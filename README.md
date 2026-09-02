# Agent Harness

An agent harness is a runtime framework that connects an LLM to external tools and capabilities. It provides a layer where an agent can:

- **Receive instructions** from a user or an application.
- **Plan and orchestrate** steps to complete a task.
- **Invoke tools** (e.g. calculators, file access, shell commands) to gather data or take actions.
- **Return results** back to the caller in a structured way.

Think of it as the scaffolding that turns a language model into an autonomous agent: the model supplies the intelligence, while the harness supplies the environment, the tool interface, and the execution loop that ties it together.

## Architecture

The Agent Harness is built around a **layered tool-execution architecture** that emphasizes separation of concerns and extensibility:

```
Agent
  ↓
Orchestrator (Planned)
  ↓
Tool Registry
  ↓
Tool Interface
  ↓
Concrete Tools
  ├── Calculator (Implemented)
  ├── Shell (Planned)
  └── Future tools (Search, File, etc.)
```

### Architecture Layers and Responsibilities

#### 1. Agent (Planned)
- Represents the high-level AI agent interface.
- Determines what tasks need to be accomplished based on user instructions.
- **Does NOT** directly implement or manage individual tools.
- Communicates with the Orchestrator to request task execution.
- **Status**: Planned for future implementation.

#### 2. Orchestrator (Planned)
- Acts as the central coordinator of the agent workflow.
- Receives the agent's requested action and analyzes what needs to be done.
- Determines which registered tool should be used for the task.
- Retrieves the required tool from the Tool Registry.
- Passes the appropriate arguments to the tool's `Execute()` method.
- Handles the tool's result or error and returns it to the agent.
- **Key distinction**: The Orchestrator is responsible for **coordinating execution** and **workflow decisions**, while the Registry is only responsible for **managing tools**.
- **Status**: Planned for future implementation.

#### 3. Tool Registry (Implemented)
Maintains a centralized collection of available tools. It acts as the directory/lookup layer that allows the Orchestrator to find and retrieve tools by name.

**Responsibilities**:
- Maintains a mapping of tool names to Tool implementations.
- Does NOT decide which tool to execute—that is the Orchestrator's responsibility.
- Provides standardized operations for tool management.

**Implemented Methods**:

- **`NewToolRegistry() *ToolRegistry`**
  - Constructor that initializes an empty Tool Registry.
  - Returns a new `ToolRegistry` instance with an empty tools map.
  - Usage: Call once at application startup to create the registry.

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

#### 4. Tool Interface (Implemented)
Defines the common contract that all tools must implement. This allows the Registry and Orchestrator to work with any tool without coupling to concrete implementations.

**Defined Interface** (`tools.go`):
```go
type Tool interface {
    Name() string
    Description() string
    Execute(args map[string]any) (any, error)
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

#### 5. Concrete Tools (Partially Implemented)

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
    - `"numbers"` ([]float64): The operands for the operation.
  - Returns: The numeric result, or an error if inputs are invalid or division by zero occurs.
  - Example: `calculator.Execute(map[string]any{"operation": "add", "numbers": []float64{2, 3, 5}})`

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
- **Shell**: Will execute shell commands and scripts.
- **Search**: Will search for information across data sources.
- **File**: Will handle file operations (read, write, delete).
- Additional tools can be added without modifying the Registry or Orchestrator.

### Current Package Structure

```
agent-harness/
├── cmd/                              (Planned: application entry points)
├── internal/
│   ├── tools/                        (Tool implementation and management)
│   │   ├── tools.go                  (Tool interface definition)
│   │   ├── ToolRegistry.go           (Tool Registry implementation)
│   │   ├── calculator.go             (Calculator tool implementation)
│   │   ├── calculator_test.go        (Calculator unit tests)
│   │   ├── registry_test.go          (Registry unit tests)
│   │   ├── registry_test_has.go      (Registry.Has() unit tests)
│   │   ├── registry_test_list.go     (Registry.List() unit tests)
│   │   └── registry_test_remove.go   (Registry.Remove() unit tests)
│   ├── agent/                        (Planned: Agent implementation)
│   └── orchestrator/                 (Planned: Orchestrator implementation)
├── shell/                            (Planned: Shell tool package)
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
| Tool Interface | ✅ Implemented | Defines `Name()`, `Description()`, `Execute()` |
| Tool Registry | ✅ Implemented | Full CRUD operations: `Register`, `Get`, `Has`, `List`, `Remove` |
| Calculator Tool | ✅ Implemented | Supports `add`, `subtract`, `multiply`, `divide`, `modulus` operations |
| Orchestrator | 🔄 Planned | Will coordinate tool execution and workflow |
| Agent | 🔄 Planned | Will provide high-level task interface |
| Shell Tool | 🔄 Planned | Will execute shell commands |
| File Tool | 🔄 Planned | Will handle file operations |
| Search Tool | 🔄 Planned | Will search across data sources |