package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
	"sercherai/backend/internal/growth/service"
	"sercherai/backend/internal/platform/config"
)

type AdminGrowthHandler struct {
	service service.GrowthService
	cfg     config.Config
}

func NewAdminGrowthHandler(service service.GrowthService, cfg config.Config) *AdminGrowthHandler {
	return &AdminGrowthHandler{service: service, cfg: cfg}
}

var allowedNewsAttachmentMimePrefixes = []string{
	"image/",
	"text/",
}

var allowedNewsAttachmentMIMEs = map[string]struct{}{
	"application/pdf":               {},
	"application/msword":            {},
	"application/vnd.ms-excel":      {},
	"application/vnd.ms-powerpoint": {},
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   {},
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         {},
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": {},
	"application/zip": {},
}

const stockDefaultSourceConfigKey = "stock.quotes.default_source_key"
const stockDefaultSourceFallback = "TUSHARE"
const schedulerJobDailyFuturesStrategy = "daily_futures_strategy"
const schedulerJobFuturesStrategyGenerate = "futures_strategy_generate"
const schedulerJobFuturesStrategyEvaluate = "futures_strategy_evaluate"
const schedulerAutoRetryEnabledConfigKey = "scheduler.auto_retry.enabled"
const schedulerAutoRetryMaxRetriesConfigKey = "scheduler.auto_retry.max_retries"
const schedulerAutoRetryBackoffSecondsConfigKey = "scheduler.auto_retry.backoff_seconds"
const schedulerAutoRetryJobsConfigKey = "scheduler.auto_retry.jobs"
const schedulerAutoRetryDefaultJob = "daily_stock_quant_pipeline"
const ossProviderConfigKey = "oss.provider"
const ossEnabledConfigKey = "oss.enabled"
const ossQiniuAccessKeyConfigKey = "oss.qiniu.access_key"
const ossQiniuSecretKeyConfigKey = "oss.qiniu.secret_key"
const ossQiniuBucketConfigKey = "oss.qiniu.bucket"
const ossQiniuDomainConfigKey = "oss.qiniu.domain"
const ossQiniuRegionConfigKey = "oss.qiniu.region"
const ossQiniuPathPrefixConfigKey = "oss.qiniu.path_prefix"
const ossQiniuUseHTTPSConfigKey = "oss.qiniu.use_https"
const ossUploadMaxSizeMBConfigKey = "oss.upload.max_size_mb"

type supportedSchedulerJob struct {
	JobName     string `json:"job_name"`
	DisplayName string `json:"display_name"`
	Module      string `json:"module"`
	AliasOf     string `json:"alias_of,omitempty"`
}

var supportedSchedulerJobs = []supportedSchedulerJob{
	{JobName: "daily_stock_quant_pipeline", DisplayName: "每日股票量化流水线", Module: "STOCK"},
	{JobName: "daily_stock_recommendation", DisplayName: "每日股票推荐", Module: "STOCK"},
	{JobName: schedulerJobDailyFuturesStrategy, DisplayName: "每日期货策略", Module: "FUTURES"},
	{JobName: schedulerJobFuturesStrategyGenerate, DisplayName: "期货策略生成(别名)", Module: "FUTURES", AliasOf: schedulerJobDailyFuturesStrategy},
	{JobName: schedulerJobFuturesStrategyEvaluate, DisplayName: "期货策略评估", Module: "FUTURES"},
	{JobName: "doc_fast_news_incremental", DisplayName: "DocFast资讯增量同步", Module: "NEWS"},
	{JobName: "tushare_news_incremental", DisplayName: "Tushare资讯增量同步", Module: "NEWS"},
	{JobName: "vip_membership_lifecycle", DisplayName: "VIP会员生命周期任务", Module: "SYSTEM"},
}

type ossUploadConfig struct {
	Provider    string
	Enabled     bool
	AccessKey   string
	SecretKey   string
	Bucket      string
	Domain      string
	Region      string
	PathPrefix  string
	UseHTTPS    bool
	MaxUploadMB int
}

func (h *AdminGrowthHandler) ListInviteRecords(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")

	items, total, err := h.service.AdminListInviteRecords(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ListRewardRecords(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")

	items, total, err := h.service.AdminListRewardRecords(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ReviewRewardRecord(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListReconciliation(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListReconciliation(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) RetryReconciliation(c *gin.Context) {
	batchID := c.Param("batch_id")
	if err := h.service.AdminRetryReconciliation(batchID); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListRiskRules(c *gin.Context) {
	items, err := h.service.AdminListRiskRules()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) CreateRiskRule(c *gin.Context) {
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

func (h *AdminGrowthHandler) UpdateRiskRule(c *gin.Context) {
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
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListRiskHits(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListRiskHits(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ReviewRiskHit(c *gin.Context) {
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
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListWithdrawRequests(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListWithdrawRequests(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ReviewWithdrawRequest(c *gin.Context) {
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
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListNewsCategories(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListNewsCategories(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateNewsCategory(c *gin.Context) {
	var req dto.NewsCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateNewsCategory(req.Name, req.Slug, req.Sort, req.Visibility, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "NEWS", "CREATE_CATEGORY", "NEWS_CATEGORY", id, "", req.Status, req.Slug)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateNewsCategory(c *gin.Context) {
	id := c.Param("id")
	var req dto.NewsCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateNewsCategory(id, req.Name, req.Slug, req.Sort, req.Visibility, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListNewsArticles(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	categoryID := c.Query("category_id")
	items, total, err := h.service.AdminListNewsArticles(status, categoryID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) GetNewsArticleDetail(c *gin.Context) {
	id := c.Param("id")
	item, err := h.service.AdminGetNewsArticleDetail(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "article not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	attachments, err := h.service.AdminListNewsAttachments(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"article":     item,
		"attachments": attachments,
	}))
}

func (h *AdminGrowthHandler) CreateNewsArticle(c *gin.Context) {
	var req dto.NewsArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operator, _ := c.Get("user_id")
	authorID, _ := operator.(string)
	if authorID == "" {
		authorID = "admin_unknown"
	}
	id, err := h.service.AdminCreateNewsArticle(req.CategoryID, req.Title, req.Summary, req.Content, req.CoverURL, req.Visibility, req.Status, authorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateNewsArticle(c *gin.Context) {
	id := c.Param("id")
	var req dto.NewsArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateNewsArticle(id, req.CategoryID, req.Title, req.Summary, req.Content, req.CoverURL, req.Visibility, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) PublishNewsArticle(c *gin.Context) {
	id := c.Param("id")
	var req dto.NewsPublishRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminPublishNewsArticle(id, req.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "article not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "NEWS", "PUBLISH_ARTICLE", "NEWS_ARTICLE", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) UploadNewsAttachment(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "missing file", Data: struct{}{}})
		return
	}

	originalName := sanitizeUploadFileName(fileHeader.Filename)
	if originalName == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "invalid file name", Data: struct{}{}})
		return
	}

	ossCfg := h.resolveOSSUploadConfig()
	maxUploadMB := h.cfg.AttachmentUploadMaxMB
	if ossCfg.MaxUploadMB > 0 {
		maxUploadMB = ossCfg.MaxUploadMB
	}
	if maxUploadMB <= 0 {
		maxUploadMB = 20
	}
	maxUploadBytes := int64(maxUploadMB) * 1024 * 1024
	if fileHeader.Size <= 0 {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "file is empty", Data: struct{}{}})
		return
	}
	if fileHeader.Size > maxUploadBytes {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Code:    40001,
			Message: fmt.Sprintf("file exceeds %dMB", maxUploadMB),
			Data:    struct{}{},
		})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: "open file failed", Data: struct{}{}})
		return
	}
	defer src.Close()

	header := make([]byte, 512)
	n, readErr := io.ReadFull(src, header)
	if readErr != nil && readErr != io.ErrUnexpectedEOF {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: "read file failed", Data: struct{}{}})
		return
	}
	mimeType := http.DetectContentType(header[:n])
	if !isAllowedNewsAttachmentMime(mimeType) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Code:    40001,
			Message: "unsupported file type",
			Data:    struct{}{},
		})
		return
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: "seek file failed", Data: struct{}{}})
		return
	}

	dateDir := time.Now().Format("20060102")
	ext := resolveUploadExt(originalName, mimeType)
	targetName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), randomHex(4), ext)
	payload, readAllErr := io.ReadAll(io.LimitReader(src, maxUploadBytes+1))
	if readAllErr != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: "read file payload failed", Data: struct{}{}})
		return
	}
	written := int64(len(payload))
	if written <= 0 || written > maxUploadBytes {
		c.JSON(http.StatusBadRequest, dto.APIResponse{
			Code:    40001,
			Message: fmt.Sprintf("file exceeds %dMB", maxUploadMB),
			Data:    struct{}{},
		})
		return
	}

	if ossCfg.Enabled && strings.EqualFold(strings.TrimSpace(ossCfg.Provider), "QINIU") {
		objectKey := joinObjectPath(ossCfg.PathPrefix, "news", dateDir, targetName)
		fileURL, uploadErr := h.uploadToQiniu(ossCfg, objectKey, payload, mimeType)
		if uploadErr != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: uploadErr.Error(), Data: struct{}{}})
			return
		}
		c.JSON(http.StatusOK, dto.OK(gin.H{
			"file_name": originalName,
			"file_url":  fileURL,
			"file_size": written,
			"mime_type": mimeType,
		}))
		return
	}

	uploadRoot := strings.TrimSpace(h.cfg.AttachmentUploadDir)
	if uploadRoot == "" {
		uploadRoot = "./uploads"
	}
	targetDir := filepath.Join(uploadRoot, "news", dateDir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: "create upload dir failed", Data: struct{}{}})
		return
	}
	targetPath := filepath.Join(targetDir, targetName)
	if err := os.WriteFile(targetPath, payload, 0o644); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: "save file failed", Data: struct{}{}})
		return
	}

	publicPath := fmt.Sprintf("/uploads/news/%s/%s", dateDir, targetName)
	fileURL := publicPath
	if baseURL := strings.TrimRight(strings.TrimSpace(h.cfg.PublicBaseURL), "/"); baseURL != "" {
		fileURL = baseURL + publicPath
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"file_name": originalName,
		"file_url":  fileURL,
		"file_size": written,
		"mime_type": mimeType,
	}))
}

