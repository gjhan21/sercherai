import test from "node:test";
import assert from "node:assert/strict";

import { buildForecastHistoryViewModel } from "./forecast-history-view-model.js";

test("buildForecastHistoryViewModel builds default latest vs previous comparison for successful runs", () => {
  const model = buildForecastHistoryViewModel({
    targetKey: "AAPL",
    targetLabel: "Apple",
    runs: [
      {
        id: "run-1",
        status: "SUCCESS",
        target_key: "AAPL",
        target_label: "Apple",
        finished_at: "2026-05-20T09:00:00Z",
        created_at: "2026-05-20T08:00:00Z",
        summary: {
          executive_summary: "等待确认。",
          scenario: "WATCH",
          confidence_label: "LOW"
        },
        report: {
          headline_verdict: "等待确认。",
          review_score: 72,
          dimension_evidence: [
            {
              key: "FLOW",
              summary: "资金未形成共振"
            }
          ]
        }
      },
      {
        id: "run-2",
        status: "SUCCESS",
        target_key: "AAPL",
        target_label: "Apple",
        finished_at: "2026-05-22T09:00:00Z",
        created_at: "2026-05-22T08:00:00Z",
        summary: {
          executive_summary: "趋势延续。",
          scenario: "CONSTRUCTIVE",
          confidence_label: "MEDIUM"
        },
        report: {
          headline_verdict: "趋势延续。",
          review_score: 81,
          dimension_evidence: [
            {
              key: "FLOW",
              summary: "资金开始回流"
            },
            {
              key: "EVENT",
              summary: "事件催化增强"
            }
          ]
        }
      },
      {
        id: "run-3",
        status: "FAILED",
        target_key: "AAPL",
        target_label: "Apple",
        finished_at: "2026-05-23T09:00:00Z"
      },
      {
        id: "run-4",
        status: "SUCCESS",
        target_key: "AAPL",
        target_label: "Apple",
        finished_at: "2026-05-24T09:00:00Z",
        created_at: "2026-05-24T08:00:00Z",
        summary: {
          executive_summary: "确认后顺势跟踪。",
          scenario: "BULLISH",
          confidence_label: "HIGH"
        },
        report: {
          headline_verdict: "确认后顺势跟踪。",
          review_score: 88,
          dimension_evidence: [
            {
              key: "FLOW",
              summary: "资金持续回流"
            },
            {
              key: "TECHNICAL",
              summary: "趋势结构完整"
            }
          ]
        }
      }
    ]
  });

  assert.equal(model.hasHistory, true);
  assert.equal(model.successRunCount, 3);
  assert.equal(model.comparePair.latest.id, "run-4");
  assert.equal(model.comparePair.previous.id, "run-2");
  assert.equal(model.compareSummary.title, "最新一次 vs 上一次");
  assert.equal(model.compareSummary.conclusionChange.changed, true);
  assert.match(model.compareSummary.conclusionChange.summary, /确认后顺势跟踪/);
  assert.equal(model.compareSummary.evidenceChange.changed, true);
  assert.match(model.compareSummary.evidenceChange.summary, /TECHNICAL|技术/);
  assert.deepEqual(model.reviewScores.map((item) => item.score), [72, 81, 88]);
  assert.equal(model.timeline.length, 3);
  assert.equal(model.timeline[0].id, "run-4");
});

test("buildForecastHistoryViewModel returns single-run guidance when only one successful run exists", () => {
  const model = buildForecastHistoryViewModel({
    targetKey: "CL",
    runs: [
      {
        id: "run-only",
        status: "SUCCESS",
        target_key: "CL",
        finished_at: "2026-05-24T09:00:00Z",
        summary: {
          executive_summary: "先观察。"
        },
        report: {
          review_score: 64
        }
      }
    ]
  });

  assert.equal(model.hasHistory, false);
  assert.equal(model.successRunCount, 1);
  assert.equal(model.comparePair.latest.id, "run-only");
  assert.equal(model.comparePair.previous, null);
  assert.equal(model.compareSummary.emptyMessage, "当前仅有 1 次成功深推演，暂不能形成历史对比");
  assert.equal(model.reviewScores[0].score, 64);
});

test("buildForecastHistoryViewModel accepts real history items without explicit status field", () => {
  const model = buildForecastHistoryViewModel({
    targetKey: "RB2609",
    runs: [
      {
        run_id: "run-old",
        target_key: "RB2609",
        target_label: "螺纹主力",
        finished_at: "2026-05-21T09:00:00Z",
        headline_verdict: "等待确认。",
        primary_scenario: "基准情景",
        review: {
          review_score: 68
        }
      },
      {
        run_id: "run-new",
        target_key: "RB2609",
        target_label: "螺纹主力",
        finished_at: "2026-05-24T09:00:00Z",
        headline_verdict: "确认后顺势跟踪。",
        primary_scenario: "乐观情景",
        review: {
          review_score: 82
        }
      }
    ]
  });

  assert.equal(model.successRunCount, 2);
  assert.equal(model.hasHistory, true);
  assert.equal(model.comparePair.latest.run_id, "run-new");
  assert.equal(model.comparePair.previous.run_id, "run-old");
});
