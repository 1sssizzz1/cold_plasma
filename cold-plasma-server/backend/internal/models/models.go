package models

import "time"

type User struct {
	ID                      int64      `db:"id" json:"id"`
	Email                   string     `db:"email" json:"email"`
	Name                    string     `db:"name" json:"name"`
	Phone                   string     `db:"phone" json:"phone"`
	PasswordHash            string     `db:"password_hash" json:"-"`
	PhotoURL                string     `db:"photo_url" json:"photo_url"`
	BonusPoints             int        `db:"bonus_points" json:"bonus_points"`
	EmailVerified           bool       `db:"email_verified" json:"email_verified"`
	PhoneVerified           bool       `db:"phone_verified" json:"phone_verified"`
	IsBlocked               bool       `db:"is_blocked" json:"is_blocked"`
	IsAdmin                 bool       `db:"is_admin" json:"is_admin"`
	EmailVerificationSentAt *time.Time `db:"email_verification_sent_at" json:"email_verification_sent_at"`
	EmailVerifiedAt         *time.Time `db:"email_verified_at" json:"email_verified_at"`
	PhoneVerifiedAt         *time.Time `db:"phone_verified_at" json:"phone_verified_at"`
	VKID                    string     `db:"vk_id" json:"vk_id"`
	AuthProvider            string     `db:"auth_provider" json:"auth_provider"`
	TelegramChatID          string     `db:"telegram_chat_id" json:"telegram_chat_id"`
	TelegramUsername        string     `db:"telegram_username" json:"telegram_username"`
	TelegramLinkedAt        *time.Time `db:"telegram_linked_at" json:"telegram_linked_at"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}

type PhoneVerificationCode struct {
	ID           int64      `db:"id" json:"id"`
	UserID       int64      `db:"user_id" json:"user_id"`
	Phone        string     `db:"phone" json:"phone"`
	CodeHash     string     `db:"code_hash" json:"-"`
	ExpiresAt    time.Time  `db:"expires_at" json:"expires_at"`
	AttemptsLeft int        `db:"attempts_left" json:"attempts_left"`
	ConsumedAt   *time.Time `db:"consumed_at" json:"consumed_at"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

type Procedure struct {
	ID           int64     `db:"id" json:"id"`
	Title        string    `db:"title" json:"title"`
	Description  string    `db:"description" json:"description"`
	DurationMins int       `db:"duration_mins" json:"duration_mins"`
	Price        int       `db:"price" json:"price"`
	BonusEarned  int       `db:"bonus_earned" json:"bonus_earned"`
	Category     string    `db:"category" json:"category"`
	ImageURL     string    `db:"image_url" json:"image_url"`
	VideoURL     string    `db:"video_url" json:"video_url"`
	Services     string    `db:"services" json:"services"`
	DurationStr  string    `db:"duration_str" json:"duration_str"`
	Popular      bool      `db:"popular" json:"popular"`
	IsActive     bool      `db:"is_active" json:"is_active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}

type ProcedureReview struct {
	ID             int64     `json:"id"`
	UserID         int64     `json:"user_id"`
	UserName       string    `json:"user_name"`
	ProcedureID    int64     `json:"procedure_id"`
	ProcedureTitle string    `json:"procedure_title,omitempty"`
	Rating         int       `json:"rating"`
	Text           string    `json:"text"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type BeforeAfterResult struct {
	ID          int64     `json:"id"`
	ProcedureID *int64    `json:"procedure_id"`
	Procedure   string    `json:"procedure"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	BeforeURL   string    `json:"before_url"`
	AfterURL    string    `json:"after_url"`
	IsFeatured  bool      `json:"is_featured"`
	SortOrder   int       `json:"sort_order"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Booking struct {
	ID                        int64      `db:"id" json:"id"`
	UserID                    int64      `db:"user_id" json:"user_id"`
	ProcedureID               int64      `db:"procedure_id" json:"procedure_id"`
	DateTime                  time.Time  `db:"datetime" json:"datetime"`
	RequestedDateTimes        []string   `db:"requested_datetimes" json:"requested_datetimes"`
	Comment                   string     `db:"comment" json:"comment"`
	Status                    string     `db:"status" json:"status"`
	BonusUsed                 int        `db:"bonus_used" json:"bonus_used"`
	NotifySMS                 bool       `db:"notify_sms" json:"notify_sms"`
	NotifyTelegram            bool       `db:"notify_telegram" json:"notify_telegram"`
	TelegramCreatedNotifiedAt *time.Time `db:"telegram_created_notified_at" json:"telegram_created_notified_at"`
	TelegramReminderSentAt    *time.Time `db:"telegram_reminder_sent_at" json:"telegram_reminder_sent_at"`
	CreatedAt                 time.Time  `db:"created_at" json:"created_at"`
}

type AdminBooking struct {
	ID                 int64     `json:"id"`
	UserID             int64     `json:"user_id"`
	UserName           string    `json:"user_name"`
	UserEmail          string    `json:"user_email"`
	UserPhone          string    `json:"user_phone"`
	ProcedureID        int64     `json:"procedure_id"`
	ProcedureTitle     string    `json:"procedure_title"`
	DateTime           time.Time `json:"datetime"`
	RequestedDateTimes []string  `json:"requested_datetimes"`
	Comment            string    `json:"comment"`
	Status             string    `json:"status"`
	NotifySMS          bool      `json:"notify_sms"`
	NotifyTelegram     bool      `json:"notify_telegram"`
	CreatedAt          time.Time `json:"created_at"`
}

type AdminNote struct {
	ID        int64     `db:"id" json:"id"`
	StartAt   time.Time `db:"start_at" json:"start_at"`
	EndAt     time.Time `db:"end_at" json:"end_at"`
	Title     string    `db:"title" json:"title"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// CalendarBooking — компактное представление записи для календаря администратора.
type CalendarBooking struct {
	ID             int64     `json:"id"`
	UserName       string    `json:"user_name"`
	UserPhone      string    `json:"user_phone"`
	ProcedureID    int64     `json:"procedure_id"`
	ProcedureTitle string    `json:"procedure_title"`
	DurationMins   int       `json:"duration_mins"`
	StartAt        time.Time `json:"start_at"`
	EndAt          time.Time `json:"end_at"`
	Status         string    `json:"status"`
}

type MonthlyRevenue struct {
	PeriodStart    time.Time `json:"period_start"`
	PeriodEnd      time.Time `json:"period_end"`
	CompletedCount int       `json:"completed_count"`
	GrossAmount    int       `json:"gross_amount"`
	BonusUsed      int       `json:"bonus_used"`
	NetAmount      int       `json:"net_amount"`
}

type TelegramAdminReminderMessage struct {
	BookingID  int64     `json:"booking_id"`
	MessageIDs []int     `json:"message_ids"`
	DeleteAt   time.Time `json:"delete_at"`
}

type PasswordResetToken struct {
	ID         int64      `db:"id" json:"id"`
	UserID     int64      `db:"user_id" json:"user_id"`
	TokenHash  string     `db:"token_hash" json:"-"`
	ExpiresAt  time.Time  `db:"expires_at" json:"expires_at"`
	ConsumedAt *time.Time `db:"consumed_at" json:"consumed_at"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

type TelegramLinkToken struct {
	ID         int64      `db:"id" json:"id"`
	UserID     int64      `db:"user_id" json:"user_id"`
	TokenHash  string     `db:"token_hash" json:"-"`
	ExpiresAt  time.Time  `db:"expires_at" json:"expires_at"`
	ConsumedAt *time.Time `db:"consumed_at" json:"consumed_at"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

type BonusLog struct {
	ID        int64     `db:"id" json:"id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Type      string    `db:"type" json:"type"`
	Amount    int       `db:"amount" json:"amount"`
	Comment   string    `db:"comment" json:"comment"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type ChatLog struct {
	ID        int64     `db:"id" json:"id"`
	SessionID string    `db:"session_id" json:"session_id"`
	UserID    *int64    `db:"user_id" json:"user_id"`
	UserName  string    `db:"user_name" json:"user_name"`
	UserEmail string    `db:"user_email" json:"user_email"`
	UserPhone string    `db:"user_phone" json:"user_phone"`
	RawInput  string    `db:"raw_input" json:"raw_input"`
	RawOutput string    `db:"raw_output" json:"raw_output"`
	AIModel   string    `db:"ai_model" json:"ai_model"`
	Intent    string    `db:"intent" json:"intent"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
