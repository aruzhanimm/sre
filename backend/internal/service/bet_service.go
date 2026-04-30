package service

import (
	"context"
	"errors"
	"strconv"
	"time"

	"BetKZ/internal/models"
	"BetKZ/internal/repository"
	ws "BetKZ/internal/websocket"

	"github.com/google/uuid"
)

type BetService struct {
	betRepo     *repository.BetRepository
	userRepo    *repository.UserRepository
	oddsRepo    *repository.OddsRepository
	marketRepo  *repository.MarketRepository
	txRepo      *repository.TransactionRepository
	oddsService *OddsService
	hub         *ws.Hub
}

func NewBetService(
	betRepo *repository.BetRepository,
	userRepo *repository.UserRepository,
	oddsRepo *repository.OddsRepository,
	marketRepo *repository.MarketRepository,
	txRepo *repository.TransactionRepository,
	oddsService *OddsService,
	hub *ws.Hub,
) *BetService {
	return &BetService{
		betRepo:     betRepo,
		userRepo:    userRepo,
		oddsRepo:    oddsRepo,
		marketRepo:  marketRepo,
		txRepo:      txRepo,
		oddsService: oddsService,
		hub:         hub,
	}
}

type PlaceBetSelection struct {
	MarketID string `json:"market_id" binding:"required"`
	OddID    string `json:"odd_id" binding:"required"`
	Outcome  string `json:"outcome" binding:"required"`
}

type PlaceBetRequest struct {
	Stake      float64             `json:"stake" binding:"required"`
	Selections []PlaceBetSelection `json:"selections" binding:"required"`
}