func (h *AdminGrowthHandler) CreateNewsAttachment(c *gin.Context) {
	articleID := c.Param("id")
	var req dto.NewsAttachmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateNewsAttachment(articleID, req.FileName, req.FileURL, req.FileSize, req.MimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) ListNewsAttachments(c *gin.Context) {
	articleID := c.Param("id")
	items, err := h.service.AdminListNewsAttachments(articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminGrowthHandler) DeleteNewsAttachment(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.AdminDeleteNewsAttachment(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40402, Message: "attachment not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "NEWS", "DELETE_ATTACHMENT", "NEWS_ATTACHMENT", id, "", "DELETED", "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func isAllowedNewsAttachmentMime(mimeType string) bool {
	normalized := strings.TrimSpace(strings.ToLower(mimeType))
	if normalized == "" {
		return false
	}
	if _, ok := allowedNewsAttachmentMIMEs[normalized]; ok {
		return true
	}
	for _, prefix := range allowedNewsAttachmentMimePrefixes {
		if strings.HasPrefix(normalized, prefix) {
			return true
		}
	}
	return false
}

func resolveUploadExt(fileName string, mimeType string) string {
	ext := strings.ToLower(strings.TrimSpace(filepath.Ext(fileName)))
	if ext != "" && len(ext) <= 10 {
		return ext
	}
	if exts, err := mime.ExtensionsByType(mimeType); err == nil && len(exts) > 0 {
		return strings.ToLower(exts[0])
	}
	return ".bin"
}

func sanitizeUploadFileName(fileName string) string {
	name := strings.TrimSpace(filepath.Base(fileName))
	if name == "" || name == "." || name == ".." {
		return ""
	}
	return name
}

func randomHex(n int) string {
	if n <= 0 {
		n = 4
	}
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}

func (h *AdminGrowthHandler) resolveOSSUploadConfig() ossUploadConfig {
	cfg := ossUploadConfig{
		Provider:    "QINIU",
		Enabled:     false,
		Region:      "z0",
		PathPrefix:  "uploads/",
		UseHTTPS:    true,
		MaxUploadMB: h.cfg.AttachmentUploadMaxMB,
	}
	items, _, err := h.service.AdminListSystemConfigs("oss.", 1, 300)
	if err != nil {
		if cfg.MaxUploadMB <= 0 {
			cfg.MaxUploadMB = 20
		}
		return cfg
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case ossProviderConfigKey:
			if value != "" {
				cfg.Provider = strings.ToUpper(value)
			}
		case ossEnabledConfigKey:
			cfg.Enabled = parseConfigBool(value, cfg.Enabled)
		case ossQiniuAccessKeyConfigKey:
			cfg.AccessKey = value
		case ossQiniuSecretKeyConfigKey:
			cfg.SecretKey = value
		case ossQiniuBucketConfigKey:
			cfg.Bucket = value
		case ossQiniuDomainConfigKey:
			cfg.Domain = value
		case ossQiniuRegionConfigKey:
			if value != "" {
				cfg.Region = strings.ToLower(value)
			}
		case ossQiniuPathPrefixConfigKey:
			cfg.PathPrefix = value
		case ossQiniuUseHTTPSConfigKey:
			cfg.UseHTTPS = parseConfigBool(value, cfg.UseHTTPS)
		case ossUploadMaxSizeMBConfigKey:
			cfg.MaxUploadMB = parseConfigInt(value, cfg.MaxUploadMB)
		}
	}
	if cfg.MaxUploadMB <= 0 {
		cfg.MaxUploadMB = 20
	}
	if cfg.Provider == "" {
		cfg.Provider = "QINIU"
	}
	if cfg.Region == "" {
		cfg.Region = "z0"
	}
	if cfg.PathPrefix == "" {
		cfg.PathPrefix = "uploads/"
	}
	return cfg
}

func (h *AdminGrowthHandler) resolveYolkPayConfig() (yolkPayRuntimeConfig, error) {
	cfg := yolkPayRuntimeConfig{
		PaymentEnabled: false,
		Enabled:        false,
		Gateway:        "https://www.yolkpay.net",
		MAPIPath:       "/mapi.php",
		PayType:        "airpay",
		Device:         "pc",
	}
	items, _, err := h.service.AdminListSystemConfigs("payment.", 1, 400)
	if err != nil {
		return cfg, err
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case paymentEnabledConfigKey:
			cfg.PaymentEnabled = parseConfigBool(value, cfg.PaymentEnabled)
		case paymentChannelYolkPayEnabledConfigKey:
			cfg.Enabled = parseConfigBool(value, cfg.Enabled)
		case paymentChannelYolkPayPIDConfigKey:
			cfg.PID = value
		case paymentChannelYolkPayKeyConfigKey:
			cfg.Key = value
		case paymentChannelYolkPayGatewayConfigKey:
			if value != "" {
				cfg.Gateway = value
			}
		case paymentChannelYolkPayMAPIPathConfigKey:
			if value != "" {
				cfg.MAPIPath = value
			}
		case paymentChannelYolkPayNotifyURLConfigKey:
			cfg.NotifyURL = value
		case paymentChannelYolkPayReturnURLConfigKey:
			cfg.ReturnURL = value
		case paymentChannelYolkPayPayTypeConfigKey:
			if value != "" {
				cfg.PayType = strings.ToLower(value)
			}
		case paymentChannelYolkPayDeviceConfigKey:
			if value != "" {
				cfg.Device = strings.ToLower(value)
			}
		}
	}
	return cfg, nil
}

func joinObjectPath(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}
	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		p := strings.TrimSpace(part)
		p = strings.Trim(p, "/")
		if p == "" {
			continue
		}
		normalized = append(normalized, p)
	}
	return strings.Join(normalized, "/")
}

func buildQiniuUploadHost(region string, useHTTPS bool) string {
	regionHost := map[string]string{
		"z0":  "up-z0.qiniup.com",
		"z1":  "up-z1.qiniup.com",
		"z2":  "up-z2.qiniup.com",
		"na0": "up-na0.qiniup.com",
		"as0": "up-as0.qiniup.com",
	}
	host, ok := regionHost[strings.ToLower(strings.TrimSpace(region))]
	if !ok || host == "" {
		host = "up-z0.qiniup.com"
	}
	scheme := "https://"
	if !useHTTPS {
		scheme = "http://"
	}
	return scheme + host
}

func buildQiniuPublicURL(cfg ossUploadConfig, objectKey string) string {
	domain := strings.TrimSpace(cfg.Domain)
	if domain == "" {
		return objectKey
	}
	if strings.HasPrefix(strings.ToLower(domain), "http://") || strings.HasPrefix(strings.ToLower(domain), "https://") {
		return strings.TrimRight(domain, "/") + "/" + strings.TrimLeft(objectKey, "/")
	}
	scheme := "https://"
	if !cfg.UseHTTPS {
		scheme = "http://"
	}
	return scheme + strings.TrimRight(domain, "/") + "/" + strings.TrimLeft(objectKey, "/")
}

func buildQiniuUploadToken(accessKey string, secretKey string, bucket string, objectKey string) (string, error) {
	ak := strings.TrimSpace(accessKey)
	sk := strings.TrimSpace(secretKey)
	bk := strings.TrimSpace(bucket)
	key := strings.TrimSpace(objectKey)
	if ak == "" || sk == "" || bk == "" || key == "" {
		return "", errors.New("qiniu credential or object key missing")
	}
	policy := map[string]interface{}{
		"scope":    bk + ":" + key,
		"deadline": time.Now().Add(1 * time.Hour).Unix(),
	}
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return "", err
	}
	encodedPolicy := base64.URLEncoding.EncodeToString(policyJSON)
	mac := hmac.New(sha1.New, []byte(sk))
	if _, err := mac.Write([]byte(encodedPolicy)); err != nil {
		return "", err
	}
	sign := base64.URLEncoding.EncodeToString(mac.Sum(nil))
	return ak + ":" + sign + ":" + encodedPolicy, nil
}

func (h *AdminGrowthHandler) uploadToQiniu(cfg ossUploadConfig, objectKey string, payload []byte, mimeType string) (string, error) {
	if strings.TrimSpace(cfg.AccessKey) == "" || strings.TrimSpace(cfg.SecretKey) == "" ||
		strings.TrimSpace(cfg.Bucket) == "" || strings.TrimSpace(cfg.Domain) == "" {
		return "", errors.New("oss qiniu config incomplete, require access_key/secret_key/bucket/domain")
	}
	if len(payload) == 0 {
		return "", errors.New("empty upload payload")
	}
	token, err := buildQiniuUploadToken(cfg.AccessKey, cfg.SecretKey, cfg.Bucket, objectKey)
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if err := writer.WriteField("token", token); err != nil {
		return "", err
	}
	if err := writer.WriteField("key", objectKey); err != nil {
		return "", err
	}
	part, err := writer.CreateFormFile("file", filepath.Base(objectKey))
	if err != nil {
		return "", err
	}
	if _, err := part.Write(payload); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPost, buildQiniuUploadHost(cfg.Region, cfg.UseHTTPS), &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if strings.TrimSpace(mimeType) != "" {
		req.Header.Set("X-Qiniu-MimeType", mimeType)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(respBytes))
		if msg == "" {
			msg = resp.Status
		}
		return "", fmt.Errorf("qiniu upload failed: %s", msg)
	}
	if len(respBytes) > 0 {
		var result map[string]interface{}
		if json.Unmarshal(respBytes, &result) == nil {
			if errText, ok := result["error"].(string); ok && strings.TrimSpace(errText) != "" {
				return "", fmt.Errorf("qiniu upload failed: %s", strings.TrimSpace(errText))
			}
			if key, ok := result["key"].(string); ok && strings.TrimSpace(key) != "" {
				objectKey = strings.TrimSpace(key)
			}
		}
	}
	return buildQiniuPublicURL(cfg, objectKey), nil
}

func normalizeAdminDateTime(value string) (string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", errors.New("is required")
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if layout == time.RFC3339 {
			if ts, err := time.Parse(layout, trimmed); err == nil {
				return ts.Format(time.RFC3339), nil
			}
			continue
		}
		if ts, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return ts.Format(time.RFC3339), nil
		}
	}
	return "", errors.New("format invalid, expected RFC3339 or YYYY-MM-DD")
}

