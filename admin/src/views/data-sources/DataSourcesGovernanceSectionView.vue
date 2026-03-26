<script setup>
import ProviderCapabilityMatrix from "../../components/data-sources/ProviderCapabilityMatrix.vue";
import ProviderGovernanceOverviewPanel from "../../components/data-sources/ProviderGovernanceOverviewPanel.vue";
import RoutingPolicyPanel from "../../components/data-sources/RoutingPolicyPanel.vue";
import { useDataSourcesWorkspaceContext } from "../../composables/useDataSourcesWorkspace.js";

const workspace = useDataSourcesWorkspaceContext();
</script>

<template>
  <ProviderGovernanceOverviewPanel
    :loading="workspace.governanceLoading"
    :stock-overview="workspace.governanceOverviewStock"
    :futures-overview="workspace.governanceOverviewFutures"
    :provider-score-rows="workspace.governanceProviderScoreRows"
    @refresh="workspace.fetchGovernanceDashboard({ preserveFeedback: true })"
  />

  <ProviderCapabilityMatrix :items="workspace.providerCapabilities" />

  <RoutingPolicyPanel
    :policies="workspace.routingPolicies"
    :dialog-visible="workspace.routingPolicyDialogVisible"
    :form="workspace.routingPolicyForm"
    :provider-options="workspace.routingPolicyProviderOptions"
    :submitting="workspace.routingPolicySubmitting"
    :can-edit-data-sources="workspace.canEditDataSources"
    @edit="workspace.handleEditRoutingPolicy"
    @save="workspace.handleSaveRoutingPolicy"
    @cancel="workspace.closeRoutingPolicyDialog"
    @update:dialog-visible="workspace.handleRoutingPolicyDialogVisible"
  />
</template>
