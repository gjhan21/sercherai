# 专题A 第一轮（Provider Registry 与供应商治理）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有多源同步与 truth 主链基线上，完成专题A第一轮的 provider registry、capability schema、统一 routing 摘要、基础质量评分和 Admin 数据治理台第一版。

**Architecture:** 保持现有 `DataSource + 多源同步 + truth 重建 + 上下文消费` 主链不推翻，通过新增治理表、治理模型和治理接口，把“散落在 repo 与前端判断里的 provider 规则”收口到统一治理层。Admin 继续沿用 `/api/v1/admin/data-sources/*` 路由组，不新增第二套后台模块，而是在现有数据源页上升级治理视图和操作路径。

**Tech Stack:** Go, Gin, MySQL, Vue 3, Element Plus, Go test, npm build

---

最后更新: 2026-03-23  
状态: Planned

## 文档定位

这份清单是：

- [专题A-多源市场数据统一模型与供应商治理.md](/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题A-多源市场数据统一模型与供应商治理.md)

对应的第一轮可执行实施计划。

它不是专题A全部收官文档，而是“从现有基线进入开发”时的第一轮启动真相源。默认顺序固定为：

1. 包0：基线校验与字段盘点
2. 包1：Provider Registry 与 Capability Schema
3. 包2：Routing / Fallback 治理层
4. 包3：Quality / Freshness / Trust Scoring
5. 包4：Admin 数据治理台第一版

## 开工前提

- 阶段8、阶段9均已 `Done`
- 当前多源同步、truth 重建、上下文消费链路已可运行
- `DataSourcesView` 和 `MarketCenterView` 继续保留，作为迁移期入口
- 默认不删除现有 config key、现有同步接口和现有 quality log
- 第一轮不动 Client，不直接改 `strategy-engine` 协议，只给下游补治理摘要

## 包0：基线校验与字段盘点

### Task 0: 盘点现有 provider 规则并锁定第一轮范围

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题A-第一轮-开发启动清单.md`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_admin.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`

- [ ] **Step 1: 确认第一轮 provider 范围**

范围固定为：

- `TUSHARE`
- `AKSHARE`
- `TICKERMD`
- `MYSELF`
- `MOCK`

- [ ] **Step 2: 确认第一轮治理数据域**

范围固定为：

- `STOCK + DAILY_BARS`
- `FUTURES + DAILY_BARS`
- `STOCK + INSTRUMENT_MASTER`
- `FUTURES + INSTRUMENT_MASTER`
- `STOCK/FUTURES + NEWS_ITEMS`
- `FUTURES + FUTURES_INVENTORY`

- [ ] **Step 3: 记录现有散落逻辑位置**

重点记录：

- source priority config key
- `resolveRequestedMarketSourceKeys`
- provider 分支 fetch 逻辑
- `supportsSyncKind` 前端硬编码
- `strategy-engine` 上下文中的 fallback / warning 判断

- [ ] **Step 4: 作为后续所有包的回归基线**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 现有市场数据、多源、truth、上下文相关测试保持通过

## 包1：Provider Registry 与 Capability Schema

### Task 1: 新增治理表与模型骨架

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260323_00_market_provider_governance.sql`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/market_provider_governance.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/market_data.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_provider_governance_repo_test.go`

- [ ] **Step 1: 先写 failing repo/model 测试**

覆盖：

- provider registry 读取
- capability matrix 读取
- routing policy 初始化读取
- 缺省 provider 种子数据可回填

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestMarketProviderGovernance'`  
Expected: 因 repo / model / schema 尚不存在而失败

- [ ] **Step 2: 新增 migration**

第一轮至少落以下表：

- `market_provider_registry`
- `market_provider_capabilities`
- `market_provider_routing_policies`
- `market_provider_quality_scores`

要求：

- 只增不改，不破坏现有 `data_sources`、`market_data_quality_logs`、truth 表
- migration 中写入第一版 provider 种子记录与 capability 初始数据

- [ ] **Step 3: 新增治理模型**

至少新增以下模型：

- `MarketProviderRegistry`
- `MarketProviderCapability`
- `MarketProviderRoutingPolicy`
- `MarketProviderQualityScore`
- `MarketProviderGovernanceOverview`

- [ ] **Step 4: 跑 repo/model 测试**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestMarketProviderGovernance'`  
Expected: schema/model 层测试通过

- [ ] **Step 5: Commit**

