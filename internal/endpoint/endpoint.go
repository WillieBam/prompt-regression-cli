package endpoint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"prompt-regression-cli/types"
)

const ChatCompletionPath = "/chat/completions"

func ChatCompletionURL(baseURL string) string {
	return fmt.Sprintf("%s%s", strings.TrimRight(baseURL, "/"), ChatCompletionPath)
}

func NewChatCompletionHTTPRequest(ctx context.Context, baseURL, apiKey string, req types.ChatCompletionRequest) (*http.Request, error) {
	bodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := ChatCompletionURL(baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.
		NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	return httpReq, nil
}
