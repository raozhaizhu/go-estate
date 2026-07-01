package user

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	mock_service "github.com/raozhaizhu/go-estate/internal/service/user/mock"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"

	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 TestCreateUser
 * =====================================================================================
 */
type createUserTC struct {
	name          string
	input         CreateUserInput
	roleToCreate  role.Role
	buildCtx      func() context.Context
	buildStubs    func(store *mock_service.MockUserStore)
	checkResponse func(t *testing.T, res *DTO, err error)
}

// TestCreateUser_Duplicate 用重复username email 创建冲突账号
// 会进入 DB 查询
func TestCreateUser_Duplicate(t *testing.T) {
	input, params, _, _ := setupCreateUserData()

	testCases := []createUserTC{
		{
			name:         "用重复 username 创建 User",
			input:        input,
			roleToCreate: role.RoleUser,
			buildStubs: func(store *mock_service.MockUserStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(params, input.Password)).
					Return(nil, db.ErrUsernameDuplicate).
					Times(1)
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildCtx: func() context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *DTO, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				require.ErrorIs(t, err, appError.ErrUserAlreadyExits)
			},
		},
		{
			name:         "用重复 email 创建 User",
			input:        input,
			roleToCreate: role.RoleUser,
			buildStubs: func(store *mock_service.MockUserStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(params, input.Password)).
					Return(nil, db.ErrEmailDuplicate).
					Times(1)
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildCtx: func() context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *DTO, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				require.ErrorIs(t, err, appError.ErrEmailAlreadyExits)
			},
		},
	}

	runCreateUserTC(t, testCases)
}

func TestCreateUser_Authorization(t *testing.T) {
	input, _, _, _ := setupCreateUserData()

	userPayload := &token.Payload{
		Username: "user",
		Role:     role.RoleUser,
	}

	testCases := []createUserTC{
		{
			name:         "用 User 创建 Vip",
			input:        input,
			roleToCreate: role.RoleVip,
			buildStubs: func(store *mock_service.MockUserStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), userPayload)
			},
			checkResponse: func(t *testing.T, res *DTO, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				require.ErrorIs(t, err, appError.ErrAuthPermissionDenied)
			},
		},
		{
			name:         "创建不被允许的角色(Admin)",
			input:        input,
			roleToCreate: role.RoleAdmin,
			buildStubs: func(store *mock_service.MockUserStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			buildCtx: func() context.Context {
				return context.Background()
			},
			checkResponse: func(t *testing.T, res *DTO, err error) {
				require.Error(t, err)
				require.Nil(t, res)
				require.ErrorIs(t, err, appError.ErrServerErr)
			},
		},
	}

	runCreateUserTC(t, testCases)
}

func TestCreateUser_Success(t *testing.T) {
	input, params, user, userDTO := setupCreateUserData()

	// 准备 ctx
	adminPayload := &token.Payload{
		Username: "admin",
		Role:     role.RoleAdmin,
	}
	userPayload := &token.Payload{
		Username: "user",
		Role:     role.RoleUser,
	}

	vipParams, err := input.toDBParams(role.RoleVip)
	require.NoError(t, err)
	vip := user
	vip.Role = int16(role.RoleVip)
	// 转化为指针类型, 这里写的相当丑, 后续重构的时候要改
	vipDTOValue := *userDTO
	vipDTOValue.Role = role.RoleVip
	vipDTO := &vipDTOValue

	testCases := []createUserTC{
		{
			name:         "用 Admin 创建 Vip",
			input:        input,
			roleToCreate: role.RoleVip,
			// 执行 svc 逻辑(先调用 create 后调用 get)
			buildStubs: func(store *mock_service.MockUserStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(vipParams, input.Password)).
					Return(driver.RowsAffected(1), nil).
					Times(1)
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(vipParams.Username)).
					Return(vip, nil).
					Times(1)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), adminPayload)
			},
			checkResponse: func(t *testing.T, res *DTO, err error) {
				assert.NoError(t, err)
				assert.Equal(t, res, vipDTO)
			},
		},
		{
			name:         "创建 User",
			input:        input,
			roleToCreate: role.RoleUser,
			// 执行 svc 逻辑(先调用 create 后调用 get)
			buildStubs: func(store *mock_service.MockUserStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(params, input.Password)).
					Return(driver.RowsAffected(1), nil).
					Times(1)
				store.EXPECT().
					GetUser(gomock.Any(), gomock.Eq(params.Username)).
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
	}

	runCreateUserTC(t, testCases)
}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

// eqCreateUserParamsMatcher 手写匹配器以规避哈希密码随机问题
type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams // mock得到的数据库参数
	password string              // 传入的明文密码
}

// Matches 匹配规则
func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	// 类型断言, 传入的 x 为对应类型
	actual, ok := x.(db.CreateUserParams) // 我实际传入的参数, 带有我所生成的哈希密码
	if !ok {
		return false
	}

	// e 的明文密码, 对应 params 的哈希密码 (密码正确)
	err := util.CheckPassword(e.password, actual.HashedPassword)
	if err != nil {
		return false
	}

	// 将 e 的哈希密码, 设置为 params 的哈希密码
	// 规避了直接使用随机生成的哈希密码, 导致前后不一致的问题
	e.arg.HashedPassword = actual.HashedPassword
	return reflect.DeepEqual(e.arg, actual)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

// EqCreateUserParams 返回自制比较器
func EqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg, password}
}

func runCreateUserTC(t *testing.T, testCases []createUserTC) {
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
			res, err := svc.CreateUser(ctx, tc.input, tc.roleToCreate)

			// 校验一致性
			tc.checkResponse(t, res, err)
		})
	}
}

func setupCreateUserData() (CreateUserInput, db.CreateUserParams, db.User, *DTO) {
	// 准备 input
	username := util.RandomUsername()
	password := util.RandomPassword()
	email := util.DeriveEmail(username)
	input := CreateUserInput{
		Username: username,
		Password: password,
		Email:    email,
	}

	// 准备 params
	params, _ := input.toDBParams(role.RoleUser)

	// 准备 预埋数据
	user := db.User{
		ID:             1,
		Username:       username,
		HashedPassword: params.HashedPassword,
		Email:          email,
		Role:           int16(role.RoleUser),
	}

	userDTO := &DTO{
		ID:       1,
		Username: username,
		Email:    email,
		Role:     role.RoleUser,
	}
	return input, params, user, userDTO
}
