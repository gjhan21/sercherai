package model

type StrategySeedSet struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	TargetType  string   `json:"target_type"`
	Status      string   `json:"status"`
	IsDefault   bool     `json:"is_default"`
	Items       []string `json:"items"`
	Description string   `json:"description,omitempty"`
	UpdatedBy   string   `json:"updated_by,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

type StrategyAgentProfile struct {
	ID                              string   `json:"id"`
	Name                            string   `json:"name"`
	TargetType                      string   `json:"target_type"`
	Status                          string   `json:"status"`
	IsDefault                       bool     `json:"is_default"`
	EnabledAgents                   []string `json:"enabled_agents"`
	PositiveThreshold               int      `json:"positive_threshold"`
	NegativeThreshold               int      `json:"negative_threshold"`
	AllowVeto                       bool     `json:"allow_veto"`
	AllowMockFallbackOnShortHistory bool     `json:"allow_mock_fallback_on_short_history"`
	Description                     string   `json:"description,omitempty"`
	UpdatedBy                       string   `json:"updated_by,omitempty"`
	UpdatedAt                       string   `json:"updated_at,omitempty"`
}

type StrategyScenarioTemplateItem struct {
	Scenario       string  `json:"scenario"`
	Label          string  `json:"label"`
	ThesisTemplate string  `json:"thesis_template"`
	Action         string  `json:"action"`
	RiskSignal     string  `json:"risk_signal"`
	ScoreBias      float64 `json:"score_bias"`
}

type StrategyScenarioTemplate struct {
	ID          string                         `json:"id"`
	Name        string                         `json:"name"`
	TargetType  string                         `json:"target_type"`
	Status      string                         `json:"status"`
	IsDefault   bool                           `json:"is_default"`
	Items       []StrategyScenarioTemplateItem `json:"items"`
	Description string                         `json:"description,omitempty"`
	UpdatedBy   string                         `json:"updated_by,omitempty"`
	UpdatedAt   string                         `json:"updated_at,omitempty"`
}

type StrategyPublishPolicy struct {
	ID                   string `json:"id"`
	Name                 string `json:"name"`
	TargetType           string `json:"target_type"`
	Status               string `json:"status"`
	IsDefault            bool   `json:"is_default"`
	MaxRiskLevel         string `json:"max_risk_level"`
	MaxWarningCount      int    `json:"max_warning_count"`
	AllowVetoedPublish   bool   `json:"allow_vetoed_publish"`
	DefaultPublisher     string `json:"default_publisher"`
	OverrideNoteTemplate string `json:"override_note_template"`
	Description          string `json:"description,omitempty"`
	UpdatedBy            string `json:"updated_by,omitempty"`
	UpdatedAt            string `json:"updated_at,omitempty"`
}
