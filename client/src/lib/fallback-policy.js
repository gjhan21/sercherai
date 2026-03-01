export function shouldUseDemoFallback() {
  if (!import.meta.env.DEV) {
    return false;
  }
  const raw = String(import.meta.env.VITE_ENABLE_CLIENT_DEMO_FALLBACK || "").trim().toLowerCase();
  return raw === "1" || raw === "true";
}
