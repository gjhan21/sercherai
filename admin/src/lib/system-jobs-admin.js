function normalizeJobName(value) {
  return String(value || "").trim();
}

function normalizeModule(value) {
  return String(value || "").trim().toUpperCase();
}

function normalizeDisplayName(value, fallback) {
  return String(value || "").trim() || String(fallback || "").trim();
}

function compareDefinitionOptions(left, right) {
  const moduleCompare = String(left?.module || "").localeCompare(String(right?.module || ""), "zh-Hans-CN");
  if (moduleCompare !== 0) {
    return moduleCompare;
  }
  return String(left?.job_name || "").localeCompare(String(right?.job_name || ""), "zh-Hans-CN");
}

function collectDefinitionJobNameMap(definitions = []) {
  const map = new Map();
  (Array.isArray(definitions) ? definitions : []).forEach((item) => {
    const jobName = normalizeJobName(item?.job_name);
    if (!jobName) {
      return;
    }
    map.set(jobName, String(item?.id || "").trim());
  });
  return map;
}

function stringifyCount(value) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return "0";
  }
  return String(Math.max(0, numeric));
}

function formatRate(value, digits = 1) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric)) {
    return "-";
  }
  return `${(numeric * 100).toFixed(digits)}%`;
}

function normalizeStatus(value) {
  return String(value || "").trim().toUpperCase();
}

export function buildSchedulerDefinitionOptions(supportedJobs = [], definitions = []) {
  const existingJobMap = collectDefinitionJobNameMap(definitions);
  const optionMap = new Map();

  (Array.isArray(supportedJobs) ? supportedJobs : []).forEach((item) => {
    const jobName = normalizeJobName(item?.job_name);
    if (!jobName) {
      return;
    }
    optionMap.set(jobName, {
      job_name: jobName,
      display_name: normalizeDisplayName(item?.display_name, jobName),
      module: normalizeModule(item?.module) || "SYSTEM",
      alias_of: normalizeJobName(item?.alias_of),
      used: existingJobMap.has(jobName)
    });
  });

  (Array.isArray(definitions) ? definitions : []).forEach((item) => {
    const jobName = normalizeJobName(item?.job_name);
    if (!jobName || optionMap.has(jobName)) {
      return;
    }
    optionMap.set(jobName, {
      job_name: jobName,
      display_name: normalizeDisplayName(item?.display_name, jobName),
      module: normalizeModule(item?.module) || "SYSTEM",
      alias_of: "",
      used: true
    });
  });

  return Array.from(optionMap.values()).sort(compareDefinitionOptions);
}

export function buildSchedulerDefinitionCreateOptions(supportedJobs = [], definitions = []) {
  return buildSchedulerDefinitionOptions(supportedJobs, definitions).filter((item) => !item.used);
}

export function validateSchedulerDefinitionJobName(jobName, definitions = [], currentDefinitionID = "") {
  const normalizedJobName = normalizeJobName(jobName);
  if (!normalizedJobName) {
    return "任务编码不能为空";
  }
  const normalizedCurrentID = String(currentDefinitionID || "").trim();
  const duplicated = (Array.isArray(definitions) ? definitions : []).find((item) => {
    const itemJobName = normalizeJobName(item?.job_name);
    if (itemJobName !== normalizedJobName) {
      return false;
    }
    if (!normalizedCurrentID) {
      return true;
    }
    return String(item?.id || "").trim() !== normalizedCurrentID;
  });
  if (duplicated) {
    return `任务编码 ${normalizedJobName} 已存在，请直接编辑原定义或改用其他编码`;
  }
  return "";
}

export function buildSystemJobsOverviewCards({
  metrics = {},
  autoRetrySummary = {},
  definitionTotal = 0
} = {}) {
  const enabled = !!autoRetrySummary?.enabled;
  const retryText = enabled
    ? `已开启 · ${stringifyCount(autoRetrySummary?.maxRetries)}次`
    : "已关闭";

  return [
    {
      key: "today_total",
      title: "今日总运行",
      value: stringifyCount(metrics?.today_total),
      tone: "primary",
      helper: "先看任务量，再判断是否异常放大"
    },
    {
      key: "today_failed",
      title: "今日失败",
      value: stringifyCount(metrics?.today_failed),
      tone: "danger",
      helper: "优先处理失败任务和失败原因"
    },
    {
      key: "today_running",
      title: "运行中",
      value: stringifyCount(metrics?.today_running),
      tone: "warning",
      helper: "适合观察是否有长时间未结束任务"
    },
    {
      key: "auto_retry",
      title: "自动重试",
      value: retryText,
      tone: enabled ? "success" : "info",
      helper: "当前会按配置自动补救首次失败任务"
    },
    {
      key: "definition_total",
      title: "任务定义数",
      value: stringifyCount(definitionTotal),
      tone: "info",
      helper: "集中维护已有调度定义和状态"
    },
    {
      key: "recovery_rate",
      title: "恢复成功率",
      value: formatRate(metrics?.recovery_hit_rate),
      tone: "gold",
      helper: "观察失败后重试是否真正恢复"
    }
  ];
}

