package repo

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

type InMemoryGrowthRepo struct{}

func NewInMemoryGrowthRepo() *InMemoryGrowthRepo {
	return &InMemoryGrowthRepo{}
}

func (r *InMemoryGrowthRepo) ListBrowseHistory(userID string, contentType string, page int, pageSize int) ([]model.BrowseHistory, int, error) {
	items := []model.BrowseHistory{
		{ID: "h_001", ContentType: "NEWS", ContentID: "news_001", Title: "政策观察", SourcePage: "/news", ViewedAt: "2026-02-24T10:00:00+08:00"},
	}
	return items, len(items), nil
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
		{ID: "sl_001", InviteCode: "ABCD1234", URL: "https://example.com/invite/ABCD1234", Channel: "wechat", Status: "ACTIVE", ExpiredAt: ""},
	}
	return items, nil
}

func (r *InMemoryGrowthRepo) CreateShareLink(userID string, channel string, expiredAt string) (model.ShareLink, error) {
	return model.ShareLink{
		ID:         "sl_002",
		InviteCode: "NEWC1234",
		URL:        "https://example.com/invite/NEWC1234",
		Channel:    channel,
		Status:     "ACTIVE",
		ExpiredAt:  expiredAt,
	}, nil
}

func (r *InMemoryGrowthRepo) ListInviteRecords(userID string, page int, pageSize int) ([]model.InviteRecord, int, error) {
	items := []model.InviteRecord{
		{ID: "iv_001", InviteeUser: "u_1002", Status: "FIRST_PAID", RegisterAt: "2026-02-20T12:00:00+08:00", FirstPayAt: "2026-02-21T09:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) ListRewardRecords(userID string, page int, pageSize int) ([]model.RewardRecord, int, error) {
	items := []model.RewardRecord{
		{ID: "rw_001", RewardType: "CASH", RewardValue: 20, TriggerEvent: "INVITEE_FIRST_RECHARGE", Status: "ISSUED", IssuedAt: "2026-02-21T10:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) GetUserProfile(userID string) (model.UserProfile, error) {
	return model.UserProfile{
		ID:          userID,
		Phone:       "13800000001",
		Email:       "demo@sercherai.local",
		KYCStatus:   "PENDING",
		MemberLevel: "FREE",
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
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) MarkMessageRead(userID string, id string) error {
	return nil
}

func (r *InMemoryGrowthRepo) GetUserAccessProfile(userID string) (model.UserAccessProfile, error) {
	return model.UserAccessProfile{
		UserID:      userID,
		Status:      "ACTIVE",
		KYCStatus:   "APPROVED",
		MemberLevel: "VIP1",
	}, nil
}

func (r *InMemoryGrowthRepo) GetMembershipQuota(userID string) (model.MembershipQuota, error) {
	return model.MembershipQuota{
		MemberLevel:            "VIP1",
		PeriodKey:              "2026-02",
		DocReadLimit:           100,
		DocReadUsed:            24,
		DocReadRemaining:       76,
		NewsSubscribeLimit:     50,
		NewsSubscribeUsed:      12,
		NewsSubscribeRemaining: 38,
		ResetCycle:             "MONTHLY",
		ResetAt:                "2026-03-01T00:00:00+08:00",
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
		{ID: "na_001", CategoryID: "nc_001", Title: "A股盘前观察", Summary: "示例摘要", Visibility: "PUBLIC", Status: "PUBLISHED", PublishedAt: "2026-02-25T09:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) GetNewsArticleDetail(userID string, articleID string) (model.NewsArticle, error) {
	return model.NewsArticle{
		ID: "na_001", CategoryID: "nc_001", Title: "A股盘前观察", Summary: "示例摘要", Content: "示例正文",
		Visibility: "PUBLIC", Status: "PUBLISHED", PublishedAt: "2026-02-25T09:00:00+08:00", AuthorID: "admin_001",
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

func (r *InMemoryGrowthRepo) AdminCreateNewsArticle(categoryID string, title string, summary string, content string, visibility string, status string, authorID string) (string, error) {
	return "na_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminUpdateNewsArticle(id string, categoryID string, title string, summary string, content string, visibility string, status string) error {
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

func (r *InMemoryGrowthRepo) AdminGenerateDailyStockRecommendations(tradeDate string) (int, error) {
	return 10, nil
}

func (r *InMemoryGrowthRepo) AdminGenerateDailyFuturesStrategies(tradeDate string) (int, error) {
	return 3, nil
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

func (r *InMemoryGrowthRepo) AdminListUsers(status string, kycStatus string, memberLevel string, page int, pageSize int) ([]model.AdminUser, int, error) {
	items := []model.AdminUser{
		{ID: "u_demo_001", Phone: "13800000001", Email: "demo@sercherai.local", Status: "ACTIVE", KYCStatus: "APPROVED", MemberLevel: "VIP1", CreatedAt: "2026-02-20T10:00:00+08:00"},
		{ID: "admin_001", Phone: "13800000000", Email: "admin@sercherai.local", Status: "ACTIVE", KYCStatus: "APPROVED", MemberLevel: "VIP1", CreatedAt: "2026-02-20T10:00:00+08:00"},
	}
	return items, len(items), nil
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

func (r *InMemoryGrowthRepo) AdminDashboardOverview() (model.AdminDashboardOverview, error) {
	return model.AdminDashboardOverview{
		TotalUsers:           1200,
		ActiveUsers:          1080,
		KYCApprovedUsers:     860,
		VIPUsers:             320,
		TodayNewUsers:        26,
		TodayPaidOrders:      14,
		TodayPublishedStocks: 10,
		TodayPublishedNews:   8,
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
	if sourceKey != "wind" && sourceKey != "ds_new_001" {
		return sql.ErrNoRows
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminDeleteDataSource(sourceKey string) error {
	if strings.TrimSpace(sourceKey) == "" {
		return sql.ErrNoRows
	}
	if sourceKey != "wind" && sourceKey != "ds_new_001" {
		return sql.ErrNoRows
	}
	return nil
}

func (r *InMemoryGrowthRepo) AdminCheckDataSourceHealth(sourceKey string) (model.DataSourceHealthCheck, error) {
	if strings.TrimSpace(sourceKey) == "" || (sourceKey != "wind" && sourceKey != "ds_new_001") {
		return model.DataSourceHealthCheck{}, sql.ErrNoRows
	}
	return model.DataSourceHealthCheck{
		SourceKey:           sourceKey,
		Status:              "HEALTHY",
		Reachable:           true,
		LatencyMS:           8,
		Message:             "ok",
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
		targets = append(targets, "wind")
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
	if strings.TrimSpace(sourceKey) == "" || (sourceKey != "wind" && sourceKey != "ds_new_001") {
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

func (r *InMemoryGrowthRepo) AdminListSystemConfigs(keyword string, page int, pageSize int) ([]model.SystemConfig, int, error) {
	items := []model.SystemConfig{
		{ID: "cfg_stock_model", ConfigKey: "stock.model.version", ConfigValue: "v1", Description: "股票推荐模型版本", UpdatedBy: "system", UpdatedAt: "2026-02-25T00:00:00+08:00"},
		{ID: "cfg_futures_model", ConfigKey: "futures.model.version", ConfigValue: "v1", Description: "期货策略模型版本", UpdatedBy: "system", UpdatedAt: "2026-02-25T00:00:00+08:00"},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminUpsertSystemConfig(configKey string, configValue string, description string, operator string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminListReviewTasks(module string, status string, submitterID string, reviewerID string, page int, pageSize int) ([]model.ReviewTask, int, error) {
	items := []model.ReviewTask{
		{
			ID:          "rt_001",
			Module:      "STOCK",
			TargetID:    "sr_demo_001",
			SubmitterID: "admin_001",
			Status:      "PENDING",
			SubmitNote:  "请审核今日策略",
			SubmittedAt: "2026-02-25T09:00:00+08:00",
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminSubmitReviewTask(module string, targetID string, submitterID string, reviewerID string, submitNote string) (string, error) {
	return "rt_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminAssignReviewTask(reviewID string, reviewerID string) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminReviewTaskDecision(reviewID string, status string, reviewerID string, reviewNote string) error {
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

func (r *InMemoryGrowthRepo) AdminCreateSchedulerJobRun(jobName string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	return "jr_new_001", nil
}

func (r *InMemoryGrowthRepo) AdminRetrySchedulerJobRun(runID string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
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

func (r *InMemoryGrowthRepo) AdminListWorkflowMessages(module string, eventType string, isRead string, receiverID string, page int, pageSize int) ([]model.WorkflowMessage, int, error) {
	items := []model.WorkflowMessage{
		{
			ID:         "wm_001",
			ReviewID:   "rt_001",
			TargetID:   "sr_demo_001",
			Module:     "STOCK",
			ReceiverID: "admin_001",
			SenderID:   "admin_001",
			EventType:  "REVIEW_SUBMITTED",
			Title:      "策略待审核",
			Content:    "有新的股票策略提交审核",
			IsRead:     false,
			CreatedAt:  "2026-02-25T10:00:00+08:00",
		},
	}
	return items, len(items), nil
}

func (r *InMemoryGrowthRepo) AdminCountUnreadWorkflowMessages(module string, eventType string, receiverID string) (int, error) {
	return 1, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateWorkflowMessageRead(id string, isRead bool) error {
	return nil
}

func (r *InMemoryGrowthRepo) AdminBulkReadWorkflowMessages(module string, eventType string, receiverID string) (int64, error) {
	return 1, nil
}

func (r *InMemoryGrowthRepo) AdminCreateWorkflowMessage(reviewID string, targetID string, module string, receiverID string, senderID string, eventType string, title string, content string) error {
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
		TodayTotal:   8,
		TodaySuccess: 6,
		TodayFailed:  1,
		TodayRunning: 1,
	}, nil
}
