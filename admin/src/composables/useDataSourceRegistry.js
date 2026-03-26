import { reactive, ref } from "vue";
import { ElMessageBox } from "element-plus";
import {
  createDataSource,
  deleteDataSource,
  listDataSources,
  listSystemConfigs,
  updateDataSource,
  upsertSystemConfig
} from "../api/admin.js";
import {
  MARKET_NEWS_DEFAULT_SOURCE_CONFIG_KEY,
  MARKET_NEWS_DEFAULT_SOURCE_FALLBACK,
  FUTURES_DEFAULT_SOURCE_CONFIG_KEY,
  FUTURES_DEFAULT_SOURCE_FALLBACK,
  STOCK_DEFAULT_SOURCE_CONFIG_KEY,
  STOCK_DEFAULT_SOURCE_FALLBACK,
  cloneConfigMap,
  defaultDataSourceForm,
  normalizeSourceKey,
  resolveBuiltinProviderKey,
  sourceTypeOptions,
  statusOptions,
  supportsDefaultStockSource,
  toSafeInt
} from "../lib/data-sources-admin.js";

const defaultDeps = {
  createDataSource,
  deleteDataSource,
  listDataSources,
  listSystemConfigs,
  updateDataSource,
  upsertSystemConfig,
  confirmDelete: (sourceKey) =>
    ElMessageBox.confirm(`确认删除数据源 ${sourceKey}？`, "删除确认", {
      type: "warning",
      confirmButtonText: "删除",
      cancelButtonText: "取消"
    })
};

function noop() {}

