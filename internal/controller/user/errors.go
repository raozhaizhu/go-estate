package user

import (
	"errors"
	"fmt"
)

const ErrEmptyDateStr = "用户参数不合规"

var ErrEmptyDate = errors.New(ErrEmptyDateStr)

const errorJsonTemplate = `{"error":"%s"}`

var (
	ErrEmptyDateJson = fmt.Sprintf(errorJsonTemplate, ErrEmptyDateStr)
)
