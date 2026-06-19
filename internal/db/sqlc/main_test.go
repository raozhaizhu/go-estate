package db

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/raozhaizhu/go-estate/internal/util"
)

var testStore Store

func TestMain(m *testing.M) {
	cfg := util.InitConfig("../../..")
	testStore = InitStore(cfg.DBSource)

	os.Exit(m.Run())
}
