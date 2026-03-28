import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function read(file) {
  return fs.readFileSync(path.join(__dirname, file), "utf8");
}

test("pc watchlist entry is routed through profile and watchlist copy is demoted", () => {
  const watchlistText = read("MyWatchlistView.vue");
  const membershipText = read("MembershipView.vue");
  const homeText = read("HomeView.vue");
  const strategyText = read("StrategyView.vue");
  const newsText = read("NewsView.vue");
  const archiveText = read("RecommendationArchiveView.vue");
  const authText = read("AuthView.vue");

  assert.doesNotMatch(watchlistText, /关注页/);
  assert.match(watchlistText, /我的 > 我的关注/);
  assert.match(watchlistText, /\/profile\/watchlist/);

  assert.match(membershipText, /buildProfileModulePath\("watchlist"\)/);
  assert.match(homeText, /buildProfileModuleRoute\("watchlist"\)/);
  assert.match(homeText, /buildProfileModuleRedirectPath\("watchlist"\)/);
  assert.match(strategyText, /buildProfileModulePath\("watchlist"\)/);
  assert.match(newsText, /buildProfileModuleRoute\("watchlist"\)/);
  assert.match(archiveText, /buildProfileModuleRoute\("watchlist"\)/);

  assert.doesNotMatch(membershipText, /\/profile\/watchlist/);
  assert.doesNotMatch(homeText, /\/profile\/watchlist/);
  assert.doesNotMatch(strategyText, /\/profile\/watchlist/);
  assert.doesNotMatch(newsText, /\/profile\/watchlist/);
  assert.doesNotMatch(archiveText, /\/profile\/watchlist/);

  assert.match(authText, /const watchlistRedirectPath = buildProfileModuleRedirectPath\("watchlist"\)/);
  assert.doesNotMatch(authText, /path: "\/profile\/watchlist"/);
});
