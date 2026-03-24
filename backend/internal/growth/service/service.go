package service

import (
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/growth/repo"
)

type GrowthService interface {
	ListBrowseHistory(userID string, contentType string, page int, pageSize int) ([]model.BrowseHistory, int, error)
	DeleteBrowseHistoryItem(userID string, id string) error
	ClearBrowseHistory(userID string) error
	ListRechargeRecords(userID string, status string, page int, pageSize int) ([]model.RechargeRecord, int, error)
	ListShareLinks(userID string) ([]model.ShareLink, error)
	CreateShareLink(userID string, channel string, expiredAt string) (model.ShareLink, error)
	ListInviteRecords(userID string, page int, pageSize int) ([]model.InviteRecord, int, error)
	GetUserInviteSummary(userID string) (model.InviteSummary, error)
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
	GetStockRecommendationInsight(userID string, recoID string) (model.StockRecommendationInsight, error)
	GetStockRecommendationVersionHistory(userID string, recoID string) ([]model.StrategyVersionHistoryItem, error)
	ListFuturesStrategies(userID string, contract string, status string, page int, pageSize int) ([]model.FuturesStrategy, int, error)
	GetFuturesStrategyDetail(userID string, strategyID string) (model.FuturesStrategy, error)
	GetFuturesStrategyInsight(userID string, strategyID string) (model.FuturesStrategyInsight, error)
	GetFuturesStrategyVersionHistory(userID string, strategyID string) ([]model.StrategyVersionHistoryItem, error)
	ListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error)
	CreateMembershipOrder(userID string, productID string, payChannel string) (model.MembershipOrderAdmin, error)
	ListMembershipOrders(userID string, status string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error)
	TrackExperimentEvent(item model.ExperimentEvent) error
	BindMembershipOrderExperiment(orderNo string, item model.ExperimentOrderAttribution) error
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
	AdminGetNewsArticleDetail(id string) (model.NewsArticle, error)
	AdminCreateNewsArticle(categoryID string, title string, summary string, content string, coverURL string, visibility string, status string, authorID string) (string, error)
	AdminUpdateNewsArticle(id string, categoryID string, title string, summary string, content string, coverURL string, visibility string, status string) error
	AdminPublishNewsArticle(id string, status string) error
	AdminCreateNewsAttachment(articleID string, fileName string, fileURL string, fileSize int64, mimeType string) (string, error)
	AdminListNewsAttachments(articleID string) ([]model.NewsAttachment, error)
	AdminDeleteNewsAttachment(id string) error
	AdminListStockRecommendations(status string, page int, pageSize int) ([]model.StockRecommendation, int, error)
	AdminCreateStockRecommendation(item model.StockRecommendation) (string, error)
	AdminUpdateStockRecommendationStatus(id string, status string) error
	AdminSyncStockQuotes(sourceKey string, symbols []string, days int) (int, error)
	AdminSyncStockQuotesDetailed(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error)
	AdminSyncFuturesQuotes(sourceKey string, contracts []string, days int) (model.MarketSyncResult, error)
	AdminSyncFuturesInventory(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error)
	AdminSyncMarketNews(sourceKey string, symbols []string, days int, limit int) (model.MarketSyncResult, error)
	BuildStrategyEngineStockSelectionContext(input model.StrategyEngineStockSelectionContextRequest) (model.StrategyEngineStockSelectionContextResponse, error)
	BuildStrategyEngineFuturesStrategyContext(input model.StrategyEngineFuturesStrategyContextRequest) (model.StrategyEngineFuturesStrategyContextResponse, error)
	AdminSyncDocFastNewsIncremental(batchSize int) (string, error)
	AdminSyncTushareNewsIncremental(batchSize int) (string, error)
	AdminSyncTushareNewsIncrementalWithOptions(opts model.TushareNewsSyncOptions) (string, []model.NewsSyncRunDetail, error)
	AdminRunVIPMembershipLifecycle() (string, error)
	AdminGetQuantTopStocks(limit int, lookbackDays int) ([]model.StockQuantScore, error)
	AdminGetQuantEvaluation(windowDays int, topN int) (model.StockQuantEvaluationSummary, []model.StockQuantEvaluationPoint, []model.StockQuantRiskPerformance, []model.StockQuantRotationPoint, error)
	AdminGenerateDailyStockRecommendations(tradeDate string) (model.AdminDailyStockRecommendationGenerationResult, error)
	AdminGetStockSelectionOverview() (model.AdminStockSelectionOverview, error)
	AdminListStockSelectionRuns(status string, reviewStatus string, profileID string, page int, pageSize int) ([]model.StockSelectionRun, int, error)
	AdminCreateStockSelectionRun(input model.StockSelectionRunCreateRequest, operator string) (model.StockSelectionRun, error)
	AdminGetStockSelectionRun(runID string) (model.StockSelectionRun, error)
	AdminGetFuturesSelectionOverview() (model.AdminFuturesSelectionOverview, error)
	AdminListFuturesSelectionRuns(status string, reviewStatus string, profileID string, page int, pageSize int) ([]model.FuturesSelectionRun, int, error)
	AdminCreateFuturesSelectionRun(input model.FuturesSelectionRunCreateRequest, operator string) (model.FuturesSelectionRun, error)
	AdminGetFuturesSelectionRun(runID string) (model.FuturesSelectionRun, error)
	AdminCompareFuturesSelectionRuns(runIDs []string) (model.FuturesSelectionRunCompareResult, error)
	AdminListFuturesSelectionProfiles(status string, page int, pageSize int) ([]model.FuturesSelectionProfile, int, error)
	AdminListFuturesSelectionProfileVersions(profileID string) ([]model.FuturesSelectionProfileVersion, error)
	AdminCreateFuturesSelectionProfile(item model.FuturesSelectionProfile, changeNote string) (model.FuturesSelectionProfile, error)
	AdminUpdateFuturesSelectionProfile(id string, item model.FuturesSelectionProfile, changeNote string) (model.FuturesSelectionProfile, error)
	AdminPublishFuturesSelectionProfile(id string, operator string) (model.FuturesSelectionProfile, error)
	AdminRollbackFuturesSelectionProfile(id string, versionNo int, changeNote string, operator string) (model.FuturesSelectionProfile, error)
	AdminListFuturesSelectionProfileTemplates(status string, page int, pageSize int) ([]model.FuturesSelectionProfileTemplate, int, error)
	AdminCreateFuturesSelectionProfileTemplate(item model.FuturesSelectionProfileTemplate) (model.FuturesSelectionProfileTemplate, error)
	AdminUpdateFuturesSelectionProfileTemplate(id string, item model.FuturesSelectionProfileTemplate) (model.FuturesSelectionProfileTemplate, error)
	AdminSetDefaultFuturesSelectionProfileTemplate(id string, operator string) (model.FuturesSelectionProfileTemplate, error)
	AdminListFuturesSelectionEvaluationLeaderboard(templateID string, profileID string, marketRegime string) ([]model.FuturesSelectionEvaluationLeaderboardItem, error)
	AdminListFuturesSelectionRunCandidates(runID string) ([]model.FuturesSelectionCandidateSnapshot, error)
	AdminListFuturesSelectionRunPortfolio(runID string) ([]model.FuturesSelectionPortfolioEntry, error)
	AdminListFuturesSelectionRunEvidence(runID string, contract string) ([]model.FuturesSelectionRunEvidence, error)
	AdminListFuturesSelectionRunEvaluations(runID string, contract string) ([]model.FuturesSelectionRunEvaluation, error)
	AdminApproveFuturesSelectionReview(runID string, operator string, reviewNote string, force bool, overrideReason string) (model.FuturesSelectionPublishReview, error)
	AdminRejectFuturesSelectionReview(runID string, operator string, reviewNote string) (model.FuturesSelectionPublishReview, error)
	AdminCompareStockSelectionRuns(runIDs []string) (model.StockSelectionRunCompareResult, error)
	AdminListStockSelectionProfiles(status string, page int, pageSize int) ([]model.StockSelectionProfile, int, error)
	AdminListStockSelectionProfileVersions(profileID string) ([]model.StockSelectionProfileVersion, error)
	AdminCreateStockSelectionProfile(item model.StockSelectionProfile, changeNote string) (model.StockSelectionProfile, error)
	AdminUpdateStockSelectionProfile(id string, item model.StockSelectionProfile, changeNote string) (model.StockSelectionProfile, error)
	AdminPublishStockSelectionProfile(id string, operator string) (model.StockSelectionProfile, error)
	AdminRollbackStockSelectionProfile(id string, versionNo int, changeNote string, operator string) (model.StockSelectionProfile, error)
	AdminListStockSelectionProfileTemplates(status string, page int, pageSize int) ([]model.StockSelectionProfileTemplate, int, error)
	AdminCreateStockSelectionProfileTemplate(item model.StockSelectionProfileTemplate) (model.StockSelectionProfileTemplate, error)
	AdminUpdateStockSelectionProfileTemplate(id string, item model.StockSelectionProfileTemplate) (model.StockSelectionProfileTemplate, error)
	AdminSetDefaultStockSelectionProfileTemplate(id string, operator string) (model.StockSelectionProfileTemplate, error)
	AdminListStockSelectionRunCandidates(runID string) ([]model.StockSelectionCandidateSnapshot, error)
	AdminListStockSelectionRunPortfolio(runID string) ([]model.StockSelectionPortfolioEntry, error)
	AdminListStockSelectionRunEvidence(runID string, symbol string) ([]model.StockSelectionRunEvidence, error)
	AdminListStockSelectionRunEvaluations(runID string, symbol string) ([]model.StockSelectionRunEvaluation, error)
	AdminListStockSelectionEvaluationLeaderboard(templateID string, profileID string, marketRegime string) ([]model.StockSelectionEvaluationLeaderboardItem, error)
	AdminListStockSelectionReviews(status string, page int, pageSize int) ([]model.StockSelectionPublishReview, int, error)
	AdminApproveStockSelectionReview(runID string, operator string, reviewNote string, force bool, overrideReason string) (model.StockSelectionPublishReview, error)
	AdminRejectStockSelectionReview(runID string, operator string, reviewNote string) (model.StockSelectionPublishReview, error)
	AdminGetStrategyGraphSnapshot(snapshotID string) (model.StrategyGraphSnapshot, error)
	AdminQueryStrategyGraphSubgraph(query model.StrategyGraphSubgraphQuery) (model.StrategyGraphSubgraph, error)
	AdminListStrategyEnginePublishHistory(jobType string) ([]model.StrategyEnginePublishRecordSummary, error)
	AdminGetStrategyEnginePublishRecord(publishID string) (model.StrategyEnginePublishRecord, error)
	AdminGetStrategyEnginePublishReplay(publishID string) (model.StrategyEnginePublishReplay, error)
	AdminCompareStrategyEnginePublishVersions(leftPublishID string, rightPublishID string) (model.StrategyEnginePublishCompareResult, error)
	AdminListStrategySeedSets(targetType string, status string, page int, pageSize int) ([]model.StrategySeedSet, int, error)
	AdminCreateStrategySeedSet(item model.StrategySeedSet) (string, error)
	AdminUpdateStrategySeedSet(id string, item model.StrategySeedSet) error
	AdminListStrategyAgentProfiles(targetType string, status string, page int, pageSize int) ([]model.StrategyAgentProfile, int, error)
	AdminCreateStrategyAgentProfile(item model.StrategyAgentProfile) (string, error)
	AdminUpdateStrategyAgentProfile(id string, item model.StrategyAgentProfile) error
	AdminListStrategyScenarioTemplates(targetType string, status string, page int, pageSize int) ([]model.StrategyScenarioTemplate, int, error)
	AdminCreateStrategyScenarioTemplate(item model.StrategyScenarioTemplate) (string, error)
	AdminUpdateStrategyScenarioTemplate(id string, item model.StrategyScenarioTemplate) error
	AdminListStrategyPublishPolicies(targetType string, status string, page int, pageSize int) ([]model.StrategyPublishPolicy, int, error)
	AdminCreateStrategyPublishPolicy(item model.StrategyPublishPolicy) (string, error)
	AdminUpdateStrategyPublishPolicy(id string, item model.StrategyPublishPolicy) error
	AdminListStrategyEngineJobs(jobType string, status string, page int, pageSize int) ([]model.StrategyEngineJobRecord, int, error)
	AdminGetStrategyEngineJob(jobID string) (model.StrategyEngineJobRecord, error)
	AdminPublishStrategyEngineJob(jobID string, operator string, force bool, overrideReason string) (model.StrategyEnginePublishRecord, error)
	AdminGenerateDailyFuturesStrategies(tradeDate string) (model.AdminDailyFuturesStrategyGenerationResult, error)
	AdminListFuturesStrategies(status string, contract string, page int, pageSize int) ([]model.FuturesStrategy, int, error)
	AdminCreateFuturesStrategy(item model.FuturesStrategy) (string, error)
	AdminUpdateFuturesStrategyStatus(id string, status string) error
	AdminListMarketEvents(eventType string, symbol string, page int, pageSize int) ([]model.MarketEvent, int, error)
	AdminCreateMarketEvent(item model.MarketEvent) (string, error)
	AdminUpdateMarketEvent(id string, item model.MarketEvent) error
	AdminListMarketRhythmTasks(taskDate string) ([]model.MarketRhythmTask, error)
	AdminEnsureMarketRhythmTasks(taskDate string) ([]model.MarketRhythmTask, error)
	AdminUpdateMarketRhythmTask(id string, owner string, notes string, sourceLinks []string, status string) (model.MarketRhythmTask, error)
	AdminUpdateMarketRhythmTaskStatus(id string, status string, owner string, notes string) (model.MarketRhythmTask, error)
	AdminListUsers(status string, kycStatus string, memberLevel string, registrationSource string, page int, pageSize int) ([]model.AdminUser, int, error)
	AdminGetUserSourceSummary(status string, kycStatus string, memberLevel string, registrationSource string) (model.AdminUserSourceSummary, error)
	AdminListBrowseHistories(userID string, contentType string, keyword string, page int, pageSize int) ([]model.AdminBrowseHistory, int, error)
	AdminGetBrowseHistorySummary() (model.AdminBrowseHistorySummary, error)
	AdminGetBrowseHistoryTrend(days int) ([]model.AdminBrowseTrendPoint, error)
	AdminListBrowseUserSegments(limit int) ([]model.AdminBrowseUserSegment, error)
	AdminListUserMessages(userID string, messageType string, readStatus string, page int, pageSize int) ([]model.AdminUserMessage, int, error)
	AdminCreateUserMessages(userIDs []string, title string, content string, messageType string) (int, []model.AdminMessageSendFailure, error)
	AdminUpdateUserStatus(id string, status string) error
	AdminUpdateUserMemberLevel(id string, memberLevel string) error
	AdminUpdateUserKYCStatus(id string, kycStatus string) error
	AdminResetUserPasswordHash(id string, passwordHash string) error
	AdminDashboardOverview() (model.AdminDashboardOverview, error)
	AdminCreateOperationLog(module string, action string, targetType string, targetID string, operatorUserID string, beforeValue string, afterValue string, reason string) error
	AdminListOperationLogs(module string, action string, operatorUserID string, page int, pageSize int) ([]model.AdminOperationLog, int, error)
	AdminListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error)
	AdminCreateMembershipProduct(name string, price float64, status string, memberLevel string, durationDays int) (string, error)
	AdminUpdateMembershipProduct(id string, name string, price float64, status string, memberLevel string, durationDays int) error
	AdminUpdateMembershipProductStatus(id string, status string) error
	AdminListMembershipOrders(status string, userID string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error)
	AdminUpdateMembershipOrderStatus(id string, status string) error
	AdminGetExperimentAnalyticsSummary(days int) (model.AdminExperimentAnalyticsSummary, error)
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
	AdminBatchCheckDataSourceHealth(sourceKeys []string) ([]model.DataSourceHealthCheck, error)
	AdminListDataSourceHealthLogs(sourceKey string, page int, pageSize int) ([]model.DataSourceHealthLog, int, error)
	AdminListMarketDataQualityLogs(assetClass string, dataKind string, severity string, issueCode string, hours int, page int, pageSize int) ([]model.MarketDataQualityLog, int, error)
	AdminGetMarketDataQualitySummary(assetClass string, hours int) (model.MarketDataQualitySummary, error)
	AdminGetMarketDerivedTruthSummary(assetClass string) (*model.MarketDerivedTruthSummary, error)
	AdminRebuildMarketDerivedTruth(assetClass string, tradeDate string, days int) (model.MarketDerivedTruthRebuildResult, error)
	AdminCreateMarketDataBackfillRun(input model.MarketBackfillCreateInput, operator string) (model.MarketBackfillRun, error)
	AdminListMarketDataBackfillRuns(status string, runType string, assetType string, sourceKey string, page int, pageSize int) ([]model.MarketBackfillRun, int, error)
	AdminGetMarketDataBackfillRun(id string) (model.MarketBackfillRun, error)
	AdminListMarketDataBackfillRunDetails(runID string, stage string, assetType string, status string, page int, pageSize int) ([]model.MarketBackfillRunDetail, int, error)
	AdminRetryMarketDataBackfillRun(runID string, input model.MarketBackfillRetryInput, operator string) (model.MarketBackfillRun, error)
	AdminListMarketUniverseSnapshots(page int, pageSize int) ([]model.MarketUniverseSnapshot, int, error)
	AdminGetMarketUniverseSnapshot(id string) (model.MarketUniverseSnapshot, []model.MarketUniverseSnapshotItem, error)
	AdminGetMarketCoverageSummary() (model.MarketCoverageSummary, error)
	AdminListSystemConfigs(keyword string, page int, pageSize int) ([]model.SystemConfig, int, error)
	AdminUpsertSystemConfig(configKey string, configValue string, description string, operator string) error
	AdminListReviewTasks(module string, status string, submitterID string, reviewerID string, page int, pageSize int) ([]model.ReviewTask, int, error)
	AdminSubmitReviewTask(module string, targetID string, submitterID string, reviewerID string, submitNote string) (string, error)
	AdminAssignReviewTask(reviewID string, reviewerID string) error
	AdminReviewTaskDecision(reviewID string, status string, reviewerID string, reviewNote string) error
	GetSchedulerJobNameByRunID(runID string) (string, error)
	AdminListSchedulerJobRuns(jobName string, status string, page int, pageSize int) ([]model.SchedulerJobRun, int, error)
	AdminListNewsSyncRunDetails(runID string, syncType string, source string, symbol string, status string, page int, pageSize int) ([]model.NewsSyncRunDetail, int, error)
	AdminCreateNewsSyncRunDetails(runID string, details []model.NewsSyncRunDetail) error
	AdminCreateSchedulerJobRun(jobName string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error)
	AdminRetrySchedulerJobRun(runID string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error)
	AdminListSchedulerJobDefinitions(status string, module string, page int, pageSize int) ([]model.SchedulerJobDefinition, int, error)
	AdminCreateSchedulerJobDefinition(item model.SchedulerJobDefinition, operatorID string) (string, error)
	AdminUpdateSchedulerJobDefinition(id string, item model.SchedulerJobDefinition, operatorID string) error
	AdminUpdateSchedulerJobDefinitionStatus(id string, status string, operatorID string) error
	AdminDeleteSchedulerJobDefinition(id string) error
	AdminListWorkflowMessages(module string, eventType string, isRead string, receiverID string, page int, pageSize int) ([]model.WorkflowMessage, int, error)
	AdminCountUnreadWorkflowMessages(module string, eventType string, receiverID string) (int, error)
	AdminUpdateWorkflowMessageRead(id string, isRead bool) error
	AdminBulkReadWorkflowMessages(module string, eventType string, receiverID string) (int64, error)
	AdminCreateWorkflowMessage(reviewID string, targetID string, module string, receiverID string, senderID string, eventType string, title string, content string) error
	AdminGetWorkflowMetrics(module string, receiverID string) (model.WorkflowMetrics, error)
	AdminGetSchedulerJobMetrics(jobName string) (model.SchedulerJobMetrics, error)
}

