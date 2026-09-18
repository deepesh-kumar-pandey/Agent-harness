# Agent Harness

Agent Harness is a Go-based runtime for building and running tool-using AI agents. It provides provider abstraction, conversation history, native tool execution, MCP integration, credential management, session storage, and an interactive CLI.

The project is designed with a local-first architecture while keeping the core runtime provider-neutral and extensible.

## Current Status

The project currently supports:

* Provider abstraction with Ollama integration
* Conversation history and agent orchestration
* Native tool calling
* Built-in calculator, shell, and filesystem tools
* Tool registry with dynamic registration and removal
* MCP client integration
* MCP tool discovery and adaptation into the native tool system
* Credential storage and retrieval
* In-memory and file-based session storage
* Interactive CLI
* CLI model inspection and session display
* Local Ollama model management
* Unit and integration tests
* Go formatting, testing, and static analysis
* GitHub Actions CI

The core runtime is functional, with the current work focused on expanding session management, CLI functionality, MCP usage, and runtime capabilities.

---

## Features

### Provider Abstraction

The runtime uses provider-neutral interfaces so the agent does not depend directly on a specific model provider.

Currently integrated:

* Ollama

The provider layer supports:

* Chat requests
* Chat responses
* Conversation messages
* Tool definitions
* Tool calls
* Model listing
* Model availability checks
* Model pulling

---

### Agent Runtime

The agent maintains conversation history and coordinates model interaction with tool execution.

The runtime supports:

* User messages
* Assistant responses
* Tool calls
* Tool results
* Multiple tool calls
* Configurable maximum tool-call iterations
* Provider-neutral execution

---

### Tool Registry

The tool registry provides a central system for managing executable tools.

Built-in tools currently include:

* Calculator
* Shell
* Filesystem

The registry supports:

* Registering tools
* Looking up tools
* Checking whether a tool exists
* Listing tools
* Removing tools
* Generating tool schemas

---

### MCP Integration

The project includes support for the Model Context Protocol (MCP).

The MCP implementation currently provides:

* MCP client creation
* MCP server connections
* MCP tool discovery
* MCP tool metadata
* MCP tool execution
* Conversion of MCP tools into native Agent Harness tools

MCP tools can therefore participate in the same tool execution flow as built-in tools.

---

### Credential Management

The project includes a credential storage abstraction for managing provider or service credentials.

Credential functionality provides:

* Credential storage
* Credential retrieval
* Credential deletion
* A storage abstraction that can be extended later

Credentials are kept separate from provider logic.

---

### Session Storage

The project includes a session abstraction and session storage layer.

Current session functionality includes:

* Session creation
* Session IDs
* In-memory session storage
* File-based session storage
* Set
* Get
* Delete

The current `Session` object contains the session ID. Conversation history is still maintained by the Agent runtime and is not automatically persisted or restored by the CLI yet.

---

## Architecture

```text
                         ┌──────────────────────┐
                         │     Interactive CLI  │
                         └──────────┬───────────┘
                                    │
                                    ▼
                         ┌──────────────────────┐
                         │        Agent         │
                         │                      │
                         │ Conversation History │
                         │ Tool Execution       │
                         └──────────┬───────────┘
                                    │
                         ┌──────────┴───────────┐
                         │                      │
                         ▼                      ▼
                ┌─────────────────┐    ┌─────────────────┐
                │    Provider     │    │  Tool Registry  │
                │                 │    │                 │
                │ Ollama          │    │ Calculator      │
                │                 │    │ Shell           │
                │                 │    │ Filesystem      │
                └─────────────────┘    └────────┬────────┘
                                                 │
                                      ┌──────────┴──────────┐
                                      │                     │
                                      ▼                     ▼
                               Native Tools          MCP Tools
                                                           │
                                                           ▼
                                                    ┌─────────────┐
                                                    │ MCP Client  │
                                                    └─────────────┘
```

### Session Layer

```text
                    ┌────────────────────┐
                    │      Session       │
                    │                    │
                    │        ID          │
                    └─────────┬──────────┘
                              │
                     ┌────────┴────────┐
                     │                 │
                     ▼                 ▼
             In-Memory Store      File Store
```

The session layer is intentionally separated from the Agent runtime so session persistence can evolve independently.

---

## Runtime Flow

### Native Tool Execution

```text
User
 │
 ▼
Agent
 │
 ▼
Provider
 │
 ▼
Model
 │
 ├── Normal response ──────────────► Agent ──► User
 │
 └── Tool call
       │
       ▼
   Tool Registry
       │
       ▼
   Tool Execution
       │
       ▼
   Tool Result
       │
       ▼
     Agent
       │
       ▼
    Provider
       │
       ▼
    Final Response
       │
       ▼
      User
```

