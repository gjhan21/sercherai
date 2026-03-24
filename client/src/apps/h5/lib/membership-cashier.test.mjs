import test from "node:test";
import assert from "node:assert/strict";
import { buildMembershipCashierModel } from "./membership-cashier.js";

test("buildMembershipCashierModel marks current and recommended plans", () => {
  const model = buildMembershipCashierModel({
    products: [
      { id: "p1", name: "VIP月卡", member_level: "VIP1", duration_days: 30, price: 99, status: "ACTIVE", feature_list: ["正文解锁"] },
      { id: "p2", name: "VIP季卡", member_level: "VIP2", duration_days: 90, price: 268, status: "ACTIVE", feature_list: ["策略跟踪"] },
      { id: "p3", name: "VIP年卡", member_level: "VIP3", duration_days: 365, price: 899, status: "ACTIVE", feature_list: ["全年深读"] }
    ],
    quota: {
      member_level: "VIP1",
      vip_status: "ACTIVE",
      activation_state: "ACTIVE"
    },
    mapMemberLevel: (value) => ({ VIP1: "VIP 1", VIP2: "VIP 2", VIP3: "VIP 3" }[value] || value),
    resolveVipStage: () => true
  });

  assert.equal(model.plans[0].current, true);
  assert.equal(model.plans[1].recommended, true);
  assert.equal(model.spotlightPlan.id, "p2");
  assert.equal(model.spotlightPlan.badge, "推荐方案");
  assert.equal(model.plans[1].sceneLabel, "适合持续跟踪主线与资讯深读");
});

test("buildMembershipCashierModel normalizes raw product names into readable mobile plan labels", () => {
  const model = buildMembershipCashierModel({
    products: [
      { id: "p1", name: "Codex QA Product 1773908535399 edit-069526", member_level: "VIP1", duration_days: 30, price: 123.45, status: "ACTIVE", feature_list: ["正文解锁", "附件下载"] }
    ],
    quota: {
      member_level: "FREE",
      vip_status: "INACTIVE",
      activation_state: "NONE"
    },
    mapMemberLevel: (value) => ({ VIP1: "VIP 1" }[value] || value),
    resolveVipStage: () => false
  });

  assert.equal(model.spotlightPlan.id, "p1");
  assert.equal(model.spotlightPlan.displayName, "VIP 1 月卡");
  assert.equal(model.spotlightPlan.badge, "立即开通");
  assert.equal(model.spotlightPlan.dailyPriceText, "约 ¥4.12 / 天");
  assert.equal(model.spotlightPlan.sceneLabel, "适合先体验完整会员能力");
  assert.deepEqual(model.spotlightPlan.highlights, ["正文解锁", "附件下载"]);
});

test("buildMembershipCashierModel falls back to the first active plan when user is not active", () => {
  const model = buildMembershipCashierModel({
    products: [
      { id: "p1", name: "VIP月卡", member_level: "VIP1", duration_days: 30, price: 99, status: "ACTIVE", feature_list: ["正文解锁"] }
    ],
    quota: {
      member_level: "FREE",
      vip_status: "INACTIVE",
      activation_state: "NONE"
    },
    mapMemberLevel: (value) => value,
    resolveVipStage: () => false
  });

  assert.equal(model.spotlightPlan.id, "p1");
  assert.equal(model.spotlightPlan.badge, "立即开通");
  assert.equal(model.plans[0].actionLabel, "立即下单");
});
