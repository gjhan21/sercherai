package service

import (
	"fmt"
	"strings"

	"sercherai/backend/internal/growth/model"
)

func (s *growthService) ListCommunityTopics(userID string, query model.CommunityTopicListQuery) ([]model.CommunityTopicListItem, int, error) {
	return s.repo.ListCommunityTopics(userID, query)
}

func (s *growthService) GetCommunityTopic(userID string, topicID string) (model.CommunityTopicDetail, error) {
	return s.repo.GetCommunityTopic(userID, topicID)
}

func (s *growthService) CreateCommunityTopic(input model.CommunityTopicCreateInput) (model.CommunityTopicDetail, error) {
	return s.repo.CreateCommunityTopic(input)
}

func (s *growthService) ListCommunityComments(userID string, topicID string, query model.CommunityCommentListQuery) ([]model.CommunityComment, int, error) {
	return s.repo.ListCommunityComments(userID, topicID, query)
}

func (s *growthService) ListMyCommunityComments(userID string, page int, pageSize int) ([]model.CommunityComment, int, error) {
	return s.repo.ListMyCommunityComments(userID, page, pageSize)
}

func (s *growthService) CreateCommunityComment(input model.CommunityCommentCreateInput) (model.CommunityComment, error) {
	comment, err := s.repo.CreateCommunityComment(input)
	if err != nil {
		return model.CommunityComment{}, err
	}

	topic, err := s.repo.GetCommunityTopic("", input.TopicID)
	if err != nil {
		return comment, nil
	}
	if strings.TrimSpace(topic.UserID) != "" && topic.UserID != input.UserID {
		title := "你的讨论有新回复"
		content := fmt.Sprintf("《%s》收到新的讨论回复。", topic.Title)
		_ = s.repo.CreateCommunityNotification(model.CommunityNotificationInput{
			UserID:      topic.UserID,
			Title:       title,
			Content:     content,
			MessageType: "COMMUNITY",
		})
	}
	if strings.TrimSpace(input.ReplyToUserID) != "" && input.ReplyToUserID != input.UserID && input.ReplyToUserID != topic.UserID {
		title := "你的评论有新回复"
		content := fmt.Sprintf("《%s》下的评论收到新的回复。", topic.Title)
		_ = s.repo.CreateCommunityNotification(model.CommunityNotificationInput{
			UserID:      input.ReplyToUserID,
			Title:       title,
			Content:     content,
			MessageType: "COMMUNITY",
		})
	}
	return comment, nil
}

func (s *growthService) CreateCommunityReaction(input model.CommunityReactionInput) error {
	return s.repo.CreateCommunityReaction(input)
}

func (s *growthService) DeleteCommunityReaction(input model.CommunityReactionInput) error {
	return s.repo.DeleteCommunityReaction(input)
}

func (s *growthService) CreateCommunityReport(input model.CommunityReportCreateInput) (model.CommunityReport, error) {
	return s.repo.CreateCommunityReport(input)
}

func (s *growthService) AdminListCommunityTopics(query model.CommunityAdminTopicQuery) ([]model.CommunityTopicListItem, int, error) {
	return s.repo.AdminListCommunityTopics(query)
}

func (s *growthService) AdminUpdateCommunityTopicStatus(id string, status string) error {
	return s.repo.AdminUpdateCommunityTopicStatus(id, status)
}

func (s *growthService) AdminListCommunityComments(query model.CommunityAdminCommentQuery) ([]model.CommunityComment, int, error) {
	return s.repo.AdminListCommunityComments(query)
}

func (s *growthService) AdminUpdateCommunityCommentStatus(id string, status string) error {
	return s.repo.AdminUpdateCommunityCommentStatus(id, status)
}

func (s *growthService) AdminListCommunityReports(query model.CommunityAdminReportQuery) ([]model.CommunityReport, int, error) {
	return s.repo.AdminListCommunityReports(query)
}

func (s *growthService) AdminReviewCommunityReport(id string, status string, reviewNote string) error {
	return s.repo.AdminReviewCommunityReport(id, status, reviewNote)
}
