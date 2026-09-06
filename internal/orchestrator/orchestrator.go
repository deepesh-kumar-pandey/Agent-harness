package orchestrator

import (
	"encoding/json"
	"fmt"

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
}

func NewOrchestrator(
	agentClient *agentpkg.Agent,
	providerClient providerpkg.Provider,
) *DefaultOrchestrator {

	fmt.Println("🚀 Creating Orchestrator...")

	return &DefaultOrchestrator{
		agentClient:    agentClient,
		providerClient: providerClient,
	}
}

func (o *DefaultOrchestrator) Run(
	name string,
	args map[string]any,
) (any, error) {

	fmt.Printf("🔧 Orchestrator executing tool: %s\n", name)

	return o.agentClient.ExecuteTool(name, args)
}

func (o *DefaultOrchestrator) Chat(
	request providerpkg.ChatRequest,
) (providerpkg.ChatResponse, error) {

	fmt.Println(" Orchestrator sending request to Provider...")

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

func (o *DefaultOrchestrator) RunAgent(request providerpkg.ChatRequest) (AgentResponse, error) {

	// Send the request to the LLM provider.
	response, err := o.providerClient.Chat(request)
	if err != nil {
		return AgentResponse{}, err
	}

	// Decode structured tool-call responses when the provider returns JSON.
	var agentResponse AgentResponse

	if json.Valid([]byte(response.Content)) {
		err = json.Unmarshal([]byte(response.Content), &agentResponse)
		if err != nil {
			return AgentResponse{}, fmt.Errorf(
				"failed to parse agent response: %w",
				err,
			)
		}
	} else {
		// Plain text is a valid provider response with no tool call.
		agentResponse.Content = response.Content
	}

	// Execute the requested tool.
	if agentResponse.ToolCall != nil {

		result, err := o.AssignTool(
			*agentResponse.ToolCall,
		)

		if err != nil {
			return AgentResponse{}, fmt.Errorf(
				"failed to execute tool: %w",
				err,
			)
		}

		return AgentResponse{
			Content: fmt.Sprintf("%v", result),
		}, nil
	}

	// Return the structured agent response.
	return agentResponse, nil
}
