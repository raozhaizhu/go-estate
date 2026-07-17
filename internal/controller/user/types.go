package user

import (
	"context"

	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	"github.com/raozhaizhu/go-estate/internal/service/user"
	service "github.com/raozhaizhu/go-estate/internal/service/user"
)

/** ====================================================================================
 * 🏁 UserController
 * =====================================================================================
 *
 */

type Controller struct {
	service Service
}

type Service interface {
	CreateUser(ctx context.Context, p user.CreateUserInput, role role.Role) (*user.DTO, error)
	GetUser(ctx context.Context, p user.GetUserInput) (*user.DTO, error)
	UpdateUser(ctx context.Context, p user.UpdateUserInput) (*user.DTO, error)
}

func New(svc Service) *Controller {
	return &Controller{service: svc}
}

/** ====================================================================================
 * 🏁 Get: GetUser
 * =====================================================================================
 */

type GetUserRequest struct {
	Username string `uri:"username" binding:"required,min=3,max=32" example:"Bob"`
}

func (r *GetUserRequest) toSvcInput() service.GetUserInput {
	return service.GetUserInput{
		Username: r.Username,
	}
}

/** ====================================================================================
 * 🏁 Post: CreateUser
 * =====================================================================================
 */

type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32" example:"Bob"`
	Password string `json:"password" binding:"required,min=8,max=16" example:"12345678"`
	Email    string `json:"email" binding:"required,email" example:"Bob@test.com"`
}

func (r *CreateUserRequest) toSvcInput() service.CreateUserInput {
	return service.CreateUserInput{
		Username: r.Username,
		Password: r.Password,
		Email:    r.Email,
	}
}

/** ====================================================================================
 * 🏁 Patch: UpdateUser
 * =====================================================================================
 */

type UpdateUserRequest struct {
	Username string  `uri:"username" binding:"required,min=3,max=32" example:"Bob"`
	Password *string `json:"password" binding:"omitempty,min=8,max=16" example:"12345678"`
	Email    *string `json:"email" binding:"omitempty,email" example:"Bob@test.com"`
}

func (r *UpdateUserRequest) toSvcInput() service.UpdateUserInput {
	return service.UpdateUserInput{
		Username: r.Username,
		Password: r.Password,
		Email:    r.Email,
	}
}
