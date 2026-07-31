package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type Store interface {
	Querier

	// 生命周期
	Close() error
	CleanTestStore(t *testing.T)

	// 事务包装方法
	ExecTx(ctx context.Context, fn func(q Querier) error) error
	GetDataAndDeductPointsTx(ctx context.Context, arg GetDataTxParams) (GetDataTxResult, error)
}

type SQLStore struct {
	db *sql.DB
	*Queries
}

func NewStore(db *sql.DB) Store {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

func InitStore(dbSource string) (Store, error) {
	conn, err := sql.Open("mysql", dbSource)
	if err != nil {
		return nil, fmt.Errorf("无法打开数据库连接: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err = conn.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("无法 ping 通数据库: %w", err)
	}

	return NewStore(conn), nil
}

/** ====================================================================================
 * 🏁 Methods
 * =====================================================================================
 */

func (s *SQLStore) ExecTx(ctx context.Context, fn func(q Querier) error) error {
	// 开始事务
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// 执行查询
	q := s.Queries.WithTx(tx)
	err = fn(q)
	if err != nil { // 执行报错
		rbErr := tx.Rollback()
		if rbErr != nil {
			return fmt.Errorf("事务错误: %v, 回滚错误: %v", err, rbErr)
		}
		return err
	}

	// 返回结果
	return tx.Commit()
}

func (s *SQLStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// CleanTestStore 清理测试数据库
func (srv *SQLStore) CleanTestStore(t *testing.T) {
	dbConn := srv.db

	// 1. 关闭外键检查
	_, err := dbConn.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	require.NoError(t, err, "关闭外键检查失败")

	// 2. 清空相关表
	_, err = dbConn.Exec("TRUNCATE TABLE users;")
	require.NoError(t, err, "清空 users 表失败")

	_, err = dbConn.Exec("TRUNCATE TABLE sessions;")
	require.NoError(t, err, "清空 sessions 表失败")

	// 3. 恢复外键检查
	_, err = dbConn.Exec("SET FOREIGN_KEY_CHECKS = 1;")
	require.NoError(t, err, "恢复外键检查失败")
}