```bash
git add backend/migrations/20260323_00_market_provider_governance.sql backend/internal/growth/model/market_provider_governance.go backend/internal/growth/model/models.go backend/internal/growth/model/market_data.go backend/internal/growth/repo/market_provider_governance_repo_test.go
git commit -m "feat: add market provider governance schema"
```

### Task 2: 新增 repo 与 in-memory 治理读写能力

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_provider_governance_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/inmemory_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_provider_governance_repo_test.go`

- [ ] **Step 1: 扩展 repo interface**

至少暴露：

- `AdminGetMarketProviderGovernanceOverview`
- `AdminListMarketProviderCapabilities`
- `AdminListMarketProviderRoutingPolicies`
- `AdminUpsertMarketProviderRoutingPolicy`

- [ ] **Step 2: 实现 MySQL repo**

要求：

- registry 与 capability 读取支持按 provider / asset_class / data_kind 过滤
- routing policy 读取必须返回主源、备源、fallback_allowed、mock_allowed、quality_threshold
- overview 需要汇总 `data_sources`、quality summary、latest health、latest truth 摘要

- [ ] **Step 3: 实现 in-memory repo**

要求：

- 单元测试可以不依赖 MySQL migration 就跑通
- 预置 demo provider / capability / policy 样本

- [ ] **Step 4: 跑 repo 全量回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo/...`  
Expected: 新 repo 测试通过，现有 market/truth/context 测试不回归

- [ ] **Step 5: Commit**

```bash
git add backend/internal/growth/repo/market_provider_governance_repo.go backend/internal/growth/repo/interfaces.go backend/internal/growth/repo/inmemory_repo.go backend/internal/growth/repo/mysql_repo.go backend/internal/growth/repo/market_provider_governance_repo_test.go
git commit -m "feat: add market provider governance repo"
```

## 包2：Routing / Fallback 治理层

### Task 3: 把 provider 选择逻辑从硬编码切到 capability + routing policy

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/market_data.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_test.go`

- [ ] **Step 1: 为 routing layer 写 failing tests**

覆盖：

- provider capability 不支持时直接跳过
- routing policy 主源失败后自动按备源顺序 fallback
- `MOCK` 在未显式允许时不可进入最终选择
- context meta 回填 `selected_source / fallback_chain / decision_reason`

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestMarketRouting|TestStrategyContext'`  
Expected: 因治理层尚未接入而失败

- [ ] **Step 2: 提炼统一 resolved source 决策器**

要求：

- 统一替换散落的 provider 支持判断
- 同步、主数据、truth、context 至少共用同一套 routing 决策入口
- 返回结构化 routing 决策，不只返回 source key 列表

- [ ] **Step 3: 扩展输出摘要**

至少补到 `MarketSyncResult` / truth summary / context meta：

- `selected_source`
- `fallback_chain`
- `decision_reason`
- `policy_key`

- [ ] **Step 4: 跑相关 repo 回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo/...`  
Expected: 多源同步、主数据、上下文测试通过

- [ ] **Step 5: Commit**

```bash
git add backend/internal/growth/repo/market_data_multi_source.go backend/internal/growth/repo/market_instrument_master_data.go backend/internal/growth/repo/market_data_admin.go backend/internal/growth/repo/strategy_engine_context_repo.go backend/internal/growth/model/market_data.go backend/internal/growth/repo/market_data_multi_source_test.go backend/internal/growth/repo/market_instrument_master_data_test.go backend/internal/growth/repo/strategy_engine_context_test.go
git commit -m "feat: unify market provider routing policy"
```

## 包3：Quality / Freshness / Trust Scoring

### Task 4: 在现有 quality log 之上增加 provider 质量画像

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_admin.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_provider_governance_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/market_data_admin_service.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_status_truth_admin_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_provider_governance_repo_test.go`

- [ ] **Step 1: 写 failing tests**

覆盖：

- 按 provider + asset_class + data_kind 聚合 freshness / stability / trust / overall score
- quality summary 继续兼容旧接口
- overview 可同时返回旧质量摘要和新 provider 画像

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestMarketQuality|TestMarketProviderGovernance'`  
Expected: 新评分字段断言失败

- [ ] **Step 2: 扩展质量聚合**

要求：

- 优先复用现有 `market_data_quality_logs`
- 不额外引入复杂计算任务
- 第一版分值要能给出 `score_reasons`

- [ ] **Step 3: Overview 聚合新增治理建议**

