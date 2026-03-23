package repo

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/platform/config"
)

type strategyEngineClient struct {
	baseURL      string
	httpClient   *http.Client
	pollInterval time.Duration
}

type strategyEngineJobAccepted struct {
	JobID string `json:"job_id"`
}

type strategyEngineJobListResponse struct {
	Items    []strategyEngineAdminJobRecord `json:"items"`
	Total    int                            `json:"total"`
	Page     int                            `json:"page"`
	PageSize int                            `json:"page_size"`
}

type strategyEngineAdminJobRecord struct {
	JobID                string                        `json:"job_id"`
	JobType              string                        `json:"job_type"`
	Status               string                        `json:"status"`
	RequestedBy          string                        `json:"requested_by"`
	TraceID              string                        `json:"trace_id"`
	Payload              map[string]any                `json:"payload"`
	Result               *strategyEngineAdminJobResult `json:"result"`
	ErrorMessage         string                        `json:"error_message"`
	PublishCount         int                           `json:"publish_count"`
	LatestPublishID      string                        `json:"latest_publish_id"`
	LatestPublishVersion int                           `json:"latest_publish_version"`
	LatestPublishAt      string                        `json:"latest_publish_at"`
	LatestPublishMode    string                        `json:"latest_publish_mode"`
	LatestPublishSource  string                        `json:"latest_publish_source"`
	CreatedAt            string                        `json:"created_at"`
	StartedAt            string                        `json:"started_at"`
	FinishedAt           string                        `json:"finished_at"`
}

type strategyEngineAdminJobResult struct {
	Summary     string         `json:"summary"`
	PayloadEcho map[string]any `json:"payload_echo"`
	Artifacts   map[string]any `json:"artifacts"`
	Warnings    []string       `json:"warnings"`
}

type strategyEngineJobDetail struct {
	Status       string                     `json:"status"`
	ErrorMessage string                     `json:"error_message"`
	Result       strategyEngineJobResultRef `json:"result"`
}

type strategyEngineJobResultRef struct {
	Artifacts strategyEngineArtifacts `json:"artifacts"`
}

type strategyEngineArtifacts struct {
	Report strategyEngineStockSelectionReport `json:"report"`
}

type strategyEngineStockStageLog struct {
	StageKey        string         `json:"stage_key"`
	StageOrder      int            `json:"stage_order"`
	Status          string         `json:"status"`
	InputCount      int            `json:"input_count"`
	OutputCount     int            `json:"output_count"`
	DurationMS      int64          `json:"duration_ms"`
	DetailMessage   string         `json:"detail_message"`
	PayloadSnapshot map[string]any `json:"payload_snapshot"`
}

type strategyEngineStockCandidateSnapshot struct {
	Symbol              string         `json:"symbol"`
	Name                string         `json:"name"`
	Stage               string         `json:"stage"`
	QuantScore          float64        `json:"quant_score"`
	TotalScore          float64        `json:"total_score"`
	RiskLevel           string         `json:"risk_level"`
	Selected            bool           `json:"selected"`
	Rank                int            `json:"rank"`
	ReasonSummary       string         `json:"reason_summary"`
	EvidenceSummary     string         `json:"evidence_summary"`
	PortfolioRole       string         `json:"portfolio_role"`
	RiskSummary         string         `json:"risk_summary"`
	FactorBreakdownJSON map[string]any `json:"factor_breakdown_json"`
}

type strategyEngineStockPortfolioEntry struct {
	Symbol              string         `json:"symbol"`
	Name                string         `json:"name"`
	Rank                int            `json:"rank"`
	QuantScore          float64        `json:"quant_score"`
	TotalScore          float64        `json:"total_score"`
	RiskLevel           string         `json:"risk_level"`
	WeightSuggestion    string         `json:"weight_suggestion"`
	ReasonSummary       string         `json:"reason_summary"`
	EvidenceSummary     string         `json:"evidence_summary"`
	PortfolioRole       string         `json:"portfolio_role"`
	RiskSummary         string         `json:"risk_summary"`
	FactorBreakdownJSON map[string]any `json:"factor_breakdown_json"`
}

type strategyEngineStockEvidenceRecord struct {
	Symbol          string           `json:"symbol"`
	Name            string           `json:"name"`
	Stage           string           `json:"stage"`
	PortfolioRole   string           `json:"portfolio_role"`
	EvidenceSummary string           `json:"evidence_summary"`
	EvidenceCards   []map[string]any `json:"evidence_cards"`
	PositiveReasons []string         `json:"positive_reasons"`
	VetoReasons     []string         `json:"veto_reasons"`
	ThemeTags       []string         `json:"theme_tags"`
	SectorTags      []string         `json:"sector_tags"`
	RiskFlags       []string         `json:"risk_flags"`
}

type strategyEngineStockEvaluationRecord struct {
	Symbol          string  `json:"symbol"`
	Name            string  `json:"name"`
	HorizonDay      int     `json:"horizon_day"`
	EvaluationScope string  `json:"evaluation_scope"`
	EntryDate       string  `json:"entry_date"`
	ExitDate        string  `json:"exit_date"`
	EntryPrice      float64 `json:"entry_price"`
	ExitPrice       float64 `json:"exit_price"`
	ReturnPct       float64 `json:"return_pct"`
	ExcessReturnPct float64 `json:"excess_return_pct"`
	MaxDrawdownPct  float64 `json:"max_drawdown_pct"`
	HitFlag         bool    `json:"hit_flag"`
	BenchmarkSymbol string  `json:"benchmark_symbol"`
}

type strategyEngineStockSelectionReport struct {
	TradeDate          string                                 `json:"trade_date"`
	ReportSummary      string                                 `json:"report_summary"`
	RiskSummary        string                                 `json:"risk_summary"`
	SelectedCount      int                                    `json:"selected_count"`
	MarketRegime       string                                 `json:"market_regime"`
	GraphSummary       string                                 `json:"graph_summary"`
	GraphSnapshotID    string                                 `json:"graph_snapshot_id"`
	ContextMeta        map[string]any                         `json:"context_meta"`
	TemplateSnapshot   map[string]any                         `json:"template_snapshot"`
	EvaluationSummary  map[string]any                         `json:"evaluation_summary"`
	RelatedEntities    []map[string]any                       `json:"related_entities"`
	MemoryFeedback     map[string]any                         `json:"memory_feedback"`
	StageCounts        map[string]int                         `json:"stage_counts"`
	StageDurationsMS   map[string]int64                       `json:"stage_durations_ms"`
	StageLogs          []strategyEngineStockStageLog          `json:"stage_logs"`
	EvidenceRecords    []strategyEngineStockEvidenceRecord    `json:"evidence_records"`
	EvaluationRecords  []strategyEngineStockEvaluationRecord  `json:"evaluation_records"`
	CandidateSnapshots []strategyEngineStockCandidateSnapshot `json:"candidate_snapshots"`
	PortfolioEntries   []strategyEngineStockPortfolioEntry    `json:"portfolio_entries"`
	PublishPayload     []strategyEngineStockPublishPayload    `json:"publish_payloads"`
	PublishID          string                                 `json:"-"`
	PublishVersion     int                                    `json:"-"`
	JobRecord          model.StrategyEngineJobRecord          `json:"-"`
}

type strategyEnginePublishRecord struct {
	PublishID     string `json:"publish_id"`
	Version       int    `json:"version"`
	ReportSummary string `json:"report_summary"`
	SelectedCount int    `json:"selected_count"`
	PayloadCount  int    `json:"payload_count"`
	TradeDate     string `json:"trade_date"`
}

type strategyEnginePublishHistoryResponse struct {
	Records []strategyEnginePublishRecordSummary `json:"records"`
}

type strategyEnginePublishRecordSummary struct {
	PublishID     string   `json:"publish_id"`
	JobID         string   `json:"job_id"`
	JobType       string   `json:"job_type"`
	Version       int      `json:"version"`
	CreatedAt     string   `json:"created_at"`
	TradeDate     string   `json:"trade_date"`
	ReportSummary string   `json:"report_summary"`
	SelectedCount int      `json:"selected_count"`
	AssetKeys     []string `json:"asset_keys"`
	PayloadCount  int      `json:"payload_count"`
}

type strategyEnginePublishRecordDetail struct {
	PublishID       string                      `json:"publish_id"`
	JobID           string                      `json:"job_id"`
	JobType         string                      `json:"job_type"`
	Version         int                         `json:"version"`
	CreatedAt       string                      `json:"created_at"`
	TradeDate       string                      `json:"trade_date"`
	ReportSummary   string                      `json:"report_summary"`
	SelectedCount   int                         `json:"selected_count"`
	AssetKeys       []string                    `json:"asset_keys"`
	PayloadCount    int                         `json:"payload_count"`
	Markdown        string                      `json:"markdown"`
	HTML            string                      `json:"html"`
	PublishPayloads []map[string]any            `json:"publish_payloads"`
	ReportSnapshot  map[string]any              `json:"report_snapshot"`
	Replay          strategyEnginePublishReplay `json:"replay"`
}

type strategyEnginePublishReplay struct {
	WarningCount      int      `json:"warning_count"`
	WarningMessages   []string `json:"warning_messages"`
	VetoedAssets      []string `json:"vetoed_assets"`
	InvalidatedAssets []string `json:"invalidated_assets"`
	Notes             []string `json:"notes"`
}

