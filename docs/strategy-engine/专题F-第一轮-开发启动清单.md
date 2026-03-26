# 专题F 第一轮（Legacy Matrix 与迁移收口）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不破坏双资产研究平台稳定性的前提下，把专题F细化成可直接按新旧入口、路由权限、客户端 fallback 主链开工的文件级任务表。

**Architecture:** 不做一次性硬切，而是通过 `Legacy Matrix -> 入口降级 -> 路由/权限/菜单统一 -> fallback 收缩 -> feature flag / rollback` 的顺序逐步收口。第一轮优先复用现有 `MarketCenterView`、Admin router、`admin-navigation.js`、Client `StrategyView.vue` / `MembershipView.vue` / fallback helper 以及后台 publish/archive/version compare 兼容链路。

**Tech Stack:** Go, Gin, Vue 3, npm build, Go test

---

最后更新: 2026-03-23  
状态: Planned

## 文档定位

这份清单是：

- [专题F-旧链路收缩与迁移收口.md](/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题F-旧链路收缩与迁移收口.md)

对应的第一轮可执行实施计划。

它不是专题F全部收官文档，而是“按当前新旧入口、路由权限与 fallback 主链直接开工”的文件级任务表。默认顺序固定为：

1. 包0：Legacy Matrix 基线盘点
2. 包1：后台入口降级与跳板化
3. 包2：路由 / 权限 / 菜单统一
4. 包3：Client fallback 收缩
5. 包4：feature flag 与 rollback path

## 开工前提

- 专题A~E 已完成主链升级，双资产研究平台已经能承接核心动作
- 所有收缩动作都必须有回退路径，不允许一次性删除旧入口
- 第一轮以“弱化旧入口 + 明确迁移状态”为主，不做彻底删库删接口
- `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js:155` 的权限跳转问题固定放在本专题包2处理，不提前插队

## 包0：Legacy Matrix 基线盘点

### Task 0: 盘点旧入口、旧页面、旧接口与 fallback 清单

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题F-第一轮-开发启动清单.md`
- Create: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题F-Legacy-Matrix.md`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/lib/admin-navigation.js`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/components/AppLayout.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/router/index.js`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/views/MembershipView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/lib/fallback-policy.js`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_archive_test.go`

- [ ] **Step 1: 冻结 Legacy Matrix 状态口径**

每条旧链路都必须标记为：

- `保留`
- `只读`
- `兜底`
- `准备删除`
- `已迁移`

- [ ] **Step 2: 先生成第一版 Legacy Matrix 文档**

要求：

- 至少覆盖旧后台入口、旧客户端 fallback、旧兼容接口、旧任务入口
- 每项记录责任页面、权限、回退路径、迁移目标
- 成为专题F后续包的唯一清单真相源

- [ ] **Step 3: 固化回归基线命令**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 兼容接口、publish/archive/version compare 相关测试保持通过

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: Admin 构建保持通过

Run: `cd /Users/gjhan21/cursor/sercherai/client && npm run build`  
Expected: Client 构建保持通过

## 包1：后台入口降级与跳板化

