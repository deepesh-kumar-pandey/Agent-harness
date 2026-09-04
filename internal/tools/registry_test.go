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
		{name: "Registered tool", toolName: "custom", register: true},
		{name: "Missing tool", toolName: "missing", expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := &ToolRegistry{tools: make(map[string]Tool)}
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
		{name: "Calculator", toolName: "calculator"},
		{name: "Shell", toolName: "shell"},
		{name: "Filesystem", toolName: "filesystem"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if !NewToolRegistry().Has(testCase.toolName) {
				t.Errorf("expected %q to be registered", testCase.toolName)
			}
		})
	}
}

func TestToolRegistryRegister(t *testing.T) {
	testCases := []struct {
		name string
		tool Tool
	}{
		{name: "Register calculator", tool: &Calculator{}},
		{name: "Register shell", tool: &ShellTool{}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := &ToolRegistry{tools: make(map[string]Tool)}
			name, returnedTool := registry.Register("tool", testCase.tool)

			if name != "tool" || returnedTool != testCase.tool {
				t.Fatalf("Register returned (%q, %T), want (%q, %T)", name, returnedTool, "tool", testCase.tool)
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
		{name: "Registered tool", register: true, expectHas: true},
		{name: "Unregistered tool", expectHas: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := &ToolRegistry{tools: make(map[string]Tool)}
			if testCase.register {
				registry.Register("calculator", &Calculator{})
			}

			if actualHas := registry.Has("calculator"); actualHas != testCase.expectHas {
				t.Errorf("expected Has to be %v, got %v", testCase.expectHas, actualHas)
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
		{name: "Empty registry"},
		{
			name:          "Multiple tools",
			tools:         []string{"calculator", "shell", "filesystem"},
			expectedTools: []string{"calculator", "shell", "filesystem"},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := &ToolRegistry{tools: make(map[string]Tool)}
			for _, name := range testCase.tools {
				registry.Register(name, &Calculator{})
			}

			actual := registry.List()
			if !reflect.DeepEqual(stringSet(actual), stringSet(testCase.expectedTools)) {
				t.Fatalf("List returned %v, want %v", actual, testCase.expectedTools)
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
		{name: "Remove existing tool", register: true},
		{name: "Remove missing tool", expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			registry := &ToolRegistry{tools: make(map[string]Tool)}
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
				t.Fatalf("expected no error, got: %v", err)
			}
			if registry.Has("calculator") {
				t.Fatal("expected tool to be removed")
			}
		})
	}
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}
