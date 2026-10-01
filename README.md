# Agent Harness

Agent Harness is a Go application and runtime for running a conversational agent that can call local tools and tools exposed by MCP servers. The interactive CLI currently connects to Ollama or OpenAI, keeps conversation history in file-backed sessions, and routes provider tool calls through a shared tool registry.

It provides:

* Provider abstraction
* Agent orchestration
* Conversation history
* Native tool execution
* MCP integration
* MCP server lifecycle management
* MCP server CLI management
* Credential storage
* Session management and persistence
* Interactive CLI
* Local Ollama model management
* Automated testing and CI

The core is provider-neutral and modular. Ollama provides the local-first path; OpenAI is also supported when an API key is configured.

---

# Current Status

The project currently supports:

* Provider abstraction with Ollama and OpenAI integration
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
* Adding additional providers
* Increasing runtime extensibility

The interactive CLI currently uses the file session store in `.sessions/`. The in-memory and SQLite stores are available as package implementations but are not selected by the CLI.

---

# Features

## Provider Abstraction

The runtime uses provider-neutral interfaces so the Agent does not depend directly on a specific model provider.

Currently integrated:

* Ollama
* OpenAI

The common `Provider` interface defines `Chat(request)`, which sends a provider-neutral request and returns response text and any tool calls.

Ollama additionally implements the separate `LocalModelManager` interface:

* `ListModels() ([]string, error)` sends a GET request to Ollama's `/api/tags` endpoint and returns the model names. Network, status, decoding, and empty-name errors are returned.
* `PullModel(name string) error` sends a POST request to Ollama's `/api/pull` endpoint and returns request, network, non-200 status, or response-read errors.

`(*OllamaProvider).HasModel(name string) (bool, error)` calls `ListModels()` and checks for an exact name match, returning listing errors unchanged.

Provider-specific behavior remains behind the Provider abstraction.

### Provider Factory

The provider factory creates the appropriate provider implementation:

```text
NewProvider
```

`NewProvider(name, baseURL, apiKey string) (Provider, error)` creates an `ollama` or `openai` provider with the supplied base URL. It passes the key to OpenAI, ignores it for Ollama, and returns an error for unsupported names. The CLI loads an OpenAI API key from the credential store before calling the factory.

Currently supported providers are:

```text
ollama
openai
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

### Provider API

The common Provider interface exposes:

```text
Chat
```

`Chat(request ChatRequest) (ChatResponse, error)` sends the model, messages, and optional tool definitions to the configured provider. It returns response text and normalized tool calls, or an error for invalid requests, transport/status failures, response-decoding failures, or a missing OpenAI choice.

Example:

```text
Chat(request)
```

Conceptually:

```text
Provider

   │

   ▼

Chat(request)

   │

   ▼

Model Response
```

Local model providers can additionally implement:

```text
ListModels

PullModel
```

Only Ollama implements `LocalModelManager` in the current code; OpenAI supports chat requests but not model listing or pulling.

`RequiresAPIKey(name string) bool` reports whether a provider requires an API key (`true` for `openai`). It returns `false` for other names, including unknown names, so `NewProvider` can report unsupported providers.

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
* Configurable maximum number of tool calls
* Provider-neutral execution
* Session-aware conversation history

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

Assigns a non-nil Session and makes its `Messages` slice the active conversation history. It returns an error for a nil Session; it does not copy the Session.

Example:

```go
err := agent.SetSession(session)
```

After this operation, Agent conversation messages are associated with the assigned Session.

### `AddMessage`

```text
AddMessage(message provider.Message) error
```

Appends the message to the active Session or, when no Session is assigned, to the Agent's in-memory Conversation. It rejects a message with an empty role.

### `GetMessages`

```text
GetMessages() []provider.Message
```

Returns the active Session's `Messages` slice or the in-memory Conversation's message slice. The returned slice is not copied.

Example:

```go
messages := agent.GetMessages()
```

The returned messages can then be used by the runtime when constructing a Provider request.

### Tool access

* `GetToolSchemas() ([]map[string]any, error)` obtains all tool schemas from the registry and propagates registry/schema errors.
* `ExecuteTool(name string, args map[string]any) (any, error)` looks up `name`, executes the tool with `args`, and returns lookup or tool errors.
* `Run(name string, args map[string]any) (any, error)` delegates directly to `ExecuteTool`; the conversational provider/tool loop is `DefaultOrchestrator.RunAgent(request)`.

The `Conversation` type backs an Agent that has no assigned Session:

* `NewConversation() *Conversation` creates an empty message slice.
* `(*Conversation).AddMessage(message provider.Message) error` rejects an empty role; otherwise appends the message and returns nil.
* `(*Conversation).GetMessages() []provider.Message` returns the underlying history slice without copying it.

---

# Tool System

## Tool Interface

Tools implement this common interface:

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

### Tool Operations

#### `Name`

```text
Name() string
```

Returns the tool name used by the registry and provider tool-call routing.

Example:

```text
tool.Name()
```

#### `Description`

```text
Description() string
```

Returns a human-readable description supplied to the model.

Example:

```text
tool.Description()
```

#### `Execute`

```text
Execute(args map[string]any) (any, error)
```

Executes the tool using the provided argument map and returns its result or an error.

Example:

```text
tool.Execute(arguments)
```

The exact arguments depend on the Tool implementation.

#### `Schema`

```text
Schema() map[string]any
```

Returns the parameter schema supplied to the model. Schema validation and argument semantics are implemented by each tool.

Example:

```text
tool.Schema()
```

---

# Tool Registry

`NewToolRegistry() *ToolRegistry` creates the registry and registers the built-in `calculator`, `shell`, and `filesystem` tools.

It supports:

```text
Register