func (s *BetService) PlaceBet(ctx context.Context, userIDStr string, req *PlaceBetRequest) (*models.Bet, error) {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	// Validate stake
	if req.Stake < 0.50 {
		return nil, errors.New("minimum stake is 0.50")
	}
	if req.Stake > 10000 {
		return nil, errors.New("maximum stake is 10000")
	}
	if len(req.Selections) < 1 {
		return nil, errors.New("at least one selection required")
	}

	// Check user balance
	balance, err := s.userRepo.GetBalance(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if balance < req.Stake {
		return nil, errors.New("insufficient balance")
	}

	// Validate selections and calculate potential return
	betType := "single"
	if len(req.Selections) > 1 {
		betType = "accumulator"
	}

	totalOdds := 1.0
	type validatedSelection struct {
		marketID uuid.UUID
		oddID    uuid.UUID
		outcome  string
		odds     float64
		eventID  string
	}
	var validated []validatedSelection

	for _, sel := range req.Selections {
		marketID, err := uuid.Parse(sel.MarketID)
		if err != nil {
			return nil, errors.New("invalid market ID")
		}
		oddID, err := uuid.Parse(sel.OddID)
		if err != nil {
			return nil, errors.New("invalid odd ID")
		}

		// Check market is open
		market, err := s.marketRepo.GetByID(ctx, marketID)
		if err != nil || market == nil {
			return nil, errors.New("market not found")
		}
		if market.Status != "open" {
			return nil, errors.New("market is not accepting bets")
		}

		// Check odds
		odd, err := s.oddsRepo.GetByID(ctx, oddID)
		if err != nil || odd == nil {
			return nil, errors.New("odds not found")
		}
		if odd.MarketID != marketID {
			return nil, errors.New("odds do not belong to specified market")
		}
		if odd.Outcome != sel.Outcome {
			return nil, errors.New("outcome mismatch")
		}

		totalOdds *= odd.CurrentOdds
		validated = append(validated, validatedSelection{
			marketID: marketID,
			oddID:    oddID,
			outcome:  sel.Outcome,
			odds:     odd.CurrentOdds,
			eventID:  market.EventID.String(),
		})
	}

	potentialReturn := req.Stake * totalOdds

	// Deduct balance
	if err := s.userRepo.UpdateBalance(ctx, userID, -req.Stake); err != nil {
		return nil, errors.New("insufficient balance")
	}

	// Create bet
	bet := &models.Bet{
		UserID:          userID,
		BetType:         betType,
		Stake:           req.Stake,
		PotentialReturn: potentialReturn,
		Status:          "pending",
	}

	if err := s.betRepo.Create(ctx, bet); err != nil {
		// Refund
		_ = s.userRepo.UpdateBalance(ctx, userID, req.Stake)
		return nil, errors.New("failed to place bet")
	}

	// Create legs and update stakes
	for _, sel := range validated {
		leg := &models.BetLeg{
			BetID:          bet.ID,
			MarketID:       sel.marketID,
			OddID:          sel.oddID,
			Outcome:        sel.outcome,
			LockedOddValue: sel.odds,
		}
		if err := s.betRepo.CreateLeg(ctx, leg); err != nil {
			return nil, err
		}
		bet.Legs = append(bet.Legs, *leg)

		// Update stake on this odd
		if err := s.oddsRepo.UpdateStake(ctx, sel.oddID, req.Stake); err != nil {
			return nil, err
		}

		// Recalculate odds
		newOdds, err := s.oddsService.CalculateOdds(ctx, sel.marketID)
		if err == nil && s.hub != nil {
			for _, o := range newOdds {
				s.hub.BroadcastOddsUpdate(sel.eventID, &ws.OddsUpdateMessage{
					MarketID:  sel.marketID.String(),
					Outcome:   o.Outcome,
					OldOdds:   sel.odds,
					NewOdds:   o.CurrentOdds,
					Change:    o.CurrentOdds - sel.odds,
					Direction: oddsDirection(sel.odds, o.CurrentOdds),
					Timestamp: time.Now().UnixMilli(),
				})
			}
		}
	}

	// Record transaction
	tx := &models.Transaction{
		UserID:        userID,
		Type:          "bet_placed",
		Amount:        -req.Stake,
		BalanceBefore: balance,
		BalanceAfter:  balance - req.Stake,
		ReferenceID:   &bet.ID,
		Status:        "completed",
	}
	_ = s.txRepo.Create(ctx, tx)

	return bet, nil
}

func (s *BetService) GetBet(ctx context.Context, idStr string) (*models.Bet, error) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		return nil, errors.New("invalid bet ID")
	}
	bet, err := s.betRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if bet == nil {
		return nil, errors.New("bet not found")
	}
	return bet, nil
}

func (s *BetService) ListBets(ctx context.Context, userIDStr string, status *string, page, limit int) (*repository.BetListResult, error) {
	filters := repository.BetFilters{
		Status: status,
		Page:   page,
		Limit:  limit,
	}

	if userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.New("invalid user ID")
		}
		filters.UserID = &uid
	}

	return s.betRepo.List(ctx, filters)
}

func (s *BetService) ListTransactions(ctx context.Context, userIDStr string, txType *string, page, limit int) (*repository.TxListResult, error) {
	filters := repository.TxFilters{
		Type:  txType,
		Page:  page,
		Limit: limit,
	}

	if userIDStr != "" {
		uid, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, errors.New("invalid user ID")
		}
		filters.UserID = &uid
	}

	return s.txRepo.List(ctx, filters)
}

// Admin: deposit money to user
func (s *BetService) Deposit(ctx context.Context, userIDStr string, amount float64) error {
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.New("invalid user ID")
	}

	if amount <= 0 {
		return errors.New("amount must be positive")
	}

	balance, err := s.userRepo.GetBalance(ctx, userID)
	if err != nil {
		return errors.New("user not found")
	}

	if err := s.userRepo.UpdateBalance(ctx, userID, amount); err != nil {
		return err
	}

	tx := &models.Transaction{
		UserID:        userID,
		Type:          "deposit",
		Amount:        amount,
		BalanceBefore: balance,
		BalanceAfter:  balance + amount,
		Status:        "completed",
	}
	return s.txRepo.Create(ctx, tx)
}

