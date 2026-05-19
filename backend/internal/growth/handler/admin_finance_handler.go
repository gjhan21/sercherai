package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/platform/utils"
)

type AdminFinanceHandler struct {
	AdminBaseHandler
}

func NewAdminFinanceHandler(base *AdminBaseHandler) *AdminFinanceHandler {
	return &AdminFinanceHandler{AdminBaseHandler: *base}
}

func (h *AdminFinanceHandler) ListInviteRecords(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")

	items, total, err := h.service.AdminListInviteRecords(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminFinanceHandler) ListRewardRecords(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")

	items, total, err := h.service.AdminListRewardRecords(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminFinanceHandler) ReviewRewardRecord(c *gin.Context) {
	id := c.Param("id")
	var req dto.ReviewRewardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminReviewRewardRecord(id, req.Status, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminFinanceHandler) ListReconciliation(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListReconciliation(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminFinanceHandler) RetryReconciliation(c *gin.Context) {
	batchID := c.Param("batch_id")
	if err := h.service.AdminRetryReconciliation(batchID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "FINANCE", "RETRY_RECONCILIATION", "BATCH", batchID, "", "RETRYING", "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminFinanceHandler) ListWithdrawRequests(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListWithdrawRequests(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminFinanceHandler) ReviewWithdrawRequest(c *gin.Context) {
	id := c.Param("id")
	var req dto.ReviewWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminReviewWithdrawRequest(id, req.Status, req.Reason); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "FINANCE", "REVIEW_WITHDRAW", "WITHDRAW_REQUEST", id, "", req.Status, req.Reason)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}
