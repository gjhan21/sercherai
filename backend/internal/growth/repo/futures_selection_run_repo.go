package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminGetFuturesSelectionOverview() (model.AdminFuturesSelectionOverview, error) {
	defaultProfile, err := r.getDefaultFuturesSelectionProfile()
	if err != nil && err != sql.ErrNoRows {
		return model.AdminFuturesSelectionOverview{}, err
	}
	latestRun, err := r.getLatestFuturesSelectionRun("")
	if err != nil && err != sql.ErrNoRows {
		return model.AdminFuturesSelectionOverview{}, err
	}
	latestSuccessRun, err := r.getLatestFuturesSelectionRun("SUCCEEDED")
	if err != nil && err != sql.ErrNoRows {
		return model.AdminFuturesSelectionOverview{}, err
	}

	dataFreshness := map[string]any{}
	latestTradeDate := ""
	if latestSuccessRun != nil {
		dataFreshness = latestSuccessRun.ContextMeta
		dataFreshness["trade_date"] = latestSuccessRun.TradeDate
		dataFreshness["completed_at"] = latestSuccessRun.CompletedAt
		latestTradeDate = latestSuccessRun.TradeDate
	} else if latestRun != nil {
		latestTradeDate = latestRun.TradeDate
	}

	var pendingReviewCount int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM futures_selection_publish_reviews WHERE review_status = 'PENDING'").Scan(&pendingReviewCount); err != nil {
		return model.AdminFuturesSelectionOverview{}, err
	}
	latestApprovedPortfolio, err := r.loadLatestApprovedFuturesContractEntries()
	if err != nil {
		return model.AdminFuturesSelectionOverview{}, err
	}
	marketRegime := ""
	if latestSuccessRun != nil {
		marketRegime = latestSuccessRun.MarketRegime
	} else if latestRun != nil {
		marketRegime = latestRun.MarketRegime
	}

	return model.AdminFuturesSelectionOverview{
		DefaultProfile:          defaultProfile,
		LatestTradeDate:         latestTradeDate,
		LatestRun:               latestRun,
		LatestSuccessRun:        latestSuccessRun,
		LatestApprovedPortfolio: latestApprovedPortfolio,
		MarketRegime:            marketRegime,
		DataFreshness:           dataFreshness,
		PendingReviewCount:      pendingReviewCount,
		QuickActions: []model.AdminStockSelectionQuickAction{
			{Key: "run_now", Label: "立即运行", ActionType: "RUN", TargetRoute: "/futures-selection/runs"},
			{Key: "open_runs", Label: "运行中心", ActionType: "NAVIGATE", TargetRoute: "/futures-selection/runs"},
		},
	}, nil
}

func (r *MySQLGrowthRepo) AdminListFuturesSelectionRuns(status string, reviewStatus string, profileID string, page int, pageSize int) ([]model.FuturesSelectionRun, int, error) {
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
	countQuery := "SELECT COUNT(*) FROM futures_selection_runs r LEFT JOIN futures_selection_publish_reviews rv ON rv.run_id = r.run_id" + filter
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
  COALESCE(r.style, ''),
  COALESCE(r.contract_scope, ''),
  r.status,
  COALESCE(r.result_summary, ''),
  COALESCE(CAST(r.warning_messages AS CHAR), ''),
  r.warning_count,
  r.universe_count,
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
  COALESCE(CAST(rv.published_contract_snapshot AS CHAR), ''),
  COALESCE(DATE_FORMAT(rv.approved_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.rejected_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.updated_at, '%Y-%m-%dT%H:%i:%sZ'), '')
FROM futures_selection_runs r
LEFT JOIN futures_selection_profile_templates t ON t.id = r.template_id
LEFT JOIN futures_selection_publish_reviews rv ON rv.run_id = r.run_id` + filter + `
ORDER BY r.created_at DESC, r.run_id DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.FuturesSelectionRun, 0)
	for rows.Next() {
		item, scanErr := scanFuturesSelectionRunRow(rows)
		if scanErr != nil {
			return nil, 0, scanErr
		}
		r.attachFuturesSelectionRunSummary(&item)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateFuturesSelectionRun(input model.FuturesSelectionRunCreateRequest, operator string) (model.FuturesSelectionRun, error) {
	if r.strategyEngine == nil {
		return model.FuturesSelectionRun{}, sql.ErrNoRows
	}
	profile, err := r.resolveFuturesSelectionProfileForRun(input.ProfileID)
	if err != nil {
		return model.FuturesSelectionRun{}, err
	}
	template, err := r.resolveFuturesSelectionTemplateForRun(input.TemplateID, profile.TemplateID)
	if err != nil && err != sql.ErrNoRows {
		return model.FuturesSelectionRun{}, err
	}
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = "admin-futures-selection"
	}
	tradeDate := normalizeFuturesSelectionTradeDate(input.TradeDate)
	runID := newID("fsr")
	startedAt := time.Now().UTC()
	if err := r.insertFuturesSelectionRunStub(runID, tradeDate, *profile, template, operator, startedAt); err != nil {
		return model.FuturesSelectionRun{}, err
	}

	runtimeProfile, _ := r.ResolveActiveStrategyAgentProfile("FUTURES")
	runtimeScenario, _ := r.ResolveActiveStrategyScenarioTemplate("FUTURES")
	runtimePolicy, _ := r.ResolveActiveStrategyPublishPolicy("FUTURES")
	payload := buildFuturesSelectionJobPayload(runID, tradeDate, *profile, template, input, runtimeProfile, runtimeScenario, runtimePolicy)
	jobRequest := map[string]any{
		"requested_by": operator,
		"payload":      payload,
	}

	accepted, err := r.strategyEngine.createFuturesStrategyJob(jobRequest)
	if err != nil {
		_ = r.markFuturesSelectionRunFailed(runID, "", err.Error())
		return model.FuturesSelectionRun{}, err
	}
	if err := r.bindFuturesSelectionRunJob(runID, accepted.JobID); err != nil {
		return model.FuturesSelectionRun{}, err
	}

	report, waitErr := r.strategyEngine.waitForFuturesStrategyJob(accepted.JobID)
	jobRecord, jobErr := r.strategyEngine.getJobRecord(accepted.JobID)
	if jobErr == nil {
		_ = r.upsertStrategyEngineJobSnapshot(jobRecord)
	}
	if waitErr != nil {
		failureSummary := waitErr.Error()
		if jobErr == nil && strings.TrimSpace(jobRecord.ErrorMessage) != "" {
			failureSummary = jobRecord.ErrorMessage
		}
		_ = r.markFuturesSelectionRunFailed(runID, accepted.JobID, failureSummary)
		return model.FuturesSelectionRun{}, waitErr
	}
	if jobErr == nil && jobRecord.Result != nil && strings.TrimSpace(jobRecord.ResultSummary) == "" {
		jobRecord.ResultSummary = jobRecord.Result.Summary
	}
	if err := r.persistCompletedFuturesSelectionRun(runID, accepted.JobID, report, jobRecord, startedAt); err != nil {
		return model.FuturesSelectionRun{}, err
	}
	return r.AdminGetFuturesSelectionRun(runID)
}

