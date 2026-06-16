package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/raozhaizhu/go-estate/controller"
	db "github.com/raozhaizhu/go-estate/db/sqlc"
	"github.com/raozhaizhu/go-estate/router"
	"github.com/raozhaizhu/go-estate/service"
	"github.com/raozhaizhu/go-estate/util"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	r := gin.Default()
	startServer(r)

	if err := r.Run(); err != nil {
		log.Fatalf("failed to run server: %v", err)
	}
}

func startServer(r *gin.Engine) {
	cfg := util.InitConfig(".")
	conn, err := sql.Open("mysql", cfg.DBSource)
	if err != nil {
		log.Fatal("无法连接到数据库", err)
	}

	if err = conn.Ping(); err != nil {
		log.Fatal("无法 ping 通数据库", err)
	}

	store := db.NewStore(conn)
	srv := service.NewDailyDataService(store)
	ctrl := controller.NewDailyDataController(srv)

	router.Setup(r, ctrl)
}
