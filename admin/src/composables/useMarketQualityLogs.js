import { computed, reactive, ref, watch } from "vue";
import {
  getMarketDataQualitySummary,
  listMarketDataQualityLogs
} from "../api/admin.js";
import {
  DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS,
  areMarketQualityFiltersEqual,
  buildMarketQualityBucketSummary,
  buildMarketQualityRouteQuery,
  buildStockGovernanceSummaryItems,
  collectMarketQualityIssueOptions,
  formatMarketQualityLookbackLabel,
  formatMarketQualityPayload,
  formatStockFallbackSourceSummary,
  normalizeMarketQualityLookbackHours,
  normalizeMarketQualityRouteContext,
  normalizeMarketQualityRouteFilters
} from "../lib/market-data-admin.js";
import {
  assetClassOptions,
  formatQualitySummaryLatest,
  formatQualitySummaryLatestError,
  qualityDataKindOptions,
  severityFilterOptions
} from "../lib/data-sources-admin.js";

const defaultDeps = {
  getMarketDataQualitySummary,
  listMarketDataQualityLogs
};

function noop() {}

export function useMarketQualityLogs(options = {}, injectedDeps = {}) {
  const deps = { ...defaultDeps, ...injectedDeps };
  const feedback = options.feedback || { clear: noop, setMessage: noop, setError: noop };
  const route = options.route;
  const router = options.router;

  const qualityLogsLoading = ref(false);
  const qualityLogItems = ref([]);
  const qualityLogTotal = ref(0);
  const qualityRouteInitialized = ref(false);
  const qualityPayloadDialogVisible = ref(false);
  const selectedQualityLog = ref(null);
  const qualitySummaryAll = ref(null);
  const qualitySummaryStock = ref(null);
  const qualitySummaryFutures = ref(null);

  const qualityFilters = reactive({
    asset_class: "",
    data_kind: "",
    severity: "",
    issue_code: "",
    hours: DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS
  });

  const qualityIssueQuickOptions = computed(() =>
    collectMarketQualityIssueOptions(qualityLogItems.value, 8)
  );
  const qualitySourceBuckets = computed(() =>
    buildMarketQualityBucketSummary(qualityLogItems.value, (item) => item?.source_key, 3)
  );
  const qualityIssueBuckets = computed(() =>
    buildMarketQualityBucketSummary(qualityLogItems.value, (item) => item?.issue_code, 3)
  );
  const qualitySeverityBuckets = computed(() =>
    buildMarketQualityBucketSummary(qualityLogItems.value, (item) => item?.severity, 3)
  );
  const qualityLookbackLabel = computed(() => formatMarketQualityLookbackLabel(qualityFilters.hours));
  const stockGovernanceSummaryItems = computed(() =>
    buildStockGovernanceSummaryItems(qualitySummaryStock.value || {})
  );
  const selectedQualityPayloadText = computed(() =>
    formatMarketQualityPayload(selectedQualityLog.value?.payload || "")
  );

  function resolveQualityLookbackHours(value) {
    return normalizeMarketQualityLookbackHours(value);
  }

  function getQualitySummaryRef(assetClass) {
    if (assetClass === "STOCK") return qualitySummaryStock;
    if (assetClass === "FUTURES") return qualitySummaryFutures;
    return qualitySummaryAll;
  }

  async function fetchQualityLogs(options = {}) {
    const { preserveFeedback = false } = options;
    qualityLogsLoading.value = true;
    if (!preserveFeedback) {
      feedback.clear();
    }
    try {
      const data = await deps.listMarketDataQualityLogs({
        ...qualityFilters,
        hours: resolveQualityLookbackHours(qualityFilters.hours),
        page: 1,
        page_size: 20
      });
      qualityLogItems.value = data?.items || [];
      qualityLogTotal.value = data?.total || 0;
    } catch (error) {
      qualityLogItems.value = [];
      qualityLogTotal.value = 0;
      feedback.setError(error.message || "加载质量日志失败");
    } finally {
      qualityLogsLoading.value = false;
    }
  }

  async function fetchQualitySummary(assetClass, options = {}) {
    const { preserveFeedback = false } = options;
    const normalizedAssetClass = String(assetClass || "").trim().toUpperCase();
    const summaryRef = getQualitySummaryRef(normalizedAssetClass);
    if (!preserveFeedback) {
      feedback.clear();
    }
    try {
      const params = { hours: resolveQualityLookbackHours(qualityFilters.hours) };
      if (normalizedAssetClass) {
        params.asset_class = normalizedAssetClass;
      }
      const data = await deps.getMarketDataQualitySummary(params);
      summaryRef.value = data || null;
    } catch (error) {
      summaryRef.value = null;
      if (!preserveFeedback) {
        feedback.setError(error.message || "加载质量概览失败");
      }
    }
  }

  async function fetchQualitySummaries(options = {}) {
    await Promise.all([
      fetchQualitySummary("", options),
      fetchQualitySummary("STOCK", options),
      fetchQualitySummary("FUTURES", options)
    ]);
  }

  async function fetchQualityDashboard(options = {}) {
    await Promise.all([
      fetchQualityLogs(options),
      fetchQualitySummaries(options)
    ]);
  }

  function syncQualityFiltersToRoute() {
    const nextQuery = {
      ...normalizeMarketQualityRouteContext(route?.query || {}),
      ...buildMarketQualityRouteQuery(qualityFilters)
    };
    if (areMarketQualityFiltersEqual(route?.query || {}, nextQuery)) {
      return false;
    }
    router?.replace({
      name: "data-sources-quality",
      query: nextQuery
    });
    return true;
  }

  function applyQualityIssueQuickFilter(issueCode) {
    qualityFilters.issue_code = String(issueCode || "").trim();
    if (!syncQualityFiltersToRoute()) {
      fetchQualityDashboard({ preserveFeedback: true });
    }
  }

  function openQualityPayload(row) {
    selectedQualityLog.value = row || null;
    qualityPayloadDialogVisible.value = true;
  }

  function closeQualityPayload() {
    qualityPayloadDialogVisible.value = false;
  }

  function resetQualityFilters() {
    Object.assign(qualityFilters, {
      asset_class: "",
      data_kind: "",
      severity: "",
      issue_code: "",
      hours: DEFAULT_MARKET_QUALITY_LOOKBACK_HOURS
    });
    if (!syncQualityFiltersToRoute()) {
      fetchQualityDashboard({ preserveFeedback: true });
    }
  }

  watch(
    () => route?.query,
    (query) => {
      const nextFilters = normalizeMarketQualityRouteFilters(query || {});
      if (areMarketQualityFiltersEqual(nextFilters, qualityFilters) && qualityRouteInitialized.value) {
        return;
      }
      Object.assign(qualityFilters, nextFilters);
      qualityRouteInitialized.value = true;
      fetchQualityDashboard({ preserveFeedback: true });
    },
    { immediate: true }
  );

  return {
    assetClassOptions,
    qualityDataKindOptions,
    severityFilterOptions,
    qualityFilters,
    qualityLogsLoading,
    qualityLogItems,
    qualityLogTotal,
    qualityPayloadDialogVisible,
    selectedQualityLog,
    qualitySummaryAll,
    qualitySummaryStock,
    qualitySummaryFutures,
    qualityIssueQuickOptions,
    qualitySourceBuckets,
    qualityIssueBuckets,
    qualitySeverityBuckets,
    qualityLookbackLabel,
    stockGovernanceSummaryItems,
    selectedQualityPayloadText,
    fetchQualityDashboard,
    syncQualityFiltersToRoute,
    applyQualityIssueQuickFilter,
    openQualityPayload,
    closeQualityPayload,
    resetQualityFilters,
    formatQualitySummaryLatest,
    formatQualitySummaryLatestError,
    formatStockFallbackSourceSummary
  };
}
