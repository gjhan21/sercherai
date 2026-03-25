package repo

import (
	"database/sql"
	"errors"
	"sort"
	"strings"
	"time"

	"sercherai/backend/internal/growth/model"
)

func (r *InMemoryGrowthRepo) seedCommunityData() {
	now := time.Date(2026, 3, 23, 9, 30, 0, 0, time.FixedZone("CST", 8*3600))
	topic := model.CommunityTopicDetail{
		ID:            "ct_demo_001",
		UserID:        "u_demo_001",
		Title:         "白酒龙头先看观察，不追高",
		Summary:       "围绕龙头股的观察观点，先确认量价和消息面是否延续。",
		Content:       "我的看法是先观察，不急着追高，等量价和消息面继续确认后再决定。",
		TopicType:     string(model.CommunityTopicTypeStock),
		Stance:        string(model.CommunityStanceWatch),
		TimeHorizon:   "SHORT",
		ReasonText:    "当前位置情绪较高，但基本面支撑仍在。",
		RiskText:      "若量能无法承接，短线容易回撤。",
		Status:        string(model.CommunityTopicStatusPublished),
		CommentCount:  1,
		LikeCount:     1,
		FavoriteCount: 0,
		ReportCount:   0,
		LastActiveAt:  now.Add(20 * time.Minute).Format(time.RFC3339),
		CreatedAt:     now.Format(time.RFC3339),
		UpdatedAt:     now.Add(20 * time.Minute).Format(time.RFC3339),
		LinkedTarget: model.CommunityTopicLink{
			TargetType:     "STOCK",
			TargetID:       "600519",
			TargetSnapshot: "600519 贵州茅台",
		},
	}
	comment := model.CommunityComment{
		ID:        "cc_demo_001",
		TopicID:   topic.ID,
		UserID:    "u_demo_002",
		Content:   "我更关心这波量能是否还能继续放大。",
		Status:    string(model.CommunityCommentStatusPublished),
		CreatedAt: now.Add(20 * time.Minute).Format(time.RFC3339),
		UpdatedAt: now.Add(20 * time.Minute).Format(time.RFC3339),
	}

	pendingTopic := model.CommunityTopicDetail{
		ID:            "ct_demo_002",
		UserID:        "u_demo_002",
		Title:         "盘后研报怎么看，是继续持有还是等回踩",
		Summary:       "围绕研报解读后的持仓节奏讨论，适合资讯页引流到讨论页联调。",
		Content:       "我更关注研报落地后的兑现节奏，如果只是情绪催化，第二天不一定适合追。",
		TopicType:     string(model.CommunityTopicTypeNews),
		Stance:        string(model.CommunityStanceWatch),
		TimeHorizon:   "SWING",
		ReasonText:    "研报热度高，但价格反应已经较快。",
		RiskText:      "如果开盘直接高开过大，追入性价比会明显下降。",
		Status:        string(model.CommunityTopicStatusPendingReview),
		CommentCount:  1,
		LikeCount:     0,
		FavoriteCount: 0,
		ReportCount:   1,
		LastActiveAt:  now.Add(45 * time.Minute).Format(time.RFC3339),
		CreatedAt:     now.Add(5 * time.Minute).Format(time.RFC3339),
		UpdatedAt:     now.Add(45 * time.Minute).Format(time.RFC3339),
		LinkedTarget: model.CommunityTopicLink{
			TargetType:     "NEWS_ARTICLE",
			TargetID:       "article_demo_002",
			TargetSnapshot: "量化策略周报",
		},
	}

	pendingComment := model.CommunityComment{
		ID:        "cc_demo_002",
		TopicID:   pendingTopic.ID,
		UserID:    "admin_003",
		Content:   "这类话题适合放资讯详情页侧栏引导，不适合做成聊天流。",
		Status:    string(model.CommunityCommentStatusPublished),
		LikeCount: 2,
		CreatedAt: now.Add(45 * time.Minute).Format(time.RFC3339),
		UpdatedAt: now.Add(45 * time.Minute).Format(time.RFC3339),
	}

	report := model.CommunityReport{
		ID:             "cr_demo_001",
		ReporterUserID: "u_demo_004",
		TargetType:     "TOPIC",
		TargetID:       pendingTopic.ID,
		Reason:         "标题偏情绪化，建议管理员复核。",
		Status:         string(model.CommunityReportStatusPending),
		CreatedAt:      now.Add(50 * time.Minute).Format(time.RFC3339),
		UpdatedAt:      now.Add(50 * time.Minute).Format(time.RFC3339),
	}

	r.communityTopics[topic.ID] = topic
	r.communityTopics[pendingTopic.ID] = pendingTopic
	r.communityComments[comment.ID] = comment
	r.communityComments[pendingComment.ID] = pendingComment
	r.communityReports[report.ID] = report
	r.communityReacts[communityReactionKey("u_demo_002", "TOPIC", topic.ID, "LIKE")] = struct{}{}
}

