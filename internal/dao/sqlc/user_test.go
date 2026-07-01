package db_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Success
 * =====================================================================================
 */

// TestCreateUser 测试CreateUser/GetUser, 能正常创建用户, 且用户信息和预期一致
func TestCreateGetUser(t *testing.T) {
	testUtil.CreateRandomUser(t, testStore)
}

func TestUpdateUser(t *testing.T) {
	// 初始化更新信息
	originUser := testUtil.CreateRandomUser(t, testStore)
	password := util.RandomPassword()
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	randomUsername := util.RandomUsername()
	email := util.DeriveEmail(randomUsername)
	passwordChangedAt := time.Now()

	// 构造参数
	params := db.UpdateUserParams{
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
	password := util.RandomPassword()

	params := testUtil.PrepareCreateUserParams(t, username, password)
	testUtil.CreateSpecificUser(t, username, password, testStore)

	// 尝试重复创建用户
	result, err := testStore.CreateUser(context.Background(), params)
	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, db.WrapDBError(err), db.ErrUsernameDuplicate)
}

// TestEmailDuplicate 基于重复Email 创建用户
func TestEmailDuplicate(t *testing.T) {
	// 创建用户
	username := util.RandomUsername()
	password := util.RandomPassword()

	params := testUtil.PrepareCreateUserParams(t, username, password)
	testUtil.CreateSpecificUser(t, username, password, testStore)

	// 尝试用相同的 email 创建用户
	params.Username = util.RandomUsername() // 使用不同的用户名, 保持其他信息不变
	result, err := testStore.CreateUser(context.Background(), params)
	require.Error(t, err)
	require.Nil(t, result)
	require.ErrorIs(t, db.WrapDBError(err), db.ErrEmailDuplicate)
}

// TestGetNonExistentUser 查询不存在的用户
func TestGetNonExistentUser(t *testing.T) {
	username := util.RandomUsername()
	user, err := testStore.GetUser(context.Background(), username)
	require.Error(t, err)
	require.Equal(t, db.User{}, user)
	require.ErrorIs(t, db.WrapDBError(err), db.ErrRecordNotFound)

}
