package model

type CommunityTopicType string

const (
	CommunityTopicTypeStock    CommunityTopicType = "STOCK"
	CommunityTopicTypeFutures  CommunityTopicType = "FUTURES"
	CommunityTopicTypeNews     CommunityTopicType = "NEWS"
	CommunityTopicTypeStrategy CommunityTopicType = "STRATEGY"
)

type CommunityTopicStatus string

const (
	CommunityTopicStatusPublished     CommunityTopicStatus = "PUBLISHED"
	CommunityTopicStatusPendingReview CommunityTopicStatus = "PENDING_REVIEW"
	CommunityTopicStatusHidden        CommunityTopicStatus = "HIDDEN"
	CommunityTopicStatusDeleted       CommunityTopicStatus = "DELETED"
)

type CommunityCommentStatus string

const (
	CommunityCommentStatusPublished     CommunityCommentStatus = "PUBLISHED"
	CommunityCommentStatusPendingReview CommunityCommentStatus = "PENDING_REVIEW"
	CommunityCommentStatusHidden        CommunityCommentStatus = "HIDDEN"
	CommunityCommentStatusDeleted       CommunityCommentStatus = "DELETED"
)

type CommunityReportStatus string

const (
	CommunityReportStatusPending  CommunityReportStatus = "PENDING"
	CommunityReportStatusResolved CommunityReportStatus = "RESOLVED"
	CommunityReportStatusRejected CommunityReportStatus = "REJECTED"
)

type CommunityStance string

const (
	CommunityStanceBullish CommunityStance = "BULLISH"
	CommunityStanceBearish CommunityStance = "BEARISH"
	CommunityStanceWatch   CommunityStance = "WATCH"
)

type CommunityReactionType string

const (
	CommunityReactionTypeLike     CommunityReactionType = "LIKE"
	CommunityReactionTypeFavorite CommunityReactionType = "FAVORITE"
)

type CommunityReactionTargetType string

const (
	CommunityReactionTargetTopic   CommunityReactionTargetType = "TOPIC"
	CommunityReactionTargetComment CommunityReactionTargetType = "COMMENT"
)

type CommunityTopicLink struct {
	TargetType     string `json:"target_type"`
	TargetID       string `json:"target_id"`
	TargetSnapshot string `json:"target_snapshot"`
}

