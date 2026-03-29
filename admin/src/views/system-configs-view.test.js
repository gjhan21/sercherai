import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "SystemConfigsView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("SystemConfigsView masks sensitive config values by default", () => {
  const text = readView();
  assert.match(text, /function isSensitiveConfigKey\(/);
  assert.match(text, /key\.includes\("source_key"\) \|\| key\.includes\("policy_key"\)/);
  assert.match(text, /function maskSensitiveConfigValue\(/);
  assert.match(text, /function formatListConfigValue\(/);
  assert.match(text, /const revealSensitiveListValues = ref\(false\);/);
  assert.match(text, /const revealDialogSensitiveValue = ref\(false\);/);
  assert.match(text, /敏感配置默认脱敏/);
  assert.match(text, /显示敏感配置值/);
  assert.match(text, /隐藏敏感配置值/);
  assert.match(text, /显示原文后可编辑/);
  assert.match(text, /include_sensitive: revealSensitiveListValues\.value && canEditSystemConfigs \? "true" : undefined/);
  assert.match(text, /仅具备编辑权限的管理员可显示敏感配置值/);
  assert.match(text, /<el-button v-if="canEditSystemConfigs" @click="toggleSensitiveListValues">/);
  assert.match(text, /编辑前请先点击“显示敏感配置值”/);
  assert.match(text, /formatListConfigValue\(row\.config_key, row\.config_value\)/);
});

test("SystemConfigsView exposes forecast enhancement config tab", () => {
  const text = readView();
  assert.match(text, /forecast-admin/);
  assert.match(text, /const forecastLoading = ref\(false\);/);
  assert.match(text, /const forecastForm = reactive\(/);
  assert.match(text, /async function fetchForecastConfig\(/);
  assert.match(text, /async function saveForecastConfig\(/);
  assert.match(text, /预测增强配置/);
  assert.match(text, /L1 仅做 explanation 增强与 advisory 提示/);
  assert.match(text, /L2 仅补充 relationship snapshot \/ stable scenarios \/ veto 摘要展示/);
  assert.match(text, /growth\.forecast_l1\.enabled/);
  assert.match(text, /growth\.forecast_l1\.advisory_priority_threshold/);
  assert.match(text, /growth\.forecast_l2\.enabled/);
  assert.match(text, /growth\.forecast_l2\.relationship_snapshot_enabled/);
  assert.match(text, /growth\.forecast_l2\.stable_scenarios_enabled/);
  assert.match(text, /growth\.forecast_l2\.veto_confidence_threshold/);
  assert.match(text, /growth\.forecast_l3\.enabled/);
  assert.match(text, /growth\.forecast_l3\.admin_manual_enabled/);
  assert.match(text, /growth\.forecast_l3\.user_request_enabled/);
  assert.match(text, /growth\.forecast_l3\.auto_priority_enabled/);
  assert.match(text, /growth\.forecast_l3\.client_read_enabled/);
  assert.match(text, /growth\.forecast_l3\.require_vip_for_full_report/);
  assert.match(text, /growth\.forecast_l3\.max_active_runs/);
  assert.match(text, /growth\.forecast_l3\.max_runs_per_day/);
  assert.match(text, /growth\.forecast_l3\.max_user_runs_per_day/);
  assert.match(text, /growth\.forecast_l3\.min_priority_threshold/);
  assert.match(text, /growth\.forecast_l3\.dispatch\.enabled/);
  assert.match(text, /growth\.forecast_l3\.dispatch\.interval_minutes/);
  assert.match(text, /growth\.forecast_l3\.quality\.enabled/);
  assert.match(text, /growth\.forecast_l3\.quality\.interval_minutes/);
  assert.match(text, /growth\.forecast_l3\.default_engine_key/);
  assert.match(text, /L3 深推演只做异步增强/);
  assert.match(text, /LOCAL_SYNTHESIS/);
});
