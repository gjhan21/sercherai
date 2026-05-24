import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "forecast", "ForecastDetailView.vue");

test("pc forecast detail view renders report and log surfaces", () => {
  const text = fs.readFileSync(filePath, "utf8");
  assert.match(text, /深推演报告/);
  assert.match(text, /同标的历史/);
  assert.match(text, /最新一次 vs 上一次/);
  assert.match(text, /结论变化/);
  assert.match(text, /证据变化/);
  assert.match(text, /完整复盘评分/);
  assert.match(text, /查看全部成功 run/);
  assert.match(text, /当前状态/);
  assert.match(text, /核心判断/);
  assert.match(text, /主情景/);
  assert.match(text, /风险边界/);
  assert.match(text, /后续操作建议/);
  assert.match(text, /证据支撑/);
  assert.match(text, /模型复核/);
  assert.match(text, /运行证据|运行日志/);
  assert.match(text, /备选情景/);
  assert.match(text, /验证清单/);
  assert.match(text, /失效信号/);
  assert.match(text, /角色分歧/);
  assert.match(text, /useForecastRunDetail/);
  assert.match(text, /getForecastRunHistory|getForecastRunCompare|getForecastRunReview/);
  assert.match(text, /buildForecastHistoryViewModel/);
  assert.match(text, /localizeForecastDimension/);
  assert.match(text, /getMembershipQuota/);
  assert.match(text, /buildForecastEvidenceSections/);
  assert.match(text, /已进入推演队列|正在生成深推演报告|本次深推演未完成|不存在或已失效/);
  assert.match(text, /一句话结论|执行摘要/);
  assert.match(text, /发起时间|完成时间|回深度推演工作台/);
  assert.match(text, /当前立场|支撑点|风险点|置信度/);
  assert.doesNotMatch(text, />Forecast L3<|NOT FOUND|QUEUED|RUNNING|FAILED/);
});

test("pc forecast detail view stacks long metric values vertically", () => {
  const text = fs.readFileSync(filePath, "utf8");
  assert.match(text, /\.forecast-grid article \{\s*display: grid;/);
  assert.match(text, /\.forecast-grid strong \{\s*text-align: left;/);
});
