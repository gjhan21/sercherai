package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

const (
	strategyEngineStockNewsWindowDays   = 14
	strategyEngineFuturesNewsWindowDays = 14
)

func (r *MySQLGrowthRepo) BuildStrategyEngineStockSelectionContext(input model.StrategyEngineStockSelectionContextRequest) (model.StrategyEngineStockSelectionContextResponse, error) {
	input = input.Normalized()
	requestedTradeDate, err := parseStrategyContextTradeDate(input.TradeDate)
	if err != nil {
		return model.StrategyEngineStockSelectionContextResponse{}, err
	}
	requestedLimit := input.Limit
	if requestedLimit <= 0 {
		requestedLimit = 10
	}
	if requestedLimit > 50 {
		requestedLimit = 50
	}

	selectionMode := resolveStrategyStockSelectionMode(input.SelectionMode, input.SeedSymbols, input.DebugSeedSymbols)
	autoMode := selectionMode == model.StrategyEngineStockSelectionModeAuto
	includeSymbols := resolveStrategyStockContextSymbols(selectionMode, input.SeedSymbols, input.DebugSeedSymbols)
	excludeSymbols := symbolSet(normalizeStockSymbolList(input.ExcludedSymbols))
	filteredInclude := make([]string, 0, len(includeSymbols))
	for _, symbol := range includeSymbols {
		if _, blocked := excludeSymbols[symbol]; blocked {
			continue
		}
		filteredInclude = append(filteredInclude, symbol)
	}
	if len(includeSymbols) > 0 && len(filteredInclude) == 0 {
		return model.StrategyEngineStockSelectionContextResponse{}, fmt.Errorf("stock selection seed symbols are empty after exclusion")
	}

	selectedTradeDate, err := r.resolveStrategyStockContextTradeDate(requestedTradeDate, filteredInclude)
	if err != nil {
		return model.StrategyEngineStockSelectionContextResponse{}, err
	}

	candidateLimit := requestedLimit
	if len(filteredInclude) == 0 {
		candidateLimit = requestedLimit * 80
		if candidateLimit < 180 {
			candidateLimit = 180
		}
		if candidateLimit > 500 {
			candidateLimit = 500
		}
	}

	candidates, err := r.loadStrategyStockContextCandidates(selectedTradeDate, filteredInclude, excludeSymbols, candidateLimit)
	if err != nil {
		return model.StrategyEngineStockSelectionContextResponse{}, err
	}
	if len(candidates) == 0 {
		return model.StrategyEngineStockSelectionContextResponse{}, fmt.Errorf("no stock truth bars available on %s", selectedTradeDate.Format("2006-01-02"))
	}
	if len(filteredInclude) > 0 {
		missingSymbols := missingCandidateSymbols(filteredInclude, candidates)
		if len(missingSymbols) > 0 {
			return model.StrategyEngineStockSelectionContextResponse{}, fmt.Errorf("missing stock truth bars on %s for symbols: %s", selectedTradeDate.Format("2006-01-02"), strings.Join(missingSymbols, ", "))
		}
	}

	candidateSymbols := make([]string, 0, len(candidates))
	for _, item := range candidates {
		candidateSymbols = append(candidateSymbols, item.Symbol)
	}

	historyMap, err := r.loadStrategyStockTruthHistory(candidateSymbols, selectedTradeDate)
	if err != nil {
		return model.StrategyEngineStockSelectionContextResponse{}, err
	}
	statusTruthMap, err := r.loadStrategyStockStatusTruthMap(candidateSymbols, selectedTradeDate)
	if err != nil {
		return model.StrategyEngineStockSelectionContextResponse{}, err
	}
	listingDateMap := map[string]time.Time{}
	listingDaysFilterEnabled := true
	listingCoverageWarning := ""
	if autoMode {
		listingDateMap, err = r.loadStrategyStockListingDateMap(candidateSymbols)
		if err != nil {
			return model.StrategyEngineStockSelectionContextResponse{}, err
		}
		if input.MinListingDays > 0 {
			coverageStart, coverageErr := r.loadStrategyTruthCoverageStart(marketAssetClassStock)
			if coverageErr != nil {
				return model.StrategyEngineStockSelectionContextResponse{}, coverageErr
			}
			if !coverageStart.IsZero() {
				coverageDays := int(selectedTradeDate.Sub(coverageStart).Hours() / 24)
				if coverageDays < input.MinListingDays {
					listingDaysFilterEnabled = false
					listingCoverageWarning = fmt.Sprintf("自动股票池因本地行情覆盖仅 %d 天，已跳过上市天数代理过滤", coverageDays)
				}
			}
		}
	}
	dailyBasics, err := r.loadStrategyStockDailyBasicsAsOf(candidateSymbols, selectedTradeDate)
	if err != nil {
		return model.StrategyEngineStockSelectionContextResponse{}, err
	}
	moneyflows, err := r.loadStrategyStockMoneyflowsAsOf(candidateSymbols, selectedTradeDate)
	if err != nil {
		return model.StrategyEngineStockSelectionContextResponse{}, err
	}
	newsSignals, err := r.loadStrategyMarketNewsSignals(candidateSymbols, selectedTradeDate, strategyEngineStockNewsWindowDays)
	if err != nil {
		return model.StrategyEngineStockSelectionContextResponse{}, err
	}

	warnings := make([]string, 0)
	if listingCoverageWarning != "" {
		warnings = append(warnings, listingCoverageWarning)
	}
	seeds := make([]model.StrategyEngineStockSeed, 0, requestedLimit)
	priceSources := make(map[string]struct{})
	filterCounters := map[string]int{
		"listing_days": 0,
		"avg_turnover": 0,
		"suspended":    0,
		"st":           0,
	}
	for _, candidate := range candidates {
		quotes := historyMap[candidate.Symbol]
		statusTruth, hasStatusTruth := statusTruthMap[candidate.Symbol]
		if autoMode {
			filtered, filterKey := shouldFilterAutoStockUniverseCandidate(candidate, statusTruth, hasStatusTruth, selectedTradeDate, quotes, listingDateMap[candidate.Symbol], listingDaysFilterEnabled, input)
			if filtered {
				filterCounters[filterKey]++
				continue
			}
		}
		score, ok := buildStockQuantScore(candidate.Symbol, quotes)
		if !ok {
			if len(filteredInclude) > 0 {
				return model.StrategyEngineStockSelectionContextResponse{}, fmt.Errorf("insufficient stock truth history for %s on %s", candidate.Symbol, selectedTradeDate.Format("2006-01-02"))
			}
			warnings = appendUniqueText(warnings, fmt.Sprintf("%s 历史样本不足 20 个交易日，已跳过", candidate.Symbol))
			continue
		}
		if basic, ok := dailyBasics[candidate.Symbol]; ok {
			score.PeTTM = basic.PeTTM
			score.PB = basic.PB
			score.TurnoverRate = basic.TurnoverRate
		}
		if flow, ok := moneyflows[candidate.Symbol]; ok {
			score.NetMFAmount = flow.NetMFAmount
		}
		if signal, ok := newsSignals[candidate.Symbol]; ok {
			score.NewsHeat = signal.Heat
			score.PositiveNewsRate = signal.PositiveRate
		} else {
			score.NewsHeat = 0
			score.PositiveNewsRate = 0.5
		}
		priceSources[candidate.PriceSource] = struct{}{}
		suspendedProxy := candidate.Volume <= 0 || candidate.Turnover <= 0
		stRiskProxy := isSTRiskCandidate(candidate)
		if hasStatusTruth {
			suspendedProxy = statusTruth.IsSuspended
			stRiskProxy = statusTruth.IsST
		}
		seeds = append(seeds, model.StrategyEngineStockSeed{
			Symbol:           candidate.Symbol,
			Name:             candidate.Name,
			TradeDate:        selectedTradeDate.Format("2006-01-02"),
			ClosePrice:       roundTo(score.ClosePrice, 4),
			Momentum5:        roundTo(score.Momentum5, 4),
			Momentum20:       roundTo(score.Momentum20, 4),
			Volatility20:     roundTo(score.Volatility20, 4),
			VolumeRatio:      roundTo(score.VolumeRatio, 4),
			Drawdown20:       roundTo(score.Drawdown20, 4),
			TrendStrength:    roundTo(score.TrendStrength, 4),
			NetMFAmount:      roundTo(score.NetMFAmount, 4),
			PeTTM:            roundTo(score.PeTTM, 4),
			PB:               roundTo(score.PB, 4),
			TurnoverRate:     roundTo(score.TurnoverRate, 4),
			NewsHeat:         score.NewsHeat,
			PositiveNewsRate: roundTo(score.PositiveNewsRate, 4),
			ListingDays:      strategyStockListingDays(selectedTradeDate, listingDateMap[candidate.Symbol], quotes),
			AvgTurnover20:    roundTo(averageStockTurnover20(quotes), 4),
			SuspendedProxy:   suspendedProxy,
			STRiskProxy:      stRiskProxy,
			Industry:         candidate.Industry,
			Sector:           candidate.Sector,
			ThemeTags:        append([]string(nil), candidate.ThemeTags...),
			RiskFlags:        buildStrategyStockRiskFlags(candidate),
		})
		if len(filteredInclude) > 0 && len(seeds) >= requestedLimit {
			break
		}
	}

	if len(seeds) == 0 {
		return model.StrategyEngineStockSelectionContextResponse{}, fmt.Errorf("stock truth data does not provide enough 20-session history for %s", selectedTradeDate.Format("2006-01-02"))
	}
	if noNewsSignals(newsSignals, seeds) {
		warnings = appendUniqueText(warnings, fmt.Sprintf("最近 %d 天暂无市场资讯信号，已回退到中性默认值", strategyEngineStockNewsWindowDays))
	}
	if autoMode {
		if filterCounters["listing_days"] > 0 {
			warnings = appendUniqueText(warnings, fmt.Sprintf("自动股票池按上市天数代理过滤 %d 只股票", filterCounters["listing_days"]))
		}
		if filterCounters["avg_turnover"] > 0 {
			warnings = appendUniqueText(warnings, fmt.Sprintf("自动股票池按 20 日均成交额代理过滤 %d 只股票", filterCounters["avg_turnover"]))
		}
		if filterCounters["suspended"] > 0 {
			warnings = appendUniqueText(warnings, fmt.Sprintf("自动股票池按停牌/成交代理过滤 %d 只股票", filterCounters["suspended"]))
		}
		if filterCounters["st"] > 0 {
			warnings = appendUniqueText(warnings, fmt.Sprintf("自动股票池按 ST/风险警示代理过滤 %d 只股票", filterCounters["st"]))
		}
	}
	priceSource := summarizeStrategyPriceSource(priceSources)
	routingSummary := buildStrategyContextRoutingSummary(priceSource, marketAssetClassStock, marketDataKindDailyBars)

	return model.StrategyEngineStockSelectionContextResponse{
		Seeds: seeds,
		Meta: model.StrategyEngineStockSelectionContextMeta{
			SelectedTradeDate:        selectedTradeDate.Format("2006-01-02"),
			PriceSource:              priceSource,
			SelectedSource:           routingSummary.SelectedSource,
			FallbackSourceKeys:       append([]string(nil), routingSummary.FallbackSourceKeys...),
			RoutingPolicyKey:         routingSummary.RoutingPolicyKey,
			DecisionReason:           routingSummary.DecisionReason,
			NewsWindowDays:           strategyEngineStockNewsWindowDays,
			ListingDaysFilterApplied: !autoMode || listingDaysFilterEnabled,
			Warnings:                 warnings,
		},
	}, nil
}

