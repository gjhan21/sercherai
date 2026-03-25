package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
)

func (h *UserGrowthHandler) ListPublicCommunityTopics(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListCommunityTopics(optionalUserID(c), model.CommunityTopicListQuery{
		TopicType: model.CommunityTopicType(strings.ToUpper(strings.TrimSpace(c.Query("topic_type")))),
		Sort:      strings.TrimSpace(c.Query("sort")),
		Mine:      strings.TrimSpace(c.Query("mine")),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) GetPublicCommunityTopic(c *gin.Context) {
	item, err := h.service.GetCommunityTopic(optionalUserID(c), c.Param("id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "topic not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) ListPublicCommunityComments(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListCommunityComments(optionalUserID(c), c.Param("id"), model.CommunityCommentListQuery{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) ListCommunityTopics(c *gin.Context) {
	h.ListPublicCommunityTopics(c)
}

func (h *UserGrowthHandler) GetCommunityTopic(c *gin.Context) {
	h.GetPublicCommunityTopic(c)
}

func (h *UserGrowthHandler) ListCommunityComments(c *gin.Context) {
	h.ListPublicCommunityComments(c)
}

func (h *UserGrowthHandler) ListMyCommunityTopics(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListCommunityTopics(userID, model.CommunityTopicListQuery{
		TopicType: model.CommunityTopicType(strings.ToUpper(strings.TrimSpace(c.Query("topic_type")))),
		Sort:      strings.TrimSpace(c.Query("sort")),
		Mine:      "topics",
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) ListMyCommunityComments(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListMyCommunityComments(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) CreateCommunityTopic(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req dto.CommunityTopicCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	item, err := h.service.CreateCommunityTopic(model.CommunityTopicCreateInput{
		UserID:         userID,
		Title:          req.Title,
		Summary:        req.Summary,
		Content:        req.Content,
		TopicType:      req.TopicType,
		Stance:         req.Stance,
		TimeHorizon:    req.TimeHorizon,
		ReasonText:     req.ReasonText,
		RiskText:       req.RiskText,
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		TargetSnapshot: req.TargetSnapshot,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) CreateCommunityComment(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req dto.CommunityCommentCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	item, err := h.service.CreateCommunityComment(model.CommunityCommentCreateInput{
		UserID:          userID,
		TopicID:         c.Param("id"),
		ParentCommentID: req.ParentCommentID,
		ReplyToUserID:   req.ReplyToUserID,
		Content:         req.Content,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "topic not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) CreateCommunityReaction(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req dto.CommunityReactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.CreateCommunityReaction(model.CommunityReactionInput{
		UserID:       userID,
		TargetType:   req.TargetType,
		TargetID:     req.TargetID,
		ReactionType: req.ReactionType,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *UserGrowthHandler) DeleteCommunityReaction(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	input := model.CommunityReactionInput{
		UserID:       userID,
		TargetType:   c.Query("target_type"),
		TargetID:     c.Query("target_id"),
		ReactionType: c.Query("reaction_type"),
	}
	if strings.TrimSpace(input.TargetType) == "" || strings.TrimSpace(input.TargetID) == "" || strings.TrimSpace(input.ReactionType) == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "target_type,target_id,reaction_type required", Data: struct{}{}})
		return
	}
	if err := h.service.DeleteCommunityReaction(input); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *UserGrowthHandler) CreateCommunityReport(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req dto.CommunityReportCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	item, err := h.service.CreateCommunityReport(model.CommunityReportCreateInput{
		ReporterUserID: userID,
		TargetType:     req.TargetType,
		TargetID:       req.TargetID,
		Reason:         req.Reason,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}
