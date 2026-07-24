package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Store interface {
	Querier
	ExecTx(ctx context.Context, fn func(q Querier) error) error
	Close() error
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
