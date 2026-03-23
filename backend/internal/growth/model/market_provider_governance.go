package model

type MarketProviderRegistry struct {
	ProviderKey        string                 `json:"provider_key"`
	ProviderName       string                 `json:"provider_name"`
	ProviderType       string                 `json:"provider_type"`
	Status             string                 `json:"status"`
	AuthMode           string                 `json:"auth_mode,omitempty"`
	Endpoint           string                 `json:"endpoint,omitempty"`
	TimeoutMS          int                    `json:"timeout_ms,omitempty"`
	RetryPolicy        map[string]interface{} `json:"retry_policy,omitempty"`
	HealthPolicy       map[string]interface{} `json:"health_policy,omitempty"`
	RateLimitPolicy    map[string]interface{} `json:"rate_limit_policy,omitempty"`
	CostTier           string                 `json:"cost_tier,omitempty"`
	SupportsTruthWrite bool                   `json:"supports_truth_write"`
	SupportsManualSync bool                   `json:"supports_manual_sync"`
	SupportsAutoSync   bool                   `json:"supports_auto_sync"`
	UpdatedAt          string                 `json:"updated_at,omitempty"`
}

type MarketProviderCapability struct {
	ProviderKey                string `json:"provider_key"`
	AssetClass                 string `json:"asset_class"`
	DataKind                   string `json:"data_kind"`
	SupportsSync               bool   `json:"supports_sync"`
	SupportsTruthRebuild       bool   `json:"supports_truth_rebuild"`
	SupportsContextSeed        bool   `json:"supports_context_seed"`
	SupportsResearchRun        bool   `json:"supports_research_run"`
	SupportsBackfill           bool   `json:"supports_backfill"`
	SupportsBatch              bool   `json:"supports_batch"`
	SupportsIntraday           bool   `json:"supports_intraday"`
	SupportsHistory            bool   `json:"supports_history"`
	SupportsMetadataEnrichment bool   `json:"supports_metadata_enrichment"`
	RequiresAuth               bool   `json:"requires_auth"`
	FallbackAllowed            bool   `json:"fallback_allowed"`
	PriorityWeight             int    `json:"priority_weight"`
	UpdatedAt                  string `json:"updated_at,omitempty"`
}

type MarketProviderRoutingPolicy struct {
	PolicyKey            string   `json:"policy_key"`
	AssetClass           string   `json:"asset_class"`
	DataKind             string   `json:"data_kind"`
	PrimaryProviderKey   string   `json:"primary_provider_key"`
	FallbackProviderKeys []string `json:"fallback_provider_keys,omitempty"`
	FallbackAllowed      bool     `json:"fallback_allowed"`
	MockAllowed          bool     `json:"mock_allowed"`
	QualityThreshold     float64  `json:"quality_threshold,omitempty"`
	UpdatedAt            string   `json:"updated_at,omitempty"`
}
