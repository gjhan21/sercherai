const STORAGE_KEY = "sercherai_admin_session";
const ROLE_LABELS = {
  ADMIN: "管理员",
  USER: "普通用户",
  CONTENT_EDITOR: "内容编辑"
};

export function getSession() {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function hasSession() {
  const session = getSession();
  return !!session?.accessToken;
}

export function getPermissionCodes() {
  return getSession()?.permissionCodes || [];
}

export function hasPermission(code) {
  if (!code) {
    return true;
  }
  const codes = getPermissionCodes();
  if (!codes.length) {
    return false;
  }
  if (codes.includes("*")) {
    return true;
  }
  return codes.includes(code);
}

export function getAccessToken() {
  return getSession()?.accessToken || "";
}

export function formatSessionRole(value) {
  const normalized = String(value || "")
    .trim()
    .toUpperCase();
  if (!normalized) {
    return "管理员";
  }
  return ROLE_LABELS[normalized] || value;
}

export function saveSession(payload) {
  const permissionCodes = Array.isArray(payload.permission_codes)
    ? payload.permission_codes.filter(Boolean)
    : [];
  const roles = Array.isArray(payload.roles) ? payload.roles : [];
  const session = {
    accessToken: payload.access_token || "",
    tokenType: payload.token_type || "Bearer",
    userID: payload.user_id || "",
    role: payload.role || "",
    expiresIn: payload.expires_in || 0,
    permissionCodes,
    roles
  };
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
  return session;
}

export function clearSession() {
  localStorage.removeItem(STORAGE_KEY);
}
