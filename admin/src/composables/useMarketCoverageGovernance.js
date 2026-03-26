import { computed, ref } from "vue";
import {
  getMarketCoverageSummary,
  listMarketDataBackfillRuns,
  listMarketUniverseSnapshots
} from "../api/admin.js";
import { buildMarketCoverageAssetRows, buildMarketCoverageOverviewCards } from "../lib/market-data-admin.js";

const defaultDeps = {
  getMarketCoverageSummary,
  listMarketDataBackfillRuns,
  listMarketUniverseSnapshots
};

function noop() {}

export function useMarketCoverageGovernance(options = {}, injectedDeps = {}) {
  const deps = { ...defaultDeps, ...injectedDeps };
  const feedback = options.feedback || { clear: noop, setMessage: noop, setError: noop };
  const router = options.router;

  const marketCoverageLoading = ref(false);
  const marketCoverageSummary = ref(null);
  const marketCoverageAssetRows = ref([]);
  const marketUniverseSnapshotsLoading = ref(false);
  const marketUniverseSnapshots = ref([]);
  const marketBackfillRunsLoading = ref(false);
  const marketBackfillRuns = ref([]);

  const marketCoverageOverviewCards = computed(() =>
    buildMarketCoverageOverviewCards(marketCoverageSummary.value || {})
  );

  async function fetchMarketCoverageGovernance(options = {}) {
    const { preserveFeedback = false } = options;
    if (!preserveFeedback) {
      feedback.clear();
    }
    marketCoverageLoading.value = true;
    marketUniverseSnapshotsLoading.value = true;
    marketBackfillRunsLoading.value = true;
    try {
      const [coverageData, snapshotData, runData] = await Promise.all([
        deps.getMarketCoverageSummary(),
        deps.listMarketUniverseSnapshots({ page: 1, page_size: 5 }),
        deps.listMarketDataBackfillRuns({ page: 1, page_size: 5 })
      ]);
      marketCoverageSummary.value = coverageData || null;
      marketCoverageAssetRows.value = buildMarketCoverageAssetRows(coverageData || {});
      marketUniverseSnapshots.value = snapshotData?.items || [];
      marketBackfillRuns.value = runData?.items || [];
    } catch (error) {
      marketCoverageSummary.value = null;
      marketCoverageAssetRows.value = [];
      marketUniverseSnapshots.value = [];
      marketBackfillRuns.value = [];
      if (!preserveFeedback) {
        feedback.setError(error.message || "加载全市场治理概览失败");
      }
    } finally {
      marketCoverageLoading.value = false;
      marketUniverseSnapshotsLoading.value = false;
      marketBackfillRunsLoading.value = false;
    }
  }

  function openMarketBackfillWorkspace() {
    router?.push({
      path: "/system-jobs",
      query: { tab: "market-data" }
    });
  }

  return {
    marketCoverageLoading,
    marketCoverageSummary,
    marketCoverageAssetRows,
    marketUniverseSnapshotsLoading,
    marketUniverseSnapshots,
    marketBackfillRunsLoading,
    marketBackfillRuns,
    marketCoverageOverviewCards,
    fetchMarketCoverageGovernance,
    openMarketBackfillWorkspace
  };
}
