package response

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	val "github.com/go-playground/validator/v10"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/validator"
)

/** ====================================================================================
 * 🏁 Constants
 * =====================================================================================
 */

const (
	BizCodeKey = "BizCode"
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

// SuccessResult 统一成功响应结构
type SuccessResult[T any] struct {
	Code int    `json:"code" example:"200"`
	Msg  string `json:"msg" example:"操作成功"`
	Data T      `json:"data,omitempty"`
}

// ClientErrorResult 客户端请求参数错误 (400)
type ClientErrorResult struct {
	Code int    `json:"code" example:"40000"`
	Msg  string `json:"msg" example:"参数校验失败"`
}

// AuthErrorResult 认证失败 (401)
type AuthErrorResult struct {
	Code int    `json:"code" example:"40100"`
	Msg  string `json:"msg" example:"未授权，请检查令牌"`
}

// NotFoundErrorResult 资源未找到 (404)
type NotFoundErrorResult struct {
	Code int    `json:"code" example:"40400"`
	Msg  string `json:"msg" example:"资源不存在"`
}

// ConflictErrorResult 资源冲突 (409)
type ConflictErrorResult struct {
	Code int    `json:"code" example:"40900"`
	Msg  string `json:"msg" example:"数据冲突，操作无法完成"`
}

// ServerErrorResult 服务器内部错误 (500)
type ServerErrorResult struct {
	Code int    `json:"code" example:"50000"`
	Msg  string `json:"msg" example:"服务器内部发生错误"`
}

// FailWithBindError 处理 req 参数绑定错误
func FailWithBindError(c *gin.Context, err error) {
	// validator 翻译器为空
	if validator.Trans == nil {
		_ = c.Error(fmt.Errorf("CRITICAL: validator.Trans 翻译器未初始化, 无法翻译校验错误"))
		c.Set(BizCodeKey, appError.CodeServerErr)
		c.JSON(http.StatusInternalServerError, Result[any]{
			Code: appError.CodeServerErr,
			Msg:  "服务器开小差了",
		})
		return
	}

	c.Set(BizCodeKey, appError.CodeInvalidParam)

	// validator 校验错误
	if errs, ok := err.(val.ValidationErrors); ok {
		var errMsgs []string
		// 将错误翻译为中文
		for _, e := range errs.Translate(validator.Trans) {
			errMsgs = append(errMsgs, e)
		}
		// 将错误添加入上下文
		c.JSON(http.StatusOK, Result[any]{
			Code: appError.CodeInvalidParam,
			Msg:  strings.Join(errMsgs, ", "),
		})
		return
	}

	// 兜底错误
	c.JSON(http.StatusOK, Result[any]{
		Code: appError.CodeInvalidParam,
		Msg:  "参数格式或类型错误",
	})
}

// Fail 请求处理失败, 集中处理参数绑定错误, 已知错误, 服务器内部错误
func Fail(c *gin.Context, err error) {
	// 将错误加入上下文, 供 logger 打印, 此处设置 errType 为 private
	_ = c.Error(err)

	// 处理参数绑定错误
	var bindErr bindError
	if errors.As(err, &bindErr) {
		FailWithBindError(c, bindErr.error)
		return
	}

	// 处理已知错误
	var bizErr *appError.BizError
	if errors.As(err, &bizErr) {
		httpCode := http.StatusOK
		if bizErr.Code >= appError.CodeServerErr {
			httpCode = 500
		}
		c.Set(BizCodeKey, bizErr.Code)
		// 如果是内部错误, 返回 http.StatusInternalServerError
		c.JSON(httpCode, Result[any]{Code: bizErr.Code, Msg: bizErr.Msg})
		return
	}

	// 兜底: 处理未知错误
	c.Set(BizCodeKey, appError.CodeServerErr)
	c.JSON(http.StatusInternalServerError, Result[any]{
		Code: appError.CodeServerErr,
		Msg:  "服务器开小差了",
	})
}

// Success 请求处理成功, 直接返回 200
func Success[T any](c *gin.Context, data T) {
	c.Set(BizCodeKey, appError.CodeSuccess)
	c.JSON(http.StatusOK, Result[T]{Code: appError.CodeSuccess, Msg: "success", Data: data})
}
