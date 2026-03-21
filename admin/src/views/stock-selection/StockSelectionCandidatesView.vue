<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
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
  formatStockSelectionMarketRegime,
  formatStockSelectionPercent,
  formatStockSelectionReviewStatus,
  formatStockSelectionRiskLevel,
  formatStockSelectionStage,
  summarizeStockSelectionDiff
} from "../../lib/stock-selection";
import { hasPermission } from "../../lib/session";

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

async function handleApprove(force = false) {
  if (!selectedRunID.value) {
    return;
  }
  actionLoading.value = true;
  try {
    const payload = {
      review_note: force ? "人工确认允许覆盖发布" : "候选池审核通过，允许发布",
      force,
      override_reason: ""
    };
    if (force) {
      const result = await ElMessageBox.prompt("请输入人工覆盖原因", "强制发布", {
        confirmButtonText: "发布",
        cancelButtonText: "取消"
      });
      payload.override_reason = result.value;
      payload.review_note = `人工覆盖发布：${result.value}`;
    }
    await approveStockSelectionReview(selectedRunID.value, payload);
    ElMessage.success(force ? "已执行强制发布" : "审核已通过并发布");
    await fetchRunArtifacts();
    await fetchRunOptions();
  } catch (error) {
    if (error === "cancel") {
      return;
    }
    ElMessage.error(error?.message || "审核通过失败");
  } finally {
    actionLoading.value = false;
  }
}

async function handleReject() {
  if (!selectedRunID.value) {
    return;
  }
  actionLoading.value = true;
  try {
    const result = await ElMessageBox.prompt("请输入驳回原因", "驳回审核", {
      confirmButtonText: "驳回",
      cancelButtonText: "取消"
    });
    await rejectStockSelectionReview(selectedRunID.value, {
      review_note: result.value
    });
    ElMessage.success("审核已驳回");
    await fetchRunArtifacts();
    await fetchRunOptions();
  } catch (error) {
    if (error === "cancel") {
      return;
    }
    ElMessage.error(error?.message || "驳回失败");
  } finally {
    actionLoading.value = false;
  }
}

watch(selectedRunID, fetchRunArtifacts);

onMounted(async () => {
  await fetchRunOptions();
  await fetchRunArtifacts();
});
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股候选与组合"
    description="这里承接最近成功运行的候选池、最终组合和审核发布动作，详细 HTML 或 Markdown 报告继续复用通用发布记录。"
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
        <el-button :loading="loading" @click="fetchRunArtifacts">刷新当前 Run</el-button>
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
            @click="handleApprove(false)"
          >
            审核通过并发布
          </el-button>
          <el-button
            v-if="canManage"
            type="warning"
            :loading="actionLoading"
            @click="handleApprove(true)"
          >
            强制发布
          </el-button>
          <el-button
            v-if="canManage"
            type="danger"
            plain
            :loading="actionLoading"
            @click="handleReject"
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
          {{ runDetail.context_meta?.price_source || "-" }}
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
          <el-table-column prop="portfolio_role" label="角色" min-width="90" />
          <el-table-column prop="risk_level" label="风险" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionRiskLevel(row.risk_level) }}</template>
          </el-table-column>
          <el-table-column prop="evaluation_status" label="评估" min-width="90" />
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
          <el-table-column prop="portfolio_role" label="角色" min-width="90" />
          <el-table-column prop="weight_suggestion" label="仓位建议" min-width="120" />
          <el-table-column prop="risk_level" label="风险" min-width="90">
            <template #default="{ row }">{{ formatStockSelectionRiskLevel(row.risk_level) }}</template>
          </el-table-column>
          <el-table-column prop="evaluation_status" label="评估" min-width="90" />
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
          <el-table-column prop="evaluation_status" label="评估" min-width="90" />
        </el-table>
      </div>
    </div>

    <div class="card" v-if="selectedCandidate">
      <div class="card-title">因子拆解</div>
      <div class="tag-wrap" style="margin-bottom: 12px">
        <el-tag type="info">{{ selectedCandidate.symbol }}</el-tag>
        <el-tag type="success">{{ selectedCandidate.portfolio_role || "候选" }}</el-tag>
        <el-tag type="warning">{{ formatStockSelectionRiskLevel(selectedCandidate.risk_level) }}</el-tag>
        <el-tag type="primary">{{ formatStockSelectionDiffStatus(selectedCandidate.previous_publish_diff) }}</el-tag>
        <el-tag v-if="selectedCandidate.evaluation_status" type="info">
          评估：{{ selectedCandidate.evaluation_status }}
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
              {{ formatStockSelectionStage(item.stage) }} / {{ item.portfolio_role || "候选" }}
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
            <div class="mini-list">状态：{{ selectedCandidate.evaluation_status || "PENDING" }}</div>
          </el-card>
          <el-table :data="selectedEvaluations" border stripe size="small" empty-text="评估尚未生成">
            <el-table-column prop="horizon_day" label="窗口" min-width="80">
              <template #default="{ row }">{{ row.horizon_day }} 日</template>
            </el-table-column>
            <el-table-column prop="evaluation_scope" label="范围" min-width="100" />
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
</style>
