package user_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	ctrl "github.com/raozhaizhu/go-estate/internal/controller/user"
	mock_controller "github.com/raozhaizhu/go-estate/internal/controller/user/mock"
	"github.com/raozhaizhu/go-estate/internal/delivery"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	svc "github.com/raozhaizhu/go-estate/internal/service/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	resp "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 TestUpdateUser_Success
 * =====================================================================================
 */

// TestUpdateUser_Success 测试成功更新新用户
// Vip/User 带自身请求头可更新自身, Admin 可以更新 User/Vip
func TestUpdateUser_Success(t *testing.T) {
	user, password := setupUserData()
	updateUserReq, updateUserInput, userDTO := setupUpdateUserData(user.Username, user.Email, password, role.RoleUser)
	updateVipReq, updateVipInput, vipDTO := setupUpdateUserData("vip", user.Email, password, role.RoleVip)

	// authWithUserFunc 携带 userToken 进行认证
	authWithUserFunc := func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// userToken
		userToken, _, err := testTokenMaker.CreateToken(updateUserReq.Username, role.RoleUser, time.Minute, token.TokenTypeAccessToken)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer "+userToken)
	}

	testCases := []GTC[ctrl.UpdateUserRequest, *svc.DTO]{
		{
			Name:      "带 Admin 请求头更新 Vip",
			Request:   updateVipReq,
			SetupAuth: authWithAdminFunc,
			BuildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					UpdateUser(gomock.Any(), updateVipInput).
					Return(vipDTO, nil).
					Times(1)
			},
			CheckResponse:    checkEqualFuncUpdateUser,
			ExpectedHTTPCode: 200,
			ExpectedBizCode:  200,
		},
		{
			Name:      "带 Admin 请求头更新 User",
			Request:   updateUserReq,
			SetupAuth: authWithAdminFunc,
			BuildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					UpdateUser(gomock.Any(), updateUserInput).
					Return(userDTO, nil).
					Times(1)
			},
			CheckResponse:    checkEqualFuncUpdateUser,
			ExpectedHTTPCode: 200,
			ExpectedBizCode:  200,
		},
		{
			Name:      "Vip 带请求头更新自己",
			Request:   updateVipReq,
			SetupAuth: authWithVipFunc,
			BuildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					UpdateUser(gomock.Any(), updateVipInput).
					Return(vipDTO, nil).
					Times(1)
			},
			CheckResponse:    checkEqualFuncUpdateUser,
			ExpectedHTTPCode: 200,
			ExpectedBizCode:  200,
		},
		{
			Name:      "User 带请求头更新自己",
			Request:   updateUserReq,
			SetupAuth: authWithUserFunc,
			BuildStubs: func(svc *mock_controller.MockService) {
				svc.EXPECT().
					UpdateUser(gomock.Any(), updateUserInput).
					Return(userDTO, nil).
					Times(1)
			},
			CheckResponse:    checkEqualFuncUpdateUser,
			ExpectedHTTPCode: 200,
			ExpectedBizCode:  200,
		},
	}

	RunGenericTC(t, testCases, updateUserReqBuilder)
}

/** ====================================================================================
 * 🏁 TestUpdateUser_Authorization
 * =====================================================================================
 */

