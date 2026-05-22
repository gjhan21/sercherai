import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";

const filePath = path.resolve("admin/src/views/stock-selection/StockSelectionOverviewView.vue");
const text = fs.readFileSync(filePath, "utf8");

test("StockSelectionOverviewView renders evaluation backfill state card", () => {
  assert.match(text, /评估回填状态/);
  assert.match(text, /evaluation_backfill_state/);
  assert.match(text, /formatStockSelectionEvaluationStatus/);
  assert.match(text, /missing_horizons/);
});