func normalizeStockSymbols(raw []string) []string {
	if len(raw) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(raw))
	result := make([]string, 0, len(raw))
	for _, item := range raw {
		symbol := strings.ToUpper(strings.TrimSpace(item))
		if symbol == "" {
			continue
		}
		if _, ok := seen[symbol]; ok {
			continue
		}
		seen[symbol] = struct{}{}
		result = append(result, symbol)
	}
	return result
}

func normalizeStockQuoteSyncMode(raw string, symbols []string) string {
	mode := strings.ToUpper(strings.TrimSpace(raw))
	switch mode {
	case "FULL_MARKET", "EXPLICIT", "SAMPLE":
		return mode
	}
	if len(symbols) > 0 {
		return "EXPLICIT"
	}
	return "SAMPLE"
}

func parseQuantEvaluationQuery(c *gin.Context) (int, int) {
	windowDays := 60
	if parsed, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("days", "60"))); err == nil && parsed > 0 {
		windowDays = parsed
	}
	if windowDays < 20 {
		windowDays = 20
	}
	if windowDays > 365 {
		windowDays = 365
	}
	topN := 10
	if parsed, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("top_n", "10"))); err == nil && parsed > 0 {
		topN = parsed
	}
	if topN < 1 {
		topN = 1
	}
	if topN > 30 {
		topN = 30
	}
	return windowDays, topN
}

func (h *AdminGrowthHandler) resolveDefaultStockQuoteSourceKey() string {
	items, _, err := h.service.AdminListSystemConfigs(stockDefaultSourceConfigKey, 1, 50)
	if err != nil {
		return stockDefaultSourceFallback
	}
	for _, item := range items {
		if !strings.EqualFold(strings.TrimSpace(item.ConfigKey), stockDefaultSourceConfigKey) {
			continue
		}
		sourceKey := strings.ToUpper(strings.TrimSpace(item.ConfigValue))
		if sourceKey != "" {
			return sourceKey
		}
	}
	return stockDefaultSourceFallback
}

func (h *AdminGrowthHandler) ListStockRecommendations(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListStockRecommendations(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateStockRecommendation(c *gin.Context) {
	var req dto.StockRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	validFrom, err := normalizeAdminDateTime(req.ValidFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "valid_from " + err.Error(), Data: struct{}{}})
		return
	}
	validTo, err := normalizeAdminDateTime(req.ValidTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "valid_to " + err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	operator = strings.TrimSpace(operator)
	sourceType := strings.ToUpper(strings.TrimSpace(req.SourceType))
	if sourceType == "" {
		sourceType = "MANUAL"
	}
	performanceLabel := strings.ToUpper(strings.TrimSpace(req.PerformanceLabel))
	if performanceLabel == "" {
		performanceLabel = "PENDING"
	}
	strategyVersion := strings.TrimSpace(req.StrategyVersion)
	if strategyVersion == "" {
		strategyVersion = "manual-v1"
	}
	publisher := strings.TrimSpace(req.Publisher)
	if publisher == "" {
		publisher = operator
	}
	id, err := h.service.AdminCreateStockRecommendation(model.StockRecommendation{
		Symbol:           req.Symbol,
		Name:             req.Name,
		Score:            req.Score,
		RiskLevel:        req.RiskLevel,
		PositionRange:    req.PositionRange,
		ValidFrom:        validFrom,
		ValidTo:          validTo,
		Status:           req.Status,
		ReasonSummary:    req.ReasonSummary,
		SourceType:       sourceType,
		StrategyVersion:  strategyVersion,
		Reviewer:         strings.TrimSpace(req.Reviewer),
		Publisher:        publisher,
		ReviewNote:       strings.TrimSpace(req.ReviewNote),
		PerformanceLabel: performanceLabel,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateStockRecommendationStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.StockRecommendationStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateStockRecommendationStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "STOCK", "UPDATE_RECOMMENDATION_STATUS", "STOCK_RECOMMENDATION", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) SyncStockQuotes(c *gin.Context) {
	var req dto.StockQuoteSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultStockQuoteSourceKey()
	}
	if sourceKey == "" {
		sourceKey = "MOCK"
	}
	days := req.Days
	if days <= 0 {
		days = 120
	}
	if days > 365 {
		days = 365
	}
	symbols := normalizeStockSymbols(req.Symbols)
	syncMode := normalizeStockQuoteSyncMode(req.SyncMode, symbols)

	var (
		result model.MarketSyncResult
		err    error
	)
	if syncMode == "FULL_MARKET" {
		result, err = h.service.AdminSyncStockQuotesFromMaster(sourceKey, days)
	} else {
		result, err = h.service.AdminSyncStockQuotesDetailed(sourceKey, symbols, days)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	count := result.TruthCount
	if count <= 0 {
		count = result.BarCount
	}
	beforeValue := requestedSourceKey
	if beforeValue == "" {
		beforeValue = "DEFAULT"
	}
	reason := fmt.Sprintf("days=%d,symbols=%d", days, len(symbols))
	h.writeOperationLog(c, "STOCK", "SYNC_QUOTES", "STOCK_QUOTES", sourceKey, beforeValue, "count="+strconv.Itoa(count), reason)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":                count,
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"days":                 days,
		"sync_mode":            syncMode,
		"symbols":              symbols,
		"result":               result,
	}))
}

func (h *AdminGrowthHandler) SyncStockInstrumentMaster(c *gin.Context) {
	var req dto.StockMasterSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	requestedSourceKey := strings.ToUpper(strings.TrimSpace(req.SourceKey))
	sourceKey := requestedSourceKey
	if sourceKey == "" {
		sourceKey = h.resolveDefaultConfigValue("stock.master.default_source_key", "TUSHARE")
	}
	symbols := normalizeStockSymbols(req.Symbols)
	result, err := h.service.AdminSyncStockInstrumentMaster(sourceKey, symbols)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	count := result.TruthCount
	beforeValue := requestedSourceKey
	if beforeValue == "" {
		beforeValue = "DEFAULT"
	}
	reason := fmt.Sprintf("symbols=%d", len(symbols))
	h.writeOperationLog(c, "STOCK", "SYNC_INSTRUMENT_MASTER", "STOCK_MASTER", sourceKey, beforeValue, "count="+strconv.Itoa(count), reason)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":                count,
		"source_key":           sourceKey,
		"requested_source_key": requestedSourceKey,
		"symbols":              symbols,
		"result":               result,
	}))
}

func (h *AdminGrowthHandler) ListQuantTopStocks(c *gin.Context) {
	limit := 10
	if parsed, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("limit", "10"))); err == nil && parsed > 0 {
		limit = parsed
	}
	if limit > 50 {
		limit = 50
	}

	lookbackDays := 120
	if parsed, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("lookback_days", "120"))); err == nil && parsed > 0 {
		lookbackDays = parsed
	}
	if lookbackDays < 30 {
		lookbackDays = 30
	}
	if lookbackDays > 365 {
		lookbackDays = 365
	}

	items, err := h.service.AdminGetQuantTopStocks(limit, lookbackDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"items":         items,
		"limit":         limit,
		"lookback_days": lookbackDays,
		"total":         len(items),
	}))
}

func (h *AdminGrowthHandler) ListQuantEvaluation(c *gin.Context) {
	windowDays, topN := parseQuantEvaluationQuery(c)
	summary, points, riskItems, rotationItems, err := h.service.AdminGetQuantEvaluation(windowDays, topN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"summary":        summary,
		"items":          points,
		"risk_items":     riskItems,
		"rotation_items": rotationItems,
		"total":          len(points),
	}))
}

func (h *AdminGrowthHandler) ExportQuantEvaluationCSV(c *gin.Context) {
	windowDays, topN := parseQuantEvaluationQuery(c)
	summary, points, riskItems, rotationItems, err := h.service.AdminGetQuantEvaluation(windowDays, topN)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"section", "field", "value"})
	_ = writer.Write([]string{"summary", "window_days", strconv.Itoa(summary.WindowDays)})
	_ = writer.Write([]string{"summary", "top_n", strconv.Itoa(summary.TopN)})
	_ = writer.Write([]string{"summary", "sample_days", strconv.Itoa(summary.SampleDays)})
	_ = writer.Write([]string{"summary", "sample_count", strconv.Itoa(summary.SampleCount)})
	_ = writer.Write([]string{"summary", "avg_return_5", fmt.Sprintf("%.6f", summary.AvgReturn5)})
	_ = writer.Write([]string{"summary", "hit_rate_5", fmt.Sprintf("%.6f", summary.HitRate5)})
	_ = writer.Write([]string{"summary", "max_drawdown_5", fmt.Sprintf("%.6f", summary.MaxDrawdown5)})
	_ = writer.Write([]string{"summary", "avg_return_10", fmt.Sprintf("%.6f", summary.AvgReturn10)})
	_ = writer.Write([]string{"summary", "hit_rate_10", fmt.Sprintf("%.6f", summary.HitRate10)})
	_ = writer.Write([]string{"summary", "max_drawdown_10", fmt.Sprintf("%.6f", summary.MaxDrawdown10)})
	_ = writer.Write([]string{"summary", "benchmark_avg_return_5", fmt.Sprintf("%.6f", summary.BenchmarkAvgReturn5)})
	_ = writer.Write([]string{"summary", "benchmark_avg_return_10", fmt.Sprintf("%.6f", summary.BenchmarkAvgReturn10)})
	_ = writer.Write([]string{"summary", "generated_at", summary.GeneratedAt})
	_ = writer.Write([]string{})

	_ = writer.Write([]string{
		"points_trade_date",
		"sample_count",
		"avg_return_5",
		"hit_rate_5",
		"benchmark_return_5",
		"avg_return_10",
		"hit_rate_10",
		"benchmark_return_10",
		"cumulative_return_5",
		"cumulative_benchmark_5",
		"cumulative_excess_5",
		"cumulative_return_10",
		"cumulative_benchmark_10",
		"cumulative_excess_10",
	})
	for _, item := range points {
		_ = writer.Write([]string{
			item.TradeDate,
			strconv.Itoa(item.SampleCount),
			fmt.Sprintf("%.6f", item.AvgReturn5),
			fmt.Sprintf("%.6f", item.HitRate5),
			fmt.Sprintf("%.6f", item.BenchmarkReturn),
			fmt.Sprintf("%.6f", item.AvgReturn10),
			fmt.Sprintf("%.6f", item.HitRate10),
			fmt.Sprintf("%.6f", item.BenchmarkReturn10),
			fmt.Sprintf("%.6f", item.CumulativeReturn5),
			fmt.Sprintf("%.6f", item.CumulativeBenchmark5),
			fmt.Sprintf("%.6f", item.CumulativeExcess5),
			fmt.Sprintf("%.6f", item.CumulativeReturn10),
			fmt.Sprintf("%.6f", item.CumulativeBenchmark10),
			fmt.Sprintf("%.6f", item.CumulativeExcess10),
		})
	}
	_ = writer.Write([]string{})

	_ = writer.Write([]string{"risk_level", "sample_count", "avg_return_5", "hit_rate_5", "avg_return_10", "hit_rate_10"})
	for _, item := range riskItems {
		_ = writer.Write([]string{
			item.RiskLevel,
			strconv.Itoa(item.SampleCount),
			fmt.Sprintf("%.6f", item.AvgReturn5),
			fmt.Sprintf("%.6f", item.HitRate5),
			fmt.Sprintf("%.6f", item.AvgReturn10),
			fmt.Sprintf("%.6f", item.HitRate10),
		})
	}
	_ = writer.Write([]string{})

	_ = writer.Write([]string{"rotation_trade_date", "top_symbols", "entered", "exited", "stayed_count", "changed_count"})
	for _, item := range rotationItems {
		_ = writer.Write([]string{
			item.TradeDate,
			strings.Join(item.TopSymbols, "|"),
			strings.Join(item.Entered, "|"),
			strings.Join(item.Exited, "|"),
			strconv.Itoa(item.StayedCount),
			strconv.Itoa(item.ChangedCount),
		})
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	fileName := fmt.Sprintf("stock_quant_evaluation_%dd_top%d.csv", windowDays, topN)
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.String(http.StatusOK, buf.String())
}

