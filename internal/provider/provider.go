package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type Message struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type ChatRequest struct {
	Model    string           `json:"model"`
	Messages []Message        `json:"messages"`
	Tools    []ToolDefinition `json:"tools,omitempty"`
}

type OllamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OllamaToolCall `json:"tool_calls,omitempty"`
}

type OllamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []OllamaMessage        `json:"messages"`
	Tools    []OllamaToolDefinition `json:"tools,omitempty"`
}

type OllamaToolDefinition struct {
	Type     string                   `json:"type"`
	Function OllamaFunctionDefinition `json:"function"`
}

type OllamaFunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type OllamaToolCall struct {
	Function OllamaFunction `json:"function"`
}

type OllamaFunction struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ToolCall struct {
	Name      string
	Arguments map[string]any
}

type ChatResponse struct {
	Content   string
	ToolCalls []ToolCall
}

type OllamaResponse struct {
	Message OllamaMessage `json:"message"`
}

type Provider interface {
	Chat(request ChatRequest) (ChatResponse, error)
}

type OllamaProvider struct {
	BaseURL string
	Client  *http.Client
}

func convertToOllamaMessages(
	messages []Message,
) []OllamaMessage {

	result := make(
		[]OllamaMessage,
		0,
		len(messages),
	)

	for _, message := range messages {

		toolCalls := make(
			[]OllamaToolCall,
			0,
			len(message.ToolCalls),
		)

		for _, toolCall := range message.ToolCalls {
			toolCalls = append(
				toolCalls,
				OllamaToolCall{
					Function: OllamaFunction{
						Name:      toolCall.Name,
						Arguments: toolCall.Arguments,
					},
				},
			)
		}

		result = append(
			result,
			OllamaMessage{
				Role:      message.Role,
				Content:   message.Content,
				ToolCalls: toolCalls,
			},
		)
	}

	return result
}

func convertFromOllamaMessage(
	message OllamaMessage,
) Message {

	toolCalls := make(
		[]ToolCall,
		0,
		len(message.ToolCalls),
	)

	for _, toolCall := range message.ToolCalls {
		toolCalls = append(
			toolCalls,
			ToolCall{
				Name:      toolCall.Function.Name,
				Arguments: toolCall.Function.Arguments,
			},
		)
	}

	return Message{
		Role:      message.Role,
		Content:   message.Content,
		ToolCalls: toolCalls,
	}
}

func convertToOllamaRequest(
	request ChatRequest,
) OllamaChatRequest {

	tools := make(
		[]OllamaToolDefinition,
		0,
		len(request.Tools),
	)

	for _, tool := range request.Tools {
		tools = append(
			tools,
			OllamaToolDefinition{
				Type: "function",
				Function: OllamaFunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			},
		)
	}

	return OllamaChatRequest{
		Model:    request.Model,
		Messages: convertToOllamaMessages(request.Messages),
		Tools:    tools,
	}
}

func (o OllamaProvider) Chat(
	request ChatRequest,
) (ChatResponse, error) {

	baseURL := o.BaseURL

	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	client := o.Client

	if client == nil {
		client = http.DefaultClient
	}

	var ollamaResp OllamaResponse

	if request.Model == "" {
		return ChatResponse{}, fmt.Errorf(
			"Model is required",
		)
	}

	if len(request.Messages) == 0 {
		return ChatResponse{}, fmt.Errorf(
			"At least one message is required",
		)
	}

	ollamaRequest := convertToOllamaRequest(request)

	data, err := json.Marshal(ollamaRequest)

	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to encode request: %w",
			err,
		)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/api/chat",
		bytes.NewBuffer(data),
	)

	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to create request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	resp, err := client.Do(req)

	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to contact ollama: %w",
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ChatResponse{}, fmt.Errorf(
			"Ollama returned status code %d",
			resp.StatusCode,
		)
	}

	err = json.NewDecoder(
		resp.Body,
	).Decode(&ollamaResp)

	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to decode ollama response: %w",
			err,
		)
	}

	message := convertFromOllamaMessage(
		ollamaResp.Message,
	)

	return ChatResponse{
		Content:   message.Content,
		ToolCalls: message.ToolCalls,
	}, nil
}
