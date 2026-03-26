<script setup>
import { nextTick, onMounted, reactive, ref } from "vue";
import StrategyEngineJobCenter from "./StrategyEngineJobCenter.vue";
import {
  createStrategyAgentProfile,
  createStrategyPublishPolicy,
  createStrategyScenarioTemplate,
  createStrategySeedSet,
  listStrategyAgentProfiles,
  listStrategyPublishPolicies,
  listStrategyScenarioTemplates,
  listStrategySeedSets,
  updateStrategyPublishPolicy,
  updateStrategyScenarioTemplate,
  updateStrategyAgentProfile,
  updateStrategySeedSet
} from "../api/admin";
import { hasPermission } from "../lib/session";

const loading = ref(false);
const seedSubmitting = ref(false);
const agentSubmitting = ref(false);
const scenarioSubmitting = ref(false);
const policySubmitting = ref(false);
const message = ref("");
const errorMessage = ref("");
const seedSets = ref([]);
const agentProfiles = ref([]);
const scenarioTemplates = ref([]);
const publishPolicies = ref([]);
const seedSetTableRef = ref(null);
const agentProfileTableRef = ref(null);
const scenarioTemplateTableRef = ref(null);
const policyTableRef = ref(null);
const focusedConfigType = ref("");
const focusedConfigID = ref("");
const jobCenterRef = ref(null);
const seedDialogVisible = ref(false);
const agentDialogVisible = ref(false);
const scenarioDialogVisible = ref(false);
const policyDialogVisible = ref(false);
const editingSeedID = ref("");
const editingAgentID = ref("");
const editingScenarioID = ref("");
const editingPolicyID = ref("");
const canEditMarket = hasPermission("market.edit");

const agentOptions = [
  { label: "趋势", value: "trend" },
  { label: "事件", value: "event" },
  { label: "流动性", value: "liquidity" },
  { label: "风控", value: "risk" },
  { label: "基差", value: "basis" }
];

const scenarioOrder = ["bull", "base", "bear", "shock"];
const scenarioLabels = {
  bull: "Bull",
  base: "Base",
  bear: "Bear",
  shock: "Shock"
};

const seedForm = reactive({
  name: "",
  target_type: "STOCK",
  status: "ACTIVE",
  is_default: false,
  itemsText: "",
  description: ""
});

const agentForm = reactive({
  name: "",
  target_type: "ALL",
  status: "ACTIVE",
  is_default: false,
  enabled_agents: ["trend", "event", "liquidity", "risk", "basis"],
  positive_threshold: 3,
  negative_threshold: 2,
  allow_veto: true,
  allow_mock_fallback_on_short_history: false,
  description: ""
});

const scenarioForm = reactive({
  name: "",
  target_type: "STOCK",
  status: "ACTIVE",
  is_default: false,
  items: [],
  description: ""
});

const policyForm = reactive({
  name: "",
  target_type: "ALL",
  status: "ACTIVE",
  is_default: false,
  max_risk_level: "MEDIUM",
  max_warning_count: 3,
  allow_vetoed_publish: false,
  default_publisher: "strategy-engine",
  override_note_template: "",
  description: ""
});

function clearMessages() {
  message.value = "";
  errorMessage.value = "";
}

function ensureCanEditMarket() {
  if (canEditMarket) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法修改策略引擎配置";
  return false;
}

function normalizeItemsText(items = []) {
  return (items || []).join("\n");
}

function parseItemsText(raw) {
  return Array.from(
    new Set(
      String(raw || "")
        .split(/[,\n]/)
        .map((item) => item.trim().toUpperCase())
        .filter(Boolean)
    )
  );
}

function buildScenarioPreset(targetType = "STOCK") {
  if (targetType === "FUTURES") {
    return [
      { scenario: "bull", label: "顺势扩张", thesis_template: "趋势顺畅推进，主方向胜率抬升。", action: "顺势推进", risk_signal: "中", score_bias: 0 },
      { scenario: "base", label: "计划执行", thesis_template: "常态波动下仍可按计划执行。", action: "按计划执行", risk_signal: "中", score_bias: 0 },
      { scenario: "bear", label: "快速收缩", thesis_template: "反向波动会压缩盈亏比，需快速收缩。", action: "收缩仓位", risk_signal: "中高", score_bias: 0 },
      { scenario: "shock", label: "止损降杠杆", thesis_template: "高波动冲击时先执行止损与降杠杆。", action: "停止新增", risk_signal: "高", score_bias: 0 }
    ];
  }
  return [
    { scenario: "bull", label: "进攻", thesis_template: "景气扩散与资金跟随，趋势继续强化。", action: "加仓", risk_signal: "低", score_bias: 0 },
    { scenario: "base", label: "常态", thesis_template: "维持当前节奏，等待下一轮验证。", action: "持有", risk_signal: "中", score_bias: 0 },
    { scenario: "bear", label: "收缩", thesis_template: "市场回撤导致估值与情绪压缩。", action: "减仓", risk_signal: "中高", score_bias: 0 },
    { scenario: "shock", label: "防守", thesis_template: "黑天鹅或流动性冲击下先保命。", action: "回避", risk_signal: "高", score_bias: 0 }
  ];
}

