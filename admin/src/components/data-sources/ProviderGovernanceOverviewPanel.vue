<script setup>
import { formatGovernanceSuggestion } from "../../lib/data-sources-admin.js";

defineProps({
  loading: Boolean,
  stockOverview: { type: Object, default: null },
  futuresOverview: { type: Object, default: null },
  providerScoreRows: { type: Array, default: () => [] }
});

const emit = defineEmits(["refresh"]);
</script>

<template>
  <div class="card" style="margin-bottom: 12px">
    <div class="section-header">
      <div>
        <h3 style="margin: 0">供应商治理总览</h3>
        <p class="muted" style="margin: 6px 0 0">
          把 provider 质量画像、治理建议和 truth 依赖摘要集中到这里，先看股票与期货日行情治理态势，再决定是否调整主源与回退链。
        </p>
      </div>
      <div class="inline-actions inline-actions--left">
        <el-button :loading="loading" @click="emit('refresh')">刷新治理台</el-button>
      </div>
    </div>

    <div class="truth-summary-grid" style="margin-top: 12px">
      <div class="truth-summary-card">
        <div class="truth-summary-card__header">
          <div class="truth-summary-card__title">股票日行情治理</div>
          <el-tag type="info" effect="plain">DAILY_BARS</el-tag>
        </div>
        <ul v-if="stockOverview" class="truth-summary-list">
          <li>质量事件：{{ stockOverview.quality_summary?.total_count || 0 }}</li>
          <li>最近问题：{{ stockOverview.quality_summary?.latest_issue_code || "-" }}</li>
          <li>治理画像：{{ stockOverview.provider_scores?.length || 0 }} 个 provider</li>
          <li>truth 摘要：{{ stockOverview.latest_derived_truth?.issue_code || "-" }}</li>
        </ul>
        <p v-else class="muted">暂无股票治理总览</p>
      </div>

      <div class="truth-summary-card">
        <div class="truth-summary-card__header">
          <div class="truth-summary-card__title">期货日行情治理</div>
          <el-tag type="info" effect="plain">DAILY_BARS</el-tag>
        </div>
        <ul v-if="futuresOverview" class="truth-summary-list">
          <li>质量事件：{{ futuresOverview.quality_summary?.total_count || 0 }}</li>
          <li>最近问题：{{ futuresOverview.quality_summary?.latest_issue_code || "-" }}</li>
          <li>治理画像：{{ futuresOverview.provider_scores?.length || 0 }} 个 provider</li>
          <li>truth 摘要：{{ futuresOverview.latest_derived_truth?.issue_code || "-" }}</li>
        </ul>
        <p v-else class="muted">暂无期货治理总览</p>
      </div>
    </div>

    <el-table
      :data="providerScoreRows"
      border
      stripe
      size="small"
      style="margin-top: 12px"
      v-loading="loading"
      empty-text="暂无 provider 质量画像"
    >
      <el-table-column prop="asset_class" label="资产" width="90" />
      <el-table-column prop="provider_key" label="Provider" min-width="120" />
      <el-table-column prop="freshness_score" label="新鲜度" width="100" />
      <el-table-column prop="stability_score" label="稳定性" width="100" />
      <el-table-column prop="trust_score" label="可信度" width="100" />
      <el-table-column prop="overall_score" label="综合分" width="100" />
      <el-table-column label="治理建议" min-width="220">
        <template #default="{ row }">
          <div>{{ formatGovernanceSuggestion(row) }}</div>
          <el-text v-if="row.latest_issue_code" type="info" size="small">
            最近问题：{{ row.latest_issue_code }}
          </el-text>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
