# Deep Forecast Phase 2 Evidence Expansion Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Upgrade deep forecast phase 2 so stock and futures reports consume richer in-repo evidence, produce domain-specific research conclusions, and present Forecast Lab / Forecast Report as a research system rather than a generic run workbench.

**Architecture:** Keep the existing phase 1 deep forecast pipeline and strengthen the middle layers: expand the internal research pack, replace generic dimension filling with stock/futures domain evidence builders, then update scenario/report composition to consume those richer structures. On the frontend, preserve the existing route and state shell while shifting Forecast Lab toward “recent research conclusions + run dynamics” and Forecast Report toward domain evidence cards that render the upgraded report schema.

**Tech Stack:** Go backend (`go test`), Vue 3 / Vite frontend, `node:test` source-contract tests, existing forecast L3 repo/orchestrator/report-builder stack

---

## File Structure

### Backend core files

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`
  - Expand `strategyForecastL3ResearchPack`
  - Extend `strategyForecastL3ContextReader` if existing repo methods are needed for futures evidence
  - Build stock/futures domain evidence from existing repo data
  - Enrich the LLM validation prompt with domain evidence summaries
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder.go`
  - Stop filling `dimension_evidence` from generic highlights
  - Build report state/scenario/action sections from domain evidence
  - Keep markdown/html as generated compatibility layers
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go`
  - Add any minimal new internal/public report structures needed for richer evidence semantics while keeping phase 1 compatibility
- Modify if required: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo.go`
  - Only if persistence helpers need to adapt to expanded JSON payloads

### Backend test files

- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator_test.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder_test.go`
- Modify if needed: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_repo_test.go`
- Modify if needed: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/handler/forecast_l3_handler_test.go`

### Frontend core files

- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastLabView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastLabView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.js`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-localization.js`
- Create if the detail pages need decompression:
  - `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-report-view-model.js`

### Frontend test files

- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-lab-view.test.mjs`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-lab-view.test.mjs`
- Modify if view-model logic is added:
  - `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.test.mjs`
  - `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-localization.test.mjs`

## Task 1: Expand Research Pack With Stock/Futures Domain Evidence

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/model/strategy_forecast_l3.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator_test.go`

- [ ] **Step 1: Add failing tests for domain evidence extraction**

Update `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator_test.go` with two focused tests:

```go
func TestBuildStrategyForecastL3ResearchPackIncludesStockDomainEvidence(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	run, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:    model.StrategyForecastL3TargetTypeStock,
		TargetID:      "sr_001",
		TargetKey:     "600519.SH",
		TargetLabel:   "贵州茅台",
		TriggerType:   model.StrategyForecastL3TriggerTypeUserRequest,
		RequestUserID: "user_001",
		Reason:        "从推荐页进入，想看更完整研究判断",
	})
	if err != nil {
		t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
	}

	pack, err := buildStrategyForecastL3ResearchPack(repo, run)
	if err != nil {
		t.Fatalf("buildStrategyForecastL3ResearchPack() error = %v", err)
	}

	if len(pack.StockEvidence.Technical.SupportingPoints) == 0 {
		t.Fatalf("expected stock technical evidence, got %+v", pack.StockEvidence)
	}
	if len(pack.StockEvidence.Valuation.SupportingPoints) == 0 {
		t.Fatalf("expected stock valuation evidence, got %+v", pack.StockEvidence)
	}
}

func TestBuildStrategyForecastL3ResearchPackIncludesFuturesDomainEvidence(t *testing.T) {
	repo := NewInMemoryGrowthRepo()
	run, err := repo.CreateStrategyForecastL3Run(model.StrategyForecastL3RunCreateInput{
		TargetType:    model.StrategyForecastL3TargetTypeFutures,
		TargetID:      "ft_001",
		TargetKey:     "AU2408",
		TargetLabel:   "沪金主力",
		TriggerType:   model.StrategyForecastL3TriggerTypeUserRequest,
		RequestUserID: "user_001",
		Reason:        "从策略页进入，想核对库存和结构信号",
	})
	if err != nil {
		t.Fatalf("CreateStrategyForecastL3Run() error = %v", err)
	}

	pack, err := buildStrategyForecastL3ResearchPack(repo, run)
	if err != nil {
		t.Fatalf("buildStrategyForecastL3ResearchPack() error = %v", err)
	}

	if len(pack.FuturesEvidence.SupplyDemand.SupportingPoints) == 0 {
		t.Fatalf("expected futures supply-demand evidence, got %+v", pack.FuturesEvidence)
	}
	if len(pack.FuturesEvidence.TermStructure.SupportingPoints) == 0 {
		t.Fatalf("expected futures term structure evidence, got %+v", pack.FuturesEvidence)
	}
}
```

