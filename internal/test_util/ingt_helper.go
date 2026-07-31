package testUtil

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raozhaizhu/go-estate/internal/domain/app"
	"github.com/raozhaizhu/go-estate/internal/server"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 TestCase
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
 * 🏁 Helper
 * =====================================================================================
 */

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
