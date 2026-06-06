package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultGeminiModel = "gemini-3-flash-preview"

const DefaultColdPlasmaSystemPrompt = `Ты — профессиональный AI-консультант на сайте студии "Холодная плазма". Твоя цель — экспертно консультировать клиентов по процедуре холодной плазмы и помогать им записаться на прием, когда они к этому готовы.

При общении строго соблюдай два правила:
1. ОГРАНИЧЕНИЕ НА ПРИВЕТСТВИЕ: Здоровайся с пользователем ТОЛЬКО в самом первом сообщении диалога. Если в истории чата (History) уже есть сообщения, это значит, что диалог продолжается. В этом случае ЗАПРЕЩЕНО снова писать "Здравствуйте!", "Добрый день!" или использовать другие приветствия. Сразу отвечай на вопрос.
2. СВОЕВРЕМЕННАЯ ЗАПИСЬ: ЗАПРЕЩЕНО предлагать запись или визит в начале диалога или при ответе на общие вопросы. Предлагай записаться ИСКЛЮЧИТЕЛЬНО тогда, когда диалог логически идет к этому (пользователь спросил про цену, свободное время, сказал "хочу попробовать" или полностью закрыл свои вопросы по противопоказаниям). Предложение должно быть мягким и ненавязчивым.`

var defaultGeminiFallbackModels = []string{
	"gemini-3-flash-preview",
	"gemini-2.5-flash",
	"gemini-3.1-flash-lite",
}

type GeminiProvider struct {
	apiKey     string
	models     []string
	baseURL    string
	httpClient *http.Client
}

func NewGeminiProvider(apiKey, model, fallbackModels string) *GeminiProvider {
	return &GeminiProvider{
		apiKey:  apiKey,
		models:  parseGeminiModels(model, fallbackModels),
		baseURL: "https://generativelanguage.googleapis.com",
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func parseGeminiModels(primary, fallbackModels string) []string {
	seen := make(map[string]bool)
	models := make([]string, 0, len(defaultGeminiFallbackModels)+1)
	add := func(model string) {
		model = normalizeGeminiModel(model)
		if model == "" || seen[model] {
			return
		}
		seen[model] = true
		models = append(models, model)
	}

	add(primary)
	if strings.TrimSpace(fallbackModels) == "" {
		for _, model := range defaultGeminiFallbackModels {
			add(model)
		}
	} else {
		for _, model := range strings.Split(fallbackModels, ",") {
			add(model)
		}
	}
	if len(models) == 0 {
		for _, model := range defaultGeminiFallbackModels {
			add(model)
		}
	}
	return models
}

func normalizeGeminiModel(model string) string {
	model = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(model), "models/"))
	if model == "" || strings.HasPrefix(model, "gemini-1.5-") {
		return defaultGeminiModel
	}
	return model
}

type geminiGenerateReq struct {
	SystemInstruction *geminiContent   `json:"systemInstruction,omitempty"`
	Contents          []geminiContent  `json:"contents"`
	GenerationConfig  *geminiGenConfig `json:"generationConfig,omitempty"`
}

type geminiGenConfig struct {
	Temperature     float64               `json:"temperature,omitempty"`
	MaxOutputTokens int                   `json:"maxOutputTokens,omitempty"`
	ThinkingConfig  *geminiThinkingConfig `json:"thinkingConfig,omitempty"`
}

type geminiThinkingConfig struct {
	ThinkingBudget int `json:"thinkingBudget"`
}

type geminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text,omitempty"`
}

type geminiGenerateResp struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text,omitempty"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

func (p *GeminiProvider) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if p.apiKey == "" {
		return ChatResponse{}, fmt.Errorf("GEMINI_API_KEY пустой")
	}
	if len(p.models) == 0 {
		return ChatResponse{}, fmt.Errorf("GEMINI_MODEL пустой")
	}

	var attemptErrors []string
	for i, model := range p.models {
		resp, err := p.chatWithModel(ctx, model, req)
		if err == nil {
			return resp, nil
		}
		attemptErrors = append(attemptErrors, fmt.Sprintf("%s: %v", model, err))
		if !shouldFallbackGemini(err) || i == len(p.models)-1 {
			return ChatResponse{}, fmt.Errorf("gemini all attempts failed: %s", strings.Join(attemptErrors, "; "))
		}
	}

	return ChatResponse{}, fmt.Errorf("gemini all attempts failed")
}