- [ ] **Step 2: Run the new orchestrator tests to verify they fail**

Run:

```bash
go test ./backend/internal/growth/repo -run 'TestBuildStrategyForecastL3ResearchPackIncludesStockDomainEvidence|TestBuildStrategyForecastL3ResearchPackIncludesFuturesDomainEvidence'
```

Expected: FAIL because the current `strategyForecastL3ResearchPack` has no stock/futures domain evidence structure.

- [ ] **Step 3: Expand the research pack structures minimally**

In `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`, extend the internal pack with focused internal types such as:

```go
type strategyForecastL3EvidenceSlice struct {
	Summary          string
	SupportingPoints []string
	RiskPoints       []string
}

type strategyForecastL3StockEvidence struct {
	Fundamental strategyForecastL3EvidenceSlice
	Technical   strategyForecastL3EvidenceSlice
	Flow        strategyForecastL3EvidenceSlice
	Valuation   strategyForecastL3EvidenceSlice
	Event       strategyForecastL3EvidenceSlice
}

type strategyForecastL3FuturesEvidence struct {
	SupplyDemand strategyForecastL3EvidenceSlice
	TermStructure strategyForecastL3EvidenceSlice
	TapeTechnical strategyForecastL3EvidenceSlice
	PositionFlow  strategyForecastL3EvidenceSlice
	MacroEvent    strategyForecastL3EvidenceSlice
}
```

Embed them into `strategyForecastL3ResearchPack` while keeping all existing phase 1 fields intact.

- [ ] **Step 4: Populate stock domain evidence from existing insight data**

Inside `buildStrategyForecastL3ResearchPack(...)`:

- Build stock technical evidence from existing momentum / trend / volatility / volume-ratio fields available on the stock recommendation insight or explanation backing data
- Build flow evidence from money-flow / turnover / flow score signals
- Build valuation evidence from `PeTTM / PB / ValueScore`
- Build event evidence from related news and news sentiment signals
- Build fundamental evidence from existing `reason summary / explanation / performance summary / value hints`

Prefer helper functions such as:

```go
func buildStrategyForecastL3StockEvidence(insight model.StockRecommendationInsight) strategyForecastL3StockEvidence
```

- [ ] **Step 5: Populate futures domain evidence from existing futures insight/context**

Inside `buildStrategyForecastL3ResearchPack(...)`:

- Build supply-demand evidence from inventory summary / inventory pressure / inventory change style signals
- Build term-structure evidence from `BasisPct / CarryPct / TermStructurePct / CurveSlopePct / BasisTermAlignment`
- Build tape-technical evidence from trend / volatility / volume / spread pressure
- Build position-flow evidence from `OIChangePct / TurnoverRatio / FlowBias` and any current public position summary available through existing insight/explanation
- Build macro-event evidence from related news, related events, and regime

Prefer helper functions such as:

```go
func buildStrategyForecastL3FuturesEvidence(insight model.FuturesStrategyInsight) strategyForecastL3FuturesEvidence
```

- [ ] **Step 6: Re-run the research-pack tests**

Run:

```bash
go test ./backend/internal/growth/repo -run 'TestBuildStrategyForecastL3ResearchPackIncludesStockDomainEvidence|TestBuildStrategyForecastL3ResearchPackIncludesFuturesDomainEvidence'
```

Expected: PASS.

- [ ] **Step 7: Commit Task 1**

```bash
git add backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go \
        backend/internal/growth/repo/strategy_forecast_l3_orchestrator_test.go \
        backend/internal/growth/model/strategy_forecast_l3.go
git commit -m "feat: expand deep forecast research pack evidence"
```

## Task 2: Rebuild Domain Evidence, Scenario Assessment, And Validation Input

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder.go`
- Modify: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`
- Test: `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder_test.go`

- [ ] **Step 1: Add failing tests for stock/futures dimension evidence and richer scenario output**

Extend `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder_test.go` with:

