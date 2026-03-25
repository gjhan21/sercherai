package repo

import (
	"database/sql"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *MySQLGrowthRepo) ListCommunityTopics(userID string, query model.CommunityTopicListQuery) ([]model.CommunityTopicListItem, int, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	filters := make([]string, 0, 3)
	args := make([]interface{}, 0, 4)

	if strings.EqualFold(query.Mine, "topics") && strings.TrimSpace(userID) != "" {
		filters = append(filters, "t.user_id = ?")
		args = append(args, strings.TrimSpace(userID))
	} else {
		filters = append(filters, "t.status = ?")
		args = append(args, string(model.CommunityTopicStatusPublished))
	}
	if topicType := strings.ToUpper(strings.TrimSpace(string(query.TopicType))); topicType != "" {
		filters = append(filters, "t.topic_type = ?")
		args = append(args, topicType)
	}

	where := ""
	if len(filters) > 0 {
		where = " WHERE " + strings.Join(filters, " AND ")
	}
	reportWhere := strings.ReplaceAll(where, "status", "rp.status")
	reportWhere = strings.ReplaceAll(reportWhere, "target_type", "rp.target_type")

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM discussion_topics t"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := " ORDER BY t.created_at DESC"
	if strings.EqualFold(query.Sort, "MOST_ACTIVE") || strings.EqualFold(query.Sort, "active") {
		orderBy = " ORDER BY t.last_active_at DESC, t.created_at DESC"
	}

	queryArgs := make([]interface{}, 0, len(args)+4)
	queryArgs = append(queryArgs, strings.TrimSpace(userID), strings.TrimSpace(userID))
	queryArgs = append(queryArgs, args...)
	queryArgs = append(queryArgs, pageSize, offset)

	rows, err := r.db.Query(`
SELECT
	t.id,
	t.user_id,
	t.title,
	t.summary,
	t.topic_type,
	t.stance,
	t.status,
	t.comment_count,
	t.like_count,
	t.favorite_count,
	t.report_count,
	t.last_active_at,
	t.created_at,
	COALESCE(l.target_type, ''),
	COALESCE(l.target_id, ''),
	COALESCE(l.target_snapshot, ''),
	EXISTS(
		SELECT 1 FROM discussion_reactions dr
		WHERE dr.user_id = ? AND dr.target_type = 'TOPIC' AND dr.target_id = t.id AND dr.reaction_type = 'LIKE'
	) AS liked_by_me,
	EXISTS(
		SELECT 1 FROM discussion_reactions dr
		WHERE dr.user_id = ? AND dr.target_type = 'TOPIC' AND dr.target_id = t.id AND dr.reaction_type = 'FAVORITE'
	) AS favorited_by_me
FROM discussion_topics t
LEFT JOIN discussion_topic_links l ON l.topic_id = t.id`+where+orderBy+`
LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.CommunityTopicListItem, 0)
	for rows.Next() {
		item, err := scanCommunityTopicListItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) GetCommunityTopic(userID string, topicID string) (model.CommunityTopicDetail, error) {
	row := r.db.QueryRow(`
SELECT
	t.id,
	t.user_id,
	t.title,
	t.summary,
	t.content,
	t.topic_type,
	t.stance,
	t.time_horizon,
	t.reason_text,
	t.risk_text,
	t.status,
	t.comment_count,
	t.like_count,
	t.favorite_count,
	t.report_count,
	t.last_active_at,
	t.created_at,
	t.updated_at,
	COALESCE(l.target_type, ''),
	COALESCE(l.target_id, ''),
	COALESCE(l.target_snapshot, ''),
	EXISTS(
		SELECT 1 FROM discussion_reactions dr
		WHERE dr.user_id = ? AND dr.target_type = 'TOPIC' AND dr.target_id = t.id AND dr.reaction_type = 'LIKE'
	) AS liked_by_me,
	EXISTS(
		SELECT 1 FROM discussion_reactions dr
		WHERE dr.user_id = ? AND dr.target_type = 'TOPIC' AND dr.target_id = t.id AND dr.reaction_type = 'FAVORITE'
	) AS favorited_by_me
FROM discussion_topics t
LEFT JOIN discussion_topic_links l ON l.topic_id = t.id
WHERE t.id = ?
LIMIT 1`, strings.TrimSpace(userID), strings.TrimSpace(userID), strings.TrimSpace(topicID))

	item, err := scanCommunityTopicDetail(row)
	if err != nil {
		return model.CommunityTopicDetail{}, err
	}
	if !communityTopicVisibleToUser(item.Status, item.UserID, userID) {
		return model.CommunityTopicDetail{}, sql.ErrNoRows
	}
	return item, nil
}

func (r *MySQLGrowthRepo) CreateCommunityTopic(input model.CommunityTopicCreateInput) (model.CommunityTopicDetail, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.CommunityTopicDetail{}, err
	}
	defer tx.Rollback()

	now := time.Now()
	id := newID("ct")
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		summary = summarizeCommunityText(input.Content)
	}

	if _, err := tx.Exec(`
INSERT INTO discussion_topics (
	id, user_id, title, summary, content, topic_type, stance, time_horizon, reason_text, risk_text,
	status, comment_count, like_count, favorite_count, report_count, last_active_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, ?, ?, ?)`,
		id,
		strings.TrimSpace(input.UserID),
		strings.TrimSpace(input.Title),
		summary,
		strings.TrimSpace(input.Content),
		strings.ToUpper(strings.TrimSpace(input.TopicType)),
		strings.ToUpper(strings.TrimSpace(input.Stance)),
		strings.ToUpper(strings.TrimSpace(input.TimeHorizon)),
		strings.TrimSpace(input.ReasonText),
		strings.TrimSpace(input.RiskText),
		string(model.CommunityTopicStatusPublished),
		now,
		now,
		now,
	); err != nil {
		return model.CommunityTopicDetail{}, err
	}

	if _, err := tx.Exec(`
INSERT INTO discussion_topic_links (topic_id, target_type, target_id, target_snapshot, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		id,
		strings.ToUpper(strings.TrimSpace(input.TargetType)),
		strings.TrimSpace(input.TargetID),
		strings.TrimSpace(input.TargetSnapshot),
		now,
		now,
	); err != nil {
		return model.CommunityTopicDetail{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.CommunityTopicDetail{}, err
	}
	return r.GetCommunityTopic(strings.TrimSpace(input.UserID), id)
}

func (r *MySQLGrowthRepo) ListCommunityComments(userID string, topicID string, query model.CommunityCommentListQuery) ([]model.CommunityComment, int, error) {
	if _, err := r.GetCommunityTopic(userID, topicID); err != nil {
		return nil, 0, err
	}

	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	args := []interface{}{strings.TrimSpace(topicID)}
	filter := " WHERE c.topic_id = ? AND c.status = ?"
	args = append(args, string(model.CommunityCommentStatusPublished))

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM discussion_comments c"+filter, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT
	c.id,
	c.topic_id,
	c.user_id,
	COALESCE(c.parent_comment_id, ''),
	COALESCE(c.reply_to_user_id, ''),
	c.content,
	c.status,
	c.like_count,
	c.created_at,
	c.updated_at,
	EXISTS(
		SELECT 1 FROM discussion_reactions dr
		WHERE dr.user_id = ? AND dr.target_type = 'COMMENT' AND dr.target_id = c.id AND dr.reaction_type = 'LIKE'
	) AS liked_by_me
FROM discussion_comments c`+filter+`
ORDER BY
	CASE WHEN c.parent_comment_id IS NULL OR c.parent_comment_id = '' THEN 0 ELSE 1 END ASC,
	c.created_at ASC
LIMIT ? OFFSET ?`, append([]interface{}{strings.TrimSpace(userID)}, append(args, pageSize, offset)...)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.CommunityComment, 0)
	for rows.Next() {
		item, err := scanCommunityComment(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) ListMyCommunityComments(userID string, page int, pageSize int) ([]model.CommunityComment, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	trimmedUserID := strings.TrimSpace(userID)

	var total int
	if err := r.db.QueryRow(
		"SELECT COUNT(*) FROM discussion_comments c WHERE c.user_id = ?",
		trimmedUserID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT
	c.id,
	c.topic_id,
	COALESCE(t.title, ''),
	COALESCE(t.summary, ''),
	COALESCE(t.status, ''),
	c.user_id,
	COALESCE(c.parent_comment_id, ''),
	COALESCE(c.reply_to_user_id, ''),
	c.content,
	c.status,
	c.like_count,
	c.created_at,
	c.updated_at,
	EXISTS(
		SELECT 1 FROM discussion_reactions dr
		WHERE dr.user_id = ? AND dr.target_type = 'COMMENT' AND dr.target_id = c.id AND dr.reaction_type = 'LIKE'
	) AS liked_by_me,
	COALESCE(l.target_type, ''),
	COALESCE(l.target_id, ''),
	COALESCE(l.target_snapshot, '')
FROM discussion_comments c
LEFT JOIN discussion_topics t ON t.id = c.topic_id
LEFT JOIN discussion_topic_links l ON l.topic_id = c.topic_id
WHERE c.user_id = ?
ORDER BY c.created_at DESC
LIMIT ? OFFSET ?`, trimmedUserID, trimmedUserID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.CommunityComment, 0)
	for rows.Next() {
		item, err := scanCommunityCommentWithTopicContext(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) CreateCommunityComment(input model.CommunityCommentCreateInput) (model.CommunityComment, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.CommunityComment{}, err
	}
	defer tx.Rollback()

	var topicExists int
	if err := tx.QueryRow("SELECT COUNT(*) FROM discussion_topics WHERE id = ?", strings.TrimSpace(input.TopicID)).Scan(&topicExists); err != nil {
		return model.CommunityComment{}, err
	}
	if topicExists == 0 {
		return model.CommunityComment{}, sql.ErrNoRows
	}

	now := time.Now()
	id := newID("cc")
	if _, err := tx.Exec(`
INSERT INTO discussion_comments (
	id, topic_id, user_id, parent_comment_id, reply_to_user_id, content, status, like_count, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, 0, ?, ?)`,
		id,
		strings.TrimSpace(input.TopicID),
		strings.TrimSpace(input.UserID),
		nullIfBlank(input.ParentCommentID),
		nullIfBlank(input.ReplyToUserID),
		strings.TrimSpace(input.Content),
		string(model.CommunityCommentStatusPublished),
		now,
		now,
	); err != nil {
		return model.CommunityComment{}, err
	}

	if _, err := tx.Exec(`
UPDATE discussion_topics
SET comment_count = comment_count + 1, last_active_at = ?, updated_at = ?
WHERE id = ?`, now, now, strings.TrimSpace(input.TopicID)); err != nil {
		return model.CommunityComment{}, err
	}

	if err := tx.Commit(); err != nil {
		return model.CommunityComment{}, err
	}
	return r.getCommunityCommentByID(strings.TrimSpace(input.UserID), id)
}

func (r *MySQLGrowthRepo) CreateCommunityReaction(input model.CommunityReactionInput) error {
	if err := ensureCommunityTargetType(input.TargetType); err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
INSERT IGNORE INTO discussion_reactions (id, user_id, target_type, target_id, reaction_type, created_at)
VALUES (?, ?, ?, ?, ?, ?)`,
		newID("dcr"),
		strings.TrimSpace(input.UserID),
		strings.ToUpper(strings.TrimSpace(input.TargetType)),
		strings.TrimSpace(input.TargetID),
		strings.ToUpper(strings.TrimSpace(input.ReactionType)),
		time.Now(),
	)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected > 0 {
		if err := r.bumpCommunityReactionCount(tx, input.TargetType, input.TargetID, input.ReactionType, 1); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MySQLGrowthRepo) DeleteCommunityReaction(input model.CommunityReactionInput) error {
	if err := ensureCommunityTargetType(input.TargetType); err != nil {
		return err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
DELETE FROM discussion_reactions
WHERE user_id = ? AND target_type = ? AND target_id = ? AND reaction_type = ?`,
		strings.TrimSpace(input.UserID),
		strings.ToUpper(strings.TrimSpace(input.TargetType)),
		strings.TrimSpace(input.TargetID),
		strings.ToUpper(strings.TrimSpace(input.ReactionType)),
	)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected > 0 {
		if err := r.bumpCommunityReactionCount(tx, input.TargetType, input.TargetID, input.ReactionType, -1); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *MySQLGrowthRepo) CreateCommunityReport(input model.CommunityReportCreateInput) (model.CommunityReport, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return model.CommunityReport{}, err
	}
	defer tx.Rollback()

	now := time.Now()
	item := model.CommunityReport{
		ID:             newID("cr"),
		ReporterUserID: strings.TrimSpace(input.ReporterUserID),
		TargetType:     strings.ToUpper(strings.TrimSpace(input.TargetType)),
		TargetID:       strings.TrimSpace(input.TargetID),
		Reason:         strings.TrimSpace(input.Reason),
		Status:         string(model.CommunityReportStatusPending),
		CreatedAt:      now.Format(time.RFC3339),
		UpdatedAt:      now.Format(time.RFC3339),
	}

	if _, err := tx.Exec(`
INSERT INTO discussion_reports (id, reporter_user_id, target_type, target_id, reason, status, review_note, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, NULL, ?, ?)`,
		item.ID,
		item.ReporterUserID,
		item.TargetType,
		item.TargetID,
		item.Reason,
		item.Status,
		now,
		now,
	); err != nil {
		return model.CommunityReport{}, err
	}

	if item.TargetType == "TOPIC" {
		if _, err := tx.Exec("UPDATE discussion_topics SET report_count = report_count + 1, updated_at = ? WHERE id = ?", now, item.TargetID); err != nil {
			return model.CommunityReport{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return model.CommunityReport{}, err
	}
	return item, nil
}

func (r *MySQLGrowthRepo) CreateCommunityNotification(input model.CommunityNotificationInput) error {
	_, err := r.db.Exec(`
INSERT INTO messages (id, user_id, title, content, type, read_status, created_at)
VALUES (?, ?, ?, ?, ?, 'UNREAD', ?)`,
		newID("msg"),
		strings.TrimSpace(input.UserID),
		strings.TrimSpace(input.Title),
		strings.TrimSpace(input.Content),
		strings.TrimSpace(input.MessageType),
		time.Now(),
	)
	return err
}

func (r *MySQLGrowthRepo) AdminListCommunityTopics(query model.CommunityAdminTopicQuery) ([]model.CommunityTopicListItem, int, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	filters := make([]string, 0, 3)
	args := make([]interface{}, 0, 4)
	if topicType := strings.ToUpper(strings.TrimSpace(query.TopicType)); topicType != "" {
		filters = append(filters, "t.topic_type = ?")
		args = append(args, topicType)
	}
	if status := strings.ToUpper(strings.TrimSpace(query.Status)); status != "" {
		filters = append(filters, "t.status = ?")
		args = append(args, status)
	}
	if userID := strings.TrimSpace(query.UserID); userID != "" {
		filters = append(filters, "t.user_id = ?")
		args = append(args, userID)
	}

	where := ""
	if len(filters) > 0 {
		where = " WHERE " + strings.Join(filters, " AND ")
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM discussion_topics t"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT
	t.id,
	t.user_id,
	t.title,
	t.summary,
	t.topic_type,
	t.stance,
	t.status,
	t.comment_count,
	t.like_count,
	t.favorite_count,
	t.report_count,
	t.last_active_at,
	t.created_at,
	COALESCE(l.target_type, ''),
	COALESCE(l.target_id, ''),
	COALESCE(l.target_snapshot, ''),
	0 AS liked_by_me,
	0 AS favorited_by_me
FROM discussion_topics t
LEFT JOIN discussion_topic_links l ON l.topic_id = t.id`+where+`
ORDER BY t.last_active_at DESC, t.created_at DESC
LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.CommunityTopicListItem, 0)
	for rows.Next() {
		item, err := scanCommunityTopicListItem(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminUpdateCommunityTopicStatus(id string, status string) error {
	res, err := r.db.Exec("UPDATE discussion_topics SET status = ?, updated_at = ? WHERE id = ?", strings.ToUpper(strings.TrimSpace(status)), time.Now(), strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminListCommunityComments(query model.CommunityAdminCommentQuery) ([]model.CommunityComment, int, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	filters := make([]string, 0, 3)
	args := make([]interface{}, 0, 4)
	if topicID := strings.TrimSpace(query.TopicID); topicID != "" {
		filters = append(filters, "c.topic_id = ?")
		args = append(args, topicID)
	}
	if status := strings.ToUpper(strings.TrimSpace(query.Status)); status != "" {
		filters = append(filters, "c.status = ?")
		args = append(args, status)
	}
	if userID := strings.TrimSpace(query.UserID); userID != "" {
		filters = append(filters, "c.user_id = ?")
		args = append(args, userID)
	}

	where := ""
	if len(filters) > 0 {
		where = " WHERE " + strings.Join(filters, " AND ")
	}

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM discussion_comments c"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT
	c.id,
	c.topic_id,
	COALESCE(t.title, ''),
	COALESCE(t.summary, ''),
	COALESCE(t.status, ''),
	c.user_id,
	COALESCE(c.parent_comment_id, ''),
	COALESCE(c.reply_to_user_id, ''),
	c.content,
	c.status,
	c.like_count,
	c.created_at,
	c.updated_at,
	0 AS liked_by_me,
	COALESCE(l.target_type, ''),
	COALESCE(l.target_id, ''),
	COALESCE(l.target_snapshot, '')
FROM discussion_comments c
LEFT JOIN discussion_topics t ON t.id = c.topic_id
LEFT JOIN discussion_topic_links l ON l.topic_id = c.topic_id`+where+`
ORDER BY c.created_at DESC
LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.CommunityComment, 0)
	for rows.Next() {
		item, err := scanCommunityCommentWithTopicContext(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminUpdateCommunityCommentStatus(id string, status string) error {
	res, err := r.db.Exec("UPDATE discussion_comments SET status = ?, updated_at = ? WHERE id = ?", strings.ToUpper(strings.TrimSpace(status)), time.Now(), strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) AdminListCommunityReports(query model.CommunityAdminReportQuery) ([]model.CommunityReport, int, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	filters := make([]string, 0, 2)
	args := make([]interface{}, 0, 3)
	if status := strings.ToUpper(strings.TrimSpace(query.Status)); status != "" {
		filters = append(filters, "status = ?")
		args = append(args, status)
	}
	if targetType := strings.ToUpper(strings.TrimSpace(query.TargetType)); targetType != "" {
		filters = append(filters, "target_type = ?")
		args = append(args, targetType)
	}

	where := ""
	if len(filters) > 0 {
		where = " WHERE " + strings.Join(filters, " AND ")
	}
	reportWhere := strings.ReplaceAll(where, "status", "rp.status")
	reportWhere = strings.ReplaceAll(reportWhere, "target_type", "rp.target_type")

	var total int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM discussion_reports"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.Query(`
SELECT
	rp.id,
	rp.reporter_user_id,
	rp.target_type,
	rp.target_id,
	COALESCE(
		CASE
			WHEN rp.target_type = 'COMMENT' THEN dc.topic_id
			WHEN rp.target_type = 'TOPIC' THEN dt.id
			ELSE ''
		END,
		''
	) AS topic_id,
	COALESCE(
		CASE
			WHEN rp.target_type = 'COMMENT' THEN dct.title
			WHEN rp.target_type = 'TOPIC' THEN dt.title
			ELSE ''
		END,
		''
	) AS topic_title,
	COALESCE(
		CASE
			WHEN rp.target_type = 'COMMENT' THEN dct.summary
			WHEN rp.target_type = 'TOPIC' THEN dt.summary
			ELSE ''
		END,
		''
	) AS topic_summary,
	COALESCE(
		CASE
			WHEN rp.target_type = 'COMMENT' THEN dc.content
			WHEN rp.target_type = 'TOPIC' THEN NULLIF(TRIM(dt.summary), '')
			ELSE ''
		END,
		CASE
			WHEN rp.target_type = 'TOPIC' THEN COALESCE(dt.content, '')
			ELSE ''
		END
	) AS target_content,
	COALESCE(
		CASE
			WHEN rp.target_type = 'COMMENT' THEN dc.status
			WHEN rp.target_type = 'TOPIC' THEN dt.status
			ELSE ''
		END,
		''
	) AS target_status,
	COALESCE(
		CASE
			WHEN rp.target_type = 'COMMENT' THEN lct.target_type
			WHEN rp.target_type = 'TOPIC' THEN lt.target_type
			ELSE ''
		END,
		''
	) AS linked_target_type,
	COALESCE(
		CASE
			WHEN rp.target_type = 'COMMENT' THEN lct.target_id
			WHEN rp.target_type = 'TOPIC' THEN lt.target_id
			ELSE ''
		END,
		''
	) AS linked_target_id,
	COALESCE(
		CASE
			WHEN rp.target_type = 'COMMENT' THEN lct.target_snapshot
			WHEN rp.target_type = 'TOPIC' THEN lt.target_snapshot
			ELSE ''
		END,
		''
	) AS linked_target_snapshot,
	rp.reason,
	rp.status,
	COALESCE(rp.review_note, ''),
	rp.created_at,
	rp.updated_at
FROM discussion_reports rp
LEFT JOIN discussion_topics dt ON rp.target_type = 'TOPIC' AND dt.id = rp.target_id
LEFT JOIN discussion_topic_links lt ON lt.topic_id = dt.id
LEFT JOIN discussion_comments dc ON rp.target_type = 'COMMENT' AND dc.id = rp.target_id
LEFT JOIN discussion_topics dct ON dct.id = dc.topic_id
LEFT JOIN discussion_topic_links lct ON lct.topic_id = dct.id`+reportWhere+`
ORDER BY rp.created_at DESC
LIMIT ? OFFSET ?`, append(args, pageSize, offset)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.CommunityReport, 0)
	for rows.Next() {
		item, err := scanCommunityReportWithContext(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *MySQLGrowthRepo) AdminReviewCommunityReport(id string, status string, reviewNote string) error {
	res, err := r.db.Exec(`
UPDATE discussion_reports
SET status = ?, review_note = ?, updated_at = ?
WHERE id = ?`,
		strings.ToUpper(strings.TrimSpace(status)),
		strings.TrimSpace(reviewNote),
		time.Now(),
		strings.TrimSpace(id),
	)
	if err != nil {
		return err
	}
	if affected, _ := res.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *MySQLGrowthRepo) getCommunityCommentByID(userID string, commentID string) (model.CommunityComment, error) {
	row := r.db.QueryRow(`
SELECT
	c.id,
	c.topic_id,
	c.user_id,
	COALESCE(c.parent_comment_id, ''),
	COALESCE(c.reply_to_user_id, ''),
	c.content,
	c.status,
	c.like_count,
	c.created_at,
	c.updated_at,
	EXISTS(
		SELECT 1 FROM discussion_reactions dr
		WHERE dr.user_id = ? AND dr.target_type = 'COMMENT' AND dr.target_id = c.id AND dr.reaction_type = 'LIKE'
	) AS liked_by_me
FROM discussion_comments c
WHERE c.id = ?
LIMIT 1`, strings.TrimSpace(userID), strings.TrimSpace(commentID))
	return scanCommunityComment(row)
}

func (r *MySQLGrowthRepo) bumpCommunityReactionCount(tx *sql.Tx, targetType string, targetID string, reactionType string, delta int) error {
	targetKind := strings.ToUpper(strings.TrimSpace(targetType))
	reactionKind := strings.ToUpper(strings.TrimSpace(reactionType))

	switch targetKind {
	case "TOPIC":
		column := ""
		switch reactionKind {
		case "LIKE":
			column = "like_count"
		case "FAVORITE":
			column = "favorite_count"
		default:
			return nil
		}
		_, err := tx.Exec(`
UPDATE discussion_topics
SET `+column+` = GREATEST(0, `+column+` + ?), updated_at = ?
WHERE id = ?`, delta, time.Now(), strings.TrimSpace(targetID))
		return err
	case "COMMENT":
		if reactionKind != "LIKE" {
			return nil
		}
		_, err := tx.Exec(`
UPDATE discussion_comments
SET like_count = GREATEST(0, like_count + ?), updated_at = ?
WHERE id = ?`, delta, time.Now(), strings.TrimSpace(targetID))
		return err
	default:
		return nil
	}
}

func scanCommunityTopicListItem(scanner interface {
	Scan(dest ...interface{}) error
}) (model.CommunityTopicListItem, error) {
	var item model.CommunityTopicListItem
	var lastActiveAt time.Time
	var createdAt time.Time
	var likedByMe int
	var favoritedByMe int
	if err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.Title,
		&item.Summary,
		&item.TopicType,
		&item.Stance,
		&item.Status,
		&item.CommentCount,
		&item.LikeCount,
		&item.FavoriteCount,
		&item.ReportCount,
		&lastActiveAt,
		&createdAt,
		&item.LinkedTarget.TargetType,
		&item.LinkedTarget.TargetID,
		&item.LinkedTarget.TargetSnapshot,
		&likedByMe,
		&favoritedByMe,
	); err != nil {
		return model.CommunityTopicListItem{}, err
	}
	item.LastActiveAt = lastActiveAt.Format(time.RFC3339)
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.LikedByMe = likedByMe > 0
	item.FavoritedByMe = favoritedByMe > 0
	return item, nil
}

func scanCommunityTopicDetail(scanner interface {
	Scan(dest ...interface{}) error
}) (model.CommunityTopicDetail, error) {
	var item model.CommunityTopicDetail
	var lastActiveAt time.Time
	var createdAt time.Time
	var updatedAt time.Time
	var likedByMe int
	var favoritedByMe int
	if err := scanner.Scan(
		&item.ID,
		&item.UserID,
		&item.Title,
		&item.Summary,
		&item.Content,
		&item.TopicType,
		&item.Stance,
		&item.TimeHorizon,
		&item.ReasonText,
		&item.RiskText,
		&item.Status,
		&item.CommentCount,
		&item.LikeCount,
		&item.FavoriteCount,
		&item.ReportCount,
		&lastActiveAt,
		&createdAt,
		&updatedAt,
		&item.LinkedTarget.TargetType,
		&item.LinkedTarget.TargetID,
		&item.LinkedTarget.TargetSnapshot,
		&likedByMe,
		&favoritedByMe,
	); err != nil {
		return model.CommunityTopicDetail{}, err
	}
	item.LastActiveAt = lastActiveAt.Format(time.RFC3339)
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	item.LikedByMe = likedByMe > 0
	item.FavoritedByMe = favoritedByMe > 0
	return item, nil
}

func scanCommunityComment(scanner interface {
	Scan(dest ...interface{}) error
}) (model.CommunityComment, error) {
	var item model.CommunityComment
	var createdAt time.Time
	var updatedAt time.Time
	var likedByMe int
	if err := scanner.Scan(
		&item.ID,
		&item.TopicID,
		&item.UserID,
		&item.ParentCommentID,
		&item.ReplyToUserID,
		&item.Content,
		&item.Status,
		&item.LikeCount,
		&createdAt,
		&updatedAt,
		&likedByMe,
	); err != nil {
		return model.CommunityComment{}, err
	}
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	item.LikedByMe = likedByMe > 0
	return item, nil
}

func scanCommunityCommentWithTopicContext(scanner interface {
	Scan(dest ...interface{}) error
}) (model.CommunityComment, error) {
	var item model.CommunityComment
	var createdAt time.Time
	var updatedAt time.Time
	var likedByMe int
	if err := scanner.Scan(
		&item.ID,
		&item.TopicID,
		&item.TopicTitle,
		&item.TopicSummary,
		&item.TopicStatus,
		&item.UserID,
		&item.ParentCommentID,
		&item.ReplyToUserID,
		&item.Content,
		&item.Status,
		&item.LikeCount,
		&createdAt,
		&updatedAt,
		&likedByMe,
		&item.LinkedTarget.TargetType,
		&item.LinkedTarget.TargetID,
		&item.LinkedTarget.TargetSnapshot,
	); err != nil {
		return model.CommunityComment{}, err
	}
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	item.LikedByMe = likedByMe > 0
	return item, nil
}

func scanCommunityReportWithContext(scanner interface {
	Scan(dest ...interface{}) error
}) (model.CommunityReport, error) {
	var item model.CommunityReport
	var createdAt time.Time
	var updatedAt time.Time
	if err := scanner.Scan(
		&item.ID,
		&item.ReporterUserID,
		&item.TargetType,
		&item.TargetID,
		&item.TopicID,
		&item.TopicTitle,
		&item.TopicSummary,
		&item.TargetContent,
		&item.TargetStatus,
		&item.LinkedTarget.TargetType,
		&item.LinkedTarget.TargetID,
		&item.LinkedTarget.TargetSnapshot,
		&item.Reason,
		&item.Status,
		&item.ReviewNote,
		&createdAt,
		&updatedAt,
	); err != nil {
		return model.CommunityReport{}, err
	}
	item.CreatedAt = createdAt.Format(time.RFC3339)
	item.UpdatedAt = updatedAt.Format(time.RFC3339)
	return item, nil
}

func communityTopicVisibleToUser(status string, ownerUserID string, userID string) bool {
	if strings.EqualFold(strings.TrimSpace(status), string(model.CommunityTopicStatusPublished)) {
		return true
	}
	return strings.TrimSpace(ownerUserID) != "" && strings.TrimSpace(ownerUserID) == strings.TrimSpace(userID)
}

func nullIfBlank(value string) interface{} {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return trimmed
}
