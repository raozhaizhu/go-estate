package auth_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	authCtrl "github.com/raozhaizhu/go-estate/internal/controller/auth"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
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
 * 🏁 Refresh
 * =====================================================================================
 */

func TestRefresh(t *testing.T) {
	// 预备数据
	username := util.RandomUsername()
	refreshStr := testUtil.RefreshStr
	testRefreshUUID := testUtil.NewUUID()
	testAccessUUID := testUtil.NewUUID()
	refreshJti := testRefreshUUID.String()

	accessPayload := &token.Payload{
		ID:        testAccessUUID,
		Username:  username,
		Role:      userDomain.RoleUser,
		TokenType: token.TokenTypeAccessToken,
	}
	refreshPayload := &token.Payload{
		ID:        testRefreshUUID,
		Username:  username,
		Role:      userDomain.RoleUser,
		TokenType: token.TokenTypeRefreshToken,
	}
	cacheSession := &cache.Session{
		Username:  username,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	cacheExpiredSession := &cache.Session{
		Username:  username,
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	cacheBlockedSession := &cache.Session{
		Username:  username,
		ExpiresAt: time.Now().Add(time.Hour),
		IsBlocked: true,
	}
	dbSession := db.Session{
		ID:        refreshJti,
		Username:  username,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	dbExpiredSession := db.Session{
		ID:        refreshJti,
		Username:  username,
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	dbBlockedSession := db.Session{
		ID:        refreshJti,
		Username:  username,
		ExpiresAt: time.Now().Add(time.Hour),
		IsBlocked: true,
	}

	// 默认行为和校验逻辑逻辑
	defaultAction := func(svc authCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		refreshStr, ok := input.(string)
		require.True(t, ok)
		dto, err := svc.Refresh(ctx, refreshStr)
		return []interface{}{dto}, err
	}
	failCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.Error(t, actualErr)
		require.ErrorIs(t, actualErr, expectedErr)
		require.NotNil(t, results)
		require.Len(t, results, 1)
	}
	successCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.NoError(t, actualErr)
		require.NotNil(t, results)
		require.Len(t, results, 1)

		dto, ok := results[0].(*auth.DTO)
		require.True(t, ok)
		require.NotNil(t, dto)
		require.NotEmpty(t, dto)
		require.Equal(t, username, dto.UserInfo.Username)
	}

	// 成功桩函数
	stubVerifyTokenSuccess := func(tokenMakerMock *mock_token.MockMaker) {
		tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).
			DoAndReturn(func(tokenStr string, tokenType token.TokenType) (*token.Payload, error) {
				require.Equal(t, tokenStr, refreshStr)
				require.Equal(t, tokenType, token.TokenType(token.TokenTypeRefreshToken))
				return refreshPayload, nil
			}).Times(1)
	}
	stubGetCacheSessionSuccess := func(cacheMock *mock_db.MockSessionCache) {
		cacheMock.EXPECT().
			GetSession(gomock.Any(), refreshJti).
			Return(cacheSession, nil).Times(1)
	}
	stubGetDBSessionSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().
			GetSession(gomock.Any(), refreshJti).
			Return(dbSession, nil).Times(1)
	}
	stubGetDBExpiredSessionSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().
			GetSession(gomock.Any(), refreshJti).
			Return(dbExpiredSession, nil).Times(1)
	}
	stubGetDBBlockedSessionSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().
			GetSession(gomock.Any(), refreshJti).
			Return(dbBlockedSession, nil).Times(1)
	}
	stubAddNewSessionSuccess := func(cacheMock *mock_db.MockSessionCache) {
		cacheMock.EXPECT().AddNewSession(gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, params cache.AddNewSessionParams) error {
			require.NotEmpty(t, params.JTI)
			require.Equal(t, username, params.Username)
			return nil
		}).Times(1)
	}
	stubCreateTokenSuccess := func(tokenMakerMock *mock_token.MockMaker) {
		tokenMakerMock.EXPECT().CreateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(_username string, role role.Role, duration time.Duration, tokenType token.TokenType) (string, *token.Payload, error) {
				require.Equal(t, _username, username)
				return testUtil.AccessStr, accessPayload, nil
			}).Times(1)
	}

	// 测试用例
	testCases := []testCase{
		{
			name:  "校验刷新令牌失败",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌失败
				tokenMakerMock.EXPECT().VerifyToken(gomock.Any(), gomock.Any()).
					DoAndReturn(func(tokenStr string, tokenType token.TokenType) (*token.Payload, error) {
						require.Equal(t, tokenStr, refreshStr)
						require.Equal(t, tokenType, token.TokenType(token.TokenTypeRefreshToken))
						return nil, appError.ErrServerErr
					}).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "Session 在缓存和数据库都不存在",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌成功
				stubVerifyTokenSuccess(tokenMakerMock)
				// 但缓存 Miss, 获取 Session 失败
				cacheMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(nil, appError.ErrMissSession).Times(1)
				// 尝试去数据库获取 Session, 同样失败(Session 不存在)
				storeMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(db.Session{}, sql.ErrNoRows).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrNoSession,
		},
		{
			name:  "缓存 Miss, 去数据库查, 但查询失败, 发生内部错误(譬如断联)",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌成功
				stubVerifyTokenSuccess(tokenMakerMock)
				// 但缓存 Miss, 获取 Session 失败
				cacheMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(nil, appError.ErrMissSession).Times(1)
				// 尝试去数据库获取 Session, 返回内部错误
				storeMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(db.Session{}, driver.ErrBadConn).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "缓存 Miss, 成功在数据库找到, 但是 Session 过期",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌成功
				stubVerifyTokenSuccess(tokenMakerMock)
				// 但缓存 Miss, 获取 Session 失败
				cacheMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(nil, appError.ErrMissSession).Times(1)
				// 尝试去数据库获取 Session, 成功找到过期 Session
				stubGetDBExpiredSessionSuccess(storeMock)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrExpiredToken,
		},
		{
			name:  "缓存 Miss, 成功在数据库找到, 但是 Session 被封禁",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌成功
				stubVerifyTokenSuccess(tokenMakerMock)
				// 但缓存 Miss, 获取 Session 失败
				cacheMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(nil, appError.ErrMissSession).Times(1)
				// 尝试去数据库获取 Session, 成功找到被封禁 Session
				stubGetDBBlockedSessionSuccess(storeMock)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrBlockedSession,
		},
		{
			name:  "缓存 Miss, 成功在数据库找到可用 Session, 写回缓存成功, 但发放 accessToken 失败",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌成功
				stubVerifyTokenSuccess(tokenMakerMock)
				// 但缓存 Miss
				cacheMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(nil, appError.ErrMissSession).Times(1)
				// 去数据库获取 Session, 成功找到合规 Session
				stubGetDBSessionSuccess(storeMock)
				// 写回缓存成功
				stubAddNewSessionSuccess(cacheMock)
				// 但在铸造 accessToken 时失败
				tokenMakerMock.EXPECT().CreateToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					DoAndReturn(func(_username string, role role.Role, duration time.Duration, tokenType token.TokenType) (string, *token.Payload, error) {
						require.Equal(t, _username, username)
						return "", nil, appError.ErrServerErr
					}).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "缓存 Miss, 成功在数据库找到可用 Session, 但在尽力而为写回缓存时失败, 并不影响整体逻辑成功",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌成功
				stubVerifyTokenSuccess(tokenMakerMock)
				// 但缓存 Miss, 获取 Session 失败
				cacheMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(nil, appError.ErrMissSession).Times(1)
				// 尝试去数据库获取 Session, 成功找到合规 Session, 尝试写回缓存但是失败
				stubGetDBSessionSuccess(storeMock)
				cacheMock.EXPECT().
					AddNewSession(gomock.Any(), gomock.Any()).
					Return(appError.ErrServerErr.WithErr(fmt.Errorf("写回 redis 失败"))). // 模拟缓存写失败
					Times(1)
				// 成功铸造 accessToken
				stubCreateTokenSuccess(tokenMakerMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},

		{
			name:  "缓存命中, 但 session 过期",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌成功
				stubVerifyTokenSuccess(tokenMakerMock)
				// 缓存命中,但得到过期 Session
				cacheMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(cacheExpiredSession, nil).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrExpiredToken,
		},
		{
			name:  "缓存命中, 但 session 封禁",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌成功
				stubVerifyTokenSuccess(tokenMakerMock)
				// 缓存命中,但得到封禁 Session
				cacheMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(cacheBlockedSession, nil).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrBlockedSession,
		},
		{
			name:  "缓存命中, 并且 session 可用, 铸造 token 成功",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌成功
				stubVerifyTokenSuccess(tokenMakerMock)
				// 缓存命中,Session可用
				stubGetCacheSessionSuccess(cacheMock)
				// 铸造 token 成功
				stubCreateTokenSuccess(tokenMakerMock)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
		{
			name:  "查询 redis 失败, 返回内部错误,直接拦截",
			input: refreshStr,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockSessionCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				// 校验刷新令牌成功
				stubVerifyTokenSuccess(tokenMakerMock)
				// 查询 redis 失败
				cacheMock.EXPECT().
					GetSession(gomock.Any(), refreshJti).
					Return(nil, appError.ErrServerErr).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
	}

	runTC(t, testCases)
}
