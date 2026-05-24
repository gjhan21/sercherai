package router

import (
	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/handler"
	"sercherai/backend/internal/platform/config"
	"sercherai/backend/internal/platform/middleware"
)

func registerUserRoutes(v1 *gin.RouterGroup, userGrowthHandler *handler.UserGrowthHandler, cfg *config.Config) {
	user := v1.Group("/user")
	user.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
	{
		user.GET("/profile", userGrowthHandler.GetUserProfile)
		user.PUT("/profile", userGrowthHandler.UpdateUserProfile)

		user.GET("/virtual-sandbox", userGrowthHandler.GetUserVirtualSandbox)
		user.POST("/virtual-sandbox", userGrowthHandler.AddUserVirtualSandbox)

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

	search := v1.Group("/search")
	search.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
	{
		search.GET("/global", userGrowthHandler.SearchGlobal)
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
		futures.GET("/strategies/:id/version-history", userGrowthHandler.GetFuturesStrategyVersionHistory)
	}

	stocks := v1.Group("/stocks")
	stocks.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
	{
		stocks.GET("/recommendations", userGrowthHandler.ListStockRecommendations)
		stocks.GET("/recommendations/:id", userGrowthHandler.GetStockRecommendationDetail)
		stocks.GET("/recommendations/:id/performance", userGrowthHandler.GetStockRecommendationPerformance)
		stocks.GET("/recommendations/:id/insight", userGrowthHandler.GetStockRecommendationInsight)
		stocks.GET("/recommendations/:id/version-history", userGrowthHandler.GetStockRecommendationVersionHistory)
	}

	forecast := v1.Group("/forecast")
	forecast.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
	{
		forecast.POST("/runs", userGrowthHandler.CreateForecastL3Run)
		forecast.GET("/runs", userGrowthHandler.ListForecastL3Runs)
		forecast.GET("/runs/:id", userGrowthHandler.GetForecastL3RunDetail)
		forecast.GET("/runs/:id/review", userGrowthHandler.GetForecastL3RunReview)
		forecast.GET("/targets/history", userGrowthHandler.ListForecastL3History)
		forecast.GET("/targets/history/compare", userGrowthHandler.GetForecastL3HistoryCompare)
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

	community := v1.Group("/community")
	community.Use(middleware.AuthRequired(cfg.JWTSecret), middleware.RoleRequired("USER", "ADMIN"))
	{
		community.GET("/topics", userGrowthHandler.ListCommunityTopics)
		community.GET("/me/topics", userGrowthHandler.ListMyCommunityTopics)
		community.GET("/me/comments", userGrowthHandler.ListMyCommunityComments)
		community.POST("/topics", userGrowthHandler.CreateCommunityTopic)
		community.GET("/topics/:id", userGrowthHandler.GetCommunityTopic)
		community.GET("/topics/:id/comments", userGrowthHandler.ListCommunityComments)
		community.POST("/topics/:id/comments", userGrowthHandler.CreateCommunityComment)
		community.POST("/reactions", userGrowthHandler.CreateCommunityReaction)
		community.DELETE("/reactions", userGrowthHandler.DeleteCommunityReaction)
		community.POST("/reports", userGrowthHandler.CreateCommunityReport)
	}
}
