import test from "node:test";
import assert from "node:assert/strict";
import { buildProfileCenterModel } from "./profile-center.js";

test("buildProfileCenterModel prioritizes activation, unread messages, and invite conversion", () => {
  const model = buildProfileCenterModel({
    profile: {
      phone: "13812345678",
      kyc_status: "PENDING"
    },
    quota: {
      member_level: "VIP1",
      vip_status: "ACTIVE",
      activation_state: "PAID_PENDING_KYC",
      vip_remaining_days: 20
    },
    orders: [
      {
        order_no: "MO20260324001",
        status: "PAID",
        paid_at: "2026-03-24T09:30:00+08:00"
      }
    ],
    messages: [
      {
        id: "m1",
        type: "STRATEGY",
        title: "策略提醒",
        content: "今日建议先看防御方向。",
        read_status: "UNREAD",
        created_at: "2026-03-24T09:40:00+08:00"
      },
      {
        id: "m2",
        type: "SYSTEM",
        title: "系统通知",
        content: "账户安全状态正常。",
        read_status: "READ",
        created_at: "2026-03-23T20:10:00+08:00"
      }
    ],
    shareLinks: [
      {
        id: "s1",
        invite_code: "VIP2026",
        url: "https://example.com/invite/VIP2026"
      }
    ],
    inviteRecords: [
      {
        id: "i1",
        invitee_user_id: "u1001",
        status: "FIRST_PAID",
        first_pay_at: "2026-03-24T12:00:00+08:00"
      }
    ],
    inviteSummary: {
      last_7d_registered_count: 3,
      last_7d_first_paid_count: 1,
      share_link_count: 1
    }
  });

  assert.equal(model.hero.displayName, "138****5678");
  assert.equal(model.hero.memberLevel, "VIP 1");
  assert.equal(model.hero.activationState, "待实名激活");
  assert.deepEqual(model.hero.metrics, [
    { label: "实名", value: "审核中", note: "支付后待完成" },
    { label: "消息", value: "1 条", note: "未读待处理" },
    { label: "邀请", value: "3 人", note: "7日注册" }
  ]);

  assert.equal(model.todos[0].id, "kyc");
  assert.equal(model.todos[0].actionLabel, "去实名");
  assert.equal(model.serviceCards[0].id, "membership");
  assert.equal(model.serviceCards[1].id, "messages");
  assert.equal(model.inviteOverview.primaryCode, "VIP2026");
  assert.equal(model.sticky.primaryTarget, "kyc");
  assert.equal(model.sticky.primaryLabel, "提交实名");
});

test("buildProfileCenterModel keeps readable fallbacks without account data", () => {
  const model = buildProfileCenterModel({});

  assert.equal(model.hero.displayName, "我的账户");
  assert.equal(model.hero.memberLevel, "普通会员");
  assert.equal(model.hero.metrics.length, 3);
  assert.equal(model.todos[0].id, "membership");
  assert.equal(model.shortcuts.length, 4);
  assert.equal(model.serviceCards[2].id, "invite");
  assert.equal(model.inviteOverview.primaryCode, "暂未创建邀请码");
  assert.equal(model.sticky.primaryTarget, "membership");
});
