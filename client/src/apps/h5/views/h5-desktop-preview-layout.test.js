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

test("h5 auth desktop preview has a dedicated wider-layout treatment", () => {
  const text = read("H5AuthView.vue");
  assert.match(text, /@media \(min-width: 521px\)/);
  assert.match(text, /\.h5-auth-main-card\s*\{[\s\S]*grid-template-columns:/);
  assert.match(text, /\.h5-auth-scene-cards\s*\{[\s\S]*grid-template-columns:/);
});

test("h5 strategy desktop preview has dedicated action layout rules", () => {
  const text = read("H5StrategyView.vue");
  assert.match(text, /@media \(min-width: 521px\)/);
  assert.match(text, /\.strategy-actions-card\s*\{[\s\S]*grid-template-columns:/);
  assert.match(text, /\.strategy-kind-strip\s*\{[\s\S]*flex-wrap:/);
});

test("h5 news desktop preview has dedicated wider action and filter layout rules", () => {
  const text = read("H5NewsView.vue");
  assert.match(text, /@media \(min-width: 521px\)/);
  assert.match(text, /\.news-feed-actions\s*\{[\s\S]*grid-template-columns:/);
  assert.match(text, /\.news-category-strip\s*\{[\s\S]*flex-wrap:/);
});
