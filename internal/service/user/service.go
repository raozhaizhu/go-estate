package user

import (
	"context"

	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 CreateUser
 * =====================================================================================
 *
 */

// CreateUser 创建用户, 返回 UserDTO
func (svc *UserService) CreateUser(ctx context.Context, input CreateUserInput, role role.Role) (UserDTO, error) {
	// 初始化参数
	params, err := input.toDBParams(role)
	if err != nil {
		return UserDTO{}, err
	}
	// -> db 创建用户
	_, err = svc.store.CreateUser(ctx, params)
	if err != nil {
		return UserDTO{}, svc.mapDBError(err)
	}
	// -> db 返回用户
	return svc.getUserDTO(ctx, params.Username)
}

/** ====================================================================================
 * 🏁 GetUser
 * =====================================================================================
 *
 */

// GetUser 获取UserDTO
func (svc *UserService) GetUser(ctx context.Context, input GetUserInput) (UserDTO, error) {
	// -> db 返回用户
	return svc.getUserDTO(ctx, input.Username)
}

/** ====================================================================================
 * 🏁 UpdateUser
 * =====================================================================================
 *
 */

// UpdateUser 更新用户信息, 返回 UserDTO
func (svc *UserService) UpdateUser(ctx context.Context, input UpdateUserInput) (UserDTO, error) {
	// 转换参数
	params, err := input.toDBParams()
	if err != nil {
		return UserDTO{}, err
	}
	// -> db 更新用户
	_, err = svc.store.UpdateUser(ctx, params)
	if err != nil {
		return UserDTO{}, svc.mapDBError(err)
	}

	// -> db 返回用户
	return svc.getUserDTO(ctx, params.Username)
}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 *
 */

// mapDBError 集中处理错误: 创建和更新用户
func (svc *UserService) mapDBError(err error) error {
	switch {
	case db.IsUserDuplicateError(err): // 用户名已存在
		return appError.ErrUserAlreadyExits
	case db.IsEmailDuplicateErr(err): // 邮箱已存在
		return appError.ErrEmailAlreadyExits
	default:
		return err
	}
}

// getUserDTO 从数据库获取用户信息, 过滤为 DTO 后返回
func (svc *UserService) getUserDTO(ctx context.Context, username string) (UserDTO, error) {
	// -> db 查询用户
	user, err := svc.store.GetUser(ctx, username)
	if err != nil { // 没查到
		if db.IsZeroRowsError(err) {
			return UserDTO{}, appError.ErrUserNotFound
		}
		// 未知错误
		return UserDTO{}, err
	}
	// -> db 返回用户
	return UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     role.Role(user.Role),
	}, nil
}
