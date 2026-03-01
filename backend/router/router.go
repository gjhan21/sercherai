package router

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"sercherai/backend/internal/growth/handler"
	"sercherai/backend/internal/growth/model"
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

	if db != nil {
		startDocFastIncrementalSyncWorker(growthSvc)
		startTushareNewsIncrementalSyncWorker(growthSvc)
		startVIPMembershipLifecycleWorker(growthSvc)
	}

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

		user := v1.Group("/user")
		user.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			user.GET("/profile", userGrowthHandler.GetUserProfile)
			user.PUT("/profile", userGrowthHandler.UpdateUserProfile)
			user.GET("/kyc/status", userGrowthHandler.GetKYCStatus)
			user.POST("/kyc/submit", userGrowthHandler.SubmitKYC)

			user.GET("/browse-history", userGrowthHandler.ListBrowseHistory)
			user.DELETE("/browse-history/:id", userGrowthHandler.DeleteBrowseHistoryItem)
			user.DELETE("/browse-history", userGrowthHandler.ClearBrowseHistory)

			user.GET("/recharge-records", userGrowthHandler.ListRechargeRecords)

			user.GET("/share-links", userGrowthHandler.ListShareLinks)
			user.POST("/share-links", userGrowthHandler.CreateShareLink)

			user.GET("/share/invites", userGrowthHandler.ListInviteRecords)
			user.GET("/share/invite-summary", userGrowthHandler.GetInviteSummary)
			user.GET("/share/rewards", userGrowthHandler.ListRewardRecords)
			user.GET("/reward-wallet", userGrowthHandler.GetRewardWallet)
			user.GET("/reward-wallet/txns", userGrowthHandler.ListRewardWalletTxns)
			user.POST("/reward-wallet/withdraw", userGrowthHandler.CreateWithdrawRequest)
		}

		subscriptions := v1.Group("/subscriptions")
		subscriptions.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			subscriptions.GET("", userGrowthHandler.ListSubscriptions)
			subscriptions.POST("", userGrowthHandler.CreateSubscription)
			subscriptions.PUT("/:id", userGrowthHandler.UpdateSubscription)
		}

		messages := v1.Group("/messages")
		messages.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			messages.GET("", userGrowthHandler.ListMessages)
			messages.PUT("/:id/read", userGrowthHandler.ReadMessage)
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
			futures.GET("/arbitrage", userGrowthHandler.ListFuturesArbitrage)
			futures.GET("/arbitrage/:id", userGrowthHandler.GetFuturesArbitrageDetail)
			futures.GET("/arbitrage/opportunities", userGrowthHandler.ListArbitrageOpportunities)
			futures.GET("/guidance/:contract", userGrowthHandler.GetFuturesGuidance)
			futures.POST("/alerts", userGrowthHandler.CreateFuturesAlert)
			futures.GET("/reviews", userGrowthHandler.ListFuturesReviews)
			futures.GET("/strategies", userGrowthHandler.ListFuturesStrategies)
			futures.GET("/strategies/:id", userGrowthHandler.GetFuturesStrategyDetail)
			futures.GET("/strategies/:id/insight", userGrowthHandler.GetFuturesStrategyInsight)
		}

		stocks := v1.Group("/stocks")
		stocks.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			stocks.GET("/recommendations", userGrowthHandler.ListStockRecommendations)
			stocks.GET("/recommendations/:id", userGrowthHandler.GetStockRecommendationDetail)
			stocks.GET("/recommendations/:id/performance", userGrowthHandler.GetStockRecommendationPerformance)
			stocks.GET("/recommendations/:id/insight", userGrowthHandler.GetStockRecommendationInsight)
		}

		news := v1.Group("/news")
		news.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			news.GET("/categories", userGrowthHandler.ListNewsCategories)
			news.GET("/articles", userGrowthHandler.ListNewsArticles)
			news.GET("/articles/:id", userGrowthHandler.GetNewsArticleDetail)
			news.GET("/articles/:id/attachments", userGrowthHandler.ListNewsAttachments)
			news.GET("/attachments/:id/signed-url", userGrowthHandler.GetAttachmentSignedURL)
		}
		v1.GET("/news/attachments/:id/download", userGrowthHandler.DownloadAttachment)

		public := v1.Group("/public")
		{
			public.GET("/holdings", userGrowthHandler.ListPublicHoldings)
			public.GET("/futures-positions", userGrowthHandler.ListPublicFuturesPositions)
			public.GET("/news/categories", userGrowthHandler.ListNewsCategories)
			public.GET("/news/articles", userGrowthHandler.ListNewsArticles)
			public.GET("/news/articles/:id", userGrowthHandler.GetNewsArticleDetail)
			public.GET("/news/articles/:id/attachments", userGrowthHandler.ListNewsAttachments)
		}

		market := v1.Group("/market")
		market.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			market.GET("/events", userGrowthHandler.ListMarketEvents)
			market.GET("/events/:id", userGrowthHandler.GetMarketEventDetail)
		}

		payment := v1.Group("/payment")
		payment.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
		{
			payment.POST("/callbacks/:channel", userGrowthHandler.HandlePaymentCallback)
		}
		v1.Any("/payment/callbacks/yolkpay/notify", userGrowthHandler.HandleYolkPayCallback)

		admin := v1.Group("/admin/growth")
		admin.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			admin.GET("/invite-records", middleware.PermissionRequired(db, "growth.view"), adminGrowthHandler.ListInviteRecords)
			admin.GET("/reward-records", middleware.PermissionRequired(db, "growth.view"), adminGrowthHandler.ListRewardRecords)
			admin.PUT("/reward-records/:id/review", middleware.PermissionRequired(db, "growth.edit"), adminGrowthHandler.ReviewRewardRecord)
		}

		adminPayment := v1.Group("/admin/payment")
		adminPayment.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminPayment.GET("/reconciliation", middleware.PermissionRequired(db, "payment.view"), adminGrowthHandler.ListReconciliation)
			adminPayment.POST("/reconciliation/:batch_id/retry", middleware.PermissionRequired(db, "payment.edit"), adminGrowthHandler.RetryReconciliation)
		}

		adminRisk := v1.Group("/admin/risk")
		adminRisk.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminRisk.GET("/rules", middleware.PermissionRequired(db, "risk.view"), adminGrowthHandler.ListRiskRules)
			adminRisk.POST("/rules", middleware.PermissionRequired(db, "risk.edit"), adminGrowthHandler.CreateRiskRule)
			adminRisk.PUT("/rules/:id", middleware.PermissionRequired(db, "risk.edit"), adminGrowthHandler.UpdateRiskRule)
			adminRisk.GET("/hits", middleware.PermissionRequired(db, "risk.view"), adminGrowthHandler.ListRiskHits)
			adminRisk.PUT("/hits/:id/review", middleware.PermissionRequired(db, "risk.edit"), adminGrowthHandler.ReviewRiskHit)
		}

		adminRewardWallet := v1.Group("/admin/reward-wallet")
		adminRewardWallet.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminRewardWallet.GET("/withdraw-requests", middleware.PermissionRequired(db, "reward_wallet.view"), adminGrowthHandler.ListWithdrawRequests)
			adminRewardWallet.PUT("/withdraw-requests/:id/review", middleware.PermissionRequired(db, "reward_wallet.edit"), adminGrowthHandler.ReviewWithdrawRequest)
		}

		adminNews := v1.Group("/admin/news")
		adminNews.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminNews.GET("/categories", middleware.PermissionRequired(db, "news.view"), adminGrowthHandler.ListNewsCategories)
			adminNews.POST("/categories", middleware.PermissionRequired(db, "news.edit"), adminGrowthHandler.CreateNewsCategory)
			adminNews.PUT("/categories/:id", middleware.PermissionRequired(db, "news.edit"), adminGrowthHandler.UpdateNewsCategory)

			adminNews.GET("/articles", middleware.PermissionRequired(db, "news.view"), adminGrowthHandler.ListNewsArticles)
			adminNews.GET("/articles/:id", middleware.PermissionRequired(db, "news.view"), adminGrowthHandler.GetNewsArticleDetail)
			adminNews.POST("/articles", middleware.PermissionRequired(db, "news.edit"), adminGrowthHandler.CreateNewsArticle)
			adminNews.PUT("/articles/:id", middleware.PermissionRequired(db, "news.edit"), adminGrowthHandler.UpdateNewsArticle)
			adminNews.PUT("/articles/:id/publish", middleware.PermissionRequired(db, "news.edit"), adminGrowthHandler.PublishNewsArticle)

			adminNews.POST("/attachments/upload", middleware.PermissionRequired(db, "news.edit"), adminGrowthHandler.UploadNewsAttachment)
			adminNews.GET("/articles/:id/attachments", middleware.PermissionRequired(db, "news.view"), adminGrowthHandler.ListNewsAttachments)
			adminNews.POST("/articles/:id/attachments", middleware.PermissionRequired(db, "news.edit"), adminGrowthHandler.CreateNewsAttachment)
			adminNews.DELETE("/attachments/:id", middleware.PermissionRequired(db, "news.edit"), adminGrowthHandler.DeleteNewsAttachment)
		}

		adminDataSources := v1.Group("/admin/data-sources")
		adminDataSources.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminDataSources.GET("", middleware.PermissionRequired(db, "data_source.view"), adminGrowthHandler.ListDataSources)
			adminDataSources.POST("/health-checks", middleware.PermissionRequired(db, "data_source.edit"), adminGrowthHandler.BatchCheckDataSourcesHealth)
			adminDataSources.POST("", middleware.PermissionRequired(db, "data_source.edit"), adminGrowthHandler.CreateDataSource)
			adminDataSources.PUT("/:source_key", middleware.PermissionRequired(db, "data_source.edit"), adminGrowthHandler.UpdateDataSource)
			adminDataSources.DELETE("/:source_key", middleware.PermissionRequired(db, "data_source.edit"), adminGrowthHandler.DeleteDataSource)
			adminDataSources.POST("/:source_key/health-check", middleware.PermissionRequired(db, "data_source.edit"), adminGrowthHandler.CheckDataSourceHealth)
			adminDataSources.GET("/:source_key/health-logs", middleware.PermissionRequired(db, "data_source.view"), adminGrowthHandler.ListDataSourceHealthLogs)
		}

		adminStocks := v1.Group("/admin/stocks")
		adminStocks.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminStocks.GET("/recommendations", middleware.PermissionRequired(db, "market.view"), adminGrowthHandler.ListStockRecommendations)
			adminStocks.POST("/recommendations", middleware.PermissionRequired(db, "market.edit"), adminGrowthHandler.CreateStockRecommendation)
			adminStocks.PUT("/recommendations/:id/status", middleware.PermissionRequired(db, "market.edit"), adminGrowthHandler.UpdateStockRecommendationStatus)
			adminStocks.POST("/quotes/sync", middleware.PermissionRequired(db, "market.edit"), adminGrowthHandler.SyncStockQuotes)
			adminStocks.GET("/quant/top", middleware.PermissionRequired(db, "market.view"), adminGrowthHandler.ListQuantTopStocks)
			adminStocks.GET("/quant/evaluation", middleware.PermissionRequired(db, "market.view"), adminGrowthHandler.ListQuantEvaluation)
			adminStocks.GET("/quant/evaluation/export.csv", middleware.PermissionRequired(db, "market.view"), adminGrowthHandler.ExportQuantEvaluationCSV)
			adminStocks.POST("/recommendations/generate-daily", middleware.PermissionRequired(db, "market.edit"), adminGrowthHandler.GenerateDailyStockRecommendations)
		}

		adminFutures := v1.Group("/admin/futures")
		adminFutures.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminFutures.GET("/strategies", middleware.PermissionRequired(db, "market.view"), adminGrowthHandler.ListFuturesStrategies)
			adminFutures.POST("/strategies", middleware.PermissionRequired(db, "market.edit"), adminGrowthHandler.CreateFuturesStrategy)
			adminFutures.PUT("/strategies/:id/status", middleware.PermissionRequired(db, "market.edit"), adminGrowthHandler.UpdateFuturesStrategyStatus)
		}

		adminMarket := v1.Group("/admin/market")
		adminMarket.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminMarket.GET("/events", middleware.PermissionRequired(db, "market.view"), adminGrowthHandler.ListMarketEvents)
			adminMarket.POST("/events", middleware.PermissionRequired(db, "market.edit"), adminGrowthHandler.CreateMarketEvent)
			adminMarket.PUT("/events/:id", middleware.PermissionRequired(db, "market.edit"), adminGrowthHandler.UpdateMarketEvent)
		}

		adminUsers := v1.Group("/admin/users")
		adminUsers.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminUsers.GET("", middleware.PermissionRequired(db, "users.view"), adminGrowthHandler.ListUsers)
			adminUsers.GET("/source-summary", middleware.PermissionRequired(db, "users.view"), adminGrowthHandler.UserSourceSummary)
			adminUsers.GET("/export.csv", middleware.PermissionRequired(db, "users.view"), adminGrowthHandler.ExportUsersCSV)
			adminUsers.GET("/browse-histories", middleware.PermissionRequired(db, "users.view"), adminGrowthHandler.ListBrowseHistories)
			adminUsers.GET("/browse-histories/summary", middleware.PermissionRequired(db, "users.view"), adminGrowthHandler.BrowseHistorySummary)
			adminUsers.GET("/browse-histories/trend", middleware.PermissionRequired(db, "users.view"), adminGrowthHandler.BrowseHistoryTrend)
			adminUsers.GET("/browse-histories/segments", middleware.PermissionRequired(db, "users.view"), adminGrowthHandler.ListBrowseUserSegments)
			adminUsers.GET("/browse-histories/export.csv", middleware.PermissionRequired(db, "users.view"), adminGrowthHandler.ExportBrowseHistoriesCSV)
			adminUsers.GET("/messages", middleware.PermissionRequired(db, "users.view"), adminGrowthHandler.ListUserMessages)
			adminUsers.POST("/messages", middleware.PermissionRequired(db, "users.edit"), adminGrowthHandler.CreateUserMessages)
			adminUsers.GET("/:id/center-overview", middleware.PermissionRequired(db, "users.view"), adminGrowthHandler.GetUserCenterOverview)
			adminUsers.PUT("/:id/subscriptions/:sub_id", middleware.PermissionRequired(db, "users.edit"), adminGrowthHandler.UpdateUserSubscription)
			adminUsers.PUT("/:id/status", middleware.PermissionRequired(db, "users.edit"), adminGrowthHandler.UpdateUserStatus)
			adminUsers.PUT("/:id/member-level", middleware.PermissionRequired(db, "users.edit"), adminGrowthHandler.UpdateUserMemberLevel)
			adminUsers.PUT("/:id/kyc-status", middleware.PermissionRequired(db, "users.edit"), adminGrowthHandler.UpdateUserKYCStatus)
		}

		adminDashboard := v1.Group("/admin/dashboard")
		adminDashboard.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminDashboard.GET("/overview", middleware.PermissionRequired(db, "dashboard.view"), adminGrowthHandler.DashboardOverview)
		}

		adminAudit := v1.Group("/admin/audit")
		adminAudit.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminAudit.GET("/operation-logs", middleware.PermissionRequired(db, "audit.view"), adminGrowthHandler.ListOperationLogs)
			adminAudit.GET("/operation-logs/export.csv", middleware.PermissionRequired(db, "audit.view"), adminGrowthHandler.ExportOperationLogsCSV)
		}

		adminMembership := v1.Group("/admin/membership")
		adminMembership.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminMembership.GET("/products", middleware.PermissionRequired(db, "membership.view"), adminGrowthHandler.ListMembershipProducts)
			adminMembership.POST("/products", middleware.PermissionRequired(db, "membership.edit"), adminGrowthHandler.CreateMembershipProduct)
			adminMembership.PUT("/products/:id", middleware.PermissionRequired(db, "membership.edit"), adminGrowthHandler.UpdateMembershipProduct)
			adminMembership.PUT("/products/:id/status", middleware.PermissionRequired(db, "membership.edit"), adminGrowthHandler.UpdateMembershipProductStatus)

			adminMembership.GET("/orders", middleware.PermissionRequired(db, "membership.view"), adminGrowthHandler.ListMembershipOrders)
			adminMembership.GET("/orders/export.csv", middleware.PermissionRequired(db, "membership.view"), adminGrowthHandler.ExportMembershipOrdersCSV)
			adminMembership.PUT("/orders/:id/status", middleware.PermissionRequired(db, "membership.edit"), adminGrowthHandler.UpdateMembershipOrderStatus)

			adminMembership.GET("/quota-configs", middleware.PermissionRequired(db, "membership.view"), adminGrowthHandler.ListVIPQuotaConfigs)
			adminMembership.POST("/quota-configs", middleware.PermissionRequired(db, "membership.edit"), adminGrowthHandler.CreateVIPQuotaConfig)
			adminMembership.PUT("/quota-configs/:id", middleware.PermissionRequired(db, "membership.edit"), adminGrowthHandler.UpdateVIPQuotaConfig)
			adminMembership.GET("/user-quotas", middleware.PermissionRequired(db, "membership.view"), adminGrowthHandler.ListUserQuotas)
			adminMembership.PUT("/user-quotas/:user_id/adjust", middleware.PermissionRequired(db, "membership.edit"), adminGrowthHandler.AdjustUserQuota)
		}

		adminSystem := v1.Group("/admin/system")
		adminSystem.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminSystem.GET("/configs", middleware.PermissionRequired(db, "system_config.view"), adminGrowthHandler.ListSystemConfigs)
			adminSystem.PUT("/configs", middleware.PermissionRequired(db, "system_config.edit"), adminGrowthHandler.UpsertSystemConfig)
			adminSystem.POST("/configs/oss/qiniu/test", middleware.PermissionRequired(db, "system_config.edit"), adminGrowthHandler.TestOSSQiniuConfig)
			adminSystem.POST("/configs/payment/yolkpay/test", middleware.PermissionRequired(db, "system_config.edit"), adminGrowthHandler.TestYolkPayConfig)

			adminSystem.GET("/job-definitions", middleware.PermissionRequired(db, "system_job.view"), adminGrowthHandler.ListSchedulerJobDefinitions)
			adminSystem.GET("/job-definitions/supported", middleware.PermissionRequired(db, "system_job.view"), adminGrowthHandler.ListSupportedSchedulerJobs)
			adminSystem.POST("/job-definitions", middleware.PermissionRequired(db, "system_job.edit"), adminGrowthHandler.CreateSchedulerJobDefinition)
			adminSystem.PUT("/job-definitions/:id", middleware.PermissionRequired(db, "system_job.edit"), adminGrowthHandler.UpdateSchedulerJobDefinition)
			adminSystem.PUT("/job-definitions/:id/status", middleware.PermissionRequired(db, "system_job.edit"), adminGrowthHandler.UpdateSchedulerJobDefinitionStatus)
			adminSystem.DELETE("/job-definitions/:id", middleware.PermissionRequired(db, "system_job.edit"), adminGrowthHandler.DeleteSchedulerJobDefinition)

			adminSystem.GET("/job-runs", middleware.PermissionRequired(db, "system_job.view"), adminGrowthHandler.ListSchedulerJobRuns)
			adminSystem.GET("/job-runs/:id/news-sync-details", middleware.PermissionRequired(db, "system_job.view"), adminGrowthHandler.ListNewsSyncRunDetails)
			adminSystem.GET("/job-runs/export.csv", middleware.PermissionRequired(db, "system_job.view"), adminGrowthHandler.ExportSchedulerJobRunsCSV)
			adminSystem.GET("/job-runs/metrics", middleware.PermissionRequired(db, "system_job.view"), adminGrowthHandler.SchedulerJobMetrics)
			adminSystem.POST("/job-runs/trigger", middleware.PermissionRequired(db, "system_job.edit"), adminGrowthHandler.TriggerSchedulerJob)
			adminSystem.POST("/job-runs/:id/retry", middleware.PermissionRequired(db, "system_job.edit"), adminGrowthHandler.RetrySchedulerJobRun)
			adminSystem.POST("/job-runs/:id/retry-news-sync-item", middleware.PermissionRequired(db, "system_job.edit"), adminGrowthHandler.RetryNewsSyncItem)
		}

		adminWorkflow := v1.Group("/admin/workflow")
		adminWorkflow.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
		{
			adminWorkflow.GET("/reviews", middleware.PermissionRequired(db, "review.view"), adminGrowthHandler.ListReviewTasks)
			adminWorkflow.GET("/reviews/export.csv", middleware.PermissionRequired(db, "review.view"), adminGrowthHandler.ExportReviewTasksCSV)
			adminWorkflow.GET("/metrics", middleware.PermissionRequired(db, "review.view"), adminGrowthHandler.WorkflowMetrics)
			adminWorkflow.POST("/reviews/submit", middleware.PermissionRequired(db, "review.edit"), adminGrowthHandler.SubmitReviewTask)
			adminWorkflow.PUT("/reviews/:id/assign", middleware.PermissionRequired(db, "review.edit"), adminGrowthHandler.AssignReviewTask)
			adminWorkflow.PUT("/reviews/:id/decision", middleware.PermissionRequired(db, "review.edit"), adminGrowthHandler.ReviewTaskDecision)
			adminWorkflow.GET("/messages", middleware.PermissionRequired(db, "workflow.view"), adminGrowthHandler.ListWorkflowMessages)
			adminWorkflow.GET("/messages/export.csv", middleware.PermissionRequired(db, "workflow.view"), adminGrowthHandler.ExportWorkflowMessagesCSV)
			adminWorkflow.GET("/messages/unread-count", middleware.PermissionRequired(db, "workflow.view"), adminGrowthHandler.CountUnreadWorkflowMessages)
			adminWorkflow.PUT("/messages/:id/read", middleware.PermissionRequired(db, "workflow.edit"), adminGrowthHandler.UpdateWorkflowMessageRead)
			adminWorkflow.PUT("/messages/read-all", middleware.PermissionRequired(db, "workflow.edit"), adminGrowthHandler.BulkReadWorkflowMessages)
		}

	}
}

