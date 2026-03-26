import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "DataSourcesView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("DataSourcesView acts as nested shell for data source subroutes", () => {
  const text = readView();
  assert.match(text, /useDataSourcesWorkspace/);
  assert.match(text, /<router-view/);
  assert.match(text, /data-sources-nav/);
  assert.match(text, /workspace\.sectionItems/);
  assert.match(text, /workspace\.buildSectionLocation/);
  assert.doesNotMatch(text, /<ProviderGovernanceOverviewPanel/);
});
