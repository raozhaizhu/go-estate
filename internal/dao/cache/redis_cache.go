package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Lifecycle
 * =====================================================================================
 */

// Close 关闭缓存
func (r *redisCache) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// CleanTestCache 清理测试缓存
func (r *redisCache) CleanTestCache(t *testing.T) {
	err := r.client.FlushDB(context.Background()).Err()
	require.NoError(t, err, "清理 Redis 测试缓存失败")
}

/** ====================================================================================
 * 🏁 General
 * =====================================================================================
 */

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

/** ====================================================================================
 * 🏁 Session_Manage
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

// AddNewSession 增加新 session
func (r *redisCache) AddNewSession(ctx context.Context, params AddNewSessionParams) error {
	// 获取 kv
	key := "session:" + params.JTI
	value := params.toValue()

	// 管道操作
	err := r.pipeHSetExpire(ctx, key, value, params.ExpiresAt)

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

/** ====================================================================================
 * 🏁 IP_Manage
 * =====================================================================================
 */

const InvalidCnt = -1

// IncrIPCnt 增加 IP 访问计数并返回当前计数
func (r *redisCache) IncrIPCnt(ctx context.Context, ip string, duration time.Duration) (int64, error) {
	key := "ratelimit" + ip
	// 搭建管道
	pipe := r.client.Pipeline()
	// 增加计数
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, duration)
	_, err := pipe.Exec(ctx)

	if err != nil {
		return InvalidCnt, nil
	}

	currCnt := incr.Val()

	return currCnt, nil
}

/** ====================================================================================
 * 🏁 Query_Manage
 * =====================================================================================
 */

// buildQueryKeyByDay 生成统一的 Key, 用于 getDataByDay 查询
func (r *redisCache) buildQueryKeyByDay(username, targetDate string) string {
	return fmt.Sprintf("query:daily_data:%s:%s", username, targetDate)
}

// GetDailyData 尝试获取已付费的查询缓存
// 返回(是否命中,错误) 并将数据序列化到 dest
func (r *redisCache) GetDailyData(ctx context.Context, username, targetDate string, dest any) (bool, error) {
	// 生成唯一键
	key := r.buildQueryKeyByDay(username, targetDate)

	// 查询唯一键
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		// 缓存不存在或已过期
		return false, nil
	} else if err != nil {
		// Redis异常
		r.logger.ErrorContext(ctx, "redis 获取数据失败", "err", err.Error())
		return false, appError.NewSrvErr(err)
	}

	// 缓存命中,反序列化回结构体
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		// 序列化异常
		r.logger.ErrorContext(ctx, "redis 序列化数据失败", "err", err.Error())
		return false, appError.NewSrvErr(err)
	}

	return true, nil
}

// SetDailyData 写入查询缓存, 设置过期时间
func (r *redisCache) SetDailyData(ctx context.Context, username, targetDate string, data any, ttl time.Duration) error {
	// 生成唯一键
	key := r.buildQueryKeyByDay(username, targetDate)

	bytes, err := json.Marshal(data)
	if err != nil {
		r.logger.ErrorContext(ctx, "redis 序列化数据失败", "err", err.Error())
		return appError.NewSrvErr(err)
	}

	err = r.client.Set(ctx, key, bytes, ttl).Err()
	if err != nil {
		r.logger.ErrorContext(ctx, "redis 设置键失败", "err", err.Error())
		return appError.NewSrvErr(err)
	}

	return nil
}
