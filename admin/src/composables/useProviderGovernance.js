import { computed, reactive, ref } from "vue";
import {
  getMarketProviderGovernanceOverview,
  listMarketProviderCapabilities,
  listMarketProviderRoutingPolicies,
  updateMarketProviderRoutingPolicy
} from "../api/admin.js";
import {
  defaultRoutingPolicyForm,
  formatGovernanceSuggestion,
  normalizeProviderKeyList,
  normalizeSourceKey
} from "../lib/data-sources-admin.js";

const defaultDeps = {
  getMarketProviderGovernanceOverview,
  listMarketProviderCapabilities,
  listMarketProviderRoutingPolicies,
  updateMarketProviderRoutingPolicy
};

function noop() {}

export function useProviderGovernance(options = {}, injectedDeps = {}) {
  const deps = { ...defaultDeps, ...injectedDeps };
  const feedback = options.feedback || { clear: noop, setMessage: noop, setError: noop };
  const canEditDataSources = options.canEditDataSources !== false;
  const getLookbackHours = options.getLookbackHours || (() => 24);

  const governanceOverviewStock = ref(null);
  const governanceOverviewFutures = ref(null);
  const providerCapabilities = ref([]);
  const routingPolicies = ref([]);
  const governanceLoading = ref(false);
  const routingPolicyDialogVisible = ref(false);
  const routingPolicySubmitting = ref(false);
  const editingRoutingPolicyKey = ref("");
  const routingPolicyForm = reactive(defaultRoutingPolicyForm());

  const governanceOverviewCards = computed(() =>
    [governanceOverviewStock.value, governanceOverviewFutures.value].filter(Boolean)
  );

  const governanceProviderScoreRows = computed(() =>
    governanceOverviewCards.value.flatMap((overview) =>
      (overview?.provider_scores || []).map((item) => ({
        ...item,
        asset_class: item.asset_class || overview.asset_class || ""
      }))
    )
  );

  const routingPolicyProviderOptions = computed(() => {
    const normalizedAssetClass = normalizeSourceKey(routingPolicyForm.asset_class);
    const normalizedDataKind = normalizeSourceKey(routingPolicyForm.data_kind);
    const values = new Set();
    providerCapabilities.value.forEach((item) => {
      if (normalizedAssetClass && normalizeSourceKey(item?.asset_class) !== normalizedAssetClass) {
        return;
      }
      if (normalizedDataKind && normalizeSourceKey(item?.data_kind) !== normalizedDataKind) {
        return;
      }
      const providerKey = normalizeSourceKey(item?.provider_key);
      if (providerKey) {
        values.add(providerKey);
      }
    });
    return Array.from(values).map((value) => ({ value, label: value }));
  });

  function ensureCanEditDataSources() {
    if (canEditDataSources) {
      return true;
    }
    feedback.setError("当前账号只有查看权限，无法修改数据源配置或执行健康检查");
    return false;
  }

  function resetRoutingPolicyForm() {
    Object.assign(routingPolicyForm, defaultRoutingPolicyForm());
    editingRoutingPolicyKey.value = "";
  }

  function closeRoutingPolicyDialog() {
    routingPolicyDialogVisible.value = false;
    resetRoutingPolicyForm();
  }

  async function fetchGovernanceDashboard(options = {}) {
    const { preserveFeedback = false } = options;
    governanceLoading.value = true;
    if (!preserveFeedback) {
      feedback.clear();
    }
    try {
      const hours = getLookbackHours();
      const [stockOverview, futuresOverview, capabilityResult, policyResult] = await Promise.all([
        deps.getMarketProviderGovernanceOverview({ asset_class: "STOCK", data_kind: "DAILY_BARS", hours }),
        deps.getMarketProviderGovernanceOverview({ asset_class: "FUTURES", data_kind: "DAILY_BARS", hours }),
        deps.listMarketProviderCapabilities({ data_kind: "DAILY_BARS" }),
        deps.listMarketProviderRoutingPolicies({})
      ]);
      governanceOverviewStock.value = stockOverview || null;
      governanceOverviewFutures.value = futuresOverview || null;
      providerCapabilities.value = capabilityResult?.items || [];
      routingPolicies.value = policyResult?.items || [];
    } catch (error) {
      governanceOverviewStock.value = null;
      governanceOverviewFutures.value = null;
      providerCapabilities.value = [];
      routingPolicies.value = [];
      if (!preserveFeedback) {
        feedback.setError(error.message || "加载供应商治理信息失败");
      }
    } finally {
      governanceLoading.value = false;
    }
  }

  function handleEditRoutingPolicy(item) {
    if (!ensureCanEditDataSources()) {
      return;
    }
    Object.assign(routingPolicyForm, {
      policy_key: item?.policy_key || "",
      asset_class: normalizeSourceKey(item?.asset_class),
      data_kind: normalizeSourceKey(item?.data_kind),
      primary_provider_key: normalizeSourceKey(item?.primary_provider_key),
      fallback_provider_keys: normalizeProviderKeyList(item?.fallback_provider_keys),
      fallback_allowed: item?.fallback_allowed !== false,
      mock_allowed: item?.mock_allowed === true,
      quality_threshold: Number(item?.quality_threshold) || 0
    });
    editingRoutingPolicyKey.value = routingPolicyForm.policy_key;
    routingPolicyDialogVisible.value = true;
  }

  async function handleSaveRoutingPolicy() {
    if (!ensureCanEditDataSources()) {
      return;
    }
    if (!routingPolicyForm.policy_key || !routingPolicyForm.data_kind || !routingPolicyForm.primary_provider_key) {
      feedback.setError("请先完整填写策略 key、数据域和主源");
      return;
    }
    routingPolicySubmitting.value = true;
    feedback.clear();
    try {
      await deps.updateMarketProviderRoutingPolicy(routingPolicyForm.policy_key, {
        asset_class: routingPolicyForm.asset_class,
        data_kind: routingPolicyForm.data_kind,
        primary_provider_key: routingPolicyForm.primary_provider_key,
        fallback_provider_keys: normalizeProviderKeyList(routingPolicyForm.fallback_provider_keys),
        fallback_allowed: routingPolicyForm.fallback_allowed,
        mock_allowed: routingPolicyForm.mock_allowed,
        quality_threshold: Number(routingPolicyForm.quality_threshold) || 0
      });
      routingPolicyDialogVisible.value = false;
      feedback.setMessage(`路由策略 ${routingPolicyForm.policy_key} 已更新`);
      resetRoutingPolicyForm();
      await fetchGovernanceDashboard({ preserveFeedback: true });
    } catch (error) {
      feedback.setError(error.message || "保存路由策略失败");
    } finally {
      routingPolicySubmitting.value = false;
    }
  }

  return {
    governanceOverviewStock,
    governanceOverviewFutures,
    providerCapabilities,
    routingPolicies,
    governanceLoading,
    routingPolicyDialogVisible,
    routingPolicySubmitting,
    editingRoutingPolicyKey,
    routingPolicyForm,
    governanceProviderScoreRows,
    routingPolicyProviderOptions,
    fetchGovernanceDashboard,
    handleEditRoutingPolicy,
    handleSaveRoutingPolicy,
    closeRoutingPolicyDialog,
    resetRoutingPolicyForm,
    formatGovernanceSuggestion
  };
}
