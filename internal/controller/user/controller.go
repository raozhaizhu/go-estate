package user

import (
	"github.com/gin-gonic/gin"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	userService "github.com/raozhaizhu/go-estate/internal/service/user"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

var _ userService.DTO

/** ====================================================================================
 * 🏁 Get: GetUser
 * =====================================================================================
 */

// GetUser 获取用户信息
// @Summary      查询指定用户信息
// @Description  获取指定用户的详细资料。（安全策略：仅限用户本人访问自己的数据，或者由 Admin 管理员强行调取）
// @Tags         User (用户管理)
// @Security     BearerAuth
// @Security     DeviceIDAuth
// @Security     UserAgentAuth
// @Produce      json
// @Param        username path     string true "目标用户名"
// @Success      200      {object}  response.SuccessResult[userService.DTO] "获取成功"
// @Failure      401      {object}  response.AuthErrorResult         "未登录或越权非法访问 (HTTP 返回 200, code: 401xx)"
// @Failure      404      {object}  response.NotFoundErrorResult     "该用户不存在 (HTTP 返回 200, code: 404xx)"
// @Failure      500      {object}  response.ServerErrorResult       "服务器内部错误 (HTTP 返回 500, code: 500xx)"
// @Router       /api/v1/user/{username} [get]
func (c *Controller) GetUser(ctx *gin.Context) (interface{}, error) {
	var req GetUserRequest
	// 参数错误
	if err := ctx.ShouldBindUri(&req); err != nil {
		return nil, response.MarkBindError(err)
	}

	// 参数转换
	params := req.toSvcInput()

	// -> svc 获取用户
	data, err := c.service.GetUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return data, nil
}

/** ====================================================================================
 * 🏁 Post: CreateUser
 * =====================================================================================
 */

// CreateNormalUser 创建普通用户
// @Summary      注册普通用户
// @Description  公开访问接口，用于普通用户的注册。任何人均可尝试创建。
// @Tags         User (用户管理)
// @Security     DeviceIDAuth
// @Security     UserAgentAuth
// @Accept       json
// @Produce      json
// @Param        request body     CreateUserRequest true "注册参数"
// @Success      200     {object}  response.SuccessResult[userService.DTO] "注册成功"
// @Failure      400     {object}  response.ClientErrorResult       "参数校验失败 (HTTP 返回 200, code: 400xx)"
// @Failure      409     {object}  response.ConflictErrorResult     "用户名或邮箱已存在 (HTTP 返回 200, code: 409xx)"
// @Failure      500     {object}  response.ServerErrorResult       "服务器内部错误 (HTTP 返回 500, code: 500xx)"
// @Router       /api/v1/user [post]
func (c *Controller) CreateNormalUser(ctx *gin.Context) (interface{}, error) {
	return c.createUser(ctx, role.RoleUser)
}

// CreateVip 创建 VIP 用户
// @Summary      管理员创建 VIP 用户
// @Description  受保护接口，仅限 Admin 管理员权限调用，用于直接创建带有 VIP 身份的用户。
// @Tags         User (用户管理)
// @Security     BearerAuth
// @Security     DeviceIDAuth
// @Security     UserAgentAuth
// @Accept       json
// @Produce      json
// @Param        request body     CreateUserRequest true "VIP 用户参数"
// @Success      200     {object}  response.SuccessResult[userService.DTO] "创建成功"
// @Failure      400     {object}  response.ClientErrorResult       "参数校验失败 (HTTP 返回 200, code: 400xx)"
// @Failure      401     {object}  response.AuthErrorResult         "未登录或角色权限不足 (HTTP 返回 200, code: 401xx)"
// @Failure      409     {object}  response.ConflictErrorResult     "用户名或邮箱已存在 (HTTP 返回 200, code: 409xx)"
// @Failure      500     {object}  response.ServerErrorResult       "服务器内部错误 (HTTP 返回 500, code: 500xx)"
// @Router       /api/v1/user/vip [post]
func (c *Controller) CreateVip(ctx *gin.Context) (interface{}, error) {
	return c.createUser(ctx, role.RoleVip)
}

func (c *Controller) createUser(ctx *gin.Context, role role.Role) (interface{}, error) {
	var req CreateUserRequest
	// 参数错误
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		return nil, response.MarkBindError(err)
	}

	// 参数转换
	input := req.toSvcInput()

	// -> svc 创建用户
	data, err := c.service.CreateUser(ctx, input, role)
	if err != nil {
		return nil, err
	}

	return data, nil
}

/** ====================================================================================
 * 🏁 Patch: UpdateUser
 * =====================================================================================
 */

// UpdateUser 更新用户信息
// @Summary      修改指定用户信息
// @Description  局部修改用户字段。（安全策略：仅限用户本人操作，或者由 Admin 管理员代为修改）
// @Tags         User (用户管理)
// @Security     BearerAuth
// @Security     DeviceIDAuth
// @Security     UserAgentAuth
// @Accept       json
// @Produce      json
// @Param        username path     string            true "目标用户名"
// @Param        request  body     UpdateUserRequest true "需要更新的字段参数"
// @Success      200      {object}  response.SuccessResult[userService.DTO] "修改成功"
// @Failure      400      {object}  response.ClientErrorResult       "参数解析错误或没有任何可更新的字段 (HTTP 返回 200, code: 400xx)"
// @Failure      401      {object}  response.AuthErrorResult         "未登录或越权非法操作 (HTTP 返回 200, code: 401xx)"
// @Failure      404      {object}  response.NotFoundErrorResult     "该用户不存在 (HTTP 返回 200, code: 404xx)"
// @Failure      500      {object}  response.ServerErrorResult       "服务器内部错误 (HTTP 返回 500, code: 500xx)"
// @Router       /api/v1/user/{username} [patch]
func (c *Controller) UpdateUser(ctx *gin.Context) (interface{}, error) {
	var req UpdateUserRequest
	// 参数错误
	if err := ctx.ShouldBindUri(&req); err != nil { // 解析 Uri
		return nil, response.MarkBindError(err)
	}
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil { // 解析 Json
		return nil, response.MarkBindError(err)
	}

	// 参数转换
	input := req.toSvcInput()

	// -> svc 更新用户
	data, err := c.service.UpdateUser(ctx, input)
	if err != nil {
		return nil, err
	}

	return data, nil
}
