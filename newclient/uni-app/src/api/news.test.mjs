import test from "node:test";
import assert from "node:assert/strict";

import { resolveArticlePath, resolveAttachmentPath } from "./news.js";

test("resolveArticlePath uses public path when visitor is not logged in", () => {
  assert.equal(resolveArticlePath("article-1", ""), "/public/news/articles/article-1");
});

test("resolveArticlePath uses private path when access token exists", () => {
  assert.equal(resolveArticlePath("article-1", "access-token"), "/news/articles/article-1");
});

test("resolveAttachmentPath follows the same public/private rule", () => {
  assert.equal(resolveAttachmentPath("article-1", ""), "/public/news/articles/article-1/attachments");
  assert.equal(resolveAttachmentPath("article-1", "access-token"), "/news/articles/article-1/attachments");
});
