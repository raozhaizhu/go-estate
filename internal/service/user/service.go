package user

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	objectStore "github.com/raozhaizhu/go-estate/pkg/object_store"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

/** ====================================================================================
 * 🏁 CreateUser
 * =====================================================================================
 */

// CreateUser 创建用户, 返回 UserDTO
func (svc *service) CreateUser(ctx context.Context, input CreateUserInput, roleToCreate role.Role) (*DTO, error) {
	// 校验权限
	err := svc.authorizeCreate(ctx, roleToCreate)
	if err != nil { // 无权创建, 直接返空
		return nil, err
	}

	// 校验头像是否存在
	if input.AvatarKey != "" {
		err := svc.objectStore.EnsureFileExists(ctx, objectStore.AvatarBucketName, input.AvatarKey)
		if err != nil { //文件不存在或内部错误
			return nil, err
		}
	}

	// 初始化参数
	params, err := input.toDBParams(roleToCreate)
	if err != nil {
		return nil, err
	}

	// -> db 创建用户
	_, err = svc.store.CreateUser(ctx, params)
	if err != nil {
		return nil, svc.mapDBError(err)
	}

	// -> db 返回用户
	return svc.getUserDTO(ctx, params.Username)
}

/** ====================================================================================
 * 🏁 GetUser
 * =====================================================================================
 */

// GetUser 获取UserDTO
func (svc *service) GetUser(ctx context.Context, input GetUserInput) (*DTO, error) {
	// 校验权限
	err := svc.authorizeAccess(ctx, input.Username)
	if err != nil {
		return nil, err
	}

	// -> db 返回用户
	return svc.getUserDTO(ctx, input.Username)
}

/** ====================================================================================
 * 🏁 UpdateUser
 * =====================================================================================
 */

// UpdateUser 更新用户信息, 返回 UserDTO
func (svc *service) UpdateUser(ctx context.Context, input UpdateUserInput) (*DTO, error) {
	// 转换参数
	params, err := input.ToDBParams()
	if err != nil {
		return nil, err
	}

	// 校验头像
	if input.AvatarKey != nil {
		err = svc.objectStore.EnsureFileExists(ctx, objectStore.AvatarBucketName, *input.AvatarKey)
		if err != nil {
			return nil, err
		}
		params.AvatarKey = sql.NullString{String: *input.AvatarKey, Valid: true}
	}

	// 校验权限
	err = svc.authorizeAccess(ctx, input.Username)
	if err != nil {
		return nil, err
	}

	// 事务函数, 用于在更新密码后删除旧的 token
	var jtis []string
	updateUserBlockSessions := func(q db.Querier) error {
		// 更新用户
		result, err := q.UpdateUser(ctx, params)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return appError.ErrServerErr.WithErr(fmt.Errorf("更新用户时发生内部错误: %w", err))
		}
		if rows == 0 { // 受影响条目为 0(用户不存在)
			return appError.ErrUserNotFound
		}
		// 获取用户 Sessions
		jtis, err = q.GetSessionIDsByUsernameForUpdate(ctx, params.Username)
		if err != nil {
			return appError.ErrServerErr.WithErr(fmt.Errorf("获取用户活跃 Session时发生内部错误: %w", err))
		}

		err = q.BlockSessionsByIDs(ctx, jtis)
		if err != nil {
			return err
		}

		return nil
	}

	// -> db 更新用户
	if params.HashedPassword.Valid { // 密码变更, 需标记 token并清除缓存
		err := svc.txRunner.ExecTx(ctx, updateUserBlockSessions) // -> db 标记 sessions
		if err != nil {
			return nil, err
		}
		err = svc.distributor.DistributeTaskDeleteSessions(ctx, jtis) // -> cache 任务入队
		if err != nil {
			return nil, err
		}
	} else { // 密码未变, 仅更新邮箱
		_, err = svc.store.UpdateUser(ctx, params)
		if err != nil {
			return nil, svc.mapDBError(err)
		}

	}

	// -> db 返回用户
	return svc.getUserDTO(ctx, params.Username)
}

/** ====================================================================================
 * 🏁 GenerateAvatarPresignUrl
 * =====================================================================================
 */

// GenerateAvatarPresignUrl 生成头像预签名链接
func (svc *service) GenerateAvatarPresignUrl(ctx context.Context, input GenerateAvatarPresignUrlInput) (string, map[string]string, string, error) {
	// 生成唯一 key
	objectKey := fmt.Sprintf("%s%s", uuid.New().String(), input.Extension)
	// 定义文件大小, 1kb-5mb
	var minSize, maxSize int64 = 1024, 5 * 1024 * 1024

	// 调用存储层生成 Policy
	postUrl, formData, err := svc.objectStore.GetPostPolicy(
		ctx,
		objectStore.AvatarBucketName,
		objectKey,
		objectStore.ExpireDuration,
		minSize,
		maxSize,
		input.ContentType,
	)
	if err != nil {
		return "", nil, "", err
	}

	return postUrl, formData, objectKey, nil

}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

// mapDBError 集中处理错误: 创建和更新用户
func (svc *service) mapDBError(err error) error {
	wrappedErr := db.WrapDBError(err)

	switch {
	case errors.Is(wrappedErr, db.ErrUsernameDuplicate): // 用户名已存在
		return appError.ErrUserAlreadyExits
	case errors.Is(wrappedErr, db.ErrEmailDuplicate): // 邮箱已存在
		return appError.ErrEmailAlreadyExits
	default:
		return err
	}
}

// getUserDTO 从数据库获取用户信息, 过滤为 DTO 后返回
func (svc *service) getUserDTO(ctx context.Context, username string) (*DTO, error) {
	// -> db 查询用户
	user, err := svc.store.GetUser(ctx, username)
	if err != nil { // 没查到
		if errors.Is(db.WrapDBError(err), db.ErrRecordNotFound) {
			return nil, appError.ErrUserNotFound
		}
		// 未知错误
		return nil, err
	}

	// -> db 返回用户
	return &DTO{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		Role:      role.Role(user.Role),
		AvatarKey: user.AvatarKey,
	}, nil
}

// authorizeCreate 校验当前用户是否有权限创建该角色
func (svc *service) authorizeCreate(ctx context.Context, roleToCreate role.Role) error {
	// 校验权限
	switch roleToCreate {
	case role.RoleUser: // 创建 User, 直接放行
		return nil
	case role.RoleVip: // 创建 Vip, 身份必须是 Admin
		// 获取 payload
		payload, err := token.GetPayload(ctx)
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
		return appError.ErrServerErr.WithErr(fmt.Errorf("verifyCreatePermission 抵达了不应抵达的位置"))
	}
}

// authorizeAccess 统一校验权限: 必须是管理员或者本人
func (svc *service) authorizeAccess(ctx context.Context, targetUsername string) error {
	// 获取 payload
	payload, err := token.GetPayload(ctx)
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
