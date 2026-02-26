import http from "../lib/http";

export function login(payload) {
  return http.post("/auth/login", payload);
}

export function mockLogin(payload) {
  return http.post("/auth/mock-login", payload);
}

export function me() {
  return http.get("/auth/me");
}
