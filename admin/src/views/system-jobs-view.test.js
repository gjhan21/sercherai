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

test("SystemJobsView reads unified audit event summary for scheduler operations", () => {
  const text = readView();
  assert.match(text, /getAuditEventSummary/);
  assert.match(text, /listAuditEvents/);
  assert.match(text, /useRoute/);
  assert.match(text, /const route = useRoute\(\);/);
  assert.match(text, /useRouter/);
  assert.match(text, /function applyRunRouteFocus\(/);
  assert.match(text, /watch\(\s*\(\) => route\.query\.run_id/);
  assert.match(text, /watch\(\s*\(\) => route\.query\.job_name/);
  assert.match(text, /query: \{ run_id: currentRun\.value\.id \}/);
  assert.match(text, /function openJobAuditInbox\(\)/);
  assert.match(text, /router\.push\("\/workflow-messages"\)/);
  assert.match(text, /const jobAuditSummary = ref\(null\);/);
  assert.match(text, /const jobAuditItems = ref\(\[\]\);/);
  assert.match(text, /async function fetchJobAuditEvents/);
  assert.match(text, /任务事件摘要/);
  assert.match(text, /查看消息中心/);
  assert.match(text, /近期任务事件/);
  assert.match(text, /SCHEDULER_JOB/);
});
