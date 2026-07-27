package user_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	userCtrl "github.com/raozhaizhu/go-estate/internal/controller/user"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	testUtil "github.com/raozhaizhu/go-estate/internal/test_util"
	"github.com/raozhaizhu/go-estate/internal/util"
	mock_worker "github.com/raozhaizhu/go-estate/internal/worker/mock"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
	mock_object_store "github.com/raozhaizhu/go-estate/pkg/object_store/mock"
	"github.com/raozhaizhu/go-estate/pkg/token"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 TestUpdateUser
 * =====================================================================================
 */

func TestUpdateUser(t *testing.T) {
	// 预备数据
	username, password, email := util.RandomUsername(), util.RandomPassword(), util.RandomEmail()
	correctUserInput := user.UpdateUserInput{
		Username: username,
		Password: &password,
		Email:    &email,
	}
	emailOnlyInput := user.UpdateUserInput{
		Username: username,
		Email:    &email,
	}
	avatarKey := "uploaded-avatar.png"
	avatarInput := user.UpdateUserInput{
		Username:  username,
		Email:     &email,
		AvatarKey: &avatarKey,
	}
	correctDBUser := db.User{
		Username: username,
		Email:    email,
		Role:     int16(userDomain.RoleUser),
	}
	emptyUserInput := user.UpdateUserInput{
		Username: username,
	}
	correctUserPayload := &token.Payload{
		Username: username,
		Role:     userDomain.RoleUser,
	}
	otherUserPayload := &token.Payload{
		Username: util.RandomUsername(),
		Role:     userDomain.RoleUser,
	}
	refreshStr := testUtil.RefreshStr
	jtis := []string{refreshStr}

	// 默认行为和校验逻辑逻辑
	defaultAction := func(svc userCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		updateInput, ok := input.(user.UpdateUserInput)
		require.True(t, ok)
		dto, err := svc.UpdateUser(ctx, updateInput)
		return []interface{}{dto}, err
	}
	buildCorrectCtx := func() context.Context {
		return context.WithValue(context.Background(), token.PayloadKey, correctUserPayload)
	}
	buildOtherUserCtx := func() context.Context {
		return context.WithValue(context.Background(), token.PayloadKey, otherUserPayload)
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

		dto, ok := results[0].(*user.DTO)
		require.True(t, ok)
		require.NotNil(t, dto)
		require.NotEmpty(t, dto)
		require.Equal(t, username, dto.Username)
	}

	// 成功桩函数
	stubExecTxSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().ExecTx(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, fn func(db.Querier) error) error {
				return fn(storeMock)
			}).Times(1)
	}
	stubUpdateUserSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
			DoAndReturn(func(ctx context.Context, arg db.UpdateUserParams) (sql.Result, error) {
				assert.Equal(t, username, arg.Username)
				return mockResult{rowsAffected: 1}, nil
			}).Times(1)
	}
	stubGetSessionIDsByUsernameForUpdateSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().GetSessionIDsByUsernameForUpdate(gomock.Any(), username).
			Return(jtis, nil).Times(1)
	}
	stubBlockSessionsByIDsSuccess := func(storeMock *mock_db.MockStore) {
		storeMock.EXPECT().BlockSessionsByIDs(gomock.Any(), jtis).
			Return(nil).Times(1)
	}
	stubDistributeTaskDeleteSessionsSuccess := func(distributor *mock_worker.MockTaskDistributor) {
		distributor.EXPECT().DistributeTaskDeleteSessions(gomock.Any(), jtis).
			Return(nil).Times(1)
	}

	testCases := []testCase{
		{
			name:  "密码邮箱为空,用户什么都没更新",
			input: emptyUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrEmptyUpdate,
		},
		{
			name:  "没有携带 payload",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "用户没有携带自己的 payload",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			action:        defaultAction,
			buildCtx:      buildOtherUserCtx,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrAuthPermissionDenied,
		},
		{
			name:       "更新不存在的头像",
			input:      avatarInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).Times(0)
			},
			buildObjectStoreStubs: func(objectStoreMock *mock_object_store.MockStorageService) {
				objectStoreMock.EXPECT().EnsureFileExists(gomock.Any(), objectStore.AvatarBucketName, avatarKey).
					Return(appError.ErrFileNotFound).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrFileNotFound,
		},
		{
			name:       "只更新邮箱和头像成功",
			input:      avatarInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, arg db.UpdateUserParams) (sql.Result, error) {
						require.Equal(t, username, arg.Username)
						require.Equal(t, avatarKey, arg.AvatarKey.String)
						require.True(t, arg.AvatarKey.Valid)
						require.False(t, arg.HashedPassword.Valid)
						return mockResult{rowsAffected: 1}, nil
					}).Times(1)
				storeMock.EXPECT().GetUser(gomock.Any(), username).Return(correctDBUser, nil).Times(1)
			},
			buildObjectStoreStubs: func(objectStoreMock *mock_object_store.MockStorageService) {
				objectStoreMock.EXPECT().EnsureFileExists(gomock.Any(), objectStore.AvatarBucketName, avatarKey).
					Return(nil).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: successCheckResponse,
		},
		{
			name:  "更新用户时发生内部错误",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				stubExecTxSuccess(storeMock)
				storeMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
					Return(nil, appError.ErrServerErr.WithErr(fmt.Errorf("更新用户时发生内部错误"))).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "读取受影响行数时发生错误",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				stubExecTxSuccess(storeMock)
				storeMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
					Return(mockResult{rowsAffectedErr: fmt.Errorf("rows affected failed")}, nil).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "更新用户时发现用户不存在",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				stubExecTxSuccess(storeMock)
				storeMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
					Return(mockResult{rowsAffected: 0}, nil).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrUserNotFound,
		},
		{
			name:  "只更新邮箱时邮箱重复",
			input: emailOnlyInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).
					Return(nil, db.ErrEmailDuplicate).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrEmailAlreadyExits,
		},
		{
			name:  "获取用户活跃会话时发生内部错误",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				stubExecTxSuccess(storeMock)
				stubUpdateUserSuccess(storeMock)
				storeMock.EXPECT().GetSessionIDsByUsernameForUpdate(gomock.Any(), username).
					Return([]string{}, appError.ErrServerErr).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "封禁用户活跃会话时发生内部错误",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				stubExecTxSuccess(storeMock)
				stubUpdateUserSuccess(storeMock)
				stubGetSessionIDsByUsernameForUpdateSuccess(storeMock)
				storeMock.EXPECT().BlockSessionsByIDs(gomock.Any(), jtis).
					Return(appError.ErrServerErr).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "派发任务时发生错误",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				stubExecTxSuccess(storeMock)
				stubUpdateUserSuccess(storeMock)
				stubGetSessionIDsByUsernameForUpdateSuccess(storeMock)
				stubBlockSessionsByIDsSuccess(storeMock)
				distributor.EXPECT().DistributeTaskDeleteSessions(gomock.Any(), jtis).
					Return(appError.ErrServerErr).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "更新成功后, 返回用户时发现用户不存在(事务执行后立刻删除了用户)",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				stubExecTxSuccess(storeMock)
				stubUpdateUserSuccess(storeMock)
				stubGetSessionIDsByUsernameForUpdateSuccess(storeMock)
				stubBlockSessionsByIDsSuccess(storeMock)
				stubDistributeTaskDeleteSessionsSuccess(distributor)
				storeMock.EXPECT().GetUser(gomock.Any(), username).
					Return(db.User{}, sql.ErrNoRows).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrUserNotFound,
		},
		{
			name:  "更新成功后, 返回用户时出现内部错误",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				stubExecTxSuccess(storeMock)
				stubUpdateUserSuccess(storeMock)
				stubGetSessionIDsByUsernameForUpdateSuccess(storeMock)
				stubBlockSessionsByIDsSuccess(storeMock)
				stubDistributeTaskDeleteSessionsSuccess(distributor)
				storeMock.EXPECT().GetUser(gomock.Any(), username).
					Return(db.User{}, appError.ErrServerErr).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:  "更新成功后, 返回用户成功",
			input: correctUserInput,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				stubExecTxSuccess(storeMock)
				stubUpdateUserSuccess(storeMock)
				stubGetSessionIDsByUsernameForUpdateSuccess(storeMock)
				stubBlockSessionsByIDsSuccess(storeMock)
				stubDistributeTaskDeleteSessionsSuccess(distributor)
				storeMock.EXPECT().GetUser(gomock.Any(), username).
					Return(correctDBUser, nil).Times(1)
			},
			action:        defaultAction,
			buildCtx:      buildCorrectCtx,
			checkResponse: successCheckResponse,
			expectedErr:   nil,
		},
	}

	runTC(t, testCases)
}
