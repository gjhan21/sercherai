import test from "node:test";
import assert from "node:assert/strict";
import {
  resolveHomeMembershipSummary,
  resolveStrategyAccessSummary
} from "./membership-display.js";

test("home membership summary does not flash non-member copy while logged-in quota is loading", () => {
  const summary = resolveHomeMembershipSummary({}, { isLoggedIn: true, loading: true });

  assert.equal(summary.value, "同步中");
  assert.equal(summary.note, "正在确认会员状态");
});

test("home membership summary still shows non-member copy for logged-out visitors", () => {
  const summary = resolveHomeMembershipSummary({}, { isLoggedIn: false, loading: false });

  assert.equal(summary.value, "普通会员");
  assert.equal(summary.note, "未开通或待激活");
});

test("home membership summary shows vip copy after quota resolves", () => {
  const summary = resolveHomeMembershipSummary(
    { member_level: "VIP2", vip_status: "ACTIVE", activation_state: "ACTIVE", vip_remaining_days: 64 },
    { isLoggedIn: true, loading: false }
  );

  assert.equal(summary.value, "VIP 2");
  assert.equal(summary.note, "生效中");
});

test("strategy access summary does not flash normal browsing while logged-in quota is loading", () => {
  const summary = resolveStrategyAccessSummary({}, { isLoggedIn: true, loading: true });

  assert.equal(summary.value, "同步中");
  assert.equal(summary.note, "正在确认会员能力");
});

test("strategy access summary shows member-ready copy after quota resolves", () => {
  const summary = resolveStrategyAccessSummary(
    { member_level: "VIP2", vip_status: "ACTIVE", activation_state: "ACTIVE", vip_remaining_days: 64 },
    { isLoggedIn: true, loading: false }
  );

  assert.equal(summary.value, "会员已同步");
  assert.equal(summary.note, "可继续深读更多内容");
});
