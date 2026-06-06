package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type OpenAIProvider struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAIProvider(apiKey, model, baseURL string) *OpenAIProvider {
	baseURL = strings.TrimRight(baseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

type openAIChatReq struct {
	Model       string          `json:"model"`
	Messages    []openAIChatMsg `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
}

type openAIChatMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIChatResp struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if p.apiKey == "" {
		return ChatResponse{}, fmt.Errorf("OPENAI_API_KEY пустой")
	}
	if p.model == "" {
		return ChatResponse{}, fmt.Errorf("OPENAI_MODEL пустой")
	}

	userMsg := req.UserMessage
	if req.UserName != "" {
		userMsg = fmt.Sprintf("Меня зовут %s.\n\n%s", req.UserName, req.UserMessage)
	}

	payload := openAIChatReq{
		Model:       p.model,
		Messages:    buildOpenAIMessages(req.SystemPrompt, req.History, userMsg),
		Temperature: 0.4,
		MaxTokens:   350,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshal: %w", err)
	}

	endpoint := p.baseURL + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("read body: %w", err)
	}

	var parsed openAIChatResp
	if err := json.Unmarshal(b, &parsed); err != nil {
		return ChatResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}
	if parsed.Error != nil {
		return ChatResponse{}, fmt.Errorf("openai api error: %s", parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("openai: пустой ответ")
	}
	text := strings.TrimSpace(parsed.Choices[0].Message.Content)
	if text == "" {
		return ChatResponse{}, fmt.Errorf("openai: пустой текст")
	}
	return ChatResponse{Text: text, Model: p.model}, nil
}

func buildOpenAIMessages(systemPrompt string, history []ChatMessage, userMsg string) []openAIChatMsg {
	messages := make([]openAIChatMsg, 0, len(history)+2)
	messages = append(messages, openAIChatMsg{Role: "system", Content: systemPrompt})
	for _, msg := range history {
		role := strings.ToLower(strings.TrimSpace(msg.Role))
		if role != "assistant" {
			role = "user"
		}
		text := strings.TrimSpace(msg.Text)
		if text == "" {
			continue
		}
		messages = append(messages, openAIChatMsg{Role: role, Content: text})
	}
	messages = append(messages, openAIChatMsg{Role: "user", Content: userMsg})
	return messages
}
