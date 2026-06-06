package postgres

import (
	"context"
	"fmt"
	"time"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
)

type ChatRepo struct {
	db db
}

func NewChatRepo(d db) *ChatRepo {
	return &ChatRepo{db: d}
}

func (r *ChatRepo) UpsertSession(ctx context.Context, p repository.UpsertChatSessionParams) error {
	const q = `
		INSERT INTO ai_chat_sessions (id, user_id, user_name, user_email, user_phone, opened_at, closed_at)
		VALUES ($1, $2, $3, $4, $5, now(), NULL)
		ON CONFLICT (id) DO UPDATE SET
			user_id = COALESCE(EXCLUDED.user_id, ai_chat_sessions.user_id),
			user_name = EXCLUDED.user_name,
			user_email = EXCLUDED.user_email,
			user_phone = EXCLUDED.user_phone,
			closed_at = NULL
	`
	_, err := r.db.Exec(ctx, q, p.ID, p.UserID, p.UserName, p.UserEmail, p.UserPhone)
	if err != nil {
		return fmt.Errorf("upsert ai chat session: %w", err)
	}
	return nil
}

func (r *ChatRepo) CloseSession(ctx context.Context, sessionID string) error {
	const q = `UPDATE ai_chat_sessions SET closed_at = COALESCE(closed_at, now()) WHERE id = $1`
	_, err := r.db.Exec(ctx, q, sessionID)
	if err != nil {
		return fmt.Errorf("close ai chat session: %w", err)
	}
	return nil
}

func (r *ChatRepo) DeleteSession(ctx context.Context, sessionID string) error {
	if _, err := r.db.Exec(ctx, `DELETE FROM chat_logs WHERE session_id = $1`, sessionID); err != nil {
		return fmt.Errorf("delete chat session logs: %w", err)
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM ai_chat_sessions WHERE id = $1`, sessionID); err != nil {
		return fmt.Errorf("delete ai chat session: %w", err)
	}
	return nil
}

func (r *ChatRepo) DeleteOlderThan(ctx context.Context, cutoff time.Time) error {
	const oldSessions = `
		SELECT id
		FROM ai_chat_sessions
		WHERE COALESCE(closed_at, opened_at) < $1
	`
	if _, err := r.db.Exec(ctx, `DELETE FROM chat_logs WHERE created_at < $1 OR session_id IN (`+oldSessions+`)`, cutoff); err != nil {
		return fmt.Errorf("delete old chat logs: %w", err)
	}
	if _, err := r.db.Exec(ctx, `DELETE FROM ai_chat_sessions WHERE COALESCE(closed_at, opened_at) < $1`, cutoff); err != nil {
		return fmt.Errorf("delete old ai chat sessions: %w", err)
	}
	return nil
}

func (r *ChatRepo) AddChatLog(ctx context.Context, p repository.CreateChatLogParams) error {
	const q = `
		INSERT INTO chat_logs (session_id, user_id, user_name, user_email, user_phone, raw_input, raw_output, ai_model, intent)
		VALUES (NULLIF($1, ''), $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := r.db.Exec(ctx, q, p.SessionID, p.UserID, p.UserName, p.UserEmail, p.UserPhone, p.RawInput, p.RawOutput, p.AIModel, p.Intent)
	if err != nil {
		return fmt.Errorf("insert chat log: %w", err)
	}
	return nil
}

func (r *ChatRepo) ListByUserID(ctx context.Context, userID int64, limit int) ([]models.ChatLog, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	const q = `
		SELECT id, COALESCE(session_id, ''), user_id, user_name, user_email, user_phone, raw_input, raw_output, ai_model, intent, created_at
		FROM chat_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list chat logs: %w", err)
	}
	defer rows.Close()

	out := make([]models.ChatLog, 0)
	for rows.Next() {
		var l models.ChatLog
		if err := rows.Scan(
			&l.ID,
			&l.SessionID,
			&l.UserID,
			&l.UserName,
			&l.UserEmail,
			&l.UserPhone,
			&l.RawInput,
			&l.RawOutput,
			&l.AIModel,
			&l.Intent,
			&l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat log: %w", err)
		}
		out = append(out, l)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list chat logs: %w", rows.Err())
	}
	return out, nil
}

func (r *ChatRepo) ListAll(ctx context.Context, limit int) ([]models.ChatLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	const q = `
		SELECT id, COALESCE(session_id, ''), user_id, user_name, user_email, user_phone, raw_input, raw_output, ai_model, intent, created_at
		FROM chat_logs
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list all chat logs: %w", err)
	}
	defer rows.Close()

	out := make([]models.ChatLog, 0)
	for rows.Next() {
		var l models.ChatLog
		if err := rows.Scan(
			&l.ID,
			&l.SessionID,
			&l.UserID,
			&l.UserName,
			&l.UserEmail,
			&l.UserPhone,
			&l.RawInput,
			&l.RawOutput,
			&l.AIModel,
			&l.Intent,
			&l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan chat log: %w", err)
		}
		out = append(out, l)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list all chat logs: %w", rows.Err())
	}
	return out, nil
}
