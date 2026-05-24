<template>
  <section class="h5-forecast-lab">
    <div class="h5-forecast-hero">
      <p class="h5-forecast-kicker">研究入口</p>
      <h1>深度研究中心</h1>
      <p>这里承接推荐、策略和 `run_id` 入口，先看研究聚焦和最近研究结论，再进入完整报告。</p>
      <div class="h5-forecast-actions">
        <button class="h5-primary-btn" type="button" @click="handlePrimaryAction" :disabled="loading || submittingFocusedRun">
          {{ actionBusyLabel }}
        </button>
        <button class="h5-ghost-btn" type="button" @click="$router.push('/markets')">去看行情</button>
      </div>
    </div>

    <div v-if="hasFocusedTarget" class="h5-forecast-card h5-focus-card">
      <div class="h5-forecast-head">
        <strong>当前推演对象</strong>
        <span>{{ focusCardMeta }}</span>
      </div>
      <div class="h5-focus-target">
        <div>
          <h3>{{ focusedTargetDisplayName }}</h3>
          <p>{{ focusedTargetKey || "等待绑定标的代码" }}</p>
        </div>
        <div class="h5-focus-badges">
          <span class="h5-chip">{{ currentFlowLabel }}</span>
          <span class="h5-chip muted">{{ focusTargetTypeLabel }}</span>
        </div>
      </div>
      <p>{{ focusedTargetNarrative }}</p>
      <div class="h5-focus-metrics">
        <div>
          <span>最近研究状态</span>
          <strong>{{ focusedResearchStatus }}</strong>
        </div>
        <div>
          <span>结构化报告</span>
          <strong>{{ focusedReportAvailability }}</strong>
        </div>
        <div>
          <span>是否完成模型复核</span>
          <strong>{{ focusedValidationStatus }}</strong>
        </div>
      </div>
      <div class="h5-forecast-actions">
        <button class="h5-primary-btn" type="button" @click="handleFocusedTargetAction" :disabled="submittingFocusedRun">
          {{ focusActionLabel }}
        </button>
        <button class="h5-ghost-btn" type="button" @click="toggleScope">
          {{ showAllContext ? "聚焦该标的最近运行" : "查看全部运行清单" }}
        </button>
      </div>
    </div>

    <div v-if="queryRunId" class="h5-forecast-card">
      <div class="h5-forecast-head">
        <strong>URL 指定运行</strong>
        <span>{{ queryRunId }}</span>
      </div>
      <p>当前地址带有 `run_id`，可直接进入该次深推演详情。</p>
      <RouterLink class="h5-forecast-link" :to="queryRunTo">查看完整深度推演</RouterLink>
    </div>

    <div class="h5-forecast-card">
      <div class="h5-forecast-head">
        <strong>历史研究概览</strong>
        <span>查看同标的历史对比</span>
      </div>
      <p>{{ historyOverviewNarrative }}</p>
      <div class="h5-focus-metrics">
        <div>
          <span>成功 run 数量</span>
          <strong>{{ historyViewModel.successRunCount }} 条</strong>
        </div>
        <div>
          <span>最近复盘评分趋势</span>
          <strong>{{ historyTrendLabel }}</strong>
        </div>
        <div>
          <span>查看同标的历史对比</span>
          <strong>{{ historyEntryLabel }}</strong>
        </div>
      </div>
      <div class="h5-forecast-actions compact">
        <button class="h5-ghost-btn" type="button" @click="goToFocusedHistory">查看同标的历史对比</button>
      </div>
    </div>

    <div class="h5-forecast-card">
      <div class="h5-forecast-head">
        <strong>研究聚焦</strong>
      </div>
      <p>{{ researchFocusNarrative }}</p>
      <div class="h5-focus-metrics">
        <div>
          <span>值得继续深挖</span>
          <strong>{{ researchFocusCount }} 个标的</strong>
        </div>
        <div>
          <span>最近研究结论</span>
          <strong>{{ recentResearchCount }} 条</strong>
        </div>
        <div>
          <span>运行动态</span>
          <strong>{{ activeRunCount }} 条待跟踪</strong>
        </div>
      </div>
    </div>

    <div class="h5-forecast-card">
      <div class="h5-forecast-head">
        <strong>最近研究结论</strong>
        <span>{{ displayedRecentEntries.length }} 条</span>
      </div>
      <div v-if="displayedRecentEntries.length" class="h5-forecast-list">
        <article v-for="entry in displayedRecentEntries" :key="entry.key" class="h5-forecast-item">
          <div class="h5-entry-top">
            <div>
              <p class="h5-entry-source">{{ entry.source }}</p>
              <h3>{{ entry.title }}</h3>
            </div>
            <span class="h5-entry-status" :class="`tone-${entry.summary.tone || 'muted'}`">{{ entry.summary.statusLabel }}</span>
          </div>
          <p>{{ entry.summary.summary }}</p>
          <div class="h5-focus-metrics compact">
            <div>
              <span>最近研究状态</span>
              <strong>{{ entry.summary.statusLabel }}</strong>
            </div>
            <div>
              <span>结构化报告</span>
              <strong>{{ entry.summary.reportAvailable ? "已生成" : "待生成" }}</strong>
            </div>
            <div>
              <span>模型复核</span>
              <strong>{{ entry.validationStatusLabel }}</strong>
            </div>
          </div>
          <div class="h5-forecast-actions compact">
            <RouterLink class="h5-forecast-link" :to="entry.detailTo">查看完整深度推演</RouterLink>
          </div>
        </article>
      </div>
      <div v-else class="h5-empty-state">
        <strong>{{ recentEntryEmptyTitle }}</strong>
        <p>{{ recentEntryEmptyCopy }}</p>
      </div>
    </div>

    <div class="h5-forecast-card">
      <div class="h5-forecast-head">
        <strong>运行动态</strong>
        <span>{{ displayedRuns.length }} 条</span>
      </div>
      <div v-if="displayedRuns.length" class="h5-forecast-list">
        <article v-for="run in displayedRuns" :key="run.id || run.run_id" class="h5-forecast-item">
          <div class="h5-entry-top">
            <div>
              <p class="h5-entry-source">{{ run.target_type || "STOCK" }}</p>
              <h3>{{ run.target_label || run.target_key || "未命名目标" }}</h3>
            </div>
            <span class="h5-entry-status" :class="`tone-${statusTone(run.status)}`">{{ statusLabel(run.status) }}</span>
          </div>
          <p>{{ run.summary?.executive_summary || "当前未补更多执行摘要。" }}</p>
          <RouterLink class="h5-forecast-link" :to="`/forecast/${encodeURIComponent(run.id || run.run_id)}`">查看完整深度推演</RouterLink>
        </article>
      </div>
      <div v-else-if="!loading" class="h5-empty-state">
        <strong>{{ runEmptyTitle }}</strong>
        <p>{{ runEmptyCopy }}</p>
      </div>
      <p v-if="errorMessage" class="h5-error-state">{{ errorMessage }}</p>
    </div>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { createForecastRun, getForecastRunCompare, getForecastRunHistory, getForecastRunReview, listForecastRuns } from "@/api/forecast.js";
