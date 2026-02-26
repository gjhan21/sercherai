<script setup>
import { onMounted, reactive, ref } from "vue";
import {
  createRiskRule,
  listInviteRecords,
  listReconciliation,
  listRewardRecords,
  listRiskHits,
  listRiskRules,
  listWithdrawRequests,
  retryReconciliation,
  reviewRewardRecord,
  reviewRiskHit,
  reviewWithdrawRequest,
  updateRiskRule
} from "../api/admin";

const activeTab = ref("rules");

const errorMessage = ref("");
const message = ref("");
const refreshingAll = ref(false);

const rulesLoading = ref(false);
const ruleCreating = ref(false);
const rules = ref([]);
const ruleDraftMap = ref({});
const ruleDialogVisible = ref(false);
const ruleForm = reactive({
  rule_code: "",
  rule_name: "",
  threshold: 10,
  status: "ACTIVE"
});

const hitsLoading = ref(false);
const hitPage = ref(1);
const hitPageSize = ref(20);
const hitTotal = ref(0);
const hits = ref([]);
const hitFilters = reactive({
  status: ""
});
const hitReviewDialogVisible = ref(false);
const hitReviewSubmitting = ref(false);
const hitReviewForm = reactive({
  id: "",
  status: "CONFIRMED",
  reason: ""
});

const rewardLoading = ref(false);
const rewardPage = ref(1);
const rewardPageSize = ref(20);
const rewardTotal = ref(0);
const rewardItems = ref([]);
const rewardFilters = reactive({
  status: ""
});
const rewardReviewDialogVisible = ref(false);
const rewardReviewSubmitting = ref(false);
const rewardReviewForm = reactive({
  id: "",
  status: "ISSUED",
  reason: ""
});

const withdrawLoading = ref(false);
const withdrawPage = ref(1);
const withdrawPageSize = ref(20);
const withdrawTotal = ref(0);
const withdrawItems = ref([]);
const withdrawReviewDialogVisible = ref(false);
const withdrawReviewSubmitting = ref(false);
const withdrawReviewForm = reactive({
  id: "",
  status: "APPROVED",
  reason: ""
});

const reconLoading = ref(false);
const reconPage = ref(1);
const reconPageSize = ref(20);
const reconTotal = ref(0);
const reconItems = ref([]);
const retryingReconID = ref("");

const inviteLoading = ref(false);
const invitePage = ref(1);
const invitePageSize = ref(20);
const inviteTotal = ref(0);
const inviteItems = ref([]);
const inviteFilters = reactive({
  status: ""
});

const ruleStatusOptions = ["ACTIVE", "DISABLED"];
const hitStatusOptions = ["PENDING", "CONFIRMED", "RELEASED"];
const rewardStatusOptions = ["PENDING", "ISSUED", "REJECTED", "FROZEN"];
const rewardReviewStatusOptions = ["ISSUED", "REJECTED", "FROZEN"];
const withdrawStatusOptions = ["PENDING", "APPROVED", "REJECTED", "PAID", "FAILED"];

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function toSafeInt(value, fallback = 0) {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? Math.trunc(parsed) : fallback;
}

function clearMessages() {
  errorMessage.value = "";
  message.value = "";
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (["ACTIVE", "SUCCESS", "ISSUED", "APPROVED", "PAID", "CONFIRMED", "RELEASED"].includes(normalized)) {
    return "success";
  }
  if (["PENDING", "RUNNING"].includes(normalized)) {
    return "warning";
  }
  if (["FAILED", "DISABLED", "REJECTED", "FROZEN"].includes(normalized)) {
    return "danger";
  }
  return "info";
}

function resetRuleForm() {
  Object.assign(ruleForm, {
    rule_code: "",
    rule_name: "",
    threshold: 10,
    status: "ACTIVE"
  });
}

function syncRuleDrafts() {
  const map = {};
  rules.value.forEach((item) => {
    map[item.id] = {
      threshold: item.threshold,
      status: item.status || "ACTIVE"
    };
  });
  ruleDraftMap.value = map;
}

