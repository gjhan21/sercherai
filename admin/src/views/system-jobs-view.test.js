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
