# Agent Harness

Agent Harness is a Go-based runtime for building and running tool-using AI agents.

It provides:

* Provider abstraction
* Agent orchestration
* Conversation history
* Native tool execution
* MCP integration
* MCP server lifecycle management
* Credential storage
* Session management and persistence
* Interactive CLI
* Local Ollama model management
* Automated testing and CI

The project is designed with a **local-first architecture** while keeping the core runtime provider-neutral, modular, testable, and extensible.

---

# Current Status

The project currently supports:

* Provider abstraction with Ollama integration
* Conversation history
* Agent orchestration
* Native tool calling
* Built-in calculator, shell, and filesystem tools
* Dynamic tool registry
* MCP client integration
* MCP server connections through command transport
* MCP tool discovery
* MCP tool adaptation into the native tool system
* MCP tool registration into the Tool Registry
* MCP runtime/session lifecycle management
* Credential storage and retrieval
* In-memory session storage
* File-based session storage
* Session-aware Agent runtime
* Persistent Agent conversation history
* Default session restoration
* CLI session creation, listing, loading, and deletion
* Interactive CLI
* CLI model inspection
* CLI session inspection
* Local Ollama model management
* Unit and integration tests
* Go formatting and static analysis
* GitHub Actions CI

The core runtime is functional.

Current development is focused on:

* Expanding MCP functionality
* Improving session management
* Improving CLI functionality
* Adding additional providers
* Increasing runtime extensibility

---

# Features

## Provider Abstraction

The runtime uses provider-neutral interfaces so the Agent does not depend directly on a specific model provider.

Currently integrated:

* Ollama

The Provider layer supports:

* Chat requests
* Chat responses
* Conversation messages
* Tool definitions
* Tool calls
* Model listing
* Model availability checks
* Model pulling

Provider-specific behavior remains behind the Provider abstraction.

---

# Agent Runtime

The Agent manages conversation state and provides access to the tool system.

The Orchestrator coordinates communication between the Agent and the Provider.

The runtime supports:

* User messages
* Assistant responses
* Tool calls
* Tool results
* Multiple tool calls
* Configurable maximum tool-call iterations
* Provider-neutral execution
* Session-aware conversation history

The Agent can operate using:

1. Its internal conversation history
2. An explicitly assigned Session

When a Session is assigned, the Session becomes the source of conversation history.

---

# Tool System

## Tool Interface

Tools expose a common interface containing:

```text
Name
Description
Execute
Schema
```

This allows different tool implementations to be treated uniformly.

The same abstraction is used for:

* Built-in tools
* MCP tools
* Future external tools

---

# Tool Registry

The Tool Registry provides a central system for managing executable tools.

It supports:

```text
Register
Get
Has
List
Remove
Schemas
```

The registry provides the common execution layer used by both native tools and MCP tools.

---

# Built-in Tools

## Calculator

The Calculator tool supports:

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

## Shell

The Shell tool allows controlled local command execution through the common Tool interface.

---

## Filesystem

The Filesystem tool provides filesystem-related operations through the common Tool interface.

---

# MCP Integration

The project uses the official Go MCP SDK for MCP integration.

The MCP layer provides:

* MCP client creation
* MCP server connections
* MCP session management
* MCP tool discovery
* MCP tool metadata
* MCP tool execution
* MCP tool adaptation
* MCP tool registration
* MCP server lifecycle management

MCP tools are converted into the existing `tools.Tool` interface so the Agent does not require a separate execution path for MCP tools.

---

# MCP Architecture

The MCP flow is:

```text
MCP Server
    │
    ▼
Command Transport
    │
    ▼
MCP Client
    │
    ▼
Client Session
    │
    ▼
List Tools
    │
    ▼
MCP Tool
    │
    ▼
Tool Adapter
    │
    ▼
Native Tool Interface
    │
    ▼
Tool Registry
    │
    ▼
Agent
```

This allows MCP tools and native tools to participate in the same Agent execution flow.

---

# MCP Client

The MCP client is implemented in:

```text
internal/mcp/client.go
```

The client provides:

* MCP client creation
* Transport-based connections
* Command-based connections
* MCP tool discovery

The main operations are:

