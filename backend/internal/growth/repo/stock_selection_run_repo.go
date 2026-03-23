package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminGetStockSelectionOverview() (model.AdminStockSelectionOverview, error) {
	defaultProfile, err := r.getDefaultStockSelectionProfile()
	if err != nil && err != sql.ErrNoRows {
		return model.AdminStockSelectionOverview{}, err
	}
	defaultTemplate, err := r.getDefaultStockSelectionProfileTemplate()
	if err != nil && err != sql.ErrNoRows {
		return model.AdminStockSelectionOverview{}, err
	}
	latestRun, err := r.getLatestStockSelectionRun("")
	if err != nil && err != sql.ErrNoRows {
		return model.AdminStockSelectionOverview{}, err
	}
	latestSuccessRun, err := r.getLatestStockSelectionRun("SUCCEEDED")
	if err != nil && err != sql.ErrNoRows {
		return model.AdminStockSelectionOverview{}, err
	}

	evaluationSummary := make(map[string]any)
	warnings := make([]string, 0)
	for _, window := range []int{5, 10, 20} {
		summary, _, _, _, evalErr := r.AdminGetQuantEvaluation(window, 5)
		if evalErr != nil {
			warnings = append(warnings, fmt.Sprintf("%d日评估摘要加载失败: %v", window, evalErr))
			continue
		}
		evaluationSummary[strconv.Itoa(window)] = summary
	}
	leaderboard, leaderboardErr := r.AdminListStockSelectionEvaluationLeaderboard("", "", "")
	if leaderboardErr != nil {
		warnings = append(warnings, fmt.Sprintf("智能选股评估榜单加载失败: %v", leaderboardErr))
	}
	evaluationSummaryV2, evaluationSummaryV2Err := r.loadStockSelectionOverviewEvaluationSummary()
	if evaluationSummaryV2Err != nil {
		warnings = append(warnings, fmt.Sprintf("智能选股评估摘要加载失败: %v", evaluationSummaryV2Err))
		evaluationSummaryV2 = map[string]any{}
	}

	dataFreshness := map[string]any{}
	latestTradeDate := ""
	latestAvailableTradeDate, latestTradeDateErr := r.resolveStrategyStockContextTradeDate(time.Now(), nil)
	if latestTradeDateErr != nil {
		warnings = append(warnings, fmt.Sprintf("智能选股最新交易日加载失败: %v", latestTradeDateErr))
	} else if !latestAvailableTradeDate.IsZero() {
		latestTradeDate = latestAvailableTradeDate.Format("2006-01-02")
	}
	if latestSuccessRun != nil {
		dataFreshness = latestSuccessRun.ContextMeta
		dataFreshness["trade_date"] = latestSuccessRun.TradeDate
		dataFreshness["completed_at"] = latestSuccessRun.CompletedAt
	} else if latestRun != nil {
		if latestTradeDate == "" {
			latestTradeDate = latestRun.TradeDate
		}
	}
	marketRegime := ""
	if latestSuccessRun != nil {
		marketRegime = latestSuccessRun.MarketRegime
	} else if latestRun != nil {
		marketRegime = latestRun.MarketRegime
	}

	var pendingReviewCount int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM stock_selection_publish_reviews WHERE review_status = 'PENDING'").Scan(&pendingReviewCount); err != nil {
		warnings = append(warnings, fmt.Sprintf("待审核数量加载失败: %v", err))
	}
	latestApprovedPortfolio, _ := r.loadLatestApprovedPortfolioEntries()

	templateSummary := map[string]any{}
	if defaultTemplate != nil {
		templateSummary = map[string]any{
			"default_template_id":   defaultTemplate.ID,
			"default_template_name": defaultTemplate.Name,
			"market_regime_bias":    defaultTemplate.MarketRegimeBias,
		}
	}
	evaluationSummaryV2["leaderboard_items"] = leaderboard

	return model.AdminStockSelectionOverview{
		DefaultProfile:          defaultProfile,
		LatestTradeDate:         latestTradeDate,
		LatestRun:               latestRun,
		LatestSuccessRun:        latestSuccessRun,
		LatestApprovedPortfolio: latestApprovedPortfolio,
		MarketRegime:            marketRegime,
		DataFreshness:           dataFreshness,
		EvaluationSummary:       evaluationSummary,
		EvaluationSummaryV2:     evaluationSummaryV2,
		TemplateSummary:         templateSummary,
		PendingReviewCount:      pendingReviewCount,
		Warnings:                warnings,
		QuickActions: []model.AdminStockSelectionQuickAction{
			{Key: "run_now", Label: "立即运行", ActionType: "RUN", TargetRoute: "/stock-selection/runs"},
			{Key: "manage_templates", Label: "策略模板", ActionType: "NAVIGATE", TargetRoute: "/stock-selection/templates"},
			{Key: "manage_profiles", Label: "策略配置", ActionType: "NAVIGATE", TargetRoute: "/stock-selection/profiles"},
			{Key: "review_candidates", Label: "候选与组合", ActionType: "NAVIGATE", TargetRoute: "/stock-selection/candidates"},
			{Key: "evaluation_board", Label: "评估复盘", ActionType: "NAVIGATE", TargetRoute: "/stock-selection/evaluation"},
		},
	}, nil
}

