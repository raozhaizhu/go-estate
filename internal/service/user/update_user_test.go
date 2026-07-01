package user

import (
	"context"
	"database/sql/driver"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	mock_service "github.com/raozhaizhu/go-estate/internal/service/user/mock"
	"github.com/raozhaizhu/go-estate/internal/util"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 TestUpdateUser
 * =====================================================================================
 */

func TestUpdateUser_Authorization(t *testing.T) {
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
			name:  "User 更新别人",
			input: input,
			buildStubs: func(store *mock_service.MockUserStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
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
			name:  "Vip 更新别人",
			input: input,
			buildStubs: func(store *mock_service.MockUserStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
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

// TestUpdateUser_Success 更新用户信息: 仅允许本人或Admin更新
func TestUpdateUser_Success(t *testing.T) {
	input, user, userDTO := setupUpdateUserData()
	params, err := input.toDBParams()
	require.NoError(t, err)

	// 准备 ctx
	adminPayload := &token.Payload{
		Username: "admin",
		Role:     role.RoleAdmin,
	}
	userPayload := &token.Payload{
		Username: input.Username,
		Role:     role.RoleUser,
	}

	testCases := []updateUserTC{
		{
			name:  "User 更新自己",
			input: input,
			buildStubs: func(store *mock_service.MockUserStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), EqUpdateUserParams(params, *input.Password)).
					Return(driver.RowsAffected(1), nil).
					Times(1)
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
			name:  "Admin 更新 User",
			input: input,
			buildStubs: func(store *mock_service.MockUserStore) {
				store.EXPECT().
					UpdateUser(gomock.Any(), EqUpdateUserParams(params, *input.Password)).
					Return(driver.RowsAffected(1), nil).
					Times(1)
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

	runUpdateUserTC(t, testCases)
}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

type updateUserTC struct {
	name          string
	input         UpdateUserInput
	buildCtx      func() context.Context
	buildStubs    func(store *mock_service.MockUserStore)
	checkResponse func(t *testing.T, res *DTO, err error)
}

func setupUpdateUserData() (UpdateUserInput, db.User, *DTO) {
	// 准备 input
	username := util.RandomUsername()
	newPassword := util.RandomPassword()
	newEmail := util.RandomEmail()

	input := UpdateUserInput{
		Username: username,
		Password: &newPassword,
		Email:    &newEmail,
	}

	// 预埋数据
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

func runUpdateUserTC(t *testing.T, testCases []updateUserTC) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 初始化 store, svc
			storeMock := mock_service.NewMockUserStore(ctrl)
			svc := New(storeMock)

			// 数据库埋桩
			tc.buildStubs(storeMock)

			// 注入上下文
			ctx := tc.buildCtx()
			res, err := svc.UpdateUser(ctx, tc.input)

			// 校验一致性
			tc.checkResponse(t, res, err)
		})
	}
}

// eqUpdateUserParamsMatcher 手写匹配器以规避哈希密码随机问题
type eqUpdateUserParamsMatcher struct {
	arg      db.UpdateUserParams // mock得到的数据库参数
	password string              // 传入的明文密码
}

// Matches 匹配规则
func (e eqUpdateUserParamsMatcher) Matches(x interface{}) bool {
	// 类型断言, 传入的 x 为对应类型
	actual, ok := x.(db.UpdateUserParams) // 我实际传入的参数, 带有我所生成的哈希密码
	if !ok {
		return false
	}

	// 校验mock密码和实际哈希密码是否对应
	if err := util.CheckPassword(e.password, actual.HashedPassword.String); err != nil {
		return false
	}

	// 校验其他字段一致
	if actual.Username != e.arg.Username || actual.Email != e.arg.Email {
		return false
	}

	return true
}

func (e eqUpdateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

// EqUpdateUserParams 返回自制比较器
func EqUpdateUserParams(arg db.UpdateUserParams, password string) gomock.Matcher {
	return eqUpdateUserParamsMatcher{arg, password}
}
