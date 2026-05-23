import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "StockDetail.vue");

test("h5 stock detail renders deep forecast summary card below AI brief", () => {
  const text = fs.readFileSync(filePath, "utf8");
  assert.match(text, /DeepForecastSummaryCard/);
  assert.match(text, /useDeepForecastEntry/);
  assert.match(text, /深度推演/);
  assert.match(text, /buildForecastContextQuery/);
  assert.match(text, /带着完整上下文进入深度推演/);
});
