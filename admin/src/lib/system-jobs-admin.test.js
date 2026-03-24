import test from "node:test";
import assert from "node:assert/strict";

import * as systemJobsAdmin from "./system-jobs-admin.js";

const {
  buildSchedulerDefinitionOptions,
  buildSchedulerDefinitionCreateOptions,
  validateSchedulerDefinitionJobName
} = systemJobsAdmin;

test("buildSchedulerDefinitionCreateOptions filters out supported jobs that already have definitions", () => {
  const supported = [
    { job_name: "daily_stock_quant_pipeline", display_name: "每日股票量化流水线", module: "STOCK" },
    { job_name: "vip_membership_lifecycle", display_name: "VIP会员生命周期任务", module: "SYSTEM" },
    { job_name: "nightly_cleanup", display_name: "夜间清理", module: "SYSTEM" }
  ];
  const definitions = [
    { job_name: "daily_stock_quant_pipeline", display_name: "每日股票量化流水线", module: "STOCK" },
    { job_name: "vip_membership_lifecycle", display_name: "VIP会员生命周期任务", module: "SYSTEM" }
  ];

  assert.deepEqual(buildSchedulerDefinitionCreateOptions(supported, definitions), [
    {
      alias_of: "",
      display_name: "夜间清理",
      job_name: "nightly_cleanup",
      module: "SYSTEM",
      used: false
    }
  ]);
});

test("buildSchedulerDefinitionOptions keeps used jobs visible for edit mode and marks them as used", () => {
  const supported = [
    { job_name: "daily_stock_quant_pipeline", display_name: "每日股票量化流水线", module: "STOCK" },
    { job_name: "vip_membership_lifecycle", display_name: "VIP会员生命周期任务", module: "SYSTEM" }
  ];
  const definitions = [
    { job_name: "daily_stock_quant_pipeline", display_name: "量化已配置", module: "STOCK" }
  ];

  assert.deepEqual(buildSchedulerDefinitionOptions(supported, definitions), [
    {
      alias_of: "",
      display_name: "每日股票量化流水线",
      job_name: "daily_stock_quant_pipeline",
      module: "STOCK",
      used: true
    },
    {
      alias_of: "",
      display_name: "VIP会员生命周期任务",
      job_name: "vip_membership_lifecycle",
      module: "SYSTEM",
      used: false
    }
  ]);
});

test("validateSchedulerDefinitionJobName rejects duplicate job names outside current edit target", () => {
  const definitions = [
    { id: "jobdef_001", job_name: "daily_stock_quant_pipeline" },
    { id: "jobdef_002", job_name: "vip_membership_lifecycle" }
  ];

  assert.equal(
    validateSchedulerDefinitionJobName("daily_stock_quant_pipeline", definitions),
    "任务编码 daily_stock_quant_pipeline 已存在，请直接编辑原定义或改用其他编码"
  );
  assert.equal(validateSchedulerDefinitionJobName("new_job_name", definitions), "");
  assert.equal(
    validateSchedulerDefinitionJobName("daily_stock_quant_pipeline", definitions, "jobdef_001"),
    ""
  );
});

test("buildSystemJobsOverviewCards builds Chinese summary cards for the task center", () => {
  assert.equal(typeof systemJobsAdmin.buildSystemJobsOverviewCards, "function");

  const cards = systemJobsAdmin.buildSystemJobsOverviewCards({
    metrics: {
      today_total: 26,
      today_failed: 4,
      today_running: 2,
      recovery_hit_rate: 0.625
    },
    autoRetrySummary: {
      enabled: true,
      maxRetries: 3
    },
    definitionTotal: 12,
    runTotal: 58
  });

  assert.deepEqual(cards, [
    {
      key: "today_total",
      title: "今日总运行",
      value: "26",
      tone: "primary",
      helper: "先看任务量，再判断是否异常放大"
    },
    {
      key: "today_failed",
      title: "今日失败",
      value: "4",
      tone: "danger",
      helper: "优先处理失败任务和失败原因"
    },
    {
      key: "today_running",
      title: "运行中",
      value: "2",
      tone: "warning",
      helper: "适合观察是否有长时间未结束任务"
    },
    {
      key: "auto_retry",
      title: "自动重试",
      value: "已开启 · 3次",
      tone: "success",
      helper: "当前会按配置自动补救首次失败任务"
    },
    {
      key: "definition_total",
      title: "任务定义数",
      value: "12",
      tone: "info",
      helper: "集中维护已有调度定义和状态"
    },
    {
      key: "recovery_rate",
      title: "恢复成功率",
      value: "62.5%",
      tone: "gold",
      helper: "观察失败后重试是否真正恢复"
    }
  ]);
});

