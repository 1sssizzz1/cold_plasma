package repository

import (
	"context"
	"time"

	"cold-plasma-server/internal/models"
)

// ErrNotFound используется, когда запись отсутствует.
// Репозиторий возвращает именно эту ошибку, чтобы сервис мог корректно маппить её на HTTP-коды.
var ErrNotFound = models.ErrNotFound

type UserRepository interface {
	Create(ctx context.Context, u CreateUserParams) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	GetByPhone(ctx context.Context, phone string) (models.User, error)
	GetByVKID(ctx context.Context, vkID string) (models.User, error)
	GetByID(ctx context.Context, id int64) (models.User, error)
	LinkVK(ctx context.Context, userID int64, vkID, email, name string) error
	UpdatePhoneUnverified(ctx context.Context, userID int64, phone string) error
	UpdateBonusPoints(ctx context.Context, userID int64, newBalance int) error
	MarkEmailVerified(ctx context.Context, userID int64) error
	MarkPhoneVerified(ctx context.Context, userID int64, phone string) error
	UpdateEmailVerificationSentAt(ctx context.Context, userID int64, t time.Time) error
	CreatePhoneVerificationCode(ctx context.Context, p CreatePhoneVerificationCodeParams) (models.PhoneVerificationCode, error)
	GetActivePhoneVerificationCode(ctx context.Context, userID int64, phone string) (models.PhoneVerificationCode, error)
	ConsumePhoneVerificationCode(ctx context.Context, codeID int64) error
	DecrementPhoneVerificationAttempts(ctx context.Context, codeID int64) error
	CreatePasswordResetToken(ctx context.Context, p CreatePasswordResetTokenParams) (models.PasswordResetToken, error)
	GetActivePasswordResetToken(ctx context.Context, userID int64) (models.PasswordResetToken, error)
	ConsumePasswordResetToken(ctx context.Context, tokenID int64) error
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
	CreateTelegramLinkToken(ctx context.Context, p CreateTelegramLinkTokenParams) (models.TelegramLinkToken, error)
	GetActiveTelegramLinkTokenByHash(ctx context.Context, tokenHash string) (models.TelegramLinkToken, error)
	ConsumeTelegramLinkToken(ctx context.Context, tokenID int64) error
	LinkTelegram(ctx context.Context, userID int64, chatID, username string) error
}

type CreateUserParams struct {
	Email         string
	Name          string
	Phone         string
	PasswordHash  string
	EmailVerified bool
	PhoneVerified bool
	VKID          string
	AuthProvider  string
}

type CreatePhoneVerificationCodeParams struct {
	UserID       int64
	Phone        string
	CodeHash     string
	ExpiresAt    time.Time
	AttemptsLeft int
}

type CreatePasswordResetTokenParams struct {
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
}

type CreateTelegramLinkTokenParams struct {
	UserID    int64
	TokenHash string
	ExpiresAt time.Time
}

type ProcedureRepository interface {
	ListActive(ctx context.Context) ([]models.Procedure, error)
	ListAll(ctx context.Context) ([]models.Procedure, error)
	GetByID(ctx context.Context, id int64) (models.Procedure, error)
	Create(ctx context.Context, p ProcedureParams) (models.Procedure, error)
	Update(ctx context.Context, id int64, p ProcedureParams) (models.Procedure, error)
	Delete(ctx context.Context, id int64) error
}

type ProcedureParams struct {
	Title        string
	Description  string
	DurationMins int
	Price        int
	BonusEarned  int
	Category     string
	ImageURL     string
	VideoURL     string
	Services     string
	DurationStr  string
	Popular      bool
	IsActive     bool
}

type ProcedureReviewRepository interface {
	ListPublic(ctx context.Context, procedureID int64, limit int) ([]models.ProcedureReview, error)
	ListAll(ctx context.Context, limit int) ([]models.ProcedureReview, error)
	CanUserReview(ctx context.Context, userID int64, procedureID int64) (bool, error)
	Upsert(ctx context.Context, userID int64, procedureID int64, rating int, text string) (models.ProcedureReview, error)
	Delete(ctx context.Context, id int64) error
}

