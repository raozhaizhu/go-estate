package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Success
 * =====================================================================================
 */

// TestCreateUser 测试CreateUser/GetUser, 能正常创建用户, 且用户信息和预期一致
func TestCreateGetUser(t *testing.T) {
	createThenGetRandomUser(t)
}

func TestUpdateUser(t *testing.T) {
	// 初始化更新信息
	originUser := createThenGetRandomUser(t)
	password := util.RandomPassword()
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	randomUsername := util.RandomUsername()
	email := util.DeriveEmail(randomUsername)
	passwordChangedAt := time.Now()

	// 构造参数
	params := UpdateUserParams{
		HashedPassword:    sql.NullString{String: hashedPassword, Valid: true},
		PasswordChangedAt: sql.NullTime{Time: passwordChangedAt, Valid: true},
		Email:             sql.NullString{String: email, Valid: true},
		Username:          originUser.Username,
	}

	// 更新用户
	result, err := testStore.UpdateUser(context.Background(), params)
	require.NoError(t, err)
	rows, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)

	// 查询并比较新老用户
	updatedUser, err := testStore.GetUser(context.Background(), originUser.Username)
	require.NoError(t, err)
	require.Equal(t, originUser.Username, updatedUser.Username)
	require.Equal(t, hashedPassword, updatedUser.HashedPassword)
	require.Equal(t, email, updatedUser.Email)
	require.Equal(t, originUser.Role, updatedUser.Role)
	require.WithinDuration(t, passwordChangedAt, updatedUser.PasswordChangedAt, time.Second)
	require.Equal(t, originUser.CreatedAt, updatedUser.CreatedAt)
	require.NotEmpty(t, originUser.ID)
	require.Equal(t, originUser.ID, updatedUser.ID)
}

/** ====================================================================================
 * 🏁 Fail
 * =====================================================================================
 */

// TestUserNameDuplicate 基于重复用户名创建用户
func TestUserNameDuplicate(t *testing.T) {
	// 创建用户
	username := util.RandomUsername()
	params := PrepareCreateUserParams(t, username)
	createThenGetSpecificUser(t, username)

	// 尝试重复创建用户
	result, err := testStore.CreateUser(context.Background(), params)
	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, WrapDBError(err), ErrUsernameDuplicate)
}

// TestEmailDuplicate 基于重复Email 创建用户
func TestEmailDuplicate(t *testing.T) {
	// 创建用户
	username := util.RandomUsername()
	params := PrepareCreateUserParams(t, username)
	createThenGetSpecificUser(t, username)

	// 尝试用相同的 email 创建用户
	params.Username = util.RandomUsername() // 使用不同的用户名, 保持其他信息不变
	result, err := testStore.CreateUser(context.Background(), params)
	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, WrapDBError(err), ErrEmailDuplicate)
}

// TestGetNonExistentUser 查询不存在的用户
func TestGetNonExistentUser(t *testing.T) {
	username := util.RandomUsername()
	user, err := testStore.GetUser(context.Background(), username)
	require.Error(t, err)
	require.Equal(t, User{}, user)
	require.ErrorIs(t, WrapDBError(err), ErrRecordNotFound)

}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

func createThenGetRandomUser(t *testing.T) User {
	// 初始化用户信息
	username := util.RandomUsername()
	return createThenGetSpecificUser(t, username)
}

func createThenGetSpecificUser(t *testing.T, username string) User {
	// 构造参数
	params := PrepareCreateUserParams(t, username)

	// 指定期望时间
	expectedPwdChangedAt := time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC)
	expectedCreatedAt := time.Now()

	// 创建用户
	result, err := testStore.CreateUser(context.Background(), params)
	require.NoError(t, err)
	rows, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)

	// 查询并比较用户
	user, err := testStore.GetUser(context.Background(), username)
	require.NoError(t, err)
	require.Equal(t, username, user.Username)
	require.Equal(t, params.HashedPassword, user.HashedPassword)
	require.Equal(t, params.Email, user.Email)
	require.Equal(t, params.Role, user.Role)
	require.Equal(t, expectedPwdChangedAt, user.PasswordChangedAt)
	require.WithinDuration(t, expectedCreatedAt, user.CreatedAt, time.Second)
	require.NotEmpty(t, user.ID)

	return user
}

func PrepareCreateUserParams(t *testing.T, username string) CreateUserParams {
	password := util.RandomPassword()
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	email := util.DeriveEmail(username)
	_role := int16(role.RoleUser)

	params := CreateUserParams{
		Username:       username,
		HashedPassword: hashedPassword,
		Email:          email,
		Role:           _role,
	}

	return params
}
