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

test("h5 shared shell and auth page consume the same width token", () => {
  const shellCss = read(shellCssPath);
  const uiCss = read(uiCssPath);
  const authView = read(authViewPath);

  assert.match(uiCss, /--h5-content-width:\s*min\(100%,\s*560px\);/);
  assert.match(shellCss, /width:\s*var\(--h5-content-width\);/);
  assert.match(authView, /width:\s*var\(--h5-content-width\);/);
});
