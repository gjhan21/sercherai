import http from "../lib/http.js";
import { buildParams } from "../lib/request.js";

export function listMembershipProducts(params) {
  return http.get("/membership/products", { params: buildParams(params) });
}

export function listMembershipOrders(params) {
  return http.get("/membership/orders", { params: buildParams(params) });
}

export function getMembershipQuota() {
  return http.get("/membership/quota");
}

export function createMembershipOrder(payload) {
  return http.post("/membership/orders", payload);
}
