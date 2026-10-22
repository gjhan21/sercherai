package repo

import (
	"database/sql"
	"time"

	"sercherai/backend/internal/growth/model"
)

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