### Task 1: 让 MarketCenter 与旧研究入口转为跳板或说明页

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/admin-navigation.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemJobsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题F-Legacy-Matrix.md`

- [ ] **Step 1: 为后台入口降级写清迁移目标**

至少覆盖：

- `MarketCenter`
- 旧市场同步聚合入口
- 旧每日生成 / 调度入口
- 与新研究平台重叠的旧配置页

- [ ] **Step 2: 将旧入口改成跳板 / 说明 / 只读模式**

要求：

- 不再允许旧入口和新研究模块编辑同一核心对象
- 页面要显式提示“迁移去哪里、为何迁移、何时只读”
- 暂不删除菜单，仅调整优先级与交互角色

- [ ] **Step 3: 跑 Admin 构建回归**

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: 入口降级后后台构建通过

## 包2：路由 / 权限 / 菜单统一

### Task 2: 收口 Admin route、菜单与无权限跳转逻辑

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/admin-navigation.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/components/AppLayout.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/NoAccessView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/admin-navigation.test.js`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/lib/session.js`

- [ ] **Step 1: 先写路由/导航回归测试或校验清单**

覆盖：

- 无权限用户不再被重定向回 `/dashboard`
- `no-access` 成为统一兜底
- 菜单只显示当前可见入口
- 新旧双入口的 `navKey` 与默认跳转口径一致

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: 变更前记录当前编译基线

- [ ] **Step 2: 在本包内收口 `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js:155` 问题**

要求：

- 禁止无权限跳转统一硬指向 `/dashboard`
- 改为 `no-access` 或首个有权限入口
- 不修改其他专题的业务流，只处理路由权限口径

- [ ] **Step 3: 同步菜单与 layout**

要求：

- `admin-navigation.js` 成为导航真相源
- `AppLayout.vue` 只消费统一菜单定义
- 老菜单项如果保留，必须有明显的“迁移中 / 只读 / 兜底”状态

- [ ] **Step 4: 跑 Admin 构建与手动校验**

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: 路由与导航构建通过

## 包3：Client fallback 收缩

### Task 3: 让客户端 fallback 只保留必要兼容，不再承接新字段缺口

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/fallback-policy.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/MembershipView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`

- [ ] **Step 1: 盘点 fallback 仍在承接的新字段缺口**

至少检查：

- explanation 新字段
- version diff / archive 对比
- 会员页 latest strategy snapshot
- profile / message / quota 等 demo fallback 入口

- [ ] **Step 2: 收口 fallback 责任边界**

要求：

- fallback 只负责老记录兼容与 demo 模式
- 新记录统一走共享 helper 与正式字段
- fallback 不再替新主链缺字段兜底

- [ ] **Step 3: 跑 Client + Backend 回归**

Run: `cd /Users/gjhan21/cursor/sercherai/client && npm run build`  
Expected: 客户端构建通过

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: explanation / archive / version compare 相关测试通过

## 包4：feature flag 与 rollback path

### Task 4: 为每个收缩动作补显式开关与回退路径

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/admin-navigation.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/fallback-policy.js`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/legacy_migration_flag.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/legacy_migration_flag_repo.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/legacy_migration_flag_repo_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题F-Legacy-Matrix.md`

- [ ] **Step 1: 为迁移开关写 failing tests**

覆盖：

- 页面级开关
- 菜单级开关
- fallback 行为开关
- rollback path 读取与回退

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestLegacyMigrationFlag'`  
Expected: 因迁移开关 repo 尚不存在而失败

- [ ] **Step 2: 新增最小 migration flag 模型**

要求：

- 不做一次性通用 feature flag 平台
- 只覆盖专题F的迁移收口动作
- 每个收缩动作在 Legacy Matrix 中都要有 rollback path

- [ ] **Step 3: 跑前后端回归**

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: 后台构建通过

Run: `cd /Users/gjhan21/cursor/sercherai/client && npm run build`  
Expected: 客户端构建通过

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 兼容链路测试通过

- [ ] **Step 4: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/admin/src/router/index.js /Users/gjhan21/cursor/sercherai/admin/src/lib/admin-navigation.js /Users/gjhan21/cursor/sercherai/client/src/lib/fallback-policy.js /Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/legacy_migration_flag.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/legacy_migration_flag_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/legacy_migration_flag_repo_test.go /Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题F-Legacy-Matrix.md
git commit -m "feat: add legacy migration flags and rollback matrix"
```

## 第一轮默认验收清单

### Backend

- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo/...`
- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`

### Admin / Client

- `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`
- `cd /Users/gjhan21/cursor/sercherai/client && npm run build`

## 第一轮明确不并入

- 一次性删除所有旧接口
- 无回退的大切换
- 不经 feature flag 的菜单硬切
- 把所有 legacy 兼容逻辑在第一轮彻底删光

## 完成后文档回写要求

第一轮完成后，至少回写：

- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题F-旧链路收缩与迁移收口.md`
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/README.md`
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/未完成专题细化任务表.md`
