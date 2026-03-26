import test from "node:test";
import assert from "node:assert/strict";

import { normalizeRedirectPath, shouldServeH5Index } from "./app-paths.js";

test("shouldServeH5Index only rewrites mobile app paths", () => {
  assert.equal(shouldServeH5Index("/m"), true);
  assert.equal(shouldServeH5Index("/m/home"), true);
  assert.equal(shouldServeH5Index("/m/news?keyword=新能源"), true);
  assert.equal(shouldServeH5Index("/"), false);
  assert.equal(shouldServeH5Index("/api/v1/news"), false);
  assert.equal(shouldServeH5Index("/pc/home"), false);
});

test("normalizeRedirectPath keeps safe in-app routes and falls back to /home", () => {
  assert.equal(normalizeRedirectPath("/news"), "/news");
  assert.equal(normalizeRedirectPath("membership"), "/membership");
  assert.equal(normalizeRedirectPath("https://evil.example.com"), "/home");
  assert.equal(normalizeRedirectPath("javascript:alert(1)"), "/home");
  assert.equal(normalizeRedirectPath(""), "/home");
});
