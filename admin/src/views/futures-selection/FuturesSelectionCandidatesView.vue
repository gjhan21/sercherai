<script setup>
import { computed, onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import FuturesSelectionModuleShell from "../../components/FuturesSelectionModuleShell.vue";
import {
  approveFuturesSelectionReview,
  getFuturesSelectionRun,
  listFuturesSelectionCandidates,
  listFuturesSelectionPortfolio,
  listFuturesSelectionRunEvaluation,
  listFuturesSelectionRunEvidence,
  listFuturesSelectionRuns,
  rejectFuturesSelectionReview
} from "../../api/admin";
import {
  formatFuturesSelectionDateTime,
  formatFuturesSelectionDiffStatus,
  formatFuturesSelectionDirection,
  formatFuturesSelectionEvaluationScope,
  formatFuturesSelectionEvaluationStatus,
  formatFuturesSelectionLabel,
  formatFuturesSelectionMarketRegime,
  formatFuturesSelectionPercent,
  formatFuturesSelectionReviewStatus,
  formatFuturesSelectionRiskLevel,
  formatFuturesSelectionStage,
  summarizeFuturesSelectionDiff
} from "../../lib/futures-selection";
import {
  extractReviewConflictReason,
  resolveReviewDialogMeta
} from "../../lib/review-action-dialog";
import { hasPermission } from "../../lib/session";

const route = useRoute();
const router = useRouter();
const canManage = hasPermission("futures_selection.manage");
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
  evidenceRecords.value.filter((item) => item.contract === selectedCandidate.value?.contract)
);

