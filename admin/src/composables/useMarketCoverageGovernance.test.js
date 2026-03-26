import test from "node:test";
import assert from "node:assert/strict";
import { useMarketCoverageGovernance } from "./useMarketCoverageGovernance.js";

function createFeedback() {
  return { clear() {}, setMessage() {}, setError(message) { throw new Error(message); } };
}

test("useMarketCoverageGovernance loads summary, snapshots, and runs", async () => {
  const pushes = [];
  const state = useMarketCoverageGovernance(
    { feedback: createFeedback(), router: { push(payload) { pushes.push(payload); } } },
    {
      getMarketCoverageSummary: async () => ({ latest_trade_date: "2026-03-24", stock_universe_count: 10 }),
      listMarketUniverseSnapshots: async () => ({ items: [{ snapshot_date: "2026-03-24" }] }),
      listMarketDataBackfillRuns: async () => ({ items: [{ status: "SUCCESS" }] })
    }
  );

  await state.fetchMarketCoverageGovernance();
  state.openMarketBackfillWorkspace();

  assert.equal(state.marketUniverseSnapshots.value.length, 1);
  assert.equal(state.marketBackfillRuns.value.length, 1);
  assert.equal(pushes[0].path, "/system-jobs");
});
