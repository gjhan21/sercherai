import { formatDateTime } from "../lib/format.js";

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

export function mapNewsCategory(item = {}) {
  return {
    id: item.id || "",
    name: item.name || "未分类"
  };
}

export function mapNewsArticle(item = {}) {
  return {
    id: item.id || "",
    categoryID: item.category_id || "",
    title: item.title || "未命名内容",
    summary: item.summary || "暂无摘要",
    content: item.content || "",
    visibility: item.visibility || "PUBLIC",
    publishedAt: formatDateTime(item.published_at),
    attachmentCount: item.attachment_count || 0,
    coverURL: item.cover_url || ""
  };
}

export function buildNewsDetailView(article = {}, attachments = []) {
  const item = mapNewsArticle(article);
  return {
    ...item,
    attachments: toArray(attachments).map((attachment) => ({
      id: attachment.id || "",
      fileName: attachment.file_name || "未命名附件",
      fileSize: attachment.file_size || 0,
      createdAt: formatDateTime(attachment.created_at)
    }))
  };
}
