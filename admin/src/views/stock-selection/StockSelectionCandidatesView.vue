<script setup>
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  approveStockSelectionReview,
  getStockSelectionRun,
  listStockSelectionCandidates,
  listStockSelectionRunEvaluation,
  listStockSelectionRunEvidence,
  listStockSelectionPortfolio,
  listStockSelectionRuns,
  rejectStockSelectionReview
} from "../../api/admin";
import {
  formatStockSelectionDateTime,
  formatStockSelectionDiffStatus,
  formatStockSelectionEvaluationScope,
  formatStockSelectionEvaluationStatus,
  formatStockSelectionLabel,
  formatStockSelectionMarketRegime,
  formatStockSelectionPercent,
  formatStockSelectionReviewStatus,
  formatStockSelectionRiskLevel,
  formatStockSelectionSource,
  formatStockSelectionStage,
  summarizeStockSelectionDiff
} from "../../lib/stock-selection";
import {
  extractReviewConflictReason,
  resolveReviewDialogMeta
} from "../../lib/review-action-dialog";
import { hasPermission } from "../../lib/session";

const route = useRoute();
const router = useRouter();
const canManage = hasPermission("stock_selection.manage");
const loading = ref(false);
const actionLoading = ref(false);
const runOptions = ref([]);
const selectedRunID = ref("");
const runDetail = ref(null);
const candidateSnapshots = ref([]);
const portfolioEntries = ref([]);
const evidenceRecords = ref([]);
const evaluationRecords = ref([]);
const selectedCandidate = ref(null);
const reviewDialogVisible = ref(false);
const reviewDialogMode = ref("approve");
const reviewDialogError = ref("");
const reviewDialogForm = reactive({
  reviewNote: "",
  overrideReason: ""
});

const candidatePool = computed(() =>
  candidateSnapshots.value.filter((item) => item.stage === "CANDIDATE_POOL")
);

const selectedEvidence = computed(() =>
  evidenceRecords.value.filter((item) => item.symbol === selectedCandidate.value?.symbol)
);

const selectedEvaluations = computed(() =>
  evaluationRecords.value.filter((item) => item.symbol === selectedCandidate.value?.symbol)
);

const watchlistCandidates = computed(() =>
  candidatePool.value.filter((item) => String(item.portfolio_role || "").toUpperCase() === "WATCHLIST")
);
const reviewWarningMessages = computed(() =>
  Array.isArray(runDetail.value?.warning_messages) ? runDetail.value.warning_messages : []
);
const reviewDialogTitle = computed(() => {
  return resolveReviewDialogMeta(reviewDialogMode.value).title;
});
const reviewDialogPrimaryText = computed(() => {
  return resolveReviewDialogMeta(reviewDialogMode.value).primaryText;
});
const reviewDialogPrimaryType = computed(() => {
  return resolveReviewDialogMeta(reviewDialogMode.value).primaryType;
});
const reviewDialogSummaryTone = computed(() => {
  return resolveReviewDialogMeta(reviewDialogMode.value).summaryTone;
});

function formatDateTime(value) {
  return formatStockSelectionDateTime(value);
}

function tagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "APPROVED") return "success";
  if (normalized === "REJECTED") return "danger";
  if (normalized === "PENDING") return "warning";
  return "info";
}

async function fetchRunOptions() {
  const data = await listStockSelectionRuns({
    status: "SUCCEEDED",
    page: 1,
    page_size: 50
  });
  runOptions.value = Array.isArray(data?.items) ? data.items : [];
  const queryRunID = String(route.query.run_id || "").trim();
  if (queryRunID && runOptions.value.some((item) => item.run_id === queryRunID)) {
    selectedRunID.value = queryRunID;
    return;
  }
  if (!selectedRunID.value && runOptions.value.length > 0) {
    selectedRunID.value = runOptions.value[0].run_id;
  }
}

