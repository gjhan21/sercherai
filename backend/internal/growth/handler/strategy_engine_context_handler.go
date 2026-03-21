package handler

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"sercherai/backend/internal/growth/dto"
	"sercherai/backend/internal/growth/model"
)

func (h *AdminGrowthHandler) InternalStrategyEngineStockSelectionContext(c *gin.Context) {
	var req dto.StrategyEngineStockSelectionContextRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	resp, err := h.service.BuildStrategyEngineStockSelectionContext(model.StrategyEngineStockSelectionContextRequest{
		TradeDate:        req.TradeDate,
		SelectionMode:    req.SelectionMode,
		UniverseScope:    req.UniverseScope,
		ProfileID:        req.ProfileID,
		DebugSeedSymbols: req.DebugSeedSymbols,
		SeedSymbols:      req.SeedSymbols,
		ExcludedSymbols:  req.ExcludedSymbols,
		Limit:            req.Limit,
		MarketScope:      req.MarketScope,
		MinListingDays:   req.MinListingDays,
		MinAvgTurnover:   req.MinAvgTurnover,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *AdminGrowthHandler) InternalStrategyEngineFuturesStrategyContext(c *gin.Context) {
	var req dto.StrategyEngineFuturesStrategyContextRequest
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	resp, err := h.service.BuildStrategyEngineFuturesStrategyContext(model.StrategyEngineFuturesStrategyContextRequest{
		TradeDate:                       req.TradeDate,
		Contracts:                       req.Contracts,
		Limit:                           req.Limit,
		AllowMockFallbackOnShortHistory: req.AllowMockFallbackOnShortHistory,
	})
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}
