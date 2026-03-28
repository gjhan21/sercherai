# Stock Futures Forecast L1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不改现有推荐主链、不新增新业务表和独立异步系统的前提下，把股票推荐与期货策略的 `insight / version-history` 四个入口升级为带研究编排、记忆反馈消费、历史理由分层和置信度校准的 `L1` 预测增强版本。

**Architecture:** 继续复用现有 `strategy-engine` 报告快照、`StrategyClientExplanation`、`version-history`、评估回补与现有 client 展示链，不新建第二套预测系统。主线做法是先在后端 explanation 生成链中补“研究编排 + 历史失效管理 + advisory 型置信度校准”纯函数层，再把这些非破坏性字段透传到现有 `StrategyView` / `RecommendationArchiveView` 等页面，最后把 admin 需求收口为可选的轻量嵌入，不作为 `L1` 主线前提。

**Tech Stack:** Go, MySQL repo layer, Gin API, existing strategy-engine snapshots, Vue 3 client, existing `client/src/lib/strategy-version.js`

---

## Scope And Guardrails

- 本计划只覆盖 `L1`，不提前实现 `L2 / L3`。
- 本计划只覆盖 4 个入口：
  - 股票推荐 `insight`
  - 股票推荐 `version-history`
  - 期货策略 `insight`
  - 期货策略 `version-history`
- `memory_feedback` 与 `confidence calibration` 必须真实参与 explanation 生成，不允许只是“多返回几个字段”。
- `L1` 中上述能力只允许影响：
  - explanation 内容
  - risk boundary
  - confidence summary / warning flags
  - reviewer advisory priority
- `L1` 中明确不允许：
  - 直接改 candidate ranking
  - 直接改 portfolio weight
  - 直接阻断 publish review
  - 直接替代现有主打分链
- 默认不新增新的持久化业务表。
- 默认不新增独立异步研究刷新系统。
- admin 不作为本计划主线阻塞项；只允许做后置、嵌入式、轻量承接。

## Current Baseline

