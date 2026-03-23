package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

type stockStatusTruthInput struct {
	TradeDate      time.Time
	Symbol         string
	DisplayName    string
	SelectedSource string
	Volume         float64
	Turnover       float64
	ListDate       time.Time
	MetadataJSON   string
}

type stockStatusTruthRecord struct {
	TradeDate      string
	Symbol         string
	ListDate       string
	SelectedSource string
	IsSuspended    bool
	IsST           bool
	RiskWarning    bool
	ReasonCodes    []string
	MetadataJSON   string
}

type strategyStockStatusTruth struct {
	ListDate     time.Time
	IsSuspended  bool
	IsST         bool
	RiskWarning  bool
	ReasonCodes  []string
	SelectedFrom string
}

type futuresContractSnapshot struct {
	InstrumentKey   string
	ProductKey      string
	ExchangeCode    string
	Turnover        float64
	OpenInterest    float64
	SelectedSource  string
	SettlePrice     float64
	PrevSettlePrice float64
}

type futuresContractMappingRecord struct {
	TradeDate              string
	ProductKey             string
	ExchangeCode           string
	DominantInstrumentKey  string
	SecondaryInstrumentKey string
	NearInstrumentKey      string
	SelectedSourceKey      string
	MappingMethod          string
	MetadataJSON           string
}

func normalizeMarketDataQualitySeverity(value string) string {
	switch strings.ToUpper(strings.TrimSpace(value)) {
	case "INFO":
		return "INFO"
	case "ERROR":
		return "ERROR"
	default:
		return "WARN"
	}
}

func buildStockStatusTruthRecord(input stockStatusTruthInput) stockStatusTruthRecord {
	symbol := strings.ToUpper(strings.TrimSpace(input.Symbol))
	displayName := strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = symbol
	}
	_, _, _, _, riskWarning := parseStockInstrumentMetadata(input.MetadataJSON)
	isSTByName := isStockNameSTRisk(displayName)
	isSuspended := input.Volume <= 0 || input.Turnover <= 0
	isST := isSTByName || riskWarning
	reasonCodes := make([]string, 0, 3)
	if isSuspended {
		reasonCodes = append(reasonCodes, "SUSPENDED_PROXY")
	}
	if isSTByName {
		reasonCodes = append(reasonCodes, "ST_NAME_PREFIX")
	}
	if riskWarning {
		reasonCodes = append(reasonCodes, "RISK_WARNING")
	}
	metadataJSON := marshalJSONSilently(filterEmptyJSONMap(map[string]interface{}{
		"display_name": displayName,
		"volume":       input.Volume,
		"turnover":     input.Turnover,
	}))
	return stockStatusTruthRecord{
		TradeDate: input.TradeDate.Format("2006-01-02"),
		Symbol:    symbol,
		ListDate: func() string {
			if input.ListDate.IsZero() {
				return ""
			}
			return input.ListDate.Format("2006-01-02")
		}(),
		SelectedSource: strings.ToUpper(strings.TrimSpace(input.SelectedSource)),
		IsSuspended:    isSuspended,
		IsST:           isST,
		RiskWarning:    riskWarning,
		ReasonCodes:    reasonCodes,
		MetadataJSON:   metadataJSON,
	}
}

func isStockNameSTRisk(value string) bool {
	name := strings.ToUpper(strings.TrimSpace(value))
	return strings.HasPrefix(name, "ST") || strings.HasPrefix(name, "*ST")
}