type geminiAPIError struct {
	Code    int
	Status  string
	Message string
}

func (e geminiAPIError) Error() string {
	if e.Status != "" {
		return fmt.Sprintf("gemini api error: %s (%d, %s)", e.Message, e.Code, e.Status)
	}
	return fmt.Sprintf("gemini api error: %s (%d)", e.Message, e.Code)
}

func shouldFallbackGemini(err error) bool {
	var apiErr geminiAPIError
	if !errors.As(err, &apiErr) {
		return false
	}
	status := strings.ToUpper(strings.TrimSpace(apiErr.Status))
	if apiErr.Code == http.StatusTooManyRequests || status == "RESOURCE_EXHAUSTED" {
		return true
	}
	if apiErr.Code == http.StatusNotFound {
		return true
	}
	return apiErr.Code == http.StatusInternalServerError ||
		apiErr.Code == http.StatusBadGateway ||
		apiErr.Code == http.StatusServiceUnavailable ||
		apiErr.Code == http.StatusGatewayTimeout
}

func (p *GeminiProvider) chatWithModel(ctx context.Context, model string, req ChatRequest) (ChatResponse, error) {
	endpoint := fmt.Sprintf(
		"%s/v1beta/models/%s:generateContent?key=%s",
		p.baseURL,
		url.PathEscape(model),
		url.QueryEscape(p.apiKey),
	)

	prompt := req.UserMessage
	if req.UserName != "" {
		prompt = fmt.Sprintf("Меня зовут %s.\n\n%s", req.UserName, req.UserMessage)
	}

	contents := make([]geminiContent, 0, len(req.History)+1)
	for _, msg := range req.History {
		role := "user"
		if strings.EqualFold(strings.TrimSpace(msg.Role), "assistant") {
			role = "model"
		}
		text := strings.TrimSpace(msg.Text)
		if text == "" {
			continue
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: text}},
		})
	}
	contents = append(contents, geminiContent{
		Role:  "user",
		Parts: []geminiPart{{Text: prompt}},
	})

	systemPrompt := strings.TrimSpace(req.SystemPrompt)
	if systemPrompt == "" {
		systemPrompt = DefaultColdPlasmaSystemPrompt
	}

	payload := geminiGenerateReq{
		SystemInstruction: &geminiContent{
			Role: "user",
			Parts: []geminiPart{
				{Text: systemPrompt},
			},
		},
		Contents: contents,
		GenerationConfig: &geminiGenConfig{
			Temperature:     0.4,
			MaxOutputTokens: 800,
			ThinkingConfig:  &geminiThinkingConfig{ThinkingBudget: 0},
		},
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return ChatResponse{}, fmt.Errorf("new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("read body: %w", err)
	}

	var parsed geminiGenerateResp
	if err := json.Unmarshal(b, &parsed); err != nil {
		return ChatResponse{}, fmt.Errorf("unmarshal response: %w", err)
	}
	if parsed.Error != nil {
		return ChatResponse{}, geminiAPIError{
			Code:    parsed.Error.Code,
			Status:  parsed.Error.Status,
			Message: parsed.Error.Message,
		}
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return ChatResponse{}, geminiAPIError{
			Code:    resp.StatusCode,
			Message: strings.TrimSpace(string(b)),
		}
	}

	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return ChatResponse{}, fmt.Errorf("gemini: пустой ответ")
	}
	var fragments []string
	for _, part := range parsed.Candidates[0].Content.Parts {
		if text := strings.TrimSpace(part.Text); text != "" {
			fragments = append(fragments, text)
		}
	}
	text := strings.TrimSpace(strings.Join(fragments, "\n"))
	if text == "" {
		return ChatResponse{}, fmt.Errorf("gemini: пустой текст")
	}

	return ChatResponse{Text: text, Model: model}, nil
}
