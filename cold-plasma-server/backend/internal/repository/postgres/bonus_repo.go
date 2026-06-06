package postgres

import (
	"context"
	"fmt"

	"cold-plasma-server/internal/models"
)

type BonusRepo struct {
	db db
}

func NewBonusRepo(d db) *BonusRepo {
	return &BonusRepo{db: d}
}

func (r *BonusRepo) AddLog(ctx context.Context, userID int64, typ string, amount int, comment string) error {
	const q = `INSERT INTO bonus_logs (user_id, type, amount, comment) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(ctx, q, userID, typ, amount, comment)
	if err != nil {
		return fmt.Errorf("insert bonus log: %w", err)
	}
	return nil
}

func (r *BonusRepo) ListLogsByUserID(ctx context.Context, userID int64, limit int) ([]models.BonusLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
		SELECT id, user_id, type, amount, comment, created_at
		FROM bonus_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list bonus logs: %w", err)
	}
	defer rows.Close()

	out := make([]models.BonusLog, 0)
	for rows.Next() {
		var l models.BonusLog
		if err := rows.Scan(&l.ID, &l.UserID, &l.Type, &l.Amount, &l.Comment, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan bonus log: %w", err)
		}
		out = append(out, l)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list bonus logs: %w", rows.Err())
	}
	return out, nil
}
