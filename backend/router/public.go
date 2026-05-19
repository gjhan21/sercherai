package router

import (
	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/handler"
)

func registerPublicRoutes(r *gin.Engine, v1 *gin.RouterGroup, userGrowthHandler *handler.UserGrowthHandler, adminHandlers *handler.AdminHandlers) {
	public := v1.Group("/public")
	{
		public.GET("/holdings", userGrowthHandler.ListPublicHoldings)
		public.GET("/futures-positions", userGrowthHandler.ListPublicFuturesPositions)
		public.GET("/news/categories", userGrowthHandler.ListNewsCategories)
		public.GET("/news/articles", userGrowthHandler.ListNewsArticles)
		public.GET("/news/articles/:id", userGrowthHandler.GetNewsArticleDetail)
		public.GET("/news/articles/:id/attachments", userGrowthHandler.ListNewsAttachments)
		public.GET("/search/global", userGrowthHandler.SearchGlobal)
		public.GET("/community/topics", userGrowthHandler.ListPublicCommunityTopics)
		public.GET("/community/topics/:id", userGrowthHandler.GetPublicCommunityTopic)
		public.GET("/community/topics/:id/comments", userGrowthHandler.ListPublicCommunityComments)
		public.POST("/experiments/events", userGrowthHandler.TrackExperimentEvent)
	}

	market := v1.Group("/market")
	{
		market.GET("/events", userGrowthHandler.ListMarketEvents)
		market.GET("/events/:id", userGrowthHandler.GetMarketEventDetail)
	}

	// Public stock K-line data
	v1.GET("/public/stocks/kline", userGrowthHandler.GetStockKline)
	v1.GET("/public/stocks/pattern-match", userGrowthHandler.PatternMatch)

	payment := v1.Group("/payment")
	{
		payment.POST("/callbacks/:channel", userGrowthHandler.HandlePaymentCallback)
	}
	v1.Any("/payment/callbacks/yolkpay/notify", userGrowthHandler.HandleYolkPayCallback)

	internalV1 := r.Group("/internal/v1")
	{
		internalStrategy := internalV1.Group("/strategy-engine")
		{
			internalStrategy.POST("/context/stock-selection", adminHandlers.Strategy.InternalStrategyEngineStockSelectionContext)
			internalStrategy.POST("/context/futures-strategy", adminHandlers.Strategy.InternalStrategyEngineFuturesStrategyContext)
		}
	}
}
