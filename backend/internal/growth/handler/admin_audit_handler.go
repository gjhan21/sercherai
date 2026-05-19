package handler

import (
	"bytes"
	"encoding/csv"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/platform/utils"
)

type AdminAuditHandler struct {
	AdminBaseHandler
}

func NewAdminAuditHandler(base *AdminBaseHandler) *AdminAuditHandler {
	return &AdminAuditHandler{AdminBaseHandler: *base}
}

func (h *AdminAuditHandler) DashboardOverview(c *gin.Context) {
	item, err := h.service.AdminDashboardOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminAuditHandler) ListOperationLogs(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	module := c.Query("module")
	action := c.Query("action")
	operator := c.Query("operator_user_id")
	items, total, err := h.service.AdminListOperationLogs(module, action, operator, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminAuditHandler) ExportOperationLogsCSV(c *gin.Context) {
	module := c.Query("module")
	action := c.Query("action")
	operator := c.Query("operator_user_id")
	items, _, err := h.service.AdminListOperationLogs(module, action, operator, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "module", "action", "target_type", "target_id", "operator_user_id", "before_value", "after_value", "reason", "created_at"})
	for _, it := range items {
		_ = writer.Write([]string{it.ID, it.Module, it.Action, it.TargetType, it.TargetID, it.OperatorUserID, it.BeforeValue, it.AfterValue, it.Reason, it.CreatedAt})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=admin_operation_logs.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminAuditHandler) ListAuditEvents(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	filter := model.AdminAuditEventFilter{
		EventDomain: c.Query("event_domain"),
		EventType:   c.Query("event_type"),
		Level:       c.Query("level"),
		Module:      c.Query("module"),
		ObjectType:  c.Query("object_type"),
		ObjectID:    c.Query("object_id"),
		ActorUserID: c.Query("actor_user_id"),
		Status:      c.Query("status"),
	}
	items, total, err := h.service.AdminListAuditEvents(filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminAuditHandler) GetAuditEventSummary(c *gin.Context) {
	summary, err := h.service.AdminGetAuditEventSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(summary))
}

func (h *AdminAuditHandler) ListWorkflowMessages(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	module := c.Query("module")
	eventType := c.Query("event_type")
	isRead := c.Query("is_read")
	receiverID := strings.TrimSpace(c.Query("receiver_id"))
	if receiverID == "" {
		operatorVal, _ := c.Get("user_id")
		receiverID, _ = operatorVal.(string)
	}
	items, total, err := h.service.AdminListWorkflowMessages(module, eventType, isRead, receiverID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminAuditHandler) ExportWorkflowMessagesCSV(c *gin.Context) {
	module := c.Query("module")
	eventType := c.Query("event_type")
	isRead := c.Query("is_read")
	receiverID := strings.TrimSpace(c.Query("receiver_id"))
	if receiverID == "" {
		operatorVal, _ := c.Get("user_id")
		receiverID, _ = operatorVal.(string)
	}
	items, _, err := h.service.AdminListWorkflowMessages(module, eventType, isRead, receiverID, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "review_id", "target_id", "module", "receiver_id", "sender_id", "event_type", "title", "content", "is_read", "created_at", "read_at"})
	for _, it := range items {
		isReadVal := "false"
		if it.IsRead {
			isReadVal = "true"
		}
		_ = writer.Write([]string{it.ID, it.ReviewID, it.TargetID, it.Module, it.ReceiverID, it.SenderID, it.EventType, it.Title, it.Content, isReadVal, it.CreatedAt, it.ReadAt})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=workflow_messages.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminAuditHandler) UpdateWorkflowMessageRead(c *gin.Context) {
	id := c.Param("id")
	var req dto.WorkflowMessageReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateWorkflowMessageRead(id, req.IsRead); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminAuditHandler) CountUnreadWorkflowMessages(c *gin.Context) {
	module := c.Query("module")
	eventType := c.Query("event_type")
	receiverID := strings.TrimSpace(c.Query("receiver_id"))
	if receiverID == "" {
		operatorVal, _ := c.Get("user_id")
		receiverID, _ = operatorVal.(string)
	}
	total, err := h.service.AdminCountUnreadWorkflowMessages(module, eventType, receiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"unread_count": total}))
}

func (h *AdminAuditHandler) BulkReadWorkflowMessages(c *gin.Context) {
	var req dto.WorkflowMessageBulkReadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	receiverID := strings.TrimSpace(req.ReceiverID)
	if receiverID == "" {
		operatorVal, _ := c.Get("user_id")
		receiverID, _ = operatorVal.(string)
	}
	affected, err := h.service.AdminBulkReadWorkflowMessages(req.Module, req.EventType, receiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"affected": affected}))
}

func (h *AdminAuditHandler) ListReviewTasks(c *gin.Context) {
	page, pageSize := utils.ParsePage(c)
	module := c.Query("module")
	status := c.Query("status")
	submitterID := c.Query("submitter_id")
	reviewerID := c.Query("reviewer_id")
	items, total, err := h.service.AdminListReviewTasks(module, status, submitterID, reviewerID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminAuditHandler) ExportReviewTasksCSV(c *gin.Context) {
	module := c.Query("module")
	status := c.Query("status")
	submitterID := c.Query("submitter_id")
	reviewerID := c.Query("reviewer_id")
	items, _, err := h.service.AdminListReviewTasks(module, status, submitterID, reviewerID, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "module", "target_id", "submitter_id", "reviewer_id", "status", "submit_note", "review_note", "submitted_at", "reviewed_at"})
	for _, it := range items {
		_ = writer.Write([]string{it.ID, it.Module, it.TargetID, it.SubmitterID, it.ReviewerID, it.Status, it.SubmitNote, it.ReviewNote, it.SubmittedAt, it.ReviewedAt})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=review_tasks.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminAuditHandler) WorkflowMetrics(c *gin.Context) {
	module := c.Query("module")
	receiverID := strings.TrimSpace(c.Query("receiver_id"))
	if receiverID == "" {
		operatorVal, _ := c.Get("user_id")
		receiverID, _ = operatorVal.(string)
	}
	item, err := h.service.AdminGetWorkflowMetrics(module, receiverID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminAuditHandler) SubmitReviewTask(c *gin.Context) {
	var req dto.ReviewSubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	id, err := h.service.AdminSubmitReviewTask(req.Module, req.TargetID, operator, req.ReviewerID, req.SubmitNote)
	if err != nil {
		status, code := resolveWorkflowReviewHTTPError(err)
		c.JSON(status, dto.APIResponse{Code: code, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "WORKFLOW", "SUBMIT_REVIEW", strings.ToUpper(req.Module), req.TargetID, "", "REVIEWING", req.SubmitNote)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminAuditHandler) ReviewTaskDecision(c *gin.Context) {
	id := c.Param("id")
	var req dto.ReviewDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminReviewTaskDecision(id, req.Status, operator, req.ReviewNote); err != nil {
		status, code := resolveWorkflowReviewHTTPError(err)
		c.JSON(status, dto.APIResponse{Code: code, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "WORKFLOW", "REVIEW_DECISION", "REVIEW_TASK", id, "PENDING", strings.ToUpper(req.Status), req.ReviewNote)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminAuditHandler) AssignReviewTask(c *gin.Context) {
	id := c.Param("id")
	var req dto.ReviewAssignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminAssignReviewTask(id, req.ReviewerID); err != nil {
		status, code := resolveWorkflowReviewHTTPError(err)
		c.JSON(status, dto.APIResponse{Code: code, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "WORKFLOW", "ASSIGN_REVIEW", "REVIEW_TASK", id, "", req.ReviewerID, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}
