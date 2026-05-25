import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "SystemJobsView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("SystemJobsView reads operation logs for auto retry config changes", () => {
  const text = readView();
  assert.match(text, /listOperationLogs/);
  assert.match(text, /async function fetchAutoRetryChangeLogs/);
  assert.match(text, /module: "SYSTEM"/);
  assert.match(text, /action: "UPSERT_CONFIG"/);
  assert.match(text, /SYSTEM_CONFIG/);
  assert.match(text, /scheduler\.auto_retry/);
  assert.match(text, /autoRetryChangeLogs = ref\(\[\]\)/);
});

test("SystemJobsView wires market sync cancel action for pending and running runs", () => {
  const text = readView();
  assert.match(text, /cancelMarketDataBackfillRun/);
  assert.match(text, /async function handleCancelMarketBackfillRun/);
  assert.match(text, /row\.status === ['"]PENDING['"] \|\| row\.status === ['"]RUNNING['"]/);
  assert.match(text, /取消任务/);
});

test("SystemJobsView keeps cancelled sync tasks retryable", () => {
  const text = readView();
  assert.match(text, /function canRetryMarketBackfillRun/);
  assert.match(text, /status === ["']CANCELLED["']/);
});

test("SystemJobsView exposes stock sync shortcuts in manual trigger tab", () => {
  const text = readView();
  assert.match(text, /股票全量同步/);
  assert.match(text, /当日股票数据同步/);
  assert.match(text, /template: ["']STOCK_FULL["']/);
  assert.match(text, /template: ["']STOCK_INCREMENTAL["']/);
  assert.match(text, /@click=["']handleQuickStockSync\(item\.template\)["']/);
  assert.match(text, /buildSyncJobPayloadFromTemplate\(templateKey/);
  assert.match(text, /createMarketDataBackfillRun\(payload\)/);
});

test("SystemJobsView keeps stock sync shortcuts separate from manual trigger task codes", () => {
  const text = readView();
  assert.doesNotMatch(text, /stock_full_sync/);
  assert.doesNotMatch(text, /stock_daily_incremental_sync/);
  assert.doesNotMatch(text, /getStockSyncTemplateForJobName/);
  assert.doesNotMatch(text, /showTriggerLoading/);
  assert.match(text, /async function submitTrigger\(\)/);
  assert.match(text, /triggerSchedulerJob\(payload\)/);
});
