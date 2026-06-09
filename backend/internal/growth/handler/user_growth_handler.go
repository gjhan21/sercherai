package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"math/rand"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/config"
)

type UserGrowthHandler struct {
	service service.GrowthService
	cfg     config.Config
}

func NewUserGrowthHandler(service service.GrowthService, cfg config.Config) *UserGrowthHandler {
	return &UserGrowthHandler{service: service, cfg: cfg}
}


func (h *UserGrowthHandler) ListBrowseHistory(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	contentType := c.Query("content_type")

	items, total, err := h.service.ListBrowseHistory(userID, contentType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) DeleteBrowseHistoryItem(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	id := c.Param("id")
	if err := h.service.DeleteBrowseHistoryItem(userID, id); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *UserGrowthHandler) ClearBrowseHistory(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	if err := h.service.ClearBrowseHistory(userID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *UserGrowthHandler) GetUserProfile(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
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
	c.JSON(http.StatusOK, dto.OK(profile))
}

func (h *UserGrowthHandler) UpdateUserProfile(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req dto.UpdateUserProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.UpdateUserProfileEmail(userID, req.Email); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *UserGrowthHandler) ListRechargeRecords(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	status := c.Query("status")

	items, total, err := h.service.ListRechargeRecords(userID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) ListShareLinks(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	items, err := h.service.ListShareLinks(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *UserGrowthHandler) CreateShareLink(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req dto.CreateShareLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	item, err := h.service.CreateShareLink(userID, req.Channel, req.ExpiredAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) ListSubscriptions(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListSubscriptions(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) CreateSubscription(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req dto.SubscriptionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.CreateSubscription(userID, req.Type, req.Scope, req.Frequency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *UserGrowthHandler) UpdateSubscription(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	id := c.Param("id")
	var req dto.SubscriptionUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.UpdateSubscription(userID, id, req.Frequency, req.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "subscription not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *UserGrowthHandler) ListMessages(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListMessages(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) ReadMessage(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	id := c.Param("id")
	if err := h.service.MarkMessageRead(userID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "message not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *UserGrowthHandler) ListInviteRecords(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListInviteRecords(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) GetInviteSummary(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	summary, err := h.service.GetUserInviteSummary(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(summary))
}

func (h *UserGrowthHandler) ListRewardRecords(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListRewardRecords(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) GetMembershipQuota(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	item, err := h.service.GetMembershipQuota(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) ListMembershipProducts(c *gin.Context) {
	_, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	status := strings.TrimSpace(c.Query("status"))
	if status == "" {
		status = "ACTIVE"
	}
	items, total, err := h.service.ListMembershipProducts(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) CreateMembershipOrder(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req dto.CreateMembershipOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	order, err := h.service.CreateMembershipOrder(userID, req.ProductID, req.PayChannel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	experimentBindingError := ""
	if req.Experiment != nil && strings.TrimSpace(req.Experiment.ExperimentKey) != "" && strings.TrimSpace(req.Experiment.VariantKey) != "" {
		conversionType := "PAYMENT_SUCCESS"
		userStage := strings.ToUpper(strings.TrimSpace(req.Experiment.UserStage))
		if userStage == "VIP" || userStage == "EXPIRED" {
			conversionType = "RENEWAL_SUCCESS"
		}
		if err := h.service.BindMembershipOrderExperiment(order.OrderNo, model.ExperimentOrderAttribution{
			OrderNo:        order.OrderNo,
			ExperimentKey:  strings.TrimSpace(req.Experiment.ExperimentKey),
			VariantKey:     strings.TrimSpace(req.Experiment.VariantKey),
			PageKey:        strings.TrimSpace(req.Experiment.PageKey),
			TargetKey:      strings.TrimSpace(req.Experiment.TargetKey),
			UserStage:      userStage,
			AnonymousID:    strings.TrimSpace(req.Experiment.AnonymousID),
			SessionID:      strings.TrimSpace(req.Experiment.SessionID),
			Pathname:       strings.TrimSpace(req.Experiment.Pathname),
			Referrer:       strings.TrimSpace(req.Experiment.Referrer),
			ConversionType: conversionType,
			Metadata:       req.Experiment.Metadata,
		}); err != nil {
			experimentBindingError = err.Error()
		}
	}
	if strings.EqualFold(strings.TrimSpace(req.PayChannel), "YOLKPAY") {
		payload := membershipOrderToMap(order)
		if experimentBindingError != "" {
			payload["experiment_binding_error"] = experimentBindingError
		}
		paymentAction, payErr := h.createYolkPayPaymentAction(c, order)
		if payErr != nil {
			payload["payment_initialized"] = false
			payload["payment_error"] = payErr.Error()
			c.JSON(http.StatusOK, dto.OK(payload))
			return
		}
		payload["payment_initialized"] = true
		payload["payment_action"] = paymentAction
		c.JSON(http.StatusOK, dto.OK(payload))
		return
	}
	if experimentBindingError != "" {
		c.JSON(http.StatusOK, dto.OK(gin.H{
			"order":                    order,
			"experiment_binding_error": experimentBindingError,
		}))
		return
	}
	c.JSON(http.StatusOK, dto.OK(order))
}

func membershipOrderToMap(order model.MembershipOrderAdmin) gin.H {
	return gin.H{
		"id":          order.ID,
		"order_no":    order.OrderNo,
		"user_id":     order.UserID,
		"product_id":  order.ProductID,
		"amount":      order.Amount,
		"pay_channel": order.PayChannel,
		"status":      order.Status,
		"paid_at":     order.PaidAt,
		"created_at":  order.CreatedAt,
	}
}

func (h *UserGrowthHandler) createYolkPayPaymentAction(c *gin.Context, order model.MembershipOrderAdmin) (gin.H, error) {
	cfg, err := h.resolveYolkPayConfig()
	if err != nil {
		return nil, fmt.Errorf("加载蛋黄支付配置失败: %w", err)
	}
	if !cfg.PaymentEnabled {
		return nil, errors.New("payment.enabled 未开启")
	}
	if !cfg.Enabled {
		return nil, errors.New("payment.channel.yolkpay.enabled 未开启")
	}
	if strings.TrimSpace(cfg.PID) == "" || strings.TrimSpace(cfg.Key) == "" {
		return nil, errors.New("payment.channel.yolkpay.pid 或 key 未配置")
	}

	clientIP := strings.TrimSpace(c.ClientIP())
	if clientIP == "" || clientIP == "::1" || clientIP == "0:0:0:0:0:0:0:1" {
		clientIP = "127.0.0.1"
	}
	params := map[string]string{
		"pid":          strings.TrimSpace(cfg.PID),
		"type":         strings.TrimSpace(cfg.PayType),
		"out_trade_no": strings.TrimSpace(order.OrderNo),
		"notify_url":   strings.TrimSpace(cfg.NotifyURL),
		"return_url":   strings.TrimSpace(cfg.ReturnURL),
		"name":         fmt.Sprintf("会员订阅-%s", strings.TrimSpace(order.OrderNo)),
		"money":        fmt.Sprintf("%.2f", order.Amount),
		"clientip":     clientIP,
		"device":       strings.TrimSpace(cfg.Device),
		"param":        strings.TrimSpace(order.UserID),
	}
	params["sign"] = buildYolkPaySign(params, cfg.Key)
	params["sign_type"] = "MD5"

	form := url.Values{}
	for key, value := range params {
		form.Set(key, value)
	}
	req, err := http.NewRequest(http.MethodPost, buildYolkPayGatewayURL(cfg.Gateway, cfg.MAPIPath), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := (&http.Client{Timeout: 15 * time.Second}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 128*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(bodyBytes))
		if msg == "" {
			msg = resp.Status
		}
		return nil, fmt.Errorf("蛋黄支付网关返回异常: %s", msg)
	}
	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		return nil, fmt.Errorf("蛋黄支付返回解析失败: %w", err)
	}
	code := parseYolkPayCode(payload["code"])
	msg := stringifyYolkPayValue(payload["msg"])
	if code != 1 {
		if msg == "" {
			msg = "下单失败"
		}
		return nil, fmt.Errorf("蛋黄支付下单失败: %s", msg)
	}
	payURL := stringifyYolkPayValue(payload["payurl"])
	qrCode := stringifyYolkPayValue(payload["qrcode"])
	urlScheme := stringifyYolkPayValue(payload["urlscheme"])
	tradeNo := stringifyYolkPayValue(payload["trade_no"])
	return gin.H{
		"channel":   "YOLKPAY",
		"pay_url":   payURL,
		"qrcode":    qrCode,
		"urlscheme": urlScheme,
		"trade_no":  tradeNo,
		"raw":       payload,
	}, nil
}

func (h *UserGrowthHandler) ListMembershipOrders(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	status := strings.TrimSpace(c.Query("status"))
	items, total, err := h.service.ListMembershipOrders(userID, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) TrackExperimentEvent(c *gin.Context) {
	var req dto.TrackExperimentEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}

	item := model.ExperimentEvent{
		ExperimentKey: strings.TrimSpace(req.ExperimentKey),
		VariantKey:    strings.TrimSpace(req.VariantKey),
		EventType:     strings.ToUpper(strings.TrimSpace(req.EventType)),
		PageKey:       strings.TrimSpace(req.PageKey),
		TargetKey:     strings.TrimSpace(req.TargetKey),
		UserStage:     strings.ToUpper(strings.TrimSpace(req.UserStage)),
		AnonymousID:   strings.TrimSpace(req.AnonymousID),
		SessionID:     strings.TrimSpace(req.SessionID),
		Pathname:      strings.TrimSpace(req.Pathname),
		Referrer:      strings.TrimSpace(req.Referrer),
		Metadata:      req.Metadata,
	}
	if item.UserStage == "" {
		item.UserStage = "UNKNOWN"
	}
	if err := h.service.TrackExperimentEvent(item); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"accepted": true}))
}

func (h *UserGrowthHandler) GetAttachmentSignedURL(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	attachmentID := c.Param("id")
	info, err := h.service.GetAttachmentFileInfo(userID, attachmentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40402, Message: "attachment not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if strings.TrimSpace(h.cfg.AttachmentSigningSecret) == "" {
		c.JSON(http.StatusOK, dto.OK(model.SignedURL{SignedURL: info.FileURL, ExpiredAt: ""}))
		return
	}
	ttl := time.Duration(h.cfg.AttachmentSigningTTLSecond) * time.Second
	expiresAt := time.Now().Add(ttl)
	token := h.signAttachmentToken(attachmentID, userID, expiresAt)
	signedURL := h.buildSignedDownloadURL(attachmentID, token)
	c.JSON(http.StatusOK, dto.OK(model.SignedURL{SignedURL: signedURL, ExpiredAt: expiresAt.Format(time.RFC3339)}))
}

func (h *UserGrowthHandler) DownloadAttachment(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40101, Message: "missing token", Data: struct{}{}})
		return
	}
	if strings.TrimSpace(h.cfg.AttachmentSigningSecret) == "" {
		c.JSON(http.StatusForbidden, dto.APIResponse{Code: 40302, Message: "attachment signing disabled", Data: struct{}{}})
		return
	}
	attachmentID := c.Param("id")
	tokenAttachmentID, userID, _, ok := h.verifyAttachmentToken(token)
	if !ok || tokenAttachmentID != attachmentID {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40103, Message: "invalid token", Data: struct{}{}})
		return
	}
	info, err := h.service.GetAttachmentFileInfo(userID, attachmentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40402, Message: "attachment not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.LogAttachmentDownload(userID, attachmentID, info.ArticleID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Redirect(http.StatusFound, info.FileURL)
}

func (h *UserGrowthHandler) ListNewsCategories(c *gin.Context) {
	userID := optionalUserID(c)
	items, err := h.service.ListNewsCategories(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *UserGrowthHandler) ListNewsArticles(c *gin.Context) {
	userID := optionalUserID(c)
	page, pageSize := parsePage(c)
	categoryID := c.Query("category_id")
	keyword := c.Query("keyword")

	items, total, err := h.service.ListNewsArticles(userID, categoryID, keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) SearchGlobal(c *gin.Context) {
	userID := optionalUserID(c)
	keyword := normalizeSearchKeyword(c.Query("keyword"))
	if keyword == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "keyword required", Data: struct{}{}})
		return
	}
	mode := normalizeSearchMode(c.Query("mode"))
	defaultLimit := 6
	if mode == "full" {
		defaultLimit = 20
	}
	limit := parseIntOrDefault(c.Query("limit"), defaultLimit)
	if limit < 1 {
		limit = 1
	}
	if limit > 20 {
		limit = 20
	}

	newsItems, newsTotal, err := h.service.ListNewsArticles(userID, "", keyword, 1, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	stocks := make([]model.StockRecommendation, 0)
	strategies := make([]model.FuturesStrategy, 0)
	stockTotal := 0
	strategyTotal := 0

	if userID != "" {
		stockItems, _, stockErr := h.service.ListStockRecommendations(userID, "", 1, 200)
		if stockErr != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: stockErr.Error(), Data: struct{}{}})
			return
		}
		stocks, stockTotal = filterStockRecommendationsByKeyword(stockItems, keyword, limit)

		strategyItems, _, strategyErr := h.service.ListFuturesStrategies(userID, "", "", 1, 200)
		if strategyErr != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: strategyErr.Error(), Data: struct{}{}})
			return
		}
		strategies, strategyTotal = filterFuturesStrategiesByKeyword(strategyItems, keyword, limit)
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{
		"keyword": keyword,
		"scope":   map[bool]string{true: "USER", false: "PUBLIC"}[userID != ""],
		"stocks": gin.H{
			"items": stocks,
			"total": stockTotal,
		},
		"strategies": gin.H{
			"items": strategies,
			"total": strategyTotal,
		},
		"news": gin.H{
			"items": newsItems,
			"total": newsTotal,
		},
	}))
}

