# 专题B 第一轮（事件真相源与审核台）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在现有新闻同步、股票研究、`strategy-graph` 和 explanation 主链之上，补齐“新闻 -> 事件 -> 图谱 -> 审核 -> explanation”文件级实施清单，让专题B可以直接按代码入口开工。

**Architecture:** 不推翻现有 `market news -> stock selection -> graph summary -> client explanation` 主链，而是在 Go 后端新增事件真相源与审核状态，在 `strategy-graph` 增加 reviewed event edges，在 Admin 增加事件审核页，在 Client explanation 增加事件证据卡。第一轮保持 run snapshot 兼容，事件真相源作为增强层逐步接入。

**Tech Stack:** Go, Gin, MySQL, Vue 3, Python, FastAPI, pytest, Go test, npm build

---

最后更新: 2026-03-23  
状态: Done（第一轮已完成）

## 文档定位

这份清单是：

- [专题B-股票事件图谱与资讯语义工作台.md](/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题B-股票事件图谱与资讯语义工作台.md)

对应的第一轮可执行实施计划。

它不是专题B全部收官文档，而是“按当前新闻、图谱、股票 explanation 主链直接开工”的文件级任务表。默认顺序固定为：

1. 包0：事件基线盘点与范围冻结
2. 包1：事件真相源 schema 与标准化模型
3. 包2：资讯抽取、聚类与 reviewed event graph
4. 包3：Admin 事件审核台第一版
5. 包4：股票 explanation 事件证据卡接入

## 开工前提

- 专题A 第一轮已完成，资讯与图谱消费可带治理摘要
- `strategy-graph` 继续保持“可降级，不阻塞 run / publish / explanation”原则
- 第一轮不删除旧新闻表、不删除 run snapshot、不强制历史记录回填 reviewed events
- Client 继续复用现有 explanation helper 和历史版本对比逻辑

## 当前落地状态

- 已落地事件真相源最小 schema / model / repo，补齐 `stock_event_clusters / items / entities / edges / reviews` 与查询过滤能力
- 已在新闻同步后生成草稿事件并完成聚类去重；cluster 默认进入 `CLUSTERED + PENDING`，低置信度或泛化 `NEWS` 事件进入 `HIGH` 优先级审核队列
- 已补齐 `GET /api/v1/admin/stock-selection/events`、`GET /api/v1/admin/stock-selection/events/:id`、`POST /api/v1/admin/stock-selection/events/:id/review`，并接入 `ReviewCenterView` 的 `STOCK_EVENT` 审核任务
- 已新增 `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionEventsView.vue`，形成股票研究模块下的事件审核台
- 已新增 `POST /internal/v1/graph/reviewed-events`；审核通过事件会以 `reviewed-event-<cluster_id>` 伪快照写入图服务，从而复用现有 `subgraph` 查询链路而不破坏 run snapshot 兼容性
- 已在股票 explanation 中新增 `related_events` 与 `event_evidence_cards`，客户端 `StrategyView.vue` 已展示“事件证据卡”，旧 `evidence_cards / graph_summary` 保持兼容

## 本轮实现备注

- `strategy-graph` 第一轮没有新建独立 reviewed-event repo 表达层，而是复用既有 snapshot 存取协议承接审核通过事件
- 审核通过写图失败时采用降级策略：审核结果仍成功落库，warning 回写到 cluster metadata，Admin 可提示稍后重试
- `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js:155` 的权限跳转问题仍未并入 Topic B，本轮没有处理

## 包0：事件基线盘点与范围冻结

### Task 0: 锁定第一轮事件范围与现有主链入口

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题B-第一轮-开发启动清单.md`
- Review: `/Users/gjhan21/cursor/sercherai/backend/router/router.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/stock_selection_admin_handler.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/strategy_graph_admin_handler.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_graph_repo.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_graph_client.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Review: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/strategy_graph_service.go`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionGraphView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionCandidatesView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/client/src/views/NewsView.vue`
- Review: `/Users/gjhan21/cursor/sercherai/services/strategy-graph/app/api/routes_graph.py`

- [ ] **Step 1: 冻结第一轮事件类型与实体范围**

范围固定为：

- 事件类型：`NEWS / ANNOUNCEMENT / POLICY / EARNINGS / INDUSTRY_THEME / SUPPLY_CHAIN_EVENT`
- 标准实体：`COMPANY / SECTOR / TOPIC / POLICY / EARNINGS_REPORT / SUPPLY_CHAIN_NODE`
- 事件状态：`raw / clustered / reviewed / published / rejected`

