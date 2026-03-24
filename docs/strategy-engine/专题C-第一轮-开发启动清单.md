# 专题C 第一轮（期货主数据画像与深层因子）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有期货真实上下文、inventory、term structure、spread、候选 explanation 主链之上，把专题C细化成可直接按代码入口开工的文件级任务表。

**Architecture:** 保持 `Go context -> strategy-engine futures pipeline -> publish/explanation -> Admin/Client evidence` 主链不推翻，通过补齐 instrument master v2、inventory/warehouse 深层因子、basis/spread/curve 深化和商品链证据卡，让期货研究从“能运行”升级到“证据层更像研究报告”。第一轮不改实盘执行策略，不改当前发布闭环。

**Tech Stack:** Go, Gin, MySQL, Python, pytest, Vue 3, npm build

---

最后更新: 2026-03-24  
状态: Done（专题C / C0 ~ C4 第一轮已完成）

## 本轮实际落地摘要

- 包0：完成期货主数据 / futures pipeline / explanation / Admin / Client 的主链盘点，并把本清单收口为代码入口级任务表
- 包1：已落 `futures_instrument_profile_v2` migration、repo、model 和 instrument master 扩展，期货主数据画像可独立沉淀
- 包2 / 包3：已把 inventory / warehouse / brand / grade / spread / curve / cross-contract linkage 推进到上下文、feature factory、pipeline、report 和测试
- 包4：已在 explanation 中补齐 `supply_chain_notes`、`structure_factor_summary`、`inventory_factor_summary`，并把商品链证据接进 Admin 候选页 / 因子页 / 图谱页，以及 Client `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- 第一轮验收已完成：`go test ./internal/growth/... ./router/...`、`pytest -q`、Admin build、Client build 均通过

## 文档定位

这份清单是：

- [专题C-期货深度因子与主数据画像.md](/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题C-期货深度因子与主数据画像.md)

对应的第一轮可执行实施计划。

它不是专题C全部收官文档，而是“按当前期货主数据、futures pipeline、候选 explanation 主链直接开工”的文件级任务表。默认顺序固定为：

1. 包0：画像与因子基线盘点
2. 包1：futures instrument master v2
3. 包2：inventory / warehouse 深层因子
4. 包3：basis / spread / term structure 深化
5. 包4：商品链证据与候选 explanation 接入

## 开工前提

- 当前继续复用已落地的期货真实上下文、多源底座与 publish/explanation 闭环；本轮不等待专题A先完成
- 阶段8 已完成真实期货上下文接入，当前 futures pipeline 已能消费 inventory 与结构字段
- 第一轮不引入期权专题、不做实盘执行、不重做期货后台页面骨架
- 现有 explanation 与 publish payload 继续兼容旧消费链路

## 包0：画像与因子基线盘点（Done）

### Task 0: 锁定第一轮范围与当前代码主链

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题C-第一轮-开发启动清单.md`
- Review: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/futures_selection_run_repo.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/futures_selection_profile_repo.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/futures_selection_admin_handler.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Review: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/seeds/futures_seed_loader.py`
- Review: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/features/futures_feature_factory.py`
- Review: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/pipelines/futures_strategy_pipeline.py`
- Review: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/reports/futures_report_builder.py`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionFactorsView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionCandidatesView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`

- [ ] **Step 1: 冻结第一轮画像与因子范围**

范围固定为：

- 画像字段：`commodity / delivery_place / warehouse / brand / grade / inventory_metric / contract_chain`
- 深层因子：`inventory_concentration / warehouse_shift / brand_grade_premium / basis_spread_curve / cross_contract_linkage`
- 证据产物：`factor_summary / evidence_cards / supply_chain_notes`

- [ ] **Step 2: 固化回归基线命令**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 期货上下文、运行中心、评估与 explanation 相关测试保持通过

Run: `cd /Users/gjhan21/cursor/sercherai/services/strategy-engine && ./.venv/bin/pytest -q`  
Expected: futures pipeline 现有测试保持通过

## 包1：futures instrument master v2（Done）

### Task 1: 建立期货主数据画像模型与 repo 骨架

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260323_02_futures_instrument_profile_v2.sql`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/futures_instrument_profile.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/futures_instrument_profile_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/inmemory_repo.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/futures_instrument_profile_repo_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data_test.go`

- [ ] **Step 1: 先写 failing repo/model 测试**

覆盖：

- 主数据画像入库与读取
- contract chain 组装
- place / warehouse / brand / grade 聚合
- 与现有 instrument master 兼容读取

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestFuturesInstrumentProfile|TestMarketInstrumentMaster'`  
Expected: 因 profile v2 schema / repo 尚不存在而失败

- [ ] **Step 2: 新增 migration 与模型**

要求：

- 只增不改，不替换现有 instrument master truth 表
- 用 profile v2 承接 place / warehouse / brand / grade 等现有字段
- 支持一个品种下多合约链与库存口径映射