func communityReactionKey(userID string, targetType string, targetID string, reactionType string) string {
	return strings.Join([]string{strings.TrimSpace(userID), strings.ToUpper(strings.TrimSpace(targetType)), strings.TrimSpace(targetID), strings.ToUpper(strings.TrimSpace(reactionType))}, "|")
}

func (r *InMemoryGrowthRepo) ListCommunityTopics(userID string, query model.CommunityTopicListQuery) ([]model.CommunityTopicListItem, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rows := make([]model.CommunityTopicListItem, 0, len(r.communityTopics))
	showMine := strings.EqualFold(query.Mine, "topics") && strings.TrimSpace(userID) != ""
	for _, topic := range r.communityTopics {
		if !showMine && topic.Status != string(model.CommunityTopicStatusPublished) {
			continue
		}
		if query.TopicType != "" && !strings.EqualFold(topic.TopicType, string(query.TopicType)) {
			continue
		}
		if showMine && topic.UserID != userID {
			continue
		}
		rows = append(rows, communityTopicListItemFromDetail(topic, userID, r.communityReacts))
	}

	sort.Slice(rows, func(i, j int) bool {
		if strings.EqualFold(query.Sort, "MOST_ACTIVE") || strings.EqualFold(query.Sort, "active") {
			return rows[i].LastActiveAt > rows[j].LastActiveAt
		}
		return rows[i].CreatedAt > rows[j].CreatedAt
	})

	total := len(rows)
	start, end := paginateBounds(query.Page, query.PageSize, total)
	if start >= total {
		return []model.CommunityTopicListItem{}, total, nil
	}
	return rows[start:end], total, nil
}

func (r *InMemoryGrowthRepo) GetCommunityTopic(userID string, topicID string) (model.CommunityTopicDetail, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	topic, ok := r.communityTopics[topicID]
	if !ok {
		return model.CommunityTopicDetail{}, sql.ErrNoRows
	}
	if topic.Status != string(model.CommunityTopicStatusPublished) && topic.UserID != strings.TrimSpace(userID) {
		return model.CommunityTopicDetail{}, sql.ErrNoRows
	}
	topic.LikedByMe = r.hasCommunityReaction(userID, "TOPIC", topicID, "LIKE")
	topic.FavoritedByMe = r.hasCommunityReaction(userID, "TOPIC", topicID, "FAVORITE")
	return topic, nil
}

func (r *InMemoryGrowthRepo) CreateCommunityTopic(input model.CommunityTopicCreateInput) (model.CommunityTopicDetail, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Format(time.RFC3339)
	id := newID("ct")
	topic := model.CommunityTopicDetail{
		ID:           id,
		UserID:       strings.TrimSpace(input.UserID),
		Title:        strings.TrimSpace(input.Title),
		Summary:      strings.TrimSpace(input.Summary),
		Content:      strings.TrimSpace(input.Content),
		TopicType:    strings.ToUpper(strings.TrimSpace(input.TopicType)),
		Stance:       strings.ToUpper(strings.TrimSpace(input.Stance)),
		TimeHorizon:  strings.ToUpper(strings.TrimSpace(input.TimeHorizon)),
		ReasonText:   strings.TrimSpace(input.ReasonText),
		RiskText:     strings.TrimSpace(input.RiskText),
		Status:       string(model.CommunityTopicStatusPublished),
		LastActiveAt: now,
		CreatedAt:    now,
		UpdatedAt:    now,
		LinkedTarget: model.CommunityTopicLink{
			TargetType:     strings.ToUpper(strings.TrimSpace(input.TargetType)),
			TargetID:       strings.TrimSpace(input.TargetID),
			TargetSnapshot: strings.TrimSpace(input.TargetSnapshot),
		},
	}
	if topic.Summary == "" {
		topic.Summary = summarizeCommunityText(topic.Content)
	}
	r.communityTopics[id] = topic
	return topic, nil
}

