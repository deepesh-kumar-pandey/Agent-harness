# Agent Harness

Agent Harness is a Go-based runtime that connects an Ollama language model to registered local tools. It provides the provider boundary, tool execution loop, conversation history, and interactive CLI needed to run a small tool-using agent locally.

## Current Status

The current MVP is implemented around Ollama and supports native tool calling, local model discovery and pulling, three built-in tools, in-memory conversation history, and an interactive terminal client. The project is local-first: Ollama must be running on the host that runs the CLI or integration tests.

## Features

- Go-based agent harness
- Provider abstraction with provider-neutral `Message`, `ChatRequest`, `ChatResponse`, `ToolCall`, and `ToolDefinition` types
- Ollama provider with native function/tool calling
- JSON Schema-based tool definitions
- Ollama local model discovery and model pulling
- Tool Registry with registration, lookup, listing, removal, and schema exposure
- Agent layer for tool access and in-memory conversation history
- Orchestrator for provider communication and tool execution
- Multiple native tool calls in a single provider response
- Structured JSON tool-call handling
- Maximum tool-call protection through `WithMaxToolCalls`
- Calculator, filesystem, and shell tools
- Interactive CLI
- Unit tests, Ollama integration tests, and end-to-end tool execution

## Architecture

```text
User
  |
  v
CLI
  |
  v
Orchestrator
  |
  v
Provider
  |
  v
Agent
  |
  v
Tool Registry
  |
  v
Concrete Tools
```

The layers have distinct responsibilities:

- **Config** loads and validates provider configuration.
- **Provider** communicates with the LLM and exposes provider-neutral types. The Ollama implementation converts those types to and from the Ollama API.
- **Agent** owns the Tool Registry and in-memory conversation history, and provides tool execution.
- **Orchestrator** coordinates the agent loop, provider communication, tool calls, tool results, and maximum tool-call protection.
- **Tool Registry** registers, retrieves, lists, removes, and exposes schemas for tools.
- **Concrete Tools** perform the requested operations.

The LLM/provider produces the tool-call decision. The Agent does not plan or decide what should happen. The Orchestrator coordinates execution, the Registry manages available tools, and the selected concrete Tool performs the operation.

## Runtime Flow

For a normal chat request, the runtime follows this sequence:

1. The CLI creates a `ChatRequest` containing the user's message.
2. The Orchestrator adds the message to the Agent's in-memory history and sends the conversation plus registered tool definitions to the Provider.
3. The Provider returns either a final response or one or more native tool calls.
4. The Orchestrator executes each tool call through the Agent and adds the tool results to the conversation.
5. The Orchestrator sends the updated conversation back to the Provider.
6. The loop ends when the Provider returns a final response or the configured maximum tool-call limit is reached.

The orchestrator also accepts structured JSON responses containing a `content` value and optional `tool_call` object. Conversation history is retained only in memory for the lifetime of the Agent instance.

## Project Structure

```text
agent-harness/
├── cmd/
│   ├── main.go
│   └── main_test.go
├── config/
│   ├── config.go
│   ├── config.json
│   └── config_test.go
├── internal/
│   ├── agent/
│   │   ├── agent.go
│   │   ├── agent_test.go
│   │   ├── history.go
│   │   └── history_test.go
│   ├── orchestrator/
│   │   ├── orchestrator.go
│   │   ├── orchestrator_integration_test.go
│   │   ├── orchestrator_test.go
│   │   ├── response.go
│   │   └── response_test.go
│   ├── provider/
│   │   ├── provider.go
│   │   ├── provider_integration_test.go
│   │   └── provider_test.go
│   └── tools/
│       ├── calculator.go
│       ├── calculator_test.go
│       ├── filesystem_tool.go
│       ├── filesystem_tool_test.go
│       ├── registry_test.go
│       ├── shell_tool.go
│       ├── shell_tool_test.go
│       ├── ToolRegistry.go
│       └── tools.go
├── filesystem/
│   ├── filesystem.go
│   └── filesystem_test.go
├── shell/
│   ├── shell.go
│   └── shell_test.go
├── go.mod
├── LICENSE
└── README.md
```

## Core Components

### Config

`config/config.go` loads and validates the JSON provider configuration. The current configuration selects Ollama, a model, its base URL, and the chat endpoint.

### Provider

The `Provider` interface exposes chat communication without coupling the rest of the application to Ollama-specific types. `OllamaProvider` sends non-streaming chat requests, converts JSON Schema tool definitions into Ollama function tools, decodes native tool calls, lists local models through `/api/tags`, and pulls models through `/api/pull`.

