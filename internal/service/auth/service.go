package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/async"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

/** ====================================================================================
 * 🏁 Login
 * =====================================================================================
 */

// Login 用户登录
func (svc *service) Login(ctx context.Context, input LoginInput) (*DTO, string, error) {
	// 校验用户存在, 密码正确
	user, err := svc.checkUser(ctx, input)
	if err != nil {
		return nil, "", err
	}

	// 获取访问令牌, 刷新令牌
	accessToken, refreshToken, accessPayload, refreshPayload, err := svc.tokenMaker.ForgeTokenPair(user, &svc.config)
	if err != nil { // 内部错误, 铸造失败
		return nil, "", err
	}

	// 在事务中: 踢掉旧会话 + 创建新会话, 并异步清理 redis
	err = svc.restoreSessionTx(ctx, refreshPayload, input)
	if err != nil {
		return nil, "", err
	}

	// 返回 DTO
	return &DTO{
		AccessToken:          accessToken,
		AccessTokenExpiredAt: accessPayload.ExpiredAt,
		UserInfo: UserInfo{
			Username:  user.Username,
			Role:      role.Role(user.Role),
			AvatarKey: user.AvatarKey,
		},
	}, refreshToken, nil
}

// checkUser 校验用户存在, 密码正确
func (svc *service) checkUser(ctx context.Context, input LoginInput) (*db.User, error) {
	// 查询用户
	user, err := svc.store.GetUser(ctx, input.Username)
	if err != nil { // 用户不存在
		return nil, appError.ErrWrongUsernamePassword // 返回账号密码错误
	}

	// 校对密码
	err = util.CheckPassword(input.Password, user.HashedPassword)
	if err != nil { // 密码错误
		return nil, appError.ErrWrongUsernamePassword // 返回账号密码错误
	}

	return &user, nil
}

// restoreSessionTx 在事务中踢掉旧会话并创建新会话, 事务成功后异步清理 redis
func (svc *service) restoreSessionTx(ctx context.Context, refreshPayload *token.Payload, input LoginInput) error {
	// 参数转化
	dbParams := refreshPayload.ToDBParams(input.UserAgent, input.ClientIp, input.DeviceID)
	cacheParams := refreshPayload.ToCacheParams()
	logoutInput := LogoutInput{
		Username: input.Username,
		DeviceID: input.DeviceID,
	}
	logoutParams := logoutInput.ToDBParams()

	// 在事务中: 踢掉旧会话 + 创建新会话
	var blockedIDs []string
	err := svc.txRunner.ExecTx(ctx, func(q db.Querier) error {
		// 获取该用户指定设备下所有活跃 session_ids
		ids, err := q.GetActiveSessionIDsByUserDeviceForUpdate(ctx, logoutParams)
		if err != nil {
			return appError.ErrServerErr.WithErr(fmt.Errorf("获取用户活跃 Sessions 失败: %w", err))
		}

		// 如果有旧会话, 先清除
		if len(ids) > 0 {
			if err := q.BlockSessionsByIDs(ctx, ids); err != nil {
				return appError.ErrServerErr.WithErr(fmt.Errorf("清除 Sessions 时失败: %w)", err))
			}
			blockedIDs = ids
		}

		// 创建新会话
		if err := q.CreateSession(ctx, dbParams); err != nil {
			return appError.ErrServerErr.WithErr(fmt.Errorf("存入会话到 DB 失败: %w)", err))
		}

		return nil
	})
	if err != nil {
		return err
	}

	// 事务成功后, 异步清理 redis 中的旧会话
	if len(blockedIDs) > 0 {
		err = svc.distributor.DistributeTaskDeleteSessions(ctx, blockedIDs)
		if err != nil {
			return appError.ErrServerErr.WithErr(fmt.Errorf("调用 worker 异步清理 redis时失败: %w)", err))
		}
	}

	// 尽力而为, 存刷新令牌到 redis
	svc.asyncAddNewSession(ctx, cacheParams)

	return nil
}

/** ====================================================================================
 * 🏁 Refresh
 * =====================================================================================
 */

