package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	configpkg "agent-harness/config"
	agentpkg "agent-harness/internal/agent"
	orchestratorpkg "agent-harness/internal/orchestrator"
	providerpkg "agent-harness/internal/provider"
	toolspkg "agent-harness/internal/tools"
)

func resolveModel(configModel string) (string, string) {
	if model := strings.TrimSpace(os.Getenv("OLLAMA_MODEL")); model != "" {
		return model, "environment"
	}

	if model := strings.TrimSpace(configModel); model != "" {
		return model, "config"
	}

	return "kirito1/qwen3-coder:4b", "default"
}

func loadConfig() (*configpkg.Config, error) {
	configPaths := []string{
		"config/config.json",
		"../config/config.json",
	}

	var lastErr error
	for _, path := range configPaths {
		appConfig, err := configpkg.Load(path)
		if err == nil {
			return appConfig, nil
		}

		lastErr = err
	}

	return nil, lastErr
}

func selectModel(
	providerClient providerpkg.Provider,
	configuredModel string,
	configuredSource string,
	input *bufio.Scanner,
	output io.Writer,
) (string, string, error) {
	manager, ok := providerClient.(providerpkg.LocalModelManager)
	if !ok {
		return configuredModel, configuredSource, nil
	}

	models, err := manager.ListModels()
	if err != nil {
		fmt.Fprintln(output, "Unable to connect to Ollama.")
		fmt.Fprintln(output, "Make sure Ollama is running:")
		fmt.Fprintln(output, "    ollama serve")
		return "", "", fmt.Errorf("unable to connect to Ollama: %w", err)
	}

	if containsModel(models, configuredModel) {
		return configuredModel, configuredSource, nil
	}

	for {
		if len(models) == 0 {
			fmt.Fprintf(
				output,
				"Model %q is not installed.\n\nNo local Ollama models are installed.\n\n",
				configuredModel,
			)
		} else {
			fmt.Fprintf(
				output,
				"Model %q is not installed.\n\nAvailable local models:\n\n",
				configuredModel,
			)

			for index, model := range models {
				fmt.Fprintf(output, "%d. %s\n", index+1, model)
			}

			fmt.Fprintln(output)
		}

		fmt.Fprintf(output, "[p] Pull %s\n", configuredModel)
		if len(models) > 0 {
			fmt.Fprintln(output, "[1-9] Use an installed model")
		}
		fmt.Fprintln(output, "[q] Quit")
		fmt.Fprint(output, "> ")

		if !input.Scan() {
			return "", "", fmt.Errorf("model selection canceled")
		}

		choice := strings.ToLower(strings.TrimSpace(input.Text()))
		switch choice {
		case "p":
			fmt.Fprintf(output, "Pulling model %q...\n", configuredModel)

			if err := manager.PullModel(configuredModel); err != nil {
				fmt.Fprintf(
					output,
					"Failed to pull model %q: %v\n",
					configuredModel,
					err,
				)
				continue
			}

			fmt.Fprintf(
				output,
				"Model %q pulled successfully.\n",
				configuredModel,
			)
			return configuredModel, "pulled", nil
		case "q":
			return "", "", fmt.Errorf("model selection canceled")
		default:
			index, err := strconv.Atoi(choice)
			if err == nil && index > 0 && index <= len(models) {
				return models[index-1], "local-selection", nil
			}

			fmt.Fprintln(output, "Invalid choice. Please choose one of the available options.")
		}
	}
}

func containsModel(models []string, target string) bool {
	for _, model := range models {
		if model == target {
			return true
		}
	}

	return false
}

func main() {
	appConfig, err := loadConfig()
	if err != nil {
		fmt.Println("Configuration error:", err)
		return
	}

	registry := toolspkg.NewToolRegistry()

	agentClient := agentpkg.NewAgent(registry)

	providerClient := providerpkg.Provider(&providerpkg.OllamaProvider{
		BaseURL: appConfig.Provider.BaseURL,
	})

	scanner := bufio.NewScanner(os.Stdin)
	model, modelSource := resolveModel(appConfig.Provider.Model)

	model, modelSource, err = selectModel(
		providerClient,
		model,
		modelSource,
		scanner,
		os.Stdout,
	)
	if err != nil {
		fmt.Println("Model selection error:", err)
		return
	}

	fmt.Printf(
		"Agent Harness starting... using model: %s (source: %s)\n",
		model,
		modelSource,
	)

	orchestratorClient := orchestratorpkg.NewOrchestrator(
		agentClient,
		providerClient,
	)

	fmt.Println("Agent Harness initialized...")

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
