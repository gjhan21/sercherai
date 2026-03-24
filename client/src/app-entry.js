export function resolveClientAppMode(rawPath = "/") {
  const value = String(rawPath || "/");
  const pathname = value.split("?")[0] || "/";
  return pathname === "/m" || pathname.startsWith("/m/") ? "h5" : "pc";
}
