import { ref } from "vue";
import {
  batchCheckDataSources,
  checkDataSourceHealth,
  listDataSourceHealthLogs
} from "../api/admin.js";

const defaultDeps = {
  batchCheckDataSources,
  checkDataSourceHealth,
  listDataSourceHealthLogs
};

function noop() {}

export function useDataSourceHealth(options = {}, injectedDeps = {}) {
  const deps = { ...defaultDeps, ...injectedDeps };
  const feedback = options.feedback || { clear: noop, setMessage: noop, setError: noop };
  const canEditDataSources = options.canEditDataSources !== false;

  const batchChecking = ref(false);
  const healthMap = ref({});
  const logsLoading = ref(false);
  const logSourceKey = ref("");
  const logItems = ref([]);

  function ensureCanEditDataSources() {
    if (canEditDataSources) {
      return true;
    }
    feedback.setError("当前账号只有查看权限，无法修改数据源配置或执行健康检查");
    return false;
  }

  async function handleCheckOne(sourceKey) {
    if (!ensureCanEditDataSources()) {
      return;
    }
    feedback.clear();
    try {
      const result = await deps.checkDataSourceHealth(sourceKey);
      healthMap.value = {
        ...healthMap.value,
        [sourceKey]: result
      };
      feedback.setMessage(`数据源 ${sourceKey} 健康检查完成`);
    } catch (error) {
      feedback.setError(error.message || "健康检查失败");
    }
  }

  async function handleBatchCheckAll() {
    if (!ensureCanEditDataSources()) {
      return;
    }
    batchChecking.value = true;
    feedback.clear();
    try {
      const result = await deps.batchCheckDataSources([]);
      const merged = { ...healthMap.value };
      const rows = result?.items || [];
      rows.forEach((row) => {
        merged[row.source_key] = row;
      });
      healthMap.value = merged;
      feedback.setMessage(`批量健康检查完成，共 ${rows.length} 个数据源`);
    } catch (error) {
      feedback.setError(error.message || "批量健康检查失败");
    } finally {
      batchChecking.value = false;
    }
  }

  async function showLogs(sourceKey) {
    logsLoading.value = true;
    logSourceKey.value = sourceKey;
    feedback.clear();
    try {
      const result = await deps.listDataSourceHealthLogs(sourceKey, {
        page: 1,
        page_size: 20
      });
      logItems.value = result?.items || [];
    } catch (error) {
      logItems.value = [];
      feedback.setError(error.message || "加载健康日志失败");
    } finally {
      logsLoading.value = false;
    }
  }

  return {
    batchChecking,
    healthMap,
    logsLoading,
    logSourceKey,
    logItems,
    handleCheckOne,
    handleBatchCheckAll,
    showLogs
  };
}
