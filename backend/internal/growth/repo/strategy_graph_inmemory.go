package repo

import "sercherai/backend/internal/growth/model"

func (r *InMemoryGrowthRepo) AdminGetStrategyGraphSnapshot(snapshotID string) (model.StrategyGraphSnapshot, error) {
	return model.StrategyGraphSnapshot{
		SnapshotID:  snapshotID,
		RunID:       "ssr_demo_graph",
		AssetDomain: "stock",
		TradeDate:   "2026-03-22",
		Summary:     "内存模式图快照示例：机器人与算力共振。",
		RelatedEntities: []model.StrategyGraphEntity{
			{EntityType: "ConceptTheme", EntityKey: "ROBOTICS", Label: "机器人", AssetDomain: "stock"},
			{EntityType: "ConceptTheme", EntityKey: "AI_COMPUTE", Label: "算力", AssetDomain: "stock"},
		},
		Entities: []model.StrategyGraphEntity{
			{EntityType: "Stock", EntityKey: "SZ300024", Label: "机器人龙头", AssetDomain: "stock"},
			{EntityType: "ConceptTheme", EntityKey: "ROBOTICS", Label: "机器人", AssetDomain: "stock"},
			{EntityType: "ConceptTheme", EntityKey: "AI_COMPUTE", Label: "算力", AssetDomain: "stock"},
		},
		Relations: []model.StrategyGraphRelation{
			{RelationType: "BELONGS_TO", SourceType: "Stock", SourceKey: "SZ300024", TargetType: "ConceptTheme", TargetKey: "ROBOTICS", Strength: 0.88},
			{RelationType: "CONFIRMS_SIGNAL", SourceType: "ConceptTheme", SourceKey: "ROBOTICS", TargetType: "ConceptTheme", TargetKey: "AI_COMPUTE", Strength: 0.73},
		},
		Meta:      map[string]any{"backend": "inmemory"},
		CreatedAt: "2026-03-22T00:00:00Z",
	}, nil
}

func (r *InMemoryGrowthRepo) AdminQueryStrategyGraphSubgraph(query model.StrategyGraphSubgraphQuery) (model.StrategyGraphSubgraph, error) {
	entity := &model.StrategyGraphEntity{
		EntityType:  query.EntityType,
		EntityKey:   query.EntityKey,
		Label:       query.EntityKey,
		AssetDomain: query.AssetDomain,
	}
	return model.StrategyGraphSubgraph{
		Entity:             entity,
		Entities:           []model.StrategyGraphEntity{*entity},
		Relations:          []model.StrategyGraphRelation{},
		MatchedSnapshotIDs: []string{"gss_inmemory_demo"},
		Backend:            "inmemory",
	}, nil
}
