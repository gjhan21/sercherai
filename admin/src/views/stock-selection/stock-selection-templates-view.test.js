import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "StockSelectionTemplatesView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("StockSelectionTemplatesView sends layered defaults buckets", () => {
  const text = readView();
  assert.match(text, /market_analysis_defaults_json/);
  assert.match(text, /candidate_pool_defaults_json/);
  assert.match(text, /short_term_head_defaults_json/);
  assert.match(text, /swing_head_defaults_json/);
});

test("StockSelectionTemplatesView exposes short-term and swing controls", () => {
  const text = readView();
  assert.match(text, /short_term_limit/);
  assert.match(text, /short_term_min_score/);
  assert.match(text, /swing_limit/);
});
