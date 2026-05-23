import { computed, onBeforeUnmount, ref } from "vue";
import { getForecastRunDetail } from "@/api/forecast.js";

export function useForecastRunDetail(runIdRef) {
  const loading = ref(false);
  const errorMessage = ref("");
  const detail = ref(null);
  let pollTimer = null;

  const run = computed(() => detail.value?.run || null);
  const report = computed(() => detail.value?.report || null);
  const logs = computed(() => (Array.isArray(detail.value?.logs) ? detail.value.logs : []));
  const runStatus = computed(() => String(run.value?.status || "").trim().toUpperCase());
  const detailState = computed(() => {
    if (errorMessage.value) {
      if (String(errorMessage.value).toLowerCase().includes("not found")) return "not_found";
      return "failed";
    }
    if (!run.value) return "idle";
    if (runStatus.value === "QUEUED") return "queued";
    if (runStatus.value === "RUNNING") return "running";
    if (runStatus.value === "FAILED" || runStatus.value === "CANCELLED") return "failed";
    if (runStatus.value === "SUCCEEDED") return "succeeded";
    return "idle";
  });
  const statusLabel = computed(() => {
    switch (detailState.value) {
      case "not_found":
        return "不存在或已失效";
      case "queued":
        return "已进入推演队列";
      case "running":
        return "正在生成深推演报告";
      case "failed":
        return "本次深推演未完成";
      case "succeeded":
        return "深推演已完成";
      default:
        return "等待结果";
    }
  });
  const statusTone = computed(() => {
    switch (detailState.value) {
      case "queued":
        return "queued";
      case "running":
        return "running";
      case "failed":
      case "not_found":
        return "failed";
      case "succeeded":
        return "success";
      default:
        return "muted";
    }
  });
  const shouldAutoRefresh = computed(() => detailState.value === "queued" || detailState.value === "running");
  const lastUpdatedAt = computed(() => {
    return (
      run.value?.updated_at ||
      run.value?.finished_at ||
      run.value?.started_at ||
      run.value?.queued_at ||
      report.value?.updated_at ||
      ""
    );
  });
  const runMeta = computed(() => ({
    targetLabel: run.value?.target_label || "",
    targetKey: run.value?.target_key || "",
    triggerType: run.value?.trigger_type || "",
    reason: run.value?.reason || "",
    queuedAt: run.value?.queued_at || "",
    startedAt: run.value?.started_at || "",
    finishedAt: run.value?.finished_at || "",
    updatedAt: lastUpdatedAt.value,
    engineKey: run.value?.engine_key || "",
    reportRequiresVip: run.value?.report_ref?.requires_vip === true,
    reportFullReadable: run.value?.report_ref?.full_readable === true
  }));

  function stopPolling() {
    if (pollTimer) {
      clearTimeout(pollTimer);
      pollTimer = null;
    }
  }

  async function loadDetail(quiet = false) {
    const runId = typeof runIdRef?.value !== "undefined" ? runIdRef.value : runIdRef;
    if (!runId) return;
    if (!quiet) loading.value = true;
    errorMessage.value = "";
    try {
      const data = await getForecastRunDetail(runId);
      detail.value = data || null;
      const status = String(data?.run?.status || "").toUpperCase();
      if (status === "QUEUED" || status === "RUNNING") {
        stopPolling();
        pollTimer = setTimeout(() => {
          loadDetail(true);
        }, 5000);
      } else {
        stopPolling();
      }
    } catch (err) {
      errorMessage.value = err?.message || "同步失败";
      stopPolling();
    } finally {
      if (!quiet) loading.value = false;
    }
  }

  onBeforeUnmount(() => {
    stopPolling();
  });

  return {
    loading,
    errorMessage,
    detail,
    run,
    report,
    logs,
    runStatus,
    detailState,
    statusLabel,
    statusTone,
    shouldAutoRefresh,
    lastUpdatedAt,
    runMeta,
    loadDetail,
    stopPolling
  };
}