import { listFuturesArbitrage, listStockRecommendations } from "@/api/market.js";
import { useClientAuth } from "@/shared/auth/client-auth";
import { canLaunchForecastContext, resolveForecastContextFromRoute } from "@/shared/lib/forecast-context.js";
import { buildForecastHistoryViewModel, getLatestSuccessfulForecastRun } from "@/shared/lib/forecast-history-view-model.js";
import { localizeForecastValidationStatus } from "@/shared/lib/forecast-localization.js";
import { buildDeepForecastSummary } from "@/shared/lib/forecast-summary.js";

const route = useRoute();
const router = useRouter();
const { isLoggedIn } = useClientAuth();
const loading = ref(false);
const submittingFocusedRun = ref(false);
const showAllContext = ref(false);
const errorMessage = ref("");
const stockItems = ref([]);
const futuresItems = ref([]);
const runs = ref([]);
const historyRuns = ref([]);

const routeContext = computed(() => resolveForecastContextFromRoute(route));
const queryRunId = computed(() => String(route.query.run_id || "").trim());
const queryRunTo = computed(() => `/forecast/${encodeURIComponent(queryRunId.value)}`);
const authRequired = computed(() => !isLoggedIn.value);
const flowSource = computed(() => routeContext.value.from);
const focusedTargetId = computed(() => routeContext.value.targetId);
const focusedTargetKey = computed(() => routeContext.value.targetKey);
const focusedTargetName = computed(() => routeContext.value.targetLabel);
const focusedTargetType = computed(() => routeContext.value.targetType);
const focusedSource = computed(() => routeContext.value.source);
const focusedSourceId = computed(() => routeContext.value.sourceId);
const focusedSourcePath = computed(() => routeContext.value.sourcePath);
const hasFocusedTarget = computed(() => Boolean(focusedTargetKey.value || focusedTargetName.value));
const hasStrictForecastContext = computed(() => canLaunchForecastContext(routeContext.value));
const focusedTargetDisplayName = computed(() => focusedTargetName.value || focusedTargetKey.value || "待确认标的");
const focusTargetTypeLabel = computed(() => (focusedTargetType.value === "FUTURES" ? "期货标的" : "股票标的"));
const currentFlowLabel = computed(() => {
  if (flowSource.value === "recommendations") return "来自每日推荐";
  if (flowSource.value === "strategies") return "来自交易策略";
  return "独立进入";
});
const primaryActionLabel = computed(() => (authRequired.value ? "登录后查看" : "刷新推演"));
const focusActionLabel = computed(() => {
  if (authRequired.value) return "登录后查看该标的深度推演";
  if (focusedRuns.value.length) return "查看这只标的最近深度推演";
  if (!hasStrictForecastContext.value) return "请先从来源页进入或先完成标的解析";
  return "立即发起这只标的的深度推演";
});
const actionBusyLabel = computed(() => {
  if (submittingFocusedRun.value) return "发起中...";
  if (loading.value) return "同步中...";
  return primaryActionLabel.value;
});

