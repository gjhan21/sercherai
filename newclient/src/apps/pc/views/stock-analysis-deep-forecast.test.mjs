import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const filePath = "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-deep-forecast-client-integration/newclient/src/apps/pc/views/analysis/StockAnalysis.vue";

test("stock analysis view renders deep forecast summary entry surface", () => {
  const text = fs.readFileSync(filePath, "utf8");
  assert.match(text, /DeepForecastSummaryCard/);
  assert.match(text, /useDeepForecastEntry/);
  assert.match(text, /深度推演/);
  assert.match(text, /forecastEntryTo/);
});
