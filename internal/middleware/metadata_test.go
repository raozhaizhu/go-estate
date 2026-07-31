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
	"github.com/raozhaizhu/go-estate/internal/middleware"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/require"
)

func TestMetadata(t *testing.T) {
	testUrl := "/test"
	_, authorization := testUtil.AccessStr, testUtil.AuthorizationAccessToken
	deviceID, userAgent, clientIP, clientIPWithPort := testUtil.DeviceID, testUtil.UserAgent, testUtil.ClientIp, testUtil.ClientIpWithPort
	correctHeader := map[string]any{
		"X-Device-ID":   deviceID,
		"User-Agent":    userAgent,
		"Authorization": authorization,
	}
	noDeviceIDHeader := map[string]any{
		"User-Agent":    userAgent,
		"Authorization": authorization,
	}
	noUserAgentHeader := map[string]any{
		"X-Device-ID":   deviceID,
		"Authorization": authorization,
	}

	// 默认执行逻辑
	defaultAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
		require.NoError(t, err)

		// 设置中间件和答复 api
		middleware := middleware.RequireMetadata()
		router.Use(middleware)
		router.GET(reqUrl, func(c *gin.Context) { // 通过中间件, 默认返回成功
			capturedCtx = c.Request.Context()
			response.Success(c, "success")
		})

		req.RemoteAddr = clientIPWithPort // 设置 IP
		for k, v := range customData {    // 设置 UserAgent 和 DeviceID
			if strVal, ok := v.(string); ok {
				req.Header.Set(k, strVal)
			}
		}

		//  执行请求
		router.ServeHTTP(writer, req)
	}
	failCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string) {
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
	successCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string) {
		var results response.Result[string]
		// 反序列化结果
		err := json.Unmarshal(writer.Body.Bytes(), &results)
		require.NoError(t, err)
		require.Equal(t, 200, writer.Code)
		require.Equal(t, expectedBizCode, results.Code)
		require.Equal(t, expectedMsg, results.Msg)
		// 校验上下文
		meta, err := middleware.GetCtxClientMeta(capturedCtx)
		require.NoError(t, err)
		require.Equal(t, deviceID, meta.DeviceID)
		require.Equal(t, userAgent, meta.UserAgent)
		require.Equal(t, clientIP, meta.ClientIP)
	}

	testCases := []testCase{
		{
			name:       "device_id不存在",
			reqUrl:     testUrl,
			customData: noDeviceIDHeader,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrEmptyDeviceID.Code,
			expectedMsg:      appError.ErrEmptyDeviceID.Msg,
		},
		{
			name:       "user_agent不存在",
			reqUrl:     testUrl,
			customData: noUserAgentHeader,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrEmptyUserAgent.Code,
			expectedMsg:      appError.ErrEmptyUserAgent.Msg,
		},
		{
			name:       "元信息齐全, KV 写入上下文成功",
			reqUrl:     testUrl,
			customData: correctHeader,
			buildStubs: func(tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).Times(0)
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
