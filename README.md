# Agent Harness

Agent Harness is a Go application and runtime for running a conversational agent that can call local tools and tools exposed by MCP servers. The interactive CLI currently supports Ollama, OpenAI, Anthropic, Google Gemini, and Mistral, keeps conversation history in file-backed sessions, routes provider tool calls through a shared tool registry, and evaluates tool actions through the Jev action-evaluation layer.

It provides:

* Provider abstraction
* Agent orchestration
* Conversation history
* Native tool execution
* Deterministic action evaluation through Jev
* MCP integration
* MCP server lifecycle management
* MCP server CLI management
* Credential storage
* Session management and persistence
* Interactive CLI
* Local Ollama model management
* Automated testing and CI

The core is provider-neutral and modular. Ollama provides the local-first path, while cloud providers use the same Provider abstraction and credential system.

---

# Current Status

The project currently supports:

* Provider abstraction with Ollama, OpenAI, Anthropic, Google Gemini, and Mistral
* Provider factory
* Provider API-key requirement detection
* Conversation history
* Agent orchestration
* Native tool calling
* Deterministic tool-action evaluation through Jev
* Allow, Confirm, and Deny action decisions
* Basic action validation policy
* Shell action confirmation policy
* Configurable Jev evaluator injection
* Built-in calculator, shell, and filesystem tools
* Dynamic tool registry
* MCP client integration
* MCP server connections through command transport
* MCP tool discovery
* MCP tool adaptation into the native tool system
* MCP tool registration into the Tool Registry
* MCP runtime/session lifecycle management
* MCP server configuration through the CLI
* MCP server listing through the CLI
* MCP server addition through the CLI
* MCP server removal through the CLI
* Dynamic MCP server registration
* Dynamic MCP server removal
* Persistent MCP server configuration
* Credential storage and retrieval
* In-memory session storage
* File-based session storage
* Database-backed session storage
* Atomic file-based session persistence
* Transactional database session persistence
* Session timestamps
* Session metadata
* Session renaming
* Session export and import
* Session-aware Agent runtime
* Persistent Agent conversation history
* Default session restoration
* CLI session creation and automatic activation
* CLI session listing
* CLI session loading
* CLI session deletion
* Interactive CLI
* CLI model display and startup model selection for Ollama
* CLI display of the active session ID
* Local Ollama model management
* Unit and integration tests
* Go formatting and static analysis
* GitHub Actions CI

The core runtime is functional.

Current development is focused on:

* Improving MCP capabilities
* Improving MCP lifecycle and error handling
* Improving CLI functionality
* Improving provider behavior and compatibility
* Increasing runtime extensibility
* Extending deterministic action-evaluation policies

The interactive CLI currently uses the file session store in `.sessions/`. The in-memory and SQLite stores are available as package implementations but are not selected by the CLI.

---

# Features

## Provider Abstraction

The runtime uses provider-neutral interfaces so the Agent does not depend directly on a specific model provider.

Currently integrated:

* Ollama
* OpenAI
* Anthropic
* Google Gemini
* Mistral

The common `Provider` interface defines `Chat(request)`, which sends a provider-neutral request and returns response text and any normalized tool calls.

Provider-specific request and response formats remain behind the Provider abstraction.

### Provider Factory

The provider factory creates the appropriate provider implementation:

```text
NewProvider
```

`NewProvider(name, baseURL, apiKey string) (Provider, error)` creates a provider based on the supplied provider name.

Currently supported providers are:

```text
ollama
openai
gemini
anthropic
mistral
```

Example:

```text
NewProvider("ollama", "http://localhost:11434", "")
```

This creates an Ollama provider.

For a provider requiring credentials:

```text
NewProvider("openai", "https://api.openai.com/v1", apiKey)
```

This creates an OpenAI provider using the supplied API key.

The factory keeps provider construction separate from the Agent and Orchestrator.

### Provider API

The common Provider interface exposes:

```text
Chat
```

`Chat(request ChatRequest) (ChatResponse, error)` sends the model, messages, and optional tool definitions to the configured provider. It returns response text and normalized tool calls, or an error for invalid requests, transport/status failures, response-decoding failures, or provider-specific response errors.

Conceptually:

```text
Provider

   │

   ▼

Chat(request)

   │

   ▼

Provider API

   │

   ▼

Normalized ChatResponse
```

Local model providers can additionally implement:

```text
ListModels

PullModel
```

Only Ollama implements `LocalModelManager` in the current code.

`RequiresAPIKey(name string) bool` reports whether a provider requires an API key before CLI initialization.

Currently:

```text
ollama    → false
openai    → true
gemini    → true
anthropic → true
mistral   → true
```

Unknown provider names return `false`; `NewProvider` is responsible for rejecting unsupported providers.

---

# Agent Runtime

The Agent manages conversation state and provides access to the tool system.

The Orchestrator coordinates communication between the Agent and the Provider.

Jev evaluates tool actions before execution on the Orchestrator's `AssignTool` path.

The runtime supports:

* User messages
* Assistant responses
* Tool calls
* Tool results
* Multiple tool calls
* Configurable maximum number of tool calls
* Provider-neutral execution
* Session-aware conversation history
* Deterministic action evaluation

The Agent can operate using:

1. Its internal conversation history
2. An explicitly assigned Session

When a Session is assigned, the Session becomes the source of conversation history.

## Agent Operations

`NewAgent(registry *tools.ToolRegistry) *Agent` creates an Agent backed by that registry and an empty in-memory `Conversation`.

### `SetSession`

```text
SetSession(session *session.Session) error
```

Assigns a non-nil Session and makes its `Messages` slice the active conversation history.

It returns an error for a nil Session.

The Session is assigned directly rather than copied.

Example:

```go
err := agent.SetSession(session)
```

After this operation, Agent conversation messages are associated with the assigned Session.

### `AddMessage`

```text
AddMessage(message provider.Message) error
```

Appends the message to the active Session or, when no Session is assigned, to the Agent's in-memory Conversation.

It rejects a message with an empty role.

Example:

```go
err := agent.AddMessage(message)
```

### `GetMessages`

```text
GetMessages() []provider.Message
```

Returns the active Session's `Messages` slice or the in-memory Conversation's message slice.

The returned slice is not copied.

Example:

```go
messages := agent.GetMessages()
```

### Tool Access

* `GetToolSchemas() ([]map[string]any, error)` obtains all tool schemas from the registry and propagates registry/schema errors.
* `ExecuteTool(name string, args map[string]any) (any, error)` looks up `name`, executes the tool with `args`, and returns lookup or tool errors.
* `Run(name string, args map[string]any) (any, error)` delegates directly to `ExecuteTool`.

The conversational provider/tool loop is implemented by:

```text
DefaultOrchestrator.RunAgent
```

---

# Conversation History

The `Conversation` type backs an Agent that has no assigned Session.

### `NewConversation`

```text
NewConversation() *Conversation
```

Creates an empty conversation message list.

### `AddMessage`

```text
(*Conversation).AddMessage(message provider.Message) error
```

Rejects a message with an empty role.

Otherwise, it appends the message to the conversation.

### `GetMessages`

```text
(*Conversation).GetMessages() []provider.Message
```

Returns the underlying conversation message slice without copying it.

---

# Tool System

## Tool Interface

Tools implement the common interface:

```text
Name() string

Description() string

Execute(args map[string]any) (any, error)

Schema() map[string]any
```

This allows different tool implementations to be treated uniformly.

The same abstraction is used for:

* Built-in tools
* MCP tools
* Future external tools

---

# Tool Operations

## `Name`

```text
Name() string
```

Returns the tool name used by the registry and provider tool-call routing.

Example:

```text
tool.Name()
```

## `Description`

```text
Description() string
```

Returns a human-readable description supplied to the model.

Example:

```text
tool.Description()
```

## `Execute`

```text
Execute(args map[string]any) (any, error)
```

Executes the tool using the provided argument map and returns its result or an error.

Example:

```text
tool.Execute(arguments)
```

The exact arguments depend on the Tool implementation.

## `Schema`

```text
Schema() map[string]any
```

Returns the parameter schema supplied to the model.

Example:

```text
tool.Schema()
```

---

# Tool Registry