func (r *MySQLGrowthRepo) AdminListStockSelectionRuns(status string, reviewStatus string, profileID string, page int, pageSize int) ([]model.StockSelectionRun, int, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	reviewStatus = strings.ToUpper(strings.TrimSpace(reviewStatus))
	profileID = strings.TrimSpace(profileID)
	offset := (page - 1) * pageSize

	args := make([]any, 0, 8)
	conditions := make([]string, 0, 3)
	if status != "" {
		conditions = append(conditions, "r.status = ?")
		args = append(args, status)
	}
	if reviewStatus != "" {
		conditions = append(conditions, "COALESCE(rv.review_status, '') = ?")
		args = append(args, reviewStatus)
	}
	if profileID != "" {
		conditions = append(conditions, "r.profile_id = ?")
		args = append(args, profileID)
	}
	filter := ""
	if len(conditions) > 0 {
		filter = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM stock_selection_runs r LEFT JOIN stock_selection_publish_reviews rv ON rv.run_id = r.run_id" + filter
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT
  r.run_id,
  DATE_FORMAT(r.trade_date, '%Y-%m-%d'),
  COALESCE(r.job_id, ''),
  r.profile_id,
  r.profile_version,
  COALESCE(r.template_id, ''),
  COALESCE(t.name, ''),
  COALESCE(r.market_regime, ''),
  r.selection_mode,
  r.universe_scope,
  r.status,
  COALESCE(r.result_summary, ''),
  COALESCE(CAST(r.warning_messages AS CHAR), ''),
  r.warning_count,
  r.universe_count,
  r.seed_count,
  r.candidate_count,
  r.selected_count,
  r.publish_count,
  COALESCE(CAST(r.context_meta AS CHAR), ''),
  COALESCE(CAST(r.template_snapshot AS CHAR), ''),
  COALESCE(CAST(r.compare_summary AS CHAR), ''),
  COALESCE(DATE_FORMAT(r.started_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.completed_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(r.created_by, ''),
  DATE_FORMAT(r.created_at, '%Y-%m-%dT%H:%i:%sZ'),
  DATE_FORMAT(r.updated_at, '%Y-%m-%dT%H:%i:%sZ'),
  COALESCE(rv.id, ''),
  COALESCE(rv.review_status, ''),
  COALESCE(rv.reviewer, ''),
  COALESCE(rv.review_note, ''),
  COALESCE(rv.override_reason, ''),
  COALESCE(rv.publish_id, ''),
  COALESCE(rv.publish_version, 0),
  COALESCE(CAST(rv.published_portfolio_snapshot AS CHAR), ''),
  COALESCE(DATE_FORMAT(rv.approved_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.rejected_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.updated_at, '%Y-%m-%dT%H:%i:%sZ'), '')
FROM stock_selection_runs r
LEFT JOIN stock_selection_profile_templates t ON t.id = r.template_id
LEFT JOIN stock_selection_publish_reviews rv ON rv.run_id = r.run_id` + filter + `
ORDER BY r.created_at DESC, r.run_id DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.StockSelectionRun, 0)
	for rows.Next() {
		item, scanErr := scanStockSelectionRunRow(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		r.attachStockSelectionRunSummary(&item)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateStockSelectionRun(input model.StockSelectionRunCreateRequest, operator string) (model.StockSelectionRun, error) {
	if r.strategyEngine == nil {
		return model.StockSelectionRun{}, sql.ErrNoRows
	}
	profile, err := r.resolveStockSelectionProfileForRun(input.ProfileID)
	if err != nil {
		return model.StockSelectionRun{}, err
	}
	template, err := r.resolveStockSelectionTemplateForRun(input.TemplateID, profile.TemplateID)
	if err != nil && err != sql.ErrNoRows {
		return model.StockSelectionRun{}, err
	}
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = "admin-stock-selection"
	}
	tradeDate := normalizeStockSelectionTradeDate(input.TradeDate)
	runID := newID("ssr")
	startedAt := time.Now().UTC()
	if err := r.insertStockSelectionRunStub(runID, tradeDate, *profile, template, operator, startedAt); err != nil {
		return model.StockSelectionRun{}, err
	}

	runtimeProfile, _ := r.ResolveActiveStrategyAgentProfile("STOCK")
	runtimeScenario, _ := r.ResolveActiveStrategyScenarioTemplate("STOCK")
	runtimePolicy, _ := r.ResolveActiveStrategyPublishPolicy("STOCK")
	payload := buildStockSelectionProfileJobPayload(tradeDate, *profile, template, input, runtimeProfile, runtimeScenario, runtimePolicy)
	jobRequest := map[string]any{
		"requested_by": operator,
		"payload":      payload,
	}

	accepted, err := r.strategyEngine.createStockSelectionJob(jobRequest)
	if err != nil {
		_ = r.markStockSelectionRunFailed(runID, "", err.Error())
		return model.StockSelectionRun{}, err
	}
	if err := r.bindStockSelectionRunJob(runID, accepted.JobID); err != nil {
		return model.StockSelectionRun{}, err
	}

	report, waitErr := r.strategyEngine.waitForStockSelectionJob(accepted.JobID)
	jobRecord, jobErr := r.strategyEngine.getJobRecord(accepted.JobID)
	if jobErr == nil {
		_ = r.upsertStrategyEngineJobSnapshot(jobRecord)
	}
	if waitErr != nil {
		failureSummary := waitErr.Error()
		if jobErr == nil && strings.TrimSpace(jobRecord.ErrorMessage) != "" {
			failureSummary = jobRecord.ErrorMessage
		}
		_ = r.markStockSelectionRunFailed(runID, accepted.JobID, failureSummary)
		return model.StockSelectionRun{}, waitErr
	}
	if jobErr == nil && jobRecord.Result != nil && strings.TrimSpace(jobRecord.ResultSummary) == "" {
		jobRecord.ResultSummary = jobRecord.Result.Summary
	}
	if err := r.persistCompletedStockSelectionRun(runID, accepted.JobID, report, jobRecord, startedAt); err != nil {
		return model.StockSelectionRun{}, err
	}
	return r.AdminGetStockSelectionRun(runID)
}

func (r *MySQLGrowthRepo) AdminGetStockSelectionRun(runID string) (model.StockSelectionRun, error) {
	rows, err := r.db.Query(`
SELECT
  r.run_id,
  DATE_FORMAT(r.trade_date, '%Y-%m-%d'),
  COALESCE(r.job_id, ''),
  r.profile_id,
  r.profile_version,
  COALESCE(r.template_id, ''),
  COALESCE(t.name, ''),
  COALESCE(r.market_regime, ''),
  r.selection_mode,
  r.universe_scope,
  r.status,
  COALESCE(r.result_summary, ''),
  COALESCE(CAST(r.warning_messages AS CHAR), ''),
  r.warning_count,
  r.universe_count,
  r.seed_count,
  r.candidate_count,
  r.selected_count,
  r.publish_count,
  COALESCE(CAST(r.context_meta AS CHAR), ''),
  COALESCE(CAST(r.template_snapshot AS CHAR), ''),
  COALESCE(CAST(r.compare_summary AS CHAR), ''),
  COALESCE(DATE_FORMAT(r.started_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(r.completed_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(r.created_by, ''),
  DATE_FORMAT(r.created_at, '%Y-%m-%dT%H:%i:%sZ'),
  DATE_FORMAT(r.updated_at, '%Y-%m-%dT%H:%i:%sZ'),
  COALESCE(rv.id, ''),
  COALESCE(rv.review_status, ''),
  COALESCE(rv.reviewer, ''),
  COALESCE(rv.review_note, ''),
  COALESCE(rv.override_reason, ''),
  COALESCE(rv.publish_id, ''),
  COALESCE(rv.publish_version, 0),
  COALESCE(CAST(rv.published_portfolio_snapshot AS CHAR), ''),
  COALESCE(DATE_FORMAT(rv.approved_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.rejected_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.updated_at, '%Y-%m-%dT%H:%i:%sZ'), '')
FROM stock_selection_runs r
LEFT JOIN stock_selection_profile_templates t ON t.id = r.template_id
LEFT JOIN stock_selection_publish_reviews rv ON rv.run_id = r.run_id
WHERE r.run_id = ?`, strings.TrimSpace(runID))
	if err != nil {
		return model.StockSelectionRun{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return model.StockSelectionRun{}, sql.ErrNoRows
	}
	item, err := scanStockSelectionRunRow(rows)
	if err != nil {
		return model.StockSelectionRun{}, err
	}
	stageLogs, err := r.loadStockSelectionRunStageLogs(item.RunID)
	if err != nil {
		return model.StockSelectionRun{}, err
	}
	item.StageLogs = stageLogs
	item.StageDurationsMS = buildStageDurationMap(stageLogs)
	r.attachStockSelectionRunSummary(&item)
	return item, nil
}

func (r *MySQLGrowthRepo) AdminListStockSelectionRunCandidates(runID string) ([]model.StockSelectionCandidateSnapshot, error) {
	_ = r.ensureStockSelectionRunEvaluations(runID)
	rows, err := r.db.Query(`
SELECT id, run_id, symbol, name, stage, quant_score, risk_level, selected, rank_no, COALESCE(reason_summary, ''), COALESCE(CAST(factor_breakdown_json AS CHAR), ''), DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM stock_selection_run_candidates
WHERE run_id = ?
ORDER BY FIELD(stage, 'UNIVERSE', 'SEED_POOL', 'CANDIDATE_POOL', 'PORTFOLIO'), rank_no ASC, symbol ASC`, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StockSelectionCandidateSnapshot, 0)
	for rows.Next() {
		var item model.StockSelectionCandidateSnapshot
		var factorJSON string
		var selected bool
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.Symbol,
			&item.Name,
			&item.Stage,
			&item.QuantScore,
			&item.RiskLevel,
			&selected,
			&item.Rank,
			&item.ReasonSummary,
			&factorJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.Selected = selected
		item.FactorBreakdownJSON = parseJSONMap(factorJSON)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	r.enrichStockSelectionCandidateSnapshots(runID, items)
	return items, nil
}

func (r *MySQLGrowthRepo) AdminListStockSelectionRunPortfolio(runID string) ([]model.StockSelectionPortfolioEntry, error) {
	_ = r.ensureStockSelectionRunEvaluations(runID)
	rows, err := r.db.Query(`
SELECT id, run_id, symbol, name, rank_no, quant_score, risk_level, COALESCE(weight_suggestion, ''), COALESCE(reason_summary, ''), COALESCE(CAST(factor_breakdown_json AS CHAR), ''), DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM stock_selection_run_portfolio
WHERE run_id = ?
ORDER BY rank_no ASC, symbol ASC`, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StockSelectionPortfolioEntry, 0)
	for rows.Next() {
		var item model.StockSelectionPortfolioEntry
		var factorJSON string
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.Symbol,
			&item.Name,
			&item.Rank,
			&item.QuantScore,
			&item.RiskLevel,
			&item.WeightSuggestion,
			&item.ReasonSummary,
			&factorJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.FactorBreakdownJSON = parseJSONMap(factorJSON)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	r.enrichStockSelectionPortfolioEntries(runID, items)
	return items, nil
}

func (r *MySQLGrowthRepo) AdminListStockSelectionReviews(status string, page int, pageSize int) ([]model.StockSelectionPublishReview, int, error) {
	status = strings.ToUpper(strings.TrimSpace(status))
	offset := (page - 1) * pageSize
	args := make([]any, 0, 3)
	filter := ""
	if status != "" {
		filter = " WHERE review_status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM stock_selection_publish_reviews"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, run_id, review_status, COALESCE(reviewer, ''), COALESCE(review_note, ''), COALESCE(override_reason, ''), COALESCE(publish_id, ''), COALESCE(publish_version, 0),
       COALESCE(DATE_FORMAT(approved_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
       COALESCE(DATE_FORMAT(rejected_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
       COALESCE(DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
       COALESCE(DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ'), '')
FROM stock_selection_publish_reviews` + filter + `
ORDER BY updated_at DESC, id DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.StockSelectionPublishReview, 0)
	for rows.Next() {
		var item model.StockSelectionPublishReview
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.ReviewStatus,
			&item.Reviewer,
			&item.ReviewNote,
			&item.OverrideReason,
			&item.PublishID,
			&item.PublishVersion,
			&item.ApprovedAt,
			&item.RejectedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminApproveStockSelectionReview(runID string, operator string, reviewNote string, force bool, overrideReason string) (model.StockSelectionPublishReview, error) {
	run, err := r.AdminGetStockSelectionRun(runID)
	if err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	if strings.TrimSpace(run.JobID) == "" {
		return model.StockSelectionPublishReview{}, sql.ErrNoRows
	}
	published, err := r.AdminPublishStrategyEngineJob(run.JobID, operator, force, overrideReason)
	if err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	payloads, err := toStrategyEngineStockPublishPayloads(published.PublishPayloads)
	if err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	if _, err := r.persistStrategyEngineStockRecommendations(payloads); err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	jobRecord, jobErr := r.AdminGetStrategyEngineJob(run.JobID)
	publishCount := run.PublishCount
	if jobErr == nil && jobRecord.PublishCount > publishCount {
		publishCount = jobRecord.PublishCount
	}
	if publishCount == 0 {
		publishCount = 1
	}
	portfolioEntries, err := r.AdminListStockSelectionRunPortfolio(runID)
	if err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	tx, err := r.db.Begin()
	if err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	_, err = tx.Exec(`
UPDATE stock_selection_publish_reviews
SET review_status = 'APPROVED',
    reviewer = ?,
    review_note = ?,
    override_reason = ?,
    publish_id = ?,
    publish_version = ?,
    published_portfolio_snapshot = ?,
    approved_at = NOW(),
    rejected_at = NULL,
    updated_at = NOW()
WHERE run_id = ?`,
		operator,
		strings.TrimSpace(reviewNote),
		strings.TrimSpace(overrideReason),
		published.PublishID,
		published.Version,
		stockSelectionMustJSON(buildPublishedPortfolioSnapshot(portfolioEntries)),
		runID,
	)
	if err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	_, err = tx.Exec("UPDATE stock_selection_runs SET publish_count = ?, updated_at = NOW() WHERE run_id = ?", publishCount, runID)
	if err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	return r.getStockSelectionReview(runID)
}

func (r *MySQLGrowthRepo) AdminRejectStockSelectionReview(runID string, operator string, reviewNote string) (model.StockSelectionPublishReview, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	_, err = tx.Exec(`
UPDATE stock_selection_publish_reviews
SET review_status = 'REJECTED',
    reviewer = ?,
    review_note = ?,
    override_reason = '',
    rejected_at = NOW(),
    approved_at = NULL,
    updated_at = NOW()
WHERE run_id = ?`,
		operator,
		strings.TrimSpace(reviewNote),
		strings.TrimSpace(runID),
	)
	if err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	return r.getStockSelectionReview(runID)
}

func (r *MySQLGrowthRepo) resolveStockSelectionProfileForRun(profileID string) (*model.StockSelectionProfile, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID != "" {
		item, err := r.getStockSelectionProfile(profileID)
		if err != nil {
			return nil, err
		}
		return &item, nil
	}
	return r.getDefaultStockSelectionProfile()
}

func normalizeStockSelectionTradeDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		return raw
	}
	return time.Now().Format("2006-01-02")
}

func buildStockSelectionProfileJobPayload(
	tradeDate string,
	profile model.StockSelectionProfile,
	template *model.StockSelectionProfileTemplate,
	input model.StockSelectionRunCreateRequest,
	runtimeProfile *model.StrategyAgentProfile,
	runtimeScenario *model.StrategyScenarioTemplate,
	runtimePolicy *model.StrategyPublishPolicy,
) map[string]any {
	universeConfig := mergeStockSelectionConfigMaps(nil, profile.UniverseConfig)
	seedConfig := mergeStockSelectionConfigMaps(nil, profile.SeedMiningConfig)
	factorConfig := mergeStockSelectionConfigMaps(nil, profile.FactorConfig)
	portfolioConfig := mergeStockSelectionConfigMaps(nil, profile.PortfolioConfig)
	publishConfig := mergeStockSelectionConfigMaps(nil, profile.PublishConfig)
	if template != nil {
		universeConfig = mergeStockSelectionConfigMaps(template.UniverseDefaults, profile.UniverseConfig)
		seedConfig = mergeStockSelectionConfigMaps(template.SeedDefaults, profile.SeedMiningConfig)
		factorConfig = mergeStockSelectionConfigMaps(template.FactorDefaults, profile.FactorConfig)
		portfolioConfig = mergeStockSelectionConfigMaps(template.PortfolioDefaults, profile.PortfolioConfig)
		publishConfig = mergeStockSelectionConfigMaps(template.PublishDefaults, profile.PublishConfig)
	}

	payload := map[string]any{
		"trade_date":                  tradeDate,
		"selection_mode":              profile.SelectionModeDefault,
		"profile_id":                  profile.ID,
		"universe_scope":              profile.UniverseScope,
		"dry_run":                     input.DryRun,
		"compare_with_last_published": input.CompareWithLastPublished,
	}
	if template != nil {
		payload["template_id"] = template.ID
		payload["template_key"] = template.TemplateKey
		payload["template_name"] = template.Name
		payload["template_snapshot"] = map[string]any{
			"id":                 template.ID,
			"template_key":       template.TemplateKey,
			"name":               template.Name,
			"market_regime_bias": template.MarketRegimeBias,
		}
	}
	if scope := stringValue(universeConfig["universe_scope"]); scope != "" {
		payload["universe_scope"] = scope
		payload["market_scope"] = scope
	}
	if limit := intValue(portfolioConfig["limit"]); limit > 0 {
		payload["limit"] = limit
	}
	if minScore := floatValue(portfolioConfig["min_score"]); minScore > 0 {
		payload["min_score"] = minScore
	}
	if maxRisk := strings.ToUpper(stringValue(portfolioConfig["max_risk_level"])); maxRisk != "" {
		payload["max_risk_level"] = maxRisk
	}
	if lookbackDays := intValue(factorConfig["lookback_days"]); lookbackDays > 0 {
		payload["lookback_days"] = lookbackDays
	}
	if listingDays := intValue(universeConfig["min_listing_days"]); listingDays > 0 {
		payload["min_listing_days"] = listingDays
	}
	if avgTurnover := floatValue(universeConfig["min_avg_turnover"]); avgTurnover > 0 {
		payload["min_avg_turnover"] = avgTurnover
	}
	if seedSymbols := stockSelectionStringSlice(seedConfig["seed_symbols"]); len(seedSymbols) > 0 {
		payload["seed_symbols"] = seedSymbols
	}
	if debugSymbols := stockSelectionStringSlice(seedConfig["debug_seed_symbols"]); len(debugSymbols) > 0 {
		payload["debug_seed_symbols"] = debugSymbols
	}
	if excluded := stockSelectionStringSlice(universeConfig["excluded_symbols"]); len(excluded) > 0 {
		payload["excluded_symbols"] = excluded
	}
	copyStockSelectionPayloadFields(payload, universeConfig, []string{"price_min", "price_max", "volatility_min", "volatility_max", "industry_whitelist", "industry_blacklist", "sector_whitelist", "sector_blacklist", "theme_whitelist", "theme_blacklist"})
	copyStockSelectionPayloadFields(payload, seedConfig, []string{"bucket_limit", "seed_pool_cap", "candidate_pool_limit", "trend_bias", "money_flow_bias", "quality_bias", "event_bias", "resonance_bias"})
	copyStockSelectionPayloadFields(payload, factorConfig, []string{"quant_weight", "event_weight", "resonance_weight", "liquidity_risk_weight"})
	copyStockSelectionPayloadFields(payload, portfolioConfig, []string{"watchlist_limit", "max_symbol_per_bucket", "max_symbols_per_sector"})
	copyStockSelectionPayloadFields(payload, publishConfig, []string{"review_required", "allow_auto_publish"})
	if runtimeProfile != nil {
		payload["enabled_agents"] = runtimeProfile.EnabledAgents
		payload["positive_threshold"] = runtimeProfile.PositiveThreshold
		payload["negative_threshold"] = runtimeProfile.NegativeThreshold
		payload["allow_veto"] = runtimeProfile.AllowVeto
	}
	if runtimeScenario != nil && len(runtimeScenario.Items) > 0 {
		payload["scenario_templates"] = runtimeScenario.Items
	}
	attachStrategyConfigPayload(payload, nil, runtimeProfile, runtimeScenario, runtimePolicy)
	return payload
}

func (r *MySQLGrowthRepo) insertStockSelectionRunStub(runID string, tradeDate string, profile model.StockSelectionProfile, template *model.StockSelectionProfileTemplate, operator string, startedAt time.Time) error {
	templateID := ""
	if template != nil {
		templateID = template.ID
	}
	_, err := r.db.Exec(`
INSERT INTO stock_selection_runs (
  run_id, trade_date, profile_id, profile_version, template_id, selection_mode, universe_scope, status,
  result_summary, warning_count, universe_count, seed_count, candidate_count, selected_count, publish_count,
  context_meta, started_at, completed_at, created_by, created_at, updated_at
) VALUES (?, ?, ?, ?, NULLIF(?, ''), ?, ?, 'RUNNING', '', 0, 0, 0, 0, 0, 0, NULL, ?, NULL, ?, NOW(), NOW())`,
		runID,
		tradeDate,
		profile.ID,
		profile.CurrentVersion,
		templateID,
		profile.SelectionModeDefault,
		profile.UniverseScope,
		startedAt,
		operator,
	)
	return err
}

func (r *MySQLGrowthRepo) bindStockSelectionRunJob(runID string, jobID string) error {
	_, err := r.db.Exec("UPDATE stock_selection_runs SET job_id = ?, updated_at = NOW() WHERE run_id = ?", strings.TrimSpace(jobID), strings.TrimSpace(runID))
	return err
}

func (r *MySQLGrowthRepo) markStockSelectionRunFailed(runID string, jobID string, summary string) error {
	_, err := r.db.Exec(`
UPDATE stock_selection_runs
SET job_id = NULLIF(?, ''),
    status = 'FAILED',
    result_summary = ?,
    completed_at = NOW(),
    updated_at = NOW()
WHERE run_id = ?`,
		strings.TrimSpace(jobID),
		strings.TrimSpace(summary),
		strings.TrimSpace(runID),
	)
	return err
}

func (r *MySQLGrowthRepo) persistCompletedStockSelectionRun(
	runID string,
	jobID string,
	report strategyEngineStockSelectionReport,
	jobRecord model.StrategyEngineJobRecord,
	startedAt time.Time,
) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	warnings := []string{}
	resultSummary := strings.TrimSpace(report.ReportSummary)
	if jobRecord.Result != nil {
		warnings = append(warnings, jobRecord.Result.Warnings...)
		if resultSummary == "" {
			resultSummary = strings.TrimSpace(jobRecord.Result.Summary)
		}
	}
	if resultSummary == "" {
		resultSummary = fmt.Sprintf("stock-selection completed with %d publish-ready candidates", report.SelectedCount)
	}
	contextMeta := buildStockSelectionRunContextMeta(report)
	if _, err = tx.Exec(`
UPDATE stock_selection_runs
SET job_id = ?,
    status = 'SUCCEEDED',
    result_summary = ?,
    warning_messages = ?,
    warning_count = ?,
    universe_count = ?,
    seed_count = ?,
    candidate_count = ?,
    selected_count = ?,
    publish_count = 0,
    market_regime = ?,
    context_meta = ?,
    template_snapshot = ?,
    compare_summary = ?,
    started_at = ?,
    completed_at = NOW(),
    updated_at = NOW()
WHERE run_id = ?`,
		jobID,
		resultSummary,
		stockSelectionMustJSON(warnings),
		len(warnings),
		report.StageCounts["UNIVERSE"],
		report.StageCounts["SEED_POOL"],
		report.StageCounts["CANDIDATE_POOL"],
		report.SelectedCount,
		strings.TrimSpace(report.MarketRegime),
		stockSelectionMustJSON(contextMeta),
		stockSelectionMustJSON(report.TemplateSnapshot),
		stockSelectionMustJSON(buildStockSelectionCompareSummary(nil, report.PortfolioEntries)),
		startedAt,
		runID,
	); err != nil {
		return err
	}

	if _, err = tx.Exec("DELETE FROM stock_selection_run_stage_logs WHERE run_id = ?", runID); err != nil {
		return err
	}
	for _, stage := range report.StageLogs {
		if _, err = tx.Exec(`
INSERT INTO stock_selection_run_stage_logs (id, run_id, stage_key, stage_order, status, input_count, output_count, duration_ms, detail_message, payload_snapshot, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			newID("sslg"),
			runID,
			stage.StageKey,
			stage.StageOrder,
			stage.Status,
			stage.InputCount,
			stage.OutputCount,
			stage.DurationMS,
			stage.DetailMessage,
			stockSelectionMustJSON(stage.PayloadSnapshot),
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec("DELETE FROM stock_selection_run_candidates WHERE run_id = ?", runID); err != nil {
		return err
	}
	for _, item := range report.CandidateSnapshots {
		if _, err = tx.Exec(`
INSERT INTO stock_selection_run_candidates (id, run_id, symbol, name, stage, quant_score, risk_level, selected, rank_no, reason_summary, factor_breakdown_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			newID("ssrc"),
			runID,
			item.Symbol,
			item.Name,
			item.Stage,
			item.QuantScore,
			strings.ToUpper(strings.TrimSpace(item.RiskLevel)),
			item.Selected,
			item.Rank,
			item.ReasonSummary,
			stockSelectionMustJSON(item.FactorBreakdownJSON),
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec("DELETE FROM stock_selection_run_portfolio WHERE run_id = ?", runID); err != nil {
		return err
	}
	for _, item := range report.PortfolioEntries {
		if _, err = tx.Exec(`
INSERT INTO stock_selection_run_portfolio (id, run_id, symbol, name, rank_no, quant_score, risk_level, weight_suggestion, reason_summary, factor_breakdown_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			newID("ssrp"),
			runID,
			item.Symbol,
			item.Name,
			item.Rank,
			item.QuantScore,
			strings.ToUpper(strings.TrimSpace(item.RiskLevel)),
			item.WeightSuggestion,
			item.ReasonSummary,
			stockSelectionMustJSON(item.FactorBreakdownJSON),
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec("DELETE FROM stock_selection_run_evidence WHERE run_id = ?", runID); err != nil {
		return err
	}
	for _, item := range report.EvidenceRecords {
		if _, err = tx.Exec(`
INSERT INTO stock_selection_run_evidence (
  id, run_id, symbol, stage, name, portfolio_role, evidence_summary,
  evidence_cards_json, positive_reasons_json, veto_reasons_json, theme_tags_json, sector_tags_json, risk_flags_json,
  created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			newID("ssev"),
			runID,
			item.Symbol,
			item.Stage,
			item.Name,
			item.PortfolioRole,
			item.EvidenceSummary,
			stockSelectionMustJSON(item.EvidenceCards),
			stockSelectionMustJSON(item.PositiveReasons),
			stockSelectionMustJSON(item.VetoReasons),
			stockSelectionMustJSON(item.ThemeTags),
			stockSelectionMustJSON(item.SectorTags),
			stockSelectionMustJSON(item.RiskFlags),
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec("DELETE FROM stock_selection_run_evaluations WHERE run_id = ?", runID); err != nil {
		return err
	}
	for _, item := range report.EvaluationRecords {
		if _, err = tx.Exec(`
INSERT INTO stock_selection_run_evaluations (
  id, run_id, symbol, horizon_day, evaluation_scope, name, entry_date, exit_date,
  entry_price, exit_price, return_pct, excess_return_pct, max_drawdown_pct, hit_flag, benchmark_symbol,
  created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			newID("ssev"),
			runID,
			item.Symbol,
			item.HorizonDay,
			item.EvaluationScope,
			item.Name,
			item.EntryDate,
			item.ExitDate,
			item.EntryPrice,
			item.ExitPrice,
			item.ReturnPct,
			item.ExcessReturnPct,
			item.MaxDrawdownPct,
			item.HitFlag,
			item.BenchmarkSymbol,
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec(`
INSERT INTO stock_selection_publish_reviews (id, run_id, review_status, reviewer, review_note, override_reason, publish_id, publish_version, approved_at, rejected_at, created_at, updated_at)
VALUES (?, ?, 'PENDING', '', '', '', '', 0, NULL, NULL, NOW(), NOW())
ON DUPLICATE KEY UPDATE review_status = VALUES(review_status), reviewer = '', review_note = '', override_reason = '', publish_id = '', publish_version = 0, approved_at = NULL, rejected_at = NULL, updated_at = NOW()`,
		newID("ssrv"),
		runID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *MySQLGrowthRepo) loadStockSelectionRunStageLogs(runID string) ([]model.StockSelectionRunStageLog, error) {
	rows, err := r.db.Query(`
SELECT id, run_id, stage_key, stage_order, status, input_count, output_count, duration_ms, COALESCE(detail_message, ''), COALESCE(CAST(payload_snapshot AS CHAR), ''), DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM stock_selection_run_stage_logs
WHERE run_id = ?
ORDER BY stage_order ASC, id ASC`, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.StockSelectionRunStageLog, 0)
	for rows.Next() {
		var item model.StockSelectionRunStageLog
		var payloadJSON string
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.StageKey,
			&item.StageOrder,
			&item.Status,
			&item.InputCount,
			&item.OutputCount,
			&item.DurationMS,
			&item.DetailMessage,
			&payloadJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.PayloadSnapshot = parseJSONMap(payloadJSON)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) getStockSelectionReview(runID string) (model.StockSelectionPublishReview, error) {
	var item model.StockSelectionPublishReview
	var snapshotJSON string
	err := r.db.QueryRow(`
SELECT id, run_id, review_status, COALESCE(reviewer, ''), COALESCE(review_note, ''), COALESCE(override_reason, ''), COALESCE(publish_id, ''), COALESCE(publish_version, 0), COALESCE(CAST(published_portfolio_snapshot AS CHAR), ''),
       COALESCE(DATE_FORMAT(approved_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
       COALESCE(DATE_FORMAT(rejected_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
       COALESCE(DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
       COALESCE(DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ'), '')
FROM stock_selection_publish_reviews
WHERE run_id = ?`, strings.TrimSpace(runID)).Scan(
		&item.ID,
		&item.RunID,
		&item.ReviewStatus,
		&item.Reviewer,
		&item.ReviewNote,
		&item.OverrideReason,
		&item.PublishID,
		&item.PublishVersion,
		&snapshotJSON,
		&item.ApprovedAt,
		&item.RejectedAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return model.StockSelectionPublishReview{}, err
	}
	item.PublishedPortfolioSnapshot = parseJSONMapSlice(snapshotJSON)
	return item, nil
}

func (r *MySQLGrowthRepo) getLatestStockSelectionRun(status string) (*model.StockSelectionRun, error) {
	items, _, err := r.AdminListStockSelectionRuns(status, "", "", 1, 1)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	cloned := items[0]
	return &cloned, nil
}

func (r *MySQLGrowthRepo) attachStockSelectionRunSummary(item *model.StockSelectionRun) {
	if item == nil {
		return
	}
	item.StageCounts = map[string]int{
		"UNIVERSE":       item.UniverseCount,
		"SEED_POOL":      item.SeedCount,
		"CANDIDATE_POOL": item.CandidateCount,
		"PORTFOLIO":      item.SelectedCount,
	}
	if strings.TrimSpace(item.JobID) != "" {
		job, err := r.AdminGetStrategyEngineJob(item.JobID)
		if err == nil {
			item.PublishCount = maxInt(item.PublishCount, job.PublishCount)
			item.LatestPublishID = job.LatestPublishID
			item.LatestPublishVersion = job.LatestPublishVersion
			item.LatestPublishAt = job.LatestPublishAt
			item.LatestPublishMode = job.LatestPublishMode
			item.LatestPublishSource = job.LatestPublishSource
		}
	}
}

func scanStockSelectionRunRow(scanner interface{ Scan(dest ...any) error }) (model.StockSelectionRun, error) {
	var item model.StockSelectionRun
	var warningJSON, contextJSON, templateSnapshotJSON, compareSummaryJSON string
	var review model.StockSelectionPublishReview
	var portfolioSnapshotJSON string
	err := scanner.Scan(
		&item.RunID,
		&item.TradeDate,
		&item.JobID,
		&item.ProfileID,
		&item.ProfileVersion,
		&item.TemplateID,
		&item.TemplateName,
		&item.MarketRegime,
		&item.SelectionMode,
		&item.UniverseScope,
		&item.Status,
		&item.ResultSummary,
		&warningJSON,
		&item.WarningCount,
		&item.UniverseCount,
		&item.SeedCount,
		&item.CandidateCount,
		&item.SelectedCount,
		&item.PublishCount,
		&contextJSON,
		&templateSnapshotJSON,
		&compareSummaryJSON,
		&item.StartedAt,
		&item.CompletedAt,
		&item.CreatedBy,
		&item.CreatedAt,
		&item.UpdatedAt,
		&review.ID,
		&review.ReviewStatus,
		&review.Reviewer,
		&review.ReviewNote,
		&review.OverrideReason,
		&review.PublishID,
		&review.PublishVersion,
		&portfolioSnapshotJSON,
		&review.ApprovedAt,
		&review.RejectedAt,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		return model.StockSelectionRun{}, err
	}
	item.WarningMessages = parseJSONStringList(warningJSON)
	item.ContextMeta = parseJSONMap(contextJSON)
	item.TemplateSnapshot = parseJSONMap(templateSnapshotJSON)
	item.CompareSummary = parseJSONMap(compareSummaryJSON)
	if strings.TrimSpace(review.ID) != "" {
		review.RunID = item.RunID
		review.PublishedPortfolioSnapshot = parseJSONMapSlice(portfolioSnapshotJSON)
		item.Review = &review
		item.ReviewStatus = review.ReviewStatus
	}
	return item, nil
}

func buildStageDurationMap(items []model.StockSelectionRunStageLog) map[string]int64 {
	result := make(map[string]int64, len(items))
	for _, item := range items {
		result[item.StageKey] = item.DurationMS
	}
	return result
}

func toStrategyEngineStockPublishPayloads(items []map[string]any) ([]strategyEngineStockPublishPayload, error) {
	if len(items) == 0 {
		return nil, nil
	}
	body, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	var payloads []strategyEngineStockPublishPayload
	if err := json.Unmarshal(body, &payloads); err != nil {
		return nil, err
	}
	return payloads, nil
}

func parseJSONMap(text string) map[string]any {
	text = strings.TrimSpace(text)
	if text == "" || text == "null" {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return map[string]any{}
	}
	if payload == nil {
		return map[string]any{}
	}
	return payload
}

func buildStockSelectionRunContextMeta(report strategyEngineStockSelectionReport) map[string]any {
	result := make(map[string]any, len(report.ContextMeta)+4)
	for key, value := range report.ContextMeta {
		result[key] = value
	}
	if strings.TrimSpace(report.GraphSummary) != "" {
		result["graph_summary"] = strings.TrimSpace(report.GraphSummary)
	}
	if strings.TrimSpace(report.GraphSnapshotID) != "" {
		result["graph_snapshot_id"] = strings.TrimSpace(report.GraphSnapshotID)
	}
	if len(report.RelatedEntities) > 0 {
		result["related_entities"] = report.RelatedEntities
	}
	if len(report.MemoryFeedback) > 0 {
		result["memory_feedback"] = report.MemoryFeedback
	}
	return result
}

func parseJSONStringList(text string) []string {
	text = strings.TrimSpace(text)
	if text == "" || text == "null" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		return nil
	}
	return items
}

func stockSelectionMustJSON(value any) string {
	body, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(body)
}

func stringValue(raw any) string {
	if raw == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprintf("%v", raw))
}

func boolValue(raw any) bool {
	switch value := raw.(type) {
	case bool:
		return value
	case string:
		return strings.EqualFold(strings.TrimSpace(value), "true")
	case float64:
		return value != 0
	case int:
		return value != 0
	default:
		return false
	}
}

func stockSelectionMapValue(raw any) map[string]any {
	if raw == nil {
		return map[string]any{}
	}
	if item, ok := raw.(map[string]any); ok {
		return item
	}
	body, err := json.Marshal(raw)
	if err != nil {
		return map[string]any{}
	}
	return parseJSONMap(string(body))
}

func intValue(raw any) int {
	switch value := raw.(type) {
	case int:
		return value
	case int32:
		return int(value)
	case int64:
		return int(value)
	case float64:
		return int(value)
	case string:
		parsed, _ := strconv.Atoi(strings.TrimSpace(value))
		return parsed
	default:
		return 0
	}
}

func floatValue(raw any) float64 {
	switch value := raw.(type) {
	case float64:
		return value
	case float32:
		return float64(value)
	case int:
		return float64(value)
	case int64:
		return float64(value)
	case string:
		parsed, _ := strconv.ParseFloat(strings.TrimSpace(value), 64)
		return parsed
	default:
		return 0
	}
}

func stockSelectionStringSlice(raw any) []string {
	items, ok := raw.([]any)
	if ok {
		result := make([]string, 0, len(items))
		for _, item := range items {
			text := strings.ToUpper(strings.TrimSpace(fmt.Sprintf("%v", item)))
			if text == "" {
				continue
			}
			result = append(result, text)
		}
		return result
	}
	stringsRaw, ok := raw.([]string)
	if ok {
		result := make([]string, 0, len(stringsRaw))
		for _, item := range stringsRaw {
			text := strings.ToUpper(strings.TrimSpace(item))
			if text == "" {
				continue
			}
			result = append(result, text)
		}
		return result
	}
	return nil
}