- [ ] **Step 3: 跑后端 repo 回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo/...`  
Expected: 主数据与上下文相关 repo 测试通过

- [ ] **Step 4: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/backend/migrations/20260323_02_futures_instrument_profile_v2.sql /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/futures_instrument_profile.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/futures_instrument_profile_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/inmemory_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/futures_instrument_profile_repo_test.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_instrument_master_data_test.go
git commit -m "feat: add futures instrument profile v2"
```

## 包2：inventory / warehouse 深层因子（Done）

### Task 2: 把库存 / 仓单结构从解释字段推进到 feature factory

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/seeds/futures_seed_loader.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/features/futures_feature_factory.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/schemas/futures.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/reports/futures_report_builder.py`
- Test: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/tests/test_futures_seed_loader.py`
- Test: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/tests/test_futures_pipeline.py`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_test.go`

- [ ] **Step 1: 为深层 inventory factors 写 failing tests**

覆盖：

- inventory concentration / shift / persistence
- brand / grade premium summary
- evidence card 所需 factor summary 输出
- 缺少深层数据时仍回退到现有 explanation 字段

Run: `cd /Users/gjhan21/cursor/sercherai/services/strategy-engine && ./.venv/bin/pytest -q -k futures`  
Expected: 因新特征尚未接入而失败

- [ ] **Step 2: 扩展 Go 上下文与 Python feature factory**

要求：

- Go 后端负责统一上下文字段口径
- Python 不重新定义第二套仓单/库存语义
- report builder 要能把深层因子转为可读证据

- [ ] **Step 3: 跑服务侧回归**

Run: `cd /Users/gjhan21/cursor/sercherai/services/strategy-engine && ./.venv/bin/pytest -q`  
Expected: futures pipeline 测试通过

## 包3：basis / spread / term structure 深化（Done）

### Task 3: 扩展期货结构因子体系与运行摘要

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/features/futures_feature_factory.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/agents/basis_agent.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/scenarios/futures_scenario_engine.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/pipelines/futures_strategy_pipeline.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/app/domain/selectors/futures_selector.py`
- Test: `/Users/gjhan21/cursor/sercherai/services/strategy-engine/tests/test_futures_pipeline.py`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_context_candidates_test.go`

- [ ] **Step 1: 写结构因子深化的 failing tests**

覆盖：

- basis / spread / curve 因子增强
- 跨品种或上下游联动的最小字段输出
- scenario / selector 对新结构因子的消费
- report 中带出“为什么这个结构重要”

Run: `cd /Users/gjhan21/cursor/sercherai/services/strategy-engine && ./.venv/bin/pytest -q -k 'futures and not stock'`  
Expected: 因新结构因子尚未接入而失败

- [ ] **Step 2: 扩展上下文与 pipeline**

要求：

- 不破坏现有 publish payload 与风险控制输出
- 结构因子优先进入 factor summary，而不是直接堆到原始 JSON
- selector / report / agent 对新因子消费口径一致

- [ ] **Step 3: 跑 Go + Python 双回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 期货上下文与后台运行中心测试通过

Run: `cd /Users/gjhan21/cursor/sercherai/services/strategy-engine && ./.venv/bin/pytest -q`  
Expected: strategy-engine 测试通过

## 包4：商品链证据与候选 explanation 接入（Done）

### Task 4: 让 Admin / Client 期货候选详情读到更深证据卡

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionCandidatesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionFactorsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionGraphView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`

- [ ] **Step 1: 为期货 explanation 证据增强写 failing tests**

覆盖：

- evidence cards 新增仓单/库存/品牌/等级/结构因子摘要
- Admin 候选详情可读更深因子证据
- Client explanation 继续兼容旧记录 fallback

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestStrategyClientExplanation'`  
Expected: 因期货深层证据字段尚未接入而失败

- [ ] **Step 2: 扩展 explanation 聚合与前端渲染**

要求：

- 优先复用现有 evidence card 区块，不重做页面骨架
- 新证据字段进入统一 helper，避免股票/期货再走分叉 helper
- 缺少新字段时仍显示旧 summary，不报错

- [ ] **Step 3: 跑 Admin / Client / Backend 回归**

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: 期货后台页面构建通过

Run: `cd /Users/gjhan21/cursor/sercherai/client && npm run build`  
Expected: 客户端构建通过

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: explanation 与候选详情相关测试通过

- [ ] **Step 4: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go /Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionCandidatesView.vue /Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionFactorsView.vue /Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionGraphView.vue /Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue /Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js
git commit -m "feat: deepen futures evidence cards"
```

## 第一轮默认验收清单

### Backend

- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo/...`
- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`

### Strategy Engine

- `cd /Users/gjhan21/cursor/sercherai/services/strategy-engine && ./.venv/bin/pytest -q`

### Admin / Client

- `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`
- `cd /Users/gjhan21/cursor/sercherai/client && npm run build`

## 第一轮明确不并入

- 期权波动率曲面
- 高频 CTA 实盘执行
- 单独交易执行工作台
- 完整商品链知识图谱重建

## 完成后文档回写要求

第一轮完成后，至少回写：

- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题C-期货深度因子与主数据画像.md`
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/README.md`
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/未完成专题细化任务表.md`
