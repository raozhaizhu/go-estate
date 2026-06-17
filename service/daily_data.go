package service

import (
	"context"
	"fmt"
	"time"

	db "github.com/raozhaizhu/go-estate/db/sqlc"
	"github.com/raozhaizhu/go-estate/util"
)

type DailyDataService struct {
	store DailyDataQuerier
}

type DailyDataQuerier interface {
	GetDataByDay(ctx context.Context, targetDate time.Time) ([]db.DailyDatum, error)
	GetDataByPeriod(ctx context.Context, arg db.GetDataByPeriodParams) ([]db.DailyDatum, error)
	GetAllData(ctx context.Context) ([]db.DailyDatum, error)
}

func NewDailyDataService(store DailyDataQuerier) *DailyDataService {
	return &DailyDataService{store: store}
}

// GetDataByDay 按天查询数据
func (srv *DailyDataService) GetDataByDay(ctx context.Context, targetDate time.Time) ([]db.DailyDatum, error) {
	// 查询时间必须在范围内
	if targetDate.Before(util.MinDate) || !targetDate.Before(util.ExpiredDate) {
		return []db.DailyDatum{}, fmt.Errorf("%w, 请求日期为:%v 合规范围为:[%v, %v)",
			ErrTimeOutOfRange, targetDate, util.MinDate, util.ExpiredDate)
	}

	return srv.store.GetDataByDay(ctx, targetDate)
}

// GetDataByPeriod 按范围查询数据
func (srv *DailyDataService) GetDataByPeriod(ctx context.Context, startDate, endDate time.Time) ([]db.DailyDatum, error) {
	// 开始时间必须晚于结束时间
	if startDate.After(endDate) {
		return []db.DailyDatum{}, ErrBadTimerOrder
	}
	// 查询时间必须在范围内
	if startDate.Before(util.MinDate) || !endDate.Before(util.ExpiredDate) {
		return []db.DailyDatum{}, fmt.Errorf("%w, 请求范围为:[%v,%v] 合规范围为:[%v, %v)",
			ErrTimeOutOfRange, startDate, endDate, util.MinDate, util.ExpiredDate)
	}

	arg := db.GetDataByPeriodParams{
		StartDate: startDate,
		EndDate:   endDate,
	}
	return srv.store.GetDataByPeriod(ctx, arg)
}

// GetAllData 获取所有数据
func (srv *DailyDataService) GetAllData(ctx context.Context) ([]db.DailyDatum, error) {
	return srv.store.GetAllData(ctx)
}
