# Stock Futures Forecast L2 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改现有推荐发布主链、不新增独立分析页和独立后台中心的前提下，为股票推荐与期货策略的 `insight / version-history` 增加轻量关系快照、三情景推演、角色意见/veto、历史回看与 admin 复用式承接。

**Architecture:** 继续复用 `L1` explanation 主链、现有 `simulations / scenario_templates / agents` 骨架、现有 `StrategyClientExplanation` 与 version-history 读取链。`L2` 不新开接口路径，不另造第二套推演系统，而是在后端 explanation 生成链中引入轻量关系快照与稳定三情景 contract，再把结果透传到现有 client 页面和现有 admin 模块。

**Tech Stack:** Go, Gin, MySQL repo layer, existing strategy-engine runtime payloads, Vue 3 client, Vue 3 admin, existing `strategy_client_explanation` and `strategy-version` helpers.

---

## Execution Status (2026-03-29)

- `L2` 主线已在 `main` 落地，范围保持在既定 4 个读取入口内。
- Task 1 ~ Task 6 已完成，且保持“非破坏性扩展 explanation / history，不改推荐发布主链”的边界。
- 对应关键提交：
  - `3e8c383 feat: add l2 relationship snapshot contracts`
  - `d61b86c feat: add stable l2 scenario snapshots`
  - `1b29758 feat: add l2 agent opinions and veto meta`
  - `303c529 feat: wire l2 scenarios into insight and history`
  - `aef4465 feat: surface l2 scenarios in client views`
  - `f5cd039 feat: add l2 forecast visibility in admin`
- fresh verification：
  - `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo` -> PASS
  - `cd /Users/gjhan21/cursor/sercherai/client && npm run build` -> PASS
  - `cd /Users/gjhan21/cursor/sercherai/admin && npm run build` -> PASS

## Scope And Guardrails

- 本计划只覆盖 `L2`，不提前实现 `L3`。
- 本计划只覆盖 4 个真实入口：
  - 股票推荐 `insight`
  - 股票推荐 `version-history`
  - 期货策略 `insight`
  - 期货策略 `version-history`
- `L2` 只允许作为 explanation / history 的非破坏性扩展。
- `L2` 不允许：
  - 直接修改 candidate ranking
  - 直接修改 portfolio weight
  - 直接修改 publish review 主流程
  - 新增独立 analysis 页作为前提
  - 新增独立预测增强后台中心
- `L2` 主情景 contract 固定为三情景：`bull / base / bear`。
- 现有 `shock` 或其它场景如果仍存在，只作为兼容输入保留；首批 `L2` 输出 contract 不以它们为主。
- admin 本轮一起做，但只能复用现有模块：
  - `SystemConfigsView.vue`
  - `MarketCenterView.vue`
  - `ReviewCenterView.vue`
  - `stock-selection/*`
  - `futures-selection/*`

## Current Baseline

- `L1` 已落地并收口：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
  - `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
- 当前已经有一部分 scenario / simulation 骨架，不是从零开始：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_client.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/futures_selection_run_repo.go`
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_runtime_repo.go`
  - `/Users/gjhan21/cursor/sercherai/admin/components/StrategyEngineConfigPanel.vue`
- 当前 client 已经读取部分 `simulations`：
  - `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- 当前 admin 已经有 `L1` 配置与摘要嵌入：
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`

## File Map

### Backend Contracts And Builders

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
  - 为 explanation 与 history 增加 `scenario_snapshots / scenario_meta / agent_opinions / relationship_snapshot` 等 `L2` 字段。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
  - 把新的关系快照、三情景与角色意见合并进 explanation / history。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_relationships.go`
  - 从现有 tags / related entities / related events / supply chain notes / market events 整理轻量关系快照。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_scenarios.go`
  - 构建 `bull / base / bear` 三情景。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_agents.go`
  - 构建多角色意见、共识与 veto。

### Backend Integration

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
  - 在四个真实读取入口上统一接入 `L2` explanation 增强。
