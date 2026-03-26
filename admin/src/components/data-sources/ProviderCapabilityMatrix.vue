<script setup>
defineProps({
  items: { type: Array, default: () => [] }
});
</script>

<template>
  <div class="card" style="margin-bottom: 12px">
    <div class="section-header">
      <div>
        <h3 style="margin: 0">能力矩阵</h3>
        <p class="muted" style="margin: 6px 0 0">
          当前能力矩阵来自 provider capability schema，帮助判断哪些供应商可参与同步、truth 重建和研究上下文。
        </p>
      </div>
    </div>
    <el-table :data="items" border stripe size="small" empty-text="暂无能力矩阵数据">
      <el-table-column prop="provider_key" label="Provider" min-width="120" />
      <el-table-column prop="asset_class" label="资产" width="90" />
      <el-table-column prop="data_kind" label="数据域" min-width="140" />
      <el-table-column prop="priority_weight" label="优先级" width="90" />
      <el-table-column label="能力" min-width="320">
        <template #default="{ row }">
          <div class="sync-result-tags">
            <el-tag v-if="row.supports_sync" type="success">同步</el-tag>
            <el-tag v-if="row.supports_truth_rebuild" type="warning">truth</el-tag>
            <el-tag v-if="row.supports_context_seed" type="info">context</el-tag>
            <el-tag v-if="row.supports_research_run" type="success">research</el-tag>
            <el-tag v-if="row.supports_metadata_enrichment" type="warning">meta</el-tag>
          </div>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>
