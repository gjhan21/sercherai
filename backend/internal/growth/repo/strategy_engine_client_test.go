package repo

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sercherai/backend/internal/growth/model"
)

func TestBuildStockSelectionJobPayloadDefaultsToAutoMode(t *testing.T) {
	payload := buildStockSelectionJobPayload("2026-03-21", nil)

	if payload["selection_mode"] != strategyEngineStockSelectionModeAuto {
		t.Fatalf("expected AUTO selection_mode, got %#v", payload["selection_mode"])
	}
	if payload["profile_id"] != strategyEngineDefaultStockSelectionProfileID {
		t.Fatalf("unexpected profile_id: %#v", payload["profile_id"])
	}
	if payload["universe_scope"] != strategyEngineDefaultStockUniverseScope || payload["market_scope"] != strategyEngineDefaultStockUniverseScope {
		t.Fatalf("unexpected universe scope payload: %#v", payload)
	}
	if _, exists := payload["seed_symbols"]; exists {
		t.Fatalf("expected AUTO payload to omit seed_symbols, got %#v", payload["seed_symbols"])
	}
}

func TestBuildStockSelectionJobPayloadUsesManualModeWhenSeedSetProvided(t *testing.T) {
	payload := buildStockSelectionJobPayload("2026-03-21", &model.StrategySeedSet{
		ID:    "seed_stock_manual",
		Items: []string{"600519.SH", "300750.SZ"},
	})

	if payload["selection_mode"] != strategyEngineStockSelectionModeManual {
		t.Fatalf("expected MANUAL selection_mode, got %#v", payload["selection_mode"])
	}
	if got, ok := payload["seed_symbols"].([]string); !ok || len(got) != 2 {
		t.Fatalf("expected seed_symbols to be populated, got %#v", payload["seed_symbols"])
	}
}

