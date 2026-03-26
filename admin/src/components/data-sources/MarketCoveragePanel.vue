<script setup>
import {
  formatFallbackSources
} from "../../lib/data-sources-admin.js";
import {
  buildUniverseSnapshotDigest,
  formatMarketAssetScopeSummary,
  formatMarketBackfillStatusLabel,
  marketBackfillStatusTagType
} from "../../lib/market-data-admin.js";

defineProps({
  loading: Boolean,
  overviewCards: { type: Array, default: () => [] },
  summary: { type: Object, default: null },
  assetRows: { type: Array, default: () => [] },
  snapshots: { type: Array, default: () => [] },
  snapshotsLoading: Boolean,
  backfillRuns: { type: Array, default: () => [] },
  backfillRunsLoading: Boolean
});

const emit = defineEmits(["refresh", "open-jobs"]);
</script>

<template>
  <div class="card" style="margin-bottom: 12px">
    <div class="section-header">
      <div>
        <h3 style="margin: 0">全市场覆盖治理</h3>
        <p class="muted" style="margin: 6px 0 0">
          这里看证券全集、行情覆盖、增强因子覆盖和最近回填情况，适合先判断数据底座是否完整，再决定去哪里补数。
        </p>
      </div>
      <div class="inline-actions inline-actions--left">
        <el-button :loading="loading || snapshotsLoading || backfillRunsLoading" @click="emit('refresh')">
          刷新治理概览
        </el-button>
        <el-button type="primary" plain @click="emit('open-jobs')">去任务中心</el-button>
      </div>
    </div>

    <div class="governance-kpi-grid" style="margin-top: 12px">
      <div v-for="item in overviewCards" :key="item.key" class="governance-kpi-card" :class="`is-${item.tone}`">
        <div class="governance-kpi-card__title">{{ item.title }}</div>
        <div class="governance-kpi-card__value">{{ item.value }}</div>
        <div class="governance-kpi-card__helper">{{ item.helper }}</div>
      </div>
    </div>

    <div class="truth-summary-grid" style="margin-top: 12px">
      <div class="truth-summary-card">
        <div class="truth-summary-card__title">回退源摘要</div>
        <ul v-if="summary" class="truth-summary-list">
          <li>最近交易日：{{ summary.latest_trade_date || "-" }}</li>
          <li>主数据覆盖：{{ summary.master_coverage_count || 0 }}</li>
          <li>回退源分布：{{ formatFallbackSources(summary.fallback_source_summary) }}</li>
        </ul>
        <p v-else class="muted">暂无覆盖率摘要</p>
      </div>
      <div class="truth-summary-card">
        <div class="truth-summary-card__title">最近 Universe 快照</div>
        <ul v-if="snapshots.length" class="truth-summary-list">
          <li>最近快照数：{{ snapshots.length }}</li>
          <li>最新快照：{{ buildUniverseSnapshotDigest(snapshots[0]) }}</li>
          <li>资产范围：{{ formatMarketAssetScopeSummary(snapshots[0]?.scope) }}</li>
        </ul>
        <p v-else class="muted">暂无 Universe 快照</p>
      </div>
      <div class="truth-summary-card">
        <div class="truth-summary-card__title">最近回填任务</div>
        <ul v-if="backfillRuns.length" class="truth-summary-list">
          <li>最近总单数：{{ backfillRuns.length }}</li>
          <li>最新状态：{{ formatMarketBackfillStatusLabel(backfillRuns[0]?.status) }}</li>
          <li>资产范围：{{ formatMarketAssetScopeSummary(backfillRuns[0]?.asset_scope) }}</li>
        </ul>
        <p v-else class="muted">暂无回填总单</p>
      </div>
    </div>

    <div class="governance-table-grid" style="margin-top: 12px">
      <div class="governance-table-card">
        <div class="toolbar jobs-table-toolbar">
          <el-text type="primary">资产覆盖拆分</el-text>
          <span class="muted">按资产类型看 universe、行情和增强因子覆盖，不把所有资产混成一个数字。</span>
        </div>
        <el-table :data="assetRows" border stripe size="small" v-loading="loading" empty-text="暂无资产覆盖数据">
          <el-table-column prop="label" label="资产类型" min-width="100" />
          <el-table-column prop="universeCount" label="全集" min-width="90" />
          <el-table-column prop="masterCoverageCount" label="主数据" min-width="90" />
          <el-table-column prop="quotesCoverageCount" label="行情" min-width="90" />
          <el-table-column prop="dailyBasicCoverageCount" label="daily_basic" min-width="110" />
          <el-table-column prop="moneyflowCoverageCount" label="moneyflow" min-width="100" />
          <el-table-column prop="latestTradeDate" label="最新交易日" min-width="120" />
        </el-table>
      </div>

      <div class="governance-table-card">
        <div class="toolbar jobs-table-toolbar">
          <el-text type="primary">最近 Universe 快照</el-text>
          <span class="muted">快照是回填入口的第一步，先看范围和来源是否对。</span>
        </div>
        <el-table :data="snapshots" border stripe size="small" v-loading="snapshotsLoading" empty-text="暂无快照">
          <el-table-column prop="snapshot_date" label="快照日期" min-width="110" />
          <el-table-column label="资产范围" min-width="180">
            <template #default="{ row }">{{ formatMarketAssetScopeSummary(row.scope) }}</template>
          </el-table-column>
          <el-table-column prop="source_key" label="来源" min-width="100" />
          <el-table-column prop="created_at" label="创建时间" min-width="170" />
        </el-table>
      </div>
    </div>

    <div class="governance-table-card" style="margin-top: 12px">
      <div class="toolbar jobs-table-toolbar">
        <el-text type="primary">最近回填总单</el-text>
        <span class="muted">在数据源页先看最近是否有人补过数，真正操作入口统一回任务中心。</span>
      </div>
      <el-table :data="backfillRuns" border stripe size="small" v-loading="backfillRunsLoading" empty-text="暂无回填总单">
        <el-table-column prop="created_at" label="创建时间" min-width="170" />
        <el-table-column label="运行类型" min-width="100">
          <template #default="{ row }">{{ row.run_type || "-" }}</template>
        </el-table-column>
        <el-table-column label="资产范围" min-width="180">
          <template #default="{ row }">{{ formatMarketAssetScopeSummary(row.asset_scope) }}</template>
        </el-table-column>
        <el-table-column label="状态" min-width="110">
          <template #default="{ row }">
            <el-tag :type="marketBackfillStatusTagType(row.status)">
              {{ formatMarketBackfillStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="current_stage" label="当前阶段" min-width="120" />
        <el-table-column prop="source_key" label="来源" min-width="100" />
      </el-table>
    </div>
  </div>
</template>
