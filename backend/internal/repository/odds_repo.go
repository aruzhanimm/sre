package repository

import (
	"context"
	"errors"

	"BetKZ/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OddsRepository struct {
	db *pgxpool.Pool
}

func NewOddsRepository(db *pgxpool.Pool) *OddsRepository {
	return &OddsRepository{db: db}
}

func (r *OddsRepository) GetByMarketID(ctx context.Context, marketID uuid.UUID) ([]models.Odd, error) {
	query := `SELECT id, market_id, outcome, initial_odds, current_odds, total_stake, bet_count,
	                 last_calculated_at, is_manual_override, created_at, updated_at
	          FROM odds WHERE market_id = $1 ORDER BY outcome`

	rows, err := r.db.Query(ctx, query, marketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var odds []models.Odd
	for rows.Next() {
		var o models.Odd
		if err := rows.Scan(&o.ID, &o.MarketID, &o.Outcome, &o.InitialOdds, &o.CurrentOdds,
			&o.TotalStake, &o.BetCount, &o.LastCalculatedAt, &o.IsManualOverride,
			&o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		odds = append(odds, o)
	}
	return odds, nil
}

func (r *OddsRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Odd, error) {
	o := &models.Odd{}
	query := `SELECT id, market_id, outcome, initial_odds, current_odds, total_stake, bet_count,
	                 last_calculated_at, is_manual_override, created_at, updated_at
	          FROM odds WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(&o.ID, &o.MarketID, &o.Outcome, &o.InitialOdds,
		&o.CurrentOdds, &o.TotalStake, &o.BetCount, &o.LastCalculatedAt, &o.IsManualOverride,
		&o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return o, nil
}

func (r *OddsRepository) Create(ctx context.Context, odd *models.Odd) error {
	query := `INSERT INTO odds (market_id, outcome, initial_odds, current_odds)
	          VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		odd.MarketID, odd.Outcome, odd.InitialOdds, odd.CurrentOdds,
	).Scan(&odd.ID, &odd.CreatedAt, &odd.UpdatedAt)
}

func (r *OddsRepository) UpdateCurrentOdds(ctx context.Context, id uuid.UUID, newOdds float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE odds SET current_odds = $2, last_calculated_at = NOW() WHERE id = $1`, id, newOdds)
	return err
}

func (r *OddsRepository) UpdateStake(ctx context.Context, id uuid.UUID, additionalStake float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE odds SET total_stake = total_stake + $2, bet_count = bet_count + 1 WHERE id = $1`,
		id, additionalStake)
	return err
}

func (r *OddsRepository) ManualOverride(ctx context.Context, id uuid.UUID, newOdds float64, lock bool) error {
	_, err := r.db.Exec(ctx,
		`UPDATE odds SET current_odds = $2, is_manual_override = $3, last_calculated_at = NOW()
		 WHERE id = $1`, id, newOdds, lock)
	return err
}

func (r *OddsRepository) GetHistory(ctx context.Context, oddID uuid.UUID, limit int) ([]models.OddsHistory, error) {
	query := `SELECT id, odd_id, odds_value, total_stake, bet_count, recorded_at
	          FROM odds_history WHERE odd_id = $1 ORDER BY recorded_at DESC LIMIT $2`

	rows, err := r.db.Query(ctx, query, oddID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []models.OddsHistory
	for rows.Next() {
		var h models.OddsHistory
		if err := rows.Scan(&h.ID, &h.OddID, &h.OddsValue, &h.TotalStake, &h.BetCount, &h.RecordedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}
