package auth_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mock_controller "github.com/raozhaizhu/go-estate/internal/controller/auth/mock"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	authSvc "github.com/raozhaizhu/go-estate/internal/service/auth"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	ctxKey "github.com/raozhaizhu/go-estate/pkg/ctx_key"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Refresh
 * =====================================================================================
 */

func TestRefresh(t *testing.T) {
	// 准备数据
	refreshUrl := delivery.CurrAPI + "/auth/refresh"
	username := "correct"
	deviceID, userAgent, _, clientIPWithPort := testUtil.DeviceID, testUtil.UserAgent, testUtil.ClientIp, testUtil.ClientIpWithPort
	refreshStr, wrongStr := testUtil.RefreshStr, "wrong"
	accessStr := testUtil.AccessStr
	correctHeaderCookie := map[string]any{
		"X-Device-ID": deviceID,
		"User-Agent":  userAgent,
		"myCookie": http.Cookie{
			Name:  ctxKey.CtxKeyRefreshToken,
			Value: refreshStr,
			Path:  "/",
		},
	}
	correctHeaderWrongCookie := map[string]any{
		"X-Device-ID": deviceID,
		"User-Agent":  userAgent,
		"myCookie": http.Cookie{
			Name:  ctxKey.CtxKeyRefreshToken,
			Value: wrongStr,
			Path:  "/",
		},
	}
	correctHeaderNoCookie := map[string]any{
		"X-Device-ID": deviceID,
		"User-Agent":  userAgent,
	}
	correctDto := &authSvc.DTO{
		AccessToken:          accessStr,
		AccessTokenExpiredAt: time.Now().Add(time.Hour),
		UserInfo: authSvc.UserInfo{
			Username: username,
			Role:     userDomain.RoleUser,
		},
	}
	// 默认执行逻辑
	defaultAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any) {
		req, err := http.NewRequest(http.MethodPost, reqUrl, nil)
		require.NoError(t, err)

		req.RemoteAddr = clientIPWithPort // 设置 IP
		for k, v := range customData {    // 设置 UserAgent 和 DeviceID
			if strVal, ok := v.(string); ok {
				req.Header.Set(k, strVal)
			}
			if _cookie, ok := v.(http.Cookie); ok {
				req.AddCookie(&_cookie)
			}
		}

		//  执行请求
		router.ServeHTTP(writer, req)
	}
	successCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string) {
		var results response.Result[*authSvc.DTO]
		// 反序列化结果
		err := json.Unmarshal(writer.Body.Bytes(), &results)
		require.NoError(t, err)
		assert.Equal(t, expectedHTTPCode, writer.Code)
		assert.Equal(t, expectedBizCode, results.Code)
		assert.Equal(t, username, results.Data.UserInfo.Username)
		assert.Equal(t, results.Msg, expectedMsg)
	}
	failCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string) {
		var results response.Result[*authSvc.DTO]
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

	// 成功桩函数
	stubIncrIPCntSuccess := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().IncrIPCnt(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(int64(1), nil).Times(1)
	}

	testCases := []testCase{
		{
			name:       "携带正确 cookie, 登录成功",
			reqUrl:     refreshUrl,
			customData: correctHeaderCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Refresh(gomock.Any(), refreshStr).
					Return(correctDto, nil).Times(1)
			},
			action:           defaultAction,
			checkResponse:    successCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
		},
		{
			name:       "未携带 cookie, 登录失败",
			reqUrl:     refreshUrl,
			customData: correctHeaderNoCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Refresh(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrCookieNoRefreshToken.Code,
			expectedMsg:      "cookie 内没有 freshToken",
		},
		{
			name:       "携带错误 cookie, 登录失败",
			reqUrl:     refreshUrl,
			customData: correctHeaderWrongCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Refresh(gomock.Any(), wrongStr).
					Return(nil, appError.ErrInvalidToken).Times(1)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrInvalidToken.Code,
			expectedMsg:      "令牌不可用",
		},
		{
			name:       "携带正确 cookie, svc 抛出底层错误, ctrl 兜底处理且不暴露内部信息",
			reqUrl:     refreshUrl,
			customData: correctHeaderCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Refresh(gomock.Any(), refreshStr).
					Return(nil, fmt.Errorf("从数据库获取 Session 失败: %w", sql.ErrNoRows)).
					Times(1)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 500,
			expectedBizCode:  appError.CodeServerErr,
			expectedMsg:      "服务器开小差了",
		},
	}

	runTC(t, testCases)
}