```go
func TestBuildForecastL3ReportBuildsStockDimensionEvidenceFromDomainEvidence(t *testing.T) {
	report := buildStrategyForecastL3Report(
		model.StrategyForecastL3Run{
			ID:          "l3run_stock_report",
			TargetType:  model.StrategyForecastL3TargetTypeStock,
			TargetKey:   "600519.SH",
			TargetLabel: "贵州茅台",
			TriggerType: model.StrategyForecastL3TriggerTypeUserRequest,
			Status:      model.StrategyForecastL3StatusRunning,
		},
		strategyForecastL3ResearchPack{
			TargetType:   model.StrategyForecastL3TargetTypeStock,
			TargetKey:    "600519.SH",
			TargetLabel:  "贵州茅台",
			CoreThesis:   "估值承压但资金仍在支撑。",
			RiskBoundary: "跌破关键支撑位则主线失效。",
			StockEvidence: strategyForecastL3StockEvidence{
				Fundamental: strategyForecastL3EvidenceSlice{Summary: "盈利质量稳定。", SupportingPoints: []string{"绩效回撤可控"}},
				Technical:   strategyForecastL3EvidenceSlice{Summary: "20日动量维持正值。", SupportingPoints: []string{"趋势强度抬升"}},
				Flow:        strategyForecastL3EvidenceSlice{Summary: "主力净流入修复。", SupportingPoints: []string{"换手温和放大"}},
				Valuation:   strategyForecastL3EvidenceSlice{Summary: "PE处于可跟踪区间。", SupportingPoints: []string{"PB未失控"}},
				Event:       strategyForecastL3EvidenceSlice{Summary: "新闻偏正向。", SupportingPoints: []string{"资讯热度回升"}},
			},
		},
		nil,
		strategyForecastL3ValidationResult{},
		time.Date(2026, 5, 24, 12, 0, 0, 0, time.UTC),
	)

	if got := len(report.DimensionEvidence); got < 5 {
		t.Fatalf("expected at least five stock dimensions, got %d (%+v)", got, report.DimensionEvidence)
	}
	if report.ScenarioAssessment == nil || len(report.ScenarioAssessment.ActionPlan) == 0 {
		t.Fatalf("expected structured scenario assessment, got %+v", report.ScenarioAssessment)
	}
}

func TestBuildForecastL3ValidationPromptIncludesDomainEvidence(t *testing.T) {
	prompt := buildStrategyForecastL3ValidationPrompt(
		model.StrategyForecastL3Run{
			TargetType:  model.StrategyForecastL3TargetTypeFutures,
			TargetKey:   "AU2408",
			TargetLabel: "沪金主力",
		},
		strategyForecastL3ResearchPack{
			TargetType:  model.StrategyForecastL3TargetTypeFutures,
			TargetKey:   "AU2408",
			TargetLabel: "沪金主力",
			CoreThesis:  "基差与库存共同支撑近月。",
			FuturesEvidence: strategyForecastL3FuturesEvidence{
				SupplyDemand: strategyForecastL3EvidenceSlice{Summary: "库存压力缓和。", SupportingPoints: []string{"库存延续去化"}},
				TermStructure: strategyForecastL3EvidenceSlice{Summary: "近月结构占优。", SupportingPoints: []string{"基差与期限结构同向"}},
			},
		},
		nil,
	)

	if !strings.Contains(prompt, "库存压力缓和") || !strings.Contains(prompt, "近月结构占优") {
		t.Fatalf("expected domain evidence in validation prompt, got %q", prompt)
	}
}
```

- [ ] **Step 2: Run report-builder tests to verify they fail**

Run:

```bash
go test ./backend/internal/growth/repo -run 'TestBuildForecastL3ReportBuildsStockDimensionEvidenceFromDomainEvidence|TestBuildForecastL3ValidationPromptIncludesDomainEvidence|TestBuildForecastL3ReportSynthesizesExecutiveSummaryAndMarkdown'
```

Expected: FAIL because dimension evidence and validation prompt still rely on generic highlights/roles.

- [ ] **Step 3: Replace generic dimension filling with domain-driven builders**

Inside `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_report_builder.go`, add focused helpers:

```go
func buildStrategyForecastL3StockDimensionEvidence(pack strategyForecastL3ResearchPack) []model.StrategyForecastL3DimensionEvidence
func buildStrategyForecastL3FuturesDimensionEvidence(pack strategyForecastL3ResearchPack) []model.StrategyForecastL3DimensionEvidence
```

Rules:

