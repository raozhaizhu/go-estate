package user

import (
	"context"
	"log"

	db "github.com/raozhaizhu/go-estate/internal/db/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/middleware"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 CreateUser
 * =====================================================================================
 */

// CreateUser 创建用户, 返回 UserDTO
func (svc *UserService) CreateUser(ctx context.Context, input CreateUserInput, roleToCreate role.Role) (UserDTO, error) {
	// 初始化参数
	params, err := input.toDBParams(roleToCreate)
	if err != nil {
		return UserDTO{}, err
	}

	// 校验权限
	err = svc.authorizeCreate(ctx, roleToCreate)
	if err != nil { // 无权创建, 直接返空
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
 */

// GetUser 获取UserDTO
func (svc *UserService) GetUser(ctx context.Context, input GetUserInput) (UserDTO, error) {
	// 校验权限
	err := svc.authorizeAccess(ctx, input.Username)
	if err != nil {
		return UserDTO{}, err
	}

	// -> db 返回用户
	return svc.getUserDTO(ctx, input.Username)
}

/** ====================================================================================
 * 🏁 UpdateUser
 * =====================================================================================
 */

// UpdateUser 更新用户信息, 返回 UserDTO
func (svc *UserService) UpdateUser(ctx context.Context, input UpdateUserInput) (UserDTO, error) {
	// 转换参数
	params, err := input.toDBParams()
	if err != nil {
		return UserDTO{}, err
	}

	// 校验权限
	err = svc.authorizeAccess(ctx, input.Username)
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

	// 校验权限

	// -> db 返回用户
	return UserDTO{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Role:     role.Role(user.Role),
	}, nil
}

// authorizeCreate 校验当前用户是否有权限创建该角色
func (svc *UserService) authorizeCreate(ctx context.Context, roleToCreate role.Role) error {
	// 校验权限
	switch roleToCreate {
	case role.RoleUser: // 创建 User, 直接放行
		return nil
	case role.RoleVip: // 创建 Vip, 身份必须是 Admin
		// 获取 payload
		payload, err := middleware.GetPayload(ctx)
		if err != nil {
			return err
		}

		// 校验当前身份为 Admin
		currRole := payload.Role
		if currRole == role.RoleAdmin {
			return nil
		}

		return appError.ErrAuthPermissionDenied
	default:
		log.Printf("verifyCreatePermission 抵达了不应抵达的位置")
		return appError.ErrServerErr
	}
}

// authorizeAccess 统一校验权限: 必须是管理员或者本人
func (svc *UserService) authorizeAccess(ctx context.Context, targetUsername string) error {
	// 获取 payload
	payload, err := middleware.GetPayload(ctx)
	if err != nil {
		return err
	}

	// 获取当前用户名和身份
	currUsername, currRole := payload.Username, payload.Role

	// 如果是 Admin, 直接通过
	if currRole == role.RoleAdmin {
		return nil
	}

	// 如果是本人, 直接通过
	if currUsername == targetUsername {
		return nil
	}

	// 无权操作, 返回错误
	return appError.ErrAuthPermissionDenied
}
