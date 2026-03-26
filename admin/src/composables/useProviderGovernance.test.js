import test from "node:test";
import assert from "node:assert/strict";
import { useProviderGovernance } from "./useProviderGovernance.js";

function createFeedback() {
  return { clear() {}, setMessage() {}, setError(message) { throw new Error(message); } };
}

test("useProviderGovernance loads dashboard and normalizes policy form", async () => {
  let savedPayload = null;
  const state = useProviderGovernance(
    { feedback: createFeedback(), canEditDataSources: true, getLookbackHours: () => 24 },
    {
      getMarketProviderGovernanceOverview: async ({ asset_class }) => ({ asset_class, provider_scores: [] }),
      listMarketProviderCapabilities: async () => ({ items: [{ provider_key: "TUSHARE", asset_class: "STOCK", data_kind: "DAILY_BARS" }] }),
      listMarketProviderRoutingPolicies: async () => ({ items: [{
        policy_key: "stock.daily_bars",
        asset_class: "stock",
        data_kind: "daily_bars",
        primary_provider_key: "tushare",
        fallback_provider_keys: ["akshare"],
        fallback_allowed: true,
        mock_allowed: false,
        quality_threshold: 0.8
      }] }),
      updateMarketProviderRoutingPolicy: async (policyKey, payload) => {
        savedPayload = { policyKey, payload };
      }
    }
  );

  await state.fetchGovernanceDashboard();
  state.handleEditRoutingPolicy(state.routingPolicies.value[0]);
  assert.equal(state.providerCapabilities.value.length, 1);
  assert.equal(state.routingPolicyForm.primary_provider_key, "TUSHARE");
  await state.handleSaveRoutingPolicy();

  assert.equal(savedPayload.policyKey, "stock.daily_bars");
  assert.equal(savedPayload.payload.primary_provider_key, "TUSHARE");
  assert.deepEqual(savedPayload.payload.fallback_provider_keys, ["AKSHARE"]);
});
