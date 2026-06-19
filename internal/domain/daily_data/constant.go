package dailyData

import (
	"fmt"
	"time"
)

const (
	MinRegion                 = 0
	MaxRegion                 = 13
	DateFormat                = "2006-01-02"
	MinDateFormatted          = "2026-05-01"
	MaxDateFormatted          = "2026-05-31"
	ExpiredDateFormatted      = "2026-06-01"
	SmallInvalidDateFormatted = "?2026-06-01"
	BigInvalidDateFormatted   = "?2099-06-01"

	DailyDataBaseUrl = "/api/v1/daily_data"
)

var (
	MinDate            = time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	MaxDate            = time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	ExpiredDate        = time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	DailyDataDayUrl    = fmt.Sprintf("%s/day", DailyDataBaseUrl)
	DailyDataPeriodUrl = fmt.Sprintf("%s/period", DailyDataBaseUrl)
	DailyDataAllUrl    = fmt.Sprintf("%s/all", DailyDataBaseUrl)
)