func TestStrategyEngineClientGeneratesStockSelectionReport(t *testing.T) {
	jobStatusCalls := 0
	var stockJobRequest map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/internal/v1/jobs/stock-selection":
			if err := json.NewDecoder(r.Body).Decode(&stockJobRequest); err != nil {
				t.Fatalf("decode stock request body: %v", err)
			}
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(strategyEngineJobAccepted{JobID: "job_test_001"})
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/jobs/job_test_001":
			jobStatusCalls++
			status := "RUNNING"
			report := strategyEngineStockSelectionReport{}
			if jobStatusCalls >= 2 {
				status = "SUCCEEDED"
				report = strategyEngineStockSelectionReport{
					SelectedCount: 1,
					PublishPayload: []strategyEngineStockPublishPayload{
						{
							Recommendation: strategyEngineRecommendation{
								Symbol:           "600519.SH",
								Name:             "贵州茅台",
								Score:            91.2,
								RiskLevel:        "LOW",
								PositionRange:    "10%-15%",
								ValidFrom:        "2026-03-17T00:00:00Z",
								ValidTo:          "2026-03-18T00:00:00Z",
								Status:           "PUBLISHED",
								ReasonSummary:    "趋势和资金面较强",
								SourceType:       "SYSTEM",
								StrategyVersion:  "stock-mvp-v1",
								Publisher:        "strategy-engine",
								PerformanceLabel: "ESTIMATED",
							},
							Detail: strategyEngineDetail{
								TechScore:      88,
								FundScore:      84,
								SentimentScore: 79,
								MoneyFlowScore: 86,
								TakeProfit:     "上涨8%-12%分批止盈",
								StopLoss:       "回撤3%止损",
								RiskNote:       "仅供参考",
							},
						},
					},
				}
			}
			_ = json.NewEncoder(w).Encode(strategyEngineAdminJobRecord{
				JobID:       "job_test_001",
				JobType:     "stock-selection",
				Status:      status,
				RequestedBy: "go-backend",
				TraceID:     "trace-stock-001",
				Payload: map[string]any{
					"trade_date": "2026-03-17",
				},
				Result: &strategyEngineAdminJobResult{
					Summary:   "本次输出 1 条可发布推荐",
					Artifacts: map[string]any{"report": map[string]any{"selected_count": report.SelectedCount, "publish_payloads": report.PublishPayload}},
					Warnings:  []string{"示例 warning"},
				},
				CreatedAt:  "2026-03-17T00:00:00Z",
				StartedAt:  "2026-03-17T00:00:10Z",
				FinishedAt: "2026-03-17T00:00:20Z",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/internal/v1/publish/jobs/job_test_001":
			_ = json.NewEncoder(w).Encode(strategyEnginePublishRecord{
				PublishID:     "publish_test_001",
				Version:       3,
				ReportSummary: "本次输出 1 条可发布推荐",
				SelectedCount: 1,
				PayloadCount:  1,
				TradeDate:     "2026-03-17",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &strategyEngineClient{
		baseURL:      server.URL,
		httpClient:   &http.Client{Timeout: 2 * time.Second},
		pollInterval: 10 * time.Millisecond,
	}

	report, err := client.generateStockSelectionReport("2026-03-17", nil, nil, nil, nil)
	if err != nil {
		t.Fatalf("generateStockSelectionReport returned error: %v", err)
	}
	if report.SelectedCount != 1 {
		t.Fatalf("expected selected_count=1, got %d", report.SelectedCount)
	}
	if len(report.PublishPayload) != 1 {
		t.Fatalf("expected one publish payload, got %d", len(report.PublishPayload))
	}
	if report.PublishPayload[0].Recommendation.Symbol != "600519.SH" {
		t.Fatalf("unexpected symbol: %s", report.PublishPayload[0].Recommendation.Symbol)
	}
	if report.PublishID != "publish_test_001" {
		t.Fatalf("unexpected publish id: %s", report.PublishID)
	}
	if report.PublishVersion != 3 {
		t.Fatalf("unexpected publish version: %d", report.PublishVersion)
	}
	if report.JobRecord.JobID != "job_test_001" || report.JobRecord.StorageSource != "" {
		t.Fatalf("unexpected archived job record: %#v", report.JobRecord)
	}
	payload, ok := stockJobRequest["payload"].(map[string]any)
	if !ok {
		t.Fatalf("expected payload in stock request")
	}
	configRefs, ok := payload["config_refs"].(map[string]any)
	if !ok {
		t.Fatalf("expected config_refs in stock payload")
	}
	if configRefs["seed_set"] != nil || configRefs["agent_profile"] != nil || configRefs["scenario_template"] != nil || configRefs["publish_policy"] != nil {
		t.Fatalf("expected nil config refs when no default config is provided, got %#v", configRefs)
	}
	if preview, exists := payload["publish_policy_preview"]; !exists || preview != nil {
		t.Fatalf("expected publish_policy_preview to exist and be nil, got %#v", preview)
	}
	if payload["selection_mode"] != strategyEngineStockSelectionModeAuto {
		t.Fatalf("expected AUTO selection_mode in request payload, got %#v", payload["selection_mode"])
	}
	if payload["profile_id"] != strategyEngineDefaultStockSelectionProfileID {
		t.Fatalf("unexpected profile_id in request payload: %#v", payload["profile_id"])
	}
	if payload["universe_scope"] != strategyEngineDefaultStockUniverseScope {
		t.Fatalf("unexpected universe_scope in request payload: %#v", payload["universe_scope"])
	}
	if _, exists := payload["seed_symbols"]; exists {
		t.Fatalf("expected AUTO request payload to omit seed_symbols, got %#v", payload["seed_symbols"])
	}
}

func TestStrategyEngineClientGeneratesFuturesStrategyReport(t *testing.T) {
	jobStatusCalls := 0
	var futuresJobRequest map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/internal/v1/jobs/futures-strategy":
			if err := json.NewDecoder(r.Body).Decode(&futuresJobRequest); err != nil {
				t.Fatalf("decode futures request body: %v", err)
			}
			w.WriteHeader(http.StatusAccepted)
			_ = json.NewEncoder(w).Encode(strategyEngineJobAccepted{JobID: "job_futures_001"})
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/jobs/job_futures_001":
			jobStatusCalls++
			status := "RUNNING"
			report := strategyEngineFuturesStrategyReport{}
			if jobStatusCalls >= 2 {
				status = "SUCCEEDED"
				report = strategyEngineFuturesStrategyReport{
					SelectedCount: 1,
					PublishPayload: []strategyEngineFuturesPublishPayload{
						{
							Strategy: strategyEngineFuturesStrategy{
								Contract:      "CU2405",
								Name:          "沪铜主连",
								Direction:     "LONG",
								RiskLevel:     "MEDIUM",
								PositionRange: "20%-30%",
								ValidFrom:     "2026-03-17T09:00:00Z",
								ValidTo:       "2026-03-18T15:00:00Z",
								Status:        "PUBLISHED",
								ReasonSummary: "趋势突破且成交量放大",
							},
							Guidance: strategyEngineFuturesGuidance{
								Contract:          "CU2405",
								GuidanceDirection: "LONG",
								PositionLevel:     "MEDIUM",
								EntryRange:        "78000-78200",
								TakeProfitRange:   "78800-79200",
								StopLossRange:     "77500-77600",
								RiskLevel:         "MEDIUM",
								InvalidCondition:  "跌破77500",
								ValidTo:           "2026-03-18T15:00:00Z",
							},
						},
					},
				}
			}
			_ = json.NewEncoder(w).Encode(strategyEngineAdminJobRecord{
				JobID:       "job_futures_001",
				JobType:     "futures-strategy",
				Status:      status,
				RequestedBy: "go-backend",
				TraceID:     "trace-futures-001",
				Payload: map[string]any{
					"trade_date": "2026-03-17",
				},
				Result: &strategyEngineAdminJobResult{
					Summary:   "本次输出 1 条期货策略",
					Artifacts: map[string]any{"report": map[string]any{"selected_count": report.SelectedCount, "publish_payloads": report.PublishPayload}},
					Warnings:  []string{"示例 futures warning"},
				},
				CreatedAt:  "2026-03-17T09:00:00Z",
				StartedAt:  "2026-03-17T09:00:10Z",
				FinishedAt: "2026-03-17T09:00:20Z",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/internal/v1/publish/jobs/job_futures_001":
			_ = json.NewEncoder(w).Encode(strategyEnginePublishRecord{
				PublishID:     "publish_futures_001",
				Version:       5,
				ReportSummary: "本次输出 1 条期货策略",
				SelectedCount: 1,
				PayloadCount:  1,
				TradeDate:     "2026-03-17",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &strategyEngineClient{
		baseURL:      server.URL,
		httpClient:   &http.Client{Timeout: 2 * time.Second},
		pollInterval: 10 * time.Millisecond,
	}

	seedSet := &model.StrategySeedSet{
		ID:         "seed_futures_default",
		Name:       "期货默认种子池",
		TargetType: "FUTURES",
		IsDefault:  true,
		UpdatedAt:  "2026-03-18T10:00:00Z",
		Items:      []string{"CU2405"},
	}
	profile := &model.StrategyAgentProfile{
		ID:                              "agent_futures_default",
		Name:                            "期货默认五角评审",
		TargetType:                      "FUTURES",
		IsDefault:                       true,
		UpdatedAt:                       "2026-03-18T10:01:00Z",
		EnabledAgents:                   []string{"trend", "risk"},
		PositiveThreshold:               3,
		NegativeThreshold:               2,
		AllowVeto:                       true,
		AllowMockFallbackOnShortHistory: false,
	}
	scenarioTemplate := &model.StrategyScenarioTemplate{
		ID:         "scenario_futures_default",
		Name:       "期货四场景模板",
		TargetType: "FUTURES",
		IsDefault:  true,
		UpdatedAt:  "2026-03-18T10:02:00Z",
	}
	publishPolicy := &model.StrategyPublishPolicy{
		ID:                   "policy_futures_default",
		Name:                 "期货默认发布策略",
		TargetType:           "FUTURES",
		IsDefault:            true,
		UpdatedAt:            "2026-03-18T10:03:00Z",
		MaxRiskLevel:         "HIGH",
		MaxWarningCount:      1,
		AllowVetoedPublish:   false,
		DefaultPublisher:     "strategy-engine",
		OverrideNoteTemplate: "人工覆盖发布需说明原因",
	}

	report, err := client.generateFuturesStrategyReport("2026-03-17", seedSet, profile, scenarioTemplate, publishPolicy)
	if err != nil {
		t.Fatalf("generateFuturesStrategyReport returned error: %v", err)
	}
	if report.SelectedCount != 1 {
		t.Fatalf("expected selected_count=1, got %d", report.SelectedCount)
	}
	if len(report.PublishPayload) != 1 {
		t.Fatalf("expected one futures publish payload, got %d", len(report.PublishPayload))
	}
	if report.PublishPayload[0].Strategy.Contract != "CU2405" {
		t.Fatalf("unexpected futures contract: %s", report.PublishPayload[0].Strategy.Contract)
	}
	if report.PublishPayload[0].Guidance.EntryRange != "78000-78200" {
		t.Fatalf("unexpected futures entry range: %s", report.PublishPayload[0].Guidance.EntryRange)
	}
	if report.PublishID != "publish_futures_001" {
		t.Fatalf("unexpected publish id: %s", report.PublishID)
	}
	if report.PublishVersion != 5 {
		t.Fatalf("unexpected publish version: %d", report.PublishVersion)
	}
	if report.JobRecord.JobID != "job_futures_001" {
		t.Fatalf("unexpected archived futures job record: %#v", report.JobRecord)
	}
	payload, ok := futuresJobRequest["payload"].(map[string]any)
	if !ok {
		t.Fatalf("expected payload in futures request")
	}
	if payload["allow_mock_fallback_on_short_history"] != false {
		t.Fatalf("expected futures payload to honor profile fallback toggle, got %#v", payload["allow_mock_fallback_on_short_history"])
	}
	configRefs, ok := payload["config_refs"].(map[string]any)
	if !ok {
		t.Fatalf("expected config_refs in futures payload")
	}
	seedRef, ok := configRefs["seed_set"].(map[string]any)
	if !ok {
		t.Fatalf("expected seed_set ref in futures payload")
	}
	if seedRef["id"] != "seed_futures_default" || seedRef["source"] != "admin-default" {
		t.Fatalf("unexpected seed ref: %#v", seedRef)
	}
	agentRef, ok := configRefs["agent_profile"].(map[string]any)
	if !ok || agentRef["id"] != "agent_futures_default" {
		t.Fatalf("unexpected agent ref: %#v", configRefs["agent_profile"])
	}
	scenarioRef, ok := configRefs["scenario_template"].(map[string]any)
	if !ok || scenarioRef["id"] != "scenario_futures_default" {
		t.Fatalf("unexpected scenario ref: %#v", configRefs["scenario_template"])
	}
	policyRef, ok := configRefs["publish_policy"].(map[string]any)
	if !ok || policyRef["id"] != "policy_futures_default" {
		t.Fatalf("unexpected policy ref: %#v", configRefs["publish_policy"])
	}
	preview, ok := payload["publish_policy_preview"].(map[string]any)
	if !ok {
		t.Fatalf("expected publish_policy_preview in futures payload")
	}
	if preview["max_risk_level"] != "HIGH" || preview["allow_vetoed_publish"] != false {
		t.Fatalf("unexpected publish policy preview: %#v", preview)
	}
}

func TestStrategyEngineClientPublishHistoryReplayAndCompare(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/history/stock-selection":
			_ = json.NewEncoder(w).Encode(strategyEnginePublishHistoryResponse{
				Records: []strategyEnginePublishRecordSummary{
					{
						PublishID:     "publish_hist_001",
						JobID:         "job_hist_001",
						JobType:       "stock-selection",
						Version:       7,
						CreatedAt:     "2026-03-17T00:00:00Z",
						TradeDate:     "2026-03-17",
						ReportSummary: "生成 3 条推荐",
						SelectedCount: 3,
						AssetKeys:     []string{"600519.SH", "601318.SH"},
						PayloadCount:  3,
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/records/publish_hist_001":
			_ = json.NewEncoder(w).Encode(strategyEnginePublishRecordDetail{
				PublishID:     "publish_hist_001",
				JobID:         "job_hist_001",
				JobType:       "stock-selection",
				Version:       7,
				CreatedAt:     "2026-03-17T00:00:00Z",
				TradeDate:     "2026-03-17",
				ReportSummary: "生成 3 条推荐",
				SelectedCount: 3,
				AssetKeys:     []string{"600519.SH", "601318.SH"},
				PayloadCount:  3,
				Markdown:      "# 股票推荐发布报告",
				HTML:          "<html><body><h1>股票推荐发布报告</h1></body></html>",
				PublishPayloads: []map[string]any{
					{"symbol": "600519.SH", "score": 91.2},
				},
				ReportSnapshot: map[string]any{
					"trade_date": "2026-03-17",
					"candidates": []map[string]any{{"symbol": "600519.SH"}},
				},
				Replay: strategyEnginePublishReplay{
					WarningCount:      1,
					WarningMessages:   []string{"风控过滤 1 条"},
					VetoedAssets:      []string{"688981.SH"},
					InvalidatedAssets: []string{"600519.SH"},
					Notes:             []string{"本次需要重点复盘仓位控制"},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/records/publish_hist_001/replay":
			_ = json.NewEncoder(w).Encode(strategyEnginePublishReplay{
				WarningCount:      1,
				WarningMessages:   []string{"风控过滤 1 条"},
				VetoedAssets:      []string{"688981.SH"},
				InvalidatedAssets: []string{"600519.SH"},
				Notes:             []string{"本次需要重点复盘仓位控制"},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/internal/v1/publish/compare":
			_ = json.NewEncoder(w).Encode(strategyEnginePublishCompareResult{
				LeftPublishID:      "publish_hist_001",
				RightPublishID:     "publish_hist_002",
				LeftVersion:        7,
				RightVersion:       8,
				SelectedCountDelta: 1,
				PayloadCountDelta:  1,
				WarningCountDelta:  -1,
				VetoCountDelta:     0,
				AddedAssets:        []string{"300750.SZ"},
				RemovedAssets:      []string{"601318.SH"},
				SharedAssets:       []string{"600519.SH"},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &strategyEngineClient{
		baseURL:      server.URL,
		httpClient:   &http.Client{Timeout: 2 * time.Second},
		pollInterval: 10 * time.Millisecond,
	}

	history, err := client.listPublishHistory("stock-selection")
	if err != nil {
		t.Fatalf("listPublishHistory returned error: %v", err)
	}
	if len(history) != 1 || history[0].PublishID != "publish_hist_001" {
		t.Fatalf("unexpected publish history: %+v", history)
	}

	record, err := client.getPublishRecord("publish_hist_001")
	if err != nil {
		t.Fatalf("getPublishRecord returned error: %v", err)
	}
	if record.Version != 7 || len(record.PublishPayloads) != 1 || record.Replay.WarningCount != 1 {
		t.Fatalf("unexpected publish record: %+v", record)
	}

	replay, err := client.getPublishReplay("publish_hist_001")
	if err != nil {
		t.Fatalf("getPublishReplay returned error: %v", err)
	}
	if replay.WarningCount != 1 || len(replay.VetoedAssets) != 1 {
		t.Fatalf("unexpected replay payload: %+v", replay)
	}

	compareResult, err := client.comparePublishVersions("publish_hist_001", "publish_hist_002")
	if err != nil {
		t.Fatalf("comparePublishVersions returned error: %v", err)
	}
	if compareResult.RightVersion != 8 || len(compareResult.AddedAssets) != 1 {
		t.Fatalf("unexpected compare result: %+v", compareResult)
	}
}

func TestStrategyEngineClientListsAndGetsJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/jobs":
			if got := r.URL.Query().Get("job_type"); got != "stock-selection" {
				t.Fatalf("expected job_type query, got %q", got)
			}
			if got := r.URL.Query().Get("status"); got != "SUCCEEDED" {
				t.Fatalf("expected status query, got %q", got)
			}
			_ = json.NewEncoder(w).Encode(strategyEngineJobListResponse{
				Items: []strategyEngineAdminJobRecord{
					{
						JobID:       "job_list_001",
						JobType:     "stock-selection",
						Status:      "SUCCEEDED",
						RequestedBy: "ops-admin",
						TraceID:     "trace-001",
						Payload: map[string]any{
							"trade_date": "2026-03-17",
							"seed_symbols": []string{
								"600519.SH",
								"601318.SH",
							},
						},
						Result: &strategyEngineAdminJobResult{
							Summary: "stock-selection completed with 2 publish-ready candidates",
							Warnings: []string{
								"risk agent vetoed 1 symbol",
							},
							Artifacts: map[string]any{
								"report": map[string]any{
									"selected_count": 2,
								},
							},
						},
						CreatedAt:  "2026-03-17T09:00:00Z",
						StartedAt:  "2026-03-17T09:00:01Z",
						FinishedAt: "2026-03-17T09:00:02Z",
					},
				},
				Total:    1,
				Page:     1,
				PageSize: 10,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/jobs/job_list_001":
			_ = json.NewEncoder(w).Encode(strategyEngineAdminJobRecord{
				JobID:       "job_list_001",
				JobType:     "stock-selection",
				Status:      "SUCCEEDED",
				RequestedBy: "ops-admin",
				TraceID:     "trace-001",
				Payload: map[string]any{
					"trade_date": "2026-03-17",
				},
				Result: &strategyEngineAdminJobResult{
					Summary: "stock-selection completed with 2 publish-ready candidates",
					Warnings: []string{
						"risk agent vetoed 1 symbol",
					},
				},
				CreatedAt: "2026-03-17T09:00:00Z",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/jobs/job_missing":
			http.NotFound(w, r)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &strategyEngineClient{
		baseURL:      server.URL,
		httpClient:   &http.Client{Timeout: 2 * time.Second},
		pollInterval: 10 * time.Millisecond,
	}

	items, total, err := client.listJobs("stock-selection", "SUCCEEDED", 1, 10)
	if err != nil {
		t.Fatalf("listJobs returned error: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected total=1, got %d", total)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 job, got %d", len(items))
	}
	if items[0].Result == nil || items[0].Result.Summary == "" {
		t.Fatalf("expected populated result summary, got %+v", items[0].Result)
	}

	item, err := client.getJobRecord("job_list_001")
	if err != nil {
		t.Fatalf("getJobRecord returned error: %v", err)
	}
	if item.JobID != "job_list_001" {
		t.Fatalf("unexpected job id: %s", item.JobID)
	}
	if item.Result == nil || len(item.Result.Warnings) != 1 {
		t.Fatalf("expected one warning, got %+v", item.Result)
	}

	_, err = client.getJobRecord("job_missing")
	if err == nil {
		t.Fatalf("expected missing job to return error")
	}
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}
