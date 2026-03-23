<script setup>
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import FuturesSelectionModuleShell from "../../components/FuturesSelectionModuleShell.vue";
import {
  approveFuturesSelectionReview,
  getFuturesSelectionRun,
  getFuturesStrategyEnginePublishRecord,
  listFuturesSelectionPortfolio,
  listFuturesSelectionRuns,
  rejectFuturesSelectionReview
} from "../../api/admin";
import {
  formatFuturesSelectionDateTime,
  formatFuturesSelectionLabel,
  formatFuturesSelectionDirection,
  formatFuturesSelectionMarketRegime,
  formatFuturesSelectionReviewStatus,
  formatFuturesSelectionRiskLevel,
  formatFuturesSelectionRunStatus
} from "../../lib/futures-selection";
import {
  extractReviewConflictReason,
  resolveReviewDialogMeta
} from "../../lib/review-action-dialog";
import { hasPermission } from "../../lib/session";

const route = useRoute();
const router = useRouter();
const canManage = hasPermission("futures_selection.manage");

const filters = ref({
  review_status: "PENDING"
});
const queueLoading = ref(false);
const detailLoading = ref(false);
const actionLoading = ref(false);
const runQueue = ref([]);
const selectedRunID = ref("");
const runDetail = ref(null);
const portfolioEntries = ref([]);
const publishRecord = ref(null);
const publishLoadError = ref("");
const reviewDialogVisible = ref(false);
const reviewDialogMode = ref("approve");
const reviewDialogError = ref("");
const reviewDialogForm = reactive({
  reviewNote: "",
  overrideReason: ""
});

const publishedSnapshot = computed(() => {
  const snapshot = runDetail.value?.review?.published_contract_snapshot;
  return Array.isArray(snapshot) ? snapshot : [];
});

const reviewStatusSummary = computed(() => {
  const pending = runQueue.value.filter((item) => item.review_status === "PENDING").length;
  const approved = runQueue.value.filter((item) => item.review_status === "APPROVED").length;
  const rejected = runQueue.value.filter((item) => item.review_status === "REJECTED").length;
  return { pending, approved, rejected };
});
const reviewWarningMessages = computed(() =>
  Array.isArray(runDetail.value?.warning_messages) ? runDetail.value.warning_messages : []
);
const reviewDialogTitle = computed(() => resolveReviewDialogMeta(reviewDialogMode.value).title);
const reviewDialogPrimaryText = computed(() => resolveReviewDialogMeta(reviewDialogMode.value).primaryText);
const reviewDialogPrimaryType = computed(() => resolveReviewDialogMeta(reviewDialogMode.value).primaryType);
const reviewDialogSummaryTone = computed(() => resolveReviewDialogMeta(reviewDialogMode.value).summaryTone);

function tagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "APPROVED") return "success";
  if (normalized === "REJECTED") return "danger";
  if (normalized === "PENDING") return "warning";
  if (normalized === "SUCCEEDED") return "success";
  if (normalized === "FAILED") return "danger";
  return "info";
}

function syncRouteQuery(runID) {
  const nextQuery = { ...route.query };
  if (runID) {
    nextQuery.run_id = runID;
  } else {
    delete nextQuery.run_id;
  }
  router.replace({ query: nextQuery }).catch(() => {});
}

async function fetchRunQueue(preferredRunID = "") {
  queueLoading.value = true;
  try {
    const data = await listFuturesSelectionRuns({
      status: "SUCCEEDED",
      review_status: filters.value.review_status || undefined,
      page: 1,
      page_size: 50
    });
    runQueue.value = Array.isArray(data?.items) ? data.items : [];
    const queryRunID = String(preferredRunID || route.query.run_id || "").trim();
    const currentRunID = String(selectedRunID.value || "").trim();
    const nextRunID =
      (queryRunID && runQueue.value.find((item) => item.run_id === queryRunID)?.run_id) ||
      (currentRunID && runQueue.value.find((item) => item.run_id === currentRunID)?.run_id) ||
      runQueue.value[0]?.run_id ||
      "";
    selectedRunID.value = nextRunID;
    syncRouteQuery(nextRunID);
  } catch (error) {
    ElMessage.error(error?.message || "加载期货审核队列失败");
    runQueue.value = [];
    selectedRunID.value = "";
    syncRouteQuery("");
  } finally {
    queueLoading.value = false;
  }
}

