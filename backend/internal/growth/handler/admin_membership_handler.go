package handler

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/platform/utils"
)

type AdminMembershipHandler struct {
	AdminBaseHandler
}

func NewAdminMembershipHandler(base *AdminBaseHandler) *AdminMembershipHandler {
	return &AdminMembershipHandler{AdminBaseHandler: *base}
}

func (h *AdminMembershipHandler) ListMembershipProducts(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListMembershipProducts(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminMembershipHandler) CreateMembershipProduct(c *gin.Context) {
	var req dto.MembershipProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateMembershipProduct(req.Name, req.Price, req.Status, req.MemberLevel, req.DurationDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "CREATE_PRODUCT", "MEMBERSHIP_PRODUCT", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminMembershipHandler) UpdateMembershipProduct(c *gin.Context) {
	id := c.Param("id")
	var req dto.MembershipProductUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateMembershipProduct(id, req.Name, req.Price, req.Status, req.MemberLevel, req.DurationDays); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "membership product not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "UPDATE_PRODUCT", "MEMBERSHIP_PRODUCT", id, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminMembershipHandler) UpdateMembershipProductStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.MembershipProductStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateMembershipProductStatus(id, req.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "membership product not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "UPDATE_PRODUCT_STATUS", "MEMBERSHIP_PRODUCT", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminMembershipHandler) ListMembershipOrders(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	userID := c.Query("user_id")
	items, total, err := h.service.AdminListMembershipOrders(status, userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminMembershipHandler) ListMembershipOrdersCSV(c *gin.Context) {
	status := c.Query("status")
	userID := c.Query("user_id")
	items, _, err := h.service.AdminListMembershipOrders(status, userID, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "user_id", "product_id", "amount", "status", "paid_at", "created_at"})
	for _, it := range items {
		_ = writer.Write([]string{
			it.ID, it.UserID, it.ProductID,
			strconv.FormatFloat(it.Amount, 'f', -1, 64),
			it.Status, it.PaidAt, it.CreatedAt,
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=membership_orders.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminMembershipHandler) UpdateMembershipOrderStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.MembershipOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateMembershipOrderStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "UPDATE_ORDER_STATUS", "MEMBERSHIP_ORDER", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminMembershipHandler) ListVIPQuotaConfigs(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	memberLevel := c.Query("member_level")
	status := c.Query("status")
	items, total, err := h.service.AdminListVIPQuotaConfigs(memberLevel, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminMembershipHandler) CreateVIPQuotaConfig(c *gin.Context) {
	var req dto.VIPQuotaConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if _, err := time.Parse(time.RFC3339, req.EffectiveAt); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "effective_at must be RFC3339 format", Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateVIPQuotaConfig(model.VIPQuotaConfig{
		MemberLevel:        req.MemberLevel,
		DocReadLimit:       req.DocReadLimit,
		NewsSubscribeLimit: req.NewsSubscribeLimit,
		ResetCycle:         req.ResetCycle,
		Status:             req.Status,
		EffectiveAt:        req.EffectiveAt,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "CREATE_VIP_QUOTA_CONFIG", "VIP_QUOTA_CONFIG", id, "", req.MemberLevel, req.EffectiveAt)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminMembershipHandler) UpdateVIPQuotaConfig(c *gin.Context) {
	id := c.Param("id")
	var req dto.VIPQuotaConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if _, err := time.Parse(time.RFC3339, req.EffectiveAt); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "effective_at must be RFC3339 format", Data: struct{}{}})
		return
	}
	err := h.service.AdminUpdateVIPQuotaConfig(id, model.VIPQuotaConfig{
		DocReadLimit:       req.DocReadLimit,
		NewsSubscribeLimit: req.NewsSubscribeLimit,
		ResetCycle:         req.ResetCycle,
		Status:             req.Status,
		EffectiveAt:        req.EffectiveAt,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "quota config not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MEMBERSHIP", "UPDATE_VIP_QUOTA_CONFIG", "VIP_QUOTA_CONFIG", id, "", req.Status, req.EffectiveAt)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminMembershipHandler) ListUserQuotas(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	userID := strings.TrimSpace(c.Query("user_id"))
	periodKey := strings.TrimSpace(c.Query("period_key"))
	items, total, err := h.service.AdminListUserQuotaUsages(userID, periodKey, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminMembershipHandler) AdjustUserQuota(c *gin.Context) {
	userID := c.Param("user_id")
	var req dto.UserQuotaAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminAdjustUserQuota(userID, req.PeriodKey, req.DocReadDelta, req.NewsSubscribeDelta); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "user not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	summary := fmt.Sprintf("period=%s,doc_delta=%d,news_delta=%d", req.PeriodKey, req.DocReadDelta, req.NewsSubscribeDelta)
	h.writeOperationLog(c, "MEMBERSHIP", "ADJUST_USER_QUOTA", "USER_QUOTA", userID, "", summary, req.Reason)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminMembershipHandler) ListInviteRecords(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListInviteRecords(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminMembershipHandler) ListRewardRecords(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListRewardRecords(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminMembershipHandler) ReviewRewardRecord(c *gin.Context) {
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
	h.writeOperationLog(c, "FINANCE", "REVIEW_REWARD", "REWARD_RECORD", id, "", req.Status, req.Reason)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminMembershipHandler) ListReconciliation(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListReconciliation(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminMembershipHandler) RetryReconciliation(c *gin.Context) {
	batchID := c.Param("batch_id")
	if err := h.service.AdminRetryReconciliation(batchID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "FINANCE", "RETRY_RECONCILIATION", "RECONCILIATION_BATCH", batchID, "", "RETRY", "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminMembershipHandler) ListWithdrawRequests(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	items, total, err := h.service.AdminListWithdrawRequests(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminMembershipHandler) ReviewWithdrawRequest(c *gin.Context) {
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