const recentEntries = computed(() => {
  const stockEntries = stockItems.value
    .map((item) => {
      const summary = buildDeepForecastSummary(item);
      if (!summary) return null;
      return {
        key: `stock:${item.symbol}`,
        source: "股票详情",
        title: `${item.symbol || ""} ${item.name || ""}`.trim() || "股票标的",
        targetKey: item.symbol || "",
        targetLabel: item.name || "",
        targetType: "STOCK",
        summary,
        validationStatusLabel: localizeForecastValidationStatus(item?.deep_forecast_summary?.validation_status || item?.summary?.validation_status) || "未触发模型复核",
        detailTo: `/forecast/${encodeURIComponent(summary.runId)}`
      };
    })
    .filter(Boolean);

  const futuresEntries = futuresItems.value
    .map((item) => {
      const summary = buildDeepForecastSummary(item);
      if (!summary) return null;
      return {
        key: `futures:${item.id}`,
        source: "期货机会",
        title: `${item.contract_a || ""} ${item.contract_b || ""}`.trim() || "期货机会",
        targetKey: item.contract_a || item.contract_b || "",
        targetLabel: item.contract_a || item.contract_b || "",
        targetType: "FUTURES",
        summary,
        validationStatusLabel: localizeForecastValidationStatus(item?.deep_forecast_summary?.validation_status || item?.summary?.validation_status) || "未触发模型复核",
        detailTo: `/forecast/${encodeURIComponent(summary.runId)}`
      };
    })
    .filter(Boolean);

  return [...stockEntries, ...futuresEntries].slice(0, 8);
});
const focusedRecentEntries = computed(() =>
  recentEntries.value.filter((entry) => matchesFocusedTarget(entry.targetKey, entry.targetLabel, entry.targetType))
);
const focusedRuns = computed(() =>
  runs.value.filter((run) => matchesFocusedTarget(run.target_key, run.target_label, run.target_type))
);
const displayedRecentEntries = computed(() => {
  if (!hasFocusedTarget.value || showAllContext.value) return recentEntries.value;
  return focusedRecentEntries.value;
});
const displayedRuns = computed(() => {
  if (!hasFocusedTarget.value || showAllContext.value) return runs.value;
  return focusedRuns.value;
});
const researchFocusCount = computed(() => {
  if (hasFocusedTarget.value && !showAllContext.value) return Math.max(focusedRecentEntries.value.length, focusedRuns.value.length, hasFocusedTarget.value ? 1 : 0);
  return new Set([...recentEntries.value.map((entry) => entry.targetKey || entry.targetLabel), ...runs.value.map((run) => run.target_key || run.target_label)].filter(Boolean)).size;
});
const recentResearchCount = computed(() => displayedRecentEntries.value.length);
const activeRunCount = computed(() =>
  displayedRuns.value.filter((run) => ["QUEUED", "RUNNING", "FAILED"].includes(formatStatus(run.status))).length
);
const focusCardMeta = computed(() => {
  if (!hasFocusedTarget.value) return "";
  if (showAllContext.value) return "正在查看全部运行清单";
  return focusedRuns.value.length || focusedRecentEntries.value.length ? "已切到该标的聚焦态" : "等待发起这只标的的深度推演";
});
const focusedSummary = computed(() => focusedRecentEntries.value[0]?.summary || null);
const focusedRun = computed(() => focusedRuns.value[0] || null);
const historyViewModel = computed(() =>
  buildForecastHistoryViewModel({
    targetKey: focusedTargetKey.value,
    targetLabel: focusedTargetName.value,
    runs: historyRuns.value
  })
);
const focusedResearchStatus = computed(() => {
  if (focusedSummary.value?.statusLabel) return focusedSummary.value.statusLabel;
  if (focusedRun.value?.status) return statusLabel(focusedRun.value.status);
  return "待补充";
});
const historyOverviewNarrative = computed(() => {
  if (!hasFocusedTarget.value) return "带着具体标的进入后，这里会展示同标的成功 run 历史、默认最新 vs 上一次对比，以及完整复盘评分趋势。";
  if (!historyViewModel.value.successRunCount) return `${focusedTargetDisplayName.value} 还没有成功完成的深推演历史。`;
  if (!historyViewModel.value.hasHistory) return "当前仅有 1 次成功深推演，暂不能形成历史对比";
  return historyViewModel.value.compareSummary?.conclusionChange?.summary || "已经生成最新一次 vs 上一次的历史对比。";
});
const historyTrendLabel = computed(() => {
  const scores = historyViewModel.value.reviewScores.map((item) => item.score).filter((item) => Number.isFinite(item));
  if (scores.length < 2) return "当前仅有 1 次成功深推演，暂不能形成历史对比";
  return scores[scores.length - 1] >= scores[0] ? "最近复盘评分趋势改善" : "最近复盘评分趋势走弱";
});
const historyEntryLabel = computed(() => {
  if (!historyViewModel.value.successRunCount) return "暂无历史";
  if (!historyViewModel.value.hasHistory) return "当前仅有 1 次成功深推演，暂不能形成历史对比";
  return "已可查看最新一次 vs 上一次";
});
const focusedReportAvailability = computed(() => {
  if (focusedSummary.value?.reportAvailable === true) return "已生成";
  if (focusedSummary.value?.reportAvailable === false) return "待生成";
  return hasFocusedTarget.value ? "待补充" : "-";
});
const focusedValidationStatus = computed(() => {
  if (focusedRun.value?.validation_status) return localizeForecastValidationStatus(focusedRun.value.validation_status) || "待补充";
  if (focusedRecentEntries.value[0]?.validationStatusLabel) return focusedRecentEntries.value[0].validationStatusLabel;
  return hasFocusedTarget.value ? "待补充" : "-";
});
const researchFocusNarrative = computed(() => {
  if (hasFocusedTarget.value && !showAllContext.value) {
    return `${focusedTargetDisplayName.value} 已经切到研究聚焦态。这里会优先展示它的最近研究结论、结构化报告可用性，以及仍需持续跟踪的运行动态。`;
  }
  return "这里优先展示值得继续深挖的标的、最近研究结论，以及仍需跟踪的运行动态，帮助你快速切到研究视角。";
});
const focusedTargetNarrative = computed(() => {
  if (authRequired.value) {
    return `你正从${currentFlowLabel.value.replace("来自", "")}带着 ${focusedTargetDisplayName.value} 进入深度推演。当前未登录，所以这里只先展示该标的的聚焦承接位；登录后可以直接发起或查看它的最近运行。`;
  }
  if (!hasStrictForecastContext.value) {
    return `${focusedTargetDisplayName.value} 当前只带了部分上下文，还不足以发起高质量深度推演。请先从来源页进入或先完成标的解析，补齐 target_id 与来源信息。`;
  }
  if (focusedRuns.value.length) {
    return `${focusedTargetDisplayName.value} 已经有 ${focusedRuns.value.length} 条可追踪的深推演运行。你可以先看这只标的的最近运行，必要时再切回全部运行清单。`;
  }
  if (focusedRecentEntries.value.length) {
    return `${focusedTargetDisplayName.value} 已经在来源页沉淀了可读摘要，但还没有命中的运行清单。你可以直接发起一轮新的深度推演，或者先查看最近入口。`;
  }
  return `${focusedTargetDisplayName.value} 还没有现成的深推演运行。这里已经切到该标的的聚焦态，你可以直接发起一轮新的深度推演，而不是在通用工作台里自己找。`;
});
const recentEntryEmptyTitle = computed(() => {
  if (hasFocusedTarget.value && !showAllContext.value) {
    return authRequired.value ? "登录后查看这只标的的最近入口" : "该标的暂时还没有最近入口";
  }
  return authRequired.value ? "登录后查看最近深推演入口" : "还没有可展示的最近推演";
});
const recentEntryEmptyCopy = computed(() => {
  if (hasFocusedTarget.value && !showAllContext.value) {
    return authRequired.value
      ? "当前未登录，因此不会主动请求受保护接口。登录后这里会优先展示这只标的的最近深推演入口。"
      : "这只标的暂时还没有历史入口。你可以直接发起一轮新的深度推演，或者切回全部运行清单浏览其他标的。";
  }
  return authRequired.value
    ? "当前未登录，因此不会主动请求受保护接口，也就不会触发 401。登录后这里会自动展示最近推演入口。"
    : "你可以先去股票分析或期货策略详情触发推演，生成 `deep_forecast_summary` 后这里就会自动出现。";
});
const runEmptyTitle = computed(() => {
  if (hasFocusedTarget.value && !showAllContext.value) {
    return authRequired.value ? "登录后查看这只标的的运行清单" : "该标的暂时还没有运行清单";
  }
  return authRequired.value ? "登录后查看运行清单" : "暂时还没有运行清单";
});
const runEmptyCopy = computed(() => {
  if (hasFocusedTarget.value && !showAllContext.value) {
    return authRequired.value
      ? "未登录时这里只展示这只标的的聚焦承接位，不会主动请求 `/forecast/runs`。"
      : "如果这只标的还没有现成运行，你可以直接发起新的深度推演，或者切到全部运行清单查看其它标的。";
  }
  return authRequired.value
    ? "未登录时这里只展示模块壳和直达入口，不会主动请求 `/forecast/runs`。"
    : "如果后端暂未返回运行清单，你仍然可以通过 `run_id` 或详情页入口进入推演报告。";
});