async function fetchRunDetail() {
  if (!selectedRunID.value) {
    runDetail.value = null;
    portfolioEntries.value = [];
    publishRecord.value = null;
    publishLoadError.value = "";
    return;
  }
  detailLoading.value = true;
  try {
    const [run, portfolio] = await Promise.all([
      getFuturesSelectionRun(selectedRunID.value),
      listFuturesSelectionPortfolio(selectedRunID.value)
    ]);
    runDetail.value = run;
    portfolioEntries.value = Array.isArray(portfolio?.items) ? portfolio.items : [];
    const publishID = String(run?.review?.publish_id || run?.latest_publish_id || "").trim();
    publishRecord.value = null;
    publishLoadError.value = "";
    if (publishID) {
      try {
        publishRecord.value = await getFuturesStrategyEnginePublishRecord(publishID);
      } catch (error) {
        publishLoadError.value = error?.message || "加载期货发布归档失败";
      }
    }
  } catch (error) {
    ElMessage.error(error?.message || "加载期货审核详情失败");
    runDetail.value = null;
    portfolioEntries.value = [];
    publishRecord.value = null;
    publishLoadError.value = "";
  } finally {
    detailLoading.value = false;
  }
}

async function refreshPage(preferredRunID = "") {
  await fetchRunQueue(preferredRunID);
  await fetchRunDetail();
}

function resetReviewDialog() {
  reviewDialogMode.value = "approve";
  reviewDialogError.value = "";
  reviewDialogForm.reviewNote = "";
  reviewDialogForm.overrideReason = "";
}

function openReviewDialog(mode) {
  if (!selectedRunID.value || !runDetail.value) {
    return;
  }
  reviewDialogMode.value = mode;
  reviewDialogError.value = "";
  if (mode === "approve") {
    reviewDialogForm.reviewNote = "审核工作台确认通过并发布";
    reviewDialogForm.overrideReason = "";
  } else if (mode === "force") {
    reviewDialogForm.reviewNote = "人工确认允许覆盖发布";
    reviewDialogForm.overrideReason = "";
  } else {
    reviewDialogForm.reviewNote = "";
    reviewDialogForm.overrideReason = "";
  }
  reviewDialogVisible.value = true;
}

function closeReviewDialog() {
  if (actionLoading.value) {
    return;
  }
  reviewDialogVisible.value = false;
  resetReviewDialog();
}

async function submitReviewDialog() {
  if (!selectedRunID.value) {
    return;
  }
  if (reviewDialogMode.value === "blocked") {
    reviewDialogMode.value = "force";
    reviewDialogError.value = "";
    if (!reviewDialogForm.reviewNote.trim()) {
      reviewDialogForm.reviewNote = "人工确认允许覆盖发布";
    }
    return;
  }
  const reviewNote = reviewDialogForm.reviewNote.trim();
  const overrideReason = reviewDialogForm.overrideReason.trim();
  if (!reviewNote) {
    reviewDialogError.value = reviewDialogMode.value === "reject" ? "请填写驳回原因" : "请填写审核说明";
    return;
  }
  if (reviewDialogMode.value === "force" && !overrideReason) {
    reviewDialogError.value = "请填写人工覆盖原因";
    return;
  }
  actionLoading.value = true;
  try {
    reviewDialogError.value = "";
    if (reviewDialogMode.value === "reject") {
      await rejectFuturesSelectionReview(selectedRunID.value, {
        review_note: reviewNote
      });
      ElMessage.success("审核已驳回");
    } else {
      const force = reviewDialogMode.value === "force";
      await approveFuturesSelectionReview(selectedRunID.value, {
        review_note: reviewNote,
        force,
        override_reason: force ? overrideReason : ""
      });
      ElMessage.success(force ? "已完成强制发布" : "审核已通过并发布");
    }
    reviewDialogVisible.value = false;
    resetReviewDialog();
    await refreshPage(selectedRunID.value);
  } catch (error) {
    const conflictReason = extractReviewConflictReason(error);
    if (conflictReason) {
      reviewDialogMode.value = "blocked";
      reviewDialogError.value = conflictReason;
      return;
    }
    reviewDialogError.value =
      error?.message || (reviewDialogMode.value === "reject" ? "驳回失败" : "审核通过失败");
  } finally {
    actionLoading.value = false;
  }
}

