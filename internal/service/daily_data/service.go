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
	txParams, err := input.toDBParams(ctx)
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
	txParams, err := input.toDBParams(ctx)
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
	startDate, endDate := txParams.StartDate.Format(dailyDataDomain.DateFormat), txParams.EndDate.Format(dailyDataDomain.DateFormat)
	// -> cache 若命中则直接返回数据
	recordHit, rawData, err := svc.queryCache.GetRecordAndData(ctx, txParams.Username, startDate, endDate)
	if err != nil { // Redis 出现内部错误
		svc.logger.ErrorContext(ctx, "redis 查询失败", "err", err.Error())
	} else if recordHit && rawData != nil { // Redis 没有报错, 得到用户记录, 且得到成交数据, 直接返回即可
		var cachedData []db.DailyDatum
		err := json.Unmarshal(rawData, &cachedData)
		if err != nil { // 序列化失败, 继续往下走访问 db 获取数据
			svc.logger.ErrorContext(ctx, "序列化数据失败", "err", err.Error())
		} else {
			return cachedData, nil
		}
	}

	// 记录 record 是否命中缓存(若命中则仅需获取成交数据即可)
	txParams.RecordHit = recordHit
	// -> db 插入凭证(?) & 扣费(?) & 获取数据(必须)
	txResult, err := svc.store.GetDataAndDeductPointsTx(ctx, txParams)
	if err != nil {
		return nil, err
	}

	// -> cache 写入缓存
	err = svc.queryCache.SetRecordAndData(ctx, txParams.Username, startDate, endDate, txResult.Data, RecordExpireDuration, DataExpireDuration)
	if err != nil {
		svc.logger.ErrorContext(ctx, "写入缓存失败", "err", err.Error())
	}

	// -> ws 是初次查询, 因此产生了扣费, 通知用户
	if !recordHit && txResult.FirstTime { // 缓存没命中, 并且查数据得知这是初次查询
		svc.wsManager.GoSendPointsMsgToUser(ctx, txParams.Username, txParams.Points)
	}

	return txResult.Data, err
}
