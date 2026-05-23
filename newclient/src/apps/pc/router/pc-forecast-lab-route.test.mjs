import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const filePath = "/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/router/index.js";

test("pc router registers forecast-lab route", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /path:\s*\"\/forecast-lab\"/);
  assert.match(text, /name:\s*\"forecast-lab\"/);
});