Get

Has

List

Remove

Schemas
```

## Registry Operations

### `Register`

```text
Register(name string, tool Tool) (string, Tool)
```

Adds or replaces the tool under `name`, prints a registration message, and returns the supplied name and tool.

Example:

```text
registry.Register("calculator", calculator)
```

### `Get`

```text
Get(name string) (Tool, error)
```

Retrieves a registered tool by name, or returns an error if it is absent.

Example:

```text
tool, err := registry.Get("calculator")
```

### `Has`

```text
Has(name string) bool
```

Reports whether a tool with the specified name is registered.

Example:

```text
exists := registry.Has("calculator")
```

### `List`

```text
List() []string
```

Returns the names of all registered tools. The order is unspecified.

Example:

```text
names := registry.List()
```

### `Remove`

```text
Remove(name string) error
```

Removes a registered tool by name or returns an error if no such tool exists.

Example:

```text
registry.Remove("calculator")
```

### `Schemas`

```text
Schemas() ([]map[string]any, error)
```

Returns the schemas of registered tools so they can be supplied to a model provider.

Example:

```text
schemas, err := registry.Schemas()
```

Returns each tool's name, description, and parameter schema. It returns an error if the registry is nil or a registered tool or its schema is nil. The registry is shared by built-in and adapted MCP tools.

---

# Built-in Tools

## Calculator

The built-in `calculator` tool takes an `operation` and a `numbers` array. It supports:

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

Division and modulus operations also handle invalid zero divisors.

---

## Shell

The built-in `shell` tool launches an installed local executable with the supplied argument list. It does not invoke a shell to parse a command string.

### Shell Execution

The shell execution operation:

```text
Execute
```

runs the requested command and returns its output or an execution error.

Example:

```text
Execute({"command": "ls", "args": ["-la"]})
```

`Shell.Execute(command, args...)` checks that the executable is available, then returns combined output or an error. Commands run with the permissions of the current process.

---

## Filesystem

The built-in `filesystem` tool operates on local paths through the common Tool interface.

### Filesystem Execution

Its `Execute(args)` operation supports:

* `read`: returns a file's contents.
* `write`: writes the supplied `content` to a path.
* `list`: returns directory entry names.
* `exists`: reports whether a path exists.
* `delete`: removes a path.

Each request supplies an `operation` and `path`; `write` also requires `content`.

This Agent tool is separate from the top-level `filesystem.FileSystem` helper, whose `Read`, `Write`, `List`, `Search`, and `Delete` methods operate directly on paths.

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

## MCP Client Operations

### `NewClient`

```text
NewClient() *Client
```

Creates an MCP client identified as `agent-harness` version `0.1.0`.

Example:

```go
client := mcp.NewClient()
```

### `Connect`

```text
Connect(ctx context.Context, transport mcpsdk.Transport) (*mcpsdk.ClientSession, error)
```

Establishes an MCP connection over the supplied transport and returns its client session or a connection error.

Example:

```go
session, err := client.Connect(ctx, transport)
```

The returned ClientSession is used for subsequent MCP operations.

### `ConnectCommand`

```text
ConnectCommand(ctx context.Context, command string, args []string) (*mcpsdk.ClientSession, error)
```

Starts `command` with `args` using `exec.CommandContext`, then connects with the MCP command transport. Returns the session or an error.

Example:

```go
session, err := client.ConnectCommand(
    ctx,
    "./mcp-server",
    []string{"--port", "8080"},
)
```

### `ListTools`

```text
ListTools(ctx context.Context, session *mcpsdk.ClientSession) ([]*mcpsdk.Tool, error)
```

Requests the connected server's tool list and returns its tools, or the SDK error.

Example:

```go
tools, err := client.ListTools(ctx, session)
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
internal/mcp/tools.go
```

It provides:

* Tool name
* Tool description
* Tool schema
* Tool execution

The MCP tool uses the MCP client session to call the remote MCP server.

## MCP Tool Operations

### Tool Metadata

The MCP wrapper exposes the remote MCP tool's:

```text
Name

Description

Schema
```

These allow the tool to participate in the native Tool abstraction.

Examples:

```text
tool.Name()

tool.Description()

tool.Schema()
```

`Name() string` and `Description() string` return the remote tool metadata. `Schema() map[string]any` converts its input schema through JSON; it returns nil if the remote schema is nil or conversion fails.

`NewTool(tool mcpsdk.Tool) *Tool` wraps the SDK tool value.

### `Execute`

```text
Execute(ctx context.Context, session *mcpsdk.ClientSession, args map[string]any) (*mcpsdk.CallToolResult, error)
```

Calls the named remote tool with `args` through `session` using `ctx`, returning the MCP result or SDK error.

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

The Agent therefore does not need to know whether a tool originated from:

* Built-in code
* MCP
* Another future tool provider

### Adapter Operations

The adapter implements the native `tools.Tool` interface. Its metadata and schema methods delegate to the MCP tool wrapper; its `Execute(args)` method invokes the MCP tool using the stored context and client session, then collects returned text content.

`NewToolAdapter(ctx context.Context, tool *Tool, session *mcpsdk.ClientSession) *ToolAdapter` binds the wrapper to the context and client session used for calls. `Name() string`, `Description() string`, and `Schema() map[string]any` delegate to the wrapper. `Execute(args map[string]any) (any, error)` returns concatenated text content; it propagates call errors and turns an MCP `IsError` result into an error. Non-text result content is not included in the returned string.

This includes:

```text
Name

