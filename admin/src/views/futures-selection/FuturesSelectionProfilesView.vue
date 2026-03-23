<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { useRouter } from "vue-router";
import { ElMessage, ElMessageBox } from "element-plus";
import FuturesSelectionModuleShell from "../../components/FuturesSelectionModuleShell.vue";
import {
  createFuturesSelectionProfile,
  listFuturesSelectionProfileVersions,
  listFuturesSelectionProfiles,
  listFuturesSelectionTemplates,
  publishFuturesSelectionProfile,
  rollbackFuturesSelectionProfile,
  updateFuturesSelectionProfile
} from "../../api/admin";
import {
  formatFuturesSelectionContractScope,
  formatFuturesSelectionDateTime,
  formatFuturesSelectionMarketRegime,
  formatFuturesSelectionRiskLevel,
  formatFuturesSelectionStatus,
  formatFuturesSelectionStyle,
  futuresSelectionContractScopeOptions,
  futuresSelectionRiskLevelOptions,
  futuresSelectionStatusOptions,
  futuresSelectionStyleOptions,
  joinFuturesSelectionContracts,
  splitFuturesSelectionContracts
} from "../../lib/futures-selection";
import { hasPermission } from "../../lib/session";

const router = useRouter();
const canManage = hasPermission("futures_selection.manage");
const loading = ref(false);
const submitting = ref(false);
const dialogVisible = ref(false);
const editingID = ref("");
const profiles = ref([]);
const templates = ref([]);

