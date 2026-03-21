<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";
import StockSelectionModuleShell from "../../components/StockSelectionModuleShell.vue";
import {
  createStockSelectionProfile,
  listStockSelectionProfiles,
  listStockSelectionTemplates,
  publishStockSelectionProfile,
  rollbackStockSelectionProfile,
  updateStockSelectionProfile
} from "../../api/admin";
import {
  formatStockSelectionMarketRegime,
  countExtraConfigKeys,
  formatStockSelectionMode,
  formatStockSelectionRiskLevel,
  formatStockSelectionStatus,
  formatStockSelectionUniverseScope,
  joinStockSelectionSymbols,
  joinStockSelectionTextList,
  splitStockSelectionSymbols,
  splitStockSelectionTextList,
  stockSelectionModeOptions,
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
const profiles = ref([]);
const templates = ref([]);

const KNOWN_CONFIG_KEYS = {
  universe: [
    "universe_scope",
    "min_listing_days",
    "min_avg_turnover",
    "exclude_st",
    "exclude_suspended",
    "excluded_symbols",
    "price_min",
    "price_max",
    "volatility_min",
    "volatility_max",
    "industry_whitelist",
    "industry_blacklist",
    "sector_whitelist",
    "sector_blacklist",
    "theme_whitelist",
    "theme_blacklist"
  ],
  seed: [
    "mode",
    "bucket_limit",
    "seed_pool_cap",
    "candidate_pool_limit",
    "seed_symbols",
    "debug_seed_symbols",
    "trend_bias",
    "money_flow_bias",
    "quality_bias",
    "event_bias",
    "resonance_bias"
  ],
  factor: ["lookback_days", "quant_weight", "event_weight", "resonance_weight", "liquidity_risk_weight"],
  portfolio: ["limit", "max_risk_level", "min_score", "watchlist_limit", "max_symbol_per_bucket", "max_symbols_per_sector"],
  publish: ["review_required", "allow_auto_publish"]
};

const DEFAULTS = {
  selectionMode: "AUTO",
  universeScope: "CN_A_ALL",
  universe: {
    min_listing_days: 180,
    min_avg_turnover: 50000000,
    exclude_st: true,
    exclude_suspended: true,
    price_min: 5,
    price_max: 300,
    volatility_min: 0,
    volatility_max: 8
  },
  seed: {
    bucket_limit: 36,
    seed_pool_cap: 180,
    candidate_pool_limit: 30,
    trend_bias: 1,
    money_flow_bias: 1,
    quality_bias: 1,
    event_bias: 1,
    resonance_bias: 1
  },
  factor: {
    lookback_days: 120,
    quant_weight_pct: 75,
    event_weight_pct: 10,
    resonance_weight_pct: 10,
    liquidity_risk_weight_pct: 5
  },
  portfolio: {
    limit: 5,
    watchlist_limit: 5,
    max_symbol_per_bucket: 2,
    max_symbols_per_sector: 2,
    max_risk_level: "MEDIUM",
    min_score: 75
  },
  publish: {
    review_required: true,
    allow_auto_publish: false
  }
};

function numberValue(value, fallback) {
  const numeric = Number(value);
  return Number.isFinite(numeric) ? numeric : fallback;
}

function booleanValue(value, fallback) {
  if (typeof value === "boolean") {
    return value;
  }
  if (value === undefined || value === null || value === "") {
    return fallback;
  }
  return Boolean(value);
}

function pickExtraConfig(config, knownKeys) {
  return Object.fromEntries(
    Object.entries(config || {}).filter(([key]) => !knownKeys.includes(key))
  );
}

function toPercent(value, fallback) {
  return numberValue(value, fallback / 100) * 100;
}

function buildFormState(profile = null) {
  const universeConfig = profile?.universe_config || {};
  const seedConfig = profile?.seed_mining_config || {};
  const factorConfig = profile?.factor_config || {};
  const portfolioConfig = profile?.portfolio_config || {};
  const publishConfig = profile?.publish_config || {};
  const selectionMode = profile?.selection_mode_default || seedConfig.mode || DEFAULTS.selectionMode;
  const universeScope = universeConfig.universe_scope || profile?.universe_scope || DEFAULTS.universeScope;

  return {
    name: profile?.name || "",
    template_id: profile?.template_id || "",
    status: profile?.status || "ACTIVE",
    is_default: Boolean(profile?.is_default),
    selection_mode_default: selectionMode,
    universe_scope: universeScope,
    universe: {
      min_listing_days: numberValue(universeConfig.min_listing_days, DEFAULTS.universe.min_listing_days),
      min_avg_turnover: numberValue(universeConfig.min_avg_turnover, DEFAULTS.universe.min_avg_turnover),
      exclude_st: booleanValue(universeConfig.exclude_st, DEFAULTS.universe.exclude_st),
      exclude_suspended: booleanValue(universeConfig.exclude_suspended, DEFAULTS.universe.exclude_suspended),
      price_min: numberValue(universeConfig.price_min, DEFAULTS.universe.price_min),
      price_max: numberValue(universeConfig.price_max, DEFAULTS.universe.price_max),
      volatility_min: numberValue(universeConfig.volatility_min, DEFAULTS.universe.volatility_min),
      volatility_max: numberValue(universeConfig.volatility_max, DEFAULTS.universe.volatility_max),
      excluded_symbols_text: joinStockSelectionSymbols(universeConfig.excluded_symbols),
      industry_whitelist_text: joinStockSelectionTextList(universeConfig.industry_whitelist),
      industry_blacklist_text: joinStockSelectionTextList(universeConfig.industry_blacklist),
      sector_whitelist_text: joinStockSelectionTextList(universeConfig.sector_whitelist),
      sector_blacklist_text: joinStockSelectionTextList(universeConfig.sector_blacklist),
      theme_whitelist_text: joinStockSelectionTextList(universeConfig.theme_whitelist),
      theme_blacklist_text: joinStockSelectionTextList(universeConfig.theme_blacklist)
    },
    seed: {
      bucket_limit: numberValue(seedConfig.bucket_limit, DEFAULTS.seed.bucket_limit),
      seed_pool_cap: numberValue(seedConfig.seed_pool_cap, DEFAULTS.seed.seed_pool_cap),
      candidate_pool_limit: numberValue(seedConfig.candidate_pool_limit, DEFAULTS.seed.candidate_pool_limit),
      trend_bias: numberValue(seedConfig.trend_bias, DEFAULTS.seed.trend_bias),
      money_flow_bias: numberValue(seedConfig.money_flow_bias, DEFAULTS.seed.money_flow_bias),
      quality_bias: numberValue(seedConfig.quality_bias, DEFAULTS.seed.quality_bias),
      event_bias: numberValue(seedConfig.event_bias, DEFAULTS.seed.event_bias),
      resonance_bias: numberValue(seedConfig.resonance_bias, DEFAULTS.seed.resonance_bias),
      seed_symbols_text: joinStockSelectionSymbols(seedConfig.seed_symbols),
      debug_seed_symbols_text: joinStockSelectionSymbols(seedConfig.debug_seed_symbols)
    },
    factor: {
      lookback_days: numberValue(factorConfig.lookback_days, DEFAULTS.factor.lookback_days),
      quant_weight_pct: toPercent(factorConfig.quant_weight, DEFAULTS.factor.quant_weight_pct),
      event_weight_pct: toPercent(factorConfig.event_weight, DEFAULTS.factor.event_weight_pct),
      resonance_weight_pct: toPercent(factorConfig.resonance_weight, DEFAULTS.factor.resonance_weight_pct),
      liquidity_risk_weight_pct: toPercent(
        factorConfig.liquidity_risk_weight,
        DEFAULTS.factor.liquidity_risk_weight_pct
      )
    },
    portfolio: {
      limit: numberValue(portfolioConfig.limit, DEFAULTS.portfolio.limit),
      watchlist_limit: numberValue(portfolioConfig.watchlist_limit, DEFAULTS.portfolio.watchlist_limit),
      max_symbol_per_bucket: numberValue(portfolioConfig.max_symbol_per_bucket, DEFAULTS.portfolio.max_symbol_per_bucket),
      max_symbols_per_sector: numberValue(portfolioConfig.max_symbols_per_sector, DEFAULTS.portfolio.max_symbols_per_sector),
      max_risk_level: portfolioConfig.max_risk_level || DEFAULTS.portfolio.max_risk_level,
      min_score: numberValue(portfolioConfig.min_score, DEFAULTS.portfolio.min_score)
    },
    publish: {
      review_required: booleanValue(publishConfig.review_required, DEFAULTS.publish.review_required),
      allow_auto_publish: booleanValue(publishConfig.allow_auto_publish, DEFAULTS.publish.allow_auto_publish)
    },
    description: profile?.description || "",
    change_note: "",
    extraConfigs: {
      universe: pickExtraConfig(universeConfig, KNOWN_CONFIG_KEYS.universe),
      seed: pickExtraConfig(seedConfig, KNOWN_CONFIG_KEYS.seed),
      factor: pickExtraConfig(factorConfig, KNOWN_CONFIG_KEYS.factor),
      portfolio: pickExtraConfig(portfolioConfig, KNOWN_CONFIG_KEYS.portfolio),
      publish: pickExtraConfig(publishConfig, KNOWN_CONFIG_KEYS.publish)
    }
  };
}

const form = reactive(buildFormState());

const factorWeightTotal = computed(
  () =>
    numberValue(form.factor.quant_weight_pct, 0) +
    numberValue(form.factor.event_weight_pct, 0) +
    numberValue(form.factor.resonance_weight_pct, 0) +
    numberValue(form.factor.liquidity_risk_weight_pct, 0)
);

const isManualMode = computed(() => form.selection_mode_default === "MANUAL");
const isDebugMode = computed(() => form.selection_mode_default === "DEBUG");
const preservedExtraCount = computed(() => countExtraConfigKeys(form.extraConfigs));

async function fetchProfiles() {
  loading.value = true;
  try {
    const data = await listStockSelectionProfiles({ page: 1, page_size: 100 });
    profiles.value = Array.isArray(data?.items) ? data.items : [];
  } catch (error) {
    ElMessage.error(error?.message || "加载配置方案失败");
  } finally {
    loading.value = false;
  }
}

async function fetchTemplates() {
  try {
    const data = await listStockSelectionTemplates({ page: 1, page_size: 100, status: "ACTIVE" });
    templates.value = Array.isArray(data?.items) ? data.items : [];
  } catch (error) {
    ElMessage.error(error?.message || "加载模板失败");
  }
}

function resetForm() {
  editingID.value = "";
  Object.assign(form, buildFormState());
}

function openCreate() {
  resetForm();
  dialogVisible.value = true;
}

function openEdit(profile) {
  editingID.value = profile.id;
  Object.assign(form, buildFormState(profile));
  dialogVisible.value = true;
}

function buildPayload() {
  const name = String(form.name || "").trim();
  if (!name) {
    throw new Error("请填写配置方案名称");
  }
  if (Math.abs(factorWeightTotal.value - 100) > 0.001) {
    throw new Error("因子权重合计必须为 100%");
  }

  const excludedSymbols = splitStockSelectionSymbols(form.universe.excluded_symbols_text);
  const industryWhitelist = splitStockSelectionTextList(form.universe.industry_whitelist_text);
  const industryBlacklist = splitStockSelectionTextList(form.universe.industry_blacklist_text);
  const sectorWhitelist = splitStockSelectionTextList(form.universe.sector_whitelist_text);
  const sectorBlacklist = splitStockSelectionTextList(form.universe.sector_blacklist_text);
  const themeWhitelist = splitStockSelectionTextList(form.universe.theme_whitelist_text);
  const themeBlacklist = splitStockSelectionTextList(form.universe.theme_blacklist_text);
  const seedSymbols = splitStockSelectionSymbols(form.seed.seed_symbols_text);
  const debugSeedSymbols = splitStockSelectionSymbols(form.seed.debug_seed_symbols_text);

  if (isManualMode.value && seedSymbols.length === 0) {
    throw new Error("手动指定模式至少需要填写 1 个股票代码");
  }
  if (isDebugMode.value && debugSeedSymbols.length === 0) {
    throw new Error("调试模式至少需要填写 1 个调试股票代码");
  }

  const universeConfig = {
    ...form.extraConfigs.universe,
    universe_scope: form.universe_scope,
    min_listing_days: numberValue(form.universe.min_listing_days, DEFAULTS.universe.min_listing_days),
    min_avg_turnover: numberValue(form.universe.min_avg_turnover, DEFAULTS.universe.min_avg_turnover),
    price_min: numberValue(form.universe.price_min, DEFAULTS.universe.price_min),
    price_max: numberValue(form.universe.price_max, DEFAULTS.universe.price_max),
    volatility_min: numberValue(form.universe.volatility_min, DEFAULTS.universe.volatility_min),
    volatility_max: numberValue(form.universe.volatility_max, DEFAULTS.universe.volatility_max),
    exclude_st: Boolean(form.universe.exclude_st),
    exclude_suspended: Boolean(form.universe.exclude_suspended)
  };
  if (excludedSymbols.length > 0) {
    universeConfig.excluded_symbols = excludedSymbols;
  }
  if (industryWhitelist.length > 0) {
    universeConfig.industry_whitelist = industryWhitelist;
  }
  if (industryBlacklist.length > 0) {
    universeConfig.industry_blacklist = industryBlacklist;
  }
  if (sectorWhitelist.length > 0) {
    universeConfig.sector_whitelist = sectorWhitelist;
  }
  if (sectorBlacklist.length > 0) {
    universeConfig.sector_blacklist = sectorBlacklist;
  }
  if (themeWhitelist.length > 0) {
    universeConfig.theme_whitelist = themeWhitelist;
  }
  if (themeBlacklist.length > 0) {
    universeConfig.theme_blacklist = themeBlacklist;
  }

  const seedMiningConfig = {
    ...form.extraConfigs.seed,
    mode: form.selection_mode_default,
    bucket_limit: numberValue(form.seed.bucket_limit, DEFAULTS.seed.bucket_limit),
    seed_pool_cap: numberValue(form.seed.seed_pool_cap, DEFAULTS.seed.seed_pool_cap),
    candidate_pool_limit: numberValue(form.seed.candidate_pool_limit, DEFAULTS.seed.candidate_pool_limit),
    trend_bias: numberValue(form.seed.trend_bias, DEFAULTS.seed.trend_bias),
    money_flow_bias: numberValue(form.seed.money_flow_bias, DEFAULTS.seed.money_flow_bias),
    quality_bias: numberValue(form.seed.quality_bias, DEFAULTS.seed.quality_bias),
    event_bias: numberValue(form.seed.event_bias, DEFAULTS.seed.event_bias),
    resonance_bias: numberValue(form.seed.resonance_bias, DEFAULTS.seed.resonance_bias)
  };
  if (seedSymbols.length > 0) {
    seedMiningConfig.seed_symbols = seedSymbols;
  }
  if (debugSeedSymbols.length > 0) {
    seedMiningConfig.debug_seed_symbols = debugSeedSymbols;
  }

  const factorConfig = {
    ...form.extraConfigs.factor,
    lookback_days: numberValue(form.factor.lookback_days, DEFAULTS.factor.lookback_days),
    quant_weight: numberValue(form.factor.quant_weight_pct, DEFAULTS.factor.quant_weight_pct) / 100,
    event_weight: numberValue(form.factor.event_weight_pct, DEFAULTS.factor.event_weight_pct) / 100,
    resonance_weight: numberValue(form.factor.resonance_weight_pct, DEFAULTS.factor.resonance_weight_pct) / 100,
    liquidity_risk_weight:
      numberValue(form.factor.liquidity_risk_weight_pct, DEFAULTS.factor.liquidity_risk_weight_pct) / 100
  };

  const portfolioConfig = {
    ...form.extraConfigs.portfolio,
    limit: numberValue(form.portfolio.limit, DEFAULTS.portfolio.limit),
    watchlist_limit: numberValue(form.portfolio.watchlist_limit, DEFAULTS.portfolio.watchlist_limit),
    max_symbol_per_bucket: numberValue(
      form.portfolio.max_symbol_per_bucket,
      DEFAULTS.portfolio.max_symbol_per_bucket
    ),
    max_symbols_per_sector: numberValue(
      form.portfolio.max_symbols_per_sector,
      DEFAULTS.portfolio.max_symbols_per_sector
    ),
    max_risk_level: form.portfolio.max_risk_level,
    min_score: numberValue(form.portfolio.min_score, DEFAULTS.portfolio.min_score)
  };

  const publishConfig = {
    ...form.extraConfigs.publish,
    review_required: Boolean(form.publish.review_required),
    allow_auto_publish: Boolean(form.publish.allow_auto_publish)
  };

  return {
    name,
    template_id: form.template_id,
    status: form.status,
    is_default: form.is_default,
    selection_mode_default: form.selection_mode_default,
    universe_scope: form.universe_scope,
    universe_config: universeConfig,
    seed_mining_config: seedMiningConfig,
    factor_config: factorConfig,
    portfolio_config: portfolioConfig,
    publish_config: publishConfig,
    description: String(form.description || "").trim(),
    change_note: String(form.change_note || "").trim()
  };
}

async function submitForm() {
  submitting.value = true;
  try {
    const payload = buildPayload();
    if (editingID.value) {
      await updateStockSelectionProfile(editingID.value, payload);
      ElMessage.success("配置方案已更新");
    } else {
      await createStockSelectionProfile(payload);
      ElMessage.success("配置方案已创建");
    }
    dialogVisible.value = false;
    await fetchProfiles();
  } catch (error) {
    ElMessage.error(error?.message || "保存配置方案失败");
  } finally {
    submitting.value = false;
  }
}

async function setDefaultProfile(profile) {
  try {
    await publishStockSelectionProfile(profile.id);
    ElMessage.success("默认配置方案已切换");
    await fetchProfiles();
  } catch (error) {
    ElMessage.error(error?.message || "切换默认配置方案失败");
  }
}

async function rollbackProfile(profile) {
  try {
    const result = await ElMessageBox.prompt("请输入要回滚到的版本号", "回滚配置方案", {
      inputPattern: /^[0-9]+$/,
      inputErrorMessage: "版本号必须是正整数",
      confirmButtonText: "回滚",
      cancelButtonText: "取消"
    });
    await rollbackStockSelectionProfile(profile.id, {
      version_no: Number(result.value),
      change_note: `后台回滚到 v${result.value}`
    });
    ElMessage.success("配置方案已回滚并生成新版本");
    await fetchProfiles();
  } catch (error) {
    if (error === "cancel") {
      return;
    }
    ElMessage.error(error?.message || "回滚失败");
  }
}

onMounted(async () => {
  await Promise.all([fetchProfiles(), fetchTemplates()]);
});
</script>

<template>
  <StockSelectionModuleShell
    title="智能选股策略配置"
    description="把股票池、种子挖掘、因子权重、组合规则和发布规则都做成中文可视化表单，保存时仍兼容现有配置结构。"
  >
    <template #actions>
      <div class="toolbar" style="margin-bottom: 0; flex-wrap: wrap">
        <el-button :loading="loading" @click="fetchProfiles">刷新配置</el-button>
        <el-button v-if="canManage" type="primary" @click="openCreate">新建配置方案</el-button>
      </div>
    </template>

    <div class="card">
      <el-table :data="profiles" border stripe size="small" v-loading="loading" empty-text="暂无配置方案">
        <el-table-column prop="name" label="方案名称" min-width="180" />
        <el-table-column label="默认" min-width="90">
          <template #default="{ row }">
            <el-tag :type="row.is_default ? 'success' : 'info'">
              {{ row.is_default ? "默认" : "候选" }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="current_version" label="当前版本" min-width="100">
          <template #default="{ row }">v{{ row.current_version || 0 }}</template>
        </el-table-column>
        <el-table-column prop="selection_mode_default" label="运行模式" min-width="120">
          <template #default="{ row }">{{ formatStockSelectionMode(row.selection_mode_default) }}</template>
        </el-table-column>
        <el-table-column prop="template_name" label="绑定模板" min-width="160">
          <template #default="{ row }">
            {{ row.template_name || "未指定模板" }}
          </template>
        </el-table-column>
        <el-table-column prop="universe_scope" label="股票池范围" min-width="120">
          <template #default="{ row }">{{ formatStockSelectionUniverseScope(row.universe_scope) }}</template>
        </el-table-column>
        <el-table-column prop="status" label="状态" min-width="100">
          <template #default="{ row }">{{ formatStockSelectionStatus(row.status) }}</template>
        </el-table-column>
        <el-table-column label="版本历史" min-width="260">
          <template #default="{ row }">
            <el-tag
              v-for="version in row.versions || []"
              :key="version.id"
              style="margin-right: 6px; margin-bottom: 6px"
            >
              v{{ version.version_no }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updated_at" label="更新时间" min-width="170" />
        <el-table-column label="操作" min-width="240" fixed="right">
          <template #default="{ row }">
            <div v-if="canManage" class="row-actions">
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="success" @click="setDefaultProfile(row)">设为默认</el-button>
              <el-button size="small" type="warning" @click="rollbackProfile(row)">回滚</el-button>
            </div>
            <span v-else class="muted">只读</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog
      v-model="dialogVisible"
      :title="editingID ? '编辑配置方案' : '新建配置方案'"
      width="980px"
    >
      <el-alert
        title="这里已经改成中文表单配置。你只需要按业务含义点选和填写，不需要再手动写 JSON。"
        type="info"
        :closable="false"
        style="margin-bottom: 12px"
      />

      <el-alert
        v-if="preservedExtraCount > 0"
        :title="`当前方案里还有 ${preservedExtraCount} 个高级扩展字段未在表单中单独展开，保存时会自动保留，不会丢失。`"
        type="warning"
        :closable="false"
        style="margin-bottom: 12px"
      />

      <el-form label-position="top" class="profile-form">
        <div class="section-card">
          <div class="section-head">
            <div>
              <div class="section-title">基础信息</div>
              <div class="section-desc">先定义这个配置方案的名称、状态、默认运行模式和股票池范围。</div>
            </div>
          </div>
          <div class="profile-form-grid">
            <el-form-item label="配置方案名称" required>
              <el-input v-model="form.name" placeholder="例如：稳健自动选股方案" />
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
            <el-form-item label="默认运行模式">
              <el-select v-model="form.selection_mode_default">
                <el-option
                  v-for="item in stockSelectionModeOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="绑定模板">
              <el-select v-model="form.template_id" clearable placeholder="不指定则走默认模板">
                <el-option
                  v-for="item in templates"
                  :key="item.id"
                  :label="`${item.name}（${formatStockSelectionMarketRegime(item.market_regime_bias)}）`"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="股票池范围">
              <el-select v-model="form.universe_scope" filterable allow-create default-first-option>
                <el-option
                  v-for="item in stockSelectionUniverseScopeOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="设为默认方案">
              <el-switch v-model="form.is_default" active-text="是" inactive-text="否" />
            </el-form-item>
          </div>
          <el-form-item label="方案说明">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="3"
              placeholder="补充说明这个方案适合什么风格、什么场景"
            />
          </el-form-item>
        </div>

        <div class="section-card">
          <div class="section-head">
            <div>
              <div class="section-title">股票池过滤</div>
              <div class="section-desc">控制系统先从哪些股票里筛选，以及过滤掉哪些不想参与的标的。</div>
            </div>
          </div>
          <div class="profile-form-grid">
            <el-form-item label="最少上市天数">
              <el-input-number v-model="form.universe.min_listing_days" :min="0" :step="30" controls-position="right" />
            </el-form-item>
            <el-form-item label="最近20日平均成交额下限（元）">
              <el-input-number
                v-model="form.universe.min_avg_turnover"
                :min="0"
                :step="10000000"
                controls-position="right"
              />
            </el-form-item>
            <el-form-item label="排除 ST 股票">
              <el-switch v-model="form.universe.exclude_st" active-text="排除" inactive-text="保留" />
            </el-form-item>
            <el-form-item label="排除停牌股票">
              <el-switch v-model="form.universe.exclude_suspended" active-text="排除" inactive-text="保留" />
            </el-form-item>
            <el-form-item label="最低股价（元）">
              <el-input-number v-model="form.universe.price_min" :min="0" :step="1" controls-position="right" />
            </el-form-item>
            <el-form-item label="最高股价（元）">
              <el-input-number v-model="form.universe.price_max" :min="0" :step="10" controls-position="right" />
            </el-form-item>
            <el-form-item label="最低 20 日波动率（%）">
              <el-input-number v-model="form.universe.volatility_min" :min="0" :step="0.5" controls-position="right" />
            </el-form-item>
            <el-form-item label="最高 20 日波动率（%）">
              <el-input-number v-model="form.universe.volatility_max" :min="0" :step="0.5" controls-position="right" />
            </el-form-item>
          </div>
          <el-form-item label="额外排除的股票代码">
            <el-input
              v-model="form.universe.excluded_symbols_text"
              type="textarea"
              :rows="3"
              placeholder="每行一个，示例：600519.SH"
            />
          </el-form-item>
          <div class="profile-form-grid">
            <el-form-item label="行业白名单">
              <el-input
                v-model="form.universe.industry_whitelist_text"
                type="textarea"
                :rows="3"
                placeholder="每行一个，留空表示不限"
              />
            </el-form-item>
            <el-form-item label="行业黑名单">
              <el-input
                v-model="form.universe.industry_blacklist_text"
                type="textarea"
                :rows="3"
                placeholder="每行一个，命中则排除"
              />
            </el-form-item>
            <el-form-item label="板块白名单">
              <el-input
                v-model="form.universe.sector_whitelist_text"
                type="textarea"
                :rows="3"
                placeholder="每行一个，留空表示不限"
              />
            </el-form-item>
            <el-form-item label="板块黑名单">
              <el-input
                v-model="form.universe.sector_blacklist_text"
                type="textarea"
                :rows="3"
                placeholder="每行一个，命中则排除"
              />
            </el-form-item>
            <el-form-item label="题材白名单">
              <el-input
                v-model="form.universe.theme_whitelist_text"
                type="textarea"
                :rows="3"
                placeholder="每行一个，留空表示不限"
              />
            </el-form-item>
            <el-form-item label="题材黑名单">
              <el-input
                v-model="form.universe.theme_blacklist_text"
                type="textarea"
                :rows="3"
                placeholder="每行一个，命中则排除"
              />
            </el-form-item>
          </div>
        </div>

        <div class="section-card">
          <div class="section-head">
            <div>
              <div class="section-title">种子挖掘</div>
              <div class="section-desc">控制系统从多大范围里挑出候选种子，以及是否手动指定或调试运行。</div>
            </div>
          </div>

          <el-alert
            v-if="form.selection_mode_default === 'AUTO'"
            title="当前是自动选股模式，系统会从股票池里自动挖掘种子，不需要手工输入股票代码。"
            type="success"
            :closable="false"
            style="margin-bottom: 12px"
          />
          <el-alert
            v-else-if="form.selection_mode_default === 'MANUAL'"
            title="当前是手动指定模式，请在下方填写希望参与筛选的股票代码。"
            type="warning"
            :closable="false"
            style="margin-bottom: 12px"
          />
          <el-alert
            v-else
            title="当前是调试模式，请填写调试股票代码，系统会用这些股票快速验证整条链路。"
            type="warning"
            :closable="false"
            style="margin-bottom: 12px"
          />

          <div class="profile-form-grid">
            <el-form-item label="每个信号桶保留数量">
              <el-input-number v-model="form.seed.bucket_limit" :min="1" :step="6" controls-position="right" />
            </el-form-item>
            <el-form-item label="种子池上限">
              <el-input-number v-model="form.seed.seed_pool_cap" :min="1" :step="10" controls-position="right" />
            </el-form-item>
            <el-form-item label="候选池上限">
              <el-input-number v-model="form.seed.candidate_pool_limit" :min="1" :step="5" controls-position="right" />
            </el-form-item>
            <el-form-item label="趋势桶偏置">
              <el-slider v-model="form.seed.trend_bias" :min="0.5" :max="1.5" :step="0.1" show-input />
            </el-form-item>
            <el-form-item label="资金桶偏置">
              <el-slider v-model="form.seed.money_flow_bias" :min="0.5" :max="1.5" :step="0.1" show-input />
            </el-form-item>
            <el-form-item label="质量桶偏置">
              <el-slider v-model="form.seed.quality_bias" :min="0.5" :max="1.5" :step="0.1" show-input />
            </el-form-item>
            <el-form-item label="事件桶偏置">
              <el-slider v-model="form.seed.event_bias" :min="0.5" :max="1.5" :step="0.1" show-input />
            </el-form-item>
            <el-form-item label="共振桶偏置">
              <el-slider v-model="form.seed.resonance_bias" :min="0.5" :max="1.5" :step="0.1" show-input />
            </el-form-item>
          </div>

          <el-form-item v-if="isManualMode" label="手动指定股票代码">
            <el-input
              v-model="form.seed.seed_symbols_text"
              type="textarea"
              :rows="4"
              placeholder="每行一个，示例：600519.SH"
            />
          </el-form-item>

          <el-form-item v-if="isDebugMode" label="调试股票代码">
            <el-input
              v-model="form.seed.debug_seed_symbols_text"
              type="textarea"
              :rows="4"
              placeholder="每行一个，示例：300750.SZ"
            />
          </el-form-item>
        </div>

        <div class="section-card">
          <div class="section-head section-head-inline">
            <div>
              <div class="section-title">因子权重</div>
              <div class="section-desc">这里决定量化、事件、共振和流动性修正各占多少比重。</div>
            </div>
            <el-tag :type="Math.abs(factorWeightTotal - 100) < 0.001 ? 'success' : 'danger'">
              权重合计 {{ factorWeightTotal.toFixed(0) }}%
            </el-tag>
          </div>

          <div class="profile-form-grid">
            <el-form-item label="回看交易日">
              <el-input-number v-model="form.factor.lookback_days" :min="20" :step="20" controls-position="right" />
            </el-form-item>
            <el-form-item label="量化主分权重（%）">
              <el-input-number v-model="form.factor.quant_weight_pct" :min="0" :max="100" :step="5" controls-position="right" />
            </el-form-item>
            <el-form-item label="事件确认权重（%）">
              <el-input-number v-model="form.factor.event_weight_pct" :min="0" :max="100" :step="5" controls-position="right" />
            </el-form-item>
            <el-form-item label="共振确认权重（%）">
              <el-input-number
                v-model="form.factor.resonance_weight_pct"
                :min="0"
                :max="100"
                :step="5"
                controls-position="right"
              />
            </el-form-item>
            <el-form-item label="流动性/风险修正（%）">
              <el-input-number
                v-model="form.factor.liquidity_risk_weight_pct"
                :min="0"
                :max="100"
                :step="5"
                controls-position="right"
              />
            </el-form-item>
          </div>
        </div>

        <div class="section-card">
          <div class="section-head">
            <div>
              <div class="section-title">组合规则</div>
              <div class="section-desc">控制最终选出多少只股票，以及允许的最高风险级别和最低入选分数。</div>
            </div>
          </div>
          <div class="profile-form-grid">
            <el-form-item label="最终组合数量">
              <el-input-number v-model="form.portfolio.limit" :min="1" :step="1" controls-position="right" />
            </el-form-item>
            <el-form-item label="观察名单数量">
              <el-input-number v-model="form.portfolio.watchlist_limit" :min="0" :step="1" controls-position="right" />
            </el-form-item>
            <el-form-item label="单个桶最多入选">
              <el-input-number
                v-model="form.portfolio.max_symbol_per_bucket"
                :min="1"
                :step="1"
                controls-position="right"
              />
            </el-form-item>
            <el-form-item label="单行业最多入选">
              <el-input-number
                v-model="form.portfolio.max_symbols_per_sector"
                :min="1"
                :step="1"
                controls-position="right"
              />
            </el-form-item>
            <el-form-item label="最高风险等级">
              <el-select v-model="form.portfolio.max_risk_level">
                <el-option
                  v-for="item in stockSelectionRiskLevelOptions"
                  :key="item.value"
                  :label="item.label"
                  :value="item.value"
                />
              </el-select>
            </el-form-item>
            <el-form-item label="最低入选分数">
              <el-input-number v-model="form.portfolio.min_score" :min="0" :max="100" :step="1" controls-position="right" />
            </el-form-item>
          </div>
          <el-text type="info" size="small">
            当前最高风险等级：{{ formatStockSelectionRiskLevel(form.portfolio.max_risk_level) }}
          </el-text>
        </div>

        <div class="section-card">
          <div class="section-head">
            <div>
              <div class="section-title">发布规则</div>
              <div class="section-desc">控制运行结束后是否必须审核，以及是否允许自动发布。</div>
            </div>
          </div>
          <div class="profile-form-grid">
            <el-form-item label="发布前必须人工审核">
              <el-switch v-model="form.publish.review_required" active-text="需要" inactive-text="不需要" />
            </el-form-item>
            <el-form-item label="允许自动发布">
              <el-switch v-model="form.publish.allow_auto_publish" active-text="允许" inactive-text="不允许" />
            </el-form-item>
          </div>
        </div>

        <div class="section-card">
          <div class="section-head">
            <div>
              <div class="section-title">变更说明</div>
              <div class="section-desc">写清楚这次为什么新增或修改，方便后面回滚和版本对比。</div>
            </div>
          </div>
          <el-form-item label="本次变更说明">
            <el-input
              v-model="form.change_note"
              type="textarea"
              :rows="3"
              placeholder="例如：调高流动性门槛，降低高风险标的占比"
            />
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">保存配置</el-button>
      </template>
    </el-dialog>
  </StockSelectionModuleShell>
</template>

<style scoped>
.row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.profile-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.section-card {
  border: 1px solid #e5e7eb;
  border-radius: 12px;
  padding: 16px;
  background: #f8fafc;
}

.section-head {
  margin-bottom: 12px;
}

.section-head-inline {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.section-title {
  font-size: 15px;
  font-weight: 700;
  color: #0f172a;
}

.section-desc {
  margin-top: 4px;
  font-size: 13px;
  color: #64748b;
  line-height: 1.5;
}

.profile-form-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 12px;
}
</style>