func (h *UserGrowthHandler) GetNewsArticleDetail(c *gin.Context) {
	userID := optionalUserID(c)
	articleID := c.Param("id")
	item, err := h.service.GetNewsArticleDetail(userID, articleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "article not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) ListNewsAttachments(c *gin.Context) {
	userID := optionalUserID(c)
	articleID := c.Param("id")
	items, err := h.service.ListNewsAttachments(userID, articleID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "article not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *UserGrowthHandler) GetRewardWallet(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	item, err := h.service.GetRewardWallet(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) ListRewardWalletTxns(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListRewardWalletTxns(userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) CreateWithdrawRequest(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req dto.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	withdrawID, err := h.service.CreateWithdrawRequest(userID, req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"withdraw_id": withdrawID}))
}

func (h *UserGrowthHandler) HandleYolkPayCallback(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusOK, "fail")
		return
	}
	params := make(map[string]string, len(c.Request.Form))
	for key, values := range c.Request.Form {
		if len(values) == 0 {
			continue
		}
		params[key] = strings.TrimSpace(values[0])
	}
	sign := strings.TrimSpace(params["sign"])
	orderNo := strings.TrimSpace(params["out_trade_no"])
	tradeNo := strings.TrimSpace(params["trade_no"])
	tradeStatus := strings.ToUpper(strings.TrimSpace(params["trade_status"]))

	cfg, err := h.resolveYolkPayConfig()
	if err != nil || !cfg.Enabled || strings.TrimSpace(cfg.Key) == "" {
		c.String(http.StatusOK, "fail")
		return
	}
	if sign == "" || !verifyYolkPaySign(params, sign, cfg.Key) {
		c.String(http.StatusOK, "fail")
		return
	}
	if orderNo == "" {
		c.String(http.StatusOK, "fail")
		return
	}
	if tradeStatus != "TRADE_SUCCESS" {
		c.String(http.StatusOK, "success")
		return
	}
	idempotencyKey := "yolkpay:" + orderNo
	if tradeNo != "" {
		idempotencyKey = "yolkpay:" + tradeNo
	}
	err = h.service.HandlePaymentCallback("YOLKPAY", orderNo, tradeNo, idempotencyKey, sign, true)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			c.String(http.StatusOK, "success")
			return
		}
		c.String(http.StatusOK, "fail")
		return
	}
	c.String(http.StatusOK, "success")
}

