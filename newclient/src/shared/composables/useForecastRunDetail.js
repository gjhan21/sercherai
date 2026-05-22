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
    loadDetail,
    stopPolling
  };
}
