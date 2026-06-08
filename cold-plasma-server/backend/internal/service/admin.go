package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
	"cold-plasma-server/internal/security"
)

const aiChatRetention = 30 * 24 * time.Hour

type AdminService struct {
	chats      repository.ChatRepository
	bookings   repository.BookingRepository
	adminNotes repository.AdminNoteRepository
	settings   repository.SettingsRepository
	notifier   AdminBookingNotifier
	crypto     *security.TextCipher
}

type AdminBookingNotifier interface {
	NotifyUserBookingConfirmed(ctx context.Context, booking models.Booking)
	NotifyUserBookingRescheduled(ctx context.Context, booking models.Booking)
	NotifyUserBookingCompleted(ctx context.Context, booking models.Booking)
}

func NewAdminService(chats repository.ChatRepository, bookings repository.BookingRepository, adminNotes repository.AdminNoteRepository, settings repository.SettingsRepository, notifier AdminBookingNotifier, crypto *security.TextCipher) *AdminService {
	return &AdminService{chats: chats, bookings: bookings, adminNotes: adminNotes, settings: settings, notifier: notifier, crypto: crypto}
}

func (s *AdminService) ChatLogs(ctx context.Context, limit int) ([]models.ChatLog, error) {
	if err := s.CleanupOldChatSessions(ctx, time.Now()); err != nil {
		return nil, err
	}
	items, err := s.chats.ListAll(ctx, limit)
	if err != nil {
		return nil, err
	}
	if s.crypto == nil {
		return items, nil
	}
	for i := range items {
		items[i].RawInput = s.crypto.DecryptIfEncrypted(items[i].RawInput)
		items[i].RawOutput = s.crypto.DecryptIfEncrypted(items[i].RawOutput)
	}
	return items, nil
}

func (s *AdminService) DeleteChatSession(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimPrefix(strings.TrimSpace(sessionID), "session:")
	if sessionID == "" {
		return fmt.Errorf("session_id обязателен: %w", ErrValidation)
	}
	return s.chats.DeleteSession(ctx, sessionID)
}

func (s *AdminService) CleanupOldChatSessions(ctx context.Context, now time.Time) error {
	return s.chats.DeleteOlderThan(ctx, now.Add(-aiChatRetention))
}

