package repo

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

type marketUniverseSourceItem struct {
	AssetType      string
	InstrumentKey  string
	ExternalSymbol string
	DisplayName    string
	ExchangeCode   string
	Status         string
	ListDate       string
	DelistDate     string
	MetadataJSON   string
}

func normalizeUniverseAssetType(value string) string {
	return normalizeMarketBackfillAssetType(value)
}

func (r *MySQLGrowthRepo) AdminBuildMarketUniverseSnapshot(sourceKey string, assetScope []string, operator string) (model.MarketUniverseSnapshot, []model.MarketUniverseSnapshotItem, error) {
	normalizedScope := normalizeMarketBackfillAssetScope(assetScope)
	if len(normalizedScope) == 0 {
		return model.MarketUniverseSnapshot{}, nil, fmt.Errorf("asset_scope is required")
	}
	normalizedSource := strings.ToUpper(strings.TrimSpace(sourceKey))
	if normalizedSource == "" {
		normalizedSource = "TUSHARE"
	}
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = "system"
	}

	now := time.Now()
	snapshotID := newID("mus")
	items := make([]model.MarketUniverseSnapshotItem, 0, len(normalizedScope)*64)
	for _, assetType := range normalizedScope {
		sourceItems, err := r.fetchMarketUniverseItemsForSource(normalizedSource, assetType)
		if err != nil {
			return model.MarketUniverseSnapshot{}, nil, err
		}
		for _, sourceItem := range sourceItems {
			items = append(items, model.MarketUniverseSnapshotItem{
				ID:             newID("musi"),
				SnapshotID:     snapshotID,
				AssetType:      assetType,
				InstrumentKey:  sourceItem.InstrumentKey,
				ExternalSymbol: sourceItem.ExternalSymbol,
				DisplayName:    sourceItem.DisplayName,
				ExchangeCode:   sourceItem.ExchangeCode,
				Status:         sourceItem.Status,
				ListDate:       sourceItem.ListDate,
				DelistDate:     sourceItem.DelistDate,
				MetadataJSON:   sourceItem.MetadataJSON,
				CreatedAt:      now.Format(time.RFC3339),
			})
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].AssetType != items[j].AssetType {
			return items[i].AssetType < items[j].AssetType
		}
		return items[i].InstrumentKey < items[j].InstrumentKey
	})

	snapshot := model.MarketUniverseSnapshot{
		ID:             snapshotID,
		Scope:          normalizedScope,
		SourceKey:      normalizedSource,
		SnapshotDate:   now.Format("2006-01-02"),
		AssetSummaries: buildMarketUniverseAssetSummaries(items, normalizedScope),
		CreatedBy:      truncateByRunes(normalizeUTF8Text(operator), 64),
		CreatedAt:      now.Format(time.RFC3339),
	}
	summaryJSON := marshalJSONText(map[string]any{
		"asset_summaries": snapshot.AssetSummaries,
		"item_count":      len(items),
	})

	tx, err := r.db.Begin()
	if err != nil {
		return model.MarketUniverseSnapshot{}, nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
INSERT INTO market_universe_snapshots (id, scope, source_key, snapshot_date, summary_json, created_by, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		snapshot.ID,
		marshalJSONText(snapshot.Scope),
		snapshot.SourceKey,
		snapshot.SnapshotDate,
		summaryJSON,
		nullableString(snapshot.CreatedBy),
		now,
	); err != nil {
		return model.MarketUniverseSnapshot{}, nil, err
	}
	for _, item := range items {
		if _, err := tx.Exec(`
INSERT INTO market_universe_snapshot_items
  (id, snapshot_id, asset_type, instrument_key, external_symbol, display_name, exchange_code, status, list_date, delist_date, raw_metadata_json, created_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			item.ID,
			item.SnapshotID,
			item.AssetType,
			item.InstrumentKey,
			nullableString(item.ExternalSymbol),
			nullableString(item.DisplayName),
			nullableString(item.ExchangeCode),
			nullableString(item.Status),
			nullableDate(item.ListDate),
			nullableDate(item.DelistDate),
			nullableString(item.MetadataJSON),
			now,
		); err != nil {
			return model.MarketUniverseSnapshot{}, nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return model.MarketUniverseSnapshot{}, nil, err
	}
	return snapshot, items, nil
}

func (r *InMemoryGrowthRepo) AdminBuildMarketUniverseSnapshot(sourceKey string, assetScope []string, operator string) (model.MarketUniverseSnapshot, []model.MarketUniverseSnapshotItem, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.buildMarketUniverseSnapshotLocked(sourceKey, assetScope, operator)
}

func (r *InMemoryGrowthRepo) buildMarketUniverseSnapshotLocked(sourceKey string, assetScope []string, operator string) (model.MarketUniverseSnapshot, []model.MarketUniverseSnapshotItem, error) {
	normalizedScope := normalizeMarketBackfillAssetScope(assetScope)
	if len(normalizedScope) == 0 {
		return model.MarketUniverseSnapshot{}, nil, fmt.Errorf("asset_scope is required")
	}
	normalizedSource := strings.ToUpper(strings.TrimSpace(sourceKey))
	if normalizedSource == "" {
		normalizedSource = "TUSHARE"
	}
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = "system"
	}

	now := time.Now()
	createdAt := now.Format(time.RFC3339)
	snapshot := model.MarketUniverseSnapshot{
		ID:           "mus_" + strings.ToLower(strings.ReplaceAll(newID("snap"), "_", "")),
		Scope:        normalizedScope,
		SourceKey:    normalizedSource,
		SnapshotDate: now.Format("2006-01-02"),
		CreatedBy:    operator,
		CreatedAt:    createdAt,
	}
	items := make([]model.MarketUniverseSnapshotItem, 0, len(normalizedScope))
	for _, assetType := range normalizedScope {
		template := inMemoryUniverseTemplateForAsset(assetType)
		items = append(items, model.MarketUniverseSnapshotItem{
			ID:             "musi_" + strings.ToLower(strings.ReplaceAll(newID("snapitem"), "_", "")),
			SnapshotID:     snapshot.ID,
			AssetType:      assetType,
			InstrumentKey:  template.InstrumentKey,
			ExternalSymbol: template.ExternalSymbol,
			DisplayName:    template.DisplayName,
			ExchangeCode:   template.ExchangeCode,
			Status:         template.Status,
			ListDate:       template.ListDate,
			DelistDate:     template.DelistDate,
			MetadataJSON:   template.MetadataJSON,
			CreatedAt:      createdAt,
		})
	}
	snapshot.AssetSummaries = buildMarketUniverseAssetSummaries(items, normalizedScope)
	r.marketUniverseSnapshots[snapshot.ID] = snapshot
	r.marketUniverseItems[snapshot.ID] = items
	return snapshot, items, nil
}

func inMemoryUniverseTemplateForAsset(assetType string) marketUniverseSourceItem {
	switch normalizeUniverseAssetType(assetType) {
	case "INDEX":
		return marketUniverseSourceItem{
			AssetType:      "INDEX",
			InstrumentKey:  "000300.SH",
			ExternalSymbol: "000300.SH",
			DisplayName:    "沪深300",
			ExchangeCode:   "SH",
			Status:         "ACTIVE",
			ListDate:       "2005-04-08",
			MetadataJSON:   `{"provider":"demo","category":"broad_index"}`,
		}
	case "ETF":
		return marketUniverseSourceItem{
			AssetType:      "ETF",
			InstrumentKey:  "510300.SH",
			ExternalSymbol: "510300.SH",
			DisplayName:    "沪深300ETF",
			ExchangeCode:   "SH",
			Status:         "ACTIVE",
			ListDate:       "2012-05-28",
			MetadataJSON:   `{"provider":"demo","fund_type":"ETF"}`,
		}
	case "LOF":
		return marketUniverseSourceItem{
			AssetType:      "LOF",
			InstrumentKey:  "161725.SZ",
			ExternalSymbol: "161725.SZ",
			DisplayName:    "白酒LOF",
			ExchangeCode:   "SZ",
			Status:         "ACTIVE",
			ListDate:       "2015-05-27",
			MetadataJSON:   `{"provider":"demo","fund_type":"LOF"}`,
		}
	case "CBOND":
		return marketUniverseSourceItem{
			AssetType:      "CBOND",
			InstrumentKey:  "113519.SH",
			ExternalSymbol: "113519.SH",
			DisplayName:    "长久转债",
			ExchangeCode:   "SH",
			Status:         "ACTIVE",
			ListDate:       "2018-11-07",
			MetadataJSON:   `{"provider":"demo","bond_type":"convertible"}`,
		}
	default:
		return marketUniverseSourceItem{
			AssetType:      "STOCK",
			InstrumentKey:  "600519.SH",
			ExternalSymbol: "600519.SH",
			DisplayName:    "贵州茅台",
			ExchangeCode:   "SH",
			Status:         "ACTIVE",
			ListDate:       "2001-08-27",
			MetadataJSON:   `{"provider":"demo","industry":"白酒"}`,
		}
	}
}

func buildMarketUniverseAssetSummaries(items []model.MarketUniverseSnapshotItem, assetScope []string) []model.MarketUniverseSnapshotAssetItem {
	orderedScope := normalizeMarketBackfillAssetScope(assetScope)
	summaries := make([]model.MarketUniverseSnapshotAssetItem, 0, len(orderedScope))
	byAsset := make(map[string]*model.MarketUniverseSnapshotAssetItem, len(orderedScope))
	for _, assetType := range orderedScope {
		item := &model.MarketUniverseSnapshotAssetItem{AssetType: assetType}
		byAsset[assetType] = item
		summaries = append(summaries, *item)
	}
	for _, item := range items {
		summary, ok := byAsset[item.AssetType]
		if !ok {
			summary = &model.MarketUniverseSnapshotAssetItem{AssetType: item.AssetType}
			byAsset[item.AssetType] = summary
			summaries = append(summaries, *summary)
		}
		summary.ItemCount++
		if isActiveMarketUniverseStatus(item.Status) {
			summary.ActiveCount++
		} else {
			summary.InactiveCount++
		}
	}
	for idx := range summaries {
		if current, ok := byAsset[summaries[idx].AssetType]; ok {
			summaries[idx] = *current
		}
	}
	return summaries
}

func isActiveMarketUniverseStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "", "ACTIVE", "LISTED", "PENDING":
		return true
	default:
		return false
	}
}

func (r *MySQLGrowthRepo) fetchMarketUniverseItemsForSource(sourceKey string, assetType string) ([]marketUniverseSourceItem, error) {
	assetType = normalizeUniverseAssetType(assetType)
	if assetType == "" {
		return nil, fmt.Errorf("unsupported asset_type")
	}
	resolvedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	if resolvedSourceKey == "" {
		resolvedSourceKey = "TUSHARE"
	}
	provider := resolvedSourceKey
	timeoutMS := 12000
	token := strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
	if sourceItem, err := r.getDataSourceBySourceKey(resolvedSourceKey); err == nil {
		if configuredProvider := strings.ToUpper(parseDataSourceStringConfig(sourceItem.Config, "provider", "vendor")); configuredProvider != "" {
			provider = configuredProvider
		}
		resolvedSourceKey = canonicalMarketSourceKey(resolvedSourceKey, provider)
		if configuredToken := strings.TrimSpace(parseDataSourceStringConfig(sourceItem.Config, "token", "api_token", "tushare_token")); configuredToken != "" {
			token = configuredToken
		}
		if configuredTimeout := parseDataSourceTimeoutMS(sourceItem.Config); configuredTimeout > 0 {
			timeoutMS = configuredTimeout
		}
	}

	if strings.EqualFold(provider, "TUSHARE") {
		items, err := fetchTushareMarketUniverseItems(token, resolvedSourceKey, assetType, timeoutMS)
		if err == nil && len(items) > 0 {
			return items, nil
		}
		if masterItems, masterErr := r.loadMarketUniverseItemsFromMaster(assetType); masterErr == nil && len(masterItems) > 0 {
			return masterItems, nil
		}
		if err != nil {
			return nil, err
		}
	}

	items, err := r.loadMarketUniverseItemsFromMaster(assetType)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no universe items found for asset_type=%s source=%s", assetType, resolvedSourceKey)
	}
	return items, nil
}

func (r *MySQLGrowthRepo) loadMarketUniverseItemsFromMaster(assetType string) ([]marketUniverseSourceItem, error) {
	rows, err := r.db.Query(`
SELECT instrument_key, COALESCE(display_name, ''), COALESCE(exchange_code, ''), COALESCE(status, ''),
       COALESCE(DATE_FORMAT(list_date, '%Y-%m-%d'), ''), COALESCE(DATE_FORMAT(delist_date, '%Y-%m-%d'), ''),
       COALESCE(CAST(metadata_json AS CHAR), '')
FROM market_instruments
WHERE asset_class = ?
ORDER BY instrument_key ASC`, assetType)
	if err != nil {
		if isMarketInstrumentSchemaCompatError(err) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	items := make([]marketUniverseSourceItem, 0, 256)
	for rows.Next() {
		var item marketUniverseSourceItem
		if err := rows.Scan(&item.InstrumentKey, &item.DisplayName, &item.ExchangeCode, &item.Status, &item.ListDate, &item.DelistDate, &item.MetadataJSON); err != nil {
			return nil, err
		}
		item.AssetType = assetType
		item.InstrumentKey = strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		item.ExternalSymbol = item.InstrumentKey
		item.DisplayName = strings.TrimSpace(item.DisplayName)
		item.ExchangeCode = strings.ToUpper(strings.TrimSpace(item.ExchangeCode))
		item.Status = normalizeMarketUniverseStatus(item.Status, item.DelistDate, time.Now())
		item.ListDate = normalizeMarketDate(item.ListDate)
		item.DelistDate = normalizeMarketDate(item.DelistDate)
		if item.InstrumentKey == "" {
			continue
		}
		if item.DisplayName == "" {
			item.DisplayName = item.InstrumentKey
		}
		if item.ExchangeCode == "" {
			item.ExchangeCode = detectInstrumentExchangeCode(item.InstrumentKey)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func fetchTushareMarketUniverseItems(token string, sourceKey string, assetType string, timeoutMS int) ([]marketUniverseSourceItem, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	switch assetType {
	case "STOCK":
		return fetchTushareStockUniverseItems(client, token, sourceKey)
	case "INDEX":
		return fetchTushareIndexUniverseItems(client, token, sourceKey)
	case "ETF", "LOF":
		return fetchTushareFundUniverseItems(client, token, sourceKey, assetType)
	case "CBOND":
		return fetchTushareCBondUniverseItems(client, token, sourceKey)
	default:
		return nil, fmt.Errorf("unsupported asset_type=%s", assetType)
	}
}

func fetchTushareStockUniverseItems(client *http.Client, token string, sourceKey string) ([]marketUniverseSourceItem, error) {
	responses, err := fetchPaginatedTushareUniverseResponses(client, token, "stock_basic", nil, "ts_code,symbol,name,area,industry,market,list_date,delist_date,list_status,is_hs,exchange")
	if err != nil {
		return nil, err
	}
	now := time.Now()
	items := make([]marketUniverseSourceItem, 0, 4096)
	seen := make(map[string]struct{}, 4096)
	for _, response := range responses {
		fieldIndex := buildTushareFieldIndex(response.Data.Fields)
		for _, row := range response.Data.Items {
			instrumentKey := strings.ToUpper(firstTushareString(row, fieldIndex, "ts_code"))
			if instrumentKey == "" {
				continue
			}
			if _, exists := seen[instrumentKey]; exists {
				continue
			}
			seen[instrumentKey] = struct{}{}
			listStatus := firstTushareString(row, fieldIndex, "list_status")
			delistDate := normalizeMarketDate(firstTushareString(row, fieldIndex, "delist_date"))
			exchangeCode := normalizeMarketExchangeCode(firstTushareString(row, fieldIndex, "exchange"), detectInstrumentExchangeCode(instrumentKey))
			items = append(items, marketUniverseSourceItem{
				AssetType:      "STOCK",
				InstrumentKey:  instrumentKey,
				ExternalSymbol: defaultString(firstTushareString(row, fieldIndex, "symbol"), instrumentKey),
				DisplayName:    defaultString(firstTushareString(row, fieldIndex, "name"), instrumentKey),
				ExchangeCode:   exchangeCode,
				Status:         normalizeMarketUniverseStatus(listStatus, delistDate, now),
				ListDate:       normalizeMarketDate(firstTushareString(row, fieldIndex, "list_date")),
				DelistDate:     delistDate,
				MetadataJSON: marshalJSONSilently(filterEmptyJSONMap(map[string]interface{}{
					"source_key": sourceKey,
					"area":       firstTushareString(row, fieldIndex, "area"),
					"industry":   firstTushareString(row, fieldIndex, "industry"),
					"market":     firstTushareString(row, fieldIndex, "market"),
					"is_hs":      firstTushareString(row, fieldIndex, "is_hs"),
				})),
			})
		}
	}
	return items, nil
}

func fetchTushareIndexUniverseItems(client *http.Client, token string, sourceKey string) ([]marketUniverseSourceItem, error) {
	responses, err := fetchPaginatedTushareUniverseResponses(client, token, "index_basic", nil, "ts_code,name,market,publisher,category,base_date,base_point,list_date,exp_date")
	if err != nil {
		return nil, err
	}
	now := time.Now()
	items := make([]marketUniverseSourceItem, 0, 1024)
	seen := make(map[string]struct{}, 1024)
	for _, response := range responses {
		fieldIndex := buildTushareFieldIndex(response.Data.Fields)
		for _, row := range response.Data.Items {
			instrumentKey := strings.ToUpper(firstTushareString(row, fieldIndex, "ts_code"))
			if instrumentKey == "" {
				continue
			}
			if _, exists := seen[instrumentKey]; exists {
				continue
			}
			seen[instrumentKey] = struct{}{}
			expDate := normalizeMarketDate(firstTushareString(row, fieldIndex, "exp_date"))
			items = append(items, marketUniverseSourceItem{
				AssetType:      "INDEX",
				InstrumentKey:  instrumentKey,
				ExternalSymbol: instrumentKey,
				DisplayName:    defaultString(firstTushareString(row, fieldIndex, "name"), instrumentKey),
				ExchangeCode:   detectInstrumentExchangeCode(instrumentKey),
				Status:         normalizeMarketUniverseStatus("", expDate, now),
				ListDate:       normalizeMarketDate(firstTushareString(row, fieldIndex, "list_date")),
				DelistDate:     expDate,
				MetadataJSON: marshalJSONSilently(filterEmptyJSONMap(map[string]interface{}{
					"source_key": sourceKey,
					"market":     firstTushareString(row, fieldIndex, "market"),
					"publisher":  firstTushareString(row, fieldIndex, "publisher"),
					"category":   firstTushareString(row, fieldIndex, "category"),
					"base_date":  normalizeMarketDate(firstTushareString(row, fieldIndex, "base_date")),
					"base_point": firstTushareString(row, fieldIndex, "base_point"),
				})),
			})
		}
	}
	return items, nil
}

func fetchTushareFundUniverseItems(client *http.Client, token string, sourceKey string, requestedAssetType string) ([]marketUniverseSourceItem, error) {
	responses, err := fetchPaginatedTushareUniverseResponses(client, token, "fund_basic", map[string]string{"market": "E"}, "ts_code,name,management,custodian,fund_type,found_date,due_date,list_date,issue_date,delist_date,invest_type,type,status,market")
	if err != nil {
		return nil, err
	}
	now := time.Now()
	items := make([]marketUniverseSourceItem, 0, 2048)
	seen := make(map[string]struct{}, 2048)
	for _, response := range responses {
		fieldIndex := buildTushareFieldIndex(response.Data.Fields)
		for _, row := range response.Data.Items {
			instrumentKey := strings.ToUpper(firstTushareString(row, fieldIndex, "ts_code"))
			if instrumentKey == "" {
				continue
			}
			assetType := classifyTushareFundAssetType(
				firstTushareString(row, fieldIndex, "name"),
				firstTushareString(row, fieldIndex, "fund_type"),
				firstTushareString(row, fieldIndex, "invest_type"),
				firstTushareString(row, fieldIndex, "type"),
			)
			if assetType != requestedAssetType {
				continue
			}
			if _, exists := seen[instrumentKey]; exists {
				continue
			}
			seen[instrumentKey] = struct{}{}
			delistDate := normalizeMarketDate(firstTushareString(row, fieldIndex, "delist_date"))
			items = append(items, marketUniverseSourceItem{
				AssetType:      assetType,
				InstrumentKey:  instrumentKey,
				ExternalSymbol: instrumentKey,
				DisplayName:    defaultString(firstTushareString(row, fieldIndex, "name"), instrumentKey),
				ExchangeCode:   detectInstrumentExchangeCode(instrumentKey),
				Status:         normalizeMarketUniverseStatus(firstTushareString(row, fieldIndex, "status"), delistDate, now),
				ListDate:       normalizeMarketDate(firstTushareString(row, fieldIndex, "list_date")),
				DelistDate:     delistDate,
				MetadataJSON: marshalJSONSilently(filterEmptyJSONMap(map[string]interface{}{
					"source_key":  sourceKey,
					"fund_type":   firstTushareString(row, fieldIndex, "fund_type"),
					"invest_type": firstTushareString(row, fieldIndex, "invest_type"),
					"type":        firstTushareString(row, fieldIndex, "type"),
					"management":  firstTushareString(row, fieldIndex, "management"),
					"custodian":   firstTushareString(row, fieldIndex, "custodian"),
					"market":      firstTushareString(row, fieldIndex, "market"),
					"found_date":  normalizeMarketDate(firstTushareString(row, fieldIndex, "found_date")),
					"issue_date":  normalizeMarketDate(firstTushareString(row, fieldIndex, "issue_date")),
					"due_date":    normalizeMarketDate(firstTushareString(row, fieldIndex, "due_date")),
				})),
			})
		}
	}
	return items, nil
}

func fetchTushareCBondUniverseItems(client *http.Client, token string, sourceKey string) ([]marketUniverseSourceItem, error) {
	responses, err := fetchPaginatedTushareUniverseResponses(client, token, "cb_basic", nil, "ts_code,bond_short_name,stk_code,stk_short_name,list_date,delist_date,conv_start_date,conv_end_date,maturity_date,issue_size,remain_size,rating")
	if err != nil {
		return nil, err
	}
	now := time.Now()
	items := make([]marketUniverseSourceItem, 0, 1024)
	seen := make(map[string]struct{}, 1024)
	for _, response := range responses {
		fieldIndex := buildTushareFieldIndex(response.Data.Fields)
		for _, row := range response.Data.Items {
			instrumentKey := strings.ToUpper(firstTushareString(row, fieldIndex, "ts_code"))
			if instrumentKey == "" {
				continue
			}
			if _, exists := seen[instrumentKey]; exists {
				continue
			}
			seen[instrumentKey] = struct{}{}
			delistDate := normalizeMarketDate(firstTushareString(row, fieldIndex, "delist_date"))
			if delistDate == "" {
				delistDate = normalizeMarketDate(firstTushareString(row, fieldIndex, "maturity_date"))
			}
			items = append(items, marketUniverseSourceItem{
				AssetType:      "CBOND",
				InstrumentKey:  instrumentKey,
				ExternalSymbol: instrumentKey,
				DisplayName:    defaultString(firstTushareString(row, fieldIndex, "bond_short_name"), instrumentKey),
				ExchangeCode:   detectInstrumentExchangeCode(instrumentKey),
				Status:         normalizeMarketUniverseStatus("", delistDate, now),
				ListDate:       normalizeMarketDate(firstTushareString(row, fieldIndex, "list_date")),
				DelistDate:     delistDate,
				MetadataJSON: marshalJSONSilently(filterEmptyJSONMap(map[string]interface{}{
					"source_key":      sourceKey,
					"stock_code":      firstTushareString(row, fieldIndex, "stk_code"),
					"stock_name":      firstTushareString(row, fieldIndex, "stk_short_name"),
					"conv_start_date": normalizeMarketDate(firstTushareString(row, fieldIndex, "conv_start_date")),
					"conv_end_date":   normalizeMarketDate(firstTushareString(row, fieldIndex, "conv_end_date")),
					"maturity_date":   normalizeMarketDate(firstTushareString(row, fieldIndex, "maturity_date")),
					"issue_size":      firstTushareString(row, fieldIndex, "issue_size"),
					"remain_size":     firstTushareString(row, fieldIndex, "remain_size"),
					"rating":          firstTushareString(row, fieldIndex, "rating"),
				})),
			})
		}
	}
	return items, nil
}

func fetchPaginatedTushareUniverseResponses(client *http.Client, token string, apiName string, baseParams map[string]string, fields string) ([]tushareStdResponse, error) {
	const (
		pageSize = 4000
		maxPages = 5
	)
	responses := make([]tushareStdResponse, 0, 2)
	for page := 0; page < maxPages; page++ {
		params := cloneTushareParams(baseParams)
		params["limit"] = strconv.Itoa(pageSize)
		if page > 0 {
			params["offset"] = strconv.Itoa(page * pageSize)
		}
		response, err := callTushareAPI(client, token, apiName, params, fields)
		if err != nil && shouldRetryTushareWithoutFields(err) {
			response, err = callTushareAPI(client, token, apiName, params, "")
		}
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
		if len(response.Data.Items) < pageSize {
			break
		}
	}
	return responses, nil
}

func cloneTushareParams(input map[string]string) map[string]string {
	if len(input) == 0 {
		return make(map[string]string)
	}
	cloned := make(map[string]string, len(input)+2)
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

func classifyTushareFundAssetType(name string, fundType string, investType string, typeValue string) string {
	combined := strings.ToUpper(strings.Join([]string{name, fundType, investType, typeValue}, " "))
	switch {
	case strings.Contains(combined, "LOF") || strings.Contains(name, "上市开放式"):
		return "LOF"
	case strings.Contains(combined, "ETF") || strings.Contains(name, "交易型开放式"):
		return "ETF"
	default:
		return ""
	}
}

func normalizeMarketUniverseStatus(status string, delistDate string, asOf time.Time) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "L", "LISTED", "LISTING", "ACTIVE":
		return "ACTIVE"
	case "P", "PENDING":
		return "PENDING"
	case "D", "DELISTED", "INACTIVE":
		return "DELISTED"
	default:
		return resolveMarketInstrumentStatus("", delistDate, asOf)
	}
}