type growthService struct {
	repo repo.GrowthRepo
}

func NewGrowthService(repo repo.GrowthRepo) GrowthService {
	return &growthService{repo: repo}
}

func (s *growthService) ListBrowseHistory(userID string, contentType string, page int, pageSize int) ([]model.BrowseHistory, int, error) {
	return s.repo.ListBrowseHistory(userID, contentType, page, pageSize)
}

func (s *growthService) DeleteBrowseHistoryItem(userID string, id string) error {
	return s.repo.DeleteBrowseHistoryItem(userID, id)
}

func (s *growthService) ClearBrowseHistory(userID string) error {
	return s.repo.ClearBrowseHistory(userID)
}

func (s *growthService) BuildStrategyEngineStockSelectionContext(input model.StrategyEngineStockSelectionContextRequest) (model.StrategyEngineStockSelectionContextResponse, error) {
	return s.repo.BuildStrategyEngineStockSelectionContext(input.Normalized())
}

func (s *growthService) BuildStrategyEngineFuturesStrategyContext(input model.StrategyEngineFuturesStrategyContextRequest) (model.StrategyEngineFuturesStrategyContextResponse, error) {
	return s.repo.BuildStrategyEngineFuturesStrategyContext(input)
}

func (s *growthService) ListRechargeRecords(userID string, status string, page int, pageSize int) ([]model.RechargeRecord, int, error) {
	return s.repo.ListRechargeRecords(userID, status, page, pageSize)
}

