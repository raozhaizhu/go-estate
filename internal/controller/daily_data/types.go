package dailyData

import (
	"context"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
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

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 */

type GetDataByDayRequest struct {
	Date string `form:"date" binding:"required,datetime=2006-01-02"`
}

func (r *GetDataByDayRequest) toSvcInput() (service.GetDataByDayInput, error) {
	// 转换为标准日期字符串
	targetTime, err := time.Parse(dailyData.DateFormat, r.Date)
	// 已知错误: 查询日期格式错误
	if err != nil {
		return service.GetDataByDayInput{}, appError.ErrBadDate
	}

	return service.GetDataByDayInput{TargetDate: targetTime}, nil
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
	// 转换为标准日期字符串
	// 已知错误: 开始日期格式错误
	start, err := time.Parse(dailyData.DateFormat, r.Start)
	if err != nil {
		return service.GetDataByPeriodInput{}, appError.ErrBadStartDate
	}
	// 已知错误: 结束日期格式错误
	end, err := time.Parse(dailyData.DateFormat, r.End)
	if err != nil {
		return service.GetDataByPeriodInput{}, appError.ErrBadEndDate
	}
	return service.GetDataByPeriodInput{StartDate: start, EndDate: end}, nil
}

/** ====================================================================================
 * 🏁 GetAllData
 * =====================================================================================
 */
