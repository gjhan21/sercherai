<script setup>
import { onMounted, reactive, ref } from "vue";
import { ElMessage } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  createStockSelectionTemplate,
  listStockSelectionTemplates,
  setDefaultStockSelectionTemplate,
  updateStockSelectionTemplate
} from "../../api/admin";
import {
  formatStockSelectionMarketRegime,
  formatStockSelectionRiskLevel,
  formatStockSelectionStatus,
  formatStockSelectionUniverseScope,
  joinStockSelectionTextList,
  splitStockSelectionTextList,
  stockSelectionMarketRegimeOptions,
  stockSelectionRiskLevelOptions,
  stockSelectionStatusOptions,
  stockSelectionUniverseScopeOptions
} from "../../lib/stock-selection";
import { hasPermission } from "../../lib/session";

const canManage = hasPermission("stock_selection.manage");
const loading = ref(false);
const dialogVisible = ref(false);
const submitting = ref(false);
const editingID = ref("");
const templates = ref([]);

function createDefaultForm() {
  return {
    template_key: "",
    name: "",
    description: "",
    market_regime_bias: "ROTATION",
    status: "ACTIVE",
    is_default: false,
    universe_scope: "CN_A_ALL",
    min_listing_days: 180,
    min_avg_turnover: 50000000,
    exclude_st: true,
    exclude_suspended: true,
    price_min: 5,
    price_max: 300,
    volatility_min: 0,
    volatility_max: 8,
    industry_whitelist_text: "",
    industry_blacklist_text: "",
    sector_whitelist_text: "",
    sector_blacklist_text: "",
    theme_whitelist_text: "",
    theme_blacklist_text: "",
    bucket_limit: 36,
    seed_pool_cap: 180,
    candidate_pool_limit: 30,
    trend_bias: 1,
    money_flow_bias: 1,
    quality_bias: 1,
    event_bias: 1,
    resonance_bias: 1,
    quant_weight_pct: 70,
    event_weight_pct: 10,
    resonance_weight_pct: 10,
    liquidity_risk_weight_pct: 10,
    limit: 5,
    watchlist_limit: 5,
    max_symbol_per_bucket: 2,
    max_symbols_per_sector: 2,
    max_risk_level: "MEDIUM",
    min_score: 75,
    review_required: true,
    allow_auto_publish: false
  };
}

const form = reactive(createDefaultForm());

function resetForm() {
  editingID.value = "";
  Object.assign(form, createDefaultForm());
}

function openCreate() {
  resetForm();
  dialogVisible.value = true;
}

