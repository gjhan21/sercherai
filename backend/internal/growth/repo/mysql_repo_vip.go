package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type vipLifecycleRuntimeConfig struct {
	Enabled         bool
	RemindDaysThree int
	RemindDaysOne   int
}

func (r *MySQLGrowthRepo) AdminRunVIPMembershipLifecycle() (string, error) {
	cfg := r.resolveVIPLifecycleRuntimeConfig()
	if !cfg.Enabled {
		return "vip lifecycle disabled", nil
	}

	now := time.Now()
	tx, err := r.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	remindThreeCount, err := r.sendVIPExpiryReminder(tx, now, cfg.RemindDaysThree, "vip_remind_3d_at")
	if err != nil {
		return "", err
	}
	remindOneCount, err := r.sendVIPExpiryReminder(tx, now, cfg.RemindDaysOne, "vip_remind_1d_at")
	if err != nil {
		return "", err
	}
	expiredCount, err := r.expireVIPUsers(tx, now)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"vip_remind_3d=%d vip_remind_1d=%d vip_expired=%d",
		remindThreeCount,
		remindOneCount,
		expiredCount,
	), nil
}

func (r *MySQLGrowthRepo) resolveVIPLifecycleRuntimeConfig() vipLifecycleRuntimeConfig {
	cfg := vipLifecycleRuntimeConfig{
		Enabled:         true,
		RemindDaysThree: 3,
		RemindDaysOne:   1,
	}
	items, _, err := r.AdminListSystemConfigs("membership.vip.lifecycle.", 1, 50)
	if err != nil {
		return cfg
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case "membership.vip.lifecycle.enabled":
			cfg.Enabled = parseRepoConfigBool(value, cfg.Enabled)
		case "membership.vip.lifecycle.remind_days_3":
			cfg.RemindDaysThree = parseRepoConfigInt(value, cfg.RemindDaysThree)
		case "membership.vip.lifecycle.remind_days_1":
			cfg.RemindDaysOne = parseRepoConfigInt(value, cfg.RemindDaysOne)
		}
	}
	if cfg.RemindDaysThree < 1 {
		cfg.RemindDaysThree = 3
	}
	if cfg.RemindDaysOne < 1 {
		cfg.RemindDaysOne = 1
	}
	if cfg.RemindDaysThree < cfg.RemindDaysOne {
		cfg.RemindDaysThree = cfg.RemindDaysOne + 1
	}
	return cfg
}

