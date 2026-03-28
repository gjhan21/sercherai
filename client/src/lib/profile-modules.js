const PROFILE_MODULE_SECTIONS = new Set(["overview", "watchlist", "community"]);

export function normalizeProfileModuleSection(value) {
  const normalized = String(value || "").trim().toLowerCase();
  if (!PROFILE_MODULE_SECTIONS.has(normalized)) {
    return "overview";
  }
  return normalized;
}

export function buildProfileModuleRoute(section = "overview") {
  const normalized = normalizeProfileModuleSection(section);
  if (normalized === "overview") {
    return { path: "/profile" };
  }
  return {
    path: "/profile",
    query: { section: normalized }
  };
}

export function buildProfileModulePath(section = "overview") {
  const route = buildProfileModuleRoute(section);
  if (!route.query?.section) {
    return route.path;
  }
  return `${route.path}?section=${encodeURIComponent(route.query.section)}`;
}

export function buildProfileModuleRedirectPath(section = "overview") {
  return buildProfileModulePath(section);
}