func resolveStrategyStockSelectionMode(raw string, seedSymbols []string, debugSeedSymbols []string) string {
	return model.StrategyEngineStockSelectionContextRequest{
		SelectionMode:    raw,
		SeedSymbols:      seedSymbols,
		DebugSeedSymbols: debugSeedSymbols,
	}.Normalized().SelectionMode
}

func resolveStrategyStockContextSymbols(selectionMode string, seedSymbols []string, debugSeedSymbols []string) []string {
	switch strings.ToUpper(strings.TrimSpace(selectionMode)) {
	case model.StrategyEngineStockSelectionModeDebug:
		normalizedDebug := normalizeStockSymbolList(debugSeedSymbols)
		if len(normalizedDebug) > 0 {
			return normalizedDebug
		}
		return normalizeStockSymbolList(seedSymbols)
	case model.StrategyEngineStockSelectionModeManual:
		return normalizeStockSymbolList(seedSymbols)
	default:
		return nil
	}
}

func (r *MySQLGrowthRepo) BuildStrategyEngineFuturesStrategyContext(input model.StrategyEngineFuturesStrategyContextRequest) (model.StrategyEngineFuturesStrategyContextResponse, error) {
	requestedTradeDate, err := parseStrategyContextTradeDate(input.TradeDate)
	if err != nil {
		return model.StrategyEngineFuturesStrategyContextResponse{}, err
	}
	requestedLimit := input.Limit
	if requestedLimit <= 0 {
		requestedLimit = 5
	}
	if requestedLimit > 20 {
		requestedLimit = 20
	}

	includeContracts := normalizeFuturesContextContractList(input.Contracts)
	selectedTradeDate, mockFallback, err := r.resolveStrategyFuturesContextTradeDate(requestedTradeDate, includeContracts)
	if err != nil {
		return model.StrategyEngineFuturesStrategyContextResponse{}, err
	}
	resp, err := r.buildStrategyEngineFuturesStrategyContextResponse(requestedTradeDate, selectedTradeDate, includeContracts, requestedLimit, mockFallback, nil)
	if err == nil {
		return resp, nil
	}
	if !input.AllowMockFallbackOnShortHistory || len(includeContracts) == 0 || mockFallback || !strings.Contains(err.Error(), "insufficient futures truth history") {
		return model.StrategyEngineFuturesStrategyContextResponse{}, err
	}
	mockTradeDate, found, mockErr := r.queryStrategyFuturesContextTradeDateBySource(requestedTradeDate, includeContracts, "MOCK")
	if mockErr != nil {
		return model.StrategyEngineFuturesStrategyContextResponse{}, mockErr
	}
	if !found || !mockTradeDate.After(selectedTradeDate) {
		return model.StrategyEngineFuturesStrategyContextResponse{}, err
	}
	return r.buildStrategyEngineFuturesStrategyContextResponse(
		requestedTradeDate,
		mockTradeDate,
		includeContracts,
		requestedLimit,
		true,
		[]string{
			fmt.Sprintf("实盘期货在 %s 的历史少于 14 个交易日，显式允许后已切换到 MOCK 真相源", selectedTradeDate.Format("2006-01-02")),
		},
	)
}

func (r *MySQLGrowthRepo) buildStrategyEngineFuturesStrategyContextResponse(requestedTradeDate time.Time, selectedTradeDate time.Time, includeContracts []string, requestedLimit int, mockFallback bool, extraWarnings []string) (model.StrategyEngineFuturesStrategyContextResponse, error) {
	candidateLimit := requestedLimit
	if len(includeContracts) == 0 {
		candidateLimit = requestedLimit * 3
		if candidateLimit < requestedLimit {
			candidateLimit = requestedLimit
		}
		if candidateLimit > 30 {
			candidateLimit = 30
		}
	}

	candidateContracts := includeContracts
	if len(candidateContracts) == 0 {
		mappedContracts, err := r.loadStrategyFuturesMappedContracts(selectedTradeDate, candidateLimit, mockFallback)
		if err != nil {
			return model.StrategyEngineFuturesStrategyContextResponse{}, err
		}
		if len(mappedContracts) > 0 {
			candidateContracts = mappedContracts
		}
	}

	candidates, err := r.loadStrategyFuturesContextCandidates(selectedTradeDate, candidateContracts, candidateLimit, mockFallback)
	if err != nil {
		return model.StrategyEngineFuturesStrategyContextResponse{}, err
	}
	if len(candidates) == 0 {
		return model.StrategyEngineFuturesStrategyContextResponse{}, fmt.Errorf("no futures truth bars available on %s", selectedTradeDate.Format("2006-01-02"))
	}
	if len(includeContracts) > 0 {
		missingContracts := missingFuturesCandidateContracts(includeContracts, candidates)
		if len(missingContracts) > 0 {
			return model.StrategyEngineFuturesStrategyContextResponse{}, fmt.Errorf("missing futures truth bars on %s for contracts: %s", selectedTradeDate.Format("2006-01-02"), strings.Join(missingContracts, ", "))
		}
	}

	instrumentKeys := make([]string, 0, len(candidates))
	aliasMap := make(map[string][]string, len(candidates))
	for _, item := range candidates {
		instrumentKeys = append(instrumentKeys, item.InstrumentKey)
		aliasMap[item.Contract] = compactStrings([]string{item.Contract, item.InstrumentKey})
	}

	historyMap, err := r.loadStrategyFuturesTruthHistory(instrumentKeys, selectedTradeDate)
	if err != nil {
		return model.StrategyEngineFuturesStrategyContextResponse{}, err
	}
	newsSignals, err := r.loadStrategyFuturesNewsSignals(aliasMap, selectedTradeDate, strategyEngineFuturesNewsWindowDays)
	if err != nil {
		return model.StrategyEngineFuturesStrategyContextResponse{}, err
	}
	curveMetricsMap, err := r.loadStrategyFuturesCurveMetrics(selectedTradeDate, candidates, mockFallback)
	if err != nil {
		return model.StrategyEngineFuturesStrategyContextResponse{}, err
	}
	inventorySignalMap, err := r.loadStrategyFuturesInventorySignals(candidates, selectedTradeDate)
	if err != nil {
		return model.StrategyEngineFuturesStrategyContextResponse{}, err
	}
	spreadSignalMap, err := r.loadStrategyFuturesSpreadSignals(candidates)
	if err != nil {
		return model.StrategyEngineFuturesStrategyContextResponse{}, err
	}

	warnings := append([]string(nil), extraWarnings...)
	if mockFallback && len(warnings) == 0 {
		warnings = append(warnings, fmt.Sprintf("截至 %s 未找到实盘期货 bars，已回退到 MOCK 真相源", requestedTradeDate.Format("2006-01-02")))
	}
	seeds := make([]model.StrategyEngineFuturesSeed, 0, requestedLimit)
	priceSources := make(map[string]struct{})
	for _, candidate := range candidates {
		quotes := historyMap[candidate.InstrumentKey]
		newsSignal, ok := newsSignals[candidate.Contract]
		if !ok {
			newsSignal = futuresNewsSignal{Heat: 0, Bias: 0}
		}
		seed, usable := buildFuturesContextSeed(candidate, quotes, newsSignal, curveMetricsMap[candidate.Contract], inventorySignalMap[candidate.Contract], spreadSignalMap[candidate.Contract])
		if !usable {
			if len(includeContracts) > 0 {
				return model.StrategyEngineFuturesStrategyContextResponse{}, fmt.Errorf("insufficient futures truth history for %s on %s", candidate.Contract, selectedTradeDate.Format("2006-01-02"))
			}
			warnings = appendUniqueText(warnings, fmt.Sprintf("%s 历史样本不足 14 个交易日，已跳过", candidate.Contract))
			continue
		}
		priceSources[candidate.PriceSource] = struct{}{}
		seeds = append(seeds, seed)
		if len(includeContracts) == 0 && len(seeds) >= requestedLimit {
			break
		}
	}
	if len(seeds) == 0 {
		return model.StrategyEngineFuturesStrategyContextResponse{}, fmt.Errorf("futures truth data does not provide enough 14-session history for %s", selectedTradeDate.Format("2006-01-02"))
	}
	if noFuturesNewsSignals(newsSignals, seeds) {
		warnings = appendUniqueText(warnings, fmt.Sprintf("最近 %d 天暂无市场资讯信号，已回退到中性默认值", strategyEngineFuturesNewsWindowDays))
	}
	priceSource := summarizeStrategyPriceSource(priceSources)
	routingSummary := buildStrategyContextRoutingSummary(priceSource, marketAssetClassFutures, marketDataKindDailyBars)

	return model.StrategyEngineFuturesStrategyContextResponse{
		Seeds: seeds,
		Meta: model.StrategyEngineFuturesStrategyContextMeta{
			SelectedTradeDate:  selectedTradeDate.Format("2006-01-02"),
			PriceSource:        priceSource,
			SelectedSource:     routingSummary.SelectedSource,
			FallbackSourceKeys: append([]string(nil), routingSummary.FallbackSourceKeys...),
			RoutingPolicyKey:   routingSummary.RoutingPolicyKey,
			DecisionReason:     routingSummary.DecisionReason,
			NewsWindowDays:     strategyEngineFuturesNewsWindowDays,
			Warnings:           warnings,
		},
	}, nil
}

type strategyStockContextCandidate struct {
	Symbol      string
	Name        string
	PriceSource string
	Volume      float64
	Turnover    float64
	RiskWarning bool
	Industry    string
	Sector      string
	ThemeTags   []string
	RiskFlags   []string
}

type strategyFuturesContextCandidate struct {
	InstrumentKey string
	Contract      string
	Name          string
	PriceSource   string
}

type futuresQuoteCandle struct {
	InstrumentKey   string
	TradeDate       time.Time
	ClosePrice      float64
	PrevClosePrice  float64
	SettlePrice     float64
	PrevSettlePrice float64
	Volume          float64
	Turnover        float64
	OpenInterest    float64
}

type futuresQuoteSupplement struct {
	SourceKey       string
	SettlePrice     float64
	PrevSettlePrice float64
	Turnover        float64
	OpenInterest    float64
}

type futuresNewsSignal struct {
	Heat int
	Bias float64
}

type futuresCurveMetrics struct {
	TermStructurePct float64
	CurveSlopePct    float64
}

type futuresSpreadSignal struct {
	Pair       string
	Percentile float64
	Pressure   float64
}

