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
  const hasFailedRuns = Number(failedRunCount) > 0;
  const firstCard = hasFailedRuns
    ? {
        key: "view-failed-runs",
        title: "先处理失败任务",
        description: `当前页有 ${stringifyCount(failedRunCount)} 条失败运行，建议先过滤查看并决定是否重跑`,
        actionText: "查看失败任务",
        tone: "danger"
      }
    : {
        key: "view-failed-runs",
        title: "查看失败任务",
        description: "当前页没有失败运行，可以切换筛选继续核对历史记录",
        actionText: "筛选失败记录",
        tone: "info"
      };

  const cards = [
    firstCard,
    {
      key: "refresh-all",
      title: "刷新任务面板",
      description: "同步最新指标、配置、任务定义和运行记录",
      actionText: "刷新全部",
      tone: "primary"
    }
  ];

  if (canEditSystemJobs) {
    cards.push(
      {
        key: "open-create-definition",
        title: "新增任务定义",
        description: "适合补充新的定时任务或补齐空缺定义",
        actionText: "新增定义",
        tone: "info"
      },
      {
        key: "scroll-trigger",
        title: "手动触发任务",
        description: "需要临时补跑、联调或验证时，从这里快速进入",
        actionText: "去触发区",
        tone: "gold"
      }
    );
    return cards;
  }

  cards.push({
    key: "scroll-definitions",
    title: "查看任务定义",
    description: "快速跳到任务定义列表，核对状态、表达式和最近执行情况",
    actionText: "去任务定义",
    tone: "gold"
  });
  return cards;
}

export function buildSystemJobsTabOptions({ canEditSystemJobs = false } = {}) {
  const tabs = [
    {
      key: "overview",
      label: "总览",
      description: "看今日运行、失败原因和使用说明"
    },
    {
      key: "config",
      label: "任务配置",
      description: "管理自动重试和任务定义"
    }
  ];

  if (canEditSystemJobs) {
    tabs.push({
      key: "trigger",
      label: "手动触发",
      description: "临时补跑、联调和手动触发任务"
    });
  }

  tabs.push({
    key: "runs",
    label: "运行记录",
    description: "查看详情、筛选失败和执行重跑"
  });

  return tabs;
}
