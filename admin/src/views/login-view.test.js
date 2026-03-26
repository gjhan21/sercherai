import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const filePath = path.join(__dirname, "LoginView.vue");

function readView() {
  return fs.readFileSync(filePath, "utf8");
}

test("LoginView avoids preset credentials in form defaults", () => {
  const text = readView();
  assert.match(text, /const allowMockLogin = import\.meta\.env\.DEV/);
  assert.match(text, /const showDevLoginAssist = import\.meta\.env\.DEV/);
  assert.match(text, /phone: ""/);
  assert.match(text, /password: ""/);
  assert.doesNotMatch(text, /phone: "19900000001"/);
  assert.doesNotMatch(text, /password: "abc123456"/);
  assert.match(text, /v-if="allowMockLogin"/);
});
