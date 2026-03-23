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

test("DataSourcesView keeps futures inventory sync in the unified sync console", () => {
  const text = readView();
  assert.match(text, /syncFuturesInventory/);
  assert.match(text, /const syncingFuturesInventory = ref\(false\);/);
  assert.match(text, /const futuresInventoryLastSyncResult = ref\(null\);/);
  assert.match(text, /const futuresInventorySyncForm = reactive\(\{/);
  assert.match(text, /case "futures_inventory":/);
  assert.match(text, /return \["TUSHARE", "MOCK"\]\.includes\(provider\);/);
  assert.match(text, /const futuresInventorySyncOptions = computed\(\(\) => buildSyncSourceOptions\("futures_inventory"\)\);/);
  assert.match(text, /async function handleSyncFuturesInventory\(\)/);
  assert.match(text, /期货仓单同步/);
  assert.match(text, /同步期货仓单/);
  assert.match(text, /futuresInventorySyncOptions/);
});

test("DataSourcesView exposes provider governance sections and routing policy editing", () => {
  const text = readView();
  assert.match(text, /getMarketProviderGovernanceOverview/);
  assert.match(text, /listMarketProviderCapabilities/);
  assert.match(text, /listMarketProviderRoutingPolicies/);
  assert.match(text, /updateMarketProviderRoutingPolicy/);
  assert.match(text, /const governanceOverviewStock = ref\(null\);/);
  assert.match(text, /const governanceOverviewFutures = ref\(null\);/);
  assert.match(text, /const providerCapabilities = ref\(\[\]\);/);
  assert.match(text, /const routingPolicies = ref\(\[\]\);/);
  assert.match(text, /const routingPolicyDialogVisible = ref\(false\);/);
  assert.match(text, /async function fetchGovernanceDashboard/);
  assert.match(text, /async function handleEditRoutingPolicy/);
  assert.match(text, /async function handleSaveRoutingPolicy/);
  assert.match(text, /供应商治理总览/);
  assert.match(text, /能力矩阵/);
  assert.match(text, /路由与真相源治理/);
  assert.match(text, /治理建议/);
});
