# Client PC/H5 Demo Alignment Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Align the real `client` PC and H5 applications with the `client/newh5demo` design language, including new independent H5 `watchlist` and `archive` pages, while preserving existing APIs, auth flows, membership rules, and data behavior.

**Architecture:** Keep the current split app structure: PC stays in `client/src/apps/pc` plus the shared `client/src/views` pages, and H5 stays in `client/src/apps/h5` with presentation-specific view-model helpers. Introduce only the smallest new pure helper modules needed to support the new H5 secondary pages and auth redirect handling, then reshape the Vue templates and CSS around those tested helpers.

**Tech Stack:** Vue 3, Vue Router, existing shared auth/session/http modules, Node built-in test runner (`node --test`), Vite, CSS in SFCs plus `pc-shell.css`, `h5-shell.css`, and `h5-ui.css`.

---

## File Map

### Shell, Navigation, and Redirects

- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/pc/layouts/PcLayout.vue`
  Responsibility: PC global shell, top nav, auth entry, desktop page container.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/pc/styles/pc-shell.css`
  Responsibility: PC shell tokens, header, container sizing, desktop chrome.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/layouts/H5Layout.vue`
  Responsibility: H5 header, account chip, bottom tab bar, shared mobile frame.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/styles/h5-shell.css`
  Responsibility: H5 shell frame, header, main area, tab bar.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/styles/h5-ui.css`
  Responsibility: H5 shared card rhythm, chip styles, typography, CTA surfaces.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/router/index.js`
  Responsibility: H5 route tree; make `archive` and `watchlist` real views.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/shell-meta.js`
  Responsibility: H5 page-scene copy for header/title/pulse text.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/shell-meta.test.mjs`
  Responsibility: Regression tests for route-to-scene mapping.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/auth-page.js`
  Responsibility: H5 auth redirect normalization, scene copy, and post-login return paths.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/auth-page.test.mjs`
  Responsibility: Regression tests for redirect normalization and auth surface copy.

### New H5 View-Model Helpers

- Create: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/watchlist-feed.js`
  Responsibility: Convert watchlist data into mobile-friendly summary cards, lead item, rows, and empty-state copy.
- Create: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/watchlist-feed.test.mjs`
  Responsibility: Prove watchlist mobile model handles populated and empty states.
- Create: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/archive-feed.js`
  Responsibility: Convert archive/history data into mobile-friendly summary cards, timeline rows, and status badges.
- Create: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/archive-feed.test.mjs`
  Responsibility: Prove archive mobile model handles populated and empty states.

### H5 Pages

- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5HomeView.vue`
  Responsibility: Demo-aligned hero rhythm plus primary entry into `/watchlist`.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5StrategyView.vue`
  Responsibility: Demo-aligned viewpoint feed and optional secondary watchlist CTA.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5NewsView.vue`
  Responsibility: Demo-aligned reading flow plus primary entry into `/archive`.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5MembershipView.vue`
  Responsibility: Mobile cashier hierarchy aligned to demo.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5ProfileView.vue`
  Responsibility: Demo-aligned account center ordering.
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5AuthView.vue`
  Responsibility: Demo-aligned source/redirect/invite presentation.
- Create: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5WatchlistView.vue`
  Responsibility: New independent H5 watchlist page.
- Create: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5ArchiveView.vue`
  Responsibility: New independent H5 archive page.

### H5 Shared Components Likely to Need Styling/Hierarchy Changes

- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/components/H5HeroCard.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/components/H5SectionBlock.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/components/H5SummaryCard.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/components/H5ArticleCard.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/components/H5StickyCta.vue`

### PC Pages

- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/HomeView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/MyWatchlistView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/NewsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/MembershipView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/ProfileView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/AuthView.vue`

### Verification

- Verify: `/Users/gjhan21/cursor/sercherai/client/package.json`
  Responsibility: source of the `build:h5`, `build:pc`, and `build:all` commands.

## Implementation Notes

- Do not copy `client/newh5demo/*.html` into Vue templates wholesale.
- Reuse existing API calls, computed data, and auth/session behavior wherever possible.
- Keep H5 bottom tabs at five items: `home`, `news`, `strategies`, `membership`, `profile`.
- `watchlist` and `archive` must be discoverable through in-page CTAs, not the tab bar.
- `/m/...` in product language corresponds to H5 runtime paths like `/watchlist` and `/archive` inside the H5 router and auth redirect helpers.
- Do not introduce snapshot tests or component-test infrastructure just for this change. Prefer new pure helper tests plus build verification.

