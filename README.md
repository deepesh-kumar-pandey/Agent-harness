# Agent Harness

Agent Harness is a Go-based runtime for building and running tool-using AI agents. It provides provider abstraction, conversation history, native tool execution, MCP integration, credential management, session storage, and an interactive CLI.

The project is designed with a local-first architecture while keeping the core runtime provider-neutral and extensible.

---

## Current Status

The project currently supports:

- Provider abstraction with Ollama integration
- Conversation history and agent orchestration
- Native tool calling
- Built-in calculator, shell, and filesystem tools
- Tool registry with dynamic registration and removal
- MCP client integration
- MCP tool discovery and adaptation into the native tool system
- Credential storage and retrieval
- In-memory and file-based session storage
- Session-aware Agent runtime
- CLI session creation, listing, loading, and deletion
- Interactive CLI
- CLI model inspection and session display
- Local Ollama model management
- Unit and integration tests
- Go formatting, testing, and static analysis
- GitHub Actions CI

The core runtime is functional. Current development is focused on strengthening session management, persistent conversation history, MCP capabilities, and runtime extensibility.

---

# Features

## Provider Abstraction

The runtime uses provider-neutral interfaces so the Agent does not depend directly on a specific model provider.

Currently integrated:

- Ollama

The provider layer supports:

- Chat requests
- Chat responses
- Conversation messages
- Tool definitions
- Tool calls
- Model listing
- Model availability checks
- Model pulling

Provider-specific behavior is kept behind the Provider abstraction.

---

## Agent Runtime

The Agent coordinates conversation history, provider interaction, and tool execution.

The runtime supports:

- User messages
- Assistant responses
- Tool calls
- Tool results
- Multiple tool calls
- Configurable maximum tool-call iterations
- Provider-neutral execution
- Session-aware conversation history

The Agent can operate with its internal conversation history or with an explicitly assigned Session.

---

## Tool Registry

The Tool Registry provides a central system for managing executable tools.

Built-in tools currently include:

- Calculator
- Shell
- Filesystem

The registry supports:

- Registering tools
- Looking up tools
- Checking whether a tool exists
- Listing tools
- Removing tools
- Generating tool schemas

The registry provides the common execution layer used by both native and adapted MCP tools.

---

## MCP Integration

The project includes support for the Model Context Protocol (MCP).

The MCP implementation currently provides:

- MCP client creation
- MCP server connections
- MCP session management
- MCP tool discovery
- MCP tool metadata
- MCP tool execution
- Conversion of MCP tools into native Agent Harness tools
- Tool adapter integration with the native Tool interface

MCP tools can participate in the same tool execution flow as built-in tools.

---

## Credential Management

The project includes a credential storage abstraction for managing provider or service credentials.

Credential functionality provides:

- Credential storage
- Credential retrieval
- Credential deletion
- A storage abstraction that can be extended later

Credentials are kept separate from provider logic so credential backends can evolve independently.

---

## Session Management

The project includes a session abstraction and session storage layer.

Current session functionality includes:

- Session creation
- Session IDs
- Session message storage
- In-memory session storage
- File-based session storage
- Set
- Get
- Delete
- Agent session assignment
- Session-aware Agent message handling

A Session currently contains:

- Session ID
- Conversation messages

The Agent can be connected to a Session using the session-aware runtime.

The current file-based store can serialize and restore the Session object, including its stored messages.

Automatic persistence after every Agent interaction and automatic session restoration when the application starts are future improvements.

---

# Architecture

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
                                                    │ MCP Client  │
                                                    └─────────────┘
```

## Session Layer

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
             In-Memory Store      File Store
```

The Session layer is intentionally separated from the Agent runtime so storage and persistence can evolve independently.

The Agent can use a Session as its conversation state while the storage implementation determines how that Session is stored.

---

# Runtime Flow

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

