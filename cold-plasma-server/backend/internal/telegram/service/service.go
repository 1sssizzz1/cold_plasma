package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"cold-plasma-server/internal/models"
	core "cold-plasma-server/internal/repository"
	"cold-plasma-server/internal/sms"
	tgrepo "cold-plasma-server/internal/telegram/repository"
	tgtransport "cold-plasma-server/internal/telegram/transport"
)

type Service struct {
	bookings         tgrepo.BookingRepository
	users            tgrepo.UserRepository
	procedures       tgrepo.ProcedureRepository
	settings         tgrepo.SettingsRepository
	sms              sms.Sender
	bot              *tgtransport.BotClient
	polling          *tgtransport.PollingClient
	botName          string
	reminderThreadID string
	revenueChatID    string
	revenueThreadID  string
}

func New(bookings tgrepo.BookingRepository, users tgrepo.UserRepository, procedures tgrepo.ProcedureRepository, settings tgrepo.SettingsRepository, smsSender sms.Sender, bot *tgtransport.BotClient, polling *tgtransport.PollingClient, botName, reminderThreadID, revenueChatID, revenueThreadID string) *Service {
	return &Service{
		bookings:         bookings,
		users:            users,
		procedures:       procedures,
		settings:         settings,
		sms:              smsSender,
		bot:              bot,
		polling:          polling,
		botName:          strings.TrimPrefix(strings.TrimSpace(botName), "@"),
		reminderThreadID: strings.TrimSpace(reminderThreadID),
		revenueChatID:    strings.TrimSpace(revenueChatID),
		revenueThreadID:  strings.TrimSpace(revenueThreadID),
	}
}

type LinkResult struct {
	Linked bool   `json:"linked"`
	URL    string `json:"url,omitempty"`
}

func (s *Service) LinkStatus(ctx context.Context, userID int64) (LinkResult, error) {
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return LinkResult{}, err
	}
	return LinkResult{Linked: strings.TrimSpace(u.TelegramChatID) != ""}, nil
}

func (s *Service) CreateLink(ctx context.Context, userID int64) (LinkResult, error) {
	if s == nil || s.users == nil {
		return LinkResult{}, fmt.Errorf("telegram service is not configured")
	}
	if s.botName == "" {
		return LinkResult{}, fmt.Errorf("TELEGRAM_BOT_USERNAME не задан")
	}
	u, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return LinkResult{}, err
	}
	if strings.TrimSpace(u.TelegramChatID) != "" {
		return LinkResult{Linked: true}, nil
	}
	token, err := randomLinkToken()
	if err != nil {
		return LinkResult{}, err
	}
	if _, err := s.users.CreateTelegramLinkToken(ctx, core.CreateTelegramLinkTokenParams{
		UserID:    userID,
		TokenHash: hashToken(token),
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}); err != nil {
		return LinkResult{}, err
	}
	return LinkResult{
		URL: "https://t.me/" + s.botName + "?start=" + token,
	}, nil
}

type Update struct {
	Message *Message `json:"message"`
}

