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

	fmt.Println("🚀 Creating Orchestrator...")

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
	for toolCallCount := 0; toolCallCount <= o.maxToolCalls; toolCallCount++ {
		response, err := o.providerClient.Chat(request)
		if err != nil {
			return AgentResponse{}, err
		}

		var agentResponse AgentResponse
		trimmedContent := strings.TrimSpace(response.Content)
		if strings.HasPrefix(trimmedContent, "{") && json.Valid([]byte(trimmedContent)) {
			err = json.Unmarshal([]byte(response.Content), &agentResponse)
			if err != nil {
				return AgentResponse{}, fmt.Errorf(
					"failed to parse agent response: %w",
					err,
				)
			}
		} else {
			agentResponse.Content = response.Content
		}

		if agentResponse.ToolCall == nil {
			return agentResponse, nil
		}

		if toolCallCount == o.maxToolCalls {
			return AgentResponse{}, fmt.Errorf(
				"maximum tool-call limit (%d) exceeded",
				o.maxToolCalls,
			)
		}

		result, err := o.AssignTool(*agentResponse.ToolCall)
		if err != nil {
			return AgentResponse{}, fmt.Errorf(
				"failed to execute tool: %w",
				err,
			)
		}

		request.Messages = append(request.Messages,
			providerpkg.Message{Role: "assistant", Content: response.Content},
			providerpkg.Message{
				Role:    "tool",
				Content: fmt.Sprintf("%v", result),
			},
		)
	}

	return AgentResponse{}, fmt.Errorf("agent loop terminated unexpectedly")
}
