<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import FuturesSelectionModuleShell from "../../components/FuturesSelectionModuleShell.vue";
import {
  createFuturesSelectionTemplate,
  listFuturesSelectionTemplates,
  setDefaultFuturesSelectionTemplate,
  updateFuturesSelectionTemplate
} from "../../api/admin";
import {
  formatFuturesSelectionContractScope,
  formatFuturesSelectionMarketRegime,
  formatFuturesSelectionRiskLevel,
  formatFuturesSelectionStatus,
  formatFuturesSelectionStyle,
  futuresSelectionContractScopeOptions,
  futuresSelectionMarketRegimeOptions,
  futuresSelectionRiskLevelOptions,
  futuresSelectionStatusOptions,
  futuresSelectionStyleOptions,
  joinFuturesSelectionContracts,
  splitFuturesSelectionContracts
} from "../../lib/futures-selection";
import { hasPermission } from "../../lib/session";

const canManage = hasPermission("futures_selection.manage");
const loading = ref(false);
const dialogVisible = ref(false);
const submitting = ref(false);
const editingID = ref("");
const templates = ref([]);

const DEFAULTS = {
  market_regime_bias: "BASE",
  status: "ACTIVE",
  is_default: false,
  style: "balanced",
  contract_scope: "DOMINANT_ALL",
  allow_mock_fallback_on_short_history: true,
  min_confidence: 55,
  limit: 3,
  max_risk_level: "HIGH",
  review_required: true,
  allow_auto_publish: false
};

function numberValue(value, fallback) {
  const numeric = Number(value);
  return Number.isFinite(numeric) ? numeric : fallback;
}

function booleanValue(value, fallback) {
  if (typeof value === "boolean") {
    return value;
  }
  if (value === undefined || value === null || value === "") {
    return fallback;
  }
  return Boolean(value);
}

function createDefaultForm() {
  return {
    template_key: "",
    name: "",
    description: "",
    market_regime_bias: DEFAULTS.market_regime_bias,
    status: DEFAULTS.status,
    is_default: DEFAULTS.is_default,
    style: DEFAULTS.style,
    contract_scope: DEFAULTS.contract_scope,
    allow_mock_fallback_on_short_history: DEFAULTS.allow_mock_fallback_on_short_history,
    contracts_text: "",
    min_confidence: DEFAULTS.min_confidence,
    limit: DEFAULTS.limit,
    max_risk_level: DEFAULTS.max_risk_level,
    review_required: DEFAULTS.review_required,
    allow_auto_publish: DEFAULTS.allow_auto_publish
  };
}

const form = reactive(createDefaultForm());

const manualContractsVisible = computed(
  () => form.contract_scope === "MANUAL" || String(form.contracts_text || "").trim() !== ""
);

function resetForm() {
  editingID.value = "";
  Object.assign(form, createDefaultForm());
}

function openCreate() {
  resetForm();
  dialogVisible.value = true;
}

function openEdit(item) {
  editingID.value = item.id;
  Object.assign(form, {
    template_key: item.template_key || "",
    name: item.name || "",
    description: item.description || "",
    market_regime_bias: item.market_regime_bias || DEFAULTS.market_regime_bias,
    status: item.status || DEFAULTS.status,
    is_default: Boolean(item.is_default),
    style: item.universe_defaults_json?.style || DEFAULTS.style,
    contract_scope: item.universe_defaults_json?.contract_scope || DEFAULTS.contract_scope,
    allow_mock_fallback_on_short_history: booleanValue(
      item.universe_defaults_json?.allow_mock_fallback_on_short_history,
      DEFAULTS.allow_mock_fallback_on_short_history
    ),
    contracts_text: joinFuturesSelectionContracts(item.universe_defaults_json?.contracts),
    min_confidence: numberValue(item.factor_defaults_json?.min_confidence, DEFAULTS.min_confidence),
    limit: numberValue(item.portfolio_defaults_json?.limit, DEFAULTS.limit),
    max_risk_level: item.portfolio_defaults_json?.max_risk_level || DEFAULTS.max_risk_level,
    review_required: booleanValue(item.publish_defaults_json?.review_required, DEFAULTS.review_required),
    allow_auto_publish: booleanValue(
      item.publish_defaults_json?.allow_auto_publish,
      DEFAULTS.allow_auto_publish
    )
  });
  dialogVisible.value = true;
}

