package user

import (
	"context"

	"github.com/gin-gonic/gin"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	service "github.com/raozhaizhu/go-estate/internal/service/user"
	response "github.com/raozhaizhu/go-estate/pkg/api"
)

/** ====================================================================================
 * 🏁 UserController
 * =====================================================================================
 *
 */
type UserController struct {
	service UserQuerier
}

type UserQuerier interface {
	CreateUser(ctx context.Context, p user.CreateUserInput, role role.Role) (user.UserDTO, error)
	GetUser(ctx context.Context, p user.GetUserInput) (user.UserDTO, error)
	UpdateUser(ctx context.Context, p user.UpdateUserInput) (user.UserDTO, error)
}

func NewUserController(svc UserQuerier) *UserController {
	return &UserController{service: svc}
}

/** ====================================================================================
 * 🏁 Get: GetUser
 * =====================================================================================
 * Uri: /api/v1/user/:username
 */
type GetUserRequest struct {
	Username string `uri:"username" binding:"required,min=1"`
}

func (r *GetUserRequest) toSvcParams() service.GetUserInput {
	return service.GetUserInput{
		Username: r.Username,
	}
}
func (c *UserController) GetUser(ctx *gin.Context) {
	var req GetUserRequest
	// 参数错误
	if err := ctx.ShouldBindUri(&req); err != nil {
		response.FailWithBindError(ctx, err)
		return
	}
	params := req.toSvcParams()
	data, err := c.service.GetUser(ctx, params)
	// 自定义错误, 或服务器内部错误
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}

	response.Success(ctx, data)
}

/** ====================================================================================
 * 🏁 Post: CreateUser
 * =====================================================================================
 * Json: /api/v1/user
 */
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=1"`
	Password string `json:"password" binding:"required,min=8,max=16"`
	Email    string `json:"email" binding:"required,email"`
}

func (r *CreateUserRequest) toSvcParams() service.CreateUserInput {
	return service.CreateUserInput{
		Username: r.Username,
		Password: r.Password,
		Email:    r.Email,
	}
}
func (c *UserController) createUser(ctx *gin.Context, role role.Role) {
	var req CreateUserRequest
	// 请求错误
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		response.FailWithBindError(ctx, err)
		return
	}
	// 构造参数
	params := req.toSvcParams()
	// 创建用户
	data, err := c.service.CreateUser(ctx, params, role)
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}

	response.Success(ctx, data)
}

func (c *UserController) CreateNormalUser(ctx *gin.Context) {
	c.createUser(ctx, role.RoleUser)
}

// func (c *UserController) CreateVip(ctx *gin.Context) {
// 	c.createUser(ctx, role.RoleVip)
// }

/** ====================================================================================
 * 🏁 Patch: UpdateUser
 * =====================================================================================
 * Json: /api/v1/user/:username
 */
type UpdateUserRequest struct {
	Username string  `uri:"username" binding:"required,min=1"`
	Password *string `json:"password" binding:"omitempty,min=8,max=16"`
	Email    *string `json:"email" binding:"omitempty,email"`
}

func (r *UpdateUserRequest) toSvcParams() service.UpdateUserInput {
	return service.UpdateUserInput{
		Username: r.Username,
		Password: r.Password,
		Email:    r.Email,
	}
}
func (c *UserController) UpdateUser(ctx *gin.Context) {
	var req UpdateUserRequest
	// 参数错误
	if err := ctx.ShouldBindUri(&req); err != nil { // 解析 Uri
		response.FailWithBindError(ctx, err)
		return
	}
	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil { // 解析 Json
		response.FailWithBindError(ctx, err)
		return
	}
	// 尝试更新信息
	params := req.toSvcParams()
	data, err := c.service.UpdateUser(ctx, params)
	// 自定义错误, 或服务器内部错误
	if err != nil {
		response.FailWithError(ctx, err)
		return
	}

	response.Success(ctx, data)
}
