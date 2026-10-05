package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const anthropicAPIVersion = "2023-06-01"

type AnthropicContentBlock struct {
	Type string `json:"type"`

	Text string `json:"text,omitempty"`

	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`

	Input map[string]any `json:"input,omitempty"`
}

type AnthropicMessage struct {
	Role    string                  `json:"role"`
	Content []AnthropicContentBlock `json:"content"`
}

type AnthropicTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type AnthropicChatRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []AnthropicMessage `json:"messages"`
	Tools     []AnthropicTool    `json:"tools,omitempty"`
}

type AnthropicResponse struct {
	Content []AnthropicContentBlock `json:"content"`
}

type AnthropicProvider struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func convertToAnthropicMessages(
	messages []Message,
) []AnthropicMessage {
	result := make(
		[]AnthropicMessage,
		0,
		len(messages),
	)

	for _, message := range messages {
		role := message.Role

		if role == "assistant" {
			content := make(
				[]AnthropicContentBlock,
				0,
				1+len(message.ToolCalls),
			)

			if message.Content != "" {
				content = append(
					content,
					AnthropicContentBlock{
						Type: "text",
						Text: message.Content,
					},
				)
			}

			for _, toolCall := range message.ToolCalls {
				content = append(
					content,
					AnthropicContentBlock{
						Type:  "tool_use",
						Name:  toolCall.Name,
						Input: toolCall.Arguments,
					},
				)
			}

			result = append(
				result,
				AnthropicMessage{
					Role:    "assistant",
					Content: content,
				},
			)

			continue
		}

		content := make(
			[]AnthropicContentBlock,
			0,
			1,
		)

		if message.Content != "" {
			content = append(
				content,
				AnthropicContentBlock{
					Type: "text",
					Text: message.Content,
				},
			)
		}

		result = append(
			result,
			AnthropicMessage{
				Role:    "user",
				Content: content,
			},
		)
	}

	return result
}

func convertToAnthropicTools(
	tools []ToolDefinition,
) []AnthropicTool {
	if len(tools) == 0 {
		return nil
	}

	result := make(
		[]AnthropicTool,
		0,
		len(tools),
	)

	for _, tool := range tools {
		result = append(
			result,
			AnthropicTool{
				Name:        tool.Name,
				Description: tool.Description,
				InputSchema: tool.Parameters,
			},
		)
	}

	return result
}

func convertToAnthropicRequest(
	request ChatRequest,
) AnthropicChatRequest {
	return AnthropicChatRequest{
		Model:     request.Model,
		MaxTokens: 1024,
		Messages:  convertToAnthropicMessages(request.Messages),
		Tools:     convertToAnthropicTools(request.Tools),
	}
}

func convertFromAnthropicResponse(
	response AnthropicResponse,
) ChatResponse {
	var content strings.Builder

	toolCalls := make(
		[]ToolCall,
		0,
	)

	for _, block := range response.Content {
		switch block.Type {
		case "text":
			content.WriteString(block.Text)

		case "tool_use":
			toolCalls = append(
				toolCalls,
				ToolCall{
					Name:      block.Name,
					Arguments: block.Input,
				},
			)
		}
	}

	return ChatResponse{
		Content:   content.String(),
		ToolCalls: toolCalls,
	}
}

func (a AnthropicProvider) Chat(
	request ChatRequest,
) (ChatResponse, error) {
	baseURL := a.BaseURL

	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	baseURL = strings.TrimRight(
		baseURL,
		"/",
	)

	client := a.Client

	if client == nil {
		client = http.DefaultClient
	}

	if a.APIKey == "" {
		return ChatResponse{}, fmt.Errorf(
			"Anthropic API key is required",
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

	anthropicRequest := convertToAnthropicRequest(
		request,
	)

	data, err := json.Marshal(
		anthropicRequest,
	)

	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to encode Anthropic request: %w",
			err,
		)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/messages",
		bytes.NewBuffer(data),
	)

	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to create Anthropic request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"x-api-key",
		a.APIKey,
	)

	req.Header.Set(
		"anthropic-version",
		anthropicAPIVersion,
	)

	resp, err := client.Do(req)

	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to contact Anthropic: %w",
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(
			resp.Body,
		)

		return ChatResponse{}, fmt.Errorf(
			"Anthropic returned status code %d: %s",
			resp.StatusCode,
			strings.TrimSpace(
				string(body),
			),
		)
	}

	var anthropicResp AnthropicResponse

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&anthropicResp); err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to decode Anthropic response: %w",
			err,
		)
	}

	return convertFromAnthropicResponse(
		anthropicResp,
	), nil
}
