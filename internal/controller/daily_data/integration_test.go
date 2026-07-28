// go:build integration

package dailyData_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"testing"

	dailyData "github.com/raozhaizhu/go-estate/internal/controller/daily_data"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	dailyDataDomain "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDailyDataHappyPathFlow 仅测试 Auth 模块的正确路径
func TestDailyDataHappyPathFlow(t *testing.T) {
	// 启动服务器, 配置 cookie
	testServer, deps := testUtil.SetupIntegrationTest(t)
	defer testServer.Close()
	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := testServer.Client()
	client.Jar = jar

	// 准备数据
	user, userPassword := testUtil.CreateRandomUser(t, deps.Store, userDomain.RoleUser, "")
	vip, vipPassword := testUtil.CreateRandomUser(t, deps.Store, userDomain.RoleVip, "")
	admin, adminPassword := testUtil.CreateRandomUser(t, deps.Store, userDomain.RoleAdmin, "")
	username, vipName, adminName := user.Username, vip.Username, admin.Username
	dateStr := util.GetRandomDayInRange().Format(dailyDataDomain.DateFormat)
	start, end := util.GetRandom2DayInRange()
	startStr, endStr := start.Format(dailyDataDomain.DateFormat), end.Format(dailyDataDomain.DateFormat)

	loginUserBody := map[string]string{
		"username": username,
		"password": userPassword,
	}
	loginVipBody := map[string]string{
		"username": vipName,
		"password": vipPassword,
	}
	loginAdminBody := map[string]string{
		"username": adminName,
		"password": adminPassword,
	}

	deviceID, userAgent := testUtil.DeviceID, testUtil.UserAgent
	// Url
	loginUrl := testServer.URL + delivery.AuthLoginAPI
	getByDayUrl := testServer.URL + delivery.DailyDataGetDayAPI + "?date=" + dateStr
	getByPeriodUrl := testServer.URL + delivery.DailyDataGetPeriodAPI + "?start=" + startStr + "&end=" + endStr
	getAllUrl := testServer.URL + delivery.DailyDataGetAllAPI

	// 防止报错
	_, _, _, _ = loginVipBody, loginAdminBody, getByPeriodUrl, getAllUrl

	// 预备数据
	var userToken, vipToken, adminToken string

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

	getToken := func() string {
		return userToken
	}
	getVipToken := func() string {
		return vipToken
	}
	getAdminToken := func() string {
		return adminToken
	}

	loginUserCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
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
	loginVipCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*auth.DTO]
		err = json.Unmarshal(respBytes, &results)
		assert.NoError(t, err)
		// 校验业务代码
		assert.Equal(t, results.Code, expectedBizCode)
		// 校验用户信息一致
		assert.Equal(t, vipName, results.Data.UserInfo.Username)

		// 设置 accessToken
		vipToken = results.Data.AccessToken
	}
	loginAdminCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*auth.DTO]
		err = json.Unmarshal(respBytes, &results)
		assert.NoError(t, err)
		// 校验业务代码
		assert.Equal(t, results.Code, expectedBizCode)
		// 校验用户信息一致
		assert.Equal(t, adminName, results.Data.UserInfo.Username)

		// 设置 accessToken
		adminToken = results.Data.AccessToken
	}

	getDataCheckResponse := func(t *testing.T, respBytes []byte, expectedBizCode int, expectedMsg string) {
		// 序列化结构体
		var results response.Result[*[]dailyData.DailyDatumSchema]
		err = json.Unmarshal(respBytes, &results)
		assert.NoError(t, err)
		// 校验业务代码
		assert.Equal(t, results.Code, expectedBizCode)
		// 校验数据非空
		assert.NotEmpty(t, results.Data)
	}

	testCases := []testUtil.IntgTestCase{
		{
			Name:             "登录 User 账号",
			ReqUrl:           loginUrl,
			Method:           "POST",
			Body:             loginUserBody,
			Action:           defaultAction,
			CheckResponse:    loginUserCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "获取日数据",
			ReqUrl:           getByDayUrl,
			Method:           "GET",
			GetToken:         getToken,
			Action:           defaultAction,
			CheckResponse:    getDataCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "登录 VIP 账号",
			ReqUrl:           loginUrl,
			Body:             loginVipBody,
			Method:           "POST",
			Action:           defaultAction,
			CheckResponse:    loginVipCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
		{
			Name:             "获取周期数据",
			ReqUrl:           getByPeriodUrl,
			Method:           "GET",
			GetToken:         getVipToken,
			Action:           defaultAction,
			CheckResponse:    getDataCheckResponse,
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
			Name:             "获取全部数据",
			ReqUrl:           getAllUrl,
			Method:           "GET",
			GetToken:         getAdminToken,
			Action:           defaultAction,
			CheckResponse:    getDataCheckResponse,
			ExpectedHTTPCode: http.StatusOK,
			ExpectedBizCode:  200,
		},
	}

	testUtil.RunIntgTC(t, testCases, client)
}
