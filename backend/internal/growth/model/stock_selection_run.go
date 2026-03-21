package model

type StockSelectionRun struct {
	RunID                string                       `json:"run_id"`
	TradeDate            string                       `json:"trade_date"`
	JobID                string                       `json:"job_id,omitempty"`
	ProfileID            string                       `json:"profile_id"`
	ProfileVersion       int                          `json:"profile_version"`
	TemplateID           string                       `json:"template_id,omitempty"`
	TemplateName         string                       `json:"template_name,omitempty"`
	MarketRegime         string                       `json:"market_regime,omitempty"`
	SelectionMode        string                       `json:"selection_mode"`
	UniverseScope        string                       `json:"universe_scope,omitempty"`
	Status               string                       `json:"status"`
	ResultSummary        string                       `json:"result_summary,omitempty"`
	WarningCount         int                          `json:"warning_count"`
	WarningMessages      []string                     `json:"warning_messages,omitempty"`
	UniverseCount        int                          `json:"universe_count"`
	SeedCount            int                          `json:"seed_count"`
	CandidateCount       int                          `json:"candidate_count"`
	SelectedCount        int                          `json:"selected_count"`
	PublishCount         int                          `json:"publish_count"`
	ContextMeta          map[string]any               `json:"context_meta,omitempty"`
	TemplateSnapshot     map[string]any               `json:"template_snapshot,omitempty"`
	CompareSummary       map[string]any               `json:"compare_summary,omitempty"`
	StageCounts          map[string]int               `json:"stage_counts,omitempty"`
	StageDurationsMS     map[string]int64             `json:"stage_durations_ms,omitempty"`
	StageLogs            []StockSelectionRunStageLog  `json:"stage_logs,omitempty"`
	Review               *StockSelectionPublishReview `json:"review,omitempty"`
	ReviewStatus         string                       `json:"review_status,omitempty"`
	LatestPublishID      string                       `json:"latest_publish_id,omitempty"`
	LatestPublishVersion int                          `json:"latest_publish_version,omitempty"`
	LatestPublishAt      string                       `json:"latest_publish_at,omitempty"`
	LatestPublishMode    string                       `json:"latest_publish_mode,omitempty"`
	LatestPublishSource  string                       `json:"latest_publish_source,omitempty"`
	StartedAt            string                       `json:"started_at,omitempty"`
	CompletedAt          string                       `json:"completed_at,omitempty"`
	CreatedBy            string                       `json:"created_by,omitempty"`
	CreatedAt            string                       `json:"created_at,omitempty"`
	UpdatedAt            string                       `json:"updated_at,omitempty"`
}

type StockSelectionRunStageLog struct {
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

type StockSelectionCandidateSnapshot struct {
	ID                  string         `json:"id"`
	RunID               string         `json:"run_id"`
	Symbol              string         `json:"symbol"`
	Name                string         `json:"name"`
	Stage               string         `json:"stage"`
	QuantScore          float64        `json:"quant_score"`
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

type StockSelectionPortfolioEntry struct {
	ID                  string         `json:"id"`
	RunID               string         `json:"run_id"`
	Symbol              string         `json:"symbol"`
	Name                string         `json:"name"`
	Rank                int            `json:"rank"`
	QuantScore          float64        `json:"quant_score"`
	RiskLevel           string         `json:"risk_level"`
	WeightSuggestion    string         `json:"weight_suggestion,omitempty"`
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

type StockSelectionPublishReview struct {
	ID             string `json:"id"`
	RunID          string `json:"run_id"`
	ReviewStatus   string `json:"review_status"`
	Reviewer       string `json:"reviewer,omitempty"`
	ReviewNote     string `json:"review_note,omitempty"`
	OverrideReason string `json:"override_reason,omitempty"`
	PublishID      string `json:"publish_id,omitempty"`
	PublishVersion int    `json:"publish_version"`
	ApprovedAt     string `json:"approved_at,omitempty"`
	RejectedAt     string `json:"rejected_at,omitempty"`
	PublishedPortfolioSnapshot []map[string]any `json:"published_portfolio_snapshot,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}
