package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

const (
	marketInstrumentStockPriorityConfigKey   = "market.instrument.stock.source_priority"
	marketInstrumentFuturesPriorityConfigKey = "market.instrument.futures.source_priority"
	localPlaceholderMarketSourceKey          = "LOCAL_PLACEHOLDER"
)

type marketInstrumentSourceFact struct {
	AssetClass      string
	InstrumentKey   string
	SourceKey       string
	ExternalSymbol  string
	DisplayName     string
	ExchangeCode    string
	ProductKey      string
	ListDate        string
	DelistDate      string
	Status          string
	MetadataJSON    string
	QualityScore    float64
	SourceUpdatedAt time.Time
}

type marketInstrumentTruth struct {
	AssetClass        string
	InstrumentKey     string
	DisplayName       string
	ExchangeCode      string
	ProductKey        string
	ListDate          string
	DelistDate        string
	Status            string
	MetadataJSON      string
	SelectedSourceKey string
	QualityScore      float64
	SourceUpdatedAt   time.Time
}

type tushareStockBasicRecord struct {
	TSCode     string
	Symbol     string
	Name       string
	Area       string
	Industry   string
	Market     string
	ListDate   string
	DelistDate string
	ListStatus string
	IsHS       string
	Exchange   string
}

type tushareStockCompanyRecord struct {
	TSCode        string
	Exchange      string
	Chairman      string
	Manager       string
	Secretary     string
	RegCapital    float64
	SetupDate     string
	Province      string
	City          string
	Website       string
	Email         string
	Employees     int
	MainBusiness  string
	BusinessScope string
}

type tushareFuturesBasicRecord struct {
	TSCode        string
	Symbol        string
	Exchange      string
	Name          string
	FutCode       string
	TradeUnit     string
	Multiplier    int
	QuoteUnit     string
	QuoteUnitDesc string
	ListDate      string
	DelistDate    string
	DModeDesc     string
	TradeTimeDesc string
	LastTradeDate string
	DeliveryMonth string
}

func marketInstrumentPriorityConfigKey(assetClass string) string {
	switch assetClass {
	case marketAssetClassFutures:
		return marketInstrumentFuturesPriorityConfigKey
	default:
		return marketInstrumentStockPriorityConfigKey
	}
}

func defaultMarketInstrumentSourcePriority(assetClass string) []string {
	switch assetClass {
	case marketAssetClassFutures:
		return []string{"TUSHARE", "AKSHARE", "TICKERMD", "MYSELF", "MOCK"}
	default:
		return []string{"TUSHARE", "AKSHARE", "TICKERMD", "MYSELF", "MOCK"}
	}
}

func (r *MySQLGrowthRepo) syncMarketInstrumentMasterData(assetClass string, requestedSourceKey string, instrumentKeys []string) error {
	if len(instrumentKeys) == 0 {
		return nil
	}
	if err := r.upsertMarketInstruments(assetClass, instrumentKeys); err != nil {
		return err
	}

	sourceKeys := r.resolveRequestedMarketSourceKeysWithGovernance(
		requestedSourceKey,
		assetClass,
		"INSTRUMENT_MASTER",
		marketInstrumentPriorityConfigKey(assetClass),
		defaultMarketInstrumentSourcePriority(assetClass),
	)

	for _, resolvedSourceKey := range sourceKeys {
		sourceItem, err := r.getDataSourceBySourceKey(resolvedSourceKey)
		if err != nil {
			r.insertMarketDataQualityLog(assetClass, "INSTRUMENT_MASTER", "", "", resolvedSourceKey, "WARN", "MASTER_SOURCE_LOOKUP_FAILED", err.Error(), "")
			continue
		}
		facts, err := r.fetchMarketInstrumentFactsForSource(sourceItem, assetClass, instrumentKeys)
		if err != nil || len(facts) == 0 {
			if err != nil {
				r.insertMarketDataQualityLog(assetClass, "INSTRUMENT_MASTER", "", "", resolvedSourceKey, "WARN", "MASTER_FETCH_FAILED", err.Error(), "")
			}
			continue
		}
		if err := r.upsertMarketInstrumentSourceFacts(facts); err != nil {
			if isMarketInstrumentSchemaCompatError(err) {
				continue
			}
			return err
		}
		if err := r.upsertMarketSymbolAliasesFromInstrumentFacts(facts); err != nil {
			return err
		}
	}

	if err := r.rebuildMarketInstrumentTruth(assetClass, instrumentKeys, sourceKeys); err != nil {
		if isMarketInstrumentSchemaCompatError(err) {
			return nil
		}
		return err
	}
	return nil
}

