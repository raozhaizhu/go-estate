package auth_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	authCtrl "github.com/raozhaizhu/go-estate/internal/controller/auth"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	"github.com/raozhaizhu/go-estate/internal/util"
	mock_worker "github.com/raozhaizhu/go-estate/internal/worker/mock"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 TestLogin
 * =====================================================================================
 */

func TestLogin(t *testing.T) {
	// 预备数据
	username := util.RandomUsername()
	password := util.RandomPassword()
	hashedPassword, _ := util.HashPassword(password)

	wrongUsername, wrongPassword := "wrong", "wrong"
	correctLoginInput := auth.LoginInput{
		Username:  username,
		Password:  password,
		UserAgent: testUtil.UserAgent,
		ClientIp:  testUtil.ClientIp,
		DeviceID:  testUtil.DeviceID,
	}
	wrongUsernameInput, wrongPasswordInput := correctLoginInput, correctLoginInput
	wrongUsernameInput.Username = wrongUsername
	wrongPasswordInput.Password = wrongPassword

	logoutInput := auth.LogoutInput{
		Username: correctLoginInput.Username,
		DeviceID: correctLoginInput.DeviceID,
	}

	activeTidsMock := []string{"token"}
	accessPayload := &token.Payload{
		Username:  username,
		Role:      userDomain.RoleUser,
		TokenType: token.TokenTypeAccessToken,
	}
	refreshPayload := &token.Payload{
		Username:  username,
		Role:      userDomain.RoleUser,
		TokenType: token.TokenTypeRefreshToken,
	}

	// 默认行为和校验逻辑逻辑
	defaultAction := func(svc authCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		logInput, ok := input.(auth.LoginInput)
		require.True(t, ok)
		dto, refreshToken, err := svc.Login(ctx, logInput)
		return []interface{}{dto, refreshToken}, err
	}
	failCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.Error(t, actualErr)
		require.ErrorIs(t, actualErr, expectedErr)
		require.NotNil(t, results)
		require.Len(t, results, 2)
	}
	successCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.NoError(t, actualErr)
		require.NotNil(t, results)
		require.Len(t, results, 2)

		dto, ok := results[0].(*auth.DTO)
		require.True(t, ok)
		require.NotNil(t, dto)
		require.NotEmpty(t, dto)
		require.Equal(t, username, dto.UserInfo.Username)
	}

	// 成功桩函数
	stubGetUserSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().
			GetUser(gomock.Any(), username).
			Return(db.User{
				Username:       username,
				HashedPassword: hashedPassword,
			}, nil).Times(1)
	}
	stubExecTxSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().ExecTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(db.Querier) error) error {
				return fn(storeMock)
			}).Times(1)
	}
	stubGetActiveSessionIDsByUserDeviceForUpdateSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().GetActiveSessionIDsByUserDeviceForUpdate(gomock.Any(), logoutInput.ToDBParams()).
			Return(activeTidsMock, nil).Times(1)
	}
	stubBlockSessionsByIDsSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().BlockSessionsByIDs(gomock.Any(), gomock.Any()).
			Return(nil).Times(1)
	}
	stubDistributeTaskDeleteSessionsSuccess := func(distributor *mock_worker.MockTaskDistributor) {
		distributor.EXPECT().DistributeTaskDeleteSessions(gomock.Any(), activeTidsMock).
			Return(nil).Times(1)
	}
	stubForgeTokenPairSuccess := func(tokenMakerMock *mock_token.MockMaker) {
		tokenMakerMock.EXPECT().ForgeTokenPair(gomock.Any(), gomock.Any()).
			DoAndReturn(func(user *db.User, config *util.Config) (string, string, *token.Payload, *token.Payload, error) {
				require.Equal(t, user.Username, username)
				require.Equal(t, user.HashedPassword, hashedPassword)
				return testUtil.AccessStr, testUtil.RefreshStr, accessPayload, refreshPayload, nil
			}).Times(1)
	}
	stubCreateSessionSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().CreateSession(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, params db.CreateSessionParams) error {
				require.Equal(t, username, params.Username)
				require.NotEmpty(t, params.ID)
				return nil
			}).Times(1)
	}
	stubAddNewSessionSuccess := func(cacheMock *mock_db.MockCache) {
		cacheMock.EXPECT().AddNewSession(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, params cache.AddNewSessionParams) error {
			require.NotEmpty(t, params.JTI)
			require.Equal(t, username, params.Username)
			return nil
		}).Times(1)
	}

	// 测试用例
	testCases := []testCase{
		{
			name:  "用户不存在",
			input: wrongUsernameInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 先获取用户,但用户不存在
				storeMock.EXPECT().
					GetUser(gomock.Any(), wrongUsername).
					Return(db.User{}, appError.ErrWrongUsernamePassword).
					Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrWrongUsernamePassword,
		},
		{
			name:  "用户存在但密码错误",
			input: wrongPasswordInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 先获取用户,用户存在,但密码错误
				stubGetUserSuccess(storeMock)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrWrongUsernamePassword,
		},
		{
			name:  "铸造 Token 时失败",
			input: correctLoginInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 先获取用户,用户存在
				stubGetUserSuccess(storeMock)
				// 但在铸造 token 时失败
				tokenMakerMock.EXPECT().ForgeTokenPair(gomock.Any(), gomock.Any()).
					DoAndReturn(func(user *db.User, config *util.Config) (string, string, *token.Payload, *token.Payload, error) {
						require.Equal(t, user.Username, username)
						require.Equal(t, user.HashedPassword, hashedPassword)
						return "", "", nil, nil, appError.ErrServerErr
					}).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "获取用户活跃 Sessions 时失败",
			input: correctLoginInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 先获取用户,用户存在
				stubGetUserSuccess(storeMock)
				// 铸造 token 成功
				stubForgeTokenPairSuccess(tokenMakerMock)
				// 开启事务, 但获取是否有活跃 Sessions 时失败
				stubExecTxSuccess(storeMock)
				storeMock.EXPECT().GetActiveSessionIDsByUserDeviceForUpdate(gomock.Any(), logoutInput.ToDBParams()).
					Return([]string{}, appError.ErrServerErr).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "清除用户活跃 Sessions 时失败",
			input: correctLoginInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 先获取用户,用户存在
				stubGetUserSuccess(storeMock)
				// 铸造 token 成功
				stubForgeTokenPairSuccess(tokenMakerMock)
				// 开启事务
				stubExecTxSuccess(storeMock)
				// 成功获取用户活跃 Sessions
				stubGetActiveSessionIDsByUserDeviceForUpdateSuccess(storeMock)
				// 但在清除 sessions 时失败
				storeMock.EXPECT().BlockSessionsByIDs(gomock.Any(), gomock.Any()).
					Return(appError.ErrServerErr).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "制造 Session 时失败",
			input: correctLoginInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 先获取用户,用户存在
				stubGetUserSuccess(storeMock)
				// 铸造 token 成功
				stubForgeTokenPairSuccess(tokenMakerMock)
				// 开启事务
				stubExecTxSuccess(storeMock)
				// 成功获取用户活跃 Sessions
				stubGetActiveSessionIDsByUserDeviceForUpdateSuccess(storeMock)
				// 成功清除 sessions
				stubBlockSessionsByIDsSuccess(storeMock)
				// 但在 createSession 失败
				storeMock.EXPECT().CreateSession(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, params db.CreateSessionParams) error {
						require.Equal(t, username, params.Username)
						require.NotEmpty(t, params.ID)
						return appError.ErrServerErr
					}).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "调用 worker 异步清理 redis失败",
			input: correctLoginInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 先获取用户,用户存在
				stubGetUserSuccess(storeMock)
				// 铸造 token 成功
				stubForgeTokenPairSuccess(tokenMakerMock)
				// 开启事务
				stubExecTxSuccess(storeMock)
				// 成功获取用户活跃 Sessions
				stubGetActiveSessionIDsByUserDeviceForUpdateSuccess(storeMock)
				// 成功清除 sessions
				stubBlockSessionsByIDsSuccess(storeMock)
				// 成功创建 Session
				stubCreateSessionSuccess(storeMock)
				// 但在派发 worker 清理 redis 时失败
				distributor.EXPECT().DistributeTaskDeleteSessions(gomock.Any(), activeTidsMock).
					Return(appError.ErrServerErr).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "尽力而为存入 Redis 时失败, 但不影响整体逻辑成功",
			input: correctLoginInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 先获取用户,用户存在
				stubGetUserSuccess(storeMock)
				// 铸造 token 成功
				stubForgeTokenPairSuccess(tokenMakerMock)
				// 开启事务
				stubExecTxSuccess(storeMock)
				// 成功获取用户活跃 Sessions
				stubGetActiveSessionIDsByUserDeviceForUpdateSuccess(storeMock)
				// 成功清除 sessions
				stubBlockSessionsByIDsSuccess(storeMock)
				// 成功创建 Session
				stubCreateSessionSuccess(storeMock)
				// 派发 worker 清理 redis 成功
				stubDistributeTaskDeleteSessionsSuccess(distributor)
				// 但在存 Session 到 Redis 时失败
				cacheMock.EXPECT().AddNewSession(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, params cache.AddNewSessionParams) error {
					require.NotEmpty(t, params.JTI)
					require.Equal(t, username, params.Username)
					return appError.ErrServerErr
				}).Times(1)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
			expectedErr:   nil,
		},
		{
			name:  "用户已有活跃 Sessions(之前登录过), 重新登录成功",
			input: correctLoginInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 先获取用户,用户存在
				stubGetUserSuccess(storeMock)
				// 铸造 token 成功
				stubForgeTokenPairSuccess(tokenMakerMock)
				// 开启事务
				stubExecTxSuccess(storeMock)
				// 成功获取用户活跃 Sessions
				stubGetActiveSessionIDsByUserDeviceForUpdateSuccess(storeMock)
				// 成功清除 sessions
				stubBlockSessionsByIDsSuccess(storeMock)
				// 成功创建 Session
				stubCreateSessionSuccess(storeMock)
				// 派发 worker 清理 redis 成功
				stubDistributeTaskDeleteSessionsSuccess(distributor)
				// 存入 Session 成功
				stubAddNewSessionSuccess(cacheMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
		{
			name:  "用户没有活跃 Sessions(之前没登录), 崭新登录成功",
			input: correctLoginInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 先获取用户,用户存在
				stubGetUserSuccess(storeMock)
				// 铸造 token 成功
				stubForgeTokenPairSuccess(tokenMakerMock)
				// 开启事务
				stubExecTxSuccess(storeMock)
				// 无法获取到用户活跃 Sessions (不存在)
				storeMock.EXPECT().GetActiveSessionIDsByUserDeviceForUpdate(gomock.Any(), logoutInput.ToDBParams()).
					Return([]string{}, nil).Times(1)
				// 跳过清除 sessions 和派发 worker 清理 redis, 因为无需登出
				// 成功创建 Session
				stubCreateSessionSuccess(storeMock)
				// 存入 Session 成功
				stubAddNewSessionSuccess(cacheMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
	}

	runTC(t, testCases)
}
