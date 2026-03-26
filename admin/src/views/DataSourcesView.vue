<script setup>
import { provide } from "vue";
import {
  DATA_SOURCES_WORKSPACE_KEY,
  useDataSourcesWorkspace
} from "../composables/useDataSourcesWorkspace.js";

const workspace = useDataSourcesWorkspace();

provide(DATA_SOURCES_WORKSPACE_KEY, workspace);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">数据源管理</h1>
        <p class="muted">统一处理数据源配置、手动同步、健康检查、本地 truth 派生重建与质量日志，并按领域切到独立子页。</p>
      </div>
      <div class="toolbar">
        <el-tag type="warning" effect="plain">默认行情源：{{ workspace.defaultStockSourceKey || "-" }}</el-tag>
        <el-tag type="warning" effect="plain">期货默认源：{{ workspace.defaultFuturesSourceKey || "-" }}</el-tag>
        <el-tag type="warning" effect="plain">资讯默认源：{{ workspace.defaultMarketNewsSourceKey || "-" }}</el-tag>
      </div>
    </div>

    <el-alert
      v-if="workspace.errorMessage"
      :title="workspace.errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />
    <el-alert
      v-if="workspace.message"
      :title="workspace.message"
      type="success"
      show-icon
      style="margin-bottom: 12px"
    />

    <div class="data-sources-nav">
      <router-link
        v-for="item in workspace.sectionItems"
        :key="item.key"
        :to="workspace.buildSectionLocation(item.key)"
        class="data-sources-nav__item"
        :class="{ 'is-active': workspace.activeSectionKey === item.key }"
      >
        <span class="data-sources-nav__label">{{ item.label }}</span>
        <span class="data-sources-nav__desc">{{ item.description }}</span>
      </router-link>
    </div>

    <div class="data-sources-shell__body">
      <router-view />
    </div>
  </div>
</template>

<style>
.config-line {
  line-height: 1.5;
  word-break: break-all;
}

.source-key-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.config-line--muted {
  color: #6b7280;
}

.inline-actions {
  display: flex;
  justify-content: flex-end;
  flex-wrap: wrap;
  gap: 8px;
}

.log-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
}

.section-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.inline-actions--left {
  justify-content: flex-start;
}

.dialog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 0 12px;
}

.data-sources-nav {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
  margin-bottom: 16px;
}

.data-sources-nav__item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 14px 16px;
  border-radius: 14px;
  border: 1px solid #dbe5f3;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  color: #475569;
  text-decoration: none;
  transition: border-color 0.2s ease, box-shadow 0.2s ease, transform 0.2s ease;
}

.data-sources-nav__item:hover {
  border-color: #93c5fd;
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.08);
  transform: translateY(-1px);
}

.data-sources-nav__item.is-active {
  border-color: #2563eb;
  background: linear-gradient(180deg, #eff6ff 0%, #ffffff 100%);
  box-shadow: 0 12px 28px rgba(37, 99, 235, 0.12);
}

.data-sources-nav__label {
  color: #0f172a;
  font-size: 15px;
  font-weight: 700;
}

.data-sources-nav__desc {
  color: #64748b;
  font-size: 12px;
  line-height: 1.6;
}

.data-sources-shell__body {
  display: grid;
  gap: 16px;
}

.truth-summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 12px;
}

.truth-summary-card {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 14px 16px;
  background: #fafafa;
}

.governance-kpi-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(190px, 1fr));
  gap: 12px;
}

.governance-kpi-card {
  min-height: 132px;
  padding: 14px 16px;
  border-radius: 14px;
  border: 1px solid #dbe5f3;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.88);
}

.governance-kpi-card.is-primary {
  border-color: #bfdbfe;
  background: linear-gradient(180deg, #eff6ff 0%, #ffffff 100%);
}

.governance-kpi-card.is-success {
  border-color: #bbf7d0;
  background: linear-gradient(180deg, #f0fdf4 0%, #ffffff 100%);
}

.governance-kpi-card.is-warning {
  border-color: #fde68a;
  background: linear-gradient(180deg, #fff8e6 0%, #ffffff 100%);
}

.governance-kpi-card.is-danger {
  border-color: #fecaca;
  background: linear-gradient(180deg, #fff5f5 0%, #ffffff 100%);
}

.governance-kpi-card.is-gold {
  border-color: #fcd34d;
  background: linear-gradient(180deg, #fffbeb 0%, #ffffff 100%);
}

.governance-kpi-card__title {
  color: #475569;
  font-size: 12px;
}

.governance-kpi-card__value {
  margin-top: 8px;
  color: #0f172a;
  font-size: 28px;
  font-weight: 700;
}

.governance-kpi-card__helper {
  margin-top: 10px;
  color: #64748b;
  font-size: 12px;
  line-height: 1.6;
}

.truth-summary-card__title {
  font-weight: 600;
  margin-bottom: 10px;
}

.truth-summary-card__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.truth-summary-list {
  margin: 0;
  padding-left: 18px;
  line-height: 1.7;
}

.truth-summary-list--compact {
  margin-top: 8px;
}

.issue-quick-filter {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 12px;
}

.issue-quick-filter__label {
  color: #6b7280;
  font-size: 13px;
}

.issue-quick-filter__tag {
  cursor: pointer;
}

.sync-form-grid {
  display: grid;
  grid-template-columns: 1.3fr 1.4fr 120px auto;
  gap: 10px;
  align-items: center;
}

.sync-inline-numbers {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}

.sync-inline-hint {
  margin-top: 10px;
}

.sync-last-result {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed #d1d5db;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sync-last-result__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
  color: #374151;
}

.sync-result-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.payload-message {
  margin: 8px 0 0;
  color: #374151;
  line-height: 1.7;
}

.payload-viewer {
  margin: 8px 0 0;
  padding: 12px;
  background: #111827;
  color: #e5e7eb;
  border-radius: 12px;
  max-height: 420px;
  overflow: auto;
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.6;
}

.governance-table-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(420px, 1fr));
  gap: 12px;
  align-items: start;
}

.governance-table-card {
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  padding: 14px;
  background: linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
}

.dialog-grid .el-form-item {
  margin-bottom: 14px;
}

.dialog-grid .el-select,
.dialog-grid .el-input-number {
  width: 100%;
}

@media (max-width: 1080px) {
  .sync-form-grid {
    grid-template-columns: 1fr;
  }

  .sync-inline-numbers {
    grid-template-columns: 1fr 1fr;
  }

  .governance-table-grid {
    grid-template-columns: 1fr;
  }
}
</style>
