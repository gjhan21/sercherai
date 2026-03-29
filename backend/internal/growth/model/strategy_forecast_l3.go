package model

const (
	StrategyForecastL3TargetTypeStock   = "STOCK"
	StrategyForecastL3TargetTypeFutures = "FUTURES"

	StrategyForecastL3TriggerTypeAdminManual  = "ADMIN_MANUAL"
	StrategyForecastL3TriggerTypeAutoPriority = "AUTO_PRIORITY"
	StrategyForecastL3TriggerTypeUserRequest  = "USER_REQUEST"

	StrategyForecastL3StatusQueued    = "QUEUED"
	StrategyForecastL3StatusRunning   = "RUNNING"
	StrategyForecastL3StatusSucceeded = "SUCCEEDED"
	StrategyForecastL3StatusFailed    = "FAILED"
	StrategyForecastL3StatusCancelled = "CANCELLED"

	StrategyForecastL3EngineLocalSynthesis = "LOCAL_SYNTHESIS"
)

type StrategyForecastL3RunCreateInput struct {
	TargetType     string         `json:"target_type"`
	TargetID       string         `json:"target_id"`
	TargetKey      string         `json:"target_key"`
	TargetLabel    string         `json:"target_label"`
	TriggerType    string         `json:"trigger_type"`
	RequestUserID  string         `json:"request_user_id"`
	OperatorUserID string         `json:"operator_user_id"`
	PriorityScore  float64        `json:"priority_score"`
	Reason         string         `json:"reason"`
	ContextMeta    map[string]any `json:"context_meta,omitempty"`
}

type StrategyForecastL3Run struct {
	ID             string                       `json:"id"`
	TargetType     string                       `json:"target_type"`
	TargetID       string                       `json:"target_id"`
	TargetKey      string                       `json:"target_key"`
	TargetLabel    string                       `json:"target_label"`
	TriggerType    string                       `json:"trigger_type"`
	RequestUserID  string                       `json:"request_user_id,omitempty"`
	OperatorUserID string                       `json:"operator_user_id,omitempty"`
	EngineKey      string                       `json:"engine_key"`
	Status         string                       `json:"status"`
	PriorityScore  float64                      `json:"priority_score"`
	Reason         string                       `json:"reason,omitempty"`
	FailureReason  string                       `json:"failure_reason,omitempty"`
	ContextMeta    map[string]any               `json:"context_meta,omitempty"`
	Summary        StrategyForecastL3Summary    `json:"summary"`
	ReportRef      *StrategyForecastL3ReportRef `json:"report_ref,omitempty"`
	QueuedAt       string                       `json:"queued_at,omitempty"`
	StartedAt      string                       `json:"started_at,omitempty"`
	FinishedAt     string                       `json:"finished_at,omitempty"`
	CancelledAt    string                       `json:"cancelled_at,omitempty"`
	CreatedAt      string                       `json:"created_at"`
	UpdatedAt      string                       `json:"updated_at"`
}

type StrategyForecastL3RunDetail struct {
	Run            StrategyForecastL3Run             `json:"run"`
	Report         *StrategyForecastL3Report         `json:"report,omitempty"`
	Logs           []StrategyForecastL3Log           `json:"logs,omitempty"`
	QualitySummary *StrategyForecastL3QualitySummary `json:"quality_summary,omitempty"`
}

type StrategyForecastL3Report struct {
	ID                   string                               `json:"id"`
	RunID                string                               `json:"run_id"`
	Version              int                                  `json:"version"`
	ExecutiveSummary     string                               `json:"executive_summary"`
	PrimaryScenario      string                               `json:"primary_scenario"`
	AlternativeScenarios []StrategyForecastL3Scenario         `json:"alternative_scenarios,omitempty"`
	TriggerChecklist     []StrategyForecastL3ChecklistItem    `json:"trigger_checklist,omitempty"`
	InvalidationSignals  []string                             `json:"invalidation_signals,omitempty"`
	RoleDisagreements    []StrategyForecastL3RoleDisagreement `json:"role_disagreements,omitempty"`
	ActionGuidance       []string                             `json:"action_guidance,omitempty"`
	MarkdownBody         string                               `json:"markdown_body,omitempty"`
	HTMLBody             string                               `json:"html_body,omitempty"`
	Summary              StrategyForecastL3Summary            `json:"summary"`
	CreatedAt            string                               `json:"created_at"`
	UpdatedAt            string                               `json:"updated_at"`
}

