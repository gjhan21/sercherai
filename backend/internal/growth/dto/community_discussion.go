package dto

type CommunityTopicCreateRequest struct {
	Title          string `json:"title" binding:"required"`
	Summary        string `json:"summary"`
	Content        string `json:"content" binding:"required"`
	TopicType      string `json:"topic_type" binding:"required,oneof=STOCK FUTURES NEWS STRATEGY"`
	Stance         string `json:"stance" binding:"required,oneof=BULLISH BEARISH WATCH"`
	TimeHorizon    string `json:"time_horizon"`
	ReasonText     string `json:"reason_text" binding:"required"`
	RiskText       string `json:"risk_text" binding:"required"`
	TargetType     string `json:"target_type" binding:"required,oneof=STOCK FUTURES NEWS_ARTICLE STRATEGY_ITEM"`
	TargetID       string `json:"target_id" binding:"required"`
	TargetSnapshot string `json:"target_snapshot"`
}

type CommunityCommentCreateRequest struct {
	ParentCommentID string `json:"parent_comment_id"`
	ReplyToUserID   string `json:"reply_to_user_id"`
	Content         string `json:"content" binding:"required"`
}

type CommunityReactionRequest struct {
	TargetType   string `json:"target_type" binding:"required,oneof=TOPIC COMMENT"`
	TargetID     string `json:"target_id" binding:"required"`
	ReactionType string `json:"reaction_type" binding:"required,oneof=LIKE FAVORITE"`
}

type CommunityReportCreateRequest struct {
	TargetType string `json:"target_type" binding:"required,oneof=TOPIC COMMENT"`
	TargetID   string `json:"target_id" binding:"required"`
	Reason     string `json:"reason" binding:"required"`
}

type CommunityStatusUpdateRequest struct {
	Status string `json:"status" binding:"required,oneof=PUBLISHED PENDING_REVIEW HIDDEN DELETED"`
}

type CommunityReportReviewRequest struct {
	Status     string `json:"status" binding:"required,oneof=RESOLVED REJECTED"`
	ReviewNote string `json:"review_note"`
}