### MCP Tool Execution

```text
Agent
 │
 ▼
Tool Registry
 │
 ▼
MCP Tool Adapter
 │
 ▼
MCP Client
 │
 ▼
MCP Server
 │
 ▼
Tool Result
 │
 ▼
Agent
```

---

## Project Structure

```text
Agent-harness/
├── cmd/
│   ├── main.go
│   └── main_test.go
│
├── internal/
│   ├── agent/
│   │   ├── agent.go
│   │   └── ...
│   │
│   ├── config/
│   │   ├── config.go
│   │   └── ...
│   │
│   ├── credentials/
│   │   ├── credentials.go
│   │   └── ...
│   │
│   ├── mcp/
│   │   ├── client.go
│   │   ├── tool.go
│   │   ├── adapter.go
│   │   └── ...
│   │
│   ├── orchestrator/
│   │   ├── orchestrator.go
│   │   └── ...
│   │
│   ├── provider/
│   │   ├── provider.go
│   │   ├── ollama.go
│   │   └── ...
│   │
│   ├── session/
│   │   ├── session.go
│   │   ├── store.go
│   │   └── ...
│   │
│   └── tools/
│       ├── tool.go
│       ├── registry.go
│       ├── calculator.go
│       ├── shell.go
│       ├── filesystem.go
│       └── ...
│
├── .github/
│   └── workflows/
│
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

---

# Core Components

## Config

Configuration controls runtime behavior such as:

* Model
* Provider
* Base URL
* Endpoint
* Other provider-specific settings

A typical configuration looks like:

```json
{
  "model": "llama3.1",
  "provider": "ollama",
  "base_url": "http://localhost:11434",
  "endpoint": "/api/chat"
}
```

The configuration layer keeps provider-specific configuration outside the core Agent implementation.

---

## Credentials

Credential management is separated from provider configuration.

The credential layer provides an abstraction for:

```text
Set
Get
Delete
```

This allows credentials to be handled independently from model configuration.

Future credential backends can be added without changing the Agent or Provider interfaces.

---

## Sessions

A session represents a logical interaction context.

The current session abstraction is intentionally small:

```go
type Session struct {
    ID string
}
```

A session can be created with:

```go
session := NewSession("default")
```

### Session Store

The session store provides:

```text
Set
Get
Delete
```

Two storage implementations are currently available:

* In-memory session store
* File-based session store

The file-based implementation provides persistence for session objects.

### Current Limitation

Session persistence is currently separate from Agent conversation history.

The current implementation does **not** automatically:

* Save Agent conversation history to a session
* Restore conversation history from a session
* Switch between sessions through the CLI
* List saved sessions through the CLI

These are planned extensions.

---

## Provider

The Provider interface keeps the Agent independent from a particular model provider.

The provider layer is responsible for:

* Sending chat requests
* Receiving model responses
* Handling tool definitions
* Handling tool calls
* Listing models
* Checking model availability
* Pulling local models where supported

### Ollama

Ollama is currently the primary integrated provider.

The implementation supports local Ollama models and native tool calling.

The default Ollama endpoint is:

```text
http://localhost:11434
```

The default chat endpoint is:

```text
/api/chat
```

---

## Agent

The Agent is responsible for maintaining the interaction between the user, model, and tools.

Conceptually:

```text
User Input
    │
    ▼
Conversation History
    │
    ▼
Provider
    │
    ▼
Model Response
    │
    ├── Text ───────► User
    │
    └── Tool Call
            │
            ▼
       Tool Registry
            │
            ▼
       Tool Execution
            │
            ▼
       Tool Result
            │
            ▼
         Provider
```

The Agent is provider-neutral and operates through the Provider and Tool abstractions.

---

## Orchestrator

The orchestrator coordinates the Agent execution loop.

It handles repeated model/tool interactions until:

* The model produces a final response
* The maximum tool-call limit is reached
* An error occurs

The maximum number of tool-call iterations can be configured.

This prevents an agent from entering an uncontrolled tool execution loop.

---

# Tool System

## Tool Interface

Tools expose a common interface containing:

* Name
* Description
* Execute
* Schema

This allows different tool implementations to be treated uniformly.

---

## Tool Registry

The registry manages all available tools.

It supports:

```text
Register
Get
Has
List
Remove
Schemas
```

This provides a central mechanism for both native and externally discovered tools.

---

## Built-in Tools

### Calculator

The calculator supports:

```text
add
subtract
multiply
divide
modulus
```

Example:

```text
10 + 20
```

Result:

```text
30
```

---

### Shell

The shell tool allows command execution through the tool system.

It is intended for controlled local agent workflows.

---

### Filesystem

The filesystem tool provides filesystem-related operations through the common Tool interface.

---

# MCP

The project uses the official Go MCP SDK for MCP integration.

The MCP layer provides:

```text
MCP Client
     │
     ▼
