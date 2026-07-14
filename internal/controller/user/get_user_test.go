package user_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mock_controller "github.com/raozhaizhu/go-estate/internal/controller/user/mock"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetUser 测试获取用户
func TestGetUser(t *testing.T) {
	// 准备数据
	basicUrl := delivery.CurrAPI + "/user"
	username, _, email := util.RandomUsername(), util.RandomPassword(), util.RandomEmail()
	deviceID, userAgent, _, clientIPWithPort := testUtil.DeviceID, testUtil.UserAgent, testUtil.ClientIp, testUtil.ClientIpWithPort
	correctRequest := basicUrl + "/" + username
	emptyRequest := basicUrl + "/" + " "
	_, authorization := testUtil.AccessStr, testUtil.AuthorizationAccessToken
	correctInput := user.GetUserInput{
		Username: username,
	}
	correctUserDto := &user.DTO{
		Username: username,
		Email:    email,
		Role:     userDomain.RoleUser,
	}
	accessUserPayload := &token.Payload{
		Username:  username,
		Role:      userDomain.RoleUser,
		TokenType: token.TokenTypeAccessToken,
	}
	correctHeaderCookie := map[string]any{
		"X-Device-ID":   deviceID,
		"User-Agent":    userAgent,
		"Authorization": authorization,
	}

	// 默认执行逻辑
	defaultAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any) {
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
		assert.NoError(t, err)

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
		var results response.Result[*user.DTO]
		// 反序列化结果
		err := json.Unmarshal(writer.Body.Bytes(), &results)
		assert.NoError(t, err)
		assert.Equal(t, 200, writer.Code)
		assert.Equal(t, expectedBizCode, results.Code)
		assert.Equal(t, username, results.Data.Username)
		assert.Equal(t, results.Msg, expectedMsg)
	}
	failCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string) {
		var results response.Result[*user.DTO]
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
				return accessUserPayload, nil
			}).Times(1)
	}

	testCases := []testCase{
		{
			name:       "带 User 请求头获取自身信息成功",
			reqUrl:     correctRequest,
			customData: correctHeaderCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				stubVerifyTokenSuccess(tokenMakerMock)
				svcMock.EXPECT().GetUser(gomock.Any(), correctInput).
					Return(correctUserDto, nil).Times(1)
			},
			action:           defaultAction,
			checkResponse:    successCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
		},
		{
			name:       "参数不存在",
			reqUrl:     emptyRequest,
			customData: correctHeaderCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				stubVerifyTokenSuccess(tokenMakerMock)
				svcMock.EXPECT().GetUser(gomock.Any(), correctInput).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Username长度必须至少为3个字符",
		},
		{
			name:       "参数正确, svc 抛出底层错误, ctrl 兜底处理且不暴露内部信息",
			reqUrl:     correctRequest,
			customData: correctHeaderCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				stubVerifyTokenSuccess(tokenMakerMock)
				svcMock.EXPECT().GetUser(gomock.Any(), correctInput).
					Return(nil, fmt.Errorf("从数据库获取 Session 失败: %w", sql.ErrNoRows)).Times(1)
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
