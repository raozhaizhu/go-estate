package dailyData

import (
	"context"

	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
)

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 *
 */

// GetDataByDay 按天查询成交数据
func (svc *DailyDataService) GetDataByDay(ctx context.Context, input GetDataByDayInput) ([]db.DailyDatum, error) {
	// 参数转换
	targetDate, err := input.toDBParams()
	if err != nil {
		return []db.DailyDatum{}, err
	}
	// -> db 获取数据
	return svc.db.GetDataByDay(ctx, targetDate)
}

/** ====================================================================================
 * 🏁 GetDataByPeriod
 * =====================================================================================
 *
 */

// GetDataByPeriod 按范围查询成交数据
func (svc *DailyDataService) GetDataByPeriod(ctx context.Context, input GetDataByPeriodInput) ([]db.DailyDatum, error) {
	// 参数转换
	params, err := input.toDBParams()
	if err != nil {
		return []db.DailyDatum{}, err
	}
	// -> db 获取数据
	return svc.db.GetDataByPeriod(ctx, params)
}

/** ====================================================================================
 * 🏁 GetAllData
 * =====================================================================================
 *
 */

// GetAllData 获取所有成交数据
func (svc *DailyDataService) GetAllData(ctx context.Context) ([]db.DailyDatum, error) {
	// -> db 获取数据
	return svc.db.GetAllData(ctx)
}
