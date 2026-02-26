package router

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"sercherai/backend/internal/growth/handler"
	"sercherai/backend/internal/growth/repo"
	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/config"
	"sercherai/backend/internal/platform/middleware"
	"sercherai/backend/internal/platform/storage"
)

func Register(r *gin.Engine) {
	cfg := config.Load()

	var growthRepo repo.GrowthRepo
	var redisClient *redis.Client
	db, err := storage.NewMySQL(cfg)
	if err != nil {
		log.Printf("mysql unavailable, fallback to in-memory repo: %v", err)
		growthRepo = repo.NewInMemoryGrowthRepo()
	} else {
		var rErr error
		redisClient, rErr = storage.NewRedis(cfg)
		if rErr != nil {
			log.Printf("redis unavailable, continue without redis cache: %v", rErr)
		}
		growthRepo = repo.NewMySQLGrowthRepo(db, redisClient)
	}

	growthSvc := service.NewGrowthService(growthRepo)
	userGrowthHandler := handler.NewUserGrowthHandler(growthSvc, cfg)
	adminGrowthHandler := handler.NewAdminGrowthHandler(growthSvc, cfg)
	authHandler := handler.NewAuthHandler(
		cfg.JWTSecret,
		cfg.JWTExpireSeconds,
		cfg.JWTRefreshExpireSeconds,
		cfg.LoginFailThreshold,
		cfg.LoginIPFailThreshold,
		cfg.LoginIPPhoneThreshold,
		cfg.LoginLockSeconds,
		cfg.AllowMockLogin,
		db,
		redisClient,
	)

	v1 := r.Group("/api/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", authHandler.Register)
			authGroup.POST("/login", authHandler.Login)
			authGroup.POST("/refresh", authHandler.Refresh)
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.POST("/logout-all", middleware.AuthRequired(cfg.JWTSecret), authHandler.LogoutAll)
			if cfg.AllowMockLogin {
				authGroup.POST("/mock-login", authHandler.MockLogin)
			}
			authGroup.GET("/me", middleware.AuthRequired(cfg.JWTSecret), authHandler.Me)
		}

		adminAuth := v1.Group("/admin/auth")
		adminAuth.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminAuth.GET("/login-logs", authHandler.AdminListLoginLogs)
			adminAuth.GET("/login-logs/export.csv", authHandler.AdminExportLoginLogsCSV)
			adminAuth.GET("/risk-config", authHandler.AdminGetRiskConfig)
			adminAuth.PUT("/risk-config", authHandler.AdminUpdateRiskConfig)
			adminAuth.GET("/risk-config-logs", authHandler.AdminListRiskConfigLogs)
			adminAuth.POST("/unlock", authHandler.AdminUnlockRiskState)
			adminAuth.GET("/unlock-logs", authHandler.AdminListUnlockLogs)
		}

		user := v1.Group("/user")
		user.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			user.GET("/browse-history", userGrowthHandler.ListBrowseHistory)
			user.DELETE("/browse-history/:id", userGrowthHandler.DeleteBrowseHistoryItem)
			user.DELETE("/browse-history", userGrowthHandler.ClearBrowseHistory)

			user.GET("/recharge-records", userGrowthHandler.ListRechargeRecords)

			user.GET("/share-links", userGrowthHandler.ListShareLinks)
			user.POST("/share-links", userGrowthHandler.CreateShareLink)

			user.GET("/share/invites", userGrowthHandler.ListInviteRecords)
			user.GET("/share/rewards", userGrowthHandler.ListRewardRecords)
			user.GET("/reward-wallet", userGrowthHandler.GetRewardWallet)
			user.GET("/reward-wallet/txns", userGrowthHandler.ListRewardWalletTxns)
			user.POST("/reward-wallet/withdraw", userGrowthHandler.CreateWithdrawRequest)
		}

		membership := v1.Group("/membership")
		membership.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			membership.GET("/products", userGrowthHandler.ListMembershipProducts)
			membership.POST("/orders", userGrowthHandler.CreateMembershipOrder)
			membership.GET("/orders", userGrowthHandler.ListMembershipOrders)
			membership.GET("/quota", userGrowthHandler.GetMembershipQuota)
		}

		futures := v1.Group("/futures")
		futures.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			futures.GET("/arbitrage/opportunities", userGrowthHandler.ListArbitrageOpportunities)
			futures.GET("/guidance/:contract", userGrowthHandler.GetFuturesGuidance)
			futures.GET("/strategies", userGrowthHandler.ListFuturesStrategies)
			futures.GET("/strategies/:id", userGrowthHandler.GetFuturesStrategyDetail)
		}

		stocks := v1.Group("/stocks")
		stocks.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			stocks.GET("/recommendations", userGrowthHandler.ListStockRecommendations)
			stocks.GET("/recommendations/:id", userGrowthHandler.GetStockRecommendationDetail)
		}

		news := v1.Group("/news")
		news.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			news.GET("/categories", userGrowthHandler.ListNewsCategories)
			news.GET("/articles", userGrowthHandler.ListNewsArticles)
			news.GET("/articles/:id", userGrowthHandler.GetNewsArticleDetail)
			news.GET("/attachments/:id/signed-url", userGrowthHandler.GetAttachmentSignedURL)
		}
		v1.GET("/news/attachments/:id/download", userGrowthHandler.DownloadAttachment)

		payment := v1.Group("/payment")
		payment.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			payment.POST("/callbacks/:channel", userGrowthHandler.HandlePaymentCallback)
		}

		admin := v1.Group("/admin/growth")
		admin.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			admin.GET("/invite-records", adminGrowthHandler.ListInviteRecords)
			admin.GET("/reward-records", adminGrowthHandler.ListRewardRecords)
			admin.PUT("/reward-records/:id/review", adminGrowthHandler.ReviewRewardRecord)
		}

		adminPayment := v1.Group("/admin/payment")
		adminPayment.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminPayment.GET("/reconciliation", adminGrowthHandler.ListReconciliation)
			adminPayment.POST("/reconciliation/:batch_id/retry", adminGrowthHandler.RetryReconciliation)
		}

		adminRisk := v1.Group("/admin/risk")
		adminRisk.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminRisk.GET("/rules", adminGrowthHandler.ListRiskRules)
			adminRisk.POST("/rules", adminGrowthHandler.CreateRiskRule)
			adminRisk.PUT("/rules/:id", adminGrowthHandler.UpdateRiskRule)
			adminRisk.GET("/hits", adminGrowthHandler.ListRiskHits)
			adminRisk.PUT("/hits/:id/review", adminGrowthHandler.ReviewRiskHit)
		}

		adminRewardWallet := v1.Group("/admin/reward-wallet")
		adminRewardWallet.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminRewardWallet.GET("/withdraw-requests", adminGrowthHandler.ListWithdrawRequests)
			adminRewardWallet.PUT("/withdraw-requests/:id/review", adminGrowthHandler.ReviewWithdrawRequest)
		}

		adminNews := v1.Group("/admin/news")
		adminNews.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminNews.GET("/categories", adminGrowthHandler.ListNewsCategories)
			adminNews.POST("/categories", adminGrowthHandler.CreateNewsCategory)
			adminNews.PUT("/categories/:id", adminGrowthHandler.UpdateNewsCategory)

			adminNews.GET("/articles", adminGrowthHandler.ListNewsArticles)
			adminNews.POST("/articles", adminGrowthHandler.CreateNewsArticle)
			adminNews.PUT("/articles/:id", adminGrowthHandler.UpdateNewsArticle)

			adminNews.GET("/articles/:id/attachments", adminGrowthHandler.ListNewsAttachments)
			adminNews.POST("/articles/:id/attachments", adminGrowthHandler.CreateNewsAttachment)
		}

		adminStocks := v1.Group("/admin/stocks")
		adminStocks.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminStocks.GET("/recommendations", adminGrowthHandler.ListStockRecommendations)
			adminStocks.POST("/recommendations", adminGrowthHandler.CreateStockRecommendation)
			adminStocks.PUT("/recommendations/:id/status", adminGrowthHandler.UpdateStockRecommendationStatus)
			adminStocks.POST("/recommendations/generate-daily", adminGrowthHandler.GenerateDailyStockRecommendations)
		}

		adminFutures := v1.Group("/admin/futures")
		adminFutures.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminFutures.GET("/strategies", adminGrowthHandler.ListFuturesStrategies)
			adminFutures.POST("/strategies", adminGrowthHandler.CreateFuturesStrategy)
			adminFutures.PUT("/strategies/:id/status", adminGrowthHandler.UpdateFuturesStrategyStatus)
		}

		adminUsers := v1.Group("/admin/users")
		adminUsers.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminUsers.GET("", adminGrowthHandler.ListUsers)
			adminUsers.GET("/export.csv", adminGrowthHandler.ExportUsersCSV)
			adminUsers.PUT("/:id/status", adminGrowthHandler.UpdateUserStatus)
			adminUsers.PUT("/:id/member-level", adminGrowthHandler.UpdateUserMemberLevel)
			adminUsers.PUT("/:id/kyc-status", adminGrowthHandler.UpdateUserKYCStatus)
		}

		adminDashboard := v1.Group("/admin/dashboard")
		adminDashboard.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminDashboard.GET("/overview", adminGrowthHandler.DashboardOverview)
		}

		adminAudit := v1.Group("/admin/audit")
		adminAudit.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminAudit.GET("/operation-logs", adminGrowthHandler.ListOperationLogs)
			adminAudit.GET("/operation-logs/export.csv", adminGrowthHandler.ExportOperationLogsCSV)
		}

		adminMembership := v1.Group("/admin/membership")
		adminMembership.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminMembership.GET("/products", adminGrowthHandler.ListMembershipProducts)
			adminMembership.POST("/products", adminGrowthHandler.CreateMembershipProduct)
			adminMembership.PUT("/products/:id/status", adminGrowthHandler.UpdateMembershipProductStatus)

			adminMembership.GET("/orders", adminGrowthHandler.ListMembershipOrders)
			adminMembership.GET("/orders/export.csv", adminGrowthHandler.ExportMembershipOrdersCSV)
			adminMembership.PUT("/orders/:id/status", adminGrowthHandler.UpdateMembershipOrderStatus)

			adminMembership.GET("/quota-configs", adminGrowthHandler.ListVIPQuotaConfigs)
			adminMembership.POST("/quota-configs", adminGrowthHandler.CreateVIPQuotaConfig)
		}

		adminSystem := v1.Group("/admin/system")
		adminSystem.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminSystem.GET("/configs", adminGrowthHandler.ListSystemConfigs)
			adminSystem.PUT("/configs", adminGrowthHandler.UpsertSystemConfig)

			adminSystem.GET("/job-definitions", adminGrowthHandler.ListSchedulerJobDefinitions)
			adminSystem.POST("/job-definitions", adminGrowthHandler.CreateSchedulerJobDefinition)
			adminSystem.PUT("/job-definitions/:id", adminGrowthHandler.UpdateSchedulerJobDefinition)
			adminSystem.PUT("/job-definitions/:id/status", adminGrowthHandler.UpdateSchedulerJobDefinitionStatus)

			adminSystem.GET("/job-runs", adminGrowthHandler.ListSchedulerJobRuns)
			adminSystem.GET("/job-runs/export.csv", adminGrowthHandler.ExportSchedulerJobRunsCSV)
			adminSystem.GET("/job-runs/metrics", adminGrowthHandler.SchedulerJobMetrics)
			adminSystem.POST("/job-runs/trigger", adminGrowthHandler.TriggerSchedulerJob)
			adminSystem.POST("/job-runs/:id/retry", adminGrowthHandler.RetrySchedulerJobRun)
		}

		adminWorkflow := v1.Group("/admin/workflow")
		adminWorkflow.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminWorkflow.GET("/reviews", adminGrowthHandler.ListReviewTasks)
			adminWorkflow.GET("/reviews/export.csv", adminGrowthHandler.ExportReviewTasksCSV)
			adminWorkflow.GET("/metrics", adminGrowthHandler.WorkflowMetrics)
			adminWorkflow.POST("/reviews/submit", adminGrowthHandler.SubmitReviewTask)
			adminWorkflow.PUT("/reviews/:id/assign", adminGrowthHandler.AssignReviewTask)
			adminWorkflow.PUT("/reviews/:id/decision", adminGrowthHandler.ReviewTaskDecision)
			adminWorkflow.GET("/messages", adminGrowthHandler.ListWorkflowMessages)
			adminWorkflow.GET("/messages/export.csv", adminGrowthHandler.ExportWorkflowMessagesCSV)
			adminWorkflow.GET("/messages/unread-count", adminGrowthHandler.CountUnreadWorkflowMessages)
			adminWorkflow.PUT("/messages/:id/read", adminGrowthHandler.UpdateWorkflowMessageRead)
			adminWorkflow.PUT("/messages/read-all", adminGrowthHandler.BulkReadWorkflowMessages)
		}

	}
}