Description

Schema

Execute
```

Examples:

```text
adapter.Name()

adapter.Description()

adapter.Schema()

adapter.Execute(arguments)
```

---

# MCP Tool Registration

The MCP package provides tool registration functionality that:

1. Lists tools from an MCP server
2. Wraps each MCP tool
3. Creates a native Tool Adapter
4. Registers the adapter in the Tool Registry

This allows MCP tools to automatically become available to the Agent.

### Registration Operation

The registration operation:

```text
RegisterTools(ctx context.Context, session *mcpsdk.ClientSession, registry *toolspkg.ToolRegistry) ([]toolspkg.Tool, error)
```

discovers available MCP tools, wraps each in a `Tool` and `ToolAdapter`, registers the adapters under their tool names, and returns the registered tools. A discovery error is returned; registry registration itself has no error return.

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

The runtime is responsible for managing MCP server connections during the lifetime of the application.

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

The runtime keeps the client sessions alive because MCP Tool Adapters use those sessions when executing tools.

## MCP Runtime Operations

### Runtime Creation

Creates a new MCP runtime with an empty set of active server sessions.

```go
runtime := mcp.NewRuntime()
```

`NewRuntime() *Runtime` initializes empty maps for active client sessions and each server's registered tools.

The Tool Registry is supplied when connecting a server because discovered MCP tools are registered during the connection process.

### `ConnectServer`

```text
ConnectServer(ctx context.Context, name, command string, args []string, registry *toolspkg.ToolRegistry) error
```

Connects to a command-launched MCP server, creates its client session, discovers its tools, and registers those tools into `registry`. It rejects duplicate active names and returns connection or registration errors; if registration fails, it closes the new session.

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

The server becomes part of the active runtime immediately after a successful connection and tool registration.

### `DisconnectServer`

```text
DisconnectServer(name string, registry *toolspkg.ToolRegistry) error
```

Disconnects an active MCP server session.

Example:

```go
err := runtime.DisconnectServer("test-server", registry)
```

The server's registered tools are removed only when the registry still maps those names to the same adapter instances, then its session is closed. It returns an error for an unknown active name or a session close failure.

### `Close`

```text
Close(registry *toolspkg.ToolRegistry) error
```

Closes active MCP sessions and releases MCP runtime resources during application shutdown.

Example:

```go
defer runtime.Close(registry)
```

`Close` removes runtime-owned registered tools, attempts to close all active sessions, clears runtime maps, and returns the first close error, if any.

---

# MCP Runtime Lifecycle

The runtime supports both startup connections and dynamic server lifecycle operations.

## Application Startup

Configured MCP servers are connected during application startup:

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

A server can also be connected while the application is running:

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

A running server can be disconnected through the CLI:

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

## Tool Execution

While the application is running:

```text
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

# MCP CLI Server Management

MCP servers can be managed directly through the interactive CLI.

The CLI currently supports:

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

Displays the available MCP CLI operations.

```text
mcp
```

Example output:

```text
Usage: mcp list | mcp add <name> <command> [args...] | mcp remove <name>
```

---

## `mcp list`

Lists all configured MCP servers.

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

The command reads the currently loaded MCP configuration and displays each configured server's name, command, and arguments.

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

Another example using the repository's MCP test server:

```text
mcp add test-server ./cmd/mcp-test-server/mcp-test-server
```

For a server requiring multiple arguments:

```text
mcp add filesystem npx -y @modelcontextprotocol/server-filesystem /tmp
```

The command performs the following operations:

1. Checks the loaded MCP configuration for a duplicate name.
2. Finds the configuration file path.
3. Connects to the MCP server and creates a client session.
4. Discovers the server's tools and registers them in the Tool Registry.
5. Adds the server to the loaded configuration and saves it.

If saving fails, the CLI restores the in-memory configuration and disconnects the server.

Example output:

```text
Added MCP server: test-server
```

If the server already exists:

```text
MCP server "test-server" already exists.
```

### Important

`mcp add` dynamically connects the server to the active MCP runtime.

A successful `mcp add` therefore makes the server available immediately without requiring an application restart.

The server is also persisted in the configuration so it will be connected again during the next application startup.

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

The command performs the following operations:

1. Searches the configured MCP servers by name.
2. Finds the matching server.
3. Disconnects the active MCP runtime session.
4. Removes the server from the configuration.
5. Saves the updated configuration.

Example output:

```text
Removed MCP server: test-server
```

If the server does not exist:

```text
MCP server "missing-server" not found.
```

If the name is missing:

```text
Usage: mcp remove <name>
```

### Important

`mcp remove` disconnects the active MCP session before removing the server from persistent configuration.

The server therefore stops being active immediately without requiring an application restart.

---

# MCP CLI Workflow

A typical MCP configuration workflow is:

### 1. Add a server

```text
mcp add test-server ./test-server --port 8080
```