function normalizeScenarioItems(items = [], targetType = "STOCK") {
  const base = buildScenarioPreset(targetType);
  const itemMap = new Map((items || []).map((item) => [String(item.scenario || "").toLowerCase(), item]));
  return scenarioOrder.map((scenario) => {
    const preset = base.find((item) => item.scenario === scenario);
    const current = itemMap.get(scenario) || {};
    return {
      scenario,
      label: current.label || preset?.label || scenarioLabels[scenario] || scenario,
      thesis_template: current.thesis_template || preset?.thesis_template || "",
      action: current.action || preset?.action || "",
      risk_signal: current.risk_signal || preset?.risk_signal || "",
      score_bias: Number(current.score_bias ?? preset?.score_bias ?? 0)
    };
  });
}

function resetSeedForm() {
  editingSeedID.value = "";
  Object.assign(seedForm, {
    name: "",
    target_type: "STOCK",
    status: "ACTIVE",
    is_default: false,
    itemsText: "",
    description: ""
  });
}

function resetAgentForm() {
  editingAgentID.value = "";
  Object.assign(agentForm, {
    name: "",
    target_type: "ALL",
    status: "ACTIVE",
    is_default: false,
    enabled_agents: ["trend", "event", "liquidity", "risk", "basis"],
    positive_threshold: 3,
    negative_threshold: 2,
    allow_veto: true,
    allow_mock_fallback_on_short_history: false,
    description: ""
  });
}

function resetScenarioForm() {
  editingScenarioID.value = "";
  Object.assign(scenarioForm, {
    name: "",
    target_type: "STOCK",
    status: "ACTIVE",
    is_default: false,
    items: normalizeScenarioItems([], "STOCK"),
    description: ""
  });
}

function resetPolicyForm() {
  editingPolicyID.value = "";
  Object.assign(policyForm, {
    name: "",
    target_type: "ALL",
    status: "ACTIVE",
    is_default: false,
    max_risk_level: "MEDIUM",
    max_warning_count: 3,
    allow_vetoed_publish: false,
    default_publisher: "strategy-engine",
    override_note_template: "",
    description: ""
  });
}

function formatPolicySource(row = {}) {
  const id = String(row.id || "");
  const updatedBy = String(row.updated_by || "");
  if (id === "policy_default_futures" || id === "policy_default_all" || updatedBy.includes("system")) {
    return "系统默认";
  }
  return "后台配置";
}

function policySourceTagType(row = {}) {
  return formatPolicySource(row) === "系统默认" ? "warning" : "info";
}

function openCreateSeedDialog() {
  if (!ensureCanEditMarket()) {
    return;
  }
  resetSeedForm();
  seedDialogVisible.value = true;
}

function openEditSeedDialog(row) {
  if (!ensureCanEditMarket()) {
    return;
  }
  editingSeedID.value = row.id || "";
  Object.assign(seedForm, {
    name: row.name || "",
    target_type: row.target_type || "STOCK",
    status: row.status || "ACTIVE",
    is_default: !!row.is_default,
    itemsText: normalizeItemsText(row.items),
    description: row.description || ""
  });
  seedDialogVisible.value = true;
}

function openCreateAgentDialog() {
  if (!ensureCanEditMarket()) {
    return;
  }
  resetAgentForm();
  agentDialogVisible.value = true;
}

function openEditAgentDialog(row) {
  if (!ensureCanEditMarket()) {
    return;
  }
  editingAgentID.value = row.id || "";
  Object.assign(agentForm, {
    name: row.name || "",
    target_type: row.target_type || "ALL",
    status: row.status || "ACTIVE",
    is_default: !!row.is_default,
    enabled_agents: Array.isArray(row.enabled_agents) && row.enabled_agents.length
      ? [...row.enabled_agents]
      : ["trend", "event", "liquidity", "risk", "basis"],
    positive_threshold: row.positive_threshold || 3,
    negative_threshold: row.negative_threshold || 2,
    allow_veto: row.allow_veto !== false,
    allow_mock_fallback_on_short_history: !!row.allow_mock_fallback_on_short_history,
    description: row.description || ""
  });
  agentDialogVisible.value = true;
}

