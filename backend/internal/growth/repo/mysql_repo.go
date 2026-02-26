package repo

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"sercherai/backend/internal/growth/model"
)

type MySQLGrowthRepo struct {
	db    *sql.DB
	redis *redis.Client
}

func NewMySQLGrowthRepo(db *sql.DB, redisClient *redis.Client) *MySQLGrowthRepo {
	return &MySQLGrowthRepo{db: db, redis: redisClient}
}

func (r *MySQLGrowthRepo) ListBrowseHistory(userID string, contentType string, page int, pageSize int) ([]model.BrowseHistory, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{userID}
	filter := ""
	if contentType != "" {
		filter = " AND content_type = ?"
		args = append(args, contentType)
	}

	countQuery := "SELECT COUNT(*) FROM browse_histories WHERE user_id = ?" + filter
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
SELECT id, content_type, content_id, source_page, viewed_at
FROM browse_histories
WHERE user_id = ?` + filter + `
ORDER BY viewed_at DESC
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
		if err := rows.Scan(&item.ID, &item.ContentType, &item.ContentID, &item.SourcePage, &viewedAt); err != nil {
			return nil, 0, err
		}
		item.Title = item.ContentID
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
	url := fmt.Sprintf("https://example.com/invite/%s", code)
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
		id, userID, code, url, channel, status, exp, time.Now())
	if err != nil {
		return model.ShareLink{}, err
	}
	return model.ShareLink{
		ID:         id,
		InviteCode: code,
		URL:        url,
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
SELECT id, invitee_user_id, status, register_at, first_pay_at
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
		if err := rows.Scan(&item.ID, &item.InviteeUser, &item.Status, &registerAt, &firstPayAt); err != nil {
			return nil, 0, err
		}
		if registerAt.Valid {
			item.RegisterAt = registerAt.Time.Format(time.RFC3339)
		}
		if firstPayAt.Valid {
			item.FirstPayAt = firstPayAt.Time.Format(time.RFC3339)
		}
		items = append(items, item)
	}
	return items, total, nil
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
	err := r.db.QueryRow("SELECT id, phone, email, kyc_status, member_level FROM users WHERE id = ?", userID).Scan(
		&profile.ID, &profile.Phone, &email, &profile.KYCStatus, &profile.MemberLevel,
	)
	if err != nil {
		return model.UserProfile{}, err
	}
	if email.Valid {
		profile.Email = email.String
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
	profile.UserID = userID
	err := r.db.QueryRow("SELECT status, kyc_status, member_level FROM users WHERE id = ?", userID).Scan(&profile.Status, &profile.KYCStatus, &profile.MemberLevel)
	if err != nil {
		return model.UserAccessProfile{}, err
	}
	return profile, nil
}

