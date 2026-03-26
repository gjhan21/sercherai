import test from "node:test";
import assert from "node:assert/strict";
import { ref } from "vue";
import { useMarketSyncConsole } from "./useMarketSyncConsole.js";

function createFeedback() {
  return { clear() {}, setMessage() {}, setError(message) { throw new Error(message); } };
}

test("useMarketSyncConsole builds source options and stores sync results per card", async () => {
  const items = ref([
    { source_key: "tushare", name: "Tushare", status: "ACTIVE", config: { provider: "tushare" } }
  ]);
  const healthMap = ref({ TUSHARE: { status: "HEALTHY" } });
  const state = useMarketSyncConsole(
    {
      feedback: createFeedback(),
      canEditMarket: true,
      items,
      healthMap,
      defaultStockSourceKey: ref("TUSHARE"),
      defaultFuturesSourceKey: ref("TUSHARE"),
      defaultMarketNewsSourceKey: ref("AKSHARE")
    },
    {
      syncStockInstrumentMaster: async () => ({ count: 12, source_key: "TUSHARE", result: { truth_count: 12 } }),
      syncStockQuotes: async () => ({ count: 12, source_key: "TUSHARE", result: { truth_count: 8 } }),
      syncFuturesQuotes: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncFuturesInventory: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncMarketNewsSource: async () => ({ count: 0, source_key: "AKSHARE", result: null })
    }
  );

  await state.handleSyncStockQuotes();

  assert.equal(state.stockSyncOptions.value.length, 1);
  assert.equal(state.stockLastSyncResult.value.count, 12);
  assert.equal(state.syncCards.value[0].key, "stock_quotes");
  assert.equal(state.syncCards.value[0].actions.length, 2);
});

test("useMarketSyncConsole full stock sync updates master before FULL_MARKET quotes when stock symbols are empty", async () => {
  const items = ref([
    { source_key: "tushare", name: "Tushare", status: "ACTIVE", config: { provider: "tushare" } }
  ]);
  const healthMap = ref({ TUSHARE: { status: "HEALTHY" } });
  const callLog = [];
  const state = useMarketSyncConsole(
    {
      feedback: createFeedback(),
      canEditMarket: true,
      items,
      healthMap,
      defaultStockSourceKey: ref("TUSHARE"),
      defaultFuturesSourceKey: ref("TUSHARE"),
      defaultMarketNewsSourceKey: ref("AKSHARE")
    },
    {
      syncStockInstrumentMaster: async (payload) => {
        callLog.push({ type: "master", payload });
        return { count: 4200, source_key: "TUSHARE", result: { truth_count: 4200 } };
      },
      syncStockQuotes: async (payload) => {
        callLog.push({ type: "quotes", payload });
        return { count: 4200, source_key: "TUSHARE", sync_mode: payload.sync_mode, result: { truth_count: 4200 } };
      },
      syncMarketDataDailyBasic: async (payload) => {
        callLog.push({ type: "daily_basic", payload });
        return { count: 4100, source_key: "TUSHARE", result: { bar_count: 4100 } };
      },
      syncMarketDataMoneyflow: async (payload) => {
        callLog.push({ type: "moneyflow", payload });
        return { count: 4088, source_key: "TUSHARE", result: { bar_count: 4088 } };
      },
      syncFuturesQuotes: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncFuturesInventory: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncMarketNewsSource: async () => ({ count: 0, source_key: "AKSHARE", result: null })
    }
  );

  state.stockSyncForm.symbols = "";
  await state.handleSyncStockFullSync();

  state.stockSyncForm.symbols = "600519.SH,000001.SZ";
  await state.handleSyncStockQuotes();

  assert.equal(callLog[0].type, "master");
  assert.deepEqual(callLog[0].payload, { source_key: "TUSHARE", symbols: [] });
  assert.equal(callLog[1].type, "quotes");
  assert.equal(callLog[1].payload.sync_mode, "FULL_MARKET");
  assert.deepEqual(callLog[1].payload.symbols, []);
  assert.equal(callLog[2].type, "daily_basic");
  assert.deepEqual(callLog[2].payload, {
    source_key: "TUSHARE",
    symbols: [],
    asset_scope: ["STOCK"],
    days: 120
  });
  assert.equal(callLog[3].type, "moneyflow");
  assert.deepEqual(callLog[3].payload, {
    source_key: "TUSHARE",
    symbols: [],
    asset_scope: ["STOCK"],
    days: 120
  });
  assert.equal(callLog[4].type, "quotes");
  assert.equal(callLog[4].payload.sync_mode, undefined);
  assert.deepEqual(callLog[4].payload.symbols, ["600519.SH", "000001.SZ"]);
});