func (h *UserGrowthHandler) HandlePaymentCallback(c *gin.Context) {
	channel := c.Param("channel")
	var req dto.PaymentCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	signVerified, err := h.verifyPaymentSignature(channel, req.OrderNo, req.ChannelTxnNo, req.IdempotencyKey, req.Sign)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: err.Error(), Data: struct{}{}})
		return
	}
	requireSign := strings.TrimSpace(h.cfg.PaymentSigningSecret) != ""
	if !requireSign {
		if dbSecret, lookupErr := h.lookupSystemConfigValue(paymentSigningSecretConfigKey); lookupErr == nil && strings.TrimSpace(dbSecret) != "" {
			requireSign = true
		}
	}
	if !signVerified && requireSign {
		c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40103, Message: "invalid signature", Data: struct{}{}})
		return
	}
	err = h.service.HandlePaymentCallback(channel, req.OrderNo, req.ChannelTxnNo, req.IdempotencyKey, req.Sign, signVerified)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			c.JSON(http.StatusConflict, dto.APIResponse{Code: 40901, Message: err.Error(), Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *UserGrowthHandler) ListFuturesArbitrage(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	typeFilter := c.Query("type")
	items, total, err := h.service.ListFuturesArbitrage(typeFilter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) GetFuturesArbitrageDetail(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	id := c.Param("id")
	item, err := h.service.GetFuturesArbitrageDetail(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "arbitrage not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) ListArbitrageOpportunities(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	typeFilter := c.Query("type")

	items, total, err := h.service.ListArbitrageOpportunities(typeFilter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) CreateFuturesAlert(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	var req dto.FuturesAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.CreateFuturesAlert(userID, req.Contract, req.AlertType, req.Threshold)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *UserGrowthHandler) ListFuturesReviews(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	items, total, err := h.service.ListFuturesReviews(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) GetFuturesGuidance(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	contract := c.Param("contract")
	item, err := h.service.GetFuturesGuidance(contract)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) ListMarketEvents(c *gin.Context) {
	_, ok := requireUserID(c)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	eventType := strings.TrimSpace(c.Query("type"))
	if eventType == "" {
		eventType = strings.TrimSpace(c.Query("event_type"))
	}
	items, total, err := h.service.ListMarketEvents(eventType, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) GetMarketEventDetail(c *gin.Context) {
	_, ok := requireUserID(c)
	if !ok {
		return
	}
	id := c.Param("id")
	item, err := h.service.GetMarketEventDetail(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "event not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) ListStockRecommendations(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	tradeDate := c.Query("trade_date")
	items, total, err := h.service.ListStockRecommendations(userID, tradeDate, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) ListStockRecommendationHistory(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	outcome := strings.TrimSpace(c.Query("outcome"))
	tradeDateFrom := strings.TrimSpace(c.Query("trade_date_from"))
	tradeDateTo := strings.TrimSpace(c.Query("trade_date_to"))
	items, summary, total, err := h.service.ListStockRecommendationHistory(userID, outcome, tradeDateFrom, tradeDateTo, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "summary": summary, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) GetStockSimulatedOverview(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	data, err := h.service.AdminGetStockSimulatedOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(data))
}

func (h *UserGrowthHandler) ListStockSimulatedPositions(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	status := c.Query("status")
	symbol := c.Query("symbol")
	items, total, err := h.service.AdminListStockSimulatedPositions(status, symbol, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) GetStockRecommendationDetail(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	id := c.Param("id")
	item, err := h.service.GetStockRecommendationDetail(userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40403, Message: "stock recommendation not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) GetStockRecommendationPerformance(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	id := c.Param("id")
	points, err := h.service.GetStockRecommendationPerformance(userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40403, Message: "stock recommendation not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"points": points}))
}

func (h *UserGrowthHandler) GetStockRecommendationInsight(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	id := c.Param("id")
	item, err := h.service.GetStockRecommendationInsight(userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40403, Message: "stock recommendation not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	// 动态合并最新的 L3 推演状态
	if latestRun, err := h.service.GetLatestStrategyForecastL3Run(model.StrategyForecastL3TargetTypeStock, id); err == nil && latestRun.ID != "" {
		if item.Explanation.DeepForecastSummary == nil || latestRun.CreatedAt > item.Explanation.DeepForecastSummary.GeneratedAt {
			summary := latestRun.Summary
			if summary.RunID == "" {
				summary.RunID = latestRun.ID
			}
			summary.Status = latestRun.Status
			summary.GeneratedAt = latestRun.CreatedAt
			item.Explanation.DeepForecastSummary = &summary
		}
	}

	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) GetStockRecommendationVersionHistory(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	id := c.Param("id")
	items, err := h.service.GetStockRecommendationVersionHistory(userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40403, Message: "stock recommendation not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *UserGrowthHandler) ListPublicHoldings(c *gin.Context) {
	page, pageSize := parsePage(c)
	symbol := strings.TrimSpace(c.Query("symbol"))
	items, total, err := h.service.ListPublicHoldings(symbol, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) ListPublicFuturesPositions(c *gin.Context) {
	page, pageSize := parsePage(c)
	contract := strings.TrimSpace(c.Query("contract"))
	items, total, err := h.service.ListPublicFuturesPositions(contract, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) ListFuturesStrategies(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	page, pageSize := parsePage(c)
	contract := c.Query("contract")
	status := c.Query("status")
	items, total, err := h.service.ListFuturesStrategies(userID, contract, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *UserGrowthHandler) GetFuturesStrategyDetail(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	id := c.Param("id")
	item, err := h.service.GetFuturesStrategyDetail(userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40404, Message: "futures strategy not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) GetFuturesStrategyInsight(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	id := c.Param("id")
	item, err := h.service.GetFuturesStrategyInsight(userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40404, Message: "futures strategy not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *UserGrowthHandler) GetFuturesStrategyVersionHistory(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	_, ok = h.loadAccessProfile(c, userID)
	if !ok {
		return
	}
	id := c.Param("id")
	items, err := h.service.GetFuturesStrategyVersionHistory(userID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40404, Message: "futures strategy not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}



func normalizeSearchKeyword(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	return strings.ToLower(strings.Join(strings.Fields(text), " "))
}

func normalizeSearchMode(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "full":
		return "full"
	default:
		return "suggest"
	}
}

func containsSearchKeyword(keyword string, fields ...string) bool {
	if keyword == "" {
		return true
	}
	for _, field := range fields {
		if strings.Contains(strings.ToLower(strings.TrimSpace(field)), keyword) {
			return true
		}
	}
	return false
}

func normalizeStockSearchKey(symbol string, name string) string {
	normalizedSymbol := strings.ToUpper(strings.TrimSpace(symbol))
	normalizedName := strings.ToUpper(strings.TrimSpace(name))
	if normalizedSymbol != "" {
		for _, suffix := range []string{".SH", ".SZ", ".BJ", ".HK"} {
			normalizedSymbol = strings.TrimSuffix(normalizedSymbol, suffix)
		}
		if normalizedSymbol != "" {
			return normalizedSymbol
		}
	}
	return normalizedName
}

func filterStockRecommendationsByKeyword(items []model.StockRecommendation, keyword string, limit int) ([]model.StockRecommendation, int) {
	needle := normalizeSearchKeyword(keyword)
	if len(items) == 0 || needle == "" || limit <= 0 {
		return []model.StockRecommendation{}, 0
	}
	result := make([]model.StockRecommendation, 0, minInt(limit, len(items)))
	seen := make(map[string]struct{}, len(items))
	total := 0
	for _, item := range items {
		if !containsSearchKeyword(needle, item.Symbol, item.Name, item.ReasonSummary) {
			continue
		}
		stockKey := normalizeStockSearchKey(item.Symbol, item.Name)
		if stockKey != "" {
			if _, exists := seen[stockKey]; exists {
				continue
			}
			seen[stockKey] = struct{}{}
		}
		total++
		if len(result) < limit {
			result = append(result, item)
		}
	}
	return result, total
}

func filterFuturesStrategiesByKeyword(items []model.FuturesStrategy, keyword string, limit int) ([]model.FuturesStrategy, int) {
	needle := normalizeSearchKeyword(keyword)
	if len(items) == 0 || needle == "" || limit <= 0 {
		return []model.FuturesStrategy{}, 0
	}
	result := make([]model.FuturesStrategy, 0, minInt(limit, len(items)))
	total := 0
	for _, item := range items {
		if !containsSearchKeyword(needle, item.Contract, item.Name, item.Direction, item.ReasonSummary) {
			continue
		}
		total++
		if len(result) < limit {
			result = append(result, item)
		}
	}
	return result, total
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func optionalUserID(c *gin.Context) string {
	if v, ok := c.Get("user_id"); ok {
		if uid, castOK := v.(string); castOK && strings.TrimSpace(uid) != "" {
			return uid
		}
	}
	return ""
}

func requireUserID(c *gin.Context) (string, bool) {
	if v, ok := c.Get("user_id"); ok {
		if uid, castOK := v.(string); castOK && strings.TrimSpace(uid) != "" {
			return uid, true
		}
	}
	c.JSON(http.StatusUnauthorized, dto.APIResponse{Code: 40103, Message: "invalid token", Data: struct{}{}})
	return "", false
}

func (h *UserGrowthHandler) loadAccessProfile(c *gin.Context, userID string) (model.UserAccessProfile, bool) {
	profile, err := h.service.GetUserAccessProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return model.UserAccessProfile{}, false
	}
	return profile, true
}

func maskStockRecommendations(items []model.StockRecommendation) {
	for i := range items {
		items[i].PositionRange = ""
		items[i].ReasonSummary = ""
	}
}

func maskFuturesStrategies(items []model.FuturesStrategy) {
	for i := range items {
		items[i].PositionRange = ""
		items[i].ReasonSummary = ""
	}
}

func (h *UserGrowthHandler) signAttachmentToken(attachmentID string, userID string, expiresAt time.Time) string {
	payload := fmt.Sprintf("%s|%s|%d", attachmentID, userID, expiresAt.Unix())
	mac := hmac.New(sha256.New, []byte(h.cfg.AttachmentSigningSecret))
	_, _ = mac.Write([]byte(payload))
	signature := hex.EncodeToString(mac.Sum(nil))
	encodedPayload := base64.RawURLEncoding.EncodeToString([]byte(payload))
	encodedSig := base64.RawURLEncoding.EncodeToString([]byte(signature))
	return encodedPayload + "." + encodedSig
}

func (h *UserGrowthHandler) verifyAttachmentToken(token string) (attachmentID string, userID string, expiresAt time.Time, ok bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", "", time.Time{}, false
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", "", time.Time{}, false
	}
	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", "", time.Time{}, false
	}
	payload := string(payloadBytes)
	fields := strings.Split(payload, "|")
	if len(fields) != 3 {
		return "", "", time.Time{}, false
	}
	expUnix, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return "", "", time.Time{}, false
	}
	mac := hmac.New(sha256.New, []byte(h.cfg.AttachmentSigningSecret))
	_, _ = mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal(sigBytes, []byte(expected)) {
		return "", "", time.Time{}, false
	}
	exp := time.Unix(expUnix, 0)
	if time.Now().After(exp) {
		return "", "", time.Time{}, false
	}
	return fields[0], fields[1], exp, true
}

func (h *UserGrowthHandler) loadSystemConfigMap(keyword string) (map[string]string, error) {
	items, _, err := h.service.AdminListSystemConfigs(keyword, 1, 400)
	if err != nil {
		return nil, err
	}
	configMap := make(map[string]string, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		if key == "" {
			continue
		}
		configMap[key] = strings.TrimSpace(item.ConfigValue)
	}
	return configMap, nil
}

func (h *UserGrowthHandler) lookupSystemConfigValue(configKey string) (string, error) {
	items, _, err := h.service.AdminListSystemConfigs(configKey, 1, 10)
	if err != nil {
		return "", err
	}
	target := strings.ToLower(strings.TrimSpace(configKey))
	for _, item := range items {
		if strings.ToLower(strings.TrimSpace(item.ConfigKey)) == target {
			return strings.TrimSpace(item.ConfigValue), nil
		}
	}
	return "", nil
}

func (h *UserGrowthHandler) resolveYolkPayConfig() (yolkPayRuntimeConfig, error) {
	cfg, err := resolveYolkPayConfig(h.service)
	if err != nil {
		return cfg, err
	}
	baseURL := strings.TrimRight(strings.TrimSpace(h.cfg.PublicBaseURL), "/")
	if cfg.NotifyURL == "" && baseURL != "" {
		cfg.NotifyURL = baseURL + "/api/v1/payment/callbacks/yolkpay/notify"
	}
	if cfg.ReturnURL == "" && baseURL != "" {
		cfg.ReturnURL = baseURL + "/payment/success"
	}
	return cfg, nil
}

func (h *UserGrowthHandler) verifyPaymentSignature(channel string, orderNo string, channelTxnNo string, idempotencyKey string, sign string) (bool, error) {
	secret := strings.TrimSpace(h.cfg.PaymentSigningSecret)
	if dbSecret, err := h.lookupSystemConfigValue(paymentSigningSecretConfigKey); err == nil && strings.TrimSpace(dbSecret) != "" {
		secret = strings.TrimSpace(dbSecret)
	}
	if secret == "" {
		if h.cfg.AppEnv == "production" {
			return false, errors.New("payment signing secret not configured")
		}
		return false, nil
	}
	signPayload := fmt.Sprintf("channel=%s&order_no=%s&channel_txn_no=%s&idempotency_key=%s",
		strings.ToUpper(strings.TrimSpace(channel)),
		orderNo,
		channelTxnNo,
		idempotencyKey,
	)
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(signPayload))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(strings.ToLower(sign)), []byte(strings.ToLower(expected))), nil
}

func (h *UserGrowthHandler) buildSignedDownloadURL(attachmentID string, token string) string {
	base := strings.TrimRight(h.cfg.PublicBaseURL, "/")
	return fmt.Sprintf("%s/api/v1/news/attachments/%s/download?token=%s", base, attachmentID, url.QueryEscape(token))
}

func (h *UserGrowthHandler) AddUserVirtualSandbox(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	var req struct {
		RecoID   string  `json:"reco_id" binding:"required"`
		AddPrice float64 `json:"add_price" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AddUserVirtualSandbox(userID, req.RecoID, req.AddPrice); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *UserGrowthHandler) GetUserVirtualSandbox(c *gin.Context) {
	userID, ok := requireUserID(c)
	if !ok {
		return
	}
	items, err := h.service.GetUserVirtualSandbox(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

var klineDB *sql.DB
var klineDBOnce sync.Once

func (h *UserGrowthHandler) getKlineDB() *sql.DB {
	klineDBOnce.Do(func() {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FShanghai",
			h.cfg.MySQLUser, h.cfg.MySQLPass, h.cfg.MySQLHost, h.cfg.MySQLPort, h.cfg.MySQLDB))
		if err == nil {
			db.SetMaxOpenConns(4)
			db.SetMaxIdleConns(2)
			klineDB = db
		}
	})
	return klineDB
}

func (h *UserGrowthHandler) GetStockKline(c *gin.Context) {
	symbol := strings.TrimSpace(c.Query("symbol"))
	if symbol == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "symbol is required", Data: struct{}{}})
		return
	}
	daysStr := c.DefaultQuery("days", "60")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 10 || days > 365 {
		days = 60
	}

	type klinePoint struct {
		Date   string  `json:"date"`
		Open   float64 `json:"open"`
		High   float64 `json:"high"`
		Low    float64 `json:"low"`
		Close  float64 `json:"close"`
		Volume int64   `json:"volume"`
		Ma5    float64 `json:"ma5"`
		Ma10   float64 `json:"ma10"`
		Ma20   float64 `json:"ma20"`
	}

	// Try reading real data from market_daily_bars (reused connection pool)
	db := h.getKlineDB()
	if db != nil {
		rows, qErr := db.Query("SELECT trade_date, open_price, high_price, low_price, close_price, volume FROM market_daily_bars WHERE asset_class = 'STOCK' AND instrument_key = ? AND source_key = 'TUSHARE' ORDER BY trade_date DESC LIMIT ?", symbol, days)
		if qErr == nil {
			seen := map[string]bool{}
			var pts []klinePoint
			for rows.Next() {
				var p klinePoint
				var t time.Time
				if err := rows.Scan(&t, &p.Open, &p.High, &p.Low, &p.Close, &p.Volume); err != nil {
					continue
				}
				dateStr := t.Format("2006-01-02")
				if seen[dateStr] {
					continue
				}
				seen[dateStr] = true
				p.Date = dateStr
				p.Open = float64(int(p.Open*100+0.5)) / 100
				p.High = float64(int(p.High*100+0.5)) / 100
				p.Low = float64(int(p.Low*100+0.5)) / 100
				p.Close = float64(int(p.Close*100+0.5)) / 100
				pts = append(pts, p)
			}
			rows.Close()
			if len(pts) > 0 {
				for i, j := 0, len(pts)-1; i < j; i, j = i+1, j-1 {
					pts[i], pts[j] = pts[j], pts[i]
				}
				n := len(pts)
				for i := 0; i < n; i++ {
					pts[i].Ma5 = pts[i].Close
					pts[i].Ma10 = pts[i].Close
					pts[i].Ma20 = pts[i].Close
					if i >= 4 {
						s := 0.0
						for j := i - 4; j <= i; j++ { s += pts[j].Close }
						pts[i].Ma5 = float64(int(s/5*100+0.5)) / 100
					}
					if i >= 9 {
						s := 0.0
						for j := i - 9; j <= i; j++ { s += pts[j].Close }
						pts[i].Ma10 = float64(int(s/10*100+0.5)) / 100
					}
					if i >= 19 {
						s := 0.0
						for j := i - 19; j <= i; j++ { s += pts[j].Close }
						pts[i].Ma20 = float64(int(s/20*100+0.5)) / 100
					}
			}
			c.JSON(http.StatusOK, dto.OK(gin.H{"symbol": symbol, "days": len(pts), "points": pts}))
			return
		}
	}
}
	// Fallback: generate synthetic data
	seed := int64(0)
	for _, ch := range symbol {
		seed += int64(ch)
	}
	rng := rand.New(rand.NewSource(seed))
	basePrice := 80.0 + rng.Float64()*200
	now := time.Now()

	points := make([]klinePoint, days)
	closes := make([]float64, days)
	price := basePrice

	for i := 0; i < days; i++ {
		change := (rng.Float64() - 0.48) * 6.0
		if i > 0 && closes[i-1] > basePrice*1.25 {
			change -= 0.5
		}
		if i > 0 && closes[i-1] < basePrice*0.85 {
			change += 0.5
		}
		open := price
		close := price + change
		if close < 0 {
			close = 0.5
		}
		high := open
		if close > high {
			high = close
		}
		high += rng.Float64() * 2.5
		low := open
		if close < low {
			low = close
		}
		low -= rng.Float64() * 2.5
		if low < 0 {
			low = 0
		}
		volume := int64(100000 + rng.Float64()*5000000)
		date := now.AddDate(0, 0, i-days+1)

		points[i] = klinePoint{
			Date:   date.Format("2006-01-02"),
			Open:   float64(int(open*100+0.5)) / 100,
			High:   float64(int(high*100+0.5)) / 100,
			Low:    float64(int(low*100+0.5)) / 100,
			Close:  float64(int(close*100+0.5)) / 100,
			Volume: volume,
		}
		closes[i] = close
		price = close
	}
	for i := 0; i < days; i++ {
		if i >= 4 {
			s := 0.0
			for j := i - 4; j <= i; j++ {
				s += closes[j]
			}
			points[i].Ma5 = float64(int(s/5*100+0.5)) / 100
		} else {
			points[i].Ma5 = points[i].Close
		}
		if i >= 9 {
			s := 0.0
			for j := i - 9; j <= i; j++ {
				s += closes[j]
			}
			points[i].Ma10 = float64(int(s/10*100+0.5)) / 100
		} else {
			points[i].Ma10 = points[i].Close
		}
		if i >= 19 {
			s := 0.0
			for j := i - 19; j <= i; j++ {
				s += closes[j]
			}
			points[i].Ma20 = float64(int(s/20*100+0.5)) / 100
		} else {
			points[i].Ma20 = points[i].Close
		}
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{
		"symbol": symbol,
		"days":   days,
		"points": points,
	}))
}

type matchResult struct {
	Rank       int       `json:"rank"`
	Stock      string    `json:"stock"`
	Similarity float64   `json:"similarity"`
	MatchDate  string    `json:"match_date"`
	Next7d     []float64 `json:"next_7d"`
	Return7d   float64   `json:"return_7d"`
}

type predictionOut struct {

	Median    []float64 `json:"median"`
	Upper75   []float64 `json:"upper_75"`
	Lower25   []float64 `json:"lower_25"`
	AvgReturn float64   `json:"avg_return"`
	WinRate   float64   `json:"win_rate"`
}


type candidate struct {
	stock      string
	date       string
	similarity float64
	next       []float64
}
func (h *UserGrowthHandler) PatternMatch(c *gin.Context) {
	symbol := strings.TrimSpace(c.Query("symbol"))
	lookback := 20
	if v := c.Query("lookback"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && (n == 20 || n == 60) {
			lookback = n
		}
	}
	topN := 10
	if v := c.Query("top"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 20 {
			topN = n
		}
	}
	if symbol == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "symbol is required", Data: struct{}{}})
		return
	}

	db := h.getKlineDB()
	if db == nil {
		c.JSON(http.StatusServiceUnavailable, dto.APIResponse{Code: 50301, Message: "database unavailable", Data: struct{}{}})
		return
	}

	// 1. Get the current stock's price segment
	rows, err := db.Query("SELECT close_price, trade_date FROM market_daily_bars WHERE asset_class='STOCK' AND instrument_key=? AND source_key='TUSHARE' ORDER BY trade_date DESC LIMIT ?", symbol, lookback+7)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: "query failed", Data: struct{}{}})
		return
	}
	var prices []float64
	var dates []string
	var currentSeg []float64
	for rows.Next() {
		var p float64
		var d string
		if err := rows.Scan(&p, &d); err != nil {
			continue
		}
		prices = append([]float64{p}, prices...)
		dates = append([]string{d}, dates...)
	}
	rows.Close()
	if len(prices) < lookback {
		c.JSON(http.StatusOK, dto.OK(gin.H{"symbol": symbol, "error": "not enough data", "matches": []matchResult{}, "prediction": nil}))
		return
	}
	currentSeg = prices[len(prices)-lookback:]
	// Normalize current segment: base=100
	base := currentSeg[0]
	normCur := make([]float64, len(currentSeg))
	for i, v := range currentSeg {
		normCur[i] = v / base * 100
	}

	// 2. Query candidate stocks (top 300 with enough data)
	candidateStocks := []string{}
	stockRows, err := db.Query("SELECT instrument_key FROM market_daily_bars WHERE asset_class='STOCK' AND source_key='TUSHARE' AND instrument_key != ? GROUP BY instrument_key HAVING COUNT(*) >= ? ORDER BY COUNT(*) DESC LIMIT 300", symbol, lookback+7)
	if err == nil {
		for stockRows.Next() {
			var s string
			stockRows.Scan(&s)
			candidateStocks = append(candidateStocks, s)
		}
		stockRows.Close()
	}

	var candidates []candidate
	for _, cs := range candidateStocks {
		r, e := db.Query("SELECT close_price, trade_date FROM market_daily_bars WHERE asset_class='STOCK' AND instrument_key=? AND source_key='TUSHARE' ORDER BY trade_date ASC", cs)
		if e != nil {
			continue
		}
		var sp []float64
		var sd []string
		for r.Next() {
			var p float64
			var t time.Time
			if err := r.Scan(&p, &t); err == nil {
				sp = append(sp, p)
				sd = append(sd, t.Format("2006-01-02"))
			} else {
				fmt.Printf("[DEBUG] cs=%s Scan error: %v\n", cs, err)
			}
		}
		r.Close()
		if len(sp) < lookback+1 {
			continue
		}
		baseSeg := currentSeg[0]
		for i := 0; i <= len(sp)-lookback-7; i++ {
			sb := sp[i]
			dot, n1, n2 := 0.0, 0.0, 0.0
			for j := 0; j < lookback; j++ {
				nv := sp[i+j] / sb * 100
				cv := currentSeg[j] / baseSeg * 100
				dot += cv * nv
				n1 += cv * cv
				n2 += nv * nv
			}
			sim := dot / (sqrt(n1*n2) + 1e-10)
			if sim > 0.85 {
				next7 := make([]float64, 7)
				for j := 0; j < 7 && i+lookback+j < len(sp); j++ {
					next7[j] = (sp[i+lookback+j]/sp[i+lookback-1] - 1) * 100
				}
				matchDate := ""
				if i+lookback-1 < len(sd) {
					matchDate = sd[i+lookback-1]
				}
				candidates = append(candidates, candidate{stock: cs, date: matchDate, similarity: sim, next: next7})
				if len(candidates) > 100 {
					sortCandidates(candidates)
					candidates = candidates[:100]
				}
			}
		}
	}

	// 4. Sort candidates by similarity, take topN
	// 4. Sort candidates by similarity, take topN
	sortCandidates(candidates)

	matches := make([]matchResult, 0, topN)
	var allNext7d [][]float64
	for i := 0; i < len(candidates) && i < topN; i++ {
		c := candidates[i]
		ret := 0.0
		if len(c.next) > 0 {
			ret = c.next[len(c.next)-1]
		}
		matches = append(matches, matchResult{
			Rank: i + 1, Stock: c.stock,
			Similarity: float64(int(c.similarity*1000)) / 1000,
			MatchDate:  c.date, Next7d: c.next, Return7d: ret,
		})
		allNext7d = append(allNext7d, c.next)
	}

	// 5. Fetch prediction from Python strategy-engine, fallback to local morph prediction if offline
	strategyEngineURL := strings.TrimRight(h.cfg.StrategyEngineBaseURL, "/")
	if strategyEngineURL == "" {
		strategyEngineURL = "http://127.0.0.1:18081"
	}

	var pred *predictionOut
	targetTradeDate := dates[len(dates)-1]
	if len(targetTradeDate) > 10 {
		targetTradeDate = targetTradeDate[:10]
	}
	reqURL := fmt.Sprintf("%s/internal/v1/predict/stock-7d?symbol=%s&trade_date=%s", strategyEngineURL, url.QueryEscape(symbol), url.QueryEscape(targetTradeDate))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err == nil {
		resp, httpErr := http.DefaultClient.Do(httpReq)
		if httpErr == nil {
			defer resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var pyPred predictionOut
				if decErr := json.NewDecoder(resp.Body).Decode(&pyPred); decErr == nil {
					pred = &pyPred
				} else {
					fmt.Printf("[DEBUG] Failed to decode python prediction response: %v\n", decErr)
				}
			} else {
				fmt.Printf("[DEBUG] Python prediction returned status: %d\n", resp.StatusCode)
			}
		} else {
			fmt.Printf("[DEBUG] Failed to send request to python prediction engine: %v\n", httpErr)
		}
	} else {
		fmt.Printf("[DEBUG] Failed to build request to python prediction engine: %v\n", err)
	}

	if pred == nil {
		// fallback to local morphological aggregation
		pred = buildPrediction(allNext7d)
	}

	c.JSON(http.StatusOK, dto.OK(gin.H{
		"symbol":     symbol,
		"lookback":   lookback,
		"current":    normCur[len(normCur)-7:],
		"matches":    matches,
		"prediction": pred,
	}))
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	z := x / 2
	for i := 0; i < 20; i++ {
		z = z - (z*z-x)/(2*z)
	}
	return z
}

func sortCandidates(c []candidate) {
	for i := 0; i < len(c); i++ {
		for j := i + 1; j < len(c); j++ {
			if c[j].similarity > c[i].similarity {
				c[i], c[j] = c[j], c[i]
			}
		}
	}
}

func buildPrediction(allNext7d [][]float64) *predictionOut {
	if len(allNext7d) == 0 {
		return nil
	}
	n := 7
	median := make([]float64, n)
	u75 := make([]float64, n)
	l25 := make([]float64, n)
	totalReturn := 0.0
	positive := 0

	for day := 0; day < n; day++ {
		var vals []float64
		for _, seq := range allNext7d {
			if day < len(seq) {
				vals = append(vals, seq[day])
			}
		}
		if len(vals) == 0 {
			continue
		}
		// Sort vals
		sortFloats(vals)
		median[day] = vals[len(vals)/2]
		u75[day] = vals[len(vals)*3/4]
		l25[day] = vals[len(vals)/4]
	}
	for _, seq := range allNext7d {
		if len(seq) > 0 {
			totalReturn += seq[len(seq)-1]
			if seq[len(seq)-1] > 0 {
				positive++
			}
		}
	}
	nValid := len(allNext7d)
	return &predictionOut{
		Median: median, Upper75: u75, Lower25: l25,
		AvgReturn: float64(int(totalReturn/float64(nValid)*100)) / 100,
		WinRate:   float64(int(float64(positive)/float64(nValid)*10000)) / 100,
	}
}

func sortFloats(v []float64) {
	for i := 0; i < len(v); i++ {
		for j := i + 1; j < len(v); j++ {
			if v[j] < v[i] {
				v[i], v[j] = v[j], v[i]
			}
		}
	}
}