func (h *AdminGrowthHandler) GenerateDailyStockRecommendations(c *gin.Context) {
	tradeDate := c.Query("trade_date")
	result, err := h.service.AdminGenerateDailyStockRecommendations(tradeDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	afterValue := "count=" + strconv.Itoa(result.Count)
	if strings.TrimSpace(result.PublishID) != "" {
		afterValue += " publish_id=" + result.PublishID
	}
	if strings.TrimSpace(result.GenerationMode) != "" {
		afterValue += " generation_mode=" + result.GenerationMode
	}
	h.writeOperationLog(c, "STOCK", "GENERATE_DAILY_RECOMMENDATIONS", "BATCH", tradeDate, "", afterValue, "")
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":           result.Count,
		"publish_id":      result.PublishID,
		"publish_version": result.PublishVersion,
		"report_summary":  result.ReportSummary,
		"generation_mode": result.GenerationMode,
		"archive_enabled": result.ArchiveEnabled,
	}))
}

func (h *AdminGrowthHandler) ListStrategyEngineStockPublishHistory(c *gin.Context) {
	items, err := h.service.AdminListStrategyEnginePublishHistory("stock-selection")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "total": len(items)}))
}

func (h *AdminGrowthHandler) GetStrategyEngineStockPublishRecord(c *gin.Context) {
	publishID := strings.TrimSpace(c.Param("publish_id"))
	if publishID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "publish_id is required", Data: struct{}{}})
		return
	}
	item, err := h.service.AdminGetStrategyEnginePublishRecord(publishID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) GetStrategyEngineStockPublishReplay(c *gin.Context) {
	publishID := strings.TrimSpace(c.Param("publish_id"))
	if publishID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "publish_id is required", Data: struct{}{}})
		return
	}
	item, err := h.service.AdminGetStrategyEnginePublishReplay(publishID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) CompareStrategyEngineStockPublishVersions(c *gin.Context) {
	var req struct {
		LeftPublishID  string `json:"left_publish_id"`
		RightPublishID string `json:"right_publish_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	result, err := h.service.AdminCompareStrategyEnginePublishVersions(strings.TrimSpace(req.LeftPublishID), strings.TrimSpace(req.RightPublishID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}

func (h *AdminGrowthHandler) GenerateDailyFuturesStrategies(c *gin.Context) {
	tradeDate := c.Query("trade_date")
	result, err := h.service.AdminGenerateDailyFuturesStrategies(tradeDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	afterValue := "count=" + strconv.Itoa(result.Count)
	if strings.TrimSpace(result.PublishID) != "" {
		afterValue += " publish_id=" + result.PublishID
	}
	if strings.TrimSpace(result.GenerationMode) != "" {
		afterValue += " generation_mode=" + result.GenerationMode
	}
	h.writeOperationLog(c, "FUTURES", "GENERATE_DAILY_STRATEGIES", "BATCH", tradeDate, "", afterValue, "")
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"count":           result.Count,
		"publish_id":      result.PublishID,
		"publish_version": result.PublishVersion,
		"report_summary":  result.ReportSummary,
		"generation_mode": result.GenerationMode,
		"archive_enabled": result.ArchiveEnabled,
	}))
}

func (h *AdminGrowthHandler) ListStrategyEngineFuturesPublishHistory(c *gin.Context) {
	items, err := h.service.AdminListStrategyEnginePublishHistory("futures-strategy")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "total": len(items)}))
}

func (h *AdminGrowthHandler) GetStrategyEngineFuturesPublishRecord(c *gin.Context) {
	publishID := strings.TrimSpace(c.Param("publish_id"))
	if publishID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "publish_id is required", Data: struct{}{}})
		return
	}
	item, err := h.service.AdminGetStrategyEnginePublishRecord(publishID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) GetStrategyEngineFuturesPublishReplay(c *gin.Context) {
	publishID := strings.TrimSpace(c.Param("publish_id"))
	if publishID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "publish_id is required", Data: struct{}{}})
		return
	}
	item, err := h.service.AdminGetStrategyEnginePublishReplay(publishID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) CompareStrategyEngineFuturesPublishVersions(c *gin.Context) {
	var req struct {
		LeftPublishID  string `json:"left_publish_id"`
		RightPublishID string `json:"right_publish_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	result, err := h.service.AdminCompareStrategyEnginePublishVersions(strings.TrimSpace(req.LeftPublishID), strings.TrimSpace(req.RightPublishID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(result))
}

func (h *AdminGrowthHandler) ListFuturesStrategies(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	contract := c.Query("contract")
	items, total, err := h.service.AdminListFuturesStrategies(status, contract, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateFuturesStrategy(c *gin.Context) {
	var req dto.FuturesStrategyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	validFrom, err := normalizeAdminDateTime(req.ValidFrom)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "valid_from " + err.Error(), Data: struct{}{}})
		return
	}
	validTo, err := normalizeAdminDateTime(req.ValidTo)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "valid_to " + err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateFuturesStrategy(model.FuturesStrategy{
		Contract:      req.Contract,
		Name:          req.Name,
		Direction:     req.Direction,
		RiskLevel:     req.RiskLevel,
		PositionRange: req.PositionRange,
		ValidFrom:     validFrom,
		ValidTo:       validTo,
		Status:        req.Status,
		ReasonSummary: req.ReasonSummary,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "FUTURES", "CREATE_STRATEGY", "FUTURES_STRATEGY", id, "", req.Status, req.Contract)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateFuturesStrategyStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.FuturesStrategyStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateFuturesStrategyStatus(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "FUTURES", "UPDATE_STRATEGY_STATUS", "FUTURES_STRATEGY", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListMarketEvents(c *gin.Context) {
	page, pageSize := parsePage(c)
	eventType := strings.TrimSpace(c.Query("event_type"))
	symbol := strings.TrimSpace(c.Query("symbol"))
	items, total, err := h.service.AdminListMarketEvents(eventType, symbol, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ListMarketRhythmTasks(c *gin.Context) {
	taskDate := strings.TrimSpace(c.DefaultQuery("date", time.Now().Format("2006-01-02")))
	items, err := h.service.AdminListMarketRhythmTasks(taskDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"date": taskDate, "items": items}))
}

func (h *AdminGrowthHandler) EnsureMarketRhythmTasks(c *gin.Context) {
	var req dto.MarketRhythmTaskEnsureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	items, err := h.service.AdminEnsureMarketRhythmTasks(req.Date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "ENSURE_RHYTHM_TASKS", "MARKET_RHYTHM", req.Date, "", req.Date, "")
	c.JSON(http.StatusOK, dto.OK(gin.H{"date": req.Date, "items": items}))
}

func (h *AdminGrowthHandler) UpdateMarketRhythmTask(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "task id required", Data: struct{}{}})
		return
	}
	var req dto.MarketRhythmTaskUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	item, err := h.service.AdminUpdateMarketRhythmTask(id, req.Owner, req.Notes, req.SourceLinks, req.Status)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "market rhythm task not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "UPDATE_RHYTHM_TASK", "MARKET_RHYTHM", id, "", item.Status, item.TaskKey)
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) UpdateMarketRhythmTaskStatus(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "task id required", Data: struct{}{}})
		return
	}
	var req dto.MarketRhythmTaskStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	item, err := h.service.AdminUpdateMarketRhythmTaskStatus(id, req.Status, req.Owner, req.Notes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "market rhythm task not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "UPDATE_RHYTHM_TASK_STATUS", "MARKET_RHYTHM", id, "", item.Status, item.TaskKey)
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) GetExperimentAnalyticsSummary(c *gin.Context) {
	days, err := strconv.Atoi(strings.TrimSpace(c.DefaultQuery("days", "7")))
	if err != nil || days <= 0 {
		days = 7
	}
	if days > 30 {
		days = 30
	}
	summary, err := h.service.AdminGetExperimentAnalyticsSummary(days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(summary))
}

