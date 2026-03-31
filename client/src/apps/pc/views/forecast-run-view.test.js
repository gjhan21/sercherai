import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function read(...parts) {
  return fs.readFileSync(path.join(__dirname, ...parts), "utf8");
}

test("forecast run route and entry surfaces are wired into client views", () => {
  const routerText = read("../router", "index.js");
  const strategyText = read("StrategyView.vue");
  const archiveText = read("RecommendationArchiveView.vue");
  const forecastViewText = read("ForecastRunView.vue");

  assert.match(routerText, /name:\s*"forecast-run"/);
  assert.match(routerText, /path:\s*"forecast\/runs\/:id"/);

  assert.match(strategyText, /发起深推演/);
  assert.match(strategyText, /查看深推演/);
  assert.match(archiveText, /发起深推演/);
  assert.match(archiveText, /查看深推演/);

  assert.match(forecastViewText, /深推演报告/);
  assert.match(forecastViewText, /运行日志/);
});
