package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	configpkg "agent-harness/config"
	agentpkg "agent-harness/internal/agent"
	mcppkg "agent-harness/internal/mcp"
	orchestratorpkg "agent-harness/internal/orchestrator"
	providerpkg "agent-harness/internal/provider"
	sessionpkg "agent-harness/internal/session"
	toolspkg "agent-harness/internal/tools"
)

type CommandResult int

const (
	CommandNotHandled CommandResult = iota
	CommandHandled
	CommandExit
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

		return "", "", fmt.Errorf(
			"unable to connect to Ollama: %w",
			err,
		)
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
			if err := input.Err(); err != nil {
				return "", "", fmt.Errorf(
					"model selection input error: %w",
					err,
				)
			}

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

			fmt.Fprintln(
				output,
				"Invalid choice. Please choose one of the available options.",
			)
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

func clearTerminal(output io.Writer) {
	fmt.Fprint(output, "\033[H\033[2J")
}

func handleCommand(
	input string,
	model string,
	modelSource string,
	currentSession **sessionpkg.Session,
	sessionStore sessionpkg.Store,
	agentClient *agentpkg.Agent,
	output io.Writer,
) CommandResult {
	parts := strings.Fields(input)

	if len(parts) == 0 {
		return CommandNotHandled
	}

	switch parts[0] {
	case "help":
		fmt.Fprintln(
			output,
			"Available commands: help, exit, model, session, clear",
		)
		return CommandHandled

	case "model":
		fmt.Fprintf(
			output,
			"Current model : %s\nSource: %s\n",
			model,
			modelSource,
		)
		return CommandHandled

	case "session":
		if len(parts) == 1 {
			fmt.Fprintf(
				output,
				"Current session: %s\n",
				(*currentSession).ID,
			)
			return CommandHandled
		}

		switch parts[1] {
		case "load":
			if len(parts) != 3 {
				fmt.Fprintln(output, "Usage: session load <id>")
				return CommandHandled
			}

			session, err := sessionStore.Get(parts[2])
			if err != nil {
				fmt.Fprintf(
					output,
					"Failed to load session: %v\n",
					err,
				)
				return CommandHandled
			}

			if err := agentClient.SetSession(session); err != nil {
				fmt.Fprintf(
					output,
					"Failed to activate session: %v\n",
					err,
				)
				return CommandHandled
			}

			*currentSession = session

			fmt.Fprintf(
				output,
				"Loaded session: %s\n",
				session.ID,
			)

			return CommandHandled

		case "list":
			sessions := sessionStore.List()

			fmt.Fprintln(output, "Sessions:")

			for _, session := range sessions {
				fmt.Fprintf(output, "- %s\n", session.ID)
			}

			return CommandHandled

		case "create":
			if len(parts) != 3 {
				fmt.Fprintln(output, "Usage: session create <id>")
				return CommandHandled
			}

			session := sessionpkg.NewSession(parts[2])

			if err := sessionStore.Set(session); err != nil {
				fmt.Fprintf(
					output,
					"Failed to create session: %v\n",
					err,
				)
				return CommandHandled
			}

			fmt.Fprintf(
				output,
				"Created session: %s\n",
				session.ID,
			)

			return CommandHandled

		case "delete":
			if len(parts) != 3 {
				fmt.Fprintln(output, "Usage: session delete <id>")
				return CommandHandled
			}

			if parts[2] == (*currentSession).ID {
				fmt.Fprintln(output, "Cannot delete current session.")
				return CommandHandled
			}

			if err := sessionStore.Delete(parts[2]); err != nil {
				fmt.Fprintf(
					output,
					"Failed to delete session: %v\n",
					err,
				)
				return CommandHandled
			}

			fmt.Fprintf(
				output,
				"Deleted session: %s\n",
				parts[2],
			)

			return CommandHandled

		default:
			fmt.Fprintln(output, "Unknown session command.")
			return CommandHandled
		}

	case "clear":
		clearTerminal(output)
		return CommandHandled

	case "exit":
		return CommandExit

	default:
		return CommandNotHandled
	}
}

func main() {
	appConfig, err := loadConfig()
	if err != nil {
		fmt.Println("Configuration error:", err)
		return
	}

	registry := toolspkg.NewToolRegistry()

	mcpRuntime := mcppkg.NewRuntime()
	defer func() {
		if err := mcpRuntime.Close(); err != nil {
			fmt.Println("MCP shutdown error:", err)
		}
	}()

	ctx := context.Background()

	for _, server := range appConfig.MCP.Servers {
		fmt.Printf("Connecting to MCP server: %s\n", server.Name)

		if err := mcpRuntime.ConnectServer(
			ctx,
			server.Name,
			server.Command,
			server.Args,
			registry,
		); err != nil {
			fmt.Printf("MCP server error: %v\n", err)
			return
		}

		fmt.Printf("Connected to MCP server: %s\n", server.Name)
	}

	agentClient := agentpkg.NewAgent(registry)

	providerClient := providerpkg.Provider(&providerpkg.OllamaProvider{
		BaseURL: appConfig.Provider.BaseURL,
	})

	scanner := bufio.NewScanner(os.Stdin)

	model, modelSource := resolveModel(
		appConfig.Provider.Model,
	)

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

	sessionStore := sessionpkg.NewFileSessionStore(".sessions")

	currentSession, err := sessionStore.Get("default")
	if err != nil {
		currentSession = sessionpkg.NewSession("default")

		if err := sessionStore.Set(currentSession); err != nil {
			fmt.Println("Session error:", err)
			return
		}
	}

	if err := agentClient.SetSession(currentSession); err != nil {
		fmt.Println("Session error:", err)
		return
	}

	fmt.Printf(
		"Agent Harness starting... using model: %s (source: %s)\n",
		model,
		modelSource,
	)

	fmt.Printf(
		"Current session: %s\n",
		currentSession.ID,
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

		if input == "" {
			continue
		}

		result := handleCommand(
			input,
			model,
			modelSource,
			&currentSession,
			sessionStore,
			agentClient,
			os.Stdout,
		)

		if result == CommandExit {
			break
		}

		if result == CommandHandled {
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

		if err := sessionStore.Set(currentSession); err != nil {
			fmt.Println("Session save error:", err)
			continue
		}

		fmt.Println(response.Content)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Input error:", err)
	}
}
