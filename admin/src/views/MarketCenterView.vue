<script setup>
import { onMounted, reactive, ref } from "vue";
import {
  createFuturesStrategy,
  createStockRecommendation,
  generateDailyStockRecommendations,
  listFuturesStrategies,
  listStockRecommendations,
  updateFuturesStrategyStatus,
  updateStockRecommendationStatus
} from "../api/admin";

const activeTab = ref("stocks");
const errorMessage = ref("");
const message = ref("");
const refreshingAll = ref(false);

const stockLoading = ref(false);
const stockSubmitting = ref(false);
const stockGenerating = ref(false);
const stockPage = ref(1);
const stockPageSize = ref(20);
const stockTotal = ref(0);
const stocks = ref([]);
const stockFilters = reactive({
  status: ""
});
const stockDraftStatusMap = ref({});
const stockTradeDate = ref("");
const stockDialogVisible = ref(false);
const stockForm = reactive({
  symbol: "",
  name: "",
  score: 80,
  risk_level: "MEDIUM",
  position_range: "10%-20%",
  valid_from: "",
  valid_to: "",
  status: "PUBLISHED",
  reason_summary: ""
});

const futuresLoading = ref(false);
const futuresSubmitting = ref(false);
const futuresPage = ref(1);
const futuresPageSize = ref(20);
const futuresTotal = ref(0);
const futures = ref([]);
const futuresFilters = reactive({
  status: "",
  contract: ""
});
const futuresDraftStatusMap = ref({});
const futuresDialogVisible = ref(false);
const futuresForm = reactive({
  contract: "",
  name: "",
  direction: "LONG",
  risk_level: "MEDIUM",
  position_range: "10%-20%",
  valid_from: "",
  valid_to: "",
  status: "ACTIVE",
  reason_summary: ""
});

const stockStatusOptions = ["PUBLISHED", "DRAFT", "DISABLED", "ACTIVE"];
const futuresStatusOptions = ["ACTIVE", "DRAFT", "DISABLED", "PUBLISHED"];
const riskLevelOptions = ["LOW", "MEDIUM", "HIGH"];
const directionOptions = ["LONG", "SHORT", "NEUTRAL"];

function normalizeErrorMessage(error, fallback) {
  return error?.message || fallback || "操作失败";
}

function clearMessages() {
  errorMessage.value = "";
  message.value = "";
}

function statusTagType(status) {
  const normalized = (status || "").toUpperCase();
  if (["ACTIVE", "PUBLISHED", "SUCCESS"].includes(normalized)) return "success";
  if (["DRAFT", "PENDING", "RUNNING"].includes(normalized)) return "warning";
  if (["DISABLED", "FAILED", "REJECTED"].includes(normalized)) return "danger";
  return "info";
}

function syncStockDrafts() {
  const map = {};
  stocks.value.forEach((item) => {
    map[item.id] = item.status || "DRAFT";
  });
  stockDraftStatusMap.value = map;
}

function syncFuturesDrafts() {
  const map = {};
  futures.value.forEach((item) => {
    map[item.id] = item.status || "ACTIVE";
  });
  futuresDraftStatusMap.value = map;
}

function resetStockForm() {
  Object.assign(stockForm, {
    symbol: "",
    name: "",
    score: 80,
    risk_level: "MEDIUM",
    position_range: "10%-20%",
    valid_from: "",
    valid_to: "",
    status: "PUBLISHED",
    reason_summary: ""
  });
}

function resetFuturesForm() {
  Object.assign(futuresForm, {
    contract: "",
    name: "",
    direction: "LONG",
    risk_level: "MEDIUM",
    position_range: "10%-20%",
    valid_from: "",
    valid_to: "",
    status: "ACTIVE",
    reason_summary: ""
  });
}

async function fetchStocks(options = {}) {
  const { keepMessage = false } = options;
  stockLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listStockRecommendations({
      status: stockFilters.status,
      page: stockPage.value,
      page_size: stockPageSize.value
    });
    stocks.value = data.items || [];
    stockTotal.value = data.total || 0;
    syncStockDrafts();
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载股票推荐失败");
  } finally {
    stockLoading.value = false;
  }
}

