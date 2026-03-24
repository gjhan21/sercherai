import http from "../lib/http.js";
import { buildParams } from "../lib/request.js";
import { getAccessToken } from "../lib/session.js";

export function listNewsCategories() {
  return http.get("/public/news/categories");
}

export function listNewsArticles(params) {
  return http.get("/public/news/articles", { params: buildParams(params) });
}

export function resolveArticlePath(id, accessToken = getAccessToken()) {
  const articleID = encodeURIComponent(id);
  if (accessToken) {
    return `/news/articles/${articleID}`;
  }
  return `/public/news/articles/${articleID}`;
}

export function resolveAttachmentPath(articleID, accessToken = getAccessToken()) {
  const id = encodeURIComponent(articleID);
  if (accessToken) {
    return `/news/articles/${id}/attachments`;
  }
  return `/public/news/articles/${id}/attachments`;
}

export function getNewsArticleDetail(id) {
  return http.get(resolveArticlePath(id));
}

export function listNewsAttachments(articleID) {
  return http.get(resolveAttachmentPath(articleID));
}

export function getAttachmentSignedURL(attachmentID) {
  return http.get(`/news/attachments/${encodeURIComponent(attachmentID)}/signed-url`);
}
