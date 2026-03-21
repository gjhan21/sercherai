package repo

import (
	"bytes"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"

	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/platform/config"
)

type MySQLGrowthRepo struct {
	db             *sql.DB
	redis          *redis.Client
	strategyEngine *strategyEngineClient
}

var repoIDSequence atomic.Uint64

func NewMySQLGrowthRepo(db *sql.DB, redisClient *redis.Client, cfg config.Config) *MySQLGrowthRepo {
	return &MySQLGrowthRepo{
		db:             db,
		redis:          redisClient,
		strategyEngine: newStrategyEngineClient(cfg),
	}
}

func normalizeInviteLinkURL(rawURL string, inviteCode string) string {
	trimmed := strings.TrimSpace(rawURL)
	code := strings.TrimSpace(inviteCode)
	if code == "" {
		return trimmed
	}
	if trimmed == "" {
		return "/invite/" + url.PathEscape(code)
	}
	lowerURL := strings.ToLower(trimmed)
	if strings.Contains(lowerURL, "example.com/invite/") {
		return "/invite/" + url.PathEscape(code)
	}
	return trimmed
}

func safeRatioValue(numerator int64, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

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
	COALESCE(NULLIF(na.title, ''), bh.content_id) AS title,
	bh.source_page,
	bh.viewed_at
FROM browse_histories bh
LEFT JOIN news_articles na
	ON bh.content_type = 'NEWS' AND na.id = bh.content_id
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
		if err := rows.Scan(&item.ID, &item.ContentType, &item.ContentID, &item.Title, &item.SourcePage, &viewedAt); err != nil {
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
	profile.ActivationState = resolveMembershipActivationState(profile.MemberLevel, profile.KYCStatus, vipExpireAt)
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

func (r *MySQLGrowthRepo) SubmitUserKYC(userID string, realName string, idNumber string, provider string) (string, error) {
	var status string
	if err := r.db.QueryRow("SELECT kyc_status FROM users WHERE id = ?", userID).Scan(&status); err != nil {
		return "", err
	}
	status = strings.ToUpper(strings.TrimSpace(status))
	if status == "APPROVED" {
		return "", errors.New("kyc already approved")
	}
	if status == "PENDING" {
		return "", errors.New("kyc pending")
	}

	kycID := newID("kyc")
	now := time.Now()
	if provider == "" {
		provider = "MANUAL"
	}
	_, err := r.db.Exec(`
INSERT INTO kyc_records (id, user_id, real_name, id_number, provider, status, reason, submitted_at, reviewed_at)
VALUES (?, ?, ?, ?, ?, 'PENDING', NULL, ?, NULL)`,
		kycID, userID, realName, idNumber, provider, now,
	)
	if err != nil {
		return "", err
	}
	if _, err := r.db.Exec("UPDATE users SET kyc_status = 'PENDING', updated_at = ? WHERE id = ?", now, userID); err != nil {
		return "", err
	}
	return "PENDING", nil
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
	profile.ActivationState = resolveMembershipActivationState(profile.MemberLevel, profile.KYCStatus, vipExpireAt)
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

	var limitDoc, limitSub int
	var resetCycle string
	err := r.db.QueryRow(`
SELECT doc_read_limit, news_subscribe_limit, reset_cycle
FROM vip_quota_configs
WHERE member_level = ? AND status = 'ACTIVE'
ORDER BY effective_at DESC
LIMIT 1`, memberLevel).Scan(&limitDoc, &limitSub, &resetCycle)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.MembershipQuota{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		limitDoc = 0
		limitSub = 0
		resetCycle = "MONTHLY"
	}

	var usedDoc, usedSub int
	err = r.db.QueryRow(`
SELECT doc_read_used, news_subscribe_used
FROM user_quota_usages
WHERE user_id = ? AND period_key = ?`,
		userID, periodKey,
	).Scan(&usedDoc, &usedSub)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.MembershipQuota{}, err
	}
	if errors.Is(err, sql.ErrNoRows) {
		usedDoc, usedSub = 0, 0
	}

	remainingDoc := limitDoc - usedDoc
	if remainingDoc < 0 {
		remainingDoc = 0
	}
	remainingSub := limitSub - usedSub
	if remainingSub < 0 {
		remainingSub = 0
	}

	resetAt := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	if resetCycle == "DAILY" {
		resetAt = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	}

	return model.MembershipQuota{
		MemberLevel:            memberLevel,
		KYCStatus:              kycStatus,
		ActivationState:        resolveMembershipActivationState(memberLevel, kycStatus, vipExpireAt),
		PeriodKey:              periodKey,
		DocReadLimit:           limitDoc,
		DocReadUsed:            usedDoc,
		DocReadRemaining:       remainingDoc,
		NewsSubscribeLimit:     limitSub,
		NewsSubscribeUsed:      usedSub,
		NewsSubscribeRemaining: remainingSub,
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
	_, err := r.db.Exec(`
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
SELECT id, category_id, title, summary, content, cover_url, visibility, status, published_at, author_id
FROM news_articles
WHERE id = ? AND status = 'PUBLISHED'`
	args := []interface{}{articleID}
	if !isVIP {
		query += " AND visibility = 'PUBLIC'"
	}
	var item model.NewsArticle
	var summary, content, coverURL, authorID sql.NullString
	var publishedAt sql.NullTime
	err = r.db.QueryRow(query, args...).Scan(
		&item.ID, &item.CategoryID, &item.Title, &summary, &content, &coverURL, &item.Visibility, &item.Status, &publishedAt, &authorID,
	)
	if err != nil {
		return model.NewsArticle{}, err
	}
	if summary.Valid {
		item.Summary = summary.String
	}
	if content.Valid {
		item.Content = content.String
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
	_, _ = r.db.Exec(`
INSERT INTO browse_histories (id, user_id, content_type, content_id, source_page, viewed_at)
VALUES (?, ?, 'NEWS', ?, '/news', ?)`,
		newID("bh"), userID, articleID, time.Now(),
	)
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

func (r *MySQLGrowthRepo) HandlePaymentCallback(channel string, orderNo string, channelTxnNo string, idempotencyKey string, sign string, signVerified bool) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	now := time.Now()
	verifiedVal := 0
	if signVerified {
		verifiedVal = 1
	}
	_, err = tx.Exec(`
INSERT INTO payment_callback_logs
(id, pay_channel, order_no, channel_txn_no, sign_verified, idempotency_key, callback_status, created_at)
VALUES (?, ?, ?, ?, ?, ?, 'RECEIVED', ?)`,
		newID("cb"), strings.ToUpper(channel), orderNo, channelTxnNo, verifiedVal, idempotencyKey, now,
	)
	if err != nil {
		_ = tx.Rollback()
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return errors.New("duplicate callback")
		}
		return err
	}

	var orderID, userID, productID, status string
	orderErr := tx.QueryRow(`
SELECT id, user_id, product_id, status
FROM membership_orders
WHERE order_no = ? OR id = ?
LIMIT 1`, orderNo, orderNo).Scan(&orderID, &userID, &productID, &status)
	if orderErr == nil {
		if strings.ToUpper(status) != "PAID" {
			if _, err := tx.Exec(`
UPDATE membership_orders
SET status = 'PAID', paid_at = ?, pay_channel = ?, updated_at = ?
WHERE id = ?`, now, strings.ToUpper(channel), now, orderID); err != nil {
				_ = tx.Rollback()
				return err
			}
			var memberLevel sql.NullString
			var durationDays sql.NullInt64
			if err := tx.QueryRow(
				"SELECT member_level, duration_days FROM membership_products WHERE id = ?",
				productID,
			).Scan(&memberLevel, &durationDays); err == nil {
				level := strings.ToUpper(strings.TrimSpace(memberLevel.String))
				days := int(durationDays.Int64)
				if level != "" && days <= 0 {
					days = 30
				}
				if level != "" {
					var currentLevel string
					var currentExpireAt sql.NullTime
					if err := tx.QueryRow(
						"SELECT member_level, vip_expire_at FROM users WHERE id = ? FOR UPDATE",
						userID,
					).Scan(&currentLevel, &currentExpireAt); err != nil {
						_ = tx.Rollback()
						return err
					}
					baseTime := now
					if strings.HasPrefix(strings.ToUpper(strings.TrimSpace(currentLevel)), "VIP") &&
						currentExpireAt.Valid &&
						currentExpireAt.Time.After(now) {
						baseTime = currentExpireAt.Time
					}
					newExpireAt := baseTime
					if days > 0 {
						newExpireAt = baseTime.AddDate(0, 0, days)
					}
					startedAt := now
					if _, err := tx.Exec(`
UPDATE users
SET member_level = ?,
    vip_started_at = ?,
    vip_expire_at = ?,
    vip_remind_3d_at = NULL,
    vip_remind_1d_at = NULL,
    updated_at = ?
WHERE id = ?`,
						level,
						startedAt,
						newExpireAt,
						now,
						userID,
					); err != nil {
						_ = tx.Rollback()
						return err
					}
					title := "VIP会员开通成功"
					content := fmt.Sprintf(
						"订单%s支付成功，会员等级已更新为%s，到期时间：%s。",
						strings.TrimSpace(orderNo),
						level,
						newExpireAt.Format("2006-01-02 15:04:05"),
					)
					if _, err := tx.Exec(`
INSERT INTO messages (id, user_id, title, content, type, read_status, created_at)
VALUES (?, ?, ?, ?, 'SYSTEM', 'UNREAD', ?)`,
						newID("msg"),
						userID,
						title,
						content,
						now,
					); err != nil {
						_ = tx.Rollback()
						return err
					}
				}
			}
			if err := createExperimentSuccessEventTx(tx, strings.TrimSpace(orderNo), now); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		return tx.Commit()
	}
	if !errors.Is(orderErr, sql.ErrNoRows) {
		_ = tx.Rollback()
		return orderErr
	}

	var rechargeStatus string
	rechargeErr := tx.QueryRow("SELECT status FROM recharge_records WHERE order_no = ?", orderNo).Scan(&rechargeStatus)
	if rechargeErr == nil {
		if strings.ToUpper(rechargeStatus) != "PAID" {
			if _, err := tx.Exec("UPDATE recharge_records SET status = 'PAID', paid_at = ? WHERE order_no = ?", now, orderNo); err != nil {
				_ = tx.Rollback()
				return err
			}
		}
		return tx.Commit()
	}
	if !errors.Is(rechargeErr, sql.ErrNoRows) {
		_ = tx.Rollback()
		return rechargeErr
	}

	_ = tx.Commit()
	return errors.New("order not found")
}

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

func (r *MySQLGrowthRepo) AdminListRewardRecords(status string, page int, pageSize int) ([]model.RewardRecord, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := ""
	if status != "" {
		filter = " WHERE status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM share_reward_records"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, inviter_user_id, invitee_user_id, reward_type, reward_value, status
FROM share_reward_records` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.RewardRecord, 0)
	for rows.Next() {
		var item model.RewardRecord
		if err := rows.Scan(&item.ID, &item.InviterUser, &item.InviteeUser, &item.RewardType, &item.RewardValue, &item.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminReviewRewardRecord(id string, status string, reason string) error {
	_, err := r.db.Exec("UPDATE share_reward_records SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *MySQLGrowthRepo) AdminListReconciliation(page int, pageSize int) ([]model.ReconciliationRecord, int, error) {
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM reconciliation_records").Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT id, pay_channel, batch_date, status, diff_count
FROM reconciliation_records
ORDER BY batch_date DESC
LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.ReconciliationRecord, 0)
	for rows.Next() {
		var item model.ReconciliationRecord
		var batchDate time.Time
		if err := rows.Scan(&item.ID, &item.PayChannel, &batchDate, &item.Status, &item.DiffCount); err != nil {
			return nil, 0, err
		}
		item.BatchDate = batchDate.Format("2006-01-02")
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminRetryReconciliation(batchID string) error {
	_, err := r.db.Exec("UPDATE reconciliation_records SET status = 'DONE' WHERE id = ?", batchID)
	return err
}

func (r *MySQLGrowthRepo) AdminListRiskRules() ([]model.RiskRule, error) {
	rows, err := r.db.Query("SELECT id, rule_code, rule_name, threshold, status FROM risk_rule_configs ORDER BY updated_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]model.RiskRule, 0)
	for rows.Next() {
		var item model.RiskRule
		if err := rows.Scan(&item.ID, &item.RuleCode, &item.RuleName, &item.Threshold, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *MySQLGrowthRepo) AdminCreateRiskRule(ruleCode string, ruleName string, threshold int, status string) (string, error) {
	id := newID("rule")
	_, err := r.db.Exec(`
INSERT INTO risk_rule_configs (id, rule_code, rule_name, threshold, status, effective_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`, id, ruleCode, ruleName, threshold, status, time.Now(), time.Now())
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateRiskRule(id string, threshold int, status string) error {
	_, err := r.db.Exec("UPDATE risk_rule_configs SET threshold = ?, status = ?, updated_at = ? WHERE id = ?", threshold, status, time.Now(), id)
	return err
}

func (r *MySQLGrowthRepo) AdminListRiskHits(status string, page int, pageSize int) ([]model.RiskHit, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := ""
	if status != "" {
		filter = " WHERE status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM risk_hit_logs"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, rule_code, user_id, risk_level, status
FROM risk_hit_logs` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.RiskHit, 0)
	for rows.Next() {
		var item model.RiskHit
		if err := rows.Scan(&item.ID, &item.RuleCode, &item.UserID, &item.RiskLevel, &item.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminReviewRiskHit(id string, status string, reason string) error {
	_, err := r.db.Exec("UPDATE risk_hit_logs SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *MySQLGrowthRepo) AdminListWithdrawRequests(page int, pageSize int) ([]model.WithdrawRequestInfo, int, error) {
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM withdraw_requests").Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT id, user_id, amount, status, applied_at
FROM withdraw_requests
ORDER BY applied_at DESC
LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.WithdrawRequestInfo, 0)
	for rows.Next() {
		var item model.WithdrawRequestInfo
		var appliedAt time.Time
		if err := rows.Scan(&item.ID, &item.UserID, &item.Amount, &item.Status, &appliedAt); err != nil {
			return nil, 0, err
		}
		item.AppliedAt = appliedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminReviewWithdrawRequest(id string, status string, reason string) error {
	_, err := r.db.Exec("UPDATE withdraw_requests SET status = ?, review_reason = ?, reviewed_at = ? WHERE id = ?", status, reason, time.Now(), id)
	return err
}

func (r *MySQLGrowthRepo) AdminListNewsCategories(status string, page int, pageSize int) ([]model.NewsCategory, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := ""
	if status != "" {
		filter = " WHERE status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM news_categories"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, name, slug, sort, visibility, status
FROM news_categories` + filter + `
ORDER BY sort ASC, created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.NewsCategory, 0)
	for rows.Next() {
		var item model.NewsCategory
		if err := rows.Scan(&item.ID, &item.Name, &item.Slug, &item.Sort, &item.Visibility, &item.Status); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateNewsCategory(name string, slug string, sort int, visibility string, status string) (string, error) {
	id := newID("nc")
	_, err := r.db.Exec(`
INSERT INTO news_categories (id, name, slug, sort, visibility, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, name, slug, sort, visibility, status, time.Now(), time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateNewsCategory(id string, name string, slug string, sort int, visibility string, status string) error {
	_, err := r.db.Exec(`
UPDATE news_categories
SET name = ?, slug = ?, sort = ?, visibility = ?, status = ?, updated_at = ?
WHERE id = ?`, name, slug, sort, visibility, status, time.Now(), id)
	return err
}

func (r *MySQLGrowthRepo) AdminListNewsArticles(status string, categoryID string, page int, pageSize int) ([]model.NewsArticle, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND na.status = ?"
		args = append(args, status)
	}
	if categoryID != "" {
		filter += " AND na.category_id = ?"
		args = append(args, categoryID)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM news_articles na"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT na.id, na.category_id, na.title, na.summary, na.cover_url, na.visibility, na.status, na.published_at, na.author_id,
  (SELECT COUNT(*) FROM news_attachments att WHERE att.article_id = na.id) AS attachment_count
FROM news_articles na` + filter + `
ORDER BY na.created_at DESC
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
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Title, &summary, &coverURL, &item.Visibility, &item.Status, &publishedAt, &authorID, &item.AttachmentCount); err != nil {
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

func (r *MySQLGrowthRepo) AdminGetNewsArticleDetail(id string) (model.NewsArticle, error) {
	query := `
SELECT na.id, na.category_id, na.title, na.summary, na.content, na.cover_url, na.visibility, na.status, na.published_at, na.author_id,
  (SELECT COUNT(*) FROM news_attachments att WHERE att.article_id = na.id) AS attachment_count
FROM news_articles na
WHERE na.id = ?`
	var item model.NewsArticle
	var summary, content, coverURL, authorID sql.NullString
	var publishedAt sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&item.ID, &item.CategoryID, &item.Title, &summary, &content, &coverURL, &item.Visibility, &item.Status, &publishedAt, &authorID, &item.AttachmentCount,
	)
	if err != nil {
		return model.NewsArticle{}, err
	}
	if summary.Valid {
		item.Summary = summary.String
	}
	if content.Valid {
		item.Content = content.String
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
	return item, nil
}

func (r *MySQLGrowthRepo) AdminCreateNewsArticle(categoryID string, title string, summary string, content string, coverURL string, visibility string, status string, authorID string) (string, error) {
	id := newID("na")
	_, err := r.db.Exec(`
INSERT INTO news_articles (id, category_id, title, summary, content, cover_url, visibility, status, published_at, author_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, categoryID, title, summary, content, nullableString(coverURL), visibility, status, time.Now(), authorID, time.Now(), time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateNewsArticle(id string, categoryID string, title string, summary string, content string, coverURL string, visibility string, status string) error {
	_, err := r.db.Exec(`
UPDATE news_articles
SET category_id = ?, title = ?, summary = ?, content = ?, cover_url = ?, visibility = ?, status = ?, updated_at = ?
WHERE id = ?`, categoryID, title, summary, content, nullableString(coverURL), visibility, status, time.Now(), id)
	return err
}

func (r *MySQLGrowthRepo) AdminPublishNewsArticle(id string, status string) error {
	result, err := r.db.Exec(`
UPDATE news_articles
SET status = ?, published_at = ?, updated_at = ?
WHERE id = ?`, status, time.Now(), time.Now(), id)
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

func (r *MySQLGrowthRepo) AdminCreateNewsAttachment(articleID string, fileName string, fileURL string, fileSize int64, mimeType string) (string, error) {
	id := newID("att")
	_, err := r.db.Exec(`
INSERT INTO news_attachments (id, article_id, file_name, file_url, file_size, mime_type, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, articleID, fileName, fileURL, fileSize, mimeType, time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminListNewsAttachments(articleID string) ([]model.NewsAttachment, error) {
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

func (r *MySQLGrowthRepo) AdminDeleteNewsAttachment(id string) error {
	result, err := r.db.Exec("DELETE FROM news_attachments WHERE id = ?", id)
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

type docFastSyncRuntimeConfig struct {
	Enabled       bool
	BatchSize     int
	SourceBaseURL string
	AuthorID      string
}

type docFastSyncCategoryTarget struct {
	CategoryID string
	Visibility string
}

type docFastSourceArticle struct {
	ID          int64
	ChannelID   int64
	Title       string
	Image       string
	Description string
	PublishUnix int64
	CreateUnix  int64
	UpdateUnix  int64
	Content     string
	DownloadURL string
}

type docFastDownloadAttachment struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Password string `json:"password"`
}

type docFastAttachmentMeta struct {
	FileName string
	MimeType string
	FileSize int64
}

const docFastSyncCheckpointKey = "doc_fast_news_incremental"
const tushareNewsIncrementalJobName = "tushare_news_incremental"

const (
	tushareNewsCategoryBriefSlug        = "ts-news-brief"
	tushareNewsCategoryMajorSlug        = "ts-news-major"
	tushareNewsCategoryResearchSlug     = "ts-report-research"
	tushareNewsCategoryForecastSlug     = "ts-report-forecast"
	tushareNewsCategoryAnnouncementSlug = "ts-announcement"
)

const (
	tushareSyncTypeNewsBrief    = "NEWS_BRIEF"
	tushareSyncTypeNewsMajor    = "NEWS_MAJOR"
	tushareSyncTypeResearch     = "RESEARCH_REPORT"
	tushareSyncTypeForecast     = "REPORT_RC"
	tushareSyncTypeAnnouncement = "ANNOUNCEMENT"
)

type tushareNewsSyncRuntimeConfig struct {
	Enabled                  bool
	BatchSize                int
	TimeoutMS                int
	AuthorID                 string
	SyncTypes                []string
	BriefSources             []string
	MajorSources             []string
	BriefLookbackHours       int
	MajorLookbackHours       int
	ResearchLookbackDays     int
	ForecastLookbackDays     int
	AnnouncementLookbackDays int
	BriefVisibility          string
	MajorVisibility          string
	ResearchVisibility       string
	ForecastVisibility       string
	AnnouncementVisibility   string
}

type tushareNewsCategoryTarget struct {
	CategoryID string
	Visibility string
}

type tushareNewsCategoryTargets struct {
	Brief        tushareNewsCategoryTarget
	Major        tushareNewsCategoryTarget
	Research     tushareNewsCategoryTarget
	Forecast     tushareNewsCategoryTarget
	Announcement tushareNewsCategoryTarget
}

type normalizedTushareSyncOptions struct {
	BatchSize   int
	SourceSet   map[string]struct{}
	SymbolSet   map[string]struct{}
	SyncTypeSet map[string]struct{}
}

type tushareNewsArticlePayload struct {
	ID             string
	CategoryID     string
	Visibility     string
	SyncType       string
	Source         string
	Symbol         string
	Title          string
	Summary        string
	Content        string
	PublishedAt    time.Time
	AuthorID       string
	AttachmentURL  string
	AttachmentName string
	AttachmentMime string
}

func (r *MySQLGrowthRepo) AdminSyncDocFastNewsIncremental(batchSize int) (string, error) {
	runtime := r.resolveDocFastSyncRuntimeConfig(batchSize)
	if !runtime.Enabled {
		return "doc_fast incremental sync disabled", nil
	}
	categoryTargets, err := r.resolveDocFastSyncCategoryTargets()
	if err != nil {
		return "", err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return "", err
	}
	rollbacked := false
	defer func() {
		if !rollbacked {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.Exec(`
INSERT INTO news_sync_checkpoints
	(sync_key, cursor_updated_at, cursor_source_id, last_status, synced_articles, synced_attachments, updated_at, created_at)
VALUES
	(?, 0, 0, 'IDLE', 0, 0, NOW(), NOW())
ON DUPLICATE KEY UPDATE
	sync_key = VALUES(sync_key)`, docFastSyncCheckpointKey); err != nil {
		rollbacked = true
		_ = tx.Rollback()
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}

	cursorUpdatedAt := int64(0)
	cursorSourceID := int64(0)
	if err := tx.QueryRow(`
SELECT cursor_updated_at, cursor_source_id
FROM news_sync_checkpoints
WHERE sync_key = ?
FOR UPDATE`, docFastSyncCheckpointKey).Scan(&cursorUpdatedAt, &cursorSourceID); err != nil {
		rollbacked = true
		_ = tx.Rollback()
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}

	sourceItems, err := r.loadDocFastSourceArticlesForSync(tx, cursorUpdatedAt, cursorSourceID, runtime.BatchSize)
	if err != nil {
		rollbacked = true
		_ = tx.Rollback()
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}

	syncedArticles := 0
	syncedAttachments := 0
	nextCursorUpdatedAt := cursorUpdatedAt
	nextCursorSourceID := cursorSourceID

	for _, item := range sourceItems {
		target, ok := categoryTargets[item.ChannelID]
		if !ok || strings.TrimSpace(target.CategoryID) == "" {
			continue
		}
		articleID := fmt.Sprintf("na_df_%d", item.ID)
		title := truncateByRunes(normalizeUTF8Text(item.Title), 256)
		if title == "" {
			title = fmt.Sprintf("doc_fast_%d", item.ID)
		}
		summary := truncateByRunes(normalizeUTF8Text(item.Description), 512)
		if summary == "" {
			summary = title
		}
		content := normalizeUTF8Text(item.Content)
		if content == "" {
			content = title
		}
		content = strings.ToValidUTF8(content, "")

		coverURL := normalizeDocFastAssetURL(runtime.SourceBaseURL, item.Image)
		createdAt := fromUnixSecondOrNow(item.CreateUnix)
		updatedAt := fromUnixSecondOrNow(item.UpdateUnix)
		publishedAt := nullableUnixSecond(item.PublishUnix)
		visibility := strings.ToUpper(strings.TrimSpace(target.Visibility))
		if visibility == "" {
			visibility = "PUBLIC"
		}

		_, err := tx.Exec(`
INSERT INTO news_articles
	(id, category_id, title, summary, content, cover_url, tags, visibility, status, published_at, author_id, created_at, updated_at)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, 'PUBLISHED', ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	category_id = VALUES(category_id),
	title = VALUES(title),
	summary = VALUES(summary),
	content = VALUES(content),
	cover_url = VALUES(cover_url),
	visibility = VALUES(visibility),
	status = VALUES(status),
	published_at = VALUES(published_at),
	author_id = VALUES(author_id),
	updated_at = VALUES(updated_at)`,
			articleID,
			target.CategoryID,
			title,
			summary,
			content,
			nullableString(coverURL),
			nil,
			visibility,
			publishedAt,
			runtime.AuthorID,
			createdAt,
			updatedAt,
		)
		if err != nil {
			rollbacked = true
			_ = tx.Rollback()
			r.markDocFastSyncFailed(err.Error())
			return "", err
		}
		syncedArticles++

		attachments := parseDocFastDownloadAttachments(item.DownloadURL)
		for idx, att := range attachments {
			normalizedURL := normalizeDocFastAssetURL(runtime.SourceBaseURL, att.URL)
			if normalizedURL == "" {
				continue
			}
			meta, metaErr := r.loadDocFastAttachmentMeta(tx, att.URL)
			if metaErr != nil {
				rollbacked = true
				_ = tx.Rollback()
				r.markDocFastSyncFailed(metaErr.Error())
				return "", metaErr
			}
			fileName := normalizeUTF8Text(meta.FileName)
			if fileName == "" {
				fileName = inferDocFastAttachmentFileName(normalizedURL, title, idx+1)
			}
			fileName = truncateByRunes(fileName, 256)
			if fileName == "" {
				fileName = fmt.Sprintf("attachment_%d", idx+1)
			}

			mimeType := truncateByRunes(normalizeUTF8Text(meta.MimeType), 128)
			attachmentID := fmt.Sprintf("nat_df_%d_%d", item.ID, idx+1)
			_, err = tx.Exec(`
INSERT INTO news_attachments
	(id, article_id, file_name, file_url, file_size, mime_type, created_at)
VALUES
	(?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	file_name = VALUES(file_name),
	file_url = VALUES(file_url),
	file_size = VALUES(file_size),
	mime_type = VALUES(mime_type)`,
				attachmentID,
				articleID,
				fileName,
				truncateByRunes(normalizedURL, 512),
				meta.FileSize,
				nullableString(mimeType),
				updatedAt,
			)
			if err != nil {
				rollbacked = true
				_ = tx.Rollback()
				r.markDocFastSyncFailed(err.Error())
				return "", err
			}
			syncedAttachments++
		}

		if item.UpdateUnix > nextCursorUpdatedAt || (item.UpdateUnix == nextCursorUpdatedAt && item.ID > nextCursorSourceID) {
			nextCursorUpdatedAt = item.UpdateUnix
			nextCursorSourceID = item.ID
		}
	}

	now := time.Now()
	_, err = tx.Exec(`
UPDATE news_sync_checkpoints
SET
	cursor_updated_at = ?,
	cursor_source_id = ?,
	last_run_at = ?,
	last_success_at = ?,
	last_status = 'SUCCESS',
	last_error = NULL,
	synced_articles = synced_articles + ?,
	synced_attachments = synced_attachments + ?,
	updated_at = ?
WHERE sync_key = ?`,
		nextCursorUpdatedAt,
		nextCursorSourceID,
		now,
		now,
		syncedArticles,
		syncedAttachments,
		now,
		docFastSyncCheckpointKey,
	)
	if err != nil {
		rollbacked = true
		_ = tx.Rollback()
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}

	if err := tx.Commit(); err != nil {
		rollbacked = true
		_ = tx.Rollback()
		r.markDocFastSyncFailed(err.Error())
		return "", err
	}
	rollbacked = true

	return fmt.Sprintf(
		"synced_articles=%d synced_attachments=%d cursor_updated_at=%d cursor_source_id=%d batch=%d",
		syncedArticles,
		syncedAttachments,
		nextCursorUpdatedAt,
		nextCursorSourceID,
		runtime.BatchSize,
	), nil
}

func (r *MySQLGrowthRepo) resolveDocFastSyncRuntimeConfig(batchSize int) docFastSyncRuntimeConfig {
	cfg := docFastSyncRuntimeConfig{
		Enabled:       true,
		BatchSize:     200,
		SourceBaseURL: "https://img.cloudup518.top",
		AuthorID:      "admin_001",
	}
	items, _, err := r.AdminListSystemConfigs("news.sync.doc_fast.", 1, 200)
	if err == nil {
		for _, item := range items {
			key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
			value := strings.TrimSpace(item.ConfigValue)
			switch key {
			case "news.sync.doc_fast.enabled":
				cfg.Enabled = parseRepoConfigBool(value, cfg.Enabled)
			case "news.sync.doc_fast.batch_size":
				cfg.BatchSize = parseRepoConfigInt(value, cfg.BatchSize)
			case "news.sync.doc_fast.source_base_url":
				if value != "" {
					cfg.SourceBaseURL = value
				}
			case "news.sync.doc_fast.author_id":
				if value != "" {
					cfg.AuthorID = value
				}
			}
		}
	}
	if batchSize > 0 {
		cfg.BatchSize = batchSize
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 200
	}
	if cfg.BatchSize > 5000 {
		cfg.BatchSize = 5000
	}
	cfg.SourceBaseURL = strings.TrimSpace(cfg.SourceBaseURL)
	if cfg.SourceBaseURL == "" {
		cfg.SourceBaseURL = "https://img.cloudup518.top"
	}
	if !strings.HasPrefix(strings.ToLower(cfg.SourceBaseURL), "http://") && !strings.HasPrefix(strings.ToLower(cfg.SourceBaseURL), "https://") {
		cfg.SourceBaseURL = "https://" + strings.TrimLeft(cfg.SourceBaseURL, "/")
	}
	cfg.SourceBaseURL = strings.TrimRight(cfg.SourceBaseURL, "/")
	cfg.AuthorID = strings.TrimSpace(cfg.AuthorID)
	if cfg.AuthorID == "" {
		cfg.AuthorID = "admin_001"
	}
	return cfg
}

func (r *MySQLGrowthRepo) resolveDocFastSyncCategoryTargets() (map[int64]docFastSyncCategoryTarget, error) {
	rows, err := r.db.Query(`
SELECT id, name, visibility
FROM news_categories
WHERE name IN ('国内研报', '海外研报', '期刊', '图书')
ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type categoryInfo struct {
		ID         string
		Visibility string
	}
	byName := make(map[string]categoryInfo)
	for rows.Next() {
		var (
			id         string
			name       string
			visibility string
		)
		if err := rows.Scan(&id, &name, &visibility); err != nil {
			return nil, err
		}
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		if _, exists := byName[name]; exists {
			continue
		}
		byName[name] = categoryInfo{
			ID:         strings.TrimSpace(id),
			Visibility: strings.ToUpper(strings.TrimSpace(visibility)),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	required := []string{"国内研报", "海外研报", "期刊", "图书"}
	missing := make([]string, 0)
	for _, name := range required {
		if _, ok := byName[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing target news categories: %s", strings.Join(missing, ","))
	}

	targets := map[int64]docFastSyncCategoryTarget{
		13: {CategoryID: byName["国内研报"].ID, Visibility: byName["国内研报"].Visibility},
		14: {CategoryID: byName["海外研报"].ID, Visibility: byName["海外研报"].Visibility},
		26: {CategoryID: byName["期刊"].ID, Visibility: byName["期刊"].Visibility},
		30: {CategoryID: byName["图书"].ID, Visibility: byName["图书"].Visibility},
		31: {CategoryID: byName["图书"].ID, Visibility: byName["图书"].Visibility},
		33: {CategoryID: byName["图书"].ID, Visibility: byName["图书"].Visibility},
		34: {CategoryID: byName["图书"].ID, Visibility: byName["图书"].Visibility},
	}
	return targets, nil
}

func (r *MySQLGrowthRepo) loadDocFastSourceArticlesForSync(tx *sql.Tx, cursorUpdatedAt int64, cursorSourceID int64, limit int) ([]docFastSourceArticle, error) {
	rows, err := tx.Query(`
SELECT
	a.id,
	a.channel_id,
	a.title,
	a.image,
	a.description,
	COALESCE(a.publishtime, 0) AS publishtime,
	COALESCE(a.createtime, 0) AS createtime,
	COALESCE(a.updatetime, 0) AS updatetime,
	COALESCE(ad.content, '') AS content,
	COALESCE(ad.downloadurl, '') AS downloadurl
FROM doc_fast.fa_cms_archives a
LEFT JOIN doc_fast.fa_cms_addondownload ad ON ad.id = a.id
WHERE
	a.status = 'normal'
	AND a.channel_id IN (13, 14, 26, 30, 31, 33, 34)
	AND (
		COALESCE(a.updatetime, 0) > ?
		OR (COALESCE(a.updatetime, 0) = ? AND a.id > ?)
	)
ORDER BY COALESCE(a.updatetime, 0) ASC, a.id ASC
LIMIT ?`, cursorUpdatedAt, cursorUpdatedAt, cursorSourceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]docFastSourceArticle, 0, limit)
	for rows.Next() {
		var item docFastSourceArticle
		if err := rows.Scan(
			&item.ID,
			&item.ChannelID,
			&item.Title,
			&item.Image,
			&item.Description,
			&item.PublishUnix,
			&item.CreateUnix,
			&item.UpdateUnix,
			&item.Content,
			&item.DownloadURL,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *MySQLGrowthRepo) loadDocFastAttachmentMeta(tx *sql.Tx, sourceURL string) (docFastAttachmentMeta, error) {
	candidate := strings.TrimSpace(sourceURL)
	if candidate == "" {
		return docFastAttachmentMeta{}, nil
	}
	alt := candidate
	if strings.HasPrefix(candidate, "/") {
		alt = strings.TrimPrefix(candidate, "/")
	} else {
		alt = "/" + candidate
	}
	var (
		fileName sql.NullString
		mimeType sql.NullString
		fileSize sql.NullInt64
	)
	err := tx.QueryRow(`
SELECT filename, mimetype, filesize
FROM doc_fast.fa_attachment
WHERE url IN (?, ?)
ORDER BY id DESC
LIMIT 1`, candidate, alt).Scan(&fileName, &mimeType, &fileSize)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return docFastAttachmentMeta{}, nil
		}
		return docFastAttachmentMeta{}, err
	}
	result := docFastAttachmentMeta{}
	if fileName.Valid {
		result.FileName = fileName.String
	}
	if mimeType.Valid {
		result.MimeType = mimeType.String
	}
	if fileSize.Valid {
		result.FileSize = fileSize.Int64
	}
	return result, nil
}

func parseDocFastDownloadAttachments(raw string) []docFastDownloadAttachment {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil
	}
	items := make([]docFastDownloadAttachment, 0)
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		return nil
	}
	return items
}

func normalizeDocFastAssetURL(baseURL string, raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	lower := strings.ToLower(value)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return value
	}
	if strings.HasPrefix(value, "//") {
		return "https:" + value
	}
	trimmedBase := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if trimmedBase == "" {
		trimmedBase = "https://img.cloudup518.top"
	}
	if strings.HasPrefix(value, "/") {
		return trimmedBase + value
	}
	return trimmedBase + "/" + strings.TrimLeft(value, "/")
}

func inferDocFastAttachmentFileName(fileURL string, title string, index int) string {
	trimmedURL := strings.TrimSpace(fileURL)
	base := path.Base(trimmedURL)
	if base != "" && base != "." && base != "/" {
		return base
	}
	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		cleanTitle = "attachment"
	}
	return fmt.Sprintf("%s_%d", cleanTitle, index)
}

func truncateByRunes(value string, size int) string {
	if size <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= size {
		return value
	}
	return string(runes[:size])
}

func normalizeUTF8Text(value string) string {
	return strings.TrimSpace(strings.ToValidUTF8(value, ""))
}

func fromUnixSecondOrNow(value int64) time.Time {
	if value > 0 {
		return time.Unix(value, 0)
	}
	return time.Now()
}

func nullableUnixSecond(value int64) interface{} {
	if value > 0 {
		return time.Unix(value, 0)
	}
	return nil
}

func (r *MySQLGrowthRepo) markDocFastSyncFailed(errText string) {
	now := time.Now()
	_, _ = r.db.Exec(`
UPDATE news_sync_checkpoints
SET
	last_run_at = ?,
	last_status = 'FAILED',
	last_error = ?,
	updated_at = ?
WHERE sync_key = ?`,
		now,
		truncateByRunes(normalizeUTF8Text(errText), 255),
		now,
		docFastSyncCheckpointKey,
	)
}

func (r *MySQLGrowthRepo) AdminSyncTushareNewsIncremental(batchSize int) (string, error) {
	summary, _, err := r.AdminSyncTushareNewsIncrementalWithOptions(model.TushareNewsSyncOptions{
		BatchSize: batchSize,
	})
	return summary, err
}

func (r *MySQLGrowthRepo) AdminSyncTushareNewsIncrementalWithOptions(opts model.TushareNewsSyncOptions) (string, []model.NewsSyncRunDetail, error) {
	normalizedOpts := normalizeTushareSyncOptions(opts)
	runtime := r.resolveTushareNewsSyncRuntimeConfig(normalizedOpts.BatchSize)
	if !runtime.Enabled {
		return "tushare news incremental sync disabled", nil, nil
	}
	if len(normalizedOpts.SyncTypeSet) == 0 {
		for _, syncType := range runtime.SyncTypes {
			key := normalizeTushareSyncType(syncType)
			if key == "" {
				continue
			}
			normalizedOpts.SyncTypeSet[key] = struct{}{}
		}
	}
	if err := r.ensureTushareNewsCategories(); err != nil {
		return "", nil, err
	}
	targets, err := r.resolveTushareNewsCategoryTargets(runtime)
	if err != nil {
		return "", nil, err
	}
	token, err := r.resolveTushareTokenForNewsSync()
	if err != nil {
		return "", nil, err
	}

	now := time.Now()
	mergedItems := make(map[string]tushareNewsArticlePayload)
	fetchWarnings := make([]string, 0)
	detailItems := make([]model.NewsSyncRunDetail, 0, 32)
	counts := map[string]int{
		"brief":        0,
		"major":        0,
		"research":     0,
		"forecast":     0,
		"announcement": 0,
	}

	mergeItems := func(counterKey string, items []tushareNewsArticlePayload, details []model.NewsSyncRunDetail) {
		counts[counterKey] += len(items)
		detailItems = append(detailItems, details...)
		for _, item := range items {
			if strings.TrimSpace(item.ID) == "" || strings.TrimSpace(item.CategoryID) == "" {
				continue
			}
			existing, exists := mergedItems[item.ID]
			if !exists || item.PublishedAt.After(existing.PublishedAt) {
				mergedItems[item.ID] = item
			}
		}
	}

	if shouldSyncTushareType(normalizedOpts.SyncTypeSet, tushareSyncTypeNewsBrief) {
		items, details, fetchErr := fetchTushareNewsBriefArticles(token, runtime, targets.Brief, now, normalizedOpts)
		mergeItems("brief", items, details)
		if fetchErr != nil {
			fetchWarnings = append(fetchWarnings, fmt.Sprintf("news: %v", fetchErr))
		}
	}

	if shouldSyncTushareType(normalizedOpts.SyncTypeSet, tushareSyncTypeNewsMajor) {
		items, details, fetchErr := fetchTushareMajorNewsArticles(token, runtime, targets.Major, now, normalizedOpts)
		mergeItems("major", items, details)
		if fetchErr != nil {
			fetchWarnings = append(fetchWarnings, fmt.Sprintf("major_news: %v", fetchErr))
		}
	}

	if shouldSyncTushareType(normalizedOpts.SyncTypeSet, tushareSyncTypeResearch) {
		items, details, fetchErr := fetchTushareResearchReportArticles(token, runtime, targets.Research, now, normalizedOpts)
		mergeItems("research", items, details)
		if fetchErr != nil {
			fetchWarnings = append(fetchWarnings, fmt.Sprintf("research_report: %v", fetchErr))
		}
	}

	if shouldSyncTushareType(normalizedOpts.SyncTypeSet, tushareSyncTypeForecast) {
		items, details, fetchErr := fetchTushareForecastArticles(token, runtime, targets.Forecast, now, normalizedOpts)
		mergeItems("forecast", items, details)
		if fetchErr != nil {
			fetchWarnings = append(fetchWarnings, fmt.Sprintf("report_rc: %v", fetchErr))
		}
	}

	if shouldSyncTushareType(normalizedOpts.SyncTypeSet, tushareSyncTypeAnnouncement) {
		items, details, fetchErr := fetchTushareAnnouncementArticles(token, runtime, targets.Announcement, now, normalizedOpts)
		mergeItems("announcement", items, details)
		if fetchErr != nil {
			fetchWarnings = append(fetchWarnings, fmt.Sprintf("anns_d: %v", fetchErr))
		}
	}

	readyItems := make([]tushareNewsArticlePayload, 0, len(mergedItems))
	for _, item := range mergedItems {
		readyItems = append(readyItems, item)
	}

	syncedArticles, syncedAttachments, err := r.upsertTushareNewsArticles(readyItems)
	if err != nil {
		return "", detailItems, err
	}
	applyTushareDetailUpsertCounts(detailItems, readyItems)

	summary := fmt.Sprintf(
		"brief=%d major=%d research=%d forecast=%d announcement=%d upserted=%d attachments=%d",
		counts["brief"],
		counts["major"],
		counts["research"],
		counts["forecast"],
		counts["announcement"],
		syncedArticles,
		syncedAttachments,
	)
	if len(normalizedOpts.SourceSet) > 0 {
		summary += fmt.Sprintf(" source_filter=%d", len(normalizedOpts.SourceSet))
	}
	if len(normalizedOpts.SymbolSet) > 0 {
		summary += fmt.Sprintf(" symbol_filter=%d", len(normalizedOpts.SymbolSet))
	}
	if len(fetchWarnings) > 0 {
		summary += " warnings=" + strings.Join(compactErrorMessages(fetchWarnings, 3), " | ")
	}
	if len(readyItems) == 0 && len(fetchWarnings) > 0 {
		if onlyRateLimitMessages(fetchWarnings) {
			summary += " rate_limited=true"
			return summary, detailItems, nil
		}
		return summary, detailItems, errors.New(strings.Join(compactErrorMessages(fetchWarnings, 2), "; "))
	}
	return summary, detailItems, nil
}

func (r *MySQLGrowthRepo) resolveTushareNewsSyncRuntimeConfig(batchSize int) tushareNewsSyncRuntimeConfig {
	cfg := tushareNewsSyncRuntimeConfig{
		Enabled:                  true,
		BatchSize:                200,
		TimeoutMS:                12000,
		AuthorID:                 "admin_001",
		SyncTypes:                []string{tushareSyncTypeNewsBrief, tushareSyncTypeNewsMajor},
		BriefSources:             []string{"cls", "yicai", "sina", "eastmoney"},
		MajorSources:             []string{"新华网", "财联社", "第一财经", "新浪财经", "华尔街见闻"},
		BriefLookbackHours:       8,
		MajorLookbackHours:       24,
		ResearchLookbackDays:     7,
		ForecastLookbackDays:     7,
		AnnouncementLookbackDays: 2,
		BriefVisibility:          "PUBLIC",
		MajorVisibility:          "PUBLIC",
		ResearchVisibility:       "VIP",
		ForecastVisibility:       "VIP",
		AnnouncementVisibility:   "PUBLIC",
	}

	items, _, err := r.AdminListSystemConfigs("news.sync.tushare.", 1, 200)
	if err == nil {
		for _, item := range items {
			key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
			value := strings.TrimSpace(item.ConfigValue)
			switch key {
			case "news.sync.tushare.enabled":
				cfg.Enabled = parseRepoConfigBool(value, cfg.Enabled)
			case "news.sync.tushare.batch_size":
				cfg.BatchSize = parseRepoConfigInt(value, cfg.BatchSize)
			case "news.sync.tushare.timeout_ms":
				cfg.TimeoutMS = parseRepoConfigInt(value, cfg.TimeoutMS)
			case "news.sync.tushare.author_id":
				if value != "" {
					cfg.AuthorID = value
				}
			case "news.sync.tushare.sync_types":
				cfg.SyncTypes = splitTushareSyncTypeList(value, cfg.SyncTypes)
			case "news.sync.tushare.sources.news_brief":
				cfg.BriefSources = splitTushareSourceList(value, cfg.BriefSources)
			case "news.sync.tushare.sources.news_major":
				cfg.MajorSources = splitTushareSourceList(value, cfg.MajorSources)
			case "news.sync.tushare.lookback_hours.news_brief":
				cfg.BriefLookbackHours = parseRepoConfigInt(value, cfg.BriefLookbackHours)
			case "news.sync.tushare.lookback_hours.news_major":
				cfg.MajorLookbackHours = parseRepoConfigInt(value, cfg.MajorLookbackHours)
			case "news.sync.tushare.lookback_days.research_report":
				cfg.ResearchLookbackDays = parseRepoConfigInt(value, cfg.ResearchLookbackDays)
			case "news.sync.tushare.lookback_days.report_rc":
				cfg.ForecastLookbackDays = parseRepoConfigInt(value, cfg.ForecastLookbackDays)
			case "news.sync.tushare.lookback_days.announcement":
				cfg.AnnouncementLookbackDays = parseRepoConfigInt(value, cfg.AnnouncementLookbackDays)
			case "news.sync.tushare.visibility.news_brief":
				cfg.BriefVisibility = value
			case "news.sync.tushare.visibility.news_major":
				cfg.MajorVisibility = value
			case "news.sync.tushare.visibility.research_report":
				cfg.ResearchVisibility = value
			case "news.sync.tushare.visibility.report_rc":
				cfg.ForecastVisibility = value
			case "news.sync.tushare.visibility.announcement":
				cfg.AnnouncementVisibility = value
			}
		}
	}

	if batchSize > 0 {
		cfg.BatchSize = batchSize
	}
	cfg.BatchSize = maxInt(20, minInt(cfg.BatchSize, 1000))
	cfg.TimeoutMS = maxInt(2000, minInt(cfg.TimeoutMS, 30000))
	cfg.BriefLookbackHours = maxInt(1, minInt(cfg.BriefLookbackHours, 72))
	cfg.MajorLookbackHours = maxInt(1, minInt(cfg.MajorLookbackHours, 168))
	cfg.ResearchLookbackDays = maxInt(1, minInt(cfg.ResearchLookbackDays, 60))
	cfg.ForecastLookbackDays = maxInt(1, minInt(cfg.ForecastLookbackDays, 60))
	cfg.AnnouncementLookbackDays = maxInt(1, minInt(cfg.AnnouncementLookbackDays, 30))
	cfg.SyncTypes = splitTushareSyncTypeList(strings.Join(cfg.SyncTypes, ","), []string{tushareSyncTypeNewsBrief, tushareSyncTypeNewsMajor})
	cfg.BriefSources = splitTushareSourceList(strings.Join(cfg.BriefSources, ","), []string{"cls", "yicai", "sina", "eastmoney"})
	cfg.MajorSources = splitTushareSourceList(strings.Join(cfg.MajorSources, ","), []string{"新华网", "财联社", "第一财经", "新浪财经", "华尔街见闻"})
	cfg.AuthorID = strings.TrimSpace(cfg.AuthorID)
	if cfg.AuthorID == "" {
		cfg.AuthorID = "admin_001"
	}
	cfg.BriefVisibility = normalizeTushareNewsVisibility(cfg.BriefVisibility)
	cfg.MajorVisibility = normalizeTushareNewsVisibility(cfg.MajorVisibility)
	cfg.ResearchVisibility = normalizeTushareNewsVisibility(cfg.ResearchVisibility)
	cfg.ForecastVisibility = normalizeTushareNewsVisibility(cfg.ForecastVisibility)
	cfg.AnnouncementVisibility = normalizeTushareNewsVisibility(cfg.AnnouncementVisibility)
	return cfg
}

func (r *MySQLGrowthRepo) ensureTushareNewsCategories() error {
	type categorySeed struct {
		ID         string
		Name       string
		Slug       string
		Sort       int
		Visibility string
	}
	seeds := []categorySeed{
		{ID: "nc_ts_news_brief", Name: "新闻快讯", Slug: tushareNewsCategoryBriefSlug, Sort: 210, Visibility: "PUBLIC"},
		{ID: "nc_ts_news_major", Name: "新闻通讯", Slug: tushareNewsCategoryMajorSlug, Sort: 220, Visibility: "PUBLIC"},
		{ID: "nc_ts_report_research", Name: "券商研报", Slug: tushareNewsCategoryResearchSlug, Sort: 230, Visibility: "VIP"},
		{ID: "nc_ts_report_forecast", Name: "盈利预测", Slug: tushareNewsCategoryForecastSlug, Sort: 240, Visibility: "VIP"},
		{ID: "nc_ts_announcement", Name: "上市公司公告", Slug: tushareNewsCategoryAnnouncementSlug, Sort: 250, Visibility: "PUBLIC"},
	}
	now := time.Now()
	for _, seed := range seeds {
		_, err := r.db.Exec(`
	INSERT INTO news_categories
		(id, name, slug, sort, visibility, status, created_at, updated_at)
	SELECT
		?, ?, ?, ?, ?, 'PUBLISHED', ?, ?
	FROM DUAL
	WHERE NOT EXISTS (
		SELECT 1 FROM news_categories WHERE slug = ?
	)`,
			seed.ID,
			seed.Name,
			seed.Slug,
			seed.Sort,
			seed.Visibility,
			now,
			now,
			seed.Slug,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) resolveTushareNewsCategoryTargets(runtime tushareNewsSyncRuntimeConfig) (tushareNewsCategoryTargets, error) {
	rows, err := r.db.Query(`
SELECT id, slug, visibility
FROM news_categories
WHERE slug IN (?, ?, ?, ?, ?)`,
		tushareNewsCategoryBriefSlug,
		tushareNewsCategoryMajorSlug,
		tushareNewsCategoryResearchSlug,
		tushareNewsCategoryForecastSlug,
		tushareNewsCategoryAnnouncementSlug,
	)
	if err != nil {
		return tushareNewsCategoryTargets{}, err
	}
	defer rows.Close()

	bySlug := make(map[string]tushareNewsCategoryTarget)
	for rows.Next() {
		var (
			id         string
			slug       string
			visibility string
		)
		if err := rows.Scan(&id, &slug, &visibility); err != nil {
			return tushareNewsCategoryTargets{}, err
		}
		bySlug[strings.TrimSpace(slug)] = tushareNewsCategoryTarget{
			CategoryID: strings.TrimSpace(id),
			Visibility: normalizeTushareNewsVisibility(visibility),
		}
	}
	if err := rows.Err(); err != nil {
		return tushareNewsCategoryTargets{}, err
	}

	missing := make([]string, 0)
	requiredSlugs := []string{
		tushareNewsCategoryBriefSlug,
		tushareNewsCategoryMajorSlug,
		tushareNewsCategoryResearchSlug,
		tushareNewsCategoryForecastSlug,
		tushareNewsCategoryAnnouncementSlug,
	}
	for _, slug := range requiredSlugs {
		if _, ok := bySlug[slug]; !ok {
			missing = append(missing, slug)
		}
	}
	if len(missing) > 0 {
		return tushareNewsCategoryTargets{}, fmt.Errorf("missing tushare news categories: %s", strings.Join(missing, ","))
	}

	targets := tushareNewsCategoryTargets{
		Brief: tushareNewsCategoryTarget{
			CategoryID: bySlug[tushareNewsCategoryBriefSlug].CategoryID,
			Visibility: resolveTushareTargetVisibility(runtime.BriefVisibility, bySlug[tushareNewsCategoryBriefSlug].Visibility),
		},
		Major: tushareNewsCategoryTarget{
			CategoryID: bySlug[tushareNewsCategoryMajorSlug].CategoryID,
			Visibility: resolveTushareTargetVisibility(runtime.MajorVisibility, bySlug[tushareNewsCategoryMajorSlug].Visibility),
		},
		Research: tushareNewsCategoryTarget{
			CategoryID: bySlug[tushareNewsCategoryResearchSlug].CategoryID,
			Visibility: resolveTushareTargetVisibility(runtime.ResearchVisibility, bySlug[tushareNewsCategoryResearchSlug].Visibility),
		},
		Forecast: tushareNewsCategoryTarget{
			CategoryID: bySlug[tushareNewsCategoryForecastSlug].CategoryID,
			Visibility: resolveTushareTargetVisibility(runtime.ForecastVisibility, bySlug[tushareNewsCategoryForecastSlug].Visibility),
		},
		Announcement: tushareNewsCategoryTarget{
			CategoryID: bySlug[tushareNewsCategoryAnnouncementSlug].CategoryID,
			Visibility: resolveTushareTargetVisibility(runtime.AnnouncementVisibility, bySlug[tushareNewsCategoryAnnouncementSlug].Visibility),
		},
	}
	return targets, nil
}

func (r *MySQLGrowthRepo) resolveTushareTokenForNewsSync() (string, error) {
	item, err := r.getDataSourceBySourceKey("TUSHARE")
	if err == nil {
		if token := parseDataSourceStringConfig(item.Config, "token", "api_token", "tushare_token"); strings.TrimSpace(token) != "" {
			return strings.TrimSpace(token), nil
		}
	}
	if token := strings.TrimSpace(os.Getenv("TUSHARE_TOKEN")); token != "" {
		return token, nil
	}
	return "", errors.New("tushare token not configured")
}

func (r *MySQLGrowthRepo) upsertTushareNewsArticles(items []tushareNewsArticlePayload) (int, int, error) {
	if len(items) == 0 {
		return 0, 0, nil
	}
	syncedArticles := 0
	syncedAttachments := 0
	now := time.Now()
	for _, item := range items {
		id := strings.TrimSpace(item.ID)
		categoryID := strings.TrimSpace(item.CategoryID)
		title := truncateByRunes(normalizeUTF8Text(item.Title), 256)
		if id == "" || categoryID == "" || title == "" {
			continue
		}
		summary := truncateByRunes(normalizeUTF8Text(item.Summary), 512)
		if summary == "" {
			summary = title
		}
		content := normalizeUTF8Text(item.Content)
		if content == "" {
			content = summary
		}
		visibility := normalizeTushareNewsVisibility(item.Visibility)
		if visibility == "" {
			visibility = "PUBLIC"
		}
		publishedAt := item.PublishedAt
		if publishedAt.IsZero() {
			publishedAt = now
		}
		authorID := strings.TrimSpace(item.AuthorID)
		if authorID == "" {
			authorID = "admin_001"
		}

		_, err := r.db.Exec(`
INSERT INTO news_articles
	(id, category_id, title, summary, content, cover_url, tags, visibility, status, published_at, author_id, created_at, updated_at)
VALUES
	(?, ?, ?, ?, ?, NULL, NULL, ?, 'PUBLISHED', ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	category_id = VALUES(category_id),
	title = VALUES(title),
	summary = VALUES(summary),
	content = VALUES(content),
	visibility = VALUES(visibility),
	status = VALUES(status),
	published_at = VALUES(published_at),
	author_id = VALUES(author_id),
	updated_at = VALUES(updated_at)`,
			id,
			categoryID,
			title,
			summary,
			content,
			visibility,
			publishedAt,
			authorID,
			now,
			now,
		)
		if err != nil {
			return syncedArticles, syncedAttachments, err
		}
		syncedArticles++

		attachmentURL := strings.TrimSpace(item.AttachmentURL)
		if attachmentURL == "" {
			continue
		}
		attachmentID := buildTushareNewsAttachmentID("att", id+"|"+attachmentURL)
		fileName := inferTushareAttachmentFileName(attachmentURL, item.AttachmentName, title)
		mimeType := strings.TrimSpace(item.AttachmentMime)
		if mimeType == "" {
			mimeType = inferAttachmentMimeTypeFromFileName(fileName)
		}
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		_, err = r.db.Exec(`
INSERT INTO news_attachments
	(id, article_id, file_name, file_url, file_size, mime_type, created_at)
VALUES
	(?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	file_name = VALUES(file_name),
	file_url = VALUES(file_url),
	file_size = VALUES(file_size),
	mime_type = VALUES(mime_type)`,
			attachmentID,
			id,
			fileName,
			truncateByRunes(attachmentURL, 512),
			1,
			mimeType,
			now,
		)
		if err != nil {
			return syncedArticles, syncedAttachments, err
		}
		syncedAttachments++
	}
	return syncedArticles, syncedAttachments, nil
}

func fetchTushareNewsBriefArticles(token string, runtime tushareNewsSyncRuntimeConfig, target tushareNewsCategoryTarget, now time.Time, opts normalizedTushareSyncOptions) ([]tushareNewsArticlePayload, []model.NewsSyncRunDetail, error) {
	startAt := now.Add(-time.Duration(runtime.BriefLookbackHours) * time.Hour)
	client := &http.Client{Timeout: time.Duration(runtime.TimeoutMS) * time.Millisecond}
	limit := maxInt(50, minInt(runtime.BatchSize, 1000))
	items := make([]tushareNewsArticlePayload, 0, limit)
	details := make([]model.NewsSyncRunDetail, 0, maxInt(1, len(runtime.BriefSources)))
	seen := make(map[string]struct{})
	errs := make([]string, 0)

	sources := filterTushareSources(runtime.BriefSources, opts.SourceSet)
	if len(sources) == 0 {
		return items, details, nil
	}
	for _, source := range sources {
		source = strings.ToLower(strings.TrimSpace(source))
		if source == "" {
			continue
		}
		detail := newTushareSyncDetailDraft(tushareSyncTypeNewsBrief, source, "")
		callStartedAt := time.Now()
		params := map[string]string{
			"src":        source,
			"start_date": startAt.Format("2006-01-02 15:04:05"),
			"end_date":   now.Format("2006-01-02 15:04:05"),
			"limit":      strconv.Itoa(limit),
		}
		parsed, err := callTushareAPI(client, token, "news", params, "datetime,title,content,channels")
		if err != nil && shouldRetryTushareWithoutFields(err) {
			parsed, err = callTushareAPI(client, token, "news", params, "")
		}
		if err != nil {
			rateLimited := isTushareRateLimitErrorText(err.Error())
			detail.Status = "FAILED"
			detail.FailedCount = 1
			detail.ErrorText = truncateByRunes(err.Error(), 512)
			if rateLimited {
				detail.WarningText = "触发Tushare限频，已停止后续来源抓取"
			}
			detail.StartedAt = callStartedAt.Format(time.RFC3339)
			detail.FinishedAt = time.Now().Format(time.RFC3339)
			details = append(details, detail)
			errs = append(errs, fmt.Sprintf("%s: %v", source, err))
			if rateLimited {
				break
			}
			continue
		}
		beforeCount := len(items)
		fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
		for _, row := range parsed.Data.Items {
			datetimeText := firstTushareString(row, fieldIndex, "datetime")
			contentText := firstTushareString(row, fieldIndex, "content")
			title := firstTushareString(row, fieldIndex, "title")
			channels := firstTushareString(row, fieldIndex, "channels")
			if title == "" {
				title = truncateByRunes(normalizeUTF8Text(contentText), 80)
			}
			title = normalizeUTF8Text(title)
			contentText = normalizeUTF8Text(contentText)
			if title == "" || contentText == "" {
				continue
			}
			publishedAt := parseTushareDateTimeOr(datetimeText, now)
			uniqueKey := fmt.Sprintf("brief|%s|%s|%s|%s", source, publishedAt.Format(time.RFC3339), title, channels)
			id := buildTushareNewsArticleID("nb", uniqueKey)
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			if channels != "" {
				contentText = fmt.Sprintf("频道：%s\n\n%s", channels, contentText)
			}
			items = append(items, tushareNewsArticlePayload{
				ID:          id,
				CategoryID:  target.CategoryID,
				Visibility:  target.Visibility,
				SyncType:    tushareSyncTypeNewsBrief,
				Source:      source,
				Title:       title,
				Summary:     buildTushareNewsSummary(contentText, title),
				Content:     contentText,
				PublishedAt: publishedAt,
				AuthorID:    runtime.AuthorID,
			})
		}
		fetchedCount := len(items) - beforeCount
		detail.Status = "SUCCESS"
		detail.FetchedCount = fetchedCount
		detail.StartedAt = callStartedAt.Format(time.RFC3339)
		detail.FinishedAt = time.Now().Format(time.RFC3339)
		details = append(details, detail)
	}
	if len(items) == 0 && len(errs) > 0 {
		return nil, details, errors.New(strings.Join(compactErrorMessages(errs, 3), " | "))
	}
	return items, details, nil
}

func fetchTushareMajorNewsArticles(token string, runtime tushareNewsSyncRuntimeConfig, target tushareNewsCategoryTarget, now time.Time, opts normalizedTushareSyncOptions) ([]tushareNewsArticlePayload, []model.NewsSyncRunDetail, error) {
	startAt := now.Add(-time.Duration(runtime.MajorLookbackHours) * time.Hour)
	client := &http.Client{Timeout: time.Duration(runtime.TimeoutMS) * time.Millisecond}
	limit := maxInt(50, minInt(runtime.BatchSize, 1000))
	items := make([]tushareNewsArticlePayload, 0, limit)
	details := make([]model.NewsSyncRunDetail, 0, maxInt(1, len(runtime.MajorSources)))
	seen := make(map[string]struct{})
	errs := make([]string, 0)

	sources := runtime.MajorSources
	if len(sources) == 0 {
		sources = []string{""}
	}
	sources = filterTushareSources(sources, opts.SourceSet)
	if len(sources) == 0 {
		return items, details, nil
	}

	for _, source := range sources {
		source = strings.TrimSpace(source)
		detailSource := source
		if detailSource == "" {
			detailSource = "all"
		}
		detail := newTushareSyncDetailDraft(tushareSyncTypeNewsMajor, detailSource, "")
		callStartedAt := time.Now()
		params := map[string]string{
			"start_date": startAt.Format("2006-01-02 15:04:05"),
			"end_date":   now.Format("2006-01-02 15:04:05"),
			"limit":      strconv.Itoa(limit),
		}
		if source != "" {
			params["src"] = source
		}
		parsed, err := callTushareAPI(client, token, "major_news", params, "pub_time,src,title,content")
		if err != nil && shouldRetryTushareWithoutFields(err) {
			parsed, err = callTushareAPI(client, token, "major_news", params, "")
		}
		if err != nil {
			rateLimited := isTushareRateLimitErrorText(err.Error())
			detail.Status = "FAILED"
			detail.FailedCount = 1
			detail.ErrorText = truncateByRunes(err.Error(), 512)
			if rateLimited {
				detail.WarningText = "触发Tushare限频，已停止后续来源抓取"
			}
			detail.StartedAt = callStartedAt.Format(time.RFC3339)
			detail.FinishedAt = time.Now().Format(time.RFC3339)
			details = append(details, detail)
			errs = append(errs, fmt.Sprintf("%s: %v", detailSource, err))
			if rateLimited {
				break
			}
			continue
		}
		beforeCount := len(items)
		fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
		for _, row := range parsed.Data.Items {
			pubTimeText := firstTushareString(row, fieldIndex, "pub_time")
			rowSource := firstTushareString(row, fieldIndex, "src")
			title := firstTushareString(row, fieldIndex, "title")
			contentText := firstTushareString(row, fieldIndex, "content")
			title = normalizeUTF8Text(title)
			contentText = normalizeUTF8Text(contentText)
			if title == "" {
				title = truncateByRunes(contentText, 80)
			}
			if title == "" || contentText == "" {
				continue
			}
			publishedAt := parseTushareDateTimeOr(pubTimeText, now)
			uniqueKey := fmt.Sprintf("major|%s|%s|%s", rowSource, publishedAt.Format(time.RFC3339), title)
			id := buildTushareNewsArticleID("nm", uniqueKey)
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			resolvedSource := normalizeUTF8Text(rowSource)
			if resolvedSource == "" {
				resolvedSource = detailSource
			}
			if rowSource != "" {
				contentText = fmt.Sprintf("来源：%s\n\n%s", rowSource, contentText)
			}
			items = append(items, tushareNewsArticlePayload{
				ID:          id,
				CategoryID:  target.CategoryID,
				Visibility:  target.Visibility,
				SyncType:    tushareSyncTypeNewsMajor,
				Source:      resolvedSource,
				Title:       title,
				Summary:     buildTushareNewsSummary(contentText, title),
				Content:     contentText,
				PublishedAt: publishedAt,
				AuthorID:    runtime.AuthorID,
			})
		}
		fetchedCount := len(items) - beforeCount
		detail.Status = "SUCCESS"
		detail.FetchedCount = fetchedCount
		detail.StartedAt = callStartedAt.Format(time.RFC3339)
		detail.FinishedAt = time.Now().Format(time.RFC3339)
		details = append(details, detail)
	}
	if len(items) == 0 && len(errs) > 0 {
		return nil, details, errors.New(strings.Join(compactErrorMessages(errs, 3), " | "))
	}
	return items, details, nil
}

func fetchTushareResearchReportArticles(token string, runtime tushareNewsSyncRuntimeConfig, target tushareNewsCategoryTarget, now time.Time, opts normalizedTushareSyncOptions) ([]tushareNewsArticlePayload, []model.NewsSyncRunDetail, error) {
	startDate := now.AddDate(0, 0, -runtime.ResearchLookbackDays).Format("20060102")
	endDate := now.Format("20060102")
	client := &http.Client{Timeout: time.Duration(runtime.TimeoutMS) * time.Millisecond}
	callStartedAt := time.Now()
	parsed, err := callTushareAPI(client, token, "research_report", map[string]string{
		"start_date": startDate,
		"end_date":   endDate,
	}, "ts_code,name,report_date,title,report_type,classify,org_name,author_name,url,summary")
	if err != nil && shouldRetryTushareWithoutFields(err) {
		parsed, err = callTushareAPI(client, token, "research_report", map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
		}, "")
	}
	if err != nil {
		detail := newTushareSyncDetailDraft(tushareSyncTypeResearch, "research_report", "")
		detail.Status = "FAILED"
		detail.FailedCount = 1
		detail.ErrorText = truncateByRunes(err.Error(), 512)
		detail.StartedAt = callStartedAt.Format(time.RFC3339)
		detail.FinishedAt = time.Now().Format(time.RFC3339)
		return nil, []model.NewsSyncRunDetail{detail}, err
	}
	fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
	items := make([]tushareNewsArticlePayload, 0, len(parsed.Data.Items))
	groupCount := make(map[string]int)
	groupDetail := make(map[string]model.NewsSyncRunDetail)
	seen := make(map[string]struct{})
	for _, row := range parsed.Data.Items {
		tsCode := firstTushareString(row, fieldIndex, "ts_code")
		name := firstTushareString(row, fieldIndex, "name")
		reportDateText := firstTushareString(row, fieldIndex, "report_date")
		title := firstTushareString(row, fieldIndex, "title")
		reportType := firstTushareString(row, fieldIndex, "report_type")
		classify := firstTushareString(row, fieldIndex, "classify")
		orgName := firstTushareString(row, fieldIndex, "org_name", "org")
		authorName := firstTushareString(row, fieldIndex, "author_name", "author")
		sourceURL := firstTushareString(row, fieldIndex, "url")
		summary := firstTushareString(row, fieldIndex, "summary", "abstr")
		resolvedSource := normalizeUTF8Text(orgName)
		resolvedSymbol := normalizeTushareSymbol(tsCode)
		if !matchTushareSourceFilter(opts.SourceSet, resolvedSource) {
			continue
		}
		if !matchTushareSymbolFilter(opts.SymbolSet, resolvedSymbol) {
			continue
		}
		title = normalizeUTF8Text(title)
		if title == "" {
			fallbackName := strings.TrimSpace(name)
			if fallbackName == "" {
				fallbackName = strings.TrimSpace(tsCode)
			}
			title = strings.TrimSpace(fallbackName + " 券商研报")
		}
		if title == "" {
			continue
		}
		publishedAt := parseTushareDateTimeOr(reportDateText, now)
		uniqueKey := fmt.Sprintf("research|%s|%s|%s|%s|%s", reportDateText, tsCode, title, orgName, sourceURL)
		id := buildTushareNewsArticleID("rr", uniqueKey)
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		detailKey := buildNewsSyncDetailKey(tushareSyncTypeResearch, resolvedSource, resolvedSymbol)
		groupCount[detailKey]++
		if _, exists := groupDetail[detailKey]; !exists {
			groupDetail[detailKey] = newTushareSyncDetailDraft(tushareSyncTypeResearch, resolvedSource, resolvedSymbol)
		}
		contentBuilder := make([]string, 0, 8)
		contentBuilder = append(contentBuilder, fmt.Sprintf("股票：%s %s", strings.TrimSpace(tsCode), strings.TrimSpace(name)))
		contentBuilder = append(contentBuilder, fmt.Sprintf("机构：%s", strings.TrimSpace(orgName)))
		contentBuilder = append(contentBuilder, fmt.Sprintf("作者：%s", strings.TrimSpace(authorName)))
		contentBuilder = append(contentBuilder, fmt.Sprintf("报告类型：%s", strings.TrimSpace(reportType)))
		contentBuilder = append(contentBuilder, fmt.Sprintf("分类：%s", strings.TrimSpace(classify)))
		if strings.TrimSpace(summary) != "" {
			contentBuilder = append(contentBuilder, "", "摘要：", strings.TrimSpace(summary))
		}
		if strings.TrimSpace(sourceURL) != "" {
			contentBuilder = append(contentBuilder, "", "原文链接：", strings.TrimSpace(sourceURL))
		}
		contentText := strings.Join(contentBuilder, "\n")
		items = append(items, tushareNewsArticlePayload{
			ID:             id,
			CategoryID:     target.CategoryID,
			Visibility:     target.Visibility,
			SyncType:       tushareSyncTypeResearch,
			Source:         resolvedSource,
			Symbol:         resolvedSymbol,
			Title:          title,
			Summary:        buildTushareNewsSummary(summary, title),
			Content:        contentText,
			PublishedAt:    publishedAt,
			AuthorID:       runtime.AuthorID,
			AttachmentURL:  strings.TrimSpace(sourceURL),
			AttachmentName: strings.TrimSpace(title) + ".pdf",
			AttachmentMime: "application/pdf",
		})
	}
	finishedAt := time.Now()
	details := make([]model.NewsSyncRunDetail, 0, len(groupDetail))
	for detailKey, draft := range groupDetail {
		draft.Status = "SUCCESS"
		draft.FetchedCount = groupCount[detailKey]
		draft.StartedAt = callStartedAt.Format(time.RFC3339)
		draft.FinishedAt = finishedAt.Format(time.RFC3339)
		details = append(details, draft)
	}
	return items, details, nil
}

func fetchTushareForecastArticles(token string, runtime tushareNewsSyncRuntimeConfig, target tushareNewsCategoryTarget, now time.Time, opts normalizedTushareSyncOptions) ([]tushareNewsArticlePayload, []model.NewsSyncRunDetail, error) {
	startDate := now.AddDate(0, 0, -runtime.ForecastLookbackDays).Format("20060102")
	endDate := now.Format("20060102")
	client := &http.Client{Timeout: time.Duration(runtime.TimeoutMS) * time.Millisecond}
	callStartedAt := time.Now()
	parsed, err := callTushareAPI(client, token, "report_rc", map[string]string{
		"start_date": startDate,
		"end_date":   endDate,
	}, "report_date,ts_code,name,report_title,org_name,author_name,quarter,op_rt,eps,pe,max_price,min_price,create_time")
	if err != nil && shouldRetryTushareWithoutFields(err) {
		parsed, err = callTushareAPI(client, token, "report_rc", map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
		}, "")
	}
	if err != nil {
		detail := newTushareSyncDetailDraft(tushareSyncTypeForecast, "report_rc", "")
		detail.Status = "FAILED"
		detail.FailedCount = 1
		detail.ErrorText = truncateByRunes(err.Error(), 512)
		detail.StartedAt = callStartedAt.Format(time.RFC3339)
		detail.FinishedAt = time.Now().Format(time.RFC3339)
		return nil, []model.NewsSyncRunDetail{detail}, err
	}
	fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
	items := make([]tushareNewsArticlePayload, 0, len(parsed.Data.Items))
	groupCount := make(map[string]int)
	groupDetail := make(map[string]model.NewsSyncRunDetail)
	seen := make(map[string]struct{})
	for _, row := range parsed.Data.Items {
		reportDateText := firstTushareString(row, fieldIndex, "report_date")
		tsCode := firstTushareString(row, fieldIndex, "ts_code")
		name := firstTushareString(row, fieldIndex, "name")
		reportTitle := firstTushareString(row, fieldIndex, "report_title", "title")
		orgName := firstTushareString(row, fieldIndex, "org_name")
		authorName := firstTushareString(row, fieldIndex, "author_name")
		quarter := firstTushareString(row, fieldIndex, "quarter")
		opRT := firstTushareString(row, fieldIndex, "op_rt")
		eps := firstTushareString(row, fieldIndex, "eps")
		pe := firstTushareString(row, fieldIndex, "pe")
		maxPrice := firstTushareString(row, fieldIndex, "max_price")
		minPrice := firstTushareString(row, fieldIndex, "min_price")
		createTime := firstTushareString(row, fieldIndex, "create_time")
		resolvedSource := normalizeUTF8Text(orgName)
		resolvedSymbol := normalizeTushareSymbol(tsCode)
		if !matchTushareSourceFilter(opts.SourceSet, resolvedSource) {
			continue
		}
		if !matchTushareSymbolFilter(opts.SymbolSet, resolvedSymbol) {
			continue
		}

		reportTitle = normalizeUTF8Text(reportTitle)
		if reportTitle == "" {
			fallbackName := strings.TrimSpace(name)
			if fallbackName == "" {
				fallbackName = strings.TrimSpace(tsCode)
			}
			reportTitle = strings.TrimSpace(fallbackName + " 盈利预测")
		}
		if reportTitle == "" {
			continue
		}
		publishedAt := parseTushareDateTimeOr(createTime, parseTushareDateTimeOr(reportDateText, now))
		uniqueKey := fmt.Sprintf("forecast|%s|%s|%s|%s|%s", reportDateText, tsCode, orgName, authorName, reportTitle)
		id := buildTushareNewsArticleID("rc", uniqueKey)
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		detailKey := buildNewsSyncDetailKey(tushareSyncTypeForecast, resolvedSource, resolvedSymbol)
		groupCount[detailKey]++
		if _, exists := groupDetail[detailKey]; !exists {
			groupDetail[detailKey] = newTushareSyncDetailDraft(tushareSyncTypeForecast, resolvedSource, resolvedSymbol)
		}
		summaryParts := make([]string, 0, 4)
		if strings.TrimSpace(orgName) != "" {
			summaryParts = append(summaryParts, strings.TrimSpace(orgName))
		}
		if strings.TrimSpace(opRT) != "" {
			summaryParts = append(summaryParts, "评级:"+strings.TrimSpace(opRT))
		}
		if strings.TrimSpace(eps) != "" {
			summaryParts = append(summaryParts, "EPS:"+strings.TrimSpace(eps))
		}
		if strings.TrimSpace(pe) != "" {
			summaryParts = append(summaryParts, "PE:"+strings.TrimSpace(pe))
		}
		summary := strings.Join(summaryParts, " ")
		contentBuilder := []string{
			fmt.Sprintf("股票：%s %s", strings.TrimSpace(tsCode), strings.TrimSpace(name)),
			fmt.Sprintf("机构：%s", strings.TrimSpace(orgName)),
			fmt.Sprintf("作者：%s", strings.TrimSpace(authorName)),
			fmt.Sprintf("季度：%s", strings.TrimSpace(quarter)),
			fmt.Sprintf("评级：%s", strings.TrimSpace(opRT)),
			fmt.Sprintf("EPS：%s", strings.TrimSpace(eps)),
			fmt.Sprintf("PE：%s", strings.TrimSpace(pe)),
			fmt.Sprintf("目标价区间：%s - %s", strings.TrimSpace(minPrice), strings.TrimSpace(maxPrice)),
		}
		contentText := strings.Join(contentBuilder, "\n")
		items = append(items, tushareNewsArticlePayload{
			ID:          id,
			CategoryID:  target.CategoryID,
			Visibility:  target.Visibility,
			SyncType:    tushareSyncTypeForecast,
			Source:      resolvedSource,
			Symbol:      resolvedSymbol,
			Title:       reportTitle,
			Summary:     buildTushareNewsSummary(summary, reportTitle),
			Content:     contentText,
			PublishedAt: publishedAt,
			AuthorID:    runtime.AuthorID,
		})
	}
	finishedAt := time.Now()
	details := make([]model.NewsSyncRunDetail, 0, len(groupDetail))
	for detailKey, draft := range groupDetail {
		draft.Status = "SUCCESS"
		draft.FetchedCount = groupCount[detailKey]
		draft.StartedAt = callStartedAt.Format(time.RFC3339)
		draft.FinishedAt = finishedAt.Format(time.RFC3339)
		details = append(details, draft)
	}
	return items, details, nil
}

func fetchTushareAnnouncementArticles(token string, runtime tushareNewsSyncRuntimeConfig, target tushareNewsCategoryTarget, now time.Time, opts normalizedTushareSyncOptions) ([]tushareNewsArticlePayload, []model.NewsSyncRunDetail, error) {
	startDate := now.AddDate(0, 0, -runtime.AnnouncementLookbackDays).Format("20060102")
	endDate := now.Format("20060102")
	client := &http.Client{Timeout: time.Duration(runtime.TimeoutMS) * time.Millisecond}
	callStartedAt := time.Now()
	parsed, err := callTushareAPI(client, token, "anns_d", map[string]string{
		"start_date": startDate,
		"end_date":   endDate,
		"limit":      strconv.Itoa(maxInt(50, minInt(runtime.BatchSize, 2000))),
	}, "ann_date,ts_code,name,title,url,rec_time")
	if err != nil && shouldRetryTushareWithoutFields(err) {
		parsed, err = callTushareAPI(client, token, "anns_d", map[string]string{
			"start_date": startDate,
			"end_date":   endDate,
			"limit":      strconv.Itoa(maxInt(50, minInt(runtime.BatchSize, 2000))),
		}, "")
	}
	if err != nil {
		detail := newTushareSyncDetailDraft(tushareSyncTypeAnnouncement, "anns_d", "")
		detail.Status = "FAILED"
		detail.FailedCount = 1
		detail.ErrorText = truncateByRunes(err.Error(), 512)
		detail.StartedAt = callStartedAt.Format(time.RFC3339)
		detail.FinishedAt = time.Now().Format(time.RFC3339)
		return nil, []model.NewsSyncRunDetail{detail}, err
	}
	fieldIndex := buildTushareFieldIndex(parsed.Data.Fields)
	items := make([]tushareNewsArticlePayload, 0, len(parsed.Data.Items))
	groupCount := make(map[string]int)
	groupDetail := make(map[string]model.NewsSyncRunDetail)
	seen := make(map[string]struct{})
	for _, row := range parsed.Data.Items {
		annDate := firstTushareString(row, fieldIndex, "ann_date")
		tsCode := firstTushareString(row, fieldIndex, "ts_code")
		name := firstTushareString(row, fieldIndex, "name")
		title := firstTushareString(row, fieldIndex, "title")
		sourceURL := firstTushareString(row, fieldIndex, "url")
		recTime := firstTushareString(row, fieldIndex, "rec_time")
		resolvedSource := "交易所公告"
		resolvedSymbol := normalizeTushareSymbol(tsCode)
		if !matchTushareSourceFilter(opts.SourceSet, resolvedSource) {
			continue
		}
		if !matchTushareSymbolFilter(opts.SymbolSet, resolvedSymbol) {
			continue
		}
		title = normalizeUTF8Text(title)
		if title == "" {
			title = strings.TrimSpace(tsCode + " 上市公司公告")
		}
		if title == "" {
			continue
		}
		publishedAt := parseTushareDateTimeOr(recTime, parseTushareDateTimeOr(annDate, now))
		uniqueKey := fmt.Sprintf("announcement|%s|%s|%s|%s", annDate, tsCode, title, sourceURL)
		id := buildTushareNewsArticleID("an", uniqueKey)
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		detailKey := buildNewsSyncDetailKey(tushareSyncTypeAnnouncement, resolvedSource, resolvedSymbol)
		groupCount[detailKey]++
		if _, exists := groupDetail[detailKey]; !exists {
			groupDetail[detailKey] = newTushareSyncDetailDraft(tushareSyncTypeAnnouncement, resolvedSource, resolvedSymbol)
		}
		contentBuilder := []string{
			fmt.Sprintf("股票：%s %s", strings.TrimSpace(tsCode), strings.TrimSpace(name)),
			fmt.Sprintf("公告日期：%s", strings.TrimSpace(annDate)),
			fmt.Sprintf("公告标题：%s", title),
		}
		if strings.TrimSpace(sourceURL) != "" {
			contentBuilder = append(contentBuilder, "公告链接："+strings.TrimSpace(sourceURL))
		}
		contentText := strings.Join(contentBuilder, "\n")
		items = append(items, tushareNewsArticlePayload{
			ID:             id,
			CategoryID:     target.CategoryID,
			Visibility:     target.Visibility,
			SyncType:       tushareSyncTypeAnnouncement,
			Source:         resolvedSource,
			Symbol:         resolvedSymbol,
			Title:          title,
			Summary:        buildTushareNewsSummary(contentText, title),
			Content:        contentText,
			PublishedAt:    publishedAt,
			AuthorID:       runtime.AuthorID,
			AttachmentURL:  strings.TrimSpace(sourceURL),
			AttachmentName: title + ".html",
			AttachmentMime: "text/html",
		})
	}
	finishedAt := time.Now()
	details := make([]model.NewsSyncRunDetail, 0, len(groupDetail))
	for detailKey, draft := range groupDetail {
		draft.Status = "SUCCESS"
		draft.FetchedCount = groupCount[detailKey]
		draft.StartedAt = callStartedAt.Format(time.RFC3339)
		draft.FinishedAt = finishedAt.Format(time.RFC3339)
		details = append(details, draft)
	}
	return items, details, nil
}

func buildTushareFieldIndex(fields []string) map[string]int {
	fieldIndex := make(map[string]int, len(fields))
	for index, field := range fields {
		key := strings.ToLower(strings.TrimSpace(field))
		if key == "" {
			continue
		}
		fieldIndex[key] = index
	}
	return fieldIndex
}

func firstTushareString(row []interface{}, fieldIndex map[string]int, keys ...string) string {
	for _, key := range keys {
		lookupKey := strings.ToLower(strings.TrimSpace(key))
		index, ok := fieldIndex[lookupKey]
		if !ok || index < 0 || index >= len(row) {
			continue
		}
		text := strings.TrimSpace(fmt.Sprintf("%v", row[index]))
		if text == "" || text == "<nil>" {
			continue
		}
		return text
	}
	return ""
}

func parseTushareDateTimeOr(raw string, fallback time.Time) time.Time {
	text := strings.TrimSpace(raw)
	if text == "" {
		return fallback
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04:05",
		"20060102150405",
		"2006-01-02",
		"20060102",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, text, time.Local); err == nil {
			return parsed
		}
	}
	return fallback
}

func parseDateTimeOrDefault(raw string, fallback time.Time) time.Time {
	return parseTushareDateTimeOr(raw, fallback)
}

func buildTushareNewsSummary(content string, title string) string {
	text := normalizeUTF8Text(content)
	if text == "" {
		return truncateByRunes(normalizeUTF8Text(title), 200)
	}
	return truncateByRunes(text, 200)
}

func splitTushareSourceList(raw string, defaults []string) []string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return append([]string(nil), defaults...)
	}
	normalized := strings.NewReplacer("，", ",", "；", ",", ";", ",", "\n", ",", "\t", ",").Replace(text)
	parts := strings.Split(normalized, ",")
	seen := make(map[string]struct{})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		lookupKey := strings.ToLower(item)
		if _, exists := seen[lookupKey]; exists {
			continue
		}
		seen[lookupKey] = struct{}{}
		result = append(result, item)
	}
	if len(result) == 0 {
		return append([]string(nil), defaults...)
	}
	return result
}

func splitTushareSyncTypeList(raw string, defaults []string) []string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return append([]string(nil), defaults...)
	}
	normalized := strings.NewReplacer("，", ",", "；", ",", ";", ",", "\n", ",", "\t", ",").Replace(text)
	parts := strings.Split(normalized, ",")
	seen := make(map[string]struct{})
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		syncType := normalizeTushareSyncType(part)
		if syncType == "" {
			continue
		}
		if _, exists := seen[syncType]; exists {
			continue
		}
		seen[syncType] = struct{}{}
		result = append(result, syncType)
	}
	if len(result) == 0 {
		return append([]string(nil), defaults...)
	}
	return result
}

func normalizeTushareSyncOptions(opts model.TushareNewsSyncOptions) normalizedTushareSyncOptions {
	normalized := normalizedTushareSyncOptions{
		BatchSize:   opts.BatchSize,
		SourceSet:   make(map[string]struct{}),
		SymbolSet:   make(map[string]struct{}),
		SyncTypeSet: make(map[string]struct{}),
	}
	if normalized.BatchSize < 0 {
		normalized.BatchSize = 0
	}
	for _, source := range opts.Sources {
		key := normalizeTushareSourceKey(source)
		if key == "" {
			continue
		}
		normalized.SourceSet[key] = struct{}{}
	}
	for _, symbol := range opts.Symbols {
		key := normalizeTushareSymbol(symbol)
		if key == "" {
			continue
		}
		normalized.SymbolSet[key] = struct{}{}
	}
	for _, syncType := range opts.SyncTypes {
		key := normalizeTushareSyncType(syncType)
		if key == "" {
			continue
		}
		normalized.SyncTypeSet[key] = struct{}{}
	}
	return normalized
}

func normalizeTushareSyncType(raw string) string {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	switch normalized {
	case "NEWS", "NEWS_BRIEF", "BRIEF":
		return tushareSyncTypeNewsBrief
	case "MAJOR", "NEWS_MAJOR", "MAJOR_NEWS":
		return tushareSyncTypeNewsMajor
	case "RESEARCH", "RESEARCH_REPORT":
		return tushareSyncTypeResearch
	case "FORECAST", "REPORT_RC":
		return tushareSyncTypeForecast
	case "ANN", "ANNS", "ANNOUNCEMENT", "ANNS_D":
		return tushareSyncTypeAnnouncement
	default:
		return ""
	}
}

func shouldSyncTushareType(syncTypeSet map[string]struct{}, targetType string) bool {
	if len(syncTypeSet) == 0 {
		return true
	}
	_, ok := syncTypeSet[targetType]
	return ok
}

func filterTushareSources(allSources []string, sourceSet map[string]struct{}) []string {
	if len(allSources) == 0 {
		return nil
	}
	if len(sourceSet) == 0 {
		return append([]string(nil), allSources...)
	}
	result := make([]string, 0, len(allSources))
	seen := make(map[string]struct{})
	for _, source := range allSources {
		normalized := normalizeTushareSourceKey(source)
		if normalized == "" {
			continue
		}
		if _, ok := sourceSet[normalized]; !ok {
			continue
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, source)
	}
	return result
}

func normalizeTushareSourceKey(raw string) string {
	return strings.ToLower(normalizeUTF8Text(raw))
}

func normalizeTushareSymbol(raw string) string {
	return strings.ToUpper(strings.TrimSpace(raw))
}

func matchTushareSourceFilter(sourceSet map[string]struct{}, source string) bool {
	if len(sourceSet) == 0 {
		return true
	}
	_, ok := sourceSet[normalizeTushareSourceKey(source)]
	return ok
}

func matchTushareSymbolFilter(symbolSet map[string]struct{}, symbol string) bool {
	if len(symbolSet) == 0 {
		return true
	}
	_, ok := symbolSet[normalizeTushareSymbol(symbol)]
	return ok
}

func newTushareSyncDetailDraft(syncType string, source string, symbol string) model.NewsSyncRunDetail {
	return model.NewsSyncRunDetail{
		JobName:   tushareNewsIncrementalJobName,
		SyncType:  strings.ToUpper(strings.TrimSpace(syncType)),
		Source:    truncateByRunes(normalizeUTF8Text(source), 128),
		Symbol:    truncateByRunes(normalizeTushareSymbol(symbol), 32),
		Status:    "SUCCESS",
		StartedAt: time.Now().Format(time.RFC3339),
	}
}

func buildNewsSyncDetailKey(syncType string, source string, symbol string) string {
	return strings.ToUpper(strings.TrimSpace(syncType)) + "|" +
		normalizeTushareSourceKey(source) + "|" +
		normalizeTushareSymbol(symbol)
}

func applyTushareDetailUpsertCounts(details []model.NewsSyncRunDetail, items []tushareNewsArticlePayload) {
	if len(details) == 0 || len(items) == 0 {
		return
	}
	groupCounts := make(map[string]int, len(items))
	for _, item := range items {
		key := buildNewsSyncDetailKey(item.SyncType, item.Source, item.Symbol)
		groupCounts[key]++
	}
	for idx := range details {
		key := buildNewsSyncDetailKey(details[idx].SyncType, details[idx].Source, details[idx].Symbol)
		if strings.EqualFold(details[idx].Status, "FAILED") {
			if details[idx].FailedCount <= 0 {
				details[idx].FailedCount = 1
			}
			continue
		}
		if upserted, ok := groupCounts[key]; ok {
			details[idx].UpsertedCount = upserted
			if details[idx].FetchedCount <= 0 {
				details[idx].FetchedCount = upserted
			}
		}
		if details[idx].FailedCount < 0 {
			details[idx].FailedCount = 0
		}
		if details[idx].Status == "" {
			details[idx].Status = "SUCCESS"
		}
	}
}

func normalizeTushareNewsVisibility(raw string) string {
	value := strings.ToUpper(strings.TrimSpace(raw))
	switch value {
	case "PUBLIC", "VIP":
		return value
	default:
		return ""
	}
}

func resolveTushareTargetVisibility(preferred string, fallback string) string {
	if value := normalizeTushareNewsVisibility(preferred); value != "" {
		return value
	}
	if value := normalizeTushareNewsVisibility(fallback); value != "" {
		return value
	}
	return "PUBLIC"
}

func buildTushareNewsArticleID(kind string, uniqueKey string) string {
	return buildTushareStableID("na_ts_"+strings.ToLower(strings.TrimSpace(kind)), uniqueKey, 32)
}

func buildTushareNewsAttachmentID(kind string, uniqueKey string) string {
	return buildTushareStableID("nat_ts_"+strings.ToLower(strings.TrimSpace(kind)), uniqueKey, 32)
}

func buildTushareStableID(prefix string, uniqueKey string, maxLen int) string {
	normalizedPrefix := strings.ToLower(strings.TrimSpace(prefix))
	if normalizedPrefix == "" {
		normalizedPrefix = "ts"
	}
	if maxLen <= 0 {
		maxLen = 32
	}
	if len(normalizedPrefix) >= maxLen-1 {
		normalizedPrefix = normalizedPrefix[:maxLen-1]
	}
	hash := sha1.Sum([]byte(strings.TrimSpace(uniqueKey)))
	hashText := hex.EncodeToString(hash[:])
	available := maxLen - len(normalizedPrefix) - 1
	if available <= 0 {
		return normalizedPrefix
	}
	if available > len(hashText) {
		available = len(hashText)
	}
	return normalizedPrefix + "_" + hashText[:available]
}

func inferTushareAttachmentFileName(fileURL string, preferredName string, title string) string {
	if preferred := strings.TrimSpace(preferredName); preferred != "" {
		return truncateByRunes(preferred, 256)
	}
	baseName := strings.TrimSpace(path.Base(strings.TrimSpace(fileURL)))
	if baseName != "" && baseName != "." && baseName != "/" {
		return truncateByRunes(baseName, 256)
	}
	cleanTitle := normalizeUTF8Text(title)
	if cleanTitle == "" {
		cleanTitle = "attachment"
	}
	return truncateByRunes(cleanTitle, 256)
}

func inferAttachmentMimeTypeFromFileName(fileName string) string {
	ext := strings.ToLower(strings.TrimSpace(path.Ext(fileName)))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".ppt":
		return "application/vnd.ms-powerpoint"
	case ".pptx":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".zip":
		return "application/zip"
	default:
		return ""
	}
}

func isTushareRateLimitErrorText(message string) bool {
	text := strings.ToLower(normalizeUTF8Text(message))
	if text == "" {
		return false
	}
	keywords := []string{
		"最多访问该接口",
		"触发访问频次",
		"请求过于频繁",
		"rate limit",
		"too many requests",
	}
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	if strings.Contains(text, "每分钟") && strings.Contains(text, "访问") {
		return true
	}
	if strings.Contains(text, "每小时") && strings.Contains(text, "访问") {
		return true
	}
	if strings.Contains(text, "每天") && strings.Contains(text, "访问") {
		return true
	}
	return false
}

func shouldRetryTushareWithoutFields(err error) bool {
	if err == nil {
		return false
	}
	return !isTushareRateLimitErrorText(err.Error())
}

func onlyRateLimitMessages(messages []string) bool {
	hasMessage := false
	for _, message := range messages {
		text := strings.TrimSpace(message)
		if text == "" {
			continue
		}
		hasMessage = true
		if !isTushareRateLimitErrorText(text) {
			return false
		}
	}
	return hasMessage
}

func compactErrorMessages(messages []string, limit int) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(messages))
	for _, message := range messages {
		text := truncateByRunes(normalizeUTF8Text(message), 160)
		if text == "" {
			continue
		}
		if _, exists := seen[text]; exists {
			continue
		}
		seen[text] = struct{}{}
		result = append(result, text)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

func (r *MySQLGrowthRepo) ListStockRecommendations(userID string, tradeDate string, page int, pageSize int) ([]model.StockRecommendation, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE status IN ('PUBLISHED', 'ACTIVE', 'TRACKING')"
	if tradeDate != "" {
		filter += " AND DATE(valid_from) = ?"
		args = append(args, tradeDate)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM stock_recommendations"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary
FROM stock_recommendations` + filter + `
ORDER BY valid_from DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.StockRecommendation, 0)
	for rows.Next() {
		var item model.StockRecommendation
		var positionRange, reasonSummary sql.NullString
		var validFrom, validTo time.Time
		if err := rows.Scan(&item.ID, &item.Symbol, &item.Name, &item.Score, &item.RiskLevel, &positionRange, &validFrom, &validTo, &item.Status, &reasonSummary); err != nil {
			return nil, 0, err
		}
		if positionRange.Valid {
			item.PositionRange = positionRange.String
		}
		if reasonSummary.Valid {
			item.ReasonSummary = reasonSummary.String
		}
		item.ValidFrom = validFrom.Format(time.RFC3339)
		item.ValidTo = validTo.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) GetStockRecommendationDetail(userID string, recoID string) (model.StockRecommendationDetail, error) {
	var item model.StockRecommendationDetail
	var takeProfit, stopLoss, riskNote sql.NullString
	err := r.db.QueryRow(`
SELECT d.reco_id, d.tech_score, d.fund_score, d.sentiment_score, d.money_flow_score, d.take_profit, d.stop_loss, d.risk_note
FROM stock_reco_details d
JOIN stock_recommendations r ON r.id = d.reco_id
WHERE d.reco_id = ? AND r.status IN ('PUBLISHED', 'ACTIVE', 'TRACKING', 'HIT_TAKE_PROFIT', 'HIT_STOP_LOSS', 'INVALIDATED', 'REVIEWED')`, recoID).Scan(
		&item.RecoID, &item.TechScore, &item.FundScore, &item.SentimentScore, &item.MoneyFlowScore, &takeProfit, &stopLoss, &riskNote,
	)
	if err != nil {
		return model.StockRecommendationDetail{}, err
	}
	if takeProfit.Valid {
		item.TakeProfit = takeProfit.String
	}
	if stopLoss.Valid {
		item.StopLoss = stopLoss.String
	}
	if riskNote.Valid {
		item.RiskNote = riskNote.String
	}
	return item, nil
}

func (r *MySQLGrowthRepo) GetStockRecommendationPerformance(userID string, recoID string) ([]model.RecommendationPerformancePoint, error) {
	var score sql.NullFloat64
	var validFrom, validTo time.Time
	err := r.db.QueryRow(`
SELECT score, valid_from, valid_to
FROM stock_recommendations
WHERE id = ? AND status IN ('PUBLISHED', 'ACTIVE', 'TRACKING', 'HIT_TAKE_PROFIT', 'HIT_STOP_LOSS', 'INVALIDATED', 'REVIEWED')`, recoID).Scan(&score, &validFrom, &validTo)
	if err != nil {
		return nil, err
	}

	base := 0.0
	if score.Valid {
		base = (score.Float64 - 50) / 100.0
	}
	points := make([]model.RecommendationPerformancePoint, 0, 5)
	for day := 0; day < 5; day++ {
		current := validFrom.AddDate(0, 0, day)
		if current.After(validTo) {
			break
		}
		value := base * (float64(day+1) / 5.0)
		value = float64(int(value*10000)) / 10000
		points = append(points, model.RecommendationPerformancePoint{
			Date:   current.Format("2006-01-02"),
			Return: value,
		})
	}
	if len(points) == 0 {
		points = append(points, model.RecommendationPerformancePoint{
			Date:   validFrom.Format("2006-01-02"),
			Return: base,
		})
	}
	return points, nil
}

func (r *MySQLGrowthRepo) GetStockRecommendationInsight(userID string, recoID string) (model.StockRecommendationInsight, error) {
	var item model.StockRecommendation
	var positionRange, reasonSummary sql.NullString
	var validFrom, validTo time.Time
	if err := r.db.QueryRow(`
SELECT id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary
FROM stock_recommendations
WHERE id = ? AND status IN ('PUBLISHED', 'ACTIVE', 'TRACKING', 'HIT_TAKE_PROFIT', 'HIT_STOP_LOSS', 'INVALIDATED', 'REVIEWED')`, recoID).Scan(
		&item.ID, &item.Symbol, &item.Name, &item.Score, &item.RiskLevel, &positionRange, &validFrom, &validTo, &item.Status, &reasonSummary,
	); err != nil {
		return model.StockRecommendationInsight{}, err
	}
	if positionRange.Valid {
		item.PositionRange = positionRange.String
	}
	if reasonSummary.Valid {
		item.ReasonSummary = reasonSummary.String
	}
	item.ValidFrom = validFrom.Format(time.RFC3339)
	item.ValidTo = validTo.Format(time.RFC3339)

	detail, detailErr := r.GetStockRecommendationDetail(userID, recoID)
	if detailErr != nil {
		if errors.Is(detailErr, sql.ErrNoRows) {
			detail = estimateStockRecoDetailByScore(recoID, item.Score)
		} else {
			return model.StockRecommendationInsight{}, detailErr
		}
	}
	if detail.RecoID == "" {
		detail.RecoID = recoID
	}

	performancePoints, err := r.GetStockRecommendationPerformance(userID, recoID)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}
	benchmarkPoints, benchmarkSymbol, benchmarkSource, err := r.buildBenchmarkSeries(performancePoints, validFrom, validTo)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}
	performanceStats := summarizeStockPerformance(performancePoints, benchmarkPoints, benchmarkSymbol, benchmarkSource)

	scoreFramework := buildStockRecoScoreFramework(item.Score, detail)
	relatedNews, err := r.listStockRelatedNews(item.Symbol, item.Name, validFrom, 6)
	if err != nil {
		return model.StockRecommendationInsight{}, err
	}

	return model.StockRecommendationInsight{
		Recommendation:   item,
		Detail:           detail,
		ScoreFramework:   scoreFramework,
		RelatedNews:      relatedNews,
		Performance:      performancePoints,
		Benchmark:        benchmarkPoints,
		PerformanceStats: performanceStats,
		Explanation:      r.buildStockStrategyExplanation(item, detail),
		GeneratedAt:      time.Now().Format(time.RFC3339),
	}, nil
}

func (r *MySQLGrowthRepo) GetStockRecommendationVersionHistory(userID string, recoID string) ([]model.StrategyVersionHistoryItem, error) {
	var item model.StockRecommendation
	var validFrom time.Time
	var reasonSummary, strategyVersion sql.NullString
	if err := r.db.QueryRow(`
SELECT id, symbol, valid_from, reason_summary, strategy_version
FROM stock_recommendations
WHERE id = ? AND status IN ('PUBLISHED', 'ACTIVE', 'TRACKING', 'HIT_TAKE_PROFIT', 'HIT_STOP_LOSS', 'INVALIDATED', 'REVIEWED')`,
		recoID,
	).Scan(&item.ID, &item.Symbol, &validFrom, &reasonSummary, &strategyVersion); err != nil {
		return nil, err
	}
	item.ValidFrom = validFrom.Format(time.RFC3339)
	if reasonSummary.Valid {
		item.ReasonSummary = reasonSummary.String
	}
	if strategyVersion.Valid {
		item.StrategyVersion = strategyVersion.String
	}

	detail, detailErr := r.GetStockRecommendationDetail(userID, recoID)
	if detailErr != nil && !errors.Is(detailErr, sql.ErrNoRows) {
		return nil, detailErr
	}
	explanation := r.buildStockStrategyExplanation(item, detail)
	contexts, err := r.listStrategyEngineAssetContexts("stock-selection", item.Symbol, defaultVersionHistoryLimit)
	if err != nil {
		return nil, err
	}
	if len(contexts) == 0 {
		return []model.StrategyVersionHistoryItem{
			buildFallbackVersionHistoryItem(
				explanation.PublishID,
				explanation.JobID,
				dateOnly(item.ValidFrom),
				explanation.PublishVersion,
				explanation.GeneratedAt,
				item.StrategyVersion,
				item.ReasonSummary,
				explanation,
			),
		}, nil
	}

	items := make([]model.StrategyVersionHistoryItem, 0, len(contexts))
	for _, ctx := range contexts {
		historyItem := buildStrategyVersionHistoryItem(
			ctx,
			asString(ctx.asset["reason_summary"]),
			firstNonEmpty(asString(ctx.asset["strategy_version"]), item.StrategyVersion),
			[]string{
				"市场种子输入",
				"特征工程与打分",
				"多角色评审",
				"情景模拟",
				"风险过滤与发布",
			},
		)
		if summary, summaryErr := r.loadStockSelectionEvaluationSummaryByContext(ctx, item.Symbol); summaryErr == nil {
			enrichStockSelectionVersionHistoryEvaluationMeta(&historyItem, summary)
		}
		items = append(items, historyItem)
	}
	return items, nil
}

func estimateStockRecoDetailByScore(recoID string, totalScore float64) model.StockRecommendationDetail {
	base := clampFloat(totalScore, 55, 98)
	return model.StockRecommendationDetail{
		RecoID:         recoID,
		TechScore:      roundTo(clampFloat(base-2, 50, 99), 2),
		FundScore:      roundTo(clampFloat(base+1, 50, 99), 2),
		SentimentScore: roundTo(clampFloat(base-4, 50, 99), 2),
		MoneyFlowScore: roundTo(clampFloat(base-1, 50, 99), 2),
		TakeProfit:     "上涨 6%-10% 分批止盈",
		StopLoss:       "跌破关键支撑位止损",
		RiskNote:       "评分由推荐总分估算，建议结合盘中波动与仓位纪律执行。",
	}
}

func buildStockRecoScoreFramework(totalScore float64, detail model.StockRecommendationDetail) model.StockRecommendationScoreFramework {
	factors := []model.StockRecommendationFactorScore{
		{Key: "tech", Label: "技术因子", Weight: 0.30, Score: detail.TechScore},
		{Key: "fund", Label: "基本面因子", Weight: 0.30, Score: detail.FundScore},
		{Key: "sentiment", Label: "情绪因子", Weight: 0.20, Score: detail.SentimentScore},
		{Key: "flow", Label: "资金流因子", Weight: 0.20, Score: detail.MoneyFlowScore},
	}
	sumWeight := 0.0
	weightedScore := 0.0
	for index := range factors {
		if factors[index].Score <= 0 {
			continue
		}
		factors[index].Contribution = roundTo(factors[index].Weight*factors[index].Score, 2)
		weightedScore += factors[index].Contribution
		sumWeight += factors[index].Weight
	}
	if sumWeight > 0 {
		weightedScore = weightedScore / sumWeight
	}
	weightedScore = roundTo(weightedScore, 2)
	return model.StockRecommendationScoreFramework{
		Method:        "growth-v1 (tech30 + fund30 + sentiment20 + flow20)",
		TotalScore:    roundTo(totalScore, 2),
		WeightedScore: weightedScore,
		ScoreGap:      roundTo(totalScore-weightedScore, 2),
		Factors:       factors,
	}
}

func buildEstimatedBenchmark(points []model.RecommendationPerformancePoint) []model.RecommendationPerformancePoint {
	result := make([]model.RecommendationPerformancePoint, 0, len(points))
	for _, point := range points {
		result = append(result, model.RecommendationPerformancePoint{
			Date:   point.Date,
			Return: roundTo(clampFloat(point.Return*0.55, -0.2, 0.2), 4),
		})
	}
	return result
}

func (r *MySQLGrowthRepo) buildBenchmarkSeries(points []model.RecommendationPerformancePoint, validFrom time.Time, validTo time.Time) ([]model.RecommendationPerformancePoint, string, string, error) {
	estimated := buildEstimatedBenchmark(points)
	if len(points) == 0 {
		return estimated, "000300.SH", "estimated: no strategy points", nil
	}

	startDate := validFrom
	endDate := validTo
	if startDate.IsZero() || endDate.IsZero() || endDate.Before(startDate) {
		for index := range points {
			ts, err := time.Parse("2006-01-02", points[index].Date)
			if err != nil {
				continue
			}
			if startDate.IsZero() || ts.Before(startDate) {
				startDate = ts
			}
			if endDate.IsZero() || ts.After(endDate) {
				endDate = ts
			}
		}
	}
	if startDate.IsZero() || endDate.IsZero() || endDate.Before(startDate) {
		return estimated, "000300.SH", "estimated: invalid date range", nil
	}

	candidates := []struct {
		Symbol string
		Name   string
	}{
		{Symbol: "000300.SH", Name: "CSI300"},
		{Symbol: "000905.SH", Name: "CSI500"},
	}

	rows, err := r.db.Query(`
SELECT symbol, trade_date, close_price, prev_close_price
FROM stock_market_quotes
WHERE symbol IN (?, ?)
  AND trade_date >= ?
  AND trade_date <= ?
ORDER BY symbol ASC, trade_date ASC`,
		candidates[0].Symbol, candidates[1].Symbol, startDate, endDate,
	)
	if err != nil {
		if isTableNotFoundError(err) {
			return estimated, "000300.SH", "estimated: stock_market_quotes table missing", nil
		}
		return nil, "", "", err
	}
	defer rows.Close()

	quoteMap := map[string]map[string]float64{
		candidates[0].Symbol: {},
		candidates[1].Symbol: {},
	}
	lastClose := map[string]float64{}
	for rows.Next() {
		var symbol string
		var tradeDate time.Time
		var closePrice, prevClosePrice sql.NullFloat64
		if err := rows.Scan(&symbol, &tradeDate, &closePrice, &prevClosePrice); err != nil {
			return nil, "", "", err
		}
		if !closePrice.Valid || closePrice.Float64 <= 0 {
			continue
		}
		ret := math.NaN()
		if prevClosePrice.Valid && prevClosePrice.Float64 > 0 {
			ret = closePrice.Float64/prevClosePrice.Float64 - 1
		} else if previous, exists := lastClose[symbol]; exists && previous > 0 {
			ret = closePrice.Float64/previous - 1
		}
		lastClose[symbol] = closePrice.Float64
		if !math.IsNaN(ret) {
			dateKey := tradeDate.Format("2006-01-02")
			quoteMap[symbol][dateKey] = roundTo(clampFloat(ret, -0.2, 0.2), 4)
		}
	}

	perfDates := make([]string, 0, len(points))
	for _, point := range points {
		if point.Date == "" {
			continue
		}
		dateKey := point.Date
		if ts, err := time.Parse("2006-01-02", point.Date); err == nil {
			dateKey = ts.Format("2006-01-02")
		}
		perfDates = append(perfDates, dateKey)
	}

	bestSymbol := candidates[0].Symbol
	bestName := candidates[0].Name
	bestCoverage := -1
	for _, candidate := range candidates {
		coverage := 0
		for _, date := range perfDates {
			if _, ok := quoteMap[candidate.Symbol][date]; ok {
				coverage++
			}
		}
		if coverage > bestCoverage {
			bestCoverage = coverage
			bestSymbol = candidate.Symbol
			bestName = candidate.Name
		}
	}
	if bestCoverage <= 0 {
		return estimated, "000300.SH", "estimated: no benchmark quotes matched", nil
	}

	result := make([]model.RecommendationPerformancePoint, 0, len(points))
	actualCount := 0
	for index, point := range points {
		dateKey := point.Date
		if ts, err := time.Parse("2006-01-02", point.Date); err == nil {
			dateKey = ts.Format("2006-01-02")
		}
		value := estimated[index].Return
		if actual, ok := quoteMap[bestSymbol][dateKey]; ok {
			value = actual
			actualCount++
		}
		result = append(result, model.RecommendationPerformancePoint{
			Date:   point.Date,
			Return: value,
		})
	}

	source := fmt.Sprintf("actual: %s (%d/%d points), missing filled by estimated", bestName, actualCount, len(points))
	if actualCount == len(points) {
		source = fmt.Sprintf("actual: %s", bestName)
	}
	return result, bestSymbol, source, nil
}

func summarizeStockPerformance(points []model.RecommendationPerformancePoint, benchmark []model.RecommendationPerformancePoint, benchmarkSymbol string, benchmarkSource string) model.StockRecommendationPerformanceSummary {
	if len(points) == 0 {
		return model.StockRecommendationPerformanceSummary{
			BenchmarkSymbol: benchmarkSymbol,
			BenchmarkSource: benchmarkSource,
		}
	}
	sampleDays := len(points)
	wins := 0
	sumDaily := 0.0
	curve := 1.0
	peak := 1.0
	maxDrawdown := 0.0
	for _, point := range points {
		if point.Return > 0 {
			wins++
		}
		sumDaily += point.Return
		curve *= 1 + point.Return
		if curve > peak {
			peak = curve
		}
		if peak > 0 {
			dd := (peak - curve) / peak
			if dd > maxDrawdown {
				maxDrawdown = dd
			}
		}
	}

	benchmarkCurve := 1.0
	for _, point := range benchmark {
		benchmarkCurve *= 1 + point.Return
	}

	cumulative := curve - 1
	benchmarkCumulative := benchmarkCurve - 1
	return model.StockRecommendationPerformanceSummary{
		SampleDays:                sampleDays,
		WinRate:                   roundTo(float64(wins)/float64(sampleDays), 4),
		AvgDailyReturn:            roundTo(sumDaily/float64(sampleDays), 4),
		CumulativeReturn:          roundTo(cumulative, 4),
		BenchmarkCumulativeReturn: roundTo(benchmarkCumulative, 4),
		ExcessReturn:              roundTo(cumulative-benchmarkCumulative, 4),
		MaxDrawdown:               roundTo(maxDrawdown, 4),
		BenchmarkSymbol:           benchmarkSymbol,
		BenchmarkSource:           benchmarkSource,
	}
}

func (r *MySQLGrowthRepo) listStockRelatedNews(symbol string, name string, anchor time.Time, limit int) ([]model.StockRecommendationRelatedNews, error) {
	keywords := buildStockNewsKeywords(symbol, name)
	if len(keywords) == 0 {
		return []model.StockRecommendationRelatedNews{}, nil
	}
	if limit <= 0 {
		limit = 6
	}
	if anchor.IsZero() {
		anchor = time.Now()
	}
	windowStart := anchor.AddDate(0, 0, -45)

	conditions := make([]string, 0, len(keywords))
	args := make([]interface{}, 0, 1+len(keywords)*3+1)
	args = append(args, windowStart)
	for _, keyword := range keywords {
		conditions = append(conditions, "(na.title LIKE ? OR COALESCE(na.summary, '') LIKE ? OR COALESCE(na.content, '') LIKE ?)")
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}

	query := `
SELECT na.id, na.title, na.summary, nc.name AS source, na.visibility, na.published_at
FROM news_articles na
LEFT JOIN news_categories nc ON nc.id = na.category_id
WHERE na.status = 'PUBLISHED'
  AND na.published_at >= ?
  AND (` + strings.Join(conditions, " OR ") + `)
ORDER BY na.published_at DESC
LIMIT ?`
	args = append(args, limit*4)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.StockRecommendationRelatedNews, 0, limit)
	for rows.Next() {
		var item model.StockRecommendationRelatedNews
		var summary, source sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Title, &summary, &source, &item.Visibility, &publishedAt); err != nil {
			return nil, err
		}
		if summary.Valid {
			item.Summary = truncateByRunes(normalizeUTF8Text(summary.String), 120)
		}
		if source.Valid {
			item.Source = source.String
		}
		if item.Source == "" {
			item.Source = "资讯中心"
		}
		if publishedAt.Valid {
			item.PublishedAt = publishedAt.Time.Format(time.RFC3339)
			item.RelevanceScore = scoreNewsRelevance(item.Title, item.Summary, keywords, publishedAt.Time, anchor)
		} else {
			item.RelevanceScore = scoreNewsRelevance(item.Title, item.Summary, keywords, time.Time{}, anchor)
		}
		items = append(items, item)
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].RelevanceScore == items[j].RelevanceScore {
			return items[i].PublishedAt > items[j].PublishedAt
		}
		return items[i].RelevanceScore > items[j].RelevanceScore
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func buildStockNewsKeywords(symbol string, name string) []string {
	items := make([]string, 0, 4)
	seen := make(map[string]struct{})
	push := func(value string) {
		key := strings.TrimSpace(value)
		if key == "" {
			return
		}
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		items = append(items, key)
	}

	normalizedName := strings.TrimSpace(name)
	push(normalizedName)

	rawSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	push(rawSymbol)
	if idx := strings.Index(rawSymbol, "."); idx > 0 {
		push(rawSymbol[:idx])
	}
	return items
}

func scoreNewsRelevance(title string, summary string, keywords []string, publishedAt time.Time, anchor time.Time) float64 {
	titleLower := strings.ToLower(title)
	summaryLower := strings.ToLower(summary)
	score := 0.0
	for _, keyword := range keywords {
		key := strings.ToLower(strings.TrimSpace(keyword))
		if key == "" {
			continue
		}
		if strings.Contains(titleLower, key) {
			score += 0.45
		}
		if strings.Contains(summaryLower, key) {
			score += 0.20
		}
	}

	if !publishedAt.IsZero() {
		diff := anchor.Sub(publishedAt)
		if diff < 0 {
			diff = -diff
		}
		hours := diff.Hours()
		if hours <= 72 {
			score += 0.2
		} else if hours <= 168 {
			score += 0.1
		}
	}
	return roundTo(clampFloat(score, 0, 1), 4)
}

func (r *MySQLGrowthRepo) ListFuturesStrategies(userID string, contract string, status string, page int, pageSize int) ([]model.FuturesStrategy, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if contract != "" {
		filter += " AND contract = ?"
		args = append(args, contract)
	}
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	} else {
		filter += " AND status IN ('PUBLISHED', 'ACTIVE')"
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM futures_strategies"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary
FROM futures_strategies` + filter + `
ORDER BY valid_from DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.FuturesStrategy, 0)
	for rows.Next() {
		var item model.FuturesStrategy
		var name, positionRange, reasonSummary sql.NullString
		var validFrom, validTo time.Time
		if err := rows.Scan(&item.ID, &item.Contract, &name, &item.Direction, &item.RiskLevel, &positionRange, &validFrom, &validTo, &item.Status, &reasonSummary); err != nil {
			return nil, 0, err
		}
		if name.Valid {
			item.Name = name.String
		}
		if positionRange.Valid {
			item.PositionRange = positionRange.String
		}
		if reasonSummary.Valid {
			item.ReasonSummary = reasonSummary.String
		}
		item.ValidFrom = validFrom.Format(time.RFC3339)
		item.ValidTo = validTo.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) GetFuturesStrategyDetail(userID string, strategyID string) (model.FuturesStrategy, error) {
	var item model.FuturesStrategy
	var name, positionRange, reasonSummary sql.NullString
	var validFrom, validTo time.Time
	err := r.db.QueryRow(`
SELECT id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary
FROM futures_strategies
WHERE id = ?`, strategyID).Scan(
		&item.ID, &item.Contract, &name, &item.Direction, &item.RiskLevel, &positionRange, &validFrom, &validTo, &item.Status, &reasonSummary,
	)
	if err != nil {
		return model.FuturesStrategy{}, err
	}
	if name.Valid {
		item.Name = name.String
	}
	if positionRange.Valid {
		item.PositionRange = positionRange.String
	}
	if reasonSummary.Valid {
		item.ReasonSummary = reasonSummary.String
	}
	item.ValidFrom = validFrom.Format(time.RFC3339)
	item.ValidTo = validTo.Format(time.RFC3339)
	return item, nil
}

func (r *MySQLGrowthRepo) GetFuturesStrategyInsight(userID string, strategyID string) (model.FuturesStrategyInsight, error) {
	strategy, err := r.GetFuturesStrategyDetail(userID, strategyID)
	if err != nil {
		return model.FuturesStrategyInsight{}, err
	}

	guidance, _ := r.getLatestFuturesGuidanceByContract(strategy.Contract)

	validFrom, validTo := parseStrategyWindow(strategy.ValidFrom, strategy.ValidTo)
	performance, statsFromReview, err := r.buildFuturesStrategyPerformance(strategyID, strategy.RiskLevel, validFrom, validTo)
	if err != nil {
		return model.FuturesStrategyInsight{}, err
	}

	benchmarkSymbol := futuresBenchmarkSymbolByContract(strategy.Contract)
	benchmark, benchmarkSource, err := r.buildFuturesBenchmarkSeries(performance, benchmarkSymbol)
	if err != nil {
		return model.FuturesStrategyInsight{}, err
	}
	stats := summarizeFuturesPerformance(performance, benchmark, benchmarkSymbol, benchmarkSource, statsFromReview.MaxDrawdown)

	relatedEvents, err := r.listFuturesRelatedEvents(strategy.Contract, 6)
	if err != nil {
		return model.FuturesStrategyInsight{}, err
	}
	anchor := validFrom
	if anchor.IsZero() {
		anchor = time.Now()
	}
	relatedNews, err := r.listStockRelatedNews(strategy.Contract, strategy.Name, anchor, 6)
	if err != nil {
		return model.FuturesStrategyInsight{}, err
	}

	scoreWeights := r.loadFuturesStrategyScoreWeights()
	scoreFramework := buildFuturesScoreFramework(strategy, guidance, stats, relatedNews, relatedEvents, scoreWeights)
	return model.FuturesStrategyInsight{
		Strategy:         strategy,
		Guidance:         guidance,
		ScoreFramework:   scoreFramework,
		RelatedNews:      relatedNews,
		RelatedEvents:    relatedEvents,
		Performance:      performance,
		Benchmark:        benchmark,
		PerformanceStats: stats,
		Explanation:      r.buildFuturesStrategyExplanation(strategy, guidance),
		GeneratedAt:      time.Now().Format(time.RFC3339),
	}, nil
}

func (r *MySQLGrowthRepo) GetFuturesStrategyVersionHistory(userID string, strategyID string) ([]model.StrategyVersionHistoryItem, error) {
	strategy, err := r.GetFuturesStrategyDetail(userID, strategyID)
	if err != nil {
		return nil, err
	}
	guidance, _ := r.getLatestFuturesGuidanceByContract(strategy.Contract)
	explanation := r.buildFuturesStrategyExplanation(strategy, guidance)
	contexts, err := r.listStrategyEngineAssetContexts("futures-strategy", strategy.Contract, defaultVersionHistoryLimit)
	if err != nil {
		return nil, err
	}
	if len(contexts) == 0 {
		return []model.StrategyVersionHistoryItem{
			buildFallbackVersionHistoryItem(
				explanation.PublishID,
				explanation.JobID,
				dateOnly(strategy.ValidFrom),
				explanation.PublishVersion,
				explanation.GeneratedAt,
				explanation.StrategyVersion,
				strategy.ReasonSummary,
				explanation,
			),
		}, nil
	}

	items := make([]model.StrategyVersionHistoryItem, 0, len(contexts))
	for _, ctx := range contexts {
		items = append(items, buildStrategyVersionHistoryItem(
			ctx,
			asString(ctx.asset["reason_summary"]),
			firstNonEmpty(asString(ctx.asset["strategy_version"]), explanation.StrategyVersion, "futures-mvp-v1"),
			[]string{
				"合约池初始化",
				"方向/价位特征评估",
				"多角色评审",
				"情景推演",
				"风险与发布过滤",
			},
		))
	}
	return items, nil
}

func (r *MySQLGrowthRepo) getLatestFuturesGuidanceByContract(contract string) (model.FuturesGuidance, error) {
	var item model.FuturesGuidance
	var entryRange, takeProfitRange, stopLossRange, invalidCondition sql.NullString
	var validTo time.Time
	err := r.db.QueryRow(`
SELECT contract, guidance_direction, position_level, entry_range, take_profit_range, stop_loss_range, risk_level, invalid_condition, valid_to
FROM futures_guidances
WHERE contract = ?
ORDER BY valid_to DESC
LIMIT 1`, strings.TrimSpace(contract)).Scan(
		&item.Contract, &item.GuidanceDirection, &item.PositionLevel, &entryRange, &takeProfitRange, &stopLossRange, &item.RiskLevel, &invalidCondition, &validTo,
	)
	if err != nil {
		return model.FuturesGuidance{}, err
	}
	if entryRange.Valid {
		item.EntryRange = entryRange.String
	}
	if takeProfitRange.Valid {
		item.TakeProfitRange = takeProfitRange.String
	}
	if stopLossRange.Valid {
		item.StopLossRange = stopLossRange.String
	}
	if invalidCondition.Valid {
		item.InvalidCondition = invalidCondition.String
	}
	item.ValidTo = validTo.Format(time.RFC3339)
	return item, nil
}

type futuresPerformanceBuildMeta struct {
	MaxDrawdown float64
}

func (r *MySQLGrowthRepo) buildFuturesStrategyPerformance(strategyID string, riskLevel string, validFrom time.Time, validTo time.Time) ([]model.RecommendationPerformancePoint, futuresPerformanceBuildMeta, error) {
	rows, err := r.db.Query(`
SELECT review_date, pnl, max_drawdown
FROM futures_reviews
WHERE strategy_id = ?
ORDER BY review_date ASC`, strategyID)
	if err != nil {
		if !isTableNotFoundError(err) {
			return nil, futuresPerformanceBuildMeta{}, err
		}
		rows = nil
	}

	points := make([]model.RecommendationPerformancePoint, 0, 12)
	meta := futuresPerformanceBuildMeta{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var reviewDate time.Time
			var pnl, maxDrawdown sql.NullFloat64
			if err := rows.Scan(&reviewDate, &pnl, &maxDrawdown); err != nil {
				return nil, futuresPerformanceBuildMeta{}, err
			}
			value := normalizeFuturesPnLToReturn(pnl)
			points = append(points, model.RecommendationPerformancePoint{
				Date:   reviewDate.Format("2006-01-02"),
				Return: roundTo(value, 4),
			})
			if maxDrawdown.Valid {
				meta.MaxDrawdown = math.Max(meta.MaxDrawdown, math.Abs(maxDrawdown.Float64))
			}
		}
	}

	if len(points) > 0 {
		return points, meta, nil
	}
	return estimateFuturesPerformanceByRisk(riskLevel, validFrom, validTo), meta, nil
}

func normalizeFuturesPnLToReturn(pnl sql.NullFloat64) float64 {
	if !pnl.Valid {
		return 0
	}
	value := pnl.Float64
	if math.Abs(value) > 1.5 {
		value = value / 100.0
	}
	return clampFloat(value, -0.3, 0.3)
}

func estimateFuturesPerformanceByRisk(riskLevel string, validFrom time.Time, validTo time.Time) []model.RecommendationPerformancePoint {
	base := 0.009
	switch strings.ToUpper(strings.TrimSpace(riskLevel)) {
	case "LOW":
		base = 0.006
	case "HIGH":
		base = 0.013
	}

	start := validFrom
	if start.IsZero() {
		start = time.Now().AddDate(0, 0, -4)
	}
	end := validTo
	if end.IsZero() || end.Before(start) {
		end = start.AddDate(0, 0, 4)
	}

	points := make([]model.RecommendationPerformancePoint, 0, 5)
	current := start
	for index := 0; index < 5; index++ {
		if current.After(end) {
			break
		}
		value := base * (0.8 + float64(index)*0.12)
		if index == 2 {
			value = -value * 0.45
		}
		points = append(points, model.RecommendationPerformancePoint{
			Date:   current.Format("2006-01-02"),
			Return: roundTo(clampFloat(value, -0.3, 0.3), 4),
		})
		current = current.AddDate(0, 0, 1)
	}
	if len(points) == 0 {
		points = append(points, model.RecommendationPerformancePoint{
			Date:   start.Format("2006-01-02"),
			Return: roundTo(base, 4),
		})
	}
	return points
}

func parseStrategyWindow(validFrom string, validTo string) (time.Time, time.Time) {
	start, _ := time.Parse(time.RFC3339, strings.TrimSpace(validFrom))
	end, _ := time.Parse(time.RFC3339, strings.TrimSpace(validTo))
	return start, end
}

func futuresBenchmarkSymbolByContract(contract string) string {
	source := strings.ToUpper(strings.TrimSpace(contract))
	switch {
	case strings.HasPrefix(source, "IF"):
		return "000300.SH"
	case strings.HasPrefix(source, "IC"):
		return "000905.SH"
	case strings.HasPrefix(source, "IH"):
		return "000016.SH"
	case strings.HasPrefix(source, "IM"):
		return "000852.SH"
	default:
		return "000300.SH"
	}
}

func (r *MySQLGrowthRepo) buildFuturesBenchmarkSeries(points []model.RecommendationPerformancePoint, benchmarkSymbol string) ([]model.RecommendationPerformancePoint, string, error) {
	estimated := buildEstimatedBenchmark(points)
	if len(points) == 0 {
		return estimated, "estimated: no strategy points", nil
	}
	if benchmarkSymbol == "" {
		return estimated, "estimated: benchmark symbol missing", nil
	}

	minDate, maxDate := "", ""
	for _, point := range points {
		dateKey := strings.TrimSpace(point.Date)
		if dateKey == "" {
			continue
		}
		if minDate == "" || dateKey < minDate {
			minDate = dateKey
		}
		if maxDate == "" || dateKey > maxDate {
			maxDate = dateKey
		}
	}
	if minDate == "" || maxDate == "" {
		return estimated, "estimated: invalid point dates", nil
	}

	rows, err := r.db.Query(`
SELECT trade_date, close_price, prev_close_price
FROM stock_market_quotes
WHERE symbol = ?
  AND trade_date >= ?
  AND trade_date <= ?
ORDER BY trade_date ASC`, benchmarkSymbol, minDate, maxDate)
	if err != nil {
		if isTableNotFoundError(err) {
			return estimated, "estimated: stock_market_quotes table missing", nil
		}
		return nil, "", err
	}
	defer rows.Close()

	valueByDate := map[string]float64{}
	lastClose := 0.0
	for rows.Next() {
		var tradeDate time.Time
		var closePrice, prevClosePrice sql.NullFloat64
		if err := rows.Scan(&tradeDate, &closePrice, &prevClosePrice); err != nil {
			return nil, "", err
		}
		if !closePrice.Valid || closePrice.Float64 <= 0 {
			continue
		}
		ret := math.NaN()
		if prevClosePrice.Valid && prevClosePrice.Float64 > 0 {
			ret = closePrice.Float64/prevClosePrice.Float64 - 1
		} else if lastClose > 0 {
			ret = closePrice.Float64/lastClose - 1
		}
		lastClose = closePrice.Float64
		if !math.IsNaN(ret) {
			valueByDate[tradeDate.Format("2006-01-02")] = roundTo(clampFloat(ret, -0.2, 0.2), 4)
		}
	}

	result := make([]model.RecommendationPerformancePoint, 0, len(points))
	actualCount := 0
	for index, point := range points {
		value := estimated[index].Return
		if actual, ok := valueByDate[point.Date]; ok {
			value = actual
			actualCount++
		}
		result = append(result, model.RecommendationPerformancePoint{
			Date:   point.Date,
			Return: value,
		})
	}

	if actualCount == 0 {
		return result, "estimated: no benchmark quotes matched", nil
	}
	if actualCount < len(points) {
		return result, fmt.Sprintf("actual mixed: %d/%d points from %s", actualCount, len(points), benchmarkSymbol), nil
	}
	return result, fmt.Sprintf("actual: %s", benchmarkSymbol), nil
}

func summarizeFuturesPerformance(points []model.RecommendationPerformancePoint, benchmark []model.RecommendationPerformancePoint, benchmarkSymbol string, benchmarkSource string, overrideMaxDrawdown float64) model.FuturesStrategyPerformanceSummary {
	if len(points) == 0 {
		return model.FuturesStrategyPerformanceSummary{
			BenchmarkSymbol: benchmarkSymbol,
			BenchmarkSource: benchmarkSource,
		}
	}

	sampleDays := len(points)
	positiveDays := 0
	sumDaily := 0.0
	curve := 1.0
	peak := 1.0
	maxDrawdown := 0.0
	for _, point := range points {
		if point.Return > 0 {
			positiveDays++
		}
		sumDaily += point.Return
		curve *= 1 + point.Return
		if curve > peak {
			peak = curve
		}
		if peak > 0 {
			drawdown := (peak - curve) / peak
			if drawdown > maxDrawdown {
				maxDrawdown = drawdown
			}
		}
	}
	if overrideMaxDrawdown > 0 {
		maxDrawdown = math.Max(maxDrawdown, clampFloat(overrideMaxDrawdown, 0, 0.99))
	}

	benchmarkCurve := 1.0
	for _, point := range benchmark {
		benchmarkCurve *= 1 + point.Return
	}
	cumulative := curve - 1
	benchmarkCumulative := benchmarkCurve - 1
	return model.FuturesStrategyPerformanceSummary{
		SampleDays:                sampleDays,
		WinRate:                   roundTo(float64(positiveDays)/float64(sampleDays), 4),
		AvgDailyReturn:            roundTo(sumDaily/float64(sampleDays), 4),
		CumulativeReturn:          roundTo(cumulative, 4),
		BenchmarkCumulativeReturn: roundTo(benchmarkCumulative, 4),
		ExcessReturn:              roundTo(cumulative-benchmarkCumulative, 4),
		MaxDrawdown:               roundTo(maxDrawdown, 4),
		BenchmarkSymbol:           benchmarkSymbol,
		BenchmarkSource:           benchmarkSource,
	}
}

func (r *MySQLGrowthRepo) listFuturesRelatedEvents(contract string, limit int) ([]model.MarketEvent, error) {
	if limit <= 0 {
		limit = 6
	}
	rows, err := r.db.Query(`
SELECT id, event_type, symbol, summary, trigger_rule, source, created_at
FROM market_events
WHERE symbol = ? OR symbol = 'ALL'
ORDER BY created_at DESC
LIMIT ?`, strings.TrimSpace(contract), limit)
	if err != nil {
		if isTableNotFoundError(err) {
			return []model.MarketEvent{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	items := make([]model.MarketEvent, 0, limit)
	for rows.Next() {
		var item model.MarketEvent
		var summary, triggerRule, source sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.EventType, &item.Symbol, &summary, &triggerRule, &source, &createdAt); err != nil {
			return nil, err
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
		items = append(items, item)
	}
	return items, nil
}

const futuresStrategyScoreWeightsConfigKey = "futures.strategy.score_weights"

var futuresStrategyScoreWeightKeys = []string{
	"trend",
	"structure",
	"flow",
	"risk",
	"news",
	"performance",
}

func defaultFuturesStrategyScoreWeights() map[string]float64 {
	return map[string]float64{
		"trend":       0.25,
		"structure":   0.20,
		"flow":        0.15,
		"risk":        0.20,
		"news":        0.10,
		"performance": 0.10,
	}
}

func normalizeFuturesStrategyScoreWeightKey(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "trend", "trend_factor":
		return "trend"
	case "structure", "structure_factor":
		return "structure"
	case "flow", "flow_factor", "event", "events":
		return "flow"
	case "risk", "risk_factor", "risk_control":
		return "risk"
	case "news", "news_factor":
		return "news"
	case "performance", "performance_factor", "perf":
		return "performance"
	default:
		return ""
	}
}

func normalizeFuturesStrategyScoreWeights(raw map[string]float64) map[string]float64 {
	weights := defaultFuturesStrategyScoreWeights()
	if raw == nil {
		return weights
	}
	sum := 0.0
	for _, key := range futuresStrategyScoreWeightKeys {
		if value, ok := raw[key]; ok {
			weights[key] = clampFloat(value, 0, 1)
		}
		sum += weights[key]
	}
	if sum <= 0 {
		return defaultFuturesStrategyScoreWeights()
	}
	for _, key := range futuresStrategyScoreWeightKeys {
		weights[key] = weights[key] / sum
	}
	return weights
}

func parseFloatFromInterface(value interface{}) (float64, bool) {
	switch raw := value.(type) {
	case float64:
		if !math.IsNaN(raw) && !math.IsInf(raw, 0) {
			return raw, true
		}
	case float32:
		parsed := float64(raw)
		if !math.IsNaN(parsed) && !math.IsInf(parsed, 0) {
			return parsed, true
		}
	case int:
		return float64(raw), true
	case int32:
		return float64(raw), true
	case int64:
		return float64(raw), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
		if err == nil && !math.IsNaN(parsed) && !math.IsInf(parsed, 0) {
			return parsed, true
		}
	}
	return 0, false
}

func parseFuturesStrategyScoreWeights(raw string) (map[string]float64, bool) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return nil, false
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(text), &payload); err != nil {
		return nil, false
	}
	weights := defaultFuturesStrategyScoreWeights()
	hasAny := false
	for key, value := range payload {
		normalizedKey := normalizeFuturesStrategyScoreWeightKey(key)
		if normalizedKey == "" {
			continue
		}
		parsed, ok := parseFloatFromInterface(value)
		if !ok {
			continue
		}
		if parsed > 1 {
			parsed = parsed / 100
		}
		weights[normalizedKey] = clampFloat(parsed, 0, 1)
		hasAny = true
	}
	if !hasAny {
		return nil, false
	}
	return normalizeFuturesStrategyScoreWeights(weights), true
}

func (r *MySQLGrowthRepo) loadFuturesStrategyScoreWeights() map[string]float64 {
	defaults := defaultFuturesStrategyScoreWeights()
	var configValue sql.NullString
	err := r.db.QueryRow(`
SELECT config_value
FROM system_configs
WHERE LOWER(config_key) = LOWER(?)
ORDER BY updated_at DESC
LIMIT 1`, futuresStrategyScoreWeightsConfigKey).Scan(&configValue)
	if err != nil || !configValue.Valid {
		return defaults
	}
	parsed, ok := parseFuturesStrategyScoreWeights(configValue.String)
	if !ok {
		return defaults
	}
	return parsed
}

func formatCompactPercent(value float64) string {
	percent := roundTo(value*100, 1)
	if math.Abs(percent-math.Round(percent)) < 0.05 {
		return fmt.Sprintf("%.0f", math.Round(percent))
	}
	return fmt.Sprintf("%.1f", percent)
}

func futuresStrategyScoreMethod(weights map[string]float64) string {
	return fmt.Sprintf(
		"futures-v1 (trend%s + structure%s + flow%s + risk%s + news%s + performance%s)",
		formatCompactPercent(weights["trend"]),
		formatCompactPercent(weights["structure"]),
		formatCompactPercent(weights["flow"]),
		formatCompactPercent(weights["risk"]),
		formatCompactPercent(weights["news"]),
		formatCompactPercent(weights["performance"]),
	)
}

func buildFuturesScoreFramework(strategy model.FuturesStrategy, guidance model.FuturesGuidance, stats model.FuturesStrategyPerformanceSummary, relatedNews []model.StockRecommendationRelatedNews, relatedEvents []model.MarketEvent, factorWeights map[string]float64) model.FuturesStrategyScoreFramework {
	weights := normalizeFuturesStrategyScoreWeights(factorWeights)

	trendScore := 68.0
	if strings.EqualFold(strategy.Direction, "LONG") || strings.EqualFold(strategy.Direction, "SHORT") {
		trendScore += 6
	}
	reasonText := strings.ToLower(strings.TrimSpace(strategy.ReasonSummary))
	if strings.Contains(reasonText, "趋势") {
		trendScore += 8
	}
	if strings.Contains(reasonText, "突破") {
		trendScore += 5
	}
	trendScore = clampFloat(trendScore, 45, 96)

	structureScore := 56.0
	if guidance.EntryRange != "" {
		structureScore += 12
	}
	if guidance.TakeProfitRange != "" {
		structureScore += 10
	}
	if guidance.StopLossRange != "" {
		structureScore += 10
	}
	if guidance.InvalidCondition != "" {
		structureScore += 6
	}
	structureScore = clampFloat(structureScore, 40, 95)

	eventScore := clampFloat(58+float64(len(relatedEvents))*5, 45, 90)

	riskScore := 74.0
	switch strings.ToUpper(strings.TrimSpace(strategy.RiskLevel)) {
	case "LOW":
		riskScore = 86
	case "MEDIUM":
		riskScore = 76
	case "HIGH":
		riskScore = 64
	}

	newsAvg := 0.0
	for _, item := range relatedNews {
		newsAvg += clampFloat(item.RelevanceScore, 0, 1)
	}
	if len(relatedNews) > 0 {
		newsAvg = newsAvg / float64(len(relatedNews))
	}
	newsScore := clampFloat(55+newsAvg*35, 45, 92)

	perfScore := clampFloat(50+stats.WinRate*25+stats.ExcessReturn*120, 40, 96)

	factors := []model.FuturesStrategyFactorScore{
		{Key: "trend", Label: "趋势因子", Weight: weights["trend"], Score: roundTo(trendScore, 2)},
		{Key: "structure", Label: "结构因子", Weight: weights["structure"], Score: roundTo(structureScore, 2)},
		{Key: "flow", Label: "资金与事件因子", Weight: weights["flow"], Score: roundTo(eventScore, 2)},
		{Key: "risk", Label: "风险控制因子", Weight: weights["risk"], Score: roundTo(riskScore, 2)},
		{Key: "news", Label: "资讯因子", Weight: weights["news"], Score: roundTo(newsScore, 2)},
		{Key: "performance", Label: "绩效因子", Weight: weights["performance"], Score: roundTo(perfScore, 2)},
	}

	total := 0.0
	weightSum := 0.0
	for index := range factors {
		factors[index].Contribution = roundTo(factors[index].Score*factors[index].Weight, 2)
		total += factors[index].Contribution
		weightSum += factors[index].Weight
	}
	weighted := total
	if weightSum > 0 {
		weighted = total / weightSum
	}
	weighted = roundTo(weighted, 2)
	totalScore := roundTo(weighted+0.4, 2)
	return model.FuturesStrategyScoreFramework{
		Method:        futuresStrategyScoreMethod(weights),
		TotalScore:    totalScore,
		WeightedScore: weighted,
		ScoreGap:      roundTo(totalScore-weighted, 2),
		Factors:       factors,
	}
}

func (r *MySQLGrowthRepo) ListMembershipProducts(status string, page int, pageSize int) ([]model.MembershipProduct, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND status = ?"
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

func (r *MySQLGrowthRepo) CreateMembershipOrder(userID string, productID string, payChannel string) (model.MembershipOrderAdmin, error) {
	var price float64
	var status string
	err := r.db.QueryRow("SELECT price, status FROM membership_products WHERE id = ?", productID).Scan(&price, &status)
	if err != nil {
		return model.MembershipOrderAdmin{}, err
	}
	if strings.ToUpper(status) != "ACTIVE" {
		return model.MembershipOrderAdmin{}, errors.New("membership product not active")
	}
	id := newID("mo")
	orderNo := id
	now := time.Now()
	_, err = r.db.Exec(`
INSERT INTO membership_orders (id, order_no, user_id, product_id, amount, pay_channel, status, paid_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, 'PENDING', NULL, ?, ?)`,
		id, orderNo, userID, productID, price, strings.ToUpper(payChannel), now, now,
	)
	if err != nil {
		return model.MembershipOrderAdmin{}, err
	}
	return model.MembershipOrderAdmin{
		ID:         id,
		OrderNo:    orderNo,
		UserID:     userID,
		ProductID:  productID,
		Amount:     price,
		PayChannel: strings.ToUpper(payChannel),
		Status:     "PENDING",
		CreatedAt:  now.Format(time.RFC3339),
	}, nil
}

func (r *MySQLGrowthRepo) ListMembershipOrders(userID string, status string, page int, pageSize int) ([]model.MembershipOrderAdmin, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{userID}
	filter := " WHERE user_id = ?"
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
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

func (r *MySQLGrowthRepo) TrackExperimentEvent(item model.ExperimentEvent) error {
	metadataJSON, err := marshalExperimentMetadata(item.Metadata)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(`
INSERT INTO experiment_events (
	id, experiment_key, variant_key, event_type, page_key, target_key, user_stage,
	anonymous_id, session_id, pathname, referrer, metadata_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("exp"),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.ExperimentKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.VariantKey)), 32),
		truncateByRunes(strings.ToUpper(strings.TrimSpace(item.EventType)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.PageKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.TargetKey)), 64),
		truncateByRunes(strings.ToUpper(strings.TrimSpace(item.UserStage)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.AnonymousID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.SessionID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Pathname)), 255),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Referrer)), 255),
		nullableString(metadataJSON),
		time.Now(),
	)
	return err
}

func (r *MySQLGrowthRepo) BindMembershipOrderExperiment(orderNo string, item model.ExperimentOrderAttribution) error {
	metadataJSON, err := marshalExperimentMetadata(item.Metadata)
	if err != nil {
		return err
	}
	conversionType := strings.ToUpper(strings.TrimSpace(item.ConversionType))
	if conversionType == "" {
		conversionType = deriveExperimentConversionType(item.UserStage)
	}
	_, err = r.db.Exec(`
INSERT INTO experiment_order_attributions (
	order_no, experiment_key, variant_key, page_key, target_key, user_stage,
	anonymous_id, session_id, pathname, referrer, conversion_type, metadata_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	experiment_key = VALUES(experiment_key),
	variant_key = VALUES(variant_key),
	page_key = VALUES(page_key),
	target_key = VALUES(target_key),
	user_stage = VALUES(user_stage),
	anonymous_id = VALUES(anonymous_id),
	session_id = VALUES(session_id),
	pathname = VALUES(pathname),
	referrer = VALUES(referrer),
	conversion_type = VALUES(conversion_type),
	metadata_json = VALUES(metadata_json)`,
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(orderNo)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.ExperimentKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.VariantKey)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.PageKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.TargetKey)), 64),
		truncateByRunes(strings.ToUpper(strings.TrimSpace(item.UserStage)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.AnonymousID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.SessionID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Pathname)), 255),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Referrer)), 255),
		truncateByRunes(conversionType, 32),
		nullableString(metadataJSON),
		time.Now(),
	)
	return err
}

func marshalExperimentMetadata(metadata map[string]interface{}) (string, error) {
	if len(metadata) == 0 {
		return "", nil
	}
	payloadBytes, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}
	return truncateByRunes(string(payloadBytes), 4096), nil
}

func deriveExperimentConversionType(userStage string) string {
	stage := strings.ToUpper(strings.TrimSpace(userStage))
	if stage == "VIP" || stage == "EXPIRED" {
		return "RENEWAL_SUCCESS"
	}
	return "PAYMENT_SUCCESS"
}

func insertExperimentEventTx(tx *sql.Tx, item model.ExperimentEvent, occurredAt time.Time) error {
	metadataJSON, err := marshalExperimentMetadata(item.Metadata)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`
INSERT INTO experiment_events (
	id, experiment_key, variant_key, event_type, page_key, target_key, user_stage,
	anonymous_id, session_id, pathname, referrer, metadata_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("exp"),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.ExperimentKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.VariantKey)), 32),
		truncateByRunes(strings.ToUpper(strings.TrimSpace(item.EventType)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.PageKey)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.TargetKey)), 64),
		truncateByRunes(strings.ToUpper(strings.TrimSpace(item.UserStage)), 32),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.AnonymousID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.SessionID)), 64),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Pathname)), 255),
		truncateByRunes(normalizeUTF8Text(strings.TrimSpace(item.Referrer)), 255),
		nullableString(metadataJSON),
		occurredAt,
	)
	return err
}

func createExperimentSuccessEventTx(tx *sql.Tx, orderNo string, occurredAt time.Time) error {
	var attribution model.ExperimentOrderAttribution
	var targetKey, userStage, anonymousID, sessionID, pathname, referrer, conversionType sql.NullString
	var metadataJSON sql.NullString
	err := tx.QueryRow(`
SELECT experiment_key, variant_key, page_key, target_key, user_stage, anonymous_id, session_id, pathname, referrer, conversion_type, metadata_json
FROM experiment_order_attributions
WHERE order_no = ?
LIMIT 1`, orderNo).Scan(
		&attribution.ExperimentKey,
		&attribution.VariantKey,
		&attribution.PageKey,
		&targetKey,
		&userStage,
		&anonymousID,
		&sessionID,
		&pathname,
		&referrer,
		&conversionType,
		&metadataJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if targetKey.Valid {
		attribution.TargetKey = targetKey.String
	}
	if userStage.Valid {
		attribution.UserStage = userStage.String
	}
	if anonymousID.Valid {
		attribution.AnonymousID = anonymousID.String
	}
	if sessionID.Valid {
		attribution.SessionID = sessionID.String
	}
	if pathname.Valid {
		attribution.Pathname = pathname.String
	}
	if referrer.Valid {
		attribution.Referrer = referrer.String
	}
	eventMetadata := map[string]interface{}{
		"order_no": orderNo,
	}
	if metadataJSON.Valid && strings.TrimSpace(metadataJSON.String) != "" {
		var extra map[string]interface{}
		if err := json.Unmarshal([]byte(metadataJSON.String), &extra); err == nil {
			for key, value := range extra {
				eventMetadata[key] = value
			}
		}
	}
	eventType := deriveExperimentConversionType(attribution.UserStage)
	if conversionType.Valid && strings.TrimSpace(conversionType.String) != "" {
		eventType = strings.ToUpper(strings.TrimSpace(conversionType.String))
	}
	sourceExperimentItems := []interface{}{}
	if rawItems, ok := eventMetadata["source_experiments"].([]interface{}); ok {
		sourceExperimentItems = rawItems
	}
	delete(eventMetadata, "source_experiments")
	primaryMetadata := map[string]interface{}{}
	for key, value := range eventMetadata {
		primaryMetadata[key] = value
	}
	primaryMetadata["attribution_role"] = "primary"
	if err := insertExperimentEventTx(tx, model.ExperimentEvent{
		ExperimentKey: attribution.ExperimentKey,
		VariantKey:    attribution.VariantKey,
		EventType:     eventType,
		PageKey:       attribution.PageKey,
		TargetKey:     attribution.TargetKey,
		UserStage:     attribution.UserStage,
		AnonymousID:   attribution.AnonymousID,
		SessionID:     attribution.SessionID,
		Pathname:      attribution.Pathname,
		Referrer:      attribution.Referrer,
		Metadata:      primaryMetadata,
	}, occurredAt); err != nil {
		return err
	}

	for _, rawItem := range sourceExperimentItems {
		itemMap, ok := rawItem.(map[string]interface{})
		if !ok {
			continue
		}
		readSourceField := func(key string) string {
			value, exists := itemMap[key]
			if !exists || value == nil {
				return ""
			}
			return strings.TrimSpace(fmt.Sprint(value))
		}
		sourceExperimentKey := readSourceField("experiment_key")
		sourceVariantKey := readSourceField("variant_key")
		sourcePageKey := readSourceField("page_key")
		if sourceExperimentKey == "" || sourceVariantKey == "" || sourcePageKey == "" {
			continue
		}
		sourceMetadata := map[string]interface{}{}
		for key, value := range eventMetadata {
			sourceMetadata[key] = value
		}
		sourceMetadata["attribution_role"] = "source"
		sourceMetadata["source_experiment_key"] = sourceExperimentKey
		if err := insertExperimentEventTx(tx, model.ExperimentEvent{
			ExperimentKey: sourceExperimentKey,
			VariantKey:    sourceVariantKey,
			EventType:     eventType,
			PageKey:       sourcePageKey,
			TargetKey:     readSourceField("target_key"),
			UserStage:     readSourceField("user_stage"),
			AnonymousID:   readSourceField("anonymous_id"),
			SessionID:     readSourceField("session_id"),
			Pathname:      readSourceField("pathname"),
			Referrer:      readSourceField("referrer"),
			Metadata:      sourceMetadata,
		}, occurredAt); err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminListMarketEvents(eventType string, symbol string, page int, pageSize int) ([]model.MarketEvent, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	trimmedType := strings.ToUpper(strings.TrimSpace(eventType))
	trimmedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	if trimmedType != "" {
		filter += " AND event_type = ?"
		args = append(args, trimmedType)
	}
	if trimmedSymbol != "" {
		filter += " AND symbol LIKE ?"
		args = append(args, "%"+trimmedSymbol+"%")
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM market_events"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, event_type, symbol, summary, trigger_rule, source, created_at
FROM market_events` + filter + `
ORDER BY created_at DESC, id DESC
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
		var summary, triggerRule, source sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.EventType, &item.Symbol, &summary, &triggerRule, &source, &createdAt); err != nil {
			return nil, 0, err
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
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateMarketEvent(item model.MarketEvent) (string, error) {
	id := newID("me")
	_, err := r.db.Exec(`
INSERT INTO market_events (id, event_type, symbol, summary, trigger_rule, source, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id,
		strings.ToUpper(strings.TrimSpace(item.EventType)),
		strings.ToUpper(strings.TrimSpace(item.Symbol)),
		strings.TrimSpace(item.Summary),
		strings.TrimSpace(item.TriggerRule),
		strings.TrimSpace(item.Source),
		time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateMarketEvent(id string, item model.MarketEvent) error {
	result, err := r.db.Exec(`
UPDATE market_events
SET event_type = ?, symbol = ?, summary = ?, trigger_rule = ?, source = ?
WHERE id = ?`,
		strings.ToUpper(strings.TrimSpace(item.EventType)),
		strings.ToUpper(strings.TrimSpace(item.Symbol)),
		strings.TrimSpace(item.Summary),
		strings.TrimSpace(item.TriggerRule),
		strings.TrimSpace(item.Source),
		id,
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

func (r *MySQLGrowthRepo) AdminListStockRecommendations(status string, page int, pageSize int) ([]model.StockRecommendation, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM stock_recommendations"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, source_type, strategy_version, reviewer, publisher, review_note, performance_label
FROM stock_recommendations` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.StockRecommendation, 0)
	for rows.Next() {
		var item model.StockRecommendation
		var pos, reason, sourceType, strategyVersion, reviewer, publisher, reviewNote, performanceLabel sql.NullString
		var vf, vt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.Symbol,
			&item.Name,
			&item.Score,
			&item.RiskLevel,
			&pos,
			&vf,
			&vt,
			&item.Status,
			&reason,
			&sourceType,
			&strategyVersion,
			&reviewer,
			&publisher,
			&reviewNote,
			&performanceLabel,
		); err != nil {
			return nil, 0, err
		}
		if pos.Valid {
			item.PositionRange = pos.String
		}
		if reason.Valid {
			item.ReasonSummary = reason.String
		}
		if sourceType.Valid {
			item.SourceType = sourceType.String
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
		item.ValidFrom = vf.Format(time.RFC3339)
		item.ValidTo = vt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateStockRecommendation(item model.StockRecommendation) (string, error) {
	id := newID("sr")
	vf, err := parseFlexibleDateTime(item.ValidFrom)
	if err != nil {
		return "", err
	}
	vt, err := parseFlexibleDateTime(item.ValidTo)
	if err != nil {
		return "", err
	}
	sourceType := strings.ToUpper(strings.TrimSpace(item.SourceType))
	if sourceType == "" {
		sourceType = "MANUAL"
	}
	strategyVersion := strings.TrimSpace(item.StrategyVersion)
	if strategyVersion == "" {
		strategyVersion = "manual-v1"
	}
	reviewer := strings.TrimSpace(item.Reviewer)
	publisher := strings.TrimSpace(item.Publisher)
	reviewNote := strings.TrimSpace(item.ReviewNote)
	performanceLabel := strings.ToUpper(strings.TrimSpace(item.PerformanceLabel))
	if performanceLabel == "" {
		performanceLabel = "PENDING"
	}
	_, err = r.db.Exec(`
INSERT INTO stock_recommendations (id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, source_type, strategy_version, reviewer, publisher, review_note, performance_label, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id,
		item.Symbol,
		item.Name,
		item.Score,
		item.RiskLevel,
		item.PositionRange,
		vf,
		vt,
		item.Status,
		item.ReasonSummary,
		sourceType,
		strategyVersion,
		reviewer,
		publisher,
		reviewNote,
		performanceLabel,
		time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateStockRecommendationStatus(id string, status string) error {
	targetStatus := normalizeStockRecommendationLifecycleStatus(status)
	if targetStatus == "" {
		return errors.New("invalid stock recommendation status")
	}

	var currentStatus string
	if err := r.db.QueryRow("SELECT status FROM stock_recommendations WHERE id = ?", id).Scan(&currentStatus); err != nil {
		return err
	}
	currentStatus = normalizeStockRecommendationLifecycleStatus(currentStatus)
	if currentStatus == "" {
		return errors.New("current stock recommendation status is empty")
	}
	if !canTransitionStockRecommendationStatus(currentStatus, targetStatus) {
		return fmt.Errorf(
			"invalid status transition: %s -> %s (allowed: %s)",
			currentStatus,
			targetStatus,
			strings.Join(allowedStockRecommendationTransitions(currentStatus), ", "),
		)
	}

	_, err := r.db.Exec("UPDATE stock_recommendations SET status = ? WHERE id = ?", targetStatus, id)
	return err
}

func (r *MySQLGrowthRepo) AdminGetQuantTopStocks(limit int, lookbackDays int) ([]model.StockQuantScore, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if lookbackDays <= 0 {
		lookbackDays = 120
	}
	if lookbackDays < 30 {
		lookbackDays = 30
	}
	if lookbackDays > 365 {
		lookbackDays = 365
	}

	sinceDate := time.Now().AddDate(0, 0, -(lookbackDays + 40))
	quotesBySymbol, err := r.loadStockQuotesBySymbol(sinceDate)
	if err != nil {
		return nil, err
	}
	items := make([]model.StockQuantScore, 0, len(quotesBySymbol))
	for symbol, quotes := range quotesBySymbol {
		item, ok := buildStockQuantScore(symbol, quotes)
		if !ok {
			continue
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return []model.StockQuantScore{}, nil
	}
	dailyBasicMap, err := r.loadLatestStockDailyBasics(sinceDate)
	if err != nil {
		return nil, err
	}
	moneyflowMap, err := r.loadLatestStockMoneyflows(sinceDate)
	if err != nil {
		return nil, err
	}
	newsSignalMap, err := r.loadStockNewsSignals(time.Now().AddDate(0, 0, -14))
	if err != nil {
		return nil, err
	}
	for index := range items {
		if basic, ok := dailyBasicMap[items[index].Symbol]; ok {
			items[index].PeTTM = basic.PeTTM
			items[index].PB = basic.PB
			items[index].TurnoverRate = basic.TurnoverRate
		}
		if flow, ok := moneyflowMap[items[index].Symbol]; ok {
			items[index].NetMFAmount = flow.NetMFAmount
		}
		if signal, ok := newsSignalMap[items[index].Symbol]; ok {
			items[index].NewsHeat = signal.Heat
			items[index].PositiveNewsRate = signal.PositiveRate
		}
	}

	momentum5Range := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.Momentum5 })
	momentum20Range := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.Momentum20 })
	volatilityRange := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.Volatility20 })
	volumeRatioRange := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.VolumeRatio })
	drawdownRange := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.Drawdown20 })
	trendRange := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.TrendStrength })
	peRange := buildMetricRangeFromValues(extractPositiveValues(items, func(item model.StockQuantScore) float64 { return item.PeTTM }))
	pbRange := buildMetricRangeFromValues(extractPositiveValues(items, func(item model.StockQuantScore) float64 { return item.PB }))
	turnoverRateRange := buildMetricRangeFromValues(extractPositiveValues(items, func(item model.StockQuantScore) float64 { return item.TurnoverRate }))
	netMoneyflowRange := buildMetricRange(items, func(item model.StockQuantScore) float64 { return item.NetMFAmount })
	newsHeatRange := buildMetricRangeFromValues(extractPositiveValues(items, func(item model.StockQuantScore) float64 { return float64(item.NewsHeat) }))

	for index := range items {
		m5Norm := normalizeMetric(items[index].Momentum5, momentum5Range, false)
		m20Norm := normalizeMetric(items[index].Momentum20, momentum20Range, false)
		volNorm := normalizeMetric(items[index].Volatility20, volatilityRange, true)
		volumeNorm := normalizeMetric(items[index].VolumeRatio, volumeRatioRange, false)
		drawdownNorm := normalizeMetric(items[index].Drawdown20, drawdownRange, true)
		trendNorm := normalizeMetric(items[index].TrendStrength, trendRange, false)
		trendScore := (m5Norm*0.25 + m20Norm*0.30 + volNorm*0.15 + drawdownNorm*0.15 + trendNorm*0.15) * 100

		netMoneyflowNorm := 0.5
		if math.Abs(items[index].NetMFAmount) > 0.001 {
			netMoneyflowNorm = normalizeMetric(items[index].NetMFAmount, netMoneyflowRange, false)
		}
		flowScore := (volumeNorm*0.42 + netMoneyflowNorm*0.58) * 100

		peNorm := 0.48
		if items[index].PeTTM > 0 {
			peNorm = normalizeMetric(items[index].PeTTM, peRange, true)
		}
		pbNorm := 0.5
		if items[index].PB > 0 {
			pbNorm = normalizeMetric(items[index].PB, pbRange, true)
		}
		turnoverNorm := 0.45
		if items[index].TurnoverRate > 0 {
			turnoverNorm = normalizeMetric(items[index].TurnoverRate, turnoverRateRange, false)
		}
		valueScore := (peNorm*0.45 + pbNorm*0.25 + turnoverNorm*0.30) * 100

		heatNorm := 0.45
		if items[index].NewsHeat > 0 {
			heatNorm = normalizeMetric(float64(items[index].NewsHeat), newsHeatRange, false)
		}
		positiveRate := items[index].PositiveNewsRate
		if positiveRate <= 0 {
			positiveRate = 0.5
		}
		newsScore := (heatNorm*0.45 + clampFloat(positiveRate, 0, 1)*0.55) * 100

		score := trendScore*0.45 + flowScore*0.25 + valueScore*0.20 + newsScore*0.10
		score -= quantRiskPenalty(items[index])
		score = clampFloat(score, 0, 100)
		items[index].TrendScore = roundTo(trendScore, 2)
		items[index].FlowScore = roundTo(flowScore, 2)
		items[index].ValueScore = roundTo(valueScore, 2)
		items[index].NewsScore = roundTo(newsScore, 2)
		items[index].Score = roundTo(score, 2)
		items[index].Momentum5 = roundTo(items[index].Momentum5, 2)
		items[index].Momentum20 = roundTo(items[index].Momentum20, 2)
		items[index].Volatility20 = roundTo(items[index].Volatility20, 2)
		items[index].VolumeRatio = roundTo(items[index].VolumeRatio, 2)
		items[index].Drawdown20 = roundTo(items[index].Drawdown20, 2)
		items[index].TrendStrength = roundTo(items[index].TrendStrength, 2)
		items[index].NetMFAmount = roundTo(items[index].NetMFAmount, 2)
		items[index].PeTTM = roundTo(items[index].PeTTM, 2)
		items[index].PB = roundTo(items[index].PB, 2)
		items[index].TurnoverRate = roundTo(items[index].TurnoverRate, 2)
		items[index].PositiveNewsRate = roundTo(items[index].PositiveNewsRate, 4)
		items[index].ClosePrice = roundTo(items[index].ClosePrice, 3)
	}
	filteredItems := make([]model.StockQuantScore, 0, len(items))
	for _, item := range items {
		if !passesQuantRiskGate(item) {
			continue
		}
		filteredItems = append(filteredItems, item)
	}
	if len(filteredItems) == 0 {
		filteredItems = items
	}
	items = filteredItems

	sort.Slice(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			if items[i].Momentum20 == items[j].Momentum20 {
				return items[i].Symbol < items[j].Symbol
			}
			return items[i].Momentum20 > items[j].Momentum20
		}
		return items[i].Score > items[j].Score
	})

	if len(items) > limit {
		items = selectDiversifiedQuantTop(items, limit)
	}

	symbols := make([]string, 0, len(items))
	for _, item := range items {
		symbols = append(symbols, item.Symbol)
	}
	nameMap, err := r.lookupStockNames(symbols)
	if err != nil {
		return nil, err
	}
	for index := range items {
		items[index].Rank = index + 1
		if name, ok := nameMap[items[index].Symbol]; ok && strings.TrimSpace(name) != "" {
			items[index].Name = name
		}
		if strings.TrimSpace(items[index].Name) == "" {
			items[index].Name = items[index].Symbol
		}
		items[index].Reasons = buildQuantReasons(items[index])
		items[index].ReasonSummary = buildQuantReasonSummary(items[index])
	}
	if err := r.persistStockQuantSnapshots(items); err != nil && !isTableNotFoundError(err) {
		return nil, err
	}

	return items, nil
}

func (r *MySQLGrowthRepo) AdminGetQuantEvaluation(windowDays int, topN int) (model.StockQuantEvaluationSummary, []model.StockQuantEvaluationPoint, []model.StockQuantRiskPerformance, []model.StockQuantRotationPoint, error) {
	if windowDays <= 0 {
		windowDays = 60
	}
	if windowDays < 20 {
		windowDays = 20
	}
	if windowDays > 365 {
		windowDays = 365
	}
	if topN <= 0 {
		topN = 10
	}
	if topN > 30 {
		topN = 30
	}
	sinceDate := time.Now().AddDate(0, 0, -windowDays)
	rankRows, err := r.db.Query(`
SELECT trade_date, symbol, rank_no
FROM stock_rank_daily
WHERE trade_date >= ? AND rank_no <= ?
ORDER BY trade_date ASC, rank_no ASC`, sinceDate.Format("2006-01-02"), topN)
	if err != nil {
		if isTableNotFoundError(err) {
			return model.StockQuantEvaluationSummary{
				WindowDays:  windowDays,
				TopN:        topN,
				GeneratedAt: time.Now().Format(time.RFC3339),
			}, []model.StockQuantEvaluationPoint{}, []model.StockQuantRiskPerformance{}, []model.StockQuantRotationPoint{}, nil
		}
		return model.StockQuantEvaluationSummary{}, nil, nil, nil, err
	}
	defer rankRows.Close()

	type rankItem struct {
		TradeDate time.Time
		Symbol    string
		RankNo    int
	}
	ranks := make([]rankItem, 0, windowDays*topN)
	for rankRows.Next() {
		var (
			item   rankItem
			symbol string
		)
		if err := rankRows.Scan(&item.TradeDate, &symbol, &item.RankNo); err != nil {
			return model.StockQuantEvaluationSummary{}, nil, nil, nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if item.Symbol == "" {
			continue
		}
		ranks = append(ranks, item)
	}
	if len(ranks) == 0 {
		return model.StockQuantEvaluationSummary{
			WindowDays:  windowDays,
			TopN:        topN,
			GeneratedAt: time.Now().Format(time.RFC3339),
		}, []model.StockQuantEvaluationPoint{}, []model.StockQuantRiskPerformance{}, []model.StockQuantRotationPoint{}, nil
	}

	quoteSince := sinceDate.AddDate(0, 0, -20)
	quotesBySymbol, err := r.loadStockQuotesBySymbol(quoteSince)
	if err != nil {
		return model.StockQuantEvaluationSummary{}, nil, nil, nil, err
	}
	benchmarkSeries, err := r.loadBenchmarkQuoteSeries(quoteSince)
	if err != nil {
		return model.StockQuantEvaluationSummary{}, nil, nil, nil, err
	}
	riskLevelMap, err := r.loadQuantRiskLevelByDateSymbol(sinceDate, topN)
	if err != nil {
		return model.StockQuantEvaluationSummary{}, nil, nil, nil, err
	}

	type dayAggregate struct {
		Sum5           float64
		Cnt5           int
		Hit5           int
		Sum10          float64
		Cnt10          int
		Hit10          int
		Benchmark5Sum  float64
		Benchmark5Cnt  int
		Benchmark10Sum float64
		Benchmark10Cnt int
	}
	byDay := make(map[string]*dayAggregate)
	riskAggMap := make(map[string]*quantRiskAggregate)
	dayTopSymbols := make(map[string][]string)
	dayTopSeen := make(map[string]map[string]struct{})
	for _, item := range ranks {
		symbolQuotes, ok := quotesBySymbol[item.Symbol]
		if !ok || len(symbolQuotes) == 0 {
			continue
		}
		dayKey := item.TradeDate.Format("2006-01-02")
		if _, ok := dayTopSeen[dayKey]; !ok {
			dayTopSeen[dayKey] = make(map[string]struct{}, topN)
		}
		if _, exists := dayTopSeen[dayKey][item.Symbol]; !exists {
			dayTopSeen[dayKey][item.Symbol] = struct{}{}
			dayTopSymbols[dayKey] = append(dayTopSymbols[dayKey], item.Symbol)
		}

		agg := byDay[dayKey]
		if agg == nil {
			agg = &dayAggregate{}
			byDay[dayKey] = agg
		}
		riskLevel := resolveQuantRiskLevel(riskLevelMap[dayKey+"|"+item.Symbol])
		riskAgg := riskAggMap[riskLevel]
		if riskAgg == nil {
			riskAgg = &quantRiskAggregate{}
			riskAggMap[riskLevel] = riskAgg
		}
		if ret5, ok := calcForwardReturn(symbolQuotes, item.TradeDate, 5); ok {
			agg.Sum5 += ret5
			agg.Cnt5++
			if ret5 > 0 {
				agg.Hit5++
			}
			riskAgg.Sum5 += ret5
			riskAgg.Cnt5++
			if ret5 > 0 {
				riskAgg.Hit5++
			}
		}
		if ret10, ok := calcForwardReturn(symbolQuotes, item.TradeDate, 10); ok {
			agg.Sum10 += ret10
			agg.Cnt10++
			if ret10 > 0 {
				agg.Hit10++
			}
			riskAgg.Sum10 += ret10
			riskAgg.Cnt10++
			if ret10 > 0 {
				riskAgg.Hit10++
			}
		}
		if benchmark5, ok := calcForwardReturn(benchmarkSeries, item.TradeDate, 5); ok {
			agg.Benchmark5Sum += benchmark5
			agg.Benchmark5Cnt++
		}
		if benchmark10, ok := calcForwardReturn(benchmarkSeries, item.TradeDate, 10); ok {
			agg.Benchmark10Sum += benchmark10
			agg.Benchmark10Cnt++
		}
	}

	tradeDates := make([]string, 0, len(byDay))
	for day := range byDay {
		tradeDates = append(tradeDates, day)
	}
	sort.Strings(tradeDates)
	points := make([]model.StockQuantEvaluationPoint, 0, len(tradeDates))
	summary := model.StockQuantEvaluationSummary{
		WindowDays:  windowDays,
		TopN:        topN,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}
	equityReturns5 := make([]float64, 0, len(tradeDates))
	equityReturns10 := make([]float64, 0, len(tradeDates))
	benchmarkReturns5 := make([]float64, 0, len(tradeDates))
	benchmarkReturns10 := make([]float64, 0, len(tradeDates))
	for _, tradeDate := range tradeDates {
		agg := byDay[tradeDate]
		point := model.StockQuantEvaluationPoint{TradeDate: tradeDate}
		if agg.Cnt5 > 0 {
			point.AvgReturn5 = roundTo(agg.Sum5/float64(agg.Cnt5), 4)
			point.HitRate5 = roundTo(float64(agg.Hit5)/float64(agg.Cnt5), 4)
			if agg.Benchmark5Cnt > 0 {
				point.BenchmarkReturn = roundTo(agg.Benchmark5Sum/float64(agg.Benchmark5Cnt), 4)
			}
			point.SampleCount = agg.Cnt5
			equityReturns5 = append(equityReturns5, point.AvgReturn5)
			benchmarkReturns5 = append(benchmarkReturns5, point.BenchmarkReturn)
			summary.AvgReturn5 += point.AvgReturn5
			summary.HitRate5 += point.HitRate5
			summary.SampleCount += agg.Cnt5
		}
		if agg.Cnt10 > 0 {
			point.AvgReturn10 = roundTo(agg.Sum10/float64(agg.Cnt10), 4)
			point.HitRate10 = roundTo(float64(agg.Hit10)/float64(agg.Cnt10), 4)
			equityReturns10 = append(equityReturns10, point.AvgReturn10)
			if agg.Benchmark10Cnt > 0 {
				point.BenchmarkReturn10 = roundTo(agg.Benchmark10Sum/float64(agg.Benchmark10Cnt), 4)
				benchmarkReturns10 = append(benchmarkReturns10, point.BenchmarkReturn10)
			}
			summary.AvgReturn10 += point.AvgReturn10
			summary.HitRate10 += point.HitRate10
		}
		points = append(points, point)
	}
	summary.SampleDays = len(points)
	if len(points) > 0 {
		if len(equityReturns5) > 0 {
			summary.AvgReturn5 = roundTo(summary.AvgReturn5/float64(len(equityReturns5)), 4)
			summary.HitRate5 = roundTo(summary.HitRate5/float64(len(equityReturns5)), 4)
		}
		if len(equityReturns10) > 0 {
			summary.AvgReturn10 = roundTo(summary.AvgReturn10/float64(len(equityReturns10)), 4)
			summary.HitRate10 = roundTo(summary.HitRate10/float64(len(equityReturns10)), 4)
		}
		summary.MaxDrawdown5 = roundTo(calcMaxDrawdownFromReturns(equityReturns5), 4)
		summary.MaxDrawdown10 = roundTo(calcMaxDrawdownFromReturns(equityReturns10), 4)
		summary.BenchmarkAvgReturn5 = roundTo(avgFloat(benchmarkReturns5), 4)
		summary.BenchmarkAvgReturn10 = roundTo(avgFloat(benchmarkReturns10), 4)
	}
	cum5 := 1.0
	cumBench5 := 1.0
	cum10 := 1.0
	cumBench10 := 1.0
	for index := range points {
		cum5 *= 1 + points[index].AvgReturn5
		cumBench5 *= 1 + points[index].BenchmarkReturn
		points[index].CumulativeReturn5 = roundTo(cum5-1, 4)
		points[index].CumulativeBenchmark5 = roundTo(cumBench5-1, 4)
		points[index].CumulativeExcess5 = roundTo(points[index].CumulativeReturn5-points[index].CumulativeBenchmark5, 4)
		cum10 *= 1 + points[index].AvgReturn10
		cumBench10 *= 1 + points[index].BenchmarkReturn10
		points[index].CumulativeReturn10 = roundTo(cum10-1, 4)
		points[index].CumulativeBenchmark10 = roundTo(cumBench10-1, 4)
		points[index].CumulativeExcess10 = roundTo(points[index].CumulativeReturn10-points[index].CumulativeBenchmark10, 4)
	}
	riskItems := buildQuantRiskPerformanceItems(riskAggMap)
	rotationItems := buildQuantRotationItems(tradeDates, dayTopSymbols)
	return summary, points, riskItems, rotationItems, nil
}

func (r *MySQLGrowthRepo) AdminGenerateDailyStockRecommendations(tradeDate string) (model.AdminDailyStockRecommendationGenerationResult, error) {
	if r.strategyEngine != nil {
		return r.generateDailyStockRecommendationsViaStrategyEngine(tradeDate)
	}
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	start, err := time.ParseInLocation("2006-01-02", tradeDate, time.Local)
	if err != nil {
		return model.AdminDailyStockRecommendationGenerationResult{}, err
	}
	end := start.Add(24 * time.Hour)
	quantItems, quantErr := r.AdminGetQuantTopStocks(10, 180)
	candidates := buildDailyStockCandidatesFromQuant(quantItems, start, end)
	if quantErr != nil || len(candidates) == 0 {
		candidates = buildDefaultDailyStockCandidates(start, end)
	}

	count := 0
	for _, candidate := range candidates {
		id := newID("sr")
		_, err := r.db.Exec(`
INSERT INTO stock_recommendations (id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, source_type, strategy_version, reviewer, publisher, review_note, performance_label, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id,
			candidate.Symbol,
			candidate.Name,
			candidate.Score,
			candidate.RiskLevel,
			candidate.PositionRange,
			candidate.ValidFrom,
			candidate.ValidTo,
			candidate.Status,
			candidate.ReasonSummary,
			"SYSTEM",
			"daily-v1",
			"",
			"system",
			"由每日量化流水线自动生成，待运营跟踪。",
			"ESTIMATED",
			time.Now(),
		)
		if err != nil {
			return model.AdminDailyStockRecommendationGenerationResult{}, err
		}
		_, _ = r.db.Exec(`
INSERT INTO stock_reco_details (reco_id, tech_score, fund_score, sentiment_score, money_flow_score, take_profit, stop_loss, risk_note)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE tech_score=VALUES(tech_score), fund_score=VALUES(fund_score), sentiment_score=VALUES(sentiment_score), money_flow_score=VALUES(money_flow_score), take_profit=VALUES(take_profit), stop_loss=VALUES(stop_loss), risk_note=VALUES(risk_note)`,
			id,
			candidate.TechScore,
			candidate.FundScore,
			candidate.SentimentScore,
			candidate.MoneyFlowScore,
			candidate.TakeProfit,
			candidate.StopLoss,
			candidate.RiskNote,
		)
		count++
	}
	return model.AdminDailyStockRecommendationGenerationResult{
		Count:          count,
		GenerationMode: "FALLBACK",
		ArchiveEnabled: false,
	}, nil
}

func (r *MySQLGrowthRepo) AdminGenerateDailyFuturesStrategies(tradeDate string) (model.AdminDailyFuturesStrategyGenerationResult, error) {
	if r.strategyEngine != nil {
		return r.generateDailyFuturesStrategiesViaStrategyEngine(tradeDate)
	}
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	start, err := time.ParseInLocation("2006-01-02", tradeDate, time.Local)
	if err != nil {
		return model.AdminDailyFuturesStrategyGenerationResult{}, err
	}
	end := start.Add(24 * time.Hour)
	samples := []model.FuturesStrategy{
		{Contract: "IF2603", Name: "股指趋势", Direction: "LONG", RiskLevel: "MEDIUM", PositionRange: "20%-30%", Status: "PUBLISHED", ReasonSummary: "趋势与量价结构一致"},
		{Contract: "IH2603", Name: "蓝筹回归", Direction: "LONG", RiskLevel: "LOW", PositionRange: "15%-25%", Status: "PUBLISHED", ReasonSummary: "估值回归驱动"},
		{Contract: "IC2603", Name: "中证成长", Direction: "SHORT", RiskLevel: "HIGH", PositionRange: "10%-20%", Status: "PUBLISHED", ReasonSummary: "风格切换压力"},
	}
	count := 0
	for _, s := range samples {
		id := newID("fs")
		_, err := r.db.Exec(`
INSERT INTO futures_strategies (id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, s.Contract, s.Name, s.Direction, s.RiskLevel, s.PositionRange, start, end, s.Status, s.ReasonSummary,
		)
		if err != nil {
			return model.AdminDailyFuturesStrategyGenerationResult{}, err
		}
		count++
	}
	return model.AdminDailyFuturesStrategyGenerationResult{
		Count:          count,
		GenerationMode: "FALLBACK",
		ArchiveEnabled: false,
	}, nil
}

func (r *MySQLGrowthRepo) AdminListFuturesStrategies(status string, contract string, page int, pageSize int) ([]model.FuturesStrategy, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	contract = strings.TrimSpace(contract)
	status = strings.TrimSpace(status)
	if contract != "" {
		filter += " AND contract = ?"
		args = append(args, contract)
	}
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM futures_strategies"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary
FROM futures_strategies` + filter + `
ORDER BY valid_from DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.FuturesStrategy, 0)
	for rows.Next() {
		var item model.FuturesStrategy
		var name, positionRange, reasonSummary sql.NullString
		var validFrom, validTo time.Time
		if err := rows.Scan(&item.ID, &item.Contract, &name, &item.Direction, &item.RiskLevel, &positionRange, &validFrom, &validTo, &item.Status, &reasonSummary); err != nil {
			return nil, 0, err
		}
		if name.Valid {
			item.Name = name.String
		}
		if positionRange.Valid {
			item.PositionRange = positionRange.String
		}
		if reasonSummary.Valid {
			item.ReasonSummary = reasonSummary.String
		}
		item.ValidFrom = validFrom.Format(time.RFC3339)
		item.ValidTo = validTo.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateFuturesStrategy(item model.FuturesStrategy) (string, error) {
	id := newID("fs")
	vf, err := parseFlexibleDateTime(item.ValidFrom)
	if err != nil {
		return "", err
	}
	vt, err := parseFlexibleDateTime(item.ValidTo)
	if err != nil {
		return "", err
	}
	_, err = r.db.Exec(`
INSERT INTO futures_strategies (id, contract, name, direction, risk_level, position_range, valid_from, valid_to, status, reason_summary)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, item.Contract, item.Name, item.Direction, item.RiskLevel, item.PositionRange, vf, vt, item.Status, item.ReasonSummary,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateFuturesStrategyStatus(id string, status string) error {
	_, err := r.db.Exec("UPDATE futures_strategies SET status = ? WHERE id = ?", status, id)
	return err
}

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

func (r *MySQLGrowthRepo) AdminListUsers(status string, kycStatus string, memberLevel string, registrationSource string, page int, pageSize int) ([]model.AdminUser, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	status = strings.ToUpper(strings.TrimSpace(status))
	kycStatus = strings.ToUpper(strings.TrimSpace(kycStatus))
	memberLevel = strings.TrimSpace(memberLevel)
	registrationSource = strings.ToUpper(strings.TrimSpace(registrationSource))
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND u.status = ?"
		args = append(args, status)
	}
	if kycStatus != "" {
		filter += " AND u.kyc_status = ?"
		args = append(args, kycStatus)
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
		item.ActivationState = resolveMembershipActivationState(item.MemberLevel, item.KYCStatus, vipExpireAt)
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminGetUserSourceSummary(status string, kycStatus string, memberLevel string, registrationSource string) (model.AdminUserSourceSummary, error) {
	result := model.AdminUserSourceSummary{}
	args := []interface{}{}
	status = strings.ToUpper(strings.TrimSpace(status))
	kycStatus = strings.ToUpper(strings.TrimSpace(kycStatus))
	memberLevel = strings.TrimSpace(memberLevel)
	registrationSource = strings.ToUpper(strings.TrimSpace(registrationSource))

	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND u.status = ?"
		args = append(args, status)
	}
	if kycStatus != "" {
		filter += " AND u.kyc_status = ?"
		args = append(args, kycStatus)
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

func (r *MySQLGrowthRepo) AdminUpdateUserKYCStatus(id string, kycStatus string) error {
	_, err := r.db.Exec("UPDATE users SET kyc_status = ?, updated_at = ? WHERE id = ?", kycStatus, time.Now(), id)
	return err
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
	if err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE kyc_status = 'APPROVED'").Scan(&result.KYCApprovedUsers); err != nil {
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
SELECT id, member_level, doc_read_limit, news_subscribe_limit, reset_cycle, status, effective_at, updated_at
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
		if err := rows.Scan(&item.ID, &item.MemberLevel, &item.DocReadLimit, &item.NewsSubscribeLimit, &item.ResetCycle, &item.Status, &effectiveAt, &updatedAt); err != nil {
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
INSERT INTO vip_quota_configs (id, member_level, doc_read_limit, news_subscribe_limit, reset_cycle, status, effective_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, item.MemberLevel, item.DocReadLimit, item.NewsSubscribeLimit, item.ResetCycle, item.Status, effectiveAt, time.Now(),
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
SET doc_read_limit = ?, news_subscribe_limit = ?, reset_cycle = ?, status = ?, effective_at = ?, updated_at = ?
WHERE id = ?`,
		item.DocReadLimit, item.NewsSubscribeLimit, item.ResetCycle, item.Status, effectiveAt, time.Now(), id,
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
SELECT u.id, u.member_level, ?, COALESCE(vqc.doc_read_limit, 0), COALESCE(uqu.doc_read_used, 0),
       COALESCE(vqc.news_subscribe_limit, 0), COALESCE(uqu.news_subscribe_used, 0), uqu.updated_at
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
			&item.UserID,
			&item.MemberLevel,
			&item.PeriodKey,
			&item.DocReadLimit,
			&item.DocReadUsed,
			&item.NewsSubscribeLimit,
			&item.NewsSubscribeUsed,
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

func (r *MySQLGrowthRepo) AdminAdjustUserQuota(userID string, periodKey string, docReadDelta int, newsSubscribeDelta int) error {
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
INSERT INTO user_quota_usages (id, user_id, member_level, period_key, doc_read_used, news_subscribe_used, updated_at)
VALUES (?, ?, ?, ?, 0, 0, ?)
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
	updated_at = ?
WHERE user_id = ? AND period_key = ?`,
		docReadDelta, newsSubscribeDelta, now, userID, periodKey,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *MySQLGrowthRepo) AdminListDataSources(page int, pageSize int) ([]model.DataSource, int, error) {
	if err := r.ensureBuiltinDataSources(); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM system_configs WHERE config_key LIKE 'data_source.%'").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT id, config_key, config_value, description, updated_at
FROM system_configs
WHERE config_key LIKE 'data_source.%'
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.DataSource, 0)
	for rows.Next() {
		var id, configKey, configValue string
		var desc sql.NullString
		var updatedAt time.Time
		if err := rows.Scan(&id, &configKey, &configValue, &desc, &updatedAt); err != nil {
			return nil, 0, err
		}
		item := model.DataSource{
			ID:         id,
			SourceKey:  strings.TrimPrefix(configKey, "data_source."),
			Name:       "",
			SourceType: "",
			Status:     "ACTIVE",
			UpdatedAt:  updatedAt.Format(time.RFC3339),
		}
		var payload struct {
			Name       string                 `json:"name"`
			SourceType string                 `json:"source_type"`
			Status     string                 `json:"status"`
			Config     map[string]interface{} `json:"config"`
		}
		if err := json.Unmarshal([]byte(configValue), &payload); err == nil {
			item.Name = payload.Name
			item.SourceType = payload.SourceType
			if strings.TrimSpace(payload.Status) != "" {
				item.Status = payload.Status
			}
			item.Config = payload.Config
		}
		if item.Name == "" {
			if desc.Valid {
				item.Name = desc.String
			} else {
				item.Name = item.SourceKey
			}
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateDataSource(item model.DataSource) (string, error) {
	sourceKey := strings.TrimSpace(item.SourceKey)
	if sourceKey == "" {
		return "", errors.New("source_key is required")
	}
	configKey := "data_source." + sourceKey
	var exists int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM system_configs WHERE config_key = ?", configKey).Scan(&exists); err != nil {
		return "", err
	}
	if exists > 0 {
		return "", errors.New("data source already exists")
	}

	payloadBytes, err := json.Marshal(map[string]interface{}{
		"name":        item.Name,
		"source_type": strings.ToUpper(strings.TrimSpace(item.SourceType)),
		"status":      strings.ToUpper(strings.TrimSpace(item.Status)),
		"config":      item.Config,
	})
	if err != nil {
		return "", err
	}

	id := newID("cfg")
	_, err = r.db.Exec(`
INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		id, configKey, string(payloadBytes), item.Name, "admin", time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func cloneDataSourceConfigMap(value map[string]interface{}) map[string]interface{} {
	if len(value) == 0 {
		return map[string]interface{}{}
	}
	clone := make(map[string]interface{}, len(value))
	for key, item := range value {
		clone[key] = item
	}
	return clone
}

func mergeDataSourceConfigMap(existing map[string]interface{}, incoming map[string]interface{}) map[string]interface{} {
	merged := cloneDataSourceConfigMap(existing)
	for key, value := range incoming {
		merged[key] = value
	}
	return merged
}

func (r *MySQLGrowthRepo) AdminUpdateDataSource(sourceKey string, item model.DataSource) error {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return sql.ErrNoRows
	}
	configKey := "data_source." + sourceKey
	var existingConfigValue string
	if err := r.db.QueryRow("SELECT config_value FROM system_configs WHERE config_key = ?", configKey).Scan(&existingConfigValue); err != nil {
		return err
	}
	var existingPayload struct {
		Name       string                 `json:"name"`
		SourceType string                 `json:"source_type"`
		Status     string                 `json:"status"`
		Config     map[string]interface{} `json:"config"`
	}
	if err := json.Unmarshal([]byte(existingConfigValue), &existingPayload); err != nil {
		existingPayload.Config = map[string]interface{}{}
	}
	name := strings.TrimSpace(item.Name)
	if name == "" {
		name = strings.TrimSpace(existingPayload.Name)
	}
	sourceType := strings.ToUpper(strings.TrimSpace(item.SourceType))
	if sourceType == "" {
		sourceType = strings.ToUpper(strings.TrimSpace(existingPayload.SourceType))
	}
	status := strings.ToUpper(strings.TrimSpace(item.Status))
	if status == "" {
		status = strings.ToUpper(strings.TrimSpace(existingPayload.Status))
	}
	if status == "" {
		status = "ACTIVE"
	}
	mergedConfig := mergeDataSourceConfigMap(existingPayload.Config, item.Config)
	payloadBytes, err := json.Marshal(map[string]interface{}{
		"name":        name,
		"source_type": sourceType,
		"status":      status,
		"config":      mergedConfig,
	})
	if err != nil {
		return err
	}

	result, err := r.db.Exec(`
UPDATE system_configs
SET config_value = ?, description = ?, updated_by = ?, updated_at = ?
WHERE config_key = ?`,
		string(payloadBytes), name, "admin", time.Now(), configKey,
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

func (r *MySQLGrowthRepo) AdminDeleteDataSource(sourceKey string) error {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return sql.ErrNoRows
	}
	configKey := "data_source." + sourceKey
	result, err := r.db.Exec("DELETE FROM system_configs WHERE config_key = ?", configKey)
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

func (r *MySQLGrowthRepo) AdminCheckDataSourceHealth(sourceKey string) (model.DataSourceHealthCheck, error) {
	item, err := r.getDataSourceBySourceKey(sourceKey)
	if err != nil {
		return model.DataSourceHealthCheck{}, err
	}
	checkedAt := time.Now()
	result := model.DataSourceHealthCheck{
		SourceKey: item.SourceKey,
		Status:    "UNKNOWN",
		CheckedAt: checkedAt.Format(time.RFC3339),
	}
	previousStatus, err := r.getLatestDataSourceHealthStatus(item.SourceKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return model.DataSourceHealthCheck{}, err
	}

	if strings.ToUpper(strings.TrimSpace(item.Status)) != "ACTIVE" {
		result.Message = "data source is disabled"
		result.FailureCategory = "DISABLED"
	} else {
		provider := strings.ToUpper(parseDataSourceStringConfig(item.Config, "provider", "vendor"))
		isTushare := strings.EqualFold(strings.TrimSpace(item.SourceKey), "TUSHARE") || provider == "TUSHARE"
		isAkshare := strings.EqualFold(strings.TrimSpace(item.SourceKey), "AKSHARE") || provider == "AKSHARE"
		endpoint := parseDataSourceStringConfig(item.Config, "endpoint", "quotes_endpoint")
		if isTushare && strings.TrimSpace(endpoint) == "" {
			endpoint = "https://api.tushare.pro"
		}
		timeoutMS := parseDataSourceTimeoutMS(item.Config)
		retryTimes := parseDataSourceRetryTimes(item.Config)
		retryIntervalMS := parseDataSourceRetryIntervalMS(item.Config)
		result.MaxAttempts = retryTimes + 1

		if isAkshare {
			for attempt := 1; attempt <= result.MaxAttempts; attempt++ {
				result.Attempts = attempt
				attemptResult := performAkshareDataSourceHealthCheckAttempt(item.Config, timeoutMS)
				result.Status = attemptResult.Status
				result.Reachable = attemptResult.Reachable
				result.HTTPStatus = attemptResult.HTTPStatus
				result.LatencyMS = attemptResult.LatencyMS
				result.Message = attemptResult.Message
				result.FailureCategory = attemptResult.FailureCategory
				if result.Status == "HEALTHY" || result.Status == "UNKNOWN" {
					break
				}
				if attempt < result.MaxAttempts && retryIntervalMS > 0 {
					time.Sleep(time.Duration(retryIntervalMS) * time.Millisecond)
				}
			}
		} else if endpoint == "" {
			result.Message = "endpoint not configured"
			result.FailureCategory = "CONFIG_ERROR"
		} else {
			parsed, parseErr := url.Parse(endpoint)
			if parseErr != nil || parsed.Scheme == "" || parsed.Host == "" {
				result.Message = "invalid endpoint"
				result.FailureCategory = "CONFIG_ERROR"
			} else if parsed.Scheme != "http" && parsed.Scheme != "https" {
				result.Message = "unsupported endpoint scheme"
				result.FailureCategory = "CONFIG_ERROR"
			} else {
				tushareToken := ""
				if isTushare {
					tushareToken = parseDataSourceStringConfig(item.Config, "token", "api_token", "tushare_token")
					if strings.TrimSpace(tushareToken) == "" {
						tushareToken = strings.TrimSpace(os.Getenv("TUSHARE_TOKEN"))
					}
					if strings.TrimSpace(tushareToken) == "" {
						result.Status = "UNHEALTHY"
						result.FailureCategory = "CONFIG_ERROR"
						result.Message = "tushare token not configured"
					}
				}

				for attempt := 1; attempt <= result.MaxAttempts; attempt++ {
					if isTushare && strings.TrimSpace(tushareToken) == "" {
						break
					}
					result.Attempts = attempt
					attemptResult := model.DataSourceHealthCheck{}
					if isTushare {
						attemptResult = performTushareDataSourceHealthCheckAttempt(endpoint, tushareToken, timeoutMS)
					} else {
						attemptResult = performDataSourceHealthCheckAttempt(endpoint, timeoutMS)
					}
					result.Status = attemptResult.Status
					result.Reachable = attemptResult.Reachable
					result.HTTPStatus = attemptResult.HTTPStatus
					result.LatencyMS = attemptResult.LatencyMS
					result.Message = attemptResult.Message
					result.FailureCategory = attemptResult.FailureCategory
					if result.Status == "HEALTHY" || result.Status == "UNKNOWN" {
						break
					}
					if attempt < result.MaxAttempts && retryIntervalMS > 0 {
						time.Sleep(time.Duration(retryIntervalMS) * time.Millisecond)
					}
				}
			}
		}
	}

	if err := r.persistDataSourceHealthLog(result, checkedAt); err != nil {
		return model.DataSourceHealthCheck{}, err
	}

	if result.Status == "UNHEALTHY" {
		consecutiveFailures, err := r.countConsecutiveDataSourceFailures(item.SourceKey, 50)
		if err != nil {
			return model.DataSourceHealthCheck{}, err
		}
		result.ConsecutiveFailures = consecutiveFailures
		threshold := parseDataSourceFailThreshold(item.Config)
		result.AlertTriggered = consecutiveFailures >= threshold
		if result.AlertTriggered && consecutiveFailures == threshold {
			r.createDataSourceHealthWorkflowMessage(item, result, "UNHEALTHY")
		}
	}
	if result.Status == "HEALTHY" && previousStatus == "UNHEALTHY" {
		r.createDataSourceHealthWorkflowMessage(item, result, "RECOVERED")
	}
	return result, nil
}

func (r *MySQLGrowthRepo) AdminBatchCheckDataSourceHealth(sourceKeys []string) ([]model.DataSourceHealthCheck, error) {
	targets := make([]string, 0)
	seen := make(map[string]struct{})
	if len(sourceKeys) == 0 {
		rows, err := r.db.Query(`
SELECT config_key
FROM system_configs
WHERE config_key LIKE 'data_source.%'
ORDER BY updated_at DESC`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		for rows.Next() {
			var configKey string
			if err := rows.Scan(&configKey); err != nil {
				return nil, err
			}
			sourceKey := strings.TrimSpace(strings.TrimPrefix(configKey, "data_source."))
			if sourceKey == "" {
				continue
			}
			if _, ok := seen[sourceKey]; ok {
				continue
			}
			seen[sourceKey] = struct{}{}
			targets = append(targets, sourceKey)
		}
	} else {
		for _, sourceKey := range sourceKeys {
			trimmed := strings.TrimSpace(sourceKey)
			if trimmed == "" {
				continue
			}
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
			targets = append(targets, trimmed)
		}
	}

	items := make([]model.DataSourceHealthCheck, 0, len(targets))
	for _, sourceKey := range targets {
		item, err := r.AdminCheckDataSourceHealth(sourceKey)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				items = append(items, model.DataSourceHealthCheck{
					SourceKey: sourceKey,
					Status:    "UNKNOWN",
					Message:   "data source not found",
					CheckedAt: time.Now().Format(time.RFC3339),
				})
				continue
			}
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *MySQLGrowthRepo) AdminListDataSourceHealthLogs(sourceKey string, page int, pageSize int) ([]model.DataSourceHealthLog, int, error) {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return nil, 0, sql.ErrNoRows
	}
	if _, err := r.getDataSourceBySourceKey(sourceKey); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM data_source_health_logs WHERE source_key = ?", sourceKey).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(`
SELECT id, source_key, status, reachable, http_status, latency_ms, message, checked_at
FROM data_source_health_logs
WHERE source_key = ?
ORDER BY checked_at DESC, id DESC
LIMIT ? OFFSET ?`, sourceKey, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.DataSourceHealthLog, 0)
	for rows.Next() {
		var item model.DataSourceHealthLog
		var reachableInt int
		var httpStatus sql.NullInt64
		var message sql.NullString
		var checkedAt time.Time
		if err := rows.Scan(&item.ID, &item.SourceKey, &item.Status, &reachableInt, &httpStatus, &item.LatencyMS, &message, &checkedAt); err != nil {
			return nil, 0, err
		}
		item.Reachable = reachableInt == 1
		if httpStatus.Valid {
			item.HTTPStatus = int(httpStatus.Int64)
		}
		if message.Valid {
			item.Message = message.String
		}
		item.CheckedAt = checkedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) getDataSourceBySourceKey(sourceKey string) (model.DataSource, error) {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return model.DataSource{}, sql.ErrNoRows
	}
	if err := r.ensureBuiltinDataSources(); err != nil {
		return model.DataSource{}, err
	}
	candidates := buildDataSourceLookupCandidates(sourceKey)
	seen := make(map[string]struct{})
	for _, candidate := range candidates {
		trimmedCandidate := strings.TrimSpace(candidate)
		if trimmedCandidate == "" {
			continue
		}
		lookupKey := strings.ToLower(trimmedCandidate)
		if _, exists := seen[lookupKey]; exists {
			continue
		}
		seen[lookupKey] = struct{}{}

		configKey := "data_source." + trimmedCandidate
		var id, configValue string
		var desc sql.NullString
		var updatedAt time.Time
		var storedConfigKey string
		err := r.db.QueryRow(`
SELECT id, config_key, config_value, description, updated_at
FROM system_configs
WHERE LOWER(config_key) = LOWER(?)
LIMIT 1`, configKey).Scan(&id, &storedConfigKey, &configValue, &desc, &updatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			return model.DataSource{}, err
		}
		actualSourceKey := strings.TrimSpace(strings.TrimPrefix(storedConfigKey, "data_source."))
		if actualSourceKey == "" {
			actualSourceKey = trimmedCandidate
		}
		item := model.DataSource{
			ID:         id,
			SourceKey:  actualSourceKey,
			Name:       "",
			SourceType: "",
			Status:     "ACTIVE",
			UpdatedAt:  updatedAt.Format(time.RFC3339),
		}
		var payload struct {
			Name       string                 `json:"name"`
			SourceType string                 `json:"source_type"`
			Status     string                 `json:"status"`
			Config     map[string]interface{} `json:"config"`
		}
		if err := json.Unmarshal([]byte(configValue), &payload); err == nil {
			item.Name = payload.Name
			item.SourceType = payload.SourceType
			if strings.TrimSpace(payload.Status) != "" {
				item.Status = payload.Status
			}
			item.Config = payload.Config
		}
		if item.Name == "" {
			if desc.Valid {
				item.Name = desc.String
			} else {
				item.Name = actualSourceKey
			}
		}
		return item, nil
	}
	return model.DataSource{}, sql.ErrNoRows
}

func buildDataSourceLookupCandidates(sourceKey string) []string {
	trimmed := strings.TrimSpace(sourceKey)
	if trimmed == "" {
		return nil
	}
	items := []string{trimmed}
	lowerKey := strings.ToLower(trimmed)
	upperKey := strings.ToUpper(trimmed)
	if lowerKey != trimmed {
		items = append(items, lowerKey)
	}
	if upperKey != trimmed && upperKey != lowerKey {
		items = append(items, upperKey)
	}
	switch upperKey {
	case "MOCK":
		items = append(items, "mock_stock", "MOCK_STOCK")
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		normalized := strings.TrimSpace(item)
		if normalized == "" {
			continue
		}
		key := strings.ToLower(normalized)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func (r *MySQLGrowthRepo) ensureBuiltinDataSources() error {
	now := time.Now()
	type builtinDataSource struct {
		ID          string
		SourceKey   string
		ConfigValue string
		Description string
	}
	builtins := []builtinDataSource{
		{
			ID:          "cfg_data_source_mock_stock",
			SourceKey:   "mock_stock",
			ConfigValue: `{"name":"Mock Stock Quotes","source_type":"STOCK","status":"ACTIVE","config":{"provider":"MOCK","endpoint":"http://127.0.0.1:18080/healthz","retry_times":0,"fail_threshold":5,"retry_interval_ms":200,"health_timeout_ms":3000,"alert_receiver_id":"admin_001"}}`,
			Description: "内置模拟股票行情数据源",
		},
		{
			ID:          "cfg_data_source_tushare",
			SourceKey:   "tushare",
			ConfigValue: `{"name":"Tushare","source_type":"STOCK","status":"ACTIVE","config":{"provider":"TUSHARE","endpoint":"https://api.tushare.pro","token":"","retry_times":1,"fail_threshold":3,"retry_interval_ms":500,"health_timeout_ms":8000,"alert_receiver_id":"admin_001"}}`,
			Description: "内置Tushare股票行情数据源",
		},
		{
			ID:          "cfg_data_source_akshare",
			SourceKey:   "akshare",
			ConfigValue: `{"name":"AkShare","source_type":"MARKET","status":"ACTIVE","config":{"provider":"AKSHARE","python_bin":"../services/strategy-engine/.venv/bin/python","bridge_script":"../services/strategy-engine/app/tools/market_bridge.py","retry_times":0,"fail_threshold":3,"retry_interval_ms":300,"health_timeout_ms":10000,"alert_receiver_id":"admin_001"}}`,
			Description: "内置 AkShare 多市场数据源",
		},
		{
			ID:          "cfg_data_source_tickermd",
			SourceKey:   "tickermd",
			ConfigValue: `{"name":"TickerMD","source_type":"MARKET","status":"ACTIVE","config":{"provider":"TICKERMD","endpoint":"http://39.107.99.235:1008","quotes_endpoint":"http://39.107.99.235:1008/getQuote.php","kline_endpoint":"http://39.107.99.235:1008/redis.php","ws_endpoint":"ws://39.107.99.235/ws","retry_times":1,"fail_threshold":3,"retry_interval_ms":500,"health_timeout_ms":8000,"alert_receiver_id":"admin_001"}}`,
			Description: "内置 TickerMD 行情备份数据源",
		},
		{
			ID:          "cfg_data_source_myself",
			SourceKey:   "myself",
			ConfigValue: `{"name":"Myself","source_type":"MARKET","status":"ACTIVE","config":{"provider":"MYSELF","endpoint":"https://qt.gtimg.cn/q=s_sh000001","stock_kline_endpoint_tencent":"https://web.ifzq.gtimg.cn/appstock/app/fqkline/get","stock_kline_endpoint_sina":"https://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData","futures_kline_endpoint_sina":"https://stock2.finance.sina.com.cn/futures/api/jsonp.php/var%20_TEST=/InnerFuturesNewService.getDailyKLine","referer":"https://finance.sina.com.cn","retry_times":1,"fail_threshold":3,"retry_interval_ms":500,"health_timeout_ms":8000,"alert_receiver_id":"admin_001"}}`,
			Description: "内置 Myself 新浪/腾讯聚合行情数据源",
		},
	}
	for _, item := range builtins {
		configKey := "data_source." + item.SourceKey
		_, err := r.db.Exec(`
	INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
	SELECT ?, ?, ?, ?, ?, ?
	FROM DUAL
	WHERE NOT EXISTS (
	  SELECT 1 FROM system_configs WHERE LOWER(config_key) = LOWER(?)
	)`,
			item.ID,
			configKey,
			item.ConfigValue,
			item.Description,
			"system",
			now,
			configKey,
		)
		if err != nil {
			return err
		}
	}
	defaultConfigs := []struct {
		Key         string
		Value       string
		Description string
	}{
		{
			Key:         "stock.quotes.default_source_key",
			Value:       "TUSHARE",
			Description: "股票行情默认数据源",
		},
		{
			Key:         "futures.quotes.default_source_key",
			Value:       "TUSHARE",
			Description: "期货行情默认数据源",
		},
		{
			Key:         marketStockPriorityConfigKey,
			Value:       "TUSHARE,AKSHARE,TICKERMD,MOCK",
			Description: "股票日线多源优先级",
		},
		{
			Key:         marketFuturesPriorityConfigKey,
			Value:       "TUSHARE,TICKERMD,AKSHARE,MOCK",
			Description: "期货日线多源优先级",
		},
		{
			Key:         marketNewsPriorityConfigKey,
			Value:       "AKSHARE,TUSHARE",
			Description: "市场资讯多源优先级",
		},
		{
			Key:         "market.news.default_source_key",
			Value:       "AKSHARE",
			Description: "市场资讯默认数据源",
		},
	}
	for _, item := range defaultConfigs {
		_, err := r.db.Exec(`
	INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
	SELECT ?, ?, ?, ?, ?, ?
	FROM DUAL
	WHERE NOT EXISTS (
	  SELECT 1 FROM system_configs WHERE LOWER(config_key) = LOWER(?)
	)`,
			newID("cfg"),
			item.Key,
			item.Value,
			item.Description,
			"system",
			now,
			item.Key,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) persistDataSourceHealthLog(item model.DataSourceHealthCheck, checkedAt time.Time) error {
	reachable := 0
	if item.Reachable {
		reachable = 1
	}
	httpStatus := interface{}(nil)
	if item.HTTPStatus > 0 {
		httpStatus = item.HTTPStatus
	}
	_, err := r.db.Exec(`
INSERT INTO data_source_health_logs (id, source_key, status, reachable, http_status, latency_ms, message, checked_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		newID("dshl"),
		item.SourceKey,
		item.Status,
		reachable,
		httpStatus,
		item.LatencyMS,
		item.Message,
		checkedAt,
	)
	return err
}

func (r *MySQLGrowthRepo) getLatestDataSourceHealthStatus(sourceKey string) (string, error) {
	var status string
	if err := r.db.QueryRow(`
SELECT status
FROM data_source_health_logs
WHERE source_key = ?
ORDER BY checked_at DESC, id DESC
LIMIT 1`, sourceKey).Scan(&status); err != nil {
		return "", err
	}
	return strings.ToUpper(strings.TrimSpace(status)), nil
}

func performDataSourceHealthCheckAttempt(endpoint string, timeoutMS int) model.DataSourceHealthCheck {
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	start := time.Now()
	req, reqErr := http.NewRequest(http.MethodGet, endpoint, nil)
	if reqErr != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNKNOWN",
			Message:         reqErr.Error(),
			FailureCategory: "REQUEST_ERROR",
		}
	}
	resp, callErr := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if callErr != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       false,
			LatencyMS:       latency,
			Message:         callErr.Error(),
			FailureCategory: classifyDataSourceRequestError(callErr),
		}
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return model.DataSourceHealthCheck{
			Status:     "HEALTHY",
			Reachable:  true,
			HTTPStatus: resp.StatusCode,
			LatencyMS:  latency,
			Message:    "ok",
		}
	}
	return model.DataSourceHealthCheck{
		Status:          "UNHEALTHY",
		Reachable:       true,
		HTTPStatus:      resp.StatusCode,
		LatencyMS:       latency,
		Message:         resp.Status,
		FailureCategory: "HTTP_ERROR",
	}
}

func performTushareDataSourceHealthCheckAttempt(endpoint string, token string, timeoutMS int) model.DataSourceHealthCheck {
	if strings.TrimSpace(token) == "" {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Message:         "tushare token not configured",
			FailureCategory: "CONFIG_ERROR",
		}
	}
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	start := time.Now()
	now := time.Now()
	startDate := now.AddDate(0, 0, -3).Format("20060102")
	endDate := now.Format("20060102")
	payload := map[string]interface{}{
		"api_name": "trade_cal",
		"token":    token,
		"params": map[string]string{
			"exchange":   "SSE",
			"start_date": startDate,
			"end_date":   endDate,
		},
		"fields": "exchange,cal_date,is_open",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNKNOWN",
			Message:         err.Error(),
			FailureCategory: "REQUEST_ERROR",
		}
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNKNOWN",
			Message:         err.Error(),
			FailureCategory: "REQUEST_ERROR",
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()
	if err != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       false,
			LatencyMS:       latency,
			Message:         err.Error(),
			FailureCategory: classifyDataSourceRequestError(err),
		}
	}
	defer resp.Body.Close()

	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       true,
			HTTPStatus:      resp.StatusCode,
			LatencyMS:       latency,
			Message:         readErr.Error(),
			FailureCategory: "RESPONSE_ERROR",
		}
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       true,
			HTTPStatus:      resp.StatusCode,
			LatencyMS:       latency,
			Message:         resp.Status,
			FailureCategory: "HTTP_ERROR",
		}
	}

	parsed := struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}{}
	if err := json.Unmarshal(responseBody, &parsed); err != nil {
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       true,
			HTTPStatus:      resp.StatusCode,
			LatencyMS:       latency,
			Message:         "invalid tushare response",
			FailureCategory: "RESPONSE_PARSE_ERROR",
		}
	}
	if parsed.Code != 0 {
		category := "AUTH_ERROR"
		msgLower := strings.ToLower(parsed.Msg)
		if !strings.Contains(msgLower, "token") && !strings.Contains(msgLower, "auth") {
			category = "API_ERROR"
		}
		return model.DataSourceHealthCheck{
			Status:          "UNHEALTHY",
			Reachable:       true,
			HTTPStatus:      resp.StatusCode,
			LatencyMS:       latency,
			Message:         strings.TrimSpace(parsed.Msg),
			FailureCategory: category,
		}
	}
	return model.DataSourceHealthCheck{
		Status:     "HEALTHY",
		Reachable:  true,
		HTTPStatus: resp.StatusCode,
		LatencyMS:  latency,
		Message:    "ok",
	}
}

func classifyDataSourceRequestError(err error) string {
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return "TIMEOUT"
	}
	errorText := strings.ToLower(err.Error())
	switch {
	case strings.Contains(errorText, "connection refused"),
		strings.Contains(errorText, "no such host"),
		strings.Contains(errorText, "dial tcp"),
		strings.Contains(errorText, "i/o timeout"):
		return "NETWORK_ERROR"
	default:
		return "REQUEST_ERROR"
	}
}

func (r *MySQLGrowthRepo) createDataSourceHealthWorkflowMessage(source model.DataSource, check model.DataSourceHealthCheck, event string) {
	receiverID := parseDataSourceAlertReceiverID(source.Config)
	if strings.TrimSpace(receiverID) == "" {
		return
	}
	switch strings.ToUpper(strings.TrimSpace(event)) {
	case "UNHEALTHY":
		threshold := parseDataSourceFailThreshold(source.Config)
		_ = r.AdminCreateWorkflowMessage(
			"",
			source.SourceKey,
			"SYSTEM",
			receiverID,
			"system",
			"DATA_SOURCE_UNHEALTHY",
			"数据源健康告警",
			fmt.Sprintf("数据源 %s 连续失败 %d 次（阈值 %d），最近状态：%s", source.SourceKey, check.ConsecutiveFailures, threshold, check.Message),
		)
	case "RECOVERED":
		_ = r.AdminCreateWorkflowMessage(
			"",
			source.SourceKey,
			"SYSTEM",
			receiverID,
			"system",
			"DATA_SOURCE_RECOVERED",
			"数据源恢复通知",
			fmt.Sprintf("数据源 %s 已恢复，最近耗时 %dms", source.SourceKey, check.LatencyMS),
		)
	}
}

func (r *MySQLGrowthRepo) countConsecutiveDataSourceFailures(sourceKey string, limit int) (int, error) {
	rows, err := r.db.Query(`
SELECT status
FROM data_source_health_logs
WHERE source_key = ?
ORDER BY checked_at DESC, id DESC
LIMIT ?`, sourceKey, limit)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return 0, err
		}
		if strings.ToUpper(strings.TrimSpace(status)) != "UNHEALTHY" {
			break
		}
		count++
	}
	return count, nil
}

func parseDataSourceFailThreshold(config map[string]interface{}) int {
	return parseDataSourceIntConfig(config, "fail_threshold", 3, 1, 20)
}

func parseDataSourceTimeoutMS(config map[string]interface{}) int {
	return parseDataSourceIntConfig(config, "health_timeout_ms", 3000, 500, 15000)
}

func parseDataSourceRetryTimes(config map[string]interface{}) int {
	return parseDataSourceIntConfig(config, "retry_times", 0, 0, 5)
}

func parseDataSourceRetryIntervalMS(config map[string]interface{}) int {
	return parseDataSourceIntConfig(config, "retry_interval_ms", 200, 0, 10000)
}

func parseDataSourceAlertReceiverID(config map[string]interface{}) string {
	if receiverID := parseDataSourceStringConfig(config, "alert_receiver_id", "receiver_id"); receiverID != "" {
		return receiverID
	}
	return "admin_001"
}

func parseDataSourceStringConfig(config map[string]interface{}, keys ...string) string {
	if config == nil {
		return ""
	}
	for _, key := range keys {
		raw, ok := config[key]
		if !ok || raw == nil {
			continue
		}
		switch value := raw.(type) {
		case string:
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		default:
			text := strings.TrimSpace(fmt.Sprintf("%v", value))
			if text != "" && text != "<nil>" {
				return text
			}
		}
	}
	return ""
}

func parseDataSourceIntConfig(config map[string]interface{}, key string, defaultValue int, minValue int, maxValue int) int {
	if config == nil {
		return defaultValue
	}
	raw, ok := config[key]
	if !ok {
		return defaultValue
	}
	value, ok := parseFlexibleInt(raw)
	if !ok {
		return defaultValue
	}
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func parseFlexibleInt(raw interface{}) (int, bool) {
	switch value := raw.(type) {
	case int:
		return value, true
	case int8:
		return int(value), true
	case int16:
		return int(value), true
	case int32:
		return int(value), true
	case int64:
		return int(value), true
	case float32:
		return int(value), true
	case float64:
		return int(value), true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(value))
		if err != nil {
			return 0, false
		}
		return parsed, true
	default:
		return 0, false
	}
}

func (r *MySQLGrowthRepo) AdminListSystemConfigs(keyword string, page int, pageSize int) ([]model.SystemConfig, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if keyword != "" {
		like := "%" + keyword + "%"
		filter += " AND (config_key LIKE ? OR description LIKE ?)"
		args = append(args, like, like)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM system_configs"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, config_key, config_value, description, updated_by, updated_at
FROM system_configs` + filter + `
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.SystemConfig, 0)
	for rows.Next() {
		var item model.SystemConfig
		var desc sql.NullString
		var updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.ConfigKey, &item.ConfigValue, &desc, &item.UpdatedBy, &updatedAt); err != nil {
			return nil, 0, err
		}
		if desc.Valid {
			item.Description = desc.String
		}
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminUpsertSystemConfig(configKey string, configValue string, description string, operator string) error {
	if operator == "" {
		operator = "admin_unknown"
	}
	_, err := r.db.Exec(`
INSERT INTO system_configs (id, config_key, config_value, description, updated_by, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	config_value = VALUES(config_value),
	description = VALUES(description),
	updated_by = VALUES(updated_by),
	updated_at = VALUES(updated_at)`,
		newID("cfg"), configKey, configValue, description, operator, time.Now(),
	)
	return err
}

func (r *MySQLGrowthRepo) AdminListReviewTasks(module string, status string, submitterID string, reviewerID string, page int, pageSize int) ([]model.ReviewTask, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if status != "" {
		filter += " AND status = ?"
		args = append(args, strings.ToUpper(status))
	}
	if submitterID != "" {
		filter += " AND submitter_id = ?"
		args = append(args, submitterID)
	}
	if reviewerID != "" {
		filter += " AND reviewer_id = ?"
		args = append(args, reviewerID)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM review_tasks"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, module, target_id, submitter_id, reviewer_id, status, submit_note, review_note, submitted_at, reviewed_at
FROM review_tasks` + filter + `
ORDER BY submitted_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.ReviewTask, 0)
	for rows.Next() {
		var item model.ReviewTask
		var reviewerID, submitNote, reviewNote sql.NullString
		var submittedAt time.Time
		var reviewedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.Module, &item.TargetID, &item.SubmitterID, &reviewerID, &item.Status, &submitNote, &reviewNote, &submittedAt, &reviewedAt); err != nil {
			return nil, 0, err
		}
		if reviewerID.Valid {
			item.ReviewerID = reviewerID.String
		}
		if submitNote.Valid {
			item.SubmitNote = submitNote.String
		}
		if reviewNote.Valid {
			item.ReviewNote = reviewNote.String
		}
		item.SubmittedAt = submittedAt.Format(time.RFC3339)
		if reviewedAt.Valid {
			item.ReviewedAt = reviewedAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminSubmitReviewTask(module string, targetID string, submitterID string, reviewerID string, submitNote string) (string, error) {
	module = strings.ToUpper(strings.TrimSpace(module))
	if module == "" {
		return "", errors.New("module is required")
	}
	var pendingCount int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM review_tasks WHERE module = ? AND target_id = ? AND status = 'PENDING'", module, targetID).Scan(&pendingCount); err != nil {
		return "", err
	}
	if pendingCount > 0 {
		return "", errors.New("pending review task already exists")
	}

	// Move target into REVIEWING state when submitting for review.
	if err := r.applyModuleTargetStatus(r.db, module, targetID, "REVIEWING"); err != nil {
		return "", err
	}

	id := newID("rt")
	now := time.Now()
	_, err := r.db.Exec(`
INSERT INTO review_tasks (id, module, target_id, submitter_id, reviewer_id, status, submit_note, submitted_at, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 'PENDING', ?, ?, ?, ?)`,
		id, module, targetID, submitterID, reviewerID, submitNote, now, now, now,
	)
	if err != nil {
		return "", err
	}
	_ = r.AdminCreateWorkflowMessage(
		id,
		targetID,
		module,
		reviewerID,
		submitterID,
		"REVIEW_SUBMITTED",
		"内容已提交审核",
		"模块 "+module+" 的目标 "+targetID+" 已提交审核",
	)
	return id, nil
}

func (r *MySQLGrowthRepo) AdminAssignReviewTask(reviewID string, reviewerID string) error {
	_, err := r.db.Exec(
		"UPDATE review_tasks SET reviewer_id = ?, updated_at = ? WHERE id = ? AND status = 'PENDING'",
		reviewerID, time.Now(), reviewID,
	)
	if err != nil {
		return err
	}
	_ = r.AdminCreateWorkflowMessage(
		reviewID,
		reviewID,
		"WORKFLOW",
		reviewerID,
		"",
		"REVIEW_ASSIGNED",
		"审核任务已分配",
		"你有新的审核任务 "+reviewID,
	)
	return nil
}

func (r *MySQLGrowthRepo) AdminReviewTaskDecision(reviewID string, status string, reviewerID string, reviewNote string) error {
	status = strings.ToUpper(strings.TrimSpace(status))
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var module, targetID, currentStatus, submitterID string
	var assignedReviewer sql.NullString
	err = tx.QueryRow("SELECT module, target_id, submitter_id, reviewer_id, status FROM review_tasks WHERE id = ? FOR UPDATE", reviewID).
		Scan(&module, &targetID, &submitterID, &assignedReviewer, &currentStatus)
	if err != nil {
		return err
	}
	if strings.ToUpper(currentStatus) != "PENDING" {
		return errors.New("review task is not pending")
	}
	if assignedReviewer.Valid && strings.TrimSpace(assignedReviewer.String) != "" && assignedReviewer.String != reviewerID {
		return errors.New("reviewer mismatch")
	}

	targetStatus := "DRAFT"
	if status == "APPROVED" {
		targetStatus = "PUBLISHED"
	}
	if err = r.applyModuleTargetStatus(tx, module, targetID, targetStatus); err != nil {
		return err
	}

	_, err = tx.Exec(`
UPDATE review_tasks
SET status = ?, reviewer_id = ?, review_note = ?, reviewed_at = ?, updated_at = ?
WHERE id = ?`,
		status, reviewerID, reviewNote, time.Now(), time.Now(), reviewID,
	)
	if err != nil {
		return err
	}
	eventType := "REVIEW_REJECTED"
	if status == "APPROVED" {
		eventType = "REVIEW_APPROVED"
	}
	_ = r.AdminCreateWorkflowMessage(
		reviewID,
		targetID,
		module,
		submitterID,
		reviewerID,
		eventType,
		"审核结果通知",
		"审核结果为 "+status+"，目标 "+targetID,
	)
	return tx.Commit()
}

func (r *MySQLGrowthRepo) GetSchedulerJobNameByRunID(runID string) (string, error) {
	var jobName string
	if err := r.db.QueryRow("SELECT job_name FROM scheduler_job_runs WHERE id = ?", runID).Scan(&jobName); err != nil {
		return "", err
	}
	return jobName, nil
}

func (r *MySQLGrowthRepo) AdminListSchedulerJobRuns(jobName string, status string, page int, pageSize int) ([]model.SchedulerJobRun, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if jobName != "" {
		filter += " AND job_name = ?"
		args = append(args, jobName)
	}
	if status != "" {
		filter += " AND status = ?"
		args = append(args, strings.ToUpper(status))
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, parent_run_id, retry_count, job_name, trigger_source, status, started_at, finished_at, result_summary, error_message, operator_id
FROM scheduler_job_runs` + filter + `
ORDER BY started_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.SchedulerJobRun, 0)
	for rows.Next() {
		var item model.SchedulerJobRun
		var finishedAt sql.NullTime
		var resultSummary, errorMsg, operatorID, parentRunID sql.NullString
		var startedAt time.Time
		if err := rows.Scan(&item.ID, &parentRunID, &item.RetryCount, &item.JobName, &item.TriggerSource, &item.Status, &startedAt, &finishedAt, &resultSummary, &errorMsg, &operatorID); err != nil {
			return nil, 0, err
		}
		item.StartedAt = startedAt.Format(time.RFC3339)
		if parentRunID.Valid {
			item.ParentRunID = parentRunID.String
		}
		if finishedAt.Valid {
			item.FinishedAt = finishedAt.Time.Format(time.RFC3339)
		}
		if resultSummary.Valid {
			item.ResultSummary = resultSummary.String
		}
		if errorMsg.Valid {
			item.ErrorMessage = errorMsg.String
		}
		if operatorID.Valid {
			item.OperatorID = operatorID.String
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminListNewsSyncRunDetails(runID string, syncType string, source string, symbol string, status string, page int, pageSize int) ([]model.NewsSyncRunDetail, int, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return []model.NewsSyncRunDetail{}, 0, nil
	}
	offset := (page - 1) * pageSize
	args := []interface{}{runID}
	filter := " WHERE run_id = ?"
	if syncType = strings.ToUpper(strings.TrimSpace(syncType)); syncType != "" {
		filter += " AND sync_type = ?"
		args = append(args, syncType)
	}
	if source = strings.TrimSpace(source); source != "" {
		filter += " AND source LIKE ?"
		args = append(args, "%"+source+"%")
	}
	if symbol = strings.ToUpper(strings.TrimSpace(symbol)); symbol != "" {
		filter += " AND symbol = ?"
		args = append(args, symbol)
	}
	if status = strings.ToUpper(strings.TrimSpace(status)); status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM news_sync_run_details"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT
	id, run_id, job_name, sync_type, source, symbol, status,
	fetched_count, upserted_count, failed_count,
	warning_text, error_text, started_at, finished_at, created_at, updated_at
FROM news_sync_run_details` + filter + `
ORDER BY created_at DESC, sync_type ASC, source ASC, symbol ASC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.NewsSyncRunDetail, 0, pageSize)
	for rows.Next() {
		var item model.NewsSyncRunDetail
		var sourceVal, symbolVal, warningVal, errorVal sql.NullString
		var finishedAt sql.NullTime
		var startedAt, createdAt, updatedAt time.Time
		if err := rows.Scan(
			&item.ID,
			&item.RunID,
			&item.JobName,
			&item.SyncType,
			&sourceVal,
			&symbolVal,
			&item.Status,
			&item.FetchedCount,
			&item.UpsertedCount,
			&item.FailedCount,
			&warningVal,
			&errorVal,
			&startedAt,
			&finishedAt,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, err
		}
		if sourceVal.Valid {
			item.Source = sourceVal.String
		}
		if symbolVal.Valid {
			item.Symbol = symbolVal.String
		}
		if warningVal.Valid {
			item.WarningText = warningVal.String
		}
		if errorVal.Valid {
			item.ErrorText = errorVal.String
		}
		item.StartedAt = startedAt.Format(time.RFC3339)
		if finishedAt.Valid {
			item.FinishedAt = finishedAt.Time.Format(time.RFC3339)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminCreateNewsSyncRunDetails(runID string, details []model.NewsSyncRunDetail) error {
	runID = strings.TrimSpace(runID)
	if runID == "" || len(details) == 0 {
		return nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
INSERT INTO news_sync_run_details
	(id, run_id, job_name, sync_type, source, symbol, status, fetched_count, upserted_count, failed_count, warning_text, error_text, started_at, finished_at, created_at, updated_at)
VALUES
	(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
	job_name = VALUES(job_name),
	sync_type = VALUES(sync_type),
	source = VALUES(source),
	symbol = VALUES(symbol),
	status = VALUES(status),
	fetched_count = VALUES(fetched_count),
	upserted_count = VALUES(upserted_count),
	failed_count = VALUES(failed_count),
	warning_text = VALUES(warning_text),
	error_text = VALUES(error_text),
	started_at = VALUES(started_at),
	finished_at = VALUES(finished_at),
	updated_at = VALUES(updated_at)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, detail := range details {
		id := strings.TrimSpace(detail.ID)
		if id == "" {
			id = newID("nsd")
		}
		jobName := strings.TrimSpace(detail.JobName)
		if jobName == "" {
			jobName = tushareNewsIncrementalJobName
		}
		syncType := normalizeTushareSyncType(detail.SyncType)
		if syncType == "" {
			syncType = strings.ToUpper(strings.TrimSpace(detail.SyncType))
		}
		if syncType == "" {
			syncType = tushareSyncTypeNewsBrief
		}
		status := strings.ToUpper(strings.TrimSpace(detail.Status))
		if status == "" {
			if detail.FailedCount > 0 || strings.TrimSpace(detail.ErrorText) != "" {
				status = "FAILED"
			} else {
				status = "SUCCESS"
			}
		}
		startedAt := parseDateTimeOrDefault(detail.StartedAt, time.Now())
		finishedAt := parseDateTimeOrDefault(detail.FinishedAt, startedAt)
		if strings.TrimSpace(detail.FinishedAt) == "" {
			finishedAt = startedAt
		}
		warningText := truncateByRunes(normalizeUTF8Text(detail.WarningText), 512)
		errorText := truncateByRunes(normalizeUTF8Text(detail.ErrorText), 512)
		source := truncateByRunes(normalizeUTF8Text(detail.Source), 128)
		symbol := truncateByRunes(normalizeTushareSymbol(detail.Symbol), 32)
		now := time.Now()
		_, err := stmt.Exec(
			id,
			runID,
			jobName,
			syncType,
			nullableString(source),
			nullableString(symbol),
			status,
			maxInt(0, detail.FetchedCount),
			maxInt(0, detail.UpsertedCount),
			maxInt(0, detail.FailedCount),
			nullableString(warningText),
			nullableString(errorText),
			startedAt,
			finishedAt,
			now,
			now,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MySQLGrowthRepo) AdminCreateSchedulerJobRun(jobName string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	id := newID("jr")
	now := time.Now()
	jobName = truncateByRunes(normalizeUTF8Text(jobName), 64)
	operatorID = truncateByRunes(normalizeUTF8Text(operatorID), 32)
	safeResultSummary := truncateByRunes(normalizeUTF8Text(resultSummary), 512)
	safeErrorMessage := truncateByRunes(normalizeUTF8Text(errorMessage), 512)
	var finishedAt interface{} = nil
	if strings.ToUpper(status) != "RUNNING" {
		finishedAt = now
	}
	_, err := r.db.Exec(`
INSERT INTO scheduler_job_runs (id, parent_run_id, retry_count, job_name, trigger_source, status, started_at, finished_at, result_summary, error_message, operator_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, nil, 0, jobName, strings.ToUpper(triggerSource), strings.ToUpper(status), now, finishedAt, safeResultSummary, safeErrorMessage, operatorID, now,
	)
	if err != nil {
		return "", err
	}
	_, _ = r.db.Exec(`
UPDATE scheduler_job_definitions
SET last_run_at = ?, updated_at = ?
WHERE job_name = ?`, now, now, strings.TrimSpace(jobName))
	_ = r.AdminCreateWorkflowMessage(
		"",
		id,
		"SYSTEM",
		operatorID,
		operatorID,
		"JOB_TRIGGERED",
		"任务已触发",
		"任务 "+jobName+" 已触发，状态 "+strings.ToUpper(status),
	)
	if strings.ToUpper(status) == "FAILED" {
		_ = r.AdminCreateWorkflowMessage(
			"",
			id,
			"SYSTEM",
			operatorID,
			operatorID,
			"JOB_FAILED",
			"任务执行失败",
			"任务 "+jobName+" 执行失败："+safeErrorMessage,
		)
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminRetrySchedulerJobRun(runID string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	var jobName string
	var prevRetry int
	err := r.db.QueryRow("SELECT job_name, retry_count FROM scheduler_job_runs WHERE id = ?", runID).Scan(&jobName, &prevRetry)
	if err != nil {
		return "", err
	}
	id := newID("jr")
	now := time.Now()
	operatorID = truncateByRunes(normalizeUTF8Text(operatorID), 32)
	safeResultSummary := truncateByRunes(normalizeUTF8Text(resultSummary), 512)
	safeErrorMessage := truncateByRunes(normalizeUTF8Text(errorMessage), 512)
	upperStatus := strings.ToUpper(status)
	var finishedAt interface{} = nil
	if upperStatus != "RUNNING" {
		finishedAt = now
	}
	upperTriggerSource := strings.ToUpper(strings.TrimSpace(triggerSource))
	if upperTriggerSource == "" {
		upperTriggerSource = "MANUAL"
	}
	_, err = r.db.Exec(`
INSERT INTO scheduler_job_runs (id, parent_run_id, retry_count, job_name, trigger_source, status, started_at, finished_at, result_summary, error_message, operator_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, runID, prevRetry+1, jobName, upperTriggerSource, upperStatus, now, finishedAt, safeResultSummary, safeErrorMessage, operatorID, now,
	)
	if err != nil {
		return "", err
	}
	_, _ = r.db.Exec(`
UPDATE scheduler_job_definitions
SET last_run_at = ?, updated_at = ?
WHERE job_name = ?`, now, now, strings.TrimSpace(jobName))
	_ = r.AdminCreateWorkflowMessage(
		"",
		id,
		"SYSTEM",
		operatorID,
		operatorID,
		"JOB_RETRIED",
		"任务重跑已触发",
		"任务 "+jobName+" 基于运行 "+runID+" 重跑，状态 "+upperStatus,
	)
	if upperStatus == "FAILED" {
		_ = r.AdminCreateWorkflowMessage(
			"",
			id,
			"SYSTEM",
			operatorID,
			operatorID,
			"JOB_FAILED",
			"任务重跑失败",
			"任务 "+jobName+" 重跑失败："+safeErrorMessage,
		)
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminListSchedulerJobDefinitions(status string, module string, page int, pageSize int) ([]model.SchedulerJobDefinition, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND status = ?"
		args = append(args, strings.ToUpper(status))
	}
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_definitions"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, job_name, display_name, module, cron_expr, status, last_run_at, updated_by, created_at, updated_at
FROM scheduler_job_definitions` + filter + `
ORDER BY updated_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.SchedulerJobDefinition, 0)
	for rows.Next() {
		var item model.SchedulerJobDefinition
		var lastRunAt sql.NullTime
		var createdAt, updatedAt time.Time
		if err := rows.Scan(&item.ID, &item.JobName, &item.DisplayName, &item.Module, &item.CronExpr, &item.Status, &lastRunAt, &item.UpdatedBy, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		if lastRunAt.Valid {
			item.LastRunAt = lastRunAt.Time.Format(time.RFC3339)
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		item.UpdatedAt = updatedAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateSchedulerJobDefinition(item model.SchedulerJobDefinition, operatorID string) (string, error) {
	id := newID("jobdef")
	now := time.Now()
	_, err := r.db.Exec(`
INSERT INTO scheduler_job_definitions (id, job_name, display_name, module, cron_expr, status, updated_by, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, item.JobName, item.DisplayName, strings.ToUpper(item.Module), item.CronExpr, strings.ToUpper(item.Status), operatorID, now, now,
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateSchedulerJobDefinition(id string, item model.SchedulerJobDefinition, operatorID string) error {
	_, err := r.db.Exec(`
UPDATE scheduler_job_definitions
SET display_name = ?, module = ?, cron_expr = ?, status = ?, updated_by = ?, updated_at = ?
WHERE id = ?`,
		item.DisplayName, strings.ToUpper(item.Module), item.CronExpr, strings.ToUpper(item.Status), operatorID, time.Now(), id,
	)
	return err
}

func (r *MySQLGrowthRepo) AdminUpdateSchedulerJobDefinitionStatus(id string, status string, operatorID string) error {
	_, err := r.db.Exec(
		"UPDATE scheduler_job_definitions SET status = ?, updated_by = ?, updated_at = ? WHERE id = ?",
		strings.ToUpper(status), operatorID, time.Now(), id,
	)
	return err
}

func (r *MySQLGrowthRepo) AdminDeleteSchedulerJobDefinition(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return sql.ErrNoRows
	}
	result, err := r.db.Exec("DELETE FROM scheduler_job_definitions WHERE id = ?", id)
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

func (r *MySQLGrowthRepo) AdminListWorkflowMessages(module string, eventType string, isRead string, receiverID string, page int, pageSize int) ([]model.WorkflowMessage, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if eventType != "" {
		filter += " AND event_type = ?"
		args = append(args, strings.ToUpper(eventType))
	}
	if receiverID != "" {
		filter += " AND receiver_id = ?"
		args = append(args, receiverID)
	}
	if isRead != "" {
		switch strings.ToLower(isRead) {
		case "true", "1":
			filter += " AND is_read = 1"
		case "false", "0":
			filter += " AND is_read = 0"
		}
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM workflow_messages"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, review_id, target_id, module, receiver_id, sender_id, event_type, title, content, is_read, created_at, read_at
FROM workflow_messages` + filter + `
ORDER BY created_at DESC
LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := make([]model.WorkflowMessage, 0)
	for rows.Next() {
		var item model.WorkflowMessage
		var reviewID, receiverID, senderID sql.NullString
		var readAt sql.NullTime
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &reviewID, &item.TargetID, &item.Module, &receiverID, &senderID, &item.EventType, &item.Title, &item.Content, &item.IsRead, &createdAt, &readAt); err != nil {
			return nil, 0, err
		}
		if reviewID.Valid {
			item.ReviewID = reviewID.String
		}
		if receiverID.Valid {
			item.ReceiverID = receiverID.String
		}
		if senderID.Valid {
			item.SenderID = senderID.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		if readAt.Valid {
			item.ReadAt = readAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCountUnreadWorkflowMessages(module string, eventType string, receiverID string) (int, error) {
	args := []interface{}{}
	filter := " WHERE is_read = 0"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if eventType != "" {
		filter += " AND event_type = ?"
		args = append(args, strings.ToUpper(eventType))
	}
	if receiverID != "" {
		filter += " AND receiver_id = ?"
		args = append(args, receiverID)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM workflow_messages"+filter, args...).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *MySQLGrowthRepo) AdminUpdateWorkflowMessageRead(id string, isRead bool) error {
	readAt := interface{}(nil)
	if isRead {
		readAt = time.Now()
	}
	_, err := r.db.Exec("UPDATE workflow_messages SET is_read = ?, read_at = ? WHERE id = ?", isRead, readAt, id)
	return err
}

func (r *MySQLGrowthRepo) AdminBulkReadWorkflowMessages(module string, eventType string, receiverID string) (int64, error) {
	args := []interface{}{}
	filter := " WHERE is_read = 0"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if eventType != "" {
		filter += " AND event_type = ?"
		args = append(args, strings.ToUpper(eventType))
	}
	if receiverID != "" {
		filter += " AND receiver_id = ?"
		args = append(args, receiverID)
	}
	query := "UPDATE workflow_messages SET is_read = 1, read_at = ?" + filter
	args = append([]interface{}{time.Now()}, args...)
	res, err := r.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) AdminCreateWorkflowMessage(reviewID string, targetID string, module string, receiverID string, senderID string, eventType string, title string, content string) error {
	reviewID = truncateByRunes(normalizeUTF8Text(reviewID), 64)
	targetID = truncateByRunes(normalizeUTF8Text(targetID), 64)
	module = truncateByRunes(strings.ToUpper(normalizeUTF8Text(module)), 32)
	receiverID = truncateByRunes(normalizeUTF8Text(receiverID), 32)
	senderID = truncateByRunes(normalizeUTF8Text(senderID), 32)
	eventType = truncateByRunes(strings.ToUpper(normalizeUTF8Text(eventType)), 32)
	title = truncateByRunes(normalizeUTF8Text(title), 128)
	content = truncateByRunes(normalizeUTF8Text(content), 512)

	var reviewVal interface{} = nil
	if reviewID != "" {
		reviewVal = reviewID
	}
	var receiverVal interface{} = nil
	if receiverID != "" {
		receiverVal = receiverID
	}
	var senderVal interface{} = nil
	if senderID != "" {
		senderVal = senderID
	}
	_, err := r.db.Exec(`
INSERT INTO workflow_messages (id, review_id, target_id, module, receiver_id, sender_id, event_type, title, content, is_read, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?)`,
		newID("wm"), reviewVal, targetID, module, receiverVal, senderVal, eventType, title, content, time.Now(),
	)
	return err
}

func (r *MySQLGrowthRepo) AdminGetWorkflowMetrics(module string, receiverID string) (model.WorkflowMetrics, error) {
	result := model.WorkflowMetrics{}
	args := []interface{}{}
	filter := " WHERE 1=1"
	if module != "" {
		filter += " AND module = ?"
		args = append(args, strings.ToUpper(module))
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM review_tasks"+filter+" AND status = 'PENDING'", args...).Scan(&result.PendingReviews); err != nil {
		return result, err
	}

	todayStart := time.Now().Truncate(24 * time.Hour)
	tomorrowStart := todayStart.Add(24 * time.Hour)
	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM review_tasks"+filter+" AND status = 'APPROVED' AND reviewed_at >= ? AND reviewed_at < ?",
		append(args, todayStart, tomorrowStart)...,
	).Scan(&result.ApprovedToday); err != nil {
		return result, err
	}
	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM review_tasks"+filter+" AND status = 'REJECTED' AND reviewed_at >= ? AND reviewed_at < ?",
		append(args, todayStart, tomorrowStart)...,
	).Scan(&result.RejectedToday); err != nil {
		return result, err
	}

	msgArgs := []interface{}{}
	msgFilter := " WHERE 1=1"
	if module != "" {
		msgFilter += " AND module = ?"
		msgArgs = append(msgArgs, strings.ToUpper(module))
	}
	if receiverID != "" {
		msgFilter += " AND receiver_id = ?"
		msgArgs = append(msgArgs, receiverID)
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM workflow_messages"+msgFilter, msgArgs...).Scan(&result.TotalMessages); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM workflow_messages"+msgFilter+" AND is_read = 0", msgArgs...).Scan(&result.UnreadMessages); err != nil {
		return result, err
	}
	return result, nil
}

func (r *MySQLGrowthRepo) AdminGetSchedulerJobMetrics(jobName string) (model.SchedulerJobMetrics, error) {
	result := model.SchedulerJobMetrics{}
	normalizedJobName := strings.TrimSpace(jobName)
	args := []interface{}{}
	filter := " WHERE started_at >= ? AND started_at < ?"
	todayStart := time.Now().Truncate(24 * time.Hour)
	tomorrowStart := todayStart.Add(24 * time.Hour)
	args = append(args, todayStart, tomorrowStart)
	if normalizedJobName != "" {
		filter += " AND job_name = ?"
		args = append(args, normalizedJobName)
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter, args...).Scan(&result.TodayTotal); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND status = 'SUCCESS'", args...).Scan(&result.TodaySuccess); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND status = 'FAILED'", args...).Scan(&result.TodayFailed); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND status = 'RUNNING'", args...).Scan(&result.TodayRunning); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND retry_count > 0", args...).Scan(&result.RetryTotal); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND retry_count > 0 AND status = 'SUCCESS'", args...).Scan(&result.RetrySuccess); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM scheduler_job_runs"+filter+" AND retry_count > 0 AND status = 'FAILED'", args...).Scan(&result.RetryFailed); err != nil {
		return result, err
	}
	if result.RetryTotal > 0 {
		result.RetryHitRate = roundTo(float64(result.RetrySuccess)/float64(result.RetryTotal), 4)
	} else {
		result.RetryHitRate = 0
	}
	var avgRetry sql.NullFloat64
	if err := r.db.QueryRow("SELECT AVG(retry_count) FROM scheduler_job_runs"+filter+" AND retry_count > 0", args...).Scan(&avgRetry); err != nil {
		return result, err
	}
	if avgRetry.Valid {
		result.AvgRetryCount = roundTo(avgRetry.Float64, 2)
	}

	jobStatsRows, err := r.db.Query(`
SELECT
	job_name,
	COUNT(*) AS today_total,
	SUM(CASE WHEN status = 'SUCCESS' THEN 1 ELSE 0 END) AS today_success,
	SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END) AS today_failed,
	SUM(CASE WHEN status = 'RUNNING' THEN 1 ELSE 0 END) AS today_running,
	SUM(CASE WHEN retry_count > 0 THEN 1 ELSE 0 END) AS retry_total,
	SUM(CASE WHEN retry_count > 0 AND status = 'SUCCESS' THEN 1 ELSE 0 END) AS retry_success,
	SUM(CASE WHEN retry_count > 0 AND status = 'FAILED' THEN 1 ELSE 0 END) AS retry_failed,
	SUM(CASE WHEN retry_count > 0 AND trigger_source = 'SYSTEM' THEN 1 ELSE 0 END) AS auto_retry_total,
	AVG(CASE WHEN retry_count > 0 THEN retry_count ELSE NULL END) AS avg_retry_count
FROM scheduler_job_runs`+filter+`
GROUP BY job_name
ORDER BY retry_total DESC, job_name ASC`, args...)
	if err != nil {
		return result, err
	}
	defer jobStatsRows.Close()
	jobStatsMap := make(map[string]*model.SchedulerJobRetryStat)
	for jobStatsRows.Next() {
		var (
			item                model.SchedulerJobRetryStat
			todayTotal          int64
			todaySuccess        int64
			todayFailed         int64
			todayRunning        int64
			retryTotal          int64
			retrySuccess        int64
			retryFailed         int64
			autoRetryTotal      int64
			avgRetryCountPerJob sql.NullFloat64
		)
		if err := jobStatsRows.Scan(
			&item.JobName,
			&todayTotal,
			&todaySuccess,
			&todayFailed,
			&todayRunning,
			&retryTotal,
			&retrySuccess,
			&retryFailed,
			&autoRetryTotal,
			&avgRetryCountPerJob,
		); err != nil {
			return result, err
		}
		item.TodayTotal = int(todayTotal)
		item.TodaySuccess = int(todaySuccess)
		item.TodayFailed = int(todayFailed)
		item.TodayRunning = int(todayRunning)
		item.RetryTotal = int(retryTotal)
		item.RetrySuccess = int(retrySuccess)
		item.RetryFailed = int(retryFailed)
		item.AutoRetryTotal = int(autoRetryTotal)
		if item.RetryTotal > 0 {
			item.RetryHitRate = roundTo(float64(item.RetrySuccess)/float64(item.RetryTotal), 4)
		}
		if avgRetryCountPerJob.Valid {
			item.AvgRetryCount = roundTo(avgRetryCountPerJob.Float64, 2)
		}
		jobStatsMap[item.JobName] = &item
	}
	if err := jobStatsRows.Err(); err != nil {
		return result, err
	}

	if err := r.fillSchedulerRecoveryStats(jobStatsMap, todayStart, tomorrowStart, normalizedJobName); err != nil {
		return result, err
	}

	jobStats := make([]model.SchedulerJobRetryStat, 0, len(jobStatsMap))
	for _, item := range jobStatsMap {
		jobStats = append(jobStats, *item)
	}
	sort.Slice(jobStats, func(i, j int) bool {
		if jobStats[i].RetryTotal == jobStats[j].RetryTotal {
			if jobStats[i].RecoveryTotal == jobStats[j].RecoveryTotal {
				return jobStats[i].JobName < jobStats[j].JobName
			}
			return jobStats[i].RecoveryTotal > jobStats[j].RecoveryTotal
		}
		return jobStats[i].RetryTotal > jobStats[j].RetryTotal
	})
	result.JobRetryStats = jobStats
	for _, item := range jobStats {
		result.AutoRetryTotal += item.AutoRetryTotal
		result.RecoveryTotal += item.RecoveryTotal
		result.RecoverySuccess += item.RecoverySuccess
	}
	if result.RecoveryTotal > 0 {
		result.RecoveryHitRate = roundTo(float64(result.RecoverySuccess)/float64(result.RecoveryTotal), 4)
	}

	failScopeDays := 7
	failSince := time.Now().AddDate(0, 0, -failScopeDays)
	failArgs := []interface{}{failSince}
	failFilter := " WHERE status = 'FAILED' AND started_at >= ?"
	if normalizedJobName != "" {
		failFilter += " AND job_name = ?"
		failArgs = append(failArgs, normalizedJobName)
	}
	failRows, err := r.db.Query(`
SELECT job_name, error_message, started_at
FROM scheduler_job_runs`+failFilter+`
ORDER BY started_at DESC
LIMIT 800`, failArgs...)
	if err != nil {
		return result, err
	}
	defer failRows.Close()
	reasonAggMap := make(map[string]*schedulerFailureReasonAggregate)
	jobReasonAggMap := make(map[string]*schedulerFailureReasonAggregate)
	for failRows.Next() {
		var (
			jobNameValue string
			errorMessage sql.NullString
			startedAt    time.Time
		)
		if err := failRows.Scan(&jobNameValue, &errorMessage, &startedAt); err != nil {
			return result, err
		}
		reason := normalizeSchedulerFailureReason(errorMessage.String)
		agg := reasonAggMap[reason]
		if agg == nil {
			agg = &schedulerFailureReasonAggregate{Reason: reason}
			reasonAggMap[reason] = agg
		}
		agg.Count++
		if agg.LastOccurredAt.IsZero() || startedAt.After(agg.LastOccurredAt) {
			agg.LastOccurredAt = startedAt
		}
		jobReasonKey := strings.TrimSpace(jobNameValue) + "|" + reason
		jobAgg := jobReasonAggMap[jobReasonKey]
		if jobAgg == nil {
			jobAgg = &schedulerFailureReasonAggregate{
				JobName: strings.TrimSpace(jobNameValue),
				Reason:  reason,
			}
			jobReasonAggMap[jobReasonKey] = jobAgg
		}
		jobAgg.Count++
		if jobAgg.LastOccurredAt.IsZero() || startedAt.After(jobAgg.LastOccurredAt) {
			jobAgg.LastOccurredAt = startedAt
		}
	}
	if err := failRows.Err(); err != nil {
		return result, err
	}
	reasons := make([]model.SchedulerJobFailureReason, 0, len(reasonAggMap))
	for _, agg := range reasonAggMap {
		reasons = append(reasons, model.SchedulerJobFailureReason{
			Reason:         agg.Reason,
			Count:          agg.Count,
			LastOccurredAt: agg.LastOccurredAt.Format(time.RFC3339),
		})
	}
	sort.Slice(reasons, func(i, j int) bool {
		if reasons[i].Count == reasons[j].Count {
			return reasons[i].Reason < reasons[j].Reason
		}
		return reasons[i].Count > reasons[j].Count
	})
	if len(reasons) > 8 {
		reasons = reasons[:8]
	}
	jobReasons := make([]model.SchedulerJobFailureByJob, 0, len(jobReasonAggMap))
	for _, agg := range jobReasonAggMap {
		jobReasons = append(jobReasons, model.SchedulerJobFailureByJob{
			JobName:        agg.JobName,
			Reason:         agg.Reason,
			Count:          agg.Count,
			LastOccurredAt: agg.LastOccurredAt.Format(time.RFC3339),
		})
	}
	sort.Slice(jobReasons, func(i, j int) bool {
		if jobReasons[i].Count == jobReasons[j].Count {
			if jobReasons[i].JobName == jobReasons[j].JobName {
				return jobReasons[i].Reason < jobReasons[j].Reason
			}
			return jobReasons[i].JobName < jobReasons[j].JobName
		}
		return jobReasons[i].Count > jobReasons[j].Count
	})
	if len(jobReasons) > 36 {
		jobReasons = jobReasons[:36]
	}
	result.FailureReasons = reasons
	result.JobFailureReasons = jobReasons
	result.FailureReasonScope = fmt.Sprintf("LAST_%d_DAYS", failScopeDays)
	return result, nil
}

type schedulerJobRecoveryState struct {
	JobName       string
	HasRetry      bool
	MaxRetryCount int
	FinalStatus   string
}

type schedulerJobRetryChild struct {
	ID          string
	ParentRunID string
	JobName     string
	Status      string
	RetryCount  int
}

func (r *MySQLGrowthRepo) fillSchedulerRecoveryStats(
	jobStatsMap map[string]*model.SchedulerJobRetryStat,
	startedAt time.Time,
	endedAt time.Time,
	jobName string,
) error {
	args := []interface{}{startedAt, endedAt}
	filter := ""
	if jobName != "" {
		filter += " AND job_name = ?"
		args = append(args, jobName)
	}
	rootRows, err := r.db.Query(`
SELECT id, job_name
FROM scheduler_job_runs
WHERE retry_count = 0
	AND status = 'FAILED'
	AND started_at >= ?
	AND started_at < ?`+filter, args...)
	if err != nil {
		return err
	}
	defer rootRows.Close()

	recoveryStates := make(map[string]*schedulerJobRecoveryState)
	runToRoot := make(map[string]string)
	frontier := make([]string, 0)
	visited := make(map[string]struct{})
	for rootRows.Next() {
		var (
			rootID      string
			rootJobName string
		)
		if err := rootRows.Scan(&rootID, &rootJobName); err != nil {
			return err
		}
		rootID = strings.TrimSpace(rootID)
		if rootID == "" {
			continue
		}
		rootJobName = strings.TrimSpace(rootJobName)
		recoveryStates[rootID] = &schedulerJobRecoveryState{
			JobName:       rootJobName,
			HasRetry:      false,
			MaxRetryCount: 0,
			FinalStatus:   "FAILED",
		}
		runToRoot[rootID] = rootID
		frontier = append(frontier, rootID)
		visited[rootID] = struct{}{}
	}
	if err := rootRows.Err(); err != nil {
		return err
	}
	if len(frontier) == 0 {
		return nil
	}

	const batchSize = 200
	for len(frontier) > 0 {
		nextFrontier := make([]string, 0)
		for i := 0; i < len(frontier); i += batchSize {
			end := i + batchSize
			if end > len(frontier) {
				end = len(frontier)
			}
			children, err := r.loadSchedulerJobRetryChildren(frontier[i:end])
			if err != nil {
				return err
			}
			for _, child := range children {
				if _, seen := visited[child.ID]; seen {
					continue
				}
				rootID, ok := runToRoot[child.ParentRunID]
				if !ok || rootID == "" {
					continue
				}
				visited[child.ID] = struct{}{}
				runToRoot[child.ID] = rootID
				nextFrontier = append(nextFrontier, child.ID)

				state := recoveryStates[rootID]
				if state == nil {
					continue
				}
				if state.JobName == "" {
					state.JobName = strings.TrimSpace(child.JobName)
				}
				if child.RetryCount > 0 {
					state.HasRetry = true
				}
				childStatus := strings.ToUpper(strings.TrimSpace(child.Status))
				if child.RetryCount > state.MaxRetryCount {
					state.MaxRetryCount = child.RetryCount
					state.FinalStatus = childStatus
					continue
				}
				if child.RetryCount == state.MaxRetryCount && state.MaxRetryCount > 0 {
					if state.FinalStatus != "SUCCESS" && childStatus == "SUCCESS" {
						state.FinalStatus = childStatus
					}
				}
			}
		}
		frontier = nextFrontier
	}

	for _, state := range recoveryStates {
		if !state.HasRetry {
			continue
		}
		jobNameValue := strings.TrimSpace(state.JobName)
		if jobNameValue == "" {
			jobNameValue = "UNKNOWN"
		}
		item := jobStatsMap[jobNameValue]
		if item == nil {
			item = &model.SchedulerJobRetryStat{JobName: jobNameValue}
			jobStatsMap[jobNameValue] = item
		}
		item.RecoveryTotal++
		if strings.EqualFold(state.FinalStatus, "SUCCESS") {
			item.RecoverySuccess++
		}
		if item.RecoveryTotal > 0 {
			item.RecoveryHitRate = roundTo(float64(item.RecoverySuccess)/float64(item.RecoveryTotal), 4)
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) loadSchedulerJobRetryChildren(parentRunIDs []string) ([]schedulerJobRetryChild, error) {
	if len(parentRunIDs) == 0 {
		return []schedulerJobRetryChild{}, nil
	}
	placeholder := strings.TrimSuffix(strings.Repeat("?,", len(parentRunIDs)), ",")
	query := fmt.Sprintf(`
SELECT id, parent_run_id, job_name, status, retry_count
FROM scheduler_job_runs
WHERE parent_run_id IN (%s)
ORDER BY retry_count ASC, id ASC`, placeholder)
	args := make([]interface{}, 0, len(parentRunIDs))
	for _, id := range parentRunIDs {
		args = append(args, id)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]schedulerJobRetryChild, 0)
	for rows.Next() {
		var item schedulerJobRetryChild
		if err := rows.Scan(&item.ID, &item.ParentRunID, &item.JobName, &item.Status, &item.RetryCount); err != nil {
			return nil, err
		}
		item.ID = strings.TrimSpace(item.ID)
		item.ParentRunID = strings.TrimSpace(item.ParentRunID)
		item.JobName = strings.TrimSpace(item.JobName)
		item.Status = strings.TrimSpace(item.Status)
		if item.ID == "" || item.ParentRunID == "" {
			continue
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

type schedulerFailureReasonAggregate struct {
	JobName        string
	Reason         string
	Count          int
	LastOccurredAt time.Time
}

func normalizeSchedulerFailureReason(raw string) string {
	text := strings.TrimSpace(raw)
	if text == "" {
		return "UNKNOWN_ERROR"
	}
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "tushare") && strings.Contains(lower, "token"):
		return "TUSHARE_TOKEN_INVALID_OR_MISSING"
	case strings.Contains(lower, "error 2003"), strings.Contains(lower, "can’t connect"), strings.Contains(lower, "can't connect"):
		return "MYSQL_CONNECTION_FAILED"
	case strings.Contains(lower, "error 1146"):
		return "TABLE_NOT_FOUND"
	case strings.Contains(lower, "timeout"):
		return "UPSTREAM_TIMEOUT"
	case strings.Contains(lower, "status: 5"), strings.Contains(lower, "status code: 5"), strings.Contains(lower, "internal server error"):
		return "UPSTREAM_5XX"
	case strings.Contains(lower, "permission"), strings.Contains(lower, "forbidden"), strings.Contains(lower, "unauthorized"):
		return "PERMISSION_DENIED"
	case strings.Contains(lower, "unknown job"):
		return "UNKNOWN_JOB_NAME"
	default:
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.Join(strings.Fields(text), " ")
		if len(text) > 96 {
			text = text[:96]
		}
		return text
	}
}

type sqlExecer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func (r *MySQLGrowthRepo) applyModuleTargetStatus(execer sqlExecer, module string, targetID string, status string) error {
	var (
		query string
		args  []interface{}
	)
	switch strings.ToUpper(module) {
	case "NEWS":
		query = "UPDATE news_articles SET status = ?, updated_at = ? WHERE id = ?"
		args = []interface{}{status, time.Now(), targetID}
	case "STOCK":
		query = "UPDATE stock_recommendations SET status = ? WHERE id = ?"
		args = []interface{}{status, targetID}
	case "FUTURES":
		query = "UPDATE futures_strategies SET status = ? WHERE id = ?"
		args = []interface{}{status, targetID}
	default:
		return errors.New("unsupported module")
	}
	res, err := execer.Exec(query, args...)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("target not found")
	}
	return nil
}

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
	return resolveMembershipActivationState(memberLevel, kycStatus, vipExpireAt) == "ACTIVE", nil
}

func newID(prefix string) string {
	seq := repoIDSequence.Add(1)
	return fmt.Sprintf("%s_%d_%d", prefix, time.Now().UnixNano(), seq)
}

func nullableString(value string) interface{} {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func formatNullTime(value sql.NullTime) string {
	if value.Valid {
		return value.Time.Format(time.RFC3339)
	}
	return ""
}

type marketRhythmTaskTemplate struct {
	Slot    string
	TaskKey string
}

func defaultMarketRhythmTaskTemplates() []marketRhythmTaskTemplate {
	return []marketRhythmTaskTemplate{
		{Slot: "08:30", TaskKey: "morning_stock_publish"},
		{Slot: "11:30", TaskKey: "midday_news_publish"},
		{Slot: "15:30", TaskKey: "close_tracking_review"},
		{Slot: "周末", TaskKey: "weekend_review_digest"},
	}
}

func normalizeMarketRhythmDate(raw string) (string, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return time.Now().Format("2006-01-02"), nil
	}
	parsed, err := time.Parse("2006-01-02", text)
	if err != nil {
		return "", err
	}
	return parsed.Format("2006-01-02"), nil
}

func normalizeMarketRhythmStatus(raw string) string {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "TODO":
		return "TODO"
	case "IN_PROGRESS":
		return "IN_PROGRESS"
	case "DONE":
		return "DONE"
	case "BLOCKED":
		return "BLOCKED"
	default:
		return ""
	}
}

func normalizeStringList(items []string) []string {
	if len(items) == 0 {
		return nil
	}
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		text := strings.TrimSpace(item)
		if text == "" {
			continue
		}
		if _, exists := seen[text]; exists {
			continue
		}
		seen[text] = struct{}{}
		result = append(result, text)
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func marshalStringList(items []string) (string, error) {
	if len(items) == 0 {
		return "", nil
	}
	payload, err := json.Marshal(items)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func parseStringList(raw sql.NullString) []string {
	text := strings.TrimSpace(raw.String)
	if !raw.Valid || text == "" {
		return nil
	}
	var items []string
	if err := json.Unmarshal([]byte(text), &items); err == nil {
		return normalizeStringList(items)
	}
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	return normalizeStringList(parts)
}

func scanMarketRhythmTask(scanner interface {
	Scan(dest ...interface{}) error
}) (model.MarketRhythmTask, error) {
	var item model.MarketRhythmTask
	var taskDate time.Time
	var sourceLinksJSON sql.NullString
	var completedAt sql.NullTime
	var createdAt sql.NullTime
	var updatedAt sql.NullTime
	if err := scanner.Scan(
		&item.ID,
		&taskDate,
		&item.Slot,
		&item.TaskKey,
		&item.Status,
		&item.Owner,
		&item.Notes,
		&sourceLinksJSON,
		&completedAt,
		&createdAt,
		&updatedAt,
	); err != nil {
		return model.MarketRhythmTask{}, err
	}
	item.Date = taskDate.Format("2006-01-02")
	item.SourceLinks = parseStringList(sourceLinksJSON)
	if completedAt.Valid {
		item.CompletedAt = completedAt.Time.Format(time.RFC3339)
	}
	if createdAt.Valid {
		item.CreatedAt = createdAt.Time.Format(time.RFC3339)
	}
	if updatedAt.Valid {
		item.UpdatedAt = updatedAt.Time.Format(time.RFC3339)
	}
	return item, nil
}

func getMarketRhythmTaskByIDTx(tx *sql.Tx, id string) (model.MarketRhythmTask, error) {
	row := tx.QueryRow(`
SELECT id, task_date, slot, task_key, status, COALESCE(owner, ''), COALESCE(notes, ''), source_links_json, completed_at, created_at, updated_at
FROM market_rhythm_tasks
WHERE id = ?
LIMIT 1`, id)
	return scanMarketRhythmTask(row)
}

func (r *MySQLGrowthRepo) adminGetMarketRhythmTaskByID(id string) (model.MarketRhythmTask, error) {
	row := r.db.QueryRow(`
SELECT id, task_date, slot, task_key, status, COALESCE(owner, ''), COALESCE(notes, ''), source_links_json, completed_at, created_at, updated_at
FROM market_rhythm_tasks
WHERE id = ?
LIMIT 1`, id)
	return scanMarketRhythmTask(row)
}

func marketRhythmCompletedAtValue(existing string, status string) interface{} {
	if normalizeMarketRhythmStatus(status) != "DONE" {
		return nil
	}
	if parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(existing)); err == nil {
		return parsed
	}
	return time.Now()
}

func vipRemainingDays(expireAt sql.NullTime, now time.Time) int {
	if !expireAt.Valid {
		return 0
	}
	if !expireAt.Time.After(now) {
		return 0
	}
	hours := expireAt.Time.Sub(now).Hours()
	if hours <= 0 {
		return 0
	}
	days := int(math.Ceil(hours / 24))
	if days < 0 {
		return 0
	}
	return days
}

func vipStatusFromLevelAndExpire(memberLevel string, expireAt sql.NullTime, now time.Time) string {
	level := strings.ToUpper(strings.TrimSpace(memberLevel))
	if !strings.HasPrefix(level, "VIP") {
		return "FREE"
	}
	if !expireAt.Valid {
		return "ACTIVE"
	}
	if expireAt.Time.After(now) {
		return "ACTIVE"
	}
	return "EXPIRED"
}

func resolveVIPStatusAndDays(memberLevel string, expireAt sql.NullTime) (string, int) {
	now := time.Now()
	return vipStatusFromLevelAndExpire(memberLevel, expireAt, now), vipRemainingDays(expireAt, now)
}

func isVerifiedKYCStatus(status string) bool {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "APPROVED", "VERIFIED":
		return true
	default:
		return false
	}
}

func resolveMembershipActivationState(memberLevel string, kycStatus string, vipExpireAt sql.NullTime) string {
	if vipStatusFromLevelAndExpire(memberLevel, vipExpireAt, time.Now()) != "ACTIVE" {
		return "NON_MEMBER"
	}
	if isVerifiedKYCStatus(kycStatus) {
		return "ACTIVE"
	}
	return "PAID_PENDING_KYC"
}

func parseRepoConfigBool(raw string, fallback bool) bool {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" {
		return fallback
	}
	switch text {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func parseRepoConfigInt(raw string, fallback int) int {
	text := strings.TrimSpace(raw)
	if text == "" {
		return fallback
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return fallback
	}
	return value
}

type stockQuoteCandle struct {
	Symbol         string
	TradeDate      time.Time
	OpenPrice      float64
	HighPrice      float64
	LowPrice       float64
	ClosePrice     float64
	PrevClosePrice float64
	Volume         float64
	Turnover       float64
}

type stockDailyBasicPoint struct {
	Symbol       string
	TradeDate    time.Time
	TurnoverRate float64
	VolumeRatio  float64
	PeTTM        float64
	PB           float64
	TotalMV      float64
	CircMV       float64
	SourceKey    string
}

type stockMoneyflowPoint struct {
	Symbol        string
	TradeDate     time.Time
	NetMFAmount   float64
	BuyLGAmount   float64
	SellLGAmount  float64
	BuyELGAmount  float64
	SellELGAmount float64
	SourceKey     string
}

type stockNewsRawPoint struct {
	SourceKey   string
	Symbol      string
	PublishedAt time.Time
	Title       string
	Content     string
	URL         string
	Sentiment   string
}

type stockNewsSignal struct {
	Heat         int
	PositiveRate float64
}

type quantRiskAggregate struct {
	Sum5  float64
	Cnt5  int
	Hit5  int
	Sum10 float64
	Cnt10 int
	Hit10 int
}

type metricRange struct {
	Min float64
	Max float64
}

type dailyStockCandidate struct {
	Symbol         string
	Name           string
	Score          float64
	RiskLevel      string
	PositionRange  string
	ValidFrom      time.Time
	ValidTo        time.Time
	Status         string
	ReasonSummary  string
	TechScore      float64
	FundScore      float64
	SentimentScore float64
	MoneyFlowScore float64
	TakeProfit     string
	StopLoss       string
	RiskNote       string
}

func normalizeStockSymbolList(symbols []string) []string {
	seen := make(map[string]struct{}, len(symbols))
	result := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func defaultMockStockSymbols() []string {
	return []string{
		"600519.SH", "601318.SH", "600036.SH", "600276.SH", "601012.SH",
		"000333.SZ", "300750.SZ", "002594.SZ", "688981.SH", "601888.SH",
		"000858.SZ", "000001.SZ", "601166.SH", "300015.SZ", "000651.SZ",
	}
}

func symbolSeed(symbol string) float64 {
	sum := 0
	for _, ch := range symbol {
		sum += int(ch)
	}
	return float64((sum%97)+3) / 10.0
}

func mockBasePriceBySymbol(symbol string) float64 {
	if strings.HasSuffix(symbol, ".SZ") {
		return 18 + symbolSeed(symbol)*4.5
	}
	return 25 + symbolSeed(symbol)*5.2
}

func buildMockStockQuotes(symbols []string, days int, sourceKey string) []model.StockMarketQuote {
	now := time.Now()
	tradeDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	items := make([]model.StockMarketQuote, 0, len(symbols)*days)
	for _, symbol := range symbols {
		seed := symbolSeed(symbol)
		prevClose := mockBasePriceBySymbol(symbol)
		for offset := days - 1; offset >= 0; offset-- {
			currentDay := tradeDay.AddDate(0, 0, -offset)
			drift := 0.0008 + 0.0015*math.Sin(float64(days-offset)/11.0+seed/5.0)
			wave := 0.006 * math.Sin(float64(offset)/2.7+seed)
			noise := 0.003 * math.Cos(float64(offset)/1.9+seed*1.8)
			closePrice := prevClose * (1 + drift + wave + noise)
			if closePrice <= 0 {
				closePrice = prevClose
			}
			openPrice := prevClose * (1 + noise*0.6)
			highPrice := math.Max(openPrice, closePrice) * (1 + 0.002 + math.Abs(wave)*0.5)
			lowPrice := math.Min(openPrice, closePrice) * (1 - 0.002 - math.Abs(wave)*0.45)
			if lowPrice <= 0 {
				lowPrice = math.Min(openPrice, closePrice)
			}

			volumeBase := 1_800_000 + seed*260_000
			volumeWave := (1 + 0.35*math.Sin(float64(offset)/3.1+seed*0.7))
			volume := int64(math.Max(200_000, math.Round(volumeBase*volumeWave)))
			turnover := closePrice * float64(volume)

			items = append(items, model.StockMarketQuote{
				Symbol:         symbol,
				TradeDate:      currentDay.Format("2006-01-02"),
				OpenPrice:      roundTo(openPrice, 4),
				HighPrice:      roundTo(highPrice, 4),
				LowPrice:       roundTo(lowPrice, 4),
				ClosePrice:     roundTo(closePrice, 4),
				PrevClosePrice: roundTo(prevClose, 4),
				Volume:         volume,
				Turnover:       roundTo(turnover, 2),
				SourceKey:      sourceKey,
			})
			prevClose = closePrice
		}
	}
	return items
}

func fetchStockQuotesFromEndpoint(endpoint string, sourceKey string, symbols []string, days int, timeoutMS int) ([]model.StockMarketQuote, error) {
	parsed, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil {
		return nil, err
	}
	query := parsed.Query()
	if len(symbols) > 0 {
		query.Set("symbols", strings.Join(symbols, ","))
	}
	if days > 0 {
		query.Set("days", strconv.Itoa(days))
	}
	parsed.RawQuery = query.Encode()
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	resp, err := client.Get(parsed.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("quotes endpoint status: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	type quoteItem struct {
		Symbol         string  `json:"symbol"`
		TradeDate      string  `json:"trade_date"`
		Date           string  `json:"date"`
		OpenPrice      float64 `json:"open_price"`
		HighPrice      float64 `json:"high_price"`
		LowPrice       float64 `json:"low_price"`
		ClosePrice     float64 `json:"close_price"`
		PrevClosePrice float64 `json:"prev_close_price"`
		Volume         float64 `json:"volume"`
		Turnover       float64 `json:"turnover"`
	}
	payload := struct {
		Data  []quoteItem `json:"data"`
		Items []quoteItem `json:"items"`
	}{}
	items := make([]quoteItem, 0)
	if err := json.Unmarshal(body, &payload); err == nil {
		if len(payload.Data) > 0 {
			items = payload.Data
		} else if len(payload.Items) > 0 {
			items = payload.Items
		}
	}
	if len(items) == 0 {
		if err := json.Unmarshal(body, &items); err != nil {
			return nil, err
		}
	}

	symbolFilter := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		symbolFilter[strings.ToUpper(strings.TrimSpace(symbol))] = struct{}{}
	}

	result := make([]model.StockMarketQuote, 0, len(items))
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		if len(symbolFilter) > 0 {
			if _, ok := symbolFilter[symbol]; !ok {
				continue
			}
		}
		dateText := strings.TrimSpace(item.TradeDate)
		if dateText == "" {
			dateText = strings.TrimSpace(item.Date)
		}
		if dateText == "" {
			continue
		}
		tradeDate, err := parseFlexibleDateTime(dateText)
		if err != nil {
			continue
		}
		closePrice := item.ClosePrice
		if closePrice <= 0 {
			continue
		}
		openPrice := item.OpenPrice
		if openPrice <= 0 {
			openPrice = closePrice
		}
		highPrice := item.HighPrice
		if highPrice <= 0 {
			highPrice = math.Max(openPrice, closePrice)
		}
		lowPrice := item.LowPrice
		if lowPrice <= 0 {
			lowPrice = math.Min(openPrice, closePrice)
		}
		prevClose := item.PrevClosePrice
		if prevClose <= 0 {
			prevClose = openPrice
		}
		volume := int64(math.Round(item.Volume))
		if volume < 0 {
			volume = 0
		}
		turnover := item.Turnover
		if turnover <= 0 && volume > 0 {
			turnover = closePrice * float64(volume)
		}
		result = append(result, model.StockMarketQuote{
			Symbol:         symbol,
			TradeDate:      tradeDate.Format("2006-01-02"),
			OpenPrice:      roundTo(openPrice, 4),
			HighPrice:      roundTo(highPrice, 4),
			LowPrice:       roundTo(lowPrice, 4),
			ClosePrice:     roundTo(closePrice, 4),
			PrevClosePrice: roundTo(prevClose, 4),
			Volume:         volume,
			Turnover:       roundTo(turnover, 2),
			SourceKey:      sourceKey,
		})
	}
	return result, nil
}

func fetchStockQuotesFromTushare(token string, sourceKey string, symbols []string, days int, timeoutMS int) ([]model.StockMarketQuote, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	startDate := time.Now().AddDate(0, 0, -(days + 20)).Format("20060102")
	endDate := time.Now().Format("20060102")
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	result := make([]model.StockMarketQuote, 0, len(symbols)*days)

	for _, symbol := range symbols {
		reqBody := map[string]interface{}{
			"api_name": "daily",
			"token":    token,
			"params": map[string]string{
				"ts_code":    symbol,
				"start_date": startDate,
				"end_date":   endDate,
			},
			"fields": "ts_code,trade_date,open,high,low,close,pre_close,vol,amount",
		}
		payload, err := json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequest(http.MethodPost, "https://api.tushare.pro", bytes.NewReader(payload))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}

		body, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			return nil, readErr
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return nil, fmt.Errorf("tushare status: %s", resp.Status)
		}

		parsed := struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				Fields []string        `json:"fields"`
				Items  [][]interface{} `json:"items"`
			} `json:"data"`
		}{}
		if err := json.Unmarshal(body, &parsed); err != nil {
			return nil, err
		}
		if parsed.Code != 0 {
			return nil, fmt.Errorf("tushare error(%s): %s", symbol, strings.TrimSpace(parsed.Msg))
		}
		if len(parsed.Data.Fields) == 0 || len(parsed.Data.Items) == 0 {
			continue
		}

		fieldIndex := make(map[string]int, len(parsed.Data.Fields))
		for idx, field := range parsed.Data.Fields {
			fieldIndex[strings.TrimSpace(field)] = idx
		}

		for _, row := range parsed.Data.Items {
			tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
			if !ok {
				continue
			}
			tradeDateRaw, ok := tushareGetString(row, fieldIndex, "trade_date")
			if !ok {
				continue
			}
			tradeDate, err := time.ParseInLocation("20060102", strings.TrimSpace(tradeDateRaw), time.Local)
			if err != nil {
				continue
			}

			openPrice, ok := tushareGetFloat(row, fieldIndex, "open")
			if !ok {
				continue
			}
			highPrice, ok := tushareGetFloat(row, fieldIndex, "high")
			if !ok {
				continue
			}
			lowPrice, ok := tushareGetFloat(row, fieldIndex, "low")
			if !ok {
				continue
			}
			closePrice, ok := tushareGetFloat(row, fieldIndex, "close")
			if !ok || closePrice <= 0 {
				continue
			}
			prevClose, ok := tushareGetFloat(row, fieldIndex, "pre_close")
			if !ok || prevClose <= 0 {
				prevClose = openPrice
			}
			volLot, _ := tushareGetFloat(row, fieldIndex, "vol")
			amountK, _ := tushareGetFloat(row, fieldIndex, "amount")
			volumeShares := int64(math.Round(volLot * 100))
			if volumeShares < 0 {
				volumeShares = 0
			}
			turnover := amountK * 1000
			if turnover <= 0 && volumeShares > 0 {
				turnover = closePrice * float64(volumeShares)
			}

			result = append(result, model.StockMarketQuote{
				Symbol:         strings.ToUpper(strings.TrimSpace(tsCode)),
				TradeDate:      tradeDate.Format("2006-01-02"),
				OpenPrice:      roundTo(openPrice, 4),
				HighPrice:      roundTo(highPrice, 4),
				LowPrice:       roundTo(lowPrice, 4),
				ClosePrice:     roundTo(closePrice, 4),
				PrevClosePrice: roundTo(prevClose, 4),
				Volume:         volumeShares,
				Turnover:       roundTo(turnover, 2),
				SourceKey:      sourceKey,
			})
		}
	}
	return result, nil
}

type tushareStdResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Fields []string        `json:"fields"`
		Items  [][]interface{} `json:"items"`
	} `json:"data"`
}

func callTushareAPI(client *http.Client, token string, apiName string, params map[string]string, fields string) (tushareStdResponse, error) {
	request := map[string]interface{}{
		"api_name": apiName,
		"token":    token,
		"params":   params,
		"fields":   fields,
	}
	payload, err := json.Marshal(request)
	if err != nil {
		return tushareStdResponse{}, err
	}
	req, err := http.NewRequest(http.MethodPost, "https://api.tushare.pro", bytes.NewReader(payload))
	if err != nil {
		return tushareStdResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return tushareStdResponse{}, err
	}
	body, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		return tushareStdResponse{}, readErr
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return tushareStdResponse{}, fmt.Errorf("tushare status: %s", resp.Status)
	}
	parsed := tushareStdResponse{}
	if err := json.Unmarshal(body, &parsed); err != nil {
		return tushareStdResponse{}, err
	}
	if parsed.Code != 0 {
		return tushareStdResponse{}, fmt.Errorf("tushare error(%s): %s", apiName, strings.TrimSpace(parsed.Msg))
	}
	return parsed, nil
}

func fetchStockDailyBasicsFromTushare(token string, sourceKey string, symbols []string, days int, timeoutMS int) ([]stockDailyBasicPoint, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 120
	}
	startDate := time.Now().AddDate(0, 0, -(days + 20)).Format("20060102")
	endDate := time.Now().Format("20060102")
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	result := make([]stockDailyBasicPoint, 0, len(symbols)*days)
	for _, symbol := range symbols {
		parsed, err := callTushareAPI(client, token, "daily_basic", map[string]string{
			"ts_code":    symbol,
			"start_date": startDate,
			"end_date":   endDate,
		}, "ts_code,trade_date,turnover_rate,volume_ratio,pe_ttm,pb,total_mv,circ_mv")
		if err != nil {
			return nil, err
		}
		fieldIndex := make(map[string]int, len(parsed.Data.Fields))
		for idx, field := range parsed.Data.Fields {
			fieldIndex[strings.TrimSpace(field)] = idx
		}
		for _, row := range parsed.Data.Items {
			tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
			if !ok {
				continue
			}
			tradeDateRaw, ok := tushareGetString(row, fieldIndex, "trade_date")
			if !ok {
				continue
			}
			tradeDate, err := time.ParseInLocation("20060102", tradeDateRaw, time.Local)
			if err != nil {
				continue
			}
			turnoverRate, _ := tushareGetFloat(row, fieldIndex, "turnover_rate")
			volumeRatio, _ := tushareGetFloat(row, fieldIndex, "volume_ratio")
			peTTM, _ := tushareGetFloat(row, fieldIndex, "pe_ttm")
			pb, _ := tushareGetFloat(row, fieldIndex, "pb")
			totalMV, _ := tushareGetFloat(row, fieldIndex, "total_mv")
			circMV, _ := tushareGetFloat(row, fieldIndex, "circ_mv")
			result = append(result, stockDailyBasicPoint{
				Symbol:       strings.ToUpper(strings.TrimSpace(tsCode)),
				TradeDate:    tradeDate,
				TurnoverRate: roundTo(turnoverRate, 4),
				VolumeRatio:  roundTo(volumeRatio, 4),
				PeTTM:        roundTo(peTTM, 4),
				PB:           roundTo(pb, 4),
				TotalMV:      roundTo(totalMV, 4),
				CircMV:       roundTo(circMV, 4),
				SourceKey:    sourceKey,
			})
		}
	}
	return result, nil
}

func fetchStockMoneyflowsFromTushare(token string, sourceKey string, symbols []string, days int, timeoutMS int) ([]stockMoneyflowPoint, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 120
	}
	startDate := time.Now().AddDate(0, 0, -(days + 20)).Format("20060102")
	endDate := time.Now().Format("20060102")
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	result := make([]stockMoneyflowPoint, 0, len(symbols)*days)
	for _, symbol := range symbols {
		parsed, err := callTushareAPI(client, token, "moneyflow", map[string]string{
			"ts_code":    symbol,
			"start_date": startDate,
			"end_date":   endDate,
		}, "ts_code,trade_date,net_mf_amount,buy_lg_amount,sell_lg_amount,buy_elg_amount,sell_elg_amount")
		if err != nil {
			return nil, err
		}
		fieldIndex := make(map[string]int, len(parsed.Data.Fields))
		for idx, field := range parsed.Data.Fields {
			fieldIndex[strings.TrimSpace(field)] = idx
		}
		for _, row := range parsed.Data.Items {
			tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
			if !ok {
				continue
			}
			tradeDateRaw, ok := tushareGetString(row, fieldIndex, "trade_date")
			if !ok {
				continue
			}
			tradeDate, err := time.ParseInLocation("20060102", tradeDateRaw, time.Local)
			if err != nil {
				continue
			}
			netMFAmount, _ := tushareGetFloat(row, fieldIndex, "net_mf_amount")
			buyLGAmount, _ := tushareGetFloat(row, fieldIndex, "buy_lg_amount")
			sellLGAmount, _ := tushareGetFloat(row, fieldIndex, "sell_lg_amount")
			buyELGAmount, _ := tushareGetFloat(row, fieldIndex, "buy_elg_amount")
			sellELGAmount, _ := tushareGetFloat(row, fieldIndex, "sell_elg_amount")
			result = append(result, stockMoneyflowPoint{
				Symbol:        strings.ToUpper(strings.TrimSpace(tsCode)),
				TradeDate:     tradeDate,
				NetMFAmount:   roundTo(netMFAmount, 4),
				BuyLGAmount:   roundTo(buyLGAmount, 4),
				SellLGAmount:  roundTo(sellLGAmount, 4),
				BuyELGAmount:  roundTo(buyELGAmount, 4),
				SellELGAmount: roundTo(sellELGAmount, 4),
				SourceKey:     sourceKey,
			})
		}
	}
	return result, nil
}

func fetchStockNewsFromTushare(token string, sourceKey string, symbols []string, days int, timeoutMS int) ([]stockNewsRawPoint, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("tushare token not configured")
	}
	if timeoutMS <= 0 {
		timeoutMS = 12000
	}
	if days <= 0 {
		days = 14
	}
	startDate := time.Now().AddDate(0, 0, -days).Format("20060102")
	endDate := time.Now().Format("20060102")
	client := &http.Client{Timeout: time.Duration(timeoutMS) * time.Millisecond}
	result := make([]stockNewsRawPoint, 0, len(symbols)*3)
	for _, symbol := range symbols {
		parsed, err := callTushareAPI(client, token, "anns_d", map[string]string{
			"ts_code":    symbol,
			"start_date": startDate,
			"end_date":   endDate,
		}, "ts_code,ann_date,title,url,rec_time")
		if err != nil {
			return nil, err
		}
		fieldIndex := make(map[string]int, len(parsed.Data.Fields))
		for idx, field := range parsed.Data.Fields {
			fieldIndex[strings.TrimSpace(field)] = idx
		}
		for _, row := range parsed.Data.Items {
			tsCode, ok := tushareGetString(row, fieldIndex, "ts_code")
			if !ok {
				continue
			}
			title, ok := tushareGetString(row, fieldIndex, "title")
			if !ok {
				continue
			}
			urlText, _ := tushareGetString(row, fieldIndex, "url")
			recTime, _ := tushareGetString(row, fieldIndex, "rec_time")
			annDate, _ := tushareGetString(row, fieldIndex, "ann_date")
			publishedAt := parseTushareNewsTime(recTime, annDate)
			sentiment := classifyNewsSentiment(title)
			result = append(result, stockNewsRawPoint{
				SourceKey:   sourceKey,
				Symbol:      strings.ToUpper(strings.TrimSpace(tsCode)),
				PublishedAt: publishedAt,
				Title:       strings.TrimSpace(title),
				URL:         strings.TrimSpace(urlText),
				Sentiment:   sentiment,
			})
		}
	}
	return result, nil
}

func parseTushareNewsTime(recTime string, annDate string) time.Time {
	recTime = strings.TrimSpace(recTime)
	annDate = strings.TrimSpace(annDate)
	if recTime != "" {
		if ts, err := time.ParseInLocation("2006-01-02 15:04:05", recTime, time.Local); err == nil {
			return ts
		}
		if ts, err := time.ParseInLocation("2006-01-02 15:04", recTime, time.Local); err == nil {
			return ts
		}
	}
	if annDate != "" {
		if ts, err := time.ParseInLocation("20060102", annDate, time.Local); err == nil {
			return ts
		}
	}
	return time.Now()
}

func classifyNewsSentiment(title string) string {
	text := strings.ToLower(strings.TrimSpace(title))
	if text == "" {
		return "NEUTRAL"
	}
	positiveKeywords := []string{"增长", "预增", "中标", "回购", "增持", "突破", "创新高", "上调", "盈利", "签约", "positive", "upgrade"}
	negativeKeywords := []string{"下滑", "预减", "亏损", "违约", "减持", "处罚", "风险", "暴跌", "下调", "诉讼", "negative", "downgrade"}
	positiveHit := 0
	negativeHit := 0
	for _, keyword := range positiveKeywords {
		if strings.Contains(text, keyword) {
			positiveHit++
		}
	}
	for _, keyword := range negativeKeywords {
		if strings.Contains(text, keyword) {
			negativeHit++
		}
	}
	if positiveHit > negativeHit {
		return "POSITIVE"
	}
	if negativeHit > positiveHit {
		return "NEGATIVE"
	}
	return "NEUTRAL"
}

func tushareGetString(row []interface{}, fieldIndex map[string]int, key string) (string, bool) {
	index, ok := fieldIndex[key]
	if !ok || index < 0 || index >= len(row) {
		return "", false
	}
	text := strings.TrimSpace(fmt.Sprintf("%v", row[index]))
	if text == "" || text == "<nil>" {
		return "", false
	}
	return text, true
}

func tushareGetFloat(row []interface{}, fieldIndex map[string]int, key string) (float64, bool) {
	index, ok := fieldIndex[key]
	if !ok || index < 0 || index >= len(row) {
		return 0, false
	}
	value := row[index]
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case json.Number:
		num, err := v.Float64()
		if err != nil {
			return 0, false
		}
		return num, true
	case string:
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			return 0, false
		}
		num, err := strconv.ParseFloat(trimmed, 64)
		if err != nil {
			return 0, false
		}
		return num, true
	default:
		text := strings.TrimSpace(fmt.Sprintf("%v", value))
		if text == "" || text == "<nil>" {
			return 0, false
		}
		num, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return 0, false
		}
		return num, true
	}
}

func (r *MySQLGrowthRepo) upsertStockMarketQuotes(items []model.StockMarketQuote) (int, error) {
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		tradeDate, err := parseFlexibleDateTime(item.TradeDate)
		if err != nil {
			return affected, err
		}
		if item.ClosePrice <= 0 {
			continue
		}
		openPrice := item.OpenPrice
		if openPrice <= 0 {
			openPrice = item.ClosePrice
		}
		highPrice := item.HighPrice
		if highPrice <= 0 {
			highPrice = math.Max(openPrice, item.ClosePrice)
		}
		lowPrice := item.LowPrice
		if lowPrice <= 0 {
			lowPrice = math.Min(openPrice, item.ClosePrice)
		}
		prevClose := item.PrevClosePrice
		if prevClose <= 0 {
			prevClose = openPrice
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if sourceKey == "" {
			sourceKey = "MOCK"
		}
		_, err = r.db.Exec(`
INSERT INTO stock_market_quotes
  (id, symbol, trade_date, open_price, high_price, low_price, close_price, prev_close_price, volume, turnover, source_key, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  open_price = VALUES(open_price),
  high_price = VALUES(high_price),
  low_price = VALUES(low_price),
  close_price = VALUES(close_price),
  prev_close_price = VALUES(prev_close_price),
  volume = VALUES(volume),
  turnover = VALUES(turnover),
  source_key = VALUES(source_key),
  updated_at = VALUES(updated_at)`,
			newID("smq"),
			symbol,
			tradeDate.Format("2006-01-02"),
			openPrice,
			highPrice,
			lowPrice,
			item.ClosePrice,
			prevClose,
			item.Volume,
			item.Turnover,
			sourceKey,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) upsertStockDailyBasics(items []stockDailyBasicPoint) (int, error) {
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if sourceKey == "" {
			sourceKey = "TUSHARE"
		}
		_, err := r.db.Exec(`
INSERT INTO stock_daily_basic
  (id, symbol, trade_date, turnover_rate, volume_ratio, pe_ttm, pb, total_mv, circ_mv, source_key, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  turnover_rate = VALUES(turnover_rate),
  volume_ratio = VALUES(volume_ratio),
  pe_ttm = VALUES(pe_ttm),
  pb = VALUES(pb),
  total_mv = VALUES(total_mv),
  circ_mv = VALUES(circ_mv),
  source_key = VALUES(source_key),
  updated_at = VALUES(updated_at)`,
			newID("sdb"),
			symbol,
			item.TradeDate.Format("2006-01-02"),
			item.TurnoverRate,
			item.VolumeRatio,
			item.PeTTM,
			item.PB,
			item.TotalMV,
			item.CircMV,
			sourceKey,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) upsertStockMoneyflows(items []stockMoneyflowPoint) (int, error) {
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		if symbol == "" {
			continue
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if sourceKey == "" {
			sourceKey = "TUSHARE"
		}
		_, err := r.db.Exec(`
INSERT INTO stock_moneyflow_daily
  (id, symbol, trade_date, net_mf_amount, buy_lg_amount, sell_lg_amount, buy_elg_amount, sell_elg_amount, source_key, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  net_mf_amount = VALUES(net_mf_amount),
  buy_lg_amount = VALUES(buy_lg_amount),
  sell_lg_amount = VALUES(sell_lg_amount),
  buy_elg_amount = VALUES(buy_elg_amount),
  sell_elg_amount = VALUES(sell_elg_amount),
  source_key = VALUES(source_key),
  updated_at = VALUES(updated_at)`,
			newID("smf"),
			symbol,
			item.TradeDate.Format("2006-01-02"),
			item.NetMFAmount,
			item.BuyLGAmount,
			item.SellLGAmount,
			item.BuyELGAmount,
			item.SellELGAmount,
			sourceKey,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) upsertStockNewsRaw(items []stockNewsRawPoint) (int, error) {
	affected := 0
	for _, item := range items {
		symbol := strings.ToUpper(strings.TrimSpace(item.Symbol))
		title := strings.TrimSpace(item.Title)
		if symbol == "" || title == "" || item.PublishedAt.IsZero() {
			continue
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.SourceKey))
		if sourceKey == "" {
			sourceKey = "TUSHARE"
		}
		sentiment := strings.ToUpper(strings.TrimSpace(item.Sentiment))
		if sentiment == "" {
			sentiment = "NEUTRAL"
		}
		_, err := r.db.Exec(`
INSERT INTO stock_news_raw
  (id, source_key, symbol, published_at, title, content, url, sentiment, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  content = VALUES(content),
  url = VALUES(url),
  sentiment = VALUES(sentiment),
  updated_at = VALUES(updated_at)`,
			newID("snr"),
			sourceKey,
			symbol,
			item.PublishedAt,
			title,
			nullableString(item.Content),
			nullableString(item.URL),
			sentiment,
			time.Now(),
			time.Now(),
		)
		if err != nil {
			return affected, err
		}
		affected++
	}
	return affected, nil
}

func (r *MySQLGrowthRepo) loadLatestStockDailyBasics(sinceDate time.Time) (map[string]stockDailyBasicPoint, error) {
	rows, err := r.db.Query(`
SELECT t.symbol, t.trade_date, t.turnover_rate, t.volume_ratio, t.pe_ttm, t.pb, t.total_mv, t.circ_mv, t.source_key
FROM stock_daily_basic t
INNER JOIN (
  SELECT symbol, MAX(trade_date) AS latest_trade_date
  FROM stock_daily_basic
  WHERE trade_date >= ?
  GROUP BY symbol
) latest
ON latest.symbol = t.symbol AND latest.latest_trade_date = t.trade_date`, sinceDate.Format("2006-01-02"))
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]stockDailyBasicPoint{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]stockDailyBasicPoint)
	for rows.Next() {
		var (
			item         stockDailyBasicPoint
			turnoverRate sql.NullFloat64
			volumeRatio  sql.NullFloat64
			peTTM        sql.NullFloat64
			pb           sql.NullFloat64
			totalMV      sql.NullFloat64
			circMV       sql.NullFloat64
			sourceKey    sql.NullString
		)
		if err := rows.Scan(
			&item.Symbol,
			&item.TradeDate,
			&turnoverRate,
			&volumeRatio,
			&peTTM,
			&pb,
			&totalMV,
			&circMV,
			&sourceKey,
		); err != nil {
			return nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
		item.TurnoverRate = sqlNullFloat(turnoverRate)
		item.VolumeRatio = sqlNullFloat(volumeRatio)
		item.PeTTM = sqlNullFloat(peTTM)
		item.PB = sqlNullFloat(pb)
		item.TotalMV = sqlNullFloat(totalMV)
		item.CircMV = sqlNullFloat(circMV)
		if sourceKey.Valid {
			item.SourceKey = strings.ToUpper(strings.TrimSpace(sourceKey.String))
		}
		if item.Symbol == "" {
			continue
		}
		result[item.Symbol] = item
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadLatestStockMoneyflows(sinceDate time.Time) (map[string]stockMoneyflowPoint, error) {
	rows, err := r.db.Query(`
SELECT t.symbol, t.trade_date, t.net_mf_amount, t.buy_lg_amount, t.sell_lg_amount, t.buy_elg_amount, t.sell_elg_amount, t.source_key
FROM stock_moneyflow_daily t
INNER JOIN (
  SELECT symbol, MAX(trade_date) AS latest_trade_date
  FROM stock_moneyflow_daily
  WHERE trade_date >= ?
  GROUP BY symbol
) latest
ON latest.symbol = t.symbol AND latest.latest_trade_date = t.trade_date`, sinceDate.Format("2006-01-02"))
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]stockMoneyflowPoint{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]stockMoneyflowPoint)
	for rows.Next() {
		var (
			item      stockMoneyflowPoint
			netMF     sql.NullFloat64
			buyLG     sql.NullFloat64
			sellLG    sql.NullFloat64
			buyELG    sql.NullFloat64
			sellELG   sql.NullFloat64
			sourceKey sql.NullString
		)
		if err := rows.Scan(
			&item.Symbol,
			&item.TradeDate,
			&netMF,
			&buyLG,
			&sellLG,
			&buyELG,
			&sellELG,
			&sourceKey,
		); err != nil {
			return nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
		item.NetMFAmount = sqlNullFloat(netMF)
		item.BuyLGAmount = sqlNullFloat(buyLG)
		item.SellLGAmount = sqlNullFloat(sellLG)
		item.BuyELGAmount = sqlNullFloat(buyELG)
		item.SellELGAmount = sqlNullFloat(sellELG)
		if sourceKey.Valid {
			item.SourceKey = strings.ToUpper(strings.TrimSpace(sourceKey.String))
		}
		if item.Symbol == "" {
			continue
		}
		result[item.Symbol] = item
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadStockNewsSignals(sinceTime time.Time) (map[string]stockNewsSignal, error) {
	rows, err := r.db.Query(`
SELECT symbol,
       COUNT(*) AS heat,
       SUM(CASE WHEN sentiment = 'POSITIVE' THEN 1 ELSE 0 END) AS positive_cnt
FROM stock_news_raw
WHERE published_at >= ?
GROUP BY symbol`, sinceTime)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]stockNewsSignal{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]stockNewsSignal)
	for rows.Next() {
		var (
			symbol      string
			heat        int
			positiveCnt int
		)
		if err := rows.Scan(&symbol, &heat, &positiveCnt); err != nil {
			return nil, err
		}
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" || heat <= 0 {
			continue
		}
		result[symbol] = stockNewsSignal{
			Heat:         heat,
			PositiveRate: float64(positiveCnt) / float64(heat),
		}
	}
	return result, nil
}

func (r *MySQLGrowthRepo) persistStockQuantSnapshots(items []model.StockQuantScore) error {
	if len(items) == 0 {
		return nil
	}
	now := time.Now()
	for _, item := range items {
		tradeDate, err := parseFlexibleDateTime(item.TradeDate)
		if err != nil {
			continue
		}
		reasonsJSON, _ := json.Marshal(item.Reasons)
		payloadJSON, _ := json.Marshal(item)
		_, err = r.db.Exec(`
INSERT INTO stock_factor_snapshot
  (id, symbol, trade_date, total_score, trend_score, flow_score, value_score, news_score, risk_level, reason_summary, reasons_json, source_key, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  total_score = VALUES(total_score),
  trend_score = VALUES(trend_score),
  flow_score = VALUES(flow_score),
  value_score = VALUES(value_score),
  news_score = VALUES(news_score),
  risk_level = VALUES(risk_level),
  reason_summary = VALUES(reason_summary),
  reasons_json = VALUES(reasons_json),
  source_key = VALUES(source_key),
  updated_at = VALUES(updated_at)`,
			newID("sfs"),
			item.Symbol,
			tradeDate.Format("2006-01-02"),
			item.Score,
			item.TrendScore,
			item.FlowScore,
			item.ValueScore,
			item.NewsScore,
			item.RiskLevel,
			nullableString(item.ReasonSummary),
			string(reasonsJSON),
			"QUANT_V2",
			now,
			now,
		)
		if err != nil {
			return err
		}
		_, err = r.db.Exec(`
INSERT INTO stock_rank_daily
  (id, trade_date, rank_no, symbol, name, total_score, risk_level, reason_summary, payload_json, created_at, updated_at)
VALUES
  (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE
  rank_no = VALUES(rank_no),
  name = VALUES(name),
  total_score = VALUES(total_score),
  risk_level = VALUES(risk_level),
  reason_summary = VALUES(reason_summary),
  payload_json = VALUES(payload_json),
  updated_at = VALUES(updated_at)`,
			newID("srd"),
			tradeDate.Format("2006-01-02"),
			item.Rank,
			item.Symbol,
			nullableString(item.Name),
			item.Score,
			item.RiskLevel,
			nullableString(item.ReasonSummary),
			string(payloadJSON),
			now,
			now,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MySQLGrowthRepo) loadStockQuotesBySymbol(sinceDate time.Time) (map[string][]stockQuoteCandle, error) {
	rows, err := r.db.Query(`
SELECT symbol, trade_date, open_price, high_price, low_price, close_price, prev_close_price, volume, turnover
FROM stock_market_quotes
WHERE trade_date >= ?
ORDER BY symbol ASC, trade_date ASC`, sinceDate.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string][]stockQuoteCandle)
	for rows.Next() {
		var item stockQuoteCandle
		var prevClose sql.NullFloat64
		var turnover sql.NullFloat64
		if err := rows.Scan(
			&item.Symbol,
			&item.TradeDate,
			&item.OpenPrice,
			&item.HighPrice,
			&item.LowPrice,
			&item.ClosePrice,
			&prevClose,
			&item.Volume,
			&turnover,
		); err != nil {
			return nil, err
		}
		item.Symbol = strings.ToUpper(strings.TrimSpace(item.Symbol))
		if prevClose.Valid {
			item.PrevClosePrice = prevClose.Float64
		}
		if turnover.Valid {
			item.Turnover = turnover.Float64
		}
		if item.Symbol == "" || item.ClosePrice <= 0 {
			continue
		}
		result[item.Symbol] = append(result[item.Symbol], item)
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadBenchmarkQuoteSeries(sinceDate time.Time) ([]stockQuoteCandle, error) {
	benchmarkSymbols := []string{
		"000300.SH",
		"399001.SZ",
		"000001.SH",
		"399006.SZ",
	}
	symbols := normalizeStockSymbolList(benchmarkSymbols)
	placeholder := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	args := make([]interface{}, 0, len(symbols)+1)
	args = append(args, sinceDate.Format("2006-01-02"))
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	query := fmt.Sprintf(`
SELECT symbol, trade_date, open_price, high_price, low_price, close_price, prev_close_price, volume, turnover
FROM stock_market_quotes
WHERE trade_date >= ? AND symbol IN (%s)
ORDER BY trade_date ASC, symbol ASC`, placeholder)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		if isTableNotFoundError(err) {
			return []stockQuoteCandle{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	seriesByDate := make(map[string][]float64)
	for rows.Next() {
		var (
			symbol     string
			tradeDate  time.Time
			openPrice  float64
			highPrice  float64
			lowPrice   float64
			closePrice float64
			prevClose  sql.NullFloat64
			volume     sql.NullInt64
			turnover   sql.NullFloat64
		)
		if err := rows.Scan(&symbol, &tradeDate, &openPrice, &highPrice, &lowPrice, &closePrice, &prevClose, &volume, &turnover); err != nil {
			return nil, err
		}
		if closePrice <= 0 {
			continue
		}
		key := tradeDate.Format("2006-01-02")
		seriesByDate[key] = append(seriesByDate[key], closePrice)
	}
	dateKeys := make([]string, 0, len(seriesByDate))
	for date := range seriesByDate {
		dateKeys = append(dateKeys, date)
	}
	sort.Strings(dateKeys)
	result := make([]stockQuoteCandle, 0, len(dateKeys))
	prev := 0.0
	for _, date := range dateKeys {
		values := seriesByDate[date]
		if len(values) == 0 {
			continue
		}
		meanClose := avgFloat(values)
		if meanClose <= 0 {
			continue
		}
		tradeDate, err := time.ParseInLocation("2006-01-02", date, time.Local)
		if err != nil {
			continue
		}
		if prev <= 0 {
			prev = meanClose
		}
		result = append(result, stockQuoteCandle{
			Symbol:         "BENCHMARK",
			TradeDate:      tradeDate,
			OpenPrice:      meanClose,
			HighPrice:      meanClose,
			LowPrice:       meanClose,
			ClosePrice:     meanClose,
			PrevClosePrice: prev,
		})
		prev = meanClose
	}
	return result, nil
}

func (r *MySQLGrowthRepo) loadQuantRiskLevelByDateSymbol(sinceDate time.Time, topN int) (map[string]string, error) {
	rows, err := r.db.Query(`
SELECT r.trade_date, r.symbol, COALESCE(NULLIF(s.risk_level, ''), 'MEDIUM') AS risk_level
FROM stock_rank_daily r
LEFT JOIN stock_factor_snapshot s
  ON s.trade_date = r.trade_date AND s.symbol = r.symbol
WHERE r.trade_date >= ? AND r.rank_no <= ?`, sinceDate.Format("2006-01-02"), topN)
	if err != nil {
		if isTableNotFoundError(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]string)
	for rows.Next() {
		var (
			tradeDate time.Time
			symbol    string
			riskLevel sql.NullString
		)
		if err := rows.Scan(&tradeDate, &symbol, &riskLevel); err != nil {
			return nil, err
		}
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if symbol == "" {
			continue
		}
		risk := "MEDIUM"
		if riskLevel.Valid {
			risk = resolveQuantRiskLevel(riskLevel.String)
		}
		result[tradeDate.Format("2006-01-02")+"|"+symbol] = risk
	}
	return result, nil
}

func resolveQuantRiskLevel(raw string) string {
	risk := strings.ToUpper(strings.TrimSpace(raw))
	switch risk {
	case "LOW", "MEDIUM", "HIGH":
		return risk
	default:
		return "MEDIUM"
	}
}

func buildQuantRiskPerformanceItems(aggMap map[string]*quantRiskAggregate) []model.StockQuantRiskPerformance {
	if len(aggMap) == 0 {
		return []model.StockQuantRiskPerformance{}
	}
	orderedRiskLevels := []string{"LOW", "MEDIUM", "HIGH"}
	items := make([]model.StockQuantRiskPerformance, 0, len(orderedRiskLevels))
	for _, riskLevel := range orderedRiskLevels {
		agg := aggMap[riskLevel]
		if agg == nil {
			continue
		}
		item := model.StockQuantRiskPerformance{
			RiskLevel:   riskLevel,
			SampleCount: agg.Cnt5,
		}
		if agg.Cnt5 > 0 {
			item.AvgReturn5 = roundTo(agg.Sum5/float64(agg.Cnt5), 4)
			item.HitRate5 = roundTo(float64(agg.Hit5)/float64(agg.Cnt5), 4)
		}
		if agg.Cnt10 > 0 {
			item.AvgReturn10 = roundTo(agg.Sum10/float64(agg.Cnt10), 4)
			item.HitRate10 = roundTo(float64(agg.Hit10)/float64(agg.Cnt10), 4)
		}
		items = append(items, item)
	}
	return items
}

func buildQuantRotationItems(tradeDates []string, dayTopSymbols map[string][]string) []model.StockQuantRotationPoint {
	if len(tradeDates) == 0 {
		return []model.StockQuantRotationPoint{}
	}
	result := make([]model.StockQuantRotationPoint, 0, len(tradeDates))
	prevSymbols := []string{}
	prevSet := make(map[string]struct{})
	for _, tradeDate := range tradeDates {
		currSymbols := uniqueUpperSymbols(dayTopSymbols[tradeDate])
		currSet := make(map[string]struct{}, len(currSymbols))
		for _, symbol := range currSymbols {
			currSet[symbol] = struct{}{}
		}
		entered := make([]string, 0)
		for _, symbol := range currSymbols {
			if _, ok := prevSet[symbol]; !ok {
				entered = append(entered, symbol)
			}
		}
		exited := make([]string, 0)
		for _, symbol := range prevSymbols {
			if _, ok := currSet[symbol]; !ok {
				exited = append(exited, symbol)
			}
		}
		stayed := len(currSymbols) - len(entered)
		if stayed < 0 {
			stayed = 0
		}
		result = append(result, model.StockQuantRotationPoint{
			TradeDate:    tradeDate,
			TopSymbols:   currSymbols,
			Entered:      entered,
			Exited:       exited,
			StayedCount:  stayed,
			ChangedCount: len(entered) + len(exited),
		})
		prevSymbols = currSymbols
		prevSet = currSet
	}
	return result
}

func uniqueUpperSymbols(symbols []string) []string {
	seen := make(map[string]struct{}, len(symbols))
	result := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func calcForwardReturn(quotes []stockQuoteCandle, tradeDate time.Time, horizon int) (float64, bool) {
	if horizon <= 0 || len(quotes) == 0 {
		return 0, false
	}
	indexByDate := make(map[string]int, len(quotes))
	for idx, quote := range quotes {
		indexByDate[quote.TradeDate.Format("2006-01-02")] = idx
	}
	key := tradeDate.Format("2006-01-02")
	startIndex, ok := indexByDate[key]
	if !ok {
		return 0, false
	}
	targetIndex := startIndex + horizon
	if targetIndex >= len(quotes) {
		return 0, false
	}
	startClose := quotes[startIndex].ClosePrice
	targetClose := quotes[targetIndex].ClosePrice
	if startClose <= 0 || targetClose <= 0 {
		return 0, false
	}
	return targetClose/startClose - 1, true
}

func buildStockQuantScore(symbol string, quotes []stockQuoteCandle) (model.StockQuantScore, bool) {
	if len(quotes) < 25 {
		return model.StockQuantScore{}, false
	}
	lastIndex := len(quotes) - 1
	index5 := lastIndex - 5
	index20 := lastIndex - 20
	if index5 < 0 || index20 < 0 {
		return model.StockQuantScore{}, false
	}
	latest := quotes[lastIndex]
	base5 := quotes[index5].ClosePrice
	base20 := quotes[index20].ClosePrice
	if base5 <= 0 || base20 <= 0 || latest.ClosePrice <= 0 {
		return model.StockQuantScore{}, false
	}

	momentum5 := (latest.ClosePrice/base5 - 1) * 100
	momentum20 := (latest.ClosePrice/base20 - 1) * 100

	returns := make([]float64, 0, 20)
	for idx := index20 + 1; idx <= lastIndex; idx++ {
		prev := quotes[idx-1].ClosePrice
		curr := quotes[idx].ClosePrice
		if prev <= 0 || curr <= 0 {
			continue
		}
		returns = append(returns, curr/prev-1)
	}
	if len(returns) < 10 {
		return model.StockQuantScore{}, false
	}
	volatility20 := calculateStdDev(returns) * 100

	volumeSum := 0.0
	volumeCount := 0
	startVolume := maxInt(0, lastIndex-20)
	for idx := startVolume; idx < lastIndex; idx++ {
		if quotes[idx].Volume <= 0 {
			continue
		}
		volumeSum += quotes[idx].Volume
		volumeCount++
	}
	avgVolume := volumeSum
	if volumeCount > 0 {
		avgVolume = volumeSum / float64(volumeCount)
	}
	volumeRatio := 1.0
	if avgVolume > 0 && latest.Volume > 0 {
		volumeRatio = latest.Volume / avgVolume
	}

	startDrawdown := maxInt(0, len(quotes)-20)
	peak := quotes[startDrawdown].ClosePrice
	maxDrawdown := 0.0
	for idx := startDrawdown; idx <= lastIndex; idx++ {
		closePrice := quotes[idx].ClosePrice
		if closePrice > peak {
			peak = closePrice
		}
		if peak > 0 {
			drawdown := (peak - closePrice) / peak * 100
			if drawdown > maxDrawdown {
				maxDrawdown = drawdown
			}
		}
	}

	ma5 := 0.0
	for idx := lastIndex - 4; idx <= lastIndex; idx++ {
		ma5 += quotes[idx].ClosePrice
	}
	ma5 /= 5
	ma20 := 0.0
	for idx := lastIndex - 19; idx <= lastIndex; idx++ {
		ma20 += quotes[idx].ClosePrice
	}
	ma20 /= 20
	trendStrength := 0.0
	if ma20 > 0 {
		trendStrength = (ma5/ma20 - 1) * 100
	}

	return model.StockQuantScore{
		Symbol:        symbol,
		TradeDate:     latest.TradeDate.Format("2006-01-02"),
		ClosePrice:    latest.ClosePrice,
		Momentum5:     momentum5,
		Momentum20:    momentum20,
		Volatility20:  volatility20,
		VolumeRatio:   volumeRatio,
		Drawdown20:    maxDrawdown,
		TrendStrength: trendStrength,
		RiskLevel:     classifyQuantRisk(volatility20, maxDrawdown),
	}, true
}

func classifyQuantRisk(volatility20 float64, drawdown20 float64) string {
	if volatility20 >= 4.0 || drawdown20 >= 12 {
		return "HIGH"
	}
	if volatility20 >= 2.5 || drawdown20 >= 7 {
		return "MEDIUM"
	}
	return "LOW"
}

func quantRiskPenalty(item model.StockQuantScore) float64 {
	penalty := 0.0
	if item.Volatility20 >= 5.0 {
		penalty += 4.0
	}
	if item.Drawdown20 >= 16 {
		penalty += 5.0
	}
	if item.NetMFAmount < 0 {
		penalty += 2.5
	}
	if item.PeTTM > 0 && item.PeTTM >= 60 {
		penalty += 2.0
	}
	if item.NewsHeat >= 3 && item.PositiveNewsRate > 0 && item.PositiveNewsRate < 0.3 {
		penalty += 2.0
	}
	return penalty
}

func passesQuantRiskGate(item model.StockQuantScore) bool {
	if item.ClosePrice <= 0 {
		return false
	}
	if item.Volatility20 >= 9.0 {
		return false
	}
	if item.Drawdown20 >= 25 {
		return false
	}
	if item.VolumeRatio <= 0.15 {
		return false
	}
	if item.Momentum20 <= -12 {
		return false
	}
	if item.PeTTM > 0 && item.PeTTM >= 200 {
		return false
	}
	if item.PB > 0 && item.PB >= 30 {
		return false
	}
	if item.NewsHeat >= 5 && item.PositiveNewsRate > 0 && item.PositiveNewsRate < 0.2 {
		return false
	}
	return true
}

func quantBucketBySymbol(symbol string) string {
	normalized := strings.ToUpper(strings.TrimSpace(symbol))
	if strings.HasPrefix(normalized, "68") && strings.HasSuffix(normalized, ".SH") {
		return "KCB"
	}
	if strings.HasPrefix(normalized, "30") && strings.HasSuffix(normalized, ".SZ") {
		return "CYB"
	}
	if strings.HasPrefix(normalized, "60") && strings.HasSuffix(normalized, ".SH") {
		return "MAIN_SH"
	}
	if (strings.HasPrefix(normalized, "00") || strings.HasPrefix(normalized, "002")) && strings.HasSuffix(normalized, ".SZ") {
		return "MAIN_SZ"
	}
	if strings.HasSuffix(normalized, ".BJ") {
		return "BJ"
	}
	if strings.HasSuffix(normalized, ".SH") {
		return "SH"
	}
	if strings.HasSuffix(normalized, ".SZ") {
		return "SZ"
	}
	return "OTHER"
}

func selectDiversifiedQuantTop(items []model.StockQuantScore, limit int) []model.StockQuantScore {
	if len(items) <= limit {
		return items
	}
	capByBucket := maxInt(1, int(math.Ceil(float64(limit)*0.4)))
	selected := make([]model.StockQuantScore, 0, limit)
	overflow := make([]model.StockQuantScore, 0, len(items))
	bucketCount := make(map[string]int)
	for _, item := range items {
		bucket := quantBucketBySymbol(item.Symbol)
		if bucketCount[bucket] < capByBucket {
			selected = append(selected, item)
			bucketCount[bucket]++
		} else {
			overflow = append(overflow, item)
		}
		if len(selected) >= limit {
			return selected
		}
	}
	for _, item := range overflow {
		if len(selected) >= limit {
			break
		}
		selected = append(selected, item)
	}
	if len(selected) > limit {
		selected = selected[:limit]
	}
	return selected
}

func calculateStdDev(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	mean := 0.0
	for _, item := range values {
		mean += item
	}
	mean /= float64(len(values))
	sumSq := 0.0
	for _, item := range values {
		diff := item - mean
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq / float64(len(values)))
}

func avgFloat(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func calcMaxDrawdownFromReturns(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	equity := 1.0
	peak := 1.0
	maxDrawdown := 0.0
	for _, dayReturn := range returns {
		equity *= 1 + dayReturn
		if equity > peak {
			peak = equity
		}
		if peak <= 0 {
			continue
		}
		drawdown := (peak - equity) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}
	return maxDrawdown
}

func buildMetricRange(items []model.StockQuantScore, selector func(model.StockQuantScore) float64) metricRange {
	if len(items) == 0 {
		return metricRange{}
	}
	minVal := selector(items[0])
	maxVal := minVal
	for index := 1; index < len(items); index++ {
		value := selector(items[index])
		if value < minVal {
			minVal = value
		}
		if value > maxVal {
			maxVal = value
		}
	}
	return metricRange{Min: minVal, Max: maxVal}
}

func buildMetricRangeFromValues(values []float64) metricRange {
	if len(values) == 0 {
		return metricRange{}
	}
	minVal := values[0]
	maxVal := values[0]
	for index := 1; index < len(values); index++ {
		if values[index] < minVal {
			minVal = values[index]
		}
		if values[index] > maxVal {
			maxVal = values[index]
		}
	}
	return metricRange{Min: minVal, Max: maxVal}
}

func extractPositiveValues(items []model.StockQuantScore, selector func(model.StockQuantScore) float64) []float64 {
	result := make([]float64, 0, len(items))
	for _, item := range items {
		value := selector(item)
		if value > 0 {
			result = append(result, value)
		}
	}
	return result
}

func normalizeMetric(value float64, valueRange metricRange, invert bool) float64 {
	span := valueRange.Max - valueRange.Min
	if math.Abs(span) < 1e-9 {
		return 0.5
	}
	normalized := (value - valueRange.Min) / span
	if invert {
		normalized = 1 - normalized
	}
	return clampFloat(normalized, 0, 1)
}

func clampFloat(value float64, minValue float64, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func roundTo(value float64, precision int) float64 {
	if precision < 0 {
		precision = 0
	}
	factor := math.Pow(10, float64(precision))
	return math.Round(value*factor) / factor
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func minInt(left int, right int) int {
	if left < right {
		return left
	}
	return right
}

func sqlNullFloat(value sql.NullFloat64) float64 {
	if !value.Valid {
		return 0
	}
	return value.Float64
}

func isTableNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(strings.TrimSpace(err.Error()))
	return strings.Contains(message, "error 1146") ||
		(strings.Contains(message, "doesn't exist") && strings.Contains(message, "table"))
}

func (r *MySQLGrowthRepo) lookupStockNames(symbols []string) (map[string]string, error) {
	result := make(map[string]string)
	symbols = normalizeStockSymbolList(symbols)
	if len(symbols) == 0 {
		return result, nil
	}
	for _, symbol := range symbols {
		result[symbol] = ""
	}
	placeholder := strings.TrimSuffix(strings.Repeat("?,", len(symbols)), ",")
	query := fmt.Sprintf(`
SELECT symbol, name
FROM stock_recommendations
WHERE symbol IN (%s)
ORDER BY created_at DESC`, placeholder)
	args := make([]interface{}, 0, len(symbols))
	for _, symbol := range symbols {
		args = append(args, symbol)
	}
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var symbol string
		var name sql.NullString
		if err := rows.Scan(&symbol, &name); err != nil {
			return nil, err
		}
		symbol = strings.ToUpper(strings.TrimSpace(symbol))
		if _, ok := result[symbol]; !ok {
			continue
		}
		if strings.TrimSpace(result[symbol]) != "" {
			continue
		}
		if name.Valid {
			result[symbol] = strings.TrimSpace(name.String)
		}
	}
	return result, nil
}

func buildQuantReasons(item model.StockQuantScore) []string {
	reasons := make([]string, 0, 6)
	if item.Momentum20 >= 6 {
		reasons = append(reasons, fmt.Sprintf("20日动量%.2f%%，中期趋势较强", item.Momentum20))
	}
	if item.VolumeRatio >= 1.2 {
		reasons = append(reasons, fmt.Sprintf("量比%.2f，成交活跃度提升", item.VolumeRatio))
	}
	if item.NetMFAmount > 0 {
		reasons = append(reasons, fmt.Sprintf("主力净流入%.2f，资金面偏强", item.NetMFAmount))
	} else if item.NetMFAmount < 0 {
		reasons = append(reasons, fmt.Sprintf("主力净流出%.2f，需控制仓位", math.Abs(item.NetMFAmount)))
	}
	if item.PeTTM > 0 {
		if item.PeTTM < 30 {
			reasons = append(reasons, fmt.Sprintf("PE(TTM) %.2f，估值处于可接受区间", item.PeTTM))
		} else {
			reasons = append(reasons, fmt.Sprintf("PE(TTM) %.2f，估值偏高需跟踪兑现", item.PeTTM))
		}
	}
	if item.NewsHeat > 0 {
		reasons = append(reasons, fmt.Sprintf("近14天资讯热度%d，正面占比%.0f%%", item.NewsHeat, item.PositiveNewsRate*100))
	}
	if item.Volatility20 <= 2.5 && item.Drawdown20 <= 7 {
		reasons = append(reasons, "波动与回撤控制在可接受范围")
	}
	if len(reasons) == 0 {
		reasons = append(reasons, fmt.Sprintf("趋势评分%.2f，资金评分%.2f，估值评分%.2f", item.TrendScore, item.FlowScore, item.ValueScore))
	}
	return reasons
}

func buildQuantReasonSummary(item model.StockQuantScore) string {
	reasons := item.Reasons
	if len(reasons) == 0 {
		reasons = buildQuantReasons(item)
	}
	return strings.Join(firstNStrings(reasons, 3), "；")
}

func firstNStrings(items []string, count int) []string {
	if count <= 0 || len(items) == 0 {
		return []string{}
	}
	if len(items) <= count {
		return items
	}
	return items[:count]
}

func buildDailyStockCandidatesFromQuant(items []model.StockQuantScore, validFrom time.Time, validTo time.Time) []dailyStockCandidate {
	if len(items) == 0 {
		return nil
	}
	candidates := make([]dailyStockCandidate, 0, len(items))
	for _, item := range items {
		riskLevel := strings.ToUpper(strings.TrimSpace(item.RiskLevel))
		if riskLevel == "" {
			riskLevel = "MEDIUM"
		}
		positionRange := "8%-12%"
		takeProfit := "上涨10%-15%分批止盈"
		stopLoss := "回撤5%止损"
		switch riskLevel {
		case "LOW":
			positionRange = "10%-15%"
			takeProfit = "上涨8%-12%分批止盈"
			stopLoss = "回撤3%止损"
		case "HIGH":
			positionRange = "5%-8%"
			takeProfit = "上涨12%-18%动态止盈"
			stopLoss = "回撤7%止损"
		}
		score := item.Score
		if score <= 0 {
			score = 80
		}
		candidate := dailyStockCandidate{
			Symbol:        strings.ToUpper(strings.TrimSpace(item.Symbol)),
			Name:          strings.TrimSpace(item.Name),
			Score:         roundTo(score, 2),
			RiskLevel:     riskLevel,
			PositionRange: positionRange,
			ValidFrom:     validFrom,
			ValidTo:       validTo,
			Status:        "PUBLISHED",
			ReasonSummary: fmt.Sprintf("量化评分%.2f，%s", score, buildQuantReasonSummary(item)),
			TechScore:     roundTo(clampFloat(58+item.Momentum20*1.6+item.TrendStrength*2.8, 55, 98), 2),
			FundScore:     roundTo(clampFloat(56+item.Momentum20*1.4+item.Momentum5*1.1, 52, 97), 2),
			SentimentScore: roundTo(
				clampFloat(54+item.VolumeRatio*9+item.Momentum5*1.2-item.Volatility20*0.6, 50, 96),
				2,
			),
			MoneyFlowScore: roundTo(
				clampFloat(55+item.VolumeRatio*11+item.TrendStrength*1.5-item.Drawdown20*0.4, 50, 97),
				2,
			),
			TakeProfit: takeProfit,
			StopLoss:   stopLoss,
			RiskNote:   "量化信号存在失效风险，仅供参考，不构成投资建议",
		}
		if candidate.Name == "" {
			candidate.Name = candidate.Symbol
		}
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Symbol < candidates[j].Symbol
		}
		return candidates[i].Score > candidates[j].Score
	})
	return candidates
}

func buildDefaultDailyStockCandidates(validFrom time.Time, validTo time.Time) []dailyStockCandidate {
	samples := []dailyStockCandidate{
		{Symbol: "600519.SH", Name: "贵州茅台", Score: 91.1, RiskLevel: "MEDIUM", PositionRange: "10%-15%", Status: "PUBLISHED", ReasonSummary: "龙头估值修复"},
		{Symbol: "601318.SH", Name: "中国平安", Score: 88.4, RiskLevel: "MEDIUM", PositionRange: "8%-12%", Status: "PUBLISHED", ReasonSummary: "估值低位"},
		{Symbol: "600036.SH", Name: "招商银行", Score: 86.8, RiskLevel: "LOW", PositionRange: "8%-10%", Status: "PUBLISHED", ReasonSummary: "基本面稳健"},
		{Symbol: "600276.SH", Name: "恒瑞医药", Score: 84.5, RiskLevel: "MEDIUM", PositionRange: "6%-10%", Status: "PUBLISHED", ReasonSummary: "创新药预期"},
		{Symbol: "601012.SH", Name: "隆基绿能", Score: 83.2, RiskLevel: "HIGH", PositionRange: "5%-8%", Status: "PUBLISHED", ReasonSummary: "景气拐点博弈"},
		{Symbol: "000333.SZ", Name: "美的集团", Score: 82.1, RiskLevel: "LOW", PositionRange: "6%-10%", Status: "PUBLISHED", ReasonSummary: "外需改善"},
		{Symbol: "300750.SZ", Name: "宁德时代", Score: 87.5, RiskLevel: "MEDIUM", PositionRange: "8%-12%", Status: "PUBLISHED", ReasonSummary: "产业链回暖"},
		{Symbol: "002594.SZ", Name: "比亚迪", Score: 85.3, RiskLevel: "MEDIUM", PositionRange: "7%-11%", Status: "PUBLISHED", ReasonSummary: "销量韧性"},
		{Symbol: "688981.SH", Name: "中芯国际", Score: 80.8, RiskLevel: "HIGH", PositionRange: "5%-8%", Status: "PUBLISHED", ReasonSummary: "国产替代"},
		{Symbol: "601888.SH", Name: "中国中免", Score: 79.9, RiskLevel: "HIGH", PositionRange: "4%-7%", Status: "PUBLISHED", ReasonSummary: "消费复苏博弈"},
	}
	for index := range samples {
		samples[index].ValidFrom = validFrom
		samples[index].ValidTo = validTo
		samples[index].TechScore = roundTo(clampFloat(samples[index].Score-3, 50, 99), 2)
		samples[index].FundScore = roundTo(clampFloat(samples[index].Score-1, 50, 99), 2)
		samples[index].SentimentScore = roundTo(clampFloat(samples[index].Score-5, 50, 99), 2)
		samples[index].MoneyFlowScore = roundTo(clampFloat(samples[index].Score-2, 50, 99), 2)
		samples[index].TakeProfit = "上涨8%-12%分批止盈"
		samples[index].StopLoss = "跌破关键位止损"
		samples[index].RiskNote = "仅供参考，不构成投资建议"
	}
	return samples
}

func normalizeStockRecommendationLifecycleStatus(status string) string {
	return strings.ToUpper(strings.TrimSpace(status))
}

func allowedStockRecommendationTransitions(status string) []string {
	switch normalizeStockRecommendationLifecycleStatus(status) {
	case "DRAFT":
		return []string{"REVIEW_PENDING", "PUBLISHED", "DISABLED", "INVALIDATED"}
	case "REVIEW_PENDING":
		return []string{"DRAFT", "PUBLISHED", "DISABLED", "INVALIDATED"}
	case "PUBLISHED":
		return []string{"REVIEW_PENDING", "TRACKING", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED", "REVIEWED", "DISABLED"}
	case "TRACKING":
		return []string{"PUBLISHED", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED", "REVIEWED", "DISABLED"}
	case "HIT_TAKE_PROFIT":
		return []string{"TRACKING", "REVIEWED", "DISABLED"}
	case "HIT_STOP_LOSS":
		return []string{"TRACKING", "REVIEWED", "DISABLED"}
	case "INVALIDATED":
		return []string{"DRAFT", "REVIEWED", "DISABLED"}
	case "REVIEWED":
		return []string{"PUBLISHED", "TRACKING", "DISABLED"}
	case "ACTIVE":
		return []string{"PUBLISHED", "REVIEW_PENDING", "TRACKING", "HIT_TAKE_PROFIT", "HIT_STOP_LOSS", "INVALIDATED", "REVIEWED", "DISABLED"}
	case "DISABLED":
		return []string{"DRAFT", "REVIEW_PENDING", "PUBLISHED"}
	default:
		return []string{}
	}
}

func canTransitionStockRecommendationStatus(currentStatus string, targetStatus string) bool {
	currentStatus = normalizeStockRecommendationLifecycleStatus(currentStatus)
	targetStatus = normalizeStockRecommendationLifecycleStatus(targetStatus)
	if currentStatus == "" || targetStatus == "" {
		return false
	}
	if currentStatus == targetStatus {
		return true
	}
	for _, allowed := range allowedStockRecommendationTransitions(currentStatus) {
		if allowed == targetStatus {
			return true
		}
	}
	return false
}

func parseFlexibleDateTime(value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return time.Time{}, errors.New("datetime is required")
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if layout == time.RFC3339 {
			if ts, err := time.Parse(layout, trimmed); err == nil {
				return ts, nil
			}
			continue
		}
		if ts, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return ts, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid datetime format: %s", trimmed)
}
