package repo

import (
	"errors"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

const marketProviderRegistryListQueryPattern = `(?s)SELECT provider_key, provider_name, provider_type, status, auth_mode, endpoint, timeout_ms,\s*COALESCE\(CAST\(retry_policy_json AS CHAR\), ''\),\s*COALESCE\(CAST\(health_policy_json AS CHAR\), ''\),\s*COALESCE\(CAST\(rate_limit_policy_json AS CHAR\), ''\),\s*cost_tier, supports_truth_write, supports_manual_sync, supports_auto_sync, updated_at\s+FROM market_provider_registry`
const marketProviderCapabilityListQueryPattern = `(?s)SELECT provider_key, asset_class, data_kind, supports_sync, supports_truth_rebuild, supports_context_seed, supports_research_run, supports_backfill, supports_batch, supports_intraday, supports_history, supports_metadata_enrichment, requires_auth, fallback_allowed, priority_weight, updated_at\s+FROM market_provider_capabilities`
const marketProviderRoutingPolicyListQueryPattern = `(?s)SELECT policy_key, asset_class, data_kind, primary_provider_key,\s*COALESCE\(CAST\(fallback_provider_keys_json AS CHAR\), ''\), fallback_allowed, mock_allowed, quality_threshold, updated_at\s+FROM market_provider_routing_policies`
const marketProviderQualityAggregateQueryPattern = `(?s)SELECT COALESCE\(source_key, ''\) AS source_key,\s*COUNT\(\*\),\s*COALESCE\(SUM\(CASE WHEN severity = 'ERROR' THEN 1 ELSE 0 END\), 0\),\s*COALESCE\(SUM\(CASE WHEN severity = 'WARN' THEN 1 ELSE 0 END\), 0\),\s*MAX\(created_at\)\s+FROM market_data_quality_logs`
const marketProviderQualityLatestIssueQueryPattern = `(?s)SELECT source_key, issue_code, created_at\s+FROM market_data_quality_logs`

func TestAdminListMarketProviderRegistriesReadsRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	updatedAt := time.Date(2026, 3, 23, 9, 0, 0, 0, time.Local)

	mock.ExpectQuery(marketProviderRegistryListQueryPattern).
		WithArgs("ACTIVE").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider_key",
			"provider_name",
			"provider_type",
			"status",
			"auth_mode",
			"endpoint",
			"timeout_ms",
			"retry_policy_json",
			"health_policy_json",
			"rate_limit_policy_json",
			"cost_tier",
			"supports_truth_write",
			"supports_manual_sync",
			"supports_auto_sync",
			"updated_at",
		}).AddRow(
			"TUSHARE",
			"Tushare Pro",
			"API",
			"ACTIVE",
			"TOKEN",
			"https://api.tushare.pro",
			12000,
			`{"max_retries":3}`,
			`{"timeout_ms":3000}`,
			`{"rpm":120}`,
			"PAID",
			true,
			true,
			true,
			updatedAt,
		))

	items, err := repo.AdminListMarketProviderRegistries("ACTIVE")
	if err != nil {
		t.Fatalf("AdminListMarketProviderRegistries returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].ProviderKey != "TUSHARE" || items[0].ProviderName != "Tushare Pro" {
		t.Fatalf("unexpected registry item: %+v", items[0])
	}
	if items[0].RetryPolicy["max_retries"] != float64(3) {
		t.Fatalf("expected retry policy to be parsed, got %+v", items[0].RetryPolicy)
	}
	if items[0].UpdatedAt == "" {
		t.Fatalf("expected updated_at to be set, got %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminListMarketProviderCapabilitiesAppliesFilters(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	updatedAt := time.Date(2026, 3, 23, 9, 5, 0, 0, time.Local)

	mock.ExpectQuery(marketProviderCapabilityListQueryPattern).
		WithArgs("TUSHARE", "STOCK", "DAILY_BARS").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider_key",
			"asset_class",
			"data_kind",
			"supports_sync",
			"supports_truth_rebuild",
			"supports_context_seed",
			"supports_research_run",
			"supports_backfill",
			"supports_batch",
			"supports_intraday",
			"supports_history",
			"supports_metadata_enrichment",
			"requires_auth",
			"fallback_allowed",
			"priority_weight",
			"updated_at",
		}).AddRow(
			"TUSHARE",
			"STOCK",
			"DAILY_BARS",
			true,
			true,
			true,
			true,
			true,
			true,
			false,
			true,
			false,
			true,
			true,
			100,
			updatedAt,
		))

	items, err := repo.AdminListMarketProviderCapabilities("TUSHARE", "STOCK", "DAILY_BARS")
	if err != nil {
		t.Fatalf("AdminListMarketProviderCapabilities returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if !items[0].SupportsSync || !items[0].SupportsTruthRebuild || items[0].SupportsIntraday {
		t.Fatalf("unexpected capability flags: %+v", items[0])
	}
	if items[0].PriorityWeight != 100 {
		t.Fatalf("expected priority weight 100, got %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminListMarketProviderCapabilitiesFallsBackWhenTableMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(marketProviderCapabilityListQueryPattern).
		WithArgs("STOCK", "DAILY_BARS").
		WillReturnError(errors.New("Error 1146 (42S02): Table 'sercherai.market_provider_capabilities' doesn't exist"))

	items, err := repo.AdminListMarketProviderCapabilities("", "STOCK", "DAILY_BARS")
	if err != nil {
		t.Fatalf("expected schema-compat fallback, got error: %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected default capabilities on fallback")
	}
}

func TestAdminListMarketProviderRoutingPoliciesParsesFallbackProviders(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	updatedAt := time.Date(2026, 3, 23, 9, 10, 0, 0, time.Local)

	mock.ExpectQuery(marketProviderRoutingPolicyListQueryPattern).
		WithArgs("STOCK", "DAILY_BARS").
		WillReturnRows(sqlmock.NewRows([]string{
			"policy_key",
			"asset_class",
			"data_kind",
			"primary_provider_key",
			"fallback_provider_keys_json",
			"fallback_allowed",
			"mock_allowed",
			"quality_threshold",
			"updated_at",
		}).AddRow(
			"market.stock.daily",
			"STOCK",
			"DAILY_BARS",
			"TUSHARE",
			`["AKSHARE","TICKERMD"]`,
			true,
			false,
			0.8,
			updatedAt,
		))

	items, err := repo.AdminListMarketProviderRoutingPolicies("STOCK", "DAILY_BARS")
	if err != nil {
		t.Fatalf("AdminListMarketProviderRoutingPolicies returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if items[0].PrimaryProviderKey != "TUSHARE" {
		t.Fatalf("unexpected primary provider: %+v", items[0])
	}
	if len(items[0].FallbackProviderKeys) != 2 || items[0].FallbackProviderKeys[0] != "AKSHARE" {
		t.Fatalf("unexpected fallback providers: %+v", items[0].FallbackProviderKeys)
	}
	if !items[0].FallbackAllowed || items[0].MockAllowed {
		t.Fatalf("unexpected routing flags: %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestAdminListMarketProviderRoutingPoliciesFallsBackWhenTableMissing(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(marketProviderRoutingPolicyListQueryPattern).
		WillReturnError(errors.New("Error 1146 (42S02): Table 'sercherai.market_provider_routing_policies' doesn't exist"))

	items, err := repo.AdminListMarketProviderRoutingPolicies("", "")
	if err != nil {
		t.Fatalf("expected schema-compat fallback, got error: %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected default routing policies on fallback")
	}
}

func TestAdminListMarketProviderQualityScoresBuildsGovernanceSuggestions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	updatedAt := time.Date(2026, 3, 23, 9, 5, 0, 0, time.Local)
	latestIssueAt := time.Date(2026, 3, 23, 8, 45, 0, 0, time.Local)

	mock.ExpectQuery(marketProviderCapabilityListQueryPattern).
		WithArgs("STOCK", "DAILY_BARS").
		WillReturnRows(sqlmock.NewRows([]string{
			"provider_key",
			"asset_class",
			"data_kind",
			"supports_sync",
			"supports_truth_rebuild",
			"supports_context_seed",
			"supports_research_run",
			"supports_backfill",
			"supports_batch",
			"supports_intraday",
			"supports_history",
			"supports_metadata_enrichment",
			"requires_auth",
			"fallback_allowed",
			"priority_weight",
			"updated_at",
		}).AddRow(
			"TUSHARE",
			"STOCK",
			"DAILY_BARS",
			true,
			true,
			true,
			true,
			true,
			true,
			false,
			true,
			false,
			true,
			true,
			100,
			updatedAt,
		))

	mock.ExpectQuery(marketProviderQualityAggregateQueryPattern).
		WithArgs("STOCK", "DAILY_BARS", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"source_key",
			"total_count",
			"error_count",
			"warn_count",
			"latest_created_at",
		}).AddRow(
			"TUSHARE",
			3,
			1,
			1,
			latestIssueAt,
		))

	mock.ExpectQuery(marketProviderQualityLatestIssueQueryPattern).
		WithArgs("STOCK", "DAILY_BARS", sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"source_key",
			"issue_code",
			"created_at",
		}).AddRow(
			"TUSHARE",
			"SOURCE_FETCH_FAILED",
			latestIssueAt,
		))

	items, err := repo.AdminListMarketProviderQualityScores("stock", "daily_bars", 24)
	if err != nil {
		t.Fatalf("AdminListMarketProviderQualityScores returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 provider score, got %d", len(items))
	}
	if items[0].ProviderKey != "TUSHARE" || items[0].AssetClass != "STOCK" || items[0].DataKind != "DAILY_BARS" {
		t.Fatalf("unexpected provider quality item: %+v", items[0])
	}
	if items[0].FreshnessScore <= 0 || items[0].StabilityScore <= 0 || items[0].TrustScore <= 0 || items[0].OverallScore <= 0 {
		t.Fatalf("expected positive quality scores, got %+v", items[0])
	}
	if items[0].LatestIssueCode != "SOURCE_FETCH_FAILED" {
		t.Fatalf("expected latest issue code SOURCE_FETCH_FAILED, got %+v", items[0])
	}
	if items[0].GovernanceSuggestion == "" {
		t.Fatalf("expected governance suggestion, got %+v", items[0])
	}
	if len(items[0].ScoreReasons) == 0 {
		t.Fatalf("expected score reasons, got %+v", items[0])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}
