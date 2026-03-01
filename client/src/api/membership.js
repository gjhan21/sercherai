import http from "../lib/http";

function buildParams(params = {}) {
  const result = {};
  Object.keys(params).forEach((key) => {
    const value = params[key];
    if (value !== undefined && value !== null && value !== "") {
      result[key] = value;
    }
  });
  return result;
}

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

export function triggerPaymentCallback(channel, payload) {
  return http.post(`/payment/callbacks/${encodeURIComponent(channel)}`, payload);
}
