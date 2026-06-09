import http from "../shared/lib/http";

function buildParams(params = {}) {
  const result = {};
  Object.keys(params).forEach((key) => {
    const value = params[key];
    if (value !== undefined && value !== null && value !== "") result[key] = value;
  });
  return result;
}

export function listStockRecommendations(params) {
  return http.get("/stocks/recommendations", { params: buildParams(params) });
}

export function listStockRecommendationHistory(params) {
  return http.get("/stocks/recommendations/history", { params: buildParams(params) });
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

export function listFuturesArbitrage(params) {
  return http.get("/futures/arbitrage", { params: buildParams(params) });
}

export function getFuturesGuidance(contract) {
  return http.get(`/futures/guidance/${encodeURIComponent(contract)}`);
}

export function listFuturesStrategies(params) {
  return http.get("/futures/strategies", { params: buildParams(params) });
}

export function getFuturesStrategyInsight(id) {
  return http.get(`/futures/strategies/${encodeURIComponent(id)}/insight`);
}

export function getStockKline(params) {
  return http.get("/public/stocks/kline", { params: buildParams(params) });
}

export function getStockPatternMatch(params) {
  return http.get("/public/stocks/pattern-match", { params: buildParams(params) });
}

export function getFuturesStrategyVersionHistory(id) {
  return http.get(`/futures/strategies/${encodeURIComponent(id)}/version-history`);
}

export function listMarketEvents(params) {
  return http.get("/market/events", { params: buildParams(params) });
}

export function getMarketEventDetail(id) {
  return http.get(`/market/events/${encodeURIComponent(id)}`);
}

export function getFuturesArbitrageDetail(id) {
  return http.get(`/futures/arbitrage/${encodeURIComponent(id)}`);
}

export function listFuturesArbitrageOpportunities(params) {
  return http.get("/futures/arbitrage/opportunities", { params: buildParams(params) });
}

export function createFuturesAlert(payload) {
  return http.post("/futures/alerts", payload);
}

export function listFuturesReviews(params) {
  return http.get("/futures/reviews", { params: buildParams(params) });
}

export function getFuturesStrategyDetail(id) {
  return http.get(`/futures/strategies/${encodeURIComponent(id)}`);
}

export function getStockSimulatedOverview() {
  return http.get("/stocks/simulated/overview");
}

export function listStockSimulatedPositions(params) {
  return http.get("/stocks/simulated/positions", { params: buildParams(params) });
}