type CommunityTopicListQuery struct {
	TopicType CommunityTopicType `json:"topic_type"`
	Sort      string             `json:"sort"`
	Mine      string             `json:"mine"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
}

type CommunityTopicListItem struct {
	ID            string              `json:"id"`
	UserID        string              `json:"user_id"`
	Title         string              `json:"title"`
	Summary       string              `json:"summary"`
	TopicType     string              `json:"topic_type"`
	Stance        string              `json:"stance"`
	Status        string              `json:"status"`
	CommentCount  int                 `json:"comment_count"`
	LikeCount     int                 `json:"like_count"`
	FavoriteCount int                 `json:"favorite_count"`
	ReportCount   int                 `json:"report_count"`
	LastActiveAt  string              `json:"last_active_at"`
	CreatedAt     string              `json:"created_at"`
	LikedByMe     bool                `json:"liked_by_me"`
	FavoritedByMe bool                `json:"favorited_by_me"`
	LinkedTarget  CommunityTopicLink  `json:"linked_target"`
}

type CommunityTopicDetail struct {
	ID            string             `json:"id"`
	UserID        string             `json:"user_id"`
	Title         string             `json:"title"`
	Summary       string             `json:"summary"`
	Content       string             `json:"content"`
	TopicType     string             `json:"topic_type"`
	Stance        string             `json:"stance"`
	TimeHorizon   string             `json:"time_horizon"`
	ReasonText    string             `json:"reason_text"`
	RiskText      string             `json:"risk_text"`
	Status        string             `json:"status"`
	CommentCount  int                `json:"comment_count"`
	LikeCount     int                `json:"like_count"`
	FavoriteCount int                `json:"favorite_count"`
	ReportCount   int                `json:"report_count"`
	LastActiveAt  string             `json:"last_active_at"`
	CreatedAt     string             `json:"created_at"`
	UpdatedAt     string             `json:"updated_at"`
	LikedByMe     bool               `json:"liked_by_me"`
	FavoritedByMe bool               `json:"favorited_by_me"`
	LinkedTarget  CommunityTopicLink `json:"linked_target"`
}

type CommunityTopicCreateInput struct {
	UserID         string `json:"user_id"`
	Title          string `json:"title"`
	Summary        string `json:"summary"`
	Content        string `json:"content"`
	TopicType      string `json:"topic_type"`
	Stance         string `json:"stance"`
	TimeHorizon    string `json:"time_horizon"`
	ReasonText     string `json:"reason_text"`
	RiskText       string `json:"risk_text"`
	TargetType     string `json:"target_type"`
	TargetID       string `json:"target_id"`
	TargetSnapshot string `json:"target_snapshot"`
}

type CommunityComment struct {
	ID              string `json:"id"`
	TopicID         string `json:"topic_id"`
	TopicTitle      string `json:"topic_title,omitempty"`
	TopicSummary    string `json:"topic_summary,omitempty"`
	TopicStatus     string `json:"topic_status,omitempty"`
	UserID          string `json:"user_id"`
	ParentCommentID string `json:"parent_comment_id"`
	ReplyToUserID   string `json:"reply_to_user_id"`
	Content         string `json:"content"`
	Status          string `json:"status"`
	LikeCount       int    `json:"like_count"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	LikedByMe       bool   `json:"liked_by_me"`
	LinkedTarget    CommunityTopicLink `json:"linked_target"`
}

type CommunityCommentListQuery struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type CommunityCommentCreateInput struct {
	UserID          string `json:"user_id"`
	TopicID         string `json:"topic_id"`
	ParentCommentID string `json:"parent_comment_id"`
	ReplyToUserID   string `json:"reply_to_user_id"`
	Content         string `json:"content"`
}

type CommunityReactionInput struct {
	UserID       string `json:"user_id"`
	TargetType   string `json:"target_type"`
	TargetID     string `json:"target_id"`
	ReactionType string `json:"reaction_type"`
}

type CommunityReport struct {
	ID             string `json:"id"`
	ReporterUserID string `json:"reporter_user_id"`
	TargetType     string `json:"target_type"`
	TargetID       string `json:"target_id"`
	TopicID        string `json:"topic_id,omitempty"`
	TopicTitle     string `json:"topic_title,omitempty"`
	TopicSummary   string `json:"topic_summary,omitempty"`
	TargetContent  string `json:"target_content,omitempty"`
	TargetStatus   string `json:"target_status,omitempty"`
	LinkedTarget   CommunityTopicLink `json:"linked_target"`
	Reason         string `json:"reason"`
	Status         string `json:"status"`
	ReviewNote     string `json:"review_note"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
}

type CommunityReportCreateInput struct {
	ReporterUserID string `json:"reporter_user_id"`
	TargetType     string `json:"target_type"`
	TargetID       string `json:"target_id"`
	Reason         string `json:"reason"`
}

type CommunityAdminTopicQuery struct {
	TopicType string `json:"topic_type"`
	Status    string `json:"status"`
	UserID    string `json:"user_id"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

type CommunityAdminCommentQuery struct {
	TopicID   string `json:"topic_id"`
	Status    string `json:"status"`
	UserID    string `json:"user_id"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

type CommunityAdminReportQuery struct {
	Status    string `json:"status"`
	TargetType string `json:"target_type"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

type CommunityNotificationInput struct {
	UserID      string `json:"user_id"`
	Title       string `json:"title"`
	Content     string `json:"content"`
	MessageType string `json:"message_type"`
}
