package testUtil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/server"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

func CreateRandomUser(t *testing.T, testStore db.Store) db.User {
	// 初始化用户信息
	username := util.RandomUsername()
	password := util.RandomPassword()

	return CreateSpecificUser(t, username, password, testStore)
}

func CreateSpecificUser(t *testing.T, username, password string, testStore db.Store) db.User {
	// 构造参数
	params := PrepareCreateUserParams(t, username, password)

	// 指定期望时间
	expectedPwdChangedAt := time.Date(1970, 1, 1, 0, 0, 1, 0, time.UTC)
	expectedCreatedAt := time.Now()

	// 创建用户
	result, err := testStore.CreateUser(context.Background(), params)
	require.NoError(t, err)
	rows, err := result.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), rows)

	// 查询并比较用户
	user, err := testStore.GetUser(context.Background(), username)
	require.NoError(t, err)
	require.Equal(t, username, user.Username)
	require.Equal(t, params.HashedPassword, user.HashedPassword)
	require.Equal(t, params.Email, user.Email)
	require.Equal(t, params.Role, user.Role)
	require.Equal(t, expectedPwdChangedAt, user.PasswordChangedAt)
	require.WithinDuration(t, expectedCreatedAt, user.CreatedAt, time.Second)
	require.NotEmpty(t, user.ID)

	return user
}

func PrepareCreateUserParams(t *testing.T, username, password string) db.CreateUserParams {
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	email := util.DeriveEmail(username)
	_role := int16(role.RoleUser)

	params := db.CreateUserParams{
		Username:       username,
		HashedPassword: hashedPassword,
		Email:          email,
		Role:           _role,
	}

	return params
}

// SetupIntegrationTest 准备所有真实依赖，并启动一个 Gin 测试服务器
func SetupIntegrationTest(t *testing.T) *httptest.Server {
	deps := app.PrepareTestDeps()
	// 清理数据库,缓存,面向对象存储
	deps.Store.CleanTestStore(t)
	deps.Cache.CleanTestCache(t)
	deps.ObjectStore.CleanTestObjectStore(t)

	server, err := server.NewServer(deps)
	require.NoError(t, err)
	testServer := httptest.NewServer(server)

	// 退出时释放资源,关闭异步任务处理器,数据库,缓存,
	t.Cleanup(func() {
		testServer.Close()
		deps.TaskProcessor.Stop()
		deps.Store.Close()
		deps.Cache.Close()
	})

	return testServer
}

// IntgTestCase 集成测试用例
type IntgTestCase struct {
	// Name 测试用例名称
	Name string
	// ReqUrl HTTP 请求路径
	ReqUrl string
	// Body HTTP 请求体
	Body map[string]string
	// Method HTTP方法
	Method string
	// AccessToken 令牌信息
	AccessToken string
	// Action 执行服务函数
	Action func(t *testing.T, reqUrl, method, accessToken string, body map[string]string, client *http.Client) *http.Response
	// CheckResponse 校验数据函数
	CheckResponse func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string)
	// ExpectedHTTPCode 期望的 HTTP 代码
	ExpectedHTTPCode int
	// ExpectedBizCode 期望的业务代码
	ExpectedBizCode int
	// ExpectedMsg 期望的业务信息
	ExpectedMsg string
}
