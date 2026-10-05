package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type MistralMessage struct {
	Role      string            `json:"role"`
	Content   string            `json:"content"`
	ToolCalls []MistralToolCall `json:"tool_calls,omitempty"`
}

type MistralChatRequest struct {
	Model    string                  `json:"model"`
	Messages []MistralMessage        `json:"messages"`
	Tools    []MistralToolDefinition `json:"tools,omitempty"`
}

type MistralToolDefinition struct {
	Type     string                    `json:"type"`
	Function MistralFunctionDefinition `json:"function"`
}

type MistralFunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type MistralToolCall struct {
	ID       string          `json:"id,omitempty"`
	Type     string          `json:"type,omitempty"`
	Function MistralFunction `json:"function"`
}

type MistralFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type MistralResponse struct {
	Choices []struct {
		Message MistralMessage `json:"message"`
	} `json:"choices"`
}

type MistralProvider struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

// convertToMistralMessages converts shared messages to Mistral messages.
func convertToMistralMessages(
	messages []Message,
) []MistralMessage {
	result := make(
		[]MistralMessage,
		0,
		len(messages),
	)

	for _, message := range messages {
		toolCalls := make(
			[]MistralToolCall,
			0,
			len(message.ToolCalls),
		)

		for _, toolCall := range message.ToolCalls {
			arguments, err := json.Marshal(
				toolCall.Arguments,
			)

			if err != nil {
				continue
			}

			toolCalls = append(
				toolCalls,
				MistralToolCall{
					Type: "function",
					Function: MistralFunction{
						Name:      toolCall.Name,
						Arguments: string(arguments),
					},
				},
			)
		}

		result = append(
			result,
			MistralMessage{
				Role:      message.Role,
				Content:   message.Content,
				ToolCalls: toolCalls,
			},
		)
	}

	return result
}

// convertToMistralRequest converts a shared chat request to Mistral format.
func convertToMistralRequest(
	request ChatRequest,
) MistralChatRequest {
	tools := make(
		[]MistralToolDefinition,
		0,
		len(request.Tools),
	)

	for _, tool := range request.Tools {
		tools = append(
			tools,
			MistralToolDefinition{
				Type: "function",
				Function: MistralFunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			},
		)
	}

	return MistralChatRequest{
		Model:    request.Model,
		Messages: convertToMistralMessages(request.Messages),
		Tools:    tools,
	}
}

// convertFromMistralMessage converts a Mistral message to the shared format.
func convertFromMistralMessage(
	message MistralMessage,
) Message {
	toolCalls := make(
		[]ToolCall,
		0,
		len(message.ToolCalls),
	)

	for _, toolCall := range message.ToolCalls {
		var arguments map[string]any

		if toolCall.Function.Arguments != "" {
			if err := json.Unmarshal(
				[]byte(toolCall.Function.Arguments),
				&arguments,
			); err != nil {
				arguments = map[string]any{}
			}
		}

		toolCalls = append(
			toolCalls,
			ToolCall{
				Name:      toolCall.Function.Name,
				Arguments: arguments,
			},
		)
	}

	return Message{
		Role:      message.Role,
		Content:   message.Content,
		ToolCalls: toolCalls,
	}
}

// Chat sends a chat request to the Mistral API.
func (m MistralProvider) Chat(
	request ChatRequest,
) (ChatResponse, error) {
	baseURL := m.BaseURL

	if baseURL == "" {
		baseURL = "https://api.mistral.ai/v1"
	}

	baseURL = strings.TrimRight(baseURL, "/")

	client := m.Client

	if client == nil {
		client = http.DefaultClient
	}

	if m.APIKey == "" {
		return ChatResponse{}, fmt.Errorf(
			"Mistral API key is required",
		)
	}

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

	mistralRequest := convertToMistralRequest(request)

	data, err := json.Marshal(mistralRequest)
	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to encode Mistral request: %w",
			err,
		)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/chat/completions",
		bytes.NewBuffer(data),
	)
	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to create Mistral request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+m.APIKey,
	)

	resp, err := client.Do(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to contact Mistral: %w",
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)

		return ChatResponse{}, fmt.Errorf(
			"Mistral returned status code %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var mistralResp MistralResponse

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&mistralResp); err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to decode Mistral response: %w",
			err,
		)
	}

	if len(mistralResp.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf(
			"Mistral response contains no choices",
		)
	}

	message := convertFromMistralMessage(
		mistralResp.Choices[0].Message,
	)

	return ChatResponse{
		Content:   message.Content,
		ToolCalls: message.ToolCalls,
	}, nil
}