至少返回：

- `freshness_score`
- `stability_score`
- `trust_score`
- `overall_score`
- `latest_issue_code`
- `governance_suggestion`

- [ ] **Step 4: 跑后台相关测试**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: admin market data handler / repo / router 测试通过

- [ ] **Step 5: Commit**

```bash
git add backend/internal/growth/model/market_data_admin.go backend/internal/growth/repo/market_data_admin.go backend/internal/growth/repo/market_provider_governance_repo.go backend/internal/growth/service/market_data_admin_service.go backend/internal/growth/repo/market_status_truth_admin_test.go backend/internal/growth/repo/market_provider_governance_repo_test.go
git commit -m "feat: add market provider quality scoring"
```

## 包4：Admin 数据治理台第一版

### Task 5: 扩展后台接口并升级数据源页为治理台

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/market_data_admin_service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/market_data_admin_handler_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/market-data-admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/DataSourcesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/data-sources-view.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`

- [ ] **Step 1: 先写 handler 和前端 failing tests**

后端新增接口至少包括：

- `GET /api/v1/admin/data-sources/governance/overview`
- `GET /api/v1/admin/data-sources/governance/capabilities`
- `GET /api/v1/admin/data-sources/governance/routing-policies`
- `PUT /api/v1/admin/data-sources/governance/routing-policies/:policy_key`

前端至少断言：

- 数据源页出现治理总览区
- 能力矩阵可读
- routing policy 可编辑主源 / 备源 / mock 允许开关

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/handler ./router/...`  
Expected: 新接口未注册前失败

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm test -- --runInBand data-sources-view.test.js market-data-admin.test.js`  
Expected: 新治理区断言失败

- [ ] **Step 2: 扩展后端 service / handler / router**

要求：

- 继续沿用 `/admin/data-sources` 路由组
- 不新建第二个一级菜单
- 保持原有同步、health check、truth 重建接口全部兼容

- [ ] **Step 3: 升级 `DataSourcesView`**

页面结构调整为：

- 供应商总览
- 能力矩阵
- 路由与真相源治理
- 质量与异常
- 原有同步与重建操作区

- [ ] **Step 4: `MarketCenterView` 只保留治理摘要和跳转**

要求：

- 不再复制一套 provider 判断
- 优先跳到 `数据源管理` 处理治理动作
- 仍保留迁移期概览提示

- [ ] **Step 5: 跑前后端构建与测试**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 通过

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: 通过

- [ ] **Step 6: Commit**

```bash
git add backend/internal/growth/service/service.go backend/internal/growth/service/market_data_admin_service.go backend/internal/growth/handler/market_data_admin_handler.go backend/router/router.go backend/internal/growth/handler/market_data_admin_handler_test.go admin/src/api/admin.js admin/src/lib/market-data-admin.js admin/src/views/DataSourcesView.vue admin/src/views/data-sources-view.test.js admin/src/views/MarketCenterView.vue
git commit -m "feat: upgrade data sources admin into governance console"
```

## 联调与验收闸门

### 闸门1：provider 治理底座可读

- migration 落库成功
- provider registry / capability / routing policy 可通过 API 读取
- in-memory 与 MySQL 两条 repo 都可通过测试

### 闸门2：routing 主链不回归

- 股票、期货、主数据、资讯同步继续可用
- `strategy-engine` 上下文继续优先消费 truth
- fallback warning 不丢失，且新增 `decision_reason`

### 闸门3：治理页可用

- `数据源管理` 页面不只是操作台，而能看到治理总览
- 能力矩阵、质量画像、routing policy 可读
- 主源 / 备源调整后可反映到后端 policy

## 第一轮默认验证命令

- Backend: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`
- Admin: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`

## 第一轮明确不并入

- 不做 Client 改版
- 不做 `strategy-engine` 协议大改
- 不做供应商成本中心和计费
- 不做事件审核台
- 不做期货深层主数据专题化能力

## 完成后文档回写要求

第一轮包0 到包4全部完成后，至少同步回写：

- [专题A-多源市场数据统一模型与供应商治理.md](/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题A-多源市场数据统一模型与供应商治理.md)
- [README.md](/Users/gjhan21/cursor/sercherai/docs/strategy-engine/README.md)

回写内容至少包括：

- 第一轮已完成范围
- 当前治理台真实入口
- 当前统一 routing 摘要字段
- 明确留给专题A第二轮或专题B/C的后续项
