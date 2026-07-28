package delivery

// 定义全局版本路由
const (
	CurrAPI = "/api/v1"

	AuthAPI        = CurrAPI + "/auth"
	AuthLoginAPI   = AuthAPI + "/login"
	AuthRefreshAPI = AuthAPI + "/refresh"
	AuthLogoutAPI  = AuthAPI + "/logout"

	UserAPI                   = CurrAPI + "/user"
	UserCreateNormalUserAPI   = UserAPI
	UserGetAvatarUploadUrlAPI = UserAPI + "/avatar/upload-url"
	UserCreateVipAPI          = UserAPI + "/vip"
	UserGetUserAPI            = UserAPI
	UserUpdateUserAPI         = UserAPI

	DailyDataAPI          = CurrAPI + "/daily_data"
	DailyDataGetDayAPI    = DailyDataAPI + "/day"
	DailyDataGetPeriodAPI = DailyDataAPI + "/period"
	DailyDataGetAllAPI    = DailyDataAPI + "/all"
)