- [ ] **Step 2: 标记当前散落逻辑位置**

至少记录：

- 资讯同步与新闻主链入口
- run snapshot 图谱构建与查询入口
- 股票候选 explanation / evidence card 入口
- 审核中心与股票审核页中可复用的审核动作入口

- [ ] **Step 3: 固化第一轮回归基线命令**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 股票研究、图谱代理、资讯与 explanation 相关测试保持通过

Run: `cd /Users/gjhan21/cursor/sercherai/services/strategy-graph && ./.venv/bin/pytest -q`  
Expected: 图服务现有 API 与 repo 测试保持通过

## 包1：事件真相源 schema 与标准化模型

### Task 1: 新增事件真相源 migration、模型与 repo 骨架

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/migrations/20260323_01_stock_event_truth_source.sql`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_event_truth.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_event_truth_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/inmemory_repo.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_event_truth_repo_test.go`

- [ ] **Step 1: 先写 failing repo/model 测试**

覆盖：

- 事件入库与读取
- cluster 合并与成员关系
- entity / edge 读写
- review 状态流转
- symbol / sector / topic 过滤查询

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestStockEventTruth'`  
Expected: 因 migration / model / repo 尚不存在而失败

- [ ] **Step 2: 新增 migration**

第一轮至少落以下表：

- `stock_event_clusters`
- `stock_event_items`
- `stock_event_entities`
- `stock_event_edges`
- `stock_event_reviews`

要求：

- 只增不改，不删除 `market_news_items` 与现有 graph snapshot 表
- 事件真相源主键与新闻原始 ID 保持可追溯映射
- `published / rejected` 状态不替代旧新闻发布状态，只服务研究链

- [ ] **Step 3: 新增统一模型与 repo interface**

至少新增：

- `StockEventCluster`
- `StockEventItem`
- `StockEventEntity`
- `StockEventEdge`
- `StockEventReview`
- `StockEventQuery`

- [ ] **Step 4: 跑 repo/model 回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestStockEventTruth'`  
Expected: schema/model 层测试通过

- [ ] **Step 5: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/backend/migrations/20260323_01_stock_event_truth_source.sql /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/stock_event_truth.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_event_truth_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/models.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/interfaces.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/inmemory_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_event_truth_repo_test.go
git commit -m "feat: add stock event truth source schema"
```

## 包2：资讯抽取、聚类与 reviewed event graph

### Task 2: 在资讯同步后补事件抽取与聚类链路

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/stock_event_service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/market_data_multi_source.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/admin_growth_handler.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_event_truth_repo_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/stock_selection_admin_handler_test.go`

- [ ] **Step 1: 为抽取与聚类入口写 failing tests**

覆盖：

- 新闻同步后生成事件草稿
- 同主题新闻聚类与去重
- symbol / sector / topic 映射
- 低置信度事件进入 review queue

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... -run 'TestStockEvent|TestMarketNews'`  
Expected: 因抽取与聚类链路尚未接入而失败

- [ ] **Step 2: 新增服务层抽取/聚类编排**

要求：

- 保持现有新闻同步入口不变
- 资讯同步完成后可按批次调用事件标准化
- 失败时只影响事件增强，不阻塞新闻原始落库

- [ ] **Step 3: 跑后端回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 新闻同步与股票研究链路保持通过

### Task 3: 让 strategy-graph 承接 reviewed event edges

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_graph.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_graph_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_graph_client.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/service/strategy_graph_service.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/strategy_graph_admin_handler.go`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-graph/app/api/routes_graph.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-graph/app/schemas/graph.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-graph/app/repo/base.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-graph/app/repo/inmemory.py`
- Modify: `/Users/gjhan21/cursor/sercherai/services/strategy-graph/app/repo/neo4j_repo.py`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_graph_client_test.go`
- Test: `/Users/gjhan21/cursor/sercherai/services/strategy-graph/tests/test_graph_api.py`

- [ ] **Step 1: 为 reviewed events graph API 写 failing tests**

覆盖：

- reviewed event edges 写入
- event subgraph 查询
- run snapshot + reviewed event edges 组合读取
- 图服务降级时回退到现有 run snapshot

Run: `cd /Users/gjhan21/cursor/sercherai/services/strategy-graph && ./.venv/bin/pytest -q`  
Expected: 因新 schema / API 尚未接入而失败

- [ ] **Step 2: 扩展 Go 图谱代理与 Python 图服务协议**

