# H5 Community Watchlist Restructure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 H5 端的关注入口与 PC 端一致：关注统一收口到“我的”页二级模块，`/watchlist` 只保留兼容跳转。

**Architecture:** 复用共享 `profile-modules` helper 统一生成 `/profile?section=watchlist`，H5 router 只做兼容重定向；`H5ProfileView` 新增二级模块承接；跨页 CTA 统一改为先进入“我的”；`H5WatchlistView` 降级为“我的 > 我的关注详情”。

**Tech Stack:** Vue 3、Vue Router、现有 H5 组件体系、Node built-in test、Vite。

---

### Task 1: H5 路由先定义目标行为

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/router/index.js`
- Test: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/router/h5-watchlist-profile-layout.test.js`

- [ ] **Step 1: 写失败测试，定义 H5 watchlist 收口行为**

```js
assert.match(text, /path: "watchlist"[\s\S]*redirect: buildProfileModuleRoute\("watchlist"\)/)
assert.doesNotMatch(text, /path: "watchlist", name: "h5-watchlist", component:/)
```

- [ ] **Step 2: 跑测试确认失败**

Run: `node --test src/apps/h5/router/h5-watchlist-profile-layout.test.js`
Expected: FAIL，说明路由还未收口

- [ ] **Step 3: 最小实现路由重定向**

在 `index.js` 引入共享 profile helper，把 `/watchlist` 改成 redirect 到 `/profile?section=watchlist`。

- [ ] **Step 4: 再跑测试确认通过**

Run: `node --test src/apps/h5/router/h5-watchlist-profile-layout.test.js`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add /Users/gjhan21/cursor/sercherai/client/src/apps/h5/router/index.js /Users/gjhan21/cursor/sercherai/client/src/apps/h5/router/h5-watchlist-profile-layout.test.js
git commit -m "test: pin h5 watchlist redirect through profile"
```

### Task 2: H5 我的页补二级模块承接

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5ProfileView.vue`
- Test: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/h5-profile-module-entry.test.js`

- [ ] **Step 1: 写失败测试，定义 H5 我的页结构**

```js
assert.match(text, /我的二级模块/)
assert.match(text, /我的关注/)
assert.match(text, /我的讨论/)
```

- [ ] **Step 2: 跑测试确认失败**

Run: `node --test src/apps/h5/views/h5-profile-module-entry.test.js`
Expected: FAIL

- [ ] **Step 3: 最小实现模块卡片和 section 聚焦逻辑**

增加：
- 二级模块区块
- `section=watchlist` 高亮
- 滚动到关注承接区

- [ ] **Step 4: 再跑测试确认通过**

Run: `node --test src/apps/h5/views/h5-profile-module-entry.test.js`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add /Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5ProfileView.vue /Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/h5-profile-module-entry.test.js
git commit -m "feat: add h5 profile module entry for watchlist"
```

### Task 3: H5 跨页 CTA 收口到“我的”

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5HomeView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5NewsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5MembershipView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5AuthView.vue`
- Test: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/router/h5-watchlist-entry-routing.test.js`

- [ ] **Step 1: 写失败测试，锁定所有关注入口新路径**

```js
assert.match(text, /buildProfileModuleRoute\("watchlist"\)|buildProfileModulePath\("watchlist"\)|buildProfileModuleRedirectPath\("watchlist"\)/)
assert.doesNotMatch(text, /"\/watchlist"/)
```

- [ ] **Step 2: 跑测试确认失败**

Run: `node --test src/apps/h5/router/h5-watchlist-entry-routing.test.js`
Expected: FAIL

- [ ] **Step 3: 最小实现 CTA 和 auth redirect 收口**

统一使用共享 helper 生成：
- 导航跳转
- auth redirect
- 会员/资讯/策略/首页入口

- [ ] **Step 4: 再跑测试确认通过**

Run: `node --test src/apps/h5/router/h5-watchlist-entry-routing.test.js`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add /Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5HomeView.vue /Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5StrategyView.vue /Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5NewsView.vue /Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5MembershipView.vue /Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5AuthView.vue /Users/gjhan21/cursor/sercherai/client/src/apps/h5/router/h5-watchlist-entry-routing.test.js
git commit -m "feat: route h5 watchlist entry through profile"
```

### Task 4: H5 Watchlist 页降级为“我的 > 我的关注详情”

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5WatchlistView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/shell-meta.js`
- Test: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/h5-watchlist-demotion.test.js`

- [ ] **Step 1: 写失败测试，定义文案降级目标**

```js
assert.match(text, /我的 > 我的关注详情|我的 > 我的关注/)
assert.doesNotMatch(text, /独立一级频道|一级入口/)
```

- [ ] **Step 2: 跑测试确认失败**

Run: `node --test src/apps/h5/views/h5-watchlist-demotion.test.js`
Expected: FAIL

- [ ] **Step 3: 最小实现文案与 shell scene 调整**

改写：
- 页面 tagline
- hero copy
- sticky CTA
- shell scene 对 `/watchlist` 的表达

- [ ] **Step 4: 再跑测试确认通过**

Run: `node --test src/apps/h5/views/h5-watchlist-demotion.test.js`
Expected: PASS

- [ ] **Step 5: 提交**

```bash
git add /Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5WatchlistView.vue /Users/gjhan21/cursor/sercherai/client/src/apps/h5/lib/shell-meta.js /Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/h5-watchlist-demotion.test.js
git commit -m "feat: demote h5 watchlist to profile detail"
```

### Task 5: 总回归

**Files:**
- Review: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/router/index.js`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5ProfileView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5WatchlistView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5HomeView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5StrategyView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5NewsView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5MembershipView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5AuthView.vue`

- [ ] **Step 1: 跑 H5 相关测试集**

Run: `cd /Users/gjhan21/cursor/sercherai/client && node --test src/apps/h5/router/h5-watchlist-profile-layout.test.js src/apps/h5/router/h5-watchlist-entry-routing.test.js src/apps/h5/views/h5-profile-module-entry.test.js src/apps/h5/views/h5-watchlist-demotion.test.js`
Expected: PASS

- [ ] **Step 2: 跑 H5 构建**

Run: `cd /Users/gjhan21/cursor/sercherai/client && npm run build:h5`
Expected: build 成功

- [ ] **Step 3: 手验页面流转**

Check:
- `/watchlist` -> `/profile?section=watchlist`
- `/profile?section=watchlist`
- 从首页/资讯/策略/会员跳关注时，先到我的页而不是独立页

- [ ] **Step 4: 提交**

```bash
git add /Users/gjhan21/cursor/sercherai/client/src/apps/h5 /Users/gjhan21/cursor/sercherai/client/src/lib/profile-modules.js
git commit -m "feat: align h5 watchlist entry with profile"
```
