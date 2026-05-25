import test from "node:test";
import assert from "node:assert/strict";

import * as systemJobsAdmin from "./system-jobs-admin.js";

const {
  buildMarketBackfillGuideCards,
  buildMarketBackfillOverviewCards,
  buildSyncJobPayloadFromTemplate,
  buildSyncJobTemplateOptions,
  buildSchedulerDefinitionOptions,
  buildSchedulerDefinitionCreateOptions,
  formatSyncJobTypeLabel,
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

test("buildSystemJobsActionCards is disabled after the task center streamline", () => {
  assert.equal(typeof systemJobsAdmin.buildSystemJobsActionCards, "function");

  assert.deepEqual(systemJobsAdmin.buildSystemJobsActionCards(), []);
  assert.deepEqual(
    systemJobsAdmin.buildSystemJobsActionCards({
      canEditSystemJobs: true,
      failedRunCount: 3
    }),
    []
  );
});

test("buildSystemJobsTabOptions returns the streamlined three-tab task center layout", () => {
  assert.equal(typeof systemJobsAdmin.buildSystemJobsTabOptions, "function");

  assert.deepEqual(systemJobsAdmin.buildSystemJobsTabOptions({ canEditSystemJobs: true }), [
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
  ]);

  assert.deepEqual(systemJobsAdmin.buildSystemJobsTabOptions({ canEditSystemJobs: false }), [
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
  ]);
});

test("buildMarketBackfillOverviewCards summarizes market backfill workspace in Chinese", () => {
  assert.deepEqual(
    buildMarketBackfillOverviewCards({
      runs: [
        { status: "SUCCESS" },
        { status: "RUNNING" },
        { status: "FAILED" },
        { status: "PARTIAL_SUCCESS" }
      ],
      snapshots: [{ id: "mus_001" }, { id: "mus_002" }]
    }),
    [
      {
        key: "total_runs",
        title: "同步任务数",
        value: "4",
        tone: "primary",
        helper: "先看总量，再判断今天是否有集中同步"
      },
      {
        key: "running_runs",
        title: "进行中",
        value: "1",
        tone: "warning",
        helper: "适合盯正在推进的阶段和卡住的批次"
      },
      {
        key: "failed_runs",
        title: "失败/部分成功",
        value: "2",
        tone: "danger",
        helper: "失败批次可重试，不需要重新发起整单"
      },
      {
        key: "snapshots",
        title: "同步快照",
        value: "2",
        tone: "info",
        helper: "先确认快照范围，再决定发起哪类同步任务"
      }
    ]
  );
});

test("buildMarketBackfillGuideCards returns Chinese operator guidance under sync task language", () => {
  assert.deepEqual(buildMarketBackfillGuideCards({ canEditSystemJobs: true }), [
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
    {
      key: "permission",
      title: "当前账号可操作",
      items: [
        "可以新建同步任务和重试失败批次",
        "可以查看同步快照、任务总单和批次明细"
      ]
    }
  ]);
});

test("sync task helper layer exposes new stock and futures task templates", () => {
  assert.equal(typeof buildSyncJobTemplateOptions, "function");
  assert.equal(typeof buildSyncJobPayloadFromTemplate, "function");
  assert.equal(typeof formatSyncJobTypeLabel, "function");

  assert.deepEqual(
    buildSyncJobTemplateOptions().map((item) => item.key),
    ["STOCK_FULL", "STOCK_INCREMENTAL", "FUTURES_FULL", "FUTURES_INCREMENTAL"]
  );
});
