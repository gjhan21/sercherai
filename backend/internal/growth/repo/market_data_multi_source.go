package repo

import (
	"bytes"
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

const (
	marketAssetClassStock   = "STOCK"
	marketAssetClassFutures = "FUTURES"

	marketDataKindDailyBars        = "DAILY_BARS"
	marketDataKindNewsItems        = "NEWS_ITEMS"
	marketDataKindFuturesInventory = "FUTURES_INVENTORY"

	marketStockPriorityConfigKey   = "market.stock.daily.source_priority"
	marketFuturesPriorityConfigKey = "market.futures.daily.source_priority"
	marketNewsPriorityConfigKey    = "market.news.source_priority"
)

type marketTouchedBarKey struct {
	InstrumentKey string
	TradeDate     string
}

type pythonBridgeDailyBarPayload struct {
	Items []pythonBridgeDailyBarItem `json:"items"`
}

type pythonBridgeDailyBarItem struct {
	InstrumentKey   string  `json:"instrument_key"`
	ExternalSymbol  string  `json:"external_symbol"`
	TradeDate       string  `json:"trade_date"`
	OpenPrice       float64 `json:"open_price"`
	HighPrice       float64 `json:"high_price"`
	LowPrice        float64 `json:"low_price"`
	ClosePrice      float64 `json:"close_price"`
	PrevClosePrice  float64 `json:"prev_close_price"`
	SettlePrice     float64 `json:"settle_price"`
	PrevSettlePrice float64 `json:"prev_settle_price"`
	Volume          float64 `json:"volume"`
	Turnover        float64 `json:"turnover"`
	OpenInterest    float64 `json:"open_interest"`
}

type pythonBridgeNewsPayload struct {
	Items []pythonBridgeNewsItem `json:"items"`
}

type pythonBridgeNewsItem struct {
	ExternalID    string   `json:"external_id"`
	NewsType      string   `json:"news_type"`
	Title         string   `json:"title"`
	Summary       string   `json:"summary"`
	Content       string   `json:"content"`
	URL           string   `json:"url"`
	PrimarySymbol string   `json:"primary_symbol"`
	Symbols       []string `json:"symbols"`
	PublishedAt   string   `json:"published_at"`
}

type marketSourceRoutingSummary struct {
	SelectedSource     string
	FallbackSourceKeys []string
	RoutingPolicyKey   string
	DecisionReason     string
}

func (r *MySQLGrowthRepo) AdminSyncStockQuotes(sourceKey string, symbols []string, days int) (int, error) {
	result, err := r.AdminSyncStockQuotesDetailed(sourceKey, symbols, days)
	if err != nil {
		return 0, err
	}
	if result.TruthCount > 0 {
		return result.TruthCount, nil
	}
	return result.BarCount, nil
}

func (r *MySQLGrowthRepo) AdminSyncStockQuotesDetailed(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error) {
	symbols = normalizeStockSymbolList(symbols)
	if len(symbols) == 0 {
		symbols = defaultMockStockSymbols()
	}
	if days <= 0 {
		days = 120
	}
	if days > 365 {
		days = 365
	}
	return r.syncMarketDailyBars(
		marketAssetClassStock,
		sourceKey,
		symbols,
		days,
		marketStockPriorityConfigKey,
		[]string{"TUSHARE", "AKSHARE", "TICKERMD", "MOCK"},
	)
}

func (r *MySQLGrowthRepo) AdminSyncFuturesQuotes(sourceKey string, contracts []string, days int) (model.MarketSyncResult, error) {
	contracts = normalizeFuturesContractList(contracts)
	if len(contracts) == 0 {
		contracts = defaultFuturesContracts()
	}
	if days <= 0 {
		days = 120
	}
	if days > 365 {
		days = 365
	}
	return r.syncMarketDailyBars(
		marketAssetClassFutures,
		sourceKey,
		contracts,
		days,
		marketFuturesPriorityConfigKey,
		[]string{"TUSHARE", "TICKERMD", "AKSHARE", "MOCK"},
	)
}

func (r *MySQLGrowthRepo) AdminSyncFuturesInventory(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error) {
	symbols = normalizeFuturesInventorySymbolList(symbols)
	if len(symbols) == 0 {
		symbols = defaultFuturesInventorySymbols()
	}
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}
	sourceKeys := r.resolveRequestedMarketSourceKeysWithGovernance(sourceKey, marketAssetClassFutures, marketDataKindFuturesInventory, "market.futures.inventory.source_priority", []string{"TUSHARE", "MOCK"})
	routingSummary := buildMarketSourceRoutingSummary(sourceKey, sourceKeys, marketAssetClassFutures, marketDataKindFuturesInventory)
	result := model.MarketSyncResult{
		AssetClass:         marketAssetClassFutures,
		DataKind:           marketDataKindFuturesInventory,
		RequestedSourceKey: strings.ToUpper(strings.TrimSpace(sourceKey)),
		ResolvedSourceKeys: sourceKeys,
		SelectedSource:     routingSummary.SelectedSource,
		FallbackSourceKeys: append([]string(nil), routingSummary.FallbackSourceKeys...),
		RoutingPolicyKey:   routingSummary.RoutingPolicyKey,
		DecisionReason:     routingSummary.DecisionReason,
		Results:            make([]model.MarketSourceSyncItemResult, 0, len(sourceKeys)),
	}

	totalInventory := 0
	totalSnapshots := 0
	successes := 0
	failures := make([]string, 0)

	for _, resolvedSourceKey := range sourceKeys {
		sourceItem, err := r.getDataSourceBySourceKey(resolvedSourceKey)
		if err != nil {
			r.insertMarketDataQualityLog(marketAssetClassFutures, marketDataKindFuturesInventory, "", "", resolvedSourceKey, "ERROR", "SOURCE_LOOKUP_FAILED", err.Error(), "")
			failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
			result.Results = append(result.Results, model.MarketSourceSyncItemResult{
				SourceKey: resolvedSourceKey,
				Status:    "FAILED",
				Message:   err.Error(),
			})
			continue
		}

		items, payload, err := r.fetchFuturesInventoryForSource(sourceItem, symbols, days)
		status := "SUCCESS"
		message := "ok"
		inventoryCount := 0
		if err != nil {
			status = "FAILED"
			message = err.Error()
			failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
		} else {
			successes++
			inventoryCount, err = r.upsertFuturesInventorySnapshots(items)
			if err != nil {
				status = "FAILED"
				message = err.Error()
				failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
			} else {
				totalInventory += inventoryCount
			}
		}

		if payload == "" && len(items) > 0 {
			payload = marshalJSONSilently(map[string]interface{}{
				"source_key":  resolvedSourceKey,
				"asset_class": marketAssetClassFutures,
				"data_kind":   marketDataKindFuturesInventory,
				"symbol_n":    len(symbols),
				"days":        days,
				"items":       items,
			})
		}
		if snapshotErr := r.insertMarketSourceSnapshot(
			resolvedSourceKey,
			marketAssetClassFutures,
			marketDataKindFuturesInventory,
			"",
			"",
			status,
			message,
			payload,
			time.Now(),
		); snapshotErr == nil {
			totalSnapshots++
		}

		result.Results = append(result.Results, model.MarketSourceSyncItemResult{
			SourceKey:      resolvedSourceKey,
			Status:         status,
			InventoryCount: inventoryCount,
			SnapshotCount:  1,
			Message:        message,
		})
	}

	result.InventoryCount = totalInventory
	result.SnapshotCount = totalSnapshots
	if successes == 0 && len(failures) > 0 {
		return result, errors.New(strings.Join(failures, "; "))
	}
	return result, nil
}

func (r *MySQLGrowthRepo) AdminSyncMarketNews(sourceKey string, symbols []string, days int, limit int) (model.MarketSyncResult, error) {
	symbols = normalizeStockSymbolList(symbols)
	if days <= 0 {
		days = 7
	}
	if limit <= 0 {
		limit = 50
	}
	sourceKeys := r.resolveRequestedMarketSourceKeysWithGovernance(sourceKey, "", marketDataKindNewsItems, marketNewsPriorityConfigKey, []string{"AKSHARE", "TUSHARE"})
	routingSummary := buildMarketSourceRoutingSummary(sourceKey, sourceKeys, "", marketDataKindNewsItems)
	result := model.MarketSyncResult{
		DataKind:           marketDataKindNewsItems,
		RequestedSourceKey: strings.ToUpper(strings.TrimSpace(sourceKey)),
		ResolvedSourceKeys: sourceKeys,
		SelectedSource:     routingSummary.SelectedSource,
		FallbackSourceKeys: append([]string(nil), routingSummary.FallbackSourceKeys...),
		RoutingPolicyKey:   routingSummary.RoutingPolicyKey,
		DecisionReason:     routingSummary.DecisionReason,
		Results:            make([]model.MarketSourceSyncItemResult, 0, len(sourceKeys)),
	}

	totalNews := 0
	totalSnapshots := 0
	successes := 0
	failures := make([]string, 0)

	for _, resolvedSourceKey := range sourceKeys {
		sourceItem, err := r.getDataSourceBySourceKey(resolvedSourceKey)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
			result.Results = append(result.Results, model.MarketSourceSyncItemResult{
				SourceKey: resolvedSourceKey,
				Status:    "FAILED",
				Message:   err.Error(),
			})
			continue
		}

		items, payload, err := r.fetchMarketNewsForSource(sourceItem, symbols, days, limit)
		status := "SUCCESS"
		message := "ok"
		newsCount := 0
		if err != nil {
			status = "FAILED"
			message = err.Error()
			failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
		} else {
			successes++
			newsCount, err = r.upsertMarketNewsItems(items)
			if err != nil {
				status = "FAILED"
				message = err.Error()
				failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
			} else {
				totalNews += newsCount
			}
		}

		if payload == "" && len(items) > 0 {
			payload = marshalJSONSilently(map[string]interface{}{
				"source_key": resolvedSourceKey,
				"data_kind":  marketDataKindNewsItems,
				"days":       days,
				"limit":      limit,
				"items":      items,
			})
		}
		if snapshotErr := r.insertMarketSourceSnapshot(
			resolvedSourceKey,
			"",
			marketDataKindNewsItems,
			"",
			"",
			status,
			message,
			payload,
			time.Now(),
		); snapshotErr == nil {
			totalSnapshots++
		}

		result.Results = append(result.Results, model.MarketSourceSyncItemResult{
			SourceKey:     resolvedSourceKey,
			Status:        status,
			NewsCount:     newsCount,
			SnapshotCount: 1,
			Message:       message,
		})
	}

	result.NewsCount = totalNews
	result.SnapshotCount = totalSnapshots
	if successes == 0 && len(failures) > 0 {
		return result, errors.New(strings.Join(failures, "; "))
	}
	return result, nil
}

