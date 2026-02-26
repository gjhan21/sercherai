import axios from "axios";
import { clearSession, getAccessToken } from "./session";

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "/api/v1",
  timeout: 10000
});

function normalizeErrorMessage(error) {
  if (error?.response?.data?.message) {
    return error.response.data.message;
  }
  if (error?.message) {
    return error.message;
  }
  return "请求失败";
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
      const businessError = new Error(payload.message || "业务错误");
      businessError.code = payload.code;
      businessError.payload = payload.data;
      throw businessError;
    }
    return payload;
  },
  (error) => {
    if (error?.response?.status === 401) {
      clearSession();
      if (window.location.pathname !== "/login") {
        window.location.href = "/login";
      }
    }
    const normalized = new Error(normalizeErrorMessage(error));
    normalized.code = error?.response?.data?.code || error?.code;
    normalized.original = error;
    throw normalized;
  }
);

export default http;
