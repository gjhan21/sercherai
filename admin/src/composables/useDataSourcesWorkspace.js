import { computed, inject, onMounted, proxyRefs, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { useDataSourceHealth } from "./useDataSourceHealth.js";
import { useDataSourceRegistry } from "./useDataSourceRegistry.js";
import { useDerivedTruthRebuild } from "./useDerivedTruthRebuild.js";
import { useMarketCoverageGovernance } from "./useMarketCoverageGovernance.js";
import { useMarketQualityLogs } from "./useMarketQualityLogs.js";
import { useMarketSyncConsole } from "./useMarketSyncConsole.js";
import { useProviderGovernance } from "./useProviderGovernance.js";
import {
  buildDataSourcesSectionRouteName,
  dataSourcesSectionItems,
  normalizeSourceKey
} from "../lib/data-sources-admin.js";
import { buildMarketQualityRouteQuery } from "../lib/market-data-admin.js";
import { hasPermission } from "../lib/session.js";

export const DATA_SOURCES_WORKSPACE_KEY = Symbol("data-sources-workspace");

export function useDataSourcesWorkspace() {
  const route = useRoute();
  const router = useRouter();

  const message = ref("");
  const errorMessage = ref("");
  const lastRegistryLogKey = ref("");

  const feedback = {
    clear() {
      message.value = "";
      errorMessage.value = "";
    },
    setMessage(value) {
      message.value = value;
      errorMessage.value = "";
    },
    setError(value) {
      errorMessage.value = value;
    }
  };

  const canEditDataSources = hasPermission("data_source.edit");
  const canEditMarket = hasPermission("market.edit");

  const registry = useDataSourceRegistry({ feedback, canEditDataSources });
  const health = useDataSourceHealth({ feedback, canEditDataSources });
  const quality = useMarketQualityLogs({ feedback, route, router });
  const governance = useProviderGovernance({
    feedback,
    canEditDataSources,
    getLookbackHours: () => Number(quality.qualityFilters.hours) || 24
  });
  const sync = useMarketSyncConsole({
    feedback,
    canEditMarket,
    items: registry.items,
    healthMap: health.healthMap,
    defaultStockSourceKey: registry.defaultStockSourceKey,
    defaultFuturesSourceKey: registry.defaultFuturesSourceKey,
    defaultMarketNewsSourceKey: registry.defaultMarketNewsSourceKey
  });
  const coverage = useMarketCoverageGovernance({ feedback, router });
  const truth = useDerivedTruthRebuild({
    feedback,
    canEditMarket,
    refreshQualityDashboard: (options) => quality.fetchQualityDashboard(options)
  });

  const activeSectionKey = computed(() => {
    const routeName = String(route.name || "");
    const matched = dataSourcesSectionItems.find(
      (item) => buildDataSourcesSectionRouteName(item.key) === routeName
    );
    return matched?.key || "governance";
  });

  function buildSectionLocation(sectionKey) {
    const name = buildDataSourcesSectionRouteName(sectionKey);
    if (sectionKey === "quality") {
      return {
        name,
        query: buildMarketQualityRouteQuery(quality.qualityFilters)
      };
    }
    if (sectionKey === "registry") {
      const query = {};
      if (route.query?.source_key) {
        query.source_key = route.query.source_key;
      }
      if (route.query?.action) {
        query.action = route.query.action;
      }
      return Object.keys(query).length ? { name, query } : { name };
    }
    return { name };
  }

  function handleQualitySearch() {
    if (!quality.syncQualityFiltersToRoute()) {
      quality.fetchQualityDashboard();
    }
  }

  function handleRoutingPolicyDialogVisible(nextValue) {
    governance.routingPolicyDialogVisible.value = nextValue;
    if (!nextValue) {
      governance.closeRoutingPolicyDialog();
    }
  }

  function handleQualityPayloadDialogVisible(nextValue) {
    quality.qualityPayloadDialogVisible.value = nextValue;
    if (!nextValue) {
      quality.closeQualityPayload();
    }
  }

  onMounted(registry.fetchDataSources);
  onMounted(() => {
    governance.fetchGovernanceDashboard({ preserveFeedback: true });
    truth.fetchDerivedTruthSummaries({ preserveFeedback: true });
    coverage.fetchMarketCoverageGovernance({ preserveFeedback: true });
  });

  watch(
    () => quality.qualityFilters.hours,
    (nextValue, previousValue) => {
      if (previousValue !== undefined && nextValue !== previousValue) {
        governance.fetchGovernanceDashboard({ preserveFeedback: true });
      }
    }
  );

  watch(
    () => [route.name, route.query?.action, route.query?.source_key],
    ([routeName, actionValue, sourceKeyValue]) => {
      if (String(routeName || "") !== "data-sources-registry") {
        return;
      }
      const action = String(actionValue || "")
        .trim()
        .toLowerCase();
      const sourceKey = normalizeSourceKey(sourceKeyValue);
      if (action !== "logs" || !sourceKey) {
        lastRegistryLogKey.value = "";
        return;
      }
      const nextKey = `${action}:${sourceKey}`;
      if (lastRegistryLogKey.value === nextKey && health.logSourceKey.value === sourceKey) {
        return;
      }
      lastRegistryLogKey.value = nextKey;
      health.showLogs(sourceKey);
    },
    { immediate: true }
  );

  return proxyRefs({
    route,
    router,
    message,
    errorMessage,
    canEditDataSources,
    canEditMarket,
    sectionItems: dataSourcesSectionItems,
    activeSectionKey,
    buildSectionLocation,
    handleQualitySearch,
    handleRoutingPolicyDialogVisible,
    handleQualityPayloadDialogVisible,
    ...registry,
    ...health,
    ...quality,
    ...governance,
    ...sync,
    ...coverage,
    ...truth
  });
}

export function useDataSourcesWorkspaceContext() {
  const workspace = inject(DATA_SOURCES_WORKSPACE_KEY, null);
  if (!workspace) {
    throw new Error("Data sources workspace is not available");
  }
  return workspace;
}
