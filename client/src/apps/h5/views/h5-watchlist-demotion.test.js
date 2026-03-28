import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const watchlistViewPath = path.join(__dirname, "H5WatchlistView.vue");
const homeViewPath = path.join(__dirname, "H5HomeView.vue");
const strategyViewPath = path.join(__dirname, "H5StrategyView.vue");
const archiveViewPath = path.join(__dirname, "H5ArchiveView.vue");
const shellMetaPath = path.join(__dirname, "..", "lib", "shell-meta.js");

test("h5 watchlist is framed as profile detail instead of primary channel", () => {
  const watchlistText = fs.readFileSync(watchlistViewPath, "utf8");
  const homeText = fs.readFileSync(homeViewPath, "utf8");
  const strategyText = fs.readFileSync(strategyViewPath, "utf8");
  const archiveText = fs.readFileSync(archiveViewPath, "utf8");
  const shellMetaText = fs.readFileSync(shellMetaPath, "utf8");

  assert.match(watchlistText, /我的 > 我的关注详情|我的 > 我的关注/);
  assert.match(shellMetaText, /我的关注详情|我的二级模块|从我的进入/);
  assert.doesNotMatch(shellMetaText, /section:\s*"关注"/);
  assert.doesNotMatch(homeText, /进入我的关注|去我的关注|关注页/);
  assert.doesNotMatch(strategyText, /去我的关注|关注页/);
  assert.doesNotMatch(archiveText, /去我的关注|关注页|回到我的关注/);
  assert.match(homeText, /我的\s*(>|&gt;)\s*我的关注|从我的进入/);
  assert.match(strategyText, /我的\s*(>|&gt;)\s*我的关注|从我的进入/);
  assert.match(archiveText, /我的\s*(>|&gt;)\s*我的关注|从我的进入/);
});
