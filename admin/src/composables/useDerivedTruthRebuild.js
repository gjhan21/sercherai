import { reactive, ref } from "vue";
import {
  getMarketDerivedTruthSummary,
  rebuildFuturesDerivedTruth,
  rebuildStockDerivedTruth
} from "../api/admin.js";
import {
  defaultRebuildForm,
  normalizeSourceKey,
  toSafeInt
} from "../lib/data-sources-admin.js";
import { formatTruthRebuildSuccessMessage } from "../lib/market-data-admin.js";

const defaultDeps = {
  getMarketDerivedTruthSummary,
  rebuildFuturesDerivedTruth,
  rebuildStockDerivedTruth
};

function noop() {}

export function useDerivedTruthRebuild(options = {}, injectedDeps = {}) {
  const deps = { ...defaultDeps, ...injectedDeps };
  const feedback = options.feedback || { clear: noop, setMessage: noop, setError: noop };
  const canEditMarket = options.canEditMarket !== false;
  const refreshQualityDashboard = options.refreshQualityDashboard || noop;

  const rebuildingStockTruth = ref(false);
  const rebuildingFuturesTruth = ref(false);
  const stockRebuildSummary = ref(null);
  const futuresRebuildSummary = ref(null);
  const rebuildForm = reactive(defaultRebuildForm());

  function ensureCanEditMarket() {
    if (canEditMarket) {
      return true;
    }
    feedback.setError("当前账号没有行情维护权限，无法执行 truth 派生重建");
    return false;
  }

  function normalizeRebuildPayload() {
    return {
      trade_date: String(rebuildForm.trade_date || "").trim(),
      days: toSafeInt(rebuildForm.days, 3)
    };
  }

  function getRebuildSummaryRef(assetClass) {
    return assetClass === "FUTURES" ? futuresRebuildSummary : stockRebuildSummary;
  }

  async function fetchDerivedTruthSummary(assetClass, options = {}) {
    const { preserveFeedback = false, preserveCurrent = false } = options;
    const normalizedAssetClass = normalizeSourceKey(assetClass);
    const summaryRef = getRebuildSummaryRef(normalizedAssetClass);
    if (!preserveFeedback) {
      feedback.clear();
    }
    try {
      const data = await deps.getMarketDerivedTruthSummary({ asset_class: normalizedAssetClass });
      summaryRef.value = data || null;
    } catch (error) {
      if (!preserveCurrent) {
        summaryRef.value = null;
      }
      if (!preserveFeedback) {
        feedback.setError(error.message || "加载 truth 重建摘要失败");
      }
    }
  }

  async function fetchDerivedTruthSummaries(options = {}) {
    await Promise.all([
      fetchDerivedTruthSummary("STOCK", options),
      fetchDerivedTruthSummary("FUTURES", options)
    ]);
  }

  async function handleRebuildTruth(assetClass) {
    if (!ensureCanEditMarket()) {
      return;
    }
    const isStock = assetClass === "STOCK";
    const loadingRef = isStock ? rebuildingStockTruth : rebuildingFuturesTruth;
    const summaryRef = getRebuildSummaryRef(assetClass);
    loadingRef.value = true;
    feedback.clear();
    try {
      const data = isStock
        ? await deps.rebuildStockDerivedTruth(normalizeRebuildPayload())
        : await deps.rebuildFuturesDerivedTruth(normalizeRebuildPayload());
      summaryRef.value = data || null;
      feedback.setMessage(formatTruthRebuildSuccessMessage(isStock ? "STOCK" : "FUTURES", data));
      await Promise.all([
        refreshQualityDashboard({ preserveFeedback: true }),
        fetchDerivedTruthSummary(assetClass, { preserveFeedback: true, preserveCurrent: true })
      ]);
    } catch (error) {
      feedback.setError(error.message || `${isStock ? "股票" : "期货"} truth 重建失败`);
    } finally {
      loadingRef.value = false;
    }
  }

  return {
    rebuildingStockTruth,
    rebuildingFuturesTruth,
    stockRebuildSummary,
    futuresRebuildSummary,
    rebuildForm,
    fetchDerivedTruthSummary,
    fetchDerivedTruthSummaries,
    handleRebuildTruth
  };
}