async function fetchFutures(options = {}) {
  const { keepMessage = false } = options;
  futuresLoading.value = true;
  errorMessage.value = "";
  if (!keepMessage) {
    message.value = "";
  }
  try {
    const data = await listFuturesStrategies({
      status: futuresFilters.status,
      contract: futuresFilters.contract.trim(),
      page: futuresPage.value,
      page_size: futuresPageSize.value
    });
    futures.value = data.items || [];
    futuresTotal.value = data.total || 0;
    syncFuturesDrafts();
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "加载期货策略失败");
  } finally {
    futuresLoading.value = false;
  }
}

async function submitStock() {
  const payload = {
    symbol: stockForm.symbol.trim(),
    name: stockForm.name.trim(),
    score: Number(stockForm.score),
    risk_level: stockForm.risk_level,
    position_range: stockForm.position_range.trim(),
    valid_from: stockForm.valid_from.trim(),
    valid_to: stockForm.valid_to.trim(),
    status: stockForm.status,
    reason_summary: stockForm.reason_summary.trim()
  };
  if (!payload.symbol || !payload.name || !payload.valid_from || !payload.valid_to) {
    errorMessage.value = "请完整填写股票推荐必填字段";
    return;
  }
  stockSubmitting.value = true;
  clearMessages();
  try {
    await createStockRecommendation(payload);
    stockDialogVisible.value = false;
    resetStockForm();
    await fetchStocks({ keepMessage: true });
    message.value = `股票推荐 ${payload.symbol} 已创建`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "创建股票推荐失败");
  } finally {
    stockSubmitting.value = false;
  }
}

async function submitFutures() {
  const payload = {
    contract: futuresForm.contract.trim(),
    name: futuresForm.name.trim(),
    direction: futuresForm.direction,
    risk_level: futuresForm.risk_level,
    position_range: futuresForm.position_range.trim(),
    valid_from: futuresForm.valid_from.trim(),
    valid_to: futuresForm.valid_to.trim(),
    status: futuresForm.status,
    reason_summary: futuresForm.reason_summary.trim()
  };
  if (!payload.contract || !payload.direction || !payload.valid_from || !payload.valid_to) {
    errorMessage.value = "请完整填写期货策略必填字段";
    return;
  }
  futuresSubmitting.value = true;
  clearMessages();
  try {
    await createFuturesStrategy(payload);
    futuresDialogVisible.value = false;
    resetFuturesForm();
    await fetchFutures({ keepMessage: true });
    message.value = `期货策略 ${payload.contract} 已创建`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "创建期货策略失败");
  } finally {
    futuresSubmitting.value = false;
  }
}