type BeforeAfterRepository interface {
	ListPublic(ctx context.Context, limit int) ([]models.BeforeAfterResult, error)
	ListAll(ctx context.Context, limit int) ([]models.BeforeAfterResult, error)
	Create(ctx context.Context, p BeforeAfterParams) (models.BeforeAfterResult, error)
	Update(ctx context.Context, id int64, p BeforeAfterParams) (models.BeforeAfterResult, error)
	Delete(ctx context.Context, id int64) error
}

type BeforeAfterParams struct {
	ProcedureID *int64
	Procedure   string
	Title       string
	Description string
	BeforeURL   string
	AfterURL    string
	IsFeatured  bool
	SortOrder   int
	IsActive    bool
}

type BookingRepository interface {
	Create(ctx context.Context, b CreateBookingParams) (models.Booking, error)
	ListByUserID(ctx context.Context, userID int64) ([]models.Booking, error)
	ListAdmin(ctx context.Context, statuses []string, limit int) ([]models.AdminBooking, error)
	UpdateStatusDateTime(ctx context.Context, bookingID int64, status string, dateTime time.Time) error
	Delete(ctx context.Context, bookingID int64) error
	MonthlyRevenue(ctx context.Context, from, to time.Time) (models.MonthlyRevenue, error)
	ListTelegramReminderDue(ctx context.Context, now time.Time, limit int) ([]models.Booking, error)
	StoreTelegramAdminReminder(ctx context.Context, bookingID int64, messageIDs []int, deleteAt time.Time) error
	ListTelegramAdminReminderDeleteDue(ctx context.Context, now time.Time, limit int) ([]models.TelegramAdminReminderMessage, error)
	MarkTelegramCreatedNotified(ctx context.Context, bookingID int64) error
	MarkTelegramReminderSent(ctx context.Context, bookingID int64) error
	MarkTelegramAdminReminderDeleted(ctx context.Context, bookingID int64) error
}

type CreateBookingParams struct {
	UserID         int64
	ProcedureID    int64
	DateTime       time.Time
	DateTimes      []time.Time
	Comment        string
	BonusUsed      int
	NotifySMS      bool
	NotifyTelegram bool
}

type BonusRepository interface {
	AddLog(ctx context.Context, userID int64, typ string, amount int, comment string) error
	ListLogsByUserID(ctx context.Context, userID int64, limit int) ([]models.BonusLog, error)
}

type ChatRepository interface {
	UpsertSession(ctx context.Context, p UpsertChatSessionParams) error
	CloseSession(ctx context.Context, sessionID string) error
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteOlderThan(ctx context.Context, cutoff time.Time) error
	AddChatLog(ctx context.Context, p CreateChatLogParams) error
	ListByUserID(ctx context.Context, userID int64, limit int) ([]models.ChatLog, error)
	ListAll(ctx context.Context, limit int) ([]models.ChatLog, error)
}

type UpsertChatSessionParams struct {
	ID        string
	UserID    *int64
	UserName  string
	UserEmail string
	UserPhone string
}

type SettingsRepository interface {
	GetSettings(ctx context.Context, keys []string) (map[string]string, error)
	UpsertSettings(ctx context.Context, values map[string]string) error
}

type CreateChatLogParams struct {
	SessionID string
	UserID    *int64
	UserName  string
	UserEmail string
	UserPhone string
	RawInput  string
	RawOutput string
	AIModel   string
	Intent    string
}

// TxManager позволяет сервисам выполнять несколько операций атомарно.
// Реализация — Postgres транзакция.
type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context, repos TxRepos) error) error
}

// TxRepos — набор репозиториев, привязанных к одной транзакции.
type TxRepos interface {
	User() UserRepository
	Booking() BookingRepository
	Bonus() BonusRepository
}
