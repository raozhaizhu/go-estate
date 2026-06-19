package dailyData

import (
	"context"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
)

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

type GetDataByDayParams struct {
	TargetDate time.Time
}

func (p *GetDataByDayParams) ToDBParams() (time.Time, error) {
	// 查询时间必须在范围内
	if p.TargetDate.Before(dailyData.MinDate) || !p.TargetDate.Before(dailyData.ExpiredDate) {
		return time.Time{}, ErrTimeOutOfRange
	}

	return p.TargetDate, nil
}

// GetDataByDay 按天查询数据
func (svc *DailyDataService) GetDataByDay(ctx context.Context, p GetDataByDayParams) ([]db.DailyDatum, error) {
	targetDate, err := p.ToDBParams()
	if err != nil {
		return []db.DailyDatum{}, err
	}

	return svc.db.GetDataByDay(ctx, targetDate)
}

type GetDataByPeriodParams struct {
	StartDate time.Time
	EndDate   time.Time
}

func (p *GetDataByPeriodParams) ToDBParams() (db.GetDataByPeriodParams, error) {
	// 开始时间必须晚于结束时间
	if p.StartDate.After(p.EndDate) {
		return db.GetDataByPeriodParams{}, ErrBadTimerOrder
	}
	// 查询时间必须在范围内
	if p.StartDate.Before(dailyData.MinDate) || !p.EndDate.Before(dailyData.ExpiredDate) {
		return db.GetDataByPeriodParams{}, ErrTimeOutOfRange
	}

	params := db.GetDataByPeriodParams{
		StartDate: p.StartDate,
		EndDate:   p.EndDate,
	}

	return params, nil
}

// GetDataByPeriod 按范围查询数据
func (svc *DailyDataService) GetDataByPeriod(ctx context.Context, p GetDataByPeriodParams) ([]db.DailyDatum, error) {
	params, err := p.ToDBParams()
	if err != nil {
		return []db.DailyDatum{}, err
	}
	return svc.db.GetDataByPeriod(ctx, params)
}

// GetAllData 获取所有数据
func (svc *DailyDataService) GetAllData(ctx context.Context) ([]db.DailyDatum, error) {
	return svc.db.GetAllData(ctx)
}
