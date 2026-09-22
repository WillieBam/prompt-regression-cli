package provider

import (
	"context"
	"time"
)

type CompletionRequest struct {
	Model       string
	Prompt      string
	Temperature float32
	MaxTokens   int
}

type CompletionMetrics struct {
	Latency          time.Duration
	TimeToFirstToken time.Duration
	InputTokens      int
	OutputTokens     int
}

type CompletionResponse struct {
	Text    string
	Metrics CompletionMetrics
}

type ModelProvider interface {
	Name() string
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
}