func (s *growthService) ListShareLinks(userID string) ([]model.ShareLink, error) {
	return s.repo.ListShareLinks(userID)
}

func (s *growthService) CreateShareLink(userID string, channel string, expiredAt string) (model.ShareLink, error) {
	return s.repo.CreateShareLink(userID, channel, expiredAt)
}

func (s *growthService) ListInviteRecords(userID string, page int, pageSize int) ([]model.InviteRecord, int, error) {
	return s.repo.ListInviteRecords(userID, page, pageSize)
}

func (s *growthService) GetUserInviteSummary(userID string) (model.InviteSummary, error) {
	return s.repo.GetUserInviteSummary(userID)
}

func (s *growthService) ListRewardRecords(userID string, page int, pageSize int) ([]model.RewardRecord, int, error) {
	return s.repo.ListRewardRecords(userID, page, pageSize)
}

func (s *growthService) GetUserProfile(userID string) (model.UserProfile, error) {
	return s.repo.GetUserProfile(userID)
}

func (s *growthService) UpdateUserProfileEmail(userID string, email string) error {
	return s.repo.UpdateUserProfileEmail(userID, email)
}

func (s *growthService) SubmitUserKYC(userID string, realName string, idNumber string, provider string) (string, error) {
	return s.repo.SubmitUserKYC(userID, realName, idNumber, provider)
}