const (
	docFastIncrementalJobName            = "doc_fast_news_incremental"
	docFastIncrementalDefaultMinutes     = 100
	docFastIncrementalMaxMinutes         = 24 * 60
	tushareNewsIncrementalJobName        = "tushare_news_incremental"
	tushareNewsIncrementalDefaultMinutes = 20
	tushareNewsIncrementalMaxMinutes     = 24 * 60
	vipLifecycleJobName                  = "vip_membership_lifecycle"
	vipLifecycleDefaultMinutes           = 30
	vipLifecycleMaxMinutes               = 24 * 60
)

func startDocFastIncrementalSyncWorker(growthSvc service.GrowthService) {
	go func() {
		log.Printf("[scheduler] start doc_fast incremental worker")
		for {
			enabled, intervalMinutes := loadDocFastIncrementalWorkerConfig(growthSvc)
			if enabled {
				runDocFastIncrementalJob(growthSvc, "SYSTEM_TIMER")
			}
			if intervalMinutes <= 0 {
				intervalMinutes = docFastIncrementalDefaultMinutes
			}
			time.Sleep(time.Duration(intervalMinutes) * time.Minute)
		}
	}()
}

func runDocFastIncrementalJob(growthSvc service.GrowthService, triggerSource string) {
	summary, runErr := growthSvc.AdminSyncDocFastNewsIncremental(0)
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	_, logErr := growthSvc.AdminCreateSchedulerJobRun(
		docFastIncrementalJobName,
		triggerSource,
		status,
		summary,
		errorMessage,
		"system",
	)
	if logErr != nil {
		log.Printf("[scheduler] create job run failed(%s): %v", docFastIncrementalJobName, logErr)
	}
	if runErr != nil {
		log.Printf("[scheduler] job failed(%s): %v", docFastIncrementalJobName, runErr)
		return
	}
	log.Printf("[scheduler] job success(%s): %s", docFastIncrementalJobName, strings.TrimSpace(summary))
}

