<template>
  <section class="forecast-lab-page">
    <div class="forecast-lab-hero">
      <div class="hero-copy">
        <p class="forecast-lab-kicker">研究入口</p>
        <h1>深度研究中心</h1>
        <p class="forecast-lab-subtitle">
          汇总值得继续深挖的股票与期货，先看研究聚焦和最近研究结论，再进入完整深推演报告。
        </p>
        <div class="journey-context">
          <span class="journey-chip">{{ currentFlowLabel }}</span>
          <span v-if="contextTargetLabel" class="journey-target">当前标的：{{ contextTargetLabel }}</span>
        </div>
      </div>
      <div class="hero-actions">
        <button class="btn-primary" type="button" @click="handlePrimaryAction" :disabled="loading || submittingFocusedRun">
          {{ actionBusy ? actionBusyLabel : primaryActionLabel }}
        </button>
        <button class="btn-ghost" type="button" @click="goToPreviousStep">{{ previousStepLabel }}</button>
      </div>
    </div>

    <section v-if="hasFocusedTarget" class="forecast-lab-card focus-card">
      <div class="card-head">
        <strong>当前推演对象</strong>
        <span class="card-meta">{{ focusCardMeta }}</span>
      </div>
      <div class="focus-target">
        <div>
          <h3>{{ focusedTargetDisplayName }}</h3>
          <p>{{ focusedTargetKey || "等待绑定标的代码" }}</p>
        </div>
        <div class="focus-badges">
          <span class="journey-chip">{{ currentFlowLabel }}</span>
          <span class="journey-target">{{ focusTargetTypeLabel }}</span>
        </div>
      </div>
      <p class="card-copy">{{ focusedTargetNarrative }}</p>
      <div class="focus-metrics">
        <article>
          <span>最近研究状态</span>
          <strong>{{ focusedResearchStatus }}</strong>
        </article>
        <article>
          <span>结构化报告</span>
          <strong>{{ focusedReportAvailability }}</strong>
        </article>
        <article>
          <span>是否完成模型复核</span>
          <strong>{{ focusedValidationStatus }}</strong>
        </article>
      </div>
      <div class="entry-actions">
        <button class="btn-primary" type="button" @click="handleFocusedTargetAction" :disabled="submittingFocusedRun">
          {{ focusActionLabel }}
        </button>
        <button class="inline-link" type="button" @click="toggleScope">
          {{ showAllContext ? "聚焦该标的最近运行" : "查看全部运行清单" }}
        </button>
      </div>
    </section>

    <section class="forecast-lab-card">
      <div class="card-head">
        <strong>当前主线</strong>
      </div>
      <div class="journey-steps">
        <span class="journey-step" :class="{ active: flowSource === 'recommendations' }">1. 每日推荐</span>
        <span class="journey-step" :class="{ active: flowSource === 'strategies' }">2. 交易策略</span>
        <span class="journey-step active">3. 深度推演</span>
      </div>
      <p class="card-copy">{{ flowNarrative }}</p>
      <div class="entry-actions">
        <button class="inline-link" type="button" @click="goToPreviousStep">{{ previousStepLabel }}</button>
        <button class="card-link" type="button" @click="goToAnalysis">去股票分析补充判断</button>
      </div>
    </section>

    <section v-if="queryRunId" class="forecast-lab-card">
      <div class="card-head">
        <strong>URL 指定运行</strong>
        <span class="card-meta">{{ queryRunId }}</span>
      </div>
      <p class="card-copy">当前地址带有 `run_id`，可直接进入该次深推演详情。</p>
      <RouterLink class="card-link" :to="queryRunTo">查看完整深度推演</RouterLink>
    </section>

    <section class="forecast-lab-card">
      <div class="card-head">
        <strong>历史研究概览</strong>
        <span class="card-meta">查看同标的历史对比</span>
      </div>
      <p class="card-copy">{{ historyOverviewNarrative }}</p>
      <div class="focus-metrics">
        <article>
          <span>成功 run 数量</span>
          <strong>{{ historyViewModel.successRunCount }} 条</strong>
        </article>
        <article>
          <span>最近复盘评分趋势</span>
          <strong>{{ historyTrendLabel }}</strong>
        </article>
        <article>
          <span>查看同标的历史对比</span>
          <strong>{{ historyEntryLabel }}</strong>
        </article>
      </div>
      <div class="entry-actions">
        <button class="card-link" type="button" @click="goToFocusedHistory">查看同标的历史对比</button>
      </div>
    </section>

    <section class="forecast-lab-card">
      <div class="card-head">
        <strong>研究聚焦</strong>
      </div>
      <p class="card-copy">{{ researchFocusNarrative }}</p>
      <div class="focus-metrics">
        <article>
          <span>值得继续深挖</span>
          <strong>{{ researchFocusCount }} 个标的</strong>
        </article>
        <article>
          <span>最近研究结论</span>
          <strong>{{ recentResearchCount }} 条</strong>
        </article>
        <article>
          <span>运行动态</span>
          <strong>{{ activeRunCount }} 条待跟踪</strong>
        </article>
      </div>
    </section>

    <section class="forecast-lab-card">
      <div class="card-head">
        <strong>最近研究结论</strong>
        <span class="card-meta">{{ displayedRecentEntries.length }} 条</span>
      </div>
      <div v-if="displayedRecentEntries.length" class="entry-list">
        <article v-for="entry in displayedRecentEntries" :key="entry.key" class="entry-item">
          <div class="entry-top">
            <div>
              <p class="entry-source">{{ entry.source }}</p>
              <h3>{{ entry.title }}</h3>
            </div>
            <span class="entry-status" :class="`tone-${entry.summary.tone || 'muted'}`">{{ entry.summary.statusLabel }}</span>
          </div>
          <p class="entry-summary">{{ entry.summary.summary }}</p>
          <div class="entry-metrics">
            <span>最近研究状态：{{ entry.summary.statusLabel }}</span>
            <span>结构化报告：{{ entry.summary.reportAvailable ? "已生成" : "待生成" }}</span>
            <span>模型复核：{{ entry.validationStatusLabel }}</span>
          </div>
          <div v-if="entry.summary.note" class="entry-note">{{ entry.summary.note }}</div>
          <div class="entry-actions">
            <RouterLink class="card-link" :to="entry.detailTo">查看完整深度推演</RouterLink>
            <button class="inline-link" type="button" @click="goToHost(entry.hostTo)">回到来源页面</button>
          </div>
        </article>
      </div>
      <div v-else class="empty-state">
        <strong>{{ recentEntryEmptyTitle }}</strong>
        <p>{{ recentEntryEmptyCopy }}</p>
      </div>
    </section>

    <section class="forecast-lab-card">
      <div class="card-head">
        <strong>运行动态</strong>
        <span class="card-meta">{{ displayedRuns.length }} 条</span>
      </div>
      <div v-if="displayedRuns.length" class="run-list">
        <article v-for="run in displayedRuns" :key="run.id || run.run_id" class="run-item">
          <div class="run-top">
            <div>
              <p class="run-target">{{ run.target_label || run.target_key || "未命名目标" }}</p>
              <h3>{{ run.target_type || "STOCK" }}</h3>
            </div>
            <span class="entry-status" :class="`tone-${statusTone(run.status)}`">{{ statusLabel(run.status) }}</span>
          </div>
          <p class="entry-summary">{{ run.summary?.executive_summary || "当前未补更多执行摘要。" }}</p>
          <div class="entry-actions">
            <RouterLink class="card-link" :to="`/forecast/${encodeURIComponent(run.id || run.run_id)}`">查看完整深度推演</RouterLink>
          </div>
        </article>
      </div>
      <div v-else-if="!loading" class="empty-state">
        <strong>{{ runEmptyTitle }}</strong>
        <p>{{ runEmptyCopy }}</p>
      </div>
      <p v-if="errorMessage" class="error-state">{{ errorMessage }}</p>
    </section>
  </section>
