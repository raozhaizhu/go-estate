package user_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	userCtrl "github.com/raozhaizhu/go-estate/internal/controller/user"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	userDomain "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	mock_worker "github.com/raozhaizhu/go-estate/internal/worker/mock"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
	mock_object_store "github.com/raozhaizhu/go-estate/pkg/object_store/mock"
	"github.com/raozhaizhu/go-estate/pkg/token"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateUser(t *testing.T) {
	username := util.RandomUsername()
	password := util.RandomPassword()
	email := util.DeriveEmail(username)
	input := user.CreateUserInput{
		Username: username,
		Password: password,
		Email:    email,
	}
	avatarKey := "uploaded-avatar.png"
	inputWithAvatar := input
	inputWithAvatar.AvatarKey = avatarKey
	longPasswordInput := input
	longPasswordInput.Password = strings.Repeat("a", 73)

	userParams := db.CreateUserParams{
		Username:  username,
		Email:     email,
		AvatarKey: objectStore.DefaultAvatarKey,
		Role:      int16(userDomain.RoleUser),
	}
	vipParams := userParams
	vipParams.AvatarKey = avatarKey
	vipParams.Role = int16(userDomain.RoleVip)
	userDTO := &user.DTO{
		ID:        1,
		Username:  username,
		Email:     email,
		Role:      userDomain.RoleUser,
		AvatarKey: objectStore.DefaultAvatarKey,
	}
	vipDTO := &user.DTO{
		ID:        2,
		Username:  username,
		Email:     email,
		Role:      userDomain.RoleVip,
		AvatarKey: avatarKey,
	}
	userRecord := db.User{
		ID:        userDTO.ID,
		Username:  username,
		Email:     email,
		Role:      int16(userDomain.RoleUser),
		AvatarKey: objectStore.DefaultAvatarKey,
	}
	vipRecord := db.User{
		ID:        vipDTO.ID,
		Username:  username,
		Email:     email,
		Role:      int16(userDomain.RoleVip),
		AvatarKey: avatarKey,
	}
	adminPayload := &token.Payload{Username: "admin", Role: userDomain.RoleAdmin}
	userPayload := &token.Payload{Username: "user", Role: userDomain.RoleUser}
	internalErr := errors.New("database unavailable")

	defaultAction := func(svc userCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		createInput, ok := input.(user.CreateUserInput)
		require.True(t, ok)
		dto, err := svc.CreateUser(ctx, createInput, userDomain.RoleUser)
		return []interface{}{dto}, err
	}
	createVipAction := func(svc userCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		createInput, ok := input.(user.CreateUserInput)
		require.True(t, ok)
		dto, err := svc.CreateUser(ctx, createInput, userDomain.RoleVip)
		return []interface{}{dto}, err
	}
	createAdminAction := func(svc userCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		createInput, ok := input.(user.CreateUserInput)
		require.True(t, ok)
		dto, err := svc.CreateUser(ctx, createInput, userDomain.RoleAdmin)
		return []interface{}{dto}, err
	}
	failCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.ErrorIs(t, actualErr, expectedErr)
		require.Len(t, results, 1)
		dto, ok := results[0].(*user.DTO)
		require.True(t, ok)
		require.Nil(t, dto)
	}
	successCheckResponse := func(expected *user.DTO) func(*testing.T, []interface{}, error, error) {
		return func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
			require.NoError(t, actualErr)
			require.Len(t, results, 1)
			dto, ok := results[0].(*user.DTO)
			require.True(t, ok)
			require.Equal(t, expected, dto)
		}
	}
	noExternalStubs := func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
	}

	testCases := []testCase{
		{
			name:          "创建 VIP 时没有认证信息",
			input:         input,
			buildStubs:    noExternalStubs,
			action:        createVipAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:       "普通用户不能创建 VIP",
			input:      input,
			buildStubs: noExternalStubs,
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), userPayload)
			},
			action:        createVipAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrAuthPermissionDenied,
		},
		{
			name:          "不支持创建管理员角色",
			input:         input,
			buildStubs:    noExternalStubs,
			action:        createAdminAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrServerErr,
		},
		{
			name:          "密码超过 bcrypt 长度限制",
			input:         longPasswordInput,
			buildStubs:    noExternalStubs,
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   bcrypt.ErrPasswordTooLong,
		},
		{
			name:       "上传的头像不存在",
			input:      inputWithAvatar,
			buildStubs: noExternalStubs,
			buildObjectStoreStubs: func(objectStoreMock *mock_object_store.MockStorageService) {
				objectStoreMock.EXPECT().
					EnsureFileExists(gomock.Any(), objectStore.AvatarBucketName, avatarKey).
					Return(appError.ErrFileNotFound).
					Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrFileNotFound,
		},
		{
			name:  "用户名重复",
			input: input,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().CreateUser(gomock.Any(), eqCreateUserParams(userParams, password)).
					Return(nil, db.ErrUsernameDuplicate).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrUserAlreadyExits,
		},
		{
			name:  "邮箱重复",
			input: input,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().CreateUser(gomock.Any(), eqCreateUserParams(userParams, password)).
					Return(nil, db.ErrEmailDuplicate).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrEmailAlreadyExits,
		},
		{
			name:  "创建用户时发生未知数据库错误",
			input: input,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().CreateUser(gomock.Any(), eqCreateUserParams(userParams, password)).
					Return(nil, internalErr).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   internalErr,
		},
		{
			name:  "创建成功后用户已不存在",
			input: input,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().CreateUser(gomock.Any(), eqCreateUserParams(userParams, password)).
					Return(mockResult{rowsAffected: 1}, nil).Times(1)
				storeMock.EXPECT().GetUser(gomock.Any(), username).Return(db.User{}, sql.ErrNoRows).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   appError.ErrUserNotFound,
		},
		{
			name:  "创建成功后读取用户发生未知错误",
			input: input,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().CreateUser(gomock.Any(), eqCreateUserParams(userParams, password)).
					Return(mockResult{rowsAffected: 1}, nil).Times(1)
				storeMock.EXPECT().GetUser(gomock.Any(), username).Return(db.User{}, internalErr).Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   internalErr,
		},
		{
			name:  "创建普通用户成功",
			input: input,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().CreateUser(gomock.Any(), eqCreateUserParams(userParams, password)).
					Return(mockResult{rowsAffected: 1}, nil).Times(1)
				storeMock.EXPECT().GetUser(gomock.Any(), username).Return(userRecord, nil).Times(1)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse(userDTO),
		},
		{
			name:  "管理员使用已上传头像创建 VIP 成功",
			input: inputWithAvatar,
			buildStubs: func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
				storeMock.EXPECT().CreateUser(gomock.Any(), eqCreateUserParams(vipParams, password)).
					Return(mockResult{rowsAffected: 1}, nil).Times(1)
				storeMock.EXPECT().GetUser(gomock.Any(), username).Return(vipRecord, nil).Times(1)
			},
			buildObjectStoreStubs: func(objectStoreMock *mock_object_store.MockStorageService) {
				objectStoreMock.EXPECT().
					EnsureFileExists(gomock.Any(), objectStore.AvatarBucketName, avatarKey).
					Return(nil).
					Times(1)
			},
			buildCtx: func() context.Context {
				return token.WithPayload(context.Background(), adminPayload)
			},
			action:        createVipAction,
			checkResponse: successCheckResponse(vipDTO),
		},
	}

	runTC(t, testCases)
}

// eqCreateUserParamsMatcher 校验除随机哈希外的所有建库参数，并验证明文密码。
type eqCreateUserParamsMatcher struct {
	arg      db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	actual, ok := x.(db.CreateUserParams)
	if !ok || util.CheckPassword(e.password, actual.HashedPassword) != nil {
		return false
	}
	e.arg.HashedPassword = actual.HashedPassword
	return reflect.DeepEqual(e.arg, actual)
}

func (e eqCreateUserParamsMatcher) String() string {
	return fmt.Sprintf("matches params %v and password %q", e.arg, e.password)
}

func eqCreateUserParams(arg db.CreateUserParams, password string) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg: arg, password: password}
}