func (r *MySQLGrowthRepo) AdminCompareFuturesSelectionRuns(runIDs []string) (model.FuturesSelectionRunCompareResult, error) {
	normalized := make([]string, 0, len(runIDs))
	seen := map[string]struct{}{}
	for _, item := range runIDs {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}
	result := model.FuturesSelectionRunCompareResult{Items: []model.FuturesSelectionRunCompareItem{}}
	if len(normalized) == 0 {
		return result, nil
	}
	var previousContracts []string
	for index, runID := range normalized {
		run, err := r.AdminGetFuturesSelectionRun(runID)
		if err != nil {
			return result, err
		}
		portfolio, err := r.AdminListFuturesSelectionRunPortfolio(runID)
		if err != nil {
			return result, err
		}
		currentContracts := make([]string, 0, len(portfolio))
		for _, item := range portfolio {
			currentContracts = append(currentContracts, item.Contract)
		}
		added, removed := diffStringSlices(previousContracts, currentContracts)
		result.Items = append(result.Items, model.FuturesSelectionRunCompareItem{
			RunID:              run.RunID,
			TradeDate:          run.TradeDate,
			ProfileID:          run.ProfileID,
			TemplateID:         run.TemplateID,
			TemplateName:       run.TemplateName,
			MarketRegime:       run.MarketRegime,
			Status:             run.Status,
			ReviewStatus:       run.ReviewStatus,
			SelectedCount:      run.SelectedCount,
			PortfolioContracts: currentContracts,
			AddedContracts:     added,
			RemovedContracts:   removed,
		})
		if index == 0 {
			result.BaseRunID = run.RunID
		}
		previousContracts = currentContracts
	}
	return result, nil
}

