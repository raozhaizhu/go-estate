package dailyData

import (
	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	dailyData "github.com/raozhaizhu/go-estate/internal/service/daily_data"
)

func RegisterDailyData(r *gin.RouterGroup, store db.Store) {
	svc := dailyData.NewDailyDataService(store)
	ctrl := NewDailyDataController(svc)

	g := r.Group("/daily_data")
	{
		g.GET("/day", ctrl.GetDataByDay)
		g.GET("/period", ctrl.GetDataByPeriod)
		g.GET("/all", ctrl.GetAllData)
	}
}
