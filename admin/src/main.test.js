import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "main.js");

test("admin entry imports Element Plus global styles for service overlays", () => {
  const text = fs.readFileSync(filePath, "utf8");

  assert.match(text, /import\s+["']element-plus\/dist\/index\.css["'];/);
});
