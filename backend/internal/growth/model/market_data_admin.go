package model

type MarketDataQualityLog struct {
	ID            string `json:"id"`
	AssetClass    string `json:"asset_class,omitempty"`
	DataKind      string `json:"data_kind"`
	InstrumentKey string `json:"instrument_key,omitempty"`
	TradeDate     string `json:"trade_date,omitempty"`
	SourceKey     string `json:"source_key,omitempty"`
	Severity      string `json:"severity"`
	IssueCode     string `json:"issue_code"`
	IssueMessage  string `json:"issue_message,omitempty"`
	Payload       string `json:"payload,omitempty"`
	CreatedAt     string `json:"created_at"`
}

type MarketDataQualitySummary struct {
	AssetClass              string `json:"asset_class,omitempty"`
	LookbackHours           int    `json:"lookback_hours"`
	TotalCount              int    `json:"total_count"`
	ErrorCount              int    `json:"error_count"`
	WarnCount               int    `json:"warn_count"`
	InfoCount               int    `json:"info_count"`
	DistinctSourceCount     int    `json:"distinct_source_count"`
	LatestSourceKey         string `json:"latest_source_key,omitempty"`
	LatestSeverity          string `json:"latest_severity,omitempty"`
	LatestIssueCode         string `json:"latest_issue_code,omitempty"`
	LatestIssueMessage      string `json:"latest_issue_message,omitempty"`
	LatestTradeDate         string `json:"latest_trade_date,omitempty"`
	LatestCreatedAt         string `json:"latest_created_at,omitempty"`
	LatestErrorSourceKey    string `json:"latest_error_source_key,omitempty"`
	LatestErrorIssueCode    string `json:"latest_error_issue_code,omitempty"`
	LatestErrorMessage      string `json:"latest_error_message,omitempty"`
	LatestErrorCreatedAt    string `json:"latest_error_created_at,omitempty"`
	StockMasterCoverage     int    `json:"stock_master_coverage,omitempty"`
	StockTruthCoverage      int    `json:"stock_truth_coverage,omitempty"`
	StockDailyBasicCoverage int    `json:"stock_daily_basic_coverage,omitempty"`
	StockMoneyflowCoverage  int    `json:"stock_moneyflow_coverage,omitempty"`
	StockNewsCoverage       int    `json:"stock_news_coverage,omitempty"`
	FallbackSourceSummary   string `json:"fallback_source_summary,omitempty"`
	CanonicalKeyGapCount    int    `json:"canonical_key_gap_count,omitempty"`
	DisplayNameMissingCount int    `json:"display_name_missing_count,omitempty"`
	ListDateMissingCount    int    `json:"list_date_missing_count,omitempty"`
}

type MarketDerivedTruthRebuildResult struct {
	AssetClass          string   `json:"asset_class"`
	TradeDate           string   `json:"trade_date,omitempty"`
	StartDate           string   `json:"start_date,omitempty"`
	EndDate             string   `json:"end_date,omitempty"`
	Days                int      `json:"days"`
	TruthBarCount       int      `json:"truth_bar_count"`
	StockStatusCount    int      `json:"stock_status_count,omitempty"`
	FuturesMappingCount int      `json:"futures_mapping_count,omitempty"`
	Warnings            []string `json:"warnings,omitempty"`
}

type MarketDerivedTruthSummary struct {
	AssetClass          string   `json:"asset_class"`
	SourceKey           string   `json:"source_key,omitempty"`
	IssueCode           string   `json:"issue_code,omitempty"`
	IssueMessage        string   `json:"issue_message,omitempty"`
	TradeDate           string   `json:"trade_date,omitempty"`
	StartDate           string   `json:"start_date,omitempty"`
	EndDate             string   `json:"end_date,omitempty"`
	Days                int      `json:"days"`
	TruthBarCount       int      `json:"truth_bar_count"`
	StockStatusCount    int      `json:"stock_status_count,omitempty"`
	FuturesMappingCount int      `json:"futures_mapping_count,omitempty"`
	Warnings            []string `json:"warnings,omitempty"`
	CreatedAt           string   `json:"created_at,omitempty"`
}

type MarketCoverageSummary struct {
	TotalUniverseCount      int                               `json:"total_universe_count"`
	MasterCoverageCount     int                               `json:"master_coverage_count"`
	QuotesCoverageCount     int                               `json:"quotes_coverage_count"`
	DailyBasicCoverageCount int                               `json:"daily_basic_coverage_count"`
	MoneyflowCoverageCount  int                               `json:"moneyflow_coverage_count"`
	LatestTradeDate         string                            `json:"latest_trade_date,omitempty"`
	FallbackSourceSummary   []MarketCoverageSourceSummaryItem `json:"fallback_source_summary,omitempty"`
	AssetItems              []MarketCoverageSummaryAssetItem  `json:"asset_items,omitempty"`
	CanonicalKeyGapCount    int                               `json:"canonical_key_gap_count"`
	DisplayNameGapCount     int                               `json:"display_name_gap_count"`
	ListDateGapCount        int                               `json:"list_date_gap_count"`
}

type MarketCoverageSummaryAssetItem struct {
	AssetType               string `json:"asset_type"`
	UniverseCount           int    `json:"universe_count"`
	MasterCoverageCount     int    `json:"master_coverage_count"`
	QuotesCoverageCount     int    `json:"quotes_coverage_count"`
	DailyBasicCoverageCount int    `json:"daily_basic_coverage_count"`
	MoneyflowCoverageCount  int    `json:"moneyflow_coverage_count"`
	LatestTradeDate         string `json:"latest_trade_date,omitempty"`
}

type MarketCoverageSourceSummaryItem struct {
	SourceKey string `json:"source_key"`
	Count     int    `json:"count"`
}