func buildFuturesContractMappingRecord(tradeDate time.Time, snapshots []futuresContractSnapshot) (futuresContractMappingRecord, bool) {
	if len(snapshots) == 0 {
		return futuresContractMappingRecord{}, false
	}
	filtered := make([]futuresContractSnapshot, 0, len(snapshots))
	for _, item := range snapshots {
		item.InstrumentKey = strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		item.ProductKey = strings.ToUpper(strings.TrimSpace(item.ProductKey))
		item.ExchangeCode = strings.ToUpper(strings.TrimSpace(item.ExchangeCode))
		item.SelectedSource = strings.ToUpper(strings.TrimSpace(item.SelectedSource))
		if item.InstrumentKey == "" || item.ProductKey == "" {
			continue
		}
		if item.ExchangeCode == "" {
			item.ExchangeCode = detectInstrumentExchangeCode(item.InstrumentKey)
		}
		filtered = append(filtered, item)
	}
	if len(filtered) == 0 {
		return futuresContractMappingRecord{}, false
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].Turnover != filtered[j].Turnover {
			return filtered[i].Turnover > filtered[j].Turnover
		}
		if filtered[i].OpenInterest != filtered[j].OpenInterest {
			return filtered[i].OpenInterest > filtered[j].OpenInterest
		}
		return filtered[i].InstrumentKey < filtered[j].InstrumentKey
	})
	dominant := filtered[0]
	near := dominant
	bestExpiry := 0
	for _, item := range filtered {
		contract := normalizeFuturesContextContract(item.InstrumentKey)
		_, expiryCode, ok := splitFuturesContractCode(contract)
		if !ok || expiryCode <= 0 {
			continue
		}
		if bestExpiry == 0 || expiryCode < bestExpiry {
			bestExpiry = expiryCode
			near = item
		}
	}
	metadataJSON := marshalJSONSilently(map[string]interface{}{
		"contract_count": len(filtered),
		"candidates":     collectFuturesMappingCandidates(filtered),
	})
	record := futuresContractMappingRecord{
		TradeDate:             tradeDate.Format("2006-01-02"),
		ProductKey:            dominant.ProductKey,
		ExchangeCode:          dominant.ExchangeCode,
		DominantInstrumentKey: dominant.InstrumentKey,
		NearInstrumentKey:     near.InstrumentKey,
		SelectedSourceKey:     dominant.SelectedSource,
		MappingMethod:         "TURNOVER_OPEN_INTEREST",
		MetadataJSON:          metadataJSON,
	}
	if len(filtered) > 1 {
		record.SecondaryInstrumentKey = filtered[1].InstrumentKey
	}
	return record, true
}

func collectFuturesMappingCandidates(items []futuresContractSnapshot) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if item.InstrumentKey == "" {
			continue
		}
		result = append(result, item.InstrumentKey)
	}
	return result
}

