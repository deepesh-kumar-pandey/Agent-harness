package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	agentpkg "agent-harness/internal/agent"
	orchestratorpkg "agent-harness/internal/orchestrator"
	providerpkg "agent-harness/internal/provider"
	toolspkg "agent-harness/internal/tools"
)

func resolveModel() string {
	if model := strings.TrimSpace(os.Getenv("OLLAMA_MODEL")); model != "" {
		return model
	}

	return "kirito1/qwen3-coder:4b"
}

func main() {
	model := resolveModel()

	fmt.Printf("Agent Harness starting... using model: %s\n", model)

	registry := toolspkg.NewToolRegistry()

	agentClient := agentpkg.NewAgent(registry)

	providerClient := providerpkg.OllamaProvider{
		BaseURL: "http://localhost:11434",
	}

	orchestratorClient := orchestratorpkg.NewOrchestrator(
		agentClient,
		providerClient,
	)

	fmt.Println("Agent Harness initialized...")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())

		if input == "exit" {
			break
		}

		if input == "help" {
			fmt.Println("Available commands: help, exit")
			continue
		}

		request := providerpkg.ChatRequest{
			Model: model,
			Messages: []providerpkg.Message{
				{
					Role:    "user",
					Content: input,
				},
			},
		}

		response, err := orchestratorClient.RunAgent(request)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		fmt.Println(response.Content)
	}
}