test("useMarketSyncConsole quotes-only stock sync skips master refresh", async () => {
  const items = ref([
    { source_key: "tushare", name: "Tushare", status: "ACTIVE", config: { provider: "tushare" } }
  ]);
  const healthMap = ref({ TUSHARE: { status: "HEALTHY" } });
  const callLog = [];
  const state = useMarketSyncConsole(
    {
      feedback: createFeedback(),
      canEditMarket: true,
      items,
      healthMap,
      defaultStockSourceKey: ref("TUSHARE"),
      defaultFuturesSourceKey: ref("TUSHARE"),
      defaultMarketNewsSourceKey: ref("AKSHARE")
    },
    {
      syncStockInstrumentMaster: async (payload) => {
        callLog.push({ type: "master", payload });
        return { count: 4200, source_key: "TUSHARE", result: { truth_count: 4200 } };
      },
      syncStockQuotes: async (payload) => {
        callLog.push({ type: "quotes", payload });
        return { count: 1395, source_key: "TUSHARE", sync_mode: payload.sync_mode, result: { truth_count: 1395 } };
      },
      syncFuturesQuotes: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncFuturesInventory: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncMarketNewsSource: async () => ({ count: 0, source_key: "AKSHARE", result: null })
    }
  );

  state.stockSyncForm.symbols = "";
  await state.handleSyncStockQuotes();

  assert.equal(callLog.length, 1);
  assert.equal(callLog[0].type, "quotes");
  assert.equal(callLog[0].payload.sync_mode, "FULL_MARKET");
});

test("useMarketSyncConsole full futures sync refreshes master before generic quotes sync", async () => {
  const items = ref([
    { source_key: "tushare", name: "Tushare", status: "ACTIVE", config: { provider: "tushare" } }
  ]);
  const healthMap = ref({ TUSHARE: { status: "HEALTHY" } });
  const callLog = [];
  const state = useMarketSyncConsole(
    {
      feedback: createFeedback(),
      canEditMarket: true,
      items,
      healthMap,
      defaultStockSourceKey: ref("TUSHARE"),
      defaultFuturesSourceKey: ref("TUSHARE"),
      defaultMarketNewsSourceKey: ref("AKSHARE")
    },
    {
      syncStockInstrumentMaster: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncStockQuotes: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncMarketDataMaster: async (payload) => {
        callLog.push({ type: "master", payload });
        return { count: 3200, source_key: "TUSHARE", result: { snapshot_count: 3200 } };
      },
      syncMarketDataQuotes: async (payload) => {
        callLog.push({ type: "quotes", payload });
        return { count: 3200, source_key: "TUSHARE", result: { truth_count: 3200 } };
      },
      syncFuturesQuotes: async (payload) => {
        callLog.push({ type: "legacy_quotes", payload });
        return { count: 88, source_key: "TUSHARE", result: { truth_count: 88 } };
      },
      syncFuturesInventory: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncMarketNewsSource: async () => ({ count: 0, source_key: "AKSHARE", result: null })
    }
  );

  state.futuresSyncForm.contracts = "";
  await state.handleSyncFuturesFullSync();

  state.futuresSyncForm.contracts = "AU2506.SHF,AG2506.SHF";
  await state.handleSyncFuturesQuotes();

  assert.equal(callLog[0].type, "master");
  assert.deepEqual(callLog[0].payload, {
    source_key: "TUSHARE",
    asset_scope: ["FUTURES"]
  });
  assert.equal(callLog[1].type, "quotes");
  assert.deepEqual(callLog[1].payload, {
    source_key: "TUSHARE",
    asset_scope: ["FUTURES"],
    days: 120
  });
  assert.equal(callLog[2].type, "legacy_quotes");
  assert.deepEqual(callLog[2].payload.contracts, ["AU2506.SHF", "AG2506.SHF"]);
});

