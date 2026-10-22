package repo

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) SavePatternMatches(matches []model.StockPatternMatch) error {
	if len(matches) == 0 {
		return nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO stock_pattern_matches (id, source_symbol, match_symbol, match_date, similarity, lookback, next_7d_returns)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE similarity = VALUES(similarity), match_date = VALUES(match_date), next_7d_returns = VALUES(next_7d_returns)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, m := range matches {
		id := m.ID
		if id == "" {
			id = newID("spm")
		}
		next7dJSON, jsonErr := json.Marshal(m.Next7d)
		if jsonErr != nil {
			err = jsonErr
			return err
		}
		_, err = stmt.Exec(id, m.SourceSymbol, m.MatchSymbol, m.MatchDate, m.Similarity, m.Lookback, string(next7dJSON))
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	return err
}

func (r *MySQLGrowthRepo) GetPatternMatches(sourceSymbol string, lookback int, limit int) ([]model.StockPatternMatch, error) {
	rows, err := r.db.Query(`
		SELECT id, source_symbol, match_symbol, match_date, similarity, lookback, next_7d_returns
		FROM stock_pattern_matches
		WHERE source_symbol = ? AND lookback = ?
		ORDER BY similarity DESC
		LIMIT ?
	`, sourceSymbol, lookback, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.StockPatternMatch
	for rows.Next() {
		var m model.StockPatternMatch
		var next7dStr string
		err := rows.Scan(&m.ID, &m.SourceSymbol, &m.MatchSymbol, &m.MatchDate, &m.Similarity, &m.Lookback, &next7dStr)
		if err != nil {
			return nil, err
		}
		if next7dStr != "" {
			var next7d []float64
			if jsonErr := json.Unmarshal([]byte(next7dStr), &next7d); jsonErr == nil {
				m.Next7d = next7d
			}
		}
		result = append(result, m)
	}
	return result, rows.Err()
}

func (r *MySQLGrowthRepo) GetLatestRecommendationBySymbol(symbol string) (model.StockRecommendation, error) {
	var item model.StockRecommendation
	var positionRange, reasonSummary, strategyVersion, reviewer, publisher, reviewNote, performanceLabel sql.NullString
	var validFrom, validTo, createdAt time.Time
	err := r.db.QueryRow(`
		SELECT id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, source_type, strategy_version, reviewer, publisher, review_note, performance_label, created_at
		FROM stock_recommendations
		WHERE symbol = ? OR symbol LIKE ?
		ORDER BY created_at DESC, valid_from DESC
		LIMIT 1
	`, symbol, symbol+".%").Scan(
		&item.ID, &item.Symbol, &item.Name, &item.Score, &item.RiskLevel, &positionRange, &validFrom, &validTo, &item.Status, &reasonSummary,
		&item.SourceType, &strategyVersion, &reviewer, &publisher, &reviewNote, &performanceLabel, &createdAt,
	)
	if err != nil {
		return model.StockRecommendation{}, err
	}
	if positionRange.Valid {
		item.PositionRange = positionRange.String
	}
	if reasonSummary.Valid {
		item.ReasonSummary = reasonSummary.String
	}
	if strategyVersion.Valid {
		item.StrategyVersion = strategyVersion.String
	}
	if reviewer.Valid {
		item.Reviewer = reviewer.String
	}
	if publisher.Valid {
		item.Publisher = publisher.String
	}
	if reviewNote.Valid {
		item.ReviewNote = reviewNote.String
	}
	if performanceLabel.Valid {
		item.PerformanceLabel = performanceLabel.String
	}
	item.ValidFrom = validFrom.Format(time.RFC3339)
	item.ValidTo = validTo.Format(time.RFC3339)
	return item, nil
}

type LLMEngine struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	BaseURL   string `json:"base_url"`
	APIKey    string `json:"api_key"`
	ModelName string `json:"model_name"`
	IsActive  bool   `json:"is_active"`
}

func (r *MySQLGrowthRepo) getActiveLLMConfig() (apiKey, baseURL, modelName string) {
	var enginesJSON string
	err := r.db.QueryRow("SELECT config_value FROM system_configs WHERE config_key = 'llm.engines'").Scan(&enginesJSON)
	if err == nil && enginesJSON != "" {
		var engines []LLMEngine
		if err := json.Unmarshal([]byte(enginesJSON), &engines); err == nil {
			for _, engine := range engines {
				if engine.IsActive {
					return engine.APIKey, engine.BaseURL, engine.ModelName
				}
			}
		}
	}

	_ = r.db.QueryRow("SELECT config_value FROM system_configs WHERE config_key = 'llm.api_key'").Scan(&apiKey)
	_ = r.db.QueryRow("SELECT config_value FROM system_configs WHERE config_key = 'llm.base_url'").Scan(&baseURL)
	_ = r.db.QueryRow("SELECT config_value FROM system_configs WHERE config_key = 'llm.model_name'").Scan(&modelName)
	return apiKey, baseURL, modelName
}

func (r *MySQLGrowthRepo) GenerateRealtimeStockInsight(userID string, symbol string) (model.StockRecommendationInsight, error) {
	llmApiKey, llmBaseUrl, llmModelName := r.getActiveLLMConfig()

	llmConfig := make(map[string]string)
	if llmApiKey != "" {
		llmConfig["api_key"] = strings.TrimSpace(llmApiKey)
	}
	if llmBaseUrl != "" {
		llmConfig["base_url"] = strings.TrimSpace(llmBaseUrl)
	}
	if llmModelName != "" {
		llmConfig["model_name"] = strings.TrimSpace(llmModelName)
	}

	resp, err := r.strategyEngine.evaluateStockRealtime(symbol, llmConfig)
	if err != nil {
		return model.StockRecommendationInsight{}, fmt.Errorf("realtime evaluation failed: %w", err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	recoID := newID("sr")
	now := time.Now()
	validFrom := now
	validTo := now.Add(24 * time.Hour * 7)

	_, err = tx.Exec(`
		INSERT INTO stock_recommendations (id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, source_type, strategy_version, reviewer, publisher, review_note, performance_label, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'REALTIME_CACHED', ?, 'SYSTEM', 'realtime-v1', '', 'system', '由系统实时智能体分析计算。', 'ESTIMATED', ?)
	`, recoID, resp.Recommendation.Symbol, resp.Recommendation.Name, resp.Recommendation.Score, resp.Recommendation.RiskLevel, resp.Recommendation.PositionRange, validFrom, validTo, resp.Recommendation.ReasonSummary, now)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}

	_, err = tx.Exec(`
		INSERT INTO stock_reco_details (reco_id, tech_score, fund_score, sentiment_score, money_flow_score, take_profit, stop_loss, risk_note)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, recoID, resp.Detail.TechScore, resp.Detail.FundScore, resp.Detail.SentimentScore, resp.Detail.MoneyFlowScore, resp.Detail.TakeProfit, resp.Detail.StopLoss, resp.Detail.RiskNote)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}

	report := map[string]any{
		"generated_at":      now.Format(time.RFC3339),
		"graph_summary":     resp.Explanation.GraphSummary,
		"consensus_summary": resp.Explanation.ConsensusSummary,
		"candidates": []map[string]any{
			{
				"symbol":           resp.Recommendation.Symbol,
				"reason_summary":   resp.Recommendation.ReasonSummary,
				"strategy_version": "realtime-v1",
				"risk_flags":       resp.Explanation.RiskFlags,
				"invalidations":    resp.Explanation.Invalidations,
			},
		},
		"simulations": []map[string]any{
			{
				"asset_key":        resp.Recommendation.Symbol,
				"asset_type":       "STOCK",
				"agents":           resp.Explanation.AgentOpinions,
				"scenarios":        resp.Explanation.ScenarioSnapshots,
				"consensus_action": "HOLD",
			},
		},
	}

	reportBytes, err := json.Marshal(report)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}

	jobID := "job_rt_" + recoID
	publishID := "pub_rt_" + recoID

	_, err = tx.Exec(`
		INSERT INTO strategy_job_runs (job_id, job_type, payload_snapshot, remote_created_at, trade_date, status, created_at, synced_at)
		VALUES (?, 'stock-selection', '{}', ?, ?, 'SUCCEEDED', ?, ?)
	`, jobID, now, now, now, now)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}

	_, err = tx.Exec(`
		INSERT INTO strategy_job_artifacts (job_id, report_snapshot, created_at, synced_at)
		VALUES (?, ?, ?, ?)
	`, jobID, string(reportBytes), now, now)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}

	replayID := "job_replay_" + publishID
	_, err = tx.Exec(`
		INSERT INTO strategy_job_replays (id, job_id, publish_id, publish_version, replay_snapshot, created_at)
		VALUES (?, ?, ?, 1, '{}', ?)
	`, replayID, jobID, publishID, now)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}

	err = tx.Commit()
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}

	return r.GetStockRecommendationInsight(userID, recoID)
}

func (r *MySQLGrowthRepo) AdminGetUsersByEmailRule(ruleType string) ([]model.AdminUser, error) {
	var query string
	switch ruleType {
	case "VIP_EXPIRING_7_DAYS":
		query = `
			SELECT u.id, u.phone, COALESCE(u.email, ''), u.status, u.kyc_status, u.member_level, u.vip_expire_at, u.created_at
			FROM users u
			WHERE u.member_level LIKE 'VIP%' AND u.vip_expire_at IS NOT NULL AND u.vip_expire_at > NOW() AND u.vip_expire_at <= DATE_ADD(NOW(), INTERVAL 7 DAY)
			ORDER BY u.vip_expire_at ASC`
	case "VIP_EXPIRED_14_DAYS":
		query = `
			SELECT u.id, u.phone, COALESCE(u.email, ''), u.status, u.kyc_status, u.member_level, u.vip_expire_at, u.created_at
			FROM users u
			WHERE u.member_level = 'FREE' AND u.vip_expire_at IS NOT NULL AND u.vip_expire_at <= DATE_SUB(NOW(), INTERVAL 14 DAY)
			ORDER BY u.vip_expire_at DESC`
	case "REGISTERED_7_DAYS_NO_VIP":
		query = `
			SELECT u.id, u.phone, COALESCE(u.email, ''), u.status, u.kyc_status, u.member_level, u.vip_expire_at, u.created_at
			FROM users u
			WHERE u.member_level = 'FREE' AND u.vip_expire_at IS NULL AND u.created_at <= DATE_SUB(NOW(), INTERVAL 7 DAY)
			ORDER BY u.created_at DESC`
	case "REGISTERED_30_DAYS_NO_VIP":
		query = `
			SELECT u.id, u.phone, COALESCE(u.email, ''), u.status, u.kyc_status, u.member_level, u.vip_expire_at, u.created_at
			FROM users u
			WHERE u.member_level = 'FREE' AND u.vip_expire_at IS NULL AND u.created_at <= DATE_SUB(NOW(), INTERVAL 30 DAY)
			ORDER BY u.created_at DESC`
	case "READING_ACTIVE":
		query = `
			SELECT u.id, u.phone, COALESCE(u.email, ''), u.status, u.kyc_status, u.member_level, u.vip_expire_at, u.created_at
			FROM users u
			JOIN (
				SELECT user_id, COUNT(*) as cnt
				FROM browse_histories
				WHERE viewed_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
				GROUP BY user_id
			) bh ON bh.user_id = u.id
			WHERE u.status = 'ACTIVE'
			ORDER BY bh.cnt DESC`
	case "READING_SILENT":
		query = `
			SELECT u.id, u.phone, COALESCE(u.email, ''), u.status, u.kyc_status, u.member_level, u.vip_expire_at, u.created_at
			FROM users u
			LEFT JOIN browse_histories bh ON bh.user_id = u.id
			WHERE u.status = 'ACTIVE'
			GROUP BY u.id, u.phone, u.email, u.status, u.kyc_status, u.member_level, u.vip_expire_at, u.created_at
			HAVING MAX(bh.viewed_at) IS NULL OR MAX(bh.viewed_at) < DATE_SUB(NOW(), INTERVAL 7 DAY)
			ORDER BY MAX(bh.viewed_at) ASC`
	default:
		return nil, fmt.Errorf("unknown rule type: %s", ruleType)
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.AdminUser, 0)
	for rows.Next() {
		var item model.AdminUser
		var phone sql.NullString
		var email sql.NullString
		var vipExpireAt sql.NullTime
		var createdAt time.Time

		if err := rows.Scan(&item.ID, &phone, &email, &item.Status, &item.KYCStatus, &item.MemberLevel, &vipExpireAt, &createdAt); err != nil {
			return nil, err
		}

		if phone.Valid {
			item.Phone = phone.String
		}
		if email.Valid {
			item.Email = email.String
		}
		if vipExpireAt.Valid {
			item.VIPExpireAt = vipExpireAt.Time.Format(time.RFC3339)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		users = append(users, item)
	}
	return users, nil
}

func (r *MySQLGrowthRepo) AdminRecordEmailLog(userID string, email string, subject string, body string, ruleType string, status string, errMsg string) error {
	id := newID("eml")
	_, err := r.db.Exec(`
		INSERT INTO user_email_logs (id, user_id, email, subject, body, rule_type, status, error_message, sent_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, userID, email, subject, body, ruleType, status, sql.NullString{String: errMsg, Valid: errMsg != ""}, time.Now(),
	)
	return err
}
