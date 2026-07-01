package auth_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/service/auth"
	mock_service "github.com/raozhaizhu/go-estate/internal/service/auth/mock"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Login
 * =====================================================================================
 */

func TestLogin_Success(t *testing.T) {
	username := util.RandomUsername()
	password := util.RandomPassword()
	hashedPassword, _ := util.HashPassword(password)
	user := testUtil.CreateSpecificUser(t, username, password, testStore)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storeMock := mock_service.NewMockAuthStore(ctrl)
	cacheMock := mock_service.NewMockSessionCache(ctrl)

	storeMock.EXPECT().
		GetUser(gomock.Any(), username).
		Return(db.User{
			Username:       username,
			HashedPassword: hashedPassword,
		}, nil).
		Times(1)

	storeMock.EXPECT().GetActiveSessionIDsByUserDevice(gomock.Any(), gomock.Any()).
		Return([]string{}, nil).
		Times(1)

	storeMock.EXPECT().
		CreateSession(gomock.Any(), gomock.AssignableToTypeOf(db.CreateSessionParams{})).
		Return(nil).
		Times(1)

	cacheMock.EXPECT().
		AddNewSession(gomock.Any(), gomock.AssignableToTypeOf(cache.AddNewSessionParams{})).
		Return(nil).
		Times(1)

	svc := auth.New(storeMock, cacheMock, testConfig, testTokenMaker, nil)

	// Act
	dto, _, err := svc.Login(context.Background(), auth.LoginInput{
		Username: username,
		Password: password,
	})

	// Assert
	require.NoError(t, err)
	require.NotNil(t, dto)
	require.Equal(t, user.Username, dto.UserInfo.Username)
}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

type loginTC struct {
	name  string
	input auth.LoginInput
	// db,cache埋桩
	buildStubs func(storeMock *mock_service.MockAuthStore, cacheMock *mock_service.MockSessionCache)
	// 注入上下文
	buildCtx func() context.Context
	// 校验数据
	checkResponse func(t *testing.T, dto *auth.DTO, err error)
}

func runLoginTC(t *testing.T, testCases []loginTC) {
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			// 初始化 store, svc
			storeMock := mock_service.NewMockAuthStore(ctrl)
			cacheMock := mock_service.NewMockSessionCache(ctrl)

			svc := auth.New(storeMock, cacheMock, testConfig, testTokenMaker, nil)

			// 数据库埋桩
			tc.buildStubs(storeMock, cacheMock)

			// 注入上下文
			ctx := tc.buildCtx()
			dto, _, err := svc.Login(ctx, tc.input)

			// 校验一致性
			tc.checkResponse(t, dto, err)
		})
	}
}
