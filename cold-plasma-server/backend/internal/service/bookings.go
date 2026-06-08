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
	adminNotes repository.AdminNoteRepository
	tx         repository.TxManager
	notifier   BookingNotifier
}

type BookingNotifier interface {
	NotifyBookingCreated(ctx context.Context, booking models.Booking)
}

func NewBookingService(procedures repository.ProcedureRepository, bookings repository.BookingRepository, adminNotes repository.AdminNoteRepository, tx repository.TxManager, notifier BookingNotifier) *BookingService {
	return &BookingService{procedures: procedures, bookings: bookings, adminNotes: adminNotes, tx: tx, notifier: notifier}
}

// DaySlots — свободные окна записи на один день.
type DaySlots struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Slots []Slot `json:"slots"`
}

// Availability возвращает свободные окна записи для процедуры на диапазон дат
// [from, to). Окна нарезаются по длительности процедуры; занятыми считаются
// подтверждённые записи и заметки администратора.
func (s *BookingService) Availability(ctx context.Context, procedureID int64, from, to time.Time) ([]DaySlots, error) {
	if procedureID <= 0 {
		return nil, fmt.Errorf("procedure_id обязателен: %w", ErrValidation)
	}
	proc, err := s.procedures.GetByID(ctx, procedureID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return nil, fmt.Errorf("процедура не найдена: %w", ErrNotFound)
		}
		return nil, err
	}
	if !proc.IsActive {
		return nil, fmt.Errorf("процедура недоступна: %w", ErrValidation)
	}
	if proc.DurationMins < 1 {
		return nil, fmt.Errorf("у процедуры не задана длительность: %w", ErrValidation)
	}

	loc := salonLocation()
	from = from.In(loc)
	to = to.In(loc)
	if !to.After(from) {
		return nil, fmt.Errorf("некорректный диапазон дат: %w", ErrValidation)
	}

	busy, err := s.busyIntervals(ctx, from, to)
	if err != nil {
		return nil, err
	}

	now := time.Now().In(loc)
	out := make([]DaySlots, 0)
	day := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, loc)
	last := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, loc)
	for ; !day.After(last); day = day.AddDate(0, 0, 1) {
		slots := generateDaySlots(day, proc.DurationMins, loc)
		free := make([]Slot, 0, len(slots))
		for _, slot := range slots {
			if slot.StartAt.Before(now) {
				continue
			}
			if intervalBusy(slot.StartAt, slot.EndAt, busy) {
				continue
			}
			free = append(free, slot)
		}
		out = append(out, DaySlots{Date: day.Format("2006-01-02"), Slots: free})
	}
	return out, nil
}

func (s *BookingService) busyIntervals(ctx context.Context, from, to time.Time) ([]interval, error) {
	notes, err := s.adminNotes.ListBetween(ctx, from, to)
	if err != nil {
		return nil, err
	}
	confirmed, err := s.bookings.ListCalendar(ctx, from, to, []string{"confirmed"})
	if err != nil {
		return nil, err
	}
	busy := make([]interval, 0, len(notes)+len(confirmed))
	for _, n := range notes {
		busy = append(busy, interval{start: n.StartAt, end: n.EndAt})
	}
	for _, b := range confirmed {
		busy = append(busy, interval{start: b.StartAt, end: b.EndAt})
	}
	return busy, nil
}

func intervalBusy(start, end time.Time, busy []interval) bool {
	for _, b := range busy {
		if overlaps(start, end, b.start, b.end) {
			return true
		}
	}
	return false
}

func (s *BookingService) Create(ctx context.Context, userID int64, procedureID int64, slotStart time.Time, comment string, bonusUsed int, notifySMS bool, notifyTelegram bool) (models.Booking, int, error) {
	comment = strings.TrimSpace(comment)
	if procedureID <= 0 {
		return models.Booking{}, 0, fmt.Errorf("procedure_id обязателен: %w", ErrValidation)
	}
	if slotStart.IsZero() {
		return models.Booking{}, 0, fmt.Errorf("выберите время записи: %w", ErrValidation)
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
	if proc.DurationMins < 1 {
		return models.Booking{}, 0, fmt.Errorf("у процедуры не задана длительность: %w", ErrValidation)
	}

	loc := salonLocation()
	slotStart = slotStart.In(loc)
	if err := s.validateSlot(ctx, slotStart, proc.DurationMins, loc); err != nil {
		return models.Booking{}, 0, err
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
			DateTime:       slotStart,
			DateTimes:      []time.Time{slotStart},
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

// validateSlot проверяет, что окно укладывается в рабочее время, выровнено по
// сетке слотов, не в прошлом и не пересекается с занятыми интервалами.
func (s *BookingService) validateSlot(ctx context.Context, slotStart time.Time, durationMins int, loc *time.Location) error {
	now := time.Now().In(loc)
	if slotStart.Before(now) {
		return se(ErrValidation, "Это время уже прошло")
	}
	day := time.Date(slotStart.Year(), slotStart.Month(), slotStart.Day(), 0, 0, 0, 0, loc)
	slots := generateDaySlots(day, durationMins, loc)
	slotEnd := slotStart.Add(time.Duration(durationMins) * time.Minute)
	matched := false
	for _, slot := range slots {
		if slot.StartAt.Equal(slotStart) {
			matched = true
			break
		}
	}
	if !matched {
		return se(ErrValidation, "Это время недоступно для записи")
	}

	busy, err := s.busyIntervals(ctx, slotStart, slotEnd)
	if err != nil {
		return err
	}
	if intervalBusy(slotStart, slotEnd, busy) {
		return se(ErrConflict, "Это время уже занято, выберите другое окно")
	}
	return nil
}

func (s *BookingService) ListMy(ctx context.Context, userID int64) ([]models.Booking, error) {
	return s.bookings.ListByUserID(ctx, userID)
}
