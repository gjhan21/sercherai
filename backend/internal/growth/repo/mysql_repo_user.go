package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) ListBrowseHistory(userID string, contentType string, page int, pageSize int) ([]model.BrowseHistory, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{userID}
	filter := ""
	if contentType != "" {
		filter = " AND bh.content_type = ?"
		args = append(args, contentType)
	}

	countQuery := "SELECT COUNT(*) FROM browse_histories bh WHERE bh.user_id = ?" + filter
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT
	bh.id,
	bh.content_type,
	bh.content_id,
	COALESCE(
		na.title,
		CASE WHEN bh.content_type = 'STOCK' THEN CONCAT(sr.name, ' (', sr.symbol, ')') END,
		CASE WHEN bh.content_type = 'FORECAST' THEN CONCAT('深度推演: ', fr.target_label, ' (', fr.target_key, ')') END,
		CASE WHEN bh.content_type = 'ATTACHMENT' THEN CONCAT('下载附件: ', att.file_name) END,
		bh.content_id
	) AS title,
	bh.source_page,
	bh.viewed_at,
	COALESCE(sr.symbol, fr.target_key, '') AS target_key
FROM browse_histories bh
LEFT JOIN news_articles na
	ON bh.content_type = 'NEWS' AND na.id = bh.content_id
LEFT JOIN stock_recommendations sr
	ON bh.content_type = 'STOCK' AND sr.id = bh.content_id
LEFT JOIN strategy_forecast_l3_runs fr
	ON bh.content_type = 'FORECAST' AND fr.id = bh.content_id
LEFT JOIN news_attachments att
	ON bh.content_type = 'ATTACHMENT' AND att.id = bh.content_id
WHERE bh.user_id = ?` + filter + `
ORDER BY bh.viewed_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.BrowseHistory, 0)
	for rows.Next() {
		var item model.BrowseHistory
		var viewedAt time.Time
		if err := rows.Scan(&item.ID, &item.ContentType, &item.ContentID, &item.Title, &item.SourcePage, &viewedAt, &item.TargetKey); err != nil {
			return nil, 0, err
		}
		item.ViewedAt = viewedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) DeleteBrowseHistoryItem(userID string, id string) error {
	_, err := r.db.Exec("DELETE FROM browse_histories WHERE user_id = ? AND id = ?", userID, id)
	return err
}

func (r *MySQLGrowthRepo) ClearBrowseHistory(userID string) error {
	_, err := r.db.Exec("DELETE FROM browse_histories WHERE user_id = ?", userID)
	return err
}

