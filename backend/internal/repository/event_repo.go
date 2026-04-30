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

type EventRepository struct {
	db *pgxpool.Pool
}

func NewEventRepository(db *pgxpool.Pool) *EventRepository {
	return &EventRepository{db: db}
}

type EventFilters struct {
	SportID  *int
	Status   *string
	Search   *string
	DateFrom *string
	DateTo   *string
	Page     int
	Limit    int
}

type EventListResult struct {
	Events []models.Event `json:"events"`
	Total  int            `json:"total"`
	Page   int            `json:"page"`
	Limit  int            `json:"limit"`
}

func (r *EventRepository) List(ctx context.Context, filters EventFilters) (*EventListResult, error) {
	if filters.Page < 1 {
		filters.Page = 1
	}
	if filters.Limit < 1 || filters.Limit > 100 {
		filters.Limit = 20
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.SportID != nil {
		conditions = append(conditions, fmt.Sprintf("e.sport_id = $%d", argIdx))
		args = append(args, *filters.SportID)
		argIdx++
	}
	if filters.Status != nil && *filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("e.status = $%d", argIdx))
		args = append(args, *filters.Status)
		argIdx++
	}
	if filters.Search != nil && *filters.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(LOWER(e.home_team) LIKE $%d OR LOWER(e.away_team) LIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(*filters.Search)+"%")
		argIdx++
	}
	if filters.DateFrom != nil && *filters.DateFrom != "" {
		conditions = append(conditions, fmt.Sprintf("e.start_time >= $%d", argIdx))
		args = append(args, *filters.DateFrom)
		argIdx++
	}
	if filters.DateTo != nil && *filters.DateTo != "" {
		conditions = append(conditions, fmt.Sprintf("e.start_time <= $%d", argIdx))
		args = append(args, *filters.DateTo)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM events e %s`, whereClause)
	var total int
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	// Fetch events
	offset := (filters.Page - 1) * filters.Limit
	query := fmt.Sprintf(`
		SELECT e.id, e.sport_id, e.home_team, e.away_team, e.start_time, e.status,
			   e.final_score_home, e.final_score_away, e.created_at, e.updated_at,
			   s.name, s.slug, s.icon,
			   (SELECT COUNT(*) FROM markets m WHERE m.event_id = e.id) as markets_count
		FROM events e
		JOIN sports s ON e.sport_id = s.id
		%s
		ORDER BY e.start_time ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filters.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []models.Event
	for rows.Next() {
		var e models.Event
		err := rows.Scan(
			&e.ID, &e.SportID, &e.HomeTeam, &e.AwayTeam, &e.StartTime, &e.Status,
			&e.FinalScoreHome, &e.FinalScoreAway, &e.CreatedAt, &e.UpdatedAt,
			&e.SportName, &e.SportSlug, &e.SportIcon, &e.MarketsCount,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return &EventListResult{
		Events: events,
		Total:  total,
		Page:   filters.Page,
		Limit:  filters.Limit,
	}, nil
}

func (r *EventRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Event, error) {
	event := &models.Event{}
	query := `
		SELECT e.id, e.sport_id, e.home_team, e.away_team, e.start_time, e.status,
			   e.final_score_home, e.final_score_away, e.created_at, e.updated_at,
			   s.name, s.slug, s.icon,
			   (SELECT COUNT(*) FROM markets m WHERE m.event_id = e.id) as markets_count
		FROM events e
		JOIN sports s ON e.sport_id = s.id
		WHERE e.id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&event.ID, &event.SportID, &event.HomeTeam, &event.AwayTeam, &event.StartTime, &event.Status,
		&event.FinalScoreHome, &event.FinalScoreAway, &event.CreatedAt, &event.UpdatedAt,
		&event.SportName, &event.SportSlug, &event.SportIcon, &event.MarketsCount,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return event, nil
}

func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	query := `
		INSERT INTO events (sport_id, home_team, away_team, start_time, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		event.SportID, event.HomeTeam, event.AwayTeam, event.StartTime, event.Status,
	).Scan(&event.ID, &event.CreatedAt, &event.UpdatedAt)
}

func (r *EventRepository) Update(ctx context.Context, event *models.Event) error {
	query := `
		UPDATE events SET home_team = $2, away_team = $3, start_time = $4, status = $5,
		       final_score_home = $6, final_score_away = $7
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		event.ID, event.HomeTeam, event.AwayTeam, event.StartTime, event.Status,
		event.FinalScoreHome, event.FinalScoreAway,
	)
	return err
}

func (r *EventRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	_, err := r.db.Exec(ctx, `UPDATE events SET status = $2 WHERE id = $1`, id, status)
	return err
}

func (r *EventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE events SET status = 'cancelled' WHERE id = $1`, id)
	return err
}