func (s *growthService) ListSubscriptions(userID string, page int, pageSize int) ([]model.Subscription, int, error) {
	return s.repo.ListSubscriptions(userID, page, pageSize)
}

func (s *growthService) CreateSubscription(userID string, subType string, scope string, frequency string) (string, error) {
	return s.repo.CreateSubscription(userID, subType, scope, frequency)
}

func (s *growthService) UpdateSubscription(userID string, id string, frequency string, status string) error {
	return s.repo.UpdateSubscription(userID, id, frequency, status)
}

func (s *growthService) ListMessages(userID string, page int, pageSize int) ([]model.UserMessage, int, error) {
	return s.repo.ListMessages(userID, page, pageSize)
}

func (s *growthService) MarkMessageRead(userID string, id string) error {
	return s.repo.MarkMessageRead(userID, id)
}

func (s *growthService) GetUserAccessProfile(userID string) (model.UserAccessProfile, error) {
	return s.repo.GetUserAccessProfile(userID)
}

func (s *growthService) GetMembershipQuota(userID string) (model.MembershipQuota, error) {
	return s.repo.GetMembershipQuota(userID)
}

func (s *growthService) GetAttachmentFileInfo(userID string, attachmentID string) (model.AttachmentFileInfo, error) {
	return s.repo.GetAttachmentFileInfo(userID, attachmentID)
}

func (s *growthService) LogAttachmentDownload(userID string, attachmentID string, articleID string) error {
	return s.repo.LogAttachmentDownload(userID, attachmentID, articleID)
}

func (s *growthService) ListNewsCategories(userID string) ([]model.NewsCategory, error) {
	return s.repo.ListNewsCategories(userID)
}

func (s *growthService) ListNewsArticles(userID string, categoryID string, keyword string, page int, pageSize int) ([]model.NewsArticle, int, error) {
	return s.repo.ListNewsArticles(userID, categoryID, keyword, page, pageSize)
}

func (s *growthService) GetNewsArticleDetail(userID string, articleID string) (model.NewsArticle, error) {
	return s.repo.GetNewsArticleDetail(userID, articleID)
}

func (s *growthService) ListNewsAttachments(userID string, articleID string) ([]model.NewsAttachment, error) {
	return s.repo.ListNewsAttachments(userID, articleID)
}

func (s *growthService) ListStockRecommendations(userID string, tradeDate string, page int, pageSize int) ([]model.StockRecommendation, int, error) {
	return s.repo.ListStockRecommendations(userID, tradeDate, page, pageSize)
}

func (s *growthService) GetStockRecommendationDetail(userID string, recoID string) (model.StockRecommendationDetail, error) {
	return s.repo.GetStockRecommendationDetail(userID, recoID)
}

func (s *growthService) GetStockRecommendationPerformance(userID string, recoID string) ([]model.RecommendationPerformancePoint, error) {
	return s.repo.GetStockRecommendationPerformance(userID, recoID)
}

func (s *growthService) GetStockRecommendationInsight(userID string, recoID string) (model.StockRecommendationInsight, error) {
	return s.repo.GetStockRecommendationInsight(userID, recoID)
}

func (s *growthService) GetStockRecommendationVersionHistory(userID string, recoID string) ([]model.StrategyVersionHistoryItem, error) {
	return s.repo.GetStockRecommendationVersionHistory(userID, recoID)
}

func (s *growthService) ListFuturesStrategies(userID string, contract string, status string, page int, pageSize int) ([]model.FuturesStrategy, int, error) {
	return s.repo.ListFuturesStrategies(userID, contract, status, page, pageSize)
}

func (s *growthService) GetFuturesStrategyDetail(userID string, strategyID string) (model.FuturesStrategy, error) {
	return s.repo.GetFuturesStrategyDetail(userID, strategyID)
}

func (s *growthService) GetFuturesStrategyInsight(userID string, strategyID string) (model.FuturesStrategyInsight, error) {
	return s.repo.GetFuturesStrategyInsight(userID, strategyID)
}

func (s *growthService) GetFuturesStrategyVersionHistory(userID string, strategyID string) ([]model.StrategyVersionHistoryItem, error) {
	return s.repo.GetFuturesStrategyVersionHistory(userID, strategyID)
}

func (s *growthService) ListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error) {
	return s.repo.ListMembershipProducts(status, page, pageSize)
}

