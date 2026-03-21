import http from "../lib/http";

function buildParams(raw) {
  const params = {};
  Object.entries(raw || {}).forEach(([key, value]) => {
    if (value === undefined || value === null || value === "") {
      return;
    }
    params[key] = value;
  });
  return params;
}

export function getDashboardOverview() {
  return http.get("/admin/dashboard/overview");
}

export function getAccessProfile() {
  return http.get("/admin/access/me");
}

export function listAccessPermissions(params) {
  return http.get("/admin/access/permissions", { params: buildParams(params) });
}

export function listAccessRoles(params) {
  return http.get("/admin/access/roles", { params: buildParams(params) });
}

export function createAccessRole(payload) {
  return http.post("/admin/access/roles", payload);
}

export function updateAccessRole(id, payload) {
  return http.put(`/admin/access/roles/${encodeURIComponent(id)}`, payload);
}

export function updateAccessRoleStatus(id, status) {
  return http.put(`/admin/access/roles/${encodeURIComponent(id)}/status`, { status });
}

export function listAdminAccounts(params) {
  return http.get("/admin/access/admin-users", { params: buildParams(params) });
}

export function createAdminAccount(payload) {
  return http.post("/admin/access/admin-users", payload);
}

export function updateAdminAccountStatus(id, status) {
  return http.put(`/admin/access/admin-users/${encodeURIComponent(id)}/status`, { status });
}

export function assignAdminAccountRoles(id, roleIDs) {
  return http.put(`/admin/access/admin-users/${encodeURIComponent(id)}/roles`, {
    role_ids: roleIDs
  });
}

export function resetAdminAccountPassword(id, password) {
  return http.put(`/admin/access/admin-users/${encodeURIComponent(id)}/password`, {
    password
  });
}

export function listOperationLogs(params) {
  return http.get("/admin/audit/operation-logs", { params: buildParams(params) });
}

export function listUsers(params) {
  return http.get("/admin/users", { params: buildParams(params) });
}

export function getUserSourceSummary(params) {
  return http.get("/admin/users/source-summary", { params: buildParams(params) });
}

export function getUserCenterOverview(id, params) {
  return http.get(`/admin/users/${encodeURIComponent(id)}/center-overview`, {
    params: buildParams(params)
  });
}

export function updateUserSubscription(userID, subscriptionID, payload) {
  return http.put(
    `/admin/users/${encodeURIComponent(userID)}/subscriptions/${encodeURIComponent(subscriptionID)}`,
    payload
  );
}

export function updateUserStatus(id, status) {
  return http.put(`/admin/users/${encodeURIComponent(id)}/status`, { status });
}

export function updateUserMemberLevel(id, memberLevel) {
  return http.put(`/admin/users/${encodeURIComponent(id)}/member-level`, {
    member_level: memberLevel
  });
}

export function updateUserKYCStatus(id, kycStatus) {
  return http.put(`/admin/users/${encodeURIComponent(id)}/kyc-status`, {
    kyc_status: kycStatus
  });
}

export function listBrowseHistories(params) {
  return http.get("/admin/users/browse-histories", { params: buildParams(params) });
}

export function getBrowseHistorySummary() {
  return http.get("/admin/users/browse-histories/summary");
}

export function getBrowseHistoryTrend(params) {
  return http.get("/admin/users/browse-histories/trend", { params: buildParams(params) });
}

export function listBrowseUserSegments(params) {
  return http.get("/admin/users/browse-histories/segments", { params: buildParams(params) });
}

export function listUserMessages(params) {
  return http.get("/admin/users/messages", { params: buildParams(params) });
}

export function createUserMessages(payload) {
  return http.post("/admin/users/messages", payload);
}

export function listNewsCategories(params) {
  return http.get("/admin/news/categories", { params: buildParams(params) });
}

export function createNewsCategory(payload) {
  return http.post("/admin/news/categories", payload);
}

