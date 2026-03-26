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