func (s *growthService) CreateMembershipOrder(userID string, productID string, payChannel string) (model.MembershipOrderAdmin, error) {
	return s.repo.CreateMembershipOrder(userID, productID, payChannel)
}

func (s *growthService) ListMembershipOrders(userID string, status string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error) {
	return s.repo.ListMembershipOrders(userID, status, page, pageSize)
}

func (s *growthService) TrackExperimentEvent(item model.ExperimentEvent) error {
	return s.repo.TrackExperimentEvent(item)
}

func (s *growthService) BindMembershipOrderExperiment(orderNo string, item model.ExperimentOrderAttribution) error {
	return s.repo.BindMembershipOrderExperiment(orderNo, item)
}

func (s *growthService) GetRewardWallet(userID string) (model.RewardWallet, error) {
	return s.repo.GetRewardWallet(userID)
}

func (s *growthService) ListRewardWalletTxns(userID string, page int, pageSize int) ([]model.RewardWalletTxn, int, error) {
	return s.repo.ListRewardWalletTxns(userID, page, pageSize)
}

func (s *growthService) CreateWithdrawRequest(userID string, amount float64) (string, error) {
	return s.repo.CreateWithdrawRequest(userID, amount)
}

func (s *growthService) HandlePaymentCallback(channel string, orderNo string, channelTxnNo string, idempotencyKey string, sign string, signVerified bool) error {
	return s.repo.HandlePaymentCallback(channel, orderNo, channelTxnNo, idempotencyKey, sign, signVerified)
}

func (s *growthService) ListArbitrageOpportunities(typeFilter string, page int, pageSize int) ([]model.ArbitrageOpportunity, int, error) {
	return s.repo.ListArbitrageOpportunities(typeFilter, page, pageSize)
}

func (s *growthService) ListFuturesArbitrage(typeFilter string, page int, pageSize int) ([]model.ArbitrageRecommendation, int, error) {
	return s.repo.ListFuturesArbitrage(typeFilter, page, pageSize)
}

func (s *growthService) GetFuturesArbitrageDetail(id string) (model.ArbitrageRecommendation, error) {
	return s.repo.GetFuturesArbitrageDetail(id)
}

func (s *growthService) CreateFuturesAlert(userID string, contract string, alertType string, threshold float64) (string, error) {
	return s.repo.CreateFuturesAlert(userID, contract, alertType, threshold)
}

func (s *growthService) ListFuturesReviews(page int, pageSize int) ([]model.FuturesReview, int, error) {
	return s.repo.ListFuturesReviews(page, pageSize)
}

func (s *growthService) ListMarketEvents(eventType string, page int, pageSize int) ([]model.MarketEvent, int, error) {
	return s.repo.ListMarketEvents(eventType, page, pageSize)
}

func (s *growthService) GetMarketEventDetail(id string) (model.MarketEvent, error) {
	return s.repo.GetMarketEventDetail(id)
}

func (s *growthService) GetFuturesGuidance(contract string) (model.FuturesGuidance, error) {
	return s.repo.GetFuturesGuidance(contract)
}

func (s *growthService) ListPublicHoldings(symbol string, page int, pageSize int) ([]model.PublicHolding, int, error) {
	return s.repo.ListPublicHoldings(symbol, page, pageSize)
}

func (s *growthService) ListPublicFuturesPositions(contract string, page int, pageSize int) ([]model.PublicFuturesPosition, int, error) {
	return s.repo.ListPublicFuturesPositions(contract, page, pageSize)
}

func (s *growthService) AdminListInviteRecords(status string, page int, pageSize int) ([]model.InviteRecord, int, error) {
	return s.repo.AdminListInviteRecords(status, page, pageSize)
}

func (s *growthService) AdminListRewardRecords(status string, page int, pageSize int) ([]model.RewardRecord, int, error) {
	return s.repo.AdminListRewardRecords(status, page, pageSize)
}

func (s *growthService) AdminReviewRewardRecord(id string, status string, reason string) error {
	return s.repo.AdminReviewRewardRecord(id, status, reason)
}

func (s *growthService) AdminListReconciliation(page int, pageSize int) ([]model.ReconciliationRecord, int, error) {
	return s.repo.AdminListReconciliation(page, pageSize)
}

func (s *growthService) AdminRetryReconciliation(batchID string) error {
	return s.repo.AdminRetryReconciliation(batchID)
}

func (s *growthService) AdminListRiskRules() ([]model.RiskRule, error) {
	return s.repo.AdminListRiskRules()
}

func (s *growthService) AdminCreateRiskRule(ruleCode string, ruleName string, threshold int, status string) (string, error) {
	return s.repo.AdminCreateRiskRule(ruleCode, ruleName, threshold, status)
}

func (s *growthService) AdminUpdateRiskRule(id string, threshold int, status string) error {
	return s.repo.AdminUpdateRiskRule(id, threshold, status)
}

func (s *growthService) AdminListRiskHits(status string, page int, pageSize int) ([]model.RiskHit, int, error) {
	return s.repo.AdminListRiskHits(status, page, pageSize)
}

func (s *growthService) AdminReviewRiskHit(id string, status string, reason string) error {
	return s.repo.AdminReviewRiskHit(id, status, reason)
}

func (s *growthService) AdminListWithdrawRequests(page int, pageSize int) ([]model.WithdrawRequestInfo, int, error) {
	return s.repo.AdminListWithdrawRequests(page, pageSize)
}

func (s *growthService) AdminReviewWithdrawRequest(id string, status string, reason string) error {
	return s.repo.AdminReviewWithdrawRequest(id, status, reason)
}

func (s *growthService) AdminListNewsCategories(status string, page int, pageSize int) ([]model.NewsCategory, int, error) {
	return s.repo.AdminListNewsCategories(status, page, pageSize)
}

func (s *growthService) AdminCreateNewsCategory(name string, slug string, sort int, visibility string, status string) (string, error) {
	return s.repo.AdminCreateNewsCategory(name, slug, sort, visibility, status)
}

func (s *growthService) AdminUpdateNewsCategory(id string, name string, slug string, sort int, visibility string, status string) error {
	return s.repo.AdminUpdateNewsCategory(id, name, slug, sort, visibility, status)
}

func (s *growthService) AdminListNewsArticles(status string, categoryID string, page int, pageSize int) ([]model.NewsArticle, int, error) {
	return s.repo.AdminListNewsArticles(status, categoryID, page, pageSize)
}

func (s *growthService) AdminGetNewsArticleDetail(id string) (model.NewsArticle, error) {
	return s.repo.AdminGetNewsArticleDetail(id)
}

func (s *growthService) AdminCreateNewsArticle(categoryID string, title string, summary string, content string, coverURL string, visibility string, status string, authorID string) (string, error) {
	return s.repo.AdminCreateNewsArticle(categoryID, title, summary, content, coverURL, visibility, status, authorID)
}

