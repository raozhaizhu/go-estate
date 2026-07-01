package dailyData

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"

	mock_controller "github.com/raozhaizhu/go-estate/internal/controller/daily_data/mock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	service "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDataByDay(t *testing.T) {
	type testCase struct {
		name             string
		date             string
		buildStubs       func(svc *mock_controller.MockService)
		expectedHTTPCode int
		expectedBizCode  int
		expectedMsg      string
		payload          *token.Payload
	}
	emptyDate := ""
	malformedDate := dailyData.MalformedSmallDate

	validDate := dailyData.MinDate
	validDateStr := dailyData.MinDateStr

	expiredDate := dailyData.ExpiredDate
	expiredDateStr := dailyData.ExpiredDateStr

	dummyData := []db.DailyDatum{
		{
			ID: 1,
		},
	}

	testCases := []testCase{
		{
			name: "无 Token 访问 GetDataByDay",
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByDay(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedHTTPCode: 401,
			expectedBizCode:  appError.ErrAuthRequired.Code,
			expectedMsg:      appError.ErrAuthRequired.Msg,
		},
		{
			name: "不填 query",
			date: emptyDate,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByDay(gomock.Any(), gomock.Any()).Times(0)
			}, expectedHTTPCode: 400,
			expectedBizCode: 40000,
			expectedMsg:     "Date为必填字段",
			payload:         userPayload,
		},
		{
			name: "query 日期格式错误",
			date: malformedDate,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByDay(gomock.Any(), gomock.Any()).Times(0)
			}, expectedHTTPCode: 400,
			expectedBizCode: 40000,
			expectedMsg:     "Date的格式必须是2006-01-02",
			payload:         userPayload,
		},
		{
			name: "query 日期超出范围",
			date: expiredDateStr,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByDay(gomock.Any(),
					service.GetDataByDayInput{expiredDate}).
					Return(nil, appError.ErrTimeOutOfRange).Times(1)
			},
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrTimeOutOfRange.Code,
			expectedMsg:      appError.ErrTimeOutOfRange.Msg,
			payload:          userPayload,
		},
		{
			name: "Success",
			date: validDateStr,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByDay(gomock.Any(),
					service.GetDataByDayInput{validDate}).
					Return(dummyData, nil).Times(1)
			},
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
			payload:          userPayload,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 构建 svc, ctrl
			svcMock := mock_controller.NewMockService(ctrl)
			controller := NewDailyDataController(svcMock)

			//  svc打桩
			tc.buildStubs(svcMock)

			//  初始化 recorder, ctx, router
			w := httptest.NewRecorder()
			ctx, router := setupTestRouter(controller, tc.payload)

			// 构建 req
			reqUrl := fmt.Sprintf("%s?date=%s", dailyData.DailyDataDayUrl, tc.date)
			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
			assert.NoError(t, err)

			// 服务 req
			ctx.Request = req
			router.ServeHTTP(w, req)

			// 校验状态码
			assert.Equal(t, tc.expectedHTTPCode, w.Code)

			// 将响应 json 反序列化为 actualResult
			var actualResult response.Result[[]db.DailyDatum]
			err = json.Unmarshal(w.Body.Bytes(), &actualResult)
			require.NoError(t, err)

			// 比较 actualResult
			assert.Equal(t, tc.expectedBizCode, actualResult.Code)
			assert.Contains(t, actualResult.Msg, tc.expectedMsg)

		})
	}

}

func TestGetDataByPeriod(t *testing.T) {
	type testCase struct {
		name             string
		inputQuery       string
		buildStubs       func(svc *mock_controller.MockService)
		expectedHTTPCode int
		expectedBizCode  int
		expectedMsg      string
		payload          *token.Payload
	}
	emptyDate := ""
	startDate := dailyData.MinDate
	endDate := dailyData.MaxDate

	validDateQuery := fmt.Sprintf("start=%s&end=%s", dailyData.MinDateStr, dailyData.MaxDateStr)
	invalidFormatQuery := fmt.Sprintf("start=%s&end=%s", dailyData.MalformedSmallDate, dailyData.MalformedBigDate)

	expiredDate := dailyData.ExpiredDate
	expiredDateFormattedQuery := fmt.Sprintf("start=%s&end=%s", dailyData.MinDateStr, dailyData.ExpiredDateStr)

	dummyData := []db.DailyDatum{
		{
			ID: 1,
		},
	}

	testCases := []testCase{
		{
			name: "无 Token 访问 GetDataByPeriod",
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByPeriod(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedHTTPCode: 401,
			expectedBizCode:  appError.ErrAuthRequired.Code,
			expectedMsg:      appError.ErrAuthRequired.Msg,
		},
		{
			name: "User 访问 GetDataByPeriod",
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByPeriod(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedHTTPCode: 401,
			expectedBizCode:  appError.ErrAuthPermissionDenied.Code,
			expectedMsg:      appError.ErrAuthPermissionDenied.Msg,
			payload:          userPayload,
		},
		{
			name:       "不填 query",
			inputQuery: emptyDate,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByPeriod(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedHTTPCode: 400,
			expectedBizCode:  40000,
			expectedMsg:      "Start为必填字段, End为必填字段",
			payload:          vipPayload,
		},
		{
			name:       "query 日期格式错误",
			inputQuery: invalidFormatQuery,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByPeriod(gomock.Any(), gomock.Any()).Times(0)
			},
			expectedHTTPCode: 400,
			expectedBizCode:  40000,
			expectedMsg:      "Start的格式必须是2006-01-02, End的格式必须是2006-01-02",
			payload:          vipPayload,
		},
		{
			name:       "query 日期超出限制范围",
			inputQuery: expiredDateFormattedQuery,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByPeriod(gomock.Any(),
					service.GetDataByPeriodInput{StartDate: startDate, EndDate: expiredDate}).
					Return(nil, appError.ErrTimeOutOfRange).Times(1)
			},
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrTimeOutOfRange.Code,
			expectedMsg:      appError.ErrTimeOutOfRange.Msg,
			payload:          vipPayload,
		},
		{
			name:       "Success",
			inputQuery: validDateQuery,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetDataByPeriod(gomock.Any(),
					service.GetDataByPeriodInput{StartDate: startDate, EndDate: endDate}).
					Return(dummyData, nil).Times(1)
			},
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
			payload:          vipPayload,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 构建 svc, ctrl
			mockService := mock_controller.NewMockService(ctrl)
			controller := NewDailyDataController(mockService)

			//  svc打桩
			tc.buildStubs(mockService)

			//  初始化 recorder, ctx, router
			w := httptest.NewRecorder()
			ctx, router := setupTestRouter(controller, tc.payload)

			// 构建 req
			reqUrl := fmt.Sprintf("%s?%s", dailyData.DailyDataPeriodUrl, tc.inputQuery)
			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
			assert.NoError(t, err)

			// 服务 req
			ctx.Request = req
			router.ServeHTTP(w, req)

			// 校验状态码
			assert.Equal(t, tc.expectedHTTPCode, w.Code)

			// 将响应 json 反序列化为 actualResult
			var actualResult response.Result[[]db.DailyDatum]
			err = json.Unmarshal(w.Body.Bytes(), &actualResult)
			require.NoError(t, err)

			// 比较 actualResult
			assert.Equal(t, tc.expectedBizCode, actualResult.Code)
			assert.Contains(t, actualResult.Msg, tc.expectedMsg)
		})
	}

}

