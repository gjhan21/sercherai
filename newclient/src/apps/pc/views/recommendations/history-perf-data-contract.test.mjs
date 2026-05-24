import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const historyPerfFile = path.join(__dirname, "HistoryPerf.vue");

test("history perf view uses real history endpoint instead of current recommendations endpoint", () => {
  const text = fs.readFileSync(historyPerfFile, "utf8");

  assert.match(text, /listStockRecommendationHistory/);
  assert.doesNotMatch(text, /listStockRecommendations\(/);
});

test("history perf view removes mock and random performance generation", () => {
  const text = fs.readFileSync(historyPerfFile, "utf8");

  assert.doesNotMatch(text, /REC_HISTORY as MOCK_HISTORY/);
  assert.doesNotMatch(text, /Math\.random/);
});

test("history perf view renders real performance columns and summary-driven statistics", () => {
  const text = fs.readFileSync(historyPerfFile, "utf8");

  assert.match(text, /summary/);
  assert.match(text, /建仓价/);
  assert.match(text, /最新\/结算价/);
  assert.match(text, /真实收益/);
  assert.match(text, /最大回撤/);
});

test("history perf view uses login gate instead of mock history when logged out", () => {
  const text = fs.readFileSync(historyPerfFile, "utf8");

  assert.match(text, /登录后查看历史推荐表现/);
  assert.doesNotMatch(text, /const historyData = ref\(MOCK_HISTORY\)/);
});