type strategyEnginePublishCompareResult struct {
	LeftPublishID      string   `json:"left_publish_id"`
	RightPublishID     string   `json:"right_publish_id"`
	LeftVersion        int      `json:"left_version"`
	RightVersion       int      `json:"right_version"`
	SelectedCountDelta int      `json:"selected_count_delta"`
	PayloadCountDelta  int      `json:"payload_count_delta"`
	WarningCountDelta  int      `json:"warning_count_delta"`
	VetoCountDelta     int      `json:"veto_count_delta"`
	AddedAssets        []string `json:"added_assets"`
	RemovedAssets      []string `json:"removed_assets"`
	SharedAssets       []string `json:"shared_assets"`
}

type strategyEngineStockPublishPayload struct {
	Recommendation strategyEngineRecommendation `json:"recommendation"`
	Detail         strategyEngineDetail         `json:"detail"`
}

type strategyEngineFuturesJobDetail struct {
	Status       string                         `json:"status"`
	ErrorMessage string                         `json:"error_message"`
	Result       strategyEngineFuturesResultRef `json:"result"`
}

type strategyEngineFuturesResultRef struct {
	Artifacts strategyEngineFuturesArtifacts `json:"artifacts"`
}

type strategyEngineFuturesArtifacts struct {
	Report strategyEngineFuturesStrategyReport `json:"report"`
}

type strategyEngineFuturesStrategyReport struct {
	TradeDate          string                                   `json:"trade_date"`
	SelectedCount      int                                      `json:"selected_count"`
	MarketRegime       string                                   `json:"market_regime"`
	GraphSummary       string                                   `json:"graph_summary"`
	GraphSnapshotID    string                                   `json:"graph_snapshot_id"`
	ReportSummary      string                                   `json:"report_summary"`
	ContextMeta        map[string]any                           `json:"context_meta"`
	TemplateSnapshot   map[string]any                           `json:"template_snapshot"`
	EvaluationSummary  map[string]any                           `json:"evaluation_summary"`
	RelatedEntities    []map[string]any                         `json:"related_entities"`
	MemoryFeedback     map[string]any                           `json:"memory_feedback"`
	StageCounts        map[string]int                           `json:"stage_counts"`
	StageDurationsMS   map[string]int64                         `json:"stage_durations_ms"`
	StageLogs          []strategyEngineStockStageLog            `json:"stage_logs"`
	EvidenceRecords    []strategyEngineFuturesEvidenceRecord    `json:"evidence_records"`
	EvaluationRecords  []strategyEngineFuturesEvaluationRecord  `json:"evaluation_records"`
	CandidateSnapshots []strategyEngineFuturesCandidateSnapshot `json:"candidate_snapshots"`
	PortfolioEntries   []strategyEngineFuturesPortfolioEntry    `json:"portfolio_entries"`
	PublishPayload     []strategyEngineFuturesPublishPayload    `json:"publish_payloads"`
	PublishID          string                                   `json:"-"`
	PublishVersion     int                                      `json:"-"`
	JobRecord          model.StrategyEngineJobRecord            `json:"-"`
}

type strategyEngineFuturesCandidateSnapshot struct {
	Contract            string         `json:"contract"`
	Name                string         `json:"name"`
	Stage               string         `json:"stage"`
	Score               float64        `json:"score"`
	Direction           string         `json:"direction"`
	RiskLevel           string         `json:"risk_level"`
	Selected            bool           `json:"selected"`
	Rank                int            `json:"rank"`
	ReasonSummary       string         `json:"reason_summary"`
	EvidenceSummary     string         `json:"evidence_summary"`
	PortfolioRole       string         `json:"portfolio_role"`
	RiskSummary         string         `json:"risk_summary"`
	FactorBreakdownJSON map[string]any `json:"factor_breakdown_json"`
}

type strategyEngineFuturesPortfolioEntry struct {
	Contract            string         `json:"contract"`
	Name                string         `json:"name"`
	Rank                int            `json:"rank"`
	Score               float64        `json:"score"`
	Direction           string         `json:"direction"`
	RiskLevel           string         `json:"risk_level"`
	PositionRange       string         `json:"position_range"`
	ReasonSummary       string         `json:"reason_summary"`
	EvidenceSummary     string         `json:"evidence_summary"`
	PortfolioRole       string         `json:"portfolio_role"`
	RiskSummary         string         `json:"risk_summary"`
	FactorBreakdownJSON map[string]any `json:"factor_breakdown_json"`
}

type strategyEngineFuturesEvidenceRecord struct {
	Contract        string           `json:"contract"`
	Name            string           `json:"name"`
	Stage           string           `json:"stage"`
	PortfolioRole   string           `json:"portfolio_role"`
	EvidenceSummary string           `json:"evidence_summary"`
	EvidenceCards   []map[string]any `json:"evidence_cards"`
	PositiveReasons []string         `json:"positive_reasons"`
	VetoReasons     []string         `json:"veto_reasons"`
	RiskFlags       []string         `json:"risk_flags"`
	RelatedEntities []map[string]any `json:"related_entities"`
}

type strategyEngineFuturesEvaluationRecord struct {
	Contract        string  `json:"contract"`
	Name            string  `json:"name"`
	HorizonDay      int     `json:"horizon_day"`
	EvaluationScope string  `json:"evaluation_scope"`
	EntryDate       string  `json:"entry_date"`
	ExitDate        string  `json:"exit_date"`
	EntryPrice      float64 `json:"entry_price"`
	ExitPrice       float64 `json:"exit_price"`
	ReturnPct       float64 `json:"return_pct"`
	ExcessReturnPct float64 `json:"excess_return_pct"`
	MaxDrawdownPct  float64 `json:"max_drawdown_pct"`
	HitFlag         bool    `json:"hit_flag"`
	BenchmarkSymbol string  `json:"benchmark_symbol"`
}

type strategyEngineFuturesPublishPayload struct {
	Strategy strategyEngineFuturesStrategy `json:"strategy"`
	Guidance strategyEngineFuturesGuidance `json:"guidance"`
}

type strategyEngineFuturesStrategy struct {
	Contract      string `json:"contract"`
	Name          string `json:"name"`
	Direction     string `json:"direction"`
	RiskLevel     string `json:"risk_level"`
	PositionRange string `json:"position_range"`
	ValidFrom     string `json:"valid_from"`
	ValidTo       string `json:"valid_to"`
	Status        string `json:"status"`
	ReasonSummary string `json:"reason_summary"`
}

type strategyEngineFuturesGuidance struct {
	Contract          string `json:"contract"`
	GuidanceDirection string `json:"guidance_direction"`
	PositionLevel     string `json:"position_level"`
	EntryRange        string `json:"entry_range"`
	TakeProfitRange   string `json:"take_profit_range"`
	StopLossRange     string `json:"stop_loss_range"`
	RiskLevel         string `json:"risk_level"`
	InvalidCondition  string `json:"invalid_condition"`
	ValidTo           string `json:"valid_to"`
}

type strategyEngineRecommendation struct {
	Symbol           string  `json:"symbol"`
	Name             string  `json:"name"`
	Score            float64 `json:"score"`
	RiskLevel        string  `json:"risk_level"`
	PositionRange    string  `json:"position_range"`
	ValidFrom        string  `json:"valid_from"`
	ValidTo          string  `json:"valid_to"`
	Status           string  `json:"status"`
	ReasonSummary    string  `json:"reason_summary"`
	SourceType       string  `json:"source_type"`
	StrategyVersion  string  `json:"strategy_version"`
	Reviewer         string  `json:"reviewer"`
	Publisher        string  `json:"publisher"`
	ReviewNote       string  `json:"review_note"`
	PerformanceLabel string  `json:"performance_label"`
}

type strategyEngineDetail struct {
	TechScore      float64 `json:"tech_score"`
	FundScore      float64 `json:"fund_score"`
	SentimentScore float64 `json:"sentiment_score"`
	MoneyFlowScore float64 `json:"money_flow_score"`
	TakeProfit     string  `json:"take_profit"`
	StopLoss       string  `json:"stop_loss"`
	RiskNote       string  `json:"risk_note"`
}

const (
	strategyEngineStockSelectionModeAuto   = model.StrategyEngineStockSelectionModeAuto
	strategyEngineStockSelectionModeManual = model.StrategyEngineStockSelectionModeManual

	strategyEngineDefaultStockSelectionProfileID = model.StrategyEngineDefaultStockSelectionProfileID
	strategyEngineDefaultStockUniverseScope      = model.StrategyEngineDefaultStockUniverseScope
	strategyEngineDefaultStockMinListingDays     = model.StrategyEngineDefaultStockMinListingDays
	strategyEngineDefaultStockMinAvgTurnover     = model.StrategyEngineDefaultStockMinAvgTurnover
)

func newStrategyEngineClient(cfg config.Config) *strategyEngineClient {
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.StrategyEngineBaseURL), "/")
	if baseURL == "" {
		return nil
	}

	timeoutMS := cfg.StrategyEngineTimeoutMS
	if timeoutMS <= 0 {
		timeoutMS = 15000
	}
	pollMS := cfg.StrategyEnginePollMS
	if pollMS <= 0 {
		pollMS = 250
	}

	return &strategyEngineClient{
		baseURL:      baseURL,
		httpClient:   &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond},
		pollInterval: time.Duration(pollMS) * time.Millisecond,
	}
}