func TestGetAllData(t *testing.T) {
	type testCase struct {
		name             string
		buildStubs       func(svc *mock_controller.MockService)
		expectedHTTPCode int
		expectedBizCode  int
		expectedMsg      string
		payload          *token.Payload
	}

	dummyData := []db.DailyDatum{
		{
			ID: 1,
		},
	}

	testCases := []testCase{
		{
			name: "无 Token 访问 GetAllData",
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetAllData(gomock.Any()).Times(0)
			},
			expectedHTTPCode: 401,
			expectedBizCode:  appError.ErrAuthRequired.Code,
			expectedMsg:      appError.ErrAuthRequired.Msg,
		},
		{
			name: "User 访问 GetAllData",
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetAllData(gomock.Any()).Times(0)
			},
			expectedHTTPCode: 401,
			expectedBizCode:  appError.ErrAuthPermissionDenied.Code,
			expectedMsg:      appError.ErrAuthPermissionDenied.Msg,
			payload:          userPayload,
		},
		{
			name: "Vip 访问 GetAllData",
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetAllData(gomock.Any()).Times(0)
			},
			expectedHTTPCode: 401,
			expectedBizCode:  appError.ErrAuthPermissionDenied.Code,
			expectedMsg:      appError.ErrAuthPermissionDenied.Msg,
			payload:          vipPayload,
		},
		{
			name: "Admin 访问 GetAllData",
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().GetAllData(gomock.Any()).Return(dummyData, nil).Times(1)
			},
			expectedHTTPCode: 200,
			expectedBizCode:  200,
			expectedMsg:      "success",
			payload:          adminPayload,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 构建 svc, ctrl
			mockService := mock_controller.NewMockService(ctrl)
			controller := NewDailyDataController(mockService)

			//  svc打桩
			tc.buildStubs(mockService)

			//  初始化 recorder, ctx, router
			w := httptest.NewRecorder()
			ctx, router := setupTestRouter(controller, tc.payload)

			// 构建 req
			reqUrl := fmt.Sprintf("%s", dailyData.DailyDataAllUrl)
			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
			assert.NoError(t, err)

			// 服务 req
			ctx.Request = req
			router.ServeHTTP(w, req)

			// 校验状态码
			assert.Equal(t, tc.expectedHTTPCode, w.Code)

			// 将响应 json 反序列化为 actualResult
			var actualResult response.Result[[]db.DailyDatum]
			err = json.Unmarshal(w.Body.Bytes(), &actualResult)
			require.NoError(t, err)

			// 比较 actualResult
			assert.Equal(t, tc.expectedBizCode, actualResult.Code)
			assert.Contains(t, actualResult.Msg, tc.expectedMsg)
		})
	}

}

func setupTestRouter(controller *Controller, mockPayload *token.Payload) (*gin.Context, *gin.Engine) {
	ctx, router := gin.CreateTestContext(httptest.NewRecorder())
	// 模拟路由结构
	apiGroup := router.Group("/api/v1")

	{
		dailyGroup := apiGroup.Group("/daily_data")
		{
			dailyGroup.Use(func(c *gin.Context) {
				if mockPayload != nil { // 传入了mockPayload, 进行配置
					c.Set(token.PayloadKey, mockPayload)
					c.Next()
				} else { // 没传入, 直接报错
					c.AbortWithStatusJSON(http.StatusUnauthorized, appError.ErrAuthRequired)
				}
			})

			// 至少是 User 才可以获取单日数据
			dailyGroup.GET("/day", middleware.RoleMiddleware(role.RoleAtLeastUser), response.Wrapper(controller.GetDataByDay))
			// 至少是 VIP 才可以获取范围数据
			dailyGroup.GET("/period", middleware.RoleMiddleware(role.RoleAtLeastVip), response.Wrapper(controller.GetDataByPeriod))
			// 至少是 Admin 才可以获取所有数据
			dailyGroup.GET("/all", middleware.RoleMiddleware(role.RoleAtLeastAdmin), response.Wrapper(controller.GetAllData))
		}
	}
	return ctx, router
}
