package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

/** ====================================================================================
 * 🏁 Types
 * =====================================================================================
 */

// Cache 缓存
type Cache interface {
	AddNewSession(ctx context.Context, params AddNewSessionParams) error
	GetSession(ctx context.Context, jti string) (*Session, error)
	BatchDelete(ctx context.Context, jtis []string) error
	IncrIPCnt(ctx context.Context, ip string, duration time.Duration) (int64, error)
}

// SessionCache 用于管理 session
type SessionCache interface {
	AddNewSession(ctx context.Context, params AddNewSessionParams) error
	GetSession(ctx context.Context, jti string) (*Session, error)
	BatchDelete(ctx context.Context, jtis []string) error
}

func NewCache(addr, password string) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("无法连接到 Redis: %w", err)
	}

	return &redisCache{client: client}, nil
}

type redisCache struct {
	client *redis.Client
}