type StrategyForecastL3Scenario struct {
	Name        string  `json:"name"`
	Probability float64 `json:"probability,omitempty"`
	Thesis      string  `json:"thesis,omitempty"`
	Action      string  `json:"action,omitempty"`
}

type StrategyForecastL3ChecklistItem struct {
	Label   string `json:"label"`
	Status  string `json:"status"`
	Note    string `json:"note,omitempty"`
	Trigger string `json:"trigger,omitempty"`
}

type StrategyForecastL3RoleDisagreement struct {
	Role    string `json:"role"`
	Stance  string `json:"stance,omitempty"`
	Summary string `json:"summary"`
	Veto    bool   `json:"veto"`
}

type StrategyForecastL3Log struct {
	ID        string         `json:"id"`
	RunID     string         `json:"run_id"`
	StepKey   string         `json:"step_key"`
	Status    string         `json:"status"`
	Message   string         `json:"message"`
	Payload   map[string]any `json:"payload,omitempty"`
	CreatedAt string         `json:"created_at"`
}

type StrategyForecastL3LearningRecord struct {
	ID                string             `json:"id"`
	RunID             string             `json:"run_id"`
	TargetType        string             `json:"target_type"`
	TargetKey         string             `json:"target_key"`
	ScenarioHit       bool               `json:"scenario_hit"`
	TriggerHit        bool               `json:"trigger_hit"`
	InvalidationEarly bool               `json:"invalidation_early"`
	BiasLabel         string             `json:"bias_label,omitempty"`
	RoleEffectiveness map[string]float64 `json:"role_effectiveness,omitempty"`
	Summary           string             `json:"summary,omitempty"`
	CreatedAt         string             `json:"created_at"`
	UpdatedAt         string             `json:"updated_at"`
}

type StrategyForecastL3QualitySummary struct {
	TargetType             string             `json:"target_type"`
	TotalRuns              int                `json:"total_runs"`
	SucceededRuns          int                `json:"succeeded_runs"`
	ScenarioHitRate        float64            `json:"scenario_hit_rate"`
	TriggerHitRate         float64            `json:"trigger_hit_rate"`
	InvalidationEarlyRate  float64            `json:"invalidation_early_rate"`
	RoleEffectiveness      map[string]float64 `json:"role_effectiveness,omitempty"`
	LastLearningRecordedAt string             `json:"last_learning_recorded_at,omitempty"`
}

type StrategyForecastL3Summary struct {
	RunID            string  `json:"run_id"`
	Status           string  `json:"status"`
	EngineKey        string  `json:"engine_key"`
	TriggerType      string  `json:"trigger_type"`
	TargetType       string  `json:"target_type"`
	TargetKey        string  `json:"target_key"`
	TargetLabel      string  `json:"target_label"`
	ExecutiveSummary string  `json:"executive_summary,omitempty"`
	PrimaryScenario  string  `json:"primary_scenario,omitempty"`
	ActionGuidance   string  `json:"action_guidance,omitempty"`
	ConfidenceLabel  string  `json:"confidence_label,omitempty"`
	PriorityScore    float64 `json:"priority_score,omitempty"`
	GeneratedAt      string  `json:"generated_at,omitempty"`
	ReportAvailable  bool    `json:"report_available"`
}

type StrategyForecastL3ReportRef struct {
	RunID        string `json:"run_id"`
	ReportID     string `json:"report_id"`
	Status       string `json:"status"`
	EngineKey    string `json:"engine_key"`
	GeneratedAt  string `json:"generated_at,omitempty"`
	RequiresVIP  bool   `json:"requires_vip"`
	FullReadable bool   `json:"full_readable"`
}