- Use the stock/futures domain evidence slices from Task 1 as the primary source
- Keep `roles` only as a secondary compatibility source when no domain evidence exists
- Emit dimensions with the exact user-facing semantics:
  - Stock: `FUNDAMENTAL`, `TECHNICAL`, `FLOW`, `VALUATION`, `EVENT`
  - Futures: `SUPPLY_DEMAND`, `TERM_STRUCTURE`, `TAPE_TECHNICAL`, `POSITION_FLOW`, `MACRO_EVENT`

- [ ] **Step 4: Strengthen state/scenario/action composition**

Still in `strategy_forecast_l3_report_builder.go`:

- Make `buildStrategyForecastL3StateAssessment(...)` produce a more meaningful `CurrentState`
- Make `buildStrategyForecastL3ScenarioAssessment(...)` derive:
  - non-empty `SecondaryScenarios`
  - trigger conditions from enriched evidence/checklists
  - invalidation conditions from enriched risk points
  - action plan from action hints plus domain evidence

Keep compatibility with existing summary fields.

- [ ] **Step 5: Enrich the validation prompt with domain evidence**

In `/Users/gjhan21/cursor/sercherai/backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go`, update `buildStrategyForecastL3ValidationPrompt(...)` so it includes:

- stock/futures evidence summaries
- supporting points
- risk points
- current structured action plan inputs

Do not change the validation output schema.

- [ ] **Step 6: Re-run the report-builder tests**

Run:

```bash
go test ./backend/internal/growth/repo -run 'TestBuildForecastL3ReportBuildsStockDimensionEvidenceFromDomainEvidence|TestBuildForecastL3ValidationPromptIncludesDomainEvidence|TestBuildForecastL3ReportSynthesizesExecutiveSummaryAndMarkdown'
```

Expected: PASS.

- [ ] **Step 7: Run the full forecast-l3 repo slice**

Run:

```bash
go test ./backend/internal/growth/repo -run 'TestExecuteForecastL3RunBuildsReportAndLogs|TestBuildStrategyForecastL3ResearchPackIncludesStockDomainEvidence|TestBuildStrategyForecastL3ResearchPackIncludesFuturesDomainEvidence|TestBuildForecastL3ReportBuildsStockDimensionEvidenceFromDomainEvidence|TestBuildForecastL3ValidationPromptIncludesDomainEvidence|TestBuildForecastL3ReportSynthesizesExecutiveSummaryAndMarkdown'
```

Expected: PASS.

- [ ] **Step 8: Commit Task 2**

```bash
git add backend/internal/growth/repo/strategy_forecast_l3_report_builder.go \
        backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go \
        backend/internal/growth/repo/strategy_forecast_l3_report_builder_test.go \
        backend/internal/growth/repo/strategy_forecast_l3_orchestrator_test.go
git commit -m "feat: rebuild deep forecast domain evidence"
```

## Task 3: Upgrade Forecast Report To Render Stock/Futures Research Dimensions

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-localization.js`
- Create if needed: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-report-view-model.js`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-detail-view.test.mjs`

- [ ] **Step 1: Add failing source-contract tests for stock/futures evidence headings**

Update `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-detail-view.test.mjs` and the H5 twin so they assert the new evidence labels exist:

```js
assert.match(text, /基本面|技术面|资金面|估值面|事件面/);
assert.match(text, /供需库存|期限结构|盘面技术|持仓资金|宏观事件/);
assert.match(text, /当前立场|支撑点|风险点|置信度/);
```

- [ ] **Step 2: Run the detail-view source tests to verify they fail**

Run:

```bash
node --test \
  newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  newclient/src/apps/h5/views/forecast-detail-view.test.mjs
