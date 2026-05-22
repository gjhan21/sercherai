<script setup>
import ProviderCapabilityMatrix from "../../components/data-sources/ProviderCapabilityMatrix.vue";
import ProviderGovernanceOverviewPanel from "../../components/data-sources/ProviderGovernanceOverviewPanel.vue";
import RoutingPolicyPanel from "../../components/data-sources/RoutingPolicyPanel.vue";
import { useDataSourcesWorkspaceContext } from "../../composables/useDataSourcesWorkspace.js";

const workspace = useDataSourcesWorkspaceContext();
</script>

<template>
  <div class="card" style="margin-bottom: 12px">
    <div class="section-header">
      <div>
        <h3 style="margin: 0">治理操作提示</h3>
        <p class="muted" style="margin: 6px 0 0">
          这里聚焦 provider 能力、路由策略与质量画像。需要补当天行情或执行手动增量同步时，请前往同步台，确保请求走当前生效的市场数据同步链路。
        </p>
      </div>
      <div class="inline-actions inline-actions--left">
        <router-link :to="workspace.buildSectionLocation('sync')">
          <el-button type="primary">前往同步台</el-button>
        </router-link>
      </div>
    </div>
  </div>

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
