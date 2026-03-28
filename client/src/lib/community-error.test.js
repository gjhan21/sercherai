import test from "node:test";
import assert from "node:assert/strict";
import { normalizeCommunityLoadError } from "./community-error.js";

test("normalizes raw database errors into user-safe community message", () => {
  const message = "Error 1146 (42S02): Table 'sercherai.discussion_topics' doesn't exist";
  assert.equal(normalizeCommunityLoadError(message), "社区讨论数据尚未初始化，当前先展示入口和说明。");
});

test("keeps generic non-sensitive community errors readable", () => {
  assert.equal(normalizeCommunityLoadError("network timeout"), "network timeout");
  assert.equal(normalizeCommunityLoadError(""), "讨论列表加载失败");
});