The Tool Registry is implemented in:

```text
internal/tools/ToolRegistry.go
```

`NewToolRegistry() *ToolRegistry` creates the registry and registers the built-in:

```text
calculator

shell

filesystem
```

tools.

It supports:

```text
Register

Get

Has

List

Remove

Schemas
```

## `Register`

```text
Register(name string, tool Tool) (string, Tool)
```

Adds or replaces a tool under `name`, prints a registration message, and returns the supplied name and tool.

Example:

```text
registry.Register("calculator", calculator)
```

## `Get`

```text
Get(name string) (Tool, error)
```

Retrieves a registered tool by name.

It returns an error when the tool is not registered.

Example:

```text
tool, err := registry.Get("calculator")
```

## `Has`

```text
Has(name string) bool
```

Reports whether a tool with the specified name is registered.

Example:

```text
exists := registry.Has("calculator")
```

## `List`

```text
List() []string
```

Returns the names of all registered tools.

The order is unspecified.

Example:

```text
names := registry.List()
```

## `Remove`

```text
Remove(name string) error
```

Removes a registered tool by name.

It returns an error if no such tool exists.

Example:

```text
registry.Remove("calculator")
```

## `Schemas`

```text
Schemas() ([]map[string]any, error)
```

Returns the schemas of registered tools so they can be supplied to a model provider.

Example:

```text
schemas, err := registry.Schemas()
```

It returns each tool's:

```text
Name

Description

Parameter Schema
```

It returns an error if the registry is nil or a registered tool or its schema is nil.

The registry is shared by built-in and adapted MCP tools.

---

# Jev Action Evaluation

Jev is the deterministic action-evaluation layer in:

```text
internal/jev/
```

Jev evaluates tool actions before they are executed through the Orchestrator's `AssignTool` path.

It does not replace the Tool interface, Provider interface, or Tool Registry.

The evaluation result is one of:

```text
Allow

Confirm

Deny
```

The current default evaluator applies policies with the following priority:

```text
Deny > Confirm > Allow
```

Conceptually:

```text
Tool Call

    │

    ▼

Orchestrator

    │

    ▼

Jev Evaluator

    │

    ├── Allow ───────► Tool Execution
    │
    ├── Confirm ─────► Confirmation Required Error
    │
    └── Deny ────────► Denied Error
```

## Action Model

Implemented in:

```text
internal/jev/action.go
```

The `Action` type represents an operation that is being evaluated:

```text
Action

    Tool
    Args
```

It keeps Jev independent from the Orchestrator's `ToolCall` type.

The Orchestrator maps:

```text
orchestrator.ToolCall
```

to:

```text
jev.Action
```

before evaluation.

## Decision Model

Implemented in:

```text
internal/jev/decision.go
```

Jev defines:

```text
Decision
```

with:

```text
Allow

Confirm

Deny
```

The evaluation result is:

```text
DecisionResult
```

which contains:

```text
Decision

Reason
```

The reason describes why the policy produced the decision.

## `Evaluator`

Implemented in:

```text
internal/jev/evaluator.go
```

The evaluator abstraction is:

```text
Evaluate(action Action) DecisionResult
```

It allows the Orchestrator to depend on action evaluation without depending on a specific evaluator implementation.

## `Policy`

Implemented in:

```text
internal/jev/policy.go
```

The policy abstraction is:

```text
Evaluate(action Action) DecisionResult
```

Policies are individual action-evaluation rules that can be composed by an evaluator.

---

# Jev Policies

## Basic Policy

Implemented in:

```text
internal/jev/basic_policy.go
```

`BasicPolicy` performs the baseline action validation.

Its current behavior is:

```text
Empty Tool Name
      │
      ▼
    Deny

Valid Tool Name
      │
      ▼
    Allow
```

An action without a tool name is denied.

Other actions are allowed by the basic policy.

## Shell Policy

Implemented in:

```text
internal/jev/shell_policy.go
```

`ShellPolicy` specifically evaluates shell actions.

Its current behavior is:

```text
Non-shell action
      │
      ▼
    Allow

Shell action
      │
      ▼
   Confirm
```

Shell actions therefore require confirmation rather than being executed directly through the Jev-protected `AssignTool` path.

## Default Evaluator

Implemented in:

```text
internal/jev/default_evaluator.go
```

`NewDefaultEvaluator() *DefaultEvaluator` creates the default evaluator with:

```text
BasicPolicy

ShellPolicy
```

The evaluator combines policy results using:

```text
Deny > Confirm > Allow
```

This means:

* Any Deny decision immediately denies the action.
* Otherwise, a Confirm decision is preserved.
* If no policy requires confirmation or denial, the action is allowed.

The default evaluator provides the current deterministic action-evaluation behavior used by the Orchestrator.

---

# Jev Orchestrator Integration

Jev is integrated into:

```text
internal/orchestrator/orchestrator.go
```

The Orchestrator owns an:

```text
jev.Evaluator
```

and defaults to:

```text
jev.NewDefaultEvaluator()
```

## `WithEvaluator`

```text
WithEvaluator(evaluator jev.Evaluator) OrchestratorOption
```

Replaces the default action evaluator when a non-nil evaluator is supplied.

This allows callers and tests to provide a custom evaluator without changing the Orchestrator interface.

Example:

```text
NewOrchestrator(
    agentClient,
    providerClient,
    WithEvaluator(customEvaluator),
)
```

## `AssignTool`

```text
AssignTool(toolCall ToolCall) (any, error)
```

Executes the named tool using the supplied arguments.

Before execution, the Orchestrator converts the `ToolCall` into a Jev `Action` and evaluates it.

The resulting decision is handled as follows:

```text
Allow
  │
  ▼
Execute Tool

Confirm
  │
  ▼
Return Confirmation Error

Deny
  │
  ▼
Return Denied Error
```

Execution errors are wrapped with the tool name.

Jev is currently integrated into `AssignTool`. The direct `Run(name, args)` method remains a separate direct tool-execution path.

---

# Built-in Tools

## Calculator

The calculator implementation is located in:

```text
internal/tools/tools.go
```

The built-in `calculator` tool takes an `operation` and a `numbers` array.

It supports:

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

### Calculator Operations

Each calculator operation:

* Validates its arguments
* Performs the requested arithmetic operation
* Returns the calculated result

Division and modulus operations reject zero divisors.

---

## Shell

The built-in Shell tool is implemented in:

```text
internal/tools/shell_tool.go
```

The underlying shell helper is implemented in:

```text
shell/shell.go
```

The Shell tool launches an installed local executable with the supplied argument list.

It does not invoke a shell to parse a command string.

### Shell Execution

```text
Execute
```

runs the requested executable and returns its output or an execution error.

Example:

```text
Execute({"command": "ls", "args": ["-la"]})
```

The shell implementation checks that the executable is available and returns combined command output or an error.

Commands run with the permissions of the current process.

---

## Filesystem

The built-in Filesystem tool is implemented in:

```text
internal/tools/filesystem_tool.go
```

The underlying filesystem helper is implemented in:

```text
filesystem/filesystem.go
```

The Agent tool operates on local paths through the common Tool interface.

Its `Execute(args)` operation supports:

* `read`: returns a file's contents.
* `write`: writes supplied `content` to a path.
* `list`: returns directory entry names.
* `exists`: reports whether a path exists.
* `delete`: removes a path.

Each request supplies:

```text
operation

path
```

`write` also requires:

```text
content
```

The top-level `filesystem.FileSystem` helper exposes:

```text
Read

Write

List

Search

Delete
```

and operates directly on filesystem paths.

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
* MCP server CLI configuration

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

## `NewClient`

```text
NewClient() *Client
```

Creates an MCP client identified as:

```text
agent-harness
```

with the current application version.

Example:

```go
client := mcp.NewClient()
```

## `Connect`

```text
Connect(ctx context.Context, transport mcpsdk.Transport) (*mcpsdk.ClientSession, error)
```

Establishes an MCP connection over the supplied transport and returns its client session or a connection error.

Example:

```go
session, err := client.Connect(ctx, transport)
```

The returned ClientSession is used for subsequent MCP operations.

## `ConnectCommand`