func (r *MySQLGrowthRepo) fetchMarketInstrumentFactsForSource(item model.DataSource, assetClass string, instrumentKeys []string) ([]marketInstrumentSourceFact, error) {
	sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
	provider := strings.ToUpper(parseDataSourceStringConfig(item.Config, "provider", "vendor"))
	if provider == "" {
		provider = sourceKey
	}
	sourceKey = canonicalMarketSourceKey(sourceKey, provider)
	switch provider {
	case "TUSHARE":
		token := parseDataSourceStringConfig(item.Config, "token", "api_token", "tushare_token")
		if strings.TrimSpace(token) == "" {
			token = strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
		}
		if assetClass == marketAssetClassFutures {
			return fetchFuturesInstrumentFactsFromTushare(token, sourceKey, instrumentKeys, parseDataSourceTimeoutMS(item.Config))
		}
		if universeAssetType := normalizeUniverseAssetType(assetClass); universeAssetType != "" && universeAssetType != marketAssetClassStock {
			universeItems, err := r.fetchMarketUniverseItemsForSource(sourceKey, universeAssetType)
			if err != nil {
				return nil, err
			}
			filteredItems := filterMarketUniverseItemsByInstrumentKeys(universeItems, instrumentKeys)
			return buildMarketInstrumentSourceFactsFromUniverseItems(sourceKey, universeAssetType, filteredItems, time.Now()), nil
		}
		return fetchStockInstrumentFactsFromTushare(token, sourceKey, instrumentKeys, parseDataSourceTimeoutMS(item.Config))
	default:
		return buildFallbackMarketInstrumentSourceFacts(sourceKey, assetClass, instrumentKeys), nil
	}
}

func fetchStockInstrumentFactsFromTushare(token string, sourceKey string, instrumentKeys []string, timeoutMS int) ([]marketInstrumentSourceFact, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	facts := make([]marketInstrumentSourceFact, 0, len(instrumentKeys))
	errs := make([]string, 0)
	fetchedAt := time.Now()
	for _, instrumentKey := range instrumentKeys {
		basic, ok, err := fetchTushareStockBasicRecord(client, token, instrumentKey)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", instrumentKey, err))
			continue
		}
		if !ok {
			continue
		}
		company, companyErr := fetchTushareStockCompanyRecord(client, token, instrumentKey)
		if companyErr != nil {
			errs = append(errs, fmt.Sprintf("%s company: %v", instrumentKey, companyErr))
		}
		fact, ok := buildTushareStockInstrumentFact(basic, company, fetchedAt)
		if !ok {
			continue
		}
		fact.SourceKey = sourceKey
		fact.QualityScore = deriveMarketInstrumentQualityScore(fact)
		facts = append(facts, fact)
	}
	if len(facts) == 0 && len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, "; "))
	}
	return facts, nil
}

func fetchFuturesInstrumentFactsFromTushare(token string, sourceKey string, instrumentKeys []string, timeoutMS int) ([]marketInstrumentSourceFact, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	facts := make([]marketInstrumentSourceFact, 0, len(instrumentKeys))
	errs := make([]string, 0)
	fetchedAt := time.Now()
	for _, instrumentKey := range instrumentKeys {
		record, ok, err := fetchTushareFuturesBasicRecord(client, token, instrumentKey)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", instrumentKey, err))
			continue
		}
		if !ok {
			continue
		}
		fact, ok := buildTushareFuturesInstrumentFact(record, fetchedAt)
		if !ok {
			continue
		}
		fact.SourceKey = sourceKey
		fact.QualityScore = deriveMarketInstrumentQualityScore(fact)
		facts = append(facts, fact)
	}
	if len(facts) == 0 && len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, "; "))
	}
	return facts, nil
}

func fetchTushareStockBasicRecord(client *http.Client, token string, instrumentKey string) (tushareStockBasicRecord, bool, error) {
	parsed, err := callTushareAPI(client, token, "stock_basic", map[string]string{
		"ts_code": strings.ToUpper(strings.TrimSpace(instrumentKey)),
	}, "")
	if err != nil {
		return tushareStockBasicRecord{}, false, err
	}
	fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
	for _, row := range parsed.Data.Items {
		tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
		if !ok {
			continue
		}
		return tushareStockBasicRecord{
			TSCode:     strings.ToUpper(strings.TrimSpace(tsCode)),
			Symbol:     fallbackTushareString(row, fieldIndex, "symbol"),
			Name:       fallbackTushareString(row, fieldIndex, "name"),
			Area:       fallbackTushareString(row, fieldIndex, "area"),
			Industry:   fallbackTushareString(row, fieldIndex, "industry"),
			Market:     fallbackTushareString(row, fieldIndex, "market"),
			ListDate:   fallbackTushareString(row, fieldIndex, "list_date"),
			DelistDate: fallbackTushareString(row, fieldIndex, "delist_date"),
			ListStatus: fallbackTushareString(row, fieldIndex, "list_status"),
			IsHS:       fallbackTushareString(row, fieldIndex, "is_hs"),
			Exchange:   fallbackTushareString(row, fieldIndex, "exchange"),
		}, true, nil
	}
	return tushareStockBasicRecord{}, false, nil
}

