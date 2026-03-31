import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const communityViewPath = path.join(__dirname, "CommunityView.vue");

test("community view presents primary entry framing on pc", () => {
  const text = fs.readFileSync(communityViewPath, "utf8");
  assert.match(text, /社区主入口/);
  assert.match(text, /发布我的观点/);
  assert.match(text, /看我的主题/);
  assert.match(text, /从策略、资讯和我的讨论继续承接/);
  assert.match(text, /community-entry-grid/);
});
