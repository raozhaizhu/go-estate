package delivery

import (
	"github.com/gin-gonic/gin"
	dailyDataController "github.com/raozhaizhu/go-estate/internal/controller/daily_data"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

func RegisterDailyData(authGroup *gin.RouterGroup, service dailyDataController.Service) {
	if service == nil {
		return
	}
	// 初始化 controller
	controller := dailyDataController.NewDailyDataController(service)

	// 初始化路由组
	dailyGroup := authGroup.Group("/daily_data")

	{
		// 至少是 User 才可以获取单日数据
		dailyGroup.GET("/day", middleware.RequireRoles(role.RoleAtLeastUser), response.Wrapper(controller.GetDataByDay))
		// 至少是 VIP 才可以获取范围数据
		dailyGroup.GET("/period", middleware.RequireRoles(role.RoleAtLeastVip), response.Wrapper(controller.GetDataByPeriod))
		// 至少是 Admin 才可以获取所有数据
		dailyGroup.GET("/all", middleware.RequireRoles(role.RoleAtLeastAdmin), response.Wrapper(controller.GetAllData))
	}

}
