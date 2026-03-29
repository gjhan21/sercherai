import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "ForecastLabView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("ForecastLabView wires the L3 run workbench and quality summary flow", () => {
  const text = readView();
  assert.match(text, /listForecastRuns/);
  assert.match(text, /createForecastRun/);
  assert.match(text, /getForecastRunDetail/);
  assert.match(text, /retryForecastRun/);
  assert.match(text, /cancelForecastRun/);
  assert.match(text, /listForecastQualitySummaries/);
  assert.match(text, /const runFilters = reactive\(/);
  assert.match(text, /const createForm = reactive\(/);
  assert.match(text, /const qualityFilters = reactive\(/);
  assert.match(text, /async function fetchRunList\(/);
  assert.match(text, /async function fetchRunDetail\(/);
  assert.match(text, /async function fetchQualitySummaries\(/);
  assert.match(text, /async function submitCreateRun\(/);
  assert.match(text, /async function handleRetryRun\(/);
  assert.match(text, /async function handleCancelRun\(/);
  assert.match(text, /Forecast Lab/);
  assert.match(text, /深推演运行队列/);
  assert.match(text, /质量回写摘要/);
  assert.match(text, /运行步骤日志/);
  assert.match(text, /管理员手动触发/);
});
