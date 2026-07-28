package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
	dailyData "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/internal/worker"
	"github.com/raozhaizhu/go-estate/pkg/validator"
)

type Server struct {
	config        util.Config
	store         db.Store
	cache         cache.Cache
	router        *gin.Engine
	taskProcessor worker.TaskProcessor
}

// NewServer 创建服务器
func NewServer(deps *app.Deps) (*Server, error) {
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
		config:        deps.Config,
		store:         deps.Store,
		cache:         deps.Cache,
		router:        router,
		taskProcessor: deps.TaskProcessor,
	}

	return server, nil
}

// Start 启动服务器, 并监听退出信号以实现优雅退出
func (srv *Server) Start(address string) error {
	lifecycleLogger := slog.Default().With(slog.String("category", "lifecycle"))
	// 构造 server
	httpSrv := &http.Server{
		Addr:    address,
		Handler: srv.router,
	}

	// 启动协程进行监听和服务
	go func() {
		lifecycleLogger.Info("服务器启动", slog.String("address", address))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lifecycleLogger.Error("HTTP server 异常退出", slog.String("error", err.Error()))
		}
	}()

	if srv.taskProcessor != nil {
		go func() {
			lifecycleLogger.Info("Asynq Worker 服务器启动...")
			if err := srv.taskProcessor.Start(); err != nil {
				lifecycleLogger.Error("Asynq Worker 异常退出", slog.String("error", err.Error()))
			}
		}()
	}

	// 监听系统 kill 信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	lifecycleLogger.Info("收到停机信号, 准备退出")

	// 关闭 HTTP 入口, 禁止新请求进入
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpSrv.Shutdown(ctx); err != nil {
		return fmt.Errorf("服务器强制退出: %w", err)
	}

	// 关闭后台异步任务处理器
	if srv.taskProcessor != nil {
		lifecycleLogger.Info("正在停止 Asynq Worker 服务器...")
		srv.taskProcessor.Stop() // 阻塞直到当前正在跑的 Job 执行完毕
		lifecycleLogger.Info("Asynq Worker 服务器已安全退出")
	}

	// 关闭数据库连接
	lifecycleLogger.Info("正在关闭数据库连接...")
	if err := srv.store.Close(); err != nil {
		lifecycleLogger.Error("关闭数据库连接失败", slog.String("error", err.Error()))
	} else {
		lifecycleLogger.Info("关闭数据库连接成功")
	}

	// 关闭缓存连接
	lifecycleLogger.Info("正在关闭缓存连接...")
	if err := srv.cache.Close(); err != nil {
		lifecycleLogger.Error("关闭缓存连接失败", slog.String("error", err.Error()))
	} else {
		lifecycleLogger.Info("关闭缓存连接成功")
	}

	// 安全退出
	lifecycleLogger.Info("HTTP服务器已安全退出")
	return nil
}

// ServeHTTP 用于集成测试
func (srv *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	srv.router.ServeHTTP(w, req)
}