```text
ConnectCommand(ctx context.Context, command string, args []string) (*mcpsdk.ClientSession, error)
```

Starts `command` with `args` using `exec.CommandContext`, then connects with the MCP command transport.

It returns the MCP client session or an error.

Example:

```go
session, err := client.ConnectCommand(
    ctx,
    "./mcp-server",
    []string{"--port", "8080"},
)
```

## `ListTools`

```text
ListTools(ctx context.Context, session *mcpsdk.ClientSession) ([]*mcpsdk.Tool, error)
```

Requests the connected server's tool list and returns its tools or the SDK error.

Example:

```go
tools, err := client.ListTools(ctx, session)
```

---

# MCP Command Transport

The current runtime supports local MCP servers through command-based transport.

The client creates the command using:

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
internal/mcp/tools.go
```

It provides:

* Tool name
* Tool description
* Tool schema
* Tool execution

The MCP tool uses the MCP client session to call the remote MCP server.

## `NewTool`

```text
NewTool(tool mcpsdk.Tool) *Tool
```

Wraps an MCP SDK tool in the Agent Harness MCP Tool abstraction.

Example:

```text
tool := mcp.NewTool(serverTool)
```

## Tool Metadata

The MCP wrapper exposes:

```text
Name

Description

Schema
```

### `Name`

```text
Name() string
```

Returns the remote MCP tool name.

### `Description`

```text
Description() string
```

Returns the remote MCP tool description.

### `Schema`

```text
Schema() map[string]any
```

Converts the remote input schema into the native map representation.

It returns nil when the remote schema is nil or conversion fails.

## `Execute`

```text
Execute(ctx context.Context, session *mcpsdk.ClientSession, args map[string]any) (*mcpsdk.CallToolResult, error)
```

Calls the named remote tool through the supplied MCP session.

Example:

```text
result, err := tool.Execute(ctx, session, arguments)
```

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

### `NewToolAdapter`

```text
NewToolAdapter(
    ctx context.Context,
    tool *Tool,
    session *mcpsdk.ClientSession,
) *ToolAdapter
```

Binds the MCP tool wrapper to the context and client session used for calls.

### `Name`

```text
Name() string
```

Delegates to the MCP Tool wrapper.

### `Description`

```text
Description() string
```

Delegates to the MCP Tool wrapper.

### `Schema`

```text
Schema() map[string]any
```

Delegates to the MCP Tool wrapper.

### `Execute`

```text
Execute(args map[string]any) (any, error)
```

Executes the MCP tool through the stored context and client session.

Returned text content is concatenated into the native Tool result.

It propagates MCP call errors and converts an MCP `IsError` result into an error.

Non-text MCP result content is not included in the returned string.

---

# MCP Tool Registration

The MCP package provides tool registration functionality that:

1. Lists tools from an MCP server.
2. Wraps each MCP tool.
3. Creates a native Tool Adapter.
4. Registers the adapter in the Tool Registry.

The registration operation is:

```text
RegisterTools(
    ctx context.Context,
    session *mcpsdk.ClientSession,
    registry *toolspkg.ToolRegistry,
) ([]toolspkg.Tool, error)
```

It discovers available MCP tools, creates wrappers and adapters, registers them under their tool names, and returns the registered tools.

A discovery error is returned.

Registry registration itself has no error return.

Example:

```text
RegisterTools(ctx, session, registry)
```

Conceptually:

```text
MCP Session

     │

     ▼

ListTools

     │

     ▼

Create MCP Tools

     │

     ▼

Create Adapters

     │

     ▼

Registry.Register
```

---

# MCP Runtime

The MCP runtime is implemented in:

```text
internal/mcp/runtime.go
```

The runtime manages MCP server connections during the lifetime of the application.

The runtime:

* Connects to configured MCP servers
* Dynamically connects MCP servers
* Dynamically disconnects MCP servers
* Starts command-based MCP server processes
* Creates MCP client sessions
* Discovers MCP tools
* Registers MCP tools into the native Tool Registry
* Keeps MCP sessions alive while the application runs
* Closes MCP sessions during shutdown

The runtime keeps client sessions alive because MCP Tool Adapters use those sessions when executing tools.

## `NewRuntime`

```text
NewRuntime() *Runtime
```

Creates an MCP runtime with empty active-session and registered-tool maps.

Example:

```go
runtime := mcp.NewRuntime()
```

## `ConnectServer`

```text
ConnectServer(
    ctx context.Context,
    name string,
    command string,
    args []string,
    registry *toolspkg.ToolRegistry,
) error
```

Connects to a command-launched MCP server, creates its client session, discovers its tools, and registers those tools into `registry`.

It rejects duplicate active names.

If tool registration fails, the newly created session is closed.

Example:

```go
err := runtime.ConnectServer(
    ctx,
    "test-server",
    "./test-server",
    []string{},
    registry,
)
```

The server becomes part of the active runtime after successful connection and registration.

## `DisconnectServer`

```text
DisconnectServer(
    name string,
    registry *toolspkg.ToolRegistry,
) error
```

Disconnects an active MCP server session.

Example:

```go
err := runtime.DisconnectServer("test-server", registry)
```

The server's registered tools are removed when the registry still maps those names to the same adapter instances.

The MCP session is then closed.

It returns an error for an unknown active server or a session close failure.

## `Close`

```text
Close(registry *toolspkg.ToolRegistry) error
```

Closes active MCP sessions and releases runtime resources during application shutdown.

Example:

```go
defer runtime.Close(registry)
```

`Close` removes runtime-owned registered tools, attempts to close all active sessions, clears runtime maps, and returns the first close error, if any.

---

# MCP Runtime Lifecycle

The runtime supports startup connections and dynamic server lifecycle operations.

## Application Startup

```text
Application Start

       │

       ▼

Load Configuration

       │

       ▼

MCP Runtime

       │

       ├── Connect Configured Server

       ├── Create Client Session

       ├── Discover Tools

       └── Register Tools

       │

       ▼

Application Running
```

## Dynamic Server Addition

```text
mcp add

       │

       ▼

Connect MCP Server

       │

       ▼

Create Client Session

       │

       ▼

Discover Tools

       │

       ▼

Register Tools

       │

       ▼

Update Configuration

       │

       ▼

Save Configuration

       │

       ▼

Server Available Immediately
```

## Dynamic Server Removal

```text
mcp remove

       │

       ▼

Find Configured Server

       │

       ▼

Disconnect Active Session

       │

       ▼

Remove Server From Configuration

       │

       ▼

Save Configuration

       │

       ▼

Server No Longer Active
```

## Application Shutdown

```text
Application Shutdown

       │

       ▼

MCP Runtime

       │

       ▼

Close Active Sessions

       │

       ▼

MCP Server Processes Exit
```

---

# MCP Runtime Configuration

MCP servers are configured through the application configuration.

Example:

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

---

# MCP CLI Server Management

The interactive CLI supports:

```text
mcp list

mcp add <name> <command> [args...]

mcp remove <name>
```

The MCP CLI manages both:

* Persistent MCP server configuration
* Active MCP runtime connections

Configured servers are loaded and connected when the application starts.

Servers added through `mcp add` are connected immediately.

Servers removed through `mcp remove` are disconnected immediately.

---

## `mcp`

Displays MCP CLI operations.

```text
mcp
```

Example:

```text
Usage: mcp list | mcp add <name> <command> [args...] | mcp remove <name>
```

---

## `mcp list`

Lists configured MCP servers.

```text
mcp list
```

Example:

```text
MCP servers:

- test-server: ./test-server --port 8080

- filesystem: npx -y @modelcontextprotocol/server-filesystem /tmp
```

If no MCP servers are configured:

```text
No MCP servers configured.
```

---

## `mcp add`

Adds and connects a new MCP server.

Syntax:

```text
mcp add <name> <command> [args...]
```

Example:

```text
mcp add test-server ./test-server --port 8080
```

The command:

1. Checks for a duplicate name.
2. Finds the configuration file path.
3. Connects to the MCP server.
4. Creates a client session.
5. Discovers the server's tools.
6. Registers those tools in the Tool Registry.
7. Adds the server to the loaded configuration.
8. Saves the configuration.

If configuration saving fails, the CLI restores the in-memory configuration and disconnects the server.

A successful command makes the server available immediately.

---

## `mcp remove`

Removes and disconnects an MCP server.

Syntax:

```text
mcp remove <name>
```

Example:

```text
mcp remove test-server
```

The command:

1. Searches configured MCP servers.
2. Finds the matching server.
3. Disconnects the active MCP runtime session.
4. Removes the server from configuration.
5. Saves the updated configuration.

If configuration saving fails, the CLI attempts to restore the previous configuration and reconnect the server.

---

# MCP Test Server

The repository contains a small MCP server used for integration and runtime testing:

```text
cmd/mcp-test-server/main.go
```

It exposes:

```text
test_tool

