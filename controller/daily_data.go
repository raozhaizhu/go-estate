package controller

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/db/sqlc"
	"github.com/raozhaizhu/go-estate/util"
)

type DailyDataQuerier interface {
	GetDataByDay(ctx context.Context, targetDate time.Time) ([]db.DailyDatum, error)
	GetDataByPeriod(ctx context.Context, startDate, endDate time.Time) ([]db.DailyDatum, error)
	GetAllData(ctx context.Context) ([]db.DailyDatum, error)
}

type DailyDataController struct {
	store DailyDataQuerier
}

func NewDailyDataController(s DailyDataQuerier) *DailyDataController {
	return &DailyDataController{store: s}
}

type GetDataByDayRequest struct {
	DateStr string `form:"date" binding:"required"`
}

// GetDataByDay HTTP GET /daily_data/day?date=2026-5-1
func (c *DailyDataController) GetDataByDay(ctx *gin.Context) {
	var req GetDataByDayRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrEmptyDate))
		return
	}

	targetDate, err := time.Parse(util.DateFormat, req.DateStr)
	if err != nil { // 日期格式错误
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrInvalidDateForm))
		return
	}

	data, err := c.store.GetDataByDay(ctx, targetDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, data)
}

type GetDataByPeriodRequest struct {
	StartDateStr string `form:"start" binding:"required"`
	EndDateStr   string `form:"end" binding:"required"`
}

// GetDataByPeriod HTTP GET /daily_data/period?start=2026-5-1&end=2026-5-20
func (c *DailyDataController) GetDataByPeriod(ctx *gin.Context) {
	var req GetDataByPeriodRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrEmptyDate))
		return
	}

	startDate, err1 := time.Parse(util.DateFormat, req.StartDateStr)
	endDate, err2 := time.Parse(util.DateFormat, req.EndDateStr)
	if err1 != nil || err2 != nil { // 日期格式错误
		ctx.JSON(http.StatusBadRequest, errorResponse(ErrInvalidDateForm))
		return
	}

	data, err := c.store.GetDataByPeriod(ctx, startDate, endDate)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, data)
}

// GetAllData HTTP GET /daily_data/all
func (c *DailyDataController) GetAllData(ctx *gin.Context) {
	data, err := c.store.GetAllData(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, data)
}