```

Expected: FAIL because the current views still expose generic `dimension` cards and not the upgraded headings.

- [ ] **Step 3: Add a shared report view-model normalizer if the detail views are getting too large**

If the detail pages are already hard to reason about, create:

```js
// /Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-report-view-model.js
export function buildForecastReportViewModel(detail) {
  return {
    stockEvidenceSections: [],
    futuresEvidenceSections: [],
    validationSummary: {},
    scenarioSummary: {},
  };
}
```

Use it to map raw `dimension_evidence` into stable UI sections with localized headings.

- [ ] **Step 4: Upgrade PC ForecastDetailView to render research-style evidence cards**

In `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastDetailView.vue`:

- Keep the seven major sections intact
- Replace the generic evidence article rendering with domain-aware cards
- Show, per evidence card:
  - heading
  - stance
  - confidence
  - summary
  - supporting points
  - risk points
- Keep `markdown_body` as a secondary/collapsed “原始报告全文” surface

- [ ] **Step 5: Mirror the same evidence rendering semantics in H5**

In `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastDetailView.vue`:

- Reuse the same view-model logic
- Keep mobile layout compressed, but preserve all evidence headings and summaries
- Do not split into a separate H5-only data model

- [ ] **Step 6: Update localization helpers for evidence headings and stance labels**

In `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-localization.js`:

- Add localized mappings for:
  - `FUNDAMENTAL`, `TECHNICAL`, `FLOW`, `VALUATION`, `EVENT`
  - `SUPPLY_DEMAND`, `TERM_STRUCTURE`, `TAPE_TECHNICAL`, `POSITION_FLOW`, `MACRO_EVENT`
- Keep existing state / verdict localization intact

- [ ] **Step 7: Re-run the detail-view tests**

Run:

```bash
node --test \
  newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  newclient/src/apps/h5/views/forecast-detail-view.test.mjs \
  newclient/src/shared/lib/forecast-localization.test.mjs
```

Expected: PASS.

- [ ] **Step 8: Commit Task 3**

```bash
git add newclient/src/apps/pc/views/forecast/ForecastDetailView.vue \
        newclient/src/apps/h5/views/forecast/ForecastDetailView.vue \
        newclient/src/shared/lib/forecast-localization.js \
        newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
        newclient/src/apps/h5/views/forecast-detail-view.test.mjs \
        newclient/src/shared/lib/forecast-localization.test.mjs \
        newclient/src/shared/lib/forecast-report-view-model.js
git commit -m "feat: upgrade deep forecast report evidence views"
```

## Task 4: Shift Forecast Lab From Run Workbench To Research Center

**Files:**
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastLabView.vue`
- Modify: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastLabView.vue`
- Modify if needed: `/Users/gjhan21/cursor/sercherai/newclient/src/shared/lib/forecast-summary.js`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast-lab-view.test.mjs`
- Test: `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast-lab-view.test.mjs`

- [ ] **Step 1: Add failing source-contract tests for “recent research conclusions” and “run dynamics”**

Update the PC/H5 lab tests to assert the new wording and structure:

```js
assert.match(text, /最近研究结论/);
assert.match(text, /运行动态/);
assert.match(text, /最近研究状态|结构化报告|模型复核/);
```

- [ ] **Step 2: Run the Forecast Lab tests to verify they fail**

Run:

```bash
node --test \
  newclient/src/apps/pc/views/forecast-lab-view.test.mjs \
  newclient/src/apps/h5/views/forecast-lab-view.test.mjs
```

Expected: FAIL because the current copy and structure are still “工作台 / 最近深推演入口 / 运行清单”.

- [ ] **Step 3: Reframe recent entries as recent research conclusions**

In `/Users/gjhan21/cursor/sercherai/newclient/src/apps/pc/views/forecast/ForecastLabView.vue`:

- Rename the “最近深推演入口” semantics to “最近研究结论”
- Surface, per card:
  - target
  - current state / status
  - one-line conclusion
  - primary scenario
  - risk boundary hint
  - validation status

If `forecast-summary.js` needs richer derived labels, extend it there instead of bloating the view.

- [ ] **Step 4: Split run list semantics into run dynamics**

Still in the PC view:

- Reframe the run list area as “运行动态”
- Keep queued/running/failed visibility, but visually separate it from research conclusions
- Preserve current strict-context entry behavior

- [ ] **Step 5: Mirror the same product semantics in H5**

In `/Users/gjhan21/cursor/sercherai/newclient/src/apps/h5/views/forecast/ForecastLabView.vue`:

- Keep mobile layout lightweight
- Preserve the same two main surfaces:
  - recent research conclusions
  - run dynamics
- Do not invent H5-only behavior

- [ ] **Step 6: Re-run the Forecast Lab tests**

Run:

```bash
node --test \
  newclient/src/apps/pc/views/forecast-lab-view.test.mjs \
  newclient/src/apps/h5/views/forecast-lab-view.test.mjs \
  newclient/src/shared/lib/forecast-summary.test.mjs
```

Expected: PASS.

- [ ] **Step 7: Run the broader frontend forecast regression slice**

Run:

