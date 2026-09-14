# Agent Harness

Agent Harness is a Go-based runtime for building and running tool-using AI agents. It provides provider abstraction, conversation history, native tool execution, MCP integration, credential management, and an interactive CLI.

The project is designed with a local-first architecture while keeping the core runtime provider-neutral and extensible.

## Current Status

The project currently supports:

* Provider-neutral LLM communication
* Ollama integration with native tool calling
* Local model discovery and pulling
* Native local tools
* MCP client creation and connection
* MCP tool discovery
* MCP tool wrappers and execution
* Credential storage
* In-memory conversation history
* Agent and orchestrator execution
* Interactive CLI
* Unit and integration testing

Ollama remains the current fully integrated provider. Additional cloud providers can be added through the provider abstraction without changing the Agent or Orchestrator layers.

MCP support is currently implemented through the client and tool-wrapper layers. The next integration step is exposing discovered MCP tools through the existing Agent and Tool Registry.

## Features

* Go-based agent runtime
* Provider abstraction with provider-neutral `Message`, `ChatRequest`, `ChatResponse`, `ToolCall`, and `ToolDefinition` types
* Ollama provider with native function/tool calling
* JSON Schema-based tool definitions
* Ollama local model discovery and model pulling
* Tool Registry with registration, lookup, listing, removal, and schema exposure
* Agent layer for tool access and in-memory conversation history
* Orchestrator for provider communication and tool execution
* Multiple native tool calls in a single provider response
* Structured JSON tool-call handling
* Maximum tool-call protection through `WithMaxToolCalls`
* Calculator, filesystem, and shell tools
* MCP client creation and server connection
* MCP tool discovery
* MCP tool abstraction
* MCP tool execution
* Local credential storage
* Interactive CLI
* Unit tests
* Ollama integration tests
* End-to-end native tool execution

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
                           Agent
                        /         \
                       /           \
                      v             v
              Tool Registry      MCP Client
                   |                 |
                   v                 v
            Native Tools        MCP Servers
