import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "FuturesStrategies.vue");

test("futures strategies page owns futures strategy list semantics", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /listFuturesStrategies/);
  assert.match(text, /期货策略/);
  assert.match(text, /进入深度推演/);
  assert.doesNotMatch(text, /回看推荐来源/);
  assert.doesNotMatch(text, /第 2 步：形成策略/);
  assert.doesNotMatch(text, /\/recommendations\/backtest/);
  assert.match(text, /targetType:\s*"FUTURES"/);
});