func (r *MySQLGrowthRepo) GetMembershipQuota(userID string) (model.MembershipQuota, error) {
	var memberLevel string
	if err := r.db.QueryRow("SELECT member_level FROM users WHERE id = ?", userID).Scan(&memberLevel); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			memberLevel = "FREE"
		} else {
			return model.MembershipQuota{}, err
		}
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
		PeriodKey:              periodKey,
		DocReadLimit:           limitDoc,
		DocReadUsed:            usedDoc,
		DocReadRemaining:       remainingDoc,
		NewsSubscribeLimit:     limitSub,
		NewsSubscribeUsed:      usedSub,
		NewsSubscribeRemaining: remainingSub,
		ResetCycle:             resetCycle,
		ResetAt:                resetAt.Format(time.RFC3339),
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
	isVIP, err := r.isVIPUser(userID)
	if err != nil {
		return nil, 0, err
	}

	args := []interface{}{}
	filter := " WHERE status = 'PUBLISHED'"
	if !isVIP {
		filter += " AND visibility = 'PUBLIC'"
	}
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
SELECT id, category_id, title, summary, visibility, status, published_at, author_id
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
		var summary, authorID sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Title, &summary, &item.Visibility, &item.Status, &publishedAt, &authorID); err != nil {
			return nil, 0, err
		}
		if summary.Valid {
			item.Summary = summary.String
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
SELECT id, category_id, title, summary, content, visibility, status, published_at, author_id
FROM news_articles
WHERE id = ? AND status = 'PUBLISHED'`
	args := []interface{}{articleID}
	if !isVIP {
		query += " AND visibility = 'PUBLIC'"
	}
	var item model.NewsArticle
	var summary, content, authorID sql.NullString
	var publishedAt sql.NullTime
	err = r.db.QueryRow(query, args...).Scan(
		&item.ID, &item.CategoryID, &item.Title, &summary, &content, &item.Visibility, &item.Status, &publishedAt, &authorID,
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
	if authorID.Valid {
		item.AuthorID = authorID.String
	}
	if publishedAt.Valid {
		item.PublishedAt = publishedAt.Time.Format(time.RFC3339)
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
			if err := tx.QueryRow("SELECT member_level FROM membership_products WHERE id = ?", productID).Scan(&memberLevel); err == nil {
				if memberLevel.Valid && strings.TrimSpace(memberLevel.String) != "" {
					if _, err := tx.Exec("UPDATE users SET member_level = ?, updated_at = ? WHERE id = ?", memberLevel.String, now, userID); err != nil {
						_ = tx.Rollback()
						return err
					}
				}
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
		filter += " AND status = ?"
		args = append(args, status)
	}
	if categoryID != "" {
		filter += " AND category_id = ?"
		args = append(args, categoryID)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM news_articles"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, category_id, title, summary, visibility, status, published_at, author_id
FROM news_articles` + filter + `
ORDER BY created_at DESC
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
		var summary, authorID sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.Title, &summary, &item.Visibility, &item.Status, &publishedAt, &authorID); err != nil {
			return nil, 0, err
		}
		if summary.Valid {
			item.Summary = summary.String
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

func (r *MySQLGrowthRepo) AdminCreateNewsArticle(categoryID string, title string, summary string, content string, visibility string, status string, authorID string) (string, error) {
	id := newID("na")
	_, err := r.db.Exec(`
INSERT INTO news_articles (id, category_id, title, summary, content, visibility, status, published_at, author_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, categoryID, title, summary, content, visibility, status, time.Now(), authorID, time.Now(), time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateNewsArticle(id string, categoryID string, title string, summary string, content string, visibility string, status string) error {
	_, err := r.db.Exec(`
UPDATE news_articles
SET category_id = ?, title = ?, summary = ?, content = ?, visibility = ?, status = ?, updated_at = ?
WHERE id = ?`, categoryID, title, summary, content, visibility, status, time.Now(), id)
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

func (r *MySQLGrowthRepo) ListStockRecommendations(userID string, tradeDate string, page int, pageSize int) ([]model.StockRecommendation, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE status = 'PUBLISHED'"
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
SELECT reco_id, tech_score, fund_score, sentiment_score, money_flow_score, take_profit, stop_loss, risk_note
FROM stock_reco_details
WHERE reco_id = ?`, recoID).Scan(
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
WHERE id = ? AND status = 'PUBLISHED'`, recoID).Scan(&score, &validFrom, &validTo)
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
		filter += " AND status = 'PUBLISHED'"
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
SELECT id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary
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
		var pos, reason sql.NullString
		var vf, vt time.Time
		if err := rows.Scan(&item.ID, &item.Symbol, &item.Name, &item.Score, &item.RiskLevel, &pos, &vf, &vt, &item.Status, &reason); err != nil {
			return nil, 0, err
		}
		if pos.Valid {
			item.PositionRange = pos.String
		}
		if reason.Valid {
			item.ReasonSummary = reason.String
		}
		item.ValidFrom = vf.Format(time.RFC3339)
		item.ValidTo = vt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
}

func (r *MySQLGrowthRepo) AdminCreateStockRecommendation(item model.StockRecommendation) (string, error) {
	id := newID("sr")
	vf, err := time.Parse(time.RFC3339, item.ValidFrom)
	if err != nil {
		return "", err
	}
	vt, err := time.Parse(time.RFC3339, item.ValidTo)
	if err != nil {
		return "", err
	}
	_, err = r.db.Exec(`
INSERT INTO stock_recommendations (id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, item.Symbol, item.Name, item.Score, item.RiskLevel, item.PositionRange, vf, vt, item.Status, item.ReasonSummary, time.Now(),
	)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateStockRecommendationStatus(id string, status string) error {
	_, err := r.db.Exec("UPDATE stock_recommendations SET status = ? WHERE id = ?", status, id)
	return err
}

func (r *MySQLGrowthRepo) AdminGenerateDailyStockRecommendations(tradeDate string) (int, error) {
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	start, err := time.ParseInLocation("2006-01-02", tradeDate, time.Local)
	if err != nil {
		return 0, err
	}
	end := start.Add(24 * time.Hour)
	samples := []model.StockRecommendation{
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

	count := 0
	for _, s := range samples {
		id := newID("sr")
		_, err := r.db.Exec(`
INSERT INTO stock_recommendations (id, symbol, name, score, risk_level, position_range, valid_from, valid_to, status, reason_summary, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			id, s.Symbol, s.Name, s.Score, s.RiskLevel, s.PositionRange, start, end, s.Status, s.ReasonSummary, time.Now(),
		)
		if err != nil {
			return count, err
		}
		_, _ = r.db.Exec(`
INSERT INTO stock_reco_details (reco_id, tech_score, fund_score, sentiment_score, money_flow_score, take_profit, stop_loss, risk_note)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE tech_score=VALUES(tech_score), fund_score=VALUES(fund_score), sentiment_score=VALUES(sentiment_score), money_flow_score=VALUES(money_flow_score), take_profit=VALUES(take_profit), stop_loss=VALUES(stop_loss), risk_note=VALUES(risk_note)`,
			id, s.Score-3, s.Score-1, s.Score-5, s.Score-2, "上涨8%-12%分批止盈", "跌破关键位止损", "仅供参考，不构成投资建议",
		)
		count++
	}
	return count, nil
}

func (r *MySQLGrowthRepo) AdminGenerateDailyFuturesStrategies(tradeDate string) (int, error) {
	if tradeDate == "" {
		tradeDate = time.Now().Format("2006-01-02")
	}
	start, err := time.ParseInLocation("2006-01-02", tradeDate, time.Local)
	if err != nil {
		return 0, err
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
			return count, err
		}
		count++
	}
	return count, nil
}

func (r *MySQLGrowthRepo) AdminListFuturesStrategies(status string, contract string, page int, pageSize int) ([]model.FuturesStrategy, int, error) {
	return r.ListFuturesStrategies("u_admin", contract, status, page, pageSize)
}

func (r *MySQLGrowthRepo) AdminCreateFuturesStrategy(item model.FuturesStrategy) (string, error) {
	id := newID("fs")
	vf, err := time.Parse(time.RFC3339, item.ValidFrom)
	if err != nil {
		return "", err
	}
	vt, err := time.Parse(time.RFC3339, item.ValidTo)
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

func (r *MySQLGrowthRepo) AdminListUsers(status string, kycStatus string, memberLevel string, page int, pageSize int) ([]model.AdminUser, int, error) {
	offset := (page - 1) * pageSize
	args := []interface{}{}
	filter := " WHERE 1=1"
	if status != "" {
		filter += " AND status = ?"
		args = append(args, status)
	}
	if kycStatus != "" {
		filter += " AND kyc_status = ?"
		args = append(args, kycStatus)
	}
	if memberLevel != "" {
		filter += " AND member_level = ?"
		args = append(args, memberLevel)
	}
	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM users"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	query := `
SELECT id, phone, email, status, kyc_status, member_level, created_at
FROM users` + filter + `
ORDER BY created_at DESC
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
		var createdAt time.Time
		if err := rows.Scan(&item.ID, &item.Phone, &email, &item.Status, &item.KYCStatus, &item.MemberLevel, &createdAt); err != nil {
			return nil, 0, err
		}
		if email.Valid {
			item.Email = email.String
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)
		items = append(items, item)
	}
	return items, total, nil
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
	if err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at >= ? AND created_at < ?", startOfDay, endOfDay).Scan(&result.TodayNewUsers); err != nil {
		return result, err
	}
	if err := r.db.QueryRow("SELECT COUNT(*) FROM membership_orders WHERE status = 'PAID' AND paid_at >= ? AND paid_at < ?", startOfDay, endOfDay).Scan(&result.TodayPaidOrders); err != nil {
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
	level := strings.TrimSpace(memberLevel)
	if level == "" {
		level = "VIP1"
	}
	days := durationDays
	if days <= 0 {
		days = 30
	}
	_, err := r.db.Exec("INSERT INTO membership_products (id, name, price, status, member_level, duration_days) VALUES (?, ?, ?, ?, ?, ?)", id, name, price, status, level, days)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (r *MySQLGrowthRepo) AdminUpdateMembershipProductStatus(id string, status string) error {
	_, err := r.db.Exec("UPDATE membership_products SET status = ? WHERE id = ?", status, id)
	return err
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
	_, err := r.db.Exec("UPDATE membership_orders SET status = ?, updated_at = ? WHERE id = ?", status, time.Now(), id)
	return err
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

func (r *MySQLGrowthRepo) AdminUpdateDataSource(sourceKey string, item model.DataSource) error {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return sql.ErrNoRows
	}
	configKey := "data_source." + sourceKey
	payloadBytes, err := json.Marshal(map[string]interface{}{
		"name":        item.Name,
		"source_type": strings.ToUpper(strings.TrimSpace(item.SourceType)),
		"status":      strings.ToUpper(strings.TrimSpace(item.Status)),
		"config":      item.Config,
	})
	if err != nil {
		return err
	}

	result, err := r.db.Exec(`
UPDATE system_configs
SET config_value = ?, description = ?, updated_by = ?, updated_at = ?
WHERE config_key = ?`,
		string(payloadBytes), item.Name, "admin", time.Now(), configKey,
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
	result := model.DataSourceHealthCheck{
		SourceKey: item.SourceKey,
		Status:    "UNKNOWN",
		CheckedAt: time.Now().Format(time.RFC3339),
	}

	if strings.ToUpper(strings.TrimSpace(item.Status)) != "ACTIVE" {
		result.Message = "data source is disabled"
		return result, nil
	}

	endpoint := strings.TrimSpace(fmt.Sprintf("%v", item.Config["endpoint"]))
	if endpoint == "" {
		result.Message = "endpoint not configured"
		return result, nil
	}

	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		result.Message = "invalid endpoint"
		return result, nil
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		result.Message = "unsupported endpoint scheme"
		return result, nil
	}

	client := &http.Client{Timeout: 3 * time.Second}
	start := time.Now()
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		result.Message = err.Error()
		return result, nil
	}
	resp, err := client.Do(req)
	result.LatencyMS = time.Since(start).Milliseconds()
	if err != nil {
		result.Status = "UNHEALTHY"
		result.Reachable = false
		result.Message = err.Error()
		return result, nil
	}
	defer resp.Body.Close()
	result.Reachable = true
	result.HTTPStatus = resp.StatusCode
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		result.Status = "HEALTHY"
		result.Message = "ok"
		return result, nil
	}
	result.Status = "UNHEALTHY"
	result.Message = resp.Status
	return result, nil
}

