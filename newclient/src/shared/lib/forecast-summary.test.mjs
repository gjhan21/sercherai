import test from "node:test";
import assert from "node:assert/strict";
import { buildDeepForecastSummary } from "./forecast-summary.js";

test("buildDeepForecastSummary normalizes readable L3 summary fields", () => {
  const result = buildDeepForecastSummary({
    deep_forecast_summary: {
      run_id: "l3run_demo_001",
      status: "SUCCEEDED",
      executive_summary: "主情景继续有效。",
      primary_scenario: "bull",
      action_guidance: "沿确认信号执行",
      generated_at: "2026-05-22T12:30:00Z",
      report_available: true
    },
    deep_forecast_report_ref: {
      run_id: "l3run_demo_001",
      report_id: "l3report_demo_001",
      requires_vip: true,
      full_readable: false
    }
  });

  assert.equal(result.runId, "l3run_demo_001");
  assert.equal(result.status, "SUCCEEDED");
  assert.equal(result.statusLabel, "已完成");
  assert.equal(result.requiresVip, true);
  assert.equal(result.reportAvailable, true);
  assert.equal(result.tone, "success");
});

test("buildDeepForecastSummary keeps running, queued, and failed states readable", () => {
  const running = buildDeepForecastSummary({
    deep_forecast_summary: { run_id: "l3run_running", status: "RUNNING", executive_summary: "推演中" }
  });
  const queued = buildDeepForecastSummary({
    deep_forecast_summary: { run_id: "l3run_queued", status: "QUEUED", executive_summary: "排队中" }
  });
  const failed = buildDeepForecastSummary({
    deep_forecast_summary: { run_id: "l3run_failed", status: "FAILED", executive_summary: "失败" }
  });

  assert.equal(running.statusLabel, "推演中");
  assert.equal(queued.statusLabel, "排队中");
  assert.equal(failed.statusLabel, "已失败");
});

test("buildDeepForecastSummary keeps CTA-safe run id and VIP metadata", () => {
  const result = buildDeepForecastSummary({
    deep_forecast_summary: {
      run_id: "l3run_entry_001",
      status: "QUEUED",
      executive_summary: "排队中"
    },
    deep_forecast_report_ref: {
      run_id: "l3run_entry_001",
      requires_vip: false
    }
  });

  assert.equal(result.runId, "l3run_entry_001");
  assert.equal(result.statusLabel, "排队中");
  assert.equal(result.requiresVip, false);
});

test("buildDeepForecastSummary returns null when no usable summary exists", () => {
  assert.equal(buildDeepForecastSummary({}), null);
  assert.equal(buildDeepForecastSummary(null), null);
});
