package server

import (
	"github.com/gin-gonic/gin"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	"github.com/raozhaizhu/go-estate/internal/router"
	"github.com/raozhaizhu/go-estate/internal/util"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
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
		return nil, appError.NewInvalidKeySizeError(len(config.TokenSymmetricKey), token.MinSecretSize)
	}

	// 初始化验证翻译器
	validator.InitTrans()

	// 初始化路由
	router := router.SetupRouter(store, config, tokenMaker)

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
