<script setup>
import {
  isDefaultSource,
  normalizeSourceKey,
  statusTagType,
  supportsDefaultStockSource
} from "../../lib/data-sources-admin.js";

defineProps({
  items: { type: Array, default: () => [] },
  loading: Boolean,
  total: Number,
  page: Number,
  pageSize: Number,
  healthMap: { type: Object, default: () => ({}) },
  canEditDataSources: Boolean,
  batchChecking: Boolean,
  settingDefaultSource: Boolean,
  settingDefaultSourceKey: String,
  defaultStockSourceKey: String
});

const emit = defineEmits([
  "refresh",
  "batch-check",
  "create",
  "edit",
  "delete",
  "check-health",
  "show-logs",
  "set-default-stock",
  "change-page"
]);
</script>

<template>
  <div class="card">
    <div class="section-header" style="margin-bottom: 12px">
      <div>
        <h3 style="margin: 0">数据源台账</h3>
        <p class="muted" style="margin: 6px 0 0">
          在这里维护数据源、默认行情源、健康检查入口和基础 CRUD。
        </p>
      </div>
      <div class="toolbar">
        <el-button :loading="loading" @click="emit('refresh')">刷新列表</el-button>
        <el-button
          v-if="canEditDataSources"
          type="primary"
          plain
          :loading="batchChecking"
          @click="emit('batch-check')"
        >
          全部健康检查
        </el-button>
        <el-button v-if="canEditDataSources" type="primary" @click="emit('create')">新增数据源</el-button>
      </div>
    </div>

    <el-table :data="items" border stripe v-loading="loading" empty-text="暂无数据源">
      <el-table-column label="source_key" min-width="190">
        <template #default="{ row }">
          <div class="source-key-cell">
            <span>{{ row.source_key }}</span>
            <el-tag
              v-if="isDefaultSource(row.source_key, defaultStockSourceKey)"
              size="small"
              type="warning"
              effect="plain"
            >
              默认行情源
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="名称" min-width="130" />
      <el-table-column prop="source_type" label="类型" min-width="110" />
      <el-table-column label="状态" min-width="110">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="配置摘要" min-width="260">
        <template #default="{ row }">
          <div class="config-line">endpoint: {{ row.config?.endpoint || "-" }}</div>
          <div class="config-line config-line--muted">token: {{ row.config?.token ? "已配置" : "未配置" }}</div>
          <div class="config-line config-line--muted">
            阈值: {{ row.config?.fail_threshold ?? 3 }} / 重试: {{ row.config?.retry_times ?? 0 }}
          </div>
        </template>
      </el-table-column>
      <el-table-column label="最近健康状态" min-width="260">
        <template #default="{ row }">
          <template v-if="healthMap[row.source_key]">
            <el-tag :type="statusTagType(healthMap[row.source_key].status)">
              {{ healthMap[row.source_key].status }}
            </el-tag>
            <div class="config-line config-line--muted">
              {{ healthMap[row.source_key].message || "-" }} · {{ healthMap[row.source_key].latency_ms || 0 }}ms
            </div>
            <div class="config-line config-line--muted">
              尝试 {{ healthMap[row.source_key].attempts || 1 }}/{{ healthMap[row.source_key].max_attempts || 1 }}
              · 连续失败 {{ healthMap[row.source_key].consecutive_failures || 0 }}
            </div>
          </template>
          <el-text v-else type="info">未检查</el-text>
        </template>
      </el-table-column>
      <el-table-column prop="updated_at" label="更新时间" min-width="180" />
      <el-table-column label="操作" align="right" min-width="420">
        <template #default="{ row }">
          <div class="inline-actions">
            <el-button
              v-if="canEditDataSources && supportsDefaultStockSource(row)"
              size="small"
              type="success"
              plain
              :loading="settingDefaultSource && settingDefaultSourceKey === normalizeSourceKey(row.source_key)"
              :disabled="isDefaultSource(row.source_key, defaultStockSourceKey)"
              @click="emit('set-default-stock', row)"
            >
              设为默认行情源
            </el-button>
            <el-button v-if="canEditDataSources" size="small" @click="emit('check-health', row.source_key)">健康检查</el-button>
            <el-button size="small" @click="emit('show-logs', row.source_key)">健康日志</el-button>
            <el-button v-if="canEditDataSources" size="small" @click="emit('edit', row)">编辑</el-button>
            <el-button v-if="canEditDataSources" size="small" type="danger" plain @click="emit('delete', row.source_key)">删除</el-button>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination">
      <el-text type="info">第 {{ page }} 页，共 {{ total }} 条</el-text>
      <el-pagination
        background
        layout="prev, pager, next"
        :current-page="page"
        :page-size="pageSize"
        :total="total"
        @current-change="emit('change-page', $event)"
      />
    </div>
  </div>
</template>
