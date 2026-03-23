package model

type FuturesSelectionProfile struct {
	ID              string                           `json:"id"`
	Name            string                           `json:"name"`
	TemplateID      string                           `json:"template_id,omitempty"`
	TemplateName    string                           `json:"template_name,omitempty"`
	Status          string                           `json:"status"`
	IsDefault       bool                             `json:"is_default"`
	CurrentVersion  int                              `json:"current_version"`
	StyleDefault    string                           `json:"style_default"`
	ContractScope   string                           `json:"contract_scope,omitempty"`
	UniverseConfig  map[string]any                   `json:"universe_config,omitempty"`
	FactorConfig    map[string]any                   `json:"factor_config,omitempty"`
	PortfolioConfig map[string]any                   `json:"portfolio_config,omitempty"`
	PublishConfig   map[string]any                   `json:"publish_config,omitempty"`
	Description     string                           `json:"description,omitempty"`
	UpdatedBy       string                           `json:"updated_by,omitempty"`
	UpdatedAt       string                           `json:"updated_at,omitempty"`
	CreatedAt       string                           `json:"created_at,omitempty"`
	Versions        []FuturesSelectionProfileVersion `json:"versions,omitempty"`
}

type FuturesSelectionProfileVersion struct {
	ID         string         `json:"id"`
	ProfileID  string         `json:"profile_id"`
	VersionNo  int            `json:"version_no"`
	Snapshot   map[string]any `json:"snapshot_json,omitempty"`
	ChangeNote string         `json:"change_note,omitempty"`
	CreatedBy  string         `json:"created_by,omitempty"`
	CreatedAt  string         `json:"created_at,omitempty"`
}

type FuturesSelectionProfileTemplate struct {
	ID                string         `json:"id"`
	TemplateKey       string         `json:"template_key"`
	Name              string         `json:"name"`
	Description       string         `json:"description,omitempty"`
	MarketRegimeBias  string         `json:"market_regime_bias,omitempty"`
	IsDefault         bool           `json:"is_default"`
	Status            string         `json:"status"`
	UniverseDefaults  map[string]any `json:"universe_defaults_json,omitempty"`
	FactorDefaults    map[string]any `json:"factor_defaults_json,omitempty"`
	PortfolioDefaults map[string]any `json:"portfolio_defaults_json,omitempty"`
	PublishDefaults   map[string]any `json:"publish_defaults_json,omitempty"`
	UpdatedBy         string         `json:"updated_by,omitempty"`
	UpdatedAt         string         `json:"updated_at,omitempty"`
	CreatedAt         string         `json:"created_at,omitempty"`
}
