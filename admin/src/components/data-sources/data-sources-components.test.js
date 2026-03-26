import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function readComponent(name) {
  return fs.readFileSync(path.join(__dirname, name), "utf8");
}

test("RoutingPolicyPanel keeps routing policy dialog and emits save hooks", () => {
  const text = readComponent("RoutingPolicyPanel.vue");
  assert.match(text, /路由与真相源治理/);
  assert.match(text, /defineEmits/);
  assert.match(text, /保存策略/);
});

test("DataSourceRegistryTable keeps registry toolbar and CRUD affordances", () => {
  const text = readComponent("DataSourceRegistryTable.vue");
  assert.match(text, /新增数据源/);
  assert.match(text, /全部健康检查/);
  assert.match(text, /defineProps/);
  assert.match(text, /defineEmits/);
});

test("QualityLogsPanel keeps filter toolbar and payload dialog", () => {
  const text = readComponent("QualityLogsPanel.vue");
  assert.match(text, /市场数据质量日志/);
  assert.match(text, /查看 payload/);
  assert.match(text, /payload 详情/);
});

test("MarketSyncConsole uses dedicated large-card layout for sync forms", () => {
  const text = readComponent("MarketSyncConsole.vue");
  assert.match(text, /market-sync-grid/);
  assert.match(text, /market-sync-card/);
  assert.match(text, /card\.actions/);
  assert.match(text, /执行日志/);
  assert.match(text, /card\.logs/);
  assert.doesNotMatch(text, /class="truth-summary-grid"/);
});