error_tool
```

`test_tool` returns:

```text
MCP test tool executed successfully
```

`error_tool` deliberately returns an error for adapter and runtime error-path testing.

The test server uses the MCP SDK's stdio transport.

When launched directly:

```bash
go run ./cmd/mcp-test-server
```

it waits for an MCP client through stdin/stdout and therefore does not produce normal terminal output while waiting.

---

# Building the MCP Test Server

Build it from the project root:

```bash
go build -o cmd/mcp-test-server/mcp-test-server ./cmd/mcp-test-server
```

This creates:

```text
cmd/mcp-test-server/mcp-test-server
```

The generated binary is local build output and should not be committed to Git.

---

# MCP Testing

MCP tests are located in:

```text
internal/mcp/
```

They cover:

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
* MCP runtime server disconnection

Run:

```bash
go test ./internal/mcp -v
```

Before command-transport tests, build the test server:

```bash
go build -o cmd/mcp-test-server/mcp-test-server ./cmd/mcp-test-server
```

Then:

```bash
go test ./internal/mcp -v
```

---

# Credential Management

The credential abstraction is implemented in:

```text
internal/credentials/store.go
```

Credential functionality provides:

* Credential storage
* Credential retrieval
* Credential deletion
* File-backed credential persistence

Credentials are kept separate from Provider logic.

The main operations are:

```text
Set

Get

Delete
```

## `NewFileStore`

```text
NewFileStore() (*FileStore, error)
```

Creates the file-backed credential store.

It resolves the current user's home directory and uses:

```text
~/.agent-harness/credentials.json
```

It returns an error if the home directory cannot be resolved.

## `Set`

```text
Set(provider string, key string) error
```

Stores or updates the credential associated with a provider.

It rejects an empty provider or key.

The parent directory is created with mode `0700`.

The credential file uses mode `0600`.

Example:

```text
credentials.Set("openai", apiKey)
```

## `Get`

```text
Get(provider string) (string, error)
```

Retrieves the stored credential for the specified provider.

It rejects an empty provider and returns `ErrCredentialNotFound` when the file or provider entry is missing.

File and JSON decoding errors are returned.

Example:

```text
key, err := credentials.Get("openai")
```

## `Delete`

```text
Delete(provider string) error
```

Removes the stored credential associated with the provider.

It returns errors for empty or unknown providers and file/JSON failures.

Example:

```text
credentials.Delete("openai")
```

---

# Session Management

The Session abstraction and storage layer are implemented in:

```text
internal/session/
```

Current functionality includes:

* Session creation
* Session IDs
* Session timestamps
* Session metadata
* Session renaming
* Session message storage
* In-memory session storage
* File-based session storage
* Database-backed session storage
* Session export
* Session import
* Set
* Get
* Delete
* List
* Agent session assignment
* Session-aware Agent message handling
* Persistent conversation history
* Default session restoration
* Active session switching
* Atomic file persistence
* Transactional database persistence

A Session contains:

```go
type Session struct {
    ID        string
    CreatedAt time.Time
    UpdatedAt time.Time
    Metadata  map[string]string
    Messages  []providerpkg.Message
}
```

`CreatedAt` records when the Session was created.

`UpdatedAt` is initialized when the Session is created and refreshed by `Rename`.

Appending conversation messages does not currently update `UpdatedAt`.

`Metadata` stores session-level information such as the display name.

`Messages` stores conversation history.

---

# Session Operations

## `NewSession`

```text
NewSession(id string) *Session
```

Creates a Session with:

* Supplied ID
* Current creation timestamp
* Current update timestamp
* Initialized metadata
* Initialized message list

Example:

```text
session := NewSession("project-a")
```

## `Rename`

```text
(*Session).Rename(name string)
```

Updates the Session display name in metadata and updates `UpdatedAt`.

The Session ID remains unchanged.

Example:

```text
session.Rename("Backend Project")
```

## `Export`

```text
(*Session).Export() ([]byte, error)
```

Serializes the complete Session into indented JSON.

Example:

```text
data, err := session.Export()
```

The exported data includes:

```text
ID

CreatedAt

UpdatedAt

Metadata

Messages
```

## `Import`

```text
Import(data []byte) (*Session, error)
```

Restores a Session from JSON.

It rejects invalid JSON and empty Session IDs.

Example:

```text
session, err := Import(data)
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

Available implementations are:

```text
SessionStore

MemoryStore

FileSessionStore

DatabaseSessionStore
```

All implement:

```go
Get(id string) (*Session, error)
Set(session *Session) error
Delete(id string) error
List() []*Session
```

The interactive CLI currently uses:

```text
FileSessionStore
```

The other implementations are package APIs and are not selected through the CLI.

## `Set`

```text
Set(session *Session) error
```

Stores or updates a Session.

Current implementations reject nil Sessions and empty IDs.

## `Get`

```text
Get(id string) (*Session, error)
```

Retrieves a Session by ID or returns a not-found/storage error.

## `Delete`

```text
Delete(id string) error
```

Removes a Session by ID or returns a not-found/storage error.

## `List`

```text
List() []*Session
```

Returns available Sessions.

The method has no error return, so storage implementations that encounter listing errors return an empty slice.

---

# In-Memory Session Store

The in-memory implementation is located in:

```text
internal/session/memory_store.go
```

`NewMemoryStore()` creates a MemoryStore.

The package also provides:

```text
NewSessionStore()
```

for the `SessionStore` implementation.

Both provide:

```text
Set

Get

Delete

List
```

Changes are lost when the application exits.

---

# File-Based Session Store

The file-based implementation is located in:

```text
internal/session/file_store.go
```

`NewFileSessionStore(dir string) *FileSessionStore` creates a store rooted at the supplied directory.

## `Set`

```text
Set(session *Session) error
```

Rejects nil Sessions and empty IDs.

It:

1. Ensures the store directory exists.
2. Serializes the Session.
3. Creates a temporary file.
4. Writes the serialized data.
5. Closes the temporary file.
6. Renames it to the final Session file.

This provides atomic replacement.

The temporary file is cleaned up after successful or failed operations.

## `Get`

```text
Get(id string) (*Session, error)
```

Reads and decodes the Session JSON file.

Missing files return a session-not-found error.

Read and JSON decoding errors are returned.

## `Delete`

```text
Delete(id string) error
```

Removes the Session JSON file.

Missing Sessions return a session-not-found error.

## `List`

```text
List() []*Session
```

Scans the Session directory for `.json` files.

Directories and entries that fail to load are skipped.

If reading the directory fails, an empty slice is returned because the method has no error result.

The helper:

```text
sessionPath(id string) string
```

maps a Session ID to:

```text
<id>.json
```

The helper:

```text
ensureDir(dir string) error
```

creates the directory tree with mode `0700`.

---

# Database Session Store

The database implementation is located in:

```text
internal/session/database_store.go
```

It uses:

```text
modernc.org/sqlite
```

SQLite is embedded and does not require a separate database server.

`NewDatabaseSessionStore(path string) (*DatabaseSessionStore, error)` opens the database and initializes the required tables.

If initialization fails, the opened database is closed before returning the error.

The store maintains:

```text
sessions

messages
```

The `sessions` table stores:

* Session ID
* Creation timestamp
* Update timestamp
* Metadata

The `messages` table stores:

* Session ID
* Message role
* Message content
* Tool calls

Messages reference Sessions through a foreign key with cascade deletion.

Database writes use transactions.

---

# Database Session Store Operations

## `Set`

```text
Set(session *Session) error
```

Rejects nil Sessions and empty IDs.

