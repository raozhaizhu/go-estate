package dailyData

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	mock_service "github.com/raozhaizhu/go-estate/internal/service/daily_data/mock"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/stretchr/testify/assert"
)

func TestGetDataByDay(t *testing.T) {
	type testCase struct {
		name          string
		inputDate     time.Time
		buildStubs    func(store *mock_service.MockDailyDataStore)
		checkResponse func(t *testing.T, res []db.DailyDatum, err error)
	}
	validDate := dailyData.MaxDate
	invalidDate := dailyData.ExpiredDate
	dummyData := []db.DailyDatum{
		{
			ID:   1,
			Date: validDate,
		},
	}

	testCases := []testCase{
		{
			name:       "ErrTimeOutOfRange invalidDate",
			inputDate:  invalidDate,
			buildStubs: func(store *mock_service.MockDailyDataStore) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, appError.ErrTimeOutOfRange)
				assert.Empty(t, res)
			},
		},
		{
			name:      "Success",
			inputDate: validDate,
			buildStubs: func(store *mock_service.MockDailyDataStore) {
				store.EXPECT().GetDataByDay(gomock.Any(), validDate).Return(dummyData, nil).Times(1)
			},
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Equal(t, int32(1), res[0].ID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mock_service.NewMockDailyDataStore(ctrl)
			svc := New(mockStore)

			tc.buildStubs(mockStore)

			res, err := svc.GetDataByDay(context.Background(), GetDataByDayInput{TargetDate: tc.inputDate})

			tc.checkResponse(t, res, err)
		})
	}

}

func TestGetDataByPeriod(t *testing.T) {
	type testCase struct {
		name          string
		inputDate     db.GetDataByPeriodParams
		buildStubs    func(store *mock_service.MockDailyDataStore)
		checkResponse func(t *testing.T, res []db.DailyDatum, err error)
	}
	validStartDate := dailyData.MinDate
	validEndDate := dailyData.MaxDate
	invalidStartDate := dailyData.MinDate.AddDate(0, 0, -1)
	invalidEndDate := dailyData.ExpiredDate.AddDate(0, 0, 1)
	dummyData := []db.DailyDatum{
		{
			ID:   1,
			Date: validStartDate,
		},
	}

	testCases := []testCase{
		{
			name:       "ErrBadTimerOrder",
			inputDate:  db.GetDataByPeriodParams{StartDate: invalidEndDate, EndDate: validStartDate},
			buildStubs: func(store *mock_service.MockDailyDataStore) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, appError.ErrBadTimerOrder)
				assert.Empty(t, res)
			},
		},
		{
			name:       "ErrTimeOutOfRange validStartDate invalidEndDate",
			inputDate:  db.GetDataByPeriodParams{StartDate: validStartDate, EndDate: invalidEndDate},
			buildStubs: func(store *mock_service.MockDailyDataStore) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, appError.ErrTimeOutOfRange)
				assert.Empty(t, res)
			},
		},
		{
			name:       "ErrTimeOutOfRange invalidStartDate validEndDate",
			inputDate:  db.GetDataByPeriodParams{StartDate: invalidStartDate, EndDate: validEndDate},
			buildStubs: func(store *mock_service.MockDailyDataStore) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, appError.ErrTimeOutOfRange)
				assert.Empty(t, res)
			},
		},
		{
			name:       "ErrTimeOutOfRange invalidStartDate invalidEndDate",
			inputDate:  db.GetDataByPeriodParams{StartDate: invalidStartDate, EndDate: invalidEndDate},
			buildStubs: func(store *mock_service.MockDailyDataStore) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, appError.ErrTimeOutOfRange)
				assert.Empty(t, res)
			},
		},
		{
			name:      "Success",
			inputDate: db.GetDataByPeriodParams{StartDate: validStartDate, EndDate: validEndDate},
			buildStubs: func(store *mock_service.MockDailyDataStore) {
				store.EXPECT().GetDataByPeriod(gomock.Any(),
					db.GetDataByPeriodParams{StartDate: validStartDate, EndDate: validEndDate}).
					Return(dummyData, nil).Times(1)
			},
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Equal(t, int32(1), res[0].ID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mock_service.NewMockDailyDataStore(ctrl)
			svc := New(mockStore)

			tc.buildStubs(mockStore)
			res, err := svc.GetDataByPeriod(context.Background(),
				GetDataByPeriodInput{StartDate: tc.inputDate.StartDate, EndDate: tc.inputDate.EndDate})
			tc.checkResponse(t, res, err)
		})
	}

}

func TestGGetAllData(t *testing.T) {
	type testCase struct {
		name          string
		inputDate     time.Time
		buildStubs    func(store *mock_service.MockDailyDataStore)
		checkResponse func(t *testing.T, res []db.DailyDatum, err error)
	}

	dummyData := []db.DailyDatum{
		{ID: 1},
	}

	testCases := []testCase{
		{
			name: "Success",
			buildStubs: func(store *mock_service.MockDailyDataStore) {
				store.EXPECT().GetAllData(gomock.Any()).Return(dummyData, nil).Times(1)
			},
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Equal(t, int32(1), res[0].ID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mock_service.NewMockDailyDataStore(ctrl)
			svc := New(mockStore)

			tc.buildStubs(mockStore)

			res, err := svc.GetAllData(context.Background())

			tc.checkResponse(t, res, err)
		})
	}

}
