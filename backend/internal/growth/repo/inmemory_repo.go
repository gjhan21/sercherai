package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"sercherai/backend/internal/growth/model"
)

type InMemoryGrowthRepo struct {
	mu                        sync.Mutex
	marketRhythmTasks         map[string]model.MarketRhythmTask
	marketBackfillRuns        map[string]model.MarketBackfillRun
	marketBackfillRunDetails  map[string][]model.MarketBackfillRunDetail
	marketUniverseSnapshots   map[string]model.MarketUniverseSnapshot
	marketUniverseItems       map[string][]model.MarketUniverseSnapshotItem
	forecastL3Runs            map[string]model.StrategyForecastL3Run
	forecastL3Reports         map[string]model.StrategyForecastL3Report
	forecastL3Logs            map[string][]model.StrategyForecastL3Log
	forecastL3Learning        map[string][]model.StrategyForecastL3LearningRecord
	stockEventClusters        map[string]model.StockEventCluster
	futuresInstrumentProfiles map[string]model.FuturesInstrumentProfile
	reviewTasks               map[string]model.ReviewTask
	communityTopics           map[string]model.CommunityTopicDetail
	communityComments         map[string]model.CommunityComment
	communityReports          map[string]model.CommunityReport
	communityReacts           map[string]struct{}
	userMessages              map[string][]model.UserMessage
	adminAuditEvents          map[string]model.AdminAuditEvent
	workflowMessages          map[string]model.WorkflowMessage
}

