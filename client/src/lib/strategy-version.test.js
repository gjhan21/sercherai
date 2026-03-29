import test from "node:test";
import assert from "node:assert/strict";
import {
  buildStrategyDeepForecastSummary,
  buildStrategyResearchOutlineRows,
  buildStrategyThesisCardRows,
  buildStrategyWatchSignalRows,
  buildStrategyConfidenceCalibrationSummary,
  buildStrategyScenarioSnapshotRows,
  buildStrategyAgentOpinionRows,
  buildStrategyScenarioMetaSummary,
  buildStrategyRelationshipSnapshotSummary,
  mapStrategyVersionHistory
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

test("l2 helpers build scenario and veto summaries", () => {
  const explanation = {
    scenario_snapshots: [
      { scenario: "bull", thesis: "顺势推进", trigger: "放量突破", action_suggestion: "顺势跟踪" },
      { scenario: "base", thesis: "按计划执行", trigger: "逻辑维持", action_suggestion: "按计划执行" }
    ],
    scenario_meta: {
      primary_scenario: "base",
      consensus_action: "继续观察",
      vetoed: true,
      veto_reason: "风险角色提示回撤超限"
    },
    agent_opinions: [
      { role: "FLOW", stance: "SUPPORT", confidence: 0.72, summary: "资金承接稳定" }
    ]
  };

  assert.equal(buildStrategyScenarioSnapshotRows(explanation)[0].scenario, "bull");
  assert.equal(buildStrategyAgentOpinionRows(explanation)[0].role, "FLOW");
  assert.match(buildStrategyScenarioMetaSummary(explanation).summary, /veto/i);
  assert.equal(
    buildStrategyRelationshipSnapshotSummary({
      relationship_snapshot: {
        relationship_count: 3,
        nodes: [{ label: "白酒" }, { label: "高端消费" }]
      }
    }).summary,
    "关系节点 3 个"
  );
});

test("buildStrategyDeepForecastSummary normalizes deep forecast fields", () => {
  const result = buildStrategyDeepForecastSummary({
    deep_forecast_summary: {
      run_id: "l3run_demo_001",
      status: "SUCCEEDED",
      executive_summary: "深推演确认主情景仍然有效。",
      primary_scenario: "bull",
      action_guidance: "沿确认信号执行",
      generated_at: "2026-03-29T12:30:00Z",
      report_available: true
    },
    deep_forecast_report_ref: {
      run_id: "l3run_demo_001",
      report_id: "l3report_demo_001",
      requires_vip: true,
      full_readable: false
    }
  });

  assert.equal(result.runID, "l3run_demo_001");
  assert.equal(result.statusLabel, "已完成");
  assert.match(result.summary, /主情景/);
  assert.equal(result.reportID, "l3report_demo_001");
});

test("mapStrategyVersionHistory carries deep forecast summary", () => {
  const result = mapStrategyVersionHistory([
    {
      publish_id: "publish_001",
      publish_version: 3,
      strategy_version: "stock-v3",
      reason_summary: "版本理由",
      deep_forecast_summary: {
        run_id: "l3run_history_001",
        status: "RUNNING",
        executive_summary: "深推演仍在运行。",
        primary_scenario: "base",
        action_guidance: "等待更多确认",
        generated_at: "2026-03-29T13:00:00Z"
      }
    }
  ]);

  assert.equal(result[0].deepForecast.runID, "l3run_history_001");
  assert.equal(result[0].deepForecast.statusLabel, "推演中");
});