func (r *MySQLGrowthRepo) generateDailyStockRecommendationsViaStrategyEngine(tradeDate string) (model.AdminDailyStockRecommendationGenerationResult, error) {
	seedSet, err := r.ResolveActiveStrategySeedSet("STOCK")
	if err != nil {
		return model.AdminDailyStockRecommendationGenerationResult{}, err
	}
	profile, err := r.ResolveActiveStrategyAgentProfile("STOCK")
	if err != nil {
		return model.AdminDailyStockRecommendationGenerationResult{}, err
	}
	scenarioTemplate, err := r.ResolveActiveStrategyScenarioTemplate("STOCK")
	if err != nil {
		return model.AdminDailyStockRecommendationGenerationResult{}, err
	}
	publishPolicy, err := r.ResolveActiveStrategyPublishPolicy("STOCK")
	if err != nil {
		return model.AdminDailyStockRecommendationGenerationResult{}, err
	}
	report, err := r.strategyEngine.generateStockSelectionReport(tradeDate, seedSet, profile, scenarioTemplate, publishPolicy)
	if err != nil {
		return model.AdminDailyStockRecommendationGenerationResult{}, err
	}
	if err := r.upsertStrategyEngineJobSnapshot(report.JobRecord); err != nil {
		return model.AdminDailyStockRecommendationGenerationResult{}, err
	}
	count, err := r.persistStrategyEngineStockRecommendations(report.PublishPayload)
	if err != nil {
		return model.AdminDailyStockRecommendationGenerationResult{}, err
	}
	return model.AdminDailyStockRecommendationGenerationResult{
		Count:          count,
		PublishID:      report.PublishID,
		PublishVersion: report.PublishVersion,
		ReportSummary:  report.ReportSummary,
		GenerationMode: "STRATEGY_ENGINE",
		ArchiveEnabled: true,
	}, nil
}

func (c *strategyEngineClient) generateStockSelectionReport(
	tradeDate string,
	seedSet *model.StrategySeedSet,
	profile *model.StrategyAgentProfile,
	scenarioTemplate *model.StrategyScenarioTemplate,
	publishPolicy *model.StrategyPublishPolicy,
) (strategyEngineStockSelectionReport, error) {
	payload := buildStockSelectionJobPayload(tradeDate, seedSet)
	body := map[string]any{
		"requested_by": "go-backend",
		"payload":      payload,
	}
	attachStrategyConfigPayload(payload, seedSet, profile, scenarioTemplate, publishPolicy)
	if profile != nil {
		payload["enabled_agents"] = profile.EnabledAgents
		payload["positive_threshold"] = profile.PositiveThreshold
		payload["negative_threshold"] = profile.NegativeThreshold
		payload["allow_veto"] = profile.AllowVeto
	}
	if scenarioTemplate != nil && len(scenarioTemplate.Items) > 0 {
		payload["scenario_templates"] = scenarioTemplate.Items
	}
	if strings.TrimSpace(tradeDate) == "" {
		delete(payload, "trade_date")
	}

	job, err := c.createStockSelectionJob(body)
	if err != nil {
		return strategyEngineStockSelectionReport{}, err
	}
	report, err := c.waitForStockSelectionJob(job.JobID)
	if err != nil {
		return strategyEngineStockSelectionReport{}, err
	}
	publishRecord, err := c.publishJob(job.JobID, "go-backend", false, "", publishPolicy)
	if err != nil {
		return strategyEngineStockSelectionReport{}, err
	}
	jobRecord, err := c.getJobRecord(job.JobID)
	if err != nil {
		return strategyEngineStockSelectionReport{}, err
	}
	report.PublishID = publishRecord.PublishID
	report.PublishVersion = publishRecord.Version
	report.ReportSummary = publishRecord.ReportSummary
	report.JobRecord = jobRecord
	return report, nil
}

func buildStockSelectionJobPayload(tradeDate string, seedSet *model.StrategySeedSet) map[string]any {
	payload := map[string]any{
		"trade_date":       tradeDate,
		"selection_mode":   strategyEngineStockSelectionModeAuto,
		"profile_id":       strategyEngineDefaultStockSelectionProfileID,
		"universe_scope":   strategyEngineDefaultStockUniverseScope,
		"market_scope":     strategyEngineDefaultStockUniverseScope,
		"min_listing_days": strategyEngineDefaultStockMinListingDays,
		"min_avg_turnover": strategyEngineDefaultStockMinAvgTurnover,
		"limit":            10,
		"max_risk_level":   "MEDIUM",
		"min_score":        75,
	}
	if seedSet != nil && len(seedSet.Items) > 0 {
		payload["selection_mode"] = strategyEngineStockSelectionModeManual
		payload["seed_symbols"] = seedSet.Items
	}
	return payload
}

func (r *MySQLGrowthRepo) generateDailyFuturesStrategiesViaStrategyEngine(tradeDate string) (model.AdminDailyFuturesStrategyGenerationResult, error) {
	seedSet, err := r.ResolveActiveStrategySeedSet("FUTURES")
	if err != nil {
		return model.AdminDailyFuturesStrategyGenerationResult{}, err
	}
	profile, err := r.ResolveActiveStrategyAgentProfile("FUTURES")
	if err != nil {
		return model.AdminDailyFuturesStrategyGenerationResult{}, err
	}
	scenarioTemplate, err := r.ResolveActiveStrategyScenarioTemplate("FUTURES")
	if err != nil {
		return model.AdminDailyFuturesStrategyGenerationResult{}, err
	}
	publishPolicy, err := r.ResolveActiveStrategyPublishPolicy("FUTURES")
	if err != nil {
		return model.AdminDailyFuturesStrategyGenerationResult{}, err
	}
	report, err := r.strategyEngine.generateFuturesStrategyReport(tradeDate, seedSet, profile, scenarioTemplate, publishPolicy)
	if err != nil {
		return model.AdminDailyFuturesStrategyGenerationResult{}, err
	}
	if err := r.upsertStrategyEngineJobSnapshot(report.JobRecord); err != nil {
		return model.AdminDailyFuturesStrategyGenerationResult{}, err
	}
	count, err := r.persistStrategyEngineFuturesStrategies(report.PublishPayload)
	if err != nil {
		return model.AdminDailyFuturesStrategyGenerationResult{}, err
	}
	return model.AdminDailyFuturesStrategyGenerationResult{
		Count:          count,
		PublishID:      report.PublishID,
		PublishVersion: report.PublishVersion,
		ReportSummary:  report.ReportSummary,
		GenerationMode: "STRATEGY_ENGINE",
		ArchiveEnabled: true,
	}, nil
}

func (c *strategyEngineClient) generateFuturesStrategyReport(
	tradeDate string,
	seedSet *model.StrategySeedSet,
	profile *model.StrategyAgentProfile,
	scenarioTemplate *model.StrategyScenarioTemplate,
	publishPolicy *model.StrategyPublishPolicy,
) (strategyEngineFuturesStrategyReport, error) {
	allowMockFallbackOnShortHistory := true
	if profile != nil {
		allowMockFallbackOnShortHistory = profile.AllowMockFallbackOnShortHistory
	}
	body := map[string]any{
		"requested_by": "go-backend",
		"payload": map[string]any{
			"trade_date":                           tradeDate,
			"limit":                                5,
			"max_risk_level":                       "HIGH",
			"allow_mock_fallback_on_short_history": allowMockFallbackOnShortHistory,
		},
	}
	attachStrategyConfigPayload(body["payload"].(map[string]any), seedSet, profile, scenarioTemplate, publishPolicy)
	if seedSet != nil && len(seedSet.Items) > 0 {
		body["payload"].(map[string]any)["contracts"] = seedSet.Items
	}
	if profile != nil {
		body["payload"].(map[string]any)["enabled_agents"] = profile.EnabledAgents
		body["payload"].(map[string]any)["positive_threshold"] = profile.PositiveThreshold
		body["payload"].(map[string]any)["negative_threshold"] = profile.NegativeThreshold
		body["payload"].(map[string]any)["allow_veto"] = profile.AllowVeto
	}
	if scenarioTemplate != nil && len(scenarioTemplate.Items) > 0 {
		body["payload"].(map[string]any)["scenario_templates"] = scenarioTemplate.Items
	}
	if strings.TrimSpace(tradeDate) == "" {
		delete(body["payload"].(map[string]any), "trade_date")
	}

	job, err := c.createFuturesStrategyJob(body)
	if err != nil {
		return strategyEngineFuturesStrategyReport{}, err
	}
	report, err := c.waitForFuturesStrategyJob(job.JobID)
	if err != nil {
		return strategyEngineFuturesStrategyReport{}, err
	}
	publishRecord, err := c.publishJob(job.JobID, "go-backend", false, "", publishPolicy)
	if err != nil {
		return strategyEngineFuturesStrategyReport{}, err
	}
	jobRecord, err := c.getJobRecord(job.JobID)
	if err != nil {
		return strategyEngineFuturesStrategyReport{}, err
	}
	report.PublishID = publishRecord.PublishID
	report.PublishVersion = publishRecord.Version
	report.ReportSummary = publishRecord.ReportSummary
	report.JobRecord = jobRecord
	return report, nil
}

func (c *strategyEngineClient) createStockSelectionJob(payload map[string]any) (strategyEngineJobAccepted, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return strategyEngineJobAccepted{}, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/internal/v1/jobs/stock-selection", bytes.NewReader(body))
	if err != nil {
		return strategyEngineJobAccepted{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return strategyEngineJobAccepted{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return strategyEngineJobAccepted{}, fmt.Errorf("strategy-engine returned %d when creating stock job: %s", resp.StatusCode, readBodyText(resp.Body))
	}
	var result strategyEngineJobAccepted
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return strategyEngineJobAccepted{}, err
	}
	if strings.TrimSpace(result.JobID) == "" {
		return strategyEngineJobAccepted{}, fmt.Errorf("strategy-engine returned empty job_id")
	}
	return result, nil
}

func attachStrategyConfigPayload(
	payload map[string]any,
	seedSet *model.StrategySeedSet,
	profile *model.StrategyAgentProfile,
	scenarioTemplate *model.StrategyScenarioTemplate,
	publishPolicy *model.StrategyPublishPolicy,
) {
	if payload == nil {
		return
	}
	payload["config_refs"] = map[string]any{
		"seed_set":          buildStrategySeedSetRef(seedSet),
		"agent_profile":     buildStrategyAgentProfileRef(profile),
		"scenario_template": buildStrategyScenarioTemplateRef(scenarioTemplate),
		"publish_policy":    buildStrategyPublishPolicyRef(publishPolicy),
	}
	payload["publish_policy_preview"] = buildStrategyPublishPolicyPreview(publishPolicy)
}

func buildStrategySeedSetRef(item *model.StrategySeedSet) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":          item.ID,
		"name":        item.Name,
		"target_type": item.TargetType,
		"is_default":  item.IsDefault,
		"updated_at":  item.UpdatedAt,
		"source":      "admin-default",
	}
}