### Task 1: Create an isolated implementation workspace and confirm a clean baseline

**Files:**
- Verify: `/Users/gjhan21/cursor/sercherai/.worktrees`
- Verify: `/Users/gjhan21/cursor/sercherai/client/package.json`

- [ ] **Step 1: Create the worktree**

Run:

```bash
git check-ignore -q .worktrees
git worktree add .worktrees/codex-client-pc-h5-demo -b codex/client-pc-h5-demo
```

Expected: `.worktrees` is ignored and the new worktree is created without touching the dirty main workspace.

- [ ] **Step 2: Move into the worktree and make sure dependencies are available**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client install
```

Expected: `client/node_modules` is present and install exits with code 0.

- [ ] **Step 3: Run the existing H5 pure-module tests as a baseline**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
node --test \
  client/src/apps/h5/lib/auth-page.test.mjs \
  client/src/apps/h5/lib/display-copy.test.mjs \
  client/src/apps/h5/lib/home-feed.test.mjs \
  client/src/apps/h5/lib/membership-cashier.test.mjs \
  client/src/apps/h5/lib/news-feed.test.mjs \
  client/src/apps/h5/lib/page-state.test.mjs \
  client/src/apps/h5/lib/profile-center.test.mjs \
  client/src/apps/h5/lib/shell-meta.test.mjs \
  client/src/apps/h5/lib/strategy-feed.test.mjs \
  client/src/apps/h5/lib/surface-tone.test.mjs
```

Expected: all existing H5 pure-module tests pass before feature work starts.

- [ ] **Step 4: Run the baseline build**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client run build:all
```

Expected: both PC and H5 builds succeed before feature edits begin.

- [ ] **Step 5: Commit nothing, just record the baseline**

Expected: the worktree is ready; if Step 3 or Step 4 fails, stop and decide whether to fix the baseline first or proceed with known failures.

### Task 2: Add tested H5 feed helpers for the new watchlist and archive pages

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/watchlist-feed.js`
- Create: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/watchlist-feed.test.mjs`
- Create: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/archive-feed.js`
- Create: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/archive-feed.test.mjs`

- [ ] **Step 1: Write the failing tests first**

Add tests shaped like:

```js
test("buildWatchlistFeedModel creates lead card and follow-up rows", () => {
  const model = buildWatchlistFeedModel({ items: [/* populated sample */] });
  assert.equal(model.leadItem.id, "w1");
  assert.equal(model.rows.length, 2);
});

test("buildArchiveFeedModel creates summary stats and timeline rows", () => {
  const model = buildArchiveFeedModel({ items: [/* populated sample */] });
  assert.equal(model.summaryCards.length, 3);
  assert.equal(model.timeline.length, 1);
});
```

- [ ] **Step 2: Run the tests to verify they fail**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
node --test \
  client/src/apps/h5/lib/watchlist-feed.test.mjs \
  client/src/apps/h5/lib/archive-feed.test.mjs
```

Expected: FAIL because the new helper modules or exports do not exist yet.

- [ ] **Step 3: Implement the minimal helpers**

Create exports shaped like:

```js
export function buildWatchlistFeedModel({ items = [] } = {}) {
  return { leadItem: null, rows: [], summaryCards: [], emptyState: {} };
}

export function buildArchiveFeedModel({ items = [] } = {}) {
  return { summaryCards: [], timeline: [], emptyState: {} };
}
```

Fill them with the minimum mapping needed for the tests and the planned mobile pages. Reuse the naming style already used in `home-feed.js`, `news-feed.js`, and `profile-center.js`.

- [ ] **Step 4: Run the tests again and confirm green**

Run the same `node --test` command from Step 2.

Expected: PASS for both new helper test files.

