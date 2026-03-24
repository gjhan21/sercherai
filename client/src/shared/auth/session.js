const CLIENT_STORAGE_KEY = "sercherai_client_session";

function parseSession(raw) {
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

export function getClientSession() {
  return parseSession(localStorage.getItem(CLIENT_STORAGE_KEY));
}

export function getAccessToken() {
  return getClientSession()?.accessToken || "";
}

export function getRefreshToken() {
  return getClientSession()?.refreshToken || "";
}

export function hasClientSession() {
  return Boolean(getClientSession()?.accessToken);
}

export function saveClientSession(payload) {
  const previous = getClientSession() || {};
  const session = {
    accessToken: payload.access_token || payload.accessToken || previous.accessToken || "",
    refreshToken: payload.refresh_token || payload.refreshToken || previous.refreshToken || "",
    tokenType: payload.token_type || payload.tokenType || "Bearer",
    userID: payload.user_id || payload.userID || previous.userID || "",
    phone: payload.phone || previous.phone || "",
    email: payload.email || previous.email || "",
    role: payload.role || previous.role || "USER",
    expiresIn: payload.expires_in || payload.expiresIn || previous.expiresIn || 0
  };
  localStorage.setItem(CLIENT_STORAGE_KEY, JSON.stringify(session));
  return session;
}

export function clearClientSession() {
  localStorage.removeItem(CLIENT_STORAGE_KEY);
}
