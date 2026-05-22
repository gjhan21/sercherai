import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "DataSourcesGovernanceSectionView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("DataSourcesGovernanceSectionView links operators to the sync workspace", () => {
  const text = readView();
  assert.match(text, /buildSectionLocation\('sync'\)/);
  assert.match(text, /前往同步台/);
});
