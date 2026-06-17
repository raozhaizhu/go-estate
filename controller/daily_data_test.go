package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mock_controller "github.com/raozhaizhu/go-estate/controller/mock"
	db "github.com/raozhaizhu/go-estate/db/sqlc"
	"github.com/raozhaizhu/go-estate/service"
	"github.com/raozhaizhu/go-estate/util"
	"github.com/stretchr/testify/assert"
)

func TestGetDataByDay(t *testing.T) {
	type testCase struct {
		name          string
		inputQuery    string
		buildStubs    func(store *mock_controller.MockDailyDataQuerier)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}
	emptyDate := ""
	validDate := util.MinDate

	validDateQuery := fmt.Sprintf("date=%s", util.MinDateFormatted)
	invalidFormatQuery := fmt.Sprintf("date=%s", util.SmallInvalidDateFormatted)

	expiredDate := util.ExpiredDate
	expiredDateFormattedQuery := fmt.Sprintf("date=%s", util.ExpiredDateFormatted)

	dummyData := []db.DailyDatum{
		{
			ID: 1,
		},
	}

	testCases := []testCase{
		{
			name:       "ErrEmptyDate",
			inputQuery: emptyDate,
			buildStubs: func(store *mock_controller.MockDailyDataQuerier) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)      // 校验状态码400
				assert.JSONEq(t, ErrEmptyDateJson, recorder.Body.String()) // 校验 json 符合预期
			},
		},
		{
			name:       "ErrInvalidDateFormStr",
			inputQuery: invalidFormatQuery,
			buildStubs: func(store *mock_controller.MockDailyDataQuerier) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)            // 校验状态码400
				assert.JSONEq(t, ErrInvalidDateFormJson, recorder.Body.String()) // 校验 json 符合预期
			},
		},
		{
			name:       "ErrQueryFailed",
			inputQuery: expiredDateFormattedQuery,
			buildStubs: func(store *mock_controller.MockDailyDataQuerier) {
				store.EXPECT().GetDataByDay(gomock.Any(), expiredDate).Return(nil, service.ErrTimeOutOfRange).Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)                   // 校验状态码400
				assert.JSONEq(t, service.ErrTimeOutOfRangeJson, recorder.Body.String()) // 校验 json 符合预期
			},
		},
		{
			name:       "Success",
			inputQuery: validDateQuery,
			buildStubs: func(store *mock_controller.MockDailyDataQuerier) {
				store.EXPECT().GetDataByDay(gomock.Any(), validDate).Return(dummyData, nil).Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				dateBytes, err := json.Marshal(dummyData)
				assert.NoError(t, err)                                      // 校验状态码 200
				assert.JSONEq(t, string(dateBytes), recorder.Body.String()) // 校验 json 符合预期
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 构建 srv, ctrl
			srvMock := mock_controller.NewMockDailyDataQuerier(ctrl)
			controller := NewDailyDataController(srvMock)
			//  srv打桩
			tc.buildStubs(srvMock)
			//  初始化 recorder, ctx, router
			w := httptest.NewRecorder()
			ctx, router := gin.CreateTestContext(w)
			// 挂载 api
			router.GET(util.DailyDataDayApiUrl, controller.GetDataByDay)
			// 构建 req
			reqUrl := fmt.Sprintf("%s?%s", util.DailyDataDayApiUrl, tc.inputQuery)
			t.Logf("====== 🚀 当前请求的 URL 是: %s ======", reqUrl)
			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
			assert.NoError(t, err)
			// writer 服务 req
			ctx.Request = req
			router.ServeHTTP(w, req)
			// 检查结果
			tc.checkResponse(t, w)
		})
	}

}

