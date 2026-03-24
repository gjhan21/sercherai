import http from "../lib/http";

export function register(payload) {
  return http.post("/auth/register", payload);
}

export function login(payload) {
  return http.post("/auth/login", payload);
}

export function refreshToken(refreshToken) {
  return http.post("/auth/refresh", { refresh_token: refreshToken });
}

export function logout(refreshToken) {
  return http.post("/auth/logout", { refresh_token: refreshToken });
}

export function logoutAll() {
  return http.post("/auth/logout-all");
}

export function getAuthProfile() {
  return http.get("/auth/me");
}
