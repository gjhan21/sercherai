package middleware

import (
	"database/sql"
	"fmt"
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

func PermissionRequired(db *sql.DB, permissionCodes ...string) gin.HandlerFunc {
	requiredCodes := normalizePermissionCodes(permissionCodes)
	return func(c *gin.Context) {
		if len(requiredCodes) == 0 || db == nil {
			c.Next()
			return
		}
		userIDValue, ok := c.Get("user_id")
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40301, "message": "role missing", "data": struct{}{}})
			return
		}
		userID, castOK := userIDValue.(string)
		if !castOK || strings.TrimSpace(userID) == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40301, "message": "role missing", "data": struct{}{}})
			return
		}
		allowed, err := checkPermission(db, userID, requiredCodes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"code": 50001, "message": err.Error(), "data": struct{}{}})
			return
		}
		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 40302, "message": "insufficient permission", "data": struct{}{}})
			return
		}
		c.Next()
	}
}

func checkPermission(db *sql.DB, userID string, permissionCodes []string) (bool, error) {
	var superAdminCount int
	err := db.QueryRow(`
SELECT COUNT(*)
FROM rbac_user_roles ur
JOIN rbac_roles r ON r.id = ur.role_id
WHERE ur.user_id = ? AND r.status = 'ACTIVE' AND r.role_key = 'SUPER_ADMIN'`, userID).Scan(&superAdminCount)
	if err != nil {
		return false, err
	}
	if superAdminCount > 0 {
		return true, nil
	}

	placeholders := make([]string, 0, len(permissionCodes))
	args := make([]interface{}, 0, len(permissionCodes)+1)
	args = append(args, userID)
	for _, code := range permissionCodes {
		placeholders = append(placeholders, "?")
		args = append(args, code)
	}

	query := `
SELECT COUNT(*)
FROM rbac_user_roles ur
JOIN rbac_roles r ON r.id = ur.role_id
JOIN rbac_role_permissions rp ON rp.role_id = r.id
JOIN rbac_permissions p ON p.code = rp.permission_code
WHERE ur.user_id = ?
  AND r.status = 'ACTIVE'
  AND p.status = 'ACTIVE'
  AND p.code IN (` + strings.Join(placeholders, ",") + `)`
	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func normalizePermissionCodes(codes []string) []string {
	if len(codes) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(codes))
	for _, code := range codes {
		normalized := strings.TrimSpace(code)
		if normalized == "" {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func BuildPermissionDeniedMessage(permissionCodes []string) string {
	return fmt.Sprintf("missing permission: %s", strings.Join(permissionCodes, ","))
}
