package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 Types
 * =====================================================================================
 */

type taskProcessor struct {
	server *asynq.Server
	store  db.Store
	cache  cache.Cache
}

type TaskProcessor interface {
	HandleDeleteSessionsTask(ctx context.Context, t *asynq.Task) error
	Start() error
	Stop()
}

func NewRedisTaskProcessor(opt asynq.RedisClientOpt, redisCache cache.Cache, store db.Store) TaskProcessor {
	// 这样这里就能正常拿到 opt 了
	server := asynq.NewServer(opt, asynq.Config{Concurrency: 10})

	return &taskProcessor{
		server: server,
		store:  store,
		cache:  redisCache,
	}
}

type TaskDistributor interface {
	DistributeTaskDeleteSessions(ctx context.Context, jtis []string) error
}

// 3. 定义实现结构体
type redisTaskDistributor struct {
	client *asynq.Client
}

// NewRedisTaskDistributor 构造函数
func NewRedisTaskDistributor(redisOpt asynq.RedisConnOpt) TaskDistributor {
	client := asynq.NewClient(redisOpt)
	return &redisTaskDistributor{
		client: client,
	}
}

// Close 释放客户端资源（可以在 main 函数退出时调用）
func (distributor *redisTaskDistributor) Close() error {
	return distributor.client.Close()
}

type DeleteSessionsPayload struct {
	JTIs []string `json:"jtis"`
}

/** ====================================================================================
 * 🏁 Constants
 * =====================================================================================
 */

const TaskDeleteSessions = "cache:delete_sessions"

/** ====================================================================================
 * 🏁 Methods
 * =====================================================================================
 */

// 4. 实现接口方法
func (distributor *redisTaskDistributor) DistributeTaskDeleteSessions(ctx context.Context, jtis []string) error {
	// 如果传入切片为空，直接返回
	if len(jtis) == 0 {
		return nil
	}

	payload := DeleteSessionsPayload{JTIs: jtis}

	// 序列化为 JSON 字节数组
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return appError.ErrServerErr.WithErr(fmt.Errorf("无法序列化任务载荷: %w", err))
	}

	// 创建 asynq 任务，并配置任务属性
	task := asynq.NewTask(
		TaskDeleteSessions,
		jsonPayload,
		asynq.MaxRetry(3),            // 最多重试 3 次
		asynq.Timeout(2*time.Minute), // 任务执行超时时间
		asynq.ProcessIn(0),           // 立即执行
	)

	// 将任务推入 Redis 队列
	info, err := distributor.client.EnqueueContext(ctx, task)
	if err != nil {
		return appError.ErrServerErr.WithErr(fmt.Errorf("任务入队失败: %w", err))
	}

	// 打印可观测性日志
	log.Printf("成功派发异步任务: type=%s, id=%s, queue=%s, payload_size=%d",
		task.Type(), info.ID, info.Queue, len(jtis))

	return nil
}

func (processor *taskProcessor) Start() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskDeleteSessions, processor.HandleDeleteSessionsTask)

	return processor.server.Run(mux)
}

func (processor *taskProcessor) Stop() {
	processor.server.Stop()
}
