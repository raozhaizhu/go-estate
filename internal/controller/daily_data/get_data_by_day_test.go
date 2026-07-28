package dailyData_test

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
	"github.com/stretchr/testify/require"

	dailyDataCtrl "github.com/raozhaizhu/go-estate/internal/controller/daily_data"
	mock_controller "github.com/raozhaizhu/go-estate/internal/controller/daily_data/mock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	dailyDataDomain "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	dailyData "github.com/raozhaizhu/go-estate/internal/service/daily_data"

	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
)

/** ====================================================================================
 * 🏁 TestGetDataByDay
 * =====================================================================================
 */

func TestGetDataByDay(t *testing.T) {
	// 预备数据
	baseUrl := delivery.CurrAPI + "/daily_data/day"
	username := util.RandomUsername()
	malformedDateStr := dailyDataDomain.MalformedDateStr
	validDate, validDateStr := dailyDataDomain.MinDate, dailyDataDomain.MinDateStr
	validInput := dailyData.GetDataByDayInput{
		TargetDate: validDate,
	}
	expiredDate, expiredDateStr := dailyDataDomain.ExpiredDate, dailyDataDomain.ExpiredDateStr
	expiredInput := dailyData.GetDataByDayInput{
		TargetDate: expiredDate,
	}
	dummyDBData := []db.DailyDatum{{ID: 1}}
	deviceID, userAgent, _, clientIPWithPort := testUtil.DeviceID, testUtil.UserAgent, testUtil.ClientIp, testUtil.ClientIpWithPort
	_, authorization := testUtil.AccessStr, testUtil.AuthorizationAccessToken
	correctHeader := map[string]any{
		"X-Device-ID":   deviceID,
		"User-Agent":    userAgent,
		"Authorization": authorization,
	}
	accessPayload := &token.Payload{
		Username:  username,
		Role:      userDomain.RoleUser,
		TokenType: token.TokenTypeAccessToken,
	}

	// 默认执行逻辑
	defaultAction := func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any) {
		req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
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
		var results response.Result[dailyDataCtrl.DailyDataList]
		// 反序列化结果
		err := json.Unmarshal(writer.Body.Bytes(), &results)
		require.NoError(t, err)
		require.Equal(t, appError.CodeSuccess, writer.Code)
		require.Equal(t, expectedBizCode, results.Code)
		require.Equal(t, expectedMsg, results.Msg)
	}
	failCheckResponse := func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string) {
		var results response.Result[dailyDataCtrl.DailyDataList]
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
			name:       "带正确格式日期访问, 获取数据成功",
			reqUrl:     baseUrl + "?date=" + validDateStr,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				stubVerifyTokenSuccess(tokenMakerMock)
				svcMock.EXPECT().GetDataByDay(gomock.Any(), validInput).
					Return(dummyDBData, nil).Times(1)
			},
			action:           defaultAction,
			checkResponse:    successCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
		},
		{
			name:       "query 日期格式错误",
			reqUrl:     baseUrl + "?date=" + malformedDateStr,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				stubVerifyTokenSuccess(tokenMakerMock)
				svcMock.EXPECT().GetDataByDay(gomock.Any(), gomock.Any()).Times(0)
			},
			action:           defaultAction,
			checkResponse:    successCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.CodeInvalidParam,
			expectedMsg:      "Date的格式必须是2006-01-02",
		},
		{
			name:       "query 日期超出范围",
			reqUrl:     baseUrl + "?date=" + expiredDateStr,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				stubVerifyTokenSuccess(tokenMakerMock)
				svcMock.EXPECT().GetDataByDay(gomock.Any(), expiredInput).
					Return([]db.DailyDatum{}, appError.ErrTimeOutOfRange).Times(1)
			},
			action:           defaultAction,
			checkResponse:    failCheckResponse,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrTimeOutOfRange.Code,
			expectedMsg:      "查询日期超出范围",
		},
		{
			name:       "带正确格式日期访问, svc 抛出底层错误, ctrl 兜底处理且不暴露内部信息",
			reqUrl:     baseUrl + "?date=" + validDateStr,
			customData: correctHeader,
			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
				stubVerifyTokenSuccess(tokenMakerMock)
				svcMock.EXPECT().GetDataByDay(gomock.Any(), validInput).
					Return([]db.DailyDatum{}, fmt.Errorf("从数据库获取 Session 失败: %w", sql.ErrNoRows)).Times(1)
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

// func TestGetAllData(t *testing.T) {

// 	dummyData := []db.DailyDatum{
// 		{
// 			ID: 1,
// 		},
// 	}

// 	testCases := []testCase{
// 		{
// 			name: "无 Token 访问 GetAllData",
// 			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
// 				svcMock.EXPECT().GetAllData(gomock.Any()).Times(0)
// 			},
// 			expectedHTTPCode: 401,
// 			expectedBizCode:  appError.ErrAuthRequired.Code,
// 			expectedMsg:      appError.ErrAuthRequired.Msg,
// 		},
// 		{
// 			name: "User 访问 GetAllData",
// 			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
// 				svcMock.EXPECT().GetAllData(gomock.Any()).Times(0)
// 			},
// 			expectedHTTPCode: 401,
// 			expectedBizCode:  appError.ErrAuthPermissionDenied.Code,
// 			expectedMsg:      appError.ErrAuthPermissionDenied.Msg,
// 			payload:          userPayload,
// 		},
// 		{
// 			name: "Vip 访问 GetAllData",
// 			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
// 				svcMock.EXPECT().GetAllData(gomock.Any()).Times(0)
// 			},
// 			expectedHTTPCode: 401,
// 			expectedBizCode:  appError.ErrAuthPermissionDenied.Code,
// 			expectedMsg:      appError.ErrAuthPermissionDenied.Msg,
// 			payload:          vipPayload,
// 		},
// 		{
// 			name: "Admin 访问 GetAllData",
// 			buildStubs: func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker) {
// 				svcMock.EXPECT().GetAllData(gomock.Any()).Return(dummyData, nil).Times(1)
// 			},
// 			expectedHTTPCode: 200,
// 			expectedBizCode:  200,
// 			expectedMsg:      "success",
// 			payload:          adminPayload,
// 		},
// 	}

// }
