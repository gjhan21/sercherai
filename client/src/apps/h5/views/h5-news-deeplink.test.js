import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const viewPath = path.join(__dirname, "H5NewsView.vue");

test("h5 news view supports article_id deep links from search results", () => {
  const text = fs.readFileSync(viewPath, "utf8");
  assert.match(text, /resolveRequestedArticleID/);
  assert.match(text, /findRequestedArticleLocation/);
  assert.match(text, /route\.query\.article_id/);
  assert.match(text, /article_id: undefined/);
});