func startTushareNewsIncrementalSyncWorker(growthSvc service.GrowthService) {
	go func() {
		log.Printf("[scheduler] start tushare news incremental worker")
		for {
			enabled, intervalMinutes := loadTushareNewsIncrementalWorkerConfig(growthSvc)
			if enabled {
				runTushareNewsIncrementalJob(growthSvc, "SYSTEM_TIMER")
			}
			if intervalMinutes <= 0 {
				intervalMinutes = tushareNewsIncrementalDefaultMinutes
			}
			time.Sleep(time.Duration(intervalMinutes) * time.Minute)
		}
	}()
}

func runTushareNewsIncrementalJob(growthSvc service.GrowthService, triggerSource string) {
	summary, details, runErr := growthSvc.AdminSyncTushareNewsIncrementalWithOptions(model.TushareNewsSyncOptions{})
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	runID, logErr := growthSvc.AdminCreateSchedulerJobRun(
		tushareNewsIncrementalJobName,
		triggerSource,
		status,
		summary,
		errorMessage,
		"system",
	)
	if logErr != nil {
		log.Printf("[scheduler] create job run failed(%s): %v", tushareNewsIncrementalJobName, logErr)
	} else if len(details) > 0 {
		if detailErr := growthSvc.AdminCreateNewsSyncRunDetails(runID, details); detailErr != nil {
			log.Printf("[scheduler] create news sync details failed(%s): %v", runID, detailErr)
		}
	}
	if runErr != nil {
		log.Printf("[scheduler] job failed(%s): %v", tushareNewsIncrementalJobName, runErr)
		return
	}
	log.Printf("[scheduler] job success(%s): %s", tushareNewsIncrementalJobName, strings.TrimSpace(summary))
}

