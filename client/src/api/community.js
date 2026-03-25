import http from "../lib/http";
import { getAccessToken } from "../lib/session";

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

function resolveReadPath(path) {
  if (getAccessToken()) {
    return `/community${path}`;
  }
  return `/public/community${path}`;
}

export function listCommunityTopics(params) {
  return http.get(resolveReadPath("/topics"), { params: buildParams(params) });
}

export function listMyCommunityTopics(params) {
  return http.get("/community/me/topics", { params: buildParams(params) });
}

export function listMyCommunityComments(params) {
  return http.get("/community/me/comments", { params: buildParams(params) });
}

export function getCommunityTopicDetail(id) {
  return http.get(resolveReadPath(`/topics/${encodeURIComponent(id)}`));
}

export function listCommunityComments(topicID, params) {
  return http.get(resolveReadPath(`/topics/${encodeURIComponent(topicID)}/comments`), {
    params: buildParams(params)
  });
}

export function createCommunityTopic(payload) {
  return http.post("/community/topics", payload);
}

export function createCommunityComment(topicID, payload) {
  return http.post(`/community/topics/${encodeURIComponent(topicID)}/comments`, payload);
}

export function createCommunityReaction(payload) {
  return http.post("/community/reactions", payload);
}

export function deleteCommunityReaction(params) {
  return http.delete("/community/reactions", { params: buildParams(params) });
}

export function createCommunityReport(payload) {
  return http.post("/community/reports", payload);
}
