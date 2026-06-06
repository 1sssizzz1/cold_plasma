package postgres

import (
	"context"
	"fmt"

	"cold-plasma-server/internal/repository"
)

type SettingsRepo struct {
	db db
}

func NewSettingsRepo(d db) *SettingsRepo {
	return &SettingsRepo{db: d}
}

var _ repository.SettingsRepository = (*SettingsRepo)(nil)

func (r *SettingsRepo) GetSettings(ctx context.Context, keys []string) (map[string]string, error) {
	out := make(map[string]string, len(keys))
	for _, key := range keys {
		out[key] = ""
	}
	if len(keys) == 0 {
		return out, nil
	}
	rows, err := r.db.Query(ctx, `SELECT key, value FROM app_settings WHERE key = ANY($1)`, keys)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("scan setting: %w", err)
		}
		out[key] = value
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("get settings: %w", rows.Err())
	}
	return out, nil
}

func (r *SettingsRepo) UpsertSettings(ctx context.Context, values map[string]string) error {
	for key, value := range values {
		if _, err := r.db.Exec(ctx, `
			INSERT INTO app_settings (key, value, updated_at)
			VALUES ($1, $2, now())
			ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()
		`, key, value); err != nil {
			return fmt.Errorf("upsert setting %s: %w", key, err)
		}
	}
	return nil
}
