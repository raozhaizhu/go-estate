package auth_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"

	"github.com/raozhaizhu/go-estate/internal/delivery"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthHappyPathFlow 仅测试 Auth 模块的正确路径
func TestAuthHappyPathFlow(t *testing.T) {
	// 启动服务器, 配置 cookie
	testServer := testUtil.SetupIntegrationTest(t)
	defer testServer.Close()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := testServer.Client()
	client.Jar = jar

	// 准备数据
	username, password, email := util.RandomUsername(), util.RandomPassword(), util.RandomEmail()
	avatarKey := objectStore.DefaultAvatarKey
	createUserBody := map[string]string{
		"username": username,
		"password": password,
		"email":    email,
	}
	loginBody := map[string]string{
		"username": username,
		"password": password,
	}
	deviceID, userAgent := testUtil.DeviceID, testUtil.UserAgent
	// Url
	createUserUrl := testServer.URL + delivery.UserAPI + "?username=" + username
	loginUrl := testServer.URL + delivery.AuthAPI + "/login"
	refreshUrl := testServer.URL + delivery.AuthAPI + "/refresh"
	logoutUrl := testServer.URL + delivery.AuthAPI + "/logout"

	// 预备数据
	var accessToken string

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
		req, err := http.NewRequest(http.MethodPost, reqUrl, reqBody)
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
	createUserCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*user.DTO]
		err = json.Unmarshal(respBytes, &results)
		assert.NoError(t, err)
		// 校验业务代码
		assert.Equal(t, results.Code, expectedBizCode)
		// 校验用户信息一致
		assert.Equal(t, username, results.Data.Username)
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
		assert.Equal(t, avatarKey, results.Data.UserInfo.AvatarKey)

		// 设置 accessToken
		accessToken = results.Data.AccessToken
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
		accessToken = results.Data.AccessToken
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
			Name:             "注册账号",
			ReqUrl:           createUserUrl,
			Method:           "POST",
			Body:             createUserBody,
			Action:           defaultAction,
			CheckResponse:    createUserCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
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
			Action:           defaultAction,
			CheckResponse:    logoutCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var resp *http.Response
			if tc.Name == "退出登录" {
				tc.AccessToken = accessToken
			}
			// 执行行动
			resp = tc.Action(t, tc.ReqUrl, tc.Method, tc.AccessToken, tc.Body, client)

			// 关闭 Body 防止泄露
			defer resp.Body.Close()
			// 断言 HTTP 状态码
			assert.Equal(t, resp.StatusCode, tc.ExpectedHTTPCode)
			// 读取响应体
			respBytes, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)

			// 校验结果
			tc.CheckResponse(t, respBytes, tc.ExpectedBizCode, tc.ExpectedMsg)
		})

	}
}