export function buildSystemJobsGuideCards({ canEditSystemJobs = false } = {}) {
  const permissionCard = canEditSystemJobs
    ? {
        key: "permission",
        title: "当前账号可操作",
        items: [
          "可以修改自动重试配置",
          "可以手动触发任务和批量重跑",
          "可以新增、编辑、删除任务定义"
        ]
      }
    : {
        key: "permission",
        title: "当前账号权限",
        items: [
          "当前账号仅支持查看任务总览和运行记录",
          "如需触发、重跑或改配置，请申请 system_job.edit 权限"
        ]
      };

  return [
    {
      key: "today-flow",
      title: "今天怎么处理",
      items: [
        "先看总览卡里的失败数和运行中数量",
        "再看失败原因表，确认是单任务异常还是系统性问题",
        "最后去运行记录里做重跑、复核和导出"
      ]
    },
    {
      key: "config-tips",
      title: "配置怎么用更稳",
      items: [
        "自动重试建议只放高频、可幂等的任务",
        "退避秒数不要过小，避免连续打满失败队列",
        "新增任务定义前先确认任务编码和调度表达式"
      ]
    },
    permissionCard
  ];
}

export function buildSystemJobsActionCards({ canEditSystemJobs = false, failedRunCount = 0 } = {}) {
  void canEditSystemJobs;
  void failedRunCount;
  return [];
}

export function buildSystemJobsTabOptions({ canEditSystemJobs = false } = {}) {
  void canEditSystemJobs;
  return [
    {
      key: "overview",
      label: "总览",
      description: "先看今天整体健康度和共性失败原因"
    },
    {
      key: "market-data",
      label: "同步任务",
      description: "发起股票期货同步、查看任务和同步快照"
    },
    {
      key: "runs",
      label: "运行记录",
      description: "查看详情、筛选失败和执行重跑"
    }
  ];
}

const STOCK_SYNC_ASSET_SCOPE = ["STOCK", "INDEX", "ETF", "LOF", "CBOND"];
const STOCK_FULL_SYNC_STAGES = ["UNIVERSE", "MASTER", "QUOTES", "DAILY_BASIC", "MONEYFLOW", "TRUTH", "COVERAGE_SUMMARY"];
const STOCK_INCREMENTAL_SYNC_STAGES = ["QUOTES", "DAILY_BASIC", "MONEYFLOW", "TRUTH", "COVERAGE_SUMMARY"];
const FUTURES_FULL_SYNC_STAGES = ["UNIVERSE", "MASTER", "QUOTES", "TRUTH", "COVERAGE_SUMMARY"];
const FUTURES_INCREMENTAL_SYNC_STAGES = ["QUOTES", "TRUTH", "COVERAGE_SUMMARY"];

function normalizeAssetScope(values = []) {
  return Array.from(
    new Set(
      (Array.isArray(values) ? values : [])
        .map((item) => String(item || "").trim().toUpperCase())
        .filter(Boolean)
    )
  );
}

function isStockSyncScope(assetScope = []) {
  const scope = normalizeAssetScope(assetScope);
  return scope.length > 0 && scope.every((item) => STOCK_SYNC_ASSET_SCOPE.includes(item));
}

function isFuturesSyncScope(assetScope = []) {
  const scope = normalizeAssetScope(assetScope);
  return scope.length > 0 && scope.every((item) => item === "FUTURES");
}

export function buildSyncJobTemplateOptions() {
  return [
    {
      key: "STOCK_FULL",
      label: "股票全量同步",
      description: "刷新股票主数据、行情、增强因子和 Truth"
    },
    {
      key: "STOCK_INCREMENTAL",
      label: "股票每日增量同步",
      description: "补当天或最近交易日行情、增强因子和 Truth"
    },
    {
      key: "FUTURES_FULL",
      label: "期货全量同步",
      description: "刷新期货主数据、行情和 Truth"
    },
    {
      key: "FUTURES_INCREMENTAL",
      label: "期货每日增量同步",
      description: "补当天或最近交易日期货行情和 Truth"
    }
  ];
}