watch(
  () => route.query.run_id,
  async (runID) => {
    const normalized = String(runID || "").trim();
    if (!normalized || normalized === selectedRunID.value) {
      return;
    }
    selectedRunID.value = normalized;
    await fetchRunDetail();
  }
);

watch(selectedRunID, async (value, oldValue) => {
  if (value === oldValue) {
    return;
  }
  syncRouteQuery(value);
  await fetchRunDetail();
});

onMounted(async () => {
  await refreshPage(String(route.query.run_id || "").trim());
});
</script>

<template>
  <FuturesSelectionModuleShell
    title="智能期货审核与发布"
    description="把待审核 run、最终合约组合和发布归档拆成独立工作台，研究页继续看证据，这里专注审核决策与发布结果。"
  >
    <template #actions>
      <div class="toolbar review-toolbar">
        <el-select
          v-model="filters.review_status"
          placeholder="审核状态"
          clearable
          style="width: 160px"
          @change="refreshPage(selectedRunID)"
        >
          <el-option label="待审核" value="PENDING" />
          <el-option label="已通过" value="APPROVED" />
          <el-option label="已驳回" value="REJECTED" />
        </el-select>
        <el-button :loading="queueLoading || detailLoading" @click="refreshPage(selectedRunID)">刷新</el-button>
        <el-button
          v-if="selectedRunID"
          type="primary"
          plain
          @click="router.push({ name: 'futures-selection-candidates', query: { run_id: selectedRunID } })"
        >
          打开候选与组合
        </el-button>
      </div>
    </template>

    <div class="review-layout">
      <div class="card queue-card" v-loading="queueLoading">
        <div class="toolbar" style="justify-content: space-between; margin-bottom: 12px">
          <div class="card-title">审核队列</div>
          <div class="tag-wrap">
            <el-tag type="warning">待审核 {{ reviewStatusSummary.pending }}</el-tag>
            <el-tag type="success">已通过 {{ reviewStatusSummary.approved }}</el-tag>
            <el-tag type="danger">已驳回 {{ reviewStatusSummary.rejected }}</el-tag>
          </div>
        </div>
        <el-table
          :data="runQueue"
          border
          stripe
          size="small"
          height="620"
          empty-text="当前筛选条件下没有期货审核记录"
          row-key="run_id"
          highlight-current-row
          :current-row-key="selectedRunID"
          @row-click="(row) => selectedRunID = row.run_id"
        >
          <el-table-column prop="trade_date" label="交易日" min-width="104" />
          <el-table-column prop="run_id" label="运行编号" min-width="190" />
          <el-table-column label="审核状态" min-width="98">
            <template #default="{ row }">
              <el-tag :type="tagType(row.review_status)">
                {{ formatFuturesSelectionReviewStatus(row.review_status || "PENDING") }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="市场状态" min-width="112">
            <template #default="{ row }">{{ formatFuturesSelectionMarketRegime(row.market_regime) }}</template>
          </el-table-column>
          <el-table-column prop="publish_count" label="发布次数" min-width="90" />
          <el-table-column label="最近发布时间" min-width="160">
            <template #default="{ row }">{{ formatFuturesSelectionDateTime(row.latest_publish_at) }}</template>
          </el-table-column>
        </el-table>
      </div>

      <div class="detail-stack">
        <div class="card" v-loading="detailLoading">
          <template v-if="runDetail">
            <div class="toolbar" style="justify-content: space-between; flex-wrap: wrap">
              <div class="tag-wrap">
                <el-tag type="info">运行 {{ runDetail.run_id }}</el-tag>
                <el-tag :type="tagType(runDetail.status)">{{ formatFuturesSelectionRunStatus(runDetail.status) }}</el-tag>
                <el-tag :type="tagType(runDetail.review_status)">
                  {{ formatFuturesSelectionReviewStatus(runDetail.review_status || "PENDING") }}
                </el-tag>
                <el-tag type="warning">{{ formatFuturesSelectionMarketRegime(runDetail.market_regime) }}</el-tag>
                <el-tag type="success">组合 {{ runDetail.selected_count || 0 }}</el-tag>
                <el-tag type="info">发布 v{{ runDetail.latest_publish_version || 0 }}</el-tag>
              </div>
              <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
                <el-button
                  v-if="canManage && runDetail.review_status === 'PENDING'"
                  type="success"
                  :loading="actionLoading"
                  @click="openReviewDialog('approve')"
                >
                  审核通过并发布
                </el-button>
                <el-button
                  v-if="canManage && runDetail.review_status === 'PENDING'"
                  type="warning"
                  :loading="actionLoading"
                  @click="openReviewDialog('force')"
                >
                  强制发布
                </el-button>
                <el-button
                  v-if="canManage && runDetail.review_status === 'PENDING'"
                  type="danger"
                  plain
                  :loading="actionLoading"
                  @click="openReviewDialog('reject')"
                >
                  驳回
                </el-button>
              </div>
            </div>

            <el-descriptions :column="2" border size="small" style="margin-top: 12px">
              <el-descriptions-item label="结果摘要" :span="2">
                {{ runDetail.result_summary || "当前没有额外摘要。" }}
              </el-descriptions-item>
              <el-descriptions-item label="配置方案">
                {{ runDetail.profile_id }} / v{{ runDetail.profile_version }}
              </el-descriptions-item>
              <el-descriptions-item label="模板">
                {{ runDetail.template_name || "未指定模板" }}
              </el-descriptions-item>
              <el-descriptions-item label="审核人">
                {{ runDetail.review?.reviewer || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="审核备注">
                {{ runDetail.review?.review_note || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="最近发布记录">
                {{ runDetail.review?.publish_id || runDetail.latest_publish_id || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="最近发布时间">
                {{ formatFuturesSelectionDateTime(runDetail.review?.approved_at || runDetail.latest_publish_at) }}
              </el-descriptions-item>
              <el-descriptions-item label="版本差异" :span="2">
                新增：{{ (runDetail.compare_summary?.added_contracts || []).join("、") || "无" }}
                / 移除：{{ (runDetail.compare_summary?.removed_contracts || []).join("、") || "无" }}
              </el-descriptions-item>
            </el-descriptions>

            <div class="card tone-card" v-if="Array.isArray(runDetail.warning_messages) && runDetail.warning_messages.length">
              <div class="card-title">运行提醒</div>
              <div class="tag-wrap">
                <el-tag v-for="warning in runDetail.warning_messages" :key="warning" type="warning">
                  {{ warning }}
                </el-tag>
              </div>
            </div>
          </template>
          <el-empty v-else description="先从左侧选择一个待审核 run" :image-size="88" />
        </div>

        <div class="detail-grid" v-if="runDetail">
          <div class="card">
            <div class="card-title">当前组合</div>
            <el-table :data="portfolioEntries" border stripe size="small" height="280" empty-text="暂无组合结果">
              <el-table-column prop="rank" label="排名" min-width="68" />
              <el-table-column prop="contract" label="合约" min-width="120" />
              <el-table-column prop="name" label="名称" min-width="120" />
              <el-table-column label="方向" min-width="90">
                <template #default="{ row }">{{ formatFuturesSelectionDirection(row.direction) }}</template>
              </el-table-column>
              <el-table-column label="角色" min-width="90">
                <template #default="{ row }">{{ formatFuturesSelectionLabel(row.portfolio_role || "SATELLITE") }}</template>
              </el-table-column>
              <el-table-column label="风险" min-width="90">
                <template #default="{ row }">{{ formatFuturesSelectionRiskLevel(row.risk_level) }}</template>
              </el-table-column>
            </el-table>
          </div>

          <div class="card">
            <div class="card-title">已发布快照</div>
            <el-table
              v-if="publishedSnapshot.length"
              :data="publishedSnapshot"
              border
              stripe
              size="small"
              height="280"
            >
              <el-table-column prop="rank" label="排名" min-width="68" />
              <el-table-column prop="contract" label="合约" min-width="120" />
              <el-table-column prop="name" label="名称" min-width="120" />
              <el-table-column label="方向" min-width="90">
                <template #default="{ row }">{{ formatFuturesSelectionDirection(row.direction) }}</template>
              </el-table-column>
              <el-table-column label="角色" min-width="90">
                <template #default="{ row }">{{ formatFuturesSelectionLabel(row.portfolio_role || "SATELLITE") }}</template>
              </el-table-column>
            </el-table>
            <el-empty v-else description="当前 run 还没有固化的发布快照" :image-size="72" />
          </div>
        </div>

        <div class="card" v-if="runDetail">
          <div class="card-title">发布归档摘要</div>
          <el-alert
            v-if="publishLoadError"
            type="warning"
            show-icon
            :closable="false"
            :title="publishLoadError"
            style="margin-bottom: 12px"
          />
          <el-descriptions v-if="publishRecord" :column="2" border size="small">
            <el-descriptions-item label="发布记录">{{ publishRecord.publish_id }}</el-descriptions-item>
            <el-descriptions-item label="发布版本">v{{ publishRecord.version || 0 }}</el-descriptions-item>
            <el-descriptions-item label="归档时间">
              {{ formatFuturesSelectionDateTime(publishRecord.created_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="入选数量">
              {{ publishRecord.selected_count || 0 }}
            </el-descriptions-item>
            <el-descriptions-item label="发布方式">
              {{ publishRecord.replay?.force_publish ? "人工覆盖发布" : "按策略发布" }}
            </el-descriptions-item>
            <el-descriptions-item label="警告数量">
              {{ publishRecord.replay?.warning_count || 0 }}
            </el-descriptions-item>
            <el-descriptions-item label="摘要" :span="2">
              {{ publishRecord.report_summary || "当前没有额外归档摘要。" }}
            </el-descriptions-item>
            <el-descriptions-item label="覆盖原因" :span="2">
              {{ publishRecord.replay?.override_reason || "-" }}
            </el-descriptions-item>
          </el-descriptions>
          <el-empty v-else description="当前还没有可读取的发布归档。" :image-size="72" />
        </div>
      </div>
    </div>

    <el-dialog
      v-model="reviewDialogVisible"
      :title="reviewDialogTitle"
      width="620px"
      class="review-action-dialog"
      destroy-on-close
      :close-on-click-modal="!actionLoading"
      :close-on-press-escape="!actionLoading"
      @close="closeReviewDialog"
    >
      <div class="review-dialog-body">
        <div class="review-summary" :class="`review-summary--${reviewDialogSummaryTone}`">
          <div class="review-summary__title">
            {{ runDetail?.template_name || "未指定模板" }} / {{ formatFuturesSelectionMarketRegime(runDetail?.market_regime) }}
          </div>
          <div class="review-summary__meta">
            <span>运行 {{ runDetail?.run_id || "-" }}</span>
            <span>候选 {{ runDetail?.candidate_count || 0 }}</span>
            <span>组合 {{ runDetail?.selected_count || 0 }}</span>
            <span>警告 {{ runDetail?.warning_count || 0 }}</span>
          </div>
        </div>

        <el-alert
          v-if="reviewDialogMode === 'approve'"
          type="info"
          :closable="false"
          show-icon
          title="将按默认发布策略检查风险阈值；若命中拦截条件，系统会先阻断本次发布。"
        />
        <el-alert
          v-else-if="reviewDialogMode === 'force'"
          type="warning"
          :closable="false"
          show-icon
          title="强制发布会绕过默认阈值拦截，请确认风险可控并留下清晰的人工覆盖原因。"
        />
        <el-alert
          v-else-if="reviewDialogMode === 'reject'"
          type="error"
          :closable="false"
          show-icon
          title="驳回后本次合约组合不会进入发布链路，建议把原因写清楚，方便后续复盘。"
        />
        <el-alert
          v-else
          type="warning"
          :closable="false"
          show-icon
          :title="reviewDialogError || '默认发布已被策略引擎拦截，请确认后再决定是否改为强制发布。'"
        />
        <el-alert
          v-if="reviewDialogError && reviewDialogMode !== 'blocked'"
          type="error"
          :closable="false"
          show-icon
          :title="reviewDialogError"
        />

        <div v-if="reviewWarningMessages.length" class="review-warning-list">
          <div class="review-warning-list__title">本次运行提醒</div>
          <div class="tag-wrap">
            <el-tag
              v-for="warning in reviewWarningMessages"
              :key="warning"
              type="warning"
              effect="light"
            >
              {{ warning }}
            </el-tag>
          </div>
        </div>

        <el-form
          v-if="reviewDialogMode !== 'blocked'"
          label-position="top"
          class="review-form"
        >
          <el-form-item :label="reviewDialogMode === 'reject' ? '驳回原因' : '审核说明'">
            <el-input
              v-model="reviewDialogForm.reviewNote"
              type="textarea"
              :rows="4"
              resize="none"
              :placeholder="reviewDialogMode === 'reject' ? '请输入驳回原因' : '请输入审核说明，便于发布审计与后续复盘'"
            />
          </el-form-item>
          <el-form-item v-if="reviewDialogMode === 'force'" label="人工覆盖原因">
            <el-input
              v-model="reviewDialogForm.overrideReason"
              type="textarea"
              :rows="3"
              resize="none"
              placeholder="请说明为什么要强制发布，以及人工判断依据"
            />
          </el-form-item>
        </el-form>
      </div>

      <template #footer>
        <div class="review-dialog-footer">
          <el-button :disabled="actionLoading" @click="closeReviewDialog">取消</el-button>
          <el-button
            :type="reviewDialogPrimaryType"
            :loading="actionLoading"
            @click="submitReviewDialog"
          >
            {{ reviewDialogPrimaryText }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </FuturesSelectionModuleShell>
</template>

<style scoped>
.review-toolbar {
  margin-bottom: 0;
  flex-wrap: wrap;
}

.review-layout {
  display: grid;
  grid-template-columns: minmax(360px, 460px) minmax(0, 1fr);
  gap: 16px;
  align-items: start;
}

.detail-stack {
  display: grid;
  gap: 16px;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.queue-card {
  min-height: 700px;
}

.tag-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
}

.tone-card {
  margin-top: 12px;
}

.review-dialog-body {
  display: grid;
  gap: 14px;
}

.review-summary {
  padding: 14px 16px;
  border-radius: 14px;
  border: 1px solid rgba(148, 163, 184, 0.22);
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.94), rgba(241, 245, 249, 0.98));
}

.review-summary--primary {
  border-color: rgba(59, 130, 246, 0.22);
  background: linear-gradient(135deg, rgba(239, 246, 255, 0.96), rgba(224, 242, 254, 0.9));
}

.review-summary--warning {
  border-color: rgba(245, 158, 11, 0.26);
  background: linear-gradient(135deg, rgba(255, 251, 235, 0.96), rgba(254, 243, 199, 0.92));
}

.review-summary--danger {
  border-color: rgba(239, 68, 68, 0.2);
  background: linear-gradient(135deg, rgba(254, 242, 242, 0.96), rgba(254, 226, 226, 0.92));
}

.review-summary__title {
  font-size: 16px;
  font-weight: 700;
  color: #0f172a;
}

.review-summary__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 14px;
  margin-top: 8px;
  color: #475569;
  font-size: 13px;
}

.review-warning-list {
  display: grid;
  gap: 8px;
}

.review-warning-list__title {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}

.review-form {
  margin-top: 2px;
}

.review-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 1280px) {
  .review-layout,
  .detail-grid {
    grid-template-columns: 1fr;
  }

  .queue-card {
    min-height: auto;
  }
}
</style>
