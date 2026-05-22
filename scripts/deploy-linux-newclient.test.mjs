import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const pkg = JSON.parse(
  fs.readFileSync(
    new URL("../newclient/package.json", import.meta.url),
    "utf8"
  )
);
const deploy = fs.readFileSync(new URL("./deploy_linux_server.sh", import.meta.url), "utf8");
const nginx = fs.readFileSync(
  new URL("../deploy/linux/sercherai.nginx.conf.template", import.meta.url),
  "utf8"
);
const guide = fs.readFileSync(new URL("../docs/DEPLOY_LINUX.md", import.meta.url), "utf8");

test("newclient package exposes aggregate build", () => {
  assert.equal(typeof pkg.scripts.build, "string");
  assert.match(pkg.scripts.build, /build:pc/);
  assert.match(pkg.scripts.build, /build:h5/);
});

test("deploy script builds and publishes newclient", () => {
  assert.match(deploy, /NEWCLIENT_PORT/);
  assert.match(deploy, /\$\{WWW_DIR\}\/newclient/);
  assert.match(deploy, /cd "\$\{ROOT_DIR\}\/newclient"/);
  assert.match(deploy, /npm run build/);
  assert.match(deploy, /dist-h5/);
  assert.doesNotMatch(
    deploy,
    /copy_dir_without_hidden "\$\{ROOT_DIR\}\/client\/dist"/
  );
});

test("nginx and deploy doc use newclient naming", () => {
  assert.match(nginx, /__NEWCLIENT_PORT__/);
  assert.match(nginx, /__NEWCLIENT_ROOT__/);
  assert.match(guide, /backend\/admin\/newclient/);
});
