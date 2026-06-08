package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cold-plasma-server/internal/models"
	"cold-plasma-server/internal/repository"
)

type BookingRepo struct {
	db db
}

func NewBookingRepo(d db) *BookingRepo {
	return &BookingRepo{db: d}
}

func (r *BookingRepo) Create(ctx context.Context, b repository.CreateBookingParams) (models.Booking, error) {
	dates := make([]string, 0, len(b.DateTimes))
	for _, dt := range b.DateTimes {
		dates = append(dates, dt.Format(time.RFC3339))
	}
	rawDates, err := json.Marshal(dates)
	if err != nil {
		return models.Booking{}, fmt.Errorf("marshal requested datetimes: %w", err)
	}
	const q = `
		INSERT INTO bookings (user_id, procedure_id, datetime, requested_datetimes, comment, status, bonus_used, notify_sms, notify_telegram)
		VALUES ($1, $2, $3, $4::jsonb, $5, 'new', $6, $7, $8)
		RETURNING id, user_id, procedure_id, datetime, requested_datetimes, comment, status, bonus_used, notify_sms, notify_telegram, telegram_created_notified_at, telegram_reminder_sent_at, created_at
	`
	var out models.Booking
	var requestedRaw []byte
	err = r.db.QueryRow(ctx, q, b.UserID, b.ProcedureID, b.DateTime, string(rawDates), b.Comment, b.BonusUsed, b.NotifySMS, b.NotifyTelegram).Scan(bookingScanDest(&out, &requestedRaw)...)
	if err != nil {
		return models.Booking{}, fmt.Errorf("create booking: %w", err)
	}
	_ = json.Unmarshal(requestedRaw, &out.RequestedDateTimes)
	return out, nil
}

func (r *BookingRepo) ListByUserID(ctx context.Context, userID int64) ([]models.Booking, error) {
	const q = `
		SELECT id, user_id, procedure_id, datetime, requested_datetimes, comment, status, bonus_used, notify_sms, notify_telegram, telegram_created_notified_at, telegram_reminder_sent_at, created_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY datetime DESC
		LIMIT 100
	`
	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list bookings: %w", err)
	}
	defer rows.Close()

	out := make([]models.Booking, 0)
	for rows.Next() {
		var b models.Booking
		var requestedRaw []byte
		if err := rows.Scan(bookingScanDest(&b, &requestedRaw)...); err != nil {
			return nil, fmt.Errorf("scan booking: %w", err)
		}
		_ = json.Unmarshal(requestedRaw, &b.RequestedDateTimes)
		out = append(out, b)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list bookings: %w", rows.Err())
	}
	return out, nil
}