### Agent

The Agent receives a Tool Registry, executes named tools, exposes their definitions, and owns a `Conversation` containing provider messages. It validates message roles when adding history. It does not select tools or make planning decisions.

### Orchestrator

The Orchestrator provides direct tool execution, provider chat, structured tool assignment, and `RunAgent`. `RunAgent` coordinates the provider response, native or structured tool calls, tool results, conversation updates, and the `WithMaxToolCalls` guard.

### Tool Registry

The default registry contains `calculator`, `shell`, and `filesystem`. The Registry supports:

- `Register` and `Remove`
- `Get` and `Has`
- `List`
- `Schemas`

Each registered tool exposes a name, description, execution method, and JSON Schema parameter definition through the `Tool` interface.

## Built-in Tools

- **Calculator** performs `add`, `subtract`, `multiply`, `divide`, and `modulus` operations over numeric arrays.
- **Filesystem** supports local `read`, `write`, `list`, `exists`, and `delete` operations.
- **Shell** executes an available operating-system command with optional string arguments and returns combined output.

The tools operate on the local machine and should be used with appropriate care.

## CLI Usage

The current CLI entry point is:

```bash
go run ./cmd
```

At startup, the CLI loads `config/config.json`, checks the configured Ollama model, and lets you pull it or select an installed local model when it is unavailable. Once started, these commands are available:

```text
help
model
clear
exit
```

Any other input is sent as a chat request. The active model can be overridden for the current run with `OLLAMA_MODEL`:

```bash
OLLAMA_MODEL="your-model-name" go run ./cmd
```

## Ollama Setup

Start Ollama before running the CLI:

```bash
ollama serve
```

The default provider configuration is stored in `config/config.json`:

```json
{
  "provider": {
    "name": "ollama",
    "model": "llama3.1",
    "base_url": "http://localhost:11434",
    "endpoint": "/api/chat"
  }
}
```

If the configured model is not installed, the CLI offers an explicit pull option or lets you choose an installed model for the current session. It does not silently download or change the configured model.

## Testing

Run the unit and package tests from the repository root:

```bash
go test ./...
```

Run static checks:

```bash
go vet ./...
```

The orchestrator integration test is opt-in and requires a running Ollama service and an available configured model:

```bash
ORCHESTRATOR_INTEGRATION=1 go test -run TestOrchestratorRunAgent_Integration ./internal/orchestrator -v
```

The provider package also contains a separate local Ollama integration test:

```bash
go test -run TestOllamaProvider_Integration ./internal/provider -v
```

## End-to-End Example

With Ollama running and a model available:

```bash
go run ./cmd
```

Then enter a request such as:

```text
Calculate 12 multiplied by 7.
```

The Provider can return a native calculator tool call. The Orchestrator passes it to the Agent, the Tool Registry resolves `calculator`, the Calculator executes it, and the result is sent back to the Provider for the final response.

## Architectural Principles

- **Separation of concerns:** provider communication, coordination, tool access, registry management, and concrete operations remain separate.
- **Provider-neutral contracts:** the Orchestrator and Agent use shared message and tool-call types rather than Ollama-specific response structures.
- **Explicit tool boundaries:** every tool implements the same interface and publishes its JSON Schema.
- **Bounded execution:** `WithMaxToolCalls` prevents an unbounded tool-execution loop.
- **In-memory state:** conversation history belongs to the Agent and is not persisted between process runs.
- **Local-first operation:** the current implementation is built around a local Ollama service.

## Implementation Status

| Component | Status |
| --- | --- |
| Configuration loading and validation | Implemented |
| Provider abstraction | Implemented |
| Ollama chat and native tool calling | Implemented |
| Local model discovery and pulling | Implemented |
| Provider-neutral message and tool types | Implemented |
| Agent tool access and conversation history | Implemented |
| Tool Registry and JSON Schema exposure | Implemented |
| Calculator, filesystem, and shell tools | Implemented |
| Orchestrator agent loop and tool-result feedback | Implemented |
| Multiple native tool calls per provider response | Implemented |
| Maximum tool-call protection | Implemented |
| Interactive CLI | Implemented |
| Unit and integration tests | Implemented |

## Roadmap

The following are planned possibilities, not current capabilities:

- Additional LLM provider implementations
- Persistent sessions and conversation storage
- Search and information-retrieval tools
- MCP integration
- Additional concrete tools and deployment options
