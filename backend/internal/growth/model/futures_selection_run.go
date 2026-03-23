package model

type FuturesSelectionRun struct {
	RunID                string                         `json:"run_id"`
	TradeDate            string                         `json:"trade_date"`
	JobID                string                         `json:"job_id,omitempty"`
	ProfileID            string                         `json:"profile_id"`
	ProfileVersion       int                            `json:"profile_version"`
	TemplateID           string                         `json:"template_id,omitempty"`
	TemplateName         string                         `json:"template_name,omitempty"`
	MarketRegime         string                         `json:"market_regime,omitempty"`
	Style                string                         `json:"style,omitempty"`
	ContractScope        string                         `json:"contract_scope,omitempty"`
	Status               string                         `json:"status"`
	ResultSummary        string                         `json:"result_summary,omitempty"`
	WarningCount         int                            `json:"warning_count"`
	WarningMessages      []string                       `json:"warning_messages,omitempty"`
	UniverseCount        int                            `json:"universe_count"`
	CandidateCount       int                            `json:"candidate_count"`
	SelectedCount        int                            `json:"selected_count"`
	PublishCount         int                            `json:"publish_count"`
	ContextMeta          map[string]any                 `json:"context_meta,omitempty"`
	TemplateSnapshot     map[string]any                 `json:"template_snapshot,omitempty"`
	CompareSummary       map[string]any                 `json:"compare_summary,omitempty"`
	StageCounts          map[string]int                 `json:"stage_counts,omitempty"`
	StageDurationsMS     map[string]int64               `json:"stage_durations_ms,omitempty"`
	StageLogs            []FuturesSelectionRunStageLog  `json:"stage_logs,omitempty"`
	Review               *FuturesSelectionPublishReview `json:"review,omitempty"`
	ReviewStatus         string                         `json:"review_status,omitempty"`
	LatestPublishID      string                         `json:"latest_publish_id,omitempty"`
	LatestPublishVersion int                            `json:"latest_publish_version,omitempty"`
	LatestPublishAt      string                         `json:"latest_publish_at,omitempty"`
	LatestPublishMode    string                         `json:"latest_publish_mode,omitempty"`
	LatestPublishSource  string                         `json:"latest_publish_source,omitempty"`
	StartedAt            string                         `json:"started_at,omitempty"`
	CompletedAt          string                         `json:"completed_at,omitempty"`
	CreatedBy            string                         `json:"created_by,omitempty"`
	CreatedAt            string                         `json:"created_at,omitempty"`
	UpdatedAt            string                         `json:"updated_at,omitempty"`
}

type FuturesSelectionRunStageLog struct {
	ID              string         `json:"id"`
	RunID           string         `json:"run_id"`
	StageKey        string         `json:"stage_key"`
	StageOrder      int            `json:"stage_order"`
	Status          string         `json:"status"`
	InputCount      int            `json:"input_count"`
	OutputCount     int            `json:"output_count"`
	DurationMS      int64          `json:"duration_ms"`
	DetailMessage   string         `json:"detail_message,omitempty"`
	PayloadSnapshot map[string]any `json:"payload_snapshot,omitempty"`
	CreatedAt       string         `json:"created_at,omitempty"`
	UpdatedAt       string         `json:"updated_at,omitempty"`
}

type FuturesSelectionCandidateSnapshot struct {
	ID                  string         `json:"id"`
	RunID               string         `json:"run_id"`
	Contract            string         `json:"contract"`
	Name                string         `json:"name"`
	Stage               string         `json:"stage"`
	Score               float64        `json:"score"`
	Direction           string         `json:"direction"`
	RiskLevel           string         `json:"risk_level"`
	Selected            bool           `json:"selected"`
	Rank                int            `json:"rank"`
	ReasonSummary       string         `json:"reason_summary,omitempty"`
	EvidenceSummary     string         `json:"evidence_summary,omitempty"`
	PortfolioRole       string         `json:"portfolio_role,omitempty"`
	PreviousPublishDiff map[string]any `json:"previous_publish_diff,omitempty"`
	EvaluationStatus    string         `json:"evaluation_status,omitempty"`
	RiskSummary         string         `json:"risk_summary,omitempty"`
	FactorBreakdownJSON map[string]any `json:"factor_breakdown_json,omitempty"`
	CreatedAt           string         `json:"created_at,omitempty"`
	UpdatedAt           string         `json:"updated_at,omitempty"`
}

