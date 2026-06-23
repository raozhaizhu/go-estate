package user_test

import (
	"bytes"
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
	userService "github.com/raozhaizhu/go-estate/internal/service/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */
type createUserTC struct {
	name    string
	request userController.CreateUserRequest
	// 设置认证头信息
	setupAuth func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	// svc埋桩
	buildStubs func(svc *mock_controller.MockService)
	// 校验数据
	checkResponse func(t *testing.T, tc createUserTC, result response.Result[userService.UserDTO])
	// 请求地址
	reqUrl string
	// 响应代码
	expectedHTTPCode int
	expectedBizCode  int
	expectedMsg      string
}

func runCreateUserTC(t *testing.T, testCases []createUserTC) {
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
			data, err := json.Marshal(tc.request)
			require.NoError(t, err)
			reqBody := bytes.NewBuffer(data)
			req, err := http.NewRequest(http.MethodPost, tc.reqUrl, reqBody)
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
 * 🏁 Variant
 * =====================================================================================
 */

// url
var createVipReqUrl = fmt.Sprintf("%s/%s", delivery.UserApi, "vip")

/** ====================================================================================
 * 🏁 TestCreateUser_Success
 * =====================================================================================
 */

// TestCreateUser_Success 测试成功创建新用户
// 不带请求头可创建 User, 带 Admin 请求头可以创建 Vip/Uer
func TestCreateUser_Success(t *testing.T) {
	createUserReq, createUserInput, _, userDTO := setupCreateUserData()

	// vip 部分
	// req
	createVipReq := createUserReq
	createVipReq.Username = "vip"
	// input
	createVipInput := createUserInput
	createVipInput.Username = "vip"
	// dto
	vipDTO := userDTO
	vipDTO.Username = "vip"
	vipDTO.Role = role.RoleVip

	testCases := []createUserTC{
		{
			name:      "不带请求头创建 User",
			request:   createUserReq,
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), createUserInput, role.RoleUser).
					Return(userDTO, nil).
					Times(1)
			},
			reqUrl:           delivery.UserApi,
			checkResponse:    checkEqualFuncCreateUser,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
		},
		{
			name:      "带 Admin 请求头创建 Vip",
			request:   createVipReq,
			setupAuth: authWithAdminFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), createVipInput, role.RoleVip).
					Return(vipDTO, nil).
					Times(1)
			},
			reqUrl:           createVipReqUrl,
			checkResponse:    checkEqualFuncCreateUser,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
		},
		{
			name:      "带 Admin 请求头创建 User",
			request:   createUserReq,
			setupAuth: authWithAdminFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), createUserInput, role.RoleUser).
					Return(userDTO, nil).
					Times(1)
			},
			reqUrl:           delivery.UserApi,
			checkResponse:    checkEqualFuncCreateUser,
			expectedHTTPCode: 200,
			expectedBizCode:  200,
		},
	}

	runCreateUserTC(t, testCases)
}

/** ====================================================================================
 * 🏁 TestCreateUser_Authorization
 * =====================================================================================
 */

// TestCreateUser_Authorization 测试因认证原因, 创建用户失败
// 不带请求头, 或者带 Vip/User 头创建 Vip 失败
func TestCreateUser_Authorization(t *testing.T) {
	createUserReq, createUserInput, _, userDTO := setupCreateUserData()

	// vip 部分
	// req
	createVipReq := createUserReq
	createVipReq.Username = "vip"
	// input
	createVipInput := createUserInput
	createVipInput.Username = "vip"
	// dto
	vipDTO := userDTO
	vipDTO.Username = "vip"
	vipDTO.Role = role.RoleVip

	// authWithUserFunc 携带 userToken 进行认证
	authWithUserFunc := func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// userToken
		userToken, _, err := testTokenMaker.CreateToken(createUserReq.Username, role.RoleUser, time.Minute, token.TokenTypeAccessToken)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer "+userToken)
	}

	testCases := []createUserTC{
		{
			name:      "不带请求头创建 Vip",
			request:   createVipReq,
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           createVipReqUrl,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 401,
			expectedBizCode:  appError.CodeAuthNoHeader,
		},
		{
			name:      "带 Vip 请求头创建 Vip",
			request:   createVipReq,
			setupAuth: authWithVipFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           createVipReqUrl,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 401,
			expectedBizCode:  appError.CodeAuthPermissionDenied,
		},
		{
			name:      "带 User 请求头创建 Vip",
			request:   createVipReq,
			setupAuth: authWithUserFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           createVipReqUrl,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 401,
			expectedBizCode:  appError.CodeAuthPermissionDenied,
		},
	}

	runCreateUserTC(t, testCases)
}