The server is connected immediately and its tools are registered.

### 2. Verify the configuration

```text
mcp list
```

Example:

```text
MCP servers:

- test-server: ./test-server --port 8080
```

### 3. Use the server immediately

No application restart is required.

The server's discovered MCP tools are available through the Tool Registry.

### 4. Remove the server when no longer required

```text
mcp remove test-server
```

The active MCP session is disconnected immediately.

### 5. Verify removal

```text
mcp list
```

If it was the only configured server:

```text
No MCP servers configured.
```

---

# MCP Test Server

The repository contains a small MCP server used for integration testing:

```text
cmd/mcp-test-server/main.go
```

It exposes two test tools:

```text
test_tool
error_tool
```

`test_tool` returns:

```text
MCP test tool executed successfully
```

`error_tool` returns a deliberate error for adapter and runtime error-path tests.

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

The binary is used by the MCP command-transport tests and is also built by the GitHub Actions CI workflow.

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
* MCP runtime server disconnection

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

Disconnect Sessions

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

The credential store is used by providers that require API credentials.

Credentials are kept separate from Provider logic so credential backends can evolve independently.

The main credential operations are:

```text
Set

Get

Delete
```

`NewFileStore()` returns the file-backed implementation of `CredentialStore`.
Its signature is `NewFileStore() (*FileStore, error)`; it resolves the current user's home directory and points the store at `~/.agent-harness/credentials.json`, returning a home-directory lookup error if resolution fails.

## Credential Operations

### `Set`

```text
Set(provider string, key string) error
```

Stores or updates the credential associated with a provider. It rejects an empty provider or key, creates the parent directory with mode `0700`, reads existing JSON when present, and writes the updated map with mode `0600`.

Example:

```text
credentials.Set("openai", apiKey)
```

### `Get`

```text
Get(provider string) (string, error)
```

Retrieves the stored credential for the specified provider. It rejects an empty provider and returns `ErrCredentialNotFound` if the file or provider entry is missing; file and JSON decoding errors are returned.

Example:

```text
key, err := credentials.Get("openai")
```

If no credential exists, the store returns a credential-not-found error.

### `Delete`

```text
Delete(provider string) error
```

Removes the stored credential associated with the provider and rewrites the JSON file with mode `0600`. It returns errors for an empty or unknown provider and for file/JSON failures.

Example:

```text
credentials.Delete("openai")
```

The current file-based credential store uses:

```text
~/.agent-harness/credentials.json
```

with restrictive filesystem permissions.

When created, the credential directory uses mode `0700` and the credentials file uses mode `0600`.

---

# Session Management

The project includes a Session abstraction and storage layer.

Current session functionality includes:

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

A Session currently contains:

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

`UpdatedAt` is initialized when the Session is created and refreshed by `Rename`; appending conversation messages does not currently update it.

`Metadata` stores session-level information such as the display name.

`Messages` stores the conversation history associated with the Session.

---

# Session Operations

## `NewSession`

```text
NewSession(id string) *Session
```

Creates a new Session with:

* The supplied Session ID
* Current creation timestamp
* Current update timestamp
* An initialized metadata map
* An initialized message list

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

Serializes the complete Session into indented JSON and returns any JSON encoding error.

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

Restores a Session from JSON. It returns a decoding error for invalid JSON and rejects an empty Session ID.

Example:

```text
session, err := Import(data)
```

This allows Session data to be reconstructed independently from the active Session Store.

`Import` rejects JSON with an empty Session ID.

---

# Session Renaming

Sessions support a display name through metadata.

The rename operation is:

```text
Rename
```

Example:

```text
session.Rename("Project Alpha")
```

Renaming updates the Session metadata and `UpdatedAt` timestamp.

The Session ID remains unchanged.

This means renaming a Session changes its display name without changing its persistent storage identity.

---

# Session Export and Import

Sessions support JSON export and import.

The export operation:

```text
Export
```

Example:

```text
data, err := session.Export()
```

serializes the complete Session into formatted JSON.

The exported data includes:

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

Example:

```text
session, err := Import(data)
```

restores a Session from JSON.

This provides a foundation for moving or backing up Session data independently from the active Session Store.

---

# Session Store

The Session Store provides:

```text
Set

Get

Delete

List
```

The package provides file and SQLite persistence plus in-memory implementations:

```text
MemoryStore

FileSessionStore

DatabaseSessionStore

SessionStore
```

`MemoryStore`, `FileSessionStore`, and `DatabaseSessionStore` implement `Store`:

```go
Get(id string) (*Session, error)
Set(session *Session) error
Delete(id string) error
List() []*Session
```

`SessionStore` is also an in-memory store with the same operations. `NewSessionStore() *SessionStore` creates it. `Set` rejects nil sessions and empty IDs; `Get` and `Delete` return a not-found error when the ID is absent; `List` returns the stored pointers and has no error result. The interactive CLI currently selects `FileSessionStore`; the other stores are package APIs and are not configurable from the CLI.

## Store Operations

### `Set`

```text
Set(session *Session) error
```

Stores or updates a Session and returns an error for invalid input or storage failure. All current stores reject nil sessions and empty IDs.

Example:

```text
store.Set(session)
```

### `Get`

```text
Get(id string) (*Session, error)
```

Retrieves a Session by its ID or returns a not-found or storage error.

Example:

```text
session, err := store.Get("project-a")
```

