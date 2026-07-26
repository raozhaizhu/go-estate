package objectStore

import (
	"context"
	"time"

	"github.com/minio/minio-go/v7"
)

/** ====================================================================================
 * 🏁 Types
 * =====================================================================================
 */

// UploadResult 通用的上传返回结果
type UploadResult struct {
	ETag string
	Size int64
	URL  string
}

// StorageService 对象存储接口
type StorageService interface {
	GetPostPolicy(ctx context.Context, bucketName, objectKey string, expires time.Duration, minSize, maxSize int64, contentType string) (string, map[string]string, error)

	GetDownloadUrl(ctx context.Context, bucketName, objectKey string, expires time.Duration) (string, error)

	EnsureFileExists(ctx context.Context, bucketName, objectKey string) error
}

// MinioStorage Minio对象存储
type MinioStorage struct {
	client *minio.Client
}

/** ====================================================================================
 * 🏁 Constants
 * =====================================================================================
 */

const (
	AvatarBucketName = "avatars"
	DefaultAvatarUrl = "avatars/default_avatar.png"
	ExpireDuration   = 5 * time.Minute
)
