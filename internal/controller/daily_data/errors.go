package dailyData

import (
	"errors"
	"fmt"
)

const ErrEmptyDateStr = "日期参数不能为空"
const ErrInvalidDateFormStr = "日期格式错误, 请使用 YYYY-MM-DD 格式"
const ErrQueryFailedStr = "查询数据失败"

var ErrEmptyDate = errors.New(ErrEmptyDateStr)
var ErrInvalidDateForm = errors.New(ErrInvalidDateFormStr)
var ErrQueryFailed = errors.New(ErrQueryFailedStr)

const errorJsonTemplate = `{"error":"%s"}`

var (
	ErrEmptyDateJson       = fmt.Sprintf(errorJsonTemplate, ErrEmptyDateStr)
	ErrInvalidDateFormJson = fmt.Sprintf(errorJsonTemplate, ErrInvalidDateFormStr)
	ErrQueryFailedJson     = fmt.Sprintf(errorJsonTemplate, ErrQueryFailedStr)
)