- Optional Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_job_snapshot_repo.go`
  - 如需补齐历史快照读取中的 `simulations`/scenario 兼容解析，可在此做最小补强。

### Backend Tests

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_relationships_test.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_scenarios_test.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_agents_test.go`

### Client

- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
  - 增加 `L2` 字段读取与格式化 helper。
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.test.js`
  - 为 `L2` helper 加最小测试。
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
  - 接入三情景、角色意见与关系快照摘要。
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`
  - 接入历史版本情景回看。

### Admin

- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.js`
  - 扩展 `L2` 只读摘要与配置 helper。
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
  - 增加 `L2` 开关与阈值型配置。
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
  - 展示发布详情中的 `L2` 三情景与 veto 摘要。
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`
  - 展示 `L2` 的 advisory / veto 只读信息，不改审核动作。
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionRolesView.vue`
  - 增强角色意见读取与展示。
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionRolesView.vue`
  - 增强角色意见读取与展示。
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/system-configs-view.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/market-center-view.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/review-center-view.test.js`

## Task 1: Add L2 Contract Fields And Relationship Snapshot

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_relationships.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_relationships_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`

- [ ] **Step 1: 写 failing tests，先把 L2 contract 钉死**

Add tests similar to:

```go
func TestBuildFallbackVersionHistoryItemCarriesL2Fields(t *testing.T) {
	explanation := model.StrategyClientExplanation{
		RelationshipSnapshot: model.StrategyExplanationRelationshipSnapshot{
			AssetKey: "600519.SH",
			RelationshipCount: 3,
			Nodes: []model.StrategyExplanationRelationshipNode{{Type: "Theme", Label: "白酒"}},
		},
		ScenarioSnapshots: []model.StrategyExplanationScenarioSnapshot{{
			Scenario: "base",
			Thesis: "主逻辑延续",
			Trigger: "放量突破",
			InvalidationSignal: "跌破 5 日线",
			ActionSuggestion: "继续跟踪",
		}},
		AgentOpinions: []model.StrategyExplanationAgentOpinion{{
			Role: "FLOW",
			Stance: "SUPPORT",
			Confidence: 0.72,
			Summary: "资金承接稳定",
		}},
	}

	item := buildFallbackVersionHistoryItem("", "", "2026-03-29", 0, "2026-03-29T09:00:00Z", "", "", explanation)
	if len(item.ScenarioSnapshots) != 1 || len(item.AgentOpinions) != 1 {
		t.Fatalf("expected L2 fields to survive fallback item: %+v", item)
	}
	if item.RelationshipSnapshot.RelationshipCount != 3 {
		t.Fatalf("expected relationship snapshot to survive fallback item: %+v", item.RelationshipSnapshot)
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuildFallbackVersionHistoryItemCarriesL2Fields'`

Expected: FAIL，提示 `StrategyClientExplanation` 或 history item 缺少 L2 字段。

- [ ] **Step 2: 在 model 中新增 L2 子结构和非破坏性字段**

Implementation notes:
- 新增：
  - `StrategyExplanationRelationshipSnapshot`
  - `StrategyExplanationRelationshipNode`
  - `StrategyExplanationRelationshipEdge`
  - `StrategyExplanationScenarioSnapshot`
  - `StrategyExplanationScenarioMeta`
  - `StrategyExplanationAgentOpinion`
- 在 `StrategyClientExplanation` 和 `StrategyVersionHistoryItem` 中新增：
  - `RelationshipSnapshot`
  - `ScenarioSnapshots`
  - `ScenarioMeta`
  - `AgentOpinions`
- 字段全部为非破坏性扩展。

- [ ] **Step 3: 在新 helper 中实现轻量关系快照构造**

Implementation notes:
- 在 `strategy_explanation_relationships.go` 中新增：
  - `buildStockRelationshipSnapshot`
  - `buildFuturesRelationshipSnapshot`
  - `compactRelationshipNodes`
  - `compactRelationshipEdges`
- 股票最小节点类型：
  - `Company`
  - `Sector`
  - `Theme`
  - `Event`
  - `Policy`
