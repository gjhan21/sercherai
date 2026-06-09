package repo

import (
	"fmt"
	"math"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminPrecomputeStockPatternMatches(lookback int) error {
	if r.db == nil {
		return fmt.Errorf("database unavailable")
	}

	// 1. Query candidate stocks (top 300 with enough data)
	candidateStocks := []string{}
	stockRows, err := r.db.Query(`
		SELECT instrument_key 
		FROM market_daily_bars 
		WHERE asset_class='STOCK' AND source_key='TUSHARE' 
		GROUP BY instrument_key 
		HAVING COUNT(*) >= ? 
		ORDER BY COUNT(*) DESC 
		LIMIT 300
	`, lookback+7)
	if err != nil {
		return fmt.Errorf("query candidates failed: %w", err)
	}
	for stockRows.Next() {
		var s string
		if err := stockRows.Scan(&s); err == nil {
			candidateStocks = append(candidateStocks, s)
		}
	}
	stockRows.Close()

	if len(candidateStocks) == 0 {
		return fmt.Errorf("no candidate stocks found")
	}

	// Make a map of candidates for fast lookup
	candidateSet := make(map[string]bool)
	for _, cs := range candidateStocks {
		candidateSet[cs] = true
	}

	// 2. Load all prices and dates for these candidate stocks in one go
	type barPoint struct {
		Price float64
		Date  string
	}
	stockData := make(map[string][]barPoint)

	rows, err := r.db.Query(`
		SELECT instrument_key, close_price, trade_date 
		FROM market_daily_bars 
		WHERE asset_class='STOCK' AND source_key='TUSHARE' 
		ORDER BY instrument_key, trade_date ASC
	`)
	if err != nil {
		return fmt.Errorf("query bars failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sym string
		var p float64
		var t time.Time
		if err := rows.Scan(&sym, &p, &t); err == nil {
			if candidateSet[sym] {
				stockData[sym] = append(stockData[sym], barPoint{
					Price: p,
					Date:  t.Format("2006-01-02"),
				})
			}
		}
	}

	// 3. For each candidate stock, calculate the Top 10 matches and save them
	allMatches := []model.StockPatternMatch{}

	for _, sourceSym := range candidateStocks {
		sourcePoints := stockData[sourceSym]
		if len(sourcePoints) < lookback {
			continue
		}
		// Current segment: last lookback points of the source stock
		currentSeg := sourcePoints[len(sourcePoints)-lookback:]
		baseSeg := currentSeg[0].Price
		if baseSeg <= 0 {
			continue
		}

		type matchCandidate struct {
			stock      string
			date       string
			similarity float64
			next       []float64
		}
		var matches []matchCandidate

		for _, matchSym := range candidateStocks {
			if sourceSym == matchSym {
				continue
			}
			matchPoints := stockData[matchSym]
			if len(matchPoints) < lookback+7 {
				continue
			}

			// Sliding window similarity search
			for i := 0; i <= len(matchPoints)-lookback-7; i++ {
				sb := matchPoints[i].Price
				if sb <= 0 {
					continue
				}
				dot, n1, n2 := 0.0, 0.0, 0.0
				for j := 0; j < lookback; j++ {
					nv := matchPoints[i+j].Price / sb * 100
					cv := currentSeg[j].Price / baseSeg * 100
					dot += cv * nv
					n1 += cv * cv
					n2 += nv * nv
				}
				sim := dot / (math.Sqrt(n1*n2) + 1e-10)
				if sim > 0.85 {
					next7 := make([]float64, 7)
					for j := 0; j < 7 && i+lookback+j < len(matchPoints); j++ {
						next7[j] = (matchPoints[i+lookback+j].Price/matchPoints[i+lookback-1].Price - 1) * 100
					}
					matches = append(matches, matchCandidate{
						stock:      matchSym,
						date:       matchPoints[i+lookback-1].Date,
						similarity: sim,
						next:       next7,
					})
				}
			}
		}

		// Sort matches by similarity descending
		for i := 0; i < len(matches); i++ {
			for j := i + 1; j < len(matches); j++ {
				if matches[j].similarity > matches[i].similarity {
					matches[i], matches[j] = matches[j], matches[i]
				}
			}
		}

		// Keep Top 10 matches
		limit := 10
		if len(matches) < limit {
			limit = len(matches)
		}
		for i := 0; i < limit; i++ {
			c := matches[i]
			allMatches = append(allMatches, model.StockPatternMatch{
				ID:           newID("spm"),
				SourceSymbol: sourceSym,
				MatchSymbol:  c.stock,
				MatchDate:    c.date,
				Similarity:   c.similarity,
				Lookback:     lookback,
				Next7d:       c.next,
			})
		}
	}

	// 4. Batch save matches
	if len(allMatches) > 0 {
		_, err := r.db.Exec("DELETE FROM stock_pattern_matches WHERE lookback = ?", lookback)
		if err != nil {
			return fmt.Errorf("failed to clear old matches: %w", err)
		}
		err = r.SavePatternMatches(allMatches)
		if err != nil {
			return fmt.Errorf("failed to save pattern matches: %w", err)
		}
	}

	return nil
}