func (r *MySQLGrowthRepo) syncMarketDailyBars(assetClass string, sourceKey string, instrumentKeys []string, days int, routeConfigKey string, defaultPriority []string) (model.MarketSyncResult, error) {
	sourceKeys := r.resolveRequestedMarketSourceKeysWithGovernance(sourceKey, assetClass, marketDataKindDailyBars, routeConfigKey, defaultPriority)
	routingSummary := buildMarketSourceRoutingSummary(sourceKey, sourceKeys, assetClass, marketDataKindDailyBars)
	result := model.MarketSyncResult{
		AssetClass:         assetClass,
		DataKind:           marketDataKindDailyBars,
		RequestedSourceKey: strings.ToUpper(strings.TrimSpace(sourceKey)),
		ResolvedSourceKeys: sourceKeys,
		SelectedSource:     routingSummary.SelectedSource,
		FallbackSourceKeys: append([]string(nil), routingSummary.FallbackSourceKeys...),
		RoutingPolicyKey:   routingSummary.RoutingPolicyKey,
		DecisionReason:     routingSummary.DecisionReason,
		Results:            make([]model.MarketSourceSyncItemResult, 0, len(sourceKeys)),
	}

	if len(instrumentKeys) == 0 {
		return result, errors.New("instrument list is empty")
	}
	if err := r.syncMarketInstrumentMasterData(assetClass, sourceKey, instrumentKeys); err != nil {
		return result, err
	}

	totalBars := 0
	totalSnapshots := 0
	failures := make([]string, 0)
	touched := make(map[string]marketTouchedBarKey)
	successes := 0

	for _, resolvedSourceKey := range sourceKeys {
		sourceItem, err := r.getDataSourceBySourceKey(resolvedSourceKey)
		if err != nil {
			failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
			result.Results = append(result.Results, model.MarketSourceSyncItemResult{
				SourceKey: resolvedSourceKey,
				Status:    "FAILED",
				Message:   err.Error(),
			})
			continue
		}
		externalSymbols, err := r.resolveMarketExternalSymbols(resolvedSourceKey, assetClass, instrumentKeys)
		if err != nil {
			r.insertMarketDataQualityLog(assetClass, marketDataKindDailyBars, "", "", resolvedSourceKey, "ERROR", "SYMBOL_ALIAS_RESOLVE_FAILED", err.Error(), "")
			failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
			result.Results = append(result.Results, model.MarketSourceSyncItemResult{
				SourceKey: resolvedSourceKey,
				Status:    "FAILED",
				Message:   err.Error(),
			})
			continue
		}

		bars, payload, err := r.fetchMarketDailyBarsForSource(sourceItem, assetClass, instrumentKeys, externalSymbols, days)
		status := "SUCCESS"
		message := "ok"
		barCount := 0
		if err != nil {
			status = "FAILED"
			message = err.Error()
			r.insertMarketDataQualityLog(assetClass, marketDataKindDailyBars, "", "", resolvedSourceKey, "ERROR", "SOURCE_FETCH_FAILED", err.Error(), "")
			failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
		} else {
			successes++
			barCount, err = r.upsertMarketDailyBars(bars)
			if err != nil {
				status = "FAILED"
				message = err.Error()
				r.insertMarketDataQualityLog(assetClass, marketDataKindDailyBars, "", "", resolvedSourceKey, "ERROR", "BAR_UPSERT_FAILED", err.Error(), "")
				failures = append(failures, fmt.Sprintf("%s: %v", resolvedSourceKey, err))
			} else {
				totalBars += barCount
				for _, bar := range bars {
					if bar.InstrumentKey == "" || bar.TradeDate == "" {
						continue
					}
					mapKey := bar.InstrumentKey + "|" + bar.TradeDate
					touched[mapKey] = marketTouchedBarKey{
						InstrumentKey: bar.InstrumentKey,
						TradeDate:     bar.TradeDate,
					}
				}
			}
		}

		if payload == "" && len(bars) > 0 {
			payload = marshalJSONSilently(map[string]interface{}{
				"source_key":   resolvedSourceKey,
				"asset_class":  assetClass,
				"data_kind":    marketDataKindDailyBars,
				"instrument_n": len(instrumentKeys),
				"days":         days,
				"items":        bars,
			})
		}
		if snapshotErr := r.insertMarketSourceSnapshot(
			resolvedSourceKey,
			assetClass,
			marketDataKindDailyBars,
			"",
			"",
			status,
			message,
			payload,
			time.Now(),
		); snapshotErr == nil {
			totalSnapshots++
		}
		result.Results = append(result.Results, model.MarketSourceSyncItemResult{
			SourceKey:     resolvedSourceKey,
			Status:        status,
			BarCount:      barCount,
			SnapshotCount: 1,
			Message:       message,
		})
	}

	result.BarCount = totalBars
	result.SnapshotCount = totalSnapshots

	if len(touched) > 0 {
		truthBars, err := r.rebuildMarketDailyBarTruth(assetClass, touched, r.loadGovernedMarketSourcePriority(assetClass, marketDataKindDailyBars, routeConfigKey, defaultPriority))
		if err != nil {
			return result, err
		}
		result.TruthCount = len(truthBars)
		if assetClass == marketAssetClassStock && len(truthBars) > 0 {
			if _, err := r.rebuildStockStatusTruth(truthBars); err != nil && !isMarketStatusSchemaCompatError(err) {
				return result, err
			}
			if err := r.syncLegacyStockQuotesFromTruthBars(truthBars); err != nil {
				return result, err
			}
		}
		if assetClass == marketAssetClassFutures && len(truthBars) > 0 {
			if _, err := r.rebuildFuturesContractMappings(truthBars); err != nil && !isMarketStatusSchemaCompatError(err) {
				return result, err
			}
		}
	}

	if successes == 0 && len(failures) > 0 {
		return result, errors.New(strings.Join(failures, "; "))
	}
	return result, nil
}

type marketStockDateRangeSyncOptions struct {
	EnsureMasterSync bool
	RebuildTruth     bool
}

func (r *MySQLGrowthRepo) syncStockMarketDailyBarsByDateRange(sourceKey string, instrumentKeys []string, tradeDateFrom string, tradeDateTo string, options marketStockDateRangeSyncOptions) (model.MarketSyncResult, map[string]marketTouchedBarKey, error) {
	resolvedSourceKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	if resolvedSourceKey == "" {
		resolvedSourceKey = "TUSHARE"
	}
	result := model.MarketSyncResult{
		AssetClass:         marketAssetClassStock,
		DataKind:           marketDataKindDailyBars,
		RequestedSourceKey: resolvedSourceKey,
		ResolvedSourceKeys: []string{resolvedSourceKey},
		Results:            make([]model.MarketSourceSyncItemResult, 0, 1),
	}
	touched := make(map[string]marketTouchedBarKey)
	if len(instrumentKeys) == 0 {
		return result, touched, errors.New("instrument list is empty")
	}
	if options.EnsureMasterSync {
		if err := r.syncMarketInstrumentMasterData(marketAssetClassStock, resolvedSourceKey, instrumentKeys); err != nil {
			return result, touched, err
		}
	}

	sourceItem, err := r.getDataSourceBySourceKey(resolvedSourceKey)
	if err != nil {
		return result, touched, err
	}
	externalSymbols, err := r.resolveMarketExternalSymbols(resolvedSourceKey, marketAssetClassStock, instrumentKeys)
	if err != nil {
		r.insertMarketDataQualityLog(marketAssetClassStock, marketDataKindDailyBars, "", "", resolvedSourceKey, "ERROR", "SYMBOL_ALIAS_RESOLVE_FAILED", err.Error(), "")
		return result, touched, err
	}

	bars, payload, err := r.fetchMarketDailyBarsForSourceDateRange(sourceItem, marketAssetClassStock, instrumentKeys, externalSymbols, tradeDateFrom, tradeDateTo)
	status := "SUCCESS"
	message := "ok"
	barCount := 0
	if err != nil {
		status = "FAILED"
		message = err.Error()
		r.insertMarketDataQualityLog(marketAssetClassStock, marketDataKindDailyBars, "", "", resolvedSourceKey, "ERROR", "SOURCE_FETCH_FAILED", err.Error(), "")
	} else {
		barCount, err = r.upsertMarketDailyBars(bars)
		if err != nil {
			status = "FAILED"
			message = err.Error()
			r.insertMarketDataQualityLog(marketAssetClassStock, marketDataKindDailyBars, "", "", resolvedSourceKey, "ERROR", "BAR_UPSERT_FAILED", err.Error(), "")
		} else {
			result.BarCount = barCount
		}
	}

	if payload == "" && len(bars) > 0 {
		payload = marshalJSONSilently(map[string]interface{}{
			"source_key":      resolvedSourceKey,
			"asset_class":     marketAssetClassStock,
			"data_kind":       marketDataKindDailyBars,
			"trade_date_from": tradeDateFrom,
			"trade_date_to":   tradeDateTo,
			"instrument_n":    len(instrumentKeys),
			"items":           bars,
		})
	}
	if status == "SUCCESS" {
		touched = buildTouchedBarKeysFromBars(bars)
		if options.RebuildTruth && len(touched) > 0 {
			truthBars, truthErr := r.rebuildMarketDailyBarTruth(marketAssetClassStock, touched, []string{resolvedSourceKey})
			if truthErr != nil {
				return result, touched, truthErr
			}
			result.TruthCount = len(truthBars)
			if len(truthBars) > 0 {
				if _, truthErr := r.rebuildStockStatusTruth(truthBars); truthErr != nil && !isMarketStatusSchemaCompatError(truthErr) {
					return result, touched, truthErr
				}
				if truthErr := r.syncLegacyStockQuotesFromTruthBars(truthBars); truthErr != nil {
					return result, touched, truthErr
				}
			}
		}
	}

	if snapshotErr := r.insertMarketSourceSnapshot(
		resolvedSourceKey,
		marketAssetClassStock,
		marketDataKindDailyBars,
		"",
		"",
		status,
		message,
		payload,
		time.Now(),
	); snapshotErr == nil {
		result.SnapshotCount = 1
	}
	result.Results = append(result.Results, model.MarketSourceSyncItemResult{
		SourceKey:     resolvedSourceKey,
		Status:        status,
		BarCount:      result.BarCount,
		SnapshotCount: result.SnapshotCount,
		TruthCount:    result.TruthCount,
		Message:       message,
	})
	if status == "FAILED" {
		return result, touched, errors.New(message)
	}
	return result, touched, nil
}