export function buildSyncJobPayloadFromTemplate(templateKey, overrides = {}) {
  const normalized = String(templateKey || "").trim().toUpperCase();
  const baseByTemplate = {
    STOCK_FULL: {
      run_type: "FULL",
      asset_scope: [...STOCK_SYNC_ASSET_SCOPE],
      stages: [...STOCK_FULL_SYNC_STAGES],
      force_refresh_universe: true,
      rebuild_truth_after_sync: true
    },
    STOCK_INCREMENTAL: {
      run_type: "INCREMENTAL",
      asset_scope: [...STOCK_SYNC_ASSET_SCOPE],
      stages: [...STOCK_INCREMENTAL_SYNC_STAGES],
      force_refresh_universe: false,
      rebuild_truth_after_sync: true
    },
    FUTURES_FULL: {
      run_type: "FULL",
      asset_scope: ["FUTURES"],
      stages: [...FUTURES_FULL_SYNC_STAGES],
      force_refresh_universe: true,
      rebuild_truth_after_sync: true
    },
    FUTURES_INCREMENTAL: {
      run_type: "INCREMENTAL",
      asset_scope: ["FUTURES"],
      stages: [...FUTURES_INCREMENTAL_SYNC_STAGES],
      force_refresh_universe: false,
      rebuild_truth_after_sync: true
    }
  };

  const base = baseByTemplate[normalized];
  if (!base) {
    return {};
  }

  return {
    ...base,
    ...overrides
  };
}

export function deriveSyncJobTemplateFromBackfillRun(run = {}) {
  const runType = normalizeStatus(run?.run_type);
  const assetScope = normalizeAssetScope(run?.asset_scope);

  if (runType === "FULL" && isStockSyncScope(assetScope)) {
    return "STOCK_FULL";
  }
  if (runType === "INCREMENTAL" && isStockSyncScope(assetScope)) {
    return "STOCK_INCREMENTAL";
  }
  if (runType === "FULL" && isFuturesSyncScope(assetScope)) {
    return "FUTURES_FULL";
  }
  if (runType === "INCREMENTAL" && isFuturesSyncScope(assetScope)) {
    return "FUTURES_INCREMENTAL";
  }
  return "";
}

export function formatSyncJobTypeLabel(run = {}) {
  const templateKey = deriveSyncJobTemplateFromBackfillRun(run);
  const option = buildSyncJobTemplateOptions().find((item) => item.key === templateKey);
  if (option) {
    return option.label;
  }
  const runType = normalizeStatus(run?.run_type);
  if (runType === "REBUILD_ONLY") {
    return "仅重建";
  }
  return "自定义同步任务";
}

export function buildMarketBackfillOverviewCards({ runs = [], snapshots = [] } = {}) {
  const normalizedRuns = Array.isArray(runs) ? runs : [];
  const runningCount = normalizedRuns.filter((item) => normalizeStatus(item?.status) === "RUNNING").length;
  const failedCount = normalizedRuns.filter((item) => {
    const status = normalizeStatus(item?.status);
    return status === "FAILED" || status === "PARTIAL_SUCCESS";
  }).length;

  return [
    {
      key: "total_runs",
      title: "同步任务数",
      value: stringifyCount(normalizedRuns.length),
      tone: "primary",
      helper: "先看总量，再判断今天是否有集中同步"
    },
    {
      key: "running_runs",
      title: "进行中",
      value: stringifyCount(runningCount),
      tone: "warning",
      helper: "适合盯正在推进的阶段和卡住的批次"
    },
    {
      key: "failed_runs",
      title: "失败/部分成功",
      value: stringifyCount(failedCount),
      tone: "danger",
      helper: "失败批次可重试，不需要重新发起整单"
    },
    {
      key: "snapshots",
      title: "同步快照",
      value: stringifyCount(Array.isArray(snapshots) ? snapshots.length : 0),
      tone: "info",
      helper: "先确认快照范围，再决定发起哪类同步任务"
    }
  ];
}

export function buildMarketBackfillGuideCards({ canEditSystemJobs = false } = {}) {
  const permissionCard = canEditSystemJobs
    ? {
        key: "permission",
        title: "当前账号可操作",
        items: [
          "可以新建同步任务和重试失败批次",
          "可以查看同步快照、任务总单和批次明细"
        ]
      }
    : {
        key: "permission",
        title: "当前账号权限",
        items: [
          "当前账号仅支持查看同步状态、快照和批次明细",
          "如需发起任务或重试失败批次，请申请 system_job.edit 权限"
        ]
      };

  return [
    {
      key: "workflow",
      title: "建议处理顺序",
      items: [
        "先选择同步任务模板，再按需要微调参数",
        "先看同步任务状态，再展开批次明细定位问题",
        "失败批次可重试，不需要重新发起整单"
      ]
    },
    {
      key: "scope",
      title: "本轮真实支持范围",
      items: [
        "股票支持全量同步与每日增量同步，默认包含行情、增强因子和 Truth",
        "期货支持全量同步与每日增量同步，默认聚焦行情与 Truth 链路",
        "当前阶段不支持的子能力会明确标记为跳过，不记成失败"
      ]
    },
    permissionCard
  ];
}