- [ ] **Step 5: Commit the helper layer**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git add client/src/apps/h5/lib/watchlist-feed.js client/src/apps/h5/lib/watchlist-feed.test.mjs client/src/apps/h5/lib/archive-feed.js client/src/apps/h5/lib/archive-feed.test.mjs
git commit -m "feat: add h5 watchlist and archive feed models"
```

### Task 3: Update H5 route metadata and auth redirects so the new pages are first-class destinations

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5WatchlistView.vue`
- Create: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5ArchiveView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/shell-meta.js`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/shell-meta.test.mjs`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/auth-page.js`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/auth-page.test.mjs`

- [ ] **Step 1: Extend the tests before changing behavior**

Add failing assertions like:

```js
assert.deepEqual(resolveShellScene("/watchlist"), {
  section: "关注",
  title: "我的关注",
  subtitle: "变化工作台按回访节奏展示跟踪对象",
  pulse: "回访中"
});

assert.equal(normalizeAuthRedirect("/archive?article=n2"), "/archive?article=n2");
assert.equal(normalizeAuthRedirect("/watchlist"), "/watchlist");
```

- [ ] **Step 2: Run the tests to verify the current behavior is wrong**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
node --test client/src/apps/h5/lib/shell-meta.test.mjs client/src/apps/h5/lib/auth-page.test.mjs
```

Expected: FAIL because `shell-meta` has no watchlist/archive scene copy and `auth-page` still rewrites legacy destinations.

- [ ] **Step 3: Implement the route and redirect updates**

- Create minimal stub views first so the router can import real modules before the first `build:h5` run.
- The stubs can render a simple `<div class="h5-page">` placeholder with a heading and note; they will be expanded in Tasks 4 and 5.
- Replace the H5 `archive` and `watchlist` redirects in `router/index.js` with real view routes.
- Update `shell-meta.js` so `resolveShellScene()` returns demo-aligned scene copy for `/watchlist` and `/archive`.
- Remove the legacy redirect rewrites in `auth-page.js`, add scene copy for the two new destinations, and keep all other valid destinations intact.

- [ ] **Step 4: Re-run tests plus an H5 build smoke check**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
node --test client/src/apps/h5/lib/shell-meta.test.mjs client/src/apps/h5/lib/auth-page.test.mjs
npm --prefix client run build:h5
```

Expected: both test files pass and the H5 build still succeeds.

- [ ] **Step 5: Commit the route/auth foundation**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git add client/src/apps/h5/views/H5WatchlistView.vue client/src/apps/h5/views/H5ArchiveView.vue client/src/apps/h5/router/index.js client/src/apps/h5/lib/shell-meta.js client/src/apps/h5/lib/shell-meta.test.mjs client/src/apps/h5/lib/auth-page.js client/src/apps/h5/lib/auth-page.test.mjs
git commit -m "feat: make h5 watchlist and archive routable"
```

### Task 4: Build the H5 watchlist page and wire its primary entry from home

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5WatchlistView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5HomeView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5StrategyView.vue`

- [ ] **Step 1: Wire the new page around the tested watchlist feed helper**

Implementation target:

```js
const watchlistModel = computed(() => buildWatchlistFeedModel({ items: hydratedWatchRows.value }));
```

Use the helper output to drive summary cards, a lead follow-up item, the list rows, and empty-state copy.

Data/state requirements:

- The primary source must be `listWatchedStocks()` from `/Users/gjhan21/cursor/sercherai/client/src/lib/watchlist.js`, not the general recommendation list.
- Reuse the local watchlist event (`WATCHLIST_EVENT`) so the page refreshes when a stock is added or removed elsewhere.
- Hydrate saved rows with the same market detail/insight/version-history sources the PC watchlist page uses where needed, keeping the data scope mobile-friendly.
- Handle four readable states: unauthenticated-with-local-items, unauthenticated-empty, logged-in empty, and inline error/fallback.
- A logged-out user with locally saved watchlist items must still see those saved items plus a login CTA; do not hide local data behind an immediate auth wall.
- A logged-out user with no local items should see the empty-state guidance plus a login CTA with `redirect=/watchlist`.
- Preserve remove-from-watchlist behavior.

- [ ] **Step 2: Add the required primary entry from home**

In `H5HomeView.vue`, make the “延伸观察” block or its primary CTA route to `/watchlist` instead of leaving it as a non-navigable preview.

- [ ] **Step 3: Add the required strategy-page watchlist behavior**

