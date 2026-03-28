# H5 Community Watchlist Restructure Design

**Goal:** 让 H5 端与 PC 端保持同一入口口径：社区保留一级入口，关注收口到“我的”页内的二级模块，不再让用户感知到独立一级关注页。

## Current Context

当前 H5 端仍保留独立 `/watchlist` 页面与独立场景文案：
- `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/router/index.js` 直接暴露 `watchlist` 路由
- `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/layouts/H5Layout.vue` 的 shell scene 仍把 `/watchlist` 当成独立内容区
- `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5ProfileView.vue` 还没有承担“我的关注 / 我的讨论”的二级模块承接职责
- `/Users/gjhan21/cursor/sercherai/client/src/apps/h5/views/H5WatchlistView.vue` 文案和入口感知仍更接近一级页面

这与 PC 端已经完成的口径不一致。用户要求 H5 端同样收口。

## Design Decision

采用与 PC 一致的轻量收口方案，而不是重写 H5 关注页：

1. `/watchlist` 不再作为主入口，只保留兼容跳转
2. 所有 H5 “去关注/我的关注”入口统一先跳 `/profile?section=watchlist`
3. `H5ProfileView` 新增“我的二级模块”区块，明确承接：
   - 我的关注
   - 我的讨论
4. `H5WatchlistView` 保留，但角色改为“我的 > 我的关注详情”，不再承担一级频道职责

这样可以保持改动最小，同时统一用户心智：
- 发现与参与讨论 -> 社区
- 个人回访与关注管理 -> 我的

## Architecture

### 1. Shared profile section helper

复用一套轻量 helper 生成 profile section 路由：
- `buildProfileModuleRoute("watchlist")`
- `buildProfileModulePath("watchlist")`
- `buildProfileModuleRedirectPath("watchlist")`

H5 和 PC 共用这套 helper，避免再次出现路径分叉。

### 2. H5 router behavior

`/watchlist` 改成兼容 redirect：
- 旧链接仍能访问
- 实际落点统一为 `/profile?section=watchlist`

不新增 H5 底部 tab，不把关注升回一级入口。

### 3. H5 profile as owner page

`H5ProfileView` 新增二级模块承接区：
- 模块卡片：我的关注 / 我的讨论
- 当 `route.query.section === "watchlist"` 时：
  - 当前模块高亮
  - 页面滚动到对应承接卡片或关注入口区
  - CTA 文案强调“进入关注详情”，避免一级页感知

### 4. H5 watchlist as detail page

`H5WatchlistView` 继续保留详情能力，但语义改为：
- 我的 > 我的关注详情
- 从我的进入，而不是站内一级频道

这要求改写：
- 顶部 tagline / hero 文案
- sticky CTA 文案
- 空态和引导动作

### 5. Cross-page CTA normalization

H5 下列页面中的“关注”入口全部收口到 `/profile?section=watchlist`：
- H5 首页
- H5 策略页
- H5 资讯页
- H5 会员页
- H5 认证页 redirect
- 任何 `/watchlist` 直接跳转逻辑

## UX Rules

- 底部导航只保留：首页 / 资讯 / 策略 / 会员 / 我的
- 社区不是 H5 底部 tab，本轮不扩 H5 社区入口，只保证“关注收口到我的”口径一致
- H5 我的页承担“账户 + 个人回访入口”双职责
- 关注详情仍可存在，但必须被表达为从“我的”进入的二级详情

## Error Handling

- `section` 非法值时回退到 `overview`
- `/watchlist` 旧路径必须稳定跳转，不允许 404
- 未登录进入 `/profile?section=watchlist` 时仍按原逻辑去 `/auth`，但 redirect 保持同一路径

## Testing Strategy

需要覆盖三类验证：

1. 路由验证
- `/watchlist` -> `/profile?section=watchlist`
- `/profile?section=watchlist` 正常打开 H5 Profile

2. 文案和结构验证
- H5 Profile 出现“我的二级模块”“我的关注”“我的讨论”
- H5 Watchlist 出现“我的 > 我的关注详情”类文案，而不是一级页感知

3. 入口收口验证
- H5 首页 / 策略 / 资讯 / 会员 / 认证中的关注入口不再直接指向独立 `/watchlist`

## Non-Goals

本轮不做：
- H5 社区完整实现
- H5 底部导航新增社区 tab
- 重写 H5 Watchlist 的数据模型
- 后台权限跳转问题修复