function openEdit(item) {
  editingID.value = item.id;
  Object.assign(form, {
    template_key: item.template_key || "",
    name: item.name || "",
    description: item.description || "",
    market_regime_bias: item.market_regime_bias || "ROTATION",
    status: item.status || "ACTIVE",
    is_default: Boolean(item.is_default),
    universe_scope: item.universe_defaults_json?.universe_scope || "CN_A_ALL",
    min_listing_days: Number(item.universe_defaults_json?.min_listing_days || 180),
    min_avg_turnover: Number(item.universe_defaults_json?.min_avg_turnover || 50000000),
    exclude_st: item.universe_defaults_json?.exclude_st !== false,
    exclude_suspended: item.universe_defaults_json?.exclude_suspended !== false,
    price_min: Number(item.universe_defaults_json?.price_min || 5),
    price_max: Number(item.universe_defaults_json?.price_max || 300),
    volatility_min: Number(item.universe_defaults_json?.volatility_min || 0),
    volatility_max: Number(item.universe_defaults_json?.volatility_max || 8),
    industry_whitelist_text: joinStockSelectionTextList(item.universe_defaults_json?.industry_whitelist),
    industry_blacklist_text: joinStockSelectionTextList(item.universe_defaults_json?.industry_blacklist),
    sector_whitelist_text: joinStockSelectionTextList(item.universe_defaults_json?.sector_whitelist),
    sector_blacklist_text: joinStockSelectionTextList(item.universe_defaults_json?.sector_blacklist),
    theme_whitelist_text: joinStockSelectionTextList(item.universe_defaults_json?.theme_whitelist),
    theme_blacklist_text: joinStockSelectionTextList(item.universe_defaults_json?.theme_blacklist),
    bucket_limit: Number(item.seed_defaults_json?.bucket_limit || 36),
    seed_pool_cap: Number(item.seed_defaults_json?.seed_pool_cap || 180),
    candidate_pool_limit: Number(item.seed_defaults_json?.candidate_pool_limit || 30),
    trend_bias: Number(item.seed_defaults_json?.trend_bias || 1),
    money_flow_bias: Number(item.seed_defaults_json?.money_flow_bias || 1),
    quality_bias: Number(item.seed_defaults_json?.quality_bias || 1),
    event_bias: Number(item.seed_defaults_json?.event_bias || 1),
    resonance_bias: Number(item.seed_defaults_json?.resonance_bias || 1),
    quant_weight_pct: Number(item.factor_defaults_json?.quant_weight || 0.7) * 100,
    event_weight_pct: Number(item.factor_defaults_json?.event_weight || 0.1) * 100,
    resonance_weight_pct: Number(item.factor_defaults_json?.resonance_weight || 0.1) * 100,
    liquidity_risk_weight_pct: Number(item.factor_defaults_json?.liquidity_risk_weight || 0.1) * 100,
    limit: Number(item.portfolio_defaults_json?.limit || 5),
    watchlist_limit: Number(item.portfolio_defaults_json?.watchlist_limit || 5),
    max_symbol_per_bucket: Number(item.portfolio_defaults_json?.max_symbol_per_bucket || 2),
    max_symbols_per_sector: Number(item.portfolio_defaults_json?.max_symbols_per_sector || 2),
    max_risk_level: item.portfolio_defaults_json?.max_risk_level || "MEDIUM",
    min_score: Number(item.portfolio_defaults_json?.min_score || 75),
    review_required: item.publish_defaults_json?.review_required !== false,
    allow_auto_publish: Boolean(item.publish_defaults_json?.allow_auto_publish)
  });
  dialogVisible.value = true;
}

function buildPayload() {
  const weightTotal =
    Number(form.quant_weight_pct) +
    Number(form.event_weight_pct) +
    Number(form.resonance_weight_pct) +
    Number(form.liquidity_risk_weight_pct);
  if (!String(form.template_key || "").trim()) {
    throw new Error("请填写模板标识");
  }
  if (!String(form.name || "").trim()) {
    throw new Error("请填写模板名称");
  }
  if (Math.abs(weightTotal - 100) > 0.001) {
    throw new Error("因子权重合计必须为 100%");
  }
  const industryWhitelist = splitStockSelectionTextList(form.industry_whitelist_text);
  const industryBlacklist = splitStockSelectionTextList(form.industry_blacklist_text);
  const sectorWhitelist = splitStockSelectionTextList(form.sector_whitelist_text);
  const sectorBlacklist = splitStockSelectionTextList(form.sector_blacklist_text);
  const themeWhitelist = splitStockSelectionTextList(form.theme_whitelist_text);
  const themeBlacklist = splitStockSelectionTextList(form.theme_blacklist_text);
  return {
    template_key: String(form.template_key || "").trim(),
    name: String(form.name || "").trim(),
    description: String(form.description || "").trim(),
    market_regime_bias: form.market_regime_bias,
    is_default: Boolean(form.is_default),
    status: form.status,
    universe_defaults_json: {
      universe_scope: form.universe_scope,
      min_listing_days: Number(form.min_listing_days),
      min_avg_turnover: Number(form.min_avg_turnover),
      exclude_st: Boolean(form.exclude_st),
      exclude_suspended: Boolean(form.exclude_suspended),
      price_min: Number(form.price_min),
      price_max: Number(form.price_max),
      volatility_min: Number(form.volatility_min),
      volatility_max: Number(form.volatility_max),
      industry_whitelist: industryWhitelist,
      industry_blacklist: industryBlacklist,
      sector_whitelist: sectorWhitelist,
      sector_blacklist: sectorBlacklist,
      theme_whitelist: themeWhitelist,
      theme_blacklist: themeBlacklist
    },
    seed_defaults_json: {
      bucket_limit: Number(form.bucket_limit),
      seed_pool_cap: Number(form.seed_pool_cap),
      candidate_pool_limit: Number(form.candidate_pool_limit),
      trend_bias: Number(form.trend_bias),
      money_flow_bias: Number(form.money_flow_bias),
      quality_bias: Number(form.quality_bias),
      event_bias: Number(form.event_bias),
      resonance_bias: Number(form.resonance_bias)
    },
    factor_defaults_json: {
      quant_weight: Number(form.quant_weight_pct) / 100,
      event_weight: Number(form.event_weight_pct) / 100,
      resonance_weight: Number(form.resonance_weight_pct) / 100,
      liquidity_risk_weight: Number(form.liquidity_risk_weight_pct) / 100
    },
    portfolio_defaults_json: {
      limit: Number(form.limit),
      watchlist_limit: Number(form.watchlist_limit),
      max_symbol_per_bucket: Number(form.max_symbol_per_bucket),
      max_symbols_per_sector: Number(form.max_symbols_per_sector),
      max_risk_level: form.max_risk_level,
      min_score: Number(form.min_score)
    },
    publish_defaults_json: {
      review_required: Boolean(form.review_required),
      allow_auto_publish: Boolean(form.allow_auto_publish)
    }
  };
}

