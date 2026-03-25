package repo

import (
	"strings"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) SyncReviewedStockEventGraph(cluster model.StockEventCluster) error {
	if r.strategyGraph == nil {
		return nil
	}
	payload := buildReviewedStockEventGraphPayload(cluster)
	_, err := r.strategyGraph.writeReviewedEvent(payload)
	return err
}

func (r *InMemoryGrowthRepo) SyncReviewedStockEventGraph(cluster model.StockEventCluster) error {
	return nil
}

func buildReviewedStockEventGraphPayload(cluster model.StockEventCluster) model.StrategyGraphReviewedEventWriteRequest {
	payload := model.StrategyGraphReviewedEventWriteRequest{
		ClusterID: strings.TrimSpace(cluster.ID),
		Approved:  strings.ToUpper(strings.TrimSpace(cluster.ReviewStatus)) == "APPROVED",
		TradeDate: dateOnly(cluster.PublishedAt),
		Summary:   strings.TrimSpace(cluster.Summary),
		Meta: map[string]any{
			"title":          strings.TrimSpace(cluster.Title),
			"event_type":     strings.TrimSpace(cluster.EventType),
			"review_status":  strings.TrimSpace(cluster.ReviewStatus),
			"primary_symbol": strings.TrimSpace(cluster.PrimarySymbol),
		},
	}
	if !payload.Approved {
		return payload
	}

	rootKey := strings.TrimSpace(cluster.ID)
	rootLabel := strings.TrimSpace(cluster.Title)
	if rootLabel == "" {
		rootLabel = rootKey
	}
	entities := []model.StrategyGraphEntity{
		{
			EntityType:  "StockEvent",
			EntityKey:   rootKey,
			Label:       rootLabel,
			AssetDomain: "stock",
			Tags:        compactStrings([]string{cluster.EventType, cluster.ReviewStatus}),
			Meta: map[string]any{
				"summary":         cluster.Summary,
				"news_count":      cluster.NewsCount,
				"confidence":      cluster.Confidence,
				"review_status":   cluster.ReviewStatus,
				"review_priority": cluster.Metadata["review_priority"],
			},
		},
	}
	relations := make([]model.StrategyGraphRelation, 0, 4)

	seenEntityKeys := map[string]struct{}{
		"StockEvent|" + rootKey: {},
	}
	appendEntity := func(entityType string, entityKey string, label string, tags []string, meta map[string]any) {
		entityKey = strings.TrimSpace(entityKey)
		if entityKey == "" {
			return
		}
		key := entityType + "|" + entityKey
		if _, ok := seenEntityKeys[key]; ok {
			return
		}
		seenEntityKeys[key] = struct{}{}
		entities = append(entities, model.StrategyGraphEntity{
			EntityType:  entityType,
			EntityKey:   entityKey,
			Label:       strings.TrimSpace(label),
			AssetDomain: "stock",
			Tags:        tags,
			Meta:        meta,
		})
	}

	symbols := make([]string, 0, len(cluster.Entities)+1)
	if strings.TrimSpace(cluster.PrimarySymbol) != "" {
		symbols = append(symbols, cluster.PrimarySymbol)
	}
	for _, entity := range cluster.Entities {
		if strings.ToUpper(strings.TrimSpace(entity.EntityType)) == "COMPANY" && strings.TrimSpace(entity.Symbol) != "" {
			symbols = append(symbols, entity.Symbol)
		}
	}
	for _, symbol := range normalizeStockSymbolList(symbols) {
		appendEntity("Stock", symbol, symbol, []string{"reviewed-event"}, map[string]any{"symbol": symbol})
		relations = append(relations, model.StrategyGraphRelation{
			RelationType: "AFFECTS",
			SourceType:   "StockEvent",
			SourceKey:    rootKey,
			TargetType:   "Stock",
			TargetKey:    symbol,
			Strength:     0.9,
			Meta:         map[string]any{"cluster_id": cluster.ID},
		})
	}
	if topic := strings.TrimSpace(cluster.TopicLabel); topic != "" {
		topicKey := "topic:" + topic
		appendEntity("Topic", topicKey, topic, []string{"reviewed-event-topic"}, map[string]any{"topic_label": topic})
		relations = append(relations, model.StrategyGraphRelation{
			RelationType: "BELONGS_TO_TOPIC",
			SourceType:   "StockEvent",
			SourceKey:    rootKey,
			TargetType:   "Topic",
			TargetKey:    topicKey,
			Strength:     0.82,
			Meta:         map[string]any{"cluster_id": cluster.ID},
		})
	}
	if sector := strings.TrimSpace(cluster.SectorLabel); sector != "" {
		sectorKey := "sector:" + sector
		appendEntity("Sector", sectorKey, sector, []string{"reviewed-event-sector"}, map[string]any{"sector_label": sector})
		relations = append(relations, model.StrategyGraphRelation{
			RelationType: "BELONGS_TO_SECTOR",
			SourceType:   "StockEvent",
			SourceKey:    rootKey,
			TargetType:   "Sector",
			TargetKey:    sectorKey,
			Strength:     0.78,
			Meta:         map[string]any{"cluster_id": cluster.ID},
		})
	}

	payload.Entities = entities
	payload.Relations = relations
	return payload
}
