import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "MarketCenterView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("MarketCenterView keeps migration messaging and jumps to dedicated modules", () => {
  const text = readView();
  assert.match(text, /策略中心/);
  assert.match(text, /数据链路已独立/);
  assert.match(text, /股票数据链路已迁移/);
  assert.match(text, /期货数据链路已迁移/);
  assert.match(text, /前往数据源管理/);
  assert.match(text, /前往智能选股/);
  assert.match(text, /前往智能期货/);
  assert.match(text, /data-sources-governance/);
  assert.match(text, /data-sources-sync/);
});

test("MarketCenterView no longer owns data-source governance logic", () => {
  const text = readView();
  assert.doesNotMatch(text, /batchCheckDataSources/);
  assert.doesNotMatch(text, /listDataSources/);
  assert.doesNotMatch(text, /listSystemConfigs/);
  assert.doesNotMatch(text, /fetchStockSourceOptions/);
  assert.doesNotMatch(text, /fetchFuturesSourceOptions/);
  assert.doesNotMatch(text, /fetchMarketNewsSourceOptions/);
  assert.doesNotMatch(text, /supportsSyncKind/);
  assert.doesNotMatch(text, /buildAutoSourceOption/);
  assert.doesNotMatch(text, /buildCustomChainSourceOption/);
  assert.doesNotMatch(text, /handleSyncStockQuotes/);
  assert.doesNotMatch(text, /handleSyncFuturesQuotes/);
  assert.doesNotMatch(text, /handleSyncMarketNews/);
  assert.doesNotMatch(text, /handleRebuildMarketTruth/);
});


test("MarketCenterView consumes audit deep-link query state", () => {
  const text = readView();
  assert.match(text, /applyMarketCenterRouteState/);
  assert.match(text, /normalizeMarketCenterRouteState\(query\)/);
  assert.match(text, /publish_id/);
  assert.match(text, /policy_id/);
  assert.match(text, /config_type/);
  assert.match(text, /config_id/);
  assert.match(text, /view/);
  assert.match(text, /focusStrategyConfigItem/);
  assert.match(text, /handleViewStockPublishDetail/);
  assert.match(text, /handleViewStockPublishReplay/);
  assert.match(text, /handleViewFuturesPublishDetail/);
  assert.match(text, /handleViewFuturesPublishReplay/);
  assert.match(text, /watch\(\s*\(\) => route\.query/);
});
