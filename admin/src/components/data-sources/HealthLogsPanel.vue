<script setup>
import { statusTagType } from "../../lib/data-sources-admin.js";

defineProps({
  loading: Boolean,
  logSourceKey: String,
  logItems: { type: Array, default: () => [] }
});
</script>

<template>
  <div class="card" style="margin-top: 12px">
    <div class="log-header">
      <h3 style="margin: 0">健康日志 {{ logSourceKey ? `(${logSourceKey})` : "" }}</h3>
      <el-text type="info">{{ loading ? "加载中..." : "最近20条" }}</el-text>
    </div>
    <el-table :data="logItems" border stripe v-loading="loading" empty-text="暂无日志，先执行健康检查">
      <el-table-column label="状态" min-width="120">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="reachable" min-width="110">
        <template #default="{ row }">
          {{ row.reachable ? "是" : "否" }}
        </template>
      </el-table-column>
      <el-table-column prop="http_status" label="HTTP" min-width="90">
        <template #default="{ row }">
          {{ row.http_status || "-" }}
        </template>
      </el-table-column>
      <el-table-column prop="latency_ms" label="latency(ms)" min-width="110" />
      <el-table-column prop="message" label="message" min-width="220">
        <template #default="{ row }">
          {{ row.message || "-" }}
        </template>
      </el-table-column>
      <el-table-column prop="checked_at" label="checked_at" min-width="180" />
    </el-table>
  </div>
</template>
