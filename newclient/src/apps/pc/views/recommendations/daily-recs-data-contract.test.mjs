import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const dailyRecsFile = path.join(__dirname, "DailyRecs.vue");

test("daily recs view requests trade_date and avoids fake quote fields", () => {
  const text = fs.readFileSync(dailyRecsFile, "utf8");

  assert.match(text, /trade_date/);
  assert.doesNotMatch(text, /Math\.random\(\)\s*\*\s*6\s*-\s*1/);
  assert.doesNotMatch(text, /String\(item\.score\s*\|\|/);
});

test("daily recs view uses login gate instead of mock recommendation pool", () => {
  const text = fs.readFileSync(dailyRecsFile, "utf8");

  assert.match(text, /登录后查看今日 AI 精选/);
  assert.doesNotMatch(text, /const dailyRecs = ref\(MOCK_RECS\)/);
});

test("daily recs view renders real card fields for recommendation credibility", () => {
  const text = fs.readFileSync(dailyRecsFile, "utf8");

  assert.match(text, /风险等级/);
  assert.match(text, /建议仓位/);
  assert.match(text, /止盈/);
  assert.match(text, /止损/);
});
