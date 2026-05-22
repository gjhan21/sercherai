# Newclient Formal Cutover Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Cut the repository's official frontend identity over from the legacy `client` name to `newclient` across development, build, deployment, active docs, and historical doc guidance while keeping the old `client` directory as read-only reference.

**Architecture:** Keep the current repo structure, but hard-cut every supported operational chain to `newclient`. Use small Node-based regression tests to pin the naming and deployment contract, then update shell scripts, frontend package scripts, deployment config, active docs, and historical notes until the tests and frontend builds all pass.

**Tech Stack:** Bash, Node.js `node:test`, Vite, Vue, nginx template rendering, Markdown documentation.

---

## Scope

This plan covers one coherent sub-project:

- make `newclient` the only official frontend for runtime, build, deployment, and onboarding

It does not include:

- deleting `/Users/gjhan21/cursor/sercherai/client`
- migrating old `client` source code into `newclient`
- redesigning frontend features

## Current Code Reality

### What is already true

- `/Users/gjhan21/cursor/sercherai/scripts/devctl.sh` already launches `/Users/gjhan21/cursor/sercherai/newclient` for the frontend service body.
- `/Users/gjhan21/cursor/sercherai/newclient/package.json` already exposes `build:pc` and `build:h5`.
- `npm run build:h5` currently emits `dist-h5/m/index.html` plus shared hashed assets under `dist-h5/assets`.

### What is still wrong

- `scripts/devctl.sh` still exposes the official service name as `client`.
- `.run/client.env`, `CLIENT_HOST`, and `CLIENT_PORT` are still the official dev env contract.
- `scripts/deploy_linux_server.sh` still builds `/Users/gjhan21/cursor/sercherai/client` and publishes to `${WWW_DIR}/client`.
- `deploy/linux/sercherai.nginx.conf.template` still renders `CLIENT_ROOT` and `CLIENT_PORT`.
- `scripts/README.md` and `docs/DEPLOY_LINUX.md` still describe the official frontend as `client`.
- historical docs still mention old `client` paths without clearly marking them as historical.

## File Structure

### New test files

- `/Users/gjhan21/cursor/sercherai/scripts/devctl-newclient.test.mjs`
- `/Users/gjhan21/cursor/sercherai/scripts/deploy-linux-newclient.test.mjs`
- `/Users/gjhan21/cursor/sercherai/docs/historical-newclient-note.test.mjs`

### Existing files to modify

- `/Users/gjhan21/cursor/sercherai/scripts/devctl.sh`
- `/Users/gjhan21/cursor/sercherai/newclient/package.json`
- `/Users/gjhan21/cursor/sercherai/scripts/deploy_linux_server.sh`
- `/Users/gjhan21/cursor/sercherai/deploy/linux/sercherai.nginx.conf.template`
- `/Users/gjhan21/cursor/sercherai/scripts/README.md`
- `/Users/gjhan21/cursor/sercherai/docs/DEPLOY_LINUX.md`

### Historical docs to modify with cutover note

- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-24-sercherai-h5-xueqiu-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-24-newclient-homepage-demo-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-24-client-pc-h5-demo-alignment-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-27-pc-community-watchlist-restructure-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-24-newclient-profile-demo-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-24-newclient-strategies-demo-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-26-client-h5demo-style-alignment-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-23-community-discussion-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-24-newclient-fullsite-demo-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-24-newclient-watchlist-demo-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-27-h5-community-watchlist-restructure-design.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-24-client-pc-h5-demo-alignment.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-27-h5-community-watchlist-restructure-plan.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-24-newclient-strategies-demo.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-24-sercherai-h5-app-refactor.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-23-community-discussion-mvp.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-27-pc-community-watchlist-restructure-plan.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-29-stock-futures-forecast-l3-implementation.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-28-stock-futures-forecast-l1-implementation.md`
- `/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-29-stock-futures-forecast-l2-implementation.md`
- `/Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth/README.md`
- `/Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth/阶段7-虚拟沙盘与动态权益墙.md`
- `/Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth/阶段3-历史档案与信任改造.md`
- `/Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth/阶段4-我的关注与回访机制.md`
- `/Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth/阶段0-画布与真相源.md`
- `/Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth/阶段1-今日决策首页.md`
- `/Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth/阶段2-推荐档案页.md`
- `/Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth/阶段6-会员转化与内容节奏.md`

## Shared Historical Note Text

Use one consistent note for the historical docs:

```md
> Historical note:
> This document describes work from the legacy `client` frontend era.
> The current official frontend has moved to `/Users/gjhan21/cursor/sercherai/newclient`.
```

## Task 1: Hard-Cut Dev Naming To `newclient`

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/scripts/devctl-newclient.test.mjs`
- Modify: `/Users/gjhan21/cursor/sercherai/scripts/devctl.sh`
- Modify: `/Users/gjhan21/cursor/sercherai/scripts/README.md`

