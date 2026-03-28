package repo

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"sercherai/backend/internal/growth/model"
)

const (
	localAssetContextQueryPattern          = `(?s)SELECT\s+r\.job_id,\s+COALESCE\(CAST\(r\.payload_snapshot AS CHAR\), ''\),\s+COALESCE\(DATE_FORMAT\(r\.remote_created_at, '%Y-%m-%dT%H:%i:%sZ'\), ''\),\s+COALESCE\(DATE_FORMAT\(r\.trade_date, '%Y-%m-%d'\), ''\),\s+COALESCE\(CAST\(a\.report_snapshot AS CHAR\), ''\),\s+COALESCE\(j\.publish_id, ''\),\s+COALESCE\(j\.publish_version, 0\),\s+COALESCE\(DATE_FORMAT\(j\.created_at, '%Y-%m-%dT%H:%i:%sZ'\), ''\),\s+COALESCE\(CAST\(j\.replay_snapshot AS CHAR\), ''\)\s+FROM strategy_job_runs r`
	stockEventEvidenceQueryPattern         = `(?s)SELECT\s+c\.id,\s+c\.title,\s+c\.event_type,.*FROM stock_event_clusters c`
	stockRecommendationVersionQueryPattern = `(?s)SELECT id, symbol, valid_from, reason_summary, strategy_version\s+FROM stock_recommendations\s+WHERE id = \?`
	stockRecommendationDetailQueryPattern  = `(?s)SELECT d\.reco_id, d\.tech_score, d\.fund_score, d\.sentiment_score, d\.money_flow_score, d\.take_profit, d\.stop_loss, d\.risk_note\s+FROM stock_reco_details d`
	futuresStrategyDetailQueryPattern      = `(?s)SELECT id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary\s+FROM futures_strategies\s+WHERE id = \?`
	futuresGuidanceQueryPattern            = `(?s)SELECT contract, guidance_direction, position_level, entry_range, take_profit_range, stop_loss_range, risk_level, invalid_condition, valid_to\s+FROM futures_guidances\s+WHERE contract = \?`
)

type strategyEngineAssetContextFixture struct {
	JobID           string
	Payload         map[string]any
	RemoteCreatedAt string
	TradeDate       string
	Report          map[string]any
	PublishID       string
	PublishVersion  int
	ReplayCreatedAt string
	Replay          map[string]any
}

