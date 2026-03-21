package model

type AdminStockSelectionOverview struct {
	DefaultProfile    *StockSelectionProfile           `json:"default_profile,omitempty"`
	LatestTradeDate   string                           `json:"latest_trade_date,omitempty"`
	LatestRun         *StockSelectionRun               `json:"latest_run,omitempty"`
	LatestSuccessRun  *StockSelectionRun               `json:"latest_success_run,omitempty"`
	LatestApprovedPortfolio []StockSelectionPortfolioEntry `json:"latest_approved_portfolio,omitempty"`
	MarketRegime      string                           `json:"market_regime,omitempty"`
	DataFreshness     map[string]any                   `json:"data_freshness,omitempty"`
	EvaluationSummary map[string]any                   `json:"evaluation_summary,omitempty"`
	EvaluationSummaryV2 map[string]any                 `json:"evaluation_summary_1_3_5_10_20,omitempty"`
	TemplateSummary   map[string]any                   `json:"template_summary,omitempty"`
	PendingReviewCount int                             `json:"pending_review_count,omitempty"`
	Warnings          []string                         `json:"warnings,omitempty"`
	QuickActions      []AdminStockSelectionQuickAction `json:"quick_actions,omitempty"`
}

type AdminStockSelectionQuickAction struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	ActionType  string `json:"action_type"`
	TargetRoute string `json:"target_route,omitempty"`
}
