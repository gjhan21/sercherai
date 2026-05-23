import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const dailyRecsFile = path.join(__dirname, "DailyRecs.vue");
const strategiesFile = path.join(__dirname, "Strategies.vue");
const forecastLabFile = path.join(__dirname, "..", "forecast", "ForecastLabView.vue");

test("daily recommendations exposes workflow CTA toward strategies and forecast lab", () => {
  const text = fs.readFileSync(dailyRecsFile, "utf8");

  assert.match(text, /发现机会/);
  assert.match(text, /查看对应策略/);
  assert.match(text, /进入深度推演/);
  assert.match(text, /buildForecastContextQuery/);
  assert.match(text, /targetId:\s*stock\?\.id/);
  assert.match(text, /source:\s*"RECOMMENDATION"/);
  assert.match(text, /sourcePath:\s*"\/recommendations"/);
});

test("strategies view exposes workflow CTA toward forecast lab and source recommendations", () => {
  const text = fs.readFileSync(strategiesFile, "utf8");

  assert.match(text, /形成策略/);
  assert.match(text, /回看推荐来源/);
  assert.match(text, /进入深度推演/);
  assert.match(text, /buildForecastContextQuery/);
  assert.match(text, /targetId:\s*strategy\?\.id/);
  assert.match(text, /source:\s*"STRATEGY"/);
  assert.match(text, /sourcePath:\s*"\/recommendations\/strategies"/);
});

test("forecast lab view surfaces recommendation strategy workflow context", () => {
  const text = fs.readFileSync(forecastLabFile, "utf8");

  assert.match(text, /当前主线/);
  assert.match(text, /来自每日推荐|来自交易策略|独立进入/);
  assert.match(text, /回到交易策略|回到每日推荐|去股票分析/);
  assert.match(text, /请先从来源页进入或先完成标的解析/);
  assert.match(text, /resolveForecastContextFromRoute/);
  assert.match(text, /target_id/);
  assert.match(text, /source_path|sourcePath/);
});
