package server

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/hibiken/asynq"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
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
	cache      cache.Cache
	redis      *redis.Client
	tokenMaker token.Maker
	router     *gin.Engine
	taskClient *asynq.Client
}

func NewServer(config util.Config, store db.Store, redisCache cache.Cache, asynqClnt *asynq.Client) (*Server, error) {
	// 初始化 JWTMaker
	tokenMaker, err := token.NewJwtMaker(config.TokenSymmetricKey)
	if err != nil {
		log.Fatal("tokenMaker初始化失败: %w", err)
	}

	// 初始化服务
	authSvc := auth.New(store, redisCache, config, tokenMaker, asynqClnt)
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
		cache:      redisCache,
		tokenMaker: tokenMaker,
		router:     router,
		taskClient: asynqClnt,
	}

	return server, nil
}

func (srv *Server) Start(address string) error {
	return srv.router.Run(address)
}
