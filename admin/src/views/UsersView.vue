<script setup>
import { computed, nextTick, onMounted, reactive, ref } from "vue";
import {
  getUserSourceSummary,
  getUserCenterOverview,
  listUsers,
  resetUserPassword,
  updateUserKYCStatus,
  updateUserMemberLevel,
  updateUserSubscription,
  updateUserStatus
} from "../api/admin";
import { getAccessToken, hasPermission } from "../lib/session";

const loading = ref(false);
const exportingFiltered = ref(false);
const retryingFailed = ref(false);
const copyingFailedDetails = ref(false);

const errorMessage = ref("");
const message = ref("");

const filters = reactive({
  status: "",
  kyc_status: "",
  member_level: "",
  registration_source: ""
});

const users = ref([]);
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const sourceSummary = ref({
  total_users: 0,
  direct_users: 0,
  invited_users: 0,
  invite_rate: 0,
  today_invited_users: 0,
  last_7d_invited_users: 0,
  last_7d_first_paid_users: 0,
  last_7d_conversion_rate: 0,
  last_30d_invited_users: 0,
  last_30d_first_paid_users: 0,
  last_30d_conversion_rate: 0
});

const userTableRef = ref(null);
const selectedRows = ref([]);

const draftStatusMap = ref({});
const draftKYCMap = ref({});
const draftLevelMap = ref({});

const batchStatus = ref("ACTIVE");
const batchKYCStatus = ref("APPROVED");
const batchMemberLevel = ref("VIP1");
const batchUpdatingStatus = ref(false);
const batchUpdatingKYC = ref(false);
const batchUpdatingLevel = ref(false);

const batchResultVisible = ref(false);
const batchResultTitle = ref("");
const batchResultRows = ref([]);
const batchResultFilter = ref("all");

const centerDialogVisible = ref(false);
const centerLoading = ref(false);
const centerErrorMessage = ref("");
const centerActiveTab = ref("vip");
const centerTargetUserID = ref("");
const centerData = ref({
  user_profile: {},
  membership_quota: {},
  payment_summary: {},
  reading_summary: {},
  subscription_summary: {},
  invite_summary: {},
  membership_orders: [],
  recharge_records: [],
  browse_history: [],
  subscriptions: [],
  share_links: [],
  invite_records: []
});
const centerSubscriptionDraft = ref({});
const centerSavingSubscriptionID = ref("");
const passwordDialogVisible = ref(false);
const passwordSubmitting = ref(false);
const passwordDialogError = ref("");
const passwordForm = reactive({
  userID: "",
  account: "",
  password: "",
  confirmPassword: ""
});

const statusOptions = ["ACTIVE", "DISABLED", "BANNED"];
const kycStatusOptions = ["PENDING", "APPROVED", "REJECTED"];
const registrationSourceOptions = [
  { value: "DIRECT", label: "自然注册" },
  { value: "INVITED", label: "邀请注册" }
];
const subscriptionStatusOptions = ["ACTIVE", "PAUSED"];
const subscriptionFrequencyOptions = ["INSTANT", "DAILY", "WEEKLY"];
const canEditUsers = hasPermission("users.edit");

const selectedCount = computed(() => selectedRows.value.length);
const failedBatchRows = computed(() => batchResultRows.value.filter((row) => row.result === "FAILED"));
const canBatchUpdateStatus = computed(() => selectedCount.value > 0 && batchStatus.value.trim() !== "");
const canBatchUpdateKYC = computed(() => selectedCount.value > 0 && batchKYCStatus.value.trim() !== "");
const canBatchUpdateLevel = computed(() => selectedCount.value > 0 && batchMemberLevel.value.trim() !== "");
const batchResultStats = computed(() => {
  const stats = {
    total: batchResultRows.value.length,
    success: 0,
    failed: 0,
    skipped: 0
  };
  batchResultRows.value.forEach((row) => {
    const result = (row.result || "").toUpperCase();
    if (result === "SUCCESS") {
      stats.success += 1;
      return;
    }
    if (result === "FAILED") {
      stats.failed += 1;
      return;
    }
    if (result === "SKIPPED") {
      stats.skipped += 1;
    }
  });
  return stats;
});
const displayBatchResultRows = computed(() => {
  if (batchResultFilter.value === "failed") {
    return batchResultRows.value.filter((row) => row.result === "FAILED");
  }
  if (batchResultFilter.value === "skipped") {
    return batchResultRows.value.filter((row) => row.result === "SKIPPED");
  }
  return batchResultRows.value;
});
const centerMembershipOrders = computed(() => centerData.value?.membership_orders || []);
const centerRechargeRecords = computed(() => centerData.value?.recharge_records || []);
const centerBrowseHistory = computed(() => centerData.value?.browse_history || []);
const centerSubscriptions = computed(() => centerData.value?.subscriptions || []);
const centerShareLinks = computed(() => centerData.value?.share_links || []);
const centerInviteRecords = computed(() => centerData.value?.invite_records || []);
const centerInviteSummaryCards = computed(() => {
  const summary = centerData.value?.invite_summary || {};
  return [
    { label: "分享链接数", value: Number(summary.share_link_count || 0) },
    { label: "邀请注册总数", value: Number(summary.registered_count || 0) },
    { label: "首单转化总数", value: Number(summary.first_paid_count || 0) },
    { label: "总转化率", value: `${(Number(summary.conversion_rate || 0) * 100).toFixed(1)}%` },
    { label: "近7天转化率", value: `${(Number(summary.last_7d_conversion_rate || 0) * 100).toFixed(1)}%` },
    { label: "近30天转化率", value: `${(Number(summary.last_30d_conversion_rate || 0) * 100).toFixed(1)}%` }
  ];
});
const centerInviteTimeline = computed(() =>
  (centerInviteRecords.value || [])
    .map((item) => {
      const statusRaw = String(item.status || "").toUpperCase();
      return {
        id: item.id || `${item.invitee_user_id || "invitee"}-${item.register_at || ""}`,
        inviteeUserID: item.invitee_user_id || "-",
        statusRaw,
        statusText: mapInviteStatus(statusRaw),
        registerAt: formatCenterDateTime(item.register_at),
        firstPayAt: formatCenterDateTime(item.first_pay_at),
        riskFlag: mapInviteRiskFlag(item.risk_flag),
        sortTS: Date.parse(item.register_at || "") || 0
      };
    })
    .sort((a, b) => b.sortTS - a.sortTS)
);
const sourceSummaryCards = computed(() => {
  const totalUsers = Number(sourceSummary.value?.total_users || 0);
  const directUsers = Number(sourceSummary.value?.direct_users || 0);
  const invitedUsers = Number(sourceSummary.value?.invited_users || 0);
  const inviteRate = Number(sourceSummary.value?.invite_rate || 0);
  const todayInvitedUsers = Number(sourceSummary.value?.today_invited_users || 0);
  const last7dInvitedUsers = Number(sourceSummary.value?.last_7d_invited_users || 0);
  const last7dConversionRate = Number(sourceSummary.value?.last_7d_conversion_rate || 0);
  const last30dInvitedUsers = Number(sourceSummary.value?.last_30d_invited_users || 0);
  const last30dConversionRate = Number(sourceSummary.value?.last_30d_conversion_rate || 0);
  return [
    { label: "用户总数", value: totalUsers },
    { label: "自然注册", value: directUsers },
    { label: "邀请注册", value: invitedUsers },
    { label: "邀请占比", value: `${(inviteRate * 100).toFixed(1)}%` },
    { label: "今日邀请注册", value: todayInvitedUsers },
    { label: "近7天邀请注册", value: last7dInvitedUsers },
    { label: "近7天转化率", value: `${(last7dConversionRate * 100).toFixed(1)}%` },
    { label: "近30天邀请注册", value: last30dInvitedUsers },
    { label: "近30天转化率", value: `${(last30dConversionRate * 100).toFixed(1)}%` }
  ];
});

