package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := ""
	if status != "" {
		filter = " WHERE status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM membership_products"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, name, price, status, member_level, duration_days
FROM membership_products` + filter + `
ORDER BY id DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.MembershipProduct, 0)
	for rows.Next() {
		var item model.MembershipProduct
		var memberLevel sql.NullString
		var durationDays sql.NullInt64
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Status, &memberLevel, &durationDays); err != nil {
			return nil, 0, err
		}
		if memberLevel.Valid {
			item.MemberLevel = memberLevel.String
		}
		if durationDays.Valid {
			item.DurationDays = int(durationDays.Int64)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateMembershipProduct(name string, price float64, status string, memberLevel string, durationDays int) (string, error) {
	id := newID("mp")
	level := strings.ToUpper(strings.TrimSpace(memberLevel))
	if level == "" {
		level = "VIP1"
	}
	if err := r.ensureMembershipLevelSynced(level); err != nil {
		return "", err
	}
	days := durationDays
	if days <= 0 {
		days = 30
	}
	_, err := r.db.Exec(
		"INSERT INTO membership_products (id, name, price, status, member_level, duration_days) VALUES (?, ?, ?, ?, ?, ?)",
		id,
		strings.TrimSpace(name),
		price,
		strings.ToUpper(strings.TrimSpace(status)),
		level,
		days,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateMembershipProduct(id string, name string, price float64, status string, memberLevel string, durationDays int) error {
	level := strings.ToUpper(strings.TrimSpace(memberLevel))
	if level == "" {
		return errors.New("member level is required")
	}
	if err := r.ensureMembershipLevelSynced(level); err != nil {
		return err
	}
	days := durationDays
	if days <= 0 {
		days = 30
	}
	res, err := r.db.Exec(
		`UPDATE membership_products
SET name = ?, price = ?, status = ?, member_level = ?, duration_days = ?
WHERE id = ?`,
		strings.TrimSpace(name),
		price,
		strings.ToUpper(strings.TrimSpace(status)),
		level,
		days,
		id,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminUpdateMembershipProductStatus(id string, status string) error {
	res, err := r.db.Exec("UPDATE membership_products SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminDeleteMembershipProduct(id string) error {
	result, err := r.db.Exec("DELETE FROM membership_products WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) ensureMembershipLevelSynced(memberLevel string) error {
	level := strings.ToUpper(strings.TrimSpace(memberLevel))
	if level == "" {
		return errors.New("member level is required")
	}
	var count int
	if err := r.db.QueryRow(
		`SELECT COUNT(*)
FROM vip_quota_configs
WHERE member_level = ? AND status = 'ACTIVE'`,
		level,
	).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf("member level %s has no active vip quota config", level)
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminListMembershipOrders(status string, userID string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}
	if userID != "" {
		filter += " AND user_id = ?"
		args = append(args, userID)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM membership_orders"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, order_no, user_id, product_id, amount, pay_channel, status, paid_at, created_at
FROM membership_orders` + filter + `
ORDER BY created_at DESC, id DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.MembershipOrderAdmin, 0)
	for rows.Next() {
		var item model.MembershipOrderAdmin
		var paidAt, createdAt sql.NullTime
		var orderNo, payChannel sql.NullString
		if err := rows.Scan(&item.ID, &orderNo, &item.UserID, &item.ProductID, &item.Amount, &payChannel, &item.Status, &paidAt, &createdAt); err != nil {
			return nil, 0, err
		}
		if orderNo.Valid {
			item.OrderNo = orderNo.String
		}
		if payChannel.Valid {
			item.PayChannel = payChannel.String
		}
		if paidAt.Valid {
			item.PaidAt = paidAt.Time.Format(time.RFC3339)
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminUpdateMembershipOrderStatus(id string, status string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	var orderNo, previousStatus string
	if err := tx.QueryRow("SELECT order_no, status FROM membership_orders WHERE id = ? FOR UPDATE", id).Scan(&orderNo, &previousStatus); err != nil {
		_ = tx.Rollback()
		return err
	}
	now := time.Now()
	if _, err := tx.Exec("UPDATE membership_orders SET status = ?, updated_at = ? WHERE id = ?", status, now, id); err != nil {
		_ = tx.Rollback()
		return err
	}
	if strings.ToUpper(strings.TrimSpace(status)) == "PAID" && strings.ToUpper(strings.TrimSpace(previousStatus)) != "PAID" {
		if err := createExperimentSuccessEventTx(tx, strings.TrimSpace(orderNo), now); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *MySQLGrowthRepo) AdminListMarketRhythmTasks(taskDate string) ([]model.MarketRhythmTask, error) {
	normalizedDate, err := normalizeMarketRhythmDate(taskDate)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(`
SELECT id, task_date, slot, task_key, status, COALESCE(owner, ''), COALESCE(notes, ''), source_links_json, completed_at, created_at, updated_at
FROM market_rhythm_tasks
WHERE task_date = ?
ORDER BY FIELD(slot, '08:30', '11:30', '15:30', '周末'), task_key ASC, created_at ASC`, normalizedDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.MarketRhythmTask, 0)
	for rows.Next() {
		item, err := scanMarketRhythmTask(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MySQLGrowthRepo) AdminEnsureMarketRhythmTasks(taskDate string) ([]model.MarketRhythmTask, error) {
	normalizedDate, err := normalizeMarketRhythmDate(taskDate)
	if err != nil {
		return nil, err
	}
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for _, template := range defaultMarketRhythmTaskTemplates() {
		if _, err := tx.Exec(`
INSERT INTO market_rhythm_tasks (
	id, task_date, slot, task_key, status, owner, notes, source_links_json, completed_at, created_at, updated_at
) VALUES (?, ?, ?, ?, 'TODO', '', '', NULL, NULL, ?, ?)
ON DUPLICATE KEY UPDATE updated_at = updated_at`,
			newID("mrt"),
			normalizedDate,
			template.Slot,
			template.TaskKey,
			now,
			now,
		); err != nil {
			_ = tx.Rollback()
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.AdminListMarketRhythmTasks(normalizedDate)
}

func (r *MySQLGrowthRepo) AdminUpdateMarketRhythmTask(id string, owner string, notes string, sourceLinks []string, status string) (model.MarketRhythmTask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.MarketRhythmTask{}, err
	}
	item, err := getMarketRhythmTaskByIDTx(tx, id)
	if err != nil {
		_ = tx.Rollback()
		return model.MarketRhythmTask{}, err
	}
	nextStatus := normalizeMarketRhythmStatus(status)
	if nextStatus == "" {
		nextStatus = normalizeMarketRhythmStatus(item.Status)
	}
	sourceLinksJSON, err := marshalStringList(normalizeStringList(sourceLinks))
	if err != nil {
		_ = tx.Rollback()
		return model.MarketRhythmTask{}, err
	}
	completedAt := marketRhythmCompletedAtValue(item.CompletedAt, nextStatus)
	now := time.Now()
	if _, err := tx.Exec(`
UPDATE market_rhythm_tasks
SET owner = ?, notes = ?, source_links_json = ?, status = ?, completed_at = ?, updated_at = ?
WHERE id = ?`,
		strings.TrimSpace(owner),
		strings.TrimSpace(notes),
		nullableString(sourceLinksJSON),
		nextStatus,
		completedAt,
		now,
		id,
	); err != nil {
		_ = tx.Rollback()
		return model.MarketRhythmTask{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.MarketRhythmTask{}, err
	}
	return r.adminGetMarketRhythmTaskByID(id)
}

func (r *MySQLGrowthRepo) AdminUpdateMarketRhythmTaskStatus(id string, status string, owner string, notes string) (model.MarketRhythmTask, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.MarketRhythmTask{}, err
	}
	item, err := getMarketRhythmTaskByIDTx(tx, id)
	if err != nil {
		_ = tx.Rollback()
		return model.MarketRhythmTask{}, err
	}
	nextStatus := normalizeMarketRhythmStatus(status)
	if nextStatus == "" {
		nextStatus = normalizeMarketRhythmStatus(item.Status)
	}
	nextOwner := item.Owner
	if strings.TrimSpace(owner) != "" {
		nextOwner = strings.TrimSpace(owner)
	}
	nextNotes := item.Notes
	if strings.TrimSpace(notes) != "" {
		nextNotes = strings.TrimSpace(notes)
	}
	sourceLinksJSON, err := marshalStringList(normalizeStringList(item.SourceLinks))
	if err != nil {
		_ = tx.Rollback()
		return model.MarketRhythmTask{}, err
	}
	completedAt := marketRhythmCompletedAtValue(item.CompletedAt, nextStatus)
	now := time.Now()
	if _, err := tx.Exec(`
UPDATE market_rhythm_tasks
SET owner = ?, notes = ?, source_links_json = ?, status = ?, completed_at = ?, updated_at = ?
WHERE id = ?`,
		nextOwner,
		nextNotes,
		nullableString(sourceLinksJSON),
		nextStatus,
		completedAt,
		now,
		id,
	); err != nil {
		_ = tx.Rollback()
		return model.MarketRhythmTask{}, err
	}
	if err := tx.Commit(); err != nil {
		return model.MarketRhythmTask{}, err
	}
	return r.adminGetMarketRhythmTaskByID(id)
}

func (r *MySQLGrowthRepo) AdminGetExperimentAnalyticsSummary(days int) (model.AdminExperimentAnalyticsSummary, error) {
	if days <= 0 {
		days = 7
	}
	since := time.Now().AddDate(0, 0, -days)
	summary := model.AdminExperimentAnalyticsSummary{
		Days:                days,
		Overview:            model.AdminExperimentAnalyticsOverview{Days: days},
		Items:               make([]model.AdminExperimentAnalyticsItem, 0),
		PageBreakdown:       make([]model.AdminExperimentAnalyticsPageItem, 0),
		DailyTrend:          make([]model.AdminExperimentAnalyticsTrendPoint, 0),
		PayChannelBreakdown: make([]model.AdminExperimentAnalyticsPayChannelItem, 0),
		DeviceBreakdown:     make([]model.AdminExperimentAnalyticsDeviceItem, 0),
		UserStageBreakdown:  make([]model.AdminExperimentAnalyticsUserStageItem, 0),
		VariantDailyTrend:   make([]model.AdminExperimentAnalyticsVariantTrendPoint, 0),
	}
	var lastEventAt sql.NullTime
	if err := r.db.QueryRow(`
SELECT
	COUNT(*) AS total_events,
	COUNT(DISTINCT experiment_key) AS total_experiments,
	COALESCE(SUM(CASE WHEN event_type = 'EXPOSURE' THEN 1 ELSE 0 END), 0) AS exposure_count,
	COALESCE(SUM(CASE WHEN event_type = 'CLICK' THEN 1 ELSE 0 END), 0) AS click_count,
	COALESCE(SUM(CASE WHEN event_type = 'UPGRADE_INTENT' THEN 1 ELSE 0 END), 0) AS upgrade_intent_count,
	COALESCE(SUM(CASE WHEN event_type = 'PAYMENT_SUCCESS' THEN 1 ELSE 0 END), 0) AS payment_success_count,
	COALESCE(SUM(CASE WHEN event_type = 'RENEWAL_SUCCESS' THEN 1 ELSE 0 END), 0) AS renewal_success_count,
	MAX(created_at) AS last_event_at
FROM experiment_events
WHERE created_at >= ?`, since).Scan(
		&summary.Overview.TotalEvents,
		&summary.Overview.TotalExperiments,
		&summary.Overview.ExposureCount,
		&summary.Overview.ClickCount,
		&summary.Overview.UpgradeIntentCount,
		&summary.Overview.PaymentSuccessCount,
		&summary.Overview.RenewalSuccessCount,
		&lastEventAt,
	); err != nil {
		return summary, err
	}
	paidSuccessTotal := summary.Overview.PaymentSuccessCount + summary.Overview.RenewalSuccessCount
	summary.Overview.ClickThroughRate = safeRatioValue(int64(summary.Overview.ClickCount), int64(summary.Overview.ExposureCount))
	summary.Overview.UpgradePerClickRate = safeRatioValue(int64(summary.Overview.UpgradeIntentCount), int64(summary.Overview.ClickCount))
	summary.Overview.UpgradePerExposureRate = safeRatioValue(int64(summary.Overview.UpgradeIntentCount), int64(summary.Overview.ExposureCount))
	summary.Overview.PaidPerUpgradeRate = safeRatioValue(int64(paidSuccessTotal), int64(summary.Overview.UpgradeIntentCount))
	summary.Overview.PaidPerClickRate = safeRatioValue(int64(paidSuccessTotal), int64(summary.Overview.ClickCount))
	summary.Overview.PaidPerExposureRate = safeRatioValue(int64(paidSuccessTotal), int64(summary.Overview.ExposureCount))
	if lastEventAt.Valid {
		summary.Overview.LastEventAt = lastEventAt.Time.Format(time.RFC3339)
	}

	rows, err := r.db.Query(`
SELECT
	experiment_key,
	variant_key,
	page_key,
	COALESCE(NULLIF(user_stage, ''), 'UNKNOWN') AS user_stage,
	COALESCE(SUM(CASE WHEN event_type = 'EXPOSURE' THEN 1 ELSE 0 END), 0) AS exposure_count,
	COALESCE(SUM(CASE WHEN event_type = 'CLICK' THEN 1 ELSE 0 END), 0) AS click_count,
	COALESCE(SUM(CASE WHEN event_type = 'UPGRADE_INTENT' THEN 1 ELSE 0 END), 0) AS upgrade_intent_count,
	COALESCE(SUM(CASE WHEN event_type = 'PAYMENT_SUCCESS' THEN 1 ELSE 0 END), 0) AS payment_success_count,
	COALESCE(SUM(CASE WHEN event_type = 'RENEWAL_SUCCESS' THEN 1 ELSE 0 END), 0) AS renewal_success_count,
	MAX(created_at) AS last_event_at
FROM experiment_events
WHERE created_at >= ?
GROUP BY experiment_key, variant_key, page_key, COALESCE(NULLIF(user_stage, ''), 'UNKNOWN')
ORDER BY upgrade_intent_count DESC, click_count DESC, exposure_count DESC, last_event_at DESC
LIMIT 100`, since)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.AdminExperimentAnalyticsItem
		var itemLastEventAt sql.NullTime
		if err := rows.Scan(
			&item.ExperimentKey,
			&item.VariantKey,
			&item.PageKey,
			&item.UserStage,
			&item.ExposureCount,
			&item.ClickCount,
			&item.UpgradeIntentCount,
			&item.PaymentSuccessCount,
			&item.RenewalSuccessCount,
			&itemLastEventAt,
		); err != nil {
			return summary, err
		}
		paidSuccessTotal := item.PaymentSuccessCount + item.RenewalSuccessCount
		item.ClickThroughRate = safeRatioValue(int64(item.ClickCount), int64(item.ExposureCount))
		item.UpgradePerClickRate = safeRatioValue(int64(item.UpgradeIntentCount), int64(item.ClickCount))
		item.UpgradePerExposureRate = safeRatioValue(int64(item.UpgradeIntentCount), int64(item.ExposureCount))
		item.PaidPerUpgradeRate = safeRatioValue(int64(paidSuccessTotal), int64(item.UpgradeIntentCount))
		item.PaidPerClickRate = safeRatioValue(int64(paidSuccessTotal), int64(item.ClickCount))
		item.PaidPerExposureRate = safeRatioValue(int64(paidSuccessTotal), int64(item.ExposureCount))
		if itemLastEventAt.Valid {
			item.LastEventAt = itemLastEventAt.Time.Format(time.RFC3339)
		}
		summary.Items = append(summary.Items, item)
		summary.UserStageBreakdown = append(summary.UserStageBreakdown, model.AdminExperimentAnalyticsUserStageItem{
			ExperimentKey:          item.ExperimentKey,
			VariantKey:             item.VariantKey,
			PageKey:                item.PageKey,
			UserStage:              item.UserStage,
			ExposureCount:          item.ExposureCount,
			ClickCount:             item.ClickCount,
			UpgradeIntentCount:     item.UpgradeIntentCount,
			PaymentSuccessCount:    item.PaymentSuccessCount,
			RenewalSuccessCount:    item.RenewalSuccessCount,
			ClickThroughRate:       item.ClickThroughRate,
			UpgradePerClickRate:    item.UpgradePerClickRate,
			UpgradePerExposureRate: item.UpgradePerExposureRate,
			PaidPerUpgradeRate:     item.PaidPerUpgradeRate,
			PaidPerClickRate:       item.PaidPerClickRate,
			PaidPerExposureRate:    item.PaidPerExposureRate,
			LastEventAt:            item.LastEventAt,
		})
	}
	if err := rows.Err(); err != nil {
		return summary, err
	}

	pageRows, err := r.db.Query(`
SELECT
	page_key,
	COALESCE(SUM(CASE WHEN event_type = 'EXPOSURE' THEN 1 ELSE 0 END), 0) AS exposure_count,
	COALESCE(SUM(CASE WHEN event_type = 'CLICK' THEN 1 ELSE 0 END), 0) AS click_count,
	COALESCE(SUM(CASE WHEN event_type = 'UPGRADE_INTENT' THEN 1 ELSE 0 END), 0) AS upgrade_intent_count,
	COALESCE(SUM(CASE WHEN event_type = 'PAYMENT_SUCCESS' THEN 1 ELSE 0 END), 0) AS payment_success_count,
	COALESCE(SUM(CASE WHEN event_type = 'RENEWAL_SUCCESS' THEN 1 ELSE 0 END), 0) AS renewal_success_count,
	MAX(created_at) AS last_event_at
FROM experiment_events
WHERE created_at >= ?
GROUP BY page_key
ORDER BY payment_success_count DESC, upgrade_intent_count DESC, click_count DESC, exposure_count DESC, last_event_at DESC
LIMIT 20`, since)
	if err != nil {
		return summary, err
	}
	defer pageRows.Close()

	for pageRows.Next() {
		var item model.AdminExperimentAnalyticsPageItem
		var itemLastEventAt sql.NullTime
		if err := pageRows.Scan(
			&item.PageKey,
			&item.ExposureCount,
			&item.ClickCount,
			&item.UpgradeIntentCount,
			&item.PaymentSuccessCount,
			&item.RenewalSuccessCount,
			&itemLastEventAt,
		); err != nil {
			return summary, err
		}
		paidSuccessTotal := item.PaymentSuccessCount + item.RenewalSuccessCount
		item.ClickThroughRate = safeRatioValue(int64(item.ClickCount), int64(item.ExposureCount))
		item.UpgradePerClickRate = safeRatioValue(int64(item.UpgradeIntentCount), int64(item.ClickCount))
		item.PaidPerUpgradeRate = safeRatioValue(int64(paidSuccessTotal), int64(item.UpgradeIntentCount))
		item.PaidPerExposureRate = safeRatioValue(int64(paidSuccessTotal), int64(item.ExposureCount))
		if itemLastEventAt.Valid {
			item.LastEventAt = itemLastEventAt.Time.Format(time.RFC3339)
		}
		summary.PageBreakdown = append(summary.PageBreakdown, item)
	}
	if err := pageRows.Err(); err != nil {
		return summary, err
	}

	trendRows, err := r.db.Query(`
SELECT
	DATE(created_at) AS metric_date,
	COALESCE(SUM(CASE WHEN event_type = 'EXPOSURE' THEN 1 ELSE 0 END), 0) AS exposure_count,
	COALESCE(SUM(CASE WHEN event_type = 'CLICK' THEN 1 ELSE 0 END), 0) AS click_count,
	COALESCE(SUM(CASE WHEN event_type = 'UPGRADE_INTENT' THEN 1 ELSE 0 END), 0) AS upgrade_intent_count,
	COALESCE(SUM(CASE WHEN event_type = 'PAYMENT_SUCCESS' THEN 1 ELSE 0 END), 0) AS payment_success_count,
	COALESCE(SUM(CASE WHEN event_type = 'RENEWAL_SUCCESS' THEN 1 ELSE 0 END), 0) AS renewal_success_count
FROM experiment_events
WHERE created_at >= ?
GROUP BY DATE(created_at)
ORDER BY metric_date ASC`, since)
	if err != nil {
		return summary, err
	}
	defer trendRows.Close()

	for trendRows.Next() {
		var metricDate time.Time
		var point model.AdminExperimentAnalyticsTrendPoint
		if err := trendRows.Scan(
			&metricDate,
			&point.ExposureCount,
			&point.ClickCount,
			&point.UpgradeIntentCount,
			&point.PaymentSuccessCount,
			&point.RenewalSuccessCount,
		); err != nil {
			return summary, err
		}
		paidSuccessTotal := point.PaymentSuccessCount + point.RenewalSuccessCount
		point.Date = metricDate.Format("2006-01-02")
		point.ClickThroughRate = safeRatioValue(int64(point.ClickCount), int64(point.ExposureCount))
		point.UpgradePerClickRate = safeRatioValue(int64(point.UpgradeIntentCount), int64(point.ClickCount))
		point.PaidPerExposureRate = safeRatioValue(int64(paidSuccessTotal), int64(point.ExposureCount))
		summary.DailyTrend = append(summary.DailyTrend, point)
	}
	if err := trendRows.Err(); err != nil {
		return summary, err
	}

	payRows, err := r.db.Query(`
SELECT pay_channel, payment_success_count, renewal_success_count, last_event_at
FROM (
	SELECT
		COALESCE(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(metadata_json, '$.pay_channel')), ''), 'UNKNOWN') AS pay_channel,
		COALESCE(SUM(CASE WHEN event_type = 'PAYMENT_SUCCESS' THEN 1 ELSE 0 END), 0) AS payment_success_count,
		COALESCE(SUM(CASE WHEN event_type = 'RENEWAL_SUCCESS' THEN 1 ELSE 0 END), 0) AS renewal_success_count,
		MAX(created_at) AS last_event_at
	FROM experiment_events
	WHERE created_at >= ?
	  AND event_type IN ('PAYMENT_SUCCESS', 'RENEWAL_SUCCESS')
	GROUP BY COALESCE(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(metadata_json, '$.pay_channel')), ''), 'UNKNOWN')
) pay_summary
ORDER BY (payment_success_count + renewal_success_count) DESC, last_event_at DESC`, since)
	if err != nil {
		return summary, err
	}
	defer payRows.Close()

	for payRows.Next() {
		var item model.AdminExperimentAnalyticsPayChannelItem
		var itemLastEventAt sql.NullTime
		if err := payRows.Scan(
			&item.PayChannel,
			&item.PaymentSuccessCount,
			&item.RenewalSuccessCount,
			&itemLastEventAt,
		); err != nil {
			return summary, err
		}
		item.PaidSuccessCount = item.PaymentSuccessCount + item.RenewalSuccessCount
		item.PaidShareRate = safeRatioValue(int64(item.PaidSuccessCount), int64(paidSuccessTotal))
		if itemLastEventAt.Valid {
			item.LastEventAt = itemLastEventAt.Time.Format(time.RFC3339)
		}
		summary.PayChannelBreakdown = append(summary.PayChannelBreakdown, item)
	}
	if err := payRows.Err(); err != nil {
		return summary, err
	}

	deviceRows, err := r.db.Query(`
SELECT
	experiment_key,
	variant_key,
	page_key,
	device_type,
	exposure_count,
	click_count,
	upgrade_intent_count,
	payment_success_count,
	renewal_success_count,
	last_event_at
FROM (
	SELECT
		experiment_key,
		variant_key,
		page_key,
		COALESCE(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(metadata_json, '$.device_type')), ''), 'UNKNOWN') AS device_type,
		COALESCE(SUM(CASE WHEN event_type = 'EXPOSURE' THEN 1 ELSE 0 END), 0) AS exposure_count,
		COALESCE(SUM(CASE WHEN event_type = 'CLICK' THEN 1 ELSE 0 END), 0) AS click_count,
		COALESCE(SUM(CASE WHEN event_type = 'UPGRADE_INTENT' THEN 1 ELSE 0 END), 0) AS upgrade_intent_count,
		COALESCE(SUM(CASE WHEN event_type = 'PAYMENT_SUCCESS' THEN 1 ELSE 0 END), 0) AS payment_success_count,
		COALESCE(SUM(CASE WHEN event_type = 'RENEWAL_SUCCESS' THEN 1 ELSE 0 END), 0) AS renewal_success_count,
		MAX(created_at) AS last_event_at
	FROM experiment_events
	WHERE created_at >= ?
	GROUP BY experiment_key, variant_key, page_key, COALESCE(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(metadata_json, '$.device_type')), ''), 'UNKNOWN')
) device_summary
ORDER BY (payment_success_count + renewal_success_count) DESC, upgrade_intent_count DESC, click_count DESC, exposure_count DESC, last_event_at DESC
LIMIT 120`, since)
	if err != nil {
		return summary, err
	}
	defer deviceRows.Close()

	for deviceRows.Next() {
		var item model.AdminExperimentAnalyticsDeviceItem
		var itemLastEventAt sql.NullTime
		if err := deviceRows.Scan(
			&item.ExperimentKey,
			&item.VariantKey,
			&item.PageKey,
			&item.DeviceType,
			&item.ExposureCount,
			&item.ClickCount,
			&item.UpgradeIntentCount,
			&item.PaymentSuccessCount,
			&item.RenewalSuccessCount,
			&itemLastEventAt,
		); err != nil {
			return summary, err
		}
		paidSuccessTotal := item.PaymentSuccessCount + item.RenewalSuccessCount
		item.ClickThroughRate = safeRatioValue(int64(item.ClickCount), int64(item.ExposureCount))
		item.UpgradePerClickRate = safeRatioValue(int64(item.UpgradeIntentCount), int64(item.ClickCount))
		item.UpgradePerExposureRate = safeRatioValue(int64(item.UpgradeIntentCount), int64(item.ExposureCount))
		item.PaidPerUpgradeRate = safeRatioValue(int64(paidSuccessTotal), int64(item.UpgradeIntentCount))
		item.PaidPerClickRate = safeRatioValue(int64(paidSuccessTotal), int64(item.ClickCount))
		item.PaidPerExposureRate = safeRatioValue(int64(paidSuccessTotal), int64(item.ExposureCount))
		if itemLastEventAt.Valid {
			item.LastEventAt = itemLastEventAt.Time.Format(time.RFC3339)
		}
		summary.DeviceBreakdown = append(summary.DeviceBreakdown, item)
	}
	if err := deviceRows.Err(); err != nil {
		return summary, err
	}

	variantTrendRows, err := r.db.Query(`
SELECT
	DATE(created_at) AS metric_date,
	experiment_key,
	variant_key,
	page_key,
	COALESCE(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(metadata_json, '$.device_type')), ''), 'UNKNOWN') AS device_type,
	COALESCE(NULLIF(user_stage, ''), 'UNKNOWN') AS user_stage,
	COALESCE(SUM(CASE WHEN event_type = 'EXPOSURE' THEN 1 ELSE 0 END), 0) AS exposure_count,
	COALESCE(SUM(CASE WHEN event_type = 'CLICK' THEN 1 ELSE 0 END), 0) AS click_count,
	COALESCE(SUM(CASE WHEN event_type = 'UPGRADE_INTENT' THEN 1 ELSE 0 END), 0) AS upgrade_intent_count,
	COALESCE(SUM(CASE WHEN event_type = 'PAYMENT_SUCCESS' THEN 1 ELSE 0 END), 0) AS payment_success_count,
	COALESCE(SUM(CASE WHEN event_type = 'RENEWAL_SUCCESS' THEN 1 ELSE 0 END), 0) AS renewal_success_count
FROM experiment_events
WHERE created_at >= ?
GROUP BY DATE(created_at), experiment_key, variant_key, page_key,
	COALESCE(NULLIF(JSON_UNQUOTE(JSON_EXTRACT(metadata_json, '$.device_type')), ''), 'UNKNOWN'),
	COALESCE(NULLIF(user_stage, ''), 'UNKNOWN')
ORDER BY metric_date ASC, experiment_key ASC, variant_key ASC, page_key ASC
LIMIT 360`, since)
	if err != nil {
		return summary, err
	}
	defer variantTrendRows.Close()

	for variantTrendRows.Next() {
		var metricDate time.Time
		var point model.AdminExperimentAnalyticsVariantTrendPoint
		if err := variantTrendRows.Scan(
			&metricDate,
			&point.ExperimentKey,
			&point.VariantKey,
			&point.PageKey,
			&point.DeviceType,
			&point.UserStage,
			&point.ExposureCount,
			&point.ClickCount,
			&point.UpgradeIntentCount,
			&point.PaymentSuccessCount,
			&point.RenewalSuccessCount,
		); err != nil {
			return summary, err
		}
		paidSuccessTotal := point.PaymentSuccessCount + point.RenewalSuccessCount
		point.Date = metricDate.Format("2006-01-02")
		point.ClickThroughRate = safeRatioValue(int64(point.ClickCount), int64(point.ExposureCount))
		point.UpgradePerClickRate = safeRatioValue(int64(point.UpgradeIntentCount), int64(point.ClickCount))
		point.PaidPerExposureRate = safeRatioValue(int64(paidSuccessTotal), int64(point.ExposureCount))
		summary.VariantDailyTrend = append(summary.VariantDailyTrend, point)
	}
	if err := variantTrendRows.Err(); err != nil {
		return summary, err
	}

	return summary, nil
}

func (r *MySQLGrowthRepo) AdminListVIPQuotaConfigs(memberLevel string, status string, page int, pageSize int) ([]model.VIPQuotaConfig, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if memberLevel != "" {
		filter += " AND member_level = ?"
		args = append(args, memberLevel)
	}
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM vip_quota_configs"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, member_level, doc_read_limit, news_subscribe_limit, download_limit, forecast_limit, stock_reco_limit, reset_cycle, status, effective_at, updated_at
FROM vip_quota_configs` + filter + `
ORDER BY effective_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.VIPQuotaConfig, 0)
	for rows.Next() {
		var item model.VIPQuotaConfig
		var effectiveAt, updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.MemberLevel, &item.DocReadLimit, &item.NewsSubscribeLimit, &item.DownloadLimit, &item.ForecastLimit, &item.StockRecoLimit, &item.ResetCycle, &item.Status, &effectiveAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		item.EffectiveAt = effectiveAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateVIPQuotaConfig(item model.VIPQuotaConfig) (string, error) {
	id := newID("vqc")
	effectiveAt, err := time.Parse(time.RFC3339, item.EffectiveAt)
	if err != nil {
		return "", err
	}
	_, err = r.db.Exec(`
INSERT INTO vip_quota_configs (id, member_level, doc_read_limit, news_subscribe_limit, download_limit, forecast_limit, stock_reco_limit, reset_cycle, status, effective_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, item.MemberLevel, item.DocReadLimit, item.NewsSubscribeLimit, item.DownloadLimit, item.ForecastLimit, item.StockRecoLimit, item.ResetCycle, item.Status, effectiveAt, time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateVIPQuotaConfig(id string, item model.VIPQuotaConfig) error {
	effectiveAt, err := time.Parse(time.RFC3339, item.EffectiveAt)
	if err != nil {
		return err
	}
	result, err := r.db.Exec(`
UPDATE vip_quota_configs
SET doc_read_limit = ?, news_subscribe_limit = ?, download_limit = ?, forecast_limit = ?, stock_reco_limit = ?, reset_cycle = ?, status = ?, effective_at = ?, updated_at = ?
WHERE id = ?`,
		item.DocReadLimit, item.NewsSubscribeLimit, item.DownloadLimit, item.ForecastLimit, item.StockRecoLimit, item.ResetCycle, item.Status, effectiveAt, time.Now(), id,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminDeleteVIPQuotaConfig(id string) error {
	result, err := r.db.Exec("DELETE FROM vip_quota_configs WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminListUserQuotaUsages(userID string, periodKey string, page int, pageSize int) ([]model.UserQuotaUsage, int, error) {
	if strings.TrimSpace(periodKey) == "" {
		periodKey = time.Now().Format("2006-01")
	}
	offset := (page - 1) * pageSize
	filter := " WHERE 1=1"
	countArgs := []interface{}{}
	dataArgs := []interface{}{periodKey, periodKey}
	if strings.TrimSpace(userID) != "" {
		filter += " AND u.id = ?"
		countArgs = append(countArgs, userID)
		dataArgs = append(dataArgs, userID)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM users u"+filter, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT COALESCE(uqu.id, ''), u.id, u.member_level, ?, 
       COALESCE(vqc.doc_read_limit, 0), COALESCE(uqu.doc_read_used, 0),
       COALESCE(vqc.news_subscribe_limit, 0), COALESCE(uqu.news_subscribe_used, 0),
       COALESCE(vqc.download_limit, 0), COALESCE(uqu.download_used, 0),
       COALESCE(vqc.forecast_limit, 0), COALESCE(uqu.forecast_used, 0),
       COALESCE(vqc.stock_reco_limit, 0), COALESCE(uqu.stock_reco_used, 0),
       uqu.updated_at
FROM users u
LEFT JOIN user_quota_usages uqu ON uqu.user_id = u.id AND uqu.period_key = ?
LEFT JOIN vip_quota_configs vqc ON vqc.id = (
    SELECT v2.id
    FROM vip_quota_configs v2
    WHERE v2.member_level = u.member_level AND v2.status = 'ACTIVE'
    ORDER BY v2.effective_at DESC
    LIMIT 1
)` + filter + `
ORDER BY u.created_at DESC
LIMIT ? OFFSET ?`
	dataArgs = append(dataArgs, pageSize, offset)
	rows, err := r.db.Query(query, dataArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.UserQuotaUsage, 0)
	for rows.Next() {
		var item model.UserQuotaUsage
		var updatedAt sql.NullTime
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.MemberLevel,
			&item.PeriodKey,
			&item.DocReadLimit,
			&item.DocReadUsed,
			&item.NewsSubscribeLimit,
			&item.NewsSubscribeUsed,
			&item.DownloadLimit,
			&item.DownloadUsed,
			&item.ForecastLimit,
			&item.ForecastUsed,
			&item.StockRecoLimit,
			&item.StockRecoUsed,
			&updatedAt,
		); err != nil {
			return nil, 0, err
		}
		if updatedAt.Valid {
			item.UpdatedAt = updatedAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminAdjustUserQuota(userID string, periodKey string, docReadDelta int, newsSubscribeDelta int, downloadDelta int, forecastDelta int, stockRecoDelta int) error {
	if strings.TrimSpace(periodKey) == "" {
		return errors.New("period_key is required")
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	var memberLevel string
	if err := tx.QueryRow("SELECT member_level FROM users WHERE id = ?", userID).Scan(&memberLevel); err != nil {
		_ = tx.Rollback()
		return err
	}

	now := time.Now()
	_, err = tx.Exec(`
INSERT INTO user_quota_usages (id, user_id, member_level, period_key, doc_read_used, news_subscribe_used, download_used, forecast_used, stock_reco_used, updated_at)
VALUES (?, ?, ?, ?, 0, 0, 0, 0, 0, ?)
ON DUPLICATE KEY UPDATE member_level = VALUES(member_level), updated_at = VALUES(updated_at)`,
		newID("uqu"), userID, memberLevel, periodKey, now,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	_, err = tx.Exec(`
UPDATE user_quota_usages
SET doc_read_used = GREATEST(0, doc_read_used + ?),
	news_subscribe_used = GREATEST(0, news_subscribe_used + ?),
	download_used = GREATEST(0, download_used + ?),
	forecast_used = GREATEST(0, forecast_used + ?),
	stock_reco_used = GREATEST(0, stock_reco_used + ?),
	updated_at = ?
WHERE user_id = ? AND period_key = ?`,
		docReadDelta, newsSubscribeDelta, downloadDelta, forecastDelta, stockRecoDelta, now, userID, periodKey,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *MySQLGrowthRepo) AdminDeleteUserQuota(id string) error {
	result, err := r.db.Exec("DELETE FROM user_quota_usages WHERE id = ?", id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}