It writes Session metadata and messages in a transaction.

Metadata and message tool calls are serialized as JSON.

If any operation fails, the transaction is rolled back.

If the Session already exists, its metadata and `updated_at` are updated while the original `created_at` is retained.

Existing messages are replaced with the supplied message history within the same transaction.

## `Get`

```text
Get(id string) (*Session, error)
```

Retrieves a Session and ordered message history from SQLite.

It returns a session-not-found error when absent.

Query, scan, and JSON decoding errors are returned.

Tool calls are reconstructed from their stored JSON representation.

## `Delete`

```text
Delete(id string) error
```

Deletes a Session.

It returns a session-not-found error if no row was deleted.

Foreign-key cascade deletion removes associated messages.

## `List`

```text
List() []*Session
```

Returns all Sessions ordered by creation time.

Each Session is reconstructed with its metadata and associated messages.

Because the method has no error return, database query/load failures result in an empty slice.

---

# Database Session Storage

The database-backed architecture is:

```text
                    ┌──────────────────────┐
                    │       Session        │
                    │                      │
                    │ ID / Timestamps      │
                    │ Metadata / Messages  │
                    └──────────┬───────────┘
                               │
             ┌─────────────────┼─────────────────┐
             │                 │                 │
             ▼                 ▼                 ▼
      ┌────────────┐   ┌──────────────┐   ┌──────────────────┐
      │ MemoryStore│   │ File Store   │   │ Database Store   │
      └────────────┘   └──────┬───────┘   └────────┬─────────┘
                              │                    │
                              ▼                    ▼
                       .sessions/*.json      SQLite Database
```

The database relationship is:

```text
Session

   │

   │ 1

   ▼

Messages

   │

   │ many

   ▼

Session Messages
```

Transactions ensure that Session metadata and message history are updated consistently.

---

# Session Export and Import

Sessions support JSON export and import.

The export operation:

```text
Export
```

serializes:

```text
ID

CreatedAt

UpdatedAt

Metadata

Messages
```

The import operation:

```text
Import
```

reconstructs a Session independently from the active Session Store.

This provides a foundation for Session backup and migration.

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

When a new Session is created through the CLI, the newly created Session is assigned to the Agent and becomes active.

---

# Session Persistence

The application persists the active Session after successful Agent interactions.

For file-based persistence:

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

Atomic Session Write

    │

    ▼

.sessions/<id>.json
```

For database-backed persistence:

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

DatabaseSessionStore

    │

    ▼

Database Transaction

    │

    ▼

SQLite Database
```

When the application starts, the `default` Session is restored if it already exists.

Creating a new Session through the CLI immediately activates it.

---

# Database Session Testing

Database tests are located in:

```text
internal/session/database_store_test.go
```

They cover:

* Database initialization
* Session persistence
* Session metadata persistence
* Message persistence
* Tool-call persistence and round-trip restoration
* Session retrieval
* Session deletion
* Cascading message deletion
* Session listing
* Nil/empty-ID validation
* Missing-session errors
* Invalid stored metadata/tool-call JSON
* Database query failures
* Transaction rollback after forced message-insert failures

The database implementation serializes message tool calls and the tests verify that tool calls survive a database round trip.

File-store tests cover:

* Invalid inputs
* Filesystem errors
* JSON errors
* Missing Sessions
* Atomic persistence
* Temporary-file cleanup

CLI tests cover rollback of MCP add/remove when configuration saving fails.

Run:

```bash
go test ./internal/session -v
```

---

# CLI

The interactive CLI is located in:

```text
cmd/main.go
```

CLI tests are located in:

```text
cmd/main_test.go
```

Run the application with:

```bash
go run ./cmd
```

Build and run:

```bash
go build -o agent-harness ./cmd
./agent-harness
```

The CLI requires:

```text
config/config.json
```

or:

```text
../config/config.json
```

when launched from a child directory.

Create a local configuration from the example:

```bash
cp config/config.example.json config/config.json
```

The example configuration includes the repository MCP test server.

Build the test server first or remove the test-server entry before starting the CLI.

The CLI starts with the default Session:

```text
default
```

---

# CLI Implementation

Important CLI functions include:

### `loadConfig`

```text
loadConfig() (*configpkg.Config, error)
```

Tries:

```text
config/config.json

../config/config.json
```

and returns the first valid configuration or the relevant load error.

### `resolveModel`

```text
resolveModel(configModel string) (string, string)
```

Chooses:

1. Trimmed `OLLAMA_MODEL` environment override.
2. Trimmed configured model.
3. Built-in default:

```text
kirito1/qwen3-coder:4b
```

It also returns the source:

```text
environment

config

default
```

### `getConfigPath`

```text
getConfigPath() (string, error)
```

Searches the supported configuration paths and returns the first existing path.

### `selectModel`

```text
selectModel(
    providerClient providerpkg.Provider,
    configuredModel string,
    configuredSource string,
    input *bufio.Scanner,
    output io.Writer,
) (string, string, error)
```

For providers without `LocalModelManager`, it returns the configured model without local model discovery.

For Ollama, it:

1. Lists installed models.
2. Checks whether the configured model is installed.
3. Offers to pull the configured model when missing.
4. Allows selection of another installed model.
5. Allows quitting.

Listing, pulling, and input errors are returned.

### `handleCommand`

```text
handleCommand(
    input string,
    model string,
    modelSource string,
    currentSession **sessionpkg.Session,
    sessionStore sessionpkg.Store,
    agentClient *agentpkg.Agent,
    appConfig *configpkg.Config,
    mcpRuntime *mcppkg.Runtime,
    registry *toolspkg.ToolRegistry,
    output io.Writer,
) CommandResult
```

Parses and handles built-in CLI commands.

It supports:

```text
help

model

session

session create

session list

session load

session delete

clear

mcp

mcp list

mcp add

mcp remove

exit
```

Session creation/loading updates both the active Session and Agent.

MCP add/remove update configuration and runtime together and attempt rollback when configuration saving fails.

---

# CLI Commands

| Command                              | Description                            |
| ------------------------------------ | -------------------------------------- |
| `help`                               | Show available commands                |
| `model`                              | Show the selected model and its source |
| `session`                            | Show the current session ID            |
| `session create <id>`                | Create and activate a new Session      |
| `session list`                       | List stored Sessions                   |
| `session load <id>`                  | Load and activate a stored Session     |
| `session delete <id>`                | Delete a stored Session                |
| `clear`                              | Clear the terminal display             |
| `mcp`                                | Show MCP command usage                 |
| `mcp list`                           | List configured MCP servers            |
| `mcp add <name> <command> [args...]` | Connect and persist an MCP server      |
| `mcp remove <name>`                  | Disconnect and remove an MCP server    |
| `exit`                               | Exit the CLI                           |

---

# `help`

Displays available CLI commands:

```text
help
```

Example:

```text
Available commands: help, exit, model, session, clear, mcp
```

---

# `model`

Displays the selected model and its source:

```text
model
```

For Ollama, the model can be overridden using:

```bash
export OLLAMA_MODEL=your-model
```

Example:

```text
Current model : llama3.1
Source: config
```

---

# `session`

Displays the active Session:

```text
session
```

Example:

```text
Current session: default
```

---

# `session create`

Creates and activates a new Session:

```text
session create project-a
```

The command:

1. Creates the Session.
2. Persists it.
3. Assigns it to the Agent.
4. Updates the active Session reference.

Example:

```text
Created and switched to session: project-a
```

---

# `session list`

Lists stored Sessions:

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

Loads an existing Session:

```text
session load project-a
```

The Agent is updated to use the loaded Session.

---

# `session delete`

Deletes a stored Session:

```text
session delete project-a
```

The CLI prevents deletion of the currently active Session.

---

# `clear`

Clears the terminal display:

```text
clear
```

It does not clear or delete conversation history.

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
config/config.go
```

Tests are located in:

```text
config/config_test.go
```

The configuration contains Provider and MCP settings.

Example:

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

The example configuration is:

```text
config/config.example.json
```

The local configuration:

```text
config/config.json
```

is ignored by Git.

`Config` contains:

```text
Provider

MCP
```

The Provider contains:

```text
Name

Model

BaseURL

