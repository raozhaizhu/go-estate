package service

import (
	"errors"
	"fmt"
)

const ErrTimeOutOfRangeStr = "查询时间超出范围"
const ErrBadTimerOrderStr = "查询的起始时间晚于结束时间"

var ErrTimeOutOfRange = errors.New(ErrTimeOutOfRangeStr)
var ErrBadTimerOrder = errors.New(ErrBadTimerOrderStr)

const errorJsonTemplate = `{"error":"%s"}`

var (
	ErrTimeOutOfRangeJson = fmt.Sprintf(errorJsonTemplate, ErrTimeOutOfRangeStr)
	ErrBadTimerOrderJson  = fmt.Sprintf(errorJsonTemplate, ErrBadTimerOrderStr)
)
