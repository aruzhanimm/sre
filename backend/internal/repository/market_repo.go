package repository

import (
	"context"
	"errors"

	"BetKZ/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MarketRepository struct {
	db *pgxpool.Pool
}

func NewMarketRepository(db *pgxpool.Pool) *MarketRepository {
	return &MarketRepository{db: db}
}

func (r *MarketRepository) GetByEventID(ctx context.Context, eventID uuid.UUID) ([]models.Market, error) {
	query := `SELECT id, event_id, market_type, name, line, status, margin_percentage, created_at, updated_at
	          FROM markets WHERE event_id = $1 ORDER BY created_at`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var markets []models.Market
	for rows.Next() {
		var m models.Market
		if err := rows.Scan(&m.ID, &m.EventID, &m.MarketType, &m.Name, &m.Line, &m.Status,
			&m.MarginPercentage, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		markets = append(markets, m)
	}
	return markets, nil
}

func (r *MarketRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Market, error) {
	m := &models.Market{}
	query := `SELECT id, event_id, market_type, name, line, status, margin_percentage, created_at, updated_at
	          FROM markets WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(&m.ID, &m.EventID, &m.MarketType, &m.Name, &m.Line,
		&m.Status, &m.MarginPercentage, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return m, nil
}

func (r *MarketRepository) Create(ctx context.Context, market *models.Market) error {
	query := `INSERT INTO markets (event_id, market_type, name, line, status, margin_percentage)
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		market.EventID, market.MarketType, market.Name, market.Line, market.Status, market.MarginPercentage,
	).Scan(&market.ID, &market.CreatedAt, &market.UpdatedAt)
}

func (r *MarketRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE markets SET status = $2 WHERE id = $1`, id, status)
	return err
}

func (r *MarketRepository) LockMarketsByEvent(ctx context.Context, eventID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE markets SET status = 'locked' WHERE event_id = $1 AND status = 'open'`, eventID)
	return err
}

func (r *MarketRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `DELETE FROM markets WHERE id = $1`, id)
	return err
}
