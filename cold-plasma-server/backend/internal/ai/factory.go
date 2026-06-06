package ai

import (
	"strings"

	"cold-plasma-server/config"
)

func NewProvider(cfg *config.Config) Provider {
	switch strings.ToLower(strings.TrimSpace(cfg.AIProvider)) {
	case "openai":
		return NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIModel, cfg.OpenAIBaseURL)
	default:
		return NewGeminiProvider(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiFallbackModels)
	}
}
