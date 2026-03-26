<script setup>
import MarketCoveragePanel from "../../components/data-sources/MarketCoveragePanel.vue";
import { useDataSourcesWorkspaceContext } from "../../composables/useDataSourcesWorkspace.js";

const workspace = useDataSourcesWorkspaceContext();
</script>

<template>
  <MarketCoveragePanel
    :loading="workspace.marketCoverageLoading"
    :overview-cards="workspace.marketCoverageOverviewCards"
    :summary="workspace.marketCoverageSummary"
    :asset-rows="workspace.marketCoverageAssetRows"
    :snapshots="workspace.marketUniverseSnapshots"
    :snapshots-loading="workspace.marketUniverseSnapshotsLoading"
    :backfill-runs="workspace.marketBackfillRuns"
    :backfill-runs-loading="workspace.marketBackfillRunsLoading"
    @refresh="workspace.fetchMarketCoverageGovernance()"
    @open-jobs="workspace.openMarketBackfillWorkspace"
  />
</template>
