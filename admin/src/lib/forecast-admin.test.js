import test from "node:test";
import assert from "node:assert/strict";

import {
  buildForecastAdminConfigPayloads,
  buildForecastPublishSummary,
  parseForecastAdminConfigMap
} from "./forecast-admin.js";

test("parseForecastAdminConfigMap reads switches and thresholds from config map", () => {
  const config = parseForecastAdminConfigMap({
    "growth.forecast_l1.enabled": "false",
    "growth.forecast_l1.explanation_enabled": "true",
    "growth.forecast_l1.memory_feedback_min_samples": "9",
    "growth.forecast_l1.advisory_priority_threshold": "0.58"
  });

  assert.equal(config.enabled, false);
  assert.equal(config.explanationEnabled, true);
  assert.equal(config.memoryFeedbackMinSamples, 9);
  assert.equal(config.advisoryPriorityThreshold, 0.58);
});

test("buildForecastAdminConfigPayloads emits stable system config keys", () => {
  const payloads = buildForecastAdminConfigPayloads({
    enabled: true,
    explanationEnabled: false,
    memoryFeedbackMinSamples: 7,
    advisoryPriorityThreshold: 0.61
  });

  assert.deepEqual(
    payloads.map((item) => item.config_key),
    [
      "growth.forecast_l1.enabled",
      "growth.forecast_l1.explanation_enabled",
      "growth.forecast_l1.memory_feedback_min_samples",
      "growth.forecast_l1.advisory_priority_threshold"
    ]
  );
  assert.equal(payloads[2].config_value, "7");
  assert.equal(payloads[3].config_value, "0.61");
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
            }
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
  assert.equal(summary.coverageRatio, "66.7%");
});
