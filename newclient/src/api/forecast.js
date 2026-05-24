import http from "../shared/lib/http";

function buildParams(params = {}) {
  const result = {};
  Object.keys(params).forEach((key) => {
    const value = params[key];
    if (value !== undefined && value !== null && value !== "") result[key] = value;
  });
  return result;
}

export function createForecastRun(payload) {
  return http.post("/forecast/runs", payload);
}

export function listForecastRuns(params) {
  return http.get("/forecast/runs", { params: buildParams(params) });
}

export function getForecastRunDetail(id) {
  return http.get(`/forecast/runs/${encodeURIComponent(id)}`);
}

export function getForecastRunHistory(params) {
  return http.get("/forecast/targets/history", { params: buildParams(params) });
}

export function getForecastRunCompare(params) {
  return http.get("/forecast/targets/history/compare", { params: buildParams(params) });
}

export function getForecastRunReview(id) {
  return http.get(`/forecast/runs/${encodeURIComponent(id)}/review`);
}
