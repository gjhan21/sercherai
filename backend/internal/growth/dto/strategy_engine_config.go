package dto

type StrategySeedSetRequest struct {
	Name        string   `json:"name" binding:"required"`
	TargetType  string   `json:"target_type" binding:"required,oneof=STOCK FUTURES"`
	Status      string   `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
	IsDefault   bool     `json:"is_default"`
	Items       []string `json:"items" binding:"required,min=1,dive,required"`
	Description string   `json:"description"`
}

type StrategyAgentProfileRequest struct {
	Name                            string   `json:"name" binding:"required"`
	TargetType                      string   `json:"target_type" binding:"required,oneof=STOCK FUTURES ALL"`
	Status                          string   `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
	IsDefault                       bool     `json:"is_default"`
	EnabledAgents                   []string `json:"enabled_agents" binding:"required,min=1,dive,required"`
	PositiveThreshold               int      `json:"positive_threshold" binding:"required,gte=1,lte=5"`
	NegativeThreshold               int      `json:"negative_threshold" binding:"required,gte=1,lte=5"`
	AllowVeto                       bool     `json:"allow_veto"`
	AllowMockFallbackOnShortHistory bool     `json:"allow_mock_fallback_on_short_history"`
	Description                     string   `json:"description"`
}

type StrategyScenarioTemplateItemRequest struct {
	Scenario       string  `json:"scenario" binding:"required,oneof=bull base bear shock"`
	Label          string  `json:"label" binding:"required"`
	ThesisTemplate string  `json:"thesis_template" binding:"required"`
	Action         string  `json:"action" binding:"required"`
	RiskSignal     string  `json:"risk_signal" binding:"required"`
	ScoreBias      float64 `json:"score_bias"`
}

type StrategyScenarioTemplateRequest struct {
	Name        string                                `json:"name" binding:"required"`
	TargetType  string                                `json:"target_type" binding:"required,oneof=STOCK FUTURES"`
	Status      string                                `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
	IsDefault   bool                                  `json:"is_default"`
	Items       []StrategyScenarioTemplateItemRequest `json:"items" binding:"required,min=1,dive"`
	Description string                                `json:"description"`
}

type StrategyPublishPolicyRequest struct {
	Name                 string `json:"name" binding:"required"`
	TargetType           string `json:"target_type" binding:"required,oneof=STOCK FUTURES ALL"`
	Status               string `json:"status" binding:"required,oneof=ACTIVE DISABLED"`
	IsDefault            bool   `json:"is_default"`
	MaxRiskLevel         string `json:"max_risk_level" binding:"required,oneof=LOW MEDIUM HIGH"`
	MaxWarningCount      int    `json:"max_warning_count" binding:"required,gte=0,lte=20"`
	AllowVetoedPublish   bool   `json:"allow_vetoed_publish"`
	DefaultPublisher     string `json:"default_publisher"`
	OverrideNoteTemplate string `json:"override_note_template"`
	Description          string `json:"description"`
}

type StrategyJobPublishRequest struct {
	Force          bool   `json:"force"`
	OverrideReason string `json:"override_reason"`
}
