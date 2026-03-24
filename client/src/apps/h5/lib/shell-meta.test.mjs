import test from "node:test";
import assert from "node:assert/strict";
import { resolveShellScene, resolveTabItems } from "./shell-meta.js";

test("resolveShellScene maps h5 routes to content-app scene copy", () => {
  assert.deepEqual(resolveShellScene("/home"), {
    section: "首页",
    title: "今日观点",
    subtitle: "先看核心判断，再顺着内容流继续阅读",
    pulse: "内容优先"
  });

  assert.deepEqual(resolveShellScene("/news?article=n1"), {
    section: "资讯",
    title: "市场资讯",
    subtitle: "栏目切换后直接进入正文与内容流",
    pulse: "阅读中"
  });

  assert.deepEqual(resolveShellScene("/strategies?id=s1"), {
    section: "策略",
    title: "精选观点",
    subtitle: "像内容 App 一样查看结论、理由和风险边界",
    pulse: "观点流"
  });
});

test("resolveShellScene falls back to profile/account scene and home defaults", () => {
  assert.deepEqual(resolveShellScene("/membership"), {
    section: "会员",
    title: "会员中心",
    subtitle: "套餐、支付与激活状态集中在一页完成",
    pulse: "收银台"
  });

  assert.deepEqual(resolveShellScene("/profile"), {
    section: "我的",
    title: "账户中心",
    subtitle: "消息、订单、实名和会员状态统一管理",
    pulse: "账户"
  });

  assert.deepEqual(resolveShellScene("/watchlist"), {
    section: "关注",
    title: "我的关注",
    subtitle: "变化工作台按回访节奏展示跟踪对象",
    pulse: "回访中"
  });

  assert.deepEqual(resolveShellScene("/archive"), {
    section: "档案",
    title: "历史档案",
    subtitle: "历史样本、结果和版本变化统一按时间线查看",
    pulse: "复盘"
  });

  assert.equal(resolveShellScene("/unknown").section, "首页");
});

test("resolveTabItems keeps app-first tab order with explicit icons", () => {
  assert.deepEqual(resolveTabItems().map((item) => ({ path: item.path, label: item.label, icon: item.icon })), [
    { path: "/home", label: "首页", icon: "home" },
    { path: "/news", label: "资讯", icon: "news" },
    { path: "/strategies", label: "策略", icon: "insight" },
    { path: "/membership", label: "会员", icon: "vip" },
    { path: "/profile", label: "我的", icon: "user" }
  ]);
});
