import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const viewPath = path.join(__dirname, "H5ProfileView.vue");

test("h5 profile view exposes watchlist and discussion as profile modules", () => {
  const text = fs.readFileSync(viewPath, "utf8");
  assert.match(text, /我的二级模块/);
  assert.match(text, /我的关注/);
  assert.match(text, /我的讨论/);
  assert.match(text, /profile-module-grid|profile-module-list/);
});
