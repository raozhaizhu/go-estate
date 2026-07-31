package dailyData

import (
	"log/slog"
	"time"

	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	myWebsocket "github.com/raozhaizhu/go-estate/internal/my_websocket"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 DailyDataService
 * =====================================================================================
 *
 */

type service struct {
	store      db.DailyDataStore
	queryCache cache.QueryCache
	wsManager  *myWebsocket.Manager

	logger *slog.Logger
}

func New(deps *app.Deps) *service {
	logger := slog.Default().With("layer", "service", "module", "dailyData_service")
	return &service{store: deps.Store, queryCache: deps.Cache, logger: logger, wsManager: deps.WebsocketManager}
}

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 *
 */

type GetDataByDayInput struct {
	Username   string
	TargetDate time.Time
}

func (input *GetDataByDayInput) toDBParams() (db.GetDataTxParams, error) {
	// 查询时间必须在范围内
	if input.TargetDate.Before(dailyData.MinDate) || !input.TargetDate.Before(dailyData.ExpiredDate) {
		return db.GetDataTxParams{}, appError.ErrTimeOutOfRange
	}

	txParams := db.GetDataTxParams{
		Username:  input.Username,
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
	Username  string
	StartDate time.Time
	EndDate   time.Time
}

func (input *GetDataByPeriodInput) toDBParams() (db.GetDataTxParams, error) {
	// 开始时间必须晚于结束时间
	if !input.StartDate.Before(input.EndDate) {
		return db.GetDataTxParams{}, appError.ErrBadTimerOrder
	}
	// 查询时间必须在范围内
	if input.StartDate.Before(dailyData.MinDate) || !input.EndDate.Before(dailyData.ExpiredDate) {
		return db.GetDataTxParams{}, appError.ErrTimeOutOfRange
	}

	params := db.GetDataTxParams{
		Username:  input.Username,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		Points:    GetDataByDayPoints,
		QueryType: db.TypeGetDataByDay,
	}

	return params, nil
}
