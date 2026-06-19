package user

import (
	"context"
	"database/sql"
	"time"

	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	appError "github.com/raozhaizhu/go-estate/pkg/apperror"
)

/** ====================================================================================
 * 🏁 UserService
 * =====================================================================================
 *
 */
type UserService struct {
	store UserQuerier
}

type UserQuerier interface {
	CreateUser(ctx context.Context, arg db.CreateUserParams) (sql.Result, error)
	GetUser(ctx context.Context, username string) (db.User, error)
	UpdateUser(ctx context.Context, arg db.UpdateUserParams) (sql.Result, error)
}

func NewUserService(store UserQuerier) *UserService {
	return &UserService{store: store}
}

type UserDTO struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

/** ====================================================================================
 * 🏁 CreateUser
 * =====================================================================================
 *
 */
type CreateUserInput struct {
	Username string
	Password string
	Email    string
}

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
 *
 */
type GetUserInput struct {
	Username string
}

/** ====================================================================================
 * 🏁 UpdateUser
 * =====================================================================================
 *
 */
type UpdateUserInput struct {
	Username string

	Password *string
	Email    *string
}

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