function handleScenarioTargetChange(targetType) {
  scenarioForm.items = normalizeScenarioItems(scenarioForm.items, targetType);
}

function openCreateScenarioDialog() {
  if (!ensureCanEditMarket()) {
    return;
  }
  resetScenarioForm();
  scenarioDialogVisible.value = true;
}

function openEditScenarioDialog(row) {
  if (!ensureCanEditMarket()) {
    return;
  }
  editingScenarioID.value = row.id || "";
  Object.assign(scenarioForm, {
    name: row.name || "",
    target_type: row.target_type || "STOCK",
    status: row.status || "ACTIVE",
    is_default: !!row.is_default,
    items: normalizeScenarioItems(row.items, row.target_type || "STOCK"),
    description: row.description || ""
  });
  scenarioDialogVisible.value = true;
}

function openCreatePolicyDialog() {
  if (!ensureCanEditMarket()) {
    return;
  }
  resetPolicyForm();
  policyDialogVisible.value = true;
}

function openEditPolicyDialog(row) {
  if (!ensureCanEditMarket()) {
    return;
  }
  editingPolicyID.value = row.id || "";
  Object.assign(policyForm, {
    name: row.name || "",
    target_type: row.target_type || "ALL",
    status: row.status || "ACTIVE",
    is_default: !!row.is_default,
    max_risk_level: row.max_risk_level || "MEDIUM",
    max_warning_count: Number(row.max_warning_count ?? 3),
    allow_vetoed_publish: !!row.allow_vetoed_publish,
    default_publisher: row.default_publisher || "strategy-engine",
    override_note_template: row.override_note_template || "",
    description: row.description || ""
  });
  policyDialogVisible.value = true;
}


function normalizeStrategyConfigRouteType(value) {
  const normalized = String(value || "").trim().toUpperCase().replace(/[-\s]+/g, "_");
  if (!normalized) {
    return "";
  }
  if (normalized === "SEED_SET" || normalized === "STRATEGY_SEED_SET") {
    return "seed-set";
  }
  if (normalized === "AGENT_PROFILE" || normalized === "STRATEGY_AGENT_PROFILE") {
    return "agent-profile";
  }
  if (normalized === "SCENARIO_TEMPLATE" || normalized === "STRATEGY_SCENARIO_TEMPLATE") {
    return "scenario-template";
  }
  if (normalized === "PUBLISH_POLICY" || normalized === "STRATEGY_PUBLISH_POLICY") {
    return "publish-policy";
  }
  return "";
}

function buildConfigRowClassName(type) {
  return ({ row }) =>
    row?.id && focusedConfigType.value === type && row.id === focusedConfigID.value ? "is-route-focus" : "";
}

const seedSetRowClassName = buildConfigRowClassName("seed-set");
const agentProfileRowClassName = buildConfigRowClassName("agent-profile");
const scenarioTemplateRowClassName = buildConfigRowClassName("scenario-template");
const policyRowClassName = buildConfigRowClassName("publish-policy");

function resolveStrategyConfigFocusSpec(type) {
  switch (normalizeStrategyConfigRouteType(type)) {
    case "seed-set":
      return {
        type: "seed-set",
        label: "种子集",
        items: seedSets.value,
        tableRef: seedSetTableRef.value,
        openEdit: openEditSeedDialog
      };
    case "agent-profile":
      return {
        type: "agent-profile",
        label: "角色配置",
        items: agentProfiles.value,
        tableRef: agentProfileTableRef.value,
        openEdit: openEditAgentDialog
      };
    case "scenario-template":
      return {
        type: "scenario-template",
        label: "场景模板",
        items: scenarioTemplates.value,
        tableRef: scenarioTemplateTableRef.value,
        openEdit: openEditScenarioDialog
      };
    case "publish-policy":
      return {
        type: "publish-policy",
        label: "发布策略",
        items: publishPolicies.value,
        tableRef: policyTableRef.value,
        openEdit: openEditPolicyDialog
      };
    default:
      return null;
  }
}

