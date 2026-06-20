package db

import (
	"context"
	"testing"
	"time"

	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/stretchr/testify/require"
)

func assertSaleValid(t *testing.T, sale DailyDatum) {
	require.NotZero(t, sale.ID)
	// region必须在 0-13 之内,房管网目前仅14 区
	require.GreaterOrEqual(t, int(sale.Region), dailyData.MinRegion, "Region必须>=0")
	require.LessOrEqual(t, int(sale.Region), dailyData.MaxRegion, "Region 必须<=13")
	require.NotZero(t, sale.HouseCount)
	require.NotEmpty(t, sale.Category)
	require.NotEmpty(t, sale.LicenseNo)
	require.NotEmpty(t, sale.ProjectName)
	require.NotEmpty(t, sale.Area)
	require.NotEmpty(t, sale.Category)
}

func TestGetAllData(t *testing.T) {
	// 查询正常
	sales, err := testStore.GetAllData(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, sales)
	// 查询结果非空
	for _, s := range sales {
		assertSaleValid(t, s)
	}
}

func assertDateEqual(t *testing.T, actualDay, expectedDay time.Time) {
	y1, m1, d1 := actualDay.Date()
	y2, m2, d2 := expectedDay.Date()
	require.Equal(t, y1, y2)
	require.Equal(t, m1, m2)
	require.Equal(t, d1, d2)
}

func TestGetDataByDay(t *testing.T) {
	day := util.GetRandomDayInRange()
	// 查询正常
	sales, err := testStore.GetDataByDay(context.Background(), day)
	require.NoError(t, err)
	require.NotEmpty(t, sales)
	// 查询结果非空, 随机日期和实际日期一致
	for _, s := range sales {
		assertSaleValid(t, s)
		assertDateEqual(t, s.Date, day)
	}
}

func assertYearMonthEqual(t *testing.T, actualDay, expectedDay time.Time) {
	y1, m1, _ := actualDay.Date()
	y2, m2, _ := expectedDay.Date()
	require.Equal(t, y1, y2)
	require.Equal(t, m1, m2)
}
func assertDayInRange(t *testing.T, actualDay, startDay, endDay int) {
	require.GreaterOrEqual(t, actualDay, startDay, "日期必须大于等于 start")
	require.LessOrEqual(t, actualDay, endDay, "日期必须小于等于  end")
}
func TestGetDataByPeriod(t *testing.T) {
	start, end := util.GetRandom2DayInRange()
	// 查询正常
	sales, err := testStore.GetDataByPeriod(context.Background(), GetDataByPeriodParams{start, end})
	require.NoError(t, err)
	require.NotEmpty(t, sales)
	// 查询结果非空, 实际日期在区间内
	for _, s := range sales {
		assertSaleValid(t, s)
		assertYearMonthEqual(t, s.Date, start)
		assertDayInRange(t, s.Date.Day(), start.Day(), end.Day())
	}
}
