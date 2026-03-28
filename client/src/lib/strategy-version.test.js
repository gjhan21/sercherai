import test from "node:test";
import assert from "node:assert/strict";
import {
  buildStrategyResearchOutlineRows,
  buildStrategyThesisCardRows,
  buildStrategyWatchSignalRows,
  buildStrategyConfidenceCalibrationSummary,
  buildStrategyScenarioSnapshotRows,
  buildStrategyAgentOpinionRows,
  buildStrategyScenarioMetaSummary,
  buildStrategyRelationshipSnapshotSummary
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
