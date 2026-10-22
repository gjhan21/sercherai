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

  <el-row :gutter="12" style="margin-bottom: 12px">
    <el-col :span="12" v-for="item in workspace.marketCoverageAssetRows.filter(r => ['STOCK', 'FUTURES'].includes(r.key))" :key="item.key">
      <div class="card" v-loading="workspace.marketCoverageLoading">
        <div style="display: flex; justify-content: space-between; align-items: center;">
          <span style="font-weight: 600; font-size: 14px; color: #303133;">{{ item.label }} 数据同步水位</span>
          <el-tag :type="item.latestTradeDate !== '-' ? 'success' : 'info'" size="small">
            {{ item.latestTradeDate !== '-' ? '已同步' : '未导入' }}
          </el-tag>
        </div>
        <div style="margin-top: 14px; display: flex; align-items: baseline; gap: 8px;">
          <span style="font-size: 24px; font-weight: 700; color: #1f2937;">
            {{ item.latestTradeDate }}
          </span>
          <span class="muted" style="font-size: 12px;">最新交易日</span>
        </div>
        <div style="margin-top: 12px; display: flex; justify-content: space-between; font-size: 13px; border-top: 1px solid #f3f4f6; padding-top: 10px;">
          <span class="muted">主数据证券全集：<strong style="color: #374151;">{{ item.universeCount }}</strong></span>
          <span class="muted">行情覆盖品种数：<strong style="color: #374151;">{{ item.quotesCoverageCount }}</strong></span>
        </div>
      </div>
    </el-col>
  </el-row>

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
