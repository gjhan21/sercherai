package repo

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) AdminListBrowseHistories(userID string, contentType string, keyword string, page int, pageSize int) ([]model.AdminBrowseHistory, int, error) {
	offset := (page - 1) * pageSize
	filter := " WHERE 1=1"
	args := []interface{}{}

	userID = strings.TrimSpace(userID)
	if userID != "" {
		filter += " AND bh.user_id = ?"
		args = append(args, userID)
	}

	contentType = strings.ToUpper(strings.TrimSpace(contentType))
	if contentType != "" {
		filter += " AND bh.content_type = ?"
		args = append(args, contentType)
	}

	keyword = strings.TrimSpace(keyword)
	if keyword != "" {
		kw := "%" + keyword + "%"
		filter += " AND (bh.content_id LIKE ? OR na.title LIKE ? OR u.phone LIKE ?)"
		args = append(args, kw, kw, kw)
	}

	countQuery := `
SELECT COUNT(*)
FROM browse_histories bh
LEFT JOIN users u ON u.id = bh.user_id
LEFT JOIN news_articles na ON bh.content_type = 'NEWS' AND na.id = bh.content_id` + filter
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT
	bh.id,
	bh.user_id,
	u.phone,
	bh.content_type,
	bh.content_id,
	COALESCE(NULLIF(na.title, ''), bh.content_id) AS title,
	bh.source_page,
	bh.viewed_at
FROM browse_histories bh
LEFT JOIN users u ON u.id = bh.user_id
LEFT JOIN news_articles na ON bh.content_type = 'NEWS' AND na.id = bh.content_id` + filter + `
ORDER BY bh.viewed_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.AdminBrowseHistory, 0)
	for rows.Next() {
		var item model.AdminBrowseHistory
		var userPhone sql.NullString
		var sourcePage sql.NullString
		var viewedAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&userPhone,
			&item.ContentType,
			&item.ContentID,
			&item.Title,
			&sourcePage,
			&viewedAt,
		); err != nil {
			return nil, 0, err
		}
		if userPhone.Valid {
			item.UserPhone = userPhone.String
		}
		if sourcePage.Valid {
			item.SourcePage = sourcePage.String
		}
		item.ViewedAt = viewedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminGetBrowseHistorySummary() (model.AdminBrowseHistorySummary, error) {
	result := model.AdminBrowseHistorySummary{}
	err := r.db.QueryRow(`
SELECT
	COUNT(*) AS total_views,
	COUNT(DISTINCT user_id) AS unique_users,
	COALESCE(SUM(CASE WHEN content_type = 'NEWS' THEN 1 ELSE 0 END), 0) AS news_views,
	COALESCE(SUM(CASE WHEN content_type = 'REPORT' THEN 1 ELSE 0 END), 0) AS report_views,
	COALESCE(SUM(CASE WHEN content_type = 'JOURNAL' THEN 1 ELSE 0 END), 0) AS journal_views,
	COALESCE(SUM(CASE WHEN DATE(viewed_at) = CURRENT_DATE() THEN 1 ELSE 0 END), 0) AS today_views,
	COALESCE(SUM(CASE WHEN viewed_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) THEN 1 ELSE 0 END), 0) AS last_7d_views
FROM browse_histories`).Scan(
		&result.TotalViews,
		&result.UniqueUsers,
		&result.NewsViews,
		&result.ReportViews,
		&result.JournalViews,
		&result.TodayViews,
		&result.Last7dViews,
	)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (r *MySQLGrowthRepo) AdminGetBrowseHistoryTrend(days int) ([]model.AdminBrowseTrendPoint, error) {
	if days <= 0 {
		days = 7
	}
	if days > 30 {
		days = 30
	}

	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	rangeStart := dayStart.AddDate(0, 0, -(days - 1))

	rows, err := r.db.Query(`
SELECT
	DATE(viewed_at) AS dt,
	COUNT(*) AS total_views,
	COALESCE(SUM(CASE WHEN content_type = 'NEWS' THEN 1 ELSE 0 END), 0) AS news_views,
	COALESCE(SUM(CASE WHEN content_type = 'REPORT' THEN 1 ELSE 0 END), 0) AS report_views,
	COALESCE(SUM(CASE WHEN content_type = 'JOURNAL' THEN 1 ELSE 0 END), 0) AS journal_views
FROM browse_histories
WHERE viewed_at >= ?
GROUP BY DATE(viewed_at)
ORDER BY dt ASC`, rangeStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	indexed := make(map[string]model.AdminBrowseTrendPoint, days)
	for rows.Next() {
		var day time.Time
		var item model.AdminBrowseTrendPoint
		if err := rows.Scan(&day, &item.TotalViews, &item.NewsViews, &item.ReportViews, &item.JournalViews); err != nil {
			return nil, err
		}
		item.Date = day.Format("2006-01-02")
		indexed[item.Date] = item
	}

	points := make([]model.AdminBrowseTrendPoint, 0, days)
	for i := 0; i < days; i++ {
		current := rangeStart.AddDate(0, 0, i).Format("2006-01-02")
		if item, exists := indexed[current]; exists {
			points = append(points, item)
		} else {
			points = append(points, model.AdminBrowseTrendPoint{Date: current})
		}
	}
	return points, nil
}

func (r *MySQLGrowthRepo) AdminListBrowseUserSegments(limit int) ([]model.AdminBrowseUserSegment, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	activeRows, err := r.db.Query(`
SELECT
	u.id,
	u.phone,
	COUNT(bh.id) AS view_count_7d,
	MAX(bh.viewed_at) AS last_viewed_at
FROM users u
JOIN browse_histories bh ON bh.user_id = u.id AND bh.viewed_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
WHERE u.status = 'ACTIVE'
GROUP BY u.id, u.phone
ORDER BY view_count_7d DESC, last_viewed_at DESC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer activeRows.Close()

	result := make([]model.AdminBrowseUserSegment, 0, limit*2)
	for activeRows.Next() {
		var item model.AdminBrowseUserSegment
		var phone sql.NullString
		var lastViewedAt sql.NullTime
		if err := activeRows.Scan(&item.UserID, &phone, &item.ViewCount7d, &lastViewedAt); err != nil {
			return nil, err
		}
		item.Segment = "ACTIVE"
		if phone.Valid {
			item.UserPhone = phone.String
		}
		if lastViewedAt.Valid {
			item.LastViewedAt = lastViewedAt.Time.Format(time.RFC3339)
		}
		_ = r.db.QueryRow(`
SELECT content_id, content_type
FROM browse_histories
WHERE user_id = ?
ORDER BY viewed_at DESC
LIMIT 1`, item.UserID).Scan(&item.LastContentID, &item.LastContentType)
		result = append(result, item)
	}

	silentRows, err := r.db.Query(`
SELECT
	u.id,
	u.phone,
	MAX(bh.viewed_at) AS last_viewed_at
FROM users u
LEFT JOIN browse_histories bh ON bh.user_id = u.id
WHERE u.status = 'ACTIVE'
GROUP BY u.id, u.phone
HAVING MAX(bh.viewed_at) IS NULL OR MAX(bh.viewed_at) < DATE_SUB(NOW(), INTERVAL 7 DAY)
ORDER BY last_viewed_at ASC
LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer silentRows.Close()

	for silentRows.Next() {
		var item model.AdminBrowseUserSegment
		var phone sql.NullString
		var lastViewedAt sql.NullTime
		if err := silentRows.Scan(&item.UserID, &phone, &lastViewedAt); err != nil {
			return nil, err
		}
		item.Segment = "SILENT"
		item.ViewCount7d = 0
		if phone.Valid {
			item.UserPhone = phone.String
		}
		if lastViewedAt.Valid {
			item.LastViewedAt = lastViewedAt.Time.Format(time.RFC3339)
		}
		_ = r.db.QueryRow(`
SELECT content_id, content_type
FROM browse_histories
WHERE user_id = ?
ORDER BY viewed_at DESC
LIMIT 1`, item.UserID).Scan(&item.LastContentID, &item.LastContentType)
		result = append(result, item)
	}
	return result, nil
}

func (r *MySQLGrowthRepo) AdminListUserMessages(userID string, messageType string, readStatus string, page int, pageSize int) ([]model.AdminUserMessage, int, error) {
	offset := (page - 1) * pageSize
	filter := " WHERE 1=1"
	args := []interface{}{}

	userID = strings.TrimSpace(userID)
	if userID != "" {
		filter += " AND m.user_id = ?"
		args = append(args, userID)
	}

	messageType = strings.ToUpper(strings.TrimSpace(messageType))
	if messageType != "" {
		filter += " AND m.type = ?"
		args = append(args, messageType)
	}

	readStatus = strings.ToUpper(strings.TrimSpace(readStatus))
	if readStatus != "" {
		filter += " AND m.read_status = ?"
		args = append(args, readStatus)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM messages m"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT m.id, m.user_id, u.phone, m.title, m.content, m.type, m.read_status, m.created_at
FROM messages m
LEFT JOIN users u ON u.id = m.user_id` + filter + `
ORDER BY m.created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.AdminUserMessage, 0)
	for rows.Next() {
		var item model.AdminUserMessage
		var userPhone sql.NullString
		var content sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.UserID, &userPhone, &item.Title, &content, &item.Type, &item.ReadStatus, &createdAt); err != nil {
			return nil, 0, err
		}
		if userPhone.Valid {
			item.UserPhone = userPhone.String
		}
		if content.Valid {
			item.Content = content.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateUserMessages(userIDs []string, title string, content string, messageType string) (int, []model.AdminMessageSendFailure, error) {
	if len(userIDs) == 0 {
		return 0, nil, errors.New("no target users")
	}

	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)
	messageType = strings.ToUpper(strings.TrimSpace(messageType))

	stmt, err := r.db.Prepare(`
INSERT INTO messages (id, user_id, title, content, type, read_status, created_at)
VALUES (?, ?, ?, ?, ?, 'UNREAD', ?)`)
	if err != nil {
		return 0, nil, err
	}
	defer stmt.Close()

	now := time.Now()
	sent := 0
	failures := make([]model.AdminMessageSendFailure, 0)
	existsCache := make(map[string]bool)
	for _, userID := range userIDs {
		userID = strings.TrimSpace(userID)
		if userID == "" {
			continue
		}

		exists, cached := existsCache[userID]
		if !cached {
			var count int
			if err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", userID).Scan(&count); err != nil {
				failures = append(failures, model.AdminMessageSendFailure{UserID: userID, Reason: err.Error()})
				continue
			}
			exists = count > 0
			existsCache[userID] = exists
		}
		if !exists {
			failures = append(failures, model.AdminMessageSendFailure{UserID: userID, Reason: "user not found"})
			continue
		}

		if _, err := stmt.Exec(newID("msg"), userID, title, content, messageType, now); err != nil {
			failures = append(failures, model.AdminMessageSendFailure{UserID: userID, Reason: err.Error()})
			continue
		}
		sent++
	}

	if sent == 0 {
		if len(failures) == 0 {
			return 0, nil, errors.New("no target users")
		}
		return 0, failures, nil
	}
	return sent, failures, nil
}

func (r *MySQLGrowthRepo) AdminListUsers(status string, memberLevel string, registrationSource string, phone string, email string, page int, pageSize int) ([]model.AdminUser, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	status = strings.ToUpper(strings.TrimSpace(status))
	memberLevel = strings.TrimSpace(memberLevel)
	registrationSource = strings.ToUpper(strings.TrimSpace(registrationSource))
	phone = strings.TrimSpace(phone)
	email = strings.TrimSpace(email)
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND u.status = ?"
		args = append(args, status)
	}
	if memberLevel != "" {
		filter += " AND u.member_level = ?"
		args = append(args, memberLevel)
	}
	if registrationSource == "INVITED" {
		filter += " AND ir.id IS NOT NULL"
	}
	if registrationSource == "DIRECT" {
		filter += " AND ir.id IS NULL"
	}
	if phone != "" {
		filter += " AND u.phone LIKE ?"
		args = append(args, "%"+phone+"%")
	}
	if email != "" {
		filter += " AND u.email LIKE ?"
		args = append(args, "%"+email+"%")
	}
	var total int
	countQuery := "SELECT COUNT(*) FROM users u LEFT JOIN invite_records ir ON ir.invitee_user_id = u.id" + filter
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT
	u.id,
	u.phone,
	u.email,
	u.status,
	u.kyc_status,
	u.member_level,
	u.vip_expire_at,
	CASE WHEN ir.id IS NULL THEN 'DIRECT' ELSE 'INVITED' END AS registration_source,
	ir.inviter_user_id,
	il.invite_code,
	ir.register_at,
	u.created_at
FROM users u
LEFT JOIN invite_records ir ON ir.invitee_user_id = u.id
LEFT JOIN invite_links il ON il.id = ir.invite_link_id` + filter + `
ORDER BY u.created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.AdminUser, 0)
	for rows.Next() {
		var item model.AdminUser
		var email sql.NullString
		var inviterUserID sql.NullString
		var inviteCode sql.NullString
		var vipExpireAt sql.NullTime
		var invitedAt sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Phone,
			&email,
			&item.Status,
			&item.KYCStatus,
			&item.MemberLevel,
			&vipExpireAt,
			&item.RegistrationSource,
			&inviterUserID,
			&inviteCode,
			&invitedAt,
			&createdAt,
		); err != nil {
			return nil, 0, err
		}
		if email.Valid {
			item.Email = email.String
		}
		if inviterUserID.Valid {
			item.InviterUserID = inviterUserID.String
		}
		if inviteCode.Valid {
			item.InviteCode = inviteCode.String
		}
		if invitedAt.Valid {
			item.InviteRegisteredAt = invitedAt.Time.Format(time.RFC3339)
		}
		item.ActivationState = resolveMembershipActivationState(item.MemberLevel, vipExpireAt)
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminGetUserSourceSummary(status string, memberLevel string, registrationSource string, phone string, email string) (model.AdminUserSourceSummary, error) {
	result := model.AdminUserSourceSummary{}
	args := []interface{}{}
	status = strings.ToUpper(strings.TrimSpace(status))
	memberLevel = strings.TrimSpace(memberLevel)
	registrationSource = strings.ToUpper(strings.TrimSpace(registrationSource))
	phone = strings.TrimSpace(phone)
	email = strings.TrimSpace(email)

	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND u.status = ?"
		args = append(args, status)
	}
	if memberLevel != "" {
		filter += " AND u.member_level = ?"
		args = append(args, memberLevel)
	}
	if registrationSource == "INVITED" {
		filter += " AND ir.id IS NOT NULL"
	}
	if registrationSource == "DIRECT" {
		filter += " AND ir.id IS NULL"
	}
	if phone != "" {
		filter += " AND u.phone LIKE ?"
		args = append(args, "%"+phone+"%")
	}
	if email != "" {
		filter += " AND u.email LIKE ?"
		args = append(args, "%"+email+"%")
	}

	query := `
SELECT
	COUNT(*) AS total_users,
	SUM(CASE WHEN ir.id IS NULL THEN 1 ELSE 0 END) AS direct_users,
	SUM(CASE WHEN ir.id IS NOT NULL THEN 1 ELSE 0 END) AS invited_users,
	SUM(CASE WHEN ir.id IS NOT NULL AND DATE(ir.register_at) = CURDATE() THEN 1 ELSE 0 END) AS today_invited_users,
	SUM(CASE WHEN ir.id IS NOT NULL AND ir.first_pay_at IS NOT NULL THEN 1 ELSE 0 END) AS total_first_paid_users,
	SUM(CASE WHEN ir.id IS NOT NULL AND ir.register_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) THEN 1 ELSE 0 END) AS last_7d_invited_users,
	SUM(CASE WHEN ir.id IS NOT NULL AND ir.register_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) AND ir.first_pay_at IS NOT NULL THEN 1 ELSE 0 END) AS last_7d_first_paid_users,
	SUM(CASE WHEN ir.id IS NOT NULL AND ir.register_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 1 ELSE 0 END) AS last_30d_invited_users,
	SUM(CASE WHEN ir.id IS NOT NULL AND ir.register_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) AND ir.first_pay_at IS NOT NULL THEN 1 ELSE 0 END) AS last_30d_first_paid_users
FROM users u
LEFT JOIN invite_records ir ON ir.invitee_user_id = u.id` + filter

	var directUsers sql.NullInt64
	var invitedUsers sql.NullInt64
	var todayInvitedUsers sql.NullInt64
	var totalFirstPaidUsers sql.NullInt64
	var last7dInvitedUsers sql.NullInt64
	var last7dFirstPaidUsers sql.NullInt64
	var last30dInvitedUsers sql.NullInt64
	var last30dFirstPaidUsers sql.NullInt64
	if err := r.db.QueryRow(query, args...).Scan(
		&result.TotalUsers,
		&directUsers,
		&invitedUsers,
		&todayInvitedUsers,
		&totalFirstPaidUsers,
		&last7dInvitedUsers,
		&last7dFirstPaidUsers,
		&last30dInvitedUsers,
		&last30dFirstPaidUsers,
	); err != nil {
		return result, err
	}
	if directUsers.Valid {
		result.DirectUsers = int(directUsers.Int64)
	}
	if invitedUsers.Valid {
		result.InvitedUsers = int(invitedUsers.Int64)
	}
	if todayInvitedUsers.Valid {
		result.TodayInvitedUsers = int(todayInvitedUsers.Int64)
	}
	if totalFirstPaidUsers.Valid {
		result.TotalFirstPaidUsers = int(totalFirstPaidUsers.Int64)
	}
	if last7dInvitedUsers.Valid {
		result.Last7dInvitedUsers = int(last7dInvitedUsers.Int64)
	}
	if last7dFirstPaidUsers.Valid {
		result.Last7dFirstPaidUsers = int(last7dFirstPaidUsers.Int64)
	}
	if last30dInvitedUsers.Valid {
		result.Last30dInvitedUsers = int(last30dInvitedUsers.Int64)
	}
	if last30dFirstPaidUsers.Valid {
		result.Last30dFirstPaidUsers = int(last30dFirstPaidUsers.Int64)
	}
	if result.TotalUsers > 0 {
		result.InviteRate = float64(result.InvitedUsers) / float64(result.TotalUsers)
	}
	if result.InvitedUsers > 0 {
		result.TotalConversionRate = float64(result.TotalFirstPaidUsers) / float64(result.InvitedUsers)
	}
	if result.Last7dInvitedUsers > 0 {
		result.Last7dConversionRate = float64(result.Last7dFirstPaidUsers) / float64(result.Last7dInvitedUsers)
	}
	if result.Last30dInvitedUsers > 0 {
		result.Last30dConversionRate = float64(result.Last30dFirstPaidUsers) / float64(result.Last30dInvitedUsers)
	}
	return result, nil
}

func (r *MySQLGrowthRepo) AdminUpdateUserStatus(id string, status string) error {
	_, err := r.db.Exec("UPDATE users SET status = ?, updated_at = ? WHERE id = ?", status, time.Now(), id)
	return err
}

func (r *MySQLGrowthRepo) AdminUpdateUserMemberLevel(id string, memberLevel string) error {
	_, err := r.db.Exec("UPDATE users SET member_level = ?, updated_at = ? WHERE id = ?", memberLevel, time.Now(), id)
	return err
}

func (r *MySQLGrowthRepo) AdminResetUserPasswordHash(id string, passwordHash string) error {
	result, err := r.db.Exec("UPDATE users SET password_hash = ?, updated_at = ? WHERE id = ?", passwordHash, time.Now(), id)
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

func (r *MySQLGrowthRepo) AdminDashboardOverview() (model.AdminDashboardOverview, error) {
	result := model.AdminDashboardOverview{}
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&result.TotalUsers); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE status = 'ACTIVE'").Scan(&result.ActiveUsers); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE member_level LIKE 'VIP%'").Scan(&result.VIPUsers); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE status = 'ACTIVE'").Scan(&result.ActiveSubscriptions); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM membership_orders WHERE status = 'PENDING'").Scan(&result.PendingMembershipOrders); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= ? AND created_at < ?", startOfDay, endOfDay).Scan(&result.TodayNewUsers); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM membership_orders WHERE status = 'PAID' AND paid_at >= ? AND paid_at < ?", startOfDay, endOfDay).Scan(&result.TodayPaidOrders); err != nil {
		return result, err
	}
	if err := r.db.QueryRow(`
SELECT COALESCE(SUM(amount), 0)
FROM membership_orders
WHERE status = 'PAID' AND paid_at >= ? AND paid_at < ?`, startOfDay, endOfDay).Scan(&result.TodayPaidAmount); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM stock_recommendations WHERE status = 'PUBLISHED' AND created_at >= ? AND created_at < ?", startOfDay, endOfDay).Scan(&result.TodayPublishedStocks); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM news_articles WHERE status = 'PUBLISHED' AND created_at >= ? AND created_at < ?", startOfDay, endOfDay).Scan(&result.TodayPublishedNews); err != nil {
		return result, err
	}
	return result, nil
}

