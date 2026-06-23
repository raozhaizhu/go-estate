package dailyData

import (
	"os"
	"testing"
	"time"

	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/pkg/token"
	"github.com/raozhaizhu/go-estate/pkg/validator"
)

var userPayload, err1 = token.NewPayload("Bob", role.RoleUser, time.Hour, token.TokenTypeAccessToken)
var vipPayload, err2 = token.NewPayload("Bob", role.RoleVip, time.Hour, token.TokenTypeAccessToken)
var adminPayload, err3 = token.NewPayload("Bob", role.RoleAdmin, time.Hour, token.TokenTypeAccessToken)

func TestMain(m *testing.M) {
	validator.InitTrans()

	os.Exit(m.Run())

}
