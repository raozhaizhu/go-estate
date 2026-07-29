package db

import (
	"context"
	"fmt"
	"time"

	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 GetDataAndDeductPointsTx
 * =====================================================================================
 */

const (
	// 查询类型
	TypeGetDataByDay    int8 = 1
	TypeGetDataByPeriod int8 = 2
)

// GetDataTxParams 获取成交数据所需参数
type GetDataTxParams struct {
	Username  string
	QueryType int8
	StartDate time.Time
	EndDate   time.Time
	Points    int
}

// GetDataTxResult 返回成交数据 & 是否初次查询
type GetDataTxResult struct {
	Data      []DailyDatum
	FirstTime bool
}

// GetDataAndDeductPointsTx 执行扣费并获取数据的完整事务
func (s *SQLStore) GetDataAndDeductPointsTx(ctx context.Context, arg GetDataTxParams) (GetDataTxResult, error) {
	var result GetDataTxResult

	err := s.ExecTx(ctx, func(q Querier) error {
		// 1. 尝试写入查询凭证
		firstTime, err := TryRecordUserQuery(q, ctx, GetUserQueryRecordParams{
			Username:  arg.Username,
			QueryType: arg.QueryType,
			StartDate: arg.StartDate,
			EndDate:   arg.EndDate,
		})
		if err != nil {
			return err
		}

		result.FirstTime = firstTime // 记录到外层返回值中

		// 2. 写入凭证成功，尝试扣费
		if firstTime {
			if err = DecreaseUserPoints(q, ctx, DecreasePointsParams{
				Amount:   uint32(arg.Points),
				Username: arg.Username,
			}); err != nil {
				return err
			}
		}

		// 3. 获取数据
		var fetchedData []DailyDatum
		switch arg.QueryType {
		case TypeGetDataByDay: // 获取单日数据
			fetchedData, err = q.GetDataByDay(ctx, arg.StartDate)
			if err != nil {
				return err
			}

		case TypeGetDataByPeriod:
			fetchedData, err = q.GetDataByPeriod(ctx, GetDataByPeriodParams{
				StartDate: arg.StartDate,
				EndDate:   arg.EndDate,
			})
			if err != nil {
				return err
			}

		default:
			return appError.NewSrvErr(fmt.Errorf("GetDataAndDeductPointsTx 抵达了不应抵达的位置"))

		}

		if len(fetchedData) == 0 { // 用户没有获取到任何数据, 当日没有录入数据或者没有成交(查询过早)
			return appError.ErrDailyDataNotFound
		}

		result.Data = fetchedData
		return nil
	})

	return result, err
}

// DecreaseUserPoints
func DecreaseUserPoints(q Querier, ctx context.Context, arg DecreasePointsParams) error {
	result, err := q.DecreasePoints(ctx, arg)
	if err != nil {
		return appError.NewSrvErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return appError.NewSrvErr(err)
	}

	if rowsAffected == 0 {
		return appError.ErrInsufficientPoints
	}

	return nil
}

// TryRecordUserQuery 尝试记录用户的查询行为
// true 插入成功, 这是用户第一次查询该日期段的数据, 应该扣费
// false 插入失败, 用户之前已经查询过, 不应该扣费
func TryRecordUserQuery(q Querier, ctx context.Context, arg GetUserQueryRecordParams) (bool, error) {
	result, err := q.GetUserQueryRecord(ctx, arg)
	if err != nil {
		return false, appError.NewSrvErr(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, appError.NewSrvErr(err)
	}

	if rowsAffected > 0 { // 成功插入, 说明这是用户第一次查询, 需要扣费
		return true, nil
	}

	return false, nil
}
