<script setup>
import {
  buildSyncMetricTags,
  formatRequestedSourceLabel,
  formatSyncRequestScope,
  formatSyncResolvedSourceKeys
} from "../../lib/data-sources-admin.js";

defineProps({
  cards: { type: Array, default: () => [] },
  defaultStockSourceKey: String,
  defaultFuturesSourceKey: String,
  defaultMarketNewsSourceKey: String
});
</script>

<template>
  <div class="card" style="margin-bottom: 12px">
    <div class="section-header">
      <div>
        <h3 style="margin: 0">数据同步操作台</h3>
        <p class="muted" style="margin: 6px 0 0">
          股票、期货、市场资讯的手动同步统一收口在这里；如果上游源失败，可直接切换默认源、AUTO 或自定义回退链路。
        </p>
      </div>
      <div class="inline-actions inline-actions--left">
        <el-tag type="warning" effect="plain">股票默认源：{{ defaultStockSourceKey || "-" }}</el-tag>
        <el-tag type="warning" effect="plain">期货默认源：{{ defaultFuturesSourceKey || "-" }}</el-tag>
        <el-tag type="warning" effect="plain">资讯默认源：{{ defaultMarketNewsSourceKey || "-" }}</el-tag>
      </div>
    </div>

    <div class="market-sync-grid" style="margin-top: 12px">
      <div v-for="card in cards" :key="card.key" class="market-sync-card truth-summary-card">
        <div class="truth-summary-card__header">
          <div class="truth-summary-card__title">{{ card.title }}</div>
          <el-tag type="info" effect="plain">
            {{ card.options.length ? `${card.options.length} 个可用源` : "无可用源" }}
          </el-tag>
        </div>

        <div class="market-sync-form-grid sync-form-grid">
          <el-select
            v-model="card.form.source_key"
            filterable
            allow-create
            default-first-option
            :placeholder="`选择${card.title.replace('同步', '')}源`"
          >
            <el-option
              v-for="item in card.options"
              :key="`${card.key}-${item.value}`"
              :label="item.label"
              :value="item.value"
              :disabled="item.disabled"
            />
          </el-select>
          <el-input v-model="card.form[card.inputKey]" :placeholder="card.placeholder" />
          <div v-if="card.hasLimit" class="market-sync-inline-numbers sync-inline-numbers">
            <el-input-number
              v-model="card.form.days"
              :min="card.minDays"
              :max="card.maxDays"
              controls-position="right"
            />
            <el-input-number
              v-model="card.form.limit"
              :min="card.minLimit"
              :max="card.maxLimit"
              controls-position="right"
            />
          </div>
          <el-input-number
            v-else
            v-model="card.form.days"
            :min="card.minDays"
            :max="card.maxDays"
            controls-position="right"
          />
          <div class="market-sync-actions">
            <el-button
              v-for="action in (card.actions && card.actions.length ? card.actions : [{ key: 'default', label: card.buttonText, type: 'primary', run: card.run }])"
              :key="`${card.key}-${action.key}`"
              :type="action.type || 'default'"
              :loading="card.loading"
              @click="action.run"
            >
              {{ action.label }}
            </el-button>
          </div>
        </div>

        <div class="sync-inline-hint">
          <el-text type="info" size="small">{{ card.hint }}</el-text>
        </div>

        <div v-if="card.logs && card.logs.length" class="market-sync-log-panel">
          <div class="market-sync-log-panel__head">
            <strong>执行日志</strong>
            <span>{{ card.logs.length }} 条</span>
          </div>
          <div class="market-sync-log-list">
            <div
              v-for="item in card.logs"
              :key="item.id"
              class="market-sync-log-item"
              :class="`is-${item.level || 'info'}`"
            >
              <span class="market-sync-log-item__time">{{ item.time }}</span>
              <span class="market-sync-log-item__message">{{ item.message }}</span>
            </div>
          </div>
        </div>

        <div v-if="card.result" class="market-sync-last-result sync-last-result">
          <div class="market-sync-last-result__head sync-last-result__head">
            <strong>最近一次{{ card.title }}</strong>
            <span>
              {{ formatRequestedSourceLabel(card.result) }}
              ->
              {{ formatSyncResolvedSourceKeys(card.result.result, card.result.source_key) }}
            </span>
          </div>
          <div class="sync-result-tags">
            <el-tag v-for="item in buildSyncMetricTags(card.result.result)" :key="`${card.key}-${item.key}`" :type="item.type">
              {{ item.label }}
            </el-tag>
            <el-tag v-if="buildSyncMetricTags(card.result.result).length === 0" type="info">
              处理 {{ card.result.count || 0 }} 条
            </el-tag>
          </div>
          <el-text type="info" size="small">
            范围：{{ formatSyncRequestScope(card.result[card.scopeKey], card.emptyScopeLabel) }}
          </el-text>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.market-sync-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(420px, 1fr));
  gap: 16px;
}

.market-sync-card {
  padding: 18px 18px 16px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  container-type: inline-size;
}

.market-sync-form-grid {
  grid-template-columns: minmax(0, 1.1fr) minmax(0, 1.3fr) minmax(128px, 164px) auto;
  align-items: start;
}

.market-sync-inline-numbers :deep(.el-input-number),
.market-sync-form-grid :deep(.el-select),
.market-sync-form-grid :deep(.el-input),
.market-sync-form-grid :deep(.el-input-number) {
  width: 100%;
}

.market-sync-last-result__head span {
  word-break: break-word;
}

.market-sync-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.market-sync-log-panel {
  margin-top: 12px;
  padding: 12px;
  border: 1px solid #d9e6f7;
  border-radius: 12px;
  background: #f7fbff;
}

.market-sync-log-panel__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
  color: #355070;
  font-size: 13px;
}

.market-sync-log-list {
  display: grid;
  gap: 6px;
  max-height: 188px;
  overflow: auto;
}

.market-sync-log-item {
  display: grid;
  grid-template-columns: 72px minmax(0, 1fr);
  gap: 8px;
  padding: 8px 10px;
  border-radius: 10px;
  background: #ffffff;
  color: #355070;
  font-size: 13px;
}

.market-sync-log-item__time {
  color: #7f8ea3;
  font-variant-numeric: tabular-nums;
}

.market-sync-log-item__message {
  word-break: break-word;
}

.market-sync-log-item.is-success {
  background: #eef9f1;
  color: #1f7a45;
}

.market-sync-log-item.is-danger {
  background: #fff1f0;
  color: #b42318;
}

@container (max-width: 620px) {
  .market-sync-form-grid {
    grid-template-columns: 1fr;
  }

  .market-sync-inline-numbers {
    grid-template-columns: 1fr 1fr;
  }

  .market-sync-actions {
    justify-content: stretch;
  }

  .market-sync-actions :deep(.el-button) {
    flex: 1 1 auto;
  }

  .market-sync-log-item {
    grid-template-columns: 1fr;
  }
}

@container (max-width: 460px) {
  .market-sync-inline-numbers {
    grid-template-columns: 1fr;
  }
}
</style>