func buildStrategyAgentProfileRef(item *model.StrategyAgentProfile) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":          item.ID,
		"name":        item.Name,
		"target_type": item.TargetType,
		"is_default":  item.IsDefault,
		"updated_at":  item.UpdatedAt,
		"source":      "admin-default",
	}
}

func buildStrategyScenarioTemplateRef(item *model.StrategyScenarioTemplate) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":          item.ID,
		"name":        item.Name,
		"target_type": item.TargetType,
		"is_default":  item.IsDefault,
		"updated_at":  item.UpdatedAt,
		"source":      "admin-default",
	}
}

func buildStrategyPublishPolicyRef(item *model.StrategyPublishPolicy) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"id":          item.ID,
		"name":        item.Name,
		"target_type": item.TargetType,
		"is_default":  item.IsDefault,
		"updated_at":  item.UpdatedAt,
		"source":      "admin-default",
	}
}

func buildStrategyPublishPolicyPreview(item *model.StrategyPublishPolicy) map[string]any {
	if item == nil {
		return nil
	}
	return map[string]any{
		"max_risk_level":         item.MaxRiskLevel,
		"max_warning_count":      item.MaxWarningCount,
		"allow_vetoed_publish":   item.AllowVetoedPublish,
		"default_publisher":      item.DefaultPublisher,
		"override_note_template": item.OverrideNoteTemplate,
	}
}

func (c *strategyEngineClient) createFuturesStrategyJob(payload map[string]any) (strategyEngineJobAccepted, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return strategyEngineJobAccepted{}, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/internal/v1/jobs/futures-strategy", bytes.NewReader(body))
	if err != nil {
		return strategyEngineJobAccepted{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return strategyEngineJobAccepted{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		return strategyEngineJobAccepted{}, fmt.Errorf("strategy-engine returned %d when creating futures job: %s", resp.StatusCode, readBodyText(resp.Body))
	}
	var result strategyEngineJobAccepted
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return strategyEngineJobAccepted{}, err
	}
	if strings.TrimSpace(result.JobID) == "" {
		return strategyEngineJobAccepted{}, fmt.Errorf("strategy-engine returned empty job_id")
	}
	return result, nil
}

func (c *strategyEngineClient) waitForStockSelectionJob(jobID string) (strategyEngineStockSelectionReport, error) {
	deadline := time.Now().Add(c.httpClient.Timeout)
	for {
		detail, err := c.getJob(jobID)
		if err != nil {
			return strategyEngineStockSelectionReport{}, err
		}
		switch strings.ToUpper(strings.TrimSpace(detail.Status)) {
		case "SUCCEEDED":
			if len(detail.Result.Artifacts.Report.PublishPayload) == 0 {
				return strategyEngineStockSelectionReport{}, fmt.Errorf("strategy-engine succeeded but returned empty publish payload")
			}
			return detail.Result.Artifacts.Report, nil
		case "FAILED":
			if strings.TrimSpace(detail.ErrorMessage) == "" {
				return strategyEngineStockSelectionReport{}, fmt.Errorf("strategy-engine stock job failed")
			}
			return strategyEngineStockSelectionReport{}, fmt.Errorf("strategy-engine stock job failed: %s", detail.ErrorMessage)
		}

		if time.Now().After(deadline) {
			return strategyEngineStockSelectionReport{}, fmt.Errorf("strategy-engine stock job timed out after %s", c.httpClient.Timeout)
		}
		time.Sleep(c.pollInterval)
	}
}

func (c *strategyEngineClient) waitForFuturesStrategyJob(jobID string) (strategyEngineFuturesStrategyReport, error) {
	deadline := time.Now().Add(c.httpClient.Timeout)
	for {
		detail, err := c.getFuturesJob(jobID)
		if err != nil {
			return strategyEngineFuturesStrategyReport{}, err
		}
		switch strings.ToUpper(strings.TrimSpace(detail.Status)) {
		case "SUCCEEDED":
			if len(detail.Result.Artifacts.Report.PublishPayload) == 0 {
				return strategyEngineFuturesStrategyReport{}, fmt.Errorf("strategy-engine succeeded but returned empty futures publish payload")
			}
			return detail.Result.Artifacts.Report, nil
		case "FAILED":
			if strings.TrimSpace(detail.ErrorMessage) == "" {
				return strategyEngineFuturesStrategyReport{}, fmt.Errorf("strategy-engine futures job failed")
			}
			return strategyEngineFuturesStrategyReport{}, fmt.Errorf("strategy-engine futures job failed: %s", detail.ErrorMessage)
		}

		if time.Now().After(deadline) {
			return strategyEngineFuturesStrategyReport{}, fmt.Errorf("strategy-engine futures job timed out after %s", c.httpClient.Timeout)
		}
		time.Sleep(c.pollInterval)
	}
}

func (c *strategyEngineClient) getJob(jobID string) (strategyEngineJobDetail, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.baseURL+"/internal/v1/jobs/"+jobID, nil)
	if err != nil {
		return strategyEngineJobDetail{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return strategyEngineJobDetail{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return strategyEngineJobDetail{}, fmt.Errorf("strategy-engine returned %d when fetching job %s: %s", resp.StatusCode, jobID, readBodyText(resp.Body))
	}
	var result strategyEngineJobDetail
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return strategyEngineJobDetail{}, err
	}
	return result, nil
}

func (c *strategyEngineClient) getFuturesJob(jobID string) (strategyEngineFuturesJobDetail, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.baseURL+"/internal/v1/jobs/"+jobID, nil)
	if err != nil {
		return strategyEngineFuturesJobDetail{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return strategyEngineFuturesJobDetail{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return strategyEngineFuturesJobDetail{}, fmt.Errorf("strategy-engine returned %d when fetching futures job %s: %s", resp.StatusCode, jobID, readBodyText(resp.Body))
	}
	var result strategyEngineFuturesJobDetail
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return strategyEngineFuturesJobDetail{}, err
	}
	return result, nil
}

func (c *strategyEngineClient) publishJob(
	jobID string,
	requestedBy string,
	force bool,
	overrideReason string,
	policy *model.StrategyPublishPolicy,
) (strategyEnginePublishRecord, error) {
	bodyMap := map[string]any{
		"requested_by":    strings.TrimSpace(requestedBy),
		"force":           force,
		"override_reason": strings.TrimSpace(overrideReason),
	}
	if bodyMap["requested_by"] == "" {
		bodyMap["requested_by"] = "system"
	}
	if policy != nil {
		bodyMap["policy"] = map[string]any{
			"max_risk_level":         policy.MaxRiskLevel,
			"max_warning_count":      policy.MaxWarningCount,
			"allow_vetoed_publish":   policy.AllowVetoedPublish,
			"default_publisher":      policy.DefaultPublisher,
			"override_note_template": policy.OverrideNoteTemplate,
		}
	}
	body, err := json.Marshal(bodyMap)
	if err != nil {
		return strategyEnginePublishRecord{}, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/internal/v1/publish/jobs/"+jobID, bytes.NewReader(body))
	if err != nil {
		return strategyEnginePublishRecord{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return strategyEnginePublishRecord{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return strategyEnginePublishRecord{}, fmt.Errorf("strategy-engine returned %d when publishing job %s: %s", resp.StatusCode, jobID, readBodyText(resp.Body))
	}
	var result strategyEnginePublishRecord
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return strategyEnginePublishRecord{}, err
	}
	if strings.TrimSpace(result.PublishID) == "" {
		return strategyEnginePublishRecord{}, fmt.Errorf("strategy-engine returned empty publish_id")
	}
	return result, nil
}

func (c *strategyEngineClient) listPublishHistory(jobType string) ([]model.StrategyEnginePublishRecordSummary, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.baseURL+"/internal/v1/publish/history/"+jobType, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("strategy-engine returned %d when fetching publish history %s: %s", resp.StatusCode, jobType, readBodyText(resp.Body))
	}
	var result strategyEnginePublishHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	items := make([]model.StrategyEnginePublishRecordSummary, 0, len(result.Records))
	for _, item := range result.Records {
		items = append(items, model.StrategyEnginePublishRecordSummary{
			PublishID:     item.PublishID,
			JobID:         item.JobID,
			JobType:       item.JobType,
			Version:       item.Version,
			CreatedAt:     item.CreatedAt,
			TradeDate:     item.TradeDate,
			ReportSummary: item.ReportSummary,
			SelectedCount: item.SelectedCount,
			AssetKeys:     item.AssetKeys,
			PayloadCount:  item.PayloadCount,
		})
	}
	return items, nil
}

func (c *strategyEngineClient) getPublishRecord(publishID string) (model.StrategyEnginePublishRecord, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.baseURL+"/internal/v1/publish/records/"+publishID, nil)
	if err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return model.StrategyEnginePublishRecord{}, fmt.Errorf("strategy-engine returned %d when fetching publish record %s: %s", resp.StatusCode, publishID, readBodyText(resp.Body))
	}
	var result strategyEnginePublishRecordDetail
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	return model.StrategyEnginePublishRecord{
		PublishID:       result.PublishID,
		JobID:           result.JobID,
		JobType:         result.JobType,
		Version:         result.Version,
		CreatedAt:       result.CreatedAt,
		TradeDate:       result.TradeDate,
		ReportSummary:   result.ReportSummary,
		SelectedCount:   result.SelectedCount,
		AssetKeys:       result.AssetKeys,
		PayloadCount:    result.PayloadCount,
		Markdown:        result.Markdown,
		HTML:            result.HTML,
		PublishPayloads: result.PublishPayloads,
		ReportSnapshot:  result.ReportSnapshot,
		Replay: model.StrategyEnginePublishReplay{
			WarningCount:      result.Replay.WarningCount,
			WarningMessages:   result.Replay.WarningMessages,
			VetoedAssets:      result.Replay.VetoedAssets,
			InvalidatedAssets: result.Replay.InvalidatedAssets,
			Notes:             result.Replay.Notes,
		},
	}, nil
}