export function updateNewsCategory(id, payload) {
  return http.put(`/admin/news/categories/${encodeURIComponent(id)}`, payload);
}

export function listNewsArticles(params) {
  return http.get("/admin/news/articles", { params: buildParams(params) });
}

export function getNewsArticleDetail(id) {
  return http.get(`/admin/news/articles/${encodeURIComponent(id)}`);
}

export function createNewsArticle(payload) {
  return http.post("/admin/news/articles", payload);
}

export function updateNewsArticle(id, payload) {
  return http.put(`/admin/news/articles/${encodeURIComponent(id)}`, payload);
}

export function publishNewsArticle(id) {
  return http.put(`/admin/news/articles/${encodeURIComponent(id)}/publish`, {
    status: "PUBLISHED"
  });
}

export function listNewsAttachments(articleID) {
  return http.get(`/admin/news/articles/${encodeURIComponent(articleID)}/attachments`);
}

export function createNewsAttachment(articleID, payload) {
  return http.post(`/admin/news/articles/${encodeURIComponent(articleID)}/attachments`, payload);
}

export function uploadNewsAttachmentFile(formData) {
  return http.post("/admin/news/attachments/upload", formData);
}

export function deleteNewsAttachment(id) {
  return http.delete(`/admin/news/attachments/${encodeURIComponent(id)}`);
}

export function listDataSources(params) {
  return http.get("/admin/data-sources", { params: buildParams(params) });
}

export function createDataSource(payload) {
  return http.post("/admin/data-sources", payload);
}

export function updateDataSource(sourceKey, payload) {
  return http.put(`/admin/data-sources/${encodeURIComponent(sourceKey)}`, payload);
}

export function deleteDataSource(sourceKey) {
  return http.delete(`/admin/data-sources/${encodeURIComponent(sourceKey)}`);
}

export function checkDataSourceHealth(sourceKey) {
  return http.post(`/admin/data-sources/${encodeURIComponent(sourceKey)}/health-check`);
}

export function batchCheckDataSources(sourceKeys = []) {
  return http.post("/admin/data-sources/health-checks", { source_keys: sourceKeys });
}

export function listDataSourceHealthLogs(sourceKey, params) {
  return http.get(`/admin/data-sources/${encodeURIComponent(sourceKey)}/health-logs`, {
    params: buildParams(params)
  });
}

export function listWorkflowMessages(params) {
  return http.get("/admin/workflow/messages", { params: buildParams(params) });
}

export function listReviewTasks(params) {
  return http.get("/admin/workflow/reviews", { params: buildParams(params) });
}

export function getWorkflowMetrics(params) {
  return http.get("/admin/workflow/metrics", { params: buildParams(params) });
}

export function submitReviewTask(payload) {
  return http.post("/admin/workflow/reviews/submit", payload);
}

export function assignReviewTask(id, reviewerID) {
  return http.put(`/admin/workflow/reviews/${encodeURIComponent(id)}/assign`, {
    reviewer_id: reviewerID
  });
}

export function reviewTaskDecision(id, status, reviewNote) {
  return http.put(`/admin/workflow/reviews/${encodeURIComponent(id)}/decision`, {
    status,
    review_note: reviewNote
  });
}

export function listSchedulerJobDefinitions(params) {
  return http.get("/admin/system/job-definitions", { params: buildParams(params) });
}

export function listSupportedSchedulerJobs(params) {
  return http.get("/admin/system/job-definitions/supported", { params: buildParams(params) });
}

export function createSchedulerJobDefinition(payload) {
  return http.post("/admin/system/job-definitions", payload);
}

export function updateSchedulerJobDefinition(id, payload) {
  return http.put(`/admin/system/job-definitions/${encodeURIComponent(id)}`, payload);
}

export function updateSchedulerJobDefinitionStatus(id, status) {
  return http.put(`/admin/system/job-definitions/${encodeURIComponent(id)}/status`, { status });
}

export function deleteSchedulerJobDefinition(id) {
  return http.delete(`/admin/system/job-definitions/${encodeURIComponent(id)}`);
}