const DEFAULTS = {
  styleDefault: "balanced",
  contractScope: "DOMINANT_ALL",
  status: "ACTIVE",
  universe: {
    allow_mock_fallback_on_short_history: true,
    contracts_text: ""
  },
  factor: {
    min_confidence: 55
  },
  portfolio: {
    limit: 3,
    max_risk_level: "HIGH"
  },
  publish: {
    review_required: true,
    allow_auto_publish: false
  }
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

function buildFormState(profile = null) {
  const universeConfig = profile?.universe_config || {};
  const factorConfig = profile?.factor_config || {};
  const portfolioConfig = profile?.portfolio_config || {};
  const publishConfig = profile?.publish_config || {};
  return {
    name: profile?.name || "",
    template_id: profile?.template_id || "",
    status: profile?.status || DEFAULTS.status,
    is_default: Boolean(profile?.is_default),
    style_default: profile?.style_default || universeConfig.style || DEFAULTS.styleDefault,
    contract_scope: profile?.contract_scope || universeConfig.contract_scope || DEFAULTS.contractScope,
    universe: {
      allow_mock_fallback_on_short_history: booleanValue(
        universeConfig.allow_mock_fallback_on_short_history,
        DEFAULTS.universe.allow_mock_fallback_on_short_history
      ),
      contracts_text: joinFuturesSelectionContracts(universeConfig.contracts)
    },
    factor: {
      min_confidence: numberValue(factorConfig.min_confidence, DEFAULTS.factor.min_confidence)
    },
    portfolio: {
      limit: numberValue(portfolioConfig.limit, DEFAULTS.portfolio.limit),
      max_risk_level: portfolioConfig.max_risk_level || DEFAULTS.portfolio.max_risk_level
    },
    publish: {
      review_required: booleanValue(publishConfig.review_required, DEFAULTS.publish.review_required),
      allow_auto_publish: booleanValue(publishConfig.allow_auto_publish, DEFAULTS.publish.allow_auto_publish)
    },
    description: profile?.description || "",
    change_note: ""
  };
}

const form = reactive(buildFormState());

const selectedTemplate = computed(() =>
  templates.value.find((item) => item.id === form.template_id) || null
);

const defaultTemplate = computed(() => templates.value.find((item) => item.is_default) || templates.value[0] || null);

const manualContractsVisible = computed(
  () => form.contract_scope === "MANUAL" || String(form.universe.contracts_text || "").trim() !== ""
);

function resetForm() {
  editingID.value = "";
  Object.assign(form, buildFormState());
}

function applyTemplateDefaults(template = selectedTemplate.value) {
  if (!template) {
    return;
  }
  const universeDefaults = template.universe_defaults_json || {};
  const factorDefaults = template.factor_defaults_json || {};
  const portfolioDefaults = template.portfolio_defaults_json || {};
  const publishDefaults = template.publish_defaults_json || {};

  form.style_default = universeDefaults.style || form.style_default || DEFAULTS.styleDefault;
  form.contract_scope = universeDefaults.contract_scope || form.contract_scope || DEFAULTS.contractScope;
  form.universe.allow_mock_fallback_on_short_history = booleanValue(
    universeDefaults.allow_mock_fallback_on_short_history,
    form.universe.allow_mock_fallback_on_short_history
  );
  if (Array.isArray(universeDefaults.contracts) && universeDefaults.contracts.length > 0) {
    form.universe.contracts_text = joinFuturesSelectionContracts(universeDefaults.contracts);
  }
  form.factor.min_confidence = numberValue(factorDefaults.min_confidence, form.factor.min_confidence);
  form.portfolio.limit = numberValue(portfolioDefaults.limit, form.portfolio.limit);
  form.portfolio.max_risk_level = portfolioDefaults.max_risk_level || form.portfolio.max_risk_level;
  form.publish.review_required = booleanValue(publishDefaults.review_required, form.publish.review_required);
  form.publish.allow_auto_publish = booleanValue(
    publishDefaults.allow_auto_publish,
    form.publish.allow_auto_publish
  );
}

function openCreate() {
  resetForm();
  if (defaultTemplate.value) {
    form.template_id = defaultTemplate.value.id;
    applyTemplateDefaults(defaultTemplate.value);
  }
  dialogVisible.value = true;
}

function openEdit(profile) {
  editingID.value = profile.id;
  Object.assign(form, buildFormState(profile));
  dialogVisible.value = true;
}

function buildPayload() {
  const name = String(form.name || "").trim();
  if (!name) {
    throw new Error("请填写配置方案名称");
  }

  const contracts = splitFuturesSelectionContracts(form.universe.contracts_text);
  if (form.contract_scope === "MANUAL" && contracts.length === 0) {
    throw new Error("手工指定合约时至少填写 1 个合约代码");
  }

  return {
    name,
    template_id: form.template_id,
    status: form.status,
    is_default: Boolean(form.is_default),
    style_default: form.style_default,
    contract_scope: form.contract_scope,
    universe_config: {
      contract_scope: form.contract_scope,
      style: form.style_default,
      allow_mock_fallback_on_short_history: Boolean(form.universe.allow_mock_fallback_on_short_history),
      ...(contracts.length > 0 ? { contracts } : {})
    },
    factor_config: {
      min_confidence: numberValue(form.factor.min_confidence, DEFAULTS.factor.min_confidence)
    },
    portfolio_config: {
      limit: numberValue(form.portfolio.limit, DEFAULTS.portfolio.limit),
      max_risk_level: form.portfolio.max_risk_level
    },
    publish_config: {
      review_required: Boolean(form.publish.review_required),
      allow_auto_publish: Boolean(form.publish.allow_auto_publish)
    },
    description: String(form.description || "").trim(),
    change_note: String(form.change_note || "").trim()
  };
}

async function fetchProfiles() {
  loading.value = true;
  try {
    const data = await listFuturesSelectionProfiles({ page: 1, page_size: 100 });
    profiles.value = Array.isArray(data?.items) ? data.items : [];
  } catch (error) {
    ElMessage.error(error?.message || "加载期货配置方案失败");
  } finally {
    loading.value = false;
  }
}

async function fetchTemplates() {
  try {
    const data = await listFuturesSelectionTemplates({ page: 1, page_size: 100, status: "ACTIVE" });
    templates.value = Array.isArray(data?.items) ? data.items : [];
  } catch (error) {
    ElMessage.error(error?.message || "加载期货模板失败");
  }
}

async function submitForm() {
  submitting.value = true;
  try {
    const payload = buildPayload();
    if (editingID.value) {
      await updateFuturesSelectionProfile(editingID.value, payload);
      ElMessage.success("配置方案已更新");
    } else {
      await createFuturesSelectionProfile(payload);
      ElMessage.success("配置方案已创建");
    }
    dialogVisible.value = false;
    await fetchProfiles();
  } catch (error) {
    ElMessage.error(error?.message || "保存配置方案失败");
  } finally {
    submitting.value = false;
  }
}

async function handlePublish(profile) {
  try {
    await ElMessageBox.confirm(`确认将「${profile.name}」设为默认期货配置吗？`, "设为默认", {
      confirmButtonText: "确认",
      cancelButtonText: "取消",
      type: "warning"
    });
    await publishFuturesSelectionProfile(profile.id);
    ElMessage.success("已切换默认期货配置");
    await fetchProfiles();
  } catch (error) {
    if (error === "cancel") {
      return;
    }
    ElMessage.error(error?.message || "切换默认配置失败");
  }
}

async function handleRollback(profile) {
  let versions = Array.isArray(profile.versions) ? profile.versions : [];
  try {
    const response = await listFuturesSelectionProfileVersions(profile.id);
    const latestVersions = Array.isArray(response?.items) ? response.items : Array.isArray(response) ? response : [];
    if (latestVersions.length > 0) {
      versions = latestVersions;
    }
  } catch (error) {
    ElMessage.warning(error?.message || "读取期货配置版本列表失败，已回退到当前列表快照");
  }
  const historyVersions = versions.filter((item) => Number(item.version_no) > 0);
  if (historyVersions.length === 0) {
    ElMessage.warning("当前配置没有可回滚的历史版本");
    return;
  }
  const versionTips = historyVersions
    .slice(0, 6)
    .map((item) => `v${item.version_no}：${item.change_note || "无备注"}`)
    .join("\n");
  try {
    const result = await ElMessageBox.prompt(
      `可回滚版本如下：\n${versionTips}\n\n请输入目标版本号`,
      "回滚配置",
      {
        confirmButtonText: "回滚",
        cancelButtonText: "取消",
        inputPattern: /^[0-9]+$/,
        inputErrorMessage: "请输入数字版本号"
      }
    );
    const versionNo = Number(result.value);
    await rollbackFuturesSelectionProfile(profile.id, {
      version_no: versionNo,
      change_note: `从版本 v${versionNo} 回滚`
    });
    ElMessage.success(`已回滚到 v${versionNo} 并生成新版本`);
    await fetchProfiles();
  } catch (error) {
    if (error === "cancel") {
      return;
    }
    ElMessage.error(error?.message || "回滚配置失败");
  }
}

onMounted(async () => {
  await fetchTemplates();
  await fetchProfiles();
});
</script>

<template>
  <FuturesSelectionModuleShell
    title="智能期货策略设计"
    description="这里收口期货研究模板、合约池规则、因子权重和配置方案。先把策略设计清楚，再去运行中心与候选审核页完成发布闭环。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchProfiles">刷新配置</el-button>
        <el-button plain @click="router.push({ name: 'futures-selection-templates' })">策略模板</el-button>
        <el-button plain @click="router.push({ name: 'futures-selection-rules' })">合约池规则</el-button>
        <el-button plain @click="router.push({ name: 'futures-selection-factors' })">因子与权重</el-button>
        <el-button v-if="canManage" type="primary" @click="openCreate">新建配置</el-button>
      </div>
    </template>

    <div class="card">
      <el-table :data="profiles" border stripe size="small" v-loading="loading" empty-text="暂无期货配置方案">
        <el-table-column prop="name" label="配置方案" min-width="180" />
        <el-table-column prop="template_name" label="模板" min-width="140" />
        <el-table-column label="风格" min-width="110">
          <template #default="{ row }">{{ formatFuturesSelectionStyle(row.style_default) }}</template>
        </el-table-column>
        <el-table-column label="合约范围" min-width="130">
          <template #default="{ row }">{{ formatFuturesSelectionContractScope(row.contract_scope) }}</template>
        </el-table-column>
        <el-table-column label="风险上限" min-width="110">
          <template #default="{ row }">{{ formatFuturesSelectionRiskLevel(row.portfolio_config?.max_risk_level) }}</template>
        </el-table-column>
        <el-table-column prop="current_version" label="当前版本" min-width="90">
          <template #default="{ row }">v{{ row.current_version || 0 }}</template>
        </el-table-column>
        <el-table-column label="状态" min-width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 'ACTIVE' ? 'success' : 'info'">
              {{ formatFuturesSelectionStatus(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="默认" min-width="90">
          <template #default="{ row }">
            <el-tag :type="row.is_default ? 'success' : 'info'">{{ row.is_default ? "默认" : "普通" }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="历史版本" min-width="260">
          <template #default="{ row }">
            <div class="version-wrap">
              <el-tag
                v-for="item in (row.versions || []).slice(0, 4)"
                :key="`${row.id}-${item.version_no}`"
                size="small"
                effect="plain"
              >
                v{{ item.version_no }} {{ item.change_note || "" }}
              </el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" min-width="170">
          <template #default="{ row }">{{ formatFuturesSelectionDateTime(row.updated_at) }}</template>
        </el-table-column>
        <el-table-column v-if="canManage" label="操作" min-width="240" fixed="right">
          <template #default="{ row }">
            <div class="action-row">
              <el-button link type="primary" @click="openEdit(row)">编辑</el-button>
              <el-button link type="success" @click="handlePublish(row)">设为默认</el-button>
              <el-button link type="warning" @click="handleRollback(row)">回滚</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="editingID ? '编辑期货配置' : '新建期货配置'"
      width="880px"
      destroy-on-close
    >
      <div class="dialog-grid">
        <div class="card">
          <div class="card-title">基础信息</div>
          <el-form label-width="120px">
            <el-form-item label="配置名称">
              <el-input v-model="form.name" placeholder="例如：均衡趋势日常版" />
            </el-form-item>
            <el-form-item label="研究模板">
              <div class="inline-form">
                <el-select v-model="form.template_id" placeholder="选择模板" style="flex: 1">
                  <el-option
                    v-for="item in templates"
                    :key="item.id"
                    :label="`${item.name}${item.is_default ? '（默认）' : ''}`"
                    :value="item.id"
                  />
                </el-select>
                <el-button @click="applyTemplateDefaults()">套用模板参数</el-button>
              </div>
            </el-form-item>
            <el-form-item label="配置状态">
              <el-select v-model="form.status" style="width: 180px">
                <el-option
                  v-for="item in futuresSelectionStatusOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="保存后设默认">
              <el-switch v-model="form.is_default" />
            </el-form-item>
            <el-form-item label="配置说明">
              <el-input
                v-model="form.description"
                type="textarea"
                :rows="3"
                placeholder="说明这个期货配置适用于什么市场、什么节奏。"
              />
            </el-form-item>
            <el-form-item label="变更备注">
              <el-input
                v-model="form.change_note"
                type="textarea"
                :rows="2"
                placeholder="例如：提高置信度阈值，缩小组合数量。"
              />
            </el-form-item>
          </el-form>
        </div>

        <div class="card" v-if="selectedTemplate">
          <div class="card-title">模板预览</div>
          <el-descriptions :column="1" border size="small">
            <el-descriptions-item label="模板名称">
              {{ selectedTemplate.name }}
            </el-descriptions-item>
            <el-descriptions-item label="偏向市场状态">
              {{ formatFuturesSelectionMarketRegime(selectedTemplate.market_regime_bias) }}
            </el-descriptions-item>
            <el-descriptions-item label="默认风格">
              {{ formatFuturesSelectionStyle(selectedTemplate.universe_defaults_json?.style) }}
            </el-descriptions-item>
            <el-descriptions-item label="默认范围">
              {{ formatFuturesSelectionContractScope(selectedTemplate.universe_defaults_json?.contract_scope) }}
            </el-descriptions-item>
            <el-descriptions-item label="模板说明">
              {{ selectedTemplate.description || "-" }}
            </el-descriptions-item>
          </el-descriptions>
        </div>
      </div>

      <div class="dialog-grid" style="margin-top: 12px">
        <div class="card">
          <div class="card-title">研究参数</div>
          <el-form label-width="120px">
            <el-form-item label="研究风格">
              <el-radio-group v-model="form.style_default">
                <el-radio-button
                  v-for="item in futuresSelectionStyleOptions"
                  :key="item.value"
                  :label="item.value"
                >
                  {{ item.label }}
                </el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="合约范围">
              <el-radio-group v-model="form.contract_scope">
                <el-radio-button
                  v-for="item in futuresSelectionContractScopeOptions"
                  :key="item.value"
                  :label="item.value"
                >
                  {{ item.label }}
                </el-radio-button>
              </el-radio-group>
            </el-form-item>
            <el-form-item label="手工合约列表" v-if="manualContractsVisible">
              <el-input
                v-model="form.universe.contracts_text"
                type="textarea"
                :rows="4"
                placeholder="每行一个，例如：IF2606&#10;RB2605&#10;AU2606"
              />
            </el-form-item>
            <el-form-item label="短历史兜底">
              <el-switch v-model="form.universe.allow_mock_fallback_on_short_history" />
            </el-form-item>
            <el-form-item label="最低置信度">
              <div class="inline-form">
                <el-slider
                  v-model="form.factor.min_confidence"
                  :min="0"
                  :max="100"
                  :step="1"
                  style="flex: 1"
                />
                <el-input-number v-model="form.factor.min_confidence" :min="0" :max="100" :step="1" />
              </div>
            </el-form-item>
            <el-form-item label="组合数量">
              <el-input-number v-model="form.portfolio.limit" :min="1" :max="5" :step="1" />
            </el-form-item>
            <el-form-item label="风险上限">
              <el-radio-group v-model="form.portfolio.max_risk_level">
                <el-radio-button
                  v-for="item in futuresSelectionRiskLevelOptions"
                  :key="item.value"
                  :label="item.value"
                >
                  {{ item.label }}
                </el-radio-button>
              </el-radio-group>
            </el-form-item>
          </el-form>
        </div>

        <div class="card">
          <div class="card-title">发布规则</div>
          <el-form label-width="120px">
            <el-form-item label="需要审核">
              <el-switch v-model="form.publish.review_required" />
            </el-form-item>
            <el-form-item label="允许自动发布">
              <el-switch v-model="form.publish.allow_auto_publish" />
            </el-form-item>
            <el-form-item label="当前摘要">
              <el-alert
                type="info"
                :closable="false"
                show-icon
                :title="`风格：${formatFuturesSelectionStyle(form.style_default)}；范围：${formatFuturesSelectionContractScope(form.contract_scope)}；最低置信度：${form.factor.min_confidence}`"
              />
            </el-form-item>
          </el-form>
        </div>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="submitting" @click="submitForm">保存配置</el-button>
        </div>
      </template>
    </el-dialog>
  </FuturesSelectionModuleShell>
</template>

<style scoped>
.dialog-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.card-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}

.inline-form {
  width: 100%;
  display: flex;
  gap: 12px;
  align-items: center;
}

.version-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.action-row {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 960px) {
  .dialog-grid {
    grid-template-columns: 1fr;
  }
}
</style>
