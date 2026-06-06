package repository

import (
	"context"
	"time"

	"cold-plasma-server/internal/models"
	core "cold-plasma-server/internal/repository"
)

type BookingRepository interface {
	MonthlyRevenue(ctx context.Context, from, to time.Time) (models.MonthlyRevenue, error)
	ListTelegramReminderDue(ctx context.Context, now time.Time, limit int) ([]models.Booking, error)
	StoreTelegramAdminReminder(ctx context.Context, bookingID int64, messageIDs []int, deleteAt time.Time) error
	ListTelegramAdminReminderDeleteDue(ctx context.Context, now time.Time, limit int) ([]models.TelegramAdminReminderMessage, error)
	MarkTelegramCreatedNotified(ctx context.Context, bookingID int64) error
	MarkTelegramReminderSent(ctx context.Context, bookingID int64) error
	MarkTelegramAdminReminderDeleted(ctx context.Context, bookingID int64) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (models.User, error)
	CreateTelegramLinkToken(ctx context.Context, p core.CreateTelegramLinkTokenParams) (models.TelegramLinkToken, error)
	GetActiveTelegramLinkTokenByHash(ctx context.Context, tokenHash string) (models.TelegramLinkToken, error)
	ConsumeTelegramLinkToken(ctx context.Context, tokenID int64) error
	LinkTelegram(ctx context.Context, userID int64, chatID, username string) error
}

type SettingsRepository interface {
	GetSettings(ctx context.Context, keys []string) (map[string]string, error)
}

type ProcedureRepository interface {
	GetByID(ctx context.Context, id int64) (models.Procedure, error)
}
