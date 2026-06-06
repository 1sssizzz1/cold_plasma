package postgres

import (
	"context"
	"errors"
	"fmt"

	"cold-plasma-server/internal/models"

	"github.com/jackc/pgx/v5"
)

type ProcedureReviewRepo struct {
	db db
}

func NewProcedureReviewRepo(d db) *ProcedureReviewRepo {
	return &ProcedureReviewRepo{db: d}
}

func (r *ProcedureReviewRepo) ListPublic(ctx context.Context, procedureID int64, limit int) ([]models.ProcedureReview, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	q := `
		SELECT r.id, r.user_id, u.name, r.procedure_id, p.title, r.rating, r.text, r.created_at, r.updated_at
		FROM procedure_reviews r
		JOIN users u ON u.id = r.user_id
		JOIN procedures p ON p.id = r.procedure_id
	`
	args := []any{}
	if procedureID > 0 {
		q += ` WHERE r.procedure_id = $1 ORDER BY r.created_at DESC LIMIT $2`
		args = append(args, procedureID, limit)
	} else {
		q += ` ORDER BY r.created_at DESC LIMIT $1`
		args = append(args, limit)
	}
	return r.list(ctx, q, args...)
}

func (r *ProcedureReviewRepo) ListAll(ctx context.Context, limit int) ([]models.ProcedureReview, error) {
	return r.ListPublic(ctx, 0, limit)
}

func (r *ProcedureReviewRepo) CanUserReview(ctx context.Context, userID int64, procedureID int64) (bool, error) {
	const q = `
		SELECT EXISTS (
			SELECT 1 FROM bookings
			WHERE user_id = $1 AND procedure_id = $2 AND status = 'completed'
		)
	`
	var ok bool
	if err := r.db.QueryRow(ctx, q, userID, procedureID).Scan(&ok); err != nil {
		return false, fmt.Errorf("check review permission: %w", err)
	}
	return ok, nil
}

func (r *ProcedureReviewRepo) Upsert(ctx context.Context, userID int64, procedureID int64, rating int, text string) (models.ProcedureReview, error) {
	const q = `
		INSERT INTO procedure_reviews (user_id, procedure_id, rating, text)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, procedure_id)
		DO UPDATE SET rating = EXCLUDED.rating, text = EXCLUDED.text, updated_at = now()
		RETURNING id
	`
	var id int64
	if err := r.db.QueryRow(ctx, q, userID, procedureID, rating, text).Scan(&id); err != nil {
		return models.ProcedureReview{}, fmt.Errorf("upsert procedure review: %w", err)
	}
	items, err := r.list(ctx, `
		SELECT r.id, r.user_id, u.name, r.procedure_id, p.title, r.rating, r.text, r.created_at, r.updated_at
		FROM procedure_reviews r
		JOIN users u ON u.id = r.user_id
		JOIN procedures p ON p.id = r.procedure_id
		WHERE r.id = $1
	`, id)
	if err != nil {
		return models.ProcedureReview{}, err
	}
	if len(items) == 0 {
		return models.ProcedureReview{}, models.ErrNotFound
	}
	return items[0], nil
}

func (r *ProcedureReviewRepo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM procedure_reviews WHERE id = $1`
	ct, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete procedure review: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *ProcedureReviewRepo) list(ctx context.Context, q string, args ...any) ([]models.ProcedureReview, error) {
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list procedure reviews: %w", err)
	}
	defer rows.Close()
	out := make([]models.ProcedureReview, 0)
	for rows.Next() {
		var item models.ProcedureReview
		if err := rows.Scan(&item.ID, &item.UserID, &item.UserName, &item.ProcedureID, &item.ProcedureTitle, &item.Rating, &item.Text, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan procedure review: %w", err)
		}
		out = append(out, item)
	}
	if rows.Err() != nil && !errors.Is(rows.Err(), pgx.ErrNoRows) {
		return nil, fmt.Errorf("list procedure reviews: %w", rows.Err())
	}
	return out, nil
}
