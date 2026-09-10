# Agent Harness

An agent harness is a runtime framework that connects an LLM to external tools and capabilities. It provides a layer where an agent can:

- **Receive instructions** from a user or an application.
- **Plan and orchestrate** steps to complete a task.
- **Invoke tools** (e.g. calculators, file access, shell commands) to gather data or take actions.
- **Return results** back to the caller in a structured way.

Think of it as the scaffolding that turns a language model into an autonomous agent: the model supplies the intelligence, while the harness supplies the environment, the tool interface, and the execution loop that ties it together.

## Recent Progress

- **Today**: Added a provider abstraction with a provider-neutral interface so the system can work across multiple LLM providers without hard-coding provider logic into the agent/orchestrator layers.
- **Today**: Added Ollama provider support, including request validation, HTTP chat calls, JSON encoding/decoding, and test-friendly dependency injection.
- **Today**: Added provider-neutral tool definitions and schemas so the tool registry and agent can expose capabilities in a format that providers can consume.
- **Today**: Implemented native Ollama function/tool calling, including Ollama-specific request conversion and tool-call decoding.
- **Today**: Added support for multiple native tool calls in a single provider response and updated the orchestrator to execute each requested tool and return their results to the provider.
- **Today**: Added configurable maximum tool-call limits to protect the orchestration loop from runaway tool execution.
- **Today**: Added the Calculator, Shell, and Filesystem tools to the Tool Registry and exposed JSON schema support for all registered tools.
- **Today**: Updated the orchestrator to work with provider-neutral tool calls and native tool execution flow without tightly coupling to Ollama-specific types.
- **Today**: Added comprehensive unit tests for provider, agent, tools, registry, and orchestrator behavior, including single and multiple native tool-call cases.
- **Today**: Added an opt-in Ollama integration test for end-to-end tool calling when a local provider is available.
- **Today**: Verified the end-to-end flow: User → Agent → Orchestrator → Provider → Native Tool Call → Tool Registry → Tool → Result → Provider → Final Response.
- **Later**: Additional tools will be added after the core v1 is stabilized and the orchestration pattern is validated.

- **Yesterday**: Added the low-level shell execution logic in `shell/shell.go`. It validates commands with `exec.LookPath`, executes them with `os/exec`, captures combined output, and returns `(string, error)`.
- **Today**: Added the `ShellTool` and `FilesystemTool`, exposing shell commands and local filesystem operations through the common `Tool` interface.
- **Today**: Added the initial Agent logic in `internal/agent/agent.go`. The Agent accepts a Tool Registry, retrieves tools by name, forwards argument maps, and returns tool results or errors.
- **Today**: Added `Agent.Run`, which provides the public entry point for executing a named registered tool.
- **Today**: Implemented the Ollama Provider in `internal/provider/provider.go`, including request validation, JSON encoding, HTTP requests, response decoding, and injectable HTTP dependencies for testing.
- **Today**: Added provider unit tests with a local `httptest` server and an Ollama integration test for the local service.
- **Today**: Added the Orchestrator layer in `internal/orchestrator`, connecting the Agent and Provider layers through `Run`, `Chat`, and `AssignTool`.
- **Today**: Added the `ToolCall` and `AgentResponse` contracts with JSON tags for structured LLM responses.
- **Today**: Added `RunAgent`, which accepts plain text or structured JSON provider responses, executes requested tool calls, and returns tool results.
- **Today**: Added comprehensive Orchestrator tests for direct tool execution, provider chat, tool assignment, plain text, JSON tool calls, invalid JSON, and tool execution through `RunAgent`.
- **Today**: Updated the Calculator to accept both direct `[]float64` arguments and JSON-decoded numeric arrays.
- **Today**: Added native Ollama tool calling support with provider-neutral `ToolDefinition` values and Ollama-specific request conversion.
- **Today**: Added JSON Schema support for tool parameters and updated the Calculator, ShellTool, and FilesystemTool schemas to proper JSON Schema.
- **Today**: Updated `RunAgent()` with native provider tool-call handling, tool execution through the Orchestrator, tool-result feedback to the provider, and maximum tool-call protection through `WithMaxToolCalls`.
- **Today**: Added and updated tests for Ollama request conversion, native tool-call conversion, tool schemas, Agent/Orchestrator tool definitions, and Ollama integration. Native Ollama tool calling has been tested end-to-end with the local Ollama integration test.