func (c *strategyEngineClient) getPublishReplay(publishID string) (model.StrategyEnginePublishReplay, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.baseURL+"/internal/v1/publish/records/"+publishID+"/replay", nil)
	if err != nil {
		return model.StrategyEnginePublishReplay{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.StrategyEnginePublishReplay{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return model.StrategyEnginePublishReplay{}, fmt.Errorf("strategy-engine returned %d when fetching publish replay %s: %s", resp.StatusCode, publishID, readBodyText(resp.Body))
	}
	var result strategyEnginePublishReplay
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.StrategyEnginePublishReplay{}, err
	}
	return model.StrategyEnginePublishReplay{
		WarningCount:      result.WarningCount,
		WarningMessages:   result.WarningMessages,
		VetoedAssets:      result.VetoedAssets,
		InvalidatedAssets: result.InvalidatedAssets,
		Notes:             result.Notes,
	}, nil
}

func (c *strategyEngineClient) comparePublishVersions(leftPublishID string, rightPublishID string) (model.StrategyEnginePublishCompareResult, error) {
	body, err := json.Marshal(map[string]string{
		"left_publish_id":  leftPublishID,
		"right_publish_id": rightPublishID,
	})
	if err != nil {
		return model.StrategyEnginePublishCompareResult{}, err
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, c.baseURL+"/internal/v1/publish/compare", bytes.NewReader(body))
	if err != nil {
		return model.StrategyEnginePublishCompareResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.StrategyEnginePublishCompareResult{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return model.StrategyEnginePublishCompareResult{}, fmt.Errorf("strategy-engine returned %d when comparing publish versions: %s", resp.StatusCode, readBodyText(resp.Body))
	}
	var result strategyEnginePublishCompareResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.StrategyEnginePublishCompareResult{}, err
	}
	return model.StrategyEnginePublishCompareResult{
		LeftPublishID:      result.LeftPublishID,
		RightPublishID:     result.RightPublishID,
		LeftVersion:        result.LeftVersion,
		RightVersion:       result.RightVersion,
		SelectedCountDelta: result.SelectedCountDelta,
		PayloadCountDelta:  result.PayloadCountDelta,
		WarningCountDelta:  result.WarningCountDelta,
		VetoCountDelta:     result.VetoCountDelta,
		AddedAssets:        result.AddedAssets,
		RemovedAssets:      result.RemovedAssets,
		SharedAssets:       result.SharedAssets,
	}, nil
}

func (c *strategyEngineClient) listJobs(jobType string, status string, page int, pageSize int) ([]model.StrategyEngineJobRecord, int, error) {
	query := url.Values{}
	if trimmed := strings.TrimSpace(jobType); trimmed != "" {
		query.Set("job_type", trimmed)
	}
	if trimmed := strings.TrimSpace(status); trimmed != "" {
		query.Set("status", trimmed)
	}
	if page > 0 {
		query.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		query.Set("page_size", strconv.Itoa(pageSize))
	}

	targetURL := c.baseURL + "/internal/v1/jobs"
	if encoded := query.Encode(); encoded != "" {
		targetURL += "?" + encoded
	}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, targetURL, nil)
	if err != nil {
		return nil, 0, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("strategy-engine returned %d when listing jobs: %s", resp.StatusCode, readBodyText(resp.Body))
	}

	var result strategyEngineJobListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	items := make([]model.StrategyEngineJobRecord, 0, len(result.Items))
	for _, item := range result.Items {
		items = append(items, toStrategyEngineJobRecord(item))
	}
	return items, result.Total, nil
}

func (c *strategyEngineClient) getJobRecord(jobID string) (model.StrategyEngineJobRecord, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, c.baseURL+"/internal/v1/jobs/"+jobID, nil)
	if err != nil {
		return model.StrategyEngineJobRecord{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return model.StrategyEngineJobRecord{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return model.StrategyEngineJobRecord{}, sql.ErrNoRows
	}
	if resp.StatusCode != http.StatusOK {
		return model.StrategyEngineJobRecord{}, fmt.Errorf("strategy-engine returned %d when fetching job %s: %s", resp.StatusCode, jobID, readBodyText(resp.Body))
	}

	var result strategyEngineAdminJobRecord
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return model.StrategyEngineJobRecord{}, err
	}
	return toStrategyEngineJobRecord(result), nil
}

func toStrategyEngineJobRecord(item strategyEngineAdminJobRecord) model.StrategyEngineJobRecord {
	record := model.StrategyEngineJobRecord{
		JobID:                item.JobID,
		JobType:              item.JobType,
		Status:               item.Status,
		RequestedBy:          item.RequestedBy,
		TraceID:              item.TraceID,
		Payload:              item.Payload,
		ErrorMessage:         item.ErrorMessage,
		PublishCount:         item.PublishCount,
		LatestPublishID:      item.LatestPublishID,
		LatestPublishVersion: item.LatestPublishVersion,
		LatestPublishAt:      item.LatestPublishAt,
		LatestPublishMode:    strings.ToUpper(strings.TrimSpace(item.LatestPublishMode)),
		LatestPublishSource:  strings.TrimSpace(item.LatestPublishSource),
		CreatedAt:            item.CreatedAt,
		StartedAt:            item.StartedAt,
		FinishedAt:           item.FinishedAt,
	}
	if item.Result != nil {
		record.Result = &model.StrategyEngineJobResult{
			Summary:     item.Result.Summary,
			PayloadEcho: item.Result.PayloadEcho,
			Artifacts:   item.Result.Artifacts,
			Warnings:    item.Result.Warnings,
		}
	}
	return hydrateStrategyEngineJobSummary(record)
}

func strategyEngineTargetTypeFromJobType(jobType string) string {
	switch strings.TrimSpace(jobType) {
	case "futures-strategy":
		return "FUTURES"
	default:
		return "STOCK"
	}
}

func (r *MySQLGrowthRepo) AdminListStrategyEnginePublishHistory(jobType string) ([]model.StrategyEnginePublishRecordSummary, error) {
	items, err := r.listStrategyEnginePublishRecordSnapshots(jobType)
	if err == nil && len(items) > 0 {
		return items, nil
	}
	if r.strategyEngine == nil {
		if err != nil {
			return nil, err
		}
		return []model.StrategyEnginePublishRecordSummary{}, nil
	}
	remoteItems, remoteErr := r.strategyEngine.listPublishHistory(jobType)
	if remoteErr != nil {
		if err != nil {
			return nil, remoteErr
		}
		return []model.StrategyEnginePublishRecordSummary{}, nil
	}
	return r.backfillStrategyEnginePublishHistory(jobType, remoteItems)
}

func (r *MySQLGrowthRepo) AdminListStrategyEngineJobs(jobType string, status string, page int, pageSize int) ([]model.StrategyEngineJobRecord, int, error) {
	items, total, err := r.listStrategyEngineJobSnapshots(jobType, status, page, pageSize)
	if err == nil && len(items) > 0 {
		for index := range items {
			items[index] = r.attachStrategyJobReplays(items[index])
		}
		return items, total, nil
	}
	if err != nil && r.strategyEngine == nil {
		return nil, 0, err
	}
	if r.strategyEngine == nil {
		return items, total, nil
	}
	remoteItems, remoteTotal, remoteErr := r.strategyEngine.listJobs(jobType, status, page, pageSize)
	if remoteErr != nil {
		if err != nil {
			return nil, 0, remoteErr
		}
		return items, total, nil
	}
	archived := make([]model.StrategyEngineJobRecord, 0, len(remoteItems))
	for _, item := range remoteItems {
		source := "REMOTE_BACKFILLED"
		if persistErr := r.upsertStrategyEngineJobSnapshot(item); persistErr != nil {
			source = "REMOTE_ONLY"
		}
		archived = append(archived, r.attachStrategyJobReplays(strategySnapshotJobRecordWithSource(item, source)))
	}
	return archived, remoteTotal, nil
}

func (r *MySQLGrowthRepo) AdminGetStrategyEngineJob(jobID string) (model.StrategyEngineJobRecord, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return model.StrategyEngineJobRecord{}, sql.ErrNoRows
	}
	item, err := r.getStrategyEngineJobSnapshot(jobID)
	if err == nil {
		return r.attachStrategyJobReplays(item), nil
	}
	if r.strategyEngine == nil {
		return model.StrategyEngineJobRecord{}, err
	}
	remoteItem, remoteErr := r.strategyEngine.getJobRecord(jobID)
	if remoteErr != nil {
		if isNotFoundStrategySnapshot(err) {
			return model.StrategyEngineJobRecord{}, remoteErr
		}
		return model.StrategyEngineJobRecord{}, remoteErr
	}
	if persistErr := r.upsertStrategyEngineJobSnapshot(remoteItem); persistErr != nil {
		return r.attachStrategyJobReplays(strategySnapshotJobRecordWithSource(remoteItem, "REMOTE_ONLY")), nil
	}
	return r.attachStrategyJobReplays(strategySnapshotJobRecordWithSource(remoteItem, "REMOTE_BACKFILLED")), nil
}