export function listSchedulerJobRuns(params) {
  return http.get("/admin/system/job-runs", { params: buildParams(params) });
}

export function listNewsSyncRunDetails(runID, params) {
  return http.get(`/admin/system/job-runs/${encodeURIComponent(runID)}/news-sync-details`, {
    params: buildParams(params)
  });
}

export function getSchedulerJobMetrics(params) {
  return http.get("/admin/system/job-runs/metrics", { params: buildParams(params) });
}

export function triggerSchedulerJob(payload) {
  return http.post("/admin/system/job-runs/trigger", payload);
}

export function getStockSelectionOverview() {
  return http.get("/admin/stock-selection/overview");
}

export function listStockSelectionRuns(params) {
  return http.get("/admin/stock-selection/runs", { params: buildParams(params) });
}

export function createStockSelectionRun(payload) {
  return http.post("/admin/stock-selection/runs", payload, { timeout: 60000 });
}

export function getStockSelectionRun(runID) {
  return http.get(`/admin/stock-selection/runs/${encodeURIComponent(runID)}`);
}

export function compareStockSelectionRuns(runIDs = []) {
  return http.get("/admin/stock-selection/runs/compare", {
    params: buildParams({ run_ids: Array.isArray(runIDs) ? runIDs.join(",") : runIDs })
  });
}

export function listStockSelectionProfiles(params) {
  return http.get("/admin/stock-selection/profiles", { params: buildParams(params) });
}

export function listStockSelectionProfileVersions(id) {
  return http.get(`/admin/stock-selection/profiles/${encodeURIComponent(id)}/versions`);
}

export function createStockSelectionProfile(payload) {
  return http.post("/admin/stock-selection/profiles", payload);
}

export function updateStockSelectionProfile(id, payload) {
  return http.put(`/admin/stock-selection/profiles/${encodeURIComponent(id)}`, payload);
}

export function publishStockSelectionProfile(id) {
  return http.post(`/admin/stock-selection/profiles/${encodeURIComponent(id)}/publish`);
}

export function rollbackStockSelectionProfile(id, payload) {
  return http.post(`/admin/stock-selection/profiles/${encodeURIComponent(id)}/rollback`, payload);
}

export function listStockSelectionTemplates(params) {
  return http.get("/admin/stock-selection/templates", { params: buildParams(params) });
}

export function createStockSelectionTemplate(payload) {
  return http.post("/admin/stock-selection/templates", payload);
}

export function updateStockSelectionTemplate(id, payload) {
  return http.put(`/admin/stock-selection/templates/${encodeURIComponent(id)}`, payload);
}

export function setDefaultStockSelectionTemplate(id) {
  return http.post(`/admin/stock-selection/templates/${encodeURIComponent(id)}/set-default`);
}

export function listStockSelectionCandidates(runID, params) {
  return http.get(`/admin/stock-selection/runs/${encodeURIComponent(runID)}/candidates`, {
    params: buildParams(params)
  });
}

export function listStockSelectionPortfolio(runID) {
  return http.get(`/admin/stock-selection/runs/${encodeURIComponent(runID)}/portfolio`);
}

export function listStockSelectionRunEvidence(runID, params) {
  return http.get(`/admin/stock-selection/runs/${encodeURIComponent(runID)}/evidence`, {
    params: buildParams(params)
  });
}

export function listStockSelectionRunEvaluation(runID, params) {
  return http.get(`/admin/stock-selection/runs/${encodeURIComponent(runID)}/evaluation`, {
    params: buildParams(params)
  });
}

export function listStockSelectionEvaluationLeaderboard(params) {
  return http.get("/admin/stock-selection/evaluation/leaderboard", {
    params: buildParams(params)
  });
}

export function listStockSelectionReviews(params) {
  return http.get("/admin/stock-selection/reviews", { params: buildParams(params) });
}

