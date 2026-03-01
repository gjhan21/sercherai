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

export function getUserProfile() {
  return http.get("/user/profile");
}

export function getMembershipQuota() {
  return http.get("/membership/quota");
}

export function listMembershipOrders(params) {
  return http.get("/membership/orders", { params: buildParams(params) });
}

export function listRechargeRecords(params) {
  return http.get("/user/recharge-records", { params: buildParams(params) });
}

export function listBrowseHistory(params) {
  return http.get("/user/browse-history", { params: buildParams(params) });
}

export function listMessages(params) {
  return http.get("/messages", { params: buildParams(params) });
}

export function listShareLinks() {
  return http.get("/user/share-links");
}

export function createShareLink(payload) {
  return http.post("/user/share-links", payload);
}

export function listInviteRecords(params) {
  return http.get("/user/share/invites", { params: buildParams(params) });
}

export function getInviteSummary() {
  return http.get("/user/share/invite-summary");
}

export function readMessage(id) {
  return http.put(`/messages/${encodeURIComponent(id)}/read`);
}

export function listSubscriptions(params) {
  return http.get("/subscriptions", { params: buildParams(params) });
}

export function createSubscription(payload) {
  return http.post("/subscriptions", payload);
}

export function updateSubscription(id, payload) {
  return http.put(`/subscriptions/${encodeURIComponent(id)}`, payload);
}