func TestBuildStockStrategyExplanationUsesLocalSnapshotWithoutRemoteFetch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildExplanationPublishRecordDetail(
		"publish_local_explanation_001",
		"job_local_explanation_001",
		14,
		"2026-03-18",
		[]string{"600519.SH", "300750.SZ"},
		"本地快照图谱显示消费龙头资金回流。",
		"三位代理一致认为继续跟踪。",
		"本地快照理由",
		"stock-local-v2",
		[]string{"跌破 5 日线"},
		[]string{"本地快照告警"},
		[]string{},
		[]string{"本地回放备注"},
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected remote request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	repo := &MySQLGrowthRepo{
		db: db,
		strategyEngine: &strategyEngineClient{
			baseURL:      server.URL,
			httpClient:   &http.Client{Timeout: 2 * time.Second},
			pollInterval: 10 * time.Millisecond,
		},
	}

	mock.ExpectQuery(localAssetContextQueryPattern).
		WithArgs("stock-selection").
		WillReturnRows(newLocalAssetContextRows(buildStrategyEngineAssetContextFixture(record, record.AssetKeys)))
	mock.ExpectQuery(stockEventEvidenceQueryPattern).
		WithArgs("600519.SH", "600519.SH", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "event_type", "primary_symbol", "topic_label", "sector_label", "review_priority", "review_note", "published_at"}).
			AddRow("sec_demo_001", "白酒景气事件", "INDUSTRY_THEME", "600519.SH", "白酒", "消费", "HIGH", "事件与龙头走势一致", time.Date(2026, 3, 17, 9, 30, 0, 0, time.UTC)))

	explanation := repo.buildStockStrategyExplanation(
		model.StockRecommendation{
			ID:              "reco_local_001",
			Symbol:          "600519.SH",
			ValidFrom:       "2026-03-18T09:35:00Z",
			ReasonSummary:   "基础原因总结",
			StrategyVersion: "stock-local-v2",
		},
		model.StockRecommendationDetail{
			RecoID:   "reco_local_001",
			RiskNote: "基础风险提示",
		},
	)

	if explanation.PublishID != record.PublishID || explanation.JobID != record.JobID {
		t.Fatalf("expected local snapshot identity, got %+v", explanation)
	}
	if explanation.PublishVersion != record.Version || explanation.TradeDate != record.TradeDate {
		t.Fatalf("unexpected local publish metadata: %+v", explanation)
	}
	if explanation.GraphSummary != "本地快照图谱显示消费龙头资金回流。" || explanation.ConsensusSummary != "三位代理一致认为继续跟踪。" {
		t.Fatalf("expected explanation to use local report snapshot, got %+v", explanation)
	}
	if strings.Join(explanation.SeedHighlights, ",") != "600519.SH,300750.SZ" {
		t.Fatalf("unexpected seed highlights: %+v", explanation.SeedHighlights)
	}
	if explanation.SeedSummary != "本次先处理 2 个种子输入，再筛到 1 个可解释候选。" {
		t.Fatalf("unexpected seed summary: %+v", explanation)
	}
	if strings.Join(explanation.RiskFlags, ",") != "本地快照告警,本地回放备注" {
		t.Fatalf("unexpected risk flags: %+v", explanation.RiskFlags)
	}
	if strings.Join(explanation.Invalidations, ",") != "跌破 5 日线" {
		t.Fatalf("unexpected invalidations: %+v", explanation.Invalidations)
	}
	if len(explanation.Simulations) != 1 || len(explanation.AgentOpinions) != 1 {
		t.Fatalf("expected explanation simulation and agent opinions, got %+v", explanation)
	}
	if !explanation.ConfidenceCalibration.AdvisoryOnly || explanation.ConfidenceCalibration.AdjustedConfidence <= 0 {
		t.Fatalf("expected stock explanation confidence calibration, got %+v", explanation.ConfidenceCalibration)
	}
	if len(explanation.RelatedEvents) != 1 || explanation.RelatedEvents[0].ClusterID != "sec_demo_001" {
		t.Fatalf("expected related reviewed events, got %+v", explanation.RelatedEvents)
	}
	if len(explanation.EventEvidenceCards) != 1 || explanation.EventEvidenceCards[0].Value != "白酒景气事件" {
		t.Fatalf("expected event evidence cards, got %+v", explanation.EventEvidenceCards)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestBuildStockStrategyExplanationIncludesRealContextSummary(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildExplanationPublishRecordDetail(
		"publish_local_context_001",
		"job_local_context_001",
		16,
		"2026-03-18",
		[]string{"600519.SH", "300750.SZ"},
		"上下游图谱显示消费与电新联动。",
		"多角色继续维持观察结论。",
		"本地上下文理由",
		"stock-context-v1",
		[]string{"跌破关键均线"},
		[]string{"资讯窗口存在缺口"},
		[]string{},
		[]string{"本地上下文备注"},
	)
	record.ReportSnapshot["context_meta"] = map[string]any{
		"selected_trade_date": "2026-03-18",
		"price_source":        "TUSHARE",
		"news_window_days":    14,
	}

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(localAssetContextQueryPattern).
		WithArgs("stock-selection").
		WillReturnRows(newLocalAssetContextRows(buildStrategyEngineAssetContextFixture(record, record.AssetKeys)))
	mock.ExpectQuery(stockEventEvidenceQueryPattern).
		WithArgs("600519.SH", "600519.SH", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "event_type", "primary_symbol", "topic_label", "sector_label", "review_priority", "review_note", "published_at"}))

	explanation := repo.buildStockStrategyExplanation(
		model.StockRecommendation{
			ID:              "reco_context_001",
			Symbol:          "600519.SH",
			ValidFrom:       "2026-03-18T09:35:00Z",
			ReasonSummary:   "基础原因总结",
			StrategyVersion: "stock-context-v1",
		},
		model.StockRecommendationDetail{
			RecoID:   "reco_context_001",
			RiskNote: "基础风险提示",
		},
	)

	if !strings.Contains(explanation.SeedSummary, "2026-03-18 行情(TUSHARE)") {
		t.Fatalf("expected seed summary to include selected trade date and source, got %+v", explanation)
	}
	if !strings.Contains(explanation.SeedSummary, "近 14 天资讯") {
		t.Fatalf("expected seed summary to include news window, got %+v", explanation)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestListReviewedStockEventEvidenceIncludesEntitySymbolMatches(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(stockEventEvidenceQueryPattern).
		WithArgs("600519.SH", "600519.SH", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "event_type", "primary_symbol", "topic_label", "sector_label", "review_priority", "review_note", "published_at"}).
			AddRow("sec_demo_002", "白酒板块联动事件", "INDUSTRY_THEME", "000858.SZ", "白酒", "消费", "NORMAL", "与次级关联股票同步受益", time.Date(2026, 3, 17, 9, 30, 0, 0, time.UTC)))

	relatedEvents, eventCards, err := repo.listReviewedStockEventEvidence("600519.SH", "2026-03-18T09:35:00Z")
	if err != nil {
		t.Fatalf("list reviewed stock event evidence: %v", err)
	}
	if len(relatedEvents) != 1 {
		t.Fatalf("expected entity-matched related events, got %+v", relatedEvents)
	}
	if relatedEvents[0].PrimarySymbol != "000858.SZ" {
		t.Fatalf("expected event to survive when primary_symbol differs, got %+v", relatedEvents[0])
	}
	if len(eventCards) != 1 || eventCards[0].Value != "白酒板块联动事件" {
		t.Fatalf("expected event evidence cards for entity-matched symbol, got %+v", eventCards)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestBuildStockStrategyExplanationBackfillsFromRemotePublishRecordUsingLocalJobSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildExplanationPublishRecordDetail(
		"publish_remote_explanation_001",
		"job_remote_explanation_001",
		15,
		"2026-03-18",
		[]string{"600519.SH", "300750.SZ"},
		"远端发布图谱显示白马股分歧收敛。",
		"远端多角色已形成继续跟踪共识。",
		"远端回填理由",
		"stock-remote-v3",
		[]string{"跌破关键支撑位"},
		[]string{"远端回填告警"},
		[]string{"688981.SH"},
		[]string{"回填备注"},
	)
	remoteHistory := strategyEnginePublishHistoryResponse{
		Records: []strategyEnginePublishRecordSummary{
			{
				PublishID:     record.PublishID,
				JobID:         record.JobID,
				JobType:       record.JobType,
				Version:       record.Version,
				CreatedAt:     record.CreatedAt,
				TradeDate:     record.TradeDate,
				ReportSummary: record.ReportSummary,
				SelectedCount: record.SelectedCount,
				AssetKeys:     record.AssetKeys,
				PayloadCount:  record.PayloadCount,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/history/stock-selection":
			_ = json.NewEncoder(w).Encode(remoteHistory)
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/records/"+record.PublishID:
			_ = json.NewEncoder(w).Encode(record)
		case strings.HasPrefix(r.URL.Path, "/internal/v1/jobs/"):
			t.Fatalf("unexpected remote job fetch: %s", r.URL.Path)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	repo := &MySQLGrowthRepo{
		db: db,
		strategyEngine: &strategyEngineClient{
			baseURL:      server.URL,
			httpClient:   &http.Client{Timeout: 2 * time.Second},
			pollInterval: 10 * time.Millisecond,
		},
	}

	mock.ExpectQuery(localAssetContextQueryPattern).
		WithArgs("stock-selection").
		WillReturnRows(newLocalAssetContextRows())
	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newJobSnapshotRows(record))
	mock.ExpectExec(`INSERT INTO strategy_job_runs`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_artifacts`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_replays`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(publishRecordSnapshotQueryPattern).
		WithArgs(record.PublishID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newJobSnapshotRows(record))
	mock.ExpectQuery(stockEventEvidenceQueryPattern).
		WithArgs("600519.SH", "600519.SH", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "event_type", "primary_symbol", "topic_label", "sector_label", "review_priority", "review_note", "published_at"}))

	explanation := repo.buildStockStrategyExplanation(
		model.StockRecommendation{
			ID:              "reco_remote_001",
			Symbol:          "600519.SH",
			ValidFrom:       "2026-03-18T10:05:00Z",
			ReasonSummary:   "客户端当前推荐说明",
			StrategyVersion: "stock-live-v1",
		},
		model.StockRecommendationDetail{
			RecoID:   "reco_remote_001",
			RiskNote: "仓位保持克制",
		},
	)

	if explanation.PublishID != record.PublishID || explanation.JobID != record.JobID {
		t.Fatalf("expected remote publish record to backfill explanation, got %+v", explanation)
	}
	if explanation.PublishVersion != record.Version || explanation.TradeDate != record.TradeDate {
		t.Fatalf("unexpected publish metadata: %+v", explanation)
	}
	if explanation.GraphSummary != "远端发布图谱显示白马股分歧收敛。" || explanation.ConsensusSummary != "远端多角色已形成继续跟踪共识。" {
		t.Fatalf("expected remote report snapshot to hydrate explanation, got %+v", explanation)
	}
	if strings.Join(explanation.SeedHighlights, ",") != "600519.SH,300750.SZ" {
		t.Fatalf("expected local job snapshot to provide seed highlights, got %+v", explanation.SeedHighlights)
	}
	if strings.Join(explanation.RiskFlags, ",") != "远端回填告警,回填备注" {
		t.Fatalf("unexpected risk flags: %+v", explanation.RiskFlags)
	}
	if len(explanation.Simulations) != 1 || len(explanation.AgentOpinions) != 1 {
		t.Fatalf("expected simulations and agents from remote backfill, got %+v", explanation)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestBuildStockStrategyExplanationKeepsL1FallbackWithoutContext(t *testing.T) {
	repo := &MySQLGrowthRepo{}
	explanation := repo.buildStockStrategyExplanation(
		model.StockRecommendation{
			ID:              "reco_fallback_001",
			Symbol:          "600519.SH",
			ValidFrom:       "2026-03-28T09:30:00Z",
			ReasonSummary:   "趋势保持，等待确认",
			StrategyVersion: "stock-fallback-v1",
		},
		model.StockRecommendationDetail{
			RecoID:   "reco_fallback_001",
			RiskNote: "跌破前低减仓",
		},
	)

	if len(explanation.ResearchOutline) == 0 || len(explanation.ActiveThesisCards) == 0 {
		t.Fatalf("expected fallback stock explanation to keep l1 blocks: %+v", explanation)
	}
	if !explanation.ConfidenceCalibration.AdvisoryOnly || explanation.ConfidenceCalibration.AdjustedConfidence <= 0 {
		t.Fatalf("expected fallback stock explanation to keep advisory calibration: %+v", explanation.ConfidenceCalibration)
	}
}

func TestBuildFuturesStrategyExplanationKeepsL1FallbackWithoutContext(t *testing.T) {
	repo := &MySQLGrowthRepo{}
	explanation := repo.buildFuturesStrategyExplanation(
		model.FuturesStrategy{
			ID:            "futures_fallback_001",
			Contract:      "RB2505",
			ValidFrom:     "2026-03-28T09:30:00Z",
			ReasonSummary: "结构未破坏，继续观察",
		},
		model.FuturesGuidance{
			Contract:         "RB2505",
			RiskLevel:        "MEDIUM",
			InvalidCondition: "跌破 3220 则失效",
		},
	)

	if len(explanation.ResearchOutline) == 0 || len(explanation.ActiveThesisCards) == 0 {
		t.Fatalf("expected fallback futures explanation to keep l1 blocks: %+v", explanation)
	}
	if len(explanation.WatchSignals) == 0 {
		t.Fatalf("expected fallback futures explanation to keep watch signals: %+v", explanation)
	}
	if !explanation.ConfidenceCalibration.AdvisoryOnly || explanation.ConfidenceCalibration.AdjustedConfidence <= 0 {
		t.Fatalf("expected fallback futures explanation to keep advisory calibration: %+v", explanation.ConfidenceCalibration)
	}
}

func TestGetStockRecommendationVersionHistoryUsesBackfilledLocalContexts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildExplanationPublishRecordDetail(
		"publish_remote_history_001",
		"job_remote_history_001",
		16,
		"2026-03-18",
		[]string{"600519.SH", "300750.SZ"},
		"远端版本图谱显示消费与新能源重新分层。",
		"版本历史已经沉淀为可追踪共识。",
		"版本历史回填理由",
		"stock-remote-v4",
		[]string{"跌破前低后废止"},
		[]string{"版本历史告警"},
		[]string{"688981.SH"},
		[]string{"版本历史备注"},
	)
	remoteHistory := strategyEnginePublishHistoryResponse{
		Records: []strategyEnginePublishRecordSummary{
			{
				PublishID:     record.PublishID,
				JobID:         record.JobID,
				JobType:       record.JobType,
				Version:       record.Version,
				CreatedAt:     record.CreatedAt,
				TradeDate:     record.TradeDate,
				ReportSummary: record.ReportSummary,
				SelectedCount: record.SelectedCount,
				AssetKeys:     record.AssetKeys,
				PayloadCount:  record.PayloadCount,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/history/stock-selection":
			_ = json.NewEncoder(w).Encode(remoteHistory)
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/records/"+record.PublishID:
			_ = json.NewEncoder(w).Encode(record)
		case strings.HasPrefix(r.URL.Path, "/internal/v1/jobs/"):
			t.Fatalf("unexpected remote job fetch: %s", r.URL.Path)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	repo := &MySQLGrowthRepo{
		db: db,
		strategyEngine: &strategyEngineClient{
			baseURL:      server.URL,
			httpClient:   &http.Client{Timeout: 2 * time.Second},
			pollInterval: 10 * time.Millisecond,
		},
	}

	validFrom := time.Date(2026, 3, 18, 10, 30, 0, 0, time.UTC)
	mock.ExpectQuery(stockRecommendationVersionQueryPattern).
		WithArgs("reco_history_001").
		WillReturnRows(sqlmock.NewRows([]string{"id", "symbol", "valid_from", "reason_summary", "strategy_version"}).
			AddRow("reco_history_001", "600519.SH", validFrom, "客户端当前推荐说明", "stock-live-v1"))
	mock.ExpectQuery(stockRecommendationDetailQueryPattern).
		WithArgs("reco_history_001").
		WillReturnRows(sqlmock.NewRows([]string{
			"reco_id", "tech_score", "fund_score", "sentiment_score", "money_flow_score", "take_profit", "stop_loss", "risk_note",
		}).AddRow(
			"reco_history_001", 88.0, 90.0, 84.0, 86.0, "上涨 8% 分批止盈", "跌破支撑位止损", "仓位保持克制",
		))
	mock.ExpectQuery(localAssetContextQueryPattern).
		WithArgs("stock-selection").
		WillReturnRows(newLocalAssetContextRows())
	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newJobSnapshotRows(record))
	mock.ExpectExec(`INSERT INTO strategy_job_runs`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_artifacts`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_replays`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(publishRecordSnapshotQueryPattern).
		WithArgs(record.PublishID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newJobSnapshotRows(record))
	mock.ExpectQuery(localAssetContextQueryPattern).
		WithArgs("stock-selection").
		WillReturnRows(newLocalAssetContextRows(buildStrategyEngineAssetContextFixture(record, record.AssetKeys)))

	items, err := repo.GetStockRecommendationVersionHistory("user_history_001", "reco_history_001")
	if err != nil {
		t.Fatalf("GetStockRecommendationVersionHistory returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one version history item, got %+v", items)
	}
	if items[0].PublishID != record.PublishID || items[0].JobID != record.JobID {
		t.Fatalf("unexpected version history identity: %+v", items[0])
	}
	if items[0].PublishVersion != record.Version || items[0].TradeDate != record.TradeDate {
		t.Fatalf("unexpected version history metadata: %+v", items[0])
	}
	if items[0].StrategyVersion != "stock-remote-v4" || items[0].ReasonSummary != "版本历史回填理由" {
		t.Fatalf("expected local backfilled context to drive history fields, got %+v", items[0])
	}
	if !strings.Contains(items[0].ConfidenceReason, "版本历史回填理由") || items[0].ConsensusSummary != "版本历史已经沉淀为可追踪共识。" {
		t.Fatalf("unexpected version history explanation summary: %+v", items[0])
	}
	if !items[0].ConfidenceCalibration.AdvisoryOnly || items[0].ConfidenceCalibration.AdjustedConfidence <= 0 {
		t.Fatalf("expected stock version history confidence calibration, got %+v", items[0].ConfidenceCalibration)
	}
	if len(items[0].HistoricalThesisCards) == 0 {
		t.Fatalf("expected stock version history to keep historical thesis cards, got %+v", items[0])
	}
	if strings.Join(items[0].RiskFlags, ",") != "版本历史告警,版本历史备注" {
		t.Fatalf("unexpected version history risk flags: %+v", items[0].RiskFlags)
	}
	if strings.Join(items[0].Invalidations, ",") != "跌破前低后废止" {
		t.Fatalf("unexpected version history invalidations: %+v", items[0].Invalidations)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestApplyStrategyVersionDiffUsesStructuredPublishChanges(t *testing.T) {
	currentRecord := buildExplanationPublishRecordDetail(
		"publish_diff_current_001",
		"job_diff_current_001",
		12,
		"2026-03-22",
		[]string{"600519.SH", "300750.SZ"},
		"新版图谱摘要",
		"新版共识摘要",
		"新版理由",
		"stock-live-v12",
		[]string{"跌破五日线"},
		[]string{"新版告警"},
		[]string{},
		[]string{"新版备注"},
	)
	currentRecord.ReportSnapshot["candidates"] = []map[string]any{
		{
			"symbol":         "600519.SH",
			"rank":           1,
			"reason_summary": "新版理由",
			"risk_summary":   "控制节奏，仓位更克制",
		},
		{
			"symbol":         "300750.SZ",
			"rank":           2,
			"reason_summary": "新增标的",
			"risk_summary":   "波动加大需分批",
		},
	}

	previousRecord := buildExplanationPublishRecordDetail(
		"publish_diff_previous_001",
		"job_diff_previous_001",
		11,
		"2026-03-21",
		[]string{"600519.SH", "601318.SH"},
		"旧版图谱摘要",
		"旧版共识摘要",
		"旧版理由",
		"stock-live-v11",
		[]string{"跌破前低"},
		[]string{"旧版告警"},
		[]string{},
		[]string{"旧版备注"},
	)
	previousRecord.ReportSnapshot["candidates"] = []map[string]any{
		{
			"symbol":         "600519.SH",
			"rank":           2,
			"reason_summary": "旧版理由",
			"risk_summary":   "常规跟踪",
		},
		{
			"symbol":         "601318.SH",
			"rank":           1,
			"reason_summary": "移除标的",
			"risk_summary":   "旧版候选",
		},
	}

	currentCtx := buildStrategyEngineAssetContext(model.StrategyEnginePublishRecord{
		PublishID:      currentRecord.PublishID,
		JobID:          currentRecord.JobID,
		JobType:        currentRecord.JobType,
		Version:        currentRecord.Version,
		CreatedAt:      currentRecord.CreatedAt,
		TradeDate:      currentRecord.TradeDate,
		SelectedCount:  currentRecord.SelectedCount,
		PayloadCount:   currentRecord.PayloadCount,
		AssetKeys:      currentRecord.AssetKeys,
		ReportSnapshot: currentRecord.ReportSnapshot,
		Replay: model.StrategyEnginePublishReplay{
			WarningCount:      currentRecord.Replay.WarningCount,
			WarningMessages:   currentRecord.Replay.WarningMessages,
			VetoedAssets:      currentRecord.Replay.VetoedAssets,
			InvalidatedAssets: currentRecord.Replay.InvalidatedAssets,
			Notes:             currentRecord.Replay.Notes,
		},
	}, model.StrategyEngineJobRecord{}, "600519.SH")
	previousCtx := buildStrategyEngineAssetContext(model.StrategyEnginePublishRecord{
		PublishID:      previousRecord.PublishID,
		JobID:          previousRecord.JobID,
		JobType:        previousRecord.JobType,
		Version:        previousRecord.Version,
		CreatedAt:      previousRecord.CreatedAt,
		TradeDate:      previousRecord.TradeDate,
		SelectedCount:  previousRecord.SelectedCount,
		PayloadCount:   previousRecord.PayloadCount,
		AssetKeys:      previousRecord.AssetKeys,
		ReportSnapshot: previousRecord.ReportSnapshot,
		Replay: model.StrategyEnginePublishReplay{
			WarningCount:      previousRecord.Replay.WarningCount,
			WarningMessages:   previousRecord.Replay.WarningMessages,
			VetoedAssets:      previousRecord.Replay.VetoedAssets,
			InvalidatedAssets: previousRecord.Replay.InvalidatedAssets,
			Notes:             previousRecord.Replay.Notes,
		},
	}, model.StrategyEngineJobRecord{}, "600519.SH")
	if currentCtx == nil || previousCtx == nil {
		t.Fatalf("expected both strategy contexts to be built")
	}

	diff := buildStrategyVersionDiffFromContexts(*currentCtx, *previousCtx, "600519.SH")
	if diff.CurrentAssetChange != "PROMOTED" {
		t.Fatalf("expected promoted current asset change, got %+v", diff)
	}
	if strings.Join(diff.Added, ",") != "300750.SZ" {
		t.Fatalf("expected added symbol diff, got %+v", diff)
	}
	if strings.Join(diff.Removed, ",") != "601318.SH" {
		t.Fatalf("expected removed symbol diff, got %+v", diff)
	}
	if strings.Join(diff.Promoted, ",") != "600519.SH" {
		t.Fatalf("expected promoted symbol diff, got %+v", diff)
	}
	if len(diff.DowngradeReasons) == 0 || !strings.Contains(strings.Join(diff.DowngradeReasons, "；"), "600519.SH") {
		t.Fatalf("expected downgrade reasons to describe current asset updates, got %+v", diff)
	}

	explanation := model.StrategyClientExplanation{}
	applyStrategyVersionDiffToExplanation(&explanation, []strategyEngineAssetContext{*currentCtx, *previousCtx}, "600519.SH")
	if explanation.VersionDiff.CurrentAssetChange != "PROMOTED" {
		t.Fatalf("expected explanation version diff to be applied, got %+v", explanation.VersionDiff)
	}

	historyItems := []model.StrategyVersionHistoryItem{{PublishID: currentRecord.PublishID}, {PublishID: previousRecord.PublishID}}
	applyStrategyVersionDiffToHistoryItems(historyItems, []strategyEngineAssetContext{*currentCtx, *previousCtx}, "600519.SH")
	if historyItems[0].VersionDiff.CurrentAssetChange != "PROMOTED" {
		t.Fatalf("expected history item version diff to be applied, got %+v", historyItems[0].VersionDiff)
	}
}

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
			BaseConfidence:     0.72,
			AdjustedConfidence: 0.64,
			AdvisoryOnly:       true,
		},
	}

	item := buildFallbackVersionHistoryItem("", "", "2026-03-28", 0, "2026-03-28T09:00:00Z", "", "", explanation)
	if len(item.ResearchOutline) != 1 {
		t.Fatalf("expected research outline to survive fallback item: %+v", item)
	}
	if len(item.ActiveThesisCards) != 1 || len(item.HistoricalThesisCards) != 1 {
		t.Fatalf("expected thesis fields to survive fallback item: %+v", item)
	}
	if len(item.WatchSignals) != 1 || item.ConfidenceCalibration.AdjustedConfidence <= 0 {
		t.Fatalf("expected watch/confidence fields to survive fallback item: %+v", item)
	}
	if !item.ConfidenceCalibration.AdvisoryOnly {
		t.Fatalf("expected advisory_only to survive fallback item: %+v", item.ConfidenceCalibration)
	}
}

func TestBuildFallbackVersionHistoryItemCarriesL2Fields(t *testing.T) {
	explanation := model.StrategyClientExplanation{
		RelationshipSnapshot: model.StrategyExplanationRelationshipSnapshot{
			AssetKey:          "600519.SH",
			RelationshipCount: 3,
			Nodes: []model.StrategyExplanationRelationshipNode{
				{Type: "Theme", Label: "白酒"},
			},
		},
		ScenarioSnapshots: []model.StrategyExplanationScenarioSnapshot{
			{
				Scenario:           "base",
				Thesis:             "主逻辑延续",
				Trigger:            "放量突破",
				InvalidationSignal: "跌破 5 日线",
				ActionSuggestion:   "继续跟踪",
			},
		},
		AgentOpinions: []model.StrategyExplanationAgentOpinion{
			{
				Role:       "FLOW",
				Stance:     "SUPPORT",
				Confidence: 0.72,
				Summary:    "资金承接稳定",
			},
		},
	}

	item := buildFallbackVersionHistoryItem("", "", "2026-03-29", 0, "2026-03-29T09:00:00Z", "", "", explanation)
	if len(item.ScenarioSnapshots) != 1 || len(item.AgentOpinions) != 1 {
		t.Fatalf("expected L2 fields to survive fallback item: %+v", item)
	}
	if item.RelationshipSnapshot.RelationshipCount != 3 {
		t.Fatalf("expected relationship snapshot to survive fallback item: %+v", item.RelationshipSnapshot)
	}
}

func TestBuildFuturesStrategyExplanationUsesLocalSnapshotWithoutRemoteFetch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildFuturesExplanationPublishRecordDetail(
		"publish_local_futures_explanation_001",
		"job_local_futures_explanation_001",
		21,
		"2026-03-18",
		[]string{"IF2606", "IH2606"},
		"本地期货图谱显示指数多头结构仍在。",
		"多角色对方向与仓位已形成一致结论。",
		"本地期货回放理由",
		"futures-local-v2",
		[]string{"跌破 3490 则失效"},
		[]string{"期货本地告警"},
		[]string{},
		[]string{"期货本地备注"},
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected remote request: %s %s", r.Method, r.URL.Path)
	}))
	defer server.Close()

	repo := &MySQLGrowthRepo{
		db: db,
		strategyEngine: &strategyEngineClient{
			baseURL:      server.URL,
			httpClient:   &http.Client{Timeout: 2 * time.Second},
			pollInterval: 10 * time.Millisecond,
		},
	}

	mock.ExpectQuery(localAssetContextQueryPattern).
		WithArgs("futures-strategy").
		WillReturnRows(newLocalAssetContextRows(buildStrategyEngineAssetContextFixture(record, record.AssetKeys)))

	explanation := repo.buildFuturesStrategyExplanation(
		model.FuturesStrategy{
			ID:            "futures_local_001",
			Contract:      "IF2606",
			ValidFrom:     "2026-03-18T09:35:00Z",
			ReasonSummary: "基础期货理由",
		},
		model.FuturesGuidance{
			Contract:         "IF2606",
			RiskLevel:        "MEDIUM",
			InvalidCondition: "跌破 3490 则失效",
		},
	)

	if explanation.PublishID != record.PublishID || explanation.JobID != record.JobID {
		t.Fatalf("expected local futures snapshot identity, got %+v", explanation)
	}
	if explanation.PublishVersion != record.Version || explanation.TradeDate != record.TradeDate {
		t.Fatalf("unexpected local futures publish metadata: %+v", explanation)
	}
	if explanation.GraphSummary != "本地期货图谱显示指数多头结构仍在。" || explanation.ConsensusSummary != "多角色对方向与仓位已形成一致结论。" {
		t.Fatalf("expected futures explanation to use local report snapshot, got %+v", explanation)
	}
	if strings.Join(explanation.SeedHighlights, ",") != "IF2606,IH2606" {
		t.Fatalf("unexpected futures seed highlights: %+v", explanation.SeedHighlights)
	}
	if strings.Join(explanation.RiskFlags, ",") != "期货本地告警,期货本地备注" {
		t.Fatalf("unexpected futures risk flags: %+v", explanation.RiskFlags)
	}
	if strings.Join(explanation.Invalidations, ",") != "跌破 3490 则失效" {
		t.Fatalf("unexpected futures invalidations: %+v", explanation.Invalidations)
	}
	if len(explanation.Simulations) != 1 || len(explanation.AgentOpinions) != 1 {
		t.Fatalf("expected futures explanation simulation and agent opinions, got %+v", explanation)
	}
	if !explanation.ConfidenceCalibration.AdvisoryOnly || explanation.ConfidenceCalibration.AdjustedConfidence <= 0 {
		t.Fatalf("expected futures explanation confidence calibration, got %+v", explanation.ConfidenceCalibration)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestBuildFuturesStrategyExplanationSurfacesSupplyChainContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildFuturesExplanationPublishRecordDetail(
		"publish_local_futures_supply_chain_001",
		"job_local_futures_supply_chain_001",
		23,
		"2026-03-18",
		[]string{"RB2609", "HC2609"},
		"期货图谱显示黑色链条与交割资源联动增强。",
		"多角色认为库存去化与结构联动同时强化。",
		"黑色链条主线继续强化",
		"futures-supply-v1",
		[]string{"跌破 3220 则失效"},
		[]string{"库存口径存在一日延迟"},
		[]string{},
		[]string{"关注主仓迁移节奏"},
	)
	record.ReportSnapshot["related_entities"] = []map[string]any{
		{
			"entity_type":  "Commodity",
			"entity_key":   "RB",
			"label":        "螺纹钢",
			"asset_domain": "futures",
		},
	}
	strategies := record.ReportSnapshot["strategies"].([]map[string]any)
	strategies[0]["inventory_factor_summary"] = "央企品牌A库存占比提升；连续3日去库"
	strategies[0]["structure_factor_summary"] = "价差联动强化主方向；期限结构保持同向"
	strategies[0]["related_entities"] = []map[string]any{
		{
			"entity_type":  "SupplyChainNode",
			"entity_key":   "RB_SUPPLY_CHAIN:BLAST_FURNACE",
			"label":        "高炉链条",
			"asset_domain": "futures",
		},
		{
			"entity_type":  "Warehouse",
			"entity_key":   "WAREHOUSE:SH",
			"label":        "上海主仓",
			"asset_domain": "futures",
		},
		{
			"entity_type":  "Brand",
			"entity_key":   "BRAND:央企品牌A",
			"label":        "央企品牌A",
			"asset_domain": "futures",
		},
		{
			"entity_type":  "Grade",
			"entity_key":   "GRADE:一级品",
			"label":        "一级品",
			"asset_domain": "futures",
		},
		{
			"entity_type":  "SpreadPair",
			"entity_key":   "SPREAD:RB-HC",
			"label":        "RB-HC 价差对",
			"asset_domain": "futures",
		},
	}
	strategies[0]["evidence_cards"] = []map[string]any{
		{
			"title": "库存画像",
			"value": "库存深度 78.4",
			"note":  "央企品牌A库存占比提升；连续3日去库",
		},
		{
			"title": "结构联动",
			"value": "结构深度 74.2",
			"note":  "价差联动强化主方向；期限结构保持同向",
		},
	}

	repo := &MySQLGrowthRepo{db: db}
	mock.ExpectQuery(localAssetContextQueryPattern).
		WithArgs("futures-strategy").
		WillReturnRows(newLocalAssetContextRows(buildStrategyEngineAssetContextFixture(record, record.AssetKeys)))

	explanation := repo.buildFuturesStrategyExplanation(
		model.FuturesStrategy{
			ID:            "futures_supply_chain_001",
			Contract:      "RB2609",
			ValidFrom:     "2026-03-18T09:35:00Z",
			ReasonSummary: "基础期货理由",
		},
		model.FuturesGuidance{
			Contract:         "RB2609",
			RiskLevel:        "MEDIUM",
			InvalidCondition: "跌破 3220 则失效",
		},
	)

	payload, err := json.Marshal(explanation)
	if err != nil {
		t.Fatalf("marshal explanation: %v", err)
	}
	raw := map[string]any{}
	if err := json.Unmarshal(payload, &raw); err != nil {
		t.Fatalf("unmarshal explanation: %v", err)
	}

	if strings.TrimSpace(asString(raw["inventory_factor_summary"])) == "" {
		t.Fatalf("expected inventory_factor_summary to be populated, got %+v", raw)
	}
	if strings.TrimSpace(asString(raw["structure_factor_summary"])) == "" {
		t.Fatalf("expected structure_factor_summary to be populated, got %+v", raw)
	}
	supplyChainNotes := stringSlice(raw["supply_chain_notes"])
	if len(supplyChainNotes) < 3 {
		t.Fatalf("expected supply_chain_notes to surface supply-chain entity notes, got %+v", raw["supply_chain_notes"])
	}
	relatedEntities := sliceOfMaps(raw["related_entities"])
	if len(relatedEntities) < 5 {
		t.Fatalf("expected futures explanation to merge asset and report related entities, got %+v", raw["related_entities"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestBuildFuturesStrategyExplanationBackfillsFromRemotePublishRecordUsingLocalJobSnapshot(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildFuturesExplanationPublishRecordDetail(
		"publish_remote_futures_explanation_001",
		"job_remote_futures_explanation_001",
		22,
		"2026-03-18",
		[]string{"IF2606", "IH2606"},
		"远端期货图谱显示主力合约趋势继续扩展。",
		"远端期货多角色一致建议保留多头思路。",
		"远端期货回填理由",
		"futures-remote-v3",
		[]string{"跌破 3470 则废止"},
		[]string{"远端期货告警"},
		[]string{"IH2606"},
		[]string{"远端期货备注"},
	)
	remoteHistory := strategyEnginePublishHistoryResponse{
		Records: []strategyEnginePublishRecordSummary{
			{
				PublishID:     record.PublishID,
				JobID:         record.JobID,
				JobType:       record.JobType,
				Version:       record.Version,
				CreatedAt:     record.CreatedAt,
				TradeDate:     record.TradeDate,
				ReportSummary: record.ReportSummary,
				SelectedCount: record.SelectedCount,
				AssetKeys:     record.AssetKeys,
				PayloadCount:  record.PayloadCount,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/history/futures-strategy":
			_ = json.NewEncoder(w).Encode(remoteHistory)
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/records/"+record.PublishID:
			_ = json.NewEncoder(w).Encode(record)
		case strings.HasPrefix(r.URL.Path, "/internal/v1/jobs/"):
			t.Fatalf("unexpected remote futures job fetch: %s", r.URL.Path)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	repo := &MySQLGrowthRepo{
		db: db,
		strategyEngine: &strategyEngineClient{
			baseURL:      server.URL,
			httpClient:   &http.Client{Timeout: 2 * time.Second},
			pollInterval: 10 * time.Millisecond,
		},
	}

	mock.ExpectQuery(localAssetContextQueryPattern).
		WithArgs("futures-strategy").
		WillReturnRows(newLocalAssetContextRows())
	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newJobSnapshotRows(record))
	mock.ExpectExec(`INSERT INTO strategy_job_runs`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_artifacts`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_replays`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(publishRecordSnapshotQueryPattern).
		WithArgs(record.PublishID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newJobSnapshotRows(record))

	explanation := repo.buildFuturesStrategyExplanation(
		model.FuturesStrategy{
			ID:            "futures_remote_001",
			Contract:      "IF2606",
			ValidFrom:     "2026-03-18T10:05:00Z",
			ReasonSummary: "客户端当前期货说明",
		},
		model.FuturesGuidance{
			Contract:         "IF2606",
			RiskLevel:        "MEDIUM",
			InvalidCondition: "跌破 3470 则废止",
		},
	)

	if explanation.PublishID != record.PublishID || explanation.JobID != record.JobID {
		t.Fatalf("expected remote futures publish record to backfill explanation, got %+v", explanation)
	}
	if explanation.PublishVersion != record.Version || explanation.TradeDate != record.TradeDate {
		t.Fatalf("unexpected futures publish metadata: %+v", explanation)
	}
	if explanation.GraphSummary != "远端期货图谱显示主力合约趋势继续扩展。" || explanation.ConsensusSummary != "远端期货多角色一致建议保留多头思路。" {
		t.Fatalf("expected remote futures report snapshot to hydrate explanation, got %+v", explanation)
	}
	if strings.Join(explanation.SeedHighlights, ",") != "IF2606,IH2606" {
		t.Fatalf("expected local futures job snapshot to provide seed highlights, got %+v", explanation.SeedHighlights)
	}
	if strings.Join(explanation.RiskFlags, ",") != "远端期货告警,远端期货备注" {
		t.Fatalf("unexpected futures risk flags: %+v", explanation.RiskFlags)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestGetFuturesStrategyVersionHistoryUsesBackfilledLocalContexts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	record := buildFuturesExplanationPublishRecordDetail(
		"publish_remote_futures_history_001",
		"job_remote_futures_history_001",
		23,
		"2026-03-18",
		[]string{"IF2606", "IH2606"},
		"远端期货版本图谱显示多头结构和价差窗口同步改善。",
		"期货版本历史已沉淀为可追踪的方向共识。",
		"期货版本回填理由",
		"futures-remote-v4",
		[]string{"跌破前低则废止"},
		[]string{"期货版本告警"},
		[]string{"IH2606"},
		[]string{"期货版本备注"},
	)
	remoteHistory := strategyEnginePublishHistoryResponse{
		Records: []strategyEnginePublishRecordSummary{
			{
				PublishID:     record.PublishID,
				JobID:         record.JobID,
				JobType:       record.JobType,
				Version:       record.Version,
				CreatedAt:     record.CreatedAt,
				TradeDate:     record.TradeDate,
				ReportSummary: record.ReportSummary,
				SelectedCount: record.SelectedCount,
				AssetKeys:     record.AssetKeys,
				PayloadCount:  record.PayloadCount,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/history/futures-strategy":
			_ = json.NewEncoder(w).Encode(remoteHistory)
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/publish/records/"+record.PublishID:
			_ = json.NewEncoder(w).Encode(record)
		case strings.HasPrefix(r.URL.Path, "/internal/v1/jobs/"):
			t.Fatalf("unexpected remote futures job fetch: %s", r.URL.Path)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	repo := &MySQLGrowthRepo{
		db: db,
		strategyEngine: &strategyEngineClient{
			baseURL:      server.URL,
			httpClient:   &http.Client{Timeout: 2 * time.Second},
			pollInterval: 10 * time.Millisecond,
		},
	}

	validFrom := time.Date(2026, 3, 18, 10, 30, 0, 0, time.UTC)
	validTo := time.Date(2026, 3, 25, 15, 0, 0, 0, time.UTC)
	guidanceValidTo := time.Date(2026, 3, 25, 15, 0, 0, 0, time.UTC)

	mock.ExpectQuery(futuresStrategyDetailQueryPattern).
		WithArgs("futures_history_001").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "contract", "name", "direction", "risk_level", "position_range", "valid_from", "valid_to", "status", "reason_summary",
		}).AddRow(
			"futures_history_001", "IF2606", "股指主力", "LONG", "MEDIUM", "20%-30%", validFrom, validTo, "PUBLISHED", "客户端当前期货说明",
		))
	mock.ExpectQuery(futuresGuidanceQueryPattern).
		WithArgs("IF2606").
		WillReturnRows(sqlmock.NewRows([]string{
			"contract", "guidance_direction", "position_level", "entry_range", "take_profit_range", "stop_loss_range", "risk_level", "invalid_condition", "valid_to",
		}).AddRow(
			"IF2606", "LONG", "MEDIUM", "3520-3545", "3590-3620", "3470-3490", "MEDIUM", "跌破前低则废止", guidanceValidTo,
		))
	mock.ExpectQuery(localAssetContextQueryPattern).
		WithArgs("futures-strategy").
		WillReturnRows(newLocalAssetContextRows())
	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newJobSnapshotRows(record))
	mock.ExpectExec(`INSERT INTO strategy_job_runs`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_artifacts`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO strategy_job_replays`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(publishRecordSnapshotQueryPattern).
		WithArgs(record.PublishID).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(jobSnapshotQueryPattern).
		WithArgs(record.JobID).
		WillReturnRows(newJobSnapshotRows(record))
	mock.ExpectQuery(localAssetContextQueryPattern).
		WithArgs("futures-strategy").
		WillReturnRows(newLocalAssetContextRows(buildStrategyEngineAssetContextFixture(record, record.AssetKeys)))

	items, err := repo.GetFuturesStrategyVersionHistory("user_futures_history_001", "futures_history_001")
	if err != nil {
		t.Fatalf("GetFuturesStrategyVersionHistory returned error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one futures version history item, got %+v", items)
	}
	if items[0].PublishID != record.PublishID || items[0].JobID != record.JobID {
		t.Fatalf("unexpected futures version history identity: %+v", items[0])
	}
	if items[0].PublishVersion != record.Version || items[0].TradeDate != record.TradeDate {
		t.Fatalf("unexpected futures version history metadata: %+v", items[0])
	}
	if items[0].StrategyVersion != "futures-remote-v4" || items[0].ReasonSummary != "期货版本回填理由" {
		t.Fatalf("expected local futures backfilled context to drive history fields, got %+v", items[0])
	}
	if !strings.Contains(items[0].ConfidenceReason, "期货版本回填理由") || items[0].ConsensusSummary != "期货版本历史已沉淀为可追踪的方向共识。" {
		t.Fatalf("unexpected futures version history explanation summary: %+v", items[0])
	}
	if !items[0].ConfidenceCalibration.AdvisoryOnly || items[0].ConfidenceCalibration.AdjustedConfidence <= 0 {
		t.Fatalf("expected futures version history confidence calibration, got %+v", items[0].ConfidenceCalibration)
	}
	if len(items[0].ResearchOutline) == 0 || len(items[0].HistoricalThesisCards) == 0 {
		t.Fatalf("expected futures version history to keep l1 research fields, got %+v", items[0])
	}
	if strings.Join(items[0].RiskFlags, ",") != "期货版本告警,期货版本备注" {
		t.Fatalf("unexpected futures version history risk flags: %+v", items[0].RiskFlags)
	}
	if strings.Join(items[0].Invalidations, ",") != "跌破前低则废止" {
		t.Fatalf("unexpected futures version history invalidations: %+v", items[0].Invalidations)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func newLocalAssetContextRows(items ...strategyEngineAssetContextFixture) *sqlmock.Rows {
	rows := sqlmock.NewRows([]string{
		"job_id",
		"payload_snapshot",
		"remote_created_at",
		"trade_date",
		"report_snapshot",
		"publish_id",
		"publish_version",
		"replay_created_at",
		"replay_snapshot",
	})
	for _, item := range items {
		rows.AddRow(
			item.JobID,
			mustJSON(item.Payload),
			item.RemoteCreatedAt,
			item.TradeDate,
			mustJSON(item.Report),
			item.PublishID,
			item.PublishVersion,
			item.ReplayCreatedAt,
			mustJSON(item.Replay),
		)
	}
	return rows
}

func buildStrategyEngineAssetContextFixture(record strategyEnginePublishRecordDetail, seedSymbols []string) strategyEngineAssetContextFixture {
	if len(seedSymbols) == 0 {
		seedSymbols = append([]string{}, record.AssetKeys...)
	}
	payload := map[string]any{
		"trade_date": record.TradeDate,
	}
	if record.JobType == "futures-strategy" {
		payload["contracts"] = seedSymbols
	} else {
		payload["seed_symbols"] = seedSymbols
	}
	return strategyEngineAssetContextFixture{
		JobID:           record.JobID,
		Payload:         payload,
		RemoteCreatedAt: record.CreatedAt,
		TradeDate:       record.TradeDate,
		Report:          record.ReportSnapshot,
		PublishID:       record.PublishID,
		PublishVersion:  record.Version,
		ReplayCreatedAt: record.CreatedAt,
		Replay: map[string]any{
			"warning_count":      record.Replay.WarningCount,
			"warning_messages":   record.Replay.WarningMessages,
			"vetoed_assets":      record.Replay.VetoedAssets,
			"invalidated_assets": record.Replay.InvalidatedAssets,
			"notes":              record.Replay.Notes,
		},
	}
}

func buildExplanationPublishRecordDetail(
	publishID string,
	jobID string,
	version int,
	tradeDate string,
	assetKeys []string,
	graphSummary string,
	consensusSummary string,
	reasonSummary string,
	strategyVersion string,
	invalidations []string,
	warningMessages []string,
	vetoedAssets []string,
	notes []string,
) strategyEnginePublishRecordDetail {
	primaryAsset := firstAssetKey(assetKeys)
	publishPayloads := []map[string]any{
		{
			"recommendation": map[string]any{
				"symbol": primaryAsset,
			},
		},
	}
	reportSnapshot := map[string]any{
		"trade_date":        tradeDate,
		"generated_at":      tradeDate + "T09:35:00Z",
		"graph_summary":     graphSummary,
		"consensus_summary": consensusSummary,
		"selected_count":    1,
		"publish_payloads":  publishPayloads,
		"candidates": []map[string]any{
			{
				"symbol":           primaryAsset,
				"reason_summary":   reasonSummary,
				"strategy_version": strategyVersion,
				"invalidations":    invalidations,
			},
		},
		"simulations": []map[string]any{
			{
				"asset_key":        primaryAsset,
				"asset_type":       "stock",
				"consensus_action": "HOLD",
				"vetoed":           false,
				"veto_reason":      "",
				"scenarios": []map[string]any{
					{
						"scenario":         "基准情景",
						"thesis":           "高质量资产仍具备持续性",
						"score_adjustment": 1.3,
						"action":           "继续观察并分批参与",
						"risk_signal":      "量价配合稳定",
					},
				},
				"agents": []map[string]any{
					{
						"agent":      "risk",
						"stance":     "support",
						"confidence": 0.86,
						"summary":    "风险代理认为波动仍可控",
						"veto":       false,
					},
				},
			},
		},
	}
	return strategyEnginePublishRecordDetail{
		PublishID:       publishID,
		JobID:           jobID,
		JobType:         "stock-selection",
		Version:         version,
		CreatedAt:       tradeDate + "T09:35:00Z",
		TradeDate:       tradeDate,
		ReportSummary:   "生成 explanation 回填快照",
		SelectedCount:   1,
		AssetKeys:       assetKeys,
		PayloadCount:    len(publishPayloads),
		Markdown:        "# 发布报告",
		HTML:            "<p>发布报告</p>",
		PublishPayloads: publishPayloads,
		ReportSnapshot:  reportSnapshot,
		Replay: strategyEnginePublishReplay{
			WarningCount:      len(warningMessages),
			WarningMessages:   warningMessages,
			VetoedAssets:      vetoedAssets,
			InvalidatedAssets: invalidations,
			Notes:             notes,
		},
	}
}

func buildFuturesExplanationPublishRecordDetail(
	publishID string,
	jobID string,
	version int,
	tradeDate string,
	assetKeys []string,
	graphSummary string,
	consensusSummary string,
	reasonSummary string,
	strategyVersion string,
	invalidations []string,
	warningMessages []string,
	vetoedAssets []string,
	notes []string,
) strategyEnginePublishRecordDetail {
	primaryAsset := firstAssetKey(assetKeys)
	publishPayloads := []map[string]any{
		{
			"strategy": map[string]any{
				"contract":         primaryAsset,
				"reason_summary":   reasonSummary,
				"strategy_version": strategyVersion,
			},
			"guidance": map[string]any{
				"contract":           primaryAsset,
				"guidance_direction": "LONG",
			},
		},
	}
	reportSnapshot := map[string]any{
		"trade_date":        tradeDate,
		"generated_at":      tradeDate + "T09:35:00Z",
		"graph_summary":     graphSummary,
		"consensus_summary": consensusSummary,
		"selected_count":    1,
		"publish_payloads":  publishPayloads,
		"strategies": []map[string]any{
			{
				"contract":         primaryAsset,
				"reason_summary":   reasonSummary,
				"strategy_version": strategyVersion,
				"invalidations":    invalidations,
			},
		},
		"simulations": []map[string]any{
			{
				"asset_key":        primaryAsset,
				"asset_type":       "futures",
				"consensus_action": "LONG",
				"vetoed":           false,
				"veto_reason":      "",
				"scenarios": []map[string]any{
					{
						"scenario":         "趋势延续",
						"thesis":           "主力合约结构与量价方向保持一致",
						"score_adjustment": 1.1,
						"action":           "保留多头仓位",
						"risk_signal":      "回撤可控",
					},
				},
				"agents": []map[string]any{
					{
						"agent":      "macro",
						"stance":     "support",
						"confidence": 0.84,
						"summary":    "宏观代理认为方向继续偏多",
						"veto":       false,
					},
				},
			},
		},
	}
	return strategyEnginePublishRecordDetail{
		PublishID:       publishID,
		JobID:           jobID,
		JobType:         "futures-strategy",
		Version:         version,
		CreatedAt:       tradeDate + "T09:35:00Z",
		TradeDate:       tradeDate,
		ReportSummary:   "生成期货 explanation 回填快照",
		SelectedCount:   1,
		AssetKeys:       assetKeys,
		PayloadCount:    len(publishPayloads),
		Markdown:        "# 期货发布报告",
		HTML:            "<p>期货发布报告</p>",
		PublishPayloads: publishPayloads,
		ReportSnapshot:  reportSnapshot,
		Replay: strategyEnginePublishReplay{
			WarningCount:      len(warningMessages),
			WarningMessages:   warningMessages,
			VetoedAssets:      vetoedAssets,
			InvalidatedAssets: invalidations,
			Notes:             notes,
		},
	}
}
