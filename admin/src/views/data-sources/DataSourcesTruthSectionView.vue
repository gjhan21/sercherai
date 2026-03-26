<script setup>
import DerivedTruthPanel from "../../components/data-sources/DerivedTruthPanel.vue";
import { useDataSourcesWorkspaceContext } from "../../composables/useDataSourcesWorkspace.js";

const workspace = useDataSourcesWorkspaceContext();
</script>

<template>
  <DerivedTruthPanel
    :can-edit-market="workspace.canEditMarket"
    :rebuild-form="workspace.rebuildForm"
    :rebuilding-stock-truth="workspace.rebuildingStockTruth"
    :rebuilding-futures-truth="workspace.rebuildingFuturesTruth"
    :stock-rebuild-summary="workspace.stockRebuildSummary"
    :futures-rebuild-summary="workspace.futuresRebuildSummary"
    @rebuild-stock="workspace.handleRebuildTruth('STOCK')"
    @rebuild-futures="workspace.handleRebuildTruth('FUTURES')"
  />
</template>
