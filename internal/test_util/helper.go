package testUtil

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/server"
	"github.com/raozhaizhu/go-estate/internal/util"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

func CreateRandomUser(t *testing.T, testStore db.Store, role userDomain.Role, avatarKey string) (db.User, string) {
	// 初始化用户信息
	username := util.RandomUsername()
	password := util.RandomPassword()

	return CreateSpecificUser(t, username, password, testStore, role, avatarKey)
}

func CreateSpecificUser(t *testing.T, username, password string, testStore db.Store, role userDomain.Role, avatarKey string) (db.User, string) {
	// 构造参数
	params := PrepareCreateUserParams(t, username, password, role, avatarKey)

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

	return user, password
}

func PrepareCreateUserParams(t *testing.T, username, password string, role userDomain.Role, avatarKey string) db.CreateUserParams {
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	email := util.DeriveEmail(username)
	if avatarKey == "" {
		avatarKey = objectStore.DefaultAvatarKey
	}

	params := db.CreateUserParams{
		Username:       username,
		HashedPassword: hashedPassword,
		Email:          email,
		Role:           int16(role),
		AvatarKey:      avatarKey,
	}

	return params
}

// SetupIntegrationTest 准备所有真实依赖，并启动一个 Gin 测试服务器
func SetupIntegrationTest(t *testing.T) (*httptest.Server, *app.Deps) {
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

	return testServer, deps
}

func RunIntgTC(t *testing.T, testCases []IntgTestCase, client *http.Client) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var resp *http.Response
			var currToken string

			// 获取 token
			if tc.GetToken != nil {
				currToken = tc.GetToken()
			}

			// 执行行动
			resp = tc.Action(t, tc.ReqUrl, tc.Method, currToken, tc.Body, client)

			// 关闭 Body 防止泄露
			defer resp.Body.Close()
			// 断言 HTTP 状态码
			require.Equal(t, resp.StatusCode, tc.ExpectedHTTPCode)
			// 读取响应体
			respBytes, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			// 校验结果
			tc.CheckResponse(t, respBytes, tc.ExpectedBizCode, tc.ExpectedMsg)
		})

	}
}

/** ====================================================================================
 * 🏁 Types
 * =====================================================================================
 */

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
	// GetToken 用于获取令牌信息
	GetToken func() string
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
