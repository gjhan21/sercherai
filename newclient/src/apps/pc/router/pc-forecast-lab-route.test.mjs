import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const filePath = new URL("./index.js", import.meta.url).pathname;

test("pc router registers forecast-lab route", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /path:\s*\"\/forecast-lab\"/);
  assert.match(text, /name:\s*\"forecast-lab\"/);
});