func fetchTushareStockCompanyRecord(client *http.Client, token string, instrumentKey string) (*tushareStockCompanyRecord, error) {
	parsed, err := callTushareAPI(client, token, "stock_company", map[string]string{
		"ts_code": strings.ToUpper(strings.TrimSpace(instrumentKey)),
	}, "")
	if err != nil {
		return nil, err
	}
	fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
	for _, row := range parsed.Data.Items {
		tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
		if !ok {
			continue
		}
		record := &tushareStockCompanyRecord{
			TSCode:        strings.ToUpper(strings.TrimSpace(tsCode)),
			Exchange:      fallbackTushareString(row, fieldIndex, "exchange"),
			Chairman:      fallbackTushareString(row, fieldIndex, "chairman"),
			Manager:       fallbackTushareString(row, fieldIndex, "manager"),
			Secretary:     fallbackTushareString(row, fieldIndex, "secretary"),
			SetupDate:     fallbackTushareString(row, fieldIndex, "setup_date"),
			Province:      fallbackTushareString(row, fieldIndex, "province"),
			City:          fallbackTushareString(row, fieldIndex, "city"),
			Website:       fallbackTushareString(row, fieldIndex, "website"),
			Email:         fallbackTushareString(row, fieldIndex, "email"),
			MainBusiness:  fallbackTushareString(row, fieldIndex, "main_business"),
			BusinessScope: fallbackTushareString(row, fieldIndex, "business_scope"),
		}
		if regCapital, ok := tushareGetFloat(row, fieldIndex, "reg_capital"); ok {
			record.RegCapital = regCapital
		}
		if employees, ok := tushareGetFloat(row, fieldIndex, "employees"); ok {
			record.Employees = int(employees)
		}
		return record, nil
	}
	return nil, nil
}

func fetchTushareFuturesBasicRecord(client *http.Client, token string, instrumentKey string) (tushareFuturesBasicRecord, bool, error) {
	parsed, err := callTushareAPI(client, token, "fut_basic", map[string]string{
		"ts_code": strings.ToUpper(strings.TrimSpace(instrumentKey)),
	}, "")
	if err != nil {
		return tushareFuturesBasicRecord{}, false, err
	}
	fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
	for _, row := range parsed.Data.Items {
		tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
		if !ok {
			continue
		}
		record := tushareFuturesBasicRecord{
			TSCode:        strings.ToUpper(strings.TrimSpace(tsCode)),
			Symbol:        fallbackTushareString(row, fieldIndex, "symbol"),
			Exchange:      fallbackTushareString(row, fieldIndex, "exchange"),
			Name:          fallbackTushareString(row, fieldIndex, "name"),
			FutCode:       fallbackTushareString(row, fieldIndex, "fut_code"),
			TradeUnit:     fallbackTushareString(row, fieldIndex, "trade_unit"),
			QuoteUnit:     fallbackTushareString(row, fieldIndex, "quote_unit"),
			QuoteUnitDesc: fallbackTushareString(row, fieldIndex, "quote_unit_desc"),
			ListDate:      fallbackTushareString(row, fieldIndex, "list_date"),
			DelistDate:    fallbackTushareString(row, fieldIndex, "delist_date"),
			DModeDesc:     fallbackTushareString(row, fieldIndex, "d_mode_desc"),
			TradeTimeDesc: fallbackTushareString(row, fieldIndex, "trade_time_desc"),
			LastTradeDate: fallbackTushareString(row, fieldIndex, "last_ddate"),
			DeliveryMonth: fallbackTushareString(row, fieldIndex, "d_month"),
		}
		if multiplier, ok := tushareGetFloat(row, fieldIndex, "multiplier"); ok {
			record.Multiplier = int(multiplier)
		}
		return record, true, nil
	}
	return tushareFuturesBasicRecord{}, false, nil
}

func fallbackTushareString(row []interface{}, fieldIndex map[string]int, key string) string {
	value, _ := tushareGetString(row, fieldIndex, key)
	return value
}

func buildTushareStockInstrumentFact(basic tushareStockBasicRecord, company *tushareStockCompanyRecord, fetchedAt time.Time) (marketInstrumentSourceFact, bool) {
	instrumentKey := strings.ToUpper(strings.TrimSpace(basic.TSCode))
	if instrumentKey == "" {
		return marketInstrumentSourceFact{}, false
	}
	exchangeCode := normalizeMarketExchangeCode(basic.Exchange, detectInstrumentExchangeCode(instrumentKey))
	if company != nil {
		exchangeCode = normalizeMarketExchangeCode(company.Exchange, exchangeCode)
	}
	productKey := deriveMarketProductKey(marketAssetClassStock, instrumentKey, basic.Symbol)
	metadata := map[string]interface{}{
		"area":        strings.TrimSpace(basic.Area),
		"industry":    strings.TrimSpace(basic.Industry),
		"market":      strings.TrimSpace(basic.Market),
		"list_status": strings.TrimSpace(basic.ListStatus),
		"is_hs":       strings.TrimSpace(basic.IsHS),
	}
	if company != nil {
		metadata["chairman"] = strings.TrimSpace(company.Chairman)
		metadata["manager"] = strings.TrimSpace(company.Manager)
		metadata["secretary"] = strings.TrimSpace(company.Secretary)
		metadata["province"] = strings.TrimSpace(company.Province)
		metadata["city"] = strings.TrimSpace(company.City)
		metadata["website"] = strings.TrimSpace(company.Website)
		metadata["email"] = strings.TrimSpace(company.Email)
		metadata["main_business"] = strings.TrimSpace(company.MainBusiness)
		metadata["business_scope"] = strings.TrimSpace(company.BusinessScope)
		if company.RegCapital > 0 {
			metadata["reg_capital"] = company.RegCapital
		}
		if company.Employees > 0 {
			metadata["employees"] = company.Employees
		}
		setupDate := normalizeMarketDate(company.SetupDate)
		if setupDate != "" {
			metadata["setup_date"] = setupDate
		}
	}
	fact := marketInstrumentSourceFact{
		AssetClass:      marketAssetClassStock,
		InstrumentKey:   instrumentKey,
		ExternalSymbol:  instrumentKey,
		DisplayName:     strings.TrimSpace(basic.Name),
		ExchangeCode:    exchangeCode,
		ProductKey:      productKey,
		ListDate:        normalizeMarketDate(basic.ListDate),
		DelistDate:      normalizeMarketDate(basic.DelistDate),
		Status:          resolveMarketInstrumentStatus(strings.TrimSpace(basic.ListStatus), normalizeMarketDate(basic.DelistDate), fetchedAt),
		MetadataJSON:    marshalJSONSilently(filterEmptyJSONMap(metadata)),
		SourceUpdatedAt: fetchedAt,
	}
	if strings.TrimSpace(fact.DisplayName) == "" {
		fact.DisplayName = instrumentKey
	}
	return fact, true
}