### `Delete`

```text
Delete(id string) error
```

Removes a Session by its ID or returns a not-found or storage error.

Example:

```text
store.Delete("project-a")
```

### `List`

```text
List() []*Session
```

Returns available Sessions as `[]*Session`; it has no error return. Consequently, store implementations that encounter listing errors return an empty slice.

Example:

```text
sessions := store.List()
```

---

# In-Memory Session Store

The in-memory store keeps Sessions for the lifetime of the process.

`NewMemoryStore()` creates a `MemoryStore`; `NewSessionStore()` also creates an in-memory `SessionStore`.

It provides:

```text
Set

Get

Delete

List
```

Changes are lost when the application exits.

Example:

```text
store := NewMemoryStore()

store.Set(session)

store.Get(session.ID)

store.List()

store.Delete(session.ID)
```

---

# File-Based Session Store

The file-based store persists Sessions as JSON files.

`NewFileSessionStore(dir string) *FileSessionStore` creates a store rooted at the supplied directory.

It provides:

```text
Set

Get

Delete

List
```

## `Set`

```text
Set(session *Session) error
```

Rejects a nil Session or empty ID, ensures the store directory exists, serializes the Session to JSON, and persists it through a temporary file followed by a rename. It returns encoding and filesystem errors.

Example:

```text
fileStore.Set(session)
```

The write flow is:

```text
Session

   │

   ▼

JSON Serialization

   │

   ▼

Temporary File

   │

   ▼

Write Session Data

   │

   ▼

Close Temporary File

   │

   ▼

Rename Temporary File

   │

   ▼

Final Session File
```

The Session is first written to a temporary file inside the Session Store directory.

After the data has been successfully written and the temporary file has been closed, the temporary file is renamed to the final Session file.

This prevents the final Session file from being replaced by partially written JSON if a write operation fails.

The temporary file is removed after the operation, including when writing or renaming fails.

## `Get`

```text
Get(id string) (*Session, error)
```

Reads and decodes the Session JSON file. Missing files return a session-not-found error; read and JSON decoding errors are also returned.

Example:

```text
session, err := fileStore.Get("project-a")
```

## `Delete`

```text
Delete(id string) error
```

Removes the Session JSON file, returning a session-not-found error for a missing file and propagating other filesystem errors.

Example:

```text
fileStore.Delete("project-a")
```

## `List`

```text
List() []*Session
```

Scans the directory for `.json` files, skips directories and entries that fail to load, and returns the successfully loaded Sessions. If reading the directory fails, it returns an empty slice because the method has no error return.

Example:

```text
sessions := fileStore.List()
```

The internal `(*FileSessionStore).sessionPath(id string) string` helper joins the store directory with `<id>.json`. The package helper `ensureDir(dir string) error` creates the directory tree with mode `0700` and returns any filesystem error.

---

# Database Session Store

The database session store persists Sessions using SQLite.

The implementation uses `modernc.org/sqlite`, a pure-Go SQLite implementation.

SQLite is embedded and does not require a separate database server or network port.

`NewDatabaseSessionStore(path)` opens the SQLite database and initializes its tables.
Its signature is `NewDatabaseSessionStore(path string) (*DatabaseSessionStore, error)`. If initialization fails, it closes the opened database and returns the error.

It provides:

```text
Set

Get

Delete

List
```

The database store maintains two tables:

```text
sessions

messages
```

The internal `(*DatabaseSessionStore).initialize() error` enables SQLite foreign keys and creates the `sessions` and `messages` tables if they do not already exist.

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

Session messages are associated with their parent Session through a foreign key with cascade deletion.

Deleting a Session therefore also deletes its associated messages.

Database writes use transactions so Session metadata and messages are persisted consistently.

## Database Session Store Operations

### `Set`

```text
Set(session *Session) error
```

Rejects a nil Session or empty ID, then writes the Session and messages in a transaction. It serializes metadata and each message's tool calls as JSON. If any transactional operation fails, the deferred rollback prevents a partial write and the error is returned.

Example:

```text
databaseStore.Set(session)
```

The operation persists:

```text
Session ID

Creation timestamp

Update timestamp

Metadata

Messages

Tool calls
```

If the Session already exists, its stored metadata and `updated_at` are updated; the original `created_at` is retained by the SQL upsert.

The existing messages for that Session are deleted and replaced by the supplied message history, including serialized tool calls, within the same transaction.

### `Get`

```text
Get(id string) (*Session, error)
```

Retrieves a Session and ordered message history from SQLite using its ID. It returns a session-not-found error when absent, and propagates query, scan, or JSON decoding errors.

Example:

```text
session, err := databaseStore.Get("project-a")
```

The operation restores:

```text
Session ID

Creation timestamp

Update timestamp

Metadata

Messages

Tool calls
```

Messages are loaded in their stored order and reconstructed into the Session's message history.

If the Session does not exist, the store returns a Session-not-found error.

### `Delete`

```text
Delete(id string) error
```

Deletes a Session from SQLite using its ID. It returns a session-not-found error if no row was deleted, or a database error otherwise.

Example:

```text
databaseStore.Delete("project-a")
```

Because the `messages` table uses a foreign key with cascade deletion, deleting a Session also removes all messages associated with that Session.

If the Session does not exist, the store returns a Session-not-found error.

### `List`

```text
List() []*Session
```

