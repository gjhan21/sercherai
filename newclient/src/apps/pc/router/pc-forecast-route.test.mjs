import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const routerPath = path.join(__dirname, "index.js");

test("pc router registers dedicated forecast detail route", () => {
  const text = fs.readFileSync(routerPath, "utf8");
  assert.match(text, /path:\s*\"\/forecast\/:id\"/);
  assert.match(text, /name:\s*\"forecast-detail\"/);
});
