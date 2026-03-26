import { formatDateTime, formatPercent, formatScore } from "../lib/format.js";
import { buildExplanationView } from "./strategy-version.js";

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

export function mapStockRecommendation(item = {}) {
  return {
    id: item.id || "",
    type: "stock",
    symbol: item.symbol || "-",
    name: item.name || "未命名推荐",
    score: formatScore(item.score),
    risk: item.risk_level || "待确认",
    position: item.position_range || "-",
    validRange: item.valid_from && item.valid_to ? `${item.valid_from} ~ ${item.valid_to}` : item.valid_to || "-",
    reason: item.reason_summary || "等待推荐理由",
    status: item.status || "-",
    sourceType: item.source_type || "-",
    performanceLabel: item.performance_label || ""
  };
}

export function mapFuturesStrategy(item = {}) {
  return {
    id: item.id || "",
    type: "futures",
    symbol: item.contract || "-",
    name: item.name || item.contract || "未命名期货策略",
    score: item.direction || "-",
    risk: item.risk_level || "待确认",
    position: item.position_range || "-",
    validRange: item.valid_from && item.valid_to ? `${item.valid_from} ~ ${item.valid_to}` : item.valid_to || "-",
    reason: item.reason_summary || "等待策略说明",
    status: item.status || "-"
  };
}

export function buildStockInsightView(insight = {}) {
  const recommendation = mapStockRecommendation(insight.recommendation);
  const explanation = buildExplanationView(insight.explanation, recommendation.reason);
  return {
    header: recommendation,
    whyNow: explanation.whyNow,
    proofTags: explanation.proofTags,
    meta: explanation.meta,
    riskBoundary: explanation.riskBoundary || insight.detail?.risk_note || "以止损线和失效条件为准",
    versionSummary: explanation.versionSummary,
    scoreFramework: toArray(insight.score_framework?.factors),
    relatedNews: toArray(insight.related_news),
    performance: {
      winRate: formatPercent(insight.performance_stats?.win_rate),
      cumulativeReturn: formatPercent(insight.performance_stats?.cumulative_return),
      excessReturn: formatPercent(insight.performance_stats?.excess_return),
      maxDrawdown: formatPercent(insight.performance_stats?.max_drawdown)
    },
    guardrails: [
      { label: "止盈建议", value: insight.detail?.take_profit || "-" },
      { label: "止损建议", value: insight.detail?.stop_loss || "-" },
      { label: "风险提示", value: insight.detail?.risk_note || "-" }
    ]
  };
}

export function buildFuturesInsightView(insight = {}) {
  const strategy = mapFuturesStrategy(insight.strategy);
  const explanation = buildExplanationView(insight.explanation, strategy.reason);
  return {
    header: strategy,
    whyNow: explanation.whyNow,
    proofTags: explanation.proofTags,
    meta: explanation.meta,
    riskBoundary: explanation.riskBoundary || insight.guidance?.invalid_condition || "以失效条件和止损区间为准",
    versionSummary: explanation.versionSummary,
    scoreFramework: toArray(insight.score_framework?.factors),
    relatedNews: toArray(insight.related_news),
    relatedEvents: toArray(insight.related_events),
    performance: {
      winRate: formatPercent(insight.performance_stats?.win_rate),
      cumulativeReturn: formatPercent(insight.performance_stats?.cumulative_return),
      excessReturn: formatPercent(insight.performance_stats?.excess_return),
      maxDrawdown: formatPercent(insight.performance_stats?.max_drawdown)
    },
    guardrails: [
      { label: "入场区间", value: insight.guidance?.entry_range || "-" },
      { label: "止盈区间", value: insight.guidance?.take_profit_range || "-" },
      { label: "止损区间", value: insight.guidance?.stop_loss_range || "-" }
    ]
  };
}

export function mapVersionHistory(items = []) {
  return toArray(items).map((item) => ({
    key: item.publish_id || item.publish_version || item.created_at,
    title: item.reason_summary || item.confidence_reason || "历史版本",
    version: item.publish_version ? `V${item.publish_version}` : item.strategy_version || "-",
    date: formatDateTime(item.created_at || item.generated_at || item.trade_date),
    meta: [item.market_regime, item.portfolio_role].filter(Boolean).join(" · "),
    riskBoundary: item.risk_boundary || "-"
  }));
}