func buildTushareFuturesInstrumentFact(item tushareFuturesBasicRecord, fetchedAt time.Time) (marketInstrumentSourceFact, bool) {
	instrumentKey := strings.ToUpper(strings.TrimSpace(item.TSCode))
	if instrumentKey == "" {
		return marketInstrumentSourceFact{}, false
	}
	exchangeCode := normalizeMarketExchangeCode(item.Exchange, detectInstrumentExchangeCode(instrumentKey))
	productKey := deriveMarketProductKey(marketAssetClassFutures, instrumentKey, item.FutCode)
	listDate := normalizeMarketDate(item.ListDate)
	delistDate := normalizeMarketDate(item.DelistDate)
	metadata := filterEmptyJSONMap(map[string]interface{}{
		"trade_unit":      strings.TrimSpace(item.TradeUnit),
		"quote_unit":      strings.TrimSpace(item.QuoteUnit),
		"quote_unit_desc": strings.TrimSpace(item.QuoteUnitDesc),
		"d_mode_desc":     strings.TrimSpace(item.DModeDesc),
		"trade_time_desc": strings.TrimSpace(item.TradeTimeDesc),
		"last_trade_date": normalizeMarketDate(item.LastTradeDate),
		"delivery_month":  strings.TrimSpace(item.DeliveryMonth),
		"multiplier":      item.Multiplier,
	})
	fact := marketInstrumentSourceFact{
		AssetClass:      marketAssetClassFutures,
		InstrumentKey:   instrumentKey,
		ExternalSymbol:  instrumentKey,
		DisplayName:     strings.TrimSpace(item.Name),
		ExchangeCode:    exchangeCode,
		ProductKey:      productKey,
		ListDate:        listDate,
		DelistDate:      delistDate,
		Status:          resolveMarketInstrumentStatus("", delistDate, fetchedAt),
		MetadataJSON:    marshalJSONSilently(metadata),
		SourceUpdatedAt: fetchedAt,
	}
	if strings.TrimSpace(fact.DisplayName) == "" {
		fact.DisplayName = instrumentKey
	}
	return fact, true
}

func buildFallbackMarketInstrumentSourceFacts(sourceKey string, assetClass string, instrumentKeys []string) []marketInstrumentSourceFact {
	now := time.Now()
	facts := make([]marketInstrumentSourceFact, 0, len(instrumentKeys))
	for _, instrumentKey := range instrumentKeys {
		normalizedKey := strings.ToUpper(strings.TrimSpace(instrumentKey))
		if normalizedKey == "" {
			continue
		}
		fact := marketInstrumentSourceFact{
			AssetClass:      assetClass,
			InstrumentKey:   normalizedKey,
			SourceKey:       strings.ToUpper(strings.TrimSpace(sourceKey)),
			ExternalSymbol:  deriveDefaultExternalSymbol(strings.ToUpper(strings.TrimSpace(sourceKey)), assetClass, normalizedKey),
			ExchangeCode:    detectInstrumentExchangeCode(normalizedKey),
			ProductKey:      deriveMarketProductKey(assetClass, normalizedKey, ""),
			Status:          "ACTIVE",
			MetadataJSON:    marshalJSONSilently(map[string]interface{}{"placeholder": true}),
			SourceUpdatedAt: now,
		}
		fact.QualityScore = deriveMarketInstrumentQualityScore(fact)
		facts = append(facts, fact)
	}
	return facts
}

