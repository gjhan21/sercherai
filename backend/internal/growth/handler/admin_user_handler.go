package handler

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/platform/utils"
)

type AdminUserHandler struct {
	AdminBaseHandler
}

func NewAdminUserHandler(base *AdminBaseHandler) *AdminUserHandler {
	return &AdminUserHandler{AdminBaseHandler: *base}
}

func (h *AdminUserHandler) ListUsers(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	status := c.Query("status")
	memberLevel := c.Query("member_level")
	registrationSource := c.Query("registration_source")
	items, total, err := h.service.AdminListUsers(status, memberLevel, registrationSource, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminUserHandler) UserSourceSummary(c *gin.Context) {
	status := c.Query("status")
	memberLevel := c.Query("member_level")
	registrationSource := c.Query("registration_source")
	item, err := h.service.AdminGetUserSourceSummary(status, memberLevel, registrationSource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminUserHandler) ExportUsersCSV(c *gin.Context) {
	status := c.Query("status")
	memberLevel := c.Query("member_level")
	registrationSource := c.Query("registration_source")
	items, _, err := h.service.AdminListUsers(status, memberLevel, registrationSource, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "phone", "email", "status", "member_level", "registration_source", "inviter_user_id", "invite_code", "invite_registered_at", "created_at"})
	for _, it := range items {
		_ = writer.Write([]string{
			it.ID,
			it.Phone,
			it.Email,
			it.Status,
			it.MemberLevel,
			it.RegistrationSource,
			it.InviterUserID,
			it.InviteCode,
			it.InviteRegisteredAt,
			it.CreatedAt,
		})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=admin_users.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminUserHandler) UpdateUserStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateUserStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "USER", "UPDATE_STATUS", "USER", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminUserHandler) UpdateUserMemberLevel(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUserMemberLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateUserMemberLevel(id, req.MemberLevel); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "USER", "UPDATE_MEMBER_LEVEL", "USER", id, "", req.MemberLevel, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminUserHandler) ResetUserPassword(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "user id required", Data: struct{}{}})
		return
	}
	if strings.HasPrefix(strings.ToLower(id), "admin_") {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40002, Message: "admin user password should be reset from access module", Data: struct{}{}})
		return
	}

	var req dto.AdminResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	passwordHash, err := bcryptHash(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminResetUserPasswordHash(id, passwordHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40404, Message: "user not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "USER", "RESET_PASSWORD", "USER", id, "", "UPDATED", "ADMIN_RESET")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminUserHandler) ListBrowseHistories(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	userID := strings.TrimSpace(c.Query("user_id"))
	contentType := strings.ToUpper(strings.TrimSpace(c.Query("content_type")))
	keyword := strings.TrimSpace(c.Query("keyword"))
	items, total, err := h.service.AdminListBrowseHistories(userID, contentType, keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminUserHandler) BrowseHistorySummary(c *gin.Context) {
	item, err := h.service.AdminGetBrowseHistorySummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminUserHandler) BrowseHistoryTrend(c *gin.Context) {
	days := parseIntOrDefault(strings.TrimSpace(c.Query("days")), 7)
	if days < 1 {
		days = 7
	}
	if days > 30 {
		days = 30
	}
	items, err := h.service.AdminGetBrowseHistoryTrend(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "days": days}))
}

func (h *AdminUserHandler) ListBrowseUserSegments(c *gin.Context) {
	limit := parseIntOrDefault(strings.TrimSpace(c.Query("limit")), 10)
	if limit > 50 {
		limit = 50
	}
	items, err := h.service.AdminListBrowseUserSegments(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminUserHandler) ExportBrowseHistoriesCSV(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("user_id"))
	contentType := strings.ToUpper(strings.TrimSpace(c.Query("content_type")))
	keyword := strings.TrimSpace(c.Query("keyword"))

	items, _, err := h.service.AdminListBrowseHistories(userID, contentType, keyword, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "user_id", "user_phone", "content_type", "content_id", "title", "source_page", "viewed_at"})
	for _, it := range items {
		_ = writer.Write([]string{it.ID, it.UserID, it.UserPhone, it.ContentType, it.ContentID, it.Title, it.SourcePage, it.ViewedAt})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=browse_histories.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminUserHandler) ListUserMessages(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	userID := strings.TrimSpace(c.Query("user_id"))
	messageType := strings.TrimSpace(c.Query("type"))
	readStatus := strings.TrimSpace(c.Query("read_status"))
	items, total, err := h.service.AdminListUserMessages(userID, messageType, readStatus, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminUserHandler) CreateUserMessages(c *gin.Context) {
	var req dto.AdminUserMessageCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	title := strings.TrimSpace(req.Title)
	content := strings.TrimSpace(req.Content)
	messageType := strings.ToUpper(strings.TrimSpace(req.Type))
	if title == "" || content == "" || messageType == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "title, content and type required", Data: struct{}{}})
		return
	}

	targetUserIDs := uniqueNonEmptyStrings(req.UserIDs)
	if len(targetUserIDs) == 0 {
		users, _, err := h.service.AdminListUsers("ACTIVE", "", "", 1, 10000)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		for _, user := range users {
			targetUserIDs = append(targetUserIDs, user.ID)
		}
		targetUserIDs = uniqueNonEmptyStrings(targetUserIDs)
	}

	if len(targetUserIDs) == 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "no target users", Data: struct{}{}})
		return
	}

	sentCount, failures, err := h.service.AdminCreateUserMessages(targetUserIDs, title, content, messageType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	h.writeOperationLog(
		c,
		"USER",
		"SEND_MESSAGES",
		"USER_MESSAGE",
		"",
		"",
		"",
		fmt.Sprintf("type=%s sent=%d failed=%d", messageType, sentCount, len(failures)),
	)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"sent_count":   sentCount,
		"failed_count": len(failures),
		"failures":     failures,
	}))
}