func TestGetDataByPeriod(t *testing.T) {
	type testCase struct {
		name          string
		inputQuery    string
		buildStubs    func(store *mock_controller.MockDailyDataQuerier)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}
	emptyDate := ""
	startDate := util.MinDate
	endDate := util.MaxDate

	validDateQuery := fmt.Sprintf("start=%s&end=%s", util.MinDateFormatted, util.MaxDateFormatted)
	invalidFormatQuery := fmt.Sprintf("start=%s&end=%s", util.SmallInvalidDateFormatted, util.BigInvalidDateFormatted)

	expiredDate := util.ExpiredDate
	expiredDateFormattedQuery := fmt.Sprintf("start=%s&end=%s", util.MinDateFormatted, util.ExpiredDateFormatted)

	dummyData := []db.DailyDatum{
		{
			ID: 1,
		},
	}

	testCases := []testCase{
		{
			name:       "ErrEmptyDate",
			inputQuery: emptyDate,
			buildStubs: func(store *mock_controller.MockDailyDataQuerier) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)      // 校验状态码400
				assert.JSONEq(t, ErrEmptyDateJson, recorder.Body.String()) // 校验 json 符合预期
			},
		},
		{
			name:       "ErrInvalidDateFormStr",
			inputQuery: invalidFormatQuery,
			buildStubs: func(store *mock_controller.MockDailyDataQuerier) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)            // 校验状态码400
				assert.JSONEq(t, ErrInvalidDateFormJson, recorder.Body.String()) // 校验 json 符合预期
			},
		},
		{
			name:       "ErrTimeOutOfRangeJson",
			inputQuery: expiredDateFormattedQuery,
			buildStubs: func(store *mock_controller.MockDailyDataQuerier) {
				store.EXPECT().GetDataByPeriod(gomock.Any(), startDate, expiredDate).Return(nil, service.ErrTimeOutOfRange).Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, recorder.Code)                   // 校验状态码400
				assert.JSONEq(t, service.ErrTimeOutOfRangeJson, recorder.Body.String()) // 校验 json 符合预期
			},
		},
		{
			name:       "Success",
			inputQuery: validDateQuery,
			buildStubs: func(store *mock_controller.MockDailyDataQuerier) {
				store.EXPECT().GetDataByPeriod(gomock.Any(), startDate, endDate).Return(dummyData, nil).Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				dateBytes, err := json.Marshal(dummyData)
				assert.NoError(t, err)                                      // 校验状态码 200
				assert.JSONEq(t, string(dateBytes), recorder.Body.String()) // 校验 json 符合预期
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 构建 srv, ctrl
			srvMock := mock_controller.NewMockDailyDataQuerier(ctrl)
			controller := NewDailyDataController(srvMock)
			//  srv打桩
			tc.buildStubs(srvMock)
			//  初始化 recorder, ctx, router
			w := httptest.NewRecorder()
			ctx, router := gin.CreateTestContext(w)
			// 挂载 api
			router.GET(util.DailyDataPeriodApiUrl, controller.GetDataByPeriod)
			// 构建 req
			reqUrl := fmt.Sprintf("%s?%s", util.DailyDataPeriodApiUrl, tc.inputQuery)
			t.Logf("====== 🚀 当前请求的 URL 是: %s ======", reqUrl)
			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
			assert.NoError(t, err)
			// writer 服务 req
			ctx.Request = req
			router.ServeHTTP(w, req)
			// 检查结果
			tc.checkResponse(t, w)
		})
	}

}

func TestGetAllData(t *testing.T) {
	type testCase struct {
		name          string
		buildStubs    func(store *mock_controller.MockDailyDataQuerier)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}

	dummyData := []db.DailyDatum{
		{
			ID: 1,
		},
	}

	testCases := []testCase{
		{
			name: "Success",
			buildStubs: func(store *mock_controller.MockDailyDataQuerier) {
				store.EXPECT().GetAllData(gomock.Any()).Return(dummyData, nil).Times(1)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, recorder.Code)
				dateBytes, err := json.Marshal(dummyData)
				assert.NoError(t, err)                                      // 校验状态码 200
				assert.JSONEq(t, string(dateBytes), recorder.Body.String()) // 校验 json 符合预期
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 构建 srv, ctrl
			srvMock := mock_controller.NewMockDailyDataQuerier(ctrl)
			controller := NewDailyDataController(srvMock)
			//  srv打桩
			tc.buildStubs(srvMock)
			//  初始化 recorder, ctx, router
			w := httptest.NewRecorder()
			ctx, router := gin.CreateTestContext(w)
			// 挂载 api
			router.GET(util.DailyDataAllApiUrl, controller.GetAllData)
			// 构建 req
			reqUrl := fmt.Sprintf("%s", util.DailyDataAllApiUrl)
			t.Logf("====== 🚀 当前请求的 URL 是: %s ======", reqUrl)
			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
			assert.NoError(t, err)
			// writer 服务 req
			ctx.Request = req
			router.ServeHTTP(w, req)
			// 检查结果
			tc.checkResponse(t, w)
		})
	}

}