function syncCenterSubscriptionDraft(rows = []) {
  const draft = {};
  (rows || []).forEach((row) => {
    if (!row?.id) {
      return;
    }
    draft[row.id] = {
      frequency: String(row.frequency || "DAILY").toUpperCase(),
      status: String(row.status || "ACTIVE").toUpperCase()
    };
  });
  centerSubscriptionDraft.value = draft;
}

function syncDrafts() {
  const statusMap = {};
  const kycMap = {};
  const levelMap = {};
  users.value.forEach((user) => {
    statusMap[user.id] = user.status || "ACTIVE";
    kycMap[user.id] = user.kyc_status || "PENDING";
    levelMap[user.id] = user.member_level || "FREE";
  });
  draftStatusMap.value = statusMap;
  draftKYCMap.value = kycMap;
  draftLevelMap.value = levelMap;
}

function clearSelection() {
  selectedRows.value = [];
  nextTick(() => {
    userTableRef.value?.clearSelection();
  });
}

function handleSelectionChange(rows) {
  selectedRows.value = rows || [];
}

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function ensureCanEditUsers() {
  if (canEditUsers) {
    return true;
  }
  errorMessage.value = "当前账号只有查看权限，无法修改用户状态、KYC、会员等级、密码或订阅配置";
  return false;
}

function resetPasswordDialogState() {
  passwordDialogError.value = "";
  passwordSubmitting.value = false;
  passwordForm.userID = "";
  passwordForm.account = "";
  passwordForm.password = "";
  passwordForm.confirmPassword = "";
}

async function fetchUsers(options = {}) {
  const { keepMessage = false } = options;
  loading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const listParams = {
      status: filters.status,
      kyc_status: filters.kyc_status,
      member_level: filters.member_level,
      registration_source: filters.registration_source,
      page: page.value,
      page_size: pageSize.value
    };
    const [listResult, summaryResult] = await Promise.allSettled([
      listUsers(listParams),
      getUserSourceSummary({
        status: filters.status,
        kyc_status: filters.kyc_status,
        member_level: filters.member_level,
        registration_source: filters.registration_source
      })
    ]);
    if (listResult.status !== "fulfilled") {
      throw listResult.reason;
    }
    const data = listResult.value;
    users.value = data.items || [];
    total.value = data.total || 0;
    if (summaryResult.status === "fulfilled" && summaryResult.value) {
      sourceSummary.value = summaryResult.value;
    }
    syncDrafts();
    clearSelection();
  } catch (error) {
    errorMessage.value = error.message || "加载用户失败";
  } finally {
    loading.value = false;
  }
}

async function handleUpdateStatus(user) {
  if (!ensureCanEditUsers()) {
    return;
  }
  const target = (draftStatusMap.value[user.id] || "").trim();
  if (!target || target === user.status) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await updateUserStatus(user.id, target);
    await fetchUsers({ keepMessage: true });
    message.value = `用户 ${user.id} 状态已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = error.message || "更新用户状态失败";
  }
}

async function handleUpdateKYC(user) {
  if (!ensureCanEditUsers()) {
    return;
  }
  const target = (draftKYCMap.value[user.id] || "").trim();
  if (!target || target === user.kyc_status) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await updateUserKYCStatus(user.id, target);
    await fetchUsers({ keepMessage: true });
    message.value = `用户 ${user.id} KYC 状态已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = error.message || "更新 KYC 状态失败";
  }
}