export function approveStockSelectionReview(runID, payload) {
  return http.post(`/admin/stock-selection/reviews/${encodeURIComponent(runID)}/approve`, payload, {
    timeout: 60000
  });
}

export function rejectStockSelectionReview(runID, payload) {
  return http.post(`/admin/stock-selection/reviews/${encodeURIComponent(runID)}/reject`, payload);
}

export function retrySchedulerJobRun(id, payload) {
  return http.post(`/admin/system/job-runs/${encodeURIComponent(id)}/retry`, payload);
}

export function retryNewsSyncItem(id, payload) {
  return http.post(`/admin/system/job-runs/${encodeURIComponent(id)}/retry-news-sync-item`, payload);
}

export function countUnreadWorkflowMessages(params) {
  return http.get("/admin/workflow/messages/unread-count", {
    params: buildParams(params)
  });
}

export function updateWorkflowMessageRead(id, isRead) {
  return http.put(`/admin/workflow/messages/${encodeURIComponent(id)}/read`, {
    is_read: isRead
  });
}

export function bulkReadWorkflowMessages(payload) {
  return http.put("/admin/workflow/messages/read-all", payload);
}

export function listInviteRecords(params) {
  return http.get("/admin/growth/invite-records", { params: buildParams(params) });
}

export function listRewardRecords(params) {
  return http.get("/admin/growth/reward-records", { params: buildParams(params) });
}

export function reviewRewardRecord(id, status, reason = "") {
  return http.put(`/admin/growth/reward-records/${encodeURIComponent(id)}/review`, {
    status,
    reason
  });
}

export function listReconciliation(params) {
  return http.get("/admin/payment/reconciliation", { params: buildParams(params) });
}

export function retryReconciliation(batchID) {
  return http.post(`/admin/payment/reconciliation/${encodeURIComponent(batchID)}/retry`);
}

export function listRiskRules() {
  return http.get("/admin/risk/rules");
}

export function createRiskRule(payload) {
  return http.post("/admin/risk/rules", payload);
}

export function updateRiskRule(id, payload) {
  return http.put(`/admin/risk/rules/${encodeURIComponent(id)}`, payload);
}

export function listRiskHits(params) {
  return http.get("/admin/risk/hits", { params: buildParams(params) });
}

export function reviewRiskHit(id, status, reason = "") {
  return http.put(`/admin/risk/hits/${encodeURIComponent(id)}/review`, {
    status,
    reason
  });
}

export function listWithdrawRequests(params) {
  return http.get("/admin/reward-wallet/withdraw-requests", {
    params: buildParams(params)
  });
}

export function reviewWithdrawRequest(id, status, reason = "") {
  return http.put(`/admin/reward-wallet/withdraw-requests/${encodeURIComponent(id)}/review`, {
    status,
    reason
  });
}

export function listStockRecommendations(params) {
  return http.get("/admin/stocks/recommendations", { params: buildParams(params) });
}

export function createStockRecommendation(payload) {
  return http.post("/admin/stocks/recommendations", payload);
}

export function updateStockRecommendationStatus(id, status) {
  return http.put(`/admin/stocks/recommendations/${encodeURIComponent(id)}/status`, {
    status
  });
}

export function syncStockQuotes(payload) {
  return http.post("/admin/stocks/quotes/sync", payload);
}

export function syncFuturesQuotes(payload) {
  return http.post("/admin/futures/quotes/sync", payload);
}

export function syncFuturesInventory(payload) {
  return http.post("/admin/futures/inventory/sync", payload);
}

export function syncMarketNewsSource(payload) {
  return http.post("/admin/news/market-sync", payload);
}

export function listQuantTopStocks(params) {
  return http.get("/admin/stocks/quant/top", { params: buildParams(params) });
}

export function listQuantEvaluation(params) {
  return http.get("/admin/stocks/quant/evaluation", { params: buildParams(params) });
}

export function generateDailyStockRecommendations(tradeDate = "") {
  return http.post("/admin/stocks/recommendations/generate-daily", null, {
    params: buildParams({ trade_date: tradeDate })
  });
}

