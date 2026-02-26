package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/platform/auth"
)

func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 40101, "message": "missing authorization", "data": struct{}{}})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 40102, "message": "invalid authorization format", "data": struct{}{}})
			return
		}

		claims, err := auth.ParseToken(secret, parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 40103, "message": "invalid token", "data": struct{}{}})
			return
		}
		if strings.TrimSpace(claims.UserID) == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 40103, "message": "invalid token", "data": struct{}{}})
			return
		}
		if strings.ToUpper(strings.TrimSpace(claims.TokenType)) != "ACCESS" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 40105, "message": "invalid token type", "data": struct{}{}})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func RoleRequired(roles ...string) gin.HandlerFunc {
	allowed := map[string]struct{}{}
	for _, role := range roles {
		allowed[strings.ToUpper(strings.TrimSpace(role))] = struct{}{}
	}

	return func(c *gin.Context) {
		roleVal, ok := c.Get("role")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40301, "message": "role missing", "data": struct{}{}})
			return
		}
		role, _ := roleVal.(string)
		if _, exists := allowed[strings.ToUpper(strings.TrimSpace(role))]; !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40302, "message": "insufficient permission", "data": struct{}{}})
			return
		}
		c.Next()
	}
}
