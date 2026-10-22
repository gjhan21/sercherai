package repo

import (
	"database/sql"
	"log"
	"strconv"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminUpdateStockRecommendationAIReview(id string, aiReviewContent string) error {
	_, err := r.db.Exec("UPDATE stock_recommendations SET ai_review_content = ? WHERE id = ?", aiReviewContent, id)
	return err
}

func (r *MySQLGrowthRepo) AddUserVirtualSandbox(userID string, recoID string, addPrice float64) error {
	id := newID("vs")
	_, err := r.db.Exec(
		"INSERT INTO user_virtual_sandbox (id, user_id, stock_recommendation_id, add_price, status, added_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, userID, recoID, addPrice, "ACTIVE", time.Now().UTC(),
	)
	return err
}

func (r *MySQLGrowthRepo) GetUserVirtualSandbox(userID string) ([]model.UserVirtualSandbox, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, stock_recommendation_id, add_price, status, added_at FROM user_virtual_sandbox WHERE user_id = ? ORDER BY added_at DESC",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sandboxes []model.UserVirtualSandbox
	for rows.Next() {
		var s model.UserVirtualSandbox
		var addedAt time.Time
		if err := rows.Scan(&s.ID, &s.UserID, &s.StockRecommendationID, &s.AddPrice, &s.Status, &addedAt); err != nil {
			return nil, err
		}
		s.AddedAt = addedAt.Format(time.RFC3339)
		sandboxes = append(sandboxes, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return sandboxes, nil
}

func (r *MySQLGrowthRepo) AdminListStockSimulatedPositions(status string, symbol string, page int, pageSize int) ([]model.StockSimulatedPosition, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	status = strings.TrimSpace(status)
	symbol = strings.TrimSpace(symbol)
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}
	if symbol != "" {
		filter += " AND symbol = ?"
		args = append(args, symbol)
	}
	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM stock_simulated_positions"+filter, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, reco_id, symbol, name, status, open_date, open_price, current_price, 
       COALESCE(close_date, ''), COALESCE(close_price, 0), COALESCE(take_profit_price, 0), COALESCE(stop_loss_price, 0),
       quantity, cost_basis, COALESCE(close_value, 0), return_rate, max_drawdown, hold_days, COALESCE(close_reason, ''), created_at
FROM stock_simulated_positions` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.StockSimulatedPosition, 0)
	for rows.Next() {
		var item model.StockSimulatedPosition
		var closeDate, closeReason sql.NullString
		var closePrice, takeProfitPrice, stopLossPrice, closeValue sql.NullFloat64
		var openDate time.Time
		var createdAt time.Time
		err := rows.Scan(
			&item.ID, &item.RecoID, &item.Symbol, &item.Name, &item.Status, &openDate, &item.OpenPrice, &item.CurrentPrice,
			&closeDate, &closePrice, &takeProfitPrice, &stopLossPrice, &item.Quantity, &item.CostBasis, &closeValue,
			&item.ReturnRate, &item.MaxDrawdown, &item.HoldDays, &closeReason, &createdAt,
		)
		if err != nil {
			return nil, 0, err
		}
		item.OpenDate = openDate.Format("2006-01-02")
		item.CreatedAt = createdAt.Format(time.RFC3339)
		if closeDate.Valid && closeDate.String != "" {
			item.CloseDate = closeDate.String
			if t, err := time.Parse("2006-01-02T15:04:05Z", closeDate.String); err == nil {
				item.CloseDate = t.Format("2006-01-02")
			} else if t, err := time.Parse("2006-01-02", closeDate.String); err == nil {
				item.CloseDate = t.Format("2006-01-02")
			}
		}
		if closePrice.Valid {
			item.ClosePrice = closePrice.Float64
		}
		if takeProfitPrice.Valid {
			item.TakeProfitPrice = takeProfitPrice.Float64
		}
		if stopLossPrice.Valid {
			item.StopLossPrice = stopLossPrice.Float64
		}
		if closeValue.Valid {
			item.CloseValue = closeValue.Float64
		}
		if closeReason.Valid {
			item.CloseReason = closeReason.String
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminGetStockSimulatedOverview() (model.StockSimulatedOverview, error) {
	var overview model.StockSimulatedOverview
	err := r.db.QueryRow("SELECT COUNT(*) FROM stock_simulated_positions").Scan(&overview.TotalTrades)
	if err != nil {
		return overview, err
	}
	err = r.db.QueryRow("SELECT COUNT(*) FROM stock_simulated_positions WHERE status = 'HOLDING'").Scan(&overview.ActiveHoldings)
	if err != nil {
		return overview, err
	}

	var closedTrades int
	err = r.db.QueryRow("SELECT COUNT(*) FROM stock_simulated_positions WHERE status = 'CLOSED'").Scan(&closedTrades)
	if err != nil {
		return overview, err
	}

	if closedTrades > 0 {
		var winTrades int
		err = r.db.QueryRow("SELECT COUNT(*) FROM stock_simulated_positions WHERE status = 'CLOSED' AND return_rate > 0").Scan(&winTrades)
		if err != nil {
			return overview, err
		}
		overview.WinRate = float64(winTrades) / float64(closedTrades)

		var avgReturn, maxProfit, maxLoss, avgHoldDays sql.NullFloat64
		err = r.db.QueryRow(`
SELECT AVG(return_rate), MAX(return_rate), MIN(return_rate), AVG(hold_days) 
FROM stock_simulated_positions 
WHERE status = 'CLOSED'`).Scan(&avgReturn, &maxProfit, &maxLoss, &avgHoldDays)
		if err != nil {
			return overview, err
		}
		if avgReturn.Valid {
			overview.AverageReturn = avgReturn.Float64
		}
		if maxProfit.Valid {
			overview.MaxProfitRate = maxProfit.Float64
		}
		if maxLoss.Valid {
			overview.MaxLossRate = maxLoss.Float64
		}
		if avgHoldDays.Valid {
			overview.AvgHoldDays = avgHoldDays.Float64
		}
	}

	var totalCost, totalCurrentValue sql.NullFloat64
	err = r.db.QueryRow(`
SELECT SUM(cost_basis), 
       SUM(CASE WHEN status = 'CLOSED' THEN close_value ELSE current_price * quantity END)
FROM stock_simulated_positions`).Scan(&totalCost, &totalCurrentValue)
	if err != nil {
		return overview, err
	}
	if totalCost.Valid && totalCost.Float64 > 0 && totalCurrentValue.Valid {
		overview.TotalReturn = (totalCurrentValue.Float64 - totalCost.Float64) / totalCost.Float64
	}

	return overview, nil
}

func (r *MySQLGrowthRepo) AdminAutoOpenSimulatedPositions(tradeDate string) error {
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	tDate, err := time.Parse("2006-01-02", tradeDate)
	if err != nil {
		return err
	}

	rows, err := r.db.Query(`
SELECT r.id, r.symbol, r.name, COALESCE(d.take_profit, ''), COALESCE(d.stop_loss, '')
FROM stock_recommendations r
LEFT JOIN stock_reco_details d ON r.id = d.reco_id
WHERE r.valid_from <= ? AND r.valid_to >= ? 
  AND r.id NOT IN (SELECT reco_id FROM stock_simulated_positions)`, tDate, tDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	type pendingPosition struct {
		recoID     string
		symbol     string
		name       string
		takeProfit float64
		stopLoss   float64
	}
	var pendings []pendingPosition
	for rows.Next() {
		var p pendingPosition
		var tpStr, slStr string
		if err := rows.Scan(&p.recoID, &p.symbol, &p.name, &tpStr, &slStr); err != nil {
			return err
		}
		p.takeProfit, _ = strconv.ParseFloat(tpStr, 64)
		p.stopLoss, _ = strconv.ParseFloat(slStr, 64)
		pendings = append(pendings, p)
	}

	for _, p := range pendings {
		var openPrice float64
		err = r.db.QueryRow(`
SELECT close_price 
FROM market_daily_bar_truth 
WHERE asset_class = 'STOCK' AND (instrument_key = ? OR instrument_key LIKE ?) AND trade_date = ? 
LIMIT 1`, p.symbol, p.symbol+".%", tDate).Scan(&openPrice)
		if err != nil {
			errLatest := r.db.QueryRow(`
SELECT close_price 
FROM market_daily_bar_truth 
WHERE asset_class = 'STOCK' AND (instrument_key = ? OR instrument_key LIKE ?) AND trade_date <= ? 
ORDER BY trade_date DESC LIMIT 1`, p.symbol, p.symbol+".%", tDate).Scan(&openPrice)
			if errLatest != nil {
				openPrice = 100.0
			}
		}

		quantity := 1000.0
		costBasis := openPrice * quantity
		posID := newID("sp")
		_, err = r.db.Exec(`
INSERT INTO stock_simulated_positions (id, reco_id, symbol, name, status, open_date, open_price, current_price, take_profit_price, stop_loss_price, quantity, cost_basis, return_rate, hold_days, created_at)
VALUES (?, ?, ?, ?, 'HOLDING', ?, ?, ?, ?, ?, ?, ?, 0.0000, 0, ?)`,
			posID, p.recoID, p.symbol, p.name, tDate, openPrice, openPrice, p.takeProfit, p.stopLoss, quantity, costBasis, time.Now())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminSettlementSimulatedPositions(tradeDate string) error {
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	tDate, err := time.Parse("2006-01-02", tradeDate)
	if err != nil {
		return err
	}

	rows, err := r.db.Query(`
SELECT p.id, p.reco_id, p.symbol, p.open_price, p.take_profit_price, p.stop_loss_price, p.quantity, p.hold_days, r.valid_to
FROM stock_simulated_positions p
JOIN stock_recommendations r ON p.reco_id = r.id
WHERE p.status = 'HOLDING'`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type holdingPos struct {
		id              string
		recoID          string
		symbol          string
		openPrice       float64
		takeProfitPrice float64
		stopLossPrice   float64
		quantity        float64
		holdDays        int
		validTo         time.Time
	}
	var holdings []holdingPos
	for rows.Next() {
		var h holdingPos
		var validToStr string
		if err := rows.Scan(&h.id, &h.recoID, &h.symbol, &h.openPrice, &h.takeProfitPrice, &h.stopLossPrice, &h.quantity, &h.holdDays, &validToStr); err != nil {
			return err
		}
		if t, parseErr := time.Parse("2006-01-02 15:04:05", validToStr); parseErr == nil {
			h.validTo = t
		} else if t, parseErr := time.Parse("2006-01-02T15:04:05Z", validToStr); parseErr == nil {
			h.validTo = t
		} else if t, parseErr := time.Parse("2006-01-02", validToStr); parseErr == nil {
			h.validTo = t
		}
		holdings = append(holdings, h)
	}

	for _, h := range holdings {
		var openPrice, highPrice, lowPrice, closePrice float64
		err = r.db.QueryRow(`
SELECT open_price, high_price, low_price, close_price
FROM market_daily_bar_truth
WHERE asset_class = 'STOCK' AND (instrument_key = ? OR instrument_key LIKE ?) AND trade_date = ?
LIMIT 1`, h.symbol, h.symbol+".%", tDate).Scan(&openPrice, &highPrice, &lowPrice, &closePrice)
		if err != nil {
			_, _ = r.db.Exec("UPDATE stock_simulated_positions SET hold_days = hold_days + 1 WHERE id = ?", h.id)
			continue
		}

		newHoldDays := h.holdDays + 1
		newReturnRate := (closePrice - h.openPrice) / h.openPrice

		var maxPriceSeen float64
		err = r.db.QueryRow(`
SELECT MAX(close_price)
FROM market_daily_bar_truth
WHERE asset_class = 'STOCK' AND (instrument_key = ? OR instrument_key LIKE ?) AND trade_date >= (SELECT open_date FROM stock_simulated_positions WHERE id = ?) AND trade_date <= ?`,
			h.symbol, h.symbol+".%", h.id, tDate).Scan(&maxPriceSeen)
		if err != nil || maxPriceSeen <= 0 {
			maxPriceSeen = closePrice
		}
		maxDrawdown := 0.0
		if maxPriceSeen > 0 {
			maxDrawdown = (maxPriceSeen - closePrice) / maxPriceSeen
			if maxDrawdown < 0 {
				maxDrawdown = 0
			}
		}

		isClosed := false
		closePriceVal := closePrice
		closeReason := ""

		if h.takeProfitPrice > 0 && highPrice >= h.takeProfitPrice {
			isClosed = true
			closePriceVal = h.takeProfitPrice
			closeReason = "TAKE_PROFIT"
		} else if h.stopLossPrice > 0 && lowPrice <= h.stopLossPrice {
			isClosed = true
			closePriceVal = h.stopLossPrice
			closeReason = "STOP_LOSS"
		} else if tDate.After(h.validTo) || tDate.Equal(h.validTo) {
			isClosed = true
			closePriceVal = closePrice
			closeReason = "EXPIRED"
		}

		if isClosed {
			finalReturn := (closePriceVal - h.openPrice) / h.openPrice
			closeValue := closePriceVal * h.quantity
			_, err = r.db.Exec(`
UPDATE stock_simulated_positions
SET status = 'CLOSED', current_price = ?, close_date = ?, close_price = ?, close_value = ?, return_rate = ?, max_drawdown = ?, hold_days = ?, close_reason = ?
WHERE id = ?`, closePrice, tDate, closePriceVal, closeValue, finalReturn, maxDrawdown, newHoldDays, closeReason, h.id)
		} else {
			_, err = r.db.Exec(`
UPDATE stock_simulated_positions
SET current_price = ?, return_rate = ?, max_drawdown = ?, hold_days = ?
WHERE id = ?`, closePrice, newReturnRate, maxDrawdown, newHoldDays, h.id)
		}
		if err != nil {
			log.Printf("[simulated-positions] settlement update failed for %s: %v", h.id, err)
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminListFuturesSimulatedPositions(status string, contract string, page int, pageSize int) ([]model.FuturesSimulatedPosition, int, error) {
	offset := (page - 1) * pageSize
	filter := " WHERE 1=1"
	var args []any

	status = strings.TrimSpace(status)
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}

	contract = strings.TrimSpace(contract)
	if contract != "" {
		filter += " AND contract = ?"
		args = append(args, contract)
	}

	var total int
	err := r.db.QueryRow("SELECT COUNT(*) FROM futures_simulated_positions"+filter, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, strategy_id, contract, name, direction, status,
	                 DATE_FORMAT(open_date, '%Y-%m-%d'), open_price, current_price,
	                 COALESCE(DATE_FORMAT(close_date, '%Y-%m-%d'), ''), COALESCE(close_price, 0),
	                 COALESCE(take_profit_price, 0), COALESCE(stop_loss_price, 0),
	                 quantity, cost_basis, COALESCE(close_value, 0),
	                 return_rate, max_drawdown, hold_days, COALESCE(close_reason, ''),
	                 DATE_FORMAT(created_at, '%Y-%m-%dT%H:%i:%sZ')
	          FROM futures_simulated_positions` + filter + `
	          ORDER BY open_date DESC, created_at DESC
	          LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var positions []model.FuturesSimulatedPosition
	for rows.Next() {
		var p model.FuturesSimulatedPosition
		err := rows.Scan(
			&p.ID, &p.StrategyID, &p.Contract, &p.Name, &p.Direction, &p.Status,
			&p.OpenDate, &p.OpenPrice, &p.CurrentPrice,
			&p.CloseDate, &p.ClosePrice, &p.TakeProfitPrice, &p.StopLossPrice,
			&p.Quantity, &p.CostBasis, &p.CloseValue,
			&p.ReturnRate, &p.MaxDrawdown, &p.HoldDays, &p.CloseReason,
			&p.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}
		positions = append(positions, p)
	}
	return positions, total, nil
}

func (r *MySQLGrowthRepo) AdminGetFuturesSimulatedOverview() (model.FuturesSimulatedOverview, error) {
	var overview model.FuturesSimulatedOverview

	err := r.db.QueryRow("SELECT COUNT(*) FROM futures_simulated_positions").Scan(&overview.TotalTrades)
	if err != nil {
		return overview, err
	}

	err = r.db.QueryRow("SELECT COUNT(*) FROM futures_simulated_positions WHERE status = 'HOLDING'").Scan(&overview.ActiveHoldings)
	if err != nil {
		return overview, err
	}

	var closedTrades int
	err = r.db.QueryRow("SELECT COUNT(*) FROM futures_simulated_positions WHERE status = 'CLOSED'").Scan(&closedTrades)
	if err != nil {
		return overview, err
	}

	if closedTrades > 0 {
		var winTrades int
		err = r.db.QueryRow("SELECT COUNT(*) FROM futures_simulated_positions WHERE status = 'CLOSED' AND return_rate > 0").Scan(&winTrades)
		if err == nil {
			overview.WinRate = float64(winTrades) / float64(closedTrades)
		}

		err = r.db.QueryRow("SELECT COALESCE(AVG(return_rate), 0) FROM futures_simulated_positions WHERE status = 'CLOSED'").Scan(&overview.AverageReturn)
		if err != nil {
			return overview, err
		}

		err = r.db.QueryRow("SELECT COALESCE(AVG(hold_days), 0) FROM futures_simulated_positions WHERE status = 'CLOSED'").Scan(&overview.AvgHoldDays)
		if err != nil {
			return overview, err
		}

		err = r.db.QueryRow("SELECT COALESCE(MAX(return_rate), 0) FROM futures_simulated_positions WHERE status = 'CLOSED'").Scan(&overview.MaxProfitRate)
		if err != nil {
			return overview, err
		}

		err = r.db.QueryRow("SELECT COALESCE(MIN(return_rate), 0) FROM futures_simulated_positions WHERE status = 'CLOSED'").Scan(&overview.MaxLossRate)
		if err != nil {
			return overview, err
		}
	}

	var totalCost, totalCurrentValue float64
	err = r.db.QueryRow(`
		SELECT COALESCE(SUM(cost_basis), 0),
		       COALESCE(SUM(CASE WHEN status = 'CLOSED' THEN close_value ELSE cost_basis * (1 + return_rate) END), 0)
		FROM futures_simulated_positions`).Scan(&totalCost, &totalCurrentValue)
	if err == nil && totalCost > 0 {
		overview.TotalReturn = (totalCurrentValue - totalCost) / totalCost
	}

	return overview, nil
}

func parseRangeValue(rangeStr string) float64 {
	rangeStr = strings.TrimSpace(rangeStr)
	if rangeStr == "" {
		return 0
	}
	if idx := strings.Index(rangeStr, "-"); idx != -1 {
		rangeStr = strings.TrimSpace(rangeStr[:idx])
	}
	val, _ := strconv.ParseFloat(rangeStr, 64)
	return val
}

func (r *MySQLGrowthRepo) AdminAutoOpenFuturesSimulatedPositions(tradeDate string) error {
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	tDate, err := time.Parse("2006-01-02", tradeDate)
	if err != nil {
		return err
	}

	rows, err := r.db.Query(`
		SELECT s.id, s.contract, s.name, s.direction, COALESCE(g.take_profit_range, ''), COALESCE(g.stop_loss_range, '')
		FROM futures_strategies s
		LEFT JOIN futures_guidances g ON s.contract = g.contract AND g.valid_to >= s.valid_from AND g.guidance_direction = s.direction
		WHERE s.valid_from <= ? AND s.valid_to >= ? AND s.status = 'PUBLISHED'
		  AND s.id NOT IN (SELECT strategy_id FROM futures_simulated_positions)`, tDate, tDate)
	if err != nil {
		return err
	}
	defer rows.Close()

	type pendingPosition struct {
		strategyID string
		contract   string
		name       string
		direction  string
		takeProfit float64
		stopLoss   float64
	}
	var pendings []pendingPosition
	for rows.Next() {
		var p pendingPosition
		var tpStr, slStr string
		if err := rows.Scan(&p.strategyID, &p.contract, &p.name, &p.direction, &tpStr, &slStr); err != nil {
			return err
		}
		p.takeProfit = parseRangeValue(tpStr)
		p.stopLoss = parseRangeValue(slStr)
		pendings = append(pendings, p)
	}

	for _, p := range pendings {
		var openPrice float64
		err = r.db.QueryRow(`
			SELECT close_price 
			FROM market_daily_bar_truth 
			WHERE asset_class = 'FUTURES' AND (instrument_key = ? OR instrument_key LIKE ?) AND trade_date = ? 
			LIMIT 1`, p.contract, p.contract+".%", tDate).Scan(&openPrice)
		if err != nil {
			errLatest := r.db.QueryRow(`
				SELECT close_price 
				FROM market_daily_bar_truth 
				WHERE asset_class = 'FUTURES' AND (instrument_key = ? OR instrument_key LIKE ?) AND trade_date <= ? 
				ORDER BY trade_date DESC LIMIT 1`, p.contract, p.contract+".%", tDate).Scan(&openPrice)
			if errLatest != nil {
				openPrice = 100.0
			}
		}

		quantity := 10.0
		costBasis := openPrice * quantity
		posID := newID("fp")
		_, err = r.db.Exec(`
			INSERT INTO futures_simulated_positions (id, strategy_id, contract, name, direction, status, open_date, open_price, current_price, take_profit_price, stop_loss_price, quantity, cost_basis, return_rate, hold_days, created_at)
			VALUES (?, ?, ?, ?, ?, 'HOLDING', ?, ?, ?, ?, ?, ?, ?, 0.0000, 0, ?)`,
			posID, p.strategyID, p.contract, p.name, p.direction, tDate, openPrice, openPrice, p.takeProfit, p.stopLoss, quantity, costBasis, time.Now())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminSettlementFuturesSimulatedPositions(tradeDate string) error {
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	tDate, err := time.Parse("2006-01-02", tradeDate)
	if err != nil {
		return err
	}

	rows, err := r.db.Query(`
		SELECT p.id, p.strategy_id, p.contract, p.direction, p.open_price, p.take_profit_price, p.stop_loss_price, p.quantity, p.cost_basis, p.hold_days, s.valid_to
		FROM futures_simulated_positions p
		JOIN futures_strategies s ON p.strategy_id = s.id
		WHERE p.status = 'HOLDING'`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type holdingPos struct {
		id              string
		strategyID      string
		contract        string
		direction       string
		openPrice       float64
		takeProfitPrice float64
		stopLossPrice   float64
		quantity        float64
		costBasis       float64
		holdDays        int
		validTo         time.Time
	}
	var holdings []holdingPos
	for rows.Next() {
		var h holdingPos
		var validToStr string
		if err := rows.Scan(&h.id, &h.strategyID, &h.contract, &h.direction, &h.openPrice, &h.takeProfitPrice, &h.stopLossPrice, &h.quantity, &h.costBasis, &h.holdDays, &validToStr); err != nil {
			return err
		}
		if t, parseErr := time.Parse("2006-01-02 15:04:05", validToStr); parseErr == nil {
			h.validTo = t
		} else if t, parseErr := time.Parse("2006-01-02T15:04:05Z", validToStr); parseErr == nil {
			h.validTo = t
		} else if t, parseErr := time.Parse("2006-01-02", validToStr); parseErr == nil {
			h.validTo = t
		}
		holdings = append(holdings, h)
	}

	for _, h := range holdings {
		var openPrice, highPrice, lowPrice, closePrice float64
		err = r.db.QueryRow(`
			SELECT open_price, high_price, low_price, close_price
			FROM market_daily_bar_truth
			WHERE asset_class = 'FUTURES' AND (instrument_key = ? OR instrument_key LIKE ?) AND trade_date = ?
			LIMIT 1`, h.contract, h.contract+".%", tDate).Scan(&openPrice, &highPrice, &lowPrice, &closePrice)
		if err != nil {
			_, _ = r.db.Exec("UPDATE futures_simulated_positions SET hold_days = hold_days + 1 WHERE id = ?", h.id)
			continue
		}

		newHoldDays := h.holdDays + 1
		var newReturnRate float64
		var maxDrawdown float64

		if strings.ToUpper(h.direction) == "SHORT" {
			newReturnRate = (h.openPrice - closePrice) / h.openPrice

			var minPriceSeen float64
			err = r.db.QueryRow(`
				SELECT MIN(close_price)
				FROM market_daily_bar_truth
				WHERE asset_class = 'FUTURES' AND (instrument_key = ? OR instrument_key LIKE ?) AND trade_date >= (SELECT open_date FROM futures_simulated_positions WHERE id = ?) AND trade_date <= ?`,
				h.contract, h.contract+".%", h.id, tDate).Scan(&minPriceSeen)
			if err != nil || minPriceSeen <= 0 {
				minPriceSeen = closePrice
			}
			if minPriceSeen > 0 {
				maxDrawdown = (closePrice - minPriceSeen) / h.openPrice
				if maxDrawdown < 0 {
					maxDrawdown = 0
				}
			}
		} else {
			newReturnRate = (closePrice - h.openPrice) / h.openPrice

			var maxPriceSeen float64
			err = r.db.QueryRow(`
				SELECT MAX(close_price)
				FROM market_daily_bar_truth
				WHERE asset_class = 'FUTURES' AND (instrument_key = ? OR instrument_key LIKE ?) AND trade_date >= (SELECT open_date FROM futures_simulated_positions WHERE id = ?) AND trade_date <= ?`,
				h.contract, h.contract+".%", h.id, tDate).Scan(&maxPriceSeen)
			if err != nil || maxPriceSeen <= 0 {
				maxPriceSeen = closePrice
			}
			if maxPriceSeen > 0 {
				maxDrawdown = (maxPriceSeen - closePrice) / maxPriceSeen
				if maxDrawdown < 0 {
					maxDrawdown = 0
				}
			}
		}

		isClosed := false
		closePriceVal := closePrice
		closeReason := ""

		if strings.ToUpper(h.direction) == "SHORT" {
			if h.takeProfitPrice > 0 && lowPrice <= h.takeProfitPrice {
				isClosed = true
				closePriceVal = h.takeProfitPrice
				closeReason = "TAKE_PROFIT"
			} else if h.stopLossPrice > 0 && highPrice >= h.stopLossPrice {
				isClosed = true
				closePriceVal = h.stopLossPrice
				closeReason = "STOP_LOSS"
			} else if tDate.After(h.validTo) || tDate.Equal(h.validTo) {
				isClosed = true
				closePriceVal = closePrice
				closeReason = "EXPIRED"
			}
		} else {
			if h.takeProfitPrice > 0 && highPrice >= h.takeProfitPrice {
				isClosed = true
				closePriceVal = h.takeProfitPrice
				closeReason = "TAKE_PROFIT"
			} else if h.stopLossPrice > 0 && lowPrice <= h.stopLossPrice {
				isClosed = true
				closePriceVal = h.stopLossPrice
				closeReason = "STOP_LOSS"
			} else if tDate.After(h.validTo) || tDate.Equal(h.validTo) {
				isClosed = true
				closePriceVal = closePrice
				closeReason = "EXPIRED"
			}
		}

		if isClosed {
			var finalReturn float64
			var closeValue float64
			if strings.ToUpper(h.direction) == "SHORT" {
				finalReturn = (h.openPrice - closePriceVal) / h.openPrice
				closeValue = h.costBasis * (1 + finalReturn)
			} else {
				finalReturn = (closePriceVal - h.openPrice) / h.openPrice
				closeValue = closePriceVal * h.quantity
			}

			_, err = r.db.Exec(`
				UPDATE futures_simulated_positions
				SET status = 'CLOSED', current_price = ?, close_date = ?, close_price = ?, close_value = ?, return_rate = ?, max_drawdown = ?, hold_days = ?, close_reason = ?
				WHERE id = ?`, closePrice, tDate, closePriceVal, closeValue, finalReturn, maxDrawdown, newHoldDays, closeReason, h.id)
		} else {
			_, err = r.db.Exec(`
				UPDATE futures_simulated_positions
				SET current_price = ?, return_rate = ?, max_drawdown = ?, hold_days = ?
				WHERE id = ?`, closePrice, newReturnRate, maxDrawdown, newHoldDays, h.id)
		}
		if err != nil {
			log.Printf("[simulated-positions] futures settlement update failed for %s: %v", h.id, err)
		}
	}
	return nil
}
