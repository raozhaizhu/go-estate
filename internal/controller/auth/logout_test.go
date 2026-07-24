package auth_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
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
	"github.com/raozhaizhu/go-estate/pkg/token"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Refresh
 * =====================================================================================
 */

func TestLogout(t *testing.T) {
	// 准备数据
	logoutUrl := delivery.CurrAPI + "/auth/logout"
	username := "correct"
	deviceID, userAgent, _, clientIPWithPort := testUtil.DeviceID, testUtil.UserAgent, testUtil.ClientIp, testUtil.ClientIpWithPort
	_, authorization := testUtil.AccessStr, testUtil.AuthorizationAccessToken
	correctHeader := map[string]any{
		"X-Device-ID":   deviceID,
		"User-Agent":    userAgent,
		"Authorization": authorization,
	}

	correctInput := auth.LogoutInput{
		Username: username,
		DeviceID: deviceID,
	}
	accessPayload := &token.Payload{
		Username:  username,
		Role:      userDomain.RoleUser,
		TokenType: token.TokenTypeAccessToken,
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
		}

		//  执行请求
		router.ServeHTTP(writer, req)
	}
	successCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string) {
		var results response.Result[*authSvc.DTO]
		// 反序列化结果
		err := json.Unmarshal(writer.Body.Bytes(), &results)
		require.NoError(t, err)
		assert.Equal(t, 200, writer.Code)
		assert.Equal(t, expectedBizCode, results.Code)
		assert.Equal(t, expectedMsg, results.Msg)
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
	stubVerifyTokenSuccess := func(tokenMakerMock *mock_token.MockMaker) {
		tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).
			DoAndReturn(func(tokenStr string, tokenType token.TokenType) (*token.Payload, error) {
				require.Equal(t, tokenType, token.TokenType(token.TokenTypeAccessToken))
				return accessPayload, nil
			}).Times(1)
	}

	testCases := []testCase{
		{
			name:       "携带正确认证信息, 登出成功",
			reqUrl:     logoutUrl,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubVerifyTokenSuccess(tokenMakerMock)
				svcMock.EXPECT().Logout(gomock.Any(), correctInput).
					Return(nil).Times(1)
			},
			action:           defaultAction,
			checkResponse:    successCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
		},
		{
			name:       "获取 payload 成功, svc抛出底层错误, ctrl 兜底处理且不暴露内部信息",
			reqUrl:     logoutUrl,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache) {
				stubVerifyTokenSuccess(tokenMakerMock)
				dbErr := &mysql.MySQLError{Number: 1205, Message: "Lock wait timeout exceeded"}
				svcMock.EXPECT().Logout(gomock.Any(), correctInput).
					Return(appError.ErrServerErr.WithErr(fmt.Errorf("获取用户活跃 Sessions 失败: %w", dbErr))).Times(1)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 500,
			expectedBizCode:  appError.ErrServerErr.Code,
			expectedMsg:      "服务器开小差了",
		},
	}

	runTC(t, testCases)
}
