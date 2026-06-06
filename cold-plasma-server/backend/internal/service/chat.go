package service

import (
	"context"
	"fmt"
	"strings"

	"cold-plasma-server/config"
	"cold-plasma-server/internal/ai"
	"cold-plasma-server/internal/repository"
	"cold-plasma-server/internal/security"
)

type ChatService struct {
	cfg      *config.Config
	provider ai.Provider
	chats    repository.ChatRepository
	users    repository.UserRepository
	crypto   *security.TextCipher
}

type ChatActor struct {
	UserID *int64
	Name   string
	Email  string
	Phone  string
}

type ChatAnswer struct {
	Text   string
	Model  string
	Intent string
}

func NewChatService(cfg *config.Config, provider ai.Provider, chats repository.ChatRepository, users repository.UserRepository, crypto *security.TextCipher) *ChatService {
	return &ChatService{cfg: cfg, provider: provider, chats: chats, users: users, crypto: crypto}
}

func (s *ChatService) OpenSession(ctx context.Context, actor ChatActor, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("session_id is required: %w", ErrValidation)
	}
	actor = s.enrichActor(ctx, actor)
	return s.chats.UpsertSession(ctx, repository.UpsertChatSessionParams{
		ID:        sessionID,
		UserID:    actor.UserID,
		UserName:  actor.Name,
		UserEmail: actor.Email,
		UserPhone: actor.Phone,
	})
}

func (s *ChatService) CloseSession(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("session_id is required: %w", ErrValidation)
	}
	return s.chats.CloseSession(ctx, sessionID)
}

func (s *ChatService) systemPrompt() string {
	return fmt.Sprintf(
		`%s

Правила ответов:
- Кратко: 2–4 коротких предложения, до 450 символов.
- Без длинных вступлений, списков и медицинских диагнозов.
- Тёпло и понятно, как говорит косметолог.
- Не уходи в посторонние темы: если вопрос не про холодную плазму или уход, мягко верни к теме.

Локальный контекст:
- Название студии: %s
- Адрес: %s
- Режим работы: %s
- Как добраться: %s
- Парковка: %s
- Частые вопросы: %s`,
		ai.DefaultColdPlasmaSystemPrompt,
		s.cfg.SalonName,
		s.cfg.SalonAddress,
		s.cfg.SalonWorkHours,
		s.cfg.SalonDirections,
		s.cfg.SalonParking,
		s.cfg.SalonPopularFAQ,
	)
}

func (s *ChatService) Ask(ctx context.Context, actor ChatActor, sessionID string, message string, history []ai.ChatMessage) (ChatAnswer, error) {
	message = strings.TrimSpace(message)
	sessionID = strings.TrimSpace(sessionID)
	actor = s.enrichActor(ctx, actor)
	if message == "" {
		return ChatAnswer{}, fmt.Errorf("message обязателен: %w", ErrValidation)
	}

	if sessionID != "" {
		if err := s.chats.UpsertSession(ctx, repository.UpsertChatSessionParams{
			ID:        sessionID,
			UserID:    actor.UserID,
			UserName:  actor.Name,
			UserEmail: actor.Email,
			UserPhone: actor.Phone,
		}); err != nil {
			return ChatAnswer{}, err
		}
	}

	intent := detectChatIntent(message)
	resp, err := s.provider.Chat(ctx, ai.ChatRequest{
		SystemPrompt: s.systemPrompt(),
		UserMessage:  message,
		UserName:     actor.Name,
		History:      normalizeChatHistory(history),
	})
	if err != nil {
		return ChatAnswer{}, se(ErrValidation, "AI-консультант временно недоступен: "+err.Error())
	}

	if params, ok := s.encryptedLogParams(actor, sessionID, message, resp.Text, resp.Model, intent); ok {
		_ = s.chats.AddChatLog(ctx, params)
	}
	return ChatAnswer{Text: resp.Text, Model: resp.Model, Intent: intent}, nil
}

func (s *ChatService) encryptedLogParams(actor ChatActor, sessionID, input, output, model, intent string) (repository.CreateChatLogParams, bool) {
	if s.crypto == nil {
		return repository.CreateChatLogParams{}, false
	}
	encryptedInput, err := s.crypto.Encrypt(input)
	if err != nil {
		return repository.CreateChatLogParams{}, false
	}
	encryptedOutput, err := s.crypto.Encrypt(output)
	if err != nil {
		return repository.CreateChatLogParams{}, false
	}
	return repository.CreateChatLogParams{
		SessionID: sessionID,
		UserID:    actor.UserID,
		UserName:  actor.Name,
		UserEmail: actor.Email,
		UserPhone: actor.Phone,
		RawInput:  encryptedInput,
		RawOutput: encryptedOutput,
		AIModel:   model,
		Intent:    intent,
	}, true
}

func (s *ChatService) enrichActor(ctx context.Context, actor ChatActor) ChatActor {
	actor.Name = strings.TrimSpace(actor.Name)
	actor.Email = strings.TrimSpace(strings.ToLower(actor.Email))
	actor.Phone = strings.TrimSpace(actor.Phone)

	if s.users == nil || actor.UserID == nil {
		return actor
	}
	u, err := s.users.GetByID(ctx, *actor.UserID)
	if err != nil {
		return actor
	}
	actor.Name = strings.TrimSpace(u.Name)
	actor.Email = strings.TrimSpace(strings.ToLower(u.Email))
	actor.Phone = strings.TrimSpace(u.Phone)
	return actor
}

func normalizeChatHistory(history []ai.ChatMessage) []ai.ChatMessage {
	if len(history) > 12 {
		history = history[len(history)-12:]
	}
	out := make([]ai.ChatMessage, 0, len(history))
	for _, msg := range history {
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		if role != "assistant" {
			role = "user"
		}
		text := strings.TrimSpace(msg.Text)
		if text == "" {
			continue
		}
		if len([]rune(text)) > 1000 {
			runes := []rune(text)
			text = string(runes[:1000])
		}
		out = append(out, ai.ChatMessage{Role: role, Text: text})
	}
	return out
}

func detectChatIntent(message string) string {
	v := strings.ToLower(strings.TrimSpace(message))
	bookingWords := []string{
		"запис", "запиш", "заброни", "бронь", "хочу прийти", "хочу на процедуру",
		"можно к вам", "свободное время", "выбрать время", "нужна запись", "оставить заявку",
	}
	for _, word := range bookingWords {
		if strings.Contains(v, word) {
			return "booking"
		}
	}
	return ""
}
