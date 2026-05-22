import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "StockSelectionCandidatesView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("StockSelectionCandidatesView surfaces layered candidate metadata", () => {
  const text = readView();
  assert.match(text, /recommendation_head/);
  assert.match(text, /selection_layer/);
  assert.match(text, /technical_pattern/);
  assert.match(text, /formatStockSelectionEvaluationScope/);
});

test("StockSelectionCandidatesView loads layered evidence and evaluation artifacts", () => {
  const text = readView();
  assert.match(text, /listStockSelectionRunEvidence/);
  assert.match(text, /listStockSelectionRunEvaluation/);
  assert.match(text, /head_label/);
  assert.match(text, /holding_contract/);
});