func (r *MySQLGrowthRepo) AdminGetFuturesSelectionRun(runID string) (model.FuturesSelectionRun, error) {
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
  COALESCE(r.style, ''),
  COALESCE(r.contract_scope, ''),
  r.status,
  COALESCE(r.result_summary, ''),
  COALESCE(CAST(r.warning_messages AS CHAR), ''),
  r.warning_count,
  r.universe_count,
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
  COALESCE(CAST(rv.published_contract_snapshot AS CHAR), ''),
  COALESCE(DATE_FORMAT(rv.approved_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.rejected_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
  COALESCE(DATE_FORMAT(rv.updated_at, '%Y-%m-%dT%H:%i:%sZ'), '')
FROM futures_selection_runs r
LEFT JOIN futures_selection_profile_templates t ON t.id = r.template_id
LEFT JOIN futures_selection_publish_reviews rv ON rv.run_id = r.run_id
WHERE r.run_id = ?`, strings.TrimSpace(runID))
	if err != nil {
		return model.FuturesSelectionRun{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return model.FuturesSelectionRun{}, sql.ErrNoRows
	}
	item, err := scanFuturesSelectionRunRow(rows)
	if err != nil {
		return model.FuturesSelectionRun{}, err
	}
	stageLogs, err := r.loadFuturesSelectionRunStageLogs(item.RunID)
	if err != nil {
		return model.FuturesSelectionRun{}, err
	}
	item.StageLogs = stageLogs
	item.StageDurationsMS = buildFuturesSelectionStageDurationMap(stageLogs)
	r.attachFuturesSelectionRunSummary(&item)
	return item, nil
}

func (r *MySQLGrowthRepo) AdminListFuturesSelectionRunCandidates(runID string) ([]model.FuturesSelectionCandidateSnapshot, error) {
	rows, err := r.db.Query(`
SELECT id, run_id, contract, name, stage, score, direction, risk_level, selected, rank_no,
       COALESCE(reason_summary, ''), COALESCE(CAST(factor_breakdown_json AS CHAR), ''),
       DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM futures_selection_run_candidates
WHERE run_id = ?
ORDER BY FIELD(stage, 'UNIVERSE', 'CANDIDATE_POOL', 'PORTFOLIO'), rank_no ASC, contract ASC`, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FuturesSelectionCandidateSnapshot, 0)
	for rows.Next() {
		var item model.FuturesSelectionCandidateSnapshot
		var factorJSON string
		var selected bool
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.Contract,
			&item.Name,
			&item.Stage,
			&item.Score,
			&item.Direction,
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
	r.enrichFuturesSelectionCandidateSnapshots(runID, items)
	return items, nil
}

func (r *MySQLGrowthRepo) AdminListFuturesSelectionRunPortfolio(runID string) ([]model.FuturesSelectionPortfolioEntry, error) {
	rows, err := r.db.Query(`
SELECT id, run_id, contract, name, rank_no, score, direction, risk_level, COALESCE(position_range, ''),
       COALESCE(reason_summary, ''), COALESCE(CAST(factor_breakdown_json AS CHAR), ''),
       DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM futures_selection_run_portfolio
WHERE run_id = ?
ORDER BY rank_no ASC, contract ASC`, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FuturesSelectionPortfolioEntry, 0)
	for rows.Next() {
		var item model.FuturesSelectionPortfolioEntry
		var factorJSON string
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.Contract,
			&item.Name,
			&item.Rank,
			&item.Score,
			&item.Direction,
			&item.RiskLevel,
			&item.PositionRange,
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
	r.enrichFuturesSelectionPortfolioEntries(runID, items)
	return items, nil
}

func (r *MySQLGrowthRepo) AdminListFuturesSelectionRunEvidence(runID string, contract string) ([]model.FuturesSelectionRunEvidence, error) {
	args := []any{strings.TrimSpace(runID)}
	query := `
SELECT id, run_id, contract, stage, COALESCE(name, ''), COALESCE(portfolio_role, ''), COALESCE(evidence_summary, ''),
       COALESCE(CAST(evidence_cards_json AS CHAR), ''), COALESCE(CAST(positive_reasons_json AS CHAR), ''),
       COALESCE(CAST(veto_reasons_json AS CHAR), ''), COALESCE(CAST(risk_flags_json AS CHAR), ''),
       COALESCE(CAST(related_entities_json AS CHAR), ''),
       DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM futures_selection_run_evidence
WHERE run_id = ?`
	if strings.TrimSpace(contract) != "" {
		query += " AND contract = ?"
		args = append(args, strings.ToUpper(strings.TrimSpace(contract)))
	}
	query += " ORDER BY FIELD(stage, 'PORTFOLIO', 'CANDIDATE_POOL', 'UNIVERSE'), contract ASC"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FuturesSelectionRunEvidence, 0)
	for rows.Next() {
		var item model.FuturesSelectionRunEvidence
		var cardsJSON, positiveJSON, vetoJSON, riskJSON, relatedJSON string
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.Contract,
			&item.Stage,
			&item.Name,
			&item.PortfolioRole,
			&item.EvidenceSummary,
			&cardsJSON,
			&positiveJSON,
			&vetoJSON,
			&riskJSON,
			&relatedJSON,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		item.EvidenceCards = parseJSONMapSlice(cardsJSON)
		item.PositiveReasons = parseJSONStringList(positiveJSON)
		item.VetoReasons = parseJSONStringList(vetoJSON)
		item.RiskFlags = parseJSONStringList(riskJSON)
		item.RelatedEntities = parseJSONMapSlice(relatedJSON)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) AdminListFuturesSelectionRunEvaluations(runID string, contract string) ([]model.FuturesSelectionRunEvaluation, error) {
	args := []any{strings.TrimSpace(runID)}
	query := `
SELECT id, run_id, contract, horizon_day, evaluation_scope, COALESCE(name, ''),
       COALESCE(DATE_FORMAT(entry_date, '%Y-%m-%d'), ''), COALESCE(DATE_FORMAT(exit_date, '%Y-%m-%d'), ''),
       entry_price, exit_price, return_pct, excess_return_pct, max_drawdown_pct, hit_flag, COALESCE(benchmark_symbol, ''),
       DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM futures_selection_run_evaluations
WHERE run_id = ?`
	if strings.TrimSpace(contract) != "" {
		query += " AND contract = ?"
		args = append(args, strings.ToUpper(strings.TrimSpace(contract)))
	}
	query += " ORDER BY contract ASC, evaluation_scope ASC, horizon_day ASC"
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.FuturesSelectionRunEvaluation, 0)
	for rows.Next() {
		var item model.FuturesSelectionRunEvaluation
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.Contract,
			&item.HorizonDay,
			&item.EvaluationScope,
			&item.Name,
			&item.EntryDate,
			&item.ExitDate,
			&item.EntryPrice,
			&item.ExitPrice,
			&item.ReturnPct,
			&item.ExcessReturnPct,
			&item.MaxDrawdownPct,
			&item.HitFlag,
			&item.BenchmarkSymbol,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) AdminApproveFuturesSelectionReview(runID string, operator string, reviewNote string, force bool, overrideReason string) (model.FuturesSelectionPublishReview, error) {
	run, err := r.AdminGetFuturesSelectionRun(runID)
	if err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	if strings.TrimSpace(run.JobID) == "" {
		return model.FuturesSelectionPublishReview{}, sql.ErrNoRows
	}
	published, err := r.AdminPublishStrategyEngineJob(run.JobID, operator, force, overrideReason)
	if err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	payloads, err := toStrategyEngineFuturesPublishPayloads(published.PublishPayloads)
	if err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	if _, err := r.persistStrategyEngineFuturesStrategies(payloads); err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	jobRecord, jobErr := r.AdminGetStrategyEngineJob(run.JobID)
	publishCount := run.PublishCount
	if jobErr == nil && jobRecord.PublishCount > publishCount {
		publishCount = jobRecord.PublishCount
	}
	if publishCount == 0 {
		publishCount = 1
	}
	portfolioEntries, err := r.AdminListFuturesSelectionRunPortfolio(runID)
	if err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	tx, err := r.db.Begin()
	if err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	_, err = tx.Exec(`
UPDATE futures_selection_publish_reviews
SET review_status = 'APPROVED',
    reviewer = ?,
    review_note = ?,
    override_reason = ?,
    publish_id = ?,
    publish_version = ?,
    published_contract_snapshot = ?,
    approved_at = NOW(),
    rejected_at = NULL,
    updated_at = NOW()
WHERE run_id = ?`,
		operator,
		strings.TrimSpace(reviewNote),
		strings.TrimSpace(overrideReason),
		published.PublishID,
		published.Version,
		stockSelectionMustJSON(buildPublishedFuturesContractSnapshot(portfolioEntries)),
		runID,
	)
	if err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	_, err = tx.Exec("UPDATE futures_selection_runs SET publish_count = ?, updated_at = NOW() WHERE run_id = ?", publishCount, runID)
	if err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	return r.getFuturesSelectionReview(runID)
}

func (r *MySQLGrowthRepo) AdminRejectFuturesSelectionReview(runID string, operator string, reviewNote string) (model.FuturesSelectionPublishReview, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	_, err = tx.Exec(`
UPDATE futures_selection_publish_reviews
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
		return model.FuturesSelectionPublishReview{}, err
	}
	if err = tx.Commit(); err != nil {
		return model.FuturesSelectionPublishReview{}, err
	}
	return r.getFuturesSelectionReview(runID)
}

func (r *MySQLGrowthRepo) resolveFuturesSelectionProfileForRun(profileID string) (*model.FuturesSelectionProfile, error) {
	profileID = strings.TrimSpace(profileID)
	if profileID != "" {
		item, err := r.getFuturesSelectionProfile(profileID)
		if err != nil {
			return nil, err
		}
		return &item, nil
	}
	return r.getDefaultFuturesSelectionProfile()
}

func (r *MySQLGrowthRepo) resolveFuturesSelectionTemplateForRun(runTemplateID string, profileTemplateID string) (*model.FuturesSelectionProfileTemplate, error) {
	for _, candidate := range []string{strings.TrimSpace(runTemplateID), strings.TrimSpace(profileTemplateID)} {
		if candidate == "" {
			continue
		}
		item, err := r.getFuturesSelectionProfileTemplate(candidate)
		if err != nil {
			return nil, err
		}
		return &item, nil
	}
	return r.getDefaultFuturesSelectionProfileTemplate()
}

func normalizeFuturesSelectionTradeDate(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw != "" {
		return raw
	}
	return time.Now().Format("2006-01-02")
}

func buildFuturesSelectionJobPayload(
	runID string,
	tradeDate string,
	profile model.FuturesSelectionProfile,
	template *model.FuturesSelectionProfileTemplate,
	input model.FuturesSelectionRunCreateRequest,
	runtimeProfile *model.StrategyAgentProfile,
	runtimeScenario *model.StrategyScenarioTemplate,
	runtimePolicy *model.StrategyPublishPolicy,
) map[string]any {
	universeConfig := mergeStockSelectionConfigMaps(nil, profile.UniverseConfig)
	factorConfig := mergeStockSelectionConfigMaps(nil, profile.FactorConfig)
	portfolioConfig := mergeStockSelectionConfigMaps(nil, profile.PortfolioConfig)
	publishConfig := mergeStockSelectionConfigMaps(nil, profile.PublishConfig)
	if template != nil {
		universeConfig = mergeStockSelectionConfigMaps(template.UniverseDefaults, profile.UniverseConfig)
		factorConfig = mergeStockSelectionConfigMaps(template.FactorDefaults, profile.FactorConfig)
		portfolioConfig = mergeStockSelectionConfigMaps(template.PortfolioDefaults, profile.PortfolioConfig)
		publishConfig = mergeStockSelectionConfigMaps(template.PublishDefaults, profile.PublishConfig)
	}

	payload := map[string]any{
		"run_id":                      runID,
		"trade_date":                  tradeDate,
		"profile_id":                  profile.ID,
		"dry_run":                     input.DryRun,
		"compare_with_last_published": input.CompareWithLastPublished,
		"style":                       firstNonEmpty(profile.StyleDefault, stringValue(universeConfig["style"])),
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
	if limit := intValue(portfolioConfig["limit"]); limit > 0 {
		payload["limit"] = limit
	}
	if maxRisk := strings.ToUpper(stringValue(portfolioConfig["max_risk_level"])); maxRisk != "" {
		payload["max_risk_level"] = maxRisk
	}
	if minConfidence := floatValue(factorConfig["min_confidence"]); minConfidence > 0 {
		payload["min_confidence"] = minConfidence
	}
	if allowMock := boolValue(universeConfig["allow_mock_fallback_on_short_history"]); allowMock {
		payload["allow_mock_fallback_on_short_history"] = true
	}
	if contracts := stockSelectionStringSlice(universeConfig["contracts"]); len(contracts) > 0 {
		payload["contracts"] = contracts
	}
	if runtimeProfile != nil {
		payload["enabled_agents"] = runtimeProfile.EnabledAgents
		payload["positive_threshold"] = runtimeProfile.PositiveThreshold
		payload["negative_threshold"] = runtimeProfile.NegativeThreshold
		payload["allow_veto"] = runtimeProfile.AllowVeto
	}
	if runtimeScenario != nil && len(runtimeScenario.Items) > 0 {
		payload["scenario_templates"] = runtimeScenario.Items
	}
	copyStockSelectionPayloadFields(payload, publishConfig, []string{"review_required", "allow_auto_publish"})
	attachStrategyConfigPayload(payload, nil, runtimeProfile, runtimeScenario, runtimePolicy)
	return payload
}

func (r *MySQLGrowthRepo) insertFuturesSelectionRunStub(runID string, tradeDate string, profile model.FuturesSelectionProfile, template *model.FuturesSelectionProfileTemplate, operator string, startedAt time.Time) error {
	templateID := ""
	templateName := ""
	if template != nil {
		templateID = template.ID
		templateName = template.Name
	}
	_, err := r.db.Exec(`
INSERT INTO futures_selection_runs (
  run_id, trade_date, profile_id, profile_version, template_id, style, contract_scope, status,
  result_summary, warning_count, universe_count, candidate_count, selected_count, publish_count,
  context_meta, template_snapshot, compare_summary, started_at, completed_at, created_by, created_at, updated_at
) VALUES (?, ?, ?, ?, NULLIF(?, ''), ?, ?, 'RUNNING', '', 0, 0, 0, 0, 0, NULL, ?, NULL, ?, NULL, ?, NOW(), NOW())`,
		runID,
		tradeDate,
		profile.ID,
		profile.CurrentVersion,
		templateID,
		firstNonEmpty(profile.StyleDefault, "balanced"),
		firstNonEmpty(profile.ContractScope, "DOMINANT_ALL"),
		stockSelectionMustJSON(map[string]any{
			"id":            templateID,
			"name":          templateName,
			"style_default": profile.StyleDefault,
		}),
		startedAt,
		operator,
	)
	return err
}

func (r *MySQLGrowthRepo) bindFuturesSelectionRunJob(runID string, jobID string) error {
	_, err := r.db.Exec("UPDATE futures_selection_runs SET job_id = ?, updated_at = NOW() WHERE run_id = ?", strings.TrimSpace(jobID), strings.TrimSpace(runID))
	return err
}

func (r *MySQLGrowthRepo) markFuturesSelectionRunFailed(runID string, jobID string, summary string) error {
	_, err := r.db.Exec(`
UPDATE futures_selection_runs
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

func (r *MySQLGrowthRepo) persistCompletedFuturesSelectionRun(
	runID string,
	jobID string,
	report strategyEngineFuturesStrategyReport,
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
		resultSummary = fmt.Sprintf("futures-selection completed with %d publish-ready strategies", report.SelectedCount)
	}
	previousSnapshot, _ := r.loadLatestApprovedFuturesContractSnapshot(runID)
	contextMeta := buildFuturesSelectionRunContextMeta(report)
	if _, err = tx.Exec(`
UPDATE futures_selection_runs
SET job_id = ?,
    status = 'SUCCEEDED',
    result_summary = ?,
    warning_messages = ?,
    warning_count = ?,
    universe_count = ?,
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
		report.StageCounts["CANDIDATE_POOL"],
		report.SelectedCount,
		strings.TrimSpace(report.MarketRegime),
		stockSelectionMustJSON(contextMeta),
		stockSelectionMustJSON(report.TemplateSnapshot),
		stockSelectionMustJSON(buildFuturesSelectionCompareSummary(previousSnapshot, report.PortfolioEntries)),
		startedAt,
		runID,
	); err != nil {
		return err
	}

	if _, err = tx.Exec("DELETE FROM futures_selection_run_stage_logs WHERE run_id = ?", runID); err != nil {
		return err
	}
	for _, stage := range report.StageLogs {
		if _, err = tx.Exec(`
INSERT INTO futures_selection_run_stage_logs (id, run_id, stage_key, stage_order, status, input_count, output_count, duration_ms, detail_message, payload_snapshot, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			newID("fslg"),
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

	if _, err = tx.Exec("DELETE FROM futures_selection_run_candidates WHERE run_id = ?", runID); err != nil {
		return err
	}
	for _, item := range report.CandidateSnapshots {
		if _, err = tx.Exec(`
INSERT INTO futures_selection_run_candidates (id, run_id, contract, name, stage, score, direction, risk_level, selected, rank_no, reason_summary, evidence_summary, portfolio_role, risk_summary, factor_breakdown_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			newID("fsrc"),
			runID,
			item.Contract,
			item.Name,
			item.Stage,
			item.Score,
			item.Direction,
			strings.ToUpper(strings.TrimSpace(item.RiskLevel)),
			item.Selected,
			item.Rank,
			item.ReasonSummary,
			item.EvidenceSummary,
			item.PortfolioRole,
			item.RiskSummary,
			stockSelectionMustJSON(item.FactorBreakdownJSON),
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec("DELETE FROM futures_selection_run_portfolio WHERE run_id = ?", runID); err != nil {
		return err
	}
	for _, item := range report.PortfolioEntries {
		if _, err = tx.Exec(`
INSERT INTO futures_selection_run_portfolio (id, run_id, contract, name, rank_no, score, direction, risk_level, position_range, reason_summary, evidence_summary, portfolio_role, risk_summary, factor_breakdown_json, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			newID("fsrp"),
			runID,
			item.Contract,
			item.Name,
			item.Rank,
			item.Score,
			item.Direction,
			strings.ToUpper(strings.TrimSpace(item.RiskLevel)),
			item.PositionRange,
			item.ReasonSummary,
			item.EvidenceSummary,
			item.PortfolioRole,
			item.RiskSummary,
			stockSelectionMustJSON(item.FactorBreakdownJSON),
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec("DELETE FROM futures_selection_run_evidence WHERE run_id = ?", runID); err != nil {
		return err
	}
	for _, item := range report.EvidenceRecords {
		if _, err = tx.Exec(`
INSERT INTO futures_selection_run_evidence (
  id, run_id, contract, stage, name, portfolio_role, evidence_summary,
  evidence_cards_json, positive_reasons_json, veto_reasons_json, risk_flags_json, related_entities_json,
  created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			newID("fsev"),
			runID,
			item.Contract,
			item.Stage,
			item.Name,
			item.PortfolioRole,
			item.EvidenceSummary,
			stockSelectionMustJSON(item.EvidenceCards),
			stockSelectionMustJSON(item.PositiveReasons),
			stockSelectionMustJSON(item.VetoReasons),
			stockSelectionMustJSON(item.RiskFlags),
			stockSelectionMustJSON(item.RelatedEntities),
		); err != nil {
			return err
		}
	}

	if _, err = tx.Exec("DELETE FROM futures_selection_run_evaluations WHERE run_id = ?", runID); err != nil {
		return err
	}
	for _, item := range report.EvaluationRecords {
		if _, err = tx.Exec(`
INSERT INTO futures_selection_run_evaluations (
  id, run_id, contract, horizon_day, evaluation_scope, name, entry_date, exit_date,
  entry_price, exit_price, return_pct, excess_return_pct, max_drawdown_pct, hit_flag, benchmark_symbol,
  created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())`,
			newID("fsee"),
			runID,
			item.Contract,
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
INSERT INTO futures_selection_publish_reviews (id, run_id, review_status, reviewer, review_note, override_reason, publish_id, publish_version, approved_at, rejected_at, created_at, updated_at)
VALUES (?, ?, 'PENDING', '', '', '', '', 0, NULL, NULL, NOW(), NOW())
ON DUPLICATE KEY UPDATE review_status = VALUES(review_status), reviewer = '', review_note = '', override_reason = '', publish_id = '', publish_version = 0, approved_at = NULL, rejected_at = NULL, updated_at = NOW()`,
		newID("fsrv"),
		runID,
	); err != nil {
		return err
	}

	return tx.Commit()
}

func buildFuturesSelectionRunContextMeta(report strategyEngineFuturesStrategyReport) map[string]any {
	context := map[string]any{}
	for key, value := range report.ContextMeta {
		context[key] = value
	}
	context["market_regime"] = report.MarketRegime
	context["graph_summary"] = report.GraphSummary
	context["graph_snapshot_id"] = report.GraphSnapshotID
	context["related_entities"] = report.RelatedEntities
	context["memory_feedback"] = report.MemoryFeedback
	return context
}

func buildFuturesSelectionCompareSummary(previousSnapshot []map[string]any, current []strategyEngineFuturesPortfolioEntry) map[string]any {
	currentContracts := make([]string, 0, len(current))
	currentSet := make(map[string]struct{}, len(current))
	for _, item := range current {
		contract := strings.ToUpper(strings.TrimSpace(item.Contract))
		if contract == "" {
			continue
		}
		currentContracts = append(currentContracts, contract)
		currentSet[contract] = struct{}{}
	}
	prevSet := make(map[string]struct{}, len(previousSnapshot))
	for _, item := range previousSnapshot {
		contract := strings.ToUpper(strings.TrimSpace(asString(item["contract"])))
		if contract == "" {
			continue
		}
		prevSet[contract] = struct{}{}
	}
	added := make([]string, 0)
	removed := make([]string, 0)
	for _, contract := range currentContracts {
		if _, ok := prevSet[contract]; !ok {
			added = append(added, contract)
		}
	}
	for contract := range prevSet {
		if _, ok := currentSet[contract]; !ok {
			removed = append(removed, contract)
		}
	}
	sort.Strings(removed)
	return map[string]any{
		"current_contracts": currentContracts,
		"added_contracts":   added,
		"removed_contracts": removed,
	}
}

func buildPublishedFuturesContractSnapshot(items []model.FuturesSelectionPortfolioEntry) []map[string]any {
	result := make([]map[string]any, 0, len(items))
	for _, item := range items {
		result = append(result, map[string]any{
			"contract":       item.Contract,
			"name":           item.Name,
			"rank":           item.Rank,
			"score":          item.Score,
			"direction":      item.Direction,
			"risk_level":     item.RiskLevel,
			"position_range": item.PositionRange,
			"reason_summary": item.ReasonSummary,
			"portfolio_role": item.PortfolioRole,
		})
	}
	return result
}

func (r *MySQLGrowthRepo) getFuturesSelectionReview(runID string) (model.FuturesSelectionPublishReview, error) {
	var item model.FuturesSelectionPublishReview
	var snapshotJSON string
	err := r.db.QueryRow(`
SELECT id, run_id, review_status, COALESCE(reviewer, ''), COALESCE(review_note, ''), COALESCE(override_reason, ''),
       COALESCE(publish_id, ''), COALESCE(publish_version, 0), COALESCE(CAST(published_contract_snapshot AS CHAR), ''),
       COALESCE(DATE_FORMAT(approved_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
       COALESCE(DATE_FORMAT(rejected_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
       COALESCE(DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), ''),
       COALESCE(DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ'), '')
FROM futures_selection_publish_reviews
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
		return model.FuturesSelectionPublishReview{}, err
	}
	item.PublishedContractSnapshot = parseJSONMapSlice(snapshotJSON)
	return item, nil
}

func (r *MySQLGrowthRepo) loadLatestApprovedFuturesContractSnapshot(excludeRunID string) ([]map[string]any, error) {
	var snapshotText sql.NullString
	err := r.db.QueryRow(`
SELECT COALESCE(CAST(published_contract_snapshot AS CHAR), '')
FROM futures_selection_publish_reviews
WHERE review_status = 'APPROVED' AND run_id <> ?
ORDER BY approved_at DESC, updated_at DESC
LIMIT 1`, strings.TrimSpace(excludeRunID)).Scan(&snapshotText)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return parseJSONMapSlice(snapshotText.String), nil
}

func (r *MySQLGrowthRepo) loadLatestApprovedFuturesContractEntries() ([]model.FuturesSelectionPortfolioEntry, error) {
	snapshot, err := r.loadLatestApprovedFuturesContractSnapshot("")
	if err != nil {
		return nil, err
	}
	items := make([]model.FuturesSelectionPortfolioEntry, 0, len(snapshot))
	for _, entry := range snapshot {
		items = append(items, model.FuturesSelectionPortfolioEntry{
			Contract:      asString(entry["contract"]),
			Name:          asString(entry["name"]),
			Rank:          asInt(entry["rank"]),
			Score:         asFloat(entry["score"]),
			Direction:     asString(entry["direction"]),
			RiskLevel:     asString(entry["risk_level"]),
			PositionRange: asString(entry["position_range"]),
			ReasonSummary: asString(entry["reason_summary"]),
			PortfolioRole: asString(entry["portfolio_role"]),
		})
	}
	return items, nil
}

func (r *MySQLGrowthRepo) loadFuturesSelectionRunStageLogs(runID string) ([]model.FuturesSelectionRunStageLog, error) {
	rows, err := r.db.Query(`
SELECT id, run_id, stage_key, stage_order, status, input_count, output_count, duration_ms, COALESCE(detail_message, ''), COALESCE(CAST(payload_snapshot AS CHAR), ''), DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ'), DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ')
FROM futures_selection_run_stage_logs
WHERE run_id = ?
ORDER BY stage_order ASC, id ASC`, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.FuturesSelectionRunStageLog, 0)
	for rows.Next() {
		var item model.FuturesSelectionRunStageLog
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

func (r *MySQLGrowthRepo) getLatestFuturesSelectionRun(status string) (*model.FuturesSelectionRun, error) {
	items, _, err := r.AdminListFuturesSelectionRuns(status, "", "", 1, 1)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, sql.ErrNoRows
	}
	cloned := items[0]
	return &cloned, nil
}

func (r *MySQLGrowthRepo) attachFuturesSelectionRunSummary(item *model.FuturesSelectionRun) {
	if item == nil {
		return
	}
	item.StageCounts = map[string]int{
		"UNIVERSE":       item.UniverseCount,
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

func (r *MySQLGrowthRepo) enrichFuturesSelectionCandidateSnapshots(runID string, items []model.FuturesSelectionCandidateSnapshot) {
	evidenceMap, _ := r.loadFuturesSelectionRunEvidenceMap(runID)
	evaluationMap, _ := r.loadFuturesSelectionRunEvaluationStatusMap(runID)
	previousSnapshot, _ := r.loadLatestApprovedFuturesContractSnapshot(runID)
	for index := range items {
		key := futuresSelectionEvidenceMapKey(items[index].Contract, items[index].Stage)
		if evidence, ok := evidenceMap[key]; ok {
			items[index].EvidenceSummary = evidence.EvidenceSummary
			items[index].PortfolioRole = evidence.PortfolioRole
			items[index].RiskSummary = strings.Join(evidence.RiskFlags, "；")
		}
		if items[index].PortfolioRole == "" && items[index].Stage == "PORTFOLIO" {
			items[index].PortfolioRole = "CORE"
		}
		if status, ok := evaluationMap[strings.ToUpper(strings.TrimSpace(items[index].Contract))]; ok {
			items[index].EvaluationStatus = status
		} else {
			items[index].EvaluationStatus = "PENDING"
		}
		items[index].PreviousPublishDiff = buildPreviousFuturesPublishDiff(items[index].Contract, previousSnapshot)
	}
}

func (r *MySQLGrowthRepo) enrichFuturesSelectionPortfolioEntries(runID string, items []model.FuturesSelectionPortfolioEntry) {
	evidenceMap, _ := r.loadFuturesSelectionRunEvidenceMap(runID)
	evaluationMap, _ := r.loadFuturesSelectionRunEvaluationStatusMap(runID)
	previousSnapshot, _ := r.loadLatestApprovedFuturesContractSnapshot(runID)
	for index := range items {
		key := futuresSelectionEvidenceMapKey(items[index].Contract, "PORTFOLIO")
		if evidence, ok := evidenceMap[key]; ok {
			items[index].EvidenceSummary = evidence.EvidenceSummary
			items[index].PortfolioRole = evidence.PortfolioRole
			items[index].RiskSummary = strings.Join(evidence.RiskFlags, "；")
		}
		if items[index].PortfolioRole == "" {
			items[index].PortfolioRole = "CORE"
		}
		if status, ok := evaluationMap[strings.ToUpper(strings.TrimSpace(items[index].Contract))]; ok {
			items[index].EvaluationStatus = status
		} else {
			items[index].EvaluationStatus = "PENDING"
		}
		items[index].PreviousPublishDiff = buildPreviousFuturesPublishDiff(items[index].Contract, previousSnapshot)
	}
}

func (r *MySQLGrowthRepo) loadFuturesSelectionRunEvidenceMap(runID string) (map[string]model.FuturesSelectionRunEvidence, error) {
	items, err := r.AdminListFuturesSelectionRunEvidence(runID, "")
	if err != nil {
		return nil, err
	}
	result := make(map[string]model.FuturesSelectionRunEvidence, len(items))
	for _, item := range items {
		result[futuresSelectionEvidenceMapKey(item.Contract, item.Stage)] = item
	}
	return result, nil
}

func futuresSelectionEvidenceMapKey(contract string, stage string) string {
	return strings.ToUpper(strings.TrimSpace(contract)) + "::" + strings.ToUpper(strings.TrimSpace(stage))
}

func buildPreviousFuturesPublishDiff(contract string, previousSnapshot []map[string]any) map[string]any {
	contract = strings.ToUpper(strings.TrimSpace(contract))
	if contract == "" {
		return map[string]any{}
	}
	for _, item := range previousSnapshot {
		if strings.ToUpper(strings.TrimSpace(asString(item["contract"]))) == contract {
			return map[string]any{
				"status":          "UNCHANGED",
				"previous_rank":   asInt(item["rank"]),
				"previous_reason": asString(item["reason_summary"]),
			}
		}
	}
	return map[string]any{"status": "ADDED"}
}

func (r *MySQLGrowthRepo) loadFuturesSelectionRunEvaluationStatusMap(runID string) (map[string]string, error) {
	items, err := r.AdminListFuturesSelectionRunEvaluations(runID, "")
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, item := range items {
		contract := strings.ToUpper(strings.TrimSpace(item.Contract))
		if contract == "" {
			continue
		}
		if item.ExitDate != "" {
			result[contract] = "DONE"
			continue
		}
		if _, ok := result[contract]; !ok {
			result[contract] = "PENDING"
		}
	}
	return result, nil
}

func toStrategyEngineFuturesPublishPayloads(items []map[string]any) ([]strategyEngineFuturesPublishPayload, error) {
	if len(items) == 0 {
		return nil, nil
	}
	body, err := json.Marshal(items)
	if err != nil {
		return nil, err
	}
	var payloads []strategyEngineFuturesPublishPayload
	if err := json.Unmarshal(body, &payloads); err != nil {
		return nil, err
	}
	return payloads, nil
}

func scanFuturesSelectionRunRow(scanner interface{ Scan(dest ...any) error }) (model.FuturesSelectionRun, error) {
	var item model.FuturesSelectionRun
	var warningJSON, contextJSON, templateSnapshotJSON, compareSummaryJSON string
	var review model.FuturesSelectionPublishReview
	var contractSnapshotJSON string
	err := scanner.Scan(
		&item.RunID,
		&item.TradeDate,
		&item.JobID,
		&item.ProfileID,
		&item.ProfileVersion,
		&item.TemplateID,
		&item.TemplateName,
		&item.MarketRegime,
		&item.Style,
		&item.ContractScope,
		&item.Status,
		&item.ResultSummary,
		&warningJSON,
		&item.WarningCount,
		&item.UniverseCount,
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
		&contractSnapshotJSON,
		&review.ApprovedAt,
		&review.RejectedAt,
		&review.CreatedAt,
		&review.UpdatedAt,
	)
	if err != nil {
		return model.FuturesSelectionRun{}, err
	}
	item.WarningMessages = parseJSONStringList(warningJSON)
	item.ContextMeta = parseJSONMap(contextJSON)
	item.TemplateSnapshot = parseJSONMap(templateSnapshotJSON)
	item.CompareSummary = parseJSONMap(compareSummaryJSON)
	if strings.TrimSpace(review.ID) != "" {
		review.RunID = item.RunID
		review.PublishedContractSnapshot = parseJSONMapSlice(contractSnapshotJSON)
		item.Review = &review
		item.ReviewStatus = review.ReviewStatus
	}
	return item, nil
}

func buildFuturesSelectionStageDurationMap(items []model.FuturesSelectionRunStageLog) map[string]int64 {
	result := make(map[string]int64, len(items))
	for _, item := range items {
		result[item.StageKey] = item.DurationMS
	}
	return result
}

func (r *MySQLGrowthRepo) getFuturesSelectionProfile(id string) (model.FuturesSelectionProfile, error) {
	var item model.FuturesSelectionProfile
	var isDefault bool
	var universeJSON, factorJSON, portfolioJSON, publishJSON string
	err := r.db.QueryRow(`
SELECT
  p.id,
  p.name,
  COALESCE(p.template_id, ''),
  COALESCE(t.name, ''),
  p.status,
  p.is_default,
  COALESCE((
    SELECT MAX(v.version_no) FROM futures_selection_profile_versions v WHERE v.profile_id = p.id
  ), 1),
  COALESCE(p.style_default, 'balanced'),
  COALESCE(p.contract_scope, ''),
  COALESCE(CAST(p.universe_config AS CHAR), ''),
  COALESCE(CAST(p.factor_config AS CHAR), ''),
  COALESCE(CAST(p.portfolio_config AS CHAR), ''),
  COALESCE(CAST(p.publish_config AS CHAR), ''),
  COALESCE(p.description, ''),
  COALESCE(p.updated_by, ''),
  DATE_FORMAT(p.updated_at, '%Y-%m-%dT%H:%i:%sZ'),
  DATE_FORMAT(p.created_at, '%Y-%m-%dT%H:%i:%sZ')
FROM futures_selection_profiles p
LEFT JOIN futures_selection_profile_templates t ON t.id = p.template_id
WHERE p.id = ?`, strings.TrimSpace(id)).Scan(
		&item.ID,
		&item.Name,
		&item.TemplateID,
		&item.TemplateName,
		&item.Status,
		&isDefault,
		&item.CurrentVersion,
		&item.StyleDefault,
		&item.ContractScope,
		&universeJSON,
		&factorJSON,
		&portfolioJSON,
		&publishJSON,
		&item.Description,
		&item.UpdatedBy,
		&item.UpdatedAt,
		&item.CreatedAt,
	)
	if err != nil {
		return model.FuturesSelectionProfile{}, err
	}
	item.IsDefault = isDefault
	item.UniverseConfig = parseJSONMap(universeJSON)
	item.FactorConfig = parseJSONMap(factorJSON)
	item.PortfolioConfig = parseJSONMap(portfolioJSON)
	item.PublishConfig = parseJSONMap(publishJSON)
	return item, nil
}

func (r *MySQLGrowthRepo) getDefaultFuturesSelectionProfile() (*model.FuturesSelectionProfile, error) {
	var id string
	err := r.db.QueryRow(`
SELECT id
FROM futures_selection_profiles
WHERE status = 'ACTIVE' AND is_default = 1
ORDER BY updated_at DESC, id ASC
LIMIT 1`).Scan(&id)
	if err != nil {
		return nil, err
	}
	item, err := r.getFuturesSelectionProfile(id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *MySQLGrowthRepo) getFuturesSelectionProfileTemplate(id string) (model.FuturesSelectionProfileTemplate, error) {
	var item model.FuturesSelectionProfileTemplate
	var isDefault bool
	var universeJSON, factorJSON, portfolioJSON, publishJSON string
	err := r.db.QueryRow(`
SELECT
  id,
  template_key,
  name,
  COALESCE(description, ''),
  COALESCE(market_regime_bias, ''),
  is_default,
  status,
  COALESCE(CAST(universe_defaults_json AS CHAR), ''),
  COALESCE(CAST(factor_defaults_json AS CHAR), ''),
  COALESCE(CAST(portfolio_defaults_json AS CHAR), ''),
  COALESCE(CAST(publish_defaults_json AS CHAR), ''),
  COALESCE(updated_by, ''),
  DATE_FORMAT(updated_at, '%Y-%m-%dT%H:%i:%sZ'),
  DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
FROM futures_selection_profile_templates
WHERE id = ?`, strings.TrimSpace(id)).Scan(
		&item.ID,
		&item.TemplateKey,
		&item.Name,
		&item.Description,
		&item.MarketRegimeBias,
		&isDefault,
		&item.Status,
		&universeJSON,
		&factorJSON,
		&portfolioJSON,
		&publishJSON,
		&item.UpdatedBy,
		&item.UpdatedAt,
		&item.CreatedAt,
	)
	if err != nil {
		return model.FuturesSelectionProfileTemplate{}, err
	}
	item.IsDefault = isDefault
	item.UniverseDefaults = parseJSONMap(universeJSON)
	item.FactorDefaults = parseJSONMap(factorJSON)
	item.PortfolioDefaults = parseJSONMap(portfolioJSON)
	item.PublishDefaults = parseJSONMap(publishJSON)
	return item, nil
}

func (r *MySQLGrowthRepo) getDefaultFuturesSelectionProfileTemplate() (*model.FuturesSelectionProfileTemplate, error) {
	var id string
	err := r.db.QueryRow(`
SELECT id
FROM futures_selection_profile_templates
WHERE status = 'ACTIVE' AND is_default = 1
ORDER BY updated_at DESC, id ASC
LIMIT 1`).Scan(&id)
	if err != nil {
		return nil, err
	}
	item, err := r.getFuturesSelectionProfileTemplate(id)
	if err != nil {
		return nil, err
	}
	return &item, nil
}