func (r *InMemoryGrowthRepo) ListCommunityComments(userID string, topicID string, query model.CommunityCommentListQuery) ([]model.CommunityComment, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rows := make([]model.CommunityComment, 0)
	for _, item := range r.communityComments {
		if item.TopicID != topicID {
			continue
		}
		if strings.TrimSpace(userID) == "" && item.Status != string(model.CommunityCommentStatusPublished) {
			continue
		}
		item.LikedByMe = r.hasCommunityReaction(userID, "COMMENT", item.ID, "LIKE")
		rows = append(rows, item)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].CreatedAt < rows[j].CreatedAt
	})
	total := len(rows)
	start, end := paginateBounds(query.Page, query.PageSize, total)
	if start >= total {
		return []model.CommunityComment{}, total, nil
	}
	return rows[start:end], total, nil
}

func (r *InMemoryGrowthRepo) ListMyCommunityComments(userID string, page int, pageSize int) ([]model.CommunityComment, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rows := make([]model.CommunityComment, 0)
	for _, item := range r.communityComments {
		if item.UserID != strings.TrimSpace(userID) {
			continue
		}
		item.LikedByMe = r.hasCommunityReaction(userID, "COMMENT", item.ID, "LIKE")
		if topic, ok := r.communityTopics[item.TopicID]; ok {
			item.TopicTitle = topic.Title
			item.TopicSummary = topic.Summary
			item.TopicStatus = topic.Status
			item.LinkedTarget = topic.LinkedTarget
		}
		rows = append(rows, item)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].CreatedAt > rows[j].CreatedAt
	})
	total := len(rows)
	start, end := paginateBounds(page, pageSize, total)
	if start >= total {
		return []model.CommunityComment{}, total, nil
	}
	return rows[start:end], total, nil
}

func (r *InMemoryGrowthRepo) CreateCommunityComment(input model.CommunityCommentCreateInput) (model.CommunityComment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	topic, ok := r.communityTopics[input.TopicID]
	if !ok {
		return model.CommunityComment{}, sql.ErrNoRows
	}
	now := time.Now().Format(time.RFC3339)
	comment := model.CommunityComment{
		ID:              newID("cc"),
		TopicID:         strings.TrimSpace(input.TopicID),
		UserID:          strings.TrimSpace(input.UserID),
		ParentCommentID: strings.TrimSpace(input.ParentCommentID),
		ReplyToUserID:   strings.TrimSpace(input.ReplyToUserID),
		Content:         strings.TrimSpace(input.Content),
		Status:          string(model.CommunityCommentStatusPublished),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	r.communityComments[comment.ID] = comment

	topic.CommentCount++
	topic.LastActiveAt = now
	topic.UpdatedAt = now
	r.communityTopics[topic.ID] = topic
	return comment, nil
}

func (r *InMemoryGrowthRepo) CreateCommunityReaction(input model.CommunityReactionInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := communityReactionKey(input.UserID, input.TargetType, input.TargetID, input.ReactionType)
	if _, exists := r.communityReacts[key]; exists {
		return nil
	}
	r.communityReacts[key] = struct{}{}
	r.bumpCommunityReactionCount(strings.ToUpper(strings.TrimSpace(input.TargetType)), strings.TrimSpace(input.TargetID), strings.ToUpper(strings.TrimSpace(input.ReactionType)), 1)
	return nil
}

func (r *InMemoryGrowthRepo) DeleteCommunityReaction(input model.CommunityReactionInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := communityReactionKey(input.UserID, input.TargetType, input.TargetID, input.ReactionType)
	if _, exists := r.communityReacts[key]; !exists {
		return nil
	}
	delete(r.communityReacts, key)
	r.bumpCommunityReactionCount(strings.ToUpper(strings.TrimSpace(input.TargetType)), strings.TrimSpace(input.TargetID), strings.ToUpper(strings.TrimSpace(input.ReactionType)), -1)
	return nil
}

func (r *InMemoryGrowthRepo) CreateCommunityReport(input model.CommunityReportCreateInput) (model.CommunityReport, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Format(time.RFC3339)
	report := model.CommunityReport{
		ID:             newID("cr"),
		ReporterUserID: strings.TrimSpace(input.ReporterUserID),
		TargetType:     strings.ToUpper(strings.TrimSpace(input.TargetType)),
		TargetID:       strings.TrimSpace(input.TargetID),
		Reason:         strings.TrimSpace(input.Reason),
		Status:         string(model.CommunityReportStatusPending),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	r.communityReports[report.ID] = report
	if topic, ok := r.communityTopics[report.TargetID]; ok {
		topic.ReportCount++
		topic.UpdatedAt = now
		r.communityTopics[topic.ID] = topic
	}
	return report, nil
}

func (r *InMemoryGrowthRepo) CreateCommunityNotification(input model.CommunityNotificationInput) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	userID := strings.TrimSpace(input.UserID)
	r.userMessages[userID] = append(r.userMessages[userID], model.UserMessage{
		ID:         newID("msg"),
		Title:      strings.TrimSpace(input.Title),
		Content:    strings.TrimSpace(input.Content),
		Type:       strings.TrimSpace(input.MessageType),
		ReadStatus: "UNREAD",
		CreatedAt:  time.Now().Format(time.RFC3339),
	})
	return nil
}

func (r *InMemoryGrowthRepo) AdminListCommunityTopics(query model.CommunityAdminTopicQuery) ([]model.CommunityTopicListItem, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rows := make([]model.CommunityTopicListItem, 0, len(r.communityTopics))
	for _, topic := range r.communityTopics {
		if query.TopicType != "" && !strings.EqualFold(topic.TopicType, query.TopicType) {
			continue
		}
		if query.Status != "" && !strings.EqualFold(topic.Status, query.Status) {
			continue
		}
		if query.UserID != "" && topic.UserID != query.UserID {
			continue
		}
		rows = append(rows, communityTopicListItemFromDetail(topic, "", r.communityReacts))
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].LastActiveAt > rows[j].LastActiveAt
	})
	total := len(rows)
	start, end := paginateBounds(query.Page, query.PageSize, total)
	if start >= total {
		return []model.CommunityTopicListItem{}, total, nil
	}
	return rows[start:end], total, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateCommunityTopicStatus(id string, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	topic, ok := r.communityTopics[id]
	if !ok {
		return sql.ErrNoRows
	}
	topic.Status = strings.ToUpper(strings.TrimSpace(status))
	topic.UpdatedAt = time.Now().Format(time.RFC3339)
	r.communityTopics[id] = topic
	return nil
}

