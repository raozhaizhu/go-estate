package controller

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var InvalidDate = errors.New("日期参数不能为空")
var InvalidDateForm = errors.New("日期格式错误, 请使用 YYYY-MM-DD 格式")
var QueryFailed = errors.New("查询数据失败")

func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
