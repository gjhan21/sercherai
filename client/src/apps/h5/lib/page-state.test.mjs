import test from "node:test";
import assert from "node:assert/strict";
import { mergeRecordById, resolveNewsAccessState } from "./page-state.js";

test("resolveNewsAccessState locks VIP content for guests", () => {
  const state = resolveNewsAccessState({ isLoggedIn: false, hasVipAccess: false, visibility: "VIP" });
  assert.equal(state.locked, true);
  assert.match(state.message, /请先登录/);
});

test("resolveNewsAccessState locks VIP content for non-members", () => {
  const state = resolveNewsAccessState({ isLoggedIn: true, hasVipAccess: false, visibility: "VIP" });
  assert.equal(state.locked, true);
  assert.match(state.message, /VIP 专享/);
});

test("resolveNewsAccessState unlocks public and entitled content", () => {
  assert.equal(resolveNewsAccessState({ isLoggedIn: false, hasVipAccess: false, visibility: "PUBLIC" }).locked, false);
  assert.equal(resolveNewsAccessState({ isLoggedIn: true, hasVipAccess: true, visibility: "VIP" }).locked, false);
});

test("mergeRecordById appends missing detail rows for deep links", () => {
  const rows = [{ id: "1", symbol: "AAA" }];
  const merged = mergeRecordById(rows, { id: "2", name: "Beta" });
  assert.equal(merged.length, 2);
  assert.deepEqual(merged[1], { id: "2", name: "Beta" });
});

test("mergeRecordById updates an existing row without duplicating it", () => {
  const rows = [{ id: "1", symbol: "AAA", name: "Alpha" }];
  const merged = mergeRecordById(rows, { id: "1", name: "Alpha 2" });
  assert.equal(merged.length, 1);
  assert.deepEqual(merged[0], { id: "1", symbol: "AAA", name: "Alpha 2" });
});
