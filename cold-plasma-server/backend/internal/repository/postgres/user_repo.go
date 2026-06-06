package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type UserRepo struct {
	db db
}

const userColumns = `
	id, email, name, phone, password_hash, photo_url, bonus_points,
	email_verified, phone_verified, is_blocked, is_admin,
	email_verification_sent_at, email_verified_at, phone_verified_at,
	vk_id, auth_provider, telegram_chat_id, telegram_username, telegram_linked_at,
	created_at, updated_at
`

func NewUserRepo(d db) *UserRepo {
	return &UserRepo{db: d}
}

func (r *UserRepo) Create(ctx context.Context, u repository.CreateUserParams) (models.User, error) {
	if u.AuthProvider == "" {
		u.AuthProvider = "password"
	}
	const q = `
		INSERT INTO users (email, name, phone, password_hash, email_verified, phone_verified, phone_verified_at, vk_id, auth_provider)
		VALUES ($1, $2, $3, $4, $5, $6, CASE WHEN $6 THEN now() ELSE NULL END, $7, $8)
		RETURNING ` + userColumns + `
	`
	var out models.User
	err := r.db.QueryRow(ctx, q, u.Email, u.Name, u.Phone, u.PasswordHash, u.EmailVerified, u.PhoneVerified, u.VKID, u.AuthProvider).Scan(userScanDest(&out)...)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, fmt.Errorf("user already exists: %w", err)
		}
		return models.User{}, fmt.Errorf("create user: %w", err)
	}
	return out, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (models.User, error) {
	const q = `
		SELECT ` + userColumns + `
		FROM users
		WHERE email = $1
	`
	var out models.User
	err := r.db.QueryRow(ctx, q, email).Scan(userScanDest(&out)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrNotFound
		}
		return models.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return out, nil
}

func (r *UserRepo) GetByPhone(ctx context.Context, phone string) (models.User, error) {
	const q = `
		SELECT ` + userColumns + `
		FROM users
		WHERE phone = $1
	`
	var out models.User
	err := r.db.QueryRow(ctx, q, phone).Scan(userScanDest(&out)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrNotFound
		}
		return models.User{}, fmt.Errorf("get user by phone: %w", err)
	}
	return out, nil
}

func (r *UserRepo) GetByVKID(ctx context.Context, vkID string) (models.User, error) {
	const q = `
		SELECT ` + userColumns + `
		FROM users
		WHERE vk_id = $1
	`
	var out models.User
	err := r.db.QueryRow(ctx, q, vkID).Scan(userScanDest(&out)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrNotFound
		}
		return models.User{}, fmt.Errorf("get user by vk_id: %w", err)
	}
	return out, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (models.User, error) {
	const q = `
		SELECT ` + userColumns + `
		FROM users
		WHERE id = $1
	`
	var out models.User
	err := r.db.QueryRow(ctx, q, id).Scan(userScanDest(&out)...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, models.ErrNotFound
		}
		return models.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return out, nil
}

