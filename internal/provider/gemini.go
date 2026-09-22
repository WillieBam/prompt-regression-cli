package provider

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type GeminiProvider struct {
	APIKey string
}

func NewGeminiProvider(apiKey string) *GeminiProvider {
	return &GeminiProvider{APIKey: apiKey}
}

func (g *GeminiProvider) Name() string {
	return "gemini"
}

func (g *GeminiProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	start := time.Now()

	// if no API key is provided, provide mock output for offline testing
	if g.APIKey == "" {
		select {
		case <-time.After(120 * time.Millisecond):
			// mock logic based on prompt keywords
			text := `{"category": "GENERAL", "urgency": "LOW"}`
			if strings.Contains(strings.ToLower(req.Prompt), "refund") || strings.Contains(strings.ToLower(req.Prompt), "charged") {
				text = `{"category": "BILLING", "urgency": "LOW"}`
			} else if strings.Contains(strings.ToLower(req.Prompt), "500") || strings.Contains(strings.ToLower(req.Prompt), "downtime") {
				text = `{"category": "TECHNICAL", "urgency": "HIGH"}`
			}
			return &CompletionResponse{
				Text: text,
				Metrics: CompletionMetrics{
					Latency:      time.Since(start),
					InputTokens:  len(req.Prompt) / 4,
					OutputTokens: len(text) / 4,
				},
			}, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// TODO: Replace with real HTTP request to https://generativelanguage.googleapis.com/v1beta/...
	return nil, fmt.Errorf("real Gemini API call not yet wired; unset API key to run mock mode")
}