// Refresh 凭借刷新令牌, 获取新的访问令牌
func (svc *service) Refresh(ctx context.Context, refreshTokenStr string) (*DTO, error) {
	// 校验刷新令牌合规
	refreshPayload, err := svc.tokenMaker.VerifyToken(refreshTokenStr, token.TokenTypeRefreshToken)
	if err != nil {
		return nil, err
	}

	// 校验令牌有效
	jti := refreshPayload.ID
	err = svc.isSessionValid(ctx, jti.String())
	if err != nil {
		return nil, err
	}

	// 注: 用户名和身份当前不允许修改
	username, userRole := refreshPayload.Username, refreshPayload.Role

	// 发放访问令牌
	accessToken, accessPayload, err := svc.tokenMaker.CreateToken(username, userRole,
		svc.config.AccessTokenDuration, token.TokenTypeAccessToken)
	if err != nil {
		return nil, err
	}

	// 返回新 accessToken
	return &DTO{
		AccessToken:          accessToken,
		AccessTokenExpiredAt: accessPayload.ExpiredAt,
		UserInfo: UserInfo{
			Username: username,
			Role:     userRole,
		},
	}, nil
}

// isSessionValid 查询 redis(若 miss 则查询 db), 校验令牌是否有效
func (svc *service) isSessionValid(ctx context.Context, jti string) error {
	// 从 redis 找 Session
	session, err := svc.sessionCache.GetSession(ctx, jti)

	// 校验错误类型
	switch err {
	case appError.ErrMissSession: // 1. 缓存 miss, 尝试去数据库取
		dbSession, dbErr := svc.store.GetSession(ctx, jti)
		if dbErr != nil { // 数据库内也没有 session
			if errors.Is(dbErr, sql.ErrNoRows) {
				return appError.ErrNoSession
			}
			return appError.ErrServerErr.WithErr(fmt.Errorf("从数据库获取 Session 失败: %w", dbErr))
		}
		err := dbSession.IsValid()
		if err != nil { // 校验注销,过期
			return err
		}
		// 尽力而为, 存到缓存
		svc.asyncAddNewSession(ctx, dbSession.ToCacheParams())

	case nil: // 2. 缓存命中, 校验 session 注销或过期
		err = session.IsValid()
		if err != nil {
			return err
		}
	default: // 3. 其他错误,直接返错
		return appError.ErrServerErr.WithErr(fmt.Errorf("缓存获取 Session 失败: %w", err))
	}

	return nil
}

// asyncAddNewSession 尽力而为, 异步将 Session 存到 Redis
func (svc *service) asyncAddNewSession(ctx context.Context, params cache.AddNewSessionParams) {
	svc.asyncRunner(ctx, svc.logger, "回写 Session 缓存", 5*time.Second, func(asyncCtx context.Context) {
		cacheErr := svc.sessionCache.AddNewSession(asyncCtx, params)
		if cacheErr != nil {
			async.LogAsyncError(asyncCtx, svc.logger, "/async/refresh_cache", "回写 Session 缓存失败", cacheErr, "jti", params.JTI)
		}
	})
}

/** ====================================================================================
 * 🏁 Logout
 * =====================================================================================
 */

// Logout 用户登出
// 将用户指定设备下的 session 禁用
func (svc *service) Logout(ctx context.Context, input LogoutInput) error {
	// 获取参数
	params := input.ToDBParams()

	// 在事务中: 获取并清除该用户指定设备下所有活跃 sessions
	var ids []string
	err := svc.txRunner.ExecTx(ctx, func(q db.Querier) error {
		// 获取该用户指定设备下所有 session_ids
		fetchedIDs, err := q.GetActiveSessionIDsByUserDeviceForUpdate(ctx, params)
		if err != nil {
			return appError.ErrServerErr.WithErr(fmt.Errorf("获取用户活跃 Sessions 失败: %w", err))
		}
		if len(fetchedIDs) == 0 { // 没有需要清理的 token, 任务已经完成
			return nil
		}

		// 将这些 sessions 清除
		if err := q.BlockSessionsByIDs(ctx, fetchedIDs); err != nil {
			return appError.ErrServerErr.WithErr(fmt.Errorf("清除 Sessions 时失败: %w)", err))
		}

		ids = fetchedIDs
		return nil
	})
	if err != nil {
		return err
	}

	// 没有需要清理的 token, 任务已经完成
	if len(ids) == 0 {
		return nil
	}

	// worker->redis 调用 worker 异步清理 redis
	err = svc.distributor.DistributeTaskDeleteSessions(ctx, ids)
	if err != nil {
		return appError.ErrServerErr.WithErr(fmt.Errorf("调用 worker 异步清理 redis时失败: %w)", err))
	}

	return nil
}