async function focusStrategyConfigItem(configType, configID) {
  const targetType = normalizeStrategyConfigRouteType(configType);
  const targetID = String(configID || "").trim();
  if (!targetType || !targetID) {
    focusedConfigType.value = "";
    focusedConfigID.value = "";
    return false;
  }
  let spec = resolveStrategyConfigFocusSpec(targetType);
  if (!spec?.items?.some((item) => item?.id === targetID)) {
    await refreshAll();
    spec = resolveStrategyConfigFocusSpec(targetType);
  }
  const matched = spec?.items?.find((item) => item?.id === targetID);
  if (!matched) {
    errorMessage.value = `未找到${spec?.label || "策略配置"} ${targetID}`;
    return false;
  }
  focusedConfigType.value = targetType;
  focusedConfigID.value = targetID;
  message.value = `已定位到${spec.label} ${matched.name || targetID}`;
  await nextTick();
  const table = spec.tableRef?.$el || spec.tableRef;
  const row = table?.querySelector?.(`tr[data-row-key="${targetID}"]`);
  row?.scrollIntoView?.({ block: "center", behavior: "smooth" });
  if (canEditMarket) {
    spec.openEdit?.(matched);
  }
  return true;
}

async function focusPublishPolicyByID(policyID) {
  return focusStrategyConfigItem("publish-policy", policyID);
}

async function refreshAll() {
  loading.value = true;
  clearMessages();
  try {
    const [seedData, agentData, scenarioData, policyData] = await Promise.all([
      listStrategySeedSets({ page: 1, page_size: 100 }),
      listStrategyAgentProfiles({ page: 1, page_size: 100 }),
      listStrategyScenarioTemplates({ page: 1, page_size: 100 }),
      listStrategyPublishPolicies({ page: 1, page_size: 100 })
    ]);
    seedSets.value = seedData.items || [];
    agentProfiles.value = agentData.items || [];
    scenarioTemplates.value = scenarioData.items || [];
    publishPolicies.value = policyData.items || [];
    await jobCenterRef.value?.refreshJobs?.();
  } catch (error) {
    errorMessage.value = error?.message || "加载策略引擎配置失败";
  } finally {
    loading.value = false;
  }
}

async function handleSubmitSeed() {
  if (!ensureCanEditMarket()) {
    return;
  }
  const items = parseItemsText(seedForm.itemsText);
  if (!seedForm.name.trim()) {
    errorMessage.value = "请输入种子集名称";
    return;
  }
  if (!items.length) {
    errorMessage.value = "请至少输入一个种子标的/合约";
    return;
  }
  seedSubmitting.value = true;
  clearMessages();
  try {
    const payload = {
      name: seedForm.name.trim(),
      target_type: seedForm.target_type,
      status: seedForm.status,
      is_default: seedForm.is_default,
      items,
      description: seedForm.description.trim()
    };
    if (editingSeedID.value) {
      await updateStrategySeedSet(editingSeedID.value, payload);
      message.value = "种子集已更新";
    } else {
      await createStrategySeedSet(payload);
      message.value = "种子集已创建";
    }
    seedDialogVisible.value = false;
    resetSeedForm();
    await refreshAll();
  } catch (error) {
    errorMessage.value = error?.message || "保存种子集失败";
  } finally {
    seedSubmitting.value = false;
  }
}

async function handleSubmitAgent() {
  if (!ensureCanEditMarket()) {
    return;
  }
  if (!agentForm.name.trim()) {
    errorMessage.value = "请输入角色配置名称";
    return;
  }
  if (!agentForm.enabled_agents.length) {
    errorMessage.value = "请至少启用一个角色";
    return;
  }
  agentSubmitting.value = true;
  clearMessages();
  try {
    const payload = {
      name: agentForm.name.trim(),
      target_type: agentForm.target_type,
      status: agentForm.status,
      is_default: agentForm.is_default,
      enabled_agents: agentForm.enabled_agents,
      positive_threshold: Number(agentForm.positive_threshold) || 3,
      negative_threshold: Number(agentForm.negative_threshold) || 2,
      allow_veto: !!agentForm.allow_veto,
      allow_mock_fallback_on_short_history: !!agentForm.allow_mock_fallback_on_short_history,
      description: agentForm.description.trim()
    };
    if (editingAgentID.value) {
      await updateStrategyAgentProfile(editingAgentID.value, payload);
      message.value = "角色配置已更新";
    } else {
      await createStrategyAgentProfile(payload);
      message.value = "角色配置已创建";
    }
    agentDialogVisible.value = false;
    resetAgentForm();
    await refreshAll();
  } catch (error) {
    errorMessage.value = error?.message || "保存角色配置失败";
  } finally {
    agentSubmitting.value = false;
  }
}

