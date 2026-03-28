import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const routerPath = path.join(__dirname, "index.js");
const layoutPath = path.join(__dirname, "..", "layouts", "PcLayout.vue");

test("pc router exposes community and profile watchlist routes", () => {
  const text = fs.readFileSync(routerPath, "utf8");
  assert.match(text, /const CommunityView =/);
  assert.match(text, /const CommunityTopicView =/);
  assert.match(text, /const CommunityComposeView =/);
  assert.match(text, /path: "community"/);
  assert.match(text, /path: "community\/topics\/:id"/);
  assert.match(text, /path: "community\/new"/);
  assert.match(text, /path: "profile\/watchlist"/);
  assert.match(text, /name: "pc-profile-watchlist"/);
  assert.match(text, /import \{ buildProfileModuleRoute \} from "\.\.\/\.\.\/\.\.\/lib\/profile-modules"/);
  assert.match(text, /path: "watchlist"[\s\S]*redirect: buildProfileModuleRoute\("watchlist"\)/);
});

test("pc nav promotes community and stops exposing watchlist as top-level tab", () => {
  const text = fs.readFileSync(layoutPath, "utf8");
  assert.match(text, /path: "\/community", label: "社区"/);
  assert.doesNotMatch(text, /{ path: "\/watchlist", label: "关注" }/);
  assert.match(text, /isTabActive/);
  assert.match(text, /item\.match/);
});