Returns all Sessions stored in the SQLite database, ordered by creation time. As this method has no error return, query, scan, row-iteration, or session-load failures result in an empty slice.

Example:

```text
sessions := databaseStore.List()
```

Sessions are retrieved in creation order.

Each Session is reconstructed with its metadata and associated messages before being returned.

---

# Database Session Storage

The database-backed Session architecture is:

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

The database store uses transactions when persisting Session metadata and messages so both parts of the Session are updated together.

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

Session writes use atomic temporary-file replacement so the final Session file is not left with partially written JSON after a failed write.

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

When a new Session is created through the CLI, the newly created Session is also assigned to the Agent and becomes the active Session.

---

# Session Persistence

The application persists the active Session after successful Agent interactions.

For file-based persistence, the flow is:

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

For database-backed persistence, the flow is:

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

Creating a new session through the CLI also immediately activates the new Session:

```text
session create project-a
```

The resulting flow is:

```text
CLI

 │

 ▼

Create Session

 │

 ▼

Persist Session

 │

 ▼

Set Active Session

 │

 ▼

Assign Session to Agent

 │

 ▼

Continue Conversation
```

---

# Database Session Testing

Database session tests are located in:

```text
internal/session/database_store_test.go
```

The tests cover:

* Database initialization
* Session persistence
* Session metadata persistence
* Message persistence
* Session retrieval
* Session deletion
* Cascading message deletion
* Session listing

The tests also exercise nil/empty-ID rejection, missing-session errors, invalid stored metadata/tool-call JSON, database query failures, and rollback after a forced message-insert failure. The implementation serializes message tool calls, but the current database tests verify messages and metadata rather than asserting a tool-call round trip. File-store tests cover invalid inputs, filesystem and JSON errors, missing sessions, and temporary-file cleanup after a successful write. CLI tests cover rollback of MCP add/remove when saving the configuration fails.

Run the session tests with:

```bash
go test ./internal/session -v
```

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

To build and then run the CLI:

```bash
go build -o agent-harness ./cmd
./agent-harness
```

The CLI requires `config/config.json` (or `../config/config.json` when launched from a child directory). Create it from the example and update its Provider settings:

```bash
cp config/config.example.json config/config.json
```

The example configures the repository MCP test server, so either build that executable first or remove the example server entry before starting the CLI.

The CLI starts with the default Session:

```text
default
```

In `cmd/main.go`, `main()` loads configuration, connects configured MCP servers, creates the built-in tool registry and provider, restores the `default` file-backed Session, and runs the input loop. Important CLI functions are:

* `loadConfig() (*configpkg.Config, error)` tries `config/config.json` and then `../config/config.json`, returning the first valid configuration or the last load error.
* `resolveModel(configModel string) (string, string)` chooses a trimmed `OLLAMA_MODEL` environment override first, then the trimmed configured model, then the built-in default `kirito1/qwen3-coder:4b`. It also returns the source label (`environment`, `config`, or `default`).
* `selectModel(providerClient providerpkg.Provider, configuredModel string, configuredSource string, input *bufio.Scanner, output io.Writer) (string, string, error)` returns the configured choice unchanged for providers without `LocalModelManager`. For Ollama, it lists models, returns the configured model if installed, or prompts to pull it, select an installed model, or quit. Listing, pulling, and input errors are returned.
* `handleCommand(input string, model string, modelSource string, currentSession **sessionpkg.Session, sessionStore sessionpkg.Store, agentClient *agentpkg.Agent, appConfig *configpkg.Config, mcpRuntime *mcppkg.Runtime, registry *toolspkg.ToolRegistry, output io.Writer) CommandResult` parses and handles built-in commands, writes user-facing results to `output`, and returns `CommandNotHandled`, `CommandHandled`, or `CommandExit`. Session create/load update the active session and Agent; current-session deletion is refused. MCP add/remove update configuration and runtime together and attempt rollback when saving fails.

---

# CLI Commands

| Command                              | Description                            |
| ------------------------------------ | -------------------------------------- |
| `help`                               | Show available commands                |
| `model`                              | Show the selected model and its source  |
| `session`                            | Show the current session ID            |
| `session create <id>`                | Create and activate a new session      |
| `session list`                       | List stored sessions                   |
| `session load <id>`                  | Load and activate a stored session     |
| `session delete <id>`                | Delete a stored session                |
| `clear`                              | Clear the terminal display              |
| `mcp`                                | Show MCP command usage                 |
| `mcp list`                           | List configured MCP servers            |
| `mcp add <name> <command> [args...]` | Connect and persist an MCP server      |
| `mcp remove <name>`                  | Disconnect and remove an MCP server    |
| `exit`                               | Exit the CLI                           |

---

# `help`

Displays available CLI commands.

```text
help
```

Example:

```text
Available commands: help, exit, model, session, clear, mcp
```

---

# `model`

Displays the model selected for the current run and its source.

```text
model
```

For Ollama, the model can also be overridden using:

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

Creates and activates a new session:

```text
session create project-a
```

Example:

```text
Created and switched to session: project-a
```

The command performs the following operations:

1. Creates the Session.
2. Persists the Session to the Session Store.
3. Assigns the Session to the Agent.
4. Updates the active Session reference.
5. Continues the CLI using the new Session.

The Session ID remains the persistent identity of the Session.

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

The loaded Session becomes the active Session for subsequent interactions.

---

# `session delete`

