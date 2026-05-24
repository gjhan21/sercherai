import test from "node:test";
import assert from "node:assert/strict";

import {
  buildSyncJobPayloadFromTemplate,
  buildSyncJobTemplateOptions,
  formatSyncJobTypeLabel
} from "../lib/system-jobs-admin.js";

test("buildSyncJobTemplateOptions exposes four stock and futures sync templates", () => {
  const options = buildSyncJobTemplateOptions();
  assert.deepEqual(
    options.map((item) => item.key),
    ["STOCK_FULL", "STOCK_INCREMENTAL", "FUTURES_FULL", "FUTURES_INCREMENTAL"]
  );
});

test("buildSyncJobPayloadFromTemplate maps stock and futures templates to backfill payloads", () => {
  assert.deepEqual(buildSyncJobPayloadFromTemplate("STOCK_FULL"), {
    run_type: "FULL",
    asset_scope: ["STOCK", "INDEX", "ETF", "LOF", "CBOND"],
    stages: ["UNIVERSE", "MASTER", "QUOTES", "DAILY_BASIC", "MONEYFLOW", "TRUTH", "COVERAGE_SUMMARY"],
    force_refresh_universe: true,
    rebuild_truth_after_sync: true
  });

  assert.deepEqual(buildSyncJobPayloadFromTemplate("STOCK_INCREMENTAL"), {
    run_type: "INCREMENTAL",
    asset_scope: ["STOCK", "INDEX", "ETF", "LOF", "CBOND"],
    stages: ["QUOTES", "DAILY_BASIC", "MONEYFLOW", "TRUTH", "COVERAGE_SUMMARY"],
    force_refresh_universe: false,
    rebuild_truth_after_sync: true
  });

  assert.deepEqual(buildSyncJobPayloadFromTemplate("FUTURES_FULL"), {
    run_type: "FULL",
    asset_scope: ["FUTURES"],
    stages: ["UNIVERSE", "MASTER", "QUOTES", "TRUTH", "COVERAGE_SUMMARY"],
    force_refresh_universe: true,
    rebuild_truth_after_sync: true
  });

  assert.deepEqual(buildSyncJobPayloadFromTemplate("FUTURES_INCREMENTAL"), {
    run_type: "INCREMENTAL",
    asset_scope: ["FUTURES"],
    stages: ["QUOTES", "TRUTH", "COVERAGE_SUMMARY"],
    force_refresh_universe: false,
    rebuild_truth_after_sync: true
  });
});

test("formatSyncJobTypeLabel derives stock and futures sync labels from old backfill runs", () => {
  assert.equal(
    formatSyncJobTypeLabel({ run_type: "FULL", asset_scope: ["STOCK", "INDEX", "ETF", "LOF", "CBOND"] }),
    "股票全量同步"
  );
  assert.equal(
    formatSyncJobTypeLabel({ run_type: "INCREMENTAL", asset_scope: ["STOCK", "ETF"] }),
    "股票每日增量同步"
  );
  assert.equal(formatSyncJobTypeLabel({ run_type: "FULL", asset_scope: ["FUTURES"] }), "期货全量同步");
  assert.equal(
    formatSyncJobTypeLabel({ run_type: "INCREMENTAL", asset_scope: ["FUTURES"] }),
    "期货每日增量同步"
  );
});
