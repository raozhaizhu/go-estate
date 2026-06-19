package dailyData

import (
	"context"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	appError "github.com/raozhaizhu/go-estate/pkg/apperror"
)

/** ====================================================================================
 * 🏁 DailyDataService
 * =====================================================================================
 *
 */

type DailyDataService struct {
	db DailyDataQuerier
}

type DailyDataQuerier interface {
	GetDataByDay(ctx context.Context, targetDate time.Time) ([]db.DailyDatum, error)
	GetDataByPeriod(ctx context.Context, arg db.GetDataByPeriodParams) ([]db.DailyDatum, error)
	GetAllData(ctx context.Context) ([]db.DailyDatum, error)
}

func NewDailyDataService(store DailyDataQuerier) *DailyDataService {
	return &DailyDataService{db: store}
}

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 *
 */

type GetDataByDayInput struct {
	TargetDate time.Time
}

func (input *GetDataByDayInput) toDBParams() (time.Time, error) {
	// 查询时间必须在范围内
	if input.TargetDate.Before(dailyData.MinDate) || !input.TargetDate.Before(dailyData.ExpiredDate) {
		return time.Time{}, appError.ErrTimeOutOfRange
	}

	return input.TargetDate, nil
}

/** ====================================================================================
 * 🏁 GetDataByPeriod
 * =====================================================================================
 *
 */

type GetDataByPeriodInput struct {
	StartDate time.Time
	EndDate   time.Time
}

func (input *GetDataByPeriodInput) toDBParams() (db.GetDataByPeriodParams, error) {
	// 开始时间必须晚于结束时间
	if input.StartDate.After(input.EndDate) {
		return db.GetDataByPeriodParams{}, appError.ErrBadTimerOrder
	}
	// 查询时间必须在范围内
	if input.StartDate.Before(dailyData.MinDate) || !input.EndDate.Before(dailyData.ExpiredDate) {
		return db.GetDataByPeriodParams{}, appError.ErrTimeOutOfRange
	}

	params := db.GetDataByPeriodParams{
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
	}

	return params, nil
}