func (r *UserRepo) LinkVK(ctx context.Context, userID int64, vkID, email, name string) error {
	const q = `
		UPDATE users
		SET vk_id = $2,
		    auth_provider = CASE WHEN auth_provider = '' OR auth_provider = 'password' THEN 'vk' ELSE auth_provider END,
		    email = CASE WHEN email = '' THEN $3 ELSE email END,
		    email_verified = CASE WHEN email = '' AND $3 <> '' THEN TRUE ELSE email_verified END,
		    name = CASE WHEN name = '' THEN $4 ELSE name END,
		    updated_at = now()
		WHERE id = $1
	`
	ct, err := r.db.Exec(ctx, q, userID, vkID, email, name)
	if err != nil {
		return fmt.Errorf("link vk: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) UpdatePhoneUnverified(ctx context.Context, userID int64, phone string) error {
	const q = `
		UPDATE users
		SET phone = $2,
		    phone_verified = FALSE,
		    phone_verified_at = NULL,
		    updated_at = now()
		WHERE id = $1
	`
	ct, err := r.db.Exec(ctx, q, userID, phone)
	if err != nil {
		return fmt.Errorf("update phone: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) UpdateBonusPoints(ctx context.Context, userID int64, newBalance int) error {
	const q = `UPDATE users SET bonus_points = $2, updated_at = now() WHERE id = $1`
	ct, err := r.db.Exec(ctx, q, userID, newBalance)
	if err != nil {
		return fmt.Errorf("update bonus_points: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) MarkEmailVerified(ctx context.Context, userID int64) error {
	const q = `
		UPDATE users
		SET email_verified = TRUE, email_verified_at = now(), updated_at = now()
		WHERE id = $1
	`
	ct, err := r.db.Exec(ctx, q, userID)
	if err != nil {
		return fmt.Errorf("mark email verified: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) MarkPhoneVerified(ctx context.Context, userID int64, phone string) error {
	const q = `
		UPDATE users
		SET phone = $2, phone_verified = TRUE, phone_verified_at = now(), updated_at = now()
		WHERE id = $1
	`
	ct, err := r.db.Exec(ctx, q, userID, phone)
	if err != nil {
		return fmt.Errorf("mark phone verified: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) UpdateEmailVerificationSentAt(ctx context.Context, userID int64, t time.Time) error {
	const q = `
		UPDATE users
		SET email_verification_sent_at = $2, updated_at = now()
		WHERE id = $1
	`
	ct, err := r.db.Exec(ctx, q, userID, t)
	if err != nil {
		return fmt.Errorf("update email_verification_sent_at: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) CreatePhoneVerificationCode(ctx context.Context, p repository.CreatePhoneVerificationCodeParams) (models.PhoneVerificationCode, error) {
	const q = `
		INSERT INTO phone_verification_codes (user_id, phone, code_hash, expires_at, attempts_left)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, phone, code_hash, expires_at, attempts_left, consumed_at, created_at
	`
	var out models.PhoneVerificationCode
	err := r.db.QueryRow(ctx, q, p.UserID, p.Phone, p.CodeHash, p.ExpiresAt, p.AttemptsLeft).Scan(
		&out.ID,
		&out.UserID,
		&out.Phone,
		&out.CodeHash,
		&out.ExpiresAt,
		&out.AttemptsLeft,
		&out.ConsumedAt,
		&out.CreatedAt,
	)
	if err != nil {
		return models.PhoneVerificationCode{}, fmt.Errorf("create phone verification code: %w", err)
	}
	return out, nil
}

func (r *UserRepo) GetActivePhoneVerificationCode(ctx context.Context, userID int64, phone string) (models.PhoneVerificationCode, error) {
	const q = `
		SELECT id, user_id, phone, code_hash, expires_at, attempts_left, consumed_at, created_at
		FROM phone_verification_codes
		WHERE user_id = $1 AND phone = $2 AND consumed_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	var out models.PhoneVerificationCode
	err := r.db.QueryRow(ctx, q, userID, phone).Scan(
		&out.ID,
		&out.UserID,
		&out.Phone,
		&out.CodeHash,
		&out.ExpiresAt,
		&out.AttemptsLeft,
		&out.ConsumedAt,
		&out.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PhoneVerificationCode{}, models.ErrNotFound
		}
		return models.PhoneVerificationCode{}, fmt.Errorf("get active phone verification code: %w", err)
	}
	return out, nil
}

func (r *UserRepo) ConsumePhoneVerificationCode(ctx context.Context, codeID int64) error {
	const q = `UPDATE phone_verification_codes SET consumed_at = now() WHERE id = $1`
	ct, err := r.db.Exec(ctx, q, codeID)
	if err != nil {
		return fmt.Errorf("consume phone verification code: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) DecrementPhoneVerificationAttempts(ctx context.Context, codeID int64) error {
	const q = `
		UPDATE phone_verification_codes
		SET attempts_left = GREATEST(attempts_left - 1, 0)
		WHERE id = $1
	`
	ct, err := r.db.Exec(ctx, q, codeID)
	if err != nil {
		return fmt.Errorf("decrement phone verification attempts: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) CreatePasswordResetToken(ctx context.Context, p repository.CreatePasswordResetTokenParams) (models.PasswordResetToken, error) {
	const q = `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, consumed_at, created_at
	`
	var out models.PasswordResetToken
	err := r.db.QueryRow(ctx, q, p.UserID, p.TokenHash, p.ExpiresAt).Scan(
		&out.ID,
		&out.UserID,
		&out.TokenHash,
		&out.ExpiresAt,
		&out.ConsumedAt,
		&out.CreatedAt,
	)
	if err != nil {
		return models.PasswordResetToken{}, fmt.Errorf("create password reset token: %w", err)
	}
	return out, nil
}

func (r *UserRepo) GetActivePasswordResetToken(ctx context.Context, userID int64) (models.PasswordResetToken, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, consumed_at, created_at
		FROM password_reset_tokens
		WHERE user_id = $1 AND consumed_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`
	var out models.PasswordResetToken
	err := r.db.QueryRow(ctx, q, userID).Scan(
		&out.ID,
		&out.UserID,
		&out.TokenHash,
		&out.ExpiresAt,
		&out.ConsumedAt,
		&out.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.PasswordResetToken{}, models.ErrNotFound
		}
		return models.PasswordResetToken{}, fmt.Errorf("get active password reset token: %w", err)
	}
	return out, nil
}

func (r *UserRepo) ConsumePasswordResetToken(ctx context.Context, tokenID int64) error {
	const q = `UPDATE password_reset_tokens SET consumed_at = now() WHERE id = $1`
	ct, err := r.db.Exec(ctx, q, tokenID)
	if err != nil {
		return fmt.Errorf("consume password reset token: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	const q = `UPDATE users SET password_hash = $2, updated_at = now() WHERE id = $1`
	ct, err := r.db.Exec(ctx, q, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) CreateTelegramLinkToken(ctx context.Context, p repository.CreateTelegramLinkTokenParams) (models.TelegramLinkToken, error) {
	const q = `
		INSERT INTO telegram_link_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id, user_id, token_hash, expires_at, consumed_at, created_at
	`
	var out models.TelegramLinkToken
	err := r.db.QueryRow(ctx, q, p.UserID, p.TokenHash, p.ExpiresAt).Scan(
		&out.ID,
		&out.UserID,
		&out.TokenHash,
		&out.ExpiresAt,
		&out.ConsumedAt,
		&out.CreatedAt,
	)
	if err != nil {
		return models.TelegramLinkToken{}, fmt.Errorf("create telegram link token: %w", err)
	}
	return out, nil
}

func (r *UserRepo) GetActiveTelegramLinkTokenByHash(ctx context.Context, tokenHash string) (models.TelegramLinkToken, error) {
	const q = `
		SELECT id, user_id, token_hash, expires_at, consumed_at, created_at
		FROM telegram_link_tokens
		WHERE token_hash = $1 AND consumed_at IS NULL
		LIMIT 1
	`
	var out models.TelegramLinkToken
	err := r.db.QueryRow(ctx, q, tokenHash).Scan(
		&out.ID,
		&out.UserID,
		&out.TokenHash,
		&out.ExpiresAt,
		&out.ConsumedAt,
		&out.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.TelegramLinkToken{}, models.ErrNotFound
		}
		return models.TelegramLinkToken{}, fmt.Errorf("get active telegram link token: %w", err)
	}
	return out, nil
}

func (r *UserRepo) ConsumeTelegramLinkToken(ctx context.Context, tokenID int64) error {
	const q = `UPDATE telegram_link_tokens SET consumed_at = now() WHERE id = $1`
	ct, err := r.db.Exec(ctx, q, tokenID)
	if err != nil {
		return fmt.Errorf("consume telegram link token: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *UserRepo) LinkTelegram(ctx context.Context, userID int64, chatID, username string) error {
	const q = `
		UPDATE users
		SET telegram_chat_id = $2,
		    telegram_username = $3,
		    telegram_linked_at = now(),
		    updated_at = now()
		WHERE id = $1
	`
	ct, err := r.db.Exec(ctx, q, userID, chatID, username)
	if err != nil {
		return fmt.Errorf("link telegram: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func userScanDest(out *models.User) []any {
	return []any{
		&out.ID,
		&out.Email,
		&out.Name,
		&out.Phone,
		&out.PasswordHash,
		&out.PhotoURL,
		&out.BonusPoints,
		&out.EmailVerified,
		&out.PhoneVerified,
		&out.IsBlocked,
		&out.IsAdmin,
		&out.EmailVerificationSentAt,
		&out.EmailVerifiedAt,
		&out.PhoneVerifiedAt,
		&out.VKID,
		&out.AuthProvider,
		&out.TelegramChatID,
		&out.TelegramUsername,
		&out.TelegramLinkedAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	}
}
