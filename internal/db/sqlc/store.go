package db

import (
	"database/sql"
	"log"
)

type Store interface {
	Querier
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
