import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "StockSelectionEvaluationView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("StockSelectionEvaluationView exposes evaluation scope filter", () => {
  const text = readView();
  assert.match(text, /evaluation_scope/);
  assert.match(text, /SHORT_TERM_PRIMARY/);
  assert.match(text, /SWING_AUXILIARY/);
  assert.match(text, /CANDIDATE_POOL/);
});

test("StockSelectionEvaluationView formats leaderboard rows with layered scope labels", () => {
  const text = readView();
  assert.match(text, /formatStockSelectionEvaluationScope/);
  assert.match(text, /listStockSelectionEvaluationLeaderboard/);
});

test("StockSelectionEvaluationView surfaces evaluation backfill state summary", () => {
  const text = readView();
  assert.match(text, /getStockSelectionOverview/);
  assert.match(text, /evaluation_backfill_state/);
  assert.match(text, /formatStockSelectionEvaluationStatus/);
});
