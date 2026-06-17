package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	mock_db "github.com/raozhaizhu/go-estate/db/mock"
	db "github.com/raozhaizhu/go-estate/db/sqlc"
	"github.com/raozhaizhu/go-estate/util"
	"github.com/stretchr/testify/assert"
)

func TestGetDataByDay(t *testing.T) {
	type testCase struct {
		name          string
		inputDate     time.Time
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, res []db.DailyDatum, err error)
	}
	validDate := util.MaxDate
	invalidDate := util.ExpiredDate
	dummyData := []db.DailyDatum{
		{
			ID:   1,
			Date: validDate,
		},
	}

	testCases := []testCase{
		{
			name:       "ErrTimeOutOfRange",
			inputDate:  invalidDate,
			buildStubs: func(store *mock_db.MockStore) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrTimeOutOfRange)
				assert.Empty(t, res)
			},
		},
		{
			name:      "Success",
			inputDate: validDate,
			buildStubs: func(store *mock_db.MockStore) {
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
		t.Run("tc.name", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storeMock := mock_db.NewMockStore(ctrl)
			srv := NewDailyDataService(storeMock)

			tc.buildStubs(storeMock)

			res, err := srv.GetDataByDay(context.Background(), tc.inputDate)

			tc.checkResponse(t, res, err)
		})
	}

}

func TestGetDataByPeriod(t *testing.T) {
	type testCase struct {
		name          string
		inputDate     db.GetDataByPeriodParams
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, res []db.DailyDatum, err error)
	}
	validStartDate := util.MinDate
	validEndDate := util.MaxDate
	invalidStartDate := util.MinDate.AddDate(0, 0, -1)
	invalidEndDate := util.ExpiredDate.AddDate(0, 0, 1)
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
			buildStubs: func(store *mock_db.MockStore) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrBadTimerOrder)
				assert.Empty(t, res)
			},
		},
		{
			name:       "ErrTimeOutOfRange",
			inputDate:  db.GetDataByPeriodParams{StartDate: validStartDate, EndDate: invalidEndDate},
			buildStubs: func(store *mock_db.MockStore) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrTimeOutOfRange)
				assert.Empty(t, res)
			},
		},
		{
			name:       "ErrTimeOutOfRange",
			inputDate:  db.GetDataByPeriodParams{StartDate: invalidStartDate, EndDate: validEndDate},
			buildStubs: func(store *mock_db.MockStore) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrTimeOutOfRange)
				assert.Empty(t, res)
			},
		},
		{
			name:       "ErrTimeOutOfRange",
			inputDate:  db.GetDataByPeriodParams{StartDate: invalidStartDate, EndDate: invalidEndDate},
			buildStubs: func(store *mock_db.MockStore) {}, // 直接拦截, 不触及数据库
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrTimeOutOfRange)
				assert.Empty(t, res)
			},
		},
		{
			name:      "Success",
			inputDate: db.GetDataByPeriodParams{StartDate: validStartDate, EndDate: validEndDate},
			buildStubs: func(store *mock_db.MockStore) {
				store.EXPECT().GetDataByPeriod(gomock.Any(), db.GetDataByPeriodParams{StartDate: validStartDate, EndDate: validEndDate}).Return(dummyData, nil).Times(1)
			},
			checkResponse: func(t *testing.T, res []db.DailyDatum, err error) {
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Equal(t, int32(1), res[0].ID)
			},
		},
	}

	for _, tc := range testCases {
		t.Run("tc.name", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storeMock := mock_db.NewMockStore(ctrl)
			srv := NewDailyDataService(storeMock)

			tc.buildStubs(storeMock)

			res, err := srv.GetDataByPeriod(context.Background(), tc.inputDate.StartDate, tc.inputDate.EndDate)

			tc.checkResponse(t, res, err)
		})
	}

}

func TestGGetAllData(t *testing.T) {
	type testCase struct {
		name          string
		inputDate     time.Time
		buildStubs    func(store *mock_db.MockStore)
		checkResponse func(t *testing.T, res []db.DailyDatum, err error)
	}

	dummyData := []db.DailyDatum{
		{ID: 1},
	}

	testCases := []testCase{
		{
			name: "Success",
			buildStubs: func(store *mock_db.MockStore) {
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
		t.Run("tc.name", func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			storeMock := mock_db.NewMockStore(ctrl)
			srv := NewDailyDataService(storeMock)

			tc.buildStubs(storeMock)

			res, err := srv.GetAllData(context.Background())

			tc.checkResponse(t, res, err)
		})
	}

}