func (r *MySQLGrowthRepo) resolveStrategyEngineArchiveJob(jobID string) (model.StrategyEngineJobRecord, error) {
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		return model.StrategyEngineJobRecord{}, sql.ErrNoRows
	}
	if localJob, err := r.getStrategyEngineJobSnapshot(jobID); err == nil {
		return localJob, nil
	}
	if r.strategyEngine == nil {
		return model.StrategyEngineJobRecord{}, sql.ErrNoRows
	}
	return r.strategyEngine.getJobRecord(jobID)
}

func (r *MySQLGrowthRepo) backfillStrategyEnginePublishRecord(publishID string) (model.StrategyEnginePublishRecord, error) {
	if r.strategyEngine == nil {
		return model.StrategyEnginePublishRecord{}, sql.ErrNoRows
	}
	record, err := r.strategyEngine.getPublishRecord(publishID)
	if err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	jobRecord, jobErr := r.resolveStrategyEngineArchiveJob(record.JobID)
	if jobErr != nil {
		return record, nil
	}
	r.archiveStrategyEnginePublishSnapshot(record, jobRecord)
	if localRecord, localErr := r.getStrategyEnginePublishRecordSnapshot(publishID); localErr == nil {
		return localRecord, nil
	}
	return record, nil
}

func (r *MySQLGrowthRepo) AdminPublishStrategyEngineJob(jobID string, operator string, force bool, overrideReason string) (model.StrategyEnginePublishRecord, error) {
	job, err := r.AdminGetStrategyEngineJob(jobID)
	if err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	policy, err := r.ResolveActiveStrategyPublishPolicy(strategyEngineTargetTypeFromJobType(job.JobType))
	if err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	if r.strategyEngine == nil {
		return r.publishStrategyEngineArchivedJob(job, operator, force, overrideReason, policy, nil)
	}

	published, err := r.strategyEngine.publishJob(jobID, operator, force, overrideReason, policy)
	if err != nil {
		if isStrategyEngineLivePublishJobMissing(err) {
			return r.publishStrategyEngineArchivedJob(job, operator, force, overrideReason, policy, nil)
		}
		return model.StrategyEnginePublishRecord{}, err
	}
	return r.persistRemoteStrategyEnginePublish(job, published, operator, force, overrideReason, policy)
}

func (r *MySQLGrowthRepo) persistRemoteStrategyEnginePublish(
	job model.StrategyEngineJobRecord,
	published strategyEnginePublishRecord,
	operator string,
	force bool,
	overrideReason string,
	policy *model.StrategyPublishPolicy,
) (model.StrategyEnginePublishRecord, error) {
	record, err := r.strategyEngine.getPublishRecord(published.PublishID)
	if err != nil {
		if !isStrategyEnginePublishRecordMissing(err) {
			return model.StrategyEnginePublishRecord{}, err
		}
		return r.publishStrategyEngineArchivedJob(job, operator, force, overrideReason, policy, &model.StrategyEnginePublishRecord{
			PublishID:     published.PublishID,
			JobID:         job.JobID,
			JobType:       job.JobType,
			Version:       published.Version,
			TradeDate:     published.TradeDate,
			ReportSummary: published.ReportSummary,
			SelectedCount: published.SelectedCount,
			PayloadCount:  published.PayloadCount,
		})
	}

	jobRecord, err := r.strategyEngine.getJobRecord(job.JobID)
	if err != nil {
		if isNotFoundStrategySnapshot(err) || isStrategyEngineJobMissing(err) {
			jobRecord = job
		} else {
			return model.StrategyEnginePublishRecord{}, err
		}
	}
	if err := r.upsertStrategyEngineJobSnapshot(jobRecord); err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	if err := r.createStrategyEngineJobReplay(job.JobID, record, operator, force, overrideReason, policy); err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	return record, nil
}

func (r *MySQLGrowthRepo) publishStrategyEngineArchivedJob(
	job model.StrategyEngineJobRecord,
	operator string,
	force bool,
	overrideReason string,
	policy *model.StrategyPublishPolicy,
	seed *model.StrategyEnginePublishRecord,
) (model.StrategyEnginePublishRecord, error) {
	job = hydrateStrategyEngineJobSummary(job)
	if strings.ToUpper(strings.TrimSpace(job.Status)) != "SUCCEEDED" || job.Result == nil {
		return model.StrategyEnginePublishRecord{}, fmt.Errorf("job is not ready for publish")
	}

	report := strategySnapshotReport(job.Result)
	if len(report) == 0 {
		return model.StrategyEnginePublishRecord{}, fmt.Errorf("job does not contain a publishable report artifact")
	}

	replay := buildStrategyEngineArchivedPublishReplay(job.JobType, report, job.Result.Warnings)
	if err := validateStrategyEngineArchivedPublishPolicy(job.JobType, report, replay, policy, force); err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}

	replay = appendStrategyEngineArchivedPublishNotes(replay, operator, force, overrideReason, policy)

	assetKeys := strategySnapshotAssetKeys(report)
	if len(assetKeys) == 0 && seed != nil {
		assetKeys = append([]string{}, seed.AssetKeys...)
	}

	publishPayloads := sliceOfMaps(report["publish_payloads"])
	if len(publishPayloads) == 0 && seed != nil && len(seed.PublishPayloads) > 0 {
		publishPayloads = append([]map[string]any{}, seed.PublishPayloads...)
	}

	selectedCount := strategyEngineSelectedCount(report, assetKeys)
	if selectedCount == 0 && seed != nil {
		selectedCount = seed.SelectedCount
	}
	payloadCount := len(publishPayloads)
	if payloadCount == 0 && seed != nil {
		payloadCount = seed.PayloadCount
	}

	createdAt := time.Now().UTC().Format(time.RFC3339)
	if seed != nil && strings.TrimSpace(seed.CreatedAt) != "" {
		createdAt = strings.TrimSpace(seed.CreatedAt)
	}
	version := strategyEngineNextPublishVersion(job)
	if seed != nil && seed.Version > version {
		version = seed.Version
	}
	publishID := newID("publish")
	if seed != nil && strings.TrimSpace(seed.PublishID) != "" {
		publishID = strings.TrimSpace(seed.PublishID)
	}

	record := model.StrategyEnginePublishRecord{
		PublishID:       publishID,
		JobID:           job.JobID,
		JobType:         job.JobType,
		Version:         version,
		CreatedAt:       createdAt,
		TradeDate:       firstNonEmptyString(job.TradeDate, strategySnapshotTradeDate(job.Payload), strategySnapshotTradeDate(job.ResultPayloadEcho()), strategySnapshotReportTradeDate(job.Result)),
		ReportSummary:   firstNonEmptyString(asString(report["report_summary"]), strings.TrimSpace(job.Result.Summary)),
		SelectedCount:   selectedCount,
		AssetKeys:       assetKeys,
		PayloadCount:    payloadCount,
		PublishPayloads: publishPayloads,
		ReportSnapshot:  report,
		Replay: model.StrategyEnginePublishReplay{
			PublishID:      publishID,
			JobID:          job.JobID,
			PublishVersion: version,
			Operator:       strings.TrimSpace(operator),
			ForcePublish:   force,
			OverrideReason: strings.TrimSpace(overrideReason),
			PolicySnapshot: buildStrategyPublishPolicyPreview(policy),
			CreatedAt:      createdAt,
			StorageSource:  "LOCAL_ARCHIVED",
			WarningCount:   replay.WarningCount,
			WarningMessages: append([]string{},
				replay.WarningMessages...),
			VetoedAssets:      append([]string{}, replay.VetoedAssets...),
			InvalidatedAssets: append([]string{}, replay.InvalidatedAssets...),
			Notes:             append([]string{}, replay.Notes...),
		},
	}
	if strings.TrimSpace(record.ReportSummary) == "" && seed != nil {
		record.ReportSummary = strings.TrimSpace(seed.ReportSummary)
	}
	if strings.TrimSpace(record.TradeDate) == "" && seed != nil {
		record.TradeDate = strings.TrimSpace(seed.TradeDate)
	}
	if err := r.upsertStrategyEngineJobSnapshot(job); err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	if err := r.createStrategyEngineJobReplay(job.JobID, record, operator, force, overrideReason, policy); err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	return record, nil
}

func buildStrategyEngineArchivedPublishReplay(jobType string, report map[string]any, warnings []string) model.StrategyEnginePublishReplay {
	items, keyName := strategyEnginePublishItems(jobType, report)

	invalidatedAssets := make([]string, 0)
	for _, item := range items {
		assetKey := strings.TrimSpace(asString(item[keyName]))
		if assetKey == "" || !strategyEngineHasInvalidations(item["invalidations"]) {
			continue
		}
		invalidatedAssets = append(invalidatedAssets, assetKey)
	}

	vetoedAssets := make([]string, 0)
	for _, item := range sliceOfMaps(report["simulations"]) {
		assetKey := strings.TrimSpace(asString(item["asset_key"]))
		if assetKey == "" || !asBool(item["vetoed"]) {
			continue
		}
		vetoedAssets = append(vetoedAssets, assetKey)
	}

	notes := make([]string, 0, 4)
	if len(warnings) > 0 {
		notes = append(notes, fmt.Sprintf("本次发布包含 %d 条风控或过滤提醒。", len(warnings)))
	}
	if len(vetoedAssets) > 0 {
		notes = append(notes, "被风险 agent 否决的标的: "+strings.Join(vetoedAssets, "、")+"。")
	}
	if len(invalidatedAssets) > 0 {
		notes = append(notes, fmt.Sprintf("已记录失效条件的标的数: %d。", len(invalidatedAssets)))
	}
	if len(notes) == 0 {
		notes = append(notes, "本次发布未出现额外警告，可作为后续复盘基线版本。")
	}

	return model.StrategyEnginePublishReplay{
		WarningCount:      len(warnings),
		WarningMessages:   append([]string{}, warnings...),
		VetoedAssets:      vetoedAssets,
		InvalidatedAssets: invalidatedAssets,
		Notes:             notes,
	}
}