Endpoint
```

MCP server entries contain:

```text
Name

Command

Args
```

---

# Configuration Operations

## `Load`

```text
Load(path string) (*Config, error)
```

Reads the configuration file, parses JSON, and validates required Provider fields.

It returns file-read, JSON-parse, or validation errors.

Example:

```text
config, err := config.Load("config/config.json")
```

## `Save`

```text
Save(path string, config *Config) error
```

Serializes the configuration as indented JSON with a trailing newline and writes it to the specified path.

A newly created file uses mode `0644`.

Example:

```text
err := config.Save("config/config.json", config)
```

## `Validate`

```text
(*Config).Validate() error
```

Checks that these Provider fields are non-empty:

```text
Name

Model

BaseURL

Endpoint
```

It does not whitelist provider names or validate MCP server contents.

Provider support is validated by the Provider Factory.

The configured `endpoint` is required by the configuration model, but the current concrete providers construct their standard API paths themselves.

---

# Providers

The Provider implementations are located in:

```text
internal/provider/
```

Currently supported:

```text
Ollama

OpenAI

Anthropic

Gemini

Mistral
```

The shared types and interfaces are defined in:

```text
internal/provider/provider.go
```

The Provider interface is:

```go
Chat(request ChatRequest) (ChatResponse, error)
```

---

# Ollama

Implementation:

```text
internal/provider/ollama.go
```

Tests:

```text
internal/provider/ollama_test.go
```

Ollama is the primary local provider.

Default base URL:

```text
http://localhost:11434
```

Default chat endpoint:

```text
/api/chat
```

Start Ollama:

```bash
ollama serve
```

Check models:

```bash
ollama list
```

Pull a model:

```bash
ollama pull llama3.1
```

## `Chat`

```text
Chat(request ChatRequest) (ChatResponse, error)
```

Sends a provider-neutral chat request to Ollama.

It validates the request, sends the model and messages, optionally supplies tools, and converts the Ollama response into the common Provider response type.

## `ListModels`

```text
ListModels() ([]string, error)
```

Sends a GET request to:

```text
/api/tags
```

and returns available model names.

Network, status, decoding, and empty-name errors are returned.

## `PullModel`

```text
PullModel(name string) error
```

Sends a POST request to:

```text
/api/pull
```

to pull the requested model.

Request, network, non-200 status, and response-read errors are returned.

## `HasModel`

```text
HasModel(name string) (bool, error)
```

Calls `ListModels()` and checks for an exact model-name match.

At startup the CLI uses these operations to check the configured model and offer pulling or selection of installed models.

---

# OpenAI

Implementation:

```text
internal/provider/openai.go
```

Tests:

```text
internal/provider/openai_test.go
```

The OpenAI provider uses the configured API base URL and API credential.

Default API base URL:

```text
https://api.openai.com/v1
```

Chat path:

```text
/chat/completions
```

## `Chat`

```text
Chat(request ChatRequest) (ChatResponse, error)
```

Sends a provider-neutral request to the OpenAI-compatible chat completions endpoint.

It converts the response into the common Provider representation.

Tool calls are normalized into:

```text
ToolCall
```

The provider validates that a model and at least one message are supplied.

---

# Anthropic

Implementation:

```text
internal/provider/anthropic.go
```

Tests:

```text
internal/provider/anthropic_test.go
```

Anthropic is integrated through the common Provider abstraction.

## `Chat`

```text
Chat(request ChatRequest) (ChatResponse, error)
```

Converts the common Agent request into the Anthropic request format, sends the request to the configured Anthropic API endpoint, and converts the response into the common Provider response type.

The provider handles:

* Request validation
* Authentication
* HTTP transport
* HTTP status errors
* Response decoding
* Text response extraction
* Tool-call conversion

Provider-specific request and response structures remain internal to the implementation.

---

# Google Gemini

Implementation:

```text
internal/provider/gemini.go
```

Tests:

```text
internal/provider/gemini_test.go
```

The Gemini provider integrates Google Gemini through the common Provider abstraction.

## `Chat`

```text
Chat(request ChatRequest) (ChatResponse, error)
```

Converts the provider-neutral request into Gemini's request structure and converts Gemini responses into the common Provider response type.

The implementation handles:

* Request validation
* Authentication
* Gemini request construction
* Tool-definition conversion
* Response conversion
* Tool-call conversion
* HTTP and decoding errors

Provider-specific structures remain internal to the Gemini implementation.

---

# Mistral

Implementation:

```text
internal/provider/mistral.go
```

Tests:

```text
internal/provider/mistral_test.go
```

Mistral is integrated through the common Provider abstraction.

## `Chat`

```text
Chat(request ChatRequest) (ChatResponse, error)
```

Sends a provider-neutral request to the Mistral API and converts the response into the common Provider representation.

The implementation handles:

* Request validation
* Authentication
* HTTP transport
* HTTP status errors
* Response decoding
* Text responses
* Tool-call conversion

---

# Provider Factory

Implementation:

```text
internal/provider/factory.go
```

Tests:

```text
internal/provider/factory_test.go
```

## `NewProvider`

```text
NewProvider(name, baseURL, apiKey string) (Provider, error)
```

Creates a Provider implementation based on the provider name.

Supported values:

```text
ollama

openai

gemini

anthropic

mistral
```

It forwards the supplied base URL to the selected provider and passes credentials to providers that require them.

Unsupported provider names return an error.

Example:

```text
provider, err := NewProvider(
    "gemini",
    "https://generativelanguage.googleapis.com",
    apiKey,
)
```

## `RequiresAPIKey`

```text
RequiresAPIKey(name string) bool
```

Reports whether the provider requires a stored API credential.

Current behavior:

```text
ollama    → false
openai    → true
gemini    → true
anthropic → true
mistral   → true
```

Unknown provider names return:

```text
false
```

---

# Model Override

The configured Ollama model can be overridden using:

```bash
export OLLAMA_MODEL=llama3.1
```

This allows local model testing without modifying the configuration file.

The override is used by the CLI's model-resolution logic.

---

# Orchestrator

The Orchestrator is implemented in:

```text
internal/orchestrator/orchestrator.go
```

Tests:

```text
internal/orchestrator/orchestrator_test.go
```

The Orchestrator coordinates the Agent execution loop and evaluates tool actions through Jev before execution on the `AssignTool` path.

It handles:

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

        Jev Evaluation

              │

       ┌──────┼──────┐
       │      │      │
       ▼      ▼      ▼
     Allow  Confirm  Deny

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

* The model produces a final response.
* The maximum tool-call limit is reached.
* An error occurs.

## `NewOrchestrator`

```text
NewOrchestrator(
    agentClient *agentpkg.Agent,
    providerClient providerpkg.Provider,
    options ...OrchestratorOption,
) *DefaultOrchestrator
```

Creates the Orchestrator.

The default maximum tool-call limit is:

```text
10
```

The default action evaluator is:

```text
jev.NewDefaultEvaluator()
```

Options can modify this behavior.

## `WithMaxToolCalls`

```text
WithMaxToolCalls(maxToolCalls int) OrchestratorOption
```

Changes the maximum tool-call limit when the supplied value is positive.

Zero or negative values leave the current limit unchanged.

## `WithEvaluator`

```text
WithEvaluator(evaluator jev.Evaluator) OrchestratorOption
```

Replaces the default Jev evaluator when a non-nil evaluator is supplied.

This allows custom action-evaluation behavior to be injected without changing the Orchestrator interface.

## `RunAgent`

```text
RunAgent(request providerpkg.ChatRequest) (AgentResponse, error)
```

Runs the complete conversational Agent loop.

It:

1. Requires at least one request message.
2. Uses the last request message as the new user message.
3. Adds the message to Agent history.
4. Replaces request tools with current registry definitions.
5. Uses the complete Agent conversation history.
6. Calls the Provider.
7. Executes native tool calls.
8. Supports multiple tool calls per response.
9. Adds tool results to conversation history.
10. Continues until a final response or error.

Legacy JSON responses containing a `tool_call` are also handled.

The method returns errors for:

* Empty requests
* Tool-schema errors
* Provider failures
* Tool failures
* Invalid AgentResponse JSON
* Maximum tool-call limit exhaustion

## `Chat`

```text
Chat(request provider.ChatRequest) (provider.ChatResponse, error)
```

Forwards one request directly to the Provider.

It also prints a request-status line.

## `Run`

```text
Run(name string, args map[string]any) (any, error)
```

Directly executes a named tool through the Agent.

It prints a tool-status line.

## `AssignTool`

```text
AssignTool(toolCall ToolCall) (any, error)
```

Evaluates and executes the named tool using the supplied arguments.

The Orchestrator first converts the supplied `ToolCall` into a Jev `Action`.

The Jev evaluator then returns:

```text
Allow

