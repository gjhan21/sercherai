package handler

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/platform/oss"
	"sercherai/backend/internal/platform/utils"
)

type AdminNewsHandler struct {
	AdminBaseHandler
}

func NewAdminNewsHandler(base *AdminBaseHandler) *AdminNewsHandler {
	return &AdminNewsHandler{AdminBaseHandler: *base}
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


func (h *AdminNewsHandler) ListNewsCategories(c *gin.Context) {
	page, pageSize := parsePage(c)
	status := c.Query("status")
	items, total, err := h.service.AdminListNewsCategories(status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "page": page, "page_size": pageSize, "total": total}))
}

func (h *AdminNewsHandler) CreateNewsCategory(c *gin.Context) {
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

func (h *AdminNewsHandler) UpdateNewsCategory(c *gin.Context) {
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

func (h *AdminNewsHandler) ListNewsArticles(c *gin.Context) {
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

func (h *AdminNewsHandler) GetNewsArticleDetail(c *gin.Context) {
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

func (h *AdminNewsHandler) CreateNewsArticle(c *gin.Context) {
	var req dto.NewsArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: err.Error(), Data: struct{}{}})
		return
	}
	authorID := currentAdminOperator(c)
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

func (h *AdminNewsHandler) UpdateNewsArticle(c *gin.Context) {
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

func (h *AdminNewsHandler) PublishNewsArticle(c *gin.Context) {
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

func (h *AdminNewsHandler) UploadNewsAttachment(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "missing file", Data: struct{}{}})
		return
	}

	originalName := utils.SanitizeUploadFileName(fileHeader.Filename)
	if originalName == "" {
		c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "invalid file name", Data: struct{}{}})
		return
	}

	ossClientCfg := h.resolveOSSUploadConfig()
	maxUploadMB := h.cfg.AttachmentUploadMaxMB
	if ossClientCfg.MaxUploadMB > 0 {
		maxUploadMB = ossClientCfg.MaxUploadMB
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
	if !h.isAllowedNewsAttachmentMime(mimeType) {
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
	ext := utils.ResolveUploadExt(originalName, mimeType)
	targetName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), utils.RandomHex(4), ext)
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

	if ossClientCfg.Enabled && strings.EqualFold(strings.TrimSpace(ossClientCfg.Provider), "QINIU") {
		objectKey := utils.JoinObjectPath(ossClientCfg.PathPrefix, "news", dateDir, targetName)
		qiniuCfg := oss.QiniuConfig{
			AccessKey: ossClientCfg.AccessKey,
			SecretKey: ossClientCfg.SecretKey,
			Bucket:    ossClientCfg.Bucket,
			Domain:    ossClientCfg.Domain,
			Region:    ossClientCfg.Region,
			UseHTTPS:  ossClientCfg.UseHTTPS,
		}
		fileURL, uploadErr := oss.UploadToQiniu(qiniuCfg, objectKey, payload, mimeType)
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

func (h *AdminNewsHandler) CreateNewsAttachment(c *gin.Context) {
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

func (h *AdminNewsHandler) ListNewsAttachments(c *gin.Context) {
	articleID := c.Param("id")
	items, err := h.service.AdminListNewsAttachments(articleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items}))
}

func (h *AdminNewsHandler) DeleteNewsAttachment(c *gin.Context) {
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

func (h *AdminNewsHandler) isAllowedNewsAttachmentMime(mimeType string) bool {
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