async function handleSubmitScenario() {
  if (!ensureCanEditMarket()) {
    return;
  }
  if (!scenarioForm.name.trim()) {
    errorMessage.value = "请输入场景模板名称";
    return;
  }
  scenarioSubmitting.value = true;
  clearMessages();
  try {
    const payload = {
      name: scenarioForm.name.trim(),
      target_type: scenarioForm.target_type,
      status: scenarioForm.status,
      is_default: scenarioForm.is_default,
      items: normalizeScenarioItems(scenarioForm.items, scenarioForm.target_type).map((item) => ({
        scenario: item.scenario,
        label: item.label.trim(),
        thesis_template: item.thesis_template.trim(),
        action: item.action.trim(),
        risk_signal: item.risk_signal.trim(),
        score_bias: Number(item.score_bias || 0)
      })),
      description: scenarioForm.description.trim()
    };
    if (editingScenarioID.value) {
      await updateStrategyScenarioTemplate(editingScenarioID.value, payload);
      message.value = "场景模板已更新";
    } else {
      await createStrategyScenarioTemplate(payload);
      message.value = "场景模板已创建";
    }
    scenarioDialogVisible.value = false;
    resetScenarioForm();
    await refreshAll();
  } catch (error) {
    errorMessage.value = error?.message || "保存场景模板失败";
  } finally {
    scenarioSubmitting.value = false;
  }
}

async function handleSubmitPolicy() {
  if (!ensureCanEditMarket()) {
    return;
  }
  if (!policyForm.name.trim()) {
    errorMessage.value = "请输入发布策略名称";
    return;
  }
  policySubmitting.value = true;
  clearMessages();
  try {
    const payload = {
      name: policyForm.name.trim(),
      target_type: policyForm.target_type,
      status: policyForm.status,
      is_default: policyForm.is_default,
      max_risk_level: policyForm.max_risk_level,
      max_warning_count: Number(policyForm.max_warning_count) || 0,
      allow_vetoed_publish: !!policyForm.allow_vetoed_publish,
      default_publisher: policyForm.default_publisher.trim(),
      override_note_template: policyForm.override_note_template.trim(),
      description: policyForm.description.trim()
    };
    if (editingPolicyID.value) {
      await updateStrategyPublishPolicy(editingPolicyID.value, payload);
      message.value = "发布策略已更新";
    } else {
      await createStrategyPublishPolicy(payload);
      message.value = "发布策略已创建";
    }
    policyDialogVisible.value = false;
    resetPolicyForm();
    await refreshAll();
  } catch (error) {
    errorMessage.value = error?.message || "保存发布策略失败";
  } finally {
    policySubmitting.value = false;
  }
}

function formatItems(items = []) {
  if (!Array.isArray(items) || !items.length) {
    return "-";
  }
  return items.join(" / ");
}

function formatScenarioSummary(items = []) {
  if (!Array.isArray(items) || !items.length) {
    return "-";
  }
  return items.map((item) => `${scenarioLabels[item.scenario] || item.scenario}:${item.action}`).join(" / ");
}

onMounted(refreshAll);

defineExpose({ refreshAll, focusPublishPolicyByID, focusStrategyConfigItem });
</script>

