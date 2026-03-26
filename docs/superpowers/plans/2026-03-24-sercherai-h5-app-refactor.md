# SercherAI H5 App Refactor Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rebuild the H5 `/m/*` experience into a Xueqiu-like mobile content app with stronger reading flow, lighter shell chrome, and app-style bottom navigation while keeping existing APIs and auth behavior intact.

**Architecture:** Keep the existing H5 route tree and data-fetching APIs, but refactor the shell and page composition so layout, typography, CTA placement, and list/detail rhythm behave like a mobile content product instead of a compressed desktop web page. Shared business logic stays intact; the H5 layer owns presentation-focused UI models and styles.

**Tech Stack:** Vue 3, Vue Router, existing shared auth/session/http modules, H5-specific CSS, Vite.

---

### Task 1: Refresh the H5 shell into a mobile content-app frame

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/layouts/H5Layout.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/styles/h5-shell.css`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/styles/h5-ui.css`

- [ ] Step 1: Inspect current shell behavior and identify desktop-like chrome to remove.
- [ ] Step 2: Simplify the top bar into a mobile headline/status strip with lighter branding.
- [ ] Step 3: Rework the bottom tab bar into a stronger app-style navigation with larger hit areas and clearer active state.
- [ ] Step 4: Update shell/background/token styles to emphasize content flow over floating dashboard cards.
- [ ] Step 5: Verify H5 shell still renders all tab routes without breaking auth/logout links.

### Task 2: Rebuild home and news around content-first mobile reading

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5HomeView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5NewsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/components/H5ArticleCard.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/components/H5SectionBlock.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/components/H5StickyCta.vue`

- [ ] Step 1: Write or extend a focused regression test for any extracted pure helper used to support new content ordering if needed.
- [ ] Step 2: Refactor home into a headline-style “today’s core view” followed by related news and lightweight actions.
- [ ] Step 3: Refactor news into a stronger channel-switch + article-feed layout with more app-like list rhythm.
- [ ] Step 4: Adjust article/detail presentation so the正文 remains primary and the sticky action area feels like a mobile app footer.
- [ ] Step 5: Verify direct article navigation and unauthenticated redirect flows still work.

### Task 3: Rebuild strategy, membership, and profile into focused mobile destinations

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5MembershipView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5ProfileView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/components/H5HeroCard.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/components/H5SummaryCard.vue`

- [ ] Step 1: Make strategy list/detail feel like a curated viewpoint feed instead of a strategy workbench.
- [ ] Step 2: Make membership feel like a mobile cashier/upgrade page with package-first hierarchy.
- [ ] Step 3: Make profile feel like an account center with status, entries, and message/order management grouped cleanly.
- [ ] Step 4: Tune shared H5 components so these three pages inherit the new mobile visual rhythm.
- [ ] Step 5: Verify authenticated pages still respect existing login redirect behavior.

### Task 4: Run validation for H5-only regressions

**Files:**
- Test: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/auth-page.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/page-state.test.mjs`
- Verify: `/Users/gjhan21/cursor/sercherai/client/package.json`

- [ ] Step 1: Run the existing H5 pure-module tests.
- [ ] Step 2: Run the H5 build to catch layout/router/import regressions.
- [ ] Step 3: If a verification fails, fix the smallest relevant issue and re-run.
- [ ] Step 4: Summarize changed files, verification evidence, and any remaining visual follow-up items.