func normalizeFuturesContractList(contracts []string) []string {
	seen := make(map[string]struct{}, len(contracts))
	items := make([]string, 0, len(contracts))
	for _, contract := range contracts {
		normalized := strings.ToUpper(strings.TrimSpace(contract))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	return items
}

func normalizeFuturesInventorySymbolList(symbols []string) []string {
	seen := make(map[string]struct{}, len(symbols))
	items := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		letters := make([]rune, 0, len(normalized))
		for _, ch := range normalized {
			if ch >= 'A' && ch <= 'Z' {
				letters = append(letters, ch)
				continue
			}
			break
		}
		if len(letters) > 0 {
			normalized = string(letters)
		}
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	return items
}

func defaultFuturesContracts() []string {
	return []string{
		"AU2406.SHF",
		"AG2406.SHF",
		"RB2406.SHF",
		"IF2406.CFX",
	}
}

func defaultFuturesInventorySymbols() []string {
	return []string{"RB", "CU", "AL", "RU", "AU"}
}

func buildMarketSourceRoutingSummary(requestedSourceKey string, resolvedSourceKeys []string, assetClass string, dataKind string) marketSourceRoutingSummary {
	selectedSource := ""
	if len(resolvedSourceKeys) > 0 {
		selectedSource = strings.ToUpper(strings.TrimSpace(resolvedSourceKeys[0]))
	}
	summary := marketSourceRoutingSummary{
		SelectedSource:     selectedSource,
		FallbackSourceKeys: appendRemainingSourceKeys(selectedSource, resolvedSourceKeys),
	}
	if policy, ok := findDefaultMarketProviderRoutingPolicy(assetClass, dataKind); ok {
		summary.RoutingPolicyKey = policy.PolicyKey
	}

	normalizedRequested := strings.ToUpper(strings.TrimSpace(requestedSourceKey))
	switch normalizedRequested {
	case "":
		if summary.RoutingPolicyKey != "" {
			summary.DecisionReason = "default_primary_source"
		} else {
			summary.DecisionReason = "legacy_default_priority"
		}
	case "AUTO":
		if summary.RoutingPolicyKey != "" {
			summary.DecisionReason = "governed_auto_priority"
		} else {
			summary.DecisionReason = "legacy_auto_priority"
		}
	default:
		summary.DecisionReason = "explicit_source"
		if summary.SelectedSource == "" {
			summary.SelectedSource = normalizedRequested
		}
	}
	return summary
}

func buildStrategyContextRoutingSummary(selectedSource string, assetClass string, dataKind string) marketSourceRoutingSummary {
	normalizedSelected := strings.ToUpper(strings.TrimSpace(selectedSource))
	summary := marketSourceRoutingSummary{SelectedSource: normalizedSelected}
	if normalizedSelected == "" {
		return summary
	}
	if policy, ok := findDefaultMarketProviderRoutingPolicy(assetClass, dataKind); ok {
		summary.RoutingPolicyKey = policy.PolicyKey
		governedChain := make([]string, 0, 1+len(policy.FallbackProviderKeys))
		governedChain = append(governedChain, policy.PrimaryProviderKey)
		governedChain = append(governedChain, policy.FallbackProviderKeys...)
		summary.FallbackSourceKeys = appendRemainingSourceKeys(normalizedSelected, governedChain)
	}
	if normalizedSelected == "MOCK" {
		summary.DecisionReason = "mock_truth_fallback"
	} else {
		summary.DecisionReason = "local_truth_price_source"
	}
	return summary
}

func appendRemainingSourceKeys(selectedSource string, sourceKeys []string) []string {
	normalizedSelected := strings.ToUpper(strings.TrimSpace(selectedSource))
	seen := make(map[string]struct{}, len(sourceKeys))
	items := make([]string, 0, len(sourceKeys))
	for _, item := range sourceKeys {
		normalized := strings.ToUpper(strings.TrimSpace(item))
		if normalized == "" || normalized == normalizedSelected {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		items = append(items, normalized)
	}
	return items
}

func (r *MySQLGrowthRepo) resolveRequestedMarketSourceKeys(raw string, routeConfigKey string, defaultPriority []string) []string {
	return r.resolveRequestedMarketSourceKeysWithGovernance(raw, "", "", routeConfigKey, defaultPriority)
}

func (r *MySQLGrowthRepo) resolveRequestedMarketSourceKeysWithGovernance(raw string, assetClass string, dataKind string, routeConfigKey string, defaultPriority []string) []string {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	if normalized == "" {
		priority := r.loadGovernedMarketSourcePriority(assetClass, dataKind, routeConfigKey, defaultPriority)
		if len(priority) > 0 {
			return []string{priority[0]}
		}
		return []string{"TUSHARE"}
	}
	if normalized == "AUTO" {
		return r.loadGovernedMarketSourcePriority(assetClass, dataKind, routeConfigKey, defaultPriority)
	}
	parts := strings.FieldsFunc(normalized, func(ch rune) bool {
		return ch == ',' || ch == ';' || ch == '|' || ch == ' '
	})
	seen := make(map[string]struct{}, len(parts))
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.ToUpper(strings.TrimSpace(part))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	if len(items) > 0 {
		return items
	}
	return r.loadGovernedMarketSourcePriority(assetClass, dataKind, routeConfigKey, defaultPriority)
}

func (r *MySQLGrowthRepo) loadGovernedMarketSourcePriority(assetClass string, dataKind string, routeConfigKey string, fallback []string) []string {
	normalizedAssetClass := normalizeMarketProviderFilter(assetClass)
	normalizedDataKind := normalizeMarketProviderFilter(dataKind)
	if normalizedDataKind != "" {
		query := `
SELECT primary_provider_key,
       COALESCE(CAST(fallback_provider_keys_json AS CHAR), ''),
       fallback_allowed,
       mock_allowed
FROM market_provider_routing_policies
WHERE asset_class = ? AND data_kind = ?
LIMIT 1`
		var (
			primaryProviderKey    sql.NullString
			fallbackProvidersJSON sql.NullString
			fallbackAllowed       bool
			mockAllowed           bool
		)
		err := r.db.QueryRow(query, normalizedAssetClass, normalizedDataKind).Scan(
			&primaryProviderKey,
			&fallbackProvidersJSON,
			&fallbackAllowed,
			&mockAllowed,
		)
		if err == nil {
			items := make([]string, 0, 4)
			seen := make(map[string]struct{}, 4)
			appendItem := func(value string) {
				value = strings.ToUpper(strings.TrimSpace(value))
				if value == "" {
					return
				}
				if value == "MOCK" && !mockAllowed {
					return
				}
				if _, ok := seen[value]; ok {
					return
				}
				seen[value] = struct{}{}
				items = append(items, value)
			}
			appendItem(primaryProviderKey.String)
			if fallbackAllowed {
				for _, item := range parseJSONStringList(fallbackProvidersJSON.String) {
					appendItem(item)
				}
			}
			if len(items) > 0 {
				return items
			}
		} else if err != sql.ErrNoRows && !isMarketProviderGovernanceSchemaCompatError(err) {
			return r.loadMarketSourcePriority(routeConfigKey, fallback)
		}
	}
	return r.loadMarketSourcePriority(routeConfigKey, fallback)
}

func (r *MySQLGrowthRepo) loadMarketSourcePriority(configKey string, fallback []string) []string {
	if strings.TrimSpace(configKey) == "" {
		return append([]string(nil), fallback...)
	}
	var raw string
	err := r.db.QueryRow(`
SELECT config_value
FROM system_configs
WHERE LOWER(config_key) = LOWER(?)
LIMIT 1`, configKey).Scan(&raw)
	if err != nil {
		return append([]string(nil), fallback...)
	}
	values := splitSourcePriorityList(raw)
	if len(values) == 0 {
		return append([]string(nil), fallback...)
	}
	return values
}

func splitSourcePriorityList(raw string) []string {
	parts := strings.FieldsFunc(strings.ToUpper(strings.TrimSpace(raw)), func(ch rune) bool {
		return ch == ',' || ch == ';' || ch == '|' || ch == ' '
	})
	seen := make(map[string]struct{}, len(parts))
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		items = append(items, part)
	}
	return items
}

func (r *MySQLGrowthRepo) upsertMarketInstruments(assetClass string, instrumentKeys []string) error {
	now := time.Now()
	for _, instrumentKey := range instrumentKeys {
		normalized := strings.ToUpper(strings.TrimSpace(instrumentKey))
		if normalized == "" {
			continue
		}
		exchangeCode := detectInstrumentExchangeCode(normalized)
		_, err := r.db.Exec(`
INSERT INTO market_instruments (id, asset_class, instrument_key, display_name, exchange_code, status, metadata_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 'ACTIVE', NULL, ?, ?)
ON DUPLICATE KEY UPDATE
  exchange_code = VALUES(exchange_code),
  updated_at = VALUES(updated_at)`,
			newID("mdi"),
			assetClass,
			normalized,
			normalized,
			nullableString(exchangeCode),
			now,
			now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func detectInstrumentExchangeCode(instrumentKey string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(instrumentKey))
	if idx := strings.LastIndex(trimmed, "."); idx >= 0 && idx < len(trimmed)-1 {
		return trimmed[idx+1:]
	}
	return ""
}

func (r *MySQLGrowthRepo) resolveMarketExternalSymbols(sourceKey string, assetClass string, instrumentKeys []string) (map[string]string, error) {
	result := make(map[string]string, len(instrumentKeys))
	if len(instrumentKeys) == 0 {
		return result, nil
	}
	holders := make([]string, 0, len(instrumentKeys))
	args := make([]interface{}, 0, len(instrumentKeys)+2)
	args = append(args, assetClass, strings.ToUpper(strings.TrimSpace(sourceKey)))
	for _, instrumentKey := range instrumentKeys {
		holders = append(holders, "?")
		args = append(args, strings.ToUpper(strings.TrimSpace(instrumentKey)))
	}
	query := `
SELECT instrument_key, external_symbol
FROM market_symbol_aliases
WHERE asset_class = ? AND UPPER(source_key) = ? AND status = 'ACTIVE' AND instrument_key IN (` + strings.Join(holders, ",") + `)`
	rows, err := r.db.Query(query, args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var instrumentKey string
			var externalSymbol string
			if scanErr := rows.Scan(&instrumentKey, &externalSymbol); scanErr != nil {
				return nil, scanErr
			}
			result[strings.ToUpper(strings.TrimSpace(instrumentKey))] = strings.TrimSpace(externalSymbol)
		}
	}
	for _, instrumentKey := range instrumentKeys {
		normalizedKey := strings.ToUpper(strings.TrimSpace(instrumentKey))
		if normalizedKey == "" {
			continue
		}
		if _, ok := result[normalizedKey]; ok {
			continue
		}
		result[normalizedKey] = deriveDefaultExternalSymbol(strings.ToUpper(strings.TrimSpace(sourceKey)), assetClass, normalizedKey)
	}
	return result, nil
}

func deriveDefaultExternalSymbol(sourceKey string, assetClass string, instrumentKey string) string {
	normalizedKey := strings.ToUpper(strings.TrimSpace(instrumentKey))
	switch assetClass {
	case marketAssetClassStock:
		code := normalizedKey
		exchangeCode := ""
		if idx := strings.LastIndex(normalizedKey, "."); idx >= 0 {
			code = normalizedKey[:idx]
			exchangeCode = normalizedKey[idx+1:]
		}
		switch sourceKey {
		case "AKSHARE":
			return code
		case "TICKERMD":
			prefix := strings.ToLower(exchangeCode)
			if prefix == "" {
				return strings.ToLower(code)
			}
			return prefix + strings.ToLower(code)
		default:
			return normalizedKey
		}
	case marketAssetClassFutures:
		switch sourceKey {
		case "AKSHARE", "TICKERMD":
			if idx := strings.LastIndex(normalizedKey, "."); idx >= 0 {
				return strings.ToLower(normalizedKey[:idx])
			}
			return strings.ToLower(normalizedKey)
		default:
			return normalizedKey
		}
	default:
		return normalizedKey
	}
}

func canonicalMarketSourceKey(sourceKey string, provider string) string {
	normalizedKey := strings.ToUpper(strings.TrimSpace(sourceKey))
	normalizedProvider := strings.ToUpper(strings.TrimSpace(provider))
	switch normalizedProvider {
	case "MOCK", "TUSHARE", "AKSHARE", "TICKERMD":
		return normalizedProvider
	}
	if normalizedKey != "" {
		return normalizedKey
	}
	return normalizedProvider
}

func (r *MySQLGrowthRepo) fetchMarketDailyBarsForSource(item model.DataSource, assetClass string, instrumentKeys []string, externalSymbols map[string]string, days int) ([]model.MarketDailyBar, string, error) {
	sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
	provider := strings.ToUpper(parseDataSourceStringConfig(item.Config, "provider", "vendor"))
	if provider == "" {
		provider = sourceKey
	}
	sourceKey = canonicalMarketSourceKey(sourceKey, provider)
	switch provider {
	case "MOCK":
		return buildMockMarketDailyBars(assetClass, sourceKey, instrumentKeys, days), "", nil
	case "TUSHARE":
		token := parseDataSourceStringConfig(item.Config, "token", "api_token", "tushare_token")
		if strings.TrimSpace(token) == "" {
			token = strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
		}
		timeoutMS := parseDataSourceTimeoutMS(item.Config)
		if assetClass == marketAssetClassStock {
			return fetchStockMarketBarsFromTushare(token, sourceKey, instrumentKeys, externalSymbols, days, timeoutMS)
		}
		return fetchFuturesMarketBarsFromTushare(token, sourceKey, instrumentKeys, externalSymbols, days, timeoutMS)
	case "AKSHARE":
		if assetClass == marketAssetClassFutures {
			return fetchFuturesMarketBarsFromAkshareBridge(item.Config, sourceKey, instrumentKeys, externalSymbols, days)
		}
		return fetchStockMarketBarsFromAkshareBridge(item.Config, sourceKey, instrumentKeys, externalSymbols, days)
	case "MYSELF":
		return fetchMarketBarsFromMyself(item.Config, sourceKey, assetClass, instrumentKeys, externalSymbols, days)
	case "TICKERMD":
		return fetchMarketBarsFromTickerMD(item.Config, sourceKey, assetClass, instrumentKeys, externalSymbols, days)
	default:
		if assetClass == marketAssetClassStock {
			items := make([]string, 0, len(instrumentKeys))
			for _, instrumentKey := range instrumentKeys {
				if external := strings.TrimSpace(externalSymbols[instrumentKey]); external != "" {
					items = append(items, external)
				}
			}
			quotes, err := fetchStockQuotesFromEndpoint(parseDataSourceStringConfig(item.Config, "quotes_endpoint", "endpoint"), sourceKey, items, days, parseDataSourceTimeoutMS(item.Config))
			if err != nil {
				return nil, "", err
			}
			return convertStockQuotesToMarketBars(quotes, instrumentKeys, externalSymbols), "", nil
		}
		return nil, "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (r *MySQLGrowthRepo) fetchMarketDailyBarsForSourceDateRange(item model.DataSource, assetClass string, instrumentKeys []string, externalSymbols map[string]string, tradeDateFrom string, tradeDateTo string) ([]model.MarketDailyBar, string, error) {
	sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
	provider := strings.ToUpper(parseDataSourceStringConfig(item.Config, "provider", "vendor"))
	if provider == "" {
		provider = sourceKey
	}
	sourceKey = canonicalMarketSourceKey(sourceKey, provider)
	if assetClass == marketAssetClassStock && provider == "TUSHARE" {
		token := parseDataSourceStringConfig(item.Config, "token", "api_token", "tushare_token")
		if strings.TrimSpace(token) == "" {
			token = strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
		}
		timeoutMS := parseDataSourceTimeoutMS(item.Config)
		return fetchStockMarketBarsFromTushareDateRange(token, sourceKey, instrumentKeys, externalSymbols, tradeDateFrom, tradeDateTo, timeoutMS)
	}
	return nil, "", fmt.Errorf("long history quote backfill only supports STOCK + TUSHARE")
}

func (r *MySQLGrowthRepo) fetchFuturesInventoryForSource(item model.DataSource, symbols []string, days int) ([]model.FuturesInventorySnapshot, string, error) {
	sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
	provider := strings.ToUpper(parseDataSourceStringConfig(item.Config, "provider", "vendor"))
	if provider == "" {
		provider = sourceKey
	}
	sourceKey = canonicalMarketSourceKey(sourceKey, provider)
	switch provider {
	case "MOCK":
		return buildMockFuturesInventorySnapshots(sourceKey, symbols, days), "", nil
	case "TUSHARE":
		token := parseDataSourceStringConfig(item.Config, "token", "api_token", "tushare_token")
		if strings.TrimSpace(token) == "" {
			token = strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
		}
		return fetchFuturesInventoryFromTushare(token, sourceKey, symbols, days, parseDataSourceTimeoutMS(item.Config))
	default:
		return nil, "", fmt.Errorf("unsupported futures inventory provider: %s", provider)
	}
}

func fetchFuturesInventoryFromTushare(token string, sourceKey string, symbols []string, days int, timeoutMS int) ([]model.FuturesInventorySnapshot, string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, "", errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 30
	}
	startDate := time.Now().AddDate(0, 0, -(days + 10)).Format("20060102")
	endDate := time.Now().Format("20060102")
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	items := make([]model.FuturesInventorySnapshot, 0, len(symbols)*days)
	rawSnapshots := make([]map[string]interface{}, 0, len(symbols))

	for _, symbol := range symbols {
		parsed, err := callTushareAPI(client, token, "fut_wsr", map[string]string{
			"symbol":     symbol,
			"start_date": startDate,
			"end_date":   endDate,
		}, "trade_date,symbol,fut_name,warehouse,wh_id,vol,unit,area,brand,place,grade,pre_vol,vol_chg")
		if err != nil {
			return nil, "", err
		}
		rawSnapshots = append(rawSnapshots, map[string]interface{}{
			"symbol":      symbol,
			"field_count": len(parsed.Data.Fields),
			"item_count":  len(parsed.Data.Items),
		})
		fieldIndex := make(map[string]int, len(parsed.Data.Fields))
		for idx, field := range parsed.Data.Fields {
			fieldIndex[strings.TrimSpace(field)] = idx
		}
		for _, row := range parsed.Data.Items {
			tradeDateRaw, ok := tushareGetString(row, fieldIndex, "trade_date")
			if !ok {
				continue
			}
			tradeDate, err := time.ParseInLocation("20060102", tradeDateRaw, time.Local)
			if err != nil {
				continue
			}
			snapshotSymbol, _ := tushareGetString(row, fieldIndex, "symbol")
			if strings.TrimSpace(snapshotSymbol) == "" {
				snapshotSymbol = symbol
			}
			futuresName, _ := tushareGetString(row, fieldIndex, "fut_name")
			warehouse, _ := tushareGetString(row, fieldIndex, "warehouse")
			warehouseID, _ := tushareGetString(row, fieldIndex, "wh_id")
			unit, _ := tushareGetString(row, fieldIndex, "unit")
			area, _ := tushareGetString(row, fieldIndex, "area")
			brand, _ := tushareGetString(row, fieldIndex, "brand")
			place, _ := tushareGetString(row, fieldIndex, "place")
			grade, _ := tushareGetString(row, fieldIndex, "grade")
			receiptVolume, _ := tushareGetFloat(row, fieldIndex, "vol")
			previousVolume, _ := tushareGetFloat(row, fieldIndex, "pre_vol")
			changeVolume, _ := tushareGetFloat(row, fieldIndex, "vol_chg")
			if previousVolume <= 0 && receiptVolume > 0 && changeVolume != 0 {
				previousVolume = receiptVolume - changeVolume
			}
			items = append(items, model.FuturesInventorySnapshot{
				Symbol:         strings.ToUpper(strings.TrimSpace(snapshotSymbol)),
				TradeDate:      tradeDate.Format("2006-01-02"),
				FuturesName:    strings.TrimSpace(futuresName),
				Warehouse:      strings.TrimSpace(warehouse),
				WarehouseID:    strings.TrimSpace(warehouseID),
				Area:           strings.TrimSpace(area),
				Brand:          strings.TrimSpace(brand),
				Place:          strings.TrimSpace(place),
				Grade:          strings.TrimSpace(grade),
				Unit:           strings.TrimSpace(unit),
				ReceiptVolume:  roundTo(receiptVolume, 4),
				PreviousVolume: roundTo(previousVolume, 4),
				ChangeVolume:   roundTo(changeVolume, 4),
				SourceKey:      sourceKey,
			})
		}
	}
	return items, marshalJSONSilently(map[string]interface{}{
		"source_key": sourceKey,
		"provider":   "TUSHARE",
		"asset":      marketAssetClassFutures,
		"data_kind":  marketDataKindFuturesInventory,
		"items":      rawSnapshots,
	}), nil
}

func buildMockFuturesInventorySnapshots(sourceKey string, symbols []string, days int) []model.FuturesInventorySnapshot {
	if days <= 0 {
		days = 10
	}
	if len(symbols) == 0 {
		symbols = defaultFuturesInventorySymbols()
	}
	start := time.Now().AddDate(0, 0, -(minInt(days, 10) - 1))
	items := make([]model.FuturesInventorySnapshot, 0, len(symbols)*minInt(days, 10))
	for symbolIndex, symbol := range symbols {
		baseLevel := 1800.0 + float64(symbolIndex)*420
		trendStep := float64((symbolIndex%3)-1) * 38
		for dayIndex := 0; dayIndex < minInt(days, 10); dayIndex++ {
			tradeDate := start.AddDate(0, 0, dayIndex)
			previousLevel := baseLevel + float64(maxInt(dayIndex-1, 0))*trendStep
			level := baseLevel + float64(dayIndex)*trendStep
			change := level - previousLevel
			items = append(items, model.FuturesInventorySnapshot{
				Symbol:         strings.ToUpper(strings.TrimSpace(symbol)),
				TradeDate:      tradeDate.Format("2006-01-02"),
				FuturesName:    strings.ToUpper(strings.TrimSpace(symbol)) + " 仓单",
				Warehouse:      "MOCK_WAREHOUSE",
				WarehouseID:    "MOCK",
				Area:           "MOCK",
				Brand:          fmt.Sprintf("MOCK_BRAND_%d", symbolIndex+1),
				Place:          "MOCK_PLACE",
				Grade:          "STANDARD",
				Unit:           "手",
				ReceiptVolume:  roundTo(level, 4),
				PreviousVolume: roundTo(previousLevel, 4),
				ChangeVolume:   roundTo(change, 4),
				SourceKey:      sourceKey,
			})
		}
	}
	return items
}

func fetchMarketBarsFromMyself(config map[string]interface{}, sourceKey string, assetClass string, instrumentKeys []string, externalSymbols map[string]string, days int) ([]model.MarketDailyBar, string, error) {
	switch assetClass {
	case marketAssetClassStock:
		quotes, payload, err := fetchStockQuotesFromMyself(config, sourceKey, instrumentKeys, externalSymbols, days)
		if err != nil {
			return nil, "", err
		}
		return convertStockQuotesToMarketBars(quotes, instrumentKeys, nil), payload, nil
	case marketAssetClassFutures:
		return fetchFuturesMarketBarsFromMyself(config, sourceKey, instrumentKeys, externalSymbols, days)
	default:
		return nil, "", fmt.Errorf("unsupported myself asset class: %s", assetClass)
	}
}

func fetchStockQuotesFromMyself(config map[string]interface{}, sourceKey string, instrumentKeys []string, externalSymbols map[string]string, days int) ([]model.StockMarketQuote, string, error) {
	timeoutMS := parseDataSourceTimeoutMS(config)
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	items := make([]model.StockMarketQuote, 0, len(instrumentKeys)*maxInt(days, 1))
	summary := make([]map[string]interface{}, 0, len(instrumentKeys))
	failures := make([]string, 0)

	for _, instrumentKey := range instrumentKeys {
		apiSymbol := myselfStockAPISymbol(instrumentKey, externalSymbols[instrumentKey])
		if apiSymbol == "" {
			failures = append(failures, instrumentKey+": invalid stock symbol")
			continue
		}
		quotes, provider, err := fetchStockQuotesFromMyselfTencent(config, sourceKey, strings.ToUpper(strings.TrimSpace(instrumentKey)), apiSymbol, days, timeoutMS)
		if err != nil || len(quotes) == 0 {
			fallbackQuotes, fallbackErr := fetchStockQuotesFromMyselfSina(config, sourceKey, strings.ToUpper(strings.TrimSpace(instrumentKey)), apiSymbol, days, timeoutMS)
			if fallbackErr != nil {
				message := fallbackErr.Error()
				if err != nil {
					message = err.Error() + "; sina fallback: " + fallbackErr.Error()
				}
				failures = append(failures, instrumentKey+": "+message)
				summary = append(summary, map[string]interface{}{
					"instrument_key": instrumentKey,
					"api_symbol":     apiSymbol,
					"provider":       "MYSELF",
					"status":         "FAILED",
					"message":        message,
				})
				continue
			}
			quotes = fallbackQuotes
			provider = "SINA_STOCK_KLINE"
		}
		items = append(items, quotes...)
		summary = append(summary, map[string]interface{}{
			"instrument_key": instrumentKey,
			"api_symbol":     apiSymbol,
			"provider":       provider,
			"status":         "SUCCESS",
			"item_count":     len(quotes),
		})
	}
	if len(items) == 0 && len(failures) > 0 {
		return nil, "", errors.New(strings.Join(failures, "; "))
	}
	return items, marshalJSONSilently(map[string]interface{}{
		"source_key": sourceKey,
		"provider":   "MYSELF",
		"asset":      marketAssetClassStock,
		"items":      summary,
		"warnings":   failures,
	}), nil
}

func fetchFuturesMarketBarsFromMyself(config map[string]interface{}, sourceKey string, instrumentKeys []string, externalSymbols map[string]string, days int) ([]model.MarketDailyBar, string, error) {
	timeoutMS := parseDataSourceTimeoutMS(config)
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	items := make([]model.MarketDailyBar, 0, len(instrumentKeys)*maxInt(days, 1))
	summary := make([]map[string]interface{}, 0, len(instrumentKeys))
	failures := make([]string, 0)

	for _, instrumentKey := range instrumentKeys {
		apiSymbol := myselfFuturesAPISymbol(instrumentKey, externalSymbols[instrumentKey])
		if apiSymbol == "" {
			failures = append(failures, instrumentKey+": invalid futures symbol")
			continue
		}
		bars, err := fetchFuturesMarketBarsFromMyselfSina(config, sourceKey, strings.ToUpper(strings.TrimSpace(instrumentKey)), apiSymbol, days, timeoutMS)
		if err != nil {
			failures = append(failures, instrumentKey+": "+err.Error())
			summary = append(summary, map[string]interface{}{
				"instrument_key": instrumentKey,
				"api_symbol":     apiSymbol,
				"provider":       "MYSELF",
				"status":         "FAILED",
				"message":        err.Error(),
			})
			continue
		}
		items = append(items, bars...)
		summary = append(summary, map[string]interface{}{
			"instrument_key": instrumentKey,
			"api_symbol":     apiSymbol,
			"provider":       "SINA_FUTURES_DAILY",
			"status":         "SUCCESS",
			"item_count":     len(bars),
		})
	}
	if len(items) == 0 && len(failures) > 0 {
		return nil, "", errors.New(strings.Join(failures, "; "))
	}
	return items, marshalJSONSilently(map[string]interface{}{
		"source_key": sourceKey,
		"provider":   "MYSELF",
		"asset":      marketAssetClassFutures,
		"items":      summary,
		"warnings":   failures,
	}), nil
}

func fetchStockQuotesFromMyselfTencent(config map[string]interface{}, sourceKey string, instrumentKey string, apiSymbol string, days int, timeoutMS int) ([]model.StockMarketQuote, string, error) {
	endpoint := parseDataSourceStringConfig(config, "stock_kline_endpoint_tencent")
	if strings.TrimSpace(endpoint) == "" {
		endpoint = "https://web.ifzq.gtimg.cn/appstock/app/fqkline/get"
	}
	lookbackDays := maxInt(days*3, 30)
	query := url.Values{}
	query.Set("param", fmt.Sprintf("%s,day,%s,%s,%d,", apiSymbol, time.Now().AddDate(0, 0, -lookbackDays).Format("2006-01-02"), time.Now().Format("2006-01-02"), maxInt(days, 1)))
	body, err := myselfHTTPGet(endpoint, query, nil, timeoutMS)
	if err != nil {
		return nil, "", err
	}
	quotes, err := parseMyselfTencentStockDailyPayload(instrumentKey, sourceKey, body, days)
	if err != nil {
		return nil, "", err
	}
	return quotes, "TENCENT_STOCK_KLINE", nil
}

func fetchStockQuotesFromMyselfSina(config map[string]interface{}, sourceKey string, instrumentKey string, apiSymbol string, days int, timeoutMS int) ([]model.StockMarketQuote, error) {
	endpoint := parseDataSourceStringConfig(config, "stock_kline_endpoint_sina")
	if strings.TrimSpace(endpoint) == "" {
		endpoint = "https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData"
	}
	query := url.Values{}
	query.Set("symbol", apiSymbol)
	query.Set("scale", "240")
	query.Set("ma", "no")
	query.Set("datalen", strconv.Itoa(maxInt(days, 1)))
	headers := map[string]string{
		"Referer": firstNonEmpty(parseDataSourceStringConfig(config, "referer"), "https://finance.sina.com.cn"),
	}
	body, err := myselfHTTPGet(endpoint, query, headers, timeoutMS)
	if err != nil {
		return nil, err
	}
	return parseMyselfSinaStockDailyPayload(instrumentKey, sourceKey, body, days)
}

func fetchFuturesMarketBarsFromMyselfSina(config map[string]interface{}, sourceKey string, instrumentKey string, apiSymbol string, days int, timeoutMS int) ([]model.MarketDailyBar, error) {
	endpoint := parseDataSourceStringConfig(config, "futures_kline_endpoint_sina")
	if strings.TrimSpace(endpoint) == "" {
		endpoint = "https://stock2.finance.sina.com.cn/futures/api/jsonp.php/var%20_TEST=/InnerFuturesNewService.getDailyKLine"
	}
	query := url.Values{}
	query.Set("symbol", apiSymbol)
	query.Set("type", "2021_04_12")
	body, err := myselfHTTPGet(endpoint, query, nil, timeoutMS)
	if err != nil {
		return nil, err
	}
	return parseMyselfSinaFuturesDailyPayload(instrumentKey, sourceKey, apiSymbol, body, days)
}

func myselfHTTPGet(endpoint string, query url.Values, headers map[string]string, timeoutMS int) ([]byte, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return nil, err
	}
	values := parsed.Query()
	for key, items := range query {
		values.Del(key)
		for _, item := range items {
			values.Add(key, item)
		}
	}
	parsed.RawQuery = values.Encode()
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	req, err := http.NewRequest(http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		if strings.TrimSpace(value) == "" {
			continue
		}
		req.Header.Set(key, value)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("myself endpoint status %s: %s", resp.Status, truncateMarketPayload(body, 160))
	}
	return body, nil
}

func myselfStockAPISymbol(instrumentKey string, externalSymbol string) string {
	for _, candidate := range []string{externalSymbol, instrumentKey} {
		normalized := strings.ToUpper(strings.TrimSpace(candidate))
		if normalized == "" {
			continue
		}
		if !strings.Contains(normalized, ".") && (strings.HasPrefix(normalized, "SH") || strings.HasPrefix(normalized, "SZ") || strings.HasPrefix(normalized, "BJ")) {
			return strings.ToLower(normalized)
		}
		code := normalized
		exchange := ""
		if idx := strings.LastIndex(normalized, "."); idx >= 0 {
			code = normalized[:idx]
			exchange = normalized[idx+1:]
		}
		prefix := myselfStockExchangePrefix(exchange, code)
		if prefix == "" {
			continue
		}
		return prefix + strings.ToLower(code)
	}
	return ""
}

func myselfStockExchangePrefix(exchange string, code string) string {
	switch strings.ToUpper(strings.TrimSpace(exchange)) {
	case "SH", "SSE":
		return "sh"
	case "SZ", "SZSE":
		return "sz"
	case "BJ", "BSE":
		return "bj"
	}
	trimmedCode := strings.TrimSpace(code)
	if len(trimmedCode) == 0 {
		return ""
	}
	switch trimmedCode[0] {
	case '5', '6', '9':
		return "sh"
	case '4', '8':
		return "bj"
	default:
		return "sz"
	}
}

func myselfFuturesAPISymbol(instrumentKey string, externalSymbol string) string {
	for _, candidate := range []string{externalSymbol, instrumentKey} {
		normalized := strings.ToUpper(strings.TrimSpace(candidate))
		if normalized == "" {
			continue
		}
		if idx := strings.LastIndex(normalized, "."); idx >= 0 {
			normalized = normalized[:idx]
		}
		return normalized
	}
	return ""
}

func parseMyselfTencentStockDailyPayload(instrumentKey string, sourceKey string, body []byte, days int) ([]model.StockMarketQuote, error) {
	payload := struct {
		Code int                                   `json:"code"`
		Msg  string                                `json:"msg"`
		Data map[string]map[string]json.RawMessage `json:"data"`
	}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if payload.Code != 0 {
		return nil, fmt.Errorf("tencent kline error(code=%d): %s", payload.Code, strings.TrimSpace(payload.Msg))
	}
	if len(payload.Data) == 0 {
		return nil, errors.New("tencent kline returned empty data")
	}
	for _, item := range payload.Data {
		var rows [][]string
		for _, key := range []string{"day", "qfqday"} {
			rawRows, ok := item[key]
			if !ok || len(rawRows) == 0 {
				continue
			}
			if err := json.Unmarshal(rawRows, &rows); err == nil && len(rows) > 0 {
				break
			}
		}
		if len(rows) == 0 {
			continue
		}
		return buildMyselfStockQuotesFromRows(instrumentKey, sourceKey, rows, days, func(row []string) (string, float64, float64, float64, float64, float64, bool) {
			if len(row) < 6 {
				return "", 0, 0, 0, 0, 0, false
			}
			openPrice, okOpen := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
			closePrice, okClose := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
			highPrice, okHigh := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
			lowPrice, okLow := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
			volume, okVolume := strconv.ParseFloat(strings.TrimSpace(row[5]), 64)
			if okOpen != nil || okClose != nil || okHigh != nil || okLow != nil || okVolume != nil {
				return "", 0, 0, 0, 0, 0, false
			}
			return strings.TrimSpace(row[0]), openPrice, highPrice, lowPrice, closePrice, volume, true
		})
	}
	return nil, errors.New("tencent kline returned no usable rows")
}

func parseMyselfSinaStockDailyPayload(instrumentKey string, sourceKey string, body []byte, days int) ([]model.StockMarketQuote, error) {
	var rows []map[string]string
	if err := json.Unmarshal(body, &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("sina stock kline returned no rows")
	}
	stringRows := make([][]string, 0, len(rows))
	for _, row := range rows {
		stringRows = append(stringRows, []string{
			strings.TrimSpace(row["day"]),
			strings.TrimSpace(row["open"]),
			strings.TrimSpace(row["high"]),
			strings.TrimSpace(row["low"]),
			strings.TrimSpace(row["close"]),
			strings.TrimSpace(row["volume"]),
		})
	}
	return buildMyselfStockQuotesFromRows(instrumentKey, sourceKey, stringRows, days, func(row []string) (string, float64, float64, float64, float64, float64, bool) {
		if len(row) < 6 {
			return "", 0, 0, 0, 0, 0, false
		}
		openPrice, okOpen := strconv.ParseFloat(strings.TrimSpace(row[1]), 64)
		highPrice, okHigh := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
		lowPrice, okLow := strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
		closePrice, okClose := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
		volume, okVolume := strconv.ParseFloat(strings.TrimSpace(row[5]), 64)
		if okOpen != nil || okHigh != nil || okLow != nil || okClose != nil || okVolume != nil {
			return "", 0, 0, 0, 0, 0, false
		}
		return strings.TrimSpace(row[0]), openPrice, highPrice, lowPrice, closePrice, volume, true
	})
}

func buildMyselfStockQuotesFromRows(instrumentKey string, sourceKey string, rows [][]string, days int, parser func([]string) (string, float64, float64, float64, float64, float64, bool)) ([]model.StockMarketQuote, error) {
	items := make([]model.StockMarketQuote, 0, len(rows))
	normalizedKey := strings.ToUpper(strings.TrimSpace(instrumentKey))
	for _, row := range rows {
		tradeDate, openPrice, highPrice, lowPrice, closePrice, volume, ok := parser(row)
		if !ok || closePrice <= 0 {
			continue
		}
		items = append(items, model.StockMarketQuote{
			Symbol:         normalizedKey,
			TradeDate:      tradeDate,
			OpenPrice:      roundTo(openPrice, 4),
			HighPrice:      roundTo(highPrice, 4),
			LowPrice:       roundTo(lowPrice, 4),
			ClosePrice:     roundTo(closePrice, 4),
			PrevClosePrice: 0,
			Volume:         int64(math.Round(volume)),
			Turnover:       0,
			SourceKey:      sourceKey,
		})
	}
	if len(items) == 0 {
		return nil, errors.New("myself stock kline returned no usable rows")
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].TradeDate < items[j].TradeDate
	})
	if days > 0 && len(items) > days {
		items = append([]model.StockMarketQuote(nil), items[len(items)-days:]...)
	}
	var prevClose float64
	for idx := range items {
		if idx == 0 {
			if items[idx].OpenPrice > 0 {
				prevClose = items[idx].OpenPrice
			} else {
				prevClose = items[idx].ClosePrice
			}
		}
		items[idx].PrevClosePrice = roundTo(prevClose, 4)
		prevClose = items[idx].ClosePrice
	}
	return items, nil
}

