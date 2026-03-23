import test from "node:test";
import assert from "node:assert/strict";

import {
  buildSchedulerDefinitionOptions,
  buildSchedulerDefinitionCreateOptions,
  validateSchedulerDefinitionJobName
} from "./system-jobs-admin.js";

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
