package model

type StockSelectionProfileTemplate struct {
	ID                string         `json:"id"`
	TemplateKey       string         `json:"template_key"`
	Name              string         `json:"name"`
	Description       string         `json:"description,omitempty"`
	MarketRegimeBias  string         `json:"market_regime_bias,omitempty"`
	IsDefault         bool           `json:"is_default"`
	Status            string         `json:"status"`
	UniverseDefaults  map[string]any `json:"universe_defaults_json,omitempty"`
	SeedDefaults      map[string]any `json:"seed_defaults_json,omitempty"`
	FactorDefaults    map[string]any `json:"factor_defaults_json,omitempty"`
	PortfolioDefaults map[string]any `json:"portfolio_defaults_json,omitempty"`
	PublishDefaults   map[string]any `json:"publish_defaults_json,omitempty"`
	UpdatedBy         string         `json:"updated_by,omitempty"`
	UpdatedAt         string         `json:"updated_at,omitempty"`
	CreatedAt         string         `json:"created_at,omitempty"`
}

type StockSelectionRunCreateRequest struct {
	TradeDate                string `json:"trade_date"`
	ProfileID                string `json:"profile_id"`
	TemplateID               string `json:"template_id,omitempty"`
	CompareWithLastPublished bool   `json:"compare_with_last_published"`
	DryRun                   bool   `json:"dry_run"`
}

type StockSelectionRunEvidence struct {
	ID              string           `json:"id"`
	RunID           string           `json:"run_id"`
	Symbol          string           `json:"symbol"`
	Name            string           `json:"name,omitempty"`
	Stage           string           `json:"stage"`
	PortfolioRole   string           `json:"portfolio_role,omitempty"`
	EvidenceSummary string           `json:"evidence_summary,omitempty"`
	EvidenceCards   []map[string]any `json:"evidence_cards_json,omitempty"`
	PositiveReasons []string         `json:"positive_reasons_json,omitempty"`
	VetoReasons     []string         `json:"veto_reasons_json,omitempty"`
	ThemeTags       []string         `json:"theme_tags_json,omitempty"`
	SectorTags      []string         `json:"sector_tags_json,omitempty"`
	RiskFlags       []string         `json:"risk_flags_json,omitempty"`
	CreatedAt       string           `json:"created_at,omitempty"`
	UpdatedAt       string           `json:"updated_at,omitempty"`
}

type StockSelectionRunEvaluation struct {
	ID              string  `json:"id"`
	RunID           string  `json:"run_id"`
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name,omitempty"`
	HorizonDay      int     `json:"horizon_day"`
	EvaluationScope string  `json:"evaluation_scope"`
	EntryDate       string  `json:"entry_date,omitempty"`
	ExitDate        string  `json:"exit_date,omitempty"`
	EntryPrice      float64 `json:"entry_price"`
	ExitPrice       float64 `json:"exit_price"`
	ReturnPct       float64 `json:"return_pct"`
	ExcessReturnPct float64 `json:"excess_return_pct"`
	MaxDrawdownPct  float64 `json:"max_drawdown_pct"`
	HitFlag         bool    `json:"hit_flag"`
	BenchmarkSymbol string  `json:"benchmark_symbol,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	UpdatedAt       string  `json:"updated_at,omitempty"`
}

type StockSelectionRunCompareItem struct {
	RunID            string   `json:"run_id"`
	TradeDate        string   `json:"trade_date"`
	ProfileID        string   `json:"profile_id"`
	TemplateID       string   `json:"template_id,omitempty"`
	TemplateName     string   `json:"template_name,omitempty"`
	MarketRegime     string   `json:"market_regime,omitempty"`
	Status           string   `json:"status"`
	ReviewStatus     string   `json:"review_status,omitempty"`
	SelectedCount    int      `json:"selected_count"`
	PortfolioSymbols []string `json:"portfolio_symbols,omitempty"`
	AddedSymbols     []string `json:"added_symbols,omitempty"`
	RemovedSymbols   []string `json:"removed_symbols,omitempty"`
}

type StockSelectionRunCompareResult struct {
	BaseRunID string                        `json:"base_run_id,omitempty"`
	Items     []StockSelectionRunCompareItem `json:"items"`
}

type StockSelectionEvaluationLeaderboardItem struct {
	TemplateID        string             `json:"template_id,omitempty"`
	TemplateName      string             `json:"template_name,omitempty"`
	ProfileID         string             `json:"profile_id,omitempty"`
	ProfileName       string             `json:"profile_name,omitempty"`
	MarketRegime      string             `json:"market_regime,omitempty"`
	SampleCount       int                `json:"sample_count"`
	ReturnByHorizon   map[string]float64 `json:"return_by_horizon,omitempty"`
	HitRateByHorizon  map[string]float64 `json:"hit_rate_by_horizon,omitempty"`
	MaxDrawdownPct    float64            `json:"max_drawdown_pct"`
}
