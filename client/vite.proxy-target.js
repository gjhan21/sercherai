export function resolveProxyTarget() {
  const explicitTarget = normalizeURL(process.env.VITE_PROXY_TARGET);
  if (explicitTarget) {
    return explicitTarget;
  }

  const backendPort = normalizePort(process.env.APP_PORT) || normalizePort(process.env.BACKEND_PORT) || "19081";
  return `http://127.0.0.1:${backendPort}`;
}

function normalizeURL(value) {
  const normalized = String(value || "").trim();
  return normalized || "";
}

function normalizePort(value) {
  const normalized = String(value || "").trim();
  return /^\d+$/.test(normalized) ? normalized : "";
}
