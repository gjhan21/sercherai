import test from "node:test";
import assert from "node:assert/strict";
import { shouldServeH5Index } from "./vite.h5-paths.js";

test("shouldServeH5Index rewrites h5 history routes to the m entry", () => {
  assert.equal(shouldServeH5Index("/m/home"), true);
  assert.equal(shouldServeH5Index("/m/news?article=n1"), true);
  assert.equal(shouldServeH5Index("/m/strategies?id=s1"), true);
});

test("shouldServeH5Index leaves assets and api routes alone", () => {
  assert.equal(shouldServeH5Index("/m/src/apps/h5/main.js"), false);
  assert.equal(shouldServeH5Index("/m/assets/index.css"), false);
  assert.equal(shouldServeH5Index("/api/v1/auth/login"), false);
  assert.equal(shouldServeH5Index("/uploads/demo.png"), false);
});
