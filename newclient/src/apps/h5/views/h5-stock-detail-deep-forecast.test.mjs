import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const filePath = "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-deep-forecast-client-integration/newclient/src/apps/h5/views/StockDetail.vue";

test("h5 stock detail renders deep forecast summary card below AI brief", () => {
  const text = fs.readFileSync(filePath, "utf8");
  assert.match(text, /DeepForecastSummaryCard/);
  assert.match(text, /useDeepForecastEntry/);
  assert.match(text, /深度推演/);
});
