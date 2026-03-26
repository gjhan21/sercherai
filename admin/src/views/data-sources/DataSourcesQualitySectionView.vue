<script setup>
import QualityLogsPanel from "../../components/data-sources/QualityLogsPanel.vue";
import { useDataSourcesWorkspaceContext } from "../../composables/useDataSourcesWorkspace.js";

const workspace = useDataSourcesWorkspaceContext();
</script>

<template>
  <QualityLogsPanel
    :filters="workspace.qualityFilters"
    :asset-class-options="workspace.assetClassOptions"
    :quality-data-kind-options="workspace.qualityDataKindOptions"
    :severity-filter-options="workspace.severityFilterOptions"
    :loading="workspace.qualityLogsLoading"
    :items="workspace.qualityLogItems"
    :total="workspace.qualityLogTotal"
    :summary-all="workspace.qualitySummaryAll"
    :summary-stock="workspace.qualitySummaryStock"
    :summary-futures="workspace.qualitySummaryFutures"
    :quality-lookback-label="workspace.qualityLookbackLabel"
    :stock-governance-summary-items="workspace.stockGovernanceSummaryItems"
    :issue-quick-options="workspace.qualityIssueQuickOptions"
    :source-buckets="workspace.qualitySourceBuckets"
    :issue-buckets="workspace.qualityIssueBuckets"
    :severity-buckets="workspace.qualitySeverityBuckets"
    :payload-dialog-visible="workspace.qualityPayloadDialogVisible"
    :selected-log="workspace.selectedQualityLog"
    :selected-payload-text="workspace.selectedQualityPayloadText"
    @search="workspace.handleQualitySearch"
    @reset-filters="workspace.resetQualityFilters"
    @apply-issue-filter="workspace.applyQualityIssueQuickFilter"
    @open-payload="workspace.openQualityPayload"
    @update:payload-dialog-visible="workspace.handleQualityPayloadDialogVisible"
  />
</template>
