import test from "node:test";
import assert from "node:assert/strict";
import { nextTick, reactive } from "vue";
import { useMarketQualityLogs } from "./useMarketQualityLogs.js";

function createFeedback() {
  return { clear() {}, setMessage() {}, setError(message) { throw new Error(message); } };
}

test("useMarketQualityLogs syncs filters to route and opens payload dialog", async () => {
  const route = reactive({ query: {} });
  const replaced = [];
  const state = useMarketQualityLogs(
    {
      feedback: createFeedback(),
      route,
      router: {
        replace(payload) {
          replaced.push(payload);
          route.query = payload.query;
        }
      }
    },
    {
      listMarketDataQualityLogs: async () => ({ items: [{ issue_code: "MISSING_BAR", payload: '{"ok":true}' }], total: 1 }),
      getMarketDataQualitySummary: async () => ({ total_count: 1 })
    }
  );

  await nextTick();
  state.qualityFilters.issue_code = "MISSING_BAR";
  state.syncQualityFiltersToRoute();
  state.openQualityPayload({ issue_code: "MISSING_BAR", payload: '{"ok":true}' });

  assert.equal(replaced.length, 1);
  assert.equal(replaced[0].name, "data-sources-quality");
  assert.equal(state.qualityPayloadDialogVisible.value, true);
  assert.equal(state.qualityLogItems.value.length, 1);
});
