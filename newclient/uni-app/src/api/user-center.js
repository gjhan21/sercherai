import http from "../lib/http.js";
import { buildParams } from "../lib/request.js";

export function getUserProfile() {
  return http.get("/user/profile");
}

export function submitKYC(payload) {
  return http.post("/user/kyc/submit", payload);
}

export function listMessages(params) {
  return http.get("/messages", { params: buildParams(params) });
}

export function listBrowseHistory(params) {
  return http.get("/user/browse-history", { params: buildParams(params) });
}

export function listInviteRecords(params) {
  return http.get("/user/share/invites", { params: buildParams(params) });
}

export function getInviteSummary() {
  return http.get("/user/share/invite-summary");
}

export function listShareLinks() {
  return http.get("/user/share-links");
}

export function listSubscriptions(params) {
  return http.get("/subscriptions", { params: buildParams(params) });
}