func (r *MySQLGrowthRepo) getDataSourceBySourceKey(sourceKey string) (model.DataSource, error) {
	sourceKey = strings.TrimSpace(sourceKey)
	if sourceKey == "" {
		return model.DataSource{}, sql.ErrNoRows
	}
	configKey := "data_source." + sourceKey

	var id, configValue string
	var desc sql.NullString
	var updatedAt time.Time
	err := r.db.QueryRow(`
SELECT id, config_value, description, updated_at
FROM system_configs
WHERE config_key = ?
LIMIT 1`, configKey).Scan(&id, &configValue, &desc, &updatedAt)
	if err != nil {
		return model.DataSource{}, err
	}
	item := model.DataSource{
		ID:         id,
		SourceKey:  sourceKey,
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
			item.Name = sourceKey
		}
	}
	return item, nil
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

func (r *MySQLGrowthRepo) AdminCreateSchedulerJobRun(jobName string, triggerSource string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	id := newID("jr")
	now := time.Now()
	var finishedAt interface{} = nil
	if strings.ToUpper(status) != "RUNNING" {
		finishedAt = now
	}
	_, err := r.db.Exec(`
INSERT INTO scheduler_job_runs (id, parent_run_id, retry_count, job_name, trigger_source, status, started_at, finished_at, result_summary, error_message, operator_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, nil, 0, jobName, strings.ToUpper(triggerSource), strings.ToUpper(status), now, finishedAt, resultSummary, errorMessage, operatorID, now,
	)
	if err != nil {
		return "", err
	}
	_ = r.AdminCreateWorkflowMessage(
		"",
		id,
		"SYSTEM",
		"",
		operatorID,
		"JOB_TRIGGERED",
		"任务已触发",
		"任务 "+jobName+" 已触发，状态 "+strings.ToUpper(status),
	)
	return id, nil
}

func (r *MySQLGrowthRepo) AdminRetrySchedulerJobRun(runID string, status string, resultSummary string, errorMessage string, operatorID string) (string, error) {
	var jobName string
	var prevRetry int
	err := r.db.QueryRow("SELECT job_name, retry_count FROM scheduler_job_runs WHERE id = ?", runID).Scan(&jobName, &prevRetry)
	if err != nil {
		return "", err
	}
	id := newID("jr")
	now := time.Now()
	upperStatus := strings.ToUpper(status)
	var finishedAt interface{} = nil
	if upperStatus != "RUNNING" {
		finishedAt = now
	}
	_, err = r.db.Exec(`
INSERT INTO scheduler_job_runs (id, parent_run_id, retry_count, job_name, trigger_source, status, started_at, finished_at, result_summary, error_message, operator_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		id, runID, prevRetry+1, jobName, "MANUAL", upperStatus, now, finishedAt, resultSummary, errorMessage, operatorID, now,
	)
	if err != nil {
		return "", err
	}
	_ = r.AdminCreateWorkflowMessage(
		"",
		id,
		"SYSTEM",
		"",
		operatorID,
		"JOB_RETRIED",
		"任务重跑已触发",
		"任务 "+jobName+" 基于运行 "+runID+" 重跑，状态 "+upperStatus,
	)
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
	var reviewVal interface{} = nil
	if strings.TrimSpace(reviewID) != "" {
		reviewVal = reviewID
	}
	var receiverVal interface{} = nil
	if strings.TrimSpace(receiverID) != "" {
		receiverVal = receiverID
	}
	var senderVal interface{} = nil
	if strings.TrimSpace(senderID) != "" {
		senderVal = senderID
	}
	_, err := r.db.Exec(`
INSERT INTO workflow_messages (id, review_id, target_id, module, receiver_id, sender_id, event_type, title, content, is_read, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?)`,
		newID("wm"), reviewVal, targetID, strings.ToUpper(module), receiverVal, senderVal, strings.ToUpper(eventType), title, content, time.Now(),
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
	args := []interface{}{}
	filter := " WHERE started_at >= ? AND started_at < ?"
	todayStart := time.Now().Truncate(24 * time.Hour)
	tomorrowStart := todayStart.Add(24 * time.Hour)
	args = append(args, todayStart, tomorrowStart)
	if jobName != "" {
		filter += " AND job_name = ?"
		args = append(args, jobName)
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
	return result, nil
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

func (r *MySQLGrowthRepo) isVIPUser(userID string) (bool, error) {
	var memberLevel string
	err := r.db.QueryRow("SELECT member_level FROM users WHERE id = ?", userID).Scan(&memberLevel)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	memberLevel = strings.ToUpper(strings.TrimSpace(memberLevel))
	return strings.HasPrefix(memberLevel, "VIP"), nil
}

func newID(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
}
