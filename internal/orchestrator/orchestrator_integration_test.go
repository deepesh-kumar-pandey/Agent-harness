package orchestrator

import (
	"os"
	"testing"

	agentpkg "agent-harness/internal/agent"
	providerpkg "agent-harness/internal/provider"
	toolspkg "agent-harness/internal/tools"
)

func TestOrchestratorRunAgent_Integration(t *testing.T) {
	if os.Getenv("ORCHESTRATOR_INTEGRATION") != "1" {
		t.Skip("set ORCHESTRATOR_INTEGRATION=1 to run the Ollama integration test")
	}

	registry := toolspkg.NewToolRegistry()
	testAgent := agentpkg.NewAgent(registry)

	ollamaProvider := &providerpkg.OllamaProvider{
		BaseURL: "http://localhost:11434",
	}

	testOrchestrator := NewOrchestrator(
		testAgent,
		ollamaProvider,
	)

	request := providerpkg.ChatRequest{
		Model: "kirito1/qwen3-coder:4b",
		Messages: []providerpkg.Message{
			{
				Role:    "system",
				Content: "You are an agent that can use the available tools. Use the calculator tool when needed.",
			},
			{
				Role:    "user",
				Content: "Calculate 10 + 20 using the calculator tool.",
			},
		},
	}

	response, err := testOrchestrator.RunAgent(request)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if response.Content == "" {
		t.Fatalf("Expected a final response, got empty content")
	}

	if response.ToolCall != nil {
		t.Fatalf("Expected final response without a tool call")
	}

	t.Logf("Final agent response: %s", response.Content)
}
