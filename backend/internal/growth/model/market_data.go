package model

type MarketDailyBar struct {
	ID              string  `json:"id"`
	AssetClass      string  `json:"asset_class"`
	InstrumentKey   string  `json:"instrument_key"`
	ExternalSymbol  string  `json:"external_symbol,omitempty"`
	TradeDate       string  `json:"trade_date"`
	OpenPrice       float64 `json:"open_price"`
	HighPrice       float64 `json:"high_price"`
	LowPrice        float64 `json:"low_price"`
	ClosePrice      float64 `json:"close_price"`
	PrevClosePrice  float64 `json:"prev_close_price"`
	SettlePrice     float64 `json:"settle_price"`
	PrevSettlePrice float64 `json:"prev_settle_price"`
	Volume          int64   `json:"volume"`
	Turnover        float64 `json:"turnover"`
	OpenInterest    float64 `json:"open_interest"`
	SourceKey       string  `json:"source_key"`
	FetchedAt       string  `json:"fetched_at,omitempty"`
	CreatedAt       string  `json:"created_at,omitempty"`
	UpdatedAt       string  `json:"updated_at,omitempty"`
}

type MarketDailyBarTruth struct {
	ID                string  `json:"id"`
	AssetClass        string  `json:"asset_class"`
	InstrumentKey     string  `json:"instrument_key"`
	TradeDate         string  `json:"trade_date"`
	SelectedSourceKey string  `json:"selected_source_key"`
	ExternalSymbol    string  `json:"external_symbol,omitempty"`
	OpenPrice         float64 `json:"open_price"`
	HighPrice         float64 `json:"high_price"`
	LowPrice          float64 `json:"low_price"`
	ClosePrice        float64 `json:"close_price"`
	PrevClosePrice    float64 `json:"prev_close_price"`
	SettlePrice       float64 `json:"settle_price"`
	PrevSettlePrice   float64 `json:"prev_settle_price"`
	Volume            int64   `json:"volume"`
	Turnover          float64 `json:"turnover"`
	OpenInterest      float64 `json:"open_interest"`
	UpdatedAt         string  `json:"updated_at,omitempty"`
}

type MarketNewsItem struct {
	ID            string   `json:"id"`
	SourceKey     string   `json:"source_key"`
	ExternalID    string   `json:"external_id"`
	NewsType      string   `json:"news_type"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary,omitempty"`
	Content       string   `json:"content,omitempty"`
	URL           string   `json:"url,omitempty"`
	PrimarySymbol string   `json:"primary_symbol,omitempty"`
	Symbols       []string `json:"symbols,omitempty"`
	PublishedAt   string   `json:"published_at"`
	CreatedAt     string   `json:"created_at,omitempty"`
	UpdatedAt     string   `json:"updated_at,omitempty"`
}

type FuturesInventorySnapshot struct {
	ID             string  `json:"id"`
	Symbol         string  `json:"symbol"`
	TradeDate      string  `json:"trade_date"`
	FuturesName    string  `json:"futures_name,omitempty"`
	Warehouse      string  `json:"warehouse,omitempty"`
	WarehouseID    string  `json:"warehouse_id,omitempty"`
	Area           string  `json:"area,omitempty"`
	Brand          string  `json:"brand,omitempty"`
	Place          string  `json:"place,omitempty"`
	Grade          string  `json:"grade,omitempty"`
	Unit           string  `json:"unit,omitempty"`
	ReceiptVolume  float64 `json:"receipt_volume"`
	PreviousVolume float64 `json:"previous_volume"`
	ChangeVolume   float64 `json:"change_volume"`
	SourceKey      string  `json:"source_key"`
	FetchedAt      string  `json:"fetched_at,omitempty"`
	CreatedAt      string  `json:"created_at,omitempty"`
	UpdatedAt      string  `json:"updated_at,omitempty"`
}

type MarketSourceSnapshot struct {
	ID             string `json:"id"`
	SourceKey      string `json:"source_key"`
	AssetClass     string `json:"asset_class,omitempty"`
	DataKind       string `json:"data_kind"`
	InstrumentKey  string `json:"instrument_key,omitempty"`
	ExternalSymbol string `json:"external_symbol,omitempty"`
	Status         string `json:"status"`
	ErrorMessage   string `json:"error_message,omitempty"`
	FetchedAt      string `json:"fetched_at"`
}

type MarketSourceSyncItemResult struct {
	SourceKey      string `json:"source_key"`
	Status         string `json:"status"`
	BarCount       int    `json:"bar_count,omitempty"`
	NewsCount      int    `json:"news_count,omitempty"`
	TruthCount     int    `json:"truth_count,omitempty"`
	InventoryCount int    `json:"inventory_count,omitempty"`
	SnapshotCount  int    `json:"snapshot_count,omitempty"`
	Message        string `json:"message,omitempty"`
}

type MarketSyncResult struct {
	AssetClass         string                       `json:"asset_class,omitempty"`
	DataKind           string                       `json:"data_kind"`
	RequestedSourceKey string                       `json:"requested_source_key,omitempty"`
	ResolvedSourceKeys []string                     `json:"resolved_source_keys,omitempty"`
	BarCount           int                          `json:"bar_count,omitempty"`
	NewsCount          int                          `json:"news_count,omitempty"`
	TruthCount         int                          `json:"truth_count,omitempty"`
	InventoryCount     int                          `json:"inventory_count,omitempty"`
	SnapshotCount      int                          `json:"snapshot_count,omitempty"`
	Results            []MarketSourceSyncItemResult `json:"results,omitempty"`
}
