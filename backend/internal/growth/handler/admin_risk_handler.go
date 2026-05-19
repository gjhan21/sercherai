package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/platform/utils"
)

type AdminRiskHandler struct {
	AdminBaseHandler
}

func NewAdminRiskHandler(base *AdminBaseHandler) *AdminRiskHandler {
	return &AdminRiskHandler{AdminBaseHandler: *base}
}

func (h *AdminRiskHandler) ListRiskRules(c *gin.Context) {
	items, err := h.service.AdminListRiskRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminRiskHandler) CreateRiskRule(c *gin.Context) {
	var req dto.RiskRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateRiskRule(req.RuleCode, req.RuleName, req.Threshold, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "RISK", "CREATE_RULE", "RISK_RULE", id, "", req.Status, req.RuleCode)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminRiskHandler) UpdateRiskRule(c *gin.Context) {
	id := c.Param("id")
	var req dto.RiskRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateRiskRule(id, req.Threshold, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "RISK", "UPDATE_RULE", "RISK_RULE", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminRiskHandler) ListRiskHits(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListRiskHits(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminRiskHandler) ReviewRiskHit(c *gin.Context) {
	id := c.Param("id")
	var req dto.ReviewRiskHitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminReviewRiskHit(id, req.Status, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "RISK", "REVIEW_HIT", "RISK_HIT", id, "", req.Status, req.Reason)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}