```

The layers have distinct responsibilities:

* **Config** loads and validates provider configuration.
* **Credentials** stores and retrieves provider credentials separately from general configuration.
* **Provider** communicates with the LLM and exposes provider-neutral types. The Ollama implementation converts those types to and from the Ollama API.
* **Agent** owns the Tool Registry and in-memory conversation history, and provides tool execution.
* **Orchestrator** coordinates the agent loop, provider communication, tool calls, tool results, and maximum tool-call protection.
* **Tool Registry** registers, retrieves, lists, removes, and exposes schemas for native tools.
* **MCP Client** creates MCP client sessions, connects to MCP servers, and discovers available MCP tools.
* **MCP Tool Wrapper** represents an MCP tool locally and forwards execution requests to the connected MCP server.
* **Concrete Tools** perform local operations.
* **MCP Servers** provide external tools through the Model Context Protocol.

The LLM/provider produces the tool-call decision. The Agent does not plan or decide what should happen. The Orchestrator coordinates execution, the Registry manages native tools, the MCP layer manages MCP tools, and the selected tool performs the operation.

## Runtime Flow

For a normal native-tool chat request, the runtime follows this sequence:

1. The CLI creates a `ChatRequest` containing the user's message.
2. The Orchestrator adds the message to the Agent's in-memory history.
3. The Agent exposes available native tool definitions.
4. The Provider receives the conversation and available tool definitions.
5. The Provider returns either a final response or one or more tool calls.
6. The Orchestrator executes each tool call through the Agent.
7. The Agent resolves the requested native tool through the Tool Registry.
8. The tool result is added to the conversation.
9. The updated conversation is sent back to the Provider.
10. The loop continues until the Provider returns a final response or the configured maximum tool-call limit is reached.

MCP currently has a separate client flow:

1. The MCP client is created.
2. The client connects to an MCP server through an SDK transport.
3. A client session is established.
4. Available MCP tools are requested from the server.
5. Each discovered MCP tool can be wrapped by the Agent Harness MCP tool abstraction.
6. The wrapper forwards execution requests to the MCP server through the active session.

The MCP tools are not yet automatically registered with the existing native Tool Registry. Integrating that boundary is the next planned step.

The orchestrator also accepts structured JSON responses containing a `content` value and optional `tool_call` object.

Conversation history is retained only in memory for the lifetime of the Agent instance.

## Project Structure

```text
agent-harness/
├── cmd/
│   └── agent/
│       ├── main.go
│       └── main_test.go
├── config/
│   ├── config.go
│   └── config.example.json
├── internal/
│   ├── agent/
│   │   ├── agent.go
│   │   ├── agent_test.go
│   │   ├── history.go
│   │   └── history_test.go
│   ├── credentials/
│   │   ├── credentials.go
│   │   ├── store.go
│   │   └── store_test.go
│   ├── mcp/
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── tools.go
│   │   └── tools_test.go
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
├── go.mod
├── go.sum
├── LICENSE
├── .gitignore
└── README.md
```

## Core Components

### Config

`config/config.go` loads and validates provider configuration.

Configuration contains provider information such as:

* Provider name
* Model
* Base URL
* API endpoint

Credentials are intentionally kept separate from the configuration system.

### Credentials

The `internal/credentials` package provides a simple credential abstraction:

```go
type CredentialStore interface {
    Set(provider string, key string) error
    Get(provider string) (string, error)
    Delete(provider string) error
}
```

The current implementation uses a local file store.

Credentials are stored outside the repository under:

```text
~/.agent-harness/credentials.json
```

The credential store provides three operations:

* **Set** stores or updates a credential for a provider.
* **Get** retrieves the stored credential for a provider.
* **Delete** removes the stored credential for a provider.

The credential layer is intentionally separated behind an interface so the underlying storage can be replaced later without changing the rest of the runtime.

### Provider

The `Provider` interface exposes LLM communication without coupling the rest of the application to a specific provider.

The current Ollama implementation:

* Sends non-streaming chat requests
* Converts JSON Schema tool definitions into Ollama function tools
* Decodes native tool calls
* Lists local models through `/api/tags`
* Pulls models through `/api/pull`

Additional providers can be implemented behind the same provider boundary.

### Agent

The Agent:

* Owns the Tool Registry
* Owns conversation history
* Exposes tool definitions
* Executes named tools
* Validates conversation messages

The Agent does not select tools or make planning decisions.

### Orchestrator

The Orchestrator coordinates:

* Provider communication
* Agent interaction
* Tool calls
* Tool results
* Conversation updates
* Structured tool-call responses
* Maximum tool-call protection

`RunAgent` contains the main execution loop that allows the provider to request tools and receive their results.

### Tool Registry

The default registry contains:

* `calculator`
* `shell`
* `filesystem`

The Registry supports:

* `Register`
* `Remove`
* `Get`
* `Has`
* `List`
* `Schemas`

Each registered native tool exposes:

* Name
* Description
* Execution method
* JSON Schema parameter definition

### MCP

The `internal/mcp` package provides the foundation for Model Context Protocol integration.

MCP allows Agent Harness to communicate with external MCP servers and use tools provided by those servers.

The MCP client currently provides the following functionality:

#### `NewClient`

Creates an Agent Harness MCP client backed by the official MCP SDK.

The client identifies itself to MCP servers using the Agent Harness implementation name and version.

```go
client := mcp.NewClient()
```

#### `Connect`

Connects the MCP client to an MCP server using an SDK transport and returns an active `ClientSession`.

```go
session, err := client.Connect(ctx, transport)
```

The transport determines how communication takes place. The current abstraction accepts any MCP SDK `Transport`, allowing transports such as in-memory or command-based transports to be used without changing the Agent Harness client API.

#### `ListTools`

Requests the available tools from an MCP server through an active client session.

```go
tools, err := client.ListTools(ctx, session)
```

The method returns the MCP SDK tool definitions discovered from the server.

#### `NewTool`

Wraps an MCP SDK tool in the Agent Harness MCP tool abstraction.

```go
tool := mcp.NewTool(toolDefinition)
```

The wrapper keeps the MCP tool definition locally while allowing execution through the associated MCP session.

#### `Name`

Returns the name provided by the MCP server for the wrapped tool.

```go
name := tool.Name()
```

#### `Description`

Returns the description provided by the MCP server for the wrapped tool.

```go
description := tool.Description()
```

#### `Execute`

Forwards a tool execution request to the MCP server through the active client session.

```go
result, err := tool.Execute(ctx, session, args)
```

The method accepts:

* `context.Context` for operation lifecycle, cancellation, and deadlines
* An active MCP `ClientSession`
* `map[string]any` containing the tool-specific arguments

The MCP server executes the requested operation and returns an MCP `CallToolResult`.

The MCP layer is kept separate from native tools:

```text
Native Tools
     |
Tool Registry
     |
     +----------------+
                      |
                    Agent
                      |
                 Orchestrator
                      |
                  MCP Client
                      |
                 MCP Servers
