import test from "node:test";
import assert from "node:assert/strict";

import {
  H5_SESSION_KEY,
  clearClientSession,
  getAccessToken,
  getClientSession,
  getRefreshToken,
  hasClientSession,
  saveClientSession
} from "./session.js";

function createStorage() {
  const bucket = new Map();
  return {
    getItem(key) {
      return bucket.has(key) ? bucket.get(key) : null;
    },
    setItem(key, value) {
      bucket.set(key, String(value));
    },
    removeItem(key) {
      bucket.delete(key);
    }
  };
}

test("saveClientSession normalizes auth payload and exposes token helpers", () => {
  globalThis.localStorage = createStorage();

  saveClientSession({
    access_token: "access-demo",
    refresh_token: "refresh-demo",
    user_id: "u_demo_001",
    phone: "13800000000"
  });

  assert.equal(getAccessToken(), "access-demo");
  assert.equal(getRefreshToken(), "refresh-demo");
  assert.equal(hasClientSession(), true);
  assert.deepEqual(getClientSession(), {
    accessToken: "access-demo",
    refreshToken: "refresh-demo",
    tokenType: "Bearer",
    userID: "u_demo_001",
    phone: "13800000000",
    email: "",
    role: "USER",
    expiresIn: 0
  });
});

test("clearClientSession removes the H5 scoped session key", () => {
  globalThis.localStorage = createStorage();

  saveClientSession({ access_token: "access-demo" });
  clearClientSession();

  assert.equal(globalThis.localStorage.getItem(H5_SESSION_KEY), null);
  assert.equal(getAccessToken(), "");
  assert.equal(hasClientSession(), false);
});
