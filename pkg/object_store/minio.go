package objectStore

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Methods
 * =====================================================================================
 */

// SetupMinIO 初始化 MinIO 客户端
func SetupMinIO(endpoint, accessKeyID, accessKey string) (StorageService, error) {
	// 初始化 minioClient, minioStorage
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, accessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, appError.NewSrvErr(err)
	}
	m := &MinioStorage{
		client: minioClient,
	}

	return m, nil
}

// CleanTestObjectStore 清理面向对象数据库
func (m *MinioStorage) CleanTestObjectStore(t *testing.T) {
	// 配置清理桶,保护文件,选项
	bucketsToClean := []string{AvatarBucketName}
	keysToProtect := []string{DefaultAvatarKey}
	opts := minio.ListObjectsOptions{Recursive: true}
	removeOpts := minio.RemoveObjectOptions{}

	// 遍历所有桶,清理非保护文件
	for _, bucket := range bucketsToClean {
		// 获取清理对象频道
		objectCh := m.client.ListObjects(context.Background(), bucket, opts)
		// 逐个清理非保护对象
		for object := range objectCh {
			require.NoError(t, object.Err, "遍历文件时出现错误,桶名: %s", bucket)
			objectKey := object.Key
			if slices.Contains(keysToProtect, objectKey) {
				continue
			}
			err := m.client.RemoveObject(context.Background(), bucket, objectKey, removeOpts)
			require.NoError(t, err, "删除文件时出现错误,对象名: %s, 桶名: %s", objectKey, bucket)
		}
	}
}

// EnsureFileExists 确认文件是否存在
func (m *MinioStorage) EnsureFileExists(ctx context.Context, bucketName, objectName string) error {
	// 校验默认头像是否存在
	_, err := m.client.StatObject(ctx, bucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return appError.ErrFileNotFound.WithErr(err)
		}
		return appError.NewSrvErr(err) // 其他未知错误
	}
	// 文件已存在
	return nil
}

// GetDownloadUrl 获取用于下载的预签名链接
func (m *MinioStorage) GetDownloadUrl(ctx context.Context, bucketName, objectKey string, expires time.Duration) (string, error) {
	u, err := m.client.PresignedGetObject(ctx, bucketName, objectKey, expires, nil)
	if err != nil {
		return "", appError.NewSrvErr(err)
	}
	return u.String(), nil
}

// GetPostPolicy 生成带安全限制的预签名 POST 表单
func (m *MinioStorage) GetPostPolicy(ctx context.Context, bucketName, objectKey string, expires time.Duration, minSize, maxSize int64, contentType string) (string, map[string]string, error) {
	// 初始化全新的 POST 策略
	policy := minio.NewPostPolicy()

	// 限制上传的目标桶,文件名和过期时间
	policy.SetBucket(bucketName)
	policy.SetKey(objectKey)
	policy.SetExpires(time.Now().UTC().Add(expires))

	// 限制文件大小
	if err := policy.SetContentLengthRange(minSize, maxSize); err != nil {
		return "", nil, appError.NewSrvErr(err)
	}

	// 限制文件类型
	if contentType != "" {
		if err := policy.SetContentType(contentType); err != nil {
			return "", nil, appError.NewSrvErr(err)
		}
	}

	// 申请策略签名
	url, formData, err := m.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return "", nil, appError.NewSrvErr(err)
	}

	return url.String(), formData, nil
}
