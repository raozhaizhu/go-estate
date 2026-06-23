package user_test

import (
	"net/http"
	"os"
	"testing"
	"time"

	controller "github.com/raozhaizhu/go-estate/internal/controller/user"
	ctrl "github.com/raozhaizhu/go-estate/internal/controller/user"
	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	service "github.com/raozhaizhu/go-estate/internal/service/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	"github.com/raozhaizhu/go-estate/pkg/token"
	"github.com/raozhaizhu/go-estate/pkg/validator"
	"github.com/stretchr/testify/require"
)

// 用于测试
var (
	testConfig     util.Config
	testTokenMaker token.Maker
)

func TestMain(m *testing.M) {
	// 初始化配置
	testConfig = util.InitConfig("../../..")

	// 初始化 JWTMaker
	var err error
	testTokenMaker, err = token.NewJwtMaker(testConfig.TokenSymmetricKey)
	if err != nil {
		panic("初始化令牌铸造器失败: " + err.Error())
	}

	// 初始化验证翻译器
	validator.InitTrans()

	exitCode := m.Run()

	os.Exit(exitCode)
}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

func setupUserData() (db.User, string) {
	username := util.RandomUsername()
	password := util.RandomPassword()
	email := util.RandomEmail()

	user := db.User{
		ID:       1,
		Username: username,
		Role:     int16(userDomain.RoleUser),
		Email:    email,
	}

	return user, password
}

func setupUpdateUserData(username string, email string, password string, role userDomain.Role) (controller.UpdateUserRequest, service.UpdateUserInput, service.UserDTO) {
	request := controller.UpdateUserRequest{Username: username, Password: &password, Email: &email}
	input := service.UpdateUserInput{Username: username, Password: &password, Email: &email}
	dto := service.UserDTO{ID: 1, Username: username, Email: email, Role: role}

	return request, input, dto
}

func setupGetUserData() (string, db.User, service.UserDTO) {
	// 准备 input
	username := util.RandomUsername()

	// 准备 预埋数据
	user := db.User{
		ID:       1,
		Username: username,
		Role:     int16(userDomain.RoleUser),
	}

	userDTO := service.UserDTO{
		ID:       1,
		Username: username,
		Role:     userDomain.RoleUser,
	}
	return username, user, userDTO
}

func setupCreateUserData() (controller.CreateUserRequest, service.CreateUserInput, db.User, service.UserDTO) {
	// 准备 input
	username := util.RandomUsername()
	password := util.RandomPassword()
	email := util.RandomEmail()

	request := controller.CreateUserRequest{Username: username, Password: password, Email: email}
	input := service.CreateUserInput{Username: username, Password: password, Email: email}

	// 准备 预埋数据
	user := db.User{
		ID:       1,
		Username: username,
		Role:     int16(userDomain.RoleUser),
	}

	userDTO := service.UserDTO{
		ID:       1,
		Username: username,
		Role:     userDomain.RoleUser,
	}
	return request, input, user, userDTO
}

func authWithFunc(t *testing.T, request *http.Request, tokenMaker token.Maker, username string, role userDomain.Role) {
	token, _, err := testTokenMaker.CreateToken(username, role, time.Minute, token.TokenTypeAccessToken)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+token)
}

// authWithVipFunc 携带 vipToken 进行认证
func authWithVipFunc(t *testing.T, request *http.Request, tokenMaker token.Maker) {
	authWithFunc(t, request, tokenMaker, "vip", role.RoleVip)
}

// authWithWrongFunc 携带 错误格式请求头 进行认证
func authWithWrongFunc(t *testing.T, request *http.Request, tokenMaker token.Maker) {
	request.Header.Set("Authorization", "Bearer ")
}

// authWithEmptyFunc 携带 空请求头 进行认证
func authWithEmptyFunc(t *testing.T, request *http.Request, tokenMaker token.Maker) {
}

// authWithAdminFunc 携带 adminToken 进行认证
func authWithAdminFunc(t *testing.T, request *http.Request, tokenMaker token.Maker) {
	authWithFunc(t, request, tokenMaker, "admin", role.RoleAdmin)
}

// checkEmptyResFuncGetUser response因失败返空
func checkEmptyResFuncGetUser(t *testing.T, tc getUserTC, result response.Result[service.UserDTO]) {
	require.Equal(t, tc.expectedBizCode, result.Code)
	require.Empty(t, result.Data)
}

// checkEqualFuncGetUser response返回的 username 和期望的 username 一致
func checkEqualFuncGetUser(t *testing.T, tc getUserTC, result response.Result[service.UserDTO]) {
	require.Equal(t, tc.expectedBizCode, result.Code)
	require.Equal(t, tc.request.Username, result.Data.Username)
}

// checkEmptyResFuncCreateUser response因失败返空
func checkEmptyResFuncCreateUser(t *testing.T, tc createUserTC, result response.Result[service.UserDTO]) {
	require.Equal(t, tc.expectedBizCode, result.Code)
	require.Empty(t, result.Data)
	t.Log(result.Msg)
	require.Contains(t, result.Msg, tc.expectedMsg)
}

// checkEqualFuncCreateUser response返回的 username 和期望的 username 一致
func checkEqualFuncCreateUser(t *testing.T, tc createUserTC, result response.Result[service.UserDTO]) {
	require.Equal(t, tc.expectedBizCode, result.Code)
	require.Equal(t, tc.request.Username, result.Data.Username)

}

// checkEmptyResFuncUpdateUser response因失败返空
func checkEmptyResFuncUpdateUser(t *testing.T, req ctrl.UpdateUserRequest, data service.UserDTO) {
	require.Empty(t, data)
}

// checkEqualFuncUpdateUser response返回的 username 和期望的 username 一致
func checkEqualFuncUpdateUser(t *testing.T, req ctrl.UpdateUserRequest, data service.UserDTO) {
	require.Equal(t, req.Username, data.Username)
}