export function listStockStrategyEnginePublishHistory() {
  return http.get("/admin/stocks/strategy-engine/publish-history");
}

export function getStockStrategyEnginePublishRecord(publishID) {
  return http.get(`/admin/stocks/strategy-engine/publish-records/${encodeURIComponent(publishID)}`);
}

export function getStockStrategyEnginePublishReplay(publishID) {
  return http.get(`/admin/stocks/strategy-engine/publish-records/${encodeURIComponent(publishID)}/replay`);
}

export function compareStockStrategyEnginePublishVersions(payload) {
  return http.post("/admin/stocks/strategy-engine/publish-compare", payload);
}

export function listFuturesStrategies(params) {
  return http.get("/admin/futures/strategies", { params: buildParams(params) });
}

export function generateDailyFuturesStrategies(tradeDate = "") {
  return http.post("/admin/futures/strategies/generate-daily", null, {
    params: buildParams({ trade_date: tradeDate })
  });
}

export function listFuturesStrategyEnginePublishHistory() {
  return http.get("/admin/futures/strategy-engine/publish-history");
}

export function getFuturesStrategyEnginePublishRecord(publishID) {
  return http.get(`/admin/futures/strategy-engine/publish-records/${encodeURIComponent(publishID)}`);
}

export function getFuturesStrategyEnginePublishReplay(publishID) {
  return http.get(`/admin/futures/strategy-engine/publish-records/${encodeURIComponent(publishID)}/replay`);
}

export function compareFuturesStrategyEnginePublishVersions(payload) {
  return http.post("/admin/futures/strategy-engine/publish-compare", payload);
}

export function createFuturesStrategy(payload) {
  return http.post("/admin/futures/strategies", payload);
}

export function updateFuturesStrategyStatus(id, status) {
  return http.put(`/admin/futures/strategies/${encodeURIComponent(id)}/status`, {
    status
  });
}

export function listMarketEvents(params) {
  return http.get("/admin/market/events", { params: buildParams(params) });
}

export function listMarketRhythmTasks(params) {
  return http.get("/admin/market/rhythm-tasks", { params: buildParams(params) });
}

export function ensureMarketRhythmTasks(payload) {
  return http.post("/admin/market/rhythm-tasks/ensure", payload);
}

export function updateMarketRhythmTask(id, payload) {
  return http.put(`/admin/market/rhythm-tasks/${encodeURIComponent(id)}`, payload);
}

export function updateMarketRhythmTaskStatus(id, payload) {
  return http.put(`/admin/market/rhythm-tasks/${encodeURIComponent(id)}/status`, payload);
}

export function getExperimentAnalyticsSummary(params) {
  return http.get("/admin/market/experiments/summary", { params: buildParams(params) });
}

export function listStrategySeedSets(params) {
  return http.get("/admin/market/strategy-engine/seed-sets", { params: buildParams(params) });
}

export function createStrategySeedSet(payload) {
  return http.post("/admin/market/strategy-engine/seed-sets", payload);
}

export function updateStrategySeedSet(id, payload) {
  return http.put(`/admin/market/strategy-engine/seed-sets/${encodeURIComponent(id)}`, payload);
}

export function listStrategyAgentProfiles(params) {
  return http.get("/admin/market/strategy-engine/agents", { params: buildParams(params) });
}

export function createStrategyAgentProfile(payload) {
  return http.post("/admin/market/strategy-engine/agents", payload);
}

export function updateStrategyAgentProfile(id, payload) {
  return http.put(`/admin/market/strategy-engine/agents/${encodeURIComponent(id)}`, payload);
}

export function listStrategyScenarioTemplates(params) {
  return http.get("/admin/market/strategy-engine/scenarios", { params: buildParams(params) });
}

export function createStrategyScenarioTemplate(payload) {
  return http.post("/admin/market/strategy-engine/scenarios", payload);
}

