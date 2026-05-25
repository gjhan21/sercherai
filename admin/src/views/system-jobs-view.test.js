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

test("SystemJobsView uses the streamlined three-tab layout without config or trigger main tabs", () => {
  const text = readView();
  assert.doesNotMatch(text, /v-show="activeTab === 'config'"/);
  assert.doesNotMatch(text, /v-show="activeTab === 'trigger'"/);
  assert.doesNotMatch(text, /按任务维度的重试恢复统计/);
  assert.doesNotMatch(text, /按任务失败原因/);
  assert.doesNotMatch(text, /当前窗口提示/);
  assert.doesNotMatch(text, /jobs-action-grid/);
  assert.doesNotMatch(text, /任务配置<\/h3>/);
  assert.doesNotMatch(text, /手动触发任务<\/h3>/);
  assert.match(text, /系统任务中心/);
  assert.match(text, /同步任务/);
  assert.match(text, /运行记录/);
});