func (r *InMemoryGrowthRepo) AdminListCommunityComments(query model.CommunityAdminCommentQuery) ([]model.CommunityComment, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rows := make([]model.CommunityComment, 0)
	for _, item := range r.communityComments {
		if query.TopicID != "" && item.TopicID != query.TopicID {
			continue
		}
		if query.UserID != "" && item.UserID != query.UserID {
			continue
		}
		if query.Status != "" && !strings.EqualFold(item.Status, query.Status) {
			continue
		}
		if topic, ok := r.communityTopics[item.TopicID]; ok {
			item.TopicTitle = topic.Title
			item.TopicSummary = topic.Summary
			item.TopicStatus = topic.Status
			item.LinkedTarget = topic.LinkedTarget
		}
		rows = append(rows, item)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].CreatedAt > rows[j].CreatedAt
	})
	total := len(rows)
	start, end := paginateBounds(query.Page, query.PageSize, total)
	if start >= total {
		return []model.CommunityComment{}, total, nil
	}
	return rows[start:end], total, nil
}

func (r *InMemoryGrowthRepo) AdminUpdateCommunityCommentStatus(id string, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.communityComments[id]
	if !ok {
		return sql.ErrNoRows
	}
	item.Status = strings.ToUpper(strings.TrimSpace(status))
	item.UpdatedAt = time.Now().Format(time.RFC3339)
	r.communityComments[id] = item
	return nil
}

