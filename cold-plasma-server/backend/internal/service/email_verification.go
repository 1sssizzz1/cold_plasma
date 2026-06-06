package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"cold-plasma-server/internal/email"
	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
	appjwt "cold-plasma-server/pkg/jwt"
)

type EmailVerificationService struct {
	users       repository.UserRepository
	mailer      email.Sender
	publicBase  string
	secret      string
	expirySec   int
	minCooldown time.Duration
}

func NewEmailVerificationService(users repository.UserRepository, mailer email.Sender, publicBaseURL, secret string, expirySec int) *EmailVerificationService {
	return &EmailVerificationService{
		users:       users,
		mailer:      mailer,
		publicBase:  strings.TrimRight(strings.TrimSpace(publicBaseURL), "/"),
		secret:      secret,
		expirySec:   expirySec,
		minCooldown: 60 * time.Second,
	}
}

func (s *EmailVerificationService) SendForUser(ctx context.Context, u models.User) (bool, error) {
	if u.EmailVerified {
		return true, nil
	}
	if _, err := mail.ParseAddress(u.Email); err != nil {
		return false, se(ErrValidation, "Email некорректный")
	}
	if s.mailer == nil {
		return false, se(ErrValidation, "Почта для подтверждения не настроена (SMTP)")
	}
	if s.secret == "" {
		return false, se(ErrValidation, "Секрет для подтверждения email не настроен")
	}
	if s.publicBase == "" {
		return false, se(ErrValidation, "PUBLIC_BASE_URL не задан")
	}

	if u.EmailVerificationSentAt != nil && time.Since(*u.EmailVerificationSentAt) < s.minCooldown {
		return false, se(ErrValidation, "Письмо уже отправлено, попробуйте позже")
	}

	token, err := appjwt.GenerateEmailVerifyToken(u.ID, u.Email, s.secret, s.expirySec)
	if err != nil {
		return false, fmt.Errorf("generate verify token: %w", err)
	}

	verifyURL := s.publicBase + "/api/v1/auth/verify?token=" + url.QueryEscape(token)

	// Используем красивый шаблон email
	tmpl := email.VerificationEmailTemplate(verifyURL, "Студия холодной плазмы Plasma Glow")

	if err := s.mailer.Send(ctx, u.Email, tmpl.Subject, tmpl.TextBody, tmpl.HTMLBody); err != nil {
		return false, fmt.Errorf("send email: %w", err)
	}

	_ = s.users.UpdateEmailVerificationSentAt(ctx, u.ID, time.Now())
	return true, nil
}

func (s *EmailVerificationService) SendByEmail(ctx context.Context, emailAddr string) (bool, error) {
	emailAddr = strings.TrimSpace(strings.ToLower(emailAddr))
	u, err := s.users.GetByEmail(ctx, emailAddr)
	if err != nil {
		// Чтобы не раскрывать существование email, отвечаем "успешно", но ничего не отправляем.
		if errors.Is(err, models.ErrNotFound) {
			return true, nil
		}
		return false, err
	}
	return s.SendForUser(ctx, u)
}

func (s *EmailVerificationService) VerifyToken(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return se(ErrValidation, "Токен подтверждения не найден")
	}
	if s.secret == "" {
		return se(ErrValidation, "Подтверждение email не настроено")
	}
	claims, err := appjwt.ValidateEmailVerifyToken(token, s.secret)
	if err != nil {
		return se(ErrValidation, "Ссылка подтверждения недействительна или устарела")
	}
	if err := s.users.MarkEmailVerified(ctx, claims.UserID); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return se(ErrNotFound, "Пользователь не найден")
		}
		return err
	}
	return nil
}
