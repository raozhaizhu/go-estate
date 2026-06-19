package appError

import "fmt"

/** ====================================================================================
 * 🏁 BizError
 * =====================================================================================
 *
 */
// 错误组
const (
	CodeGroupClientError = 40000
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

	// 404 资源不存在
	CodeUserNotFound = CodeGroupNotFound + iota // 用户不存在

	// 409 值冲突
	CodeUserAlreadyExits  = CodeGroupConflict + iota // 用户已存在
	CodeEmailAlreadyExits                            // 邮箱已存在

	// 500 内部错误
	CodeServerErr = CodeGroupServer + iota // 服务器内部错误
)

type BizError struct {
	Code int
	Msg  string
}

func (e *BizError) Error() string {
	return fmt.Sprintf("code:% d, msg: %s", e.Code, e.Msg)
}

func New(code int, msg string) *BizError {
	return &BizError{Code: code, Msg: msg}
}

var (
	ErrEmptyUpdate       = New(CodeEmptyUpdate, "没有任何可更新的字段")
	ErrUserNotFound      = New(CodeUserNotFound, "该用户不存在")
	ErrUserAlreadyExits  = New(CodeUserAlreadyExits, "该用户已经存在")
	ErrEmailAlreadyExits = New(CodeEmailAlreadyExits, "该邮箱已经存在")
	ErrBadDate           = New(CodeBadDate, "查询日期格式错误")
	ErrBadStartDate      = New(CodeBadStartDate, "开始日期格式错误")
	ErrBadEndDate        = New(CodeBadEndDate, "结束日期格式错误")
	ErrTimeOutOfRange    = New(CodeTimeOutOfRange, "查询日期超出范围")
	ErrBadTimerOrder     = New(CodeBadTimerOrder, "开始日期晚于结束日期")
)
