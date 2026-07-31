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

func TestGetDataByDay(t *testing.T) {
	// 预备数据
	username, validDate, expiredDate := util.RandomUsername(), dailyDataDomain.MinDate, dailyDataDomain.ExpiredDate
	validDateStr := dailyDataDomain.MinDateStr
	validInput := &dailyData.GetDataByDayInput{
		Username:   username,
		TargetDate: validDate,
	}
	invalidInput := &dailyData.GetDataByDayInput{
		Username:   username,
		TargetDate: expiredDate,
	}
	doubleMissTxParams := db.GetDataTxParams{
		RecordHit: false,
		DataHit:   false,
		Username:  username,
		QueryType: db.TypeGetDataByDay,
		StartDate: validDate,
		EndDate:   validDate,
		Points:    dailyData.GetDataByDayPoints,
	}
	dataHitTxParams := db.GetDataTxParams{
		RecordHit: false,
		DataHit:   true,
		Username:  username,
		QueryType: db.TypeGetDataByDay,
		StartDate: validDate,
		EndDate:   validDate,
		Points:    dailyData.GetDataByDayPoints,
	}
	recordHitTxParams := db.GetDataTxParams{
		RecordHit: true,
		DataHit:   false,
		Username:  username,
		QueryType: db.TypeGetDataByDay,
		StartDate: validDate,
		EndDate:   validDate,
		Points:    dailyData.GetDataByDayPoints,
	}
	dummyData := []db.DailyDatum{{ID: 1}}
	dummyDataRaw, _ := json.Marshal(dummyData)
	firstTxResult := db.GetDataTxResult{Data: dummyData, FirstTime: true}
	notFirstTxResult := db.GetDataTxResult{Data: dummyData, FirstTime: false}

	// 默认行为和校验逻辑
	defaultAction := func(svc dailyDataCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		dayInput, ok := input.(*dailyData.GetDataByDayInput)
		require.True(t, ok)
		data, err := svc.GetDataByDay(ctx, dayInput)
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

	// FailStubs
	doubleMissCache := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			GetRecordAndData(gomock.Any(), username, validDateStr, validDateStr).
			Return(false, nil, appError.ErrServerErr).Times(1)
	}
	failSetCacheDoubleMiss := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			SetRecordAndData(gomock.Any(), false, false, username, validDateStr, validDateStr,
				dummyData, dailyData.RecordExpireDuration, dailyData.DataExpireDuration).
			Return(appError.ErrServerErr).Times(1)
	}
	// SuccessStubs
	// Store 部分
	storeTxDoubleMissIsFirst := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().
			GetDataAndDeductPointsTx(gomock.Any(), doubleMissTxParams).
			Return(firstTxResult, nil).Times(1)

	}
	storeTxDoubleMissNotFirst := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().
			GetDataAndDeductPointsTx(gomock.Any(), doubleMissTxParams).
			Return(notFirstTxResult, nil).Times(1)
	}
	// Record Hit 说明 Cache 里有凭证, 则 DB 一定也有凭证, 绝对不是第一次查询
	storeTxRecordHit := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().
			GetDataAndDeductPointsTx(gomock.Any(), recordHitTxParams).
			Return(notFirstTxResult, nil).Times(1)
	}
	storeTxDataHitIsFirst := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().
			GetDataAndDeductPointsTx(gomock.Any(), dataHitTxParams).
			Return(firstTxResult, nil).Times(1)
	}
	storeTxDataHitNotFirst := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().
			GetDataAndDeductPointsTx(gomock.Any(), dataHitTxParams).
			Return(notFirstTxResult, nil).Times(1)
	}
	// Cache 部分
	cacheSetDoubleMiss := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			SetRecordAndData(gomock.Any(), false, false, username, validDateStr, validDateStr,
				dummyData, dailyData.RecordExpireDuration, dailyData.DataExpireDuration).
			Return(nil).Times(1)
	}
	cacheSetRecordHit := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			SetRecordAndData(gomock.Any(), true, false, username, validDateStr, validDateStr,
				dummyData, dailyData.RecordExpireDuration, dailyData.DataExpireDuration).
			Return(nil).Times(1)
	}
	cacheSetDataHit := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			SetRecordAndData(gomock.Any(), false, true, username, validDateStr, validDateStr,
				dummyData, dailyData.RecordExpireDuration, dailyData.DataExpireDuration).
			Return(nil).Times(1)
	}
	cacheNoSet := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			SetRecordAndData(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
	}
	cacheDoubleHit := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			GetRecordAndData(gomock.Any(), username, validDateStr, validDateStr).
			Return(true, dummyDataRaw, nil).Times(1)
	}
	cacheRecordHit := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			GetRecordAndData(gomock.Any(), username, validDateStr, validDateStr).
			Return(true, nil, nil).Times(1)
	}
	cacheDataHit := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().
			GetRecordAndData(gomock.Any(), username, validDateStr, validDateStr).
			Return(false, dummyDataRaw, nil).Times(1)
	}

	testCases := []testCase{
		{
			name:          "日期超出范围",
			input:         invalidInput,
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrTimeOutOfRange,
		},
		{
			name:  "[获取缓存发生内部错误] -> [DB查询失败,发生内部错误]",
			input: validInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager) {
				doubleMissCache(cacheMock)
				storeMock.EXPECT().
					GetDataAndDeductPointsTx(gomock.Any(), doubleMissTxParams).
					Return(db.GetDataTxResult{}, appError.ErrServerErr).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "[获取缓存发生内部错误] -> [DB查询成功] -> [是用户第一次查询数据]-> [写回缓存失败,不影响查询成功]",
			input: validInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager) {
				doubleMissCache(cacheMock)
				storeTxDoubleMissIsFirst(storeMock)
				failSetCacheDoubleMiss(cacheMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
		{
			name:  "[获取缓存发生内部错误] -> [DB查询成功] -> [是用户第一次查询数据]-> [写回缓存成功,查询成功]",
			input: validInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager) {
				doubleMissCache(cacheMock)
				storeTxDoubleMissIsFirst(storeMock)
				cacheSetDoubleMiss(cacheMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
		{
			name:  "[获取缓存发生内部错误] -> [DB查询成功] -> [并非用户第一次查询数据]-> [写回缓存成功,查询成功]",
			input: validInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager) {
				doubleMissCache(cacheMock)
				storeTxDoubleMissNotFirst(storeMock)
				cacheSetDoubleMiss(cacheMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
		{
			name:  "[缓存DoubleHit,用户查询过且数据已缓存] -> 不写回缓存,查询成功",
			input: validInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager) {
				cacheDoubleHit(cacheMock)
				cacheNoSet(cacheMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
		{
			name:  "[缓存SingleHit,凭证已缓存,数据未缓存] -> [DB查询成功, 仅获取数据] -> 写回缓存成功,查询成功",
			input: validInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager) {
				cacheRecordHit(cacheMock)
				storeTxRecordHit(storeMock)
				cacheSetRecordHit(cacheMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
		{
			name:  "[缓存SingleHit,凭证未缓存,数据已缓存] -> [DB查询成功, 仅获取凭证] -> [是初次查询] -> 写回缓存成功,查询成功",
			input: validInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager) {
				cacheDataHit(cacheMock)
				storeTxDataHitIsFirst(storeMock)
				cacheSetDataHit(cacheMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
		{
			name:  "[缓存SingleHit,凭证未缓存,数据已缓存] -> [DB查询成功, 仅获取凭证] -> [不是初次查询] -> 写回缓存成功,查询成功",
			input: validInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager) {
				cacheDataHit(cacheMock)
				storeTxDataHitNotFirst(storeMock)
				cacheSetDataHit(cacheMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
	}

	runTC(t, testCases)

}
