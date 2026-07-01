package worker

import (
	"context"

	"github.com/hibiken/asynq"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
)

/** ====================================================================================
 * 🏁 Types
 * =====================================================================================
 */

type redisTaskProcessor struct {
	sessionCache cache.SessionCache
}

type RedisTaskProcessor interface {
	HandleDeleteSessionsTask(ctx context.Context, t *asynq.Task) error
}

func NewRedisTaskProcessor(sessionCache cache.SessionCache) RedisTaskProcessor {
	return &redisTaskProcessor{sessionCache: sessionCache}
}