## MCP Tool Execution

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
```

---

# Project Structure

```text
Agent-harness/
│
├── cmd/
│   ├── main.go
│   └── main_test.go
│
├── internal/
│   │
│   ├── agent/
│   │   ├── agent.go
│   │   ├── history.go
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
│   │   ├── memory_store.go
│   │   ├── file_store.go
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

- Model
- Provider
- Base URL
- Endpoint
- Other provider-specific settings

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

A Session represents a logical interaction context.

The current Session structure is:

```go
type Session struct {
	ID       string
	Messages []providerpkg.Message
}
```

A session can be created with:

```go
session := NewSession("default")
```

A Session contains both its identifier and its conversation messages.

### Session Store

The Session Store provides:

```text
Set
Get
Delete
```

The project currently has two storage implementations:

- In-memory session store
- File-based session store

The in-memory store keeps sessions for the lifetime of the process.

The file-based store serializes Session objects as JSON files and can restore stored session data.

### Agent Session Integration

The Agent can be connected to a Session:

```go
err := agent.SetSession(session)
```

When a Session is assigned, Agent message operations use the Session's message history.

The Agent can retrieve the active conversation through:

```go
messages := agent.GetMessages()
```

This allows session state and Agent execution to share the same conversation history.

### Current Persistence Limitation

The storage layer can persist Session objects, but the full application lifecycle is not yet automatically persistent.

The current implementation does not automatically:

- Persist the active Session after every Agent interaction
- Automatically restore the previous Session when the application starts
- Automatically migrate the Agent's active Session to disk
- Automatically synchronize every conversation update with the file store

These are future extensions.

---

## Provider

The Provider interface keeps the Agent independent from a particular model provider.

The provider layer is responsible for:

- Sending chat requests
- Receiving model responses
- Handling tool definitions
- Handling tool calls
- Listing models
- Checking model availability
- Pulling local models where supported

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

The Agent is responsible for coordinating the interaction between the user, model, tools, and session state.

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
            │
            ▼
      Final Response
```

The Agent is provider-neutral and operates through the Provider and Tool abstractions.

When a Session is assigned, the Session becomes the source of conversation messages.

---

## Orchestrator

The Orchestrator coordinates the Agent execution loop.

It handles repeated model/tool interactions until:

- The model produces a final response
- The maximum tool-call limit is reached
- An error occurs

The maximum number of tool-call iterations can be configured.

This prevents an Agent from entering an uncontrolled tool execution loop.

---

# Tool System

## Tool Interface

Tools expose a common interface containing:

- Name
- Description
- Execute
- Schema

This allows different tool implementations to be treated uniformly.

The same abstraction can represent native tools and adapted MCP tools.

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

This provides a central mechanism for native tools and externally discovered tools.

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

The Shell tool allows command execution through the tool system.

It is intended for controlled local Agent workflows.

---

### Filesystem

The Filesystem tool provides filesystem-related operations through the common Tool interface.

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

- Connecting to an MCP server
- Managing the MCP session
- Discovering tools
- Returning discovered tools

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

Discovered MCP tools can be adapted and registered into the existing Tool Registry.

This means the Agent can treat:

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

| Command | Description |
| --- | --- |
| `help` | Show available commands |
| `model` | Show the currently configured model |
| `session` | Show the current session ID |
| `session create <id>` | Create a new session |
| `session list` | List stored sessions |
| `session load <id>` | Load and activate a stored session |
| `session delete <id>` | Delete a stored session |
| `clear` | Clear the current conversation history |
| `exit` | Exit the CLI |

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

---

## `session create`

Creates a new session and stores it in the active Session Store.

```text
session create project-a
```

Example:

```text
Session created: project-a
```

---

## `session list`

Lists the sessions currently stored in the active Session Store.

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

## `session load`

Loads an existing session and makes it the active session.

```text
session load project-a
```

Example:

```text
Session loaded: project-a
```

The Agent is updated to use the loaded session.

---

## `session delete`

Deletes a stored session.

```text
session delete project-a
```

The CLI prevents deletion of the currently active session.

---

## `clear`

Clears the current Agent conversation history.

```text
clear
```

This affects the active conversation history.

It does not delete the Session itself.

---

## `exit`

Exits the CLI.

```text
exit
```

---

# Ollama Setup

Install and start Ollama before running Agent Harness.

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

The Ollama Orchestrator integration test can be enabled with:

```bash
ORCHESTRATOR_INTEGRATION=1 go test ./...
```

The integration test requires:

- Ollama running locally
- A compatible model available
- The configured model supporting the required tool-calling behavior

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

A simplified native tool execution looks like this:

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

A simplified MCP tool execution looks like this:

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

A simplified session-aware flow looks like this:

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
```

