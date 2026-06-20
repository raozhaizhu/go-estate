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
	CodeGroupClientError = 40000
	CodeGroupAuthError   = 40100
	CodeGroupNotFound    = 40400
	CodeGroupConflict    = 40900
	CodeGroupServer      = 50000
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

	// 401 认证错误
	CodeAuthInvalidToken      = CodeGroupAuthError + iota // 令牌不可用
	CodeAuthExpiredToken                                  // 令牌已过期
	CodeWrongUsernamePassword                             // 账户名或密码错误
	CodeAuthNoHeader                                      // 认证头不存在
	CodeAuthBadHeader                                     // 认证头格式错误
	CodeAuthRequired                                      // 未登录或者登录已经失效
	CodeAuthPermissionDenied                              // 角色权限不足

	// 404 资源不存在
	CodeUserNotFound = CodeGroupNotFound + iota // 用户不存在

	// 409 值冲突
	CodeUserAlreadyExits  = CodeGroupConflict + iota // 用户已存在
	CodeEmailAlreadyExits                            // 邮箱已存在

	// 500 内部错误
	CodeServerErr    = CodeGroupServer + iota // 服务器内部错误
	CodeWrongSizeKey                          // 密钥尺寸错误
)

type BizError struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (e *BizError) Error() string {
	return fmt.Sprintf("code:% d, msg: %s", e.Code, e.Msg)
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

	// 401 认证错误
	ErrInvalidToken          = New(CodeAuthInvalidToken, "令牌不可用")
	ErrExpiredToken          = New(CodeAuthExpiredToken, "令牌已过期")
	ErrWrongUsernamePassword = New(CodeWrongUsernamePassword, "账户名或密码错误")
	ErrAuthNoHeader          = New(CodeAuthNoHeader, "没有认证头")
	ErrAuthBadHeader         = New(CodeAuthBadHeader, "认证头格式错误")
	ErrAuthRequired          = New(CodeAuthRequired, "未登录或者登录已经失效")
	ErrAuthPermissionDenied  = New(CodeAuthPermissionDenied, "角色权限不足")

	// 404 资源不存在
	ErrUserNotFound = New(CodeUserNotFound, "该用户不存在")

	// 409 值冲突
	ErrUserAlreadyExits  = New(CodeUserAlreadyExits, "该用户已经存在")
	ErrEmailAlreadyExits = New(CodeEmailAlreadyExits, "该邮箱已经存在")

	// 500 服务器内部错误
	ErrServerErr = New(CodeServerErr, "服务器内部错误")
)

func NewInvalidKeySizeError(actual, minSize int) error {
	return New(CodeWrongSizeKey, fmt.Sprintf("invalid key size: current is %d, must be at least %d characters", actual, minSize))
}
