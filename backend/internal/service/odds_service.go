package service

import (
	"context"
	"errors"
	"math"

	"BetKZ/internal/models"
	"BetKZ/internal/repository"

	"github.com/google/uuid"
)

type OddsService struct {
	oddsRepo   *repository.OddsRepository
	marketRepo *repository.MarketRepository
}

func NewOddsService(oddsRepo *repository.OddsRepository, marketRepo *repository.MarketRepository) *OddsService {
	return &OddsService{
		oddsRepo:   oddsRepo,
		marketRepo: marketRepo,
	}
}

// CalculateOdds recalculates odds for all outcomes in a market based on stakes
func (s *OddsService) CalculateOdds(ctx context.Context, marketID uuid.UUID) ([]models.Odd, error) {
	market, err := s.marketRepo.GetByID(ctx, marketID)
	if err != nil || market == nil {
		return nil, errors.New("market not found")
	}

	odds, err := s.oddsRepo.GetByMarketID(ctx, marketID)
	if err != nil {
		return nil, err
	}

	// Calculate total pool
	totalPool := 0.0
	for _, odd := range odds {
		totalPool += odd.TotalStake
	}

	// If no bets, keep initial odds
	if totalPool == 0 {
		return odds, nil
	}

	margin := market.MarginPercentage / 100.0

	for i, odd := range odds {
		if odd.IsManualOverride {
			continue
		}

		// Implied probability from stakes
		impliedProb := odd.TotalStake / totalPool

		// Adjust with house margin
		adjustedProb := impliedProb * (1 + margin)

		// Prevent division by zero
		if adjustedProb <= 0 {
			adjustedProb = 0.01
		}

		// Calculate new odds
		newOdds := 1.0 / adjustedProb

		// Apply boundaries (1.01 to 100.00)
		newOdds = math.Max(1.01, math.Min(newOdds, 100.00))

		// Round to 2 decimal places
		newOdds = math.Round(newOdds*100) / 100

		// Update in DB
		if err := s.oddsRepo.UpdateCurrentOdds(ctx, odd.ID, newOdds); err != nil {
			return nil, err
		}

		odds[i].CurrentOdds = newOdds
	}

	return odds, nil
}

// GetMarketOdds returns all odds for a market
func (s *OddsService) GetMarketOdds(ctx context.Context, marketIDStr string) ([]models.Odd, error) {
	marketID, err := uuid.Parse(marketIDStr)
	if err != nil {
		return nil, errors.New("invalid market ID")
	}
	return s.oddsRepo.GetByMarketID(ctx, marketID)
}

// GetEventMarkets returns all markets with their odds for an event
func (s *OddsService) GetEventMarkets(ctx context.Context, eventIDStr string) ([]models.Market, error) {
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		return nil, errors.New("invalid event ID")
	}

	markets, err := s.marketRepo.GetByEventID(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Load odds for each market
	for i := range markets {
		odds, err := s.oddsRepo.GetByMarketID(ctx, markets[i].ID)
		if err != nil {
			return nil, err
		}
		markets[i].Odds = odds
	}

	return markets, nil
}

type CreateMarketRequest struct {
	EventID          string            `json:"event_id" binding:"required"`
	MarketType       string            `json:"market_type" binding:"required"`
	Name             string            `json:"name" binding:"required"`
	Line             *float64          `json:"line"`
	MarginPercentage float64           `json:"margin_percentage"`
	InitialOdds      []InitialOddInput `json:"initial_odds" binding:"required"`
}

type InitialOddInput struct {
	Outcome string  `json:"outcome" binding:"required"`
	Odds    float64 `json:"odds" binding:"required"`
}

func (s *OddsService) CreateMarket(ctx context.Context, req *CreateMarketRequest) (*models.Market, error) {
	eventID, err := uuid.Parse(req.EventID)
	if err != nil {
		return nil, errors.New("invalid event ID")
	}

	validTypes := map[string]bool{"1x2": true, "over_under": true, "both_teams_score": true, "double_chance": true, "handicap": true, "custom": true}
	if !validTypes[req.MarketType] {
		return nil, errors.New("invalid market type")
	}

	if req.MarginPercentage <= 0 {
		req.MarginPercentage = 5.00
	}

	market := &models.Market{
		EventID:          eventID,
		MarketType:       req.MarketType,
		Name:             req.Name,
		Line:             req.Line,
		Status:           "open",
		MarginPercentage: req.MarginPercentage,
	}

	if err := s.marketRepo.Create(ctx, market); err != nil {
		return nil, err
	}

	// Create odds for each outcome
	seenOutcomes := make(map[string]bool)
	for _, input := range req.InitialOdds {
		if seenOutcomes[input.Outcome] {
			return nil, errors.New("duplicate outcome: " + input.Outcome)
		}
		seenOutcomes[input.Outcome] = true

		if input.Odds < 1.01 {
			return nil, errors.New("odds must be >= 1.01")
		}

		odd := &models.Odd{
			MarketID:    market.ID,
			Outcome:     input.Outcome,
			InitialOdds: input.Odds,
			CurrentOdds: input.Odds,
		}
		if err := s.oddsRepo.Create(ctx, odd); err != nil {
			return nil, err
		}
		market.Odds = append(market.Odds, *odd)
	}

	return market, nil
}

func (s *OddsService) ManualOverride(ctx context.Context, oddIDStr string, newOdds float64, lock bool) error {
	oddID, err := uuid.Parse(oddIDStr)
	if err != nil {
		return errors.New("invalid odd ID")
	}

	if newOdds < 1.01 {
		return errors.New("odds must be >= 1.01")
	}

	return s.oddsRepo.ManualOverride(ctx, oddID, newOdds, lock)
}

func (s *OddsService) GetOddsHistory(ctx context.Context, oddIDStr string) ([]models.OddsHistory, error) {
	oddID, err := uuid.Parse(oddIDStr)
	if err != nil {
		return nil, errors.New("invalid odd ID")
	}
	return s.oddsRepo.GetHistory(ctx, oddID, 100)
}
