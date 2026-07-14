package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
)

type Store interface {
	Querier
	ExecTx(ctx context.Context, fn func(q Querier) error) error
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

func InitStore(dbSource string) Store {
	conn, err := sql.Open("mysql", dbSource)
	if err != nil {
		log.Fatal("无法连接到数据库", err)
	}

	if err = conn.Ping(); err != nil {
		log.Fatal("无法 ping 通数据库", err)
	}

	return NewStore(conn)
}

/** ====================================================================================
 * 🏁 TX
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
