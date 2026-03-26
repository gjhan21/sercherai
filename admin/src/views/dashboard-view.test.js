import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "DashboardView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("DashboardView links data source operations to dedicated child pages", () => {
  const text = readView();
  assert.match(text, /数据源管理/);
  assert.match(text, /数据治理/);
  assert.match(text, /数据同步/);
  assert.match(text, /数据质量/);
  assert.match(text, /path: "\/data-sources\/governance"/);
  assert.match(text, /path: "\/data-sources\/sync"/);
  assert.match(text, /path: "\/data-sources\/quality"/);
  assert.match(text, /path: "\/data-sources\/registry"/);
});
