package response

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	val "github.com/go-playground/validator/v10"
	appError "github.com/raozhaizhu/go-estate/pkg/apperror"
	"github.com/raozhaizhu/go-estate/pkg/validator"
)

func Error(err error) gin.H {
	return gin.H{"error": err.Error()}
}

type Result struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// FailWithBindError 处理 req 参数绑定错误
func FailWithBindError(c *gin.Context, err error) {
	if validator.Trans == nil {
		log.Printf("翻译器没有初始化")
		return
	}
	// err 属于校验错误
	if errs, ok := err.(val.ValidationErrors); ok {
		var errMsgs []string
		// 将错误翻译为中文
		for _, e := range errs.Translate(validator.Trans) {
			errMsgs = append(errMsgs, e)
		}
		// 将错误添加入上下文
		c.JSON(http.StatusBadRequest, Result{
			Code: appError.CodeInvalidParam,
			Msg:  strings.Join(errMsgs, ", "),
		})
		return
	}
	// err 属于非校验错误
	c.JSON(http.StatusBadRequest, Result{
		Code: appError.CodeInvalidParam,
		Msg:  "参数格式或类型错误",
	})
}

// caller: FailWithError
func convertToBizError(err error) *appError.BizError {
	// 本身就是 BizError，直接返回
	var bizErr *appError.BizError
	if errors.As(err, &bizErr) {
		return bizErr
	}
	// 其他未知错误，返回 nil，交给 Caller 处理
	return nil
}

// FailWithError 处理自定义错误, 或者服务器内部错误
func FailWithError(c *gin.Context, err error) {
	log.Printf("[ERROR] 类型: %T | 内容: %v", err, err)
	// 处理已知错误
	bizErr := convertToBizError(err)
	if bizErr != nil {
		FailWithBizError(c, bizErr)
		return
	}
	// 处理未知错误
	log.Printf("未知 [ERROR] 类型: %T | 内容: %v", err, err)
	c.JSON(http.StatusInternalServerError, Result{
		Code: appError.CodeServerErr,
		Msg:  "服务器开小差了",
	})
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Result{Code: 200, Msg: "success", Data: data})
}

func FailWithBizError(c *gin.Context, err *appError.BizError) {
	c.JSON(http.StatusOK, Result{Code: err.Code, Msg: err.Msg})
}
