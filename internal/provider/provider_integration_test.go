package tools

import "testing"

func TestOllamaProvider_Integration(t *testing.T) {

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
		t.Fatalf("❌ Ollama integration test failed: %v", err)
	}

	if response.Content == "" {
		t.Fatalf("❌ Ollama returned an empty response")
	}

	t.Logf("✅ Ollama response: %s", response.Content)
}
