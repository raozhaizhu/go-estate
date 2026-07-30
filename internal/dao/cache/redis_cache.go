package cache

import (
	"context"
	"encoding/json"
	"errors"
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

// buildUserQueryKey 生成统一的 Key, 用于记录用户是否查询过该数据
func (r *redisCache) buildUserQueryKey(username, startDate, endDate string) string {
	return fmt.Sprintf("query:user_query_record:%s:%s:%s", username, startDate, endDate)
}

// buildDataKey 生成统一的 Key, 用于缓存日成交数据
func (r *redisCache) buildDataKey(startDate, endDate string) string {
	return fmt.Sprintf("query:daily_data:%s:%s", startDate, endDate)
}

// GetRecordAndData 获取用户查询记录和成交数据
func (r *redisCache) GetRecordAndData(ctx context.Context, username, startDate, endDate string) (bool, []byte, error) {
	// 用户查询键: 记录用户是否查询过该数据
	recordKey := r.buildUserQueryKey(username, startDate, endDate)
	// 数据查询键: 获取缓存的成交数据
	dataKey := r.buildDataKey(startDate, endDate)

	// 开启管道
	pipe := r.client.Pipeline()
	recordCmd, dataCmd := pipe.Get(ctx, recordKey), pipe.Get(ctx, dataKey)

	// 执行并获取报错
	pipe.Exec(ctx) // 忽略总错误处理, 后续会处理分支错误
	recordErr, dataErr := recordCmd.Err(), dataCmd.Err()

	// 校验用户查询记录
	hasRecord := true
	if errors.Is(recordErr, redis.Nil) { // 为空, 说明没有记录
		hasRecord = false
	} else if recordErr != nil { // 其他内部错误
		return false, nil, recordErr
	}

	// 校验成交数据
	var rawData []byte
	var err error
	if errors.Is(dataErr, redis.Nil) { // 为空, 说明没有数据
		rawData = nil
	} else if dataErr != nil { // 其他内部错误
		return hasRecord, nil, dataErr
	} else { // 没有错误, 成功拿到数据
		rawData, err = dataCmd.Bytes()
		if err != nil {
			return hasRecord, nil, err
		}
	}

	return hasRecord, rawData, nil
}

// SetRecordAndData 写入用户查询记录和成交数据
func (r *redisCache) SetRecordAndData(ctx context.Context, username, startDate, endDate string, data any, recordTTL, dataTTL time.Duration) error {
	// 用户查询键: 记录用户是否查询过该数据
	recordKey := r.buildUserQueryKey(username, startDate, endDate)
	// 写入用户查询记录
	err := r.client.Set(ctx, recordKey, true, recordTTL).Err()
	if err != nil {
		return appError.NewSrvErr(err)
	}

	// 序列化数据
	bytes, err := json.Marshal(data)
	if err != nil {
		return appError.NewSrvErr(err)
	}

	// 数据查询键: 获取缓存的成交数据
	dataKey := r.buildDataKey(startDate, endDate)
	// 写入成交数据
	err = r.client.Set(ctx, dataKey, bytes, dataTTL).Err()
	if err != nil {
		return appError.NewSrvErr(err)
	}

	return nil
}
