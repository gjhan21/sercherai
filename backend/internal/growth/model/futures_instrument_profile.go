package model

type FuturesInstrumentProfile struct {
	AssetClass          string         `json:"asset_class"`
	ProductKey          string         `json:"product_key"`
	CommodityLabel      string         `json:"commodity_label"`
	ExchangeCode        string         `json:"exchange_code,omitempty"`
	ContractChain       []string       `json:"contract_chain,omitempty"`
	DeliveryPlaces      []string       `json:"delivery_places,omitempty"`
	Warehouses          []string       `json:"warehouses,omitempty"`
	Brands              []string       `json:"brands,omitempty"`
	Grades              []string       `json:"grades,omitempty"`
	InventoryMetricKeys []string       `json:"inventory_metric_keys,omitempty"`
	Metadata            map[string]any `json:"metadata,omitempty"`
	SourceUpdatedAt     string         `json:"source_updated_at,omitempty"`
	CreatedAt           string         `json:"created_at,omitempty"`
	UpdatedAt           string         `json:"updated_at,omitempty"`
}