// TestUpdateUser_Authorization 测试因认证原因, 更新用户失败
// 不带请求头, 或者带 Vip/User 请求头更新其他用户, 导致失败
func TestUpdateUser_Authorization(t *testing.T) {
	user, password := setupUserData()
	updateUserReq, updateUserInput, _ := setupUpdateUserData(user.Username, user.Email, password, role.RoleUser)
	updateAnotherUserReq, updateAnotherUserInput, _ := setupUpdateUserData("another_user", user.Email, password, role.RoleVip)
	// updateVipReq, updateVipInput, vipDTO := setupUpdateUserData("vip", user.Email, password, role.RoleVip)
	updateAnotherVipReq, updateAnotherVipInput, _ := setupUpdateUserData("another_vip", user.Email, password, role.RoleVip)

	// authWithUserFunc 携带 userToken 进行认证
	authWithUserFunc := func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		// userToken
		userToken, _, err := testTokenMaker.CreateToken(updateUserReq.Username, role.RoleUser, time.Minute, token.TokenTypeAccessToken)
		require.NoError(t, err)
		request.Header.Set("Authorization", "Bearer "+userToken)
	}

	testCases := []GTC[ctrl.UpdateUserRequest, *svc.DTO]{
		{
			Name:      "Vip 带请求头更新 其他 Vip",
			Request:   updateAnotherVipReq,
			SetupAuth: authWithVipFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), updateAnotherVipInput).
					Return(nil, appError.ErrAuthPermissionDenied).
					Times(1)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 200,
			ExpectedBizCode:  appError.CodeAuthPermissionDenied,
		},
		{
			Name:      "Vip 带请求头更新 User",
			Request:   updateUserReq,
			SetupAuth: authWithVipFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), updateUserInput).
					Return(nil, appError.ErrAuthPermissionDenied).
					Times(1)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 200,
			ExpectedBizCode:  appError.CodeAuthPermissionDenied,
		},
		{
			Name:      "User 带请求头更新 Vip",
			Request:   updateAnotherVipReq,
			SetupAuth: authWithUserFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), updateAnotherVipInput).
					Return(nil, appError.ErrAuthPermissionDenied).
					Times(1)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 200,
			ExpectedBizCode:  appError.CodeAuthPermissionDenied,
		},
		{
			Name:      "User 带请求头更新其他 User",
			Request:   updateAnotherUserReq,
			SetupAuth: authWithUserFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), updateAnotherUserInput).
					Return(nil, appError.ErrAuthPermissionDenied).
					Times(1)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 200,
			ExpectedBizCode:  appError.CodeAuthPermissionDenied,
		},
		{
			Name:      "不带请求头更新 User",
			Request:   updateUserReq,
			SetupAuth: authWithEmptyFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), updateUserInput).
					Return(nil, appError.ErrAuthPermissionDenied).
					Times(0)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 401,
			ExpectedBizCode:  appError.CodeAuthNoHeader,
		},
		{
			Name:      "携带错误格式请求头更新 User",
			Request:   updateUserReq,
			SetupAuth: authWithWrongFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), updateUserInput).
					Return(nil, appError.ErrAuthPermissionDenied).
					Times(0)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 401,
			ExpectedBizCode:  appError.CodeAuthBadHeader,
		},
	}

	RunGenericTC(t, testCases, updateUserReqBuilder)
}

/** ====================================================================================
 * 🏁 TestUpdateUser_Validation
 * =====================================================================================
 */

// TestUpdateUser_Validation 测试因参数原因, 更新用户失败
// 用户名正确, 但邮箱/密码格式错误, 导致更新失败
// 请求格式如下
// type UpdateUserRequest struct {
// 	Username string  `uri:"username" binding:"required,min=3,max=32"`
// 	Password *string `json:"password" binding:"omitempty,min=8,max=16"`
// 	Email    *string `json:"email" binding:"omitempty,email"`
// }

