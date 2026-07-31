package dailyData

import (
	"context"
	"encoding/json"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	dailyDataDomain "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
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
	RecordExpireDuration = 12 * time.Hour
	DataExpireDuration   = 7 * 24 * time.Hour
)

// GetDataByDay 按日获取楼盘成交数据
func (svc *service) GetDataByDay(ctx context.Context, input *GetDataByDayInput) ([]db.DailyDatum, error) {
	// 参数转换
	txParams, err := input.toDBParams()
	if err != nil {
		return nil, err
	}

	// -> db 获取数据
	return svc.fetchAndDeductData(ctx, txParams)
}

/** ====================================================================================
 * 🏁 GetDataByPeriod
 * =====================================================================================
 */

// GetDataByPeriod 按周期获取楼盘成交数据
func (svc *service) GetDataByPeriod(ctx context.Context, input *GetDataByPeriodInput) ([]db.DailyDatum, error) {
	// 参数转换
	txParams, err := input.toDBParams()
	if err != nil {
		return nil, err
	}

	// -> db 获取数据
	return svc.fetchAndDeductData(ctx, txParams)
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
func (svc *service) fetchAndDeductData(ctx context.Context, txParams db.GetDataTxParams) ([]db.DailyDatum, error) {
	// 格式化日期
	startDate, endDate := txParams.StartDate.Format(dailyDataDomain.DateFormat), txParams.EndDate.Format(dailyDataDomain.DateFormat)

	// -> cache 获取缓存情况
	recordHit, rawData, err := svc.queryCache.GetRecordAndData(ctx, txParams.Username, startDate, endDate)
	var cachedData []db.DailyDatum
	var dataHit bool
	if rawData != nil { // 成交数据非空 = DataHit
		if err := json.Unmarshal(rawData, &cachedData); err == nil {
			dataHit = true
		} else {
			svc.logger.ErrorContext(ctx, "序列化缓存数据失败，降级为穿透查DB", "err", err.Error())
		}
	}

	// DoubleHit 直接返回数据
	if recordHit && dataHit {
		return cachedData, nil
	}

	// 记录 Hit 情况, 准备开启事务
	txParams.RecordHit, txParams.DataHit = recordHit, dataHit

	// -> db 执行事务, 按需获取数据
	txResult, err := svc.store.GetDataAndDeductPointsTx(ctx, txParams)
	if err != nil {
		return nil, err
	}

	// -> cache 按需写回缓存
	err = svc.queryCache.SetRecordAndData(ctx, recordHit, dataHit, txParams.Username, startDate, endDate, txResult.Data,
		RecordExpireDuration, DataExpireDuration)
	if err != nil {
		svc.logger.ErrorContext(ctx, "写入缓存失败", "err", err.Error())
	}

	// -> ws 是初次查询, 因此产生了扣费, 通知用户
	if txResult.FirstTime {
		svc.wsManager.GoSendPointsMsgToUser(ctx, txParams.Username, txParams.Points)
	}

	// 组装最终的返回结果
	finalData := cachedData
	if !dataHit {
		finalData = txResult.Data
	}

	return finalData, nil
}
