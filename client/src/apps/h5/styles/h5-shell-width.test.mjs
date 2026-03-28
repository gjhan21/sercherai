import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const shellCssPath = path.join(__dirname, "h5-shell.css");
const uiCssPath = path.join(__dirname, "h5-ui.css");
const authViewPath = path.join(__dirname, "..", "views", "H5AuthView.vue");

function read(filePath) {
  return fs.readFileSync(filePath, "utf8");
}

test("shared h5 shell width matches auth page width", () => {
  const shellCss = read(shellCssPath);
  const uiCss = read(uiCssPath);
  const authView = read(authViewPath);

  assert.match(authView, /width:\s*var\(--h5-content-width\);/);
  assert.match(shellCss, /\.h5-header-inner,\s*\.h5-main\s*\{[\s\S]*width:\s*var\(--h5-content-width\);/);
  assert.match(uiCss, /--h5-content-width:\s*min\(100%,\s*560px\);/);
});

test("fixed h5 chrome scales with the wider shell", () => {
  const shellCss = read(shellCssPath);
  const uiCss = read(uiCssPath);

  assert.match(shellCss, /\.h5-tabbar\s*\{[\s\S]*width:\s*min\(calc\(100% - 44px\),\s*504px\);/);
  assert.match(uiCss, /\.h5-sticky-cta\s*\{[\s\S]*width:\s*min\(calc\(100% - 44px\),\s*504px\);/);
});
