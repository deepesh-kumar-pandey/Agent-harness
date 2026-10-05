package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GeminiPart struct {
	Text         string              `json:"text,omitempty"`
	FunctionCall *GeminiFunctionCall `json:"functionCall,omitempty"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
}

type GeminiFunctionDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type GeminiTool struct {
	FunctionDeclarations []GeminiFunctionDeclaration `json:"functionDeclarations"`
}

type GeminiChatRequest struct {
	Contents []GeminiContent `json:"contents"`
	Tools    []GeminiTool    `json:"tools,omitempty"`
}

type GeminiCandidate struct {
	Content GeminiContent `json:"content"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"candidates"`
}

type GeminiProvider struct {
	APIKey  string
	BaseURL string
	Client  *http.Client
}

func convertToGeminiContents(
	messages []Message,
) []GeminiContent {
	result := make(
		[]GeminiContent,
		0,
		len(messages),
	)

	for _, message := range messages {
		role := message.Role

		if role == "assistant" {
			role = "model"
		}

		parts := make(
			[]GeminiPart,
			0,
			1+len(message.ToolCalls),
		)

		if message.Content != "" {
			parts = append(
				parts,
				GeminiPart{
					Text: message.Content,
				},
			)
		}

		for _, toolCall := range message.ToolCalls {
			parts = append(
				parts,
				GeminiPart{
					FunctionCall: &GeminiFunctionCall{
						Name: toolCall.Name,
						Args: toolCall.Arguments,
					},
				},
			)
		}

		result = append(
			result,
			GeminiContent{
				Role:  role,
				Parts: parts,
			},
		)
	}

	return result
}

func convertToGeminiTools(
	tools []ToolDefinition,
) []GeminiTool {
	if len(tools) == 0 {
		return nil
	}

	declarations := make(
		[]GeminiFunctionDeclaration,
		0,
		len(tools),
	)

	for _, tool := range tools {
		declarations = append(
			declarations,
			GeminiFunctionDeclaration{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters:  tool.Parameters,
			},
		)
	}

	return []GeminiTool{
		{
			FunctionDeclarations: declarations,
		},
	}
}

func convertToGeminiRequest(
	request ChatRequest,
) GeminiChatRequest {
	return GeminiChatRequest{
		Contents: convertToGeminiContents(request.Messages),
		Tools:    convertToGeminiTools(request.Tools),
	}
}

func convertFromGeminiResponse(
	response GeminiResponse,
) ChatResponse {
	if len(response.Candidates) == 0 {
		return ChatResponse{}
	}

	message := response.Candidates[0].Content

	toolCalls := make(
		[]ToolCall,
		0,
	)

	var content strings.Builder

	for _, part := range message.Parts {
		if part.Text != "" {
			content.WriteString(part.Text)
		}

		if part.FunctionCall != nil {
			toolCalls = append(
				toolCalls,
				ToolCall{
					Name:      part.FunctionCall.Name,
					Arguments: part.FunctionCall.Args,
				},
			)
		}
	}

	return ChatResponse{
		Content:   content.String(),
		ToolCalls: toolCalls,
	}
}

func (g GeminiProvider) Chat(
	request ChatRequest,
) (ChatResponse, error) {
	baseURL := g.BaseURL

	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}

	baseURL = strings.TrimRight(baseURL, "/")

	client := g.Client

	if client == nil {
		client = http.DefaultClient
	}

	if g.APIKey == "" {
		return ChatResponse{}, fmt.Errorf(
			"Gemini API key is required",
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

	geminiRequest := convertToGeminiRequest(request)

	data, err := json.Marshal(geminiRequest)
	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to encode Gemini request: %w",
			err,
		)
	}

	url := fmt.Sprintf(
		"%s/models/%s:generateContent",
		baseURL,
		request.Model,
	)

	req, err := http.NewRequest(
		http.MethodPost,
		url,
		bytes.NewBuffer(data),
	)
	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to create Gemini request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"x-goog-api-key",
		g.APIKey,
	)

	resp, err := client.Do(req)
	if err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to contact Gemini: %w",
			err,
		)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)

		return ChatResponse{}, fmt.Errorf(
			"Gemini returned status code %d: %s",
			resp.StatusCode,
			strings.TrimSpace(string(body)),
		)
	}

	var geminiResp GeminiResponse

	if err := json.NewDecoder(
		resp.Body,
	).Decode(&geminiResp); err != nil {
		return ChatResponse{}, fmt.Errorf(
			"failed to decode Gemini response: %w",
			err,
		)
	}

	if len(geminiResp.Candidates) == 0 {
		return ChatResponse{}, fmt.Errorf(
			"Gemini response contains no candidates",
		)
	}

	return convertFromGeminiResponse(geminiResp), nil
}
