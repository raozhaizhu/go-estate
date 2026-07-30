package middleware_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	myWebsocket "github.com/raozhaizhu/go-estate/internal/my_websocket"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	ctxKey "github.com/raozhaizhu/go-estate/pkg/ctx_key"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/require"
)

func TestServeWS(t *testing.T) {
	testUrl := "/ws"
	username := util.RandomUsername()
	rawStr := util.RandomString(16)
	wsKey := base64.StdEncoding.EncodeToString([]byte(rawStr))
	invalidHeader := map[string]any{
		"Connection":            "upgrade",
		"Upgrade":               "websocket",
		"Sec-Websocket-Version": "13",
		"Sec-Websocket-Key":     "wrong",
	}
	correctHeader := map[string]any{
		"Connection":            "upgrade",
		"Upgrade":               "websocket",
		"Sec-Websocket-Version": "13",
		"Sec-Websocket-Key":     wsKey,
	}

	// 默认执行逻辑
	defaultAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
		require.NoError(t, err)

		// 设置中间件和答复 api
		manager := myWebsocket.NewManager()
		middleware := middleware.ServeWS(manager)
		router.Use(middleware)
		router.GET(reqUrl, func(c *gin.Context) { // 通过中间件, 默认返回成功
			capturedCtx = c.Request.Context()
			response.Success(c, "success")
		})

		// 设置 header
		for k, v := range customData {
			if strVal, ok := v.(string); ok {
				req.Header.Set(k, strVal)
			}
		}

		//  执行请求
		router.ServeHTTP(writer, req)
	}
	authAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
		require.NoError(t, err)

		// 设置中间件和答复 api
		router.Use(func(c *gin.Context) { // 临时中间件用于挂载 username
			c.Set(ctxKey.CtxKeyUserName, username) // 设置 username
			c.Next()
		})

		// 挂载 ws 中间件
		manager := myWebsocket.NewManager()
		middleware := middleware.ServeWS(manager)
		router.Use(middleware)
		router.GET(reqUrl, middleware)

		// 设置 header
		for k, v := range customData {
			if strVal, ok := v.(string); ok {
				req.Header.Set(k, strVal)
			}
		}

		//  执行请求
		router.ServeHTTP(writer, req)
	}
	successAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
		manager := myWebsocket.NewManager()

		// 挂载路由和中间件
		router.Use(func(c *gin.Context) {
			c.Set(ctxKey.CtxKeyUserName, username)
			c.Next()
		})
		router.GET(reqUrl, middleware.ServeWS(manager))

		// 启动真实的本地服务
		server := httptest.NewServer(router)
		defer server.Close()

		// 构造 ws 地址(去掉前面的 http 并换成 ws)
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + reqUrl
		// 构造 dialer 发起 tcp 连接
		conn, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)

		// 断言升级成功
		require.NoError(t, err)
		require.Equal(t, http.StatusSwitchingProtocols, resp.StatusCode)
		defer conn.Close()

		// 验证升级成功后，Manager 是否真的注册了 Client (可以考虑给点时间)
		// time.Sleep(10 * time.Millisecond)
		connecting := manager.UserConnecting(username)
		require.True(t, connecting)
	}

	failCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string) {
		var results response.Result[any]
		// 反序列化结果
		err := json.Unmarshal(writer.Body.Bytes(), &results)
		require.NoError(t, err)
		require.Equal(t, expectedHTTPCode, writer.Code)
		require.Equal(t, expectedBizCode, results.Code)
		// 校验 Msg 是否一致
		expSlice := strings.Split(expectedMsg, ", ")
		actSlice := strings.Split(results.Msg, ", ")
		sort.Strings(expSlice)
		sort.Strings(actSlice)
		require.Equal(t, expSlice, actSlice)
	}
	upgradeFailResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string) {
		var results response.Result[any]
		// 反序列化结果
		err := json.Unmarshal(writer.Body.Bytes(), &results)
		require.NoError(t, err)
		require.Equal(t, expectedHTTPCode, writer.Code)
		require.Equal(t, expectedBizCode, results.Code)
		// 校验 Msg 是否一致
		expSlice := strings.Split(expectedMsg, ", ")
		actSlice := strings.Split(results.Msg, ", ")
		sort.Strings(expSlice)
		sort.Strings(actSlice)
		require.Equal(t, expSlice, actSlice)
	}

	testCases := []testCase{
		{
			name:             "账户未登录",
			reqUrl:           testUrl,
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrAuthRequired.Code,
			expectedMsg:      appError.ErrAuthRequired.Msg,
		},
		{
			name:             "账户已登录, 参数错误导致升级协议失败",
			reqUrl:           testUrl,
			action:           authAction,
			customData:       invalidHeader,
			checkResponse:    upgradeFailResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrBadWSUpgradeHeader.Code,
			expectedMsg:      "Websocket 升级参数错误",
		},
		{
			name:       "账户已登录, 升级协议成功",
			reqUrl:     testUrl,
			action:     successAction,
			customData: correctHeader,
		},
	}

	runTC(t, testCases)

}
