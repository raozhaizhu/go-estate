package controller

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/service"
	"github.com/raozhaizhu/go-estate/util"
)

type DailyDataController struct {
	service service.DailyDataServiceInterface
}

func NewDailyDataController(s service.DailyDataServiceInterface) *DailyDataController {
	return &DailyDataController{service: s}
}

// GetDataByDay HTTP GET /daily_data/day?date=2026-5-1
func (c *DailyDataController) GetDataByDay(ctx *gin.Context) {
	dateStr := ctx.Query("date")
	if dateStr == "" { // 日期为空
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrEmptyDate))
		return
	}

	targetDate, err := time.Parse(util.DateFormat, dateStr)
	if err != nil { // 日期格式不正确
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrInvalidDateForm))
		return
	}

	data, err := c.service.GetDataByDay(ctx, targetDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, data)
}

// GetDataByPeriod HTTP GET /daily_data/period?start=2026-5-1&end=2026-5-20
func (c *DailyDataController) GetDataByPeriod(ctx *gin.Context) {
	startDateStr, endDateStr := ctx.Query("start"), ctx.Query("end")
	if startDateStr == "" || endDateStr == "" { // 日期非空
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrEmptyDate))
		return
	}

	startDate, err1 := time.Parse(util.DateFormat, startDateStr)
	endDate, err2 := time.Parse(util.DateFormat, endDateStr)
	if err1 != nil || err2 != nil { // 日期格式化正确
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrInvalidDateForm))
		return
	}

	data, err := c.service.GetDataByPeriod(ctx, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, data)
}

// GetAllData HTTP GET /daily_data/all
func (c *DailyDataController) GetAllData(ctx *gin.Context) {
	data, err := c.service.GetAllData(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, data)
}
