package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
)

type BookingService struct {
	procedures repository.ProcedureRepository
	bookings   repository.BookingRepository
	tx         repository.TxManager
	notifier   BookingNotifier
}

type BookingNotifier interface {
	NotifyBookingCreated(ctx context.Context, booking models.Booking)
}

func NewBookingService(procedures repository.ProcedureRepository, bookings repository.BookingRepository, tx repository.TxManager, notifier BookingNotifier) *BookingService {
	return &BookingService{procedures: procedures, bookings: bookings, tx: tx, notifier: notifier}
}

func (s *BookingService) Create(ctx context.Context, userID int64, procedureID int64, dateTimes []time.Time, comment string, bonusUsed int, notifySMS bool, notifyTelegram bool) (models.Booking, int, error) {
	comment = strings.TrimSpace(comment)
	if procedureID <= 0 {
		return models.Booking{}, 0, fmt.Errorf("procedure_id обязателен: %w", ErrValidation)
	}
	dateTimes = normalizeBookingDates(dateTimes)
	if len(dateTimes) == 0 {
		return models.Booking{}, 0, fmt.Errorf("укажите хотя бы одну дату записи: %w", ErrValidation)
	}
	if bonusUsed != 0 {
		return models.Booking{}, 0, fmt.Errorf("бонусы нельзя списывать при записи: %w", ErrValidation)
	}

	proc, err := s.procedures.GetByID(ctx, procedureID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.Booking{}, 0, fmt.Errorf("процедура не найдена: %w", ErrNotFound)
		}
		return models.Booking{}, 0, err
	}
	if !proc.IsActive {
		return models.Booking{}, 0, fmt.Errorf("процедура недоступна: %w", ErrValidation)
	}

	var booking models.Booking
	var newBalance int

	err = s.tx.WithTx(ctx, func(ctx context.Context, repos repository.TxRepos) error {
		u, err := repos.User().GetByID(ctx, userID)
		if err != nil {
			if errors.Is(err, models.ErrNotFound) {
				return fmt.Errorf("пользователь не найден: %w", ErrNotFound)
			}
			return err
		}
		if u.IsBlocked {
			return fmt.Errorf("пользователь заблокирован: %w", ErrForbidden)
		}
		if !u.PhoneVerified {
			return se(ErrForbidden, "Подтвердите российский номер телефона перед записью")
		}
		newBalance = u.BonusPoints

		booking, err = repos.Booking().Create(ctx, repository.CreateBookingParams{
			UserID:         userID,
			ProcedureID:    procedureID,
			DateTime:       dateTimes[0],
			DateTimes:      dateTimes,
			Comment:        comment,
			BonusUsed:      0,
			NotifySMS:      notifySMS,
			NotifyTelegram: notifyTelegram,
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return models.Booking{}, 0, err
	}
	if s.notifier != nil {
		s.notifier.NotifyBookingCreated(ctx, booking)
	}
	return booking, newBalance, nil
}

func (s *BookingService) ListMy(ctx context.Context, userID int64) ([]models.Booking, error) {
	return s.bookings.ListByUserID(ctx, userID)
}

func normalizeBookingDates(items []time.Time) []time.Time {
	out := make([]time.Time, 0, len(items))
	seen := make(map[string]bool)
	for _, item := range items {
		if item.IsZero() {
			continue
		}
		key := item.UTC().Format(time.RFC3339)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	return out
}
