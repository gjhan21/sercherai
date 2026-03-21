package model

import "strings"

const (
	StrategyEngineStockSelectionModeAuto   = "AUTO"
	StrategyEngineStockSelectionModeManual = "MANUAL"
	StrategyEngineStockSelectionModeDebug  = "DEBUG"

	StrategyEngineDefaultStockSelectionProfileID = "profile_default_stock_auto"
	StrategyEngineDefaultStockUniverseScope      = "CN_A_ALL"
	StrategyEngineDefaultStockMinListingDays     = 180
	StrategyEngineDefaultStockMinAvgTurnover     = 50_000_000
)

type StrategyEngineStockSelectionContextRequest struct {
	TradeDate        string   `json:"trade_date"`
	SelectionMode    string   `json:"selection_mode"`
	UniverseScope    string   `json:"universe_scope"`
	ProfileID        string   `json:"profile_id"`
	DebugSeedSymbols []string `json:"debug_seed_symbols"`
	SeedSymbols      []string `json:"seed_symbols"`
	ExcludedSymbols  []string `json:"excluded_symbols"`
	Limit            int      `json:"limit"`
	MarketScope      string   `json:"market_scope"`
	MinListingDays   int      `json:"min_listing_days"`
	MinAvgTurnover   float64  `json:"min_avg_turnover"`
}

func (r StrategyEngineStockSelectionContextRequest) Normalized() StrategyEngineStockSelectionContextRequest {
	normalized := r
	normalized.TradeDate = strings.TrimSpace(normalized.TradeDate)
	normalized.ProfileID = strings.TrimSpace(normalized.ProfileID)
	normalized.UniverseScope = strings.TrimSpace(normalized.UniverseScope)
	normalized.MarketScope = strings.TrimSpace(normalized.MarketScope)
	normalized.DebugSeedSymbols = normalizeStrategyEngineStockContextSymbols(normalized.DebugSeedSymbols)
	normalized.SeedSymbols = normalizeStrategyEngineStockContextSymbols(normalized.SeedSymbols)
	normalized.ExcludedSymbols = normalizeStrategyEngineStockContextSymbols(normalized.ExcludedSymbols)
	normalized.SelectionMode = resolveStrategyEngineStockSelectionMode(
		normalized.SelectionMode,
		normalized.SeedSymbols,
		normalized.DebugSeedSymbols,
	)
	normalized.UniverseScope = resolveStrategyEngineStockUniverseScope(normalized.UniverseScope, normalized.MarketScope)
	if normalized.MarketScope == "" {
		normalized.MarketScope = normalized.UniverseScope
	}
	if normalized.ProfileID == "" {
		normalized.ProfileID = StrategyEngineDefaultStockSelectionProfileID
	}
	if normalized.MinListingDays <= 0 {
		normalized.MinListingDays = StrategyEngineDefaultStockMinListingDays
	}
	if normalized.MinAvgTurnover <= 0 {
		normalized.MinAvgTurnover = StrategyEngineDefaultStockMinAvgTurnover
	}
	return normalized
}

func resolveStrategyEngineStockSelectionMode(raw string, seedSymbols []string, debugSeedSymbols []string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case StrategyEngineStockSelectionModeAuto,
		StrategyEngineStockSelectionModeManual,
		StrategyEngineStockSelectionModeDebug:
		return strings.ToUpper(strings.TrimSpace(raw))
	}
	if len(debugSeedSymbols) > 0 {
		return StrategyEngineStockSelectionModeDebug
	}
	if len(seedSymbols) > 0 {
		return StrategyEngineStockSelectionModeManual
	}
	return StrategyEngineStockSelectionModeAuto
}

func resolveStrategyEngineStockUniverseScope(universeScope string, marketScope string) string {
	scope := strings.TrimSpace(universeScope)
	if scope != "" {
		return scope
	}
	scope = strings.TrimSpace(marketScope)
	if scope != "" {
		return scope
	}
	return StrategyEngineDefaultStockUniverseScope
}

func normalizeStrategyEngineStockContextSymbols(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item))
		if symbol == "" {
			continue
		}
		if _, exists := seen[symbol]; exists {
			continue
		}
		seen[symbol] = struct{}{}
		result = append(result, symbol)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

