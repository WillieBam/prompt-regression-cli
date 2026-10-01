package provider

import "testing"

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name         string
		wantProvider string
		wantErr      bool
	}{
		{name: "", wantProvider: "gemini"},
		{name: "gemini", wantProvider: "gemini"},
		{name: "openai", wantProvider: "openai"},
		{name: "unknown", wantProvider: "unknown", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewProvider(tt.name, "test-key", "")
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewProvider() error = %v, wantErr %t", err, tt.wantErr)
			}
			if err == nil && p.Name() != tt.wantProvider {
				t.Errorf("NewProvider() name = %q, want %q", p.Name(), tt.wantProvider)
			}
		})
	}
}

func TestProviderEnvResolution(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "test-openai-key")
	t.Setenv("LLM_BASE_URL", "https://custom.endpoint.com")

	p := NewOpenAIProvider("", "")
	if p.apiKey != "test-openai-key" {
		t.Errorf("expected apiKey %q, got %q", "test-openai-key", p.apiKey)
	}
	if p.baseURL != "https://custom.endpoint.com" {
		t.Errorf("expected baseURL %q, got %q", "https://custom.endpoint.com", p.baseURL)
	}
}

func TestProviderDefaultBaseURL(t *testing.T) {
	p := NewGeminiProvider("key", "")
	if p.baseURL != "https://generativelanguage.googleapis.com" {
		t.Errorf("expected default gemini baseURL, got %q", p.baseURL)
	}
}