func (h *AdminGrowthHandler) CreateMarketEvent(c *gin.Context) {
	var req dto.MarketEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateMarketEvent(model.MarketEvent{
		EventType:   req.EventType,
		Symbol:      req.Symbol,
		Summary:     req.Summary,
		TriggerRule: req.TriggerRule,
		Source:      req.Source,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "CREATE_EVENT", "MARKET_EVENT", id, "", strings.ToUpper(strings.TrimSpace(req.EventType)), req.Symbol)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateMarketEvent(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "event id required", Data: struct{}{}})
		return
	}
	var req dto.MarketEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateMarketEvent(id, model.MarketEvent{
		EventType:   req.EventType,
		Symbol:      req.Symbol,
		Summary:     req.Summary,
		TriggerRule: req.TriggerRule,
		Source:      req.Source,
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "market event not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "MARKET", "UPDATE_EVENT", "MARKET_EVENT", id, "", strings.ToUpper(strings.TrimSpace(req.EventType)), req.Symbol)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ListUsers(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	kycStatus := c.Query("kyc_status")
	memberLevel := c.Query("member_level")
	registrationSource := c.Query("registration_source")
	items, total, err := h.service.AdminListUsers(status, kycStatus, memberLevel, registrationSource, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) UserSourceSummary(c *gin.Context) {
	status := c.Query("status")
	kycStatus := c.Query("kyc_status")
	memberLevel := c.Query("member_level")
	registrationSource := c.Query("registration_source")
	item, err := h.service.AdminGetUserSourceSummary(status, kycStatus, memberLevel, registrationSource)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) ListBrowseHistories(c *gin.Context) {
	page, pageSize := parsePage(c)
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

func (h *AdminGrowthHandler) BrowseHistorySummary(c *gin.Context) {
	item, err := h.service.AdminGetBrowseHistorySummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) BrowseHistoryTrend(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListBrowseUserSegments(c *gin.Context) {
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

func (h *AdminGrowthHandler) ExportBrowseHistoriesCSV(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListUserMessages(c *gin.Context) {
	page, pageSize := parsePage(c)
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

func (h *AdminGrowthHandler) CreateUserMessages(c *gin.Context) {
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
		users, _, err := h.service.AdminListUsers("ACTIVE", "", "", "", 1, 10000)
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
		fmt.Sprintf("type=%s", messageType),
		fmt.Sprintf("sent=%d failed=%d", sentCount, len(failures)),
	)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"sent_count":   sentCount,
		"failed_count": len(failures),
		"failures":     failures,
	}))
}

func (h *AdminGrowthHandler) GetUserCenterOverview(c *gin.Context) {
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

func parseTimeSafe(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return parsed, true
	}
	parsed, err = time.Parse("2006-01-02 15:04:05", value)
	if err == nil {
		return parsed, true
	}
	return time.Time{}, false
}

func safeRatio(numerator int, denominator int) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func (h *AdminGrowthHandler) UpdateUserSubscription(c *gin.Context) {
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

func (h *AdminGrowthHandler) ExportUsersCSV(c *gin.Context) {
	status := c.Query("status")
	kycStatus := c.Query("kyc_status")
	memberLevel := c.Query("member_level")
	registrationSource := c.Query("registration_source")
	items, _, err := h.service.AdminListUsers(status, kycStatus, memberLevel, registrationSource, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "phone", "email", "status", "kyc_status", "member_level", "registration_source", "inviter_user_id", "invite_code", "invite_registered_at", "created_at"})
	for _, it := range items {
		_ = writer.Write([]string{
			it.ID,
			it.Phone,
			it.Email,
			it.Status,
			it.KYCStatus,
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

func (h *AdminGrowthHandler) UpdateUserStatus(c *gin.Context) {
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

func (h *AdminGrowthHandler) UpdateUserMemberLevel(c *gin.Context) {
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

func (h *AdminGrowthHandler) UpdateUserKYCStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.UpdateUserKYCStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if err := h.service.AdminUpdateUserKYCStatus(id, req.KYCStatus); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "USER", "UPDATE_KYC_STATUS", "USER", id, "", req.KYCStatus, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) ResetUserPassword(c *gin.Context) {
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

func (h *AdminGrowthHandler) DashboardOverview(c *gin.Context) {
	item, err := h.service.AdminDashboardOverview()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) ListOperationLogs(c *gin.Context) {
	page, pageSize := parsePage(c)
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

func (h *AdminGrowthHandler) ListAuditEvents(c *gin.Context) {
	page, pageSize := parsePage(c)
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

func (h *AdminGrowthHandler) GetAuditEventSummary(c *gin.Context) {
	summary, err := h.service.AdminGetAuditEventSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(summary))
}

func (h *AdminGrowthHandler) ExportOperationLogsCSV(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListMembershipProducts(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListMembershipProducts(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateMembershipProduct(c *gin.Context) {
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

func (h *AdminGrowthHandler) UpdateMembershipProduct(c *gin.Context) {
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

func (h *AdminGrowthHandler) UpdateMembershipProductStatus(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListMembershipOrders(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	userID := c.Query("user_id")
	items, total, err := h.service.AdminListMembershipOrders(status, userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ExportMembershipOrdersCSV(c *gin.Context) {
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

func (h *AdminGrowthHandler) UpdateMembershipOrderStatus(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListVIPQuotaConfigs(c *gin.Context) {
	page, pageSize := parsePage(c)
	memberLevel := c.Query("member_level")
	status := c.Query("status")
	items, total, err := h.service.AdminListVIPQuotaConfigs(memberLevel, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateVIPQuotaConfig(c *gin.Context) {
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

func (h *AdminGrowthHandler) UpdateVIPQuotaConfig(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListUserQuotas(c *gin.Context) {
	page, pageSize := parsePage(c)
	userID := strings.TrimSpace(c.Query("user_id"))
	periodKey := strings.TrimSpace(c.Query("period_key"))
	items, total, err := h.service.AdminListUserQuotaUsages(userID, periodKey, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) AdjustUserQuota(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListDataSources(c *gin.Context) {
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListDataSources(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) CreateDataSource(c *gin.Context) {
	var req dto.DataSourceCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	id, err := h.service.AdminCreateDataSource(model.DataSource{
		SourceKey:  req.SourceKey,
		Name:       req.Name,
		SourceType: req.SourceType,
		Status:     req.Status,
		Config:     req.Config,
	})
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "exists") {
			c.JSON(http.StatusConflict, dto.APIResponse{Code: 40901, Message: err.Error(), Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "CREATE_DATA_SOURCE", "DATA_SOURCE", id, "", req.Status, req.SourceKey)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateDataSource(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	var req dto.DataSourceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	err := h.service.AdminUpdateDataSource(sourceKey, model.DataSource{
		Name:       req.Name,
		SourceType: req.SourceType,
		Status:     req.Status,
		Config:     req.Config,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "UPDATE_DATA_SOURCE", "DATA_SOURCE", sourceKey, "", req.Status, req.Name)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) DeleteDataSource(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	if err := h.service.AdminDeleteDataSource(sourceKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "DELETE_DATA_SOURCE", "DATA_SOURCE", sourceKey, "", "DELETED", "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) CheckDataSourceHealth(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	item, err := h.service.AdminCheckDataSourceHealth(sourceKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "CHECK_DATA_SOURCE_HEALTH", "DATA_SOURCE", sourceKey, "", item.Status, item.Message)
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) ListDataSourceHealthLogs(c *gin.Context) {
	sourceKey := strings.TrimSpace(c.Param("source_key"))
	page, pageSize := parsePage(c)
	items, total, err := h.service.AdminListDataSourceHealthLogs(sourceKey, page, pageSize)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "data source not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) BatchCheckDataSourcesHealth(c *gin.Context) {
	var req dto.DataSourceBatchHealthCheckRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
			return
		}
	}
	items, err := h.service.AdminBatchCheckDataSourceHealth(req.SourceKeys)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "BATCH_CHECK_DATA_SOURCE_HEALTH", "DATA_SOURCE", "BATCH", "", fmt.Sprintf("count=%d", len(items)), "")
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "count": len(items)}))
}

func (h *AdminGrowthHandler) ListSystemConfigs(c *gin.Context) {
	page, pageSize := parsePage(c)
	keyword := c.Query("keyword")
	items, total, err := h.service.AdminListSystemConfigs(keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) UpsertSystemConfig(c *gin.Context) {
	var req dto.SystemConfigUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpsertSystemConfig(req.ConfigKey, req.ConfigValue, req.Description, operator); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "UPSERT_CONFIG", "SYSTEM_CONFIG", req.ConfigKey, "", req.ConfigValue, req.Description)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) TestOSSQiniuConfig(c *gin.Context) {
	cfg := h.resolveOSSUploadConfig()
	if !strings.EqualFold(strings.TrimSpace(cfg.Provider), "QINIU") {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "oss provider is not QINIU", Data: struct{}{}})
		return
	}
	if !cfg.Enabled {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "oss is disabled", Data: struct{}{}})
		return
	}
	objectKey := joinObjectPath(
		cfg.PathPrefix,
		"probe",
		time.Now().Format("20060102"),
		fmt.Sprintf("probe_%d_%s.txt", time.Now().Unix(), randomHex(3)),
	)
	payload := []byte(fmt.Sprintf("sercherai oss qiniu probe %s", time.Now().Format(time.RFC3339)))
	fileURL, err := h.uploadToQiniu(cfg, objectKey, payload, "text/plain")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SYSTEM", "TEST_OSS_QINIU", "SYSTEM_CONFIG", "oss.provider", "", "SUCCESS", objectKey)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"provider":   "QINIU",
		"region":     cfg.Region,
		"bucket":     cfg.Bucket,
		"object_key": objectKey,
		"file_url":   fileURL,
		"size":       len(payload),
	}))
}

func (h *AdminGrowthHandler) TestYolkPayConfig(c *gin.Context) {
	cfg, err := h.resolveYolkPayConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if !cfg.PaymentEnabled {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "payment.enabled 未开启", Data: struct{}{}})
		return
	}
	if !cfg.Enabled {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "payment.channel.yolkpay.enabled 未开启", Data: struct{}{}})
		return
	}
	if strings.TrimSpace(cfg.PID) == "" || strings.TrimSpace(cfg.Key) == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "payment.channel.yolkpay.pid 或 key 未配置", Data: struct{}{}})
		return
	}

	queryEndpoint := strings.TrimRight(buildYolkPayGatewayURL(cfg.Gateway, "/api.php"), "/")
	values := url.Values{}
	values.Set("act", "query")
	values.Set("pid", strings.TrimSpace(cfg.PID))
	values.Set("key", strings.TrimSpace(cfg.Key))
	requestURL := queryEndpoint + "?" + values.Encode()

	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	resp, err := (&http.Client{Timeout: 12 * time.Second}).Do(req)
	if err != nil {
		c.JSON(http.StatusBadGateway, dto.APIResponse{Code: 50201, Message: err.Error(), Data: struct{}{}})
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 128*1024))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		message := strings.TrimSpace(string(bodyBytes))
		if message == "" {
			message = resp.Status
		}
		c.JSON(http.StatusBadGateway, dto.APIResponse{Code: 50201, Message: message, Data: struct{}{}})
		return
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &payload); err != nil {
		c.JSON(http.StatusBadGateway, dto.APIResponse{Code: 50201, Message: "蛋黄支付返回格式异常", Data: struct{}{}})
		return
	}
	if parseYolkPayCode(payload["code"]) != 1 {
		msg := stringifyYolkPayValue(payload["msg"])
		if msg == "" {
			msg = "蛋黄支付配置校验失败"
		}
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: msg, Data: payload})
		return
	}
	result := gin.H{
		"gateway": queryEndpoint,
		"pid":     stringifyYolkPayValue(payload["pid"]),
		"active":  stringifyYolkPayValue(payload["active"]),
		"money":   stringifyYolkPayValue(payload["money"]),
		"orders":  stringifyYolkPayValue(payload["orders"]),
		"raw":     payload,
	}
	pid := stringifyYolkPayValue(result["pid"])
	h.writeOperationLog(c, "SYSTEM", "TEST_PAYMENT_YOLKPAY", "SYSTEM_CONFIG", paymentChannelYolkPayPIDConfigKey, "", "SUCCESS", pid)
	c.JSON(http.StatusOK, dto.OK(result))
}

func (h *AdminGrowthHandler) ListReviewTasks(c *gin.Context) {
	page, pageSize := parsePage(c)
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

func (h *AdminGrowthHandler) ExportReviewTasksCSV(c *gin.Context) {
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

func (h *AdminGrowthHandler) WorkflowMetrics(c *gin.Context) {
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

func (h *AdminGrowthHandler) SubmitReviewTask(c *gin.Context) {
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

func (h *AdminGrowthHandler) ReviewTaskDecision(c *gin.Context) {
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

func (h *AdminGrowthHandler) AssignReviewTask(c *gin.Context) {
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

func (h *AdminGrowthHandler) ListSchedulerJobRuns(c *gin.Context) {
	page, pageSize := parsePage(c)
	jobName := c.Query("job_name")
	status := c.Query("status")
	items, total, err := h.service.AdminListSchedulerJobRuns(jobName, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ListNewsSyncRunDetails(c *gin.Context) {
	runID := strings.TrimSpace(c.Param("id"))
	if runID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "run id is required", Data: struct{}{}})
		return
	}
	page, pageSize := parsePage(c)
	syncType := c.Query("sync_type")
	source := c.Query("source")
	symbol := c.Query("symbol")
	status := c.Query("status")
	items, total, err := h.service.AdminListNewsSyncRunDetails(runID, syncType, source, symbol, status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ExportSchedulerJobRunsCSV(c *gin.Context) {
	jobName := c.Query("job_name")
	status := c.Query("status")
	items, _, err := h.service.AdminListSchedulerJobRuns(jobName, status, 1, 10000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	_ = writer.Write([]string{"id", "parent_run_id", "retry_count", "job_name", "trigger_source", "status", "started_at", "finished_at", "result_summary", "error_message", "operator_id"})
	for _, it := range items {
		_ = writer.Write([]string{it.ID, it.ParentRunID, strconv.Itoa(it.RetryCount), it.JobName, it.TriggerSource, it.Status, it.StartedAt, it.FinishedAt, it.ResultSummary, it.ErrorMessage, it.OperatorID})
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=scheduler_job_runs.csv")
	c.String(http.StatusOK, buf.String())
}

func (h *AdminGrowthHandler) SchedulerJobMetrics(c *gin.Context) {
	jobName := c.Query("job_name")
	item, err := h.service.AdminGetSchedulerJobMetrics(jobName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminGrowthHandler) ListSchedulerJobDefinitions(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	module := c.Query("module")
	items, total, err := h.service.AdminListSchedulerJobDefinitions(status, module, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminGrowthHandler) ListSupportedSchedulerJobs(c *gin.Context) {
	items := make([]supportedSchedulerJob, 0, len(supportedSchedulerJobs))
	moduleFilter := strings.ToUpper(strings.TrimSpace(c.Query("module")))
	for _, item := range supportedSchedulerJobs {
		if moduleFilter != "" && strings.ToUpper(item.Module) != moduleFilter {
			continue
		}
		items = append(items, item)
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "total": len(items)}))
}

func (h *AdminGrowthHandler) CreateSchedulerJobDefinition(c *gin.Context) {
	var req dto.SchedulerJobDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	id, err := h.service.AdminCreateSchedulerJobDefinition(model.SchedulerJobDefinition{
		JobName:     req.JobName,
		DisplayName: req.DisplayName,
		Module:      req.Module,
		CronExpr:    req.CronExpr,
		Status:      req.Status,
	}, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "CREATE_JOB_DEFINITION", "JOB_DEFINITION", id, "", req.Status, req.JobName)
	c.JSON(http.StatusOK, dto.OK(gin.H{"id": id}))
}

func (h *AdminGrowthHandler) UpdateSchedulerJobDefinition(c *gin.Context) {
	id := c.Param("id")
	var req dto.SchedulerJobDefinitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpdateSchedulerJobDefinition(id, model.SchedulerJobDefinition{
		JobName:     req.JobName,
		DisplayName: req.DisplayName,
		Module:      req.Module,
		CronExpr:    req.CronExpr,
		Status:      req.Status,
	}, operator); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "UPDATE_JOB_DEFINITION", "JOB_DEFINITION", id, "", req.Status, req.CronExpr)
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) UpdateSchedulerJobDefinitionStatus(c *gin.Context) {
	id := c.Param("id")
	var req dto.SchedulerJobDefinitionStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	if err := h.service.AdminUpdateSchedulerJobDefinitionStatus(id, req.Status, operator); err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "UPDATE_JOB_DEFINITION_STATUS", "JOB_DEFINITION", id, "", req.Status, "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) DeleteSchedulerJobDefinition(c *gin.Context) {
	id := strings.TrimSpace(c.Param("id"))
	if err := h.service.AdminDeleteSchedulerJobDefinition(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "job definition not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	h.writeOperationLog(c, "SCHEDULER", "DELETE_JOB_DEFINITION", "JOB_DEFINITION", id, "", "DELETED", "")
	c.JSON(http.StatusOK, dto.OK(struct{}{}))
}

func (h *AdminGrowthHandler) TriggerSchedulerJob(c *gin.Context) {
	var req dto.SchedulerTriggerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	simulateStatus := strings.ToUpper(strings.TrimSpace(req.SimulateStatus))
	if simulateStatus != "" && h.cfg.AllowJobSimulation {
		id, err := h.service.AdminCreateSchedulerJobRun(req.JobName, req.TriggerSource, simulateStatus, req.ResultSummary, req.ErrorMessage, operator)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		h.writeOperationLog(c, "SCHEDULER", "TRIGGER_JOB", "JOB", req.JobName, "", simulateStatus, req.TriggerSource)
		c.JSON(http.StatusOK, dto.OK(gin.H{"id": id, "status": simulateStatus}))
		return
	}
	syncOptions := buildTushareNewsSyncOptions(req.NewsSources, req.Symbols, req.SyncTypes, req.BatchSize)
	execResult, err := h.runSchedulerJob(req.JobName, syncOptions)
	status := "SUCCESS"
	errorMessage := ""
	if err != nil {
		status = "FAILED"
		errorMessage = err.Error()
	}
	resultSummary := execResult.Summary
	id, err := h.service.AdminCreateSchedulerJobRun(req.JobName, req.TriggerSource, status, resultSummary, errorMessage, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if len(execResult.NewsSyncDetails) > 0 {
		if detailErr := h.service.AdminCreateNewsSyncRunDetails(id, execResult.NewsSyncDetails); detailErr != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: detailErr.Error(), Data: struct{}{}})
			return
		}
	}
	finalRunID, finalStatus, finalSummary, finalError, retryAttempts, retryErr := h.executeSchedulerAutoRetry(
		req.JobName,
		id,
		status,
		resultSummary,
		errorMessage,
		operator,
		syncOptions,
	)
	if retryErr != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: retryErr.Error(), Data: struct{}{}})
		return
	}
	reason := req.TriggerSource
	if retryAttempts > 0 {
		reason = fmt.Sprintf("%s,auto_retry=%d", req.TriggerSource, retryAttempts)
	}
	h.writeOperationLog(c, "SCHEDULER", "TRIGGER_JOB", "JOB", req.JobName, "", finalStatus, reason)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"id":                 finalRunID,
		"status":             finalStatus,
		"first_run_id":       id,
		"retry_attempts":     retryAttempts,
		"result_summary":     finalSummary,
		"error_message":      finalError,
		"auto_retry_applied": retryAttempts > 0,
	}))
}

func (h *AdminGrowthHandler) RetrySchedulerJobRun(c *gin.Context) {
	runID := c.Param("id")
	var req dto.SchedulerRetryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	simulateStatus := strings.ToUpper(strings.TrimSpace(req.SimulateStatus))
	if simulateStatus != "" && h.cfg.AllowJobSimulation {
		id, err := h.service.AdminRetrySchedulerJobRun(runID, "MANUAL", simulateStatus, req.ResultSummary, req.ErrorMessage, operator)
		if err != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
			return
		}
		h.writeOperationLog(c, "SCHEDULER", "RETRY_JOB", "JOB_RUN", runID, "", simulateStatus, req.ResultSummary)
		c.JSON(http.StatusOK, dto.OK(gin.H{"id": id, "status": simulateStatus}))
		return
	}
	jobName, err := h.service.GetSchedulerJobNameByRunID(runID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "job run not found", Data: struct{}{}})
		return
	}
	syncOptions := buildTushareNewsSyncOptions(req.NewsSources, req.Symbols, req.SyncTypes, req.BatchSize)
	execResult, runErr := h.runSchedulerJob(jobName, syncOptions)
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	resultSummary := execResult.Summary
	id, err := h.service.AdminRetrySchedulerJobRun(runID, "MANUAL", status, resultSummary, errorMessage, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if len(execResult.NewsSyncDetails) > 0 {
		if detailErr := h.service.AdminCreateNewsSyncRunDetails(id, execResult.NewsSyncDetails); detailErr != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: detailErr.Error(), Data: struct{}{}})
			return
		}
	}
	finalRunID, finalStatus, finalSummary, finalError, retryAttempts, retryErr := h.executeSchedulerAutoRetry(
		jobName,
		id,
		status,
		resultSummary,
		errorMessage,
		operator,
		syncOptions,
	)
	if retryErr != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: retryErr.Error(), Data: struct{}{}})
		return
	}
	reason := resultSummary
	if retryAttempts > 0 {
		reason = fmt.Sprintf("auto_retry=%d, %s", retryAttempts, strings.TrimSpace(resultSummary))
	}
	h.writeOperationLog(c, "SCHEDULER", "RETRY_JOB", "JOB_RUN", runID, "", finalStatus, reason)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"id":                 finalRunID,
		"status":             finalStatus,
		"first_run_id":       id,
		"retry_attempts":     retryAttempts,
		"result_summary":     finalSummary,
		"error_message":      finalError,
		"auto_retry_applied": retryAttempts > 0,
	}))
}