func (r *MySQLGrowthRepo) sendVIPExpiryReminder(tx *sql.Tx, now time.Time, remindDays int, remindColumn string) (int, error) {
	if remindDays <= 0 {
		return 0, nil
	}
	if remindColumn != "vip_remind_3d_at" && remindColumn != "vip_remind_1d_at" {
		return 0, nil
	}
	endAt := now.AddDate(0, 0, remindDays)
	query := fmt.Sprintf(`
SELECT id, member_level, vip_expire_at
FROM users
WHERE member_level LIKE 'VIP%%'
  AND vip_expire_at IS NOT NULL
  AND vip_expire_at > ?
  AND vip_expire_at <= ?
  AND %s IS NULL
ORDER BY vip_expire_at ASC
LIMIT 500`, remindColumn)

	rows, err := tx.Query(query, now, endAt)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type reminderCandidate struct {
		UserID      string
		MemberLevel string
		ExpireAt    time.Time
	}
	candidates := make([]reminderCandidate, 0)
	for rows.Next() {
		var item reminderCandidate
		if err := rows.Scan(&item.UserID, &item.MemberLevel, &item.ExpireAt); err != nil {
			return 0, err
		}
		item.UserID = strings.TrimSpace(item.UserID)
		item.MemberLevel = strings.ToUpper(strings.TrimSpace(item.MemberLevel))
		if item.UserID == "" || !strings.HasPrefix(item.MemberLevel, "VIP") {
			continue
		}
		candidates = append(candidates, item)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	count := 0
	for _, item := range candidates {
		res, err := tx.Exec(
			fmt.Sprintf("UPDATE users SET %s = ?, updated_at = ? WHERE id = ? AND %s IS NULL", remindColumn, remindColumn),
			now,
			now,
			item.UserID,
		)
		if err != nil {
			return count, err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			continue
		}
		daysLeft := vipRemainingDays(sql.NullTime{Time: item.ExpireAt, Valid: true}, now)
		if daysLeft < 1 {
			daysLeft = 1
		}
		title := "VIP到期提醒"
		content := fmt.Sprintf(
			"您的%s将于%s到期（约剩余%d天），到期后VIP权限将自动暂停。",
			item.MemberLevel,
			item.ExpireAt.Format("2006-01-02 15:04:05"),
			daysLeft,
		)
		if err := insertSystemMessageTx(tx, item.UserID, title, content, now); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func (r *MySQLGrowthRepo) expireVIPUsers(tx *sql.Tx, now time.Time) (int, error) {
	rows, err := tx.Query(`
SELECT id, member_level, vip_expire_at
FROM users
WHERE member_level LIKE 'VIP%'
  AND vip_expire_at IS NOT NULL
  AND vip_expire_at <= ?
ORDER BY vip_expire_at ASC
LIMIT 2000`, now)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	type expiredCandidate struct {
		UserID      string
		MemberLevel string
		ExpireAt    time.Time
	}
	candidates := make([]expiredCandidate, 0)
	for rows.Next() {
		var item expiredCandidate
		if err := rows.Scan(&item.UserID, &item.MemberLevel, &item.ExpireAt); err != nil {
			return 0, err
		}
		item.UserID = strings.TrimSpace(item.UserID)
		item.MemberLevel = strings.ToUpper(strings.TrimSpace(item.MemberLevel))
		if item.UserID == "" || !strings.HasPrefix(item.MemberLevel, "VIP") {
			continue
		}
		candidates = append(candidates, item)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	count := 0
	for _, item := range candidates {
		res, err := tx.Exec(`
UPDATE users
SET member_level = 'FREE',
    vip_started_at = NULL,
    vip_expire_at = NULL,
    vip_remind_3d_at = NULL,
    vip_remind_1d_at = NULL,
    updated_at = ?
WHERE id = ? AND member_level LIKE 'VIP%'`,
			now,
			item.UserID,
		)
		if err != nil {
			return count, err
		}
		affected, _ := res.RowsAffected()
		if affected == 0 {
			continue
		}
		title := "VIP已到期暂停"
		content := fmt.Sprintf(
			"您的%s已于%s到期，当前已自动切换为免费权限。",
			item.MemberLevel,
			item.ExpireAt.Format("2006-01-02 15:04:05"),
		)
		if err := insertSystemMessageTx(tx, item.UserID, title, content, now); err != nil {
			return count, err
		}
		count++
	}
	return count, nil
}

func insertSystemMessageTx(tx *sql.Tx, userID string, title string, content string, now time.Time) error {
	_, err := tx.Exec(`
INSERT INTO messages (id, user_id, title, content, type, read_status, created_at)
VALUES (?, ?, ?, ?, 'SYSTEM', 'UNREAD', ?)`,
		newID("msg"),
		userID,
		truncateByRunes(strings.TrimSpace(title), 128),
		truncateByRunes(strings.TrimSpace(content), 2048),
		now,
	)
	return err
}

func (r *MySQLGrowthRepo) isVIPUser(userID string) (bool, error) {
	var memberLevel string
	var kycStatus string
	var vipExpireAt sql.NullTime
	err := r.db.QueryRow("SELECT member_level, kyc_status, vip_expire_at FROM users WHERE id = ?", userID).Scan(&memberLevel, &kycStatus, &vipExpireAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	memberLevel = strings.ToUpper(strings.TrimSpace(memberLevel))
	if !strings.HasPrefix(memberLevel, "VIP") {
		return false, nil
	}
	now := time.Now()
	if vipExpireAt.Valid && !vipExpireAt.Time.After(now) {
		if _, err := r.db.Exec(`
UPDATE users
SET member_level = 'FREE',
    vip_started_at = NULL,
    vip_expire_at = NULL,
    vip_remind_3d_at = NULL,
    vip_remind_1d_at = NULL,
    updated_at = ?
WHERE id = ? AND member_level LIKE 'VIP%'`, now, userID); err != nil {
			return false, err
		}
		return false, nil
	}
	return resolveMembershipActivationState(memberLevel, vipExpireAt) == "ACTIVE", nil
}

func (r *MySQLGrowthRepo) CheckAndConsumeQuota(userID string, quotaType string, contentID string) (bool, error) {
	var memberLevel string
	var vipExpireAt sql.NullTime
	if err := r.db.QueryRow("SELECT member_level, vip_expire_at FROM users WHERE id = ?", userID).Scan(&memberLevel, &vipExpireAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			memberLevel = "FREE"
		} else {
			return false, err
		}
	}
	memberLevel = strings.ToUpper(strings.TrimSpace(memberLevel))

	if strings.HasPrefix(memberLevel, "VIP") && vipExpireAt.Valid && !vipExpireAt.Time.After(time.Now()) {
		now := time.Now()
		if _, err := r.db.Exec(`
UPDATE users
SET member_level = 'FREE',
    vip_started_at = NULL,
    vip_expire_at = NULL,
    vip_remind_3d_at = NULL,
    vip_remind_1d_at = NULL,
    updated_at = ?
WHERE id = ? AND member_level LIKE 'VIP%'`, now, userID); err != nil {
			return false, err
		}
		memberLevel = "FREE"
	}

	now := time.Now()
	periodKey := now.Format("2006-01")
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	if contentID != "" {
		var count int
		var checkErr error
		if quotaType == "news_subscribe" || quotaType == "doc_read" {
			checkErr = r.db.QueryRow("SELECT COUNT(*) FROM browse_histories WHERE user_id = ? AND content_type = 'NEWS' AND content_id = ? AND viewed_at >= ?", userID, contentID, periodStart).Scan(&count)
		} else if quotaType == "download" {
			checkErr = r.db.QueryRow("SELECT COUNT(*) FROM attachment_download_logs WHERE user_id = ? AND attachment_id = ? AND downloaded_at >= ?", userID, contentID, periodStart).Scan(&count)
		} else if quotaType == "stock_reco" {
			checkErr = r.db.QueryRow("SELECT COUNT(*) FROM browse_histories WHERE user_id = ? AND content_type = 'STOCK' AND content_id = ? AND viewed_at >= ?", userID, contentID, periodStart).Scan(&count)
		}
		if checkErr == nil && count > 0 {
			return true, nil
		}
	}

	var limit int
	limitColumn := ""
	switch quotaType {
	case "doc_read":
		limitColumn = "doc_read_limit"
	case "news_subscribe":
		limitColumn = "news_subscribe_limit"
	case "download":
		limitColumn = "download_limit"
	case "forecast":
		limitColumn = "forecast_limit"
	case "stock_reco":
		limitColumn = "stock_reco_limit"
	default:
		return false, fmt.Errorf("invalid quota type: %s", quotaType)
	}

	err := r.db.QueryRow(fmt.Sprintf(`
SELECT %s
FROM vip_quota_configs
WHERE member_level = ? AND status = 'ACTIVE'
ORDER BY effective_at DESC
LIMIT 1`, limitColumn), memberLevel).Scan(&limit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			limit = 0
		} else {
			return false, err
		}
	}

	var used int
	usedColumn := quotaType + "_used"
	err = r.db.QueryRow(fmt.Sprintf(`
SELECT %s
FROM user_quota_usages
WHERE user_id = ? AND period_key = ?`, usedColumn), userID, periodKey).Scan(&used)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			used = 0
		} else {
			return false, err
		}
	}

	if used >= limit {
		return false, nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.Exec(`
INSERT INTO user_quota_usages (id, user_id, member_level, period_key, updated_at)
VALUES (?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE member_level = VALUES(member_level), updated_at = VALUES(updated_at)`,
		newID("uqu"), userID, memberLevel, periodKey, now,
	)
	if err != nil {
		return false, err
	}

	_, err = tx.Exec(fmt.Sprintf(`
UPDATE user_quota_usages
SET %s = GREATEST(0, %s + 1), updated_at = ?
WHERE user_id = ? AND period_key = ?`, usedColumn, usedColumn), now, userID, periodKey)
	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	return true, nil
}
