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

func buildStrategyForecastL3LearningRecord(detail model.StrategyForecastL3RunDetail, now time.Time) model.StrategyForecastL3LearningRecord {
	report := detail.Report
	roleEffectiveness := map[string]float64{}
	biasLabel := "UNCALIBRATED"
	summary := "No report available for calibration."
	scenarioHit := false
	triggerHit := false
	invalidationEarly := false
	if report != nil {
		for _, item := range report.RoleDisagreements {
			if strings.TrimSpace(item.Role) == "" {
				continue
			}
			score := 0.58
			if item.Veto {
				score = 0.62
			}
			roleEffectiveness[item.Role] = score
		}
		scenarioHit = report.PrimaryScenario != ""
		triggerHit = len(report.TriggerChecklist) > 0
		invalidationEarly = report.PrimaryScenario == "bear" || report.PrimaryScenario == "reversal"
		summary = firstNonEmpty(report.ExecutiveSummary, detail.Run.Summary.ExecutiveSummary, "Forecast L3 report completed.")
		if invalidationEarly {
			biasLabel = "RISK_FIRST"
		}
	}
	return model.StrategyForecastL3LearningRecord{
		ID:                newID("l3learn"),
		RunID:             detail.Run.ID,
		TargetType:        detail.Run.TargetType,
		TargetKey:         detail.Run.TargetKey,
		ScenarioHit:       scenarioHit,
		TriggerHit:        triggerHit,
		InvalidationEarly: invalidationEarly,
		BiasLabel:         biasLabel,
		RoleEffectiveness: roleEffectiveness,
		Summary:           summary,
		CreatedAt:         now.UTC().Format(time.RFC3339),
		UpdatedAt:         now.UTC().Format(time.RFC3339),
	}
}

