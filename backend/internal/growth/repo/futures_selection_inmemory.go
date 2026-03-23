package repo

import (
	"database/sql"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *InMemoryGrowthRepo) AdminGetFuturesSelectionOverview() (model.AdminFuturesSelectionOverview, error) {
	profile := model.FuturesSelectionProfile{
		ID:             "profile_default_futures_auto",
		Name:           "默认智能期货",
		TemplateID:     "fstpl_balanced_trend",
		TemplateName:   "均衡趋势",
		Status:         "ACTIVE",
		IsDefault:      true,
		CurrentVersion: 1,
		StyleDefault:   "balanced",
		ContractScope:  "DOMINANT_ALL",
		UniverseConfig: map[string]any{"contract_scope": "DOMINANT_ALL", "allow_mock_fallback_on_short_history": true},
		FactorConfig:   map[string]any{"min_confidence": 55},
		PortfolioConfig: map[string]any{
			"limit":          3,
			"max_risk_level": "HIGH",
		},
		PublishConfig: map[string]any{"review_required": true, "allow_auto_publish": false},
		UpdatedBy:     "system",
		UpdatedAt:     time.Now().UTC().Format(time.RFC3339),
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
	}
	run := model.FuturesSelectionRun{
		RunID:          "fsr_demo_001",
		TradeDate:      time.Now().Format("2006-01-02"),
		ProfileID:      profile.ID,
		ProfileVersion: 1,
		TemplateID:     profile.TemplateID,
		TemplateName:   profile.TemplateName,
		MarketRegime:   "BASE",
		Style:          "balanced",
		ContractScope:  "DOMINANT_ALL",
		Status:         "SUCCEEDED",
		UniverseCount:  6,
		CandidateCount: 4,
		SelectedCount:  3,
		ReviewStatus:   "PENDING",
		ContextMeta: map[string]any{
			"graph_snapshot_id":   "gss_futures_demo",
			"graph_write_status":  "WRITTEN",
			"selected_trade_date": time.Now().Format("2006-01-02"),
		},
		StageCounts: map[string]int{"UNIVERSE": 6, "CANDIDATE_POOL": 4, "PORTFOLIO": 3},
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}
	return model.AdminFuturesSelectionOverview{
		DefaultProfile:          &profile,
		LatestTradeDate:         run.TradeDate,
		LatestRun:               &run,
		LatestSuccessRun:        &run,
		LatestApprovedPortfolio: []model.FuturesSelectionPortfolioEntry{{Contract: "RB9999", Name: "螺纹钢主连", Rank: 1, Direction: "LONG", PortfolioRole: "CORE"}},
		MarketRegime:            run.MarketRegime,
		DataFreshness:           run.ContextMeta,
		PendingReviewCount:      1,
		QuickActions: []model.AdminStockSelectionQuickAction{
			{Key: "run_now", Label: "立即运行", ActionType: "RUN", TargetRoute: "/futures-selection/runs"},
			{Key: "open_runs", Label: "运行中心", ActionType: "NAVIGATE", TargetRoute: "/futures-selection/runs"},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListFuturesSelectionRuns(status string, reviewStatus string, profileID string, page int, pageSize int) ([]model.FuturesSelectionRun, int, error) {
	item := model.FuturesSelectionRun{
		RunID:          "fsr_demo_001",
		TradeDate:      time.Now().Format("2006-01-02"),
		ProfileID:      "profile_default_futures_auto",
		ProfileVersion: 1,
		TemplateID:     "fstpl_balanced_trend",
		TemplateName:   "均衡趋势",
		MarketRegime:   "BASE",
		Style:          "balanced",
		ContractScope:  "DOMINANT_ALL",
		Status:         "SUCCEEDED",
		UniverseCount:  6,
		CandidateCount: 4,
		SelectedCount:  3,
		ReviewStatus:   "PENDING",
		ContextMeta:    map[string]any{"graph_snapshot_id": "gss_futures_demo", "graph_write_status": "WRITTEN"},
		StageCounts:    map[string]int{"UNIVERSE": 6, "CANDIDATE_POOL": 4, "PORTFOLIO": 3},
		CreatedAt:      time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
	}
	return []model.FuturesSelectionRun{item}, 1, nil
}

func (r *InMemoryGrowthRepo) AdminListFuturesSelectionProfiles(status string, page int, pageSize int) ([]model.FuturesSelectionProfile, int, error) {
	item := model.FuturesSelectionProfile{
		ID:             "profile_default_futures_auto",
		Name:           "默认智能期货",
		TemplateID:     "fstpl_balanced_trend",
		TemplateName:   "均衡趋势",
		Status:         "ACTIVE",
		IsDefault:      true,
		CurrentVersion: 2,
		StyleDefault:   "balanced",
		ContractScope:  "DOMINANT_ALL",
		UniverseConfig: map[string]any{
			"contract_scope":                       "DOMINANT_ALL",
			"allow_mock_fallback_on_short_history": true,
			"contracts":                            []string{"IF2606", "RB2605"},
		},
		FactorConfig: map[string]any{"min_confidence": 55},
		PortfolioConfig: map[string]any{
			"limit":          3,
			"max_risk_level": "HIGH",
		},
		PublishConfig: map[string]any{
			"review_required":    true,
			"allow_auto_publish": false,
		},
		Description: "默认期货研究配置，适合做日常同步验证。",
		Versions: []model.FuturesSelectionProfileVersion{
			{
				ID:        "profile_default_futures_auto_v2",
				ProfileID: "profile_default_futures_auto",
				VersionNo: 2,
				Snapshot: map[string]any{
					"name":             "默认智能期货",
					"style_default":    "balanced",
					"contract_scope":   "DOMINANT_ALL",
					"factor_config":    map[string]any{"min_confidence": 55},
					"portfolio_config": map[string]any{"limit": 3, "max_risk_level": "HIGH"},
				},
				ChangeNote: "调整默认组合上限与置信度阈值",
				CreatedBy:  "system",
				CreatedAt:  time.Now().Add(-12 * time.Hour).UTC().Format(time.RFC3339),
			},
			{
				ID:        "profile_default_futures_auto_v1",
				ProfileID: "profile_default_futures_auto",
				VersionNo: 1,
				Snapshot: map[string]any{
					"name":             "默认智能期货",
					"style_default":    "trend",
					"contract_scope":   "DOMINANT_ALL",
					"factor_config":    map[string]any{"min_confidence": 50},
					"portfolio_config": map[string]any{"limit": 2, "max_risk_level": "MEDIUM"},
				},
				ChangeNote: "初始化版本",
				CreatedBy:  "system",
				CreatedAt:  time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339),
			},
		},
		UpdatedBy: "system",
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		CreatedAt: time.Now().Add(-72 * time.Hour).UTC().Format(time.RFC3339),
	}
	return []model.FuturesSelectionProfile{item}, 1, nil
}

func (r *InMemoryGrowthRepo) AdminListFuturesSelectionProfileVersions(profileID string) ([]model.FuturesSelectionProfileVersion, error) {
	items, _, _ := r.AdminListFuturesSelectionProfiles("", 1, 1)
	if len(items) == 0 {
		return nil, nil
	}
	if profileID != "" && items[0].ID != profileID {
		return []model.FuturesSelectionProfileVersion{}, nil
	}
	return items[0].Versions, nil
}

func (r *InMemoryGrowthRepo) AdminCreateFuturesSelectionProfile(item model.FuturesSelectionProfile, changeNote string) (model.FuturesSelectionProfile, error) {
	item.ID = "fspf_demo_001"
	item.CurrentVersion = 1
	item.TemplateName = "均衡趋势"
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	item.CreatedAt = item.UpdatedAt
	item.Versions = []model.FuturesSelectionProfileVersion{
		{
			ID:         "fspf_demo_001_v1",
			ProfileID:  "fspf_demo_001",
			VersionNo:  1,
			Snapshot:   map[string]any{"name": item.Name},
			ChangeNote: firstNonEmpty(changeNote, "创建期货配置"),
			CreatedBy:  firstNonEmpty(item.UpdatedBy, "admin"),
			CreatedAt:  item.UpdatedAt,
		},
	}
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateFuturesSelectionProfile(id string, item model.FuturesSelectionProfile, changeNote string) (model.FuturesSelectionProfile, error) {
	item.ID = id
	if item.CurrentVersion == 0 {
		item.CurrentVersion = 2
	}
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminPublishFuturesSelectionProfile(id string, operator string) (model.FuturesSelectionProfile, error) {
	items, _, _ := r.AdminListFuturesSelectionProfiles("", 1, 1)
	if len(items) == 0 {
		return model.FuturesSelectionProfile{}, sql.ErrNoRows
	}
	item := items[0]
	item.ID = id
	item.IsDefault = true
	item.UpdatedBy = operator
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminRollbackFuturesSelectionProfile(id string, versionNo int, changeNote string, operator string) (model.FuturesSelectionProfile, error) {
	items, _, _ := r.AdminListFuturesSelectionProfiles("", 1, 1)
	if len(items) == 0 {
		return model.FuturesSelectionProfile{}, sql.ErrNoRows
	}
	item := items[0]
	item.ID = id
	item.CurrentVersion = versionNo + 1
	item.UpdatedBy = operator
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminListFuturesSelectionProfileTemplates(status string, page int, pageSize int) ([]model.FuturesSelectionProfileTemplate, int, error) {
	items := []model.FuturesSelectionProfileTemplate{
		{
			ID:               "fstpl_balanced_trend",
			TemplateKey:      "BALANCED_TREND",
			Name:             "均衡趋势",
			Description:      "兼顾趋势延续、流动性和风险上限的默认期货模板。",
			MarketRegimeBias: "BASE",
			IsDefault:        true,
			Status:           "ACTIVE",
			UniverseDefaults: map[string]any{
				"contract_scope":                       "DOMINANT_ALL",
				"allow_mock_fallback_on_short_history": true,
				"style":                                "balanced",
			},
			FactorDefaults:    map[string]any{"min_confidence": 55},
			PortfolioDefaults: map[string]any{"limit": 3, "max_risk_level": "HIGH"},
			PublishDefaults:   map[string]any{"review_required": true, "allow_auto_publish": false},
			UpdatedBy:         "system",
			UpdatedAt:         time.Now().UTC().Format(time.RFC3339),
			CreatedAt:         time.Now().Add(-72 * time.Hour).UTC().Format(time.RFC3339),
		},
		{
			ID:               "fstpl_trend_attack",
			TemplateKey:      "TREND_ATTACK",
			Name:             "趋势进攻",
			Description:      "提升趋势风格与风险预算，适合趋势延续阶段。",
			MarketRegimeBias: "TREND_CONTINUE",
			IsDefault:        false,
			Status:           "ACTIVE",
			UniverseDefaults: map[string]any{"contract_scope": "DOMINANT_ALL", "style": "trend"},
			FactorDefaults:   map[string]any{"min_confidence": 62},
			PortfolioDefaults: map[string]any{
				"limit":          4,
				"max_risk_level": "HIGH",
			},
			PublishDefaults: map[string]any{"review_required": true, "allow_auto_publish": false},
			UpdatedBy:       "system",
			UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
			CreatedAt:       time.Now().Add(-48 * time.Hour).UTC().Format(time.RFC3339),
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminCreateFuturesSelectionProfileTemplate(item model.FuturesSelectionProfileTemplate) (model.FuturesSelectionProfileTemplate, error) {
	item.ID = "fstpl_demo_001"
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	item.CreatedAt = item.UpdatedAt
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateFuturesSelectionProfileTemplate(id string, item model.FuturesSelectionProfileTemplate) (model.FuturesSelectionProfileTemplate, error) {
	item.ID = id
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminSetDefaultFuturesSelectionProfileTemplate(id string, operator string) (model.FuturesSelectionProfileTemplate, error) {
	items, _, _ := r.AdminListFuturesSelectionProfileTemplates("", 1, 10)
	if len(items) == 0 {
		return model.FuturesSelectionProfileTemplate{}, sql.ErrNoRows
	}
	item := items[0]
	item.ID = id
	item.IsDefault = true
	item.UpdatedBy = operator
	item.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminListFuturesSelectionEvaluationLeaderboard(templateID string, profileID string, marketRegime string) ([]model.FuturesSelectionEvaluationLeaderboardItem, error) {
	return []model.FuturesSelectionEvaluationLeaderboardItem{
		{
			TemplateID:       "fstpl_balanced_trend",
			TemplateName:     "均衡趋势",
			ProfileID:        "profile_default_futures_auto",
			ProfileName:      "默认智能期货",
			MarketRegime:     "BASE",
			SampleCount:      6,
			ReturnByHorizon:  map[string]float64{"1": 0.003, "3": 0.012, "5": 0.021, "10": 0.035, "20": 0.052},
			HitRateByHorizon: map[string]float64{"1": 0.57, "3": 0.6, "5": 0.63, "10": 0.68, "20": 0.72},
			MaxDrawdownPct:   -0.028,
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminCreateFuturesSelectionRun(input model.FuturesSelectionRunCreateRequest, operator string) (model.FuturesSelectionRun, error) {
	items, _, _ := r.AdminListFuturesSelectionRuns("", "", "", 1, 1)
	return items[0], nil
}

func (r *InMemoryGrowthRepo) AdminGetFuturesSelectionRun(runID string) (model.FuturesSelectionRun, error) {
	items, _, _ := r.AdminListFuturesSelectionRuns("", "", "", 1, 1)
	if len(items) == 0 {
		return model.FuturesSelectionRun{}, sql.ErrNoRows
	}
	item := items[0]
	item.RunID = runID
	item.StageLogs = []model.FuturesSelectionRunStageLog{
		{ID: "fslg_001", RunID: runID, StageKey: "MARKET_REGIME", StageOrder: 1, Status: "SUCCEEDED", InputCount: 6, OutputCount: 6, DetailMessage: "detected futures_regime=BASE"},
		{ID: "fslg_002", RunID: runID, StageKey: "UNIVERSE", StageOrder: 2, Status: "SUCCEEDED", InputCount: 6, OutputCount: 6, DetailMessage: "contract universe normalized"},
		{ID: "fslg_003", RunID: runID, StageKey: "GRAPH_ENRICHMENT", StageOrder: 3, Status: "SUCCEEDED", InputCount: 6, OutputCount: 7, DetailMessage: "futures graph enrichment completed"},
		{ID: "fslg_004", RunID: runID, StageKey: "CANDIDATE_POOL", StageOrder: 4, Status: "SUCCEEDED", InputCount: 6, OutputCount: 4, DetailMessage: "ranked futures candidates by conviction score"},
		{ID: "fslg_005", RunID: runID, StageKey: "PORTFOLIO", StageOrder: 5, Status: "SUCCEEDED", InputCount: 4, OutputCount: 3, DetailMessage: "portfolio limit=3 max_risk=HIGH"},
	}
	item.StageDurationsMS = buildFuturesSelectionStageDurationMap(item.StageLogs)
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminCompareFuturesSelectionRuns(runIDs []string) (model.FuturesSelectionRunCompareResult, error) {
	return model.FuturesSelectionRunCompareResult{
		BaseRunID: "fsr_demo_001",
		Items: []model.FuturesSelectionRunCompareItem{
			{
				RunID:              "fsr_demo_001",
				TradeDate:          time.Now().Format("2006-01-02"),
				ProfileID:          "profile_default_futures_auto",
				TemplateID:         "fstpl_balanced_trend",
				TemplateName:       "均衡趋势",
				MarketRegime:       "BASE",
				Status:             "SUCCEEDED",
				ReviewStatus:       "PENDING",
				SelectedCount:      3,
				PortfolioContracts: []string{"IF2606", "RB2605", "CU2605"},
			},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListFuturesSelectionRunCandidates(runID string) ([]model.FuturesSelectionCandidateSnapshot, error) {
	return []model.FuturesSelectionCandidateSnapshot{
		{
			ID:               "fsc_demo_001",
			RunID:            runID,
			Contract:         "IF2606",
			Name:             "沪深300股指主力",
			Stage:            "CANDIDATE_POOL",
			Score:            72.5,
			Direction:        "SHORT",
			RiskLevel:        "HIGH",
			Selected:         true,
			Rank:             1,
			ReasonSummary:    "趋势与风险过滤后保留的空头候选。",
			EvidenceSummary:  "股指风险偏好偏弱，空头信号确认。",
			PortfolioRole:    "CORE",
			EvaluationStatus: "PENDING",
			RiskSummary:      "注意宏观消息冲击。",
			FactorBreakdownJSON: map[string]any{
				"trend": 68,
				"basis": 55,
				"macro": 71,
				"risk":  60,
				"total": 72.5,
			},
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListFuturesSelectionRunPortfolio(runID string) ([]model.FuturesSelectionPortfolioEntry, error) {
	return []model.FuturesSelectionPortfolioEntry{
		{
			ID:               "fsp_demo_001",
			RunID:            runID,
			Contract:         "IF2606",
			Name:             "沪深300股指主力",
			Rank:             1,
			Score:            72.5,
			Direction:        "SHORT",
			RiskLevel:        "HIGH",
			PositionRange:    "20%-30%",
			ReasonSummary:    "趋势确认后纳入组合。",
			EvidenceSummary:  "图谱与趋势信号共振。",
			PortfolioRole:    "CORE",
			EvaluationStatus: "PENDING",
			RiskSummary:      "注意波动率放大。",
			CreatedAt:        time.Now().UTC().Format(time.RFC3339),
			UpdatedAt:        time.Now().UTC().Format(time.RFC3339),
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListFuturesSelectionRunEvidence(runID string, contract string) ([]model.FuturesSelectionRunEvidence, error) {
	return []model.FuturesSelectionRunEvidence{
		{
			ID:              "fsev_demo_001",
			RunID:           runID,
			Contract:        "IF2606",
			Name:            "沪深300股指主力",
			Stage:           "PORTFOLIO",
			PortfolioRole:   "CORE",
			EvidenceSummary: "趋势、宏观和图谱风险传导共同确认空头方向。",
			EvidenceCards: []map[string]any{
				{"title": "市场状态", "value": "BASE"},
				{"title": "方向", "value": "SHORT"},
			},
			PositiveReasons: []string{"趋势拐点确认", "图谱关系支持空头"},
			VetoReasons:     []string{"需要防范政策反转"},
			RiskFlags:       []string{"高波动", "事件驱动"},
			RelatedEntities: []map[string]any{{"label": "BASE 期货状态", "entity_type": "Policy"}},
			CreatedAt:       time.Now().UTC().Format(time.RFC3339),
			UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListFuturesSelectionRunEvaluations(runID string, contract string) ([]model.FuturesSelectionRunEvaluation, error) {
	return []model.FuturesSelectionRunEvaluation{
		{
			ID:              "fsee_demo_001",
			RunID:           runID,
			Contract:        "IF2606",
			Name:            "沪深300股指主力",
			HorizonDay:      5,
			EvaluationScope: "PORTFOLIO",
			EntryDate:       time.Now().AddDate(0, 0, -5).Format("2006-01-02"),
			ExitDate:        "",
			EntryPrice:      3620,
			ExitPrice:       0,
			ReturnPct:       0,
			ExcessReturnPct: 0,
			MaxDrawdownPct:  0,
			HitFlag:         false,
			BenchmarkSymbol: "IF",
			CreatedAt:       time.Now().UTC().Format(time.RFC3339),
			UpdatedAt:       time.Now().UTC().Format(time.RFC3339),
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminApproveFuturesSelectionReview(runID string, operator string, reviewNote string, force bool, overrideReason string) (model.FuturesSelectionPublishReview, error) {
	return model.FuturesSelectionPublishReview{
		ID:             "fsrv_demo_001",
		RunID:          runID,
		ReviewStatus:   "APPROVED",
		Reviewer:       operator,
		ReviewNote:     reviewNote,
		OverrideReason: overrideReason,
		PublishID:      "publish_demo_futures_001",
		PublishVersion: 1,
		ApprovedAt:     time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:      time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (r *InMemoryGrowthRepo) AdminRejectFuturesSelectionReview(runID string, operator string, reviewNote string) (model.FuturesSelectionPublishReview, error) {
	return model.FuturesSelectionPublishReview{
		ID:           "fsrv_demo_001",
		RunID:        runID,
		ReviewStatus: "REJECTED",
		Reviewer:     operator,
		ReviewNote:   reviewNote,
		RejectedAt:   time.Now().UTC().Format(time.RFC3339),
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}
