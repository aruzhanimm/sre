package repository

import (
	"context"

	"BetKZ/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SportRepository struct {
	db *pgxpool.Pool
}

func NewSportRepository(db *pgxpool.Pool) *SportRepository {
	return &SportRepository{db: db}
}

func (r *SportRepository) List(ctx context.Context) ([]models.Sport, error) {
	query := `SELECT id, name, slug, icon, is_active FROM sports WHERE is_active = true ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sports []models.Sport
	for rows.Next() {
		var s models.Sport
		if err := rows.Scan(&s.ID, &s.Name, &s.Slug, &s.Icon, &s.IsActive); err != nil {
			return nil, err
		}
		sports = append(sports, s)
	}
	return sports, nil
}

func (r *SportRepository) GetByID(ctx context.Context, id int) (*models.Sport, error) {
	sport := &models.Sport{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, slug, icon, is_active FROM sports WHERE id = $1`, id,
	).Scan(&sport.ID, &sport.Name, &sport.Slug, &sport.Icon, &sport.IsActive)
	if err != nil {
		return nil, err
	}
	return sport, nil
}
