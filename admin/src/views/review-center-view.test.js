import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "ReviewCenterView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("ReviewCenterView reads unified audit event summary for review operations", () => {
  const text = readView();
  assert.match(text, /getAuditEventSummary/);
  assert.match(text, /listAuditEvents/);
  assert.match(text, /useRoute/);
  assert.match(text, /const route = useRoute\(\);/);
  assert.match(text, /useRouter/);
  assert.match(text, /function applyReviewRouteFocus\(/);
  assert.match(text, /watch\(\s*\(\) => route\.query\.review_id/);
  assert.match(text, /query: \{ review_id: task\.id \}/);
  assert.match(text, /function openReviewAuditInbox\(\)/);
  assert.match(text, /router\.push\("\/workflow-messages"\)/);
  assert.match(text, /const reviewAuditSummary = ref\(null\);/);
  assert.match(text, /const reviewAuditItems = ref\(\[\]\);/);
  assert.match(text, /async function fetchReviewAuditEvents/);
  assert.match(text, /审核事件摘要/);
  assert.match(text, /查看消息中心/);
  assert.match(text, /待处理审核事件/);
  assert.match(text, /REVIEW_TASK/);
});

test("ReviewCenterView shows forecast advisory-only configuration summary", () => {
  const text = readView();
  assert.match(text, /listSystemConfigs/);
  assert.match(text, /forecast-admin/);
  assert.match(text, /const forecastReviewConfig = ref\(null\);/);
  assert.match(text, /async function fetchForecastReviewConfig\(/);
  assert.match(text, /预测增强审核提示/);
  assert.match(text, /advisory only/);
  assert.match(text, /记忆反馈样本阈值/);
  assert.match(text, /优先级阈值/);
});
