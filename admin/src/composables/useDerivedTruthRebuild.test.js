import test from "node:test";
import assert from "node:assert/strict";
import { useDerivedTruthRebuild } from "./useDerivedTruthRebuild.js";

function createFeedback() {
  return { clear() {}, setMessage() {}, setError(message) { throw new Error(message); } };
}

test("useDerivedTruthRebuild refreshes summaries after stock rebuild", async () => {
  let refreshCalls = 0;
  const state = useDerivedTruthRebuild(
    {
      feedback: createFeedback(),
      canEditMarket: true,
      refreshQualityDashboard: async () => {
        refreshCalls += 1;
      }
    },
    {
      getMarketDerivedTruthSummary: async ({ asset_class }) => ({ asset_class, truth_bar_count: asset_class === "STOCK" ? 6 : 3 }),
      rebuildStockDerivedTruth: async () => ({ trade_date: "2026-03-24", truth_bar_count: 6, stock_status_count: 5 }),
      rebuildFuturesDerivedTruth: async () => ({ trade_date: "2026-03-24", truth_bar_count: 3, futures_mapping_count: 2 })
    }
  );

  await state.fetchDerivedTruthSummaries();
  await state.handleRebuildTruth("STOCK");

  assert.equal(state.stockRebuildSummary.value.truth_bar_count, 6);
  assert.equal(refreshCalls, 1);
});
