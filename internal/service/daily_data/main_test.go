package dailyData_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	dailyDataCtrl "github.com/raozhaizhu/go-estate/internal/controller/daily_data"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	myWebsocket "github.com/raozhaizhu/go-estate/internal/my_websocket"
	dailyData "github.com/raozhaizhu/go-estate/internal/service/daily_data"
	"github.com/raozhaizhu/go-estate/internal/util"
)

var (
	testConfig util.Config
	testLogger *slog.Logger
)

func TestMain(m *testing.M) {
	testConfig = util.InitTestConfig()
	testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

	os.Exit(m.Run())
}

type testCase struct {
	name  string
	input interface{}
	// buildStubs 埋桩
	buildStubs func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, wsManager *myWebsocket.Manager)
	// buildCtx 注入上下文
	buildCtx func() context.Context
	// action 执行动作
	action func(svc dailyDataCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error)
	// checkResponse 校验数据
	checkResponse func(t *testing.T, results []interface{}, actualErr, expectedErr error)
	// expectedErr 预期错误
	expectedErr error
}

func runTC(t *testing.T, testCases []testCase) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 初始化 deps
			storeMock := mock_db.NewMockStore(ctrl)
			cacheMock := mock_db.NewMockCache(ctrl)
			wsManager := myWebsocket.NewManager()
			deps := &app.Deps{
				Store:            storeMock,
				Cache:            cacheMock,
				Config:           testConfig,
				WebsocketManager: wsManager,
			}
			// 初始化 svc
			svc := dailyData.New(deps)
			// 数据库埋桩
			if tc.buildStubs != nil {
				tc.buildStubs(storeMock, cacheMock, wsManager)
			}
			// 注入上下文
			ctx := context.Background()
			if tc.buildCtx != nil {
				ctx = tc.buildCtx()
			}
			// 执行动作
			results, err := tc.action(svc, ctx, tc.input)
			// 校验一致性
			tc.checkResponse(t, results, err, tc.expectedErr)
		})
	}
}