Deletes a stored session:

```text
session delete project-a
```

The CLI prevents deletion of the currently active session.

---

# `clear`

Clears the terminal display; it does not clear or delete conversation history:

```text
clear
```

The active Session and its messages are unchanged.

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

The CLI loads the local configuration from `config/config.json` (or `../config/config.json` when started from a child directory). The local configuration:

```text
config/config.json
```

is ignored by Git because it may contain local development settings.

`Config` contains a `Provider` (`name`, `model`, `base_url`, and `endpoint`) and optional `MCP.Servers` entries (`name`, `command`, and `args`). `Load(path)` reads JSON and validates it; `Save(path, config)` writes indented JSON; `Validate()` checks that the required Provider values are non-empty.

---

# Configuration Operations

The configuration layer provides:

```text
Load

Save

Validate
```

## `Load`

```text
Load(path string) (*Config, error)
```

Reads the file at `path`, parses its JSON into a `Config`, and validates the required Provider fields. File-read, JSON-parse, and validation errors are returned.

Example:

```text
config, err := config.Load("config/config.json")
```

The configuration is parsed and validated before being returned.

## `Save`

```text
Save(path string, config *Config) error
```

Serializes `config` as indented JSON with a trailing newline and writes it to `path` (using mode `0644` if the file is created). Encoding and write errors are returned.

Example:

```text
err := config.Save("config/config.json", config)
```

The configuration is written as formatted JSON.

## `Validate`

```text
(*Config).Validate() error
```

Checks that Provider `Name`, `Model`, `BaseURL`, and `Endpoint` are non-empty. It returns the first missing-field error; it does not validate supported provider names or MCP server contents.

Example:

```text
err := config.Validate()
```

---

# Configuration Validation

The configuration layer validates that these Provider fields are non-empty:

```text
Provider Name

Provider Model

Provider Base URL

Provider Endpoint
```

MCP configuration is optional. `endpoint` is required by validation, but the current provider implementations construct their standard chat paths themselves (`/api/chat` for Ollama and `/chat/completions` for OpenAI); they do not use this configured field.

---

# Provider

The Provider interface keeps the Agent independent from a particular model provider.

The common `Provider` contract is:

```go
Chat(request ChatRequest) (ChatResponse, error)
```

`Chat` sends the requested model, messages, and optional tool definitions, then returns response text and normalized tool calls or an error. Ollama and OpenAI validate that a model and at least one message are supplied; provider-specific transport, status, response-decoding, and no-choice errors are returned.

## Provider Operations

### `Chat`

```text
Chat(request ChatRequest) (ChatResponse, error)
```

Sends a provider-neutral chat request and returns the provider response, including tool calls when supplied by the model.

Example:

```text
response, err := provider.Chat(request)
```

### `ListModels`

```text
(*OllamaProvider).ListModels() ([]string, error)
```

Uses Ollama's `/api/tags` endpoint to return locally available model names. It returns errors for connection, non-200 status, invalid JSON, or empty model names.

Example:

```text
models, err := ollamaProvider.ListModels()
```

The separate `LocalModelManager` interface requires `ListModels() ([]string, error)` and `PullModel(name string) error`; only Ollama implements it. `(*OllamaProvider).HasModel(name string) (bool, error)` calls `ListModels()` and checks for an exact match. At startup, the CLI uses model listing and pulling to check the configured Ollama model, offer installed alternatives, or pull it. OpenAI does not implement local model management.

Example:

```text
err := ollamaProvider.PullModel("llama3.1")
```

---

# Ollama

Ollama is the primary local Provider.

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

At startup, the CLI lists local models. If the configured model (or `OLLAMA_MODEL` override) is missing, it lets the user pull that model, choose an installed model, or quit.

Then start Agent Harness:

```bash
go run ./cmd
```

---

# OpenAI

OpenAI is supported through the Provider abstraction.

The OpenAI provider is implemented in:

```text
internal/provider/openai.go
```

The provider uses the configured API base URL (default `https://api.openai.com/v1`) and API credential, then posts to `/chat/completions`.

API credentials are retrieved through the credential store rather than being stored directly in the Provider configuration.

This keeps credential handling separate from provider-specific request logic.

The CLI looks up the provider key in `~/.agent-harness/credentials.json`. The file store's `Set`, `Get`, and `Delete` methods manage credentials; the interactive CLI has no credential-management command.

## OpenAI Operations

### `Chat`

Sends a chat request to the configured OpenAI-compatible API endpoint and converts the response into the common Provider response type.

Example:

```text
provider.Chat(request)
```

### Tool Call Handling

Converts model-generated tool calls into the common:

```text
ToolCall
```

representation used by the Agent runtime.

Example flow:

```text
OpenAI Response

      │

      ▼

Tool Call

      │

      ▼

Common ToolCall

      │

      ▼

Agent Runtime
```

---

# Provider Factory

The Provider factory is implemented in:

```text
internal/provider/factory.go
```

It creates providers based on the configured provider name.

The main operation is:

```text
NewProvider
```

### `NewProvider`

```text
NewProvider(name, baseURL, apiKey string) (Provider, error)
```

Creates `*OllamaProvider` or `*OpenAIProvider` with the supplied base URL and, for OpenAI, API key. Returns an error for unsupported provider names.

Example:

```text
provider, err := NewProvider(
    "ollama",
    "http://localhost:11434",
    "",
)
```