<template>
  <div class="strategy-config-panel" v-loading="loading">
    <el-alert v-if="errorMessage" :title="errorMessage" type="error" show-icon class="panel-alert" />
    <el-alert v-if="message" :title="message" type="success" show-icon class="panel-alert" />

    <div class="config-grid">
      <section class="card">
        <div class="toolbar panel-head">
          <div>
            <h3>种子集管理</h3>
            <p class="muted">让后台决定默认股票 / 期货输入池，生成任务时自动带入。</p>
          </div>
          <el-button v-if="canEditMarket" type="primary" @click="openCreateSeedDialog">新建种子集</el-button>
        </div>
        <el-table ref="seedSetTableRef" :data="seedSets" row-key="id" :row-class-name="seedSetRowClassName" border>
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="target_type" label="类型" width="100" />
          <el-table-column label="默认" width="80">
            <template #default="{ row }">
              <el-tag :type="row.is_default ? 'success' : 'info'">{{ row.is_default ? "默认" : "候选" }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" />
          <el-table-column label="种子项" min-width="220">
            <template #default="{ row }">{{ formatItems(row.items) }}</template>
          </el-table-column>
          <el-table-column prop="updated_at" label="更新时间" min-width="180" />
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }">
              <el-button v-if="canEditMarket" link type="primary" @click="openEditSeedDialog(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="card">
        <div class="toolbar panel-head">
          <div>
            <h3>多角色配置</h3>
            <p class="muted">后台决定启用哪些角色、正负阈值和 veto 策略，生成任务时自动带入。</p>
          </div>
          <el-button v-if="canEditMarket" type="primary" @click="openCreateAgentDialog">新建角色配置</el-button>
        </div>
        <el-table ref="agentProfileTableRef" :data="agentProfiles" row-key="id" :row-class-name="agentProfileRowClassName" border>
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="target_type" label="作用范围" width="110" />
          <el-table-column label="默认" width="80">
            <template #default="{ row }">
              <el-tag :type="row.is_default ? 'success' : 'info'">{{ row.is_default ? "默认" : "候选" }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" />
          <el-table-column label="启用角色" min-width="220">
            <template #default="{ row }">
              <el-space wrap>
                <el-tag v-for="item in row.enabled_agents || []" :key="item" size="small">{{ item }}</el-tag>
              </el-space>
            </template>
          </el-table-column>
          <el-table-column label="阈值" width="140">
            <template #default="{ row }">+{{ row.positive_threshold }} / -{{ row.negative_threshold }}</template>
          </el-table-column>
          <el-table-column label="Veto" width="90">
            <template #default="{ row }">
              <el-tag :type="row.allow_veto ? 'danger' : 'info'">{{ row.allow_veto ? "开启" : "关闭" }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="短历史回退" width="120">
            <template #default="{ row }">
              <el-tag :type="row.allow_mock_fallback_on_short_history ? 'warning' : 'info'">{{ row.allow_mock_fallback_on_short_history ? "允许" : "关闭" }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="updated_at" label="更新时间" min-width="180" />
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }">
              <el-button v-if="canEditMarket" link type="primary" @click="openEditAgentDialog(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="card">
        <div class="toolbar panel-head">
          <div>
            <h3>场景模板</h3>
            <p class="muted">后台管理 bull / base / bear / shock 的解释文案、动作建议和分数偏置。</p>
          </div>
          <el-button v-if="canEditMarket" type="primary" @click="openCreateScenarioDialog">新建场景模板</el-button>
        </div>
        <el-table ref="scenarioTemplateTableRef" :data="scenarioTemplates" row-key="id" :row-class-name="scenarioTemplateRowClassName" border>
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="target_type" label="作用范围" width="110" />
          <el-table-column label="默认" width="80">
            <template #default="{ row }">
              <el-tag :type="row.is_default ? 'success' : 'info'">{{ row.is_default ? "默认" : "候选" }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" />
          <el-table-column label="场景摘要" min-width="260">
            <template #default="{ row }">{{ formatScenarioSummary(row.items) }}</template>
          </el-table-column>
          <el-table-column prop="updated_at" label="更新时间" min-width="180" />
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }">
              <el-button v-if="canEditMarket" link type="primary" @click="openEditScenarioDialog(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="card">
        <div class="toolbar panel-head">
          <div>
            <h3>发布策略</h3>
            <p class="muted">统一管理最大风险等级、警告阈值、veto 放行规则和人工覆盖提示。</p>
          </div>
          <el-button v-if="canEditMarket" type="primary" @click="openCreatePolicyDialog">新建发布策略</el-button>
        </div>
        <el-table ref="policyTableRef" :data="publishPolicies" row-key="id" :row-class-name="policyRowClassName" border>
          <el-table-column prop="name" label="名称" min-width="160" />
          <el-table-column prop="target_type" label="作用范围" width="110" />
          <el-table-column label="默认" width="80">
            <template #default="{ row }">
              <el-tag :type="row.is_default ? 'success' : 'info'">{{ row.is_default ? "默认" : "候选" }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="来源" width="110">
            <template #default="{ row }">
              <el-tag :type="policySourceTagType(row)">{{ formatPolicySource(row) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100" />
          <el-table-column label="风险门槛" width="150">
            <template #default="{ row }">{{ row.max_risk_level }} / 警告≤{{ row.max_warning_count }}</template>
          </el-table-column>
          <el-table-column label="Veto 放行" width="100">
            <template #default="{ row }">
              <el-tag :type="row.allow_vetoed_publish ? 'warning' : 'danger'">{{ row.allow_vetoed_publish ? "允许" : "拦截" }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="default_publisher" label="默认发布者" min-width="140" />
          <el-table-column prop="updated_by" label="维护者" min-width="130" />
          <el-table-column prop="updated_at" label="更新时间" min-width="180" />
          <el-table-column label="操作" width="90" fixed="right">
            <template #default="{ row }">
              <el-button v-if="canEditMarket" link type="primary" @click="openEditPolicyDialog(row)">编辑</el-button>
            </template>
          </el-table-column>
        </el-table>
      </section>
    </div>

    <StrategyEngineJobCenter ref="jobCenterRef" />

    <el-dialog v-model="seedDialogVisible" :title="editingSeedID ? '编辑种子集' : '新建种子集'" width="720px">
      <el-form label-position="top">
        <el-form-item label="名称">
          <el-input v-model="seedForm.name" placeholder="例如：核心蓝筹观察池" />
        </el-form-item>
        <div class="form-grid">
          <el-form-item label="类型">
            <el-select v-model="seedForm.target_type">
              <el-option label="股票" value="STOCK" />
              <el-option label="期货" value="FUTURES" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="seedForm.status">
              <el-option label="启用" value="ACTIVE" />
              <el-option label="停用" value="DISABLED" />
            </el-select>
          </el-form-item>
        </div>
        <el-form-item label="设为默认">
          <el-switch v-model="seedForm.is_default" />
        </el-form-item>
        <el-form-item label="种子项（逗号或换行分隔）">
          <el-input v-model="seedForm.itemsText" type="textarea" :rows="6" placeholder="600519.SH\n601318.SH\n300750.SZ" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="seedForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="seedDialogVisible = false">取消</el-button>
        <el-button v-if="canEditMarket" type="primary" :loading="seedSubmitting" @click="handleSubmitSeed">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="agentDialogVisible" :title="editingAgentID ? '编辑角色配置' : '新建角色配置'" width="760px">
      <el-form label-position="top">
        <el-form-item label="名称">
          <el-input v-model="agentForm.name" placeholder="例如：默认五角评审" />
        </el-form-item>
        <div class="form-grid triple">
          <el-form-item label="作用范围">
            <el-select v-model="agentForm.target_type">
              <el-option label="全部" value="ALL" />
              <el-option label="股票" value="STOCK" />
              <el-option label="期货" value="FUTURES" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="agentForm.status">
              <el-option label="启用" value="ACTIVE" />
              <el-option label="停用" value="DISABLED" />
            </el-select>
          </el-form-item>
          <el-form-item label="设为默认">
            <el-switch v-model="agentForm.is_default" />
          </el-form-item>
        </div>
        <el-form-item label="启用角色">
          <el-checkbox-group v-model="agentForm.enabled_agents">
            <el-checkbox v-for="item in agentOptions" :key="item.value" :label="item.value">{{ item.label }}</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        <div class="form-grid triple">
          <el-form-item label="正向通过阈值">
            <el-input-number v-model="agentForm.positive_threshold" :min="1" :max="5" />
          </el-form-item>
          <el-form-item label="负向收缩阈值">
            <el-input-number v-model="agentForm.negative_threshold" :min="1" :max="5" />
          </el-form-item>
          <el-form-item label="允许 veto">
            <el-switch v-model="agentForm.allow_veto" />
          </el-form-item>
        </div>
        <el-form-item label="短历史回退 MOCK">
          <el-switch v-model="agentForm.allow_mock_fallback_on_short_history" />
          <div class="muted" style="margin-top: 8px;">仅期货链路生效；真实历史不足时允许受控回退到 MOCK，并在作业 warning 中明确标记。</div>
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="agentForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="agentDialogVisible = false">取消</el-button>
        <el-button v-if="canEditMarket" type="primary" :loading="agentSubmitting" @click="handleSubmitAgent">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="scenarioDialogVisible" :title="editingScenarioID ? '编辑场景模板' : '新建场景模板'" width="860px">
      <el-form label-position="top">
        <el-form-item label="名称">
          <el-input v-model="scenarioForm.name" placeholder="例如：股票四象限模板" />
        </el-form-item>
        <div class="form-grid triple">
          <el-form-item label="作用范围">
            <el-select v-model="scenarioForm.target_type" @change="handleScenarioTargetChange">
              <el-option label="股票" value="STOCK" />
              <el-option label="期货" value="FUTURES" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="scenarioForm.status">
              <el-option label="启用" value="ACTIVE" />
              <el-option label="停用" value="DISABLED" />
            </el-select>
          </el-form-item>
          <el-form-item label="设为默认">
            <el-switch v-model="scenarioForm.is_default" />
          </el-form-item>
        </div>

        <div class="scenario-editor">
          <section v-for="item in scenarioForm.items" :key="item.scenario" class="scenario-card">
            <div class="scenario-card__head">
              <h4>{{ scenarioLabels[item.scenario] || item.scenario }}</h4>
              <span class="muted">{{ item.scenario }}</span>
            </div>
            <div class="form-grid triple">
              <el-form-item label="标签">
                <el-input v-model="item.label" />
              </el-form-item>
              <el-form-item label="动作建议">
                <el-input v-model="item.action" />
              </el-form-item>
              <el-form-item label="风险信号">
                <el-input v-model="item.risk_signal" />
              </el-form-item>
            </div>
            <div class="form-grid">
              <el-form-item label="分数偏置">
                <el-input-number v-model="item.score_bias" :step="0.5" :min="-10" :max="10" />
              </el-form-item>
              <div></div>
            </div>
            <el-form-item label="解释模板">
              <el-input v-model="item.thesis_template" type="textarea" :rows="3" />
            </el-form-item>
          </section>
        </div>

        <el-form-item label="说明">
          <el-input v-model="scenarioForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="scenarioDialogVisible = false">取消</el-button>
        <el-button v-if="canEditMarket" type="primary" :loading="scenarioSubmitting" @click="handleSubmitScenario">保存</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="policyDialogVisible" :title="editingPolicyID ? '编辑发布策略' : '新建发布策略'" width="760px">
      <el-form label-position="top">
        <el-form-item label="名称">
          <el-input v-model="policyForm.name" placeholder="例如：默认发布门槛" />
        </el-form-item>
        <div class="form-grid triple">
          <el-form-item label="作用范围">
            <el-select v-model="policyForm.target_type">
              <el-option label="全部" value="ALL" />
              <el-option label="股票" value="STOCK" />
              <el-option label="期货" value="FUTURES" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="policyForm.status">
              <el-option label="启用" value="ACTIVE" />
              <el-option label="停用" value="DISABLED" />
            </el-select>
          </el-form-item>
          <el-form-item label="设为默认">
            <el-switch v-model="policyForm.is_default" />
          </el-form-item>
        </div>
        <div class="form-grid triple">
          <el-form-item label="最大风险等级">
            <el-select v-model="policyForm.max_risk_level">
              <el-option label="LOW" value="LOW" />
              <el-option label="MEDIUM" value="MEDIUM" />
              <el-option label="HIGH" value="HIGH" />
            </el-select>
          </el-form-item>
          <el-form-item label="最大警告数">
            <el-input-number v-model="policyForm.max_warning_count" :min="0" :max="20" />
          </el-form-item>
          <el-form-item label="允许带 veto 发布">
            <el-switch v-model="policyForm.allow_vetoed_publish" />
          </el-form-item>
        </div>
        <el-form-item label="默认发布者">
          <el-input v-model="policyForm.default_publisher" placeholder="strategy-engine" />
        </el-form-item>
        <el-form-item label="人工覆盖说明模板">
          <el-input v-model="policyForm.override_note_template" type="textarea" :rows="3" />
        </el-form-item>
        <el-form-item label="说明">
          <el-input v-model="policyForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="policyDialogVisible = false">取消</el-button>
        <el-button v-if="canEditMarket" type="primary" :loading="policySubmitting" @click="handleSubmitPolicy">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.strategy-config-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.panel-alert {
  margin-bottom: 0;
}

.config-grid {
  display: grid;
  gap: 16px;
}

.panel-head {
  margin-bottom: 12px;
}

.panel-head h3 {
  margin: 0 0 4px;
  font-size: 18px;
}

.form-grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.form-grid.triple {
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.scenario-editor {
  display: grid;
  gap: 12px;
}

.scenario-card {
  border: 1px solid var(--el-border-color-light);
  border-radius: 14px;
  padding: 16px;
  background: var(--el-fill-color-lighter);
}

.scenario-card__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.scenario-card__head h4 {
  margin: 0;
  font-size: 16px;
}

:deep(.is-route-focus td) {
  background: rgba(64, 158, 255, 0.12);
}

@media (max-width: 960px) {
  .form-grid,
  .form-grid.triple {
    grid-template-columns: 1fr;
  }
}
</style>