- 期货最小节点类型：
  - `Contract`
  - `Commodity`
  - `SupplyChainNode`
  - `Warehouse`
  - `MacroEvent`
- 数据源只复用现有 explanation/context/report 字段，不新建表。

- [ ] **Step 4: 补齐 fallback / merge 透传逻辑**

Implementation notes:
- 修改 `buildStrategyVersionHistoryItem`
- 修改 `buildFallbackVersionHistoryItem`
- 修改 `mergeStrategyExplanation`
- 当前这一步只确保字段在 explanation/history 中不会丢失。

- [ ] **Step 5: 运行 targeted tests，确认 contract 与快照生成通过**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuildFallbackVersionHistoryItemCarriesL2Fields|TestBuild.*RelationshipSnapshot'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/model/strategy_client_explanation.go \
  backend/internal/growth/repo/strategy_explanation_relationships.go \
  backend/internal/growth/repo/strategy_explanation_relationships_test.go \
  backend/internal/growth/repo/strategy_client_explanation_test.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add l2 relationship snapshot contracts"
```

## Task 2: Add Stable Bull/Base/Bear Scenario Snapshots

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_scenarios.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_scenarios_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`

- [ ] **Step 1: 写 failing tests，固定三情景 contract**

Add tests similar to:

```go
func TestBuildStockScenarioSnapshotsReturnsBullBaseBear(t *testing.T) {
	ctx := strategyEngineAssetContext{
		record: model.StrategyEnginePublishRecord{TradeDate: "2026-03-29"},
		asset: map[string]any{
			"symbol": "600519.SH",
			"reason_summary": "资金回流叠加趋势延续",
			"risk_summary": "跌破 5 日线失效",
			"invalidations": []any{"跌破 5 日线"},
		},
	}

	scenarios, meta := buildStockScenarioSnapshots(ctx, nil, nil)
	if len(scenarios) != 3 {
		t.Fatalf("expected 3 stable scenarios, got %+v", scenarios)
	}
	if scenarios[0].Scenario == "" || scenarios[0].Trigger == "" || scenarios[0].ActionSuggestion == "" {
		t.Fatalf("expected complete scenario contract, got %+v", scenarios[0])
	}
	if meta.PrimaryScenario == "" {
		t.Fatalf("expected scenario meta, got %+v", meta)
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuild.*ScenarioSnapshots'`

Expected: FAIL，提示 helper 不存在或三情景不稳定。

- [ ] **Step 2: 实现股票 / 期货三情景 builder**

Implementation notes:
- 新增：
  - `buildStockScenarioSnapshots`
  - `buildFuturesScenarioSnapshots`
  - `buildScenarioMeta`
- 每个情景必须输出：
  - `scenario`
  - `thesis`
  - `trigger`
  - `confirmation_signal`
  - `invalidation_signal`
  - `expected_window`
  - `action_suggestion`
- 情景优先级固定三情景：
  - `bull`
  - `base`
  - `bear`
- 可以从现有 `simulations` 输入映射，但对外输出统一成稳定骨架。

- [ ] **Step 3: 将情景快照接入 explanation 生成链**

Implementation notes:
- 在 `buildStrategyExplanationFromContext` 中：
  - 先生成 relationship snapshot
  - 再生成 scenario snapshots
  - 再填充 scenario meta
- 没有现成 simulation 时，也要从 `L1` explanation + risk boundary + invalidations 生成降级版三情景。

- [ ] **Step 4: 保留旧 `simulations` 兼容，不让 client 现有读取断掉**

Implementation notes:
- 不删除 explanation 里的旧 `Simulations`
- `L2` 先新增稳定 `ScenarioSnapshots`，client 后续逐步优先读取新字段。

- [ ] **Step 5: 运行 targeted tests，确认三情景稳定**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuild.*ScenarioSnapshots|TestBuildStockStrategyExplanation.*|TestBuildFuturesStrategyExplanation.*'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/strategy_explanation_scenarios.go \
  backend/internal/growth/repo/strategy_explanation_scenarios_test.go \
  backend/internal/growth/repo/strategy_client_explanation.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add stable l2 scenario snapshots"
