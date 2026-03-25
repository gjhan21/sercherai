import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "AuditLogsView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("AuditLogsView stays as audit and operation log fallback entry", () => {
  const text = readView();
  assert.match(text, /listAuditEvents/);
  assert.match(text, /getAuditEventSummary/);
  assert.match(text, /const activeTab = ref\("audit-events"\);/);
  assert.match(text, /const auditItems = ref\(\[\]\);/);
  assert.match(text, /const auditSummary = ref\(null\);/);
  assert.match(text, /async function fetchAuditEvents/);
  assert.match(text, /async function fetchAuditSummary/);
  assert.match(text, /function resolvedAuditCount\(/);
  assert.match(text, /function resolveAuditObjectRoute\(/);
  assert.match(text, /function openAuditObject\(/);
  assert.match(text, /review_id/);
  assert.match(text, /run_id/);
  assert.match(text, /source_key/);
  assert.match(text, /STRATEGY_JOB/);
  assert.match(text, /STRATEGY_PUBLISH_POLICY/);
  assert.match(text, /publish_id/);
  assert.match(text, /policy_id/);
  assert.match(text, /engine-config/);
  assert.match(text, /market-center/);
  assert.match(text, /action: "logs"/);
  assert.match(text, /审计与操作日志/);
  assert.match(text, /返回消息中心/);
  assert.match(text, /统一审计事件/);
  assert.match(text, /事件域/);
  assert.match(text, /告警等级/);
  assert.match(text, /开放事件/);
  assert.match(text, /已关闭事件/);
  assert.match(text, /跳转对象页/);
});
