package async

import (
	"context"
	"log/slog"
	"time"

	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

type AsyncRunner func(ctx context.Context, logger *slog.Logger, name string, timeout time.Duration, fn func(ctx context.Context))

// Go 异步协程执行业务逻辑
func Go(ctx context.Context, logger *slog.Logger, taskName string, timeout time.Duration, fn func(asyncCtx context.Context)) {
	// 剥离主请求的 Cancel 信号，确保异步任务不会随着请求结束而中断
	bgCtx := context.WithoutCancel(ctx)

	go func() {
		// 统一捕获 Panic，防止整个进程崩溃
		defer func() {
			r := recover()
			if r != nil {
				logger.Error("后台异步任务发生崩溃(Panic)",
					"task", taskName,
					"panic", r,
				)
			}
		}()

		// 设定超时时间
		timeoutCtx, cancel := context.WithTimeout(bgCtx, timeout)
		defer cancel()

		// 执行业务逻辑
		fn(timeoutCtx)
	}()
}

// LogAsyncError 专门用于在异步协程中记录灾难级/500级错误
func LogAsyncError(ctx context.Context, logger *slog.Logger, path string, taskMsg string, err error, extraArgs ...any) {
	if err == nil {
		return
	}

	// 初始化标准 500 级异步错误属性
	logArgs := make([]any, 0, 4+len(extraArgs))
	logArgs = append(logArgs,
		slog.Int("status", 500),
		slog.Int("biz_code", appError.CodeServerErr), // 使用当前包内的 50000 错误码
		slog.String("path", path),                    // 标记异步通路
		slog.String("errors", err.Error()),
	)

	// 动态追加业务特有的上下文
	if len(extraArgs) > 0 {
		logArgs = append(logArgs, extraArgs...)
	}

	// 统一加上 "ASYNC_ERROR" 前缀，方便在日志平台一眼识别
	logger.ErrorContext(ctx, "ASYNC_ERROR: "+taskMsg, logArgs...)
}
