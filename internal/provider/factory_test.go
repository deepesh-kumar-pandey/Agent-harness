package provider

import (
	"testing"
)

// TestNewProvider verifies provider selection and configuration.
func TestNewProvider(t *testing.T) {
	testCases := []struct {
		name             string
		provName         string
		baseURL          string
		apiKey           string
		wantProviderType string
		wantBaseURL      string
		wantAPIKey       string
		expectErr        bool
	}{
		{
			name:             "ollama selects OllamaProvider",
			provName:         "ollama",
			baseURL:          "http://localhost:11434",
			wantProviderType: "ollama",
			wantBaseURL:      "http://localhost:11434",
		},
		{
			name:             "openai selects OpenAIProvider",
			provName:         "openai",
			baseURL:          "https://api.openai.com/v1",
			apiKey:           "sk-test-key",
			wantProviderType: "openai",
			wantBaseURL:      "https://api.openai.com/v1",
			wantAPIKey:       "sk-test-key",
		},
		{
			name:             "gemini selects GeminiProvider",
			provName:         "gemini",
			baseURL:          "https://generativelanguage.googleapis.com/v1beta",
			apiKey:           "gemini-test-key",
			wantProviderType: "gemini",
			wantBaseURL:      "https://generativelanguage.googleapis.com/v1beta",
			wantAPIKey:       "gemini-test-key",
		},
		{
			name:             "anthropic selects AnthropicProvider",
			provName:         "anthropic",
			baseURL:          "https://api.anthropic.com/v1",
			apiKey:           "anthropic-test-key",
			wantProviderType: "anthropic",
			wantBaseURL:      "https://api.anthropic.com/v1",
			wantAPIKey:       "anthropic-test-key",
		},
		{
			name:             "mistral selects MistralProvider",
			provName:         "mistral",
			baseURL:          "https://api.mistral.ai/v1",
			apiKey:           "mistral-test-key",
			wantProviderType: "mistral",
			wantBaseURL:      "https://api.mistral.ai/v1",
			wantAPIKey:       "mistral-test-key",
		},
		{
			name:      "unknown provider returns error",
			provName:  "unknown",
			expectErr: true,
		},
		{
			name:      "empty provider name returns error",
			provName:  "",
			expectErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			p, err := NewProvider(testCase.provName, testCase.baseURL, testCase.apiKey)

			if testCase.expectErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				if p != nil {
					t.Errorf("expected nil Provider on error, got %T", p)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p == nil {
				t.Fatal("expected a non-nil Provider")
			}

			switch testCase.wantProviderType {
			case "ollama":
				got, ok := p.(*OllamaProvider)
				if !ok {
					t.Fatalf("expected *OllamaProvider, got %T", p)
				}
				if got.BaseURL != testCase.wantBaseURL {
					t.Errorf("BaseURL: want %q, got %q", testCase.wantBaseURL, got.BaseURL)
				}

			case "openai":
				got, ok := p.(*OpenAIProvider)
				if !ok {
					t.Fatalf("expected *OpenAIProvider, got %T", p)
				}
				if got.BaseURL != testCase.wantBaseURL {
					t.Errorf("BaseURL: want %q, got %q", testCase.wantBaseURL, got.BaseURL)
				}
				if got.APIKey != testCase.wantAPIKey {
					t.Errorf("APIKey: want %q, got %q", testCase.wantAPIKey, got.APIKey)
				}

			case "gemini":
				got, ok := p.(*GeminiProvider)
				if !ok {
					t.Fatalf("expected *GeminiProvider, got %T", p)
				}
				if got.BaseURL != testCase.wantBaseURL {
					t.Errorf("BaseURL: want %q, got %q", testCase.wantBaseURL, got.BaseURL)
				}
				if got.APIKey != testCase.wantAPIKey {
					t.Errorf("APIKey: want %q, got %q", testCase.wantAPIKey, got.APIKey)
				}

			case "anthropic":
				got, ok := p.(*AnthropicProvider)
				if !ok {
					t.Fatalf("expected *AnthropicProvider, got %T", p)
				}
				if got.BaseURL != testCase.wantBaseURL {
					t.Errorf("BaseURL: want %q, got %q", testCase.wantBaseURL, got.BaseURL)
				}
				if got.APIKey != testCase.wantAPIKey {
					t.Errorf("APIKey: want %q, got %q", testCase.wantAPIKey, got.APIKey)
				}

			case "mistral":
				got, ok := p.(*MistralProvider)
				if !ok {
					t.Fatalf("expected *MistralProvider, got %T", p)
				}
				if got.BaseURL != testCase.wantBaseURL {
					t.Errorf("BaseURL: want %q, got %q", testCase.wantBaseURL, got.BaseURL)
				}
				if got.APIKey != testCase.wantAPIKey {
					t.Errorf("APIKey: want %q, got %q", testCase.wantAPIKey, got.APIKey)
				}
			}
		})
	}
}

// TestNewProvider_LocalModelManager verifies LocalModelManager implementations.
func TestNewProvider_LocalModelManager(t *testing.T) {
	testCases := []struct {
		name         string
		provName     string
		wantsManager bool
	}{
		{"ollama implements LocalModelManager", "ollama", true},
		{"openai does not implement LocalModelManager", "openai", false},
		{"gemini does not implement LocalModelManager", "gemini", false},
		{"anthropic does not implement LocalModelManager", "anthropic", false},
		{"mistral does not implement LocalModelManager", "mistral", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			p, err := NewProvider(testCase.provName, "", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			_, ok := p.(LocalModelManager)
			if ok != testCase.wantsManager {
				t.Errorf("LocalModelManager implementation: want %v, got %v", testCase.wantsManager, ok)
			}
		})
	}
}

// TestRequiresAPIKey verifies API key requirements for each provider.
func TestRequiresAPIKey(t *testing.T) {
	testCases := []struct {
		name     string
		provName string
		expected bool
	}{
		{"openai requires an API key", "openai", true},
		{"gemini requires an API key", "gemini", true},
		{"anthropic requires an API key", "anthropic", true},
		{"mistral requires an API key", "mistral", true},
		{"ollama does not require an API key", "ollama", false},
		{"unknown provider does not require an API key", "unknown", false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := RequiresAPIKey(testCase.provName)
			if got != testCase.expected {
				t.Errorf("RequiresAPIKey(%q) = %v, want %v", testCase.provName, got, testCase.expected)
			}
		})
	}
}
