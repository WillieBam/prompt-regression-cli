package provider

import (
	"os"
	"strings"

	"prompt-regression-cli/internal/config"
)

func NewProvider(name, apiKey, baseURL string) (ModelProvider, error) {
	resolvedName := strings.ToLower(strings.TrimSpace(name))
	if resolvedName == "" {
		resolvedName = strings.ToLower(strings.TrimSpace(config.Get().Provider))
	}
	if resolvedName == "" {
		resolvedName = strings.ToLower(strings.TrimSpace(os.Getenv("LLM_PROVIDER")))
	}
	if resolvedName == "" {
		resolvedName = "gemini"
	}

	return NewBaseProvider(ProviderConfig{
		Name:      resolvedName,
		AllowMock: true,
	}, apiKey, baseURL), nil
}
