package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"cold-plasma-server/internal/email"
	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type PasswordResetService struct {
	users      repository.UserRepository
	mailer     email.Sender
	publicBase string
	ttl        time.Duration
}

func NewPasswordResetService(users repository.UserRepository, mailer email.Sender, publicBaseURL string) *PasswordResetService {
	return &PasswordResetService{
		users:      users,
		mailer:     mailer,
		publicBase: strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"),
		ttl:        30 * time.Minute,
	}
}

func (s *PasswordResetService) Request(ctx context.Context, emailAddr string) error {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	if _, err := mail.ParseAddress(emailAddr); err != nil {
		return se(ErrValidation, "Введите корректный email")
	}
	u, err := s.users.GetByEmail(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil
		}
		return err
	}
	if s.mailer == nil {
		return se(ErrValidation, "Почта для восстановления пароля не настроена")
	}
	token, err := randomResetToken()
	if err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash reset token: %w", err)
	}
	if _, err := s.users.CreatePasswordResetToken(ctx, repository.CreatePasswordResetTokenParams{
		UserID:    u.ID,
		TokenHash: string(hash),
		ExpiresAt: time.Now().Add(s.ttl),
	}); err != nil {
		return err
	}
	resetURL := s.publicBase + "/account?reset_token=" + url.QueryEscape(token) + "&email=" + url.QueryEscape(emailAddr)

	// Используем красивый шаблон email
	tmpl := email.PasswordResetEmailTemplate(resetURL, "Студия холодной плазмы Plasma Glow")

	if err := s.mailer.Send(ctx, emailAddr, tmpl.Subject, tmpl.TextBody, tmpl.HTMLBody); err != nil {
		return fmt.Errorf("send reset email: %w", err)
	}
	return nil
}

func (s *PasswordResetService) Reset(ctx context.Context, emailAddr, token, password string) error {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	token = strings.TrimSpace(token)
	if emailAddr == "" || token == "" || password == "" {
		return se(ErrValidation, "Email, токен и новый пароль обязательны")
	}
	if len(password) < 6 {
		return se(ErrValidation, "Пароль должен быть не короче 6 символов")
	}
	u, err := s.users.GetByEmail(ctx, emailAddr)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return se(ErrValidation, "Ссылка восстановления недействительна")
		}
		return err
	}
	item, err := s.users.GetActivePasswordResetToken(ctx, u.ID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return se(ErrValidation, "Ссылка восстановления недействительна")
		}
		return err
	}
	if time.Now().After(item.ExpiresAt) {
		return se(ErrValidation, "Ссылка восстановления устарела")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(item.TokenHash), []byte(token)); err != nil {
		return se(ErrValidation, "Ссылка восстановления недействительна")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, u.ID, string(hash)); err != nil {
		return err
	}
	return s.users.ConsumePasswordResetToken(ctx, item.ID)
}

func randomResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate reset token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
