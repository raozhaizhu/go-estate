package cache

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 Types
 * =====================================================================================
 */

// Cache 缓存
type Cache interface {
	// Session
	AddNewSession(ctx context.Context, params AddNewSessionParams) error
	GetSession(ctx context.Context, jti string) (*Session, error)
	BatchDelete(ctx context.Context, jtis []string) error

	// IP
	IncrIPCnt(ctx context.Context, ip string, duration time.Duration) (int64, error)

	// QueryDailyData
	GetDailyData(ctx context.Context, username, targetDate string, dest any) (bool, error)
	SetDailyData(ctx context.Context, username, targetDate string, data any, ttl time.Duration) error

	// Lifecycle
	Close() error
	CleanTestCache(t *testing.T)
}

// SessionCache 用于管理 session
type SessionCache interface {
	AddNewSession(ctx context.Context, params AddNewSessionParams) error
	GetSession(ctx context.Context, jti string) (*Session, error)
	BatchDelete(ctx context.Context, jtis []string) error
}

// QueryCache 用于管理 dailyData 的查询记录
type QueryCache interface {
	GetDailyData(ctx context.Context, username, targetDate string, dest any) (bool, error)
	SetDailyData(ctx context.Context, username, targetDate string, data any, ttl time.Duration) error
}

func NewCache(addr, password string) (Cache, error) {
	// 初始化 Redis 客户端
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// 初始化日志
	logger := slog.Default().With("layer", "cache")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// PING 通Redis
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, appError.NewSrvErr(fmt.Errorf("无法连接到 Redis: %w", err))
	}

	return &redisCache{client: client, logger: logger}, nil
}

type redisCache struct {
	client *redis.Client
	logger *slog.Logger
}

type AddNewSessionParams struct {
	JTI string
	Session
}

type Session struct {
	Username  string
	IsBlocked bool
	ExpiresAt time.Time
}

func (p *AddNewSessionParams) toValue() map[string]interface{} {
	value := map[string]interface{}{
		"username":   p.Username,
		"is_blocked": p.IsBlocked,
		"expires_at": p.ExpiresAt.Unix(),
	}

	return value
}

func (s *Session) IsValid() error {
	// 校验阻断
	if s.IsBlocked {
		return appError.ErrBlockedSession
	}
	// 校验过期
	if time.Now().After(s.ExpiresAt) {
		return appError.ErrExpiredToken
	}

	return nil
}
