package server

import (
	"log"

	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
	dailyData "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/pkg/token"
	"github.com/raozhaizhu/go-estate/pkg/validator"
)

type Server struct {
	config     util.Config
	store      db.Store
	tokenMaker token.Maker
	router     *gin.Engine
}

func NewServer(config util.Config, store db.Store) (*Server, error) {
	// 初始化 JWTMaker
	tokenMaker, err := token.NewJwtMaker(config.TokenSymmetricKey)
	if err != nil {
		log.Fatal("tokenMaker初始化失败: %w", err)
	}

	// 初始化服务
	authSvc := auth.New(store, config, tokenMaker)
	userSvc := user.New(store)
	dailyDataSvc := dailyData.New(store)
	services := delivery.Services{
		UserSvc:      userSvc,
		AuthSvc:      authSvc,
		DailyDataSvc: dailyDataSvc,
	}

	// 初始化路由
	router := delivery.SetupRouter(services, config, tokenMaker)

	// 初始化验证翻译器
	validator.InitTrans()

	// 初始化服务器
	server := &Server{
		config:     config,
		store:      store,
		tokenMaker: tokenMaker,
		router:     router,
	}

	return server, nil
}

func (srv *Server) Start(address string) error {
	return srv.router.Run(address)
}
