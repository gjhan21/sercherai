package dto

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
