package dailyData_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	dailyDataCtrl "github.com/raozhaizhu/go-estate/internal/controller/daily_data"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	dailyDataDomain "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	myWebsocket "github.com/raozhaizhu/go-estate/internal/my_websocket"
	dailyData "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	"github.com/raozhaizhu/go-estate/internal/util"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/stretchr/testify/require"
)

func TestGetDataByPeriod(t *testing.T) {
	// 预备数据
	username, minDate, maxDate, expiredDate := util.RandomUsername(), dailyDataDomain.MinDate, dailyDataDomain.MaxDate, dailyDataDomain.ExpiredDate
	minDateStr, maxDateStr := dailyDataDomain.MinDateStr, dailyDataDomain.MaxDateStr
	validInput := &dailyData.GetDataByPeriodInput{
		Username:  username,
		StartDate: minDate,
		EndDate:   maxDate,
	}
	startEqualEndInput := &dailyData.GetDataByPeriodInput{
		Username:  username,
		StartDate: minDate,
		EndDate:   minDate,
	}
	expiredInput := &dailyData.GetDataByPeriodInput{
		Username:  username,
		StartDate: minDate,
		EndDate:   expiredDate,
	}
	dummyData := []db.DailyDatum{{ID: 1}}
	dummyDataRaw, _ := json.Marshal(dummyData)

	// 默认行为和校验逻辑
	defaultAction := func(svc dailyDataCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		periodInput, ok := input.(*dailyData.GetDataByPeriodInput)
		require.True(t, ok)
		data, err := svc.GetDataByPeriod(ctx, periodInput)
		return []interface{}{data}, err
	}
	failCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.Error(t, actualErr)
		require.ErrorIs(t, actualErr, expectedErr)
		require.NotNil(t, results)
		require.Len(t, results, 1)

		data, ok := results[0].([]db.DailyDatum)
		require.True(t, ok)
		require.Nil(t, data)
		require.Empty(t, data)
	}
	successCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.NoError(t, actualErr)
		require.NotNil(t, results)
		require.Len(t, results, 1)

		data, ok := results[0].([]db.DailyDatum)
		require.True(t, ok)
		require.NotNil(t, data)
		require.NotEmpty(t, data)
		require.Equal(t, dummyData[0].ID, data[0].ID)
	}

	// 成功桩函数
	cacheNoSet := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			SetRecordAndData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	}
	cacheDoubleHit := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			GetRecordAndData(gomock.Any(), username, minDateStr, maxDateStr).
			Return(true, dummyDataRaw, nil).Times(1)
	}

	testCases := []testCase{
		{
			name:          "开始时间等于结束时间",
			input:         startEqualEndInput,
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrBadTimerOrder,
		},
		{
			name:          "日期超出范围",
			input:         expiredInput,
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrTimeOutOfRange,
		},
		{
			name:  "参数正确,直接命中缓存,返回数据",
			input: validInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager) {
				cacheDoubleHit(cacheMock)
				cacheNoSet(cacheMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
	}

	runTC(t, testCases)

}
