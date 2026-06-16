package service

import "errors"

var ErrTimeOutOfRange = errors.New("查询时间超出范围")
var ErrBadTimerOrder = errors.New("查询的起始时间晚于结束时间")
