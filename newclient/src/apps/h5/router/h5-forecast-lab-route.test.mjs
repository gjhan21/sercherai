import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const filePath = "/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/router/index.js";

test("h5 router registers forecast-lab route", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /path:\s*\"\/forecast-lab\"/);
  assert.match(text, /name:\s*\"h5-forecast-lab\"/);
});
