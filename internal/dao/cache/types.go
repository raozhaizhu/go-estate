package cache

import (
	"context"

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
}

// SessionCache 用于管理 session
type SessionCache interface {
	AddNewSession(ctx context.Context, params AddNewSessionParams) error
	GetSession(ctx context.Context, jti string) (*Session, error)
	BatchDelete(ctx context.Context, jtis []string) error
}

func NewSessionCache(addr, password string) SessionCache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	return &redisCache{client: client}
}

type redisCache struct {
	client *redis.Client
}