func (h *AdminGrowthHandler) RetryNewsSyncItem(c *gin.Context) {
	runID := strings.TrimSpace(c.Param("id"))
	if runID == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "run id is required", Data: struct{}{}})
		return
	}
	var req dto.RetryNewsSyncItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	jobName, err := h.service.GetSchedulerJobNameByRunID(runID)
	if err != nil {
		c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "job run not found", Data: struct{}{}})
		return
	}
	if !strings.EqualFold(strings.TrimSpace(jobName), "tushare_news_incremental") {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "only tushare_news_incremental supports fine-grained retry", Data: struct{}{}})
		return
	}
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)

	syncOptions := buildTushareNewsSyncOptions(
		[]string{req.Source},
		[]string{req.Symbol},
		[]string{req.SyncType},
		req.BatchSize,
	)
	execResult, runErr := h.runSchedulerJob(jobName, syncOptions)
	status := "SUCCESS"
	errorMessage := ""
	if runErr != nil {
		status = "FAILED"
		errorMessage = runErr.Error()
	}
	resultSummary := execResult.Summary
	if resultSummary == "" {
		resultSummary = "retry single news sync item"
	}

	newRunID, err := h.service.AdminRetrySchedulerJobRun(runID, "MANUAL", status, resultSummary, errorMessage, operator)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	if len(execResult.NewsSyncDetails) > 0 {
		if detailErr := h.service.AdminCreateNewsSyncRunDetails(newRunID, execResult.NewsSyncDetails); detailErr != nil {
			c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: detailErr.Error(), Data: struct{}{}})
			return
		}
	}
	h.writeOperationLog(
		c,
		"SCHEDULER",
		"RETRY_NEWS_SYNC_ITEM",
		"JOB_RUN",
		runID,
		"",
		status,
		fmt.Sprintf("sync_type=%s source=%s symbol=%s", strings.TrimSpace(req.SyncType), strings.TrimSpace(req.Source), strings.TrimSpace(req.Symbol)),
	)
	c.JSON(http.StatusOK, dto.OK(gin.H{
		"id":             newRunID,
		"status":         status,
		"result_summary": resultSummary,
		"error_message":  errorMessage,
	}))
}

