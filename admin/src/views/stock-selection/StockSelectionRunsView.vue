<script setup>
import { onMounted, reactive, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  compareStockSelectionRuns,
  createStockSelectionRun,
  getStockSelectionRun,
  listStockSelectionProfiles,
  listStockSelectionRuns
} from "../../api/admin";
import {
  formatStockSelectionDateTime,
  formatStockSelectionMarketRegime,
  formatStockSelectionMode,
  formatStockSelectionPercent,
  formatStockSelectionReviewStatus,
  formatStockSelectionRunStatus,
  formatStockSelectionStage,
  formatStockSelectionUniverseScope
} from "../../lib/stock-selection";
import { hasPermission } from "../../lib/session";

const route = useRoute();
const router = useRouter();
const canManage = hasPermission("stock_selection.manage");
const loading = ref(false);
const creating = ref(false);
const detailLoading = ref(false);
const errorMessage = ref("");
const page = ref(1);
const pageSize = ref(20);
const total = ref(0);
const runs = ref([]);
const profiles = ref([]);
const detailVisible = ref(false);
const selectedRun = ref(null);
const selectedCompareRunIDs = ref([]);
const compareResult = ref(null);

const filters = reactive({
  status: "",
  review_status: "",
  profile_id: ""
});

function formatDateTime(value) {
  return formatStockSelectionDateTime(value);
}

function tagType(status) {
  const normalized = String(status || "").toUpperCase();
  if (normalized === "SUCCEEDED" || normalized === "APPROVED") return "success";
  if (normalized === "FAILED" || normalized === "REJECTED") return "danger";
  if (normalized === "RUNNING" || normalized === "PENDING") return "warning";
  return "info";
}

async function fetchProfiles() {
  const data = await listStockSelectionProfiles({ page: 1, page_size: 100 });
  profiles.value = Array.isArray(data?.items) ? data.items : [];
}

async function fetchRuns() {
  loading.value = true;
  errorMessage.value = "";
  try {
    const data = await listStockSelectionRuns({
      ...filters,
      page: page.value,
      page_size: pageSize.value
    });
    runs.value = Array.isArray(data?.items) ? data.items : [];
    total.value = Number(data?.total) || 0;
  } catch (error) {
    errorMessage.value = error?.message || "加载运行中心失败";
  } finally {
    loading.value = false;
  }
}

async function openRunDetail(runID) {
  if (!runID) {
    return;
  }
  detailVisible.value = true;
  detailLoading.value = true;
  try {
    selectedRun.value = await getStockSelectionRun(runID);
  } catch (error) {
    ElMessage.error(error?.message || "加载运行详情失败");
  } finally {
    detailLoading.value = false;
  }
}

async function handleCompare() {
  if (selectedCompareRunIDs.value.length < 2 || selectedCompareRunIDs.value.length > 3) {
    ElMessage.warning("请选择 2 到 3 条运行记录做对比");
    return;
  }
  try {
    compareResult.value = await compareStockSelectionRuns(selectedCompareRunIDs.value);
  } catch (error) {
    ElMessage.error(error?.message || "加载运行对比失败");
  }
}

function handleSelectionChange(rows) {
  selectedCompareRunIDs.value = Array.isArray(rows)
    ? rows.map((item) => item.run_id).slice(0, 3)
    : [];
}

async function handleQuickRun() {
  creating.value = true;
  try {
    const run = await createStockSelectionRun({});
    ElMessage.success(`已完成运行 ${run.run_id}`);
    await fetchRuns();
    router.replace({
      name: "stock-selection-runs",
      query: { run_id: run.run_id }
    });
    await openRunDetail(run.run_id);
  } catch (error) {
    ElMessage.error(error?.message || "创建运行失败");
  } finally {
    creating.value = false;
  }
}

watch(
  () => route.query.run_id,
  async (runID) => {
    if (runID) {
      await openRunDetail(String(runID));
    }
  }
);

