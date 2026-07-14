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
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuth(t *testing.T) {
	accessStr, authorization := testUtil.AccessStr, testUtil.AuthorizationAccessToken
	testUrl := "/test"
	correctHeader := map[string]any{"Authorization": authorization}
	emptyHeader := map[string]any{}
	noBearerHeader := map[string]any{"Authorization": "hello 1"}
	onlyBearerHeader := map[string]any{"Authorization": "bearer"}
	accessPayload := &token.Payload{
		Role:      userDomain.RoleUser,
		TokenType: token.TokenTypeAccessToken,
	}
	// 默认执行逻辑
	defaultAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any, tokenMakerMock *mock_token.MockMaker, ctx *gin.Context) {
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
		require.NoError(t, err)

		// 执行中间件
		authMiddleware := middleware.RequireAuth(tokenMakerMock)
		router.Use(authMiddleware)

		router.GET(reqUrl, func(c *gin.Context) {
			ctx.Keys = c.Keys
			response.Success(c, "success")
		})

		for k, v := range customData { // 设置 UserAgent 和 DeviceID
			if strVal, ok := v.(string); ok {
				req.Header.Set(k, strVal)
			}
		}

		//  执行请求
		router.ServeHTTP(writer, req)

	}
	failCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string, ctx *gin.Context) {
		var results response.Result[interface{}]
		// 反序列化结果
		err := json.Unmarshal(writer.Body.Bytes(), &results)
		require.NoError(t, err)
		assert.Equal(t, expectedHTTPCode, writer.Code)
		assert.Equal(t, expectedBizCode, results.Code)
		// 校验 Msg 是否一致
		expSlice := strings.Split(expectedMsg, ", ")
		actSlice := strings.Split(results.Msg, ", ")
		sort.Strings(expSlice)
		sort.Strings(actSlice)
		assert.Equal(t, expSlice, actSlice)
	}
	successCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string, ctx *gin.Context) {
		var results response.Result[string]
		// 反序列化结果
		err := json.Unmarshal(writer.Body.Bytes(), &results)
		require.NoError(t, err)
		assert.Equal(t, 200, writer.Code)
		assert.Equal(t, expectedBizCode, results.Code)
		assert.Equal(t, expectedMsg, results.Msg)
		// 校验上下文
		payload, ok := ctx.Get(token.PayloadKey)
		assert.True(t, ok)
		assert.Equal(t, accessPayload, payload)
	}
	// 成功桩函数
	stubVerifyTokenSuccess := func(tokenMakerMock *mock_token.MockMaker) {
		tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).
			DoAndReturn(func(tokenStr string, tokenType token.TokenType) (*token.Payload, error) {
				require.Equal(t, tokenType, token.TokenType(token.TokenTypeAccessToken))
				return accessPayload, nil
			}).Times(1)
	}

	testCases := []testCase{
		{
			name:       "认证头不存在 ",
			reqUrl:     testUrl,
			customData: emptyHeader,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker) {
				tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrAuthNoHeader.Code,
			expectedMsg:      "没有认证头",
		},
		{
			name:       "认证头格式错误: 不带 bearer",
			reqUrl:     testUrl,
			customData: noBearerHeader,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker) {
				tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrAuthBadHeader.Code,
			expectedMsg:      "认证头格式错误",
		},
		{
			name:       "认证头格式错误: 只有 bearer",
			reqUrl:     testUrl,
			customData: onlyBearerHeader,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker) {
				tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrAuthBadHeader.Code,
			expectedMsg:      "认证头格式错误",
		},
		{
			name:       "认证头格式正确, 验证失败返回内部错误, 中间件正确处理不暴露内部信息",
			reqUrl:     testUrl,
			customData: correctHeader,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker) {
				tokenMakerMock.EXPECT().VerifyToken(accessStr, token.TokenType(token.TokenTypeAccessToken)).
					Return(nil, appError.ErrServerErr).Times(1)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 500,
			expectedBizCode:  appError.ErrServerErr.Code,
			expectedMsg:      "服务器开小差了",
		},
		{
			name:       "认证头格式正确, 验证成功",
			reqUrl:     testUrl,
			customData: correctHeader,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker) {
				stubVerifyTokenSuccess(tokenMakerMock)
			},
			action:           defaultAction,
			checkResponse:    successCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
		},
	}

	runTC(t, testCases)
}
