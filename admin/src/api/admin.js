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

export function listOperationLogs(params) {
  return http.get("/admin/audit/operation-logs", { params: buildParams(params) });
}

export function listUsers(params) {
  return http.get("/admin/users", { params: buildParams(params) });
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

export function createSchedulerJobDefinition(payload) {
  return http.post("/admin/system/job-definitions", payload);
}

export function updateSchedulerJobDefinition(id, payload) {
  return http.put(`/admin/system/job-definitions/${encodeURIComponent(id)}`, payload);
}

export function updateSchedulerJobDefinitionStatus(id, status) {
  return http.put(`/admin/system/job-definitions/${encodeURIComponent(id)}/status`, { status });
}

export function listSchedulerJobRuns(params) {
  return http.get("/admin/system/job-runs", { params: buildParams(params) });
}

export function getSchedulerJobMetrics(params) {
  return http.get("/admin/system/job-runs/metrics", { params: buildParams(params) });
}

export function triggerSchedulerJob(payload) {
  return http.post("/admin/system/job-runs/trigger", payload);
}

export function retrySchedulerJobRun(id, payload) {
  return http.post(`/admin/system/job-runs/${encodeURIComponent(id)}/retry`, payload);
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

export function generateDailyStockRecommendations(tradeDate = "") {
  return http.post("/admin/stocks/recommendations/generate-daily", null, {
    params: buildParams({ trade_date: tradeDate })
  });
}

export function listFuturesStrategies(params) {
  return http.get("/admin/futures/strategies", { params: buildParams(params) });
}

export function createFuturesStrategy(payload) {
  return http.post("/admin/futures/strategies", payload);
}

export function updateFuturesStrategyStatus(id, status) {
  return http.put(`/admin/futures/strategies/${encodeURIComponent(id)}/status`, {
    status
  });
}

export function listMembershipProducts(params) {
  return http.get("/admin/membership/products", { params: buildParams(params) });
}

export function createMembershipProduct(payload) {
  return http.post("/admin/membership/products", payload);
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