</template>

<script setup>
import { computed, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { listStockRecommendations, listFuturesStrategies } from "@/api/market.js";
import { createForecastRun, getForecastRunCompare, getForecastRunHistory, getForecastRunReview, listForecastRuns } from "@/api/forecast.js";
import { useClientAuth } from "@/shared/auth/client-auth";
import { canLaunchForecastContext, resolveForecastContextFromRoute } from "@/shared/lib/forecast-context.js";
import { buildForecastHistoryViewModel, getLatestSuccessfulForecastRun } from "@/shared/lib/forecast-history-view-model.js";
import { buildDeepForecastSummary } from "@/shared/lib/forecast-summary.js";
import { localizeForecastValidationStatus } from "@/shared/lib/forecast-localization.js";

const route = useRoute();
const router = useRouter();
const { isLoggedIn } = useClientAuth();
const loading = ref(false);
const errorMessage = ref("");
const stockItems = ref([]);
const strategyItems = ref([]);
const runs = ref([]);
const historyRuns = ref([]);
const showAllContext = ref(false);
const submittingFocusedRun = ref(false);

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
const contextTargetLabel = computed(() => String(route.query.name || route.query.symbol || "").trim());
const focusTargetTypeLabel = computed(() => (focusedTargetType.value === "FUTURES" ? "期货标的" : "股票标的"));
const focusedTargetDisplayName = computed(() => focusedTargetName.value || focusedTargetKey.value || "待确认标的");
const focusActionLabel = computed(() => {
  if (authRequired.value) return "登录后查看该标的深度推演";
  if (focusedRuns.value.length) return "查看这只标的最近深度推演";
  if (!hasStrictForecastContext.value) return "请先从来源页进入或先完成标的解析";
  return "立即发起这只标的的深度推演";
});
const primaryActionLabel = computed(() => {
  if (hasFocusedTarget.value) return focusActionLabel.value;
  return authRequired.value ? "登录后查看" : "刷新推演";
});
const actionBusy = computed(() => loading.value || submittingFocusedRun.value);
const actionBusyLabel = computed(() => (submittingFocusedRun.value ? "发起中..." : "同步中..."));
const currentFlowLabel = computed(() => {
  if (flowSource.value === "recommendations") return "来自每日推荐";
  if (flowSource.value === "strategies") return "来自交易策略";
  return "独立进入";
});
const previousStepLabel = computed(() => {
  if (flowSource.value === "strategies") return "回到交易策略";
  if (flowSource.value === "recommendations") return "回到每日推荐";
  return "去股票分析";
});
const flowNarrative = computed(() => {
  if (flowSource.value === "recommendations") {
    return `你现在处在第 3 步：深度推演。前一步来自每日推荐${contextTargetLabel.value ? `，当前关注 ${contextTargetLabel.value}` : ""}，建议继续查看运行结果或回到策略页补充执行方案。`;
  }
  if (flowSource.value === "strategies") {
    return `你现在处在第 3 步：深度推演。前一步来自交易策略${contextTargetLabel.value ? `，当前关注 ${contextTargetLabel.value}` : ""}，可以继续查看运行详情，或回到策略页复核方向与仓位。`;
  }
  return "你现在处在第 3 步：深度推演。这里负责承接推荐与策略之后的 L3 深挖、运行状态与完整报告。";
});
const focusCardMeta = computed(() => {
  if (!hasFocusedTarget.value) return "";
  if (showAllContext.value) return "正在查看全部运行清单";
  return focusedRuns.value.length || focusedRecentEntries.value.length ? "已切到该标的聚焦态" : "等待发起这只标的的深度推演";
});

const recentEntries = computed(() => {
  const stockEntries = stockItems.value
    .map((item) => {
      const summary = buildDeepForecastSummary(item);
      if (!summary) return null;
      return {
        key: `stock:${item.symbol}`,
        source: "股票分析",
        title: `${item.symbol || ""} ${item.name || ""}`.trim() || "股票标的",
        targetKey: item.symbol || "",
        targetLabel: item.name || "",
        targetType: "STOCK",
        summary,
        validationStatusLabel: localizeForecastValidationStatus(item?.deep_forecast_summary?.validation_status || item?.summary?.validation_status) || "未触发模型复核",
        detailTo: `/forecast/${encodeURIComponent(summary.runId)}`,
        hostTo: { path: `/identify/${item.symbol}` }
      };
    })
    .filter(Boolean);

  const strategyEntries = strategyItems.value
    .map((item) => {
      const summary = buildDeepForecastSummary(item.explanation || item);
      if (!summary) return null;
      return {
        key: `strategy:${item.id}`,
        source: "期货策略",
        title: item.contract || item.name || "期货策略",
        targetKey: item.contract || item.name || "",
        targetLabel: item.name || item.contract || "",
        targetType: "FUTURES",
        summary,
        validationStatusLabel: localizeForecastValidationStatus(item?.explanation?.deep_forecast_summary?.validation_status || item?.explanation?.summary?.validation_status) || "未触发模型复核",
        detailTo: `/forecast/${encodeURIComponent(summary.runId)}`,
        hostTo: { path: `/futures/strategy/${item.id}` }
      };
    })
    .filter(Boolean);

  return [...stockEntries, ...strategyEntries].slice(0, 8);
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
  return "这里优先展示值得继续深挖的标的、最近研究结论，以及仍需跟踪的运行动态，帮助你从推荐或策略快速进入研究视角。";
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
    : "如果后端暂未返回 `/forecast/runs` 列表，本页仍可通过最近入口和 `run_id` 直达详情。";
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
    strategyItems.value = [];
    runs.value = [];
    errorMessage.value = "";
    loading.value = false;
    return;
  }

  loading.value = true;
  errorMessage.value = "";
  const tasks = await Promise.allSettled([
    listStockRecommendations({ page: 1, page_size: 12 }),
    listFuturesStrategies({ page: 1, page_size: 8 }),
    listForecastRuns({ page: 1, page_size: 10, target_type: route.query.target_type || "" })
  ]);

  const [stockResult, strategyResult, runResult] = tasks;

  if (stockResult.status === "fulfilled") {
    stockItems.value = Array.isArray(stockResult.value?.items) ? stockResult.value.items : [];
  }

  if (strategyResult.status === "fulfilled") {
    strategyItems.value = Array.isArray(strategyResult.value?.items) ? strategyResult.value.items : [];
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

function goToAnalysis() {
  router.push("/identify");
}

function toggleScope() {
  showAllContext.value = !showAllContext.value;
}

function goToPreviousStep() {
  if (flowSource.value === "strategies") {
    router.push({
      path: route.query.target_type === "FUTURES" ? "/futures/strategies" : "/recommendations/strategies",
      query: {
        ...(route.query.target_id ? { reco_id: route.query.target_id } : {}),
        ...(route.query.symbol ? { symbol: route.query.symbol } : {}),
        ...(route.query.name ? { name: route.query.name } : {}),
        from: "forecast-lab"
      }
    });
    return;
  }

  if (flowSource.value === "recommendations") {
    router.push({
      path: "/recommendations",
      query: {
        ...(route.query.symbol ? { symbol: route.query.symbol } : {}),
        ...(route.query.name ? { name: route.query.name } : {}),
        from: "forecast-lab"
      }
    });
    return;
  }

  goToAnalysis();
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

function goToHost(location) {
  router.push(location);
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
.forecast-lab-page { display: grid; gap: 16px; max-width: 1180px; }
.forecast-lab-hero, .forecast-lab-card { background: var(--bg-card); border: 1px solid var(--border); border-radius: var(--radius-lg); padding: 24px; }
.forecast-lab-hero { display: flex; justify-content: space-between; align-items: flex-start; gap: 18px; }
.forecast-lab-kicker { font-size: 12px; color: var(--accent-gold); text-transform: uppercase; letter-spacing: .08em; margin-bottom: 6px; }
.forecast-lab-hero h1 { font-size: 28px; font-weight: 800; margin-bottom: 8px; }
.forecast-lab-subtitle { color: var(--text-secondary); line-height: 1.8; max-width: 760px; }
.journey-context { display: flex; gap: 8px; flex-wrap: wrap; margin-top: 12px; }
.journey-chip, .journey-target, .journey-step { padding: 7px 11px; border-radius: var(--radius-full); border: 1px solid var(--border); font-size: 12px; color: var(--text-secondary); }
.journey-chip { border-color: var(--border-gold); color: var(--accent-gold); background: var(--accent-gold-glow); }
.journey-steps { display: flex; gap: 8px; flex-wrap: wrap; margin-bottom: 12px; }
.journey-step.active { border-color: var(--border-gold); color: var(--accent-gold); background: var(--accent-gold-glow); }
.focus-card { background: linear-gradient(135deg, rgba(240,185,11,.08), rgba(255,255,255,.02)); }
.focus-target { display: flex; justify-content: space-between; gap: 18px; align-items: center; margin-bottom: 14px; }
.focus-target h3 { font-size: 22px; font-weight: 800; margin-bottom: 6px; }
.focus-target p { color: var(--text-secondary); font-family: var(--font-mono); }
.focus-badges { display: flex; gap: 8px; flex-wrap: wrap; }
.focus-metrics { display: grid; grid-template-columns: repeat(3,1fr); gap: 12px; margin-top: 14px; }
.focus-metrics article { display: grid; gap: 8px; padding: 14px; border-radius: var(--radius-md); background: rgba(255,255,255,.02); border: 1px solid var(--border); }
.focus-metrics span, .entry-metrics span { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: .08em; }
.hero-actions { display: flex; gap: 10px; flex-wrap: wrap; }
.btn-primary, .btn-ghost, .card-link, .inline-link { display: inline-flex; align-items: center; justify-content: center; border-radius: var(--radius-full); font-size: 13px; font-weight: 600; cursor: pointer; text-decoration: none; }
.btn-primary { padding: 10px 18px; border: none; background: linear-gradient(135deg,var(--accent-gold),var(--accent-gold-dim)); color: #000; }
.btn-ghost, .card-link, .inline-link { padding: 10px 16px; border: 1px solid var(--border-gold); background: none; color: var(--accent-gold); }
.card-head { display: flex; justify-content: space-between; gap: 12px; align-items: baseline; margin-bottom: 12px; }
.card-copy, .entry-summary, .entry-note, .empty-state p, .error-state { color: var(--text-secondary); line-height: 1.7; }
.card-meta { font-size: 12px; color: var(--text-muted); font-family: var(--font-mono); }
.entry-list, .run-list { display: grid; gap: 12px; }
.entry-item, .run-item { padding: 16px; border-radius: var(--radius-md); background: rgba(255,255,255,.02); border: 1px solid var(--border); }
.entry-top, .run-top { display: flex; justify-content: space-between; gap: 14px; align-items: flex-start; margin-bottom: 8px; }
.entry-source, .run-target { font-size: 11px; color: var(--text-muted); text-transform: uppercase; letter-spacing: .08em; margin-bottom: 6px; }
.entry-item h3, .run-item h3 { font-size: 18px; font-weight: 700; }
.entry-status { padding: 6px 10px; border-radius: var(--radius-full); font-size: 12px; border: 1px solid transparent; }
.tone-queued { color: var(--accent-gold); background: rgba(240,185,11,.08); border-color: rgba(240,185,11,.2); }
.tone-running { color: var(--accent-blue); background: rgba(59,130,246,.08); border-color: rgba(59,130,246,.2); }
.tone-success { color: var(--positive); background: rgba(0,200,151,.08); border-color: rgba(0,200,151,.2); }
.tone-failed { color: var(--negative); background: rgba(255,90,95,.08); border-color: rgba(255,90,95,.2); }
.tone-muted { color: var(--text-secondary); background: rgba(255,255,255,.04); border-color: var(--border); }
.entry-note { margin-top: 6px; }
.entry-metrics { display: grid; gap: 6px; margin-top: 8px; }
.entry-actions { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 14px; }
.inline-link { border-color: var(--border); color: var(--text-secondary); }
.empty-state { display: grid; gap: 8px; padding: 20px; border-radius: var(--radius-md); border: 1px dashed var(--border); background: rgba(255,255,255,.015); }
.empty-state strong { font-size: 15px; }
.error-state { margin-top: 12px; color: var(--negative); }
@media (max-width: 900px) {
  .forecast-lab-hero, .focus-target, .focus-metrics { display: grid; }
}
</style>
