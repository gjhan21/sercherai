import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const devctl = fs.readFileSync(new URL("./devctl.sh", import.meta.url), "utf8");
const readme = fs.readFileSync(new URL("./README.md", import.meta.url), "utf8");

test("devctl exposes newclient instead of client", () => {
  assert.match(devctl, /\bnewclient\b/);
  assert.doesNotMatch(devctl, /\[strategy-graph\|strategy-engine\|backend\|admin\|client\|all\]/);
  assert.match(devctl, /\.run\/newclient\.env/);
  assert.match(devctl, /NEWCLIENT_PORT/);
});

test("scripts README documents newclient env and commands", () => {
  assert.match(readme, /newclient/);
  assert.match(readme, /\.run\/newclient\.env/);
  assert.match(readme, /NEWCLIENT_PORT/);
});
