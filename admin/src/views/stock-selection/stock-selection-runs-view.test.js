import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "StockSelectionRunsView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("StockSelectionRunsView keeps run detail context section", () => {
  const text = readView();
  assert.match(text, /上下文与过滤条件/);
  assert.match(text, /selected_trade_date/);
});

test("StockSelectionRunsView surfaces governance routing summary", () => {
  const text = readView();
  assert.match(text, /detailGovernanceRows/);
  assert.match(text, /治理路由摘要/);
  assert.match(text, /路由主源/);
  assert.match(text, /回退链路/);
  assert.match(text, /决策原因/);
  assert.match(text, /策略键/);
  assert.match(text, /selected_source/);
  assert.match(text, /fallback_chain/);
  assert.match(text, /decision_reason/);
  assert.match(text, /policy_key/);
});

test("StockSelectionRunsView surfaces evaluation backfill state from run context", () => {
  const text = readView();
  assert.match(text, /evaluation_backfill_state/);
  assert.match(text, /评估回填状态/);
  assert.match(text, /formatStockSelectionEvaluationStatus/);
});

test("StockSelectionRunsView runs with overview trade date and default profile", () => {
  const text = readView();
  assert.match(text, /getStockSelectionOverview/);
  assert.match(text, /overview = ref/);
  assert.match(text, /latest_trade_date/);
  assert.match(text, /default_profile/);
  assert.doesNotMatch(text, /createStockSelectionRun\(\{\}\)/);
  assert.match(text, /trade_date:\s*tradeDate/);
  assert.match(text, /profile_id:\s*profileID/);
});

test("StockSelectionRunsView guides succeeded pending runs to review publish", () => {
  const text = readView();
  assert.match(text, /isRunPendingPublish/);
  assert.match(text, /运行已完成，尚未发布到今日 AI 精选/);
  assert.match(text, /去审核发布/);
  assert.match(text, /stock-selection-candidates/);
});
