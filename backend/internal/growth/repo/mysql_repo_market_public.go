package repo

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) ListArbitrageOpportunities(typeFilter string, page int, pageSize int) ([]model.ArbitrageOpportunity, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := ""
	if typeFilter != "" {
		filter = " WHERE type = ?"
		args = append(args, typeFilter)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM arbitrage_recos"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, type, contract_a, contract_b, spread, percentile, status
FROM arbitrage_recos` + filter + `
ORDER BY id DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.ArbitrageOpportunity, 0)
	for rows.Next() {
		var item model.ArbitrageOpportunity
		var spread, percentile sql.NullFloat64
		if err := rows.Scan(&item.ID, &item.Type, &item.ContractA, &item.ContractB, &spread, &percentile, &item.Status); err != nil {
			return nil, 0, err
		}
		if spread.Valid {
			item.Spread = spread.Float64
		}
		if percentile.Valid {
			item.Percentile = percentile.Float64
		}
		item.RiskLevel = "MEDIUM"
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) ListFuturesArbitrage(typeFilter string, page int, pageSize int) ([]model.ArbitrageRecommendation, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := ""
	if typeFilter != "" {
		filter = " WHERE type = ?"
		args = append(args, typeFilter)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM arbitrage_recos"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, type, contract_a, contract_b, spread, percentile, entry_point, exit_point, stop_point, status
FROM arbitrage_recos` + filter + `
ORDER BY id DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.ArbitrageRecommendation, 0)
	for rows.Next() {
		var item model.ArbitrageRecommendation
		var spread, percentile, entryPoint, exitPoint, stopPoint sql.NullFloat64
		if err := rows.Scan(&item.ID, &item.Type, &item.ContractA, &item.ContractB, &spread, &percentile, &entryPoint, &exitPoint, &stopPoint, &item.Status); err != nil {
			return nil, 0, err
		}
		if spread.Valid {
			item.Spread = spread.Float64
		}
		if percentile.Valid {
			item.Percentile = percentile.Float64
		}
		if entryPoint.Valid {
			item.EntryPoint = entryPoint.Float64
		}
		if exitPoint.Valid {
			item.ExitPoint = exitPoint.Float64
		}
		if stopPoint.Valid {
			item.StopPoint = stopPoint.Float64
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) GetFuturesArbitrageDetail(id string) (model.ArbitrageRecommendation, error) {
	var item model.ArbitrageRecommendation
	var spread, percentile, entryPoint, exitPoint, stopPoint sql.NullFloat64
	var triggerRule sql.NullString
	err := r.db.QueryRow(`
SELECT id, type, contract_a, contract_b, spread, percentile, entry_point, exit_point, stop_point, trigger_rule, status
FROM arbitrage_recos
WHERE id = ?
LIMIT 1`, id).Scan(
		&item.ID, &item.Type, &item.ContractA, &item.ContractB, &spread, &percentile,
		&entryPoint, &exitPoint, &stopPoint, &triggerRule, &item.Status,
	)
	if err != nil {
		return model.ArbitrageRecommendation{}, err
	}
	if spread.Valid {
		item.Spread = spread.Float64
	}
	if percentile.Valid {
		item.Percentile = percentile.Float64
	}
	if entryPoint.Valid {
		item.EntryPoint = entryPoint.Float64
	}
	if exitPoint.Valid {
		item.ExitPoint = exitPoint.Float64
	}
	if stopPoint.Valid {
		item.StopPoint = stopPoint.Float64
	}
	if triggerRule.Valid {
		item.TriggerRule = triggerRule.String
	}
	return item, nil
}

func (r *MySQLGrowthRepo) CreateFuturesAlert(userID string, contract string, alertType string, threshold float64) (string, error) {
	id := newID("fa")
	_, err := r.db.Exec(`
INSERT INTO futures_alerts (id, user_id, contract, alert_type, threshold, status, created_at)
VALUES (?, ?, ?, ?, ?, 'ACTIVE', ?)`,
		id, userID, strings.TrimSpace(contract), strings.ToUpper(strings.TrimSpace(alertType)), threshold, time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) ListFuturesReviews(page int, pageSize int) ([]model.FuturesReview, int, error) {
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM futures_reviews").Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT id, strategy_id, hit_rate, pnl, max_drawdown, review_date
FROM futures_reviews
ORDER BY review_date DESC, id DESC
LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.FuturesReview, 0)
	for rows.Next() {
		var item model.FuturesReview
		var hitRate, pnl, maxDrawdown sql.NullFloat64
		var reviewDate time.Time
		if err := rows.Scan(&item.ID, &item.StrategyID, &hitRate, &pnl, &maxDrawdown, &reviewDate); err != nil {
			return nil, 0, err
		}
		if hitRate.Valid {
			item.HitRate = hitRate.Float64
		}
		if pnl.Valid {
			item.PnL = pnl.Float64
		}
		if maxDrawdown.Valid {
			item.MaxDrawdown = maxDrawdown.Float64
		}
		item.ReviewDate = reviewDate.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) ListMarketEvents(eventType string, page int, pageSize int) ([]model.MarketEvent, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := ""
	if eventType != "" {
		filter = " WHERE event_type = ?"
		args = append(args, eventType)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM market_events"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, event_type, symbol, summary, trigger_rule, created_at
FROM market_events` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.MarketEvent, 0)
	for rows.Next() {
		var item model.MarketEvent
		var summary, triggerRule sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.EventType, &item.Symbol, &summary, &triggerRule, &createdAt); err != nil {
			return nil, 0, err
		}
		if summary.Valid {
			item.Summary = summary.String
		}
		if triggerRule.Valid {
			item.TriggerRule = triggerRule.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) GetMarketEventDetail(id string) (model.MarketEvent, error) {
	var item model.MarketEvent
	var createdAt time.Time
	var summary, triggerRule, source sql.NullString
	err := r.db.QueryRow(`
SELECT id, event_type, symbol, summary, trigger_rule, source, created_at
FROM market_events
WHERE id = ?
LIMIT 1`, id).Scan(
		&item.ID, &item.EventType, &item.Symbol, &summary, &triggerRule, &source, &createdAt,
	)
	if err != nil {
		return model.MarketEvent{}, err
	}
	if summary.Valid {
		item.Summary = summary.String
	}
	if triggerRule.Valid {
		item.TriggerRule = triggerRule.String
	}
	if source.Valid {
		item.Source = source.String
	}
	item.CreatedAt = createdAt.Format(time.RFC3339)
	return item, nil
}

func (r *MySQLGrowthRepo) GetFuturesGuidance(contract string) (model.FuturesGuidance, error) {
	var item model.FuturesGuidance
	var validTo time.Time
	err := r.db.QueryRow(`
SELECT contract, guidance_direction, position_level, entry_range, take_profit_range, stop_loss_range, risk_level, invalid_condition, valid_to
FROM futures_guidances
WHERE contract = ?
ORDER BY valid_to DESC
LIMIT 1`, contract).Scan(
		&item.Contract, &item.GuidanceDirection, &item.PositionLevel, &item.EntryRange,
		&item.TakeProfitRange, &item.StopLossRange, &item.RiskLevel, &item.InvalidCondition, &validTo,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.FuturesGuidance{}, nil
		}
		return model.FuturesGuidance{}, err
	}
	item.ValidTo = validTo.Format(time.RFC3339)
	return item, nil
}

func (r *MySQLGrowthRepo) ListPublicHoldings(symbol string, page int, pageSize int) ([]model.PublicHolding, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if symbol != "" {
		filter += " AND symbol = ?"
		args = append(args, symbol)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM public_holdings"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, holder, symbol, ratio, disclosed_at, source
FROM public_holdings` + filter + `
ORDER BY disclosed_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.PublicHolding, 0)
	for rows.Next() {
		var item model.PublicHolding
		var ratio sql.NullFloat64
		var source sql.NullString
		var disclosedAt time.Time
		if err := rows.Scan(&item.ID, &item.Holder, &item.Symbol, &ratio, &disclosedAt, &source); err != nil {
			return nil, 0, err
		}
		if ratio.Valid {
			item.Ratio = ratio.Float64
		}
		if source.Valid {
			item.Source = source.String
		}
		item.DisclosedAt = disclosedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) ListPublicFuturesPositions(contract string, page int, pageSize int) ([]model.PublicFuturesPosition, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if contract != "" {
		filter += " AND contract = ?"
		args = append(args, contract)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM futures_positions_public"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, contract, long_position, short_position, disclosed_at, source
FROM futures_positions_public` + filter + `
ORDER BY disclosed_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.PublicFuturesPosition, 0)
	for rows.Next() {
		var item model.PublicFuturesPosition
		var longPosition, shortPosition sql.NullFloat64
		var source sql.NullString
		var disclosedAt time.Time
		if err := rows.Scan(&item.ID, &item.Contract, &longPosition, &shortPosition, &disclosedAt, &source); err != nil {
			return nil, 0, err
		}
		if longPosition.Valid {
			item.LongPosition = longPosition.Float64
		}
		if shortPosition.Valid {
			item.ShortPosition = shortPosition.Float64
		}
		if source.Valid {
			item.Source = source.String
		}
		item.DisclosedAt = disclosedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminListInviteRecords(status string, page int, pageSize int) ([]model.InviteRecord, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := ""
	if status != "" {
		filter = " WHERE status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM invite_records"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, inviter_user_id, invitee_user_id, status, risk_flag
FROM invite_records` + filter + `
ORDER BY register_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.InviteRecord, 0)
	for rows.Next() {
		var item model.InviteRecord
		if err := rows.Scan(&item.ID, &item.InviterUser, &item.InviteeUser, &item.Status, &item.RiskFlag); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}
