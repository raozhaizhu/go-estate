package response

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	val "github.com/go-playground/validator/v10"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/validator"
)

/** ====================================================================================
 * 🏁 bindError
 * =====================================================================================
 */

// bindError 供 Controller 使用, 当 req 绑定失败时抛出
type bindError struct{ error }

// MarkBindError 供 Controller 使用, 为参数绑定错误
func MarkBindError(err error) error {
	return bindError{err}
}

/** ====================================================================================
 * 🏁 Fail
 * =====================================================================================
 *
 */

type Result[T any] struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data T      `json:"data,omitempty"`
}

// FailWithBindError 处理 req 参数绑定错误
func FailWithBindError(c *gin.Context, err error) {
	if validator.Trans == nil {
		log.Printf("翻译器没有初始化")
		return
	}
	// validator 校验错误
	if errs, ok := err.(val.ValidationErrors); ok {
		var errMsgs []string
		// 将错误翻译为中文
		for _, e := range errs.Translate(validator.Trans) {
			errMsgs = append(errMsgs, e)
		}
		// 将错误添加入上下文
		c.JSON(http.StatusBadRequest, Result[any]{
			Code: appError.CodeInvalidParam,
			Msg:  strings.Join(errMsgs, ", "),
		})
		return
	}

	// 兜底错误
	c.JSON(http.StatusBadRequest, Result[any]{
		Code: appError.CodeInvalidParam,
		Msg:  "参数格式或类型错误",
	})
}

// Fail 请求处理失败, 集中处理参数绑定错误, 已知错误, 服务器内部错误
func Fail(c *gin.Context, err error) {
	// log.Printf("[ERROR] 类型: %T | 内容: %v", err, err)

	// 处理参数绑定错误
	var bindErr bindError
	if errors.As(err, &bindErr) {
		FailWithBindError(c, bindErr.error)
		return
	}

	// 处理已知错误
	var bizErr *appError.BizError
	if errors.As(err, &bizErr) {
		c.JSON(http.StatusOK, Result[any]{Code: bizErr.Code, Msg: bizErr.Msg})
		return
	}

	// 兜底: 处理未知错误
	log.Printf("未知 [ERROR] 类型: %T | 内容: %v", err, err)
	c.JSON(http.StatusInternalServerError, Result[any]{
		Code: appError.CodeServerErr,
		Msg:  "服务器开小差了",
	})
}

// Success 请求处理成功, 直接返回 200
func Success[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Result[T]{Code: 200, Msg: "success", Data: data})
}
