package repo

import "sercherai/backend/internal/growth/model"

func buildStockRelationshipSnapshot(
	ctx strategyEngineAssetContext,
	relatedEntities []model.StrategyExplanationRelatedEntity,
) model.StrategyExplanationRelationshipSnapshot {
	assetKey := firstNonEmpty(asString(ctx.asset["symbol"]), asString(ctx.asset["asset_key"]))
	nodes := []model.StrategyExplanationRelationshipNode{}

	nodes = append(nodes, model.StrategyExplanationRelationshipNode{
		Type:  "Company",
		Key:   assetKey,
		Label: firstNonEmpty(asString(ctx.asset["name"]), assetKey),
	})
	for _, sector := range compactStrings(anySliceToStrings(ctx.asset["sector_tags"])) {
		nodes = append(nodes, model.StrategyExplanationRelationshipNode{
			Type:  "Sector",
			Key:   sector,
			Label: sector,
		})
	}
	for _, theme := range compactStrings(anySliceToStrings(ctx.asset["theme_tags"])) {
		nodes = append(nodes, model.StrategyExplanationRelationshipNode{
			Type:  "Theme",
			Key:   theme,
			Label: theme,
		})
	}
	nodes = append(nodes, relationshipNodesFromRelatedEntities(relatedEntities)...)
	for _, item := range sliceOfMaps(ctx.simulation["related_events"]) {
		nodes = append(nodes, relationshipNodeFromEvent(item))
	}

	nodes = compactRelationshipNodes(nodes)
	return model.StrategyExplanationRelationshipSnapshot{
		AssetKey:          assetKey,
		AssetType:         "stock",
		RelationshipCount: len(nodes),
		Nodes:             nodes,
		Edges:             compactRelationshipEdges(nil),
	}
}

func buildFuturesRelationshipSnapshot(
	ctx strategyEngineAssetContext,
	relatedEntities []model.StrategyExplanationRelatedEntity,
	inventorySummary string,
	structureSummary string,
) model.StrategyExplanationRelationshipSnapshot {
	assetKey := firstNonEmpty(asString(ctx.asset["contract"]), asString(ctx.asset["asset_key"]))
	nodes := []model.StrategyExplanationRelationshipNode{
		{
			Type:  "Contract",
			Key:   assetKey,
			Label: firstNonEmpty(asString(ctx.asset["name"]), assetKey),
		},
	}
	nodes = append(nodes, relationshipNodesFromRelatedEntities(relatedEntities)...)
	for _, item := range sliceOfMaps(ctx.simulation["related_events"]) {
		nodes = append(nodes, relationshipNodeFromEvent(item))
	}
	if inventorySummary != "" {
		nodes = append(nodes, model.StrategyExplanationRelationshipNode{
			Type:  "InventorySignal",
			Key:   inventorySummary,
			Label: "库存画像",
			Meta:  map[string]any{"summary": inventorySummary},
		})
	}
	if structureSummary != "" {
		nodes = append(nodes, model.StrategyExplanationRelationshipNode{
			Type:  "StructureSignal",
			Key:   structureSummary,
			Label: "结构联动",
			Meta:  map[string]any{"summary": structureSummary},
		})
	}
	nodes = compactRelationshipNodes(nodes)
	return model.StrategyExplanationRelationshipSnapshot{
		AssetKey:          assetKey,
		AssetType:         "futures",
		RelationshipCount: len(nodes),
		Nodes:             nodes,
		Edges:             compactRelationshipEdges(nil),
	}
}

func relationshipNodesFromRelatedEntities(items []model.StrategyExplanationRelatedEntity) []model.StrategyExplanationRelationshipNode {
	nodes := make([]model.StrategyExplanationRelationshipNode, 0, len(items))
	for _, item := range items {
		label := firstNonEmpty(item.Label, item.EntityKey)
		if label == "" {
			continue
		}
		nodes = append(nodes, model.StrategyExplanationRelationshipNode{
			Type:  firstNonEmpty(item.EntityType, "Entity"),
			Key:   firstNonEmpty(item.EntityKey, label),
			Label: label,
			Meta:  item.Meta,
		})
	}
	return nodes
}

func relationshipNodeFromEvent(item map[string]any) model.StrategyExplanationRelationshipNode {
	label := firstNonEmpty(asString(item["title"]), asString(item["event_type"]))
	if label == "" {
		return model.StrategyExplanationRelationshipNode{}
	}
	return model.StrategyExplanationRelationshipNode{
		Type:  firstNonEmpty(asString(item["event_type"]), "Event"),
		Key:   firstNonEmpty(asString(item["cluster_id"]), label),
		Label: label,
	}
}

func compactRelationshipNodes(items []model.StrategyExplanationRelationshipNode) []model.StrategyExplanationRelationshipNode {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	result := make([]model.StrategyExplanationRelationshipNode, 0, len(items))
	for _, item := range items {
		if item.Type == "" && item.Key == "" && item.Label == "" {
			continue
		}
		key := item.Type + "::" + firstNonEmpty(item.Key, item.Label)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func compactRelationshipEdges(items []model.StrategyExplanationRelationshipEdge) []model.StrategyExplanationRelationshipEdge {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	result := make([]model.StrategyExplanationRelationshipEdge, 0, len(items))
	for _, item := range items {
		key := item.From + "::" + item.To + "::" + item.Label
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func anySliceToStrings(value any) []string {
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text := asString(item); text != "" {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}