func (r *MySQLGrowthRepo) loadStrategyStockStatusTruthMap(symbols []string, selectedTradeDate time.Time) (map[string]strategyStockStatusTruth, error) {
	result := make(map[string]strategyStockStatusTruth, len(symbols))
	if len(symbols) == 0 {
		return result, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	args := make([]any, 0, len(symbols)+1)
	args = append(args, selectedTradeDate.Format("2006-01-02"))
	for _, symbol := range symbols {
		args = append(args, strings.ToUpper(strings.TrimSpace(symbol)))
	}
	query := fmt.Sprintf(`
SELECT instrument_key, list_date, is_suspended, is_st, risk_warning, status_reason_codes_json
FROM stock_status_truth
WHERE trade_date = ? AND instrument_key IN (%s)`, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isMarketStatusSchemaCompatError(err) {
			return result, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			symbol          string
			listDate        sql.NullTime
			isSuspended     int
			isST            int
			riskWarning     int
			reasonCodesJSON sql.NullString
		)
		if err := rows.Scan(&symbol, &listDate, &isSuspended, &isST, &riskWarning, &reasonCodesJSON); err != nil {
			return nil, err
		}
		item := strategyStockStatusTruth{
			IsSuspended: isSuspended != 0,
			IsST:        isST != 0,
			RiskWarning: riskWarning != 0,
			ReasonCodes: parseMarketStatusReasonCodes(reasonCodesJSON.String),
		}
		if listDate.Valid {
			item.ListDate = listDate.Time
		}
		result[strings.ToUpper(strings.TrimSpace(symbol))] = item
	}
	return result, rows.Err()
}

func parseMarketStatusReasonCodes(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return compactStrings(items)
}

func (r *MySQLGrowthRepo) rebuildStockStatusTruth(items []model.MarketDailyBar) (int, error) {
	symbols := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.AssetClass != marketAssetClassStock {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		if symbol == "" {
			continue
		}
		if _, ok := seen[symbol]; ok {
			continue
		}
		seen[symbol] = struct{}{}
		symbols = append(symbols, symbol)
	}
	if len(symbols) == 0 {
		return 0, nil
	}
	instrumentBaseMap, err := r.loadStockInstrumentStatusBaseMap(symbols)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	count := 0
	for _, item := range items {
		if item.AssetClass != marketAssetClassStock {
			continue
		}
		symbol := strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		if symbol == "" {
			continue
		}
		base := instrumentBaseMap[symbol]
		record := buildStockStatusTruthRecord(stockStatusTruthInput{
			TradeDate:      mustParseTradeDate(item.TradeDate),
			Symbol:         symbol,
			DisplayName:    base.DisplayName,
			SelectedSource: item.SourceKey,
			Volume:         float64(item.Volume),
			Turnover:       item.Turnover,
			ListDate:       base.ListDate,
			MetadataJSON:   base.MetadataJSON,
		})
		_, err := r.db.Exec(`
INSERT INTO stock_status_truth
  (id, trade_date, instrument_key, list_date, selected_source_key, is_suspended, is_st, risk_warning, status_reason_codes_json, metadata_json, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  list_date = VALUES(list_date),
  selected_source_key = VALUES(selected_source_key),
  is_suspended = VALUES(is_suspended),
  is_st = VALUES(is_st),
  risk_warning = VALUES(risk_warning),
  status_reason_codes_json = VALUES(status_reason_codes_json),
  metadata_json = VALUES(metadata_json),
  updated_at = VALUES(updated_at)`,
			newID("sst"),
			record.TradeDate,
			record.Symbol,
			nullableDate(record.ListDate),
			nullableString(record.SelectedSource),
			boolToTinyInt(record.IsSuspended),
			boolToTinyInt(record.IsST),
			boolToTinyInt(record.RiskWarning),
			nullableString(marshalJSONSilently(record.ReasonCodes)),
			nullableString(record.MetadataJSON),
			now,
			now,
		)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (r *MySQLGrowthRepo) rebuildFuturesContractMappings(items []model.MarketDailyBar) (int, error) {
	type groupKey struct {
		TradeDate    string
		ProductKey   string
		ExchangeCode string
	}
	groups := make(map[groupKey][]futuresContractSnapshot)
	for _, item := range items {
		if item.AssetClass != marketAssetClassFutures {
			continue
		}
		instrumentKey := strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		if instrumentKey == "" || strings.TrimSpace(item.TradeDate) == "" {
			continue
		}
		productKey := deriveMarketProductKey(marketAssetClassFutures, instrumentKey, "")
		exchangeCode := detectInstrumentExchangeCode(instrumentKey)
		key := groupKey{
			TradeDate:    item.TradeDate,
			ProductKey:   productKey,
			ExchangeCode: exchangeCode,
		}
		groups[key] = append(groups[key], futuresContractSnapshot{
			InstrumentKey:  instrumentKey,
			ProductKey:     productKey,
			ExchangeCode:   exchangeCode,
			Turnover:       item.Turnover,
			OpenInterest:   item.OpenInterest,
			SelectedSource: item.SourceKey,
		})
	}
	if len(groups) == 0 {
		return 0, nil
	}
	now := time.Now()
	count := 0
	for key, snapshots := range groups {
		tradeDate, err := time.ParseInLocation("2006-01-02", key.TradeDate, time.Local)
		if err != nil {
			continue
		}
		record, ok := buildFuturesContractMappingRecord(tradeDate, snapshots)
		if !ok {
			continue
		}
		_, err = r.db.Exec(`
INSERT INTO futures_contract_mappings
  (id, trade_date, product_key, exchange_code, dominant_instrument_key, secondary_instrument_key, near_instrument_key, selected_source_key, mapping_method, metadata_json, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  dominant_instrument_key = VALUES(dominant_instrument_key),
  secondary_instrument_key = VALUES(secondary_instrument_key),
  near_instrument_key = VALUES(near_instrument_key),
  selected_source_key = VALUES(selected_source_key),
  mapping_method = VALUES(mapping_method),
  metadata_json = VALUES(metadata_json),
  updated_at = VALUES(updated_at)`,
			newID("fcm"),
			record.TradeDate,
			record.ProductKey,
			record.ExchangeCode,
			record.DominantInstrumentKey,
			nullableString(record.SecondaryInstrumentKey),
			nullableString(record.NearInstrumentKey),
			nullableString(record.SelectedSourceKey),
			record.MappingMethod,
			nullableString(record.MetadataJSON),
			now,
			now,
		)
		if err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (r *MySQLGrowthRepo) insertMarketDataQualityLog(assetClass string, dataKind string, instrumentKey string, tradeDate string, sourceKey string, severity string, issueCode string, issueMessage string, payload string) {
	now := time.Now()
	_, _ = r.db.Exec(`
INSERT INTO market_data_quality_logs
  (id, asset_class, data_kind, instrument_key, trade_date, source_key, severity, issue_code, issue_message, payload_json, created_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("mdq"),
		nullableString(strings.ToUpper(strings.TrimSpace(assetClass))),
		strings.ToUpper(strings.TrimSpace(dataKind)),
		nullableString(strings.ToUpper(strings.TrimSpace(instrumentKey))),
		nullableDate(tradeDate),
		nullableString(strings.ToUpper(strings.TrimSpace(sourceKey))),
		normalizeMarketDataQualitySeverity(severity),
		strings.ToUpper(strings.TrimSpace(issueCode)),
		nullableString(truncateByRunes(strings.TrimSpace(issueMessage), 255)),
		nullableString(strings.TrimSpace(payload)),
		now,
	)
}

func (r *MySQLGrowthRepo) loadStockInstrumentStatusBaseMap(symbols []string) (map[string]struct {
	DisplayName  string
	ListDate     time.Time
	MetadataJSON string
}, error) {
	result := make(map[string]struct {
		DisplayName  string
		ListDate     time.Time
		MetadataJSON string
	}, len(symbols))
	if len(symbols) == 0 {
		return result, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	args := make([]any, 0, len(symbols)+1)
	args = append(args, marketAssetClassStock)
	for _, symbol := range symbols {
		args = append(args, strings.ToUpper(strings.TrimSpace(symbol)))
	}
	query := fmt.Sprintf(`
SELECT instrument_key, display_name, list_date, COALESCE(CAST(metadata_json AS CHAR), '')
FROM market_instruments
WHERE asset_class = ? AND instrument_key IN (%s)`, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isMarketStatusSchemaCompatError(err) {
			return result, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			symbol       string
			displayName  sql.NullString
			listDate     sql.NullTime
			metadataJSON sql.NullString
		)
		if err := rows.Scan(&symbol, &displayName, &listDate, &metadataJSON); err != nil {
			return nil, err
		}
		item := struct {
			DisplayName  string
			ListDate     time.Time
			MetadataJSON string
		}{
			DisplayName:  strings.TrimSpace(displayName.String),
			MetadataJSON: strings.TrimSpace(metadataJSON.String),
		}
		if listDate.Valid {
			item.ListDate = listDate.Time
		}
		result[strings.ToUpper(strings.TrimSpace(symbol))] = item
	}
	return result, rows.Err()
}

func isMarketStatusSchemaCompatError(err error) bool {
	if err == nil {
		return false
	}
	if isTableNotFoundError(err) {
		return true
	}
	text := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(text, "unknown column") || strings.Contains(text, "error 1054")
}

func mustParseTradeDate(value string) time.Time {
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(value), time.Local)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func boolToTinyInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
