package ctxKey

import (
	"context"
	"fmt"

	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

const (
	CtxKeyRequestID    string = "request_id"
	CtxKeyClientMeta   string = "ctx_client_meta"
	CtxKeyPayload      string = "authorization_payload"
	CtxKeyRefreshToken string = "refresh_token"
	CtxKeyUserName     string = "username"
)

func GetUsername(ctx context.Context) (string, error) {
	str, err := getStringFromCtx(ctx, CtxKeyUserName)
	if err != nil {
		return "", err
	}

	return str, nil
}

func getStringFromCtx(ctx context.Context, key string) (string, error) {
	// 提取 val
	str, ok := ctx.Value(key).(string)
	if !ok {
		return "", appError.NewSrvErr(fmt.Errorf("该 Key 并非 String 类型"))
	}

	return str, nil
}
