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

type StrategyExplanationRelatedEntity struct {
	EntityType  string         `json:"entity_type"`
	EntityKey   string         `json:"entity_key"`
	Label       string         `json:"label"`
	AssetDomain string         `json:"asset_domain"`
	Tags        []string       `json:"tags"`
	Meta        map[string]any `json:"meta"`
}

type StrategyExplanationRelatedEvent struct {
	ClusterID      string   `json:"cluster_id"`
	Title          string   `json:"title"`
	EventType      string   `json:"event_type"`
	PrimarySymbol  string   `json:"primary_symbol"`
	TopicLabel     string   `json:"topic_label"`
	SectorLabel    string   `json:"sector_label"`
	ReviewPriority string   `json:"review_priority"`
	ReviewNote     string   `json:"review_note"`
	PublishedAt    string   `json:"published_at"`
	Tags           []string `json:"tags"`
}

type StrategyExplanationMemoryFeedbackItem struct {
	Title      string `json:"title"`
	Level      string `json:"level"`
	Detail     string `json:"detail"`
	Suggestion string `json:"suggestion"`
	Source     string `json:"source"`
}

type StrategyExplanationMemoryFeedback struct {
	Summary        string                                  `json:"summary"`
	Suggestions    []string                                `json:"suggestions"`
	FailureSignals []string                                `json:"failure_signals"`
	Items          []StrategyExplanationMemoryFeedbackItem `json:"items"`
}

type StrategyVersionDiff struct {
	ComparePublishID   string   `json:"compare_publish_id,omitempty"`
	CompareVersion     int      `json:"compare_version,omitempty"`
	Added              []string `json:"added,omitempty"`
	Removed            []string `json:"removed,omitempty"`
	Promoted           []string `json:"promoted,omitempty"`
	DowngradeReasons   []string `json:"downgrade_reasons,omitempty"`
	CurrentAssetChange string   `json:"current_asset_change,omitempty"`
	Summary            string   `json:"summary,omitempty"`
}

type StrategyClientExplanation struct {
	SeedSummary        string                             `json:"seed_summary"`
	SeedHighlights     []string                           `json:"seed_highlights"`
	GraphSummary       string                             `json:"graph_summary"`
	GraphSnapshotID    string                             `json:"graph_snapshot_id"`
	ConsensusSummary   string                             `json:"consensus_summary"`
	Simulations        []StrategyExplanationSimulation    `json:"simulations"`
	AgentOpinions      []StrategyExplanationAgentOpinion  `json:"agent_opinions"`
	RiskFlags          []string                           `json:"risk_flags"`
	Invalidations      []string                           `json:"invalidations"`
	ConfidenceReason   string                             `json:"confidence_reason"`
	MarketRegime       string                             `json:"market_regime"`
	EvidenceCards      []StrategyExplanationEvidenceCard  `json:"evidence_cards"`
	PortfolioRole      string                             `json:"portfolio_role"`
	RiskBoundary       string                             `json:"risk_boundary"`
	ThemeTags          []string                           `json:"theme_tags"`
	SectorTags         []string                           `json:"sector_tags"`
	SupplyChainNotes   []string                           `json:"supply_chain_notes"`
	StructureSummary   string                             `json:"structure_factor_summary"`
	InventorySummary   string                             `json:"inventory_factor_summary"`
	RelatedEntities    []StrategyExplanationRelatedEntity `json:"related_entities"`
	RelatedEvents      []StrategyExplanationRelatedEvent  `json:"related_events"`
	EventEvidenceCards []StrategyExplanationEvidenceCard  `json:"event_evidence_cards"`
	MemoryFeedback     StrategyExplanationMemoryFeedback  `json:"memory_feedback"`
	EvaluationMeta     map[string]any                     `json:"evaluation_meta"`
	VersionDiff        StrategyVersionDiff                `json:"version_diff,omitempty"`
	WorkloadSummary    StrategyWorkloadSummary            `json:"workload_summary"`
	StrategyVersion    string                             `json:"strategy_version"`
	PublishID          string                             `json:"publish_id"`
	JobID              string                             `json:"job_id"`
	TradeDate          string                             `json:"trade_date"`
	PublishVersion     int                                `json:"publish_version"`
	GeneratedAt        string                             `json:"generated_at"`
}

type StrategyVersionHistoryItem struct {
	PublishID        string                             `json:"publish_id"`
	JobID            string                             `json:"job_id"`
	TradeDate        string                             `json:"trade_date"`
	PublishVersion   int                                `json:"publish_version"`
	CreatedAt        string                             `json:"created_at"`
	StrategyVersion  string                             `json:"strategy_version"`
	ReasonSummary    string                             `json:"reason_summary"`
	ConfidenceReason string                             `json:"confidence_reason"`
	ConsensusSummary string                             `json:"consensus_summary"`
	GraphSummary     string                             `json:"graph_summary"`
	GraphSnapshotID  string                             `json:"graph_snapshot_id"`
	MarketRegime     string                             `json:"market_regime"`
	PortfolioRole    string                             `json:"portfolio_role"`
	RiskBoundary     string                             `json:"risk_boundary"`
	ThemeTags        []string                           `json:"theme_tags"`
	SectorTags       []string                           `json:"sector_tags"`
	RelatedEntities  []StrategyExplanationRelatedEntity `json:"related_entities"`
	MemoryFeedback   StrategyExplanationMemoryFeedback  `json:"memory_feedback"`
	RiskFlags        []string                           `json:"risk_flags"`
	Invalidations    []string                           `json:"invalidations"`
	EvaluationMeta   map[string]any                     `json:"evaluation_meta"`
	VersionDiff      StrategyVersionDiff                `json:"version_diff,omitempty"`
	GeneratedAt      string                             `json:"generated_at"`
}
