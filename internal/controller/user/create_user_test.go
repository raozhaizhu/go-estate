package user_test

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

// TestCreateUser 测试创建新用户
// 不带请求头可创建 User, 带 Admin 请求头可以创建 Vip
func TestCreateUser(t *testing.T) {
	// 准备数据
	createUserUrl := delivery.CurrAPI + "/user"
	createVipUrl := delivery.CurrAPI + "/user/vip"
	username, password, email := util.RandomUsername(), util.RandomPassword(), util.RandomEmail()
	shortUsername, shortPassword, longUsername, longPassword := util.RandomString(1), util.RandomString(1), util.RandomString(33), util.RandomString(17)
	deviceID, userAgent, _, clientIPWithPort := testUtil.DeviceID, testUtil.UserAgent, testUtil.ClientIp, testUtil.ClientIpWithPort
	correctBody := gin.H{
		"username": username,
		"password": password,
		"email":    email,
	}
	_, authorization := testUtil.AccessStr, testUtil.AuthorizationAccessToken
	emptyBody := gin.H{}
	brokenBody := `{"username":"test","password":`
	wrongBody := gin.H{
		"usename": username,
		"pasword": password,
		"emai":    email,
	}
	shortBody := gin.H{
		"username": shortUsername,
		"password": shortPassword,
		"email":    email,
	}
	longBody := gin.H{
		"username": longUsername,
		"password": longPassword,
		"email":    email,
	}
	badEmailBody := gin.H{
		"username": username,
		"password": password,
		"email":    "123.com",
	}
	correctInput := user.CreateUserInput{
		Username: username,
		Password: password,
		Email:    email,
	}
	correctUserDto := &user.DTO{
		Username: username,
		Email:    email,
		Role:     userDomain.RoleUser,
	}
	correctVipDto := &user.DTO{
		Username: username,
		Email:    email,
		Role:     userDomain.RoleVip,
	}
	correctHeaderCookie := map[string]any{
		"X-Device-ID":   deviceID,
		"User-Agent":    userAgent,
		"Authorization": authorization,
	}
	correctHeaderNoCookie := map[string]any{
		"X-Device-ID": deviceID,
		"User-Agent":  userAgent,
	}
	accessAdminPayload := &token.Payload{
		Username:  username,
		Role:      userDomain.RoleAdmin,
		TokenType: token.TokenTypeAccessToken,
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
				return accessAdminPayload, nil
			}).Times(1)
	}

	testCases := []testCase{
		{
			name:       "无 Cookie 直接创建 User",
			reqUrl:     createUserUrl,
			body:       correctBody,
			customData: correctHeaderNoCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				svcMock.EXPECT().CreateUser(gomock.Any(), correctInput, userDomain.RoleUser).
					Return(correctUserDto, nil).Times(1)
			},
			action:           defaultAction,
			checkResponse:    successCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
		},
		{
			name:       "带 Admin 请求头创建 Vip",
			reqUrl:     createVipUrl,
			body:       correctBody,
			customData: correctHeaderCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				stubVerifyTokenSuccess(tokenMakerMock)
				svcMock.EXPECT().CreateUser(gomock.Any(), correctInput, userDomain.RoleVip).
					Return(correctVipDto, nil).Times(1)
			},
			action:           defaultAction,
			checkResponse:    successCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
		},
		{
			name:       "参数不存在, 创建用户失败",
			reqUrl:     createUserUrl,
			body:       emptyBody,
			customData: correctHeaderNoCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				svcMock.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Email为必填字段, Password为必填字段, Username为必填字段",
		},
		{
			name:       "Body破损, 创建用户失败",
			reqUrl:     createUserUrl,
			body:       brokenBody,
			customData: correctHeaderNoCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				svcMock.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "参数格式或类型错误",
		},
		{
			name:       "Body格式错误, 创建用户失败",
			reqUrl:     createUserUrl,
			body:       wrongBody,
			customData: correctHeaderNoCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				svcMock.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Email为必填字段, Password为必填字段, Username为必填字段",
		},
		{
			name:       "密码和用户名过短, 创建用户失败",
			reqUrl:     createUserUrl,
			body:       shortBody,
			customData: correctHeaderNoCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				svcMock.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Password长度必须至少为8个字符, Username长度必须至少为3个字符",
		},
		{
			name:       "密码和用户名过长, 创建用户失败",
			reqUrl:     createUserUrl,
			body:       longBody,
			customData: correctHeaderNoCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				svcMock.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Password长度不能超过16个字符, Username长度不能超过32个字符",
		},
		{
			name:       "Email 格式错误, 创建用户失败",
			reqUrl:     createUserUrl,
			body:       badEmailBody,
			customData: correctHeaderNoCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				svcMock.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Email必须是一个有效的邮箱",
		},
		{
			name:       "参数正确, svc 抛出底层错误, ctrl 兜底处理且不暴露内部信息",
			reqUrl:     createUserUrl,
			body:       correctBody,
			customData: correctHeaderNoCookie,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				svcMock.EXPECT().CreateUser(gomock.Any(), correctInput, userDomain.RoleUser).
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
