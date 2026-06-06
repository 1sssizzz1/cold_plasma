package service

import (
	"fmt"
	"strings"
	"time"

	"cold-plasma-server/internal/models"
)

var leadEmoji = map[string]string{
	"hot":          "🔥",
	"warm":         "🌡",
	"cold":         "🧊",
	"notify_admin": "⚠️",
}

var priorityEmoji = map[string]string{
	"high":   "🔴",
	"normal": "🟡",
	"low":    "🟢",
}

func bookingAdminText(title string, user models.User, booking models.Booking, procedureTitle string, lead LeadAnalysis) string {
	lEmoji := firstNonEmpty(leadEmoji[lead.Temperature], "📋")
	pEmoji := firstNonEmpty(priorityEmoji[lead.Priority], "🟡")

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>%s</b>\n\n", escapeHTML(title)))

	sb.WriteString(fmt.Sprintf("%s Лид: <b>%s</b>\n", lEmoji, escapeHTML(lead.Temperature)))
	sb.WriteString(fmt.Sprintf("%s Приоритет: <b>%s</b>\n", pEmoji, escapeHTML(lead.Priority)))
	if lead.Reason != "" {
		sb.WriteString(fmt.Sprintf("💬 Причина: %s\n", escapeHTML(lead.Reason)))
	}

	sb.WriteString("\n<b>Клиент</b>\n")
	sb.WriteString(fmt.Sprintf("👤 %s\n", escapeHTML(firstNonEmpty(user.Name, "не указано"))))
	sb.WriteString(fmt.Sprintf("📞 %s\n", escapeHTML(firstNonEmpty(user.Phone, "не указан"))))

	if procedureTitle != "" {
		sb.WriteString(fmt.Sprintf("💉 Процедура: <b>%s</b>\n", escapeHTML(procedureTitle)))
	} else if booking.ProcedureID != 0 {
		sb.WriteString(fmt.Sprintf("💉 Процедура ID: %d\n", booking.ProcedureID))
	}

	sb.WriteString("\n<b>Дата и время</b>\n")
	if len(booking.RequestedDateTimes) > 0 {
		for _, dt := range booking.RequestedDateTimes {
			sb.WriteString(fmt.Sprintf("🗓 %s\n", escapeHTML(formatDatetime(dt))))
		}
	} else {
		sb.WriteString(fmt.Sprintf("🗓 %s\n", formatTime(booking.DateTime)))
	}

	comment := firstNonEmpty(booking.Comment, "без сообщения")
	sb.WriteString(fmt.Sprintf("\n💭 <i>%s</i>", escapeHTML(comment)))

	return sb.String()
}

func bookingUserText(title, subtitle string, booking models.Booking, procedureTitle string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("✨ <b>%s</b>\n\n", escapeHTML(title)))
	sb.WriteString(escapeHTML(subtitle) + "\n")

	if procedureTitle != "" {
		sb.WriteString(fmt.Sprintf("\n💆 Процедура: <b>%s</b>\n", escapeHTML(procedureTitle)))
	}
	sb.WriteString(fmt.Sprintf("🗓 Дата и время: <b>%s</b>\n", formatTime(booking.DateTime)))

	if booking.Comment != "" {
		sb.WriteString(fmt.Sprintf("\n💬 Ваш комментарий: <i>%s</i>\n", escapeHTML(booking.Comment)))
	}

	sb.WriteString("\n<i>С уважением, команда Plasma Glow 💫</i>")
	return sb.String()
}

