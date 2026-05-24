import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "forecast", "ForecastLabView.vue");

test("pc forecast lab view exposes visible landing surfaces", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /深度研究中心|研究入口|研究中心/);
  assert.match(text, /listForecastRuns/);
  assert.match(text, /getForecastRunHistory|getForecastRunCompare|getForecastRunReview/);
  assert.match(text, /历史研究概览/);
  assert.match(text, /查看同标的历史对比/);
  assert.match(text, /成功 run 数量/);
  assert.match(text, /最近复盘评分趋势/);
  assert.match(text, /当前仅有 1 次成功深推演，暂不能形成历史对比/);
  assert.match(text, /getLatestSuccessfulForecastRun/);
  assert.match(text, /最近研究结论/);
  assert.match(text, /运行动态/);
  assert.match(text, /值得继续深挖|研究聚焦/);
});

test("pc forecast lab view exposes focused target context and next action", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /当前推演对象/);
  assert.match(text, /最近研究状态|是否完成模型复核|结构化报告/);
  assert.match(text, /登录后查看该标的深度推演|查看这只标的最近深度推演|立即发起这只标的的深度推演|请先从来源页进入或先完成标的解析/);
  assert.match(text, /聚焦该标的最近运行|查看全部运行清单/);
  assert.match(text, /createForecastRun/);
  assert.match(text, /target_id/);
  assert.match(text, /source_path/);
});
