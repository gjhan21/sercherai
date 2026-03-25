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

test("StrategyEngineConfigPanel exposes publish policy focus helpers for deep links", () => {
  const text = readView();
  assert.match(text, /focusPublishPolicyByID/);
  assert.match(text, /policyTableRef/);
  assert.match(text, /focusedPolicyID/);
  assert.match(text, /row-class-name/);
  assert.match(text, /defineExpose\(\{ refreshAll, focusPublishPolicyByID \}\)/);
});