export function updateStrategyScenarioTemplate(id, payload) {
  return http.put(`/admin/market/strategy-engine/scenarios/${encodeURIComponent(id)}`, payload);
}

export function listStrategyPublishPolicies(params) {
  return http.get("/admin/market/strategy-engine/publish-policies", { params: buildParams(params) });
}

export function createStrategyPublishPolicy(payload) {
  return http.post("/admin/market/strategy-engine/publish-policies", payload);
}

export function updateStrategyPublishPolicy(id, payload) {
  return http.put(`/admin/market/strategy-engine/publish-policies/${encodeURIComponent(id)}`, payload);
}

export function listStrategyEngineJobs(params) {
  return http.get("/admin/market/strategy-engine/jobs", { params: buildParams(params) });
}

export function getStrategyEngineJob(jobID) {
  return http.get(`/admin/market/strategy-engine/jobs/${encodeURIComponent(jobID)}`);
}

export function publishStrategyEngineJob(jobID, payload) {
  return http.post(`/admin/market/strategy-engine/jobs/${encodeURIComponent(jobID)}/publish`, payload);
}

export function createMarketEvent(payload) {
  return http.post("/admin/market/events", payload);
}

export function updateMarketEvent(id, payload) {
  return http.put(`/admin/market/events/${encodeURIComponent(id)}`, payload);
}

export function listMembershipProducts(params) {
  return http.get("/admin/membership/products", { params: buildParams(params) });
}

export function createMembershipProduct(payload) {
  return http.post("/admin/membership/products", payload);
}

export function updateMembershipProduct(id, payload) {
  return http.put(`/admin/membership/products/${encodeURIComponent(id)}`, payload);
}

export function updateMembershipProductStatus(id, status) {
  return http.put(`/admin/membership/products/${encodeURIComponent(id)}/status`, {
    status
  });
}

export function listMembershipOrders(params) {
  return http.get("/admin/membership/orders", { params: buildParams(params) });
}

export function updateMembershipOrderStatus(id, status) {
  return http.put(`/admin/membership/orders/${encodeURIComponent(id)}/status`, {
    status
  });
}

export function listVIPQuotaConfigs(params) {
  return http.get("/admin/membership/quota-configs", { params: buildParams(params) });
}

export function createVIPQuotaConfig(payload) {
  return http.post("/admin/membership/quota-configs", payload);
}

export function updateVIPQuotaConfig(id, payload) {
  return http.put(`/admin/membership/quota-configs/${encodeURIComponent(id)}`, payload);
}

export function listUserQuotas(params) {
  return http.get("/admin/membership/user-quotas", { params: buildParams(params) });
}

export function adjustUserQuota(userID, payload) {
  return http.put(`/admin/membership/user-quotas/${encodeURIComponent(userID)}/adjust`, payload);
}

export function listSystemConfigs(params) {
  return http.get("/admin/system/configs", { params: buildParams(params) });
}

export function upsertSystemConfig(payload) {
  return http.put("/admin/system/configs", payload);
}

export function testOSSQiniuConfig() {
  return http.post("/admin/system/configs/oss/qiniu/test");
}

export function testPaymentYolkPayConfig() {
  return http.post("/admin/system/configs/payment/yolkpay/test");
}

export function listAuthLoginLogs(params) {
  return http.get("/admin/auth/login-logs", { params: buildParams(params) });
}

export function getAuthRiskConfig() {
  return http.get("/admin/auth/risk-config");
}

export function updateAuthRiskConfig(payload) {
  return http.put("/admin/auth/risk-config", payload);
}

export function listAuthRiskConfigLogs(params) {
  return http.get("/admin/auth/risk-config-logs", { params: buildParams(params) });
}

export function unlockAuthRiskState(payload) {
  return http.post("/admin/auth/unlock", payload);
}

export function listAuthUnlockLogs(params) {
  return http.get("/admin/auth/unlock-logs", { params: buildParams(params) });
}
