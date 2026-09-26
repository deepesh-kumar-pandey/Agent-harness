package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type OpenAIMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []OpenAIToolCall `json:"tool_calls,omitempty"`
}

type OpenAIChatRequest struct {
	Model    string                 `json:"model"`
	Messages []OpenAIMessage        `json:"messages"`
	Tools    []OpenAIToolDefinition `json:"tools,omitempty"`
}

type OpenAIToolDefinition struct {
	Type     string                   `json:"type"`
	Function OpenAIFunctionDefinition `json:"function"`
}

type OpenAIFunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type OpenAIToolCall struct {
	Function OpenAIFunction `json:"function"`
}

type OpenAIFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message OpenAIMessage `json:"message"`
	} `json:"choices"`
}

type OpenAIProvider struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func convertToOpenAIMessages(
	messages []Message,
) []OpenAIMessage {
	result := make(
		[]OpenAIMessage,
		0,
		len(messages),
	)

	for _, message := range messages {
		toolCalls := make(
			[]OpenAIToolCall,
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
				OpenAIToolCall{
					Function: OpenAIFunction{
						Name:      toolCall.Name,
						Arguments: string(arguments),
					},
				},
			)
		}

		result = append(
			result,
			OpenAIMessage{
				Role:      message.Role,
				Content:   message.Content,
				ToolCalls: toolCalls,
			},
		)
	}

	return result
}

func convertToOpenAIRequest(
	request ChatRequest,
) OpenAIChatRequest {
	tools := make(
		[]OpenAIToolDefinition,
		0,
		len(request.Tools),
	)

	for _, tool := range request.Tools {
		tools = append(
			tools,
			OpenAIToolDefinition{
				Type: "function",
				Function: OpenAIFunctionDefinition{
					Name:        tool.Name,
					Description: tool.Description,
					Parameters:  tool.Parameters,
				},
			},
		)
	}

	return OpenAIChatRequest{
		Model:    request.Model,
		Messages: convertToOpenAIMessages(request.Messages),
		Tools:    tools,
	}
}

func convertFromOpenAIMessage(
	message OpenAIMessage,
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

func (o OpenAIProvider) Chat(
	request ChatRequest,
) (ChatResponse, error) {
	baseURL := o.BaseURL

	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	baseURL = strings.TrimRight(baseURL, "/")

	client := o.Client

	if client == nil {
		client = http.DefaultClient
	}

	if o.APIKey == "" {
		return ChatResponse{}, fmt.Errorf(
			"OpenAI API key is required",
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

	openAIRequest := convertToOpenAIRequest(request)

	data, err := json.Marshal(openAIRequest)
	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to encode OpenAI request: %w",
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
			"failed to create OpenAI request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"Authorization",
		"Bearer "+o.APIKey,
	)

	resp, err := client.Do(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to contact OpenAI: %w",
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)

		return ChatResponse{}, fmt.Errorf(
			"OpenAI returned status code %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var openAIResp OpenAIResponse

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&openAIResp); err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to decode OpenAI response: %w",
			err,
		)
	}

	if len(openAIResp.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf(
			"OpenAI response contains no choices",
		)
	}

	message := convertFromOpenAIMessage(
		openAIResp.Choices[0].Message,
	)

	return ChatResponse{
		Content:   message.Content,
		ToolCalls: message.ToolCalls,
	}, nil
}
