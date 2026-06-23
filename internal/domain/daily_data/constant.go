package dailyDataDomain

import (
	"fmt"
	"time"
)

const (
	MinRegion          = 0
	MaxRegion          = 13
	DateFormat         = "2006-01-02"
	MinDateStr         = "2026-05-01"
	MaxDateStr         = "2026-05-31"
	ExpiredDateStr     = "2026-06-01"
	MalformedSmallDate = "?2026-06-01"
	MalformedBigDate   = "?2099-06-01"
)

var (
	MinDate     = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	MaxDate     = time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	ExpiredDate = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	DailyDataBaseUrl   = "/api/v1/daily_data"
	DailyDataDayUrl    = fmt.Sprintf("%s/day", DailyDataBaseUrl)
	DailyDataPeriodUrl = fmt.Sprintf("%s/period", DailyDataBaseUrl)
	DailyDataAllUrl    = fmt.Sprintf("%s/all", DailyDataBaseUrl)
)