- [ ] **Step 1: Write the failing dev naming regression test**

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const devctl = fs.readFileSync("/Users/gjhan21/cursor/sercherai/scripts/devctl.sh", "utf8");
const readme = fs.readFileSync("/Users/gjhan21/cursor/sercherai/scripts/README.md", "utf8");

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
```

- [ ] **Step 2: Run the dev naming test to verify it fails**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/scripts/devctl-newclient.test.mjs
```

Expected:

- FAIL because `scripts/devctl.sh` and `scripts/README.md` still use `client`, `.run/client.env`, and `CLIENT_*`.

- [ ] **Step 3: Update `devctl` and script docs to the new formal naming**

Required implementation:

- rename the service from `client` to `newclient` in `ALL_SERVICES`, usage text, case arms, service URL logic, and service command handling
- change env lookup from `.run/client.env` to `.run/newclient.env`
- change `CLIENT_HOST/CLIENT_PORT` to `NEWCLIENT_HOST/NEWCLIENT_PORT`
- update `scripts/README.md` examples, service lists, default port section, env file section, and restart examples to use `newclient`

- [ ] **Step 4: Run the dev naming test again**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/scripts/devctl-newclient.test.mjs
```

Expected:

- PASS

- [ ] **Step 5: Commit**

```bash
git add \
  scripts/devctl-newclient.test.mjs \
  scripts/devctl.sh \
  scripts/README.md
git commit -m "refactor: rename frontend dev service to newclient"
```

## Task 2: Switch Build And Deployment To `newclient`

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/scripts/deploy-linux-newclient.test.mjs`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/package.json`
- Modify: `/Users/gjhan21/cursor/sercherai/scripts/deploy_linux_server.sh`
- Modify: `/Users/gjhan21/cursor/sercherai/deploy/linux/sercherai.nginx.conf.template`
- Modify: `/Users/gjhan21/cursor/sercherai/docs/DEPLOY_LINUX.md`

- [ ] **Step 1: Write the failing deployment regression test**

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const pkg = JSON.parse(fs.readFileSync("/Users/gjhan21/cursor/sercherai/newclient/package.json", "utf8"));
const deploy = fs.readFileSync("/Users/gjhan21/cursor/sercherai/scripts/deploy_linux_server.sh", "utf8");
const nginx = fs.readFileSync("/Users/gjhan21/cursor/sercherai/deploy/linux/sercherai.nginx.conf.template", "utf8");
const guide = fs.readFileSync("/Users/gjhan21/cursor/sercherai/docs/DEPLOY_LINUX.md", "utf8");

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
  assert.doesNotMatch(deploy, /copy_dir_without_hidden "\$\{ROOT_DIR\}\/client\/dist"/);
});

test("nginx and deploy doc use newclient naming", () => {
  assert.match(nginx, /__NEWCLIENT_PORT__/);
  assert.match(nginx, /__NEWCLIENT_ROOT__/);
  assert.match(guide, /backend\/admin\/newclient/);
});
```

- [ ] **Step 2: Run the deployment regression test to verify it fails**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/scripts/deploy-linux-newclient.test.mjs
```

Expected:

- FAIL because the package lacks a unified `build` script, deploy still references `client`, and nginx/docs still use `CLIENT_*`.

- [ ] **Step 3: Implement the build and deployment cutover**

Required implementation:

- add `build` to `newclient/package.json` so it runs both `build:pc` and `build:h5`
- rename deploy env/placeholder naming from `CLIENT_*` to `NEWCLIENT_*`
- build `/Users/gjhan21/cursor/sercherai/newclient` instead of the old frontend
- publish to `${WWW_DIR}/newclient`
- assemble one runtime tree that includes:
  - `dist/index.html` for `/`
  - `dist-h5/m/index.html` for `/m/`
  - both hashed asset sets under the final static root
- update `docs/DEPLOY_LINUX.md` to describe `newclient` as the official frontend

- [ ] **Step 4: Run the deployment regression test and frontend builds**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/scripts/deploy-linux-newclient.test.mjs
cd /Users/gjhan21/cursor/sercherai/newclient && npm run build
```

Expected:

- deployment regression test PASS
- `npm run build` PASS
- PC build emits `dist/index.html`
- H5 build emits `dist-h5/m/index.html`

- [ ] **Step 5: Commit**

```bash
git add \
  scripts/deploy-linux-newclient.test.mjs \
  newclient/package.json \
  scripts/deploy_linux_server.sh \
  deploy/linux/sercherai.nginx.conf.template \
  docs/DEPLOY_LINUX.md
git commit -m "refactor: deploy newclient as the official frontend"
```

## Task 3: Add Historical Cutover Notes To Searchable Old-Client Docs

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/docs/historical-newclient-note.test.mjs`
- Modify: all historical docs listed in the File Structure section

- [ ] **Step 1: Write the failing historical-note regression test**

```js
import test from "node:test";
import assert from "node:assert/strict";
import fs from "node:fs";

const note = [
  "> Historical note:",
  "> This document describes work from the legacy `client` frontend era.",
  "> The current official frontend has moved to `/Users/gjhan21/cursor/sercherai/newclient`."
].join("\n");

