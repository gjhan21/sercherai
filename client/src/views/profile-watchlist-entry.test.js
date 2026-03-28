import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const profileViewPath = path.join(__dirname, "ProfileView.vue");

test("profile view surfaces a dedicated watchlist entry card", () => {
  const text = fs.readFileSync(profileViewPath, "utf8");
  assert.match(text, /我的二级模块/);
  assert.match(text, /我的讨论/);
  assert.match(text, /profile-module-grid/);
  assert.match(text, /profile-watchlist-card/);
  assert.match(text, /我的关注/);
  assert.match(text, /个人跟踪清单/);
  assert.match(text, /进入我的关注/);
  assert.match(text, /\/profile\/watchlist/);
});
