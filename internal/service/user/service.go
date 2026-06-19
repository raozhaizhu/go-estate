package user

import (
	"context"

	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	appError "github.com/raozhaizhu/go-estate/pkg/apperror"
)

/** ====================================================================================
 * 🏁 CreateUser
 * =====================================================================================
 *
 */

func (svc *UserService) CreateUser(ctx context.Context, input CreateUserInput, role role.Role) (UserDTO, error) {
	// 初始化参数
	params, err := input.toParams(role)
	if err != nil {
		return UserDTO{}, err
	}
	// 创建用户
	_, err = svc.store.CreateUser(ctx, params)
	if err != nil {
		return UserDTO{}, svc.mapDatabaseError(err)
	}
	// 返回用户
	return svc.toDTO(ctx, params.Username)
}

/** ====================================================================================
 * 🏁 GetUser
 * =====================================================================================
 *
 */
func (svc *UserService) GetUser(ctx context.Context, input GetUserInput) (UserDTO, error) {
	// 返回用户
	return svc.toDTO(ctx, input.Username)
}

/** ====================================================================================
 * 🏁 UpdateUser
 * =====================================================================================
 *
 */
func (svc *UserService) UpdateUser(ctx context.Context, input UpdateUserInput) (UserDTO, error) {
	// 转换参数
	params, err := input.toParams()
	if err != nil {
		return UserDTO{}, err
	}
	// 更新用户
	_, err = svc.store.UpdateUser(ctx, params)
	if err != nil {
		return UserDTO{}, svc.mapDatabaseError(err)
	}

	// 返回用户
	return svc.toDTO(ctx, params.Username)
}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 *
 */
func (svc *UserService) mapDatabaseError(err error) error {
	switch {
	case db.IsUserDuplicateError(err): // 用户名已存在
		return appError.ErrUserAlreadyExits
	case db.IsEmailDuplicateErr(err): // 邮箱已存在
		return appError.ErrEmailAlreadyExits
	default:
		return err
	}
}
func (svc *UserService) toDTO(ctx context.Context, username string) (UserDTO, error) {
	// 查询用户
	user, err := svc.store.GetUser(ctx, username)
	if err != nil { // 没查到, 直接返错
		if db.IsZeroRowsError(err) {
			return UserDTO{}, appError.ErrUserNotFound
		}
		return UserDTO{}, err
	}
	// 返回用户
	return UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}, nil
}