function normalizeToken(value) {
  return String(value || "").trim().toUpperCase();
}

function matchesFocusedTarget(targetKey, targetLabel, targetType) {
  if (!hasFocusedTarget.value) return true;
  const expectedType = normalizeToken(focusedTargetType.value);
  const actualType = normalizeToken(targetType);
  const expectedKey = normalizeToken(focusedTargetKey.value);
  const expectedLabel = normalizeToken(focusedTargetName.value);
  const actualKey = normalizeToken(targetKey);
  const actualLabel = normalizeToken(targetLabel);

  if (expectedType && actualType && expectedType !== actualType) return false;
  if (expectedKey && actualKey === expectedKey) return true;
  if (expectedLabel && actualLabel === expectedLabel) return true;
  return false;
}

function formatStatus(status) {
  return String(status || "").toUpperCase();
}

function statusLabel(status) {
  return (
    {
      QUEUED: "排队中",
      RUNNING: "推演中",
      SUCCEEDED: "已完成",
      FAILED: "已失败",
      CANCELLED: "已取消"
    }[formatStatus(status)] || "L3 状态"
  );
}

function statusTone(status) {
  return (
    {
      QUEUED: "queued",
      RUNNING: "running",
      SUCCEEDED: "success",
      FAILED: "failed",
      CANCELLED: "muted"
    }[formatStatus(status)] || "muted"
  );
}

