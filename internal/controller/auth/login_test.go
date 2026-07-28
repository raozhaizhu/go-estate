package auth_test

import (
	"bytes"
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
	"github.com/raozhaizhu/go-estate/internal/service/auth"
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
 * 🏁 Login
 * =====================================================================================
 */

func TestLogin(t *testing.T) {
	// 准备数据
	loginUrl := delivery.CurrAPI + "/auth/login"
	username, password := "correct", "12345678"
	shortPassword, longPassword := "1", "12345678912345678"
	deviceID, userAgent, clientIP, clientIPWithPort := testUtil.DeviceID, testUtil.UserAgent, testUtil.ClientIp, testUtil.ClientIpWithPort
	correctBody := gin.H{
		"username": username,
		"password": password,
	}
	emptyBody := gin.H{}
	brokenBody := `{"username":"test","password":`
	wrongBody := gin.H{
		"usename": username,
		"pasword": password,
	}
	shortBody := gin.H{
		"username": username,
		"password": shortPassword,
	}
	longBody := gin.H{
		"username": username,
		"password": longPassword,
	}
	correctInput := auth.LoginInput{
		Username:  username,
		Password:  password,
		DeviceID:  deviceID,
		UserAgent: userAgent,
		ClientIp:  clientIP,
	}
	refreshStr := testUtil.RefreshStr
	accessStr := testUtil.AccessStr
	correctHeader := map[string]any{
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
		var req *http.Request
		var err error

		if body != nil { // 带 body
			jsonData, err := json.Marshal(body)
			require.NoError(t, err)
			req, err = http.NewRequest(http.MethodPost, reqUrl, bytes.NewBuffer(jsonData))
			req.Header.Set("Content-type", "application/json")
		} else { //不带 body
			req, err = http.NewRequest(http.MethodPost, reqUrl, nil)
		}
		require.NoError(t, err)

		req.RemoteAddr = clientIPWithPort // 设置 IP
		for k, v := range customData {    // 设置 UserAgent 和 DeviceID
			if strVal, ok := v.(string); ok {
				req.Header.Set(k, strVal)
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
		// 校验 cookie 是否设置成功
		cookies := writer.Header().Values("Set-Cookie")
		require.NotEmpty(t, cookies)
		cookiesFound := false
		for _, cookie := range cookies {
			if strings.Contains(cookie, string(ctxKey.CtxKeyRefreshToken)+"="+refreshStr) {
				cookiesFound = true
				assert.Contains(t, cookie, "HttpOnly")
				break
			}
		}
		assert.True(t, cookiesFound, "Cookie里没有找到对应的 refreshToken")
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
			name:       "账号密码正确, 登录成功",
			reqUrl:     loginUrl,
			body:       correctBody,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Login(gomock.Any(), correctInput).
					Return(correctDto, refreshStr, nil).Times(1)
			},
			action:           defaultAction,
			checkResponse:    successCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
		},
		{
			name:       "Body为空, 登录失败",
			reqUrl:     loginUrl,
			body:       emptyBody,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Login(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Username为必填字段, Password为必填字段",
		},
		{
			name:       "Body破损, 登录失败",
			reqUrl:     loginUrl,
			body:       brokenBody,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Login(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "参数格式或类型错误",
		},
		{
			name:       "Body格式错误, 登录失败",
			reqUrl:     loginUrl,
			body:       wrongBody,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Login(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Username为必填字段, Password为必填字段",
		},
		{
			name:       "密码格式错误(过短), 登录失败",
			reqUrl:     loginUrl,
			body:       shortBody,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Login(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Password长度必须至少为8个字符",
		},
		{
			name:       "密码格式错误(过长), 登录失败",
			reqUrl:     loginUrl,
			body:       longBody,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Login(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Password长度不能超过16个字符",
		},
		{
			name:       "账号密码错误, 登录失败",
			reqUrl:     loginUrl,
			body:       correctBody,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Login(gomock.Any(), correctInput).
					Return(nil, "", appError.ErrWrongUsernamePassword).Times(1)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeWrongUsernamePassword,
			expectedMsg:      "账户名或密码错误",
		},
		{
			name:       "账号密码正确, svc 抛出底层错误, ctrl 兜底处理且不暴露内部信息",
			reqUrl:     loginUrl,
			body:       correctBody,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubIncrIPCntSuccess(cacheMock)
				svcMock.EXPECT().Login(gomock.Any(), correctInput).
					Return(nil, "", fmt.Errorf("从数据库获取 Session 失败: %w", sql.ErrNoRows)).Times(1)
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
