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

// ─────────────────────────────────────────────
// Direct Tool Execution
// ─────────────────────────────────────────────

func (o *DefaultOrchestrator) Run(
	name string,
	args map[string]any,
) (any, error) {

	fmt.Printf("Orchestrator executing tool: %s\n", name)

	return o.agentClient.ExecuteTool(name, args)
}

// ─────────────────────────────────────────────
// Provider Communication
// ─────────────────────────────────────────────

func (o *DefaultOrchestrator) Chat(
	request providerpkg.ChatRequest,
) (providerpkg.ChatResponse, error) {

	fmt.Println("Orchestrator sending request to Provider...")

	return o.providerClient.Chat(request)
}

// ─────────────────────────────────────────────
// Tool Assignment
// ─────────────────────────────────────────────

func (o *DefaultOrchestrator) AssignTool(
	toolCall ToolCall,
) (any, error) {

	result, err := o.agentClient.ExecuteTool(
		toolCall.Tool,
		toolCall.Args,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to execute tool %q: %w",
			toolCall.Tool,
			err,
		)
	}

	return result, nil
}

// ─────────────────────────────────────────────
// Agent Execution
// ─────────────────────────────────────────────

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

	if len(request.Messages) == 0 {
		return AgentResponse{}, fmt.Errorf(
			"request contains no messages",
		)
	}

	// Add the new user message to the persistent conversation.
	userMessage := request.Messages[len(request.Messages)-1]

	if err := o.agentClient.AddMessage(userMessage); err != nil {
		return AgentResponse{}, fmt.Errorf(
			"failed to add user message to conversation: %w",
			err,
		)
	}

	// Use the persistent conversation as the provider request history.
	request.Messages = o.agentClient.GetMessages()

	toolCallCount := 0

	for {

		// Send the complete conversation to the provider.
		response, err := o.Chat(request)
		if err != nil {
			return AgentResponse{}, err
		}

		// ─────────────────────────────────────────────
		// Native Provider Tool Calls
		// ─────────────────────────────────────────────

		if len(response.ToolCalls) > 0 {

			// Add the assistant's response and tool calls
			// to the persistent conversation.
			assistantMessage := providerpkg.Message{
				Role:      "assistant",
				Content:   response.Content,
				ToolCalls: response.ToolCalls,
			}

			if err := o.agentClient.AddMessage(assistantMessage); err != nil {
				return AgentResponse{}, fmt.Errorf(
					"failed to add assistant message to conversation: %w",
					err,
				)
			}

			// Execute every requested tool call.
			for _, providerToolCall := range response.ToolCalls {

				// Check whether the maximum number of tool calls
				// has been reached before executing the tool.
				if toolCallCount >= o.maxToolCalls {
					return AgentResponse{}, fmt.Errorf(
						"maximum tool-call limit (%d) exceeded",
						o.maxToolCalls,
					)
				}

				toolCall := ToolCall{
					Tool: providerToolCall.Name,
					Args: providerToolCall.Arguments,
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

				// Add the tool result to the persistent conversation.
				toolMessage := providerpkg.Message{
					Role:    "tool",
					Content: fmt.Sprintf("%v", result),
				}

				if err := o.agentClient.AddMessage(toolMessage); err != nil {
					return AgentResponse{}, fmt.Errorf(
						"failed to add tool message to conversation: %w",
						err,
					)
				}
			}

			// Refresh the provider request with the updated
			// persistent conversation.
			request.Messages = o.agentClient.GetMessages()

			continue
		}

		// ─────────────────────────────────────────────
		// Decode Provider Response
		// ─────────────────────────────────────────────

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

		// ─────────────────────────────────────────────
		// Final Response
		// ─────────────────────────────────────────────

		// No tool call means the LLM has produced
		// the final answer.
		if agentResponse.ToolCall == nil {

			assistantMessage := providerpkg.Message{
				Role:    "assistant",
				Content: response.Content,
			}

			if err := o.agentClient.AddMessage(assistantMessage); err != nil {
				return AgentResponse{}, fmt.Errorf(
					"failed to add assistant message to conversation: %w",
					err,
				)
			}

			return agentResponse, nil
		}

		// ─────────────────────────────────────────────
		// Legacy Tool Call
		// ─────────────────────────────────────────────

		// Check whether the maximum number of tool calls
		// has been reached.
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
		// to the persistent conversation.
		assistantMessage := providerpkg.Message{
			Role:    "assistant",
			Content: response.Content,
		}

		if err := o.agentClient.AddMessage(assistantMessage); err != nil {
			return AgentResponse{}, fmt.Errorf(
				"failed to add assistant message to conversation: %w",
				err,
			)
		}

		// Add the tool result to the persistent conversation.
		toolMessage := providerpkg.Message{
			Role:    "tool",
			Content: fmt.Sprintf("%v", result),
		}

		if err := o.agentClient.AddMessage(toolMessage); err != nil {
			return AgentResponse{}, fmt.Errorf(
				"failed to add tool message to conversation: %w",
				err,
			)
		}

		// Refresh the provider request with the updated
		// persistent conversation.
		request.Messages = o.agentClient.GetMessages()
	}
}

// ─────────────────────────────────────────────
// Tool Schemas
// ─────────────────────────────────────────────

func (o *DefaultOrchestrator) GetToolSchemas() ([]map[string]any, error) {
	return o.agentClient.GetToolSchemas()
}

// ─────────────────────────────────────────────
// Provider Tool Definitions
// ─────────────────────────────────────────────

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