### Current Status

The provider boundary is now provider-neutral. The shared `Provider` interface and generic `Message`, `ChatRequest`, `ChatResponse`, `ToolCall`, and `ToolDefinition` types are used by the orchestrator, while Ollama-specific request and response types remain inside the provider package.

The provider converts generic messages and tool definitions to Ollama's API format and converts native Ollama tool calls back to generic tool calls. The orchestrator executes those calls through the Agent, returns tool results to the provider, and continues until a final response is produced or the tool-call limit is reached.

The full unit-test suite currently passes:

```bash
go test ./...
```

### Verified local Ollama setup

The CLI is configured to use the installed local model by default:

```bash
kirito1/qwen3-coder:4b
```

You can override it at runtime with:

```bash
OLLAMA_MODEL="your-model-name" go run ./cmd
```

The real end-to-end orchestrator integration test is opt-in and requires the local Ollama service to be running:

```bash
ORCHESTRATOR_INTEGRATION=1 go test -run TestOrchestratorRunAgent_Integration ./internal/orchestrator -v
```

## Architecture

The Agent Harness is built around a **layered tool-execution architecture** that emphasizes separation of concerns and extensibility:

```
Config → Provider → Agent → Orchestrator → Tool Registry → Tools
```

Configuration is loaded from `config/config.json`, the Provider communicates with the LLM, the Agent exposes registered tool capabilities, the Orchestrator coordinates requests and tool calls, the Tool Registry manages available tools, and the Tools perform concrete operations.

### Architecture Layers and Responsibilities

#### 1. Config Layer (Implemented)
- Manages application and provider configuration.
- Stores configuration data in JSON format in the `config/` folder.
- Does NOT make API calls or interact with external services.
- Provides configuration settings that will be consumed by the Provider layer.
- **Implemented**: `config/config.go` defines the configuration structures, loading logic, and validation.
- **Tests**: `config/config_test.go` covers loading valid and missing files, plus required-field validation.
- **Configuration includes**:
  - Provider name (e.g., "ollama")
  - Model name (e.g., "llama3.1")
  - Base URL for the provider (e.g., "http://localhost:11434")
  - API endpoint (e.g., "/api/chat")

**Configuration File** (`config/config.json`):
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

**Config API** (in `config/config.go`):
- **`Load(path string) (*Config, error)`** - Read, parse, and validate a JSON configuration file.
- **`Validate() error`** - Validate the required provider fields: name, model, base URL, and endpoint.

Run the config package tests from the repository root with:

```bash
go test -v ./config
```

#### 2. Provider Layer (Implemented)
- Manages connections to LLM providers (e.g., Ollama, OpenAI, etc.).
- The Ollama implementation sends chat requests to `/api/chat` and decodes the returned assistant message.
- Supports native Ollama tool calling by converting provider-neutral `ToolDefinition` values into Ollama function-tool request definitions.
- Validates that a model and at least one message are provided before making a network request.
- Supports configurable `BaseURL` and `http.Client` values so tests do not need a running external service.
- Acts as the bridge between the Agent and external AI services.
- **Implemented types**: `Provider`, `ToolDefinition`, `ChatRequest`, `Message`, `ChatResponse`, Ollama request/response tool-call types, and `OllamaProvider`.
- **Tool parameters**: Provider-neutral definitions expose JSON Schema through `ToolDefinition.Parameters`.
- **Tests**: `internal/provider/provider_test.go` covers request validation, Ollama request conversion, and the provider response path using `httptest`.
- **Integration test**: `internal/provider/provider_integration_test.go` verifies communication with a local Ollama instance.