type StrategyEngineStockSeed struct {
	Symbol           string  `json:"symbol"`
	Name             string  `json:"name"`
	TradeDate        string  `json:"trade_date"`
	ClosePrice       float64 `json:"close_price"`
	Momentum5        float64 `json:"momentum5"`
	Momentum20       float64 `json:"momentum20"`
	Volatility20     float64 `json:"volatility20"`
	VolumeRatio      float64 `json:"volume_ratio"`
	Drawdown20       float64 `json:"drawdown20"`
	TrendStrength    float64 `json:"trend_strength"`
	NetMFAmount      float64 `json:"net_mf_amount"`
	PeTTM            float64 `json:"pe_ttm"`
	PB               float64 `json:"pb"`
	TurnoverRate     float64 `json:"turnover_rate"`
	NewsHeat         int     `json:"news_heat"`
	PositiveNewsRate float64 `json:"positive_news_rate"`
	ListingDays      int     `json:"listing_days"`
	AvgTurnover20    float64 `json:"avg_turnover_20"`
	SuspendedProxy   bool    `json:"suspended_proxy"`
	STRiskProxy      bool    `json:"st_risk_proxy"`
	Industry         string  `json:"industry,omitempty"`
	Sector           string  `json:"sector,omitempty"`
	ThemeTags        []string `json:"theme_tags,omitempty"`
	RiskFlags        []string `json:"risk_flags,omitempty"`
}

type StrategyEngineStockSelectionContextMeta struct {
	SelectedTradeDate         string   `json:"selected_trade_date"`
	PriceSource               string   `json:"price_source"`
	NewsWindowDays            int      `json:"news_window_days"`
	ListingDaysFilterApplied  bool     `json:"listing_days_filter_applied"`
	Warnings                  []string `json:"warnings,omitempty"`
}

type StrategyEngineStockSelectionContextResponse struct {
	Seeds []StrategyEngineStockSeed               `json:"seeds"`
	Meta  StrategyEngineStockSelectionContextMeta `json:"meta"`
}

type StrategyEngineFuturesStrategyContextRequest struct {
	TradeDate                       string   `json:"trade_date"`
	Contracts                       []string `json:"contracts"`
	Limit                           int      `json:"limit"`
	AllowMockFallbackOnShortHistory bool     `json:"allow_mock_fallback_on_short_history"`
}

type StrategyEngineFuturesSeed struct {
	Contract                string  `json:"contract"`
	Name                    string  `json:"name"`
	TradeDate               string  `json:"trade_date"`
	LastPrice               float64 `json:"last_price"`
	BasisPct                float64 `json:"basis_pct"`
	Volatility14            float64 `json:"volatility14"`
	TrendStrength           float64 `json:"trend_strength"`
	OIChangePct             float64 `json:"oi_change_pct"`
	VolumeRatio             float64 `json:"volume_ratio"`
	TurnoverRatio           float64 `json:"turnover_ratio"`
	FlowBias                float64 `json:"flow_bias"`
	CarryPct                float64 `json:"carry_pct"`
	TermStructurePct        float64 `json:"term_structure_pct"`
	CurveSlopePct           float64 `json:"curve_slope_pct"`
	InventoryLevel          float64 `json:"inventory_level"`
	InventoryChangePct      float64 `json:"inventory_change_pct"`
	InventoryPressure       float64 `json:"inventory_pressure"`
	InventoryFocusArea      string  `json:"inventory_focus_area,omitempty"`
	InventoryFocusWarehouse string  `json:"inventory_focus_warehouse,omitempty"`
	InventoryFocusBrand     string  `json:"inventory_focus_brand,omitempty"`
	InventoryFocusPlace     string  `json:"inventory_focus_place,omitempty"`
	InventoryFocusGrade     string  `json:"inventory_focus_grade,omitempty"`
	InventoryAreaShare      float64 `json:"inventory_area_share"`
	InventoryWarehouseShare float64 `json:"inventory_warehouse_share"`
	InventoryBrandShare     float64 `json:"inventory_brand_share"`
	InventoryPlaceShare     float64 `json:"inventory_place_share"`
	InventoryGradeShare     float64 `json:"inventory_grade_share"`
	SpreadPressure          float64 `json:"spread_pressure"`
	SpreadPercentile        float64 `json:"spread_percentile"`
	SpreadPair              string  `json:"spread_pair"`
	NewsBias                float64 `json:"news_bias"`
	Regime                  string  `json:"regime"`
}

type StrategyEngineFuturesStrategyContextMeta struct {
	SelectedTradeDate string   `json:"selected_trade_date"`
	PriceSource       string   `json:"price_source"`
	NewsWindowDays    int      `json:"news_window_days"`
	Warnings          []string `json:"warnings,omitempty"`
}

type StrategyEngineFuturesStrategyContextResponse struct {
	Seeds []StrategyEngineFuturesSeed              `json:"seeds"`
	Meta  StrategyEngineFuturesStrategyContextMeta `json:"meta"`
}
