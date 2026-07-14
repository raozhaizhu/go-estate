package server

import (
	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
	dailyData "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/pkg/validator"
)

type Server struct {
	config util.Config
	router *gin.Engine
}

func NewServer(deps app.Deps) (*Server, error) {
	// 初始化服务
	authSvc := auth.New(deps)
	userSvc := user.New(deps)
	dailyDataSvc := dailyData.New(deps)
	services := delivery.Services{
		UserSvc:      userSvc,
		AuthSvc:      authSvc,
		DailyDataSvc: dailyDataSvc,
	}

	// 初始化路由
	router := delivery.SetupRouter(services, deps)

	// 初始化验证翻译器
	validator.InitTrans()

	// 初始化服务器
	server := &Server{
		config: deps.Config,
		router: router,
	}

	return server, nil
}

func (srv *Server) Start(address string) error {
	return srv.router.Run(address)
}