async function handleUpdateMemberLevel(user) {
  if (!ensureCanEditUsers()) {
    return;
  }
  const target = (draftLevelMap.value[user.id] || "").trim();
  if (!target || target === user.member_level) {
    return;
  }
  errorMessage.value = "";
  message.value = "";
  try {
    await updateUserMemberLevel(user.id, target);
    await fetchUsers({ keepMessage: true });
    message.value = `用户 ${user.id} 会员等级已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = error.message || "更新会员等级失败";
  }
}

function openResetPasswordDialog(user) {
  if (!ensureCanEditUsers()) {
    return;
  }
  resetPasswordDialogState();
  passwordForm.userID = user?.id || "";
  passwordForm.account = user?.phone || user?.email || "-";
  passwordDialogVisible.value = true;
}

async function handleResetUserPassword() {
  if (!ensureCanEditUsers()) {
    return;
  }
  const userID = String(passwordForm.userID || "").trim();
  const password = String(passwordForm.password || "");
  const confirmPassword = String(passwordForm.confirmPassword || "");

  passwordDialogError.value = "";
  errorMessage.value = "";
  message.value = "";

  if (!userID) {
    passwordDialogError.value = "缺少用户ID，请关闭后重试";
    return;
  }
  if (password.length < 8) {
    passwordDialogError.value = "新密码长度至少为8位";
    return;
  }
  if (password !== confirmPassword) {
    passwordDialogError.value = "两次输入的密码不一致";
    return;
  }

  passwordSubmitting.value = true;
  try {
    await resetUserPassword(userID, password);
    passwordDialogVisible.value = false;
    resetPasswordDialogState();
    message.value = `用户 ${userID} 密码已更新`;
  } catch (error) {
    passwordDialogError.value = normalizeErrorMessage(error, "修改用户密码失败");
  } finally {
    passwordSubmitting.value = false;
  }
}

function mapDisplayStatus(raw, type = "") {
  const normalized = (raw || "").toUpperCase();
  if (type === "kyc") {
    if (normalized === "APPROVED") return "已认证";
    if (normalized === "PENDING") return "审核中";
    if (normalized === "REJECTED") return "未通过";
  }
  if (type === "payment") {
    if (normalized === "PAID" || normalized === "SUCCESS") return "已支付";
    if (normalized === "PENDING") return "处理中";
    if (normalized === "FAILED") return "失败";
    if (normalized === "REFUNDED" || normalized === "REFUND") return "已退款";
  }
  if (type === "subscription") {
    if (normalized === "ACTIVE") return "生效中";
    if (normalized === "PAUSED" || normalized === "INACTIVE") return "已暂停";
  }
  return raw || "-";
}

function mapRegistrationSource(raw) {
  const normalized = (raw || "").toUpperCase();
  if (normalized === "INVITED") return "邀请注册";
  if (normalized === "DIRECT") return "自然注册";
  return raw || "-";
}

function mapSubscriptionFrequency(raw) {
  const normalized = (raw || "").toUpperCase();
  if (normalized === "INSTANT") return "实时";
  if (normalized === "DAILY") return "每日";
  if (normalized === "WEEKLY") return "每周";
  return raw || "-";
}

function mapInviteStatus(raw) {
  const normalized = (raw || "").toUpperCase();
  if (normalized === "REGISTERED") return "已注册";
  if (normalized === "FIRST_PAID") return "首单完成";
  if (normalized === "INVALID") return "失效";
  return raw || "-";
}

function mapInviteRiskFlag(raw) {
  const normalized = (raw || "").toUpperCase();
  if (normalized === "NORMAL") return "正常";
  if (normalized === "RISK") return "疑似风险";
  if (normalized === "BLOCKED") return "已拦截";
  return raw || "-";
}

function statusTagByType(raw, type = "") {
  const normalized = (raw || "").toUpperCase();
  if (type === "payment") {
    if (normalized === "PAID" || normalized === "SUCCESS") return "success";
    if (normalized === "PENDING") return "warning";
    return "danger";
  }
  if (type === "subscription") {
    if (normalized === "ACTIVE") return "success";
    if (normalized === "PAUSED" || normalized === "INACTIVE") return "warning";
    return "info";
  }
  return "info";
}

function registrationSourceTagType(source) {
  const normalized = (source || "").toUpperCase();
  if (normalized === "INVITED") return "success";
  if (normalized === "DIRECT") return "info";
  return "warning";
}

function formatCenterDateTime(value) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return date.toLocaleString("zh-CN", { hour12: false });
}

function formatCenterAmount(value) {
  const num = Number(value || 0);
  return `¥${Number.isFinite(num) ? num.toFixed(2).replace(/\.00$/, "") : "0"}`;
}

async function openUserCenter(user) {
  if (!user?.id) {
    return;
  }
  centerDialogVisible.value = true;
  centerTargetUserID.value = user.id;
  centerActiveTab.value = "vip";
  centerErrorMessage.value = "";
  centerSavingSubscriptionID.value = "";
  centerLoading.value = true;

  try {
    const data = await getUserCenterOverview(user.id, { limit: 80 });
    const nextData = {
      user_profile: data?.user_profile || {},
      membership_quota: data?.membership_quota || {},
      payment_summary: data?.payment_summary || {},
      reading_summary: data?.reading_summary || {},
      subscription_summary: data?.subscription_summary || {},
      invite_summary: data?.invite_summary || {},
      membership_orders: data?.membership_orders || [],
      recharge_records: data?.recharge_records || [],
      browse_history: data?.browse_history || [],
      subscriptions: data?.subscriptions || [],
      share_links: data?.share_links || [],
      invite_records: data?.invite_records || []
    };
    centerData.value = nextData;
    syncCenterSubscriptionDraft(nextData.subscriptions);
  } catch (error) {
    centerErrorMessage.value = normalizeErrorMessage(error, "加载客户中心数据失败");
  } finally {
    centerLoading.value = false;
  }
}

async function handleSaveCenterSubscription(row) {
  if (!ensureCanEditUsers()) {
    return;
  }
  if (!centerTargetUserID.value || !row?.id) {
    return;
  }
  const draft = centerSubscriptionDraft.value[row.id];
  if (!draft) {
    centerErrorMessage.value = "订阅草稿数据缺失，请刷新后重试";
    return;
  }

  const frequency = String(draft.frequency || "").toUpperCase();
  const status = String(draft.status || "").toUpperCase();
  const prevFrequency = String(row.frequency || "").toUpperCase();
  const prevStatus = String(row.status || "").toUpperCase();

  if (!subscriptionFrequencyOptions.includes(frequency)) {
    centerErrorMessage.value = "订阅频率不合法";
    return;
  }
  if (!subscriptionStatusOptions.includes(status)) {
    centerErrorMessage.value = "订阅状态不合法";
    return;
  }
  if (frequency === prevFrequency && status === prevStatus) {
    message.value = `订阅 ${row.id} 未发生变化`;
    return;
  }

  centerSavingSubscriptionID.value = row.id;
  centerErrorMessage.value = "";
  try {
    await updateUserSubscription(centerTargetUserID.value, row.id, {
      frequency,
      status
    });
    row.frequency = frequency;
    row.status = status;
    message.value = `订阅 ${row.id} 更新成功`;
  } catch (error) {
    centerErrorMessage.value = normalizeErrorMessage(error, "更新订阅失败");
  } finally {
    centerSavingSubscriptionID.value = "";
  }
}

function openBatchResultDialog(title, rows) {
  batchResultTitle.value = title;
  batchResultRows.value = rows;
  batchResultFilter.value = "all";
  batchResultVisible.value = true;
}

async function runBatchUpdate(config) {
  if (!ensureCanEditUsers()) {
    return;
  }
  const target = (config.targetValue || "").trim();
  if (!target) {
    errorMessage.value = config.emptyTargetMessage || "请先设置目标值";
    return;
  }
  if (selectedRows.value.length <= 0) {
    errorMessage.value = "请先勾选用户";
    return;
  }

  errorMessage.value = "";
  message.value = "";
  config.loadingRef.value = true;

  let success = 0;
  let failed = 0;
  let skipped = 0;
  const resultRows = [];

  try {
    for (const user of selectedRows.value) {
      const currentValue = (config.currentValue(user) || "").trim();
      if (currentValue === target) {
        skipped += 1;
        resultRows.push({
          id: user.id,
          action: config.action,
          action_key: config.actionKey,
          target,
          result: "SKIPPED",
          reason: config.skippedReason || "已是目标值"
        });
        continue;
      }

      try {
        await config.executor(user.id, target);
        success += 1;
        resultRows.push({
          id: user.id,
          action: config.action,
          action_key: config.actionKey,
          target,
          result: "SUCCESS",
          reason: config.successReason(target)
        });
      } catch (error) {
        failed += 1;
        resultRows.push({
          id: user.id,
          action: config.action,
          action_key: config.actionKey,
          target,
          result: "FAILED",
          reason: normalizeErrorMessage(error, "更新失败")
        });
      }
    }
  } finally {
    config.loadingRef.value = false;
  }

  await fetchUsers({ keepMessage: true });
  message.value = `${config.action}完成：成功 ${success}，失败 ${failed}，跳过 ${skipped}`;
  openBatchResultDialog(config.title, resultRows);
}

async function handleBatchUpdateStatus() {
  await runBatchUpdate({
    title: "批量更新用户状态结果",
    action: "批量更新状态",
    actionKey: "USER_STATUS",
    targetValue: batchStatus.value,
    loadingRef: batchUpdatingStatus,
    emptyTargetMessage: "请先选择目标用户状态",
    currentValue: (user) => user.status,
    skippedReason: "用户状态已是目标值",
    successReason: (target) => `状态已更新为 ${target}`,
    executor: updateUserStatus
  });
}

async function handleBatchUpdateKYC() {
  await runBatchUpdate({
    title: "批量更新 KYC 状态结果",
    action: "批量更新KYC",
    actionKey: "KYC_STATUS",
    targetValue: batchKYCStatus.value,
    loadingRef: batchUpdatingKYC,
    emptyTargetMessage: "请先选择目标 KYC 状态",
    currentValue: (user) => user.kyc_status,
    skippedReason: "KYC 状态已是目标值",
    successReason: (target) => `KYC 状态已更新为 ${target}`,
    executor: updateUserKYCStatus
  });
}

async function handleBatchUpdateLevel() {
  await runBatchUpdate({
    title: "批量更新会员等级结果",
    action: "批量更新会员等级",
    actionKey: "MEMBER_LEVEL",
    targetValue: batchMemberLevel.value,
    loadingRef: batchUpdatingLevel,
    emptyTargetMessage: "请先输入目标会员等级",
    currentValue: (user) => user.member_level,
    skippedReason: "会员等级已是目标值",
    successReason: (target) => `会员等级已更新为 ${target}`,
    executor: updateUserMemberLevel
  });
}

async function executeBatchResultRow(row) {
  const actionKey = (row.action_key || "").toUpperCase();
  const target = (row.target || "").trim();
  if (!target) {
    throw new Error("缺少目标值");
  }

  if (actionKey === "USER_STATUS") {
    await updateUserStatus(row.id, target);
    return `状态已更新为 ${target}`;
  }
  if (actionKey === "KYC_STATUS") {
    await updateUserKYCStatus(row.id, target);
    return `KYC 状态已更新为 ${target}`;
  }
  if (actionKey === "MEMBER_LEVEL") {
    await updateUserMemberLevel(row.id, target);
    return `会员等级已更新为 ${target}`;
  }
  throw new Error(`未知动作: ${actionKey || "-"}`);
}

async function retryFailedBatchRows() {
  if (failedBatchRows.value.length <= 0) {
    return;
  }

  retryingFailed.value = true;
  errorMessage.value = "";
  message.value = "";

  let success = 0;
  let failed = 0;
  const resultRows = [];

  try {
    for (const row of failedBatchRows.value) {
      try {
        const tip = await executeBatchResultRow(row);
        success += 1;
        resultRows.push({
          ...row,
          action: `${row.action}重试`,
          result: "SUCCESS",
          reason: tip
        });
      } catch (error) {
        failed += 1;
        resultRows.push({
          ...row,
          action: `${row.action}重试`,
          result: "FAILED",
          reason: normalizeErrorMessage(error, "重试失败")
        });
      }
    }
  } finally {
    retryingFailed.value = false;
  }

  await fetchUsers({ keepMessage: true });
  message.value = `失败任务重试完成：成功 ${success}，失败 ${failed}`;
  openBatchResultDialog("失败任务重试结果", resultRows);
}

function buildFailedDetailsText() {
  return failedBatchRows.value
    .map((row) => {
      return [
        `用户ID=${row.id || "-"}`,
        `动作=${row.action || "-"}`,
        `目标=${row.target || "-"}`,
        `结果=${row.result || "-"}`,
        `原因=${row.reason || "-"}`
      ].join(" | ");
    })
    .join("\n");
}

async function copyFailedDetails() {
  if (failedBatchRows.value.length <= 0) {
    return;
  }
  const text = buildFailedDetailsText();
  if (!text) {
    return;
  }

  copyingFailedDetails.value = true;
  try {
    if (navigator?.clipboard?.writeText) {
      await navigator.clipboard.writeText(text);
    } else {
      const textarea = document.createElement("textarea");
      textarea.value = text;
      textarea.style.position = "fixed";
      textarea.style.opacity = "0";
      document.body.appendChild(textarea);
      textarea.focus();
      textarea.select();
      document.execCommand("copy");
      document.body.removeChild(textarea);
    }
    message.value = `已复制失败明细，共 ${failedBatchRows.value.length} 条`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "复制失败明细失败");
  } finally {
    copyingFailedDetails.value = false;
  }
}

function csvEscape(value) {
  const text = String(value ?? "");
  if (/[",\n]/.test(text)) {
    return `"${text.replace(/"/g, '""')}"`;
  }
  return text;
}

