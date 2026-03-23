package repo

import (
	"database/sql"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func normalizeMarketProviderFilter(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func isMarketProviderGovernanceSchemaCompatError(err error) bool {
	if err == nil {
		return false
	}
	if isTableNotFoundError(err) {
		return true
	}
	text := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(text, "unknown column") || strings.Contains(text, "error 1054")
}

type marketProviderQualityAggregate struct {
	TotalCount       int
	ErrorCount       int
	WarnCount        int
	LatestObservedAt time.Time
}

type marketProviderLatestIssue struct {
	IssueCode string
	CreatedAt time.Time
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

func (r *MySQLGrowthRepo) AdminListMarketProviderQualityScores(assetClass string, dataKind string, hours int) ([]model.MarketProviderQualityScore, error) {
	if hours <= 0 {
		hours = 24
	}
	normalizedAssetClass := normalizeMarketProviderFilter(assetClass)
	normalizedDataKind := normalizeMarketProviderFilter(dataKind)

	capabilities, err := r.AdminListMarketProviderCapabilities("", normalizedAssetClass, normalizedDataKind)
	if err != nil {
		if isMarketProviderGovernanceSchemaCompatError(err) {
			capabilities = defaultMarketProviderCapabilitiesFiltered("", normalizedAssetClass, normalizedDataKind)
		} else {
			return nil, err
		}
	}

	lookbackStart := time.Now().Add(-time.Duration(hours) * time.Hour)
	aggregates := make(map[string]marketProviderQualityAggregate)
	rows, err := r.db.Query(`
SELECT COALESCE(source_key, '') AS source_key,
       COUNT(*),
       COALESCE(SUM(CASE WHEN severity = 'ERROR' THEN 1 ELSE 0 END), 0),
       COALESCE(SUM(CASE WHEN severity = 'WARN' THEN 1 ELSE 0 END), 0),
       MAX(created_at)
FROM market_data_quality_logs
WHERE asset_class = ? AND data_kind = ? AND created_at >= ?
GROUP BY COALESCE(source_key, '')`, normalizedAssetClass, normalizedDataKind, lookbackStart)
	if err != nil {
		if !isMarketStatusSchemaCompatError(err) {
			return nil, err
		}
	} else {
		defer rows.Close()
		for rows.Next() {
			var (
				sourceKey        sql.NullString
				aggregate        marketProviderQualityAggregate
				latestObservedAt sql.NullTime
			)
			if err := rows.Scan(
				&sourceKey,
				&aggregate.TotalCount,
				&aggregate.ErrorCount,
				&aggregate.WarnCount,
				&latestObservedAt,
			); err != nil {
				return nil, err
			}
			if latestObservedAt.Valid {
				aggregate.LatestObservedAt = latestObservedAt.Time
			}
			aggregates[strings.ToUpper(strings.TrimSpace(sourceKey.String))] = aggregate
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	latestIssues := make(map[string]marketProviderLatestIssue)
	issueRows, err := r.db.Query(`
SELECT source_key, issue_code, created_at
FROM market_data_quality_logs
WHERE asset_class = ? AND data_kind = ? AND created_at >= ?
ORDER BY created_at DESC, id DESC`, normalizedAssetClass, normalizedDataKind, lookbackStart)
	if err != nil {
		if !isMarketStatusSchemaCompatError(err) {
			return nil, err
		}
	} else {
		defer issueRows.Close()
		for issueRows.Next() {
			var (
				sourceKey sql.NullString
				issueCode sql.NullString
				createdAt time.Time
			)
			if err := issueRows.Scan(&sourceKey, &issueCode, &createdAt); err != nil {
				return nil, err
			}
			normalizedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey.String))
			if normalizedSourceKey == "" {
				continue
			}
			if _, exists := latestIssues[normalizedSourceKey]; exists {
				continue
			}
			latestIssues[normalizedSourceKey] = marketProviderLatestIssue{
				IssueCode: strings.ToUpper(strings.TrimSpace(issueCode.String)),
				CreatedAt: createdAt,
			}
		}
		if err := issueRows.Err(); err != nil {
			return nil, err
		}
	}

	items := make([]model.MarketProviderQualityScore, 0, len(capabilities))
	for _, item := range capabilities {
		providerKey := strings.ToUpper(strings.TrimSpace(item.ProviderKey))
		items = append(items, buildMarketProviderQualityScore(item, hours, aggregates[providerKey], latestIssues[providerKey]))
	}
	return items, nil
}

func (r *MySQLGrowthRepo) AdminUpsertMarketProviderRoutingPolicy(policyKey string, item model.MarketProviderRoutingPolicy) (model.MarketProviderRoutingPolicy, error) {
	normalizedPolicyKey := normalizeMarketProviderFilter(policyKey)
	if normalizedPolicyKey == "" {
		return model.MarketProviderRoutingPolicy{}, sql.ErrNoRows
	}
	normalizedItem := model.MarketProviderRoutingPolicy{
		PolicyKey:            normalizedPolicyKey,
		AssetClass:           normalizeMarketProviderFilter(item.AssetClass),
		DataKind:             normalizeMarketProviderFilter(item.DataKind),
		PrimaryProviderKey:   normalizeMarketProviderFilter(item.PrimaryProviderKey),
		FallbackProviderKeys: appendRemainingSourceKeys(normalizeMarketProviderFilter(item.PrimaryProviderKey), item.FallbackProviderKeys),
		FallbackAllowed:      item.FallbackAllowed,
		MockAllowed:          item.MockAllowed,
		QualityThreshold:     roundTo(item.QualityThreshold, 4),
		UpdatedAt:            time.Now().Format(time.RFC3339),
	}
	if normalizedItem.DataKind == "" || normalizedItem.PrimaryProviderKey == "" {
		return model.MarketProviderRoutingPolicy{}, sql.ErrNoRows
	}
	_, err := r.db.Exec(`
INSERT INTO market_provider_routing_policies
  (policy_key, asset_class, data_kind, primary_provider_key, fallback_provider_keys_json, fallback_allowed, mock_allowed, quality_threshold)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  asset_class = VALUES(asset_class),
  data_kind = VALUES(data_kind),
  primary_provider_key = VALUES(primary_provider_key),
  fallback_provider_keys_json = VALUES(fallback_provider_keys_json),
  fallback_allowed = VALUES(fallback_allowed),
  mock_allowed = VALUES(mock_allowed),
  quality_threshold = VALUES(quality_threshold)`,
		normalizedItem.PolicyKey,
		normalizedItem.AssetClass,
		normalizedItem.DataKind,
		normalizedItem.PrimaryProviderKey,
		marshalJSONSilently(normalizedItem.FallbackProviderKeys),
		normalizedItem.FallbackAllowed,
		normalizedItem.MockAllowed,
		normalizedItem.QualityThreshold,
	)
	if err != nil {
		return model.MarketProviderRoutingPolicy{}, err
	}
	return normalizedItem, nil
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

func findDefaultMarketProviderRoutingPolicy(assetClass string, dataKind string) (model.MarketProviderRoutingPolicy, bool) {
	normalizedAssetClass := normalizeMarketProviderFilter(assetClass)
	normalizedDataKind := normalizeMarketProviderFilter(dataKind)
	for _, item := range defaultMarketProviderRoutingPolicies() {
		if normalizeMarketProviderFilter(item.AssetClass) != normalizedAssetClass {
			continue
		}
		if normalizeMarketProviderFilter(item.DataKind) != normalizedDataKind {
			continue
		}
		return item, true
	}
	return model.MarketProviderRoutingPolicy{}, false
}

func defaultMarketProviderCapabilitiesFiltered(providerKey string, assetClass string, dataKind string) []model.MarketProviderCapability {
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
	return filtered
}

func clampMarketProviderScore(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return roundTo(value, 4)
}

func calculateMarketProviderCoverageScore(item model.MarketProviderCapability) float64 {
	score := 0.2
	if item.SupportsSync {
		score += 0.15
	}
	if item.SupportsTruthRebuild {
		score += 0.15
	}
	if item.SupportsContextSeed {
		score += 0.1
	}
	if item.SupportsResearchRun {
		score += 0.1
	}
	if item.SupportsBackfill {
		score += 0.1
	}
	if item.SupportsBatch {
		score += 0.05
	}
	if item.SupportsHistory {
		score += 0.1
	}
	if item.SupportsMetadataEnrichment {
		score += 0.05
	}
	return clampMarketProviderScore(score)
}

func calculateMarketProviderFreshnessScore(now time.Time, hours int, aggregate marketProviderQualityAggregate) float64 {
	if aggregate.TotalCount == 0 || aggregate.LatestObservedAt.IsZero() {
		return 0.75
	}
	if hours <= 0 {
		hours = 24
	}
	ageHours := now.Sub(aggregate.LatestObservedAt).Hours()
	switch {
	case ageHours <= 6:
		return 0.95
	case ageHours <= 24:
		return 0.85
	case ageHours <= float64(hours):
		return 0.7
	case ageHours <= float64(hours*2):
		return 0.55
	default:
		return 0.35
	}
}

func calculateMarketProviderStabilityScore(aggregate marketProviderQualityAggregate) float64 {
	penalty := float64(aggregate.ErrorCount)*0.25 + float64(aggregate.WarnCount)*0.08
	return clampMarketProviderScore(1 - penalty)
}

func calculateMarketProviderTrustScore(item model.MarketProviderCapability, aggregate marketProviderQualityAggregate, latestIssue marketProviderLatestIssue) float64 {
	score := 0.65
	if item.SupportsTruthRebuild {
		score += 0.1
	}
	if item.RequiresAuth {
		score += 0.05
	}
	if item.PriorityWeight >= 90 {
		score += 0.05
	}
	if strings.Contains(latestIssue.IssueCode, "FAILED") {
		score -= 0.15
	}
	score -= float64(aggregate.ErrorCount) * 0.08
	score -= float64(aggregate.WarnCount) * 0.03
	return clampMarketProviderScore(score)
}

func buildMarketProviderGovernanceSuggestion(overall float64, latestIssueCode string, aggregate marketProviderQualityAggregate) string {
	switch {
	case aggregate.ErrorCount >= 2 || overall < 0.55:
		return "建议降级为备源并优先排查连续异常"
	case strings.Contains(latestIssueCode, "FAILED"):
		return "建议保留 fallback，并跟踪最近失败告警"
	case aggregate.WarnCount >= 2 || overall < 0.75:
		return "建议观察质量波动，暂不提升优先级"
	default:
		return "可保持当前优先级"
	}
}

func buildMarketProviderScoreReasons(hours int, aggregate marketProviderQualityAggregate, latestIssueCode string, coverage float64) []string {
	reasons := make([]string, 0, 4)
	if aggregate.TotalCount == 0 {
		reasons = append(reasons, "近窗口暂无质量日志，按中性新鲜度估算")
	} else {
		reasons = append(reasons, "近窗口统计 "+strconv.Itoa(aggregate.ErrorCount)+" 次 ERROR / "+strconv.Itoa(aggregate.WarnCount)+" 次 WARN")
	}
	if latestIssueCode != "" {
		reasons = append(reasons, "最近问题码: "+latestIssueCode)
	}
	reasons = append(reasons, "能力覆盖评分 "+strconv.FormatFloat(coverage, 'f', 2, 64))
	if hours > 0 {
		reasons = append(reasons, "统计窗口 "+strconv.Itoa(hours)+" 小时")
	}
	return reasons
}

func buildMarketProviderQualityScore(item model.MarketProviderCapability, hours int, aggregate marketProviderQualityAggregate, latestIssue marketProviderLatestIssue) model.MarketProviderQualityScore {
	now := time.Now()
	coverageScore := calculateMarketProviderCoverageScore(item)
	freshnessScore := calculateMarketProviderFreshnessScore(now, hours, aggregate)
	stabilityScore := calculateMarketProviderStabilityScore(aggregate)
	trustScore := calculateMarketProviderTrustScore(item, aggregate, latestIssue)
	overallScore := clampMarketProviderScore(freshnessScore*0.25 + coverageScore*0.2 + stabilityScore*0.3 + trustScore*0.25)

	result := model.MarketProviderQualityScore{
		ProviderKey:          item.ProviderKey,
		AssetClass:           item.AssetClass,
		DataKind:             item.DataKind,
		FreshnessScore:       freshnessScore,
		CoverageScore:        coverageScore,
		TrustScore:           trustScore,
		StabilityScore:       stabilityScore,
		OverallScore:         overallScore,
		LatestIssueCode:      latestIssue.IssueCode,
		GovernanceSuggestion: buildMarketProviderGovernanceSuggestion(overallScore, latestIssue.IssueCode, aggregate),
		ScoreReasons:         buildMarketProviderScoreReasons(hours, aggregate, latestIssue.IssueCode, coverageScore),
	}
	if !aggregate.LatestObservedAt.IsZero() {
		result.LatestObservedAt = aggregate.LatestObservedAt.Format(time.RFC3339)
	}
	return result
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

func (r *InMemoryGrowthRepo) AdminListMarketProviderQualityScores(assetClass string, dataKind string, hours int) ([]model.MarketProviderQualityScore, error) {
	if hours <= 0 {
		hours = 24
	}
	capabilities := defaultMarketProviderCapabilitiesFiltered("", assetClass, dataKind)
	items := make([]model.MarketProviderQualityScore, 0, len(capabilities))
	for _, item := range capabilities {
		items = append(items, buildMarketProviderQualityScore(item, hours, marketProviderQualityAggregate{}, marketProviderLatestIssue{}))
	}
	return items, nil
}

func (r *InMemoryGrowthRepo) AdminUpsertMarketProviderRoutingPolicy(policyKey string, item model.MarketProviderRoutingPolicy) (model.MarketProviderRoutingPolicy, error) {
	normalizedPolicyKey := normalizeMarketProviderFilter(policyKey)
	normalizedItem := model.MarketProviderRoutingPolicy{
		PolicyKey:            normalizedPolicyKey,
		AssetClass:           normalizeMarketProviderFilter(item.AssetClass),
		DataKind:             normalizeMarketProviderFilter(item.DataKind),
		PrimaryProviderKey:   normalizeMarketProviderFilter(item.PrimaryProviderKey),
		FallbackProviderKeys: appendRemainingSourceKeys(normalizeMarketProviderFilter(item.PrimaryProviderKey), item.FallbackProviderKeys),
		FallbackAllowed:      item.FallbackAllowed,
		MockAllowed:          item.MockAllowed,
		QualityThreshold:     roundTo(item.QualityThreshold, 4),
		UpdatedAt:            time.Now().Format(time.RFC3339),
	}
	if normalizedItem.PolicyKey == "" || normalizedItem.DataKind == "" || normalizedItem.PrimaryProviderKey == "" {
		return model.MarketProviderRoutingPolicy{}, sql.ErrNoRows
	}
	return normalizedItem, nil
}
