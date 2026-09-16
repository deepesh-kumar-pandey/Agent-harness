package provider

import (
	"os"
	"testing"
)

func TestOllamaProvider_Integration(t *testing.T) {
	if os.Getenv("OLLAMA_INTEGRATION") != "1" {
		t.Skip("set OLLAMA_INTEGRATION=1 to run the Ollama integration test")
	}

	provider := OllamaProvider{
		BaseURL: "http://localhost:11434",
	}

	request := ChatRequest{
		Model: "kirito1/qwen3-coder:4b",
		Messages: []Message{
			{
				Role:    "user",
				Content: "Explain Go in one sentence.",
			},
		},
	}

	response, err := provider.Chat(request)

	if err != nil {
		t.Fatalf(" Ollama integration test failed: %v", err)
	}

	if response.Content == "" {
		t.Fatalf(" Ollama returned an empty response")
	}

	t.Logf("✅ Ollama response: %s", response.Content)
}
