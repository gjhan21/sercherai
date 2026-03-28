import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const routerPath = path.join(__dirname, "index.js");

test("h5 router redirects watchlist through profile section", () => {
  const text = fs.readFileSync(routerPath, "utf8");
  assert.match(text, /import \{ buildProfileModuleRoute \} from "\.\.\/\.\.\/\.\.\/lib\/profile-modules"/);
  assert.match(text, /path: "watchlist"[\s\S]*redirect: buildProfileModuleRoute\("watchlist"\)/);
  assert.doesNotMatch(text, /path: "watchlist", name: "h5-watchlist", component:/);
});
