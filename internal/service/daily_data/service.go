package dailyData

import (
	"context"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
)

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 */

// GetDataByDay 按日获取楼盘成交数据
func (svc *service) GetDataByDay(ctx context.Context, input GetDataByDayInput) ([]db.DailyDatum, error) {
	// 参数转换
	targetDate, err := input.toDBParams()
	if err != nil {
		return nil, err
	}
	// -> db 获取数据
	return svc.store.GetDataByDay(ctx, targetDate)
}

/** ====================================================================================
 * 🏁 GetDataByPeriod
 * =====================================================================================
 */

// GetDataByPeriod 按周期获取楼盘成交数据
func (svc *service) GetDataByPeriod(ctx context.Context, input GetDataByPeriodInput) ([]db.DailyDatum, error) {
	// 参数转换
	params, err := input.toDBParams()
	if err != nil {
		return nil, err
	}
	// -> db 获取数据
	return svc.store.GetDataByPeriod(ctx, params)
}

/** ====================================================================================
 * 🏁 GetAllData
 * =====================================================================================
 */

// GetAllData 获取所有楼盘成交数据
func (svc *service) GetAllData(ctx context.Context) ([]db.DailyDatum, error) {
	// -> db 获取数据
	return svc.store.GetAllData(ctx)
}
