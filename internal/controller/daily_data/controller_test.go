package dailyData

// import (
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"github.com/gin-gonic/gin"
// 	"github.com/golang/mock/gomock"

// 	mock_controller "github.com/raozhaizhu/go-estate/internal/controller/daily_data/mock"
// 	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
// 	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
// 	service "github.com/raozhaizhu/go-estate/internal/service/daily_data"
// 	"github.com/stretchr/testify/assert"
// )

// func TestGetDataByDay(t *testing.T) {
// 	type testCase struct {
// 		name          string
// 		inputQuery    string
// 		buildStubs    func(svc *mock_controller.MockDailyDataQuerier)
// 		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
// 	}
// 	emptyDate := ""
// 	validDate := dailyData.MinDate

// 	validDateQuery := fmt.Sprintf("date=%s", dailyData.MinDateFormatted)
// 	invalidFormatQuery := fmt.Sprintf("date=%s", dailyData.SmallInvalidDateFormatted)

// 	expiredDate := dailyData.ExpiredDate
// 	expiredDateFormattedQuery := fmt.Sprintf("date=%s", dailyData.ExpiredDateFormatted)

// 	dummyData := []db.DailyDatum{
// 		{
// 			ID: 1,
// 		},
// 	}

// 	testCases := []testCase{
// 		{
// 			name:       "ErrEmptyDate",
// 			inputQuery: emptyDate,
// 			buildStubs: func(svc *mock_controller.MockDailyDataQuerier) {}, // 直接拦截, 不触及数据库
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)      // 校验状态码400
// 				assert.JSONEq(t, ErrEmptyDateJson, recorder.Body.String()) // 校验 json 符合预期
// 			},
// 		},
// 		{
// 			name:       "ErrInvalidDateFormStr",
// 			inputQuery: invalidFormatQuery,
// 			buildStubs: func(svc *mock_controller.MockDailyDataQuerier) {}, // 直接拦截, 不触及数据库
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)            // 校验状态码400
// 				assert.JSONEq(t, ErrInvalidDateFormJson, recorder.Body.String()) // 校验 json 符合预期
// 			},
// 		},
// 		{
// 			name:       "ErrTimeOutOfRange",
// 			inputQuery: expiredDateFormattedQuery,
// 			buildStubs: func(svc *mock_controller.MockDailyDataQuerier) {
// 				svc.EXPECT().GetDataByDay(gomock.Any(),
// 					service.GetDataByDayParams{expiredDate}).
// 					Return(nil, service.ErrTimeOutOfRange).Times(1)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)                   // 校验状态码400
// 				assert.JSONEq(t, service.ErrTimeOutOfRangeJson, recorder.Body.String()) // 校验 json 符合预期
// 			},
// 		},
// 		{
// 			name:       "Success",
// 			inputQuery: validDateQuery,
// 			buildStubs: func(svc *mock_controller.MockDailyDataQuerier) {
// 				svc.EXPECT().GetDataByDay(gomock.Any(),
// 					service.GetDataByDayParams{validDate}).
// 					Return(dummyData, nil).Times(1)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, recorder.Code)
// 				dateBytes, err := json.Marshal(dummyData)
// 				assert.NoError(t, err)                                      // 校验状态码 200
// 				assert.JSONEq(t, string(dateBytes), recorder.Body.String()) // 校验 json 符合预期
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			// 构建 svc, ctrl
// 			svcMock := mock_controller.NewMockDailyDataQuerier(ctrl)
// 			controller := NewDailyDataController(svcMock)
// 			//  svc打桩
// 			tc.buildStubs(svcMock)
// 			//  初始化 recorder, ctx, router
// 			w := httptest.NewRecorder()
// 			ctx, router := gin.CreateTestContext(w)
// 			// 挂载 api
// 			router.GET(dailyData.DailyDataDayUrl, controller.GetDataByDay)
// 			// 构建 req
// 			reqUrl := fmt.Sprintf("%s?%s", dailyData.DailyDataDayUrl, tc.inputQuery)
// 			// t.Logf("====== 🚀 当前请求的 URL 是: %s ======", reqUrl)
// 			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
// 			assert.NoError(t, err)
// 			// writer 服务 req
// 			ctx.Request = req
// 			router.ServeHTTP(w, req)
// 			// 检查结果
// 			tc.checkResponse(t, w)
// 		})
// 	}

// }

// func TestGetDataByPeriod(t *testing.T) {
// 	type testCase struct {
// 		name          string
// 		inputQuery    string
// 		buildStubs    func(svc *mock_controller.MockDailyDataQuerier)
// 		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
// 	}
// 	emptyDate := ""
// 	startDate := dailyData.MinDate
// 	endDate := dailyData.MaxDate

// 	validDateQuery := fmt.Sprintf("start=%s&end=%s", dailyData.MinDateFormatted, dailyData.MaxDateFormatted)
// 	invalidFormatQuery := fmt.Sprintf("start=%s&end=%s", dailyData.SmallInvalidDateFormatted, dailyData.BigInvalidDateFormatted)

// 	expiredDate := dailyData.ExpiredDate
// 	expiredDateFormattedQuery := fmt.Sprintf("start=%s&end=%s", dailyData.MinDateFormatted, dailyData.ExpiredDateFormatted)