type FuturesSelectionRunEvidence struct {
	ID              string           `json:"id"`
	RunID           string           `json:"run_id"`
	Contract        string           `json:"contract"`
	Name            string           `json:"name,omitempty"`
	Stage           string           `json:"stage"`
	PortfolioRole   string           `json:"portfolio_role,omitempty"`
	EvidenceSummary string           `json:"evidence_summary,omitempty"`
	EvidenceCards   []map[string]any `json:"evidence_cards_json,omitempty"`
	PositiveReasons []string         `json:"positive_reasons_json,omitempty"`
	VetoReasons     []string         `json:"veto_reasons_json,omitempty"`
	RiskFlags       []string         `json:"risk_flags_json,omitempty"`
	RelatedEntities []map[string]any `json:"related_entities_json,omitempty"`
	CreatedAt       string           `json:"created_at,omitempty"`
	UpdatedAt       string           `json:"updated_at,omitempty"`
}

type FuturesSelectionRunEvaluation struct {
	ID              string  `json:"id"`
	RunID           string  `json:"run_id"`
	Contract        string  `json:"contract"`
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

type FuturesSelectionRunCompareItem struct {
	RunID              string   `json:"run_id"`
	TradeDate          string   `json:"trade_date"`
	ProfileID          string   `json:"profile_id"`
	TemplateID         string   `json:"template_id,omitempty"`
	TemplateName       string   `json:"template_name,omitempty"`
	MarketRegime       string   `json:"market_regime,omitempty"`
	Status             string   `json:"status"`
	ReviewStatus       string   `json:"review_status,omitempty"`
	SelectedCount      int      `json:"selected_count"`
	PortfolioContracts []string `json:"portfolio_contracts,omitempty"`
	AddedContracts     []string `json:"added_contracts,omitempty"`
	RemovedContracts   []string `json:"removed_contracts,omitempty"`
}

type FuturesSelectionRunCompareResult struct {
	BaseRunID string                           `json:"base_run_id,omitempty"`
	Items     []FuturesSelectionRunCompareItem `json:"items"`
}

type FuturesSelectionPortfolioEntry struct {
	ID                  string         `json:"id"`
	RunID               string         `json:"run_id"`
	Contract            string         `json:"contract"`
	Name                string         `json:"name"`
	Rank                int            `json:"rank"`
	Score               float64        `json:"score"`
	Direction           string         `json:"direction"`
	RiskLevel           string         `json:"risk_level"`
	PositionRange       string         `json:"position_range,omitempty"`
	ReasonSummary       string         `json:"reason_summary,omitempty"`
	EvidenceSummary     string         `json:"evidence_summary,omitempty"`
	PortfolioRole       string         `json:"portfolio_role,omitempty"`
	PreviousPublishDiff map[string]any `json:"previous_publish_diff,omitempty"`
	EvaluationStatus    string         `json:"evaluation_status,omitempty"`
	RiskSummary         string         `json:"risk_summary,omitempty"`
	FactorBreakdownJSON map[string]any `json:"factor_breakdown_json,omitempty"`
	CreatedAt           string         `json:"created_at,omitempty"`
	UpdatedAt           string         `json:"updated_at,omitempty"`
}

type FuturesSelectionPublishReview struct {
	ID                        string           `json:"id"`
	RunID                     string           `json:"run_id"`
	ReviewStatus              string           `json:"review_status"`
	Reviewer                  string           `json:"reviewer,omitempty"`
	ReviewNote                string           `json:"review_note,omitempty"`
	OverrideReason            string           `json:"override_reason,omitempty"`
	PublishID                 string           `json:"publish_id,omitempty"`
	PublishVersion            int              `json:"publish_version"`
	ApprovedAt                string           `json:"approved_at,omitempty"`
	RejectedAt                string           `json:"rejected_at,omitempty"`
	PublishedContractSnapshot []map[string]any `json:"published_contract_snapshot,omitempty"`
	CreatedAt                 string           `json:"created_at,omitempty"`
	UpdatedAt                 string           `json:"updated_at,omitempty"`
}

type FuturesSelectionRunCreateRequest struct {
	TradeDate                string `json:"trade_date"`
	ProfileID                string `json:"profile_id"`
	TemplateID               string `json:"template_id,omitempty"`
	CompareWithLastPublished bool   `json:"compare_with_last_published"`
	DryRun                   bool   `json:"dry_run"`
}
