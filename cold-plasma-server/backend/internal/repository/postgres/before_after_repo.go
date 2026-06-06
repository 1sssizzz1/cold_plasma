package postgres

import (
	"context"
	"errors"
	"fmt"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"

	"github.com/jackc/pgx/v5"
)

type BeforeAfterRepo struct {
	db db
}

func NewBeforeAfterRepo(d db) *BeforeAfterRepo {
	return &BeforeAfterRepo{db: d}
}

func (r *BeforeAfterRepo) ListPublic(ctx context.Context, limit int) ([]models.BeforeAfterResult, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	const q = `
		SELECT id, procedure_id, procedure, title, description, before_url, after_url, is_featured, sort_order, is_active, created_at, updated_at
		FROM before_after_results
		WHERE is_active = true
		ORDER BY sort_order ASC, created_at DESC
		LIMIT $1
	`
	return r.list(ctx, q, limit)
}

func (r *BeforeAfterRepo) ListAll(ctx context.Context, limit int) ([]models.BeforeAfterResult, error) {
	if limit <= 0 || limit > 300 {
		limit = 100
	}
	const q = `
		SELECT id, procedure_id, procedure, title, description, before_url, after_url, is_featured, sort_order, is_active, created_at, updated_at
		FROM before_after_results
		ORDER BY sort_order ASC, created_at DESC
		LIMIT $1
	`
	return r.list(ctx, q, limit)
}

func (r *BeforeAfterRepo) Create(ctx context.Context, p repository.BeforeAfterParams) (models.BeforeAfterResult, error) {
	const q = `
		INSERT INTO before_after_results (procedure_id, procedure, title, description, before_url, after_url, is_featured, sort_order, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, procedure_id, procedure, title, description, before_url, after_url, is_featured, sort_order, is_active, created_at, updated_at
	`
	var out models.BeforeAfterResult
	if err := r.db.QueryRow(ctx, q, p.ProcedureID, p.Procedure, p.Title, p.Description, p.BeforeURL, p.AfterURL, p.IsFeatured, p.SortOrder, p.IsActive).Scan(beforeAfterScanDest(&out)...); err != nil {
		return models.BeforeAfterResult{}, fmt.Errorf("create before after result: %w", err)
	}
	return out, nil
}

func (r *BeforeAfterRepo) Update(ctx context.Context, id int64, p repository.BeforeAfterParams) (models.BeforeAfterResult, error) {
	const q = `
		UPDATE before_after_results
		SET procedure_id = $2,
			procedure = $3,
			title = $4,
			description = $5,
			before_url = $6,
			after_url = $7,
			is_featured = $8,
			sort_order = $9,
			is_active = $10,
			updated_at = now()
		WHERE id = $1
		RETURNING id, procedure_id, procedure, title, description, before_url, after_url, is_featured, sort_order, is_active, created_at, updated_at
	`
	var out models.BeforeAfterResult
	err := r.db.QueryRow(ctx, q, id, p.ProcedureID, p.Procedure, p.Title, p.Description, p.BeforeURL, p.AfterURL, p.IsFeatured, p.SortOrder, p.IsActive).Scan(beforeAfterScanDest(&out)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.BeforeAfterResult{}, models.ErrNotFound
		}
		return models.BeforeAfterResult{}, fmt.Errorf("update before after result: %w", err)
	}
	return out, nil
}

func (r *BeforeAfterRepo) Delete(ctx context.Context, id int64) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM before_after_results WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete before after result: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *BeforeAfterRepo) list(ctx context.Context, query string, args ...any) ([]models.BeforeAfterResult, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list before after results: %w", err)
	}
	defer rows.Close()

	out := make([]models.BeforeAfterResult, 0)
	for rows.Next() {
		var item models.BeforeAfterResult
		if err := rows.Scan(beforeAfterScanDest(&item)...); err != nil {
			return nil, fmt.Errorf("scan before after result: %w", err)
		}
		out = append(out, item)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list before after results: %w", rows.Err())
	}
	return out, nil
}

func beforeAfterScanDest(item *models.BeforeAfterResult) []any {
	return []any{
		&item.ID,
		&item.ProcedureID,
		&item.Procedure,
		&item.Title,
		&item.Description,
		&item.BeforeURL,
		&item.AfterURL,
		&item.IsFeatured,
		&item.SortOrder,
		&item.IsActive,
		&item.CreatedAt,
		&item.UpdatedAt,
	}
}