func parseMyselfSinaFuturesDailyPayload(instrumentKey string, sourceKey string, apiSymbol string, body []byte, days int) ([]model.MarketDailyBar, error) {
	payloadText := strings.TrimSpace(string(body))
	start := strings.Index(payloadText, "=(")
	if start < 0 {
		return nil, fmt.Errorf("unsupported sina futures payload: %s", truncateMarketPayload(body, 160))
	}
	payloadText = strings.TrimSpace(payloadText[start+2:])
	payloadText = strings.TrimSuffix(payloadText, ");")
	var rows []map[string]string
	if err := json.Unmarshal([]byte(payloadText), &rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, errors.New("sina futures kline returned no rows")
	}
	items := make([]model.MarketDailyBar, 0, len(rows))
	normalizedKey := strings.ToUpper(strings.TrimSpace(instrumentKey))
	normalizedSymbol := strings.ToUpper(strings.TrimSpace(apiSymbol))
	for _, row := range rows {
		tradeDate := strings.TrimSpace(row["d"])
		if tradeDate == "" {
			continue
		}
		openPrice, errOpen := strconv.ParseFloat(strings.TrimSpace(row["o"]), 64)
		highPrice, errHigh := strconv.ParseFloat(strings.TrimSpace(row["h"]), 64)
		lowPrice, errLow := strconv.ParseFloat(strings.TrimSpace(row["l"]), 64)
		closePrice, errClose := strconv.ParseFloat(strings.TrimSpace(row["c"]), 64)
		volume, errVolume := strconv.ParseFloat(strings.TrimSpace(row["v"]), 64)
		openInterest, errHold := strconv.ParseFloat(strings.TrimSpace(row["p"]), 64)
		settlePrice, _ := strconv.ParseFloat(strings.TrimSpace(row["s"]), 64)
		if errOpen != nil || errHigh != nil || errLow != nil || errClose != nil || errVolume != nil || errHold != nil || closePrice <= 0 {
			continue
		}
		items = append(items, model.MarketDailyBar{
			AssetClass:      marketAssetClassFutures,
			InstrumentKey:   normalizedKey,
			ExternalSymbol:  normalizedSymbol,
			TradeDate:       tradeDate,
			OpenPrice:       roundTo(openPrice, 4),
			HighPrice:       roundTo(highPrice, 4),
			LowPrice:        roundTo(lowPrice, 4),
			ClosePrice:      roundTo(closePrice, 4),
			PrevClosePrice:  0,
			SettlePrice:     roundTo(settlePrice, 4),
			PrevSettlePrice: 0,
			Volume:          int64(math.Round(volume)),
			Turnover:        0,
			OpenInterest:    roundTo(openInterest, 4),
			SourceKey:       sourceKey,
		})
	}
	if len(items) == 0 {
		return nil, errors.New("sina futures kline returned no usable rows")
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].TradeDate < items[j].TradeDate
	})
	if days > 0 && len(items) > days {
		items = append([]model.MarketDailyBar(nil), items[len(items)-days:]...)
	}
	var prevClose float64
	var prevSettle float64
	for idx := range items {
		if idx == 0 {
			if items[idx].OpenPrice > 0 {
				prevClose = items[idx].OpenPrice
			} else {
				prevClose = items[idx].ClosePrice
			}
			if items[idx].SettlePrice > 0 {
				prevSettle = items[idx].SettlePrice
			} else {
				prevSettle = prevClose
			}
		}
		items[idx].PrevClosePrice = roundTo(prevClose, 4)
		items[idx].PrevSettlePrice = roundTo(prevSettle, 4)
		if items[idx].SettlePrice <= 0 {
			items[idx].SettlePrice = items[idx].ClosePrice
		}
		prevClose = items[idx].ClosePrice
		prevSettle = items[idx].SettlePrice
	}
	return items, nil
}

