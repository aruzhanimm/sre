package handlers

import (
	"net/http"

	"BetKZ/internal/service"

	"github.com/gin-gonic/gin"
)

type BetHandler struct {
	betService *service.BetService
}

func NewBetHandler(betService *service.BetService) *BetHandler {
	return &BetHandler{betService: betService}
}

// POST /api/bets
func (h *BetHandler) PlaceBet(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req service.PlaceBetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bet, err := h.betService.PlaceBet(c.Request.Context(), userID.(string), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, bet)
}

// GET /api/bets
func (h *BetHandler) ListBets(c *gin.Context) {
	userID, _ := c.Get("user_id")
	status := c.Query("status")
	page, limit := service.ParsePageLimit(c.Query("page"), c.Query("limit"))

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	result, err := h.betService.ListBets(c.Request.Context(), userID.(string), statusPtr, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bets"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GET /api/bets/:id
func (h *BetHandler) GetBet(c *gin.Context) {
	bet, err := h.betService.GetBet(c.Request.Context(), c.Param("id"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, bet)
}

// GET /api/transactions
func (h *BetHandler) ListTransactions(c *gin.Context) {
	userID, _ := c.Get("user_id")
	txType := c.Query("type")
	page, limit := service.ParsePageLimit(c.Query("page"), c.Query("limit"))

	var typePtr *string
	if txType != "" {
		typePtr = &txType
	}

	result, err := h.betService.ListTransactions(c.Request.Context(), userID.(string), typePtr, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// POST /api/admin/users/:id/deposit
type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required"`
}

func (h *BetHandler) Deposit(c *gin.Context) {
	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.betService.Deposit(c.Request.Context(), c.Param("id"), req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deposit successful"})
}

// POST /api/admin/deposit-by-email
type DepositByEmailRequest struct {
	Email  string  `json:"email" binding:"required"`
	Amount float64 `json:"amount" binding:"required"`
}

func (h *BetHandler) DepositByEmail(c *gin.Context) {
	var req DepositByEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.betService.DepositByEmail(c.Request.Context(), req.Email, req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":     "deposit successful",
		"user_email":  user.Email,
		"new_balance": user.Balance,
	})
}

// POST /api/admin/place-bet
type AdminPlaceBetRequest struct {
	UserEmail  string                      `json:"user_email" binding:"required"`
	Stake      float64                     `json:"stake" binding:"required"`
	Selections []service.PlaceBetSelection `json:"selections" binding:"required"`
}

func (h *BetHandler) AdminPlaceBet(c *gin.Context) {
	var req AdminPlaceBetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	betReq := &service.PlaceBetRequest{
		Stake:      req.Stake,
		Selections: req.Selections,
	}

	bet, err := h.betService.AdminPlaceBet(c.Request.Context(), req.UserEmail, betReq)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, bet)
}

// GET /api/admin/bets
func (h *BetHandler) AdminListBets(c *gin.Context) {
	status := c.Query("status")
	page, limit := service.ParsePageLimit(c.Query("page"), c.Query("limit"))

	var statusPtr *string
	if status != "" {
		statusPtr = &status
	}

	result, err := h.betService.ListBets(c.Request.Context(), "", statusPtr, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bets"})
		return
	}
	c.JSON(http.StatusOK, result)
}

// POST /api/admin/settle
func (h *BetHandler) SettleMarket(c *gin.Context) {
	var req service.SettleMarketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settled, err := h.betService.SettleMarket(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "settlement complete", "settled_count": settled})
}

// GET /api/admin/stats
func (h *BetHandler) GetDashboardStats(c *gin.Context) {
	stats, err := h.betService.GetDashboardStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}