export function useDataSourceRegistry(options = {}, injectedDeps = {}) {
  const deps = { ...defaultDeps, ...injectedDeps };
  const feedback = options.feedback || { clear: noop, setMessage: noop, setError: noop };
  const canEditDataSources = options.canEditDataSources !== false;

  const loading = ref(false);
  const submitting = ref(false);
  const settingDefaultSource = ref(false);
  const settingDefaultSourceKey = ref("");
  const page = ref(1);
  const pageSize = ref(20);
  const total = ref(0);
  const items = ref([]);

  const formVisible = ref(false);
  const formMode = ref("create");
  const form = reactive(defaultDataSourceForm());
  const editingConfigSnapshot = ref({});

  const defaultStockSourceKey = ref(STOCK_DEFAULT_SOURCE_FALLBACK);
  const defaultFuturesSourceKey = ref(FUTURES_DEFAULT_SOURCE_FALLBACK);
  const defaultMarketNewsSourceKey = ref(MARKET_NEWS_DEFAULT_SOURCE_FALLBACK);

  function ensureCanEditDataSources() {
    if (canEditDataSources) {
      return true;
    }
    feedback.setError("当前账号只有查看权限，无法修改数据源配置或执行健康检查");
    return false;
  }

  function resetForm() {
    Object.assign(form, defaultDataSourceForm());
    formMode.value = "create";
    editingConfigSnapshot.value = {};
  }

  function closeForm() {
    formVisible.value = false;
    resetForm();
  }

  function buildPayload() {
    const sourceKey = form.source_key.trim().toUpperCase();
    const endpointValue = form.endpoint.trim();
    const builtinProviderKey = resolveBuiltinProviderKey(sourceKey);
    const config = {
      ...(formMode.value === "edit" ? cloneConfigMap(editingConfigSnapshot.value) : {}),
      endpoint: sourceKey === "TUSHARE" && !endpointValue ? "https://api.tushare.pro" : endpointValue,
      fail_threshold: toSafeInt(form.fail_threshold, 3),
      retry_times: toSafeInt(form.retry_times, 0),
      retry_interval_ms: toSafeInt(form.retry_interval_ms, 200),
      health_timeout_ms: toSafeInt(form.health_timeout_ms, 3000)
    };
    delete config.token;
    delete config.api_token;
    delete config.tushare_token;
    if (form.alert_receiver_id.trim()) {
      config.alert_receiver_id = form.alert_receiver_id.trim();
    } else {
      delete config.alert_receiver_id;
    }
    if (form.token.trim()) {
      config.token = form.token.trim();
    }
    if (builtinProviderKey && !String(config.provider || "").trim()) {
      config.provider = builtinProviderKey;
    }
    return {
      source_key: sourceKey,
      name: form.name.trim(),
      source_type: form.source_type.trim(),
      status: form.status,
      config
    };
  }

  async function loadDefaultSourceKey(configKey, fallback, targetRef, options = {}) {
    const { silent = false } = options;
    try {
      const data = await deps.listSystemConfigs({
        keyword: configKey,
        page: 1,
        page_size: 50
      });
      const rows = Array.isArray(data?.items) ? data.items : [];
      const matched = rows.find(
        (item) => normalizeSourceKey(item?.config_key) === normalizeSourceKey(configKey)
      );
      targetRef.value = normalizeSourceKey(matched?.config_value) || fallback;
    } catch (error) {
      targetRef.value = fallback;
      if (!silent) {
        throw error;
      }
    }
  }

  async function loadDefaultSourceKeys(options = {}) {
    await Promise.all([
      loadDefaultSourceKey(STOCK_DEFAULT_SOURCE_CONFIG_KEY, STOCK_DEFAULT_SOURCE_FALLBACK, defaultStockSourceKey, options),
      loadDefaultSourceKey(FUTURES_DEFAULT_SOURCE_CONFIG_KEY, FUTURES_DEFAULT_SOURCE_FALLBACK, defaultFuturesSourceKey, options),
      loadDefaultSourceKey(MARKET_NEWS_DEFAULT_SOURCE_CONFIG_KEY, MARKET_NEWS_DEFAULT_SOURCE_FALLBACK, defaultMarketNewsSourceKey, options)
    ]);
  }

  async function fetchDataSources(options = {}) {
    const { preserveFeedback = false } = options;
    loading.value = true;
    if (!preserveFeedback) {
      feedback.clear();
    }
    try {
      const [data] = await Promise.all([
        deps.listDataSources({ page: page.value, page_size: pageSize.value }),
        loadDefaultSourceKeys({ silent: true })
      ]);
      items.value = data?.items || [];
      total.value = data?.total || 0;
    } catch (error) {
      feedback.setError(error.message || "加载数据源失败");
    } finally {
      loading.value = false;
    }
  }

  async function submitForm() {
    if (!ensureCanEditDataSources()) {
      return;
    }
    submitting.value = true;
    feedback.clear();
    try {
      const payload = buildPayload();
      if (!payload.name || !payload.source_type) {
        throw new Error("请完整填写必填字段");
      }
      if (formMode.value === "create") {
        if (!payload.source_key) {
          throw new Error("source_key 不能为空");
        }
        await deps.createDataSource(payload);
        feedback.setMessage(`数据源 ${payload.source_key} 创建成功`);
      } else {
        await deps.updateDataSource(payload.source_key, {
          name: payload.name,
          source_type: payload.source_type,
          status: payload.status,
          config: payload.config
        });
        feedback.setMessage(`数据源 ${payload.source_key} 更新成功`);
      }
      closeForm();
      await fetchDataSources({ preserveFeedback: true });
    } catch (error) {
      feedback.setError(error.message || "提交失败");
    } finally {
      submitting.value = false;
    }
  }

  function handleCreate() {
    if (!ensureCanEditDataSources()) {
      return;
    }
    resetForm();
    formVisible.value = true;
  }

  function handleEdit(item) {
    if (!ensureCanEditDataSources()) {
      return;
    }
    const cfg = item?.config || {};
    editingConfigSnapshot.value = cloneConfigMap(cfg);
    Object.assign(form, {
      source_key: item?.source_key || "",
      name: item?.name || "",
      source_type: item?.source_type || "MARKET",
      status: item?.status || "ACTIVE",
      endpoint: cfg.endpoint || "",
      token: cfg.token || cfg.api_token || cfg.tushare_token || "",
      fail_threshold: toSafeInt(cfg.fail_threshold, 3),
      retry_times: toSafeInt(cfg.retry_times, 0),
      retry_interval_ms: toSafeInt(cfg.retry_interval_ms, 200),
      health_timeout_ms: toSafeInt(cfg.health_timeout_ms, 3000),
      alert_receiver_id: cfg.alert_receiver_id || "admin_001"
    });
    formMode.value = "edit";
    formVisible.value = true;
  }

  async function handleDelete(sourceKey) {
    if (!ensureCanEditDataSources()) {
      return;
    }
    try {
      await deps.confirmDelete(sourceKey);
    } catch {
      return;
    }

    feedback.clear();
    try {
      await deps.deleteDataSource(sourceKey);
      feedback.setMessage(`数据源 ${sourceKey} 已删除`);
      await fetchDataSources({ preserveFeedback: true });
    } catch (error) {
      feedback.setError(error.message || "删除失败");
    }
  }

  async function handleSetDefaultStockSource(row) {
    if (!ensureCanEditDataSources()) {
      return;
    }
    if (!supportsDefaultStockSource(row)) {
      feedback.setError("仅支持设置可用于股票行情同步的数据源为默认行情源");
      return;
    }
    const sourceKey = normalizeSourceKey(row?.source_key);
    if (!sourceKey) {
      feedback.setError("source_key 不能为空");
      return;
    }
    settingDefaultSource.value = true;
    settingDefaultSourceKey.value = sourceKey;
    feedback.clear();
    try {
      await deps.upsertSystemConfig({
        config_key: STOCK_DEFAULT_SOURCE_CONFIG_KEY,
        config_value: sourceKey,
        description: "股票行情默认数据源"
      });
      defaultStockSourceKey.value = sourceKey;
      feedback.setMessage(`默认行情源已设置为 ${sourceKey}`);
    } catch (error) {
      feedback.setError(error.message || "设置默认行情源失败");
    } finally {
      settingDefaultSource.value = false;
      settingDefaultSourceKey.value = "";
    }
  }

  function handlePageChange(nextPage) {
    if (nextPage === page.value) {
      return;
    }
    page.value = nextPage;
    fetchDataSources();
  }

  return {
    loading,
    submitting,
    settingDefaultSource,
    settingDefaultSourceKey,
    page,
    pageSize,
    total,
    items,
    formVisible,
    formMode,
    form,
    sourceTypeOptions,
    statusOptions,
    defaultStockSourceKey,
    defaultFuturesSourceKey,
    defaultMarketNewsSourceKey,
    fetchDataSources,
    submitForm,
    handleCreate,
    handleEdit,
    handleDelete,
    handleSetDefaultStockSource,
    handlePageChange,
    closeForm,
    resetForm,
    supportsDefaultStockSource
  };
}
