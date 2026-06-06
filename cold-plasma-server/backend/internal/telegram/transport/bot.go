package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	maxMessageLen = 4096
	maxRetries    = 3
)

type BotClient struct {
	token  string
	chatID string
	client *http.Client
}

func NewBotClient(token, chatID string) *BotClient {
	return &BotClient{
		token:  strings.TrimSpace(token),
		chatID: strings.TrimSpace(chatID),
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (b *BotClient) Enabled() bool {
	return b.token != "" && b.chatID != ""
}

func (b *BotClient) SendMessage(ctx context.Context, text string) error {
	_, err := b.SendMessageResult(ctx, text)
	return err
}

func (b *BotClient) SendMessageResult(ctx context.Context, text string) ([]int, error) {
	if !b.Enabled() {
		return nil, nil
	}
	return b.SendMessageToThreadResult(ctx, b.chatID, "", text)
}

func (b *BotClient) SendMessageToDefaultThread(ctx context.Context, threadID, text string) error {
	_, err := b.SendMessageToDefaultThreadResult(ctx, threadID, text)
	return err
}

func (b *BotClient) SendMessageToDefaultThreadResult(ctx context.Context, threadID, text string) ([]int, error) {
	if !b.Enabled() {
		return nil, nil
	}
	return b.SendMessageToThreadResult(ctx, b.chatID, threadID, text)
}

func (b *BotClient) SendMessageTo(ctx context.Context, chatID, text string) error {
	_, err := b.SendMessageToThreadResult(ctx, chatID, "", text)
	return err
}

func (b *BotClient) SendMessageToThread(ctx context.Context, chatID, threadID, text string) error {
	_, err := b.SendMessageToThreadResult(ctx, chatID, threadID, text)
	return err
}

func (b *BotClient) SendMessageToThreadResult(ctx context.Context, chatID, threadID, text string) ([]int, error) {
	if strings.TrimSpace(b.token) == "" || strings.TrimSpace(chatID) == "" {
		return nil, nil
	}
	messageIDs := make([]int, 0, 1)
	for _, chunk := range splitMessage(text, maxMessageLen) {
		messageID, err := b.sendWithRetry(ctx, chatID, threadID, chunk)
		if err != nil {
			return messageIDs, err
		}
		if messageID != 0 {
			messageIDs = append(messageIDs, messageID)
		}
	}
	return messageIDs, nil
}

func (b *BotClient) sendWithRetry(ctx context.Context, chatID, threadID, text string) (int, error) {
	var lastErr error
	delay := 500 * time.Millisecond
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return 0, ctx.Err()
			case <-time.After(delay):
				delay *= 2
			}
		}
		var messageID int
		messageID, lastErr = b.doSend(ctx, chatID, threadID, text)
		if lastErr == nil {
			return messageID, nil
		}
		if isFatalTelegramErr(lastErr) {
			return 0, lastErr
		}
		log.Printf("telegram send attempt %d failed: %v", attempt+1, lastErr)
	}
	return 0, fmt.Errorf("telegram: all %d send attempts failed: %w", maxRetries, lastErr)
}

func (b *BotClient) doSend(ctx context.Context, chatID, threadID, text string) (int, error) {
	payload := map[string]any{
		"chat_id":    strings.TrimSpace(chatID),
		"text":       text,
		"parse_mode": "HTML",
	}
	if threadID = strings.TrimSpace(threadID); threadID != "" {
		id, err := strconv.Atoi(threadID)
		if err != nil {
			return 0, fmt.Errorf("telegram message_thread_id must be integer: %w", err)
		}
		payload["message_thread_id"] = id
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("telegram marshal: %w", err)
	}
	endpoint := "https://api.telegram.org/bot" + b.token + "/sendMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("telegram send: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return 0, &telegramAPIError{
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(raw)),
		}
	}
	var out struct {
		Result struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return 0, fmt.Errorf("telegram response decode: %w", err)
	}
	return out.Result.MessageID, nil
}

func (b *BotClient) DeleteDefaultMessage(ctx context.Context, messageID int) error {
	if !b.Enabled() || messageID <= 0 {
		return nil
	}
	return b.DeleteMessage(ctx, b.chatID, messageID)
}

func (b *BotClient) DeleteMessage(ctx context.Context, chatID string, messageID int) error {
	if strings.TrimSpace(b.token) == "" || strings.TrimSpace(chatID) == "" || messageID <= 0 {
		return nil
	}
	body, err := json.Marshal(map[string]any{
		"chat_id":    strings.TrimSpace(chatID),
		"message_id": messageID,
	})
	if err != nil {
		return fmt.Errorf("telegram delete marshal: %w", err)
	}
	endpoint := "https://api.telegram.org/bot" + b.token + "/deleteMessage"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram delete request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram delete: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return &telegramAPIError{
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(raw)),
		}
	}
	return nil
}

func splitMessage(text string, maxLen int) []string {
	runes := []rune(text)
	if len(runes) <= maxLen {
		return []string{text}
	}
	var parts []string
	for len(runes) > 0 {
		if len(runes) <= maxLen {
			parts = append(parts, string(runes))
			break
		}
		cut := maxLen
		for i := maxLen - 1; i > maxLen/2; i-- {
			if runes[i] == '\n' {
				cut = i + 1
				break
			}
		}
		parts = append(parts, string(runes[:cut]))
		runes = runes[cut:]
	}
	return parts
}

type telegramAPIError struct {
	StatusCode int
	Body       string
}

func (e *telegramAPIError) Error() string {
	return fmt.Sprintf("telegram status %d: %s", e.StatusCode, e.Body)
}

func isFatalTelegramErr(err error) bool {
	var tErr *telegramAPIError
	if errors.As(err, &tErr) {
		return tErr.StatusCode >= 400 && tErr.StatusCode < 500
	}
	return false
}
