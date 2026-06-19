package user

import "errors"

const ErrBadRoleStr = "角色不存在"

var ErrBadRole = errors.New(ErrBadRoleStr)