type Message struct {
	Text string `json:"text"`
	Chat Chat   `json:"chat"`
	From From   `json:"from"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type From struct {
	Username string `json:"username"`
}

func (s *Service) HandleUpdate(ctx context.Context, update Update) error {
	if s == nil || update.Message == nil {
		return nil
	}
	text := strings.TrimSpace(update.Message.Text)
	if !strings.HasPrefix(text, "/start") {
		return nil
	}
	payload := strings.TrimSpace(strings.TrimPrefix(text, "/start"))
	if payload == "" {
		_ = s.sendToUserChat(ctx, update.Message.Chat.ID, "Откройте ссылку подключения Telegram из личного кабинета на сайте.")
		return nil
	}
	item, err := s.users.GetActiveTelegramLinkTokenByHash(ctx, hashToken(payload))
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			_ = s.sendToUserChat(ctx, update.Message.Chat.ID, "Ссылка подключения устарела или уже использована. Создайте новую ссылку в личном кабинете.")
			return nil
		}
		return err
	}
	if time.Now().After(item.ExpiresAt) {
		_ = s.sendToUserChat(ctx, update.Message.Chat.ID, "Ссылка подключения устарела. Создайте новую ссылку в личном кабинете.")
		return nil
	}
	chatID := fmt.Sprintf("%d", update.Message.Chat.ID)
	if err := s.users.LinkTelegram(ctx, item.UserID, chatID, update.Message.From.Username); err != nil {
		return err
	}
	_ = s.users.ConsumeTelegramLinkToken(ctx, item.ID)
	_ = s.sendToUserChat(ctx, update.Message.Chat.ID, "Telegram подключён. Теперь сюда будут приходить уведомления о ваших записях.")
	return nil
}

func (s *Service) NotifyBookingCreated(ctx context.Context, booking models.Booking) {
	if s == nil {
		return
	}
	settings := s.notificationSettings(ctx)
	user := s.getUser(ctx, booking.UserID)
	procedureTitle := s.getProcedureTitle(ctx, booking.ProcedureID)
	lead := s.analyzeLead(ctx, s.leadPayload(ctx, booking, user))
	text := bookingAdminText("Заявка на запись", user, booking, procedureTitle, lead)
	sent := false
	if settings["admin_notify_telegram"] == "true" {
		if s.bot == nil || !s.bot.Enabled() {
			log.Printf("telegram notify skipped: TELEGRAM_BOT_TOKEN or TELEGRAM_ADMIN_CHAT_ID is empty")
		} else if err := s.bot.SendMessage(ctx, text); err == nil {
			sent = true
		} else {
			log.Printf("telegram notify failed for booking %d: %v", booking.ID, err)
		}
	}
	if settings["admin_notify_sms"] == "true" && s.sms != nil && settings["admin_sms_phone"] != "" {
		if err := s.sms.SendText(ctx, settings["admin_sms_phone"], compactSMS(text)); err == nil {
			sent = true
		} else {
			log.Printf("admin sms notify failed for booking %d: %v", booking.ID, err)
		}
	}
	if sent {
		_ = s.bookings.MarkTelegramCreatedNotified(ctx, booking.ID)
	}
	s.notifyUserBooking(ctx, booking, "Заявка создана", "Мы получили вашу заявку и скоро подтвердим подходящее время.")
}

func (s *Service) NotifyUserBookingCreated(ctx context.Context, booking models.Booking) {
	s.notifyUserBooking(ctx, booking, "Заявка создана", "Мы получили вашу заявку и скоро подтвердим подходящее время.")
}

func (s *Service) NotifyUserBookingConfirmed(ctx context.Context, booking models.Booking) {
	s.notifyUserBooking(ctx, booking, "Запись подтверждена", "Администратор подтвердил вашу запись.")
}

func (s *Service) NotifyUserBookingRescheduled(ctx context.Context, booking models.Booking) {
	s.notifyUserBooking(ctx, booking, "Запись перенесена", "Администратор изменил дату и время записи.")
}

func (s *Service) NotifyUserBookingCompleted(ctx context.Context, booking models.Booking) {
	s.notifyUserBooking(ctx, booking, "Запись завершена", "Спасибо за визит! Будем рады видеть вас снова.")
	s.notifyMonthlyRevenue(ctx, booking.DateTime)
}

func (s *Service) notifyMonthlyRevenue(ctx context.Context, bookingTime time.Time) {
	if s == nil || s.bot == nil || s.bookings == nil || strings.TrimSpace(s.revenueChatID) == "" {
		return
	}
	loc := telegramLocation()
	if bookingTime.IsZero() {
		bookingTime = time.Now()
	}
	localTime := bookingTime.In(loc)
	periodStart := time.Date(localTime.Year(), localTime.Month(), 1, 0, 0, 0, 0, loc)
	periodEnd := periodStart.AddDate(0, 1, 0)

	summary, err := s.bookings.MonthlyRevenue(ctx, periodStart, periodEnd)
	if err != nil {
		log.Printf("telegram revenue notify failed: %v", err)
		return
	}
	text := monthlyRevenueText(summary)
	if err := s.bot.SendMessageToThread(ctx, s.revenueChatID, s.revenueThreadID, text); err != nil {
		log.Printf("telegram revenue notify send failed: %v", err)
	}
}

func (s *Service) notifyUserBooking(ctx context.Context, booking models.Booking, title, subtitle string) bool {
	if s == nil || s.bot == nil || !booking.NotifyTelegram {
		return false
	}
	user := s.getUser(ctx, booking.UserID)
	if strings.TrimSpace(user.TelegramChatID) == "" {
		return false
	}
	procedureTitle := s.getProcedureTitle(ctx, booking.ProcedureID)
	text := bookingUserText(title, subtitle, booking, procedureTitle)
	if err := s.bot.SendMessageTo(ctx, user.TelegramChatID, text); err != nil {
		log.Printf("telegram user notify failed for booking %d: %v", booking.ID, err)
		return false
	}
	return true
}

func (s *Service) SendDueReminders(ctx context.Context, now time.Time) {
	if s == nil || s.bot == nil {
		return
	}
	s.deleteDueAdminReminderMessages(ctx, now)

	loc := telegramLocation()
	localNow := now.In(loc)
	periodStart := time.Date(localNow.Year(), localNow.Month(), localNow.Day(), 0, 0, 0, 0, loc)
	items, err := s.bookings.ListTelegramReminderDue(ctx, periodStart, 50)
	if err != nil {
		log.Printf("telegram reminders: list failed: %v", err)
		return
	}
	for _, booking := range items {
		s.sendReminder(ctx, booking, now)
	}
}

func (s *Service) sendReminder(ctx context.Context, booking models.Booking, now time.Time) {
	user := s.getUser(ctx, booking.UserID)
	procedureTitle := s.getProcedureTitle(ctx, booking.ProcedureID)
	sent := false

	if s.bot.Enabled() {
		text := reminderAdminText(user, booking, procedureTitle, now)
		if messageIDs, err := s.bot.SendMessageToDefaultThreadResult(ctx, s.reminderThreadID, text); err == nil {
			sent = true
			s.scheduleAdminReminderDelete(ctx, booking, messageIDs)
		} else {
			log.Printf("telegram reminder admin notify failed for booking %d: %v", booking.ID, err)
		}
	}

	if booking.NotifyTelegram && strings.TrimSpace(user.TelegramChatID) != "" {
		text := reminderUserText(booking, procedureTitle, now)
		if err := s.bot.SendMessageTo(ctx, user.TelegramChatID, text); err == nil {
			sent = true
		} else {
			log.Printf("telegram reminder user notify failed for booking %d: %v", booking.ID, err)
		}
	}

	if sent {
		_ = s.bookings.MarkTelegramReminderSent(ctx, booking.ID)
	}
}

func (s *Service) scheduleAdminReminderDelete(ctx context.Context, booking models.Booking, messageIDs []int) {
	if s == nil || s.bot == nil || len(messageIDs) == 0 {
		return
	}
	deleteAt := reminderDeleteAt(booking.DateTime)
	if s.bookings != nil {
		if err := s.bookings.StoreTelegramAdminReminder(ctx, booking.ID, messageIDs, deleteAt); err != nil {
			log.Printf("telegram reminder admin store failed for booking %d: %v", booking.ID, err)
		}
	}
	delay := time.Until(deleteAt)
	if delay < 0 {
		delay = 0
	}
	go func(ids []int, wait time.Duration) {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		<-timer.C

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		s.deleteAdminReminderMessages(ctx, models.TelegramAdminReminderMessage{
			BookingID:  booking.ID,
			MessageIDs: ids,
			DeleteAt:   deleteAt,
		})
	}(append([]int(nil), messageIDs...), delay)
}

func (s *Service) deleteDueAdminReminderMessages(ctx context.Context, now time.Time) {
	if s == nil || s.bot == nil || s.bookings == nil {
		return
	}
	items, err := s.bookings.ListTelegramAdminReminderDeleteDue(ctx, now, 50)
	if err != nil {
		log.Printf("telegram reminder admin delete due list failed: %v", err)
		return
	}
	for _, item := range items {
		s.deleteAdminReminderMessages(ctx, item)
	}
}

func (s *Service) deleteAdminReminderMessages(ctx context.Context, item models.TelegramAdminReminderMessage) {
	failed := false
	for _, id := range item.MessageIDs {
		if err := s.bot.DeleteDefaultMessage(ctx, id); err != nil {
			failed = true
			log.Printf("telegram reminder admin delete failed for booking %d message %d: %v", item.BookingID, id, err)
		}
	}
	if !failed && s.bookings != nil {
		if err := s.bookings.MarkTelegramAdminReminderDeleted(ctx, item.BookingID); err != nil {
			log.Printf("telegram reminder admin mark deleted failed for booking %d: %v", item.BookingID, err)
		}
	}
}

func reminderDeleteAt(bookingTime time.Time) time.Time {
	loc := telegramLocation()
	localTime := bookingTime.In(loc)
	dayStart := time.Date(localTime.Year(), localTime.Month(), localTime.Day(), 0, 0, 0, 0, loc)
	return dayStart.AddDate(0, 0, 1)
}

func (s *Service) sendToUserChat(ctx context.Context, chatID int64, text string) error {
	if s == nil || s.bot == nil {
		return nil
	}
	return s.bot.SendMessageTo(ctx, fmt.Sprintf("%d", chatID), text)
}

func (s *Service) getUser(ctx context.Context, userID int64) models.User {
	if s.users == nil {
		return models.User{}
	}
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		log.Printf("telegram: getUser(%d) failed: %v", userID, err)
	}
	return user
}

func (s *Service) getProcedureTitle(ctx context.Context, procedureID int64) string {
	if s.procedures == nil || procedureID == 0 {
		return ""
	}
	procedure, err := s.procedures.GetByID(ctx, procedureID)
	if err != nil {
		log.Printf("telegram: getProcedure(%d) failed: %v", procedureID, err)
		return ""
	}
	return procedure.Title
}

func (s *Service) StartPolling(ctx context.Context) {
	if s == nil || s.polling == nil {
		return
	}
	log.Printf("telegram: starting polling mode")
	s.polling.StartPolling(ctx, func(ctx context.Context, update tgtransport.Update) error {
		// Преобразуем update из polling в формат service
		var serviceMsg *Message
		if update.Message != nil {
			serviceMsg = &Message{
				Text: update.Message.Text,
				Chat: Chat{
					ID: update.Message.Chat.ID,
				},
				From: From{
					Username: update.Message.From.Username,
				},
			}
		}
		serviceUpdate := Update{
			Message: serviceMsg,
		}
		return s.HandleUpdate(ctx, serviceUpdate)
	})
}

func (s *Service) leadPayload(ctx context.Context, booking models.Booking, user models.User) map[string]any {
	return map[string]any{
		"name":                user.Name,
		"phone":               user.Phone,
		"email":               user.Email,
		"message":             booking.Comment,
		"procedure_id":        booking.ProcedureID,
		"requested_datetimes": booking.RequestedDateTimes,
		"booking_id":          booking.ID,
	}
}

func (s *Service) notificationSettings(ctx context.Context) map[string]string {
	defaults := map[string]string{
		"admin_notify_telegram": "true",
		"admin_notify_sms":      "false",
		"admin_sms_phone":       "",
	}
	if s.settings == nil {
		return defaults
	}
	keys := make([]string, 0, len(defaults))
	for key := range defaults {
		keys = append(keys, key)
	}
	values, err := s.settings.GetSettings(ctx, keys)
	if err != nil {
		log.Printf("telegram: failed to load notification settings: %v", err)
		return defaults
	}
	for key, value := range values {
		defaults[key] = value
	}
	return defaults
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

func randomLinkToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate telegram link token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}
