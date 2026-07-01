package auth_test

import (
	"os"
	"testing"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

var (
	testConfig     util.Config
	testStore      db.Store
	testTokenMaker token.Maker
)

func TestMain(m *testing.M) {
	testConfig = util.InitConfig("../../..")
	testStore = db.InitStore(testConfig.DBSource)
	var err error
	testTokenMaker, err = token.NewJwtMaker(testConfig.TokenSymmetricKey)
	if err != nil {
		panic("初始化令牌铸造器失败: " + err.Error())
	}

	os.Exit(m.Run())
}