```bash
node --test \
  newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  newclient/src/apps/h5/views/forecast-detail-view.test.mjs \
  newclient/src/apps/pc/views/forecast-lab-view.test.mjs \
  newclient/src/apps/h5/views/forecast-lab-view.test.mjs \
  newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs \
  newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs \
  newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs \
  newclient/src/apps/h5/views/h5-futures-detail-deep-forecast.test.mjs \
  newclient/src/shared/lib/forecast-summary.test.mjs \
  newclient/src/shared/lib/forecast-localization.test.mjs
```

Expected: PASS.

- [ ] **Step 8: Run the production build**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/newclient && npm run build
```

Expected: PASS with generated `dist/index.html` and `dist-h5/m/index.html`.

- [ ] **Step 9: Commit Task 4**

```bash
git add newclient/src/apps/pc/views/forecast/ForecastLabView.vue \
        newclient/src/apps/h5/views/forecast/ForecastLabView.vue \
        newclient/src/shared/lib/forecast-summary.js \
        newclient/src/apps/pc/views/forecast-lab-view.test.mjs \
        newclient/src/apps/h5/views/forecast-lab-view.test.mjs \
        newclient/src/shared/lib/forecast-summary.test.mjs
git commit -m "feat: reposition forecast lab as research center"
```

## Final Verification

- [ ] **Step 1: Run the backend phase 2 verification slice**

Run:

```bash
go test ./backend/internal/growth/repo -run 'TestExecuteForecastL3RunBuildsReportAndLogs|TestBuildStrategyForecastL3ResearchPackIncludesStockDomainEvidence|TestBuildStrategyForecastL3ResearchPackIncludesFuturesDomainEvidence|TestBuildForecastL3ReportBuildsStockDimensionEvidenceFromDomainEvidence|TestBuildForecastL3ValidationPromptIncludesDomainEvidence|TestBuildForecastL3ReportSynthesizesExecutiveSummaryAndMarkdown'
```

Expected: PASS.

- [ ] **Step 2: Run the frontend phase 2 verification slice**

Run:

```bash
node --test \
  newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
  newclient/src/apps/h5/views/forecast-detail-view.test.mjs \
  newclient/src/apps/pc/views/forecast-lab-view.test.mjs \
  newclient/src/apps/h5/views/forecast-lab-view.test.mjs \
  newclient/src/apps/pc/views/stock-analysis-deep-forecast.test.mjs \
  newclient/src/apps/pc/views/futures-strategy-deep-forecast.test.mjs \
  newclient/src/apps/h5/views/h5-stock-detail-deep-forecast.test.mjs \
  newclient/src/apps/h5/views/h5-futures-detail-deep-forecast.test.mjs \
  newclient/src/shared/lib/forecast-summary.test.mjs \
  newclient/src/shared/lib/forecast-localization.test.mjs
```

Expected: PASS.

- [ ] **Step 3: Run the frontend production build**

Run:

```bash
cd /Users/gjhan21/cursor/sercherai/newclient && npm run build
```

Expected: PASS.

- [ ] **Step 4: Final checkpoint commit**

```bash
git add backend/internal/growth/model/strategy_forecast_l3.go \
        backend/internal/growth/repo/strategy_forecast_l3_orchestrator.go \
        backend/internal/growth/repo/strategy_forecast_l3_report_builder.go \
        backend/internal/growth/repo/strategy_forecast_l3_orchestrator_test.go \
        backend/internal/growth/repo/strategy_forecast_l3_report_builder_test.go \
        newclient/src/apps/pc/views/forecast/ForecastLabView.vue \
        newclient/src/apps/h5/views/forecast/ForecastLabView.vue \
        newclient/src/apps/pc/views/forecast/ForecastDetailView.vue \
        newclient/src/apps/h5/views/forecast/ForecastDetailView.vue \
        newclient/src/shared/lib/forecast-summary.js \
        newclient/src/shared/lib/forecast-localization.js \
        newclient/src/shared/lib/forecast-report-view-model.js \
        newclient/src/apps/pc/views/forecast-detail-view.test.mjs \
        newclient/src/apps/h5/views/forecast-detail-view.test.mjs \
        newclient/src/apps/pc/views/forecast-lab-view.test.mjs \
        newclient/src/apps/h5/views/forecast-lab-view.test.mjs \
        newclient/src/shared/lib/forecast-summary.test.mjs \
        newclient/src/shared/lib/forecast-localization.test.mjs
git commit -m "feat: deliver deep forecast phase2 evidence expansion"
```
