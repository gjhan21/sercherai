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

func (h *AdminGrowthHandler) ListCommunityTopics(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListCommunityTopics(model.CommunityAdminTopicQuery{
		TopicType: strings.TrimSpace(c.Query("topic_type")),
		Status:    strings.TrimSpace(c.Query("status")),
		UserID:    strings.TrimSpace(c.Query("user_id")),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) UpdateCommunityTopicStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.CommunityStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateCommunityTopicStatus(id, req.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "topic not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "COMMUNITY", "UPDATE_TOPIC_STATUS", "COMMUNITY_TOPIC", id, "", strings.ToUpper(strings.TrimSpace(req.Status)), "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListCommunityComments(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListCommunityComments(model.CommunityAdminCommentQuery{
		TopicID:  strings.TrimSpace(c.Query("topic_id")),
		Status:   strings.TrimSpace(c.Query("status")),
		UserID:   strings.TrimSpace(c.Query("user_id")),
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) UpdateCommunityCommentStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.CommunityStatusUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateCommunityCommentStatus(id, req.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "comment not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "COMMUNITY", "UPDATE_COMMENT_STATUS", "COMMUNITY_COMMENT", id, "", strings.ToUpper(strings.TrimSpace(req.Status)), "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListCommunityReports(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListCommunityReports(model.CommunityAdminReportQuery{
		Status:     strings.TrimSpace(c.Query("status")),
		TargetType: strings.TrimSpace(c.Query("target_type")),
		Page:       page,
		PageSize:   pageSize,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ReviewCommunityReport(c *gin.Context) {
	id := c.Param("id")
	var req dto.CommunityReportReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminReviewCommunityReport(id, req.Status, req.ReviewNote); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "report not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "COMMUNITY", "REVIEW_REPORT", "COMMUNITY_REPORT", id, "PENDING", strings.ToUpper(strings.TrimSpace(req.Status)), strings.TrimSpace(req.ReviewNote))
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}