要求：

- Go 后端仍作为唯一图谱查询入口
- Python 图服务只增加 reviewed event edges 协议，不破坏现有 snapshot API
- `subgraph` 查询要支持事件、股票、主题混合节点

- [ ] **Step 3: 跑图谱与后端双回归**

Run: `cd /Users/gjhan21/cursor/sercherai/services/strategy-graph && ./.venv/bin/pytest -q`  
Expected: 图服务测试通过

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: 图谱代理与股票研究链路测试通过

- [ ] **Step 4: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_graph.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_graph_repo.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_graph_client.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/service/strategy_graph_service.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/strategy_graph_admin_handler.go /Users/gjhan21/cursor/sercherai/services/strategy-graph/app/api/routes_graph.py /Users/gjhan21/cursor/sercherai/services/strategy-graph/app/schemas/graph.py /Users/gjhan21/cursor/sercherai/services/strategy-graph/app/repo/base.py /Users/gjhan21/cursor/sercherai/services/strategy-graph/app/repo/inmemory.py /Users/gjhan21/cursor/sercherai/services/strategy-graph/app/repo/neo4j_repo.py /Users/gjhan21/cursor/sercherai/services/strategy-graph/tests/test_graph_api.py
git commit -m "feat: add reviewed stock event graph edges"
```

## 包3：Admin 事件审核台第一版

### Task 4: 在股票研究后台增加事件审核页与治理动作

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionEventsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/router/index.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionReviewsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionGraphView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/review-action-dialog.js`
- Test: `/Users/gjhan21/cursor/sercherai/admin/src/lib/review-action-dialog.test.js`

- [ ] **Step 1: 在 `admin.js` 增加事件审核 API**

至少新增：

- `listStockEventClusters(params)`
- `getStockEventCluster(id)`
- `reviewStockEventCluster(id, payload)`
- `queryStockEventSubgraph(params)`

- [ ] **Step 2: 新增事件审核页**

要求：

- 作为股票研究模块新页面或二级标签页接入
- 支持聚类校正、标签修正、误判回收、主题/行业归属校正
- 复用现有审核弹窗交互，不再另造一套审核动作模式

- [ ] **Step 3: 跑 Admin 构建回归**

Run: `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`  
Expected: 股票研究后台新增事件审核页后编译通过

## 包4：股票 explanation 事件证据卡接入

### Task 5: 让客户端 explanation 可以读到 reviewed events 证据链

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/NewsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/api/news.js`

- [ ] **Step 1: 为 explanation 事件证据卡写 failing tests**

覆盖：

- explanation 中带出事件证据列表
- 能展示“为什么这条新闻影响这个股票/主题”
- 缺少 reviewed events 时仍回退到旧 evidence card / graph summary

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestStrategyClientExplanation'`  
Expected: 因事件证据字段尚未接入而失败

- [ ] **Step 2: 扩展 explanation 聚合与前端渲染**

要求：

- explanation payload 新增事件证据卡，但不删除旧 `graph_summary / evidence_cards`
- `StrategyView.vue` 默认展示事件链摘要
- `NewsView.vue` 可承接事件与股票/主题关联说明

- [ ] **Step 3: 跑后端 + Client 回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`  
Expected: explanation 与股票研究链路通过

Run: `cd /Users/gjhan21/cursor/sercherai/client && npm run build`  
Expected: 客户端构建通过

- [ ] **Step 4: Commit**

```bash
git add /Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go /Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go /Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue /Users/gjhan21/cursor/sercherai/client/src/views/NewsView.vue /Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js /Users/gjhan21/cursor/sercherai/client/src/api/news.js
git commit -m "feat: add stock event evidence cards"
```

## 第一轮默认验收清单

### Backend

- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo/...`
- `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/... ./router/...`

### Graph Service

- `cd /Users/gjhan21/cursor/sercherai/services/strategy-graph && ./.venv/bin/pytest -q`

### Admin / Client

- `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`
- `cd /Users/gjhan21/cursor/sercherai/client && npm run build`

## 第一轮明确不并入

- 全自动事件审核闭环
- 高频情绪量化生产化
- 对 C 端开放图谱探索社区
- 期货事件治理与股票事件真相源混做

## 完成后文档回写要求

第一轮完成后，至少回写：

- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/专题B-股票事件图谱与资讯语义工作台.md`
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/README.md`
- `/Users/gjhan21/cursor/sercherai/docs/strategy-engine/未完成专题细化任务表.md`