type AdminChatMessage struct {
	Role      string    `json:"role"`
	Text      string    `json:"text"`
	Model     string    `json:"model,omitempty"`
	Intent    string    `json:"intent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminChatSession struct {
	ID       string             `json:"id"`
	Title    string             `json:"title"`
	UserID   *int64             `json:"user_id"`
	Name     string             `json:"name"`
	Email    string             `json:"email"`
	Phone    string             `json:"phone"`
	LastAt   time.Time          `json:"last_at"`
	Messages []AdminChatMessage `json:"messages"`
}

func (s *AdminService) ChatSessions(ctx context.Context, limit int) ([]AdminChatSession, error) {
	logs, err := s.ChatLogs(ctx, limit)
	if err != nil {
		return nil, err
	}
	grouped := make(map[string]*AdminChatSession)
	for _, item := range logs {
		key := chatSessionKey(item)
		session := grouped[key]
		if session == nil {
			title := firstNonEmpty(item.UserName, item.UserEmail, item.UserPhone, "Гость")
			session = &AdminChatSession{
				ID:     key,
				Title:  title,
				UserID: item.UserID,
				Name:   item.UserName,
				Email:  item.UserEmail,
				Phone:  item.UserPhone,
			}
			grouped[key] = session
		}
		session.Messages = append(session.Messages,
			AdminChatMessage{Role: "user", Text: item.RawInput, CreatedAt: item.CreatedAt},
			AdminChatMessage{Role: "assistant", Text: item.RawOutput, Model: item.AIModel, Intent: item.Intent, CreatedAt: item.CreatedAt},
		)
		if item.CreatedAt.After(session.LastAt) {
			session.LastAt = item.CreatedAt
		}
	}
	out := make([]AdminChatSession, 0, len(grouped))
	for _, session := range grouped {
		sort.Slice(session.Messages, func(i, j int) bool {
			return session.Messages[i].CreatedAt.Before(session.Messages[j].CreatedAt)
		})
		out = append(out, *session)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].LastAt.After(out[j].LastAt)
	})
	return out, nil
}

func (s *AdminService) BookingRequests(ctx context.Context) ([]models.AdminBooking, error) {
	return s.bookings.ListAdmin(ctx, []string{"new"}, 200)
}

func (s *AdminService) ActiveBookings(ctx context.Context) ([]models.AdminBooking, error) {
	return s.bookings.ListAdmin(ctx, []string{"confirmed"}, 200)
}

func (s *AdminService) CompletedBookings(ctx context.Context) ([]models.AdminBooking, error) {
	return s.bookings.ListAdmin(ctx, []string{"completed"}, 200)
}

// CalendarData — данные календаря администратора за период.
type CalendarData struct {
	From     time.Time                `json:"from"`
	To       time.Time                `json:"to"`
	Bookings []models.CalendarBooking `json:"bookings"`
	Notes    []models.AdminNote       `json:"notes"`
}

// Calendar возвращает записи и заметки администратора, пересекающие [from, to).
func (s *AdminService) Calendar(ctx context.Context, from, to time.Time) (CalendarData, error) {
	loc := salonLocation()
	from = from.In(loc)
	to = to.In(loc)
	if !to.After(from) {
		return CalendarData{}, fmt.Errorf("некорректный диапазон дат: %w", ErrValidation)
	}
	bookings, err := s.bookings.ListCalendar(ctx, from, to, []string{"new", "confirmed", "completed"})
	if err != nil {
		return CalendarData{}, err
	}
	notes, err := s.adminNotes.ListBetween(ctx, from, to)
	if err != nil {
		return CalendarData{}, err
	}
	return CalendarData{From: from, To: to, Bookings: bookings, Notes: notes}, nil
}

func (s *AdminService) CreateNote(ctx context.Context, startAt, endAt time.Time, title string) (models.AdminNote, error) {
	title = strings.TrimSpace(title)
	if startAt.IsZero() || endAt.IsZero() {
		return models.AdminNote{}, fmt.Errorf("укажите время начала и конца: %w", ErrValidation)
	}
	loc := salonLocation()
	startAt = startAt.In(loc)
	endAt = endAt.In(loc)
	if !endAt.After(startAt) {
		return models.AdminNote{}, fmt.Errorf("конец должен быть позже начала: %w", ErrValidation)
	}
	return s.adminNotes.Create(ctx, repository.CreateAdminNoteParams{StartAt: startAt, EndAt: endAt, Title: title})
}

func (s *AdminService) DeleteNote(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("id обязателен: %w", ErrValidation)
	}
	if err := s.adminNotes.Delete(ctx, id); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return se(ErrNotFound, "Заметка не найдена")
		}
		return err
	}
	return nil
}

func (s *AdminService) ConfirmBooking(ctx context.Context, bookingID int64, dateTime time.Time) error {
	booking, err := s.updateBookingStatusDateTime(ctx, bookingID, "confirmed", dateTime)
	if err != nil {
		return err
	}
	if s.notifier != nil {
		s.notifier.NotifyUserBookingConfirmed(ctx, booking)
	}
	return nil
}

func (s *AdminService) CompleteBooking(ctx context.Context, bookingID int64) error {
	booking, err := s.updateBookingStatusDateTime(ctx, bookingID, "completed", time.Time{})
	if err != nil {
		return err
	}
	if s.notifier != nil {
		s.notifier.NotifyUserBookingCompleted(ctx, booking)
	}
	return nil
}

func (s *AdminService) RescheduleBooking(ctx context.Context, bookingID int64, dateTime time.Time) error {
	if dateTime.IsZero() {
		return fmt.Errorf("datetime обязателен: %w", ErrValidation)
	}
	booking, err := s.updateBookingStatusDateTime(ctx, bookingID, "confirmed", dateTime)
	if err != nil {
		return err
	}
	if s.notifier != nil {
		s.notifier.NotifyUserBookingRescheduled(ctx, booking)
	}
	return nil
}

func (s *AdminService) updateBookingStatusDateTime(ctx context.Context, bookingID int64, status string, dateTime time.Time) (models.Booking, error) {
	if bookingID <= 0 {
		return models.Booking{}, fmt.Errorf("booking_id обязателен: %w", ErrValidation)
	}
	var current models.AdminBooking
	found := false
	if dateTime.IsZero() {
		items, err := s.bookings.ListAdmin(ctx, []string{"new", "confirmed", "completed"}, 500)
		if err != nil {
			return models.Booking{}, err
		}
		for _, item := range items {
			if item.ID == bookingID {
				current = item
				found = true
				dateTime = item.DateTime
				break
			}
		}
	} else {
		items, err := s.bookings.ListAdmin(ctx, []string{"new", "confirmed", "completed"}, 500)
		if err != nil {
			return models.Booking{}, err
		}
		for _, item := range items {
			if item.ID == bookingID {
				current = item
				found = true
				break
			}
		}
	}
	if dateTime.IsZero() {
		return models.Booking{}, fmt.Errorf("datetime обязателен: %w", ErrValidation)
	}
	if !found {
		return models.Booking{}, se(ErrNotFound, "Запись не найдена")
	}
	if err := s.bookings.UpdateStatusDateTime(ctx, bookingID, status, dateTime); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return models.Booking{}, se(ErrNotFound, "Запись не найдена")
		}
		return models.Booking{}, err
	}
	return models.Booking{
		ID:                 current.ID,
		UserID:             current.UserID,
		ProcedureID:        current.ProcedureID,
		DateTime:           dateTime,
		RequestedDateTimes: current.RequestedDateTimes,
		Comment:            current.Comment,
		Status:             status,
		NotifySMS:          current.NotifySMS,
		NotifyTelegram:     current.NotifyTelegram,
		CreatedAt:          current.CreatedAt,
	}, nil
}

func (s *AdminService) DeleteBooking(ctx context.Context, bookingID int64) error {
	if bookingID <= 0 {
		return fmt.Errorf("booking_id обязателен: %w", ErrValidation)
	}
	if err := s.bookings.Delete(ctx, bookingID); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			return se(ErrNotFound, "Заявка не найдена")
		}
		return err
	}
	return nil
}

type AdminNotificationSettings struct {
	NotifyTelegram bool   `json:"notify_telegram"`
	NotifySMS      bool   `json:"notify_sms"`
	AdminSMSPhone  string `json:"admin_sms_phone"`
}

type MasterProfile struct {
	Name         string   `json:"name"`
	Title        string   `json:"title"`
	Bio          string   `json:"bio"`
	PhotoURL     string   `json:"photo_url"`
	Certificates []string `json:"certificates"`
	Gallery      []string `json:"gallery"`
}

func (s *AdminService) MasterProfile(ctx context.Context) (MasterProfile, error) {
	values, err := s.settings.GetSettings(ctx, []string{"master_profile"})
	if err != nil {
		return MasterProfile{}, err
	}
	var out MasterProfile
	if strings.TrimSpace(values["master_profile"]) != "" {
		_ = json.Unmarshal([]byte(values["master_profile"]), &out)
	}
	return out, nil
}

func (s *AdminService) UpdateMasterProfile(ctx context.Context, profile MasterProfile) (MasterProfile, error) {
	profile.Name = strings.TrimSpace(profile.Name)
	profile.Title = strings.TrimSpace(profile.Title)
	profile.Bio = strings.TrimSpace(profile.Bio)
	profile.PhotoURL = strings.TrimSpace(profile.PhotoURL)
	profile.Certificates = cleanStringList(profile.Certificates)
	profile.Gallery = cleanStringList(profile.Gallery)
	raw, err := json.Marshal(profile)
	if err != nil {
		return MasterProfile{}, fmt.Errorf("marshal master profile: %w", err)
	}
	if err := s.settings.UpsertSettings(ctx, map[string]string{"master_profile": string(raw)}); err != nil {
		return MasterProfile{}, err
	}
	return profile, nil
}

func (s *AdminService) NotificationSettings(ctx context.Context) (AdminNotificationSettings, error) {
	values, err := s.settings.GetSettings(ctx, []string{
		"admin_notify_telegram",
		"admin_notify_sms",
		"admin_sms_phone",
	})
	if err != nil {
		return AdminNotificationSettings{}, err
	}
	return AdminNotificationSettings{
		NotifyTelegram: values["admin_notify_telegram"] == "true",
		NotifySMS:      values["admin_notify_sms"] == "true",
		AdminSMSPhone:  values["admin_sms_phone"],
	}, nil
}

func (s *AdminService) UpdateNotificationSettings(ctx context.Context, settings AdminNotificationSettings) (AdminNotificationSettings, error) {
	if err := s.settings.UpsertSettings(ctx, map[string]string{
		"admin_notify_telegram": boolString(settings.NotifyTelegram),
		"admin_notify_sms":      boolString(settings.NotifySMS),
		"admin_sms_phone":       settings.AdminSMSPhone,
	}); err != nil {
		return AdminNotificationSettings{}, err
	}
	return s.NotificationSettings(ctx)
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func chatSessionKey(item models.ChatLog) string {
	if strings.TrimSpace(item.SessionID) != "" {
		return "session:" + strings.TrimSpace(item.SessionID)
	}
	if item.UserID != nil {
		return fmt.Sprintf("user:%d", *item.UserID)
	}
	if strings.TrimSpace(item.UserEmail) != "" {
		return "email:" + strings.ToLower(strings.TrimSpace(item.UserEmail))
	}
	if strings.TrimSpace(item.UserPhone) != "" {
		return "phone:" + strings.TrimSpace(item.UserPhone)
	}
	return "guest:" + strings.TrimSpace(item.UserName)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v != "" {
			return v
		}
	}
	return ""
}

func cleanStringList(items []string) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}
