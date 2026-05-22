import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "forecast", "ForecastDetailView.vue");

test("h5 forecast detail view renders report and log surfaces", () => {
  const text = fs.readFileSync(filePath, "utf8");
  assert.match(text, /深度推演报告/);
  assert.match(text, /运行日志/);
  assert.match(text, /useForecastRunDetail/);
  assert.match(text, /getMembershipQuota/);
});
