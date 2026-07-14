package middleware

import (
	"slices"

	"github.com/gin-gonic/gin"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	response "github.com/raozhaizhu/go-estate/pkg/api"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

// RequireRoles 角色权限认证中间件
// 校验用户是否在权限组内
func RequireRoles(allowedRoles []role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取荷载
		payload, ok := ctx.Get(token.PayloadKey)
		if !ok {
			response.Fail(ctx, appError.ErrAuthRequired)
			ctx.Abort()
			return
		}

		// 提取身份
		currRole := payload.(*token.Payload).Role

		// 确认权限
		hasPermission := slices.Contains(allowedRoles, currRole)

		// 没权限, 退出
		if !hasPermission {
			response.Fail(ctx, appError.ErrAuthPermissionDenied)
			ctx.Abort()
			return
		}

		// 有权限, 放行
		ctx.Next()
	}
}