func summarizeStrategyForecastL3LearningRecords(items []model.StrategyForecastL3LearningRecord) []model.StrategyForecastL3QualitySummary {
	type aggregate struct {
		total             int
		scenarioHit       int
		triggerHit        int
		invalidationEarly int
		roleScores        map[string]float64
		roleCounts        map[string]int
		lastRecordedAt    string
	}

	aggregates := make(map[string]*aggregate)
	for _, item := range items {
		targetType := normalizeStrategyForecastL3TargetType(item.TargetType)
		if targetType == "" {
			continue
		}
		current, ok := aggregates[targetType]
		if !ok {
			current = &aggregate{
				roleScores: make(map[string]float64),
				roleCounts: make(map[string]int),
			}
			aggregates[targetType] = current
		}
		current.total++
		if item.ScenarioHit {
			current.scenarioHit++
		}
		if item.TriggerHit {
			current.triggerHit++
		}
		if item.InvalidationEarly {
			current.invalidationEarly++
		}
		for role, value := range item.RoleEffectiveness {
			current.roleScores[role] += value
			current.roleCounts[role]++
		}
		if item.CreatedAt > current.lastRecordedAt {
			current.lastRecordedAt = item.CreatedAt
		}
	}

	keys := make([]string, 0, len(aggregates))
	for key := range aggregates {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	results := make([]model.StrategyForecastL3QualitySummary, 0, len(keys))
	for _, key := range keys {
		current := aggregates[key]
		roleEffectiveness := make(map[string]float64, len(current.roleScores))
		for role, total := range current.roleScores {
			roleEffectiveness[role] = total / float64(current.roleCounts[role])
		}
		results = append(results, model.StrategyForecastL3QualitySummary{
			TargetType:             key,
			TotalRuns:              current.total,
			SucceededRuns:          current.total,
			ScenarioHitRate:        safeForecastL3Ratio(current.scenarioHit, current.total),
			TriggerHitRate:         safeForecastL3Ratio(current.triggerHit, current.total),
			InvalidationEarlyRate:  safeForecastL3Ratio(current.invalidationEarly, current.total),
			RoleEffectiveness:      roleEffectiveness,
			LastLearningRecordedAt: current.lastRecordedAt,
		})
	}
	return results
}

func buildStrategyForecastL3RunReview(item model.StrategyForecastL3LearningRecord) model.StrategyForecastL3RunReview {
	review := model.StrategyForecastL3RunReview{
		RunID:             strings.TrimSpace(item.RunID),
		TargetType:        normalizeStrategyForecastL3TargetType(item.TargetType),
		TargetKey:         strings.TrimSpace(item.TargetKey),
		ScenarioHit:       item.ScenarioHit,
		TriggerHit:        item.TriggerHit,
		InvalidationEarly: item.InvalidationEarly,
		BiasLabel:         strings.TrimSpace(item.BiasLabel),
		RoleEffectiveness: cloneForecastL3RoleEffectiveness(item.RoleEffectiveness),
		ReviewedAt:        strings.TrimSpace(item.CreatedAt),
	}

	switch {
	case item.InvalidationEarly:
		review.ReviewGrade = "D"
		review.ReviewScore = 35
		review.ReviewVerdict = "风险边界过早触发，本次深推演需要明显降权看待。"
		review.ReviewNotes = append(review.ReviewNotes, "失效条件在主情景充分展开前就被触发。")
	case item.ScenarioHit && item.TriggerHit:
		review.ReviewGrade = "A"
		review.ReviewScore = 85
		review.ReviewVerdict = "主情景与触发条件均得到验证，这次深推演具有较高参考价值。"
		review.ReviewNotes = append(review.ReviewNotes, "主情景得到验证，且关键触发条件落地。")
	case item.ScenarioHit:
		review.ReviewGrade = "B"
		review.ReviewScore = 70
		review.ReviewVerdict = "方向判断基本成立，但确认信号不够完整，仍需控制节奏。"
		review.ReviewNotes = append(review.ReviewNotes, "主情景命中，但触发条件确认不充分。")
	default:
		review.ReviewGrade = "C"
		review.ReviewScore = 55
		review.ReviewVerdict = "主情景未充分走出，但风险边界尚未提前失效，结论需要继续观察。"
		review.ReviewNotes = append(review.ReviewNotes, "主情景未明显兑现，仍需跟踪后续验证。")
	}

	if note := strings.TrimSpace(item.Summary); note != "" {
		review.ReviewNotes = append(review.ReviewNotes, note)
	}
	switch review.BiasLabel {
	case "UNDERCONFIRMED":
		review.ReviewNotes = append(review.ReviewNotes, "历史复盘提示：当时更像提前预判，确认度偏弱。")
	case "RISK_FIRST":
		review.ReviewNotes = append(review.ReviewNotes, "历史复盘提示：风险信号先于收益兑现，需要优先守住边界。")
	case "UNCALIBRATED":
		review.ReviewNotes = append(review.ReviewNotes, "历史复盘提示：当前样本仍偏少，建议结合更多历史推演观察。")
	}

	return review
}

func safeForecastL3Ratio(numerator int, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func (r *MySQLGrowthRepo) hasStrategyForecastL3LearningRecord(runID string) (bool, error) {
	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM strategy_forecast_l3_learning_records WHERE run_id = ?`, runID).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *MySQLGrowthRepo) insertStrategyForecastL3LearningRecord(item model.StrategyForecastL3LearningRecord) error {
	_, err := r.db.Exec(`
INSERT INTO strategy_forecast_l3_learning_records (
	id, run_id, target_type, target_key, scenario_hit, trigger_hit, invalidation_early,
	bias_label, role_effectiveness_json, summary_text, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	scenario_hit = VALUES(scenario_hit),
	trigger_hit = VALUES(trigger_hit),
	invalidation_early = VALUES(invalidation_early),
	bias_label = VALUES(bias_label),
	role_effectiveness_json = VALUES(role_effectiveness_json),
	summary_text = VALUES(summary_text),
	updated_at = VALUES(updated_at)`,
		item.ID,
		item.RunID,
		item.TargetType,
		item.TargetKey,
		item.ScenarioHit,
		item.TriggerHit,
		item.InvalidationEarly,
		nullableString(item.BiasLabel),
		nullableForecastL3JSON(marshalJSONText(item.RoleEffectiveness)),
		nullableString(item.Summary),
		parseForecastL3TimestampOrNow(item.CreatedAt),
		parseForecastL3TimestampOrNow(item.UpdatedAt),
	)
	return err
}

func (r *MySQLGrowthRepo) loadStrategyForecastL3LearningRecords(targetType string, days int) ([]model.StrategyForecastL3LearningRecord, error) {
	whereParts := make([]string, 0, 2)
	args := make([]interface{}, 0, 2)
	if normalized := normalizeStrategyForecastL3TargetType(targetType); normalized != "" {
		whereParts = append(whereParts, "target_type = ?")
		args = append(args, normalized)
	}
	if days > 0 {
		whereParts = append(whereParts, "created_at >= DATE_SUB(NOW(), INTERVAL ? DAY)")
		args = append(args, days)
	}
	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = " WHERE " + strings.Join(whereParts, " AND ")
	}
	rows, err := r.db.Query(`
SELECT
	id,
	run_id,
	target_type,
	target_key,
	scenario_hit,
	trigger_hit,
	invalidation_early,
	COALESCE(bias_label, ''),
	COALESCE(CAST(role_effectiveness_json AS CHAR), ''),
	COALESCE(summary_text, ''),
	created_at,
	updated_at
FROM strategy_forecast_l3_learning_records`+whereClause+`
ORDER BY created_at DESC`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyForecastL3LearningRecord, 0)
	for rows.Next() {
		var item model.StrategyForecastL3LearningRecord
		var roleEffectivenessJSON sql.NullString
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.TargetType,
			&item.TargetKey,
			&item.ScenarioHit,
			&item.TriggerHit,
			&item.InvalidationEarly,
			&item.BiasLabel,
			&roleEffectivenessJSON,
			&item.Summary,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.RoleEffectiveness = parseStrategyForecastL3RoleEffectiveness(roleEffectivenessJSON.String)
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MySQLGrowthRepo) loadStrategyForecastL3LearningRecordsForRun(runID string) ([]model.StrategyForecastL3LearningRecord, error) {
	rows, err := r.db.Query(`
SELECT
	id,
	run_id,
	target_type,
	target_key,
	scenario_hit,
	trigger_hit,
	invalidation_early,
	COALESCE(bias_label, ''),
	COALESCE(CAST(role_effectiveness_json AS CHAR), ''),
	COALESCE(summary_text, ''),
	created_at,
	updated_at
FROM strategy_forecast_l3_learning_records
WHERE run_id = ?
ORDER BY created_at DESC, id DESC`, strings.TrimSpace(runID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StrategyForecastL3LearningRecord, 0)
	for rows.Next() {
		var item model.StrategyForecastL3LearningRecord
		var roleEffectivenessJSON sql.NullString
		var createdAt time.Time
		var updatedAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.TargetType,
			&item.TargetKey,
			&item.ScenarioHit,
			&item.TriggerHit,
			&item.InvalidationEarly,
			&item.BiasLabel,
			&roleEffectivenessJSON,
			&item.Summary,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, err
		}
		item.RoleEffectiveness = parseStrategyForecastL3RoleEffectiveness(roleEffectivenessJSON.String)
		item.CreatedAt = createdAt.UTC().Format(time.RFC3339)
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339)
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func parseStrategyForecastL3RoleEffectiveness(raw string) map[string]float64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var items map[string]float64
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	return items
}

func buildForecastL3LearningSummaryMessage(item model.StrategyForecastL3LearningRecord) string {
	return fmt.Sprintf("scenario_hit=%t trigger_hit=%t invalidation_early=%t", item.ScenarioHit, item.TriggerHit, item.InvalidationEarly)
}

func cloneForecastL3RoleEffectiveness(source map[string]float64) map[string]float64 {
	if len(source) == 0 {
		return nil
	}
	cloned := make(map[string]float64, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}