// 	dummyData := []db.DailyDatum{
// 		{
// 			ID: 1,
// 		},
// 	}

// 	testCases := []testCase{
// 		{
// 			name:       "ErrEmptyDate",
// 			inputQuery: emptyDate,
// 			buildStubs: func(svc *mock_controller.MockDailyDataQuerier) {}, // 直接拦截, 不触及数据库
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)      // 校验状态码400
// 				assert.JSONEq(t, ErrEmptyDateJson, recorder.Body.String()) // 校验 json 符合预期
// 			},
// 		},
// 		{
// 			name:       "ErrInvalidDateFormStr",
// 			inputQuery: invalidFormatQuery,
// 			buildStubs: func(svc *mock_controller.MockDailyDataQuerier) {}, // 直接拦截, 不触及数据库
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)            // 校验状态码400
// 				assert.JSONEq(t, ErrInvalidDateFormJson, recorder.Body.String()) // 校验 json 符合预期
// 			},
// 		},
// 		{
// 			name:       "ErrTimeOutOfRangeJson",
// 			inputQuery: expiredDateFormattedQuery,
// 			buildStubs: func(svc *mock_controller.MockDailyDataQuerier) {
// 				svc.EXPECT().GetDataByPeriod(gomock.Any(),
// 					service.GetDataByPeriodParams{StartDate: startDate, EndDate: expiredDate}).
// 					Return(nil, service.ErrTimeOutOfRange).Times(1)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusBadRequest, recorder.Code)                   // 校验状态码400
// 				assert.JSONEq(t, service.ErrTimeOutOfRangeJson, recorder.Body.String()) // 校验 json 符合预期
// 			},
// 		},
// 		{
// 			name:       "Success",
// 			inputQuery: validDateQuery,
// 			buildStubs: func(svc *mock_controller.MockDailyDataQuerier) {
// 				svc.EXPECT().GetDataByPeriod(gomock.Any(),
// 					service.GetDataByPeriodParams{StartDate: startDate, EndDate: endDate}).
// 					Return(dummyData, nil).Times(1)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, recorder.Code)
// 				dateBytes, err := json.Marshal(dummyData)
// 				assert.NoError(t, err)                                      // 校验状态码 200
// 				assert.JSONEq(t, string(dateBytes), recorder.Body.String()) // 校验 json 符合预期
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			// 构建 svc, ctrl
// 			svcMock := mock_controller.NewMockDailyDataQuerier(ctrl)
// 			controller := NewDailyDataController(svcMock)
// 			//  svc打桩
// 			tc.buildStubs(svcMock)
// 			//  初始化 recorder, ctx, router
// 			w := httptest.NewRecorder()
// 			ctx, router := gin.CreateTestContext(w)
// 			// 挂载 api
// 			router.GET(dailyData.DailyDataPeriodUrl, controller.GetDataByPeriod)
// 			// 构建 req
// 			reqUrl := fmt.Sprintf("%s?%s", dailyData.DailyDataPeriodUrl, tc.inputQuery)
// 			// t.Logf("====== 🚀 当前请求的 URL 是: %s ======", reqUrl)
// 			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
// 			assert.NoError(t, err)
// 			// writer 服务 req
// 			ctx.Request = req
// 			router.ServeHTTP(w, req)
// 			// 检查结果
// 			tc.checkResponse(t, w)
// 		})
// 	}

// }

// func TestGetAllData(t *testing.T) {
// 	type testCase struct {
// 		name          string
// 		buildStubs    func(svc *mock_controller.MockDailyDataQuerier)
// 		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
// 	}

// 	dummyData := []db.DailyDatum{
// 		{
// 			ID: 1,
// 		},
// 	}

// 	testCases := []testCase{
// 		{
// 			name: "Success",
// 			buildStubs: func(svc *mock_controller.MockDailyDataQuerier) {
// 				svc.EXPECT().GetAllData(gomock.Any()).Return(dummyData, nil).Times(1)
// 			},
// 			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
// 				assert.Equal(t, http.StatusOK, recorder.Code)
// 				dateBytes, err := json.Marshal(dummyData)
// 				assert.NoError(t, err)                                      // 校验状态码 200
// 				assert.JSONEq(t, string(dateBytes), recorder.Body.String()) // 校验 json 符合预期
// 			},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			// 构建 svc, ctrl
// 			svcMock := mock_controller.NewMockDailyDataQuerier(ctrl)
// 			controller := NewDailyDataController(svcMock)
// 			//  svc打桩
// 			tc.buildStubs(svcMock)
// 			//  初始化 recorder, ctx, router
// 			w := httptest.NewRecorder()
// 			ctx, router := gin.CreateTestContext(w)
// 			// 挂载 api
// 			router.GET(dailyData.DailyDataAllUrl, controller.GetAllData)
// 			// 构建 req
// 			reqUrl := fmt.Sprintf("%s", dailyData.DailyDataAllUrl)
// 			// t.Logf("====== 🚀 当前请求的 URL 是: %s ======", reqUrl)
// 			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
// 			assert.NoError(t, err)
// 			// writer 服务 req
// 			ctx.Request = req
// 			router.ServeHTTP(w, req)
// 			// 检查结果
// 			tc.checkResponse(t, w)
// 		})
// 	}

// }
