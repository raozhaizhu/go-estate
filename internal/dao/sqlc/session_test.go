package db_test

import (
	"context"
	"testing"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/stretchr/testify/require"
)

// TestCreateGetSession_Success 测试成功创建并获取会话
func TestCreateGetSession_Success(t *testing.T) {
	user := testUtil.CreateRandomUser(t, testStore)
	deviceID := testUtil.DeviceID
	createThenGetSessionByUser(t, user, deviceID)
}

func createThenGetSessionByUser(t *testing.T, user db.User, deviceID string) db.Session {
	// arrange
	id := util.RandomUUID()
	userAgent := testUtil.UserAgent
	clientIp := testUtil.ClientIp
	createdAt := time.Now()
	expiresAt := createdAt.Add(config.RefreshTokenDuration)
	// act
	err := testStore.CreateSession(context.Background(), db.CreateSessionParams{
		ID:        id,
		Username:  user.Username,
		DeviceID:  deviceID,
		UserAgent: userAgent,
		ClientIp:  clientIp,
		ExpiresAt: expiresAt,
	})
	// assert
	require.NoError(t, err)

	// act
	session, err := testStore.GetSession(context.Background(), id)
	// assert
	require.NoError(t, err)
	require.NotEmpty(t, session)
	require.Equal(t, id, session.ID)
	require.Equal(t, user.Username, session.Username)
	require.Equal(t, deviceID, session.DeviceID)
	require.Equal(t, userAgent, session.UserAgent)
	require.Equal(t, clientIp, session.ClientIp)
	require.Equal(t, false, session.IsBlocked)
	require.WithinDuration(t, createdAt, session.CreatedAt, time.Second)
	require.WithinDuration(t, expiresAt, session.ExpiresAt, time.Second)

	return session
}

// TestGetActiveSessionIDsByUserDevice_Success 测试成功获取会话ids
func TestGetActiveSessionIDsByUserDevice_Success(t *testing.T) {
	// arrange
	user := testUtil.CreateRandomUser(t, testStore)
	deviceID := testUtil.DeviceID
	session1 := createThenGetSessionByUser(t, user, deviceID)
	session2 := createThenGetSessionByUser(t, user, deviceID)
	// act
	ids, err := testStore.GetActiveSessionIDsByUserDevice(context.Background(), db.GetActiveSessionIDsByUserDeviceParams{
		Username: user.Username,
		DeviceID: deviceID,
	})
	// assert
	require.NoError(t, err)
	require.NotEmpty(t, ids)
	require.ElementsMatch(t, ids, []string{session1.ID, session2.ID})
}

// TestBlockSessionsByIDs_Success 测试使用 ids封锁相应会话
func TestBlockSessionsByIDs_Success(t *testing.T) {
	// arrange
	user1 := testUtil.CreateRandomUser(t, testStore)
	user2 := testUtil.CreateRandomUser(t, testStore)
	deviceID := testUtil.DeviceID
	session1 := createThenGetSessionByUser(t, user1, deviceID)
	session2 := createThenGetSessionByUser(t, user2, deviceID)
	// act
	err := testStore.BlockSessionsByIDs(context.Background(), []string{session1.ID, session2.ID})
	// assert
	require.NoError(t, err)
	session1Changed, _ := testStore.GetSession(context.Background(), session1.ID)
	session2Changed, _ := testStore.GetSession(context.Background(), session2.ID)
	require.True(t, session1Changed.IsBlocked)
	require.True(t, session2Changed.IsBlocked)
}

// TestBlockAllUserSessions_Success 测试封锁 user 名下所有 sessions
func TestBlockAllUserSessions_Success(t *testing.T) {
	// arrange
	user := testUtil.CreateRandomUser(t, testStore)
	deviceID := testUtil.DeviceID
	session1 := createThenGetSessionByUser(t, user, deviceID)
	session2 := createThenGetSessionByUser(t, user, deviceID)
	// act
	err := testStore.BlockAllUserSessions(context.Background(), user.Username)
	// assert
	require.NoError(t, err)
	session1Changed, _ := testStore.GetSession(context.Background(), session1.ID)
	session2Changed, _ := testStore.GetSession(context.Background(), session2.ID)
	require.True(t, session1Changed.IsBlocked)
	require.True(t, session2Changed.IsBlocked)
}
