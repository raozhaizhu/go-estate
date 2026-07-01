package testUtil

import (
	"context"
	"testing"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

func CreateRandomUser(t *testing.T, testStore db.Store) db.User {
	// 初始化用户信息
	username := util.RandomUsername()
	password := util.RandomPassword()

	return CreateSpecificUser(t, username, password, testStore)
}

func CreateSpecificUser(t *testing.T, username, password string, testStore db.Store) db.User {
	// 构造参数
	params := PrepareCreateUserParams(t, username, password)

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

func PrepareCreateUserParams(t *testing.T, username, password string) db.CreateUserParams {
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	email := util.DeriveEmail(username)
	_role := int16(role.RoleUser)

	params := db.CreateUserParams{
		Username:       username,
		HashedPassword: hashedPassword,
		Email:          email,
		Role:           _role,
	}

	return params
}