func (r *MySQLGrowthRepo) AdminCreateOperationLog(module string, action string, targetType string, targetID string, operatorUserID string, beforeValue string, afterValue string, reason string) error {
	_, err := r.db.Exec(`
INSERT INTO admin_operation_logs (id, module, action, target_type, target_id, operator_user_id, before_value, after_value, reason, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("aol"), strings.ToUpper(module), strings.ToUpper(action), strings.ToUpper(targetType), targetID, operatorUserID, beforeValue, afterValue, reason, time.Now(),
	)
	return err
}

func (r *MySQLGrowthRepo) AdminListOperationLogs(module string, action string, operatorUserID string, page int, pageSize int) ([]model.AdminOperationLog, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if action != "" {
		filter += " AND action = ?"
		args = append(args, strings.ToUpper(action))
	}
	if operatorUserID != "" {
		filter += " AND operator_user_id = ?"
		args = append(args, operatorUserID)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM admin_operation_logs"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, module, action, target_type, target_id, operator_user_id, before_value, after_value, reason, created_at
FROM admin_operation_logs` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.AdminOperationLog, 0)
	for rows.Next() {
		var item model.AdminOperationLog
		var beforeVal, afterVal, reason sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Module, &item.Action, &item.TargetType, &item.TargetID, &item.OperatorUserID, &beforeVal, &afterVal, &reason, &createdAt); err != nil {
			return nil, 0, err
		}
		if beforeVal.Valid {
			item.BeforeValue = beforeVal.String
		}
		if afterVal.Valid {
			item.AfterValue = afterVal.String
		}
		if reason.Valid {
			item.Reason = reason.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}
