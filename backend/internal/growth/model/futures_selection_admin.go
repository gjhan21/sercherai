package model

type AdminFuturesSelectionOverview struct {
	DefaultProfile          *FuturesSelectionProfile         `json:"default_profile,omitempty"`
	LatestTradeDate         string                           `json:"latest_trade_date,omitempty"`
	LatestRun               *FuturesSelectionRun             `json:"latest_run,omitempty"`
	LatestSuccessRun        *FuturesSelectionRun             `json:"latest_success_run,omitempty"`
	LatestApprovedPortfolio []FuturesSelectionPortfolioEntry `json:"latest_approved_portfolio,omitempty"`
	MarketRegime            string                           `json:"market_regime,omitempty"`
	DataFreshness           map[string]any                   `json:"data_freshness,omitempty"`
	PendingReviewCount      int                              `json:"pending_review_count,omitempty"`
	Warnings                []string                         `json:"warnings,omitempty"`
	QuickActions            []AdminStockSelectionQuickAction `json:"quick_actions,omitempty"`
}

type FuturesSelectionEvaluationLeaderboardItem struct {
	TemplateID       string             `json:"template_id,omitempty"`
	TemplateName     string             `json:"template_name,omitempty"`
	ProfileID        string             `json:"profile_id,omitempty"`
	ProfileName      string             `json:"profile_name,omitempty"`
	MarketRegime     string             `json:"market_regime,omitempty"`
	SampleCount      int                `json:"sample_count"`
	ReturnByHorizon  map[string]float64 `json:"return_by_horizon,omitempty"`
	HitRateByHorizon map[string]float64 `json:"hit_rate_by_horizon,omitempty"`
	MaxDrawdownPct   float64            `json:"max_drawdown_pct"`
}
