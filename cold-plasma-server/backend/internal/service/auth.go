package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
	appjwt "cold-plasma-server/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     repository.UserRepository
	verifier  *EmailVerificationService
	jwtSecret string
	jwtExpiry int
}

func NewAuthService(users repository.UserRepository, verifier *EmailVerificationService, jwtSecret string, jwtExpiry int) *AuthService {
	return &AuthService{
		users:     users,
		verifier:  verifier,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password, name, phone string) (models.User, bool, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)
	phone = strings.TrimSpace(phone)

	if email == "" || password == "" || name == "" || phone == "" {
		return models.User{}, false, se(ErrValidation, "Email, пароль, имя, фамилия и российский телефон обязательны")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return models.User{}, false, se(ErrValidation, "Введите корректный email")
	}
	normalizedPhone, err := NormalizeRussianPhone(phone)
	if err != nil {
		return models.User{}, false, err
	}
	if len(password) < 6 {
		return models.User{}, false, se(ErrValidation, "Пароль должен быть не короче 6 символов")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return models.User{}, false, fmt.Errorf("bcrypt: %w", err)
	}

	u, err := s.users.Create(ctx, repository.CreateUserParams{
		Email:        email,
		Name:         name,
		Phone:        normalizedPhone,
		PasswordHash: string(hash),
		AuthProvider: "password",
	})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return models.User{}, false, se(ErrConflict, "Этот email или телефон уже зарегистрирован")
		}
		return models.User{}, false, err
	}

	u.PasswordHash = ""

	sent := false
	if s.verifier != nil {
		if ok, err := s.verifier.SendForUser(ctx, u); err == nil {
			sent = ok
		} else {
			// Не ломаем регистрацию из-за почты: пользователь сможет запросить повторную отправку.
			sent = false
		}
	}

	return u, sent, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (models.User, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" || password == "" {
		return models.User{}, "", se(ErrValidation, "Email и пароль обязательны")
	}

	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.User{}, "", se(ErrUnauthorized, "Неверный email или пароль")
		}
		return models.User{}, "", err
	}

	if u.IsBlocked {
		return models.User{}, "", se(ErrForbidden, "Аккаунт заблокирован")
	}
	if !u.EmailVerified {
		return models.User{}, "", se(ErrForbidden, "Подтвердите email (ссылка в письме)")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return models.User{}, "", se(ErrUnauthorized, "Неверный email или пароль")
	}

	token, err := appjwt.GenerateToken(int(u.ID), u.Email, u.Name, u.IsAdmin, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return models.User{}, "", fmt.Errorf("jwt: %w", err)
	}
	u.PasswordHash = ""
	return u, token, nil
}

func (s *AuthService) Me(ctx context.Context, userID int64) (models.User, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.User{}, se(ErrNotFound, "Пользователь не найден")
		}
		return models.User{}, err
	}
	u.PasswordHash = ""
	return u, nil
}
