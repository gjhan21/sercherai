import { computed } from "vue";
import {
  compareDateAsc,
  formatDate,
  formatPercent,
  trendClassByNumber,
  formatRate
} from "../utils/finance";

/**
 * Composable for strategy performance logic
 */
export function useStrategyPerformance(activeView, performanceMap, benchmarkMap, statsMap) {
  const performanceRows = computed(() => {
    const id = activeView.value?.id;
    if (!id) return [];

    const points = Array.isArray(performanceMap.value[id])
      ? [...performanceMap.value[id]]
      : [];
    const benchmarkPoints = Array.isArray(benchmarkMap.value[id])
      ? [...benchmarkMap.value[id]]
      : [];
    
    if (points.length === 0) return [];

    points.sort((a, b) => compareDateAsc(a?.date, b?.date));
    const benchmarking = new Map();
    benchmarkPoints.forEach((item) => {
      if (item?.date) {
        benchmarking.set(item.date, Number(item.return));
      }
    });

    let acc = 1;
    let benchmarkAcc = 1;
    let hasValidValue = false;

    return points.map((point) => {
      const daily = Number(point?.return);
      const benchmarkDaily = benchmarking.has(point?.date) ? Number(benchmarking.get(point?.date)) : NaN;
      
      if (Number.isFinite(daily)) {
        acc *= 1 + daily;
        hasValidValue = true;
      }
      if (Number.isFinite(benchmarkDaily)) {
        benchmarkAcc *= 1 + benchmarkDaily;
      }

      const cumulative = hasValidValue ? acc - 1 : null;
      const benchmarkCumulative = Number.isFinite(benchmarkDaily) ? benchmarkAcc - 1 : null;
      const excess = Number.isFinite(cumulative) && Number.isFinite(benchmarkCumulative)
        ? cumulative - benchmarkCumulative
        : null;

      return {
        date: formatDate(point?.date),
        dailyReturn: formatPercent(daily),
        cumulativeReturn: formatPercent(cumulative),
        benchmarkReturn: formatPercent(benchmarkCumulative),
        excessReturn: formatPercent(excess),
        dailyClass: trendClassByNumber(daily),
        cumulativeClass: trendClassByNumber(cumulative),
        benchmarkClass: trendClassByNumber(benchmarkCumulative),
        excessClass: trendClassByNumber(excess),
        dailyRaw: daily,
        cumulativeRaw: cumulative,
        benchmarkRaw: benchmarkCumulative,
        excessRaw: excess
      };
    });
  });

  const performanceSummary = computed(() => {
    const id = activeView.value?.id;
    if (id) {
      const stats = statsMap.value[id];
      if (stats) {
        const benchmarkLabel = stats.benchmark_symbol ? `基准(${stats.benchmark_symbol})` : "基准";
        return `样本 ${Number(stats.sample_days || 0)} 日 · 胜率 ${formatRate(stats.win_rate)} · 累计 ${formatPercent(
          stats.cumulative_return
        )} · ${benchmarkLabel} ${formatPercent(stats.benchmark_cumulative_return)} · 超额 ${formatPercent(
          stats.excess_return
        )} · 回撤 ${formatPercent(stats.max_drawdown)}`;
      }
    }

    const rows = performanceRows.value.filter((item) => Number.isFinite(item.dailyRaw));
    if (rows.length === 0) return "暂无统计";

    const positiveDays = rows.filter((item) => item.dailyRaw > 0).length;
    const cumulative = rows[rows.length - 1].cumulativeRaw;
    return `样本 ${rows.length} 日 · 胜率 ${formatRate(positiveDays / rows.length)} · 累计 ${formatPercent(cumulative)}`;
  });

  return {
    performanceRows,
    performanceSummary
  };
}
