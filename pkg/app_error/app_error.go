package appError

import (
	"fmt"
)

/** ====================================================================================
 * 🏁 BizError
 * =====================================================================================
 *
 */

// 错误组
const (
	CodeSuccess              = 200
	CodeGroupClientError     = 40000
	CodeGroupAuthError       = 40100
	CodeGroupNotFound        = 40400
	CodeGroupConflict        = 40900
	CodeGroupTooManyRequests = 42900
	CodeGroupServer          = 50000
)

// 已知错误类型
const (
	// 400 客户端错误
	CodeInvalidParam   = CodeGroupClientError + iota // 格式错误, validator 会拦截
	CodeEmptyUpdate                                  // 更新为空
	CodeBadDate                                      // 查询日期格式错误
	CodeBadStartDate                                 // 开始日期格式错误
	CodeBadEndDate                                   // 结束日期格式错误
	CodeTimeOutOfRange                               // 查询日期超出范围
	CodeBadTimerOrder                                // 开始日期晚于结束日期
	CodeEmptyDeviceID                                // 设备 ID 不得为空
	CodeEmptyUserAgent                               // 用户代理不得为空
)

const (
	// 401 认证错误
	// 令牌相关
	CodeAuthInvalidToken     = CodeGroupAuthError + iota // 令牌不可用
	CodeAuthExpiredToken                                 // 令牌已过期
	CodeCookieNoRefreshToken                             // cookie 内没有 freshToken
	CodeNoSession                                        // 数据内没有 session
	CodeMissRefreshToken                                 // 缓存内没有 session
	CodeBlockedSession                                   // session 已注销

	// 其他错误
	CodeWrongUsernamePassword // 账户名或密码错误
	CodeAuthNoHeader          // 认证头不存在
	CodeAuthBadHeader         // 认证头格式错误
	CodeAuthRequired          // 未登录或者登录已经失效
	CodeAuthPermissionDenied  // 角色权限不足
)

const (
	// 404 资源不存在
	CodePathNotFound = CodeGroupNotFound + iota // 路径不存在
	CodeUserNotFound
)

const (
	// 409 值冲突
	CodeUserAlreadyExits  = CodeGroupConflict + iota // 用户已存在
	CodeEmailAlreadyExits                            // 邮箱已存在
)

const (
	// 429 访问次数过高
	CodeTooManyRequests = CodeGroupTooManyRequests + iota // 访问次数过高
)

const (
	// 500 内部错误
	CodeServerErr    = CodeGroupServer + iota // 服务器内部错误
	CodeWrongSizeKey                          // 密钥尺寸错误
)

type BizError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Err  error  `json:"-"`
}

func (e *BizError) Error() string {
	return fmt.Sprintf("code:% d, msg: %s", e.Code, e.Msg)
}

func (e *BizError) Unwrap() error {
	return e.Err
}
func (e *BizError) Is(targetErr error) bool {
	t, ok := targetErr.(*BizError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

func (e *BizError) WithErr(err error) *BizError {
	return &BizError{
		Code: e.Code,
		Msg:  e.Msg,
		Err:  err,
	}
}

func New(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

var (
	// 400 客户端错误
	ErrEmptyUpdate    = New(CodeEmptyUpdate, "没有任何可更新的字段")
	ErrBadDate        = New(CodeBadDate, "查询日期格式错误")
	ErrBadStartDate   = New(CodeBadStartDate, "开始日期格式错误")
	ErrBadEndDate     = New(CodeBadEndDate, "结束日期格式错误")
	ErrTimeOutOfRange = New(CodeTimeOutOfRange, "查询日期超出范围")
	ErrBadTimerOrder  = New(CodeBadTimerOrder, "开始日期晚于结束日期")
	ErrEmptyDeviceID  = New(CodeEmptyDeviceID, "设备 ID 不得为空")
	ErrEmptyUserAgent = New(CodeEmptyUserAgent, "用户代理不得为空")

	// 401 认证错误
	ErrWrongUsernamePassword = New(CodeWrongUsernamePassword, "账户名或密码错误")
	ErrAuthNoHeader          = New(CodeAuthNoHeader, "没有认证头")
	ErrAuthBadHeader         = New(CodeAuthBadHeader, "认证头格式错误")
	ErrAuthRequired          = New(CodeAuthRequired, "未登录或者登录已经失效")
	ErrAuthPermissionDenied  = New(CodeAuthPermissionDenied, "角色权限不足")

	ErrInvalidToken         = New(CodeAuthInvalidToken, "令牌不可用")
	ErrExpiredToken         = New(CodeAuthExpiredToken, "令牌已过期")
	ErrCookieNoRefreshToken = New(CodeCookieNoRefreshToken, "cookie 内没有 freshToken")

	ErrMissSession    = New(CodeMissRefreshToken, "缓存内没有 session")
	ErrNoSession      = New(CodeNoSession, "会话不存在")
	ErrBlockedSession = New(CodeBlockedSession, "session 已注销")

	// 404 资源不存在
	ErrPathNotFound = New(CodePathNotFound, "请求路径不存在")
	ErrUserNotFound = New(CodeUserNotFound, "用户不存在")

	// 409 值冲突
	ErrUserAlreadyExits  = New(CodeUserAlreadyExits, "该用户已经存在")
	ErrEmailAlreadyExits = New(CodeEmailAlreadyExits, "该邮箱已经存在")

	ErrTooManyRequests = New(CodeTooManyRequests, "请求过多,之后再试")

	// 500 服务器内部错误
	ErrServerErr = New(CodeServerErr, "服务器开小差了")
)

func NewInvalidKeySizeError(actual, minSize int) error {
	return New(CodeWrongSizeKey, fmt.Sprintf("invalid key size: current is %d, must be at least %d characters", actual, minSize))
}
