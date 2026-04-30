package handlers

import (
	"net/http"

	"BetKZ/internal/service"

	"github.com/gin-gonic/gin"
)

type MarketHandler struct {
	oddsService *service.OddsService
}

func NewMarketHandler(oddsService *service.OddsService) *MarketHandler {
	return &MarketHandler{oddsService: oddsService}
}

// GET /api/events/:id/markets
func (h *MarketHandler) GetEventMarkets(c *gin.Context) {
	markets, err := h.oddsService.GetEventMarkets(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"markets": markets})
}

// GET /api/markets/:id/odds
func (h *MarketHandler) GetMarketOdds(c *gin.Context) {
	odds, err := h.oddsService.GetMarketOdds(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"odds": odds})
}

// POST /api/admin/markets
func (h *MarketHandler) CreateMarket(c *gin.Context) {
	var req service.CreateMarketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	market, err := h.oddsService.CreateMarket(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, market)
}

// PUT /api/admin/odds/:id
type ManualOverrideRequest struct {
	Odds float64 `json:"odds" binding:"required"`
	Lock bool    `json:"lock"`
}

func (h *MarketHandler) ManualOverride(c *gin.Context) {
	var req ManualOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.oddsService.ManualOverride(c.Request.Context(), c.Param("id"), req.Odds, req.Lock)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "odds updated"})
}

// GET /api/admin/odds/:id/history
func (h *MarketHandler) GetOddsHistory(c *gin.Context) {
	history, err := h.oddsService.GetOddsHistory(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"history": history})
}