func buildMockMarketDailyBars(assetClass string, sourceKey string, instrumentKeys []string, days int) []model.MarketDailyBar {
	switch assetClass {
	case marketAssetClassStock:
		return convertStockQuotesToMarketBars(buildMockStockQuotes(instrumentKeys, days, sourceKey), instrumentKeys, nil)
	default:
		normalizedAssetClass := strings.ToUpper(strings.TrimSpace(assetClass))
		if normalizedAssetClass == "" {
			normalizedAssetClass = marketAssetClassFutures
		}
		now := time.Now()
		tradeDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		items := make([]model.MarketDailyBar, 0, len(instrumentKeys)*days)
		for _, instrumentKey := range instrumentKeys {
			seed := symbolSeed(instrumentKey)
			prevClose := 3000 + seed*25
			for offset := days - 1; offset >= 0; offset-- {
				currentDay := tradeDay.AddDate(0, 0, -offset)
				closePrice := prevClose * (1 + 0.002*math.Sin(float64(offset)/5.0+seed))
				openPrice := prevClose * (1 + 0.001*math.Cos(float64(offset)/7.0+seed))
				highPrice := math.Max(openPrice, closePrice) * 1.004
				lowPrice := math.Min(openPrice, closePrice) * 0.996
				volume := int64(20000 + math.Round(seed*1300))
				turnover := closePrice * float64(volume)
				items = append(items, model.MarketDailyBar{
					AssetClass:     normalizedAssetClass,
					InstrumentKey:  instrumentKey,
					ExternalSymbol: deriveDefaultExternalSymbol("MOCK", normalizedAssetClass, instrumentKey),
					TradeDate:      currentDay.Format("2006-01-02"),
					OpenPrice:      roundTo(openPrice, 4),
					HighPrice:      roundTo(highPrice, 4),
					LowPrice:       roundTo(lowPrice, 4),
					ClosePrice:     roundTo(closePrice, 4),
					PrevClosePrice: roundTo(prevClose, 4),
					SettlePrice:    roundTo(closePrice, 4),
					Volume:         volume,
					Turnover:       roundTo(turnover, 2),
					OpenInterest:   roundTo(float64(volume)*1.8, 2),
					SourceKey:      sourceKey,
				})
				prevClose = closePrice
			}
		}
		return items
	}
}

func convertStockQuotesToMarketBars(quotes []model.StockMarketQuote, instrumentKeys []string, externalSymbols map[string]string) []model.MarketDailyBar {
	reverse := make(map[string]string, len(instrumentKeys))
	for _, instrumentKey := range instrumentKeys {
		normalized := strings.ToUpper(strings.TrimSpace(instrumentKey))
		if normalized == "" {
			continue
		}
		external := normalized
		if externalSymbols != nil {
			if value := strings.TrimSpace(externalSymbols[normalized]); value != "" {
				external = strings.ToUpper(value)
			}
		}
		reverse[external] = normalized
		reverse[normalized] = normalized
	}
	items := make([]model.MarketDailyBar, 0, len(quotes))
	for _, quote := range quotes {
		externalSymbol := strings.ToUpper(strings.TrimSpace(quote.Symbol))
		instrumentKey := reverse[externalSymbol]
		if instrumentKey == "" {
			instrumentKey = externalSymbol
		}
		items = append(items, model.MarketDailyBar{
			AssetClass:     marketAssetClassStock,
			InstrumentKey:  instrumentKey,
			ExternalSymbol: externalSymbol,
			TradeDate:      quote.TradeDate,
			OpenPrice:      quote.OpenPrice,
			HighPrice:      quote.HighPrice,
			LowPrice:       quote.LowPrice,
			ClosePrice:     quote.ClosePrice,
			PrevClosePrice: quote.PrevClosePrice,
			Volume:         quote.Volume,
			Turnover:       quote.Turnover,
			SourceKey:      quote.SourceKey,
		})
	}
	return items
}

func fetchStockMarketBarsFromTushare(token string, sourceKey string, instrumentKeys []string, externalSymbols map[string]string, days int, timeoutMS int) ([]model.MarketDailyBar, string, error) {
	items := make([]string, 0, len(instrumentKeys))
	for _, instrumentKey := range instrumentKeys {
		external := strings.TrimSpace(externalSymbols[instrumentKey])
		if external == "" {
			external = instrumentKey
		}
		items = append(items, external)
	}
	quotes, err := fetchStockQuotesFromTushare(token, sourceKey, items, days, timeoutMS)
	if err != nil {
		return nil, "", err
	}
	return convertStockQuotesToMarketBars(quotes, instrumentKeys, externalSymbols), "", nil
}