func buildMarketInstrumentSourceFactsFromUniverseItems(sourceKey string, assetClass string, items []marketUniverseSourceItem, fetchedAt time.Time) []marketInstrumentSourceFact {
	facts := make([]marketInstrumentSourceFact, 0, len(items))
	for _, item := range items {
		instrumentKey := strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		if instrumentKey == "" {
			continue
		}
		fact := marketInstrumentSourceFact{
			AssetClass:      strings.ToUpper(strings.TrimSpace(assetClass)),
			InstrumentKey:   instrumentKey,
			SourceKey:       strings.ToUpper(strings.TrimSpace(sourceKey)),
			ExternalSymbol:  defaultString(strings.TrimSpace(item.ExternalSymbol), instrumentKey),
			DisplayName:     defaultString(strings.TrimSpace(item.DisplayName), instrumentKey),
			ExchangeCode:    normalizeMarketExchangeCode(item.ExchangeCode, detectInstrumentExchangeCode(instrumentKey)),
			ProductKey:      deriveMarketProductKey(assetClass, instrumentKey, ""),
			ListDate:        normalizeMarketDate(item.ListDate),
			DelistDate:      normalizeMarketDate(item.DelistDate),
			Status:          normalizeMarketUniverseStatus(item.Status, item.DelistDate, fetchedAt),
			MetadataJSON:    strings.TrimSpace(item.MetadataJSON),
			SourceUpdatedAt: fetchedAt,
		}
		fact.QualityScore = deriveMarketInstrumentQualityScore(fact)
		facts = append(facts, fact)
	}
	return facts
}

