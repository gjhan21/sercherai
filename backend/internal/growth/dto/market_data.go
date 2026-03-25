package dto

type StockMarketDataSyncRequest struct {
	SourceKey string   `json:"source_key"`
	Symbols   []string `json:"symbols"`
	Days      int      `json:"days" binding:"omitempty,gte=1,lte=365"`
}

type StockMarketDataBackfillRequest struct {
	SourceKey string   `json:"source_key"`
	Symbols   []string `json:"symbols"`
	Days      int      `json:"days" binding:"omitempty,gte=1,lte=365"`
}

type FuturesQuoteSyncRequest struct {
	SourceKey string   `json:"source_key"`
	Contracts []string `json:"contracts"`
	Days      int      `json:"days" binding:"omitempty,gte=5,lte=365"`
}

type FuturesInventorySyncRequest struct {
	SourceKey string   `json:"source_key"`
	Symbols   []string `json:"symbols"`
	Days      int      `json:"days" binding:"omitempty,gte=5,lte=365"`
}

type MarketNewsSyncRequest struct {
	SourceKey string   `json:"source_key"`
	Symbols   []string `json:"symbols"`
	Days      int      `json:"days" binding:"omitempty,gte=1,lte=90"`
	Limit     int      `json:"limit" binding:"omitempty,gte=1,lte=500"`
}

type MarketDerivedTruthRebuildRequest struct {
	TradeDate string `json:"trade_date"`
	Days      int    `json:"days" binding:"omitempty,gte=1,lte=365"`
}

type MarketDataBackfillRequest struct {
	RunType               string   `json:"run_type" binding:"required,oneof=FULL INCREMENTAL REBUILD_ONLY"`
	AssetScope            []string `json:"asset_scope" binding:"required,min=1,dive,oneof=STOCK INDEX ETF LOF CBOND"`
	SourceKey             string   `json:"source_key"`
	TradeDateFrom         string   `json:"trade_date_from"`
	TradeDateTo           string   `json:"trade_date_to"`
	BatchSize             int      `json:"batch_size" binding:"omitempty,gte=20,lte=500"`
	Stages                []string `json:"stages"`
	ForceRefreshUniverse  bool     `json:"force_refresh_universe"`
	RebuildTruthAfterSync bool     `json:"rebuild_truth_after_sync"`
}

type MarketDataBackfillRetryRequest struct {
	RetryMode string   `json:"retry_mode" binding:"required,oneof=FAILED_ONLY FROM_STAGE"`
	Stage     string   `json:"stage"`
	AssetType string   `json:"asset_type"`
	BatchKeys []string `json:"batch_keys"`
}

type MarketDataSyncRequest struct {
	SourceKey          string   `json:"source_key"`
	AssetScope         []string `json:"asset_scope"`
	UniverseSnapshotID string   `json:"universe_snapshot_id"`
	Symbols            []string `json:"symbols"`
	TradeDateFrom      string   `json:"trade_date_from"`
	TradeDateTo        string   `json:"trade_date_to"`
	Days               int      `json:"days" binding:"omitempty,gte=1,lte=3650"`
	BatchSize          int      `json:"batch_size" binding:"omitempty,gte=20,lte=500"`
}
