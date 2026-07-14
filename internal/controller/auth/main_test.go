package auth_test

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mock_controller "github.com/raozhaizhu/go-estate/internal/controller/auth/mock"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	"github.com/raozhaizhu/go-estate/internal/util"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/raozhaizhu/go-estate/pkg/validator"
)

var (
	testConfig util.Config
	testLogger *slog.Logger
)

func TestMain(m *testing.M) {
	testConfig = util.InitConfig("../../..")
	testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))
	validator.InitTrans()

	os.Exit(m.Run())
}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

type testCase struct {
	name       string
	reqUrl     string
	body       interface{}
	customData map[string]any
	// svc埋桩
	buildStubs func(svcMock *mock_controller.MockService, tokenMakerMock *mock_token.MockMaker)
	// 执行服务
	action func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any)
	// 校验数据
	checkResponse func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string)
	// expectedHTTPCode
	expectedHTTPCode int
	// expectedBizCode
	expectedBizCode int
	// expectedMsg
	expectedMsg string
}

func runTC(t *testing.T, testCases []testCase) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 构建 svc, tokenMaker
			svcMock := mock_controller.NewMockService(ctrl)
			tokenMakerMock := mock_token.NewMockMaker(ctrl)
			//  svc打桩
			tc.buildStubs(svcMock, tokenMakerMock)

			//  初始化 recorder,router,ctx
			writer := httptest.NewRecorder()
			svcs := delivery.Services{
				AuthSvc: svcMock,
			}
			deps := app.Deps{
				Config:     testConfig,
				TokenMaker: tokenMakerMock,
				Logger:     testLogger,
			}
			router := delivery.SetupRouter(svcs, deps)

			// 执行行动
			tc.action(t, tc.reqUrl, tc.body, router, writer, tc.customData)

			// 校验结果
			tc.checkResponse(t, writer, tc.expectedHTTPCode, tc.expectedBizCode, tc.expectedMsg)
		})
	}
}