func filterMarketUniverseItemsByInstrumentKeys(items []marketUniverseSourceItem, instrumentKeys []string) []marketUniverseSourceItem {
	if len(instrumentKeys) == 0 {
		return items
	}
	allowed := make(map[string]struct{}, len(instrumentKeys))
	for _, instrumentKey := range instrumentKeys {
		normalized := strings.ToUpper(strings.TrimSpace(instrumentKey))
		if normalized == "" {
			continue
		}
		allowed[normalized] = struct{}{}
	}
	if len(allowed) == 0 {
		return items
	}
	filtered := make([]marketUniverseSourceItem, 0, len(items))
	for _, item := range items {
		if _, ok := allowed[strings.ToUpper(strings.TrimSpace(item.InstrumentKey))]; ok {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func (r *MySQLGrowthRepo) upsertMarketInstrumentSourceFacts(facts []marketInstrumentSourceFact) error {
	now := time.Now()
	for _, fact := range facts {
		if strings.TrimSpace(fact.AssetClass) == "" || strings.TrimSpace(fact.InstrumentKey) == "" || strings.TrimSpace(fact.SourceKey) == "" {
			continue
		}
		sourceUpdatedAt := nullableTime(fact.SourceUpdatedAt)
		_, err := r.db.Exec(`
INSERT INTO market_instrument_source_facts
  (id, asset_class, instrument_key, source_key, external_symbol, display_name, exchange_code, product_key, list_date, delist_date, status, metadata_json, quality_score, source_updated_at, fetched_at, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  external_symbol = VALUES(external_symbol),
  display_name = VALUES(display_name),
  exchange_code = VALUES(exchange_code),
  product_key = VALUES(product_key),
  list_date = VALUES(list_date),
  delist_date = VALUES(delist_date),
  status = VALUES(status),
  metadata_json = VALUES(metadata_json),
  quality_score = VALUES(quality_score),
  source_updated_at = VALUES(source_updated_at),
  fetched_at = VALUES(fetched_at),
  updated_at = VALUES(updated_at)`,
			newID("mif"),
			fact.AssetClass,
			fact.InstrumentKey,
			fact.SourceKey,
			nullableString(fact.ExternalSymbol),
			nullableString(fact.DisplayName),
			nullableString(fact.ExchangeCode),
			nullableString(fact.ProductKey),
			nullableDate(fact.ListDate),
			nullableDate(fact.DelistDate),
			nullableString(defaultString(fact.Status, "ACTIVE")),
			nullableString(strings.TrimSpace(fact.MetadataJSON)),
			fact.QualityScore,
			sourceUpdatedAt,
			now,
			now,
			now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) upsertMarketSymbolAliasesFromInstrumentFacts(facts []marketInstrumentSourceFact) error {
	now := time.Now()
	for _, fact := range facts {
		sourceKey := strings.ToUpper(strings.TrimSpace(fact.SourceKey))
		if strings.TrimSpace(fact.AssetClass) == "" || strings.TrimSpace(fact.InstrumentKey) == "" || sourceKey == "" {
			continue
		}
		externalSymbol := strings.TrimSpace(fact.ExternalSymbol)
		if externalSymbol == "" {
			externalSymbol = deriveDefaultExternalSymbol(sourceKey, fact.AssetClass, fact.InstrumentKey)
		}
		metadata := ""
		if fact.DisplayName != "" {
			metadata = marshalJSONSilently(map[string]interface{}{
				"display_name": fact.DisplayName,
				"product_key":  fact.ProductKey,
			})
		}
		_, err := r.db.Exec(`
INSERT INTO market_symbol_aliases
  (id, asset_class, instrument_key, source_key, external_symbol, status, metadata_json, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, 'ACTIVE', ?, ?, ?)
ON DUPLICATE KEY UPDATE
  external_symbol = VALUES(external_symbol),
  status = VALUES(status),
  metadata_json = VALUES(metadata_json),
  updated_at = VALUES(updated_at)`,
			newID("msa"),
			fact.AssetClass,
			fact.InstrumentKey,
			sourceKey,
			externalSymbol,
			nullableString(metadata),
			now,
			now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) rebuildMarketInstrumentTruth(assetClass string, instrumentKeys []string, priority []string) error {
	byInstrument, err := r.loadMarketInstrumentSourceFacts(assetClass, instrumentKeys)
	if err != nil {
		return err
	}
	now := time.Now()
	for _, instrumentKey := range instrumentKeys {
		truth := resolveMarketInstrumentTruth(assetClass, instrumentKey, byInstrument[strings.ToUpper(strings.TrimSpace(instrumentKey))], priority)
		truthVersion := truth.SourceUpdatedAt.Unix()
		if truthVersion <= 0 {
			truthVersion = now.Unix()
		}
		_, err := r.db.Exec(`
UPDATE market_instruments
SET
  display_name = ?,
  exchange_code = ?,
  selected_source_key = ?,
  product_key = ?,
  list_date = ?,
  delist_date = ?,
  truth_version = ?,
  quality_score = ?,
  source_updated_at = ?,
  status = ?,
  metadata_json = ?,
  updated_at = ?
WHERE asset_class = ? AND instrument_key = ?`,
			truth.DisplayName,
			nullableString(truth.ExchangeCode),
			nullableString(truth.SelectedSourceKey),
			nullableString(truth.ProductKey),
			nullableDate(truth.ListDate),
			nullableDate(truth.DelistDate),
			truthVersion,
			truth.QualityScore,
			nullableTime(truth.SourceUpdatedAt),
			defaultString(truth.Status, "ACTIVE"),
			nullableString(strings.TrimSpace(truth.MetadataJSON)),
			now,
			assetClass,
			strings.ToUpper(strings.TrimSpace(instrumentKey)),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) loadMarketInstrumentSourceFacts(assetClass string, instrumentKeys []string) (map[string][]marketInstrumentSourceFact, error) {
	result := make(map[string][]marketInstrumentSourceFact, len(instrumentKeys))
	if len(instrumentKeys) == 0 {
		return result, nil
	}
	holders := make([]string, 0, len(instrumentKeys))
	args := make([]interface{}, 0, len(instrumentKeys)+1)
	args = append(args, assetClass)
	for _, instrumentKey := range instrumentKeys {
		normalizedKey := strings.ToUpper(strings.TrimSpace(instrumentKey))
		if normalizedKey == "" {
			continue
		}
		holders = append(holders, "?")
		args = append(args, normalizedKey)
	}
	if len(holders) == 0 {
		return result, nil
	}
	query := `
SELECT
  asset_class,
  instrument_key,
  source_key,
  external_symbol,
  display_name,
  exchange_code,
  product_key,
  DATE_FORMAT(list_date, '%Y-%m-%d') AS list_date,
  DATE_FORMAT(delist_date, '%Y-%m-%d') AS delist_date,
  status,
  metadata_json,
  quality_score,
  source_updated_at
FROM market_instrument_source_facts
WHERE asset_class = ? AND instrument_key IN (` + strings.Join(holders, ",") + `)`
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isMarketInstrumentSchemaCompatError(err) {
			return result, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			item            marketInstrumentSourceFact
			externalSymbol  sql.NullString
			displayName     sql.NullString
			exchangeCode    sql.NullString
			productKey      sql.NullString
			listDate        sql.NullString
			delistDate      sql.NullString
			status          sql.NullString
			metadataJSON    sql.NullString
			qualityScore    sql.NullFloat64
			sourceUpdatedAt sql.NullTime
		)
		if err := rows.Scan(
			&item.AssetClass,
			&item.InstrumentKey,
			&item.SourceKey,
			&externalSymbol,
			&displayName,
			&exchangeCode,
			&productKey,
			&listDate,
			&delistDate,
			&status,
			&metadataJSON,
			&qualityScore,
			&sourceUpdatedAt,
		); err != nil {
			return nil, err
		}
		item.AssetClass = strings.ToUpper(strings.TrimSpace(item.AssetClass))
		item.InstrumentKey = strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		item.SourceKey = strings.ToUpper(strings.TrimSpace(item.SourceKey))
		item.ExternalSymbol = strings.TrimSpace(externalSymbol.String)
		item.DisplayName = strings.TrimSpace(displayName.String)
		item.ExchangeCode = strings.ToUpper(strings.TrimSpace(exchangeCode.String))
		item.ProductKey = strings.ToUpper(strings.TrimSpace(productKey.String))
		item.ListDate = strings.TrimSpace(listDate.String)
		item.DelistDate = strings.TrimSpace(delistDate.String)
		item.Status = strings.ToUpper(strings.TrimSpace(status.String))
		item.MetadataJSON = strings.TrimSpace(metadataJSON.String)
		if qualityScore.Valid {
			item.QualityScore = qualityScore.Float64
		}
		if sourceUpdatedAt.Valid {
			item.SourceUpdatedAt = sourceUpdatedAt.Time
		}
		result[item.InstrumentKey] = append(result[item.InstrumentKey], item)
	}
	return result, nil
}

func resolveMarketInstrumentTruth(assetClass string, instrumentKey string, facts []marketInstrumentSourceFact, priority []string) marketInstrumentTruth {
	normalizedKey := strings.ToUpper(strings.TrimSpace(instrumentKey))
	if normalizedKey == "" {
		return marketInstrumentTruth{}
	}
	if len(facts) == 0 {
		return buildPlaceholderMarketInstrumentTruth(assetClass, normalizedKey)
	}
	orderedFacts := orderMarketInstrumentFacts(facts, priority)
	selected, ok := pickPreferredMarketInstrumentFact(orderedFacts)
	if !ok {
		return buildPlaceholderMarketInstrumentTruth(assetClass, normalizedKey)
	}
	truth := marketInstrumentTruth{
		AssetClass:        assetClass,
		InstrumentKey:     normalizedKey,
		DisplayName:       strings.TrimSpace(selected.DisplayName),
		ExchangeCode:      strings.ToUpper(strings.TrimSpace(selected.ExchangeCode)),
		ProductKey:        strings.ToUpper(strings.TrimSpace(selected.ProductKey)),
		ListDate:          strings.TrimSpace(selected.ListDate),
		DelistDate:        strings.TrimSpace(selected.DelistDate),
		Status:            strings.ToUpper(strings.TrimSpace(selected.Status)),
		MetadataJSON:      strings.TrimSpace(selected.MetadataJSON),
		SelectedSourceKey: strings.ToUpper(strings.TrimSpace(selected.SourceKey)),
		QualityScore:      selected.QualityScore,
		SourceUpdatedAt:   selected.SourceUpdatedAt,
	}
	for _, fact := range orderedFacts {
		if truth.DisplayName == "" && fact.DisplayName != "" {
			truth.DisplayName = strings.TrimSpace(fact.DisplayName)
		}
		if truth.ExchangeCode == "" && fact.ExchangeCode != "" {
			truth.ExchangeCode = strings.ToUpper(strings.TrimSpace(fact.ExchangeCode))
		}
		if truth.ProductKey == "" && fact.ProductKey != "" {
			truth.ProductKey = strings.ToUpper(strings.TrimSpace(fact.ProductKey))
		}
		if truth.ListDate == "" && fact.ListDate != "" {
			truth.ListDate = strings.TrimSpace(fact.ListDate)
		}
		if truth.DelistDate == "" && fact.DelistDate != "" {
			truth.DelistDate = strings.TrimSpace(fact.DelistDate)
		}
		if truth.MetadataJSON == "" && fact.MetadataJSON != "" {
			truth.MetadataJSON = strings.TrimSpace(fact.MetadataJSON)
		}
		if truth.SourceUpdatedAt.IsZero() && !fact.SourceUpdatedAt.IsZero() {
			truth.SourceUpdatedAt = fact.SourceUpdatedAt
		}
		if truth.QualityScore < fact.QualityScore {
			truth.QualityScore = fact.QualityScore
		}
	}
	if truth.DisplayName == "" {
		truth.DisplayName = normalizedKey
	}
	if truth.ExchangeCode == "" {
		truth.ExchangeCode = detectInstrumentExchangeCode(normalizedKey)
	}
	if truth.ProductKey == "" {
		truth.ProductKey = deriveMarketProductKey(assetClass, normalizedKey, "")
	}
	if truth.Status == "" {
		truth.Status = resolveMarketInstrumentStatus("", truth.DelistDate, truth.SourceUpdatedAt)
	}
	if truth.QualityScore <= 0 {
		truth.QualityScore = deriveMarketInstrumentQualityScore(selected)
	}
	return truth
}

func buildPlaceholderMarketInstrumentTruth(assetClass string, instrumentKey string) marketInstrumentTruth {
	return marketInstrumentTruth{
		AssetClass:        assetClass,
		InstrumentKey:     instrumentKey,
		DisplayName:       instrumentKey,
		ExchangeCode:      detectInstrumentExchangeCode(instrumentKey),
		ProductKey:        deriveMarketProductKey(assetClass, instrumentKey, ""),
		Status:            "ACTIVE",
		SelectedSourceKey: localPlaceholderMarketSourceKey,
		QualityScore:      0.2,
	}
}

func orderMarketInstrumentFacts(facts []marketInstrumentSourceFact, priority []string) []marketInstrumentSourceFact {
	if len(facts) == 0 {
		return nil
	}
	priorityIndex := make(map[string]int, len(priority))
	for idx, sourceKey := range priority {
		priorityIndex[strings.ToUpper(strings.TrimSpace(sourceKey))] = idx
	}
	ordered := append([]marketInstrumentSourceFact(nil), facts...)
	sort.SliceStable(ordered, func(i, j int) bool {
		leftKey := strings.ToUpper(strings.TrimSpace(ordered[i].SourceKey))
		rightKey := strings.ToUpper(strings.TrimSpace(ordered[j].SourceKey))
		leftIndex, leftOK := priorityIndex[leftKey]
		rightIndex, rightOK := priorityIndex[rightKey]
		switch {
		case leftOK && rightOK:
			if leftIndex != rightIndex {
				return leftIndex < rightIndex
			}
		case leftOK:
			return true
		case rightOK:
			return false
		}
		return ordered[i].QualityScore > ordered[j].QualityScore
	})
	return ordered
}

func pickPreferredMarketInstrumentFact(facts []marketInstrumentSourceFact) (marketInstrumentSourceFact, bool) {
	for _, fact := range facts {
		if strings.TrimSpace(fact.DisplayName) != "" {
			return fact, true
		}
	}
	for _, fact := range facts {
		if isUsableMarketInstrumentFact(fact) {
			return fact, true
		}
	}
	return marketInstrumentSourceFact{}, false
}

func isUsableMarketInstrumentFact(fact marketInstrumentSourceFact) bool {
	return strings.TrimSpace(fact.DisplayName) != "" ||
		strings.TrimSpace(fact.ProductKey) != "" ||
		strings.TrimSpace(fact.ListDate) != "" ||
		strings.TrimSpace(fact.DelistDate) != "" ||
		strings.TrimSpace(fact.ExternalSymbol) != ""
}

func deriveMarketInstrumentQualityScore(fact marketInstrumentSourceFact) float64 {
	score := 0.15
	if strings.TrimSpace(fact.DisplayName) != "" {
		score += 0.35
	}
	if strings.TrimSpace(fact.ExchangeCode) != "" {
		score += 0.1
	}
	if strings.TrimSpace(fact.ProductKey) != "" {
		score += 0.1
	}
	if strings.TrimSpace(fact.ListDate) != "" {
		score += 0.1
	}
	if strings.TrimSpace(fact.DelistDate) != "" {
		score += 0.05
	}
	if strings.TrimSpace(fact.MetadataJSON) != "" && strings.TrimSpace(fact.MetadataJSON) != "{}" {
		score += 0.1
	}
	switch strings.ToUpper(strings.TrimSpace(fact.SourceKey)) {
	case "TUSHARE":
		score += 0.15
	case "AKSHARE", "MYSELF":
		score += 0.05
	}
	if score > 1 {
		score = 1
	}
	return roundTo(score, 2)
}

func deriveMarketProductKey(assetClass string, instrumentKey string, fallback string) string {
	if trimmed := strings.ToUpper(strings.TrimSpace(fallback)); trimmed != "" {
		return trimmed
	}
	normalizedKey := strings.ToUpper(strings.TrimSpace(instrumentKey))
	if normalizedKey == "" {
		return ""
	}
	code := normalizedKey
	if idx := strings.LastIndex(code, "."); idx >= 0 {
		code = code[:idx]
	}
	if assetClass == marketAssetClassFutures {
		letters := make([]rune, 0, len(code))
		for _, ch := range code {
			if ch >= 'A' && ch <= 'Z' {
				letters = append(letters, ch)
				continue
			}
			break
		}
		if len(letters) > 0 {
			return string(letters)
		}
	}
	return code
}

func normalizeMarketExchangeCode(raw string, fallback string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "SSE", "SH":
		return "SH"
	case "SZSE", "SZ":
		return "SZ"
	case "BSE", "BJ":
		return "BJ"
	case "CFFEX", "CFX":
		return "CFX"
	case "SHFE", "SHF":
		return "SHF"
	case "DCE":
		return "DCE"
	case "CZCE", "ZCE":
		return "CZCE"
	case "INE":
		return "INE"
	case "GFEX", "GFE":
		return "GFEX"
	}
	return strings.ToUpper(strings.TrimSpace(defaultString(fallback, raw)))
}

func normalizeMarketDate(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "<nil>" {
		return ""
	}
	layouts := []string{
		"20060102",
		"2006-01-02",
		"2006/01/02",
		time.RFC3339,
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return parsed.Format("2006-01-02")
		}
	}
	return ""
}

func resolveMarketInstrumentStatus(listStatus string, delistDate string, asOf time.Time) string {
	switch strings.ToUpper(strings.TrimSpace(listStatus)) {
	case "D":
		return "DELISTED"
	case "P":
		return "PENDING"
	}
	normalizedDelistDate := normalizeMarketDate(delistDate)
	if normalizedDelistDate == "" {
		return "ACTIVE"
	}
	parsed, err := time.ParseInLocation("2006-01-02", normalizedDelistDate, time.Local)
	if err != nil {
		return "ACTIVE"
	}
	reference := asOf
	if reference.IsZero() {
		reference = time.Now()
	}
	if parsed.Before(reference.Truncate(24*time.Hour)) || parsed.Equal(reference.Truncate(24*time.Hour)) {
		return "DELISTED"
	}
	return "ACTIVE"
}

func filterEmptyJSONMap(values map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{}, len(values))
	for key, value := range values {
		switch current := value.(type) {
		case string:
			if strings.TrimSpace(current) == "" {
				continue
			}
		case int:
			if current == 0 {
				continue
			}
		case float64:
			if current == 0 {
				continue
			}
		case nil:
			continue
		}
		result[key] = value
	}
	return result
}

func nullableDate(value string) interface{} {
	normalized := normalizeMarketDate(value)
	if normalized == "" {
		return nil
	}
	return normalized
}

func nullableTime(value time.Time) interface{} {
	if value.IsZero() {
		return nil
	}
	return value
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func isMarketInstrumentSchemaCompatError(err error) bool {
	if err == nil {
		return false
	}
	if isTableNotFoundError(err) {
		return true
	}
	text := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(text, "unknown column") || strings.Contains(text, "error 1054")
}
