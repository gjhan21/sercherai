import test from "node:test";
import assert from "node:assert/strict";
import {
  buildLastSyncResult,
  buildSyncSourceOptions,
  resolveDataSourceProvider,
  resolveDataSourcesSectionRoute
} from "./data-sources-admin.js";

test("resolveDataSourcesSectionRoute sends log and source deep links to registry", () => {
  assert.equal(resolveDataSourcesSectionRoute({ action: "logs" }), "data-sources-registry");
  assert.equal(resolveDataSourcesSectionRoute({ source_key: "tushare" }), "data-sources-registry");
});

test("resolveDataSourcesSectionRoute defaults plain entry to governance", () => {
  assert.equal(resolveDataSourcesSectionRoute({}), "data-sources-governance");
});

test("resolveDataSourceProvider falls back to source_key for built-in providers", () => {
  assert.equal(resolveDataSourceProvider({ source_key: "tushare", config: {} }), "TUSHARE");
  assert.equal(resolveDataSourceProvider({ source_key: "akshare" }), "AKSHARE");
});

test("buildSyncSourceOptions keeps built-in TUSHARE sources even when provider config is missing", () => {
  const options = buildSyncSourceOptions(
    [
      { source_key: "tushare", name: "Tushare", status: "ACTIVE", source_type: "STOCK", config: {} },
      { source_key: "custom", name: "Custom Feed", status: "ACTIVE", source_type: "STOCK", config: {} }
    ],
    { TUSHARE: { status: "HEALTHY" } },
    "stock_quotes"
  );

  assert.deepEqual(options.map((item) => item.value), ["TUSHARE"]);
});

test("buildLastSyncResult keeps explicit zero count instead of falling back to snapshot count", () => {
  const result = buildLastSyncResult({
    count: 0,
    result: {
      snapshot_count: 1
    }
  });
  assert.equal(result.count, 0);
});
