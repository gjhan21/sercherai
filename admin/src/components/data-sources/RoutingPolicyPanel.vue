<script setup>
import { formatRoutingFallbackProviders } from "../../lib/data-sources-admin.js";

defineProps({
  policies: { type: Array, default: () => [] },
  dialogVisible: Boolean,
  form: { type: Object, required: true },
  providerOptions: { type: Array, default: () => [] },
  submitting: Boolean,
  canEditDataSources: Boolean
});

const emit = defineEmits(["edit", "save", "cancel", "update:dialogVisible"]);

function handleClose() {
  emit("update:dialogVisible", false);
  emit("cancel");
}
</script>

<template>
  <div class="card" style="margin-bottom: 12px">
    <div class="section-header">
      <div>
        <h3 style="margin: 0">路由与真相源治理</h3>
        <p class="muted" style="margin: 6px 0 0">
          在这里直接维护主源、备源、MOCK 开关和质量阈值；同步链、truth 链、context 链都会共用这套策略。
        </p>
      </div>
    </div>
    <el-table :data="policies" border stripe size="small" empty-text="暂无路由策略">
      <el-table-column prop="policy_key" label="策略 Key" min-width="180" />
      <el-table-column prop="asset_class" label="资产" width="90" />
      <el-table-column prop="data_kind" label="数据域" min-width="140" />
      <el-table-column prop="primary_provider_key" label="主源" width="120" />
      <el-table-column label="备源" min-width="180">
        <template #default="{ row }">
          {{ formatRoutingFallbackProviders(row) }}
        </template>
      </el-table-column>
      <el-table-column label="策略" min-width="180">
        <template #default="{ row }">
          <div class="sync-result-tags">
            <el-tag :type="row.fallback_allowed ? 'success' : 'info'">
              {{ row.fallback_allowed ? "允许回退" : "禁止回退" }}
            </el-tag>
            <el-tag :type="row.mock_allowed ? 'warning' : 'info'">
              {{ row.mock_allowed ? "允许 MOCK" : "禁止 MOCK" }}
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <el-button size="small" type="primary" plain :disabled="!canEditDataSources" @click="emit('edit', row)">
            编辑
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog
      :model-value="dialogVisible"
      :title="form.policy_key ? `编辑路由策略：${form.policy_key}` : '编辑路由策略'"
      width="720px"
      destroy-on-close
      @update:model-value="emit('update:dialogVisible', $event)"
      @close="handleClose"
    >
      <el-form label-width="120px">
        <el-form-item label="策略 Key"><el-input v-model="form.policy_key" disabled /></el-form-item>
        <el-form-item label="资产"><el-input v-model="form.asset_class" disabled /></el-form-item>
        <el-form-item label="数据域"><el-input v-model="form.data_kind" disabled /></el-form-item>
        <el-form-item label="主源">
          <el-select v-model="form.primary_provider_key" placeholder="选择主源">
            <el-option v-for="item in providerOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="备源链">
          <el-select v-model="form.fallback_provider_keys" multiple collapse-tags collapse-tags-tooltip placeholder="选择备源链">
            <el-option
              v-for="item in providerOptions"
              :key="`fallback-${item.value}`"
              :label="item.label"
              :value="item.value"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="质量阈值">
          <el-input-number v-model="form.quality_threshold" :min="0" :max="1" :step="0.05" controls-position="right" />
        </el-form-item>
        <el-form-item label="允许回退"><el-switch v-model="form.fallback_allowed" /></el-form-item>
        <el-form-item label="允许 MOCK"><el-switch v-model="form.mock_allowed" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="handleClose">取消</el-button>
        <el-button v-if="canEditDataSources" type="primary" :loading="submitting" @click="emit('save')">
          保存策略
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>
