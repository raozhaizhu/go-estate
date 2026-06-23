package user_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	userController "github.com/raozhaizhu/go-estate/internal/controller/user"
	mock_controller "github.com/raozhaizhu/go-estate/internal/controller/user/mock"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	service "github.com/raozhaizhu/go-estate/internal/service/user"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Type
 * =====================================================================================
 */
type getUserTC struct {
	name    string
	request userController.GetUserRequest
	// 设置认证头信息
	setupAuth func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	// svc埋桩
	buildStubs func(svc *mock_controller.MockService)
	// 校验数据
	checkResponse func(t *testing.T, tc getUserTC, result response.Result[service.UserDTO])
	// 请求地址
	reqUrl string
	// 响应代码
	expectedHTTPCode int
	expectedBizCode  int
}

func runGetUserTC(t *testing.T, testCases []getUserTC) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 初始化 mockService
			mockService := mock_controller.NewMockService(ctrl)
			router := delivery.SetupRouter(delivery.Services{
				UserSvc: mockService,
			}, testConfig, testTokenMaker)

			// svc埋桩
			tc.buildStubs(mockService)

			//  初始化 recorder, ctx, router
			w := httptest.NewRecorder()

			// 构建 req
			reqUrl := fmt.Sprintf("%s/%s", delivery.UserApi, tc.request.Username)
			t.Logf("reqUrl: %s", reqUrl)
			req, err := http.NewRequest(http.MethodGet, reqUrl, nil)
			require.NoError(t, err)

			// 设置认证头信息
			tc.setupAuth(t, req, testTokenMaker)

			// 服务 req
			router.ServeHTTP(w, req)

			// 校验状态码
			require.Equal(t, tc.expectedHTTPCode, w.Code)

			// 将响应 json 反序列化为 actualResult
			var actualResult response.Result[service.UserDTO]
			err = json.Unmarshal(w.Body.Bytes(), &actualResult)
			require.NoError(t, err)

			// 比较 actualResult
			tc.checkResponse(t, tc, actualResult)
		})
	}
}

/** ====================================================================================
 * 🏁 TestGetUser_Success
 * =====================================================================================
 */

// TestGetUser_Success 测试获取用户信息
// Admin 可以获取所有用户信息, User/Vip 可以获取自身信息
func TestGetUser_Success(t *testing.T) {
	username, _, userDTO := setupGetUserData()
	vipDto := userDTO
	vipDto.Username = "vip"
	vipDto.Role = role.RoleVip

	// authWithUserFunc 携带 userToken 进行认证
	authWithUserFunc := func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// userToken
		userToken, _, err := testTokenMaker.CreateToken(username, role.RoleUser, time.Minute, token.TokenTypeAccessToken)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer "+userToken)
	}

	testCases := []getUserTC{
		{
			name:      "Admin 获取 Vip",
			request:   userController.GetUserRequest{Username: "vip"},
			setupAuth: authWithAdminFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), service.GetUserInput{Username: "vip"}).
					Return(vipDto, nil).
					Times(1)
			},
			checkResponse:    checkEqualFuncGetUser,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
		},
		{
			name:      "Admin 获取 User",
			request:   userController.GetUserRequest{Username: username},
			setupAuth: authWithAdminFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), service.GetUserInput{Username: username}).
					Return(userDTO, nil).
					Times(1)
			},
			checkResponse:    checkEqualFuncGetUser,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
		},
		{
			name:      "Vip 获取自身",
			request:   userController.GetUserRequest{Username: "vip"},
			setupAuth: authWithVipFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), service.GetUserInput{Username: "vip"}).
					Return(vipDto, nil).
					Times(1)
			},
			checkResponse:    checkEqualFuncGetUser,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
		},
		{
			name:      "User 获取自身",
			request:   userController.GetUserRequest{Username: username},
			setupAuth: authWithUserFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), service.GetUserInput{Username: username}).
					Return(userDTO, nil).
					Times(1)
			},
			checkResponse:    checkEqualFuncGetUser,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
		},
	}

	runGetUserTC(t, testCases)
}

/** ====================================================================================
 * 🏁 TestGetUser_Authorization
 * =====================================================================================
 */

