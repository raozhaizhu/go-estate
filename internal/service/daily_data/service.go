package dailyData

import (
	"context"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
)

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 */

const (
	// 点数消费
	GetDataByDayPoints    = 1
	GetDataByPeriodPoints = 5

	// 过期时间
	GetDataExpireDuration = 12 * time.Hour
)

// GetDataByDay 按日获取楼盘成交数据
func (svc *service) GetDataByDay(ctx context.Context, input GetDataByDayInput) ([]db.DailyDatum, error) {
	// 参数转换
	txParams, err := input.toDBParams(ctx)
	if err != nil {
		return nil, err
	}
	cacheKey := input.TargetDate.String()

	// -> db 获取数据
	return svc.fetchAndDeductData(ctx, cacheKey, txParams)
}

/** ====================================================================================
 * 🏁 GetDataByPeriod
 * =====================================================================================
 */

// GetDataByPeriod 按周期获取楼盘成交数据
func (svc *service) GetDataByPeriod(ctx context.Context, input GetDataByPeriodInput) ([]db.DailyDatum, error) {
	// 参数转换
	txParams, err := input.toDBParams(ctx)
	if err != nil {
		return nil, err
	}
	cacheKey := txParams.StartDate.String() + txParams.EndDate.String()

	// -> db 获取数据
	return svc.fetchAndDeductData(ctx, cacheKey, txParams)
}

/** ====================================================================================
 * 🏁 GetAllData
 * =====================================================================================
 */

// GetAllData 获取所有楼盘成交数据
// 该方法仅有 admin 可以调动, 不需要做任何验证和缓存策略
func (svc *service) GetAllData(ctx context.Context) ([]db.DailyDatum, error) {
	// -> db 获取数据
	return svc.store.GetAllData(ctx)
}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

// fetchAndDeductData 获取数据并尝试扣减用户积分, 若扣费成功将通知用户
func (svc *service) fetchAndDeductData(ctx context.Context, cacheKey string, txParams db.GetDataTxParams) ([]db.DailyDatum, error) {
	// -> cache 若命中则直接返回数据
	var cachedData []db.DailyDatum
	hit, err := svc.queryCache.GetDailyData(ctx, txParams.Username, cacheKey, &cachedData)
	if err != nil { // Redis 出现内部错误
		svc.logger.ErrorContext(ctx, "redis 查询失败", "err", err.Error())
	} else if hit { // Redis 没有报错且命中
		return cachedData, nil
	}

	// -> db 插入凭证 & 扣费 & 获取数据
	txResult, err := svc.store.GetDataAndDeductPointsTx(ctx, txParams)
	if err != nil {
		return nil, err
	}

	// -> cache 写入缓存
	_ = svc.queryCache.SetDailyData(ctx, txParams.Username, cacheKey, txResult.Data, GetDataExpireDuration)

	// -> ws 扣费成功,通知用户
	if txResult.FirstTime {
		svc.wsManager.GoSendPointsMsgToUser(ctx, txParams.Username, GetDataByDayPoints)
	}

	return txResult.Data, err
}
