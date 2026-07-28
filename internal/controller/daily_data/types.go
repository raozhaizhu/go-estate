package dailyData

import (
	"context"
	"log/slog"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	service "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 DailyDataController
 * =====================================================================================
 */

type Service interface {
	GetDataByDay(ctx context.Context, p service.GetDataByDayInput) ([]db.DailyDatum, error)
	GetDataByPeriod(ctx context.Context, p service.GetDataByPeriodInput) ([]db.DailyDatum, error)
	GetAllData(ctx context.Context) ([]db.DailyDatum, error)
}

type Controller struct {
	service Service
	logger  *slog.Logger
}

func NewDailyDataController(svc Service) *Controller {
	logger := slog.Default().With("layer", "controller", "module", "dailyData_controller")
	return &Controller{service: svc, logger: logger}
}

type DailyDataList []DailyDatumSchema

type DailyDatumSchema struct {
	ID          int32     `json:"id" example:"1"`
	Date        time.Time `json:"date" example:"2026-05-01"`
	Region      int16     `json:"region" example:"1"`
	Category    string    `json:"category" example:"住宅"`
	LicenseNo   string    `json:"license_no" example:"温房预许字（2026）第00051号"`
	ProjectName string    `json:"project_name" example:"臻玉园(三期)"`
	HouseCount  int16     `json:"house_count" example:"1"`
	Area        string    `json:"area" example:"1003.96"`
	AvgPrice    string    `json:"avg_price" example:"14331.00"`
}

func toResponse(list []db.DailyDatum) DailyDataList {
	var dailyDataList DailyDataList

	for _, item := range list {
		// 清洗: 只有 Valid 为 true 时才取值
		avgPriceStr := ""
		if item.AvgPrice.Valid {
			avgPriceStr = item.AvgPrice.String
		}

		dailyDataList = append(dailyDataList, DailyDatumSchema{
			ID:          item.ID,
			Date:        item.Date,
			Region:      item.Region,
			Category:    item.Category,
			LicenseNo:   item.LicenseNo,
			ProjectName: item.ProjectName,
			HouseCount:  item.HouseCount,
			Area:        item.Area,
			AvgPrice:    avgPriceStr,
		})
	}

	return dailyDataList
}

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 */

type GetDataByDayRequest struct {
	Date string `form:"date" binding:"required,datetime=2006-01-02" example:"2026-05-01"`
}

func (r *GetDataByDayRequest) toSvcInput() service.GetDataByDayInput {
	// 转换为标准日期字符串
	targetTime, _ := time.Parse(dailyData.DateFormat, r.Date)

	return service.GetDataByDayInput{TargetDate: targetTime}
}

/** ====================================================================================
 * 🏁 GetDataByPeriod
 * =====================================================================================
 */

type GetDataByPeriodRequest struct {
	Start string `form:"start" binding:"required,datetime=2006-01-02" example:"2026-05-01"`
	End   string `form:"end" binding:"required,datetime=2006-01-02" example:"2026-05-02"`
}

func (r *GetDataByPeriodRequest) toSvcInput() (service.GetDataByPeriodInput, error) {
	start, _ := time.Parse(dailyData.DateFormat, r.Start)
	end, _ := time.Parse(dailyData.DateFormat, r.End)

	if end.Before(start) {
		return service.GetDataByPeriodInput{}, appError.ErrBadTimerOrder
	}

	return service.GetDataByPeriodInput{StartDate: start, EndDate: end}, nil
}

/** ====================================================================================
 * 🏁 GetAllData
 * =====================================================================================
 */
