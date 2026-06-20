package db

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/go-sql-driver/mysql"
)

func IsUserDuplicateError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "users.username")
	}
	return false
}

func IsZeroRowsError(err error) bool {
	if errors.Is(err, sql.ErrNoRows) {
		return true
	}
	return false
}

func IsEmailDuplicateErr(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062 && strings.Contains(mysqlErr.Message, "users.email")
	}
	return false
}
