import http from "../lib/http.js";
import { buildParams } from "../lib/request.js";

export function listStockRecommendations(params) {
  return http.get("/stocks/recommendations", { params: buildParams(params) });
}

export function getStockRecommendationDetail(id) {
  return http.get(`/stocks/recommendations/${encodeURIComponent(id)}`);
}

export function getStockRecommendationPerformance(id) {
  return http.get(`/stocks/recommendations/${encodeURIComponent(id)}/performance`);
}

export function getStockRecommendationInsight(id) {
  return http.get(`/stocks/recommendations/${encodeURIComponent(id)}/insight`);
}

export function getStockRecommendationVersionHistory(id) {
  return http.get(`/stocks/recommendations/${encodeURIComponent(id)}/version-history`);
}

export function listFuturesStrategies(params) {
  return http.get("/futures/strategies", { params: buildParams(params) });
}

export function getFuturesStrategyInsight(id) {
  return http.get(`/futures/strategies/${encodeURIComponent(id)}/insight`);
}

export function getFuturesStrategyVersionHistory(id) {
  return http.get(`/futures/strategies/${encodeURIComponent(id)}/version-history`);
}

export function listMarketEvents(params) {
  return http.get("/market/events", { params: buildParams(params) });
}
