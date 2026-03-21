package model

type StockSelectionProfile struct {
	ID                   string                         `json:"id"`
	Name                 string                         `json:"name"`
	TemplateID           string                         `json:"template_id,omitempty"`
	TemplateName         string                         `json:"template_name,omitempty"`
	Status               string                         `json:"status"`
	IsDefault            bool                           `json:"is_default"`
	CurrentVersion       int                            `json:"current_version"`
	SelectionModeDefault string                         `json:"selection_mode_default"`
	UniverseScope        string                         `json:"universe_scope,omitempty"`
	UniverseConfig       map[string]any                 `json:"universe_config,omitempty"`
	SeedMiningConfig     map[string]any                 `json:"seed_mining_config,omitempty"`
	FactorConfig         map[string]any                 `json:"factor_config,omitempty"`
	PortfolioConfig      map[string]any                 `json:"portfolio_config,omitempty"`
	PublishConfig        map[string]any                 `json:"publish_config,omitempty"`
	Description          string                         `json:"description,omitempty"`
	UpdatedBy            string                         `json:"updated_by,omitempty"`
	UpdatedAt            string                         `json:"updated_at,omitempty"`
	CreatedAt            string                         `json:"created_at,omitempty"`
	Versions             []StockSelectionProfileVersion `json:"versions,omitempty"`
}

type StockSelectionProfileVersion struct {
	ID         string         `json:"id"`
	ProfileID  string         `json:"profile_id"`
	VersionNo  int            `json:"version_no"`
	Snapshot   map[string]any `json:"snapshot_json,omitempty"`
	ChangeNote string         `json:"change_note,omitempty"`
	CreatedBy  string         `json:"created_by,omitempty"`
	CreatedAt  string         `json:"created_at,omitempty"`
}
