// go:build integration

package auth_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/raozhaizhu/go-estate/internal/delivery"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthHappyPathFlow 仅测试 Auth 模块的正确路径
func TestAuthHappyPathFlow(t *testing.T) {
	// 启动服务器, 配置 cookie
	testServer, deps := testUtil.SetupIntegrationTest(t)
	defer testServer.Close()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := testServer.Client()
	client.Jar = jar

	// 准备数据
	user, password := testUtil.CreateRandomUser(t, deps.Store, userDomain.RoleUser, "")
	username := user.Username

	loginBody := map[string]string{
		"username": username,
		"password": password,
	}
	deviceID, userAgent := testUtil.DeviceID, testUtil.UserAgent
	// Url
	loginUrl := testServer.URL + delivery.AuthAPI + "/login"
	refreshUrl := testServer.URL + delivery.AuthAPI + "/refresh"
	logoutUrl := testServer.URL + delivery.AuthAPI + "/logout"

	// 预备数据
	var userToken string

	// 默认行为
	defaultAction := func(t *testing.T, reqUrl, method, accessToken string, body map[string]string, client *http.Client) *http.Response {
		// 设置 Body(如有)
		var reqBody io.Reader
		if body != nil {
			jsonBytes, err := json.Marshal(body)
			require.NoError(t, err)
			reqBody = bytes.NewBuffer(jsonBytes)
		}
		// 初始化请求
		req, err := http.NewRequest(method, reqUrl, reqBody)
		require.NoError(t, err)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		// 添加头信息
		req.Header.Set("X-Device-Id", deviceID)
		req.Header.Set("User-Agent", userAgent)
		if accessToken != "" {
			req.Header.Set("Authorization", "Bearer "+accessToken)
		}
		// 执行请求
		resp, err := client.Do(req)
		require.NoError(t, err)

		return resp
	}

	getUserToken := func() string {
		return userToken
	}

	loginCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*auth.DTO]
		err = json.Unmarshal(respBytes, &results)
		assert.NoError(t, err)
		// 校验业务代码
		assert.Equal(t, results.Code, expectedBizCode)
		// 校验用户信息一致
		assert.Equal(t, username, results.Data.UserInfo.Username)

		// 设置 accessToken
		userToken = results.Data.AccessToken
	}
	refreshCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*auth.DTO]
		err = json.Unmarshal(respBytes, &results)
		assert.NoError(t, err)
		// 校验业务代码
		assert.Equal(t, results.Code, expectedBizCode)
		// 校验用户信息一致
		assert.Equal(t, username, results.Data.UserInfo.Username)

		// 设置 accessToken
		userToken = results.Data.AccessToken
	}
	logoutCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[any]
		err = json.Unmarshal(respBytes, &results)
		assert.NoError(t, err)
		// 校验业务代码
		assert.Equal(t, results.Code, expectedBizCode)
	}

	testCases := []testUtil.IntgTestCase{
		{
			Name:             "登录账号",
			ReqUrl:           loginUrl,
			Method:           "POST",
			Body:             loginBody,
			Action:           defaultAction,
			CheckResponse:    loginCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "重刷令牌",
			ReqUrl:           refreshUrl,
			Method:           "POST",
			Action:           defaultAction,
			CheckResponse:    refreshCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "退出登录",
			ReqUrl:           logoutUrl,
			Method:           "POST",
			GetToken:         getUserToken,
			Action:           defaultAction,
			CheckResponse:    logoutCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
	}

	testUtil.RunIntgTC(t, testCases, client)
}
