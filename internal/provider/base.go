package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"prompt-regression-cli/internal/config"
	"prompt-regression-cli/internal/endpoint"
	"prompt-regression-cli/types"
)

type ProviderConfig struct {
	Name           string
	DefaultBaseURL string
	DefaultModel   string
	AllowMock      bool
}

type BaseProvider struct {
	config     ProviderConfig
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

var defaultHttpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	},
}

func NewBaseProvider(cfg ProviderConfig, apiKey, baseURL string) *BaseProvider {
	prefix := strings.ToUpper(cfg.Name)
	appCfg := config.Get()

	resolvedKey := resolveValue(
		apiKey,
		os.Getenv(prefix+"_API_KEY"),
		appCfg.APIKey,
		os.Getenv("LLM_API_KEY"),
		os.Getenv("PROMPTCTL_API_KEY"),
	)

	resolvedURL := resolveValue(
		baseURL,
		os.Getenv(prefix+"_BASE_URL"),
		appCfg.BaseURL,
		os.Getenv("LLM_BASE_URL"),
		cfg.DefaultBaseURL,
	)

	return &BaseProvider{
		config:     cfg,
		apiKey:     resolvedKey,
		baseURL:    strings.TrimRight(resolvedURL, "/"),
		httpClient: defaultHttpClient,
	}
}

func (b *BaseProvider) Name() string {
	return b.config.Name
}

func (b *BaseProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	start := time.Now()

	if b.config.AllowMock && b.apiKey == "" {
		return mockComplete(ctx, req, start)
	}

	if b.baseURL == "" {
		return nil, fmt.Errorf("no base URL configured for provider %q (set LLM_BASE_URL in .env or via config)", b.config.Name)
	}

	prefix := strings.ToUpper(b.config.Name)
	appCfg := config.Get()

	model := resolveValue(
		req.Model,
		os.Getenv(prefix+"_MODEL"),
		appCfg.Model,
		os.Getenv("LLM_MODEL"),
		b.config.DefaultModel,
	)

	if model == "" {
		return nil, fmt.Errorf("no model specified for provider %q (set in suite YAML or LLM_MODEL in .env)", b.config.Name)
	}

	payload := types.ChatCompletionRequest{
		Model: model,
		Messages: []types.ChatMessages{
			{Role: "user", Content: req.Prompt},
		},
		Temperature: req.Temperature,
	}

	httpReq, err := endpoint.NewChatCompletionHTTPRequest(ctx, b.baseURL, b.apiKey, payload)
	if err != nil {
		return nil, fmt.Errorf("%s request preparation failed: %w", b.config.Name, err)
	}

	resp, err := b.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %w", b.config.Name, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s response: %w", b.config.Name, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s API error (status %d): %s", b.config.Name, resp.StatusCode, string(respBody))
	}

	var chatResp types.ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse %s response: %w", b.config.Name, err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choice returned by %s", b.config.Name)
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

func resolveValue(explicit string, sources ...string) string {
	if explicit != "" {
		return explicit
	}
	for _, src := range sources {
		if strings.TrimSpace(src) != "" {
			return strings.TrimSpace(src)
		}
	}
	return ""
}

func mockComplete(ctx context.Context, mockReq CompletionRequest, startTime time.Time) (*CompletionResponse, error) {
	select {
	case <-time.After(120 * time.Millisecond):
		text := `{"category": "GENERAL", "urgency": "LOW"}`
		lower := strings.ToLower(mockReq.Prompt)
		if strings.Contains(lower, "refund") || strings.Contains(lower, "charged") {
			text = `{"category": "BILLING", "urgency": "LOW"}`
		} else if strings.Contains(lower, "500") || strings.Contains(lower, "downtime") {
			text = `{"category": "TECHNICAL", "urgency": "HIGH"}`
		}
		return &CompletionResponse{
			Text: text,
			Metrics: CompletionMetrics{
				Latency:      time.Since(startTime),
				InputTokens:  len(mockReq.Prompt) / 4,
				OutputTokens: len(text) / 4,
			},
		}, nil

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
