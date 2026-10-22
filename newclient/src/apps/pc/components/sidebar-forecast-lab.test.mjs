import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const filePath = new URL("./SidebarNav.vue", import.meta.url).pathname;

test("pc sidebar exposes forecast-lab as a first-class entry", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /to=\"\/forecast-lab\"/);
  assert.match(text, /深度推演/);
});
