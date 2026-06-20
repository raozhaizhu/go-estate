package dailyData

import (
	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	dailyData "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

func RegisterDailyData(protectedGroup *gin.RouterGroup, store db.Store) {
	// 初始化service controller
	svc := dailyData.NewDailyDataService(store)
	ctrl := NewDailyDataController(svc)

	// 初始化路由组
	dailyGroup := protectedGroup.Group("/daily_data")

	{
		// 至少是 User 才可以获取单日数据
		dailyGroup.GET("/day", middleware.RoleMiddleware(role.RoleAtLeastUser), response.Wrapper(ctrl.GetDataByDay))
		// 至少是 VIP 才可以获取范围数据
		dailyGroup.GET("/period", middleware.RoleMiddleware(role.RoleAtLeastVip), response.Wrapper(ctrl.GetDataByPeriod))
		// 至少是 Admin 才可以获取所有数据
		dailyGroup.GET("/all", middleware.RoleMiddleware(role.RoleAtLeastAdmin), response.Wrapper(ctrl.GetAllData))
	}

}