func (r *MySQLGrowthRepo) ListRechargeRecords(userID string, status string, page int, pageSize int) ([]model.RechargeRecord, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{userID}
	filter := ""
	if status != "" {
		filter = " AND status = ?"
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM recharge_records WHERE user_id = ?"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, order_no, amount, pay_channel, status, paid_at, remark
FROM recharge_records
WHERE user_id = ?` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.RechargeRecord, 0)
	for rows.Next() {
		var item model.RechargeRecord
		var paidAt sql.NullTime
		var remark sql.NullString
		if err := rows.Scan(&item.ID, &item.OrderNo, &item.Amount, &item.PayChannel, &item.Status, &paidAt, &remark); err != nil {
			return nil, 0, err
		}
		if paidAt.Valid {
			item.PaidAt = paidAt.Time.Format(time.RFC3339)
		}
		if remark.Valid {
			item.Remark = remark.String
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) ListShareLinks(userID string) ([]model.ShareLink, error) {
	rows, err := r.db.Query(`
SELECT id, invite_code, url, channel, status, expired_at
FROM invite_links
WHERE user_id = ?
ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.ShareLink, 0)
	for rows.Next() {
		var item model.ShareLink
		var expiredAt sql.NullTime
		var channel sql.NullString
		if err := rows.Scan(&item.ID, &item.InviteCode, &item.URL, &channel, &item.Status, &expiredAt); err != nil {
			return nil, err
		}
		item.URL = normalizeInviteLinkURL(item.URL, item.InviteCode)
		if channel.Valid {
			item.Channel = channel.String
		}
		if expiredAt.Valid {
			item.ExpiredAt = expiredAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *MySQLGrowthRepo) CreateShareLink(userID string, channel string, expiredAt string) (model.ShareLink, error) {
	id := newID("sl")
	code := strings.ToUpper(newID("code"))
	shareURL := fmt.Sprintf("/invite/%s", url.PathEscape(code))
	status := "ACTIVE"

	var exp interface{} = nil
	if expiredAt != "" {
		t, err := time.Parse(time.RFC3339, expiredAt)
		if err == nil {
			exp = t
		}
	}
	_, err := r.db.Exec(`
INSERT INTO invite_links (id, user_id, invite_code, url, channel, status, expired_at, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, userID, code, shareURL, channel, status, exp, time.Now())
	if err != nil {
		return model.ShareLink{}, err
	}
	return model.ShareLink{
		ID:         id,
		InviteCode: code,
		URL:        shareURL,
		Channel:    channel,
		Status:     status,
		ExpiredAt:  expiredAt,
	}, nil
}

func (r *MySQLGrowthRepo) ListInviteRecords(userID string, page int, pageSize int) ([]model.InviteRecord, int, error) {
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM invite_records WHERE inviter_user_id = ?", userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT id, invitee_user_id, status, register_at, first_pay_at, risk_flag
FROM invite_records
WHERE inviter_user_id = ?
ORDER BY register_at DESC
LIMIT ? OFFSET ?`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.InviteRecord, 0)
	for rows.Next() {
		var item model.InviteRecord
		var registerAt, firstPayAt sql.NullTime
		var riskFlag sql.NullString
		if err := rows.Scan(&item.ID, &item.InviteeUser, &item.Status, &registerAt, &firstPayAt, &riskFlag); err != nil {
			return nil, 0, err
		}
		if registerAt.Valid {
			item.RegisterAt = registerAt.Time.Format(time.RFC3339)
		}
		if firstPayAt.Valid {
			item.FirstPayAt = firstPayAt.Time.Format(time.RFC3339)
		}
		if riskFlag.Valid {
			item.RiskFlag = riskFlag.String
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) GetUserInviteSummary(userID string) (model.InviteSummary, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return model.InviteSummary{}, sql.ErrNoRows
	}

	result := model.InviteSummary{}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM invite_links WHERE user_id = ?", userID).Scan(&result.ShareLinkCount); err != nil {
		return model.InviteSummary{}, err
	}

	var (
		registeredCount        sql.NullInt64
		firstPaidCount         sql.NullInt64
		last7dRegisteredCount  sql.NullInt64
		last7dFirstPaidCount   sql.NullInt64
		last30dRegisteredCount sql.NullInt64
		last30dFirstPaidCount  sql.NullInt64
	)
	err := r.db.QueryRow(`
SELECT
	SUM(CASE WHEN status IN ('REGISTERED', 'FIRST_PAID') THEN 1 ELSE 0 END) AS registered_count,
	SUM(CASE WHEN status = 'FIRST_PAID' OR first_pay_at IS NOT NULL THEN 1 ELSE 0 END) AS first_paid_count,
	SUM(CASE WHEN status IN ('REGISTERED', 'FIRST_PAID') AND register_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) THEN 1 ELSE 0 END) AS last_7d_registered_count,
	SUM(CASE WHEN status IN ('REGISTERED', 'FIRST_PAID') AND register_at >= DATE_SUB(NOW(), INTERVAL 7 DAY) AND (status = 'FIRST_PAID' OR first_pay_at IS NOT NULL) THEN 1 ELSE 0 END) AS last_7d_first_paid_count,
	SUM(CASE WHEN status IN ('REGISTERED', 'FIRST_PAID') AND register_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) THEN 1 ELSE 0 END) AS last_30d_registered_count,
	SUM(CASE WHEN status IN ('REGISTERED', 'FIRST_PAID') AND register_at >= DATE_SUB(NOW(), INTERVAL 30 DAY) AND (status = 'FIRST_PAID' OR first_pay_at IS NOT NULL) THEN 1 ELSE 0 END) AS last_30d_first_paid_count
FROM invite_records
WHERE inviter_user_id = ?`, userID).Scan(
		&registeredCount,
		&firstPaidCount,
		&last7dRegisteredCount,
		&last7dFirstPaidCount,
		&last30dRegisteredCount,
		&last30dFirstPaidCount,
	)
	if err != nil {
		return model.InviteSummary{}, err
	}

	if registeredCount.Valid {
		result.RegisteredCount = int(registeredCount.Int64)
	}
	if firstPaidCount.Valid {
		result.FirstPaidCount = int(firstPaidCount.Int64)
	}
	if last7dRegisteredCount.Valid {
		result.Last7dRegisteredCount = int(last7dRegisteredCount.Int64)
	}
	if last7dFirstPaidCount.Valid {
		result.Last7dFirstPaidCount = int(last7dFirstPaidCount.Int64)
	}
	if last30dRegisteredCount.Valid {
		result.Last30dRegisteredCount = int(last30dRegisteredCount.Int64)
	}
	if last30dFirstPaidCount.Valid {
		result.Last30dFirstPaidCount = int(last30dFirstPaidCount.Int64)
	}

	result.ConversionRate = safeRatioValue(int64(result.FirstPaidCount), int64(result.RegisteredCount))
	result.Last7dConversionRate = safeRatioValue(int64(result.Last7dFirstPaidCount), int64(result.Last7dRegisteredCount))
	result.Last30dConversionRate = safeRatioValue(int64(result.Last30dFirstPaidCount), int64(result.Last30dRegisteredCount))
	return result, nil
}

func (r *MySQLGrowthRepo) ListRewardRecords(userID string, page int, pageSize int) ([]model.RewardRecord, int, error) {
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM share_reward_records WHERE inviter_user_id = ?", userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT id, reward_type, reward_value, trigger_event, status, issued_at
FROM share_reward_records
WHERE inviter_user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.RewardRecord, 0)
	for rows.Next() {
		var item model.RewardRecord
		var issuedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.RewardType, &item.RewardValue, &item.TriggerEvent, &item.Status, &issuedAt); err != nil {
			return nil, 0, err
		}
		if issuedAt.Valid {
			item.IssuedAt = issuedAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) GetUserProfile(userID string) (model.UserProfile, error) {
	var profile model.UserProfile
	var email sql.NullString
	var inviterUserID sql.NullString
	var inviteCode sql.NullString
	var inviteLinkID sql.NullString
	var invitedAt sql.NullTime
	var vipStartedAt sql.NullTime
	var vipExpireAt sql.NullTime
	var registrationSource string
	err := r.db.QueryRow(`
SELECT
	u.id,
	u.phone,
	u.email,
	u.kyc_status,
	u.member_level,
	u.vip_started_at,
	u.vip_expire_at,
	CASE WHEN ir.id IS NULL THEN 'DIRECT' ELSE 'INVITED' END AS registration_source,
	ir.inviter_user_id,
	il.invite_code,
	ir.invite_link_id,
	ir.register_at
FROM users u
LEFT JOIN invite_records ir ON ir.invitee_user_id = u.id
LEFT JOIN invite_links il ON il.id = ir.invite_link_id
WHERE u.id = ?
LIMIT 1`, userID).Scan(
		&profile.ID,
		&profile.Phone,
		&email,
		&profile.KYCStatus,
		&profile.MemberLevel,
		&vipStartedAt,
		&vipExpireAt,
		&registrationSource,
		&inviterUserID,
		&inviteCode,
		&inviteLinkID,
		&invitedAt,
	)
	if err != nil {
		return model.UserProfile{}, err
	}
	if email.Valid {
		profile.Email = email.String
	}
	if vipStartedAt.Valid {
		profile.VIPStartedAt = vipStartedAt.Time.Format(time.RFC3339)
	}
	if vipExpireAt.Valid {
		profile.VIPExpireAt = vipExpireAt.Time.Format(time.RFC3339)
	}
	profile.VIPStatus, profile.VIPRemainingDays = resolveVIPStatusAndDays(profile.MemberLevel, vipExpireAt)
	profile.ActivationState = resolveMembershipActivationState(profile.MemberLevel, vipExpireAt)
	profile.RegistrationSource = registrationSource
	if inviterUserID.Valid {
		profile.InviterUserID = inviterUserID.String
	}
	if inviteCode.Valid {
		profile.InviteCode = inviteCode.String
	}
	if inviteLinkID.Valid {
		profile.InviteLinkID = inviteLinkID.String
	}
	if invitedAt.Valid {
		profile.InvitedAt = invitedAt.Time.Format(time.RFC3339)
	}
	return profile, nil
}

func (r *MySQLGrowthRepo) UpdateUserProfileEmail(userID string, email string) error {
	_, err := r.db.Exec("UPDATE users SET email = ?, updated_at = ? WHERE id = ?", strings.TrimSpace(email), time.Now(), userID)
	return err
}

func (r *MySQLGrowthRepo) ListSubscriptions(userID string, page int, pageSize int) ([]model.Subscription, int, error) {
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE user_id = ?", userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT id, type, scope, frequency, status
FROM subscriptions
WHERE user_id = ?
ORDER BY id DESC
LIMIT ? OFFSET ?`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.Subscription, 0)
	for rows.Next() {
		var item model.Subscription
		var scope sql.NullString
		if err := rows.Scan(&item.ID, &item.Type, &scope, &item.Frequency, &item.Status); err != nil {
			return nil, 0, err
		}
		if scope.Valid {
			item.Scope = scope.String
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) CreateSubscription(userID string, subType string, scope string, frequency string) (string, error) {
	id := newID("sub")
	_, err := r.db.Exec(`
INSERT INTO subscriptions (id, user_id, type, scope, frequency, status)
VALUES (?, ?, ?, ?, ?, 'ACTIVE')`,
		id, userID, strings.ToUpper(strings.TrimSpace(subType)), strings.TrimSpace(scope), strings.ToUpper(strings.TrimSpace(frequency)),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) UpdateSubscription(userID string, id string, frequency string, status string) error {
	res, err := r.db.Exec(`
UPDATE subscriptions
SET frequency = ?, status = ?
WHERE id = ? AND user_id = ?`,
		strings.ToUpper(strings.TrimSpace(frequency)), strings.ToUpper(strings.TrimSpace(status)), id, userID,
	)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) ListMessages(userID string, page int, pageSize int) ([]model.UserMessage, int, error) {
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM messages WHERE user_id = ?", userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT id, title, type, read_status, created_at
FROM messages
WHERE user_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.UserMessage, 0)
	for rows.Next() {
		var item model.UserMessage
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Title, &item.Type, &item.ReadStatus, &createdAt); err != nil {
			return nil, 0, err
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) MarkMessageRead(userID string, id string) error {
	res, err := r.db.Exec("UPDATE messages SET read_status = 'READ' WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) GetUserAccessProfile(userID string) (model.UserAccessProfile, error) {
	var profile model.UserAccessProfile
	var vipExpireAt sql.NullTime
	profile.UserID = userID
	err := r.db.QueryRow("SELECT status, kyc_status, member_level, vip_expire_at FROM users WHERE id = ?", userID).Scan(
		&profile.Status,
		&profile.KYCStatus,
		&profile.MemberLevel,
		&vipExpireAt,
	)
	if err != nil {
		return model.UserAccessProfile{}, err
	}
	profile.ActivationState = resolveMembershipActivationState(profile.MemberLevel, vipExpireAt)
	return profile, nil
}

func (r *MySQLGrowthRepo) GetMembershipQuota(userID string) (model.MembershipQuota, error) {
	var memberLevel string
	var kycStatus string
	var vipExpireAt sql.NullTime
	if err := r.db.QueryRow("SELECT member_level, kyc_status, vip_expire_at FROM users WHERE id = ?", userID).Scan(&memberLevel, &kycStatus, &vipExpireAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			memberLevel = "FREE"
			kycStatus = ""
		} else {
			return model.MembershipQuota{}, err
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
			return model.MembershipQuota{}, err
		}
		memberLevel = "FREE"
		vipExpireAt = sql.NullTime{}
	}

	now := time.Now()
	periodKey := now.Format("2006-01")

	var limitDoc, limitSub, limitDownload, limitForecast, limitStockReco int
	var resetCycle string
	err := r.db.QueryRow(`
SELECT doc_read_limit, news_subscribe_limit, download_limit, forecast_limit, stock_reco_limit, reset_cycle
FROM vip_quota_configs
WHERE member_level = ? AND status = 'ACTIVE'
ORDER BY effective_at DESC
LIMIT 1`, memberLevel).Scan(&limitDoc, &limitSub, &limitDownload, &limitForecast, &limitStockReco, &resetCycle)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.MembershipQuota{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		limitDoc = 0
		limitSub = 0
		limitDownload = 0
		limitForecast = 0
		limitStockReco = 0
		resetCycle = "MONTHLY"
	}

	var usedDoc, usedSub, usedDownload, usedForecast, usedStockReco int
	err = r.db.QueryRow(`
SELECT doc_read_used, news_subscribe_used, download_used, forecast_used, stock_reco_used
FROM user_quota_usages
WHERE user_id = ? AND period_key = ?`,
		userID, periodKey,
	).Scan(&usedDoc, &usedSub, &usedDownload, &usedForecast, &usedStockReco)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.MembershipQuota{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		usedDoc, usedSub, usedDownload, usedForecast, usedStockReco = 0, 0, 0, 0, 0
	}

	remainingDoc := limitDoc - usedDoc
	if remainingDoc < 0 {
		remainingDoc = 0
	}
	remainingSub := limitSub - usedSub
	if remainingSub < 0 {
		remainingSub = 0
	}
	remainingDownload := limitDownload - usedDownload
	if remainingDownload < 0 {
		remainingDownload = 0
	}
	remainingForecast := limitForecast - usedForecast
	if remainingForecast < 0 {
		remainingForecast = 0
	}
	remainingStockReco := limitStockReco - usedStockReco
	if remainingStockReco < 0 {
		remainingStockReco = 0
	}

	resetAt := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	if resetCycle == "DAILY" {
		resetAt = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	}

	return model.MembershipQuota{
		MemberLevel:            memberLevel,
		KYCStatus:              kycStatus,
		ActivationState:        resolveMembershipActivationState(memberLevel, vipExpireAt),
		PeriodKey:              periodKey,
		DocReadLimit:           limitDoc,
		DocReadUsed:            usedDoc,
		DocReadRemaining:       remainingDoc,
		NewsSubscribeLimit:     limitSub,
		NewsSubscribeUsed:      usedSub,
		NewsSubscribeRemaining: remainingSub,
		DownloadLimit:          limitDownload,
		DownloadUsed:           usedDownload,
		DownloadRemaining:      remainingDownload,
		ForecastLimit:          limitForecast,
		ForecastUsed:           usedForecast,
		ForecastRemaining:      remainingForecast,
		StockRecoLimit:         limitStockReco,
		StockRecoUsed:          usedStockReco,
		StockRecoRemaining:     remainingStockReco,
		ResetCycle:             resetCycle,
		ResetAt:                resetAt.Format(time.RFC3339),
		VIPExpireAt:            formatNullTime(vipExpireAt),
		VIPStatus:              vipStatusFromLevelAndExpire(memberLevel, vipExpireAt, now),
		VIPRemainingDays:       vipRemainingDays(vipExpireAt, now),
	}, nil
}

func (r *MySQLGrowthRepo) GetAttachmentFileInfo(userID string, attachmentID string) (model.AttachmentFileInfo, error) {
	isVIP, err := r.isVIPUser(userID)
	if err != nil {
		return model.AttachmentFileInfo{}, err
	}
	query := `
SELECT a.file_url, a.article_id
FROM news_attachments a
JOIN news_articles n ON a.article_id = n.id
WHERE a.id = ? AND n.status = 'PUBLISHED'`
	if !isVIP {
		query += " AND n.visibility = 'PUBLIC'"
	}
	var info model.AttachmentFileInfo
	if err := r.db.QueryRow(query, attachmentID).Scan(&info.FileURL, &info.ArticleID); err != nil {
		return model.AttachmentFileInfo{}, err
	}
	return info, nil
}

func (r *MySQLGrowthRepo) LogAttachmentDownload(userID string, attachmentID string, articleID string) error {
	allowed, err := r.CheckAndConsumeQuota(userID, "download", attachmentID)
	if err != nil {
		return err
	}
	if !allowed {
		return errors.New("quota exceeded")
	}

	_, _ = r.db.Exec(`
INSERT INTO browse_histories (id, user_id, content_type, content_id, source_page, viewed_at)
VALUES (?, ?, 'ATTACHMENT', ?, '/news', ?)`,
		newID("bh"), userID, attachmentID, time.Now(),
	)

	_, err = r.db.Exec(`
INSERT INTO attachment_download_logs (id, user_id, attachment_id, article_id, downloaded_at)
VALUES (?, ?, ?, ?, ?)`,
		newID("adl"), userID, attachmentID, articleID, time.Now(),
	)
	return err
}

func (r *MySQLGrowthRepo) ListNewsCategories(userID string) ([]model.NewsCategory, error) {
	rows, err := r.db.Query(`
SELECT id, name, slug, sort, visibility, status
FROM news_categories
WHERE status = 'PUBLISHED'
ORDER BY sort ASC, created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.NewsCategory, 0)
	for rows.Next() {
		var item model.NewsCategory
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Sort, &item.Visibility, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *MySQLGrowthRepo) ListNewsArticles(userID string, categoryID string, keyword string, page int, pageSize int) ([]model.NewsArticle, int, error) {
	offset := (page - 1) * pageSize

	args := []interface{}{}
	filter := " WHERE status = 'PUBLISHED'"
	if categoryID != "" {
		filter += " AND category_id = ?"
		args = append(args, categoryID)
	}
	if keyword != "" {
		filter += " AND (title LIKE ? OR summary LIKE ?)"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM news_articles"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, category_id, title, summary, cover_url, visibility, status, published_at, author_id
FROM news_articles` + filter + `
ORDER BY published_at DESC, created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.NewsArticle, 0)
	for rows.Next() {
		var item model.NewsArticle
		var summary, coverURL, authorID sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Title, &summary, &coverURL, &item.Visibility, &item.Status, &publishedAt, &authorID); err != nil {
			return nil, 0, err
		}
		if summary.Valid {
			item.Summary = summary.String
		}
		if coverURL.Valid {
			item.CoverURL = coverURL.String
		}
		if authorID.Valid {
			item.AuthorID = authorID.String
		}
		if publishedAt.Valid {
			item.PublishedAt = publishedAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) GetNewsArticleDetail(userID string, articleID string) (model.NewsArticle, error) {
	isVIP, err := r.isVIPUser(userID)
	if err != nil {
		return model.NewsArticle{}, err
	}
	query := `
SELECT na.id, na.category_id, na.title, na.summary, na.content, na.cover_url, na.visibility, na.status, na.published_at, na.author_id,
  (SELECT COUNT(*) FROM news_attachments att WHERE att.article_id = na.id) AS attachment_count
FROM news_articles na
WHERE na.id = ? AND na.status = 'PUBLISHED'`

	var item model.NewsArticle
	var summary, content, coverURL, authorID sql.NullString
	var publishedAt sql.NullTime
	err = r.db.QueryRow(query, articleID).Scan(
		&item.ID, &item.CategoryID, &item.Title, &summary, &content, &coverURL, &item.Visibility, &item.Status, &publishedAt, &authorID, &item.AttachmentCount,
	)
	if err != nil {
		return model.NewsArticle{}, err
	}
	if summary.Valid {
		item.Summary = summary.String
	}
	if coverURL.Valid {
		item.CoverURL = coverURL.String
	}
	if authorID.Valid {
		item.AuthorID = authorID.String
	}
	if publishedAt.Valid {
		item.PublishedAt = publishedAt.Time.Format(time.RFC3339)
	}

	quotaType := "doc_read"
	if item.Visibility == "VIP" {
		quotaType = "news_subscribe"
	}

	if item.Visibility == "VIP" && !isVIP {
		item.IsLocked = true
		item.Content = "此资讯为 VIP 专享，请先升级为 VIP 会员查看完整内容及下载附件。"
	} else {
		allowed := true
		if userID != "" {
			var checkErr error
			allowed, checkErr = r.CheckAndConsumeQuota(userID, quotaType, articleID)
			if checkErr != nil {
				return model.NewsArticle{}, checkErr
			}
		}
		if !allowed {
			item.IsLocked = true
			item.Content = "您的阅读额度已用尽，请下月再试或调整配额。"
		} else {
			item.IsLocked = false
			if content.Valid {
				item.Content = content.String
			}
			if userID != "" {
				_, _ = r.db.Exec(`
INSERT INTO browse_histories (id, user_id, content_type, content_id, source_page, viewed_at)
VALUES (?, ?, 'NEWS', ?, '/news', ?)`,
					newID("bh"), userID, articleID, time.Now(),
				)
			}
		}
	}

	return item, nil
}

func (r *MySQLGrowthRepo) ListNewsAttachments(userID string, articleID string) ([]model.NewsAttachment, error) {
	isVIP, err := r.isVIPUser(userID)
	if err != nil {
		return nil, err
	}

	query := "SELECT COUNT(*) FROM news_articles WHERE id = ? AND status = 'PUBLISHED'"
	args := []interface{}{articleID}
	if !isVIP {
		query += " AND visibility = 'PUBLIC'"
	}
	var articleCount int
	if err := r.db.QueryRow(query, args...).Scan(&articleCount); err != nil {
		return nil, err
	}
	if articleCount == 0 {
		return nil, sql.ErrNoRows
	}

	rows, err := r.db.Query(`
SELECT id, article_id, file_name, file_url, file_size, mime_type, created_at
FROM news_attachments
WHERE article_id = ?
ORDER BY created_at DESC`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.NewsAttachment, 0)
	for rows.Next() {
		var item model.NewsAttachment
		var mimeType sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.ArticleID, &item.FileName, &item.FileURL, &item.FileSize, &mimeType, &createdAt); err != nil {
			return nil, err
		}
		if mimeType.Valid {
			item.MimeType = mimeType.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, nil
}

func (r *MySQLGrowthRepo) GetRewardWallet(userID string) (model.RewardWallet, error) {
	var item model.RewardWallet
	err := r.db.QueryRow(`
SELECT cash_balance, cash_frozen, coupon_balance, vip_days_balance
FROM reward_wallets
WHERE user_id = ?`, userID).Scan(
		&item.CashBalance, &item.CashFrozen, &item.CouponBalance, &item.VIPDaysBalance,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.RewardWallet{}, nil
		}
		return model.RewardWallet{}, err
	}
	return item, nil
}

func (r *MySQLGrowthRepo) ListRewardWalletTxns(userID string, page int, pageSize int) ([]model.RewardWalletTxn, int, error) {
	offset := (page - 1) * pageSize
	var walletID string
	if err := r.db.QueryRow("SELECT id FROM reward_wallets WHERE user_id = ?", userID).Scan(&walletID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []model.RewardWalletTxn{}, 0, nil
		}
		return nil, 0, err
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM reward_wallet_txns WHERE wallet_id = ?", walletID).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT id, txn_type, amount, status, created_at
FROM reward_wallet_txns
WHERE wallet_id = ?
ORDER BY created_at DESC
LIMIT ? OFFSET ?`, walletID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.RewardWalletTxn, 0)
	for rows.Next() {
		var item model.RewardWalletTxn
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.TxnType, &item.Amount, &item.Status, &createdAt); err != nil {
			return nil, 0, err
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) CreateWithdrawRequest(userID string, amount float64) (string, error) {
	id := newID("wd")
	_, err := r.db.Exec(`
INSERT INTO withdraw_requests (id, user_id, amount, status, applied_at)
VALUES (?, ?, ?, 'PENDING', ?)`, id, userID, amount, time.Now())
	if err != nil {
		return "", err
	}
	return id, nil
}