func TestUpdateUser_Validation(t *testing.T) {
	// username
	tooShortUsernameReq, _, _ := setupUpdateUserData(util.RandomString(2), "", "", role.RoleUser)
	tooLongUsernameReq, _, _ := setupUpdateUserData(util.RandomString(33), "", "", role.RoleUser)
	// pwd
	tooShortPwdReq, _, _ := setupUpdateUserData(util.RandomUsername(), util.RandomEmail(), util.RandomString(7), role.RoleUser)
	tooLongPwdReq, _, _ := setupUpdateUserData(util.RandomUsername(), util.RandomEmail(), util.RandomString(17), role.RoleUser)
	// email
	malformedEmailReq, _, _ := setupUpdateUserData(util.RandomUsername(), "a123.com", util.RandomPassword(), role.RoleUser)

	// 根本到不了 svc 校验用户名是否一致这一步, token 随便造即可
	authWithUserFunc := func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
		authWithFunc(t, request, tokenMaker, util.RandomUsername(), role.RoleUser)
	}

	testCases := []GTC[ctrl.UpdateUserRequest, *svc.DTO]{
		{
			Name:      "用户名过短",
			Request:   tooShortUsernameReq,
			SetupAuth: authWithUserFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 400,
			ExpectedBizCode:  appError.CodeInvalidParam,
			ExpectedMsg:      "Username长度必须至少为3个字符",
		},
		{
			Name:      "用户名过长",
			Request:   tooLongUsernameReq,
			SetupAuth: authWithUserFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 400,
			ExpectedBizCode:  appError.CodeInvalidParam,
			ExpectedMsg:      "Username长度不能超过32个字符",
		},
		{
			Name:      "密码过短",
			Request:   tooShortPwdReq,
			SetupAuth: authWithUserFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 400,
			ExpectedBizCode:  appError.CodeInvalidParam,
			ExpectedMsg:      "Password长度必须至少为8个字符",
		},
		{
			Name:      "密码过长",
			Request:   tooLongPwdReq,
			SetupAuth: authWithUserFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 400,
			ExpectedBizCode:  appError.CodeInvalidParam,
			ExpectedMsg:      "Password长度不能超过16个字符",
		},
		{
			Name:      "邮箱格式错误",
			Request:   malformedEmailReq,
			SetupAuth: authWithUserFunc,
			BuildStubs: func(mockSvc *mock_controller.MockService) {
				mockSvc.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			CheckResponse:    checkEmptyResFuncUpdateUser,
			ExpectedHTTPCode: 400,
			ExpectedBizCode:  appError.CodeInvalidParam,
			ExpectedMsg:      "Email必须是一个有效的邮箱",
		},
	}

	RunGenericTC(t, testCases, updateUserReqBuilder)
}

/** ====================================================================================
 * 🏁 Generic
 * =====================================================================================
 */

// GTC 泛型测试用例
type GTC[Req any, Data any] struct {
	Name             string
	Request          Req
	SetupAuth        func(t *testing.T, request *http.Request, tokenMaker token.Maker)
	BuildStubs       func(svc *mock_controller.MockService)
	CheckResponse    func(t *testing.T, req Req, data Data)
	ExpectedHTTPCode int
	ExpectedBizCode  int
	ExpectedMsg      string
}

// RunGenericTC 泛型运行器
func RunGenericTC[Req any, Data any](
	t *testing.T,
	testCases []GTC[Req, Data],
	reqBuilder func(t *testing.T, req Req, reqBody io.Reader) *http.Request,
) {
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mock_controller.NewMockService(ctrl)
			router := delivery.SetupRouter(delivery.Services{
				UserSvc: mockService,
			}, testConfig, testTokenMaker)

			tc.BuildStubs(mockService)
			w := httptest.NewRecorder()

			// 序列化 req
			data, err := json.Marshal(tc.Request)
			require.NoError(t, err)

			// 构建 body
			reqBody := bytes.NewBuffer(data)
			req := reqBuilder(t, tc.Request, reqBody)
			t.Logf("req is :%s", req.URL)

			tc.SetupAuth(t, req, testTokenMaker)
			router.ServeHTTP(w, req)

			// 反序列化得到 response
			var response resp.Result[Data]
			json.Unmarshal(w.Body.Bytes(), &response)

			// 校验 http 码, 错误码, 错误信息是否正确
			require.Equal(t, tc.ExpectedHTTPCode, w.Code)
			require.Equal(t, tc.ExpectedBizCode, response.Code)
			require.Contains(t, response.Msg, tc.ExpectedMsg)

			if tc.CheckResponse != nil {
				tc.CheckResponse(t, tc.Request, response.Data)
			}
		})
	}
}

// updateUserReqBuilder 铸造request
func updateUserReqBuilder(t *testing.T, req ctrl.UpdateUserRequest, reqBody io.Reader) *http.Request {
	request, err := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/%s", delivery.UserApi, req.Username), reqBody)
	require.NoError(t, err)
	return request
}
