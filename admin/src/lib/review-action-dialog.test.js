import test from "node:test";
import assert from "node:assert/strict";

import {
  extractReviewConflictReason,
  resolveReviewDialogMeta
} from "./review-action-dialog.js";

test("resolveReviewDialogMeta returns the expected labels for each review mode", () => {
  assert.deepEqual(resolveReviewDialogMeta("approve"), {
    title: "审核通过并发布",
    primaryText: "确认发布",
    primaryType: "primary",
    summaryTone: "primary"
  });

  assert.deepEqual(resolveReviewDialogMeta("blocked"), {
    title: "默认发布已被拦截",
    primaryText: "改为强制发布",
    primaryType: "primary",
    summaryTone: "warning"
  });

  assert.deepEqual(resolveReviewDialogMeta("force"), {
    title: "强制发布确认",
    primaryText: "确认强制发布",
    primaryType: "warning",
    summaryTone: "warning"
  });

  assert.deepEqual(resolveReviewDialogMeta("reject"), {
    title: "驳回审核",
    primaryText: "确认驳回",
    primaryType: "danger",
    summaryTone: "danger"
  });
});

test("extractReviewConflictReason prefers structured 409 payloads", () => {
  const reason = extractReviewConflictReason({
    code: 40901,
    payload: {
      conflict_type: "PUBLISH_POLICY_BLOCKED",
      detail: "警告数量 4 超过阈值 3"
    },
    message: "业务错误"
  });

  assert.equal(reason, "警告数量 4 超过阈值 3");
});

test("extractReviewConflictReason still parses legacy raw conflict messages", () => {
  const reason = extractReviewConflictReason({
    message:
      'strategy-engine returned 409 when publishing job job_demo: {"detail":"发布策略拦截: 警告数量 5 超过阈值 3"}'
  });

  assert.equal(reason, "发布策略拦截: 警告数量 5 超过阈值 3");
});

test("extractReviewConflictReason ignores unrelated errors", () => {
  const reason = extractReviewConflictReason({
    code: 50001,
    message: "network error"
  });

  assert.equal(reason, "");
});