Connect to Server
     │
     ▼
List Tools
     │
     ▼
Convert MCP Tools
     │
     ▼
Native Tool Interface
     │
     ▼
Tool Registry
```

MCP tools are adapted into the existing `tools.Tool` interface so the Agent does not need a separate execution mechanism for MCP tools.

---

## MCP Client

The MCP client is responsible for:

* Connecting to an MCP server
* Managing the MCP session
* Discovering tools
* Returning discovered tools

The client can discover MCP tools through `ListTools`.

---

## MCP Tool

An MCP tool exposes information such as:

```text
Name
Description
Schema
```

The MCP tool wrapper also implements the native Tool interface.

This allows an MCP tool to be executed by the existing Agent tool system.

---

## MCP Tool Adapter

The MCP adapter bridges the MCP SDK and the native Tool interface.

Conceptually:

```text
MCP SDK Tool
     │
     ▼
Tool Adapter
     │
     ▼
tools.Tool
     │
     ▼
Tool Registry
     │
     ▼
Agent
```

This keeps MCP-specific implementation details isolated from the Agent.

---

## MCP Tool Registration

Discovered MCP tools can be adapted and registered into the existing tool registry.

This means the Agent can eventually treat:

```text
Built-in Tool
MCP Tool
Future External Tool
```

as the same general tool type.

---

# CLI Usage

The interactive CLI is located in:

```text
cmd/
```

Run it with:

```bash
go run ./cmd
```

The CLI starts with a default active session:

```text
default
```

The current session ID can be displayed with the `session` command.

---

## CLI Commands

| Command   | Description                            |
| --------- | -------------------------------------- |
| `help`    | Show available commands                |
| `model`   | Show the currently configured model    |
| `session` | Show the current session ID            |
| `clear`   | Clear the current conversation history |
| `exit`    | Exit the CLI                           |

---

## `help`

Displays the available CLI commands.

```text
help
```

---

## `model`

Displays the currently configured model.

```text
model
```

The model can also be overridden through:

```bash
export OLLAMA_MODEL=your-model
```

---

## `session`

Displays the currently active session ID.

```text
session
```

Example:

```text
Current session: default
```

At the moment, the CLI does not yet provide commands for:

```text
session create
session list
session load
session delete
```

These are planned for future session management work.

---

## `clear`

Clears the current conversation history held by the Agent.

```text
clear
```

This currently affects the active Agent conversation history and does not represent persistent session deletion.

---

## `exit`

Exits the CLI.

```text
exit
```

---

# Ollama Setup

Install and start Ollama before running the Agent Harness.

Start the Ollama server:

```bash
ollama serve
```

Verify available models:

```bash
ollama list
```

Pull a model if required:

```bash
ollama pull llama3.1
```

The Agent Harness can then be started with:

```bash
go run ./cmd
```

---

## Model Override

The model can be overridden using the environment variable:

```bash
export OLLAMA_MODEL=llama3.1
```

This is useful for testing different local models without changing the configuration file.

---

# Testing

The project uses Go's standard testing framework.

Run all tests from the project root:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

The CLI package can also be tested directly:

```bash
go test ./cmd
```

---

## Integration Testing

The Ollama orchestrator integration test can be enabled with:

```bash
ORCHESTRATOR_INTEGRATION=1 go test ./...
```

The integration test requires:

* Ollama running locally
* A compatible model available
* The configured model supporting the required tool-calling behavior

---

# Formatting

Format Go files with:

```bash
gofmt -w $(find . -name '*.go')
```

Verify formatting with:

```bash
gofmt -l $(find . -name '*.go')
```

The second command should produce no output when all Go files are formatted.

---

# Static Analysis

Run Go vet with:

```bash
go vet ./...
```

The expected result is no reported issues.

---

# Continuous Integration

GitHub Actions is used to automate project validation.

The CI workflow runs project checks such as:

```text
go test
go vet
```

The workflow is intended to validate changes on the repository's configured branches and pull requests.

---

# End-to-End Example

A simplified execution looks like this:

```text
User
 │
 │ "What is 10 + 20?"
 ▼
CLI
 │
 ▼
Agent
 │
 ▼
Ollama Provider
 │
 ▼
Local Model
 │
 │ Tool Call
 ▼
Tool Registry
 │
 ▼
Calculator
 │
 │ 10 + 20
 ▼
30
 │
 ▼
Agent
 │
 ▼
Ollama Provider
 │
 ▼
