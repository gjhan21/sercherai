export const H5_BASE_PATH = "/m/";

export function shouldServeH5Index(rawUrl = "/") {
  const pathname = String(rawUrl || "/").split("?")[0] || "/";
  return pathname === "/m" || pathname.startsWith("/m/");
}

export function normalizeRedirectPath(rawPath = "/home") {
  const value = String(rawPath || "").trim();
  if (!value) {
    return "/home";
  }
  if (/^https?:\/\//i.test(value) || /^javascript:/i.test(value)) {
    return "/home";
  }
  if (value.startsWith("/")) {
    return value;
  }
  return `/${value}`;
}