func fetchStockMarketBarsFromTushareDateRange(token string, sourceKey string, instrumentKeys []string, externalSymbols map[string]string, tradeDateFrom string, tradeDateTo string, timeoutMS int) ([]model.MarketDailyBar, string, error) {
	items := make([]string, 0, len(instrumentKeys))
	for _, instrumentKey := range instrumentKeys {
		external := strings.TrimSpace(externalSymbols[instrumentKey])
		if external == "" {
			external = instrumentKey
		}
		items = append(items, external)
	}
	startDate := strings.ReplaceAll(strings.TrimSpace(tradeDateFrom), "-", "")
	endDate := strings.ReplaceAll(strings.TrimSpace(tradeDateTo), "-", "")
	quotes, err := fetchStockQuotesFromTushareDateRange(token, sourceKey, items, startDate, endDate, timeoutMS)
	if err != nil {
		return nil, "", err
	}
	return convertStockQuotesToMarketBars(quotes, instrumentKeys, externalSymbols), marshalJSONSilently(map[string]interface{}{
		"source_key":      sourceKey,
		"provider":        "TUSHARE",
		"asset":           marketAssetClassStock,
		"trade_date_from": strings.TrimSpace(tradeDateFrom),
		"trade_date_to":   strings.TrimSpace(tradeDateTo),
		"symbols":         items,
	}), nil
}

func fetchFuturesMarketBarsFromTushare(token string, sourceKey string, instrumentKeys []string, externalSymbols map[string]string, days int, timeoutMS int) ([]model.MarketDailyBar, string, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, "", errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 120
	}
	startDate := time.Now().AddDate(0, 0, -(days + 20)).Format("20060102")
	endDate := time.Now().Format("20060102")
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	items := make([]model.MarketDailyBar, 0, len(instrumentKeys)*days)
	rawSnapshots := make([]map[string]interface{}, 0, len(instrumentKeys))
	reverse := make(map[string]string, len(instrumentKeys))
	for _, instrumentKey := range instrumentKeys {
		externalSymbol := strings.ToUpper(strings.TrimSpace(externalSymbols[instrumentKey]))
		if externalSymbol == "" {
			externalSymbol = strings.ToUpper(strings.TrimSpace(instrumentKey))
		}
		reverse[externalSymbol] = strings.ToUpper(strings.TrimSpace(instrumentKey))
		parsed, err := callTushareAPI(client, token, "fut_daily", map[string]string{
			"ts_code":    externalSymbol,
			"start_date": startDate,
			"end_date":   endDate,
		}, "ts_code,trade_date,pre_close,pre_settle,open,high,low,close,settle,vol,amount,oi")
		if err != nil {
			return nil, "", err
		}
		rawSnapshots = append(rawSnapshots, map[string]interface{}{
			"external_symbol": externalSymbol,
			"field_count":     len(parsed.Data.Fields),
			"item_count":      len(parsed.Data.Items),
		})
		fieldIndex := make(map[string]int, len(parsed.Data.Fields))
		for idx, field := range parsed.Data.Fields {
			fieldIndex[strings.TrimSpace(field)] = idx
		}
		for _, row := range parsed.Data.Items {
			tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
			if !ok {
				continue
			}
			tradeDateRaw, ok := tushareGetString(row, fieldIndex, "trade_date")
			if !ok {
				continue
			}
			tradeDate, err := time.ParseInLocation("20060102", tradeDateRaw, time.Local)
			if err != nil {
				continue
			}
			openPrice, _ := tushareGetFloat(row, fieldIndex, "open")
			highPrice, _ := tushareGetFloat(row, fieldIndex, "high")
			lowPrice, _ := tushareGetFloat(row, fieldIndex, "low")
			closePrice, ok := tushareGetFloat(row, fieldIndex, "close")
			if !ok || closePrice <= 0 {
				continue
			}
			prevClose, _ := tushareGetFloat(row, fieldIndex, "pre_close")
			prevSettle, _ := tushareGetFloat(row, fieldIndex, "pre_settle")
			settlePrice, _ := tushareGetFloat(row, fieldIndex, "settle")
			volume, _ := tushareGetFloat(row, fieldIndex, "vol")
			turnover, _ := tushareGetFloat(row, fieldIndex, "amount")
			openInterest, _ := tushareGetFloat(row, fieldIndex, "oi")
			upperCode := strings.ToUpper(strings.TrimSpace(tsCode))
			instrumentKey := reverse[upperCode]
			if instrumentKey == "" {
				instrumentKey = upperCode
			}
			items = append(items, model.MarketDailyBar{
				AssetClass:      marketAssetClassFutures,
				InstrumentKey:   instrumentKey,
				ExternalSymbol:  upperCode,
				TradeDate:       tradeDate.Format("2006-01-02"),
				OpenPrice:       roundTo(openPrice, 4),
				HighPrice:       roundTo(highPrice, 4),
				LowPrice:        roundTo(lowPrice, 4),
				ClosePrice:      roundTo(closePrice, 4),
				PrevClosePrice:  roundTo(prevClose, 4),
				SettlePrice:     roundTo(settlePrice, 4),
				PrevSettlePrice: roundTo(prevSettle, 4),
				Volume:          int64(math.Round(volume)),
				Turnover:        roundTo(turnover, 4),
				OpenInterest:    roundTo(openInterest, 4),
				SourceKey:       sourceKey,
			})
		}
	}
	return items, marshalJSONSilently(map[string]interface{}{
		"source_key": sourceKey,
		"provider":   "TUSHARE",
		"asset":      marketAssetClassFutures,
		"items":      rawSnapshots,
	}), nil
}

func fetchStockMarketBarsFromAkshareBridge(config map[string]interface{}, sourceKey string, instrumentKeys []string, externalSymbols map[string]string, days int) ([]model.MarketDailyBar, string, error) {
	pythonBin, scriptPath, timeoutMS, err := resolvePythonBridgeRuntime(config)
	if err != nil {
		return nil, "", err
	}
	args := []string{scriptPath, "stock_daily", "--days", strconv.Itoa(days)}
	if len(instrumentKeys) > 0 {
		values := make([]string, 0, len(instrumentKeys))
		for _, instrumentKey := range instrumentKeys {
			externalSymbol := strings.TrimSpace(externalSymbols[instrumentKey])
			if externalSymbol == "" {
				externalSymbol = deriveDefaultExternalSymbol(sourceKey, marketAssetClassStock, instrumentKey)
			}
			values = append(values, externalSymbol)
		}
		args = append(args, "--symbols", strings.Join(values, ","))
	}
	output, err := runPythonBridgeCommand(pythonBin, args, timeoutMS)
	if err != nil {
		return nil, "", err
	}
	return decodeAkshareBridgeDailyBars(output, marketAssetClassStock, sourceKey, instrumentKeys, externalSymbols)
}

func fetchFuturesMarketBarsFromAkshareBridge(config map[string]interface{}, sourceKey string, instrumentKeys []string, externalSymbols map[string]string, days int) ([]model.MarketDailyBar, string, error) {
	pythonBin, scriptPath, timeoutMS, err := resolvePythonBridgeRuntime(config)
	if err != nil {
		return nil, "", err
	}
	args := []string{scriptPath, "futures_daily", "--days", strconv.Itoa(days)}
	if len(instrumentKeys) > 0 {
		values := make([]string, 0, len(instrumentKeys))
		for _, instrumentKey := range instrumentKeys {
			externalSymbol := strings.TrimSpace(externalSymbols[instrumentKey])
			if externalSymbol == "" {
				externalSymbol = deriveDefaultExternalSymbol(sourceKey, marketAssetClassFutures, instrumentKey)
			}
			values = append(values, externalSymbol)
		}
		args = append(args, "--symbols", strings.Join(values, ","))
	}
	output, err := runPythonBridgeCommand(pythonBin, args, timeoutMS)
	if err != nil {
		return nil, "", err
	}
	return decodeAkshareBridgeDailyBars(output, marketAssetClassFutures, sourceKey, instrumentKeys, externalSymbols)
}

func decodeAkshareBridgeDailyBars(output []byte, assetClass string, sourceKey string, instrumentKeys []string, externalSymbols map[string]string) ([]model.MarketDailyBar, string, error) {
	payload := pythonBridgeDailyBarPayload{}
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, "", err
	}
	return convertAkshareBridgeDailyBars(assetClass, sourceKey, instrumentKeys, externalSymbols, payload.Items), string(output), nil
}

func convertAkshareBridgeDailyBars(assetClass string, sourceKey string, instrumentKeys []string, externalSymbols map[string]string, payloadItems []pythonBridgeDailyBarItem) []model.MarketDailyBar {
	reverse := make(map[string]string, len(instrumentKeys))
	for _, instrumentKey := range instrumentKeys {
		externalSymbol := strings.TrimSpace(externalSymbols[instrumentKey])
		if externalSymbol == "" {
			externalSymbol = deriveDefaultExternalSymbol(sourceKey, assetClass, instrumentKey)
		}
		reverse[strings.ToUpper(externalSymbol)] = strings.ToUpper(strings.TrimSpace(instrumentKey))
	}
	items := make([]model.MarketDailyBar, 0, len(payloadItems))
	for _, item := range payloadItems {
		tradeDate, err := parseFlexibleDateTime(item.TradeDate)
		if err != nil {
			continue
		}
		externalSymbol := strings.ToUpper(strings.TrimSpace(item.ExternalSymbol))
		instrumentKey := strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		if instrumentKey == "" {
			instrumentKey = reverse[externalSymbol]
		}
		if instrumentKey == "" {
			instrumentKey = externalSymbol
		}
		bar := model.MarketDailyBar{
			AssetClass:      assetClass,
			InstrumentKey:   instrumentKey,
			ExternalSymbol:  externalSymbol,
			TradeDate:       tradeDate.Format("2006-01-02"),
			OpenPrice:       roundTo(item.OpenPrice, 4),
			HighPrice:       roundTo(item.HighPrice, 4),
			LowPrice:        roundTo(item.LowPrice, 4),
			ClosePrice:      roundTo(item.ClosePrice, 4),
			PrevClosePrice:  roundTo(item.PrevClosePrice, 4),
			SettlePrice:     roundTo(item.SettlePrice, 4),
			PrevSettlePrice: roundTo(item.PrevSettlePrice, 4),
			Volume:          int64(math.Round(item.Volume)),
			Turnover:        roundTo(item.Turnover, 4),
			OpenInterest:    roundTo(item.OpenInterest, 4),
			SourceKey:       sourceKey,
		}
		if assetClass == marketAssetClassFutures {
			if bar.SettlePrice <= 0 {
				bar.SettlePrice = bar.ClosePrice
			}
			if bar.PrevSettlePrice <= 0 {
				bar.PrevSettlePrice = bar.PrevClosePrice
			}
		}
		items = append(items, bar)
	}
	return items
}

func fetchMarketBarsFromTickerMD(config map[string]interface{}, sourceKey string, assetClass string, instrumentKeys []string, externalSymbols map[string]string, days int) ([]model.MarketDailyBar, string, error) {
	endpoint := parseDataSourceStringConfig(config, "kline_endpoint", "endpoint")
	if strings.TrimSpace(endpoint) == "" {
		endpoint = "http://39.107.99.235:1008/redis.php"
	}
	timeoutMS := parseDataSourceTimeoutMS(config)
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	items := make([]model.MarketDailyBar, 0, len(instrumentKeys)*days)
	payloadSummary := make([]map[string]interface{}, 0, len(instrumentKeys))

	for _, instrumentKey := range instrumentKeys {
		externalSymbol := strings.TrimSpace(externalSymbols[instrumentKey])
		if externalSymbol == "" {
			externalSymbol = deriveDefaultExternalSymbol(sourceKey, assetClass, instrumentKey)
		}
		parsedURL, err := url.Parse(endpoint)
		if err != nil {
			return nil, "", err
		}
		query := parsedURL.Query()
		query.Set("code", externalSymbol)
		query.Set("time", "1d")
		query.Set("rows", strconv.Itoa(days))
		parsedURL.RawQuery = query.Encode()
		resp, err := client.Get(parsedURL.String())
		if err != nil {
			return nil, "", err
		}
		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, "", readErr
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, "", fmt.Errorf("tickermd kline status: %s", resp.Status)
		}
		rows, err := decodeTickerMDDailyRows(body)
		if err != nil {
			return nil, "", err
		}
		bars := parseTickerMDDailyBars(assetClass, sourceKey, strings.ToUpper(strings.TrimSpace(instrumentKey)), externalSymbol, rows)
		items = append(items, bars...)
		payloadSummary = append(payloadSummary, map[string]interface{}{
			"instrument_key":  instrumentKey,
			"external_symbol": externalSymbol,
			"item_count":      len(rows),
		})
	}

	return items, marshalJSONSilently(map[string]interface{}{
		"source_key": sourceKey,
		"provider":   "TICKERMD",
		"asset":      assetClass,
		"items":      payloadSummary,
	}), nil
}

func decodeTickerMDDailyRows(body []byte) ([][]interface{}, error) {
	var directRows [][]interface{}
	if err := json.Unmarshal(body, &directRows); err == nil {
		return directRows, nil
	}

	var payload interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}
	if rows, ok := extractTickerMDDailyRows(payload); ok {
		return rows, nil
	}
	if message := extractTickerMDErrorMessage(payload); message != "" {
		return nil, errors.New(message)
	}
	return nil, fmt.Errorf("unsupported tickermd kline payload: %s", truncateMarketPayload(body, 160))
}