async function fetchTemplates() {
  loading.value = true;
  try {
    const data = await listStockSelectionTemplates({ page: 1, page_size: 100 });
    templates.value = Array.isArray(data?.items) ? data.items : [];
  } catch (error) {
    ElMessage.error(error?.message || "加载策略模板失败");
  } finally {
    loading.value = false;
  }
}

async function submitForm() {
  submitting.value = true;
  try {
    const payload = buildPayload();
    if (editingID.value) {
      await updateStockSelectionTemplate(editingID.value, payload);
      ElMessage.success("策略模板已更新");
    } else {
      await createStockSelectionTemplate(payload);
      ElMessage.success("策略模板已创建");
    }
    dialogVisible.value = false;
    await fetchTemplates();
  } catch (error) {
    ElMessage.error(error?.message || "保存策略模板失败");
  } finally {
    submitting.value = false;
  }
}

async function handleSetDefault(item) {
  try {
    await setDefaultStockSelectionTemplate(item.id);
    ElMessage.success("默认模板已切换");
    await fetchTemplates();
  } catch (error) {
    ElMessage.error(error?.message || "切换默认模板失败");
  }
}

onMounted(fetchTemplates);
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股策略模板"
    description="系统模板与运营自定义模板统一在这里管理。模板定义默认参数和市场状态偏好，但不会直接发布。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchTemplates">刷新模板</el-button>
        <el-button v-if="canManage" type="primary" @click="openCreate">新建模板</el-button>
      </div>
    </template>

    <div class="card">
      <el-table :data="templates" border stripe size="small" v-loading="loading" empty-text="暂无策略模板">
        <el-table-column prop="name" label="模板名称" min-width="160" />
        <el-table-column prop="template_key" label="模板标识" min-width="160" />
        <el-table-column prop="market_regime_bias" label="市场偏好" min-width="120">
          <template #default="{ row }">{{ formatStockSelectionMarketRegime(row.market_regime_bias) }}</template>
        </el-table-column>
        <el-table-column label="默认股票池" min-width="130">
          <template #default="{ row }">
            {{ formatStockSelectionUniverseScope(row.universe_defaults_json?.universe_scope) }}
          </template>
        </el-table-column>
        <el-table-column label="组合约束" min-width="220">
          <template #default="{ row }">
            组合 {{ row.portfolio_defaults_json?.limit || 5 }} 只 / 观察 {{ row.portfolio_defaults_json?.watchlist_limit || 5 }} 只 /
            {{ formatStockSelectionRiskLevel(row.portfolio_defaults_json?.max_risk_level) }}
          </template>
        </el-table-column>
        <el-table-column label="默认" min-width="90">
          <template #default="{ row }">
            <el-tag :type="row.is_default ? 'success' : 'info'">{{ row.is_default ? "默认" : "候选" }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="100">
          <template #default="{ row }">{{ formatStockSelectionStatus(row.status) }}</template>
        </el-table-column>
        <el-table-column prop="description" label="说明" min-width="260" show-overflow-tooltip />
        <el-table-column label="操作" min-width="200" fixed="right">
          <template #default="{ row }">
            <div class="row-actions">
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button v-if="canManage" size="small" type="success" @click="handleSetDefault(row)">
                设为默认
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="dialogVisible" :title="editingID ? '编辑策略模板' : '新建策略模板'" width="980px">
      <el-form label-position="top" class="template-form">
        <div class="profile-form-grid">
          <el-form-item label="模板标识" required>
            <el-input v-model="form.template_key" placeholder="例如 trend_growth" />
          </el-form-item>
          <el-form-item label="模板名称" required>
            <el-input v-model="form.name" placeholder="例如 趋势成长" />
          </el-form-item>
          <el-form-item label="市场状态偏好">
            <el-select v-model="form.market_regime_bias">
              <el-option
                v-for="item in stockSelectionMarketRegimeOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="form.status">
              <el-option
                v-for="item in stockSelectionStatusOptions"
                :key="item.value"
                :label="item.label"
                :value="item.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="模板说明" class="grid-span-2">
            <el-input v-model="form.description" type="textarea" :rows="3" placeholder="说明这个模板更适合什么市场环境和研究目标" />
          </el-form-item>
        </div>

        <div class="section-card">
          <div class="section-title">股票池与种子</div>
          <div class="profile-form-grid">
            <el-form-item label="默认股票池范围">
              <el-select v-model="form.universe_scope">
                <el-option
                  v-for="item in stockSelectionUniverseScopeOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="最少上市天数">
              <el-input-number v-model="form.min_listing_days" :min="0" :step="30" />
            </el-form-item>
            <el-form-item label="最少 20 日均成交额">
              <el-input-number v-model="form.min_avg_turnover" :min="0" :step="10000000" />
            </el-form-item>
            <el-form-item label="排除 ST 股票">
              <el-switch v-model="form.exclude_st" active-text="排除" inactive-text="保留" />
            </el-form-item>
            <el-form-item label="排除停牌股票">
              <el-switch v-model="form.exclude_suspended" active-text="排除" inactive-text="保留" />
            </el-form-item>
            <el-form-item label="最低股价（元）">
              <el-input-number v-model="form.price_min" :min="0" :step="1" />
            </el-form-item>
            <el-form-item label="最高股价（元）">
              <el-input-number v-model="form.price_max" :min="0" :step="10" />
            </el-form-item>
            <el-form-item label="最低 20 日波动率（%）">
              <el-input-number v-model="form.volatility_min" :min="0" :step="0.5" />
            </el-form-item>
            <el-form-item label="最高 20 日波动率（%）">
              <el-input-number v-model="form.volatility_max" :min="0" :step="0.5" />
            </el-form-item>
            <el-form-item label="单桶基础名额">
              <el-input-number v-model="form.bucket_limit" :min="1" :max="120" />
            </el-form-item>
            <el-form-item label="种子池上限">
              <el-input-number v-model="form.seed_pool_cap" :min="1" :max="500" />
            </el-form-item>
            <el-form-item label="候选池上限">
              <el-input-number v-model="form.candidate_pool_limit" :min="1" :max="120" />
            </el-form-item>
            <el-form-item label="趋势桶偏置">
              <el-slider v-model="form.trend_bias" :min="0.5" :max="1.5" :step="0.1" show-input />
            </el-form-item>
            <el-form-item label="资金桶偏置">
              <el-slider v-model="form.money_flow_bias" :min="0.5" :max="1.5" :step="0.1" show-input />
            </el-form-item>
            <el-form-item label="质量桶偏置">
              <el-slider v-model="form.quality_bias" :min="0.5" :max="1.5" :step="0.1" show-input />
            </el-form-item>
            <el-form-item label="事件桶偏置">
              <el-slider v-model="form.event_bias" :min="0.5" :max="1.5" :step="0.1" show-input />
            </el-form-item>
            <el-form-item label="共振桶偏置">
              <el-slider v-model="form.resonance_bias" :min="0.5" :max="1.5" :step="0.1" show-input />
            </el-form-item>
          </div>
          <div class="profile-form-grid">
            <el-form-item label="行业白名单">
              <el-input v-model="form.industry_whitelist_text" type="textarea" :rows="3" placeholder="每行一个，留空表示不限" />
            </el-form-item>
            <el-form-item label="行业黑名单">
              <el-input v-model="form.industry_blacklist_text" type="textarea" :rows="3" placeholder="每行一个，命中则排除" />
            </el-form-item>
            <el-form-item label="板块白名单">
              <el-input v-model="form.sector_whitelist_text" type="textarea" :rows="3" placeholder="每行一个，留空表示不限" />
            </el-form-item>
            <el-form-item label="板块黑名单">
              <el-input v-model="form.sector_blacklist_text" type="textarea" :rows="3" placeholder="每行一个，命中则排除" />
            </el-form-item>
            <el-form-item label="题材白名单">
              <el-input v-model="form.theme_whitelist_text" type="textarea" :rows="3" placeholder="每行一个，留空表示不限" />
            </el-form-item>
            <el-form-item label="题材黑名单">
              <el-input v-model="form.theme_blacklist_text" type="textarea" :rows="3" placeholder="每行一个，命中则排除" />
            </el-form-item>
          </div>
        </div>

        <div class="section-card">
          <div class="section-title">因子与组合</div>
          <div class="profile-form-grid">
            <el-form-item label="量化主分权重(%)">
              <el-input-number v-model="form.quant_weight_pct" :min="0" :max="100" />
            </el-form-item>
            <el-form-item label="事件确认权重(%)">
              <el-input-number v-model="form.event_weight_pct" :min="0" :max="100" />
            </el-form-item>
            <el-form-item label="共振确认权重(%)">
              <el-input-number v-model="form.resonance_weight_pct" :min="0" :max="100" />
            </el-form-item>
            <el-form-item label="风险修正权重(%)">
              <el-input-number v-model="form.liquidity_risk_weight_pct" :min="0" :max="100" />
            </el-form-item>
            <el-form-item label="发布组合数量">
              <el-input-number v-model="form.limit" :min="1" :max="10" />
            </el-form-item>
            <el-form-item label="观察名单数量">
              <el-input-number v-model="form.watchlist_limit" :min="0" :max="20" />
            </el-form-item>
            <el-form-item label="单桶最多入选">
              <el-input-number v-model="form.max_symbol_per_bucket" :min="1" :max="10" />
            </el-form-item>
            <el-form-item label="单行业最多入选">
              <el-input-number v-model="form.max_symbols_per_sector" :min="1" :max="10" />
            </el-form-item>
            <el-form-item label="最高风险级别">
              <el-select v-model="form.max_risk_level">
                <el-option
                  v-for="item in stockSelectionRiskLevelOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="最小分数">
              <el-slider v-model="form.min_score" :min="50" :max="95" :step="1" show-input />
            </el-form-item>
          </div>
        </div>

        <div class="section-card">
          <div class="section-title">发布规则</div>
          <div class="profile-form-grid">
            <el-form-item label="需人工审核">
              <el-switch v-model="form.review_required" active-text="需要" inactive-text="不需要" />
            </el-form-item>
            <el-form-item label="允许自动发布">
              <el-switch v-model="form.allow_auto_publish" active-text="允许" inactive-text="关闭" />
            </el-form-item>
            <el-form-item label="设为默认建议模板">
              <el-switch v-model="form.is_default" active-text="是" inactive-text="否" />
            </el-form-item>
          </div>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">保存模板</el-button>
      </template>
    </el-dialog>
  </StockSelectionModuleShell>
</template>

<style scoped>
.row-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.profile-form-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.grid-span-2 {
  grid-column: span 2;
}

.section-card {
  padding: 16px;
  margin-top: 12px;
  border: 1px solid var(--el-border-color-light);
  border-radius: 14px;
  background: var(--el-bg-color-page);
}

.section-title {
  margin-bottom: 12px;
  font-size: 15px;
  font-weight: 600;
}
</style>
