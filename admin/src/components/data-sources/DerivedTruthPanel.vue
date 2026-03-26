<script setup>
import { formatSummaryWarnings } from "../../lib/data-sources-admin.js";

defineProps({
  canEditMarket: Boolean,
  rebuildForm: { type: Object, required: true },
  rebuildingStockTruth: Boolean,
  rebuildingFuturesTruth: Boolean,
  stockRebuildSummary: { type: Object, default: null },
  futuresRebuildSummary: { type: Object, default: null }
});

const emit = defineEmits(["rebuild-stock", "rebuild-futures"]);
</script>

<template>
  <div class="card" style="margin-bottom: 12px">
    <div class="section-header">
      <div>
        <h3 style="margin: 0">本地 Truth 派生重建</h3>
        <p class="muted" style="margin: 6px 0 0">
          从本地 truth 行情重建股票状态和期货主力映射，不直接触发外部源拉取。
        </p>
      </div>
      <div class="inline-actions inline-actions--left">
        <el-date-picker
          v-model="rebuildForm.trade_date"
          type="date"
          value-format="YYYY-MM-DD"
          placeholder="可选：指定交易日"
          style="width: 180px"
        />
        <el-input-number v-model="rebuildForm.days" :min="1" :max="30" controls-position="right" />
        <el-button v-if="canEditMarket" type="primary" plain :loading="rebuildingStockTruth" @click="emit('rebuild-stock')">
          重建股票状态
        </el-button>
        <el-button v-if="canEditMarket" type="primary" :loading="rebuildingFuturesTruth" @click="emit('rebuild-futures')">
          重建期货主力映射
        </el-button>
      </div>
    </div>

    <div class="truth-summary-grid" style="margin-top: 12px">
      <div class="truth-summary-card">
        <div class="truth-summary-card__title">最近一次股票状态重建</div>
        <template v-if="stockRebuildSummary">
          <ul class="truth-summary-list">
            <li>交易日：{{ stockRebuildSummary.trade_date || "-" }}</li>
            <li>窗口：{{ stockRebuildSummary.start_date || "-" }} ~ {{ stockRebuildSummary.end_date || "-" }}</li>
            <li>窗口天数：{{ stockRebuildSummary.days || 0 }}</li>
            <li>truth 行情：{{ stockRebuildSummary.truth_bar_count || 0 }}</li>
            <li>状态条数：{{ stockRebuildSummary.stock_status_count || 0 }}</li>
            <li>来源：{{ stockRebuildSummary.source_key || "LOCAL_TRUTH" }}</li>
            <li>记录时间：{{ stockRebuildSummary.created_at || "-" }}</li>
            <li>警告：{{ formatSummaryWarnings(stockRebuildSummary) }}</li>
          </ul>
        </template>
        <p v-else class="muted">暂无股票 truth 重建摘要</p>
      </div>

      <div class="truth-summary-card">
        <div class="truth-summary-card__title">最近一次期货主力映射重建</div>
        <template v-if="futuresRebuildSummary">
          <ul class="truth-summary-list">
            <li>交易日：{{ futuresRebuildSummary.trade_date || "-" }}</li>
            <li>窗口：{{ futuresRebuildSummary.start_date || "-" }} ~ {{ futuresRebuildSummary.end_date || "-" }}</li>
            <li>窗口天数：{{ futuresRebuildSummary.days || 0 }}</li>
            <li>truth 行情：{{ futuresRebuildSummary.truth_bar_count || 0 }}</li>
            <li>映射条数：{{ futuresRebuildSummary.futures_mapping_count || 0 }}</li>
            <li>来源：{{ futuresRebuildSummary.source_key || "LOCAL_TRUTH" }}</li>
            <li>记录时间：{{ futuresRebuildSummary.created_at || "-" }}</li>
            <li>警告：{{ formatSummaryWarnings(futuresRebuildSummary) }}</li>
          </ul>
        </template>
        <p v-else class="muted">暂无期货 truth 重建摘要</p>
      </div>
    </div>
  </div>
</template>
