import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const searchViewPath = path.join(__dirname, "..", "..", "..", "views", "SearchView.vue");

test("shared search view exposes dedicated h5 surface styling hook", () => {
  const text = fs.readFileSync(searchViewPath, "utf8");

  assert.match(text, /search-page-h5/);
  assert.match(text, /isH5SearchSurface/);
  assert.match(text, /route\.name/);
});

test("shared search view provides compact h5 overview and dense result card styling", () => {
  const text = fs.readFileSync(searchViewPath, "utf8");

  assert.match(text, /search-overview/);
  assert.match(text, /overviewItems/);
  assert.match(text, /search-overview-item/);
  assert.match(text, /search-page-h5 \.search-overview/);
  assert.match(text, /search-page-h5 \.search-result-card p/);
  assert.match(text, /-webkit-line-clamp: 2/);
});