In `H5StrategyView.vue`, preserve the spec-required “加入关注 / 取消关注” behavior by reusing `isWatchedStock()`, `saveWatchedStock()`, and `removeWatchedStock()` from `/Users/gjhan21/cursor/sercherai/client/src/lib/watchlist.js`. Keep the page’s secondary navigation into `/watchlist` if there is a natural CTA slot, but the mutation behavior itself is required and cannot be replaced by a route-only shortcut.

- [ ] **Step 4: Verify the page compiles and the entry path works**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client run build:h5
```

Expected: build passes and the home-to-watchlist route is present in the compiled H5 app.

- [ ] **Step 5: Commit the H5 watchlist slice**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git add client/src/apps/h5/views/H5WatchlistView.vue client/src/apps/h5/views/H5HomeView.vue client/src/apps/h5/views/H5StrategyView.vue
git commit -m "feat: add h5 watchlist page"
```

### Task 5: Build the H5 archive page and wire its primary entry from news

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5ArchiveView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5NewsView.vue`

- [ ] **Step 1: Wire the new page around the tested archive feed helper**

Implementation target:

```js
const archiveModel = computed(() => buildArchiveFeedModel({ items: archiveRows.value }));
```

Use the helper output to render summary cards, a concise reading order, and a single-column archive timeline.

Data/state requirements:

- Reuse the PC archive page’s data pattern: `listStockRecommendations()` for the base rows, then hydrate with `getStockRecommendationInsight()` and `getStockRecommendationVersionHistory()` as needed.
- Preserve the staged access model from the spec: visitor sees a readable public subset, registered users see a larger subset, VIP users see the full archive.
- Preserve readable downgrade behavior: visitor CTA to login, registered CTA to membership, inline error state, and safe demo/fallback content when API data is unavailable.
- Do not rely on unrelated `news` list items as the archive data source; the news page only provides the entry point.

- [ ] **Step 2: Add the required primary entry from news**

In `H5NewsView.vue`, add the `/archive` CTA inside the existing list-view action area (`news-feed-actions`) so it sits alongside “继续加载” / “去看策略” and remains visible even when no article detail is open. This CTA is the required primary discoverability path from news into archive.

- [ ] **Step 3: Keep the page secondary, not tabbed**

Do not add `archive` to `TAB_ITEMS`. The page should be reachable by CTA and auth redirect, not global tab-bar promotion.

- [ ] **Step 4: Verify the page compiles and the entry path works**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client run build:h5
```

Expected: build passes and the news-to-archive route is present in the compiled H5 app.

- [ ] **Step 5: Commit the H5 archive slice**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git add client/src/apps/h5/views/H5ArchiveView.vue client/src/apps/h5/views/H5NewsView.vue
git commit -m "feat: add h5 archive page"
```

### Task 6: Refresh the H5 shell and shared mobile components to match the demo tone

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/layouts/H5Layout.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/styles/h5-shell.css`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/styles/h5-ui.css`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/components/H5HeroCard.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/components/H5SectionBlock.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/components/H5SummaryCard.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/components/H5ArticleCard.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/components/H5StickyCta.vue`

- [ ] **Step 1: Rework the shell hierarchy**

Keep the brand/header/tab bar structure, but restyle them to feel like the finance demo rather than a generic content app.

- [ ] **Step 2: Make the shared card system match the demo rhythm**

Update typography, chip tone, card radius, shadows, and spacing in `h5-ui.css` so all H5 views inherit the same blue-gold finance language.

- [ ] **Step 3: Tune shared components instead of duplicating page-local one-off CSS**

Use `H5HeroCard`, `H5SectionBlock`, `H5SummaryCard`, `H5ArticleCard`, and `H5StickyCta` as the main reuse points for the new hierarchy.

- [ ] **Step 4: Verify that all five tab destinations still render**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client run build:h5
```

Expected: build passes and there are no missing imports, broken shared styles, or shell-level route regressions.

- [ ] **Step 5: Commit the H5 shell refresh**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git add client/src/apps/h5/layouts/H5Layout.vue client/src/apps/h5/styles/h5-shell.css client/src/apps/h5/styles/h5-ui.css client/src/apps/h5/components/H5HeroCard.vue client/src/apps/h5/components/H5SectionBlock.vue client/src/apps/h5/components/H5SummaryCard.vue client/src/apps/h5/components/H5ArticleCard.vue client/src/apps/h5/components/H5StickyCta.vue
git commit -m "feat: refresh h5 shell for finance demo"
```

