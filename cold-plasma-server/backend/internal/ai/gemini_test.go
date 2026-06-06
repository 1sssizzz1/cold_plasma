package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGeminiProviderFallsBackOnQuota(t *testing.T) {
	var called []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = append(called, r.URL.Path)
		switch {
		case strings.Contains(r.URL.Path, "gemini-3-flash-preview"):
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"code":429,"message":"quota exceeded","status":"RESOURCE_EXHAUSTED"}}`))
		case strings.Contains(r.URL.Path, "gemini-2.5-flash"):
			_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"Ответ со второй модели"}]}}]}`))
		default:
			t.Fatalf("unexpected model path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	provider := NewGeminiProvider(
		"test-key",
		"gemini-3-flash-preview",
		"gemini-3-flash-preview,gemini-2.5-flash,gemini-3.1-flash-lite",
	)
	provider.baseURL = server.URL
	provider.httpClient = server.Client()

	resp, err := provider.Chat(context.Background(), ChatRequest{
		SystemPrompt: "Ты консультант.",
		UserMessage:  "Что такое холодная плазма?",
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if resp.Model != "gemini-2.5-flash" {
		t.Fatalf("expected fallback model gemini-2.5-flash, got %q", resp.Model)
	}
	if resp.Text != "Ответ со второй модели" {
		t.Fatalf("unexpected text: %q", resp.Text)
	}
	if len(called) != 2 {
		t.Fatalf("expected 2 model attempts, got %d", len(called))
	}
}

func TestGeminiProviderUsesDefaultSystemPromptWhenEmpty(t *testing.T) {
	var gotPrompt string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload geminiGenerateReq
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if payload.SystemInstruction != nil && len(payload.SystemInstruction.Parts) > 0 {
			gotPrompt = payload.SystemInstruction.Parts[0].Text
		}
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"parts":[{"text":"Ответ"}]}}]}`))
	}))
	defer server.Close()

	provider := NewGeminiProvider("test-key", "gemini-3-flash-preview", "")
	provider.baseURL = server.URL
	provider.httpClient = server.Client()

	_, err := provider.Chat(context.Background(), ChatRequest{
		UserMessage: "Что такое холодная плазма?",
	})
	if err != nil {
		t.Fatalf("Chat returned error: %v", err)
	}
	if gotPrompt != DefaultColdPlasmaSystemPrompt {
		t.Fatalf("expected default system prompt, got %q", gotPrompt)
	}
}