const files = [
  "/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-24-client-pc-h5-demo-alignment-design.md",
  "/Users/gjhan21/cursor/sercherai/docs/superpowers/plans/2026-03-24-client-pc-h5-demo-alignment.md",
  "/Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth/README.md"
];

test("historical docs with old client references carry a cutover note", () => {
  for (const file of files) {
    const text = fs.readFileSync(file, "utf8");
    assert.match(text, /client/);
    assert.match(text, new RegExp(note.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")));
  }
});
```

- [ ] **Step 2: Run the historical-note test to verify it fails**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/docs/historical-newclient-note.test.mjs
```

Expected:

- FAIL because the targeted historical docs do not yet contain the standard cutover note.

- [ ] **Step 3: Add the standard historical note to every targeted doc**

Required implementation:

- prepend or place the shared historical note near the top of each targeted document
- do not rewrite the underlying old `client` implementation details
- keep the historical body intact except for minimal wording needed to avoid present-tense confusion

- [ ] **Step 4: Run the historical-note test and a repo scan**

Run:

```bash
node --test /Users/gjhan21/cursor/sercherai/docs/historical-newclient-note.test.mjs
find /Users/gjhan21/cursor/sercherai/docs/superpowers/specs /Users/gjhan21/cursor/sercherai/docs/superpowers/plans /Users/gjhan21/cursor/sercherai/docs/vibe-stock-growth -name '*.md' -print0 | xargs -0 rg -l "/Users/gjhan21/cursor/sercherai/client|client/src"
```

Expected:

- note regression test PASS
- repo scan still returns historical docs, but those docs now clearly identify themselves as legacy-client material

- [ ] **Step 5: Commit**

```bash
git add \
  docs/historical-newclient-note.test.mjs \
  docs/superpowers/specs \
  docs/superpowers/plans \
  docs/vibe-stock-growth
git commit -m "docs: mark legacy client references as historical"
```

## Task 4: Run Final Integrated Verification

**Files:**
- Verify: `/Users/gjhan21/cursor/sercherai/scripts/devctl.sh`
- Verify: `/Users/gjhan21/cursor/sercherai/scripts/deploy_linux_server.sh`
- Verify: `/Users/gjhan21/cursor/sercherai/deploy/linux/sercherai.nginx.conf.template`
- Verify: `/Users/gjhan21/cursor/sercherai/scripts/README.md`
- Verify: `/Users/gjhan21/cursor/sercherai/docs/DEPLOY_LINUX.md`
- Verify: `/Users/gjhan21/cursor/sercherai/newclient/package.json`

- [ ] **Step 1: Run all regression tests**

Run:

```bash
node --test \
  /Users/gjhan21/cursor/sercherai/scripts/devctl-newclient.test.mjs \
  /Users/gjhan21/cursor/sercherai/scripts/deploy-linux-newclient.test.mjs \
  /Users/gjhan21/cursor/sercherai/docs/historical-newclient-note.test.mjs
```

Expected:

- PASS

- [ ] **Step 2: Run frontend builds**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/newclient && npm run build
```

Expected:

- PASS
- `dist/index.html` exists
- `dist-h5/m/index.html` exists

- [ ] **Step 3: Scan for old official naming in live operational files**

Run:

```bash
rg -n "\\bclient\\b|CLIENT_|client\\.env" \
  /Users/gjhan21/cursor/sercherai/scripts/devctl.sh \
  /Users/gjhan21/cursor/sercherai/scripts/deploy_linux_server.sh \
  /Users/gjhan21/cursor/sercherai/scripts/README.md \
  /Users/gjhan21/cursor/sercherai/docs/DEPLOY_LINUX.md \
  /Users/gjhan21/cursor/sercherai/deploy/linux/sercherai.nginx.conf.template
```

Expected:

- no remaining matches for official operational usage

- [ ] **Step 4: Check git diff for accidental old-client source edits**

Run:

```bash
git diff --stat -- /Users/gjhan21/cursor/sercherai/client
```

Expected:

- no changes under the legacy `client` source tree

- [ ] **Step 5: Commit**

```bash
git add \
  scripts/devctl-newclient.test.mjs \
  scripts/deploy-linux-newclient.test.mjs \
  docs/historical-newclient-note.test.mjs \
  scripts/devctl.sh \
  scripts/deploy_linux_server.sh \
  deploy/linux/sercherai.nginx.conf.template \
  scripts/README.md \
  docs/DEPLOY_LINUX.md \
  newclient/package.json \
  docs/superpowers/specs \
  docs/superpowers/plans \
  docs/vibe-stock-growth
git commit -m "refactor: complete newclient formal cutover"
```

## Plan Review Notes

This plan should be reviewed against:

- `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-05-22-newclient-formal-cutover-design.md`

Specific review questions:

- does the deployment assembly step preserve both `/` and `/m/` runtime behavior?
- does the dev naming cutover remove all official `client` aliases as requested?
- do the historical-doc steps annotate history without rewriting it?