### Task 7: Align the remaining H5 views to the demo information order

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5HomeView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5NewsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5MembershipView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5ProfileView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/views/H5AuthView.vue`

- [ ] **Step 1: Reorder home and strategy around demo-first reading flow**

Home should emphasize “今日主线 -> 延伸观察 -> 相关资讯 -> 今日行动”. Strategy should emphasize “结论 -> 理由 -> 风险 -> 下一步”.

- [ ] **Step 2: Reorder news and membership around deep-read / conversion flow**

News should foreground the focus article and archive handoff. Membership should foreground current status, todo, recommended package, and payment action.

- [ ] **Step 3: Reorder profile and auth around account / source-entry flow**

Profile should put account status and todo ahead of secondary management blocks. Auth should keep invite code, source scene, and redirect clarity visible before the form.

- [ ] **Step 4: Verify the whole H5 app after the view edits**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client run build:h5
```

Expected: all H5 view changes compile together.

- [ ] **Step 5: Commit the H5 page-alignment pass**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git add client/src/apps/h5/views/H5HomeView.vue client/src/apps/h5/views/H5StrategyView.vue client/src/apps/h5/views/H5NewsView.vue client/src/apps/h5/views/H5MembershipView.vue client/src/apps/h5/views/H5ProfileView.vue client/src/apps/h5/views/H5AuthView.vue
git commit -m "feat: align h5 pages with finance demo"
```

### Task 8: Refresh the PC shell into the demo-style finance terminal frame

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/pc/layouts/PcLayout.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/pc/styles/pc-shell.css`

- [ ] **Step 1: Replace the thin white shell chrome with the demo-style dark finance header**

Keep the existing route-aware nav and auth/logout behavior, but change the shell hierarchy to match the finance terminal tone.

- [ ] **Step 2: Remove the temporary desktop banner**

Delete the “PC 端骨架已独立” banner and use the main page area itself as the first product surface.

- [ ] **Step 3: Tune container width, header pills, and shared desktop spacing**

Make the shell wide enough for the new hero + rail layout without creating a second competing design system.

- [ ] **Step 4: Verify the PC shell compiles**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client run build:pc
```

Expected: PC build passes and the shell still wraps all existing routes.

- [ ] **Step 5: Commit the PC shell refresh**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git add client/src/apps/pc/layouts/PcLayout.vue client/src/apps/pc/styles/pc-shell.css
git commit -m "feat: refresh pc shell for finance demo"
```

### Task 9: Re-layout PC home, strategy, and watchlist around the demo’s decision-workbench structure

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/views/HomeView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/views/StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/views/MyWatchlistView.vue`

- [ ] **Step 1: Re-layout home without deleting its existing data richness**

Keep the current recommendation/news/watchlist logic, but restructure it into `hero + main rail + side rail` instead of many equally weighted blocks.

- [ ] **Step 2: Re-layout strategy into a lead explanation page, not a dense control board**

Prioritize the active recommendation, evidence, risk boundaries, and version drift before lower-priority controls.

- [ ] **Step 3: Re-layout watchlist into a follow-up workstation**

Keep the existing watchlist logic, but surface lead item, change cues, and next-action guidance in the order shown by the PC demo.

- [ ] **Step 4: Verify the PC build after these three heavy page edits**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client run build:pc
```

Expected: build passes and these three pages still compile together.

- [ ] **Step 5: Commit the first PC page cluster**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git add client/src/views/HomeView.vue client/src/views/StrategyView.vue client/src/views/MyWatchlistView.vue
git commit -m "feat: align pc decision pages with finance demo"
```

### Task 10: Re-layout PC news, archive, and membership around deep-read and conversion flow

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/views/NewsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/views/RecommendationArchiveView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/views/MembershipView.vue`

- [ ] **Step 1: Re-layout news into a clearer left/right deep-read structure**

Keep article data and access logic intact, but make the reading path and detail/attachment handoff match the demo.

- [ ] **Step 2: Re-layout archive into a true historical review center**

Keep the current timeline/version/history data, but make “结果 -> 理由 -> 版本变化 -> 数据来源” the visible reading order.