- explanation 结构定义在 `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`。
- 股票与期货 explanation 生成主入口在：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- 现有四个读取入口在：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- 股票评估回补与 `evaluation_summary` 补强在：
  - `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
- 现有 client 读取入口在：
  - `/Users/gjhan21/cursor/sercherai/client/src/api/market.js`
  - `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
  - `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
  - `/Users/gjhan21/cursor/sercherai/client/src/views/HomeView.vue`
  - `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`
- admin 现有可复用页面已存在于：
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/stock-selection/*`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/futures-selection/*`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
  - `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`

## File Map

### Backend Models And Contracts

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
  - 为 explanation 与 version-history 增加 `research_outline / active_thesis_cards / historical_thesis_cards / watch_signals / confidence_calibration` 非破坏性字段和配套子结构。

### Backend Explanation Builders

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
  - 保留现有主入口和现有 SQL 读取逻辑，接入新的 research/memory/calibration 纯函数并把字段合并进 `StrategyClientExplanation` 与 `StrategyVersionHistoryItem`。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_research.go`
  - 负责股票 / 期货的研究编排、理由分层、观察信号生成。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_memory.go`
  - 负责把现有 `memory_feedback`、`evaluation_summary`、version diff 等事实源整理为 explanation 可消费的 advisory 输入。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_confidence.go`
  - 负责轻量置信度与风险边界校准，并把结果写回 explanation。

### Backend Repo Integration

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
  - 在四个真实入口上统一接入 `L1` explanation 增强，不新增接口。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
  - 只做最小补强，统一 `evaluation_summary` 的读取和空态处理，方便股票 explanation 消费。

### Backend Tests

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`
  - 增加 stock/futures explanation 与 history 的端到端断言。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_research_test.go`
  - 纯函数测试：研究编排、活跃理由、历史理由、watch signals。
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_confidence_test.go`
  - 纯函数测试：confidence calibration 与 advisory 调整。
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill_test.go`
  - 增加 evaluation summary 归一化与空态兜底断言。

### Client

- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
  - 增加前端对 `research_outline / active_thesis_cards / historical_thesis_cards / watch_signals / confidence_calibration` 的读取与格式化 helpers。
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
  - 在现有策略页 explanation 区增加研究编排、当前有效理由、历史弱化理由、watch signals、置信度校准卡。
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`
  - 在现有复盘页补 history item 的理由分层和后验记忆展示。
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/HomeView.vue`
  - 只做轻量承接，把首页主推荐卡的“为什么选它 / 风险边界 / 下一步观察”切到新字段优先，不重做页面结构。

### Optional Admin Follow-Up (Not Blocking L1 Mainline)

- Optional Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
- Optional Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
- Optional Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`
- Optional Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`

这些 admin 文件不进入主线任务；只有在后端和 client 主线稳定后才允许进入。

## Task 1: Extend Explanation Contracts For L1 Fields

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_client_explanation.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`

- [ ] **Step 1: 写 failing tests，先把 L1 字段 contract 钉死**

Add tests similar to:

```go
func TestBuildFallbackVersionHistoryItemCarriesL1Fields(t *testing.T) {
	explanation := model.StrategyClientExplanation{
		ResearchOutline: []model.StrategyResearchOutlineStep{
			{Slot: "TREND", Title: "趋势与结构", Summary: "价格结构仍强"},
		},
		ActiveThesisCards: []model.StrategyExplanationThesisCard{
			{Key: "trend", Title: "趋势延续", Summary: "主升趋势未破坏"},
		},
		HistoricalThesisCards: []model.StrategyExplanationThesisCard{
			{Key: "event", Title: "事件催化弱化", Summary: "旧催化已进入兑现期"},
		},
		WatchSignals: []model.StrategyExplanationWatchSignal{
			{Title: "跌破止损", SignalType: "INVALIDATION", Trigger: "跌破 5 日线"},
		},
		ConfidenceCalibration: model.StrategyExplanationConfidenceCalibration{
			BaseConfidence: 0.72,
			AdjustedConfidence: 0.64,
			AdvisoryOnly: true,
		},
	}

	item := buildFallbackVersionHistoryItem("", "", "2026-03-28", 0, "2026-03-28T09:00:00Z", "", "", explanation)
	if len(item.ActiveThesisCards) != 1 || len(item.HistoricalThesisCards) != 1 {
		t.Fatalf("expected L1 thesis fields to survive fallback item: %+v", item)
	}
	if len(item.WatchSignals) != 1 || item.ConfidenceCalibration.AdjustedConfidence <= 0 {
		t.Fatalf("expected L1 watch/confidence fields to survive fallback item: %+v", item)
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuildFallbackVersionHistoryItemCarriesL1Fields'`

Expected: FAIL，提示 `StrategyClientExplanation` 或 `StrategyVersionHistoryItem` 缺少新字段。

- [ ] **Step 2: 在 model 中新增 L1 子结构和非破坏性字段**

Implementation notes:
- 在 `strategy_client_explanation.go` 中新增：
  - `StrategyResearchOutlineStep`
  - `StrategyExplanationThesisCard`
  - `StrategyExplanationWatchSignal`
  - `StrategyExplanationConfidenceDriver`
  - `StrategyExplanationConfidenceCalibration`
- 在 `StrategyClientExplanation` 中新增：
  - `ResearchOutline []StrategyResearchOutlineStep`
  - `ActiveThesisCards []StrategyExplanationThesisCard`
  - `HistoricalThesisCards []StrategyExplanationThesisCard`
  - `WatchSignals []StrategyExplanationWatchSignal`
  - `ConfidenceCalibration StrategyExplanationConfidenceCalibration`
- 在 `StrategyVersionHistoryItem` 中同步新增同名字段。

Recommended type sketch:

```go
type StrategyResearchOutlineStep struct {
	Slot         string `json:"slot"`
	Title        string `json:"title"`
	Summary      string `json:"summary"`
	Status       string `json:"status"`
	EvidenceHint string `json:"evidence_hint"`
}

type StrategyExplanationThesisCard struct {
	Key            string `json:"key"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	Status         string `json:"status"`
	EvidenceSource string `json:"evidence_source"`
	Note           string `json:"note"`
}

type StrategyExplanationWatchSignal struct {
	Title      string `json:"title"`
	SignalType string `json:"signal_type"`
	Trigger    string `json:"trigger"`
	Action     string `json:"action"`
	Priority   string `json:"priority"`
}

type StrategyExplanationConfidenceCalibration struct {
	BaseConfidence     float64                              `json:"base_confidence"`
	AdjustedConfidence float64                              `json:"adjusted_confidence"`
	Drivers            []StrategyExplanationConfidenceDriver `json:"drivers"`
	AdvisoryOnly       bool                                 `json:"advisory_only"`
}
```

- [ ] **Step 3: 把 fallback/history 构造函数先补齐新字段透传**

Implementation notes:
- 修改 `buildStrategyVersionHistoryItem`
- 修改 `buildFallbackVersionHistoryItem`
- 修改 `mergeStrategyExplanation`
- 这一小步先只做“字段可透传”，不做生成逻辑。

- [ ] **Step 4: 运行 targeted tests，确认 contract 成型**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuildFallbackVersionHistoryItemCarriesL1Fields'`

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/model/strategy_client_explanation.go \
  backend/internal/growth/repo/strategy_client_explanation_test.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add l1 explanation contracts"
```

## Task 2: Add Research Outline, Active Thesis, Historical Thesis, And Watch Signals

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_research.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_research_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`

- [ ] **Step 1: 写 failing tests，先固定股票与期货研究编排的最小行为**

Add tests similar to:

```go
func TestBuildStockResearchBlocksCreatesActiveHistoricalAndWatchLayers(t *testing.T) {
	report := map[string]any{
		"market_regime": "ROTATION",
		"memory_feedback": map[string]any{
			"summary": "高波动题材需要缩短验证窗口",
			"failure_signals": []any{"题材切换过快"},
		},
	}
	ctx := strategyEngineAssetContext{
		record: model.StrategyEnginePublishRecord{TradeDate: "2026-03-28"},
		asset: map[string]any{
			"symbol": "600519.SH",
			"reason_summary": "资金回流叠加趋势延续",
			"risk_summary": "跌破 5 日线失效",
			"risk_flags": []any{"高位分歧"},
			"invalidations": []any{"跌破 5 日线"},
			"theme_tags": []any{"消费龙头"},
		},
	}

	outline, active, historical, watch := buildStockResearchBlocks(ctx, report, nil, nil)
	if len(outline) == 0 || len(active) == 0 || len(watch) == 0 {
		t.Fatalf("expected stock research blocks, got outline=%+v active=%+v watch=%+v", outline, active, watch)
	}
	if len(historical) == 0 {
		t.Fatalf("expected memory/version inputs to create historical thesis cards")
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuild.*ResearchBlocks'`

Expected: FAIL，提示 research helper 不存在或没有生成分层结果。

- [ ] **Step 2: 在新文件里实现股票与期货分域 research builders**

Implementation notes:
- 在 `strategy_explanation_research.go` 中新增：
  - `buildStockResearchBlocks`
  - `buildFuturesResearchBlocks`
  - `buildResearchOutlineSteps`
  - `buildActiveThesisCards`
  - `buildHistoricalThesisCards`
  - `buildWatchSignals`
- 股票固定 5 个问题槽位：
  - 趋势与价量结构
  - 资金承接与流向
  - 估值与基本面约束
  - 行业 / 主题 / 事件催化
  - 风险边界与失效条件
- 期货固定 5 个问题槽位：
  - 方向与关键价位
  - 供需 / 库存 / 产业链
  - 基差 / 期限结构 / 价差
  - 宏观 / 政策 / 事件扰动
  - 风险边界与失效条件
- 当前有效理由优先来自：
  - `reason_summary`
  - `evidence_cards`
  - `theme_tags / sector_tags`
  - `related_entities / related_events`
- 历史理由优先来自：
  - 旧版本 `reason_summary`
  - version diff 中被替换的理由
  - `memory_feedback.failure_signals`
- watch signals 优先来自：
  - `invalidations`
  - `risk_flags`
  - `memory_feedback.suggestions`
  - 校准器要求增加的 sentinel

Recommended helper signature:

```go
func buildStockResearchBlocks(
	ctx strategyEngineAssetContext,
	report map[string]any,
	previous *strategyEngineAssetContext,
	evaluation map[string]any,
) (
	outline []model.StrategyResearchOutlineStep,
	active []model.StrategyExplanationThesisCard,
	historical []model.StrategyExplanationThesisCard,
	watch []model.StrategyExplanationWatchSignal,
)
```

- [ ] **Step 3: 接入 `buildStrategyExplanationFromContext`，先让 explanation 自身能生成这些字段**

Implementation notes:
- 在 `buildStrategyExplanationFromContext` 中：
  - 识别股票 / 期货资产域
  - 根据当前 `ctx.asset` 与 `report` 生成 `ResearchOutline`
  - 填充 `ActiveThesisCards`
  - 在没有历史上下文时，允许 `HistoricalThesisCards` 退化为 `memory_feedback` / invalidation 衍生的“已弱化理由”
  - 填充基础 `WatchSignals`

- [ ] **Step 4: 用 history 上下文补强 historical thesis**

Implementation notes:
- 新增一个后处理 helper，例如：
  - `applyStrategyL1HistoryFromContexts(explanation *model.StrategyClientExplanation, contexts []strategyEngineAssetContext, assetKey string)`
- `GetStockRecommendationInsight` 与 `GetFuturesStrategyInsight` 在拿到 `contexts` 之后调用它。
- `applyStrategyVersionDiffToHistoryItems` 旁边新增 history item 对应的 L1 helper，例如：
  - `applyStrategyL1HistoryToHistoryItems(items []model.StrategyVersionHistoryItem, contexts []strategyEngineAssetContext, assetKey string)`

- [ ] **Step 5: 运行 tests，确认 research 分层稳定**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuild.*ResearchBlocks|TestBuildStockStrategyExplanation.*|TestBuildFallbackVersionHistoryItemCarriesL1Fields'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/strategy_explanation_research.go \
  backend/internal/growth/repo/strategy_explanation_research_test.go \
  backend/internal/growth/repo/strategy_client_explanation.go \
  backend/internal/growth/repo/strategy_client_explanation_test.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add l1 research outline and thesis layers"
```

## Task 3: Make Memory Feedback And Evaluation Summary Actively Influence Explanation

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_memory.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/stock_selection_run_evaluation_backfill_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`

- [ ] **Step 1: 写 failing tests，明确“被消费”而不是“仅返回”**

Add tests similar to:

```go
func TestApplyL1AdvisoryMemoryAdjustmentsAppendsRiskBoundaryAndWatchSignal(t *testing.T) {
	explanation := model.StrategyClientExplanation{
		ConfidenceReason: "当前趋势与资金共振",
		RiskBoundary: "跌破 5 日线失效",
		MemoryFeedback: model.StrategyExplanationMemoryFeedback{
			Summary: "过去同类题材高波动、验证慢",
			Suggestions: []string{"缩短验证窗口"},
			FailureSignals: []string{"高位放量不涨"},
		},
	}
	evaluation := map[string]any{
		"status": "COMPLETED",
		"5": map[string]any{"return_pct": 0.03, "max_drawdown_pct": -0.09},
	}

	applyL1AdvisoryMemoryAdjustments(&explanation, evaluation)
	if !strings.Contains(explanation.RiskBoundary, "缩短验证窗口") {
		t.Fatalf("expected memory feedback to update risk boundary: %+v", explanation)
	}
	if len(explanation.WatchSignals) == 0 {
		t.Fatalf("expected memory feedback to create watch signals: %+v", explanation)
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestApplyL1AdvisoryMemoryAdjustments|TestBuildStockStrategyExplanation.*'`

Expected: FAIL，提示 advisory memory helper 不存在，或 `RiskBoundary / WatchSignals` 未变化。

- [ ] **Step 2: 在新 helper 文件里实现 memory/evaluation 归一化**

Implementation notes:
- 在 `strategy_explanation_memory.go` 中新增：
  - `normalizeExplanationEvaluationSummary`
  - `buildMemoryFailureSignals`
  - `buildMemorySuggestionSignals`
  - `applyL1AdvisoryMemoryAdjustments`
- 股票优先消费：
  - `report["memory_feedback"]`
  - `report["evaluation_summary"]`
  - `loadStockSelectionEvaluationSummaryByContext`
- 期货优先消费：
  - futures report 自带的 `memory_feedback`
  - futures report 自带的 `evaluation_summary`
- 不新增存储，只在读取期同步消费。

- [ ] **Step 3: 统一股票 evaluation summary 的空态与 map 结构**

Implementation notes:
- 在 `stock_selection_run_evaluation_backfill.go` 中新增一个很小的 helper，例如：

```go
func normalizeStockSelectionEvaluationSummary(summary map[string]any) map[string]any {
	if len(summary) == 0 {
		return map[string]any{"status": "PENDING"}
	}
	if _, ok := summary["status"]; !ok {
		summary["status"] = "COMPLETED"
	}
	return summary
}
```

- `loadStockSelectionEvaluationSummaryByContext`
- `enrichStockSelectionEvaluationMetaFromSummary`
- `enrichStockSelectionVersionHistoryEvaluationMeta`

统一走这个 helper，避免后面 confidence 校准因空 map 失真。

- [ ] **Step 4: 在 explanation 生成链中显式调用 advisory memory adjustments**

Implementation notes:
- `buildStrategyExplanationFromContext` 生成基础 explanation 后，立即调用：
  - `applyL1AdvisoryMemoryAdjustments`
- 要求至少带来 3 个可见效果中的 2 个：
  - 更新 `RiskBoundary`
  - 增加 `WatchSignals`
  - 补充 `RiskFlags` 或 `ConfidenceReason`

- [ ] **Step 5: 运行 tests，确认 memory feedback 已真正生效**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestApplyL1AdvisoryMemoryAdjustments|TestBuildStockStrategyExplanation.*|TestBuildFuturesStrategyExplanation.*|TestNormalizeStockSelectionEvaluationSummary'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/strategy_explanation_memory.go \
  backend/internal/growth/repo/strategy_client_explanation.go \
  backend/internal/growth/repo/strategy_client_explanation_test.go \
  backend/internal/growth/repo/stock_selection_run_evaluation_backfill.go \
  backend/internal/growth/repo/stock_selection_run_evaluation_backfill_test.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: consume l1 memory feedback in explanations"
```

## Task 4: Add Advisory Confidence Calibration And Risk Boundary Calibration

**Files:**
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_confidence.go`
- Create: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_explanation_confidence_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`

- [ ] **Step 1: 写 failing tests，钉死 `confidence_calibration` 的 advisory contract**

Add tests similar to:

```go
func TestBuildExplanationConfidenceCalibrationIsAdvisoryOnly(t *testing.T) {
	explanation := model.StrategyClientExplanation{
		ConfidenceReason: "趋势延续",
		MemoryFeedback: model.StrategyExplanationMemoryFeedback{
			Summary: "高波动模板近几次回撤偏大",
		},
		EvaluationMeta: map[string]any{
			"status": "COMPLETED",
			"5": map[string]any{
				"return_pct": 0.02,
				"max_drawdown_pct": -0.11,
			},
		},
	}

	calibration := buildExplanationConfidenceCalibration(explanation)
	if !calibration.AdvisoryOnly {
		t.Fatalf("expected advisory_only=true")
	}
	if calibration.AdjustedConfidence >= calibration.BaseConfidence {
		t.Fatalf("expected drawdown-heavy case to reduce confidence: %+v", calibration)
	}
	if len(calibration.Drivers) == 0 {
		t.Fatalf("expected confidence drivers")
	}
}
```

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuildExplanationConfidenceCalibration'`

Expected: FAIL，提示 calibration helper 不存在或没有生成 advisory 结果。

- [ ] **Step 2: 实现轻量校准器**

Implementation notes:
- 在 `strategy_explanation_confidence.go` 中新增：
  - `buildExplanationConfidenceCalibration`
  - `deriveBaseConfidenceScore`
  - `buildConfidenceDrivers`
  - `applyConfidenceCalibrationToExplanation`
- 固定输入：
  - `evaluation_meta`
  - `memory_feedback`
  - `risk_flags / invalidations`
  - `evidence_cards`
  - `related_events`
- 固定输出：
  - `BaseConfidence`
  - `AdjustedConfidence`
  - `Drivers`
  - `AdvisoryOnly`
- 调整范围只允许：
  - 更新 `ConfidenceReason`
  - 增加 `RiskFlags`
  - 强化 `RiskBoundary`
  - 补充 `WatchSignals`

Recommended skeleton:

```go
func buildExplanationConfidenceCalibration(
	explanation model.StrategyClientExplanation,
) model.StrategyExplanationConfidenceCalibration {
	base := deriveBaseConfidenceScore(explanation)
	drivers, adjusted := buildConfidenceDrivers(explanation, base)
	return model.StrategyExplanationConfidenceCalibration{
		BaseConfidence: base,
		AdjustedConfidence: adjusted,
		Drivers: drivers,
		AdvisoryOnly: true,
	}
}
```

- [ ] **Step 3: 在 explanation 生成链和 history item 链统一接入 calibration**

Implementation notes:
- `buildStrategyExplanationFromContext`
- `buildStrategyVersionHistoryItem`
- `buildFallbackVersionHistoryItem`
- `mergeStrategyExplanation`

都必须能保留 `ConfidenceCalibration`。

- [ ] **Step 4: 在四个 repo 读取入口上确认不会改排序/权重主链**

Implementation notes:
- `GetStockRecommendationInsight`
- `GetStockRecommendationVersionHistory`
- `GetFuturesStrategyInsight`
- `GetFuturesStrategyVersionHistory`

只允许 explanation 内容变化，严禁改 `item.Score`、策略排序、持仓权重或 review status。

- [ ] **Step 5: 运行 tests，确认 advisory calibration 生效且不越权**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestBuildExplanationConfidenceCalibration|TestBuildStockStrategyExplanation.*|TestBuildFuturesStrategyExplanation.*|TestGet.*VersionHistory.*'`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/strategy_explanation_confidence.go \
  backend/internal/growth/repo/strategy_explanation_confidence_test.go \
  backend/internal/growth/repo/strategy_client_explanation.go \
  backend/internal/growth/repo/strategy_client_explanation_test.go \
  backend/internal/growth/repo/mysql_repo.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: add advisory confidence calibration for l1"
```

## Task 5: Wire L1 End-To-End Through The Four Real Read Entrypoints

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/mysql_repo.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_client_explanation_test.go`

- [ ] **Step 1: 写 failing tests，覆盖四个真实入口的 L1 字段可见性**

Add or extend tests so they assert:
- `GetStockRecommendationInsight(...).Explanation.ActiveThesisCards`
- `GetStockRecommendationInsight(...).Explanation.ConfidenceCalibration`
- `GetStockRecommendationVersionHistory(...)[0].HistoricalThesisCards`
- `GetFuturesStrategyInsight(...).Explanation.WatchSignals`
- `GetFuturesStrategyVersionHistory(...)[0].ResearchOutline`

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo -run 'TestGetStockRecommendationInsight|TestGetStockRecommendationVersionHistory|TestGetFuturesStrategyInsight|TestGetFuturesStrategyVersionHistory'`

Expected: FAIL，提示四个入口尚未统一带出 L1 字段。

- [ ] **Step 2: 在 stock insight / history 入口挂接 L1 helpers**

Implementation notes:
- `GetStockRecommendationInsight`
  - `buildStockStrategyExplanation` 仍是基础入口
  - 取到 `contexts` 后同时调用：
    - `applyStrategyVersionDiffToExplanation`
    - `applyStrategyL1HistoryFromContexts`
- `GetStockRecommendationVersionHistory`
  - history items 构造完成后调用：
    - `applyStrategyVersionDiffToHistoryItems`
    - `applyStrategyL1HistoryToHistoryItems`

- [ ] **Step 3: 在 futures insight / history 入口挂接同一套 L1 helpers**

Implementation notes:
- `GetFuturesStrategyInsight`
- `GetFuturesStrategyVersionHistory`

逻辑与 stock 保持一致，但 research slot 用 futures 模板。

- [ ] **Step 4: 做一次空上下文和 fallback 路径回归**

Implementation notes:
- 没有历史上下文时：
  - explanation 仍应有 `ResearchOutline`
  - `ActiveThesisCards` 仍可从当前 reason/evidence 生成
  - `ConfidenceCalibration.AdvisoryOnly` 仍必须为 `true`
- `buildFallbackVersionHistoryItem` 路径不能因为缺 contexts 而丢字段。

- [ ] **Step 5: 运行 repo 主回归**

Run: `cd /Users/gjhan21/cursor/sercherai/backend && go test ./internal/growth/repo`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  backend/internal/growth/repo/mysql_repo.go \
  backend/internal/growth/repo/strategy_client_explanation.go \
  backend/internal/growth/repo/strategy_client_explanation_test.go
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: wire l1 forecast enrichment into insight and history"
```

## Task 6: Update Client To Read And Render L1 Explanation Fields

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.js`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/StrategyView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/RecommendationArchiveView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/client/src/views/HomeView.vue`

- [ ] **Step 1: 写前端 helper 层的 failing tests；如果当前 client 无单测，则先补 helper 级测试文件**

Preferred new test file:
- Create: `/Users/gjhan21/cursor/sercherai/client/src/lib/strategy-version.test.js`

Recommended test sketch:

```js
import {
  buildStrategyResearchOutlineRows,
  buildStrategyThesisCardRows,
  buildStrategyWatchSignalRows,
  buildStrategyConfidenceCalibrationSummary
} from "./strategy-version";

test("buildStrategyConfidenceCalibrationSummary returns advisory label", () => {
  const result = buildStrategyConfidenceCalibrationSummary({
    confidence_calibration: {
      base_confidence: 0.72,
      adjusted_confidence: 0.64,
      advisory_only: true,
      drivers: [{ label: "回撤偏大", effect: "down" }]
    }
  });
  expect(result.summary).toContain("建议");
  expect(result.deltaLabel).toContain("-8");
});
```

Run: `cd /Users/gjhan21/cursor/sercherai/client && npm test -- strategy-version.test.js`

Expected: 如果仓库没有测试脚本，则记录为“当前 client 无通用 test 脚本”，改为先让 helper 代码编译并在 build 阶段验证；不要为了这一轮引入整套新测试框架。

- [ ] **Step 2: 在 `strategy-version.js` 增加新字段格式化 helpers**

Implementation notes:
- 新增：
  - `buildStrategyResearchOutlineRows`
  - `buildStrategyThesisCardRows`
  - `buildStrategyWatchSignalRows`
  - `buildStrategyConfidenceCalibrationSummary`
- 保持兼容：
  - 旧 explanation 没有这些字段时返回空数组或空 summary
  - 不破坏现有 `buildStrategyInsightSections / buildStrategyRiskBoundaryText`

- [ ] **Step 3: 在 `StrategyView.vue` explanation 区接入新模块**

Implementation notes:
- 在现有 explanation 区新增 4 个卡组，顺序固定：
  - 研究拆解
  - 当前有效理由
  - 历史弱化理由
  - 观察信号与置信度校准
- 继续保留现有：
  - why now
  - 风险边界
  - 版本差异
  - 历史版本
- 不要新开页面，不要把当前策略页改成分析聊天页。

- [ ] **Step 4: 在 `RecommendationArchiveView.vue` 和 `HomeView.vue` 做轻量承接**

Implementation notes:
- `RecommendationArchiveView.vue`
  - 历史样本卡优先显示：
    - 历史理由
    - 后续失效 / 止盈止损原因
    - `HistoricalThesisCards`
    - `ConfidenceCalibration` 摘要
- `HomeView.vue`
  - 主推荐卡只补 2 个点：
    - 新 explanation 的首条 active thesis
    - 首条 watch signal / calibration 提示
- 首页不要扩成长页。

- [ ] **Step 5: 运行 client build**

Run: `cd /Users/gjhan21/cursor/sercherai/client && npm run build`

Expected: PASS

- [ ] **Step 6: Commit**

```bash
git -C /Users/gjhan21/cursor/sercherai add \
  client/src/lib/strategy-version.js \
  client/src/views/StrategyView.vue \
  client/src/views/RecommendationArchiveView.vue \
  client/src/views/HomeView.vue
git -C /Users/gjhan21/cursor/sercherai commit -m "feat: surface l1 forecast insights in client views"
```

## Task 7: Optional Admin Embedding After Mainline Stabilizes

**Files:**
- Optional Modify: `/Users/gjhan21/cursor/sercherai/admin/src/api/admin.js`
- Optional Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/SystemConfigsView.vue`
- Optional Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/MarketCenterView.vue`
- Optional Modify: `/Users/gjhan21/cursor/sercherai/admin/src/views/ReviewCenterView.vue`

- [ ] **Step 1: 只有在 Task 1-6 已稳定后，才确认是否真的需要 admin 嵌入**

Decision gate:
- 如果 owner 只需要前台 explanation 增强，本任务直接跳过。
- 如果 owner 需要运营可见性，只允许做嵌入式只读/轻配置，不开新后台中心。

- [ ] **Step 2: 如需做 admin，只做 3 类轻量能力**

Allowed scope only:
- `SystemConfigsView`
  - 全局 kill switch
  - `memory_feedback` 最小样本阈值
  - explanation 增强展示总开关
- `MarketCenterView`
  - 最近增强 explanation 覆盖数
  - 高 advisory priority 样本摘要
- `ReviewCenterView`
  - 仅展示 advisory priority，不改审核主流程

- [ ] **Step 3: 禁止事项复核**

Must not do:
- 不新开“预测增强中心”
- 不新增平行审核流
- 不让 admin 配置直接改排序/权重

- [ ] **Step 4: 如果真的进入 admin，再单独写 mini plan 并执行**

Run before touching admin:
- `cd /Users/gjhan21/cursor/sercherai/admin && npm run build`

Expected: PASS 基线确认后，再开始 admin 嵌入开发。

## Verification Checklist

- [ ] `memory_feedback` 在 explanation 中被真实消费，而不是只作为返回字段存在。
- [ ] `confidence_calibration.advisory_only` 在 stock/futures explanation 与 history 中都恒为 `true`。
- [ ] stock/futures 四个入口都能返回新字段。
- [ ] 排序、持仓权重、审核流没有被 `L1` 误改。
- [ ] 没有新增新业务表或独立异步研究系统。
- [ ] client 在旧字段缺失时仍能正常渲染，不出现白屏。

## Full Validation Commands

- Backend targeted:

```bash
cd /Users/gjhan21/cursor/sercherai/backend
go test ./internal/growth/repo -run 'TestBuild.*ResearchBlocks|TestApplyL1AdvisoryMemoryAdjustments|TestBuildExplanationConfidenceCalibration|TestGetStockRecommendationInsight|TestGetStockRecommendationVersionHistory|TestGetFuturesStrategyInsight|TestGetFuturesStrategyVersionHistory'
```

- Backend full:

```bash
cd /Users/gjhan21/cursor/sercherai/backend
go test ./internal/growth/repo
```

- Client build:

```bash
cd /Users/gjhan21/cursor/sercherai/client
npm run build
```

- Optional admin baseline:

```bash
cd /Users/gjhan21/cursor/sercherai/admin
npm run build
```

## Recommended Commit Sequence

1. `feat: add l1 explanation contracts`
2. `feat: add l1 research outline and thesis layers`
3. `feat: consume l1 memory feedback in explanations`
4. `feat: add advisory confidence calibration for l1`
5. `feat: wire l1 forecast enrichment into insight and history`
6. `feat: surface l1 forecast insights in client views`
7. `chore: add optional admin forecast controls` only if Task 7 truly executed

## Handoff Notes For Other Threads

- 开工前必须先读：
  - `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-roadmap.md`
  - `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-l1-design.md`
  - `/Users/gjhan21/cursor/sercherai/docs/superpowers/specs/2026-03-28-stock-futures-forecast-thread-handoff.md`
  - 本计划文件
- 任何线程如果想做 `L2 / L3`、新页面、新表、独立异步研究系统，都必须先停下，不允许按本计划继续往下写。
- 当前计划默认 backend 优先，client 次之，admin 可选后置。