```

## Task 3: Add Agent Opinions, Consensus, And Veto

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_agents.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_agents_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`

- [ ] **Step 1: 写 failing tests，固定角色意见与 veto 行为**

Add tests similar to:

```go
func TestBuildStockAgentOpinionsReturnsConsensusAndVeto(t *testing.T) {
	ctx := strategyEngineAssetContext{
		asset: map[string]any{
			"reason_summary": "趋势延续但高位分歧",
			"risk_flags": []any{"高位分歧"},
			"invalidations": []any{"跌破 5 日线"},
		},
	}
	opinions, meta := buildStockAgentOpinions(ctx, nil, nil)
	if len(opinions) < 3 {
		t.Fatalf("expected multiple opinions, got %+v", opinions)
	}
	if meta.ConsensusAction == "" {
		t.Fatalf("expected consensus action, got %+v", meta)
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuild.*AgentOpinions'`

Expected: FAIL，提示角色 helper 不存在。

- [ ] **Step 2: 实现股票 / 期货角色 opinion builder**

Implementation notes:
- 股票角色最少包括：
  - `FLOW`
  - `THEME`
  - `EVENT`
  - `RISK`
- 期货角色最少包括：
  - `DIRECTION`
  - `SUPPLY`
  - `BASIS`
  - `MACRO`
  - `RISK`
- 每个角色输出：
  - `role`
  - `stance`
  - `confidence`
  - `summary`
  - `veto_flag`
- 新增汇总 helper：
  - `buildScenarioConsensusMeta`
- `RISK` 角色允许触发 veto，但只作用于 explanation/meta，不直接改排序或审批流。

- [ ] **Step 3: 将 agent opinions 与 scenario meta 串起来**

Implementation notes:
- `ScenarioMeta` 至少补：
  - `PrimaryScenario`
  - `ConsensusAction`
  - `Vetoed`
  - `VetoReason`
  - `ScenarioConfidenceSpread`
- explanation 中同时保留：
  - `ScenarioSnapshots`
  - `AgentOpinions`
  - `ScenarioMeta`

- [ ] **Step 4: 运行 targeted tests，确认意见层稳定**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuild.*AgentOpinions|TestBuild.*ScenarioSnapshots'`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/strategy_explanation_agents.go \
  backend/internal/growth/repo/strategy_explanation_agents_test.go \
  backend/internal/growth/repo/strategy_client_explanation.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add l2 agent opinions and veto meta"
```

## Task 4: Wire L2 Through The Four Real Read Entrypoints And History

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`
- Optional Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_engine_job_snapshot_repo.go`

- [ ] **Step 1: 写 failing tests，覆盖四个入口的 L2 字段可见性**

Add or extend tests so they assert:
- `GetStockRecommendationInsight(...).Explanation.ScenarioSnapshots`
- `GetStockRecommendationInsight(...).Explanation.AgentOpinions`
- `GetStockRecommendationVersionHistory(...)[0].ScenarioMeta`
- `GetFuturesStrategyInsight(...).Explanation.RelationshipSnapshot`
- `GetFuturesStrategyVersionHistory(...)[0].ScenarioSnapshots`

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestGetStockRecommendationInsight|TestGetStockRecommendationVersionHistory|TestGetFuturesStrategyInsight|TestGetFuturesStrategyVersionHistory'`

Expected: FAIL，提示四个入口尚未统一返回 L2 字段。

- [ ] **Step 2: 在 stock insight / history 接入 L2 helpers**

Implementation notes:
- `GetStockRecommendationInsight`
- `GetStockRecommendationVersionHistory`
- 在 `L1` explanation 生成完成后追加：
  - relationship snapshot
  - scenario snapshots
  - agent opinions
  - scenario meta

- [ ] **Step 3: 在 futures insight / history 接入同一套 L2 helpers**

Implementation notes:
- `GetFuturesStrategyInsight`
- `GetFuturesStrategyVersionHistory`
- 共享骨架，按 futures 模板生成关系与情景。

- [ ] **Step 4: 做一次 fallback / 空上下文回归**

