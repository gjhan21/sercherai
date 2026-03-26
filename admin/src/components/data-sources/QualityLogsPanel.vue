<script setup>
import { MARKET_QUALITY_LOOKBACK_OPTIONS, formatStockFallbackSourceSummary, marketQualitySeverityTagType } from "../../lib/market-data-admin.js";
import {
  formatQualitySummaryLatest,
  formatQualitySummaryLatestError
} from "../../lib/data-sources-admin.js";

defineProps({
  filters: { type: Object, required: true },
  assetClassOptions: { type: Array, default: () => [] },
  qualityDataKindOptions: { type: Array, default: () => [] },
  severityFilterOptions: { type: Array, default: () => [] },
  loading: Boolean,
  items: { type: Array, default: () => [] },
  total: Number,
  summaryAll: { type: Object, default: null },
  summaryStock: { type: Object, default: null },
  summaryFutures: { type: Object, default: null },
  qualityLookbackLabel: String,
  stockGovernanceSummaryItems: { type: Array, default: () => [] },
  issueQuickOptions: { type: Array, default: () => [] },
  sourceBuckets: { type: Array, default: () => [] },
  issueBuckets: { type: Array, default: () => [] },
  severityBuckets: { type: Array, default: () => [] },
  payloadDialogVisible: Boolean,
  selectedLog: { type: Object, default: null },
  selectedPayloadText: String
});

const emit = defineEmits([
  "search",
  "reset-filters",
  "apply-issue-filter",
  "open-payload",
  "update:payloadDialogVisible"
]);
</script>