func (r *BookingRepo) ListAdmin(ctx context.Context, statuses []string, limit int) ([]models.AdminBooking, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	if len(statuses) == 0 {
		statuses = []string{"new", "confirmed"}
	}
	const q = `
		SELECT b.id, b.user_id, u.name, u.email, u.phone,
		       b.procedure_id, p.title, b.datetime, b.requested_datetimes,
		       b.comment, b.status, b.notify_sms, b.notify_telegram, b.created_at
		FROM bookings b
		JOIN users u ON u.id = b.user_id
		JOIN procedures p ON p.id = b.procedure_id
		WHERE b.status = ANY($1)
		ORDER BY b.created_at DESC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, q, statuses, limit)
	if err != nil {
		return nil, fmt.Errorf("list admin bookings: %w", err)
	}
	defer rows.Close()

	out := make([]models.AdminBooking, 0)
	for rows.Next() {
		var b models.AdminBooking
		var requestedRaw []byte
		if err := rows.Scan(
			&b.ID,
			&b.UserID,
			&b.UserName,
			&b.UserEmail,
			&b.UserPhone,
			&b.ProcedureID,
			&b.ProcedureTitle,
			&b.DateTime,
			&requestedRaw,
			&b.Comment,
			&b.Status,
			&b.NotifySMS,
			&b.NotifyTelegram,
			&b.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan admin booking: %w", err)
		}
		_ = json.Unmarshal(requestedRaw, &b.RequestedDateTimes)
		out = append(out, b)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list admin bookings: %w", rows.Err())
	}
	return out, nil
}

func (r *BookingRepo) ListCalendar(ctx context.Context, from, to time.Time, statuses []string) ([]models.CalendarBooking, error) {
	if len(statuses) == 0 {
		statuses = []string{"new", "confirmed", "completed"}
	}
	const q = `
		SELECT b.id, u.name, u.phone, b.procedure_id, p.title, p.duration_mins, b.datetime, b.status
		FROM bookings b
		JOIN users u ON u.id = b.user_id
		JOIN procedures p ON p.id = b.procedure_id
		WHERE b.status = ANY($3)
		  AND b.datetime < $2
		  AND b.datetime + make_interval(mins => GREATEST(p.duration_mins, 1)) > $1
		ORDER BY b.datetime ASC
	`
	rows, err := r.db.Query(ctx, q, from, to, statuses)
	if err != nil {
		return nil, fmt.Errorf("list calendar bookings: %w", err)
	}
	defer rows.Close()

	out := make([]models.CalendarBooking, 0)
	for rows.Next() {
		var b models.CalendarBooking
		if err := rows.Scan(&b.ID, &b.UserName, &b.UserPhone, &b.ProcedureID, &b.ProcedureTitle, &b.DurationMins, &b.StartAt, &b.Status); err != nil {
			return nil, fmt.Errorf("scan calendar booking: %w", err)
		}
		mins := b.DurationMins
		if mins < 1 {
			mins = 1
		}
		b.EndAt = b.StartAt.Add(time.Duration(mins) * time.Minute)
		out = append(out, b)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list calendar bookings: %w", rows.Err())
	}
	return out, nil
}

func (r *BookingRepo) UpdateStatusDateTime(ctx context.Context, bookingID int64, status string, dateTime time.Time) error {
	const q = `UPDATE bookings SET status = $2, datetime = $3 WHERE id = $1`
	ct, err := r.db.Exec(ctx, q, bookingID, status, dateTime)
	if err != nil {
		return fmt.Errorf("update booking status: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *BookingRepo) Delete(ctx context.Context, bookingID int64) error {
	const q = `DELETE FROM bookings WHERE id = $1`
	ct, err := r.db.Exec(ctx, q, bookingID)
	if err != nil {
		return fmt.Errorf("delete booking: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *BookingRepo) MonthlyRevenue(ctx context.Context, from, to time.Time) (models.MonthlyRevenue, error) {
	const q = `
		SELECT COUNT(*)::int,
		       COALESCE(SUM(p.price), 0)::int,
		       COALESCE(SUM(b.bonus_used), 0)::int,
		       COALESCE(SUM(GREATEST(p.price - b.bonus_used, 0)), 0)::int
		FROM bookings b
		JOIN procedures p ON p.id = b.procedure_id
		WHERE b.status = 'completed'
		  AND b.datetime >= $1
		  AND b.datetime < $2
	`
	out := models.MonthlyRevenue{
		PeriodStart: from,
		PeriodEnd:   to,
	}
	if err := r.db.QueryRow(ctx, q, from, to).Scan(&out.CompletedCount, &out.GrossAmount, &out.BonusUsed, &out.NetAmount); err != nil {
		return models.MonthlyRevenue{}, fmt.Errorf("monthly revenue: %w", err)
	}
	return out, nil
}

func (r *BookingRepo) ListTelegramReminderDue(ctx context.Context, now time.Time, limit int) ([]models.Booking, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	const q = `
		SELECT id, user_id, procedure_id, datetime, requested_datetimes, comment, status, bonus_used, notify_sms, notify_telegram, telegram_created_notified_at, telegram_reminder_sent_at, created_at
		FROM bookings
		WHERE telegram_reminder_sent_at IS NULL
		  AND datetime >= $1
		  AND datetime < $1 + interval '2 days'
		  AND status = 'confirmed'
		ORDER BY datetime ASC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, q, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list telegram reminders: %w", err)
	}
	defer rows.Close()
	out := make([]models.Booking, 0)
	for rows.Next() {
		var b models.Booking
		var requestedRaw []byte
		if err := rows.Scan(bookingScanDest(&b, &requestedRaw)...); err != nil {
			return nil, fmt.Errorf("scan telegram reminder: %w", err)
		}
		_ = json.Unmarshal(requestedRaw, &b.RequestedDateTimes)
		out = append(out, b)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list telegram reminders: %w", rows.Err())
	}
	return out, nil
}

