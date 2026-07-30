package testUtil

import (
	"context"
	"testing"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

func CreateRandomUser(t *testing.T, testStore db.Store, role userDomain.Role, avatarKey string) (db.User, string) {
	// 初始化用户信息
	username := util.RandomUsername()
	password := util.RandomPassword()

	return CreateSpecificUser(t, username, password, testStore, role, avatarKey)
}

func CreateSpecificUser(t *testing.T, username, password string, testStore db.Store, role userDomain.Role, avatarKey string) (db.User, string) {
	// 构造参数
	params := PrepareCreateUserParams(t, username, password, role, avatarKey)

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

	return user, password
}

func PrepareCreateUserParams(t *testing.T, username, password string, role userDomain.Role, avatarKey string) db.CreateUserParams {
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	email := util.DeriveEmail(username)
	if avatarKey == "" {
		avatarKey = objectStore.DefaultAvatarKey
	}

	params := db.CreateUserParams{
		Username:       username,
		HashedPassword: hashedPassword,
		Email:          email,
		Role:           int16(role),
		AvatarKey:      avatarKey,
	}

	return params
}

