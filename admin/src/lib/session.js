const STORAGE_KEY = "sercherai_admin_session";

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

export function getAccessToken() {
  return getSession()?.accessToken || "";
}

export function saveSession(payload) {
  const session = {
    accessToken: payload.access_token || "",
    tokenType: payload.token_type || "Bearer",
    userID: payload.user_id || "",
    role: payload.role || "",
    expiresIn: payload.expires_in || 0
  };
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
  return session;
}

export function clearSession() {
  localStorage.removeItem(STORAGE_KEY);
}
