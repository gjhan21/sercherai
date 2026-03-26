import http from "../lib/http.js";

export function register(payload) {
  return http.post("/auth/register", payload);
}

export function login(payload) {
  return http.post("/auth/login", payload);
}

export function refreshToken(refreshTokenValue) {
  return http.post("/auth/refresh", { refresh_token: refreshTokenValue });
}

export function logout(refreshTokenValue) {
  return http.post("/auth/logout", { refresh_token: refreshTokenValue });
}

export function getAuthProfile() {
  return http.get("/auth/me");
}
