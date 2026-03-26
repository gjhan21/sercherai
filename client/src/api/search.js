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

export function searchGlobal(params) {
  return http.get("/search/global", { params: buildParams(params) });
}

export function searchGlobalPublic(params) {
  return http.get("/public/search/global", { params: buildParams(params) });
}
