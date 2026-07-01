package user

import (
	"database/sql"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 UserService
 * =====================================================================================
 */

// service 用户服务
type service struct {
	store db.UserStore
}

// New 返回用户服务指针
func New(store db.UserStore) *service {
	return &service{store: store}
}

// DTO 返回给 Controller 的 User 数据结构
type DTO struct {
	ID       int32     `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Role     role.Role `json:"role"`
}

/** ====================================================================================
 * 🏁 CreateUser
 * =====================================================================================
 */

// CreateUserInput
type CreateUserInput struct {
	Username string
	Password string
	Email    string
}

// toDBParams
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
	// 创建用户
	params := db.CreateUserParams{
		Username:       input.Username,
		HashedPassword: hashedPassword,
		Email:          input.Email,
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

	Password *string
	Email    *string
}

// toDBParams
func (input *UpdateUserInput) toDBParams() (db.UpdateUserParams, error) {
	params := db.UpdateUserParams{
		Username: input.Username,
	}
	// 什么都没改, 不进入数据库, 直接返错
	if input.Password == nil && input.Email == nil {
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
