package dailyData_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	dailyDataCtrl "github.com/raozhaizhu/go-estate/internal/controller/daily_data"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	myWebsocket "github.com/raozhaizhu/go-estate/internal/my_websocket"
	"github.com/stretchr/testify/require"
)

func TestGetAllData(t *testing.T) {
	dummyData := []db.DailyDatum{
		{ID: 1},
	}
	// 默认行为和校验逻辑逻辑
	defaultAction := func(svc dailyDataCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		data, err := svc.GetAllData(ctx)
		return []interface{}{data}, err
	}
	successCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.NoError(t, actualErr)
		require.NotNil(t, results)
		require.Len(t, results, 1)

		data, ok := results[0].([]db.DailyDatum)
		require.True(t, ok)
		require.NotNil(t, data)
		require.NotEmpty(t, data)
	}

	testCases := []testCase{
		{
			name: "Success",
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager) {
				storeMock.EXPECT().GetAllData(gomock.Any()).Return(dummyData, nil).Times(1)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
	}

	runTC(t, testCases)
}