func NewInMemoryGrowthRepo() *InMemoryGrowthRepo {
	now := time.Now().Format(time.RFC3339)
	snapshot := model.MarketUniverseSnapshot{
		ID:           "mus_demo_001",
		Scope:        []string{"STOCK", "INDEX"},
		SourceKey:    "TUSHARE",
		SnapshotDate: "2026-03-24",
		AssetSummaries: []model.MarketUniverseSnapshotAssetItem{
			{AssetType: "STOCK", ItemCount: 2, ActiveCount: 2, InactiveCount: 0},
			{AssetType: "INDEX", ItemCount: 1, ActiveCount: 1, InactiveCount: 0},
		},
		CreatedBy: "system",
		CreatedAt: now,
	}
	run := model.MarketBackfillRun{
		ID:                 "mbr_demo_001",
		SchedulerRunID:     "jr_demo_market_001",
		RunType:            "FULL",
		AssetScope:         []string{"STOCK", "INDEX"},
		SourceKey:          "TUSHARE",
		BatchSize:          200,
		UniverseSnapshotID: snapshot.ID,
		Status:             "RUNNING",
		CurrentStage:       "MASTER",
		StageProgress: []model.MarketBackfillStageProgress{
			{Stage: "UNIVERSE", Status: "SUCCESS", TotalBatches: 1, CompletedBatches: 1},
			{Stage: "MASTER", Status: "RUNNING", TotalBatches: 2, CompletedBatches: 1},
		},
		Summary: map[string]any{
			"latest_trade_date": "2026-03-24",
		},
		CreatedBy: "system",
		CreatedAt: now,
		UpdatedAt: now,
	}
	repo := &InMemoryGrowthRepo{
		marketRhythmTasks:  make(map[string]model.MarketRhythmTask),
		marketBackfillRuns: map[string]model.MarketBackfillRun{run.ID: run},
		marketBackfillRunDetails: map[string][]model.MarketBackfillRunDetail{
			run.ID: {
				{
					ID:             "mbd_demo_001",
					RunID:          run.ID,
					SchedulerRunID: run.SchedulerRunID,
					Stage:          "UNIVERSE",
					AssetType:      "STOCK",
					BatchKey:       "UNIVERSE-STOCK-001",
					Status:         "SUCCESS",
					SymbolCount:    2,
					SymbolSample:   []string{"600519.SH", "000001.SZ"},
					StartedAt:      now,
					FinishedAt:     now,
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
		},
		marketUniverseSnapshots: map[string]model.MarketUniverseSnapshot{snapshot.ID: snapshot},
		marketUniverseItems: map[string][]model.MarketUniverseSnapshotItem{
			snapshot.ID: {
				{ID: "musi_demo_001", SnapshotID: snapshot.ID, AssetType: "STOCK", InstrumentKey: "600519.SH", ExternalSymbol: "600519.SH", DisplayName: "贵州茅台", ExchangeCode: "SH", Status: "LISTED", CreatedAt: now},
				{ID: "musi_demo_002", SnapshotID: snapshot.ID, AssetType: "STOCK", InstrumentKey: "000001.SZ", ExternalSymbol: "000001.SZ", DisplayName: "平安银行", ExchangeCode: "SZ", Status: "LISTED", CreatedAt: now},
				{ID: "musi_demo_003", SnapshotID: snapshot.ID, AssetType: "INDEX", InstrumentKey: "000300.SH", ExternalSymbol: "000300.SH", DisplayName: "沪深300", ExchangeCode: "SH", Status: "ACTIVE", CreatedAt: now},
			},
		},
		forecastL3Runs:            make(map[string]model.StrategyForecastL3Run),
		forecastL3Reports:         make(map[string]model.StrategyForecastL3Report),
		forecastL3Logs:            make(map[string][]model.StrategyForecastL3Log),
		forecastL3Learning:        make(map[string][]model.StrategyForecastL3LearningRecord),
		stockEventClusters:        make(map[string]model.StockEventCluster),
		futuresInstrumentProfiles: make(map[string]model.FuturesInstrumentProfile),
		reviewTasks:               make(map[string]model.ReviewTask),
		communityTopics:           make(map[string]model.CommunityTopicDetail),
		communityComments:         make(map[string]model.CommunityComment),
		communityReports:          make(map[string]model.CommunityReport),
		communityReacts:           make(map[string]struct{}),
		userMessages:              make(map[string][]model.UserMessage),
		adminAuditEvents:          make(map[string]model.AdminAuditEvent),
		workflowMessages:          make(map[string]model.WorkflowMessage),
	}
	repo.seedCommunityData()
	return repo
}

func (r *InMemoryGrowthRepo) ListBrowseHistory(userID string, contentType string, page int, pageSize int) ([]model.BrowseHistory, int, error) {
	items := []model.BrowseHistory{
		{ID: "h_001", ContentType: "NEWS", ContentID: "news_001", Title: "政策观察", SourcePage: "/news", ViewedAt: "2026-02-24T10:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminUpsertFuturesInstrumentProfile(item model.FuturesInstrumentProfile) (model.FuturesInstrumentProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item = normalizeFuturesInstrumentProfile(item)
	now := time.Now().Format(time.RFC3339)
	if item.CreatedAt == "" {
		item.CreatedAt = now
	}
	item.UpdatedAt = now
	r.futuresInstrumentProfiles[item.ProductKey] = item
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminGetFuturesInstrumentProfile(productKey string) (model.FuturesInstrumentProfile, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	productKey = strings.ToUpper(strings.TrimSpace(productKey))
	item, ok := r.futuresInstrumentProfiles[productKey]
	if !ok {
		return model.FuturesInstrumentProfile{}, sql.ErrNoRows
	}
	return item, nil
}

func (r *InMemoryGrowthRepo) DeleteBrowseHistoryItem(userID string, id string) error {
	return nil
}

func (r *InMemoryGrowthRepo) ClearBrowseHistory(userID string) error {
	return nil
}

func (r *InMemoryGrowthRepo) ListRechargeRecords(userID string, status string, page int, pageSize int) ([]model.RechargeRecord, int, error) {
	items := []model.RechargeRecord{
		{ID: "rc_001", OrderNo: "O20260224001", Amount: 99, PayChannel: "ALIPAY", Status: "PAID", PaidAt: "2026-02-24T11:00:00+08:00", Remark: ""},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) ListShareLinks(userID string) ([]model.ShareLink, error) {
	items := []model.ShareLink{
		{ID: "sl_001", InviteCode: "ABCD1234", URL: "/invite/ABCD1234", Channel: "wechat", Status: "ACTIVE", ExpiredAt: ""},
	}
	return items, nil
}

func (r *InMemoryGrowthRepo) CreateShareLink(userID string, channel string, expiredAt string) (model.ShareLink, error) {
	return model.ShareLink{
		ID:         "sl_002",
		InviteCode: "NEWC1234",
		URL:        "/invite/NEWC1234",
		Channel:    channel,
		Status:     "ACTIVE",
		ExpiredAt:  expiredAt,
	}, nil
}

func (r *InMemoryGrowthRepo) ListInviteRecords(userID string, page int, pageSize int) ([]model.InviteRecord, int, error) {
	items := []model.InviteRecord{
		{ID: "iv_001", InviteeUser: "u_1002", Status: "FIRST_PAID", RegisterAt: "2026-02-20T12:00:00+08:00", FirstPayAt: "2026-02-21T09:00:00+08:00", RiskFlag: "NORMAL"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) GetUserInviteSummary(userID string) (model.InviteSummary, error) {
	return model.InviteSummary{
		ShareLinkCount:         1,
		RegisteredCount:        1,
		FirstPaidCount:         1,
		ConversionRate:         1,
		Last7dRegisteredCount:  1,
		Last7dFirstPaidCount:   1,
		Last7dConversionRate:   1,
		Last30dRegisteredCount: 1,
		Last30dFirstPaidCount:  1,
		Last30dConversionRate:  1,
	}, nil
}

func (r *InMemoryGrowthRepo) ListRewardRecords(userID string, page int, pageSize int) ([]model.RewardRecord, int, error) {
	items := []model.RewardRecord{
		{ID: "rw_001", RewardType: "CASH", RewardValue: 20, TriggerEvent: "INVITEE_FIRST_RECHARGE", Status: "ISSUED", IssuedAt: "2026-02-21T10:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) GetUserProfile(userID string) (model.UserProfile, error) {
	return model.UserProfile{
		ID:                 userID,
		Phone:              "13800000001",
		Email:              "demo@sercherai.local",
		KYCStatus:          "PENDING",
		MemberLevel:        "VIP1",
		ActivationState:    "PAID_PENDING_KYC",
		VIPStartedAt:       "2026-02-20T00:00:00+08:00",
		VIPExpireAt:        "2026-03-20T23:59:59+08:00",
		VIPStatus:          "ACTIVE",
		VIPRemainingDays:   20,
		RegistrationSource: "INVITED",
		InviterUserID:      "u_demo_inviter",
		InviteCode:         "DEMO2026",
		InviteLinkID:       "sl_demo_001",
		InvitedAt:          "2026-02-20T10:00:00+08:00",
	}, nil
}

func (r *InMemoryGrowthRepo) UpdateUserProfileEmail(userID string, email string) error {
	return nil
}

func (r *InMemoryGrowthRepo) SubmitUserKYC(userID string, realName string, idNumber string, provider string) (string, error) {
	return "PENDING", nil
}

func (r *InMemoryGrowthRepo) ListSubscriptions(userID string, page int, pageSize int) ([]model.Subscription, int, error) {
	items := []model.Subscription{
		{ID: "sub_demo_001", Type: "STOCK_RECO", Scope: "ALL", Frequency: "DAILY", Status: "ACTIVE"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) CreateSubscription(userID string, subType string, scope string, frequency string) (string, error) {
	return "sub_demo_002", nil
}

func (r *InMemoryGrowthRepo) UpdateSubscription(userID string, id string, frequency string, status string) error {
	return nil
}

func (r *InMemoryGrowthRepo) ListMessages(userID string, page int, pageSize int) ([]model.UserMessage, int, error) {
	items := []model.UserMessage{
		{ID: "msg_demo_001", Title: "系统通知", Type: "SYSTEM", ReadStatus: "UNREAD", CreatedAt: "2026-02-25T08:30:00+08:00"},
	}

	if extra := r.userMessages[strings.TrimSpace(userID)]; len(extra) > 0 {
		items = append(items, extra...)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})

	total := len(items)
	start, end := paginateBounds(page, pageSize, total)
	if start >= total {
		return []model.UserMessage{}, total, nil
	}
	return items[start:end], total, nil
}

func (r *InMemoryGrowthRepo) MarkMessageRead(userID string, id string) error {
	return nil
}

func (r *InMemoryGrowthRepo) GetUserAccessProfile(userID string) (model.UserAccessProfile, error) {
	return model.UserAccessProfile{
		UserID:          userID,
		Status:          "ACTIVE",
		KYCStatus:       "PENDING",
		MemberLevel:     "VIP1",
		ActivationState: "PAID_PENDING_KYC",
	}, nil
}

func (r *InMemoryGrowthRepo) GetMembershipQuota(userID string) (model.MembershipQuota, error) {
	return model.MembershipQuota{
		MemberLevel:            "VIP1",
		KYCStatus:              "PENDING",
		ActivationState:        "PAID_PENDING_KYC",
		PeriodKey:              "2026-02",
		DocReadLimit:           100,
		DocReadUsed:            24,
		DocReadRemaining:       76,
		NewsSubscribeLimit:     50,
		NewsSubscribeUsed:      12,
		NewsSubscribeRemaining: 38,
		ResetCycle:             "MONTHLY",
		ResetAt:                "2026-03-01T00:00:00+08:00",
		VIPExpireAt:            "2026-03-20T23:59:59+08:00",
		VIPStatus:              "ACTIVE",
		VIPRemainingDays:       20,
	}, nil
}

func (r *InMemoryGrowthRepo) GetAttachmentFileInfo(userID string, attachmentID string) (model.AttachmentFileInfo, error) {
	return model.AttachmentFileInfo{
		FileURL:   "https://example.com/file.pdf?sign=demo",
		ArticleID: "article_demo_001",
	}, nil
}

func (r *InMemoryGrowthRepo) LogAttachmentDownload(userID string, attachmentID string, articleID string) error {
	return nil
}

func (r *InMemoryGrowthRepo) ListNewsCategories(userID string) ([]model.NewsCategory, error) {
	return []model.NewsCategory{
		{ID: "nc_001", Name: "盘前速递", Slug: "pre-market", Sort: 1, Visibility: "PUBLIC", Status: "PUBLISHED"},
		{ID: "nc_002", Name: "VIP深度", Slug: "vip-insight", Sort: 2, Visibility: "VIP", Status: "PUBLISHED"},
	}, nil
}

func (r *InMemoryGrowthRepo) ListNewsArticles(userID string, categoryID string, keyword string, page int, pageSize int) ([]model.NewsArticle, int, error) {
	items := []model.NewsArticle{
		{
			ID:              "na_001",
			CategoryID:      "nc_001",
			Title:           "A股盘前观察",
			Summary:         "示例摘要",
			CoverURL:        "https://images.unsplash.com/photo-1611974789855-9c2a0a7236a3?auto=format&fit=crop&w=800&q=80",
			Visibility:      "PUBLIC",
			Status:          "PUBLISHED",
			PublishedAt:     "2026-02-25T09:00:00+08:00",
			AttachmentCount: 1,
		},
		{
			ID:              "na_002",
			CategoryID:      "nc_002",
			Title:           "VIP盘后复盘",
			Summary:         "仅供会员查看的复盘摘要",
			CoverURL:        "https://images.unsplash.com/photo-1460925895917-afdab827c52f?auto=format&fit=crop&w=800&q=80",
			Visibility:      "VIP",
			Status:          "PUBLISHED",
			PublishedAt:     "2026-02-25T20:30:00+08:00",
			AttachmentCount: 1,
		},
	}
	filtered := make([]model.NewsArticle, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(categoryID) != "" && item.CategoryID != categoryID {
			continue
		}
		if strings.TrimSpace(keyword) != "" {
			kw := strings.ToLower(keyword)
			merged := strings.ToLower(item.Title + " " + item.Summary)
			if !strings.Contains(merged, kw) {
				continue
			}
		}
		filtered = append(filtered, item)
	}
	return filtered, len(filtered), nil
}

func (r *InMemoryGrowthRepo) GetNewsArticleDetail(userID string, articleID string) (model.NewsArticle, error) {
	return model.NewsArticle{
		ID:              "na_001",
		CategoryID:      "nc_001",
		Title:           "A股盘前观察",
		Summary:         "示例摘要",
		Content:         "示例正文",
		CoverURL:        "https://images.unsplash.com/photo-1611974789855-9c2a0a7236a3?auto=format&fit=crop&w=800&q=80",
		Visibility:      "PUBLIC",
		Status:          "PUBLISHED",
		PublishedAt:     "2026-02-25T09:00:00+08:00",
		AuthorID:        "admin_001",
		AttachmentCount: 1,
	}, nil
}

func (r *InMemoryGrowthRepo) ListNewsAttachments(userID string, articleID string) ([]model.NewsAttachment, error) {
	return []model.NewsAttachment{
		{ID: "att_001", ArticleID: articleID, FileName: "daily-note.pdf", FileURL: "https://example.com/files/daily-note.pdf", FileSize: 20480, MimeType: "application/pdf", CreatedAt: "2026-02-25T10:00:00+08:00"},
	}, nil
}

func (r *InMemoryGrowthRepo) ListStockRecommendations(userID string, tradeDate string, page int, pageSize int) ([]model.StockRecommendation, int, error) {
	items := []model.StockRecommendation{
		{ID: "sr_001", Symbol: "600519.SH", Name: "贵州茅台", Score: 91.2, RiskLevel: "MEDIUM", PositionRange: "10%-15%", ValidFrom: "2026-02-25T09:00:00+08:00", ValidTo: "2026-02-26T15:00:00+08:00", Status: "PUBLISHED", ReasonSummary: "基本面和资金流共振"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) GetStockRecommendationDetail(userID string, recoID string) (model.StockRecommendationDetail, error) {
	return model.StockRecommendationDetail{
		RecoID: recoID, TechScore: 88, FundScore: 92, SentimentScore: 85, MoneyFlowScore: 90,
		TakeProfit: "上涨8%-12%分批止盈", StopLoss: "跌破关键支撑位止损", RiskNote: "注意市场系统性波动",
	}, nil
}

func (r *InMemoryGrowthRepo) GetStockRecommendationPerformance(userID string, recoID string) ([]model.RecommendationPerformancePoint, error) {
	return []model.RecommendationPerformancePoint{
		{Date: "2026-02-24", Return: 0.012},
		{Date: "2026-02-25", Return: 0.021},
		{Date: "2026-02-26", Return: 0.018},
	}, nil
}

func (r *InMemoryGrowthRepo) GetStockRecommendationInsight(userID string, recoID string) (model.StockRecommendationInsight, error) {
	recommendation := model.StockRecommendation{
		ID:            recoID,
		Symbol:        "600519.SH",
		Name:          "贵州茅台",
		Score:         91.2,
		RiskLevel:     "MEDIUM",
		PositionRange: "10%-15%",
		ValidFrom:     "2026-02-25T09:00:00+08:00",
		ValidTo:       "2026-03-01T15:00:00+08:00",
		Status:        "PUBLISHED",
		ReasonSummary: "基本面和资金流共振，阶段性趋势延续",
	}
	detail, _ := r.GetStockRecommendationDetail(userID, recoID)
	performance, _ := r.GetStockRecommendationPerformance(userID, recoID)
	benchmark := make([]model.RecommendationPerformancePoint, 0, len(performance))
	for _, item := range performance {
		benchmark = append(benchmark, model.RecommendationPerformancePoint{
			Date:   item.Date,
			Return: roundTo(item.Return*0.55, 4),
		})
	}
	framework := model.StockRecommendationScoreFramework{
		Method:        "growth-v1 (tech30 + fund30 + sentiment20 + flow20)",
		TotalScore:    recommendation.Score,
		WeightedScore: 89.5,
		ScoreGap:      1.7,
		Factors: []model.StockRecommendationFactorScore{
			{Key: "tech", Label: "技术因子", Weight: 0.30, Score: detail.TechScore, Contribution: roundTo(detail.TechScore*0.30, 2)},
			{Key: "fund", Label: "基本面因子", Weight: 0.30, Score: detail.FundScore, Contribution: roundTo(detail.FundScore*0.30, 2)},
			{Key: "sentiment", Label: "情绪因子", Weight: 0.20, Score: detail.SentimentScore, Contribution: roundTo(detail.SentimentScore*0.20, 2)},
			{Key: "flow", Label: "资金流因子", Weight: 0.20, Score: detail.MoneyFlowScore, Contribution: roundTo(detail.MoneyFlowScore*0.20, 2)},
		},
	}

	return model.StockRecommendationInsight{
		Recommendation: recommendation,
		Detail:         detail,
		ScoreFramework: framework,
		RelatedNews: []model.StockRecommendationRelatedNews{
			{
				ID:             "na_demo_001",
				Title:          "白酒龙头获资金连续净流入",
				Summary:        "板块景气度维持，主力资金延续净买入。",
				Source:         "新闻快讯",
				Visibility:     "PUBLIC",
				PublishedAt:    "2026-02-26T10:12:00+08:00",
				RelevanceScore: 0.89,
			},
			{
				ID:             "na_demo_002",
				Title:          "机构研报上调消费龙头盈利预测",
				Summary:        "渠道恢复节奏稳健，估值回到中枢区间。",
				Source:         "券商研报",
				Visibility:     "VIP",
				PublishedAt:    "2026-02-25T20:08:00+08:00",
				RelevanceScore: 0.81,
			},
		},
		Performance: performance,
		Benchmark:   benchmark,
		PerformanceStats: model.StockRecommendationPerformanceSummary{
			SampleDays:                3,
			WinRate:                   1.0,
			AvgDailyReturn:            0.017,
			CumulativeReturn:          0.0518,
			BenchmarkCumulativeReturn: 0.0282,
			ExcessReturn:              0.0236,
			MaxDrawdown:               0,
			BenchmarkSymbol:           "000300.SH",
			BenchmarkSource:           "estimated: 55% of strategy daily return",
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (r *InMemoryGrowthRepo) GetStockRecommendationVersionHistory(userID string, recoID string) ([]model.StrategyVersionHistoryItem, error) {
	insight, err := r.GetStockRecommendationInsight(userID, recoID)
	if err != nil {
		return nil, err
	}
	return []model.StrategyVersionHistoryItem{
		{
			TradeDate:        "2026-02-25",
			PublishVersion:   1,
			CreatedAt:        insight.GeneratedAt,
			StrategyVersion:  firstNonEmpty(insight.Explanation.StrategyVersion, insight.Recommendation.StrategyVersion, "growth-v1"),
			ReasonSummary:    insight.Recommendation.ReasonSummary,
			ConfidenceReason: firstNonEmpty(insight.Explanation.ConfidenceReason, insight.Recommendation.ReasonSummary),
			ConsensusSummary: insight.Explanation.ConsensusSummary,
			RiskFlags:        insight.Explanation.RiskFlags,
			Invalidations:    insight.Explanation.Invalidations,
			GeneratedAt:      firstNonEmpty(insight.Explanation.GeneratedAt, insight.GeneratedAt),
		},
	}, nil
}

func (r *InMemoryGrowthRepo) ListFuturesStrategies(userID string, contract string, status string, page int, pageSize int) ([]model.FuturesStrategy, int, error) {
	items := []model.FuturesStrategy{
		{ID: "fs_001", Contract: "IF2603", Name: "股指趋势跟踪", Direction: "LONG", RiskLevel: "MEDIUM", PositionRange: "20%-30%", ValidFrom: "2026-02-25T09:00:00+08:00", ValidTo: "2026-02-26T15:00:00+08:00", Status: "PUBLISHED", ReasonSummary: "趋势与量价结构一致"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) GetFuturesStrategyDetail(userID string, strategyID string) (model.FuturesStrategy, error) {
	return model.FuturesStrategy{
		ID: strategyID, Contract: "IF2603", Name: "股指趋势跟踪", Direction: "LONG", RiskLevel: "MEDIUM",
		PositionRange: "20%-30%", ValidFrom: "2026-02-25T09:00:00+08:00", ValidTo: "2026-02-26T15:00:00+08:00",
		Status: "PUBLISHED", ReasonSummary: "趋势与量价结构一致",
	}, nil
}

func (r *InMemoryGrowthRepo) GetFuturesStrategyInsight(userID string, strategyID string) (model.FuturesStrategyInsight, error) {
	strategy, _ := r.GetFuturesStrategyDetail(userID, strategyID)
	guidance, _ := r.GetFuturesGuidance(strategy.Contract)
	performance := []model.RecommendationPerformancePoint{
		{Date: "2026-02-24", Return: 0.008},
		{Date: "2026-02-25", Return: 0.012},
		{Date: "2026-02-26", Return: -0.004},
		{Date: "2026-02-27", Return: 0.011},
	}
	benchmark := []model.RecommendationPerformancePoint{
		{Date: "2026-02-24", Return: 0.004},
		{Date: "2026-02-25", Return: 0.006},
		{Date: "2026-02-26", Return: -0.002},
		{Date: "2026-02-27", Return: 0.005},
	}
	return model.FuturesStrategyInsight{
		Strategy: strategy,
		Guidance: guidance,
		ScoreFramework: model.FuturesStrategyScoreFramework{
			Method:        "futures-v1 (trend25 + structure20 + flow15 + risk20 + news10 + performance10)",
			TotalScore:    84.6,
			WeightedScore: 83.9,
			ScoreGap:      0.7,
			Factors: []model.FuturesStrategyFactorScore{
				{Key: "trend", Label: "趋势因子", Weight: 0.25, Score: 86.0, Contribution: 21.5},
				{Key: "structure", Label: "结构因子", Weight: 0.20, Score: 82.0, Contribution: 16.4},
				{Key: "flow", Label: "资金因子", Weight: 0.15, Score: 80.0, Contribution: 12.0},
				{Key: "risk", Label: "风险控制", Weight: 0.20, Score: 84.0, Contribution: 16.8},
				{Key: "news", Label: "资讯因子", Weight: 0.10, Score: 81.0, Contribution: 8.1},
				{Key: "performance", Label: "绩效因子", Weight: 0.10, Score: 91.0, Contribution: 9.1},
			},
		},
		RelatedNews: []model.StockRecommendationRelatedNews{
			{
				ID:             "na_fut_demo_001",
				Title:          "股指期货主力合约成交放量，短线趋势延续",
				Summary:        "IF 主力合约成交量放大，盘中波动率回落，趋势信号维持。",
				Source:         "新闻快讯",
				Visibility:     "PUBLIC",
				PublishedAt:    "2026-02-27T10:22:00+08:00",
				RelevanceScore: 0.86,
			},
			{
				ID:             "na_fut_demo_002",
				Title:          "机构观点：股指期货套保盘活跃，风险偏好回升",
				Summary:        "套保盘与投机盘同步增加，短期偏多策略胜率提升。",
				Source:         "券商研报",
				Visibility:     "VIP",
				PublishedAt:    "2026-02-26T20:15:00+08:00",
				RelevanceScore: 0.79,
			},
		},
		RelatedEvents: []model.MarketEvent{
			{
				ID:          "me_fut_demo_001",
				EventType:   "PRICE",
				Symbol:      strategy.Contract,
				Summary:     "15分钟级别价格突破关键压力位。",
				TriggerRule: "15分钟涨幅超过1.2%",
				Source:      "system",
				CreatedAt:   "2026-02-27T10:16:00+08:00",
			},
		},
		Performance: performance,
		Benchmark:   benchmark,
		PerformanceStats: model.FuturesStrategyPerformanceSummary{
			SampleDays:                4,
			WinRate:                   0.75,
			AvgDailyReturn:            0.0068,
			CumulativeReturn:          0.0271,
			BenchmarkCumulativeReturn: 0.0131,
			ExcessReturn:              0.014,
			MaxDrawdown:               0.004,
			BenchmarkSymbol:           "000300.SH",
			BenchmarkSource:           "actual: CSI300",
		},
		GeneratedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (r *InMemoryGrowthRepo) GetFuturesStrategyVersionHistory(userID string, strategyID string) ([]model.StrategyVersionHistoryItem, error) {
	insight, err := r.GetFuturesStrategyInsight(userID, strategyID)
	if err != nil {
		return nil, err
	}
	return []model.StrategyVersionHistoryItem{
		{
			TradeDate:        "2026-02-25",
			PublishVersion:   1,
			CreatedAt:        insight.GeneratedAt,
			StrategyVersion:  firstNonEmpty(insight.Explanation.StrategyVersion, "futures-mvp-v1"),
			ReasonSummary:    insight.Strategy.ReasonSummary,
			ConfidenceReason: firstNonEmpty(insight.Explanation.ConfidenceReason, insight.Strategy.ReasonSummary),
			ConsensusSummary: insight.Explanation.ConsensusSummary,
			RiskFlags:        insight.Explanation.RiskFlags,
			Invalidations:    insight.Explanation.Invalidations,
			GeneratedAt:      firstNonEmpty(insight.Explanation.GeneratedAt, insight.GeneratedAt),
		},
	}, nil
}

func (r *InMemoryGrowthRepo) ListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error) {
	items := []model.MembershipProduct{
		{ID: "mp_demo_001", Name: "VIP月卡", Price: 99, Status: "ACTIVE", MemberLevel: "VIP1", DurationDays: 30},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) CreateMembershipOrder(userID string, productID string, payChannel string) (model.MembershipOrderAdmin, error) {
	return model.MembershipOrderAdmin{
		ID:         "mo_demo_002",
		OrderNo:    "mo_demo_002",
		UserID:     userID,
		ProductID:  productID,
		Amount:     99,
		PayChannel: payChannel,
		Status:     "PENDING",
		CreatedAt:  "2026-02-25T12:00:00+08:00",
	}, nil
}

func (r *InMemoryGrowthRepo) ListMembershipOrders(userID string, status string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error) {
	items := []model.MembershipOrderAdmin{
		{ID: "mo_demo_001", OrderNo: "mo_demo_001", UserID: userID, ProductID: "mp_demo_001", Amount: 99, PayChannel: "ALIPAY", Status: "PAID", PaidAt: "2026-02-24T11:00:00+08:00", CreatedAt: "2026-02-24T10:50:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) TrackExperimentEvent(item model.ExperimentEvent) error {
	return nil
}

func (r *InMemoryGrowthRepo) BindMembershipOrderExperiment(orderNo string, item model.ExperimentOrderAttribution) error {
	return nil
}

func (r *InMemoryGrowthRepo) GetRewardWallet(userID string) (model.RewardWallet, error) {
	return model.RewardWallet{
		CashBalance:    100.5,
		CashFrozen:     20,
		CouponBalance:  30,
		VIPDaysBalance: 15,
	}, nil
}

func (r *InMemoryGrowthRepo) ListRewardWalletTxns(userID string, page int, pageSize int) ([]model.RewardWalletTxn, int, error) {
	items := []model.RewardWalletTxn{
		{ID: "txn_001", TxnType: "ISSUE", Amount: 20, Status: "SUCCESS", CreatedAt: "2026-02-24T12:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) CreateWithdrawRequest(userID string, amount float64) (string, error) {
	return "wd_001", nil
}

func (r *InMemoryGrowthRepo) HandlePaymentCallback(channel string, orderNo string, channelTxnNo string, idempotencyKey string, sign string, signVerified bool) error {
	if idempotencyKey == "duplicate" {
		return errors.New("duplicate callback")
	}
	return nil
}

func (r *InMemoryGrowthRepo) ListArbitrageOpportunities(typeFilter string, page int, pageSize int) ([]model.ArbitrageOpportunity, int, error) {
	items := []model.ArbitrageOpportunity{
		{
			ID:         "arb_001",
			Type:       "CALENDAR",
			ContractA:  "RB2405",
			ContractB:  "RB2409",
			Spread:     110,
			Percentile: 0.86,
			RiskLevel:  "MEDIUM",
			Status:     "WATCH",
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) ListFuturesArbitrage(typeFilter string, page int, pageSize int) ([]model.ArbitrageRecommendation, int, error) {
	items := []model.ArbitrageRecommendation{
		{
			ID:         "arb_001",
			Type:       "CALENDAR",
			ContractA:  "RB2405",
			ContractB:  "RB2409",
			Spread:     110,
			Percentile: 0.86,
			EntryPoint: 120,
			ExitPoint:  80,
			StopPoint:  150,
			Status:     "WATCH",
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) GetFuturesArbitrageDetail(id string) (model.ArbitrageRecommendation, error) {
	return model.ArbitrageRecommendation{
		ID:          id,
		Type:        "CALENDAR",
		ContractA:   "RB2405",
		ContractB:   "RB2409",
		Spread:      110,
		Percentile:  0.86,
		EntryPoint:  120,
		ExitPoint:   80,
		StopPoint:   150,
		TriggerRule: "跨期价差回归",
		Status:      "WATCH",
	}, nil
}

func (r *InMemoryGrowthRepo) CreateFuturesAlert(userID string, contract string, alertType string, threshold float64) (string, error) {
	return "fa_demo_001", nil
}

func (r *InMemoryGrowthRepo) ListFuturesReviews(page int, pageSize int) ([]model.FuturesReview, int, error) {
	items := []model.FuturesReview{
		{ID: "fr_demo_001", StrategyID: "fs_demo_001", HitRate: 0.62, PnL: 1.8, MaxDrawdown: 0.12, ReviewDate: "2026-02-24T00:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) ListMarketEvents(eventType string, page int, pageSize int) ([]model.MarketEvent, int, error) {
	items := []model.MarketEvent{
		{ID: "me_demo_001", EventType: "PRICE", Symbol: "RB2405", Summary: "价格突破关键区间", TriggerRule: "突破120", CreatedAt: "2026-02-25T09:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) GetMarketEventDetail(id string) (model.MarketEvent, error) {
	return model.MarketEvent{
		ID:          id,
		EventType:   "PRICE",
		Symbol:      "RB2405",
		Summary:     "价格突破关键区间",
		TriggerRule: "突破120",
		Source:      "demo",
		CreatedAt:   "2026-02-25T09:00:00+08:00",
	}, nil
}

func (r *InMemoryGrowthRepo) GetFuturesGuidance(contract string) (model.FuturesGuidance, error) {
	return model.FuturesGuidance{
		Contract:          contract,
		GuidanceDirection: "LONG_SPREAD",
		PositionLevel:     "LIGHT",
		EntryRange:        "100-120",
		TakeProfitRange:   "60-80",
		StopLossRange:     "150-160",
		RiskLevel:         "MEDIUM",
		InvalidCondition:  "相关性失效",
		ValidTo:           "2026-03-01T15:00:00+08:00",
	}, nil
}

func (r *InMemoryGrowthRepo) ListPublicHoldings(symbol string, page int, pageSize int) ([]model.PublicHolding, int, error) {
	items := []model.PublicHolding{
		{ID: "ph_001", Holder: "某头部机构", Symbol: "600519.SH", Ratio: 2.356, DisclosedAt: "2026-02-24T18:00:00+08:00", Source: "交易所披露"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) ListPublicFuturesPositions(contract string, page int, pageSize int) ([]model.PublicFuturesPosition, int, error) {
	items := []model.PublicFuturesPosition{
		{ID: "pfp_001", Contract: "RB2405", LongPosition: 10234, ShortPosition: 9842, DisclosedAt: "2026-02-24T18:00:00+08:00", Source: "交易所持仓排名"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminListInviteRecords(status string, page int, pageSize int) ([]model.InviteRecord, int, error) {
	items := []model.InviteRecord{
		{ID: "iv_001", InviterUser: "u_1001", InviteeUser: "u_1002", Status: "FIRST_PAID", RiskFlag: "NORMAL"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminListRewardRecords(status string, page int, pageSize int) ([]model.RewardRecord, int, error) {
	items := []model.RewardRecord{
		{ID: "rw_001", InviterUser: "u_1001", InviteeUser: "u_1002", RewardType: "CASH", RewardValue: 20, Status: "PENDING"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminReviewRewardRecord(id string, status string, reason string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListReconciliation(page int, pageSize int) ([]model.ReconciliationRecord, int, error) {
	items := []model.ReconciliationRecord{
		{ID: "rec_001", PayChannel: "ALIPAY", BatchDate: "2026-02-24", Status: "DONE", DiffCount: 0},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminRetryReconciliation(batchID string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListRiskRules() ([]model.RiskRule, error) {
	items := []model.RiskRule{
		{ID: "rule_001", RuleCode: "DEVICE_DUP", RuleName: "同设备重复邀请", Threshold: 3, Status: "ACTIVE"},
	}
	return items, nil
}

func (r *InMemoryGrowthRepo) AdminCreateRiskRule(ruleCode string, ruleName string, threshold int, status string) (string, error) {
	return "rule_002", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateRiskRule(id string, threshold int, status string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListRiskHits(status string, page int, pageSize int) ([]model.RiskHit, int, error) {
	items := []model.RiskHit{
		{ID: "hit_001", RuleCode: "DEVICE_DUP", UserID: "u_1002", RiskLevel: "MEDIUM", Status: "PENDING_REVIEW"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminReviewRiskHit(id string, status string, reason string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListWithdrawRequests(page int, pageSize int) ([]model.WithdrawRequestInfo, int, error) {
	items := []model.WithdrawRequestInfo{
		{ID: "wd_001", UserID: "u_demo_001", Amount: 50, Status: "PENDING", AppliedAt: "2026-02-24T12:10:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminReviewWithdrawRequest(id string, status string, reason string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListNewsCategories(status string, page int, pageSize int) ([]model.NewsCategory, int, error) {
	items, _ := r.ListNewsCategories("u_demo_001")
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminCreateNewsCategory(name string, slug string, sort int, visibility string, status string) (string, error) {
	return "nc_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateNewsCategory(id string, name string, slug string, sort int, visibility string, status string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListNewsArticles(status string, categoryID string, page int, pageSize int) ([]model.NewsArticle, int, error) {
	items, total, _ := r.ListNewsArticles("u_demo_001", categoryID, "", page, pageSize)
	return items, total, nil
}

func (r *InMemoryGrowthRepo) AdminGetNewsArticleDetail(id string) (model.NewsArticle, error) {
	return r.GetNewsArticleDetail("u_demo_001", id)
}

func (r *InMemoryGrowthRepo) AdminCreateNewsArticle(categoryID string, title string, summary string, content string, coverURL string, visibility string, status string, authorID string) (string, error) {
	return "na_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateNewsArticle(id string, categoryID string, title string, summary string, content string, coverURL string, visibility string, status string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminPublishNewsArticle(id string, status string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminCreateNewsAttachment(articleID string, fileName string, fileURL string, fileSize int64, mimeType string) (string, error) {
	return "att_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminListNewsAttachments(articleID string) ([]model.NewsAttachment, error) {
	return []model.NewsAttachment{
		{ID: "att_001", ArticleID: articleID, FileName: "daily-note.pdf", FileURL: "https://example.com/files/daily-note.pdf", FileSize: 20480, MimeType: "application/pdf", CreatedAt: "2026-02-25T10:00:00+08:00"},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminDeleteNewsAttachment(id string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListMarketEvents(eventType string, symbol string, page int, pageSize int) ([]model.MarketEvent, int, error) {
	items := []model.MarketEvent{
		{
			ID:          "me_demo_001",
			EventType:   "PRICE",
			Symbol:      "IF2603",
			Summary:     "指数期货早盘快速拉升，触发突破提醒",
			TriggerRule: "15分钟涨幅超过1.2%",
			Source:      "system",
			CreatedAt:   "2026-02-26T09:40:00+08:00",
		},
		{
			ID:          "me_demo_002",
			EventType:   "POLICY",
			Symbol:      "ALL",
			Summary:     "交易所发布保证金调整通知",
			TriggerRule: "保证金比率上调",
			Source:      "exchange",
			CreatedAt:   "2026-02-26T08:20:00+08:00",
		},
	}

	trimmedType := strings.ToUpper(strings.TrimSpace(eventType))
	trimmedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	filtered := make([]model.MarketEvent, 0, len(items))
	for _, item := range items {
		if trimmedType != "" && strings.ToUpper(item.EventType) != trimmedType {
			continue
		}
		if trimmedSymbol != "" && !strings.Contains(strings.ToUpper(item.Symbol), trimmedSymbol) {
			continue
		}
		filtered = append(filtered, item)
	}

	return filtered, len(filtered), nil
}

func (r *InMemoryGrowthRepo) AdminCreateMarketEvent(item model.MarketEvent) (string, error) {
	return "me_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateMarketEvent(id string, item model.MarketEvent) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListStockEventClusters(query model.StockEventQuery) ([]model.StockEventCluster, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	query = normalizeStockEventQuery(query)
	items := make([]model.StockEventCluster, 0, len(r.stockEventClusters))
	for _, item := range r.stockEventClusters {
		if query.ReviewStatus != "" && item.ReviewStatus != query.ReviewStatus {
			continue
		}
		if query.EventType != "" && item.EventType != query.EventType {
			continue
		}
		if query.ReviewPriority != "" {
			priority, _ := item.Metadata["review_priority"].(string)
			if strings.ToUpper(strings.TrimSpace(priority)) != query.ReviewPriority {
				continue
			}
		}
		if query.Symbol != "" && item.PrimarySymbol != query.Symbol {
			continue
		}
		if query.Sector != "" && item.SectorLabel != query.Sector {
			continue
		}
		if query.Topic != "" && item.TopicLabel != query.Topic {
			continue
		}
		items = append(items, item)
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminGetStockEventCluster(id string) (model.StockEventCluster, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.stockEventClusters[strings.TrimSpace(id)]
	if !ok {
		return model.StockEventCluster{}, sql.ErrNoRows
	}
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminUpsertStockEventCluster(cluster model.StockEventCluster) (model.StockEventCluster, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	cluster = normalizeStockEventCluster(cluster)
	now := time.Now().Format(time.RFC3339)
	if cluster.CreatedAt == "" {
		cluster.CreatedAt = now
	}
	cluster.UpdatedAt = now
	for index := range cluster.Items {
		cluster.Items[index] = normalizeStockEventItem(cluster.Items[index], cluster.ID)
		cluster.Items[index].CreatedAt = now
		cluster.Items[index].UpdatedAt = now
	}
	for index := range cluster.Entities {
		cluster.Entities[index] = normalizeStockEventEntity(cluster.Entities[index], cluster.ID)
		cluster.Entities[index].CreatedAt = now
		cluster.Entities[index].UpdatedAt = now
	}
	for index := range cluster.Edges {
		cluster.Edges[index] = normalizeStockEventEdge(cluster.Edges[index], cluster.ID)
		cluster.Edges[index].CreatedAt = now
		cluster.Edges[index].UpdatedAt = now
	}
	r.stockEventClusters[cluster.ID] = cluster
	r.ensureHighPriorityStockEventTaskLocked(cluster)
	return cluster, nil
}

func (r *InMemoryGrowthRepo) AdminCreateStockEventReview(review model.StockEventReview) (model.StockEventReview, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	review = normalizeStockEventReview(review)
	cluster, ok := r.stockEventClusters[review.ClusterID]
	if !ok {
		return model.StockEventReview{}, sql.ErrNoRows
	}
	now := time.Now().Format(time.RFC3339)
	review.ReviewedAt = now
	review.CreatedAt = now
	review.UpdatedAt = now
	cluster.Status = stockEventClusterStatusFromReview(review.ReviewStatus)
	cluster.ReviewStatus = review.ReviewStatus
	cluster.LatestReview = &review
	cluster.UpdatedAt = now
	r.stockEventClusters[cluster.ID] = cluster
	return review, nil
}

func (r *InMemoryGrowthRepo) AdminListMarketRhythmTasks(taskDate string) ([]model.MarketRhythmTask, error) {
	dateText, err := normalizeMarketRhythmDate(taskDate)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	items := r.ensureInMemoryMarketRhythmTasksLocked(dateText)
	return cloneMarketRhythmTasks(items), nil
}

func (r *InMemoryGrowthRepo) AdminEnsureMarketRhythmTasks(taskDate string) ([]model.MarketRhythmTask, error) {
	dateText, err := normalizeMarketRhythmDate(taskDate)
	if err != nil {
		return nil, err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	items := r.ensureInMemoryMarketRhythmTasksLocked(dateText)
	return cloneMarketRhythmTasks(items), nil
}

func (r *InMemoryGrowthRepo) AdminUpdateMarketRhythmTask(id string, owner string, notes string, sourceLinks []string, status string) (model.MarketRhythmTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.marketRhythmTasks[id]
	if !ok {
		return model.MarketRhythmTask{}, sql.ErrNoRows
	}
	item.Owner = strings.TrimSpace(owner)
	item.Notes = strings.TrimSpace(notes)
	item.SourceLinks = normalizeStringList(sourceLinks)
	if normalizedStatus := normalizeMarketRhythmStatus(status); normalizedStatus != "" {
		item.Status = normalizedStatus
	}
	if item.Status == "DONE" {
		item.CompletedAt = time.Now().Format(time.RFC3339)
	} else {
		item.CompletedAt = ""
	}
	item.UpdatedAt = time.Now().Format(time.RFC3339)
	r.marketRhythmTasks[id] = item
	return item, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateMarketRhythmTaskStatus(id string, status string, owner string, notes string) (model.MarketRhythmTask, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, ok := r.marketRhythmTasks[id]
	if !ok {
		return model.MarketRhythmTask{}, sql.ErrNoRows
	}
	if normalizedStatus := normalizeMarketRhythmStatus(status); normalizedStatus != "" {
		item.Status = normalizedStatus
	}
	if strings.TrimSpace(owner) != "" {
		item.Owner = strings.TrimSpace(owner)
	}
	if strings.TrimSpace(notes) != "" {
		item.Notes = strings.TrimSpace(notes)
	}
	if item.Status == "DONE" {
		item.CompletedAt = time.Now().Format(time.RFC3339)
	} else {
		item.CompletedAt = ""
	}
	item.UpdatedAt = time.Now().Format(time.RFC3339)
	r.marketRhythmTasks[id] = item
	return item, nil
}

func cloneMarketRhythmTasks(items []model.MarketRhythmTask) []model.MarketRhythmTask {
	result := make([]model.MarketRhythmTask, 0, len(items))
	for _, item := range items {
		cloned := item
		if len(item.SourceLinks) > 0 {
			cloned.SourceLinks = append([]string(nil), item.SourceLinks...)
		}
		result = append(result, cloned)
	}
	return result
}

func (r *InMemoryGrowthRepo) ensureInMemoryMarketRhythmTasksLocked(taskDate string) []model.MarketRhythmTask {
	now := time.Now().Format(time.RFC3339)
	items := make([]model.MarketRhythmTask, 0, len(defaultMarketRhythmTaskTemplates()))
	for _, template := range defaultMarketRhythmTaskTemplates() {
		id := fmt.Sprintf("mrt_%s_%s", strings.ReplaceAll(taskDate, "-", ""), template.TaskKey)
		item, ok := r.marketRhythmTasks[id]
		if !ok {
			item = model.MarketRhythmTask{
				ID:          id,
				Date:        taskDate,
				Slot:        template.Slot,
				TaskKey:     template.TaskKey,
				Status:      "TODO",
				Owner:       "",
				Notes:       "",
				SourceLinks: nil,
				CompletedAt: "",
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			r.marketRhythmTasks[id] = item
		}
		items = append(items, item)
	}
	return items
}

func (r *InMemoryGrowthRepo) AdminListStockRecommendations(status string, page int, pageSize int) ([]model.StockRecommendation, int, error) {
	items, total, _ := r.ListStockRecommendations("u_demo_001", "", page, pageSize)
	return items, total, nil
}

func (r *InMemoryGrowthRepo) AdminCreateStockRecommendation(item model.StockRecommendation) (string, error) {
	return "sr_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateStockRecommendationStatus(id string, status string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminSyncStockInstrumentMaster(sourceKey string, symbols []string) (model.MarketSyncResult, error) {
	symbols = normalizeStockSymbolList(symbols)
	if len(symbols) == 0 {
		symbols = []string{
			"600519.SH", "601318.SH", "600036.SH", "600276.SH", "601012.SH",
			"000333.SZ", "300750.SZ", "002594.SZ", "688981.SH", "601888.SH",
			"000858.SZ", "000001.SZ", "601166.SH", "300015.SZ", "000651.SZ",
		}
	}
	sourceKey = strings.ToUpper(strings.TrimSpace(sourceKey))
	if sourceKey == "" {
		sourceKey = "TUSHARE"
	}
	count := len(symbols)
	return model.MarketSyncResult{
		AssetClass:         "STOCK",
		DataKind:           "INSTRUMENT_MASTER",
		RequestedSourceKey: sourceKey,
		ResolvedSourceKeys: []string{sourceKey},
		TruthCount:         count,
		Results: []model.MarketSourceSyncItemResult{
			{
				SourceKey:     sourceKey,
				Status:        "SUCCESS",
				TruthCount:    count,
				Message:       "in-memory stock instrument sync",
				SnapshotCount: 1,
			},
		},
		SnapshotCount: 1,
	}, nil
}

func (r *InMemoryGrowthRepo) AdminSyncStockQuotes(sourceKey string, symbols []string, days int) (int, error) {
	result, err := r.AdminSyncStockQuotesDetailed(sourceKey, symbols, days)
	if err != nil {
		return 0, err
	}
	if result.TruthCount > 0 {
		return result.TruthCount, nil
	}
	return result.BarCount, nil
}

func (r *InMemoryGrowthRepo) AdminSyncStockQuotesDetailed(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error) {
	if days <= 0 {
		days = 90
	}
	if len(symbols) == 0 {
		symbols = []string{"600519.SH", "601318.SH", "600036.SH", "300750.SZ", "000333.SZ"}
	}
	sourceKey = strings.ToUpper(strings.TrimSpace(sourceKey))
	if sourceKey == "" {
		sourceKey = "TUSHARE"
	}
	count := len(symbols) * days
	return model.MarketSyncResult{
		AssetClass:         "STOCK",
		DataKind:           "DAILY_BARS",
		RequestedSourceKey: sourceKey,
		ResolvedSourceKeys: []string{sourceKey},
		BarCount:           count,
		TruthCount:         count,
		SnapshotCount:      1,
		Results: []model.MarketSourceSyncItemResult{
			{
				SourceKey:     sourceKey,
				Status:        "SUCCESS",
				BarCount:      count,
				TruthCount:    count,
				SnapshotCount: 1,
				Message:       "in-memory market sync",
			},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminSyncStockQuotesFromMaster(sourceKey string, days int) (model.MarketSyncResult, error) {
	if days <= 0 {
		days = 90
	}
	symbols := []string{
		"000001.SZ", "000333.SZ", "002594.SZ", "300750.SZ",
		"600000.SH", "600036.SH", "600519.SH", "601318.SH",
		"601398.SH", "601899.SH", "688111.SH", "688981.SH",
	}
	sourceKey = strings.ToUpper(strings.TrimSpace(sourceKey))
	if sourceKey == "" {
		sourceKey = "TUSHARE"
	}
	count := len(symbols) * days
	return model.MarketSyncResult{
		AssetClass:         "STOCK",
		DataKind:           "DAILY_BARS",
		RequestedSourceKey: sourceKey,
		ResolvedSourceKeys: []string{sourceKey},
		BarCount:           count,
		TruthCount:         count,
		SnapshotCount:      1,
		Results: []model.MarketSourceSyncItemResult{
			{
				SourceKey:     sourceKey,
				Status:        "SUCCESS",
				BarCount:      count,
				TruthCount:    count,
				SnapshotCount: 1,
				Message:       "in-memory full-market stock sync",
			},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminSyncStockDailyBasics(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error) {
	if days <= 0 {
		days = 30
	}
	if len(symbols) == 0 {
		symbols = []string{
			"000001.SZ", "000333.SZ", "002594.SZ", "300750.SZ",
			"600000.SH", "600036.SH", "600519.SH", "601318.SH",
			"601398.SH", "601899.SH", "688111.SH", "688981.SH",
		}
	}
	sourceKey = strings.ToUpper(strings.TrimSpace(sourceKey))
	if sourceKey == "" {
		sourceKey = "TUSHARE"
	}
	count := len(symbols) * days
	return model.MarketSyncResult{
		AssetClass:         "STOCK",
		DataKind:           "STOCK_DAILY_BASIC",
		RequestedSourceKey: sourceKey,
		ResolvedSourceKeys: []string{sourceKey},
		TruthCount:         count,
		SnapshotCount:      1,
		Results: []model.MarketSourceSyncItemResult{{
			SourceKey:     sourceKey,
			Status:        "SUCCESS",
			TruthCount:    count,
			SnapshotCount: 1,
			Message:       "in-memory stock daily basic sync",
		}},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminSyncStockMoneyflows(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error) {
	if days <= 0 {
		days = 30
	}
	if len(symbols) == 0 {
		symbols = []string{
			"000001.SZ", "000333.SZ", "002594.SZ", "300750.SZ",
			"600000.SH", "600036.SH", "600519.SH", "601318.SH",
			"601398.SH", "601899.SH", "688111.SH", "688981.SH",
		}
	}
	sourceKey = strings.ToUpper(strings.TrimSpace(sourceKey))
	if sourceKey == "" {
		sourceKey = "TUSHARE"
	}
	count := len(symbols) * days
	return model.MarketSyncResult{
		AssetClass:         "STOCK",
		DataKind:           "STOCK_MONEYFLOW",
		RequestedSourceKey: sourceKey,
		ResolvedSourceKeys: []string{sourceKey},
		TruthCount:         count,
		SnapshotCount:      1,
		Results: []model.MarketSourceSyncItemResult{{
			SourceKey:     sourceKey,
			Status:        "SUCCESS",
			TruthCount:    count,
			SnapshotCount: 1,
			Message:       "in-memory stock moneyflow sync",
		}},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminSyncStockNewsRaw(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error) {
	if days <= 0 {
		days = 7
	}
	if len(symbols) == 0 {
		symbols = []string{
			"600519.SH", "601318.SH", "300750.SZ", "000333.SZ",
		}
	}
	sourceKey = strings.ToUpper(strings.TrimSpace(sourceKey))
	if sourceKey == "" {
		sourceKey = "TUSHARE"
	}
	count := len(symbols) * 2
	return model.MarketSyncResult{
		AssetClass:         "STOCK",
		DataKind:           "STOCK_NEWS_RAW",
		RequestedSourceKey: sourceKey,
		ResolvedSourceKeys: []string{sourceKey},
		NewsCount:          count,
		SnapshotCount:      1,
		Results: []model.MarketSourceSyncItemResult{{
			SourceKey:     sourceKey,
			Status:        "SUCCESS",
			NewsCount:     count,
			SnapshotCount: 1,
			Message:       "in-memory stock news sync",
		}},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminSyncFuturesQuotes(sourceKey string, contracts []string, days int) (model.MarketSyncResult, error) {
	if days <= 0 {
		days = 60
	}
	if len(contracts) == 0 {
		contracts = []string{"AU2406.SHF", "AG2406.SHF", "IF2406.CFX"}
	}
	sourceKey = strings.ToUpper(strings.TrimSpace(sourceKey))
	if sourceKey == "" {
		sourceKey = "TUSHARE"
	}
	count := len(contracts) * days
	return model.MarketSyncResult{
		AssetClass:         "FUTURES",
		DataKind:           "DAILY_BARS",
		RequestedSourceKey: sourceKey,
		ResolvedSourceKeys: []string{sourceKey},
		BarCount:           count,
		TruthCount:         count,
		SnapshotCount:      1,
		Results: []model.MarketSourceSyncItemResult{
			{
				SourceKey:     sourceKey,
				Status:        "SUCCESS",
				BarCount:      count,
				TruthCount:    count,
				SnapshotCount: 1,
				Message:       "in-memory futures sync",
			},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminSyncFuturesInventory(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error) {
	if days <= 0 {
		days = 30
	}
	if len(symbols) == 0 {
		symbols = []string{"RB", "CU", "AU"}
	}
	sourceKey = strings.ToUpper(strings.TrimSpace(sourceKey))
	if sourceKey == "" {
		sourceKey = "MOCK"
	}
	count := len(symbols) * minInt(days, 10)
	return model.MarketSyncResult{
		AssetClass:         "FUTURES",
		DataKind:           "FUTURES_INVENTORY",
		RequestedSourceKey: sourceKey,
		ResolvedSourceKeys: []string{sourceKey},
		InventoryCount:     count,
		SnapshotCount:      1,
		Results: []model.MarketSourceSyncItemResult{
			{
				SourceKey:      sourceKey,
				Status:         "SUCCESS",
				InventoryCount: count,
				SnapshotCount:  1,
				Message:        "in-memory futures inventory sync",
			},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminSyncMarketNews(sourceKey string, symbols []string, days int, limit int) (model.MarketSyncResult, error) {
	if limit <= 0 {
		limit = 20
	}
	sourceKey = strings.ToUpper(strings.TrimSpace(sourceKey))
	if sourceKey == "" {
		sourceKey = "AKSHARE"
	}
	return model.MarketSyncResult{
		DataKind:           "NEWS_ITEMS",
		RequestedSourceKey: sourceKey,
		ResolvedSourceKeys: []string{sourceKey},
		NewsCount:          limit,
		SnapshotCount:      1,
		Results: []model.MarketSourceSyncItemResult{
			{
				SourceKey:     sourceKey,
				Status:        "SUCCESS",
				NewsCount:     limit,
				SnapshotCount: 1,
				Message:       "in-memory news sync",
			},
		},
	}, nil
}

func (r *InMemoryGrowthRepo) BuildStrategyEngineStockSelectionContext(input model.StrategyEngineStockSelectionContextRequest) (model.StrategyEngineStockSelectionContextResponse, error) {
	return model.StrategyEngineStockSelectionContextResponse{}, errors.New("strategy-engine stock context is unavailable in memory mode")
}

func (r *InMemoryGrowthRepo) BuildStrategyEngineFuturesStrategyContext(input model.StrategyEngineFuturesStrategyContextRequest) (model.StrategyEngineFuturesStrategyContextResponse, error) {
	return model.StrategyEngineFuturesStrategyContextResponse{}, errors.New("strategy-engine futures context is unavailable in memory mode")
}

func (r *InMemoryGrowthRepo) AdminSyncDocFastNewsIncremental(batchSize int) (string, error) {
	if batchSize <= 0 {
		batchSize = 200
	}
	return "doc_fast incremental sync skipped (in-memory repo)", nil
}

func (r *InMemoryGrowthRepo) AdminSyncTushareNewsIncremental(batchSize int) (string, error) {
	if batchSize <= 0 {
		batchSize = 200
	}
	return "tushare news incremental sync skipped (in-memory repo)", nil
}

func (r *InMemoryGrowthRepo) AdminSyncTushareNewsIncrementalWithOptions(opts model.TushareNewsSyncOptions) (string, []model.NewsSyncRunDetail, error) {
	batchSize := opts.BatchSize
	if batchSize <= 0 {
		batchSize = 200
	}
	details := []model.NewsSyncRunDetail{
		{
			ID:            "nsd_demo_001",
			JobName:       "tushare_news_incremental",
			SyncType:      "NEWS_BRIEF",
			Source:        "cls",
			Symbol:        "",
			Status:        "SUCCESS",
			FetchedCount:  0,
			UpsertedCount: 0,
			FailedCount:   0,
			StartedAt:     "2026-02-28T00:00:00+08:00",
			FinishedAt:    "2026-02-28T00:00:00+08:00",
		},
	}
	return fmt.Sprintf("tushare news incremental sync skipped (in-memory repo, batch=%d)", batchSize), details, nil
}

func (r *InMemoryGrowthRepo) AdminRunVIPMembershipLifecycle() (string, error) {
	return "vip lifecycle skipped (in-memory repo)", nil
}

func (r *InMemoryGrowthRepo) AdminGetQuantTopStocks(limit int, lookbackDays int) ([]model.StockQuantScore, error) {
	items := []model.StockQuantScore{
		{
			Rank: 1, Symbol: "600519.SH", Name: "贵州茅台", TradeDate: "2026-02-27", ClosePrice: 1728.5,
			Momentum5: 3.2, Momentum20: 8.6, Volatility20: 1.9, VolumeRatio: 1.28, Drawdown20: 3.1, TrendStrength: 2.4,
			TrendScore: 91.2, FlowScore: 89.4, ValueScore: 86.8, NewsScore: 78.0,
			NetMFAmount: 18234.5, PeTTM: 26.8, PB: 9.6, TurnoverRate: 0.86, NewsHeat: 4, PositiveNewsRate: 0.75,
			Score: 92.6, RiskLevel: "LOW", ReasonSummary: "20日动量8.60%，主力净流入18234.50", Reasons: []string{"20日动量8.60%，中期趋势较强", "主力净流入18234.50，资金面偏强", "近14天资讯热度4，正面占比75%"},
		},
		{
			Rank: 2, Symbol: "601318.SH", Name: "中国平安", TradeDate: "2026-02-27", ClosePrice: 49.6,
			Momentum5: 2.8, Momentum20: 7.9, Volatility20: 2.1, VolumeRatio: 1.35, Drawdown20: 4.0, TrendStrength: 2.0,
			TrendScore: 89.6, FlowScore: 86.5, ValueScore: 88.2, NewsScore: 73.1,
			NetMFAmount: 10453.8, PeTTM: 10.5, PB: 1.2, TurnoverRate: 1.05, NewsHeat: 3, PositiveNewsRate: 0.67,
			Score: 90.4, RiskLevel: "LOW", ReasonSummary: "20日动量7.90%，估值区间合理", Reasons: []string{"20日动量7.90%，中期趋势较强", "PE(TTM) 10.50，估值处于可接受区间", "波动与回撤控制在可接受范围"},
		},
		{
			Rank: 3, Symbol: "600036.SH", Name: "招商银行", TradeDate: "2026-02-27", ClosePrice: 36.2,
			Momentum5: 2.1, Momentum20: 6.8, Volatility20: 2.0, VolumeRatio: 1.22, Drawdown20: 3.6, TrendStrength: 1.8,
			TrendScore: 87.1, FlowScore: 83.4, ValueScore: 85.9, NewsScore: 69.5,
			NetMFAmount: 7234.2, PeTTM: 8.9, PB: 0.9, TurnoverRate: 0.96, NewsHeat: 2, PositiveNewsRate: 0.50,
			Score: 88.7, RiskLevel: "LOW", ReasonSummary: "趋势稳健，估值与资金表现平衡", Reasons: []string{"20日动量6.80%，中期趋势较强", "主力净流入7234.20，资金面偏强", "PE(TTM) 8.90，估值处于可接受区间"},
		},
		{
			Rank: 4, Symbol: "300750.SZ", Name: "宁德时代", TradeDate: "2026-02-27", ClosePrice: 211.8,
			Momentum5: 3.9, Momentum20: 9.5, Volatility20: 2.8, VolumeRatio: 1.54, Drawdown20: 6.4, TrendStrength: 2.9,
			TrendScore: 92.3, FlowScore: 90.8, ValueScore: 66.2, NewsScore: 71.4,
			NetMFAmount: 21230.0, PeTTM: 38.2, PB: 7.6, TurnoverRate: 1.48, NewsHeat: 5, PositiveNewsRate: 0.60,
			Score: 87.2, RiskLevel: "MEDIUM", ReasonSummary: "动量和资金强，但估值偏高", Reasons: []string{"20日动量9.50%，中期趋势较强", "主力净流入21230.00，资金面偏强", "PE(TTM) 38.20，估值偏高需跟踪兑现"},
		},
		{
			Rank: 5, Symbol: "000333.SZ", Name: "美的集团", TradeDate: "2026-02-27", ClosePrice: 71.4,
			Momentum5: 1.9, Momentum20: 6.1, Volatility20: 1.7, VolumeRatio: 1.11, Drawdown20: 3.5, TrendStrength: 1.6,
			TrendScore: 84.2, FlowScore: 80.3, ValueScore: 83.4, NewsScore: 67.8,
			NetMFAmount: 5312.7, PeTTM: 13.2, PB: 2.7, TurnoverRate: 0.82, NewsHeat: 2, PositiveNewsRate: 0.50,
			Score: 86.8, RiskLevel: "LOW", ReasonSummary: "趋势延续，波动收敛", Reasons: []string{"20日动量6.10%，中期趋势较强", "PE(TTM) 13.20，估值处于可接受区间", "波动与回撤控制在可接受范围"},
		},
	}
	if limit <= 0 || limit > len(items) {
		limit = len(items)
	}
	return items[:limit], nil
}

func (r *InMemoryGrowthRepo) AdminGetQuantEvaluation(windowDays int, topN int) (model.StockQuantEvaluationSummary, []model.StockQuantEvaluationPoint, []model.StockQuantRiskPerformance, []model.StockQuantRotationPoint, error) {
	if windowDays <= 0 {
		windowDays = 60
	}
	if topN <= 0 {
		topN = 10
	}
	points := []model.StockQuantEvaluationPoint{
		{
			TradeDate: "2026-02-21", SampleCount: 10,
			AvgReturn5: 0.021, HitRate5: 0.70, BenchmarkReturn: 0.012,
			AvgReturn10: 0.036, HitRate10: 0.68, BenchmarkReturn10: 0.019,
			CumulativeReturn5: 0.021, CumulativeBenchmark5: 0.012, CumulativeExcess5: 0.009,
			CumulativeReturn10: 0.036, CumulativeBenchmark10: 0.019, CumulativeExcess10: 0.017,
		},
		{
			TradeDate: "2026-02-24", SampleCount: 10,
			AvgReturn5: 0.013, HitRate5: 0.60, BenchmarkReturn: 0.009,
			AvgReturn10: 0.025, HitRate10: 0.61, BenchmarkReturn10: 0.014,
			CumulativeReturn5: 0.034, CumulativeBenchmark5: 0.021, CumulativeExcess5: 0.013,
			CumulativeReturn10: 0.061, CumulativeBenchmark10: 0.033, CumulativeExcess10: 0.028,
		},
		{
			TradeDate: "2026-02-26", SampleCount: 9,
			AvgReturn5: 0.017, HitRate5: 0.67, BenchmarkReturn: 0.010,
			AvgReturn10: 0.031, HitRate10: 0.64, BenchmarkReturn10: 0.015,
			CumulativeReturn5: 0.052, CumulativeBenchmark5: 0.031, CumulativeExcess5: 0.021,
			CumulativeReturn10: 0.094, CumulativeBenchmark10: 0.048, CumulativeExcess10: 0.046,
		},
	}
	summary := model.StockQuantEvaluationSummary{
		WindowDays:           windowDays,
		TopN:                 topN,
		SampleDays:           len(points),
		SampleCount:          29,
		AvgReturn5:           0.017,
		HitRate5:             0.6567,
		MaxDrawdown5:         0.024,
		AvgReturn10:          0.0307,
		HitRate10:            0.6433,
		MaxDrawdown10:        0.031,
		BenchmarkAvgReturn5:  0.0103,
		BenchmarkAvgReturn10: 0.0162,
		GeneratedAt:          "2026-02-28T10:00:00+08:00",
	}
	riskItems := []model.StockQuantRiskPerformance{
		{RiskLevel: "LOW", SampleCount: 18, AvgReturn5: 0.018, HitRate5: 0.72, AvgReturn10: 0.032, HitRate10: 0.67},
		{RiskLevel: "MEDIUM", SampleCount: 11, AvgReturn5: 0.014, HitRate5: 0.55, AvgReturn10: 0.028, HitRate10: 0.59},
		{RiskLevel: "HIGH", SampleCount: 0, AvgReturn5: 0, HitRate5: 0, AvgReturn10: 0, HitRate10: 0},
	}
	rotationItems := []model.StockQuantRotationPoint{
		{
			TradeDate: "2026-02-21", TopSymbols: []string{"600519.SH", "601318.SH", "600036.SH", "300750.SZ", "000333.SZ"},
			Entered: []string{"600519.SH", "601318.SH", "600036.SH", "300750.SZ", "000333.SZ"}, Exited: []string{}, StayedCount: 0, ChangedCount: 5,
		},
		{
			TradeDate: "2026-02-24", TopSymbols: []string{"600519.SH", "601318.SH", "300750.SZ", "000333.SZ", "688981.SH"},
			Entered: []string{"688981.SH"}, Exited: []string{"600036.SH"}, StayedCount: 4, ChangedCount: 2,
		},
		{
			TradeDate: "2026-02-26", TopSymbols: []string{"600519.SH", "601318.SH", "600036.SH", "688981.SH", "002594.SZ"},
			Entered: []string{"600036.SH", "002594.SZ"}, Exited: []string{"300750.SZ", "000333.SZ"}, StayedCount: 3, ChangedCount: 4,
		},
	}
	return summary, points, riskItems, rotationItems, nil
}

func (r *InMemoryGrowthRepo) AdminGenerateDailyStockRecommendations(tradeDate string) (model.AdminDailyStockRecommendationGenerationResult, error) {
	return model.AdminDailyStockRecommendationGenerationResult{Count: 10}, nil
}

func (r *InMemoryGrowthRepo) AdminListStrategyEnginePublishHistory(jobType string) ([]model.StrategyEnginePublishRecordSummary, error) {
	return []model.StrategyEnginePublishRecordSummary{}, nil
}

func (r *InMemoryGrowthRepo) AdminGetStrategyEnginePublishRecord(publishID string) (model.StrategyEnginePublishRecord, error) {
	return model.StrategyEnginePublishRecord{}, nil
}

func (r *InMemoryGrowthRepo) AdminGetStrategyEnginePublishReplay(publishID string) (model.StrategyEnginePublishReplay, error) {
	return model.StrategyEnginePublishReplay{}, nil
}

func (r *InMemoryGrowthRepo) AdminCompareStrategyEnginePublishVersions(leftPublishID string, rightPublishID string) (model.StrategyEnginePublishCompareResult, error) {
	return model.StrategyEnginePublishCompareResult{}, nil
}

func (r *InMemoryGrowthRepo) AdminGenerateDailyFuturesStrategies(tradeDate string) (model.AdminDailyFuturesStrategyGenerationResult, error) {
	return model.AdminDailyFuturesStrategyGenerationResult{Count: 3}, nil
}

func (r *InMemoryGrowthRepo) AdminListFuturesStrategies(status string, contract string, page int, pageSize int) ([]model.FuturesStrategy, int, error) {
	items, total, _ := r.ListFuturesStrategies("u_demo_001", contract, status, page, pageSize)
	return items, total, nil
}

func (r *InMemoryGrowthRepo) AdminCreateFuturesStrategy(item model.FuturesStrategy) (string, error) {
	return "fs_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateFuturesStrategyStatus(id string, status string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListUsers(status string, kycStatus string, memberLevel string, registrationSource string, page int, pageSize int) ([]model.AdminUser, int, error) {
	registrationSource = strings.ToUpper(strings.TrimSpace(registrationSource))
	items := []model.AdminUser{
		{
			ID:                 "u_demo_001",
			Phone:              "13800000001",
			Email:              "demo@sercherai.local",
			Status:             "ACTIVE",
			KYCStatus:          "PENDING",
			MemberLevel:        "VIP1",
			ActivationState:    "PAID_PENDING_KYC",
			RegistrationSource: "INVITED",
			InviterUserID:      "u_demo_inviter",
			InviteCode:         "DEMO2026",
			InviteRegisteredAt: "2026-02-20T10:00:00+08:00",
			CreatedAt:          "2026-02-20T10:00:00+08:00",
		},
		{
			ID:                 "admin_001",
			Phone:              "13800000000",
			Email:              "admin@sercherai.local",
			Status:             "ACTIVE",
			KYCStatus:          "APPROVED",
			MemberLevel:        "VIP1",
			ActivationState:    "ACTIVE",
			RegistrationSource: "DIRECT",
			CreatedAt:          "2026-02-20T10:00:00+08:00",
		},
	}
	filtered := make([]model.AdminUser, 0, len(items))
	for _, item := range items {
		if registrationSource == "INVITED" && strings.ToUpper(item.RegistrationSource) != "INVITED" {
			continue
		}
		if registrationSource == "DIRECT" && strings.ToUpper(item.RegistrationSource) != "DIRECT" {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, len(filtered), nil
}

func (r *InMemoryGrowthRepo) AdminGetUserSourceSummary(status string, kycStatus string, memberLevel string, registrationSource string) (model.AdminUserSourceSummary, error) {
	summary := model.AdminUserSourceSummary{
		TotalUsers:            2,
		DirectUsers:           1,
		InvitedUsers:          1,
		InviteRate:            0.5,
		TodayInvitedUsers:     0,
		Last7dInvitedUsers:    1,
		Last7dFirstPaidUsers:  0,
		Last7dConversionRate:  0,
		Last30dInvitedUsers:   1,
		Last30dFirstPaidUsers: 0,
		Last30dConversionRate: 0,
		TotalFirstPaidUsers:   0,
		TotalConversionRate:   0,
	}
	registrationSource = strings.ToUpper(strings.TrimSpace(registrationSource))
	if registrationSource == "DIRECT" {
		summary.InvitedUsers = 0
		summary.InviteRate = 0
		summary.TodayInvitedUsers = 0
		summary.Last7dInvitedUsers = 0
		summary.Last7dFirstPaidUsers = 0
		summary.Last7dConversionRate = 0
		summary.Last30dInvitedUsers = 0
		summary.Last30dFirstPaidUsers = 0
		summary.Last30dConversionRate = 0
		summary.TotalFirstPaidUsers = 0
		summary.TotalConversionRate = 0
		summary.TotalUsers = summary.DirectUsers
	}
	if registrationSource == "INVITED" {
		summary.DirectUsers = 0
		summary.InviteRate = 1
		summary.TotalUsers = summary.InvitedUsers
	}
	return summary, nil
}

func (r *InMemoryGrowthRepo) AdminListBrowseHistories(userID string, contentType string, keyword string, page int, pageSize int) ([]model.AdminBrowseHistory, int, error) {
	items := []model.AdminBrowseHistory{
		{
			ID:          "bh_demo_001",
			UserID:      "u_demo_001",
			UserPhone:   "13800000001",
			ContentType: "NEWS",
			ContentID:   "na_001",
			Title:       "A股盘前观察",
			SourcePage:  "/news",
			ViewedAt:    "2026-02-27T09:25:00+08:00",
		},
		{
			ID:          "bh_demo_002",
			UserID:      "u_demo_001",
			UserPhone:   "13800000001",
			ContentType: "REPORT",
			ContentID:   "na_002",
			Title:       "量价协同模型在震荡市的适配边界",
			SourcePage:  "/news",
			ViewedAt:    "2026-02-26T21:12:00+08:00",
		},
	}

	targetUserID := strings.TrimSpace(userID)
	targetType := strings.ToUpper(strings.TrimSpace(contentType))
	targetKeyword := strings.TrimSpace(keyword)
	filtered := make([]model.AdminBrowseHistory, 0, len(items))
	for _, item := range items {
		if targetUserID != "" && item.UserID != targetUserID {
			continue
		}
		if targetType != "" && strings.ToUpper(strings.TrimSpace(item.ContentType)) != targetType {
			continue
		}
		if targetKeyword != "" {
			merged := strings.ToLower(item.ContentID + " " + item.Title + " " + item.UserPhone)
			if !strings.Contains(merged, strings.ToLower(targetKeyword)) {
				continue
			}
		}
		filtered = append(filtered, item)
	}
	return filtered, len(filtered), nil
}

func (r *InMemoryGrowthRepo) AdminGetBrowseHistorySummary() (model.AdminBrowseHistorySummary, error) {
	return model.AdminBrowseHistorySummary{
		TotalViews:   128,
		UniqueUsers:  42,
		NewsViews:    79,
		ReportViews:  31,
		JournalViews: 18,
		TodayViews:   14,
		Last7dViews:  65,
	}, nil
}

func (r *InMemoryGrowthRepo) AdminGetBrowseHistoryTrend(days int) ([]model.AdminBrowseTrendPoint, error) {
	points := []model.AdminBrowseTrendPoint{
		{Date: "2026-02-21", TotalViews: 8, NewsViews: 5, ReportViews: 2, JournalViews: 1},
		{Date: "2026-02-22", TotalViews: 10, NewsViews: 6, ReportViews: 3, JournalViews: 1},
		{Date: "2026-02-23", TotalViews: 12, NewsViews: 7, ReportViews: 3, JournalViews: 2},
		{Date: "2026-02-24", TotalViews: 9, NewsViews: 5, ReportViews: 2, JournalViews: 2},
		{Date: "2026-02-25", TotalViews: 7, NewsViews: 4, ReportViews: 2, JournalViews: 1},
		{Date: "2026-02-26", TotalViews: 11, NewsViews: 6, ReportViews: 3, JournalViews: 2},
		{Date: "2026-02-27", TotalViews: 14, NewsViews: 8, ReportViews: 4, JournalViews: 2},
	}
	if days <= 0 || days >= len(points) {
		return points, nil
	}
	return points[len(points)-days:], nil
}

func (r *InMemoryGrowthRepo) AdminListBrowseUserSegments(limit int) ([]model.AdminBrowseUserSegment, error) {
	items := []model.AdminBrowseUserSegment{
		{
			Segment:         "ACTIVE",
			UserID:          "u_demo_001",
			UserPhone:       "13800000001",
			ViewCount7d:     12,
			LastViewedAt:    "2026-02-27T09:25:00+08:00",
			LastContentID:   "na_001",
			LastContentType: "NEWS",
		},
		{
			Segment:         "ACTIVE",
			UserID:          "u_demo_002",
			UserPhone:       "13800000002",
			ViewCount7d:     9,
			LastViewedAt:    "2026-02-27T08:18:00+08:00",
			LastContentID:   "na_007",
			LastContentType: "REPORT",
		},
		{
			Segment:         "SILENT",
			UserID:          "u_demo_010",
			UserPhone:       "13800000010",
			ViewCount7d:     0,
			LastViewedAt:    "2026-02-15T19:20:00+08:00",
			LastContentID:   "na_003",
			LastContentType: "NEWS",
		},
		{
			Segment:         "SILENT",
			UserID:          "u_demo_011",
			UserPhone:       "13800000011",
			ViewCount7d:     0,
			LastViewedAt:    "",
			LastContentID:   "",
			LastContentType: "",
		},
	}
	if limit <= 0 || limit >= len(items) {
		return items, nil
	}
	return items[:limit], nil
}

func (r *InMemoryGrowthRepo) AdminListUserMessages(userID string, messageType string, readStatus string, page int, pageSize int) ([]model.AdminUserMessage, int, error) {
	items := []model.AdminUserMessage{
		{
			ID:         "msg_demo_001",
			UserID:     "u_demo_001",
			UserPhone:  "13800000001",
			Title:      "策略提醒",
			Content:    "今日股指波动加大，建议收缩仓位至 70%。",
			Type:       "STRATEGY",
			ReadStatus: "UNREAD",
			CreatedAt:  "2026-02-27T10:20:00+08:00",
		},
		{
			ID:         "msg_demo_002",
			UserID:     "u_demo_001",
			UserPhone:  "13800000001",
			Title:      "系统公告",
			Content:    "本周六 02:00-03:00 进行系统维护。",
			Type:       "SYSTEM",
			ReadStatus: "READ",
			CreatedAt:  "2026-02-26T20:00:00+08:00",
		},
	}

	filtered := make([]model.AdminUserMessage, 0, len(items))
	targetUserID := strings.TrimSpace(userID)
	targetType := strings.ToUpper(strings.TrimSpace(messageType))
	targetReadStatus := strings.ToUpper(strings.TrimSpace(readStatus))
	for _, item := range items {
		if targetUserID != "" && item.UserID != targetUserID {
			continue
		}
		if targetType != "" && strings.ToUpper(strings.TrimSpace(item.Type)) != targetType {
			continue
		}
		if targetReadStatus != "" && strings.ToUpper(strings.TrimSpace(item.ReadStatus)) != targetReadStatus {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered, len(filtered), nil
}

func (r *InMemoryGrowthRepo) AdminCreateUserMessages(userIDs []string, title string, content string, messageType string) (int, []model.AdminMessageSendFailure, error) {
	if len(userIDs) == 0 {
		return 0, nil, errors.New("no target users")
	}
	failures := make([]model.AdminMessageSendFailure, 0)
	success := 0
	validUsers := map[string]struct{}{
		"u_demo_001": {},
		"u_demo_002": {},
		"u_demo_010": {},
		"u_demo_011": {},
		"admin_001":  {},
	}
	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" {
			continue
		}
		if _, ok := validUsers[userID]; !ok {
			failures = append(failures, model.AdminMessageSendFailure{UserID: userID, Reason: "user not found"})
			continue
		}
		success++
	}
	return success, failures, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateUserStatus(id string, status string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminUpdateUserMemberLevel(id string, memberLevel string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminUpdateUserKYCStatus(id string, kycStatus string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminResetUserPasswordHash(id string, passwordHash string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminDashboardOverview() (model.AdminDashboardOverview, error) {
	return model.AdminDashboardOverview{
		TotalUsers:              1200,
		ActiveUsers:             1080,
		KYCApprovedUsers:        860,
		VIPUsers:                320,
		ActiveSubscriptions:     486,
		PendingMembershipOrders: 9,
		TodayNewUsers:           26,
		TodayPaidOrders:         14,
		TodayPaidAmount:         2688,
		TodayPublishedStocks:    10,
		TodayPublishedNews:      8,
	}, nil
}

func (r *InMemoryGrowthRepo) AdminCreateOperationLog(module string, action string, targetType string, targetID string, operatorUserID string, beforeValue string, afterValue string, reason string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListOperationLogs(module string, action string, operatorUserID string, page int, pageSize int) ([]model.AdminOperationLog, int, error) {
	items := []model.AdminOperationLog{
		{
			ID:             "aol_001",
			Module:         "USER",
			Action:         "UPDATE_STATUS",
			TargetType:     "USER",
			TargetID:       "u_demo_001",
			OperatorUserID: "admin_001",
			BeforeValue:    "ACTIVE",
			AfterValue:     "DISABLED",
			Reason:         "demo",
			CreatedAt:      "2026-02-25T12:00:00+08:00",
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error) {
	items := []model.MembershipProduct{
		{ID: "mp_001", Name: "VIP月卡", Price: 99, Status: "ACTIVE", MemberLevel: "VIP1", DurationDays: 30},
		{ID: "mp_002", Name: "VIP年卡", Price: 899, Status: "ACTIVE", MemberLevel: "VIP2", DurationDays: 365},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminCreateMembershipProduct(name string, price float64, status string, memberLevel string, durationDays int) (string, error) {
	return "mp_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateMembershipProduct(id string, name string, price float64, status string, memberLevel string, durationDays int) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminUpdateMembershipProductStatus(id string, status string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListMembershipOrders(status string, userID string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error) {
	items := []model.MembershipOrderAdmin{
		{ID: "mo_001", OrderNo: "mo_001", UserID: "u_demo_001", ProductID: "mp_001", Amount: 99, PayChannel: "ALIPAY", Status: "PAID", PaidAt: "2026-02-25T10:00:00+08:00", CreatedAt: "2026-02-25T09:59:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminUpdateMembershipOrderStatus(id string, status string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminGetExperimentAnalyticsSummary(days int) (model.AdminExperimentAnalyticsSummary, error) {
	overview := model.AdminExperimentAnalyticsOverview{
		Days:                   days,
		TotalEvents:            0,
		TotalExperiments:       0,
		ExposureCount:          0,
		ClickCount:             0,
		UpgradeIntentCount:     0,
		PaymentSuccessCount:    0,
		RenewalSuccessCount:    0,
		ClickThroughRate:       0,
		UpgradePerClickRate:    0,
		UpgradePerExposureRate: 0,
		PaidPerUpgradeRate:     0,
		PaidPerClickRate:       0,
		PaidPerExposureRate:    0,
	}
	return model.AdminExperimentAnalyticsSummary{
		Days:                days,
		Overview:            overview,
		Items:               []model.AdminExperimentAnalyticsItem{},
		PageBreakdown:       []model.AdminExperimentAnalyticsPageItem{},
		DailyTrend:          []model.AdminExperimentAnalyticsTrendPoint{},
		PayChannelBreakdown: []model.AdminExperimentAnalyticsPayChannelItem{},
		DeviceBreakdown:     []model.AdminExperimentAnalyticsDeviceItem{},
		UserStageBreakdown:  []model.AdminExperimentAnalyticsUserStageItem{},
		VariantDailyTrend:   []model.AdminExperimentAnalyticsVariantTrendPoint{},
	}, nil
}

func (r *InMemoryGrowthRepo) AdminListVIPQuotaConfigs(memberLevel string, status string, page int, pageSize int) ([]model.VIPQuotaConfig, int, error) {
	items := []model.VIPQuotaConfig{
		{ID: "vqc_001", MemberLevel: "FREE", DocReadLimit: 3, NewsSubscribeLimit: 2, ResetCycle: "MONTHLY", Status: "ACTIVE", EffectiveAt: "2026-02-01T00:00:00+08:00", UpdatedAt: "2026-02-25T00:00:00+08:00"},
		{ID: "vqc_002", MemberLevel: "VIP2", DocReadLimit: 200, NewsSubscribeLimit: 100, ResetCycle: "MONTHLY", Status: "ACTIVE", EffectiveAt: "2026-02-01T00:00:00+08:00", UpdatedAt: "2026-02-25T00:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminCreateVIPQuotaConfig(item model.VIPQuotaConfig) (string, error) {
	return "vqc_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateVIPQuotaConfig(id string, item model.VIPQuotaConfig) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListUserQuotaUsages(userID string, periodKey string, page int, pageSize int) ([]model.UserQuotaUsage, int, error) {
	items := []model.UserQuotaUsage{
		{
			UserID:             "u_demo_001",
			MemberLevel:        "VIP1",
			PeriodKey:          "2026-02",
			DocReadLimit:       100,
			DocReadUsed:        24,
			NewsSubscribeLimit: 50,
			NewsSubscribeUsed:  12,
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminAdjustUserQuota(userID string, periodKey string, docReadDelta int, newsSubscribeDelta int) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListDataSources(page int, pageSize int) ([]model.DataSource, int, error) {
	items := []model.DataSource{
		{
			ID:         "ds_001",
			SourceKey:  "wind",
			Name:       "Wind",
			SourceType: "MARKET",
			Status:     "ACTIVE",
			Config: map[string]interface{}{
				"endpoint": "https://api.example.com/wind",
			},
			UpdatedAt: "2026-02-26T09:00:00+08:00",
		},
		{
			ID:         "ds_002",
			SourceKey:  "mock_stock",
			Name:       "Mock Stock Quotes",
			SourceType: "STOCK",
			Status:     "ACTIVE",
			Config: map[string]interface{}{
				"provider":          "MOCK",
				"endpoint":          "http://127.0.0.1:18080/healthz",
				"retry_times":       0,
				"retry_interval_ms": 200,
				"fail_threshold":    5,
				"health_timeout_ms": 3000,
			},
			UpdatedAt: "2026-02-27T20:00:00+08:00",
		},
		{
			ID:         "ds_003",
			SourceKey:  "tushare",
			Name:       "Tushare",
			SourceType: "STOCK",
			Status:     "ACTIVE",
			Config: map[string]interface{}{
				"provider":          "TUSHARE",
				"endpoint":          "https://api.tushare.pro",
				"retry_times":       1,
				"retry_interval_ms": 500,
				"fail_threshold":    3,
				"health_timeout_ms": 8000,
			},
			UpdatedAt: "2026-02-27T20:00:00+08:00",
		},
		{
			ID:         "ds_004",
			SourceKey:  "tushare_cn",
			Name:       "Tushare CN Mirror",
			SourceType: "STOCK",
			Status:     "ACTIVE",
			Config: map[string]interface{}{
				"provider":          "TUSHARE",
				"endpoint":          "https://api.tushare.pro",
				"retry_times":       1,
				"retry_interval_ms": 500,
				"fail_threshold":    3,
				"health_timeout_ms": 8000,
			},
			UpdatedAt: "2026-02-27T20:00:00+08:00",
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminCreateDataSource(item model.DataSource) (string, error) {
	return "ds_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateDataSource(sourceKey string, item model.DataSource) error {
	if strings.TrimSpace(sourceKey) == "" {
		return sql.ErrNoRows
	}
	if !inMemoryDataSourceExists(sourceKey) {
		return sql.ErrNoRows
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminDeleteDataSource(sourceKey string) error {
	if strings.TrimSpace(sourceKey) == "" {
		return sql.ErrNoRows
	}
	if !inMemoryDataSourceExists(sourceKey) {
		return sql.ErrNoRows
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminCheckDataSourceHealth(sourceKey string) (model.DataSourceHealthCheck, error) {
	sourceKey = strings.TrimSpace(sourceKey)
	if !inMemoryDataSourceExists(sourceKey) {
		return model.DataSourceHealthCheck{}, sql.ErrNoRows
	}
	message := "ok"
	if strings.EqualFold(sourceKey, "tushare") {
		message = "tushare mock healthy"
	}
	return model.DataSourceHealthCheck{
		SourceKey:           sourceKey,
		Status:              "HEALTHY",
		Reachable:           true,
		LatencyMS:           8,
		Message:             message,
		Attempts:            1,
		MaxAttempts:         1,
		ConsecutiveFailures: 0,
		AlertTriggered:      false,
		CheckedAt:           time.Now().Format(time.RFC3339),
	}, nil
}

func (r *InMemoryGrowthRepo) AdminBatchCheckDataSourceHealth(sourceKeys []string) ([]model.DataSourceHealthCheck, error) {
	targets := make([]string, 0)
	if len(sourceKeys) == 0 {
		targets = append(targets, "tushare", "mock_stock", "wind")
	} else {
		for _, key := range sourceKeys {
			trimmed := strings.TrimSpace(key)
			if trimmed != "" {
				targets = append(targets, trimmed)
			}
		}
	}
	items := make([]model.DataSourceHealthCheck, 0, len(targets))
	for _, key := range targets {
		item, err := r.AdminCheckDataSourceHealth(key)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				items = append(items, model.DataSourceHealthCheck{
					SourceKey: key,
					Status:    "UNKNOWN",
					Message:   "data source not found",
					CheckedAt: time.Now().Format(time.RFC3339),
				})
				continue
			}
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *InMemoryGrowthRepo) AdminListDataSourceHealthLogs(sourceKey string, page int, pageSize int) ([]model.DataSourceHealthLog, int, error) {
	sourceKey = strings.TrimSpace(sourceKey)
	if !inMemoryDataSourceExists(sourceKey) {
		return nil, 0, sql.ErrNoRows
	}
	items := []model.DataSourceHealthLog{
		{
			ID:         "dshl_001",
			SourceKey:  sourceKey,
			Status:     "HEALTHY",
			Reachable:  true,
			HTTPStatus: 200,
			LatencyMS:  8,
			Message:    "ok",
			CheckedAt:  time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
		},
	}
	return items, len(items), nil
}

func inMemoryDataSourceExists(sourceKey string) bool {
	switch strings.ToLower(strings.TrimSpace(sourceKey)) {
	case "wind", "ds_new_001", "mock_stock", "tushare", "tushare_cn", "akshare", "tickermd", "myself":
		return true
	default:
		return false
	}
}

func (r *InMemoryGrowthRepo) AdminListSystemConfigs(keyword string, page int, pageSize int) ([]model.SystemConfig, int, error) {
	items := []model.SystemConfig{
		{ID: "cfg_stock_model", ConfigKey: "stock.model.version", ConfigValue: "v1", Description: "股票推荐模型版本", UpdatedBy: "system", UpdatedAt: "2026-02-25T00:00:00+08:00"},
		{ID: "cfg_futures_model", ConfigKey: "futures.model.version", ConfigValue: "v1", Description: "期货策略模型版本", UpdatedBy: "system", UpdatedAt: "2026-02-25T00:00:00+08:00"},
		{ID: "cfg_stock_quotes_default_source", ConfigKey: "stock.quotes.default_source_key", ConfigValue: "TUSHARE", Description: "股票行情默认数据源", UpdatedBy: "system", UpdatedAt: "2026-02-28T00:00:00+08:00"},
		{ID: "cfg_market_instrument_stock_priority", ConfigKey: marketInstrumentStockPriorityConfigKey, ConfigValue: "TUSHARE,AKSHARE,TICKERMD,MYSELF,MOCK", Description: "股票主数据多源优先级", UpdatedBy: "system", UpdatedAt: "2026-03-22T00:00:00+08:00"},
		{ID: "cfg_market_instrument_futures_priority", ConfigKey: marketInstrumentFuturesPriorityConfigKey, ConfigValue: "TUSHARE,AKSHARE,TICKERMD,MYSELF,MOCK", Description: "期货主数据多源优先级", UpdatedBy: "system", UpdatedAt: "2026-03-22T00:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminUpsertSystemConfig(configKey string, configValue string, description string, operator string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListReviewTasks(module string, status string, submitterID string, reviewerID string, page int, pageSize int) ([]model.ReviewTask, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	upperModule := strings.ToUpper(strings.TrimSpace(module))
	upperStatus := strings.ToUpper(strings.TrimSpace(status))
	items := make([]model.ReviewTask, 0, len(r.reviewTasks))
	for _, item := range r.reviewTasks {
		if upperModule != "" && strings.ToUpper(item.Module) != upperModule {
			continue
		}
		if upperStatus != "" && strings.ToUpper(item.Status) != upperStatus {
			continue
		}
		if submitterID != "" && item.SubmitterID != submitterID {
			continue
		}
		if reviewerID != "" && item.ReviewerID != reviewerID {
			continue
		}
		items = append(items, item)
	}
	if len(items) == 0 && upperModule == "" && upperStatus == "" && submitterID == "" && reviewerID == "" {
		items = append(items, model.ReviewTask{
			ID:          "rt_001",
			Module:      "STOCK",
			TargetID:    "sr_demo_001",
			SubmitterID: "admin_001",
			Status:      "PENDING",
			SubmitNote:  "请审核今日策略",
			SubmittedAt: "2026-02-25T09:00:00+08:00",
		})
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminSubmitReviewTask(module string, targetID string, submitterID string, reviewerID string, submitNote string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	module = strings.ToUpper(strings.TrimSpace(module))
	for _, item := range r.reviewTasks {
		if item.Module == module && item.TargetID == strings.TrimSpace(targetID) && item.Status == "PENDING" {
			return "", errors.New("pending review task already exists")
		}
	}
	if err := r.applyModuleTargetStatusLocked(module, targetID, "REVIEWING"); err != nil {
		return "", err
	}
	now := time.Now().Format(time.RFC3339)
	item := model.ReviewTask{
		ID:          newID("rt"),
		Module:      module,
		TargetID:    strings.TrimSpace(targetID),
		SubmitterID: strings.TrimSpace(submitterID),
		ReviewerID:  strings.TrimSpace(reviewerID),
		Status:      "PENDING",
		SubmitNote:  strings.TrimSpace(submitNote),
		SubmittedAt: now,
	}
	r.reviewTasks[item.ID] = item
	return item.ID, nil
}

func (r *InMemoryGrowthRepo) AdminAssignReviewTask(reviewID string, reviewerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.reviewTasks[strings.TrimSpace(reviewID)]
	if !ok {
		return sql.ErrNoRows
	}
	item.ReviewerID = strings.TrimSpace(reviewerID)
	r.reviewTasks[item.ID] = item
	return nil
}

func (r *InMemoryGrowthRepo) AdminReviewTaskDecision(reviewID string, status string, reviewerID string, reviewNote string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.reviewTasks[strings.TrimSpace(reviewID)]
	if !ok {
		return sql.ErrNoRows
	}
	if item.Status != "PENDING" {
		return errors.New("review task is not pending")
	}
	if item.ReviewerID != "" && item.ReviewerID != strings.TrimSpace(reviewerID) {
		return errors.New("reviewer mismatch")
	}
	item.Status = strings.ToUpper(strings.TrimSpace(status))
	item.ReviewerID = strings.TrimSpace(reviewerID)
	item.ReviewNote = strings.TrimSpace(reviewNote)
	item.ReviewedAt = time.Now().Format(time.RFC3339)
	r.reviewTasks[item.ID] = item

	targetStatus := "DRAFT"
	if item.Status == "APPROVED" {
		targetStatus = "PUBLISHED"
	}
	return r.applyModuleTargetStatusLocked(item.Module, item.TargetID, targetStatus)
}

func (r *InMemoryGrowthRepo) ensureHighPriorityStockEventTaskLocked(cluster model.StockEventCluster) {
	priority, _ := cluster.Metadata["review_priority"].(string)
	if strings.ToUpper(strings.TrimSpace(priority)) != "HIGH" || strings.ToUpper(strings.TrimSpace(cluster.ReviewStatus)) != "PENDING" {
		return
	}
	for _, item := range r.reviewTasks {
		if item.Module == "STOCK_EVENT" && item.TargetID == cluster.ID && item.Status == "PENDING" {
			return
		}
	}
	now := time.Now().Format(time.RFC3339)
	item := model.ReviewTask{
		ID:          newID("rt"),
		Module:      "STOCK_EVENT",
		TargetID:    cluster.ID,
		SubmitterID: "system",
		Status:      "PENDING",
		SubmitNote:  "高优先级股票事件待审核",
		SubmittedAt: now,
	}
	r.reviewTasks[item.ID] = item
}

func (r *InMemoryGrowthRepo) applyModuleTargetStatusLocked(module string, targetID string, status string) error {
	switch strings.ToUpper(strings.TrimSpace(module)) {
	case "STOCK_EVENT":
		cluster, ok := r.stockEventClusters[strings.TrimSpace(targetID)]
		if !ok {
			return errors.New("target not found")
		}
		switch strings.ToUpper(strings.TrimSpace(status)) {
		case "PUBLISHED":
			cluster.Status = "REVIEWED"
			cluster.ReviewStatus = "APPROVED"
		case "DRAFT":
			cluster.Status = "REJECTED"
			cluster.ReviewStatus = "REJECTED"
		default:
			cluster.Status = "CLUSTERED"
			cluster.ReviewStatus = "PENDING"
		}
		cluster.UpdatedAt = time.Now().Format(time.RFC3339)
		r.stockEventClusters[cluster.ID] = cluster
		return nil
	default:
		return nil
	}
	return nil
}

func (r *InMemoryGrowthRepo) GetSchedulerJobNameByRunID(runID string) (string, error) {
	return "daily_stock_recommendation", nil
}

func (r *InMemoryGrowthRepo) AdminListSchedulerJobRuns(jobName string, status string, page int, pageSize int) ([]model.SchedulerJobRun, int, error) {
	items := []model.SchedulerJobRun{
		{
			ID:            "jr_001",
			JobName:       "daily_stock_recommendation",
			TriggerSource: "SYSTEM",
			Status:        "SUCCESS",
			StartedAt:     "2026-02-25T08:30:00+08:00",
			FinishedAt:    "2026-02-25T08:30:04+08:00",
			ResultSummary: "generated 10 recommendations",
			OperatorID:    "system",
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminListNewsSyncRunDetails(runID string, syncType string, source string, symbol string, status string, page int, pageSize int) ([]model.NewsSyncRunDetail, int, error) {
	items := []model.NewsSyncRunDetail{
		{
			ID:            "nsd_demo_001",
			RunID:         runID,
			JobName:       "tushare_news_incremental",
			SyncType:      "NEWS_BRIEF",
			Source:        "cls",
			Symbol:        "",
			Status:        "SUCCESS",
			FetchedCount:  8,
			UpsertedCount: 8,
			FailedCount:   0,
			StartedAt:     "2026-02-28T00:00:00+08:00",
			FinishedAt:    "2026-02-28T00:00:02+08:00",
			CreatedAt:     "2026-02-28T00:00:02+08:00",
			UpdatedAt:     "2026-02-28T00:00:02+08:00",
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminCreateNewsSyncRunDetails(runID string, details []model.NewsSyncRunDetail) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminCreateSchedulerJobRun(jobName string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	return "jr_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminRetrySchedulerJobRun(runID string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	return "jr_retry_001", nil
}

func (r *InMemoryGrowthRepo) AdminListSchedulerJobDefinitions(status string, module string, page int, pageSize int) ([]model.SchedulerJobDefinition, int, error) {
	items := []model.SchedulerJobDefinition{
		{
			ID:          "jobdef_stock_daily",
			JobName:     "daily_stock_recommendation",
			DisplayName: "每日股票推荐生成",
			Module:      "STOCK",
			CronExpr:    "0 30 8 * * *",
			Status:      "ACTIVE",
			UpdatedBy:   "system",
			CreatedAt:   "2026-02-25T00:00:00+08:00",
			UpdatedAt:   "2026-02-25T00:00:00+08:00",
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminCreateSchedulerJobDefinition(item model.SchedulerJobDefinition, operatorID string) (string, error) {
	return "jobdef_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateSchedulerJobDefinition(id string, item model.SchedulerJobDefinition, operatorID string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminUpdateSchedulerJobDefinitionStatus(id string, status string, operatorID string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminDeleteSchedulerJobDefinition(id string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListWorkflowMessages(module string, eventType string, isRead string, receiverID string, page int, pageSize int) ([]model.WorkflowMessage, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	module = strings.ToUpper(strings.TrimSpace(module))
	eventType = strings.ToUpper(strings.TrimSpace(eventType))
	isRead = strings.ToLower(strings.TrimSpace(isRead))
	receiverID = strings.TrimSpace(receiverID)

	items := make([]model.WorkflowMessage, 0, len(r.workflowMessages))
	for _, item := range r.workflowMessages {
		if module != "" && strings.ToUpper(strings.TrimSpace(item.Module)) != module {
			continue
		}
		if eventType != "" && strings.ToUpper(strings.TrimSpace(item.EventType)) != eventType {
			continue
		}
		if receiverID != "" && strings.TrimSpace(item.ReceiverID) != receiverID {
			continue
		}
		switch isRead {
		case "true", "1", "yes":
			if !item.IsRead {
				continue
			}
		case "false", "0", "no":
			if item.IsRead {
				continue
			}
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].CreatedAt == items[j].CreatedAt {
			return items[i].ID > items[j].ID
		}
		return items[i].CreatedAt > items[j].CreatedAt
	})

	total := len(items)
	if pageSize <= 0 {
		pageSize = total
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * pageSize
	if start >= total {
		return []model.WorkflowMessage{}, total, nil
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return append([]model.WorkflowMessage(nil), items[start:end]...), total, nil
}

func (r *InMemoryGrowthRepo) AdminCountUnreadWorkflowMessages(module string, eventType string, receiverID string) (int, error) {
	items, _, err := r.AdminListWorkflowMessages(module, eventType, "false", receiverID, 1, 100000)
	if err != nil {
		return 0, err
	}
	return len(items), nil
}

func (r *InMemoryGrowthRepo) AdminUpdateWorkflowMessageRead(id string, isRead bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.workflowMessages[strings.TrimSpace(id)]
	if !ok {
		return nil
	}
	item.IsRead = isRead
	if isRead {
		item.ReadAt = time.Now().Format(time.RFC3339)
	} else {
		item.ReadAt = ""
	}
	r.workflowMessages[item.ID] = item
	return nil
}

func (r *InMemoryGrowthRepo) AdminBulkReadWorkflowMessages(module string, eventType string, receiverID string) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	module = strings.ToUpper(strings.TrimSpace(module))
	eventType = strings.ToUpper(strings.TrimSpace(eventType))
	receiverID = strings.TrimSpace(receiverID)

	var affected int64
	now := time.Now().Format(time.RFC3339)
	for id, item := range r.workflowMessages {
		if module != "" && strings.ToUpper(strings.TrimSpace(item.Module)) != module {
			continue
		}
		if eventType != "" && strings.ToUpper(strings.TrimSpace(item.EventType)) != eventType {
			continue
		}
		if receiverID != "" && strings.TrimSpace(item.ReceiverID) != receiverID {
			continue
		}
		if item.IsRead {
			continue
		}
		item.IsRead = true
		item.ReadAt = now
		r.workflowMessages[id] = item
		affected++
	}
	return affected, nil
}

func (r *InMemoryGrowthRepo) AdminCreateWorkflowMessage(reviewID string, targetID string, module string, receiverID string, senderID string, eventType string, title string, content string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	reviewID = strings.TrimSpace(reviewID)
	targetID = strings.TrimSpace(targetID)
	module = strings.ToUpper(strings.TrimSpace(module))
	receiverID = strings.TrimSpace(receiverID)
	senderID = strings.TrimSpace(senderID)
	eventType = strings.ToUpper(strings.TrimSpace(eventType))
	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)

	for id, item := range r.workflowMessages {
		if item.ReviewID == reviewID &&
			item.TargetID == targetID &&
			strings.ToUpper(strings.TrimSpace(item.Module)) == module &&
			strings.TrimSpace(item.ReceiverID) == receiverID &&
			strings.ToUpper(strings.TrimSpace(item.EventType)) == eventType {
			item.Title = title
			item.Content = content
			item.SenderID = senderID
			r.workflowMessages[id] = item
			return nil
		}
	}

	now := time.Now().Format(time.RFC3339)
	id := fmt.Sprintf("wm_%03d", len(r.workflowMessages)+1)
	r.workflowMessages[id] = model.WorkflowMessage{
		ID:         id,
		ReviewID:   reviewID,
		TargetID:   targetID,
		Module:     module,
		ReceiverID: receiverID,
		SenderID:   senderID,
		EventType:  eventType,
		Title:      title,
		Content:    content,
		IsRead:     false,
		CreatedAt:  now,
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminGetWorkflowMetrics(module string, receiverID string) (model.WorkflowMetrics, error) {
	return model.WorkflowMetrics{
		PendingReviews: 2,
		ApprovedToday:  3,
		RejectedToday:  1,
		UnreadMessages: 4,
		TotalMessages:  15,
	}, nil
}

func (r *InMemoryGrowthRepo) AdminGetSchedulerJobMetrics(jobName string) (model.SchedulerJobMetrics, error) {
	return model.SchedulerJobMetrics{
		TodayTotal:         8,
		TodaySuccess:       6,
		TodayFailed:        1,
		TodayRunning:       1,
		RetryTotal:         3,
		RetrySuccess:       2,
		RetryFailed:        1,
		RetryHitRate:       0.6667,
		AvgRetryCount:      1.33,
		AutoRetryTotal:     2,
		RecoveryTotal:      2,
		RecoverySuccess:    1,
		RecoveryHitRate:    0.5,
		FailureReasonScope: "LAST_7_DAYS",
		FailureReasons: []model.SchedulerJobFailureReason{
			{Reason: "TUSHARE_TOKEN_INVALID_OR_MISSING", Count: 3, LastOccurredAt: "2026-02-28T09:15:00+08:00"},
			{Reason: "UPSTREAM_TIMEOUT", Count: 2, LastOccurredAt: "2026-02-28T08:30:00+08:00"},
			{Reason: "MYSQL_CONNECTION_FAILED", Count: 1, LastOccurredAt: "2026-02-27T22:10:00+08:00"},
		},
		JobRetryStats: []model.SchedulerJobRetryStat{
			{
				JobName:         "daily_stock_quant_pipeline",
				TodayTotal:      4,
				TodaySuccess:    3,
				TodayFailed:     1,
				RetryTotal:      2,
				RetrySuccess:    1,
				RetryFailed:     1,
				RetryHitRate:    0.5,
				AvgRetryCount:   1.5,
				AutoRetryTotal:  2,
				RecoveryTotal:   1,
				RecoverySuccess: 1,
				RecoveryHitRate: 1,
			},
			{
				JobName:         "daily_stock_recommendation",
				TodayTotal:      4,
				TodaySuccess:    3,
				TodayFailed:     0,
				TodayRunning:    1,
				RetryTotal:      1,
				RetrySuccess:    1,
				RetryHitRate:    1,
				AvgRetryCount:   1,
				AutoRetryTotal:  0,
				RecoveryTotal:   1,
				RecoverySuccess: 0,
				RecoveryHitRate: 0,
			},
		},
		JobFailureReasons: []model.SchedulerJobFailureByJob{
			{JobName: "daily_stock_quant_pipeline", Reason: "TUSHARE_TOKEN_INVALID_OR_MISSING", Count: 3, LastOccurredAt: "2026-02-28T09:15:00+08:00"},
			{JobName: "daily_stock_quant_pipeline", Reason: "UPSTREAM_TIMEOUT", Count: 2, LastOccurredAt: "2026-02-28T08:30:00+08:00"},
			{JobName: "daily_stock_recommendation", Reason: "MYSQL_CONNECTION_FAILED", Count: 1, LastOccurredAt: "2026-02-27T22:10:00+08:00"},
		},
	}, nil
}