// Admin: settle all bets for a market
type SettleMarketRequest struct {
	MarketID       string `json:"market_id" binding:"required"`
	WinningOutcome string `json:"winning_outcome" binding:"required"`
}

func (s *BetService) SettleMarket(ctx context.Context, req *SettleMarketRequest) (int, error) {
	marketID, err := uuid.Parse(req.MarketID)
	if err != nil {
		return 0, errors.New("invalid market ID")
	}

	market, err := s.marketRepo.GetByID(ctx, marketID)
	if err != nil || market == nil {
		return 0, errors.New("market not found")
	}

	if market.Status == "settled" {
		return 0, errors.New("market already settled")
	}

	// Lock the market
	if err := s.marketRepo.UpdateStatus(ctx, marketID, "settled"); err != nil {
		return 0, err
	}

	// Get all pending bets with legs on this market
	pendingBets, err := s.betRepo.GetPendingBetsByMarket(ctx, marketID)
	if err != nil {
		return 0, err
	}

	settled := 0
	for _, bet := range pendingBets {
		legs, err := s.betRepo.GetLegsByBetID(ctx, bet.ID)
		if err != nil {
			continue
		}

		allLegsSettled := true
		allLegsWon := true

		for _, leg := range legs {
			if leg.MarketID == marketID {
				result := "lost"
				if leg.Outcome == req.WinningOutcome {
					result = "won"
				}
				_ = s.betRepo.UpdateLegResult(ctx, leg.ID, result)
				if result == "lost" {
					allLegsWon = false
				}
			} else if leg.Result == "" || leg.Result == "pending" {
				allLegsSettled = false
			} else if leg.Result == "lost" {
				allLegsWon = false
			}
		}

		if !allLegsSettled {
			continue
		}

		if allLegsWon {
			_ = s.betRepo.UpdateStatus(ctx, bet.ID, "won", bet.PotentialReturn)
			_ = s.userRepo.UpdateBalance(ctx, bet.UserID, bet.PotentialReturn)

			balance, _ := s.userRepo.GetBalance(ctx, bet.UserID)
			tx := &models.Transaction{
				UserID:        bet.UserID,
				Type:          "bet_won",
				Amount:        bet.PotentialReturn,
				BalanceBefore: balance - bet.PotentialReturn,
				BalanceAfter:  balance,
				ReferenceID:   &bet.ID,
				Status:        "completed",
			}
			_ = s.txRepo.Create(ctx, tx)
		} else {
			_ = s.betRepo.UpdateStatus(ctx, bet.ID, "lost", 0)
		}
		settled++
	}

	return settled, nil
}

// Admin: deposit money to user by email
func (s *BetService) DepositByEmail(ctx context.Context, email string, amount float64) (*models.User, error) {
	if email == "" {
		return nil, errors.New("email is required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("internal error")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if err := s.Deposit(ctx, user.ID.String(), amount); err != nil {
		return nil, err
	}

	// Re-fetch to get updated balance
	user, _ = s.userRepo.GetByID(ctx, user.ID)
	return user, nil
}

// Admin: place bet on behalf of a user by email
func (s *BetService) AdminPlaceBet(ctx context.Context, email string, req *PlaceBetRequest) (*models.Bet, error) {
	if email == "" {
		return nil, errors.New("user email is required")
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("internal error")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return s.PlaceBet(ctx, user.ID.String(), req)
}

// Dashboard stats
func (s *BetService) GetDashboardStats(ctx context.Context) (*repository.DashboardStats, error) {
	return s.betRepo.GetDashboardStats(ctx)
}

func oddsDirection(old, new float64) string {
	if new > old {
		return "up"
	}
	if new < old {
		return "down"
	}
	return "stable"
}

// ParsePageLimit parses page and limit from string parameters
func ParsePageLimit(pageStr, limitStr string) (int, int) {
	page := 1
	limit := 20
	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
		limit = l
	}
	return page, limit
}
