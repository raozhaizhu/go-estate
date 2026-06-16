package router

import (
	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/controller"
)

func Setup(r *gin.Engine, ctrl *controller.DailyDataController) {
	api := r.Group("/api/v1")
	{
		api.GET("/daily_data/day", ctrl.GetDataByDay)
		api.GET("/daily_data/period", ctrl.GetDataByPeriod)
		api.GET("/daily_data/all", ctrl.GetAllData)
	}
}