function triggerCSVDownload(content, fileName) {
  const blob = new Blob([`\uFEFF${content}`], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = fileName;
  document.body.appendChild(anchor);
  anchor.click();
  document.body.removeChild(anchor);
  URL.revokeObjectURL(url);
}

function buildCSVRows(items) {
  const header = [
    "id",
    "phone",
    "email",
    "status",
    "kyc_status",
    "member_level",
    "registration_source",
    "inviter_user_id",
    "invite_code",
    "invite_registered_at",
    "created_at"
  ];
  const rows = items.map((item) => [
    item.id || "",
    item.phone || "",
    item.email || "",
    item.status || "",
    item.kyc_status || "",
    item.member_level || "",
    item.registration_source || "",
    item.inviter_user_id || "",
    item.invite_code || "",
    item.invite_registered_at || "",
    item.created_at || ""
  ]);
  return [header, ...rows].map((row) => row.map(csvEscape).join(",")).join("\n");
}

function exportCurrentPageCSV() {
  const csv = buildCSVRows(users.value);
  const fileName = `admin_users_page_${new Date().toISOString().slice(0, 10)}.csv`;
  triggerCSVDownload(csv, fileName);
  message.value = `已导出当前页 CSV，共 ${users.value.length} 条`;
}

async function exportFilteredCSV() {
  exportingFiltered.value = true;
  errorMessage.value = "";
  message.value = "";

  try {
    const params = new URLSearchParams();
    if (filters.status) params.set("status", filters.status);
    if (filters.kyc_status) params.set("kyc_status", filters.kyc_status);
    if (filters.member_level.trim()) params.set("member_level", filters.member_level.trim());
    if (filters.registration_source) params.set("registration_source", filters.registration_source);

    const baseURL = (import.meta.env.VITE_API_BASE_URL || "/api/v1").replace(/\/$/, "");
    const query = params.toString();
    const requestURL = `${baseURL}/admin/users/export.csv${query ? `?${query}` : ""}`;

    const headers = {};
    const token = getAccessToken();
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }

    const response = await fetch(requestURL, { method: "GET", headers });
    if (!response.ok) {
      const text = await response.text();
      throw new Error(text || `导出失败(${response.status})`);
    }

    const blob = await response.blob();
    const blobURL = URL.createObjectURL(blob);
    const fileName = `admin_users_filtered_${new Date().toISOString().slice(0, 10)}.csv`;
    const anchor = document.createElement("a");
    anchor.href = blobURL;
    anchor.download = fileName;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(blobURL);
    message.value = "已发起筛选结果 CSV 下载";
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "导出筛选结果失败");
  } finally {
    exportingFiltered.value = false;
  }
}

