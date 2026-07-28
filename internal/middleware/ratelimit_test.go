package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter(t *testing.T) {
	testUrl := "/test"
	clientIP, clientIPWithPort := testUtil.ClientIp, testUtil.ClientIpWithPort
	limit, duration := 5, time.Minute
	onceCnt, exceededCnt := int64(1), int64(6)

	// 默认执行逻辑
	defaultAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any, tokenMakerMock *mock_token.MockMaker, ctx *gin.Context, cacheMock *mock_db.MockCache) {
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
		require.NoError(t, err)
		// 执行中间件
		rateLimitMiddleware := middleware.RateLimiter(cacheMock, limit, duration)
		router.Use(rateLimitMiddleware)

		router.GET(reqUrl, func(c *gin.Context) {
			ctx.Keys = c.Keys
			response.Success(c, "success")
		})

		// 设置 IP
		req.RemoteAddr = clientIPWithPort

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

	testCases := []testCase{
		{
			name:   "访问次数未超过",
			reqUrl: testUrl,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				cacheMock.EXPECT().IncrIPCnt(gomock.Any(), clientIP, duration).
					Return(onceCnt, nil).Times(1)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
		},
		{
			name:   "访问次数超过限定次数",
			reqUrl: testUrl,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				cacheMock.EXPECT().IncrIPCnt(gomock.Any(), clientIP, duration).
					Return(exceededCnt, nil).Times(1)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrTooManyRequests.Code,
			expectedMsg:      "请求过多,之后再试",
		},
	}

	runTC(t, testCases)
}
