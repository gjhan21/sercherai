package repo

import "sercherai/backend/internal/growth/model"

type GrowthRepo interface {
	ListBrowseHistory(userID string, contentType string, page int, pageSize int) ([]model.BrowseHistory, int, error)
	DeleteBrowseHistoryItem(userID string, id string) error
	ClearBrowseHistory(userID string) error

	ListRechargeRecords(userID string, status string, page int, pageSize int) ([]model.RechargeRecord, int, error)

	ListShareLinks(userID string) ([]model.ShareLink, error)
	CreateShareLink(userID string, channel string, expiredAt string) (model.ShareLink, error)

	ListInviteRecords(userID string, page int, pageSize int) ([]model.InviteRecord, int, error)
	ListRewardRecords(userID string, page int, pageSize int) ([]model.RewardRecord, int, error)
	GetUserProfile(userID string) (model.UserProfile, error)
	UpdateUserProfileEmail(userID string, email string) error
	SubmitUserKYC(userID string, realName string, idNumber string, provider string) (string, error)
	ListSubscriptions(userID string, page int, pageSize int) ([]model.Subscription, int, error)
	CreateSubscription(userID string, subType string, scope string, frequency string) (string, error)
	UpdateSubscription(userID string, id string, frequency string, status string) error
	ListMessages(userID string, page int, pageSize int) ([]model.UserMessage, int, error)
	MarkMessageRead(userID string, id string) error
	GetUserAccessProfile(userID string) (model.UserAccessProfile, error)
	GetMembershipQuota(userID string) (model.MembershipQuota, error)
	GetAttachmentFileInfo(userID string, attachmentID string) (model.AttachmentFileInfo, error)
	LogAttachmentDownload(userID string, attachmentID string, articleID string) error
	ListNewsCategories(userID string) ([]model.NewsCategory, error)
	ListNewsArticles(userID string, categoryID string, keyword string, page int, pageSize int) ([]model.NewsArticle, int, error)
	GetNewsArticleDetail(userID string, articleID string) (model.NewsArticle, error)
	ListNewsAttachments(userID string, articleID string) ([]model.NewsAttachment, error)
	ListStockRecommendations(userID string, tradeDate string, page int, pageSize int) ([]model.StockRecommendation, int, error)
	GetStockRecommendationDetail(userID string, recoID string) (model.StockRecommendationDetail, error)
	GetStockRecommendationPerformance(userID string, recoID string) ([]model.RecommendationPerformancePoint, error)
	ListFuturesStrategies(userID string, contract string, status string, page int, pageSize int) ([]model.FuturesStrategy, int, error)
	GetFuturesStrategyDetail(userID string, strategyID string) (model.FuturesStrategy, error)
	ListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error)
	CreateMembershipOrder(userID string, productID string, payChannel string) (model.MembershipOrderAdmin, error)
	ListMembershipOrders(userID string, status string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error)
	GetRewardWallet(userID string) (model.RewardWallet, error)
	ListRewardWalletTxns(userID string, page int, pageSize int) ([]model.RewardWalletTxn, int, error)
	CreateWithdrawRequest(userID string, amount float64) (string, error)
	HandlePaymentCallback(channel string, orderNo string, channelTxnNo string, idempotencyKey string, sign string, signVerified bool) error
	ListArbitrageOpportunities(typeFilter string, page int, pageSize int) ([]model.ArbitrageOpportunity, int, error)
	ListFuturesArbitrage(typeFilter string, page int, pageSize int) ([]model.ArbitrageRecommendation, int, error)
	GetFuturesArbitrageDetail(id string) (model.ArbitrageRecommendation, error)
	CreateFuturesAlert(userID string, contract string, alertType string, threshold float64) (string, error)
	ListFuturesReviews(page int, pageSize int) ([]model.FuturesReview, int, error)
	ListMarketEvents(eventType string, page int, pageSize int) ([]model.MarketEvent, int, error)
	GetMarketEventDetail(id string) (model.MarketEvent, error)
	GetFuturesGuidance(contract string) (model.FuturesGuidance, error)
	ListPublicHoldings(symbol string, page int, pageSize int) ([]model.PublicHolding, int, error)
	ListPublicFuturesPositions(contract string, page int, pageSize int) ([]model.PublicFuturesPosition, int, error)

	AdminListInviteRecords(status string, page int, pageSize int) ([]model.InviteRecord, int, error)
	AdminListRewardRecords(status string, page int, pageSize int) ([]model.RewardRecord, int, error)
	AdminReviewRewardRecord(id string, status string, reason string) error
	AdminListReconciliation(page int, pageSize int) ([]model.ReconciliationRecord, int, error)
	AdminRetryReconciliation(batchID string) error
	AdminListRiskRules() ([]model.RiskRule, error)
	AdminCreateRiskRule(ruleCode string, ruleName string, threshold int, status string) (string, error)
	AdminUpdateRiskRule(id string, threshold int, status string) error
	AdminListRiskHits(status string, page int, pageSize int) ([]model.RiskHit, int, error)
	AdminReviewRiskHit(id string, status string, reason string) error
	AdminListWithdrawRequests(page int, pageSize int) ([]model.WithdrawRequestInfo, int, error)
	AdminReviewWithdrawRequest(id string, status string, reason string) error
	AdminListNewsCategories(status string, page int, pageSize int) ([]model.NewsCategory, int, error)
	AdminCreateNewsCategory(name string, slug string, sort int, visibility string, status string) (string, error)
	AdminUpdateNewsCategory(id string, name string, slug string, sort int, visibility string, status string) error
	AdminListNewsArticles(status string, categoryID string, page int, pageSize int) ([]model.NewsArticle, int, error)
	AdminCreateNewsArticle(categoryID string, title string, summary string, content string, visibility string, status string, authorID string) (string, error)
	AdminUpdateNewsArticle(id string, categoryID string, title string, summary string, content string, visibility string, status string) error
	AdminPublishNewsArticle(id string, status string) error
	AdminCreateNewsAttachment(articleID string, fileName string, fileURL string, fileSize int64, mimeType string) (string, error)
	AdminListNewsAttachments(articleID string) ([]model.NewsAttachment, error)
	AdminDeleteNewsAttachment(id string) error
	AdminListStockRecommendations(status string, page int, pageSize int) ([]model.StockRecommendation, int, error)
	AdminCreateStockRecommendation(item model.StockRecommendation) (string, error)
	AdminUpdateStockRecommendationStatus(id string, status string) error
	AdminGenerateDailyStockRecommendations(tradeDate string) (int, error)
	AdminGenerateDailyFuturesStrategies(tradeDate string) (int, error)
	AdminListFuturesStrategies(status string, contract string, page int, pageSize int) ([]model.FuturesStrategy, int, error)
	AdminCreateFuturesStrategy(item model.FuturesStrategy) (string, error)
	AdminUpdateFuturesStrategyStatus(id string, status string) error
	AdminListUsers(status string, kycStatus string, memberLevel string, page int, pageSize int) ([]model.AdminUser, int, error)
	AdminUpdateUserStatus(id string, status string) error
	AdminUpdateUserMemberLevel(id string, memberLevel string) error
	AdminUpdateUserKYCStatus(id string, kycStatus string) error
	AdminDashboardOverview() (model.AdminDashboardOverview, error)
	AdminCreateOperationLog(module string, action string, targetType string, targetID string, operatorUserID string, beforeValue string, afterValue string, reason string) error
	AdminListOperationLogs(module string, action string, operatorUserID string, page int, pageSize int) ([]model.AdminOperationLog, int, error)
	AdminListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error)
	AdminCreateMembershipProduct(name string, price float64, status string, memberLevel string, durationDays int) (string, error)
	AdminUpdateMembershipProductStatus(id string, status string) error
	AdminListMembershipOrders(status string, userID string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error)
	AdminUpdateMembershipOrderStatus(id string, status string) error
	AdminListVIPQuotaConfigs(memberLevel string, status string, page int, pageSize int) ([]model.VIPQuotaConfig, int, error)
	AdminCreateVIPQuotaConfig(item model.VIPQuotaConfig) (string, error)
	AdminUpdateVIPQuotaConfig(id string, item model.VIPQuotaConfig) error
	AdminListUserQuotaUsages(userID string, periodKey string, page int, pageSize int) ([]model.UserQuotaUsage, int, error)
	AdminAdjustUserQuota(userID string, periodKey string, docReadDelta int, newsSubscribeDelta int) error
	AdminListDataSources(page int, pageSize int) ([]model.DataSource, int, error)
	AdminCreateDataSource(item model.DataSource) (string, error)
	AdminUpdateDataSource(sourceKey string, item model.DataSource) error
	AdminDeleteDataSource(sourceKey string) error
	AdminCheckDataSourceHealth(sourceKey string) (model.DataSourceHealthCheck, error)
	AdminListSystemConfigs(keyword string, page int, pageSize int) ([]model.SystemConfig, int, error)
	AdminUpsertSystemConfig(configKey string, configValue string, description string, operator string) error
	AdminListReviewTasks(module string, status string, submitterID string, reviewerID string, page int, pageSize int) ([]model.ReviewTask, int, error)
	AdminSubmitReviewTask(module string, targetID string, submitterID string, reviewerID string, submitNote string) (string, error)
	AdminAssignReviewTask(reviewID string, reviewerID string) error
	AdminReviewTaskDecision(reviewID string, status string, reviewerID string, reviewNote string) error
	GetSchedulerJobNameByRunID(runID string) (string, error)
	AdminListSchedulerJobRuns(jobName string, status string, page int, pageSize int) ([]model.SchedulerJobRun, int, error)
	AdminCreateSchedulerJobRun(jobName string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error)
	AdminRetrySchedulerJobRun(runID string, status string, resultSummary string, errorMessage string, operatorID string) (string, error)
	AdminListSchedulerJobDefinitions(status string, module string, page int, pageSize int) ([]model.SchedulerJobDefinition, int, error)
	AdminCreateSchedulerJobDefinition(item model.SchedulerJobDefinition, operatorID string) (string, error)
	AdminUpdateSchedulerJobDefinition(id string, item model.SchedulerJobDefinition, operatorID string) error
	AdminUpdateSchedulerJobDefinitionStatus(id string, status string, operatorID string) error
	AdminListWorkflowMessages(module string, eventType string, isRead string, receiverID string, page int, pageSize int) ([]model.WorkflowMessage, int, error)
	AdminCountUnreadWorkflowMessages(module string, eventType string, receiverID string) (int, error)
	AdminUpdateWorkflowMessageRead(id string, isRead bool) error
	AdminBulkReadWorkflowMessages(module string, eventType string, receiverID string) (int64, error)
	AdminCreateWorkflowMessage(reviewID string, targetID string, module string, receiverID string, senderID string, eventType string, title string, content string) error
	AdminGetWorkflowMetrics(module string, receiverID string) (model.WorkflowMetrics, error)
	AdminGetSchedulerJobMetrics(jobName string) (model.SchedulerJobMetrics, error)
}
