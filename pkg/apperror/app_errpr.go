package appError

import "fmt"

const (
	CodeInvalidParam      = 40000 // 格式错误, validator 会拦截
	CodeEmptyUpdate       = 40001 // 更新为空
	CodeUserNotFound      = 40401 // 用户不存在
	CodeUserAlreadyExits  = 40901 // 用户已存在
	CodeEmailAlreadyExits = 40902 // 邮箱已存在
	CodeServerErr         = 50000 // 服务器内部错误
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
)
