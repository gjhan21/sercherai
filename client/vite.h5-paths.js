const STATIC_EXT_RE = /\.[a-z0-9]+$/i;

export function shouldServeH5Index(rawUrl = "") {
  const value = String(rawUrl || "");
  const pathname = value.split("?")[0] || "";

  if (!pathname.startsWith("/m/")) {
    return false;
  }

  if (
    pathname === "/m/index.html" ||
    pathname.startsWith("/m/src/") ||
    pathname.startsWith("/m/node_modules/") ||
    pathname.startsWith("/m/@") ||
    pathname.startsWith("/m/assets/") ||
    STATIC_EXT_RE.test(pathname)
  ) {
    return false;
  }

  return true;
}
