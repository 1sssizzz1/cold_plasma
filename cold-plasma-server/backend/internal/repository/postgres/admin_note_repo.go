package postgres

import (
	"context"
	"fmt"
	"time"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
)

type AdminNoteRepo struct {
	db db
}

func NewAdminNoteRepo(d db) *AdminNoteRepo {
	return &AdminNoteRepo{db: d}
}

var _ repository.AdminNoteRepository = (*AdminNoteRepo)(nil)

func (r *AdminNoteRepo) Create(ctx context.Context, p repository.CreateAdminNoteParams) (models.AdminNote, error) {
	const q = `
		INSERT INTO admin_notes (start_at, end_at, title)
		VALUES ($1, $2, $3)
		RETURNING id, start_at, end_at, title, created_at
	`
	var out models.AdminNote
	if err := r.db.QueryRow(ctx, q, p.StartAt, p.EndAt, p.Title).Scan(&out.ID, &out.StartAt, &out.EndAt, &out.Title, &out.CreatedAt); err != nil {
		return models.AdminNote{}, fmt.Errorf("create admin note: %w", err)
	}
	return out, nil
}

func (r *AdminNoteRepo) ListBetween(ctx context.Context, from, to time.Time) ([]models.AdminNote, error) {
	const q = `
		SELECT id, start_at, end_at, title, created_at
		FROM admin_notes
		WHERE start_at < $2 AND end_at > $1
		ORDER BY start_at ASC
	`
	rows, err := r.db.Query(ctx, q, from, to)
	if err != nil {
		return nil, fmt.Errorf("list admin notes: %w", err)
	}
	defer rows.Close()

	out := make([]models.AdminNote, 0)
	for rows.Next() {
		var n models.AdminNote
		if err := rows.Scan(&n.ID, &n.StartAt, &n.EndAt, &n.Title, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan admin note: %w", err)
		}
		out = append(out, n)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list admin notes: %w", rows.Err())
	}
	return out, nil
}

func (r *AdminNoteRepo) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM admin_notes WHERE id = $1`
	ct, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete admin note: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}