async function saveStockStatus(item) {
  const target = (stockDraftStatusMap.value[item.id] || "").trim();
  if (!target || target === item.status) {
    return;
  }
  clearMessages();
  try {
    await updateStockRecommendationStatus(item.id, target);
    await fetchStocks({ keepMessage: true });
    message.value = `股票推荐 ${item.id} 状态已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "更新股票推荐状态失败");
  }
}

async function saveFuturesStatus(item) {
  const target = (futuresDraftStatusMap.value[item.id] || "").trim();
  if (!target || target === item.status) {
    return;
  }
  clearMessages();
  try {
    await updateFuturesStrategyStatus(item.id, target);
    await fetchFutures({ keepMessage: true });
    message.value = `期货策略 ${item.id} 状态已更新为 ${target}`;
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "更新期货策略状态失败");
  }
}

async function handleGenerateDailyStocks() {
  stockGenerating.value = true;
  clearMessages();
  try {
    const data = await generateDailyStockRecommendations(stockTradeDate.value.trim());
    message.value = `已生成每日股票推荐 ${data.count || 0} 条`;
    await fetchStocks({ keepMessage: true });
  } catch (error) {
    errorMessage.value = normalizeErrorMessage(error, "生成每日股票推荐失败");
  } finally {
    stockGenerating.value = false;
  }
}

function applyStockFilters() {
  stockPage.value = 1;
  fetchStocks();
}

function resetStockFilters() {
  stockFilters.status = "";
  stockPage.value = 1;
  fetchStocks();
}

function handleStockPageChange(nextPage) {
  if (nextPage === stockPage.value) {
    return;
  }
  stockPage.value = nextPage;
  fetchStocks();
}

function applyFuturesFilters() {
  futuresPage.value = 1;
  fetchFutures();
}

function resetFuturesFilters() {
  futuresFilters.status = "";
  futuresFilters.contract = "";
  futuresPage.value = 1;
  fetchFutures();
}

function handleFuturesPageChange(nextPage) {
  if (nextPage === futuresPage.value) {
    return;
  }
  futuresPage.value = nextPage;
  fetchFutures();
}

async function refreshCurrentTab() {
  if (activeTab.value === "stocks") {
    await fetchStocks();
    return;
  }
  await fetchFutures();
}

async function refreshAll() {
  refreshingAll.value = true;
  clearMessages();
  try {
    await Promise.all([fetchStocks({ keepMessage: true }), fetchFutures({ keepMessage: true })]);
    message.value = "策略中心数据已刷新";
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
        <h1 class="page-title">策略中心</h1>
        <p class="muted">股票推荐与期货策略维护、发布与批量生成</p>
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
      <el-tab-pane label="股票推荐" name="stocks">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select
              v-model="stockFilters.status"
              clearable
              filterable
              allow-create
              default-first-option
              placeholder="状态"
              style="width: 160px"
            >
              <el-option v-for="item in stockStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-input v-model="stockTradeDate" clearable placeholder="trade_date(YYYY-MM-DD，可选)" style="width: 220px" />
            <el-button type="primary" plain :loading="stockGenerating" @click="handleGenerateDailyStocks">
              生成每日推荐
            </el-button>
            <el-button type="primary" plain @click="applyStockFilters">查询</el-button>
            <el-button @click="resetStockFilters">重置</el-button>
            <el-button type="primary" @click="stockDialogVisible = true">新增股票推荐</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="stocks" border stripe v-loading="stockLoading" empty-text="暂无股票推荐">
            <el-table-column prop="id" label="推荐ID" min-width="130" />
            <el-table-column prop="symbol" label="代码" min-width="90" />
            <el-table-column prop="name" label="名称" min-width="120" />
            <el-table-column prop="score" label="评分" min-width="80" />
            <el-table-column prop="risk_level" label="风险等级" min-width="90" />
            <el-table-column prop="position_range" label="仓位建议" min-width="110" />
            <el-table-column prop="valid_from" label="生效起" min-width="150" />
            <el-table-column prop="valid_to" label="生效止" min-width="150" />
            <el-table-column label="状态" min-width="250">
              <template #default="{ row }">
                <div class="inline-actions">
                  <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
                  <el-select
                    v-model="stockDraftStatusMap[row.id]"
                    filterable
                    allow-create
                    default-first-option
                    style="width: 120px"
                  >
                    <el-option v-for="item in stockStatusOptions" :key="item" :label="item" :value="item" />
                  </el-select>
                  <el-button size="small" @click="saveStockStatus(row)">保存</el-button>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="推荐理由" min-width="220">
              <template #default="{ row }">
                {{ row.reason_summary || "-" }}
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ stockPage }} 页，共 {{ stockTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="stockPage"
              :page-size="stockPageSize"
              :total="stockTotal"
              @current-change="handleStockPageChange"
            />
          </div>
        </div>
      </el-tab-pane>

      <el-tab-pane label="期货策略" name="futures">
        <div class="card" style="margin-bottom: 12px">
          <div class="toolbar" style="margin-bottom: 0">
            <el-select
              v-model="futuresFilters.status"
              clearable
              filterable
              allow-create
              default-first-option
              placeholder="状态"
              style="width: 160px"
            >
              <el-option v-for="item in futuresStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <el-input v-model="futuresFilters.contract" clearable placeholder="合约代码" style="width: 180px" />
            <el-button type="primary" plain @click="applyFuturesFilters">查询</el-button>
            <el-button @click="resetFuturesFilters">重置</el-button>
            <el-button type="primary" @click="futuresDialogVisible = true">新增期货策略</el-button>
          </div>
        </div>

        <div class="card">
          <el-table :data="futures" border stripe v-loading="futuresLoading" empty-text="暂无期货策略">
            <el-table-column prop="id" label="策略ID" min-width="130" />
            <el-table-column prop="contract" label="合约" min-width="110" />
            <el-table-column prop="name" label="名称" min-width="120" />
            <el-table-column prop="direction" label="方向" min-width="90" />
            <el-table-column prop="risk_level" label="风险等级" min-width="90" />
            <el-table-column prop="position_range" label="仓位建议" min-width="110" />
            <el-table-column prop="valid_from" label="生效起" min-width="150" />
            <el-table-column prop="valid_to" label="生效止" min-width="150" />
            <el-table-column label="状态" min-width="250">
              <template #default="{ row }">
                <div class="inline-actions">
                  <el-tag :type="statusTagType(row.status)">{{ row.status }}</el-tag>
                  <el-select
                    v-model="futuresDraftStatusMap[row.id]"
                    filterable
                    allow-create
                    default-first-option
                    style="width: 120px"
                  >
                    <el-option v-for="item in futuresStatusOptions" :key="item" :label="item" :value="item" />
                  </el-select>
                  <el-button size="small" @click="saveFuturesStatus(row)">保存</el-button>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="策略说明" min-width="220">
              <template #default="{ row }">
                {{ row.reason_summary || "-" }}
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination">
            <el-text type="info">第 {{ futuresPage }} 页，共 {{ futuresTotal }} 条</el-text>
            <el-pagination
              background
              layout="prev, pager, next"
              :current-page="futuresPage"
              :page-size="futuresPageSize"
              :total="futuresTotal"
              @current-change="handleFuturesPageChange"
            />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="stockDialogVisible" title="新增股票推荐" width="620px" destroy-on-close>
      <el-form label-width="120px">
        <div class="dialog-grid">
          <el-form-item label="代码" required>
            <el-input v-model="stockForm.symbol" placeholder="如 600519.SH" />
          </el-form-item>
          <el-form-item label="名称" required>
            <el-input v-model="stockForm.name" placeholder="贵州茅台" />
          </el-form-item>
          <el-form-item label="评分" required>
            <el-input-number v-model="stockForm.score" :min="0" :max="100" :step="1" style="width: 100%" />
          </el-form-item>
          <el-form-item label="风险等级" required>
            <el-select v-model="stockForm.risk_level" style="width: 100%">
              <el-option v-for="item in riskLevelOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="仓位建议">
            <el-input v-model="stockForm.position_range" placeholder="10%-20%" />
          </el-form-item>
          <el-form-item label="状态" required>
            <el-select
              v-model="stockForm.status"
              filterable
              allow-create
              default-first-option
              style="width: 100%"
            >
              <el-option v-for="item in stockStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="生效起" required>
            <el-input v-model="stockForm.valid_from" placeholder="2026-03-01" />
          </el-form-item>
          <el-form-item label="生效止" required>
            <el-input v-model="stockForm.valid_to" placeholder="2026-03-31" />
          </el-form-item>
        </div>
        <el-form-item label="推荐理由">
          <el-input
            v-model="stockForm.reason_summary"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
            placeholder="趋势强、估值合理、行业景气上行"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="stockDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="stockSubmitting" @click="submitStock">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog v-model="futuresDialogVisible" title="新增期货策略" width="620px" destroy-on-close>
      <el-form label-width="120px">
        <div class="dialog-grid">
          <el-form-item label="合约" required>
            <el-input v-model="futuresForm.contract" placeholder="如 IF2406" />
          </el-form-item>
          <el-form-item label="名称">
            <el-input v-model="futuresForm.name" placeholder="股指趋势策略" />
          </el-form-item>
          <el-form-item label="方向" required>
            <el-select v-model="futuresForm.direction" style="width: 100%">
              <el-option v-for="item in directionOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="风险等级" required>
            <el-select v-model="futuresForm.risk_level" style="width: 100%">
              <el-option v-for="item in riskLevelOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="仓位建议">
            <el-input v-model="futuresForm.position_range" placeholder="10%-20%" />
          </el-form-item>
          <el-form-item label="状态" required>
            <el-select
              v-model="futuresForm.status"
              filterable
              allow-create
              default-first-option
              style="width: 100%"
            >
              <el-option v-for="item in futuresStatusOptions" :key="item" :label="item" :value="item" />
            </el-select>
          </el-form-item>
          <el-form-item label="生效起" required>
            <el-input v-model="futuresForm.valid_from" placeholder="2026-03-01" />
          </el-form-item>
          <el-form-item label="生效止" required>
            <el-input v-model="futuresForm.valid_to" placeholder="2026-03-31" />
          </el-form-item>
        </div>
        <el-form-item label="策略说明">
          <el-input
            v-model="futuresForm.reason_summary"
            type="textarea"
            :rows="3"
            maxlength="300"
            show-word-limit
            placeholder="趋势突破+波动率收敛，按风险等级动态止损"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="futuresDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="futuresSubmitting" @click="submitFutures">创建</el-button>
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

.dialog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 0 12px;
}
</style>
