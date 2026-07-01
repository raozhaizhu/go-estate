package db_test

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/util"
)

var testStore db.Store
var config util.Config

func TestMain(m *testing.M) {
	config = util.InitConfig("../../..")
	testStore = db.InitStore(config.DBSource)

	os.Exit(m.Run())
}