- [ ] **Step 3: Re-layout membership to put status and todo before price**

Preserve package/order/payment logic while moving status, activation, and next-step messaging ahead of pricing blocks.

- [ ] **Step 4: Verify the PC build after this second page cluster**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client run build:pc
```

Expected: build passes and all three pages compile together.

- [ ] **Step 5: Commit the second PC page cluster**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git add client/src/views/NewsView.vue client/src/views/RecommendationArchiveView.vue client/src/views/MembershipView.vue
git commit -m "feat: align pc reading and membership pages"
```

### Task 11: Re-layout PC profile and auth around account management and source-aware entry

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/views/ProfileView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/views/AuthView.vue`

- [ ] **Step 1: Re-layout profile as an account management console**

Prioritize account status, membership status, unread messages, invite progress, and today’s todo before deep management tables.

- [ ] **Step 2: Re-layout auth as a demo-style dual-column identity entry**

Keep current account/password auth and redirect logic, but surface source context, redirect path, and invite code before form submission.

- [ ] **Step 3: Preserve the archive/watchlist naming already used by desktop auth flows**

Do not regress the existing PC redirect scene descriptions for `/archive` and `/watchlist`.

- [ ] **Step 4: Verify the PC build**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client run build:pc
```

Expected: build passes and the auth/profile route pair still compile together.

- [ ] **Step 5: Commit the final PC page cluster**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git add client/src/views/ProfileView.vue client/src/views/AuthView.vue
git commit -m "feat: align pc account and auth pages"
```

### Task 12: Run full verification and smoke-check the new route/entry expectations

**Files:**
- Verify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/package.json`
- Verify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/auth-page.test.mjs`
- Verify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/shell-meta.test.mjs`
- Verify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/watchlist-feed.test.mjs`
- Verify: `/Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo/client/src/apps/h5/lib/archive-feed.test.mjs`

- [ ] **Step 1: Run the targeted H5 pure-module tests**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
node --test \
  client/src/apps/h5/lib/auth-page.test.mjs \
  client/src/apps/h5/lib/shell-meta.test.mjs \
  client/src/apps/h5/lib/watchlist-feed.test.mjs \
  client/src/apps/h5/lib/archive-feed.test.mjs
```

Expected: PASS for all redirect/scene/feed tests.

- [ ] **Step 2: Run the full dual-build verification**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
npm --prefix client run build:all
```

Expected: both PC and H5 builds succeed with the final integrated changes.

- [ ] **Step 3: Smoke-check route discoverability and auth return behavior**

Verify manually in the dev server or preview build:

- `/m/home` contains a primary path into `/m/watchlist`
- `/m/news` contains a primary path into `/m/archive`
- unauthenticated direct access to `/m/watchlist` renders a readable degraded page first, not an immediate forced redirect
- unauthenticated direct access to `/m/watchlist` with locally saved items still shows those items plus a login CTA
- unauthenticated direct access to `/m/archive` renders a readable degraded page first, not an immediate forced redirect
- unauthenticated H5 navigation to `/m/watchlist` returns to `/m/watchlist` after login
- unauthenticated H5 navigation to `/m/archive` returns to `/m/archive` after login
- adding a stock from `/m/strategies` makes it appear on `/m/watchlist`, and removing it from `/m/watchlist` still works
- H5 bottom-tab active state still behaves correctly for `home`, `news`, `strategies`, `membership`, and `profile`
- `/m/...` routes still open the H5 app while non-`/m` routes still open the PC app
- H5 auth still preserves invite-code registration and source-return behavior from `/m/auth`
- membership primary CTA, continue-payment flow, and KYC activation handoff still work on both PC and H5
- news正文 / 附件 / 权限承接 flow still works on both PC and H5 without bypassing the intended access state
- PC nav still reaches all eight desktop pages

- [ ] **Step 4: If any verification fails, fix the smallest relevant issue and re-run the failed command**

Do not widen scope during verification. Repair the exact regression, then re-run the specific failing test/build before rerunning the full suite.

- [ ] **Step 5: Commit the verification fixes and summarize evidence**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/.worktrees/codex-client-pc-h5-demo
git status --short
```

Expected: only intentional implementation files remain changed. If verification required fixes, commit them with the smallest accurate message before handoff.
