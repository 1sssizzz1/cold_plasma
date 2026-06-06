package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// PollingClient опрашивает Telegram API для получения обновлений (альтернатива webhook).
// Используется когда webhook недоступен (блокировки, firewall, локальная разработка).
type PollingClient struct {
	token  string
	client *http.Client
	offset int64
}

func NewPollingClient(token string) *PollingClient {
	return &PollingClient{
		token:  token,
		client: &http.Client{Timeout: 60 * time.Second}, // Long polling timeout
		offset: 0,
	}
}

type GetUpdatesResponse struct {
	Ok     bool     `json:"ok"`
	Result []Update `json:"result"`
}

type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

type Message struct {
	MessageID int64  `json:"message_id"`
	From      From   `json:"from"`
	Chat      Chat   `json:"chat"`
	Date      int64  `json:"date"`
	Text      string `json:"text"`
}

type From struct {
	ID           int64  `json:"id"`
	IsBot        bool   `json:"is_bot"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Username     string `json:"username"`
	LanguageCode string `json:"language_code"`
}

type Chat struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// GetUpdates получает новые обновления от Telegram API (long polling).
func (p *PollingClient) GetUpdates(ctx context.Context) ([]Update, error) {
	if p.token == "" {
		return nil, nil
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30", p.token, p.offset)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("telegram API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result GetUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if !result.Ok {
		return nil, fmt.Errorf("telegram API returned ok=false")
	}

	// Обновляем offset для следующего запроса
	if len(result.Result) > 0 {
		lastUpdate := result.Result[len(result.Result)-1]
		p.offset = lastUpdate.UpdateID + 1
	}

	return result.Result, nil
}

// StartPolling запускает бесконечный цикл опроса Telegram API.
// Вызывает handler для каждого полученного обновления.
func (p *PollingClient) StartPolling(ctx context.Context, handler func(context.Context, Update) error) {
	log.Printf("telegram polling: started")

	for {
		select {
		case <-ctx.Done():
			log.Printf("telegram polling: stopped")
			return
		default:
		}

		updates, err := p.GetUpdates(ctx)
		if err != nil {
			log.Printf("telegram polling: get updates failed: %v", err)
			// Ждём перед повторной попыткой
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
				continue
			}
		}

		for _, update := range updates {
			if err := handler(ctx, update); err != nil {
				log.Printf("telegram polling: handle update %d failed: %v", update.UpdateID, err)
			}
		}
	}
}
