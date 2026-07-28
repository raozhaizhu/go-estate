package middleware_test

import (
	"context"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	"github.com/raozhaizhu/go-estate/internal/util"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/raozhaizhu/go-estate/pkg/validator"
)

var (
	testConfig util.Config
	testLogger *slog.Logger
)

func TestMain(m *testing.M) {
	testConfig = util.InitConfig("../..")
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
	buildStubs func(tokenMakerMock *mock_token.MockMaker, cacheMock *mock_db.MockCache)
	// 执行服务
	action func(t *testing.T, reqUrl string, body interface{}, router *gin.Engine, writer *httptest.ResponseRecorder, customData map[string]any, tokenMakerMock *mock_token.MockMaker, ctx *gin.Context, cacheMock *mock_db.MockCache)
	// 校验数据
	checkResponse func(t *testing.T, writer *httptest.ResponseRecorder, expectedHTTPCode, expectedBizCode int, expectedMsg string, ctx *gin.Context)
	// expectedHTTPCode
	expectedHTTPCode int
	// expectedBizCode
	expectedBizCode int
	// expectedMsg
	expectedMsg string
}

var capturedCtx context.Context

func runTC(t *testing.T, testCases []testCase) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 开启测试
			gin.SetMode(gin.TestMode)
			writer := httptest.NewRecorder()
			ctx, router := gin.CreateTestContext(writer)

			// 模拟请求
			tokenMakerMock := mock_token.NewMockMaker(ctrl)
			cacheMock := mock_db.NewMockCache(ctrl)
			tc.buildStubs(tokenMakerMock, cacheMock)

			// 执行行动
			tc.action(t, tc.reqUrl, tc.body, router, writer, tc.customData, tokenMakerMock, ctx, cacheMock)

			// 校验结果
			tc.checkResponse(t, writer, tc.expectedHTTPCode, tc.expectedBizCode, tc.expectedMsg, ctx)
		})
	}
}