func validateStrategyEngineArchivedPublishPolicy(
	jobType string,
	report map[string]any,
	replay model.StrategyEnginePublishReplay,
	policy *model.StrategyPublishPolicy,
	force bool,
) error {
	if policy == nil {
		return nil
	}

	items, keyName := strategyEnginePublishItems(jobType, report)
	riskRank := map[string]int{"LOW": 1, "MEDIUM": 2, "HIGH": 3}
	allowedRank := riskRank[strings.ToUpper(strings.TrimSpace(policy.MaxRiskLevel))]
	if allowedRank == 0 {
		allowedRank = riskRank["MEDIUM"]
	}

	breaches := make([]string, 0, 3)
	highRiskAssets := make([]string, 0)
	for _, item := range items {
		assetKey := strings.TrimSpace(asString(item[keyName]))
		if assetKey == "" {
			continue
		}
		if riskRank[strings.ToUpper(strings.TrimSpace(asString(item["risk_level"])))] > allowedRank {
			highRiskAssets = append(highRiskAssets, assetKey)
		}
	}
	if len(highRiskAssets) > 0 {
		breaches = append(breaches, fmt.Sprintf("存在风险等级超过 %s 的标的: %s", policy.MaxRiskLevel, strings.Join(highRiskAssets, "、")))
	}
	if replay.WarningCount > policy.MaxWarningCount {
		breaches = append(breaches, fmt.Sprintf("警告数量 %d 超过阈值 %d", replay.WarningCount, policy.MaxWarningCount))
	}
	if len(replay.VetoedAssets) > 0 && !policy.AllowVetoedPublish {
		breaches = append(breaches, "存在被 veto 的标的: "+strings.Join(replay.VetoedAssets, "、"))
	}

	if len(breaches) > 0 && !force {
		return fmt.Errorf("发布策略拦截: %s", strings.Join(breaches, "；"))
	}
	return nil
}

func appendStrategyEngineArchivedPublishNotes(
	replay model.StrategyEnginePublishReplay,
	operator string,
	force bool,
	overrideReason string,
	policy *model.StrategyPublishPolicy,
) model.StrategyEnginePublishReplay {
	if force && strings.TrimSpace(overrideReason) != "" {
		replay.Notes = append(replay.Notes, "人工覆盖发布原因: "+strings.TrimSpace(overrideReason)+"。")
	} else if force {
		replay.Notes = append(replay.Notes, "本次发布通过人工覆盖放行。")
	}
	if policy != nil && strings.TrimSpace(policy.DefaultPublisher) != "" {
		replay.Notes = append(replay.Notes, "策略默认发布者: "+strings.TrimSpace(policy.DefaultPublisher)+"。")
	}
	if policy != nil && strings.TrimSpace(policy.OverrideNoteTemplate) != "" {
		replay.Notes = append(replay.Notes, strings.TrimSpace(policy.OverrideNoteTemplate))
	}
	replay.Notes = append(replay.Notes, "发布操作人: "+firstNonEmpty(strings.TrimSpace(operator), "system")+"。")
	return replay
}

func strategyEnginePublishItems(jobType string, report map[string]any) ([]map[string]any, string) {
	if strings.TrimSpace(jobType) == "stock-selection" {
		return sliceOfMaps(report["candidates"]), "symbol"
	}
	return sliceOfMaps(report["strategies"]), "contract"
}

func strategyEngineHasInvalidations(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case bool:
		return typed
	case string:
		return strings.TrimSpace(typed) != ""
	case []string:
		return len(typed) > 0
	case []map[string]any:
		return len(typed) > 0
	case []any:
		return len(typed) > 0
	case map[string]any:
		return len(typed) > 0
	default:
		return false
	}
}

func strategyEngineSelectedCount(report map[string]any, assetKeys []string) int {
	if raw, ok := report["selected_count"]; ok {
		return asInt(raw)
	}
	return len(assetKeys)
}

func strategyEngineNextPublishVersion(job model.StrategyEngineJobRecord) int {
	nextVersion := maxInt(job.LatestPublishVersion, 0)
	for _, replay := range job.Replays {
		nextVersion = maxInt(nextVersion, replay.PublishVersion)
	}
	return nextVersion + 1
}

func isStrategyEngineLivePublishJobMissing(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "returned 404 when publishing job") && strings.Contains(message, "job not found")
}

func isStrategyEnginePublishRecordMissing(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "returned 404 when fetching publish record")
}

func isStrategyEngineJobMissing(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "returned 404 when fetching job") && strings.Contains(message, "job not found")
}

func (r *MySQLGrowthRepo) AdminGetStrategyEnginePublishRecord(publishID string) (model.StrategyEnginePublishRecord, error) {
	publishID = strings.TrimSpace(publishID)
	if publishID == "" {
		return model.StrategyEnginePublishRecord{}, sql.ErrNoRows
	}
	record, err := r.getStrategyEnginePublishRecordSnapshot(publishID)
	if err == nil {
		return record, nil
	}
	if r.strategyEngine == nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	record, err = r.backfillStrategyEnginePublishRecord(publishID)
	if err != nil {
		return model.StrategyEnginePublishRecord{}, err
	}
	return record, nil
}

func (r *MySQLGrowthRepo) AdminGetStrategyEnginePublishReplay(publishID string) (model.StrategyEnginePublishReplay, error) {
	if replay, err := r.getStrategyEngineJobReplayByPublishID(publishID); err == nil {
		return replay, nil
	}
	if r.strategyEngine == nil {
		return model.StrategyEnginePublishReplay{}, sql.ErrNoRows
	}
	record, recordErr := r.AdminGetStrategyEnginePublishRecord(publishID)
	if recordErr == nil {
		if replay, replayErr := r.getStrategyEngineJobReplayByPublishID(publishID); replayErr == nil {
			return replay, nil
		}
		if strings.TrimSpace(record.PublishID) != "" {
			replay := record.Replay
			replay.PublishID = firstNonEmpty(replay.PublishID, record.PublishID)
			replay.JobID = firstNonEmpty(replay.JobID, record.JobID)
			replay.PublishVersion = maxInt(replay.PublishVersion, record.Version)
			replay.CreatedAt = firstNonEmpty(replay.CreatedAt, record.CreatedAt)
			if replay.JobID != "" && r.upsertStrategyEngineJobReplaySnapshot(replay.JobID, replay, "", false, "", nil) == nil {
				if localReplay, replayErr := r.getStrategyEngineJobReplayByPublishID(publishID); replayErr == nil {
					return localReplay, nil
				}
			}
			if strings.TrimSpace(replay.StorageSource) == "" {
				replay.StorageSource = "REMOTE_ONLY"
			}
			return replay, nil
		}
	}
	replay, err := r.strategyEngine.getPublishReplay(publishID)
	if err != nil {
		return model.StrategyEnginePublishReplay{}, err
	}
	replay.PublishID = strings.TrimSpace(publishID)
	replay.StorageSource = "REMOTE_ONLY"
	return replay, nil
}

func (r *MySQLGrowthRepo) AdminCompareStrategyEnginePublishVersions(leftPublishID string, rightPublishID string) (model.StrategyEnginePublishCompareResult, error) {
	leftRecord, err := r.AdminGetStrategyEnginePublishRecord(leftPublishID)
	if err != nil {
		return model.StrategyEnginePublishCompareResult{}, err
	}
	rightRecord, err := r.AdminGetStrategyEnginePublishRecord(rightPublishID)
	if err != nil {
		return model.StrategyEnginePublishCompareResult{}, err
	}
	return compareStrategyEnginePublishRecords(leftRecord, rightRecord), nil
}

func compareStrategyEnginePublishRecords(left model.StrategyEnginePublishRecord, right model.StrategyEnginePublishRecord) model.StrategyEnginePublishCompareResult {
	addedAssets, removedAssets, sharedAssets := diffStrategyAssetKeys(left.AssetKeys, right.AssetKeys)
	return model.StrategyEnginePublishCompareResult{
		LeftPublishID:      left.PublishID,
		RightPublishID:     right.PublishID,
		LeftVersion:        left.Version,
		RightVersion:       right.Version,
		SelectedCountDelta: right.SelectedCount - left.SelectedCount,
		PayloadCountDelta:  right.PayloadCount - left.PayloadCount,
		WarningCountDelta:  right.Replay.WarningCount - left.Replay.WarningCount,
		VetoCountDelta:     len(right.Replay.VetoedAssets) - len(left.Replay.VetoedAssets),
		AddedAssets:        addedAssets,
		RemovedAssets:      removedAssets,
		SharedAssets:       sharedAssets,
	}
}

