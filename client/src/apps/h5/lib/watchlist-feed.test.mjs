import test from "node:test";
import assert from "node:assert/strict";
import { buildWatchlistFeedModel } from "./watchlist-feed.js";

test("buildWatchlistFeedModel creates summary cards, lead item, and follow-up rows", () => {
  const model = buildWatchlistFeedModel({
    items: [
      {
        id: "w1",
        name: "600519.SH 贵州茅台",
        summary: "先看结论变化，再决定是否继续持有。",
        status: "持续跟踪",
        risk: "中风险",
        nextAction: "跟踪财报确认",
        performance: "+6.2%",
        performanceClass: "up",
        proofTags: ["高景气", "资金回流"],
        latestNewsTitle: "高端消费继续修复",
        latestNewsTime: "今天 09:30"
      },
      {
        id: "w2",
        name: "300750.SZ 宁德时代",
        summary: "观察量能是否继续确认。",
        status: "等待确认",
        risk: "高风险",
        nextAction: "等趋势确认",
        performance: "-1.8%",
        performanceClass: "down",
        proofTags: ["量能不足"],
        latestNewsTitle: "板块分化加剧",
        latestNewsTime: "今天 10:10"
      }
    ]
  });

  assert.deepEqual(model.summaryCards, [
    { label: "已保存", value: "2 项", note: "本地关注可跨页面继续跟踪" },
    { label: "优先跟踪", value: "持续跟踪", note: "默认展示第一条已保存对象" },
    { label: "最新动态", value: "高端消费继续修复", note: "今天 09:30" }
  ]);
  assert.equal(model.leadItem.id, "w1");
  assert.equal(model.leadItem.performance, "+6.2%");
  assert.equal(model.leadItem.proofTags.length, 2);
  assert.deepEqual(model.rows.map((item) => item.id), ["w2"]);
  assert.equal(model.rows[0].headline, "300750.SZ 宁德时代");
});

test("buildWatchlistFeedModel keeps readable empty-state output", () => {
  const model = buildWatchlistFeedModel({ items: [] });

  assert.equal(model.leadItem, null);
  assert.equal(model.rows.length, 0);
  assert.deepEqual(model.summaryCards, [
    { label: "已保存", value: "0 项", note: "先从首页或策略页加入关注" },
    { label: "优先跟踪", value: "暂无", note: "保存后这里会出现回访主线" },
    { label: "最新动态", value: "待同步", note: "关注后的变化会集中显示" }
  ]);
  assert.deepEqual(model.emptyState, {
    title: "还没有保存任何关注对象",
    description: "可以先从首页或策略页加入关注，再回到这里持续跟踪变化。"
  });
});