func reminderAdminText(user models.User, booking models.Booking, procedureTitle string, now time.Time) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔔 <b>Напоминание о записи %s</b>\n\n", relativeBookingDay(booking.DateTime, now)))
	sb.WriteString(fmt.Sprintf("👤 %s\n", escapeHTML(firstNonEmpty(user.Name, "не указано"))))
	sb.WriteString(fmt.Sprintf("📞 %s\n", escapeHTML(firstNonEmpty(user.Phone, "не указан"))))
	sb.WriteString(fmt.Sprintf("🗓 %s\n", formatTime(booking.DateTime)))
	if procedureTitle != "" {
		sb.WriteString(fmt.Sprintf("💉 Процедура: <b>%s</b>\n", escapeHTML(procedureTitle)))
	} else if booking.ProcedureID != 0 {
		sb.WriteString(fmt.Sprintf("💉 Процедура ID: %d\n", booking.ProcedureID))
	}
	return sb.String()
}

func reminderUserText(booking models.Booking, procedureTitle string, now time.Time) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("🔔 <b>Напоминание о записи</b>\n\n"))
	sb.WriteString(fmt.Sprintf("Здравствуйте! Ждём вас %s в Plasma Glow!\n", relativeBookingDay(booking.DateTime, now)))

	if procedureTitle != "" {
		sb.WriteString(fmt.Sprintf("\n💆 Процедура: <b>%s</b>\n", escapeHTML(procedureTitle)))
	}
	sb.WriteString(fmt.Sprintf("🗓 Дата и время: <b>%s</b>\n", formatTime(booking.DateTime)))

	sb.WriteString("\n<i>До встречи! 💫</i>")
	return sb.String()
}

func relativeBookingDay(bookingTime, now time.Time) string {
	loc := telegramLocation()
	bookingLocal := bookingTime.In(loc)
	nowLocal := now.In(loc)
	bookingDay := time.Date(bookingLocal.Year(), bookingLocal.Month(), bookingLocal.Day(), 0, 0, 0, 0, loc)
	today := time.Date(nowLocal.Year(), nowLocal.Month(), nowLocal.Day(), 0, 0, 0, 0, loc)
	switch {
	case bookingDay.Equal(today):
		return "сегодня"
	case bookingDay.Equal(today.AddDate(0, 0, 1)):
		return "завтра"
	default:
		return "скоро"
	}
}

func monthlyRevenueText(summary models.MonthlyRevenue) string {
	period := summary.PeriodStart.In(telegramLocation()).Format("01.2006")
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("<b>Выручка за %s</b>\n\n", escapeHTML(period)))
	sb.WriteString(fmt.Sprintf("Итого: <b>%s ₽</b>\n", formatMoney(summary.NetAmount)))
	sb.WriteString(fmt.Sprintf("Завершённых записей: <b>%d</b>", summary.CompletedCount))
	if summary.BonusUsed > 0 {
		sb.WriteString(fmt.Sprintf("\n\nДо бонусов: %s ₽\nСписано бонусов: %s ₽", formatMoney(summary.GrossAmount), formatMoney(summary.BonusUsed)))
	}
	return sb.String()
}

func formatMoney(amount int) string {
	sign := ""
	if amount < 0 {
		sign = "-"
		amount = -amount
	}
	raw := fmt.Sprintf("%d", amount)
	if len(raw) <= 3 {
		return sign + raw
	}
	groups := make([]string, 0, len(raw)/3+1)
	for len(raw) > 3 {
		groups = append([]string{raw[len(raw)-3:]}, groups...)
		raw = raw[:len(raw)-3]
	}
	groups = append([]string{raw}, groups...)
	return sign + strings.Join(groups, " ")
}

func compactSMS(text string) string {
	text = stripHTMLTags(text)
	text = strings.ReplaceAll(text, "\n\n", "\n")
	runes := []rune(strings.TrimSpace(text))
	if len(runes) > 480 {
		return string(runes[:480])
	}
	return text
}

func stripHTMLTags(s string) string {
	var result strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			result.WriteRune(r)
		}
	}
	return result.String()
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func formatDatetime(raw string) string {
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return raw
	}
	return formatTime(t)
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "не указано"
	}
	return t.In(telegramLocation()).Format("02.01.2006 15:04")
}

func telegramLocation() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.FixedZone("MSK", 3*60*60)
	}
	return loc
}