async function fetchRunArtifacts() {
  if (!selectedRunID.value) {
    runDetail.value = null;
    candidateSnapshots.value = [];
    portfolioEntries.value = [];
    selectedCandidate.value = null;
    return;
  }
  loading.value = true;
  try {
    const [run, candidates, portfolio, evidence, evaluation] = await Promise.all([
      getStockSelectionRun(selectedRunID.value),
      listStockSelectionCandidates(selectedRunID.value),
      listStockSelectionPortfolio(selectedRunID.value),
      listStockSelectionRunEvidence(selectedRunID.value),
      listStockSelectionRunEvaluation(selectedRunID.value)
    ]);
    runDetail.value = run;
    candidateSnapshots.value = Array.isArray(candidates?.items) ? candidates.items : [];
    portfolioEntries.value = Array.isArray(portfolio?.items) ? portfolio.items : [];
    evidenceRecords.value = Array.isArray(evidence) ? evidence : Array.isArray(evidence?.items) ? evidence.items : [];
    evaluationRecords.value = Array.isArray(evaluation) ? evaluation : Array.isArray(evaluation?.items) ? evaluation.items : [];
    selectedCandidate.value = candidatePool.value[0] || portfolioEntries.value[0] || null;
  } catch (error) {
    ElMessage.error(error?.message || "加载候选与组合失败");
  } finally {
    loading.value = false;
  }
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
    reviewDialogForm.reviewNote = "候选池审核通过，允许发布";
    reviewDialogForm.overrideReason = "";
  } else if (mode === "force") {
    reviewDialogForm.reviewNote = "人工确认允许覆盖发布";
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
  reviewDialogError.value = "";
  try {
    if (reviewDialogMode.value === "reject") {
      await rejectStockSelectionReview(selectedRunID.value, {
        review_note: reviewNote
      });
      ElMessage.success("审核已驳回");
    } else {
      const force = reviewDialogMode.value === "force";
      await approveStockSelectionReview(selectedRunID.value, {
        review_note: reviewNote,
        force,
        override_reason: force ? overrideReason : ""
      });
      ElMessage.success(force ? "已执行强制发布" : "审核已通过并发布");
    }
    reviewDialogVisible.value = false;
    resetReviewDialog();
    await fetchRunArtifacts();
    await fetchRunOptions();
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

watch(selectedRunID, fetchRunArtifacts);

watch(
  () => route.query.run_id,
  async (runID) => {
    const normalized = String(runID || "").trim();
    if (!normalized || normalized === selectedRunID.value) {
      return;
    }
    selectedRunID.value = normalized;
  }
);

onMounted(async () => {
  await fetchRunOptions();
  await fetchRunArtifacts();
});
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股候选与审核发布"
    description="这里统一承接候选池、最终组合、证据面板和审核发布动作，不再拆成独立审核页。详细 HTML 或 Markdown 报告继续复用通用发布记录。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-select
          v-model="selectedRunID"
          placeholder="选择成功运行"
          style="width: 320px"
          filterable
        >
          <el-option
            v-for="run in runOptions"
            :key="run.run_id"
            :label="`${run.run_id} / ${run.trade_date} / ${run.profile_id}`"
            :value="run.run_id"
          />
        </el-select>
        <el-button plain @click="router.push({ name: 'stock-selection-runs' })">返回运行中心</el-button>
        <el-button :loading="loading" @click="fetchRunArtifacts">刷新当前运行</el-button>
      </div>
    </template>

    <div class="card" v-if="runDetail">
      <div class="toolbar" style="justify-content: space-between; flex-wrap: wrap">
        <div style="display: flex; gap: 8px; flex-wrap: wrap">
          <el-tag type="info">运行 {{ runDetail.run_id }}</el-tag>
          <el-tag :type="tagType(runDetail.review_status)">
            {{ formatStockSelectionReviewStatus(runDetail.review_status || "PENDING") }}
          </el-tag>
          <el-tag type="warning">{{ formatStockSelectionMarketRegime(runDetail.market_regime) }}</el-tag>
          <el-tag type="success">候选 {{ runDetail.candidate_count || 0 }}</el-tag>
          <el-tag type="warning">组合 {{ runDetail.selected_count || 0 }}</el-tag>
          <el-tag type="info">最近发布 v{{ runDetail.latest_publish_version || 0 }}</el-tag>
        </div>
        <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
          <el-button
            v-if="canManage"
            type="success"
            :loading="actionLoading"
            @click="openReviewDialog('approve')"
          >
            审核通过并发布
          </el-button>
          <el-button
            v-if="canManage"
            type="warning"
            :loading="actionLoading"
            @click="openReviewDialog('force')"
          >
            强制发布
          </el-button>
          <el-button
            v-if="canManage"
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
          {{ runDetail.result_summary || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="最近发布时间">
          {{ formatDateTime(runDetail.latest_publish_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="审核人">
          {{ runDetail.review?.reviewer || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="模板 / 市场状态" :span="2">
          {{ runDetail.template_name || "未指定模板" }} / {{ formatStockSelectionMarketRegime(runDetail.market_regime) }}
        </el-descriptions-item>
        <el-descriptions-item label="真实交易日">
          {{ runDetail.context_meta?.selected_trade_date || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="行情来源">
          {{ formatStockSelectionSource(runDetail.context_meta?.price_source) }}
        </el-descriptions-item>
      </el-descriptions>

      <div class="detail-grid" style="margin-top: 12px">
        <el-card shadow="never">
          <template #header>版本差异</template>
          <div class="mini-list">
            新增：{{ (runDetail.compare_summary?.added_symbols || []).join("、") || "无" }}
          </div>
          <div class="mini-list">
            移除：{{ (runDetail.compare_summary?.removed_symbols || []).join("、") || "无" }}
          </div>
          <div class="mini-list">
            当前组合：{{ (runDetail.compare_summary?.current_symbols || []).join("、") || "-" }}
          </div>
        </el-card>

        <el-card shadow="never">
          <template #header>运行提醒</template>
          <div v-if="(runDetail.warning_messages || []).length" class="tag-wrap">
            <el-tag
              v-for="warning in runDetail.warning_messages || []"
              :key="warning"
              type="warning"
            >
              {{ warning }}
            </el-tag>
          </div>
          <div v-else class="muted">本次运行没有额外提醒。</div>
        </el-card>
      </div>
    </div>

    <div class="candidate-layout">
      <div class="card" v-loading="loading">
        <div class="card-title">候选池</div>
        <el-table
          :data="candidatePool"
          border
          stripe
          size="small"
          height="420"
          empty-text="暂无候选池"
          @row-click="(row) => selectedCandidate = row"
        >
          <el-table-column prop="rank" label="排名" min-width="70" />
          <el-table-column prop="symbol" label="代码" min-width="120" />
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column prop="quant_score" label="量化分" min-width="90" />
          <el-table-column label="角色" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionLabel(row.portfolio_role || "PORTFOLIO") }}</template>
          </el-table-column>
          <el-table-column prop="risk_level" label="风险" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionRiskLevel(row.risk_level) }}</template>
          </el-table-column>
          <el-table-column prop="evaluation_status" label="评估" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionEvaluationStatus(row.evaluation_status) }}</template>
          </el-table-column>
          <el-table-column prop="selected" label="入组合" min-width="80">
            <template #default="{ row }">
              <el-tag :type="row.selected ? 'success' : 'info'">
                {{ row.selected ? "是" : "否" }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">最终组合</div>
        <el-table
          :data="portfolioEntries"
          border
          stripe
          size="small"
          height="420"
          empty-text="暂无最终组合"
          @row-click="(row) => selectedCandidate = row"
        >
          <el-table-column prop="rank" label="排名" min-width="70" />
          <el-table-column prop="symbol" label="代码" min-width="120" />
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column prop="quant_score" label="量化分" min-width="90" />
          <el-table-column label="角色" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionLabel(row.portfolio_role || "WATCHLIST") }}</template>
          </el-table-column>
          <el-table-column prop="weight_suggestion" label="仓位建议" min-width="120" />
          <el-table-column prop="risk_level" label="风险" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionRiskLevel(row.risk_level) }}</template>
          </el-table-column>
          <el-table-column prop="evaluation_status" label="评估" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionEvaluationStatus(row.evaluation_status) }}</template>
          </el-table-column>
        </el-table>
      </div>

      <div class="card" v-loading="loading">
        <div class="card-title">观察名单</div>
        <el-table
          :data="watchlistCandidates"
          border
          stripe
          size="small"
          height="420"
          empty-text="暂无观察名单"
          @row-click="(row) => selectedCandidate = row"
        >
          <el-table-column prop="rank" label="排名" min-width="70" />
          <el-table-column prop="symbol" label="代码" min-width="120" />
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column prop="quant_score" label="量化分" min-width="90" />
          <el-table-column prop="risk_level" label="风险" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionRiskLevel(row.risk_level) }}</template>
          </el-table-column>
          <el-table-column prop="evaluation_status" label="评估" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionEvaluationStatus(row.evaluation_status) }}</template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <div class="card" v-if="selectedCandidate">
      <div class="card-title">因子拆解</div>
      <div class="tag-wrap" style="margin-bottom: 12px">
        <el-tag type="info">{{ selectedCandidate.symbol }}</el-tag>
        <el-tag type="success">{{ formatStockSelectionLabel(selectedCandidate.portfolio_role || "PORTFOLIO") }}</el-tag>
        <el-tag type="warning">{{ formatStockSelectionRiskLevel(selectedCandidate.risk_level) }}</el-tag>
        <el-tag type="primary">{{ formatStockSelectionDiffStatus(selectedCandidate.previous_publish_diff) }}</el-tag>
        <el-tag v-if="selectedCandidate.evaluation_status" type="info">
          评估：{{ formatStockSelectionEvaluationStatus(selectedCandidate.evaluation_status) }}
        </el-tag>
      </div>
      <el-descriptions :column="3" border size="small">
        <el-descriptions-item label="标的">
          {{ selectedCandidate.symbol }} / {{ selectedCandidate.name }}
        </el-descriptions-item>
        <el-descriptions-item label="阶段">
          {{ formatStockSelectionStage(selectedCandidate.stage || "PORTFOLIO") }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ formatDateTime(runDetail?.updated_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="趋势">
          {{ selectedCandidate.factor_breakdown_json?.trend ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="资金流">
          {{ selectedCandidate.factor_breakdown_json?.money_flow ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="质量">
          {{ selectedCandidate.factor_breakdown_json?.quality ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="事件">
          {{ selectedCandidate.factor_breakdown_json?.event ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="共振">
          {{ selectedCandidate.factor_breakdown_json?.resonance ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="总分">
          {{ selectedCandidate.factor_breakdown_json?.total_score ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="风险修正">
          {{ selectedCandidate.factor_breakdown_json?.risk_adjustment ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="摘要" :span="3">
          {{ selectedCandidate.reason_summary || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="证据摘要" :span="3">
          {{ selectedCandidate.evidence_summary || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="风险边界" :span="3">
          {{ selectedCandidate.risk_summary || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="与上版相比" :span="3">
          {{ summarizeStockSelectionDiff(selectedCandidate.previous_publish_diff) }}
        </el-descriptions-item>
      </el-descriptions>

      <div class="detail-grid">
        <div>
          <div class="sub-title">证据卡片</div>
          <el-card
            v-for="item in selectedEvidence"
            :key="`${item.stage}-${item.symbol}`"
            class="mini-card"
            shadow="never"
          >
            <div class="mini-card-title">
              {{ formatStockSelectionStage(item.stage) }} / {{ formatStockSelectionLabel(item.portfolio_role || "PORTFOLIO") }}
            </div>
            <div class="muted" style="margin-bottom: 8px">{{ item.evidence_summary || "-" }}</div>
            <el-tag
              v-for="card in item.evidence_cards_json || []"
              :key="`${card.title}-${card.value}`"
              style="margin-right: 8px; margin-bottom: 8px"
            >
              {{ card.title }}：{{ card.value }}
            </el-tag>
            <div class="mini-list">入选原因：{{ (item.positive_reasons_json || []).join("；") || "-" }}</div>
            <div class="mini-list">淘汰/风险：{{ (item.veto_reasons_json || []).join("；") || "-" }}</div>
            <div class="mini-list">题材标签：{{ (item.theme_tags_json || []).join("、") || "-" }}</div>
            <div class="mini-list">行业标签：{{ (item.sector_tags_json || []).join("、") || "-" }}</div>
            <div class="mini-list">风险旗标：{{ (item.risk_flags_json || []).join("、") || "-" }}</div>
          </el-card>
        </div>

        <div>
          <div class="sub-title">版本差异与评估</div>
          <el-card class="mini-card" shadow="never">
            <div class="mini-list">差异状态：{{ formatStockSelectionDiffStatus(selectedCandidate.previous_publish_diff) }}</div>
            <div class="mini-list">差异说明：{{ summarizeStockSelectionDiff(selectedCandidate.previous_publish_diff) }}</div>
            <div class="mini-list">状态：{{ formatStockSelectionEvaluationStatus(selectedCandidate.evaluation_status || "PENDING") }}</div>
          </el-card>
          <el-table :data="selectedEvaluations" border stripe size="small" empty-text="评估尚未生成">
            <el-table-column prop="horizon_day" label="窗口" min-width="80">
              <template #default="{ row }">{{ row.horizon_day }} 日</template>
            </el-table-column>
            <el-table-column prop="evaluation_scope" label="范围" min-width="100">
              <template #default="{ row }">{{ formatStockSelectionEvaluationScope(row.evaluation_scope) }}</template>
            </el-table-column>
            <el-table-column prop="entry_date" label="入场日" min-width="110" />
            <el-table-column prop="exit_date" label="出场日" min-width="110" />
            <el-table-column prop="return_pct" label="收益" min-width="90">
              <template #default="{ row }">{{ formatStockSelectionPercent(row.return_pct) }}</template>
            </el-table-column>
            <el-table-column prop="excess_return_pct" label="超额" min-width="90">
              <template #default="{ row }">{{ formatStockSelectionPercent(row.excess_return_pct) }}</template>
            </el-table-column>
            <el-table-column prop="max_drawdown_pct" label="回撤" min-width="90">
              <template #default="{ row }">{{ formatStockSelectionPercent(row.max_drawdown_pct) }}</template>
            </el-table-column>
            <el-table-column prop="benchmark_symbol" label="基准" min-width="100" />
            <el-table-column prop="hit_flag" label="命中" min-width="80">
              <template #default="{ row }">{{ row.hit_flag ? "是" : "否" }}</template>
            </el-table-column>
          </el-table>
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
            {{ runDetail?.template_name || "未指定模板" }} / {{ formatStockSelectionMarketRegime(runDetail?.market_regime) }}
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
          title="将按默认发布策略检查风险阈值；若警告数量超限，系统会拦截本次发布。"
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
          title="驳回后本次组合不会进入发布链路，建议把原因写清楚，方便后续复盘。"
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
  </StockSelectionModuleShell>
</template>

<style scoped>
.candidate-layout {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 12px;
  margin-bottom: 12px;
}

.card-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}

.detail-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 12px;
  margin-top: 12px;
}

.sub-title {
  margin-bottom: 10px;
  font-weight: 600;
}

.mini-card {
  margin-bottom: 12px;
}

.mini-card-title {
  margin-bottom: 8px;
  font-weight: 600;
}

.mini-list {
  margin-top: 8px;
  color: var(--el-text-color-regular);
  font-size: 13px;
}

.tag-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
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
</style>
