package user

import (
	"context"
	"log/slog"

	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	service "github.com/raozhaizhu/go-estate/internal/service/user"
)

/** ====================================================================================
 * 🏁 UserController
 * =====================================================================================
 *
 */

type Controller struct {
	service Service
	logger  *slog.Logger
}

type Service interface {
	CreateUser(ctx context.Context, p user.CreateUserInput, role role.Role) (*user.DTO, error)
	GetUser(ctx context.Context, p user.GetUserInput) (*user.DTO, error)
	UpdateUser(ctx context.Context, p user.UpdateUserInput) (*user.DTO, error)
	GenerateAvatarPresignUrl(ctx context.Context, input service.GenerateAvatarPresignUrlInput) (string, map[string]string, string, error)
}

func New(svc Service) *Controller {
	logger := slog.Default().With("layer", "user", "module", "user_controller")
	return &Controller{service: svc, logger: logger}
}

/** ====================================================================================
 * 🏁 Get: GetUser
 * =====================================================================================
 */

type GetUserRequest struct {
	Username string `uri:"username" binding:"required,min=3,max=32" example:"Bob"`
}

func (r *GetUserRequest) toSvcInput() service.GetUserInput {
	return service.GetUserInput{
		Username: r.Username,
	}
}

/** ====================================================================================
 * 🏁 Post: CreateUser
 * =====================================================================================
 */

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username  string  `json:"username" binding:"required,min=3,max=32" example:"Bob"`
	Password  string  `json:"password" binding:"required,min=8,max=16" example:"12345678"`
	Email     string  `json:"email" binding:"required,email" example:"Bob@test.com"`
	AvatarKey *string `json:"avatar_key" binding:"omitempty" example:"default_avatar.png"`
}

// toSvcInput 将CreateUserRequest转化为CreateUserInput
func (r *CreateUserRequest) toSvcInput() service.CreateUserInput {
	avatarKey := ""
	if r.AvatarKey != nil {
		avatarKey = *r.AvatarKey
	}
	return service.CreateUserInput{
		Username:  r.Username,
		Password:  r.Password,
		Email:     r.Email,
		AvatarKey: avatarKey,
	}
}

/** ====================================================================================
 * 🏁 Patch: UpdateUser
 * =====================================================================================
 */

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Username string  `uri:"username" binding:"required,min=3,max=32" example:"Bob"`
	Password *string `json:"password" binding:"omitempty,min=8,max=16" example:"12345678"`
	Email    *string `json:"email" binding:"omitempty,email" example:"Bob@test.com"`
}

// toSvcInput 将UpdateUserRequest转化为UpdateUserInput
func (r *UpdateUserRequest) toSvcInput() service.UpdateUserInput {
	return service.UpdateUserInput{
		Username: r.Username,
		Password: r.Password,
		Email:    r.Email,
	}
}

/** ====================================================================================
 * 🏁 Get: GetAvatarUploadUrl
 * =====================================================================================
 */

// GetAvatarUploadUrlRequest 上传文件请求
type GetAvatarUploadUrlRequest struct {
	Extension string `form:"extension" binding:"omitempty,oneof=png jpg jpeg webp" example:"png"`
}

// AvatarUploadData 所上传头像的数据结构, 返回给前端
type AvatarUploadData struct {
	PostUrl   string            `json:"post_url"`
	FormData  map[string]string `json:"form_data"`
	ObjectKey string            `json:"object_key"`
}

func (r *GetAvatarUploadUrlRequest) toSvcInput() service.GenerateAvatarPresignUrlInput {
	var ext, contentType string
	switch r.Extension {
	case "webp":
		ext = ".webp"
		contentType = "image/webp"
	case "jpg", "jpeg":
		ext = ".jpg"
		contentType = "image/jpeg"
	default:
		ext = ".png"
		contentType = "image/png"
	}
	return service.GenerateAvatarPresignUrlInput{
		Extension:   ext,
		ContentType: contentType,
	}
}