test("buildSystemJobsGuideCards returns editable and read only Chinese usage guides", () => {
  assert.equal(typeof systemJobsAdmin.buildSystemJobsGuideCards, "function");

  const editableCards = systemJobsAdmin.buildSystemJobsGuideCards({ canEditSystemJobs: true });
  const readonlyCards = systemJobsAdmin.buildSystemJobsGuideCards({ canEditSystemJobs: false });

  assert.equal(editableCards[0].title, "今天怎么处理");
  assert.deepEqual(editableCards[0].items, [
    "先看总览卡里的失败数和运行中数量",
    "再看失败原因表，确认是单任务异常还是系统性问题",
    "最后去运行记录里做重跑、复核和导出"
  ]);
  assert.equal(editableCards[2].title, "当前账号可操作");
  assert.deepEqual(editableCards[2].items, [
    "可以修改自动重试配置",
    "可以手动触发任务和批量重跑",
    "可以新增、编辑、删除任务定义"
  ]);

  assert.equal(readonlyCards[2].title, "当前账号权限");
  assert.deepEqual(readonlyCards[2].items, [
    "当前账号仅支持查看任务总览和运行记录",
    "如需触发、重跑或改配置，请申请 system_job.edit 权限"
  ]);
});

test("buildSystemJobsActionCards prioritizes failed runs and edit actions", () => {
  assert.equal(typeof systemJobsAdmin.buildSystemJobsActionCards, "function");

  assert.deepEqual(
    systemJobsAdmin.buildSystemJobsActionCards({
      canEditSystemJobs: true,
      failedRunCount: 3
    }),
    [
      {
        key: "view-failed-runs",
        title: "先处理失败任务",
        description: "当前页有 3 条失败运行，建议先过滤查看并决定是否重跑",
        actionText: "查看失败任务",
        tone: "danger"
      },
      {
        key: "refresh-all",
        title: "刷新任务面板",
        description: "同步最新指标、配置、任务定义和运行记录",
        actionText: "刷新全部",
        tone: "primary"
      },
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
    ]
  );

  assert.deepEqual(
    systemJobsAdmin.buildSystemJobsActionCards({
      canEditSystemJobs: false,
      failedRunCount: 0
    }),
    [
      {
        key: "view-failed-runs",
        title: "查看失败任务",
        description: "当前页没有失败运行，可以切换筛选继续核对历史记录",
        actionText: "筛选失败记录",
        tone: "info"
      },
      {
        key: "refresh-all",
        title: "刷新任务面板",
        description: "同步最新指标、配置、任务定义和运行记录",
        actionText: "刷新全部",
        tone: "primary"
      },
      {
        key: "scroll-definitions",
        title: "查看任务定义",
        description: "快速跳到任务定义列表，核对状态、表达式和最近执行情况",
        actionText: "去任务定义",
        tone: "gold"
      }
    ]
  );
});

test("buildSystemJobsTabOptions returns the recommended four-tab task center layout", () => {
  assert.equal(typeof systemJobsAdmin.buildSystemJobsTabOptions, "function");

  assert.deepEqual(systemJobsAdmin.buildSystemJobsTabOptions({ canEditSystemJobs: true }), [
    {
      key: "overview",
      label: "总览",
      description: "看今日运行、失败原因和使用说明"
    },
    {
      key: "config",
      label: "任务配置",
      description: "管理自动重试和任务定义"
    },
    {
      key: "trigger",
      label: "手动触发",
      description: "临时补跑、联调和手动触发任务"
    },
    {
      key: "runs",
      label: "运行记录",
      description: "查看详情、筛选失败和执行重跑"
    }
  ]);

  assert.deepEqual(systemJobsAdmin.buildSystemJobsTabOptions({ canEditSystemJobs: false }), [
    {
      key: "overview",
      label: "总览",
      description: "看今日运行、失败原因和使用说明"
    },
    {
      key: "config",
      label: "任务配置",
      description: "管理自动重试和任务定义"
    },
    {
      key: "runs",
      label: "运行记录",
      description: "查看详情、筛选失败和执行重跑"
    }
  ]);
});