Confirm

Deny
```

An `Allow` decision executes the tool.

A `Confirm` decision returns an error indicating that confirmation is required.

A `Deny` decision returns an error indicating that the action was denied.

Execution errors are wrapped with the tool name.

## `GetToolSchemas`

```text
GetToolSchemas() ([]map[string]any, error)
```

Returns the Agent registry schemas.

## `GetToolDefinitions`

```text
GetToolDefinitions() ([]provider.ToolDefinition, error)
```

Converts registry schemas into Provider tool definitions.

It returns errors when required name, description, or parameter schema information is missing or invalid.

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
                                    ▼
                         ┌──────────────────────┐
                         │    Orchestrator      │
                         │                      │
                         │ Agent / Provider     │
                         │ Execution Loop       │
                         └───────┬───────┬──────┘
                                 │       │
                    ┌────────────┘       └────────────┐
                    ▼                                 ▼
             ┌──────────────┐                 ┌──────────────┐
             │   Provider   │                 │     Jev      │
             │              │                 │  Evaluator   │
             │ Ollama       │                 │              │
             │ OpenAI       │                 │ Allow        │
             │ Anthropic    │                 │ Confirm      │
             │ Gemini       │                 │ Deny         │
             │ Mistral      │                 └──────┬───────┘
             └──────────────┘                        │
                                                     ▼
                                            ┌─────────────────┐
                                            │  Tool Registry  │
                                            │                 │
                                            │ Calculator      │
                                            │ Shell           │
                                            │ Filesystem      │
                                            │ MCP Tools       │
                                            └────────┬────────┘
                                                     │
                                      ┌──────────────┴──────────────┐
                                      │                             │
                                      ▼                             ▼
                               Native Tools                   MCP Adapters
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

Jev is an action-evaluation boundary between the Orchestrator and tool execution.

The existing Provider and Tool interfaces remain independent of Jev.

---

# Session Architecture

```text
                    ┌────────────────────────┐
                    │        Session         │
                    │                        │
                    │          ID            │
                    │      CreatedAt         │
                    │      UpdatedAt         │
                    │       Metadata         │
                    │       Messages         │
                    └───────────┬────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              │                 │                 │
              ▼                 ▼                 ▼
      ┌──────────────┐  ┌──────────────┐  ┌──────────────────────┐
      │ MemoryStore  │  │ File Store   │  │ Database Store       │
      └──────────────┘  └──────┬───────┘  └──────────┬───────────┘
                               │                     │
                               ▼                     ▼
                        .sessions/*.json       SQLite Database
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

   Orchestrator

       │

       ▼

   Jev Evaluator

       │

       ├── Allow
       │
       ├── Confirm
       │
       └── Deny

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

Jev is evaluated on the Orchestrator `AssignTool` path. Direct calls through `Run()` remain a separate direct tool-execution path.

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

Orchestrator

 │

 ▼

Jev Evaluator

 │

 ├── Allow / Confirm / Deny

 │

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

 ├── CreatedAt

 ├── UpdatedAt

 ├── Metadata

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

 ├── File Store

 │      │

 │      ▼

 │   Atomic File Persistence

 │

 └── Database Store

        │

        ▼

     SQLite Transaction
```

---

# Project Structure

The following structure reflects the current repository.

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
│   │
│   └── mcp-test-server/
│       └── main.go
│
├── config/
│   ├── config.go
│   ├── config_test.go
│   └── config.example.json
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
│   │   ├── store.go
│   │   └── store_test.go
│   │
│   ├── jev/
│   │   ├── action.go
│   │   ├── basic_policy.go
│   │   ├── basic_policy_test.go
│   │   ├── decision.go
│   │   ├── default_evaluator.go
│   │   ├── default_evaluator_test.go
│   │   ├── evaluator.go
│   │   ├── policy.go
│   │   ├── shell_policy.go
│   │   └── shell_policy_test.go
│   │
│   ├── mcp/
│   │   ├── adapter.go
│   │   ├── adapter_test.go
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── runtime.go
│   │   ├── runtime_test.go
│   │   ├── tools.go
│   │   └── tools_test.go
│   │
│   ├── orchestrator/
│   │   ├── orchestrator.go
│   │   └── orchestrator_test.go
│   │
│   ├── provider/
│   │   ├── anthropic.go
│   │   ├── anthropic_test.go
│   │   ├── factory.go
│   │   ├── factory_test.go
│   │   ├── gemini.go
│   │   ├── gemini_test.go
│   │   ├── mistral.go
│   │   ├── mistral_test.go
│   │   ├── ollama.go
│   │   ├── ollama_test.go
│   │   ├── openai.go
│   │   ├── openai_test.go
│   │   └── provider.go
│   │
│   ├── session/
│   │   ├── database_store.go
│   │   ├── database_store_test.go
│   │   ├── file_store.go
│   │   ├── file_store_test.go
│   │   ├── memory_store.go
│   │   ├── memory_store_test.go
│   │   ├── session.go
│   │   ├── session_test.go
│   │   ├── store.go
│   │   └── store_test.go
│   │
│   └── tools/
│       ├── ToolRegistry.go
│       ├── filesystem_tool.go
│       ├── filesystem_tool_test.go
│       ├── shell_tool.go
│       ├── shell_tool_test.go
│       ├── tools.go
│       └── tools_test.go
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

is local build output and is intentionally excluded from Git.

The runtime-created Session directory:

```text
.sessions/
```

is also local application data and is not part of the repository structure.

The local configuration:

```text
config/config.json
```

is ignored by Git and therefore is not included in the tracked repository structure.

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

Run provider tests:

```bash
go test ./internal/provider -v
```

Run MCP tests:

```bash
go test ./internal/mcp -v
```

Run Jev tests:

```bash
go test ./internal/jev -v
```

Jev tests cover:

* Action and decision models
* Basic policy
* Shell policy
* Default evaluator
* Allow decisions
* Confirm decisions
* Deny decisions
* Policy priority
* Invalid actions
* Orchestrator evaluator integration

The Jev package currently has 100% statement coverage.

The overall project coverage is currently approximately 85.3%.

Run coverage for the entire project:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Run CLI tests:

```bash
go test ./cmd
```

Run session tests:

```bash
go test ./internal/session -v
```

Run database session tests:

```bash
go test ./internal/session -run Database -v
```

---

# Integration Testing

## Ollama Orchestrator Integration

Enable the Ollama-backed Orchestrator integration test using:

```bash
ORCHESTRATOR_INTEGRATION=1 go test ./...
```

Requirements:

* Ollama running locally
* Compatible model installed
* Model supporting the required tool-calling behavior

## Direct Ollama Integration

The direct provider integration test can be enabled using:

```bash
OLLAMA_INTEGRATION=1 go test ./internal/provider -run TestOllamaProvider_Integration
```

Integration tests are opt-in so normal test runs do not require a running Ollama instance.

---

# Formatting

Format Go files with:

```bash
gofmt -w .
```

Check formatting:

```bash
test -z "$(gofmt -l .)"
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

GitHub Actions validates the repository automatically.

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

Build MCP Test Server

   │

   ▼

go test -v ./...

   │

   ▼

go vet ./...
```

The workflow runs for pushes to:

```text
main

feature/**
```

and pull requests targeting:

```text
main
```

The MCP test server is built using:

```bash
go build -o cmd/mcp-test-server/mcp-test-server ./cmd/mcp-test-server
```

---

# Development Workflow

Create a feature branch:

```bash
git switch -c feature/<name>
```

Format the project:

```bash
gofmt -w .
```

Run tests:

```bash
go test ./...
```

Run static analysis:

```bash
go vet ./...
```

Check formatting:

```bash
test -z "$(gofmt -l .)"
```

Build the MCP test server:

```bash
go build -o cmd/mcp-test-server/mcp-test-server ./cmd/mcp-test-server
```

Check the working tree:

```bash
git status
```

Review changes:

```bash
git diff
```

Stage a logical change:

```bash
git add <files>
```

Commit:

```bash
git commit -m "your commit message"
```

Push:

```bash
git push origin feature/<name>
```

Then open a pull request against `main`.

---

# Architectural Principles

## Provider Neutrality

The Agent does not depend directly on provider-specific APIs.

Provider-specific behavior belongs behind the Provider interface.

Current providers are:

```text
Ollama

OpenAI

Anthropic

Gemini

Mistral
```

---

## Tool Agnosticism

The Agent does not need to know whether a tool is:

* Built into the project
* Provided by an MCP server
* Added by another future integration

All tools use the common Tool abstraction.

---

## Deterministic Action Evaluation

Tool actions can be evaluated independently of the model provider.

Jev separates action evaluation from tool implementation:

```text
Tool Call

    +

Action Policies

    │

    ▼

Allow / Confirm / Deny
```

The default evaluator currently applies:

```text
Deny > Confirm > Allow
```

This allows action-evaluation policies to evolve independently from Providers and Tools.

---

## Session Independence

Session storage is separated from the Agent runtime.

Current storage backends are:

```text
In-Memory

File

SQLite
```

Different storage implementations can therefore evolve without rewriting Agent execution.

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

Action Evaluation

Tools

MCP

CLI
```

Each subsystem has a focused responsibility.

Jev provides the action-evaluation boundary without coupling the evaluation layer to provider-specific behavior or individual Tool implementations.

---

## Local First

The implementation supports local models through Ollama.

This allows development and testing without requiring a hosted model API.

Cloud providers can use the same Provider abstraction when credentials are available.

---

## Extensibility

The architecture allows new capabilities to be added without rewriting the Agent core.

Potential future extensions include:

```text
Additional Providers

Additional Tool Sources

Additional Credential Backends

Additional Session Backends

Additional Action Policies

More Granular Tool Policies

Remote MCP Servers

Additional MCP Capabilities
```

---

# Implementation Status

| Component                             | Status      |
| ------------------------------------- | ----------- |
| Go project structure                  | Implemented |
| Provider abstraction                  | Implemented |
| Provider factory                      | Implemented |
| Provider API-key detection            | Implemented |
| Ollama provider                       | Implemented |
| OpenAI provider                       | Implemented |
| Anthropic provider                    | Implemented |
| Google Gemini provider                | Implemented |
| Mistral provider                      | Implemented |
| Credential-aware provider creation    | Implemented |
| Conversation history                  | Implemented |
| Agent runtime                         | Implemented |
| Session-aware Agent runtime           | Implemented |
| Tool abstraction                      | Implemented |
| Tool registry                         | Implemented |
| Jev action model                      | Implemented |
| Jev decision model                    | Implemented |
| Jev evaluator abstraction             | Implemented |
| Jev policy abstraction                | Implemented |
| Jev basic policy                      | Implemented |
| Jev shell policy                      | Implemented |
| Jev default evaluator                 | Implemented |
| Jev policy priority                   | Implemented |
| Jev orchestrator integration          | Implemented |
| Jev unit tests                        | Implemented |
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
| MCP dynamic server registration       | Implemented |
| MCP dynamic server removal            | Implemented |
| MCP test server                       | Implemented |
| MCP CLI listing                       | Implemented |
| MCP CLI server addition               | Implemented |
| MCP CLI server removal                | Implemented |
| MCP CLI configuration persistence     | Implemented |
| Credential abstraction                | Implemented |
| Credential storage                    | Implemented |
| Credential retrieval                  | Implemented |
| Session abstraction                   | Implemented |
| Session message storage               | Implemented |
| Session timestamps                    | Implemented |
| Session metadata                      | Implemented |
| Session renaming                      | Implemented |
| Session export                        | Implemented |
| Session import                        | Implemented |
| In-memory session store               | Implemented |
| File-based session store              | Implemented |
| Database-backed session store         | Implemented |
| Persistent Agent conversation history | Implemented |
| Automatic session persistence         | Implemented |
| Atomic file persistence               | Implemented |
| Transactional database persistence    | Implemented |
| Default session restoration           | Implemented |
| Active session switching on creation  | Implemented |
| Interactive CLI                       | Implemented |
| CLI model command                     | Implemented |
| CLI session display                   | Implemented |
| CLI terminal clearing                 | Implemented |
| CLI session creation                  | Implemented |
| CLI session listing                   | Implemented |
| CLI session loading                   | Implemented |
| CLI session deletion                  | Implemented |
| Unit tests                            | Implemented |
| Integration tests                     | Implemented |
| Go formatting checks                  | Implemented |
| Go vet checks                         | Implemented |
| GitHub Actions CI                     | Implemented |

---

# Roadmap

## Improved MCP Support

The core MCP integration, CLI server management, and dynamic MCP server lifecycle are implemented.

Future MCP work includes:

* Improved MCP server lifecycle management
* Better MCP error handling
* MCP resource support
* MCP prompts
* MCP sampling where applicable
* Remote MCP transports
* Additional MCP protocol capabilities
* Improved MCP tool lifecycle management
* Better handling of tools when an MCP server is disconnected

The current runtime primarily focuses on local command-based MCP servers.

---

## Provider Improvements

The current Provider abstraction supports:

```text
Ollama

OpenAI

Anthropic

Gemini

Mistral
```

Future provider work may include:

* Additional local model runtimes
* Additional OpenAI-compatible APIs
* Provider-specific capability detection
* More complete provider-specific tool-use behavior
* Improved streaming support

New providers should continue to use the existing Provider abstraction rather than introducing provider-specific logic into the Agent.

---

## Improved CLI

Planned CLI improvements include:

```text
Provider Management

Tool Listing

Configuration Inspection

Advanced Model Management

Persistent Session Selection

Interactive Tool Inspection
```

MCP server management is already implemented through:

```text
mcp list

mcp add

mcp remove
```

Future CLI improvements can build richer inspection and management capabilities on top of the existing commands.

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

Orchestrator

 │

 ▼

Jev Evaluator

 │

 ▼

Allow

 │

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

Provider

 │

 │ Tool Call

 ▼

Orchestrator

 │

 ▼

Jev Evaluator

 │

 ▼

Allow

 │

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

Provider

 │

 ▼

Final Response
```

A simplified MCP server addition workflow looks like:

```text
User

 │

 │ mcp add test-server ./test-server --port 8080

 ▼

CLI

 │

 ▼

MCP Runtime

 │

 ├── Connect MCP Server

 ├── Create Client Session

 ├── Discover Tools

 └── Register Tools

 │

 ▼

Update Configuration

 │

 ▼

Save config.json

 │

 ▼

Server Available Immediately
```

A simplified MCP server removal workflow looks like:

```text
User

 │

 │ mcp remove test-server

 ▼

CLI

 │

 ▼

MCP Runtime

 │

 ▼

Disconnect Active Session

 │

 ▼

Remove From Configuration

 │

 ▼

Save config.json

 │

 ▼

Server No Longer Active
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

 ├── CreatedAt

 ├── UpdatedAt

 ├── Metadata

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

Updated Session

 │

 ▼

Session Store

 │

 ├── FileSessionStore

 │      │

 │      ▼

 │   Atomic File Persistence

 │

 └── DatabaseSessionStore

        │

        ▼

     SQLite Transaction
```

A Session creation flow looks like:

```text
User

 │

 │ session create project-a

 ▼

CLI

 │

 ▼

NewSession

 │

 ▼

Session Store

 │

 ▼

Agent.SetSession

 │

 ▼

Active Session

 │

 ▼

Agent Runtime
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

Action Evaluation

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

The project is designed to evolve toward a more complete Agent runtime while maintaining clear boundaries between:

```text
Providers

Tools

Jev

Sessions

Credentials

MCP

CLI
```

---

# License

This project is licensed under the MIT License.

See the `LICENSE` file for the full license text.
