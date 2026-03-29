import http from "../lib/http";

function buildParams(params = {}) {
  const result = {};
  Object.keys(params).forEach((key) => {
    const value = params[key];
    if (value !== undefined && value !== null && value !== "") {
      result[key] = value;
    }
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