func (r *MySQLGrowthRepo) backfillStrategyEnginePublishHistory(jobType string, remoteItems []model.StrategyEnginePublishRecordSummary) ([]model.StrategyEnginePublishRecordSummary, error) {
	if len(remoteItems) == 0 {
		return []model.StrategyEnginePublishRecordSummary{}, nil
	}
	backfilledItems := make([]model.StrategyEnginePublishRecordSummary, 0, len(remoteItems))
	for _, item := range remoteItems {
		if strings.TrimSpace(item.PublishID) == "" {
			continue
		}
		record, err := r.backfillStrategyEnginePublishRecord(item.PublishID)
		if err != nil {
			continue
		}
		backfilledItems = append(backfilledItems, strategyEnginePublishSummaryFromRecord(record))
	}
	localItems, err := r.listStrategyEnginePublishRecordSnapshots(jobType)
	if err != nil || len(localItems) == 0 {
		if len(backfilledItems) > 0 {
			return mergeStrategyEnginePublishHistorySummaries(backfilledItems, remoteItems), nil
		}
		return remoteItems, nil
	}
	return mergeStrategyEnginePublishHistorySummaries(localItems, remoteItems), nil
}

func mergeStrategyEnginePublishHistorySummaries(localItems []model.StrategyEnginePublishRecordSummary, remoteItems []model.StrategyEnginePublishRecordSummary) []model.StrategyEnginePublishRecordSummary {
	if len(localItems) == 0 {
		return remoteItems
	}
	if len(remoteItems) == 0 {
		return localItems
	}
	localByPublishID := make(map[string]model.StrategyEnginePublishRecordSummary, len(localItems))
	for _, item := range localItems {
		publishID := strings.TrimSpace(item.PublishID)
		if publishID == "" {
			continue
		}
		localByPublishID[publishID] = item
	}
	merged := make([]model.StrategyEnginePublishRecordSummary, 0, maxInt(len(localItems), len(remoteItems)))
	seen := make(map[string]struct{}, len(remoteItems))
	for _, item := range remoteItems {
		publishID := strings.TrimSpace(item.PublishID)
		if publishID == "" {
			continue
		}
		seen[publishID] = struct{}{}
		if localItem, ok := localByPublishID[publishID]; ok {
			merged = append(merged, localItem)
			continue
		}
		merged = append(merged, item)
	}
	for _, item := range localItems {
		publishID := strings.TrimSpace(item.PublishID)
		if publishID == "" {
			continue
		}
		if _, ok := seen[publishID]; ok {
			continue
		}
		merged = append(merged, item)
	}
	return merged
}

func strategyEnginePublishSummaryFromRecord(record model.StrategyEnginePublishRecord) model.StrategyEnginePublishRecordSummary {
	return model.StrategyEnginePublishRecordSummary{
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
	}
}

func diffStrategyAssetKeys(left []string, right []string) ([]string, []string, []string) {
	leftSet := make(map[string]struct{}, len(left))
	rightSet := make(map[string]struct{}, len(right))
	for _, item := range left {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			leftSet[trimmed] = struct{}{}
		}
	}
	for _, item := range right {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			rightSet[trimmed] = struct{}{}
		}
	}

	added := make([]string, 0)
	removed := make([]string, 0)
	shared := make([]string, 0)
	for item := range rightSet {
		if _, ok := leftSet[item]; ok {
			shared = append(shared, item)
			continue
		}
		added = append(added, item)
	}
	for item := range leftSet {
		if _, ok := rightSet[item]; !ok {
			removed = append(removed, item)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(shared)
	return added, removed, shared
}

func (r *MySQLGrowthRepo) attachStrategyJobReplays(item model.StrategyEngineJobRecord) model.StrategyEngineJobRecord {
	replays, err := r.listStrategyEngineJobReplays(item.JobID)
	if err == nil {
		item.Replays = replays
	}
	return hydrateStrategyEngineJobSummary(item)
}

func mergeStrategyPublishReplay(base model.StrategyEnginePublishReplay, local model.StrategyEnginePublishReplay) model.StrategyEnginePublishReplay {
	merged := base
	if strings.TrimSpace(local.PublishID) != "" {
		merged.PublishID = local.PublishID
	}
	if strings.TrimSpace(local.JobID) != "" {
		merged.JobID = local.JobID
	}
	if local.PublishVersion > 0 {
		merged.PublishVersion = local.PublishVersion
	}
	if strings.TrimSpace(local.Operator) != "" {
		merged.Operator = local.Operator
	}
	merged.ForcePublish = local.ForcePublish
	if strings.TrimSpace(local.OverrideReason) != "" {
		merged.OverrideReason = local.OverrideReason
	}
	if len(local.PolicySnapshot) > 0 {
		merged.PolicySnapshot = local.PolicySnapshot
	}
	if strings.TrimSpace(local.CreatedAt) != "" {
		merged.CreatedAt = local.CreatedAt
	}
	if strings.TrimSpace(local.StorageSource) != "" {
		merged.StorageSource = local.StorageSource
	}
	if local.WarningCount > 0 || len(local.WarningMessages) > 0 || len(local.VetoedAssets) > 0 || len(local.InvalidatedAssets) > 0 || len(local.Notes) > 0 {
		merged.WarningCount = local.WarningCount
		merged.WarningMessages = local.WarningMessages
		merged.VetoedAssets = local.VetoedAssets
		merged.InvalidatedAssets = local.InvalidatedAssets
		merged.Notes = local.Notes
	}
	return merged
}

func (r *MySQLGrowthRepo) persistStrategyEngineStockRecommendations(payloads []strategyEngineStockPublishPayload) (int, error) {
	if len(payloads) == 0 {
		return 0, nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	count := 0
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, payload := range payloads {
		validFrom, parseErr := parseStrategyEngineTime(payload.Recommendation.ValidFrom)
		if parseErr != nil {
			err = parseErr
			return count, err
		}
		validTo, parseErr := parseStrategyEngineTime(payload.Recommendation.ValidTo)
		if parseErr != nil {
			err = parseErr
			return count, err
		}

		id := newID("sr")
		_, err = tx.Exec(`
INSERT INTO stock_recommendations (id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, source_type, strategy_version, reviewer, publisher, review_note, performance_label, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id,
			payload.Recommendation.Symbol,
			payload.Recommendation.Name,
			payload.Recommendation.Score,
			payload.Recommendation.RiskLevel,
			payload.Recommendation.PositionRange,
			validFrom,
			validTo,
			payload.Recommendation.Status,
			payload.Recommendation.ReasonSummary,
			payload.Recommendation.SourceType,
			payload.Recommendation.StrategyVersion,
			payload.Recommendation.Reviewer,
			payload.Recommendation.Publisher,
			payload.Recommendation.ReviewNote,
			payload.Recommendation.PerformanceLabel,
			time.Now(),
		)
		if err != nil {
			return count, err
		}

		_, err = tx.Exec(`
INSERT INTO stock_reco_details (reco_id, tech_score, fund_score, sentiment_score, money_flow_score, take_profit, stop_loss, risk_note)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE tech_score=VALUES(tech_score), fund_score=VALUES(fund_score), sentiment_score=VALUES(sentiment_score), money_flow_score=VALUES(money_flow_score), take_profit=VALUES(take_profit), stop_loss=VALUES(stop_loss), risk_note=VALUES(risk_note)`,
			id,
			payload.Detail.TechScore,
			payload.Detail.FundScore,
			payload.Detail.SentimentScore,
			payload.Detail.MoneyFlowScore,
			payload.Detail.TakeProfit,
			payload.Detail.StopLoss,
			payload.Detail.RiskNote,
		)
		if err != nil {
			return count, err
		}
		count++
	}

	if err = tx.Commit(); err != nil {
		return count, err
	}
	return count, nil
}

func (r *MySQLGrowthRepo) persistStrategyEngineFuturesStrategies(payloads []strategyEngineFuturesPublishPayload) (int, error) {
	if len(payloads) == 0 {
		return 0, nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}
	count := 0
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	for _, payload := range payloads {
		validFrom, parseErr := parseStrategyEngineTime(payload.Strategy.ValidFrom)
		if parseErr != nil {
			err = parseErr
			return count, err
		}
		validTo, parseErr := parseStrategyEngineTime(payload.Strategy.ValidTo)
		if parseErr != nil {
			err = parseErr
			return count, err
		}

		strategyID := newID("fs")
		_, err = tx.Exec(`
INSERT INTO futures_strategies (id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			strategyID,
			payload.Strategy.Contract,
			payload.Strategy.Name,
			payload.Strategy.Direction,
			payload.Strategy.RiskLevel,
			payload.Strategy.PositionRange,
			validFrom,
			validTo,
			payload.Strategy.Status,
			payload.Strategy.ReasonSummary,
		)
		if err != nil {
			return count, err
		}

		guidanceID := newID("fg")
		_, err = tx.Exec(`
INSERT INTO futures_guidances (id, contract, guidance_direction, position_level, entry_range, take_profit_range, stop_loss_range, risk_level, invalid_condition, valid_to)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			guidanceID,
			payload.Guidance.Contract,
			payload.Guidance.GuidanceDirection,
			payload.Guidance.PositionLevel,
			payload.Guidance.EntryRange,
			payload.Guidance.TakeProfitRange,
			payload.Guidance.StopLossRange,
			payload.Guidance.RiskLevel,
			payload.Guidance.InvalidCondition,
			validTo,
		)
		if err != nil {
			return count, err
		}
		count++
	}

	if err = tx.Commit(); err != nil {
		return count, err
	}
	return count, nil
}

func parseStrategyEngineTime(raw string) (time.Time, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Time{}, fmt.Errorf("strategy-engine returned empty datetime")
	}
	if parsed, err := time.Parse(time.RFC3339, text); err == nil {
		return parsed, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02", text, time.Local); err == nil {
		return parsed, nil
	}
	return time.Time{}, fmt.Errorf("invalid strategy-engine datetime: %s", text)
}

func readBodyText(body io.Reader) string {
	data, err := io.ReadAll(io.LimitReader(body, 2048))
	if err != nil {
		return err.Error()
	}
	return strings.TrimSpace(string(data))
}