Run provider unit tests from the repository root with:

```bash
go test -v ./internal/provider
```

Run the Ollama integration test when a local Ollama service is running:

```bash
go test -v -run TestOllamaProvider_Integration ./internal/provider
```

Run the orchestrator integration test when Ollama and the configured model are available:

```bash
ORCHESTRATOR_INTEGRATION=1 go test -run TestOrchestratorRunAgent_Integration ./internal/orchestrator -v
```

#### 3. Agent (Implemented)
- Represents the high-level AI agent interface.
- Receives a Tool Registry during construction.
- Retrieves a named tool and forwards the provided arguments to its `Execute()` method.
- `GetToolSchemas()` returns the registered tool names, descriptions, and argument schemas, or an error when the registry contains invalid tool metadata.
- `GetToolDefinitions()` converts registered tool schemas into provider-neutral `ToolDefinition` values for the Provider layer.
- `Run(name, args)` is the public execution entry point and delegates to `ExecuteTool`.
- Returns the tool result or execution error to the caller.
- **Tests**: `internal/agent/agent_test.go` covers agent construction, calculator and shell execution, unknown tools, `Run`, and tool schema retrieval.
- **Status**: Tool-execution logic and tool-definition exposure implemented.

#### 4. Orchestrator (Implemented)
- Acts as the central coordinator of the agent workflow.
- `Run(name, args)` executes a named tool through the Agent.
- `Chat(request)` forwards chat requests to the configured Provider.
- `AssignTool(toolCall)` executes a structured `ToolCall` through the Agent.
- `RunAgent(request)` sends a request to the Provider, parses responses, handles native provider tool calls, executes requested tools through the Agent, and sends tool results back to the Provider.
- Supports both plain-text/structured JSON responses and native Ollama tool-call responses.
- Applies maximum tool-call protection through `WithMaxToolCalls`.
- Converts registered tool schemas into provider-neutral tool definitions before sending requests.
- Uses `json.Valid()` before unmarshalling structured responses.
- Wraps provider, parsing, and tool execution errors with context.
- **Key distinction**: The Orchestrator is responsible for **coordinating execution** and **workflow decisions**, while the Registry is only responsible for **managing tools**.
- **Status**: Implemented, including native provider tool-call execution and tool-result feedback.

#### 5. Agent Response Contract (Implemented)
- Defined in `internal/orchestrator/response.go`.
- `ToolCall` contains a tool name and argument map.
- `AgentResponse` contains response content and an optional tool call.
- JSON tags map structured provider responses to `content` and `tool_call`.

#### 6. Agent Loop (Implemented)
- Flow: Provider response -> native tool call or structured `ToolCall` -> execute tool -> send the tool result back to the Provider -> final response.
- `RunAgent` repeats the loop until the Provider returns a final response or the configured maximum tool-call limit is reached.

#### 7. Tool Registry (Implemented)
Maintains a centralized collection of available tools. It acts as the directory/lookup layer that allows the Orchestrator to find and retrieve tools by name.

**Responsibilities**:
- Maintains a mapping of tool names to Tool implementations.
- Does NOT decide which tool to execute—that is the Orchestrator's responsibility.
- Provides standardized operations for tool management.

**Implemented Methods**:

- **`NewToolRegistry() *ToolRegistry`**
  - Constructor that initializes a Tool Registry with the calculator, shell, and filesystem tools.
  - Usage: Call once at application startup to create the registry with its default tools.

- **`Register(name string, tool Tool) (string, Tool)`**
  - Registers a new tool in the registry.
  - Parameters:
    - `name`: The identifier for the tool (e.g., "calculator").
    - `tool`: The Tool implementation to register.
  - Returns: The tool name and the registered Tool instance.
  - Usage: Register each tool when the application starts (e.g., `registry.Register("calculator", Calculator{})`).

