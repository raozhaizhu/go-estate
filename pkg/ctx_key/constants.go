package ctxKey

type contextKey string

const (
	CtxKeyRequestID  contextKey = "request_id"
	CtxKeyClientMeta contextKey = "ctx_client_meta"
)
