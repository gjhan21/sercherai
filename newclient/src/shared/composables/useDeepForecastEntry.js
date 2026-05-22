import { computed } from "vue";
import { buildDeepForecastSummary } from "@/shared/lib/forecast-summary.js";

export function useDeepForecastEntry(source, options = {}) {
  const mode = options.mode === "h5" ? "h5" : "pc";
  const basePath = mode === "h5" ? "/forecast/" : "/forecast/";

  const summary = computed(() => buildDeepForecastSummary(source?.value ?? source));
  const to = computed(() => {
    const runId = summary.value?.runId;
    if (!runId) return "";
    return `${basePath}${encodeURIComponent(runId)}`;
  });
  const visible = computed(() => Boolean(summary.value));

  return {
    summary,
    to,
    visible
  };
}
