import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "FuturesStrategyDetail.vue");

test("futures strategy detail routes back to futures center and keeps futures forecast semantics", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /\/futures\/strategies/);
  assert.match(text, /targetType:\s*"FUTURES"/);
  assert.match(text, /sourcePath:\s*current\.id \? `\/futures\/strategy\/\$\{current\.id\}` : "\/futures\/strategies"/);
  assert.match(text, /DeepForecastSummaryCard/);
  assert.match(text, /useDeepForecastEntry/);
});
