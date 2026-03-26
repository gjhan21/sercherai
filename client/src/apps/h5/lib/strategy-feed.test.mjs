import test from "node:test";
import assert from "node:assert/strict";
import { buildStrategyFeedModel } from "./strategy-feed.js";

test("buildStrategyFeedModel promotes the first strategy into a lead card", () => {
  const model = buildStrategyFeedModel([
    {
      id: "sr_1",
      title: "600036.SH 招商银行",
      summary: "资金重新回到防御资产，今天更适合先看高确定性方向。",
      meta: "2026/03/24 09:30",
      badge: "中风险",
      tone: "brand",
      tags: ["资金回流", "防御属性", "风控边界"]
    },
    {
      id: "sr_2",
      title: "300750.SZ 宁德时代",
      summary: "高景气赛道维持趋势，但波动更高，适合第二顺位观察。",
      meta: "2026/03/24 09:12",
      badge: "中风险",
      tone: "brand",
      tags: ["趋势延续", "成交配合"]
    }
  ], "股票策略");

  assert.equal(model.lead?.id, "sr_1");
  assert.equal(model.lead?.sectionLabel, "股票策略");
  assert.equal(model.lead?.signalLabel, "中风险");
  assert.deepEqual(model.digest, [
    "600036.SH 招商银行",
    "300750.SZ 宁德时代"
  ]);
  assert.deepEqual(model.stream[0], {
    id: "sr_2",
    title: "300750.SZ 宁德时代",
    summary: "高景气赛道维持趋势，但波动更高，适合第二顺位观察。",
    meta: "2026/03/24 09:12",
    signalLabel: "中风险",
    tone: "brand",
    kicker: "股票策略",
    tags: ["趋势延续", "成交配合"]
  });
});

test("buildStrategyFeedModel falls back to readable defaults for sparse rows", () => {
  const model = buildStrategyFeedModel([
    {
      id: "ev_1",
      title: "ALL · 政策事件",
      summary: "",
      meta: "",
      badge: "",
      tone: "",
      tags: []
    }
  ], "事件提醒");

  assert.equal(model.lead?.summary, "等待更多观点摘要");
  assert.equal(model.lead?.signalLabel, "待跟踪");
  assert.equal(model.lead?.tone, "brand");
  assert.deepEqual(model.stream, []);
});
