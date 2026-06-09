package router

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/handler"
	"sercherai/backend/internal/platform/config"
	"sercherai/backend/internal/platform/middleware"
)

func registerAuthRoutes(v1 *gin.RouterGroup, authHandler *handler.AuthHandler, cfg *config.Config, db *sql.DB) {
	authGroup := v1.Group("/auth")
	{
		authGroup.POST("/register", authHandler.Register)
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/forgot-password", authHandler.ForgotPassword)
		authGroup.POST("/reset-password", authHandler.ResetPassword)
		authGroup.POST("/refresh", authHandler.Refresh)
		authGroup.POST("/logout", authHandler.Logout)
		authGroup.POST("/logout-all", middleware.AuthRequired(cfg.JWTSecret), authHandler.LogoutAll)
		authGroup.POST("/change-password", middleware.AuthRequired(cfg.JWTSecret), authHandler.ChangePassword)
		if cfg.AllowMockLogin {
			authGroup.POST("/mock-login", authHandler.MockLogin)
		}
		authGroup.GET("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Me)
	}

	adminAuth := v1.Group("/admin/auth")
	adminAuth.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminAuth.GET("/login-logs", middleware.PermissionRequired(db, "auth_security.view"), authHandler.AdminListLoginLogs)
		adminAuth.GET("/login-logs/export.csv", middleware.PermissionRequired(db, "auth_security.view"), authHandler.AdminExportLoginLogsCSV)
		adminAuth.GET("/risk-config", middleware.PermissionRequired(db, "auth_security.view"), authHandler.AdminGetRiskConfig)
		adminAuth.PUT("/risk-config", middleware.PermissionRequired(db, "auth_security.edit"), authHandler.AdminUpdateRiskConfig)
		adminAuth.GET("/risk-config-logs", middleware.PermissionRequired(db, "auth_security.view"), authHandler.AdminListRiskConfigLogs)
		adminAuth.POST("/unlock", middleware.PermissionRequired(db, "auth_security.edit"), authHandler.AdminUnlockRiskState)
		adminAuth.GET("/unlock-logs", middleware.PermissionRequired(db, "auth_security.view"), authHandler.AdminListUnlockLogs)
	}

	adminAccess := v1.Group("/admin/access")
	adminAccess.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminAccess.GET("/me", authHandler.AdminGetAccessProfile)
		adminAccess.GET("/permissions", middleware.PermissionRequired(db, "access.view"), authHandler.AdminListPermissions)
		adminAccess.GET("/roles", middleware.PermissionRequired(db, "access.view"), authHandler.AdminListRoles)
		adminAccess.POST("/roles", middleware.PermissionRequired(db, "access.edit"), authHandler.AdminCreateRole)
		adminAccess.PUT("/roles/:id", middleware.PermissionRequired(db, "access.edit"), authHandler.AdminUpdateRole)
		adminAccess.PUT("/roles/:id/status", middleware.PermissionRequired(db, "access.edit"), authHandler.AdminUpdateRoleStatus)

		adminAccess.GET("/admin-users", middleware.PermissionRequired(db, "access.view"), authHandler.AdminListAdminUsers)
		adminAccess.POST("/admin-users", middleware.PermissionRequired(db, "access.edit"), authHandler.AdminCreateAdminUser)
		adminAccess.PUT("/admin-users/:id/status", middleware.PermissionRequired(db, "access.edit"), authHandler.AdminUpdateAdminUserStatus)
		adminAccess.PUT("/admin-users/:id/roles", middleware.PermissionRequired(db, "access.edit"), authHandler.AdminAssignAdminUserRoles)
		adminAccess.PUT("/admin-users/:id/password", middleware.PermissionRequired(db, "access.edit"), authHandler.AdminResetAdminUserPassword)
	}
}
