import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const consolePath = path.join(__dirname, "../../components/data-sources/MarketSyncConsole.vue");

function readConsoleView() {
  return fs.readFileSync(consolePath, "utf8");
}

test("MarketSyncConsole renders card actions provided by the sync workspace", () => {
  const text = readConsoleView();
  assert.match(text, /card\.actions/);
  assert.match(text, /action\.label/);
  assert.match(text, /任务中心/);
});
