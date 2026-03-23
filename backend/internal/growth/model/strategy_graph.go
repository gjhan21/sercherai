package model

type StrategyGraphEntity struct {
	EntityType  string         `json:"entity_type"`
	EntityKey   string         `json:"entity_key"`
	Label       string         `json:"label"`
	AssetDomain string         `json:"asset_domain,omitempty"`
	Tags        []string       `json:"tags,omitempty"`
	Meta        map[string]any `json:"meta,omitempty"`
}

type StrategyGraphRelation struct {
	RelationType string         `json:"relation_type"`
	SourceType   string         `json:"source_type"`
	SourceKey    string         `json:"source_key"`
	TargetType   string         `json:"target_type"`
	TargetKey    string         `json:"target_key"`
	Strength     float64        `json:"strength"`
	Note         string         `json:"note,omitempty"`
	Meta         map[string]any `json:"meta,omitempty"`
}

type StrategyGraphSnapshot struct {
	SnapshotID      string                  `json:"snapshot_id"`
	RunID           string                  `json:"run_id"`
	AssetDomain     string                  `json:"asset_domain"`
	TradeDate       string                  `json:"trade_date"`
	Summary         string                  `json:"summary,omitempty"`
	RelatedEntities []StrategyGraphEntity   `json:"related_entities,omitempty"`
	Entities        []StrategyGraphEntity   `json:"entities,omitempty"`
	Relations       []StrategyGraphRelation `json:"relations,omitempty"`
	Meta            map[string]any          `json:"meta,omitempty"`
	CreatedAt       string                  `json:"created_at,omitempty"`
}

type StrategyGraphSubgraph struct {
	Entity             *StrategyGraphEntity    `json:"entity,omitempty"`
	Entities           []StrategyGraphEntity   `json:"entities,omitempty"`
	Relations          []StrategyGraphRelation `json:"relations,omitempty"`
	MatchedSnapshotIDs []string                `json:"matched_snapshot_ids,omitempty"`
	Backend            string                  `json:"backend,omitempty"`
}

type StrategyGraphSubgraphQuery struct {
	EntityType  string `json:"entity_type"`
	EntityKey   string `json:"entity_key"`
	Depth       int    `json:"depth"`
	AssetDomain string `json:"asset_domain"`
}
