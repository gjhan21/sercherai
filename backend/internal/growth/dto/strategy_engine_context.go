package dto

type StrategyEngineStockSelectionContextRequest struct {
	TradeDate        string   `json:"trade_date"`
	SelectionMode    string   `json:"selection_mode" binding:"omitempty,oneof=AUTO MANUAL DEBUG"`
	UniverseScope    string   `json:"universe_scope"`
	ProfileID        string   `json:"profile_id"`
	DebugSeedSymbols []string `json:"debug_seed_symbols"`
	SeedSymbols      []string `json:"seed_symbols"`
	ExcludedSymbols  []string `json:"excluded_symbols"`
	Limit            int      `json:"limit" binding:"omitempty,gte=1,lte=50"`
	MarketScope      string   `json:"market_scope"`
	MinListingDays   int      `json:"min_listing_days" binding:"omitempty,gte=0,lte=3650"`
	MinAvgTurnover   float64  `json:"min_avg_turnover" binding:"omitempty,gte=0"`
}

type StrategyEngineFuturesStrategyContextRequest struct {
	TradeDate                       string   `json:"trade_date"`
	Contracts                       []string `json:"contracts"`
	Limit                           int      `json:"limit" binding:"omitempty,gte=1,lte=20"`
	AllowMockFallbackOnShortHistory bool     `json:"allow_mock_fallback_on_short_history"`
}
