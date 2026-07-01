package cache

import (
	"context"
	"strconv"
	"time"

	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 Types
 * =====================================================================================
 */

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

/** ====================================================================================
 * 🏁 GetSession
 * =====================================================================================
 */

func (r *redisCache) GetSession(ctx context.Context, jti string) (*Session, error) {
	key := "session:" + jti

	// 提取 val, 校验是否存在
	val, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(val) == 0 {
		return nil, appError.ErrMissSession
	}

	// 转码 val 得到 session
	session, err := mapToSession(val)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func mapToSession(val map[string]string) (*Session, error) {
	// 参数转化
	isBlocked := val["is_blocked"] == "true"
	expireUnix, err := strconv.ParseInt(val["expires_at"], 10, 64)
	if err != nil {
		return nil, err
	}
	expiresAt := time.Unix(expireUnix, 0)

	// 构造 session
	username := val["username"]
	session := &Session{
		Username:  username,
		IsBlocked: isBlocked,
		ExpiresAt: expiresAt,
	}

	return session, nil
}

/** ====================================================================================
 * 🏁 AddNewSession
 * =====================================================================================
 */

// AddNewSession 增加新 session
func (r *redisCache) AddNewSession(ctx context.Context, params AddNewSessionParams) error {
	// 获取 kv
	key := "session:" + params.JTI
	value := params.toValue()

	// 管道操作
	err := r.pipeHSetExpire(ctx, key, value, params.ExpiresAt)

	return err
}

// pipeHSetExpire 以管道方式添加 Session 到 redis
func (r *redisCache) pipeHSetExpire(ctx context.Context, key string, val map[string]interface{}, expireAt time.Time) error {
	// 获取实际持续时间
	duration, err := getSessionDuration(expireAt)
	if err != nil {
		return err
	}

	pipe := r.client.Pipeline()
	pipe.HSet(ctx, key, val)
	pipe.Expire(ctx, key, duration)
	_, err = pipe.Exec(ctx)

	return err
}

// getSessionDuration 校验是否过期, 并获取实际持续时间(不得大于 MaxSessionDuration)
func getSessionDuration(expireAt time.Time) (time.Duration, error) {
	duration := time.Until(expireAt)

	// 持续时间不可超过限制
	if duration > maxSessionDuration {
		duration = maxSessionDuration
	}

	// 持续时间不得为负
	if duration <= 0 {
		return 0, appError.ErrExpiredToken
	}

	return duration, nil
}

/** ====================================================================================
 * 🏁 BatchDelete
 * =====================================================================================
 */

// BatchDelete 批量删除 sessions
func (r *redisCache) BatchDelete(ctx context.Context, jtis []string) error {
	// 搭建管道
	pipe := r.client.Pipeline()
	// 填充删除命令
	for _, jti := range jtis {
		key := "session:" + jti
		pipe.Unlink(ctx, key)
	}
	// 执行批量删除
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}

	return nil
}
