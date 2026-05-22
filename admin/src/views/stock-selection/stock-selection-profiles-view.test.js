import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "StockSelectionProfilesView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("StockSelectionProfilesView builds dual-head config payloads", () => {
  const text = readView();
  assert.match(text, /market_analysis_config/);
  assert.match(text, /candidate_pool_config/);
  assert.match(text, /short_term_head_config/);
  assert.match(text, /swing_head_config/);
});

test("StockSelectionProfilesView keeps market and head level defaults visible", () => {
  const text = readView();
  assert.match(text, /market_regime_bias/);
  assert.match(text, /candidate_pool_limit/);
  assert.match(text, /watchlist_limit/);
});
