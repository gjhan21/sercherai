import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const routerPath = path.join(__dirname, "index.js");
const layoutPath = path.join(__dirname, "..", "layouts", "H5Layout.vue");

test("h5 router registers dedicated /search route", () => {
  const text = fs.readFileSync(routerPath, "utf8");
  assert.match(text, /const SearchView =/);
  assert.match(text, /path: "search"/);
  assert.match(text, /name: "h5-search"/);
});

test("h5 layout renders fixed search bar for home and search pages", () => {
  const text = fs.readFileSync(layoutPath, "utf8");
  assert.match(text, /h5-search-shell/);
  assert.match(text, /showGlobalSearchBar/);
  assert.match(text, /searchDropdownVisible/);
  assert.match(text, /查看更多搜索结果/);
  assert.match(text, /route\.path === "\/home" \|\| route\.path === "\/search"/);
  assert.match(text, /searchKeyword/);
});