func extractTickerMDDailyRows(payload interface{}) ([][]interface{}, bool) {
	switch value := payload.(type) {
	case []interface{}:
		rows := make([][]interface{}, 0, len(value))
		for _, item := range value {
			switch typed := item.(type) {
			case []interface{}:
				rows = append(rows, typed)
			case map[string]interface{}:
				row := normalizeTickerMDDailyRowObject(typed)
				if len(row) == 0 {
					return nil, false
				}
				rows = append(rows, row)
			default:
				if len(value) == 0 {
					return rows, true
				}
				return nil, false
			}
		}
		return rows, true
	case map[string]interface{}:
		for _, key := range []string{"data", "rows", "items", "list", "result", "kline"} {
			if child, ok := value[key]; ok {
				if rows, ok := extractTickerMDDailyRows(child); ok {
					return rows, true
				}
			}
		}
		for key, child := range value {
			normalizedKey := strings.ToLower(strings.TrimSpace(key))
			switch normalizedKey {
			case "error", "msg", "message", "code", "status", "success":
				continue
			}
			if rows, ok := extractTickerMDDailyRows(child); ok {
				return rows, true
			}
		}
	}
	return nil, false
}

func normalizeTickerMDDailyRowObject(item map[string]interface{}) []interface{} {
	tradeDate := firstTickerMDMapValue(item, "trade_date", "date", "day", "time", "datetime")
	timestamp := firstTickerMDMapValue(item, "timestamp", "ts", "trade_ts")
	if timestamp == nil {
		timestamp = tradeDate
	}
	openValue := firstTickerMDMapValue(item, "open", "o")
	highValue := firstTickerMDMapValue(item, "high", "h")
	lowValue := firstTickerMDMapValue(item, "low", "l")
	closeValue := firstTickerMDMapValue(item, "close", "c", "price")
	if closeValue == nil {
		return nil
	}
	volumeValue := firstTickerMDMapValue(item, "volume", "vol", "amount")
	return []interface{}{timestamp, openValue, highValue, lowValue, closeValue, tradeDate, volumeValue}
}

func firstTickerMDMapValue(item map[string]interface{}, keys ...string) interface{} {
	for _, key := range keys {
		if value, ok := item[key]; ok {
			return value
		}
	}
	return nil
}

func extractTickerMDErrorMessage(payload interface{}) string {
	item, ok := payload.(map[string]interface{})
	if !ok {
		return ""
	}
	message := ""
	for _, key := range []string{"error", "msg", "message"} {
		if value, ok := item[key]; ok {
			text := strings.TrimSpace(fmt.Sprintf("%v", value))
			if text != "" && text != "<nil>" {
				message = text
				break
			}
		}
	}
	if message == "" {
		return ""
	}
	code := ""
	if value, ok := item["code"]; ok {
		code = strings.TrimSpace(fmt.Sprintf("%v", value))
	}
	if code != "" && code != "0" && code != "200" {
		return fmt.Sprintf("tickermd error (code=%s): %s", code, message)
	}
	return "tickermd error: " + message
}

func truncateMarketPayload(body []byte, limit int) string {
	text := strings.TrimSpace(string(body))
	if limit <= 0 || len(text) <= limit {
		return text
	}
	return text[:limit] + "..."
}

func parseTickerMDDailyBars(assetClass string, sourceKey string, instrumentKey string, externalSymbol string, rows [][]interface{}) []model.MarketDailyBar {
	type sortableBar struct {
		TradeDate time.Time
		Bar       model.MarketDailyBar
	}
	parsed := make([]sortableBar, 0, len(rows))
	for _, row := range rows {
		if len(row) < 5 {
			continue
		}
		tradeDate, ok := parseTickerMDTradeDate(row)
		if !ok {
			continue
		}
		openPrice, ok := flexibleFloat(row, 1)
		if !ok {
			continue
		}
		highPrice, _ := flexibleFloat(row, 2)
		lowPrice, _ := flexibleFloat(row, 3)
		closePrice, ok := flexibleFloat(row, 4)
		if !ok || closePrice <= 0 {
			continue
		}
		volume, _ := flexibleFloat(row, 6)
		parsed = append(parsed, sortableBar{
			TradeDate: tradeDate,
			Bar: model.MarketDailyBar{
				AssetClass:     assetClass,
				InstrumentKey:  instrumentKey,
				ExternalSymbol: strings.ToUpper(strings.TrimSpace(externalSymbol)),
				TradeDate:      tradeDate.Format("2006-01-02"),
				OpenPrice:      roundTo(openPrice, 4),
				HighPrice:      roundTo(highPrice, 4),
				LowPrice:       roundTo(lowPrice, 4),
				ClosePrice:     roundTo(closePrice, 4),
				Volume:         int64(math.Round(volume)),
				Turnover:       0,
				SourceKey:      sourceKey,
			},
		})
	}
	sort.Slice(parsed, func(i, j int) bool {
		return parsed[i].TradeDate.Before(parsed[j].TradeDate)
	})
	items := make([]model.MarketDailyBar, 0, len(parsed))
	var prevClose float64
	for idx, item := range parsed {
		bar := item.Bar
		if idx == 0 {
			prevClose = bar.OpenPrice
		}
		bar.PrevClosePrice = roundTo(prevClose, 4)
		if bar.Turnover <= 0 && bar.Volume > 0 {
			bar.Turnover = roundTo(bar.ClosePrice*float64(bar.Volume), 2)
		}
		items = append(items, bar)
		prevClose = bar.ClosePrice
	}
	return items
}

func parseTickerMDTradeDate(row []interface{}) (time.Time, bool) {
	if len(row) > 5 {
		if text := strings.TrimSpace(fmt.Sprintf("%v", row[5])); text != "" && text != "<nil>" {
			if parsed, err := parseFlexibleDateTime(text); err == nil {
				return parsed, true
			}
		}
	}
	if len(row) > 0 {
		switch value := row[0].(type) {
		case float64:
			ts := int64(value)
			if ts > 1_000_000_000_000 {
				return time.UnixMilli(ts), true
			}
			if ts > 0 {
				return time.Unix(ts, 0), true
			}
		case int64:
			if value > 1_000_000_000_000 {
				return time.UnixMilli(value), true
			}
			if value > 0 {
				return time.Unix(value, 0), true
			}
		}
	}
	return time.Time{}, false
}

func flexibleFloat(row []interface{}, index int) (float64, bool) {
	if index < 0 || index >= len(row) {
		return 0, false
	}
	return tushareGetFloat(row, map[string]int{"value": index}, "value")
}

func resolvePythonBridgeRuntime(config map[string]interface{}) (string, string, int, error) {
	pythonBin := parseDataSourceStringConfig(config, "python_bin")
	if pythonBin == "" {
		pythonBin = "../services/strategy-engine/.venv/bin/python"
	}
	scriptPath := parseDataSourceStringConfig(config, "bridge_script")
	if scriptPath == "" {
		scriptPath = "../services/strategy-engine/app/tools/market_bridge.py"
	}
	pythonBin = resolveExistingLocalPath(pythonBin)
	scriptPath = resolveExistingLocalPath(scriptPath)
	if _, err := os.Stat(scriptPath); err != nil {
		return "", "", 0, fmt.Errorf("market bridge script not found: %s", scriptPath)
	}
	timeoutMS := parseDataSourceTimeoutMS(config)
	if timeoutMS <= 0 {
		timeoutMS = 10000
	}
	return pythonBin, scriptPath, timeoutMS, nil
}

func resolveExistingLocalPath(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	candidates := []string{trimmed}
	if cwd, err := os.Getwd(); err == nil {
		candidates = append(candidates,
			filepath.Clean(filepath.Join(cwd, trimmed)),
			filepath.Clean(filepath.Join(cwd, "backend", trimmed)),
			filepath.Clean(filepath.Join(cwd, "..", trimmed)),
		)
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return trimmed
}

func runPythonBridgeCommand(pythonBin string, args []string, timeoutMS int) ([]byte, error) {
	if strings.TrimSpace(pythonBin) == "" {
		pythonBin = "python3"
	}
	if timeoutMS <= 0 {
		timeoutMS = 10000
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()
	cmd := exec.CommandContext(ctx, pythonBin, args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, errors.New("python market bridge timed out")
		}
		errorText := strings.TrimSpace(stderr.String())
		if errorText == "" {
			errorText = err.Error()
		}
		return nil, errors.New(errorText)
	}
	return stdout.Bytes(), nil
}

func (r *MySQLGrowthRepo) fetchMarketNewsForSource(item model.DataSource, symbols []string, days int, limit int) ([]model.MarketNewsItem, string, error) {
	sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
	provider := strings.ToUpper(parseDataSourceStringConfig(item.Config, "provider", "vendor"))
	if provider == "" {
		provider = sourceKey
	}
	switch provider {
	case "AKSHARE":
		return fetchMarketNewsFromAkshareBridge(item.Config, sourceKey, symbols, days, limit)
	case "TUSHARE":
		token := parseDataSourceStringConfig(item.Config, "token", "api_token", "tushare_token")
		if strings.TrimSpace(token) == "" {
			token = strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
		}
		if len(symbols) == 0 {
			symbols = defaultMockStockSymbols()
		}
		items, err := fetchStockNewsFromTushare(token, sourceKey, symbols, minInt(days, 30), parseDataSourceTimeoutMS(item.Config))
		if err != nil {
			return nil, "", err
		}
		return convertStockNewsRawToMarketNews(items), "", nil
	default:
		return nil, "", fmt.Errorf("unsupported news provider: %s", provider)
	}
}

func fetchMarketNewsFromAkshareBridge(config map[string]interface{}, sourceKey string, symbols []string, days int, limit int) ([]model.MarketNewsItem, string, error) {
	pythonBin, scriptPath, timeoutMS, err := resolvePythonBridgeRuntime(config)
	if err != nil {
		return nil, "", err
	}
	args := []string{scriptPath, "market_news", "--days", strconv.Itoa(days), "--limit", strconv.Itoa(limit)}
	if len(symbols) > 0 {
		values := make([]string, 0, len(symbols))
		for _, symbol := range symbols {
			values = append(values, deriveDefaultExternalSymbol("AKSHARE", marketAssetClassStock, symbol))
		}
		args = append(args, "--symbols", strings.Join(values, ","))
	}
	output, err := runPythonBridgeCommand(pythonBin, args, timeoutMS)
	if err != nil {
		return nil, "", err
	}
	payload := pythonBridgeNewsPayload{}
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, "", err
	}
	items := make([]model.MarketNewsItem, 0, len(payload.Items))
	for _, item := range payload.Items {
		publishedAt, err := parseFlexibleDateTime(item.PublishedAt)
		if err != nil {
			continue
		}
		symbolsCopy := normalizeStockSymbolList(item.Symbols)
		items = append(items, model.MarketNewsItem{
			SourceKey:     sourceKey,
			ExternalID:    strings.TrimSpace(item.ExternalID),
			NewsType:      strings.ToUpper(strings.TrimSpace(item.NewsType)),
			Title:         strings.TrimSpace(item.Title),
			Summary:       strings.TrimSpace(item.Summary),
			Content:       strings.TrimSpace(item.Content),
			URL:           strings.TrimSpace(item.URL),
			PrimarySymbol: strings.ToUpper(strings.TrimSpace(item.PrimarySymbol)),
			Symbols:       symbolsCopy,
			PublishedAt:   publishedAt.Format(time.RFC3339),
		})
	}
	return items, string(output), nil
}

func convertStockNewsRawToMarketNews(items []stockNewsRawPoint) []model.MarketNewsItem {
	result := make([]model.MarketNewsItem, 0, len(items))
	for _, item := range items {
		result = append(result, model.MarketNewsItem{
			SourceKey:     strings.ToUpper(strings.TrimSpace(item.SourceKey)),
			ExternalID:    buildMarketNewsExternalID(item.SourceKey, item.Symbol, item.Title, item.PublishedAt),
			NewsType:      "ANNOUNCEMENT",
			Title:         strings.TrimSpace(item.Title),
			Summary:       strings.TrimSpace(item.Content),
			Content:       strings.TrimSpace(item.Content),
			URL:           strings.TrimSpace(item.URL),
			PrimarySymbol: strings.ToUpper(strings.TrimSpace(item.Symbol)),
			Symbols:       []string{strings.ToUpper(strings.TrimSpace(item.Symbol))},
			PublishedAt:   item.PublishedAt.Format(time.RFC3339),
		})
	}
	return result
}

func buildMarketNewsExternalID(sourceKey string, symbol string, title string, publishedAt time.Time) string {
	base := strings.ToUpper(strings.TrimSpace(sourceKey)) + "|" + strings.ToUpper(strings.TrimSpace(symbol)) + "|" + strings.TrimSpace(title) + "|" + publishedAt.Format(time.RFC3339)
	sum := sha1String(base)
	return sum[:24]
}

func buildFuturesInventoryExternalID(sourceKey string, symbol string, tradeDate string, warehouse string, warehouseID string, area string, brand string, place string, grade string) string {
	base := strings.ToUpper(strings.TrimSpace(sourceKey)) +
		"|" + strings.ToUpper(strings.TrimSpace(symbol)) +
		"|" + strings.TrimSpace(tradeDate) +
		"|" + strings.TrimSpace(warehouse) +
		"|" + strings.TrimSpace(warehouseID) +
		"|" + strings.TrimSpace(area) +
		"|" + strings.TrimSpace(brand) +
		"|" + strings.TrimSpace(place) +
		"|" + strings.TrimSpace(grade)
	sum := sha1String(base)
	return sum[:24]
}

