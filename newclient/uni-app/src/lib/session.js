export const H5_SESSION_KEY = "sercherai_newclient_h5_session";

function getStorage() {
  if (
    typeof globalThis.localStorage !== "undefined" &&
    globalThis.localStorage &&
    typeof globalThis.localStorage.getItem === "function" &&
    typeof globalThis.localStorage.setItem === "function" &&
    typeof globalThis.localStorage.removeItem === "function"
  ) {
    return globalThis.localStorage;
  }
  return {
    getItem() {
      return null;
    },
    setItem() {},
    removeItem() {}
  };
}

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
  return parseSession(getStorage().getItem(H5_SESSION_KEY));
}

export function getAccessToken() {
  return getClientSession()?.accessToken || "";
}

export function getRefreshToken() {
  return getClientSession()?.refreshToken || "";
}

export function hasClientSession() {
  return Boolean(getAccessToken());
}

export function saveClientSession(payload = {}) {
  const previous = getClientSession() || {};
  const session = {
    accessToken: payload.access_token || payload.accessToken || previous.accessToken || "",
    refreshToken: payload.refresh_token || payload.refreshToken || previous.refreshToken || "",
    tokenType: payload.token_type || payload.tokenType || previous.tokenType || "Bearer",
    userID: payload.user_id || payload.userID || previous.userID || "",
    phone: payload.phone || previous.phone || "",
    email: payload.email || previous.email || "",
    role: payload.role || previous.role || "USER",
    expiresIn: payload.expires_in || payload.expiresIn || previous.expiresIn || 0
  };
  getStorage().setItem(H5_SESSION_KEY, JSON.stringify(session));
  return session;
}

export function clearClientSession() {
  getStorage().removeItem(H5_SESSION_KEY);
}
