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

func (h *AdminGrowthHandler) GetStrategyGraphSnapshot(c *gin.Context) {
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

func (h *AdminGrowthHandler) QueryStrategyGraphSubgraph(c *gin.Context) {
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
