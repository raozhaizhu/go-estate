package db

import (
	"context"
	"database/sql"
	"time"

	"github.com/raozhaizhu/go-estate/internal/dao/cache"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 Stores
 * =====================================================================================
 */

type SessionStore interface {
	GetSession(ctx context.Context, id string) (Session, error)
	CreateSession(ctx context.Context, arg CreateSessionParams) error
	GetActiveSessionIDsByUserDevice(ctx context.Context, arg GetActiveSessionIDsByUserDeviceParams) ([]string, error)
	BlockSessionsByIDs(ctx context.Context, ids []string) error
}

type UserStore interface {
	CreateUser(ctx context.Context, arg CreateUserParams) (sql.Result, error)
	GetUser(ctx context.Context, username string) (User, error)
	UpdateUser(ctx context.Context, arg UpdateUserParams) (sql.Result, error)
}

type DailyDataStore interface {
	GetDataByDay(ctx context.Context, targetDate time.Time) ([]DailyDatum, error)
	GetDataByPeriod(ctx context.Context, arg GetDataByPeriodParams) ([]DailyDatum, error)
	GetAllData(ctx context.Context) ([]DailyDatum, error)
}

type AuthStore interface {
	GetUser(ctx context.Context, username string) (User, error)
	SessionStore
}

/** ====================================================================================
 * 🏁 Models
 * =====================================================================================
 */

func (s *Session) ToCacheParams() cache.AddNewSessionParams {
	return cache.AddNewSessionParams{
		JTI: s.ID,
		Session: cache.Session{
			Username:  s.Username,
			IsBlocked: s.IsBlocked,
			ExpiresAt: s.ExpiresAt,
		},
	}
}

func (s *Session) IsValid() error {
	// 会话已注销
	if s.IsBlocked {
		return appError.ErrBlockedSession
	}

	// 会话已过期
	if time.Now().After(s.ExpiresAt) {
		return appError.ErrExpiredToken
	}

	return nil
}
