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
	AdminSyncMarketMasterDetailed(assetType string, sourceKey string, instrumentKeys []string) (model.MarketSyncResult, error)
	AdminSyncMarketQuotesDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error)
	AdminSyncMarketDailyBasicDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error)
	AdminSyncMarketMoneyflowDetailed(assetType string, sourceKey string, instrumentKeys []string, days int) (model.MarketSyncResult, error)
	AdminRebuildMarketDailyTruthDetailed(assetType string, sourceKey string, instrumentKeys []string, tradeDateFrom string, tradeDateTo string) (model.MarketSyncResult, error)
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
	AdminListMarketProviderRegistries(status string) ([]model.MarketProviderRegistry, error)
	AdminListMarketProviderCapabilities(providerKey string, assetClass string, dataKind string) ([]model.MarketProviderCapability, error)
	AdminListMarketProviderRoutingPolicies(assetClass string, dataKind string) ([]model.MarketProviderRoutingPolicy, error)
	AdminListMarketProviderQualityScores(assetClass string, dataKind string, hours int) ([]model.MarketProviderQualityScore, error)
	AdminGetMarketProviderGovernanceOverview(assetClass string, dataKind string, hours int) (model.MarketProviderGovernanceOverview, error)
	AdminBuildMarketUniverseSnapshot(sourceKey string, assetScope []string, operator string) (model.MarketUniverseSnapshot, []model.MarketUniverseSnapshotItem, error)
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
