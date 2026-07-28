//go:build integration

package user_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"

	userController "github.com/raozhaizhu/go-estate/internal/controller/user"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
	userService "github.com/raozhaizhu/go-estate/internal/service/user"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	"github.com/stretchr/testify/require"
)

// TestUserHappyPathFlow 仅测试 User 模块的正确路径
func TestUserHappyPathFlow(t *testing.T) {
	// 启动服务器, 配置 cookie
	testServer, deps := testUtil.SetupIntegrationTest(t)
	defer testServer.Close()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := testServer.Client()
	client.Jar = jar

	// 准备数据
	vip, _ := testUtil.CreateRandomUser(t, deps.Store, userDomain.RoleVip, "")
	admin, adminPassword := testUtil.CreateRandomUser(t, deps.Store, userDomain.RoleAdmin, "")
	_, adminName := vip.Username, admin.Username

	username, password, email := util.RandomUsername(), util.RandomPassword(), util.RandomEmail()
	vipNameForCreate, vipPasswordForCreate, vipEmailForCreate := util.RandomUsername(), util.RandomPassword(), util.RandomEmail()
	vipNewPassword := util.RandomPassword()

	createUserBody := map[string]string{
		"username": username,
		"password": password,
		"email":    email,
	}
	createVipBody := map[string]string{
		"username": vipNameForCreate,
		"password": vipPasswordForCreate,
		"email":    vipEmailForCreate,
	}
	loginUserBody := map[string]string{
		"username": username,
		"password": password,
	}
	loginAdminBody := map[string]string{
		"username": adminName,
		"password": adminPassword,
	}
	updateUserBody := map[string]string{
		"password": vipNewPassword,
	}
	deviceID, userAgent := testUtil.DeviceID, testUtil.UserAgent
	// Url
	loginUrl := testServer.URL + delivery.AuthLoginAPI
	createUserUrl := testServer.URL + delivery.UserCreateNormalUserAPI
	getAvatarUploadUrl := testServer.URL + delivery.UserGetAvatarUploadUrlAPI + "?ext=webp"
	createVipUrl := testServer.URL + delivery.UserCreateVipAPI
	getUserUrl := testServer.URL + delivery.UserGetUserAPI + "/" + username
	updateUrl := testServer.URL + delivery.UserUpdateUserAPI + "/" + username

	// 预备凭证
	var userToken, adminToken string

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

	getAdminToken := func() string {
		return adminToken
	}

	createUserCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*userService.DTO]
		err = json.Unmarshal(respBytes, &results)
		require.NoError(t, err)
		// 校验业务代码
		require.Equal(t, results.Code, expectedBizCode)
		// 校验用户信息一致
		require.Equal(t, username, results.Data.Username)
	}
	getAvatarUploadUrlCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*userController.AvatarUploadData]
		err = json.Unmarshal(respBytes, &results)
		require.NoError(t, err)
		// 校验业务代码
		require.Equal(t, results.Code, expectedBizCode)
		// 校验非空
		require.NotEmpty(t, results.Data.PostUrl)
		require.NotEmpty(t, results.Data.FormData)
		require.NotEmpty(t, results.Data.ObjectKey)
	}
	loginUserCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*auth.DTO]
		err = json.Unmarshal(respBytes, &results)
		require.NoError(t, err)
		// 校验业务代码
		require.Equal(t, results.Code, expectedBizCode)
		// 校验用户信息一致
		require.Equal(t, username, results.Data.UserInfo.Username)

		// 设置 accessToken
		userToken = results.Data.AccessToken
	}
	loginAdminCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*auth.DTO]
		err = json.Unmarshal(respBytes, &results)
		require.NoError(t, err)
		// 校验业务代码
		require.Equal(t, results.Code, expectedBizCode)
		// 校验用户信息一致
		require.Equal(t, adminName, results.Data.UserInfo.Username)

		// 设置 accessToken
		adminToken = results.Data.AccessToken
	}
	createVipCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*userService.DTO]
		err = json.Unmarshal(respBytes, &results)
		require.NoError(t, err)
		// 校验业务代码
		require.Equal(t, results.Code, expectedBizCode)
		// 校验用户信息一致
		require.Equal(t, vipNameForCreate, results.Data.Username)
		require.Equal(t, userDomain.RoleVip, results.Data.Role)
	}
	getUserCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*userService.DTO]
		err = json.Unmarshal(respBytes, &results)
		require.NoError(t, err)
		// 校验业务代码
		require.Equal(t, results.Code, expectedBizCode)
		// 校验用户信息一致
		require.Equal(t, username, results.Data.Username)
	}

	testCases := []testUtil.IntgTestCase{
		{
			Name:             "创建账号",
			ReqUrl:           createUserUrl,
			Method:           "POST",
			Body:             createUserBody,
			Action:           defaultAction,
			CheckResponse:    createUserCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "上传头像",
			ReqUrl:           getAvatarUploadUrl,
			Method:           "GET",
			Action:           defaultAction,
			CheckResponse:    getAvatarUploadUrlCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "登录 Admin 账号",
			ReqUrl:           loginUrl,
			Body:             loginAdminBody,
			Method:           "POST",
			Action:           defaultAction,
			CheckResponse:    loginAdminCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "创建VIP",
			ReqUrl:           createVipUrl,
			Method:           "POST",
			Body:             createVipBody,
			GetToken:         getAdminToken,
			Action:           defaultAction,
			CheckResponse:    createVipCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "登录 User 账号",
			ReqUrl:           loginUrl,
			Body:             loginUserBody,
			Method:           "POST",
			Action:           defaultAction,
			CheckResponse:    loginUserCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "获取用户信息",
			ReqUrl:           getUserUrl,
			Method:           "GET",
			GetToken:         getUserToken,
			Action:           defaultAction,
			CheckResponse:    getUserCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "更新用户信息",
			ReqUrl:           updateUrl,
			Method:           "PATCH",
			Body:             updateUserBody,
			GetToken:         getUserToken,
			Action:           defaultAction,
			CheckResponse:    getUserCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
	}

	testUtil.RunIntgTC(t, testCases, client)
}
