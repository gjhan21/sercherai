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
