package orchestrator

import (
	"encoding/json"
	"fmt"
	"strings"

	agentpkg "agent-harness/internal/agent"
	providerpkg "agent-harness/internal/provider"
)

type Orchestrator interface {
	Run(name string, args map[string]any) (any, error)
	Chat(request providerpkg.ChatRequest) (providerpkg.ChatResponse, error)
}

type DefaultOrchestrator struct {
	agentClient    *agentpkg.Agent
	providerClient providerpkg.Provider
	maxToolCalls   int
}

type OrchestratorOption func(*DefaultOrchestrator)

func WithMaxToolCalls(maxToolCalls int) OrchestratorOption {
	return func(orchestrator *DefaultOrchestrator) {
		if maxToolCalls > 0 {
			orchestrator.maxToolCalls = maxToolCalls
		}
	}
}

func NewOrchestrator(
	agentClient *agentpkg.Agent,
	providerClient providerpkg.Provider,
	options ...OrchestratorOption,
) *DefaultOrchestrator {

	fmt.Println("Creating Orchestrator...")

	orchestrator := &DefaultOrchestrator{
		agentClient:    agentClient,
		providerClient: providerClient,
		maxToolCalls:   10,
	}

	for _, option := range options {
		option(orchestrator)
	}

	return orchestrator
}

func (o *DefaultOrchestrator) Run(
	name string,
	args map[string]any,
) (any, error) {

	fmt.Printf("Orchestrator executing tool: %s\n", name)

	return o.agentClient.ExecuteTool(name, args)
}

func (o *DefaultOrchestrator) Chat(
	request providerpkg.ChatRequest,
) (providerpkg.ChatResponse, error) {

	fmt.Println("Orchestrator sending request to Provider...")

	return o.providerClient.Chat(request)
}

func (o *DefaultOrchestrator) AssignTool(
	toolCall ToolCall,
) (any, error) {

	return o.agentClient.ExecuteTool(
		toolCall.Tool,
		toolCall.Args,
	)
}

func convertProviderToolCalls(
	toolCalls []providerpkg.ToolCall,
) []providerpkg.OllamaToolCall {
	result := make([]providerpkg.OllamaToolCall, 0, len(toolCalls))

	for _, toolCall := range toolCalls {
		result = append(
			result,
			providerpkg.OllamaToolCall{
				Function: providerpkg.OllamaFunction{
					Name:      toolCall.Name,
					Arguments: toolCall.Arguments,
				},
			},
		)
	}

	return result
}

func (o *DefaultOrchestrator) RunAgent(
	request providerpkg.ChatRequest,
) (AgentResponse, error) {

	toolDefinitions, err := o.GetToolDefinitions()
	if err != nil {
		return AgentResponse{}, fmt.Errorf(
			"failed to get tool definitions: %w",
			err,
		)
	}

	request.Tools = toolDefinitions

	toolCallCount := 0

	for {

		// Send the request to the LLM provider.
		response, err := o.Chat(request)
		if err != nil {
			return AgentResponse{}, err
		}

		// Handle native provider tool calls.
		if len(response.ToolCalls) > 0 {

			toolCall := ToolCall{
				Tool: response.ToolCalls[0].Name,
				Args: response.ToolCalls[0].Arguments,
			}

			// Check whether the maximum number of tool calls has been reached.
			if toolCallCount >= o.maxToolCalls {
				return AgentResponse{}, fmt.Errorf(
					"maximum tool-call limit (%d) exceeded",
					o.maxToolCalls,
				)
			}

			// Execute the requested tool.
			result, err := o.AssignTool(toolCall)

			if err != nil {
				return AgentResponse{}, fmt.Errorf(
					"failed to execute tool: %w",
					err,
				)
			}

			toolCallCount++

			// Add the assistant's tool-call response
			// and the tool result to the conversation.
			request.Messages = append(
				request.Messages,
				providerpkg.Message{
					Role:      "assistant",
					Content:   response.Content,
					ToolCalls: response.ToolCalls,
				},
				providerpkg.Message{
					Role:    "tool",
					Content: fmt.Sprintf("%v", result),
				},
			)

			continue
		}

		// Decode the provider response.
		var agentResponse AgentResponse

		trimmedContent := strings.TrimSpace(response.Content)

		if strings.HasPrefix(trimmedContent, "{") &&
			json.Valid([]byte(trimmedContent)) {

			err = json.Unmarshal(
				[]byte(response.Content),
				&agentResponse,
			)

			if err != nil {
				return AgentResponse{}, fmt.Errorf(
					"failed to parse agent response: %w",
					err,
				)
			}

		} else {

			// Plain text is a valid provider response.
			agentResponse.Content = response.Content
		}

		// No tool call means the LLM has produced the final answer.
		if agentResponse.ToolCall == nil {
			return agentResponse, nil
		}

		// Check whether the maximum number of tool calls has been reached.
		if toolCallCount >= o.maxToolCalls {
			return AgentResponse{}, fmt.Errorf(
				"maximum tool-call limit (%d) exceeded",
				o.maxToolCalls,
			)
		}

		// Execute the requested tool.
		result, err := o.AssignTool(
			*agentResponse.ToolCall,
		)

		if err != nil {
			return AgentResponse{}, fmt.Errorf(
				"failed to execute tool: %w",
				err,
			)
		}

		toolCallCount++

		// Add the assistant's tool-call response
		// and the tool result to the conversation.
		request.Messages = append(
			request.Messages,
			providerpkg.Message{
				Role:    "assistant",
				Content: response.Content,
			},
			providerpkg.Message{
				Role:    "tool",
				Content: fmt.Sprintf("%v", result),
			},
		)
	}
}

func (o *DefaultOrchestrator) GetToolSchemas() ([]map[string]any, error) {
	return o.agentClient.GetToolSchemas()
}

func (o *DefaultOrchestrator) GetToolDefinitions() ([]providerpkg.ToolDefinition, error) {

	schemas, err := o.agentClient.GetToolSchemas()
	if err != nil {
		return nil, err
	}

	definitions := make(
		[]providerpkg.ToolDefinition,
		0,
		len(schemas),
	)

	for _, schema := range schemas {

		name, ok := schema["name"].(string)
		if !ok || name == "" {
			return nil, fmt.Errorf(
				"tool schema has invalid name",
			)
		}

		description, ok := schema["description"].(string)
		if !ok || description == "" {
			return nil, fmt.Errorf(
				"tool schema has invalid description",
			)
		}

		toolParameters, ok := schema["schema"].(map[string]any)
		if !ok || toolParameters == nil {
			return nil, fmt.Errorf(
				"tool schema has invalid parameters",
			)
		}

		definitions = append(
			definitions,
			providerpkg.ToolDefinition{
				Name:        name,
				Description: description,
				Parameters:  toolParameters,
			},
		)
	}

	return definitions, nil
}
