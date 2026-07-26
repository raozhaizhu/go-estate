package user

import (
	"database/sql"
	"time"

	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	"github.com/raozhaizhu/go-estate/internal/domain/app"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/internal/worker"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
)

/** ====================================================================================
 * 🏁 UserService
 * =====================================================================================
 */

// service 用户服务
type service struct {
	store        db.UserStore
	txRunner     db.TxRunner
	sessionCache cache.SessionCache
	distributor  worker.TaskDistributor
	objectStore  objectStore.StorageService
}

// New 返回用户服务指针
func New(deps app.Deps) *service {
	return &service{store: deps.Store, txRunner: deps.Store, sessionCache: deps.Cache, distributor: deps.Distributor, objectStore: deps.ObjectStore}
}

// DTO 返回给 Controller 的 User 数据结构
type DTO struct {
	ID        int32     `json:"id" example:"1"`
	Username  string    `json:"username" example:"Bob"`
	Email     string    `json:"email" example:"Bob@test.com"`
	Role      role.Role `json:"role" example:"1"`
	AvatarKey string    `json:"avatar_key" example:"avatars/default_avatar.png"`
}

/** ====================================================================================
 * 🏁 CreateUser
 * =====================================================================================
 */

// CreateUserInput 创建用户输入
type CreateUserInput struct {
	Username  string
	Password  string
	Email     string
	AvatarKey string
}

// toDBParams 将CreateUserInput转化为CreateUserParams
func (input *CreateUserInput) toDBParams(role role.Role) (db.CreateUserParams, error) {
	// 角色类型必须合法
	if !role.IsValid() {
		return db.CreateUserParams{}, ErrBadRole
	}

	// 哈希密码
	hashedPassword, err := util.HashPassword(input.Password)
	if err != nil {
		return db.CreateUserParams{}, err
	}

	// 设置头像
	if input.AvatarKey != "" {
		input.AvatarKey = objectStore.DefaultAvatarUrl
	}

	// 创建用户
	params := db.CreateUserParams{
		Username:       input.Username,
		HashedPassword: hashedPassword,
		Email:          input.Email,
		AvatarKey:      input.AvatarKey,
		Role:           int16(role),
	}

	return params, nil
}

/** ====================================================================================
 * 🏁 GetUser
 * =====================================================================================
 */

// GetUserInput
type GetUserInput struct {
	Username string
}

/** ====================================================================================
 * 🏁 UpdateUser
 * =====================================================================================
 */

// UpdateUserInput
type UpdateUserInput struct {
	Username string

	Password  *string
	Email     *string
	AvatarKey *string
}

// ToDBParams
func (input *UpdateUserInput) ToDBParams() (db.UpdateUserParams, error) {
	params := db.UpdateUserParams{
		Username: input.Username,
	}
	// 什么都没改, 不进入数据库, 直接返错
	if input.Password == nil && input.Email == nil && input.AvatarKey == nil {
		return db.UpdateUserParams{}, appError.ErrEmptyUpdate
	}
	// 更新密码和改密时间
	if input.Password != nil {
		hashedPassword, err := util.HashPassword(*input.Password)
		if err != nil {
			return db.UpdateUserParams{}, err
		}
		params.HashedPassword = sql.NullString{String: hashedPassword, Valid: true}
		params.PasswordChangedAt = sql.NullTime{Time: time.Now(), Valid: true}
	}
	// 更新 Email
	if input.Email != nil {
		params.Email = sql.NullString{String: *input.Email, Valid: true}
	}

	return params, nil
}

/** ====================================================================================
 * 🏁 GenerateAvatarPresignUrl
 * =====================================================================================
 */

type GenerateAvatarPresignUrlInput struct {
	Extension   string
	ContentType string
}