func (s *growthService) AdminUpdateNewsArticle(id string, categoryID string, title string, summary string, content string, coverURL string, visibility string, status string) error {
	return s.repo.AdminUpdateNewsArticle(id, categoryID, title, summary, content, coverURL, visibility, status)
}

func (s *growthService) AdminPublishNewsArticle(id string, status string) error {
	return s.repo.AdminPublishNewsArticle(id, status)
}

func (s *growthService) AdminCreateNewsAttachment(articleID string, fileName string, fileURL string, fileSize int64, mimeType string) (string, error) {
	return s.repo.AdminCreateNewsAttachment(articleID, fileName, fileURL, fileSize, mimeType)
}

func (s *growthService) AdminListNewsAttachments(articleID string) ([]model.NewsAttachment, error) {
	return s.repo.AdminListNewsAttachments(articleID)
}

func (s *growthService) AdminDeleteNewsAttachment(id string) error {
	return s.repo.AdminDeleteNewsAttachment(id)
}

func (s *growthService) AdminListStockRecommendations(status string, page int, pageSize int) ([]model.StockRecommendation, int, error) {
	return s.repo.AdminListStockRecommendations(status, page, pageSize)
}

func (s *growthService) AdminCreateStockRecommendation(item model.StockRecommendation) (string, error) {
	return s.repo.AdminCreateStockRecommendation(item)
}

func (s *growthService) AdminUpdateStockRecommendationStatus(id string, status string) error {
	return s.repo.AdminUpdateStockRecommendationStatus(id, status)
}

func (s *growthService) AdminSyncStockQuotes(sourceKey string, symbols []string, days int) (int, error) {
	return s.repo.AdminSyncStockQuotes(sourceKey, symbols, days)
}

func (s *growthService) AdminSyncStockQuotesDetailed(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error) {
	return s.repo.AdminSyncStockQuotesDetailed(sourceKey, symbols, days)
}

func (s *growthService) AdminSyncFuturesQuotes(sourceKey string, contracts []string, days int) (model.MarketSyncResult, error) {
	return s.repo.AdminSyncFuturesQuotes(sourceKey, contracts, days)
}

func (s *growthService) AdminSyncFuturesInventory(sourceKey string, symbols []string, days int) (model.MarketSyncResult, error) {
	return s.repo.AdminSyncFuturesInventory(sourceKey, symbols, days)
}

func (s *growthService) AdminSyncMarketNews(sourceKey string, symbols []string, days int, limit int) (model.MarketSyncResult, error) {
	return s.repo.AdminSyncMarketNews(sourceKey, symbols, days, limit)
}

func (s *growthService) AdminSyncDocFastNewsIncremental(batchSize int) (string, error) {
	return s.repo.AdminSyncDocFastNewsIncremental(batchSize)
}

func (s *growthService) AdminSyncTushareNewsIncremental(batchSize int) (string, error) {
	return s.repo.AdminSyncTushareNewsIncremental(batchSize)
}

func (s *growthService) AdminSyncTushareNewsIncrementalWithOptions(opts model.TushareNewsSyncOptions) (string, []model.NewsSyncRunDetail, error) {
	return s.repo.AdminSyncTushareNewsIncrementalWithOptions(opts)
}

func (s *growthService) AdminRunVIPMembershipLifecycle() (string, error) {
	return s.repo.AdminRunVIPMembershipLifecycle()
}

func (s *growthService) AdminGetQuantTopStocks(limit int, lookbackDays int) ([]model.StockQuantScore, error) {
	return s.repo.AdminGetQuantTopStocks(limit, lookbackDays)
}

func (s *growthService) AdminGetQuantEvaluation(windowDays int, topN int) (model.StockQuantEvaluationSummary, []model.StockQuantEvaluationPoint, []model.StockQuantRiskPerformance, []model.StockQuantRotationPoint, error) {
	return s.repo.AdminGetQuantEvaluation(windowDays, topN)
}

func (s *growthService) AdminGenerateDailyStockRecommendations(tradeDate string) (model.AdminDailyStockRecommendationGenerationResult, error) {
	return s.repo.AdminGenerateDailyStockRecommendations(tradeDate)
}

func (s *growthService) AdminGetFuturesSelectionOverview() (model.AdminFuturesSelectionOverview, error) {
	return s.repo.AdminGetFuturesSelectionOverview()
}

func (s *growthService) AdminListFuturesSelectionRuns(status string, reviewStatus string, profileID string, page int, pageSize int) ([]model.FuturesSelectionRun, int, error) {
	return s.repo.AdminListFuturesSelectionRuns(status, reviewStatus, profileID, page, pageSize)
}

func (s *growthService) AdminCreateFuturesSelectionRun(input model.FuturesSelectionRunCreateRequest, operator string) (model.FuturesSelectionRun, error) {
	return s.repo.AdminCreateFuturesSelectionRun(input, operator)
}

func (s *growthService) AdminGetFuturesSelectionRun(runID string) (model.FuturesSelectionRun, error) {
	return s.repo.AdminGetFuturesSelectionRun(runID)
}

func (s *growthService) AdminListStrategyEnginePublishHistory(jobType string) ([]model.StrategyEnginePublishRecordSummary, error) {
	return s.repo.AdminListStrategyEnginePublishHistory(jobType)
}

func (s *growthService) AdminGetStrategyEnginePublishRecord(publishID string) (model.StrategyEnginePublishRecord, error) {
	return s.repo.AdminGetStrategyEnginePublishRecord(publishID)
}

func (s *growthService) AdminGetStrategyEnginePublishReplay(publishID string) (model.StrategyEnginePublishReplay, error) {
	return s.repo.AdminGetStrategyEnginePublishReplay(publishID)
}

func (s *growthService) AdminCompareStrategyEnginePublishVersions(leftPublishID string, rightPublishID string) (model.StrategyEnginePublishCompareResult, error) {
	return s.repo.AdminCompareStrategyEnginePublishVersions(leftPublishID, rightPublishID)
}

func (s *growthService) AdminGenerateDailyFuturesStrategies(tradeDate string) (model.AdminDailyFuturesStrategyGenerationResult, error) {
	return s.repo.AdminGenerateDailyFuturesStrategies(tradeDate)
}

func (s *growthService) AdminListFuturesStrategies(status string, contract string, page int, pageSize int) ([]model.FuturesStrategy, int, error) {
	return s.repo.AdminListFuturesStrategies(status, contract, page, pageSize)
}

func (s *growthService) AdminCreateFuturesStrategy(item model.FuturesStrategy) (string, error) {
	return s.repo.AdminCreateFuturesStrategy(item)
}

func (s *growthService) AdminUpdateFuturesStrategyStatus(id string, status string) error {
	return s.repo.AdminUpdateFuturesStrategyStatus(id, status)
}

