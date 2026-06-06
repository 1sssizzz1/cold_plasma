package postgres

import (
	"context"
	"errors"
	"fmt"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"

	"github.com/jackc/pgx/v5"
)

type ProcedureRepo struct {
	db db
}

func NewProcedureRepo(d db) *ProcedureRepo {
	return &ProcedureRepo{db: d}
}

func (r *ProcedureRepo) ListActive(ctx context.Context) ([]models.Procedure, error) {
	const q = `
		SELECT id, title, description, duration_mins, price, bonus_earned, category, image_url, video_url, services, duration_str, popular, is_active, created_at, updated_at
		FROM procedures
		WHERE is_active = TRUE AND deleted_at IS NULL
		ORDER BY popular DESC, id ASC
	`
	return r.list(ctx, q)
}

func (r *ProcedureRepo) ListAll(ctx context.Context) ([]models.Procedure, error) {
	const q = `
		SELECT id, title, description, duration_mins, price, bonus_earned, category, image_url, video_url, services, duration_str, popular, is_active, created_at, updated_at
		FROM procedures
		WHERE deleted_at IS NULL
		ORDER BY id DESC
	`
	return r.list(ctx, q)
}

func (r *ProcedureRepo) list(ctx context.Context, q string) ([]models.Procedure, error) {
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("list procedures: %w", err)
	}
	defer rows.Close()

	out := make([]models.Procedure, 0)
	for rows.Next() {
		var p models.Procedure
		if err := rows.Scan(
			&p.ID,
			&p.Title,
			&p.Description,
			&p.DurationMins,
			&p.Price,
			&p.BonusEarned,
			&p.Category,
			&p.ImageURL,
			&p.VideoURL,
			&p.Services,
			&p.DurationStr,
			&p.Popular,
			&p.IsActive,
			&p.CreatedAt,
			&p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan procedure: %w", err)
		}
		out = append(out, p)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list procedures: %w", rows.Err())
	}
	return out, nil
}

func (r *ProcedureRepo) GetByID(ctx context.Context, id int64) (models.Procedure, error) {
	const q = `
		SELECT id, title, description, duration_mins, price, bonus_earned, category, image_url, video_url, services, duration_str, popular, is_active, created_at, updated_at
		FROM procedures
		WHERE id = $1 AND deleted_at IS NULL
	`
	var p models.Procedure
	err := r.db.QueryRow(ctx, q, id).Scan(
		&p.ID,
		&p.Title,
		&p.Description,
		&p.DurationMins,
		&p.Price,
		&p.BonusEarned,
		&p.Category,
		&p.ImageURL,
		&p.VideoURL,
		&p.Services,
		&p.DurationStr,
		&p.Popular,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Procedure{}, models.ErrNotFound
		}
		return models.Procedure{}, fmt.Errorf("get procedure: %w", err)
	}
	return p, nil
}

func (r *ProcedureRepo) Create(ctx context.Context, p repository.ProcedureParams) (models.Procedure, error) {
	const q = `
		INSERT INTO procedures (title, description, duration_mins, price, bonus_earned, category, image_url, video_url, services, duration_str, popular, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		RETURNING id
	`
	var id int64
	if err := r.db.QueryRow(ctx, q, p.Title, p.Description, p.DurationMins, p.Price, p.BonusEarned, p.Category, p.ImageURL, p.VideoURL, p.Services, p.DurationStr, p.Popular, p.IsActive).Scan(&id); err != nil {
		return models.Procedure{}, fmt.Errorf("create procedure: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *ProcedureRepo) Update(ctx context.Context, id int64, p repository.ProcedureParams) (models.Procedure, error) {
	const q = `
		UPDATE procedures
		SET title=$2, description=$3, duration_mins=$4, price=$5, bonus_earned=$6,
		    category=$7, image_url=$8, video_url=$9, services=$10, duration_str=$11,
		    popular=$12, is_active=$13, updated_at=now()
		WHERE id=$1 AND deleted_at IS NULL
	`
	ct, err := r.db.Exec(ctx, q, id, p.Title, p.Description, p.DurationMins, p.Price, p.BonusEarned, p.Category, p.ImageURL, p.VideoURL, p.Services, p.DurationStr, p.Popular, p.IsActive)
	if err != nil {
		return models.Procedure{}, fmt.Errorf("update procedure: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.Procedure{}, models.ErrNotFound
	}
	return r.GetByID(ctx, id)
}

func (r *ProcedureRepo) Delete(ctx context.Context, id int64) error {
	const q = `
		UPDATE procedures
		SET is_active = FALSE, deleted_at = now(), updated_at = now()
		WHERE id = $1 AND deleted_at IS NULL
	`
	ct, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete procedure: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}
