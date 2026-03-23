package repo

import (
	"database/sql"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *InMemoryGrowthRepo) AdminGetStockSelectionOverview() (model.AdminStockSelectionOverview, error) {
	profile := model.StockSelectionProfile{
		ID:                   model.StrategyEngineDefaultStockSelectionProfileID,
		Name:                 "默认自动选股",
		Status:               "ACTIVE",
		IsDefault:            true,
		CurrentVersion:       1,
		SelectionModeDefault: "AUTO",
		UniverseScope:        model.StrategyEngineDefaultStockUniverseScope,
		UniverseConfig:       map[string]any{"min_listing_days": 180, "min_avg_turnover": 50000000},
		PortfolioConfig:      map[string]any{"limit": 5, "max_risk_level": "MEDIUM", "min_score": 75},
		UpdatedBy:            "system",
		UpdatedAt:            time.Now().UTC().Format(time.RFC3339),
		CreatedAt:            time.Now().UTC().Format(time.RFC3339),
	}
	run := model.StockSelectionRun{
		RunID:          "ssr_demo_001",
		TradeDate:      time.Now().Format("2006-01-02"),
		ProfileID:      profile.ID,
		ProfileVersion: 1,
		SelectionMode:  "AUTO",
		Status:         "SUCCEEDED",
		UniverseCount:  10,
		SeedCount:      8,
		CandidateCount: 5,
		SelectedCount:  3,
		StageCounts:    map[string]int{"UNIVERSE": 10, "SEED_POOL": 8, "CANDIDATE_POOL": 5, "PORTFOLIO": 3},
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	return model.AdminStockSelectionOverview{
		DefaultProfile:   &profile,
		LatestTradeDate:  run.TradeDate,
		LatestRun:        &run,
		LatestSuccessRun: &run,
		DataFreshness:    map[string]any{"source": "in-memory"},
		EvaluationSummary: map[string]any{
			"5":  map[string]any{"win_rate": 0.6},
			"10": map[string]any{"win_rate": 0.62},
			"20": map[string]any{"win_rate": 0.64},
		},
		EvaluationSummaryV2: map[string]any{
			"1": map[string]any{
				"horizon_day":            1,
				"sample_count":           12,
				"avg_return_pct":         0.004,
				"avg_excess_return_pct":  0.001,
				"hit_rate":               0.58,
				"avg_max_drawdown_pct":   -0.008,
				"worst_max_drawdown_pct": -0.019,
				"generated_at":           time.Now().UTC().Format(time.RFC3339),
			},
			"3": map[string]any{
				"horizon_day":            3,
				"sample_count":           12,
				"avg_return_pct":         0.011,
				"avg_excess_return_pct":  0.004,
				"hit_rate":               0.61,
				"avg_max_drawdown_pct":   -0.014,
				"worst_max_drawdown_pct": -0.027,
				"generated_at":           time.Now().UTC().Format(time.RFC3339),
			},
			"5": map[string]any{
				"horizon_day":            5,
				"sample_count":           12,
				"avg_return_pct":         0.023,
				"avg_excess_return_pct":  0.009,
				"hit_rate":               0.66,
				"avg_max_drawdown_pct":   -0.021,
				"worst_max_drawdown_pct": -0.031,
				"generated_at":           time.Now().UTC().Format(time.RFC3339),
			},
			"10": map[string]any{
				"horizon_day":            10,
				"sample_count":           12,
				"avg_return_pct":         0.041,
				"avg_excess_return_pct":  0.016,
				"hit_rate":               0.69,
				"avg_max_drawdown_pct":   -0.028,
				"worst_max_drawdown_pct": -0.039,
				"generated_at":           time.Now().UTC().Format(time.RFC3339),
			},
			"20": map[string]any{
				"horizon_day":            20,
				"sample_count":           12,
				"avg_return_pct":         0.068,
				"avg_excess_return_pct":  0.025,
				"hit_rate":               0.72,
				"avg_max_drawdown_pct":   -0.035,
				"worst_max_drawdown_pct": -0.048,
				"generated_at":           time.Now().UTC().Format(time.RFC3339),
			},
			"leaderboard_items": []model.StockSelectionEvaluationLeaderboardItem{
				{
					TemplateID:       "sstpl_balanced_steady",
					TemplateName:     "均衡稳健",
					ProfileID:        model.StrategyEngineDefaultStockSelectionProfileID,
					ProfileName:      "默认自动选股",
					MarketRegime:     "ROTATION",
					SampleCount:      12,
					ReturnByHorizon:  map[string]float64{"1": 0.004, "3": 0.011, "5": 0.023, "10": 0.041, "20": 0.068},
					HitRateByHorizon: map[string]float64{"1": 0.58, "3": 0.61, "5": 0.66, "10": 0.69, "20": 0.72},
					MaxDrawdownPct:   -0.031,
				},
			},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListStockSelectionRuns(status string, reviewStatus string, profileID string, page int, pageSize int) ([]model.StockSelectionRun, int, error) {
	item := model.StockSelectionRun{
		RunID:          "ssr_demo_001",
		TradeDate:      time.Now().Format("2006-01-02"),
		ProfileID:      model.StrategyEngineDefaultStockSelectionProfileID,
		ProfileVersion: 1,
		TemplateID:     "sstpl_balanced_steady",
		TemplateName:   "均衡稳健",
		MarketRegime:   "ROTATION",
		SelectionMode:  "AUTO",
		Status:         "SUCCEEDED",
		UniverseCount:  10,
		SeedCount:      8,
		CandidateCount: 5,
		SelectedCount:  3,
		ReviewStatus:   "PENDING",
		StageCounts:    map[string]int{"UNIVERSE": 10, "SEED_POOL": 8, "CANDIDATE_POOL": 5, "PORTFOLIO": 3},
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	return []model.StockSelectionRun{item}, 1, nil
}

func (r *InMemoryGrowthRepo) AdminCreateStockSelectionRun(input model.StockSelectionRunCreateRequest, operator string) (model.StockSelectionRun, error) {
	items, _, _ := r.AdminListStockSelectionRuns("", "", "", 1, 1)
	return items[0], nil
}

func (r *InMemoryGrowthRepo) AdminGetStockSelectionRun(runID string) (model.StockSelectionRun, error) {
	items, _, _ := r.AdminListStockSelectionRuns("", "", "", 1, 1)
	if len(items) == 0 {
		return model.StockSelectionRun{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r *InMemoryGrowthRepo) AdminCompareStockSelectionRuns(runIDs []string) (model.StockSelectionRunCompareResult, error) {
	return model.StockSelectionRunCompareResult{
		BaseRunID: "ssr_demo_001",
		Items: []model.StockSelectionRunCompareItem{
			{
				RunID:            "ssr_demo_001",
				TradeDate:        time.Now().Format("2006-01-02"),
				ProfileID:        model.StrategyEngineDefaultStockSelectionProfileID,
				TemplateID:       "sstpl_balanced_steady",
				TemplateName:     "均衡稳健",
				MarketRegime:     "ROTATION",
				Status:           "SUCCEEDED",
				ReviewStatus:     "PENDING",
				SelectedCount:    3,
				PortfolioSymbols: []string{"600519.SH", "601318.SH", "300750.SZ"},
			},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListStockSelectionProfiles(status string, page int, pageSize int) ([]model.StockSelectionProfile, int, error) {
	item := model.StockSelectionProfile{
		ID:                   model.StrategyEngineDefaultStockSelectionProfileID,
		Name:                 "默认自动选股",
		TemplateID:           "sstpl_balanced_steady",
		TemplateName:         "均衡稳健",
		Status:               "ACTIVE",
		IsDefault:            true,
		CurrentVersion:       1,
		SelectionModeDefault: "AUTO",
		UniverseScope:        model.StrategyEngineDefaultStockUniverseScope,
		UniverseConfig:       map[string]any{"min_listing_days": 180, "min_avg_turnover": 50000000},
		PortfolioConfig:      map[string]any{"limit": 5, "max_risk_level": "MEDIUM", "min_score": 75},
		Versions: []model.StockSelectionProfileVersion{
			{ID: "profile_default_stock_auto_v1", ProfileID: model.StrategyEngineDefaultStockSelectionProfileID, VersionNo: 1, Snapshot: map[string]any{"name": "默认自动选股"}},
		},
	}
	return []model.StockSelectionProfile{item}, 1, nil
}

func (r *InMemoryGrowthRepo) AdminListStockSelectionProfileVersions(profileID string) ([]model.StockSelectionProfileVersion, error) {
	items, _, _ := r.AdminListStockSelectionProfiles("", 1, 1)
	if len(items) == 0 {
		return nil, nil
	}
	return items[0].Versions, nil
}

func (r *InMemoryGrowthRepo) AdminCreateStockSelectionProfile(item model.StockSelectionProfile, changeNote string) (model.StockSelectionProfile, error) {
	item.ID = "ssp_demo_001"
	item.CurrentVersion = 1
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateStockSelectionProfile(id string, item model.StockSelectionProfile, changeNote string) (model.StockSelectionProfile, error) {
	item.ID = id
	if item.CurrentVersion == 0 {
		item.CurrentVersion = 2
	}
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminPublishStockSelectionProfile(id string, operator string) (model.StockSelectionProfile, error) {
	items, _, _ := r.AdminListStockSelectionProfiles("", 1, 1)
	if len(items) == 0 {
		return model.StockSelectionProfile{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r *InMemoryGrowthRepo) AdminRollbackStockSelectionProfile(id string, versionNo int, changeNote string, operator string) (model.StockSelectionProfile, error) {
	items, _, _ := r.AdminListStockSelectionProfiles("", 1, 1)
	if len(items) == 0 {
		return model.StockSelectionProfile{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r *InMemoryGrowthRepo) AdminListStockSelectionProfileTemplates(status string, page int, pageSize int) ([]model.StockSelectionProfileTemplate, int, error) {
	return []model.StockSelectionProfileTemplate{
		{
			ID:                "sstpl_balanced_steady",
			TemplateKey:       "BALANCED_STEADY",
			Name:              "均衡稳健",
			MarketRegimeBias:  "DEFENSIVE",
			IsDefault:         true,
			Status:            "ACTIVE",
			UniverseDefaults:  map[string]any{"price_min": 5, "price_max": 300},
			SeedDefaults:      map[string]any{"bucket_limit": 36, "seed_pool_cap": 180},
			FactorDefaults:    map[string]any{"quant_weight": 0.7, "event_weight": 0.1, "resonance_weight": 0.1, "liquidity_risk_weight": 0.1},
			PortfolioDefaults: map[string]any{"limit": 5, "watchlist_limit": 5, "max_risk_level": "MEDIUM"},
			PublishDefaults:   map[string]any{"review_required": true},
		},
	}, 1, nil
}

func (r *InMemoryGrowthRepo) AdminCreateStockSelectionProfileTemplate(item model.StockSelectionProfileTemplate) (model.StockSelectionProfileTemplate, error) {
	item.ID = "sstpl_demo_001"
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateStockSelectionProfileTemplate(id string, item model.StockSelectionProfileTemplate) (model.StockSelectionProfileTemplate, error) {
	item.ID = id
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminSetDefaultStockSelectionProfileTemplate(id string, operator string) (model.StockSelectionProfileTemplate, error) {
	items, _, _ := r.AdminListStockSelectionProfileTemplates("", 1, 1)
	if len(items) == 0 {
		return model.StockSelectionProfileTemplate{}, sql.ErrNoRows
	}
	return items[0], nil
}

func (r *InMemoryGrowthRepo) AdminListStockSelectionRunCandidates(runID string) ([]model.StockSelectionCandidateSnapshot, error) {
	return []model.StockSelectionCandidateSnapshot{
		{ID: "cand_demo_001", RunID: runID, Symbol: "600519.SH", Name: "贵州茅台", Stage: "PORTFOLIO", QuantScore: 88.2, RiskLevel: "LOW", Selected: true, Rank: 1, EvidenceSummary: "趋势、资金与质量三项同时达标", PortfolioRole: "CORE", EvaluationStatus: "PENDING", RiskSummary: "中期回撤控制", FactorBreakdownJSON: map[string]any{"trend": 82, "money_flow": 79, "quality": 81, "event": 68, "resonance": 77, "risk_adjustment": 74, "total_score": 84.2}},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListStockSelectionRunPortfolio(runID string) ([]model.StockSelectionPortfolioEntry, error) {
	return []model.StockSelectionPortfolioEntry{
		{ID: "port_demo_001", RunID: runID, Symbol: "600519.SH", Name: "贵州茅台", Rank: 1, QuantScore: 88.2, RiskLevel: "LOW", WeightSuggestion: "10%-15%", EvidenceSummary: "趋势、资金与质量三项同时达标", PortfolioRole: "CORE", EvaluationStatus: "PENDING", RiskSummary: "中期回撤控制", FactorBreakdownJSON: map[string]any{"trend": 82, "money_flow": 79, "quality": 81, "event": 68, "resonance": 77, "risk_adjustment": 74, "total_score": 84.2}},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListStockSelectionRunEvidence(runID string, symbol string) ([]model.StockSelectionRunEvidence, error) {
	return []model.StockSelectionRunEvidence{
		{
			ID:              "ssev_demo_001",
			RunID:           runID,
			Symbol:          firstNonEmpty(symbol, "600519.SH"),
			Name:            "贵州茅台",
			Stage:           "PORTFOLIO",
			PortfolioRole:   "CORE",
			EvidenceSummary: "趋势、资金、质量三项共振，风险边界清晰。",
			EvidenceCards: []map[string]any{
				{"label": "趋势", "value": "20日动量 6.8%", "note": "中期趋势延续"},
				{"label": "资金", "value": "主力净流入", "note": "资金面偏正"},
			},
			PositiveReasons: []string{"20日动量较强", "资金流入稳定"},
			VetoReasons:     []string{"若量比回落则降级为观察"},
			ThemeTags:       []string{"趋势成长", "资金共振"},
			SectorTags:      []string{"消费"},
			RiskFlags:       []string{"回撤超过5%需重审"},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListStockSelectionRunEvaluations(runID string, symbol string) ([]model.StockSelectionRunEvaluation, error) {
	return []model.StockSelectionRunEvaluation{
		{
			ID:              "sseval_demo_001",
			RunID:           runID,
			Symbol:          firstNonEmpty(symbol, "600519.SH"),
			Name:            "贵州茅台",
			HorizonDay:      5,
			EvaluationScope: "PORTFOLIO",
			EntryDate:       time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
			ExitDate:        time.Now().Format("2006-01-02"),
			EntryPrice:      1680,
			ExitPrice:       1720,
			ReturnPct:       0.0238,
			ExcessReturnPct: 0.0112,
			MaxDrawdownPct:  -0.012,
			HitFlag:         true,
			BenchmarkSymbol: "000300.SH",
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListStockSelectionEvaluationLeaderboard(templateID string, profileID string, marketRegime string) ([]model.StockSelectionEvaluationLeaderboardItem, error) {
	return []model.StockSelectionEvaluationLeaderboardItem{
		{
			TemplateID:       "sstpl_balanced_steady",
			TemplateName:     "均衡稳健",
			ProfileID:        model.StrategyEngineDefaultStockSelectionProfileID,
			ProfileName:      "默认自动选股",
			MarketRegime:     firstNonEmpty(marketRegime, "ROTATION"),
			SampleCount:      12,
			ReturnByHorizon:  map[string]float64{"1": 0.004, "3": 0.011, "5": 0.023, "10": 0.041, "20": 0.068},
			HitRateByHorizon: map[string]float64{"1": 0.58, "3": 0.61, "5": 0.66, "10": 0.69, "20": 0.72},
			MaxDrawdownPct:   -0.031,
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListStockSelectionReviews(status string, page int, pageSize int) ([]model.StockSelectionPublishReview, int, error) {
	return []model.StockSelectionPublishReview{
		{ID: "review_demo_001", RunID: "ssr_demo_001", ReviewStatus: "PENDING"},
	}, 1, nil
}

func (r *InMemoryGrowthRepo) AdminApproveStockSelectionReview(runID string, operator string, reviewNote string, force bool, overrideReason string) (model.StockSelectionPublishReview, error) {
	return model.StockSelectionPublishReview{ID: "review_demo_001", RunID: runID, ReviewStatus: "APPROVED", Reviewer: operator, ReviewNote: reviewNote}, nil
}

func (r *InMemoryGrowthRepo) AdminRejectStockSelectionReview(runID string, operator string, reviewNote string) (model.StockSelectionPublishReview, error) {
	return model.StockSelectionPublishReview{ID: "review_demo_001", RunID: runID, ReviewStatus: "REJECTED", Reviewer: operator, ReviewNote: reviewNote}, nil
}