/** ====================================================================================
 * 🏁 TestCreateUser_Validation
 * =====================================================================================
 */

// TestCreateUser_Validation
// 路径为: /api/v1/user
// 结构体要求如下
//
//	type CreateUserRequest struct {
//		Username string `json:"username" binding:"required,min=3,max=32"`
//		Password string `json:"password" binding:"required,min=8,max=16"`
//		Email    string `json:"email" binding:"required,email"`
//	}
func TestCreateUser_Validation(t *testing.T) {
	createUserReq, _, _, _ := setupCreateUserData()

	// 构建错误 req
	// username
	noUsernameReq := createUserReq
	noUsernameReq.Username = ""

	tooShortUsernameReq := createUserReq
	tooShortUsernameReq.Username = "a"

	tooLongUsernameReq := createUserReq
	tooLongUsernameReq.Username = util.RandomString(33)

	// password
	noPasswordReq := createUserReq
	noPasswordReq.Password = ""

	tooShortPasswordReq := createUserReq
	tooShortPasswordReq.Password = "a"

	tooLongPasswordReq := createUserReq
	tooLongPasswordReq.Password = util.RandomString(17)

	// email
	noEmailReq := createUserReq
	noEmailReq.Email = ""

	malformedEmailReq := createUserReq
	malformedEmailReq.Email = "123.com"

	testCases := []createUserTC{
		{
			name:      "参数 Username 不存在",
			request:   noUsernameReq,
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           delivery.UserApi,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 400,
			expectedBizCode:  appError.CodeGroupClientError,
			expectedMsg:      "Username为必填字段",
		},
		{
			name:      "参数 Username 过短",
			request:   tooShortUsernameReq,
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           delivery.UserApi,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 400,
			expectedBizCode:  appError.CodeGroupClientError,
			expectedMsg:      "Username长度必须至少为3个字符",
		},
		{
			name:      "参数 Username 过长",
			request:   tooLongUsernameReq,
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           delivery.UserApi,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 400,
			expectedBizCode:  appError.CodeGroupClientError,
			expectedMsg:      "Username长度不能超过32个字符",
		},
		{
			name:      "参数 Password 不存在",
			request:   noPasswordReq,
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           delivery.UserApi,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 400,
			expectedBizCode:  appError.CodeGroupClientError,
			expectedMsg:      "Password为必填字段",
		},
		{
			name:      "参数 Password 过短",
			request:   tooShortPasswordReq,
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           delivery.UserApi,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 400,
			expectedBizCode:  appError.CodeGroupClientError,
			expectedMsg:      "Password长度必须至少为8个字符",
		},
		{
			name:      "参数 Password 过长",
			request:   tooLongPasswordReq,
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           delivery.UserApi,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 400,
			expectedBizCode:  appError.CodeGroupClientError,
			expectedMsg:      "Password长度不能超过16个字符",
		},
		{
			name:      "参数 Email 不存在",
			request:   noEmailReq,
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           delivery.UserApi,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 400,
			expectedBizCode:  appError.CodeGroupClientError,
			expectedMsg:      "Email为必填字段",
		},
		{
			name:      "参数 Email 格式错误",
			request:   malformedEmailReq,
			setupAuth: authWithEmptyFunc,
			buildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			reqUrl:           delivery.UserApi,
			checkResponse:    checkEmptyResFuncCreateUser,
			expectedHTTPCode: 400,
			expectedBizCode:  appError.CodeGroupClientError,
			expectedMsg:      "Email必须是一个有效的邮箱",
		},
	}

	runCreateUserTC(t, testCases)
}