func (s *growthService) AdminListMarketEvents(eventType string, symbol string, page int, pageSize int) ([]model.MarketEvent, int, error) {
	return s.repo.AdminListMarketEvents(eventType, symbol, page, pageSize)
}

func (s *growthService) AdminCreateMarketEvent(item model.MarketEvent) (string, error) {
	return s.repo.AdminCreateMarketEvent(item)
}

func (s *growthService) AdminUpdateMarketEvent(id string, item model.MarketEvent) error {
	return s.repo.AdminUpdateMarketEvent(id, item)
}

func (s *growthService) AdminListMarketRhythmTasks(taskDate string) ([]model.MarketRhythmTask, error) {
	return s.repo.AdminListMarketRhythmTasks(taskDate)
}

func (s *growthService) AdminEnsureMarketRhythmTasks(taskDate string) ([]model.MarketRhythmTask, error) {
	return s.repo.AdminEnsureMarketRhythmTasks(taskDate)
}

func (s *growthService) AdminUpdateMarketRhythmTask(id string, owner string, notes string, sourceLinks []string, status string) (model.MarketRhythmTask, error) {
	return s.repo.AdminUpdateMarketRhythmTask(id, owner, notes, sourceLinks, status)
}

func (s *growthService) AdminUpdateMarketRhythmTaskStatus(id string, status string, owner string, notes string) (model.MarketRhythmTask, error) {
	return s.repo.AdminUpdateMarketRhythmTaskStatus(id, status, owner, notes)
}

func (s *growthService) AdminListUsers(status string, kycStatus string, memberLevel string, registrationSource string, page int, pageSize int) ([]model.AdminUser, int, error) {
	return s.repo.AdminListUsers(status, kycStatus, memberLevel, registrationSource, page, pageSize)
}

func (s *growthService) AdminGetUserSourceSummary(status string, kycStatus string, memberLevel string, registrationSource string) (model.AdminUserSourceSummary, error) {
	return s.repo.AdminGetUserSourceSummary(status, kycStatus, memberLevel, registrationSource)
}

func (s *growthService) AdminListBrowseHistories(userID string, contentType string, keyword string, page int, pageSize int) ([]model.AdminBrowseHistory, int, error) {
	return s.repo.AdminListBrowseHistories(userID, contentType, keyword, page, pageSize)
}

func (s *growthService) AdminGetBrowseHistorySummary() (model.AdminBrowseHistorySummary, error) {
	return s.repo.AdminGetBrowseHistorySummary()
}

func (s *growthService) AdminGetBrowseHistoryTrend(days int) ([]model.AdminBrowseTrendPoint, error) {
	return s.repo.AdminGetBrowseHistoryTrend(days)
}

func (s *growthService) AdminListBrowseUserSegments(limit int) ([]model.AdminBrowseUserSegment, error) {
	return s.repo.AdminListBrowseUserSegments(limit)
}

func (s *growthService) AdminListUserMessages(userID string, messageType string, readStatus string, page int, pageSize int) ([]model.AdminUserMessage, int, error) {
	return s.repo.AdminListUserMessages(userID, messageType, readStatus, page, pageSize)
}

func (s *growthService) AdminCreateUserMessages(userIDs []string, title string, content string, messageType string) (int, []model.AdminMessageSendFailure, error) {
	return s.repo.AdminCreateUserMessages(userIDs, title, content, messageType)
}

func (s *growthService) AdminUpdateUserStatus(id string, status string) error {
	return s.repo.AdminUpdateUserStatus(id, status)
}

func (s *growthService) AdminUpdateUserMemberLevel(id string, memberLevel string) error {
	return s.repo.AdminUpdateUserMemberLevel(id, memberLevel)
}

func (s *growthService) AdminUpdateUserKYCStatus(id string, kycStatus string) error {
	return s.repo.AdminUpdateUserKYCStatus(id, kycStatus)
}

func (s *growthService) AdminResetUserPasswordHash(id string, passwordHash string) error {
	return s.repo.AdminResetUserPasswordHash(id, passwordHash)
}

func (s *growthService) AdminDashboardOverview() (model.AdminDashboardOverview, error) {
	return s.repo.AdminDashboardOverview()
}

func (s *growthService) AdminCreateOperationLog(module string, action string, targetType string, targetID string, operatorUserID string, beforeValue string, afterValue string, reason string) error {
	return s.repo.AdminCreateOperationLog(module, action, targetType, targetID, operatorUserID, beforeValue, afterValue, reason)
}

func (s *growthService) AdminListOperationLogs(module string, action string, operatorUserID string, page int, pageSize int) ([]model.AdminOperationLog, int, error) {
	return s.repo.AdminListOperationLogs(module, action, operatorUserID, page, pageSize)
}

func (s *growthService) AdminListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error) {
	return s.repo.AdminListMembershipProducts(status, page, pageSize)
}

func (s *growthService) AdminCreateMembershipProduct(name string, price float64, status string, memberLevel string, durationDays int) (string, error) {
	return s.repo.AdminCreateMembershipProduct(name, price, status, memberLevel, durationDays)
}

func (s *growthService) AdminUpdateMembershipProduct(id string, name string, price float64, status string, memberLevel string, durationDays int) error {
	return s.repo.AdminUpdateMembershipProduct(id, name, price, status, memberLevel, durationDays)
}

func (s *growthService) AdminUpdateMembershipProductStatus(id string, status string) error {
	return s.repo.AdminUpdateMembershipProductStatus(id, status)
}

func (s *growthService) AdminListMembershipOrders(status string, userID string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error) {
	return s.repo.AdminListMembershipOrders(status, userID, page, pageSize)
}

func (s *growthService) AdminUpdateMembershipOrderStatus(id string, status string) error {
	return s.repo.AdminUpdateMembershipOrderStatus(id, status)
}

func (s *growthService) AdminGetExperimentAnalyticsSummary(days int) (model.AdminExperimentAnalyticsSummary, error) {
	return s.repo.AdminGetExperimentAnalyticsSummary(days)
}

func (s *growthService) AdminListVIPQuotaConfigs(memberLevel string, status string, page int, pageSize int) ([]model.VIPQuotaConfig, int, error) {
	return s.repo.AdminListVIPQuotaConfigs(memberLevel, status, page, pageSize)
}

func (s *growthService) AdminCreateVIPQuotaConfig(item model.VIPQuotaConfig) (string, error) {
	return s.repo.AdminCreateVIPQuotaConfig(item)
}

func (s *growthService) AdminUpdateVIPQuotaConfig(id string, item model.VIPQuotaConfig) error {
	return s.repo.AdminUpdateVIPQuotaConfig(id, item)
}

func (s *growthService) AdminListUserQuotaUsages(userID string, periodKey string, page int, pageSize int) ([]model.UserQuotaUsage, int, error) {
	return s.repo.AdminListUserQuotaUsages(userID, periodKey, page, pageSize)
}