Implementation notes:
- 即使没有完整 simulation/context：
  - explanation 仍要有 3 个降级场景
  - history item 不能丢 `ScenarioMeta`
  - `Vetoed` 默认 false，不能因为空数据误触发 veto

- [ ] **Step 5: 运行 backend repo 主回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/mysql_repo.go \
  backend/internal/growth/repo/strategy_client_explanation.go \
  backend/internal/growth/repo/strategy_client_explanation_test.go \
  backend/internal/growth/repo/strategy_engine_job_snapshot_repo.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: wire l2 scenarios into insight and history"
```

## Task 5: Update Client To Read And Render L2 Scenario Data

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`

- [ ] **Step 1: 写 helper 层 failing tests，先固定 L2 读取 contract**

Add tests similar to:

```js
import {
  buildStrategyScenarioSnapshotRows,
  buildStrategyAgentOpinionRows,
  buildStrategyScenarioMetaSummary
} from "./strategy-version.js";

test("buildStrategyScenarioMetaSummary returns veto-aware summary", () => {
  const result = buildStrategyScenarioMetaSummary({
    scenario_meta: {
      primary_scenario: "base",
      consensus_action: "继续观察",
      vetoed: true,
      veto_reason: "风险角色提示回撤超限"
    }
  });
  assert.match(result.summary, /veto/i);
});
```

Run: `cd /Users/gjhan21/cursor/sercherai/client && node --test src/lib/strategy-version.test.js`

Expected: FAIL，提示 helper 不存在。

- [ ] **Step 2: 在 `strategy-version.js` 增加 L2 helper**

Implementation notes:
- 新增：
  - `buildStrategyScenarioSnapshotRows`
  - `buildStrategyAgentOpinionRows`
  - `buildStrategyScenarioMetaSummary`
  - `buildStrategyRelationshipSnapshotSummary`
- 保持旧字段兼容；没有新字段时返回空数组/空摘要。

- [ ] **Step 3: 在 `StrategyView.vue` 接入 L2 卡组**

Implementation notes:
- 在现有 explanation 区新增 3 个区块：
  - 三情景推演
  - 角色意见与 veto
  - 关系快照摘要
- 不删除现有 `simulations` 区块，先过渡为优先读取 `ScenarioSnapshots`，旧字段作兼容回退。

- [ ] **Step 4: 在 `RecommendationArchiveView.vue` 增加历史情景回看**

Implementation notes:
- 每个历史项至少显示：
  - primary scenario
  - consensus action
  - veto reason
  - bull/base/bear 摘要

- [ ] **Step 5: 运行 client build**