func (r *BookingRepo) StoreTelegramAdminReminder(ctx context.Context, bookingID int64, messageIDs []int, deleteAt time.Time) error {
	raw, err := json.Marshal(messageIDs)
	if err != nil {
		return fmt.Errorf("marshal telegram admin reminder message ids: %w", err)
	}
	const q = `
		UPDATE bookings
		SET telegram_admin_reminder_message_ids = $2::jsonb,
		    telegram_admin_reminder_delete_at = $3,
		    telegram_admin_reminder_deleted_at = NULL
		WHERE id = $1
	`
	ct, err := r.db.Exec(ctx, q, bookingID, string(raw), deleteAt)
	if err != nil {
		return fmt.Errorf("store telegram admin reminder: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *BookingRepo) ListTelegramAdminReminderDeleteDue(ctx context.Context, now time.Time, limit int) ([]models.TelegramAdminReminderMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	const q = `
		SELECT id, telegram_admin_reminder_message_ids, telegram_admin_reminder_delete_at
		FROM bookings
		WHERE telegram_admin_reminder_delete_at IS NOT NULL
		  AND telegram_admin_reminder_deleted_at IS NULL
		  AND telegram_admin_reminder_delete_at <= $1
		ORDER BY telegram_admin_reminder_delete_at ASC
		LIMIT $2
	`
	rows, err := r.db.Query(ctx, q, now, limit)
	if err != nil {
		return nil, fmt.Errorf("list telegram admin reminder delete due: %w", err)
	}
	defer rows.Close()

	out := make([]models.TelegramAdminReminderMessage, 0)
	for rows.Next() {
		var item models.TelegramAdminReminderMessage
		var raw []byte
		if err := rows.Scan(&item.BookingID, &raw, &item.DeleteAt); err != nil {
			return nil, fmt.Errorf("scan telegram admin reminder delete due: %w", err)
		}
		_ = json.Unmarshal(raw, &item.MessageIDs)
		out = append(out, item)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("list telegram admin reminder delete due: %w", rows.Err())
	}
	return out, nil
}

func (r *BookingRepo) MarkTelegramCreatedNotified(ctx context.Context, bookingID int64) error {
	return r.markTelegramTime(ctx, bookingID, "telegram_created_notified_at")
}

func (r *BookingRepo) MarkTelegramReminderSent(ctx context.Context, bookingID int64) error {
	return r.markTelegramTime(ctx, bookingID, "telegram_reminder_sent_at")
}

func (r *BookingRepo) MarkTelegramAdminReminderDeleted(ctx context.Context, bookingID int64) error {
	return r.markTelegramTime(ctx, bookingID, "telegram_admin_reminder_deleted_at")
}

func (r *BookingRepo) markTelegramTime(ctx context.Context, bookingID int64, column string) error {
	q := fmt.Sprintf("UPDATE bookings SET %s = now() WHERE id = $1", column)
	ct, err := r.db.Exec(ctx, q, bookingID)
	if err != nil {
		return fmt.Errorf("mark telegram time: %w", err)
	}
	if ct.RowsAffected() == 0 {
		return models.ErrNotFound
	}
	return nil
}

func bookingScanDest(out *models.Booking, requestedRaw *[]byte) []any {
	return []any{
		&out.ID,
		&out.UserID,
		&out.ProcedureID,
		&out.DateTime,
		requestedRaw,
		&out.Comment,
		&out.Status,
		&out.BonusUsed,
		&out.NotifySMS,
		&out.NotifyTelegram,
		&out.TelegramCreatedNotifiedAt,
		&out.TelegramReminderSentAt,
		&out.CreatedAt,
	}
}
