import { NO_ACCESS_ROUTE_PATH } from "./admin-navigation.js";

export function resolvePermissionDeniedRedirect({ deniedPath, firstAccessiblePath }) {
  const denied = String(deniedPath || "").trim();
  const target = String(firstAccessiblePath || "").trim();

  if (!target || target === denied) {
    return NO_ACCESS_ROUTE_PATH;
  }

  return target;
}