test("useMarketSyncConsole full futures sync falls back to legacy quotes when market snapshot table is missing", async () => {
  const items = ref([
    { source_key: "tushare", name: "Tushare", status: "ACTIVE", config: { provider: "tushare" } }
  ]);
  const healthMap = ref({ TUSHARE: { status: "HEALTHY" } });
  const callLog = [];
  const state = useMarketSyncConsole(
    {
      feedback: createFeedback(),
      canEditMarket: true,
      items,
      healthMap,
      defaultStockSourceKey: ref("TUSHARE"),
      defaultFuturesSourceKey: ref("TUSHARE"),
      defaultMarketNewsSourceKey: ref("AKSHARE")
    },
    {
      syncStockInstrumentMaster: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncStockQuotes: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncMarketDataMaster: async (payload) => {
        callLog.push({ type: "master", payload });
        throw new Error("Error 1146 (42S02): Table 'sercherai.market_universe_snapshots' doesn't exist");
      },
      syncMarketDataQuotes: async (payload) => {
        callLog.push({ type: "quotes", payload });
        return { count: 0, source_key: "TUSHARE", result: { truth_count: 0 } };
      },
      syncFuturesQuotes: async (payload) => {
        callLog.push({ type: "legacy_quotes", payload });
        return { count: 1888, source_key: "TUSHARE", result: { truth_count: 1888 } };
      },
      syncFuturesInventory: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncMarketNewsSource: async () => ({ count: 0, source_key: "AKSHARE", result: null })
    }
  );

  state.futuresSyncForm.contracts = "";
  await state.handleSyncFuturesFullSync();

  assert.equal(callLog.length, 2);
  assert.equal(callLog[0].type, "master");
  assert.equal(callLog[1].type, "legacy_quotes");
  const futuresCard = state.syncCards.value.find((item) => item.key === "futures_quotes");
  const messages = futuresCard.logs.map((item) => item.message);
  assert.equal(messages.some((item) => /自动降级/.test(item)), true);
  assert.equal(state.futuresLastSyncResult.value.count, 1888);
});

test("useMarketSyncConsole records per-card execution logs for stock full sync", async () => {
  const items = ref([
    { source_key: "tushare", name: "Tushare", status: "ACTIVE", config: { provider: "tushare" } }
  ]);
  const healthMap = ref({ TUSHARE: { status: "HEALTHY" } });
  const state = useMarketSyncConsole(
    {
      feedback: createFeedback(),
      canEditMarket: true,
      items,
      healthMap,
      defaultStockSourceKey: ref("TUSHARE"),
      defaultFuturesSourceKey: ref("TUSHARE"),
      defaultMarketNewsSourceKey: ref("AKSHARE")
    },
    {
      syncStockInstrumentMaster: async () => ({ count: 4200, source_key: "TUSHARE", result: { snapshot_count: 4200 } }),
      syncStockQuotes: async () => ({ count: 1395, source_key: "TUSHARE", result: { truth_count: 1395 } }),
      syncMarketDataDailyBasic: async () => ({ count: 1360, source_key: "TUSHARE", result: { bar_count: 1360 } }),
      syncMarketDataMoneyflow: async () => ({ count: 1328, source_key: "TUSHARE", result: { bar_count: 1328 } }),
      syncFuturesQuotes: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncFuturesInventory: async () => ({ count: 0, source_key: "TUSHARE", result: null }),
      syncMarketNewsSource: async () => ({ count: 0, source_key: "AKSHARE", result: null })
    }
  );

  await state.handleSyncStockFullSync();

  const stockCard = state.syncCards.value.find((item) => item.key === "stock_quotes");
  const messages = stockCard.logs.map((item) => item.message);
  assert.equal(stockCard.logs.length >= 10, true);
  assert.match(messages[0], /开始执行全量同步/);
  assert.match(messages[1], /当前数据源：TUSHARE/);
  assert.match(messages[2], /股票代码表同步请求已发送/);
  assert.match(messages[3], /股票代码表同步完成/);
  assert.match(messages[4], /股票行情同步请求已发送/);
  assert.match(messages[5], /股票行情同步完成/);
  assert.match(messages[6], /股票日度指标同步请求已发送/);
  assert.match(messages[7], /股票日度指标同步完成/);
  assert.match(messages[8], /股票资金流向同步请求已发送/);
  assert.match(messages[9], /股票资金流向同步完成/);
  assert.match(messages.at(-1), /本次执行结束/);
});
