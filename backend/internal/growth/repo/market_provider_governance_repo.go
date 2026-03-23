package repo

import (
	"database/sql"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func normalizeMarketProviderFilter(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func (r *MySQLGrowthRepo) AdminListMarketProviderRegistries(status string) ([]model.MarketProviderRegistry, error) {
	filters := make([]string, 0, 1)
	args := make([]any, 0, 1)
	if normalizedStatus := normalizeMarketProviderFilter(status); normalizedStatus != "" {
		filters = append(filters, "status = ?")
		args = append(args, normalizedStatus)
	}

	query := `
SELECT provider_key, provider_name, provider_type, status, auth_mode, endpoint, timeout_ms,
       COALESCE(CAST(retry_policy_json AS CHAR), ''),
       COALESCE(CAST(health_policy_json AS CHAR), ''),
       COALESCE(CAST(rate_limit_policy_json AS CHAR), ''),
       cost_tier, supports_truth_write, supports_manual_sync, supports_auto_sync, updated_at
FROM market_provider_registry`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY provider_key ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.MarketProviderRegistry, 0)
	for rows.Next() {
		var (
			item             model.MarketProviderRegistry
			authMode         sql.NullString
			endpoint         sql.NullString
			retryPolicyJSON  sql.NullString
			healthPolicyJSON sql.NullString
			rateLimitJSON    sql.NullString
			costTier         sql.NullString
			updatedAt        time.Time
		)
		if err := rows.Scan(
			&item.ProviderKey,
			&item.ProviderName,
			&item.ProviderType,
			&item.Status,
			&authMode,
			&endpoint,
			&item.TimeoutMS,
			&retryPolicyJSON,
			&healthPolicyJSON,
			&rateLimitJSON,
			&costTier,
			&item.SupportsTruthWrite,
			&item.SupportsManualSync,
			&item.SupportsAutoSync,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.AuthMode = authMode.String
		item.Endpoint = endpoint.String
		item.RetryPolicy = parseJSONMap(retryPolicyJSON.String)
		item.HealthPolicy = parseJSONMap(healthPolicyJSON.String)
		item.RateLimitPolicy = parseJSONMap(rateLimitJSON.String)
		item.CostTier = costTier.String
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) AdminListMarketProviderCapabilities(providerKey string, assetClass string, dataKind string) ([]model.MarketProviderCapability, error) {
	filters := make([]string, 0, 3)
	args := make([]any, 0, 3)
	if normalizedProvider := normalizeMarketProviderFilter(providerKey); normalizedProvider != "" {
		filters = append(filters, "provider_key = ?")
		args = append(args, normalizedProvider)
	}
	if normalizedAssetClass := normalizeMarketProviderFilter(assetClass); normalizedAssetClass != "" {
		filters = append(filters, "asset_class = ?")
		args = append(args, normalizedAssetClass)
	}
	if normalizedDataKind := normalizeMarketProviderFilter(dataKind); normalizedDataKind != "" {
		filters = append(filters, "data_kind = ?")
		args = append(args, normalizedDataKind)
	}

	query := `
SELECT provider_key, asset_class, data_kind, supports_sync, supports_truth_rebuild, supports_context_seed, supports_research_run,
       supports_backfill, supports_batch, supports_intraday, supports_history, supports_metadata_enrichment,
       requires_auth, fallback_allowed, priority_weight, updated_at
FROM market_provider_capabilities`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY provider_key ASC, asset_class ASC, data_kind ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.MarketProviderCapability, 0)
	for rows.Next() {
		var (
			item      model.MarketProviderCapability
			updatedAt time.Time
		)
		if err := rows.Scan(
			&item.ProviderKey,
			&item.AssetClass,
			&item.DataKind,
			&item.SupportsSync,
			&item.SupportsTruthRebuild,
			&item.SupportsContextSeed,
			&item.SupportsResearchRun,
			&item.SupportsBackfill,
			&item.SupportsBatch,
			&item.SupportsIntraday,
			&item.SupportsHistory,
			&item.SupportsMetadataEnrichment,
			&item.RequiresAuth,
			&item.FallbackAllowed,
			&item.PriorityWeight,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) AdminListMarketProviderRoutingPolicies(assetClass string, dataKind string) ([]model.MarketProviderRoutingPolicy, error) {
	filters := make([]string, 0, 2)
	args := make([]any, 0, 2)
	if normalizedAssetClass := normalizeMarketProviderFilter(assetClass); normalizedAssetClass != "" {
		filters = append(filters, "asset_class = ?")
		args = append(args, normalizedAssetClass)
	}
	if normalizedDataKind := normalizeMarketProviderFilter(dataKind); normalizedDataKind != "" {
		filters = append(filters, "data_kind = ?")
		args = append(args, normalizedDataKind)
	}

	query := `
SELECT policy_key, asset_class, data_kind, primary_provider_key,
       COALESCE(CAST(fallback_provider_keys_json AS CHAR), ''), fallback_allowed, mock_allowed, quality_threshold, updated_at
FROM market_provider_routing_policies`
	if len(filters) > 0 {
		query += " WHERE " + strings.Join(filters, " AND ")
	}
	query += " ORDER BY asset_class ASC, data_kind ASC, policy_key ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.MarketProviderRoutingPolicy, 0)
	for rows.Next() {
		var (
			item                  model.MarketProviderRoutingPolicy
			fallbackProvidersJSON sql.NullString
			updatedAt             time.Time
		)
		if err := rows.Scan(
			&item.PolicyKey,
			&item.AssetClass,
			&item.DataKind,
			&item.PrimaryProviderKey,
			&fallbackProvidersJSON,
			&item.FallbackAllowed,
			&item.MockAllowed,
			&item.QualityThreshold,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.FallbackProviderKeys = parseJSONStringList(fallbackProvidersJSON.String)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, rows.Err()
}

func defaultMarketProviderRegistries() []model.MarketProviderRegistry {
	return []model.MarketProviderRegistry{
		{
			ProviderKey:        "AKSHARE",
			ProviderName:       "AkShare",
			ProviderType:       "API",
			Status:             "ACTIVE",
			AuthMode:           "NONE",
			Endpoint:           "https://akshare.akfamily.xyz",
			TimeoutMS:          12000,
			RetryPolicy:        map[string]interface{}{"max_retries": 2},
			HealthPolicy:       map[string]interface{}{"timeout_ms": 3000},
			RateLimitPolicy:    map[string]interface{}{"rpm": 90},
			CostTier:           "FREE",
			SupportsTruthWrite: true,
			SupportsManualSync: true,
			SupportsAutoSync:   true,
			UpdatedAt:          time.Now().Format(time.RFC3339),
		},
		{
			ProviderKey:        "MOCK",
			ProviderName:       "Mock Source",
			ProviderType:       "INTERNAL",
			Status:             "ACTIVE",
			AuthMode:           "NONE",
			Endpoint:           "",
			TimeoutMS:          1000,
			RetryPolicy:        map[string]interface{}{"max_retries": 0},
			HealthPolicy:       map[string]interface{}{"timeout_ms": 500},
			RateLimitPolicy:    map[string]interface{}{"rpm": 9999},
			CostTier:           "FREE",
			SupportsTruthWrite: false,
			SupportsManualSync: true,
			SupportsAutoSync:   false,
			UpdatedAt:          time.Now().Format(time.RFC3339),
		},
		{
			ProviderKey:        "MYSELF",
			ProviderName:       "Myself Bridge",
			ProviderType:       "BRIDGE",
			Status:             "ACTIVE",
			AuthMode:           "NONE",
			Endpoint:           "internal://myself",
			TimeoutMS:          8000,
			RetryPolicy:        map[string]interface{}{"max_retries": 1},
			HealthPolicy:       map[string]interface{}{"timeout_ms": 2000},
			RateLimitPolicy:    map[string]interface{}{"rpm": 120},
			CostTier:           "FREE",
			SupportsTruthWrite: true,
			SupportsManualSync: true,
			SupportsAutoSync:   true,
			UpdatedAt:          time.Now().Format(time.RFC3339),
		},
		{
			ProviderKey:        "TICKERMD",
			ProviderName:       "TickerMD",
			ProviderType:       "API",
			Status:             "ACTIVE",
			AuthMode:           "TOKEN",
			Endpoint:           "https://api.tickermd.com",
			TimeoutMS:          12000,
			RetryPolicy:        map[string]interface{}{"max_retries": 2},
			HealthPolicy:       map[string]interface{}{"timeout_ms": 3000},
			RateLimitPolicy:    map[string]interface{}{"rpm": 60},
			CostTier:           "PAID",
			SupportsTruthWrite: true,
			SupportsManualSync: true,
			SupportsAutoSync:   true,
			UpdatedAt:          time.Now().Format(time.RFC3339),
		},
		{
			ProviderKey:        "TUSHARE",
			ProviderName:       "Tushare Pro",
			ProviderType:       "API",
			Status:             "ACTIVE",
			AuthMode:           "TOKEN",
			Endpoint:           "https://api.tushare.pro",
			TimeoutMS:          12000,
			RetryPolicy:        map[string]interface{}{"max_retries": 3},
			HealthPolicy:       map[string]interface{}{"timeout_ms": 3000},
			RateLimitPolicy:    map[string]interface{}{"rpm": 120},
			CostTier:           "PAID",
			SupportsTruthWrite: true,
			SupportsManualSync: true,
			SupportsAutoSync:   true,
			UpdatedAt:          time.Now().Format(time.RFC3339),
		},
	}
}

func defaultMarketProviderCapabilities() []model.MarketProviderCapability {
	now := time.Now().Format(time.RFC3339)
	return []model.MarketProviderCapability{
		{
			ProviderKey:                "TUSHARE",
			AssetClass:                 "STOCK",
			DataKind:                   "DAILY_BARS",
			SupportsSync:               true,
			SupportsTruthRebuild:       true,
			SupportsContextSeed:        true,
			SupportsResearchRun:        true,
			SupportsBackfill:           true,
			SupportsBatch:              true,
			SupportsIntraday:           false,
			SupportsHistory:            true,
			SupportsMetadataEnrichment: false,
			RequiresAuth:               true,
			FallbackAllowed:            true,
			PriorityWeight:             100,
			UpdatedAt:                  now,
		},
		{
			ProviderKey:                "AKSHARE",
			AssetClass:                 "STOCK",
			DataKind:                   "NEWS_ITEMS",
			SupportsSync:               true,
			SupportsTruthRebuild:       false,
			SupportsContextSeed:        true,
			SupportsResearchRun:        true,
			SupportsBackfill:           true,
			SupportsBatch:              true,
			SupportsIntraday:           false,
			SupportsHistory:            true,
			SupportsMetadataEnrichment: true,
			RequiresAuth:               false,
			FallbackAllowed:            true,
			PriorityWeight:             90,
			UpdatedAt:                  now,
		},
		{
			ProviderKey:                "MYSELF",
			AssetClass:                 "FUTURES",
			DataKind:                   "DAILY_BARS",
			SupportsSync:               true,
			SupportsTruthRebuild:       true,
			SupportsContextSeed:        true,
			SupportsResearchRun:        true,
			SupportsBackfill:           true,
			SupportsBatch:              true,
			SupportsIntraday:           false,
			SupportsHistory:            true,
			SupportsMetadataEnrichment: false,
			RequiresAuth:               false,
			FallbackAllowed:            true,
			PriorityWeight:             80,
			UpdatedAt:                  now,
		},
	}
}

func defaultMarketProviderRoutingPolicies() []model.MarketProviderRoutingPolicy {
	now := time.Now().Format(time.RFC3339)
	return []model.MarketProviderRoutingPolicy{
		{
			PolicyKey:            "market.stock.daily",
			AssetClass:           "STOCK",
			DataKind:             "DAILY_BARS",
			PrimaryProviderKey:   "TUSHARE",
			FallbackProviderKeys: []string{"AKSHARE", "TICKERMD"},
			FallbackAllowed:      true,
			MockAllowed:          false,
			QualityThreshold:     0.8,
			UpdatedAt:            now,
		},
		{
			PolicyKey:            "market.futures.daily",
			AssetClass:           "FUTURES",
			DataKind:             "DAILY_BARS",
			PrimaryProviderKey:   "TUSHARE",
			FallbackProviderKeys: []string{"TICKERMD", "AKSHARE"},
			FallbackAllowed:      true,
			MockAllowed:          false,
			QualityThreshold:     0.8,
			UpdatedAt:            now,
		},
		{
			PolicyKey:            "market.news",
			AssetClass:           "",
			DataKind:             "NEWS_ITEMS",
			PrimaryProviderKey:   "AKSHARE",
			FallbackProviderKeys: []string{"TUSHARE"},
			FallbackAllowed:      true,
			MockAllowed:          false,
			QualityThreshold:     0.7,
			UpdatedAt:            now,
		},
	}
}

func (r *InMemoryGrowthRepo) AdminListMarketProviderRegistries(status string) ([]model.MarketProviderRegistry, error) {
	normalizedStatus := normalizeMarketProviderFilter(status)
	items := defaultMarketProviderRegistries()
	if normalizedStatus == "" {
		return items, nil
	}
	filtered := make([]model.MarketProviderRegistry, 0, len(items))
	for _, item := range items {
		if item.Status == normalizedStatus {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func (r *InMemoryGrowthRepo) AdminListMarketProviderCapabilities(providerKey string, assetClass string, dataKind string) ([]model.MarketProviderCapability, error) {
	normalizedProviderKey := normalizeMarketProviderFilter(providerKey)
	normalizedAssetClass := normalizeMarketProviderFilter(assetClass)
	normalizedDataKind := normalizeMarketProviderFilter(dataKind)
	items := defaultMarketProviderCapabilities()
	filtered := make([]model.MarketProviderCapability, 0, len(items))
	for _, item := range items {
		if normalizedProviderKey != "" && item.ProviderKey != normalizedProviderKey {
			continue
		}
		if normalizedAssetClass != "" && item.AssetClass != normalizedAssetClass {
			continue
		}
		if normalizedDataKind != "" && item.DataKind != normalizedDataKind {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}

func (r *InMemoryGrowthRepo) AdminListMarketProviderRoutingPolicies(assetClass string, dataKind string) ([]model.MarketProviderRoutingPolicy, error) {
	normalizedAssetClass := normalizeMarketProviderFilter(assetClass)
	normalizedDataKind := normalizeMarketProviderFilter(dataKind)
	items := defaultMarketProviderRoutingPolicies()
	filtered := make([]model.MarketProviderRoutingPolicy, 0, len(items))
	for _, item := range items {
		if normalizedAssetClass != "" && item.AssetClass != normalizedAssetClass {
			continue
		}
		if normalizedDataKind != "" && item.DataKind != normalizedDataKind {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, nil
}
