# PC 社区替代关注入口实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 PC 端 `社区` 成为一级主入口，并把 `关注` 收口到 `我的 > 我的关注`，同时保留旧 `/watchlist` 的兼容跳转。

**Architecture:** 复用现有 `CommunityView.vue`、`MyWatchlistView.vue` 和 `ProfileView.vue`，只调整 PC 路由、一级导航、我的页入口卡片和关键 CTA。兼容策略通过 `/watchlist -> /profile/watchlist` 重定向完成，不改 H5 导航，不改 watchlist 数据模型。

**Tech Stack:** Vue 3、Vue Router、Node `--test` 静态断言测试、Vite PC 构建

---

### Task 1: 路由与一级导航收口

**Files:**
- Create: `client/src/apps/pc/router/pc-community-watchlist-layout.test.js`
- Modify: `client/src/apps/pc/router/index.js`
- Modify: `client/src/apps/pc/layouts/PcLayout.vue`
- Test: `client/src/apps/pc/router/pc-community-watchlist-layout.test.js`

- [ ] **Step 1: 写失败测试，约束 PC 路由和导航**

```js
assert.match(routerText, /const CommunityView =/);
assert.match(routerText, /path: "community"/);
assert.match(routerText, /path: "profile\\/watchlist"/);
assert.match(routerText, /path: "watchlist"[\\s\\S]*redirect:/);
assert.match(layoutText, /{ path: "\\/community", label: "社区" }/);
assert.doesNotMatch(layoutText, /{ path: "\\/watchlist", label: "关注" }/);
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /Users/gjhan21/cursor/sercherai/client && node --test src/apps/pc/router/pc-community-watchlist-layout.test.js`
Expected: FAIL，提示缺少 `community` / `profile/watchlist` / 导航 `社区`

- [ ] **Step 3: 最小实现路由与导航**

```js
{ path: "community", name: "pc-community", component: CommunityView },
{ path: "profile/watchlist", name: "pc-profile-watchlist", component: MyWatchlistView, meta: { requiresAuth: true } },
{ path: "watchlist", redirect: "/profile/watchlist" }
```

```js
{ path: "/community", label: "社区" }
```

- [ ] **Step 4: 重跑测试确认通过**

Run: `cd /Users/gjhan21/cursor/sercherai/client && node --test src/apps/pc/router/pc-community-watchlist-layout.test.js`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add client/src/apps/pc/router/index.js client/src/apps/pc/layouts/PcLayout.vue client/src/apps/pc/router/pc-community-watchlist-layout.test.js
git commit -m "feat: route watchlist through profile on pc"
```

### Task 2: 我的页增加“我的关注”入口卡片

**Files:**
- Create: `client/src/views/profile-watchlist-entry.test.js`
- Modify: `client/src/views/ProfileView.vue`
- Test: `client/src/views/profile-watchlist-entry.test.js`

- [ ] **Step 1: 写失败测试，约束 `ProfileView` 出现“我的关注”入口**

```js
assert.match(text, /我的关注/);
assert.match(text, /进入我的关注/);
assert.match(text, /\\/profile\\/watchlist/);
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /Users/gjhan21/cursor/sercherai/client && node --test src/views/profile-watchlist-entry.test.js`
Expected: FAIL，提示缺少入口卡片或路径仍是旧 `/watchlist`

- [ ] **Step 3: 最小实现入口卡片**

在 `ProfileView.vue` 的工作台区域新增一张“我的关注”卡片，说明它是个人跟踪清单入口，CTA 指向 `/profile/watchlist`。

- [ ] **Step 4: 重跑测试确认通过**

Run: `cd /Users/gjhan21/cursor/sercherai/client && node --test src/views/profile-watchlist-entry.test.js`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add client/src/views/ProfileView.vue client/src/views/profile-watchlist-entry.test.js
git commit -m "feat: add profile watchlist entry card"
```

### Task 3: 关键 CTA 与关注页定位收口

**Files:**
- Create: `client/src/views/pc-watchlist-cta-routing.test.js`
- Modify: `client/src/views/MyWatchlistView.vue`
- Modify: `client/src/views/MembershipView.vue`
- Modify: `client/src/views/HomeView.vue`
- Modify: `client/src/views/StrategyView.vue`
- Modify: `client/src/views/NewsView.vue`
- Modify: `client/src/views/RecommendationArchiveView.vue`
- Modify: `client/src/views/AuthView.vue`
- Test: `client/src/views/pc-watchlist-cta-routing.test.js`

- [ ] **Step 1: 写失败测试，约束主链路都跳 `/profile/watchlist`**

```js
assert.match(membershipText, /\\/profile\\/watchlist/);
assert.match(authText, /redirectPath.value === "\\/profile\\/watchlist"/);
assert.match(strategyText, /navigateWithStrategyTracking\\("\\/profile\\/watchlist"/);
assert.doesNotMatch(homeText, /router\\.push\\("\\/watchlist"\\)/);
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /Users/gjhan21/cursor/sercherai/client && node --test src/views/pc-watchlist-cta-routing.test.js`
Expected: FAIL，提示仍存在旧 `/watchlist` 一级入口写法

- [ ] **Step 3: 最小实现 CTA 收口**

把 PC 端关键 CTA、认证回跳说明和关注页文案统一改到 `/profile/watchlist`，同时保留旧重定向兼容。

- [ ] **Step 4: 重跑测试确认通过**

Run: `cd /Users/gjhan21/cursor/sercherai/client && node --test src/views/pc-watchlist-cta-routing.test.js`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add client/src/views/MyWatchlistView.vue client/src/views/MembershipView.vue client/src/views/HomeView.vue client/src/views/StrategyView.vue client/src/views/NewsView.vue client/src/views/RecommendationArchiveView.vue client/src/views/AuthView.vue client/src/views/pc-watchlist-cta-routing.test.js
git commit -m "feat: move pc watchlist entry under profile"
```

### Task 4: 社区页补一级入口感并做整体验证

**Files:**
- Create: `client/src/views/community-entry-surface.test.js`
- Modify: `client/src/views/CommunityView.vue`
- Test: `client/src/views/community-entry-surface.test.js`

- [ ] **Step 1: 写失败测试，约束社区一级入口文案与 CTA**

```js
assert.match(text, /社区主入口/);
assert.match(text, /看我的主题/);
assert.match(text, /发布我的观点/);
assert.match(text, /从策略、资讯和我的讨论继续承接/);
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd /Users/gjhan21/cursor/sercherai/client && node --test src/views/community-entry-surface.test.js`
Expected: FAIL，提示缺少一级入口强化文案

- [ ] **Step 3: 最小实现社区首屏强化**

在 `CommunityView.vue` 的首屏补足“一级主入口”定位文案、快捷承接说明和已登录用户常用操作 CTA，不重写社区数据流。

- [ ] **Step 4: 跑整体验证**

Run: `cd /Users/gjhan21/cursor/sercherai/client && node --test src/apps/pc/router/pc-community-watchlist-layout.test.js src/views/profile-watchlist-entry.test.js src/views/pc-watchlist-cta-routing.test.js src/views/community-entry-surface.test.js && npm run build`
Expected: 所有测试 PASS，PC 构建成功

- [ ] **Step 5: 提交**

```bash
git add client/src/views/CommunityView.vue client/src/views/community-entry-surface.test.js
git commit -m "feat: promote community as primary pc entry"
```