async function fetchRules(options = {}) {
  const { keepMessage = false } = options;
  rulesLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listRiskRules();
    rules.value = data.items || [];
    syncRuleDrafts();
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载风控规则失败");
  } finally {
    rulesLoading.value = false;
  }
}

async function submitRule() {
  const payload = {
    rule_code: ruleForm.rule_code.trim(),
    rule_name: ruleForm.rule_name.trim(),
    threshold: toSafeInt(ruleForm.threshold, 0),
    status: ruleForm.status
  };
  if (!payload.rule_code || !payload.rule_name) {
    errorMessage.value = "rule_code 和 rule_name 不能为空";
    return;
  }

  ruleCreating.value = true;
  clearMessages();
  try {
    await createRiskRule(payload);
    ruleDialogVisible.value = false;
    resetRuleForm();
    await fetchRules({ keepMessage: true });
    message.value = `风控规则 ${payload.rule_code} 已创建`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "创建风控规则失败");
  } finally {
    ruleCreating.value = false;
  }
}

async function saveRule(item) {
  const draft = ruleDraftMap.value[item.id] || {};
  const targetThreshold = toSafeInt(draft.threshold, item.threshold);
  const targetStatus = (draft.status || "").trim();
  if (!targetStatus) {
    errorMessage.value = "规则状态不能为空";
    return;
  }
  if (targetThreshold === item.threshold && targetStatus === item.status) {
    return;
  }

  clearMessages();
  try {
    await updateRiskRule(item.id, {
      rule_code: item.rule_code,
      rule_name: item.rule_name,
      threshold: targetThreshold,
      status: targetStatus
    });
    await fetchRules({ keepMessage: true });
    message.value = `风控规则 ${item.rule_code} 已更新`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "更新风控规则失败");
  }
}

async function fetchRiskHitList(options = {}) {
  const { keepMessage = false } = options;
  hitsLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listRiskHits({
      status: hitFilters.status,
      page: hitPage.value,
      page_size: hitPageSize.value
    });
    hits.value = data.items || [];
    hitTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载风险命中失败");
  } finally {
    hitsLoading.value = false;
  }
}

function openHitReviewDialog(item, status) {
  Object.assign(hitReviewForm, {
    id: item.id,
    status,
    reason: ""
  });
  hitReviewDialogVisible.value = true;
}

async function submitHitReview() {
  if (!hitReviewForm.id || !hitReviewForm.status) {
    errorMessage.value = "缺少风险命中审核参数";
    return;
  }

  hitReviewSubmitting.value = true;
  clearMessages();
  try {
    await reviewRiskHit(hitReviewForm.id, hitReviewForm.status, hitReviewForm.reason.trim());
    hitReviewDialogVisible.value = false;
    await fetchRiskHitList({ keepMessage: true });
    message.value = `风险命中 ${hitReviewForm.id} 已处理为 ${hitReviewForm.status}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "处理风险命中失败");
  } finally {
    hitReviewSubmitting.value = false;
  }
}

function applyHitFilters() {
  hitPage.value = 1;
  fetchRiskHitList();
}

function resetHitFilters() {
  hitFilters.status = "";
  hitPage.value = 1;
  fetchRiskHitList();
}

function handleHitPageChange(nextPage) {
  if (nextPage === hitPage.value) {
    return;
  }
  hitPage.value = nextPage;
  fetchRiskHitList();
}

async function fetchRewardList(options = {}) {
  const { keepMessage = false } = options;
  rewardLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listRewardRecords({
      status: rewardFilters.status,
      page: rewardPage.value,
      page_size: rewardPageSize.value
    });
    rewardItems.value = data.items || [];
    rewardTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载奖励记录失败");
  } finally {
    rewardLoading.value = false;
  }
}

function openRewardReviewDialog(item, status) {
  Object.assign(rewardReviewForm, {
    id: item.id,
    status,
    reason: ""
  });
  rewardReviewDialogVisible.value = true;
}

async function submitRewardReview() {
  if (!rewardReviewForm.id || !rewardReviewForm.status) {
    errorMessage.value = "缺少奖励审核参数";
    return;
  }

  rewardReviewSubmitting.value = true;
  clearMessages();
  try {
    await reviewRewardRecord(rewardReviewForm.id, rewardReviewForm.status, rewardReviewForm.reason.trim());
    rewardReviewDialogVisible.value = false;
    await fetchRewardList({ keepMessage: true });
    message.value = `奖励记录 ${rewardReviewForm.id} 已处理为 ${rewardReviewForm.status}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "审核奖励记录失败");
  } finally {
    rewardReviewSubmitting.value = false;
  }
}

