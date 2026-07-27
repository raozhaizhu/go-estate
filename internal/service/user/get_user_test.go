package user_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	userCtrl "github.com/raozhaizhu/go-estate/internal/controller/user"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	mock_worker "github.com/raozhaizhu/go-estate/internal/worker/mock"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/require"
)

func TestGetUser(t *testing.T) {
	username := "target-user"
	input := user.GetUserInput{Username: username}
	userRecord := db.User{
		ID:        1,
		Username:  username,
		Email:     "target-user@example.com",
		Role:      int16(userDomain.RoleUser),
		AvatarKey: "avatar.png",
	}
	userDTO := &user.DTO{
		ID:        userRecord.ID,
		Username:  userRecord.Username,
		Email:     userRecord.Email,
		Role:      userDomain.RoleUser,
		AvatarKey: userRecord.AvatarKey,
	}
	selfPayload := &token.Payload{Username: username, Role: userDomain.RoleUser}
	adminPayload := &token.Payload{Username: "admin", Role: userDomain.RoleAdmin}
	otherUserPayload := &token.Payload{Username: "other-user", Role: userDomain.RoleUser}
	vipPayload := &token.Payload{Username: "vip", Role: userDomain.RoleVip}
	internalErr := errors.New("database unavailable")

	defaultAction := func(svc userCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		getInput, ok := input.(user.GetUserInput)
		require.True(t, ok)
		dto, err := svc.GetUser(ctx, getInput)
		return []interface{}{dto}, err
	}
	failCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.ErrorIs(t, actualErr, expectedErr)
		require.Len(t, results, 1)
		dto, ok := results[0].(*user.DTO)
		require.True(t, ok)
		require.Nil(t, dto)
	}
	successCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.NoError(t, actualErr)
		require.Len(t, results, 1)
		dto, ok := results[0].(*user.DTO)
		require.True(t, ok)
		require.Equal(t, userDTO, dto)
	}
	noExternalStubs := func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
	}

	testCases := []testCase{
		{
			name:          "没有认证信息",
			input:         input,
			buildStubs:    noExternalStubs,
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:       "普通用户查询其他用户",
			input:      input,
			buildStubs: noExternalStubs,
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), otherUserPayload)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrAuthPermissionDenied,
		},
		{
			name:       "VIP 查询其他用户",
			input:      input,
			buildStubs: noExternalStubs,
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), vipPayload)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrAuthPermissionDenied,
		},
		{
			name:  "用户查询自己但记录不存在",
			input: input,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().GetUser(gomock.Any(), username).Return(db.User{}, sql.ErrNoRows).Times(1)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), selfPayload)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrUserNotFound,
		},
		{
			name:  "管理员查询时发生未知数据库错误",
			input: input,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().GetUser(gomock.Any(), username).Return(db.User{}, internalErr).Times(1)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), adminPayload)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   internalErr,
		},
		{
			name:  "用户查询自己成功",
			input: input,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().GetUser(gomock.Any(), username).Return(userRecord, nil).Times(1)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), selfPayload)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
		{
			name:  "管理员查询用户成功",
			input: input,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().GetUser(gomock.Any(), username).Return(userRecord, nil).Times(1)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), adminPayload)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
	}

	runTC(t, testCases)
}
