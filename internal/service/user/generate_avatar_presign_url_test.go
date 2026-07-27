package user_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	userCtrl "github.com/raozhaizhu/go-estate/internal/controller/user"
	mock_db "github.com/raozhaizhu/go-estate/internal/dao/mock"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	mock_worker "github.com/raozhaizhu/go-estate/internal/worker/mock"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
	mock_object_store "github.com/raozhaizhu/go-estate/pkg/object_store/mock"
	mock_token "github.com/raozhaizhu/go-estate/pkg/token/mock"
	"github.com/stretchr/testify/require"
)

func TestGenerateAvatarPresignURL(t *testing.T) {
	input := user.GenerateAvatarPresignUrlInput{
		Extension:   ".webp",
		ContentType: "image/webp",
	}
	postURL := "https://object-store.example.com/upload"
	formData := map[string]string{"policy": "signed-policy"}
	storageErr := errors.New("object store unavailable")

	defaultAction := func(svc userCtrl.Service, ctx context.Context, input interface{}) ([]interface{}, error) {
		presignInput, ok := input.(user.GenerateAvatarPresignUrlInput)
		require.True(t, ok)
		url, data, objectKey, err := svc.GenerateAvatarPresignUrl(ctx, presignInput)
		return []interface{}{url, data, objectKey}, err
	}
	noExternalStubs := func(storeMock *mock_db.MockStore, cacheMock *mock_db.MockCache, distributor *mock_worker.MockTaskDistributor, tokenMakerMock *mock_token.MockMaker) {
	}
	failCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.ErrorIs(t, actualErr, expectedErr)
		require.Equal(t, []interface{}{"", map[string]string(nil), ""}, results)
	}
	successCheckResponse := func(t *testing.T, results []interface{}, actualErr, expectedErr error) {
		require.NoError(t, actualErr)
		require.Len(t, results, 3)
		require.Equal(t, postURL, results[0])
		require.Equal(t, formData, results[1])
		objectKey, ok := results[2].(string)
		require.True(t, ok)
		require.True(t, strings.HasSuffix(objectKey, input.Extension))
		_, err := uuid.Parse(strings.TrimSuffix(objectKey, input.Extension))
		require.NoError(t, err)
	}

	testCases := []testCase{
		{
			name:       "对象存储生成策略失败",
			input:      input,
			buildStubs: noExternalStubs,
			buildObjectStoreStubs: func(objectStoreMock *mock_object_store.MockStorageService) {
				objectStoreMock.EXPECT().
					GetPostPolicy(gomock.Any(), objectStore.AvatarBucketName, gomock.Any(), objectStore.ExpireDuration, int64(1024), int64(5*1024*1024), input.ContentType).
					Return("", nil, storageErr).
					Times(1)
			},
			action:        defaultAction,
			checkResponse: failCheckResponse,
			expectedErr:   storageErr,
		},
		{
			name:       "生成策略成功",
			input:      input,
			buildStubs: noExternalStubs,
			buildObjectStoreStubs: func(objectStoreMock *mock_object_store.MockStorageService) {
				objectStoreMock.EXPECT().
					GetPostPolicy(gomock.Any(), objectStore.AvatarBucketName, gomock.Any(), objectStore.ExpireDuration, int64(1024), int64(5*1024*1024), input.ContentType).
					DoAndReturn(func(ctx context.Context, bucketName, objectKey string, expires time.Duration, minSize, maxSize int64, contentType string) (string, map[string]string, error) {
						require.Equal(t, objectStore.AvatarBucketName, bucketName)
						require.Equal(t, objectStore.ExpireDuration, expires)
						require.Equal(t, int64(1024), minSize)
						require.Equal(t, int64(5*1024*1024), maxSize)
						require.Equal(t, input.ContentType, contentType)
						require.True(t, strings.HasSuffix(objectKey, input.Extension))
						return postURL, formData, nil
					}).
					Times(1)
			},
			action:        defaultAction,
			checkResponse: successCheckResponse,
		},
	}

	runTC(t, testCases)
}
