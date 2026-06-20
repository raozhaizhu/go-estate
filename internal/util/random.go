package util

import (
	"math/rand/v2"
	"strings"
	"time"

	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
)

/** ====================================================================================
 * 🏁 RandomDate
 * =====================================================================================
 */
func RandomDate(startDate, endDate time.Time) (time.Time, error) {
	// 开始时间不能晚于结束时间
	if startDate.After(endDate) {
		return time.Time{}, appError.ErrBadTimerOrder
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

/** ====================================================================================
 * 🏁 RandomString
 * =====================================================================================
 */

// alphabet 小写字母表
const alphabet = "abcdefghijklmnopqrstuvwxyz"

// RandomString 返回长度为 n 的随机字符串
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.IntN(k)] // 随机字符
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomUsername 返回随机用户名
func RandomUsername() string {
	return RandomString(6)
}
