package dailyData

import (
	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	dailyData "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

func RegisterDailyData(r *gin.RouterGroup, store db.Store) {
	svc := dailyData.NewDailyDataService(store)
	ctrl := NewDailyDataController(svc)

	g := r.Group("/daily_data")
	{
		g.GET("/day", response.Wrapper(ctrl.GetDataByDay))
		g.GET("/period", response.Wrapper(ctrl.GetDataByPeriod))
		g.GET("/all", response.Wrapper(ctrl.GetAllData))
	}
}
