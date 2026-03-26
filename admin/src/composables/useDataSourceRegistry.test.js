import test from "node:test";
import assert from "node:assert/strict";
import { useDataSourceRegistry } from "./useDataSourceRegistry.js";

function createFeedback() {
  return {
    clearCalls: 0,
    messages: [],
    errors: [],
    clear() {
      this.clearCalls += 1;
    },
    setMessage(message) {
      this.messages.push(message);
    },
    setError(message) {
      this.errors.push(message);
    }
  };
}

test("useDataSourceRegistry loads items and stock default source", async () => {
  const feedback = createFeedback();
  const state = useDataSourceRegistry(
    { feedback, canEditDataSources: true },
    {
      listDataSources: async () => ({ items: [{ source_key: "tushare", name: "Tushare" }], total: 1 }),
      listSystemConfigs: async ({ keyword }) => ({
        items: [{ config_key: keyword, config_value: keyword.includes("market.news") ? "akshare" : "tushare" }]
      })
    }
  );

  await state.fetchDataSources();

  assert.equal(state.items.value.length, 1);
  assert.equal(state.total.value, 1);
  assert.equal(state.defaultStockSourceKey.value, "TUSHARE");
  assert.equal(state.defaultFuturesSourceKey.value, "TUSHARE");
  assert.equal(state.defaultMarketNewsSourceKey.value, "AKSHARE");
});

test("useDataSourceRegistry persists built-in provider metadata for TUSHARE", async () => {
  const feedback = createFeedback();
  let createdPayload = null;
  const state = useDataSourceRegistry(
    { feedback, canEditDataSources: true },
    {
      listDataSources: async () => ({ items: [], total: 0 }),
      listSystemConfigs: async () => ({ items: [] }),
      createDataSource: async (payload) => {
        createdPayload = payload;
        return payload;
      }
    }
  );

  state.form.source_key = "tushare";
  state.form.name = "Tushare";
  state.form.source_type = "STOCK";
  state.form.endpoint = "";
  state.form.token = "demo-token";

  await state.submitForm();

  assert.equal(createdPayload.source_key, "TUSHARE");
  assert.equal(createdPayload.config.provider, "TUSHARE");
  assert.equal(createdPayload.config.endpoint, "https://api.tushare.pro");
});
