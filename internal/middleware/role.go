package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	role "github.com/raozhaizhu/go-estate/internal/domain/user"
	appError "github.com/raozhaizhu/go-estate/pkg/app_error"
	"github.com/raozhaizhu/go-estate/pkg/token"
)

func RoleMiddleware(allowedRoles []role.Role) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取荷载
		payload, ok := ctx.Get(PayloadKey)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, appError.ErrAuthRequired)
			return
		}

		// 提取身份
		currRole := payload.(*token.Payload).Role

		// 确认权限
		hasPermission := slices.Contains(allowedRoles, currRole)

		// 没权限, 退出
		if !hasPermission {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, appError.ErrAuthPermissionDenied)
			return
		}

		// 有权限, 放行
		ctx.Next()
	}
}
