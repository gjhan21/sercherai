import test from "node:test";
import assert from "node:assert/strict";
import {
  buildStrategyResearchOutlineRows,
  buildStrategyThesisCardRows,
  buildStrategyWatchSignalRows,
  buildStrategyConfidenceCalibrationSummary
} from "./strategy-version.js";

test("buildStrategyConfidenceCalibrationSummary returns advisory label", () => {
  const result = buildStrategyConfidenceCalibrationSummary({
    confidence_calibration: {
      base_confidence: 0.72,
      adjusted_confidence: 0.64,
      advisory_only: true,
      drivers: [{ label: "回撤偏大", impact: -0.08 }]
    }
  });

  assert.match(result.summary, /建议/);
  assert.match(result.deltaLabel, /-8/);
});

test("research, thesis and watch helpers keep readable rows", () => {
  const explanation = {
    research_outline: [{ slot: "TREND", title: "趋势与结构", summary: "趋势延续", status: "ACTIVE" }],
    active_thesis_cards: [{ key: "trend", title: "当前理由", summary: "资金回流", status: "ACTIVE" }],
    watch_signals: [{ title: "风险触发", trigger: "跌破 5 日线", action: "降低暴露", priority: "HIGH" }]
  };

  assert.equal(buildStrategyResearchOutlineRows(explanation)[0].title, "趋势与结构");
  assert.equal(buildStrategyThesisCardRows(explanation, "active")[0].title, "当前理由");
  assert.equal(buildStrategyWatchSignalRows(explanation)[0].trigger, "跌破 5 日线");
});
