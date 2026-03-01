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

export function listNewsCategories() {
  return http.get("/public/news/categories");
}

export function listNewsArticles(params) {
  return http.get("/public/news/articles", { params: buildParams(params) });
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

function resolveArticlePath(id) {
  const articleID = encodeURIComponent(id);
  if (getAccessToken()) {
    return `/news/articles/${articleID}`;
  }
  return `/public/news/articles/${articleID}`;
}

function resolveAttachmentPath(articleID) {
  const id = encodeURIComponent(articleID);
  if (getAccessToken()) {
    return `/news/articles/${id}/attachments`;
  }
  return `/public/news/articles/${id}/attachments`;
}
