package user

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	mock_worker "github.com/raozhaizhu/go-estate/internal/worker/mock"
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
	buildStubs    func(store *mock_db.MockStore)
	checkResponse func(t *testing.T, res *DTO, err error)
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
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), randomUserPayload)
			},
			checkResponse: func(t *testing.T, res *DTO, err error) {
				require.Error(t, err)
				require.ErrorIs(t, err, appError.ErrAuthPermissionDenied)
				require.Nil(t, res)
			},
		},
		{
			name:  "Vip 查别人",
			input: input,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), vipPayload)
			},
			checkResponse: func(t *testing.T, res *DTO, err error) {
				require.Error(t, err)
				require.ErrorIs(t, err, appError.ErrAuthPermissionDenied)
				require.Nil(t, res)
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
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), input.Username).
					Return(user, nil).
					Times(1)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), userPayload)
			},
			checkResponse: func(t *testing.T, res *DTO, err error) {
				assert.NoError(t, err)
				assert.Equal(t, res, userDTO)
			},
		},
		{
			name:  "Admin 查询 User",
			input: input,
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().
					GetUser(gomock.Any(), input.Username).
					Return(user, nil).
					Times(1)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), adminPayload)
			},
			checkResponse: func(t *testing.T, res *DTO, err error) {
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

func setupGetUserData() (GetUserInput, db.User, *DTO) {
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

	userDTO := &DTO{
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
			storeMock := mock_db.NewMockStore(ctrl)
			cacheMock := mock_db.NewMockCache(ctrl)
			distributorMock := mock_worker.NewMockTaskDistributor(ctrl)
			deps := app.Deps{
				Store:       storeMock,
				Cache:       cacheMock,
				Distributor: distributorMock,
			}
			svc := New(deps)

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
