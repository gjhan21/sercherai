package repo

import "sercherai/backend/internal/growth/model"

func buildStockRelationshipSnapshot(ctx strategyEngineAssetContext) model.StrategyExplanationRelationshipSnapshot {
	assetKey := firstNonEmpty(asString(ctx.asset["symbol"]), asString(ctx.asset["asset_key"]))
	nodes := []model.StrategyExplanationRelationshipNode{}

	if name := asString(ctx.asset["name"]); name != "" {
		nodes = append(nodes, model.StrategyExplanationRelationshipNode{
			Type:  "Company",
			Key:   assetKey,
			Label: name,
		})
	}
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
	for _, item := range sliceOfMaps(ctx.simulation["related_events"]) {
		label := firstNonEmpty(asString(item["title"]), asString(item["event_type"]))
		if label == "" {
			continue
		}
		nodes = append(nodes, model.StrategyExplanationRelationshipNode{
			Type:  firstNonEmpty(asString(item["event_type"]), "Event"),
			Key:   firstNonEmpty(asString(item["cluster_id"]), label),
			Label: label,
		})
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

func buildFuturesRelationshipSnapshot(ctx strategyEngineAssetContext) model.StrategyExplanationRelationshipSnapshot {
	assetKey := firstNonEmpty(asString(ctx.asset["contract"]), asString(ctx.asset["asset_key"]))
	nodes := []model.StrategyExplanationRelationshipNode{
		{
			Type:  "Contract",
			Key:   assetKey,
			Label: firstNonEmpty(asString(ctx.asset["name"]), assetKey),
		},
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

func compactRelationshipNodes(items []model.StrategyExplanationRelationshipNode) []model.StrategyExplanationRelationshipNode {
	if len(items) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	result := make([]model.StrategyExplanationRelationshipNode, 0, len(items))
	for _, item := range items {
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
