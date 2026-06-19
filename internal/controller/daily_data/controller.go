package dailyData

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	service "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

type DailyDataQuerier interface {
	GetDataByDay(ctx context.Context, p service.GetDataByDayParams) ([]db.DailyDatum, error)
	GetDataByPeriod(ctx context.Context, p service.GetDataByPeriodParams) ([]db.DailyDatum, error)
	GetAllData(ctx context.Context) ([]db.DailyDatum, error)
}

type DailyDataController struct {
	service DailyDataQuerier
}

func NewDailyDataController(svc DailyDataQuerier) *DailyDataController {
	return &DailyDataController{service: svc}
}

type GetDataByDayRequest struct {
	DateStr string `form:"date" binding:"required"`
}

func (r *GetDataByDayRequest) toSvcParams() (service.GetDataByDayParams, error) {
	targetTime, err := time.Parse(dailyData.DateFormat, r.DateStr)
	if err != nil {
		return service.GetDataByDayParams{}, err
	}

	return service.GetDataByDayParams{TargetDate: targetTime}, nil
}

// GetDataByDay HTTP GET /daily_data/day?date=2026-5-1
func (c *DailyDataController) GetDataByDay(ctx *gin.Context) {
	var req GetDataByDayRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error(ErrEmptyDate))
		return
	}

	params, err := req.toSvcParams()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error(ErrInvalidDateForm))
		return
	}

	data, err := c.service.GetDataByDay(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, data)
}

type GetDataByPeriodRequest struct {
	StartDateStr string `form:"start" binding:"required"`
	EndDateStr   string `form:"end" binding:"required"`
}

func (r *GetDataByPeriodRequest) toSvcParams() (service.GetDataByPeriodParams, error) {
	start, err1 := time.Parse(dailyData.DateFormat, r.StartDateStr)
	end, err2 := time.Parse(dailyData.DateFormat, r.EndDateStr)
	if err1 != nil || err2 != nil {
		return service.GetDataByPeriodParams{}, ErrInvalidDateForm
	}
	return service.GetDataByPeriodParams{StartDate: start, EndDate: end}, nil
}

// GetDataByPeriod HTTP GET /daily_data/period?start=2026-5-1&end=2026-5-20
func (c *DailyDataController) GetDataByPeriod(ctx *gin.Context) {
	var req GetDataByPeriodRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error(ErrEmptyDate))
		return
	}

	params, err := req.toSvcParams()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error(err))
		return
	}

	data, err := c.service.GetDataByPeriod(ctx, params)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, data)
}

// GetAllData HTTP GET /daily_data/all
func (c *DailyDataController) GetAllData(ctx *gin.Context) {
	data, err := c.service.GetAllData(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error(err))
		return
	}

	ctx.JSON(http.StatusOK, data)
}
