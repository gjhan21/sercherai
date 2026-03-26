import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "StrategyEngineConfigPanel.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("StrategyEngineConfigPanel exposes strategy config focus helpers for deep links", () => {
  const text = readView();
  assert.match(text, /focusPublishPolicyByID/);
  assert.match(text, /focusStrategyConfigItem/);
  assert.match(text, /seedSetTableRef/);
  assert.match(text, /agentProfileTableRef/);
  assert.match(text, /scenarioTemplateTableRef/);
  assert.match(text, /focusedConfigType/);
  assert.match(text, /focusedConfigID/);
  assert.match(text, /policyTableRef/);
  assert.match(text, /row-class-name/);
  assert.match(text, /defineExpose\(\{ refreshAll, focusPublishPolicyByID, focusStrategyConfigItem \}\)/);
});
