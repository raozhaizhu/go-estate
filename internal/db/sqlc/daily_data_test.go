package db

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	dailyData "github.com/raozhaizhu/go-estate/internal/domain/daily_data"
	"github.com/raozhaizhu/go-estate/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

/** ====================================================================================
 * 🏁 Success
 * =====================================================================================
 */

// TestGetAllData 测试GetAllData, 能正常得到所有成交数据
func TestGetAllData(t *testing.T) {
	// 查询所有成交数据
	sales, err := testStore.GetAllData(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, sales)

	// 查询结果非空
	for _, s := range sales {
		assertSaleValid(t, s)
	}
}

// TestGetDataByDay 测试GetDataByDay, 能正常获取数据, 并且数据格式正常, 日期和查询日期一致
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

// TestGetDataByPeriod 测试 GetDataByPeriod, 能正常获取数据, 并且数据格式正常, 数据日期都在指定范围内
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

/** ====================================================================================
 * 🏁 Fail
 * =====================================================================================
 */

type TestCase struct {
	name          string
	mockAction    func(mock sqlmock.Sqlmock)
	expectedError string
}

// TestGetAllData_ErrorScenarios 测试所有函数在 DB意外失败时, 行为是否符合预期(借助 sqlmock)
func TestGetAllData_ErrorScenarios(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	queries := New(db)
	columns := []string{"id", "date", "region", "category", "license_no", "project_name", "house_count", "area", "avg_price"}
	funcsToTest := []struct {
		name   string
		action func() error
	}{
		{"GetAllData", func() error { _, err := queries.GetAllData(context.Background()); return err }},
		{"GetDataByDay", func() error { _, err := queries.GetDataByDay(context.Background(), time.Now()); return err }},
	}

	errorScenarios := []TestCase{
		{
			name: "QueryError",
			mockAction: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").WillReturnError(errors.New("fail"))
			},
			expectedError: "fail",
		},
		{
			name: "QueryError",
			mockAction: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT").WillReturnError(errors.New("fail"))
			},
			expectedError: "fail",
		},
		{
			name: "ScanError",
			mockAction: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).AddRow(1, "bad-date", 0, "C", "L", "P", 10, 1.1, 1.1)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			expectedError: "Scan error",
		},
		{
			name: "ScanError",
			mockAction: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).AddRow(1, "bad-date", 0, "C", "L", "P", 10, 1.1, 1.1)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			expectedError: "Scan error",
		},
		{
			name: "RowsErr",
			mockAction: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).AddRow(1, time.Now(), 0, "C", "L", "P", 10, 1.1, 1.1).
					RowError(0, errors.New("stream err")) // 当第 0 行被读取时, 抛出读取错误
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			expectedError: "stream err",
		},
		{
			name: "RowsErr",
			mockAction: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows(columns).AddRow(1, time.Now(), 0, "C", "L", "P", 10, 1.1, 1.1).
					RowError(0, errors.New("stream err")) // 当第 0 行被读取时, 抛出读取错误
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			expectedError: "stream err",
		},
	}

	for _, f := range funcsToTest {
		for _, s := range errorScenarios {
			t.Run(s.name, func(t *testing.T) {
				// 执行 mock 模拟动作(准备查询数据, 埋下错误)
				s.mockAction(mock)
				// 获取错误, 并识别错误是否符合预期
				err := f.action()
				require.Error(t, err)
				assert.Contains(t, err.Error(), s.expectedError)
			})
		}
	}
}

/** ====================================================================================
 * 🏁 Helper
 * =====================================================================================
 */

// assertDateEqual 校验 2 日期的 年/月/日 是否相等
func assertDateEqual(t *testing.T, actualDay, expectedDay time.Time) {
	y1, m1, d1 := actualDay.Date()
	y2, m2, d2 := expectedDay.Date()
	require.Equal(t, y1, y2)
	require.Equal(t, m1, m2)
	require.Equal(t, d1, d2)
}

// assertSaleValid 校验 2 成交数据是否相等
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

// assertYearMonthEqual 校验 2 日期的 年/月 是否相等
func assertYearMonthEqual(t *testing.T, actualDay, expectedDay time.Time) {
	y1, m1, _ := actualDay.Date()
	y2, m2, _ := expectedDay.Date()
	require.Equal(t, y1, y2)
	require.Equal(t, m1, m2)
}

// assertDayInRange 校验日期是否在范围内
func assertDayInRange(t *testing.T, actualDay, startDay, endDay int) {
	require.GreaterOrEqual(t, actualDay, startDay, "日期必须大于等于 start")
	require.LessOrEqual(t, actualDay, endDay, "日期必须小于等于  end")
}
