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
	ListRewardRecords(userID string, page int, pageSize int) ([]model.RewardRecord, int, error)
	GetUserAccessProfile(userID string) (model.UserAccessProfile, error)
	GetMembershipQuota(userID string) (model.MembershipQuota, error)
	GetAttachmentFileInfo(userID string, attachmentID string) (model.AttachmentFileInfo, error)
	LogAttachmentDownload(userID string, attachmentID string, articleID string) error
	ListNewsCategories(userID string) ([]model.NewsCategory, error)
	ListNewsArticles(userID string, categoryID string, keyword string, page int, pageSize int) ([]model.NewsArticle, int, error)
	GetNewsArticleDetail(userID string, articleID string) (model.NewsArticle, error)
	ListStockRecommendations(userID string, tradeDate string, page int, pageSize int) ([]model.StockRecommendation, int, error)
	GetStockRecommendationDetail(userID string, recoID string) (model.StockRecommendationDetail, error)
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
	GetFuturesGuidance(contract string) (model.FuturesGuidance, error)
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
	AdminCreateNewsAttachment(articleID string, fileName string, fileURL string, fileSize int64, mimeType string) (string, error)
	AdminListNewsAttachments(articleID string) ([]model.NewsAttachment, error)
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

func (s *growthService) ListRewardRecords(userID string, page int, pageSize int) ([]model.RewardRecord, int, error) {
	return s.repo.ListRewardRecords(userID, page, pageSize)
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

func (s *growthService) ListStockRecommendations(userID string, tradeDate string, page int, pageSize int) ([]model.StockRecommendation, int, error) {
	return s.repo.ListStockRecommendations(userID, tradeDate, page, pageSize)
}

func (s *growthService) GetStockRecommendationDetail(userID string, recoID string) (model.StockRecommendationDetail, error) {
	return s.repo.GetStockRecommendationDetail(userID, recoID)
}

func (s *growthService) ListFuturesStrategies(userID string, contract string, status string, page int, pageSize int) ([]model.FuturesStrategy, int, error) {
	return s.repo.ListFuturesStrategies(userID, contract, status, page, pageSize)
}

func (s *growthService) GetFuturesStrategyDetail(userID string, strategyID string) (model.FuturesStrategy, error) {
	return s.repo.GetFuturesStrategyDetail(userID, strategyID)
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

func (s *growthService) GetFuturesGuidance(contract string) (model.FuturesGuidance, error) {
	return s.repo.GetFuturesGuidance(contract)
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

func (s *growthService) AdminCreateNewsArticle(categoryID string, title string, summary string, content string, visibility string, status string, authorID string) (string, error) {
	return s.repo.AdminCreateNewsArticle(categoryID, title, summary, content, visibility, status, authorID)
}

func (s *growthService) AdminUpdateNewsArticle(id string, categoryID string, title string, summary string, content string, visibility string, status string) error {
	return s.repo.AdminUpdateNewsArticle(id, categoryID, title, summary, content, visibility, status)
}

func (s *growthService) AdminCreateNewsAttachment(articleID string, fileName string, fileURL string, fileSize int64, mimeType string) (string, error) {
	return s.repo.AdminCreateNewsAttachment(articleID, fileName, fileURL, fileSize, mimeType)
}

func (s *growthService) AdminListNewsAttachments(articleID string) ([]model.NewsAttachment, error) {
	return s.repo.AdminListNewsAttachments(articleID)
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

func (s *growthService) AdminGenerateDailyStockRecommendations(tradeDate string) (int, error) {
	return s.repo.AdminGenerateDailyStockRecommendations(tradeDate)
}

func (s *growthService) AdminGenerateDailyFuturesStrategies(tradeDate string) (int, error) {
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

func (s *growthService) AdminListUsers(status string, kycStatus string, memberLevel string, page int, pageSize int) ([]model.AdminUser, int, error) {
	return s.repo.AdminListUsers(status, kycStatus, memberLevel, page, pageSize)
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

func (s *growthService) AdminUpdateMembershipProductStatus(id string, status string) error {
	return s.repo.AdminUpdateMembershipProductStatus(id, status)
}

func (s *growthService) AdminListMembershipOrders(status string, userID string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error) {
	return s.repo.AdminListMembershipOrders(status, userID, page, pageSize)
}

func (s *growthService) AdminUpdateMembershipOrderStatus(id string, status string) error {
	return s.repo.AdminUpdateMembershipOrderStatus(id, status)
}

func (s *growthService) AdminListVIPQuotaConfigs(memberLevel string, status string, page int, pageSize int) ([]model.VIPQuotaConfig, int, error) {
	return s.repo.AdminListVIPQuotaConfigs(memberLevel, status, page, pageSize)
}

func (s *growthService) AdminCreateVIPQuotaConfig(item model.VIPQuotaConfig) (string, error) {
	return s.repo.AdminCreateVIPQuotaConfig(item)
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

func (s *growthService) AdminCreateSchedulerJobRun(jobName string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	return s.repo.AdminCreateSchedulerJobRun(jobName, triggerSource, status, resultSummary, errorMessage, operatorID)
}

func (s *growthService) AdminRetrySchedulerJobRun(runID string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	return s.repo.AdminRetrySchedulerJobRun(runID, status, resultSummary, errorMessage, operatorID)
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
