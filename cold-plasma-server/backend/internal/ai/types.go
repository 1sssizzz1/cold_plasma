package ai

import "context"

type ChatRequest struct {
	SystemPrompt string
	UserMessage  string
	UserName     string
	History      []ChatMessage
}

type ChatResponse struct {
	Text  string
	Model string
}

type ChatMessage struct {
	Role string
	Text string
}

type Provider interface {
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
}