async function loadLab() {
  if (authRequired.value) {
    stockItems.value = [];
    futuresItems.value = [];
    runs.value = [];
    errorMessage.value = "";
    loading.value = false;
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  const tasks = await Promise.allSettled([
    listStockRecommendations({ page: 1, page_size: 8 }),
    listFuturesArbitrage({ page: 1, page_size: 8 }),
    listForecastRuns({ page: 1, page_size: 10, target_type: route.query.target_type || "" })
  ]);

  const [stockResult, futuresResult, runResult] = tasks;

  if (stockResult.status === "fulfilled") {
    stockItems.value = Array.isArray(stockResult.value?.items) ? stockResult.value.items : [];
  }

  if (futuresResult.status === "fulfilled") {
    futuresItems.value = Array.isArray(futuresResult.value?.items) ? futuresResult.value.items : [];
  }

  if (runResult.status === "fulfilled") {
    runs.value = Array.isArray(runResult.value?.items) ? runResult.value.items : [];
  } else {
    runs.value = [];
  }

  if (tasks.every((item) => item.status === "rejected")) {
    errorMessage.value = "暂时无法同步深推演入口，请稍后重试。";
  }

  if (hasFocusedTarget.value && !authRequired.value) {
    try {
      const history = await getForecastRunHistory({
        target_type: focusedTargetType.value || route.query.target_type || "STOCK",
        target_key: focusedTargetKey.value || route.query.symbol || route.query.name || "",
        page: 1,
        page_size: 12
      });
      historyRuns.value = Array.isArray(history?.items) ? history.items : [];
      if (historyViewModel.value.hasHistory) {
        await Promise.allSettled([
          getForecastRunCompare({
            target_type: focusedTargetType.value || route.query.target_type || "STOCK",
            target_key: focusedTargetKey.value || route.query.symbol || route.query.name || ""
          }),
          focusedRuns.value[0]?.id || focusedRuns.value[0]?.run_id ? getForecastRunReview(focusedRuns.value[0]?.id || focusedRuns.value[0]?.run_id) : Promise.resolve(null)
        ]);
      }
    } catch {
      historyRuns.value = [];
    }
  } else {
    historyRuns.value = [];
  }

  loading.value = false;
}

function toggleScope() {
  showAllContext.value = !showAllContext.value;
}

async function handleFocusedTargetAction() {
  if (authRequired.value) {
    router.push("/login");
    return;
  }
  if (focusedRuns.value.length) {
    const targetRun = focusedRuns.value[0];
    router.push(`/forecast/${encodeURIComponent(targetRun.id || targetRun.run_id)}`);
    return;
  }
  if (!hasFocusedTarget.value) {
    showAllContext.value = true;
    return;
  }
  if (!hasStrictForecastContext.value) {
    errorMessage.value = "请先从来源页进入或先完成标的解析";
    return;
  }

  submittingFocusedRun.value = true;
  errorMessage.value = "";
  try {
    const created = await createForecastRun({
      target_type: focusedTargetType.value || "STOCK",
      target_id: focusedTargetId.value,
      target_key: focusedTargetKey.value,
      target_label: focusedTargetName.value || focusedTargetKey.value,
      source: focusedSource.value,
      source_id: focusedSourceId.value || focusedTargetId.value,
      source_path: focusedSourcePath.value,
      reason: `from ${flowSource.value || "forecast-lab"} focused handoff`
    });
    await loadLab();
    router.push(`/forecast/${encodeURIComponent(created.id || created.run_id)}`);
  } catch (error) {
    errorMessage.value = error?.message || "发起深度推演失败，请稍后重试。";
  } finally {
    submittingFocusedRun.value = false;
  }
}

function handlePrimaryAction() {
  if (hasFocusedTarget.value) {
    handleFocusedTargetAction();
    return;
  }
  if (authRequired.value) {
    router.push("/login");
    return;
  }
  loadLab();
}

function goToFocusedHistory() {
  const latestSuccessfulRun = getLatestSuccessfulForecastRun(historyRuns.value);
  if (latestSuccessfulRun?.id || latestSuccessfulRun?.run_id) {
    router.push(`/forecast/${encodeURIComponent(latestSuccessfulRun.id || latestSuccessfulRun.run_id)}`);
    return;
  }
  if (queryRunId.value) {
    router.push(`/forecast/${encodeURIComponent(queryRunId.value)}`);
  }
}

onMounted(loadLab);
watch(
  () => [route.query.symbol, route.query.name, route.query.target_type, route.query.target_id, route.query.source, route.query.source_path, isLoggedIn.value].join("|"),
  () => {
    showAllContext.value = false;
    loadLab();
  }
);
</script>

<style scoped>
.h5-forecast-lab { display: grid; gap: 14px; }
.h5-forecast-hero, .h5-forecast-card { padding: 16px; border-radius: var(--radius-lg); background: var(--bg-card); border: 1px solid var(--border); }
.h5-forecast-kicker { font-size: 11px; color: var(--accent-gold); text-transform: uppercase; letter-spacing: .08em; margin-bottom: 6px; }
.h5-forecast-hero h1 { font-size: 24px; font-weight: 800; margin-bottom: 8px; }
.h5-forecast-hero p, .h5-forecast-card p { color: var(--text-secondary); line-height: 1.7; }
.h5-forecast-actions { display: grid; gap: 8px; margin-top: 14px; }
.h5-forecast-actions.compact { margin-top: 10px; }
.h5-primary-btn, .h5-ghost-btn, .h5-forecast-link { width: 100%; display: inline-flex; align-items: center; justify-content: center; padding: 12px; border-radius: var(--radius-full); font-size: 14px; font-weight: 600; text-decoration: none; }
.h5-primary-btn { border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.h5-ghost-btn, .h5-forecast-link { border: 1px solid var(--accent-gold); color: var(--accent-gold); background: none; }
.h5-forecast-head { display: flex; justify-content: space-between; gap: 10px; align-items: baseline; margin-bottom: 12px; }
.h5-forecast-head span { font-size: 12px; color: var(--text-muted); font-family: var(--font-mono); }
.h5-focus-card { background: linear-gradient(135deg, rgba(240,185,11,.08), rgba(255,255,255,.02)); }
.h5-focus-target { display: grid; gap: 10px; margin-bottom: 12px; }
.h5-focus-target h3 { font-size: 18px; font-weight: 800; margin-bottom: 4px; }
.h5-focus-target p { font-family: var(--font-mono); }
.h5-focus-badges { display: flex; gap: 8px; flex-wrap: wrap; }
.h5-focus-metrics { display: grid; gap: 10px; margin-top: 12px; }
.h5-focus-metrics.compact { margin-top: 8px; }
.h5-focus-metrics div { display: grid; gap: 6px; padding: 10px; border-radius: var(--radius-sm); border: 1px solid var(--border); background: rgba(255,255,255,.02); }
.h5-focus-metrics span { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: .08em; }
.h5-focus-metrics strong { line-height: 1.6; white-space: normal; word-break: break-word; }
.h5-chip { padding: 6px 10px; border-radius: var(--radius-full); border: 1px solid var(--border-gold); color: var(--accent-gold); background: var(--accent-gold-glow); font-size: 11px; }
.h5-chip.muted { border-color: var(--border); color: var(--text-secondary); background: rgba(255,255,255,.04); }
.h5-forecast-list { display: grid; gap: 10px; }
.h5-forecast-item { display: grid; gap: 10px; padding: 14px; border-radius: var(--radius-md); border: 1px solid var(--border); background: rgba(255,255,255,.02); }
.h5-entry-top { display: flex; justify-content: space-between; gap: 12px; align-items: flex-start; }
.h5-entry-source { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: .08em; margin-bottom: 4px; }
.h5-forecast-item h3 { font-size: 16px; font-weight: 700; }
.h5-entry-status { padding: 5px 9px; border-radius: var(--radius-full); font-size: 11px; border: 1px solid transparent; }
.tone-queued { color: var(--accent-gold); background: rgba(240,185,11,.08); border-color: rgba(240,185,11,.2); }
.tone-running { color: var(--accent-blue); background: rgba(59,130,246,.08); border-color: rgba(59,130,246,.2); }
.tone-success { color: var(--positive); background: rgba(0,200,151,.08); border-color: rgba(0,200,151,.2); }
.tone-failed { color: var(--negative); background: rgba(255,90,95,.08); border-color: rgba(255,90,95,.2); }
.tone-muted { color: var(--text-secondary); background: rgba(255,255,255,.04); border-color: var(--border); }
.h5-empty-state { display: grid; gap: 8px; padding: 14px; border-radius: var(--radius-md); border: 1px dashed var(--border); }
.h5-error-state { color: var(--negative); margin-top: 8px; }
</style>
