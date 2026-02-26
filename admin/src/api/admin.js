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
