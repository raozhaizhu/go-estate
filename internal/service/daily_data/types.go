package dailyData

import (
	"context"
	"log/slog"
	"time"

	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	"github.com/raozhaizhu/go-estate/internal/myWebsocket"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	ctxKey "github.com/raozhaizhu/go-estate/pkg/ctx_key"
)

/** ====================================================================================
 * 🏁 DailyDataService
 * =====================================================================================
 *
 */

type service struct {
	store      db.DailyDataStore
	txRunner   db.TxRunner
	queryCache cache.QueryCache
	logger     *slog.Logger
	wsManager  *myWebsocket.Manager
}

func New(deps *app.Deps) *service {
	logger := slog.Default().With("layer", "service", "module", "dailyData_service")
	return &service{store: deps.Store, txRunner: deps.Store, logger: logger, wsManager: deps.WebsocketManager}
}

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 *
 */

type GetDataByDayInput struct {
	TargetDate time.Time
}

func (input *GetDataByDayInput) toDBParams(ctx context.Context) (db.GetDataTxParams, error) {
	// 获取用户名
	username, err := ctxKey.GetUsername(ctx)
	if err != nil {
		return db.GetDataTxParams{}, err
	}
	// 查询时间必须在范围内
	if input.TargetDate.Before(dailyData.MinDate) || !input.TargetDate.Before(dailyData.ExpiredDate) {
		return db.GetDataTxParams{}, appError.ErrTimeOutOfRange
	}

	txParams := db.GetDataTxParams{
		Username:  username,
		StartDate: input.TargetDate,
		EndDate:   input.TargetDate,
		Points:    GetDataByDayPoints,
		QueryType: db.TypeGetDataByDay,
	}

	return txParams, nil
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

func (input *GetDataByPeriodInput) toDBParams(ctx context.Context) (db.GetDataTxParams, error) {
	// 获取用户名
	username, err := ctxKey.GetUsername(ctx)
	if err != nil {
		return db.GetDataTxParams{}, err
	}

	// 开始时间必须晚于结束时间
	if input.StartDate.After(input.EndDate) {
		return db.GetDataTxParams{}, appError.ErrBadTimerOrder
	}
	// 查询时间必须在范围内
	if input.StartDate.Before(dailyData.MinDate) || !input.EndDate.Before(dailyData.ExpiredDate) {
		return db.GetDataTxParams{}, appError.ErrTimeOutOfRange
	}

	params := db.GetDataTxParams{
		Username:  username,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		Points:    GetDataByDayPoints,
		QueryType: db.TypeGetDataByDay,
	}

	return params, nil
}
