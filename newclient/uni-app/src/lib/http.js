import axios from "axios";

import { clearClientAuthState, setClientAuthState } from "./auth-state.js";
import { getAccessToken, getRefreshToken } from "./session.js";

const baseURL = import.meta?.env?.VITE_API_BASE_URL || "/api/v1";

const http = axios.create({
  baseURL,
  timeout: 12000
});

const refreshClient = axios.create({
  baseURL,
  timeout: 12000
});

let refreshPromise = null;

function shouldRefreshToken(error) {
  const status = error?.response?.status;
  if (status !== 401) {
    return false;
  }
  const originalConfig = error?.config || {};
  if (originalConfig._retry) {
    return false;
  }
  const url = String(originalConfig.url || "");
  if (
    url.includes("/auth/login") ||
    url.includes("/auth/register") ||
    url.includes("/auth/refresh") ||
    url.includes("/auth/logout")
  ) {
    return false;
  }
  return Boolean(getRefreshToken());
}

async function refreshSession() {
  if (refreshPromise) {
    return refreshPromise;
  }
  refreshPromise = refreshClient
    .post("/auth/refresh", { refresh_token: getRefreshToken() })
    .then((response) => {
      const payload = response?.data;
      if (!payload || payload.code !== 0 || !payload.data) {
        throw new Error(payload?.message || "刷新登录状态失败");
      }
      setClientAuthState(payload.data);
      return payload.data;
    })
    .finally(() => {
      refreshPromise = null;
    });
  return refreshPromise;
}

http.interceptors.request.use((config) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

http.interceptors.response.use(
  (response) => {
    const payload = response.data;
    if (payload && typeof payload.code !== "undefined") {
      if (payload.code === 0) {
        return payload.data;
      }
      const err = new Error(payload.message || "请求失败");
      err.code = payload.code;
      err.payload = payload.data;
      throw err;
    }
    return payload;
  },
  async (error) => {
    if (shouldRefreshToken(error)) {
      const originalConfig = error.config || {};
      originalConfig._retry = true;
      try {
        await refreshSession();
        const nextToken = getAccessToken();
        if (nextToken) {
          originalConfig.headers = originalConfig.headers || {};
          originalConfig.headers.Authorization = `Bearer ${nextToken}`;
        }
        return http(originalConfig);
      } catch {
        clearClientAuthState();
      }
    }
    const message = error?.response?.data?.message || error?.message || "网络请求失败";
    const err = new Error(message);
    err.code = error?.response?.data?.code || error?.response?.status || error?.code;
    err.original = error;
    throw err;
  }
);

export default http;
