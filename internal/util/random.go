package util

import (
	"fmt"
	"math/rand/v2"
	"time"

	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
)

func RandomDate(startDate, endDate time.Time) (time.Time, error) {
	// endDate 必须大于 startDate
	if startDate.After(endDate) {
		return time.Time{}, fmt.Errorf("下界不能晚于上界")
	}

	// 将时间格式化为当天的 0 点, 排除干扰
	start := startOfDay(startDate)
	end := startOfDay(endDate)

	// 计得起始日期相差多少天, 得随机日期
	daysBetween := int(end.Sub(start).Hours() / 24)
	randomOffset := rand.IntN(daysBetween + 1)

	return start.AddDate(0, 0, randomOffset), nil
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func GetRandomDayInRange() time.Time {
	day, _ := RandomDate(dailyData.MinDate, dailyData.MaxDate)
	return day
}

func GetRandom2DayInRange() (time.Time, time.Time) {
	// 默认 day1 在前,day2 在后, 若不满足则互换顺序
	day1, _ := RandomDate(dailyData.MinDate, dailyData.MaxDate)
	day2, _ := RandomDate(dailyData.MinDate, dailyData.MaxDate)
	if day1.After(day2) {
		day1, day2 = day2, day1
	}

	return day1, day2
}