onMounted(async () => {
  await Promise.all([fetchProfiles(), fetchRuns()]);
  if (route.query.run_id) {
    await openRunDetail(String(route.query.run_id));
  }
});
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股运行中心"
    description="集中查看运行记录、阶段摘要、审核状态和策略引擎发布摘要；新运行仍然同步等待 Python 返回结果。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchRuns">刷新列表</el-button>
        <el-button v-if="canManage" type="primary" :loading="creating" @click="handleQuickRun">
          立即运行
        </el-button>
        <el-button :disabled="selectedCompareRunIDs.length < 2" @click="handleCompare">对比选中 Run</el-button>
      </div>
    </template>

    <div class="card">
      <div class="toolbar" style="flex-wrap: wrap">
        <el-select v-model="filters.status" clearable placeholder="运行状态" style="width: 140px">
          <el-option label="运行中" value="RUNNING" />
          <el-option label="成功" value="SUCCEEDED" />
          <el-option label="失败" value="FAILED" />
        </el-select>
        <el-select v-model="filters.review_status" clearable placeholder="审核状态" style="width: 140px">
          <el-option label="待审核" value="PENDING" />
          <el-option label="已通过" value="APPROVED" />
          <el-option label="已驳回" value="REJECTED" />
        </el-select>
        <el-select v-model="filters.profile_id" clearable placeholder="配置方案" style="width: 240px">
          <el-option
            v-for="profile in profiles"
            :key="profile.id"
            :label="`${profile.name}（v${profile.current_version || 0}）`"
            :value="profile.id"
          />
        </el-select>
        <el-button type="primary" @click="page = 1; fetchRuns()">查询</el-button>
      </div>

      <el-alert
        v-if="errorMessage"
        :title="errorMessage"
        type="error"
        show-icon
        style="margin-bottom: 12px"
      />

      <el-table
        :data="runs"
        border
        stripe
        size="small"
        v-loading="loading"
        empty-text="暂无运行记录"
        @selection-change="handleSelectionChange"
        @row-click="(row) => openRunDetail(row.run_id)"
      >
        <el-table-column type="selection" width="48" />
        <el-table-column prop="run_id" label="运行编号" min-width="180" />
        <el-table-column prop="trade_date" label="交易日" min-width="110" />
        <el-table-column prop="profile_id" label="配置方案" min-width="180" />
        <el-table-column prop="template_name" label="模板" min-width="140" />
        <el-table-column prop="market_regime" label="市场状态" min-width="120">
          <template #default="{ row }">{{ formatStockSelectionMarketRegime(row.market_regime) }}</template>
        </el-table-column>
        <el-table-column label="状态" min-width="100">
          <template #default="{ row }">
            <el-tag :type="tagType(row.status)">{{ formatStockSelectionRunStatus(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="审核" min-width="100">
          <template #default="{ row }">
            <el-tag :type="tagType(row.review_status)">{{ formatStockSelectionReviewStatus(row.review_status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="universe_count" label="股票池" min-width="90" />
        <el-table-column prop="seed_count" label="种子池" min-width="90" />
        <el-table-column prop="candidate_count" label="候选池" min-width="100" />
        <el-table-column prop="selected_count" label="最终组合" min-width="100" />
        <el-table-column prop="publish_count" label="发布次数" min-width="100" />
        <el-table-column prop="latest_publish_version" label="最近版本" min-width="100" />
        <el-table-column prop="latest_publish_at" label="最近发布" min-width="170">
          <template #default="{ row }">{{ formatDateTime(row.latest_publish_at) }}</template>
        </el-table-column>
        <el-table-column prop="job_id" label="作业编号" min-width="170" />
      </el-table>

      <div class="toolbar" style="justify-content: flex-end; margin-top: 12px">
        <el-pagination
          background
          layout="prev, pager, next, total"
          :current-page="page"
          :page-size="pageSize"
          :total="total"
          @current-change="(value) => { page = value; fetchRuns(); }"
        />
      </div>
    </div>

    <div class="card" v-if="compareResult?.items?.length">
      <div class="card-title">运行对比</div>
      <el-table :data="compareResult.items" border stripe size="small" empty-text="暂无对比数据">
        <el-table-column prop="run_id" label="运行编号" min-width="180" />
        <el-table-column prop="template_name" label="模板" min-width="140" />
        <el-table-column prop="market_regime" label="市场状态" min-width="120">
          <template #default="{ row }">{{ formatStockSelectionMarketRegime(row.market_regime) }}</template>
        </el-table-column>
        <el-table-column prop="selected_count" label="组合数量" min-width="100" />
        <el-table-column prop="added_symbols" label="新增标的" min-width="180">
          <template #default="{ row }">{{ (row.added_symbols || []).join("、") || "-" }}</template>
        </el-table-column>
        <el-table-column prop="removed_symbols" label="移除标的" min-width="180">
          <template #default="{ row }">{{ (row.removed_symbols || []).join("、") || "-" }}</template>
        </el-table-column>
        <el-table-column prop="portfolio_symbols" label="组合快照" min-width="220">
          <template #default="{ row }">{{ (row.portfolio_symbols || []).join("、") || "-" }}</template>
        </el-table-column>
      </el-table>
    </div>

    <el-drawer v-model="detailVisible" size="760px" title="运行详情">
      <div v-loading="detailLoading">
        <template v-if="selectedRun">
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="运行编号">{{ selectedRun.run_id }}</el-descriptions-item>
            <el-descriptions-item label="状态">
              <el-tag :type="tagType(selectedRun.status)">{{ formatStockSelectionRunStatus(selectedRun.status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="配置方案">
              {{ selectedRun.profile_id }} / v{{ selectedRun.profile_version }}
            </el-descriptions-item>
            <el-descriptions-item label="模板 / 市场状态">
              {{ selectedRun.template_name || "未指定模板" }} / {{ formatStockSelectionMarketRegime(selectedRun.market_regime) }}
            </el-descriptions-item>
            <el-descriptions-item label="审核">
              <el-tag :type="tagType(selectedRun.review_status)">{{ formatStockSelectionReviewStatus(selectedRun.review_status) }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="最近发布">
              {{ selectedRun.latest_publish_id || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="发布时间">
              {{ formatDateTime(selectedRun.latest_publish_at) }}
            </el-descriptions-item>
            <el-descriptions-item label="运行模式">
              {{ formatStockSelectionMode(selectedRun.selection_mode) }}
            </el-descriptions-item>
            <el-descriptions-item label="股票池范围">
              {{ formatStockSelectionUniverseScope(selectedRun.universe_scope) }}
            </el-descriptions-item>
            <el-descriptions-item label="真实交易日">
              {{ selectedRun.context_meta?.selected_trade_date || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="行情来源">
              {{ selectedRun.context_meta?.price_source || "-" }}
            </el-descriptions-item>
            <el-descriptions-item label="结果摘要" :span="2">
              {{ selectedRun.result_summary || "-" }}
            </el-descriptions-item>
          </el-descriptions>

          <div class="card" style="margin-top: 12px" v-if="selectedRun.context_meta">
            <div class="card-title">上下文与过滤条件</div>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="列表真相交易日">
                {{ selectedRun.context_meta?.selected_trade_date || "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="listing_days 代理">
                {{ selectedRun.context_meta?.listing_days_filter_applied === false ? "已跳过" : "已启用" }}
              </el-descriptions-item>
              <el-descriptions-item label="价格区间">
                {{ selectedRun.context_meta?.universe_filters?.price_min ?? "-" }} ~
                {{ selectedRun.context_meta?.universe_filters?.price_max ?? "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="波动率区间">
                {{ selectedRun.context_meta?.universe_filters?.volatility_min ?? "-" }} ~
                {{ selectedRun.context_meta?.universe_filters?.volatility_max ?? "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="最少上市天数">
                {{ selectedRun.context_meta?.universe_filters?.min_listing_days ?? "-" }}
              </el-descriptions-item>
              <el-descriptions-item label="20日均成交额下限">
                {{ selectedRun.context_meta?.universe_filters?.min_avg_turnover ?? "-" }}
              </el-descriptions-item>
            </el-descriptions>
          </div>

          <div class="card" style="margin-top: 12px" v-if="selectedRun.compare_summary">
            <div class="card-title">与最近发布版本对比</div>
            <div class="chip-wrap">
              <el-tag type="success">
                新增：{{ (selectedRun.compare_summary?.added_symbols || []).join("、") || "无" }}
              </el-tag>
              <el-tag type="danger">
                移除：{{ (selectedRun.compare_summary?.removed_symbols || []).join("、") || "无" }}
              </el-tag>
            </div>
            <div class="mini-text">
              当前组合：{{ (selectedRun.compare_summary?.current_symbols || []).join("、") || "-" }}
            </div>
          </div>

          <div class="card" style="margin-top: 12px" v-if="(selectedRun.warning_messages || []).length">
            <div class="card-title">运行提醒</div>
            <div class="chip-wrap">
              <el-tag
                v-for="warning in selectedRun.warning_messages || []"
                :key="warning"
                type="warning"
              >
                {{ warning }}
              </el-tag>
            </div>
          </div>

          <div class="card" style="margin-top: 12px">
            <div class="card-title">阶段计数</div>
            <el-tag
              v-for="(count, key) in selectedRun.stage_counts || {}"
              :key="key"
              type="info"
              style="margin-right: 8px; margin-bottom: 8px"
            >
              {{ formatStockSelectionStage(key) }}：{{ count }}
            </el-tag>
          </div>

          <div class="card" style="margin-top: 12px" v-if="selectedRun.stage_durations_ms">
            <div class="card-title">阶段耗时</div>
            <div class="chip-wrap">
              <el-tag
                v-for="(duration, key) in selectedRun.stage_durations_ms || {}"
                :key="`duration-${key}`"
                type="info"
              >
                {{ formatStockSelectionStage(key) }}：{{ duration }} ms
              </el-tag>
            </div>
          </div>

          <div class="card" style="margin-top: 12px">
            <div class="card-title">阶段日志</div>
            <el-table
              :data="selectedRun.stage_logs || []"
              border
              stripe
              size="small"
              empty-text="暂无阶段日志"
            >
              <el-table-column prop="stage_key" label="阶段" min-width="120">
                <template #default="{ row }">{{ formatStockSelectionStage(row.stage_key) }}</template>
              </el-table-column>
              <el-table-column prop="input_count" label="输入" min-width="80" />
              <el-table-column prop="output_count" label="输出" min-width="80" />
              <el-table-column prop="duration_ms" label="耗时(ms)" min-width="100" />
              <el-table-column prop="detail_message" label="说明" min-width="220" />
              <el-table-column label="转化率" min-width="110">
                <template #default="{ row }">
                  {{
                    row.input_count > 0
                      ? formatStockSelectionPercent(row.output_count / row.input_count, 1)
                      : "-"
                  }}
                </template>
              </el-table-column>
            </el-table>
          </div>
        </template>
      </div>
    </el-drawer>
  </StockSelectionModuleShell>
</template>

<style scoped>
.card-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}

.chip-wrap {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.mini-text {
  margin-top: 10px;
  color: var(--el-text-color-regular);
  font-size: 13px;
}
</style>
