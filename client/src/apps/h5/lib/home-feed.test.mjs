import test from "node:test";
import assert from "node:assert/strict";
import { buildHomeFeedModel } from "./home-feed.js";

test("buildHomeFeedModel promotes the first stock into the daily lead card", () => {
  const model = buildHomeFeedModel({
    stocks: [
      {
        id: "s1",
        title: "600036.SH 招商银行",
        summary: "今天优先看高确定性防御资产。",
        risk: "中风险",
        takeProfit: "上涨6%-10%",
        position: "10%-15%",
        meta: "2026/03/24 09:30",
        tags: ["资金回流", "防御属性", "风控边界"]
      },
      {
        id: "s2",
        title: "300750.SZ 宁德时代",
        summary: "更适合第二顺位观察。",
        risk: "中风险",
        takeProfit: "上涨5%-9%",
        position: "8%-12%",
        meta: "2026/03/24 09:12",
        tags: ["趋势延续", "成交配合"]
      }
    ],
    article: {
      id: "n1",
      title: "盘前资金风格切换，银行与高股息板块走强",
      summary: "市场风险偏好回落，早盘更关注防御属性板块轮动。"
    }
  });

  assert.equal(model.leadStock?.id, "s1");
  assert.deepEqual(model.marketPulse, [
    { label: "主线风险", value: "中风险" },
    { label: "建议仓位", value: "10%-15%" },
    { label: "目标区间", value: "上涨6%-10%" }
  ]);
  assert.deepEqual(model.watchlistPreview, [
    {
      id: "s2",
      title: "300750.SZ 宁德时代",
      summary: "更适合第二顺位观察。",
      meta: "2026/03/24 09:12",
      risk: "中风险",
      tags: ["趋势延续", "成交配合"]
    }
  ]);
  assert.equal(model.newsBrief?.id, "n1");
});

test("buildHomeFeedModel keeps readable fallbacks when no data exists", () => {
  const model = buildHomeFeedModel({ stocks: [], article: null });

  assert.equal(model.leadStock, null);
  assert.equal(model.marketPulse.length, 0);
  assert.equal(model.watchlistPreview.length, 0);
  assert.equal(model.newsBrief, null);
});
