package repository

import (
	"context"
	"errors"

	"BetKZ/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (email, password_hash, balance, role)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRow(ctx, query,
		user.Email, user.PasswordHash, user.Balance, user.Role,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, password_hash, balance, role, created_at, updated_at FROM users WHERE id = $1`

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Balance,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	query := `SELECT id, email, password_hash, balance, role, created_at, updated_at FROM users WHERE email = $1`

	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.Balance,
		&user.Role, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdateBalance(ctx context.Context, id uuid.UUID, amount float64) error {
	query := `UPDATE users SET balance = balance + $2 WHERE id = $1 AND balance + $2 >= 0`
	result, err := r.db.Exec(ctx, query, id, amount)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("insufficient balance or user not found")
	}
	return nil
}

func (r *UserRepository) GetBalance(ctx context.Context, id uuid.UUID) (float64, error) {
	var balance float64
	err := r.db.QueryRow(ctx, `SELECT balance FROM users WHERE id = $1`, id).Scan(&balance)
	return balance, err
}
