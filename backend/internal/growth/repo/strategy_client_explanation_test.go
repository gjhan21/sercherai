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
	if items[0].ConfidenceReason != "版本历史回填理由" || items[0].ConsensusSummary != "版本历史已经沉淀为可追踪共识。" {
		t.Fatalf("unexpected version history explanation summary: %+v", items[0])
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
	if items[0].ConfidenceReason != "期货版本回填理由" || items[0].ConsensusSummary != "期货版本历史已沉淀为可追踪的方向共识。" {
		t.Fatalf("unexpected futures version history explanation summary: %+v", items[0])
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
