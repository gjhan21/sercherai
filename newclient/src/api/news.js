import http from "../shared/lib/http";
import { getAccessToken } from "../shared/auth/session";

function buildParams(params = {}) {
  const result = {};
  Object.keys(params).forEach((key) => {
    const value = params[key];
    if (value !== undefined && value !== null && value !== "") result[key] = value;
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
  return getAccessToken() ? `/news/articles/${articleID}` : `/public/news/articles/${articleID}`;
}

function resolveAttachmentPath(articleID) {
  const id = encodeURIComponent(articleID);
  return getAccessToken() ? `/news/articles/${id}/attachments` : `/public/news/articles/${id}/attachments`;
}
