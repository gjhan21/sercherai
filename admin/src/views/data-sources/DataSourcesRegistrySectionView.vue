<script setup>
import DataSourceRegistryTable from "../../components/data-sources/DataSourceRegistryTable.vue";
import HealthLogsPanel from "../../components/data-sources/HealthLogsPanel.vue";
import { useDataSourcesWorkspaceContext } from "../../composables/useDataSourcesWorkspace.js";

const workspace = useDataSourcesWorkspaceContext();
</script>

<template>
  <DataSourceRegistryTable
    :items="workspace.items"
    :loading="workspace.loading"
    :total="workspace.total"
    :page="workspace.page"
    :page-size="workspace.pageSize"
    :health-map="workspace.healthMap"
    :can-edit-data-sources="workspace.canEditDataSources"
    :batch-checking="workspace.batchChecking"
    :setting-default-source="workspace.settingDefaultSource"
    :setting-default-source-key="workspace.settingDefaultSourceKey"
    :default-stock-source-key="workspace.defaultStockSourceKey"
    @refresh="workspace.fetchDataSources"
    @batch-check="workspace.handleBatchCheckAll"
    @create="workspace.handleCreate"
    @edit="workspace.handleEdit"
    @delete="workspace.handleDelete"
    @check-health="workspace.handleCheckOne"
    @show-logs="workspace.showLogs"
    @set-default-stock="workspace.handleSetDefaultStockSource"
    @change-page="workspace.handlePageChange"
  />

  <HealthLogsPanel
    :loading="workspace.logsLoading"
    :log-source-key="workspace.logSourceKey"
    :log-items="workspace.logItems"
  />

  <el-dialog
    v-model="workspace.formVisible"
    :title="workspace.formMode === 'create' ? '新增数据源' : `编辑数据源：${workspace.form.source_key}`"
    width="860px"
    destroy-on-close
  >
    <el-form label-width="120px">
      <div class="dialog-grid">
        <el-form-item label="source_key" required>
          <el-input v-model="workspace.form.source_key" :disabled="workspace.formMode === 'edit'" placeholder="wind" />
        </el-form-item>
        <el-form-item label="名称" required>
          <el-input v-model="workspace.form.name" placeholder="Wind 数据源" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-select v-model="workspace.form.source_type">
            <el-option v-for="item in workspace.sourceTypeOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="workspace.form.status">
            <el-option v-for="item in workspace.statusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="endpoint">
          <el-input v-model="workspace.form.endpoint" placeholder="Tushare: https://api.tushare.pro" />
        </el-form-item>
        <el-form-item label="token">
          <el-input
            v-model="workspace.form.token"
            type="password"
            show-password
            placeholder="Tushare Token（可留空使用后端环境变量）"
          />
        </el-form-item>
        <el-form-item label="fail_threshold">
          <el-input-number v-model="workspace.form.fail_threshold" :min="1" :step="1" controls-position="right" />
        </el-form-item>
        <el-form-item label="retry_times">
          <el-input-number v-model="workspace.form.retry_times" :min="0" :max="5" :step="1" controls-position="right" />
        </el-form-item>
        <el-form-item label="retry_interval_ms">
          <el-input-number v-model="workspace.form.retry_interval_ms" :min="0" :step="100" controls-position="right" />
        </el-form-item>
        <el-form-item label="health_timeout_ms">
          <el-input-number v-model="workspace.form.health_timeout_ms" :min="500" :step="100" controls-position="right" />
        </el-form-item>
        <el-form-item label="alert_receiver_id">
          <el-input v-model="workspace.form.alert_receiver_id" placeholder="admin_001" />
        </el-form-item>
      </div>
    </el-form>

    <template #footer>
      <el-button @click="workspace.closeForm">取消</el-button>
      <el-button v-if="workspace.canEditDataSources" type="primary" :loading="workspace.submitting" @click="workspace.submitForm">
        提交
      </el-button>
    </template>
  </el-dialog>
</template>