func (r *InMemoryGrowthRepo) AdminListCommunityReports(query model.CommunityAdminReportQuery) ([]model.CommunityReport, int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	rows := make([]model.CommunityReport, 0)
	for _, item := range r.communityReports {
		if query.Status != "" && !strings.EqualFold(item.Status, query.Status) {
			continue
		}
		if query.TargetType != "" && !strings.EqualFold(item.TargetType, query.TargetType) {
			continue
		}
		targetType := strings.ToUpper(strings.TrimSpace(item.TargetType))
		if targetType == "TOPIC" {
			if topic, ok := r.communityTopics[item.TargetID]; ok {
				item.TopicID = topic.ID
				item.TopicTitle = topic.Title
				item.TopicSummary = topic.Summary
				item.TargetContent = communityFirstNonEmpty(topic.Summary, topic.Content)
				item.TargetStatus = topic.Status
				item.LinkedTarget = topic.LinkedTarget
			}
		}
		if targetType == "COMMENT" {
			if comment, ok := r.communityComments[item.TargetID]; ok {
				item.TargetContent = comment.Content
				item.TargetStatus = comment.Status
				item.TopicID = comment.TopicID
				if topic, ok := r.communityTopics[comment.TopicID]; ok {
					item.TopicTitle = topic.Title
					item.TopicSummary = topic.Summary
					item.LinkedTarget = topic.LinkedTarget
				}
			}
		}
		rows = append(rows, item)
	}
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].CreatedAt > rows[j].CreatedAt
	})
	total := len(rows)
	start, end := paginateBounds(query.Page, query.PageSize, total)
	if start >= total {
		return []model.CommunityReport{}, total, nil
	}
	return rows[start:end], total, nil
}

func (r *InMemoryGrowthRepo) AdminReviewCommunityReport(id string, status string, reviewNote string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.communityReports[id]
	if !ok {
		return sql.ErrNoRows
	}
	item.Status = strings.ToUpper(strings.TrimSpace(status))
	item.ReviewNote = strings.TrimSpace(reviewNote)
	item.UpdatedAt = time.Now().Format(time.RFC3339)
	r.communityReports[id] = item
	return nil
}

func (r *InMemoryGrowthRepo) hasCommunityReaction(userID string, targetType string, targetID string, reactionType string) bool {
	if strings.TrimSpace(userID) == "" {
		return false
	}
	_, ok := r.communityReacts[communityReactionKey(userID, targetType, targetID, reactionType)]
	return ok
}

func (r *InMemoryGrowthRepo) bumpCommunityReactionCount(targetType string, targetID string, reactionType string, delta int) {
	if targetType == "TOPIC" {
		item, ok := r.communityTopics[targetID]
		if !ok {
			return
		}
		switch reactionType {
		case "LIKE":
			item.LikeCount = communityMaxInt(0, item.LikeCount+delta)
		case "FAVORITE":
			item.FavoriteCount = communityMaxInt(0, item.FavoriteCount+delta)
		}
		item.UpdatedAt = time.Now().Format(time.RFC3339)
		r.communityTopics[targetID] = item
		return
	}
	if targetType == "COMMENT" {
		item, ok := r.communityComments[targetID]
		if !ok {
			return
		}
		if reactionType == "LIKE" {
			item.LikeCount = communityMaxInt(0, item.LikeCount+delta)
		}
		item.UpdatedAt = time.Now().Format(time.RFC3339)
		r.communityComments[targetID] = item
	}
}

func paginateBounds(page int, pageSize int, total int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	return start, end
}

func summarizeCommunityText(content string) string {
	text := strings.TrimSpace(content)
	if len(text) <= 80 {
		return text
	}
	return text[:80]
}

func communityFirstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func communityTopicListItemFromDetail(topic model.CommunityTopicDetail, userID string, reacts map[string]struct{}) model.CommunityTopicListItem {
	_, liked := reacts[communityReactionKey(userID, "TOPIC", topic.ID, "LIKE")]
	_, favorited := reacts[communityReactionKey(userID, "TOPIC", topic.ID, "FAVORITE")]
	return model.CommunityTopicListItem{
		ID:            topic.ID,
		UserID:        topic.UserID,
		Title:         topic.Title,
		Summary:       topic.Summary,
		TopicType:     topic.TopicType,
		Stance:        topic.Stance,
		Status:        topic.Status,
		CommentCount:  topic.CommentCount,
		LikeCount:     topic.LikeCount,
		FavoriteCount: topic.FavoriteCount,
		ReportCount:   topic.ReportCount,
		LastActiveAt:  topic.LastActiveAt,
		CreatedAt:     topic.CreatedAt,
		LikedByMe:     liked,
		FavoritedByMe: favorited,
		LinkedTarget:  topic.LinkedTarget,
	}
}

func communityMaxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func ensureCommunityTargetType(raw string) error {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case "TOPIC", "COMMENT":
		return nil
	default:
		return errors.New("unsupported community target type")
	}
}
