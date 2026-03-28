import test from "node:test";
import assert from "node:assert/strict";
import { resolveProxyTarget } from "./vite.proxy-target.js";

test("resolveProxyTarget prefers explicit VITE_PROXY_TARGET", () => {
  const previousProxy = process.env.VITE_PROXY_TARGET;
  const previousAppPort = process.env.APP_PORT;
  const previousBackendPort = process.env.BACKEND_PORT;

  process.env.VITE_PROXY_TARGET = "http://127.0.0.1:19999";
  delete process.env.APP_PORT;
  delete process.env.BACKEND_PORT;

  assert.equal(resolveProxyTarget(), "http://127.0.0.1:19999");

  restoreEnv(previousProxy, previousAppPort, previousBackendPort);
});

test("resolveProxyTarget falls back to APP_PORT when proxy target is absent", () => {
  const previousProxy = process.env.VITE_PROXY_TARGET;
  const previousAppPort = process.env.APP_PORT;
  const previousBackendPort = process.env.BACKEND_PORT;

  delete process.env.VITE_PROXY_TARGET;
  process.env.APP_PORT = "19123";
  delete process.env.BACKEND_PORT;

  assert.equal(resolveProxyTarget(), "http://127.0.0.1:19123");

  restoreEnv(previousProxy, previousAppPort, previousBackendPort);
});

test("resolveProxyTarget defaults to local backend port 19081", () => {
  const previousProxy = process.env.VITE_PROXY_TARGET;
  const previousAppPort = process.env.APP_PORT;
  const previousBackendPort = process.env.BACKEND_PORT;

  delete process.env.VITE_PROXY_TARGET;
  delete process.env.APP_PORT;
  delete process.env.BACKEND_PORT;

  assert.equal(resolveProxyTarget(), "http://127.0.0.1:19081");

  restoreEnv(previousProxy, previousAppPort, previousBackendPort);
});

function restoreEnv(proxy, appPort, backendPort) {
  setOrDeleteEnv("VITE_PROXY_TARGET", proxy);
  setOrDeleteEnv("APP_PORT", appPort);
  setOrDeleteEnv("BACKEND_PORT", backendPort);
}

function setOrDeleteEnv(key, value) {
  if (typeof value === "string") {
    process.env[key] = value;
    return;
  }
  delete process.env[key];
}
