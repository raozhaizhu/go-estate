package db

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrUsernameDuplicate = errors.New("用户名重复")
	ErrEmailDuplicate    = errors.New("邮箱重复")
	ErrRecordNotFound    = sql.ErrNoRows
)

func WrapDBError(err error) error {
	if err == nil {
		return nil
	}

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		if int(mysqlErr.Number) == 1062 {
			msg := mysqlErr.Message
			if strings.Contains(msg, "users.username") {
				return ErrUsernameDuplicate
			}
			if strings.Contains(msg, "users.email") {
				return ErrEmailDuplicate
			}
		}
	}

	return err
}

// func IsUsernameDuplicateError(err error) bool {
// 	var mysqlErr *mysql.MySQLError
// 	if errors.As(err, &mysqlErr) {
// 		return int(mysqlErr.Number) == 1062 && strings.Contains(mysqlErr.Message, "users.username")
// 	}
// 	return false
// }

// func IsZeroRowsError(err error) bool {
// 	if errors.Is(err, sql.ErrNoRows) {
// 		return true
// 	}
// 	return false
// }

// func IsEmailDuplicateErr(err error) bool {
// 	var mysqlErr *mysql.MySQLError
// 	if errors.As(err, &mysqlErr) {
// 		return int(mysqlErr.Number) == 1062 && strings.Contains(mysqlErr.Message, "users.email")
// 	}
// 	return false
// }
