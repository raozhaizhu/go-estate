package user_test

import (
	"context"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	userCtrl "github.com/raozhaizhu/go-estate/internal/controller/user"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	mock_worker "github.com/raozhaizhu/go-estate/internal/worker/mock"
	mock_object_store "github.com/raozhaizhu/go-estate/pkg/object_store/mock"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
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

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

type testCase struct {
	name  string
	input interface{}
	// db,cache埋桩
	buildStubs func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker)
	// 对象存储埋桩
	buildObjectStoreStubs func(objectStoreMock *mock_object_store.MockStorageService)
	// 注入上下文
	buildCtx func() context.Context
	// action 执行动作
	action func(svc userCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error)
	// 校验数据
	checkResponse func(t *testing.T, results []interface{}, actualErr, expectedErr error)
	// expectedErr
	expectedErr error
}

func runTC(t *testing.T, testCases []testCase) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 初始化 store, svc, distributor
			storeMock := mock_db.NewMockStore(ctrl)
			cacheMock := mock_db.NewMockCache(ctrl)
			distributorMock := mock_worker.NewMockTaskDistributor(ctrl)
			tokenMakerMock := mock_token.NewMockMaker(ctrl)
			objectStoreMock := mock_object_store.NewMockStorageService(ctrl)
			asyncGo := func(ctx context.Context, logger *slog.Logger, name string, timeout time.Duration, fn func(ctx context.Context)) {
				fn(ctx)
			}
			deps := &app.Deps{
				Store:       storeMock,
				Cache:       cacheMock,
				ObjectStore: objectStoreMock,
				Config:      testConfig,
				TokenMaker:  tokenMakerMock,
				Distributor: distributorMock,
				AsyncRunner: asyncGo,
			}
			// 初始化 svc
			svc := user.New(deps)
			// 数据库埋桩
			tc.buildStubs(storeMock, cacheMock, distributorMock, tokenMakerMock)
			if tc.buildObjectStoreStubs != nil {
				tc.buildObjectStoreStubs(objectStoreMock)
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

type mockResult struct {
	rowsAffected    int64
	rowsAffectedErr error
}

func (m mockResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (m mockResult) RowsAffected() (int64, error) {
	return m.rowsAffected, m.rowsAffectedErr
}
