package middleware

import (
	"net/http"
	"strings"

	"delugewarning/internal/auth"

	"github.com/gin-gonic/gin"
)

const (
	ctxUserID = "user_id"
	ctxRole   = "role"
)

// Auth 校验 JWT 并把用户信息写入 context。
func Auth(mgr *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token := strings.TrimPrefix(header, "Bearer ")
		if token == "" || token == header {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "缺少或非法的 Authorization 头"})
			return
		}
		claims, err := mgr.Parse(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token 校验失败"})
			return
		}
		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxRole, claims.Role)
		c.Next()
	}
}

// RequireRole 限定只有指定角色可访问。
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if RoleOf(c) != role {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "无权访问"})
			return
		}
		c.Next()
	}
}

// UserIDOf 从 context 取当前用户ID。
func UserIDOf(c *gin.Context) int64 {
	if v, ok := c.Get(ctxUserID); ok {
		if id, ok := v.(int64); ok {
			return id
		}
	}
	return 0
}

// RoleOf 从 context 取当前角色。
func RoleOf(c *gin.Context) string {
	if v, ok := c.Get(ctxRole); ok {
		if r, ok := v.(string); ok {
			return r
		}
	}
	return ""
}
