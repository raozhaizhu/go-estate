package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/raozhaizhu/go-estate/util"
)

var testStore Store

func TestMain(m *testing.M) {

	conn, err := sql.Open("mysql", util.DB_URL)
	if err != nil {
		log.Fatal("无法连接到数据库", err)
	}

	if err = conn.Ping(); err != nil {
		log.Fatal("无法 ping 通数据库", err)
	}

	testStore = NewStore(conn)

	os.Exit(m.Run())
}
