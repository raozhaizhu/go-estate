package auth

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/hibiken/asynq"
	db "github.com/raozhaizhu/go-estate/internal/dao/sqlc"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/raozhaizhu/go-estate/internal/worker"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
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

	// 踢掉该用户在该设备下的所有旧会话
	_ = svc.Logout(ctx, LogoutInput{
		Username: input.Username,
		DeviceID: input.DeviceID,
	})

	// 获取访问令牌, 刷新令牌
	accessToken, refreshToken, accessPayload, refreshPayload, err := svc.forgeTokenPair(user)
	if err != nil {
		return nil, "", err
	}

	// 存刷新令牌到 db, redis
	err = svc.restoreSession(ctx, refreshPayload, input)
	if err != nil {
		return nil, "", err
	}

	// 返回 DTO
	return &DTO{
		AccessToken:          accessToken,
		AccessTokenExpiredAt: accessPayload.ExpiredAt,
		UserInfo: UserInfo{
			Username: user.Username,
			Role:     role.Role(user.Role),
		},
	}, refreshToken, nil
}

// checkUser 校验用户存在, 密码正确
func (svc *service) checkUser(ctx context.Context, input LoginInput) (*db.User, error) {
	// 查询用户
	user, err := svc.store.GetUser(ctx, input.Username)
	if err != nil { // 用户不存在
		return nil, appError.ErrWrongUsernamePassword
	}

	// 校对密码
	err = util.CheckPassword(input.Password, user.HashedPassword)
	if err != nil { // 密码错误
		return nil, appError.ErrWrongUsernamePassword // 返回账号密码错误
	}

	return &user, nil
}

// forgeTokenPair 锻造 访问令牌 + 刷新令牌
func (svc *service) forgeTokenPair(user *db.User) (accessToken, refreshToken string, accessPayload, refreshPayload *token.Payload, err error) {
	// 发放访问令牌
	accessToken, accessPayload, err = svc.tokenMaker.CreateToken(user.Username, role.Role(user.Role),
		svc.config.AccessTokenDuration, token.TokenTypeAccessToken)
	if err != nil {
		return
	}

	// 发放刷新令牌
	refreshToken, refreshPayload, err = svc.tokenMaker.CreateToken(user.Username, role.Role(user.Role),
		svc.config.RefreshTokenDuration, token.TokenTypeRefreshToken)
	if err != nil {
		return
	}

	return accessToken, refreshToken, accessPayload, refreshPayload, nil
}

// restoreSession 将会话保存到 db redis
func (svc *service) restoreSession(ctx context.Context, refreshPayload *token.Payload, input LoginInput) error {
	// 参数转化
	dbParams := refreshPayload.ToDBParams(input.UserAgent, input.ClientIp, input.DeviceID)
	cacheParams := refreshPayload.ToCacheParams()

	// 存刷新令牌到 db
	err := svc.store.CreateSession(ctx, dbParams)
	if err != nil {
		return err
	}

	// 存刷新令牌到 redis
	err = svc.sessionCache.AddNewSession(ctx, cacheParams)
	if err != nil {
		return err
	}

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
	// 查 redis
	session, err := svc.sessionCache.GetSession(ctx, jti)

	// 校验错误类型
	switch err {
	case appError.ErrMissSession: // 1. 缓存 miss, 尝试去数据库取
		dbSession, dbErr := svc.store.GetSession(ctx, jti)
		if dbErr != nil { // 数据库内也没有 session
			return appError.ErrNoSession
		}
		err := dbSession.IsValid()
		if err != nil { // 校验注销,过期
			return err
		}
		// 尽力而为, 存到缓存
		svc.sessionCache.AddNewSession(ctx, dbSession.ToCacheParams())
	case nil: // 2. 缓存命中, 校验 session 注销或过期
		err = session.IsValid()
		if err != nil {
			return err
		}
	default: // 3. 其他错误,直接返错
		return err
	}

	return nil
}

/** ====================================================================================
 * 🏁 Logout
 * =====================================================================================
 */

// Logout 用户登出
// 将用户指定设备下的 session 禁用
func (svc *service) Logout(ctx context.Context, input LogoutInput) error {
	// 获取参数
	params := input.toDBParams()

	// ->db 获取该用户指定设备下所有 session_ids
	ids, err := svc.store.GetActiveSessionIDsByUserDevice(ctx, params)
	if err != nil {
		return err
	}
	if len(ids) == 0 { //没有需要清理的 token,任务已经完成
		return nil
	}

	// ->db 将这些 sessions 清除
	err = svc.store.BlockSessionsByIDs(ctx, ids)
	if err != nil {
		return err
	}

	// worker->redis 调用 worker 异步清理 redis
	err = svc.deleteSessionsTaskEnqueue(ids)
	if err != nil {
		return err
	}

	return nil
}

// deleteSessionsTaskEnqueue 将清理会话任务加入队列
func (svc *service) deleteSessionsTaskEnqueue(ids []string) error {
	// worker->redis 调用 worker 异步清理 redis
	payload := worker.DeleteSessionsPayload{JTIs: ids}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	task := asynq.NewTask(worker.TaskDeleteSessions, payloadBytes, asynq.MaxRetry(3))
	_, err = svc.taskClient.Enqueue(task)
	if err != nil {
		slog.Error("failed to enqueue delete cache task", "error", err)
	}

	return nil
}
