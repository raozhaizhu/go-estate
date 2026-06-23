package user

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	mock_service "github.com/raozhaizhu/go-estate/internal/service/user/mock"
	"github.com/raozhaizhu/go-estate/internal/util"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 TestGetUser
 * =====================================================================================
 */

type getUserTC struct {
	name          string
	input         GetUserInput
	buildCtx      func() context.Context
	buildStubs    func(store *mock_service.MockStore)
	checkResponse func(t *testing.T, res UserDTO, err error)
}

func TestGetUser_Authorization(t *testing.T) {
	input, _, _ := setupGetUserData()

	// 准备 ctx
	vipPayload := &token.Payload{
		Username: "vip",
		Role:     role.RoleVip,
	}
	randomUserPayload := &token.Payload{
		Username: util.RandomUsername(),
		Role:     role.RoleUser,
	}

	testCases := []getUserTC{
		{
			name:  "User 查别人",
			input: input,
			buildStubs: func(store *mock_service.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), randomUserPayload)
			},
			checkResponse: func(t *testing.T, res UserDTO, err error) {
				require.Error(t, err)
				require.ErrorIs(t, err, appError.ErrAuthPermissionDenied)
				require.Equal(t, res, UserDTO{})
			},
		},
		{
			name:  "Vip 查别人",
			input: input,
			buildStubs: func(store *mock_service.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), vipPayload)
			},
			checkResponse: func(t *testing.T, res UserDTO, err error) {
				require.Error(t, err)
				require.ErrorIs(t, err, appError.ErrAuthPermissionDenied)
				require.Equal(t, res, UserDTO{})
			},
		},
	}

	runGetUserTC(t, testCases)

}

func TestGetUser_Success(t *testing.T) {
	input, user, userDTO := setupGetUserData()

	// 准备 ctx
	adminPayload := &token.Payload{
		Username: "admin",
		Role:     role.RoleAdmin,
	}
	userPayload := &token.Payload{
		Username: input.Username,
		Role:     role.RoleUser,
	}

	testCases := []getUserTC{
		{
			name:  "User 查询自己",
			input: input,
			buildStubs: func(store *mock_service.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), input.Username).
					Return(user, nil).
					Times(1)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), userPayload)
			},
			checkResponse: func(t *testing.T, res UserDTO, err error) {
				assert.NoError(t, err)
				assert.Equal(t, res, userDTO)
			},
		},
		{
			name:  "Admin 查询 User",
			input: input,
			buildStubs: func(store *mock_service.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), input.Username).
					Return(user, nil).
					Times(1)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), adminPayload)
			},
			checkResponse: func(t *testing.T, res UserDTO, err error) {
				assert.NoError(t, err)
				assert.Equal(t, res, userDTO)
			},
		},
	}

	runGetUserTC(t, testCases)

}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

func setupGetUserData() (GetUserInput, db.User, UserDTO) {
	// 准备 input
	username := util.RandomUsername()
	input := GetUserInput{
		Username: username,
	}

	// 准备 预埋数据
	user := db.User{
		ID:       1,
		Username: username,
		Role:     int16(role.RoleUser),
	}

	userDTO := UserDTO{
		ID:       1,
		Username: username,
		Role:     role.RoleUser,
	}
	return input, user, userDTO
}

func runGetUserTC(t *testing.T, testCases []getUserTC) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 初始化 store, svc
			storeMock := mock_service.NewMockStore(ctrl)
			svc := New(storeMock)

			// 数据库埋桩
			tc.buildStubs(storeMock)

			// 注入上下文
			ctx := tc.buildCtx()
			res, err := svc.GetUser(ctx, tc.input)

			// 校验一致性
			tc.checkResponse(t, res, err)
		})
	}
}