- **`Get(name string) (Tool, error)`**
  - Retrieves a registered tool by name.
  - Parameters:
    - `name`: The identifier of the tool to retrieve.
  - Returns: The Tool instance if found, or an error if not found.
  - Error: Returns `fmt.Errorf("tool %q not found", name)` if the tool does not exist.
  - Usage: Called by the Orchestrator to get a tool before execution.

- **`Has(name string) bool`**
  - Checks whether a tool with the given name is registered.
  - Parameters:
    - `name`: The identifier to check.
  - Returns: `true` if the tool exists, `false` otherwise.
  - Usage: Used for validation before attempting to execute a tool.

- **`List() []string`**
  - Retrieves a list of all registered tool names.
  - Returns: A slice of strings containing all registered tool names.
  - Usage: Used to display available tools or for tool discovery.

- **`Remove(name string) error`**
  - Removes a registered tool from the registry.
  - Parameters:
    - `name`: The identifier of the tool to remove.
  - Returns: `nil` on success, or an error if the tool does not exist.
  - Error: Returns `fmt.Errorf("tool %q not found", name)` if the tool is not registered.
  - Usage: Used to deregister tools dynamically at runtime.

- **`Schemas() ([]map[string]any, error)`**
  - Returns the name, description, and JSON Schema argument definition for every registered tool.
  - Returns an error if the registry is nil, a registered tool is nil, or a tool returns a nil schema.
  - Usage: Used to expose tool capabilities to callers such as an LLM provider.

#### 8. Tool Interface (Implemented)
Defines the common contract that all tools must implement. This allows the Registry and Orchestrator to work with any tool without coupling to concrete implementations.

**Defined Interface** (`tools.go`):
```go
type Tool interface {
    Name() string
    Description() string
    Execute(args map[string]any) (any, error)
  Schema() map[string]any
}
```

**Methods every Tool must implement**:

- **`Name() string`**
  - Returns the tool's identifier.
  - Usage: Used by the Registry and Orchestrator to refer to the tool.

- **`Description() string`**
  - Returns a human-readable description of what the tool does.
  - Usage: Displayed to users or agents to explain tool capabilities.

- **`Execute(args map[string]any) (any, error)`**
  - Executes the tool's core operation with the provided arguments.
  - Parameters:
    - `args`: A map of argument names to values. The specific keys and types depend on the tool.
  - Returns: The result of the operation, or an error if execution fails.
  - Usage: Called by the Orchestrator to perform the requested task.

- **`Schema() map[string]any`**
  - Returns the tool's arguments as a JSON Schema object.
  - Usage: Used by `ToolRegistry.Schemas()` to describe available tools.

#### 9. Concrete Tools (Partially Implemented)

Concrete tools implement the Tool interface and perform actual operations. Each tool encapsulates its own execution logic and validation.

##### Calculator Tool (Implemented)

The Calculator is the first concrete tool implementation. It performs basic arithmetic operations.

**Tool Information**:
- **Name**: `"calculator"`
- **Description**: `"Performs basic arithmetic operations"`

**Interface Implementation**:
- **`Name() string`**: Returns `"calculator"`
- **`Description() string`**: Returns `"Performs basic arithmetic operations"`
- **`Execute(args map[string]any) (any, error)`**: Routes to the appropriate arithmetic operation based on the `args` map:
  - Expected `args`:
    - `"operation"` (string): The operation to perform (`"add"`, `"multiply"`, `"subtract"`, `"divide"`, `"modulus"`).
    - `"numbers"` ([]float64 or JSON-decoded numeric array): The operands for the operation.
  - Returns: The numeric result, or an error if inputs are invalid or division by zero occurs.
  - Example: `calculator.Execute(map[string]any{"operation": "add", "numbers": []float64{2, 3, 5}})`