func startVIPMembershipLifecycleWorker(growthSvc service.GrowthService) {
	go func() {
		log.Printf("[scheduler] start vip membership lifecycle worker")
		for {
			enabled, intervalMinutes := loadVIPMembershipLifecycleWorkerConfig(growthSvc)
			if enabled {
				runVIPMembershipLifecycleJob(growthSvc, "SYSTEM_TIMER")
			}
			if intervalMinutes <= 0 {
				intervalMinutes = vipLifecycleDefaultMinutes
			}
			time.Sleep(time.Duration(intervalMinutes) * time.Minute)
		}
	}()
}

func runVIPMembershipLifecycleJob(growthSvc service.GrowthService, triggerSource string) {
	summary, runErr := growthSvc.AdminRunVIPMembershipLifecycle()
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	_, logErr := growthSvc.AdminCreateSchedulerJobRun(
		vipLifecycleJobName,
		triggerSource,
		status,
		summary,
		errorMessage,
		"system",
	)
	if logErr != nil {
		log.Printf("[scheduler] create job run failed(%s): %v", vipLifecycleJobName, logErr)
	}
	if runErr != nil {
		log.Printf("[scheduler] job failed(%s): %v", vipLifecycleJobName, runErr)
		return
	}
	log.Printf("[scheduler] job success(%s): %s", vipLifecycleJobName, strings.TrimSpace(summary))
}