<template>
  <div class="card" style="margin-top: 12px">
    <div class="section-header">
      <div>
        <h3 style="margin: 0">市场数据质量日志</h3>
        <p class="muted" style="margin: 6px 0 0">
          查看多源行情、主数据和派生重建过程中产生的告警、回退与补写记录。
        </p>
      </div>
      <div class="inline-actions inline-actions--left">
        <el-select v-model="filters.hours" style="width: 140px">
          <el-option v-for="item in MARKET_QUALITY_LOOKBACK_OPTIONS" :key="`hours-${item.value}`" :label="item.label" :value="item.value" />
        </el-select>
        <el-select v-model="filters.asset_class" style="width: 140px">
          <el-option v-for="item in assetClassOptions" :key="item.value || 'asset-all'" :label="item.label" :value="item.value" />
        </el-select>
        <el-select v-model="filters.data_kind" style="width: 150px">
          <el-option v-for="item in qualityDataKindOptions" :key="item.value || 'kind-all'" :label="item.label" :value="item.value" />
        </el-select>
        <el-select v-model="filters.severity" style="width: 140px">
          <el-option v-for="item in severityFilterOptions" :key="item.value || 'severity-all'" :label="item.label" :value="item.value" />
        </el-select>
        <el-input v-model="filters.issue_code" clearable placeholder="输入 issue_code" style="width: 180px" />
        <el-button :loading="loading" @click="emit('search')">查询日志</el-button>
        <el-button @click="emit('reset-filters')">重置</el-button>
      </div>
    </div>

    <div class="truth-summary-grid" style="margin-top: 12px">
      <div class="truth-summary-card">
        <div class="truth-summary-card__title">{{ qualityLookbackLabel }}总体概览</div>
        <ul class="truth-summary-list">
          <li>日志总数：{{ summaryAll?.total_count || 0 }}</li>
          <li>错误：{{ summaryAll?.error_count || 0 }}</li>
          <li>警告：{{ summaryAll?.warn_count || 0 }}</li>
          <li>信息：{{ summaryAll?.info_count || 0 }}</li>
          <li>涉及来源：{{ summaryAll?.distinct_source_count || 0 }}</li>
          <li>最近事件：{{ formatQualitySummaryLatest(summaryAll) }}</li>
          <li>最近错误：{{ formatQualitySummaryLatestError(summaryAll) }}</li>
        </ul>
      </div>
      <div class="truth-summary-card">
        <div class="truth-summary-card__header">
          <div class="truth-summary-card__title">{{ qualityLookbackLabel }}股票质量概览</div>
        </div>
        <ul class="truth-summary-list">
          <li>日志总数：{{ summaryStock?.total_count || 0 }}</li>
          <li>错误：{{ summaryStock?.error_count || 0 }}</li>
          <li>警告：{{ summaryStock?.warn_count || 0 }}</li>
          <li>信息：{{ summaryStock?.info_count || 0 }}</li>
          <li>涉及来源：{{ summaryStock?.distinct_source_count || 0 }}</li>
          <li>最新交易日：{{ summaryStock?.latest_trade_date || "暂无" }}</li>
          <li>最近事件：{{ formatQualitySummaryLatest(summaryStock) }}</li>
          <li>最近错误：{{ formatQualitySummaryLatestError(summaryStock) }}</li>
          <li>回填来源：{{ formatStockFallbackSourceSummary(summaryStock) }}</li>
        </ul>
        <ul class="truth-summary-list truth-summary-list--compact">
          <li v-for="item in stockGovernanceSummaryItems" :key="item.key">
            {{ item.label }}：{{ item.value }}
          </li>
        </ul>
      </div>
      <div class="truth-summary-card">
        <div class="truth-summary-card__header">
          <div class="truth-summary-card__title">{{ qualityLookbackLabel }}期货质量概览</div>
        </div>
        <ul class="truth-summary-list">
          <li>日志总数：{{ summaryFutures?.total_count || 0 }}</li>
          <li>错误：{{ summaryFutures?.error_count || 0 }}</li>
          <li>警告：{{ summaryFutures?.warn_count || 0 }}</li>
          <li>信息：{{ summaryFutures?.info_count || 0 }}</li>
          <li>涉及来源：{{ summaryFutures?.distinct_source_count || 0 }}</li>
          <li>最近事件：{{ formatQualitySummaryLatest(summaryFutures) }}</li>
          <li>最近错误：{{ formatQualitySummaryLatestError(summaryFutures) }}</li>
        </ul>
      </div>
    </div>

    <div v-if="issueQuickOptions.length" class="issue-quick-filter">
      <span class="issue-quick-filter__label">快速筛选：</span>
      <el-tag
        v-for="item in issueQuickOptions"
        :key="item.value"
        class="issue-quick-filter__tag"
        :type="filters.issue_code === item.value ? 'primary' : 'info'"
        effect="plain"
        @click="emit('apply-issue-filter', item.value)"
      >
        {{ item.label }}
      </el-tag>
      <el-button v-if="filters.issue_code" link type="primary" @click="emit('apply-issue-filter', '')">
        清空 issue_code
      </el-button>
    </div>

    <div class="truth-summary-grid" style="margin-top: 12px">
      <div class="truth-summary-card">
        <div class="truth-summary-card__title">当前筛选结果 · 来源分布</div>
        <ul class="truth-summary-list">
          <li v-for="item in sourceBuckets" :key="`source-bucket-${item.value}`">{{ item.label }}：{{ item.count }}</li>
          <li v-if="!sourceBuckets.length">暂无统计</li>
        </ul>
      </div>
      <div class="truth-summary-card">
        <div class="truth-summary-card__title">当前筛选结果 · 问题分布</div>
        <ul class="truth-summary-list">
          <li v-for="item in issueBuckets" :key="`issue-bucket-${item.value}`">{{ item.label }}：{{ item.count }}</li>
          <li v-if="!issueBuckets.length">暂无统计</li>
        </ul>
      </div>
      <div class="truth-summary-card">
        <div class="truth-summary-card__title">当前筛选结果 · 级别分布</div>
        <ul class="truth-summary-list">
          <li v-for="item in severityBuckets" :key="`severity-bucket-${item.value}`">{{ item.label }}：{{ item.count }}</li>
          <li v-if="!severityBuckets.length">暂无统计</li>
        </ul>
      </div>
    </div>

    <el-table :data="items" border stripe v-loading="loading" empty-text="暂无质量日志" style="margin-top: 12px">
      <el-table-column prop="asset_class" label="资产" min-width="90" />
      <el-table-column prop="data_kind" label="数据类型" min-width="140" />
      <el-table-column prop="instrument_key" label="标的" min-width="140">
        <template #default="{ row }">{{ row.instrument_key || "-" }}</template>
      </el-table-column>
      <el-table-column prop="trade_date" label="交易日" min-width="110">
        <template #default="{ row }">{{ row.trade_date || "-" }}</template>
      </el-table-column>
      <el-table-column prop="source_key" label="来源" min-width="110">
        <template #default="{ row }">{{ row.source_key || "-" }}</template>
      </el-table-column>
      <el-table-column label="级别" min-width="100">
        <template #default="{ row }">
          <el-tag :type="marketQualitySeverityTagType(row.severity)">{{ row.severity || "-" }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="issue_code" label="问题编码" min-width="180">
        <template #default="{ row }">
          <el-button link type="primary" @click="emit('apply-issue-filter', row.issue_code)">{{ row.issue_code || "-" }}</el-button>
        </template>
      </el-table-column>
      <el-table-column prop="issue_message" label="问题说明" min-width="260">
        <template #default="{ row }">{{ row.issue_message || "-" }}</template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" min-width="180" />
      <el-table-column label="操作" min-width="220" align="center">
        <template #default="{ row }">
          <div class="inline-actions inline-actions--left">
            <el-button size="small" :disabled="!row.payload" @click="emit('open-payload', row)">查看 payload</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-text type="info">共 {{ total }} 条日志，当前展示最近 20 条</el-text>
    </div>

    <el-dialog
      :model-value="payloadDialogVisible"
      title="质量日志 payload 详情"
      width="760px"
      destroy-on-close
      @update:model-value="emit('update:payloadDialogVisible', $event)"
    >
      <el-descriptions :column="2" border>
        <el-descriptions-item label="资产">{{ selectedLog?.asset_class || "-" }}</el-descriptions-item>
        <el-descriptions-item label="来源">{{ selectedLog?.source_key || "-" }}</el-descriptions-item>
        <el-descriptions-item label="问题编码">{{ selectedLog?.issue_code || "-" }}</el-descriptions-item>
        <el-descriptions-item label="级别">{{ selectedLog?.severity || "-" }}</el-descriptions-item>
        <el-descriptions-item label="交易日">{{ selectedLog?.trade_date || "-" }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ selectedLog?.created_at || "-" }}</el-descriptions-item>
      </el-descriptions>

      <div style="margin-top: 12px">
        <div class="truth-summary-card__title">问题说明</div>
        <p class="payload-message">{{ selectedLog?.issue_message || "-" }}</p>
      </div>

      <div style="margin-top: 12px">
        <div class="truth-summary-card__title">payload</div>
        <pre class="payload-viewer">{{ selectedPayloadText || "无 payload" }}</pre>
      </div>

      <template #footer>
        <el-button @click="emit('update:payloadDialogVisible', false)">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>