func (h *AdminGrowthHandler) ListWorkflowMessages(c *gin.Context) {
	page, pageSize := parsePage(c)
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

func (h *AdminGrowthHandler) ExportWorkflowMessagesCSV(c *gin.Context) {
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

func (h *AdminGrowthHandler) UpdateWorkflowMessageRead(c *gin.Context) {
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

func (h *AdminGrowthHandler) CountUnreadWorkflowMessages(c *gin.Context) {
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

func (h *AdminGrowthHandler) BulkReadWorkflowMessages(c *gin.Context) {
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

func uniqueNonEmptyStrings(values []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))
	for _, value := range values {
		current := strings.TrimSpace(value)
		if current == "" {
			continue
		}
		if _, exists := seen[current]; exists {
			continue
		}
		seen[current] = struct{}{}
		result = append(result, current)
	}
	return result
}

func normalizeUpperValues(values []string) []string {
	seen := make(map[string]struct{})
	result := make([]string, 0, len(values))
	for _, value := range values {
		current := strings.ToUpper(strings.TrimSpace(value))
		if current == "" {
			continue
		}
		if _, exists := seen[current]; exists {
			continue
		}
		seen[current] = struct{}{}
		result = append(result, current)
	}
	return result
}

func buildTushareNewsSyncOptions(newsSources []string, symbols []string, syncTypes []string, batchSize int) model.TushareNewsSyncOptions {
	opts := model.TushareNewsSyncOptions{
		BatchSize: batchSize,
		Sources:   uniqueNonEmptyStrings(newsSources),
		Symbols:   normalizeUpperValues(symbols),
		SyncTypes: normalizeUpperValues(syncTypes),
	}
	if opts.BatchSize < 0 {
		opts.BatchSize = 0
	}
	return opts
}

func (h *AdminGrowthHandler) writeOperationLog(c *gin.Context, module string, action string, targetType string, targetID string, beforeValue string, afterValue string, reason string) {
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = "admin_unknown"
	}
	_ = h.service.AdminCreateOperationLog(module, action, targetType, targetID, operator, beforeValue, afterValue, reason)
}

func (h *AdminGrowthHandler) writeAuditEvent(c *gin.Context, item model.AdminAuditEvent) {
	operatorVal, _ := c.Get("user_id")
	operator, _ := operatorVal.(string)
	operator = strings.TrimSpace(operator)
	if operator == "" {
		operator = "admin_unknown"
	}
	if strings.TrimSpace(item.ActorUserID) == "" {
		item.ActorUserID = operator
	}
	_ = h.service.AdminCreateAuditEvent(item)
}

type schedulerJobExecutionResult struct {
	Summary         string
	NewsSyncDetails []model.NewsSyncRunDetail
}

type schedulerAutoRetryPolicy struct {
	Enabled        bool
	MaxRetries     int
	BackoffSeconds int
}

func parseConfigBool(raw string, fallback bool) bool {
	text := strings.ToLower(strings.TrimSpace(raw))
	if text == "" {
		return fallback
	}
	switch text {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}

func parseConfigInt(raw string, fallback int) int {
	text := strings.TrimSpace(raw)
	if text == "" {
		return fallback
	}
	value, err := strconv.Atoi(text)
	if err != nil {
		return fallback
	}
	return value
}

func (h *AdminGrowthHandler) resolveSchedulerAutoRetryPolicy(jobName string) schedulerAutoRetryPolicy {
	policy := schedulerAutoRetryPolicy{
		Enabled:        strings.EqualFold(strings.TrimSpace(jobName), schedulerAutoRetryDefaultJob),
		MaxRetries:     2,
		BackoffSeconds: 2,
	}
	items, _, err := h.service.AdminListSystemConfigs("scheduler.auto_retry", 1, 200)
	if err != nil || len(items) == 0 {
		return policy
	}
	allowedJobs := map[string]struct{}{
		strings.ToLower(strings.TrimSpace(schedulerAutoRetryDefaultJob)): {},
	}
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.ConfigKey))
		value := strings.TrimSpace(item.ConfigValue)
		switch key {
		case strings.ToLower(schedulerAutoRetryEnabledConfigKey):
			policy.Enabled = parseConfigBool(value, policy.Enabled)
		case strings.ToLower(schedulerAutoRetryMaxRetriesConfigKey):
			policy.MaxRetries = parseConfigInt(value, policy.MaxRetries)
		case strings.ToLower(schedulerAutoRetryBackoffSecondsConfigKey):
			policy.BackoffSeconds = parseConfigInt(value, policy.BackoffSeconds)
		case strings.ToLower(schedulerAutoRetryJobsConfigKey):
			allowedJobs = map[string]struct{}{}
			for _, name := range strings.Split(value, ",") {
				normalized := strings.ToLower(strings.TrimSpace(name))
				if normalized == "" {
					continue
				}
				allowedJobs[normalized] = struct{}{}
			}
		}
	}
	if policy.MaxRetries < 0 {
		policy.MaxRetries = 0
	}
	if policy.MaxRetries > 5 {
		policy.MaxRetries = 5
	}
	if policy.BackoffSeconds < 0 {
		policy.BackoffSeconds = 0
	}
	if policy.BackoffSeconds > 60 {
		policy.BackoffSeconds = 60
	}
	if len(allowedJobs) > 0 {
		if _, ok := allowedJobs[strings.ToLower(strings.TrimSpace(jobName))]; !ok {
			policy.Enabled = false
		}
	}
	if policy.MaxRetries <= 0 {
		policy.Enabled = false
	}
	return policy
}