function applyRewardFilters() {
  rewardPage.value = 1;
  fetchRewardList();
}

function resetRewardFilters() {
  rewardFilters.status = "";
  rewardPage.value = 1;
  fetchRewardList();
}

function handleRewardPageChange(nextPage) {
  if (nextPage === rewardPage.value) {
    return;
  }
  rewardPage.value = nextPage;
  fetchRewardList();
}

async function fetchWithdrawList(options = {}) {
  const { keepMessage = false } = options;
  withdrawLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listWithdrawRequests({
      page: withdrawPage.value,
      page_size: withdrawPageSize.value
    });
    withdrawItems.value = data.items || [];
    withdrawTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载提现申请失败");
  } finally {
    withdrawLoading.value = false;
  }
}

function openWithdrawReviewDialog(item, status) {
  Object.assign(withdrawReviewForm, {
    id: item.id,
    status,
    reason: ""
  });
  withdrawReviewDialogVisible.value = true;
}

async function submitWithdrawReview() {
  if (!withdrawReviewForm.id || !withdrawReviewForm.status) {
    errorMessage.value = "缺少提现审核参数";
    return;
  }

  withdrawReviewSubmitting.value = true;
  clearMessages();
  try {
    await reviewWithdrawRequest(withdrawReviewForm.id, withdrawReviewForm.status, withdrawReviewForm.reason.trim());
    withdrawReviewDialogVisible.value = false;
    await fetchWithdrawList({ keepMessage: true });
    message.value = `提现申请 ${withdrawReviewForm.id} 已处理为 ${withdrawReviewForm.status}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "审核提现申请失败");
  } finally {
    withdrawReviewSubmitting.value = false;
  }
}

function handleWithdrawPageChange(nextPage) {
  if (nextPage === withdrawPage.value) {
    return;
  }
  withdrawPage.value = nextPage;
  fetchWithdrawList();
}

async function fetchReconciliationList(options = {}) {
  const { keepMessage = false } = options;
  reconLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listReconciliation({
      page: reconPage.value,
      page_size: reconPageSize.value
    });
    reconItems.value = data.items || [];
    reconTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载对账记录失败");
  } finally {
    reconLoading.value = false;
  }
}

async function handleRetryReconciliation(item) {
  retryingReconID.value = item.id;
  clearMessages();
  try {
    await retryReconciliation(item.id);
    await fetchReconciliationList({ keepMessage: true });
    message.value = `对账批次 ${item.id} 已触发重试`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "重试对账失败");
  } finally {
    retryingReconID.value = "";
  }
}

function handleReconPageChange(nextPage) {
  if (nextPage === reconPage.value) {
    return;
  }
  reconPage.value = nextPage;
  fetchReconciliationList();
}

async function fetchInviteList(options = {}) {
  const { keepMessage = false } = options;
  inviteLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listInviteRecords({
      status: inviteFilters.status,
      page: invitePage.value,
      page_size: invitePageSize.value
    });
    inviteItems.value = data.items || [];
    inviteTotal.value = data.total || 0;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载邀请记录失败");
  } finally {
    inviteLoading.value = false;
  }
}

function applyInviteFilters() {
  invitePage.value = 1;
  fetchInviteList();
}

function resetInviteFilters() {
  inviteFilters.status = "";
  invitePage.value = 1;
  fetchInviteList();
}

function handleInvitePageChange(nextPage) {
  if (nextPage === invitePage.value) {
    return;
  }
  invitePage.value = nextPage;
  fetchInviteList();
}

async function refreshCurrentTab() {
  if (activeTab.value === "rules") {
    await fetchRules();
    return;
  }
  if (activeTab.value === "hits") {
    await fetchRiskHitList();
    return;
  }
  if (activeTab.value === "rewards") {
    await fetchRewardList();
    return;
  }
  if (activeTab.value === "withdraws") {
    await fetchWithdrawList();
    return;
  }
  if (activeTab.value === "reconciliation") {
    await fetchReconciliationList();
    return;
  }
  await fetchInviteList();
}

async function refreshAll() {
  refreshingAll.value = true;
  clearMessages();
  try {
    await Promise.all([
      fetchRules({ keepMessage: true }),
      fetchRiskHitList({ keepMessage: true }),
      fetchRewardList({ keepMessage: true }),
      fetchWithdrawList({ keepMessage: true }),
      fetchReconciliationList({ keepMessage: true }),
      fetchInviteList({ keepMessage: true })
    ]);
    message.value = "风控中心数据已刷新";
  } finally {
    refreshingAll.value = false;
  }
}

onMounted(refreshAll);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">风控中心</h1>
        <p class="muted">规则配置、风险命中、奖励/提现审核、对账重试与邀请追踪</p>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <el-button :loading="refreshingAll" @click="refreshAll">刷新全部</el-button>
        <el-button type="primary" plain :loading="refreshingAll" @click="refreshCurrentTab">刷新当前页签</el-button>
      </div>
    </div>

    <el-alert
      v-if="errorMessage"
      :title="errorMessage"
      type="error"
      show-icon
      style="margin-bottom: 12px"
    />
    <el-alert
      v-if="message"
      :title="message"
      type="success"
      show-icon
      style="margin-bottom: 12px"
    />

    <el-tabs v-model="activeTab" type="border-card">
      <el-tab-pane label="风控规则" name="rules">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-button type="primary" @click="ruleDialogVisible = true">新增规则</el-button>
            <el-button @click="fetchRules">刷新规则</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="rules" border stripe v-loading="rulesLoading" empty-text="暂无风控规则">
            <el-table-column prop="id" label="规则ID" min-width="130" />
            <el-table-column prop="rule_code" label="规则编码" min-width="130" />
            <el-table-column prop="rule_name" label="规则名称" min-width="180" />
            <el-table-column label="阈值" min-width="140">
              <template #default="{ row }">
                <el-input-number
                  v-model="ruleDraftMap[row.id].threshold"
                  :min="0"
                  :step="1"
                  controls-position="right"
                  style="width: 120px"
                />
              </template>
            </el-table-column>
            <el-table-column label="状态" min-width="200">
              <template #default="{ row }">
                <div class="inline-actions">
                  <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
                  <el-select v-model="ruleDraftMap[row.id].status" style="width: 120px">
                    <el-option v-for="item in ruleStatusOptions" :key="item" :label="item" :value="item" />
                  </el-select>
                  <el-button size="small" @click="saveRule(row)">保存</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane label="风险命中" name="hits">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select v-model="hitFilters.status" clearable placeholder="状态" style="width: 160px">
              <el-option v-for="item in hitStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-button type="primary" plain @click="applyHitFilters">查询</el-button>
            <el-button @click="resetHitFilters">重置</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="hits" border stripe v-loading="hitsLoading" empty-text="暂无风险命中记录">
            <el-table-column prop="id" label="命中ID" min-width="140" />
            <el-table-column prop="rule_code" label="规则编码" min-width="130" />
            <el-table-column prop="user_id" label="用户ID" min-width="130" />
            <el-table-column prop="risk_level" label="风险等级" min-width="110" />
            <el-table-column label="状态" min-width="110">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" align="right" min-width="220">
              <template #default="{ row }">
                <div class="inline-actions inline-actions--right">
                  <el-button size="small" type="success" plain @click="openHitReviewDialog(row, 'CONFIRMED')">确认</el-button>
                  <el-button size="small" type="warning" plain @click="openHitReviewDialog(row, 'RELEASED')">释放</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ hitPage }} 页，共 {{ hitTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="hitPage"
              :page-size="hitPageSize"
              :total="hitTotal"
              @current-change="handleHitPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="奖励审核" name="rewards">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select v-model="rewardFilters.status" clearable placeholder="状态" style="width: 160px">
              <el-option v-for="item in rewardStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-button type="primary" plain @click="applyRewardFilters">查询</el-button>
            <el-button @click="resetRewardFilters">重置</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="rewardItems" border stripe v-loading="rewardLoading" empty-text="暂无奖励记录">
            <el-table-column prop="id" label="奖励ID" min-width="140" />
            <el-table-column prop="inviter_user_id" label="邀请人" min-width="130" />
            <el-table-column prop="invitee_user_id" label="被邀请人" min-width="130" />
            <el-table-column prop="reward_type" label="奖励类型" min-width="110" />
            <el-table-column prop="reward_value" label="奖励值" min-width="90" />
            <el-table-column prop="trigger_event" label="触发事件" min-width="130" />
            <el-table-column label="状态" min-width="110">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="issued_at" label="发放时间" min-width="160" />
            <el-table-column label="操作" align="right" min-width="270">
              <template #default="{ row }">
                <div class="inline-actions inline-actions--right">
                  <el-button size="small" type="success" plain @click="openRewardReviewDialog(row, 'ISSUED')">发放</el-button>
                  <el-button size="small" type="danger" plain @click="openRewardReviewDialog(row, 'REJECTED')">驳回</el-button>
                  <el-button size="small" type="warning" plain @click="openRewardReviewDialog(row, 'FROZEN')">冻结</el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ rewardPage }} 页，共 {{ rewardTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="rewardPage"
              :page-size="rewardPageSize"
              :total="rewardTotal"
              @current-change="handleRewardPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="提现审核" name="withdraws">
        <div class="card">
          <el-table :data="withdrawItems" border stripe v-loading="withdrawLoading" empty-text="暂无提现申请">
            <el-table-column prop="id" label="申请ID" min-width="140" />
            <el-table-column prop="user_id" label="用户ID" min-width="130" />
            <el-table-column prop="amount" label="金额" min-width="90" />
            <el-table-column label="状态" min-width="110">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="applied_at" label="申请时间" min-width="170" />
            <el-table-column label="操作" align="right" min-width="320">
              <template #default="{ row }">
                <div class="inline-actions inline-actions--right">
                  <el-button size="small" type="success" plain @click="openWithdrawReviewDialog(row, 'APPROVED')">
                    通过
                  </el-button>
                  <el-button size="small" type="danger" plain @click="openWithdrawReviewDialog(row, 'REJECTED')">
                    驳回
                  </el-button>
                  <el-button size="small" type="primary" plain @click="openWithdrawReviewDialog(row, 'PAID')">
                    置为已打款
                  </el-button>
                  <el-button size="small" type="warning" plain @click="openWithdrawReviewDialog(row, 'FAILED')">
                    标记失败
                  </el-button>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ withdrawPage }} 页，共 {{ withdrawTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="withdrawPage"
              :page-size="withdrawPageSize"
              :total="withdrawTotal"
              @current-change="handleWithdrawPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="支付对账" name="reconciliation">
        <div class="card">
          <el-table :data="reconItems" border stripe v-loading="reconLoading" empty-text="暂无对账记录">
            <el-table-column prop="id" label="批次ID" min-width="150" />
            <el-table-column prop="pay_channel" label="支付渠道" min-width="110" />
            <el-table-column prop="batch_date" label="批次日期" min-width="120" />
            <el-table-column prop="diff_count" label="差异条数" min-width="100" />
            <el-table-column label="状态" min-width="110">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" align="right" min-width="130">
              <template #default="{ row }">
                <el-button
                  size="small"
                  :loading="retryingReconID === row.id"
                  @click="handleRetryReconciliation(row)"
                >
                  重试对账
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ reconPage }} 页，共 {{ reconTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="reconPage"
              :page-size="reconPageSize"
              :total="reconTotal"
              @current-change="handleReconPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="邀请记录" name="invites">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-input v-model="inviteFilters.status" clearable placeholder="状态（可选）" style="width: 180px" />
            <el-button type="primary" plain @click="applyInviteFilters">查询</el-button>
            <el-button @click="resetInviteFilters">重置</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="inviteItems" border stripe v-loading="inviteLoading" empty-text="暂无邀请记录">
            <el-table-column prop="id" label="邀请记录ID" min-width="150" />
            <el-table-column prop="inviter_user_id" label="邀请人" min-width="130" />
            <el-table-column prop="invitee_user_id" label="被邀请人" min-width="130" />
            <el-table-column label="状态" min-width="100">
              <template #default="{ row }">
                <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="register_at" label="注册时间" min-width="170" />
            <el-table-column prop="first_pay_at" label="首付时间" min-width="170" />
            <el-table-column prop="risk_flag" label="风控标记" min-width="120" />
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ invitePage }} 页，共 {{ inviteTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="invitePage"
              :page-size="invitePageSize"
              :total="inviteTotal"
              @current-change="handleInvitePageChange"
            />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="ruleDialogVisible" title="新增风控规则" width="520px" destroy-on-close>
      <el-form label-width="110px">
        <el-form-item label="规则编码" required>
          <el-input v-model="ruleForm.rule_code" placeholder="如 LOGIN_FAIL" />
        </el-form-item>
        <el-form-item label="规则名称" required>
          <el-input v-model="ruleForm.rule_name" placeholder="登录失败阈值" />
        </el-form-item>
        <el-form-item label="阈值" required>
          <el-input-number v-model="ruleForm.threshold" :min="0" :step="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="状态" required>
          <el-select v-model="ruleForm.status" style="width: 100%">
            <el-option v-for="item in ruleStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="ruleCreating" @click="submitRule">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="hitReviewDialogVisible" title="处理风险命中" width="520px" destroy-on-close>
      <el-form label-width="110px">
        <el-form-item label="命中ID">
          <el-text>{{ hitReviewForm.id || "-" }}</el-text>
        </el-form-item>
        <el-form-item label="处理结果">
          <el-select v-model="hitReviewForm.status" style="width: 100%">
            <el-option label="CONFIRMED" value="CONFIRMED" />
            <el-option label="RELEASED" value="RELEASED" />
          </el-select>
        </el-form-item>
        <el-form-item label="处理说明">
          <el-input
            v-model="hitReviewForm.reason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="可选，记录处理原因"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="hitReviewDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="hitReviewSubmitting" @click="submitHitReview">确认提交</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="rewardReviewDialogVisible" title="审核奖励记录" width="520px" destroy-on-close>
      <el-form label-width="110px">
        <el-form-item label="奖励ID">
          <el-text>{{ rewardReviewForm.id || "-" }}</el-text>
        </el-form-item>
        <el-form-item label="审核结果">
          <el-select v-model="rewardReviewForm.status" style="width: 100%">
            <el-option v-for="item in rewardReviewStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核备注">
          <el-input
            v-model="rewardReviewForm.reason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="可选，记录审核理由"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rewardReviewDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="rewardReviewSubmitting" @click="submitRewardReview">确认提交</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="withdrawReviewDialogVisible" title="审核提现申请" width="520px" destroy-on-close>
      <el-form label-width="110px">
        <el-form-item label="申请ID">
          <el-text>{{ withdrawReviewForm.id || "-" }}</el-text>
        </el-form-item>
        <el-form-item label="审核结果">
          <el-select v-model="withdrawReviewForm.status" style="width: 100%">
            <el-option v-for="item in withdrawStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
        </el-form-item>
        <el-form-item label="审核备注">
          <el-input
            v-model="withdrawReviewForm.reason"
            type="textarea"
            :rows="3"
            maxlength="200"
            show-word-limit
            placeholder="可选，记录审核说明"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="withdrawReviewDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="withdrawReviewSubmitting" @click="submitWithdrawReview">确认提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.inline-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.inline-actions--right {
  justify-content: flex-end;
}
</style>
