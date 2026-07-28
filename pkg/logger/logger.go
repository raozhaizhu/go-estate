package logger

import (
	"context"
	"log/slog"
	"os"

	ctxKey "github.com/raozhaizhu/go-estate/pkg/ctx_key"
)

// ContextHandler 包装原生的 slog.Handler
type ContextHandler struct {
	slog.Handler
}

// Handle 拦截每一条日志记录
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	// 从 context 中拿出 request_id
	if reqID, ok := ctx.Value(ctxKey.CtxKeyRequestID).(string); ok {
		// 追加到这条日志的属性里
		r.AddAttrs(slog.String("request_id", reqID))
	}
	// 继续交给底层的原生 Handler 去输出（比如输出成 JSON）
	return h.Handler.Handle(ctx, r)
}

func InitSlogger() {
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	ctxHandler := ContextHandler{Handler: jsonHandler}
	slog.SetDefault(slog.New(ctxHandler))
}
