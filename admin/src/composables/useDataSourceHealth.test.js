import test from "node:test";
import assert from "node:assert/strict";
import { useDataSourceHealth } from "./useDataSourceHealth.js";

function createFeedback() {
  return { clear() {}, setMessage() {}, setError(message) { throw new Error(message); } };
}

test("useDataSourceHealth stores single-check result and logs", async () => {
  const state = useDataSourceHealth(
    { feedback: createFeedback(), canEditDataSources: true },
    {
      checkDataSourceHealth: async (sourceKey) => ({ source_key: sourceKey, status: "HEALTHY" }),
      batchCheckDataSources: async () => ({ items: [] }),
      listDataSourceHealthLogs: async () => ({ items: [{ status: "HEALTHY", checked_at: "2026-03-25" }] })
    }
  );

  await state.handleCheckOne("TUSHARE");
  await state.showLogs("TUSHARE");

  assert.equal(state.healthMap.value.TUSHARE.status, "HEALTHY");
  assert.equal(state.logSourceKey.value, "TUSHARE");
  assert.equal(state.logItems.value.length, 1);
});
