package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Content string
}

type Provider interface {
	Chat(request ChatRequest) (ChatResponse, error)
}

type OllamaProvider struct {
	BaseURL string
	Client  *http.Client
}

type OllamaResponse struct {
	Message Message `json:"message"`
}

func (o OllamaProvider) Chat(request ChatRequest) (ChatResponse, error) {
	baseURL := o.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	client := o.Client
	if client == nil {
		client = http.DefaultClient
	}

	var OllamaResp OllamaResponse
	if request.Model == "" {
		return ChatResponse{}, fmt.Errorf("Model is required")
	}

	if len(request.Messages) == 0 {
		return ChatResponse{}, fmt.Errorf("At least one message is required")
	}

	data, err := json.Marshal(request)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("failed to encode request: %w", err)
	}
	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/api/chat",
		bytes.NewBuffer(data),
	)

	if err != nil {
		return ChatResponse{}, fmt.Errorf("Failed to create request")
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)

	if err != nil {
		return ChatResponse{}, fmt.Errorf("Failed to contact ollama")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ChatResponse{}, fmt.Errorf("Ollama returned status code %d", resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(&OllamaResp)

	if err != nil {
		return ChatResponse{}, fmt.Errorf("Failed to decode ollama response: %w", err)
	}
	return ChatResponse{Content: OllamaResp.Message.Content}, nil
}
