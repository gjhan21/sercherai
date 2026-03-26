import test from "node:test";
import assert from "node:assert/strict";

import { buildMembershipStatusView, mapActivationStateLabel } from "./membership.js";

test("mapActivationStateLabel renders Chinese labels", () => {
  assert.equal(mapActivationStateLabel("ACTIVE"), "已激活");
  assert.equal(mapActivationStateLabel("PAID_PENDING_KYC"), "待实名激活");
  assert.equal(mapActivationStateLabel("NON_MEMBER"), "未开通");
});

test("buildMembershipStatusView highlights paid but pending kyc members", () => {
  const view = buildMembershipStatusView(
    {
      member_level: "VIP1",
      kyc_status: "PENDING",
      activation_state: "PAID_PENDING_KYC"
    },
    {
      vip_expire_at: "2026-04-20T23:59:59+08:00",
      doc_read_remaining: 18,
      news_subscribe_remaining: 9
    }
  );

  assert.equal(view.badge, "待实名激活");
  assert.match(view.title, /已开通/);
  assert.match(view.description, /实名/);
  assert.deepEqual(view.stats, [
    { label: "会员层级", value: "VIP1" },
    { label: "研报额度", value: "18" },
    { label: "资讯额度", value: "9" }
  ]);
});
