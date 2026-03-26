import { computed, reactive, ref, watch } from "vue";
import {
  syncMarketDataMaster,
  syncMarketDataQuotes,
  syncMarketDataDailyBasic,
  syncMarketDataMoneyflow,
  syncStockInstrumentMaster,
  syncFuturesInventory,
  syncFuturesQuotes,
  syncMarketNewsSource,
  syncStockQuotes
} from "../api/admin.js";
import {
  FUTURES_INVENTORY_SOURCE_FALLBACK,
  buildLastSyncResult,
  buildSyncSourceOptions,
  defaultFuturesInventorySyncForm,
  defaultFuturesSyncForm,
  defaultMarketNewsSyncForm,
  defaultStockSyncForm,
  ensurePreferredSource,
  splitSyncInput
} from "../lib/data-sources-admin.js";

const defaultDeps = {
  syncMarketDataMaster,
  syncMarketDataQuotes,
  syncMarketDataDailyBasic,
  syncMarketDataMoneyflow,
  syncStockInstrumentMaster,
  syncFuturesInventory,
  syncFuturesQuotes,
  syncMarketNewsSource,
  syncStockQuotes
};

function noop() {}

function buildSyncLogEntry(level, message) {
  return {
    id: `${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    level,
    message,
    time: new Date().toLocaleTimeString("zh-CN", { hour12: false })
  };
}

function isMissingMarketUniverseSnapshotsError(error) {
  const message = String(error?.message || "").toLowerCase();
  return (
    message.includes("market_universe_snapshots") &&
    (message.includes("doesn't exist") || message.includes("does not exist") || message.includes("error 1146"))
  );
}

export function useMarketSyncConsole(options = {}, injectedDeps = {}) {
  const deps = { ...defaultDeps, ...injectedDeps };
  const feedback = options.feedback || { clear: noop, setMessage: noop, setError: noop };
  const canEditMarket = options.canEditMarket !== false;
  const items = options.items;
  const healthMap = options.healthMap;
  const defaultStockSourceKey = options.defaultStockSourceKey;
  const defaultFuturesSourceKey = options.defaultFuturesSourceKey;
  const defaultMarketNewsSourceKey = options.defaultMarketNewsSourceKey;

  const stockSyncForm = reactive(defaultStockSyncForm());
  const futuresSyncForm = reactive(defaultFuturesSyncForm());
  const futuresInventorySyncForm = reactive(defaultFuturesInventorySyncForm());
  const marketNewsSyncForm = reactive(defaultMarketNewsSyncForm());

  const syncingStockQuotes = ref(false);
  const syncingFuturesQuotes = ref(false);
  const syncingFuturesInventory = ref(false);
  const syncingMarketNews = ref(false);
  const stockLastSyncResult = ref(null);
  const futuresLastSyncResult = ref(null);
  const futuresInventoryLastSyncResult = ref(null);
  const marketNewsLastSyncResult = ref(null);
  const stockSyncLogs = ref([]);
  const futuresSyncLogs = ref([]);
  const futuresInventorySyncLogs = ref([]);
  const marketNewsSyncLogs = ref([]);

  const stockSyncOptions = computed(() => buildSyncSourceOptions(items?.value || [], healthMap?.value || {}, "stock_quotes"));
  const futuresSyncOptions = computed(() => buildSyncSourceOptions(items?.value || [], healthMap?.value || {}, "futures_quotes"));
  const futuresInventorySyncOptions = computed(() => buildSyncSourceOptions(items?.value || [], healthMap?.value || {}, "futures_inventory"));
  const marketNewsSyncOptions = computed(() => buildSyncSourceOptions(items?.value || [], healthMap?.value || {}, "market_news"));

  function resetCardLogs(logsRef, actionLabel, sourceKey) {
    logsRef.value = [];
    appendCardLog(logsRef, "info", `开始执行${actionLabel}`);
    if (sourceKey) {
      appendCardLog(logsRef, "info", `当前数据源：${String(sourceKey).trim().toUpperCase()}`);
    }
  }

  function appendCardLog(logsRef, level, message) {
    logsRef.value = [...logsRef.value, buildSyncLogEntry(level, message)];
  }

  watch(
    [items, healthMap, defaultStockSourceKey, defaultFuturesSourceKey, defaultMarketNewsSourceKey],
    () => {
      ensurePreferredSource(stockSyncForm, stockSyncOptions.value, defaultStockSourceKey?.value);
      ensurePreferredSource(futuresSyncForm, futuresSyncOptions.value, defaultFuturesSourceKey?.value);
      ensurePreferredSource(
        futuresInventorySyncForm,
        futuresInventorySyncOptions.value,
        FUTURES_INVENTORY_SOURCE_FALLBACK
      );
      ensurePreferredSource(marketNewsSyncForm, marketNewsSyncOptions.value, defaultMarketNewsSourceKey?.value);
    },
    { immediate: true, deep: true }
  );

  function ensureCanEditMarket() {
    if (canEditMarket) {
      return true;
    }
    feedback.setError("当前账号没有行情维护权限，无法执行 truth 派生重建");
    return false;
  }

  async function executeStockSync({ refreshMaster }) {
    if (!ensureCanEditMarket()) {
      return;
    }
    syncingStockQuotes.value = true;
    feedback.clear();
    const sourceKey = String(stockSyncForm.source_key || "").trim();
    resetCardLogs(stockSyncLogs, refreshMaster ? "全量同步" : "仅行情同步", sourceKey);
    try {
      const symbols = splitSyncInput(stockSyncForm.symbols);
      const syncDays = Number(stockSyncForm.days) || 120;
      if (refreshMaster) {
        appendCardLog(stockSyncLogs, "info", "股票代码表同步请求已发送");
        const masterData = await deps.syncStockInstrumentMaster({
          source_key: sourceKey,
          symbols
        });
        const masterResult = buildLastSyncResult(masterData);
        appendCardLog(stockSyncLogs, "success", `股票代码表同步完成，处理 ${masterResult.count || 0} 条`);
      }
      appendCardLog(
        stockSyncLogs,
        "info",
        symbols.length === 0 ? "股票行情同步请求已发送（范围：全市场）" : "股票行情同步请求已发送"
      );
      const data = await deps.syncStockQuotes({
        source_key: sourceKey,
        symbols,
        sync_mode: symbols.length === 0 ? "FULL_MARKET" : undefined,
        days: syncDays
      });
      stockLastSyncResult.value = buildLastSyncResult(data);
      appendCardLog(stockSyncLogs, "success", `股票行情同步完成，处理 ${stockLastSyncResult.value.count || 0} 条`);
      let dailyBasicCount = 0;
      let moneyflowCount = 0;
      if (refreshMaster) {
        const enhancementPayload = {
          source_key: sourceKey,
          symbols,
          asset_scope: ["STOCK"],
          days: syncDays
        };
        appendCardLog(stockSyncLogs, "info", "股票日度指标同步请求已发送");
        const dailyBasicData = await deps.syncMarketDataDailyBasic(enhancementPayload);
        const dailyBasicResult = buildLastSyncResult(dailyBasicData);
        dailyBasicCount = dailyBasicResult.count || 0;
        appendCardLog(stockSyncLogs, "success", `股票日度指标同步完成，处理 ${dailyBasicCount} 条`);
        appendCardLog(stockSyncLogs, "info", "股票资金流向同步请求已发送");
        const moneyflowData = await deps.syncMarketDataMoneyflow(enhancementPayload);
        const moneyflowResult = buildLastSyncResult(moneyflowData);
        moneyflowCount = moneyflowResult.count || 0;
        appendCardLog(stockSyncLogs, "success", `股票资金流向同步完成，处理 ${moneyflowCount} 条`);
      }
      feedback.setMessage(
        refreshMaster && symbols.length === 0
          ? `股票全量同步完成（全市场）：行情 ${stockLastSyncResult.value.count || 0} 条，日度指标 ${dailyBasicCount} 条，资金流向 ${moneyflowCount} 条`
          : refreshMaster
            ? `股票全量同步完成：行情 ${stockLastSyncResult.value.count || 0} 条，日度指标 ${dailyBasicCount} 条，资金流向 ${moneyflowCount} 条`
          : `股票行情同步完成，处理 ${stockLastSyncResult.value.count || 0} 条`
      );
    } catch (error) {
      appendCardLog(stockSyncLogs, "danger", error.message || "股票行情同步失败");
      feedback.setError(error.message || "股票行情同步失败");
    } finally {
      appendCardLog(stockSyncLogs, "info", "本次执行结束");
      syncingStockQuotes.value = false;
    }
  }

  async function handleSyncStockFullSync() {
    return executeStockSync({ refreshMaster: true });
  }

  async function handleSyncStockQuotes() {
    return executeStockSync({ refreshMaster: false });
  }

  async function handleSyncFuturesQuotes() {
    if (!ensureCanEditMarket()) {
      return;
    }
    syncingFuturesQuotes.value = true;
    feedback.clear();
    const sourceKey = String(futuresSyncForm.source_key || "").trim();
    resetCardLogs(futuresSyncLogs, "仅行情同步", sourceKey);
    try {
      const contracts = splitSyncInput(futuresSyncForm.contracts);
      appendCardLog(
        futuresSyncLogs,
        "info",
        contracts.length === 0 ? "期货行情同步请求已发送（范围：默认合约池）" : "期货行情同步请求已发送"
      );
      const data = await deps.syncFuturesQuotes({
        source_key: sourceKey,
        contracts,
        days: Number(futuresSyncForm.days) || 120
      });
      futuresLastSyncResult.value = buildLastSyncResult(data);
      appendCardLog(futuresSyncLogs, "success", `期货行情同步完成，处理 ${futuresLastSyncResult.value.count || 0} 条`);
      feedback.setMessage(`期货行情同步完成，处理 ${futuresLastSyncResult.value.count || 0} 条`);
    } catch (error) {
      appendCardLog(futuresSyncLogs, "danger", error.message || "期货行情同步失败");
      feedback.setError(error.message || "期货行情同步失败");
    } finally {
      appendCardLog(futuresSyncLogs, "info", "本次执行结束");
      syncingFuturesQuotes.value = false;
    }
  }

  async function handleSyncFuturesFullSync() {
    if (!ensureCanEditMarket()) {
      return;
    }
    syncingFuturesQuotes.value = true;
    feedback.clear();
    const sourceKey = String(futuresSyncForm.source_key || "").trim();
    resetCardLogs(futuresSyncLogs, "全量同步", sourceKey);
    try {
      const contracts = splitSyncInput(futuresSyncForm.contracts);
      const syncDays = Number(futuresSyncForm.days) || 120;
      const payload = {
        source_key: sourceKey,
        asset_scope: ["FUTURES"]
      };
      if (contracts.length > 0) {
        payload.symbols = contracts;
      }
      appendCardLog(futuresSyncLogs, "info", "期货代码表同步请求已发送");
      try {
        await deps.syncMarketDataMaster(payload);
        appendCardLog(futuresSyncLogs, "success", "期货代码表同步完成");
        appendCardLog(
          futuresSyncLogs,
          "info",
          contracts.length === 0 ? "期货行情同步请求已发送（范围：全合约池）" : "期货行情同步请求已发送"
        );
        const data = await deps.syncMarketDataQuotes({
          ...payload,
          days: syncDays
        });
        futuresLastSyncResult.value = buildLastSyncResult(data);
        futuresLastSyncResult.value.contracts = contracts;
        appendCardLog(futuresSyncLogs, "success", `期货行情同步完成，处理 ${futuresLastSyncResult.value.count || 0} 条`);
        feedback.setMessage(
          contracts.length === 0
            ? `期货主数据与全合约行情同步完成，处理 ${futuresLastSyncResult.value.count || 0} 条`
            : `期货主数据与行情同步完成，处理 ${futuresLastSyncResult.value.count || 0} 条`
        );
      } catch (marketSyncError) {
        if (!isMissingMarketUniverseSnapshotsError(marketSyncError)) {
          throw marketSyncError;
        }
        appendCardLog(
          futuresSyncLogs,
          "warning",
          "检测到未初始化 market_universe_snapshots，已自动降级为期货全池行情同步"
        );
        appendCardLog(
          futuresSyncLogs,
          "info",
          contracts.length === 0 ? "期货行情同步请求已发送（降级路径：全合约池）" : "期货行情同步请求已发送（降级路径）"
        );
        const data = await deps.syncFuturesQuotes({
          source_key: sourceKey,
          contracts,
          days: syncDays
        });
        futuresLastSyncResult.value = buildLastSyncResult(data);
        futuresLastSyncResult.value.contracts = contracts;
        appendCardLog(futuresSyncLogs, "success", `期货行情同步完成，处理 ${futuresLastSyncResult.value.count || 0} 条`);
        feedback.setMessage(
          contracts.length === 0
            ? `期货全量同步已降级为全合约行情同步，处理 ${futuresLastSyncResult.value.count || 0} 条`
            : `期货全量同步已降级为行情同步，处理 ${futuresLastSyncResult.value.count || 0} 条`
        );
      }
    } catch (error) {
      appendCardLog(futuresSyncLogs, "danger", error.message || "期货全量同步失败");
      feedback.setError(error.message || "期货全量同步失败");
    } finally {
      appendCardLog(futuresSyncLogs, "info", "本次执行结束");
      syncingFuturesQuotes.value = false;
    }
  }

  async function handleSyncFuturesInventory() {
    if (!ensureCanEditMarket()) {
      return;
    }
    syncingFuturesInventory.value = true;
    feedback.clear();
    const sourceKey = String(futuresInventorySyncForm.source_key || "").trim();
    resetCardLogs(futuresInventorySyncLogs, "仓单同步", sourceKey);
    try {
      appendCardLog(futuresInventorySyncLogs, "info", "期货仓单同步请求已发送");
      const data = await deps.syncFuturesInventory({
        source_key: sourceKey,
        symbols: splitSyncInput(futuresInventorySyncForm.symbols),
        days: Number(futuresInventorySyncForm.days) || 30
      });
      futuresInventoryLastSyncResult.value = buildLastSyncResult(data);
      appendCardLog(futuresInventorySyncLogs, "success", `期货仓单同步完成，处理 ${futuresInventoryLastSyncResult.value.count || 0} 条`);
      feedback.setMessage(`期货仓单同步完成，处理 ${futuresInventoryLastSyncResult.value.count || 0} 条`);
    } catch (error) {
      appendCardLog(futuresInventorySyncLogs, "danger", error.message || "期货仓单同步失败");
      feedback.setError(error.message || "期货仓单同步失败");
    } finally {
      appendCardLog(futuresInventorySyncLogs, "info", "本次执行结束");
      syncingFuturesInventory.value = false;
    }
  }

  async function handleSyncMarketNews() {
    if (!ensureCanEditMarket()) {
      return;
    }
    syncingMarketNews.value = true;
    feedback.clear();
    const sourceKey = String(marketNewsSyncForm.source_key || "").trim();
    resetCardLogs(marketNewsSyncLogs, "资讯同步", sourceKey);
    try {
      appendCardLog(marketNewsSyncLogs, "info", "市场资讯同步请求已发送");
      const data = await deps.syncMarketNewsSource({
        source_key: sourceKey,
        symbols: splitSyncInput(marketNewsSyncForm.symbols),
        days: Number(marketNewsSyncForm.days) || 3,
        limit: Number(marketNewsSyncForm.limit) || 50
      });
      marketNewsLastSyncResult.value = buildLastSyncResult(data);
      appendCardLog(marketNewsSyncLogs, "success", `市场资讯同步完成，处理 ${marketNewsLastSyncResult.value.count || 0} 条`);
      feedback.setMessage(`市场资讯同步完成，处理 ${marketNewsLastSyncResult.value.count || 0} 条`);
    } catch (error) {
      appendCardLog(marketNewsSyncLogs, "danger", error.message || "市场资讯同步失败");
      feedback.setError(error.message || "市场资讯同步失败");
    } finally {
      appendCardLog(marketNewsSyncLogs, "info", "本次执行结束");
      syncingMarketNews.value = false;
    }
  }

  const syncCards = computed(() => [
    {
      key: "stock_quotes",
      title: "股票同步",
      options: stockSyncOptions.value,
      form: stockSyncForm,
      loading: syncingStockQuotes.value,
      result: stockLastSyncResult.value,
      logs: stockSyncLogs.value,
      inputKey: "symbols",
      scopeKey: "symbols",
      emptyScopeLabel: "全市场",
      placeholder: "股票代码，逗号或换行分隔；留空按当前主数据全市场同步",
      hint: "建议优先点“全量同步”：先刷新股票代码表，再同步行情；“仅行情”会直接基于当前主数据范围跑 `FULL_MARKET` 或指定股票。",
      minDays: 20,
      maxDays: 365,
      actions: [
        { key: "full", label: "全量同步", type: "primary", run: handleSyncStockFullSync },
        { key: "quotes", label: "仅行情", type: "default", run: handleSyncStockQuotes }
      ]
    },
    {
      key: "futures_quotes",
      title: "期货行情同步",
      options: futuresSyncOptions.value,
      form: futuresSyncForm,
      loading: syncingFuturesQuotes.value,
      result: futuresLastSyncResult.value,
      logs: futuresSyncLogs.value,
      inputKey: "contracts",
      scopeKey: "contracts",
      emptyScopeLabel: "全合约池",
      placeholder: "合约代码，逗号或换行分隔；留空先刷新期货代码表，再按全合约同步",
      hint: "建议优先点“全量同步”：先刷新期货代码表，再按代码表同步行情；“仅行情”会直接基于指定合约或当前默认合约池拉取。",
      minDays: 20,
      maxDays: 365,
      actions: [
        { key: "full", label: "全量同步", type: "primary", run: handleSyncFuturesFullSync },
        { key: "quotes", label: "仅行情", type: "default", run: handleSyncFuturesQuotes }
      ]
    },
    {
      key: "futures_inventory",
      title: "期货仓单同步",
      options: futuresInventorySyncOptions.value,
      form: futuresInventorySyncForm,
      loading: syncingFuturesInventory.value,
      result: futuresInventoryLastSyncResult.value,
      logs: futuresInventorySyncLogs.value,
      inputKey: "symbols",
      scopeKey: "symbols",
      emptyScopeLabel: "默认品种池",
      placeholder: "品种代码，逗号或换行分隔；留空使用默认品种池",
      hint: "当前最小闭环先支持 `TUSHARE / MOCK / AUTO`；品种代码会自动归一到根符号，例如 `RB2405 -> RB`。",
      minDays: 1,
      maxDays: 90,
      actions: [{ key: "inventory", label: "同步期货仓单", type: "primary", run: handleSyncFuturesInventory }]
    },
    {
      key: "market_news",
      title: "市场资讯同步",
      options: marketNewsSyncOptions.value,
      form: marketNewsSyncForm,
      loading: syncingMarketNews.value,
      result: marketNewsLastSyncResult.value,
      logs: marketNewsSyncLogs.value,
      inputKey: "symbols",
      scopeKey: "symbols",
      emptyScopeLabel: "不限股票代码",
      placeholder: "关联股票代码，逗号或换行分隔；留空表示不限",
      hint: "支持输入 `AUTO` 或自定义链路，如 `AKSHARE,TUSHARE`。",
      minDays: 1,
      maxDays: 30,
      hasLimit: true,
      minLimit: 1,
      maxLimit: 500,
      actions: [{ key: "news", label: "同步市场资讯", type: "primary", run: handleSyncMarketNews }]
    }
  ]);

  return {
    stockSyncForm,
    futuresSyncForm,
    futuresInventorySyncForm,
    marketNewsSyncForm,
    syncingStockQuotes,
    syncingFuturesQuotes,
    syncingFuturesInventory,
    syncingMarketNews,
    stockLastSyncResult,
    futuresLastSyncResult,
    futuresInventoryLastSyncResult,
    marketNewsLastSyncResult,
    stockSyncOptions,
    futuresSyncOptions,
    futuresInventorySyncOptions,
    marketNewsSyncOptions,
    syncCards,
    handleSyncStockFullSync,
    handleSyncStockQuotes,
    handleSyncFuturesFullSync,
    handleSyncFuturesQuotes,
    handleSyncFuturesInventory,
    handleSyncMarketNews
  };
}