```text
NewClient
Connect
ConnectCommand
ListTools
```

---

# MCP Command Transport

The current runtime supports local MCP servers through command-based transport.

The client creates the command using Go's:

```go
exec.CommandContext
```

The command is then provided to the MCP SDK's command transport.

Conceptually:

```text
Agent Harness
     │
     ▼
exec.CommandContext
     │
     ▼
MCP Test / External Server
     │
     ▼
stdin/stdout
     │
     ▼
MCP Protocol
```

Using `exec.CommandContext` allows the MCP server process to be associated with a context and terminated when that context is cancelled.

---

# MCP Tool

The MCP tool wrapper is implemented in:

```text
internal/mcp/tool.go
```

It provides:

* Tool name
* Tool description
* Tool schema
* Tool execution

The MCP tool uses the MCP client session to call the remote MCP server.

---

# MCP Tool Adapter

The MCP Tool Adapter is implemented in:

```text
internal/mcp/adapter.go
```

The adapter bridges the MCP SDK and the native Tool interface.

Conceptually:

```text
MCP SDK Tool
     │
     ▼
MCP Tool Wrapper
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

The Agent therefore does not need to know whether a tool originated from:

* Built-in code
* MCP
* Another future tool provider

---

# MCP Tool Registration

The MCP package provides tool registration functionality that:

1. Lists tools from an MCP server
2. Wraps each MCP tool
3. Creates a native Tool Adapter
4. Registers the adapter in the Tool Registry

This allows MCP tools to automatically become available to the Agent.

---

# MCP Runtime

The MCP runtime is implemented in:

```text
internal/mcp/runtime.go
```

The runtime is responsible for managing MCP server connections during the lifetime of the application.

The runtime:

* Connects to configured MCP servers
* Starts command-based MCP server processes
* Creates MCP client sessions
* Discovers MCP tools
* Registers MCP tools into the native Tool Registry
* Keeps MCP sessions alive while the application runs
* Closes MCP sessions during shutdown

The runtime keeps the client sessions alive because MCP Tool Adapters use those sessions when executing tools.

---

# MCP Runtime Lifecycle

The lifecycle is:

```text
Application Start
       │
       ▼
Load Configuration
       │
       ▼
MCP Runtime
       │
       ├── Start MCP Server
       │
       ├── Connect MCP Client
       │
       ├── Create Session
       │
       ├── Discover Tools
       │
       └── Register Tools
       │
       ▼
Application Running
       │
       ▼
Agent Executes MCP Tool
       │
       ▼
MCP Client Session
       │
       ▼
MCP Server
       │
       ▼
Tool Result
       │
       ▼
Application Shutdown
       │
       ▼