function applyFilters() {
  page.value = 1;
  fetchUsers();
}

function resetFilters() {
  filters.status = "";
  filters.kyc_status = "";
  filters.member_level = "";
  filters.registration_source = "";
  page.value = 1;
  fetchUsers();
}

function handlePageChange(nextPage) {
  if (nextPage === page.value) {
    return;
  }
  page.value = nextPage;
  fetchUsers();
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "ACTIVE") return "success";
  if (normalized === "DISABLED" || normalized === "BANNED") return "danger";
  return "info";
}

function kycTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (normalized === "APPROVED") return "success";
  if (normalized === "REJECTED") return "danger";
  if (normalized === "PENDING") return "warning";
  return "info";
}

onMounted(fetchUsers);
</script>

<template>
  <div class="page">
    <div class="page-header">
      <div>
        <h1 class="page-title">用户管理</h1>
        <p class="muted">用户状态、实名状态、会员等级维护</p>
      </div>
      <el-button :loading="loading" @click="fetchUsers">刷新</el-button>
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

    <div class="card" style="margin-bottom: 12px">
      <div class="toolbar">
        <el-select v-model="filters.status" clearable placeholder="全部用户状态" style="width: 160px">
          <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-select v-model="filters.kyc_status" clearable placeholder="全部 KYC 状态" style="width: 170px">
          <el-option v-for="item in kycStatusOptions" :key="item" :label="item" :value="item" />
        </el-select>
        <el-input v-model="filters.member_level" clearable placeholder="会员等级，如 VIP1" style="width: 180px" />
        <el-select
          v-model="filters.registration_source"
          clearable
          placeholder="全部注册来源"
          style="width: 160px"
        >
          <el-option
            v-for="item in registrationSourceOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </el-select>
        <el-button :loading="exportingFiltered" @click="exportFilteredCSV">导出筛选CSV</el-button>
        <el-button @click="exportCurrentPageCSV">导出当前页CSV</el-button>
        <el-button type="primary" plain @click="applyFilters">查询</el-button>
        <el-button @click="resetFilters">重置</el-button>
      </div>
    </div>

    <div class="card" style="margin-bottom: 12px">
      <div class="summary-cards">
        <div v-for="item in sourceSummaryCards" :key="item.label" class="summary-card">
          <p>{{ item.label }}</p>
          <h3>{{ item.value }}</h3>
        </div>
      </div>
    </div>

    <div v-if="canEditUsers" class="card" style="margin-bottom: 12px">
      <div class="section-header">
        <h3 style="margin: 0">批量操作</h3>
        <el-text type="info">已勾选 {{ selectedCount }} 个用户</el-text>
      </div>
      <div class="toolbar" style="margin-bottom: 0">
        <div class="field-stack">
          <el-select v-model="batchStatus" style="width: 140px">
            <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
          </el-select>
          <el-popconfirm
            width="300"
            :title="`确认将选中用户状态批量更新为 ${batchStatus || '-'} 吗？`"
            @confirm="handleBatchUpdateStatus"
          >
            <template #reference>
              <el-button :loading="batchUpdatingStatus" :disabled="!canBatchUpdateStatus">批量更新状态</el-button>
            </template>
          </el-popconfirm>
        </div>

        <div class="field-stack">
          <el-select v-model="batchKYCStatus" style="width: 140px">
            <el-option v-for="item in kycStatusOptions" :key="item" :label="item" :value="item" />
          </el-select>
          <el-popconfirm
            width="300"
            :title="`确认将选中用户 KYC 状态批量更新为 ${batchKYCStatus || '-'} 吗？`"
            @confirm="handleBatchUpdateKYC"
          >
            <template #reference>
              <el-button :loading="batchUpdatingKYC" :disabled="!canBatchUpdateKYC">批量更新KYC</el-button>
            </template>
          </el-popconfirm>
        </div>

        <div class="field-stack">
          <el-input v-model="batchMemberLevel" placeholder="目标会员等级" style="width: 140px" />
          <el-popconfirm
            width="300"
            :title="`确认将选中用户会员等级批量更新为 ${batchMemberLevel || '-'} 吗？`"
            @confirm="handleBatchUpdateLevel"
          >
            <template #reference>
              <el-button :loading="batchUpdatingLevel" :disabled="!canBatchUpdateLevel">批量更新等级</el-button>
            </template>
          </el-popconfirm>
        </div>

        <el-button @click="clearSelection">清空勾选</el-button>
      </div>
    </div>

    <div class="card">
      <el-table
        ref="userTableRef"
        :data="users"
        row-key="id"
        border
        stripe
        v-loading="loading"
        empty-text="暂无用户数据"
        @selection-change="handleSelectionChange"
      >
        <el-table-column v-if="canEditUsers" type="selection" width="52" reserve-selection />
        <el-table-column prop="id" label="用户ID" min-width="160" />
        <el-table-column prop="phone" label="手机号" min-width="130" />
        <el-table-column label="邮箱" min-width="180">
          <template #default="{ row }">
            {{ row.email || "-" }}
          </template>
        </el-table-column>
        <el-table-column label="注册来源" min-width="120">
          <template #default="{ row }">
            <el-tag :type="registrationSourceTagType(row.registration_source)">
              {{ mapRegistrationSource(row.registration_source) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="邀请关系" min-width="220">
          <template #default="{ row }">
            <div class="stack-meta">
              <span>邀请人：{{ row.inviter_user_id || "-" }}</span>
              <span>邀请码：{{ row.invite_code || "-" }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="用户状态" min-width="300">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-tag :type="statusTagType(row.status)">{{ row.status || "-" }}</el-tag>
              <template v-if="canEditUsers">
                <el-select v-model="draftStatusMap[row.id]" style="width: 120px">
                  <el-option v-for="item in statusOptions" :key="item" :label="item" :value="item" />
                </el-select>
                <el-button size="small" @click="handleUpdateStatus(row)">保存</el-button>
              </template>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="KYC 状态" min-width="300">
          <template #default="{ row }">
            <div class="inline-actions">
              <el-tag :type="kycTagType(row.kyc_status)">{{ row.kyc_status || "-" }}</el-tag>
              <template v-if="canEditUsers">
                <el-select v-model="draftKYCMap[row.id]" style="width: 120px">
                  <el-option v-for="item in kycStatusOptions" :key="item" :label="item" :value="item" />
                </el-select>
                <el-button size="small" @click="handleUpdateKYC(row)">保存</el-button>
              </template>
            </div>
          </template>
        </el-table-column>

        <el-table-column label="会员等级" min-width="260">
          <template #default="{ row }">
            <div class="inline-actions">
              <template v-if="canEditUsers">
                <el-input v-model="draftLevelMap[row.id]" style="width: 120px" />
                <el-button size="small" @click="handleUpdateMemberLevel(row)">保存</el-button>
              </template>
              <span v-else>{{ row.member_level || "-" }}</span>
            </div>
          </template>
        </el-table-column>

        <el-table-column prop="created_at" label="创建时间" min-width="180" />
        <el-table-column label="操作" width="190" fixed="right">
          <template #default="{ row }">
            <div class="inline-actions inline-actions--compact">
              <el-button size="small" @click="openUserCenter(row)">查看</el-button>
              <el-button
                v-if="canEditUsers"
                size="small"
                type="warning"
                plain
                @click="openResetPasswordDialog(row)"
              >
                修改密码
              </el-button>
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
          @current-change="handlePageChange"
        />
      </div>
    </div>

    <el-dialog v-model="batchResultVisible" :title="batchResultTitle" width="760px" destroy-on-close>
      <div class="batch-result-summary">
        <el-tag type="success">成功 {{ batchResultStats.success }}</el-tag>
        <el-tag type="danger">失败 {{ batchResultStats.failed }}</el-tag>
        <el-tag type="warning">跳过 {{ batchResultStats.skipped }}</el-tag>
        <el-text type="info">总计 {{ batchResultStats.total }} 条</el-text>
      </div>
      <div class="batch-result-toolbar">
        <el-select v-model="batchResultFilter" style="width: 180px">
          <el-option label="查看全部结果" value="all" />
          <el-option label="仅看失败" value="failed" />
          <el-option label="仅看跳过" value="skipped" />
        </el-select>
        <el-button
          :disabled="failedBatchRows.length <= 0"
          :loading="copyingFailedDetails"
          @click="copyFailedDetails"
        >
          复制失败明细
        </el-button>
      </div>
      <el-table :data="displayBatchResultRows" border stripe max-height="360" empty-text="当前筛选条件下暂无结果">
        <el-table-column prop="id" label="用户ID" min-width="140" />
        <el-table-column prop="action" label="动作" min-width="130" />
        <el-table-column prop="target" label="目标值" min-width="120" />
        <el-table-column label="结果" min-width="110">
          <template #default="{ row }">
            <el-tag :type="row.result === 'SUCCESS' ? 'success' : row.result === 'SKIPPED' ? 'warning' : 'danger'">
              {{ row.result }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reason" label="说明" min-width="260" />
      </el-table>
      <template #footer>
        <el-button
          type="warning"
          plain
          :disabled="failedBatchRows.length <= 0"
          :loading="retryingFailed"
          @click="retryFailedBatchRows"
        >
          重试失败任务
        </el-button>
        <el-button type="primary" @click="batchResultVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="centerDialogVisible"
      :title="`客户中心详情 · ${centerTargetUserID || '-'}`"
      width="980px"
      destroy-on-close
    >
      <el-alert
        v-if="centerErrorMessage"
        :title="centerErrorMessage"
        type="error"
        show-icon
        style="margin-bottom: 10px"
      />

      <div v-loading="centerLoading">
        <el-descriptions :column="2" border style="margin-bottom: 10px">
          <el-descriptions-item label="用户ID">
            {{ centerData.user_profile?.id || centerTargetUserID || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="会员等级">
            {{ centerData.membership_quota?.member_level || centerData.user_profile?.member_level || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="KYC状态">
            {{ mapDisplayStatus(centerData.user_profile?.kyc_status, "kyc") }}
          </el-descriptions-item>
          <el-descriptions-item label="邮箱">
            {{ centerData.user_profile?.email || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="注册来源">
            {{ mapRegistrationSource(centerData.user_profile?.registration_source) }}
          </el-descriptions-item>
          <el-descriptions-item label="邀请人ID">
            {{ centerData.user_profile?.inviter_user_id || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="邀请码">
            {{ centerData.user_profile?.invite_code || "-" }}
          </el-descriptions-item>
          <el-descriptions-item label="文档阅读剩余">
            {{ centerData.membership_quota?.doc_read_remaining ?? 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="资讯订阅剩余">
            {{ centerData.membership_quota?.news_subscribe_remaining ?? 0 }}
          </el-descriptions-item>
        </el-descriptions>

        <div class="center-summary-row">
          <el-tag type="success">
            已支付订单 {{ centerData.payment_summary?.paid_order_count ?? 0 }}
          </el-tag>
          <el-tag type="warning">
            待处理订单 {{ centerData.payment_summary?.pending_order_count ?? 0 }}
          </el-tag>
          <el-tag type="info">
            支付总额 {{ formatCenterAmount(centerData.payment_summary?.paid_amount_total ?? 0) }}
          </el-tag>
          <el-tag type="info">
            阅读记录 {{ centerData.reading_summary?.browse_count ?? 0 }}
          </el-tag>
          <el-tag type="success">
            生效订阅 {{ centerData.subscription_summary?.active_count ?? 0 }}
          </el-tag>
        </div>

        <el-tabs v-model="centerActiveTab">
          <el-tab-pane label="VIP情况" name="vip">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="配额周期">
                {{ centerData.membership_quota?.period_key || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="重置周期">
                {{ centerData.membership_quota?.reset_cycle || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="文档阅读">
                {{ centerData.membership_quota?.doc_read_used ?? 0 }} /
                {{ centerData.membership_quota?.doc_read_limit ?? 0 }}
              </el-descriptions-item>
              <el-descriptions-item label="资讯订阅">
                {{ centerData.membership_quota?.news_subscribe_used ?? 0 }} /
                {{ centerData.membership_quota?.news_subscribe_limit ?? 0 }}
              </el-descriptions-item>
              <el-descriptions-item label="下次重置">
                {{ formatCenterDateTime(centerData.membership_quota?.reset_at) }}
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>

          <el-tab-pane label="支付情况" name="payment">
            <p class="center-subtitle">会员订单</p>
            <el-table :data="centerMembershipOrders" border stripe max-height="240" empty-text="暂无会员订单">
              <el-table-column prop="order_no" label="订单号" min-width="150">
                <template #default="{ row }">{{ row.order_no || row.id || "-" }}</template>
              </el-table-column>
              <el-table-column prop="product_id" label="产品ID" min-width="120" />
              <el-table-column prop="pay_channel" label="支付方式" min-width="110" />
              <el-table-column label="金额" min-width="90">
                <template #default="{ row }">{{ formatCenterAmount(row.amount) }}</template>
              </el-table-column>
              <el-table-column label="状态" min-width="90">
                <template #default="{ row }">
                  <el-tag :type="statusTagByType(row.status, 'payment')">
                    {{ mapDisplayStatus(row.status, "payment") }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="支付时间" min-width="170">
                <template #default="{ row }">
                  {{ formatCenterDateTime(row.paid_at || row.created_at) }}
                </template>
              </el-table-column>
            </el-table>

            <p class="center-subtitle">充值记录</p>
            <el-table :data="centerRechargeRecords" border stripe max-height="220" empty-text="暂无充值记录">
              <el-table-column prop="order_no" label="订单号" min-width="150" />
              <el-table-column label="金额" min-width="90">
                <template #default="{ row }">{{ formatCenterAmount(row.amount) }}</template>
              </el-table-column>
              <el-table-column prop="pay_channel" label="支付方式" min-width="110" />
              <el-table-column label="状态" min-width="90">
                <template #default="{ row }">
                  <el-tag :type="statusTagByType(row.status, 'payment')">
                    {{ mapDisplayStatus(row.status, "payment") }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="支付时间" min-width="170">
                <template #default="{ row }">{{ formatCenterDateTime(row.paid_at) }}</template>
              </el-table-column>
            </el-table>
          </el-tab-pane>

          <el-tab-pane label="阅读情况" name="reading">
            <el-table :data="centerBrowseHistory" border stripe max-height="360" empty-text="暂无阅读记录">
              <el-table-column prop="content_type" label="类型" min-width="90" />
              <el-table-column prop="title" label="标题" min-width="260" />
              <el-table-column prop="source_page" label="来源页面" min-width="120" />
              <el-table-column label="阅读时间" min-width="170">
                <template #default="{ row }">{{ formatCenterDateTime(row.viewed_at) }}</template>
              </el-table-column>
            </el-table>
          </el-tab-pane>

          <el-tab-pane label="订阅情况" name="subscription">
            <el-table :data="centerSubscriptions" border stripe max-height="360" empty-text="暂无订阅记录">
              <el-table-column prop="type" label="订阅类型" min-width="140" />
              <el-table-column prop="scope" label="范围" min-width="120" />
              <el-table-column label="频率" min-width="140">
                <template #default="{ row }">
                  <template v-if="canEditUsers">
                    <el-select
                      v-model="centerSubscriptionDraft[row.id].frequency"
                      size="small"
                      style="width: 110px"
                    >
                      <el-option
                        v-for="item in subscriptionFrequencyOptions"
                        :key="item"
                        :label="mapSubscriptionFrequency(item)"
                        :value="item"
                      />
                    </el-select>
                  </template>
                  <span v-else>{{ mapSubscriptionFrequency(row.frequency) }}</span>
                </template>
              </el-table-column>
              <el-table-column label="状态" min-width="100">
                <template #default="{ row }">
                  <template v-if="canEditUsers">
                    <el-select
                      v-model="centerSubscriptionDraft[row.id].status"
                      size="small"
                      style="width: 110px"
                    >
                      <el-option
                        v-for="item in subscriptionStatusOptions"
                        :key="item"
                        :label="mapDisplayStatus(item, 'subscription')"
                        :value="item"
                      />
                    </el-select>
                  </template>
                  <span v-else>{{ mapDisplayStatus(row.status, "subscription") }}</span>
                </template>
              </el-table-column>
              <el-table-column v-if="canEditUsers" label="操作" min-width="110">
                <template #default="{ row }">
                  <el-button
                    size="small"
                    type="primary"
                    :loading="centerSavingSubscriptionID === row.id"
                    @click="handleSaveCenterSubscription(row)"
                  >
                    保存
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </el-tab-pane>

          <el-tab-pane label="邀请关系" name="invite">
            <div class="center-summary-row">
              <el-tag v-for="item in centerInviteSummaryCards" :key="item.label" type="info">
                {{ item.label }} {{ item.value }}
              </el-tag>
            </div>

            <p class="center-subtitle">分享链接</p>
            <el-table :data="centerShareLinks" border stripe max-height="200" empty-text="暂无分享链接">
              <el-table-column prop="invite_code" label="邀请码" min-width="120" />
              <el-table-column prop="channel" label="渠道" min-width="110" />
              <el-table-column prop="status" label="状态" min-width="100" />
              <el-table-column label="过期时间" min-width="170">
                <template #default="{ row }">
                  {{ formatCenterDateTime(row.expired_at) }}
                </template>
              </el-table-column>
              <el-table-column prop="url" label="分享链接" min-width="280" />
            </el-table>

            <p class="center-subtitle">邀请时间线</p>
            <el-timeline class="center-invite-timeline">
              <el-timeline-item
                v-for="item in centerInviteTimeline"
                :key="item.id"
                :timestamp="item.registerAt"
                placement="top"
              >
                <div class="timeline-title">
                  邀请用户 {{ item.inviteeUserID }} · {{ item.statusText }}
                </div>
                <div class="timeline-meta">
                  <span>首单支付：{{ item.firstPayAt }}</span>
                  <span>风控：{{ item.riskFlag }}</span>
                </div>
              </el-timeline-item>
            </el-timeline>
          </el-tab-pane>

          <el-tab-pane label="其他信息" name="other">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="手机号">
                {{ centerData.user_profile?.phone || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="邮箱">
                {{ centerData.user_profile?.email || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="记录读取上限">
                {{ centerMembershipOrders.length + centerRechargeRecords.length + centerBrowseHistory.length }}
              </el-descriptions-item>
              <el-descriptions-item label="接口返回订阅条数">
                {{ centerSubscriptions.length }}
              </el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
        </el-tabs>
      </div>

      <template #footer>
        <el-button @click="centerDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="passwordDialogVisible"
      title="修改用户密码"
      width="420px"
      destroy-on-close
      @closed="resetPasswordDialogState"
    >
      <el-alert
        v-if="passwordDialogError"
        :title="passwordDialogError"
        type="error"
        show-icon
        style="margin-bottom: 12px"
      />

      <div class="password-dialog-stack">
        <div class="password-dialog-meta">
          <span>用户ID</span>
          <strong>{{ passwordForm.userID || "-" }}</strong>
        </div>
        <div class="password-dialog-meta">
          <span>账号</span>
          <strong>{{ passwordForm.account || "-" }}</strong>
        </div>
        <el-input
          v-model="passwordForm.password"
          type="password"
          show-password
          clearable
          maxlength="64"
          placeholder="请输入新密码，至少8位"
        />
        <el-input
          v-model="passwordForm.confirmPassword"
          type="password"
          show-password
          clearable
          maxlength="64"
          placeholder="请再次输入新密码"
          @keyup.enter="handleResetUserPassword"
        />
      </div>

      <template #footer>
        <el-button @click="passwordDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="passwordSubmitting" @click="handleResetUserPassword">
          确认修改
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.inline-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.inline-actions--compact {
  flex-wrap: wrap;
}

.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 10px;
  flex-wrap: wrap;
}

.field-stack {
  display: flex;
  align-items: center;
  gap: 8px;
}

.batch-result-summary {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.batch-result-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.center-summary-row {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 10px;
}

.center-subtitle {
  margin: 10px 0 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-regular);
}

.summary-cards {
  display: grid;
  gap: 10px;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
}

.summary-card {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  padding: 10px 12px;
  background: var(--el-fill-color-extra-light);
}

.summary-card p {
  margin: 0;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.summary-card h3 {
  margin: 6px 0 0;
  font-size: 20px;
  font-weight: 600;
}

.stack-meta {
  display: grid;
  gap: 4px;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.center-invite-timeline {
  margin-top: 8px;
}

.timeline-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.timeline-meta {
  margin-top: 4px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.password-dialog-stack {
  display: grid;
  gap: 12px;
}

.password-dialog-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 8px;
  background: var(--el-fill-color-extra-light);
}

.password-dialog-meta span {
  color: var(--el-text-color-secondary);
  font-size: 12px;
}

.password-dialog-meta strong {
  color: var(--el-text-color-primary);
  font-size: 13px;
  font-weight: 600;
}
</style>
