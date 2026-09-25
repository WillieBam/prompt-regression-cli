package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"prompt-regression-cli/internal/endpoint"
	"prompt-regression-cli/types"
	"strings"
	"time"
)

type GeminiProvider struct {
	APIKey     string
	BaseURL    string
	httpClient *http.Client
}

func NewGeminiProvider(apiKey, baseurl string) *GeminiProvider {
	if baseurl == "" {
		baseurl = "https://generativelanguage.googleapis.com"
	}
	return &GeminiProvider{
		APIKey:  apiKey,
		BaseURL: strings.TrimRight(baseurl, "/"),
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 20,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
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

	model := req.Model
	if model == "" {
		model = "gemini-3.5-flash"
	}

	payload := types.ChatCompletionRequest{
		Model: model,
		Messages: []types.ChatMessages{
			{
				Role:    "user",
				Content: req.Prompt,
			},
		},
		Temperature: req.Temperature,
	}

	httpReq, err := endpoint.NewChatCompletionHTTPRequest(ctx, g.BaseURL, g.APIKey, payload)
	if err != nil {
		return nil, err
	}

	resp, err := g.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("requested failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API Error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp types.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choice returned by provider")
	}

	return &CompletionResponse{
		Text: chatResp.Choices[0].Message.Content,
		Metrics: CompletionMetrics{
			Latency:      time.Since(start),
			InputTokens:  chatResp.Usage.PromptTokens,
			OutputTokens: chatResp.Usage.CompletionTokens,
		},
	}, nil
}