type futuresInventorySignal struct {
	Level          float64
	ChangePct      float64
	Pressure       float64
	FocusArea      string
	FocusWarehouse string
	FocusBrand     string
	FocusPlace     string
	FocusGrade     string
	AreaShare      float64
	WarehouseShare float64
	BrandShare     float64
	PlaceShare     float64
	GradeShare     float64
}

type futuresTermStructureQuote struct {
	Contract   string
	Root       string
	ExpiryCode int
	ClosePrice float64
}

func (r *MySQLGrowthRepo) resolveStrategyStockContextTradeDate(requestedTradeDate time.Time, includeSymbols []string) (time.Time, error) {
	args := []any{marketAssetClassStock, requestedTradeDate.Format("2006-01-02")}
	query := `
SELECT MAX(trade_date)
FROM market_daily_bar_truth
WHERE asset_class = ? AND trade_date <= ?`
	if len(includeSymbols) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(includeSymbols)), ",")
		query += " AND instrument_key IN (" + placeholders + ")"
		for _, symbol := range includeSymbols {
			args = append(args, symbol)
		}
	}
	var selected sql.NullTime
	if err := r.db.QueryRow(query, args...).Scan(&selected); err != nil {
		return time.Time{}, err
	}
	if !selected.Valid {
		return time.Time{}, fmt.Errorf("no stock truth bars available on or before %s", requestedTradeDate.Format("2006-01-02"))
	}
	return selected.Time, nil
}

func (r *MySQLGrowthRepo) loadStrategyStockContextCandidates(selectedTradeDate time.Time, includeSymbols []string, excludeSymbols map[string]struct{}, limit int) ([]strategyStockContextCandidate, error) {
	args := []any{marketAssetClassStock, selectedTradeDate.Format("2006-01-02")}
	query := `
SELECT t.instrument_key,
       COALESCE(NULLIF(mi.display_name, ''), t.instrument_key) AS display_name,
       t.selected_source_key,
       COALESCE(t.volume, 0),
       COALESCE(t.turnover, 0),
       COALESCE(CAST(mi.metadata_json AS CHAR), '')
FROM market_daily_bar_truth t
LEFT JOIN market_instruments mi
  ON mi.asset_class = t.asset_class AND mi.instrument_key = t.instrument_key
WHERE t.asset_class = ? AND t.trade_date = ?`
	if len(includeSymbols) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(includeSymbols)), ",")
		query += " AND t.instrument_key IN (" + placeholders + ")"
		for _, symbol := range includeSymbols {
			args = append(args, symbol)
		}
		query += " ORDER BY t.instrument_key ASC"
	} else {
		query += " ORDER BY t.turnover DESC, t.instrument_key ASC LIMIT ?"
		args = append(args, limit)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]strategyStockContextCandidate, 0)
	for rows.Next() {
		var item strategyStockContextCandidate
		var metadataJSON string
		if err := rows.Scan(&item.Symbol, &item.Name, &item.PriceSource, &item.Volume, &item.Turnover, &metadataJSON); err != nil {
			return nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
		item.Name = strings.TrimSpace(item.Name)
		item.PriceSource = strings.ToUpper(strings.TrimSpace(item.PriceSource))
		item.Industry, item.Sector, item.ThemeTags, item.RiskFlags, item.RiskWarning = parseStockInstrumentMetadata(metadataJSON)
		if item.Symbol == "" {
			continue
		}
		if _, blocked := excludeSymbols[item.Symbol]; blocked {
			continue
		}
		if item.Name == "" {
			item.Name = item.Symbol
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(includeSymbols) == 0 && len(items) > 1 {
		items = dedupeStrategyStockContextCandidates(items)
	}
	if len(includeSymbols) == 0 || len(items) <= 1 {
		return items, nil
	}
	order := make(map[string]int, len(includeSymbols))
	for index, symbol := range includeSymbols {
		order[symbol] = index
	}
	sort.SliceStable(items, func(i, j int) bool {
		return order[items[i].Symbol] < order[items[j].Symbol]
	})
	return items, nil
}

func dedupeStrategyStockContextCandidates(items []strategyStockContextCandidate) []strategyStockContextCandidate {
	seen := make(map[string]struct{}, len(items))
	result := make([]strategyStockContextCandidate, 0, len(items))
	for _, item := range items {
		key := stockContextCandidateDedupKey(item.Symbol)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	return result
}

func stockContextCandidateDedupKey(symbol string) string {
	text := strings.ToUpper(strings.TrimSpace(symbol))
	if text == "" {
		return ""
	}
	if idx := strings.Index(text, "."); idx > 0 {
		return text[:idx]
	}
	return text
}

func (r *MySQLGrowthRepo) loadStrategyStockListingDateMap(symbols []string) (map[string]time.Time, error) {
	if len(symbols) == 0 {
		return map[string]time.Time{}, nil
	}
	result := make(map[string]time.Time, len(symbols))
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	args := make([]any, 0, len(symbols)+1)
	args = append(args, marketAssetClassStock)
	for _, symbol := range symbols {
		args = append(args, symbol)
	}

	instrumentQuery := fmt.Sprintf(`
SELECT instrument_key, list_date
FROM market_instruments
WHERE asset_class = ? AND instrument_key IN (%s)`, placeholders)
	rows, err := r.db.Query(instrumentQuery, args...)
	if err != nil {
		if !isMarketStatusSchemaCompatError(err) {
			return nil, err
		}
	} else {
		defer rows.Close()
		for rows.Next() {
			var symbol string
			var listDate sql.NullTime
			if err := rows.Scan(&symbol, &listDate); err != nil {
				return nil, err
			}
			if !listDate.Valid {
				continue
			}
			result[strings.ToUpper(strings.TrimSpace(symbol))] = listDate.Time
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	missingSymbols := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		if _, ok := result[normalized]; ok {
			continue
		}
		missingSymbols = append(missingSymbols, normalized)
	}
	if len(missingSymbols) == 0 {
		return result, nil
	}

	baseSymbols := make([]string, 0, len(missingSymbols))
	baseSeen := make(map[string]struct{}, len(missingSymbols))
	missingByBase := make(map[string][]string, len(missingSymbols))
	for _, symbol := range missingSymbols {
		base := stockContextCandidateDedupKey(symbol)
		if base == "" {
			continue
		}
		if _, ok := baseSeen[base]; !ok {
			baseSeen[base] = struct{}{}
			baseSymbols = append(baseSymbols, base)
		}
		missingByBase[base] = append(missingByBase[base], symbol)
	}

	placeholders = strings.TrimSuffix(strings.Repeat("?,", len(missingSymbols)), ",")
	basePlaceholders := strings.TrimSuffix(strings.Repeat("?,", len(baseSymbols)), ",")
	args = make([]any, 0, len(missingSymbols)+len(baseSymbols)+1)
	args = append(args, marketAssetClassStock)
	for _, symbol := range missingSymbols {
		args = append(args, symbol)
	}
	for _, base := range baseSymbols {
		args = append(args, base)
	}
	query := fmt.Sprintf(`
SELECT instrument_key, MIN(trade_date)
FROM market_daily_bar_truth
WHERE asset_class = ?
  AND (instrument_key IN (%s) OR SUBSTRING_INDEX(instrument_key, '.', 1) IN (%s))
GROUP BY instrument_key`, placeholders, basePlaceholders)
	rows, err = r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var symbol string
		var firstTradeDate time.Time
		if err := rows.Scan(&symbol, &firstTradeDate); err != nil {
			return nil, err
		}
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized != "" {
			if existing, ok := result[normalized]; !ok || firstTradeDate.Before(existing) {
				result[normalized] = firstTradeDate
			}
		}
		base := stockContextCandidateDedupKey(normalized)
		for _, missingSymbol := range missingByBase[base] {
			if existing, ok := result[missingSymbol]; !ok || firstTradeDate.Before(existing) {
				result[missingSymbol] = firstTradeDate
			}
		}
	}
	return result, rows.Err()
}

func (r *MySQLGrowthRepo) loadStrategyFuturesMappedContracts(selectedTradeDate time.Time, limit int, allowMock bool) ([]string, error) {
	if limit <= 0 {
		return nil, nil
	}
	args := []any{marketAssetClassFutures, selectedTradeDate.Format("2006-01-02")}
	query := `
SELECT m.dominant_instrument_key, COALESCE(t.turnover, 0)
FROM futures_contract_mappings m
LEFT JOIN market_daily_bar_truth t
  ON t.asset_class = ? AND t.trade_date = m.trade_date AND t.instrument_key = m.dominant_instrument_key
WHERE m.trade_date = ?`
	if !allowMock {
		query += " AND UPPER(COALESCE(m.selected_source_key, '')) <> ?"
		args = append(args, "MOCK")
	}
	query += " ORDER BY COALESCE(t.turnover, 0) DESC, m.product_key ASC LIMIT ?"
	args = append(args, limit)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isMarketStatusSchemaCompatError(err) {
			return nil, nil
		}
		return nil, err
	}
	defer rows.Close()

	contracts := make([]string, 0, limit)
	seen := make(map[string]struct{}, limit)
	for rows.Next() {
		var instrumentKey string
		var turnover float64
		if err := rows.Scan(&instrumentKey, &turnover); err != nil {
			return nil, err
		}
		contract := normalizeFuturesContextContract(instrumentKey)
		if contract == "" {
			continue
		}
		if _, ok := seen[contract]; ok {
			continue
		}
		seen[contract] = struct{}{}
		contracts = append(contracts, contract)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return contracts, nil
}

func (r *MySQLGrowthRepo) loadStrategyTruthCoverageStart(assetClass string) (time.Time, error) {
	var firstTradeDate sql.NullTime
	if err := r.db.QueryRow(`
SELECT MIN(trade_date)
FROM market_daily_bar_truth
WHERE asset_class = ?`, strings.TrimSpace(assetClass)).Scan(&firstTradeDate); err != nil {
		return time.Time{}, err
	}
	if !firstTradeDate.Valid {
		return time.Time{}, nil
	}
	return firstTradeDate.Time, nil
}

func shouldFilterAutoStockUniverseCandidate(
	candidate strategyStockContextCandidate,
	statusTruth strategyStockStatusTruth,
	hasStatusTruth bool,
	selectedTradeDate time.Time,
	quotes []stockQuoteCandle,
	firstTradeDate time.Time,
	listingDaysFilterEnabled bool,
	input model.StrategyEngineStockSelectionContextRequest,
) (bool, string) {
	if hasStatusTruth {
		if statusTruth.IsSuspended {
			return true, "suspended"
		}
		if statusTruth.IsST {
			return true, "st"
		}
	} else if candidate.Volume <= 0 || candidate.Turnover <= 0 {
		return true, "suspended"
	} else if isSTRiskCandidate(candidate) {
		return true, "st"
	}
	if listingDaysFilterEnabled && !firstTradeDate.IsZero() {
		listingDays := int(selectedTradeDate.Sub(firstTradeDate).Hours() / 24)
		if listingDays < input.MinListingDays {
			return true, "listing_days"
		}
	}
	if averageStockTurnover20(quotes) < input.MinAvgTurnover {
		return true, "avg_turnover"
	}
	return false, ""
}

func averageStockTurnover20(quotes []stockQuoteCandle) float64 {
	if len(quotes) == 0 {
		return 0
	}
	total := 0.0
	count := 0
	for index := len(quotes) - 1; index >= 0 && count < 20; index-- {
		if quotes[index].Turnover <= 0 {
			continue
		}
		total += quotes[index].Turnover
		count++
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

func isSTRiskCandidate(candidate strategyStockContextCandidate) bool {
	name := strings.ToUpper(strings.TrimSpace(candidate.Name))
	return strings.HasPrefix(name, "ST") || strings.HasPrefix(name, "*ST") || candidate.RiskWarning
}

func parseStockInstrumentMetadata(raw string) (string, string, []string, []string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "null" {
		return "", "", nil, nil, false
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", "", nil, nil, false
	}
	industry := firstNonEmpty(strings.TrimSpace(asString(payload["industry"])), strings.TrimSpace(asString(payload["industry_name"])))
	sector := firstNonEmpty(strings.TrimSpace(asString(payload["sector"])), strings.TrimSpace(asString(payload["sector_name"])))
	themeTags := compactStrings(append(append(stringSlice(payload["theme_tags"]), stringSlice(payload["themes"])...), stringSlice(payload["concept_tags"])...))
	riskFlags := compactStrings(stringSlice(payload["risk_flags"]))
	riskWarning := parseStockRiskWarningValue(payload["risk_warning"])
	if riskWarning {
		riskFlags = appendUniqueText(riskFlags, "风险警示")
	}
	return industry, sector, themeTags, riskFlags, riskWarning
}

func parseStockRiskWarningValue(value any) bool {
	if value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	case float64:
		return typed != 0
	default:
		return false
	}
}

func strategyStockListingDays(selectedTradeDate time.Time, firstTradeDate time.Time, quotes []stockQuoteCandle) int {
	if firstTradeDate.IsZero() && len(quotes) > 0 {
		firstTradeDate = quotes[0].TradeDate
	}
	if firstTradeDate.IsZero() {
		return 0
	}
	return int(selectedTradeDate.Sub(firstTradeDate).Hours() / 24)
}

func buildStrategyStockRiskFlags(candidate strategyStockContextCandidate) []string {
	flags := append([]string(nil), candidate.RiskFlags...)
	if candidate.Volume <= 0 || candidate.Turnover <= 0 {
		flags = appendUniqueText(flags, "流动性异常")
	}
	if isSTRiskCandidate(candidate) {
		flags = appendUniqueText(flags, "ST代理风险")
	}
	return compactStrings(flags)
}

func (r *MySQLGrowthRepo) resolveStrategyFuturesContextTradeDate(requestedTradeDate time.Time, includeContracts []string) (time.Time, bool, error) {
	selectedTradeDate, found, err := r.queryStrategyFuturesContextTradeDate(requestedTradeDate, includeContracts, true)
	if err != nil {
		return time.Time{}, false, err
	}
	if found {
		return selectedTradeDate, false, nil
	}
	selectedTradeDate, found, err = r.queryStrategyFuturesContextTradeDate(requestedTradeDate, includeContracts, false)
	if err != nil {
		return time.Time{}, false, err
	}
	if !found {
		return time.Time{}, false, fmt.Errorf("no futures truth bars available on or before %s", requestedTradeDate.Format("2006-01-02"))
	}
	return selectedTradeDate, true, nil
}

func (r *MySQLGrowthRepo) queryStrategyFuturesContextTradeDate(requestedTradeDate time.Time, includeContracts []string, excludeMock bool) (time.Time, bool, error) {
	args := []any{marketAssetClassFutures, requestedTradeDate.Format("2006-01-02")}
	query := `
SELECT MAX(trade_date)
FROM market_daily_bar_truth
WHERE asset_class = ? AND trade_date <= ?`
	if excludeMock {
		query += " AND UPPER(selected_source_key) <> ?"
		args = append(args, "MOCK")
	}
	if len(includeContracts) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(includeContracts)), ",")
		query += " AND UPPER(SUBSTRING_INDEX(instrument_key, '.', 1)) IN (" + placeholders + ")"
		for _, contract := range includeContracts {
			args = append(args, contract)
		}
	}
	var selected sql.NullTime
	if err := r.db.QueryRow(query, args...).Scan(&selected); err != nil {
		return time.Time{}, false, err
	}
	if !selected.Valid {
		return time.Time{}, false, nil
	}
	return selected.Time, true, nil
}

func (r *MySQLGrowthRepo) queryStrategyFuturesContextTradeDateBySource(requestedTradeDate time.Time, includeContracts []string, sourceKey string) (time.Time, bool, error) {
	sourceKey = strings.ToUpper(strings.TrimSpace(sourceKey))
	if sourceKey == "" {
		return time.Time{}, false, nil
	}
	args := []any{marketAssetClassFutures, requestedTradeDate.Format("2006-01-02"), sourceKey}
	query := `
SELECT MAX(trade_date)
FROM market_daily_bar_truth
WHERE asset_class = ? AND trade_date <= ? AND UPPER(selected_source_key) = ?`
	if len(includeContracts) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(includeContracts)), ",")
		query += " AND UPPER(SUBSTRING_INDEX(instrument_key, '.', 1)) IN (" + placeholders + ")"
		for _, contract := range includeContracts {
			args = append(args, contract)
		}
	}
	var selected sql.NullTime
	if err := r.db.QueryRow(query, args...).Scan(&selected); err != nil {
		return time.Time{}, false, err
	}
	if !selected.Valid {
		return time.Time{}, false, nil
	}
	return selected.Time, true, nil
}

