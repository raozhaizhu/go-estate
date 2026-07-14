package dailyData

import (
	"github.com/gin-gonic/gin"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

/** ====================================================================================
 * 🏁 GetDataByDay
 * =====================================================================================
 */

// GetDataByDay 按日获取楼盘成交数据
// @Summary      按日获取楼盘成交数据
// @Description  获取指定单日的楼盘成交数据（需登录，且至少需要 User 权限）
// @Tags         DailyData (成交数据)
// @Security     BearerAuth
// @Security     DeviceIDAuth
// @Security     UserAgentAuth
// @Produce      json
// @Param        request query     GetDataByDayRequest true "查询参数"
// @Success      200     {object}  response.SuccessResult[dailyData.DailyDataList] "获取成功"
// @Failure      400     {object}  response.ClientErrorResult              "参数校验失败 (HTTP 返回 200, code: 400xx)"
// @Failure      401     {object}  response.AuthErrorResult                "未登录或权限不足 (HTTP 返回 200, code: 401xx)"
// @Failure      500     {object}  response.ServerErrorResult              "服务器内部错误 (HTTP 返回 500, code: 500xx)"
// @Router       /api/v1/daily_data/day [get]
func (c *Controller) GetDataByDay(ctx *gin.Context) (interface{}, error) {
	var req GetDataByDayRequest
	// 参数错误
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, response.MarkBindError(err)
	}
	// 参数转换
	input := req.toSvcInput()

	// -> svc 获得日成交数据
	data, err := c.service.GetDataByDay(ctx, input)
	if err != nil {
		return nil, err
	}

	return data, nil
}

/** ====================================================================================
 * 🏁 GetDataByPeriod
 * =====================================================================================
 */

// GetDataByPeriod 按周期获取楼盘成交数据
// @Summary      按周期获取楼盘成交数据
// @Description  获取指定时间范围内的楼盘成交数据（需登录，且至少需要 VIP 权限）
// @Tags         DailyData (成交数据)
// @Security     BearerAuth
// @Security     DeviceIDAuth
// @Security     UserAgentAuth
// @Produce      json
// @Param        request query     GetDataByPeriodRequest true "时间范围查询参数"
// @Success      200     {object}  response.SuccessResult[dailyData.DailyDataList] "获取成功"
// @Failure      400     {object}  response.ClientErrorResult              "参数校验失败 (HTTP 返回 200, code: 400xx)"
// @Failure      401     {object}  response.AuthErrorResult                "未登录或权限不足 (HTTP 返回 200, code: 401xx)"
// @Failure      500     {object}  response.ServerErrorResult              "服务器内部错误 (HTTP 返回 500, code: 500xx)"
// @Router       /api/v1/daily_data/period [get]
func (c *Controller) GetDataByPeriod(ctx *gin.Context) (interface{}, error) {
	var req GetDataByPeriodRequest
	// 参数错误
	if err := ctx.ShouldBindQuery(&req); err != nil {
		return nil, response.MarkBindError(err)
	}
	// 参数转换
	params, err := req.toSvcInput()
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
 */

// GetAllData 获取所有楼盘成交数据
// @Summary      获取所有楼盘成交数据
// @Description  获取系统内所有的楼盘成交数据（需登录，且仅限 Admin 权限访问）
// @Tags         DailyData (成交数据)
// @Security     BearerAuth
// @Security     DeviceIDAuth
// @Security     UserAgentAuth
// @Produce      json
// @Success      200     {object}  response.SuccessResult[dailyData.DailyDataList] "获取成功"
// @Failure      401     {object}  response.AuthErrorResult                "未登录或权限不足 (HTTP 返回 200, code: 401xx)"
// @Failure      500     {object}  response.ServerErrorResult              "服务器内部错误 (HTTP 返回 500, code: 500xx)"
// @Router       /api/v1/daily_data/all [get]
func (c *Controller) GetAllData(ctx *gin.Context) (interface{}, error) {
	// -> svc 获得所有数据
	data, err := c.service.GetAllData(ctx)
	if err != nil {
		return nil, err
	}

	return data, nil
}