```

MCP is an extension of the existing tool architecture rather than a replacement for native tools.

The current MCP implementation provides client creation, server connection, tool discovery, tool wrapping, and remote tool execution. The next step is integrating these wrapped tools with the existing Agent and Tool Registry.

## Built-in Tools

### Calculator

Performs:

* `add`
* `subtract`
* `multiply`
* `divide`
* `modulus`

over numeric arrays.

### Filesystem

Supports:

* `read`
* `write`
* `list`
* `exists`
* `delete`

### Shell

Executes an available operating-system command with optional string arguments and returns combined output.

The tools operate on the local machine and should be used with appropriate care.

## CLI Usage

The CLI entry point is:

```bash
go run ./cmd/agent
```

The CLI loads the configured provider and model and provides an interactive terminal interface.

Current commands include:

```text
help
model
clear
exit
```

Any other input is treated as a chat request.

The active Ollama model can be overridden for the current run with:

```bash
OLLAMA_MODEL="your-model-name" go run ./cmd/agent
```

Credential management is handled separately from provider configuration.

## Ollama Setup

Start Ollama before running the CLI:

```bash
ollama serve
```

The default provider configuration is stored in:

```text
config/config.example.json
```

Example:

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

If the configured model is not installed, the CLI can offer an explicit pull option or allow selection of an installed local model.

It does not silently download or change the configured model.

## Credentials

Credentials are stored separately from provider configuration.

The credential store is located at:

```text
~/.agent-harness/credentials.json
```

The credentials layer currently provides:

* **Set** — stores or updates a provider credential.
* **Get** — retrieves a provider credential.
* **Delete** — removes a provider credential.

Provider configuration does not contain API keys or other credentials.

## Testing

Run all tests from the repository root:

```bash
go test ./...
```

Run static checks:

```bash
go vet ./...
```

The MCP package includes unit tests for:

* MCP client creation
* MCP client connection
* MCP tool discovery
* MCP tool wrapper creation
* MCP tool name and description access
* MCP tool execution

The MCP tests use the SDK's in-memory transport, allowing client/server communication to be tested without requiring an external MCP server.

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
go run ./cmd/agent
```

Then enter:

```text
Calculate 12 multiplied by 7.
```

The Provider can return a native calculator tool call.

The Orchestrator passes the call to the Agent.

The Agent resolves `calculator` through the Tool Registry.

The Calculator executes the operation.

The result is added to the conversation and sent back to the Provider.

The Provider then produces the final response.

The same architecture is being extended to support tools provided through MCP.

## Architectural Principles

### Separation of Concerns

Provider communication, coordination, tool access, registry management, credential storage, MCP communication, and concrete operations remain separate.

### Provider-Neutral Contracts

The Agent and Orchestrator use shared message and tool-call types rather than provider-specific response structures.

### Explicit Tool Boundaries

Every native tool implements the same interface and publishes its JSON Schema.

MCP tools are exposed through the MCP layer while remaining separate from native tool implementations until they are integrated into the shared tool boundary.

### Bounded Execution

`WithMaxToolCalls` prevents an unbounded tool-execution loop.

### In-Memory State

Conversation history belongs to the Agent and is not persisted between process runs.

### Local-First Operation

The current fully integrated provider is Ollama, allowing the runtime to operate against a local LLM.

### Extensible Architecture

Provider implementations, native tools, MCP servers, and credential storage can be extended without requiring the core Agent and Orchestrator architecture to be rewritten.

## Implementation Status

| Component                                        | Status      |
| ------------------------------------------------ | ----------- |
| Configuration loading and validation             | Implemented |
| Provider abstraction                             | Implemented |
| Ollama chat and native tool calling              | Implemented |
| Local model discovery and pulling                | Implemented |
| Provider-neutral message and tool types          | Implemented |
| Agent tool access and conversation history       | Implemented |
| Tool Registry and JSON Schema exposure           | Implemented |
| Calculator, filesystem, and shell tools          | Implemented |
| Orchestrator agent loop and tool-result feedback | Implemented |
| Multiple native tool calls per provider response | Implemented |
| Maximum tool-call protection                     | Implemented |
| Interactive CLI                                  | Implemented |
| Credential abstraction                           | Implemented |
| Local file credential store                      | Implemented |
| Credential Set/Get/Delete operations             | Implemented |
| MCP client creation                              | Implemented |
| MCP server connection                            | Implemented |
| MCP tool discovery                               | Implemented |
| MCP tool wrapper                                 | Implemented |
| MCP tool name and description access             | Implemented |
| MCP tool execution                               | Implemented |
| MCP integration with Agent/Tool Registry         | Planned     |
| Unit and integration tests                       | Implemented |

## Roadmap

The following are planned extensions:

* Integrate MCP tools with the existing Agent and Tool Registry
* Expose native and MCP tools through a unified tool boundary
* Complete broader MCP session and transport management
* Additional LLM provider implementations
* CLI credential management commands
* Persistent sessions and conversation storage
* Search and information-retrieval tools
* Additional native and MCP tools
* Additional deployment options
