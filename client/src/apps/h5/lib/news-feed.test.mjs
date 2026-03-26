import test from "node:test";
import assert from "node:assert/strict";
import { buildNewsFeedRows, sortNewsCategories } from "./news-feed.js";

test("sortNewsCategories sorts backend categories and falls back to defaults", () => {
  const categories = sortNewsCategories([
    { id: "journal", name: "期刊", sort: 3 },
    { id: "news", name: "新闻", sort: 1 },
    { id: "report", name: "研报", sort: 2 }
  ], [
    { id: "fallback", name: "默认" }
  ]);

  assert.deepEqual(categories, [
    { key: "news", label: "新闻" },
    { key: "report", label: "研报" },
    { key: "journal", label: "期刊" }
  ]);

  assert.deepEqual(sortNewsCategories([], [{ id: "fallback", name: "默认" }]), [
    { key: "fallback", label: "默认" }
  ]);
});

test("buildNewsFeedRows creates a lead card, ticker rows and dense feed rows", () => {
  const rows = buildNewsFeedRows([
    {
      id: "n1",
      title: "北向资金连续回流，银行权重重新活跃",
      summary: "盘前更适合先看低波动与现金流资产。",
      meta: "2026/03/24 09:30 · 新闻",
      badge: "公开",
      tone: "brand",
      tags: ["新闻", "公开"],
      visibility: "PUBLIC"
    },
    {
      id: "n2",
      title: "机构周报：景气板块进入估值切换窗口",
      summary: "",
      content: "继续观察高景气板块与盈利预期的匹配度。",
      meta: "2026/03/24 08:40 · 研报",
      badge: "VIP",
      tone: "gold",
      tags: ["研报", "VIP"],
      visibility: "VIP"
    },
    {
      id: "n3",
      title: "盘前速递：指数期货贴水收敛",
      summary: "短线情绪保持修复。",
      meta: "2026/03/24 08:10 · 快讯",
      badge: "公开",
      tone: "brand",
      tags: ["快讯", "公开"],
      visibility: "PUBLIC"
    }
  ]);

  assert.equal(rows.lead?.id, "n1");
  assert.deepEqual(rows.tickerItems, [
    "北向资金连续回流，银行权重重新活跃",
    "机构周报：景气板块进入估值切换窗口",
    "盘前速递：指数期货贴水收敛"
  ]);
  assert.equal(rows.feedItems.length, 2);
  assert.deepEqual(rows.feedItems[0], {
    id: "n2",
    title: "机构周报：景气板块进入估值切换窗口",
    summary: "继续观察高景气板块与盈利预期的匹配度。",
    meta: "2026/03/24 08:40 · 研报",
    primaryTag: "研报",
    accessLabel: "VIP",
    tone: "gold",
    visibility: "VIP"
  });
});
