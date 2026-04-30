package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"BetKZ/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BetRepository struct {
	db *pgxpool.Pool
}

func NewBetRepository(db *pgxpool.Pool) *BetRepository {
	return &BetRepository{db: db}
}

func (r *BetRepository) Create(ctx context.Context, bet *models.Bet) error {
	query := `INSERT INTO bets (user_id, bet_type, stake, potential_return, status)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id, placed_at`

	return r.db.QueryRow(ctx, query,
		bet.UserID, bet.BetType, bet.Stake, bet.PotentialReturn, bet.Status,
	).Scan(&bet.ID, &bet.PlacedAt)
}

func (r *BetRepository) CreateLeg(ctx context.Context, leg *models.BetLeg) error {
	query := `INSERT INTO bet_legs (bet_id, market_id, odd_id, outcome, locked_odd_value)
	          VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`

	return r.db.QueryRow(ctx, query,
		leg.BetID, leg.MarketID, leg.OddID, leg.Outcome, leg.LockedOddValue,
	).Scan(&leg.ID, &leg.CreatedAt)
}

func (r *BetRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Bet, error) {
	bet := &models.Bet{}
	query := `SELECT id, user_id, bet_type, stake, potential_return, actual_return, status, placed_at, settled_at
	          FROM bets WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&bet.ID, &bet.UserID, &bet.BetType, &bet.Stake, &bet.PotentialReturn,
		&bet.ActualReturn, &bet.Status, &bet.PlacedAt, &bet.SettledAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// Load legs
	legs, err := r.GetLegsByBetID(ctx, bet.ID)
	if err != nil {
		return nil, err
	}
	bet.Legs = legs

	return bet, nil
}

func (r *BetRepository) GetLegsByBetID(ctx context.Context, betID uuid.UUID) ([]models.BetLeg, error) {
	query := `SELECT bl.id, bl.bet_id, bl.market_id, bl.odd_id, bl.outcome, bl.locked_odd_value, bl.result, bl.created_at,
	                 COALESCE(e.home_team || ' vs ' || e.away_team, '') as event_name,
	                 COALESCE(m.market_type, '') as market_type
	          FROM bet_legs bl
	          JOIN markets m ON bl.market_id = m.id
	          JOIN events e ON m.event_id = e.id
	          WHERE bl.bet_id = $1`

	rows, err := r.db.Query(ctx, query, betID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var legs []models.BetLeg
	for rows.Next() {
		var l models.BetLeg
		if err := rows.Scan(&l.ID, &l.BetID, &l.MarketID, &l.OddID, &l.Outcome,
			&l.LockedOddValue, &l.Result, &l.CreatedAt, &l.EventName, &l.MarketType); err != nil {
			return nil, err
		}
		legs = append(legs, l)
	}
	return legs, nil
}

type BetFilters struct {
	UserID *uuid.UUID
	Status *string
	Page   int
	Limit  int
}

type BetListResult struct {
	Bets  []models.Bet `json:"bets"`
	Total int          `json:"total"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
}

func (r *BetRepository) List(ctx context.Context, filters BetFilters) (*BetListResult, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *filters.UserID)
		argIdx++
	}
	if filters.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filters.Status)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM bets %s`, whereClause)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Fetch
	offset := (filters.Page - 1) * filters.Limit
	query := fmt.Sprintf(`SELECT id, user_id, bet_type, stake, potential_return, actual_return, status, placed_at, settled_at
	          FROM bets %s ORDER BY placed_at DESC LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
	args = append(args, filters.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bets []models.Bet
	for rows.Next() {
		var b models.Bet
		if err := rows.Scan(&b.ID, &b.UserID, &b.BetType, &b.Stake, &b.PotentialReturn,
			&b.ActualReturn, &b.Status, &b.PlacedAt, &b.SettledAt); err != nil {
			return nil, err
		}
		bets = append(bets, b)
	}

	return &BetListResult{
		Bets:  bets,
		Total: total,
		Page:  filters.Page,
		Limit: filters.Limit,
	}, nil
}

func (r *BetRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, actualReturn float64) error {
	_, err := r.db.Exec(ctx,
		`UPDATE bets SET status = $2, actual_return = $3, settled_at = NOW() WHERE id = $1`,
		id, status, actualReturn)
	return err
}

func (r *BetRepository) UpdateLegResult(ctx context.Context, legID uuid.UUID, result string) error {
	_, err := r.db.Exec(ctx, `UPDATE bet_legs SET result = $2 WHERE id = $1`, legID, result)
	return err
}

func (r *BetRepository) GetPendingBetsByMarket(ctx context.Context, marketID uuid.UUID) ([]models.Bet, error) {
	query := `SELECT DISTINCT b.id, b.user_id, b.bet_type, b.stake, b.potential_return, b.actual_return, b.status, b.placed_at, b.settled_at
	          FROM bets b
	          JOIN bet_legs bl ON b.id = bl.bet_id
	          WHERE bl.market_id = $1 AND b.status = 'pending'`

	rows, err := r.db.Query(ctx, query, marketID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bets []models.Bet
	for rows.Next() {
		var b models.Bet
		if err := rows.Scan(&b.ID, &b.UserID, &b.BetType, &b.Stake, &b.PotentialReturn,
			&b.ActualReturn, &b.Status, &b.PlacedAt, &b.SettledAt); err != nil {
			return nil, err
		}
		bets = append(bets, b)
	}
	return bets, nil
}

type DashboardStats struct {
	TotalUsers     int     `json:"total_users"`
	TotalBets      int     `json:"total_bets"`
	PendingBets    int     `json:"pending_bets"`
	TotalStaked    float64 `json:"total_staked"`
	TotalPaidOut   float64 `json:"total_paid_out"`
	ActiveEvents   int     `json:"active_events"`
	TotalDeposited float64 `json:"total_deposited"`
}

func (r *BetRepository) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	stats := &DashboardStats{}

	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&stats.TotalUsers)
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM bets`).Scan(&stats.TotalBets)
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM bets WHERE status = 'pending'`).Scan(&stats.PendingBets)
	_ = r.db.QueryRow(ctx, `SELECT COALESCE(SUM(stake), 0) FROM bets`).Scan(&stats.TotalStaked)
	_ = r.db.QueryRow(ctx, `SELECT COALESCE(SUM(actual_return), 0) FROM bets WHERE status = 'won'`).Scan(&stats.TotalPaidOut)
	_ = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM events WHERE status IN ('upcoming', 'live')`).Scan(&stats.ActiveEvents)
	_ = r.db.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM transactions WHERE type = 'deposit'`).Scan(&stats.TotalDeposited)

	return stats, nil
}
