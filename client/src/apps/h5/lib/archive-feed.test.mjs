import test from "node:test";
import assert from "node:assert/strict";
import { buildArchiveFeedModel } from "./archive-feed.js";

test("buildArchiveFeedModel creates summary cards and a mobile timeline", () => {
  const model = buildArchiveFeedModel({
    items: [
      {
        id: "a1",
        name: "600519.SH 贵州茅台",
        status: "已止盈",
        statusClass: "success",
        sourceLabel: "已验证",
        sourceClass: "verified",
        reason: "消费龙头估值修复。",
        explanationSummary: "高股息和盈利确定性共振。",
        cumulativeReturn: "+8.1%",
        excessReturn: "+3.2%",
        date: "2026/03/12",
        proofTags: ["高股息", "估值修复"]
      },
      {
        id: "a2",
        name: "300750.SZ 宁德时代",
        status: "已失效",
        statusClass: "warning",
        sourceLabel: "预估",
        sourceClass: "estimated",
        reason: "趋势修复不及预期。",
        explanationSummary: "量能没有继续确认。",
        cumulativeReturn: "-2.6%",
        excessReturn: "-1.1%",
        date: "2026/03/08",
        proofTags: ["量能不足"]
      }
    ]
  });

  assert.deepEqual(model.summaryCards, [
    { label: "历史样本", value: "2 条", note: "成功与失效样本都会保留" },
    { label: "优先查看", value: "已止盈", note: "先看第一条档案的结果与理由" },
    { label: "来源标签", value: "已验证", note: "每条档案都带来源说明" }
  ]);
  assert.equal(model.leadItem.id, "a1");
  assert.equal(model.timeline.length, 2);
  assert.equal(model.timeline[1].headline, "300750.SZ 宁德时代");
  assert.equal(model.timeline[1].meta, "2026/03/08 · 预估");
});

test("buildArchiveFeedModel keeps readable empty-state output", () => {
  const model = buildArchiveFeedModel({ items: [] });

  assert.equal(model.leadItem, null);
  assert.equal(model.timeline.length, 0);
  assert.deepEqual(model.summaryCards, [
    { label: "历史样本", value: "0 条", note: "公开历史样本会显示在这里" },
    { label: "优先查看", value: "暂无", note: "后续会按时间线展示" },
    { label: "来源标签", value: "待补充", note: "来源说明会跟随档案一起出现" }
  ]);
  assert.deepEqual(model.emptyState, {
    title: "当前还没有可展示的历史档案",
    description: "刷新后会按时间顺序展示历史推荐、结果和来源说明。"
  });
});
