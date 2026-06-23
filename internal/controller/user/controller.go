package user

import (
	"github.com/gin-gonic/gin"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

/** ====================================================================================
 * 🏁 Get: GetUser
 * =====================================================================================
 */

// GetUser 获取用户信息
// Get: /api/v1/user/:username
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
// Post: /api/v1/user
func (c *Controller) CreateNormalUser(ctx *gin.Context) (interface{}, error) {
	return c.createUser(ctx, role.RoleUser)
}

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
	params := req.toSvcInput()

	// -> svc 创建用户
	data, err := c.service.CreateUser(ctx, params, role)
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
// Patch: /api/v1/user/:username
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
	params := req.toSvcInput()

	// -> svc 更新用户
	data, err := c.service.UpdateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	return data, nil
}
