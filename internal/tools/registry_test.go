package tools

import (
	"reflect"
	"testing"
)

func TestToolRegistryGet(t *testing.T) {
	testCases := []struct {
		name        string
		toolName    string
		register    bool
		expectError bool
	}{
		{
			name:     "Registered tool",
			toolName: "custom",
			register: true,
		},
		{
			name:        "Missing tool",
			toolName:    "missing",
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := &ToolRegistry{
				tools: make(map[string]Tool),
			}

			if testCase.register {
				registry.Register(testCase.toolName, &Calculator{})
			}

			actualTool, err := registry.Get(testCase.toolName)

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if actualTool != nil {
					t.Fatalf("expected no tool, got %T", actualTool)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if actualTool == nil {
				t.Fatal("expected tool, got nil")
			}
		})
	}
}

func TestToolRegistryNew(t *testing.T) {
	testCases := []struct {
		name     string
		toolName string
	}{
		{
			name:     "Calculator",
			toolName: "calculator",
		},
		{
			name:     "Shell",
			toolName: "shell",
		},
		{
			name:     "Filesystem",
			toolName: "filesystem",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := NewToolRegistry()

			if !registry.Has(testCase.toolName) {
				t.Errorf(
					"expected %q to be registered",
					testCase.toolName,
				)
			}
		})
	}
}

func TestToolRegistryRegister(t *testing.T) {
	testCases := []struct {
		name string
		tool Tool
	}{
		{
			name: "Register calculator",
			tool: &Calculator{},
		},
		{
			name: "Register shell",
			tool: &ShellTool{},
		},
		{
			name: "Register filesystem",
			tool: &FilesystemTool{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := &ToolRegistry{
				tools: make(map[string]Tool),
			}

			name, returnedTool := registry.Register(
				"tool",
				testCase.tool,
			)

			if name != "tool" {
				t.Fatalf(
					"expected name %q, got %q",
					"tool",
					name,
				)
			}

			if returnedTool != testCase.tool {
				t.Fatalf(
					"expected returned tool %T, got %T",
					testCase.tool,
					returnedTool,
				)
			}

			if !registry.Has("tool") {
				t.Fatal("expected tool to be registered")
			}
		})
	}
}

func TestToolRegistryHas(t *testing.T) {
	testCases := []struct {
		name      string
		register  bool
		expectHas bool
	}{
		{
			name:      "Registered tool",
			register:  true,
			expectHas: true,
		},
		{
			name:      "Unregistered tool",
			register:  false,
			expectHas: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := &ToolRegistry{
				tools: make(map[string]Tool),
			}

			if testCase.register {
				registry.Register("calculator", &Calculator{})
			}

			if actualHas := registry.Has("calculator"); actualHas != testCase.expectHas {
				t.Errorf(
					"expected Has to be %v, got %v",
					testCase.expectHas,
					actualHas,
				)
			}
		})
	}
}

func TestToolRegistryList(t *testing.T) {
	testCases := []struct {
		name          string
		tools         []string
		expectedTools []string
	}{
		{
			name: "Empty registry",
		},
		{
			name:          "Multiple tools",
			tools:         []string{"calculator", "shell", "filesystem"},
			expectedTools: []string{"calculator", "shell", "filesystem"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := &ToolRegistry{
				tools: make(map[string]Tool),
			}

			for _, name := range testCase.tools {
				registry.Register(name, &Calculator{})
			}

			actual := registry.List()

			if !reflect.DeepEqual(
				stringSet(actual),
				stringSet(testCase.expectedTools),
			) {
				t.Fatalf(
					"List returned %v, want %v",
					actual,
					testCase.expectedTools,
				)
			}
		})
	}
}

func TestToolRegistryRemove(t *testing.T) {
	testCases := []struct {
		name        string
		register    bool
		expectError bool
	}{
		{
			name:     "Remove existing tool",
			register: true,
		},
		{
			name:        "Remove missing tool",
			register:    false,
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := &ToolRegistry{
				tools: make(map[string]Tool),
			}

			if testCase.register {
				registry.Register("calculator", &Calculator{})
			}

			err := registry.Remove("calculator")

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf(
					"expected no error, got: %v",
					err,
				)
			}

			if registry.Has("calculator") {
				t.Fatal("expected tool to be removed")
			}
		})
	}
}

func TestToolRegistrySchemas(t *testing.T) {
	testCases := []struct {
		name                string
		registry            *ToolRegistry
		expectError         bool
		expectedCount       int
		expectedName        string
		expectedDescription string
		expectedSchema      map[string]any
	}{
		{
			name: "calculator schema",
			registry: &ToolRegistry{
				tools: map[string]Tool{
					"calculator": &Calculator{},
				},
			},
			expectError:         false,
			expectedCount:       1,
			expectedName:        "calculator",
			expectedDescription: "Performs basic arithmetic operations",
			expectedSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"operation": map[string]any{
						"type": "string",
					},
					"numbers": map[string]any{
						"type": "array",
						"items": map[string]any{
							"type": "number",
						},
					},
				},
				"required": []string{
					"operation",
					"numbers",
				},
			},
		},
		{
			name:        "nil registry",
			registry:    nil,
			expectError: true,
		},
		{
			name: "nil tool",
			registry: &ToolRegistry{
				tools: map[string]Tool{
					"broken": nil,
				},
			},
			expectError: true,
		},
		{
			name: "nil schema",
			registry: &ToolRegistry{
				tools: map[string]Tool{
					"broken": nilSchemaTool{},
				},
			},
			expectError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			schemas, err := testCase.registry.Schemas()

			if testCase.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}

				if schemas != nil {
					t.Fatalf("expected no schemas, got %v", schemas)
				}

				return
			}

			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if len(schemas) != testCase.expectedCount {
				t.Fatalf(
					"expected %d schema, got %d",
					testCase.expectedCount,
					len(schemas),
				)
			}

			if schemas[0]["name"] != testCase.expectedName {
				t.Fatalf(
					"expected %s schema, got %v",
					testCase.expectedName,
					schemas[0],
				)
			}

			if schemas[0]["description"] != testCase.expectedDescription {
				t.Fatalf(
					"expected description %q, got %v",
					testCase.expectedDescription,
					schemas[0]["description"],
				)
			}

			if !reflect.DeepEqual(
				schemas[0]["schema"],
				testCase.expectedSchema,
			) {
				t.Fatalf(
					"expected schema %v, got %v",
					testCase.expectedSchema,
					schemas[0]["schema"],
				)
			}
		})
	}
}

type nilSchemaTool struct{}

func (nilSchemaTool) Name() string {
	return "broken"
}

func (nilSchemaTool) Description() string {
	return "broken tool"
}

func (nilSchemaTool) Execute(map[string]any) (any, error) {
	return nil, nil
}

func (nilSchemaTool) Schema() map[string]any {
	return nil
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))

	for _, value := range values {
		result[value] = true
	}

	return result
}
