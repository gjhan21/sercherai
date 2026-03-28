import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const routerPath = path.join(__dirname, "index.js");
const layoutPath = path.join(__dirname, "..", "layouts", "PcLayout.vue");

test("pc router registers dedicated /search route", () => {
  const text = fs.readFileSync(routerPath, "utf8");
  assert.match(text, /const SearchView =/);
  assert.match(text, /path: "search"/);
  assert.match(text, /name: "pc-search"/);
});

test("pc layout renders nav search controls beside account actions", () => {
  const text = fs.readFileSync(layoutPath, "utf8");
  assert.match(text, /pc-search/);
  assert.match(text, /searchKeyword/);
  assert.match(text, /查看更多搜索结果/);
  assert.match(text, /searchDropdownVisible/);
  assert.match(text, /focus_type/);
  assert.match(text, /focus_id/);
  assert.match(text, /登录 \/ 注册/);
  assert.doesNotMatch(text, /currentTabLabel/);
  assert.doesNotMatch(text, /退出中\.\.\./);
});
