import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const filePath = "/Users/gjhan21/cursor/sercherai/.worktrees/newclient-deep-forecast-client-integration/newclient/src/apps/h5/views/futures/FuturesArbitrageDetail.vue";

test("h5 futures detail exposes a minimal deep forecast entry surface", () => {
  const text = fs.readFileSync(filePath, "utf8");
  assert.match(text, /深度推演|查看完整深度推演|forecast/i);
});
