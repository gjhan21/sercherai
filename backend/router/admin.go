package router

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/handler"
	"sercherai/backend/internal/platform/config"
	"sercherai/backend/internal/platform/middleware"
)

func registerAdminRoutes(v1 *gin.RouterGroup, adminHandlers *handler.AdminHandlers, cfg *config.Config, db *sql.DB) {
	admin := v1.Group("/admin/growth")
	admin.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		admin.GET("/invite-records", middleware.PermissionRequired(db, "growth.view"), adminHandlers.Finance.ListInviteRecords)
		admin.GET("/reward-records", middleware.PermissionRequired(db, "growth.view"), adminHandlers.Finance.ListRewardRecords)
		admin.PUT("/reward-records/:id/review", middleware.PermissionRequired(db, "growth.edit"), adminHandlers.Finance.ReviewRewardRecord)
	}

	adminPayment := v1.Group("/admin/payment")
	adminPayment.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminPayment.GET("/reconciliation", middleware.PermissionRequired(db, "payment.view"), adminHandlers.Finance.ListReconciliation)
		adminPayment.POST("/reconciliation/:batch_id/retry", middleware.PermissionRequired(db, "payment.edit"), adminHandlers.Finance.RetryReconciliation)
	}

	adminRisk := v1.Group("/admin/risk")
	adminRisk.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminRisk.GET("/rules", middleware.PermissionRequired(db, "risk.view"), adminHandlers.Risk.ListRiskRules)
		adminRisk.POST("/rules", middleware.PermissionRequired(db, "risk.edit"), adminHandlers.Risk.CreateRiskRule)
		adminRisk.PUT("/rules/:id", middleware.PermissionRequired(db, "risk.edit"), adminHandlers.Risk.UpdateRiskRule)
		adminRisk.GET("/hits", middleware.PermissionRequired(db, "risk.view"), adminHandlers.Risk.ListRiskHits)
		adminRisk.PUT("/hits/:id/review", middleware.PermissionRequired(db, "risk.edit"), adminHandlers.Risk.ReviewRiskHit)
	}

	adminRewardWallet := v1.Group("/admin/reward-wallet")
	adminRewardWallet.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminRewardWallet.GET("/withdraw-requests", middleware.PermissionRequired(db, "reward_wallet.view"), adminHandlers.Finance.ListWithdrawRequests)
		adminRewardWallet.PUT("/withdraw-requests/:id/review", middleware.PermissionRequired(db, "reward_wallet.edit"), adminHandlers.Finance.ReviewWithdrawRequest)
	}

	adminNews := v1.Group("/admin/news")
	adminNews.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminNews.GET("/categories", middleware.PermissionRequired(db, "news.view"), adminHandlers.News.ListNewsCategories)
		adminNews.POST("/categories", middleware.PermissionRequired(db, "news.edit"), adminHandlers.News.CreateNewsCategory)
		adminNews.PUT("/categories/:id", middleware.PermissionRequired(db, "news.edit"), adminHandlers.News.UpdateNewsCategory)

		adminNews.GET("/articles", middleware.PermissionRequired(db, "news.view"), adminHandlers.News.ListNewsArticles)
		adminNews.GET("/articles/:id", middleware.PermissionRequired(db, "news.view"), adminHandlers.News.GetNewsArticleDetail)
		adminNews.POST("/articles", middleware.PermissionRequired(db, "news.edit"), adminHandlers.News.CreateNewsArticle)
		adminNews.PUT("/articles/:id", middleware.PermissionRequired(db, "news.edit"), adminHandlers.News.UpdateNewsArticle)
		adminNews.PUT("/articles/:id/publish", middleware.PermissionRequired(db, "news.edit"), adminHandlers.News.PublishNewsArticle)

		adminNews.POST("/attachments/upload", middleware.PermissionRequired(db, "news.edit"), adminHandlers.News.UploadNewsAttachment)
		adminNews.GET("/articles/:id/attachments", middleware.PermissionRequired(db, "news.view"), adminHandlers.News.ListNewsAttachments)
		adminNews.POST("/articles/:id/attachments", middleware.PermissionRequired(db, "news.edit"), adminHandlers.News.CreateNewsAttachment)
		adminNews.DELETE("/attachments/:id", middleware.PermissionRequired(db, "news.edit"), adminHandlers.News.DeleteNewsAttachment)
		adminNews.POST("/market-sync", middleware.PermissionRequired(db, "news.edit"), adminHandlers.MarketData.SyncMarketNewsSource)
	}

	adminCommunity := v1.Group("/admin/community")
	adminCommunity.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminCommunity.GET("/topics", middleware.PermissionRequired(db, "community.view"), adminHandlers.Community.ListCommunityTopics)
		adminCommunity.PUT("/topics/:id/status", middleware.PermissionRequired(db, "community.edit"), adminHandlers.Community.UpdateCommunityTopicStatus)
		adminCommunity.GET("/comments", middleware.PermissionRequired(db, "community.view"), adminHandlers.Community.ListCommunityComments)
		adminCommunity.PUT("/comments/:id/status", middleware.PermissionRequired(db, "community.edit"), adminHandlers.Community.UpdateCommunityCommentStatus)
		adminCommunity.GET("/reports", middleware.PermissionRequired(db, "community.view"), adminHandlers.Community.ListCommunityReports)
		adminCommunity.PUT("/reports/:id/review", middleware.PermissionRequired(db, "community.review"), adminHandlers.Community.ReviewCommunityReport)
	}

	adminDataSources := v1.Group("/admin/data-sources")
	adminDataSources.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminDataSources.GET("", middleware.PermissionRequired(db, "data_source.view"), adminHandlers.System.ListDataSources)
		adminDataSources.GET("/governance/overview", middleware.PermissionRequired(db, "data_source.view"), adminHandlers.MarketData.GetMarketProviderGovernanceOverview)
		adminDataSources.GET("/governance/capabilities", middleware.PermissionRequired(db, "data_source.view"), adminHandlers.MarketData.ListMarketProviderCapabilities)
		adminDataSources.GET("/governance/routing-policies", middleware.PermissionRequired(db, "data_source.view"), adminHandlers.MarketData.ListMarketProviderRoutingPolicies)
		adminDataSources.PUT("/governance/routing-policies/:policy_key", middleware.PermissionRequired(db, "data_source.edit"), adminHandlers.MarketData.UpdateMarketProviderRoutingPolicy)
		adminDataSources.GET("/market-quality-logs", middleware.PermissionRequired(db, "data_source.view"), adminHandlers.MarketData.ListMarketDataQualityLogs)
		adminDataSources.GET("/market-quality-summary", middleware.PermissionRequired(db, "data_source.view"), adminHandlers.MarketData.GetMarketDataQualitySummary)
		adminDataSources.GET("/market-derived-truth-summary", middleware.PermissionRequired(db, "data_source.view"), adminHandlers.MarketData.GetMarketDerivedTruthSummary)
		adminDataSources.POST("/health-checks", middleware.PermissionRequired(db, "data_source.edit"), adminHandlers.System.BatchCheckDataSourcesHealth)
		adminDataSources.POST("", middleware.PermissionRequired(db, "data_source.edit"), adminHandlers.System.CreateDataSource)
		adminDataSources.PUT("/:source_key", middleware.PermissionRequired(db, "data_source.edit"), adminHandlers.System.UpdateDataSource)
		adminDataSources.DELETE("/:source_key", middleware.PermissionRequired(db, "data_source.edit"), adminHandlers.System.DeleteDataSource)
		adminDataSources.POST("/:source_key/health-check", middleware.PermissionRequired(db, "data_source.edit"), adminHandlers.System.CheckDataSourceHealth)
		adminDataSources.GET("/:source_key/health-logs", middleware.PermissionRequired(db, "data_source.view"), adminHandlers.System.ListDataSourceHealthLogs)
	}

	adminMarketData := v1.Group("/admin/market-data")
	adminMarketData.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminMarketData.POST("/backfill", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.CreateMarketDataBackfillRun)
		adminMarketData.POST("/master/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncMarketDataMaster)
		adminMarketData.POST("/quotes/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncMarketDataQuotes)
		adminMarketData.POST("/daily-basic/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncMarketDataDailyBasic)
		adminMarketData.POST("/moneyflow/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncMarketDataMoneyflow)
		adminMarketData.POST("/truth/rebuild", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.RebuildMarketDataTruth)
		adminMarketData.GET("/backfill-runs", middleware.PermissionRequired(db, "market.view"), adminHandlers.MarketData.ListMarketDataBackfillRuns)
		adminMarketData.GET("/backfill-runs/:id", middleware.PermissionRequired(db, "market.view"), adminHandlers.MarketData.GetMarketDataBackfillRun)
		adminMarketData.GET("/backfill-runs/:id/details", middleware.PermissionRequired(db, "market.view"), adminHandlers.MarketData.ListMarketDataBackfillRunDetails)
		adminMarketData.POST("/backfill-runs/:id/retry", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.RetryMarketDataBackfillRun)
		adminMarketData.POST("/backfill-runs/:id/cancel", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.CancelMarketDataBackfillRun)
		adminMarketData.GET("/universe-snapshots", middleware.PermissionRequired(db, "market.view"), adminHandlers.MarketData.ListMarketUniverseSnapshots)
		adminMarketData.GET("/universe-snapshots/:id", middleware.PermissionRequired(db, "market.view"), adminHandlers.MarketData.GetMarketUniverseSnapshot)
		adminMarketData.GET("/coverage-summary", middleware.PermissionRequired(db, "market.view"), adminHandlers.MarketData.GetMarketCoverageSummary)
	}

	adminForecast := v1.Group("/admin/forecast")
	adminForecast.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminForecast.GET("/runs", middleware.PermissionRequired(db, "forecast_l3.view"), adminHandlers.Forecast.ListForecastL3Runs)
		adminForecast.POST("/runs", middleware.PermissionRequired(db, "forecast_l3.edit"), adminHandlers.Forecast.CreateForecastL3Run)
		adminForecast.GET("/runs/:id", middleware.PermissionRequired(db, "forecast_l3.view"), adminHandlers.Forecast.GetForecastL3RunDetail)
		adminForecast.POST("/runs/:id/retry", middleware.PermissionRequired(db, "forecast_l3.edit"), adminHandlers.Forecast.RetryForecastL3Run)
		adminForecast.POST("/runs/:id/cancel", middleware.PermissionRequired(db, "forecast_l3.edit"), adminHandlers.Forecast.CancelForecastL3Run)
		adminForecast.GET("/quality", middleware.PermissionRequired(db, "forecast_l3.view"), adminHandlers.Forecast.ListForecastL3Quality)
	}

	adminStocks := v1.Group("/admin/stocks")
	adminStocks.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminStocks.GET("/recommendations", middleware.PermissionRequired(db, "market.view"), adminHandlers.StockSelection.ListStockRecommendations)
		adminStocks.POST("/recommendations", middleware.PermissionRequired(db, "market.edit"), adminHandlers.StockSelection.CreateStockRecommendation)
		adminStocks.PUT("/recommendations/:id/status", middleware.PermissionRequired(db, "market.edit"), adminHandlers.StockSelection.UpdateStockRecommendationStatus)
		adminStocks.POST("/recommendations/:id/generate-review", middleware.PermissionRequired(db, "market.edit"), adminHandlers.StockSelection.GenerateStockRecommendationAIReview)
		adminStocks.POST("/master/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.StockSelection.SyncStockInstrumentMaster)
		adminStocks.POST("/quotes/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.StockSelection.SyncStockQuotes)
		adminStocks.POST("/daily-basic/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncStockDailyBasics)
		adminStocks.POST("/moneyflow/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncStockMoneyflows)
		adminStocks.POST("/news/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncStockNewsSource)
		adminStocks.POST("/kpl-list/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncStockKPLList)
		adminStocks.POST("/top-list/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncStockTopList)
		adminStocks.POST("/backfill", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.BackfillStockMarketData)
		adminStocks.POST("/quotes/rebuild-derived-truth", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.RebuildStockDerivedTruth)
			adminStocks.POST("/quotes/full-sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.FullSyncStockQuotes)
			adminStocks.POST("/quotes/incremental-sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.IncrementalSyncStockQuotes)
			adminStocks.GET("/quotes/sync-progress", adminHandlers.MarketData.GetSyncProgress)
		adminStocks.GET("/quant/top", middleware.PermissionRequired(db, "market.view"), adminHandlers.StockSelection.ListQuantTopStocks)
		adminStocks.GET("/quant/evaluation", middleware.PermissionRequired(db, "market.view"), adminHandlers.StockSelection.ListQuantEvaluation)
		adminStocks.GET("/quant/evaluation/export.csv", middleware.PermissionRequired(db, "market.view"), adminHandlers.StockSelection.ExportQuantEvaluationCSV)
		adminStocks.POST("/recommendations/generate-daily", middleware.PermissionRequired(db, "market.edit"), adminHandlers.StockSelection.GenerateDailyStockRecommendations)
		adminStocks.GET("/strategy-engine/publish-history", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.ListStrategyEngineStockPublishHistory)
		adminStocks.GET("/strategy-engine/publish-records/:publish_id", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.GetStrategyEngineStockPublishRecord)
		adminStocks.GET("/strategy-engine/publish-records/:publish_id/replay", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.GetStrategyEngineStockPublishReplay)
		adminStocks.POST("/strategy-engine/publish-compare", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.CompareStrategyEngineStockPublishVersions)
	}

	adminStockSelection := v1.Group("/admin/stock-selection")
	adminStockSelection.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminStockSelection.GET("/overview", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.GetStockSelectionOverview)
		adminStockSelection.GET("/runs", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockSelectionRuns)
		adminStockSelection.POST("/runs", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.CreateStockSelectionRun)
		adminStockSelection.GET("/runs/compare", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.CompareStockSelectionRuns)
		adminStockSelection.GET("/runs/:run_id", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.GetStockSelectionRun)
		adminStockSelection.GET("/runs/:run_id/candidates", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockSelectionRunCandidates)
		adminStockSelection.GET("/runs/:run_id/portfolio", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockSelectionRunPortfolio)
		adminStockSelection.GET("/runs/:run_id/evidence", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockSelectionRunEvidence)
		adminStockSelection.GET("/runs/:run_id/evaluation", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockSelectionRunEvaluations)
		adminStockSelection.GET("/profiles", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockSelectionProfiles)
		adminStockSelection.GET("/profiles/:id/versions", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockSelectionProfileVersions)
		adminStockSelection.POST("/profiles", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.CreateStockSelectionProfile)
		adminStockSelection.PUT("/profiles/:id", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.UpdateStockSelectionProfile)
		adminStockSelection.POST("/profiles/:id/publish", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.PublishStockSelectionProfile)
		adminStockSelection.POST("/profiles/:id/rollback", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.RollbackStockSelectionProfile)
		adminStockSelection.GET("/templates", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockSelectionProfileTemplates)
		adminStockSelection.POST("/templates", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.CreateStockSelectionProfileTemplate)
		adminStockSelection.PUT("/templates/:id", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.UpdateStockSelectionProfileTemplate)
		adminStockSelection.POST("/templates/:id/set-default", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.SetDefaultStockSelectionProfileTemplate)
		adminStockSelection.GET("/evaluation/leaderboard", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockSelectionEvaluationLeaderboard)
		adminStockSelection.GET("/reviews", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockSelectionReviews)
		adminStockSelection.POST("/reviews/:run_id/approve", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.ApproveStockSelectionReview)
		adminStockSelection.POST("/reviews/:run_id/reject", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.RejectStockSelectionReview)
		adminStockSelection.GET("/events", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.ListStockEventClusters)
		adminStockSelection.GET("/events/:id", middleware.PermissionRequired(db, "stock_selection.view"), adminHandlers.StockSelection.GetStockEventCluster)
		adminStockSelection.POST("/events/:id/review", middleware.PermissionRequired(db, "stock_selection.manage"), adminHandlers.StockSelection.ReviewStockEventCluster)
	}

	adminFuturesSelection := v1.Group("/admin/futures-selection")
	adminFuturesSelection.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminFuturesSelection.GET("/overview", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.GetFuturesSelectionOverview)
		adminFuturesSelection.GET("/runs", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.ListFuturesSelectionRuns)
		adminFuturesSelection.POST("/runs", middleware.PermissionRequired(db, "futures_selection.manage"), adminHandlers.FuturesSelection.CreateFuturesSelectionRun)
		adminFuturesSelection.GET("/runs/compare", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.CompareFuturesSelectionRuns)
		adminFuturesSelection.GET("/runs/:run_id", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.GetFuturesSelectionRun)
		adminFuturesSelection.GET("/runs/:run_id/candidates", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.ListFuturesSelectionRunCandidates)
		adminFuturesSelection.GET("/runs/:run_id/portfolio", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.ListFuturesSelectionRunPortfolio)
		adminFuturesSelection.GET("/runs/:run_id/evidence", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.ListFuturesSelectionRunEvidence)
		adminFuturesSelection.GET("/runs/:run_id/evaluation", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.ListFuturesSelectionRunEvaluations)
		adminFuturesSelection.GET("/profiles", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.ListFuturesSelectionProfiles)
		adminFuturesSelection.GET("/profiles/:id/versions", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.ListFuturesSelectionProfileVersions)
		adminFuturesSelection.POST("/profiles", middleware.PermissionRequired(db, "futures_selection.manage"), adminHandlers.FuturesSelection.CreateFuturesSelectionProfile)
		adminFuturesSelection.PUT("/profiles/:id", middleware.PermissionRequired(db, "futures_selection.manage"), adminHandlers.FuturesSelection.UpdateFuturesSelectionProfile)
		adminFuturesSelection.POST("/profiles/:id/publish", middleware.PermissionRequired(db, "futures_selection.manage"), adminHandlers.FuturesSelection.PublishFuturesSelectionProfile)
		adminFuturesSelection.POST("/profiles/:id/rollback", middleware.PermissionRequired(db, "futures_selection.manage"), adminHandlers.FuturesSelection.RollbackFuturesSelectionProfile)
		adminFuturesSelection.GET("/templates", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.ListFuturesSelectionProfileTemplates)
		adminFuturesSelection.POST("/templates", middleware.PermissionRequired(db, "futures_selection.manage"), adminHandlers.FuturesSelection.CreateFuturesSelectionProfileTemplate)
		adminFuturesSelection.PUT("/templates/:id", middleware.PermissionRequired(db, "futures_selection.manage"), adminHandlers.FuturesSelection.UpdateFuturesSelectionProfileTemplate)
		adminFuturesSelection.POST("/templates/:id/set-default", middleware.PermissionRequired(db, "futures_selection.manage"), adminHandlers.FuturesSelection.SetDefaultFuturesSelectionProfileTemplate)
		adminFuturesSelection.GET("/evaluation/leaderboard", middleware.PermissionRequired(db, "futures_selection.view"), adminHandlers.FuturesSelection.ListFuturesSelectionEvaluationLeaderboard)
		adminFuturesSelection.POST("/reviews/:run_id/approve", middleware.PermissionRequired(db, "futures_selection.manage"), adminHandlers.FuturesSelection.ApproveFuturesSelectionReview)
		adminFuturesSelection.POST("/reviews/:run_id/reject", middleware.PermissionRequired(db, "futures_selection.manage"), adminHandlers.FuturesSelection.RejectFuturesSelectionReview)
	}

	adminStrategyGraph := v1.Group("/admin/strategy-graph")
	adminStrategyGraph.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminStrategyGraph.GET("/snapshots/:snapshot_id", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.GetStrategyGraphSnapshot)
		adminStrategyGraph.GET("/subgraph", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.QueryStrategyGraphSubgraph)
	}

	adminFutures := v1.Group("/admin/futures")
	adminFutures.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminFutures.GET("/strategies", middleware.PermissionRequired(db, "market.view"), adminHandlers.FuturesSelection.ListFuturesStrategies)
		adminFutures.POST("/strategies", middleware.PermissionRequired(db, "market.edit"), adminHandlers.FuturesSelection.CreateFuturesStrategy)
		adminFutures.PUT("/strategies/:id/status", middleware.PermissionRequired(db, "market.edit"), adminHandlers.FuturesSelection.UpdateFuturesStrategyStatus)
		adminFutures.POST("/quotes/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncFuturesQuotes)
		adminFutures.POST("/quotes/rebuild-derived-truth", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.RebuildFuturesDerivedTruth)
		adminFutures.POST("/inventory/sync", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.SyncFuturesInventory)
		adminFutures.POST("/strategies/generate-daily", middleware.PermissionRequired(db, "market.edit"), adminHandlers.FuturesSelection.GenerateDailyFuturesStrategies)
		adminFutures.GET("/strategy-engine/publish-history", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.ListStrategyEngineFuturesPublishHistory)
		adminFutures.GET("/strategy-engine/publish-records/:publish_id", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.GetStrategyEngineFuturesPublishRecord)
		adminFutures.GET("/strategy-engine/publish-records/:publish_id/replay", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.GetStrategyEngineFuturesPublishReplay)
		adminFutures.POST("/strategy-engine/publish-compare", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.CompareStrategyEngineFuturesPublishVersions)
	}

	adminMarket := v1.Group("/admin/market")
	adminMarket.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminMarket.GET("/events", middleware.PermissionRequired(db, "market.view"), adminHandlers.MarketData.ListMarketEvents)
		adminMarket.GET("/rhythm-tasks", middleware.PermissionRequired(db, "market.view"), adminHandlers.MarketData.ListMarketRhythmTasks)
		adminMarket.POST("/rhythm-tasks/ensure", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.EnsureMarketRhythmTasks)
		adminMarket.PUT("/rhythm-tasks/:id", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.UpdateMarketRhythmTask)
		adminMarket.PUT("/rhythm-tasks/:id/status", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.UpdateMarketRhythmTaskStatus)
		adminMarket.GET("/experiments/summary", middleware.PermissionRequired(db, "market.view"), adminHandlers.System.GetExperimentAnalyticsSummary)
		adminMarket.GET("/strategy-engine/seed-sets", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.ListStrategySeedSets)
		adminMarket.POST("/strategy-engine/seed-sets", middleware.PermissionRequired(db, "market.edit"), adminHandlers.Strategy.CreateStrategySeedSet)
		adminMarket.PUT("/strategy-engine/seed-sets/:id", middleware.PermissionRequired(db, "market.edit"), adminHandlers.Strategy.UpdateStrategySeedSet)
		adminMarket.GET("/strategy-engine/agents", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.ListStrategyAgentProfiles)
		adminMarket.POST("/strategy-engine/agents", middleware.PermissionRequired(db, "market.edit"), adminHandlers.Strategy.CreateStrategyAgentProfile)
		adminMarket.PUT("/strategy-engine/agents/:id", middleware.PermissionRequired(db, "market.edit"), adminHandlers.Strategy.UpdateStrategyAgentProfile)
		adminMarket.GET("/strategy-engine/scenarios", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.ListStrategyScenarioTemplates)
		adminMarket.POST("/strategy-engine/scenarios", middleware.PermissionRequired(db, "market.edit"), adminHandlers.Strategy.CreateStrategyScenarioTemplate)
		adminMarket.PUT("/strategy-engine/scenarios/:id", middleware.PermissionRequired(db, "market.edit"), adminHandlers.Strategy.UpdateStrategyScenarioTemplate)
		adminMarket.GET("/strategy-engine/publish-policies", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.ListStrategyPublishPolicies)
		adminMarket.POST("/strategy-engine/publish-policies", middleware.PermissionRequired(db, "market.edit"), adminHandlers.Strategy.CreateStrategyPublishPolicy)
		adminMarket.PUT("/strategy-engine/publish-policies/:id", middleware.PermissionRequired(db, "market.edit"), adminHandlers.Strategy.UpdateStrategyPublishPolicy)
		adminMarket.GET("/strategy-engine/jobs", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.ListStrategyEngineJobs)
		adminMarket.GET("/strategy-engine/jobs/:job_id", middleware.PermissionRequired(db, "market.view"), adminHandlers.Strategy.GetStrategyEngineJob)
		adminMarket.POST("/strategy-engine/jobs/:job_id/publish", middleware.PermissionRequired(db, "market.edit"), adminHandlers.Strategy.PublishStrategyEngineJob)
		adminMarket.POST("/events", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.CreateMarketEvent)
		adminMarket.PUT("/events/:id", middleware.PermissionRequired(db, "market.edit"), adminHandlers.MarketData.UpdateMarketEvent)
	}

	adminUsers := v1.Group("/admin/users")
	adminUsers.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminUsers.GET("", middleware.PermissionRequired(db, "users.view"), adminHandlers.User.ListUsers)
		adminUsers.GET("/source-summary", middleware.PermissionRequired(db, "users.view"), adminHandlers.User.UserSourceSummary)
		adminUsers.GET("/export.csv", middleware.PermissionRequired(db, "users.view"), adminHandlers.User.ExportUsersCSV)
		adminUsers.GET("/browse-histories", middleware.PermissionRequired(db, "users.view"), adminHandlers.User.ListBrowseHistories)
		adminUsers.GET("/browse-histories/summary", middleware.PermissionRequired(db, "users.view"), adminHandlers.User.BrowseHistorySummary)
		adminUsers.GET("/browse-histories/trend", middleware.PermissionRequired(db, "users.view"), adminHandlers.User.BrowseHistoryTrend)
		adminUsers.GET("/browse-histories/segments", middleware.PermissionRequired(db, "users.view"), adminHandlers.User.ListBrowseUserSegments)
		adminUsers.GET("/browse-histories/export.csv", middleware.PermissionRequired(db, "users.view"), adminHandlers.User.ExportBrowseHistoriesCSV)
		adminUsers.GET("/messages", middleware.PermissionRequired(db, "users.view"), adminHandlers.User.ListUserMessages)
		adminUsers.POST("/messages", middleware.PermissionRequired(db, "users.edit"), adminHandlers.User.CreateUserMessages)
		adminUsers.GET("/:id/center-overview", middleware.PermissionRequired(db, "users.view"), adminHandlers.User.GetUserCenterOverview)
		adminUsers.PUT("/:id/subscriptions/:sub_id", middleware.PermissionRequired(db, "users.edit"), adminHandlers.User.UpdateUserSubscription)
		adminUsers.PUT("/:id/status", middleware.PermissionRequired(db, "users.edit"), adminHandlers.User.UpdateUserStatus)
		adminUsers.PUT("/:id/member-level", middleware.PermissionRequired(db, "users.edit"), adminHandlers.User.UpdateUserMemberLevel)
		adminUsers.PUT("/:id/password", middleware.PermissionRequired(db, "users.edit"), adminHandlers.User.ResetUserPassword)
	}

	adminDashboard := v1.Group("/admin/dashboard")
	adminDashboard.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminDashboard.GET("/overview", middleware.PermissionRequired(db, "dashboard.view"), adminHandlers.Audit.DashboardOverview)
	}

	adminAudit := v1.Group("/admin/audit")
	adminAudit.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminAudit.GET("/events/summary", middleware.PermissionRequired(db, "audit.view"), adminHandlers.Audit.GetAuditEventSummary)
		adminAudit.GET("/events", middleware.PermissionRequired(db, "audit.view"), adminHandlers.Audit.ListAuditEvents)
		adminAudit.GET("/operation-logs", middleware.PermissionRequired(db, "audit.view"), adminHandlers.Audit.ListOperationLogs)
		adminAudit.GET("/operation-logs/export.csv", middleware.PermissionRequired(db, "audit.view"), adminHandlers.Audit.ExportOperationLogsCSV)
	}

	adminMembership := v1.Group("/admin/membership")
	adminMembership.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminMembership.GET("/products", middleware.PermissionRequired(db, "membership.view"), adminHandlers.Membership.ListMembershipProducts)
		adminMembership.POST("/products", middleware.PermissionRequired(db, "membership.edit"), adminHandlers.Membership.CreateMembershipProduct)
		adminMembership.PUT("/products/:id", middleware.PermissionRequired(db, "membership.edit"), adminHandlers.Membership.UpdateMembershipProduct)
		adminMembership.PUT("/products/:id/status", middleware.PermissionRequired(db, "membership.edit"), adminHandlers.Membership.UpdateMembershipProductStatus)

		adminMembership.GET("/orders", middleware.PermissionRequired(db, "membership.view"), adminHandlers.Membership.ListMembershipOrders)
		adminMembership.GET("/orders/export.csv", middleware.PermissionRequired(db, "membership.view"), adminHandlers.Membership.ListMembershipOrdersCSV)
		adminMembership.PUT("/orders/:id/status", middleware.PermissionRequired(db, "membership.edit"), adminHandlers.Membership.UpdateMembershipOrderStatus)

		adminMembership.GET("/quota-configs", middleware.PermissionRequired(db, "membership.view"), adminHandlers.Membership.ListVIPQuotaConfigs)
		adminMembership.POST("/quota-configs", middleware.PermissionRequired(db, "membership.edit"), adminHandlers.Membership.CreateVIPQuotaConfig)
		adminMembership.PUT("/quota-configs/:id", middleware.PermissionRequired(db, "membership.edit"), adminHandlers.Membership.UpdateVIPQuotaConfig)
		adminMembership.GET("/user-quotas", middleware.PermissionRequired(db, "membership.view"), adminHandlers.Membership.ListUserQuotas)
		adminMembership.PUT("/user-quotas/:user_id/adjust", middleware.PermissionRequired(db, "membership.edit"), adminHandlers.Membership.AdjustUserQuota)
	}

	adminSystem := v1.Group("/admin/system")
	adminSystem.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminSystem.GET("/configs", middleware.PermissionRequired(db, "system_config.view"), adminHandlers.System.ListSystemConfigs)
		adminSystem.PUT("/configs", middleware.PermissionRequired(db, "system_config.edit"), adminHandlers.System.UpsertSystemConfig)
		adminSystem.POST("/configs/oss/qiniu/test", middleware.PermissionRequired(db, "system_config.edit"), adminHandlers.System.TestOSSQiniuConfig)
		adminSystem.POST("/configs/payment/yolkpay/test", middleware.PermissionRequired(db, "system_config.edit"), adminHandlers.System.TestYolkPayConfig)

		adminSystem.GET("/job-definitions", middleware.PermissionRequired(db, "system_job.view"), adminHandlers.System.ListSchedulerJobDefinitions)
		adminSystem.GET("/job-definitions/supported", middleware.PermissionRequired(db, "system_job.view"), adminHandlers.System.ListSupportedSchedulerJobs)
		adminSystem.POST("/job-definitions", middleware.PermissionRequired(db, "system_job.edit"), adminHandlers.System.CreateSchedulerJobDefinition)
		adminSystem.PUT("/job-definitions/:id", middleware.PermissionRequired(db, "system_job.edit"), adminHandlers.System.UpdateSchedulerJobDefinition)
		adminSystem.PUT("/job-definitions/:id/status", middleware.PermissionRequired(db, "system_job.edit"), adminHandlers.System.UpdateSchedulerJobDefinitionStatus)
		adminSystem.DELETE("/job-definitions/:id", middleware.PermissionRequired(db, "system_job.edit"), adminHandlers.System.DeleteSchedulerJobDefinition)

		adminSystem.GET("/job-runs", middleware.PermissionRequired(db, "system_job.view"), adminHandlers.System.ListSchedulerJobRuns)
		adminSystem.GET("/job-runs/:id/news-sync-details", middleware.PermissionRequired(db, "system_job.view"), adminHandlers.System.ListNewsSyncRunDetails)
		adminSystem.GET("/job-runs/export.csv", middleware.PermissionRequired(db, "system_job.view"), adminHandlers.System.ExportSchedulerJobRunsCSV)
		adminSystem.GET("/job-runs/metrics", middleware.PermissionRequired(db, "system_job.view"), adminHandlers.System.SchedulerJobMetrics)
		adminSystem.POST("/job-runs/trigger", middleware.PermissionRequired(db, "system_job.edit"), adminHandlers.System.TriggerSchedulerJob)
		adminSystem.POST("/job-runs/:id/retry", middleware.PermissionRequired(db, "system_job.edit"), adminHandlers.System.RetrySchedulerJobRun)
		adminSystem.POST("/job-runs/:id/retry-news-sync-item", middleware.PermissionRequired(db, "system_job.edit"), adminHandlers.System.RetryNewsSyncItem)
	}

	adminWorkflow := v1.Group("/admin/workflow")
	adminWorkflow.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("ADMIN"))
	{
		adminWorkflow.GET("/reviews", middleware.PermissionRequired(db, "review.view"), adminHandlers.Audit.ListReviewTasks)
		adminWorkflow.GET("/reviews/export.csv", middleware.PermissionRequired(db, "review.view"), adminHandlers.Audit.ExportReviewTasksCSV)
		adminWorkflow.GET("/metrics", middleware.PermissionRequired(db, "review.view"), adminHandlers.Audit.WorkflowMetrics)
		adminWorkflow.POST("/reviews/submit", middleware.PermissionRequired(db, "review.edit"), adminHandlers.Audit.SubmitReviewTask)
		adminWorkflow.PUT("/reviews/:id/assign", middleware.PermissionRequired(db, "review.edit"), adminHandlers.Audit.AssignReviewTask)
		adminWorkflow.PUT("/reviews/:id/decision", middleware.PermissionRequired(db, "review.edit"), adminHandlers.Audit.ReviewTaskDecision)
		adminWorkflow.GET("/messages", middleware.PermissionRequired(db, "workflow.view"), adminHandlers.Audit.ListWorkflowMessages)
		adminWorkflow.GET("/messages/export.csv", middleware.PermissionRequired(db, "workflow.view"), adminHandlers.Audit.ExportWorkflowMessagesCSV)
		adminWorkflow.GET("/messages/unread-count", middleware.PermissionRequired(db, "workflow.view"), adminHandlers.Audit.CountUnreadWorkflowMessages)
		adminWorkflow.PUT("/messages/:id/read", middleware.PermissionRequired(db, "workflow.edit"), adminHandlers.Audit.UpdateWorkflowMessageRead)
		adminWorkflow.PUT("/messages/read-all", middleware.PermissionRequired(db, "workflow.edit"), adminHandlers.Audit.BulkReadWorkflowMessages)
	}
}