func (h *AdminGrowthHandler) executeSchedulerAutoRetry(jobName string, baseRunID string, baseStatus string, baseSummary string, baseError string, operator string, syncOptions model.TushareNewsSyncOptions) (string, string, string, string, int, error) {
	finalRunID := strings.TrimSpace(baseRunID)
	finalStatus := strings.ToUpper(strings.TrimSpace(baseStatus))
	finalSummary := baseSummary
	finalError := baseError
	retryAttempts := 0
	if finalStatus != "FAILED" || finalRunID == "" {
		return finalRunID, finalStatus, finalSummary, finalError, retryAttempts, nil
	}
	policy := h.resolveSchedulerAutoRetryPolicy(jobName)
	if !policy.Enabled {
		return finalRunID, finalStatus, finalSummary, finalError, retryAttempts, nil
	}
	currentRunID := finalRunID
	for attempt := 1; attempt <= policy.MaxRetries && finalStatus == "FAILED"; attempt++ {
		if policy.BackoffSeconds > 0 {
			time.Sleep(time.Duration(policy.BackoffSeconds*attempt) * time.Second)
		}
		execResult, runErr := h.runSchedulerJob(jobName, syncOptions)
		summary := execResult.Summary
		status := "SUCCESS"
		errorMessage := ""
		if runErr != nil {
			status = "FAILED"
			errorMessage = runErr.Error()
		}
		newRunID, createErr := h.service.AdminRetrySchedulerJobRun(currentRunID, "SYSTEM", status, summary, errorMessage, operator)
		if createErr != nil {
			return finalRunID, finalStatus, finalSummary, finalError, retryAttempts, createErr
		}
		if len(execResult.NewsSyncDetails) > 0 {
			if detailErr := h.service.AdminCreateNewsSyncRunDetails(newRunID, execResult.NewsSyncDetails); detailErr != nil {
				return finalRunID, finalStatus, finalSummary, finalError, retryAttempts, detailErr
			}
		}
		retryAttempts = attempt
		currentRunID = newRunID
		finalRunID = newRunID
		finalStatus = status
		finalSummary = summary
		finalError = errorMessage
	}
	return finalRunID, finalStatus, finalSummary, finalError, retryAttempts, nil
}

func (h *AdminGrowthHandler) runSchedulerJob(jobName string, syncOptions model.TushareNewsSyncOptions) (schedulerJobExecutionResult, error) {
	switch strings.ToLower(strings.TrimSpace(jobName)) {
	case "daily_stock_quant_pipeline":
		tradeDate := time.Now().Format("2006-01-02")
		sourceKey := strings.ToUpper(strings.TrimSpace(h.resolveDefaultStockQuoteSourceKey()))
		if sourceKey == "" {
			sourceKey = stockDefaultSourceFallback
		}
		usedSourceKey := sourceKey
		quoteCount, err := h.service.AdminSyncStockQuotes(sourceKey, nil, 180)
		if err != nil && sourceKey != "MOCK" {
			fallbackCount, fallbackErr := h.service.AdminSyncStockQuotes("MOCK", nil, 180)
			if fallbackErr != nil {
				return schedulerJobExecutionResult{}, fmt.Errorf("sync quotes failed(%s): %v, fallback MOCK failed: %w", sourceKey, err, fallbackErr)
			}
			quoteCount = fallbackCount
			usedSourceKey = "MOCK"
		}
		if err != nil && sourceKey == "MOCK" {
			return schedulerJobExecutionResult{}, err
		}
		topItems, err := h.service.AdminGetQuantTopStocks(10, 180)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		recoResult, err := h.service.AdminGenerateDailyStockRecommendations(tradeDate)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{
			Summary: fmt.Sprintf(
				"trade_date=%s source=%s quotes=%d top=%d recommendations=%d",
				tradeDate,
				usedSourceKey,
				quoteCount,
				len(topItems),
				recoResult.Count,
			),
		}, nil
	case "daily_stock_recommendation":
		tradeDate := time.Now().Format("2006-01-02")
		result, err := h.service.AdminGenerateDailyStockRecommendations(tradeDate)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{Summary: fmt.Sprintf("generated %d recommendations", result.Count)}, nil
	case schedulerJobDailyFuturesStrategy, schedulerJobFuturesStrategyGenerate:
		tradeDate := time.Now().Format("2006-01-02")
		result, err := h.service.AdminGenerateDailyFuturesStrategies(tradeDate)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{
			Summary: fmt.Sprintf("trade_date=%s generated=%d", tradeDate, result.Count),
		}, nil
	case schedulerJobFuturesStrategyEvaluate:
		summary, err := h.runFuturesStrategyEvaluate(20)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{Summary: summary}, nil
	case "doc_fast_news_incremental":
		summary, err := h.service.AdminSyncDocFastNewsIncremental(0)
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{Summary: summary}, nil
	case "tushare_news_incremental":
		summary, details, err := h.service.AdminSyncTushareNewsIncrementalWithOptions(syncOptions)
		if err != nil {
			if isTushareRateLimitRuntimeError(err.Error()) {
				safeSummary := strings.TrimSpace(summary)
				if safeSummary == "" {
					safeSummary = "tushare_news_incremental rate_limited=true"
				} else if !strings.Contains(strings.ToLower(safeSummary), "rate_limited=true") {
					safeSummary += " rate_limited=true"
				}
				return schedulerJobExecutionResult{Summary: safeSummary, NewsSyncDetails: details}, nil
			}
			return schedulerJobExecutionResult{Summary: summary, NewsSyncDetails: details}, err
		}
		return schedulerJobExecutionResult{Summary: summary, NewsSyncDetails: details}, nil
	case "vip_membership_lifecycle":
		summary, err := h.service.AdminRunVIPMembershipLifecycle()
		if err != nil {
			return schedulerJobExecutionResult{}, err
		}
		return schedulerJobExecutionResult{Summary: summary}, nil
	default:
		return schedulerJobExecutionResult{}, fmt.Errorf("unknown job: %s", jobName)
	}
}

func isTushareRateLimitRuntimeError(message string) bool {
	normalized := strings.NewReplacer("|", ";", "\n", ";").Replace(strings.TrimSpace(message))
	segments := strings.Split(normalized, ";")
	hasSegment := false
	for _, segment := range segments {
		text := strings.TrimSpace(segment)
		if text == "" {
			continue
		}
		hasSegment = true
		if isTusharePermissionRuntimeErrorText(text) {
			return false
		}
		if !isTushareRateLimitRuntimeSegment(text) {
			return false
		}
	}
	return hasSegment
}

func isTusharePermissionRuntimeErrorText(message string) bool {
	text := strings.ToLower(strings.TrimSpace(message))
	if text == "" {
		return false
	}
	keywords := []string{
		"没有接口访问权限",
		"无权限",
		"permission denied",
		"no permission",
	}
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func isTushareRateLimitRuntimeSegment(message string) bool {
	text := strings.ToLower(strings.TrimSpace(message))
	if text == "" {
		return false
	}
	keywords := []string{
		"最多访问该接口",
		"触发访问频次",
		"请求过于频繁",
		"rate limit",
		"too many requests",
	}
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	if strings.Contains(text, "每分钟") && strings.Contains(text, "访问") {
		return true
	}
	if strings.Contains(text, "每小时") && strings.Contains(text, "访问") {
		return true
	}
	if strings.Contains(text, "每天") && strings.Contains(text, "访问") {
		return true
	}
	return false
}

type futuresStrategyEvaluateResult struct {
	TradeDate            string
	EvaluatedCount       int
	SuccessCount         int
	FailedCount          int
	AvgScore             float64
	AvgWinRate           float64
	AvgExcessReturn      float64
	MaxDrawdown          float64
	BenchmarkActualCount int
	TopStrategyID        string
	TopScore             float64
}

func (h *AdminGrowthHandler) runFuturesStrategyEvaluate(limit int) (string, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	collectors := make([]model.FuturesStrategy, 0, limit)
	seen := make(map[string]struct{})
	statuses := []string{"PUBLISHED", "ACTIVE"}
	for _, status := range statuses {
		if len(collectors) >= limit {
			break
		}
		items, _, err := h.service.AdminListFuturesStrategies(status, "", 1, limit)
		if err != nil {
			return "", err
		}
		for _, item := range items {
			if len(collectors) >= limit {
				break
			}
			id := strings.TrimSpace(item.ID)
			if id == "" {
				continue
			}
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			collectors = append(collectors, item)
		}
	}

	if len(collectors) == 0 {
		items, _, err := h.service.AdminListFuturesStrategies("", "", 1, limit)
		if err != nil {
			return "", err
		}
		for _, item := range items {
			id := strings.TrimSpace(item.ID)
			if id == "" {
				continue
			}
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			collectors = append(collectors, item)
			if len(collectors) >= limit {
				break
			}
		}
	}

	result := futuresStrategyEvaluateResult{
		TradeDate:      time.Now().Format("2006-01-02"),
		EvaluatedCount: len(collectors),
	}

	scoreSum := 0.0
	scoreCount := 0
	winRateSum := 0.0
	winRateCount := 0
	excessSum := 0.0
	excessCount := 0
	topScore := -1.0
	for _, strategy := range collectors {
		insight, err := h.service.GetFuturesStrategyInsight("admin_scheduler", strategy.ID)
		if err != nil {
			result.FailedCount++
			continue
		}
		result.SuccessCount++

		score := insight.ScoreFramework.TotalScore
		if score <= 0 {
			score = insight.ScoreFramework.WeightedScore
		}
		if score > 0 {
			scoreSum += score
			scoreCount++
			if score > topScore {
				topScore = score
				result.TopScore = score
				result.TopStrategyID = strategy.ID
			}
		}

		stats := insight.PerformanceStats
		if stats.SampleDays > 0 {
			winRateSum += stats.WinRate
			winRateCount++
			excessSum += stats.ExcessReturn
			excessCount++
			if stats.MaxDrawdown > result.MaxDrawdown {
				result.MaxDrawdown = stats.MaxDrawdown
			}
		}
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(stats.BenchmarkSource)), "actual") {
			result.BenchmarkActualCount++
		}
	}
	if result.EvaluatedCount-result.SuccessCount > 0 {
		result.FailedCount = result.EvaluatedCount - result.SuccessCount
	}
	if scoreCount > 0 {
		result.AvgScore = scoreSum / float64(scoreCount)
	}
	if winRateCount > 0 {
		result.AvgWinRate = winRateSum / float64(winRateCount)
	}
	if excessCount > 0 {
		result.AvgExcessReturn = excessSum / float64(excessCount)
	}

	return fmt.Sprintf(
		"trade_date=%s evaluated=%d success=%d failed=%d avg_score=%.2f avg_win_rate=%.4f avg_excess=%.4f max_drawdown=%.4f benchmark_actual=%d/%d top_id=%s top_score=%.2f",
		result.TradeDate,
		result.EvaluatedCount,
		result.SuccessCount,
		result.FailedCount,
		result.AvgScore,
		result.AvgWinRate,
		result.AvgExcessReturn,
		result.MaxDrawdown,
		result.BenchmarkActualCount,
		result.SuccessCount,
		strings.TrimSpace(result.TopStrategyID),
		result.TopScore,
	), nil
}