Final Response
 │
 ▼
CLI
```

For an MCP tool:

```text
User
 │
 ▼
Agent
 │
 ▼
Ollama
 │
 │ MCP Tool Call
 ▼
Tool Registry
 │
 ▼
MCP Adapter
 │
 ▼
MCP Client
 │
 ▼
MCP Server
 │
 ▼
Tool Result
 │
 ▼
Agent
 │
 ▼
Final Response
```

---

# Architectural Principles

## Provider Neutrality

The Agent should not depend directly on Ollama-specific APIs.

Provider-specific behavior belongs behind the Provider interface.

---

## Tool Agnosticism

The Agent should not need to know whether a tool is:

* Built into the project
* Provided by an MCP server
* Added by another future integration

All tools should be exposed through the common Tool abstraction.

---

## Separation of Concerns

The project separates:

```text
Configuration
Credentials
Sessions
Providers
Agent
Orchestration
Tools
MCP
CLI
```

Each subsystem has a focused responsibility.

---

## Local First

The current implementation is designed to work with local models through Ollama.

This allows development and testing without requiring a hosted model API.

---

## Extensibility

The architecture is designed so new capabilities can be added without rewriting the Agent core.

Potential future extensions include:

```text
Additional Providers
Additional Tool Sources
Additional Credential Backends
Additional Session Backends
Remote MCP Servers
Persistent Conversation Storage
```

---

# Implementation Status

| Component                             | Status      |
| ------------------------------------- | ----------- |
| Go project structure                  | Implemented |
| Provider abstraction                  | Implemented |
| Ollama provider                       | Implemented |
| Conversation history                  | Implemented |
| Agent runtime                         | Implemented |
| Tool abstraction                      | Implemented |
| Tool registry                         | Implemented |
| Calculator tool                       | Implemented |
| Shell tool                            | Implemented |
| Filesystem tool                       | Implemented |
| MCP client                            | Implemented |
| MCP tool discovery                    | Implemented |
| MCP tool adapter                      | Implemented |
| MCP tool registration foundation      | Implemented |
| Credential abstraction                | Implemented |
| Credential storage                    | Implemented |
| Session abstraction                   | Implemented |
| In-memory session store               | Implemented |
| File-based session store              | Implemented |
| Interactive CLI                       | Implemented |
| CLI model command                     | Implemented |
| CLI session display                   | Implemented |
| CLI conversation clearing             | Implemented |
| CLI session management                | Planned     |
| Persistent Agent conversation history | Planned     |
| Session restore                       | Planned     |
| Multiple active sessions              | Planned     |
| Additional providers                  | Planned     |

---

# Roadmap

## Session Management

Planned CLI commands:

```text
session create
session list
session load
session delete
```

The goal is to allow users to manage multiple persistent sessions directly from the CLI.

---

## Persistent Conversation History

The next stage of session development is connecting Agent conversation history with the session storage layer.

Future flow:

```text
Session
 │
 ▼
Conversation History
 │
 ▼
Persistent Storage
```

This will allow conversations to survive application restarts.

---

## Improved MCP Support

Future MCP work includes:

* CLI configuration of MCP servers
* MCP server lifecycle management
* Dynamic MCP server registration
* Better error handling
* More complete MCP resource support
* MCP prompts
* MCP sampling where applicable

---

## Additional Providers

The provider abstraction allows future integrations such as:

```text
OpenAI-compatible APIs
Anthropic
Google
Other local model runtimes
```

Provider implementations can be added without changing the core Agent architecture.

---

## Improved CLI

Planned CLI improvements include:

```text
Session management
Provider management
Tool listing
MCP server management
Configuration inspection
Model management
```

---

# Development Workflow

A typical development workflow is:

```bash
git checkout -b feature/<name>
```

Make the required changes and run:

```bash
gofmt -w $(find . -name '*.go')
go test ./...
go vet ./...
```

Then review the changes:

```bash
git status
git diff
```

Commit the completed logical change:

```bash
git add .
git commit -m "your commit message"
```

Push the feature branch:

```bash
git push -u origin feature/<name>
```

Then open a pull request against the appropriate base branch.

---

# Project Goals

The long-term goal of Agent Harness is to provide a modular runtime for building tool-using AI agents while keeping the underlying architecture understandable and extensible.

The project focuses on:

```text
Provider Abstraction
        +
Agent Runtime
        +
Tool Execution
        +
MCP
        +
Credentials
        +
Sessions
        +
CLI
```

The architecture is intentionally built incrementally so each subsystem can be tested independently before being connected to the larger runtime.

---

# License

This project is licensed under the MIT License.

See the [LICENSE](LICENSE) file for the full license text.