---

# Architectural Principles

## Provider Neutrality

The Agent should not depend directly on Ollama-specific APIs.

Provider-specific behavior belongs behind the Provider interface.

---

## Tool Agnosticism

The Agent should not need to know whether a tool is:

- Built into the project
- Provided by an MCP server
- Added by another future integration

All tools should be exposed through the common Tool abstraction.

---

## Session Independence

Session storage is separated from the Agent runtime.

This allows different storage implementations to be introduced without requiring major changes to Agent execution.

Potential session backends include:

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

Persistent Conversation Storage

Database-backed Sessions

```

---

# Implementation Status

| Component | Status |
| --- | --- |
| Go project structure | Implemented |
| Provider abstraction | Implemented |
| Ollama provider | Implemented |
| Conversation history | Implemented |
| Agent runtime | Implemented |
| Session-aware Agent runtime | Implemented |
| Tool abstraction | Implemented |
| Tool registry | Implemented |
| Calculator tool | Implemented |
| Shell tool | Implemented |
| Filesystem tool | Implemented |
| MCP client | Implemented |
| MCP tool discovery | Implemented |
| MCP tool adapter | Implemented |
| MCP tool registration foundation | Implemented |
| Credential abstraction | Implemented |
| Credential storage | Implemented |
| Session abstraction | Implemented |
| Session message storage | Implemented |
| In-memory session store | Implemented |
| File-based session store | Implemented |
| Interactive CLI | Implemented |
| CLI model command | Implemented |
| CLI session display | Implemented |
| CLI conversation clearing | Implemented |
| CLI session creation | Implemented |
| CLI session listing | Implemented |
| CLI session loading | Implemented |
| CLI session deletion | Implemented |
| Persistent Agent conversation history | Partial |
| Automatic session persistence | Planned |
| Automatic session restore | Planned |
| Database-backed sessions | Planned |
| Additional providers | Planned |
| Expanded MCP functionality | Planned |

---

# Roadmap

## Persistent Conversation History

The next stage of session development is connecting Agent conversation updates with the session storage layer.

Future flow:

```text
User Interaction
      │
      ▼
Agent
      │
      ▼
Session Messages
      │
      ▼
Session Store
      │
      ▼
Persistent Storage
```

This will allow conversations to survive application restarts automatically.

---

## Improved Session Management

Future session improvements include:

```text
Automatic Session Persistence

Automatic Session Restore

Database-backed Sessions

Session Metadata

Session Timestamps

Session Renaming

Session Export

Session Import
```

---

## Improved MCP Support

Future MCP work includes:

- CLI configuration of MCP servers
- MCP server lifecycle management
- Dynamic MCP server registration
- Better error handling
- More complete MCP resource support
- MCP prompts
- MCP sampling where applicable

---

## Additional Providers

The provider abstraction allows future integrations such as:

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
Session Management

Provider Management

Tool Listing

MCP Server Management

Configuration Inspection

Model Management

Persistent Session Selection

Interactive Tool Inspection
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

The project is designed to evolve toward a more complete agent runtime while maintaining clear boundaries between providers, tools, sessions, credentials, MCP, and the CLI.

---

# License

This project is licensed under the MIT License.

See the [LICENSE](LICENSE) file for the full license text.