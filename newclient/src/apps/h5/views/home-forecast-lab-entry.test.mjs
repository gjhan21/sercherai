import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const filePath = "/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/Home.vue";

test("h5 home exposes forecast-lab quick entry", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /forecast-lab/);
  assert.match(text, /深度推演/);
});