Run: `cd /Users/gjhan21/cursor/sercherai/client && npm run build`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  client/src/lib/strategy-version.js \
  client/src/lib/strategy-version.test.js \
  client/src/views/StrategyView.vue \
  client/src/views/RecommendationArchiveView.vue
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: surface l2 scenarios in client views"
```

## Task 6: Extend Admin For L2 Visibility And Controls

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/lib/forecast-admin.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/StockSelectionRolesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/FuturesSelectionRolesView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/system-configs-view.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/market-center-view.test.js`
- Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/review-center-view.test.js`

- [ ] **Step 1: 写 failing tests，先固定 L2 admin 最小可见性**

Add tests or extend text-based tests so they assert:
- `SystemConfigsView` 出现 `forecast-l2` 配置区或在现有预测增强页出现 `L2` 字段
- `MarketCenterView` 出现三情景 / veto 摘要
- `ReviewCenterView` 出现 veto / consensus 只读提示

Run: `cd /Users/gjhan21/cursor/sercherai/admin && node --test src/lib/forecast-admin.test.js src/views/system-configs-view.test.js src/views/market-center-view.test.js src/views/review-center-view.test.js`

Expected: FAIL，提示 L2 admin 内容未接入。

- [ ] **Step 2: 扩展 `forecast-admin.js` 的 L2 helper**

Implementation notes:
- 增加 `L2` 配置解析与摘要 helper，例如：
  - `growth.forecast_l2.enabled`
  - `growth.forecast_l2.scenario_enabled`
  - `growth.forecast_l2.veto_threshold`
- 扩展发布详情摘要：
  - 三情景覆盖数
  - veto 样本数
  - primary scenario 分布

- [ ] **Step 3: 在现有 admin 页面嵌入 L2 信息，不新开中心**

Implementation notes:
- `SystemConfigsView.vue`
  - 增加 `L2` 开关/阈值型配置
- `MarketCenterView.vue`
  - 发布详情展示三情景、consensus、veto
- `ReviewCenterView.vue`
  - 展示 advisory-only / veto 只读提示，不新增审核动作
- `StockSelectionRolesView.vue` / `FuturesSelectionRolesView.vue`
  - 展示角色意见与 scenario template 的对应关系

- [ ] **Step 4: 运行 admin tests 与 build**

Run:
- `cd /Users/gjhan21/cursor/sercherai/admin && node --test src/lib/forecast-admin.test.js src/views/system-configs-view.test.js src/views/market-center-view.test.js src/views/review-center-view.test.js`
- `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  admin/src/lib/forecast-admin.js \
  admin/src/lib/forecast-admin.test.js \
  admin/src/views/SystemConfigsView.vue \
  admin/src/views/MarketCenterView.vue \
  admin/src/views/ReviewCenterView.vue \
  admin/src/views/stock-selection/StockSelectionRolesView.vue \
  admin/src/views/futures-selection/FuturesSelectionRolesView.vue \
  admin/src/views/system-configs-view.test.js \
  admin/src/views/market-center-view.test.js \
  admin/src/views/review-center-view.test.js
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add l2 forecast visibility in admin"
```

## Verification Checklist

- [ ] explanation 与 history 都能返回 `RelationshipSnapshot / ScenarioSnapshots / ScenarioMeta / AgentOpinions`。
- [ ] 三情景固定为 `bull / base / bear`，字段完整。
- [ ] 角色意见可以形成 `consensus_action / veto_reason`，但不改排序/审批主链。
- [ ] fallback / 空上下文路径不会丢失 `L2` 字段，也不会误触发 veto。
- [ ] client 不读取新字段时仍兼容；读取新字段时能展示三情景与角色意见。
- [ ] admin 只做复用式承接，没有新开独立 `L2` 后台中心。

## Full Validation Commands

- Backend targeted:

```bash
cd /Users/gjhan21/cursor/sercherai/backend
go test ./internal/growth/repo -run 'TestBuild.*RelationshipSnapshot|TestBuild.*ScenarioSnapshots|TestBuild.*AgentOpinions|TestGetStockRecommendationInsight|TestGetStockRecommendationVersionHistory|TestGetFuturesStrategyInsight|TestGetFuturesStrategyVersionHistory'
```

- Backend full:

```bash
cd /Users/gjhan21/cursor/sercherai/backend
go test ./internal/growth/repo
```

- Client:

```bash
cd /Users/gjhan21/cursor/sercherai/client
node --test src/lib/strategy-version.test.js
npm run build
```

- Admin:

```bash
cd /Users/gjhan21/cursor/sercherai/admin
node --test src/lib/forecast-admin.test.js src/views/system-configs-view.test.js src/views/market-center-view.test.js src/views/review-center-view.test.js
npm run build
```

## Recommended Commit Sequence

1. `feat: add l2 relationship snapshot contracts`
2. `feat: add stable l2 scenario snapshots`
3. `feat: add l2 agent opinions and veto meta`
4. `feat: wire l2 scenarios into insight and history`
5. `feat: surface l2 scenarios in client views`
6. `feat: add l2 forecast visibility in admin`

## Handoff Notes For Other Threads

- 开工前必须先读：
  - `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md`
  - `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l2-design.md`
  - `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md`
  - 本计划文件
- 任何线程如果想把 `L2` 扩成新 analysis 页、独立后台中心、独立异步模拟系统，都必须先停下。
- 本计划默认顺序固定：先 backend contract，再三情景，再角色意见，再四入口集成，再 client，再 admin。