// TestGetUser_Authorization 测试因认证原因, 获取用户信息失败
// 带Vip/User请求头, 不能获取其他Vip/User的信息
// 不带请求头, 不能获取任何用户信息
func TestGetUser_Authorization(t *testing.T) {
	// 准备 username, DTO
	username, _, _ := setupGetUserData()

	// authWithUserFunc 携带 userToken 进行认证
	authWithUserFunc := func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// userToken
		userToken, _, err := testTokenMaker.CreateToken(username, role.RoleUser, time.Minute, token.TokenTypeAccessToken)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer "+userToken)
	}

	testCases := []getUserTC{
		{
			name:      "Vip 获取其他 Vip",
			request:   userController.GetUserRequest{Username: "vip_another"}, // 获取其他 vip
			setupAuth: authWithVipFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), service.GetUserInput{Username: "vip_another"}).
					Return(service.UserDTO{}, appError.ErrAuthPermissionDenied).
					Times(1)
			},
			checkResponse:    checkEmptyResFuncGetUser,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrAuthPermissionDenied.Code,
		},
		{
			name:      "Vip 获取其他 User",
			request:   userController.GetUserRequest{Username: username}, // 获取其他 User
			setupAuth: authWithVipFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), service.GetUserInput{Username: username}).
					Return(service.UserDTO{}, appError.ErrAuthPermissionDenied).
					Times(1)
			},
			checkResponse:    checkEmptyResFuncGetUser,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrAuthPermissionDenied.Code,
		},
		{
			name:      "User 获取其他 Vip",
			request:   userController.GetUserRequest{Username: "vip_another"}, // 获取其他 vip
			setupAuth: authWithUserFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), service.GetUserInput{Username: "vip_another"}).
					Return(service.UserDTO{}, appError.ErrAuthPermissionDenied).
					Times(1)
			},
			checkResponse:    checkEmptyResFuncGetUser,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrAuthPermissionDenied.Code,
		},
		{
			name:      "User 获取其他 User",
			request:   userController.GetUserRequest{Username: "user_another"}, // 获取其他 User
			setupAuth: authWithUserFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), service.GetUserInput{Username: "user_another"}).
					Return(service.UserDTO{}, appError.ErrAuthPermissionDenied).
					Times(1)
			},
			checkResponse:    checkEmptyResFuncGetUser,
			expectedHTTPCode: 200,
			expectedBizCode:  appError.ErrAuthPermissionDenied.Code,
		},
		{
			name:      "带格式错误请求头获取 User",
			request:   userController.GetUserRequest{Username: username},
			setupAuth: authWithWrongFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse:    checkEmptyResFuncGetUser,
			expectedHTTPCode: 401,
			expectedBizCode:  appError.ErrAuthBadHeader.Code,
		},
		{
			name:      "不带请求头获取 User",
			request:   userController.GetUserRequest{Username: username},
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse:    checkEmptyResFuncGetUser,
			expectedHTTPCode: 401,
			expectedBizCode:  appError.ErrAuthNoHeader.Code,
		},
	}

	runGetUserTC(t, testCases)
}

/** ====================================================================================
 * 🏁 TestGetUser_Validation
 * =====================================================================================
 */

// TestGetUser_Validation 测试参数错误
// 路径为: /api/v1/user/:username, 其中 username 必须为 string 且>3
// username 不得为空, 否则会触发 404 路径错误
func TestGetUser_Validation(t *testing.T) {
	// 准备 username, DTO
	username, _, _ := setupGetUserData()

	// authWithUserFunc 携带 userToken 进行认证
	authWithUserFunc := func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// userToken
		userToken, _, err := testTokenMaker.CreateToken(username, role.RoleUser, time.Minute, token.TokenTypeAccessToken)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer "+userToken)
	}

	testCases := []getUserTC{
		{
			name:      "携带UserToken, 访问空路径",
			request:   userController.GetUserRequest{Username: ""},
			setupAuth: authWithUserFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse:    checkEmptyResFuncGetUser,
			expectedHTTPCode: 404,
			expectedBizCode:  appError.ErrPathNotFound.Code,
		},
		{
			name:      "携带UserToken, 用错误格式请求 User",
			request:   userController.GetUserRequest{Username: "!"},
			setupAuth: authWithUserFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse:    checkEmptyResFuncGetUser,
			expectedHTTPCode: 400,
			expectedBizCode:  40000,
		},
	}

	runGetUserTC(t, testCases)
}
