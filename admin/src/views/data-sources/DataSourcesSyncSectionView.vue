<script setup>
import MarketSyncConsole from "../../components/data-sources/MarketSyncConsole.vue";
import { useDataSourcesWorkspaceContext } from "../../composables/useDataSourcesWorkspace.js";

const workspace = useDataSourcesWorkspaceContext();
</script>

<template>
  <MarketSyncConsole
    :cards="workspace.syncCards"
    :default-stock-source-key="workspace.defaultStockSourceKey"
    :default-futures-source-key="workspace.defaultFuturesSourceKey"
    :default-market-news-source-key="workspace.defaultMarketNewsSourceKey"
  />
</template>
