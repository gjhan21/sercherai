package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
)

func (h *AdminStrategyHandler) ListStrategyEngineStockPublishHistory(c *gin.Context) {
	items, err := h.service.AdminListStrategyEnginePublishHistory("stock-selection")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "total": len(items)}))
}

func (h *AdminStrategyHandler) GetStrategyEngineStockPublishRecord(c *gin.Context) {
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

func (h *AdminStrategyHandler) GetStrategyEngineStockPublishReplay(c *gin.Context) {
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

func (h *AdminStrategyHandler) CompareStrategyEngineStockPublishVersions(c *gin.Context) {
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

func (h *AdminStrategyHandler) ListStrategyEngineFuturesPublishHistory(c *gin.Context) {
	items, err := h.service.AdminListStrategyEnginePublishHistory("futures-strategy")
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(gin.H{"items": items, "total": len(items)}))
}

func (h *AdminStrategyHandler) GetStrategyEngineFuturesPublishRecord(c *gin.Context) {
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

func (h *AdminStrategyHandler) GetStrategyEngineFuturesPublishReplay(c *gin.Context) {
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

func (h *AdminStrategyHandler) CompareStrategyEngineFuturesPublishVersions(c *gin.Context) {
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

func (h *AdminStrategyHandler) GetStrategyGraphSnapshot(c *gin.Context) {
	item, err := h.service.AdminGetStrategyGraphSnapshot(c.Param("snapshot_id"))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "graph snapshot not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}

func (h *AdminStrategyHandler) QueryStrategyGraphSubgraph(c *gin.Context) {
	depth := 1
	if rawDepth := strings.TrimSpace(c.Query("depth")); rawDepth != "" {
		value, err := strconv.Atoi(rawDepth)
		if err != nil || value <= 0 {
			c.JSON(http.StatusBadRequest, dto.APIResponse{Code: 40001, Message: "depth must be a positive integer", Data: struct{}{}})
			return
		}
		depth = value
	}
	item, err := h.service.AdminQueryStrategyGraphSubgraph(model.StrategyGraphSubgraphQuery{
		EntityType:  c.Query("entity_type"),
		EntityKey:   c.Query("entity_key"),
		Depth:       depth,
		AssetDomain: c.Query("asset_domain"),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, dto.APIResponse{Code: 40401, Message: "graph subgraph not found", Data: struct{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.APIResponse{Code: 50001, Message: err.Error(), Data: struct{}{}})
		return
	}
	c.JSON(http.StatusOK, dto.OK(item))
}