function buildPayload() {
  const templateKey = String(form.template_key || "").trim();
  const name = String(form.name || "").trim();
  if (!templateKey) {
    throw new Error("请填写模板标识");
  }
  if (!name) {
    throw new Error("请填写模板名称");
  }

  const contracts = splitFuturesSelectionContracts(form.contracts_text);
  if (form.contract_scope === "MANUAL" && contracts.length === 0) {
    throw new Error("手工指定合约时至少填写 1 个合约代码");
  }

  return {
    template_key: templateKey,
    name,
    description: String(form.description || "").trim(),
    market_regime_bias: form.market_regime_bias,
    is_default: Boolean(form.is_default),
    status: form.status,
    universe_defaults_json: {
      style: form.style,
      contract_scope: form.contract_scope,
      allow_mock_fallback_on_short_history: Boolean(form.allow_mock_fallback_on_short_history),
      ...(contracts.length > 0 ? { contracts } : {})
    },
    factor_defaults_json: {
      min_confidence: numberValue(form.min_confidence, DEFAULTS.min_confidence)
    },
    portfolio_defaults_json: {
      limit: numberValue(form.limit, DEFAULTS.limit),
      max_risk_level: form.max_risk_level
    },
    publish_defaults_json: {
      review_required: Boolean(form.review_required),
      allow_auto_publish: Boolean(form.allow_auto_publish)
    }
  };
}

async function fetchTemplates() {
  loading.value = true;
  try {
    const data = await listFuturesSelectionTemplates({ page: 1, page_size: 100 });
    templates.value = Array.isArray(data?.items) ? data.items : [];
  } catch (error) {
    ElMessage.error(error?.message || "加载期货策略模板失败");
  } finally {
    loading.value = false;
  }
}

async function submitForm() {
  submitting.value = true;
  try {
    const payload = buildPayload();
    if (editingID.value) {
      await updateFuturesSelectionTemplate(editingID.value, payload);
      ElMessage.success("策略模板已更新");
    } else {
      await createFuturesSelectionTemplate(payload);
      ElMessage.success("策略模板已创建");
    }
    dialogVisible.value = false;
    await fetchTemplates();
  } catch (error) {
    ElMessage.error(error?.message || "保存策略模板失败");
  } finally {
    submitting.value = false;
  }
}

async function handleSetDefault(item) {
  try {
    await ElMessageBox.confirm(`确认将「${item.name}」设为默认模板吗？`, "设为默认模板", {
      confirmButtonText: "确认",
      cancelButtonText: "取消",
      type: "warning"
    });
    await setDefaultFuturesSelectionTemplate(item.id);
    ElMessage.success("默认模板已切换");
    await fetchTemplates();
  } catch (error) {
    if (error === "cancel") {
      return;
    }
    ElMessage.error(error?.message || "切换默认模板失败");
  }
}

onMounted(fetchTemplates);
</script>

