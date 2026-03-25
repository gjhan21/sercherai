import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "FuturesSelectionRunsView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("FuturesSelectionRunsView keeps run detail panel", () => {
  const text = readView();
  assert.match(text, /运行详情/);
  assert.match(text, /selected_trade_date/);
});

test("FuturesSelectionRunsView surfaces governance routing summary", () => {
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
