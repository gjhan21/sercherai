package model

type StrategyExplanationScenario struct {
	Scenario        string  `json:"scenario"`
	Thesis          string  `json:"thesis"`
	ScoreAdjustment float64 `json:"score_adjustment"`
	Action          string  `json:"action"`
	RiskSignal      string  `json:"risk_signal"`
}

type StrategyExplanationAgentOpinion struct {
	Agent      string  `json:"agent"`
	Stance     string  `json:"stance"`
	Confidence float64 `json:"confidence"`
	Summary    string  `json:"summary"`
	Veto       bool    `json:"veto"`
}

type StrategyExplanationSimulation struct {
	AssetKey        string                            `json:"asset_key"`
	AssetType       string                            `json:"asset_type"`
	Scenarios       []StrategyExplanationScenario     `json:"scenarios"`
	Agents          []StrategyExplanationAgentOpinion `json:"agents"`
	ConsensusAction string                            `json:"consensus_action"`
	Vetoed          bool                              `json:"vetoed"`
	VetoReason      string                            `json:"veto_reason"`
}

type StrategyWorkloadSummary struct {
	SeedCount      int      `json:"seed_count"`
	CandidateCount int      `json:"candidate_count"`
	SelectedCount  int      `json:"selected_count"`
	AgentCount     int      `json:"agent_count"`
	ScenarioCount  int      `json:"scenario_count"`
	FilterSteps    []string `json:"filter_steps"`
}

type StrategyExplanationEvidenceCard struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

type StrategyClientExplanation struct {
	SeedSummary      string                            `json:"seed_summary"`
	SeedHighlights   []string                          `json:"seed_highlights"`
	GraphSummary     string                            `json:"graph_summary"`
	ConsensusSummary string                            `json:"consensus_summary"`
	Simulations      []StrategyExplanationSimulation   `json:"simulations"`
	AgentOpinions    []StrategyExplanationAgentOpinion `json:"agent_opinions"`
	RiskFlags        []string                          `json:"risk_flags"`
	Invalidations    []string                          `json:"invalidations"`
	ConfidenceReason string                            `json:"confidence_reason"`
	MarketRegime     string                            `json:"market_regime"`
	EvidenceCards    []StrategyExplanationEvidenceCard `json:"evidence_cards"`
	PortfolioRole    string                            `json:"portfolio_role"`
	RiskBoundary     string                            `json:"risk_boundary"`
	ThemeTags        []string                          `json:"theme_tags"`
	SectorTags       []string                          `json:"sector_tags"`
	EvaluationMeta   map[string]any                    `json:"evaluation_meta"`
	WorkloadSummary  StrategyWorkloadSummary           `json:"workload_summary"`
	StrategyVersion  string                            `json:"strategy_version"`
	PublishID        string                            `json:"publish_id"`
	JobID            string                            `json:"job_id"`
	TradeDate        string                            `json:"trade_date"`
	PublishVersion   int                               `json:"publish_version"`
	GeneratedAt      string                            `json:"generated_at"`
}

type StrategyVersionHistoryItem struct {
	PublishID        string   `json:"publish_id"`
	JobID            string   `json:"job_id"`
	TradeDate        string   `json:"trade_date"`
	PublishVersion   int      `json:"publish_version"`
	CreatedAt        string   `json:"created_at"`
	StrategyVersion  string   `json:"strategy_version"`
	ReasonSummary    string   `json:"reason_summary"`
	ConfidenceReason string   `json:"confidence_reason"`
	ConsensusSummary string   `json:"consensus_summary"`
	MarketRegime     string   `json:"market_regime"`
	PortfolioRole    string   `json:"portfolio_role"`
	RiskBoundary     string   `json:"risk_boundary"`
	ThemeTags        []string `json:"theme_tags"`
	SectorTags       []string `json:"sector_tags"`
	RiskFlags        []string `json:"risk_flags"`
	Invalidations    []string `json:"invalidations"`
	EvaluationMeta   map[string]any `json:"evaluation_meta"`
	GeneratedAt      string   `json:"generated_at"`
}