- **`Schema() map[string]any`**: Returns a JSON Schema object for the `operation` and `numbers` parameters.
- Input validation rejects nil arguments, missing or empty operations, invalid number arrays, and empty number arrays.

**Arithmetic Methods**:
- **`Add(numbers ...float64) (float64, error)`**
  - Sums all provided numbers.
  - Usage: `calculator.Add(2, 3, 5)` → `10.0`

- **`Subtract(numbers ...float64) (float64, error)`**
  - Subtracts all numbers from the first number sequentially.
  - Usage: `calculator.Subtract(10, 3, 2)` → `5.0`

- **`Multiply(numbers ...float64) (float64, error)`**
  - Multiplies all provided numbers together.
  - Usage: `calculator.Multiply(2, 3, 4)` → `24.0`

- **`Divide(numbers ...float64) (float64, error)`**
  - Divides the first number by all subsequent numbers sequentially.
  - Checks for division by zero and returns an error if encountered.
  - Usage: `calculator.Divide(100, 2, 5)` → `10.0`

- **`Modulus(a, b float64) (float64, error)`**
  - Returns the remainder of `a` divided by `b`.
  - Checks for division by zero and returns an error if `b` is 0.
  - Usage: `calculator.Modulus(10, 3)` → `1.0`

##### Future Tools (Planned)
- **Search**: Will search for information across data sources.
- Additional tools can be added without modifying the Registry or Orchestrator.

##### File System Tool (Implemented)

The File System package provides local file and directory operations.

**Implementation**:
- **`filesystem/filesystem.go`**: Defines the `FileSystem` type and its file operation methods.
- **`Read(path string) ([]byte, error)`**: Reads a file and returns its contents.
- **`Write(path string, data []byte) error`**: Creates or overwrites a file.
- **`Remove(path string) ([]string, error)`**: Lists the entries in a directory.
- **`Search(path string, pattern string) ([]string, error)`**: Recursively searches for files matching a pattern.
- **`Delete(path string) error`**: Deletes a file or empty directory.
- **Tests**: `filesystem/filesystem_test.go` covers read, write, listing, search, and delete scenarios.

Run the File System package tests from the repository root with:

```bash
go test -v ./filesystem
```

##### Shell Tool (Implemented)

The Shell package executes commands through the operating system and reports clear errors for missing commands or failed execution.

**Implementation**:
- **`shell/shell.go`**: Defines `Shell.Execute(command string, args ...string) error`.
- Uses `exec.LookPath` to validate that a command is available before execution.
- Captures combined command output and returns it with any execution error.
- **`Shell.Execute(command string, args ...string) (string, error)`**: Executes a command and returns its combined output.
- **Tests**: `shell/shell_test.go` covers successful `echo` and `pwd` commands, missing commands, and failed commands.

Run the Shell package tests from the repository root with:

```bash
go test -v ./shell
```

### Current Package Structure

```
agent-harness/
├── cmd/                              (Planned: application entry points)
├── config/                           (Configuration layer)
│   ├── config.go                     (Configuration types, loading, and validation)
│   ├── config.json                   (Provider configuration)
│   └── config_test.go                (Configuration unit tests)
├── internal/
│   ├── tools/                        (Tool implementation and management)
│   │   ├── tools.go                  (Tool interface definition)
│   │   ├── ToolRegistry.go           (Tool Registry implementation)
│   │   ├── calculator.go             (Calculator tool implementation)
│   │   ├── calculator_test.go        (Calculator unit tests)
│   │   ├── filesystem_tool.go        (Filesystem tool implementation)
│   │   ├── filesystem_tool_test.go   (Filesystem tool unit tests)
│   │   ├── shell_tool.go             (Shell tool implementation)
│   │   ├── shell_tool_test.go        (Shell tool unit tests)
│   │   └── registry_test.go          (Registry unit tests)
│   ├── agent/                        (Initial Agent implementation)
│   │   ├── agent.go                  (Agent tool execution logic)
│   │   └── agent_test.go             (Agent unit tests)
│   ├── orchestrator/                 (Orchestrator implementation)
│   │   ├── orchestrator.go           (Orchestrator workflow and tool execution)
│   │   ├── orchestrator_test.go      (Orchestrator unit tests)
│   │   ├── response.go               (AgentResponse and ToolCall contracts)
│   │   └── response_test.go          (Response parsing tests)
│   └── provider/                     (Ollama Provider and provider tests)
│       ├── provider.go               (Provider interface and Ollama implementation)
│       ├── provider_test.go           (Provider unit tests)
│       └── provider_integration_test.go (Local Ollama integration test)
├── shell/                            (Shell tool package)
│   ├── shell.go                       (Shell command execution)
│   └── shell_test.go                  (Shell unit tests)
├── filesystem/                       (File system tool package)
│   ├── filesystem.go                  (File and directory operations)
│   └── filesystem_test.go             (File system unit tests)
├── go.mod                            (Go module definition)
├── go.sum                            (Go module checksums)
└── README.md                         (This file)
```