func loadDocFastIncrementalWorkerConfig(growthSvc service.GrowthService) (bool, int) {
	enabled := true
	intervalMinutes := docFastIncrementalDefaultMinutes

	items, _, err := growthSvc.AdminListSystemConfigs("news.sync.doc_fast.", 1, 200)
	if err != nil {
		return enabled, intervalMinutes
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case "news.sync.doc_fast.enabled":
			enabled = parseRouterBoolConfig(value, enabled)
		case "news.sync.doc_fast.interval_minutes":
			intervalMinutes = parseRouterIntConfig(value, intervalMinutes)
		}
	}
	if intervalMinutes <= 0 {
		intervalMinutes = docFastIncrementalDefaultMinutes
	}
	if intervalMinutes > docFastIncrementalMaxMinutes {
		intervalMinutes = docFastIncrementalMaxMinutes
	}
	return enabled, intervalMinutes
}

func loadTushareNewsIncrementalWorkerConfig(growthSvc service.GrowthService) (bool, int) {
	enabled := true
	intervalMinutes := tushareNewsIncrementalDefaultMinutes

	items, _, err := growthSvc.AdminListSystemConfigs("news.sync.tushare.", 1, 200)
	if err != nil {
		return enabled, intervalMinutes
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case "news.sync.tushare.enabled":
			enabled = parseRouterBoolConfig(value, enabled)
		case "news.sync.tushare.interval_minutes":
			intervalMinutes = parseRouterIntConfig(value, intervalMinutes)
		}
	}
	if intervalMinutes <= 0 {
		intervalMinutes = tushareNewsIncrementalDefaultMinutes
	}
	if intervalMinutes > tushareNewsIncrementalMaxMinutes {
		intervalMinutes = tushareNewsIncrementalMaxMinutes
	}
	return enabled, intervalMinutes
}

func loadVIPMembershipLifecycleWorkerConfig(growthSvc service.GrowthService) (bool, int) {
	enabled := true
	intervalMinutes := vipLifecycleDefaultMinutes

	items, _, err := growthSvc.AdminListSystemConfigs("membership.vip.lifecycle.", 1, 50)
	if err != nil {
		return enabled, intervalMinutes
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case "membership.vip.lifecycle.enabled":
			enabled = parseRouterBoolConfig(value, enabled)
		case "membership.vip.lifecycle.interval_minutes":
			intervalMinutes = parseRouterIntConfig(value, intervalMinutes)
		}
	}
	if intervalMinutes <= 0 {
		intervalMinutes = vipLifecycleDefaultMinutes
	}
	if intervalMinutes > vipLifecycleMaxMinutes {
		intervalMinutes = vipLifecycleMaxMinutes
	}
	return enabled, intervalMinutes
}

func parseRouterBoolConfig(raw string, fallback bool) bool {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" {
		return fallback
	}
	switch text {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func parseRouterIntConfig(raw string, fallback int) int {
	text := strings.TrimSpace(raw)
	if text == "" {
		return fallback
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return fallback
	}
	return value
}
