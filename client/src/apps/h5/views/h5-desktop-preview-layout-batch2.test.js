import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

function read(name) {
  return fs.readFileSync(path.join(__dirname, name), "utf8");
}

test("h5 membership desktop preview has dedicated grid rules", () => {
  const text = read("H5MembershipView.vue");
  assert.match(text, /@media \(min-width: 521px\)/);
  assert.match(text, /\.membership-plan-list\s*\{[\s\S]*grid-template-columns:/);
  assert.match(text, /\.membership-action-buttons\s*\{[\s\S]*grid-template-columns:/);
});

test("h5 profile desktop preview has dedicated grid rules", () => {
  const text = read("H5ProfileView.vue");
  assert.match(text, /@media \(min-width: 521px\)/);
  assert.match(text, /\.profile-shortcut-grid\s*\{[\s\S]*grid-template-columns:/);
  assert.match(text, /\.profile-module-grid\s*\{[\s\S]*grid-template-columns:/);
});

test("h5 watchlist desktop preview has dedicated grid rules", () => {
  const text = read("H5WatchlistView.vue");
  assert.match(text, /@media \(min-width: 521px\)/);
  assert.match(text, /\.watchlist-list\s*\{[\s\S]*grid-template-columns:/);
  assert.match(text, /\.watchlist-item-actions\s*\{[\s\S]*grid-template-columns:/);
});