Close MCP Sessions
```

---

# MCP Runtime Configuration

MCP servers are configured through the application configuration.

The configuration supports:

```json
{
  "provider": {
    "name": "ollama",
    "model": "llama3.1",
    "base_url": "http://localhost:11434",
    "endpoint": "/api/chat"
  },
  "mcp": {
    "servers": [
      {
        "name": "example",
        "command": "example-mcp-server",
        "args": []
      }
    ]
  }
}
```

Each MCP server contains:

```text
Name
Command
Args
```

`Args` allows command-line arguments to be passed to the MCP server without requiring shell parsing.

For example:

```json
{
  "name": "example",
  "command": "example-mcp-server",
  "args": ["--port", "8080"]
}
```

---

# MCP Test Server

The repository contains a small MCP server used for integration testing:

```text
cmd/mcp-test-server/main.go
```

It exposes a test tool:

```text
test_tool
```

The tool returns:

```text
MCP test tool executed successfully
```

The test server uses the MCP SDK's stdio transport.

When launched directly:

```bash
go run ./cmd/mcp-test-server
```

it waits for an MCP client to communicate through stdin/stdout. Therefore, it is expected to produce no normal terminal output while waiting.

---

# Building the MCP Test Server

Build the test server from the project root:

```bash
go build -o cmd/mcp-test-server/mcp-test-server ./cmd/mcp-test-server
```

This creates:

```text
cmd/mcp-test-server/mcp-test-server
```

The generated binary is local build output and should not be committed to Git.

The binary is used by the MCP command-transport tests.

---

# MCP Testing

MCP tests are located in:

```text
internal/mcp/
```

The tests cover:

* MCP client creation
* MCP transport connection
* Command-based MCP connection
* MCP tool discovery
* MCP tool execution
* MCP tool schema handling
* Tool adapter creation
* Tool adapter execution
* MCP tool registration
* Runtime creation
* MCP runtime server connection
* MCP session cleanup

Run the MCP tests with:

```bash
go test ./internal/mcp -v
```

Before running the command-transport test, build the test server:

```bash
go build -o cmd/mcp-test-server/mcp-test-server ./cmd/mcp-test-server
```

Then:

```bash
go test ./internal/mcp -v
```

---

# MCP Runtime Tests

The runtime tests are located in:

```text
internal/mcp/runtime_test.go
```

They verify the MCP runtime lifecycle independently from the CLI.

The runtime tests ensure that the runtime can:

```text
Create Runtime
Connect MCP Servers
Discover MCP Tools
Register Tools
Keep Sessions
Close Sessions
```

This keeps MCP runtime behavior independently testable before it is coupled further with the CLI.

---

# Credential Management

The project includes a credential storage abstraction.

Credential functionality provides:

* Credential storage
* Credential retrieval
* Credential deletion
* Storage abstraction

Credentials are kept separate from Provider logic so credential backends can evolve independently.

---

# Session Management

The project includes a Session abstraction and storage layer.

Current session functionality includes:

* Session creation
* Session IDs
* Session message storage
* In-memory session storage
* File-based session storage
* Set
* Get
* Delete
* Agent session assignment
* Session-aware Agent message handling
* Persistent conversation history
* Default session restoration

A Session currently contains:

```go
type Session struct {
    ID       string
    Messages []providerpkg.Message
}
```

---

# Session Store

The Session Store provides:

```text
Set
Get
Delete
List
```

The project currently provides two implementations:

```text
MemoryStore
FileSessionStore
```

The in-memory store keeps sessions for the lifetime of the process.

The file-based store serializes Session objects as JSON files.

---

# File-Based Session Storage

File-based sessions are stored in:

```text
.sessions/
```

The directory is intentionally excluded from Git.

Each session is stored as a JSON file.

For example:

```text
.sessions/
├── default.json
├── project-a.json
└── project-b.json
```

This allows session history to survive application restarts.

---

# Agent Session Integration

The Agent can be connected to a Session:

```go
err := agent.SetSession(session)
```

When a Session is assigned, Agent message operations use the Session's message history.

The active conversation can be retrieved using:

```go
messages := agent.GetMessages()
```

This allows the Agent and Session to share the same conversation state.

---

# Session Persistence

The application persists the active Session after successful Agent interactions.

The persistence flow is:

```text
User Input
    │
    ▼
Agent
    │
    ▼
Provider / Tools
    │
    ▼
Updated Session
    │
    ▼
FileSessionStore
    │
    ▼
