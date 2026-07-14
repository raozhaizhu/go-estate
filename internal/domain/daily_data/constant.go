package dailyDataDomain

import (
	"fmt"
	"time"
)

const (
	MinRegion        = 0
	MaxRegion        = 13
	DateFormat       = "2006-01-02"
	MinDateStr       = "2026-05-01"
	MaxDateStr       = "2026-05-31"
	ExpiredDateStr   = "2026-06-01"
	MalformedDateStr = "?2026-06-01"
)

var (
	MinDate, _     = time.Parse(DateFormat, MinDateStr)
	MaxDate, _     = time.Parse(DateFormat, MaxDateStr)
	ExpiredDate, _ = time.Parse(DateFormat, ExpiredDateStr)

	DailyDataBaseUrl   = "/api/v1/daily_data"
	DailyDataDayUrl    = fmt.Sprintf("%s/day", DailyDataBaseUrl)
	DailyDataPeriodUrl = fmt.Sprintf("%s/period", DailyDataBaseUrl)
	DailyDataAllUrl    = fmt.Sprintf("%s/all", DailyDataBaseUrl)
)
