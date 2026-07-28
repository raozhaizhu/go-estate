package main

import (
	"log/slog"
	"os"

	_ "github.com/raozhaizhu/go-estate/docs"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	"github.com/raozhaizhu/go-estate/internal/server"
	"github.com/raozhaizhu/go-estate/internal/util"

	_ "github.com/go-sql-driver/mysql"
)

// @title           Go Estate API 文档
// @version         1.0
// @description     房地产项目的后端 API 接口文档
// @host            localhost
// @BasePath        /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @securityDefinitions.apikey DeviceIDAuth
// @in                         header
// @name                       X-Device-ID

// @securityDefinitions.apikey UserAgentAuth
// @in                         header
// @name                       User-Agent
func main() {
	config := util.InitConfig(".")
	deps := app.PrepareDeps(config)
	lifecycleLogger := slog.Default().With("layer", "main", "category", "lifecycle")

	// 初始化服务器
	srv, err := server.NewServer(deps)

	if err != nil {
		lifecycleLogger.Error("服务器初始化失败", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 运行服务器在指定端口
	err = srv.Start(":" + deps.Config.ServerPort)
	if err != nil {
		lifecycleLogger.Error("服务器运行失败", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
