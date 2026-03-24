import test from "node:test";
import assert from "node:assert/strict";
import { resolveClientAppMode } from "./app-entry.js";

test("resolveClientAppMode sends /m routes to the h5 app", () => {
  assert.equal(resolveClientAppMode("/m"), "h5");
  assert.equal(resolveClientAppMode("/m/home"), "h5");
  assert.equal(resolveClientAppMode("/m/news?article=n1"), "h5");
});

test("resolveClientAppMode keeps non-mobile paths on the pc app", () => {
  assert.equal(resolveClientAppMode("/"), "pc");
  assert.equal(resolveClientAppMode("/home"), "pc");
  assert.equal(resolveClientAppMode("/auth"), "pc");
});
