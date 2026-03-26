import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "WorkflowMessagesView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("WorkflowMessagesView upgrades into unified inbox shell", () => {
  const text = readView();
  assert.match(text, /listAuditEvents/);
  assert.match(text, /getAuditEventSummary/);
  assert.match(text, /const messageCenterTab = ref\("workflow"\);/);
  assert.match(text, /const auditSummary = ref\(null\);/);
  assert.match(text, /const relatedAuditItems = ref\(\[\]\);/);
  assert.match(text, /const auditFeatureUnavailable = ref\(false\);/);
  assert.match(text, /const auditFeatureError = ref\(""\);/);
  assert.match(text, /const auditFilters = reactive\(/);
  assert.match(text, /isAuditFeatureUnavailableError/);
  assert.match(text, /markAuditFeatureUnavailable/);
  assert.match(text, /clearAuditFeatureUnavailable/);
  assert.match(text, /async function fetchAuditSummary/);
  assert.match(text, /async function fetchRelatedAuditEvents/);
  assert.match(text, /开放事件（未启用）/);
  assert.match(text, /开放事件当前不可用/);
  assert.match(text, /migrate audit/);
  assert.match(text, /复制迁移命令/);
  assert.match(text, /function resolvedAuditCount\(/);
  assert.match(text, /function resolveAuditObjectRoute\(/);
  assert.match(text, /function openAuditObject\(/);
  assert.match(text, /review_id/);
  assert.match(text, /run_id/);
  assert.match(text, /source_key/);
  assert.match(text, /STRATEGY_JOB/);
  assert.match(text, /STRATEGY_PUBLISH_POLICY/);
  assert.match(text, /STRATEGY_SEED_SET/);
  assert.match(text, /STRATEGY_AGENT_PROFILE/);
  assert.match(text, /STRATEGY_SCENARIO_TEMPLATE/);
  assert.match(text, /publish_id/);
  assert.match(text, /policy_id/);
  assert.match(text, /config_type/);
  assert.match(text, /config_id/);
  assert.match(text, /engine-config/);
  assert.match(text, /market-center/);
  assert.match(text, /name: "data-sources-registry"/);
  assert.match(text, /action: "logs"/);
  assert.match(text, /对象类型/);
  assert.match(text, /已关闭事件/);
  assert.match(text, /跳转对象页/);
  assert.match(text, /消息中心/);
  assert.match(text, /流程待办/);
  assert.match(text, /开放事件/);
  assert.match(text, /前往审计日志/);
  assert.match(text, /待处理审计事件/);
  assert.match(text, /openAuditDetail/);
});
