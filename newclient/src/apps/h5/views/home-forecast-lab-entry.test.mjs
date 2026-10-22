import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const filePath = new URL("./Home.vue", import.meta.url).pathname;

test("h5 home exposes forecast-lab quick entry", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /forecast-lab/);
  assert.match(text, /深度推演/);
});
