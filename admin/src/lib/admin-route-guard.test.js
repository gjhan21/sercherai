import test from "node:test";
import assert from "node:assert/strict";

import { NO_ACCESS_ROUTE_PATH } from "./admin-navigation.js";
import { resolvePermissionDeniedRedirect } from "./admin-route-guard.js";

test("resolvePermissionDeniedRedirect skips redirecting back to the denied route", () => {
  assert.equal(
    resolvePermissionDeniedRedirect({
      deniedPath: "/dashboard",
      firstAccessiblePath: "/dashboard"
    }),
    NO_ACCESS_ROUTE_PATH
  );
});

test("resolvePermissionDeniedRedirect prefers the first accessible route when it differs", () => {
  assert.equal(
    resolvePermissionDeniedRedirect({
      deniedPath: "/dashboard",
      firstAccessiblePath: "/users"
    }),
    "/users"
  );
});

test("resolvePermissionDeniedRedirect falls back to no-access when no accessible route exists", () => {
  assert.equal(
    resolvePermissionDeniedRedirect({
      deniedPath: "/dashboard",
      firstAccessiblePath: ""
    }),
    NO_ACCESS_ROUTE_PATH
  );
});