func (s *growthService) AdminAdjustUserQuota(userID string, periodKey string, docReadDelta int, newsSubscribeDelta int) error {
	return s.repo.AdminAdjustUserQuota(userID, periodKey, docReadDelta, newsSubscribeDelta)
}

func (s *growthService) AdminListDataSources(page int, pageSize int) ([]model.DataSource, int, error) {
	return s.repo.AdminListDataSources(page, pageSize)
}

func (s *growthService) AdminCreateDataSource(item model.DataSource) (string, error) {
	return s.repo.AdminCreateDataSource(item)
}

func (s *growthService) AdminUpdateDataSource(sourceKey string, item model.DataSource) error {
	return s.repo.AdminUpdateDataSource(sourceKey, item)
}

func (s *growthService) AdminDeleteDataSource(sourceKey string) error {
	return s.repo.AdminDeleteDataSource(sourceKey)
}

func (s *growthService) AdminCheckDataSourceHealth(sourceKey string) (model.DataSourceHealthCheck, error) {
	return s.repo.AdminCheckDataSourceHealth(sourceKey)
}

func (s *growthService) AdminBatchCheckDataSourceHealth(sourceKeys []string) ([]model.DataSourceHealthCheck, error) {
	return s.repo.AdminBatchCheckDataSourceHealth(sourceKeys)
}

func (s *growthService) AdminListDataSourceHealthLogs(sourceKey string, page int, pageSize int) ([]model.DataSourceHealthLog, int, error) {
	return s.repo.AdminListDataSourceHealthLogs(sourceKey, page, pageSize)
}

func (s *growthService) AdminListSystemConfigs(keyword string, page int, pageSize int) ([]model.SystemConfig, int, error) {
	return s.repo.AdminListSystemConfigs(keyword, page, pageSize)
}

func (s *growthService) AdminUpsertSystemConfig(configKey string, configValue string, description string, operator string) error {
	return s.repo.AdminUpsertSystemConfig(configKey, configValue, description, operator)
}

func (s *growthService) AdminListReviewTasks(module string, status string, submitterID string, reviewerID string, page int, pageSize int) ([]model.ReviewTask, int, error) {
	return s.repo.AdminListReviewTasks(module, status, submitterID, reviewerID, page, pageSize)
}

func (s *growthService) AdminSubmitReviewTask(module string, targetID string, submitterID string, reviewerID string, submitNote string) (string, error) {
	return s.repo.AdminSubmitReviewTask(module, targetID, submitterID, reviewerID, submitNote)
}

func (s *growthService) AdminAssignReviewTask(reviewID string, reviewerID string) error {
	return s.repo.AdminAssignReviewTask(reviewID, reviewerID)
}

func (s *growthService) AdminReviewTaskDecision(reviewID string, status string, reviewerID string, reviewNote string) error {
	return s.repo.AdminReviewTaskDecision(reviewID, status, reviewerID, reviewNote)
}

func (s *growthService) GetSchedulerJobNameByRunID(runID string) (string, error) {
	return s.repo.GetSchedulerJobNameByRunID(runID)
}

func (s *growthService) AdminListSchedulerJobRuns(jobName string, status string, page int, pageSize int) ([]model.SchedulerJobRun, int, error) {
	return s.repo.AdminListSchedulerJobRuns(jobName, status, page, pageSize)
}

func (s *growthService) AdminListNewsSyncRunDetails(runID string, syncType string, source string, symbol string, status string, page int, pageSize int) ([]model.NewsSyncRunDetail, int, error) {
	return s.repo.AdminListNewsSyncRunDetails(runID, syncType, source, symbol, status, page, pageSize)
}

func (s *growthService) AdminCreateNewsSyncRunDetails(runID string, details []model.NewsSyncRunDetail) error {
	return s.repo.AdminCreateNewsSyncRunDetails(runID, details)
}

func (s *growthService) AdminCreateSchedulerJobRun(jobName string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	return s.repo.AdminCreateSchedulerJobRun(jobName, triggerSource, status, resultSummary, errorMessage, operatorID)
}

func (s *growthService) AdminRetrySchedulerJobRun(runID string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	return s.repo.AdminRetrySchedulerJobRun(runID, triggerSource, status, resultSummary, errorMessage, operatorID)
}

func (s *growthService) AdminListSchedulerJobDefinitions(status string, module string, page int, pageSize int) ([]model.SchedulerJobDefinition, int, error) {
	return s.repo.AdminListSchedulerJobDefinitions(status, module, page, pageSize)
}

func (s *growthService) AdminCreateSchedulerJobDefinition(item model.SchedulerJobDefinition, operatorID string) (string, error) {
	return s.repo.AdminCreateSchedulerJobDefinition(item, operatorID)
}

func (s *growthService) AdminUpdateSchedulerJobDefinition(id string, item model.SchedulerJobDefinition, operatorID string) error {
	return s.repo.AdminUpdateSchedulerJobDefinition(id, item, operatorID)
}

func (s *growthService) AdminUpdateSchedulerJobDefinitionStatus(id string, status string, operatorID string) error {
	return s.repo.AdminUpdateSchedulerJobDefinitionStatus(id, status, operatorID)
}

func (s *growthService) AdminDeleteSchedulerJobDefinition(id string) error {
	return s.repo.AdminDeleteSchedulerJobDefinition(id)
}

func (s *growthService) AdminListWorkflowMessages(module string, eventType string, isRead string, receiverID string, page int, pageSize int) ([]model.WorkflowMessage, int, error) {
	return s.repo.AdminListWorkflowMessages(module, eventType, isRead, receiverID, page, pageSize)
}

func (s *growthService) AdminCountUnreadWorkflowMessages(module string, eventType string, receiverID string) (int, error) {
	return s.repo.AdminCountUnreadWorkflowMessages(module, eventType, receiverID)
}

func (s *growthService) AdminUpdateWorkflowMessageRead(id string, isRead bool) error {
	return s.repo.AdminUpdateWorkflowMessageRead(id, isRead)
}

func (s *growthService) AdminBulkReadWorkflowMessages(module string, eventType string, receiverID string) (int64, error) {
	return s.repo.AdminBulkReadWorkflowMessages(module, eventType, receiverID)
}

func (s *growthService) AdminCreateWorkflowMessage(reviewID string, targetID string, module string, receiverID string, senderID string, eventType string, title string, content string) error {
	return s.repo.AdminCreateWorkflowMessage(reviewID, targetID, module, receiverID, senderID, eventType, title, content)
}

func (s *growthService) AdminGetWorkflowMetrics(module string, receiverID string) (model.WorkflowMetrics, error) {
	return s.repo.AdminGetWorkflowMetrics(module, receiverID)
}

func (s *growthService) AdminGetSchedulerJobMetrics(jobName string) (model.SchedulerJobMetrics, error) {
	return s.repo.AdminGetSchedulerJobMetrics(jobName)
}
