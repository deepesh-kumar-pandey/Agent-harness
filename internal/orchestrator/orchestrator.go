package orchestrator

import (
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

	fmt.Println("💬 Orchestrator sending request to Provider...")

	return o.providerClient.Chat(request)
}