Currently supported providers are:

```text
ollama

openai
```

The factory also provides:

```text
RequiresAPIKey
```

### `RequiresAPIKey`

```text
RequiresAPIKey(name string) bool
```

Reports whether `name` requires a stored API credential before CLI initialization. It returns true for OpenAI and false for Ollama or unknown names.

Example:

```text
RequiresAPIKey("openai")
```

Currently:

```text
ollama → false

openai → true
```

---

# Model Override

The configured Ollama model can be overridden using:

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

`NewOrchestrator(agentClient *agentpkg.Agent, providerClient providerpkg.Provider, options ...OrchestratorOption) *DefaultOrchestrator` creates the orchestrator, prints a creation message, and sets the default tool-call limit to 10 before applying the options. `WithMaxToolCalls(maxToolCalls int) OrchestratorOption` changes that limit only when the supplied value is positive; zero or a negative value leaves the current limit unchanged.

This prevents uncontrolled tool execution loops.

### Orchestrator Operation

The main conversational operation:

```text
RunAgent(request providerpkg.ChatRequest) (AgentResponse, error)
```

Requires at least one request message, uses only the last request message as the new user message, adds it to Agent history, and replaces request tools/messages with registry definitions and full conversation history. It then calls the Provider and executes native tool calls (including multiple calls per response) until a final response or error. Legacy JSON responses containing a `tool_call` are also executed. It returns errors for empty requests, schema/provider/tool failures, valid JSON objects that cannot be decoded as an `AgentResponse`, or exceeding the configured tool-call limit; conversation messages are appended as the loop proceeds.

Example:

```text
response, err := orchestrator.RunAgent(request)
```

Other `DefaultOrchestrator` methods:

* `Chat(request provider.ChatRequest) (provider.ChatResponse, error)` forwards one request to the Provider and returns its response/error; it also prints a request-status line.
* `Run(name string, args map[string]any) (any, error)` directly executes the named tool through the Agent and prints a tool-status line.
* `AssignTool(toolCall ToolCall) (any, error)` executes the named tool with the supplied arguments and wraps execution errors with the tool name.
* `GetToolSchemas() ([]map[string]any, error)` returns the Agent registry schemas.
* `GetToolDefinitions() ([]provider.ToolDefinition, error)` converts registry schemas to Provider definitions; it errors if a name, description, or parameter schema is absent or invalid.

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
                │ OpenAI          │    │ Shell           │
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

Current storage backends are:

```text
In-Memory

File

Database
```

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

│   │   ├── factory.go

│   │   ├── factory_test.go

│   │   ├── openai.go

│   │   ├── openai_test.go

│   │   ├── ollama.go

│   │   ├── ollama_test.go

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

│       ├── calculator.go

│       ├── calculator_test.go

│       ├── filesystem.go

│       ├── filesystem_test.go

│       ├── registry.go

│       ├── registry_test.go

│       ├── shell.go

│       ├── shell_test.go

│       ├── tool.go

│       └── tool_test.go

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

is local build output and is intentionally not included in the repository.

The runtime-created session directory:

```text
.sessions/
```

is also local application data and is not part of the repository structure.

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

Run session tests:

```bash
go test ./internal/session -v
```

Run database session tests:

```bash
go test ./internal/session -run Database -v
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

The provider's direct Ollama integration test can be enabled separately:

```bash
OLLAMA_INTEGRATION=1 go test ./internal/provider -run TestOllamaProvider_Integration
```

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

GitHub Actions is used to validate the repository automatically.

The current workflow performs:

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

and for pull requests targeting:

```text
main
```

The MCP test server is built during CI using:

```bash
go build -o cmd/mcp-test-server/mcp-test-server ./cmd/mcp-test-server
```

This ensures the command-transport MCP tests have the required test-server binary available.

---

# Development Workflow

Create a feature branch:

```bash
git switch -c feature/<name>
```

Make changes and format the project:

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
git push origin feature/<name>
```

Then open a pull request against `main`.

---

# Architectural Principles

## Provider Neutrality

The Agent should not depend directly on Ollama-specific APIs or OpenAI-specific APIs.

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

Current storage backends include:

```text
In-Memory

File

Database
```

This allows different storage implementations to be introduced without rewriting Agent execution.

Future storage backends may include:

```text
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

The architecture also supports cloud providers through the same Provider abstraction.

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
```

---

# Implementation Status

| Component                             | Status      |
| ------------------------------------- | ----------- |
| Go project structure                  | Implemented |
| Provider abstraction                  | Implemented |
| Provider factory                      | Implemented |
| Ollama provider                       | Implemented |
| OpenAI provider                       | Implemented |
| Credential-aware provider creation    | Implemented |
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
| Additional providers                  | Planned     |

---

# Roadmap

## Improved MCP Support

The basic MCP integration, CLI server management, and dynamic MCP server lifecycle are implemented.

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

## Additional Providers

The Provider abstraction allows future integrations such as:

```text
Anthropic

Google

Other Local Model Runtimes

Additional OpenAI-compatible APIs
```

Provider implementations can be added without changing the core Agent architecture.

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

MCP Server Management is no longer listed here because basic MCP server management and dynamic runtime operations have already been implemented.

Future CLI improvements may build on the existing MCP commands with richer inspection capabilities.

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

Ollama

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

A session creation flow looks like:

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