.sessions/<id>.json
```

When the application starts, the `default` session is restored if it already exists.

---

# CLI

The interactive CLI is located in:

```text
cmd/main.go
```

Run the application with:

```bash
go run ./cmd
```

The CLI starts with the default session:

```text
default
```

---

# CLI Commands

| Command               | Description                            |
| --------------------- | -------------------------------------- |
| `help`                | Show available commands                |
| `model`               | Show the currently configured model    |
| `session`             | Show the current session ID            |
| `session create <id>` | Create a new session                   |
| `session list`        | List stored sessions                   |
| `session load <id>`   | Load and activate a stored session     |
| `session delete <id>` | Delete a stored session                |
| `clear`               | Clear the current conversation history |
| `exit`                | Exit the CLI                           |

---

# `help`

Displays available CLI commands.

```text
help
```

---

# `model`

Displays the currently configured model.

```text
model
```

The model can also be overridden using:

```bash
export OLLAMA_MODEL=your-model
```

---

# `session`

Displays the currently active session.

```text
session
```

Example:

```text
Current session: default
```

---

# `session create`

Creates a new session:

```text
session create project-a
```

Example:

```text
Session created: project-a
```

---

# `session list`

Lists stored sessions:

```text
session list
```

Example:

```text
Sessions:
- default
- project-a
- project-b
```

---

# `session load`

Loads an existing session:

```text
session load project-a
```

The Agent is updated to use the loaded Session.

---

# `session delete`

Deletes a stored session:

```text
session delete project-a
```

The CLI prevents deletion of the currently active session.

---

# `clear`

Clears the current conversation history:

```text
clear
```

This affects the active conversation but does not delete the Session itself.

---

# `exit`

Exits the CLI:

```text
exit
```

---

# Configuration

Configuration is implemented in:

```text
config/
```

The configuration contains Provider and MCP settings.

Current structure:

```json
{
  "provider": {
    "name": "ollama",
    "model": "llama3.1",
    "base_url": "http://localhost:11434",
    "endpoint": "/api/chat"
  },
  "mcp": {
    "servers": []
  }
}
```

The example configuration is available at:

```text
config/config.example.json
```

The local configuration:

```text
config/config.json
```

is ignored by Git because it may contain local development settings.

---

# Configuration Validation

The configuration layer validates required Provider fields:

```text
Provider Name
Provider Model
Provider Base URL
Provider Endpoint
```

MCP configuration is optional, allowing existing configurations without MCP servers to remain valid.

---

# Provider

The Provider interface keeps the Agent independent from a particular model provider.

The Provider layer is responsible for:

* Sending chat requests
* Receiving model responses
* Handling tool definitions
* Handling tool calls
* Listing models
* Checking model availability
* Pulling local models where supported

---

# Ollama

Ollama is currently the primary integrated Provider.

The default Ollama URL is:

```text
http://localhost:11434
```

The default chat endpoint is:

```text
/api/chat
```

Start Ollama:

```bash
ollama serve
```

Check available models:

```bash
ollama list
```

Pull a model:

```bash
ollama pull llama3.1
```

Then start Agent Harness:

```bash
go run ./cmd
```

---

# Model Override

The configured model can be overridden using:

```bash
export OLLAMA_MODEL=llama3.1
```

This allows different local models to be tested without modifying the configuration file.

---

# Orchestrator

The Orchestrator coordinates the Agent execution loop.

It handles repeated:

```text
Model Request
      │
      ▼
Model Response
      │
      ├── Final Response
      │
      └── Tool Call
              │
              ▼
        Tool Execution
              │
              ▼
        Tool Result
              │
              ▼
        Model Request
```

The loop continues until:

* The model produces a final response
* The maximum tool-call limit is reached
* An error occurs

The maximum number of tool-call iterations can be configured.

This prevents uncontrolled tool execution loops.

---

# Architecture

```text
                         ┌──────────────────────┐
                         │    Interactive CLI   │
                         └──────────┬───────────┘
                                    │
                                    ▼
                         ┌──────────────────────┐
                         │        Agent         │
                         │                      │
                         │ Conversation History │
                         │ Session Integration  │
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
                                                    │ MCP Runtime │
                                                    └──────┬──────┘
                                                           │
                                                           ▼
                                                    ┌─────────────┐
                                                    │ MCP Client  │
                                                    └──────┬──────┘
                                                           │
                                                           ▼
                                                    ┌─────────────┐
                                                    │ MCP Server  │
                                                    └─────────────┘
