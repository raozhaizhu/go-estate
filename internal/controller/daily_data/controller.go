package dailyData

import (
	"github.com/gin-gonic/gin"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 *
 */
func (c *DailyDataController) GetDataByDay(ctx *gin.Context) (interface{}, error) {
	var req GetDataByDayRequest
	// 参数错误
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, response.MarkBindError(err)
	}
	// 参数转换
	params, err := req.toSvcParams()
	if err != nil {
		return nil, err
	}
	// -> svc 获得日成交数据
	data, err := c.service.GetDataByDay(ctx, params)
	if err != nil {
		return nil, err
	}

	return data, nil
}

/** ====================================================================================
 * 🏁 GetDataByPeriod
 * =====================================================================================
 *
 */
func (c *DailyDataController) GetDataByPeriod(ctx *gin.Context) (interface{}, error) {
	var req GetDataByPeriodRequest
	// 参数错误
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, response.MarkBindError(err)
	}
	// 参数转换
	params, err := req.toSvcParams()
	if err != nil {
		return nil, err

	}
	// -> svc 获得周期成交数据
	data, err := c.service.GetDataByPeriod(ctx, params)
	if err != nil {
		return nil, err

	}

	return data, nil
}

/** ====================================================================================
 * 🏁 GetAllData
 * =====================================================================================
 *
 */
func (c *DailyDataController) GetAllData(ctx *gin.Context) (interface{}, error) {
	// -> svc 获得所有数据
	data, err := c.service.GetAllData(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}