func (h *AdminUserHandler) GetUserCenterOverview(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "user id required", Data: struct{}{}})
		return
	}

	limit := parseIntOrDefault(strings.TrimSpace(c.Query("limit")), 50)
	if limit > 200 {
		limit = 200
	}

	profile, err := h.service.GetUserProfile(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "user not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	quota, err := h.service.GetMembershipQuota(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	subscriptions, _, err := h.service.ListSubscriptions(userID, 1, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	browseHistory, _, err := h.service.ListBrowseHistory(userID, "", 1, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	rechargeRecords, _, err := h.service.ListRechargeRecords(userID, "", 1, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	membershipOrders, _, err := h.service.ListMembershipOrders(userID, "", 1, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	shareLinks, err := h.service.ListShareLinks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	inviteSummary, err := h.service.GetUserInviteSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	inviteRecords, _, err := h.service.ListInviteRecords(userID, 1, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	paidOrderCount := 0
	pendingOrderCount := 0
	paidAmountTotal := 0.0
	for _, item := range membershipOrders {
		status := strings.ToUpper(strings.TrimSpace(item.Status))
		if status == "PAID" || status == "SUCCESS" {
			paidOrderCount += 1
			paidAmountTotal += item.Amount
		} else {
			pendingOrderCount += 1
		}
	}
	for _, item := range rechargeRecords {
		status := strings.ToUpper(strings.TrimSpace(item.Status))
		if status == "PAID" || status == "SUCCESS" {
			paidAmountTotal += item.Amount
		}
	}

	activeSubscriptionCount := 0
	for _, item := range subscriptions {
		if strings.ToUpper(strings.TrimSpace(item.Status)) == "ACTIVE" {
			activeSubscriptionCount += 1
		}
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{
		"user_profile":     profile,
		"membership_quota": quota,
		"payment_summary": gin.H{
			"paid_order_count":    paidOrderCount,
			"pending_order_count": pendingOrderCount,
			"paid_amount_total":   paidAmountTotal,
			"recharge_count":      len(rechargeRecords),
		},
		"reading_summary": gin.H{
			"browse_count": len(browseHistory),
		},
		"subscription_summary": gin.H{
			"total_count":  len(subscriptions),
			"active_count": activeSubscriptionCount,
		},
		"invite_summary":    inviteSummary,
		"membership_orders": membershipOrders,
		"recharge_records":  rechargeRecords,
		"browse_history":    browseHistory,
		"subscriptions":     subscriptions,
		"share_links":       shareLinks,
		"invite_records":    inviteRecords,
	}))
}

func (h *AdminUserHandler) UpdateUserSubscription(c *gin.Context) {
	userID := strings.TrimSpace(c.Param("id"))
	subscriptionID := strings.TrimSpace(c.Param("sub_id"))
	if userID == "" || subscriptionID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "user id and subscription id required", Data: struct{}{}})
		return
	}

	var req dto.SubscriptionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	if err := h.service.UpdateSubscription(userID, subscriptionID, req.Frequency, req.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "subscription not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	h.writeOperationLog(
		c,
		"USER",
		"UPDATE_SUBSCRIPTION",
		"SUBSCRIPTION",
		subscriptionID,
		"",
		req.Status,
		fmt.Sprintf("user_id=%s frequency=%s", userID, req.Frequency),
	)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

