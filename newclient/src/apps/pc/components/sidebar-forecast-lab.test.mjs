import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const filePath = "/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/components/SidebarNav.vue";

test("pc sidebar exposes forecast-lab as a first-class entry", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /to=\"\/forecast-lab\"/);
  assert.match(text, /深度推演/);
});
