package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
	"cold-plasma-server/internal/sms"

	"golang.org/x/crypto/bcrypt"
)

type PhoneVerificationService struct {
	users      repository.UserRepository
	sender     sms.Sender
	ttlSeconds int
}

func NewPhoneVerificationService(users repository.UserRepository, sender sms.Sender, ttlSeconds int) *PhoneVerificationService {
	return &PhoneVerificationService{users: users, sender: sender, ttlSeconds: ttlSeconds}
}

func (s *PhoneVerificationService) SendCode(ctx context.Context, userID int64) (string, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return "", se(ErrNotFound, "Пользователь не найден")
		}
		return "", err
	}
	phone, err := NormalizeRussianPhone(u.Phone)
	if err != nil {
		return "", err
	}
	if u.PhoneVerified && u.Phone == phone {
		return phone, nil
	}

	code, err := generatePhoneCode()
	if err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash phone code: %w", err)
	}
	if _, err := s.users.CreatePhoneVerificationCode(ctx, repository.CreatePhoneVerificationCodeParams{
		UserID:       userID,
		Phone:        phone,
		CodeHash:     string(hash),
		ExpiresAt:    phoneCodeExpiresAt(s.ttlSeconds),
		AttemptsLeft: 5,
	}); err != nil {
		return "", err
	}
	if s.sender != nil {
		if err := s.sender.SendCode(ctx, phone, code); err != nil {
			return "", se(ErrValidation, "Не удалось отправить SMS-код")
		}
	}
	return phone, nil
}

func (s *PhoneVerificationService) UpdatePhone(ctx context.Context, userID int64, phone string) (models.User, error) {
	phone, err := NormalizeRussianPhone(phone)
	if err != nil {
		return models.User{}, err
	}
	existing, err := s.users.GetByPhone(ctx, phone)
	if err == nil && existing.ID != userID {
		return models.User{}, se(ErrConflict, "Этот телефон уже привязан к другому аккаунту")
	}
	if err != nil && !errors.Is(err, models.ErrNotFound) {
		return models.User{}, err
	}
	if err := s.users.UpdatePhoneUnverified(ctx, userID, phone); err != nil {
		return models.User{}, err
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}
	u.PasswordHash = ""
	return u, nil
}

func (s *PhoneVerificationService) Verify(ctx context.Context, userID int64, code string) (models.User, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.User{}, se(ErrNotFound, "Пользователь не найден")
		}
		return models.User{}, err
	}
	phone, err := NormalizeRussianPhone(u.Phone)
	if err != nil {
		return models.User{}, err
	}
	item, err := s.users.GetActivePhoneVerificationCode(ctx, userID, phone)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.User{}, se(ErrValidation, "Запросите новый SMS-код")
		}
		return models.User{}, err
	}
	if time.Now().After(item.ExpiresAt) {
		return models.User{}, se(ErrValidation, "SMS-код истёк")
	}
	if item.AttemptsLeft <= 0 {
		return models.User{}, se(ErrValidation, "Попытки закончились, запросите новый код")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(item.CodeHash), []byte(code)); err != nil {
		_ = s.users.DecrementPhoneVerificationAttempts(ctx, item.ID)
		return models.User{}, se(ErrValidation, "Неверный SMS-код")
	}
	if err := s.users.ConsumePhoneVerificationCode(ctx, item.ID); err != nil {
		return models.User{}, err
	}
	if err := s.users.MarkPhoneVerified(ctx, userID, phone); err != nil {
		return models.User{}, err
	}
	u, err = s.users.GetByID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}
	u.PasswordHash = ""
	return u, nil
}