func sha1String(text string) string {
	hash := sha1.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (r *MySQLGrowthRepo) upsertFuturesInventorySnapshots(items []model.FuturesInventorySnapshot) (int, error) {
	now := time.Now()
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		tradeDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(item.TradeDate), time.Local)
		if err != nil || symbol == "" {
			continue
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if sourceKey == "" {
			sourceKey = "UNKNOWN"
		}
		id := item.ID
		if strings.TrimSpace(id) == "" {
			id = buildFuturesInventoryExternalID(
				sourceKey,
				symbol,
				tradeDate.Format("2006-01-02"),
				item.Warehouse,
				item.WarehouseID,
				item.Area,
				item.Brand,
				item.Place,
				item.Grade,
			)
		}
		_, err = r.db.Exec(`
INSERT INTO futures_inventory_snapshots
  (id, symbol, trade_date, futures_name, warehouse, warehouse_id, area, brand, place, grade, unit, receipt_volume, previous_volume, change_volume, source_key, fetched_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  futures_name = VALUES(futures_name),
  area = VALUES(area),
  brand = VALUES(brand),
  place = VALUES(place),
  grade = VALUES(grade),
  unit = VALUES(unit),
  receipt_volume = VALUES(receipt_volume),
  previous_volume = VALUES(previous_volume),
  change_volume = VALUES(change_volume),
  source_key = VALUES(source_key),
  fetched_at = VALUES(fetched_at),
  updated_at = VALUES(updated_at)`,
			id,
			symbol,
			tradeDate,
			nullableString(strings.TrimSpace(item.FuturesName)),
			nullableString(strings.TrimSpace(item.Warehouse)),
			nullableString(strings.TrimSpace(item.WarehouseID)),
			nullableString(strings.TrimSpace(item.Area)),
			nullableString(strings.TrimSpace(item.Brand)),
			nullableString(strings.TrimSpace(item.Place)),
			nullableString(strings.TrimSpace(item.Grade)),
			nullableString(strings.TrimSpace(item.Unit)),
			roundTo(item.ReceiptVolume, 4),
			roundTo(item.PreviousVolume, 4),
			roundTo(item.ChangeVolume, 4),
			sourceKey,
			now,
			now,
			now,
		)
		if err != nil {
			if isTableNotFoundError(err) {
				return 0, errors.New("futures_inventory_snapshots table is missing, please run migrations")
			}
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) upsertMarketDailyBars(items []model.MarketDailyBar) (int, error) {
	now := time.Now()
	affected := 0
	for _, item := range items {
		instrumentKey := strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		assetClass := strings.ToUpper(strings.TrimSpace(item.AssetClass))
		if instrumentKey == "" || assetClass == "" {
			continue
		}
		tradeDate, err := parseFlexibleDateTime(item.TradeDate)
		if err != nil {
			return affected, err
		}
		if item.ClosePrice <= 0 {
			continue
		}
		externalSymbol := strings.ToUpper(strings.TrimSpace(item.ExternalSymbol))
		if externalSymbol == "" {
			externalSymbol = deriveDefaultExternalSymbol(item.SourceKey, assetClass, instrumentKey)
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if sourceKey == "" {
			sourceKey = "UNKNOWN"
		}
		fetchedAt := now
		if parsedFetchedAt, err := parseFlexibleDateTime(item.FetchedAt); err == nil {
			fetchedAt = parsedFetchedAt
		}
		_, err = r.db.Exec(`
INSERT INTO market_daily_bars
  (id, asset_class, instrument_key, external_symbol, trade_date, open_price, high_price, low_price, close_price, prev_close_price, settle_price, prev_settle_price, volume, turnover, open_interest, source_key, fetched_at, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  external_symbol = VALUES(external_symbol),
  open_price = VALUES(open_price),
  high_price = VALUES(high_price),
  low_price = VALUES(low_price),
  close_price = VALUES(close_price),
  prev_close_price = VALUES(prev_close_price),
  settle_price = VALUES(settle_price),
  prev_settle_price = VALUES(prev_settle_price),
  volume = VALUES(volume),
  turnover = VALUES(turnover),
  open_interest = VALUES(open_interest),
  fetched_at = VALUES(fetched_at),
  updated_at = VALUES(updated_at)`,
			newID("mdb"),
			assetClass,
			instrumentKey,
			externalSymbol,
			tradeDate.Format("2006-01-02"),
			item.OpenPrice,
			item.HighPrice,
			item.LowPrice,
			item.ClosePrice,
			nullableFloat(item.PrevClosePrice),
			nullableFloat(item.SettlePrice),
			nullableFloat(item.PrevSettlePrice),
			item.Volume,
			item.Turnover,
			item.OpenInterest,
			sourceKey,
			fetchedAt,
			now,
			now,
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func nullableFloat(value float64) interface{} {
	if value == 0 {
		return nil
	}
	return value
}

func (r *MySQLGrowthRepo) rebuildMarketDailyBarTruth(assetClass string, touched map[string]marketTouchedBarKey, priority []string) ([]model.MarketDailyBar, error) {
	keys := make([]marketTouchedBarKey, 0, len(touched))
	for _, item := range touched {
		keys = append(keys, item)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].InstrumentKey == keys[j].InstrumentKey {
			return keys[i].TradeDate < keys[j].TradeDate
		}
		return keys[i].InstrumentKey < keys[j].InstrumentKey
	})

	priorityIndex := make(map[string]int, len(priority))
	for idx, sourceKey := range priority {
		priorityIndex[strings.ToUpper(strings.TrimSpace(sourceKey))] = idx
	}

	selectedBars := make([]model.MarketDailyBar, 0, len(keys))
	now := time.Now()
	for _, key := range keys {
		rows, err := r.db.Query(`
SELECT instrument_key, external_symbol, trade_date, open_price, high_price, low_price, close_price, COALESCE(prev_close_price, 0), COALESCE(settle_price, 0), COALESCE(prev_settle_price, 0), volume, turnover, COALESCE(open_interest, 0), source_key
FROM market_daily_bars
WHERE asset_class = ? AND instrument_key = ? AND trade_date = ?
ORDER BY updated_at DESC, id DESC`, assetClass, key.InstrumentKey, key.TradeDate)
		if err != nil {
			return nil, err
		}
		candidates := make([]model.MarketDailyBar, 0)
		for rows.Next() {
			var item model.MarketDailyBar
			var tradeDate time.Time
			if err := rows.Scan(
				&item.InstrumentKey,
				&item.ExternalSymbol,
				&tradeDate,
				&item.OpenPrice,
				&item.HighPrice,
				&item.LowPrice,
				&item.ClosePrice,
				&item.PrevClosePrice,
				&item.SettlePrice,
				&item.PrevSettlePrice,
				&item.Volume,
				&item.Turnover,
				&item.OpenInterest,
				&item.SourceKey,
			); err != nil {
				_ = rows.Close()
				return nil, err
			}
			item.AssetClass = assetClass
			item.TradeDate = tradeDate.Format("2006-01-02")
			candidates = append(candidates, item)
		}
		_ = rows.Close()
		if len(candidates) == 0 {
			continue
		}
		sort.SliceStable(candidates, func(i, j int) bool {
			left, ok := priorityIndex[strings.ToUpper(strings.TrimSpace(candidates[i].SourceKey))]
			if !ok {
				left = len(priorityIndex) + 10
			}
			right, ok := priorityIndex[strings.ToUpper(strings.TrimSpace(candidates[j].SourceKey))]
			if !ok {
				right = len(priorityIndex) + 10
			}
			return left < right
		})
		selected := candidates[0]
		_, err = r.db.Exec(`
INSERT INTO market_daily_bar_truth
  (id, asset_class, instrument_key, trade_date, selected_source_key, external_symbol, open_price, high_price, low_price, close_price, prev_close_price, settle_price, prev_settle_price, volume, turnover, open_interest, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  selected_source_key = VALUES(selected_source_key),
  external_symbol = VALUES(external_symbol),
  open_price = VALUES(open_price),
  high_price = VALUES(high_price),
  low_price = VALUES(low_price),
  close_price = VALUES(close_price),
  prev_close_price = VALUES(prev_close_price),
  settle_price = VALUES(settle_price),
  prev_settle_price = VALUES(prev_settle_price),
  volume = VALUES(volume),
  turnover = VALUES(turnover),
  open_interest = VALUES(open_interest),
  updated_at = VALUES(updated_at)`,
			newID("mdt"),
			assetClass,
			selected.InstrumentKey,
			selected.TradeDate,
			selected.SourceKey,
			selected.ExternalSymbol,
			selected.OpenPrice,
			selected.HighPrice,
			selected.LowPrice,
			selected.ClosePrice,
			nullableFloat(selected.PrevClosePrice),
			nullableFloat(selected.SettlePrice),
			nullableFloat(selected.PrevSettlePrice),
			selected.Volume,
			selected.Turnover,
			selected.OpenInterest,
			now,
			now,
		)
		if err != nil {
			return nil, err
		}
		selectedBars = append(selectedBars, selected)
	}
	return selectedBars, nil
}

func (r *MySQLGrowthRepo) syncLegacyStockQuotesFromTruthBars(items []model.MarketDailyBar) error {
	quotes := make([]model.StockMarketQuote, 0, len(items))
	for _, item := range items {
		if strings.ToUpper(strings.TrimSpace(item.AssetClass)) != marketAssetClassStock {
			continue
		}
		quotes = append(quotes, model.StockMarketQuote{
			Symbol:         item.InstrumentKey,
			TradeDate:      item.TradeDate,
			OpenPrice:      item.OpenPrice,
			HighPrice:      item.HighPrice,
			LowPrice:       item.LowPrice,
			ClosePrice:     item.ClosePrice,
			PrevClosePrice: item.PrevClosePrice,
			Volume:         item.Volume,
			Turnover:       item.Turnover,
			SourceKey:      item.SourceKey,
		})
	}
	if len(quotes) == 0 {
		return nil
	}
	_, err := r.upsertStockMarketQuotes(quotes)
	return err
}

func (r *MySQLGrowthRepo) upsertMarketNewsItems(items []model.MarketNewsItem) (int, error) {
	now := time.Now()
	affected := 0
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if title == "" || sourceKey == "" {
			continue
		}
		externalID := strings.TrimSpace(item.ExternalID)
		if externalID == "" {
			publishedAt, _ := parseFlexibleDateTime(item.PublishedAt)
			externalID = buildMarketNewsExternalID(sourceKey, item.PrimarySymbol, title, publishedAt)
		}
		publishedAt, err := parseFlexibleDateTime(item.PublishedAt)
		if err != nil {
			continue
		}
		symbolsJSON := marshalJSONSilently(item.Symbols)
		_, err = r.db.Exec(`
INSERT INTO market_news_items
  (id, source_key, external_id, news_type, title, summary, content, url, primary_symbol, symbols_json, metadata_json, published_at, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  news_type = VALUES(news_type),
  title = VALUES(title),
  summary = VALUES(summary),
  content = VALUES(content),
  url = VALUES(url),
  primary_symbol = VALUES(primary_symbol),
  symbols_json = VALUES(symbols_json),
  published_at = VALUES(published_at),
  updated_at = VALUES(updated_at)`,
			newID("mni"),
			sourceKey,
			externalID,
			coalesceUpper(item.NewsType, "MARKET"),
			title,
			nullableString(strings.TrimSpace(item.Summary)),
			nullableString(strings.TrimSpace(item.Content)),
			nullableString(strings.TrimSpace(item.URL)),
			nullableString(strings.ToUpper(strings.TrimSpace(item.PrimarySymbol))),
			nullableString(symbolsJSON),
			publishedAt,
			now,
			now,
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func coalesceUpper(value string, fallback string) string {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return fallback
	}
	return normalized
}

func (r *MySQLGrowthRepo) insertMarketSourceSnapshot(sourceKey string, assetClass string, dataKind string, instrumentKey string, externalSymbol string, status string, errorMessage string, payload string, fetchedAt time.Time) error {
	_, err := r.db.Exec(`
INSERT INTO market_source_snapshots
  (id, source_key, asset_class, data_kind, instrument_key, external_symbol, status, error_message, payload_text, fetched_at, created_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("mss"),
		strings.ToUpper(strings.TrimSpace(sourceKey)),
		nullableString(strings.ToUpper(strings.TrimSpace(assetClass))),
		strings.ToUpper(strings.TrimSpace(dataKind)),
		nullableString(strings.ToUpper(strings.TrimSpace(instrumentKey))),
		nullableString(strings.ToUpper(strings.TrimSpace(externalSymbol))),
		coalesceUpper(status, "SUCCESS"),
		nullableString(truncateString(strings.TrimSpace(errorMessage), 255)),
		nullableString(payload),
		fetchedAt,
		time.Now(),
	)
	return err
}

func truncateString(value string, max int) string {
	if max <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= max {
		return value
	}
	return string(runes[:max])
}

func marshalJSONSilently(value interface{}) string {
	if value == nil {
		return ""
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(payload)
}

func performAkshareDataSourceHealthCheckAttempt(config map[string]interface{}, timeoutMS int) model.DataSourceHealthCheck {
	start := time.Now()
	pythonBin, scriptPath, _, err := resolvePythonBridgeRuntime(config)
	if err != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Message:         err.Error(),
			FailureCategory: "CONFIG_ERROR",
		}
	}
	output, err := runPythonBridgeCommand(pythonBin, []string{scriptPath, "healthcheck"}, timeoutMS)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       false,
			LatencyMS:       latency,
			Message:         err.Error(),
			FailureCategory: classifyDataSourceRequestError(err),
		}
	}
	parsed := map[string]interface{}{}
	if jsonErr := json.Unmarshal(output, &parsed); jsonErr != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       true,
			LatencyMS:       latency,
			Message:         "invalid akshare health response",
			FailureCategory: "RESPONSE_PARSE_ERROR",
		}
	}
	return model.DataSourceHealthCheck{
		Status:     "HEALTHY",
		Reachable:  true,
		LatencyMS:  latency,
		Message:    "ok",
		HTTPStatus: 200,
	}
}
