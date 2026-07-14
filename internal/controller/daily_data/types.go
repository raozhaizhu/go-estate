package dailyData

import (
	"context"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	service "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 DailyDataController
 * =====================================================================================
 */

type Service interface {
	GetDataByDay(ctx context.Context, p service.GetDataByDayInput) ([]db.DailyDatum, error)
	GetDataByPeriod(ctx context.Context, p service.GetDataByPeriodInput) ([]db.DailyDatum, error)
	GetAllData(ctx context.Context) ([]db.DailyDatum, error)
}

type Controller struct {
	service Service
}

func NewDailyDataController(svc Service) *Controller {
	return &Controller{service: svc}
}

type DailyDataList []db.DailyDatum

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 */

type GetDataByDayRequest struct {
	Date string `form:"date" binding:"required,datetime=2006-01-02"`
}

func (r *GetDataByDayRequest) toSvcInput() service.GetDataByDayInput {
	// 转换为标准日期字符串
	targetTime, _ := time.Parse(dailyData.DateFormat, r.Date)

	return service.GetDataByDayInput{TargetDate: targetTime}
}

/** ====================================================================================
 * 🏁 GetDataByPeriod
 * =====================================================================================
 */

type GetDataByPeriodRequest struct {
	Start string `form:"start" binding:"required,datetime=2006-01-02"`
	End   string `form:"end" binding:"required,datetime=2006-01-02"`
}

func (r *GetDataByPeriodRequest) toSvcInput() (service.GetDataByPeriodInput, error) {
	start, _ := time.Parse(dailyData.DateFormat, r.Start)
	end, _ := time.Parse(dailyData.DateFormat, r.End)

	if end.Before(start) {
		return service.GetDataByPeriodInput{}, appError.ErrBadTimerOrder
	}

	return service.GetDataByPeriodInput{StartDate: start, EndDate: end}, nil
}

/** ====================================================================================
 * 🏁 GetAllData
 * =====================================================================================
 */
