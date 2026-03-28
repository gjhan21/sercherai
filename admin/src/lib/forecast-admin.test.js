import test from "node:test";
import assert from "node:assert/strict";

import {
  buildForecastAdminConfigPayloads,
  buildForecastL2Summary,
  buildForecastPublishSummary,
  parseForecastAdminConfigMap
} from "./forecast-admin.js";

test("parseForecastAdminConfigMap reads switches and thresholds from config map", () => {
  const config = parseForecastAdminConfigMap({
    "growth.forecast_l1.enabled": "false",
    "growth.forecast_l1.explanation_enabled": "true",
    "growth.forecast_l1.memory_feedback_min_samples": "9",
    "growth.forecast_l1.advisory_priority_threshold": "0.58",
    "growth.forecast_l2.enabled": "true",
    "growth.forecast_l2.relationship_snapshot_enabled": "true",
    "growth.forecast_l2.stable_scenarios_enabled": "false",
    "growth.forecast_l2.veto_confidence_threshold": "0.31"
  });

  assert.equal(config.enabled, false);
  assert.equal(config.explanationEnabled, true);
  assert.equal(config.memoryFeedbackMinSamples, 9);
  assert.equal(config.advisoryPriorityThreshold, 0.58);
  assert.equal(config.l2Enabled, true);
  assert.equal(config.relationshipSnapshotEnabled, true);
  assert.equal(config.stableScenariosEnabled, false);
  assert.equal(config.vetoConfidenceThreshold, 0.31);
});

test("buildForecastAdminConfigPayloads emits stable system config keys", () => {
  const payloads = buildForecastAdminConfigPayloads({
    enabled: true,
    explanationEnabled: false,
    memoryFeedbackMinSamples: 7,
    advisoryPriorityThreshold: 0.61,
    l2Enabled: true,
    relationshipSnapshotEnabled: true,
    stableScenariosEnabled: true,
    vetoConfidenceThreshold: 0.28
  });

  assert.deepEqual(
    payloads.map((item) => item.config_key),
    [
      "growth.forecast_l1.enabled",
      "growth.forecast_l1.explanation_enabled",
      "growth.forecast_l1.memory_feedback_min_samples",
      "growth.forecast_l1.advisory_priority_threshold",
      "growth.forecast_l2.enabled",
      "growth.forecast_l2.relationship_snapshot_enabled",
      "growth.forecast_l2.stable_scenarios_enabled",
      "growth.forecast_l2.veto_confidence_threshold"
    ]
  );
  assert.equal(payloads[2].config_value, "7");
  assert.equal(payloads[3].config_value, "0.61");
  assert.equal(payloads[7].config_value, "0.28");
});

test("buildForecastPublishSummary summarizes enhanced explanation coverage", () => {
  const summary = buildForecastPublishSummary(
    {
      payload_count: 3,
      publish_payloads: [
        {
          explanation: {
            research_outline: [{ slot: "TREND" }],
            active_thesis_cards: [{ key: "trend" }],
            watch_signals: [{ title: "跌破 5 日线" }],
            memory_feedback: { summary: "题材切换快" },
            confidence_calibration: {
              adjusted_confidence: 0.42,
              advisory_only: true
            },
            relationship_snapshot: {
              relationship_count: 3,
              nodes: [{ label: "白酒" }, { label: "高端消费" }, { label: "贵州茅台" }]
            },
            scenario_snapshots: [{ scenario: "bull" }, { scenario: "base" }, { scenario: "bear" }],
            scenario_meta: {
              primary_scenario: "base",
              vetoed: true
            },
            agent_opinions: [{ role: "FLOW" }, { role: "RISK" }]
          }
        },
        {
          detail: {
            confidence_calibration: {
              adjusted_confidence: 0.77,
              advisory_only: true
            }
          }
        },
        {
          report_snapshot: {
            research_outline: [{ slot: "FLOW" }],
            watch_signals: [{ title: "缩量回踩" }]
          }
        }
      ],
      report_snapshot: {}
    },
    0.55
  );

  assert.equal(summary.enhancedCount, 2);
  assert.equal(summary.researchOutlineCount, 2);
  assert.equal(summary.watchSignalCount, 2);
  assert.equal(summary.memoryFeedbackCount, 1);
  assert.equal(summary.highAdvisoryCount, 1);
  assert.equal(summary.scenarioSnapshotCount, 3);
  assert.equal(summary.relationshipNodeCount, 3);
  assert.equal(summary.vetoedCount, 1);
  assert.equal(summary.agentOpinionCount, 2);
  assert.deepEqual(summary.primaryScenarios, ["base"]);
  assert.equal(summary.coverageRatio, "66.7%");
});

test("buildForecastL2Summary extracts primary scenario and veto summary", () => {
  const summary = buildForecastL2Summary({
    publish_payloads: [
      {
        explanation: {
          relationship_snapshot: {
            relationship_count: 4,
            nodes: [{ label: "铜" }, { label: "库存" }]
          },
          scenario_snapshots: [{ scenario: "bull" }, { scenario: "base" }, { scenario: "bear" }],
          scenario_meta: {
            primary_scenario: "base",
            consensus_action: "继续观察",
            vetoed: true,
            veto_reason: "风险角色提示波动超限"
          },
          agent_opinions: [{ role: "SUPPLY" }, { role: "RISK" }]
        }
      }
    ]
  });

  assert.equal(summary.primaryScenario, "base");
  assert.equal(summary.consensusAction, "继续观察");
  assert.equal(summary.vetoed, true);
  assert.equal(summary.vetoReason, "风险角色提示波动超限");
  assert.equal(summary.relationshipCount, 2);
  assert.equal(summary.scenarioCount, 3);
  assert.deepEqual(summary.topRoles, ["SUPPLY", "RISK"]);
});
