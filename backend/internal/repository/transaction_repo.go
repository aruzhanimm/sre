package repository

import (
	"context"
	"fmt"
	"strings"

	"BetKZ/internal/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	query := `INSERT INTO transactions (user_id, type, amount, balance_before, balance_after, reference_id, status)
	          VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at`

	return r.db.QueryRow(ctx, query,
		tx.UserID, tx.Type, tx.Amount, tx.BalanceBefore, tx.BalanceAfter, tx.ReferenceID, tx.Status,
	).Scan(&tx.ID, &tx.CreatedAt)
}

type TxFilters struct {
	UserID *uuid.UUID
	Type   *string
	Page   int
	Limit  int
}

type TxListResult struct {
	Transactions []models.Transaction `json:"transactions"`
	Total        int                  `json:"total"`
	Page         int                  `json:"page"`
	Limit        int                  `json:"limit"`
}

func (r *TransactionRepository) List(ctx context.Context, filters TxFilters) (*TxListResult, error) {
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
	if filters.Type != nil {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIdx))
		args = append(args, *filters.Type)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM transactions %s`, whereClause)
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, err
	}

	offset := (filters.Page - 1) * filters.Limit
	query := fmt.Sprintf(`SELECT id, user_id, type, amount, balance_before, balance_after, reference_id, status, created_at
	          FROM transactions %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
	args = append(args, filters.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		if err := rows.Scan(&t.ID, &t.UserID, &t.Type, &t.Amount, &t.BalanceBefore,
			&t.BalanceAfter, &t.ReferenceID, &t.Status, &t.CreatedAt); err != nil {
			return nil, err
		}
		transactions = append(transactions, t)
	}

	return &TxListResult{
		Transactions: transactions,
		Total:        total,
		Page:         filters.Page,
		Limit:        filters.Limit,
	}, nil
}
