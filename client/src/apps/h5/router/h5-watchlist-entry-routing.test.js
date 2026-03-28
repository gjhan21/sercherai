import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const viewsDir = path.join(__dirname, "..", "views");
const libDir = path.join(__dirname, "..", "lib");

function readView(name) {
  return fs.readFileSync(path.join(viewsDir, name), "utf8");
}

function readLib(name) {
  return fs.readFileSync(path.join(libDir, name), "utf8");
}

test("h5 cross-page watchlist entry routes through profile section", () => {
  const homeText = readView("H5HomeView.vue");
  const strategyText = readView("H5StrategyView.vue");
  const archiveText = readView("H5ArchiveView.vue");
  const authHelperText = readLib("auth-page.js");

  assert.match(homeText, /buildProfileModuleRoute\("watchlist"\)|buildProfileModulePath\("watchlist"\)|buildProfileModuleRedirectPath\("watchlist"\)/);
  assert.match(strategyText, /buildProfileModuleRoute\("watchlist"\)|buildProfileModulePath\("watchlist"\)|buildProfileModuleRedirectPath\("watchlist"\)/);
  assert.match(archiveText, /buildProfileModuleRoute\("watchlist"\)|buildProfileModulePath\("watchlist"\)|buildProfileModuleRedirectPath\("watchlist"\)/);
  assert.match(authHelperText, /buildProfileModuleRedirectPath\("watchlist"\)/);

  assert.doesNotMatch(homeText, /router\.push\("\/watchlist"\)/);
  assert.doesNotMatch(strategyText, /router\.push\("\/watchlist"\)/);
  assert.doesNotMatch(archiveText, /router\.push\("\/watchlist"\)/);
});
