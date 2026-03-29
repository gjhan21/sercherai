import test from "node:test";
import assert from "node:assert/strict";

import { buildForecastLabRouteQuery, normalizeForecastLabRouteState } from "./forecast-lab.js";

test("normalizeForecastLabRouteState keeps only supported route filters", () => {
  const state = normalizeForecastLabRouteState({
    run_id: "run_123",
    target_type: "stock",
    status: "running",
    trigger_type: "user_request",
    user_id: "u_001",
    extra: "ignored"
  });

  assert.deepEqual(state, {
    runID: "run_123",
    targetType: "STOCK",
    status: "RUNNING",
    triggerType: "USER_REQUEST",
    userID: "u_001"
  });
});

test("buildForecastLabRouteQuery drops stale empty filters instead of keeping old query state", () => {
  const query = buildForecastLabRouteQuery({
    runID: "run_123",
    targetType: "",
    status: "",
    triggerType: "",
    userID: ""
  });

  assert.deepEqual(query, {
    run_id: "run_123"
  });
});

test("buildForecastLabRouteQuery serializes supported filters with stable keys", () => {
  const query = buildForecastLabRouteQuery({
    runID: "run_123",
    targetType: "FUTURES",
    status: "FAILED",
    triggerType: "ADMIN_MANUAL",
    userID: "admin_001"
  });

  assert.deepEqual(query, {
    run_id: "run_123",
    target_type: "FUTURES",
    status: "FAILED",
    trigger_type: "ADMIN_MANUAL",
    user_id: "admin_001"
  });
});
