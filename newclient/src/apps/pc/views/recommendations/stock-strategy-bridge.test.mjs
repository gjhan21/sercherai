import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "StockStrategyBridge.vue");

test("stock strategy bridge consumes stock recommendation context and no futures list", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /reco_id/);
  assert.match(text, /symbol/);
  assert.match(text, /name/);
  assert.match(text, /请先从每日推荐进入策略承接页/);
  assert.match(text, /getStockRecommendationDetail/);
  assert.match(text, /getStockRecommendationInsight/);
  assert.match(text, /buildForecastContextQuery/);
  assert.match(text, /targetType:\s*"STOCK"/);
  assert.doesNotMatch(text, /listFuturesStrategies/);
});