```

---

# Session Architecture

```text
                    ┌────────────────────┐
                    │      Session       │
                    │                    │
                    │        ID          │
                    │     Messages       │
                    └─────────┬──────────┘
                              │
                     ┌────────┴────────┐
                     │                 │
                     ▼                 ▼
             ┌──────────────┐   ┌──────────────┐
             │ MemoryStore  │   │ File Store   │
             └──────────────┘   └──────┬───────┘
                                       │
                                       ▼
                                .sessions/*.json
```

The Session layer is separated from the Agent runtime so storage implementations can evolve independently.

---

# Runtime Flows

## Native Tool Execution

```text
User
 │
 ▼
CLI
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
 ├── Normal response ──────────► Agent ──► User
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

---

## MCP Tool Execution

```text
User
 │
 ▼
Agent
 │
 ▼
Provider
 │
 │ Tool Call
 ▼
Tool Registry
 │
 ▼
MCP Tool Adapter
 │
 ▼
MCP Client Session
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
Provider
 │
 ▼
Final Response
```

---

## Session Flow

```text
CLI
 │
 ▼
Session Store
 │
 ├── Create
 ├── Get
 ├── Set
 └── Delete
 │
 ▼
Session
 │
 ├── ID
 └── Messages
 │
 ▼
Agent
 │
 ▼
Provider / Tools
 │
 ▼
Updated Session
 │
 ▼
Session Store
 │
 ▼
Persistent Storage
```

---

# Project Structure

```text
Agent-harness/
│
├── .github/
│   └── workflows/
│       └── go.yml
│
├── cmd/
│   ├── main.go
│   ├── main_test.go
│   └── mcp-test-server/
│       └── main.go
│
├── config/
│   ├── config.go
│   ├── config_test.go
│   ├── config.example.json
│   └── config.json
│
├── filesystem/
│   ├── filesystem.go
│   └── filesystem_test.go
│
├── shell/
│   ├── shell.go
│   └── shell_test.go
│
├── internal/
│   │
│   ├── agent/
│   │   ├── agent.go
│   │   ├── agent_test.go
│   │   ├── history.go
│   │   └── history_test.go
│   │
│   ├── credentials/
│   │   ├── credentials.go
│   │   └── credentials_test.go
│   │
│   ├── mcp/
│   │   ├── adapter.go
│   │   ├── adapter_test.go
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── runtime.go
│   │   ├── runtime_test.go
│   │   ├── tool.go
│   │   └── tool_test.go
│   │
│   ├── orchestrator/
│   │   ├── orchestrator.go
│   │   └── orchestrator_test.go
│   │
│   ├── provider/
│   │   ├── provider.go
│   │   ├── ollama.go
│   │   └── ollama_test.go
│   │
│   ├── session/
│   │   ├── session.go
│   │   ├── session_test.go
│   │   ├── store.go
│   │   ├── store_test.go
│   │   ├── memory_store.go
│   │   ├── memory_store_test.go
│   │   ├── file_store.go
│   │   └── file_store_test.go
│   │
│   └── tools/
│       ├── tool.go
│       ├── tool_test.go
│       ├── registry.go
│       ├── registry_test.go
│       ├── calculator.go
│       ├── calculator_test.go
│       ├── shell.go
│       ├── shell_test.go
│       ├── filesystem.go
│       └── filesystem_test.go
│
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
└── README.md
```

The generated MCP test-server binary:

```text
cmd/mcp-test-server/mcp-test-server
```

is intentionally not included in the repository and should remain ignored by Git.

---

# Testing

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
```

Run MCP tests:

```bash
go test ./internal/mcp -v
```

Run CLI tests:

```bash
go test ./cmd
```

---

# Ollama Integration Testing

The Ollama Orchestrator integration test can be enabled using:

```bash
ORCHESTRATOR_INTEGRATION=1 go test ./...
```

The integration test requires:

* Ollama running locally
* A compatible model installed
* A model supporting the required tool-calling behavior

---

# Formatting

Format Go files with:

```bash
gofmt -w $(find . -name '*.go')
```

Check formatting:

```bash
gofmt -l $(find . -name '*.go')
```

The formatting check should produce no output.

---

# Static Analysis

Run:

```bash
go vet ./...
```

The expected result is no reported issues.

---

# Continuous Integration

GitHub Actions is used to validate the repository automatically.

The workflow performs:

```text
Checkout
   │
   ▼
Setup Go
   │
   ▼
Check Formatting
   │
   ▼
go test ./...
   │
   ▼
go vet ./...
```

The workflow runs for pushes to:

```text
main
feature/**
```

and for pull requests targeting:

```text
main
```

---

# Development Workflow

Create a feature branch:

```bash
git checkout -b feature/<name>
```

Make changes and format the project:

```bash
gofmt -w $(find . -name '*.go')
```

Run tests:

```bash
go test ./...
```

Run static analysis:

```bash
go vet ./...
```

Check the working tree:

```bash
git status
```

Review changes:

```bash
git diff
```

Stage the completed logical change:

```bash
git add .
```

Commit:

```bash
git commit -m "your commit message"
```

Push the feature branch:

```bash
git push -u origin feature/<name>
```

Then open a pull request against `main`.

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

All tools are exposed through the common Tool abstraction.

---

## Session Independence

Session storage is separated from the Agent runtime.

This allows different storage implementations to be introduced without rewriting Agent execution.

Potential future backends include:

```text
In-Memory
File
Database
Remote Storage
```

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
Database-backed Sessions
Session Metadata
Session Export
Session Import
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
| Session-aware Agent runtime           | Implemented |
| Tool abstraction                      | Implemented |
| Tool registry                         | Implemented |
| Calculator tool                       | Implemented |
| Shell tool                            | Implemented |
| Filesystem tool                       | Implemented |
| MCP client                            | Implemented |
| MCP command transport                 | Implemented |
| MCP tool discovery                    | Implemented |
| MCP tool wrapper                      | Implemented |
| MCP tool adapter                      | Implemented |
| MCP tool registration                 | Implemented |
| MCP runtime                           | Implemented |
| MCP session lifecycle                 | Implemented |
| MCP test server                       | Implemented |
| Credential abstraction                | Implemented |
| Credential storage                    | Implemented |
| Session abstraction                   | Implemented |
| Session message storage               | Implemented |
| In-memory session store               | Implemented |
| File-based session store              | Implemented |
| Persistent Agent conversation history | Implemented |
| Automatic session persistence         | Implemented |
| Default session restoration           | Implemented |
| Interactive CLI                       | Implemented |
| CLI model command                     | Implemented |
| CLI session display                   | Implemented |
| CLI conversation clearing             | Implemented |
| CLI session creation                  | Implemented |
| CLI session listing                   | Implemented |
| CLI session loading                   | Implemented |
| CLI session deletion                  | Implemented |
| Unit tests                            | Implemented |
| Integration tests                     | Implemented |
| Go formatting checks                  | Implemented |
| Go vet checks                         | Implemented |
| GitHub Actions CI                     | Implemented |
| Database-backed sessions              | Planned     |
| Additional providers                  | Planned     |
| Expanded MCP functionality            | Planned     |

---

# Roadmap

## Improved Session Management

Future session improvements include:

```text
Database-backed Sessions
Session Metadata
Session Timestamps
Session Renaming
Session Export
Session Import
Improved Persistence Synchronization
```

---

## Improved MCP Support

Future MCP work includes:

* CLI configuration of MCP servers
* Dynamic MCP server registration
* Improved MCP server lifecycle management
* Better MCP error handling
* MCP resource support
* MCP prompts
* MCP sampling where applicable
* Remote MCP transports

The current runtime primarily focuses on local command-based MCP servers.

---

## Additional Providers

The Provider abstraction allows future integrations such as:

```text
OpenAI-compatible APIs
Anthropic
Google
Other Local Model Runtimes
```

Provider implementations can be added without changing the core Agent architecture.

---

## Improved CLI

Planned CLI improvements include:

```text
Provider Management
Tool Listing
MCP Server Management
Configuration Inspection
Advanced Model Management
Persistent Session Selection
Interactive Tool Inspection
```

---

# End-to-End Example

A simplified native tool execution looks like:

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

A simplified MCP tool execution looks like:

```text
User
 │
 ▼
Agent
 │
 ▼
Ollama
 │
 │ Tool Call
 ▼
Tool Registry
 │
 ▼
MCP Adapter
 │
 ▼
MCP Client Session
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

A simplified session-aware flow looks like:

```text
User
 │
 ▼
CLI
 │
 ▼
Active Session
 │
 ├── ID
 └── Messages
 │
 ▼
Agent
 │
 ▼
Provider
 │
 ▼
Tool Execution
 │
 ▼
Updated Session Messages
 │
 ▼
Session Store
 │
 ▼
Persistent Storage
```

---

# Project Goals

The long-term goal of Agent Harness is to provide a modular runtime for building tool-using AI agents while keeping the underlying architecture understandable, testable, and extensible.

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

The project is designed to evolve toward a more complete agent runtime while maintaining clear boundaries between:

```text
Providers
Tools
Sessions
Credentials
MCP
CLI
```

---

# License

This project is licensed under the MIT License.

See the `LICENSE` file for the full license text.
