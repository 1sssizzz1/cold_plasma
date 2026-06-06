package sms

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type HTTPSender struct {
	url    string
	token  string
	from   string
	client *http.Client
}

func NewHTTPSender(url, token, from string) *HTTPSender {
	return &HTTPSender{
		url:   url,
		token: token,
		from:  from,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (s *HTTPSender) SendCode(ctx context.Context, phone, code string) error {
	return s.SendText(ctx, phone, "Код подтверждения: "+code)
}

func (s *HTTPSender) SendText(ctx context.Context, phone, text string) error {
	if s.url == "" {
		return fmt.Errorf("SMS_HTTP_URL пустой")
	}
	body, err := json.Marshal(map[string]string{
		"phone": phone,
		"text":  text,
		"from":  s.from,
	})
	if err != nil {
		return fmt.Errorf("marshal sms request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new sms request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if s.token != "" {
		req.Header.Set("Authorization", "Bearer "+s.token)
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send sms: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("send sms: status %d", resp.StatusCode)
	}
	return nil
}
