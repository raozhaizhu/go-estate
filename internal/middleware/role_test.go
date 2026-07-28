package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	ctxKey "github.com/raozhaizhu/go-estate/pkg/ctx_key"
	"github.com/raozhaizhu/go-estate/pkg/token"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/require"
)

func TestRole(t *testing.T) {
	testUrl := "/test"
	accessUserPayload := &token.Payload{
		Role:      userDomain.RoleUser,
		TokenType: token.TokenTypeAccessToken,
	}

	// 默认执行逻辑
	noPayloadAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any, tokenMakerMock *mock_token.MockMaker, ctx *gin.Context, cacheMock *mock_db.MockCache) {
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
		require.NoError(t, err)

		//  挂载 Role 中间件
		roleMiddleware := middleware.RequireRoles([]role.Role{userDomain.RoleVip, userDomain.RoleVip})
		router.Use(roleMiddleware)

		router.GET(reqUrl, func(c *gin.Context) {
			response.Success(c, "success")
		})

		//  执行请求
		router.ServeHTTP(writer, req)

	}
	stoppedUserAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any, tokenMakerMock *mock_token.MockMaker, ctx *gin.Context, cacheMock *mock_db.MockCache) {
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
		require.NoError(t, err)

		// 挂载临时中间件, 往 ctx 注入 payload
		router.Use(func(c *gin.Context) {
			c.Set(ctxKey.CtxKeyPayload, accessUserPayload)
			c.Next()
		})
		//  挂载 Role 中间件
		roleMiddleware := middleware.RequireRoles(userDomain.RoleAtLeastVip)
		router.Use(roleMiddleware)

		router.GET(reqUrl, func(c *gin.Context) {
			response.Success(c, "success")
		})

		//  执行请求
		router.ServeHTTP(writer, req)

	}
	passedUserAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any, tokenMakerMock *mock_token.MockMaker, ctx *gin.Context, cacheMock *mock_db.MockCache) {
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
		require.NoError(t, err)

		// 挂载临时中间件, 往 ctx 注入 payload
		router.Use(func(c *gin.Context) {
			c.Set(ctxKey.CtxKeyPayload, accessUserPayload)
			c.Next()
		})
		//  挂载 Role 中间件
		roleMiddleware := middleware.RequireRoles(userDomain.RoleAtLeastUser)
		router.Use(roleMiddleware)

		router.GET(reqUrl, func(c *gin.Context) {
			response.Success(c, "success")
		})

		//  执行请求
		router.ServeHTTP(writer, req)

	}
	failCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string, ctx *gin.Context) {
		var results response.Result[interface{}]
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
	successCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string, ctx *gin.Context) {
		var results response.Result[string]
		// 反序列化结果
		err := json.Unmarshal(writer.Body.Bytes(), &results)
		require.NoError(t, err)
		require.Equal(t, 200, writer.Code)
		require.Equal(t, expectedBizCode, results.Code)
		require.Equal(t, expectedMsg, results.Msg)
	}

	testCases := []testCase{
		{
			name:   "无法从上下文获取 payload, 被拒绝",
			reqUrl: testUrl,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           noPayloadAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrAuthRequired.Code,
			expectedMsg:      appError.ErrAuthRequired.Msg,
		},
		{
			name:   "User 访问 VIP/ADMIN 用户组, 被拒绝",
			reqUrl: testUrl,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           stoppedUserAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrAuthPermissionDenied.Code,
			expectedMsg:      appError.ErrAuthPermissionDenied.Msg,
		},
		{
			name:   "User 访问 USER/VIP/ADMIN 用户组, 成功",
			reqUrl: testUrl,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           passedUserAction,
			checkResponse:    successCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
		},
	}

	runTC(t, testCases)
}