<template>
  <FuturesSelectionModuleShell
    title="智能期货策略模板"
    description="系统模板和运营自定义模板统一在这里管理。模板只定义默认研究参数和市场状态偏好，不会直接发布。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchTemplates">刷新模板</el-button>
        <el-button v-if="canManage" type="primary" @click="openCreate">新建模板</el-button>
      </div>
    </template>

    <div class="card">
      <el-table :data="templates" border stripe size="small" v-loading="loading" empty-text="暂无期货策略模板">
        <el-table-column prop="name" label="模板名称" min-width="160" />
        <el-table-column prop="template_key" label="模板标识" min-width="160" />
        <el-table-column label="市场偏好" min-width="120">
          <template #default="{ row }">
            {{ formatFuturesSelectionMarketRegime(row.market_regime_bias) }}
          </template>
        </el-table-column>
        <el-table-column label="默认风格" min-width="120">
          <template #default="{ row }">
            {{ formatFuturesSelectionStyle(row.universe_defaults_json?.style) }}
          </template>
        </el-table-column>
        <el-table-column label="合约范围" min-width="160">
          <template #default="{ row }">
            {{ formatFuturesSelectionContractScope(row.universe_defaults_json?.contract_scope) }}
          </template>
        </el-table-column>
        <el-table-column label="组合约束" min-width="220">
          <template #default="{ row }">
            组合 {{ row.portfolio_defaults_json?.limit || DEFAULTS.limit }} 个 /
            {{ formatFuturesSelectionRiskLevel(row.portfolio_defaults_json?.max_risk_level) }} /
            置信度 {{ row.factor_defaults_json?.min_confidence || DEFAULTS.min_confidence }}+
          </template>
        </el-table-column>
        <el-table-column label="默认" min-width="90">
          <template #default="{ row }">
            <el-tag :type="row.is_default ? 'success' : 'info'">{{ row.is_default ? "默认" : "候选" }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="100">
          <template #default="{ row }">
            {{ formatFuturesSelectionStatus(row.status) }}
          </template>
        </el-table-column>
        <el-table-column prop="description" label="说明" min-width="260" show-overflow-tooltip />
        <el-table-column label="操作" min-width="200" fixed="right">
          <template #default="{ row }">
            <div class="row-actions">
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button v-if="canManage" size="small" type="success" @click="handleSetDefault(row)">
                设为默认
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="editingID ? '编辑策略模板' : '新建策略模板'" width="920px">
      <el-form label-position="top" class="template-form">
        <div class="profile-form-grid">
          <el-form-item label="模板标识" required>
            <el-input v-model="form.template_key" placeholder="例如 trend_attack" />
          </el-form-item>
          <el-form-item label="模板名称" required>
            <el-input v-model="form.name" placeholder="例如 趋势进攻" />
          </el-form-item>
          <el-form-item label="市场状态偏好">
            <el-select v-model="form.market_regime_bias">
              <el-option
                v-for="item in futuresSelectionMarketRegimeOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="form.status">
              <el-option
                v-for="item in futuresSelectionStatusOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="模板说明" class="grid-span-2">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="3"
              placeholder="说明这个模板适合什么市场环境、目标风格和风险边界"
            />
          </el-form-item>
        </div>

        <div class="section-card">
          <div class="section-title">合约池规则</div>
          <div class="profile-form-grid">
            <el-form-item label="默认研究风格">
              <el-segmented v-model="form.style" :options="futuresSelectionStyleOptions" />
            </el-form-item>
            <el-form-item label="合约范围">
              <el-select v-model="form.contract_scope">
                <el-option
                  v-for="item in futuresSelectionContractScopeOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="历史不足时允许模拟补齐">
              <el-switch
                v-model="form.allow_mock_fallback_on_short_history"
                active-text="允许"
                inactive-text="关闭"
              />
            </el-form-item>
            <el-form-item label="设为默认建议模板">
              <el-switch v-model="form.is_default" active-text="是" inactive-text="否" />
            </el-form-item>
            <el-form-item v-if="manualContractsVisible" label="手工指定合约" class="grid-span-2">
              <el-input
                v-model="form.contracts_text"
                type="textarea"
                :rows="4"
                placeholder="每行一个合约代码，例如 IF2506 或 CU2505"
              />
            </el-form-item>
          </div>
        </div>

        <div class="section-card">
          <div class="section-title">评分与组合</div>
          <div class="profile-form-grid">
            <el-form-item label="最低置信度">
              <el-slider v-model="form.min_confidence" :min="30" :max="95" :step="1" show-input />
            </el-form-item>
            <el-form-item label="组合数量">
              <el-input-number v-model="form.limit" :min="1" :max="10" />
            </el-form-item>
            <el-form-item label="最高风险级别">
              <el-select v-model="form.max_risk_level">
                <el-option
                  v-for="item in futuresSelectionRiskLevelOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
          </div>
        </div>

        <div class="section-card">
          <div class="section-title">发布规则</div>
          <div class="profile-form-grid">
            <el-form-item label="需要人工审核">
              <el-switch v-model="form.review_required" active-text="需要" inactive-text="不需要" />
            </el-form-item>
            <el-form-item label="允许自动发布">
              <el-switch v-model="form.allow_auto_publish" active-text="允许" inactive-text="关闭" />
            </el-form-item>
          </div>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">保存模板</el-button>
      </template>
    </el-dialog>
  </FuturesSelectionModuleShell>
</template>

<style scoped>
.row-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.profile-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.grid-span-2 {
  grid-column: span 2;
}

.section-card {
  padding: 16px;
  margin-top: 12px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 14px;
  background: var(--el-bg-color-page);
}

.section-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}

@media (max-width: 900px) {
  .profile-form-grid {
    grid-template-columns: minmax(0, 1fr);
  }

  .grid-span-2 {
    grid-column: span 1;
  }
}
</style>