func (r *MySQLGrowthRepo) loadStrategyFuturesContextCandidates(selectedTradeDate time.Time, includeContracts []string, limit int, allowMock bool) ([]strategyFuturesContextCandidate, error) {
	args := []any{marketAssetClassFutures, selectedTradeDate.Format("2006-01-02")}
	query := `
SELECT t.instrument_key,
       SUBSTRING_INDEX(t.instrument_key, '.', 1) AS contract_key,
       COALESCE(NULLIF(mi.display_name, ''), SUBSTRING_INDEX(t.instrument_key, '.', 1)) AS display_name,
       t.selected_source_key
FROM market_daily_bar_truth t
LEFT JOIN market_instruments mi
  ON mi.asset_class = t.asset_class AND mi.instrument_key = t.instrument_key
WHERE t.asset_class = ? AND t.trade_date = ?`
	if !allowMock {
		query += " AND UPPER(t.selected_source_key) <> ?"
		args = append(args, "MOCK")
	}
	if len(includeContracts) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(includeContracts)), ",")
		query += " AND UPPER(SUBSTRING_INDEX(t.instrument_key, '.', 1)) IN (" + placeholders + ")"
		for _, contract := range includeContracts {
			args = append(args, contract)
		}
		query += " ORDER BY t.instrument_key ASC"
	} else {
		query += " ORDER BY t.turnover DESC, t.instrument_key ASC LIMIT ?"
		args = append(args, limit)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]strategyFuturesContextCandidate, 0)
	for rows.Next() {
		var item strategyFuturesContextCandidate
		if err := rows.Scan(&item.InstrumentKey, &item.Contract, &item.Name, &item.PriceSource); err != nil {
			return nil, err
		}
		item.InstrumentKey = strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		item.Contract = normalizeFuturesContextContract(item.Contract)
		item.Name = strings.TrimSpace(item.Name)
		item.PriceSource = strings.ToUpper(strings.TrimSpace(item.PriceSource))
		if item.InstrumentKey == "" || item.Contract == "" {
			continue
		}
		if item.Name == "" {
			item.Name = item.Contract
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(includeContracts) == 0 || len(items) <= 1 {
		return items, nil
	}
	order := make(map[string]int, len(includeContracts))
	for index, contract := range includeContracts {
		order[contract] = index
	}
	sort.SliceStable(items, func(i, j int) bool {
		return order[items[i].Contract] < order[items[j].Contract]
	})
	return items, nil
}

func (r *MySQLGrowthRepo) loadStrategyFuturesTruthHistory(instrumentKeys []string, selectedTradeDate time.Time) (map[string][]futuresQuoteCandle, error) {
	if len(instrumentKeys) == 0 {
		return map[string][]futuresQuoteCandle{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(instrumentKeys)), ",")
	args := make([]any, 0, len(instrumentKeys)+3)
	args = append(args, marketAssetClassFutures)
	for _, instrumentKey := range instrumentKeys {
		args = append(args, instrumentKey)
	}
	args = append(args, selectedTradeDate.AddDate(0, 0, -45).Format("2006-01-02"), selectedTradeDate.Format("2006-01-02"))
	query := fmt.Sprintf(`
SELECT instrument_key, trade_date, close_price, prev_close_price, settle_price, prev_settle_price, volume, turnover, open_interest
FROM market_daily_bar_truth
WHERE asset_class = ?
  AND instrument_key IN (%s)
  AND trade_date >= ?
  AND trade_date <= ?
ORDER BY instrument_key ASC, trade_date ASC`, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]futuresQuoteCandle, len(instrumentKeys))
	for rows.Next() {
		var item futuresQuoteCandle
		var prevClose sql.NullFloat64
		var settle sql.NullFloat64
		var prevSettle sql.NullFloat64
		var turnover sql.NullFloat64
		var openInterest sql.NullFloat64
		if err := rows.Scan(
			&item.InstrumentKey,
			&item.TradeDate,
			&item.ClosePrice,
			&prevClose,
			&settle,
			&prevSettle,
			&item.Volume,
			&turnover,
			&openInterest,
		); err != nil {
			return nil, err
		}
		item.InstrumentKey = strings.ToUpper(strings.TrimSpace(item.InstrumentKey))
		item.PrevClosePrice = sqlNullFloat(prevClose)
		item.SettlePrice = sqlNullFloat(settle)
		item.PrevSettlePrice = sqlNullFloat(prevSettle)
		item.Turnover = sqlNullFloat(turnover)
		item.OpenInterest = sqlNullFloat(openInterest)
		if item.InstrumentKey == "" || item.ClosePrice <= 0 {
			continue
		}
		result[item.InstrumentKey] = append(result[item.InstrumentKey], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if !needsFuturesQuoteSupplement(result) {
		return result, nil
	}
	supplements, err := r.loadStrategyFuturesQuoteSupplements(instrumentKeys, selectedTradeDate)
	if err != nil {
		return nil, err
	}
	mergeFuturesQuoteSupplements(result, supplements)
	return result, nil
}

func needsFuturesQuoteSupplement(history map[string][]futuresQuoteCandle) bool {
	for _, quotes := range history {
		for _, item := range quotes {
			if item.ClosePrice <= 0 {
				continue
			}
			if item.SettlePrice <= 0 || item.PrevSettlePrice <= 0 || item.Turnover <= 0 || item.OpenInterest <= 0 {
				return true
			}
		}
	}
	return false
}

func (r *MySQLGrowthRepo) loadStrategyFuturesQuoteSupplements(instrumentKeys []string, selectedTradeDate time.Time) (map[string]map[string]futuresQuoteSupplement, error) {
	if len(instrumentKeys) == 0 {
		return map[string]map[string]futuresQuoteSupplement{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(instrumentKeys)), ",")
	args := make([]any, 0, len(instrumentKeys)+3)
	args = append(args, marketAssetClassFutures)
	for _, instrumentKey := range instrumentKeys {
		args = append(args, instrumentKey)
	}
	args = append(args, selectedTradeDate.AddDate(0, 0, -45).Format("2006-01-02"), selectedTradeDate.Format("2006-01-02"))
	query := fmt.Sprintf(`
SELECT instrument_key, trade_date, source_key, close_price, settle_price, prev_settle_price, turnover, open_interest
FROM market_daily_bars
WHERE asset_class = ?
  AND instrument_key IN (%s)
  AND trade_date >= ?
  AND trade_date <= ?
ORDER BY instrument_key ASC, trade_date ASC`, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]map[string]futuresQuoteSupplement{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	priority := r.buildStrategyFuturesSourcePriorityMap()
	result := make(map[string]map[string]futuresQuoteSupplement, len(instrumentKeys))
	for rows.Next() {
		var (
			instrumentKey string
			tradeDate     time.Time
			sourceKey     sql.NullString
			closePrice    sql.NullFloat64
			settle        sql.NullFloat64
			prevSettle    sql.NullFloat64
			turnover      sql.NullFloat64
			openInterest  sql.NullFloat64
		)
		if err := rows.Scan(&instrumentKey, &tradeDate, &sourceKey, &closePrice, &settle, &prevSettle, &turnover, &openInterest); err != nil {
			return nil, err
		}
		instrumentKey = strings.ToUpper(strings.TrimSpace(instrumentKey))
		if instrumentKey == "" {
			continue
		}
		tradeDateKey := tradeDate.Format("2006-01-02")
		candidate := futuresQuoteSupplement{
			SourceKey:       strings.ToUpper(strings.TrimSpace(sourceKey.String)),
			SettlePrice:     sqlNullFloat(settle),
			PrevSettlePrice: sqlNullFloat(prevSettle),
			Turnover:        sqlNullFloat(turnover),
			OpenInterest:    sqlNullFloat(openInterest),
		}
		if candidate.SettlePrice <= 0 && closePrice.Valid {
			candidate.SettlePrice = closePrice.Float64
		}
		if result[instrumentKey] == nil {
			result[instrumentKey] = make(map[string]futuresQuoteSupplement)
		}
		existing, ok := result[instrumentKey][tradeDateKey]
		if !ok || preferFuturesQuoteSupplement(candidate, existing, priority) {
			result[instrumentKey][tradeDateKey] = candidate
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func (r *MySQLGrowthRepo) buildStrategyFuturesSourcePriorityMap() map[string]int {
	priorityList := r.loadMarketSourcePriority(marketFuturesPriorityConfigKey, []string{"TUSHARE", "TICKERMD", "AKSHARE", "MYSELF", "MOCK"})
	priority := make(map[string]int, len(priorityList)+2)
	for index, sourceKey := range priorityList {
		priority[sourceKey] = index
	}
	if _, ok := priority["MYSELF"]; !ok {
		priority["MYSELF"] = len(priority)
	}
	if _, ok := priority["MOCK"]; !ok {
		priority["MOCK"] = len(priority) + 1
	}
	return priority
}

func preferFuturesQuoteSupplement(candidate futuresQuoteSupplement, existing futuresQuoteSupplement, priority map[string]int) bool {
	candidateMock := strings.EqualFold(candidate.SourceKey, "MOCK")
	existingMock := strings.EqualFold(existing.SourceKey, "MOCK")
	if candidateMock != existingMock {
		return !candidateMock
	}
	candidateScore := scoreFuturesQuoteSupplement(candidate)
	existingScore := scoreFuturesQuoteSupplement(existing)
	if candidateScore != existingScore {
		return candidateScore > existingScore
	}
	return futuresSourcePriorityRank(candidate.SourceKey, priority) < futuresSourcePriorityRank(existing.SourceKey, priority)
}

func scoreFuturesQuoteSupplement(item futuresQuoteSupplement) int {
	score := 0
	if item.SettlePrice > 0 {
		score += 2
	}
	if item.PrevSettlePrice > 0 {
		score += 2
	}
	if item.Turnover > 0 {
		score++
	}
	if item.OpenInterest > 0 {
		score++
	}
	return score
}

func futuresSourcePriorityRank(sourceKey string, priority map[string]int) int {
	normalized := strings.ToUpper(strings.TrimSpace(sourceKey))
	if rank, ok := priority[normalized]; ok {
		return rank
	}
	return len(priority) + 10
}

func mergeFuturesQuoteSupplements(history map[string][]futuresQuoteCandle, supplements map[string]map[string]futuresQuoteSupplement) {
	for instrumentKey, quotes := range history {
		byDate := supplements[instrumentKey]
		if len(byDate) == 0 {
			continue
		}
		for index := range quotes {
			tradeDateKey := quotes[index].TradeDate.Format("2006-01-02")
			supplement, ok := byDate[tradeDateKey]
			if !ok {
				continue
			}
			if (quotes[index].SettlePrice <= 0 || (nearlyEqualFloat(quotes[index].SettlePrice, quotes[index].ClosePrice) && !nearlyEqualFloat(supplement.SettlePrice, quotes[index].ClosePrice))) && supplement.SettlePrice > 0 {
				quotes[index].SettlePrice = supplement.SettlePrice
			}
			if quotes[index].PrevSettlePrice <= 0 && supplement.PrevSettlePrice > 0 {
				quotes[index].PrevSettlePrice = supplement.PrevSettlePrice
			}
			if quotes[index].Turnover <= 0 && supplement.Turnover > 0 {
				quotes[index].Turnover = supplement.Turnover
			}
			if quotes[index].OpenInterest <= 0 && supplement.OpenInterest > 0 {
				quotes[index].OpenInterest = supplement.OpenInterest
			}
		}
		history[instrumentKey] = quotes
	}
}

func (r *MySQLGrowthRepo) loadStrategyFuturesNewsSignals(aliasMap map[string][]string, selectedTradeDate time.Time, windowDays int) (map[string]futuresNewsSignal, error) {
	if len(aliasMap) == 0 {
		return map[string]futuresNewsSignal{}, nil
	}
	aliasToContract := make(map[string]string)
	for contract, aliases := range aliasMap {
		for _, alias := range aliases {
			normalized := strings.ToUpper(strings.TrimSpace(alias))
			if normalized == "" {
				continue
			}
			aliasToContract[normalized] = contract
			if canonical := normalizeFuturesContextContract(normalized); canonical != "" {
				aliasToContract[canonical] = contract
			}
		}
	}
	start := selectedTradeDate.AddDate(0, 0, -(windowDays - 1))
	end := selectedTradeDate.AddDate(0, 0, 1)
	rows, err := r.db.Query(`
SELECT primary_symbol, symbols_json, title
FROM market_news_items
WHERE published_at >= ? AND published_at < ?
ORDER BY published_at DESC`, start, end)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]futuresNewsSignal{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	heatMap := make(map[string]int, len(aliasMap))
	positiveMap := make(map[string]int, len(aliasMap))
	negativeMap := make(map[string]int, len(aliasMap))
	for rows.Next() {
		var primary sql.NullString
		var symbolsJSON sql.NullString
		var title string
		if err := rows.Scan(&primary, &symbolsJSON, &title); err != nil {
			return nil, err
		}
		matched := collectFuturesNewsContracts(primary, symbolsJSON, aliasToContract)
		if len(matched) == 0 {
			continue
		}
		sentiment := classifyNewsSentiment(title)
		for _, contract := range matched {
			heatMap[contract]++
			if sentiment == "POSITIVE" {
				positiveMap[contract]++
			}
			if sentiment == "NEGATIVE" {
				negativeMap[contract]++
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result := make(map[string]futuresNewsSignal, len(aliasMap))
	for contract := range aliasMap {
		heat := heatMap[contract]
		if heat <= 0 {
			continue
		}
		result[contract] = futuresNewsSignal{
			Heat: heat,
			Bias: clampFloat(float64(positiveMap[contract]-negativeMap[contract])/float64(heat), -1, 1),
		}
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadStrategyFuturesCurveMetrics(selectedTradeDate time.Time, candidates []strategyFuturesContextCandidate, allowMock bool) (map[string]futuresCurveMetrics, error) {
	if len(candidates) == 0 {
		return map[string]futuresCurveMetrics{}, nil
	}
	rootSet := make(map[string]struct{}, len(candidates))
	roots := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		root, _, ok := splitFuturesContractCode(candidate.Contract)
		if !ok {
			continue
		}
		if _, exists := rootSet[root]; exists {
			continue
		}
		rootSet[root] = struct{}{}
		roots = append(roots, root)
	}
	if len(roots) == 0 {
		return map[string]futuresCurveMetrics{}, nil
	}

	args := []any{marketAssetClassFutures, selectedTradeDate.Format("2006-01-02")}
	query := `
SELECT instrument_key, close_price
FROM market_daily_bar_truth
WHERE asset_class = ? AND trade_date = ?`
	if !allowMock {
		query += " AND UPPER(selected_source_key) <> ?"
		args = append(args, "MOCK")
	}
	query += " AND ("
	for index, root := range roots {
		if index > 0 {
			query += " OR "
		}
		query += "UPPER(SUBSTRING_INDEX(instrument_key, '.', 1)) LIKE ?"
		args = append(args, root+"%")
	}
	query += ") ORDER BY instrument_key ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	grouped := make(map[string][]futuresTermStructureQuote, len(roots))
	for rows.Next() {
		var instrumentKey string
		var closePrice float64
		if err := rows.Scan(&instrumentKey, &closePrice); err != nil {
			return nil, err
		}
		contract := normalizeFuturesContextContract(instrumentKey)
		root, expiryCode, ok := splitFuturesContractCode(contract)
		if !ok || closePrice <= 0 {
			continue
		}
		if _, exists := rootSet[root]; !exists {
			continue
		}
		grouped[root] = append(grouped[root], futuresTermStructureQuote{
			Contract:   contract,
			Root:       root,
			ExpiryCode: expiryCode,
			ClosePrice: closePrice,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make(map[string]futuresCurveMetrics, len(candidates))
	for _, items := range grouped {
		sort.SliceStable(items, func(i, j int) bool {
			if items[i].ExpiryCode == items[j].ExpiryCode {
				return items[i].Contract < items[j].Contract
			}
			return items[i].ExpiryCode < items[j].ExpiryCode
		})
		curveSlopePct := 0.0
		if len(items) >= 2 && items[0].ClosePrice > 0 && items[len(items)-1].ClosePrice > 0 {
			curveSlopePct = (items[len(items)-1].ClosePrice/items[0].ClosePrice - 1) * 100
		}
		for index, item := range items {
			if item.ClosePrice <= 0 {
				continue
			}
			termStructurePct := 0.0
			if index+1 < len(items) && items[index+1].ClosePrice > 0 {
				termStructurePct = (items[index+1].ClosePrice/item.ClosePrice - 1) * 100
			} else if index > 0 && items[index-1].ClosePrice > 0 {
				termStructurePct = (item.ClosePrice/items[index-1].ClosePrice - 1) * 100
			}
			result[item.Contract] = futuresCurveMetrics{
				TermStructurePct: roundTo(termStructurePct, 4),
				CurveSlopePct:    roundTo(curveSlopePct, 4),
			}
		}
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadStrategyFuturesInventorySignals(candidates []strategyFuturesContextCandidate, selectedTradeDate time.Time) (map[string]futuresInventorySignal, error) {
	if len(candidates) == 0 {
		return map[string]futuresInventorySignal{}, nil
	}
	rootToContracts := make(map[string][]string, len(candidates))
	roots := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		root, _, ok := splitFuturesContractCode(candidate.Contract)
		if !ok {
			continue
		}
		if _, exists := rootToContracts[root]; !exists {
			roots = append(roots, root)
		}
		rootToContracts[root] = append(rootToContracts[root], candidate.Contract)
	}
	if len(roots) == 0 {
		return map[string]futuresInventorySignal{}, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(roots)), ",")
	args := make([]any, 0, len(roots)+1)
	args = append(args, selectedTradeDate.Format("2006-01-02"))
	for _, root := range roots {
		args = append(args, root)
	}
	query := fmt.Sprintf(`
SELECT symbol, trade_date, warehouse, area, brand, place, grade, receipt_volume, previous_volume, change_volume
FROM futures_inventory_snapshots
WHERE trade_date <= ? AND symbol IN (%s)
ORDER BY symbol ASC, trade_date ASC`, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]futuresInventorySignal{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	type inventoryRow struct {
		TradeDate      time.Time
		Warehouse      string
		Area           string
		Brand          string
		Place          string
		Grade          string
		ReceiptVolume  float64
		PreviousVolume float64
		ChangeVolume   float64
	}
	grouped := make(map[string][]inventoryRow, len(roots))
	for rows.Next() {
		var symbol string
		var tradeDate time.Time
		var warehouse sql.NullString
		var area sql.NullString
		var brand sql.NullString
		var place sql.NullString
		var grade sql.NullString
		var receiptVolume sql.NullFloat64
		var previousVolume sql.NullFloat64
		var changeVolume sql.NullFloat64
		if err := rows.Scan(&symbol, &tradeDate, &warehouse, &area, &brand, &place, &grade, &receiptVolume, &previousVolume, &changeVolume); err != nil {
			return nil, err
		}
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" {
			continue
		}
		grouped[symbol] = append(grouped[symbol], inventoryRow{
			TradeDate:      tradeDate,
			Warehouse:      strings.TrimSpace(warehouse.String),
			Area:           strings.TrimSpace(area.String),
			Brand:          strings.TrimSpace(brand.String),
			Place:          strings.TrimSpace(place.String),
			Grade:          strings.TrimSpace(grade.String),
			ReceiptVolume:  sqlNullFloat(receiptVolume),
			PreviousVolume: sqlNullFloat(previousVolume),
			ChangeVolume:   sqlNullFloat(changeVolume),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	result := make(map[string]futuresInventorySignal, len(candidates))
	for root, items := range grouped {
		if len(items) == 0 {
			continue
		}
		latestTradeDate := items[len(items)-1].TradeDate
		level := 0.0
		previous := 0.0
		latestRows := make([]inventoryRow, 0, 4)
		for _, item := range items {
			if !sameDay(item.TradeDate, latestTradeDate) {
				continue
			}
			level += item.ReceiptVolume
			previous += item.PreviousVolume
			latestRows = append(latestRows, item)
		}
		if previous <= 0 && len(items) >= len(latestRows)+1 {
			previousTradeDate := time.Time{}
			fallbackPrevious := 0.0
			for idx := len(items) - len(latestRows) - 1; idx >= 0; idx-- {
				if previousTradeDate.IsZero() {
					previousTradeDate = items[idx].TradeDate
				}
				if !sameDay(items[idx].TradeDate, previousTradeDate) {
					break
				}
				fallbackPrevious += items[idx].ReceiptVolume
			}
			previous = fallbackPrevious
		}
		if previous <= 0 {
			for _, item := range latestRows {
				if item.ChangeVolume != 0 {
					previous += item.ReceiptVolume - item.ChangeVolume
				}
			}
		}
		changePct := 0.0
		if previous > 0 && level >= 0 {
			changePct = (level/previous - 1) * 100
		}
		areaName, areaShare := dominantInventoryBucket(latestRows, func(item inventoryRow) string { return item.Area }, func(item inventoryRow) float64 { return item.ReceiptVolume }, level)
		warehouseName, warehouseShare := dominantInventoryBucket(latestRows, func(item inventoryRow) string { return item.Warehouse }, func(item inventoryRow) float64 { return item.ReceiptVolume }, level)
		brandName, brandShare := dominantInventoryBucket(latestRows, func(item inventoryRow) string { return item.Brand }, func(item inventoryRow) float64 { return item.ReceiptVolume }, level)
		placeName, placeShare := dominantInventoryBucket(latestRows, func(item inventoryRow) string { return item.Place }, func(item inventoryRow) float64 { return item.ReceiptVolume }, level)
		gradeName, gradeShare := dominantInventoryBucket(latestRows, func(item inventoryRow) string { return item.Grade }, func(item inventoryRow) float64 { return item.ReceiptVolume }, level)
		pressure := clampFloat(-changePct/20, -1, 1)
		structureShare := math.Max(areaShare, math.Max(warehouseShare, math.Max(brandShare, math.Max(placeShare, gradeShare))))
		if structureShare > 0 && pressure != 0 {
			pressure = clampFloat(pressure*(1+structureShare*0.35), -1, 1)
		}
		signal := futuresInventorySignal{
			Level:          roundTo(level, 4),
			ChangePct:      roundTo(changePct, 4),
			Pressure:       roundTo(pressure, 4),
			FocusArea:      areaName,
			FocusWarehouse: warehouseName,
			FocusBrand:     brandName,
			FocusPlace:     placeName,
			FocusGrade:     gradeName,
			AreaShare:      roundTo(areaShare, 4),
			WarehouseShare: roundTo(warehouseShare, 4),
			BrandShare:     roundTo(brandShare, 4),
			PlaceShare:     roundTo(placeShare, 4),
			GradeShare:     roundTo(gradeShare, 4),
		}
		for _, contract := range rootToContracts[root] {
			result[contract] = signal
		}
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadStrategyFuturesSpreadSignals(candidates []strategyFuturesContextCandidate) (map[string]futuresSpreadSignal, error) {
	if len(candidates) == 0 {
		return map[string]futuresSpreadSignal{}, nil
	}
	contractSet := make(map[string]struct{}, len(candidates))
	contracts := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		contract := normalizeFuturesContextContract(candidate.Contract)
		if contract == "" {
			continue
		}
		if _, exists := contractSet[contract]; exists {
			continue
		}
		contractSet[contract] = struct{}{}
		contracts = append(contracts, contract)
	}
	if len(contracts) == 0 {
		return map[string]futuresSpreadSignal{}, nil
	}

	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(contracts)), ",")
	args := make([]any, 0, len(contracts)*2)
	for _, contract := range contracts {
		args = append(args, contract)
	}
	for _, contract := range contracts {
		args = append(args, contract)
	}
	query := fmt.Sprintf(`
SELECT contract_a, contract_b, percentile, status
FROM arbitrage_recos
WHERE UPPER(contract_a) IN (%s) OR UPPER(contract_b) IN (%s)
ORDER BY CASE UPPER(status)
  WHEN 'OPEN' THEN 0
  WHEN 'WATCH' THEN 1
  WHEN 'PUBLISHED' THEN 2
  ELSE 3
END ASC,
ABS(COALESCE(percentile, 0.5) - 0.5) DESC`, placeholders, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]futuresSpreadSignal{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]futuresSpreadSignal, len(contracts))
	for rows.Next() {
		var contractA sql.NullString
		var contractB sql.NullString
		var percentile sql.NullFloat64
		var status sql.NullString
		if err := rows.Scan(&contractA, &contractB, &percentile, &status); err != nil {
			return nil, err
		}
		_ = status
		a := normalizeFuturesContextContract(contractA.String)
		b := normalizeFuturesContextContract(contractB.String)
		if a == "" || b == "" || a == b {
			continue
		}
		value := 0.5
		if percentile.Valid {
			value = clampFloat(percentile.Float64, 0, 1)
		}
		pair := a + "/" + b
		assignFuturesSpreadSignal(result, a, futuresSpreadSignal{
			Pair:       pair,
			Percentile: roundTo(value, 4),
			Pressure:   roundTo(clampFloat(0.5-value, -0.5, 0.5), 4),
		})
		assignFuturesSpreadSignal(result, b, futuresSpreadSignal{
			Pair:       pair,
			Percentile: roundTo(value, 4),
			Pressure:   roundTo(clampFloat(value-0.5, -0.5, 0.5), 4),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func dominantInventoryBucket[T any](items []T, keyFn func(T) string, volumeFn func(T) float64, total float64) (string, float64) {
	if len(items) == 0 || total <= 0 {
		return "", 0
	}
	totals := make(map[string]float64)
	for _, item := range items {
		key := strings.TrimSpace(keyFn(item))
		if key == "" {
			continue
		}
		totals[key] += volumeFn(item)
	}
	bestKey := ""
	bestValue := 0.0
	for key, value := range totals {
		if value > bestValue || (nearlyEqualFloat(value, bestValue) && (bestKey == "" || key < bestKey)) {
			bestKey = key
			bestValue = value
		}
	}
	if bestKey == "" || bestValue <= 0 {
		return "", 0
	}
	return bestKey, clampFloat(bestValue/total, 0, 1)
}

func sameDay(left time.Time, right time.Time) bool {
	return left.Format("2006-01-02") == right.Format("2006-01-02")
}

func assignFuturesSpreadSignal(target map[string]futuresSpreadSignal, contract string, candidate futuresSpreadSignal) {
	if contract == "" || candidate.Pair == "" {
		return
	}
	existing, ok := target[contract]
	if !ok || math.Abs(candidate.Pressure) > math.Abs(existing.Pressure) {
		target[contract] = candidate
	}
}

func (r *MySQLGrowthRepo) loadStrategyStockTruthHistory(symbols []string, selectedTradeDate time.Time) (map[string][]stockQuoteCandle, error) {
	if len(symbols) == 0 {
		return map[string][]stockQuoteCandle{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	args := make([]any, 0, len(symbols)+3)
	args = append(args, marketAssetClassStock)
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	args = append(args, selectedTradeDate.AddDate(0, 0, -60).Format("2006-01-02"), selectedTradeDate.Format("2006-01-02"))
	query := fmt.Sprintf(`
SELECT instrument_key, trade_date, open_price, high_price, low_price, close_price, prev_close_price, volume, turnover
FROM market_daily_bar_truth
WHERE asset_class = ?
  AND instrument_key IN (%s)
  AND trade_date >= ?
  AND trade_date <= ?
ORDER BY instrument_key ASC, trade_date ASC`, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]stockQuoteCandle, len(symbols))
	for rows.Next() {
		var item stockQuoteCandle
		var prevClose sql.NullFloat64
		var turnover sql.NullFloat64
		if err := rows.Scan(
			&item.Symbol,
			&item.TradeDate,
			&item.OpenPrice,
			&item.HighPrice,
			&item.LowPrice,
			&item.ClosePrice,
			&prevClose,
			&item.Volume,
			&turnover,
		); err != nil {
			return nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
		if prevClose.Valid {
			item.PrevClosePrice = prevClose.Float64
		}
		if turnover.Valid {
			item.Turnover = turnover.Float64
		}
		if item.Symbol == "" || item.ClosePrice <= 0 {
			continue
		}
		result[item.Symbol] = append(result[item.Symbol], item)
	}
	return result, rows.Err()
}

func (r *MySQLGrowthRepo) loadStrategyStockDailyBasicsAsOf(symbols []string, selectedTradeDate time.Time) (map[string]stockDailyBasicPoint, error) {
	if len(symbols) == 0 {
		return map[string]stockDailyBasicPoint{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	args := make([]any, 0, len(symbols)*2+1)
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	args = append(args, selectedTradeDate.Format("2006-01-02"))
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	query := fmt.Sprintf(`
SELECT t.symbol, t.trade_date, t.turnover_rate, t.volume_ratio, t.pe_ttm, t.pb, t.total_mv, t.circ_mv, t.source_key
FROM stock_daily_basic t
INNER JOIN (
  SELECT symbol, MAX(trade_date) AS latest_trade_date
  FROM stock_daily_basic
  WHERE symbol IN (%s) AND trade_date <= ?
  GROUP BY symbol
) latest
ON latest.symbol = t.symbol AND latest.latest_trade_date = t.trade_date
WHERE t.symbol IN (%s)`, placeholders, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]stockDailyBasicPoint{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]stockDailyBasicPoint, len(symbols))
	for rows.Next() {
		var (
			item         stockDailyBasicPoint
			turnoverRate sql.NullFloat64
			volumeRatio  sql.NullFloat64
			peTTM        sql.NullFloat64
			pb           sql.NullFloat64
			totalMV      sql.NullFloat64
			circMV       sql.NullFloat64
			sourceKey    sql.NullString
		)
		if err := rows.Scan(&item.Symbol, &item.TradeDate, &turnoverRate, &volumeRatio, &peTTM, &pb, &totalMV, &circMV, &sourceKey); err != nil {
			return nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
		item.TurnoverRate = sqlNullFloat(turnoverRate)
		item.VolumeRatio = sqlNullFloat(volumeRatio)
		item.PeTTM = sqlNullFloat(peTTM)
		item.PB = sqlNullFloat(pb)
		item.TotalMV = sqlNullFloat(totalMV)
		item.CircMV = sqlNullFloat(circMV)
		if sourceKey.Valid {
			item.SourceKey = strings.ToUpper(strings.TrimSpace(sourceKey.String))
		}
		if item.Symbol != "" {
			result[item.Symbol] = item
		}
	}
	return result, rows.Err()
}

func (r *MySQLGrowthRepo) loadStrategyStockMoneyflowsAsOf(symbols []string, selectedTradeDate time.Time) (map[string]stockMoneyflowPoint, error) {
	if len(symbols) == 0 {
		return map[string]stockMoneyflowPoint{}, nil
	}
	placeholders := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	args := make([]any, 0, len(symbols)*2+1)
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	args = append(args, selectedTradeDate.Format("2006-01-02"))
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	query := fmt.Sprintf(`
SELECT t.symbol, t.trade_date, t.net_mf_amount, t.buy_lg_amount, t.sell_lg_amount, t.buy_elg_amount, t.sell_elg_amount, t.source_key
FROM stock_moneyflow_daily t
INNER JOIN (
  SELECT symbol, MAX(trade_date) AS latest_trade_date
  FROM stock_moneyflow_daily
  WHERE symbol IN (%s) AND trade_date <= ?
  GROUP BY symbol
) latest
ON latest.symbol = t.symbol AND latest.latest_trade_date = t.trade_date
WHERE t.symbol IN (%s)`, placeholders, placeholders)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]stockMoneyflowPoint{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]stockMoneyflowPoint, len(symbols))
	for rows.Next() {
		var (
			item      stockMoneyflowPoint
			netMF     sql.NullFloat64
			buyLG     sql.NullFloat64
			sellLG    sql.NullFloat64
			buyELG    sql.NullFloat64
			sellELG   sql.NullFloat64
			sourceKey sql.NullString
		)
		if err := rows.Scan(&item.Symbol, &item.TradeDate, &netMF, &buyLG, &sellLG, &buyELG, &sellELG, &sourceKey); err != nil {
			return nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
		item.NetMFAmount = sqlNullFloat(netMF)
		item.BuyLGAmount = sqlNullFloat(buyLG)
		item.SellLGAmount = sqlNullFloat(sellLG)
		item.BuyELGAmount = sqlNullFloat(buyELG)
		item.SellELGAmount = sqlNullFloat(sellELG)
		if sourceKey.Valid {
			item.SourceKey = strings.ToUpper(strings.TrimSpace(sourceKey.String))
		}
		if item.Symbol != "" {
			result[item.Symbol] = item
		}
	}
	return result, rows.Err()
}

func (r *MySQLGrowthRepo) loadStrategyMarketNewsSignals(symbols []string, selectedTradeDate time.Time, windowDays int) (map[string]stockNewsSignal, error) {
	if len(symbols) == 0 {
		return map[string]stockNewsSignal{}, nil
	}
	start := selectedTradeDate.AddDate(0, 0, -(windowDays - 1))
	end := selectedTradeDate.AddDate(0, 0, 1)
	rows, err := r.db.Query(`
SELECT primary_symbol, symbols_json, title
FROM market_news_items
WHERE published_at >= ? AND published_at < ?
ORDER BY published_at DESC`, start, end)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]stockNewsSignal{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	targetSymbols := symbolSet(symbols)
	heatMap := make(map[string]int, len(symbols))
	positiveMap := make(map[string]int, len(symbols))
	for rows.Next() {
		var primary sql.NullString
		var symbolsJSON sql.NullString
		var title string
		if err := rows.Scan(&primary, &symbolsJSON, &title); err != nil {
			return nil, err
		}
		matched := collectNewsSymbols(primary, symbolsJSON, targetSymbols)
		if len(matched) == 0 {
			continue
		}
		isPositive := classifyNewsSentiment(title) == "POSITIVE"
		for _, symbol := range matched {
			heatMap[symbol]++
			if isPositive {
				positiveMap[symbol]++
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	result := make(map[string]stockNewsSignal, len(symbols))
	for _, symbol := range symbols {
		heat := heatMap[symbol]
		if heat <= 0 {
			continue
		}
		result[symbol] = stockNewsSignal{Heat: heat, PositiveRate: float64(positiveMap[symbol]) / float64(heat)}
	}
	return result, nil
}

func parseStrategyContextTradeDate(raw string) (time.Time, error) {
	if strings.TrimSpace(raw) == "" {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local), nil
	}
	parsed, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(raw), time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid trade_date: %w", err)
	}
	return parsed, nil
}

func symbolSet(symbols []string) map[string]struct{} {
	result := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		result[normalized] = struct{}{}
	}
	return result
}

func appendUniqueText(items []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return items
	}
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func missingCandidateSymbols(expected []string, candidates []strategyStockContextCandidate) []string {
	seen := make(map[string]struct{}, len(candidates))
	for _, item := range candidates {
		seen[item.Symbol] = struct{}{}
	}
	missing := make([]string, 0)
	for _, symbol := range expected {
		if _, ok := seen[symbol]; !ok {
			missing = append(missing, symbol)
		}
	}
	return missing
}

func summarizeStrategyPriceSource(sourceSet map[string]struct{}) string {
	if len(sourceSet) == 0 {
		return ""
	}
	if len(sourceSet) == 1 {
		for source := range sourceSet {
			return source
		}
	}
	return "MIXED"
}

func noNewsSignals(signals map[string]stockNewsSignal, seeds []model.StrategyEngineStockSeed) bool {
	for _, seed := range seeds {
		if signal, ok := signals[seed.Symbol]; ok && signal.Heat > 0 {
			return false
		}
	}
	return true
}

func collectNewsSymbols(primary sql.NullString, symbolsJSON sql.NullString, targetSymbols map[string]struct{}) []string {
	matched := make([]string, 0)
	seen := make(map[string]struct{})
	push := func(symbol string) {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			return
		}
		if _, ok := targetSymbols[normalized]; !ok {
			return
		}
		if _, ok := seen[normalized]; ok {
			return
		}
		seen[normalized] = struct{}{}
		matched = append(matched, normalized)
	}
	if primary.Valid {
		push(primary.String)
	}
	if !symbolsJSON.Valid || strings.TrimSpace(symbolsJSON.String) == "" {
		return matched
	}
	var rawSymbols []string
	if err := json.Unmarshal([]byte(symbolsJSON.String), &rawSymbols); err == nil {
		for _, symbol := range rawSymbols {
			push(symbol)
		}
		return matched
	}
	for _, item := range strings.FieldsFunc(symbolsJSON.String, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '\n' || r == '\t'
	}) {
		push(item)
	}
	return matched
}

func normalizeFuturesContextContractList(contracts []string) []string {
	seen := make(map[string]struct{}, len(contracts))
	items := make([]string, 0, len(contracts))
	for _, contract := range contracts {
		normalized := normalizeFuturesContextContract(contract)
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

func normalizeFuturesContextContract(value string) string {
	normalized := strings.ToUpper(strings.TrimSpace(value))
	if normalized == "" {
		return ""
	}
	if idx := strings.Index(normalized, "."); idx > 0 {
		return normalized[:idx]
	}
	return normalized
}

func missingFuturesCandidateContracts(expected []string, candidates []strategyFuturesContextCandidate) []string {
	seen := make(map[string]struct{}, len(candidates))
	for _, item := range candidates {
		seen[item.Contract] = struct{}{}
	}
	missing := make([]string, 0)
	for _, contract := range expected {
		if _, ok := seen[contract]; !ok {
			missing = append(missing, contract)
		}
	}
	return missing
}

func buildFuturesContextSeed(candidate strategyFuturesContextCandidate, quotes []futuresQuoteCandle, newsSignal futuresNewsSignal, curveMetrics futuresCurveMetrics, inventorySignal futuresInventorySignal, spreadSignal futuresSpreadSignal) (model.StrategyEngineFuturesSeed, bool) {
	if len(quotes) < 15 {
		return model.StrategyEngineFuturesSeed{}, false
	}
	lastIndex := len(quotes) - 1
	latest := quotes[lastIndex]
	if latest.ClosePrice <= 0 {
		return model.StrategyEngineFuturesSeed{}, false
	}

	returnStart := maxInt(1, lastIndex-13)
	returns := make([]float64, 0, lastIndex-returnStart+1)
	for idx := returnStart; idx <= lastIndex; idx++ {
		prev := quotes[idx-1].ClosePrice
		curr := quotes[idx].ClosePrice
		if prev <= 0 || curr <= 0 {
			continue
		}
		returns = append(returns, curr/prev-1)
	}
	if len(returns) < 8 {
		return model.StrategyEngineFuturesSeed{}, false
	}
	volatility14 := calculateStdDev(returns) * 100

	volumeStart := maxInt(0, lastIndex-14)
	volumeSum := 0.0
	volumeCount := 0
	turnoverSum := 0.0
	turnoverCount := 0
	for idx := volumeStart; idx < lastIndex; idx++ {
		if quotes[idx].Volume > 0 {
			volumeSum += quotes[idx].Volume
			volumeCount++
		}
		if quotes[idx].Turnover > 0 {
			turnoverSum += quotes[idx].Turnover
			turnoverCount++
		}
	}
	avgVolume := volumeSum
	if volumeCount > 0 {
		avgVolume = volumeSum / float64(volumeCount)
	}
	volumeRatio := 1.0
	if avgVolume > 0 && latest.Volume > 0 {
		volumeRatio = latest.Volume / avgVolume
	}
	avgTurnover := turnoverSum
	if turnoverCount > 0 {
		avgTurnover = turnoverSum / float64(turnoverCount)
	}
	turnoverRatio := 1.0
	if avgTurnover > 0 && latest.Turnover > 0 {
		turnoverRatio = latest.Turnover / avgTurnover
	}

	ma5, ok := avgFuturesClose(quotes, 5)
	if !ok {
		return model.StrategyEngineFuturesSeed{}, false
	}
	ma14, ok := avgFuturesClose(quotes, 14)
	if !ok {
		return model.StrategyEngineFuturesSeed{}, false
	}
	trendStrength := 0.0
	if ma14 > 0 {
		trendStrength = (ma5/ma14 - 1) * 100
	}

	oiChangePct := 0.0
	if lastIndex > 0 && latest.OpenInterest > 0 {
		prevOI := quotes[lastIndex-1].OpenInterest
		if prevOI > 0 {
			oiChangePct = (latest.OpenInterest/prevOI - 1) * 100
		}
	}

	basisPct := 0.0
	if latest.SettlePrice > 0 && !nearlyEqualFloat(latest.SettlePrice, latest.ClosePrice) {
		basisPct = (latest.ClosePrice/latest.SettlePrice - 1) * 100
	}
	carryPct := 0.0
	if latest.SettlePrice > 0 && latest.PrevSettlePrice > 0 {
		carryPct = (latest.SettlePrice/latest.PrevSettlePrice - 1) * 100
	}

	priceChangePct := 0.0
	if latest.PrevClosePrice > 0 {
		priceChangePct = (latest.ClosePrice/latest.PrevClosePrice - 1) * 100
	}
	structureSupport := 0.0
	if carryPct != 0 && curveMetrics.TermStructurePct != 0 {
		if hasSameDirection(carryPct, curveMetrics.TermStructurePct) {
			structureSupport = math.Min(math.Abs(curveMetrics.TermStructurePct), 3) * 0.05
		} else {
			structureSupport = -math.Min(math.Abs(curveMetrics.TermStructurePct), 3) * 0.03
		}
	}
	curveSupport := 0.0
	if carryPct != 0 && curveMetrics.CurveSlopePct != 0 {
		if hasSameDirection(carryPct, curveMetrics.CurveSlopePct) {
			curveSupport = math.Min(math.Abs(curveMetrics.CurveSlopePct), 5) * 0.03
		} else {
			curveSupport = -math.Min(math.Abs(curveMetrics.CurveSlopePct), 5) * 0.02
		}
	}
	inventorySupport := inventorySignal.Pressure * 0.3
	spreadSupport := spreadSignal.Pressure * 0.4
	flowBias := clampFloat(priceChangePct*0.1+oiChangePct*0.08+(volumeRatio-1)*0.35+(turnoverRatio-1)*0.35+carryPct*0.04+structureSupport+curveSupport+inventorySupport+spreadSupport, -1, 1)
	regime := classifyFuturesRegime(trendStrength, volatility14, flowBias, carryPct, newsSignal.Bias, turnoverRatio, curveMetrics.TermStructurePct, curveMetrics.CurveSlopePct, inventorySignal.Pressure, spreadSignal.Pressure)

	return model.StrategyEngineFuturesSeed{
		Contract:                candidate.Contract,
		Name:                    candidate.Name,
		TradeDate:               latest.TradeDate.Format("2006-01-02"),
		LastPrice:               roundTo(latest.ClosePrice, 4),
		BasisPct:                roundTo(basisPct, 4),
		Volatility14:            roundTo(volatility14, 4),
		TrendStrength:           roundTo(trendStrength, 4),
		OIChangePct:             roundTo(oiChangePct, 4),
		VolumeRatio:             roundTo(volumeRatio, 4),
		TurnoverRatio:           roundTo(turnoverRatio, 4),
		FlowBias:                roundTo(flowBias, 4),
		CarryPct:                roundTo(carryPct, 4),
		TermStructurePct:        roundTo(curveMetrics.TermStructurePct, 4),
		CurveSlopePct:           roundTo(curveMetrics.CurveSlopePct, 4),
		InventoryLevel:          roundTo(inventorySignal.Level, 4),
		InventoryChangePct:      roundTo(inventorySignal.ChangePct, 4),
		InventoryPressure:       roundTo(inventorySignal.Pressure, 4),
		InventoryFocusArea:      inventorySignal.FocusArea,
		InventoryFocusWarehouse: inventorySignal.FocusWarehouse,
		InventoryFocusBrand:     inventorySignal.FocusBrand,
		InventoryFocusPlace:     inventorySignal.FocusPlace,
		InventoryFocusGrade:     inventorySignal.FocusGrade,
		InventoryAreaShare:      roundTo(inventorySignal.AreaShare, 4),
		InventoryWarehouseShare: roundTo(inventorySignal.WarehouseShare, 4),
		InventoryBrandShare:     roundTo(inventorySignal.BrandShare, 4),
		InventoryPlaceShare:     roundTo(inventorySignal.PlaceShare, 4),
		InventoryGradeShare:     roundTo(inventorySignal.GradeShare, 4),
		SpreadPressure:          roundTo(spreadSignal.Pressure, 4),
		SpreadPercentile:        roundTo(spreadSignal.Percentile, 4),
		SpreadPair:              spreadSignal.Pair,
		NewsBias:                roundTo(newsSignal.Bias, 4),
		Regime:                  regime,
	}, true
}

func avgFuturesClose(quotes []futuresQuoteCandle, window int) (float64, bool) {
	if window <= 0 || len(quotes) < window {
		return 0, false
	}
	sum := 0.0
	for idx := len(quotes) - window; idx < len(quotes); idx++ {
		if quotes[idx].ClosePrice <= 0 {
			return 0, false
		}
		sum += quotes[idx].ClosePrice
	}
	return sum / float64(window), true
}

func classifyFuturesRegime(trendStrength float64, volatility14 float64, flowBias float64, carryPct float64, newsBias float64, turnoverRatio float64, termStructurePct float64, curveSlopePct float64, inventoryPressure float64, spreadPressure float64) string {
	if volatility14 >= 3.5 {
		return "VOLATILE"
	}
	if trendStrength <= -0.35 || flowBias <= -0.2 || spreadPressure <= -0.18 || inventoryPressure <= -0.2 {
		return "WEAK"
	}
	if trendStrength >= 0.45 && flowBias >= 0 && turnoverRatio >= 0.9 {
		return "TREND"
	}
	if carryPct > 0 && termStructurePct > 0 && curveSlopePct >= 0 {
		return "DEFENSIVE"
	}
	if inventoryPressure >= 0.18 && carryPct >= 0 {
		return "DEFENSIVE"
	}
	if spreadPressure >= 0.18 && curveSlopePct >= 0 {
		return "DEFENSIVE"
	}
	if carryPct > 0 || newsBias > 0.2 {
		return "DEFENSIVE"
	}
	if flowBias >= 0 && turnoverRatio >= 1 {
		return "TREND"
	}
	return "DEFENSIVE"
}

func nearlyEqualFloat(left float64, right float64) bool {
	return math.Abs(left-right) < 1e-9
}

func hasSameDirection(left float64, right float64) bool {
	return (left > 0 && right > 0) || (left < 0 && right < 0)
}

func splitFuturesContractCode(contract string) (string, int, bool) {
	normalized := normalizeFuturesContextContract(contract)
	if normalized == "" {
		return "", 0, false
	}
	rootBuilder := strings.Builder{}
	expiryCode := 0
	hasDigits := false
	for _, ch := range normalized {
		switch {
		case ch >= 'A' && ch <= 'Z' && !hasDigits:
			rootBuilder.WriteRune(ch)
		case ch >= '0' && ch <= '9':
			hasDigits = true
			expiryCode = expiryCode*10 + int(ch-'0')
		default:
			return "", 0, false
		}
	}
	if rootBuilder.Len() == 0 || !hasDigits {
		return "", 0, false
	}
	return rootBuilder.String(), expiryCode, true
}

func noFuturesNewsSignals(signals map[string]futuresNewsSignal, seeds []model.StrategyEngineFuturesSeed) bool {
	for _, seed := range seeds {
		if signal, ok := signals[seed.Contract]; ok && signal.Heat > 0 {
			return false
		}
	}
	return true
}

func collectFuturesNewsContracts(primary sql.NullString, symbolsJSON sql.NullString, aliasToContract map[string]string) []string {
	matched := make([]string, 0)
	seen := make(map[string]struct{})
	push := func(symbol string) {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			return
		}
		contract, ok := aliasToContract[normalized]
		if !ok {
			contract, ok = aliasToContract[normalizeFuturesContextContract(normalized)]
		}
		if !ok || contract == "" {
			return
		}
		if _, ok := seen[contract]; ok {
			return
		}
		seen[contract] = struct{}{}
		matched = append(matched, contract)
	}
	if primary.Valid {
		push(primary.String)
	}
	if !symbolsJSON.Valid || strings.TrimSpace(symbolsJSON.String) == "" {
		return matched
	}
	var rawSymbols []string
	if err := json.Unmarshal([]byte(symbolsJSON.String), &rawSymbols); err == nil {
		for _, symbol := range rawSymbols {
			push(symbol)
		}
		return matched
	}
	for _, item := range strings.FieldsFunc(symbolsJSON.String, func(r rune) bool {
		return r == ',' || r == ';' || r == '|' || r == '\n' || r == '\t'
	}) {
		push(item)
	}
	return matched
}
