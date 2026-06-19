package dailyData

import (
	"errors"
	"fmt"
)

const ErrTimeOutOfRangeStr = "查询时间超出范围"
const ErrBadTimerOrderStr = "查询的起始时间晚于结束时间"
const ErrForbiddenCreateStr = "您的用户无权创建"

var ErrTimeOutOfRange = errors.New(ErrTimeOutOfRangeStr)
var ErrBadTimerOrder = errors.New(ErrBadTimerOrderStr)
var ErrForbiddenCreate = errors.New(ErrForbiddenCreateStr)

const errorJsonTemplate = `{"error":"%s"}`

var (
	ErrTimeOutOfRangeJson = fmt.Sprintf(errorJsonTemplate, ErrTimeOutOfRangeStr)
	ErrBadTimerOrderJson  = fmt.Sprintf(errorJsonTemplate, ErrBadTimerOrderStr)
)
