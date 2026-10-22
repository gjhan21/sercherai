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

// The review centre was rewritten around three tabs. These assertions pin the
// behaviour that rewrite kept: one API-driven workspace, SLA-aware metrics,
// batch decisions and a CSV escape hatch.
test("ReviewCenterView drives one API-backed workspace with SLA metrics", () => {
  const text = readView();

  assert.match(text, /getWorkflowMetrics/);
  assert.match(text, /listReviewTasks/);
  assert.match(text, /reviewTaskDecision/);
  assert.match(text, /const slaStats = computed\(/);
  assert.match(text, /SLA超时/);
  assert.match(text, /智能运营审核中心/);
});

test("ReviewCenterView maps each tab onto its own status filter", () => {
  const text = readView();

  assert.match(text, /apiFilters\.status = "PENDING"/);
  assert.match(text, /apiFilters\.reviewer_id = "NULL"/);
  assert.match(text, /apiFilters\.status = "APPROVED,REJECTED"/);
  assert.match(text, /待我处理/);
  assert.match(text, /未分配队列/);
  assert.match(text, /全量历史/);
});

test("ReviewCenterView supports batch approval and CSV export", () => {
  const text = readView();

  assert.match(text, /async function handleBatchApprove\(/);
  assert.match(text, /function exportCSV\(/);
  assert.match(text, /导出 CSV/);
});

test("ReviewCenterView parses the forecast admin config map it loads", () => {
  const text = readView();

  assert.match(text, /listSystemConfigs/);
  assert.match(text, /parseForecastAdminConfigMap/);
  assert.match(text, /const forecastReviewConfig = ref\(null\);/);
  assert.match(text, /keyword: "growth\.forecast_"/);
});