const selectedEvaluations = computed(() =>
  evaluationRecords.value.filter((item) => item.contract === selectedCandidate.value?.contract)
);
const selectedSupplyChainCards = computed(() => {
  const titles = new Set(["库存画像", "结构联动", "商品链"]);
  return selectedEvidence.value
    .flatMap((item) => (Array.isArray(item.evidence_cards_json) ? item.evidence_cards_json : []))
    .filter((card) => titles.has(String(card?.title || "").trim()));
});
const selectedSupplyChainEntities = computed(() => {
  const allowed = new Set(["Commodity", "SupplyChainNode", "SpreadPair", "Index", "DeliveryPlace", "Warehouse", "Brand", "Grade"]);
  const seen = new Set();
  return selectedEvidence.value
    .flatMap((item) => (Array.isArray(item.related_entities_json) ? item.related_entities_json : []))
    .map((item) => ({
      entityType: String(item?.entity_type || "").trim(),
      label: String(item?.label || item?.entity_key || "").trim()
    }))
    .filter((item) => item.entityType && item.label && allowed.has(item.entityType))
    .filter((item) => {
      const key = `${item.entityType}:${item.label}`;
      if (seen.has(key)) return false;
      seen.add(key);
      return true;
    });
});
const selectedInventorySummary = computed(() => {
  const matched = selectedSupplyChainCards.value.find((item) => String(item?.title || "").trim() === "库存画像");
  if (matched?.note) {
    return matched.note;
  }
  const depth = selectedCandidate.value?.factor_breakdown_json?.inventory_depth;
  return depth == null ? "-" : `库存深度 ${depth}`;
});
const selectedStructureSummary = computed(() => {
  const matched = selectedSupplyChainCards.value.find((item) => String(item?.title || "").trim() === "结构联动");
  if (matched?.note) {
    return matched.note;
  }
  const basisTerm = selectedCandidate.value?.factor_breakdown_json?.basis_term;
  const depth = selectedCandidate.value?.factor_breakdown_json?.structure_depth;
  if (basisTerm != null && depth != null) {
    return `结构深度 ${depth} / 基差结构 ${basisTerm}`;
  }
  if (depth != null) {
    return `结构深度 ${depth}`;
  }
  return "-";
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
  return "info";
}

async function fetchRunOptions() {
  const data = await listFuturesSelectionRuns({
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
    evidenceRecords.value = [];
    evaluationRecords.value = [];
    selectedCandidate.value = null;
    return;
  }
  loading.value = true;
  try {
    const [run, candidates, portfolio, evidence, evaluation] = await Promise.all([
      getFuturesSelectionRun(selectedRunID.value),
      listFuturesSelectionCandidates(selectedRunID.value),
      listFuturesSelectionPortfolio(selectedRunID.value),
      listFuturesSelectionRunEvidence(selectedRunID.value),
      listFuturesSelectionRunEvaluation(selectedRunID.value)
    ]);
    runDetail.value = run;
    candidateSnapshots.value = Array.isArray(candidates?.items) ? candidates.items : [];
    portfolioEntries.value = Array.isArray(portfolio?.items) ? portfolio.items : [];
    evidenceRecords.value = Array.isArray(evidence?.items) ? evidence.items : [];
    evaluationRecords.value = Array.isArray(evaluation?.items) ? evaluation.items : [];
    selectedCandidate.value = portfolioEntries.value[0] || candidatePool.value[0] || null;
  } catch (error) {
    ElMessage.error(error?.message || "加载期货候选与组合失败");
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
    reviewDialogForm.reviewNote = "期货候选审核通过，允许发布";
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
  <FuturesSelectionModuleShell
    title="智能期货候选与审核发布"
    description="这里统一承接期货候选池、最终组合、证据卡片、评估状态和审核发布动作，不再拆成独立审核页。"
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
        <el-button plain @click="router.push({ name: 'futures-selection-runs' })">返回运行中心</el-button>
        <el-button :loading="loading" @click="fetchRunArtifacts">刷新当前运行</el-button>
      </div>
    </template>

    <div class="card" v-if="runDetail">
      <div class="toolbar" style="justify-content: space-between; flex-wrap: wrap">
        <div style="display: flex; gap: 8px; flex-wrap: wrap">
          <el-tag type="info">运行 {{ runDetail.run_id }}</el-tag>
          <el-tag :type="tagType(runDetail.review_status)">
            {{ formatFuturesSelectionReviewStatus(runDetail.review_status || "PENDING") }}
          </el-tag>
          <el-tag type="warning">{{ formatFuturesSelectionMarketRegime(runDetail.market_regime) }}</el-tag>
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
          {{ formatFuturesSelectionDateTime(runDetail.latest_publish_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="审核人">
          {{ runDetail.review?.reviewer || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="模板 / 市场状态" :span="2">
          {{ runDetail.template_name || "未指定模板" }} / {{ formatFuturesSelectionMarketRegime(runDetail.market_regime) }}
        </el-descriptions-item>
        <el-descriptions-item label="真实交易日">
          {{ runDetail.context_meta?.selected_trade_date || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="图快照">
          {{ runDetail.context_meta?.graph_snapshot_id || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="版本差异" :span="2">
          新增：{{ (runDetail.compare_summary?.added_contracts || []).join("、") || "无" }}
          / 移除：{{ (runDetail.compare_summary?.removed_contracts || []).join("、") || "无" }}
        </el-descriptions-item>
      </el-descriptions>
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
          <el-table-column prop="contract" label="合约" min-width="120" />
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column prop="direction" label="方向" min-width="90">
            <template #default="{ row }">{{ formatFuturesSelectionDirection(row.direction) }}</template>
          </el-table-column>
          <el-table-column prop="score" label="评分" min-width="90" />
          <el-table-column label="角色" min-width="90">
            <template #default="{ row }">{{ formatFuturesSelectionLabel(row.portfolio_role || "SATELLITE") }}</template>
          </el-table-column>
          <el-table-column prop="risk_level" label="风险" min-width="90">
            <template #default="{ row }">{{ formatFuturesSelectionRiskLevel(row.risk_level) }}</template>
          </el-table-column>
          <el-table-column label="版本差异" min-width="110">
            <template #default="{ row }">
              {{ formatFuturesSelectionDiffStatus(row.previous_publish_diff) }}
            </template>
          </el-table-column>
          <el-table-column prop="evaluation_status" label="评估" min-width="90">
            <template #default="{ row }">{{ formatFuturesSelectionEvaluationStatus(row.evaluation_status) }}</template>
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
          <el-table-column prop="contract" label="合约" min-width="120" />
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column prop="direction" label="方向" min-width="90">
            <template #default="{ row }">{{ formatFuturesSelectionDirection(row.direction) }}</template>
          </el-table-column>
          <el-table-column prop="score" label="评分" min-width="90" />
          <el-table-column prop="position_range" label="仓位建议" min-width="120" />
          <el-table-column prop="risk_level" label="风险" min-width="90">
            <template #default="{ row }">{{ formatFuturesSelectionRiskLevel(row.risk_level) }}</template>
          </el-table-column>
          <el-table-column label="版本差异" min-width="110">
            <template #default="{ row }">
              {{ formatFuturesSelectionDiffStatus(row.previous_publish_diff) }}
            </template>
          </el-table-column>
          <el-table-column prop="evaluation_status" label="评估" min-width="90">
            <template #default="{ row }">{{ formatFuturesSelectionEvaluationStatus(row.evaluation_status) }}</template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <div class="card" v-if="selectedCandidate">
      <div class="card-title">合约因子拆解</div>
      <div class="tag-wrap" style="margin-bottom: 12px">
        <el-tag type="info">{{ selectedCandidate.contract }}</el-tag>
        <el-tag type="success">{{ formatFuturesSelectionLabel(selectedCandidate.portfolio_role || "SATELLITE") }}</el-tag>
        <el-tag type="warning">{{ formatFuturesSelectionRiskLevel(selectedCandidate.risk_level) }}</el-tag>
        <el-tag type="primary">{{ formatFuturesSelectionDirection(selectedCandidate.direction) }}</el-tag>
        <el-tag type="primary">{{ formatFuturesSelectionDiffStatus(selectedCandidate.previous_publish_diff) }}</el-tag>
        <el-tag v-if="selectedCandidate.evaluation_status" type="info">
          评估：{{ formatFuturesSelectionEvaluationStatus(selectedCandidate.evaluation_status) }}
        </el-tag>
      </div>

      <el-descriptions :column="3" border size="small">
        <el-descriptions-item label="合约">
          {{ selectedCandidate.contract }} / {{ selectedCandidate.name }}
        </el-descriptions-item>
        <el-descriptions-item label="阶段">
          {{ formatFuturesSelectionStage(selectedCandidate.stage || "PORTFOLIO") }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ formatFuturesSelectionDateTime(runDetail?.updated_at) }}
        </el-descriptions-item>
        <el-descriptions-item label="趋势">
          {{ selectedCandidate.factor_breakdown_json?.trend ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="库存深度">
          {{ selectedCandidate.factor_breakdown_json?.inventory_depth ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="结构深度">
          {{ selectedCandidate.factor_breakdown_json?.structure_depth ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="基差结构">
          {{ selectedCandidate.factor_breakdown_json?.basis_term ?? selectedCandidate.factor_breakdown_json?.basis ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="总分">
          {{ selectedCandidate.factor_breakdown_json?.total_score ?? selectedCandidate.factor_breakdown_json?.total ?? selectedCandidate.score ?? "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="摘要" :span="3">
          {{ selectedCandidate.reason_summary || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="证据摘要" :span="3">
          {{ selectedCandidate.evidence_summary || "-" }}
        </el-descriptions-item>
        <el-descriptions-item label="与上版相比" :span="3">
          {{ summarizeFuturesSelectionDiff(selectedCandidate.previous_publish_diff) }}
        </el-descriptions-item>
        <el-descriptions-item label="风险边界" :span="3">
          {{ selectedCandidate.risk_summary || "-" }}
        </el-descriptions-item>
      </el-descriptions>

      <div class="detail-grid">
        <div>
          <div class="sub-title">商品链证据</div>
          <el-card class="mini-card" shadow="never">
            <div class="mini-list"><strong>结构联动摘要：</strong>{{ selectedStructureSummary }}</div>
            <div class="mini-list"><strong>库存画像摘要：</strong>{{ selectedInventorySummary }}</div>
            <div class="mini-list">
              <strong>商品链实体：</strong>
              {{ selectedSupplyChainEntities.map((item) => item.label).join("、") || "-" }}
            </div>
            <div class="tag-wrap" v-if="selectedSupplyChainEntities.length" style="margin-top: 8px">
              <el-tag
                v-for="item in selectedSupplyChainEntities"
                :key="`${item.entityType}-${item.label}`"
                type="success"
                effect="light"
              >
                {{ item.entityType }} / {{ item.label }}
              </el-tag>
            </div>
          </el-card>
          <el-card
            v-for="card in selectedSupplyChainCards"
            :key="`${card.title}-${card.value}-${card.note}`"
            class="mini-card"
            shadow="never"
          >
            <div class="mini-card-title">{{ card.title || "商品链证据" }}</div>
            <div class="muted" style="margin-bottom: 8px">{{ card.value || "-" }}</div>
            <div class="mini-list">{{ card.note || "-" }}</div>
          </el-card>
        </div>

        <div>
          <div class="sub-title">证据卡片</div>
          <el-card
            v-for="item in selectedEvidence"
            :key="`${item.stage}-${item.contract}`"
            class="mini-card"
            shadow="never"
          >
            <div class="mini-card-title">
              {{ formatFuturesSelectionStage(item.stage) }} / {{ formatFuturesSelectionLabel(item.portfolio_role || "SATELLITE") }}
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
            <div class="mini-list">风险旗标：{{ (item.risk_flags_json || []).join("、") || "-" }}</div>
            <div class="mini-list">
              关联实体：
              {{ (item.related_entities_json || []).map((entity) => entity.label || entity.entity_key).filter(Boolean).join("、") || "-" }}
            </div>
          </el-card>
        </div>

        <div>
          <div class="sub-title">评估结果</div>
          <el-card class="mini-card" shadow="never">
            <div class="mini-list">差异状态：{{ formatFuturesSelectionDiffStatus(selectedCandidate.previous_publish_diff) }}</div>
            <div class="mini-list">差异说明：{{ summarizeFuturesSelectionDiff(selectedCandidate.previous_publish_diff) }}</div>
            <div class="mini-list">评估状态：{{ formatFuturesSelectionEvaluationStatus(selectedCandidate.evaluation_status || "PENDING") }}</div>
            <div class="mini-list">方向：{{ formatFuturesSelectionDirection(selectedCandidate.direction) }}</div>
            <div class="mini-list">仓位建议：{{ selectedCandidate.position_range || "-" }}</div>
          </el-card>
          <el-table :data="selectedEvaluations" border stripe size="small" empty-text="评估尚未生成">
            <el-table-column prop="horizon_day" label="窗口" min-width="80">
              <template #default="{ row }">{{ row.horizon_day }} 日</template>
            </el-table-column>
            <el-table-column prop="evaluation_scope" label="范围" min-width="100">
              <template #default="{ row }">{{ formatFuturesSelectionEvaluationScope(row.evaluation_scope) }}</template>
            </el-table-column>
            <el-table-column prop="entry_date" label="入场日" min-width="110" />
            <el-table-column prop="exit_date" label="出场日" min-width="110" />
            <el-table-column prop="return_pct" label="收益" min-width="90">
              <template #default="{ row }">{{ formatFuturesSelectionPercent(row.return_pct) }}</template>
            </el-table-column>
            <el-table-column prop="excess_return_pct" label="超额" min-width="90">
              <template #default="{ row }">{{ formatFuturesSelectionPercent(row.excess_return_pct) }}</template>
            </el-table-column>
            <el-table-column prop="max_drawdown_pct" label="回撤" min-width="90">
              <template #default="{ row }">{{ formatFuturesSelectionPercent(row.max_drawdown_pct) }}</template>
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
  margin-bottom: 8px;
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