### Architectural Principles

#### Separation of Concerns
- **Agent** = Decides/requests what needs to happen.
- **Orchestrator** = Coordinates the workflow and execution (determines *how* and *when*).
- **Tool Registry** = Stores and retrieves available tools (manages the tool directory).
- **Tool Interface** = Defines the contract all tools must follow.
- **Concrete Tools** = Perform the actual operations.

#### Extensibility
The architecture is intentionally designed to be extensible. New tools can be registered and used without modifying the core tool-management layer:
- Implement the `Tool` interface with new `Name()`, `Description()`, and `Execute()` methods.
- Register the new tool using `registry.Register("tool-name", NewTool())`.
- The Orchestrator can immediately use the new tool without any changes.

#### Type Safety Through Interfaces
The `Tool` interface allows tools to be added and retrieved without type coupling. The Orchestrator and Registry depend on the interface, not on concrete implementations. This design makes it easy to:
- Add new tools without recompiling core logic.
- Test tools in isolation.
- Replace tool implementations without affecting orchestration logic.

### Implementation Status

| Component | Status | Details |
|-----------|--------|---------|
| Config Layer | ✅ Implemented | `config/config.go` loads and validates provider configuration from `config/config.json` |
| Provider Layer | ✅ Implemented | Ollama chat provider with native tool calling, provider-neutral tool definitions, JSON Schema parameters, request conversion, validation, HTTP requests, response decoding, and testable dependencies |
| Tool Interface | ✅ Implemented | Defines `Name()`, `Description()`, `Execute()`, and `Schema()` |
| Tool Registry | ✅ Implemented | Tool registration, lookup, removal, listing, and schema discovery with error handling |
| Calculator Tool | ✅ Implemented | Supports `add`, `subtract`, `multiply`, `divide`, `modulus` operations |
| Agent | ✅ Implemented | Executes named tools through `Run` and `ExecuteTool`, and exposes tool schemas and definitions |
| Orchestrator | ✅ Implemented | Runs tools, converts tool definitions, handles native provider tool calls, loops with tool results, and enforces `WithMaxToolCalls` |
| AgentResponse | ✅ Implemented | Defines structured response content and optional tool calls |
| Tool execution from RunAgent | ✅ Implemented | Executes requested native or structured tool calls and sends results back to the Provider |
| Agent loop | ✅ Implemented | Handles native provider tool calls, tool results, final responses, and maximum tool-call protection |
| Tool schemas | ✅ Implemented | Calculator, ShellTool, and FilesystemTool expose proper JSON Schema parameter definitions |
| Shell Tool | ✅ Implemented | Executes shell commands, reports command or execution failures, and exposes JSON Schema parameters |
| File System Tool | ✅ Implemented | Reads, writes, lists, searches, deletes local files and directories, and exposes JSON Schema parameters |
| Search Tool | 🔄 Planned | Will search across data sources |